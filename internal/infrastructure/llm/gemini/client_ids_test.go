package gemini

import (
	"encoding/json"
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
// never a silent mispair) and still carries its own (unmatched) id.
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
	if fr["id"] != "call_zzz" {
		t.Errorf("unmatched-id result id = %v, want call_zzz (its own id)", fr["id"])
	}
}
