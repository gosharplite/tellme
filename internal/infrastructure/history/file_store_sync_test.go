package history

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// recordingFS is a durableFS test double: it logs the durability primitives the
// rollback rewrite invokes, in order, and can fail the file fsync or the
// directory fsync on demand. It is the witness carrier for round 084 (ADR 0056):
// the durability guarantee is unobservable, so it is bound to a mechanism-seam
// pin rather than a Gherkin journey.
type recordingFS struct {
	events  []string
	syncErr error
	dirErr  error
}

func (r *recordingFS) Sync(f *os.File) error {
	r.events = append(r.events, "sync:"+filepath.Base(f.Name()))
	return r.syncErr
}

func (r *recordingFS) Rename(oldpath, newpath string) error {
	r.events = append(r.events, "rename:"+filepath.Base(oldpath)+"->"+filepath.Base(newpath))
	return os.Rename(oldpath, newpath)
}

func (r *recordingFS) SyncDir(dir string) error {
	r.events = append(r.events, "syncdir:"+filepath.Base(dir))
	return r.dirErr
}

// TestFileStore_Rollback_SyncsTempFileBeforeRename is the round-084 witness
// (ADR 0056): the rollback's durability clause — "the surviving entries are
// written to a temp file, the temp file is fsync'd BEFORE it is atomically
// renamed over the active file" — is asserted here as the ordered durability
// events. Discriminating mutation: remove the `s.fs.Sync(f)` call in writeRaw ⇒
// this pin reddens (the event list starts with the rename).
func TestFileStore_Rollback_SyncsTempFileBeforeRename(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "one", Answer: "a1"},
		domainhistory.Entry{Prompt: "two", Answer: "a2"},
	)
	fs := &recordingFS{}
	s.fs = fs

	removed, err := s.Rollback(1)
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	// Assert the ORDER RELATION the claim names ("the temp file is fsync'd BEFORE
	// it is renamed") rather than an exact event sequence — exact equality would
	// over-couple to an extra/benign primitive (fold N-084-2). The pin also, as a
	// side effect, guards that the rewrite routes through the seam: a direct
	// f.Sync() records no sync event.
	syncIdx, renameIdx := -1, -1
	for i, e := range fs.events {
		switch {
		case strings.HasPrefix(e, "sync:"):
			syncIdx = i
		case strings.HasPrefix(e, "rename:"):
			renameIdx = i
		}
	}
	if syncIdx < 0 || renameIdx < 0 {
		t.Fatalf("the rollback must fsync the temp file then rename it via the seam; events = %v", fs.events)
	}
	if syncIdx > renameIdx {
		t.Fatalf("the temp-file fsync must PRECEDE the rename; events = %v", fs.events)
	}
	if want := "sync:" + activeFileName + ".tmp"; fs.events[syncIdx] != want {
		t.Fatalf("the fsync must target the temp file (%q); events = %v", want, fs.events)
	}
}

// TestFileStore_Rollback_SyncErrorLeavesPriorHistoryIntact pins round-084 EC-001:
// when the temp-file fsync fails, the rollback must abort — no rename, the prior
// history byte-identical, no temp-file residue.
func TestFileStore_Rollback_SyncErrorLeavesPriorHistoryIntact(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "one", Answer: "a1"},
		domainhistory.Entry{Prompt: "two", Answer: "a2"},
	)
	before := readActive(t, ws)
	fs := &recordingFS{syncErr: errors.New("fsync failed")}
	s.fs = fs

	removed, err := s.Rollback(1)
	if err == nil {
		t.Fatalf("Rollback succeeded despite a failed fsync; removed=%d", removed)
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0 on a failed fsync", removed)
	}
	for _, e := range fs.events {
		if strings.HasPrefix(e, "rename:") {
			t.Fatalf("a rename was attempted after a failed fsync: events = %v", fs.events)
		}
	}
	if got := readActive(t, ws); got != before {
		t.Fatalf("prior history changed by a failed rollback:\n got %q\nwant %q", got, before)
	}
	if _, err := os.Stat(filepath.Join(ws, activeFileName+".tmp")); !os.IsNotExist(err) {
		t.Fatalf("temp-file residue after a failed rollback (err=%v), want none", err)
	}
}

// TestFileStore_Rollback_DirSyncErrorDoesNotFail pins round-084 EC-002 and ADR
// 0056 D4: the directory fsync is best-effort — a failure must NOT fail the
// rollback (its durability effect is an accepted-unwitnessed limit).
func TestFileStore_Rollback_DirSyncErrorDoesNotFail(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s, domainhistory.Entry{Prompt: "a", Answer: "b"})
	s.fs = &recordingFS{dirErr: errors.New("directory fsync unsupported")}

	removed, err := s.Rollback(1)
	if err != nil {
		t.Fatalf("a best-effort directory fsync must not fail the rollback: %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	if entries, _ := s.Load(); len(entries) != 0 {
		t.Fatalf("survivors = %+v, want empty", entries)
	}
}

// TestFileStore_Rollback_PreservesSurvivorBytesOnToolWrittenPath pins round-084
// EC-004 (the grill-flagged sub-clause): on the TOOL-WRITTEN path, the surviving
// lines are copied raw, byte-identical to the original prefix.
func TestFileStore_Rollback_PreservesSurvivorBytesOnToolWrittenPath(t *testing.T) {
	ws := t.TempDir()
	s := NewFileStore(ws)
	seedEntries(t, s,
		domainhistory.Entry{Prompt: "one", Answer: "a1"},
		domainhistory.Entry{Prompt: "two", Answer: "a2"},
		domainhistory.Entry{Prompt: "three", Answer: "a3"},
	)
	before := readActive(t, ws)
	// The first two lines, exactly as written (a prefix rewrite).
	parts := strings.SplitAfterN(before, "\n", 3)
	if len(parts) < 3 {
		t.Fatalf("unexpected seeded file shape: %q", before)
	}
	want := parts[0] + parts[1]

	if _, err := s.Rollback(1); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if got := readActive(t, ws); got != want {
		t.Fatalf("survivor bytes = %q, want the original prefix %q", got, want)
	}
}
