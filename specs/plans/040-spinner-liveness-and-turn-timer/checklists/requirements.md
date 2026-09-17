# Specification Quality Checklist: spinner liveness while a command streams + a per-turn elapsed timer (round 040)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/040-spinner-liveness-and-turn-timer`

**Spec Path**: `specs/plans/040-spinner-liveness-and-turn-timer/spec.md`

## How to use

- Check each item against the current `spec.md`.
- Record any gap and its fix direction under "Issues & Fix Log".
- If any `NEEDS CLARIFICATION` remains, state whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (two folded spinner workstreams: issue #82 liveness while streaming; issue #83 dual elapsed timer)
- [x] No implementation technology, framework, or code detail is written as a requirement (the idle-gap mechanism and the per-call boundary are stated as behaviour; the `ui`/`spinner.go` home is an enabler, not a requirement)
- [x] Edge cases cover the high-risk situations (tool-less turn, quiet streaming command, AI-call boundary at a resume, very long turn, resize, non-TTY/`-r`/`-i`, no-usage)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by business value / delivery order (US1 liveness-while-streaming P1 — the anchor issue #82; US2 dual timer P2 — additive readability)
- [x] Each user story is independently verifiable (US1: unit spin-cycle pin + E2E resumed frame; US2: unit arithmetic pin + E2E two-figure assertion)
- [x] Each user story carries acceptance scenarios
- [x] Story-scoped FR / NFR are attached under the story (FR-001..FR-004; FR-005..FR-008)
- [x] Global requirements hold only cross-story / non-attributable items (FR-009/FR-010/FR-011)
- [x] No formal requirement is duplicated between the story and the global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps escalate to `/axb-clarify`
- [ ] This round's clarify is **partially** resolved: QB1 (reset boundary) is operator-locked; QA1–QA3 (WS-A mechanism/threshold/scope) and QB2/QB3 (format/phases) are **proposed** and pending operator confirmation at the pre-`/axb-tasks` review gate
- [x] Low-risk undecided details are disclosed as assumptions
- [ ] `NEEDS CLARIFICATION` remains for QA1–QA3/QB2/QB3 — **does not block** the plan+truth half (spec-by-example / research / analysis / dsl-refine proceed against the proposed readings); the operator confirms them at the gate before `/axb-tasks`

## Verifiability & success criteria

- [x] Acceptance scenarios verify the main success path
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only (no smuggled requirements)
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & Fix Log

- **Anchored + folded**: issue [#82](https://github.com/gosharplite/tellme/issues/82) is the anchor (WS-A); issue [#83](https://github.com/gosharplite/tellme/issues/83) is the folded second workstream (WS-B) — both requested by the operator on 2026-09-17 (the round-038 #78+#76 / round-039 #80+spacing precedent).
- **Supersession, not amendment**: WS-A changes round-034 FR-012/G8 + **ADR 0005 D7** (the whole-block pause). Recorded via a **new ADR superseding D7** (an `Accepted` ADR is immutable but for its `Status`). WS-B **amends** round-019 D4 ("never reset").
- **Recorded divergence**: `tell-me-go` shows no turn-scoped **dual** timer and never re-activates a spinner inside its streamed output; both are recorded divergences.
- **Interaction of the two workstreams**: a resumed indicator during a streaming block must keep the turn-scoped total and the per-call **model-call** figure **consistent across the pause/resume** — pinned by an E2E that combines a quiet command with a second call.
- **Witness layering**: WS-A's idle-gap mechanism is observable at both the unit layer (injected `newTicker`, incl. the race + anti-vacuity cases) and the E2E layer (a **real** idle gap: a small nonzero forced threshold + a **child** `sleep` exceeding `N + 2·P` with margin); WS-B's arithmetic is a unit-layer pin (injected `now`) plus an E2E two-figure assertion.
- **Width safety**: the two-figure line is longer than the round-019/025 line, so the round-025 rune-based `rowsForLine`/`eraseRows` bound is the carrier for "no residue" and must be re-witnessed.

## Ready verdict

- [x] Ready for downstream planning (plan + truth half)
- [ ] Still needs high-impact requirement gaps closed **before `/axb-tasks`** (QA1–QA3/QB2/QB3 confirmation)

**Note**: two `stderr`-presentation workstreams on the round-019/025 spinner. `/axb-spec-by-example` is **NOT** a NOOP (both are operator-visible). `/axb-ui-plan` is **skipped** (line-oriented, no chrome change). The load-bearing downstream steps are `/axb-technical-research` (a `techstack.md` MODIFY + a **new ADR superseding ADR 0005 D7** + the reference divergence check) → `/axb-system-analysis` (1 CLI end → `/axb-dsl-refine`; api/data NOOP) → `/axb-dsl-refine` (amend the "paused while streaming" Rule + a new dual-timer Rule + `chat/dsl.md` rows) → **STOP for the operator review gate** → `/axb-tasks` → `/axb-implement`.
