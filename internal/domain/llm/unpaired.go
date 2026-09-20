package llm

// UnpairedToolCalls returns, in call order, the ids of tool calls that a
// conversation's rounds left UNANSWERED — a call whose round produced no result
// for it by the round boundary (round 068; ADR 0038; surfaces ADR 0037 §Forward
// RF-067-1).
//
// It is the family-neutral, single owner of the round-boundary pairing account:
// the Gemini/Vertex adapter's M < N drop is the producer (its `UnpairedCallIDs`
// delegates here), and the CLI's diagnostic decorator is the consumer. The walk
// mirrors the adapter's round grouping: a round opens on a message carrying
// tool calls, closes on the next such message, a plain-text turn, or the end;
// each result binds to its call by `ToolCallID` (identity), else by the FIFO
// fallback (first unpaired), exactly as the adapter pairs them. A standalone
// media turn is not a boundary.
func UnpairedToolCalls(msgs []Message) []string {
	var p roundPairing
	for _, m := range msgs {
		switch {
		case len(m.ToolCalls) > 0:
			p.open(m.ToolCalls)
		case m.ToolCallID != "" || (m.Role == "tool" && len(m.Media) == 0):
			p.result(m.ToolCallID)
		case len(m.Media) > 0:
			// a standalone media turn is not a round boundary
		default:
			p.close()
		}
	}
	p.close()
	return p.unpaired
}

// roundPairing accumulates one conversation's round-boundary account: the current
// round's call ids and whether each has been answered, plus the residue collected
// at each boundary.
type roundPairing struct {
	ids      []string
	used     []bool
	unpaired []string
}

// open starts a new round (closing the previous one) and records the calls.
func (p *roundPairing) open(calls []ToolCall) {
	p.close()
	for _, tc := range calls {
		p.ids = append(p.ids, tc.ID)
		p.used = append(p.used, false)
	}
}

// result answers a call: by identity (`ToolCallID`), else the FIFO fallback.
func (p *roundPairing) result(resultID string) {
	if i := p.bind(resultID); i >= 0 {
		p.used[i] = true
	}
}

// bind returns the call a result pairs with (first unused identity match, else the
// first unused), or -1 when every call is already answered.
func (p *roundPairing) bind(resultID string) int {
	if resultID != "" {
		for i := range p.ids {
			if !p.used[i] && p.ids[i] == resultID {
				return i
			}
		}
	}
	for i := range p.ids {
		if !p.used[i] {
			return i
		}
	}
	return -1
}

// close flushes the current round: every unanswered call id joins the residue.
func (p *roundPairing) close() {
	for i, id := range p.ids {
		if !p.used[i] {
			p.unpaired = append(p.unpaired, id)
		}
	}
	p.ids, p.used = nil, nil
}
