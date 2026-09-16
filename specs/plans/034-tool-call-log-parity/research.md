# Technical Research — round 034 (tool-call log parity)

Plan package: `specs/plans/034-tool-call-log-parity`
Owner of truth touched: `specs/truth/techstack.md` (rows: Agent tool loop · Turn chrome · Post-turn status lines · Turn progress spinner · `execute_command` · Read-only filesystem tools).

> Decision-driven research. All §-numbered decisions are **settled**; the grilled spec (`spec.md` G1–G10) is the input. The recorded divergences from `tell-me-go` are cited where they occur.

## Context / constraints

- **Rendering-only round** (FR-016): no tool schema, tool result, transport, `history.jsonl` shape, flag, exit code, or `stdout` byte changes. `stdout` stays byte-exact; everything here is `stderr`.
- **stdlib-only, POSIX-only, hermetic** (no pty): NFR-002/004/005.
- The reference behaviour was verified at source in `tell-me-go` (`internal/ui/renderer_metrics.go`, `internal/tools/workspace/shell.go`, `internal/agent/orchestrator/{engine,engine_phases,engine_execution,middleware}.go`).

## Decisions

### D1 — A pure `internal/ui` formatter layer
All rendering lives in `internal/ui` as pure functions (no I/O), mirroring the existing `FormatToolLog`/`FormatPayloadStatus`/`FormatMetrics` discipline; the timestamp comes from the shared `formatClock` token. New pure formatters:
- `FormatToolEngine(t, i, m) string` → `[HH:MM:SS] [Tool Engine] Step <i>/<m>` (FR-001).
- `FormatToolReason(t, reason) string` → `[HH:MM:SS] [Tool Reason] <reason>` (FR-002).
- `FormatToolAction(t, tool string, args map[string]json.RawMessage) string` → `[HH:MM:SS] [Tool Action] <tool>(<args>)` (FR-003).
- `FormatToolResult(t, tool, text string) string` → `[HH:MM:SS] [Tool Result] <tool>: <snippet>` (FR-004).
- `FormatToolOutputHeader(t) string` and `ToolOutputSeparator` (a fixed hyphen literal), plus a per-line writer `[HH:MM:SS] [Tool Output] <line>` (FR-010).

The existing `FormatToolLog` (round 022) is **retired**.

### D2 — Two loop observer seams; the CLI is the only renderer/accountant
The loop emits nothing itself for the frame/tail; it drives the **existing** `agentport.LoopObserver` seam, extended with:
- a **call-begin** hook carrying `(callIndex int, estimatedPayload int)`, fired once per `Complete`, right after the request is assembled;
- a **call-end** hook carrying `(callIndex int, usage llm.Usage, roundReasons []string, measuredPayload int, final bool)`, fired once per `Complete` return.

Because `loop.Observer` is today the **spinner** (`loop.Observer = sp`), the CLI wires a **composite presenter** that renders the block and conditionally drives the spinner. The `internal/cli.runTurn` composition root keeps all pricing (`ui.Pricing`), the session roll-up (`newUsageStore(res.Workspace).Totals()`) and the persistence flush; `internal/agent` gains no persona/pricing/store (ADR-0005 D1).

### D3 — Rune-safe truncation caps (recorded divergence)
`cap = maximum total rendered rune length`, evaluated on the **folded** value; ellipsis = exactly one **U+2026** counted **inside** the cap; cut on a **rune boundary** using `utf8.RuneCountInString` / a rune-safe slice.
- Argument values: fold `\n`/`\r`→space; truncate iff runes **> 189**; keep the first **188** runes + `…` (≤189 runes).
- Result snippets: fold; truncate iff runes **> 200**; keep the first **199** runes + `…` (≤200 runes).

Reuse the existing rune-safe precedents — `truncateToBudget` (`internal/infrastructure/tools/filesystem.go`) and `truncateBytes` (`internal/infrastructure/mcp/naming.go`) — rather than inventing a third; a small shared helper in `internal/ui`. **Divergence:** the reference cuts **bytes** and uses three ASCII dots (`[:186]+"..."` = 189 bytes; `[:197]+"..."` = 200 bytes, cut-then-fold). Tellme's rune totals equal 189/200; the glyph and rendered byte length differ.

### D4 — Argument rendering: sorted keys + `json.Number` (recorded divergence)
Decode the call's `Arguments` (**a raw `string`** — `llm.ToolCall.Arguments`) with `json.Decoder` + `UseNumber()`; **remove `reason` first**; sort the remaining keys **ascending** (byte-wise); render each `key: value`, joined by `, `.
- Values: `json.Number.String()` (raw literal — `1000000`, never `1e+06`), strings unquoted, booleans/`null` literal, arrays/objects compact (source order inside nested values).
- Unparseable → empty argument list (`<tool>()`).

The transports currently pass the provider's raw text through (`openai/client.go` copies `tc.Function.Arguments`; `gemini/client.go` takes `json.RawMessage`), so declaration order is stable **by transport accident**, not by contract — hence sorted keys. **Divergence:** the reference iterates a Go map and `%v`-formats `float64`. Ordering is **presentation-only**: `tc.Arguments`, the persisted `history.Step.Arguments`, and the replayed `functionCall` keep the model's original bytes (round-014 replay fidelity).

### D5 — Per-call frame; per-call tail with final-call deferral
The status **frame** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload line) is emitted once per AI-endpoint call, at call begin (the call-begin hook). `N` = the session's prior AI-call count + `k` — it **advances within a prompt** (the reference recomputes `SessionTurnsAtStart` per Turn; `GuardStep`, `engine_phases.go`).

