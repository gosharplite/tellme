package llm

// Pricing is a model's per-million-token cost rates (round-018 config `MODELS`).
// Relocated here from `internal/ui` by round 051 (R5.5 of #92; ADR 0020) — cost is
// a provider/model concern, and ComputeCost feeds PERSISTENCE (the usage record),
// not only display.
type Pricing struct {
	Hit  float64 // USD per million cached (hit) input tokens
	Miss float64 // USD per million uncached (miss) input tokens
	Comp float64 // USD per million output (completion + thinking) tokens
}

// ComputeCost returns the USD cost of one API call from its token counts and the
// pricing rates: (miss·Miss + hit·Hit + (completion + thinking)·Comp) / 1e6
// (round-018 research Decision 3). Completion and thinking are the DISJOINT
// counters from research D1, billed additively at the completion rate.
func ComputeCost(p Pricing, in CostInputs) float64 {
	return (float64(in.Miss)*p.Miss + float64(in.Hit)*p.Hit + float64(in.Completion+in.Thinking)*p.Comp) / 1_000_000
}

// CostInputs names one call's four disjoint token counters (round 098; ADR 0068).
// They are four adjacent ints in a fixed order with no local witness: the call site
// used to read `ComputeCost(p, 4, 6, 3, 2) // miss 4, hit 6, completion 3,
// thinking 2` — a comment standing in for a name the signature could not carry.
type CostInputs struct {
	Miss       int
	Hit        int
	Completion int
	Thinking   int
}

// HitRate returns the cache-hit percentage H/(M+H), or 0 when the prompt is empty.
func HitRate(hit, miss int) float64 {
	total := hit + miss
	if total <= 0 {
		return 0
	}
	return float64(hit) / float64(total) * 100
}
