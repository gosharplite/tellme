# Feature Specification: tellme Tool-Loop Log Line Reshape (round 022)

**Feature Branch**: `022-tool-loop-log-line`

**Created**: 2026-09-15

**Status**: Draft — operator-locked decisions Q1–Q3

**Input**: Operator request: "I want below tool log and an empty line before the final message" — reshape the per-tool diagnostic line tellme writes to `stderr` during a tool-using run into a single timestamped line `[HH:MM:SS] [Tool] <tool name> - <reason>` (dropping the raw `arguments=` / `result=` dumps the current line carries), and emit one empty line between the tool-log block and the final answer.

**Operator-locked decisions (clarify)** — the operator resolved all three scope decisions up front:

- **Q1 — strict scope**: this round changes **only** (a) the tool-loop `stderr` log line shape and (b) the blank line before the final answer of a tool-using turn. The pre-flight/post-turn **payload line** (rounds 009/018) and the round-019 **spinner labels** are **out of scope** and unchanged.
- **Q2 — blank line only on tool-using turns**: exactly one blank line is emitted between the last `[Tool] …` line and the answer **only when ≥1 tool log line was written**. A non-tool turn (a plain question/answer) keeps today's spacing — no blank line is added.
- **Q3 — missing `reason` renders `[Tool] <name>`**: when a tool call carries no top-level `reason`, the line drops the separator entirely (`[HH:MM:SS] [Tool] <name>`); there is no dangling ` - ` and no placeholder text.

> **Scope note**: this is a **tool-loop diagnostic rendering** round, not a capability round. It reshapes one `stderr` line (emitted by `internal/agent/agentloop.go`'s `logStep`) and adds one blank line for tool-using turns. It changes **no** tool schema, tool result, provider transport, session-history record, CLI flag, exit code, or `stdout` byte. The round-021 generic `reason` extraction (a top-level `reason` string in a call's arguments JSON) is reused unchanged.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A single, timestamped tool-log line (Priority: P1)

As a developer reading tellme's terminal output, I want each tool call logged as **one timestamped line naming the tool and its reason** (`[HH:MM:SS] [Tool] <name> - <reason>`), so I can follow what the agent did at a glance — without the raw argument/result dumps that currently re-print the whole tool result into the diagnostics.

**Why this priority**: it is the operator's primary request and the most-visible formatting defect (today the line carries `arguments=<json>` and a truncated `result=<…>`, which duplicates the tool output — e.g. the repo tree renders twice).

**Independent verification**: run a tool-using turn (scripted provider asks for a tool with a `reason`); assert `stderr` carries one line matching `[HH:MM:SS] [Tool] <name> - <reason>` and that no `arguments=` / `result=` text appears; then run a call with no top-level `reason` and assert the line is `[HH:MM:SS] [Tool] <name>`.

**Acceptance Scenarios**:

1. **Given** a tool-using turn where the model requests a tool with a `reason`, **When** the loop logs the call, **Then** `stderr` carries the line `[HH:MM:SS] [Tool] <name> - <reason>` and contains no `arguments=` or `result=` segment.
2. **Given** a tool call whose arguments carry no top-level `reason`, **When** the loop logs it, **Then** the line is `[HH:MM:SS] [Tool] <name>` (no trailing ` - `).
3. **Given** two tool calls in one turn, **When** the loop logs them, **Then** two lines appear in call order, each with its own timestamp.

**Functional Requirements**:

- **FR-001**: The tool-loop diagnostic line written to `stderr` MUST be `[HH:MM:SS] [Tool] <tool name> - <reason>` for a call carrying a top-level `reason`.
- **FR-002**: The line MUST NOT contain the call's arguments or its result — the current `arguments=<json>` and `result=<…>` segments MUST be removed from the operator log.
- **FR-003**: When the call has no top-level `reason`, the line MUST be `[HH:MM:SS] [Tool] <tool name>` — the ` - <reason>` separator is omitted (no dangling separator, no placeholder).
- **FR-004**: The line MUST begin with a `[HH:MM:SS]` wall-clock timestamp (24-hour, zero-padded).
- **FR-005**: Exactly one line MUST be emitted per tool call, in call order, and the `reason` text MUST be folded to a single line (newlines collapsed) so the log line stays single-line.

---

### User Story 2 - A blank line separating the tool log from the answer (Priority: P2)

As a developer, I want an empty line between the tool-log block and the final answer, so the answer is visually separated from the diagnostics that precede it.

**Why this priority**: secondary readability polish on the same surface; it makes the boundary between "what the agent did" and "what it answered" obvious.

