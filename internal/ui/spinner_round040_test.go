package ui

import (
	"strings"
	"testing"
	"time"
)

// Round 040 (issue #83) — the WS-B dual-timer arithmetic under an injected clock
// (ADR 0009 D1/D2). The T006 pin: the TOTAL figure counts from the turn's
// prompt-capture epoch and NEVER resets; the SECOND figure (the current model
// call's elapsed) RESETS at each OnInferenceStart.

// TestSpinnerDualTimerArithmetic pins the two-figure arithmetic (per-AI-endpoint
// call reset) under an injected `now` + ticker (no time.Sleep).
func TestSpinnerDualTimerArithmetic(t *testing.T) {
	w := newSignalWriter()
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	nowV := base
	s := NewSpinner(w, "m", base, nil, nil)
	tick := make(chan time.Time, 8)
	s.newTicker = func() (<-chan time.Time, func()) { return tick, func() {} }
	s.now = func() time.Time { return nowV }

	// First call begins at the prompt epoch → both figures are 0.
	s.OnInferenceStart()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (0s 0s)") {
		t.Fatalf("first frame = %q, want the dual (0s 0s)", w.String())
	}

	// A later render within the same call → both figures grow together.
	nowV = base.Add(3 * time.Second)
	tick <- nowV
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (3s 3s)") {
		t.Fatalf("later render = %q, want the dual (3s 3s)", w.String())
	}

	// A SECOND OnInferenceStart (a new AI-endpoint call) → the total GROWS (5s) and
	// the second figure RESETS (0s) — the per-call boundary (QB1).
	nowV = base.Add(5 * time.Second)
	s.OnInferenceStart()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (5s 0s)") {
		t.Fatalf("second call = %q, want the total grown (5s) with the second figure reset (0s)", w.String())
	}

	// A further render within the second call → the second figure counts within its
	// own call (2s), the total keeps growing (7s) and never decreases.
	nowV = base.Add(7 * time.Second)
	tick <- nowV
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (7s 2s)") {
		t.Fatalf("second call later render = %q, want (7s 2s)", w.String())
	}
	s.Stop()
}

// TestSpinnerRowAwareClearOnLongDualFrame re-witnesses the round-025 row-aware
// clear on the LONGER two-figure 3+-digit frame (1234s 567s) at a narrow width
// (QB3 / SC-006): the clear erases every row the wrapped frame occupied.
func TestSpinnerRowAwareClearOnLongDualFrame(t *testing.T) {
	w := newSignalWriter()
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	nowV := base
	s := NewSpinner(w, "m", base, nil, func() int { return 20 })
	tick := make(chan time.Time, 8)
	s.newTicker = func() (<-chan time.Time, func()) { return tick, func() {} }
	s.now = func() time.Time { return nowV }

	// The second call starts at base+667s; a render at base+1234s shows (1234s 567s).
	nowV = base.Add(667 * time.Second)
	s.OnInferenceStart()
	awaitWrite(t, w)
	nowV = base.Add(1234 * time.Second)
	tick <- nowV
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "(1234s 567s)") {
		t.Fatalf("the 3+-digit dual frame was not rendered: %q", w.String())
	}
	// The line is 1 + 16 + 13 = 30 runes → 2 rows at 20 columns → a row-aware clear.
	s.Stop()
	if !strings.HasSuffix(w.String(), eraseRows(2)) {
		t.Fatalf("the long dual frame's clear was not row-aware: %q", w.String())
	}
}
