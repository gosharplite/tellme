package telemetry

import (
	"math"
	"testing"
)

// Round-019 machine-wide sampler unit tests (T014), driven by injected readings
// (no real /proc access). The parser tables pin the host CPU (Δ of Σcpu − idle)
// and memory percent computations.

const sampleProcStat = `cpu  100 0 100 700 0 0 0 0 0 0
cpu0 50 0 50 350 0 0 0 0 0 0
cpu1 50 0 50 350 0 0 0 0 0 0
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
	// total = user+nice+system+idle+iowait+irq+softirq+steal = 100+0+100+700 = 900
	// idle  = idle+iowait = 700
	if got.total != 900 || got.idle != 700 {
		t.Fatalf("ticks = {total:%d idle:%d}, want {900, 700}", got.total, got.idle)
	}
}

func TestCPUPercentFromDelta(t *testing.T) {
	// A single reading (prev zero) → the sample's own busy fraction.
	cur := procCPUTicks{total: 900, idle: 700}
	if got := cpuPercentFromDelta(procCPUTicks{}, cur); !approx(got, 200.0/900.0*100) {
		t.Errorf("cpuPercentFromDelta(zero→cur) = %v, want %v", got, 200.0/900.0*100)
	}
	// A delta over a second reading: Δtotal=100, Δidle=40 → 60% busy.
	prev := procCPUTicks{total: 900, idle: 700}
	next := procCPUTicks{total: 1000, idle: 740}
	if got := cpuPercentFromDelta(prev, next); !approx(got, 60) {
		t.Errorf("cpuPercentFromDelta(prev→next) = %v, want 60", got)
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
