// This file adds the shared global prompt-log adapter to the history package: a
// second local-state artifact, distinct from the per-session history.jsonl store
// (see file_store.go). Round 015 placed the log at the environment-scoped
// `output/` root; round 028 relocates it to the user-global `~/.tellme/` root.

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

// globalPromptLogFile is the shared prompt-log filename.
//
// Round 028: the log lives at the USER-global `<user-home>/.tellme/global_prompts.jsonl`
// (NOT under output/<mode>/), read + written only there. The environment-scoped
// `<home>/output/global_prompts.jsonl` (the round-015 location, shared with
// tell-me-go) is used only as the one-time seed SOURCE (specs/truth/data/data-model.dbml,
// table `prompt_log_entry`).
const globalPromptLogFile = "global_prompts.jsonl"

// GlobalPromptTracker is the file adapter for the shared, append-only global
// prompt log. Since round 028 the log lives at the USER-global
// `<user-home>/.tellme/global_prompts.jsonl`, resolved via the injected
// `userHome` resolver (`deps.Dependencies.UserHomeDir`, supplied by the
// composition root, mirroring `NewToolUsageStore`); the environment-scoped
// `<home>/output/global_prompts.jsonl`
// is used only as the first-use seed source. It implements
// domainhistory.PromptTracker.
type GlobalPromptTracker struct {
	// home is the runtime home (TELL_ME_HOME) — the seed SOURCE root.
	home string
	// userHome resolves the user's home directory — the `~/.tellme` DESTINATION
	// root (round 028). It is injected so CLI unit tests stay hermetic.
	userHome func() (string, error)
	mu       sync.Mutex
	// wg is a RESERVED drain hook for a future asynchronous background compaction
	// (round-015 research Decision 3 / PR #38 review directive ⑤). It is NOT
	// incremented today — the adapter's writes are synchronous — so Close's
	// wg.Wait() is a deliberate no-op drain that keeps the lifecycle contract
	// stable for a future async compactor (PR #59 architect review TD-2).
	wg sync.WaitGroup
}

var _ domainhistory.PromptTracker = (*GlobalPromptTracker)(nil)

// NewGlobalPromptTracker builds the tracker rooted at the runtime home (the seed
// source) plus the injected user-home resolver (the destination). The
// constructor is PURE — it performs no disk I/O; the first-use seed is a
// separate, explicitly-invoked Seed(ctx) (round 028 TD-2).
func NewGlobalPromptTracker(home string, userHome func() (string, error)) *GlobalPromptTracker {
	return &GlobalPromptTracker{home: home, userHome: userHome}
}

// destPath resolves the user-global log path
// (`<user-home>/.tellme/global_prompts.jsonl`). A resolver failure is returned so
// callers can degrade to a no-op (never breaking the interactive prompt).
func (t *GlobalPromptTracker) destPath() (string, error) {
	dir, err := t.userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".tellme", globalPromptLogFile), nil
}

// sourcePath is the environment-scoped seed source
// (`<home>/output/global_prompts.jsonl`).
func (t *GlobalPromptTracker) sourcePath() string {
	return filepath.Join(t.home, "output", globalPromptLogFile)
}

// Seed performs the first-use migration (round-028 FR-005..FR-007): when the
// user-global log is ABSENT, it copies the environment-scoped source VERBATIM
// into it. It is a COPY, not a move (the source is left in place); it NEVER
// overwrites an existing destination; a missing source — or a blank runtime home
// — is a no-op. It is best-effort: any I/O error is returned for the caller to
// ignore, and a failure never aborts the interactive prompt.
func (t *GlobalPromptTracker) Seed(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// A blank runtime home (TELL_ME_HOME unset) has no seed source; skip it so no
	// cwd-relative `output/global_prompts.jsonl` is ever read (round-028 review
	// micro-note).
	if t.home == "" {
		return nil
	}
	dest, err := t.destPath()
	if err != nil {
		return err
	}
	// Never overwrite an existing destination.
	if _, statErr := os.Stat(dest); statErr == nil {
		return nil
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	data, err := os.ReadFile(t.sourcePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no source: nothing to carry over
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	// O_EXCL so a concurrent seeder that won the race is not clobbered.
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil // a concurrent writer created it; leave it
		}
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// Append records one operator prompt (round-015 FR-009). The write is append-only
// (O_APPEND|O_CREATE|O_WRONLY) and carries the frozen shape {timestamp,prompt}.
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

	dest, err := t.destPath()
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
// (round-015 FR-008). A missing file yields an empty slice; an unresolvable user
// home degrades to empty (never an error — the prompt is best-effort).
func (t *GlobalPromptTracker) Recent(ctx context.Context, n int) ([]domainhistory.PromptLogEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dest, err := t.destPath()
	if err != nil {
		return nil, nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	data, err := os.ReadFile(dest)
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
// (round-015 research Decision 3 / PR #38 review directive ⑤). Today the adapter
// runs NO background work — `wg` is a reserved hook (see the struct) — so this is
// a no-op drain that keeps the lifecycle contract stable for a future async
// compactor (PR #59 architect review TD-2).
func (t *GlobalPromptTracker) Close(ctx context.Context) error {
	t.wg.Wait()
	return nil
}
