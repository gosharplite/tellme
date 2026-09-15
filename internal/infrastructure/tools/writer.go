package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// The write tools (round 029) — tellme's first WRITE capability: `write_file`
// (create-only, atomic) and `replace_text` (strict-unique, atomic), behind the
// unchanged domain/tools.Tool port. There is NO security/consent gate and NO
// backup/undo (round-029 D5; tellme's settled exclusions). Both tools slot into
// the round-024 resource contract: they declare the two resource params and a
// per-tool default timeout of 30 s (matching the readers), bound their result at
// the source, and return a nil-error timeout result when they observe their
// deadline (FR-018).
//
// Every write is ATOMIC (round-029 D3 / FR-009): the content goes to a temp file
// in the target's directory and is moved into place — `write_file` via an atomic
// create-only `os.Link` (an existing destination fails with EEXIST, no Stat→Rename
// TOCTOU window; requires hard-link support), `replace_text` via `rename`. The
// temp file is removed on any failure, so the destination is never observed
// partial.

// The write tools' per-tool default timeout (round-029 D6, matching the readers)
// and the modes they create (round-029 review findings 2/3).
const (
	writerDefaultTimeout = 30 * time.Second
	createdFileMode      = 0o644
	createdDirMode       = 0o755
)

// writeTempContent writes the tool's new content into its temporary file. It is a
// package-level var so the unit tests can fault-inject a partial write — the
// unit-tier atomicity witness for FR-009 (no E2E fault injection exists). The
// production value performs the whole write.
var writeTempContent = func(f *os.File, data []byte) error {
	_, err := f.Write(data)
	return err
}

// writeAtomic writes data to dest atomically: the content is written to a temp
// file IN dest's directory (so the move is a same-filesystem, atomic operation),
// the temp's mode is set to `mode`, and it is moved into place. When createOnly
// is true the move is an atomic create-only link (`os.Link` fails with EEXIST on
// an existing destination — no Stat-then-Rename TOCTOU); otherwise it is a
// `rename` (which atomically replaces the destination). The temp file is removed
// on any failure, so dest is never observed partial.
func writeAtomic(dest string, data []byte, createOnly bool, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".tellme-write-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file: %w", err)
	}
	tmpName := tmp.Name()
	// The temp file is removed on any exit: after a successful move the name is
	// already gone (a harmless ENOENT on the deferred Remove), and on any failure
	// this cleans it up so the destination is never left partial.
	defer func() { _ = os.Remove(tmpName) }()
	if err := writeTempContent(tmp, data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close the temporary file: %w", err)
	}
	// Set the final mode on the temp before the move, so a moved temp carries the
	// intended mode: os.CreateTemp would otherwise leave it owner-only (0600).
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("failed to set the temporary file mode: %w", err)
	}
	if createOnly {
		if err := os.Link(tmpName, dest); err != nil {
			if errors.Is(err, os.ErrExist) {
				return fmt.Errorf("the file already exists: %s", dest)
			}
			return fmt.Errorf("failed to move the temporary file into place: %w", err)
		}
		if err := os.Remove(tmpName); err != nil {
			return fmt.Errorf("failed to remove the temporary file: %w", err)
		}
		return nil
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("failed to move the temporary file into place: %w", err)
	}
	return nil
}

// mkdirAll0755 creates dir and any missing parents, forcing mode 0755 on every
// directory it creates. os.MkdirAll is umask-masked, so its mode is only "0755
// if the ambient umask allows it"; the write tool's directory guarantee must not
// depend on the ambient umask (ADR-036 determinism — round-029 review finding 2).
func mkdirAll0755(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		return fmt.Errorf("%s exists and is not a directory", dir)
	}
	if parent := filepath.Dir(dir); parent != dir {
		if err := mkdirAll0755(parent); err != nil {
			return err
		}
	}
	if err := os.Mkdir(dir, createdDirMode); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	return os.Chmod(dir, createdDirMode)
}

// writeFile creates a NEW file (create-only; it never overwrites).
type writeFile struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (writeFile) Name() string { return "write_file" }

// Description is the model-facing summary.
func (writeFile) Description() string {
	return "Create a new file with the given content (fails if the path already exists)."
}

// Contract declares the write tool's per-tool default timeout (round-029 D6).
func (writeFile) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: writerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments (round-029 D6/FR-012:
// `filepath`, `content`, `reason` required, plus the two resource params).
func (writeFile) Parameters() json.RawMessage {
	return resourceSchema(`"filepath":{"type":"string","description":"The path of the file to create."},"content":{"type":"string","description":"The exact content to write to the new file."}`, `"filepath","content","reason"`, writerDefaultTimeout)
}

// Execute creates the file at filepath with exactly content (create-only, atomic).
// A missing `content` key is rejected (only an explicit "" writes an empty file);
// missing parent directories are created (mode 0755); an existing destination is
// refused. A deadline observed before the write returns a nil-error timeout
// result (FR-018).
func (writeFile) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	path, content, err := parseWriteFileArgs(arguments)
	if err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}
	if content == nil {
		return "", errors.New(`write_file: the content argument is required (use "" to create an empty file)`)
	}
	if err := mkdirAll0755(filepath.Dir(path)); err != nil {
		return "", fmt.Errorf("write_file: failed to create parent directories: %w", err)
	}
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	if err := writeAtomic(path, []byte(*content), true, createdFileMode); err != nil {
		return "", fmt.Errorf("write_file: %w", err)
	}
	return boundWriteResult(writeFileSuccessMsg, budget), nil
}

