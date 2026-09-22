# Tasks — Persist a failed turn's completed tool steps (round 080)

**Plan Package**: `specs/plans/080-failed-turn-partial-persistence`
**Inputs**: `spec.md`, `plan.md`, `research.md`, `features/acceptance/…`

> Executed One-Shot (red → green → refactor). Each task is marked `[X]` after its verification passes. The Test-Alignment layer (Phase 3) carries the interface carrier (the truth Rule/Examples); the Feature phases carry the green/refactor work.

## Phase 1 — Setup & Foundational

- [X] **T001** Add the two failure synthetic answers to the domain record: `history.ProviderFailedTurnAnswer` / `history.ToolFailedTurnAnswer` (`internal/domain/history/history.go`) — stored record values, schema unchanged, distinct from the round-079 interruption text.

## Phase 2 — Foundational (the persistence seam)

- [X] **T002** Extract `failTurn(env, store, prompt, err, result, ctx)` in `internal/cli/cli.go`: the operator-interruption branch broadened to `len(result.Steps) > 0 && (ctx.Err() != nil || errors.Is(err, context.Canceled))` (folds issue #161 hole #2); otherwise a failed turn with steps is persisted (best-effort) with a class-specific synthetic answer, then the unchanged failure surface + the informational keep line. Add `emitKeptStepsLine`.
- [X] **T003** Extend the in-process fake provider: a per-reply `Reply.Drop` (a transport drop on a specific scripted request) and `Reply.ErrorStatus` (a non-retryable rejection on a specific request), preserving the round-078 whole-fake precedence.

## Phase 3 — Test Alignment & Implementation

- [X] **T004** (Test Alignment) Extend the truth contract: new Rules + 4 Examples in `specs/truth/features/cli/chat/remembering-the-conversation.feature` (a failed turn keeps its steps — retry-exhausted and rejected-outright; the session continues from the kept turn; a zero-step failure writes nothing); the `chat/dsl.md` Given/Then rows (round 080) + the round-080 note.
- [X] **T005** (Test Alignment) Extend the hermetic E2E: the `… asks tellme to read "{path}" and then always drops / rejects …` Givens + the round-080 Thens (`step_r080_failed_turn.go`); `TELL_ME_FORCE_RETRY_DELAY_MS=0`.
- [X] **T006** (Green) **US1** — a provider failure (retry exhausted) after a completed step persists the partial turn (one entry, the provider-failure answer, `calls`/steps) and still exits 6 with the frozen phrase + the keep line; `stdout` empty.
- [X] **T007** (Green) **US1** — a non-retryable 400 equivalent persists the partial turn and still exits 6.
- [X] **T008** (Green) **US1** — a session holding the kept failed turn resumes: the replayed conversation ends the earlier turn with an assistant message (role alternation).
- [X] **T009** (Green) **US2** — a failure with zero completed steps writes nothing (today's clean abort).
- [X] **T010** (Green) The exit-7 tool-loop class (`ErrIncomplete` with steps) is persisted (unit-carried) with the tool-failure answer and still exits 7.
- [X] **T011** (Refactor under green) Keep `failTurn` small; the two answers single-owned in `internal/domain/history`; the fake change additive.
- [X] **T012** (Regression) `make verify` + `go test -count=1 ./...` green; the round-079 interrupted path still works (the union predicate); the best-effort append failure does not mask the failure; `stdout` byte-exact.

## Unit pins

- [X] **U1** `internal/cli/failed_turn_test.go` — the two failure-answer literals; the provider failure with steps (persist + phrase + exit 6 + keep line + empty stdout); the tool-loop failure with steps (persist + tool answer + phrase + exit 7); zero steps writes nothing; the append failure does not mask the failure (phrase + exit 6, no keep line).
- [X] **U2** `internal/cli/interrupted_turn_test.go` — `TestRunTurn_NonCancellationErrorWithStepsPersistsAsFailure` updated: a non-cancellation failure with steps is NOT an interruption (exit 6) but IS persisted with the **failure** answer (the round-079 test asserted the opposite; round 080 changed it).

## Deliberate narrowings (recorded)

- **N-1 — the exit-7 (tool-loop) carrier is unit-only.** Driving a bound-reached / cap-exhausted turn with a *kept step* hermetically is awkward (the loop would have to request a tool it cannot run); the exit-7 persistence is carried by the unit pin `TestRunTurn_ToolLoopFailureWithStepsIsPersisted`. Recorded in ADR 0052 §Forward (RF-080-5).
- **N-2 — hole #2's behavioural E2E is unit-witnessed at the predicate.** A `Ctrl+C` landing inside a real 1–3 s retry wait is timing-sensitive; the broadened predicate is covered by the union test + the round-079 interrupted tests. Recorded (RF-080-6).

## Fold ledger

- **F-080-1** (self-found at implementation) — the round-079 unit test `TestRunTurn_NonCancellationErrorWithStepsStillFails` asserted that a transport failure with steps is NOT persisted; round 080 changes exactly that. Folded: reworked to `TestRunTurn_NonCancellationErrorWithStepsPersistsAsFailure` (it still exits 6, but now persists with the failure answer — NOT the interruption one). A round-079 test was the *specification of the old behaviour*; the change is recorded, not silently reversed.
