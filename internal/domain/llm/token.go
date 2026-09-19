package llm

// bytesPerToken and perMessageOverhead are the deterministic, dependency-free
// heuristic constants shared by the estimation helpers (round-009 research
// Decision 1). The exact ratio is NOT a contract term: only that the estimate is
// deterministic, offline, and computed over the wire payload (round-011).
const (
	bytesPerToken      = 4
	perMessageOverhead = 4
)

// estimateTerm is the shared per-item heuristic term — a fixed bytes-per-token
// ratio plus a small overhead — so the message, persona, and tool-declaration
// terms cannot drift apart (round-011 N-1).
func estimateTerm(s string) int {
	return (len(s)+bytesPerToken-1)/bytesPerToken + perMessageOverhead
}

// estimateMediaTerm approximates the token cost of one attached media part
// (round 062; PR #129 fold F-062-2): its base64 wire expansion (4 bytes per 3)
// over the bytes-per-token ratio, plus the per-part overhead. So the pre-flight
// estimate RESPONDS to media and an image-bearing turn is not under-reported.
func estimateMediaTerm(nBytes int) int {
	base64Len := 4 * ((nBytes + 2) / 3)
	return (base64Len+bytesPerToken-1)/bytesPerToken + perMessageOverhead
}

// EstimateTokens returns a deterministic, offline estimate of the token size of
// an assembled conversation (round-009 research Decision 1). It is a
// dependency-free heuristic — a fixed bytes-per-token ratio plus a small
// per-message overhead — so the pre-flight payload status needs no network and
// no BPE tokenizer, and the same conversation always yields the same number.
//
// The exact ratio is deliberately NOT a contract term: the DSL pins only that
// the estimate is deterministic (round-009 chat/dsl.md, `the payload status
// measures against a budget of {budget} tokens`).
//
// Round 062 (PR #129 fold F-062-2): an attached media part is counted too — its
// base64 wire expansion over the same ratio — so the pre-flight figure (and thus
// the budget view) reflects an image, not only text.
func EstimateTokens(messages []Message) int {
	total := 0
	for _, m := range messages {
		total += estimateTerm(m.Content)
		for _, mp := range m.Media {
			total += estimateMediaTerm(len(mp.Data))
		}
	}
	return total
}

// EstimatePayload returns the wire-faithful pre-flight estimate of a request
// (round-011 research Decision 4): the leading persona message, the tool
// declarations sent in the request, and the conversation messages — the three
// input components the provider's `prompt_tokens` covers. It is deterministic
// and dependency-free; it is NOT required to equal the provider's reported count
// (round-011 research Decision 5), only to be computed over the wired inputs and
// to be responsive to their size.
func EstimatePayload(persona string, tools []ToolDef, messages []Message) int {
	total := EstimateTokens(messages)
	if persona != "" {
		total += estimateTerm(persona)
	}
	for _, t := range tools {
		total += estimateTerm(t.Name)
		total += estimateTerm(t.Description)
		total += estimateTerm(string(t.Parameters))
	}
	return total
}
