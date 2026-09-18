# Specification Quality Checklist: de-couple `internal/cli` from the turn loop — sub-slice 2: invert the `AgentLoop` construction/execution into an injected domain port (round 050)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/050-agentloop-construction-inversion`

**Spec Path**: `specs/plans/050-agentloop-construction-inversion/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (sub-slice 2 — invert the `AgentLoop` construction/execution into a domain port; the `cli → agent` edge removed; **baseline 2 → 1**)
- [x] No implementation/framework detail written as a *requirement* (port/adapter homes and the alias/interface choice are explicit RD decisions behind FR-002 / RULE-A·C)
- [x] Edge cases cover the main high-risk situations (partial inversion; interface-seam `Validate()` blindness; `ui` drift; observer ordering; `Now`/`ToolUsage` seams; build-tag scope; cycles)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (loop via injected domain port; CLI `→ agent` refs = 0) → US2 (the two ratchet removals + `→ ui` untouched) → US3 (ADR + truth + caveats)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-010…FR-012)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] **Q1 → A LOCKED** — port-only inversion: the `→ agent` construction is inverted into a domain port; the `ui` wiring (Lines renderer, spinner, coordinator, composite observer) **stays in `internal/cli`** by design; baseline **2 → 1**; the `→ ui` edge is the later slice
- [x] **Q2 → (i) LOCKED** — domain interface `agentport.Loop` (`Run(...) (Result, error)`) + domain `LoopSpec` + a **func-typed** factory in `deps.Dependencies` (`LoopFactory`), adapter `agent.NewLoop` in `internal/agent`; `Validate()` already covers a func-typed field (no interface-seam assertion needed)
- [ ] **Q3 → TBD** — `Lines`/observer ownership
- [x] Questions asked **one at a time**; capped at 1–3 per round
- [x] High-impact gap scoped to a single first question (Q1), with options A/B/C
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (port names A3; ADR number A4; NOOP set A5/A6)
- [ ] No remaining `NEEDS CLARIFICATION` — **not yet**: Q1 pending → clarify round 1 **OPEN**

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev` at baseline **1**; the CLI `→ agent` refs = 0; a partial inversion → red; `→ ui` surface byte-stable at 19)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 2 today** (re-measured 2026-09-18 @ `dev` `684e41e`): `internal/cli → internal/agent`, `internal/cli → internal/ui`. This round targets **2 → 1**.
- **⚠️ TWO ratchet removals, not one** (ADR 0018 §Forward, fold **F-4**): when the `→ agent` edge disappears the gate reds **twice** — the RULE-E baseline line **and** the RULE-F `couplingSurface["internal/cli -> internal/agent"]` key both go stale. Both are folded deliberately (FR-005/FR-006; SC-006).
- **R-1 cross-slice coupling** (ADR 0018 §Forward): the surviving `→ agent` site shares `cli.go:699-745` with **three** `→ ui` refs — Q1 must decide the split.
- **`Validate()` interface-seam caveat** (ADR 0017 §Forward): an interface-typed field on `deps.Dependencies`/`cli.Options` is invisible to `Kind()==reflect.Func` → needs its own assertion (FR-009).
- **RULE-F metric** (ADR 0018 N-8): distinct **selectors**, not call sites; production-only scope (F-3).
- **R5.x lineage**: round 047 = R5.1 · round 048 = R5.2 · round 049 = R5.3/sub-slice 1 · **this round = sub-slice 2** (the baseline-moving round).
- **Witness discipline**: the witness is the **gate + the identifier-count check + unit seams** (FR-011), never the E2E suite (#92 AC5).
- **Atomicity** — inversion + truth/ADR + the two removals land as one PR (NFR-002). **No new Makefile target.**

## Ready determination

- [ ] Ready to proceed to downstream planning — **blocked on Q1** (clarify round 1 OPEN)
- [x] A high-impact requirement gap must be closed first — **Q1** (the R-1 assembly split)

**Note**: plan package is the `/axb-specify` skeleton; it is **paused pending `/axb-clarify` Q1**. Next pipeline step after Q1–Q3 close: `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
