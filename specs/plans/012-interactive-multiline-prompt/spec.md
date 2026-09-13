# Feature Specification: tellme Interactive Multi-line Prompt Capture (round 012)

**Feature Branch**: `012-interactive-multiline-prompt`

**Created**: 2026-09-13

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "tellme needs the interactive multi-line prompt capture that `tell-me-go` has" (`tell-me-go` prints `[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]`, then reads stdin to EOF). Today `tellme` has **no** interactive read: a bare `tellme` on a terminal prints the boot report, and stdin is read only when it is **not** a terminal (round 005). This round adds an interactive multi-line prompt reader: on a TTY with no prompt argument, `tellme` prints a hint and reads the operator's prompt until EOF (Ctrl+D), bounded and cancellable.

This round's behaviour intent is **ADD** (a new capability) that **MODIFIES** the bare no-prompt path on a terminal. It does not change the piped/positional prompt paths, the answer stream, or the payload-status/ordering contracts. **(Review-response amendment, PR #31 BLOCKER B1):** the round originally scoped "no new dependency"; the review showed the round-005 `os.ModeCharDevice` probe is unsound as a behaviour gate — it is true for `/dev/null`, so a redirected null device engaged the reader and masked a configuration failure as exit `0`. The probe is now a **real isatty** (`golang.org/x/term.IsTerminal` — already in the module graph transitively via glamour, promoted to a direct require; `go.sum` unchanged) recorded in **ADR 0003**. See NFR-001/NFR-004/SC-006/A6 below.

**Clarify Round 1 (2026-09-13)** resolved three high-impact decisions:

- **Q1 → Option 1 (invocation)**: a bare `tellme` with **no prompt argument and a TTY stdin** starts the reader. The boot report remains for non-TTY/piped no-prompt runs and explicit subcommands.
- **Q2 → Option 1 (hint stream)**: the hint is written to **`stderr`**, preserving tellme's byte-exact `stdout` contract.
- **Q3 → Option 1 (empty/cancel)**: an empty or cancelled submission sends **no request** and does not contact the provider.

**Per instruction — POSIX-only, no Windows variant**: the reader targets POSIX terminals (`Ctrl+D` to send); there is no Windows-specific hint or `Ctrl+Z`+Enter branch and no Windows build tag for it.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Composing a prompt across multiple lines (Priority: P1)

As an operator, I want to type a multi-line prompt at the terminal and send it with `Ctrl+D`, so I can compose a longer instruction without fighting shell quoting or a single-line readline.

**Why this priority**: It is the round's core capability and the reason the round exists; the empty/cancel handling (Story 2) only matters once the reader exists.

**Independent verification**: force the injected `isTTY` seam to report stdin as a terminal, feed a scripted multi-line stdin ending in EOF, and confirm the hint is written to `stderr` and the captured text becomes the prompt of exactly one reasoning turn whose answer is printed to `stdout`.

**Acceptance Scenarios**:

1. **Given** `tellme` is invoked with no positional prompt and stdin is a terminal, **When** the operator types one or more lines and sends EOF (`Ctrl+D`), **Then** the system prints the multi-line hint to `stderr` and performs one reasoning turn whose prompt is the typed text.
2. **Given** the same invocation, **When** the operator types a multi-line prompt, **Then** the captured prompt preserves the line breaks (a trailing newline trimmed) and `stdout` carries only the answer.

**Functional Requirements**:

- **FR-001**: When `tellme` resolves to an empty prompt, has no positional arguments, and stdin is a terminal, the system MUST print the multi-line hint to `stderr` and read the prompt from stdin until **EOF** (`Ctrl+D` on POSIX).
- **FR-002**: The interactive read MUST be **bounded** at a fixed 1 MiB cap (matching the round-005 stdin cap); content beyond the cap is not read.
- **FR-003**: The captured text, trimmed, MUST become the prompt of **one** reasoning turn — the identical downstream of a positional/piped prompt (one provider request; the answer printed to `stdout`).
- **FR-004**: The hint MUST be written to **`stderr`**; `stdout` MUST remain byte-exact (the hint MUST NOT appear on `stdout`).
- **FR-005**: The send trigger is **EOF** (`Ctrl+D`); the reader reads to EOF, not to a blank line, so line breaks inside the prompt are preserved.

**Non-Functional Requirements**:

- **NFR-001**: The reader MUST be deterministic and offline-testable through the injected terminal-detection seam and a scripted stdin (no pty; the probe is a real isatty via `golang.org/x/term` — ADR 0003, amending the original "no new dependency" intent).
- **NFR-002**: The reader MUST NOT buffer without bound — the 1 MiB cap MUST bound memory.

---

### User Story 2 - Cancelling or submitting an empty prompt (Priority: P2)

As an operator, I want to cancel the reader (`Ctrl+C`) or accidentally send nothing (immediate EOF) without triggering a provider request, so a stray keystroke never spends tokens or sends an unintended prompt.

**Why this priority**: It depends on Story 1 existing and defines the failure/empty contract that makes the reader safe and testable.

**Independent verification**: force the terminal seam and feed either an immediate EOF (no content) or a cancelled context, and confirm **no** provider request is made and no answer is printed.

**Acceptance Scenarios**:

1. **Given** the reader is reading, **When** the operator cancels with `Ctrl+C` (SIGINT), **Then** the run ends without a provider request and without printing an answer.
2. **Given** the reader is reading, **When** the operator sends EOF with no content (empty/whitespace only), **Then** the run ends without a provider request and without printing an answer.