**Independent verification**: run a tool-using turn and confirm exactly one blank line sits between the last `[Tool] …` line and the answer; run a non-tool turn and confirm no blank line is introduced by this change.

**Acceptance Scenarios**:

1. **Given** a tool-using turn (≥1 tool log line), **When** the run finishes and the answer is written, **Then** exactly one blank line separates the last `[Tool] …` line from the answer.
2. **Given** a non-tool turn (no tool calls), **When** the run finishes, **Then** no additional blank line is introduced by this change.

**Functional Requirements**:

- **FR-006**: On a **tool-using** turn (at least one tool log line was written), exactly one blank line MUST appear between the last tool-log line and the final answer.
- **FR-007**: On a **non-tool** turn, this change MUST NOT introduce any additional blank line.

---

### Edge cases

- **Tool call with a reason** → `[HH:MM:SS] [Tool] <name> - <reason>` (FR-001).
- **Tool call without a reason** (a non-filesystem/other tool, or malformed arguments) → `[HH:MM:SS] [Tool] <name>` (FR-003).
- **Multiple tool calls in one turn** → one line each in call order (FR-005); the blank line follows the **last** line (FR-006).
- **A `reason` containing newlines** → folded to a single line (FR-005).
- **Non-tool turn** → no tool-log line and no added blank line (FR-007).
- **Streams split** (`stdout` vs `stderr` redirected separately) → the tool-log lines and the blank line are on `stderr`; `stdout` is unchanged (FR-008).

### Key entities

- **Tool-loop log line** — the per-call diagnostic tellme writes to `stderr` during a tool-using run; now `[HH:MM:SS] [Tool] <tool name> - <reason>` (or without the reason segment).
- **Tool call** — one model-requested invocation (`{name, arguments[, signature]}`); its **top-level `reason`** string is the value echoed onto the log line.
- **Diagnostic stream (`stderr`)** — the stream that carries the tool-log line(s) and the separating blank line; the answer stays on `stdout`.

### Global requirements

- **FR-008**: The answer stream `stdout` MUST stay byte-exact; the tool-log line and the separating blank line MUST be written to the diagnostic stream (`stderr`).
- **FR-009**: The tool-loop bounds and failure contract MUST be unchanged — `MAX_TOOL_LOOP` (default 1000, env/config) and the frozen class phrase `tellme: the tool request failed` with exit code `7`.
- **FR-010**: The tool surface MUST be unchanged — `list_files`, `read_files`, `get_tree`, their required `reason` schema argument, and their results are untouched; this round changes only the operator-visible **log rendering** of a call.
- **FR-011**: The frozen class-phrase vocabulary MUST stay unchanged (11) and no new failure class MUST be introduced.
- **FR-012**: The `reason` extraction MUST remain generic — a top-level `reason` string in the call's arguments JSON (round-021 contract), so any tool carrying `reason` is echoed and any tool without one omits the segment.

### Success criteria

- **SC-001**: A tool-using run's `stderr` carries, per tool call, the line `[HH:MM:SS] [Tool] <name> - <reason>` with no `arguments=` / `result=` text.
- **SC-002**: A tool call with no top-level `reason` renders `[HH:MM:SS] [Tool] <name>`.
- **SC-003**: A tool-using turn has exactly one blank line between the last tool-log line and the answer; a non-tool turn has no blank line added.
- **SC-004**: `stdout` is byte-identical to before the change for the same run.
- **SC-005**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and a falsifiability witness (reverting the line shape, or omitting the blank line) fails the corresponding scenario.

### Assumptions

- The `[HH:MM:SS]` timestamp reuses the **same clock/format as the pre-flight payload line** (local wall-clock, captured at emission time); the loop is given that clock (an injection detail resolved in `/axb-technical-research`).
- The blank line is emitted on `stderr` as the tail of the tool-log block (an emission-site/stream detail resolved in `/axb-technical-research`); `stdout` remains byte-exact.
- `reason` remains a **required** schema argument on the three filesystem tools (round 021 D2 unchanged); the log stays **generic** for tools without one (FR-012).
- No new dependency; POSIX + the existing hermetic harness (fake provider + working-directory runner) is sufficient; no pty is needed.
- The pre-flight/post-turn **payload line** and the round-019 **spinner labels** are **unchanged** (Q1 strict scope).
- The round-008/021 tool-loop log contract (`[tool] …` on `stderr`) is **round history** and is rewritten in place (per `fresh-package-per-round`, the prior packages stay frozen).
