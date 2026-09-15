package history

// ToolOutcome is the three-way classification of one EXECUTED agent-tool
// invocation (round 026, `tool_usage_outcome`): `ok` / `error` / `timeout`. It is
// derived from the loop's structural signals only (the tool's returned error and
// the loop-owned per-call deadline) — never from the tool result text.
type ToolOutcome string

const (
	// ToolOutcomeOK — the tool returned a result (including a bounded/truncated
	// one).
	ToolOutcomeOK ToolOutcome = "ok"
	// ToolOutcomeError — the tool returned a non-nil error.
	ToolOutcomeError ToolOutcome = "error"
	// ToolOutcomeTimeout — the tool returned a nil-error result at the loop's
	// per-call deadline (the round-024 FR-018 "stopped at the time limit" path).
	ToolOutcomeTimeout ToolOutcome = "timeout"
)

// ToolUsageRecord is one executed agent-tool invocation's recorded outcome — the
// persisted shape of the user-global tool-usage log (round 026,
// `tool_usage_record`): `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"…"}`.
type ToolUsageRecord struct {
	Timestamp string      `json:"timestamp"`
	Tool      string      `json:"tool"`
	Outcome   ToolOutcome `json:"outcome"`
}

// ToolUsageSink records one executed agent-tool invocation (round 026). The
// implementation stamps the record timestamp and is best-effort: a nil sink is a
// no-op, and a sink error must NEVER break a turn.
type ToolUsageSink interface {
	Record(tool string, outcome ToolOutcome) error
}
