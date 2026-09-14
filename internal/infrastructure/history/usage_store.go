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
	usageSummaryFileName = "tokens.summary.json"
)

// usageStore is the per-mode JSON-Lines usage-log adapter (round 018): an
// append-only `tokens.log` under the session workspace, a `tokens.summary.json`
// cumulative roll-up companion (round-018 review #1), and a
// `tokens.archive.jsonl` for `--new`.
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
func (s *usageStore) summaryPath() string { return filepath.Join(s.workspace, usageSummaryFileName) }

// Append writes one API call's usage as a single append-only JSON line.
func (s *usageStore) Append(rec domainhistory.UsageRecord) error {
	return s.AppendBatch([]domainhistory.UsageRecord{rec})
}

// AppendBatch writes a turn's records as append-only JSON lines in ONE
// open/write/sync cycle, then folds them into the persisted roll-up (review #3).
func (s *usageStore) AppendBatch(recs []domainhistory.UsageRecord) error {
	if len(recs) == 0 {
		return nil
	}
	var buf []byte
	for _, rec := range recs {
		line, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		buf = append(buf, line...)
		buf = append(buf, '\n')
	}
	f, err := os.OpenFile(s.activePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(buf); err != nil {
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
	return s.bumpSummary(recs)
}

// loadSummary reads the persisted roll-up; ok is false when it is absent/corrupt.
func (s *usageStore) loadSummary() (domainhistory.UsageSummary, bool) {
	data, err := os.ReadFile(s.summaryPath())
	if err != nil {
		return domainhistory.UsageSummary{}, false
	}
	var sum domainhistory.UsageSummary
	if err := json.Unmarshal(data, &sum); err != nil {
		return domainhistory.UsageSummary{}, false
	}
	return sum, true
}

// bumpSummary folds recs into the persisted roll-up (atomic temp+rename).
func (s *usageStore) bumpSummary(recs []domainhistory.UsageRecord) error {
	sum, _ := s.loadSummary()
	for _, rec := range recs {
		sum.Add(rec)
	}
	return s.writeSummary(sum)
}

func (s *usageStore) writeSummary(sum domainhistory.UsageSummary) error {
	data, err := json.Marshal(sum)
	if err != nil {
		return err
	}
	tmp := s.summaryPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.summaryPath())
}

// Totals returns the cumulative session roll-up. It reads the persisted summary
// (O(1)); when absent it computes the totals from the active log in ONE streaming
// pass (O(N) time, O(1) memory) and persists them (review #1).
func (s *usageStore) Totals() (domainhistory.UsageSummary, error) {
	if sum, ok := s.loadSummary(); ok {
		return sum, nil
	}
	sum, err := s.computeTotals()
	if err != nil {
		return domainhistory.UsageSummary{}, err
	}
	_ = s.writeSummary(sum)
	return sum, nil
}

// computeTotals streams the active log into a session summary (no slice).
func (s *usageStore) computeTotals() (domainhistory.UsageSummary, error) {
	f, err := os.Open(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domainhistory.UsageSummary{}, nil
		}
		return domainhistory.UsageSummary{}, err
	}
	defer func() { _ = f.Close() }()

	var sum domainhistory.UsageSummary
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			var rec domainhistory.UsageRecord
			if jerr := json.Unmarshal([]byte(trimmed), &rec); jerr != nil {
				return domainhistory.UsageSummary{}, jerr
			}
			sum.Add(rec)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return domainhistory.UsageSummary{}, err
		}
	}
	return sum, nil
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

// Archive moves the active usage log into the archive file, clears the active
// file, and drops the roll-up (so a fresh session starts empty). A missing (or
// empty) active log is a no-op returning nil.
func (s *usageStore) Archive() error {
	data, err := os.ReadFile(s.activePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = os.Remove(s.summaryPath())
			return nil
		}
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		_ = os.Remove(s.summaryPath())
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
	if err := os.Remove(s.activePath()); err != nil {
		return err
	}
	_ = os.Remove(s.summaryPath())
	return nil
}
