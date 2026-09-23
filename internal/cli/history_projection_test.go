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

// Round 082 (ADR 0054) — the backward turn index.

// TestListingMessagesStampsBackwardTurnIndex pins that each message carries its
// turn's distance from the most recent turn (1 = newest), and that both messages
// of one entry share it.
func TestListingMessagesStampsBackwardTurnIndex(t *testing.T) {
	entries := []history.Entry{
		{Prompt: "q1", Answer: "a1"},
		{Prompt: "q2", Answer: "a2"},
		{Prompt: "q3", Answer: "a3"},
	}
	msgs := listingMessages(entries, 6)
	want := []int{3, 3, 2, 2, 1, 1}
	if len(msgs) != len(want) {
		t.Fatalf("listingMessages = %d messages, want %d", len(msgs), len(want))
	}
	for i, m := range msgs {
		if m.TurnIndex != want[i] {
			t.Errorf("message[%d].TurnIndex = %d, want %d", i, m.TurnIndex, want[i])
		}
	}
}

// TestListingMessagesKeepsTrueDistanceOnPartialSlice pins the load-bearing
// arithmetic: the index is the TRUE distance from the end of the loaded history,
// computed BEFORE the message-count truncation. An odd `-l N` leaves a lone
// leading [MODEL] that must keep its true distance (- 2), never be renumbered to
// - 1 by its position in the printed window.
func TestListingMessagesKeepsTrueDistanceOnPartialSlice(t *testing.T) {
	entries := []history.Entry{
		{Prompt: "q1", Answer: "a1"},
		{Prompt: "q2", Answer: "a2"},
	}
	msgs := listingMessages(entries, 3) // the last 3 of 4 messages
	want := []struct {
		role  render.ListingRole
		index int
	}{
		{render.ListingModel, 2},
		{render.ListingOperator, 1},
		{render.ListingModel, 1},
	}
	if len(msgs) != len(want) {
		t.Fatalf("listingMessages(…, 3) = %d messages, want %d", len(msgs), len(want))
	}
	for i, w := range want {
		if msgs[i].Role != w.role || msgs[i].TurnIndex != w.index {
			t.Errorf("message[%d] = %+v, want role=%v index=%d", i, msgs[i], w.role, w.index)
		}
	}
}
