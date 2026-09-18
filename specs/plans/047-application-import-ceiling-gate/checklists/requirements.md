# Specification Quality Checklist: application import-ceiling gate (RULE-E) + application-tier import baseline (round 047)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/047-application-import-ceiling-gate`

**Spec Path**: `specs/plans/047-application-import-ceiling-gate/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (table shape / guard layout are explicit RD decisions, not FRs)
- [x] Edge cases cover the main high-risk situations (rule-before-baseline, new vs baselined vs stale vs unused-sanctioned entries, default-deny, cycles, test-file scope, third-party boundary, build-tag scope)
- [x] Key entities and success criteria are present (or their inapplicability stated)

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (detect the unsanctioned import) → US2 (baseline + fail-on-stale ratchet) → US3 (recorded + single-sourced rule)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-011…FR-013, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 increment model · Q2 rule shape + normative home · Q3 enforcement target · Q4 sanctioned set · Q5 round identity)
- [x] This round's clarify questions capped at 1–3 **per round**; the operator answered **one at a time** across **two** short rounds (Q1 → then Q2/Q3/Q4/Q5) — every answer changed a scope/user-story boundary, so all five were high-impact (no low-risk detail was escalated)
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (rule form A3; ADR A4; NOOP set A5/A6)
- [x] No remaining `NEEDS CLARIFICATION` — all five answers **locked**:
  - **Q1 → A** — gate-first slice (RULE-E + ADR + ratchet baseline; zero product code).
  - **Q2 → A** — a normative application-tier allow-list in the gate's tier table; default-deny; **fail-on-stale allow-list**.
  - **Q3 → A** — RULE-E binds **both** application tiers (`internal/app/**` + `internal/cli`).
  - **Q4 → A** — sanctioned set = `domain/**` + stdlib + `internal/config` + `internal/home` + `internal/app/**`; residuals = `cli → agent`, `cli → ui`, `cli → ui/tui/prompt`.
  - **Q5 → A** — ADR `0016` · slug `047-application-import-ceiling-gate` · F-4/F-6/F-7/F-8 deferred · #101 stays open.

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev`; red on a new unsanctioned import; red on a stale entry; red on an unused sanctioned entry)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 3** (re-measured 2026-09-18 @ `dev` `b42f868`): `internal/cli -> internal/agent`, `internal/cli -> internal/ui`, `internal/cli -> internal/ui/tui/prompt`. `internal/app/**` has **0** residual (it imports only `domain` + the sanctioned `config`). RULE-A/B/C stay **0**; the total committed baseline becomes **3** (all RULE-E) and ratchets to **0** across the later R5 slices.
- **R5 is not one round**: #101 states the de-coupling "is a multi-round programme, not a slice"; therefore 047 is **R5.1** (the gate + its enablers) and the residuals + F-4/F-6/F-7/F-8 are **out of scope** (recorded on the live issue, FR-012).
- **Witness discipline**: the witness is the **gate + unit seams** (NFR-004), never the E2E suite (round-009 trap / #92 AC5); falsifiability witnesses (a)/(b)/(c) reproduced then reverted (FR-013).
- **Rule is a CEILING, distinct from RULE-A/B/C/D** — an allow-list on downward imports; the ADR states **what the rule is *not*** (application tiers only; `internal/**` imports only, so stdlib/third-party are out of scope; no existing rule changes verdict).
- **Fail-on-stale allow-list** (Q2-A) is a deliberate symmetry with ADR 0011 **D3**: an unused sanctioned entry fails, forcing the allow-list to shrink to truth.
- **Atomicity** — rule + baseline + tier-table extension land as one delivery (NFR-002, round-040 TD-1).
- **No new Makefile target** — RULE-E rides the existing `verify-architecture` member of `verify`.

## Ready determination

- [x] Ready to proceed to downstream planning (`/axb-spec-by-example` NOOP → `/axb-technical-research`)
- [ ] A high-impact requirement gap must be closed first — **none remaining** (round-1 answers locked)

**Note**: plan package only — `/axb-specify` stops here. Next pipeline step is `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record RULE-E as a dev-surface gate (0 CLI interfaces; api/data/dsl-refine NOOP).
