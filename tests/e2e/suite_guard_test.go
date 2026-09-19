package e2e

import (
	"strings"
	"testing"
)

// TestE2EConcurrency pins the concurrency resolver's branchy fallbacks (round
// 055 review TD-055-2). A value that does not parse, or is < 1, or is absent,
// falls back to the default; a valid value >= 1 is honoured.
func TestE2EConcurrency(t *testing.T) {
	cases := []struct {
		env  string
		want int
	}{
		{"", e2eDefaultConcurrency},    // unset
		{"0", e2eDefaultConcurrency},   // < 1
		{"-1", e2eDefaultConcurrency},  // < 1
		{"abc", e2eDefaultConcurrency}, // not an int
		{" 3", e2eDefaultConcurrency},  // Atoi rejects the space
		{"1", 1},                       // honoured
		{"8", 8},                       // honoured
		{" 8 ", e2eDefaultConcurrency}, // stray whitespace is not trimmed by the resolver
	}
	for _, tc := range cases {
		t.Setenv("TELL_ME_E2E_CONCURRENCY", tc.env)
		if got := e2eConcurrency(); got != tc.want {
			t.Errorf("e2eConcurrency() with TELL_ME_E2E_CONCURRENCY=%q = %d, want %d", tc.env, got, tc.want)
		}
	}
}

// TestSplitFeaturePaths pins the selection tokenizer: blanks are trimmed,
// empties dropped.
func TestSplitFeaturePaths(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{",", nil},
		{" a , , b ", []string{"a", "b"}},
		{"a,b", []string{"a", "b"}},
	}
	for _, tc := range cases {
		got := splitFeaturePaths(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("splitFeaturePaths(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitFeaturePaths(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}

// TestE2EPaths pins the path resolver: no selection ⇒ the whole contract root
// (the gate); a selection ⇒ exactly those paths.
func TestE2EPaths(t *testing.T) {
	old := e2ePathsOverride
	defer func() { e2ePathsOverride = old }()

	e2ePathsOverride = ""
	if got := e2ePaths(); len(got) != 1 || got[0] != e2eFeaturesRoot {
		t.Errorf("e2ePaths() with no selection = %v, want [%s] (the gate)", got, e2eFeaturesRoot)
	}

	e2ePathsOverride = "../../specs/truth/features/cli/history"
	if got := e2ePaths(); len(got) != 1 || got[0] != "../../specs/truth/features/cli/history" {
		t.Errorf("e2ePaths() with a selection = %v, want the selection", got)
	}
}

// TestGuardSelection pins the never-the-gate invariant (round 055 review
// B-055-1, TD-055-2): a provided selection that RESOLVES to the whole contract
// is refused — including the traversal and all-modules spellings a
// string-equality guard would wave through — while a genuine subset passes and
// the gate itself (no selection) is never refused.
func TestGuardSelection(t *testing.T) {
	const root = "../../specs/truth/features/cli"
	allModules := []string{root + "/chat", root + "/configuration", root + "/history",
		root + "/workspace", root + "/usage", root + "/diagnostics"}

	cases := []struct {
		name      string
		selection string
		wantErr   bool
	}{
		{"no selection is the gate (allowed)", "", false},
		{"a single module is a subset", root + "/history", false},
		{"two modules are a subset", root + "/history," + root + "/usage", false},
		{"the root itself is refused", root, true},
		{"traversal to the root is refused", root + "/../cli", true},
		{"all modules are refused", strings.Join(allModules, ","), true},
		{"a superset (root + extra) is refused", root + ",../../specs/plans/001-cli-bootstrap-and-config/features/acceptance", true},
		{"a missing path is refused (fail loud)", "../../specs/truth/features/cli/nope", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := guardSelection(tc.selection)
			if tc.wantErr && err == nil {
				t.Fatalf("guardSelection(%q) = nil, want an error", tc.selection)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("guardSelection(%q) = %v, want nil", tc.selection, err)
			}
		})
	}
}
