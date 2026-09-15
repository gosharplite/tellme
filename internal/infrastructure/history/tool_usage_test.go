package history

import (
	"os"
	"path/filepath"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// toolStore returns a store over a fresh temp user home.
func toolStore(t *testing.T) (*ToolUsageStore, string) {
	t.Helper()
	home := t.TempDir()
	return NewToolUsageStore(func() (string, error) { return home, nil }), home
}

// TestToolUsageStoreLazyCreationAndRoundTrip: the dir/file are created on the
// first Record (not at construction), and the record round-trips.
func TestToolUsageStoreLazyCreationAndRoundTrip(t *testing.T) {
	s, home := toolStore(t)
	if _, err := os.Stat(filepath.Join(home, ".tellme")); !os.IsNotExist(err) {
		t.Fatalf("~/.tellme must not exist before the first Record (lazy creation)")
	}
	if err := s.Record("read_files", domainhistory.ToolOutcomeOK); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := s.Record("read_files", domainhistory.ToolOutcomeError); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := s.Record("list_files", domainhistory.ToolOutcomeTimeout); err != nil {
		t.Fatalf("Record: %v", err)
	}
	recs, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("Load = %d records, want 3", len(recs))
	}
	if recs[0].Tool != "read_files" || recs[0].Outcome != domainhistory.ToolOutcomeOK || recs[0].Timestamp == "" {
		t.Fatalf("record 0 = %+v, want read_files/ok with a timestamp", recs[0])
	}
	if recs[2].Tool != "list_files" || recs[2].Outcome != domainhistory.ToolOutcomeTimeout {
		t.Fatalf("record 2 = %+v, want list_files/timeout", recs[2])
	}
}

// TestToolUsageStoreAggregate: the single-pass fold tallies per tool/outcome.
func TestToolUsageStoreAggregate(t *testing.T) {
	s, _ := toolStore(t)
	_ = s.Record("read_files", domainhistory.ToolOutcomeOK)
	_ = s.Record("read_files", domainhistory.ToolOutcomeOK)
	_ = s.Record("read_files", domainhistory.ToolOutcomeError)
	_ = s.Record("get_tree", domainhistory.ToolOutcomeTimeout)

	counts, err := s.Aggregate()
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}
	if c := counts["read_files"]; c.OK != 2 || c.Error != 1 || c.Total() != 3 {
		t.Errorf("read_files = %+v, want ok 2 error 1 total 3", c)
	}
	if c := counts["get_tree"]; c.Timeout != 1 || c.Total() != 1 {
		t.Errorf("get_tree = %+v, want timeout 1", c)
	}
	if _, ok := counts["list_files"]; ok {
		t.Errorf("a zero-use tool must not appear in the aggregate map")
	}
}

// TestToolUsageStoreAggregateEmpty: a missing log is an empty aggregate (not an error).
func TestToolUsageStoreAggregateEmpty(t *testing.T) {
	s, _ := toolStore(t)
	counts, err := s.Aggregate()
	if err != nil || len(counts) != 0 {
		t.Fatalf("Aggregate = (%v, %v), want an empty map and nil error", counts, err)
	}
}

// TestToolUsageStoreAggregateSkipsMalformedLines: a torn/malformed line (and an
// unknown outcome) is skipped; the report never fails on log content.
func TestToolUsageStoreAggregateSkipsMalformedLines(t *testing.T) {
	s, home := toolStore(t)
	dir := filepath.Join(home, ".tellme")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"timestamp":"t","tool":"read_files","outcome":"ok"}` + "\n" +
		`{not valid json` + "\n" +
		`{"timestamp":"t","tool":"read_files","outcome":"bogus"}` + "\n" +
		`{"timestamp":"t","tool":"list_files","outcome":"error"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "tools-count.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	counts, err := s.Aggregate()
	if err != nil {
		t.Fatalf("Aggregate must not fail on malformed content: %v", err)
	}
	if counts["read_files"].OK != 1 || counts["read_files"].Total() != 1 {
		t.Errorf("read_files = %+v, want a single ok (malformed/bogus skipped)", counts["read_files"])
	}
	if counts["list_files"].Error != 1 {
		t.Errorf("list_files = %+v, want error 1", counts["list_files"])
	}
}

// TestToolUsageStoreUnresolvableHomeIsBestEffort: an unresolvable home makes
// Record fail (the caller swallows it) and Aggregate empty — never a crash.
func TestToolUsageStoreUnresolvableHomeIsBestEffort(t *testing.T) {
	s := NewToolUsageStore(func() (string, error) { return "", nil })
	if err := s.Record("read_files", domainhistory.ToolOutcomeOK); err == nil {
		t.Fatalf("Record with an unresolvable home must return an error (best-effort: the loop swallows it)")
	}
	counts, err := s.Aggregate()
	if err != nil || len(counts) != 0 {
		t.Fatalf("Aggregate = (%v, %v), want an empty map and nil error", counts, err)
	}
}

// TestToolUsageStoreAggregateReadError pins the round-026 implementation-review
// C/D behaviour: a GENUINE open failure is RETURNED (not silently emptied), so
// "unreadable" is distinguishable from "never used". It arranges an ENOTDIR
// failure (the log's parent `~/.tellme` is a regular file) — deterministic and
// independent of file permissions / the test user.
func TestToolUsageStoreAggregateReadError(t *testing.T) {
	s, home := toolStore(t)
	if err := os.WriteFile(filepath.Join(home, ".tellme"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("arrange the ENOTDIR failure: %v", err)
	}
	counts, err := s.Aggregate()
	if err == nil {
		t.Fatalf("Aggregate must return a genuine read error, got nil (which would read as never-used)")
	}
	if len(counts) != 0 {
		t.Errorf("counts = %v, want empty on a read error", counts)
	}
}
