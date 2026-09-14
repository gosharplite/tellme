# Phase 0 Research: turn spinner — a live progress indicator for prompt turns (round 019)

## Decision 1: A hand-written `internal/ui` spinner (no dependency)

- **Decision**: implement the spinner as hand-written Go in `internal/ui` — a pure formatter plus a small ticker-driven presenter, mirroring the reference's `renderer_spinner.go` — alongside the round-009/017/018 helper package. It reuses the injected clock seam (a `Ticker`) and the injected `stderr` writer; no third-party spinner or TUI library is used on the non-TUI path.
- **Rationale**: keeps the non-TUI path dependency-light (the Bubble Tea family is loaded only for `-i`); the reference hand-rolls its spinner; a hand-written presenter reuses the existing clock/stream seams, keeping the round deterministic and `make verify`-clean.
- **Alternatives considered**:
  - `bubbles/spinner` — pulls the Bubble Tea model/event loop into the plain path; unnecessary for a single stderr line and couples the non-TTY path to the TUI stack.
  - A third-party spinner library (e.g. `briandowns/spinner`) — a new dependency for one line.

## Decision 2: Frames, cadence, and the drawing primitive

- **Decision**: a fixed braille frame set `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` advanced on a ~200 ms injected ticker; each frame is drawn to `stderr` as `\r<clear-line>{frame}{status} ({n}s)`; the first frame is drawn **synchronously** on start (so even a sub-tick wait shows a frame); clearing writes `\r<clear-line>` with **no** trailing newline, so the answer starts on the same line.
- **Rationale**: byte-for-byte the reference behaviour; a carriage-return line redraw is the minimal primitive needing no ANSI cursor addressing; the synchronous first frame removes the reference's documented 200 ms startup gap.
- **Alternatives considered**:
  - Full-screen cursor addressing (ANSI `\033[…]`) — unnecessary and TERM-dependent.
  - Newline-per-frame — would scroll and corrupt the transcript.

## Decision 3: Phase labels (identifiers interpolated)

- **Decision**: derive the label from the waiting phase — awaiting the model → ` Thinking [<model>]...` (the `[<model>]` segment omitted when the active provider names no model); executing tools → ` Executing [<tool>]...` for one tool, ` Executing tools [<a>, <b>]...` for several, ` Executing tools...` when names are unavailable. Every label carries a leading space (reference parity). `tellme` has no automatic context-compression phase, so the reference's ` Compressing context...` label is not reproduced (nor a retry label — no such phase this round).
- **Rationale**: the operator locked full reference-parity labels with identifiers (clarify Q1=2); `tellme`'s waiting phases are exactly inference and tool execution.
- **Alternatives considered**:
  - Fixed labels without identifiers (clarify Q1=1) — rejected by the operator.
  - Reproduce the summarisation/retry labels — no such phases exist in `tellme`.

## Decision 4: The elapsed counter

- **Decision**: the line renders `({n}s)` with `n` = whole seconds since the current spinner started; when a phase transition updates only the label in place, the counter is **not** reset; a fresh waiting interval after interleaved output restarts it.
- **Rationale**: matches the reference (a per-spinner `startTime`; `UpdateSpinnerStatus` preserves it); keeps a single wait's counter monotonic while the phase changes.
- **Alternatives considered**:
  - Reset on every label change — a single wait would look like it restarts, which reads as a hang reset.
  - Milliseconds — noise on a terminal line.

## Decision 5: Host CPU / memory sampling (machine-wide, dependency-free, POSIX-only)

- **Decision**: sample **machine-wide** host CPU and memory by hand, POSIX-only — CPU = the Δ of (`Σcpu − idle`) across the host's `/proc/stat` ticks; memory = host used/total percent; macOS via `sysctl`/mach host statistics (a **cgo** and a **nocgo** variant). Put the sampler behind a small domain **port**, with the OS adapters under `internal/infrastructure/telemetry/` (build-tagged `_linux` / `_darwin_cgo` / `_darwin_nocgo`), so unit tests inject fixed readings; render the ` [CPU: <c>% | MEM: <m>%]` segment only on the tool-execution spinner, `%.1f%%`.
- **Rationale**: the operator keeps the reference's metrics variant (clarify Q2=2) and locked Linux/macOS only. The reference samples **system-wide** (`internal/infrastructure/telemetry/system_metrics_linux.go` reads `/proc/stat`'s host `cpu ` line — commented *"host CPU stats … for better visibility of tool execution"*), so **machine-wide** is the parity target and matches tellme's contract wording ("the machine's resource usage"); a process-scoped sample would both break parity and contradict the contract. Placement behind a port in `internal/infrastructure/telemetry/` (never `internal/ui`, a pure formatter package) matches the reference and keeps `internal/ui` dependency-free.
- **Alternatives considered**:
  - **Process-scoped** (own CPU/RSS) — rejected: breaks parity and contradicts the "machine" wording (review B2).
  - Put the sampler in `internal/ui` — rejected: `internal/ui` is a pure formatter home; OS sampling belongs behind an infrastructure adapter (review TD1).
  - `shirou/gopsutil` — a large new dependency for two numbers; conflicts with NFR-005.
  - Drop the segment (clarify Q2=1) — rejected by the operator.

## Decision 6: Gating — the diagnostic-stream (`stderr`) terminal probe

