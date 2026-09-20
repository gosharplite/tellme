package gemini

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// decodeContents is a tiny helper shared by the round-066 id pins.
func decodeContents(t *testing.T, body []byte) []struct {
	Role  string           `json:"role"`
	Parts []map[string]any `json:"parts"`
} {
	t.Helper()
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return decoded.Contents
}

// TestRequestBody_ToolPartsCarryIDs pins round 066 (ADR 0036; closes #134),
// FR-001/FR-002 (US1): on a Gemini tool round every `functionCall` part carries
// the call's `id` and every `functionResponse` part carries the `id` of the call
// it answers — the two ids are equal for a matched pair.
func TestRequestBody_ToolPartsCarryIDs(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`},
			{ID: "call_2", Name: "list_files", Arguments: `{"path":"."}`},
		}},
		{Role: "tool", Content: "note a", ToolCallID: "call_1"},
		{Role: "tool", Content: "listing", ToolCallID: "call_2"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	if len(turns) != 2 {
		t.Fatalf("contents len = %d, want 2 (model + batched user turn): %s", len(turns), body)
	}
	wantIDs := []string{"call_1", "call_2"}
	for i, want := range wantIDs {
		fc, _ := turns[0].Parts[i]["functionCall"].(map[string]any)
		if fc == nil {
			t.Fatalf("model turn part %d is not a functionCall: %+v", i, turns[0].Parts[i])
		}
		if fc["id"] != want {
			t.Errorf("model turn part %d functionCall.id = %v, want %q", i, fc["id"], want)
		}
		fr, _ := turns[1].Parts[i]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("batched turn part %d is not a functionResponse: %+v", i, turns[1].Parts[i])
		}
		if fr["id"] != want {
			t.Errorf("batched turn part %d functionResponse.id = %v, want %q (equal to its call's id)", i, fr["id"], want)
		}
	}
}

// TestRequestBody_OutOfOrderResultsPairByIdentity pins FR-006 (US2 Scenario 1):
// when a round's results are presented in an order different from the calls'
// order, each result binds to ITS call by `ToolCallID` — the response carries
// the right name and id, and the batched turn lists them in CALL order.
func TestRequestBody_OutOfOrderResultsPairByIdentity(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_A", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`},
			{ID: "call_B", Name: "list_files", Arguments: `{"path":"."}`},
		}},
		// Presented out of order: B's result arrives first.
		{Role: "tool", Content: "B listing", ToolCallID: "call_B"},
		{Role: "tool", Content: "A note", ToolCallID: "call_A"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	batch := turns[1]
	if batch.Role != "user" || len(batch.Parts) != 2 {
		t.Fatalf("batched turn = role %q with %d parts, want user with 2: %+v", batch.Role, len(batch.Parts), batch.Parts)
	}
	// In CALL order (A then B), each part named/ided by IDENTITY, not arrival.
	want := []struct{ id, name, content string }{
		{"call_A", "read_files", "A note"},
		{"call_B", "list_files", "B listing"},
	}
	for i, w := range want {
		fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("part %d is not a functionResponse: %+v", i, batch.Parts[i])
		}
		resp, _ := fr["response"].(map[string]any)
		if fr["id"] != w.id || fr["name"] != w.name || resp["content"] != w.content {
			t.Errorf("part %d = {id:%v name:%v content:%v}, want {id:%q name:%q content:%q}",
				i, fr["id"], fr["name"], resp["content"], w.id, w.name, w.content)
		}
	}
}

