# Spec Quality Checklist: tellme `-i` Interactive Prompt — Strict Visual Parity (round 016)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/016-interactive-prompt-visual-parity`

**Spec Path**: `specs/plans/016-interactive-prompt-visual-parity/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (strict `-i` prompt parity: framed editor + styled suggestion list + debounce/insert + resize; the dashboard retired)
- [x] No implementation technology, framework, or code detail is written as a requirement (the Bubble Tea family, the exact chrome tokens, and the debounce value are named as the parity target/assumptions, not as FR text; behaviour is stated observably)
- [x] Edge cases cover the main high-risk scenarios (narrow width, monochrome, empty list, non-TTY `-i`, multi-word last-token accept, over-long suggestion, stale fetch)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (chrome parity P1 · interaction parity P1 → responsive width P2)
- [x] Each user story is independently verifiable (hermetic render + scripted keys + style/state assertions; no real pty)
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries (no-regression + retired dashboard + unchanged subsystems; determinism)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify` (none — the operator locked the single high-impact decision, "strict parity", in issue #39)
- [x] This round's clarify questions were held to the budget (0 questions; the decision was pre-made)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A6) rather than asked
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths (chrome present; dashboard absent; debounce/drop/insert; resize)
- [x] Success criteria are measurable, verifiable, and technology-neutral (chrome presence/absence; 100% insert; debounce/stale behaviour; prior rounds green)
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (operator decision, 2026-09-14 — issue #39)

- **Strict parity** — the `-i` prompt surface matches `tell-me-go`'s: a bordered multi-line editor over a styled suggestion list, with **no dashboard header** and **no status/keybinding line** (hints live in the placeholder). This **supersedes round-015 US3** for the `-i` surface (`FR-004`, `FR-010`, `A2`).

### Folded assumptions (accepted, not asked)

- **A1** — the reference is `tell-me-go`'s `internal/ui/tui/prompt/*` + `docs/user/tui-prompt.md`.
- **A3** — the shared log, suggestion sources, non-TTY fallback, and the round-012 plain reader are unchanged.
- **A4** — POSIX-only; no new dependency (Bubble Tea family already present).
- **A5** — the exact chrome tokens are a `/axb-ui-plan` (terminal) + `/axb-technical-research` determination.
- **A6** — hermetic verification (injected streams + scripted keys; no pty).

### Still open (non-blocking — later-phase determinations)

- **Exact chrome tokens** (border style/foreground, padding, height/width, placeholder text, list header text, selection styles) — a `/axb-ui-plan` (terminal) + `/axb-technical-research` determination.
- **Debounce value + async/cancel mechanism** — a `/axb-technical-research` determination.
- **DSL step vocabulary changes** — the rows for "the prompt shows the reference chrome", "Tab inserts the selection", and the **retired** dashboard row — a `/axb-dsl-refine` determination.
- **`ui/**` terminal-mode artifact** — the plan-side frames matching `tell-me-go` — a `/axb-ui-plan` (terminal mode) output; reviewed (not re-planned) by `/axb-system-analysis`.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Three stories stand (chrome parity P1 · interaction parity P1 · responsive width P2). The single high-impact decision (strict parity) is locked in issue [#39](https://github.com/gosharplite/tellme/issues/39); only ui-plan / research / DSL-layer determinations remain. `/axb-ui-plan` (terminal mode) and `/axb-technical-research` can proceed in parallel.
