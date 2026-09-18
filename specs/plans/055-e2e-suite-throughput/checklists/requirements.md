# Specification Quality Checklist: E2E suite throughput (round 055)

**Created**: 2026-09-19
**Feature Directory**: `specs/plans/055-e2e-suite-throughput`
**Spec Path**: `specs/plans/055-e2e-suite-throughput/spec.md`

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (US1 parallel scenarios; US2 fast subset)
- [x] No implementation/framework detail written as a *requirement* beyond the contract that must be named (godog `Concurrency` + `godog.paths`; the pinned version's exact option surface is grounding, not a requirement)
- [x] Edge cases cover the main high-risk situations (timing observers under load; the shared binary; per-scenario resources; `-count=1`; `StopOnFailure`/`Randomize`; small hosts)
- [x] Key entities and success criteria present

## User stories & requirement attribution

- [x] Ordered by value: US1 (parallel scenarios — the only one that shortens the **gate**) → US2 (subset — inner-loop convenience)
- [x] Both independently verifiable
- [x] Both carry acceptance scenarios
- [x] Story-specific FR under each (FR-001…004; FR-005/006)
- [x] No global requirements needed (no cross-story FR)

## Gaps & clarify strategy

- [ ] **Q1 OPEN — the parallelism default + timing protection** (blocks FR-001's default and FR-004's mechanism)
- [ ] **Q2 OPEN — the subset's selector + non-gating contract** (blocks FR-005's selector and the target name)
- [ ] **Q3 OPEN — the measurable bar + stability evidence N** (blocks SC-002/SC-003's numbers)
- [x] Clarify budget respected: **3** questions, asked **one at a time**, within the 1–3/round and ≤5/session budget
- [x] The proposed invariant (a subset selects for convenience, never excludes from the gate) is stated as **not open**
- [x] Non-blocking unknowns (the exact target name; the exact concurrency default) are surfaced rather than silently assumed

## Verifiability & success criteria

- [x] Acceptance covers the success path + the invariants (all 240 still executed; `Strict` still on; subset ≠ gate)
- [x] SC measurable (wall-clock **ratio/ceiling**, not a tight absolute — ADR-0010; Example-count equality; N green runs)
- [x] Assumptions are premises only (A1–A5; A5 deliberately refuses to silently exclude a non-isolated scenario)

## Issues & corrections

- The `Concurrency` field exists in the pinned godog **v0.16.0** and is clamped to ≥1 — so this is an options change, not a dependency bump.
- **Correction recorded vs. the pre-spec discussion**: godog **v0.16.0 has no name-filter flag**; the only scenario selectors are `Paths` and `Tags`, and there are **zero tags** in the tree ⇒ a tag-based subset would require **`specs/truth/**` edits** (`/axb-dsl-refine`), so the subset is scoped to `godog.paths` (Q2).
- Scenario isolation was **measured** before scoping parallelism (per-scenario `TELL_ME_HOME`+`HOME`; init-only shared state) — the decisive feasibility check.
- The #87 `di` flake is the precedent for why timing scenarios need an explicit protection decision (Q1), not an assumption.

## Ready determination

- [ ] Ready to proceed to `/axb-spec-by-example` — **NOOP anticipated** (A1): a tooling-only round with no user-visible journey
- [ ] Still needs the high-impact gaps resolved — **Q1/Q2/Q3 via `/axb-clarify` (one at a time)** before the acceptance/implementation decisions they name are settled

**備註**: the plan package + spec are complete in shape; the three open questions are **decision gaps**, and the round is otherwise ready to continue to `/axb-technical-research` once they are locked.