The call's **tail** (grouped `[Tool Reason]` lines + measured payload line + metrics line + `╰─⠿ Ready`) is emitted at call end (the call-end hook), with the **final** call's tail **deferred** past `env.writeAnswer(...)`. This reproduces the reference's observable order (it streams the answer to `stdout` during inference, so its `PhasePersisting` tail already follows the text) and preserves the existing Rule "the post-turn status trails the answer". **Divergence (recorded):** none — the reference's tail is per call (one Turn per call).

### D6 — Injected per-call estimator seam
`AgentLoop` has no persona (the adapter owns it) and the CLI only has call 1's inputs. Add `AgentLoop.PayloadEstimate func(messages []llm.Message) int` (nil ⇒ no estimate), set in `runTurn` with a closure capturing `res.Person` + `agent.ToolDefs(reg)` (the augmented registry). The loop computes `wire := base + turn` before each `Complete` and passes it. Call 1's value is **byte-identical** to the previous once-per-prompt estimate; calls 2..k grow monotonically. Typed on messages (not `llm.Request`) so the adapter's prompt-vs-messages rule is not duplicated. It is the **only** path to that number (`emitPayloadStatus`'s hard-wired call is retired for calls). (ADR-0005 D2.)

### D7 — `[Tool Output]`: shell-class only, live, single-writer, bounded-and-stopped
- **Scope:** shell-class calls that do **not** carry `output_file` (NFR-001). `output_file` (`runToFile`) binds the child's streams straight to a file handle — no bytes enter memory — so **no** block renders (G4).
- **Shape:** header `[HH:MM:SS] [Tool Output] Executing... (Output shown below)` at **call begin** (before the child starts), a fixed separator line, each **complete** stdout/stderr line as it arrives, a closing separator. A trailing partial line is **dropped, not flushed** (G4).
- **Sink:** a struct-bound, prompt-path-lazily-bound per-line sink on the command tool (round-029 TD-1 struct-field `contentWriter` / round-033 FR-009 lazy `bindSkillsCatalog` precedent), so `agentTools()` stays parameterless and read-free and round-031's assembler gate is untouched. `command.go runCaptured` must tee both `StdoutPipe`/`StderrPipe` through the sink while keeping the `boundedBuffer` byte budget intact. The sink **owns its own mutex** (two `io.Copy` goroutines).
- **Bounded-and-stopped (G4):** the stream carries every line up to the call's round-024 byte budget (`callByteBudget` → 1,000,000 B at the shipped budget); on overflow `runCaptured` sets `trimmed` and `abortCapture`→`killGroup` (SIGKILL) — the bound is observable as a **stop, not a cap**. **Divergence:** the reference's `warnWriter` has no such termination.
- **Single-writer spinner yield (G8):** the presenter clears/stops the spinner **before** the child starts; the sink owns `stderr` for the whole block; the presenter resumes after the closing separator. The block renders **unconditionally** (pause/resume is a no-op when the spinner is gated off). The `BeforeToolLog`/`AfterToolLog` pair is **not** the mechanism (loop-goroutine-only; `deactivate()`'s positional `lastRows` clear cannot be issued from a tool goroutine).

### D8 — Persistence cadence unchanged; a recorded display-only divergence
Persistence stays **one `AppendBatch` per turn** after `loop.Run`, gated on the **final** call's `Reported` flag (round-018 FR-012). The one-append-per-call alternative is **rejected** (it would persist calls of a turn that later fails — a durability change; round 030). Consequence (recorded, FR-010b): on a failed turn, or a completed turn whose final call reports no usage, the per-call `Ready` session field names totals that are never persisted (G2). Pinned invariant: the persisted record set is the `Reported` subset of `result.Calls` on a completed turn and **empty** on every error exit.

### D9 — `Step i/M` is in executed-round units
`MAX_TOOL_LOOP` bounds **executed tool rounds**, but the loop performs one extra `Complete` at the bound (the bound-reached call runs no round). Emit `[Tool Engine]` at the **execution site** (after the `i >= maxLoops` check, immediately before the round's calls), so `i` and `M` are both executed-round units and `1 ≤ i ≤ M` holds **by construction**; the bound-reached call renders its **frame** but **no** `[Tool Engine]` line. **Divergence:** the reference guards on the AI-call index before inference, so it never makes or frames an over-bound call.

### D10 — No new dependency; hermetic
`json.Number`/`utf8` are stdlib. The estimator seam is injected. The E2E layer drives the tool-loop rendering through the existing fake provider + the `TELL_ME_FORCE_STDERR_TTY`/`TELL_ME_FORCE_STDIN_TTY` seams (no pty); the `[Tool Output]` block is driven by a scripted `execute_command` (e.g. `printf`/`yes`) under the existing command-tool fixtures. `stdout` byte-exact.

## Recorded divergences from `tell-me-go` (summary)
1. `[Tool Engine]`-family decomposition **reverses** round 022's operator-chosen single line.
2. Rune-safe caps + one U+2026 vs byte slices + three ASCII dots.
3. Sorted keys + `json.Number` vs map iteration + `%v` on `float64`.
4. `[Tool Output]` bounded-and-stopped vs the reference's un-terminated `warnWriter`.
5. `Step i/M` in executed-round units vs the reference's AI-call guard.
6. `output_file` emits no block; trailing partial line dropped.
7. The display-only `Ready` session overstatement (G2) has no reference analogue (the reference persists per Turn).

## Truth impact
`specs/truth/techstack.md` MODIFY on the six rows named above; `specs/truth/features/cli/chat/**` MODIFY (see `truth-delta.md`); `specs/truth/contracts/**` NOOP (no HTTP surface); `specs/truth/data/data-model.dbml` NOOP (**re-derived**: persistence cadence + `usage_record` shape unchanged — D8).
