# Feature Specification: tellme Agent Tool Surface Parity (round 021)

**Feature Branch**: `021-tool-surface-parity`

**Created**: 2026-09-14

**Status**: Draft — operator-locked decisions D1–D5

**Input**: Operator request "Let these three tools be the theme of slice 21 — remove `summarize_history`; `list_files` and `read_files` must have a `reason` param; their params and functionality will base on `tell-me-go` as close as possible." The round aligns tellme's agent tool surface with the reference (`~/tmp/github/gosharplite/tell-me-go/internal/tools/workspace/registration.go` + `reader.go`) and removes the round-008 summarisation tool.

**Operator-locked decisions (clarify)** — the operator resolved all five scope decisions up front:

- **D1 — `read_files` is multi-file**: adopt the reference signature `filepaths: string[]` (not a single `path`). The round-008 single-`path` contract is rewritten in place.
- **D2 — `reason` is required and echoed**: every filesystem tool requires `reason`; its value is accepted and **echoed into the tool-loop diagnostic (`stderr`) log line**. tellme has **no consent layer** (below), so `reason` is the reference's audit parameter surfaced as an operator-visible tail, not a consent trigger.
- **D3 — reference limits/edge handling verbatim**: per-file cap **100000 bytes** → `... (truncated)`; binary files → `(Binary file, cannot display as text)`; a directory given to `read_files` → `ERROR: path is a directory, use list_files instead`; **≤50** files per call; unreadable/unsafe paths render an **inline `ERROR: …`** result (the loop continues, it is not a tool failure).
- **D3a — bounded results (tellme-specific)**: while the per-file/≤50 limits match the reference, tellme additionally caps **each** reader tool's whole result at **1 MiB** (NFR-001) so the round-008 "a read cannot exhaust the context window" property holds. This is a **deliberate, recorded divergence** — the reference has no aggregate cap and tellme has no in-session overflow recovery (no pruning/summarisation).
- **D4 — no security/consent layer**: `SafePath` / `UserInteractor` consent remain **settled exclusions** (round-008 Clarify Q3). The tools read whatever path the model gives; no `IsPathSafe` gate, no new failure class.
- **D5 — add `get_tree`**: the reference bundles a third reader tool, `get_tree`; this round **adds** it (`{path?, max_depth?, reason*}`), so the surface is exactly the reference's read trio.

> **Scope note**: this is a **tool-contract parity** round, not a new capability round. It removes one tool, reshapes two, and adds one — all inside the existing `AgentLoop`, tool registry, and `stderr` tool-loop logging. **No** provider-transport, session-history, CLI-flag, or exit-code changes.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read many files in one tool call (Priority: P1)

