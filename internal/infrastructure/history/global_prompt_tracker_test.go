package history

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGlobalPromptTrackerAppendAndRecent (round-015 T028): the shared store
// appends lines and reads them back newest-first + deduplicated, in the frozen
// {timestamp,prompt} shape; Close drains cleanly.
func TestGlobalPromptTrackerAppendAndRecent(t *testing.T) {
	home := t.TempDir()
	tr := NewGlobalPromptTracker(home)
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

	// The on-disk line carries the frozen shape {timestamp, prompt}.
	data, err := os.ReadFile(filepath.Join(home, "output", "global_prompts.jsonl"))
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
