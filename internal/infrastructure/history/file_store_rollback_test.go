package history

import (
	"os"
	"path/filepath"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// readActive reads the raw bytes of the active history file.
func readActive(t *testing.T, ws string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(ws, activeFileName))
	if err != nil {
		t.Fatalf("read active: %v", err)
	}
	return string(b)
}

func seedEntries(t *testing.T, s *fileStore, entries ...domainhistory.Entry) {
	t.Helper()
	for _, e := range entries {
		if err := s.Append(e); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
}

// TestFileStore_Rollback_Normal removes the last N turns and keeps the survivors.
func TestFileStore_Rollback_Normal(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "one", Answer: "a1"},
		domainhistory.Entry{Prompt: "two", Answer: "a2"},
		domainhistory.Entry{Prompt: "three", Answer: "a3"},
	)
	before := readActive(t, ws)

	removed, err := s.Rollback(1)
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	entries, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 || entries[len(entries)-1].Prompt != "two" {
		t.Fatalf("survivors = %+v, want one/two", entries)
	}
	// The surviving lines must be byte-identical to their originals (the file is a
	// prefix rewrite, not a re-encode with new fields).
	if got, want := readActive(t, ws), before[:len(before)-len(`{"prompt":"three","answer":"a3"}`)-1]; got != want {
		t.Fatalf("survivors bytes = %q, want the original prefix %q", got, want)
	}
}

// TestFileStore_Rollback_Clamp removes all turns when N exceeds the count.
func TestFileStore_Rollback_Clamp(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s, domainhistory.Entry{Prompt: "only", Answer: "a"})

	removed, err := s.Rollback(5)
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1 (clamped)", removed)
	}
	entries, _ := s.Load()
	if len(entries) != 0 {
		t.Fatalf("survivors = %+v, want empty", entries)
	}
}

// TestFileStore_Rollback_NonPositiveIsNoop leaves the file untouched for n <= 0.
func TestFileStore_Rollback_NonPositiveIsNoop(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s, domainhistory.Entry{Prompt: "a", Answer: "b"})
	before := readActive(t, ws)

	for _, n := range []int{0, -1} {
		removed, err := s.Rollback(n)
		if err != nil || removed != 0 {
			t.Fatalf("Rollback(%d) = (%d, %v), want (0, nil)", n, removed, err)
		}
	}
	if readActive(t, ws) != before {
		t.Fatalf("n <= 0 must not modify the file")
	}
}

// TestFileStore_Rollback_MissingFile is 0 removed, not an error.
func TestFileStore_Rollback_MissingFile(t *testing.T) {
	s := NewFileStore(t.TempDir())
	removed, err := s.Rollback(1)
	if err != nil || removed != 0 {
		t.Fatalf("Rollback on missing file = (%d, %v), want (0, nil)", removed, err)
	}
}

// TestFileStore_Rollback_DoesNotTouchArchive pins round 081 FR-004 (rollback ≠ --new).
func TestFileStore_Rollback_DoesNotTouchArchive(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s, domainhistory.Entry{Prompt: "a", Answer: "b"})

	if _, err := s.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws, archiveFileName)); !os.IsNotExist(err) {
		t.Fatalf("archive exists (err=%v), want it absent — a rollback must not archive", err)
	}
}

// TestFileStore_Rollback_LeavesNoTempFile pins the atomic rename (no `.tmp` residue).
func TestFileStore_Rollback_LeavesNoTempFile(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "a", Answer: "b"},
		domainhistory.Entry{Prompt: "c", Answer: "d"},
	)
	if _, err := s.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws, activeFileName+".tmp")); !os.IsNotExist(err) {
		t.Fatalf("temp file residue (err=%v), want none", err)
	}
}
