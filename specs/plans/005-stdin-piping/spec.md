# Feature Specification: tellme Prompt Piping (round 005)

**Feature Branch**: `005-stdin-piping`

**Created**: 2026-09-12

**Status**: Draft

**Input**: User request "005 — piping / raw-output" narrowed by Clarify Round 1 (reissued, 2026-09-12), which verified the recommendation set against the `tell-me-go` reference before ratification:
- **Q1 -> Option 1**: Piping only; keep raw output; **defer `-r`**. `tellme` has no renderer, so its output today already equals `tell-me-go`'s `-r` (raw) output; a literal `-r` would be a no-op flag. Full output-rendering parity (rendered default + `-r`) belongs to a future slice.
- **Q2 -> Option 3**: **Combine** — the prompt is the positional argument(s) (joined by single spaces), then a newline, then the piped standard-input content. This matches `tell-me-go`'s main chat path (`internal/ui/capture.go`: `prompt + "\n" + stdin`).
- **Q3 -> Option 2**: Adopt `tell-me-go`'s **TTY-aware** output contract — presentation (color/spinner/rendering) is suppressed when standard output is not a terminal.

*(Deferred to `/axb-technical-research` as assumptions: the TTY-detection mechanism and the exact stdin size cap; a 1 MiB cap is assumed pending ratification. Not asked this round because no user-visible decision remains.)*

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pipe a prompt into tellme (Priority: P1)

As an operator or automation-script author, I want to feed content into tellme through standard input — optionally together with an instruction argument — so that I can run reasoning over piped data in ordinary shell workflows (for example `cat logs.txt | tellme "summarize"` or `git diff | tellme "review this"`).

**Why this priority**: This is the slice's core capability and the first thing that makes tellme composable in a pipeline. Without stdin capture there is no piping at all, so it is the first value to prove.

**Independent verification**: `printf 'hello' | tellme` (no argument) sends exactly one provider request whose prompt is `hello`; `printf 'body' | tellme "instr"` sends exactly one request whose prompt is `instr\nbody`. Verified end-to-end against the local fake provider.

**Acceptance Scenarios**:

1. **Given** piped standard input carrying text and no positional prompt argument, **When** I run tellme, **Then** tellme treats the piped text (trimmed) as the prompt, sends exactly one provider request carrying it, prints the answer, and exits successfully.
2. **Given** a positional instruction together with piped standard input, **When** I run tellme, **Then** tellme's prompt is the instruction, then a newline, then the piped text, sent in exactly one provider request.

**Functional Requirements**:

- **FR-001**: When standard input is not a terminal (piped or redirected), the system MUST read standard input and use it as prompt content, bounded by a fixed maximum size.
- **FR-002**: When standard input is piped and a positional argument is also present, the system MUST form the prompt as the positional argument, followed by a newline, followed by the piped content.
- **FR-003**: When several positional arguments are present, the system MUST join them with single spaces before combining them with piped content.
- **FR-004**: When standard input is a terminal (interactive), or when no standard input is available, the system MUST NOT read standard input and MUST use the positional argument alone — preserving the round-004 behavior.
- **FR-005**: The system MUST trim surrounding whitespace from the final prompt before sending it.

**Non-Functional Requirements**:

- **NFR-001**: Reading standard input MUST be bounded by a fixed maximum so that an unbounded stream cannot exhaust memory.

---

### User Story 2 - Pipe-friendly, TTY-aware output (Priority: P2)

As an automation-script author, I want tellme's output to be clean and deterministic when it is piped or redirected, so that downstream tools consume exactly the answer text and the exit code, without terminal decoration.

**Why this priority**: Piping in is only useful if piping out is trustworthy. This pins tellme to `tell-me-go`'s TTY-aware posture — presentation is suppressed when standard output is not a terminal — and guarantees the process never blocks waiting for a terminal.

**Independent verification**: run a turn with standard output redirected; confirm the answer is the only thing on standard output (the answer bytes verbatim, plus a single CLI-appended terminating newline), carries no system-added presentation escape sequences, that any failure phrase goes to standard error only, and that the run never blocks waiting for terminal input.

**Acceptance Scenarios**:

1. **Given** a successful turn whose standard output is not a terminal, **When** the turn completes, **Then** the answer bytes appear verbatim on standard output alone, followed by a single CLI-appended terminating newline and no system-added presentation escape sequences, and the process exits with the success code.
2. **Given** a turn is run with piped or redirected standard input, **When** the run ends, **Then** the process does not block waiting for terminal input, and any failure is reported on standard error (never on standard output) with the frozen class phrase.

**Functional Requirements**:

