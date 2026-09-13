package openai

import (
	"encoding/json"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// TDD regression (review PR #25 BLOCKER-1): an empty prompt must NOT append a
// trailing user message, so a tool-loop round whose active turn is folded into
// Messages keeps the chronology `user → assistant(tool_calls) → tool(result)`.
func TestRequestBody_EmptyPromptOmitsUserMessage(t *testing.T) {
	prior := []llm.Message{
		{Role: "user", Content: "read notes.txt"},
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "call_1", Name: "read_files", Arguments: `{"path":"notes.txt"}`}}},
		{Role: "tool", Content: "ORANGE", ToolCallID: "call_1"},
	}
	body, err := requestBody("m", "", prior, nil, 0, "")
	if err != nil {
		t.Fatalf("requestBody: %v", err)
	}
	var decoded struct {
		Messages []struct {
			Role       string `json:"role"`
			ToolCallID string `json:"tool_call_id"`
			ToolCalls  []struct {
				ID string `json:"id"`
			} `json:"tool_calls"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Messages) != 3 {
		t.Fatalf("messages = %d, want 3 (no trailing empty user message): %+v", len(decoded.Messages), decoded.Messages)
	}
	roles := []string{decoded.Messages[0].Role, decoded.Messages[1].Role, decoded.Messages[2].Role}
	if roles[0] != "user" || roles[1] != "assistant" || roles[2] != "tool" {
		t.Fatalf("message roles = %v, want [user assistant tool]", roles)
	}
	if len(decoded.Messages[1].ToolCalls) != 1 || decoded.Messages[1].ToolCalls[0].ID != "call_1" {
		t.Errorf("assistant tool_calls = %+v, want one call_1", decoded.Messages[1].ToolCalls)
	}
	if decoded.Messages[2].ToolCallID != "call_1" {
		t.Errorf("tool result tool_call_id = %q, want call_1", decoded.Messages[2].ToolCallID)
	}
}
