# Feature Specification: tellme turn surface — operator chrome parity with tell-me-go (round 017)

**Feature Branch**: `017-turn-chrome-parity`

**Created**: 2026-09-14

**Status**: Draft — surface scope locked (A + B; input-capture line only)

**Input**: Operator request "I want `tellme` to feel and look like `tell-me-go` on: **input-capture line**, **turn framing**, **spacing**. Let's not concern on post-turn lines in this slice 017." Grounded against the reference: `tell-me-go` `internal/ui/capture.go` (`finalizePrompt`) and `internal/ui/renderer_metrics.go` (`renderTurnHeader`). Anchor issue: [#42](https://github.com/gosharplite/tellme/issues/42).

Behaviour intent: **ADD** the reference's pre-turn operator chrome — an **input-capture acknowledgement**, a **horizontal rule + `╭─⠿ Turn <N> - <mode>` header** wrapping the existing pre-flight payload line, and the reference's **blank-line spacing** — to `tellme`'s **non-TUI** prompt surfaces: **(A)** the positional/piped prompt turn and **(B)** the round-012 interactive plain reader. **NOOP** on the `-i` TUI surface (round 016), on every non-prompt path (boot / `-d` / `-l` / `--version` / prompt-less `--new`), and on the **post-turn** surface (explicitly out of scope this round).

**Locked decisions (2026-09-14, by the operator)**:
- **Surface scope** — the chrome applies to **(A) the positional/piped prompt turn** and **(B) the round-012 interactive plain reader**; the `-i` TUI surface stays exactly as round 016 delivered it.
- **Startup set** — **input-capture line only**: add `[HH:MM:SS] Input captured. Processing...`; the reference's second startup line (`[HH:MM:SS] [Info] Starting chat...`) is **not** added this round.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The turn acknowledges input capture (Priority: P1)

As an operator, I want `tellme` to print the reference's input-capture acknowledgement the moment it has taken my prompt, so a turn opens the same way `tell-me-go`'s does.

**Why this priority**: it is the first of the three things asked for and the first line the operator sees; the framing (Story 2) only reads correctly once the capture line precedes it.

**Independent verification**: run a prompt-bearing invocation hermetically (injected streams + injected clock); assert the diagnostic stream carries `[HH:MM:SS] Input captured. Processing...` exactly once, before the turn frame, and that `stdout` does not carry it.

**Acceptance Scenarios**:

1. **Given** `tellme` is invoked with a positional prompt and a resolved setup, **When** the turn runs, **Then** the diagnostic stream carries exactly one line `[HH:MM:SS] Input captured. Processing...` before the turn frame.
2. **Given** `tellme` receives a prompt through a pipe, **When** the turn runs, **Then** the same acknowledgement line appears exactly once.
3. **Given** a prompt-less invocation at a terminal that reads a non-empty prompt through the round-012 reader, **When** the prompt is accepted, **Then** the same acknowledgement line appears exactly once.
4. **Given** a non-prompt path (`--version`, `-d`, `-l`, boot, or a prompt-less `--new`), **When** it runs, **Then** no acknowledgement line appears.
5. **Given** the `-i` TUI path, **When** a prompt is submitted, **Then** no acknowledgement line appears (the round-016 surface is unchanged).

**Functional Requirements**:

- **FR-001**: On surfaces (A) and (B), when a non-empty prompt is captured, the system MUST write exactly one acknowledgement line `[HH:MM:SS] Input captured. Processing...` to the diagnostic stream (`stderr`).
- **FR-002**: The acknowledgement MUST be written once, before the turn frame, and MUST NOT be written to `stdout`.

**Non-Functional Requirements**:

- **NFR-001**: The acknowledgement text MUST be exact (no `tellme: ` prefix, so the frozen class-phrase vocabulary is untouched); its timestamp MUST come from the injected clock seam so assertions stay deterministic. The chrome is rendered as **plain text unconditionally** this round (no ANSI); TTY-gated colour is a recorded forward item.

---

### User Story 2 - The turn is framed like tell-me-go (Priority: P1)

As an operator, I want each prompt turn framed with the reference's rule, its `Turn <N> - <mode>` header, and the existing payload line — with the reference's blank-line spacing — so a turn reads identically to `tell-me-go`'s.

**Why this priority**: the frame is the round's central request ("turn framing" + "spacing"); it is what makes the turn *look* the same.

**Independent verification**: run a prompt turn hermetically; assert the diagnostic-stream sequence (leading blank → rule → `╭─⠿ Turn <N> - <mode>` → payload line → blank gap) precedes the answer, that `<N>` equals the session's completed-turn count + 1, and that the answer on `stdout` follows the gap.

**Acceptance Scenarios**:

1. **Given** a prompt turn on a fresh (`--new`) session, **When** it runs, **Then** the diagnostic stream carries a blank line, then the 80-column horizontal rule, then `╭─⠿ Turn 1 - <mode>`, then the pre-flight payload line, then a blank gap — all before the answer.
2. **Given** a session resumed with prior persisted turns, **When** a new prompt turn runs, **Then** the header shows `Turn <prior turns + 1>`.
3. **Given** the frame is rendered, **When** the turn completes, **Then** the rule, header, and payload appear in that order on the diagnostic stream, and the answer (on `stdout`) follows the blank gap.
4. **Given** a non-terminal diagnostic stream, **When** the frame renders, **Then** it is plain text (no ANSI) and the payload line's text format is unchanged from round 009.

**Functional Requirements**:

- **FR-003**: On a prompt-bearing turn, the system MUST write a leading blank line and then the reference's horizontal rule (the fixed 80-column `─` literal) to the diagnostic stream, before the header.
- **FR-004**: Immediately after the rule, the system MUST write a header line `╭─⠿ Turn <N> - <mode>` to the diagnostic stream, where `<N>` is the count of the session's **completed turns** + 1 (tellme stores **one `history_entry` line per completed turn** in the active `history.jsonl`, so the count is the number of such lines — **not** a message count) and `<mode>` is the effective mode.
- **FR-005**: Immediately after the header, the system MUST write the pre-flight payload status line (the round-009 format, unchanged) to the diagnostic stream.
- **FR-006**: After the payload line the system MUST write a blank gap (one blank line), so the answer that follows is visually separated.

**Non-Functional Requirements**:

- **NFR-002**: The frame MUST be written to the diagnostic stream only (`stdout` stays byte-exact), MUST be rendered as **plain text unconditionally** this round (no ANSI; TTY-gated colour is a recorded forward item), and MUST NOT alter the pinned payload-line text (round 009).

---

### User Story 3 - The chrome stays off the paths the operator did not ask for (Priority: P2)

As an operator, I want the new chrome limited to the non-TUI prompt surfaces, so the `-i` TUI (round 016) and the non-prompt commands keep their current output.

**Why this priority**: it bounds the change (the surface-scope decision) and protects prior rounds; it is a constraint on Stories 1–2 rather than a new capability.

**Independent verification**: run the `-i` path and each non-prompt path hermetically; assert no acknowledgement line and no turn frame are written; re-run the prior rounds' suites.

**Acceptance Scenarios**:

1. **Given** `tellme -i` on a terminal, **When** a prompt is submitted, **Then** no input-capture line and no turn frame are written (the round-016 chrome is unchanged).
2. **Given** `-d`, `-l N`, `--version`, boot, or a prompt-less `--new`, **When** it runs, **Then** no input-capture line and no turn frame are written.

**Functional Requirements**:

- **FR-007**: The acknowledgement and the turn frame MUST be emitted only on surfaces (A) and (B); they MUST NOT be emitted on the `-i` TUI surface or on any non-prompt path.
- **FR-008**: The round-016 `-i` surface, the round-009/010 payload status line, and every prior acceptance scenario (rounds 001–016) MUST remain green.

---

### Edge Cases

- **No prompt**: an empty positional prompt, an empty pipe, or an aborted/empty reader submission sends no request — no acknowledgement line and no frame (no turn).
- **`-r` / `--raw`**: the frame still renders on the diagnostic stream (reference behaviour), uncoloured; the answer stays raw on `stdout`.
- **Non-terminal diagnostic stream** (piped `stderr`, `-r`): the frame renders plain (no ANSI); the payload line's text is unchanged.
- **Resumed session**: the header's `<N>` reflects the completed-turn count + 1 (a resumed session does not restart at 1).
- **Resolve failure** (bad config/home): the acknowledgement may precede the boot error (the reference captures input before setup resolution); the error path and its class phrase are unchanged.
- **`-i` on a non-terminal stdin**: the round-016 fallback applies unchanged — no frame.
- **Very narrow terminal**: the rule is a fixed 80-column literal (reference parity); it does not reflow and MUST NOT panic.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-009**: The round MUST NOT introduce a new frozen class phrase (the class-phrase vocabulary is unchanged, still 11) and `stdout` MUST remain byte-exact on every path — the new chrome is written to the diagnostic stream only.

#### Non-Functional Requirements

- **NFR-003**: All new assertions MUST be deterministic (no `time.Sleep`, no real pty); the injected streams and the injected clock seam are the verification surface.

### Key Entities *(include if feature involves data)*

- **Input-capture acknowledgement**: the `[HH:MM:SS] Input captured. Processing...` line written to the diagnostic stream once a non-empty prompt is captured.
- **Turn frame**: the reference chrome around a prompt turn — a leading blank line, the 80-column `─` rule, the `╭─⠿ Turn <N> - <mode>` header, the (unchanged) pre-flight payload line, and a trailing blank gap.
- **Turn number (`<N>`)**: the session's **completed-turn count** + 1 (one `history_entry` line per completed turn; the reference's `SessionTurns + 1`); `--new` → 1.
- **Reference chrome tokens**: the exact character/color tokens the surface must reproduce (the `─` rule width, the `╭─⠿` glyph, the header/blank-line layout, the colour codes and the TTY gate) — a `/axb-technical-research` + `/axb-dsl-refine` determination.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator running `tellme "hi"` (or a piped prompt, or the round-012 reader) sees the reference turn opening — `[HH:MM:SS] Input captured. Processing...`, the rule, `╭─⠿ Turn 1 - <mode>`, the payload line, and a gap before the answer — verified hermetically (covers FR-001–FR-006).
- **SC-002**: The frame's turn number equals the session's completed-turn count + 1 (`Turn 1` on `--new`; `Turn <prior+1>` on a resumed session).
- **SC-003**: On the `-i` TUI and on every non-prompt path, none of the new lines appear, and the round-016 surface plus every prior output are unchanged (SC-003 covers FR-007–FR-008).
- **SC-004**: `stdout` is byte-exact on every path (the chrome lives on the diagnostic stream), the frozen class-phrase vocabulary is unchanged, and the round-009 payload-line text is unchanged.
- **SC-005**: The behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (reference)**: the parity target is `tell-me-go` `internal/ui/capture.go` (`finalizePrompt`: `[HH:MM:SS] Input captured. Processing...`, green on a TTY) and `internal/ui/renderer_metrics.go` (`renderTurnHeader`: a leading blank, the 80-column `─` rule, `╭─⠿ Turn <N>[ - <mode>]`, the payload line, a trailing blank).
- **A2 (surface scope — locked)**: the chrome applies to (A) the positional/piped prompt turn and (B) the round-012 interactive plain reader; the `-i` TUI surface (round 016) is unchanged.
- **A3 (startup set — locked)**: only the input-capture line is added; the reference's `[Info] Starting chat...` line is out of scope this round.
- **A4 (turn number)**: `<N>` derives from the count of the session's **completed turns** + 1 (tellme persists **one `history_entry` line per completed turn** in `history.jsonl`; this is the reference's `SessionTurns + 1`, **not** a message count and **not** `entries/2`). A session archives only on `--new` (which resets the count, so it shows `Turn 1`); no mid-session archive path can shrink the count today — a future summarisation/archive path must preserve this.
- **A5 (chrome tokens)**: the exact tokens (the `─` rule width, the `╭─⠿` glyph, the header/blank-line layout, the colour codes and the TTY gate) are a `/axb-technical-research` + `/axb-dsl-refine` determination; the existing pre-flight payload line's **text** format is unchanged from round 009.
- **A6 (platform / dependency / verification)**: POSIX-only (round-012/015/016 precedent); no new dependency; verification is hermetic (injected streams + the clock seam; no pty).
- **A7 (post-turn — out of scope)**: the post-turn lines (the measured payload line, the metrics line, the final `╰─⠿ Ready` summary) are explicitly out of scope this round.