**Functional Requirements**:

- **FR-006**: On `Ctrl+C` (SIGINT) during the interactive read, the system MUST cancel without sending a prompt and without contacting the provider.
- **FR-007**: On EOF with empty (or whitespace-only) content, the system MUST NOT send a prompt and MUST NOT contact the provider.

**Non-Functional Requirements**:

- **NFR-003**: The cancel/empty paths MUST be deterministic and offline-verifiable (no provider contact, no network).

---

### Edge Cases

- When stdin is a **pipe** (not a terminal), the round-005 behaviour is unchanged — the content is combined with any positional instruction; the interactive reader does **not** engage.
- When stdin is a **pipe with no content** and no prompt argument, the existing no-prompt **boot report** path is unchanged (the reader does not engage on a non-terminal).
- When a **positional prompt** is given, the reader does **not** engage (the argument is the prompt).
- When a **non-prompt dispatch path** runs (`--version`, `-d`, `-l`, a prompt-less `--new`), the reader does **not** engage; those paths are unchanged.
- When the operator types more than the 1 MiB cap, the read is truncated at the cap (matching round 005).
- When the reader is engaged but stdin reaches EOF immediately with no content, the run cancels (FR-007), it does not fall through to the boot report.
- **No Windows variant**: there is no Windows hint/`Ctrl+Z`+Enter branch; the interactive path is a POSIX-terminal capability.

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-008**: The interactive reader MUST engage **only** when the resolved prompt is empty, there are **no positional arguments and no piped stdin**, and stdin is a terminal. The round-005 argument/pipe behaviour MUST be unchanged.
- **FR-009**: The non-prompt dispatch paths (`--version`, `-d`, `-l`, and a prompt-less `--new`) MUST be unchanged and MUST NOT enter the reader.
- **FR-010**: The hint MUST NOT carry the reserved `tellme: ` class-phrase prefix; the frozen class-phrase vocabulary is unchanged.
- **FR-011**: The round MUST NOT regress rounds 001–011; `stdout` MUST remain byte-exact and all prior acceptance scenarios MUST stay green.

#### Non-Functional Requirements

- **NFR-004**: The round MUST NOT add a new *module* (`go.sum` unchanged) — it reuses the bounded stdin read, and (review-response amendment, PR #31 B1) the terminal probe is a **real isatty** (`golang.org/x/term`, already in the module graph transitively via glamour; `go.mod` promotes it from indirect to direct — ADR 0003), superseding the original "dependency-free `os.ModeCharDevice` seam" wording.
- **NFR-005**: All new assertions MUST be deterministic (no `time.Sleep`; the terminal seam + scripted stdin are the verification surface).

### Key Entities *(include if feature involves data)*

- **Interactive multi-line prompt**: the operator-entered text captured from a terminal until EOF; the prompt for one reasoning turn.
- **Multi-line hint**: the `stderr` line announcing that the reader is active and how to send/cancel.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With stdin reported as a terminal and no prompt argument, 100% of runs print the multi-line hint to `stderr` and read the prompt to EOF, performing one reasoning turn on the captured text.
- **SC-002**: 100% of cancelled or empty interactive submissions make **no** provider request and print **no** answer.
- **SC-003**: In every interactive run, `stdout` carries only the answer (the hint is on `stderr`); `stdout` is byte-exact.
- **SC-004**: The piped-input, positional-prompt, and non-prompt dispatch paths are byte-identical to rounds 001–011; all prior acceptance scenarios remain green.
- **SC-005**: The interactive capture is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.
- **SC-006**: No new *module* is introduced (`go.sum` unchanged); the terminal probe (`golang.org/x/term`) was already in the module graph and is promoted to a direct require (ADR 0003 — review-response amendment of the original "no new dependency" criterion).

## Assumptions

- **POSIX-only (A1)**: the reader targets POSIX terminals (`Ctrl+D` to send); there is no Windows variant (no `Ctrl+Z`+Enter branch, no Windows build tag).
- **Hint wording (A2)**: the hint mirrors the reference's POSIX line: `[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]`.
- **Send = EOF (A3)**: the read runs to EOF, not to a blank line; a trailing newline is trimmed.
- **Bound (A4)**: the interactive read is capped at 1 MiB (rounds 005/006 stdin cap).
- **Combination (A5)**: the reader engages only with no positional prompt and no piped stdin; combined prompt sources are unchanged from round 005.
- **Terminal probe (A6)**: ~~reuse the dependency-free TTY seam (`os.ModeCharDevice`) + `io.LimitReader`/context handling — not `golang.org/x/term`~~ — **SUPERSEDED (PR #31 review BLOCKER B1)**: the probe is a **real isatty** (`golang.org/x/term.IsTerminal`), because `os.ModeCharDevice` is true for non-terminal character devices (`/dev/null`) and masked a configuration failure as exit `0`. The bounded read still uses `io.LimitReader`/context handling. See ADR 0003.
- **Cancel semantics (A7)**: `Ctrl+C` (SIGINT) cancels the read without sending a prompt (consistent with rounds 004/005 SIGINT handling).
- Single-turn execution (round 004), stdin piping (round 005), rendered/raw output (round 006), durable history (round 007), the agent tool loop (round 008), the payload status line (round 009), cross-stream ordering (round 010), and persona-on-the-wire + wire-faithful estimate (round 011) are unchanged. The round adds **no** streaming, pinning, `-b`/`--retry`, pruning, MCP, memory, or TUI (no Bubble Tea prompt mode).
