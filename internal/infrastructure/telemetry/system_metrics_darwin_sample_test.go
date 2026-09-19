//go:build darwin

package telemetry

import "testing"

// Round-059 darwin sampler pin: the macOS provider reports a real, non-zero
// memory figure (the round-019 stub/mis-decode reported 0.0% on every host) and
// a well-formed CPU figure. This runs on the darwin host and is a falsifiability
// witness for the decode fix (it would be red at round 058's code).

func TestDarwinSampleIsPopulated(t *testing.T) {
	cpu, mem := NewSystemMetricsProvider().Sample()
	if mem <= 0 {
		t.Errorf("darwin MEM = %v, want > 0 (the round-019 decoder reported 0.0)", mem)
	}
	if mem > 100 {
		t.Errorf("darwin MEM = %v, want <= 100", mem)
	}
	if cpu < 0 || cpu > 100 {
		t.Errorf("darwin CPU = %v, want within [0, 100]", cpu)
	}
}
