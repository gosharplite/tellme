package gemini

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// TestRequestBody_MediaBecomesInlineData pins round-063 (ADR 0033 D2/D3): a
// media-bearing message becomes its OWN `user` contents entry, carrying one
// `inlineData` part per media part (camelCase proto-JSON keys, base64
// StdEncoding) — emitted AFTER the tool-result functionResponse turn, so the
// Vertex parser's [InlineData][FunctionResponse] ordering hazard (#1441) cannot
// arise. Round 062's loud refusal is retired: the family now CARRIES the image.
func TestRequestBody_MediaBecomesInlineData(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_image", Arguments: `{"filepath":"shot.png"}`}}},
		{Role: "tool", Content: "Successfully read image", ToolCallID: "call_1"},
		{Role: "user", Media: []domaintools.MediaPart{{MIMEType: "image/png", Data: data}}},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	if strings.Contains(string(body), `"inline_data"`) || strings.Contains(string(body), `"mime_type"`) {
		t.Errorf("the wire must use the camelCase proto-JSON keys (`inlineData`/`mimeType`): %s", body)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Contents) != 3 {
		t.Fatalf("contents len = %d, want 3: %s", len(decoded.Contents), body)
	}
	// The tool result is the functionResponse turn...
	fr, _ := decoded.Contents[1].Parts[0]["functionResponse"].(map[string]any)
	if decoded.Contents[1].Role != "user" || fr == nil {
		t.Fatalf("tool result turn is not a functionResponse: %+v", decoded.Contents[1])
	}
	// ...and the image is carried by the FOLLOWING user turn (after it).
	media := decoded.Contents[2]
	if media.Role != "user" {
		t.Errorf("media turn role = %q, want user", media.Role)
	}
	blob, _ := media.Parts[0]["inlineData"].(map[string]any)
	if blob == nil {
		t.Fatalf("media turn carries no inlineData part: %+v", media.Parts)
	}
	if blob["mimeType"] != "image/png" {
		t.Errorf("inlineData.mimeType = %v, want image/png", blob["mimeType"])
	}
	raw, _ := blob["data"].(string)
	got, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("inlineData.data is not base64: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("inlineData bytes = %x, want the file's exact bytes %x", got, data)
	}
}

// TestRequestBody_MultiCallRound_BatchesFunctionResponses pins round 065
// (ADR 0035; closes #132): a round with TWO tool calls emits the round's tool
// results as ONE `user` turn carrying TWO `functionResponse` parts (in call
// order) — the count Vertex checks (`len(functionResponse) == len(functionCall)`
// per turn) — and the round's media turns FOLLOW the batched turn:
//
//	[model: fcA, fcB] [user: frA, frB] [user: inlineData(A)] [user: inlineData(B)]
//
// This supersedes the round-063 TD-063-1 interleave pin (the pre-065 shape that
// made Vertex reject a >=2-call round).
//
// Round 083 (ADR 0055; closes #167) note: this pin feeds the adapter a SYNTHETIC
// interleaved `prior` (`tool, user(media), tool, user(media)`) to exercise the
// adapter's per-message buffering directly. The post-083 LIVE loop no longer
// emits that shape — it hands the adapter ONE media message per round — so the
// live shape is `[model: fcA, fcB] [user: frA, frB] [user: inlineData(A), inlineData(B)]`
// (pinned by TestRequestBody_RoundScopedMedia_OneTurnTwoParts).
func TestRequestBody_MultiCallRound_BatchesFunctionResponses(t *testing.T) {
	dataA := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	dataB := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_image", Arguments: `{"filepath":"a.png"}`},
			{ID: "call_2", Name: "read_image", Arguments: `{"filepath":"b.jpg"}`},
		}},
		{Role: "tool", Content: "a ok", ToolCallID: "call_1"},
		{Role: "user", Media: []domaintools.MediaPart{{MIMEType: "image/png", Data: dataA}}},
		{Role: "tool", Content: "b ok", ToolCallID: "call_2"},
		{Role: "user", Media: []domaintools.MediaPart{{MIMEType: "image/jpeg", Data: dataB}}},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Contents) != 4 {
		t.Fatalf("contents len = %d, want 4 (model, batched fr, mediaA, mediaB): %s", len(decoded.Contents), body)
	}
	if decoded.Contents[0].Role != "model" {
		t.Errorf("turn 0 role = %q, want model", decoded.Contents[0].Role)
	}
	// The batched function-response turn: TWO parts, in call order.
	batch := decoded.Contents[1]
	if batch.Role != "user" {
		t.Errorf("batched turn role = %q, want user", batch.Role)
	}
	if len(batch.Parts) != 2 {
		t.Fatalf("batched turn parts = %d, want 2 (one functionResponse per call): %+v", len(batch.Parts), batch.Parts)
	}
	for i, want := range []string{"a ok", "b ok"} {
		fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("batched turn part %d is not a functionResponse: %+v", i, batch.Parts[i])
		}
		resp, _ := fr["response"].(map[string]any)
		if resp["content"] != want {
			t.Errorf("batched turn part %d functionResponse content = %v, want %q", i, resp["content"], want)
		}
	}
	// The media turns follow the batch, standalone, each with its own blob.
	for i, want := range map[int][]byte{2: dataA, 3: dataB} {
		blob, _ := decoded.Contents[i].Parts[0]["inlineData"].(map[string]any)
		if blob == nil {
			t.Fatalf("turn %d carries no inlineData part: %+v", i, decoded.Contents[i].Parts)
		}
		raw, _ := blob["data"].(string)
		got, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			t.Fatalf("turn %d inlineData.data is not base64: %v", i, err)
		}
		if string(got) != string(want) {
			t.Errorf("turn %d inlineData bytes = %x, want %x", i, got, want)
		}
	}
}

