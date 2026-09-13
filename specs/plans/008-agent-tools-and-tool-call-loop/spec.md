# Feature Specification: tellme Agent Tools & the Tool-Call Loop (round 008)

**Feature Branch**: `008-agent-tools-and-tool-call-loop`

**Created**: 2026-09-12

**Status**: Draft — Clarify Rounds 1–3 resolved

**Input**: User request "008 — agent tools / the tool-call loop". This round gives tellme an **agentic** capability: a prompt run may **call declared tools and iterate** (think → act → observe) until a final answer, instead of a single answer-only provider round trip. Round 007 deliberately excluded tool calls ("a turn carries only prompt/answer content"); the deferred **history summarisation** capability lands with this round **as an agent tool**, while **token-budget pruning** remains a **settled exclusion** ([#22](https://github.com/gosharplite/tellme/issues/22)).

**Clarify Round 1 (2026-09-12)** resolved three high-impact scope decisions:

- **Q1 -> Option 2 (read-only filesystem tools)**: the first slice ships exactly **two read-only local tools** — **`list_files`** and **`read_files`** — plus Story 4's **session-summarisation** tool `summarize_history`. No writes, no process execution, no network beyond the existing provider turn.
- **Q2 -> Option 2 (widen the persisted turn)**: a completed tool-using turn persists its tool calls/results — a data-truth **MODIFY** to `history_entry` (owned by `/axb-data-plan`). This reverses round 007's "tool calls are not in scope" record shape.
- **Q3 -> Option 1 (no boundary)**: the read tools have **no** path/safety boundary; `SafePath` and interactive consent remain **settled exclusions** (not re-opened).

**Clarify Round 2 (2026-09-12)** resolved two items: **(Q1)** the loop's terminal failure contract — a tool-using run that cannot complete reports a **new dedicated frozen class phrase** `tellme: the tool request failed` and a **new exit code** `7` (vocabulary 10→11); **(Q2)** the loop bound — `MAX_TOOL_LOOP`, default **1000**, overridable via env/config.

**Clarify Round 3 (2026-09-12, user-invited)** resolved operator visibility: tool-loop activity is **surfaced live during the run** (the operator sees tool-loop logs as it works), written to the **diagnostic stream (`stderr`)** so `stdout` stays the answer stream, while **`-l N` continues to list only user/assistant messages** (the round-007 contract is unchanged). The loop is a **sequence of discrete provider calls + tool executions**, not token-level streaming.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Answer with a declared tool (Priority: P1)

As a developer using tellme from the terminal, I want tellme to **use a declared tool when the prompt requires it** and fold the tool's result into its answer, so that the answer reflects real data instead of the model's guess.

**Why this priority**: This is the core value of the round — without at least one working tool invocation, there is no agent, and Stories 2–4 are meaningless.

**Independent verification**: start tellme in a directory with known files; send a prompt that can only be answered by listing and reading them; confirm the tools were executed and the final answer reflects their results. Separately, send a prompt that needs no tool and confirm tellme still answers in **one** provider round trip with no tool invocation.

**Acceptance Scenarios**:

1. **Given** a directory containing known files and a prompt that requires reading them, **When** the operator runs tellme, **Then** tellme invokes the read/list tools and produces a final answer that reflects their results.
2. **Given** a prompt that needs no tool, **When** the operator runs tellme, **Then** tellme answers in a single provider round trip and invokes no tool.

**Functional Requirements**:

- **FR-001**: The system MUST register exactly **three wire-valid tools** for a prompt run — two **read-only filesystem** tools, **`list_files`** (enumerate a directory's entries) and **`read_files`** (return a file's contents), and the **session-summarisation** tool **`summarize_history`** (Story 4 / FR-012) — and make them available to the model. Tool identifiers MUST be valid per the provider tool-name schema (`^[a-zA-Z0-9_-]{1,64}$`; no whitespace).
- **FR-002**: On a prompt run, the system MUST offer the available tool definitions to the model and allow the model to request a tool instead of (or in addition to) answering.
- **FR-003**: The system MUST execute a requested tool and return its result to the model for a subsequent request.
- **FR-004**: The system MUST keep the think→act→observe cycle going until the model produces a final answer or a bound is reached (Story 3).
- **FR-005**: A prompt that requires no tool MUST remain a single provider request (no regression of rounds 004–007 behaviour).

**Non-Functional Requirements**:

- **NFR-001**: The read tools MUST be **read-only** and MUST NOT mutate the filesystem or spawn processes.

---

### User Story 2 - Watch the tool loop work (Priority: P2)

As an operator, I want to **see the tool loop's activity while it is running** — which tools are called, with what arguments, and what they return — so that I can follow and trust what an autonomous agent is doing, and diagnose a bad loop as it happens.

**Why this priority**: Observability is what makes an autonomous loop usable from a terminal; it depends on Story 1's loop existing. (The user explicitly requires seeing tool-loop logs while the run is in progress.)

**Independent verification**: run a prompt that triggers several tool iterations and confirm the run prints a tool-loop log line per tool call (and its result) **as it happens**, before the final answer; then run a no-tool prompt and confirm no tool-loop lines appear.

**Acceptance Scenarios**:

1. **Given** a prompt that triggers tool calls, **When** the run executes the loop, **Then** tellme prints a tool-loop log line for each tool call **and** its result as the loop proceeds, before the final answer.
2. **Given** a prompt that needs no tool, **When** the run executes, **Then** tellme prints no tool-loop log lines.

**Functional Requirements**:

- **FR-006**: The system MUST surface the tool loop's activity to the operator **during the run** — for each tool call: the tool name, its arguments, and its result (or error) — written to the **diagnostic stream (`stderr`)**, keeping `stdout` the answer stream. It is a sequence of **discrete log lines** emitted at each loop step (after each provider response and each tool execution), **not** a token-level stream.
- **FR-007**: The tool-loop log MUST NOT alter the final-answer output contract: the answer is still printed as today (rendered by default, raw under `-r`).

**Non-Functional Requirements**:

- **NFR-002**: Tool-loop log lines MUST be emitted **as they happen** (not buffered to the end) and in a deterministic order.

---

### User Story 3 - A bounded, terminating loop with a clear failure report (Priority: P3)

As an operator, I want a tool-using run to **always terminate deterministically** and to fail understandably when a tool errors, a tool is unknown, a tool exceeds its time budget, or the iteration bound is reached, so that a stuck or failing agent never hangs the CLI or leaves me guessing.

**Why this priority**: It is what makes the loop safe to run unattended; it depends on Story 1's loop existing.

**Independent verification**: force each terminating condition (a tool that always errors, an unknown tool, a tool that exceeds its budget, a model that keeps requesting tools) and confirm the run ends with the documented outcome and a clear message, never a hang.

**Acceptance Scenarios**:

1. **Given** a tool that returns an error, **When** the operator runs a prompt that triggers it, **Then** tellme terminates the run deterministically and reports the outcome with the frozen failure class for the loop.
2. **Given** a model that keeps requesting tools past `MAX_TOOL_LOOP`, **When** the operator runs the prompt, **Then** tellme stops at the bound and reports the failure per the failure contract instead of looping forever.
3. **Given** a single tool execution that exceeds its time budget, **When** it is requested, **Then** tellme treats it as a tool failure (per scenario 1) and the run still terminates.

**Functional Requirements**:

- **FR-008**: The system MUST bound the number of tool iterations within a single prompt run via **`MAX_TOOL_LOOP`** (default **1000**), overridable via environment variable and configuration, and MUST stop deterministically when the bound is reached.
- **FR-009**: The system MUST bound each individual tool execution by a time budget and treat a timeout as a tool failure.
- **FR-010**: A tool-using run that cannot complete (iteration bound reached, or an unrecoverable tool/protocol failure) MUST report a deterministic, operator-facing outcome using the new dedicated frozen class phrase `tellme: the tool request failed` and the new exit code `7`. A tool that merely returns an error is fed back to the model as its result and is **not** terminal (Story 1).
- **FR-011**: The system MUST NOT hang: every tool-using run terminates.

**Non-Functional Requirements**:

- **NFR-003**: The loop's failure outcomes MUST be deterministic and covered by fast, isolated tests, complementing the black-box acceptance path.

---

### User Story 4 - History summarisation as an agent tool (Priority: P4)

As a user with a long conversation, I want tellme to be able to **summarise earlier history through an agent tool** so that a long session can be condensed on demand, while the persisted history itself is not silently rewritten.

**Why this priority**: It is the deferred round-007 capability ([#22](https://github.com/gosharplite/tellme/issues/22)) and it demonstrates an **LLM-backed** tool alongside the local read tools of Story 1; it depends on the tool infrastructure of Story 1.

**Independent verification**: with a multi-turn history, trigger the summarisation tool and confirm it produces a summary that the model can use, and confirm the persisted history file is unchanged except for the recorded turn.

**Acceptance Scenarios**:

1. **Given** a session with persisted history, **When** the summarisation agent tool is invoked, **Then** tellme returns a summary of the earlier conversation that the model uses in subsequent reasoning.
2. **Given** the summarisation tool has run, **When** the operator inspects the persisted history, **Then** the earlier records are **not** silently rewritten (any change is only the newly completed turn).

**Functional Requirements**:

- **FR-012**: The system MUST provide history summarisation **as an agent tool** named **`summarize_history`** — an LLM-backed tool that reads the persisted conversation and returns a condensed summary as its tool result — distinct from automatic context pruning.
- **FR-013**: Invoking the summarisation tool MUST NOT silently rewrite persisted history records.
- **FR-014**: The summarisation tool MUST be subject to the same loop bounds and failure contract as Story 3.

**Non-Functional Requirements**:

- **NFR-004**: The summarisation tool MUST NOT introduce token-budget pruning; it is an on-demand tool, not an automatic pruning policy. (Settled exclusion, [#22](https://github.com/gosharplite/tellme/issues/22).)

---

### Edge Cases

- When a requested tool is **not declared**, the system MUST treat it as a deterministic failure (fail fast), not a silent no-op.
- When **`read_files`** targets a missing or unreadable path, the system MUST treat it as a tool result (error), not crash the run.
- When **`read_files`** targets a **large file**, the returned content MUST be **size-bounded** so it cannot exhaust the model's context window — token-budget pruning is out of scope, so the read tool must not inject unbounded input.
- When **`list_files`** targets a non-directory path, the system MUST return a deterministic error result.
- When the model returns **no tool request**, the run MUST behave exactly as a single-answer turn (Story 1, scenario 2) and print no tool-loop logs (Story 2, scenario 2).
- When a prompt-bearing run **fails mid-loop**, the system MUST NOT write a partial/interrupted history record (round-007 append-after-complete is preserved at the turn level).
- When the model requests **multiple tools in one response**, the system MUST handle the batch deterministically (sequential by default this round unless `/axb-technical-research` decides otherwise).
- The offline commands (`--version`, `-d`, `-l`, prompt-less boot) MUST remain strictly offline; only a prompt-bearing run (the chat path) may contact a provider.

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-015**: The system MUST keep its existing flag surface (`-c/--config`, `-d/--diagnostics`, `--version`, `-r/--raw`, `--new`, `-l N`) and MUST NOT regress the acceptance behaviour of rounds 001–007, except where this round deliberately extends the chat turn (a recorded truth MODIFY under `/axb-dsl-refine`).
- **FR-016**: A completed tool-using turn MUST persist its tool calls/results by **widening** the session-history record (`history_entry`), while remaining **append-after-complete** (the turn is written only when it completes).
- **FR-017**: The `-l N` command MUST continue to list **only user/assistant messages**; the widened tool activity is **not** surfaced by `-l`.
- **FR-018**: The read tools MUST NOT introduce a path/safety boundary; `SafePath` and interactive consent remain **settled exclusions**.

#### Non-Functional Requirements

- **NFR-005**: The round MUST remain **stdlib-first**: introduce a new dependency only if a tool genuinely requires it, and record the reason.
- **NFR-006**: The loop's iteration count, tool outcomes, and the wire requests MUST be observable by tests (the local fake provider records what was sent) so the loop is falsifiable.
- **NFR-007**: Tool results fed back into the conversation MUST be **size-bounded** so a single tool result cannot exhaust the model's context window (token-budget pruning is out of scope this round).

### Key Entities *(include if feature involves data)*

- **Tool declaration**: the model-facing description of a capability — a name, a description, and a parameter schema. The first slice declares `list_files`, `read_files`, and `summarize_history` (all wire-valid snake_case).
- **Tool request**: the model's request to run a tool with concrete arguments.
- **Tool result**: the outcome of one tool execution (content, or an error/timeout), size-bounded.
- **Turn (extended)**: one prompt run, which may now span **multiple** provider requests interleaved with tool executions; persisted as before, but now carrying its tool activity.
- **History entry (widened)**: the persisted record of a completed turn, extended to embed the turn's tool calls/results (round-007 shape MODIFY).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of prompts that require the read/list tools execute them and produce a final answer that reflects their results.
- **SC-002**: In the acceptance set, 100% of tool-using runs print a tool-loop log line per tool call and result **as it happens**, before the final answer; no-tool runs print none.
- **SC-003**: In the acceptance set, 100% of tool-using runs terminate deterministically within `MAX_TOOL_LOOP` (no hang).
- **SC-004**: In the acceptance set, 100% of prompts that need no tool remain a single provider request (round-004–007 behaviour preserved).
- **SC-005**: In the acceptance set, 100% of tool-using failures (unknown tool, timeout, bound reached, unrecoverable error) report `the tool request failed` and exit `7`.
- **SC-006**: All pre-existing round-001–007 acceptance scenarios remain green, except where a recorded truth MODIFY updates them for the now-tool-capable turn; `-l N` remains user/assistant-only.
- **SC-007**: The history-summarisation agent tool produces a usable summary and leaves earlier persisted records unrewritten.
- **SC-008**: 100% of `read_files` results are size-bounded, so no single read can exhaust the assembled context.

## Assumptions

- **Tool set (Q1)**: exactly two read-only filesystem tools, `list_files` and `read_files`, plus the LLM-backed session-summarisation tool `summarize_history` (Story 4); no writes, no process execution, no network beyond the provider turn.
- **Persistence (Q2)**: a completed tool-using turn's tool calls/results are persisted via a widened `history_entry` (data-truth MODIFY, `/axb-data-plan`); `-l` still lists only user/assistant messages.
- **Safety (Q3)**: no path/safety boundary; `SafePath` and interactive consent remain settled exclusions.
- **Visibility (Round 3)**: the tool loop is observable live during the run; the log is written to the diagnostic stream (`stderr`), so `stdout` remains the answer stream. The loop is a **sequence of discrete provider calls + tool executions**, **not** token-level streaming (streaming responses remain a settled exclusion).
- The **first provider family** remains the OpenAI-compatible family (round-004 Clarify Q2); structured tool calls for Anthropic/Gemini are **out of scope** for this round.
- A **turn** remains one user prompt + one final provider answer from the operator's point of view; internally it may span several provider requests.
- Tool results (file contents) are **size-bounded** by a fixed cap (a determination for `/axb-technical-research`), so the out-of-scope status of token-budget pruning is not violated.
- Batched tool calls in one model response are handled **sequentially** this round unless `/axb-technical-research` decides otherwise (a disclosed, non-blocking default).
- The round keeps the single-turn execution model (round 004), stdin piping (round 005), rendered/raw output (round 006), and durable history (round 007). It does **not** add streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, or a TUI.
