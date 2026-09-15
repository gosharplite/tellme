# Spec Quality Checklist: tellme Agent Tool Surface Parity (round 021)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/021-tool-surface-parity`

**Spec Path**: `specs/plans/021-tool-surface-parity/spec.md`

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
- [x] This round's clarify questions are held to the budget (operator locked D1–D5 before specify)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (operator-locked, pre-specify)

- **D1 -> `read_files` multi-file** — `FR-001`: adopt `filepaths: string[]`; the round-008 single `path` is rewritten.
- **D2 -> `reason` required + echoed** — `FR-012`: required on all three tools; value echoed into the `stderr` tool-loop log line.
- **D3 -> reference limits/edge handling** — `FR-003`/`FR-004`/`FR-005`: 100000-byte cap + `... (truncated)`; binary marker; directory `ERROR:`; ≤50-file cap; inline `ERROR:` text.
- **D4 -> no security/consent layer** — `FR-013`: settled exclusion stands; no `SafePath`, no consent, no new failure class.
- **D5 -> add `get_tree`** — `FR-008`/`FR-009`: `{path?, max_depth?, reason*}`, default `max_depth` 2, connector tree, `.git` not recursed.

### Still open (non-blocking — expected `/axb-technical-research` detail)

- **FR-004 wording of unreadable-path errors** — the exact `ERROR: …` text tied to the read failure; a research-level string, not a scope gap.
- **`read_files` multi-file output ordering** — request order (reference behaviour); confirm in research.
- **`reason` echo placement in the tool-loop line** — the exact `stderr` tail format; a research/interface detail under the existing round-008/010 logging contract.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Four stories stand (read-many-files · list-a-directory · view-a-tree · remove-summarisation). All scope decisions are operator-locked; only interface-level wording/ordering details remain for `/axb-technical-research`. `/axb-spec-by-example` can proceed.
