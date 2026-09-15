# Phase 0 Research: tellme Tool-Loop Log Line Reshape (Round 022)

Topic: reshape the per-call diagnostic line tellme writes to `stderr` during a tool-using run into a single operator-chosen line — `[HH:MM:SS] [Tool] <tool name> - <reason>` when the call states a reason, else `[HH:MM:SS] [Tool] <tool name>` — dropping the raw `arguments=<json>` and truncated `result=<…>` segments the round-008/021 line carries, and emit **one blank line** between the tool-log block and the final answer of a tool-using turn. A **pure diagnostic-rendering** round: no new system end, no new dependency, no schema, transport, history, flag, or exit-code change.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), the agent tool loop (`AgentLoop`, `MAX_TOOL_LOOP`, per-tool timeout, the `stderr` tool-loop log), the payload status line (round 009), the turn chrome (round 017), the post-turn status lines (round 018), and the turn spinner (round 019) were locked in rounds 001–021. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. **Settled exclusions** (out of scope): `SafePath`/consent, `-b`/`--retry`, history pinning, streaming, token-budget pruning. The **operator locked Q1–Q3** before `/axb-specify` (see `spec.md`); this document fixes the *how*.

Reference: `~/tmp/github/gosharplite/tell-me-go/internal/ui/renderer_metrics.go` (`LogToolCall` / `LogToolResult` / `metrics_logToolReason`).

---

## Decision 1: The tool-loop log line becomes a single timestamped line `[HH:MM:SS] [Tool] <name> - <reason>`

- **Decision**: Replace the round-008/021 line `[tool] <name> arguments=<json> reason=<reason> result=<result>` with one line per tool call, in call order: `[HH:MM:SS] [Tool] <tool name> - <reason>` when the call states a top-level `reason`, else `[HH:MM:SS] [Tool] <tool name>`.
- **Rationale**: Q1/Q3 (operator). The current line dumps the call's raw arguments and a truncated result into the operator's diagnostics, duplicating the tool output (the pasted run rendered the repo tree **twice** — once in `result=`, once as the answer). The new line keeps the operator's "what did the agent do, and why" signal — the tool name plus its stated reason — and nothing else.
- **Alternatives considered**:
  - The reference's decomposed shape — `[Tool Engine] Step N/M`, `[Tool Reason] <reason>`, `[Tool Action] <name>(<args minus reason>)`, `[Tool Result] <name>: <snippet>` (`renderer_metrics.go`) — rejected: the operator chose a single compact line, not the four-line decomposition.
  - Keep the round-021 `[tool] … reason=<value>` line — rejected: the raw argument/result dumps are exactly the defect.

## Decision 2: The timestamp comes from the standing injected clock seam (a new `internal/ui` formatter + an `AgentLoop.Now` field)

- **Decision**: Add a pure formatter `internal/ui.FormatToolLog(t time.Time, name, reason string) string` reusing the `15:04:05` format already used by `FormatPayloadStatus` / `FormatInputCaptured`; add a `Now func() time.Time` field to `AgentLoop` (nil → `time.Now`), mirroring the round-009/017 clock seam; `logStep` calls `a.now()` at emission. The CLI sets `loop.Now = env.now` in `runTurn`, so the loop and the chrome/payload lines share one clock.
- **Rationale**: keeps `internal/ui` a pure formatter package (round-019 precedent), keeps the loop the single tool-log emitter (round-008 Decision 7), and makes the timestamp deterministic in tests (round-009 Decision 4). Reusing the `15:04:05` format keeps the three `stderr` surfaces (`Payload:`, `Input captured.`, `[Tool]`) visually consistent.
- **Alternatives considered**:
  - Format inline in `agentloop.go` — rejected: heavy formatting stays in `internal/ui`; the loop currently imports only `fmt`.
  - Reuse an existing exported formatter — rejected: the payload/input-hint shapes differ, so none fits.
  - Pass a preformatted timestamp string into the loop — rejected: the loop emits per call as it runs, so it needs the clock, not one fixed value.

## Decision 3: Drop the `arguments=` and `result=` segments entirely

- **Decision**: The line carries neither the call's arguments nor its (truncated) result.
- **Rationale**: Q1 (operator). Those segments are what make the line noisy and duplicate the tool output. The result is **not** lost — it still goes to the model (the tool-role message) and to `history.jsonl` (the persisted step); it is removed only from the operator log.
- **Alternatives considered**: keep a truncated `result=` (rejected — the operator removed it); keep `arguments=` (rejected — same).

## Decision 4: A call with no top-level `reason` renders `[Tool] <name>` (generic extraction kept)