// TestRequestBody_EmptyToolCallIDOmitsID pins FR-003/FR-007 (US1 Scenario 2):
// a result with an empty `ToolCallID` emits NO `id` key (never `"id":""`) and
// pairs via the FIFO name fallback.
func TestRequestBody_EmptyToolCallIDOmitsID(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`},
		}},
		// No ToolCallID: an id-less (replay-shape) result.
		{Role: "tool", Content: "note a"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	if len(turns) != 2 {
		t.Fatalf("contents len = %d, want 2 (model + batched user turn): %s", len(turns), body)
	}
	fr, _ := turns[1].Parts[0]["functionResponse"].(map[string]any)
	if fr == nil {
		t.Fatalf("part 0 is not a functionResponse: %+v", turns[1].Parts[0])
	}
	if _, present := fr["id"]; present {
		t.Errorf("an id-less result must omit the `id` key, got %v", fr["id"])
	}
	if fr["name"] != "read_files" {
		t.Errorf("id-less result name = %v, want read_files (FIFO fallback)", fr["name"])
	}
	resp, _ := fr["response"].(map[string]any)
	if resp["content"] != "note a" {
		t.Errorf("id-less result content = %v, want %q", resp["content"], "note a")
	}
}

// TestRequestBody_UnmatchedToolCallIDFallsBackToFIFO pins FR-007: a result whose
// id matches no call of the round pairs by the FIFO fallback (deterministic,
// never a silent mispair) and — TD-066-1 — emits NO `id` key, because its id is on
// no `functionCall` part of the round (the reference's `response.id == call.id`
// invariant; a foreign id could itself 400 a strict provider).
func TestRequestBody_UnmatchedToolCallIDFallsBackToFIFO(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{}`},
		}},
		{Role: "tool", Content: "note a", ToolCallID: "call_zzz"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	fr, _ := turns[1].Parts[0]["functionResponse"].(map[string]any)
	if fr == nil {
		t.Fatalf("part 0 is not a functionResponse: %+v", turns[1].Parts[0])
	}
	if fr["name"] != "read_files" {
		t.Errorf("unmatched-id result name = %v, want read_files (FIFO fallback)", fr["name"])
	}
	if _, present := fr["id"]; present {
		t.Errorf("a foreign id must be omitted (no functionCall carries it), got %v", fr["id"])
	}
}

// TestRequestBody_ToolRoleMediaStillCarriesMedia pins F-066-2: the id-less-tool
// widening is restricted to a MEDIA-FREE `tool` message, so a `tool`-role message
// that carries media is still emitted as an inlineData turn (I-4: never silently
// lose an image).
func TestRequestBody_ToolRoleMediaStillCarriesMedia(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_image", Arguments: `{"filepath":"a.png"}`}}},
		{Role: "tool", Content: "attached", Media: []llm.MediaPart{{MIMEType: "image/png", Data: data}}},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	mediaFound := false
	for _, tn := range turns {
		for _, p := range tn.Parts {
			if _, ok := p["inlineData"]; ok {
				mediaFound = true
			}
		}
	}
	if !mediaFound {
		t.Errorf("a media-bearing `tool`-role message must still carry its image (no silent media loss): %s", body)
	}
}

