# Feature Specification: tellme Tool Resource Contract, `execute_command` & Reader Retrofit (round 024)

**Feature Branch**: `024-tool-resource-contract-and-execute-command`

**Created**: 2026-09-15

**Status**: Draft — operator-locked decisions D1–D7 + clarify Q1–Q3 (session 2026-09-15); **folded after the PR #54 grill round** — Q1 (byte-budget wording) · Q4 (timeout-as-result, FR-005a narrowed, FR-018 added) (session 2026-09-15)

**Input**: Operator direction: *"`tellme` will target bash shell with no security overhead … no security and no windows"*; *"rewrite `tell-me-go` with better methods (bdd/sdd/tdd)"*; *"`token bound` and `timeout` should both be params of agent tools … the current 3 read tools are limited by 1 Mb; this should be a param which can be altered by AI … same goes to timeout"*; *"024 `execute_command` is such an important tool, we should get it correct right away"*; *"Let's have 024 do `execute_command` and the 3 read tools together."*

**Operator-locked decisions (session 2026-09-15)**:

- **D1 — no security layer.** No `SafePath` check, consent prompt, command whitelist, or forbidden-character gate. The destructive-command risk is an **explicitly accepted** decision. (Lineage: round 008 Clarify Q3, round 021 D4.)
- **D2 — no Windows.** POSIX/bash only.
- **D3 — bash-first.** `execute_command` runs through `bash -c` and is a first-class primitive; **`pipe_commands` is not offered** (bash already pipes).
- **D4 — small surface.** A dedicated tool must beat bash on boundedness, determinism, or reliability.
- **D5 — tool resource contract.** A token bound and a timeout are **uniform tool parameters** with three tiers — **default (protective) + param (AI tunes) + ceiling (hard clamp)** — enforced centrally.
- **D6 — one aggregate bound.** `read_files` uses a single aggregate bound: **no** per-file cap, **no** fair-share math, **no** paging; a file larger than any safe bound is read in slices via the shell.
- **D7 — scope.** This round = the tool resource contract + `execute_command` + the reader retrofit (`list_files`, `read_files`, `get_tree`).

> **Scope note**: a **capability** round for the CLI end's agent tool surface. It adds `execute_command`, introduces the shared token-bound/timeout contract, and retrofits the three readers. It does **not** change the provider transport, the session-history/usage records, the CLI flag surface, the exit-code contract, or any `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run a shell command (`execute_command`) (Priority: P1)

As a developer using `tellme` from a terminal, I want the model to be able to **run shell commands**, so that it can build, test, inspect, and act on the workspace — the same reasoning-agent loop the reference provides.

**Why this priority**: it is the single largest capability gap — with only read tools the agent cannot act. A bash-first `execute_command` is the highest-leverage next tool and the first consumer of the resource contract.

**Independent verification**: script a provider response that requests `execute_command`; confirm the command runs via `bash -c`, its output is returned to the model bounded to the effective token bound, a long-running command is stopped at its timeout, `reason` is echoed on the diagnostic stream, and `stdout` is unchanged.

**Acceptance Scenarios**:

1. **Given** a working directory, **When** the model requests `execute_command` with `echo hi`, **Then** the tool result carries the command's output and exit status, and the loop continues.
2. **Given** a command whose output exceeds the bound, **When** it runs, **Then** the tool result is truncated to the effective `max_output_tokens` with a truncation marker.
3. **Given** a command that never returns, **When** the tool's `timeout` elapses, **Then** the process tree is terminated and a timeout result is surfaced, and the loop continues.
4. **Given** a multi-stage need, **When** the model writes `a | b` in the `command` string, **Then** it runs via `bash -c` (no separate `pipe_commands` tool is offered).
5. **Given** a command whose output is large, **When** the model passes `output_file`, **Then** the output is written to that file rather than returned in full, the result reports the capture target, and a bounded `read_files` can read it back.

**Functional Requirements (FR)**:

