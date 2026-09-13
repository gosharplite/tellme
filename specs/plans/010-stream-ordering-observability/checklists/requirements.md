# Spec Quality Checklist: tellme Stream-Ordering Observability (round 010)

**Created**: 2026-09-13

**Feature Directory**: `specs/plans/010-stream-ordering-observability`

**Spec Path**: `specs/plans/010-stream-ordering-observability/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement (the merged-capture *mechanism* is stated as an observable requirement, not a named technology)
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

- **Q1 → Option 2 (scope)** — required orderings are (a) payload-status bracketing (`FR-001`–`FR-004`) and (b) tool-loop-precedes-answer (`FR-005`–`FR-006`); the degraded-render warning is **incidental** (`FR-011`).
- **Q2 → Option 1 (witness)** — merged single-buffer capture (`FR-007`); no product change.
- **Q3 → Option 1 (layering)** — asserted at both the unit and E2E layers (`FR-008`).

### Still open (non-blocking)

- **Merged-capture mechanism** — the exact harness plumbing (a `RunMerged`-style variant sharing one buffer vs a `2>&1`-equivalent); a `/axb-technical-research` determination. No new dependency (`NFR-003`).
- **No-final-answer ordering semantics** — how the contract reads when a turn ends without an answer (tool-loop bound / tool error); a `/axb-dsl-refine` wording determination (disclosed as an edge case, non-blocking).
- **Incidental-interleave record** — the location where the degrade warning is documented as incidental (assumption in the spec vs a DSL/rationale note); a `/axb-dsl-refine` determination.
- **Unit emit-order pattern** — whether a general "ordered writes" unit helper is introduced or the existing `runTurn` order test is extended; a `/axb-tasks`/`/axb-implement` determination.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (payload-status bracketing · tool-loop-precedes-answer); the witness + layering + falsifiability are global requirements. All clarify items are resolved; only research/DSL-level determinations remain. `/axb-spec-by-example` and `/axb-technical-research` can proceed.
