# Technical Research: Persist a failed turn's completed tool steps (round 080)

**Plan Package**: `specs/plans/080-failed-turn-partial-persistence`
**Created**: 2026-09-22
**Anchor**: issue [#161](https://github.com/gosharplite/tellme/issues/161)
**Grounding**: `dev` @ `b1c929e`

> Decision-driven research (no code). The behaviour is anchored by the issue; its five *Open Design Decisions* are resolved below (D1–D7). The design is tellme's own; the reference is not asserted.

---

## 0. The problem, precisely

Round 079 (ADR 0051) persists a partial turn **only** for an operator interruption (`errors.Is(err, context.Canceled) && len(result.Steps) > 0`). Every other loop error falls through to `emitToolError` (exit 7) / `emitProviderError` (exit 6) and **skips** `store.Append` — so a **failed** turn discards the prompt and every completed tool step, exactly as before round 079. The loop already returns the completed steps **alongside** the error (`agentport.Result{Steps, Calls}, err`), so the work is in memory and the loss is purely the caller's policy (`store.Append` is reached only when `err == nil`).

Two consequences:
- **The motivating case**: a long tool loop (7 completed steps) whose `N+1`-th inference fails (retry exhausted, or a non-retryable 4xx) loses all 7 steps.
- **Hole #2 (unrecorded until now)**: the round-078 retry decorator returns `lastErr` — the previous attempt's provider error — on its two abort paths (`ctx.Err() != nil` and `!sleep(ctx, delay)`), **not** a `context.Canceled`-wrapping error. So a `Ctrl+C` **during a retry wait** is not classified as an interruption, and round 079's branch does not fire → the partial turn is lost. Narrow (the signal must land in the 1–3 s window) but real.

## 1. Decisions

### D1 — Persist a failed turn with completed steps (**the deliverable**)

When `loop.Run` returns `err != nil` and `len(result.Steps) > 0`, and the turn was **not** an operator interruption, `runTurn` appends **one** atomic `history.Entry` (the round-079 shape) and then reports the **unchanged** failure surface (the frozen phrase + exit 6/7) plus one informational `stderr` line. Zero steps ⇒ nothing (today's clean abort). *Rejected:* persisting the work only when it "looks safe" — the loss is identical regardless of the failure class.

### D2 — Scope: **any** failed turn with completed steps (S-1 → B)

The predicate is a single `len(result.Steps) > 0` gate, covering **provider failures** (retry exhausted / non-retryable) **and** the tool-loop failure (`agentport.ErrIncomplete`: `MAX_TOOL_LOOP` reached / the unknown-tool cap exhausted / no tools registered → exit 7). *Rationale:* the loss is identical across classes; one predicate is simpler than two; the issue recommends (B). *Rejected:* (A) provider-only — it would leave the exit-7 class silently losing work for no benefit (the same one-line change covers it).

### D3 — The failure surface is **unchanged** (S-3): keep the phrase + exit code

A failed-but-kept turn still exits **6** (provider) / **7** (tool loop) with its frozen class phrase. *Rejected:* exit 0 — it would misreport a genuine failure (round 079's exit 0 is right only because the operator chose to stop); exit 130 — an eleventh code, breaking the pinned ten. No new class phrase. The persistence adds **one informational `stderr` line only**: `[HH:MM:SS] kept N completed tool step(s) in the session history` — chrome-styled, control-free, carrying **no** `tellme: ` prefix (the round-017 *exactly-one-`tellme:`-line* contract holds) and **not** routed to `turns.log` (the ADR 0050/0051 precedent). It is emitted **only when the append actually succeeded** (never claim a keep that did not happen).

### D4 — A **distinct** synthetic answer for the failure case (S-4)

Two new stored constants in `internal/domain/history`, chosen by the error class:
- `history.ProviderFailedTurnAnswer = "[Turn ended early: the provider request failed]"`
- `history.ToolFailedTurnAnswer = "[Turn ended early: the tool loop did not complete]"`

*Rationale:* the `answer` is the only marker in the stored history; a reader (`-l`) must not mistake an operator stop (round 079's `[Turn interrupted by operator via Ctrl+C]`) for a failure. The two strings name the class without duplicating the frozen **phrases** (`the provider request failed` / `the tool request failed` are `tellme:`-class surfaces, not history values).

### D5 — The interruption predicate is the round-079 seam, broadened (S-2): fold hole #2

The operator-interruption branch becomes:

```go
if len(result.Steps) > 0 && (ctx.Err() != nil || errors.Is(err, context.Canceled)) { … interrupted path … }
```

`ctx.Err() != nil` is the **structural** signal of a `SIGINT`/`SIGTERM` (runTurn's `defer cancel()` runs only *after* this branch). It is strictly broader than the round-079 test and **catches hole #2**: a cancellation that surfaced as the retry decorator's `lastErr` (a transport `*llm.ProviderError`) now routes to the interrupted path (round 079's exit 0 + interruption line) — which is what the operator intends. The union keeps the typed `errors.Is` term (an adapter that wraps `context.Canceled` without the parent ctx being cancelled). Both terms are **structural** — never a string match (NFR-004). *Rejected:* `ctx.Err() != nil` alone — it would drop the error-wrapping case the round-079 unit tests exercise (and any adapter that only surfaces the cancellation via the error).

### D6 — Best-effort persistence on the failure path; **do not mask the failure**

On the failed-then-kept path, `store.Append` is **best-effort**: if it fails, the reported failure surface is **unchanged** (phrase + exit 6/7) and the informational line is **not** emitted (we do not claim a keep that failed). *Rationale:* D3's invariant "the failure surface is unchanged" must hold — the persistence is additional value, not a precondition, and exit 4 (the `emitHistoryError` route) would mask the real failure. The divergence from the operator-interruption path (round 079 routes an append failure to `emitHistoryError`/exit 4) is deliberate and recorded: there exit 0 would be a *lie* if the work was not kept, so the environment phrase is the honest surface there. *Rejected:* exit 4 on the failure path (masks 6/7). **Recorded as a forward item:** a silent best-effort append failure on this path (no operator signal that the keep failed).

### D7 — Records (S-7): ADR 0052 + truth + domain model + CLI feature

- **ADR 0052** (`docs/decisions/0052-failed-turn-partial-persistence.md` + index row).
- `specs/truth/techstack.md`: **MODIFY** the round-079 *Interrupted-turn persistence (operator)* row → a broader **Partial-turn persistence (operator interruption or failure)** row (naming both synthetic answers, both paths, the unchanged failure surface, the informational line, and the best-effort failure-path append), and **MODIFY** the *Session history store* row's append-after-complete qualification to include the failed-turn case.
- `docs/domain-model/tellme.modelith.{yaml,md}`: **MODIFY** `history-append-after-complete` and `turn-answer-stored-verbatim` to extend the round-079 exception from *operator-interrupted* to *operator-interrupted or failed* (same-PR per ADR 0041; re-rendered; `modelith-check` green). A recorded divergence from the reference (which discards the partial turn in both cases).
- `specs/truth/features/cli/chat/remembering-the-conversation.feature`: new **Rules** + Examples (a failed turn keeps its work; the session continues from it; a failure before any step writes nothing); `chat/dsl.md`: the new Given/Then rows.

## 2. Rejected alternatives (recorded — not forward items)

| Alternative | Why rejected |
| --- | --- |
| Provider-only scope (A) | The exit-7 class loses work identically; one predicate covers both (D2). |
| Exit 0 / 130 on the failed-but-kept path | Misreports the failure; breaks the pinned ten-code set (D3). |
| Reuse the round-079 interrupted text | A reader would think the operator stopped it (D4). |
| `emitHistoryError`/exit 4 on a failed-turn append failure | Masks the real failure (D6). |
| Persisting without a synthetic close | Breaks role alternation (round 079's reason; I-1). |

## 3. Residual risks (→ §Forward of ADR 0052)

- **RF-080-1** the failed-turn persistence is a small steady disk cost (one line per failed turn with ≥ 1 step).
- **RF-080-2** a silent best-effort append failure on the failure path (D6) — no operator signal that the keep failed.
- **RF-080-3** no usage is persisted for the failed-but-kept turn (mirrors round 079 D6).
- **RF-080-4** the two new synthetic answers are stored as the turn's `answer`; a downstream reader has no marker beyond the text.
- **RF-080-5** the exit-7 (tool-loop) case is kept + reported (exit 7) but its Gherkin carrier is the provider-failure one plus the unit pins (a deliberate narrowing if the E2E cannot drive a bound-reached turn with a kept step cheaply — see `plan.md`).
- **RF-080-6** hole #2's full behavioural E2E (a `Ctrl+C` landing inside a real 1–3 s retry wait) is unit-witnessed at the predicate; a timing-sensitive E2E is declined.
- **RF-080-7** no `history.archive.jsonl` interaction beyond the existing semantics.
