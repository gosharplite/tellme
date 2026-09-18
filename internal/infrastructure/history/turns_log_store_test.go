package history

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTurnsLogStoreRoundTrip covers the round-053 (ADR 0022) per-session turn
// log: append via Writer, read back, archive on `--new`, and the missing-file
// tolerance.
func TestTurnsLogStoreRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewTurnsLogStore(dir)

	// A missing active log reads as empty, not an error.
	got, err := s.Read()
	if err != nil {
		t.Fatalf("Read on a missing log: %v", err)
	}
	if got != "" {
		t.Fatalf("Read on a missing log = %q, want empty", got)
	}

	// Append two chrome lines.
	wc, err := s.Writer()
	if err != nil {
		t.Fatalf("Writer: %v", err)
	}
	if _, err := wc.Write([]byte("line-one\nline-two\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err = s.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got != "line-one\nline-two\n" {
		t.Fatalf("Read = %q, want the two appended lines", got)
	}

	// Archive moves the active log to the archive and clears the active file.
	if err := s.Archive(); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, turnsLogActiveFileName)); !os.IsNotExist(err) {
		t.Fatalf("the active turns.log must be removed by Archive (stat err = %v)", err)
	}
	archived, err := os.ReadFile(filepath.Join(dir, turnsLogArchiveFileName))
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	if string(archived) != "line-one\nline-two\n" {
		t.Fatalf("archive = %q, want the archived lines", string(archived))
	}
	got, err = s.Read()
	if err != nil {
		t.Fatalf("Read after Archive: %v", err)
	}
	if got != "" {
		t.Fatalf("Read after Archive = %q, want empty", got)
	}
}
