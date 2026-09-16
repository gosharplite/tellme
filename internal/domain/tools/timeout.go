package tools

import "time"

// The shared tool-timeout contract (round-024 FR-016), single-sourced here so
// the loop (internal/agent) and any tool adapter — including the MCP adapter —
// resolve a call's effective timeout the same way, with no mirrored ceiling
// constants that can drift (round-032 implementation-review F4).

// TimeoutCeiling is the hard upper bound on a tool call's effective timeout: a
// `timeout` param above it is clamped, never rejected. It is DISTINCT from the
// round-024 TOKEN-bound ceiling (`effectiveBudget ÷ 2`), which bounds
// `max_output_tokens`/bytes rather than time.
const TimeoutCeiling = 7200 * time.Second

// DefaultToolTimeout bounds an individual tool execution when neither a
// per-call `timeout` param nor a tool-declared default supplies one.
const DefaultToolTimeout = 300 * time.Second

// ResolveTimeout resolves a call's effective timeout in three tiers: the call's
// `timeout` param when positive, else the tool's declared default
// (toolDefault), else DefaultToolTimeout; then clamped to TimeoutCeiling.
func ResolveTimeout(param, toolDefault time.Duration) time.Duration {
	t := param
	if t <= 0 {
		t = toolDefault
	}
	if t <= 0 {
		t = DefaultToolTimeout
	}
	if t > TimeoutCeiling {
		t = TimeoutCeiling
	}
	return t
}
