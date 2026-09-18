package ui

import (
	"strings"
	"testing"
	"time"
)

// Round 045 (R3 of #92) — the yield-policy owner pins (SC-001/SC-003). These are
// UNIT pins: the yield is an ordering/route contract over the diagnostic stream,
// which a flat E2E byte capture cannot witness (a green suite alone is false
// confidence — the round-009 trap). No time.Sleep.

// TestYieldControllerNilIsANoOp pins the gated-off case: a controller over a nil
// spinner reports Enabled()==false and every method is a safe no-op, so callers
// need no nil check of their own.
func TestYieldControllerNilIsANoOp(t *testing.T) {
	yc := NewYieldController(nil)
	if yc.Enabled() {
		t.Fatal("a controller over a nil spinner must report Enabled()==false")
	}
	yc.Yield()
	yc.Restore()
	yc.Admit()
}

// TestYieldControllerOwnsTheYieldMechanism pins that the OWNER is the single
// route to the spinner's yield mechanism: Yield clears, Restore resumes (keeping
// the turn-scoped total), and Admit resumes a live-but-stopped indicator.
func TestYieldControllerOwnsTheYieldMechanism(t *testing.T) {
	w := newSignalWriter()
	base := time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)
	nowV := base
	s := NewSpinner(w, "m", base, nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return nowV }

	yc := NewYieldController(s)
	if !yc.Enabled() {
		t.Fatal("a controller over a live spinner must report Enabled()==true")
	}

	// A live phase label so a resume is not the label-less no-op.
	s.OnInferenceStart()
	awaitWrite(t, w)

	// Yield clears (a synchronous, goroutine-joined clear): the last write ends
	// with the clear control.
	nowV = base.Add(2 * time.Second)
	yc.Yield()
	awaitWrite(t, w)
	if !strings.HasSuffix(w.String(), "\r\x1b[K") {
		t.Fatalf("Yield did not write a trailing clear: %q", w.String())
	}

	// Restore resumes; the total figure continues (7s) rather than resetting.
	nowV = base.Add(7 * time.Second)
	yc.Restore()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (7s 7s)") {
		t.Fatalf("Restore did not resume with the total preserved: %q", w.String())
	}

	// Admit resumes a live-but-stopped indicator with a goroutine-drawn frame
	// (the WS-A path); after Yield it re-appears without waiting a full poll.
	yc.Yield()
	awaitWrite(t, w)
	nowV = base.Add(9 * time.Second)
	yc.Admit()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (9s 9s)") {
		t.Fatalf("Admit did not resume with a drawn frame: %q", w.String())
	}
	s.Stop()
}

// TestSpinnerPortHooksDelegateToOwner pins that the loop-facing port adapters on
// *Spinner route through the owner (the yield policy has ONE implementation):
// YieldIndicator clears and RestoreIndicator resumes, exactly like Yield/Restore.
func TestSpinnerPortHooksDelegateToOwner(t *testing.T) {
	w := newSignalWriter()
	base := time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)
	nowV := base
	s := NewSpinner(w, "m", base, nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return nowV }

	s.OnInferenceStart()
	awaitWrite(t, w)

	nowV = base.Add(3 * time.Second)
	s.YieldIndicator()
	awaitWrite(t, w)
	if !strings.HasSuffix(w.String(), "\r\x1b[K") {
		t.Fatalf("YieldIndicator did not clear: %q", w.String())
	}

	nowV = base.Add(4 * time.Second)
	s.RestoreIndicator()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (4s 4s)") {
		t.Fatalf("RestoreIndicator did not resume: %q", w.String())
	}
	s.Stop()
}
