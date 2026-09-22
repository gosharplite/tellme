# ADR 0051 — Preserve completed tool steps on an operator-interrupted turn

**Status**: Accepted (round 079)

**Date**: 2026-09-22

**Related**: anchor issue [#159](https://github.com/gosharplite/tellme/issues/159) · round 004 (the frozen provider phrase / exit 6) · round 007 (`history.jsonl` append-after-complete) · round 008 (the loop's tool steps) · round 011/`BuildMessages` (the replay projection) · round 012 (the interactive reader's operator-cancellation → `Success` convention) · round 027 (the `calls` counter) · round 034 (`persistTurnUsage`'s one-`AppendBatch`-per-turn gate) · round 056 (`refuseReasonless` — the reason-less refusal precedent for a recoverable, non-terminal path) · round 078 / [ADR 0050](0050-bounded-provider-transport-retry.md) (the `retryNotifier` inline-`stderr`-line precedent + the `DropFirst` fake mode) · `specs/truth/techstack.md` §Session management · `specs/truth/features/cli/chat/remembering-the-conversation.feature` · `docs/domain-model/tellme.modelith.yaml` (`history-append-after-complete`, `turn-answer-stored-verbatim`).

## Context

A prompt turn persists **only** when it reaches a final answer: `runTurn` (`internal/cli/cli.go`) calls `store.Append(...)` **after** `loop.Run` returns `err == nil`, and on any error it maps to a class phrase + exit code and **skips the append**. When the operator interrupts a long tool turn with `Ctrl+C` (`SIGINT`) — or `SIGTERM` — the turn context is cancelled (`signal.NotifyContext`), `Complete` returns a `*llm.ProviderError` wrapping `context.Canceled`, and the run exits with `tellme: the provider request failed: context canceled` (**exit 6**). The prompt and **every completed tool step** (held in memory in `agentport.Result{Steps, Calls}`) are **discarded** — so a turn that gathered valuable context across many steps (searches, multi-file analysis, tests/builds) is lost and must be redone.

A **partial** save is unsafe by itself: `history.jsonl` is replayed verbatim into the next request by `BuildMessages` as `user(prompt) → [assistant(tool_call) → tool(result)]… → assistant(answer)`. A turn persisted **ending on a `tool` result** would make the next request carry two consecutive `user` roles — the Gemini/Vertex wire (which rides `functionResponse` under `role:"user"`) **rejects it with HTTP 400** ("roles must alternate between user and model"), and the OpenAI-compatible wire rejects the alternation too. **An interrupted turn can only be persisted if it is closed with an assistant message.**

## Decision

1. **Persist the completed steps, closing the turn with a synthetic answer (D1).** When the turn is interrupted **and** `len(result.Steps) > 0`, `runTurn` appends **one** atomic `history.Entry` — `{prompt, answer: history.InterruptedTurnAnswer, calls: len(result.Calls), steps: result.Steps}` — via the existing `store.Append`. *Rejected:* persisting without a closing answer (violates role alternation — non-negotiable); fabricating tool steps.

2. **Typed interruption detection (D2).** The check is `errors.Is(err, context.Canceled)` — **typed, never string-matched** (NFR-003). It unwraps through the `*llm.ProviderError` the adapters return and through the round-078 retry decorator's "return the last error unchanged" path. `SIGINT` and `SIGTERM` both cancel the same context, so one predicate covers both.

3. **The synthetic answer is one named constant (D3).** `history.InterruptedTurnAnswer = "[Turn interrupted by operator via Ctrl+C]"` — owned by `internal/domain/history` (it is a **stored record value**), fixed/deterministic, control-free, single-line, unmistakably synthetic. A `SIGTERM` interruption shares the wording (the cause is not distinguished).

4. **Always-on; no config toggle (D4).** A completed-step interruption always keeps the work. **Declined alternatives (recorded, not forward items):** a config toggle; a double-Ctrl+C discard protocol; a per-provider switch — scope creep with a new failure surface; the operator's default intent is "keep what I did", and discarding remains available (`--new` archives the session).

5. **Exit code & diagnostic on the interrupted path (D5).**
   - *Interrupted with completed steps, partial save succeeds* → exit **`Success` (0)**, `stdout` **empty**, and one informational line on `stderr`: `[HH:MM:SS] interrupted by operator; kept N completed tool step(s) in the session history` — chrome-styled, control-free, carrying **no** `tellme: ` prefix (the round-017 *exactly-one-`tellme:`-line* contract holds) and **not** routed to `turns.log` (the ADR 0050/0038 precedent). Emitted in the `retryNotifier` style (an inline CLI line with the spinner already stopped).
   - *Interrupted with zero steps* → **today's surface, unchanged** (phrase + exit 6) — the issue's item 3 ("keep the current clean-abort behavior"); there is no saved work to signal success for.
   - *Partial-save `Append` fails* → the existing `emitHistoryError` path (environment phrase + exit 4); never a crash, never a half-written file.
   *Rationale for 0:* the operator deliberately stopped and the work was **saved** — nothing failed; it matches tellme's convention that an operator-initiated interruption is a `Success` (the round-012 reader), and it keeps the exit-code set **frozen at ten** (an eleventh code **130** would break `TestExitCodesMatchPinnedContract`). The synthetic answer is **not** printed to `stdout`.

6. **No token-usage persistence on the interrupted path (D6).** The completed calls' usage is not written to `tokens.log`: the existing `persistTurnUsage` gate (the *final* call's reported usage) already yields nothing on an interruption, and a failed turn writes no usage today — so **no new accounting path** is added. The bounded loss is recorded (§Forward).

