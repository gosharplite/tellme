package cli

import (
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// T007 (round 027) — the turn header's number is the session's running
// AI-endpoint-call index: Σ of the prior entries' persisted call counts + 1,
// where an entry lacking a count (a legacy or arranged plain line) counts as 1.

func TestTurnNumber(t *testing.T) {
	tests := []struct {
		name  string
		prior []history.Entry
		want  int
	}{
		{"no prior turns is the first call", nil, 1},
		{"two plain turns made two calls", []history.Entry{{}, {}}, 3},
		{"a single tool-using turn made two calls", []history.Entry{{Calls: 2}}, 3},
		{"two tool-using turns made four calls", []history.Entry{{Calls: 2}, {Calls: 2}}, 5},
		{"a legacy entry (no count) counts as one", []history.Entry{{}}, 2},
		{"mixed tool-using and plain turns", []history.Entry{{Calls: 2}, {}}, 4},
		{"a three-round tool turn advances by three", []history.Entry{{Calls: 3}}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := turnNumber(tt.prior); got != tt.want {
				t.Fatalf("turnNumber(%+v) = %d, want %d", tt.prior, got, tt.want)
			}
		})
	}
}
