# Feature Specification: tellme turn spinner — a live progress indicator for prompt turns (round 019)

**Feature Branch**: `019-turn-spinner`

**Created**: 2026-09-14

**Status**: Draft — spinner visible form, metrics, and scope locked (clarify Q1=2, Q2=2)

**Input**: Operator request, verbatim: "spinner will be the theme of slice 019. Between `user prompt is captured` and `user re-gain terminal control`, a spinner will always be visible." Locked by the operator: **reference-parity stop/resume** (option A); the **`-i` TUI is out of scope** (it precedes prompt capture); **Linux/macOS only** (no Windows).

Grounded against the reference: `tell-me-go` `internal/ui/renderer_spinner.go` (braille frames; `{status} ({elapsed}s)`; the `[CPU: … | MEM: …]` metrics variant) and `internal/agent/session/ui/spinner.go` (the per-phase start/stop/resume lifecycle), plus the per-phase status labels in `internal/domain/events/types.go`.

Behaviour intent: **ADD** a live, animated progress spinner to `tellme`'s non-TUI prompt turns, so that between prompt capture and the moment the user regains terminal control the run is never silently waiting.

**Locked decisions (operator clarify, 2026-09-14)**:

- **Q1 → 2 (full reference parity, with identifiers)**: `⠋ Thinking [<model>]... (3s)` / `⠋ Executing [<tool>]... (3s)` / `⠋ Executing tools [<a>, <b>]... (3s)`.
- **Q2 → 2 (keep the metrics variant)**: the tool-execution spinner also shows ` [CPU: x.x% | MEM: y.y%]`.
- **Stop/resume (A)**: the spinner is visible throughout every waiting interval; it is cleared only to emit interleaved output; it is restored while waiting resumes.
- **Scope**: non-TUI surfaces only (positional / piped / the round-012 reader); the `-i` TUI is out of scope.
- **Platform**: Linux/macOS only (no Windows).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The run is never silently waiting (Priority: P1)

As an operator, I want a live spinner on the diagnostic stream for the whole time `tellme` is working on my prompt — from the moment my input is captured until I regain terminal control — so the run never looks hung while the model or a tool takes time.

**Why this priority**: it is the round's entire reason to exist; without a live indicator the operator cannot tell "still working" from "dead".

**Independent verification**: run a prompt turn hermetically on a forced-terminal stream (injected clock + injected streams); assert an animated spinner is present on the diagnostic stream while waiting, that its label tracks the phase, that it is cleared before the answer, and that it is absent once the turn completes.

**Acceptance Scenarios**:

1. **Given** a resolved setup and a prompt on a terminal, **When** the turn captures the input and begins waiting, **Then** the diagnostic stream shows a spinner whose frame advances over time (animated) until the answer is written.
2. **Given** the turn is awaiting the model, **When** the spinner is drawn, **Then** its status label reads ` Thinking [<model>]...` and its `(elapsed)` counter advances in whole seconds.
3. **Given** the turn is executing tools, **When** the spinner is drawn, **Then** its status label reads ` Executing [<tool>]...` for a single tool or ` Executing tools [<a>, <b>]...` for several, and the line carries the ` [CPU: <cpu>% | MEM: <mem>%]` segment.
4. **Given** the turn emits an interleaved line (a tool trace) and then waits again, **When** output resumes, **Then** the spinner is cleared for the line and restored once waiting resumes.
5. **Given** the answer is about to be written, **When** the response arrives, **Then** the spinner is cleared and no spinner line survives into the completed turn.

**Functional Requirements**:

- **FR-001**: On every prompt-bearing non-TUI turn, from prompt capture until the answer begins, the system MUST display a spinner on the diagnostic stream whenever it is waiting (awaiting the model, or executing tools); there MUST be no silent waiting interval except while interleaved output is being emitted.
- **FR-002**: The system MUST clear the spinner before writing any interleaved output (a tool-trace line, the answer, or the post-turn status lines) and MUST restore the spinner if waiting resumes within the same turn; the spinner MUST NOT persist into the completed turn.
- **FR-003**: The spinner MUST render `{frame}{status} ({elapsed}s)` — a single braille spinner frame that advances over time, the phase status label, and the whole-seconds elapsed counter — and MUST be animated.
- **FR-004**: The status label MUST track the waiting phase: awaiting the model → ` Thinking [<model>]...` (the `[<model>]` bracket omitted when the active provider names no model); executing tools → ` Executing [<tool>]...` for one tool, ` Executing tools [<a>, <b>]...` for several, or ` Executing tools...` when tool names are unavailable. Each label carries a leading space.
- **FR-005**: The tool-execution spinner MUST additionally render ` [CPU: <cpu>% | MEM: <mem>%]` (host-process CPU and memory percentages, one decimal place); the awaiting-the-model spinner MUST NOT render that segment.

