//go:build darwin && cgo

package telemetry

import "github.com/gosharplite/tellme/internal/domain/metrics"

// Round-019 macOS (cgo) machine-wide CPU/memory sampler (research Decision 5).
// The memory percentage is read via sysctl; macOS host CPU requires mach host
// statistics, which this cgo variant would provide — until then the CPU is
// reported as 0 so the resource segment still renders.

// darwinMetricsProvider samples the machine's CPU and memory.
type darwinMetricsProvider struct{}

// NewSystemMetricsProvider builds the macOS (cgo) machine-wide metrics provider.
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinMetricsProvider{} }

// Sample returns the machine-wide CPU and memory usage percentages.
func (p *darwinMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	return 0, darwinMemPercent()
}
