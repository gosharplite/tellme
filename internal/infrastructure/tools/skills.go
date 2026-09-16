package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domainskills "github.com/gosharplite/tellme/internal/domain/skills"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// listSkillsToolName is the wire-valid canonical identifier — single-sourced so
// the tool's Name() and the catalog-binding lookup cannot drift (round 033).
const listSkillsToolName = "list_skills"

// noSkillsMessage is the empty-catalog result (FR-004): an empty catalog is
// reported as a RESULT, never an error.
const noSkillsMessage = "No skills are available.\n"

// listSkills is the read-only `list_skills` agent tool (round 033): it surfaces
// the runtime home's loaded skills catalog — each skill's name, description, and
// readable location — in a deterministic (path-sorted) order. It performs NO
// injection and reads nothing at construction: the catalog source is a lazy seam
// resolved only inside Execute (FR-009).
type listSkills struct {
	catalog func() ([]domainskills.Skill, error)
}

// NewSkillsTool builds the read-only `list_skills` tool over the given (lazy)
// catalog source. A nil source yields an empty catalog — the unbound default
// `agentTools()` constructs, which performs no filesystem read (the catalog is
// loaded and bound on the prompt path; round-033 FR-009).
func NewSkillsTool(catalog func() ([]domainskills.Skill, error)) domaintools.Tool {
	if catalog == nil {
		catalog = func() ([]domainskills.Skill, error) { return nil, nil }
	}
	return &listSkills{catalog: catalog}
}

// BindSkillsCatalog rebinds the catalog source on a `list_skills` tool (the
// prompt-path wiring seam; FR-009). It is a no-op when the registry carries no
// bindable `list_skills` tool (e.g. a test registry).
func BindSkillsCatalog(reg domaintools.Registry, catalog func() ([]domainskills.Skill, error)) {
	if t, ok := reg.Lookup(listSkillsToolName); ok {
		if ls, ok := t.(*listSkills); ok {
			ls.catalog = catalog
		}
	}
}

// Name is the wire-valid canonical identifier.
func (listSkills) Name() string { return listSkillsToolName }

// Description is the model-facing summary.
func (listSkills) Description() string {
	return "List the skills available in this environment."
}

// Contract declares the tool's per-tool default timeout — the local-reader class
// (a local directory read, like the reader trio; round-033 Decision 7).
func (listSkills) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: readerDefaultTimeout}
}

// Parameters is the JSON-schema for the tool's arguments: the mandatory `reason`
// plus the two resource params the shared builder declares (it has no bespoke
// params of its own), so `required ⊆ properties` holds (round 031).
func (listSkills) Parameters() json.RawMessage {
	return resourceSchema("", `"reason"`, readerDefaultTimeout)
}

// Execute lists the runtime home's loaded skills — name + description + readable
// location — in a deterministic, path-sorted order; an empty catalog is reported
// as a RESULT (noSkillsMessage), not an error (FR-004). The result is bounded by
// the resolved byte budget at the source, and a deadline observed before or after
// the read returns a nil-error timeout result (FR-018).
func (l listSkills) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	catalog, err := l.catalog()
	if err != nil {
		return "", fmt.Errorf("list_skills: %w", err)
	}
	skills := append([]domainskills.Skill(nil), catalog...)
	sort.Slice(skills, func(i, j int) bool { return skills[i].Location < skills[j].Location })
	if timedOut(ctx) {
		return timeoutMarker, nil
	}
	return truncateToBudget(renderSkills(skills), int(budget)), nil
}

// renderSkills formats the catalog: a `Skills (<n>):` header then one
// `- <name>: <description> (<location>)` line per skill; an empty catalog renders
// the empty-catalog message.
func renderSkills(skills []domainskills.Skill) string {
	if len(skills) == 0 {
		return noSkillsMessage
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Skills (%d):\n", len(skills))
	for _, s := range skills {
		fmt.Fprintf(&sb, "- %s: %s (%s)\n", s.Name, s.Description, s.Location)
	}
	return sb.String()
}
