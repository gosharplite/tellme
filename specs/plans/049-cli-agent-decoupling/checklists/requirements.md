# Specification Quality Checklist: de-couple `internal/cli` from the turn loop — R5.3, the `cli → agent` slice (round 049)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/049-cli-agent-decoupling`

**Spec Path**: `specs/plans/049-cli-agent-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (remove `internal/cli → internal/agent` via an injected port; RULE-E baseline 2 → 1)
- [x] No implementation/framework detail written as a *requirement* (the port shape/name/home are explicit RD decisions behind FR-002 / RULE-A·C)
- [x] Edge cases cover the main high-risk situations (domain-purity leak → port home; RULE-A adapter placement; test-only imports; half-refactor; `ErrIncomplete`/`ToolDefs` byte-contracts; interface-seam `Validate()` blindness; re-introduced edge; stale entry; missing injection; cycles; build-tag scope)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (invert the loop seam → no import) → US2 (baseline 2 → 1, gate-proven) → US3 (recorded in truth + ADR)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-010…FR-012, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [ ] **Q1 (slice + sizing) PENDING** — must be resolved before this checklist can pass
- [ ] Q2 (port home/shape) and Q3 (ride-along) held until Q1 lands (asked one at a time)
- [x] High-impact gap identified and scoped to a single first question (Q1), with options A/B/C
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (port package/name A3; ADR number A4; NOOP set A5/A6)
- [ ] No remaining `NEEDS CLARIFICATION` — **blocked on Q1**

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev` with baseline 1; red on a re-introduced edge; red on a stale entry; loud failure on a missing injection)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 2 today** (re-measured 2026-09-18 @ `dev` `12964d6`): `internal/cli → internal/agent`, `internal/cli → internal/ui`. The round removes the **`→ agent`** edge → baseline **1**; the `→ ui` edge stays for a later slice.
- **Sizing correction (important)** — ADR 0017 §Forward classifies the **`→ agent`** edge as *"the deepest slice"* and the **`→ ui`** edge as *"the next natural slice"* (itself **not edge-sized**: ~20 call sites · 3 crossing value types · 4 stateful objects · 8 formatters). An earlier shorthand implied the `→ agent` edge was the *smaller* one — that was **wrong**; Q1 asks the operator to confirm the full-inversion vs. a re-cut.
- **RULE-A shapes the adapter** (not just RULE-C): only tiers ≥ 4 may import `internal/agent`, so the adapter that constructs the loop lives at a tier ≥ 4 (or the exempt `cmd/tellme`); only the port declaration's home (Q2) is in question.
- **R5.3** — round 047 delivered **R5.1** (the RULE-E gate + baseline); round 048 delivered **R5.2** (`→ ui/tui/prompt`); this is the next de-coupling slice; the remaining edge + F-6/F-7/F-8 stay on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).
- **Witness discipline**: the witness is the **gate + unit seams** (NFR-004), never the E2E suite (#92 AC5); falsifiability witnesses (a)/(b)/(c) reproduced then reverted (FR-011).
- **Atomicity** — de-coupling + baseline regeneration + truth/ADR land as one PR (NFR-002, round-040 TD-1).
- **No new Makefile target** — the gate rides the existing `verify-architecture` member of `verify`.

## Ready determination

- [ ] Ready to proceed to downstream planning — **blocked on Q1**
- [x] A high-impact requirement gap must be closed first — **Q1 (slice + sizing)**

**Note**: plan package initial; **do not** advance to `/axb-technical-research` until Q1 (and then Q2/Q3) are locked. `/axb-spec-by-example` is **NOOP** (no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
