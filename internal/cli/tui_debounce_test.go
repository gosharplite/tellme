package cli

import (
	"testing"
	"time"
)

// TestTUIDebounceDurationResolution (round-048 review TD-1): the debounce
// resolver is pinned at the unit tier — the hermetic env override wins when it is
// a valid non-negative millisecond count, and otherwise the injected port's
// default applies. The fake's sentinel is distinct from the production value, so
// a mutant that hard-codes a constant reddens this pin.
func TestTUIDebounceDurationResolution(t *testing.T) {
	p := fakePrompter{}
	for _, tc := range []struct {
		name string
		val  string
		want time.Duration
	}{
		{"unset falls back to the port default", "", fakePrompterDebounce},
		{"valid override wins", "42", 42 * time.Millisecond},
		{"zero override wins", "0", 0},
		{"non-numeric falls back", "abc", fakePrompterDebounce},
		{"negative falls back", "-1", fakePrompterDebounce},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TELL_ME_TUI_DEBOUNCE", tc.val)
			if got := tuiDebounceDuration(p); got != tc.want {
				t.Fatalf("tuiDebounceDuration(%q) = %v, want %v", tc.val, got, tc.want)
			}
		})
	}
}