- **Decision**: Keep the generic round-021 extraction — a top-level `reason` string in the call's arguments JSON. When absent (a non-filesystem/other tool, or malformed arguments), the line is `[HH:MM:SS] [Tool] <tool name>`: no separator, no placeholder. The `reason` text is folded to one line (`oneLine`) so the log stays single-line.
- **Rationale**: Q3 (operator). The tool schemas still **require** `reason` (round-021 D2 unchanged), so the three filesystem tools always carry one; the log stays generic so any other/MCP tool logs cleanly without a reason.
- **Alternatives considered**: a placeholder (`(no reason)`) or a trailing ` - ` (rejected — Q3); assume `reason` is always present (rejected — the loop logs generically).

## Decision 5: The blank line before the answer is emitted by the CLI, only on tool-using turns

- **Decision**: On a completed turn with `len(result.Steps) > 0`, the CLI writes **one** blank line to `stderr` immediately before writing the answer (`env.writeAnswer`). A run with no tool steps writes nothing extra.
- **Rationale**: Q2 (operator). The CLI owns the answer write and the existing turn-chrome gap (round 017), so it is the precise emission point; emitting immediately before the answer guarantees the merged (`2>&1`) ordering `tool line(s) → blank → answer`. The loop must **not** emit it: it also returns on the error/incomplete paths where no answer follows (a stray blank), and it does not know an answer will follow.
- **Alternatives considered**:
  - The loop emits a trailing blank after the tool block — rejected: stray blank on a non-answer exit.
  - Always emit before the answer — rejected: Q2 (only tool-using turns).
  - Emit the blank on `stdout` — rejected: it would change `stdout` bytes (FR-008).

## Decision 6: Streams — the line and the blank line stay on `stderr`; `stdout` byte-exact

- **Decision**: Both the tool-log line and the separating blank line go to the diagnostic stream (`stderr`). `stdout` is unchanged.
- **Rationale**: FR-008; matches the round-008/017 convention (diagnostics on `stderr`, the answer on `stdout`).
- **Alternatives considered**: `stdout` (rejected — breaks byte-exactness and the piping contract).

## Decision 7: Testing — a new formatter unit test + an updated reason test; E2E re-pin; no pty

- **Decision**:
  - **Unit** (stdlib `testing`): `internal/ui.FormatToolLog` (the `[HH:MM:SS] [Tool] <name> - <reason>` shape with an injected time; the no-reason form; the `15:04:05` format); update `internal/agent/agentloop_reason_test.go` to the new shape — assert the line names the tool, carries the reason, and contains **no** `arguments=` / `result=`; assert the no-reason form.
  - **E2E** (godog): re-pin `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — the reason Then moves from `reason=<value>` to the `[Tool] <tool> - <reason>` shape, and a Then is added for the blank line before the answer on a tool-using turn (plus the negative: no added blank line on a no-tool turn). Keep the ordering Thens (`… before the answer`) — they already run on the merged-stream capture (round 010).
  - No pty; the existing hermetic harness + the round-010 merged-stream witness suffice.
- **Rationale**: the line shape is a pure formatting concern (unit-testable with an injected clock); the blank-line **separation** is an emit-order concern that only the merged-stream witness can assert (round 010 Decision 2).
- **Alternatives considered**: unit-only (rejected — the blank-line ordering needs the merged view); E2E-only (rejected — a fixed clock makes the exact line cheaply unit-pinnable).

## Decision 8: No new dependency; no schema/loop-contract change

- **Decision**: stdlib-only; `go.mod`/`go.sum` untouched. The tool schemas, the `MAX_TOOL_LOOP` bound, the `the tool request failed`/exit-7 contract, the frozen class-phrase vocabulary (11), and the payload/spinner surfaces are unchanged.
- **Rationale**: a pure rendering change (Q1 strict scope); nothing beyond the standard library is needed.
- **Alternatives considered**: widen the loop contract or the schemas (rejected — scope); touch the payload/spinner lines (rejected — Q1).

---

## Residual risks / forward links

- **Operator-chosen shape, not reference parity**: the reference decomposes tool activity into `[Tool Engine] Step N/M`, `[Tool Reason] <reason>`, `[Tool Action] <name>(<args minus reason>)`, and `[Tool Result] <name>: <snippet>` lines (`renderer_metrics.go`). tellme **deliberately diverges** — a single `[Tool] <name> - <reason>` line. Recorded so this is not later read as a parity bug.
- **`-i` submit path**: the round-017 frame is gated by `chrome` (false on the `-i` submit path), but the loop's tool-log line is **ungated**, so the blank line (which follows the tool-log lines) is likewise **ungated** — it appears on **any** tool-using turn, including `-i` submissions. The DSL must pin the trigger as "≥1 tool log line", **not** "chrome on".
- **Ordering**: the merged-stream witness must show `tool line(s) → blank → answer`; the round-018 post-turn lines continue to trail the answer.
- **Interface truth to update** (`/axb-dsl-refine`): MODIFY the `chat/dsl.md` **`the run reported the reason "{reason}" for the tool call "{tool}"`** row (its `必查` moves from `reason={reason}` to the `[Tool] <tool> - <reason>` shape) and the **`the run reported the tool call "{tool}" on its diagnostic output"`** row (note the timestamped `[Tool]` line); ADD a Then row for the blank-line separation and, if needed, a Rule/Example in `watching-the-tool-loop.feature`. **`/axb-api-plan` and `/axb-data-plan` are expected NOOP** (no request/response contract, no persisted-record change).
- **`stdout` byte-exactness**: a regression pin that a tool-using run's `stdout` is unchanged is required (FR-008) — the existing byte-exact assertions cover it; nothing new lands on `stdout`.