**Non-Functional Requirements**:

- **NFR-001**: The elapsed counter MUST be whole seconds.
- **NFR-002**: The spinner line MUST be plain text (no ANSI colour).

---

### User Story 2 - The spinner stays off the data streams and other surfaces (Priority: P2)

As an operator, I want the spinner limited to the interactive terminal surfaces and kept off the data streams, so piping, `-r`/raw output, the `-i` TUI, and the non-prompt commands keep their current behaviour.

**Why this priority**: it bounds the round and protects rounds 004–018 (piping, rendering, chrome, post-turn lines); it is a constraint on Story 1, not an independent capability.

**Independent verification**: run a piped/redirected run, a `-r` run, and each non-prompt path hermetically; assert no spinner is drawn, that `stdout` is byte-exact, and that the class phrase + exit code are unchanged on failure.

**Acceptance Scenarios**:

1. **Given** a run whose **diagnostic stream (`stderr`)** is **not** a terminal, or a `-r`/`--raw` run, **When** the turn runs, **Then** no spinner is drawn and `stdout` is byte-exact.
2. **Given** `--version`, `-d`, `-l N`, the boot path, or a prompt-less `--new`, **When** it runs, **Then** no spinner is drawn.
3. **Given** the `-i` TUI, **When** the turn runs, **Then** the `-i` surface is unchanged (out of scope).
4. **Given** a turn that fails (provider / tool / history), **When** it fails, **Then** the spinner is cleared before the class phrase and the class phrase + exit code are unchanged.

**Functional Requirements**:

