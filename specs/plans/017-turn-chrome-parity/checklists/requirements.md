# Spec Quality Checklist: tellme turn surface — operator chrome parity (round 017)

**Created**: 2026-09-14

**Feature Directory**: `specs/plans/017-turn-chrome-parity`

**Spec Path**: `specs/plans/017-turn-chrome-parity/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear (non-TUI prompt-turn chrome parity: input-capture line + rule/`Turn N` frame + blank-line spacing; the `-i` TUI and post-turn lines are out of scope)
- [x] No implementation technology, framework, or code detail is written as a requirement (the `─` rule width, the `╭─⠿` glyph, the colour codes/`TTY` gate, and the exact stream plumbing are named as the parity target/assumptions, not as FR text; behaviour is stated observably)
- [x] Edge cases cover the main high-risk scenarios (no prompt, `-r`, non-terminal stream, resumed session, resolve failure, `-i`, narrow width)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (capture line P1 · turn frame P1 → scope boundary P2)
- [x] Each user story is independently verifiable (hermetic run + injected streams/clock + stream/order assertions; no real pty)
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries (no-new-class-phrase + byte-exact stdout; determinism)
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps were escalated to `/axb-clarify` (surface scope; startup-chrome set)
- [x] This round's clarify questions were held to the budget (2 questions, one at a time; both resolved)
- [x] Low-risk undecided details are disclosed via assumptions (A1–A7) rather than asked
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking (none remain)

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths (capture line present/once; frame order + turn number; scope boundary/absence)
- [x] Success criteria are measurable, verifiable, and technology-neutral (line presence/absence + ordering; `Turn N` = persisted turns + 1; byte-exact stdout; prior rounds green)
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Operator decisions (2026-09-14, via `/axb-clarify` round 1)

- **Surface scope → (A) + (B)** — the chrome applies to the positional/piped prompt turn and the round-012 interactive plain reader; the `-i` TUI surface (round 016) is **unchanged** (`FR-007`, `A2`).
- **Startup set → input-capture line only** — add `[HH:MM:SS] Input captured. Processing...`; the reference's `[Info] Starting chat...` line is **out of scope** this round (`A3`).

### Folded assumptions (accepted, not asked — low risk)

- **A1** — the reference is `tell-me-go`'s `internal/ui/capture.go` + `internal/ui/renderer_metrics.go`.
- **A4** — `<N>` = persisted turn count + 1 (reference `SessionTurns + 1`); `--new` → `Turn 1`.
- **A5** — the exact chrome tokens are a `/axb-technical-research` + `/axb-dsl-refine` determination; the round-009 payload-line text is unchanged.
- **A6** — POSIX-only; no new dependency; hermetic verification.
- **A7** — post-turn lines are explicitly out of scope this round.

### Still open (non-blocking — later-phase determinations)

- **Exact chrome tokens** — the `─` rule width, the `╭─⠿` glyph, the header/blank-line layout, the colour codes and the TTY gate: a `/axb-technical-research` determination, then the executable rows by `/axb-dsl-refine`.
- **Where the acknowledgement and frame are emitted** (call-site vs turn entry; the `-i`/reader/A distinction) — an `/axb-technical-research` / `/axb-tasks` implementation determination.
- **DSL step vocabulary + interface feature** — the rows carrying "the turn acknowledges the captured input", "the turn is framed with the rule and header", and the "no frame on `-i`/non-prompt paths" boundary: a `/axb-dsl-refine` determination.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Three stories stand (capture line P1 · turn frame P1 · scope boundary P2). Both high-impact decisions were locked by the operator in `/axb-clarify` round 1; only research / DSL-layer determinations remain. `/axb-spec-by-example` (acceptance Gherkin) and `/axb-technical-research` can proceed.
