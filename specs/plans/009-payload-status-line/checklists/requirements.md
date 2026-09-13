# Spec Quality Checklist: tellme Payload Status Line (round 009)

**Created**: 2026-09-13

**Feature Directory**: `specs/plans/009-payload-status-line`

**Spec Path**: `specs/plans/009-payload-status-line/spec.md`

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
- [x] This round's clarify questions are held to the budget (Round 1 = 3 questions)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-13)

- **Q1 -> Option 1 (stderr)** — `FR-001`/`FR-006`: the status line is written to the diagnostic stream (`stderr`); `stdout` stays byte-exact (`FR-005`/`FR-013`).
- **Q2 -> Option 1 (estimate + actual)** — `FR-002` (estimate) + `FR-007` (actual via widened gateway `Response` + parsed provider `usage`).
- **Q3 -> Option 1 (always-on)** — `FR-009`: not TTY-gated, not suppressed by `-r`.
- **User-locked default** — `FR-010`: `MAX_HISTORY_TOKENS` default = **1000000**.

### Still open (non-blocking)

- **Estimator constant (NFR-004)** — the exact heuristic (e.g. bytes/4); a `/axb-technical-research` determination. No new dependency.
- **`MAX_HISTORY_TOKENS: 0` semantics** — mirrors the tool-loop resolver (non-positive → default); a research determination; disclosed, non-blocking.
- **Timestamp format/clock seam (NFR-005)** — injectable clock seam; the exact rendering is a research/DSL determination.
- **Metrics line (`M:/H:/C:` / cost)** — explicitly **out of scope** (no pricing table); recorded as an assumption.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Three stories stand (pre-flight estimate · post-turn actual · configurable budget). All clarify items are resolved; only research-level defaults (estimator constant, `0` semantics, clock seam) remain. `/axb-spec-by-example` can proceed.