- **FR-006**: The spinner MUST be drawn only when the **diagnostic stream (`stderr`) is a terminal** and `-r`/`--raw` is NOT set; when `stderr` is not a terminal, or `-r` is set, the system MUST draw no spinner. *(Mirrors the reference — its `stderr` spinner is gated on `stderr` terminality (`ui.stderr`), never on `stdout`; a `stdout` gate would both write `\r` frames into a redirected diagnostic and drop feedback when only `stdout` is redirected. This does **not** wire a standard-output probe — round-006 / PR #16 **Obs 1** stays **OPEN**.)*
- **FR-007**: The spinner MUST be written to `stderr` only; `stdout` MUST remain byte-exact.
- **FR-008**: The spinner MUST appear only when **both** hold: (a) the prompt surface is a non-TUI prompt-bearing turn (positional, piped, or the round-012 interactive reader), **and** (b) the FR-006 gate is satisfied (`stderr` is a terminal and `-r` is not set). It MUST NOT appear on `--version`, `-d`, `-l N`, the boot path, a prompt-less `--new`, or the `-i` TUI — the surface exclusion is a designed constraint, not merely an emergent consequence of the gate.
- **FR-009**: On any turn failure (provider / tool / history), the system MUST clear the spinner before writing the class phrase; the class phrase and the exit code MUST be unchanged.
- **FR-010**: The round MUST NOT introduce a new frozen class phrase, and the class-phrase vocabulary MUST remain unchanged.

**Non-Functional Requirements**:

- **NFR-003**: The behaviour MUST be POSIX-only (Linux/macOS); no Windows variant is provided.
- **NFR-004**: The behaviour MUST be deterministic and verifiable without a real terminal (an injected clock seam + injected streams; no pty).

---

### Edge Cases

- **Waiting interval shorter than one frame tick** → the first frame is still drawn synchronously (the spinner appears for even very short waits) and is cleared with no residue line.
- **Tool-loop turn** → the spinner cycles per phase (awaiting the model → executing tools → awaiting the model …); each tool-execution state lists the current tool(s) and carries the metrics segment.
- **Provider with no model name** → the label renders ` Thinking...` (no bracket); ` Executing tools...` when tool names are unavailable.
- **Interleaved output mid-turn** → the spinner is cleared before the line, then restored while waiting resumes.
- **Turn failure** → the spinner is cleared; the class phrase + exit code are unchanged; no spinner residue remains.
- **Non-terminal `stdout` / `-r`** → no spinner at all.
- **Non-prompt paths** → no spinner.
- **Very long wait** → the elapsed counter grows without bound and formats without overflow/panic.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-011**: The round MUST NOT change the round-009 payload status lines, the round-017 turn chrome, or the round-018 post-turn status lines (their bytes are unchanged).

#### Non-Functional Requirements

- **NFR-005**: The round MUST add **no new third-party dependency** (the host CPU/memory sampling is dependency-free, POSIX-only).

### Key Entities *(include if feature involves data)*

- **Spinner state**: the current waiting phase's frame index, status label, and elapsed counter — recreated for each waiting interval, cleared when the interval ends.
- **Phase status label**: the text derived from the waiting phase (awaiting the model vs. executing tools), with the interpolated model / tool identifier(s) and the leading space.
- **Resource segment**: the ` [CPU: … | MEM: …]` host-process CPU/memory percentages shown only during tool execution.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A hermetically-run prompt turn on a forced-terminal stream shows an animated spinner whose label tracks the phase (awaiting the model / executing tools), while `stdout` stays byte-exact (covers FR-001–FR-004, FR-007).
- **SC-002**: The tool-execution spinner carries the ` [CPU: … | MEM: …]` segment and the awaiting-the-model spinner does not (covers FR-005).
- **SC-003**: No spinner is emitted when the diagnostic stream (`stderr`) is not a terminal, when `-r` is set, on any non-prompt path, or on the `-i` TUI (covers FR-006–FR-008).
- **SC-004**: On failure the spinner is cleared and the class phrase + exit code are unchanged; the class-phrase vocabulary is unchanged (covers FR-009, FR-010).
- **SC-005**: The behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (reference)**: the parity target is the reference's `renderer_spinner.go` (frames, `{status} ({elapsed}s)`, the metrics variant) + `spinner.go` (phase lifecycle), with the phase labels from `internal/domain/events/types.go`.
- **A2 (labels)**: `<model>` = the active provider's configured model; the bracket is omitted when empty; tool names come from the tool call(s) in the current phase, joined `[a, b]` when several.
- **A3 (metrics)**: the ` [CPU: <cpu>% | MEM: <mem>%]` segment reports **machine-wide** CPU and memory (host CPU = Δ of `Σcpu − idle`; host memory percent), sampled by a dependency-free, POSIX-only mechanism (Linux `/proc/stat` + `/proc/meminfo`; macOS `sysctl`/mach — cgo and nocgo variants) behind a domain port — pinned by `/axb-technical-research`.
- **A4 (frames / cadence)**: a single-character braille frame advancing over time (the reference's frame set and tick cadence are the parity target) — pinned by `/axb-technical-research`.
- **A5 (gate)**: drawn only when the **`stderr`** is a terminal and `-r` is off — mirroring the reference's `IsTerminalContext()` (which reads `ui.stderr`), driven in E2E by a `TELL_ME_FORCE_STDERR_TTY` seam (mirroring the round-012 stdin seam). This does **not** wire a standard-output probe; round-006 / PR #16 **Obs 1** stays **OPEN**.
- **A6 (surfaces)**: positional, piped, and the round-012 reader; **not** the `-i` TUI; **not** the non-prompt paths.
- **A7 (streams)**: `stderr` only; `stdout` byte-exact; plain text (no ANSI).
- **A8 (elapsed)**: whole seconds for the current waiting interval; a fresh interval after interleaved output restarts the counter. The exact reset behaviour across a phase transition is pinned by `/axb-technical-research` (the reference updates the status in place within a turn).
- **A9 (platform / verification)**: POSIX-only (Linux/macOS); no new third-party dependency; verification is hermetic (injected clock + streams; no pty).
- **A10 (scope-bound)**: the round closes the deferred standard-output terminal probe (Obs 1) but changes no other behaviour; the class-phrase vocabulary stays unchanged.
