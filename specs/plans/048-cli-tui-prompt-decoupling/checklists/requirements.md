# Specification Quality Checklist: de-couple `internal/cli` from the TUI prompt — R5.2, first de-coupling slice (round 048)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/048-cli-tui-prompt-decoupling`

**Spec Path**: `specs/plans/048-cli-tui-prompt-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (remove `internal/cli → internal/ui/tui/prompt` via an injected port; RULE-E baseline 3 → 2)
- [x] No implementation/framework detail written as a *requirement* (the port shape/name/home are explicit RD decisions behind FR-002 / RULE-A·C)
- [x] Edge cases cover the main high-risk situations (domain-purity leak → port home; RULE-A adapter placement; test-only imports; half-refactor; re-introduced edge; stale entry; missing injection; cycles; build-tag scope)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (invert the TUI seam → no import) → US2 (baseline 3 → 2, gate-proven) → US3 (recorded in truth + ADR)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-009…FR-011, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] Only high-impact gaps escalated to `/axb-clarify` — **Q1 (slice selection)** was asked first; **Q2 (port home)** and **Q3 (F-4 ride-along)** are design-/scope-shaping
- [x] Questions asked **one at a time** (Q1 answered; Q2/Q3 pending); capped at 1–3 per round
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (port shape A3; ADR number A4; NOOP set A5/A6)
- [ ] **NEEDS CLARIFICATION (blocking)**: **Q2 — port home** (`internal/domain/**` vs `internal/app/**`) shapes the design + the ADR (a domain port must stay RULE-C-pure; the adapter stays inside `internal/ui/**` under RULE-A). **Q3 — F-4 ride-along** is a scope question (non-blocking, but bundled with Q2 for a single follow-up).
  - **Q1 → B LOCKED**: slice = the **TUI prompt** edge; slug/branch `048-cli-tui-prompt-decoupling`.

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev` with baseline 2; red on a re-introduced edge; red on a stale entry; loud failure on a missing injection)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 3 today** (re-measured 2026-09-18 @ `dev` `e5db873`): `internal/cli → internal/agent`, `internal/cli → internal/ui`, `internal/cli → internal/ui/tui/prompt`. Q1-B removes the **TUI prompt** edge → baseline **2**; the other two stay baselined for later slices.
- **R5.2, not R5**: round 047 delivered **R5.1** (the RULE-E gate + baseline); this is the first de-coupling slice; the remaining edges + F-4/F-6/F-7/F-8 stay on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).
- **RULE-A shapes the port** (not just RULE-C): only tiers ≥ 5 may import `internal/ui/**`, so the adapter that calls `tuiprompt.Run` stays inside `internal/ui/**`; only the port declaration's home (Q2) is in question.
- **Witness discipline**: the witness is the **gate + unit seams** (NFR-004), never the E2E suite (#92 AC5); falsifiability witnesses (a)/(b)/(c) reproduced then reverted (FR-010).
- **The existing `tuiPromptRunner` seam** (`cli.go:173`/`180`) is the natural locus of the inversion; the nil-fallback that live-imports the TUI package must go, and the dispatch/submit tests are adapted to the port.
- **Atomicity** — de-coupling + baseline regeneration + truth/ADR land as one PR (NFR-002, round-040 TD-1).
- **No new Makefile target** — the gate rides the existing `verify-architecture` member of `verify`.

## Ready determination

- [ ] Ready to proceed to downstream planning — **blocked on Q2**
- [x] A high-impact requirement gap must be closed first — **Q2 (port home) blocks** (it fixes the design + ADR; Q3 rides along)

**Note**: plan package updated with the Q1 fold. On the Q2/Q3 answers the spec will be finalised (port home + FR-002 refinement + F-4 riding along or not) and the pipeline continues to `/axb-technical-research` (with `/axb-spec-by-example` **NOOP** — no user-facing journey).
