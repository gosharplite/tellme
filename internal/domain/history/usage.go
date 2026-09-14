package history

// UsageRecord is one API call's recorded usage in the per-mode usage log
// (round-018, `usage_record` in specs/truth/data/data-model.dbml): a single JSON
// object per line, appended to `$TELL_ME_HOME/output/<mode>/tokens.log`.
//
// Field order matches the reference's `tokens.log` shape so the session totals
// are a plain sum. `PromptTokens` INCLUDES the cached portion (miss = prompt −
// cached); `ResponseTokens` is the EXCLUSIVE completion count and `ThinkingTokens`
// the reasoning count (disjoint, `total = prompt + response + thinking`); `Cost`
// is the call cost computed from the config `MODELS` pricing (0 when un-priced).
type UsageRecord struct {
	Timestamp      string  `json:"timestamp"`
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	CachedTokens   int     `json:"cached_tokens"`
	PromptTokens   int     `json:"prompt_tokens"`
	ResponseTokens int     `json:"response_tokens"`
	TotalTokens    int     `json:"total_tokens"`
	ThinkingTokens int     `json:"thinking_tokens"`
	Cost           float64 `json:"cost"`
}

// UsageStore is the network-free per-mode usage-log port (round 018): append one
// call's record, load the whole active log (for the session totals), and archive
// the active log into the archive file (`--new`).
type UsageStore interface {
	// Append writes one API call's usage as a single append-only JSON line.
	Append(UsageRecord) error
	// Load returns the active log's records in call order. A missing active log
	// is an empty session, not an error.
	Load() ([]UsageRecord, error)
	// Archive moves the active log into the archive file and clears the active
	// file. A missing active log is a no-op returning nil.
	Archive() error
}
