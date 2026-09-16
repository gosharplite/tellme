# Feature Specification: tool-call log parity — the reference's per-call tool rendering (round 034)

**Feature Branch**: `034-tool-call-log-parity`

**Created**: 2026-09-16

**Status**: Draft — operator-locked decisions Q1–Q7 **plus grill-round (PR #70) folds G1–G10**. See [`docs/decisions/0005-tool-call-log-parity.md`](../../../docs/decisions/0005-tool-call-log-parity.md).

**Input**: Operator request: *"I want tellme to show tool calls similar to tell-me-go."* The operator supplied the reference (`tell-me-go`) and tellme renderings side by side.

- **tellme today** writes **one** line per tool call — `[HH:MM:SS] [Tool] <name> - <reason>` (round 022), deliberately omitting the call's arguments and its result.
- **tell-me-go** decomposes each call into `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` lines and emits a status block **per AI-endpoint call**.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → 1 — full per-AI-call parity.** tellme emits the status block once per **AI-endpoint call**, with the decomposed tool-call rendering inside. `[Tool Reason]` is printed **twice**: once immediately after `[Tool Engine]`, and once **grouped at the post-call tail** (the round's reasons, re-emitted after the results, before the measured payload line).
- **Q2 → 1 — `[Tool Output]` is a live stream.**
- **Q3 → 1 — verbatim reference templates & constants** (amended by G3 for rune-safety; see below). **All tellme tool calls carry `reason`**.
- **Q4 → 1 — keep tellme's status-block line formats** (header/payload/metrics/`Ready`); only the cadence and the tool-call rendering change.
- **Q5 → 1 — the `[Tool Output]` stream is not artificially capped** (the operator's signal that a command is producing a lot) — **amended by G4**: the round-024 byte budget already bounds and *stops* it.
- **Q6 → 1 — keep the round-019 spinner** as the live indicator — **amended by G6/G8**: it is yielded once per call (single-writer), not per line.
- **Q7 → 1 — all prompt surfaces**; `-r` does **not** suppress the tool-call lines.

## Grill-round folds (PR #70) — G1–G10

A grill round (`architect` subject vs `griller`, both initialised by executing `SESSION-BOOTSTRAP.md`) ran 10 verified questions against this specification. Verdict: **proceed with changes**. The following corrections are now folded **into this spec** (full transcript: see the PR #70 grill comment / gist):

- **G1 — one observer seam pair; estimation stays in the CLI.** The loop exposes a **call-begin** hook `(callIndex, messages []llm.Message)` and a **call-end** hook `(callIndex, usage, roundReasons, final)`. The **CLI** observer computes the estimate from the messages (`llm.EstimatePayload(person, toolDefs, messages)`) — **no estimator closure is injected into `AgentLoop`** — and `usage.PromptTokens` *is* the measured payload, so no separate `measuredPayload` is passed. Pricing / session roll-up / the persistence flush stay in the CLI (`runTurn`); `AppendBatch` (one batch per turn) is unchanged.
- **G2 — display vs persistence divergence (recorded).** With a per-call tail, a turn that **fails** (`the tool request failed` / `the provider request failed`), or a **completed** turn whose **final** call reports no usage, prints `Ready` footers whose **session** field includes calls that will **never** be persisted (round 030 discards a failed turn's usage; `if !result.Usage.Reported` gates persistence on the final call). Pinned invariant: the persisted record set is the `Reported` subset of `result.Calls` written by **one** `AppendBatch` on completion, and **empty** on every error exit. Recorded as an explicit **display-only divergence**.
- **G3 — canonical truncation rule (normative).** Cap = **maximum total rendered RUNE length**, the ellipsis is exactly one **U+2026 `…` counted inside the cap**, the cut is on a **rune boundary** (never a byte slice), and the predicate is evaluated on the **folded** value. Argument values (reason excluded): fold `\n`/`\r`→space, truncate iff folded runes **> 189**, keep the first **188** runes + `…` → **≤189 runes**. Result snippet: fold first, truncate iff **> 200**, keep the first **199** runes + `…` → **≤200 runes**. **Recorded divergence:** the reference cuts **bytes** and marks with **three ASCII dots** (its `[:186]+"..."` = 189 *bytes*; `[:197]+"..."` = 200 *bytes*, cut-then-fold); tellme's rendered **rune** totals equal 189/200 while the glyph differs (+2 rendered bytes for ASCII input). The false `truncateToCap` citation is struck — the real precedents are `truncateToBudget` (`internal/infrastructure/tools/filesystem.go`) and `truncateBytes` (`internal/infrastructure/mcp/naming.go`).
- **G4 — `[Tool Output]` is bounded-and-stopped, not "unbounded".** `AgentLoop.callByteBudget` (1,000,000 B at the shipped budget) makes `runCaptured` set `trimmed` and `abortCapture`→`killGroup` (SIGKILL) — the byte budget **ends the stream and kills the producer**. FR-011 is restated: the stream carries every complete line the child produces **up to that call's round-024 byte budget**, after which the process group is **stopped** (the existing `trimmed` outcome + `TruncationMarker`); the bound is observable as a **stop, not a cap**. `output_file` calls (`runToFile`) emit **no** `[Tool Output]` block (their bytes never enter memory). A **trailing partial line is dropped, not flushed** (reference parity). The "deliberate CTRL+C signal" rationale is withdrawn as unachievable.
- **G5 — the tail trails the answer.** The existing Rule "the post-turn status trails the answer" (`presenting-the-post-turn-status.feature`) would break on a 1-call plain turn under a uniform per-call tail. **Last-call deferral**: the final call's tail is emitted **after** `writeAnswer`, so the closing status still trails the answer.
- **G6 — the reference tail is per call (A5's single-tail claim is withdrawn).** The status middleware wraps **every** phase processor; `ExecutionStep` returns `NextPhase: PhasePersisting` on both exits; `Engine.Run` allocates a fresh Turn per call ⇒ **N calls = N tails**. Per-call tails are reinstated (each call's grouped reasons + measured payload + metrics + `Ready`), with G5's final-call deferral. The header frame carries the **same** `N` per prompt? — **No (corrected):** `GuardStep` recomputes `Turn.SessionTurnsAtStart` per Turn, so the reference's number **advances within a prompt**; tellme's k-th frame carries `prior-calls + k`.
- **G7 — argument ordering + number rendering.** Decode `Arguments` with `json.Decoder` `UseNumber()`; **remove `reason` first**; sort the remaining keys **ascending** (byte-wise); render `key: value` joined by `, `; numbers render their **raw literal** (`1000000`, never `1e+06`), strings unquoted, booleans/`null` literal, arrays/objects compact. Unparseable arguments → `[Tool Action] <tool>()`. **Recorded divergence** from the reference's `%v`-on-`float64`. Ordering is a **presentation** rule in the pure `internal/ui` formatter; it never mutates `tc.Arguments` or the value fed back to the provider (round-014 replay fidelity).
- **G8 — single-writer spinner yield.** `[Tool Output]` lines come from **two tool-owned `io.Copy` goroutines** while the loop is blocked in `tool.Execute`; `BeforeToolLog`/`AfterToolLog` are loop-goroutine-only and `deactivate()`'s positional `lastRows` clear cannot be issued from a tool goroutine. The block is **yielded once per call**: the presenter clears/stops the spinner **before** the child starts; the sink owns `stderr` exclusively for the whole block (its own mutex serialising the two copy goroutines); the presenter resumes after the closing separator. The block renders **unconditionally** (pause/resume is a no-op when the spinner is gated off). Consequence: the spinner is **not visible** while a shell command streams (reference parity), and the turn-scoped elapsed survives (round-019 D4).
- **G9 — `Step i/M` unit.** `MAX_TOOL_LOOP` bounds **executed tool rounds**, so a bound-reached call must not render `Step 1001/1000`. `i` and `M` are both in **executed-round** units; the `[Tool Engine]` line is emitted at the round's **execution site** (after the `i >= maxLoops` check) so `1 ≤ i ≤ M` holds **by construction**; the bound-reached call renders its **frame** but **no** `[Tool Engine]` line. **Recorded divergence**: the reference guards on the call index before inference, so an over-bound call is never made/framed.
- **G10 — the CLI computes the estimate from the observer's messages.** `AgentLoop` gains **no** estimator field and **no** persona (the adapter owns it). The call-begin hook carries the loop's **fused** `base+turn` wire slice (not `llm.Request`, so the adapter's prompt-vs-messages rule is not duplicated); the CLI observer runs `llm.EstimatePayload(res.Person, agent.ToolDefs(reg), messages)`. Call 1's value is byte-identical to today's `emitPayloadStatus`; calls 2..k grow monotonically. It is the **only** path to that number (`emitPayloadStatus`'s hard-wired call is retired for calls).

**Recorded ADR**: `docs/decisions/0005-tool-call-log-parity.md` (observer-port extension, the `PayloadEstimate` seam, the per-call cadence + final-call deferral + the display/persistence divergence, and the sorted-keys / `json.Number` / rune-cap divergences).

**Scope note**: a **tool-call diagnostic rendering** round, not a capability round. It reshapes what tellme writes to `stderr` during a tool-using turn and how the turn's status frames are laid out. It changes **no** tool schema, tool result, provider transport, session-history record shape, CLI flag, exit code, or `stdout` byte. It **supersedes** round 022's single-line `[Tool]` log and its blank-line rule, and re-cuts rounds 017/023/027's turn-chrome cadence to **per AI call**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The decomposed tool-call log (Priority: P1)

As a developer reading tellme's terminal, I want each tool call shown the way tell-me-go shows it — a step counter, the reason, the action (tool + arguments), and the result — so I can see *what* the agent ran and *with what arguments*, not just the tool name.

**Why this priority**: the operator's primary request and the substance of the feature.

**Independent verification**: run a tool-using turn (scripted provider asks for a tool with `reason` + args); assert `stderr` carries, in order, `[Tool Engine] Step i/M`, `[Tool Reason] <reason>`, `[Tool Action] <tool>(sorted k: v)`, `[Tool Result] <tool>: <snippet>`, and a grouped `[Tool Reason]` at the post-call tail; assert no round-022 `[Tool] <name> - <reason>` line remains.

**Acceptance Scenarios**:

1. **Given** a tool-using turn, **When** the loop **executes** a tool round, **Then** `stderr` carries `[HH:MM:SS] [Tool Engine] Step <i>/<M>` (i = 1-based round index, M = the effective `MAX_TOOL_LOOP`) and, per call, `[Tool Reason]` then `[Tool Action] <tool>(<args>)`; and a call that executes **no** round emits **no** `[Tool Engine]` line.
2. **Given** a call whose arguments carry a top-level `reason`, **When** it is logged, **Then** the `reason` renders as its own `[Tool Reason]` line and is **excluded** from the `[Tool Action]` argument list (and **not counted** in the sort).
3. **Given** an argument value whose folded length exceeds 189 runes, **When** the action is rendered, **Then** the rendered value is **≤189 runes**: the first 188 runes + one U+2026 `…`; the cut is on a rune boundary.
4. **Given** a completed call with a textual result, **When** the result is rendered, **Then** `stderr` carries `[HH:MM:SS] [Tool Result] <tool>: <snippet>` — the `Text` folded (newlines → spaces) **then** truncated iff > 200 runes → first 199 runes + `…` (**≤200 runes**), rune-safe.
5. **Given** a round with several tool calls, **When** the round completes, **Then** each call's `[Tool Reason]` precedes its `[Tool Action]` (interleaved per call), and at the post-call tail the round's reasons are re-emitted, grouped, before the measured payload line.
6. **Given** arguments that do not parse as a JSON object, **When** the action is rendered, **Then** the argument list is empty: `[Tool Action] <tool>()`.

**Functional Requirements**:

- **FR-001**: The tool-call diagnostic MUST emit `[HH:MM:SS] [Tool Engine] Step <i>/<M>` once per **executed** tool round, at that round's execution site (immediately before its calls, after the loop bound check); `i` = the 1-based round index within the turn, `M` = the effective tool-loop bound (`MAX_TOOL_LOOP`, default `1000`); `1 ≤ i ≤ M` MUST hold. A call that requests tools but executes none MUST emit **no** `[Tool Engine]` line (G9). Recorded divergence: the reference guards on the AI-call index before inference.
- **FR-002**: For each call, tellme MUST emit `[HH:MM:SS] [Tool Reason] <reason>` (from the loop's existing top-level `reason` extraction) immediately before the call's `[Tool Action]` line, when the `reason` is non-empty.
- **FR-003**: For each call, tellme MUST emit `[HH:MM:SS] [Tool Action] <tool>(<args>)`, where `<args>` is: the call's `Arguments` decoded with `json.Decoder` + `UseNumber()`; the `reason` key **removed first**; the remaining keys **sorted ascending**; each rendered `key: value` and joined by `, `. Values render: numbers via `json.Number.String()` (the raw literal — `1000000`, never `1e+06`), strings unquoted, booleans/`null` literal, arrays/objects as compact JSON (source order inside nested values). The **cap is a maximum total rendered rune length of 189**: fold `\n`/`\r`→space first, truncate iff the folded rune count **> 189**, keep the first **188** runes + exactly one **U+2026 `…` counted inside the cap**, cut on a **rune boundary**. Unparseable `Arguments` render the argument list as empty (`<tool>()`). Ordering/value rendering MUST NOT mutate the call's arguments or the value fed back to the provider (round-014 replay fidelity). (G3, G7)
- **FR-004**: For each completed call, tellme MUST emit `[HH:MM:SS] [Tool Result] <tool>: <snippet>`, where `<snippet>` is the result's text folded (newlines → spaces) **first**, then truncated iff the folded rune count **> 200** to the first **199** runes + one **U+2026** (**≤200 runes** total), cut on a **rune boundary**. (G3) A binary result is **not** rendered specially — a binary file surfaces as the readers' inline text marker `(Binary file, cannot display as text)` on this same line (the reference's `Received <mime> (<n> bytes)` branch is **struck**: the `Tool` port has no binary channel).
- **FR-005**: At the **post-call tail** of each call, tellme MUST re-emit the round's reasons, grouped, as `[HH:MM:SS] [Tool Reason] <reason>` lines — immediately before that call's measured payload line (the second `[Tool Reason]` occurrence).
- **FR-006**: Every tellme tool call carries a top-level `reason`, so `[Tool Reason]` always renders. A call whose `reason` is empty/absent MUST render **no** `[Tool Reason]` line (defensive tolerance); round 022's "reports a call that states no reason" behavior is **retired**.
- **FR-007**: The lines MUST be emitted one per event, in call order within a round, each prefixed with a `[HH:MM:SS]` (24-hour, zero-padded) timestamp from the shared clock seam.

---

### User Story 2 - A status frame per AI-endpoint call, with the tail trailing the answer (Priority: P1)

As a developer, I want tellme's turn output framed the way tell-me-go frames it — one **frame** per AI-endpoint call, with the closing status **once per call** and the final one trailing the answer.

**Why this priority**: the per-call cadence is what makes the tool log legible "like tell-me-go".

**Independent verification**: run a tool-using turn (≥2 inference rounds); assert the output contains a `╭─⠿ Turn N` frame per AI call (N advancing), a per-call tail (grouped `[Tool Reason]` + measured payload + metrics + `╰─⠿ Ready`), and that the **final** call's tail appears **after** the answer; a tool-less turn shows exactly one frame.

**Acceptance Scenarios**:

1. **Given** a turn that makes more than one AI-endpoint call, **When** it runs, **Then** each call's status **frame** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload) appears at that call's begin, and each call emits its **tail** (grouped reasons + measured payload + metrics + `╰─⠿ Ready`) at its end.
2. **Given** the final AI-endpoint call, **When** it ends, **Then** its tail is emitted **after** the answer bytes (the closing status trails the answer).
3. **Given** a 1-call plain turn, **When** it runs, **Then** exactly one frame renders and the metrics + `Ready` appear **after** the answer.
4. **Given** a prompt-bearing turn, **When** each call's pre-flight payload line is emitted, **Then** its estimate is computed **before** that call and is **non-decreasing** across the turn.

**Functional Requirements**:

- **FR-008**: tellme MUST emit the status **frame** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload line) once per **AI-endpoint call**, at call begin; `N` = the session's prior AI-call count + `k` (the 1-based call index within the turn) — the number **advances within a prompt**. tellme MUST emit the call's **tail** (grouped `[Tool Reason]` lines + measured payload line + metrics line + `╰─⠿ Ready`) at that call's end, with the **final** call's tail emitted **after** the answer bytes (last-call deferral). (G1, G5, G6)
- **FR-009**: the header / pre-flight payload / measured payload / metrics / `Ready` **line formats** MUST stay tellme's existing round-018/027 formats (this round changes only the cadence and the tool-call rendering).
- **FR-010a**: each call's pre-flight payload estimate MUST be computed **before** that call by the **CLI** observer from the call-begin hook's fused `base+turn` wire messages (`llm.EstimatePayload(res.Person, agent.ToolDefs(reg), messages)`); the loop injects **no** estimator closure and carries no persona. Call 1's value MUST be byte-identical to today's estimate, and calls 2..k MUST grow monotonically with the appended turn. (G1, G10)
- **FR-010b**: the persistence/accounting flush MUST stay **one** `AppendBatch` per turn and the persistence gate MUST stay the **final** call's `Reported` flag; the persisted record set is the `Reported` subset of `result.Calls` on a completed turn and **empty** on every error exit. The per-call tail's `Ready` session field is a **recorded display-only divergence**: on a failed turn, or a completed turn whose final call reports no usage, it names session totals that will never be persisted. (G2)

---

### User Story 3 - Live `[Tool Output]` for shell commands (Priority: P2)

As a developer, I want `execute_command`'s output streamed to the terminal as `[Tool Output]` lines the way tell-me-go does, so I get a real-time view (and can CTRL+C a runaway command).

**Why this priority**: the element the operator explicitly called out as missing.

**Independent verification**: run a turn whose scripted provider asks for `execute_command` printing output; assert the `[Tool Output]` header, separators, and the output lines.

**Acceptance Scenarios**:

1. **Given** a tool round that runs a shell-class call **without** `output_file`, **When** it executes, **Then** `stderr` carries `[HH:MM:SS] [Tool Output] Executing... (Output shown below)` (emitted at call begin, before the child starts), a separator line, the command's stdout/stderr **line by line as they arrive**, and a closing separator line.
2. **Given** a command that produces more than the effective byte budget, **When** it runs, **Then** the stream carries the lines up to the budget and then **stops** (the process group is stopped; the result carries the truncation marker) — it is not an unbounded print.
3. **Given** a call that carries `output_file`, **When** it runs, **Then** **no** `[Tool Output]` block renders.
4. **Given** a trailing partial line (no terminating newline before the child exits/stops), **Then** it is **dropped**, not flushed.

**Functional Requirements**:

- **FR-010**: When a round runs a **shell-class call that does not carry `output_file`**, tellme MUST emit, in order, the header `[HH:MM:SS] [Tool Output] Executing... (Output shown below)` (at call begin, before the child starts), a separator line, each **complete** stdout/stderr line as it arrives (prefixed `[HH:MM:SS] [Tool Output] `), and a closing separator line. A shell-class call that carries `output_file` MUST emit **no** block. A trailing partial line MUST be **dropped, not flushed**. (G4)
- **FR-011**: The `[Tool Output]` stream MUST be **bounded by the round-024 resource contract**: it carries every complete line the child produces **up to that call's byte budget**, after which the process group is **stopped** (the existing `trimmed` outcome + `TruncationMarker` in the result); the bound MUST be observable as a **stop, not a cap**. ("Unbounded" is unachievable and MUST NOT be specified.) (G4)
- **FR-012**: The round-019 spinner MUST be **kept** and, for a shell command, **yielded once per call** (single writer): the presenter clears/stops the spinner before the child starts; the sink owns `stderr` exclusively for the whole block (serialising the two copy goroutines with its own mutex); the presenter resumes after the closing separator. The block MUST render **unconditionally**, including when the spinner is gated off (non-TTY / `-r`), where the pause/resume is a no-op. (G6, G8)

**Non-Functional Requirements**:

- **NFR-001**: `[Tool Output]` is a **shell-class** surface only — in tellme, `execute_command`. Readers/writers/`list_skills`/MCP tools MUST NOT emit it.

---

### Edge cases

- **A round with several tool calls** → one `[Tool Engine]` per round; `[Tool Reason]`/`[Tool Action]` interleaved per call; `[Tool Result]` per call; grouped reasons at the post-call tail.
- **A long argument value** → ≤189 runes (188 + U+2026), rune-safe (FR-003).
- **A long textual result** → folded, ≤200 runes (199 + U+2026), rune-safe (FR-004).
- **Arguments that do not parse** → `[Tool Action] <tool>()` (FR-003).
- **A binary result** → the readers' inline text marker `(Binary file, cannot display as text)` on the `[Tool Result]` line (the special branch is struck, FR-004).
- **A command producing unbounded output** → printed up to the byte budget, then **stopped** (FR-011).
- **An `output_file` call** → no `[Tool Output]` block (FR-010).
- **A trailing partial line** → dropped (FR-010).
- **A call whose `reason` is empty/absent** → no `[Tool Reason]` line (defensive; FR-006).
- **A tool-less turn** → exactly one frame, no tool lines, tail after the answer.
- **A bound-reached turn** → the final call renders a **frame but no `[Tool Engine]` line**; the turn fails with `the tool request failed` (exit 7) and persists **nothing** (FR-001, FR-008, FR-010b).
- **A failed turn** → the already-shown per-call `Ready` session totals exceed what `us.Totals()` will report next; the **next prompt's** frame numbering restarts at `prior + 1` (the failed turn persists nothing) — a recorded display/accounting skew (FR-010b).
- **`stdout`/`stderr` interleave inside a `[Tool Output]` block** → non-deterministic (two `io.Copy` goroutines); assertions use **presence + per-stream relative order**, never a global line order.
- **The `-i` submit / piped / positional surfaces** → all carry the rendering; `-r` does not suppress it.
- **Offline paths** → unaffected.

### Key entities

- **Tool-call log block** — the `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` lines for one tool round.
- **Status frame** — the per-AI-call bracket (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload) and its closing **tail** (grouped `[Tool Reason]` + measured payload + metrics + `╰─⠿ Ready`).
- **`[Tool Output]` stream** — the live, byte-bounded, stop-on-overflow per-line stream of an `execute_command` child's stdout/stderr.
- **Tool call** — one model-requested invocation; its top-level `reason` and its `Arguments` are the values rendered.

## Requirements *(mandatory)*

### Global requirements

- **FR-013**: `stdout` MUST stay **byte-exact**; every tool-call and status line MUST be written to `stderr`.
- **FR-014**: The rendering MUST appear on all prompt-bearing surfaces (positional / piped / `-i` submit); `-r` MUST NOT suppress the tool-call lines.
- **FR-015**: The offline paths MUST remain unchanged (no tool-call rendering, no new output).
- **FR-016**: The round MUST NOT change tellme's tool surface, tool schemas (incl. the required `reason`), tool results, the tool resource contract, the provider transport, the persisted `history.jsonl` record shape, the frozen class-phrase vocabulary, or the exit-code set.
- **FR-017**: The reference's own status-block line formats (its cost/timing metrics line, its `Ready` footer without ` - ` groups) MUST NOT be adopted (Q4).

### Non-Functional Requirements

- **NFR-002**: **POSIX-only** (no Windows variant).
- **NFR-003**: **No security/consent gate** (a settled exclusion); the reference's `[Bypassed]` line is **not** reproduced.
- **NFR-004**: **stdlib-only** — no new module dependency (`go.mod` / `go.sum` unchanged).
- **NFR-005**: The rendering MUST be exercisable **hermetically** (offline, no pty, injected streams/clock/estimator).

### Out of scope (recorded)

- The reference's pre-turn `[Info] Starting chat...` line (round-017 divergence).
- The reference's `[Bypassed]` security line.
- The reference's status-block line formats (Q4).
- Reference-shaped binary reporting (`Received <mime> (<n> bytes)`) — a tool-result change (FR-016).

## Success criteria *(mandatory)*

- **SC-001**: A tool-using run's `stderr` carries, per round, `[Tool Engine] Step i/M`, `[Tool Reason]`, `[Tool Action] <tool>(sorted k: v)` (reason excluded, value ≤189 runes), and `[Tool Result] <tool>: <snippet>` (single line, ≤200 runes) — no round-022 `[Tool]` line (US1). Unparseable args → `<tool>()`.
- **SC-002**: The round's reasons reappear, grouped, at the post-call tail before the measured payload line (US1/FR-005).
- **SC-003**: A tool-using turn shows a **frame per AI call** (`╭─⠿ Turn N`, N advancing) with a per-call tail; the **final** tail trails the answer; a tool-less turn shows exactly one frame (US2).
- **SC-004**: `execute_command` (non-`output_file`) streams a live `[Tool Output]` block (header + separators + per-line output), **bounded-and-stopped** at the byte budget; `output_file` emits no block (US3).
- **SC-005**: `stdout` byte-exact; offline paths unchanged; no new dependency; tool surface / class-phrase vocabulary / exit codes unchanged.
- **SC-006**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and each pinned element has a **falsifiability witness** (reverting it fails the corresponding scenario) — including: a 190-rune argument renders 189 runes; rune-boundary cut; sorted keys vs source order; `json.Number` vs `%v`; no engine line on the bound-reached call; the frozen estimator seam; `output_file` block suppression; the stop-vs-cap wording; the per-call frame count; tail-before-answer rejected.

## Assumptions

- **The pins above (G1–G10) are normative and supersede the corresponding earlier text.**
- The `[Tool Output]` header text is `Executing... (Output shown below)` and the separator is the reference's fixed hyphen literal — exact literals pinned in `/axb-dsl-refine`.
- The `[HH:MM:SS]` timestamps reuse tellme's shared `formatClock` seam.
- Every tellme tool call carries `reason`; the renderer stays defensively tolerant of an empty reason (FR-006).
- The emitting seams: a call-begin hook `(callIndex, messages []llm.Message)` and a call-end hook `(callIndex, usage, roundReasons, final)` on the existing `agentport.LoopObserver` (which becomes a composite so the CLI can both render the block and drive the spinner); the CLI computes the estimate from the messages (`llm.EstimatePayload`) and remains the only renderer/accounting owner; `AgentLoop` gains no estimator field and no persona; the `[Tool Output]` sink is a struct-bound, prompt-path-lazily-bound per-line writer on the command tool (round-029 TD-1 / round-033 FR-009 precedent), so `agentTools()` stays parameterless and read-free.
- **Truth scope (explicit MODIFY set)** — `specs/truth/techstack.md` (Agent tool loop · Turn chrome · Post-turn status lines · Turn progress spinner · `execute_command` · Read-only filesystem tools rows); `specs/truth/features/cli/chat/`: `watching-the-tool-loop.feature`, `presenting-the-turn.feature`, `presenting-the-post-turn-status.feature`, `presenting-the-progress-spinner.feature`, `reporting-the-payload-status.feature`, `estimating-the-wire-payload.feature`, `failing-the-tool-loop.feature`, and `chat/dsl.md`. `/axb-api-plan` is **NOOP** (no HTTP surface). `/axb-data-plan` is a **re-derived, checked NOOP** (see `truth-delta.md`). `/axb-system-analysis` records the CLI end carried to `/axb-dsl-refine`.
- **This spec's Assumptions list replaces the earlier three-file list** (the grill found the earlier scope incomplete).
- **Recorded divergences** (all documented above): the decomposed `[Tool Engine]`-family shape (round 022's reversal); the rune-cap vs byte-cut + U+2026 vs ASCII dots; sorted-keys + `json.Number` vs map-iteration + `%v`; the bounded-and-stopped stream vs the reference's un-terminated `warnWriter`; the per-call cadence + final-call deferral + the display/persistence divergence; `output_file` → no block; trailing partial line dropped; the `Step i/M` executed-round unit vs the reference's call index.
- This is a **CLI-interface** round: `/axb-spec-by-example` renders the acceptance journeys; `/axb-dsl-refine` is the CLI end contract owner.
