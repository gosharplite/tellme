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

## Decision 5: Host CPU / memory sampling (dependency-free, POSIX-only)

- **Decision**: sample the **current process's** CPU percentage (from cumulative CPU-time deltas — not the whole-machine tick) and memory percentage (RSS against total physical memory) by hand, POSIX-only: Linux via `/proc/self/stat` (+ `/proc/meminfo`); macOS via `sysctl` / `mach` host statistics. Expose the sampling behind a small interface so unit tests inject fixed readings; render the ` [CPU: <c>% | MEM: <m>%]` segment only on the tool-execution spinner, `%.1f%%`.
- **Rationale**: the operator keeps the reference's metrics variant (clarify Q2=2) and locked Linux/macOS only — which is exactly what makes hand-rolled sampling feasible with **no new dependency** (the reference does the same; the stdlib has no portable form).
- **Alternatives considered**:
  - `shirou/gopsutil` — a large new dependency for two numbers; conflicts with NFR-005.
  - Whole-machine CPU — noisier and less meaningful than the process's own usage.
  - Drop the segment (clarify Q2=1) — rejected by the operator.

## Decision 6: Gating — the standard-output terminal probe (closes Obs 1)

- **Decision**: draw the spinner only when **standard output** is a terminal **and** `-r` is off (matching the reference's `UseColor = isTTY && !raw`). This **wires the standard-output real-isatty probe** (`golang.org/x/term.IsTerminal`, already in the module graph) — the probe round 006 left as a **named pin** (PR #16 **Obs 1**) because no own presentation existed to gate. Add a diagnostic `TELL_ME_FORCE_STDOUT_TTY` seam (mirroring the existing `TELL_ME_FORCE_STDIN_TTY`) so the stdout-terminal branch is E2E-drivable **without a pty**.
- **Rationale**: the operator locked reference-parity gating; the spinner is tellme's first own presentation chrome, so its gate finally has a purpose and the deferred probe can be wired and tested. The forced seam keeps the E2E suite hermetic (the harness pipes stdout, which is not a terminal).
- **Alternatives considered**:
  - Gate on the **stderr** terminal (where the spinner is drawn) — diverges from the pinned round-005/006 contract ("presentation suppressed when standard output is not a terminal").
  - A pty harness — forsworn (round 005 grill Q6 / round 006 Q2); the forced seam covers the branch.

## Decision 7: Lifecycle placement

- **Decision**: the turn layer owns the spinner lifecycle — start on entering a waiting phase (awaiting the model; executing tools), stop before any interleaved write (a tool-trace line, the answer, the post-turn lines), resume if waiting resumes, and always stop before the turn completes or a failure class phrase is written (a deferred stop guard). Phase labels are passed in per phase; the presenter is a value on the runtime environment so tests inject a recording `stderr` + a fake clock/ticker.
- **Rationale**: one owner avoids double-starts; the deferred stop guarantees no residue on any exit path, including panics.
- **Alternatives considered**:
  - Event-bus/actor lifecycle (the reference's `uiBridge` actor) — overkill for a single CLI turn with no concurrent producers.
  - A per-call spinner owned inside the agent loop — splits the lifecycle across layers.

## Decision 8: Verification (unit + E2E, no pty)

- **Decision**: unit tests for the pure helpers — the label/elapsed formatter, the `%.1f%%` metrics formatter, and the CPU/memory samplers driven by injected readings (Linux/macOS parser tables); E2E through the existing harness by setting `TELL_ME_FORCE_STDOUT_TTY` — assert the spinner frames appear on `stderr` while waiting, that the label names the model/tools, that the metrics segment appears only for tools, and that `stdout` stays byte-exact; assert the negatives (piped, `-r`, non-prompt paths, `-i`) draw none.
- **Rationale**: matches the round-011/014/018 pattern (unit pure helpers + E2E acceptance) and keeps `make verify` deterministic; the forced seam replaces the forsworn pty.
- **Alternatives considered**:
  - Assert via a pty — forsworn.
  - Snapshot the animated bytes exactly — brittle; assert presence/animation/labels by pattern instead.

## Decision 9: Platform and dependency footprint

- **Decision**: POSIX-only (Linux/macOS); **no** new third-party dependency (`go.mod`/`go.sum` unchanged).
- **Rationale**: the operator locked Linux/macOS only; the spinner, the sampling, and the probe all reuse already-present facilities.
- **Alternatives considered**: a Windows branch — out of scope.

## Must-ask questions (settled by existing truth — unchanged this round)

- **BDD techstack**: `godog` (existing `specs/truth/techstack.md`).
- **Test strategy**: E2E (black-box) for the acceptance path + pure-helper unit tests.
- **System ends**: a single CLI end.
