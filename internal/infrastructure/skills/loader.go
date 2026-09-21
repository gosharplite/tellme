// Package skills loads the runtime home's skills catalog from disk: it walks the
// skills directory recursively and parses each Markdown file's YAML frontmatter,
// treating a file as a skill iff it declares both `name` and `description`
// (round-033 Decision 2 — the reference's parseSkill rule). Best-effort: a
// missing, empty, or unreadable directory is an empty catalog, and an individual
// malformed/unreadable file is skipped — the run is never failed (FR-001/NFR-001).
package skills

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	domainskills "github.com/gosharplite/tellme/internal/domain/skills"
)

// Load reads every skill definition under dir (recursively) and returns them in
// walk (lexical) order. A missing, empty, or unreadable dir yields an empty
// catalog and a nil error (FR-001); an individual malformed/unreadable file is
// skipped best-effort (NFR-001) — the run is never failed.
func Load(dir string) ([]domainskills.Skill, error) {
	var out []domainskills.Skill
	seen := make(map[string]bool)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // best-effort: skip an unreadable entry or subtree
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil // best-effort: skip an unreadable file
		}
		name, description, ok := parseFrontmatter(data)
		if !ok {
			return nil // not a skill — no valid name/description frontmatter
		}
		if seen[name] {
			return nil // duplicate name: the first discovered wins (Decision 2)
		}
		seen[name] = true
		out = append(out, domainskills.Skill{Name: name, Description: description, Location: path})
		return nil
	})
	if err != nil {
		return out, err
	}
	return out, nil
}

// frontmatterFence bounds a skill file's YAML frontmatter block (round 033).
const frontmatterFence = "---"

// parseFrontmatter reports whether data opens with a `---` YAML frontmatter block
// declaring both `name` and `description`, returning their values. CRLF line
// endings are normalized first (Decision 2). Round 075 (ADR 0047): a value may be
// a YAML block scalar (`>`/`|`, with chomping) — the following indented block is
// folded/literal-resolved instead of taking the indicator text as the value.
func parseFrontmatter(data []byte) (name, description string, ok bool) {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(s, frontmatterFence+"\n") {
		return "", "", false
	}
	rest := s[len(frontmatterFence)+1:]
	end := strings.Index(rest, "\n"+frontmatterFence)
	if end < 0 {
		return "", "", false
	}
	lines := strings.Split(rest[:end], "\n")
	for i := 0; i < len(lines); i++ {
		key, val, found := strings.Cut(lines[i], ":")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "name":
			name, i = resolveFrontmatterValue(strings.TrimSpace(val), lines, i)
		case "description":
			description, i = resolveFrontmatterValue(strings.TrimSpace(val), lines, i)
		}
	}
	if name == "" || description == "" {
		return "", "", false
	}
	return name, description, true
}

// resolveFrontmatterValue resolves a frontmatter scalar at lines[i] whose
// trimmed value is raw. A bare block-scalar indicator (isBlockIndicator) opens an
// indented block on the following lines (round 075); any other value is an inline
// scalar, resolved exactly as before. It returns the resolved value and the index
// of the last line consumed, so the caller's loop skips a block body.
func resolveFrontmatterValue(raw string, lines []string, i int) (value string, last int) {
	if !isBlockIndicator(raw) {
		return unquote(raw), i
	}
	last, block := collectBlock(lines, i+1)
	return strings.TrimSpace(foldBlockScalar(raw, block)), last
}

// isBlockIndicator reports whether a trimmed frontmatter value is a bare YAML
// block-scalar indicator: `>` or `|`, optionally with a chomping modifier (`-`
// strip / `+` keep). A value that merely contains `>` (e.g. `a > b`) is not one.
func isBlockIndicator(raw string) bool {
	if len(raw) < 1 || len(raw) > 2 {
		return false
	}
	if raw[0] != '>' && raw[0] != '|' {
		return false
	}
	if len(raw) == 2 && raw[1] != '-' && raw[1] != '+' {
		return false
	}
	return true
}

// collectBlock gathers the lines of a block scalar starting at index start: the
// following blank lines and indented lines, down to the first non-blank line
// without leading whitespace (the next key). It returns the index of the last
// consumed line (start-1 when the block is empty).
func collectBlock(lines []string, start int) (int, []string) {
	last := start - 1
	var block []string
	for j := start; j < len(lines); j++ {
		ln := lines[j]
		if strings.TrimSpace(ln) == "" {
			block = append(block, "")
			last = j
			continue
		}
		if ln[0] != ' ' && ln[0] != '\t' {
			break
		}
		block = append(block, ln)
		last = j
	}
	return last, block
}

// foldBlockScalar resolves a block scalar to its text: `>` folds intra-block line
// breaks to single spaces (a blank line becomes a newline) while `|` keeps them;
// chomping (`-` strip / `+` keep / default clip) governs the trailing newline.
// The block's common indentation (the first non-blank line's) is stripped.
func foldBlockScalar(indicator string, block []string) string {
	return applyChomping(indicator, joinBlockScalar(indicator, stripBlockIndent(block)))
}

// stripBlockIndent removes the block's common indentation (the first non-blank
// line's leading whitespace) from every content line; blank lines stay blank.
func stripBlockIndent(block []string) []string {
	indent := ""
	for _, ln := range block {
		if strings.TrimSpace(ln) != "" {
			indent = ln[:len(ln)-len(strings.TrimLeft(ln, " \t"))]
			break
		}
	}
	content := make([]string, len(block))
	for k, ln := range block {
		if strings.TrimSpace(ln) != "" {
			content[k] = strings.TrimPrefix(ln, indent)
		}
	}
	return content
}

// joinBlockScalar joins a block's content lines: `>` folds consecutive non-empty
// lines with a single space and a blank line to a newline; `|` keeps newlines.
func joinBlockScalar(indicator string, content []string) string {
	if indicator[0] != '>' {
		return strings.Join(content, "\n")
	}
	var b strings.Builder
	prevEmpty := true
	for _, ln := range content {
		switch {
		case ln == "":
			b.WriteString("\n")
			prevEmpty = true
		case prevEmpty:
			b.WriteString(ln)
			prevEmpty = false
		default:
			b.WriteString(" ")
			b.WriteString(ln)
		}
	}
	return b.String()
}

// applyChomping governs a block scalar's trailing newline: `-` strips it, `+`
// keeps it, and the default (clip) leaves a single trailing newline. NOTE
// (round-075 fold F-2): the caller, resolveFrontmatterValue, `TrimSpace`s the
// resolved value, which normalises ALL trailing-newline chomping — so for the
// trimmed scalar `>`, `>-`, `>+` (and `|`, `|-`, `|+`) are indistinguishable.
// These branches preserve YAML fidelity (and would matter for an untrimmed
// value); they are intentionally inert for the frontmatter scalar.
func applyChomping(indicator, out string) string {
	if len(indicator) == 2 {
		switch indicator[1] {
		case '-':
			return strings.TrimRight(out, "\n")
		case '+':
			return out
		}
	}
	if out != "" && !strings.HasSuffix(out, "\n") {
		return out + "\n"
	}
	return out
}

// unquote strips a single pair of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
