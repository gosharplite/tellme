//go:build darwin

package telemetry

import "syscall"

// Round-019 macOS machine-wide sampling helpers (research Decision 5). Shared by
// the cgo and nocgo provider variants.

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
