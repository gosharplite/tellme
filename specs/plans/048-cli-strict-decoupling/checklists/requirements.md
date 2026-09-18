# Specification Quality Checklist: `internal/cli` strict de-coupling — R5.2, the first de-coupling slice (round 048)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/048-cli-strict-decoupling`

**Spec Path**: `specs/plans/048-cli-strict-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (de-couple `internal/cli` from a selected unsanctioned internal package; shrink the RULE-E baseline)
- [x] No implementation/framework detail written as a *requirement* (the port shape/name/home are explicit RD decisions behind FR-002's purity constraint)
- [x] Edge cases cover the main high-risk situations (domain-purity leak → port home; test-only imports; half-refactor; re-introduced edge; stale entry; missing injection; cycles; build-tag scope)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (invert the selected edge → no import) → US2 (baseline shrinks, gate-proven) → US3 (recorded in truth + ADR)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-009…FR-011, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [ ] Only high-impact gaps escalated to `/axb-clarify` — **Q1 (slice selection) is OPEN; the operator has not yet answered**
- [ ] This round's clarify questions capped at 1–3 per round; asked **one at a time** (Q1 now; the port-home and any F-item ride-along follow)
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (port shape A3; ADR A4; NOOP set A5/A6)
- [ ] **NEEDS CLARIFICATION (blocking)**: **Q1 — which residual edge(s) does round 048 de-couple?** (see `spec.md` §Open decision). Blocking the package slug, the user-story split, and the acceptance surface.
  - Provisional candidates locked by nothing yet: A (all 3) · **B (tui prompt — recommended)** · C (agent) · D (ui) · E (ui + tui prompt).

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev`; red on a re-introduced edge; red on a stale entry; loud failure on a missing injection)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 3 today** (re-measured 2026-09-18 @ `dev` `e5db873`): `internal/cli -> internal/agent`, `internal/cli -> internal/ui`, `internal/cli -> internal/ui/tui/prompt`. The round removes exactly the Q1-selected edge(s); the baseline ratchets toward **0** across the remaining slices.
- **R5.2, not R5**: round 047 delivered **R5.1** (the RULE-E gate + baseline); this is the first de-coupling slice; the remaining edges + F-4/F-6/F-7/F-8 stay on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).
- **Witness discipline**: the witness is the **gate + unit seams** (NFR-004), never the E2E suite (#92 AC5); falsifiability witnesses (a)/(b)/(c) reproduced then reverted (FR-010).
- **Port purity** (FR-002) is the pivotal RD constraint: a `ui`/`agent` type may not enter `internal/domain/**` (RULE-C); if unavoidable, the port lives in `internal/app/**`.
- **Atomicity** — de-coupling + baseline regeneration + truth/ADR land as one PR (NFR-002, round-040 TD-1).
- **No new Makefile target** — the gate rides the existing `verify-architecture` member of `verify`.

## Ready determination

- [ ] Ready to proceed to downstream planning — **blocked on Q1**
- [x] A high-impact requirement gap must be closed first — **Q1 (slice selection) blocks**

**Note**: plan package only — `/axb-specify` stops here pending clarify Q1. On the answer the slug (if needed), stories, and FR slice-parameters will be folded, and the pipeline continues to `/axb-technical-research` (with `/axb-spec-by-example` **NOOP** — no user-facing journey).
