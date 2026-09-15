package agent

import (
	"testing"
	"time"
)

// T032 — the pure tool-resource-contract resolver.

func TestResolveBoundThreeTiers(t *testing.T) {
	cases := []struct{ param, ceiling, def, want int }{
		{0, 200, 100, 100},   // no param -> default
		{50, 200, 100, 50},   // param wins
		{300, 200, 100, 200}, // param clamped to ceiling
		{-5, 200, 100, 100},  // negative param -> default
		{0, 0, 0, 1},         // floor of 1
	}
	for _, c := range cases {
		if got := resolveBound(c.param, c.ceiling, c.def); got != c.want {
			t.Errorf("resolveBound(%d,%d,%d) = %d; want %d", c.param, c.ceiling, c.def, got, c.want)
		}
	}
}

func TestClampBytesUsesRawLength(t *testing.T) {
	if got := clampBytes("hello", 10); got != "hello" {
		t.Errorf("under-budget string changed: %q", got)
	}
	got := clampBytes("abcdefghij", 5)
	if len(got) <= 5 || got[:5] != "abcde" {
		t.Errorf("clampBytes = %q; want prefix abcde plus a marker", got)
	}
}

func TestResolveTimeoutThreeTiers(t *testing.T) {
	if got := resolveTimeout(0, 30*time.Second); got != 30*time.Second {
		t.Errorf("default = %v; want 30s", got)
	}
	if got := resolveTimeout(5*time.Second, 30*time.Second); got != 5*time.Second {
		t.Errorf("param = %v; want 5s", got)
	}
	if got := resolveTimeout(9999*time.Second, 30*time.Second); got != TimeoutCeiling {
		t.Errorf("over-ceiling = %v; want %v", got, TimeoutCeiling)
	}
}

func TestCallByteBudgetByteConversion(t *testing.T) {
	a := &AgentLoop{EffectiveBudget: 1000000}
	if got := a.callByteBudget(`{}`); got != 250000*BytesPerToken {
		t.Errorf("default byte budget = %d; want %d", got, 250000*BytesPerToken)
	}
	if got := a.callByteBudget(`{"max_output_tokens":100}`); got != 100*BytesPerToken {
		t.Errorf("param byte budget = %d; want %d", got, 100*BytesPerToken)
	}
	// over-ceiling is clamped to effectiveBudget/2.
	if got := a.callByteBudget(`{"max_output_tokens":999999999}`); got != 500000*BytesPerToken {
		t.Errorf("clamped byte budget = %d; want %d", got, 500000*BytesPerToken)
	}
}
