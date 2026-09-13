# Phase 0 Research: tellme Interactive Multi-line Prompt Capture (Round 012)

Topic: add the reference's **interactive multi-line prompt reader**. `tell-me-go` prints `[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]` and `io.ReadAll`s stdin to EOF when no prompt argument is given on a terminal. `tellme` has no such path: a bare `tellme` on a TTY prints the boot report, and stdin is read only when it is **not** a terminal (round 005). This round reads a multi-line prompt from a terminal, bounded and cancellable, and turns it into the same single reasoning turn as a positional/piped prompt. **No new dependency**; **POSIX-only (no Windows variant)**.

Scope note: the language, module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transport, rendering, session history, the agent tool loop (round 008), the payload status line (round 009), cross-stream ordering (round 010), and persona/estimator (round 011) were locked in rounds 001–011. The system still has **one CLI end**, adds **no new system end**, **no new external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. Clarify Round 1 (in `spec.md`) locked the round's three high-impact decisions: Q1 = bare no-prompt TTY starts the reader; Q2 = hint to `stderr`; Q3 = empty/cancel sends no request. **IN**: the interactive read, the hint, and the empty/cancel contract. **OUT**: any Windows variant; any new dependency; any change to the pipe/positional paths, the answer stream, or the status/ordering contracts.

---

## Decision 1: Reuse the existing dependency-free terminal seam — no `golang.org/x/term`

- **Decision**: The interactive reader keys off tellme's existing terminal-detection seam (`os.ModeCharDevice` on the stream, behind the injected `isTTY` function in `runtimeEnv`). `golang.org/x/term` (used by the reference for `MakeRaw`/`TerminalLock`) is **not** adopted.
- **Rationale**: tellme's round-005 seam already answers "is stdin a terminal?", which is the only TTY fact this reader needs. The read itself is a plain `bufio`/`io.ReadAll` over stdin — no raw-mode or terminal-size API is required. Keeps `go.mod` unchanged and is directly unit-testable via the seam.
- **Alternatives considered**:
  - **`golang.org/x/term`** — a new dependency for a capability (raw mode) tellme does not need — rejected.
  - **A pty harness/library** — forsworn (round 005 grill Q6; round 006 Q2) — rejected.

## Decision 2: Read stdin to EOF, bounded at 1 MiB

- **Decision**: The reader reads stdin **to EOF** with `io.ReadAll(io.LimitReader(stdin, 1 MiB))` (the round-005 cap), preserving interior line breaks and trimming a trailing newline.
- **Rationale**: On a POSIX terminal, `Ctrl+D` generates EOF, exactly as the reference relies on; reading to EOF (not a blank line) is what lets a multi-line prompt contain blank lines. The 1 MiB cap matches the existing stdin bound and bounds memory.
- **Alternatives considered**:
  - **A blank-line terminator** — cannot express an intentional blank line and diverges from the reference — rejected.
  - **Unbounded read** — memory-exhaustion risk — rejected.

## Decision 3: The hint is written to `stderr`, with no class prefix

- **Decision**: The hint (`[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]`) is written to **`stderr`** and carries **no** `tellme: ` prefix. `stdout` stays byte-exact.
- **Rationale**: tellme keeps `stdout` byte-exact and writes diagnostics to `stderr` (rounds 005/006, 009/010). The reference prints the hint to stdout because it has no such contract; tellme must not. (Clarify Q2 → 1.)
- **Alternatives considered**:
  - **stdout** (reference parity) — pollutes the answer stream — rejected.

## Decision 4: Empty / cancelled submissions send no request

- **Decision**: On `Ctrl+C` (SIGINT) the read is cancelled; on EOF with empty/whitespace-only content the run is treated as a cancel. Both send **no** prompt and contact **no** provider, exiting without an answer.
- **Rationale**: Matches the reference (`TrimSpace(result) == "" → context.Canceled`) and makes the reader safe (a stray keystroke never spends tokens). Reuses tellme's existing SIGINT context (`signal.NotifyContext`). (Clarify Q3 → 1.)
- **Alternatives considered**:
  - **Empty → usage error (exit 2)** — treats "no input" as misuse; diverges from the reference — rejected.
  - **Empty → boot report** — surprising after the reader has announced itself — rejected.

