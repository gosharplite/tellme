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
	usageActiveFileName  = "tokens.log"
	usageArchiveFileName = "tokens.archive.jsonl"
)

// usageStore is the per-mode JSON-Lines usage-log adapter (round 018): an
// append-only `tokens.log` under the session workspace, with a
// `tokens.archive.jsonl` for `--new` (mirroring history.jsonl / history.archive.jsonl).
type usageStore struct {
	workspace string
}

var _ domainhistory.UsageStore = (*usageStore)(nil)

// NewUsageStore returns a file-backed UsageStore rooted at the session workspace.
func NewUsageStore(workspace string) *usageStore {
	return &usageStore{workspace: workspace}
}

func (s *usageStore) activePath() string  { return filepath.Join(s.workspace, usageActiveFileName) }
func (s *usageStore) archivePath() string { return filepath.Join(s.workspace, usageArchiveFileName) }

// Append writes one API call's usage as a single JSON line (append-only,
// O_APPEND|O_CREATE, flushed).
func (s *usageStore) Append(rec domainhistory.UsageRecord) error {
	line, err := json.Marshal(rec)
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

// Load returns the active log's records in call order. A missing active file is
// an empty session, not an error.
func (s *usageStore) Load() ([]domainhistory.UsageRecord, error) {
	f, err := os.Open(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var recs []domainhistory.UsageRecord
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			var rec domainhistory.UsageRecord
			if jerr := json.Unmarshal([]byte(trimmed), &rec); jerr != nil {
				return nil, jerr
			}
			recs = append(recs, rec)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return recs, nil
}

// Archive moves the active usage log into the archive file and clears the active
// file. A missing (or empty) active log is a no-op returning nil.
func (s *usageStore) Archive() error {
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