7. **The step trigger is `len(result.Steps) > 0` (D7).** A step is recorded whenever a tool **returned** — including a tool killed mid-execution by the cancellation, whose `result` is a **nil-error** `Exit Code: -1` — **not** an `error: …` string (the loop's `error: ` prefix applies only to a non-nil tool error, e.g. a `Start` failure; a kill/timeout is a nil-error result, round-024 FR-018). A "completed tool step" therefore means a tool that **returned** a result (success or kill/timeout). This is the issue's literal rule; requiring a *successful* step would need a success predicate the loop does not expose.

8. **Records; the entry schema is unchanged (D10).** ADR 0051 + index; `techstack.md` MODIFY (the *Session history store* append-after-complete qualification) + ADD (the interrupted-turn persistence row); the CLI truth feature (`remembering-the-conversation.feature` + a new Rule/Examples) + `chat/dsl.md` rows; and — per the ADR-0041 same-PR rule — a **domain-model amendment** of `history-append-after-complete` (scope the exception) and `turn-answer-stored-verbatim` (the synthetic-answer exception), re-rendered. `history.Entry` / `history.Step` JSON shapes are **unchanged**.

9. **Scope (D9).** The prompt/tool-loop turn only; MCP-backed tool steps are ordinary steps on the same seam. The `-b`/`--retry` flags and the offline readers are untouched. Hermetic: no pty, no live network — the E2E lands a **real** `SIGINT` by scripting an `execute_command` whose command is `kill -INT $PPID; sleep 30` (a signal to the turn process itself), so the in-flight tool call is the interruption point and the completed step is what gets persisted; the resume leg arranges the persisted entry by hand (the established replay pattern).

## Consequences

- A long tool turn interrupted mid-flight keeps every completed step, so the operator can resume with *"continue"* / *"summarize what you found so far"* instead of starting over.
- The resumed request is structurally valid on **both** families: the partial turn replays ending with an `assistant` message (the synthetic answer), so the role sequence alternates.
- No exit code or class phrase is added (the vocabulary stays ten); the interrupted-with-work path exits **0** with `stdout` empty; the zero-step path is byte-identical to today.
- The turn counter advances by the rounds actually made (`calls` = the completed inference rounds); the cancelled call is not counted.
- `history.Entry` / `history.Step` are unchanged; one extra line is appended per interrupted turn (append-only preserved).
- **Domain-model amendment (recorded divergence from the reference):** the model's `history-append-after-complete` and `turn-answer-stored-verbatim` invariants gain the operator-interrupted exception — the reference discards the partial turn, so this is a recorded divergence *beyond* the reference.
- The completed calls' usage is absent from `tokens.log` for an interrupted turn (a deliberate, bounded loss).

## Forward

- **RF-079-1** — the synthetic answer is stored as the turn's `answer`; a downstream reader (`-l`) shows it as the model's "answer" (there is no marker beyond the text).
- **RF-079-2** — a step whose tool was killed mid-execution is counted (D7), so a 1-step partial turn can hold an error-result step.
- **RF-079-3** — no usage is persisted for the interrupted turn (D6).
- **RF-079-4** — the exit-0 vs zero-step-exit-6 asymmetry (D5).
- **RF-079-5** — `SIGTERM` shares the "via Ctrl+C" wording (D3).
- **RF-079-A** — no scenario chains the **product-written** interrupted entry into a resume (a second `tellme` run or `-l`); the resume leg arranges the entry by hand, so the entry's replay is witnessed by composition, not end-to-end (supersedes the stub RF-079-6, whose stated reason — a stall seam — does not exist: the real reason a signal *before* the first tool step cannot be landed is that the live signal is delivered **from inside** a tool call).
- **RF-079-B** — the aborted call's chrome frame (`╭─⠿ Turn N` + its estimated-payload line, no measured/metrics/`Ready` tail) reaches `stderr` **and** `turns.log` (the round-053 sink); suppressing the frame when the ctx is already cancelled is a candidate, not done (see `plan.md` §6 N-3).
- **RF-079-C** — the persisted `calls` now has an E2E carrier (fold F-079-2); every other boundary of the interrupted path (zero-step, append failure, non-cancellation) stays unit-carried by design.
- **RF-079-7** — no `history.archive.jsonl` interaction beyond the existing append/archive semantics.
- **RF-079-8** — the informational line is `stderr`-only, not in `turns.log` (consistent with the ADR 0050/0038 precedent).
- **RF-079-9** — the ~150 ms window between `ind.Stop()` and the append is not itself exercised (the interruption is observed at the request, not the teardown).
