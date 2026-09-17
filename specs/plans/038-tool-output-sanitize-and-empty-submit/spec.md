# Feature Specification: tool-output control-sequence sanitization + empty-submit no-op (round 038)

**Feature Branch**: `038-tool-output-sanitize-and-empty-submit`

**Created**: 2026-09-17

**Status**: Draft — behavior/parity change. Operator-locked decisions Q1–Q3 (see below).

**Input**: Two operator-filed issues, folded into one round at the operator's request:

- **Issue [#78](https://github.com/gosharplite/tellme/issues/78)** — *"text colour leak in the terminal"*: the live `[Tool Output]` block forwards a shell command's **raw** bytes to `stderr`, so a colouring command that never resets (or is killed mid-output) leaves the terminal tinted for everything printed afterwards.
- **Issue [#76](https://github.com/gosharplite/tellme/issues/76)** — with `tellme -i`, `Ctrl+S`/`Alt+Enter` on an **empty** editor quits the prompt and exits `0`, whereas the reference (`tell-me-go`) treats an empty submit as a **no-op** and stays in the prompt.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → Strip the whole ANSI/control-sequence class.** The streamed `[Tool Output]` presentation removes **every** terminal control sequence (SGR colour/attributes, cursor movement, `\x1b[?25l` style private sequences, OSC titles, …) plus stray C0 control characters, keeping the visible text and the `[HH:MM:SS] [Tool Output] ` framing. Rationale: it fixes the reported colour leak **and** the sibling control classes in one rule, and cannot strand a terminal state.
- **Q2 → Always close the block in a neutral state.** Independently of stripping, the block's close (and its start-failure / abort / trim / timeout paths) leaves the terminal in a **neutral** attribute state. Rationale: cheap defence-in-depth covering the deliberately-**dropped trailing partial line** and the **killed-mid-output** cases, so the block can never strand the terminal even if stripping ever missed.
- **Q3 → Mirror the reference: an empty submit is a no-op.** `Ctrl+S`/`Alt+Enter` with an **empty** (or whitespace-only) editor keeps the prompt open; only `Esc`/`Ctrl+C` aborts, and a **non-empty** submit is unchanged. Rationale: restores reference parity and matches this round's parity intent; an accidental empty `Ctrl+S` can no longer end the session.

**Scope note**: two independent presentation/parity changes on the `stderr`-bound interactive surface. Neither changes a CLI flag, exit code, the frozen class-phrase vocabulary, the tool **result** fed to the model, the provider transport, any persisted record, or a `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A shell command's output never leaks terminal control sequences (Priority: P1)

As an operator running a turn whose tool runs a colouring command, I want the streamed `[Tool Output]` block to be free of terminal control sequences — so my terminal is not left tinted or otherwise altered after the command runs.

**Why this priority**: it is the operator's reported symptom (issue #78) and the round's whole substance for the tool-output surface.

**Independent verification**: script a shell command that emits a colour escape with no reset, run a turn, and assert the `[Tool Output]` **content** lines carry no escape byte while the visible text is preserved; unit-pin the sanitizer over a hostile fixture.

**Acceptance Scenarios**:

1. **Given** a tool-using turn whose command emits a colour escape that is never reset, **When** the run streams the command's output, **Then** the streamed `[Tool Output]` lines carry **no** terminal control sequence and the command's visible text is preserved.
2. **Given** a command that emits an OSC title / cursor-hiding sequence, **When** the run streams its output, **Then** those sequences do not appear in the streamed lines.

**Functional Requirements**:

- **FR-001**: Each streamed `[Tool Output]` **content** line MUST carry no terminal control sequence — no ANSI escape sequence (CSI/SGR, OSC, or any other ESC-introduced sequence) and no stray C0/DEL control character — while preserving the line's visible text and the `[HH:MM:SS] [Tool Output] ` framing.

- **FR-002**: The sanitization MUST be **presentation-only**: it MUST NOT alter the tool **result** fed back to the model, `stdout`, the `-r` output, the `[Tool Output]` header/separator literals, or any other diagnostic line.

### User Story 2 - An empty submit keeps the interactive prompt open (Priority: P2)

As an operator at the `-i` prompt, I want `Ctrl+S`/`Alt+Enter` on an empty editor to do **nothing** — so an accidental empty submit does not end my session.

**Why this priority**: it is the operator's second folded issue (#76) and restores reference parity; it is independent of US1 but shares the round.

**Independent verification**: script `-i` with an empty submit followed by a real prompt submit; assert the run still makes exactly one request and exits successfully (an empty submit that quit would end the run with no request).

**Acceptance Scenarios**:

1. **Given** the interactive prompt with an empty editor, **When** the operator presses `Ctrl+S`, **Then** the prompt stays open (no turn, no exit) and the operator can keep typing.
2. **Given** the prompt stayed open after an empty submit, **When** the operator types a prompt and submits, **Then** exactly one reasoning turn runs.

**Functional Requirements**:

- **FR-003**: Submitting an **empty** (or whitespace-only) editor at the `-i` prompt MUST be a **no-op** — the prompt stays open and the run continues; only an abort (`Esc`/`Ctrl+C`) or a non-empty submit ends it.

- **FR-004**: A **non-empty** submission MUST behave exactly as today (`Ctrl+S`/`Alt+Enter` submits the composed prompt and runs one turn).

### User Story 3 - The streamed output block always closes in a neutral state (Priority: P3)

As an operator, I want the `[Tool Output]` block to leave my terminal in its default state when it closes — including when the command is stopped mid-output — so no partial state survives the block.

**Why this priority**: it is hardening for US1 (defence-in-depth); with US1 stripping it is a redundant guarantee, and the residual cases (dropped partial line, killed mid-output) make it worth keeping.

**Independent verification**: unit-pin the writer so `End()` (normal, and the start-failure path) emits a neutral-state restore after the closing separator.

**Acceptance Scenarios**:

1. **Given** a `[Tool Output]` block, **When** it closes, **Then** the presentation restores the terminal's default attribute state.

**Functional Requirements**:

- **FR-005**: When a `[Tool Output]` block closes — normal close, start failure, or the abort/trim/timeout path — the presentation MUST leave the terminal in a neutral attribute state.

### Edge cases

- **A command whose reset lands in the dropped trailing partial line** → US3's neutral-close restore covers it; US1 stripping means the earlier set-colour line never tinted in the first place.
- **A command killed at the byte budget or its `timeout`** → the block still closes with the neutral restore.
- **A command that emits colour *with* a reset** → the visible text is preserved; the escapes are removed.
- **An over-long line / a line split across `Write` calls** → sanitization runs on each **assembled** line, so a sequence split across writes is still removed.
- **Repeated empty submits** → the prompt stays open every time (no accumulation).
- **A whitespace-only editor** (`"   "`) → treated as empty (no-op).
- **Non-terminal `stderr` / `-r` / a non-shell tool** → no `[Tool Output]` block; unchanged.

### Key entities

- **Streamed tool-output line** — a `[HH:MM:SS] [Tool Output] <text>` diagnostic line for a shell-class call without `output_file`.
- **Terminal control sequence** — an ESC-introduced ANSI sequence (CSI/SGR, OSC, …) or a stray C0/DEL control character; the class removed by FR-001.
- **Neutral terminal state** — the terminal's default attributes (no colour/attribute left set); the FR-005 close guarantee.
- **Interactive submit** — the `Ctrl+S`/`Alt+Enter` action at the `-i` prompt; empty ⇒ no-op (FR-003).

## Requirements *(mandatory)*

### Global requirements

- **FR-006**: The round MUST NOT change any CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record shape (`history.jsonl` / `tokens.log` / `global_prompts.jsonl` / `tools-count.jsonl`), the tool **result** fed to the model, or `stdout` bytes; the only affected surfaces are the `stderr`-bound `[Tool Output]` block and the `-i` prompt's empty-submit handling.
- **FR-007**: The offline paths (`--version`, `-d`, `-l`, `--tool-usage`) MUST remain unchanged.

### Out of scope (recorded)

- The reference (`tell-me-go`), which has **no** output sanitizer — this is a deliberate, **recorded divergence** (it fixes a real terminal-state leak the reference also has).
- Any change to `output_file` handling, the block's bound/stop semantics, the spinner yield, or the `[Tool Output]` header/separator literals.
- Any change to the suggestion engine, editor, placeholder, or other `-i` keybindings.

## Success criteria *(mandatory)*

- **SC-001**: A colouring command's streamed `[Tool Output]` **content** lines carry **zero** escape bytes (US1) with the visible text preserved.
- **SC-002**: The `[Tool Output]` block closes in a neutral state on the normal and the start-failure paths (US3).
- **SC-003**: An empty submit at the `-i` prompt does **not** end the run; a following non-empty submit runs **exactly one** turn (US2).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit (re-run after the truth MODIFY) are green.
- **SC-005**: The change has **falsifiability witnesses**: removing the sanitizer fails the `[Tool Output]` pin; restoring quit-on-empty fails the empty-submit E2E; removing the neutral close fails the writer unit pin — each reproduced then reverted.

## Assumptions

- **Truth ownership.** `specs/truth/features/cli/chat/**` + `chat/dsl.md` are `TruthArtifact`s owned by `/axb-dsl-refine`; `techstack.md` is owned by `/axb-technical-research`; `/axb-api-plan` is **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted-state change). All truth edits are recorded in this round's `truth-delta.md`.
- **Frozen history.** Earlier plan packages (incl. `034-tool-call-log-parity`, `037-interactive-prompt-no-selection`) are never edited; 038 is a fresh package.
- **`/axb-spec-by-example` is NOT a NOOP.** The control-sequence-free block is a **user-visible acceptance** change, so 038 writes two plan-side acceptance rules (US1/US3 and US2), which `/axb-dsl-refine` maps 1:1 onto interface `Then`/`When` rows (`acceptance-coverage`).
- **`/axb-ui-plan` is skipped.** No TUI chrome changes and no new screen; this is a text-stream + key-handling change.
- **Sanitizer home.** The removal lives at the `internal/ui` presentation seam (the `[Tool Output]` line formatter) so the tool result is untouched (FR-002) and the rule is unit-testable; the neutral-close restore lives in the same writer.
- **Folded issue.** Issue #76 is folded into this round per the operator's direction (recorded here and in `truth-delta.md`).
