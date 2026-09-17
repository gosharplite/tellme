# Specification Quality Checklist: tool reason sanitize (round 036)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/036-tool-reason-sanitize`

**Spec Path**: `specs/plans/036-tool-reason-sanitize/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (a presentation-contract defect: the model-authored reason breaks its own one-line log row)
- [x] No implementation technology, framework, or code detail is written as a requirement (the pure-formatter discipline is stated as a contract, not a code recipe; the cap value is an assumption)
- [x] Edge cases cover the high-risk situations (both `\n` and `\r`, both-surfaces blank suppression, mid-rune cap, absent reason, tool-less turn)
- [x] Key entities and success criteria are present, or their non-applicability is stated

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (US1 row integrity P1; US2 blank suppression P2)
- [x] Each user story is independently verifiable (US1: unit hostile fixtures; US2: unit emission-path pins)
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story
- [x] Global requirements hold only cross-story / non-attributable items (FR-007/FR-008)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's clarify is 1–2 questions (Q1 witness scope; Q2 blank-reason behaviour — both resolved with the operator)
- [x] Low-risk undecided details are disclosed as assumptions (cap value = 200; `oneLine` home; the ADR-0005 divergence record) rather than asserted as decided
- [x] Any remaining `NEEDS CLARIFICATION` is marked for blocking/non-blocking status (none remain — Q1/Q2 resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Demotion review (pre-clarify)**: two candidate decisions were **deliberately not escalated** per the specify rule — the **cap value** (a single numeric threshold → assumption `200`) and **`oneLine`'s home** (structure → research detail). The **ADR-0005 divergence record** is a discipline default, not a choice. Only the witness scope (Q1) and the blank-reason acceptance rule (Q2) reached `/axb-clarify`.
- **Witness narrowing disclosed (Q1 → Option 1)**: the E2E suite is blind to this defect class (round-035 residue rows look for a braille frame + phase status), and a count-based E2E carrier would silently miss either `\n` or `\r`; the unit hostile-fixture pin is the deterministic carrier. Flagged so the "no new E2E row" choice is not misread as incomplete coverage.
- **Cross-invariant guard**: the spec records that this fix is **independent** of round 035's spinner-yield fix (row intact vs. shared-row clear) so a reader cannot conflate them.
- **No product-code scope creep**: FR-007 records the unchanged surfaces (flags, exit codes, vocabulary, schemas, transport, records, `stdout`).

## Ready verdict

- [x] Ready for downstream planning
- [ ] Still needs high-impact requirement gaps closed

**Note**: a presentation-contract bug fix — `/axb-spec-by-example` is expected **NOOP** (the one-line promise is a contract; the violated guarantee is a code behaviour, not a PM journey). The load-bearing downstream steps are `/axb-technical-research` (cap-value ratification + the ADR-0005 divergence record + the `oneLine` home) and `/axb-dsl-refine` (the owned `chat/dsl.md` reason-row prose + cap-constant note), then `/axb-tasks` → `/axb-implement` (hostile-fixture Red → Green → Refactor).
