// Package tools holds the concrete agent-tool adapters behind
// internal/domain/tools. Round 024 retrofits the read-only reader tools —
// list_files, read_files, get_tree — onto the tool resource contract: each takes
// the resolved BYTE budget and bounds its own output at the source, and each
// returns a NIL-ERROR timeout result when it observes its deadline (FR-018)
// rather than the retired ctx.Err() -> error path. There is still no
// path/safety boundary (round-021 Decision 8).
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Reader bounds (round 024): the per-call ≤50 cap is kept (degenerate-input
// guard); the former fixed 100000-byte per-file cap and 1 MiB aggregate constant
// are RETIRED — the bound is now the resolved byte budget.
const readMaxPerCall = 50

// readerDefaultTimeout is the readers' protective per-call timeout default
// (round-024 FR-016), declared upward via Contract().
const readerDefaultTimeout = 30 * time.Second

// The result markers (round 024; wording fixed at the step-definition layer).
const (
	capMarker        = domaintools.TruncationMarker                // list/tree/read cut at the byte budget (single-sourced, TD3)
	readBudgetMarker = "\n... (truncated at the read budget)\n"    // read_files aggregate cut
	skipMarkerPrefix = "\n... (not read: result budget reached): " // read_files files not returned
	timeoutMarker    = "\n... (stopped at the time limit)\n"       // FR-018 nil-error timeout result
)

// maxOutputTokensDesc is the SINGLE home for the `max_output_tokens` description
// shared by every agent tool (round-029 implementation re-review: the command tool
// had advertised a slightly different cap contract).
const maxOutputTokensDesc = "Optional soft cap on this tool's result size, in tokens (the result is bounded to bytes = tokens x 4); default = the effective budget divided by 4, ceiling = the effective budget divided by 2."

// reasonDesc is the SINGLE home for the `reason` property description (round 031):
// the shared builder declares it for every tool it backs, so `required ⊆ properties`
// holds — every caller passes "reason" as required, and a strict provider
// (Vertex/Gemini) rejects a request whose tool schema lists a required name with no
// matching property (issue #64).
const reasonDesc = "Reason for calling this tool."

// resourceSchema builds EVERY agent tool's JSON schema (round-024 B3/FR-013): the
// mandatory `reason` property (round 031) plus the two resource params are declared
// for every tool, their prose single-sourced (reasonDesc; maxOutputTokensDesc; the
// timeout description names the tool generically). Declaring `reason` here keeps
// `required ⊆ properties` true for every tool the builder backs — each caller still
// passes "reason" as required (issue #64). It backs the readers and the write pair;
// `execute_command` builds inline, reusing maxOutputTokensDesc while keeping its
// bespoke props + process-tree timeout wording (round-029 review finding 4).
func resourceSchema(extraProps, required string, defaultTimeout time.Duration) json.RawMessage {
	secs := int(defaultTimeout / time.Second)
	// extraProps may be empty for a tool with no bespoke parameters (round-033
	// `list_skills`, whose properties are the mandatory `reason` + the two resource
	// params) — only insert the separating comma when it carries properties.
	prefix := ""
	if strings.TrimSpace(extraProps) != "" {
		prefix = strings.TrimSpace(extraProps) + ","
	}
	return json.RawMessage(fmt.Sprintf(
		`{"type":"object","properties":{%s"reason":{"type":"string","description":%q},"max_output_tokens":{"type":"integer","description":%q},"timeout":{"type":"number","description":"Optional seconds before this tool is stopped and returns a timeout result; default %d."}},"required":[%s]}`,
		prefix, reasonDesc, maxOutputTokensDesc, secs, required))
}

// timedOut reports whether ctx has passed its deadline (the FR-018 trigger).
func timedOut(ctx context.Context) bool {
	return errors.Is(ctx.Err(), context.DeadlineExceeded)
}

// truncateToBudget bounds a tool result to the resolved byte budget; the marker
// is counted toward the ceiling and the retained content is cut at a rune
// boundary (get_tree's box-drawing glyphs are 3 bytes each), so the whole result
// — content plus terminator — stays within the budget (round-024 NFR-001).
func truncateToBudget(out string, budget int) string {
	if budget < 1 {
		budget = 1
	}
	if len(out) <= budget {
		return out
	}
	keep := budget - len(capMarker)
	if keep < 0 {
		keep = 0
	}
	return strings.ToValidUTF8(out[:keep], "") + capMarker
}

// listFiles enumerates a directory's entries (read-only).
type listFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (listFiles) Name() string { return "list_files" }

// Description is the model-facing summary.
func (listFiles) Description() string { return "List the entries of a directory." }

// Contract declares the reader's per-tool default timeout (round-024 Q2).
func (listFiles) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: readerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments.
func (listFiles) Parameters() json.RawMessage {
	return resourceSchema(`"path":{"type":"string","description":"The directory path to list (defaults to the current directory '.')."}`, `"reason"`, readerDefaultTimeout)
}

// Execute lists the entries at the given path (defaulting to "."), one `[d]` or
// `[f]` line per entry under a `Contents of <path>:` header, in os.ReadDir order.
// The result is bounded by the resolved byte budget (round-024 FR-011); a
// deadline observed before or after the read returns a nil-error timeout result
// (FR-018).
func (listFiles) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	var args struct {
		Path   string `json:"path"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("list_files: invalid arguments: %w", err)
	}
	path := args.Path
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("list_files: failed to list directory: %w", err)
	}
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, fmt.Sprintf("Contents of %s:", path))
	for _, e := range entries {
		kind := "f"
		if e.IsDir() {
			kind = "d"
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", kind, e.Name()))
	}
	out := strings.Join(lines, "\n") + "\n"
	return truncateToBudget(out, int(budget)), nil
}

// readFiles returns files' contents (read-only), budget-bounded.
type readFiles struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (readFiles) Name() string { return "read_files" }

// Description is the model-facing summary.
func (readFiles) Description() string { return "Read a file's contents." }

// Contract declares the reader's per-tool default timeout (round-024 Q2).
func (readFiles) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: readerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments (round 021: the
// multi-file `filepaths` array; `reason` is required by schema but not validated).
func (readFiles) Parameters() json.RawMessage {
	return resourceSchema(`"filepaths":{"type":"array","items":{"type":"string"},"description":"The list of file paths to read."}`, `"filepaths","reason"`, readerDefaultTimeout)
}

// Execute reads each requested file WHOLE, in request order, stopping at the
// aggregate BYTE budget (round-024 D6 / FR-009): files that fit are returned
// framed; a file that alone exceeds the budget is truncated at the bound with a
// truncation marker; a request that cannot return every file names the ones it
// did not (a skip marker); the ≤50 per-call cap is kept. A deadline observed
// returns a nil-error timeout result (FR-018).
func (readFiles) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	var args struct {
		FilePaths []string `json:"filepaths"`
		Reason    string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("read_files: invalid arguments: %w", err)
	}
	if len(args.FilePaths) == 0 {
		return "", fmt.Errorf("read_files: filepaths argument is required and cannot be empty")
	}
	if len(args.FilePaths) > readMaxPerCall {
		return fmt.Sprintf("Error: requested too many files (%d). Maximum is %d files per call.", len(args.FilePaths), readMaxPerCall), nil
	}
	b := int(budget)
	if b < 1 {
		b = 1
	}
	// Reserve room for the aggregate + skip markers so the whole result stays
	// within the budget (round-024 D4: the tool bounds to the exact byte budget so
	// the loop's backstop stays inert). The skip marker carries the joined unread
	// paths, so reserve their worst-case size (review TD2), not a magic constant.
	skipReserve := 0
	for _, p := range args.FilePaths {
		skipReserve += len(p) + 2
	}
	limit := b - len(readBudgetMarker) - len(skipMarkerPrefix) - skipReserve
	if limit < 1 {
		limit = 1
	}
	var sb strings.Builder
	var skipped []string
	for i, path := range args.FilePaths {
		if timedOut(ctx) {
			return timeoutMarker, nil
		}
		room := limit - sb.Len()
		block, oversized := renderFile(path, room)
		if !oversized {
			sb.WriteString(block)
			continue
		}
		// The file does not fit in the remaining room.
		if sb.Len() == 0 {
			// A single file larger than the bound is truncated at the bound
			// (round-024 FR-010); the remaining files are not read.
			sb.WriteString(block)
			skipped = append(skipped, args.FilePaths[i+1:]...)
		} else {
			// A request that cannot return every file names the rest (FR-009); an
			// omitted file gets no header.
			skipped = append(skipped, args.FilePaths[i:]...)
		}
		break
	}
	if len(skipped) > 0 {
		sb.WriteString(readBudgetMarker)
		sb.WriteString(skipMarkerPrefix + strings.Join(skipped, ", ") + "\n")
	}
	return sb.String(), nil
}

// renderFile renders one read_files block — a `--- File: <path> ---` header
// followed by the file's body (`\n\n`-terminated), or an inline ERROR / binary /
// directory message — bounded to room bytes. It reports truncation when the file
// does not fit within room (so the caller stops and names the rest).
func renderFile(path string, room int) (block string, truncated bool) {
	if room < 1 {
		room = 1
	}
	var sb strings.Builder
	header := fmt.Sprintf("--- File: %s ---\n", path)
	sb.WriteString(header)
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String(), false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String(), false
	}
	if info.IsDir() {
		sb.WriteString("ERROR: path is a directory, use list_files instead\n\n")
		return sb.String(), false
	}
	contentRoom := room - len(header) - len(capMarker)
	if contentRoom < 1 {
		contentRoom = 1
	}
	data, err := io.ReadAll(io.LimitReader(f, int64(contentRoom)+1))
	if err != nil {
		fmt.Fprintf(&sb, "ERROR: failed to read file: %v\n\n", err)
		return sb.String(), false
	}
	if isBinary(data) {
		sb.WriteString("(Binary file, cannot display as text)\n\n")
		return sb.String(), false
	}
	if len(data) > contentRoom {
		sb.WriteString(strings.ToValidUTF8(string(data[:contentRoom]), ""))
		sb.WriteString(capMarker)
		return sb.String(), true
	}
	sb.Write(data)
	sb.WriteString("\n\n")
	return sb.String(), false
}

// NewFilesystemTools returns the read-only filesystem reader tools in offer order
// (round 021: list_files, read_files, get_tree).
func NewFilesystemTools() []domaintools.Tool {
	return []domaintools.Tool{listFiles{}, readFiles{}, getTree{}}
}