- **Decision**: draw the spinner only when the **diagnostic stream (`stderr`)** is a terminal **and** `-r` is off. This mirrors the reference exactly: its `IsTerminalContext()` (`internal/ui/renderer.go`) reads **`ui.stderr`**, and the spinner is drawn to `stderr`; it uses **no** stdout probe. Drive the branch in E2E with a diagnostic `TELL_ME_FORCE_STDERR_TTY` seam (mirroring the round-012 `TELL_ME_FORCE_STDIN_TTY`) so it is testable without a pty. **This does not wire a standard-output probe, and does not close round-006 / PR #16 Obs 1** — Obs 1 (tellme's own *stdout* chrome) stays **OPEN**; wiring a stdout probe is neither necessary nor faithful for a `stderr` diagnostic.
- **Rationale**: the operator locked reference parity. Gating a **`stderr`** spinner on **`stderr`** is the correct, corruption-free behaviour: `tellme "hi" > out.txt` (stderr still a TTY) keeps the operator's feedback, while `tellme "hi" 2> err.txt` (stderr a file) draws **nothing** — an `isatty(stdout)` gate instead writes `\r` frame bytes into `err.txt` and drops feedback exactly when the operator is watching. The round-005/006 "presentation suppressed when standard output is not a terminal" contract governs the **answer's rendering / stdout chrome**, not a `stderr` diagnostic — the two are different gates (this reverses the earlier D6 draft).
- **Alternatives considered**:
  - Gate on **`stdout`** terminality — rejected (review B1): corrupts a redirected diagnostic and inverts the feedback; conflates the answer-rendering gate with the diagnostic gate.
  - Also require `isatty(stdout)` as an *additional* constraint — unnecessary; the surface conjunction (spec FR-008) plus the `stderr` gate already cover it, and adding it would re-introduce the feedback loss.
  - A pty harness — forsworn (round 005 grill Q6 / round 006 Q2); the forced `stderr` seam covers the branch.

## Decision 7: Lifecycle placement — a `LoopObserver` seam on `AgentLoop`

- **Decision**: the turn is a single blocking `AgentLoop.Run(ctx, prompt, prior)` call (`internal/cli/cli.go`), and inside it inference, tool execution, and the tool-trace log (`internal/agent/agentloop.go` `logStep` → `a.Stderr`) all happen — so `cli.go` cannot observe phase transitions from outside. Add a **`LoopObserver` port** (`internal/domain/agent/`) and make `AgentLoop` invoke it: `OnInferenceStart(model)` / `OnInferenceEnd()` around each `Gateway.Complete`, `OnToolsStart(names)` / `OnToolsEnd()` around the tool batch, and `BeforeToolLog()` / `AfterToolLog()` **around `logStep`** so the spinner is cleared before the trace line and restored after. The `internal/ui` spinner presenter implements the port; `cli.go` constructs it and injects it into the loop — the CLI *owns* the lifecycle (it builds/injects the observer), the loop *invokes* it. A deferred stop guarantees no residue on any exit path (incl. panics/failure).
- **Rationale**: without a seam on the loop, `cli.go` cannot switch the label to ` Executing tools …` nor clear the spinner before `logStep` writes to `stderr` — `FR-002` would be unimplementable within the task boundary (review B1). The observer keeps a single owner (the CLI) while letting the loop emit the timing.
- **Alternatives considered**:
  - Edit `cli.go` only, treating `loop.Run` as a black box — impossible: the phase transitions and the `logStep` write happen inside the loop.
  - Event-bus/actor lifecycle (the reference's `uiBridge` actor) — overkill for a single CLI turn with no concurrent producers.
  - A per-call spinner owned inside the agent loop — splits the lifecycle across layers.

## Decision 8: Verification (unit + E2E, no pty)

- **Decision**: unit tests for the pure helpers — the label/elapsed formatter, the `%.1f%%` metrics formatter, and the **machine-wide** CPU/memory samplers (in `internal/infrastructure/telemetry`) driven by injected readings (Linux/macOS parser tables); E2E through the existing harness by setting `TELL_ME_FORCE_STDERR_TTY` — assert the spinner frames appear on `stderr` while waiting, that the label names the model/tools, that the metrics segment appears only for tools, and that `stdout` stays byte-exact; assert the negatives (non-terminal `stderr`, `-r`, non-prompt paths, `-i`) draw none.
- **Rationale**: matches the round-011/014/018 pattern (unit pure helpers + E2E acceptance) and keeps `make verify` deterministic; the forced seam replaces the forsworn pty.
- **Alternatives considered**:
  - Assert via a pty — forsworn.
  - Snapshot the animated bytes exactly — brittle; assert presence/animation/labels by pattern instead.

## Decision 9: Platform and dependency footprint

- **Decision**: POSIX-only (Linux/macOS); **no** new third-party dependency (`go.mod`/`go.sum` unchanged).
- **Rationale**: the operator locked Linux/macOS only; the spinner, the sampling, and the probe all reuse already-present facilities.
- **Alternatives considered**: a Windows branch — out of scope.

## Decision 10: Synchronous clear + serialized `stderr` writes

- **Decision**: the spinner presenter owns one I/O mutex; every spinner frame update **and** every `stderr` write during an active turn are serialized through it, and `Stop()`/`Clear()` are **synchronous** — they block until the ticker goroutine has terminated and the final clear frame (`\r` + clear-line) has been written. This mirrors the reference (`renderer_spinner.go`'s `ioMu`).
- **Rationale**: a ~200 ms background redraw can otherwise interleave with a `logStep`/answer write and wipe part of it (review R1); a synchronous clear guarantees the following line starts from a clean column.
- **Alternatives considered**:
  - Fire-and-forget stop (cancel a context, return) — rejected: a final in-flight frame could land *after* the following write.
  - Leaving serialization to the caller — rejected: multiple writers (the spinner + the round-017/018 emitters) → one owner is safer.

## Must-ask questions (settled by existing truth — unchanged this round)

- **BDD techstack**: `godog` (existing `specs/truth/techstack.md`).
- **Test strategy**: E2E (black-box) for the acceptance path + pure-helper unit tests.
- **System ends**: a single CLI end.