---

## Review fold — PR #50 (round 022)

The plan+truth half was reviewed (PR #50). The following findings are resolved **in-round**:

- **B1 (blocker) — FR-007's negative had no executable carrier + was ambiguous vs the round-017 frame gap.** Resolution: added the negative Then **`the tool loop added no blank line before the answer`** + a `chat/dsl.md` row + a Rule/Example, carried on the **non-chrome `-i` submit surface** (`renderTurn(..., turnOptions{raw})`, `chrome=false`), where no frame gap exists — so it pins round-022's blank specifically, not round-017's `emitTurnGap`. The `separating-the-tools-from-the-answer.feature` acceptance Example 2 is also reworded to make the surface explicit.
- **TD1 — the "ungated" claim was prose-only.** Resolution: added an **`-i` tool-using Example** asserting `the tool report is separated from the answer`, so a future implementation that wrongly gates the blank on `chrome` fails. The ungated property is now executable.
- **TD2 — the no-reason Example drove `read_files`, whose schema requires `reason`.** Resolution: documented explicitly (feature comment + module note) that the fixture **deliberately scripts a schema-nonconforming call** (a `read_files` call without `reason`, which round-021 D2 forbids) to cover the **generic** renderer for a tool that states no reason — exactly Decision 4's stated purpose.
- **TD3 — FR-005 "in call order" was not carried.** Resolution: added the ordered Then **`the run reported the tool calls in order "{tool_a}" and "{tool_b}"`** + a `chat/dsl.md` row + the sequential two-tool Given **`a configured provider "…" whose endpoint shows the folder tree and then reads "…" and then answers with "…"`** (tree → read → answer), and a new Rule using them.
- **R1 — stale acceptance pointer.** Resolution: `watching-the-tool-loop.feature`'s header now points at `reporting-each-tool-use.feature` + `separating-the-tools-from-the-answer.feature`.
- **R2 — clock-format duplication.** Resolution (implementation): factor a small `formatClock(t) string` in `internal/ui`, used by `FormatToolLog` and the existing `FormatPayloadStatus` / `FormatInputCaptured`, so the three `stderr` surfaces stay in lockstep (carried in `tasks.md` T001/T011).

**Implementation notes carried into `tasks.md`** (non-blocking, from the review): emit **exactly one** `\n` for the blank line, positioned **after** `sp.Stop()` and **before** `env.writeAnswer(...)`; set `loop.Now = env.now`; keep the nil→`time.Now` fallback for unit tests; re-check the round-010/017 ordering Thens in the regression scope.

---

## Review fold 2 — PR #50 (round 022), head `8c38579`

Second review: **FULL ARCHITECTURAL APPROVAL — CERTIFIED READY FOR IMPLEMENTATION**. Non-blocking recommendations folded:

- **R-1 (refactor) — extend the `formatClock` consolidation to all `internal/ui` timestamp formatters.** Folded into `tasks.md` T001/T011: the shared `formatClock(t)` is used by `FormatToolLog` **and** the existing `FormatPayloadStatus`, `FormatInputCaptured`, **and `FormatMetrics`** (`internal/ui/metrics.go`), so all four `stderr` diagnostic formatters share one clock-format token.
- **TD-1 (debt) — stale PR description metrics.** The PR body is updated to `tasks.md (T001–T016)` and `213 module rows · 1087 steps` (branch state).

Implementation directives (for `/axb-implement`) folded into `tasks.md` T014:

1. **Preserve the `LoopObserver` hook lifecycle** — `logStep` keeps wrapping emission with `a.Observer.BeforeToolLog()` / `a.Observer.AfterToolLog()` (no round-019 spinner regression).
2. **Blank line emission placement** — in `internal/cli/cli.go` (`runTurn`), write the single newline (`fmt.Fprintln(env.stderr)`) strictly **after** `sp.Stop()` **and** `store.Append`, and immediately **before** `env.writeAnswer(...)`.
3. **Clock-seam fallback** — `AgentLoop.now()` returns `a.Now()` when set, else `time.Now()`.
