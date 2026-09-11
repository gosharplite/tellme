# Specification Quality Checklist: tellme Follow-up Cleanups (round 002)

**Created**: 2026-09-11

**Feature Directory**: `specs/plans/002-followup-cleanups`

**Spec Path**: `specs/plans/002-followup-cleanups/spec.md`

## How to use

- Check each item against the current `spec.md`.
- When an item does not pass, record the concrete gap and correction in "Issues & Correction Log".
- If `NEEDS CLARIFICATION` items remain, state explicitly whether they block subsequent planning.

## Content Completeness

- [x] All mandatory sections are completed
- [x] Feature topic, scope, and main flows are clearly expressed
- [x] No implementation technology, framework, or code detail is written as a requirement
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
- [x] This round's clarify budget stayed within 1–3 questions
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as non-blocking

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions express only premises and boundaries, with no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Correction Log

- **Clarify Q1 (user-ratified)**: `--json` is **removed entirely** (no machine-readable diagnostic mode). This **reverses round-001 FR-013**. Recorded as a `DELETE` intent for `/axb-dsl-refine`; the frozen round-001 package is left untouched.
- **Clarify Q2 (user-ratified), three items**: (1) quality-gate hardening — an ignored-error gate and a dependency-vulnerability gate; (2) pin the NFR-004 error-message wording; (3) pin the FR-014 exit-code numeric values.
- **F4 resolved by deletion, not by a new rule**: with `--json` no longer a flag, `tellme --json` / `tellme -d --json` fall to the existing unrecognized-flag usage contract; no new rejection mechanism is added.
- Round-001 exit-code numeric values (`0/2/3/4/5`) and the exact stderr strings become **fixed contract** this round (previously an implementation choice / unfixed wording, respectively).
- F9 (pure-helper unit tests) reframes round-001 research **Decision 5** (E2E-only): D5 governs the *acceptance* path; it does not forbid unit tests for pure helpers. Re-decision is owned by `/axb-technical-research` (recorded here as an RD follow-up, not a PM gap).
- Tooling for the ignored-error and vulnerability gates is deliberately **not** named in the spec (technology-neutral); it is chosen by `/axb-technical-research` and recorded in `techstack.md`.

## Ready Determination

- [x] Ready for subsequent planning
- [ ] Still requires high-impact requirement gaps to be filled

**Note**: A `/axb-clarify` round **was** performed (2 questions, within the 1–3 budget). Both decisions are ratified: `--json` removal (Q1) and the Q2 trio (gates + wording pin + exit-code pin). The only remaining undecided detail is the *tooling* for the two gates, which is intentionally deferred to `/axb-technical-research` — non-blocking.
