# Spec Quality Checklist: tellme Interactive Prompt Teardown & Submit-Surface Parity (round 023)

**Created**: 2026-09-15

**Feature Directory**: `specs/plans/023-interactive-prompt-teardown`

**Spec Path**: `specs/plans/023-interactive-prompt-teardown/spec.md`

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
- [x] Global requirements keep only cross-story or non-attributable entries
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps are escalated to `/axb-clarify`
- [x] This round's clarify questions are held to the budget (3 questions: Q1 teardown · Q2 surface parity · Q3 echo form)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (operator-locked, pre-specify clarify Q1–Q3)

- **Q1 → 1a — clear the frame on submit** — `FR-001`/`FR-002`/`FR-003`: the editor frame is cleared on submit/abort (reference parity), reversing the round-016 "always render".
- **Q2 → 2a — resume the standard surface** — `FR-005`/`FR-006`: the `-i` submit renders the input-capture line, the turn chrome, the live spinner, and the post-turn status, reversing the round-017 "no turn chrome" and the round-019 "`-i` excluded" rules.
- **Q3 → A′ — echo the prompt, keep the single captured line** — `FR-007`/`FR-008`/`FR-009`: the `-i` surface echoes the submitted prompt before the input-capture line; the positional / Ctrl+D surfaces do not echo; the reference's two-line form is not adopted.

### Still open (non-blocking — expected downstream detail)

- **Teardown mechanism** — clearing the editor frame on the submit/abort transition; the exact `View()`/model detail is a `/axb-technical-research`/interface concern (disclosed in Assumptions), not a scope gap.
- **E2E witness rework** — the round-016 capture ("final rendered frame") no longer holds once the frame is cleared; the new witness (frame absent post-submit · standard chrome present) is a `/axb-dsl-refine` + `/axb-tasks` detail.
- **Echo emission site** — the echoed prompt is emitted on `stderr` as part of the `-i` submit chrome (before the input-capture line); whether it lives in the CLI chrome emitter or the TUI runner is an interface detail.
- **`-r` interaction** — under `-r`, the spinner is suppressed but the echo and chrome still render, matching the other surfaces (FR-006); confirmed as the intended parity.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (the editor releases the terminal on submit · the `-i` submit continues on the standard turn surface). All three scope decisions are operator-locked; only interface-level teardown/echo/witness details remain. `/axb-spec-by-example` can proceed.
