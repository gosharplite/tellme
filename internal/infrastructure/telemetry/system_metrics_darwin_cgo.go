//go:build darwin && cgo

package telemetry

// Round-059 macOS machine-wide sampler (cgo leg) — research Decision D2. macOS
// host CPU requires the Mach host_statistics64 API (host_statistics64 needs
// cgo); with cgo available we sample the machine-wide tick deltas exactly like
// the reference's default build, so `[CPU: x% | MEM: y%]` reports the host, not
// the process. The cgo-less sibling (system_metrics_darwin_nocgo.go) covers the
// `CGO_ENABLED=0` cross-compile leg with a non-zero agent-CPU fallback.

/*
#include <mach/mach.h>
#include <mach/mach_host.h>
*/
import "C"

import (
	"unsafe"

	"github.com/gosharplite/tellme/internal/domain/metrics"
	"golang.org/x/sys/unix"
)

// darwinCGOMetricsProvider samples the machine's CPU (Mach tick deltas) and
// memory (Mach VM statistics).
type darwinCGOMetricsProvider struct {
	last     procCPUTicks
	haveLast bool
}

// NewSystemMetricsProvider builds the macOS machine-wide metrics provider (cgo
// build).
func NewSystemMetricsProvider() metrics.SystemMetricsProvider { return &darwinCGOMetricsProvider{} }

// Sample returns the machine-wide CPU (from the Mach tick delta since the
// previous sample; 0 on the first call) and memory usage percentages.
func (p *darwinCGOMetricsProvider) Sample() (cpuPercent, memPercent float64) {
	if total, idle, ok := machCPUTicks(); ok {
		cur := procCPUTicks{total: total, idle: idle}
		if p.haveLast {
			cpuPercent = cpuPercentFromDelta(p.last, cur)
		}
		p.last = cur
		p.haveLast = true
	}
	return cpuPercent, darwinCGOMemPercent()
}

// machCPUTicks reads the aggregate host CPU tick counters via
// host_statistics64(HOST_CPU_LOAD_INFO); ok is false when the Mach call fails.
func machCPUTicks() (total, idle uint64, ok bool) {
	var cpuInfo C.host_cpu_load_info_data_t
	count := C.mach_msg_type_number_t(C.HOST_CPU_LOAD_INFO_COUNT)
	host := C.mach_host_self()
	// architect-acceptance: Darwin-only Mach API (no cgo-free equivalent).
	if ret := C.host_statistics64(host, C.HOST_CPU_LOAD_INFO, C.host_info64_t(unsafe.Pointer(&cpuInfo)), &count); ret != C.KERN_SUCCESS {
		return 0, 0, false
	}
	user := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_USER])
	system := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_SYSTEM])
	idleTicks := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_IDLE])
	nice := uint64(cpuInfo.cpu_ticks[C.CPU_STATE_NICE])
	return user + nice + system + idleTicks, idleTicks, true
}

// darwinCGOMemPercent returns the used-memory percentage as
// (active + wire + compressor) / hw.memsize (the reference's cgo definition).
func darwinCGOMemPercent() float64 {
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil || total == 0 {
		return 0
	}
	var vmStats C.vm_statistics64_data_t
	count := C.mach_msg_type_number_t(C.HOST_VM_INFO64_COUNT)
	host := C.mach_host_self()
	if ret := C.host_statistics64(host, C.HOST_VM_INFO64, C.host_info64_t(unsafe.Pointer(&vmStats)), &count); ret != C.KERN_SUCCESS {
		return 0
	}
	return darwinMemPercentCGO(total, darwinPageSize(),
		uint64(vmStats.active_count), uint64(vmStats.wire_count), uint64(vmStats.compressor_page_count))
}
