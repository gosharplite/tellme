// This file adds the shared global prompt-log adapter (round 015) to the
// history package: a second local-state artifact, distinct from the per-session
// history.jsonl store (see file_store.go).

package history

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// globalPromptLogFile is the shared prompt-log filename at the runtime home's
// `output/` ROOT (round 015; NOT under output/<mode>/), shared across
// modes/personas with tell-me-go (specs/truth/data/data-model.dbml,
// table `prompt_log_entry`).
const globalPromptLogFile = "global_prompts.jsonl"

// GlobalPromptTracker is the file adapter for the shared, append-only global
// prompt log at $TELL_ME_HOME/output/global_prompts.jsonl (round 015). It
// implements domainhistory.PromptTracker.
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

// Append records one operator prompt (round-015 FR-009). The write is append-only
// (O_APPEND|O_CREATE|O_WRONLY) so a file also written by other personas/modes is
// never corrupted or truncated, and it carries the frozen shape
// {timestamp,prompt} (so lines round-trip byte-for-byte with tell-me-go).
func (t *GlobalPromptTracker) Append(ctx context.Context, prompt string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	rec, err := json.Marshal(domainhistory.PromptLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Prompt:    prompt,
	})
	if err != nil {
		return err
	}
	rec = append(rec, '\n')

	t.mu.Lock()
	defer t.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(t.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(t.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(rec); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// Recent returns the newest-first, deduplicated prompts, bounded by n
// (round-015 FR-008). A missing file yields an empty slice.
func (t *GlobalPromptTracker) Recent(ctx context.Context, n int) ([]domainhistory.PromptLogEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	data, err := os.ReadFile(t.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []domainhistory.PromptLogEntry
	seen := map[string]bool{}
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if n > 0 && len(out) >= n {
			break
		}
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var e domainhistory.PromptLogEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if seen[e.Prompt] {
			continue
		}
		seen[e.Prompt] = true
		out = append(out, e)
	}
	return out, nil
}

// Close drains any background writes/compaction before the process exits
// (round-015 research Decision 3 / PR #38 review directive ⑤).
func (t *GlobalPromptTracker) Close(ctx context.Context) error {
	t.wg.Wait()
	return nil
}
