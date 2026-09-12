# Phase 0 Research: tellme Prompt Piping (Round 005)

Topic: the round-005 piping slice — feed the prompt through **standard input** (combined with the positional instruction) and make the **output contract** pipe-friendly, per Clarify Round 1 (Q1 piping only / defer `-r`; Q2 combine `args + "\n"` + piped stdin; Q3 adopt `tell-me-go`'s TTY-aware output contract).

Scope note: the language (`Go 1.26`), module, CLI flag layer, testing harness (`godog` + stdlib `testing`), and base tooling were locked in rounds 001–004. The system still has **one CLI end** (the operator terminal). The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the CLI Gherkin; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. This round adds **no new system end**.

The two items the spec deferred here — the TTY-detection mechanism and the stdin size cap — are settled below (Decisions 1–2).

---

## Decision 1: TTY detection = dependency-free stdlib char-device check, behind an injected seam

- **Decision**: Detect "is a terminal" with the standard library — `os.File.Stat()` and `(mode & os.ModeCharDevice) != 0` — wrapped behind a small injected function (a general stream probe, defaulting to the real check) so unit tests can drive it. Wired to **stdin** this round; the **stdout** probe is applied when presentation is introduced (Decision 5, review finding F2). No new module.
- **Rationale**: Matches round-004's minimal-dependency stance (stdlib over SDKs) and is correct for the cases that matter — an interactive TTY is a character device (so stdin is *not* read), while a pipe/FIFO and a regular-file redirect are *not* (so stdin *is* read and presentation is suppressed). Keeping the probe behind an injected seam makes the I/O-mode selection unit-testable without a real terminal (Decision 4).
- **Alternatives considered**:
  - **`golang.org/x/term` (`term.IsTerminal`)** — the reference's approach and more precise (ioctl-based); rejected for now because it is a **new direct dependency** not currently in `go.mod`, and the char-device check suffices for tellme's two modes.
  - **No TTY detection (assume piped)** — rejected: would read the terminal and hang interactive use (violates FR-004/FR-008).

## Decision 2: Bounded stdin read with a fixed 1 MiB cap

- **Decision**: When stdin is piped/redirected, read it with `io.ReadAll(io.LimitReader(stdin, 1<<20))` — a fixed **1 MiB** cap (matching the reference's `maxPromptSize`). Content beyond the cap is not read (truncate, no error).
- **Rationale**: Satisfies NFR-001 (a bounded read that cannot exhaust memory) and adopts the reference's constant, so piping behaves the same as `tell-me-go`. `io.LimitReader` caps the read without buffering the whole stream first.
- **Alternatives considered**:
  - **Unbounded `io.ReadAll`** — rejected: a runaway stream (e.g. `yes | tellme`) could exhaust memory.
  - **A different cap or an error on overflow** — rejected: no basis for a different value; the reference truncates silently, and a mid-stream error would be a poor pipeline experience.

## Decision 3: Prompt assembly — combine positional argument(s) and piped content

- **Decision**: Build the prompt as `strings.Join(args, " ")`; when the piped bytes are non-empty, append `"\n"` followed by the piped text; then `strings.TrimSpace`. If the result is empty **and** no positional argument was given, there is no prompt — the run falls through to the existing boot path (round-004 behavior).
- **Rationale**: Reproduces the reference's **main** chat path exactly (`internal/ui/capture.go`: `combined = prompt + "\n" + string(bytes)`) and Clarify Q2, enabling `cat file | tellme "instruction"`. Folding empty-to-boot preserves round 004's no-prompt offline behavior and avoids a spurious empty turn.
- **Alternatives considered**:
  - **Argument wins; stdin read only when the argument is empty** — the reference's *callback-worker* path; rejected (Clarify Q2) — it is not the main path and would drop piped content whenever an instruction is given.
  - **Concatenate without the `"\n"` delimiter** — rejected: loses the instruction-vs-body boundary and produces a run-on prompt.

## Decision 4: I/O-mode selection as an injected, unit-testable seam (folds issue #14)

- **Decision**: Refactor `cli.Run` into a thin entrypoint that injects `stdin io.Reader`, `stdout/stderr io.Writer`, and the TTY-detection seam (`cmd/tellme/main.go` supplies the real `os.*` and the real detector), and extract the pure helpers (`combinePrompt`, the input/output-mode computation). Cover **flag parsing + input/output-mode selection** with table-driven unit tests driven by fakes, complementing the E2E acceptance path.
- **Rationale**: Satisfies NFR-002 / SC-005 and folds **issue #14** (F9): the flag-parsing + two-axis I/O-mode matrix (TTY vs pipe × argument vs stdin) is exactly what a black-box E2E suite under-covers and what cheap unit tests cover well. The seam mirrors the round-003 `EnvLookupFunc` port pattern (dependency inversion so pure logic is testable without real OS state).
- **Alternatives considered**:
  - **Keep `Run(args, version)` and rely on E2E only** — rejected: leaves the mode matrix under-tested and leaves #14 open.
  - **A full `App` dependency-injection container** (as the reference has) — rejected as disproportionate to three flags and two I/O modes.

## Decision 5: Output contract = TTY-aware suppression, pinned (no observable change today)

- **Decision**: Emit the answer to **stdout only**, as plain text with exactly one trailing newline; never write the answer to stderr; and adopt the reference's rule that any presentation (color, spinner, rendering) is suppressed when stdout is not a terminal (`session_manager.go`: `UseColor = isTTY && !RawOutput`). Because tellme emits **no presentation yet**, the **stdout probe is not wired this round** — the `isTTY` seam gates the **stdin read** only (Decision 1); the stdout probe is applied when presentation is introduced (review finding F2). The observable bytes are therefore identical today, and the contract is pinned.
- **Rationale**: Satisfies FR-006/FR-007 and Clarify Q3 (adopt `tell-me-go`'s posture); it is forward-compatible with the future renderer without any present-day behaviour change.
- **Alternatives considered**:
  - **TTY-agnostic (always identical output)** — rejected (Clarify Q3): diverges from the reference's posture.
  - **Detect and strip ANSI after the fact** — rejected: nothing emits ANSI yet; gating at the source is cleaner than post-hoc stripping.

## Decision 6: No `-r` flag and no renderer this round; read stdin only on the prompt-turn path

- **Decision**: Do **not** add a `-r`/raw-output flag or any renderer; keep the flag surface `-c/--config`, `-d/--diagnostics`, `--version`. Read stdin **only** on the prompt-turn dispatch — `--version` and `-d` must never read stdin (FR-010), so the offline paths stay network-free and non-blocking.
- **Rationale**: tellme's output already equals the reference's `-r` (raw) output, so a literal `-r` would be a vacuous flag (Clarify Q1). Gating the stdin read on the turn path preserves round-004's offline guarantees (NFR-004) and prevents `-d` from blocking on a piped stream.
- **Alternatives considered**:
  - **Add `-r` now as a documented no-op** — rejected: a flag that does nothing invites confusion and a dead contract.
  - **Add a renderer plus `-r` this round** — rejected: a rendering capability belongs to its own slice with its own acceptance.

## Residual risks / forward links

- **Terminal detection is a proxy (Decision 1)**: the char-device check reports "terminal" for any character device, notably `/dev/null`. This is benign here: a subprocess with no stdin reads `/dev/null` (a char device) → treated as a terminal → stdin is not read → the existing no-prompt behavior is unchanged. Recorded honestly rather than claimed as exact.
- **E2E harness (Decisions 3–4)**: the current subprocess runner sets no `cmd.Stdin` (the child reads `/dev/null`), so it must be extended to inject a scripted stdin (a pipe built from a `strings.Reader`) for the piped scenarios. Implementation note for `/axb-tasks`; no truth impact beyond the testing rows.
- **Rendered-output parity (Decision 6)**: the reference's default renders Markdown and `-r` bypasses it; tellme ships neither yet. A future "rendered output + `-r`" slice closes that gap.
- **Exit codes / class phrases**: unchanged this round — the round-004 table `0/2/3/4/5/6` and all frozen phrases stand (FR-009/FR-012).
