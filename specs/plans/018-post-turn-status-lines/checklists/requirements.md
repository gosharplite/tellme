# Spec Quality Checklist: tellme post-turn status lines (round 018)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/018-post-turn-status-lines`

**Spec Path**: `specs/plans/018-post-turn-status-lines/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, note the concrete gap and the fix direction under "Issues & Corrections".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content Completeness

- [x] All mandatory sections are present
- [x] The feature topic, scope, and main flow are clear
- [x] No implementation technology / framework / code detail is written as a requirement
- [x] Edge cases cover the main high-risk situations
- [x] Key entities and success criteria are filled in (or their non-applicability is stated)

## User Stories & Requirement Routing

- [x] User stories are ordered by business value / delivery order
- [x] Each user story is independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story / non-attributable entries
- [x] No formal requirement is duplicated between a story and the global section

## Gaps & Clarification Strategy

- [ ] Only high-impact gaps were escalated to `/axb-clarify`
- [x] This round's clarify stayed within 1–3 questions
- [x] Low-risk undecided details are disclosed as `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as (non-)blocking

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises / boundaries, importing no new requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Corrections

- The PM-level decisions were settled by the five-question operator interview (lines, fields, three `$` semantics, line-2 source, `tokens.log` storage, presence) plus one pricing clarify question (config `MODELS` only; un-priced → `$0.0000`). The clarify budget is considered spent for this session.
- The two former `NEEDS CLARIFICATION` items are now **resolved** by the RD half: **A4** → `--new` archives/rotates `tokens.log` into `tokens.archive.jsonl`; **A8** → config-only `MODELS` (`{HIT, MISS, COMP}`), **not** env-overrideable.

## Ready Determination

- [x] Ready to proceed to downstream planning
- [ ] Still needs a high-impact requirement gap filled first

**Note**: Ready for `/axb-spec-by-example` + `/axb-technical-research`. The two `NEEDS CLARIFICATION` items (A4, A8) are technical and non-blocking for the plan half; the RD skills own them.
