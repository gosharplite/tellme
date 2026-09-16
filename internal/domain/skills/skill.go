// Package skills defines the skills domain value type: one loaded skill
// definition — its declared name and description, and the on-disk location the
// agent can open with the existing read_files tool (round 033).
//
// It is deliberately minimal: no repository, composite, selector, or refresh
// framework (round-033 Decision 5 — the catalog is read from disk each
// prompt-bearing turn and is not persisted).
package skills

// Skill is one loaded skill definition: its declared `name` and `description`
// (parsed from the file's YAML frontmatter) and its readable `Location` (the
// on-disk path the agent opens with read_files).
type Skill struct {
	Name        string
	Description string
	Location    string
}
