# Specification Quality Checklist: spinner tail residue (round 035)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/035-spinner-tail-residue`

**Spec Path**: `specs/plans/035-spinner-tail-residue/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (a display defect: the per-call tail collides with the live spinner frame)
- [x] No implementation technology, framework, or code detail is written as a requirement (the mechanism — clear-only vs clear+resume — is deferred to research)
- [x] Edge cases cover the high-risk situations (multi-call turn, tool-less turn, bound-reached turn, gated-off run, narrow terminal)
- [x] Key entities and success criteria are present, or their non-applicability is stated

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (a single P1 story)
- [x] Each user story is independently verifiable
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story
- [x] Global requirements hold only cross-story / non-attributable items (FR-006/FR-007)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's clarify is 1–4 questions (Q1–Q4, all resolved with the operator)
- [x] Low-risk undecided details are disclosed as assumptions (the re-activation point) rather than asserted
- [x] Any remaining `NEEDS CLARIFICATION` is marked for blocking/non-blocking status (none remain — Q1–Q4 resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Q1–Q4 resolved** with the operator one at a time (feedback approach; pure bug fix; siblings out; E2E + unit witness).
- **Fix-mechanism disclosure**: FR-004 deliberately leaves the indicator's re-activation point to research, because a naïve resume-after-tail would relocate the residue to the next call's frame write; the acceptance scenarios (SC-001/SC-002) are the arbiter. Flagged so it is not misread as an unspecified requirement.
- **No product-code scope creep**: FR-006 records the unchanged surfaces (flags, exit codes, formats, cadence, records, `stdout`).

## Ready verdict

- [x] Ready for downstream planning
- [ ] Still needs high-impact requirement gaps closed

**Note**: this is an implementation-vs-truth bug fix — the violated rules already exist as executable truth, so `/axb-spec-by-example` is NOOP and the load-bearing downstream phases are `/axb-technical-research` (witness + mechanism) and `/axb-dsl-refine` (carrier re-anchor).
