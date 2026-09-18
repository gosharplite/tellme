package cli

import (
	"io"
	"strings"
	"testing"
)

// TestParseFlagsListDefault pins the round-054 (ADR 0023) `-l` default: a bare
// `-l` (and `--list`) means 1; an adjacent integer is consumed as the value; a
// non-integer token is left as a positional argument (a prompt).
func TestParseFlagsListDefault(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		args     []string
		wantList int
		wantArgs []string
		wantSet  bool
	}{
		{"bare short", []string{"-l"}, 1, nil, true},
		{"bare long", []string{"--list"}, 1, nil, true},
		{"adjacent value short", []string{"-l", "5"}, 5, nil, true},
		{"adjacent value long", []string{"--list", "5"}, 5, nil, true},
		{"attached value", []string{"-l=3"}, 3, nil, true},
		{"followed by a prompt", []string{"-l", "hello"}, 1, []string{"hello"}, true},
		{"followed by another flag", []string{"-l", "-r"}, 1, nil, true},
		{"absent", []string{"hi"}, 0, []string{"hi"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o, rest, ok := parseFlags(tc.args, io.Discard)
			if !ok {
				t.Fatalf("parseFlags(%v) ok = false", tc.args)
			}
			if o.listSet != tc.wantSet || o.list != tc.wantList {
				t.Errorf("parseFlags(%v): listSet=%v list=%d, want %v/%d", tc.args, o.listSet, o.list, tc.wantSet, tc.wantList)
			}
			if strings.Join(rest, ",") != strings.Join(tc.wantArgs, ",") {
				t.Errorf("parseFlags(%v) args = %v, want %v", tc.args, rest, tc.wantArgs)
			}
		})
	}
}
