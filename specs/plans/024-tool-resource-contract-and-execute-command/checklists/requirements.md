# Spec Quality Checklist: tellme Tool Resource Contract, `execute_command` & Reader Retrofit (round 024)

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/024-tool-resource-contract-and-execute-command`

**Spec Path**: `specs/plans/024-tool-resource-contract-and-execute-command/spec.md`

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
- [x] Global requirements keep only cross-story or non-attributable entries (the tool resource contract)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps are escalated to `/axb-clarify`
- [x] This round's clarify questions are held to the budget
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (operator-locked, session 2026-09-15 — D1–D7)

- **D1–D3 — no security · no Windows · bash-first** → `FR-005`/`FR-007` + the whole contract: no gate; POSIX only; `bash -c`; `pipe_commands` omitted.
- **D5 — three-tier contract** → `FR-013`/`FR-014`/`NFR-001`: uniform params; **default + param + ceiling**; central clamp.
- **D6 — one aggregate bound** → `FR-009`/`FR-010`: whole-file reads, no per-file cap, no fair-share, no paging; the shell slices big files.
- **D7 — scope** → the two stories + the retrofit FRs.

### Resolved (clarify Q1–Q3, session 2026-09-15)

- **Q1 → 1 — non-zero-exit semantics** (`FR-005a`): a non-zero exit is a **successful tool result carrying the exit code**; the loop continues. A **tool failure** (a **non-nil error**) is reserved for a **tool/argument error** only — a **timeout** is a **result**, not a failure (`FR-018`). (Reference parity.)
- **Q2 → 1 — `output_file`/`append` in scope** (`FR-005b`): `execute_command` accepts `output_file` + `append`, capturing large output to disk (the escape hatch for output bigger than the bound).
- **Q3 → 1 — mechanism only**: the default/ceiling *mechanism* (default derived from the resolved budget) is locked; the concrete numbers are a `/axb-technical-research` decision.

### Still open (non-blocking — expected downstream detail)

- **Timeout defaults** — per-tool default/ceiling numbers: a `/axb-technical-research` decision.
- **Truncation-marker wording** — exact markers for the command/reader bounds: a `/axb-dsl-refine` detail.
- **Central-enforcement site** — the decorator/executor seam that clamps bounds and applies the timeout: an interface detail.
- **Reader-cap retirement mechanics** — how the round-021 `readAggregateCap`/`readMaxPerFile` constants are retired: implementation detail (the round-021 package stays frozen).

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (run a shell command · bounded reader results), plus the cross-story tool resource contract. All seven design decisions (D1–D7) are operator-locked and clarify **Q1–Q3 are resolved** (non-zero exit = success result carrying the exit code; `output_file`/`append` in scope; default/ceiling mechanism locked, numbers deferred to research). No blocking gap remains — `/axb-spec-by-example` can proceed.
