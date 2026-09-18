package metrics

// UsageCounts is the per-call token breakdown shown on the metrics line
// (round-018): the derived miss (M) and the reported cached (H), completion (C),
// and reasoning (Th) tokens. Relocated here from `internal/ui` by round 051
// (R5.5 of #92; ADR 0020).
type UsageCounts struct {
	Miss       int
	Hit        int
	Completion int
	Thinking   int
}
