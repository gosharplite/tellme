package history

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixedUserHome returns a resolver pointing at dir (a hermetic ~/.tellme root).
func fixedUserHome(dir string) func() (string, error) {
	return func() (string, error) { return dir, nil }
}

// writeEnvLog arranges the environment-scoped seed source
// (<home>/output/global_prompts.jsonl).
func writeEnvLog(t *testing.T, home, prompt string) {
	t.Helper()
	p := filepath.Join(home, "output", "global_prompts.jsonl")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"timestamp":"2026-01-01T00:00:00Z","prompt":"` + prompt + `"}` + "\n"
	if err := os.WriteFile(p, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestGlobalPromptTrackerAppendAndRecent (round-015 T028, retargeted round-028
// T004): the shared store appends lines and reads them back newest-first +
// deduplicated, in the frozen {timestamp,prompt} shape, at the USER-global path;
// Close drains cleanly.
func TestGlobalPromptTrackerAppendAndRecent(t *testing.T) {
	home := t.TempDir()     // runtime home (seed source)
	userHome := t.TempDir() // ~/.tellme root (destination)
	tr := NewGlobalPromptTracker(home, fixedUserHome(userHome))
	ctx := context.Background()
	if err := tr.Append(ctx, "deploy to staging"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := tr.Append(ctx, "review the last two commits"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := tr.Recent(ctx, 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(got) != 2 || got[0].Prompt != "review the last two commits" {
		t.Fatalf("Recent = %v, want newest-first (review…, deploy…)", got)
	}

	// The on-disk line carries the frozen shape, at the user-global path.
	data, err := os.ReadFile(filepath.Join(userHome, ".tellme", "global_prompts.jsonl"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), `"prompt":"deploy to staging"`) {
		t.Fatalf("the log line does not carry the frozen shape: %s", data)
	}
	if err := tr.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestGlobalPromptTrackerUsesInjectedResolver (round-028 TD-1): the destination
// is resolved via the injected resolver, not a direct os.UserHomeDir() call.
func TestGlobalPromptTrackerUsesInjectedResolver(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	tr := NewGlobalPromptTracker(home, fixedUserHome(userHome))
	if err := tr.Append(context.Background(), "x"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if _, err := os.Stat(filepath.Join(userHome, ".tellme", "global_prompts.jsonl")); err != nil {
		t.Fatalf("the log was not written under the injected user home: %v", err)
	}
}

// TestGlobalPromptTrackerSeed_CopiesWhenAbsent (round-028 FR-005/006 + RF-3): the
// seed copies the environment-scoped source verbatim into the absent user-global
// file and leaves the source in place. It also pins the "seed fires with no
// submission" contract at the adapter layer — Seed ALONE carries the history
// over, with no Append.
func TestGlobalPromptTrackerSeed_CopiesWhenAbsent(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	writeEnvLog(t, home, "review the last two commits")
	tr := NewGlobalPromptTracker(home, fixedUserHome(userHome))
	if err := tr.Seed(context.Background()); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	got, err := tr.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(got) != 1 || got[0].Prompt != "review the last two commits" {
		t.Fatalf("Recent after Seed = %v, want the carried-over prompt", got)
	}
	if _, err := os.Stat(filepath.Join(home, "output", "global_prompts.jsonl")); err != nil {
		t.Fatalf("the seed source must be left in place: %v", err)
	}
}

// TestGlobalPromptTrackerSeed_NoOverwrite (round-028 FR-006): an existing
// user-global log is not replaced by a carry-over.
func TestGlobalPromptTrackerSeed_NoOverwrite(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	writeEnvLog(t, home, "review the last two commits")
	tr := NewGlobalPromptTracker(home, fixedUserHome(userHome))
	if err := tr.Append(context.Background(), "deploy to staging"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := tr.Seed(context.Background()); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	got, err := tr.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(got) != 1 || got[0].Prompt != "deploy to staging" {
		t.Fatalf("Recent after Seed = %v, want only the pre-existing prompt (no carry-over)", got)
	}
}

// TestGlobalPromptTrackerSeed_MissingSource (round-028 FR-007): a missing source
// starts empty with no error.
func TestGlobalPromptTrackerSeed_MissingSource(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	tr := NewGlobalPromptTracker(home, fixedUserHome(userHome))
	if err := tr.Seed(context.Background()); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	got, err := tr.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Recent = %v, want empty", got)
	}
}

// TestGlobalPromptTrackerSeed_BlankHomeSkips (round-028 review micro-note): a
// blank runtime home (TELL_ME_HOME unset) has no source, so the seed is skipped
// and never reads a cwd-relative output/global_prompts.jsonl.
func TestGlobalPromptTrackerSeed_BlankHomeSkips(t *testing.T) {
	userHome := t.TempDir()
	tr := NewGlobalPromptTracker("", fixedUserHome(userHome))
	if err := tr.Seed(context.Background()); err != nil {
		t.Fatalf("Seed with blank home: %v", err)
	}
	if _, err := os.Stat(filepath.Join(userHome, ".tellme", "global_prompts.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("a blank home must not seed the user-global log: stat err = %v", err)
	}
}

// TestGlobalPromptTracker_UnresolvableHomeDegrades (round-028 FR-005): a resolver
// failure degrades the log to a no-op (Append errors, Recent is empty) — the
// prompt is never aborted.
func TestGlobalPromptTracker_UnresolvableHomeDegrades(t *testing.T) {
	tr := NewGlobalPromptTracker(t.TempDir(), func() (string, error) { return "", os.ErrNotExist })
	if err := tr.Append(context.Background(), "x"); err == nil {
		t.Fatalf("Append with an unresolvable home should error (best-effort at the call site)")
	}
	got, err := tr.Recent(context.Background(), 10)
	if err != nil || got != nil {
		t.Fatalf("Recent with an unresolvable home = (%v, %v), want (nil, nil)", got, err)
	}
}
