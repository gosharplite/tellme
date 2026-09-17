# Specification Quality Checklist: interactive prompt no-selection (round 037)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/037-interactive-prompt-no-selection`

**Spec Path**: `specs/plans/037-interactive-prompt-no-selection/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (the `-i` prompt opens with no suggestion highlighted — reference parity)
- [x] No implementation technology, framework, or code detail is written as a requirement (the `-1` sentinel is stated as a behaviour, not a recipe; the reset-on-refresh is a contract)
- [x] Edge cases cover the high-risk situations (single-item list, refresh-while-highlighted, empty list, submit/abort teardown, non-terminal)
- [x] Key entities and success criteria are present, or their non-applicability is stated

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (US1 no pre-selection P1; US2 first-`Tab` chooses the first suggestion P2)
- [x] Each user story is independently verifiable (US1: at-rest E2E + unit default pin; US2: unit `Next()` pin + the existing accept journey)
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story
- [x] Global requirements hold only cross-story / non-attributable items (FR-005/FR-006)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [x] This round's clarify is one round, four decisions (Q1 scope; Q2 reference-exact arithmetic; Q3 rule restatement; Q4 witness) — all resolved with the operator
- [x] Low-risk undecided details are disclosed as assumptions (cursor-sentinel semantics; the terminal-mode `ui/`; truth ownership) rather than asserted as decided
- [x] Any remaining `NEEDS CLARIFICATION` is marked for blocking/non-blocking status (none remain — Q1–Q4 resolved)

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Scope guard (Q1)**: the **empty-`Ctrl+S` divergence** (reference no-op vs tellme quit) is **verified** but deliberately **out of round scope**; recorded as a forward item and filed as a **live issue** (durable surface, round-035 G3 lesson). Flagged so the "cursor only" choice is not misread as a missed fix.
- **Contrast with round 036 (Q4/assumption)**: unlike round 036's structurally-blind E2E class, the cursor row is **directly observable**, so this round keeps a **valid E2E carrier** (the `Then` flips to expect 0) plus updated unit pins — not a unit-only narrowing.
- **Supersession disclosed**: the frozen round-016 at-rest acceptance rule is **superseded** by 037's rule; the round-016 package (and its `ui/screens/entry.txt`) stays frozen and is never edited.
- **No product-code scope creep**: FR-005 records the unchanged surfaces (ordering/sources, glyph, styling, editor, placeholder, debounce, flags, exit codes, vocabulary, records, `stdout`).

## Ready verdict

- [x] Ready for downstream planning
- [ ] Still needs high-impact requirement gaps closed

**Note**: a presentation-parity change. `/axb-spec-by-example` is **NOT** a NOOP (the at-rest highlight is a user-visible acceptance change). The load-bearing downstream steps are `/axb-technical-research` (the `techstack.md` suggestion-row MODIFY) → `/axb-system-analysis` (1 CLI end → `/axb-dsl-refine`; api/data NOOP) → `/axb-ui-plan` (terminal-mode at-rest frames) → `/axb-dsl-refine` (the flipped `Then` + the `chat/dsl.md` row MODIFY) → `/axb-tasks` → `/axb-implement` (unit pins Red → Green → Refactor; the E2E `Then` flip).