- **FR-001**: `execute_command` MUST accept `{ "command": string (required), "timeout": number (optional), "max_output_tokens": integer (optional), "output_file": string (optional), "append": boolean (optional), "reason": string (required) }` and run the command through `bash -c`.
- **FR-002**: The tool result MUST be bounded to the effective `max_output_tokens` (global contract), with an explicit truncation marker when cut.
- **FR-003**: A command that exceeds the effective `timeout` MUST be terminated (**its process tree** — see FR-018) and surfaced as a **timeout result** — a **nil-error** tool result carrying a fixed "stopped at its time limit" marker (never the loop's `error: ` path); the run MUST continue (the loop is not failed).
- **FR-004**: `reason` MUST be a required schema field and MUST be echoed into the tool-loop diagnostic (`stderr`) line for the call.
- **FR-005**: `execute_command` MUST NOT sit behind any security/consent gate, and `pipe_commands` MUST NOT be registered (D1/D3).
- **FR-006**: The command's output is a **tool result fed to the model**; the child's `stdout`/`stderr` MUST be bound to the tool's own buffers — **never** inherited from tellme's `os.Stdout`/`os.Stderr` — so `execute_command` never alters the operator `stdout` (byte-exact; NFR-002).
- **FR-007**: `execute_command` MUST run on POSIX only (no Windows branch).

- **FR-005a — non-zero-exit semantics (resolved, clarify Q1→1)**: a command that exits non-zero MUST yield a **successful tool result carrying the exit code** (with its bounded output); the loop MUST continue. A **tool failure** — a **non-nil tool error** — is reserved for a **tool/argument error** only (e.g. a missing `command`): **neither** a normal non-zero exit **nor** a timeout (FR-003/FR-018) is a failure.
- **FR-005b — output capture (clarify Q2→1)**: `execute_command` MUST accept an optional `output_file` (string) and `append` (boolean). When `output_file` is set, the child's output — **both `stdout` and `stderr`** — MUST be bound to that file handle (streamed to disk — **never** routed through the tool's memory; `append: true` appends, otherwise it truncates), and the tool result MUST report the exit status and the capture target **with no inline preview** (a preview is obtained by a subsequent bounded `read_files`). **Recorded:** because both streams are redirected, a failing command with `output_file` set surfaces only the exit status + target (no inline error text). `output_file` is the sanctioned **large-output escape hatch**: capture to a file, then read it back in bounded slices (`read_files` / the shell). No path gate applies (D1).

---

### User Story 2 - Read files and directories without a hard per-file cap (Priority: P2)

As a developer, I want the reader tools to honour a **single context-budget bound** the model can set, instead of a fixed 1 MiB / 100000-byte ceiling — so a large file is not silently capped when there is budget to read it, and the cap always tracks the model's actual window.

