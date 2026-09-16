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
// endings are normalized first (Decision 2).
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
	for _, line := range strings.Split(rest[:end], "\n") {
		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "name":
			name = unquote(strings.TrimSpace(val))
		case "description":
			description = unquote(strings.TrimSpace(val))
		}
	}
	if name == "" || description == "" {
		return "", "", false
	}
	return name, description, true
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
