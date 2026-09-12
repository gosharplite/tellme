# Feature Specification: tellme Rendered Output & Raw Flag (round 006)

**Feature Branch**: `006-rendered-output-and-raw-flag`

**Created**: 2026-09-12

**Status**: Draft

**Input**: User request "006 — rendered output + `-r`". Resolved by Clarify Round 1 (2026-09-12); the recommendations were **verified against the `tell-me-go` reference** before ratification:

- **Q1 -> Option 1 (Reference parity — default rendered output; `-r` raw; TTY gates only tellme's own decoration)**: with no `-r`, the answer is Markdown-rendered on **all** standard-output streams; `-r`/`--raw` prints plain text; the terminal check (`isTTY && !raw`) gates only tellme's **own** presentation (color/spinner/labels) — never the rendering. Reference: `renderTextLocked` gates rendering on `-r` alone; `UseColor = isTTY && !RawOutput`. **This amends round-005 FR-007**: the byte-exact/plain piped stream is obtained via `-r`, not automatically at a non-terminal.
- **Q2 -> Option 1 (Renderer library only — glamour; no pty)**: adopt the reference's renderer (`github.com/charmbracelet/glamour`, with its transitive `termenv`/`reflow`) as the round's first presentation dependency; add **no** pty. The reference ships no pty (`go.mod`), so the TTY-sensitive decoration branch stays a **named pin** (round-005 precedent).
- **Q3 -> Option 1 (Include `WRAP_WIDTH` + `TELL_ME_WRAP_WIDTH`)**: word-wrap width; `>= 0`; `0` = the renderer's default (80); applies to rendered output and is ignored under `-r`. Reference: `WrapWidth int` (`yaml:"WRAP_WIDTH"`), `BindEnv("WRAP_WIDTH","TELL_ME_WRAP_WIDTH")`.

*(Deferred to `/axb-technical-research` as **parity details to pin** — reference behaviours the contract relies on: the renderer's build options + `GLAMOUR_STYLE`; the rendered/raw **byte handling** (trailing newline); the LaTeX→Unicode **sanitization** applied before rendering; and the graceful **degradation path** when the renderer fails to initialize.)*

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A formatted answer by default (Priority: P1)

As a developer at a terminal, I want tellme to render the model's answer as formatted Markdown by default, so that headings, emphasis, lists and code blocks are readable — matching the reference's default output.

**Why this priority**: The default output is the primary thing every user sees, and it is the counterpart that gives the `-r` flag meaning. Without it there is no rendered output at all, so it is the first value to prove.

**Independent verification**: run a turn whose answer contains Markdown (e.g. `**bold**`, a list, a fenced code block) with standard output captured; confirm the default output is the **rendered** form and not the literal Markdown source. Run it both at a pseudo-terminal and with output redirected to confirm rendering is stream-independent while tellme's own presentation is not.

**Acceptance Scenarios**:

1. **Given** a successful turn and no `-r`/`--raw` flag, **When** the turn completes, **Then** the answer is rendered as formatted Markdown on standard output (the literal Markdown source is not printed verbatim), and the process exits with the success code.
2. **Given** a successful turn with no `-r` flag whose standard output is **not** a terminal, **When** the turn completes, **Then** tellme adds none of its **own** terminal presentation (color, spinner, labels) — the rendered answer is the only thing emitted.

**Functional Requirements**:

- **FR-001**: With no `-r`/`--raw` flag, the system MUST render the answer as Markdown to standard output.
- **FR-002**: Rendering MUST be governed by the `-r` flag alone — it MUST NOT be suppressed by whether standard output is a terminal.
- **FR-003**: The system MUST use its **own** terminal presentation (color, spinner, labels) only when standard output is a terminal, and MUST suppress that presentation when standard output is not a terminal. *(Amends round-005 FR-007: the non-terminal suppression now covers tellme's own decoration, not the answer's rendering.)*

### User Story 2 - Raw output for pipelines (`-r`) (Priority: P2)

As an automation-script author, I want a `-r`/`--raw` flag that prints the model's answer as plain text (no Markdown rendering), so that I can pipe tellme's answer into other tools without terminal formatting.

**Why this priority**: It is the inverse of Story 1 and the escape hatch that makes tellme composable in a pipeline; it is only meaningful once the default output is rendered.

**Independent verification**: run a turn with `-r` whose answer contains Markdown; confirm the output contains the literal Markdown source and carries no renderer-introduced ANSI escape sequences.

**Acceptance Scenarios**:

1. **Given** a successful turn and the `-r` flag, **When** the turn completes, **Then** the answer is printed as **plain text** — no Markdown rendering and no renderer-introduced escape sequences — and the process exits with the success code.

**Functional Requirements**:

- **FR-004**: The system MUST provide a `-r`/`--raw` flag; when set, the system MUST print the answer as plain text with no Markdown rendering.
- **FR-005**: When `-r` is set, the rendered path MUST be bypassed entirely — no renderer-introduced escape sequences may appear in the answer.

### User Story 3 - Configurable wrap width (Priority: P3)

As a user with a wide or narrow terminal, I want to control the column width at which rendered output wraps — via configuration or the `TELL_ME_WRAP_WIDTH` environment variable — so the formatting fits my display.

**Why this priority**: A refinement of the rendered output (Story 1): valuable for readability, but not required for the core default/raw pair to be useful.

**Independent verification**: set `TELL_ME_WRAP_WIDTH=40`; confirm rendered output wraps at approximately 40 columns; unset it (or set `WRAP_WIDTH: 0`) and confirm the renderer's default width is used.

**Acceptance Scenarios**:

1. **Given** `WRAP_WIDTH` set in the configuration, or `TELL_ME_WRAP_WIDTH` set in the environment, **When** a turn's answer is rendered, **Then** the rendered output wraps at that column width.
2. **Given** `WRAP_WIDTH` is `0` (or unset), **When** an answer is rendered, **Then** the renderer's default width is used.

**Functional Requirements**:

- **FR-006**: The system MUST accept a `WRAP_WIDTH` integer configuration and a `TELL_ME_WRAP_WIDTH` environment override, honouring the existing `TELL_ME_*` environment-over-file precedence; `0` MUST mean the renderer's default width, and a negative value MUST be rejected as a configuration error.
- **FR-007**: The wrap width MUST apply only to rendered output and MUST be ignored when `-r` is set.

### Edge Cases

- When the renderer fails to initialize, the system MUST degrade gracefully to **raw (unrendered)** output and MUST NOT fail the turn; at most a single warning may be emitted on standard error (reference ADR-007 parity).
- When `-r` is combined with a prompt turn, the existing answer-on-standard-output contract and exit codes MUST be unchanged (the answer is simply raw).
- When `WRAP_WIDTH` is negative, the system MUST reject the configuration with the existing configuration-error contract and exit code.
- When the answer contains no Markdown, the rendered and raw forms MAY be close but MUST NOT be assumed byte-identical.
- When the answer carries control bytes or ANSI sequences, the `-r` path MUST pass the answer bytes through unchanged (rendering is bypassed).

## Requirements *(mandatory)*

> Story-specific FR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-008**: The system MUST keep its existing flag surface (`-c/--config`, `-d/--diagnostics`, `--version`) and MUST add **only** the `-r`/`--raw` flag in this round.
- **FR-009**: The system MUST NOT regress the existing acceptance behaviour of rounds 001–005; the pre-existing scenarios MUST remain green, **except** where this round deliberately amends round-005 FR-007 (the plain, byte-exact piped stream is now obtained via `-r`, not automatically at a non-terminal).

#### Non-Functional Requirements

- **NFR-001**: The rendering pipeline MUST NOT introduce network access; the offline paths (`--version`, `-d`, and a prompt-less boot) MUST remain strictly offline and deterministic.
- **NFR-002**: For a given answer, the **raw** (`-r`) bytes written to standard output MUST be deterministic across repeated runs. Rendered output is a presentation zone and is deliberately **outside** the byte-determinism contract.
- **NFR-003**: The render/raw mode selection and the wrap-width resolution MUST be covered by fast, isolated unit tests, complementing the black-box end-to-end path.

### Key Entities *(include if feature involves data)*

- **Answer rendering mode**: rendered (default) or raw (`-r`) — the output-mode branch (FR-001/FR-004).
- **Wrap width**: the effective column width for rendered output, resolved from `WRAP_WIDTH` / `TELL_ME_WRAP_WIDTH`; `0` = renderer default (FR-006).
- **Output stream mode**: whether standard output is a terminal — governs tellme's **own** presentation only, never rendering (FR-002/FR-003).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of default (no `-r`) runs render the answer as Markdown to standard output, on both terminal and non-terminal streams — the literal Markdown source is not printed verbatim.
- **SC-002**: In the acceptance set, 100% of `-r`/`--raw` runs print the answer as plain text with no renderer-introduced escape sequences.
- **SC-003**: In the acceptance set, 100% of wrap-width runs wrap rendered output at the configured/environment width, and `0`/unset uses the renderer's default width.
- **SC-004**: All pre-existing round-001–round-005 acceptance scenarios remain green, with the round-005 FR-007 Examples updated to use `-r` where they assert a plain, byte-exact redirected answer.
- **SC-005**: The render/raw mode selection and the wrap-width resolution are covered by isolated unit tests.

## Assumptions

- The round delivers **reference output parity** only: default rendered output, the `-r`/`--raw` flag, `WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH`, and the stdout terminal probe for tellme's own presentation. It does **not** add history/session rendering (`-l`), a TUI (`-i`), thoughts/tools, streaming, or the `--callback` worker.
- The renderer is the reference's renderer (glamour) — required by the Q1 parity decision (Q2). The exact build options, the rendered/raw byte handling, the LaTeX→Unicode sanitization, and the degradation path are **technical parity details owned by `/axb-technical-research`**; a code fence, emphasis, and a trailing-newline contract are assumed pending ratification.
- No pty dependency is added (Q2); the "standard output is a terminal" branch of FR-003 is a **named pin**, verified only by stand-ins (as in round 005), not by a real pty.
- The round **amends round-005 FR-007**: round 005 pinned "suppress rendering when standard output is not a terminal"; under reference parity the answer is rendered regardless of stream and the non-terminal suppression applies to tellme's **own** presentation only (FR-003). Truth for round 005's `piping-the-answer-out` interface feature is expected to be recorded as a **MODIFY** by `/axb-dsl-refine`.
- No persistence, history, or session state is introduced; the round keeps the single-turn model from round 004 and the stdin piping from round 005.
- The answer continues to be written to standard output only, followed by exactly one CLI-appended terminating newline on the raw path (round-005 FR-006 stands for the raw mode).
