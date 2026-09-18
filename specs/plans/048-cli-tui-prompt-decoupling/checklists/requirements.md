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

- [x] Only high-impact gaps escalated to `/axb-clarify` (Q1 slice selection · Q2 port home · Q3 F-4 ride-along)
- [x] Questions asked **one at a time** (Q1 → Q2 → Q3); capped at 1–3 per round; **all answered**
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (port package/name A3; ADR number A4; NOOP set A5/A6)
- [x] No remaining `NEEDS CLARIFICATION` — all three answers **locked**:
  - **Q1 → B** — slice = the **TUI prompt** edge (`internal/cli → internal/ui/tui/prompt`); slug/branch `048-cli-tui-prompt-decoupling`; baseline 3 → 2.
  - **Q2 → A** — the port lives in **`internal/domain/**`** (RULE-C-pure; adapter in a tier ≥ 5 / exempt `cmd/tellme`).
  - **Q3 → A** — **F-4 folded in** (delete `cli.Options.RunTUIPrompt` + the `tuiPromptRunner` func type).

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

- [x] Ready to proceed to downstream planning — clarify closed (Q1/Q2/Q3 locked)
- [ ] A high-impact requirement gap must be closed first — **none remaining**

**Note**: plan package final after the clarify fold. Next pipeline step is `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
