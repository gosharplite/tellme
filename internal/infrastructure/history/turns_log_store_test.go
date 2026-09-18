package history

import (
	"io"
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
	if got := readAll(t, s); got != "" {
		t.Fatalf("Reader on a missing log = %q, want empty", got)
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

	if got := readAll(t, s); got != "line-one\nline-two\n" {
		t.Fatalf("Reader = %q, want the two appended lines", got)
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
	if got := readAll(t, s); got != "" {
		t.Fatalf("Reader after Archive = %q, want empty", got)
	}
}

// readAll drains the store's streaming Reader (the RF-53-2 port shape).
func readAll(t *testing.T, s *turnsLogStore) string {
	t.Helper()
	rc, err := s.Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	return string(data)
}