// parseWriteFileArgs decodes write_file's arguments, distinguishing a MISSING
// `content` key (a nil pointer → rejected) from an explicit "" (an empty file). A
// JSON `null` is treated like absent (round-029 review nit) — only an explicit ""
// writes an empty file (FR-010).
func parseWriteFileArgs(arguments string) (path string, content *string, err error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(arguments), &raw); err != nil {
		return "", nil, fmt.Errorf("invalid arguments: %w", err)
	}
	pv, ok := raw["filepath"]
	if !ok {
		return "", nil, errors.New("the filepath argument is required")
	}
	if err := json.Unmarshal(pv, &path); err != nil {
		return "", nil, fmt.Errorf("invalid filepath: %w", err)
	}
	if cv, ok := raw["content"]; ok && !isJSONNull(cv) {
		var c string
		if err := json.Unmarshal(cv, &c); err != nil {
			return "", nil, fmt.Errorf("invalid content: %w", err)
		}
		content = &c
	}
	return path, content, nil
}

// isJSONNull reports whether a raw JSON value is the literal `null`.
func isJSONNull(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

// replaceText replaces a uniquely-identified block in an EXISTING file.
type replaceText struct{}

// Name is the wire-valid canonical identifier (round-008 BLOCKER-1).
func (replaceText) Name() string { return "replace_text" }

// Description is the model-facing summary.
func (replaceText) Description() string {
	return "Replace the single occurrence of a block of text in an existing file."
}

// Contract declares the write tool's per-tool default timeout (round-029 D6).
func (replaceText) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: writerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments (round-029 D6/FR-012:
// `filepath`, `old_text`, `new_text`, `reason` required, plus the two resource
// params).
func (replaceText) Parameters() json.RawMessage {
	return resourceSchema(`"filepath":{"type":"string","description":"The path of the file to edit."},"old_text":{"type":"string","description":"The exact block to replace (it must occur exactly once)."},"new_text":{"type":"string","description":"The replacement text."}`, `"filepath","old_text","new_text","reason"`, writerDefaultTimeout)
}

// Execute replaces the single occurrence of old_text with new_text (strict-unique,
// atomic). An empty old_text, a missing/unreadable file, an absent block, or a
// non-unique block is a recoverable error; a no-op (new_text == old_text, on a
// uniquely-present block) short-circuits to success without a write. The edit
// PRESERVES the destination's own permissions (round-029 review finding 1). A
// deadline observed before the write returns a nil-error timeout result (FR-018).
func (replaceText) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	var args struct {
		FilePath string `json:"filepath"`
		OldText  string `json:"old_text"`
		NewText  string `json:"new_text"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("replace_text: invalid arguments: %w", err)
	}
	if args.OldText == "" {
		return "", errors.New("replace_text: the old_text argument is required and cannot be empty")
	}
	data, err := os.ReadFile(args.FilePath)
	if err != nil {
		return "", fmt.Errorf("replace_text: failed to read the file: %w", err)
	}
	content := string(data)
	count := strings.Count(content, args.OldText)
	if count == 0 {
		return "", fmt.Errorf("replace_text: the block is not present: %q", args.OldText)
	}
	if count > 1 {
		return "", fmt.Errorf("replace_text: the block is not unique (%d occurrences): %q", count, args.OldText)
	}
	// A no-op short-circuits to success WITHOUT a write (round-029 D4) — evaluated
	// AFTER the presence/uniqueness gates, so an absent block still fails (FR-002;
	// round-029 review finding 5).
	if args.OldText == args.NewText {
		return boundWriteResult(replaceTextNoOpMsg, budget), nil
	}
	// Preserve the destination's own permissions on the edit (round-029 review
	// finding 1): a rename replaces the inode, so an unconditional 0644 would
	// silently re-mode a 0600 / 0755 file.
	mode := os.FileMode(createdFileMode)
	if info, statErr := os.Stat(args.FilePath); statErr == nil {
		mode = info.Mode().Perm()
	}
	idx := strings.Index(content, args.OldText)
	updated := content[:idx] + args.NewText + content[idx+len(args.OldText):]
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	// NOTE (recorded limitation, round-029 review finding 6): the read follows a
	// symlink and the atomic rename then replaces the LINK with a regular file.
	if err := writeAtomic(args.FilePath, []byte(updated), false, mode); err != nil {
		return "", fmt.Errorf("replace_text: %w", err)
	}
	return boundWriteResult(replaceTextSuccessMsg, budget), nil
}

// The write tools' success messages (single-sourced; a no-op replace reports the
// short-circuit).
const (
	writeFileSuccessMsg   = "File written successfully.\n"
	replaceTextSuccessMsg = "File updated successfully.\n"
	replaceTextNoOpMsg    = "File unchanged: the replacement is identical to the original.\n"
)

// boundWriteResult bounds a write tool's short success message to the resolved
// byte budget (round-024 contract; shared by both write tools).
func boundWriteResult(message string, budget domaintools.ByteBudget) string {
	return truncateToBudget(message, int(budget))
}

// NewWriteTools returns the two write tools in offer order (round-029 D1:
// write_file, replace_text).
func NewWriteTools() []domaintools.Tool {
	return []domaintools.Tool{writeFile{}, replaceText{}}
}

// Compile-time port conformance.
var (
	_ domaintools.Tool = writeFile{}
	_ domaintools.Tool = replaceText{}
)
