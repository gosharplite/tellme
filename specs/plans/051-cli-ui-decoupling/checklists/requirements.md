# Specification Quality Checklist: close [#101](https://github.com/gosharplite/tellme/issues/101) — de-couple `internal/cli` from `internal/ui` (last RULE-E residual) + F-6/F-7/F-8 (round 051)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/051-cli-ui-decoupling`

**Spec Path**: `specs/plans/051-cli-ui-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (close #101: the `cli → ui` edge → baseline 0; plus F-6/F-7/F-8)
- [x] No implementation/framework detail written as a *requirement* (the port shape/homes are explicit RD decisions behind FR-001/FR-002 / RULE-C)
- [x] Edge cases cover the main high-risk situations (partial removal; port-shape leak; formatter single-ownership; stateful lifecycle; degraded render; interface-seam `Validate()`; no release valve)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (CLI names no `internal/ui` id; baseline → 0) → US2 (the two ratchet removals) → US3 (F-6/F-7/F-8)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-009…FR-012)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [ ] **Q1 → OPEN** — the closing-programme shape (A re-cut sub-slice 1 / B the `→ ui` collapse / C F-6/F-7/F-8 first / D mega-round). **Blocks the story split + the round's DoD.**
- [ ] **Q2 → TBD** — the value-type homes
- [ ] **Q3 → TBD** — the `→ ui` inversion mechanism
- [ ] **Q4 → TBD** — F-6/F-7/F-8 folding + F-8 acceptability
- [x] Questions asked **one at a time**; capped at 1–3 per round (this round may need Q1–Q4 given the programme size — disclosed)
- [x] High-impact gap scoped to a single first question (Q1), with options A/B/C/D
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (ADR number A4; NOOP set A5/A6; sanctioned set fixed A7)
- [ ] No remaining `NEEDS CLARIFICATION` — **not yet**: Q1 pending

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (baseline 0; CLI `→ ui` refs = 0; a partial removal → red; byte-identical streams)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 1 today** (measured @ `dev` `310def4`): `internal/cli -> internal/ui`. This round/programme targets **1 → 0**.
- **⚠️ TWO ratchet removals, not one** (ADR 0018 fold F-4): when the `→ ui` edge disappears the gate reds **twice** — the RULE-E baseline line **and** the RULE-F `couplingSurface["internal/cli -> internal/ui"]` key (19 identifiers).
- **Not edge-sized** (ADR 0017 §Forward): ≈20 call sites · 3 value types · 4 stateful objects · 8 formatters → a **re-cut** is expected (Q1).
- **No `internal/pkg`** in tellme → the "pure utility" escape does not exist; the sanctioned set is fixed (ADR 0016 D1).
- **F-6/F-7/F-8** verified still open (F-7 still a bare `func()`; F-8 still a struct-of-funcs).
- **Interface-seam `Validate()` caveat** (ADR 0017 §Forward): a new **interface**-typed `Dependencies` field needs its own assertion.
- **No release valve** at baseline 0 (ADR 0011/0016).
- **Witness discipline**: the witness is the **gate + unit seams** (FR-011), never the E2E suite (#92 AC5).
- **Atomicity** — per round: inversion/re-home + truth/ADR + the removals land as one PR (NFR-002-style). **No new Makefile target.**

## Ready determination

- [ ] Ready to proceed to downstream planning — **blocked on Q1** (clarify round 1 OPEN)
- [x] A high-impact requirement gap must be closed first — **Q1** (the closing-programme shape)

**Note**: plan package is the `/axb-specify` skeleton; it is **paused pending `/axb-clarify` Q1**. Next pipeline step after Q1–Q4 close: `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
