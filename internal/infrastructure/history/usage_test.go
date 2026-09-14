package history

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

func sampleRecs() []domainhistory.UsageRecord {
	return []domainhistory.UsageRecord{
		{Timestamp: "t1", Provider: "p", Model: "m", CachedTokens: 6, PromptTokens: 10, ResponseTokens: 3, TotalTokens: 15, ThinkingTokens: 2, Cost: 0.0001},
		{Timestamp: "t2", Provider: "p", Model: "m", CachedTokens: 0, PromptTokens: 5, ResponseTokens: 7, TotalTokens: 12, ThinkingTokens: 0, Cost: 0.0002},
	}
}

func TestUsageStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)

	if recs, err := s.Load(); err != nil || len(recs) != 0 {
		t.Fatalf("empty log: recs=%v err=%v", recs, err)
	}

	recs := sampleRecs()
	if err := s.Append(recs[0]); err != nil {
		t.Fatalf("append r1: %v", err)
	}
	if err := s.Append(recs[1]); err != nil {
		t.Fatalf("append r2: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 2 || got[0] != recs[0] || got[1] != recs[1] {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if got[0].TotalTokens != got[0].PromptTokens+got[0].ResponseTokens+got[0].ThinkingTokens {
		t.Errorf("total tokens must be prompt + response + thinking: %+v", got[0])
	}
}

func TestUsageStoreBatchAndTotals(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)

	if sum, err := s.Totals(); err != nil || sum != (domainhistory.UsageSummary{}) {
		t.Fatalf("empty totals = %+v err=%v, want the zero summary", sum, err)
	}

	recs := sampleRecs()
	if err := s.AppendBatch(recs); err != nil {
		t.Fatalf("append batch: %v", err)
	}
	sum, err := s.Totals()
	if err != nil {
		t.Fatalf("totals: %v", err)
	}
	// miss = (10-6)+(5-0)=9 · hit = 6+0=6 · out = (3+2)+(7+0)=12 · cost = 0.0003
	if sum.Miss != 9 || sum.Hit != 6 || sum.Out != 12 {
		t.Errorf("totals = %+v, want miss 9 hit 6 out 12", sum)
	}
	if math.Abs(sum.Cost-0.0003) > 1e-12 {
		t.Errorf("total cost = %v, want 0.0003", sum.Cost)
	}
}

func TestUsageStoreTotalsRecoversFromMissingSummary(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)
	if err := s.AppendBatch(sampleRecs()); err != nil {
		t.Fatalf("append: %v", err)
	}
	// Drop the roll-up; Totals() must recompute it from the log (review #1).
	if err := os.Remove(filepath.Join(dir, "tokens.summary.json")); err != nil {
		t.Fatalf("remove summary: %v", err)
	}
	sum, err := s.Totals()
	if err != nil || sum.Miss != 9 || sum.Hit != 6 || sum.Out != 12 {
		t.Fatalf("recovered totals = %+v err=%v", sum, err)
	}
}

func TestUsageStoreArchive(t *testing.T) {
	dir := t.TempDir()
	s := NewUsageStore(dir)
	if err := s.AppendBatch(sampleRecs()); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := s.Archive(); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if recs, _ := s.Load(); len(recs) != 0 {
		t.Errorf("active log must be empty after archive, got %d", len(recs))
	}
	if sum, _ := s.Totals(); sum != (domainhistory.UsageSummary{}) {
		t.Errorf("the roll-up must reset on archive, got %+v", sum)
	}
	if _, err := os.Stat(filepath.Join(dir, "tokens.archive.jsonl")); err != nil {
		t.Errorf("the archive file must exist: %v", err)
	}
}