// TestRequestBody_MultiCallRound_NoMedia_BatchesResults pins the media-free
// multi-call case (the exact #132 reproduction): a model turn with TWO tool calls
// followed by their two results serializes to ONE batched `user` turn.
func TestRequestBody_MultiCallRound_NoMedia_BatchesResults(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["first.txt"]}`},
			{ID: "call_2", Name: "read_files", Arguments: `{"filepaths":["second.txt"]}`},
		}},
		{Role: "tool", Content: "the first note", ToolCallID: "call_1"},
		{Role: "tool", Content: "the second note", ToolCallID: "call_2"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Contents) != 2 {
		t.Fatalf("contents len = %d, want 2 (model + one batched user turn): %s", len(decoded.Contents), body)
	}
	batch := decoded.Contents[1]
	if batch.Role != "user" || len(batch.Parts) != 2 {
		t.Fatalf("batched turn = role %q with %d parts, want user with 2 functionResponse parts: %+v", batch.Role, len(batch.Parts), batch.Parts)
	}
	for i, want := range []string{"the first note", "the second note"} {
		fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("part %d is not a functionResponse: %+v", i, batch.Parts[i])
		}
		resp, _ := fr["response"].(map[string]any)
		if resp["content"] != want {
			t.Errorf("part %d functionResponse content = %v, want %q", i, resp["content"], want)
		}
		// Round 066 (ADR 0036): the part carries its call's id.
		if wantID := []string{"call_1", "call_2"}[i]; fr["id"] != wantID {
			t.Errorf("part %d functionResponse id = %v, want %q", i, fr["id"], wantID)
		}
	}
}

// TestRequestBody_ThreeCallRound_BatchesResults pins the N == 3 boundary with the
// same claim as the N == 2 pin: one batched `user` turn with three
// `functionResponse` parts, correct role, names in call order.
func TestRequestBody_ThreeCallRound_BatchesResults(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`},
			{ID: "call_2", Name: "list_files", Arguments: `{"path":"."}`},
			{ID: "call_3", Name: "get_tree", Arguments: `{"path":"."}`},
		}},
		{Role: "tool", Content: "note a", ToolCallID: "call_1"},
		{Role: "tool", Content: "listing", ToolCallID: "call_2"},
		{Role: "tool", Content: "tree", ToolCallID: "call_3"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Contents) != 2 {
		t.Fatalf("contents len = %d, want 2 (model + one batched user turn): %s", len(decoded.Contents), body)
	}
	if decoded.Contents[0].Role != "model" {
		t.Errorf("turn 0 role = %q, want model", decoded.Contents[0].Role)
	}
	batch := decoded.Contents[1]
	if batch.Role != "user" || len(batch.Parts) != 3 {
		t.Fatalf("batched turn = role %q with %d parts, want user with 3 functionResponse parts: %+v", batch.Role, len(batch.Parts), batch.Parts)
	}
	for i, want := range []struct{ name, content string }{
		{"read_files", "note a"}, {"list_files", "listing"}, {"get_tree", "tree"},
	} {
		fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("part %d is not a functionResponse: %+v", i, batch.Parts[i])
		}
		if fr["name"] != want.name {
			t.Errorf("part %d functionResponse name = %v, want %q (call order)", i, fr["name"], want.name)
		}
		resp, _ := fr["response"].(map[string]any)
		if resp["content"] != want.content {
			t.Errorf("part %d functionResponse content = %v, want %q", i, resp["content"], want.content)
		}
		// Round 066 (ADR 0036): the part carries its call's id, in call order.
		if wantID := []string{"call_1", "call_2", "call_3"}[i]; fr["id"] != wantID {
			t.Errorf("part %d functionResponse id = %v, want %q", i, fr["id"], wantID)
		}
	}
}

