package llm

import (
	"testing"

	"github.com/gosharplite/tellme/internal/domain/tools"
)

// TestUnpairedToolCalls pins round 068 (ADR 0038; surfaces ADR 0037 RF-067-1):
// the family-neutral single owner of the round-boundary account. The
// Gemini/Vertex adapter's UnpairedCallIDs delegates here.
func TestUnpairedToolCalls(t *testing.T) {
	calls := func(ids ...string) Message {
		tcs := make([]ToolCall, 0, len(ids))
		for _, id := range ids {
			tcs = append(tcs, ToolCall{ID: id, Name: "t"})
		}
		return Message{Role: "assistant", ToolCalls: tcs}
	}
	tests := []struct {
		name string
		msgs []Message
		want []string
	}{
		{"all-paired", []Message{calls("a", "b"), {Role: "tool", Content: "x", ToolCallID: "a"}, {Role: "tool", Content: "y", ToolCallID: "b"}}, nil},
		{"short-round", []Message{calls("a", "b"), {Role: "tool", Content: "x", ToolCallID: "a"}}, []string{"b"}},
		{"zero-results", []Message{calls("a", "b"), calls("c")}, []string{"a", "b", "c"}}, // both rounds have M=0
		{"multi-round", []Message{calls("r1a", "r1b"), {Role: "tool", Content: "x", ToolCallID: "r1a"}, calls("r2a")}, []string{"r1b", "r2a"}},
		{"out-of-order", []Message{calls("a", "b"), {Role: "tool", Content: "y", ToolCallID: "b"}, {Role: "tool", Content: "x", ToolCallID: "a"}}, nil},
		{"id-less-fifo", []Message{calls("a"), {Role: "tool", Content: "x"}}, nil},
		{"duplicate-id", []Message{calls("dup", "dup"), {Role: "tool", Content: "1", ToolCallID: "dup"}, {Role: "tool", Content: "2", ToolCallID: "dup"}}, nil},
		{"terminal-text-closes-round", []Message{calls("a"), {Role: "user", Content: "hi"}}, []string{"a"}},
		{"media-not-a-boundary", []Message{calls("a", "b"), {Role: "user", Media: []tools.MediaPart{{MIMEType: "image/png"}}}, {Role: "tool", Content: "x", ToolCallID: "a"}}, []string{"b"}},
		{"media-after-round", []Message{calls("a"), {Role: "tool", Content: "x", ToolCallID: "a"}, {Role: "user", Media: []tools.MediaPart{{MIMEType: "image/png"}}}}, nil},
		{"tool-role-media-not-a-result", []Message{calls("a"), {Role: "tool", Content: "m", Media: []tools.MediaPart{{MIMEType: "image/png"}}}}, []string{"a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := UnpairedToolCalls(tc.msgs)
			if len(got) != len(tc.want) {
				t.Fatalf("UnpairedToolCalls = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("UnpairedToolCalls = %v, want %v", got, tc.want)
				}
			}
		})
	}
}
