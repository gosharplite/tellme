# Specification Quality Checklist: blank-reason owner + the loop's presentation de-coupling (round 046)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`

**Spec Path**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (the seam shape, its home, and the test-harness placement are explicit RD decisions — `research.md` D-x, not FRs)
- [x] Edge cases cover the main high-risk situations (absent reason; reason sanitizing to nothing; over-cap; tool-less turn; deferred final tail; nil presenter; a future second consumer of the predicate; the baseline at 0)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (the predicate has one owner) → US2 (the loop no longer owns presentation; baseline 1 → 0) → US3 (zero behavioural change)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-009, FR-010, FR-011, NFR-005, NFR-006)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] **Clarify round 1 LOCKED** (on [#108](https://github.com/gosharplite/tellme/issues/108)):
  - **C-R4-1 → (A)** an injected domain-typed presentation port (`agentport.ToolLineRenderer`, implemented in `internal/ui`); the loop keeps its write schedule + the round-045 yield bracket → **ADR 0014 untouched**.
  - **C-R4-2 → delete the dead site** (single owner): the predicate's owner is the port's `ReasonLine(reason) (line string, renders bool)`; the dead `callRenderer.OnCallEnd` re-check is removed.
  - **C-R4-3 → unit pins + the gate** (the baseline's `1 → 0`); **no** new E2E Example.
  - **C-R4-4 → a new ADR 0015** recording the loop/presenter ownership split.
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (mechanism is RD A2; ADR numbering A3; `/axb-dsl-refine` NOOP A5)

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (one owner; the dead site removed; the gate at 0; falsifiability witnesses red→green)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Two coupled axes held together deliberately.** The blank-reason predicate (a *policy*) and the loop→`ui` import (a *layer* edge) are one ledger item (#92 scope #4) and one DoD (the baseline 1 → 0 depends on the loop owning no presentation), so they are specified as one round with two P1 stories rather than split.
- **The yield route is a boundary, not part of R4.** `ADR 0014` owns the yield policy; C-R4-1 (C) is the only mechanism that would move the wrapping, and it is specified as an **explicit ADR amendment** — never a side effect.
- **Thin carrier acknowledged.** The round is behaviour-preserving, so `/axb-spec-by-example` is **NOOP** and the E2E suite is a weak witness — the witness is **unit pins + the gate** (A1 / FR-008 / SC-004/SC-005).
- **Truth-current is an obligation**: the techstack row(s) naming the loop's tool-line rendering / the `internal/ui` coupling are real **MODIFY**s, not NOOPs (FR-010).
- **The baseline shrink is an enabler, not a bypass.** Per ADR 0011 + round-040 TD-1, the compliant code change and the `tools/arch/baseline.txt` **1 → 0** edit land together; the anti-bypass rule (an emptied baseline with violations still fails; a stale line still fails) is preserved and pinned by witness (c).

## Ready determination

- [x] Ready to proceed to downstream planning
- [ ] A high-impact requirement gap must be closed first — **none remaining** (C-R4-1 … C-R4-4 locked)

**Note**: this branch runs the full pipeline (`/axb-specify` → `/axb-clarify` → `/axb-spec-by-example` (expected NOOP) → `/axb-technical-research` → `/axb-system-analysis` → `/axb-tasks` → `/axb-implement`); the plan + implementation land on the single `046-blank-reason-owner-and-presentation-decoupling` branch (round-043/045 single-PR precedent).
