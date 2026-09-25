package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// search_files (round 071; ADR 0043): the missing half of the reader trio — a
// bounded, deterministic in-file content search over a directory subtree. It
// clears tellme's design-intent bar (a dedicated tool
// beats bash on a real axis): context-boundness (an unbounded `grep -rn` can blow
// the window) and determinism (a fixed result order — a tool's output is
// executable truth). Against the reference
// (`tell-me-go/internal/tools/workspace/search.go`,
// `internal/pkg/concurrentsearch/`), the ledger records THREE DELIBERATE
// DIVERGENCES:
//
//   - NO `SafePath`/consent gate — the settled no-security-layer direction; the
//     tool reads whatever path it is given, like its sibling readers.
//   - NO `WorkspacePolicy` directory ignore list — that policy serves a
//     secret-scanning concern tellme does not own; the caller scopes with `path`.
//   - DETERMINISTIC order — the reference's result order is worker-dependent;
//     tellme sorts matches by path, then line.
//
// PLUS THREE FURTHER RECORDED DIFFERENCES (review A1 / TD-071-4): the reference
// appends `" (truncated)"` when it cuts a line at 500 (tellme cuts SILENTLY —
// the line is capped at 500 BYTES), skips any file > 1 MiB (tellme does not; the
// budget bounds the result), and probes binaries at 1024 B (tellme probes 8000 B).
//
// Bounds (round-071 clarify Q1–Q3): a literal query by default with an `is_regex`
// opt-in (Q1); skip binary files + a max-line token of 10 MB, no directory ignore
// list, unreadable paths skipped best-effort (Q2); the round-024 byte budget is
// the primary bound on EVERY return path, with a hard degenerate match cap of 100
// (the first 100 by path of the first 1000 walk-order candidates) and a per-line
// trim of 500 BYTES, matches sorted path-then-line (Q3).

const (
	// searchMaxMatches is the hard degenerate cap on reported matches (the
	// reference's 100); the byte budget is the primary bound.
	searchMaxMatches = 100
	// searchScanCap bounds the collected candidate set before sorting/capping, so
	// a pathological tree cannot exhaust memory while the ordering stays
	// deterministic (far beyond the reported cap).
	searchScanCap = 1000
	// searchLineCap trims each reported match line at 500 BYTES (rune-boundary safe).
	searchLineCap = 500
	// searchMaxLineBytes is the scanner's max token size (10 MB, reference
	// parity); a line longer than this ends the scan of that file best-effort.
	searchMaxLineBytes = 10 * 1024 * 1024
	// searchBinProbe is how many leading bytes are probed for binary content.
	searchBinProbe = 8000
)

// searchFiles searches file contents under a directory subtree (read-only).
type searchFiles struct{}

// Name is the wire-valid canonical identifier.
func (searchFiles) Name() string { return "search_files" }

// Description is the model-facing summary.
func (searchFiles) Description() string {
	return "Search file contents under a directory for a text pattern."
}

