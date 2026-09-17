# Feature Specification: terminal-safe `[Tool …]` lines + turn-output blank-line grouping (round 039)

**Feature Branch**: `039-terminal-safe-lines-and-turn-spacing`

**Created**: 2026-09-17

**Status**: Draft — two folded workstreams. Operator-locked decisions Q1–Q8 (see below).

**Input**: Two workstreams folded into one round at the operator's request:

- **Issue [#80](https://github.com/gosharplite/tellme/issues/80)** — *"Terminal-safe line policy: sanitize every `[Tool …]` stderr formatter, not just `[Tool Output]`"*. Round 038 (PR [#79](https://github.com/gosharplite/tellme/pull/79)) closed the control-sequence leak class on **one** `stderr` surface — the live `[Tool Output]` block. Its siblings render text of the **same provenance class** (model-authored or externally-sourced) and remain unsanitized, so the underlying class — *model-authored or externally-sourced text can tint / alter the operator's terminal* — is still reachable: `FormatToolReason` (model `reason`), `FormatToolResult` (tool-result snippet), and `FormatToolAction`'s argument values.
- **Operator request (no anchor issue)** — *add blank lines to the live turn output*. Today a tool-using turn's `stderr` block is a dense wall of lines; the operator wants it grouped: a blank before each call's begin-block, a blank before the trailing grouped `[Tool Reason]` block, and a blank before the post-status group (`Payload:` + metrics + `Ready`).

## Operator-locked decisions (clarify — resolved one at a time with the operator)

### Workstream A — issue #80 (terminal-safe lines)

- **Q1 → Generalize the control-sequence class to every `[Tool …]` formatter.** The class removed is exactly round 038's: every **7-bit** ESC-introduced sequence (CSI/SGR, OSC, any other ESC sequence) and every stray C0/DEL byte; TAB kept; interior CR dropped; bytes `≥ 0x80` untouched; the ESC consumption is ASCII-gated and the scan window bounded per kind, so the sanitizer never *introduces* invalid UTF-8. `FormatToolReason`, `FormatToolResult`, and `FormatToolAction` (keys **and** values) all become control-free, like `FormatToolOutputLine`.
- **Q2 → Single-owned home.** The sanitizer (+ helpers) moves out of `tooloutput.go` into one dedicated `internal/ui` home (e.g. `sanitize.go`) and is consumed by **all four** `[Tool …]` formatters — exactly one definition of the class (the single-ownership theme also tracked on [#69](https://github.com/gosharplite/tellme/issues/69)).
- **Q3 → Sanitize *before* the rune cap** (after the round-036 fold+trim), so the cap bounds the **visible** output and the round-036 one-line contract is unchanged.
- **Q4 → The neutral-close restore stays `[Tool Output]`-scoped.** No per-line reset is added to the single-line formatters (a reset line per `[Tool Reason]`/`[Tool Result]`/`[Tool Action]` would be noise, and those formatters never leave a multi-line block). Recorded by **superseding ADR 0007** with the new **ADR 0008** (research D9 — not amended in place; an `Accepted` ADR is immutable but for its `Status` line).

### Workstream B — blank-line grouping (operator-requested)

- **Q5 → Blank before each call's begin-block.** One blank line precedes the **first** line of the call-begin block — the `[Tool Reason]` line when the call states a reason, else the `[Tool Action]` line (Q6). This is **per call**, so a round with `k` tool calls shows `k` blanks in its begin sequence (the blank before the round's first reason is just the first instance — no separate round-level blank).
- **Q6 → A reason-less call still separates.** When a call states **no** reason (blank reasons are suppressed since round 036), the blank precedes its `[Tool Action]` line — i.e. the blank is tied to the call block, not to the reason line's presence.
- **Q7 → One blank before the grouped tail block, none inside it.** The trailing re-emitted `[Tool Reason]` block gets exactly **one** blank line before it and **no** blanks between its lines.
- **Q8 → One blank before the post-status group.** The grouped measured `Payload:` line + metrics line + `Ready` footer get exactly **one** blank line before them (emitted only when the group is written — the usage-reporting gate).

**Scope note**: two independent `stderr`-presentation changes (a generalization of an existing sanitize policy; an additive spacing policy). Neither changes a CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, any persisted record, the tool **result** fed to the model, or a `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - No tool-loop line can leak terminal control data (Priority: P1)

As an operator, I want **every** `[Tool …]` diagnostic line — the reason, the result, and the action's arguments — to be free of terminal control sequences, so that neither a model-authored `reason` nor an escape-bearing file/argument can tint or alter my terminal.

**Why this priority**: it is issue #80, the durable home of the round-038 scope boundary; it closes the same class on the three remaining surfaces.

**Independent verification**: unit-pin `FormatToolReason`, `FormatToolResult`, and `FormatToolAction` over a hostile fixture (an ESC-bearing reason, an ESC-bearing result snippet, an escape-bearing key and value); assert each output is control-free and valid UTF-8, and that the visible text + the round-036 cap are preserved. Script an E2E run whose `reason`/result carry an escape and assert the emitted `stderr` lines are control-free.

**Acceptance Scenarios**:

1. **Given** a tool-using turn whose model-authored `reason` contains a colour escape, **When** the run reports the reason, **Then** the `[Tool Reason]` line carries **no** terminal control sequence and the visible reason text is preserved.
2. **Given** a tool whose result snippet contains a control sequence (e.g. a `read_files` of an escape-bearing file, or an escape-bearing filename), **When** the run reports the result, **Then** the `[Tool Result]` line is control-free.
3. **Given** a call whose argument value (or key) contains an escape, **When** the run reports the action, **Then** the `[Tool Action]` line is control-free and its sorted, `reason`-excluded argument list is otherwise unchanged.

**Functional Requirements**:

- **FR-001**: Each `[Tool Reason]`, `[Tool Result]`, and `[Tool Action]` line MUST carry **no** terminal control sequence — no ANSI escape sequence (CSI/SGR, OSC, or any other ESC-introduced sequence) and no stray C0/DEL control character — the **same class** already removed from `[Tool Output]` (TAB kept; bytes `≥ 0x80` untouched; ASCII-gated ESC consumption; per-kind bounded scan window).
- **FR-002**: The removal MUST be **presentation-only**: it MUST NOT alter the tool **result** fed to the model, the call's raw arguments, any persisted record, `stdout`, the `[Tool Output]` header/separator literals, or the class-phrase vocabulary.
- **FR-003**: The control-removal policy MUST be **single-owned** in `internal/ui` — **one** definition of the class, consumed by **all four** `[Tool …]` formatters (`[Tool Output]`, `[Tool Reason]`, `[Tool Result]`, `[Tool Action]`).
- **FR-004**: The removal MUST be applied **before** the rune cap (after the round-036 fold + trim), so the existing caps bound the **visible** output and the round-036 one-line / cap semantics for the reason and result are preserved (cap set stays `{189, 200, 200}`).
- **FR-005**: A reason whose text is **only** control data (non-blank before removal, blank after) MUST NOT render a dangling `[Tool Reason]` row — i.e. the blank-reason suppression is evaluated on the **sanitized, folded, trimmed** value.

### User Story 2 - The live turn output is grouped by blank lines (Priority: P2)

As an operator reading a tool-using turn, I want the begin blocks, the trailing reason block, and the post-status group each to be set off by a blank line, so the run is scannable instead of a dense wall of lines.

**Why this priority**: it is the operator's spacing request; it is additive and independent of US1.

**Independent verification**: script a **multi-call** round and a **reason-less** call, and assert the exact blank-line positions in the captured `stderr` (one before each call's begin block, one before the grouped tail block, one before the post-status group).

**Acceptance Scenarios**:

1. **Given** a turn whose round makes `k` tool calls each stating a reason, **When** the run logs the calls, **Then** a blank line precedes each call's `[Tool Reason]` line (and, within a call, the `[Tool Action]`/`[Tool Result]` lines follow the reason with no blank between them).
2. **Given** a call that states **no** reason, **When** the run logs it, **Then** a blank line precedes its `[Tool Action]` line.
3. **Given** a non-final call's tail, **When** the run emits the grouped `[Tool Reason]` block, **Then** exactly **one** blank line precedes the block and **no** blank line separates its lines; **And** when a measured post-status group follows, exactly **one** blank line precedes it.

**Functional Requirements**:

- **FR-006**: Exactly one blank line MUST precede each executed call's **begin block** — the `[Tool Reason]` line when the call states a reason, else the `[Tool Action]` line. Per call, so a `k`-call round emits `k` such blanks.
- **FR-007**: Exactly one blank line MUST precede the grouped post-call **tail** `[Tool Reason]` block, when that block is non-empty; **no** blank line may separate the block's lines.
- **FR-008**: Exactly one blank line MUST precede the **post-status group** (measured `Payload:` line + metrics line + `Ready` footer), emitted only when that group is written (the usage-reporting gate).

### Edge cases

- **A reason that is only an escape** (`"\x1b[31m"`) → US1 removes the escape; FR-005 makes the row a no-op (no dangling `[Tool Reason]`).
- **An escape split across a `[Tool Output]` `Write` boundary** → unchanged from round 038 (sanitization runs on each assembled line).
- **A `\r`/`\n` inside a reason** → still folded to a space by the round-036 fold; the fold precedes the sanitize (FR-004).
- **An escape-bearing argument key** (not just a value) → US1 covers keys too (FR-001/FR-003).
- **A multi-call round / a reason-less call** → US2's blank rule is per call (FR-006), pinned by the multi-call and reason-less Examples.
- **A provider that reports no usage** → no measured `Payload:`/metrics/`Ready` group → no blank (FR-008).
- **A non-tool turn** → no tool-loop lines and no new blanks; the frame gap before the answer is unchanged.
- **`raw` / `-i` / a non-terminal `stderr`** → the tool-loop lines (and their blanks) ride the same diagnostic stream wherever the lines do; `stdout` is untouched.

### Key entities

- **Terminal control sequence** — an ESC-introduced ANSI sequence (CSI/SGR, OSC, …) or a stray C0/DEL control character; the class removed by FR-001.
- **`[Tool …]` diagnostic line** — one of `[Tool Reason]` / `[Tool Result]` / `[Tool Action]` / `[Tool Output]` on the shared `stderr`.
- **Turn begin block / tail block / post-status group** — the three grouping units US2 separates with one blank line each.

## Requirements *(mandatory)*

### Global requirements

- **FR-009**: The round MUST NOT change any CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record shape (`history.jsonl` / `tokens.log` / `global_prompts.jsonl` / `tools-count.jsonl`), the tool **result** fed to the model, or `stdout` bytes; a non-tool turn is unchanged.
- **FR-010**: The offline paths (`--version`, `-d`, `-l`, `--tool-usage`) MUST remain unchanged.
- **FR-011**: The blank lines MUST be plain newlines on the diagnostic stream only (no trailing whitespace, no `stdout` byte, no persisted byte).

### Out of scope (recorded)

- The neutral-close restore stays `[Tool Output]`-scoped (Q4); no per-line reset is added to the single-line formatters.
- No change to `output_file` handling, the `[Tool Output]` bound/stop semantics, the spinner yield (round 035), the block literals, the tool schemas, or the provider transports.
- Consolidating the **three-site** blank-reason predicate (round 036) into one owner remains on [#69](https://github.com/gosharplite/tellme/issues/69); this round only **extends** the predicate's input to the sanitized value (FR-005).

## Success criteria *(mandatory)*

- **SC-001**: Hostile-fixture unit pins show `FormatToolReason`, `FormatToolResult`, and `FormatToolAction` output is control-free and valid UTF-8, with visible text and the rune caps preserved (US1).
- **SC-002**: An E2E run whose reason/result carry an escape emits control-free `[Tool Reason]` / `[Tool Result]` lines (US1).
- **SC-003**: E2E assertions pin the exact blank-line positions for a multi-call round, a reason-less call, the grouped tail block, and the post-status group (US2).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit (re-run after the truth MODIFY) are green.
- **SC-005**: The change has **falsifiability witnesses**: reverting a sibling's sanitize fails its unit pin; removing a blank fails the corresponding E2E position assertion; narrowing the reason guard back to the raw value fails the escape-only-reason pin — each reproduced then reverted.

## Assumptions

- **Truth ownership.** `specs/truth/features/cli/chat/**` + `chat/dsl.md` are `TruthArtifact`s owned by `/axb-dsl-refine`; `techstack.md` is owned by `/axb-technical-research`; **ADR 0008** is ADDED and **supersedes ADR 0007** (governance, not truth; research D9 — 0007's `Status` flips to `Superseded by 0008`, it is not amended in place); `/axb-api-plan` is a checked **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted-state change). All edits are recorded in this round's `truth-delta.md`.
- **Frozen history.** Every earlier plan package (incl. `034-tool-call-log-parity`, `035-spinner-tail-residue`, `036-tool-reason-sanitize`, `038-tool-output-sanitize-and-empty-submit`) is never edited; 039 is a fresh package.
- **`/axb-spec-by-example` is NOT a NOOP** — both workstreams are user-visible (a control-free line; a blank line), so 039 writes plan-side acceptance rules that `/axb-dsl-refine` maps onto interface rows (`acceptance-coverage`).
- **`/axb-ui-plan` is skipped** — no TUI chrome change and no new screen; this is a text-stream change.
- **Single-ownership enabler.** The sanitizer's move to one home (`sanitize.go`) is an internal enabler for US1 (not a separate user story); it must keep **byte-identical** behavior for `[Tool Output]` (round 038 regression guard).
- **Reference check.** `/axb-technical-research` verifies the reference (`tell-me-go`) for sibling sanitization and for blank-line spacing; any divergence is recorded (round-037/038 pattern).
- **Folded issue.** Issue #80 is the anchor; the blank-line workstream is an operator request (no issue).
- **Branch.** `039-terminal-safe-lines-and-turn-spacing` off `dev`.