## Decision 5: POSIX-only — no Windows variant

- **Decision**: The reader targets POSIX terminals (`Ctrl+D`). There is **no** Windows-specific hint or `Ctrl+Z`+Enter branch and **no** Windows build tag for it.
- **Rationale**: Per instruction; keeps the surface minimal. On Windows the interactive path is simply not provided (`tellme` stays arg/pipe-based there). This is a deliberate divergence from the reference's cross-platform branch.
- **Alternatives considered**:
  - **Mirror the reference's Windows `Ctrl+Z`+Enter` variant** — out of scope per instruction — rejected.

## Decision 6: Verification — unit seam for the positive path; E2E for the negative/unchanged paths

- **Decision**: The **positive** interactive read (hint to `stderr`, read to EOF, prompt used, cancel/empty) is verified at the **unit** layer in `internal/cli`, driving the injected `isTTY` seam to report stdin as a terminal and feeding a scripted `io.Reader`. The **E2E** black-box suite (godog, pty-less) asserts the **negative/unchanged** facts it *can* observe: a **piped** run prints **no** reading announcement and behaves exactly as round 005; positional/non-prompt paths are unchanged; `stdout` stays byte-exact.
- **Rationale**: tellme's E2E harness invokes the **built binary** as a subprocess, so it cannot inject the Go `isTTY` seam and cannot allocate a pty (a forsworn dependency). The exact interleave of the interactive read is therefore **not** E2E-observable — the same limitation round 005 accepted for its "stdout is a terminal" branch (a **named pin**). Asserting the positive path at the unit layer (where the seam *is* injectable) keeps it truly verified; asserting the observable negatives E2E keeps the black-box oracle honest.
- **Alternatives considered**:
  - **A test-only env seam (`TELL_ME_FORCE_STDIN_TTY`)** so the harness can force the TTY branch over a pipe — would make the positive path E2E-drivable, but adds a product env knob purely for tests (no precedent in tellme) — **held as a fallback**, to revisit only if unit coverage proves insufficient.
  - **A pty harness** — forsworn — rejected.
  - **Leaving the positive path unasserted** — unacceptable — rejected.

## Decision 7: No new dependency; BDD techstack & strategy unchanged

- **Decision**: No third-party dependency (`go.mod` / `go.sum` unchanged); the BDD techstack (`godog` over the built binary) and the strategy (E2E acceptance + fast unit tests) are unchanged. The round adds a **`[UNIT]`** assertion set for the interactive path (Decision 6) plus the E2E negative assertions.
- **Rationale**: The read is plain stdlib; the seam exists.
- **Alternatives considered**:
  - **A new runner/framework** — no need — rejected.

## Decision 8: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog`; strategy = E2E + pure-helper units — unchanged, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round extends an existing CLI end's input handling; it introduces no new interface, service, or framework.
- **Alternatives considered**:
  - **Re-open the techstack questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **Positive-path observability (Decision 6)**: the interactive read is unit-verified, not E2E-verified; the executable interface truth pins the *observable negatives* (no announcement on a pipe; unchanged paths) and documents the interactive read as a **named pin** (mirroring round 005's pty pin). The `TELL_ME_FORCE_STDIN_TTY` env seam is the fallback if unit coverage proves insufficient.
- **Hint colour/gating**: whether the hint is colourised when `stderr` is a terminal (and how that interacts with the byte-exactness gate) is a `/axb-dsl-refine`/implementation determination.
- **Empty-EOF exit code**: the exit code for a cancelled/empty interactive run (success `0`, matching the reference) is to be pinned in the interface truth — a `/axb-dsl-refine` determination.
- **Dispatch insertion point**: the reader engages on the no-prompt path only when stdin is a terminal and no terminal-less subcommand (`--version`/`-d`/`-l`/prompt-less `--new`) is active — an implementation detail (round-005 dispatch order).
- **Deferred, still out of scope**: a Windows variant; a TUI/Bubble Tea prompt mode; streaming; pinning; `-b`/`--retry`; pruning; MCP; memory.
