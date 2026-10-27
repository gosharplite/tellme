package telemetry

// Round-059 platform-agnostic metrics math. Kept build-tag-free (like
// system_metrics_parse.go) so the formulas pin on any host: the CPU-tick delta
// is shared by the linux adapter and the darwin cgo leg, and the two darwin
// memory definitions are pinned here rather than inside the build-tagged
// adapters (which cannot run everywhere).

// agentCPUPercent converts a process-CPU delta to a percentage of one core's
// capacity over the wall-clock window, per runtime.NumCPU() —
// `(ΔcpuSeconds / Δwall) * 100 / NumCPU`. On the darwin `!cgo` leg it is fed by
// `getrusage` (see system_metrics_darwin_nocgo.go); it is a machine-fraction
// normalisation (ADR 0029 D1a/D7), not the reference's cgo-less
// `runtime/metrics` (available-CPU) figure. Returns 0 for a non-positive window
// or core count.
func agentCPUPercent(cpuNs int64, wallSeconds float64, ncpu int) float64 {
	if wallSeconds <= 0 || ncpu <= 0 || cpuNs <= 0 {
		return 0
	}
	return (float64(cpuNs) / 1e9) / wallSeconds * 100.0 / float64(ncpu)
}

// darwinMemoryUsageFactor scales the raw "used" ratio to approximate the cgo
// definition of used memory (active + wired + compressor) from the free-page
// view available without cgo. It is the reference's
// `system_metrics_darwin_nocgo.go` factor.
const darwinMemoryUsageFactor = 0.6

// darwinMemPercentCGO returns the darwin used-memory percentage from the Mach
// VM statistics — `(active + wired + compressor) × pagesize / total`
// (the reference's cgo `GetMemoryPercent`). It returns 0 when a denominator is
// unavailable and clamps to [0, 100].
func darwinMemPercentCGO(in darwinMachVM) float64 {
	if in.totalBytes == 0 || in.pageSize == 0 {
		return 0
	}
	used := (in.active + in.wired + in.compressor) * in.pageSize
	return clampPercent(100.0 * float64(used) / float64(in.totalBytes))
}

// darwinMachVM names the Mach VM page counts behind the CGO memory view (round
// 098; ADR 0068). Five adjacent uint64s meant five different things and swapped
// silently.
type darwinMachVM struct {
	totalBytes uint64
	pageSize   uint64
	active     uint64
	wired      uint64
	compressor uint64
}

// darwinMemPercentNoCGO returns the darwin used-memory percentage from the
// free-page view — `(total − (free+speculative+purgeable)×pagesize) / total`,
// scaled by darwinMemoryUsageFactor (the reference's cgo-less approximation).
// It returns 0 when a denominator is unavailable and clamps to [0, 100].
func darwinMemPercentNoCGO(in darwinSysctlVM) float64 {
	if in.totalBytes == 0 || in.pageSize == 0 {
		return 0
	}
	available := (in.free + in.speculative + in.purgeable) * in.pageSize
	if available > in.totalBytes {
		available = in.totalBytes
	}
	rawUsedRatio := float64(in.totalBytes-available) / float64(in.totalBytes)
	return clampPercent(rawUsedRatio * 100.0 * darwinMemoryUsageFactor)
}

// darwinSysctlVM names the sysctl page counts behind the cgo-less memory view
// (round 098; ADR 0068). It is a distinct struct from darwinMachVM because the
// page categories are different — the earlier positional shape made them look
// interchangeable.
type darwinSysctlVM struct {
	totalBytes  uint64
	pageSize    uint64
	free        uint64
	speculative uint64
	purgeable   uint64
}

// clampPercent bounds a percentage to [0, 100].
func clampPercent(p float64) float64 {
	switch {
	case p < 0:
		return 0
	case p > 100:
		return 100
	default:
		return p
	}
}
