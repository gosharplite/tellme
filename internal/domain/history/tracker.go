package history

import "context"

// PromptLogEntry is one recorded operator prompt in the shared global prompt log
// (round 015) — the row `{timestamp, prompt}` of
// `$TELL_ME_HOME/output/global_prompts.jsonl` (specs/truth/data/data-model.dbml,
// table `prompt_log_entry`). Both fields are strings.
type PromptLogEntry struct {
	// Timestamp is the RFC3339 record time.
	Timestamp string
	// Prompt is the recorded operator prompt text.
	Prompt string
}

// PromptTracker is the domain port for the shared, append-only global prompt log
// at the `output/` root of the runtime home (round 015). It is implemented by
// internal/infrastructure/history (the file adapter) and consumed by the CLI
// (record on an `-i` submit) and the suggestion engine (recent prompts). The log
// is shared across modes/personas with tell-me-go, so its record shape must
// round-trip byte-for-byte.
type PromptTracker interface {
	// Append records one operator prompt in the shared log (round-015 FR-009).
	// The write is append-only and must never truncate lines another writer added.
	Append(ctx context.Context, prompt string) error

	// Recent returns the newest-first, deduplicated prompts, bounded by n
	// (round-015 FR-008).
	Recent(ctx context.Context, n int) ([]PromptLogEntry, error)

	// Close drains any background writes/compaction (round-015 research
	// Decision 3 / PR #38 review directive ⑤) before the process exits.
	Close(ctx context.Context) error
}
