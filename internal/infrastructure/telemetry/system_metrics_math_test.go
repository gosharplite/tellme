package telemetry

import "testing"

// Round-059 metrics-math unit pins (build-tag-free): the process-CPU percentage
// and the two darwin memory definitions. Small totals + pageSize 1 give exact
// expected values, independent of the host.

func TestAgentCPUPercent(t *testing.T) {
	tests := []struct {
		name string
		cpu  int64
		wall float64
		ncpu int
		want float64
	}{
		{"one full core over one second", int64(1e9), 1.0, 1, 100},
		{"half a core over one second on 1 cpu", int64(5e8), 1.0, 1, 50},
		{"half a core over one second on 4 cpus", int64(5e8), 1.0, 4, 12.5},
		{"no cpu delta", 0, 1.0, 4, 0},
		{"non-positive window", int64(5e8), 0, 4, 0},
		{"no core count", int64(5e8), 1.0, 0, 0},
	}
	for _, tt := range tests {
		if got := agentCPUPercent(tt.cpu, tt.wall, tt.ncpu); !approx(got, tt.want) {
			t.Errorf("%s: agentCPUPercent = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestDarwinMemPercentCGO(t *testing.T) {
	// (active+wired+compressor) = 300 pages × pageSize 1 over 1000 bytes = 30%.
	if got := darwinMemPercentCGO(1000, 1, 100, 100, 100); !approx(got, 30) {
		t.Errorf("darwinMemPercentCGO = %v, want 30", got)
	}
	// Clamped to 100 when used exceeds total.
	if got := darwinMemPercentCGO(100, 1, 200, 0, 0); got != 100 {
		t.Errorf("darwinMemPercentCGO clamp = %v, want 100", got)
	}
	// No denominator → 0.
	if got := darwinMemPercentCGO(0, 1, 1, 1, 1); got != 0 {
		t.Errorf("darwinMemPercentCGO(no total) = %v, want 0", got)
	}
}

func TestDarwinMemPercentNoCGO(t *testing.T) {
	// available = 400 pages × pageSize 1 over 1000 ⇒ raw used ratio 0.6 ⇒
	// 0.6 × 100 × 0.6 (factor) = 36%.
	if got := darwinMemPercentNoCGO(1000, 1, 400, 0, 0); !approx(got, 36) {
		t.Errorf("darwinMemPercentNoCGO = %v, want 36", got)
	}
	// available clamps to total ⇒ 0 used.
	if got := darwinMemPercentNoCGO(100, 1, 500, 0, 0); got != 0 {
		t.Errorf("darwinMemPercentNoCGO(over-available) = %v, want 0", got)
	}
	// No denominator → 0.
	if got := darwinMemPercentNoCGO(0, 1, 1, 1, 1); got != 0 {
		t.Errorf("darwinMemPercentNoCGO(no total) = %v, want 0", got)
	}
}
