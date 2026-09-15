# Specification Quality Checklist: user-global interactive prompt log

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/028-user-global-prompt-log`

**Spec Path**: `specs/plans/028-user-global-prompt-log/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and fix direction under "Issues & Fix Record".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks later planning.

## Content completeness

- [x] All mandatory sections completed
- [x] Feature theme, scope, and main flow expressed clearly
- [x] No implementation technology, framework, or code detail written as a requirement
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria present (or explicitly noted N/A)

## User stories and requirement attribution

- [x] User stories ordered by business value / delivery order
- [x] Each user story independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable items
- [x] No formal requirement duplicated between a story and the global section

## Gaps and clarification strategy

- [x] Only high-impact gaps escalated to `/axb-clarify`
- [x] This round's clarify held to 1–3 questions (0 asked — see below)
- [x] Low-risk undecided details disclosed as assumptions (exact seed trigger wording; `HOME` resolution; the seed's source-across-environments subtlety)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Operator-directed scope (this session).** The `-i` shared prompt log moves to the user-global `~/.tellme/global_prompts.jsonl`; when it is absent, the existing `<TELL_ME_HOME>/output/global_prompts.jsonl` is copied into it verbatim. The record shape, the `-i`-only write rule, and the interactive chrome are unchanged.
- **No clarify needed.** The operator gave the target path and the seed rule verbatim; the two decisions that could have changed story boundaries or acceptance were settled in-session — (1) the user-global file is the sole read/write target (the env-scoped file becomes a one-time seed source only), and (2) the seed is a verbatim copy fired only while the destination is absent. No high-impact gap remains; the persistence/migration mechanism is an RD decision delegated to `/axb-data-plan` + `/axb-technical-research`.
- **Recorded contract change (must reach the truth owners).** This round **MODIFIES** the round-015/016 truth that the log is "the same file the other personas use" — i.e. shared byte-for-byte with `tell-me-go`. After this round tellme does **not** share with `tell-me-go` (which keeps `<TELL_ME_HOME>/output/global_prompts.jsonl`), and tellme skips `tell-me-go`'s legacy locations. The sharing unit changes from *per-`TELL_ME_HOME` (+ with `tell-me-go`)* to *per-user (across all tellme environments, not with `tell-me-go`)*. Flagged to the operator before opening the round; recorded as an explicit operator-directed divergence in `spec.md` → Assumptions. `/axb-data-plan`, `/axb-technical-research` and `/axb-dsl-refine` must carry this into their truth edits and the `truth-delta.md`.
- **Single-source seed subtlety (disclosed, non-blocking).** Only the first interactive run after the move (in whichever environment is active while the user-global file is absent) seeds the global log; other environments' env-scoped logs are not merged. Low-risk and operator-accepted (verbatim copy), so disclosed in `spec.md` → edge cases rather than escalated.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope is operator-directed and the seed rule is explicit. Remaining open points (the exact `HOME`-resolution failure behaviour; the exact seed trigger wording) are low-risk and disclosed as assumptions; none blocks planning. This plan may proceed to `/axb-spec-by-example` and `/axb-technical-research`.
