package llm

// EstimateTokens returns a deterministic, offline estimate of the token size of
// an assembled conversation (round-009 research Decision 1). It is a
// dependency-free heuristic — a fixed bytes-per-token ratio plus a small
// per-message overhead — so the pre-flight payload status needs no network and
// no BPE tokenizer, and the same conversation always yields the same number.
//
// The exact ratio is deliberately NOT a contract term: the DSL pins only that
// the estimate is deterministic (round-009 chat/dsl.md, `the payload status
// measures against a budget of {budget} tokens`).
func EstimateTokens(messages []Message) int {
	const bytesPerToken = 4
	const perMessageOverhead = 4
	total := 0
	for _, m := range messages {
		total += (len(m.Content)+bytesPerToken-1)/bytesPerToken + perMessageOverhead
	}
	return total
}
