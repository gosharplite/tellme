# Spec Quality Checklist: round 098 `098-param-swap-guard`

**Created Date**: 2026-10-27

**Feature Directory**: `specs/plans/098-param-swap-guard`

**Spec Path**: `specs/plans/098-param-swap-guard/spec.md`

## How to Use

- Check each item against the current `spec.md` content.
- If an item fails, add the concrete gap and fix direction in "Issues & Fix Log".
- If `NEEDS CLARIFICATION` remains, explicitly state whether it blocks downstream planning.

## Content Completeness

- [x] All required sections are complete
- [x] The feature topic, scope, and main flows are clearly expressed
- [x] No implementation technologies, frameworks, or code details are written as requirements
- [x] Edge cases cover the main high-risk scenarios
- [x] Key entities and success criteria are filled in, or explicitly explained as not applicable

## User Stories & Requirement Attribution

- [x] User stories are ordered by business value and delivery order
- [x] Every user story can be independently verified
- [x] Every user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story hang directly under that story
- [x] Global requirements keep only cross-story or un-attributable entries
- [x] Formal requirements are not duplicated between the story area and the global requirements area

## Gaps & Clarify Strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's clarify question count is kept to 1 to 3 questions (zero needed — the operator's
      directive and issue #210 fix the scope, and the four open verification questions in the issue were
      **answered by measurement** in `research.md` rather than by asking)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked whether they block downstream planning (none
      remain)

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] All normative items (FR / NFR / SC / EC) are annotated with a Verification Intent
- [x] Assumptions express only premises and boundaries, without smuggling in new requirements
- [x] Requirements, edge cases, key entities, and success criteria are consistent with each other

## Issues & Fix Log

- **The round's carrier is a check, not a screen.** The three user stories describe a *quality instrument*
  (Story 1), the *tree it measures* (Story 2) and the *file layout it lands in* (Story 3). Each is
  independently verifiable: the gate's red run, the E2E contract's unchanged counts, and `wc -l` + the
  unchanged architecture baseline. No story depends on another being delivered first.
- **No `NEEDS CLARIFICATION` was needed.** Issue #210 already publishes a labelled baseline, a rule
  definition and a DoD; the four `Open verification` questions it lists (§1 script-first, §3 arm (c)
  option 1 vs 2, §5 arm (d) implementability, §6 model impact) are **research** questions answerable from
  the tree — `research.md` answers each with a measurement instead of an interview.
- **Scope narrowed deliberately at `n ≥ 4`.** Widening the arity threshold (or adding a result-list arm)
  would red ~50 more sites and turn a guard round into a repo-wide renaming exercise; ADR 0068 §Forward
  records the remainder with its measurement.

## Ready Judgment

- [x] Ready to proceed to downstream planning
- [ ] High-impact requirement gaps still need filling first

**Note**: the plan proceeds to `/axb-spec-by-example` → `/axb-technical-research` (the UI plan is skipped:
a plain line-oriented CLI with no governed UX surface this round).
