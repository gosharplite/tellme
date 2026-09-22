# Technical Research: Preserve completed tool steps on an interrupted turn (round 079)

**Plan Package**: `specs/plans/079-interrupted-turn-partial-persistence`
**Created**: 2026-09-22
**Anchor**: issue [#159](https://github.com/gosharplite/tellme/issues/159)
**Grounding**: `dev` @ `4f9c96d`

> Decision-driven research (no code here). The **behaviour is locked by the issue** (persist the completed steps with a synthetic closing answer; discard when zero steps; keep the immediate child kill; role-alternation on resume). The issue's five *Open Design Decisions* are resolved below (D3–D7). The design is tellme's own; the reference is not asserted.

---

## 0. The problem, precisely

`history.jsonl` is replayed into the next request by `BuildMessages` (`internal/agent/agentloop.go`): per entry `user(prompt)` → per step `assistant(tool_call) + tool(result)` → `assistant(answer)`. The turn is persisted **only** when `loop.Run` returns `err == nil` (`internal/cli/cli.go` `runTurn`). On an operator interruption (`SIGINT`/`SIGTERM`) the turn context is cancelled (`signal.NotifyContext`, line 743); `Complete` returns a `*llm.ProviderError` wrapping `context.Canceled`, the loop returns `(Result{Steps, Calls}, err)`, and `runTurn` maps the error to `the provider request failed` + **exit 6**, **skipping the append** — so the prompt and every completed step are discarded.

Why a naive partial save is unsafe: a persisted turn that **ends on a `tool` result** replays as `… user(prompt) → [assistant(tool_call) → tool(result)]…` and then the next prompt adds another `user` — two consecutive `user` roles. The Gemini/Vertex wire carries `functionResponse` parts under `role:"user"` and **strictly rejects** it (HTTP 400); the OpenAI-compatible wire rejects the alternation too. **The turn must end with an assistant message.** Hence the synthetic answer.

---

## 1. Decisions

### D1 — Persist on interrupt with completed steps (**locked by the issue**; mechanism here)

When the turn context is cancelled **and** `len(result.Steps) > 0`, `runTurn` appends **one** atomic `history.Entry` — `{prompt, answer: <synthetic>, calls: len(result.Calls), steps: result.Steps}` — via the existing `store.Append`, then reports success (D5). Detection is at the CLI turn seam (the only place that holds both the loop `Result` and the `store`).

### D2 — Typed interruption detection (I-6)

The check is `errors.Is(err, context.Canceled)`. This is **typed, not string-matched** (never `strings.Contains(err.Error(), "context canceled")`). It unwraps correctly through the `*llm.ProviderError` the adapters return (`Unwrap()` → the `net/http` error → `context.Canceled`) and through the round-078 retry decorator's "return the last error unchanged" path. `SIGTERM` and `SIGINT` both cancel the same context, so both are covered by one predicate.

### D3 — The synthetic closing answer (S-6)

A single named constant, **`[Turn interrupted by operator via Ctrl+C]`** (the issue's suggested wording), owned by `internal/domain/history` (`history.InterruptedTurnAnswer`) — it is a **stored record value**, so it lives with the record, not in the CLI. Properties: fixed/deterministic, control-free, single-line, and unmistakably synthetic (bracketed + no `tellme:` prefix). A `SIGTERM` interruption shares the same text (the interruption cause is not distinguished — a deliberate simplification).

### D4 — Always-on, no config toggle (S-5)

A completed-step interruption **always** keeps the work; there is **no** config key and **no** dual-interruption (abort-and-discard) mode. Rationale: (a) the operator's intent when they press Ctrl+C mid-long-turn is overwhelmingly "stop, keep what I did"; (b) tellme's config carries **no** security/consent keys and adding a toggle is scope creep with a new failure surface; (c) discarding remains available to the operator **without** a flag — `--new` archives the session, and a `# [need clarification]`-free abandoned turn is simply not resumed. **Declined alternatives (recorded, not forward items):** a config toggle; a "double Ctrl+C = discard" protocol; a per-provider switch.

### D5 — Exit code & diagnostic on the interrupted path (S-7)

| Case | Exit code | Stream |
| --- | --- | --- |
| Interrupted **with** completed steps (partial save succeeds) | **`Success` (0)** | one informational line on `stderr`; `stdout` **empty** |
| Interrupted **with zero** steps (nothing to save) | **`ProviderError` (6)** — **today's surface, unchanged** | today's `tellme: the provider request failed: context canceled` |
| Partial save **append fails** | **`EnvironmentError` (4)** via the existing `emitHistoryError` | today's history phrase |

**Rationale for 0:** the operator *deliberately* stopped the run and the completed work was **saved** — nothing failed. This matches tellme's existing convention that an operator-initiated interruption is a success (the round-012 interactive reader returns `Success` on a cancelled read), and it keeps the exit-code set **frozen at ten** (`TestExitCodesMatchPinnedContract`) — **130** (128+SIGINT) would add an eleventh code and break the pinned contract for no operator benefit. **The zero-step case is deliberately left on today's path** (the issue's item 3: "keep the current clean-abort behavior") — there is no saved work to signal success for, and leaving it untouched is the minimal-risk choice.

The informational line is emitted by the CLI in the **`retryNotifier` precedent** (round 078): an inline, chrome-styled, **control-free** line on `stderr`, written with the spinner already stopped, carrying **no** `tellme: ` class prefix (so the round-017 "exactly one `tellme:` line" contract is untouched) and **not** routed to `turns.log` (the retry / unpaired-diagnostic precedent):

```text
[HH:MM:SS] interrupted by operator; kept N completed tool step(s) in the session history
```

The synthetic answer is **not** printed to `stdout` (the interrupt writes no answer) — `stdout` stays `empty`, so existing byte-exact `stdout` assertions are unaffected.

### D6 — No token-usage persistence on the interrupted path (S-8)

The completed calls' usage is **not** written to `tokens.log`. The existing gate (`persistTurnUsage` requires `result.Usage.Reported`, the *final* call's usage) already yields nothing on an interruption, and a **failed** turn writes no usage today — so **no new accounting path** is added; the behaviour matches today's failure semantics. The consequence (the interrupted calls' spend is missing from the token log) is **recorded** as a deliberate, bounded loss (the accounting log is a display/log artifact; a failed turn already loses it). **Alternative declined (recorded):** persist the *reported* completed calls — rejected as added surface with no operator value this round.

### D7 — Trigger semantics: `len(result.Steps) > 0` (S-12)

The trigger is the loop's recorded step count — a step is recorded whenever a tool **returned** (success, error, **or** the per-call timeout/kill result). A tool killed mid-execution by the cancellation still appends a step — but its `result` is a **nil-error** `Exit Code: -1`, **not** an `error: …` string (the loop's `error: ` prefix applies only to a non-nil tool error, e.g. a `Start` failure; a kill/timeout is a nil-error result, round-024 FR-018) — so a turn interrupted **during its first tool's execution** counts as "one step" and is persisted. "Completed tool step" therefore means a tool that **returned** a result (a success, or a kill/timeout result). This is the issue's literal rule (I-2 is `len(steps) == 0`), and it is the honest reading: the model *requested and saw* that tool call's outcome. Requiring a *successful* step would need a success predicate the loop does not expose and would drop a genuinely-attempted tool. **Recorded consequence:** an interrupted-in-flight first tool persists a single step whose result is a kill/timeout marker (`Exit Code: -1`).

