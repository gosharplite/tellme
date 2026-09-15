# Specification Quality Checklist: tellme turn counter counts AI-endpoint calls

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/027-ai-call-turn-counter`

**Spec Path**: `specs/plans/027-ai-call-turn-counter/spec.md`

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
- [x] Low-risk undecided details disclosed as assumptions (persistence mechanism; display wording)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Operator-locked semantics (this session).** The counter's **unit** changes from completed user turns to **AI-endpoint calls**; the **cadence/format** stays as-is (one header + one `Ready` per prompt). `<N>` = prior calls + 1 (this prompt's first-call index), matching `tell-me-go`'s value at that prompt's first call. "Call" = an inference round; provider-internal retries do not count.
- **No clarify needed.** The three decisions that could have changed story boundaries, acceptance, or the success criteria were resolved directly with the operator in-session — (1) what the number counts (AI calls), (2) what "a call" is (inference round, not retries), (3) keeping the cadence as-is (no per-call header). No high-impact gap remains; the persisted-count mechanism is an RD decision delegated to `/axb-data-plan` + `/axb-technical-research`.
- **Cross-owner intent recorded.** This round **modifies** the round-017 turn-chrome truth (feature Rule + `chat/dsl.md` row + `techstack.md`) and **adds** a persisted per-turn call count to `specs/truth/data/**`. Those are intent only; the owners make the actual truth changes.
- **Reference parity is value-level, not cadence-level.** The reference emits a header/footer **per provider call**; tellme keeps one per prompt. This is a recorded divergence the operator accepted ("tellme as-is"); only the number's unit is aligned.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Semantics are operator-locked. Remaining open points (the exact persisted-count representation; the exact display wording, which stays tellme's current `Turn`) are low-risk and disclosed as assumptions; none blocks planning. This plan may proceed to `/axb-spec-by-example` and `/axb-technical-research`.
