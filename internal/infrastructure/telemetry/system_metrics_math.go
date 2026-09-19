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
func darwinMemPercentCGO(totalBytes, pageSize uint64, active, wired, compressor uint64) float64 {
	if totalBytes == 0 || pageSize == 0 {
		return 0
	}
	used := (active + wired + compressor) * pageSize
	return clampPercent(100.0 * float64(used) / float64(totalBytes))
}

// darwinMemPercentNoCGO returns the darwin used-memory percentage from the
// free-page view — `(total − (free+speculative+purgeable)×pagesize) / total`,
// scaled by darwinMemoryUsageFactor (the reference's cgo-less approximation).
// It returns 0 when a denominator is unavailable and clamps to [0, 100].
func darwinMemPercentNoCGO(totalBytes, pageSize uint64, free, speculative, purgeable uint64) float64 {
	if totalBytes == 0 || pageSize == 0 {
		return 0
	}
	available := (free + speculative + purgeable) * pageSize
	if available > totalBytes {
		available = totalBytes
	}
	rawUsedRatio := float64(totalBytes-available) / float64(totalBytes)
	return clampPercent(rawUsedRatio * 100.0 * darwinMemoryUsageFactor)
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
