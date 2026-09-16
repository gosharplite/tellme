package steps

import (
	"path/filepath"
	"strings"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round-033 skills helpers: the shared vocabulary for the `list_skills` interface
// truth — arranging a skill's frontmatter file under the runtime home's skills
// directory and reading back the recorded `list_skills` result. Kept in ONE file
// so the per-sentence step files stay independent (Zero Shared Edits) and never
// collide on a top-level helper name.

// skillRelPath returns the runtime-home-relative path of a skill's definition
// file — docs/skills/<name>/SKILL.md (the skill's folder is its declared name;
// round-033 dsl `Given (round 033)`).
func skillRelPath(name string) string {
	return filepath.ToSlash(filepath.Join("docs", "skills", name, "SKILL.md"))
}

// skillReferenceRelPath returns the runtime-home-relative path of a non-skill
// reference file inside a skill's folder (e.g. rules/REFERENCE-SENTINEL.md).
func skillReferenceRelPath(name, relpath string) string {
	return filepath.ToSlash(filepath.Join("docs", "skills", name, relpath))
}

// skillFileMarkdown renders a SKILL.md body: a YAML frontmatter block declaring
// `name` and `description`, followed by a Markdown body. Its content starts with
// `---\n`, the loader's frontmatter trigger.
func skillFileMarkdown(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n# " + name + "\n"
}

// listingHasSkill reports whether a `list_skills` result carries an entry for
// name — the tool renders one `- <name>: …` line per skill.
func listingHasSkill(result, name string) bool {
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimRight(line, "\r"), "- "+name+":") {
			return true
		}
	}
	return false
}

// toolResultFedBack reports whether the run fed the named tool's result back into
// the conversation: some recorded request carries an assistant tool-call naming
// `tool` AND (in the same request) a subsequent tool-role message — the loop's
// executed-tool → result round trip.
func toolResultFedBack(f *fakeprovider.Provider, tool string) bool {
	for _, msgs := range toolRounds(f) {
		called := false
		for _, m := range msgs {
			if m.Role == "assistant" {
				for _, tc := range m.ToolCalls {
					if tc.Function.Name == tool {
						called = true
					}
				}
			}
			if called && m.Role == "tool" {
				return true
			}
		}
	}
	return false
}
