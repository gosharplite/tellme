# ADR 0052 — Persist a failed turn's completed tool steps

**Status**: Accepted (round 080)

**Date**: 2026-09-22

**Related**: anchor issue [#161](https://github.com/gosharplite/tellme/issues/161) · round 079 / [ADR 0051](0051-interrupted-turn-partial-persistence.md) (the operator-interruption persistence shape this extends — one `history.Entry`, a synthetic closing answer, `history.Entry`/`history.Step` schema unchanged, the informational `stderr` line) · round 007 (append-after-complete) · round 008 (the loop's tool steps) · round 078 / [ADR 0050](0050-bounded-provider-transport-retry.md) (the retry decorator whose `lastErr`-on-abort is hole #2) · round 030 (the output-cap truncation guard — a non-retryable failure) · `specs/truth/techstack.md` §Session management · `specs/truth/features/cli/chat/remembering-the-conversation.feature` · `docs/domain-model/tellme.modelith.yaml` (`history-append-after-complete`, `turn-answer-stored-verbatim`).

## Context

Round 079 (ADR 0051) persists a partial turn for an **operator interruption only** (`errors.Is(err, context.Canceled) && len(result.Steps) > 0`). Every other loop error falls through to `emitToolError` (exit 7) / `emitProviderError` (exit 6) and **skips** `store.Append` (`internal/cli/cli.go` `runTurn`) — so a **failed** turn discards the prompt and every completed tool step, exactly as before round 079. The loop already returns the completed steps **alongside** the error (`internal/agent/agentloop.go` returns `agentport.Result{Steps, Calls}, err`), so the work is in memory and the loss is purely the caller's persistence policy.

Two observed consequences drove this decision:
1. **The motivating case.** A long tool loop (7 completed steps) whose next inference fails — the round-078 retry exhausts, or a non-retryable 4xx/auth/decode/truncation — loses all 7 steps, forcing a re-run.
2. **Hole #2 (unrecorded).** The round-078 retry decorator (`internal/cli/retry_gateway.go`) returns `lastErr` — the *previous attempt's* provider error (a transport `*llm.ProviderError`), **not** a `context.Canceled`-wrapping error — on both of its abort paths (`ctx.Err() != nil`; `!sleep(ctx, delay)`). So a `Ctrl+C` landing **during a retry wait** is not classified as an interruption, and round 079's branch does not fire — the partial turn is lost, reported as `the provider request failed` + exit 6.

## Decision

1. **Persist a failed turn with completed steps (D1).** When `loop.Run` returns `err != nil`, `len(result.Steps) > 0`, and the turn was **not** an operator interruption, `runTurn` appends **one** atomic `history.Entry` (the round-079 shape) — closing the partial turn with a synthetic answer — and then reports the failure. A failed turn with **zero** steps writes nothing (today's clean abort).

2. **Scope: any failed turn with completed steps (D2).** One `len(result.Steps) > 0` gate covers both the **provider** failure (retry exhausted / non-retryable) and the **tool-loop** failure (`agentport.ErrIncomplete`: `MAX_TOOL_LOOP` reached / the unknown-tool cap exhausted / no tools registered). *Rejected:* provider-only — the exit-7 class loses work identically; one predicate is simpler. (Operator-vetoable; adopted as the issue recommended.)

3. **The failure surface is unchanged (D3).** A failed-but-kept turn still exits **6** (provider) / **7** (tool loop) with its frozen class phrase. No new phrase; the exit-code set stays **ten**. The persistence adds **one informational `stderr` line only** — `[HH:MM:SS] kept N completed tool step(s) in the session history` — chrome-styled, control-free, carrying **no** `tellme: ` prefix (the round-017 *exactly-one-`tellme:`-line* contract holds) and **not** routed to `turns.log` (the ADR 0050/0051 precedent). It is emitted **only when the append succeeded**. *Rejected:* exit 0 (misreports a genuine failure); exit 130 (an eleventh code).

4. **A distinct synthetic answer for the failure case (D4).** Two stored constants in `internal/domain/history`, chosen by the error class: `history.ProviderFailedTurnAnswer = "[Turn ended early: the provider request failed]"` and `history.ToolFailedTurnAnswer = "[Turn ended early: the tool loop did not complete]"`. *Rejected:* reusing round 079's `[Turn interrupted by operator via Ctrl+C]` (a reader would think the operator stopped it). The strings name the class without duplicating the frozen **phrases** (which are `tellme:`-class `stderr` surfaces, not stored history values).

5. **The interruption predicate is broadened (D5) — folding hole #2.** The operator-interruption branch becomes `len(result.Steps) > 0 && (ctx.Err() != nil || errors.Is(err, context.Canceled))`. `ctx.Err() != nil` is the **structural** signal of a `SIGINT`/`SIGTERM` (`runTurn`'s `defer cancel()` runs only *after* this branch); it catches a cancellation that surfaced as the retry decorator's `lastErr` (hole #2), routing it to the interrupted path (round 079's exit 0 + interruption line) — the operator's intent. The union retains the typed `errors.Is` term (an adapter that wraps `context.Canceled` without the parent ctx being cancelled). Both terms are structural — never a string match. **Precedence (fold F-080-3):** `interrupted` is evaluated **before** the failure classification, so a `Ctrl+C` that lands *after* a genuine failure has been detected converts that turn into the **interruption** outcome (exit 0 + `history.InterruptedTurnAnswer`) rather than the failure surface (exit 6/7 + the failure answer). This window is created by the widening (round 079 did not have it — a transport/400 error does not wrap `context.Canceled`). The choice is deliberate: the operator's stop is the last actor and the completed work is kept either way.

6. **Best-effort persistence on the failure path — never mask the failure (D6).** If `store.Append` fails on the failed-then-kept path, the reported failure surface is **unchanged** (phrase + exit 6/7) and the informational line is **not** emitted. The divergence from the operator-interruption path (round 079 routes an append failure to `emitHistoryError`/exit 4) is deliberate: there exit 0 would be a lie if the work was not kept; here the failure exit is already honest, so the persistence is additional value, not a precondition. *Rejected:* exit 4 (masks 6/7).

7. **Records (D7).** ADR 0052 + index; `techstack.md` — the round-079 *Interrupted-turn persistence (operator)* row → a broader **Partial-turn persistence (operator interruption or failure)** row, and the *Session history store* append-after-complete qualification extended to the failed-turn case; the CLI truth feature (`remembering-the-conversation.feature` Rules/Examples) + `chat/dsl.md` rows; and a **domain-model** amendment of `history-append-after-complete` + `turn-answer-stored-verbatim` (extend the exception from *operator-interrupted* to *operator-interrupted or failed*), re-rendered. `history.Entry` / `history.Step` JSON shapes are **unchanged**.

8. **Scope (S-8).** The prompt/tool-loop turn only; MCP steps are ordinary steps on the same seam; `-b`/`--retry` and the offline readers are untouched; the in-flight child `SIGKILL` is untouched. Hermetic: no pty, no live network — the failure is produced by the in-process fake provider (an always-drop transport → retry exhaustion; an outright rejection → non-retryable), and hole #2's predicate is unit-witnessed.

## Consequences

- A failed tool turn keeps every completed step (the 7-of-10 case survives), so the operator can resume with *"continue"* / *"summarize what you found so far"* instead of starting over.
- The failure surface is **unchanged**: the class phrase (`the provider request failed` + exit 6, or `the tool request failed` + exit 7) and `stdout` (empty) are byte-identical to today; `stderr` gains exactly one non-class informational line.
- The persisted partial turn replays as a valid `user … assistant` sequence on **both** families (the synthetic close).
- Hole #2 is folded: a `Ctrl+C` during a retry wait now takes the interrupted path (exit 0 + the interruption line) and keeps the work.
- One extra `history.jsonl` line per failed turn with ≥ 1 completed step; the schema is unchanged.
- **Domain-model amendment (recorded divergence from the reference):** `history-append-after-complete` and `turn-answer-stored-verbatim` gain the failed-turn exception (the reference discards the partial turn in both the interruption and the failure case).
- The failed-turn `Append` is best-effort (a silent loss on append failure — a recorded forward item); no usage is persisted for the partial turn.

## Forward

- **RF-080-1** — the failed-turn persistence is a small steady disk cost (one line per failed turn with ≥ 1 step).
- **RF-080-2** — a silent best-effort append failure on the failure path (D6): no operator signal that the keep failed.
- **RF-080-3** — no usage is persisted for the failed-but-kept turn (mirrors round 079 D6).
- **RF-080-4** — the synthetic answers are stored as the turn's `answer`; a downstream reader has no marker beyond the text.
- **RF-080-5** — the exit-7 (tool-loop) case is kept + reported (exit 7), witnessed by unit pins; its Gherkin carrier is narrower than the provider-failure one.
- **RF-080-6** — hole #2's full behavioural E2E (a `Ctrl+C` inside a real 1–3 s retry wait) is **declined** (timing-sensitive); the **predicate is unit-witnessed** by a direct `failTurn`-seam pin (a cancelled ctx + a transport error → the interrupted path), falsifiable by deleting the `ctx.Err()` term (fold F-080-1).
- **RF-080-A** — the union's `errors.Is(err, context.Canceled)` term is now **unreachable in production** (`runTurn` owns the ctx; any child cancellation surfaces as `ctx.Err() != nil`) and is load-bearing only for the round-079 unit pin, so that pin no longer reflects the shipping mechanism. A future seam (injecting the turn ctx into `runTurn`) would restore unit-testability of the shipping mechanism.
- **RF-080-B** — the failed-turn store is now also the storage path for the exit-7 tool-loop failures (`MAX_TOOL_LOOP` reached, the round-076 unknown-tool cap), so `-l` after such a failure shows a synthetic answer; no scenario asserts what `-l` shows for a failed turn (the round-079/080 reader-marker half).
- **Note (N-080-4)** — the author's self-found `F-080-1` (tasks.md §Fold ledger) and the reviewer's `F-080-1` share a number (a pre-existing author/reviewer collision, cf. round 079).
- **RF-080-7** — no `history.archive.jsonl` interaction beyond the existing semantics.
