# Spec Quality Checklist: tellme spinner width safety — a bounded tool label and a residue-free clear (round 025)

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/025-spinner-width-safety`

**Spec Path**: `specs/plans/025-spinner-width-safety/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement
- [x] Edge cases cover the main high-risk scenarios
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order
- [x] Each user story is independently verifiable
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps are escalated to `/axb-clarify`
- [x] This round's clarify questions are held to the budget (the two scoping questions were answered directly by the operator: Q1 → 1, Q2 → 1)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (operator-locked, session 2026-09-15)

- **Q1 → 1 — fix both defects**: the several-tool label cap **and** a width-safe clear. Rationale: the clear residue is a real violation of the round-019 teardown contract, not a cosmetic symptom.
- **Q2 → 1 — width-safe mechanism**: track the last frame's rendered-row count and erase **all** its rows (rather than clamping the line to the terminal width). Rationale: robust for any wrap cause (label, resource segment, terminal resize).

### Reference finding (recorded)

- **`tell-me-go` has the same two defects** — it enumerates every tool name (`internal/domain/events/types.go` `ToolExecutionStartedEvent.SpinnerInfo`) and clears a **single** row (`\r` + `\033[2K`, `internal/ui/renderer_spinner.go`). There is **no upstream fix to port**; both fixes are deliberate **divergences**.

### Still open (non-blocking — expected downstream detail)

- **Width source / seam** — how the terminal width reaches the spinner (an injected seam defaulting to a real probe; `golang.org/x/term` is already in the module graph) and the unknown-width fallback: a `/axb-technical-research` decision.
- **Exact bounded-label wording** and the multi-row clear byte sequence: a `/axb-dsl-refine` / step-definition detail.
- **Verification split** — the label bound is E2E; the row-aware clear is a **unit** pin (a flat capture cannot reproduce a terminal grid): confirmed as the round's approach.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (the bounded several-tool label · the residue-free clear), plus the unchanged-surface constraints. Both operator-locked decisions (Q1 → 1, Q2 → 1) are settled and no blocking gap remains — `/axb-spec-by-example` can proceed.
