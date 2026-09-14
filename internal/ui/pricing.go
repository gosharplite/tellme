package ui

// Pricing is a model's per-million-token cost rates (round-018 config `MODELS`).
type Pricing struct {
	Hit  float64 // USD per million cached (hit) input tokens
	Miss float64 // USD per million uncached (miss) input tokens
	Comp float64 // USD per million output (completion + thinking) tokens
}

// ComputeCost returns the USD cost of one API call from its token counts and the
// pricing rates: (miss·Miss + hit·Hit + (completion + thinking)·Comp) / 1e6
// (round-018 research Decision 3). Completion and thinking are the DISJOINT
// counters from research D1, billed additively at the completion rate.
func ComputeCost(p Pricing, miss, hit, completion, thinking int) float64 {
	return (float64(miss)*p.Miss + float64(hit)*p.Hit + float64(completion+thinking)*p.Comp) / 1_000_000
}

// HitRate returns the cache-hit percentage H/(M+H), or 0 when the prompt is empty.
func HitRate(hit, miss int) float64 {
	total := hit + miss
	if total <= 0 {
		return 0
	}
	return float64(hit) / float64(total) * 100
}
