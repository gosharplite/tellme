# Spec Quality Checklist: Round 097 — `097-prompt-ack-blue-accent`

**Created Date**: 2026-10-22

**Feature Directory**: `specs/plans/097-prompt-ack-blue-accent`

**Spec Path**: `specs/plans/097-prompt-ack-blue-accent/spec.md`

## How to Use

- Check each item against the current `spec.md` content.
- If an item fails, add the concrete gap and fix direction in "Issues & Fix Log".
- If `NEEDS CLARIFICATION` remains, explicitly state whether it blocks downstream planning.

## Content Completeness

- [x] All required sections are complete
- [x] The feature topic, scope, and main flows are clearly expressed
- [x] No implementation technologies, frameworks, or code details are written as requirements
      *(the spec names files/constants only as **grounding evidence** for the recorded decision, not
      as requirements; the requirements themselves are behavioural)*
- [x] Edge cases cover the main high-risk scenarios (EC-001..EC-004: redirected stream, `-r`,
      non-prompt paths, `turns.log`)
- [x] Key entities and success criteria are filled in

## User Stories & Requirement Attribution

- [x] User stories are ordered by business value and delivery order (Story 1 = the accent itself;
      Story 2 = its uniform reach, which is what justifies the element choice)
- [x] Every user story can be independently verified
- [x] Every user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story hang directly under that story
- [x] Global requirements keep only cross-story or un-attributable entries (**none** — Stories 1–2
      carry the round)
- [x] Formal requirements are not duplicated between the story area and the global requirements area

## Gaps & Clarify Strategy

- [x] Only high-impact gaps escalate to `/axb-clarify` (**none escalated** — the colour, the scope,
      and the surface set were locked by the operator in issue #204)
- [x] This round's clarify question count is kept to 1 to 3 questions (**0**)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A4)
- [x] Remaining `NEEDS CLARIFICATION` items are marked whether they block downstream planning
      (**none outstanding**)

## Verifiability & Success Criteria

- [x] Acceptance scenarios suffice to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] All normative items (FR / NFR / SC / EC) are annotated with a Verification Intent
- [x] Assumptions express only premises and boundaries, without smuggling in new requirements
- [x] Requirements, edge cases, key entities, and success criteria are consistent with each other

## Issues & Fix Log

- **ISSUE-1 (resolved in-spec)** — the first draft of this theme (issue #204 revision 1) coloured the
  `-i` **prompt text**, which appears blue/white/absent across four surfaces. The spec now scopes the
  accent to the **acknowledgement** instead, and records the withdrawn alternative in §1 so the
  reasoning is not re-litigated downstream.
- **ISSUE-2 (resolved in-spec)** — revision 2 proposed **green**; the operator rejected it as "the
  dollar accent", which `internal/ui/metrics.go:38-48` confirms literally. The spec locks **blue**
  and records why green is unavailable.
- **ISSUE-3 (design constraint carried, not a gap)** — ADR 0023 **D3** is an operator-locked
  four-element set and ADR bodies are immutable ⇒ the fifth element needs a **new ADR** (A2). This
  is a recorded decision, not an open question.

## Ready Judgment

- [x] Ready to proceed to downstream planning
- [ ] High-impact requirement gaps still need filling first

**Note**: the round is deliberately narrow (one accent on one line). Its weight sits in the
**witness quality**: the accent must be shown present *and* discriminatingly absent (off the gate,
under `-r`), and the surface uniformity must be shown on at least two distinct surfaces, because
uniformity is the reason this element was chosen over the prompt text.
