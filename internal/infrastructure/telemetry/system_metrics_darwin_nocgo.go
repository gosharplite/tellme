//go:build darwin && !cgo

package telemetry

import (
	"runtime"
	runtimemetrics "runtime/metrics"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
	"golang.org/x/sys/unix"
)

// Round-059 macOS sampler (cgo-less leg) — research Decision D2. Without cgo the
// Mach host_statistics64 API is unavailable, so — matching the reference's
// darwin `!cgo` build — the CPU figure is the process's own CPU from Go's
// `runtime/metrics` (non-zero under load, never a hardcoded 0), and the memory
// figure is the sysctl free-page approximation scaled by
// darwinMemoryUsageFactor. This leg is what the `CGO_ENABLED=0` cross-compile
// gate builds.

// darwinNoCGoMetricsProvider samples the process CPU (runtime/metrics) and the
// host memory (sysctl).
type darwinNoCGoMetricsProvider struct {
	lastCPUNs int64
	lastWall  time.Time
	haveLast  bool
}

// NewSystemMetricsProvider builds the macOS metrics provider (cgo-less build).
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinNoCGoMetricsProvider{} }

// Sample returns the process's own CPU (from the runtime/metrics CPU-seconds
// delta since the previous sample; 0 on the first call) and the host memory
// usage percentage.
func (p *darwinNoCGoMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	now := time.Now()
	if ns, ok := runtimeCPUNanoseconds(); ok {
		if p.haveLast {
			cpuPercent = agentCPUPercent(ns-p.lastCPUNs, now.Sub(p.lastWall).Seconds(), runtime.NumCPU())
		}
		p.lastCPUNs = ns
		p.lastWall = now
		p.haveLast = true
	}
	return cpuPercent, darwinNoCGoMemPercent()
}

// runtimeCPUNanoseconds reads the process's cumulative CPU seconds from the Go
// runtime metrics (/cpu/classes/total:cpu-seconds) as nanoseconds; ok is false
// if the metric is unavailable or of an unexpected kind.
func runtimeCPUNanoseconds() (int64, bool) {
	samples := make([]runtimemetrics.Sample, 1)
	samples[0].Name = "/cpu/classes/total:cpu-seconds"
	runtimemetrics.Read(samples)
	if samples[0].Value.Kind() != runtimemetrics.KindFloat64 {
		return 0, false
	}
	return int64(samples[0].Value.Float64() * 1e9), true
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
