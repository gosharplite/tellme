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

// UsageSummary is the cumulative session roll-up of the usage log (round-018
// review #1): the summed miss / cached (hit) / output tokens and the summed
// cost. It is persisted as a companion `tokens.summary.json` and maintained
// incrementally so the post-turn summary reads it in O(1) instead of re-parsing
// the whole log on every turn.
type UsageSummary struct {
	Miss int     `json:"miss"`
	Hit  int     `json:"hit"`
	Out  int     `json:"out"`
	Cost float64 `json:"cost"`
}

// Add folds one usage record into the summary.
func (s *UsageSummary) Add(r UsageRecord) {
	s.Miss += r.PromptTokens - r.CachedTokens
	s.Hit += r.CachedTokens
	s.Out += r.ResponseTokens + r.ThinkingTokens
	s.Cost += r.Cost
}

// UsageStore is the network-free per-mode usage-log port (round 018): append
// calls' records, read the cumulative session roll-up, load the whole active log
// (for a round-trip), and archive the active log into the archive file (`--new`).
type UsageStore interface {
	// Append writes one API call's usage as a single append-only JSON line.
	Append(UsageRecord) error
	// AppendBatch writes a turn's records in one open/write/sync cycle and folds
	// them into the persisted roll-up (round-018 review #3).
	AppendBatch([]UsageRecord) error
	// Load returns the active log's records in call order. A missing active log
	// is an empty session, not an error.
	Load() ([]UsageRecord, error)
	// Totals returns the cumulative session roll-up. A missing summary is
	// computed once from the active log and persisted (round-018 review #1).
	Totals() (UsageSummary, error)
	// Archive moves the active log into the archive file, clears the active
	// file, and drops the roll-up. A missing active log is a no-op returning nil.
	Archive() error
}
