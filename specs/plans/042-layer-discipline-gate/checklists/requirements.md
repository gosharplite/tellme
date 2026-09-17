# Specification Quality Checklist: layer-discipline gate + violation baseline (round 042)

**Created**: 2026-09-17

**Feature Directory**: `specs/plans/042-layer-discipline-gate`

**Spec Path**: `specs/plans/042-layer-discipline-gate/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear
- [x] No implementation/framework detail written as a *requirement* (guard form is an explicit RD decision, not an FR)
- [x] Edge cases cover the main high-risk situations (baseline-before-gate, host-independence, new vs baselined vs stale entries, build-tag hiding, `tests/` scope)
- [x] Key entities and success criteria are present (or their inapplicability stated)

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (detect the new violation) → US2 (baseline + ratchet + fail-on-stale) → US3 (recorded truth)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-012, FR-013, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 rule model · Q2 package scope · Q3 stale-entry policy)
- [x] This round's clarify questions capped at 1–3 — **two were asked; Q2 was resolved by measurement** (the gate's scope does not change the baseline)
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (guard form A4; ADR A5; `/axb-dsl-refine` NOOP-or-not A6)
- [x] No remaining `NEEDS CLARIFICATION` — round-1 answers **locked**:
  - **Q1 → Option 2** — a **broad** import-direction rule (baseline = **8**, incl. `internal/agent → internal/ui`); the pinned layer ranking is in `spec.md`.
  - **Q2 → measured** — package scope pinned (production `internal/**`; `cmd/**` + `tests/**` exempt); baseline unchanged by scope.
  - **Q3 → Option 1** — a **stale** baseline entry **fails** the gate (FR-008).

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev`; red on a new violation; red on a stale entry)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-005)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- Baseline count restated as **8** (broad rule): the 7 `cli → infrastructure` edges + `internal/agent → internal/ui`.
- **PR #94 review fold (B-1):** the rule is now a **two-part predicate** (A direction + B application target rule + C domain purity + D default-deny) — the earlier one-part "no upward import" statement could not reproduce the 7 `cli → infrastructure` entries; the baseline is now **derivable from the rule**. Worked examples in **ADR 0011**.
- Ratchet refinement recorded: the `agent → ui` entry is R3/R4's to remove, so the count reaches **0 across R2–R4**, not R2 alone (how [#92](https://github.com/gosharplite/tellme/issues/92) AC1 is to be read).
- The baseline MUST be re-measured from the gate's own output at freeze (no hand-transcription); it is **generated**, not transcribed.
- **Other folds:** B-2 (anchor the enumeration to the module root + self-test the graph), TD-1 (`CROSS_TARGETS` union), TD-2 (build-tag claim withdrawn), TD-3 (default-deny), TD-4 (AC4 cycles asserted; AC2's 7→8 revision recorded on [#93](https://github.com/gosharplite/tellme/issues/93)), RF-1 (**ADR 0011** required), RF-2 (deterministic baseline format), RF-3 (one normative ranking source).

## Ready determination

- [x] Ready to proceed to downstream planning
- [ ] A high-impact requirement gap must be closed first — **none remaining** (round-1 answers locked)

**Note**: plan half only — stop before `/axb-tasks` and `/axb-implement`. Everything up to and including `/axb-dsl-refine` is in scope for this branch. Next pipeline step is `/axb-spec-by-example` (or `/axb-technical-research`; operator's call).
