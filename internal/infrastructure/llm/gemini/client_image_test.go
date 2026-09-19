package gemini

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
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
		{Role: "user", Media: []llm.MediaPart{{MIMEType: "image/png", Data: data}}},
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

// TestRequestBody_MultiCallRound_MediaTurnsInterleave pins the TD-063-1 shape
// (recorded, not fixed): a round with TWO read_image calls appends one media
// message per call, so the serialized body interleaves
//
//	[model: fcA, fcB] [user: frA] [user: inlineData(A)] [user: frB] [user: inlineData(B)]
//
// — the media turn of the first call sits BETWEEN the two functionResponse
// turns. Each blob is carried exactly once and in order, but the round's
// functionResponses are SPLIT across turns (a pre-existing trait this round does
// not introduce; the media insertion does not add a new hazard — every inline
// turn is still standalone). The round-scoped placement that would remove the
// question is ADR 0033 RF-063-7.
func TestRequestBody_MultiCallRound_MediaTurnsInterleave(t *testing.T) {
	dataA := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	dataB := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	prior := []llm.Message{
		{Role: "assistant", ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "read_image", Arguments: `{"filepath":"a.png"}`},
			{ID: "call_2", Name: "read_image", Arguments: `{"filepath":"b.jpg"}`},
		}},
		{Role: "tool", Content: "a ok", ToolCallID: "call_1"},
		{Role: "user", Media: []llm.MediaPart{{MIMEType: "image/png", Data: dataA}}},
		{Role: "tool", Content: "b ok", ToolCallID: "call_2"},
		{Role: "user", Media: []llm.MediaPart{{MIMEType: "image/jpeg", Data: dataB}}},
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
	if len(decoded.Contents) != 5 {
		t.Fatalf("contents len = %d, want 5 (model, frA, mediaA, frB, mediaB): %s", len(decoded.Contents), body)
	}
	if decoded.Contents[0].Role != "model" {
		t.Errorf("turn 0 role = %q, want model", decoded.Contents[0].Role)
	}
	for i, want := range map[int]string{1: "a ok", 3: "b ok"} {
		fr, _ := decoded.Contents[i].Parts[0]["functionResponse"].(map[string]any)
		if fr == nil {
			t.Fatalf("turn %d is not a functionResponse: %+v", i, decoded.Contents[i])
		}
		resp, _ := fr["response"].(map[string]any)
		if resp["content"] != want {
			t.Errorf("turn %d functionResponse content = %v, want %q", i, resp["content"], want)
		}
	}
	for i, want := range map[int][]byte{2: dataA, 4: dataB} {
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
