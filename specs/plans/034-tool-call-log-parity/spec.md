# Feature Specification: tool-call log parity — the reference's per-call tool rendering (round 034)

**Feature Branch**: `034-tool-call-log-parity`

**Created**: 2026-09-16

**Status**: Draft — operator-locked decisions Q1–Q7 (clarify held one decision at a time)

**Input**: Operator request: *"I want tellme to show tool calls similar to tell-me-go."* The operator supplied the reference (`tell-me-go`) and tellme renderings side by side.

- **tellme today** writes **one** line per tool call — `[HH:MM:SS] [Tool] <name> - <reason>` (round 022), deliberately omitting the call's arguments and its result.
- **tell-me-go** decomposes each call into `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` lines and emits a full status block **per AI-endpoint call**.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → 1 — full per-AI-call parity.** tellme emits the status block once per **AI-endpoint call** (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload + … + post-call + `╰─⠿ Ready`), with the decomposed tool-call rendering inside. `[Tool Reason]` is printed **twice**: once immediately after `[Tool Engine]`, and once **grouped at the post-call status** (the round's reasons, re-emitted after the results, before the measured payload line) — the reference's actual mechanism.
- **Q2 → 1 — `[Tool Output]` is a live stream.** `execute_command` streams each complete stdout/stderr line as it arrives (the reference's `warnWriter` behavior).
- **Q3 → 1 — verbatim reference templates & constants.** Engine `Step i/M`; Reason; Action `key: value` (`reason` **excluded**, each value capped at **189** chars + `…`); Output; Result `Text` (newlines → spaces, capped at **200** chars + `…`). **All tellme tool calls carry `reason`**, so `[Tool Reason]` always renders — round 022's "states no reason" form **retires**.
- **Q4 → 1 — keep tellme's status-block line formats.** The header / payload / metrics / `Ready` lines keep tellme's round-018/027 formats; **only** the cadence and the tool-call rendering change.
- **Q5 → 1 — unbounded live stream.** `[Tool Output]` prints every line with **no cap** (the operator's deliberate CTRL+C signal that the command is producing a lot).
- **Q6 → 1 — keep the round-019 spinner**, drawn inside the `[Tool Output]` block, yielding the line to output then redrawing.
- **Q7 → 1 — all prompt surfaces** (positional / piped / `-i` submit); `-r` does **not** suppress the tool-call lines (they are `stderr` diagnostics).

**Scope note**: a **tool-call diagnostic rendering** round, not a capability round. It reshapes what tellme writes to `stderr` during a tool-using turn and how the turn's status blocks are laid out. It changes **no** tool schema, tool result, provider transport, session-history record, CLI flag, exit code, or `stdout` byte. It **supersedes** round 022's single-line `[Tool]` log and its blank-line rule, and re-cuts rounds 017/023/027's turn-chrome cadence to **per AI call**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The decomposed tool-call log (Priority: P1)

As a developer reading tellme's terminal, I want each tool call shown the way tell-me-go shows it — a step counter, the reason, the action (tool + arguments), and the result — so I can see *what* the agent ran and *with what arguments*, not just the tool name.

**Why this priority**: it is the operator's primary request and the substance of the feature; without it nothing else in the round has value.

**Independent verification**: run a tool-using turn (scripted provider asks for a tool with `reason` + args); assert `stderr` carries, in order, `[HH:MM:SS] [Tool Engine] Step i/M`, `[HH:MM:SS] [Tool Reason] <reason>`, `[HH:MM:SS] [Tool Action] <tool>(k: v)`, `[HH:MM:SS] [Tool Result] <tool>: <snippet>`, and — at the post-call status — a grouped `[HH:MM:SS] [Tool Reason] <reason>`; assert no round-022 `[Tool] <name> - <reason>` line remains.

**Acceptance Scenarios**:

1. **Given** a tool-using turn, **When** the loop runs a tool round, **Then** `stderr` carries `[HH:MM:SS] [Tool Engine] Step <i>/<M>` (i = 1-based round index, M = the effective loop bound) and, per call, `[Tool Reason]` then `[Tool Action] <tool>(<args>)`.
2. **Given** a call whose arguments carry a top-level `reason`, **When** it is logged, **Then** the `reason` renders as its own `[Tool Reason]` line and is **excluded** from the `[Tool Action]` argument list.
3. **Given** an argument value longer than 189 characters, **When** the action is rendered, **Then** that value is truncated to 189 chars followed by `…`.
4. **Given** a completed call with a textual result, **When** the result is rendered, **Then** `stderr` carries `[HH:MM:SS] [Tool Result] <tool>: <snippet>`, the `Text` folded to a single line (newlines → spaces) and truncated to 200 chars followed by `…`.
5. **Given** a round with several tool calls, **When** the round completes, **Then** each call's `[Tool Reason]` precedes its `[Tool Action]` (interleaved per call), and **all** of the round's reasons are re-emitted, grouped, as `[Tool Reason]` lines at the post-call status (the second occurrence), before the measured payload line.

**Functional Requirements**:

- **FR-001**: The tool-call diagnostic on `stderr` MUST begin each round with `[HH:MM:SS] [Tool Engine] Step <i>/<M>`, where `i` is the 1-based index of the round within the call and `M` is the effective tool-loop bound (`MAX_TOOL_LOOP`, default `1000`).
- **FR-002**: For each call in the round, tellme MUST emit `[HH:MM:SS] [Tool Reason] <reason>` immediately before the call's `[Tool Action]` line, when the call's top-level `reason` is a non-empty string.
- **FR-003**: For each call, tellme MUST emit `[HH:MM:SS] [Tool Action] <tool>(<args>)`, where `<args>` is the call's arguments rendered as `key: value`, **excluding the `reason` key**, joined by `, `, each value truncated to 189 characters followed by `…` when longer.
- **FR-004**: For each completed call, tellme MUST emit `[HH:MM:SS] [Tool Result] <tool>: <snippet>`, where `<snippet>` is the result's text folded to a single line (newlines replaced by spaces) and truncated to 200 characters followed by `…` when longer; a binary result MUST instead render `[Tool Result] <tool>: Received <mime> (<n> bytes)`.
- **FR-005**: At the post-call status, tellme MUST re-emit the round's reasons, grouped, as `[HH:MM:SS] [Tool Reason] <reason>` lines — after the round's results and immediately before the measured payload line (the second `[Tool Reason]` occurrence).
- **FR-006**: Every tool call carries a top-level `reason`; the log therefore always renders `[Tool Reason]`. A call whose `reason` is empty/absent MUST render **no** `[Tool Reason]` line (defensive tolerance) — the round-022 "reports a call that states no reason" behavior is **retired** from the truth.
- **FR-007**: The lines MUST be emitted one per event, in call order per round, each prefixed with a `[HH:MM:SS]` (24-hour, zero-padded) wall-clock timestamp from the shared clock seam.

---

### User Story 2 - A status block per AI-endpoint call (Priority: P1)

As a developer, I want tellme's turn output framed the way tell-me-go frames it — one status block **per AI-endpoint call** — so the tool activity of each inference round sits inside its own block, with its own payload/metrics/`Ready` summary.

**Why this priority**: the per-call cadence is what makes the tool log legible "like tell-me-go"; it is bracket to US1's content and is required for the reference reading.

**Independent verification**: run a tool-using turn (≥2 inference rounds); assert the output contains more than one `╭─⠿ Turn N - <mode>` header and a matching `╰─⠿ Ready` per AI call, and that a tool-less turn still shows exactly one block.

**Acceptance Scenarios**:

1. **Given** a turn that makes more than one AI-endpoint call, **When** it runs, **Then** each call's block (rule + `╭─⠿ Turn N - <mode>` + pre-flight payload + post-call + `╰─⠿ Ready`) appears around that call's tool activity.
2. **Given** a turn that makes exactly one AI-endpoint call (a plain answer), **When** it runs, **Then** exactly one block appears — unchanged from today.

**Functional Requirements**:

- **FR-008**: The status block (rule + `╭─⠿ Turn N - <mode>` header + pre-flight payload line + post-call reasons/measured payload/metrics + `╰─⠿ Ready`) MUST be emitted once per **AI-endpoint call**, not once per prompt.
- **FR-009**: The header/payload/metrics/`Ready` **line formats** MUST stay tellme's existing round-018/027 formats (this round changes only the cadence and the tool-call rendering).

---

### User Story 3 - Live `[Tool Output]` for shell commands (Priority: P2)

As a developer, I want `execute_command`'s live output streamed to the terminal as `[Tool Output]` lines the way tell-me-go does, so I get a clear, real-time signal (and can CTRL+C a runaway command).

**Why this priority**: it is the element the operator explicitly called out as missing; it builds on US1's block and is the most novel sub-capability. It ranks below US1/US2 because it concerns one tool class, but it is a first-class requirement.

**Independent verification**: run a turn whose scripted provider asks for `execute_command` with a benign command that prints output; assert `stderr` carries the `[Tool Output]` header, separators, and the command's output lines.

**Acceptance Scenarios**:

1. **Given** a tool round that runs `execute_command`, **When** the command executes, **Then** `stderr` carries `[HH:MM:SS] [Tool Output] Executing... (Output shown below)`, a hyphen separator line, the command's stdout/stderr **line by line as they arrive**, and a closing hyphen separator line.
2. **Given** a command producing a large amount of output, **When** it runs, **Then** every line is printed (no cap) — the deliberate CTRL+C signal.
3. **Given** a shell tool streaming output while the round-019 spinner is active, **When** an output line arrives, **Then** the spinner yields the line, the output is printed, and the spinner redraws.

**Functional Requirements**:

- **FR-010**: When a round runs `execute_command`, tellme MUST emit `[HH:MM:SS] [Tool Output] Executing... (Output shown below)`, then a separator line, then each complete stdout/stderr line as it arrives (prefixed `[HH:MM:SS] [Tool Output] `), then a closing separator line.
- **FR-011**: The `[Tool Output]` stream MUST be **unbounded** (no cap) and MUST NOT alter the bounded tool result.
- **FR-012**: The round-019 live spinner MUST be **kept** and drawn inside the `[Tool Output]` block; it yields the line to an output line and redraws thereafter.

**Non-Functional Requirements**:

- **NFR-001**: `[Tool Output]` is a **shell-class** surface only — in tellme, `execute_command`. Readers/writers/`list_skills`/MCP tools MUST NOT emit it (they render only via `[Tool Result]`).

---

### Edge cases

- **A round with several tool calls** → one `[Tool Engine]` per round; `[Tool Reason]`/`[Tool Action]` interleaved per call; `[Tool Result]` per call; the round's reasons re-emitted grouped at the post-call status (FR-005).
- **A long argument value** → capped at 189 chars + `…` (FR-003).
- **A long textual result** → folded single-line, capped at 200 chars + `…` (FR-004).
- **A binary result** → `[Tool Result] <tool>: Received <mime> (<n> bytes)` (FR-004).
- **A command producing unbounded output** → printed in full (FR-011).
- **A call whose `reason` is empty/absent** → no `[Tool Reason]` line (defensive; not expected — all calls carry `reason`, FR-006).
- **A tool-less turn** → exactly one status block, no tool lines (FR-008).
- **The `-i` submit / piped / positional surfaces** → all carry the rendering; `-r` does not suppress it (Q7).
- **Offline paths** (`--version`, `-d`, `--tool-usage`, prompt-less `--new`, boot) → unaffected (FR-015).

### Key entities

- **Tool-call log block** — the `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]` lines tellme writes to `stderr` for one tool round.
- **Status block** — the per-AI-call frame (rule + header + pre-flight payload + post-call + `Ready`) that contains the tool-call log block and the answer.
- **`[Tool Output]` stream** — the live, unbounded per-line stream of an `execute_command` child's stdout/stderr.
- **Tool call** — one model-requested invocation; its top-level `reason` string and its arguments JSON are the values rendered on the log lines.

## Requirements *(mandatory)*

### Global requirements

#### Functional Requirements

- **FR-013**: The answer stream `stdout` MUST stay **byte-exact**; every tool-call and status line MUST be written to the diagnostic stream (`stderr`).
- **FR-014**: The rendering MUST appear on all prompt-bearing surfaces (positional / piped / `-i` submit); `-r` MUST NOT suppress the tool-call lines.
- **FR-015**: The offline paths MUST remain unchanged (no tool-call rendering, no new output).
- **FR-016**: The round MUST NOT change tellme's tool surface, tool schemas (including the required `reason` argument), tool results, the tool resource contract, the provider transport, the persisted history record, the frozen class-phrase vocabulary, or the exit-code set.

#### Non-Functional Requirements

- **NFR-002**: tellme MUST remain **POSIX-only** (no Windows variant).
- **NFR-003**: There MUST be **no security/consent gate** (a settled exclusion); the reference's `[Bypassed]` line is **not** reproduced.
- **NFR-004**: The round MUST be **stdlib-only** — no new module dependency (`go.mod` / `go.sum` unchanged).
- **NFR-005**: The rendering MUST be exercisable **hermetically** (offline, no pty, injected streams/clock), consistent with the existing harness.

### Out of scope (recorded)

- The reference's pre-turn `[Info] Starting chat...` line **stays out of scope** (round-017 recorded divergence).
- The reference's **`[Bypassed]`** security line is **not** reproduced (no security layer).
- The reference's **status-block line formats** (its metrics line with `($cost) [timing]`, its `Ready` footer without ` - ` groups) are **not** adopted (Q4).

## Success criteria *(mandatory)*

- **SC-001**: A tool-using run's `stderr` carries, per round, `[Tool Engine] Step i/M`, `[Tool Reason]`, `[Tool Action] <tool>(<args>)` (reason excluded, values capped at 189), and `[Tool Result] <tool>: <snippet>` (single line, capped at 200) — and no round-022 `[Tool] <name> - <reason>` line (US1).
- **SC-002**: The round's reasons reappear, grouped, at the post-call status (the second `[Tool Reason]`), before the measured payload line (US1/FR-005).
- **SC-003**: A tool-using turn shows one status block per AI-endpoint call (multiple headers + `Ready` footers); a tool-less turn shows exactly one (US2).
- **SC-004**: `execute_command` streams a live, unbounded `[Tool Output]` block (header + separators + per-line output) with the spinner drawn inside it (US3).
- **SC-005**: `stdout` is byte-identical to before; the offline paths are unchanged; no new dependency; the tool surface, class-phrase vocabulary, and exit codes are unchanged.
- **SC-006**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and a falsifiability witness (reverting a rendered element) fails the corresponding scenario.

## Assumptions

- The reference's line templates and constants are adopted **verbatim** (Q3): `[Tool Engine] Step i/M`; `[Tool Reason]`; `[Tool Action] <tool>(k: v, …)` (`reason` excluded; `%v`; values capped at 189 + `…`); `[Tool Result] <tool>: <snippet>` (`Text`; newlines → spaces; capped at 200 + `…`; binary → `Received <mime> (<n> bytes)`).
- The `[Tool Output]` header text is `Executing... (Output shown below)` and the separator is the reference's fixed hyphen literal — exact literals pinned in `/axb-dsl-refine`.
- The `[HH:MM:SS]` timestamps reuse tellme's shared `formatClock` seam (round 022).
- tellme's status-block line formats remain unchanged (Q4); this round is a rendering/cadence change only.
- Every tellme tool call carries `reason` (tellme's tools declare it required, round 021 D2 onward); the renderer stays defensively tolerant of an empty reason (FR-006).
- The round touches `specs/truth/techstack.md` (the tool-loop log row), `specs/truth/features/cli/chat/watching-the-tool-loop.feature` + `chat/dsl.md` (MODIFY); `/axb-api-plan` is **NOOP**; `/axb-data-plan` is **NOOP** (the rendering is not persisted state); `/axb-system-analysis` records the CLI end carried to `/axb-dsl-refine`.
- **Recorded divergence**: this round **removes** the round-022 operator-chosen divergence (single-line log, no args/result) and restores the reference's decomposed rendering — an operator-directed reversal, recorded rather than silent.
