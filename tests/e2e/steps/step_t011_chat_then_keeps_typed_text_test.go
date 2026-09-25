package steps

import "testing"

// TestJoinedRowCatchesTheJoinedState pins the round-093 discriminating clause of
// the keep-the-typed-text assertion: an editor row that carries two distinct
// typed lines is a defect (the pre-093 `line oneline two` row passed the old
// substring check vacuously). Separate rows must report no join.
func TestJoinedRowCatchesTheJoinedState(t *testing.T) {
	parts := []string{"line one", "line two"}
	joined := []string{"│┃ line oneline two", "│┃"}
	if got := joinedRow(joined, parts); got == "" {
		t.Fatalf("joinedRow did not catch a row carrying both typed lines: %q", joined)
	}
	separate := []string{"│┃ line one", "│┃ line two"}
	if got := joinedRow(separate, parts); got != "" {
		t.Fatalf("joinedRow flagged separate rows: %q", got)
	}
	if got := joinedRow(separate, []string{"line one"}); got != "" {
		t.Fatalf("joinedRow must not flag a single-line value: %q", got)
	}
}
