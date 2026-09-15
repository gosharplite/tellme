# Specification Quality Checklist: tellme tool-usage accounting

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/026-tool-usage-accounting`

**Spec Path**: `specs/plans/026-tool-usage-accounting/spec.md`

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
- [x] This round's clarify held to 1–3 questions (3 asked: outcome taxonomy · storage shape · surfacing channel)
- [x] Low-risk undecided details disclosed as assumptions (retention; flag-name; report wording)
- [x] Remaining undecided items marked as non-blocking

## Verifiability and success criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises/boundaries, not smuggled new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Record

- **Operator-locked decisions (this session).** **Q1 → 1**: three-way outcome `ok` / `error` / `timeout`; a failure = error **or** timeout (both structural at the loop: `err != nil` and the per-call `ctx` deadline). **Q2 → Others**: a **global** append-only JSONL log at `~/.tellme/tools-count.jsonl` (not the round-018 per-mode store), never reset by `--new`. **Q3 → 1**: a dedicated **offline reporting flag** printing the roll-up to `stdout`.
- **Q2 shape conformance.** The proposed `~/.tellme` global location is a **deliberate divergence** from the `TellMeHome`-is-the-namespace convention (no `~/.tellme` exists today; no `os.UserHomeDir()` usage). The YAML form was refined to **append-only JSONL** to match tellme's state convention and to be concurrency-safe without `flock`. Recorded in `spec.md` Assumptions and delegated to `/axb-data-plan`.
- **Hermeticity.** Because the write resolves the user home, tests must point the user home at a temp dir; an unresolvable home is a silent no-op. Captured as NFR-004 + an edge case.
- **Retention.** The global append-only log grows unbounded; compaction/rotation is a recorded **forward item**, not in scope.

## Ready verdict

- [x] Ready to proceed to later planning
- [ ] Still needs a high-impact requirement gap closed first

**Note**: Scope and mechanism are operator-locked (Q1–Q3). Remaining undecided items (retention, exact flag name, exact report wording) are low-risk and disclosed as non-blocking assumptions; none blocks planning. This plan may proceed to `/axb-spec-by-example` and `/axb-technical-research`.
