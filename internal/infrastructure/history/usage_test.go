package history

import (
	"os"
	"path/filepath"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

func TestUsageStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)

	if recs, err := s.Load(); err != nil || len(recs) != 0 {
		t.Fatalf("empty log: recs=%v err=%v", recs, err)
	}

	r1 := domainhistory.UsageRecord{Timestamp: "t1", Provider: "p", Model: "m", CachedTokens: 6, PromptTokens: 10, ResponseTokens: 3, TotalTokens: 15, ThinkingTokens: 2, Cost: 0.0001}
	r2 := domainhistory.UsageRecord{Timestamp: "t2", Provider: "p", Model: "m", CachedTokens: 0, PromptTokens: 5, ResponseTokens: 7, TotalTokens: 12, ThinkingTokens: 0, Cost: 0.0002}
	if err := s.Append(r1); err != nil {
		t.Fatalf("append r1: %v", err)
	}
	if err := s.Append(r2); err != nil {
		t.Fatalf("append r2: %v", err)
	}

	recs, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d records, want 2", len(recs))
	}
	if recs[0] != r1 || recs[1] != r2 {
		t.Errorf("round-trip mismatch: %+v", recs)
	}
	// total = prompt + response + thinking.
	if recs[0].TotalTokens != recs[0].PromptTokens+recs[0].ResponseTokens+recs[0].ThinkingTokens {
		t.Errorf("total tokens must be prompt + response + thinking: %+v", recs[0])
	}
}

func TestUsageStoreArchive(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)
	if err := s.Append(domainhistory.UsageRecord{Provider: "p", Cost: 0.5}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := s.Archive(); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if recs, _ := s.Load(); len(recs) != 0 {
		t.Errorf("active log must be empty after archive, got %d", len(recs))
	}
	if _, err := os.Stat(filepath.Join(dir, "tokens.archive.jsonl")); err != nil {
		t.Errorf("the archive file must exist: %v", err)
	}
}
