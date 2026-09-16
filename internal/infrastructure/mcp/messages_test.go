package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestReachabilityHint pins the safe cause classification for the skip warning
// (round-032 SC-002 review): it returns only a fixed token, never the raw error.
func TestReachabilityHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"deadline sentinel", context.DeadlineExceeded, "timed out"},
		{"canceled sentinel", context.Canceled, "timed out"},
		{"401", errors.New("failed to initialize: 401 Unauthorized"), "401 unauthorized"},
		{"403 uppercase", errors.New("HTTP 403 Forbidden"), "403 forbidden"},
		{"404", errors.New("unexpected status 404"), "404 not found"},
		{"429", errors.New("429 Too Many Requests"), "429 rate limited"},
		{"500", errors.New("server returned 500"), "500 http error"},
		{"deadline text", errors.New("context deadline exceeded"), "timed out"},
		{"connection refused", errors.New("dial tcp 127.0.0.1:9: connect: connection refused"), "connection refused"},
		{"dns", errors.New("dial tcp: lookup nope.example: no such host"), "name resolution failed"},
		{"unknown", errors.New("some opaque failure"), ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reachabilityHint(tt.err); got != tt.want {
				t.Errorf("reachabilityHint(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestUnreachableWarningWithHint pins the operator-facing shape: the hint appends
// to the un-hinted base (which stays a prefix, so existing Contains assertions
// hold), and an unrecognised error yields the base unchanged.
func TestUnreachableWarningWithHint(t *testing.T) {
	t.Parallel()

	base := UnreachableWarning("hf")

	hinted := UnreachableWarningWithHint("hf", errors.New("failed to initialize: 401 Unauthorized"))
	if want := base + " (401 unauthorized)"; hinted != want {
		t.Errorf("hinted warning = %q, want %q", hinted, want)
	}
	if !strings.HasPrefix(hinted, base) {
		t.Errorf("the base sentence must remain a prefix so Contains(base) still holds; got %q", hinted)
	}
	if got := UnreachableWarningWithHint("hf", errors.New("mystery")); got != base {
		t.Errorf("an unrecognised error must yield the un-hinted base; got %q", got)
	}
	if got := UnreachableWarningWithHint("hf", nil); got != base {
		t.Errorf("a nil error must yield the un-hinted base; got %q", got)
	}
}