// TestRequestBody_ShortRound_DropsUnpairedNames pins the TD-065-2 fold for ANY
// N: a round that yields FEWER results than its model turn's calls (M < N) emits
// the M results it has in one batched turn (the documented shape — a provider
// that rejects it surfaces the same 400, never a silent name mispair), and EVERY
// name left unpaired at the round boundary is dropped so a LATER round's parts are
// named correctly. Covers M == 1 with N == 2 and N == 3, and M == 2 with N == 3.
func TestRequestBody_ShortRound_DropsUnpairedNames(t *testing.T) {
	cases := []struct {
		name     string
		calls    []llm.ToolCall
		results  []llm.Message
		wantLeft []string // round-1 batch names
		wantNext string   // round-2 (single call) batch name
	}{
		{
			name: "N=2 M=1",
			calls: []llm.ToolCall{
				{ID: "call_1", Name: "read_files", Arguments: `{}`},
				{ID: "call_2", Name: "list_files", Arguments: `{}`},
			},
			results:  []llm.Message{{Role: "tool", Content: "only a", ToolCallID: "call_1"}},
			wantLeft: []string{"read_files"}, wantNext: "search",
		},
		{
			name: "N=3 M=1",
			calls: []llm.ToolCall{
				{ID: "call_1", Name: "read_files", Arguments: `{}`},
				{ID: "call_2", Name: "list_files", Arguments: `{}`},
				{ID: "call_3", Name: "get_tree", Arguments: `{}`},
			},
			results:  []llm.Message{{Role: "tool", Content: "only a", ToolCallID: "call_1"}},
			wantLeft: []string{"read_files"}, wantNext: "search",
		},
		{
			name: "N=3 M=2",
			calls: []llm.ToolCall{
				{ID: "call_1", Name: "read_files", Arguments: `{}`},
				{ID: "call_2", Name: "list_files", Arguments: `{}`},
				{ID: "call_3", Name: "get_tree", Arguments: `{}`},
			},
			results: []llm.Message{
				{Role: "tool", Content: "note a", ToolCallID: "call_1"},
				{Role: "tool", Content: "listing", ToolCallID: "call_2"},
			},
			wantLeft: []string{"read_files", "list_files"}, wantNext: "search",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prior := append([]llm.Message{{Role: "assistant", ToolCalls: tc.calls}}, tc.results...)
			// A second round with its own single call + result must pair correctly.
			prior = append(prior,
				llm.Message{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_9", Name: "search", Arguments: `{}`}}},
				llm.Message{Role: "tool", Content: "tree", ToolCallID: "call_9"},
			)
			body, err := requestBody("", prior, nil, 0, 0, "", "")
			if err != nil {
				t.Fatalf("requestBody: %v", err)
			}
			var decoded struct {
				Contents []struct {
					Role  string           `json:"role"`
					Parts []map[string]any `json:"parts"`
				} `json:"contents"`
			}
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(decoded.Contents) != 4 {
				t.Fatalf("contents len = %d, want 4 (model·batch·model·batch): %s", len(decoded.Contents), body)
			}
			// Round 1's batched turn carries exactly the M results it has, named in order.
			batch := decoded.Contents[1]
			if batch.Role != "user" || len(batch.Parts) != len(tc.wantLeft) {
				t.Fatalf("round-1 batch = role %q with %d parts, want user with %d: %+v", batch.Role, len(batch.Parts), len(tc.wantLeft), batch.Parts)
			}
			for i, want := range tc.wantLeft {
				fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
				if fr["name"] != want {
					t.Errorf("round-1 part %d name = %v, want %q", i, fr["name"], want)
				}
			}
			// Round 2's part must be named for ITS call (no mispair from round 1).
			fr1, _ := decoded.Contents[3].Parts[0]["functionResponse"].(map[string]any)
			if fr1["name"] != tc.wantNext {
				t.Errorf("round-2 functionResponse name = %v, want %q (no mispair from the short round)", fr1["name"], tc.wantNext)
			}
		})
	}
}

