package history

// Step is one tool execution performed during a completed turn (round 008, the
// agent tool loop). It is embedded in the turn's line as an element of the
// ordered Steps array (specs/truth/data/data-model.dbml, `history_step`). No
// per-step id is stored; the adapter synthesises a deterministic tool-call id
// on replay.
type Step struct {
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
	// Signature is the provider's opaque token for this tool call (for example
	// the Gemini 3 `thoughtSignature`), persisted only when the provider
	// supplies one (omitempty keeps a signature-less step byte-identical to the
	// round-008 shape) and replayed verbatim on resume (round 014).
	Signature string `json:"signature,omitempty"`
}

// Entry is one completed exchange in the session history: the operator's prompt
// and the provider's answer, stored verbatim (content, not the rendered form),
// plus — since round 008 — the tool steps the turn performed (empty when the
// turn made no tool calls). It is the persisted record of the session-history
// data truth (specs/truth/data/data-model.dbml, `history_entry`).
type Entry struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
	// Calls is the number of AI-endpoint calls (provider inference rounds) this
	// completed turn made (round 027): 1 for a tool-less turn, 1 + the number of
	// tool rounds otherwise (a provider-internal retry does not add). It is
	// summed across the active session to number the round-017 turn header
	// (Σ calls + 1); a line without it (a legacy or arranged entry) counts as 1.
	// Omitted when zero so a field-less line stays byte-identical.
	Calls int    `json:"calls,omitempty"`
	Steps []Step `json:"steps,omitempty"`
}

// InterruptedTurnAnswer is the synthetic closing answer tellme persists for a
// turn the operator interrupted (SIGINT/SIGTERM) after at least one tool step
// completed (round 079; ADR 0051). It is a STORED record value: the partial turn
// is closed with it so the persisted `history.jsonl` line replays as a valid
// `user … assistant` sequence on both provider families — a partial turn ending
// on a `tool` result would violate role alternation (Gemini/Vertex rides
// `functionResponse` under `role:"user"` and rejects two consecutive `user`
// roles). It is fixed, control-free, single-line, and unmistakably synthetic. A
// `SIGTERM` interruption shares this wording (the cause is not distinguished).
const InterruptedTurnAnswer = "[Turn interrupted by operator via Ctrl+C]"

// TotalCalls is the domain reading of the persisted call counts (round 027): the
// total number of AI-endpoint calls (provider inference rounds) the given
// completed turns made. Each entry contributes its persisted Calls count; an
// entry without one (a legacy or arranged plain line, Calls <= 0) contributes 1.
// It is the basis of the turn-header number (Σ + 1), so the legacy fallback and
// the accumulation live with the record, not the presentation layer.
func TotalCalls(entries []Entry) int {
	total := 0
	for _, e := range entries {
		if e.Calls > 0 {
			total += e.Calls
		} else {
			total++
		}
	}
	return total
}

// Store is the network-free session-history port: load the whole conversation,
// append one completed exchange, and archive the active history into the
// archive file (round-007 research Decisions 1 & 3).
type Store interface {
	// Load returns the active entries in conversation order. A missing active
	// history is an empty conversation, not an error.
	Load() ([]Entry, error)
	// Append writes one completed exchange as a single append-only line.
	Append(Entry) error
	// Archive moves the active history into the archive file and clears the
	// active file. A missing active history is a no-op returning nil.
	Archive() error
}
