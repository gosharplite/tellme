package history

import "testing"

// T007 (round 027) — the domain reading of the persisted call counts: each entry
// contributes its Calls; an entry without one counts as 1.

func TestTotalCalls(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    int
	}{
		{"no entries", nil, 0},
		{"a plain turn counts as one", []Entry{{}}, 1},
		{"a tool-using turn counts its calls", []Entry{{Calls: 2}}, 2},
		{"a three-round tool turn", []Entry{{Calls: 3}}, 3},
		{"mixed legacy and counted", []Entry{{Calls: 2}, {}, {Calls: 3}}, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalCalls(tt.entries); got != tt.want {
				t.Fatalf("TotalCalls(%+v) = %d, want %d", tt.entries, got, tt.want)
			}
		})
	}
}