### D8 — The persisted entry (schema unchanged)

```go
history.Entry{Prompt: prompt, Answer: history.InterruptedTurnAnswer, Calls: len(result.Calls), Steps: result.Steps}
```

One `Append` (one JSON line, append-only). **No** change to the `history.Entry` / `history.Step` JSON shape — the synthetic answer is the `answer` value. `calls` = the **completed** inference rounds (the cancelled call is not counted), consistent with the round-027 AI-endpoint-call semantics (the next turn's header advances by the rounds actually made). No per-step id is added (replay still synthesises `call_step_<n>`).

### D9 — The hermetic interruption seam (S-9)

The E2E must land a `SIGINT` **while the child is mid-turn**, deterministically, with **no pty and no live network**:

**The shipped seam (reconciled at fold F-079-1):** a **real** `SIGINT` to the turn process, produced **from inside a tool call** — no fake-provider stall mode and no harness signal helper were needed (the round declined both).
1. **The Given** scripts the fake to return **one** `execute_command` tool call whose command is `kill -INT $PPID; sleep 30` (a signal to tellme itself; the `sleep 30` keeps the tool in flight until the signal lands). `MAX_TOOL_LOOP=5` bounds an unlanded signal so it fails fast rather than running the default 1000-round bound.
2. **The sequence:** request 1 → the tool call executes → tellme's own `SIGINT` cancels the turn context → the in-flight command aborts → the loop returns `(Steps=[1], Calls=[1])` + the cancelled error → `runTurn` persists the partial turn and exits 0. The E2E asserts `history.jsonl` (one entry: the step count, the persisted `calls`, the synthetic answer), the `stderr` informational line, `stdout` empty, and exit 0.
3. **The resume leg** arranges the persisted interrupted entry **by hand** (the round-014 replay pattern) and asserts the replayed request's `messages` alternate and close the earlier turn with an `assistant` message. The **zero-step** negative (a signal *before* the first step) is carried by unit pins — it cannot be landed from inside a tool call; recorded as RF-079-A.

### D10 — Records (S-10)

- **ADR 0051** (`docs/decisions/0051-interrupted-turn-partial-persistence.md` + index row) — the decision record (this research's D1–D9).
- `specs/truth/techstack.md`: **MODIFY** the *Session history store* row (append-after-complete qualified: an operator-interrupted turn with completed steps is completed by a synthetic answer) and **ADD** a *Interrupted-turn persistence* row; the *Prompt input* row's `runTurn`-interruption note is reconciled.
- `docs/domain-model/tellme.modelith.{yaml,md}`: **MODIFY** `history-append-after-complete` (scope the exception) + `turn-answer-stored-verbatim` (the synthetic-answer exception) — a **recorded divergence from the reference** (which discards the partial turn); re-render with `make modelith-render` and keep `modelith-check` green.
- `specs/truth/features/cli/chat/remembering-the-conversation.feature`: a new **Rule** (the interrupted turn) + Examples; `chat/dsl.md`: the new Given/Then rows.
- `internal/domain/history/history.go` (the constant), `internal/cli/cli.go` (the detection + persist + line), `tests/e2e/**` (the seam).

---

## 2. Rejected alternatives (recorded — not forward items)

| Alternative | Why rejected |
| --- | --- |
| Persist without a synthetic answer (stop on the last tool result) | Violates role alternation — the next request 400s on Gemini/Vertex and is ill-formed on OpenAI. Non-negotiable (I-1). |
| Exit **130** (`128+SIGINT`) | Adds an eleventh exit code, breaking the pinned ten (`TestExitCodesMatchPinnedContract`); tellme's convention for an operator stop is `Success`. |
| Keep exit **6** on a *successful* partial save | Would report "the provider request failed" for an operator-initiated, successfully-saved stop — misleading. |
| A **config toggle** / double-Ctrl+C discard | Scope creep; a new config key and a new failure surface; the operator's intent is captured by the default. |
| Persist the completed calls' **usage** | No new accounting path this round; a failed turn already writes none (D6). |
| Restore exit-0 on the **zero-step** case too | The issue keeps the current clean abort for zero steps; changing it is out of scope (D5). |

---

## 3. Residual risks (→ §Forward of ADR 0051)

- **RF-079-1** The synthetic answer is stored as the turn's `answer` — a downstream reader (e.g. `-l`) shows it as the model's "answer"; there is no marker beyond the text itself.
- **RF-079-2** A step whose tool was **killed mid-execution** is counted (D7), so a `1`-step partial turn can hold an error-result step.
- **RF-079-3** No usage is persisted for the interrupted turn (D6).
- **RF-079-4** The exit-0/zero-step-exit-6 asymmetry (D5).
- **RF-079-5** `SIGTERM` shares the "via Ctrl+C" wording (D3).
- **RF-079-A** No scenario chains the **product-written** interrupted entry into a resume (a second `tellme` run or `-l`); the resume leg arranges the entry by hand (supersedes the stub RF-079-6 — the shipped seam has no stall mode).
- **RF-079-B** The aborted call's chrome frame (`╭─⠿ Turn N` + estimated payload, no tail) reaches `stderr` **and** `turns.log` (see `plan.md` §6 N-3).
- **RF-079-C** The persisted `calls` now has an E2E carrier (fold F-079-2).
- **RF-079-7** No `history.archive.jsonl` interaction is specified beyond the existing append/archive semantics.
- **RF-079-8** The informational line is `stderr`-only and not in `turns.log` — a trace of an interrupted turn is absent from `-t` (consistent with the retry/unpaired precedent).
