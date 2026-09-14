//go:build darwin

package telemetry

import (
	"encoding/binary"
	"fmt"
	"syscall"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// Round-019 macOS machine-wide sampler (research Decision 5).
//
// Memory is read via sysctl. macOS host CPU requires mach host statistics
// (host_statistics64), which needs cgo; until that lands the CPU is reported as
// 0.0 so the resource segment still renders (matching only the reference's
// *nocgo* path). This is a recorded forward item — research.md D5 /
// specs/truth/techstack.md (System metrics provider (telemetry)).

// darwinMetricsProvider samples the machine's CPU and memory.
type darwinMetricsProvider struct{}

// NewSystemMetricsProvider builds the macOS machine-wide metrics provider.
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinMetricsProvider{} }

// Sample returns the machine-wide CPU (0.0 pending a cgo mach sampler) and
// memory usage percentages.
func (p *darwinMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	return 0, darwinMemPercent()
}

// darwinMemPercent reads the machine-wide used-memory percentage via sysctl:
// total = hw.memsize; available = (vm.page_free_count + vm.page_speculative_count
// + vm.page_purgeable_count) × hw.pagesize — the reference's available set. The
// speculative/purgeable sysctls are read best-effort (a missing one contributes
// 0). It degrades to 0 when a required value is unavailable.
func darwinMemPercent() float64 {
	total, err := sysctlUint64("hw.memsize")
	if err != nil || total == 0 {
		return 0
	}
	pageSize, err := sysctlUint64("hw.pagesize")
	if err != nil || pageSize == 0 {
		return 0
	}
	availablePages, err := sysctlUint64("vm.page_free_count")
	if err != nil {
		return 0
	}
	for _, name := range []string{"vm.page_speculative_count", "vm.page_purgeable_count"} {
		if v, err := sysctlUint64(name); err == nil {
			availablePages += v
		}
	}
	availableBytes := availablePages * pageSize
	if availableBytes > total {
		availableBytes = total
	}
	return float64(total-availableBytes) / float64(total) * 100
}

// sysctlUint64 reads an integer sysctl by name. syscall.SysctlUint64 is not
// defined on darwin, so the raw value bytes from syscall.Sysctl are decoded
// natively (little-endian on darwin's supported architectures; a 4-byte value
// is zero-extended).
func sysctlUint64(name string) (uint64, error) {
	raw, err := syscall.Sysctl(name)
	if err != nil {
		return 0, err
	}
	b := []byte(raw)
	switch {
	case len(b) >= 8:
		return binary.LittleEndian.Uint64(b[:8]), nil
	case len(b) >= 4:
		return uint64(binary.LittleEndian.Uint32(b[:4])), nil
	default:
		return 0, fmt.Errorf("sysctl %s: unexpected value size %d", name, len(b))
	}
}