- **FR-006**: The system MUST write the answer text to standard output only — the answer text **verbatim**, followed by exactly one terminating newline **appended by the CLI**; the answer MUST NOT be written to standard error. (The answer's own bytes — including any trailing or embedded newline it carries — are passed through unchanged; only the single terminating newline is added.)
- **FR-007**: When standard output is not a terminal, the system MUST suppress presentation output (color, spinner, and any rendered decoration) so piped output is plain and byte-exact; presentation MAY be used only when standard output is a terminal.
- **FR-008**: When standard input is not a terminal, the system MUST NOT wait for interactive terminal input before running the turn (it MUST NOT block for a TTY).
- **FR-009**: Failure reporting MUST keep the existing contract — the frozen class phrase on standard error and the existing distinct exit codes — with no change to the round-001 through round-004 codes and class phrases.

**Non-Functional Requirements**:

- **NFR-002**: The flag-parsing and input/output-mode-selection logic MUST be covered by fast, isolated unit tests (folding issue #14 / F9), complementing the black-box end-to-end path.
- **NFR-003**: For a given provider answer, the bytes written to the **redirected (non-terminal)** standard output MUST be identical across repeated runs. Terminal presentation (FR-007) is deliberately **outside** the byte-determinism contract: the redirected stream is the pipeline-consumed stream, and terminal output is a presentation zone. (The two requirements therefore cover **disjoint** streams — non-terminal vs terminal.)

---

### Edge Cases

- When standard input is piped but empty **and** no positional argument is supplied, the system MUST keep its existing no-prompt behavior (boot) and MUST NOT contact any provider.
- When standard input is piped but empty **and** a positional argument is supplied, the prompt MUST be the positional argument alone (trimmed), with no extra newline.
- When an explicit mode flag (`--version` or `-d`) is present together with piped standard input, the explicit mode MUST take precedence, standard input MUST NOT be read, and no provider request is made.
- When the piped content is very large, the read MUST be bounded (NFR-001) and MUST NOT exhaust memory.
- When standard input is a terminal and no positional argument is supplied, the system MUST keep its existing no-prompt behavior (boot) and MUST NOT block reading the terminal.
- When several positional arguments are supplied together with piped content, they MUST be joined before combination (FR-003), not concatenated without separators.

## Requirements *(mandatory)*

> Story-specific FR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-010**: Explicit modes (`--version`, `-d`) MUST take precedence over piped standard input and over any positional prompt; in those modes standard input MUST NOT be read and no provider request is made.
- **FR-011**: The system MUST keep its existing flag surface (`-c/--config`, `-d/--diagnostics`, `--version`) and MUST NOT introduce new flags in this round — notably, there is no `-r`/raw-output flag (Clarify Q1).
- **FR-012**: The round MUST NOT regress existing round-001 through round-004 behavior; all pre-existing end-to-end scenarios MUST remain green.

#### Non-Functional Requirements

- **NFR-004**: The offline paths (`--version`, `-d`, and a prompt-less boot) MUST remain strictly offline and deterministic with zero network calls; the piped prompt turn follows the round-004 rule and may dial the provider. In particular, an empty piped stdin with no argument MUST remain on the offline boot path with zero network calls.

### Key Entities *(include if feature involves data)*

- **Prompt**: The final prompt string assembled from the positional argument(s) and/or piped standard-input content (FR-002/FR-003/FR-005).
- **Input source mode**: Whether standard input is a terminal (interactive — not read) or piped/redirected (read, bounded) — the input-side branch (FR-001/FR-004).
- **Output stream mode**: Whether standard output is a terminal (presentation permitted) or redirected (presentation suppressed) — the output-side branch (FR-006/FR-007).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of piped-input runs — with and without a positional argument — form the prompt per FR-002/FR-003 and send exactly one provider request, exiting with code `0`.
- **SC-002**: In the acceptance set, 100% of redirected-standard-output runs place the answer bytes verbatim on standard output only, followed by a single CLI-appended terminating newline and no system-added escape sequences, exiting with code `0`.
- **SC-003**: In the acceptance set, 100% of offline/no-prompt runs (empty piped stdin with no argument; or an explicit mode flag with piped stdin) make zero provider requests and preserve the existing exit code.
- **SC-004**: All pre-existing round-001 through round-004 acceptance scenarios remain green.
- **SC-005**: The flag-parsing and input/output-mode-selection logic is covered by isolated unit tests (issue #14), and the piped prompt path is verified end-to-end against the local fake provider with no external network access.

## Assumptions

- Round 005 delivers stdin piping and the TTY-aware output contract only. It does **not** add a renderer or the `-r` flag (Clarify Q1); `tellme`'s output already equals `tell-me-go`'s `-r` (raw) output. Full output-rendering parity is deferred to a future slice.
- The prompt-combination rule follows `tell-me-go`'s main chat path: positional argument(s) joined by single spaces, then `"\n"`, then piped content (Clarify Q2). (`tell-me-go`'s callback-worker path — argument wins, stdin only when the argument is empty — is not adopted here.)
- Output is TTY-aware in `tell-me-go`'s sense: presentation is suppressed when standard output is not a terminal (Clarify Q3). Since `tellme` has no color/spinner/rendering yet, the observable bytes are unchanged today, but the contract is pinned for when such presentation is added.
- The TTY-detection mechanism and the exact stdin size cap are technical decisions owned by `/axb-technical-research`; a 1 MiB stdin cap is assumed pending ratification.
- **Verification boundary (this round)**: the harness exercises the input/output branch *logic* through stand-ins (`/dev/null` — a character device, standing for "stdin is a terminal → not read"; a pipe standing for redirected stdout), but it **cannot exercise a real TTY** — a pty is a new dependency this round forswears, and no renderer exists yet. Therefore the stdout-is-a-terminal branch of `FR-007` (presentation shown at a terminal) and real-pty stdin fidelity are **unverifiable this round** — a **named pin, not a verified contract**. A pty-capable harness is deferred to a future dependency-authorized round.
- No persistence, history, or session state is introduced (round-004 Clarify Q1 stands); this round adds no multi-turn loop and no streaming.
- The prompt content is not otherwise transformed beyond the join, newline delimiter, and trimming defined above (the shell performs any expansion before `tellme` receives it).
