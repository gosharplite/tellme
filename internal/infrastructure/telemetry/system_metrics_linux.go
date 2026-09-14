//go:build linux

package telemetry

import (
	"os"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// Round-019 Linux machine-wide CPU/memory sampler (research Decision 5): reads
// /proc/stat's host `cpu ` line and /proc/meminfo. The machine-wide CPU is the
// delta of (Σcpu − idle) between successive samples.

// linuxMetricsProvider samples the machine's CPU and memory from /proc.
type linuxMetricsProvider struct {
	last     procCPUTicks
	haveLast bool
}

// NewSystemMetricsProvider builds the Linux machine-wide metrics provider.
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &linuxMetricsProvider{} }

// Sample returns the machine-wide CPU (from the tick delta since the previous
// sample; 0 on the first call) and memory usage percentages.
func (p *linuxMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	if cur, ok := readProcStatHostCPU(); ok {
		if p.haveLast {
			cpuPercent = cpuPercentFromDelta(p.last, cur)
		}
		p.last = cur
		p.haveLast = true
	}
	memPercent = readProcMemPercent()
	return cpuPercent, memPercent
}

// readProcStatHostCPU reads and parses the aggregate host CPU line.
func readProcStatHostCPU() (procCPUTicks, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return procCPUTicks{}, false
	}
	return parseProcStatHostCPU(string(data))
}

// readProcMemPercent reads and parses the host used-memory percentage.
func readProcMemPercent() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	pct, ok := parseProcMeminfo(string(data))
	if !ok {
		return 0
	}
	return pct
}