As a developer using tellme from the terminal, I want the model to be able to **read several files in a single tool call** (the reference's `read_files`), so that multi-file work costs fewer provider round trips and matches the reference's behaviour.

**Why this priority**: the multi-file `filepaths` array is the single largest parity gap and the operator's primary "as close as possible" target (D1). It reshapes the most-used reader tool.

**Independent verification**: with a directory holding known files, send a prompt whose scripted provider response asks tellme to read two files; confirm both file contents appear, each preceded by `--- File: <path> ---`, that a >100 KB file is truncated with `... (truncated)`, that a binary file renders the binary marker, that a directory path renders the directory `ERROR:` line, and that >50 paths returns the too-many-files message.

**Acceptance Scenarios**:

1. **Given** a working directory containing `a.txt` and `b.txt`, **When** the model requests `read_files` with `filepaths: ["a.txt","b.txt"]`, **Then** the tool result contains both files' contents, each framed by `--- File: <path> ---`.
2. **Given** a file larger than the cap and a binary file, **When** the model requests both, **Then** the large file is truncated with the `... (truncated)` marker and the binary file prints `(Binary file, cannot display as text)`.
3. **Given** a directory path and more than 50 file paths, **When** the model requests them, **Then** the directory renders `ERROR: path is a directory, use list_files instead` and the 51+ request returns `Error: requested too many files (N). Maximum is 50 files per call.`

**Functional Requirements**:

- **FR-001**: `read_files` MUST accept the argument schema `{ "filepaths": string[] (required), "reason": string (required) }`, replacing the round-008 single `path` argument.
- **FR-002**: The system MUST return each requested file's content prefixed by a `--- File: <path> ---` header line, with a blank line separating file blocks.
- **FR-003**: A file whose content exceeds **100000 bytes** MUST be truncated to that bound and terminated with a `... (truncated)` marker.
- **FR-004**: A binary file MUST render `(Binary file, cannot display as text)`; a directory path MUST render `ERROR: path is a directory, use list_files instead`; an unreadable path MUST render an inline `ERROR: …` result. These render **inside the tool result** (not as a tool/process failure); the loop continues.
- **FR-005**: A request MUST carry between **1 and 50** file paths; an empty/omitted `filepaths` is a **tool argument error** (a *structural* argument, validated at runtime — distinct from the advisory `reason`, FR-012); more than 50 paths MUST return `Error: requested too many files (N). Maximum is 50 files per call.` The **whole** result MUST additionally be capped at **1 MiB**, deterministically: a file block (its `--- File: <path> ---` header plus body) is appended only while it fits within the cap; on the first block that would exceed it, the tool appends `... (truncated at the read budget)` and stops — the omitted file gets **no** header.

**Non-Functional Requirements**:

- **NFR-001**: Every reader tool's result MUST be bounded so it cannot exhaust the model's context window (round-008 NFR-006 lineage): `read_files` by ≤50 files × ≤100000 bytes **per file** *and* an aggregate cap of **1 MiB** for the whole result; `list_files` and `get_tree` by a **1 MiB** result cap. Truncation markers: `read_files` uses `... (truncated)` per file and `... (truncated at the read budget)` for its aggregate cap; `list_files` and `get_tree` append `... (truncated)`.

---

### User Story 2 - List a directory the reference way (Priority: P1)

As a developer, I want `list_files` to return directory entries in the reference's shape, so results are consistent with `tell-me-go` and the model/prompt can rely on the same format.

**Why this priority**: `list_files` is the companion reader tool and its output format currently diverges from the reference (D5/D3 parity).

**Independent verification**: point `list_files` at a directory containing a file and a sub-directory; confirm the result starts with `Contents of <path>:` and lists `[f] name` / `[d] name`; call it with no `path` and confirm it lists the current directory.

**Acceptance Scenarios**:

1. **Given** a directory with a file and a sub-directory, **When** the model requests `list_files`, **Then** the result begins `Contents of <path>:` and each entry is prefixed `[f] ` or `[d] `.
2. **Given** a request with no `path`, **When** the model requests `list_files`, **Then** the current directory (`.`) is listed.

**Functional Requirements**:

- **FR-006**: `list_files` MUST accept `{ "path": string (optional), "reason": string (required) }`; an omitted/empty `path` MUST default to the current directory (`"."`).
- **FR-007**: The result MUST begin with `Contents of <path>:` and list one entry per line prefixed `[d] ` for directories and `[f] ` for files.

---

### User Story 3 - View a directory tree (Priority: P2)

As a developer, I want a `get_tree` tool that renders a visual directory tree, so the model can survey a subtree in one call like the reference.

**Why this priority**: the operator chose to add the reference's third reader tool (D5) so the surface matches exactly.

**Independent verification**: point `get_tree` at a small nested directory; confirm a connector tree (`├── `/`└── `) is rendered, that recursion stops at the requested depth (default 2), and that `.git` is not recursed into.

**Acceptance Scenarios**:

1. **Given** a nested directory, **When** the model requests `get_tree`, **Then** the result is a connector tree of entries.
2. **Given** a `get_tree` request with `max_depth`, **When** the model requests it, **Then** entries deeper than `max_depth` are not listed; `.git` is not recursed into.

**Functional Requirements**:

- **FR-008**: `get_tree` MUST accept `{ "path": string (optional), "max_depth": integer (optional), "reason": string (required) }`; an omitted `path` MUST default to `"."`; an omitted/≤0 `max_depth` MUST default to **2**.
- **FR-009**: The result MUST be a connector tree using `├── ` / `└── ` prefixes and a continuation indent; directories MUST be recursed into **except `.git`**; recursion MUST stop at `max_depth`.

---

### User Story 4 - Offer only the reference's reader tools (Priority: P2)

As an operator, I want tellme's tool surface to be exactly the reference's read trio — `list_files`, `read_files`, `get_tree` — so tellme does not offer a summarisation tool the operator removed.

**Why this priority**: the operator's explicit first instruction is to **remove `summarize_history`**; a smaller, single-purpose surface is the goal.

**Independent verification**: run a prompt-bearing turn; confirm the provider request declares exactly three tools (`list_files`, `read_files`, `get_tree`) and that no path offers or invokes `summarize_history`; confirm the removed tool's truth feature / DSL rows no longer exist.

**Acceptance Scenarios**:

1. **Given** any prompt-bearing run, **When** tellme offers tools to the provider, **Then** the offered tool set is exactly `list_files`, `read_files`, `get_tree`.
2. **Given** a provider that asks for `summarize_history`, **When** the loop runs, **Then** no such tool is registered (an unknown tool is terminal, per the round-008 contract).

**Functional Requirements**:

- **FR-010**: The tool registry MUST offer exactly three tools for a prompt run — `list_files`, `read_files`, `get_tree` — and MUST NOT register `summarize_history`; the summarisation adapter MUST be removed.
- **FR-011**: No user-facing behaviour, truth feature, or DSL row MUST reference the summarisation tool after this round; the tool-loop behaviour is otherwise unchanged.

---

### Edge cases

- **`read_files` with empty/omitted `filepaths`** → tool argument error (FR-005).
- **`read_files` with >50 paths** → inline `Error: requested too many files (N). …` result (FR-005).
- **File exactly at the cap** → not truncated; **one byte over** → truncated with the marker (FR-003).
- **Unreadable / missing file among several** → that block renders `ERROR: …`; the other files still render (FR-004).
- **Directory passed to `read_files`** → the directory `ERROR:` line, not a crash (FR-004).
- **`list_files` on an empty directory** → `Contents of <path>:` with no entry lines (FR-007).
- **`get_tree` on a directory containing `.git`** → `.git` is listed but not recursed into (FR-009).
- **`get_tree` with `max_depth` huge / a deep tree** → recursion is still depth-bounded (FR-009).
- **`reason` missing** → the tool argument schema is violated (required); a missing `reason` is an argument error (FR-012).
- **A path outside the working directory** → read normally (no boundary; D4).

### Key entities

- **Agent tool** — one model-invocable capability: now exactly `list_files`, `read_files`, `get_tree`. Each carries a wire name, a description, a JSON-schema (including the required `reason`), and read-only local behaviour.
- **Tool registry** — the ordered set offered to the model; order = offer order.
- **Tool step** — the persisted record of one tool execution inside a completed turn (`{tool, arguments, result[, signature]}`); its `arguments` now carry `filepaths[]` for `read_files`.
- **Tool-loop log line** — the per-tool diagnostic written to `stderr` during the run; now carries the echoed `reason`.

### Global requirements

- **FR-012**: Every filesystem tool (`list_files`, `read_files`, `get_tree`) MUST declare `reason` as a **required** argument in its tool schema, and the system MUST echo that value into the tool-loop diagnostic (`stderr`) log line for the call. The tools do **not** otherwise validate `reason` at runtime (a **schema-only** requirement, matching the reference); a missing `reason` is therefore a schema matter and not a distinct runtime failure class. (Deliberate asymmetry: `filepaths` — FR-005 — is a *structural* argument validated at runtime; `reason` is *advisory* and schema-only.)
- **FR-013**: The filesystem tools MUST have **no** path/safety boundary — no `SafePath` check, no consent prompt, no new failure class (settled exclusion, D4).
- **FR-014**: The tool-loop bounds and failure contract MUST be unchanged — `MAX_TOOL_LOOP` (default 1000, env/config) and the frozen class phrase `tellme: the tool request failed` with exit code `7`.
- **FR-015**: A completed tool-using turn MUST persist its steps as today (`{tool, arguments, result[, signature]}`), with `read_files` arguments now carrying `filepaths[]`.
- **FR-016**: `stdout` MUST stay byte-exact (the answer stream); the tool-loop log lines and the echoed `reason` MUST go to `stderr`.

### Success criteria

- **SC-001**: The provider request for a prompt-bearing run declares exactly `list_files`, `read_files`, `get_tree` (three tools; no `summarize_history`).
- **SC-002**: A multi-file `read_files` returns each file framed by `--- File: <path> ---`, with the 100 KB per-file truncation, the binary marker, the directory `ERROR:`, and the ≤50-file behaviours observable — **except** the **1 MiB aggregate cap**, which is **unit-observable** (T032 / witness #6; no interface Example exercises an aggregate truncation).
- **SC-003**: `list_files` returns `Contents of <path>:` with `[d]`/`[f]` prefixes and defaults `path` to `.` when omitted.
- **SC-004**: `get_tree` returns a connector tree honouring a default `max_depth` of 2 and skipping `.git` recursion.
- **SC-005**: Every filesystem tool call records its `reason` in the `stderr` tool-loop log and no `stdout` byte changes.
- **SC-006**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and a falsifiability witness per reshaped/new tool fails when the behaviour is removed.

### Assumptions

- **No security/consent layer** is introduced (D4); the round-008 `no boundary` record stands.
- **Provider tool-name schema is unchanged** (`^[a-zA-Z0-9_-]{1,64}$`); the three names are already wire-valid.
- **No new dependency**; the tools remain stdlib-only and local (no network, no `process execution`).
- **POSIX + the existing hermetic harness** (fake provider + working-dir runner) is sufficient; no pty is needed.
- **The reference numbers are the target**: `maxReadSize = 100000`, `maxFilesPerCall = 50`, default `max_depth = 2`, default `path = "."`.
- The prior single-`path` `read_files` contract and the `summarize_history` tool are **round-008 history** and are rewritten/removed in place (per `fresh-package-per-round`, the round-008 package stays frozen).
