// This file adds the shared global prompt-log adapter (round 015) to the
// history package: a second local-state artifact, distinct from the per-session
// history.jsonl store (see file_store.go).

package history

import (
	"context"
	"path/filepath"
	"sync"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// globalPromptLogFile is the shared prompt-log filename at the runtime home's
// `output/` ROOT (round 015; NOT under output/<mode>/), shared across
// modes/personas with tell-me-go (specs/truth/data/data-model.dbml,
// table `prompt_log_entry`).
const globalPromptLogFile = "global_prompts.jsonl"

// GlobalPromptTracker is the file adapter for the shared, append-only global
// prompt log at $TELL_ME_HOME/output/global_prompts.jsonl (round 015). It
// implements domainhistory.PromptTracker. The behaviour lands with the Feature
// phase (round-015 T036); this is the landing skeleton carrying the log path,
// the append lock, and the background-drain seam.
type GlobalPromptTracker struct {
	path string
	mu   sync.Mutex
	wg   sync.WaitGroup
}

var _ domainhistory.PromptTracker = (*GlobalPromptTracker)(nil)

// NewGlobalPromptTracker builds the tracker rooted at the runtime home; the log
// lives at <home>/output/global_prompts.jsonl.
func NewGlobalPromptTracker(home string) *GlobalPromptTracker {
	return &GlobalPromptTracker{path: filepath.Join(home, "output", globalPromptLogFile)}
}

// Append records one operator prompt in the shared log (round-015 FR-009). The
// eventual write is append-only (O_APPEND|O_CREATE|O_WRONLY, never truncating
// lines another writer added) and detached/bounded so it never blocks the
// visible prompt. Skeleton: the record serialisation + append lands with T036.
func (t *GlobalPromptTracker) Append(_ context.Context, _ string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_ = t.path
	return nil
}

// Recent returns the newest-first, deduplicated prompts, bounded by n
// (round-015 FR-008). Skeleton: the reverse-chunked scan lands with T036.
func (t *GlobalPromptTracker) Recent(_ context.Context, _ int) ([]domainhistory.PromptLogEntry, error) {
	return nil, nil
}

// Close drains any background writes/compaction before the process exits
// (round-015 research Decision 3 / PR #38 review directive ⑤).
func (t *GlobalPromptTracker) Close(_ context.Context) error {
	t.wg.Wait()
	return nil
}
