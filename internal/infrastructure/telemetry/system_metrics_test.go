package telemetry

import (
	"math"
	"testing"
)

// Round-019 machine-wide sampler unit tests (T014), driven by injected readings
// (no real /proc access). The parser tables pin the host CPU (Δ of Σcpu − idle;
// the idle column ONLY — reference parity) and memory percent computations.

const sampleProcStat = `cpu  100 0 100 700 50 0 0 0 0 0
cpu0 50 0 50 350 25 0 0 0 0 0
cpu1 50 0 50 350 25 0 0 0 0 0
`

const sampleMeminfo = `MemTotal:       1000000 kB
MemFree:         100000 kB
MemAvailable:    400000 kB
`

func approx(got, want float64) bool { return math.Abs(got-want) < 0.05 }

func TestParseProcStatHostCPU(t *testing.T) {
	got, ok := parseProcStatHostCPU(sampleProcStat)
	if !ok {
		t.Fatal("parseProcStatHostCPU returned ok=false")
	}
	// total = 100+0+100+700+50+0+0+0+0+0 = 950; idle = idle column ONLY = 700
	// (iowait 50 is deliberately NOT counted — reference parity).
	if got.total != 950 || got.idle != 700 {
		t.Fatalf("ticks = {total:%d idle:%d}, want {950, 700}", got.total, got.idle)
	}
}

func TestCPUPercentFromDelta(t *testing.T) {
	// A single reading (prev zero) → the sample's own busy fraction.
	cur := procCPUTicks{total: 950, idle: 700}
	if got := cpuPercentFromDelta(procCPUTicks{}, cur); !approx(got, 250.0/950.0*100) {
		t.Errorf("cpuPercentFromDelta(zero→cur) = %v, want %v", got, 250.0/950.0*100)
	}
	// A delta over a second reading: Δtotal=100, Δidle=20 → 80% busy.
	prev := procCPUTicks{total: 950, idle: 700}
	next := procCPUTicks{total: 1050, idle: 720}
	if got := cpuPercentFromDelta(prev, next); !approx(got, 80) {
		t.Errorf("cpuPercentFromDelta(prev→next) = %v, want 80", got)
	}
	// A zero delta never divides by zero.
	if got := cpuPercentFromDelta(next, next); got != 0 {
		t.Errorf("cpuPercentFromDelta(no change) = %v, want 0", got)
	}
}

func TestParseProcMeminfo(t *testing.T) {
	got, ok := parseProcMeminfo(sampleMeminfo)
	if !ok {
		t.Fatal("parseProcMeminfo returned ok=false")
	}
	// used = (1000000 - 400000) / 1000000 * 100 = 60%
	if !approx(got, 60) {
		t.Errorf("parseProcMeminfo = %v, want 60", got)
	}
}
