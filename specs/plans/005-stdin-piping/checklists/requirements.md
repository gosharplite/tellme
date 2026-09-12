# Specification Quality Checklist: tellme Prompt Piping (round 005)

**Created**: 2026-09-12

**Feature Directory**: `specs/plans/005-stdin-piping`

**Spec Path**: `specs/plans/005-stdin-piping/spec.md`

## How to use

- Check each item against the current `spec.md`.
- When an item does not pass, record the concrete gap and correction in "Issues & Correction Log".
- If `NEEDS CLARIFICATION` items remain, state explicitly whether they block subsequent planning.

## Content Completeness

- [x] All mandatory sections are completed
- [x] Feature topic, scope, and main flows are clearly expressed
- [x] No implementation technology, framework, or code detail is written as a requirement (the TTY-detection mechanism and the stdin size cap are deferred to `/axb-technical-research`)
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria are present, or explicitly marked not applicable

## User Stories & Requirement Attribution

- [x] User stories are ordered by business value and delivery order
- [x] Each user story can be verified independently
- [x] Each user story includes acceptance scenarios
- [x] FR/NFR attributable to a single story are attached directly under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement is duplicated between the story and global sections

## Gaps & Clarification Strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify`
- [x] This round's clarify budget stayed within 1–3 questions (exactly 3 questions addressed)
- [x] Low-risk undecided details are disclosed via assumptions or non-blocking edge cases
- [x] All high-impact decisions are ratified with zero remaining blocking gaps

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises and boundaries, with no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Correction Log

- **Clarify Round 1 reissued and verified against `tell-me-go`** (the first draft's recommendations were corrected after checking the reference source):
  - **Q1 -> Option 1 (user-ratified)**: Piping only; keep raw output; defer `-r`. `tellme` has no renderer, so its output already equals `tell-me-go`'s `-r` (raw) output; `-r` today would be a no-op flag. Full rendering parity deferred.
  - **Q2 -> Option 3 (user-ratified)**: Combine — `args (joined by spaces)` + `"\n"` + piped stdin, matching `tell-me-go`'s main chat path. (The first draft recommended "arg wins", which matches only `tell-me-go`'s callback-worker path; corrected.)
  - **Q3 -> Option 2 (user-ratified)**: Adopt `tell-me-go`'s TTY-aware output contract — presentation suppressed when stdout is not a terminal. (The first draft recommended "TTY-agnostic byte-exact", which diverges from `tell-me-go`'s posture; corrected.)
- **Deferred to `/axb-technical-research` (not asked)**: the TTY-detection mechanism and the exact stdin size cap (a 1 MiB cap is assumed pending ratification). Recorded as assumptions in the spec.
- **In scope (folding issue #14 / F9)**: fast, isolated unit tests for the flag-parsing and I/O-mode-selection logic, as NFR-002 / SC-005.
- **Out of scope (deferred)**: any renderer or `-r`/raw-output flag; streaming output; `Turn`/`History` persistence and the session loop; tools/MCP/memory; TUI/`-i`; cost/metrics.

## Ready Determination

- [x] Ready for subsequent planning
- [ ] Still requires high-impact requirement gaps to be filled

**Note**: All 3 high-impact questions were resolved in Clarify Round 1 (reissued after the recommendations were checked against `tell-me-go` at the user's request). The two remaining undecided items (TTY-detection mechanism, stdin size cap) are technical decisions owned by `/axb-technical-research` and are recorded as explicit assumptions here, so they do not block `/axb-spec-by-example` or `/axb-technical-research`. The specification is complete, self-consistent, and ready for acceptance-criteria formalization.
