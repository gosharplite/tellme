//go:build darwin && !cgo

package telemetry

import "github.com/gosharplite/tellme/internal/domain/metrics"

// Round-019 macOS (nocgo) machine-wide CPU/memory sampler (research Decision 5).
// The memory percentage is read via sysctl without cgo; macOS host CPU requires
// mach host statistics, so the CPU is reported as 0 until a cgo-enabled build
// supplies it, keeping the resource segment rendering.

// darwinNoCgoMetricsProvider samples the machine's CPU and memory.
type darwinNoCgoMetricsProvider struct{}

// NewSystemMetricsProvider builds the macOS (nocgo) machine-wide metrics provider.
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinNoCgoMetricsProvider{} }

// Sample returns the machine-wide CPU and memory usage percentages.
func (p *darwinNoCgoMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	return 0, darwinMemPercent()
}
