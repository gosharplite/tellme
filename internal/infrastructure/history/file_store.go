// Package history is the file-backed session-history adapter: an append-only
// JSON-Lines store (history.jsonl) under the session workspace, with an archive
// file (history.archive.jsonl) for `--new` (round-007 research Decisions 1 & 3).
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

const (
	activeFileName  = "history.jsonl"
	archiveFileName = "history.archive.jsonl"
)

// fileStore is the JSON-Lines adapter behind domainhistory.Store.
type fileStore struct {
	workspace string
}

var _ domainhistory.Store = (*fileStore)(nil)

// NewFileStore returns a file-backed Store rooted at the session workspace
// directory. Load reads wholesale with a bounded bufio.Reader (not a
// bufio.Scanner) so large prompts round-trip without a 64 KB token-cap failure
// (round-007 TD-2).
func NewFileStore(workspace string) *fileStore {
	return &fileStore{workspace: workspace}
}

func (s *fileStore) activePath() string  { return filepath.Join(s.workspace, activeFileName) }
func (s *fileStore) archivePath() string { return filepath.Join(s.workspace, archiveFileName) }

// Load returns the active entries in order. A missing active file is an empty
// history, not an error.
func (s *fileStore) Load() ([]domainhistory.Entry, error) {
	f, err := os.Open(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var entries []domainhistory.Entry
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			var e domainhistory.Entry
			if jerr := json.Unmarshal([]byte(trimmed), &e); jerr != nil {
				return nil, jerr
			}
			entries = append(entries, e)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return entries, nil
}

// Append writes one completed entry as a single JSON line (append-only,
// O_APPEND|O_CREATE, flushed). No timestamp or id is stored, so the bytes are
// deterministic (NFR-002).
func (s *fileStore) Append(e domainhistory.Entry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	f, err := os.OpenFile(s.activePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(line); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// Archive moves the active history into the archive file and clears the active
// file. A missing (or empty) active history is a no-op returning nil (RF-1).
// The archive is opened with O_APPEND|O_CREATE|O_WRONLY and flushed before the
// active file is removed, so a mid-write failure never loses the active history.
func (s *fileStore) Archive() error {
	data, err := os.ReadFile(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return os.Remove(s.activePath())
	}
	f, err := os.OpenFile(s.archivePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(s.activePath())
}

// Rollback removes the last n complete turns from the active history and returns
// the number removed (round 081 / ADR 0053). It clamps n to the available turns
// (n <= 0 is a no-op, returning 0 removed without touching the file) and is
// DURABLE: the surviving entries are written to a temp file in the same
// directory, fsync'd, then renamed over the active file (atomic on POSIX), so a
// crash mid-rollback leaves the prior history intact (the live file is never
// truncated in place). The archive is never touched. A missing active file is 0
// removed; a decode failure returns an error and never writes.
func (s *fileStore) Rollback(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}
	entries, err := s.Load()
	if err != nil {
		return 0, err
	}
	removed := n
	if removed > len(entries) {
		removed = len(entries)
	}
	if removed == 0 {
		return 0, nil
	}
	survivors := entries[:len(entries)-removed]
	if err := s.rewrite(survivors); err != nil {
		return 0, err
	}
	return removed, nil
}

// rewrite durably replaces the active history with the given entries: a
// same-directory temp file, fsync, then an atomic rename over the active file.
// The directory is fsync'd best-effort so the rename survives a crash.
func (s *fileStore) rewrite(entries []domainhistory.Entry) error {
	var buf []byte
	for _, e := range entries {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		buf = append(buf, line...)
		buf = append(buf, '\n')
	}
	tmp := s.activePath() + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(buf); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, s.activePath()); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	syncDir(filepath.Dir(s.activePath()))
	return nil
}

// syncDir fsyncs a directory so a rename is durable across a crash; best-effort
// (a directory fsync is unsupported on some filesystems).
func syncDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	_ = d.Sync()
	_ = d.Close()
}
