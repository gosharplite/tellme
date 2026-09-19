//go:build darwin && !cgo

package telemetry

import (
	"runtime"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
	"golang.org/x/sys/unix"
)

// Round-059 macOS sampler (cgo-less leg) — research Decision D2. Without cgo the
// Mach host_statistics64 API is unavailable, so — matching the reference's
// darwin `!cgo` build — the CPU figure is the process's own CPU and the memory
// figure is the sysctl free-page approximation scaled by
// darwinMemoryUsageFactor. This leg is what the `CGO_ENABLED=0` cross-compile
// gate builds; it must report a REAL non-zero figure (the round-019 stub, and a
// first `runtime/metrics`-based attempt, both read 0.0% — see ADR 0029).

// darwinNoCGoMetricsProvider samples the process CPU (getrusage) and the host
// memory (sysctl).
type darwinNoCGoMetricsProvider struct {
	lastCPUNs int64
	lastWall  time.Time
	haveLast  bool
}

// NewSystemMetricsProvider builds the macOS metrics provider (cgo-less build).
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinNoCGoMetricsProvider{} }

// Sample returns the process's own CPU (from the getrusage CPU-time delta since
// the previous sample; 0 on the first call) and the host memory usage
// percentage.
func (p *darwinNoCGoMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	now := time.Now()
	if ns, ok := procCPUNanoseconds(); ok {
		if p.haveLast {
			cpuPercent = agentCPUPercent(ns-p.lastCPUNs, now.Sub(p.lastWall).Seconds(), runtime.NumCPU())
		}
		p.lastCPUNs = ns
		p.lastWall = now
		p.haveLast = true
	}
	return cpuPercent, darwinNoCGoMemPercent()
}

// procCPUNanoseconds reads the process's cumulative CPU time (user + system) via
// getrusage(RUSAGE_SELF) as nanoseconds; ok is false when the call fails.
//
// getrusage is used rather than Go's `runtime/metrics`
// `/cpu/classes/total:cpu-seconds`: that metric is documented as the *available*
// CPU budget (GOMAXPROCS integrated over wall time), is not consumed CPU, and on
// this darwin/cgo-less host is never updated at all (measured `0` under load) —
// so it cannot back a real `CPU: x%` figure. `unix` (golang.org/x/sys) is
// already the round's dependency for the memory leg.
func procCPUNanoseconds() (int64, bool) {
	var ru unix.Rusage
	if err := unix.Getrusage(unix.RUSAGE_SELF, &ru); err != nil {
		return 0, false
	}
	return int64(ru.Utime.Sec)*1e9 + int64(ru.Utime.Usec)*1e3 +
		int64(ru.Stime.Sec)*1e9 + int64(ru.Stime.Usec)*1e3, true
}

// darwinNoCGoMemPercent returns the used-memory percentage from the free +
// speculative + purgeable page counts (the reference's cgo-less approximation).
func darwinNoCGoMemPercent() float64 {
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil || total == 0 {
		return 0
	}
	free, err := unix.SysctlUint32("vm.page_free_count")
	if err != nil {
		return 0
	}
	var speculative, purgeable uint64
	if v, err := unix.SysctlUint32("vm.page_speculative_count"); err == nil {
		speculative = uint64(v)
	}
	if v, err := unix.SysctlUint32("vm.page_purgeable_count"); err == nil {
		purgeable = uint64(v)
	}
	return darwinMemPercentNoCGO(total, darwinPageSize(), uint64(free), speculative, purgeable)
}
