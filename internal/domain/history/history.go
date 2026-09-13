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
	Steps  []Step `json:"steps,omitempty"`
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
