// Package metrics defines the machine-wide system-metrics port (round-019
// research Decision 5). It is the domain-facing abstraction the spinner's
// resource segment reads; the concrete POSIX samplers live in
// internal/infrastructure/telemetry behind this seam, so internal/ui stays a
// pure formatter package.
package metrics

// SystemMetricsProvider reports the machine's current resource usage.
type SystemMetricsProvider interface {
	// Sample returns the machine-wide CPU and memory usage as percentages
	// (0–100). CPU is the Δ of (Σcpu − idle) across the host; memory is the
	// host used/total percent.
	Sample() (cpuPercent, memPercent float64)
}