// TestRequestBody_MediaFreeIsByteIdentical pins I-1 on the Gemini family: a
// media-free conversation serializes BYTE-FOR-BYTE as before — any drift in the
// text path reds this (the round-062 byte-identity control, carried to Gemini).
func TestRequestBody_MediaFreeIsByteIdentical(t *testing.T) {
	prior := []llm.Message{{Role: "user", Content: "hi"}}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	const want = `{"contents":[{"parts":[{"text":"hi"}],"role":"user"}]}`
	if string(body) != want {
		t.Errorf("media-free body drifted:\n got %s\nwant %s", body, want)
	}
}

// TestRequestBody_RoundScopedMedia_OneTurnTwoParts pins round 083 (ADR 0055;
// closes #167) D3: the post-083 LOOP hands the adapter ONE media message per
// round carrying N media parts (call order), so the Gemini adapter emits ONE
// media `user` turn with N `inlineData` parts — not N single-part media turns.
// (Pre-083 the loop handed it per-call messages, which the adapter turned into N
// turns; that interleaved `prior` is still what the round-065 pin feeds directly,
// but the live loop no longer produces it — see
// TestRequestBody_MultiCallRound_BatchesFunctionResponses.)
func TestRequestBody_RoundScopedMedia_OneTurnTwoParts(t *testing.T) {
	dataA := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	dataB := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	// The post-083 round shape: assistant(tool_calls ×2), the two tool results
	// CONTIGUOUS, then ONE media message carrying BOTH images.
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_image", Arguments: `{"filepath":"a.png"}`},
			{ID: "call_2", Name: "read_image", Arguments: `{"filepath":"b.jpg"}`},
		}},
		{Role: "tool", Content: "a ok", ToolCallID: "call_1"},
		{Role: "tool", Content: "b ok", ToolCallID: "call_2"},
		{Role: "user", Media: []domaintools.MediaPart{
			{MIMEType: "image/png", Data: dataA},
			{MIMEType: "image/jpeg", Data: dataB},
		}},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// [model(fcA,fcB)] [user(frA,frB)] [user(inlineData A, inlineData B)]
	if len(decoded.Contents) != 3 {
		t.Fatalf("contents len = %d, want 3 (model, batched fr, ONE media turn): %s", len(decoded.Contents), body)
	}
	mediaTurn := decoded.Contents[2]
	if mediaTurn.Role != "user" {
		t.Errorf("media turn role = %q, want user", mediaTurn.Role)
	}
	if len(mediaTurn.Parts) != 2 {
		t.Fatalf("media turn parts = %d, want 2 (one inlineData per media part): %+v", len(mediaTurn.Parts), mediaTurn.Parts)
	}
	for i, want := range [][]byte{dataA, dataB} {
		blob, _ := mediaTurn.Parts[i]["inlineData"].(map[string]any)
		if blob == nil {
			t.Fatalf("media turn part %d carries no inlineData: %+v", i, mediaTurn.Parts[i])
		}
		raw, _ := blob["data"].(string)
		got, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			t.Fatalf("media turn part %d inlineData.data is not base64: %v", i, err)
		}
		if string(got) != string(want) {
			t.Errorf("media turn part %d bytes = %x, want %x", i, got, want)
		}
	}
}
