//go:build darwin

package telemetry

import (
	"syscall"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// Round-019 macOS machine-wide sampler (research Decision 5).
//
// Memory is read via sysctl. macOS host CPU requires mach host statistics
// (host_statistics64), which needs cgo; until that lands the CPU is reported as
// 0.0 so the resource segment still renders. This is a recorded forward item —
// research.md D5 / specs/truth/techstack.md (System metrics provider (telemetry)).

// darwinMetricsProvider samples the machine's CPU and memory.
type darwinMetricsProvider struct{}

// NewSystemMetricsProvider builds the macOS machine-wide metrics provider.
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinMetricsProvider{} }

// Sample returns the machine-wide CPU (0.0 pending a cgo mach sampler) and
// memory usage percentages.
func (p *darwinMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	return 0, darwinMemPercent()
}

// darwinMemPercent reads the machine-wide used-memory percentage via sysctl
// (hw.memsize total, hw.pagesize × vm.page_free_count free). It degrades to 0
// when a value is unavailable.
func darwinMemPercent() float64 {
	total, err := syscall.SysctlUint64("hw.memsize")
	if err != nil || total == 0 {
		return 0
	}
	pageSize, err := syscall.SysctlUint64("hw.pagesize")
	if err != nil || pageSize == 0 {
		return 0
	}
	freePages, err := syscall.SysctlUint64("vm.page_free_count")
	if err != nil {
		return 0
	}
	freeBytes := freePages * pageSize
	if freeBytes > total {
		freeBytes = total
	}
	return float64(total-freeBytes) / float64(total) * 100
}
