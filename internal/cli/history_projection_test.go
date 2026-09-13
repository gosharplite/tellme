package cli

import (
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// T024 (round 008) — the `-l` projection surfaces only the operator's messages;
// the widened tool steps never leak into the listing (FR-017).

func TestToMessagesOmitsToolSteps(t *testing.T) {
	entries := []history.Entry{{
		Prompt: "My name is Alice.",
		Answer: "Noted.",
		Steps:  []history.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE"}},
	}}
	msgs := toMessages(entries)
	if len(msgs) != 2 {
		t.Fatalf("toMessages = %d messages, want 2 (prompt + answer)", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "My name is Alice." {
		t.Errorf("message[0] = %+v, want user/My name is Alice.", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "Noted." {
		t.Errorf("message[1] = %+v, want assistant/Noted.", msgs[1])
	}
	for _, m := range msgs {
		if m.Content == "ORANGE" || m.Content == "read_files" {
			t.Errorf("tool step leaked into the -l projection: %+v", m)
		}
	}
}
