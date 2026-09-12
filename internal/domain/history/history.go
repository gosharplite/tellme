package history

// Entry is one completed exchange in the session history: the operator's prompt
// and the provider's answer, stored verbatim (content, not the rendered form).
// It is the persisted record of the session-history data truth
// (specs/truth/data/data-model.dbml, `history_entry`).
type Entry struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
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
