# Specification Quality Checklist: yield-policy owner + `LoopObserver` hook split (round 045)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/045-yield-policy-owner`

**Spec Path**: `specs/plans/045-yield-policy-owner/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (the exact `YieldController` method set, the coordinator's field change, and the test-harness placement are explicit RD decisions — `research.md` D-x, not FRs)
- [x] Edge cases cover the main high-risk situations (gated-off nil spinner; the deferred final tail; `End` while a line write is in flight; the one-concurrent-block record; `Stop()` vs a yield; a future out-of-owner yield primitive)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (the yield policy has one named owner) → US2 (the `LoopObserver` hook split) → US3 (zero behavioural change)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-010, FR-011, FR-012, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` — **six decisions** locked one at a time (the standard cap was exceeded deliberately: the coupled pair has a policy axis *and* a port axis, and each answer constrained the next)
- [x] No remaining `NEEDS CLARIFICATION` — round-1 answers **locked**:
  - **C-R3-1 → (1)** a single `internal/ui` yield owner (`YieldController`).
  - **C-R3-2 → (A)** the loop keeps calling the port (`YieldIndicator`/`RestoreIndicator`).
  - **C-R3-3 → replace** the log-named pair (no alias kept).
  - **C-R3-4 →** the baseline stays **1** (R4 owns 1 → 0).
  - **C-R3-5 →** unit ordering pins only (no new E2E Example).
  - **C-R3-6 →** the two `#92` records stay records.
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (mechanism is RD A2; ADR A3; `/axb-dsl-refine` NOOP A5)

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (one owner; the split exists; orderings preserved; falsifiability witnesses red→green)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-007)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **C-R3-3 deviates from the #105 recommendation, deliberately.** The issue's candidate C-R3-3 *recommended* keeping `BeforeToolLog`/`AfterToolLog` as intent aliases for H1 and *adding* the general pair. Adopted instead: **replace** the pair (the #105 *Goal* wording — *"the overloaded `BeforeToolLog`/`AfterToolLog` pair is replaced by intent-named hooks"*). Keeping four names with two meanings would re-introduce the very overload the split removes; the port now states only the loop's **route**, and the **policy** lives with `YieldController`. Recorded in `spec.md` §Locked decisions C-R3-3.
- **C-R3-4 boundary made explicit.** R3 does **not** lift the `internal/agent -> internal/ui` baseline entry: that entry is driven by the four formatters + `ui.ToolReasonRenders`, **not** the yield port. Removing it is **R4**'s DoD (1 → 0). Recorded so the ratchet cannot silently regress.
- **Thin carrier acknowledged.** The round is behaviour-preserving, so `/axb-spec-by-example` is **NOOP** and the E2E suite is a weak witness — the witness is the **unit ordering pins** + the ADR (A1 / FR-009 / SC-003).
- **Truth-current is an obligation**: the techstack rows naming the observer seam and the spinner's yield behaviour are real **MODIFY**s, not NOOPs (FR-011).
- **Stale-row guard.** A renamed seam can leave a stale truth row passing green (the audit checks *feature → row* only), so `/axb-dsl-refine`'s NOOP is evidenced by a **grep** of the truth tree for the old hook names — not by the audit alone.

## Ready determination

- [x] Ready to proceed to downstream planning
- [ ] A high-impact requirement gap must be closed first — **none remaining** (C-R3-1 … C-R3-6 locked)

**Note**: this branch runs the full pipeline (`/axb-specify` → `/axb-spec-by-example` (expected NOOP) → `/axb-technical-research` → `/axb-system-analysis` → `/axb-tasks` → `/axb-implement`); the plan + implementation land on the single `045-yield-policy-owner` branch (round-043 single-PR precedent).