// Contract declares the reader's per-tool default timeout (round-024 Q2).
func (searchFiles) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: readerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments (built through the
// shared resourceSchema so `required ⊆ properties` holds — round 031).
func (searchFiles) Parameters() json.RawMessage {
	return resourceSchema(
		`"path":{"type":"string","description":"The directory to search (defaults to '.')."},"query":{"type":"string","description":"The text or pattern to search for (a literal string unless is_regex is true)."},"is_regex":{"type":"boolean","description":"If true, treat query as a regular expression; default false (literal string search)."}`,
		`"query","reason"`, readerDefaultTimeout)
}

// Execute walks `path` (default ".") and returns each matching line as
// `path:line: text`, sorted by path then line, bounded by the resolved byte
// budget, with a hard 100-match cap. An empty query is a stable tool error; in
// regex mode an invalid pattern is a stable tool error naming it. A deadline
// observed returns a nil-error timeout result (FR-018).
func (searchFiles) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	var args struct {
		Path    string `json:"path"`
		Query   string `json:"query"`
		IsRegex bool   `json:"is_regex"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("search_files: invalid arguments: %w", err)
	}
	if args.Query == "" {
		return "", fmt.Errorf("search_files: the query argument is required")
	}
	path := args.Path
	if path == "" {
		path = "."
	}
	matcher, err := newSearchMatcher(args.Query, args.IsRegex)
	if err != nil {
		return "", err
	}

	matches, err := collectMatches(ctx, path, matcher)
	if err != nil {
		if timedOut(ctx) {
			return timeoutMarker, nil
		}
		return "", fmt.Errorf("search_files: %w", err)
	}
	if len(matches) == 0 {
		return truncateToBudget(fmt.Sprintf("0 matches found for %q in %q\n", args.Query, path), int(budget)), nil
	}
	return truncateToBudget(formatMatches(matches), int(budget)), nil
}

// searchMatch is one match: its path, 1-based line number, and the line text.
type searchMatch struct {
	path string
	line int
	text string
}

// collector accumulates matches up to searchScanCap.
type collector struct {
	matches []searchMatch
}

// add records one match and reports whether there is still room (false once the
// scan cap is reached, so the walk can stop early).
func (c *collector) add(m searchMatch) bool {
	c.matches = append(c.matches, m)
	return len(c.matches) < searchScanCap
}

// full reports whether the scan cap has been reached.
func (c *collector) full() bool { return len(c.matches) >= searchScanCap }

// newSearchMatcher builds the line predicate for the query: a literal substring
// match by default, or a compiled regular expression when isRegex is set. An
// invalid pattern is a stable tool error naming it.
func newSearchMatcher(query string, isRegex bool) (func(string) bool, error) {
	if !isRegex {
		return func(line string) bool { return strings.Contains(line, query) }, nil
	}
	re, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("search_files: invalid regex %q: %w", query, err)
	}
	return re.MatchString, nil
}

// collectMatches walks path (a file is searched directly), collecting matches in
// a deterministic order (path asc, then line asc), bounded by searchScanCap so
// the candidate set cannot exhaust memory before the reported cap applies.
func collectMatches(ctx context.Context, path string, match func(string) bool) ([]searchMatch, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to search %q: %w", path, err)
	}
	c := &collector{}
	if !info.IsDir() {
		if err := scanFile(ctx, path, match, c); err != nil {
			return nil, err
		}
	} else if err := walkDir(ctx, path, match, c); err != nil {
		return nil, err
	}
	sortMatches(c.matches)
	return c.matches, nil
}

// walkDir recurses (os.ReadDir returns entries sorted by name, so the walk is
// deterministic); it skips binary files (scanFile) and unreadable entries
// best-effort. There is deliberately NO directory ignore list (round-071 Q2).
func walkDir(ctx context.Context, dir string, match func(string) bool, c *collector) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // unreadable directory: skip best-effort
	}
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			if err := walkDir(ctx, full, match, c); err != nil {
				return err
			}
		} else if err := scanFile(ctx, full, match, c); err != nil {
			return err
		}
		if c.full() {
			return nil
		}
	}
	return nil
}

// scanFile probes a file for binary content (skipping it if binary), then scans
// it line-by-line. A read error or an over-long line ends that file's scan
// best-effort (never a hard failure).
func scanFile(ctx context.Context, path string, match func(string) bool, c *collector) error {
	f, err := os.Open(path)
	if err != nil {
		return nil // unreadable file: skip best-effort
	}
	defer func() { _ = f.Close() }()

	head := make([]byte, searchBinProbe)
	n, err := io.ReadFull(f, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil
	}
	if isBinary(head[:n]) {
		return nil
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 4096), searchMaxLineBytes)
	lineNum := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		lineNum++
		line := scanner.Text()
		if !match(line) {
			continue
		}
		if !c.add(searchMatch{path: path, line: lineNum, text: trimLine(line)}) {
			return nil
		}
	}
	return nil
}

// trimLine trims surrounding whitespace and caps the line at searchLineCap bytes
// on a rune boundary.
func trimLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) <= searchLineCap {
		return trimmed
	}
	return strings.ToValidUTF8(trimmed[:searchLineCap], "")
}

// sortMatches orders matches deterministically: by path (asc), then line (asc).
func sortMatches(matches []searchMatch) {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].path != matches[j].path {
			return matches[i].path < matches[j].path
		}
		return matches[i].line < matches[j].line
	})
}

// formatMatches renders the matches as `path:line: text` lines (deterministically
// ordered), capping at searchMaxMatches with the shared truncation marker.
func formatMatches(matches []searchMatch) string {
	var sb strings.Builder
	limit := len(matches)
	capped := false
	if limit > searchMaxMatches {
		limit = searchMaxMatches
		capped = true
	}
	for _, m := range matches[:limit] {
		fmt.Fprintf(&sb, "%s:%d: %s\n", filepath.ToSlash(m.path), m.line, m.text)
	}
	if capped {
		sb.WriteString(domaintools.TruncationMarker)
	}
	return sb.String()
}

// NewSearchTool returns the in-file content search tool (round 071; ADR 0043) as
// a one-element group, offered after the reader trio.
func NewSearchTool() []domaintools.Tool { return []domaintools.Tool{searchFiles{}} }

// Compile-time port conformance.
var _ domaintools.Tool = searchFiles{}
