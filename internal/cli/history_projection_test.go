package cli

import (
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// T024 (round 008) — the `-l` projection surfaces only the operator's messages;
// the widened tool steps never leak into the listing (FR-017). Round 073
// (ADR 0045): the projection returns the listing's role-tagged messages.

func TestListingMessagesOmitsToolSteps(t *testing.T) {
	entries := []history.Entry{{
		Prompt: "My name is Alice.",
		Answer: "Noted.",
		Steps:  []history.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE"}},
	}}
	msgs := listingMessages(entries, 10)
	if len(msgs) != 2 {
		t.Fatalf("listingMessages = %d messages, want 2 (prompt + answer)", len(msgs))
	}
	if msgs[0].Role != render.ListingOperator || msgs[0].Body != "My name is Alice." {
		t.Errorf("message[0] = %+v, want operator/My name is Alice.", msgs[0])
	}
	if msgs[1].Role != render.ListingModel || msgs[1].Body != "Noted." {
		t.Errorf("message[1] = %+v, want model/Noted.", msgs[1])
	}
	for _, m := range msgs {
		if m.Body == "ORANGE" || m.Body == "read_files" {
			t.Errorf("tool step leaked into the -l projection: %+v", m)
		}
	}
	if got := listingMessages(entries, 1); len(got) != 1 || got[0].Role != render.ListingModel {
		t.Errorf("listingMessages(…, 1) = %+v, want the last (model) message only", got)
	}
}
