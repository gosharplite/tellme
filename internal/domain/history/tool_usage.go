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

// ToolUsageCounts is a tool's tally by outcome (round 026) — the domain shape the
// streaming reader returns per tool and the report renders.
type ToolUsageCounts struct {
	OK      int
	Error   int
	Timeout int
}

// Total is the tool's total invocation count.
func (c ToolUsageCounts) Total() int { return c.OK + c.Error + c.Timeout }

// ToolUsageReader streams the user-global tool-usage log into per-tool counts
// (round 026). The offline `--tool-usage` report consumes it; the agent loop
// consumes only the write side (ToolUsageSink). Keeping the read capability on a
// domain interface lets the CLI stay free of the concrete adapter (PR #57
// principal-architect review — composition-root port asymmetry).
type ToolUsageReader interface {
	// Aggregate streams the log in ONE pass into per-tool counts — O(tools)
	// memory, never materialising every record. A missing log is an empty map; a
	// malformed/torn line is skipped best-effort; a genuine read failure is
	// returned so the caller can diagnose it ("unreadable" ≠ "never used").
	Aggregate() (map[string]ToolUsageCounts, error)
}

// ToolUsageStore is the full read+write tool-usage port: the loop's write sink
// and the report's streaming reader.
type ToolUsageStore interface {
	ToolUsageSink
	ToolUsageReader
}
