//go:build darwin && cgo

package telemetry

import "testing"

// F-059-1 (cgo leg) — a seeded-delta pin so the Mach CPU path has a
// falsifiability carrier (a hardcoded-zero CPU would fail it). The seed is a
// synthetic prior tick pair 1000 total / 100 idle behind the live reading, so
// the computed busy fraction is 90%.
func TestDarwinCGoCPUIsPopulated(t *testing.T) {
	total, idle, ok := machCPUTicks()
	if !ok {
		t.Skip("host_statistics64 unavailable")
	}
	p := &darwinCGOMetricsProvider{haveLast: true, last: procCPUTicks{total: total - 1000, idle: idle - 100}}
	if cpu, _ := p.Sample(); cpu <= 0 {
		t.Fatalf("darwin cgo CPU = %v, want > 0", cpu)
	}
}
