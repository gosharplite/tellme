package ui

import (
	"strings"
	"testing"
	"time"
)

// Round-059 throttle pin: the tool-phase resource figures are sampled at most
// once per ResourceSampleInterval (1 s) while the frame keeps its 200 ms
// cadence. A counting metrics double + the injected clock make it deterministic
// (no time.Sleep, no real ticker: the ticker seam is replaced with a dead
// channel so only the direct renderLocked calls draw frames).

type countingMetrics struct {
	calls int
	cpu   float64
	mem   float64
}

func (c *countingMetrics) Sample() (float64, float64) {
	c.calls++
	return c.cpu, c.mem
}

func TestSpinnerResourceSampleThrottle(t *testing.T) {
	epoch := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	cm := &countingMetrics{cpu: 12.5, mem: 34.5}
	var buf strings.Builder
	s := NewSpinner(&buf, "m", epoch, cm, nil)
	// Dead ticker + fixed clock: the test drives frames itself.
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	cur := epoch
	s.now = func() time.Time { return cur }
	defer s.Stop()

	// First tool-phase frame samples once.
	s.OnToolsStart([]string{"read_files"})
	if cm.calls != 1 {
		t.Fatalf("after the first frame: Sample calls = %d, want 1", cm.calls)
	}

	// Frames within the same second reuse the cached pair — no new sample.
	for i := 1; i <= 4; i++ { // 200 ms … 800 ms
		cur = epoch.Add(time.Duration(i) * 200 * time.Millisecond)
		s.mu.Lock()
		s.renderLocked()
		s.mu.Unlock()
		if cm.calls != 1 {
			t.Fatalf("at +%dms: Sample calls = %d, want 1 (throttled)", i*200, cm.calls)
		}
	}
	if got := buf.String(); !strings.Contains(got, "CPU: 12.5%") || !strings.Contains(got, "MEM: 34.5%") {
		t.Errorf("the throttled frames did not render the cached figures: %q", got)
	}

	// Once a full second has elapsed the provider is sampled again.
	cur = epoch.Add(ResourceSampleInterval)
	s.mu.Lock()
	s.renderLocked()
	s.mu.Unlock()
	if cm.calls != 2 {
		t.Fatalf("at +1s: Sample calls = %d, want 2 (one re-sample)", cm.calls)
	}
}

// TestSpinnerResourceSampleDisabledByNilProvider pins that a nil metrics
// provider renders a zero segment without panicking.
func TestSpinnerResourceSampleDisabledByNilProvider(t *testing.T) {
	epoch := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	var buf strings.Builder
	s := NewSpinner(&buf, "m", epoch, nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return epoch }
	defer s.Stop()
	s.OnToolsStart([]string{"read_files"})
	if got := buf.String(); !strings.Contains(got, "CPU: 0.0%") || !strings.Contains(got, "MEM: 0.0%") {
		t.Errorf("nil provider segment = %q, want zeros", got)
	}
}
