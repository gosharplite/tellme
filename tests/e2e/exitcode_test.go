package e2e

import (
	"testing"

	"github.com/gosharplite/tellme/internal/cli"
)

// TestExitCodesAreDistinct is the host-harness exit-code oracle (FR-014). It
// asserts the four failure codes are pairwise distinct, distinct from success,
// and non-zero; the black-box runs elsewhere compare against these values.
func TestExitCodesAreDistinct(t *testing.T) {
	if cli.Success != 0 {
		t.Fatalf("cli.Success = %d, want 0", cli.Success)
	}

	codes := map[string]int{
		"UsageError":                cli.UsageError,
		"ConfigError":               cli.ConfigError,
		"EnvironmentError":          cli.EnvironmentError,
		"DiagnosticUnresolvedError": cli.DiagnosticUnresolvedError,
	}

	seen := make(map[int]string, len(codes))
	for name, code := range codes {
		if code <= 0 {
			t.Errorf("%s = %d, want > 0 (non-zero)", name, code)
		}
		if code == cli.Success {
			t.Errorf("%s = %d collides with Success", name, code)
		}
		if other, dup := seen[code]; dup {
			t.Errorf("%s and %s share exit code %d", name, other, code)
		}
		seen[code] = name
	}
}

// TestExitCodesMatchPinnedContract pins the numeric values themselves (round
// 002 / FR-005), so the published contract cannot drift silently: editing a
// cli.* constant to a different number fails here even though the black-box
// runs — which compare against the constants — would stay green.
func TestExitCodesMatchPinnedContract(t *testing.T) {
	cases := map[string]struct {
		got  int
		want int
	}{
		"Success":                   {cli.Success, 0},
		"UsageError":                {cli.UsageError, 2},
		"ConfigError":               {cli.ConfigError, 3},
		"EnvironmentError":          {cli.EnvironmentError, 4},
		"DiagnosticUnresolvedError": {cli.DiagnosticUnresolvedError, 5},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("cli.%s = %d, want %d (pinned FR-005)", name, c.got, c.want)
		}
	}
}
