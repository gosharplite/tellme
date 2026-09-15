package history

import "context"

// PromptLogEntry is one recorded operator prompt in the shared global prompt log
// (round 015), now the user-global row `{timestamp, prompt}` of
// `~/.tellme/global_prompts.jsonl` (round 028 relocated it out of the runtime
// home; specs/truth/data/data-model.dbml, table `prompt_log_entry`). Both fields
// are strings.
type PromptLogEntry struct {
	// Timestamp is the RFC3339 record time.
	Timestamp string `json:"timestamp"`
	// Prompt is the recorded operator prompt text.
	Prompt string `json:"prompt"`
}

// PromptTracker is the domain port for the shared, append-only global prompt log
// (round 015), relocated to the **user-global** `~/.tellme/` root in round 028.
// It is implemented by internal/infrastructure/history (the file adapter) and
// consumed by the CLI (record on an `-i` submit; seed before the first
// suggestion read) and the suggestion engine (recent prompts).
//
// Lifecycle: the adapter resolves its path via an **injected** user-home resolver
// (the CLI-owned `userHomeDir` seam, so unit tests stay hermetic). The log is
// written only under `-i`; when the user-global file is absent it is seeded once
// with a verbatim copy of the environment-scoped `<TELL_ME_HOME>/output/global_prompts.jsonl`
// (copy, not move; never overwritten) via the adapter's explicit `Seed(ctx)`
// method, invoked once at the composition root. Round 028 intentionally **does
// not** share the log with tell-me-go (which keeps the environment-scoped file).
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
