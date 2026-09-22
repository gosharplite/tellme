# Tasks — Preserve completed tool steps on an interrupted turn (round 079)

**Plan Package**: `specs/plans/079-interrupted-turn-partial-persistence`
**Inputs**: `spec.md`, `plan.md`, `research.md`, `features/acceptance/…`

> Executed One-Shot (red → green → refactor). Each task is marked `[X]` after its verification passes. The Test-Alignment layer (Phase 3) carries the interface carrier (the truth Rule/Examples); the Feature phases carry the green/refactor work.

## Phase 1 — Setup & Foundational

- [X] **T001** Add the synthetic-answer constant `history.InterruptedTurnAnswer` (`internal/domain/history/history.go`) — a stored record value, fixed/control-free/single-line, with its rationale (the role-alternation close).

## Phase 2 — Foundational (the persistence seam)

- [X] **T002** Add the interrupted-turn branch + helper in `internal/cli/cli.go`: on `errors.Is(err, context.Canceled) && len(result.Steps) > 0`, `persistInterruptedTurn(env, store, prompt, result)` appends ONE entry closed with the synthetic answer, emits the informational `stderr` line (chrome-styled, no `tellme: ` prefix), and returns `Success`; a failed append degrades via `emitHistoryError`.
- [X] **T003** Extend the in-package loop double (`internal/cli/testdeps_test.go`): `fakeLoop.errResult` is returned ALONGSIDE `runErr`, so a partial-result failure (the interruption shape) is exercisable.

## Phase 3 — Test Alignment & Implementation

- [X] **T004** (Test Alignment) Extend the truth contract: two new Rules + 2 Examples in `specs/truth/features/cli/chat/remembering-the-conversation.feature` (a turn interrupted after a tool step is persisted + closed with the synthetic answer + exits 0; a session with a kept interrupted turn resumes, the replayed conversation closing the earlier turn with an assistant message); the new `chat/dsl.md` Given/Then rows + the round-079 note.
- [X] **T005** (Test Alignment) Extend the hermetic E2E: the `a configured provider "…" whose endpoint runs the command "…" and then answers with "…"` Given (a scripted `execute_command` running `kill -INT $PPID; sleep 30` — a REAL SIGINT to the turn process) + the `the session history already holds an interrupted exchange with 1 tool step` Given + the round-079 Thens (`step_r079_interrupt.go`).
- [X] **T006** (Green) **US1** — an operator interruption with ≥1 completed step persists the partial turn (exactly one entry, the completed steps, the synthetic answer, `calls` = the completed rounds) and exits **0** with `stdout` empty and the informational `stderr` line.
- [X] **T007** (Green) **US1** — a session holding the kept interrupted turn resumes: the replayed conversation ends the earlier turn with an assistant message (role alternation), so the next prompt is accepted.
- [X] **T008** (Green) **US2** (unit-carried, deliberate narrowing) — a turn interrupted with **zero** completed steps writes NOTHING (today's clean abort: frozen phrase + exit 6).
- [X] **T009** (Refactor under green) Keep `persistInterruptedTurn` a small function; no new port; the constant single-owned in `internal/domain/history`.
- [X] **T010** (Regression) `make verify` + `go test -count=1 ./...` green; the non-cancellation failure path is unchanged (a transport failure with steps is still a failure); `stdout` byte-exact.

## Unit pins

- [X] **U1** `internal/cli/interrupted_turn_test.go` — the synthetic-answer LITERAL; the interrupted-with-steps persist (typed detection through the `ProviderError` wrap → Success, one entry, synthetic answer, `calls`/steps); the zero-step no-write (ProviderError); the append-failure degradation (EnvironmentError); the non-cancellation guard (a transport error with steps stays a failure).

## Deliberate narrowings (recorded)

- **N-1 — the zero-step negative is unit-carried, not E2E-carried.** Landing a signal *before the first tool step* deterministically requires a pty/stall harness the round declines; the E2E produces the interruption **inside a tool call** (`kill -INT $PPID`) so it necessarily has ≥1 step. The zero-step invariant (I-2/FR-005) is carried by the unit pin `TestRunTurn_InterruptedWithoutStepsWritesNothing` (the round-036 unit-only-narrowing precedent). Recorded in `plan.md` §6 and ADR 0051 §Forward (RF-079-6).
- **N-2 — the informational line is not E2E-byte-pinned, only sub-string-asserted** (`interrupted by operator`); the exact wording is contract-free (like the class-phrase trailing detail).

## Fold ledger

- **F-079-1** (self-found at the E2E) — the reuse of `the request carried the earlier exchange "{p}" and "{a}"` failed for a **tool-using** prior turn (it asserts an adjacent user→assistant pair, but a tool round intervenes). Folded: a dedicated Then `the request closes the earlier turn with the assistant answer "{answer}"` asserts an assistant message equal to `{answer}` followed later by a `user` message — the role-alternation witness. The `the request replayed the earlier tool step "execute_command"` Then (unchanged) carries the step replay.
