package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainskills "github.com/gosharplite/tellme/internal/domain/skills"
)

// writeFixture writes content at dir/rel, creating parent directories.
func writeFixture(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

// skillBody renders a valid skill file body (frontmatter + Markdown body).
func skillBody(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n# " + name + "\n"
}

// skillNames projects a catalog to a name set (for membership assertions).
func skillNames(skills []domainskills.Skill) map[string]bool {
	out := make(map[string]bool, len(skills))
	for _, s := range skills {
		out[s.Name] = true
	}
	return out
}

// T016 [UNIT] — pins the round-033 loader contract (Decision 2 / techstack
// `Skills catalog (load)` row).

// TestLoadParsesFrontmatterRecursivelyAndIgnoresNonSkills covers the recursive
// frontmatter parse and the "non-skill Markdown is skipped" rule.
func TestLoadParsesFrontmatterRecursivelyAndIgnoresNonSkills(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "golang-patterns/SKILL.md", skillBody("golang-patterns", "Idiomatic Go patterns"))
	writeFixture(t, dir, "nested/deep/SKILL.md", skillBody("deep", "a nested skill"))
	writeFixture(t, dir, "golang-patterns/rules/REFERENCE-SENTINEL.md", "# Reference\n\nno frontmatter\n")
	writeFixture(t, dir, "STANDARDS.md", "# Standards\n")
	writeFixture(t, dir, "notes.txt", "not markdown\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	names := skillNames(got)
	if len(got) != 2 {
		t.Fatalf("Load returned %d skills (%v); want 2 (golang-patterns, deep)", len(got), names)
	}
	for _, want := range []string{"golang-patterns", "deep"} {
		if !names[want] {
			t.Errorf("missing skill %q; got %v", want, names)
		}
	}
	// A non-skill Markdown file must not be listed: assert no listed entry's
	// *location* names it (names is keyed by frontmatter `name`, so a
	// filename-keyed check there would be vacuous — the E2E asserts the same on
	// the listing text).
	for _, s := range got {
		if strings.Contains(s.Location, "REFERENCE-SENTINEL") || strings.Contains(s.Location, "STANDARDS") {
			t.Errorf("a non-skill Markdown file must not be listed; got location %q", s.Location)
		}
	}
	if got[0].Name == "golang-patterns" && got[0].Description != "Idiomatic Go patterns" {
		t.Errorf("golang-patterns description = %q; want %q", got[0].Description, "Idiomatic Go patterns")
	}
}

// TestLoadMissingOrEmptyDirYieldsEmptyCatalog covers the missing/empty-dir rule.
func TestLoadMissingOrEmptyDirYieldsEmptyCatalog(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if got, err := Load(missing); err != nil {
		t.Fatalf("Load(missing): %v", err)
	} else if len(got) != 0 {
		t.Errorf("missing dir returned %d skills; want 0", len(got))
	}
	if got, err := Load(t.TempDir()); err != nil {
		t.Fatalf("Load(empty): %v", err)
	} else if len(got) != 0 {
		t.Errorf("empty dir returned %d skills; want 0", len(got))
	}
}

// TestLoadDuplicateNameKeepsFirstDiscovered covers the duplicate-name rule.
func TestLoadDuplicateNameKeepsFirstDiscovered(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "a/SKILL.md", skillBody("dup", "first"))
	writeFixture(t, dir, "b/SKILL.md", skillBody("dup", "second"))

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("duplicate names returned %d skills; want 1", len(got))
	}
	if got[0].Description != "first" {
		t.Errorf("kept %q; want the first discovered (first)", got[0].Description)
	}
}

// TestLoadCRLFNormalized covers CRLF normalization.
func TestLoadCRLFNormalized(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "crlf/SKILL.md", "---\r\nname: crlf\r\ndescription: has CRLF\r\n---\r\n# body\r\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Name != "crlf" {
		t.Fatalf("CRLF skill not parsed; got %v", skillNames(got))
	}
}

// TestLoadSkipsMalformedBestEffort covers the best-effort skip of an individual
// malformed file (a missing closing fence; a missing description).
func TestLoadSkipsMalformedBestEffort(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "good/SKILL.md", skillBody("good", "a valid skill"))
	writeFixture(t, dir, "bad/SKILL.md", "---\nname: bad\n")
	writeFixture(t, dir, "nodesc/SKILL.md", "---\nname: nodesc\n---\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Name != "good" {
		t.Fatalf("expected only the valid skill; got %v", skillNames(got))
	}
}

// Round 075 (ADR 0047) — the frontmatter reader resolves YAML block-scalar values.

func TestLoadFoldedBlockScalarDescription(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "folded/SKILL.md", "---\nname: folded\ndescription: >\n  Idiomatic Go patterns\n  for robust code\n---\n# folded\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 skill; got %d (%v)", len(got), skillNames(got))
	}
	want := "Idiomatic Go patterns for robust code"
	if got[0].Description != want {
		t.Errorf("folded description = %q; want %q (the indicator must be resolved, not read)", got[0].Description, want)
	}
}

func TestLoadLiteralAndChompingBlockScalars(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "literal/SKILL.md", "---\nname: literal\ndescription: |\n  line one\n  line two\n---\n")
	writeFixture(t, dir, "strip/SKILL.md", "---\nname: strip\ndescription: >-\n  a\n  b\n---\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	byName := map[string]string{}
	for _, s := range got {
		byName[s.Name] = s.Description
	}
	if byName["literal"] != "line one\nline two" {
		t.Errorf("literal description = %q; want %q (breaks preserved)", byName["literal"], "line one\nline two")
	}
	if byName["strip"] != "a b" {
		t.Errorf("chomping (`>-`) description = %q; want %q", byName["strip"], "a b")
	}
}

func TestLoadInlineGreaterThanIsNotABlock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "inline/SKILL.md", "---\nname: inline\ndescription: a > b\n---\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Description != "a > b" {
		t.Fatalf("an inline value containing '>' must stay literal; got %v", got)
	}
}

func TestLoadBlockScalarWithNoBodyIsSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "good/SKILL.md", skillBody("good", "a valid skill"))
	writeFixture(t, dir, "nobody/SKILL.md", "---\nname: nobody\ndescription: >\n---\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Name != "good" {
		t.Fatalf("an indicator with no body must resolve empty and skip the file; got %v", skillNames(got))
	}
}

func TestLoadBlockScalarName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFixture(t, dir, "x/SKILL.md", "---\nname: >\n  folded name\ndescription: a description\n---\n")

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Name != "folded name" {
		t.Fatalf("a block-scalar name must be resolved; got %v", got)
	}
}
