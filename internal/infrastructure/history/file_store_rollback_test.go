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

// TestFileStore_Rollback_FailureLeavesPriorHistoryIntact is the round-081
// durability witness (review fold F-081-1): blocking the temp path (a directory
// at `<active>.tmp` makes the temp open fail with EISDIR) must (a) return an
// error and (b) leave the prior `history.jsonl` BYTE-IDENTICAL. An in-place
// truncate would fail this: it would have already destroyed the file.
func TestFileStore_Rollback_FailureLeavesPriorHistoryIntact(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "one", Answer: "a1"},
		domainhistory.Entry{Prompt: "two", Answer: "a2"},
	)
	before := readActive(t, ws)

	// Block the temp path so the durable write cannot proceed.
	if err := os.Mkdir(filepath.Join(ws, activeFileName+".tmp"), 0o755); err != nil {
		t.Fatalf("mkdir tmp block: %v", err)
	}

	removed, err := s.Rollback(1)
	if err == nil {
		t.Fatalf("Rollback succeeded despite a blocked temp path; removed=%d", removed)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0 on a failed rollback", removed)
	}
	if got := readActive(t, ws); got != before {
		t.Fatalf("prior history was modified by a failed rollback:\n got %q\nwant %q", got, before)
	}
}

// TestFileStore_Rollback_DoesNotRewriteSurvivorBytes pins TD-081-2: survivors are
// copied RAW, so a line carrying a field this binary does not know survives a
// rollback byte-for-byte.
func TestFileStore_Rollback_DoesNotRewriteSurvivorBytes(t *testing.T) {
	ws := t.TempDir()
	// A hand-written line with an unknown field (which Load/Entry drops) and
	// non-canonical key order — a decode+re-marshal would silently rewrite it.
	unknown := `{"prompt":"kept","answer":"a","unknown_future_field":42}` + "\n"
	drop := `{"prompt":"dropped","answer":"b"}` + "\n"
	if err := os.WriteFile(filepath.Join(ws, activeFileName), []byte(unknown+drop), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewFileStore(ws)
	if _, err := s.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if got := readActive(t, ws); got != unknown {
		t.Fatalf("survivor bytes = %q, want the original line verbatim %q", got, unknown)
	}
}

// TestFileStore_Rollback_DecodeFailureLeavesFileIntact pins the round-081 edge
// case (review fold F-081-3): a malformed line makes Rollback return an error and
// never write (no half-written file).
func TestFileStore_Rollback_DecodeFailureLeavesFileIntact(t *testing.T) {
	ws := t.TempDir()
	body := `{"prompt":"ok","answer":"a"}` + "\n" + "{not valid json" + "\n"
	path := filepath.Join(ws, activeFileName)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewFileStore(ws)

	removed, err := s.Rollback(1)
	if err == nil {
		t.Fatalf("Rollback on a malformed history succeeded; removed=%d", removed)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0 on a decode failure", removed)
	}
	if got, _ := os.ReadFile(path); string(got) != body {
		t.Fatalf("history file changed on a decode failure:\n got %q\nwant %q", got, body)
	}
}

// TestFileStore_Rollback_LeavesAPreExistingArchiveByteIdentical pins the strongest
// form of "the archive is untouched" (review nit N-081-1): a pre-seeded archive
// must be byte-identical after a rollback (a version that APPENDS to the archive
// would be caught here, unlike the absent-form witness).
func TestFileStore_Rollback_LeavesAPreExistingArchiveByteIdentical(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "a", Answer: "b"},
		domainhistory.Entry{Prompt: "c", Answer: "d"},
	)
	seededArchive := `{"prompt":"old","answer":"archived"}` + "\n"
	if err := os.WriteFile(filepath.Join(ws, archiveFileName), []byte(seededArchive), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(ws, archiveFileName))
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	if string(got) != seededArchive {
		t.Fatalf("archive changed by a rollback: got %q, want %q", got, seededArchive)
	}
}
