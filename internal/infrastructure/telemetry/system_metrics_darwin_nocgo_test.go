//go:build darwin && !cgo

package telemetry

import (
	"testing"
	"time"
)

// F-059-1 (cgo-less leg) — a seeded-delta pin so the cgo-less CPU path has a
// falsifiability carrier. The round-019 stub (and a runtime/metrics attempt)
// returns `0` and would fail this: the seeded delta forces a measurable figure.
//
// R-059-4: the seed is `lastCPUNs: 0` with a 1 s wall window, so the pin relies
// on the test process having consumed ≥ 1 ns of CPU since start — true in
// practice (the runtime itself burns CPU before any test runs). A future
// zero-CPU-count context should read a negative result here as this design
// assumption, not as a flake.
func TestDarwinNoCGoCPUIsPopulated(t *testing.T) {
	p := &darwinNoCGoMetricsProvider{haveLast: true, lastCPUNs: 0, lastWall: time.Now().Add(-time.Second)}
	if cpu, _ := p.Sample(); cpu <= 0 {
		t.Fatalf("darwin cgo-less CPU = %v, want > 0", cpu)
	}
}
