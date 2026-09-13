# Spec Quality Checklist: tellme Interactive Multi-line Prompt Capture (round 012)

**Created**: 2026-09-13

**Feature Directory**: `specs/plans/012-interactive-multiline-prompt`

**Spec Path**: `specs/plans/012-interactive-multiline-prompt/spec.md`

## How to use

- Check each item against the current `spec.md`.
- If an item fails, record the concrete gap and the fix direction under "Issues & corrections".
- If any `NEEDS CLARIFICATION` remains, state explicitly whether it blocks the next planning step.

## Content completeness

- [x] All mandatory sections are present
- [x] The feature theme, scope, and main flow are clear
- [x] No implementation technology, framework, or code detail is written as a requirement (the pty/dependency choices are stated as observable constraints; the reader mechanism is left to research/DSL)
- [x] Edge cases cover the main high-risk scenarios (pipe, empty pipe, positional prompt, non-prompt paths, cap, immediate EOF, no-Windows)
- [x] Key entities and success criteria are present, or their absence is justified

## User stories & requirement attribution

- [x] User stories are ordered by business value and delivery order (capture P1 before cancel/empty P2, which depends on it)
- [x] Each user story is independently verifiable
- [x] Each user story includes acceptance scenarios
- [x] FR / NFR attributable to a single story are attached under that story
- [x] Global requirements keep only cross-story or non-attributable entries
- [x] No formal requirement is duplicated across story and global sections

## Gaps & clarification strategy

- [x] Only high-impact gaps are escalated to `/axb-clarify`
- [x] This round's clarify questions are held to the budget (Round 1 = 3 questions)
- [x] Low-risk undecided details are disclosed via `NEEDS CLARIFICATION` or assumptions
- [x] Remaining `NEEDS CLARIFICATION` items are marked as blocking or non-blocking

## Verifiability & success criteria

- [x] Acceptance scenarios are sufficient to verify the main success paths
- [x] Success criteria are measurable, verifiable, and technology-neutral
- [x] Assumptions state premises and boundaries only — no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

### Resolved (Clarify Round 1, 2026-09-13)

- **Q1 → Option 1 (invocation)** — bare `tellme` (no prompt, TTY stdin) starts the reader (`FR-001`, `FR-008`); the boot report remains for non-TTY/`--new`/subcommands (`FR-009`).
- **Q2 → Option 1 (hint stream)** — the hint is written to `stderr` (`FR-004`); `stdout` stays byte-exact.
- **Q3 → Option 1 (empty/cancel)** — empty or cancelled submission sends no request (`FR-006`, `FR-007`).

### Folded assumptions (accepted, not re-asked)

- **A1** — POSIX-only; `Ctrl+D` to send; no Windows variant (per instruction).
- **A2** — hint wording mirrors the reference's POSIX line.
- **A3** — send = EOF; line breaks preserved; trailing newline trimmed.
- **A4** — 1 MiB cap.
- **A5** — engages only with no positional prompt and no piped stdin.
- **A6** — no new dependency.
- **A7** — `Ctrl+C` cancels without sending.

### Still open (non-blocking — later-phase determinations)

- **Reader mechanism** — the exact insertion point in the dispatch and how the bounded read + cancel interleave (reuse round-005's `io.LimitReader` + the SIGINT context); a `/axb-technical-research` / `/axb-tasks` determination.
- **Hint colour/gating** — whether the hint is colourised when stderr is a terminal (and how it interacts with the byte-exactness gate); a `/axb-dsl-refine`/implementation determination.
- **Testability seam** — the interactive path is asserted through the injected `isTTY` seam + scripted stdin (unit layer); the E2E black-box suite remains pty-less, so the interactive branch is **not** E2E-observable (a named limitation, mirroring the round-005 pty pin).
- **Empty-EOF exit code** — the precise exit code for a cancelled/empty interactive run (success `0`, matching the reference's cancel) is to be pinned in the interface truth; a `/axb-dsl-refine` determination.

## Ready determination

- [x] Ready to proceed to later planning
- [ ] Must first resolve the high-impact requirement gaps

**Note**: Two stories stand (multi-line capture · cancel/empty). The interactive branch is POSIX-only and asserted through the injected terminal seam (the E2E suite stays pty-less). All clarify items are resolved; only research/DSL-level determinations remain. `/axb-spec-by-example` and `/axb-technical-research` can proceed in parallel.
