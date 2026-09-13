# Spec Quality Checklist: tellme Agent Tools & the Tool-Call Loop (round 008)

**Created**: 2026-09-12

**Feature Directory**: `specs/plans/008-agent-tools-and-tool-call-loop`

**Spec Path**: `specs/plans/008-agent-tools-and-tool-call-loop/spec.md`

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
- [x] This round's clarify questions are held to the budget (Rounds 1–3)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved

- **Q1 (R1) -> read-only filesystem tools** — `FR-001` = `list_files`, `read_files`; no writes/process/network.
- **Q2 (R1) -> widen the persisted turn** — `FR-016`; data-truth MODIFY (`/axb-data-plan`).
- **Q3 (R1) -> no boundary** — `FR-018`; `SafePath` + consent remain exclusions.
- **Q1 (R2) -> failure contract** — `FR-010`: new class phrase `tellme: the tool request failed` + exit `7`.
- **Q2 (R2) -> loop bound** — `FR-008`: `MAX_TOOL_LOOP`, default 1000, env/config override.
- **Round 3 -> visibility** — `FR-006`/`FR-007`, `NFR-002`: tool-loop activity is surfaced live during the run; `FR-017`: `-l` stays user/assistant-only.

### Still open (non-blocking)

- **FR-006 / tool-loop log stream** — **RESOLVED (extra Q2, Option 1)**: the live tool-loop log goes to the diagnostic stream (`stderr`); `stdout` stays the answer stream. The loop is discrete provider calls, not token streaming.
- **Tool-result size cap (NFR-007)** — the numeric cap; a `/axb-technical-research` determination.
- **Tool concurrency** — default sequential this round unless `/axb-technical-research` decides otherwise.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Four stories stand (tool answer · watch the loop · bounded loop · summarisation tool). All clarify items are resolved; only research-level defaults (tool-result size cap, tool concurrency) remain. `/axb-spec-by-example` can proceed.
