package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

const toolUsageFileName = "tools-count.jsonl"

// ToolUsageStore is the USER-GLOBAL, append-only tool-usage log adapter (round
// 026): it records each executed agent-tool invocation to
// `~/.tellme/tools-count.jsonl` (the user home resolved through an injected seam
// — deliberately OUTSIDE `TELL_ME_HOME`) and streams the log into per-tool counts
// for the offline report.
//
// It is best-effort: an unresolvable user home is a silent no-op on Record, the
// directory/file are created LAZILY on the first Record (never at construction),
// and the reader is resilient to malformed/torn lines.
type ToolUsageStore struct {
	homeDir func() (string, error)

	// mkdirOnce hoists the `~/.tellme/` creation to once per store (the first
	// Record), so a turn with k tool calls pays one MkdirAll, not k.
	mkdirOnce sync.Once
	mkdirErr  error
}

var _ domainhistory.ToolUsageSink = (*ToolUsageStore)(nil)

// NewToolUsageStore returns a tool-usage store rooted at the user home resolved
// by homeDir (typically os.UserHomeDir). The home is resolved per call, so a
// test can inject a temporary directory.
func NewToolUsageStore(homeDir func() (string, error)) *ToolUsageStore {
	return &ToolUsageStore{homeDir: homeDir}
}

// path resolves `~/.tellme/tools-count.jsonl`.
func (s *ToolUsageStore) path() (string, error) {
	if s.homeDir == nil {
		return "", errors.New("tool usage: no user-home resolver")
	}
	home, err := s.homeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("tool usage: the user home is not resolvable")
	}
	return filepath.Join(home, ".tellme", toolUsageFileName), nil
}

// Record lazily creates `~/.tellme/` on the first write and appends one JSON line
// for the invocation. The timestamp is stamped here with time.Now (UTC RFC3339)
// — explicitly UNASSERTED (round 026 research D2).
func (s *ToolUsageStore) Record(tool string, outcome domainhistory.ToolOutcome) error {
	p, err := s.path()
	if err != nil {
		return err
	}
	if err := s.ensureDir(filepath.Dir(p)); err != nil {
		return err
	}
	rec := domainhistory.ToolUsageRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Tool:      tool,
		Outcome:   outcome,
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	// No f.Sync(): the log is best-effort (FR-004 — a lost tail on crash is
	// acceptable), and one fsync per record in the loop's hot path buys nothing
	// the spec asks for. Durability is left to the OS (round 018 batches once per
	// turn — implementation review A).
	return f.Close()
}

// ensureDir creates the log directory once per store (the first Record).
func (s *ToolUsageStore) ensureDir(dir string) error {
	s.mkdirOnce.Do(func() { s.mkdirErr = os.MkdirAll(dir, 0o755) })
	return s.mkdirErr
}

// Aggregate streams the log in ONE pass into per-tool counts — O(tools) memory,
// never materialising every record (round 026 NFR-001 / research D2). A missing
// log (or an unresolvable home) is an empty aggregate; a malformed/torn line is
// SKIPPED best-effort (it never fails on log CONTENT — FR-012). A GENUINE open
// failure (e.g. a permission error) is RETURNED so the caller can diagnose it
// rather than silently reporting all-zero — "unreadable" must be distinguishable
// from "never used" (implementation review C/D).
func (s *ToolUsageStore) Aggregate() (map[string]ToolUsageCounts, error) {
	counts := map[string]ToolUsageCounts{}
	p, err := s.path()
	if err != nil {
		return counts, nil
	}
	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return counts, nil // no log yet → empty
		}
		return counts, err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	for {
		line, rerr := r.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			var rec domainhistory.ToolUsageRecord
			if jerr := json.Unmarshal([]byte(trimmed), &rec); jerr == nil && rec.Tool != "" {
				c := counts[rec.Tool]
				switch rec.Outcome {
				case domainhistory.ToolOutcomeOK:
					c.OK++
				case domainhistory.ToolOutcomeError:
					c.Error++
				case domainhistory.ToolOutcomeTimeout:
					c.Timeout++
				}
				counts[rec.Tool] = c
			}
			// A malformed line (or unknown outcome) is skipped best-effort.
		}
		if rerr != nil {
			// io.EOF or a torn final line: stop; the report still succeeds.
			break
		}
	}
	return counts, nil
}

// Load returns the log's records in file order. It exists for round-trip unit
// tests only — the REPORT must use Aggregate (streaming). A missing log is empty.
func (s *ToolUsageStore) Load() ([]domainhistory.ToolUsageRecord, error) {
	p, err := s.path()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var recs []domainhistory.ToolUsageRecord
	r := bufio.NewReader(f)
	for {
		line, rerr := r.ReadString('\n')
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			var rec domainhistory.ToolUsageRecord
			if jerr := json.Unmarshal([]byte(trimmed), &rec); jerr == nil {
				recs = append(recs, rec)
			}
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}
			return recs, rerr
		}
	}
	return recs, nil
}

// ToolUsageCounts is a tool's tally by outcome (round 026).
type ToolUsageCounts struct {
	OK      int
	Error   int
	Timeout int
}

// Total is the tool's total invocation count.
func (c ToolUsageCounts) Total() int { return c.OK + c.Error + c.Timeout }
