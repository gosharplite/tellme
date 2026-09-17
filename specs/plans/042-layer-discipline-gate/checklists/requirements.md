# Specification Quality Checklist: layer-discipline gate + violation baseline (round 042)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/042-layer-discipline-gate`

**Spec Path**: `specs/plans/042-layer-discipline-gate/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (guard form is an explicit RD decision, not an FR)
- [x] Edge cases cover the main high-risk situations (baseline-before-gate, host-independence, new vs baselined vs stale entries, build-tag hiding, the `tests/` scope)
- [x] Key entities and success criteria are present (or their inapplicability stated)

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (detect the new violation) → US2 (baseline + ratchet) → US3 (recorded truth)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-011, FR-012, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 rule model · Q2 package scope · Q3 stale-entry policy)
- [x] This round's clarify questions are capped at 1–3 (three)
- [x] Lower-impact undecided details disclosed as `NEEDS CLARIFICATION`/assumptions (guard form A4; ADR A5; `/axb-dsl-refine` NOOP-or-not A6)
- [ ] Remaining `NEEDS CLARIFICATION` marked as blocking or not — **pending `/axb-clarify` (Q1–Q3)**: FR-012's item is **blocking** the freeze of the rule model + scope; A4/A5/A6 are **non-blocking** (RD decisions)

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev`; red on a new violation)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-005)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- Baseline count (7) is stated **under the assumed** narrow rule model + production-only scope; it MUST be re-measured from the gate's own output before the spec is frozen (Q1/Q2 decide the exact set).
- The broad rule model would additionally surface `internal/agent → internal/ui` (and possibly `infrastructure/* → config`); whether those are violations or ranked-legal is Q1's substance.
- The `tests/e2e/**` harness imports `internal/infrastructure/**`; whether it is governed by the gate is Q2.

## Ready determination

- [ ] Ready to proceed to downstream planning
- [x] A high-impact requirement gap must be closed first (`/axb-clarify` Q1–Q3)

**Note**: plan half only — stop before `/axb-tasks` and `/axb-implement`. Everything up to and including `/axb-dsl-refine` is in scope for this branch.
