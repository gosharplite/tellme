# Feature Specification: spinner liveness while a command streams + a per-turn elapsed timer (round 040)

**Feature Branch**: `040-spinner-liveness-and-turn-timer`

**Created**: 2026-09-17

**Status**: Draft — two folded spinner workstreams. WS-B's reset boundary is **operator-locked**; WS-A's mechanism and the remaining format details are **proposed** (pending operator confirmation at the pre-`/axb-tasks` review gate).

**Input**: Two spinner workstreams folded into one round (both touch `internal/ui/spinner.go` and `presenting-the-progress-spinner.feature` — the round-038/039 fold precedent):

- **Issue [#82](https://github.com/gosharplite/tellme/issues/82)** — *spinner liveness while a command's `[Tool Output]` streams*. Today the round-019/025 spinner is **paused for the whole live output block** (round-034 FR-012/G8, ADR 0005 D7): the block owns `stderr` alone from `Begin` (before the child starts) to `End` (after the closing separator). On a **quiet/slow** command (a line, then a long computation with no output) there is **no animation at all** in that window, so the operator cannot tell the run is still alive.
- **Issue [#83](https://github.com/gosharplite/tellme/issues/83)** — *a dual elapsed timer*. The spinner shows one figure, whole seconds since the **prompt** was captured (turn-scoped, never resets — round-019 D4): `⠋ Thinking [<model>]... (120s)`. The operator wants a **second** figure: the **current turn's** own duration, reset at each AI-endpoint call: `⠋ Thinking [<model>]... (120s 21s)`.

## Operator-locked / proposed decisions (clarify)

### Workstream B — the dual elapsed timer (issue #83)

- **QB1 → Reset boundary = per AI-endpoint call (option A).** The new figure resets at the start of each **AI-endpoint call / inference round** — matching the round-027 turn header (`╭─⠿ Turn <N>` counts AI-endpoint calls). Rejected: per waiting phase (B), per tool execution (C). *Mechanically clean:* `AgentLoop.Run` fires `notifyInferenceStart()` **exactly once per AI-endpoint call**, and the spinner already reacts to `OnInferenceStart`, so resetting the second timer there is per-call **by construction** with **no loop change**. *(Operator-locked 2026-09-17.)*
- **QB2 → Both phases carry both figures.** The second figure appears in the model-wait label and in the tool-execution label (which also keeps its ` [CPU: x% | MEM: y%]` segment). *(Proposed.)*
- **QB3 → Format `({total}s {turn}s)`** — one ASCII space between the two figures, both `s`-suffixed, inside the existing single parentheses. *(Proposed.)*

### Workstream A — liveness while the output block streams (issue #82)

- **QA1 → Mechanism: idle-gap resume.** Keep the single-writer `stderr` ownership, but do not stay hidden for the **whole** block: after **N** seconds with **no new output line**, the indicator resumes; the next output line clears it again. (Alternatives considered and rejected in the issue: a permanent output-progress marker; a safe per-line yield — mechanically hardest.) *(Proposed — the mechanism is the largest remaining decision.)*
- **QA2 → The idle threshold N is a small fixed value** (assume **3 s**, the round-019 research's fast-fail precedent) with a hermetic env seam (mirroring round-016's `TELL_ME_TUI_DEBOUNCE=0`) so E2E can force it. *(Proposed.)*
- **QA3 → Scope guard.** Only the tool-execution phase of a **streaming `[Tool Output]` block** is affected; the model-wait phase, the block literals, the bounded-and-stopped semantics, the non-TTY/`-r` gates, and the `-i` surface are unchanged. *(Proposed.)*

> **These decisions shape the acceptance criteria below.** QB1 is locked; QA1–QA3 and QB2/QB3 are the round's `[NEEDS CLARIFICATION]`-class items, disclosed as assumptions and to be confirmed by the operator at the pre-`/axb-tasks` review gate.

**Scope note**: two `stderr`-presentation changes to the round-019/025 progress spinner. Neither changes a CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record, the tool **result** fed to the model, or a `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A long-running command still shows a live indicator (Priority: P1)

As an operator, when a shell command runs for a while and streams its `[Tool Output]` block, I want the progress spinner to reappear during quiet stretches, so that I can tell the run is still alive instead of staring at a static header.

**Why this priority**: it is issue #82 — the anchor; the pause is the anchor's stated liveness gap.

**Independent verification**: unit-pin the spin cycle's resume-after-idle / clear-on-next-line behaviour under an injected `newTicker` seam (no `time.Sleep`); script an E2E run whose command prints a line, then is quiet past the idle threshold, and assert a spinner frame is drawn **within the block** on the captured `stderr`.

**Acceptance Scenarios**:

1. **Given** a terminal whose diagnostics are shown, **When** a command streams an output line and then produces no further output for at least the idle threshold, **Then** the run shows the progress spinner again during that quiet stretch, and the next output line clears it before being written.
2. **Given** the same run, **When** the block closes, **Then** the indicator leaves **no** residue and the turn-scoped total elapsed is preserved.

**Functional Requirements**:

- **FR-001**: While a `[Tool Output]` block is open, the indicator MUST NOT remain hidden for the whole block: after **no new output line for the idle threshold N**, the indicator MUST reappear; when the next output line arrives, the indicator MUST be cleared synchronously **before** that line is written.
- **FR-002**: The resume/clear MUST be **single-writer**-safe — it MUST NOT interleave a spinner frame with a partial output line and MUST NOT drop or reorder any output byte; exactly one writer owns `stderr` at any instant.
- **FR-003**: The resumed indicator MUST keep the **turn-scoped total** elapsed (round-019 D4) and MUST leave **no residue** on block close (round-019/025/035).
- **FR-004**: The block literals (header `... (Output shown below)` + the 60-hyphen separator), the bounded-and-stopped semantics (round-024), and the non-TTY / `-r` gates MUST be unchanged.

### User Story 2 - The spinner shows the total time and the current turn's time (Priority: P2)

As an operator, I want the spinner to show both how long ago I submitted my prompt and how long the current turn has been running, so that I can separate a slow single call from a slow series of calls.

**Why this priority**: it is issue #83, an additive readability change to the same line; independent of US1.

**Independent verification**: unit-pin `FormatSpinnerLine` / the two-epoch arithmetic under an injected `now` seam; script a multi-call tool turn and assert the two figures in the captured frames (the total never decreases; the turn figure drops at each new AI-endpoint call).

**Acceptance Scenarios**:

1. **Given** a prompt that makes one AI-endpoint call, **When** the spinner renders, **Then** the line shows two whole-second figures — the total since prompt capture and the current turn's duration — and for a single-call turn the two are approximately equal.
2. **Given** a prompt that makes several AI-endpoint calls, **When** each call begins, **Then** the second figure **resets** to that call's own duration while the first keeps growing from the prompt epoch.

**Functional Requirements**:

- **FR-005**: The spinner line MUST carry **two** whole-second figures inside the existing parentheses: the **total** since prompt capture first, then the **current turn's** duration — `({total}s {turn}s)`.
- **FR-006**: The **total** figure MUST measure whole seconds since the turn's prompt-capture epoch and MUST **not** reset within a prompt (round-019 D4 preserved).
- **FR-007**: The **turn** figure MUST measure whole seconds since the **current AI-endpoint call** began and MUST **reset at the start of each AI-endpoint call** (the per-call boundary of QB1; round-027 AI-call semantics).
- **FR-008**: Both figures MUST appear in the model-wait label **and** the tool-execution label; the tool-phase resource segment ` [CPU: <c>% | MEM: <m>%]` MUST be unchanged.

### Edge cases

- **A tool-less turn** (exactly one AI-endpoint call) → the two figures are approximately equal; no special-casing.
- **A quiet streaming command** (WS-A) → the indicator reappears after the idle threshold; the total keeps growing and the turn figure keeps counting **within its call**.
- **The AI-endpoint boundary coinciding with a resume** → the turn figure resets on the next call's `OnInferenceStart`; the total is unaffected.
- **A very long turn** (both figures reach 3+ digits) → the line is longer, so the round-025 **rune-based** `rowsForLine`/`eraseRows` bound MUST still erase every occupied row (no residue).
- **A terminal resize mid-frame** → the round-025 recorded limitation (a stale row count may over-erase one row) is unchanged and out of scope.
- **Non-terminal `stderr`, `-r`, or the `-i` submit surface** → the spinner gate is unchanged; `-i`/`-r`/non-TTY behaviour is untouched.
- **A provider that reports no usage** → the spinner is orthogonal (the spinner is not the post-turn status group); unchanged.

### Key entities

- **Spinner line** — one rendered frame: `{braille}{status} ({total}s {turn}s){resource}`.
- **Total elapsed epoch** — the turn's **prompt-capture** time; drives the first figure; never resets.
- **Turn elapsed epoch** — the **current AI-endpoint call's** start time; drives the second figure; reset per call.
- **`[Tool Output]` block** — the live streamed shell output; the spinner must resume/clear around its quiet stretches (WS-A).

## Requirements *(mandatory)*

### Global requirements

- **FR-009**: The round MUST NOT change any CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record shape (`history.jsonl` / `tokens.log` / `global_prompts.jsonl` / `tools-count.jsonl`), the tool **result** fed to the model, or `stdout` bytes.
- **FR-010**: The offline paths (`--version`, `-d`, `-l`, `--tool-usage`) MUST remain unchanged.
- **FR-011**: The spinner MUST remain gated to `isatty(stderr) && !-r`; a non-terminal diagnostic stream MUST draw no spinner, and the WS-A resume/clear MUST be a no-op when the spinner is gated off.

### Out of scope (recorded)

- Per-line spinner redraw inside the block (G8's stated blocker). WS-A resumes on an **idle gap**, not per line.
- Any change to the block literals, the bounded-and-stopped semantics, `output_file` handling, or the `-i` TUI surface.
- The reference divergence: `tell-me-go` shows no turn-scoped dual timer and does not re-activate a spinner inside its streamed output — both are recorded divergences.

## Success criteria *(mandatory)*

- **SC-001**: Unit pins show the spinner reappears after the idle threshold and clears on the next output line, with no residue and the total elapsed preserved (US1), and that the two-figure arithmetic resets the turn figure per AI-endpoint call while the total never decreases (US2).
- **SC-002**: A scripted E2E run with a quiet-then-resuming command shows a spinner frame **inside** the block window on the captured `stderr` (US1).
- **SC-003**: A scripted E2E multi-call turn shows both figures, with the total monotonic and the turn figure dropping at each call boundary (US2).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit (re-run after the truth MODIFY) are green.
- **SC-005**: The change has **falsifiability witnesses**: removing the idle-gap resume fails the WS-A Example; freezing the turn figure (no per-call reset) fails the WS-B Example; widening the line past the terminal width without the row-aware clear strands a residue frame — each reproduced then reverted.

## Assumptions

- **Truth ownership.** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md` are owned by `/axb-dsl-refine`; `techstack.md` by `/axb-technical-research`; the round's new **ADR (supersedes ADR 0005's D7 pause semantics)** is governance, not truth. `/axb-api-plan` is a checked **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted-state change). All edits are recorded in this round's `truth-delta.md`.
- **Supersession.** Round-034 FR-012/G8 and **ADR 0005 D7** (the whole-block pause) are `Accepted`/frozen → WS-A's change is recorded via a **new ADR that supersedes D7** (an `Accepted` ADR is immutable but for its `Status` line; the ADR 0008→0007 precedent). Round-019's D4 ("the elapsed epoch … never reset") is **amended** by WS-B → also recorded in the new ADR.
- **Frozen history.** Every earlier plan package (incl. `019-turn-spinner`, `025-spinner-width-safety`, `034-tool-call-log-parity`, `035-spinner-tail-residue`) is never edited; 040 is a fresh package.
- **Enabler.** The idle threshold exposes a hermetic seam so E2E can force it without wall-clock sleeps (round-016 `TELL_ME_TUI_DEBOUNCE` precedent).