// TestRequestBody_ReplayedStepIDsPairByIdentity pins TD-066-2: the replay path
// synthesises the SAME deterministic id on both sides (`call_step_<n>`), so a
// replayed round binds via the id-primary match (not FIFO), and the emitted
// parts carry equal ids. This is the real mechanism behind ADR 0036's replay
// fidelity claim (I-5).
func TestRequestBody_ReplayedStepIDsPairByIdentity(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_step_1", Name: "read_files", Arguments: `{}`}}},
		{Role: "tool", Content: "note a", ToolCallID: "call_step_1"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	fc, _ := turns[0].Parts[0]["functionCall"].(map[string]any)
	fr, _ := turns[1].Parts[0]["functionResponse"].(map[string]any)
	if fc == nil || fr == nil {
		t.Fatalf("expected a functionCall then a functionResponse: %+v", turns)
	}
	if fc["id"] != "call_step_1" || fr["id"] != "call_step_1" {
		t.Errorf("replay parts must carry equal ids: functionCall.id=%v functionResponse.id=%v, want call_step_1", fc["id"], fr["id"])
	}
}

// TestParseResponse_PrefersProviderFunctionCallID pins round 067 (ADR 0037;
// closes #136) FR-006 (US2 Scenario 1 / SC-002): when the Vertex response carries
// a `functionCall.id`, `parseResponse` PREFERS it over the synthetic `call_<n>`,
// and the preferred id then flows to both the emitted `functionCall` and its
// `functionResponse` (a single-family session keeps this Gemini-local).
func TestParseResponse_PrefersProviderFunctionCallID(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"parts":[{"functionCall":{"id":"vertex-abc","name":"read_files","args":{"filepaths":["a.txt"]}}}]}}]}`)
	resp, err := parseResponse(raw)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].ID != "vertex-abc" {
		t.Fatalf("parseResponse id preference: got %+v, want a single call with id vertex-abc", resp.ToolCalls)
	}
	// The preferred id flows through to the wire (call + its answer).
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: resp.ToolCalls},
		{Role: "tool", Content: "note a", ToolCallID: "vertex-abc"},
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	fc, _ := turns[0].Parts[0]["functionCall"].(map[string]any)
	fr, _ := turns[1].Parts[0]["functionResponse"].(map[string]any)
	if fc == nil || fr == nil {
		t.Fatalf("expected a functionCall then a functionResponse: %+v", turns)
	}
	if fc["id"] != "vertex-abc" || fr["id"] != "vertex-abc" {
		t.Errorf("provider id must flow to the wire: functionCall.id=%v functionResponse.id=%v, want vertex-abc", fc["id"], fr["id"])
	}
}

// TestParseResponse_FallsBackToDeterministicID pins FR-007/FR-008 (SC-002): a
// response WITHOUT a provider id — or with a blank/whitespace one — falls back to
// the deterministic `call_<n>`, identical across builds (never an empty id).
func TestParseResponse_FallsBackToDeterministicID(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"absent", `{"candidates":[{"content":{"parts":[{"functionCall":{"name":"a","args":{}}},{"functionCall":{"name":"b","args":{}}}]}}]}`},
		{"blank", `{"candidates":[{"content":{"parts":[{"functionCall":{"id":"   ","name":"a","args":{}}}]}}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, err := parseResponse([]byte(tc.raw))
			if err != nil {
				t.Fatalf("parseResponse: %v", err)
			}
			for i, tc2 := range first.ToolCalls {
				if want := fmt.Sprintf("call_%d", i+1); tc2.ID != want {
					t.Errorf("call %d id = %q, want deterministic %q", i, tc2.ID, want)
				}
			}
			// Deterministic across builds: a second parse yields the same ids.
			second, err := parseResponse([]byte(tc.raw))
			if err != nil {
				t.Fatalf("parseResponse (2): %v", err)
			}
			for i := range first.ToolCalls {
				if first.ToolCalls[i].ID != second.ToolCalls[i].ID {
					t.Errorf("call %d id not deterministic: %q vs %q", i, first.ToolCalls[i].ID, second.ToolCalls[i].ID)
				}
			}
		})
	}
}

// TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall pins round 067 (ADR 0037;
// closes #136) FR-001/FR-004 (US1 Scenario 1 / SC-001): a round that yields
// M < N results makes the UNPAIRED calls' ids accountable, in call order. (The
// round-066 `N=2 M=1` residual RF-066-8 is retired not by THIS short-round pin —
// at N=2, M=1 the pre-fold conditional partial drop is arithmetically equivalent
// and leaves this pin green — but by the CROSS-ROUND account in
// TestUnpairedCallIDs_MultiRound; see F-067-1.)
func TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{}`},
			{ID: "call_2", Name: "list_files", Arguments: `{}`},
		}},
		{Role: "tool", Content: "note a", ToolCallID: "call_1"}, // only call_1 is answered
	}
	got := UnpairedCallIDs(prior)
	if len(got) != 1 || got[0] != "call_2" {
		t.Fatalf("UnpairedCallIDs = %v, want [call_2] (the unpaired call's id, call order)", got)
	}
	// The emitted batched turn still carries only the M produced parts (I-2).
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	if len(turns) != 2 || len(turns[1].Parts) != 1 {
		t.Fatalf("emitted batched turn = %+v, want one user turn with exactly 1 part", turns)
	}
}

// TestUnpairedCallIDs_AllPairedIsEmpty pins FR-002 (US1 Scenario 2): a round whose
// every call is answered reports no unpaired call (the accounting invents no gap).
func TestUnpairedCallIDs_AllPairedIsEmpty(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_files", Arguments: `{}`},
			{ID: "call_2", Name: "list_files", Arguments: `{}`},
		}},
		{Role: "tool", Content: "a", ToolCallID: "call_1"},
		{Role: "tool", Content: "b", ToolCallID: "call_2"},
	}
	if got := UnpairedCallIDs(prior); len(got) != 0 {
		t.Fatalf("UnpairedCallIDs = %v, want empty (every call paired)", got)
	}
}

// TestUnpairedCallIDs_MultiRound pins FR-004: the account spans rounds, in call
// order — an unpaired call of an earlier round is not lost when a later round
// flushes, and each round's unpaired ids appear in call order.
func TestUnpairedCallIDs_MultiRound(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_r1a", Name: "read_files", Arguments: `{}`},
			{ID: "call_r1b", Name: "list_files", Arguments: `{}`},
		}},
		{Role: "tool", Content: "r1a", ToolCallID: "call_r1a"}, // call_r1b unpaired
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_r2a", Name: "get_tree", Arguments: `{}`},
		}},
		// call_r2a unpaired (the next model turn / prompt flushes round 2).
	}
	got := UnpairedCallIDs(prior)
	want := []string{"call_r1b", "call_r2a"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("UnpairedCallIDs = %v, want %v (call order across rounds)", got, want)
	}
}

// TestUnpairedCallIDs_MiddleRoundZeroResults pins spec Edge Case M = 0 (F-067-5):
// a round that yields ZERO results emits no batched turn and leaves ALL its call
// ids unpaired — including a MIDDLE round (results-less round followed by a
// further model round), whose ids are still accounted at that round's boundary.
func TestUnpairedCallIDs_MiddleRoundZeroResults(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_r1a", Name: "read_files", Arguments: `{}`},
			{ID: "call_r1b", Name: "list_files", Arguments: `{}`},
		}},
		// M = 0 for round 1: no results. The next model turn flushes it.
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_r2a", Name: "get_tree", Arguments: `{}`},
		}},
		{Role: "tool", Content: "tree", ToolCallID: "call_r2a"},
	}
	got := UnpairedCallIDs(prior)
	want := []string{"call_r1a", "call_r1b"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("UnpairedCallIDs = %v, want %v (all of the middle round's calls)", got, want)
	}
	// The M = 0 round emits no batched turn; only round 2 contributes a part.
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	userTurns := 0
	for _, tn := range turns {
		if tn.Role == "user" {
			userTurns++
		}
	}
	if userTurns != 1 {
		t.Fatalf("emitted user turns = %d, want 1 (the M=0 round emits no batched turn): %+v", userTurns, turns)
	}
}

// TestParseResponse_TrimsProviderID pins F-067-4: a provider id with surrounding
// whitespace is echoed TRIMMED (one normalisation shared with the blank-as-absent
// case), so the wire never carries stray padding.
func TestParseResponse_TrimsProviderID(t *testing.T) {
	raw := []byte(`{"candidates":[{"content":{"parts":[{"functionCall":{"id":"  vertex-abc  ","name":"read_files","args":{}}}]}}]}`)
	resp, err := parseResponse(raw)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].ID != "vertex-abc" {
		t.Fatalf("provider id must be trimmed: got %+v, want id vertex-abc", resp.ToolCalls)
	}
}

// TestUnpairedCallIDs_DuplicateProviderIDIsDeterministic pins the spec Edge Case
// (F-067-7): a degenerate response carrying the SAME provider id on two calls
// still binds each result deterministically to the first UNUSED identity match,
// in call order — never a silent mispair. Recorded in ADR 0037 §Forward RF-067-6.
func TestUnpairedCallIDs_DuplicateProviderIDIsDeterministic(t *testing.T) {
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "dup", Name: "read_files", Arguments: `{}`},
			{ID: "dup", Name: "list_files", Arguments: `{}`},
		}},
		{Role: "tool", Content: "first", ToolCallID: "dup"},
		{Role: "tool", Content: "second", ToolCallID: "dup"},
	}
	if got := UnpairedCallIDs(prior); len(got) != 0 {
		t.Fatalf("both duplicate-id calls must pair (first-unused match): got unpaired %v", got)
	}
	body, err := requestBody("", prior, nil, 0, 0, "", "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	turns := decodeContents(t, body)
	batch := turns[len(turns)-1]
	wantNames := []string{"read_files", "list_files"}
	for i, want := range wantNames {
		fr, _ := batch.Parts[i]["functionResponse"].(map[string]any)
		if fr == nil || fr["name"] != want {
			t.Fatalf("part %d name = %v, want %q (deterministic first-unused pairing)", i, fr["name"], want)
		}
	}
}
