# Feature Specification: tellme tool-usage accounting — per-tool invocation outcomes and an offline roll-up (round 026)

**Feature Branch**: `026-tool-usage-accounting`

**Created**: 2026-09-15

**Status**: Draft — scope and mechanism **operator-locked** (session 2026-09-15): **Q1 → 1** (classify each invocation as **`ok` / `error` / `timeout`**; a "failure" = **error or timeout**) · **Q2 → Others** (persist to a **global** append-only JSONL log at **`~/.tellme/tools-count.jsonl`**, one record per invocation `{timestamp, tool, outcome}`, shared across repos/envs/modes and **never reset by `--new`**) · **Q3 → 1** (surface via a **dedicated offline reporting flag**, e.g. `--tool-usage`, that prints the roll-up to **stdout** and exits).

**Input**: Follow-on issue [issue #53](https://github.com/gosharplite/tellme/issues/53), verbatim: *"A **measurement** slice: count how often each agent tool is invoked and whether it **passes or fails**, so the operator can see which tools the AI actually uses and which fail."* Motivation from the issue: *"The surface-minimalism thesis ('does this tool beat bash?') needs **data**, not opinion … Reveals tools the AI **never uses** (prune candidates) and tools that **fail often** (fix candidates / signals a bad contract)."*

**Scope note**: a **measurement** round. It adds (a) a per-invocation **outcome classification** in the agent loop, (b) a **global append-only** tool-usage log, and (c) a **new offline reporting flag**. It does **not** change the tool execution semantics (round 024), the tool set (round 024/021), the round-018 token usage store (`tokens.log` / `tokens.summary.json`), the `stdout` of the prompt path (byte-exact), the frozen class-phrase vocabulary, or the post-turn status lines.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Each tool invocation is recorded with its outcome (Priority: P1)

As an operator, I want every agent-tool invocation recorded with its outcome — `ok`, `error`, or `timeout` — in a machine-wide log, so I can see which tools the AI actually calls and how they fare.

**Why this priority**: it is the measurement itself — without the per-invocation record there is nothing to roll up. The issue's whole purpose (evidence for the surface-minimalism thesis) depends on it.

**Independent verification**: run a turn whose fake provider batches several tools — one that succeeds, one whose tool returns an error, and one that is stopped at its time limit — then assert the log gained exactly one record per invocation, each with the correct outcome (`ok` / `error` / `timeout`), and that the prompt path's `stdout` is unchanged.

**Acceptance Scenarios**:

1. **Given** a turn that invokes a tool that returns a result, **When** the invocation completes, **Then** the tool-usage log gains one record for that tool with outcome `ok`.
2. **Given** a turn that invokes a tool that returns an error, **When** the invocation completes, **Then** the log gains one record with outcome `error`.
3. **Given** a turn that invokes a tool that is stopped at its effective time limit, **When** the invocation completes, **Then** the log gains one record with outcome `timeout`.
4. **Given** a turn that invokes several tools, **When** the turn completes, **Then** the log gains exactly one record per invocation, in call order, and the prompt path's `stdout` is byte-identical to a run without accounting.

**Functional Requirements (FR)**:

- **FR-001**: For every agent-tool invocation that the loop actually **executes** (the tool name resolves in the registry and `Execute` runs), the loop MUST record **exactly one** usage record carrying the tool's **wire name** and its **outcome**.
- **FR-002**: The outcome MUST be classified as exactly one of: **`error`** (the tool returned a non-nil error), **`timeout`** (the tool returned a nil-error result **and the loop's per-call deadline was exceeded** — the round-024 FR-018 "stopped at the time limit" path), or **`ok`** (otherwise — including a bounded/truncated result, which is a success). Classification MUST read **only** the loop-owned `err` and the **per-call `ctx` deadline** — never the result text. This rests on a stated **loop↔tool invariant**: *a tool produces a timeout result only when the loop-owned per-call deadline has expired; a tool MUST NOT fabricate a timeout result outside that deadline.* **Tie-break**: on a rare trim-vs-deadline collision (the byte budget and the deadline both fire), the record follows the **loop's deadline signal** (`timeout`) even though the returned result text is a bounded success; the tool's trim-vs-stop guard remains a *result-shape* concern, not the accounting's.
- **FR-003**: The record MUST be appended to a **global, append-only JSON-Lines log** at `~/.tellme/tools-count.jsonl` (the user home resolved via `os.UserHomeDir()`), one JSON object per line: `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"<ok|error|timeout>"}`. The `timestamp` comes from the adapter's `time.Now()` (an explicitly **unasserted** value — no injected-clock determinism is claimed).
- **FR-004**: The write MUST be **best-effort** — a log I/O failure (or an unresolvable user home) MUST NOT break or alter the turn; the answer still prints and the turn still succeeds. The `~/.tellme/` directory and the log file MUST be created **lazily on the first `Record`** (never at construction), so a turn that uses no tool leaves `~/.tellme` untouched.
- **FR-005**: The accounting MUST NOT change the prompt path's `stdout` (byte-exact manual/`-r` output) nor the frozen class-phrase vocabulary; the write MUST NOT emit anything on `stdout` or `stderr`.

### User Story 2 - The operator reviews a per-tool roll-up (Priority: P2)

As an operator, I want an offline command that prints, per tool, how many times it was invoked and how it fared, so I can spot tools the AI never uses (prune candidates) and tools that fail often (fix candidates).

**Why this priority**: the record (US1) is only useful if it can be read back as evidence. It is the operator-facing payoff of the measurement.

**Independent verification**: seed the global log with known records, run the reporting flag, and assert its `stdout` shows a deterministic per-tool roll-up (invocations / ok / error / timeout) that includes every registered tool — including tools with zero recorded use — and that the run contacted no provider and read no stdin.

**Acceptance Scenarios**:

1. **Given** a global log recording several tools' usage, **When** the operator runs the tool-usage report, **Then** tellme prints, for each tool, its invocation count broken down by outcome, and exits successfully without contacting a provider.
2. **Given** a tool that is registered but has never been recorded, **When** the operator runs the report, **Then** that tool appears with zero invocations (so "never used" is visible).
3. **Given** no global log exists, **When** the operator runs the report, **Then** every registered tool is listed with zero invocations and the run exits successfully.

**Functional Requirements (FR)**:

- **FR-006**: tellme MUST offer a **dedicated offline reporting path** — the **`--tool-usage`** flag (pinned by this round's truth) — that prints the per-tool roll-up to `stdout` and exits. The report MUST be **`--version`-class**: it requires **neither** `-c`, **nor** `TELL_ME_HOME` (or a session workspace) — it reads only the **user home** (`os.UserHomeDir()`) and the **tool registry**. Its output is **plain text on `stdout`**, **not TTY-gated**, and **not affected by `-r`**.
- **FR-007**: The report MUST list **every registered agent tool** — enumerated from the **live registry** (the round-024 set of four is today's value, but the report MUST NOT hardcode it, so a future tool addition/removal can never pass vacuously) — and, for each, the total invocations plus the `ok` / `error` / `timeout` breakdown, aggregated over the **entire** global log (not session-scoped).
- **FR-008**: The report MUST be **strictly offline**: it MUST NOT contact a provider and MUST NOT read stdin; it writes its report to `stdout`.
- **FR-009**: The report MUST be **deterministic** — the tool order is fixed (the registry's offer order), independent of the log's line order or history.

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The reporting path's aggregation MUST be a single streaming pass over the log (O(N) time, O(1) memory in the number of tools) — suitable for an unbounded append-only file.

---

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-010**: The tool-usage log MUST be **global and cumulative** — shared across repos, environments, and personas on the machine, and **never reset, archived, or truncated by `--new`** (unlike the per-mode round-018 token log).
- **FR-011**: The round MUST NOT change the round-018 per-mode token usage store (`tokens.log` / `tokens.summary.json` / `tokens.archive.jsonl`), the tool execution semantics, the tool set, the post-turn status lines, or the spinner.
- **FR-012**: The report reader MUST be **resilient** — a malformed or torn line (e.g. a concurrent writer's half-written final line) is **skipped** best-effort; the report MUST never fail, abort, or corrupt the roll-up because of log content. (This is the read-side complement of the write-side best-effort rule, FR-004.)

#### Non-Functional Requirements

- **NFR-002**: The round MUST add **no new third-party module** — the log uses the Go standard library (`encoding/json` over `os`).
- **NFR-003**: The behaviour MUST be **POSIX-only**; the append is a single line to an `O_APPEND|O_CREATE` file, so concurrent writers never rewrite each other's bytes (no `flock` — a settled exclusion).
- **NFR-004**: Verification MUST be **hermetic** — the write path resolves the user home, so tests point the user home at a temporary directory and never touch the operator's real `~/.tellme`.

### Edge cases

- A turn that invokes **no** tool → **no record** is written.
- A turn that invokes the same tool **several times** → **one record per invocation**.
- A tool whose result is **truncated/bounded** (round 024) → outcome **`ok`** (a bounded success is not a failure).
- A tool the model **requests but which is not registered** → the run aborts with the existing incomplete-loop condition (`the tool request failed`); **no usage record** is written (an unavailable name is not a registered tool).
- The **user home cannot be resolved** (unset `HOME`) → the write is a **silent no-op**; the turn is unaffected.
- The **report** is run with **no log yet** → every registered tool is listed with zero invocations; exit success.
- A **write failure mid-turn** → swallowed (best-effort); the turn, the answer, and the exit code are unchanged.
- A **malformed/torn log line** (a concurrent writer's half-written final line) → skipped; the report still succeeds (FR-012).
- `TELL_ME_HOME` is **unset/empty** → the report still succeeds (it is `--version`-class and needs no runtime home) — a normal prompt turn with `TELL_ME_HOME` unset keeps its existing behaviour.

### Key entities

- **Tool-usage record** — one JSON object per line in the global log: `timestamp` (RFC3339), `tool` (the wire name), `outcome` (`ok` | `error` | `timeout`).
- **Tool-usage log** — the global, append-only JSON-Lines file `~/.tellme/tools-count.jsonl`; read whole and summed for the report; never reset by `--new`.
- **Outcome** — the three-way classification of one invocation (`ok` / `error` / `timeout`); `error` and `timeout` are the two "failure" kinds.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: A turn that invokes N tools writes exactly N records to the global log, one per invocation, each carrying the wire name and a correct `ok` / `error` / `timeout` outcome (covers FR-001–FR-003).
- **SC-002**: The reporting flag prints a deterministic per-tool roll-up (invocations / ok / error / timeout) that includes every registered tool — zero-use tools shown with zero invocations — and exits success without a provider request or a stdin read (covers FR-006–FR-009).
- **SC-003**: The prompt path's `stdout` is byte-exact, the class-phrase vocabulary is unchanged, and the round-018 token store / tool semantics / post-turn lines are unchanged (covers FR-005, FR-010, FR-011).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with falsifiability witnesses for the outcome classification and the roll-up.

## Assumptions

- The **counting seam** is the agent loop (round-008 `AgentLoop.Run`): it is the only place that holds both the tool's returned error and the per-call deadline, so `error` / `timeout` are structural signals and no tool-result text is sniffed (Q1 → 1).
- The log is **global per user machine** at `~/.tellme/tools-count.jsonl` — a **recorded divergence** from `tellme`'s `TellMeHome`-is-the-namespace convention (Q2 → Others); it is still local state, modelled by `/axb-data-plan`.
- The **reporting flag is `--tool-usage`** (pinned by this round's truth — `chat/dsl.md` + `techstack.md`); only its output *wording* remains an implementation detail pinned at the step-definition layer.
- The **report's resolution footprint** is `--version`-class: no `-c`, no `TELL_ME_HOME`, no workspace (FR-006). The **record's timestamp** is the adapter's `time.Now()` (unasserted), and the log dir/file are created **lazily on the first `Record`** (FR-003/FR-004).
- **Retention** (compaction/rotation of the unbounded append-only file) is a **forward item**, out of scope this round.
- Only **executed** (registry-resolved) invocations are recorded; an unregistered tool name aborts the run (existing behaviour) and records nothing.
- The report's exact line/table format is an implementation detail pinned at the step-definition layer, subject to the deterministic-order and every-registered-tool requirements.
- This round **adds** new truth (a data-model record + a CLI reporting module/rows); it **modifies in place** `specs/truth/techstack.md`. No frozen plan package is touched (`fresh-package-per-round`).
