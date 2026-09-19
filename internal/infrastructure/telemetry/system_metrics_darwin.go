//go:build darwin

package telemetry

import "golang.org/x/sys/unix"

// Round-059 darwin shared helpers. The macOS CPU/memory sources differ by build
// leg (cgo Mach vs cgo-less sysctl/runtime-metrics) and live in
// system_metrics_darwin_cgo.go / system_metrics_darwin_nocgo.go; this
// darwin-tagged file holds only the pieces both legs share.

// darwinPageSize reads the host page size via syscall (golang.org/x/sys/unix),
// which handles the NUL-trimmed darwin sysctl value correctly — the round-019
// hand-rolled fixed-width decode did not (it read hw.memsize's first four bytes
// and reported 0 on a 16 GiB host).
func darwinPageSize() uint64 { return uint64(unix.Getpagesize()) }