**Why this priority**: it fixes the round-021 fixed-constant problem — the cap cannot scale with the model (issue #49) and blocks reading files larger than 100000 bytes — and completes the contract across the existing surface.

**Independent verification**: point `read_files` at a file larger than 100000 bytes with a raised `max_output_tokens` and confirm more of the file is returned (no 100000 hard cap); point it at many files and confirm the aggregate bound, the skip marker, and the ≤50 cap behave; confirm `get_tree`/`list_files` honour the same bound and that `timeout` applies.

**Acceptance Scenarios**:

1. **Given** a file larger than 100000 bytes and a `max_output_tokens` that accommodates it, **When** the model requests `read_files`, **Then** the file is returned without the 100000-byte hard cap (bounded only by the aggregate).
2. **Given** a multi-file request whose total exceeds the aggregate bound, **When** the model requests `read_files`, **Then** the result contains the files that fit plus an explicit marker naming what was skipped — no per-file arithmetic.
3. **Given** a single file larger than the bound, **When** the model requests `read_files`, **Then** the file is truncated at the bound with a marker hinting that slices can be read via the shell.
4. **Given** `list_files` / `get_tree` over a large directory, **When** the result exceeds the bound, **Then** it is truncated with the marker.

**Functional Requirements (FR)**:

- **FR-008**: `list_files`, `read_files`, and `get_tree` MUST accept the optional `max_output_tokens` and `timeout` parameters (defaulting per the global contract).
- **FR-009**: `read_files` MUST read the requested files **whole** (no 100000-byte per-file cap), in request order, **stopping at the aggregate `max_output_tokens` bound**; on overflow it MUST emit a marker stating what was read and what was skipped (an omitted file gets no header).
- **FR-010**: For a single file larger than the bound, `read_files` MUST truncate at the bound and MUST include a marker hinting that slices can be read via the shell.
- **FR-011**: `list_files` and `get_tree` results MUST be bounded by the same `max_output_tokens` (truncation marker on overflow).
- **FR-012**: `read_files` MUST keep the **≤50 files per call** cap (degenerate-input guard); the `filepaths` argument remains required and non-empty.

---

### Edge cases

- `execute_command` without `reason` → schema violation (argument error).
- `execute_command` `timeout` above the ceiling → clamped to the ceiling (no error).
- `read_files` `["big.txt","small.txt"]` where `big.txt` exhausts the budget → `big.txt` truncated at the bound, `small.txt` not read, the marker names it.
- `read_files` with empty/omitted `filepaths` → tool argument error (unchanged).
- `read_files` with more than 50 paths → the too-many-files message (unchanged).
- A reader file exactly at the bound → not truncated; one byte over → truncated with the marker.
- `max_output_tokens` above the ceiling → clamped to the ceiling.
- A directory passed to `read_files` → the directory `ERROR:` line (unchanged).

### Key entities

- **Agent tool** — a model-invocable capability; now `list_files`, `read_files`, `get_tree`, and **`execute_command`**.
- **Tool resource contract** — the uniform `max_output_tokens` + `timeout` parameters (default/param/ceiling) enforced centrally for every agent tool.
- **Effective budget / bound / ceiling** — `effectiveBudget` = `min(configured MAX_HISTORY_TOKENS, the active model's context window)`; the per-call result bound (param or `effectiveBudget ÷ 4`) and its hard clamp (`effectiveBudget ÷ 2`).
- **Command result** — the `execute_command` tool result: bounded output plus exit status (semantics per FR-005a).
- **Timeout result** — a **nil-error** tool result carrying a fixed "stopped at its time limit" marker, emitted when any tool observes its effective `timeout` (FR-003/FR-018); distinct from a truncation (the byte-budget cut) and from a tool failure (a non-nil error).
- **Tool step** — the persisted record of a tool execution; `execute_command` adds `{tool:"execute_command", arguments, result}`.

---

## Requirements *(mandatory)*

### Global requirements

> The per-story FRs live under their stories; this section holds the **cross-story tool resource contract**.

#### Functional Requirements

- **FR-013**: **Every** agent tool (`list_files`, `read_files`, `get_tree`, `execute_command`) MUST accept an optional `max_output_tokens` (integer) and an optional `timeout` (number, seconds).
- **FR-014**: Enforcement MUST be **centralised** — a single execution path applies the timeout (`context.WithTimeout`) and clamps/truncates the result to the effective bound. A param above the ceiling MUST be **clamped, never rejected**.
- **FR-015**: The tool bound MUST derive from the **effective budget** = `min(configured MAX_HISTORY_TOKENS, the active model's configured context window)`, where the window is an optional per-model `MODELS.<model>.CONTEXT_WINDOW` (absent → the configured `MAX_HISTORY_TOKENS`). The **default** `max_output_tokens` MUST be `effectiveBudget ÷ 4` (retiring the fixed 1 MiB; the cap tracks the effective budget — issue #49), and the **ceiling** MUST be `effectiveBudget ÷ 2` (a headroom reservation so one tool result cannot fill the window). The resolved token bound MUST be realised as a single **byte budget** via one **contract-owned** conversion — `byteBudget = resolvedTokenBound × bytesPerToken`, with `bytesPerToken = 4` a **new, separately-named contract constant** (deliberately **not** the token estimator's heuristic ratio, whose "not a contract term" stance is unchanged) — so a tool bounds its output **at the source** and the loop's backstop clamps on **raw byte length** (`len(result) > byteBudget`), **never** on the estimator (which would double-trim via its `+4` per-message overhead). At the shipped default (`MAX_HISTORY_TOKENS = 1000000`, no window) the default **byte** budget is `1000000 B` (≈ 0.95 MiB), replacing the round-021 fixed `1 MiB` (`1048576 B`).
- **FR-016**: The `timeout` default MUST be per-tool (**`execute_command` 300 s; the readers 30 s**) with a hard **ceiling of 7200 s**; a param above the ceiling is clamped.
- **FR-017**: No new failure class MUST be introduced and the frozen class-phrase vocabulary MUST stay unchanged (11); the tool-loop contract (`MAX_TOOL_LOOP`, the frozen phrase, exit codes) is otherwise unchanged (a non-zero exit is a success result — FR-005a).
- **FR-018 (timeout is a result — uniform across every agent tool)**: A tool that observes its effective `timeout` MUST yield a **nil-error tool result** carrying a fixed "stopped at its time limit" marker — **never** the loop's `error: ` path — on **both** `execute_command` (FR-003, a process-group kill) **and** the readers (FR-016's 30 s default; their current `ctx.Err()` → error path is retired). A returned **non-nil tool error** therefore means a genuine **tool/argument error** (FR-005a). A tool honours the deadline only when it **observes** it (a blocked `io.ReadAll` notices it when the read returns). The frozen class-phrase vocabulary is unchanged (FR-017).

#### Non-Functional Requirements

- **NFR-001**: Every agent tool's result MUST be bounded so it cannot exhaust the model's context window; the bound MUST track the **effective budget** (the configured `MAX_HISTORY_TOKENS`, capped by the model's configured context window) and be clampable to a headroom-reserved ceiling.
- **NFR-002**: `stdout` MUST stay byte-exact; all tool diagnostics (the `reason` echo, timeout notices, truncation markers) are tool-result content or `stderr` only.
- **NFR-003**: No new dependency; the tools remain stdlib-only and local (no network); POSIX/bash only.

---

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: A prompt-bearing run offers exactly `list_files`, `read_files`, `get_tree`, `execute_command` (four tools; no `pipe_commands`, no security tooling).
- **SC-002**: `execute_command` runs via `bash -c`; output is bounded with a marker; a long command is stopped at the timeout; `reason` is echoed; `stdout` is byte-identical to a run without the tool.
- **SC-003**: The readers honour `max_output_tokens` + `timeout`; a file larger than 100000 bytes is no longer hard-capped; the aggregate bound, the skip marker, and the ≤50 cap behave.
- **SC-004**: A default derived from the **effective budget** — `min(configured MAX_HISTORY_TOKENS, the active model's configured context window)` — replaces the fixed 1 MiB constant (the cap is tied to the resolved budget: issue #49); an over-ceiling param is clamped.
- **SC-005**: A non-zero exit yields a successful result carrying the exit code (loop continues); `output_file`/`append` capture output to a file, readable back by a bounded `read_files`.
- **SC-006**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with falsifiability witnesses for each new/changed behaviour.

---

## Assumptions

- **Numbers (per `research.md` D5).** The default `max_output_tokens` is `effectiveBudget ÷ 4` and the ceiling `effectiveBudget ÷ 2`; the shell `timeout` default is **300 s** (aligning with the agent loop's existing per-tool timeout), the readers' **30 s**, ceiling **7200 s**. (The reference's ~15 s figure is a starting point, not the chosen value.) The token bound is realised as a **byte** budget via the contract-owned `bytesPerToken = 4` conversion — default **`1000000 B`** at the shipped 1 M budget (FR-015).
- **No security, no Windows, bash-first** (D1–D3) are inherited from `README.md` → *Design Intent & Direction*; the destructive-command risk is an accepted decision.
- **`output_file`/`append` on `execute_command`** are **in** this round (clarify Q2→1); the sanctioned "large output" path is capture-to-file then a bounded `read_files`.
- This round **rewrites in place** the round-021 fixed caps in `specs/truth/features/cli/chat/**` and `chat/dsl.md`; the round-021 plan package stays frozen (`fresh-package-per-round`).
- No pty is needed; the existing hermetic harness (fake provider + working-directory runner) drives the E2E.
