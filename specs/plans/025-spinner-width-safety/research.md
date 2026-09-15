# Phase 0 Research: tellme spinner width safety — bounded tool label + residue-free clear (round 025)

<!--
  Decision-driven research for round 025. Each decision keeps:
  Decision / Rationale / Alternatives considered.
-->

**Context.** Round-019 shipped the non-TUI turn progress spinner. Round-024 enlarged the agent tool
set (`execute_command`) and the model can batch several tools; the tool-phase label
(`internal/ui/spinner.go` `ExecutingToolsLabel`) joins **every** tool name
(` strings.Join(names, ", ")`), so the line grows without bound and can exceed the terminal width —
clipping the trailing `[CPU: … | MEM: …]` segment and soft-wrapping the frame. Because the spinner
clear is a **single-row** erase (`clearControl = "\r\x1b[K"`), a wrapped frame leaves residue past
the turn, violating the round-019 contract `the progress spinner no longer appears once the answer is
written` (issue #55).

**Reference finding.** `tell-me-go` has the **same** two properties — it enumerates every tool name
(`internal/domain/events/types.go`, `ToolExecutionStartedEvent.SpinnerInfo`) and clears a single row
(`\r` + `\033[2K`, `internal/ui/renderer_spinner.go`). There is **no upstream fix to port**; both
round-025 changes are deliberate **divergences** from the reference (permitted — tellme is a
re-specification, and it already records divergences).

## Decision 1: Bound the several-tool label (first name + count)

- **Decision**: `ExecutingToolsLabel` keeps its three branches but changes the several-tool branch: for
  `N == 0` → ` Executing tools...` (unchanged); `N == 1` → ` Executing [<name>]...` (unchanged); `N >= 2`
  → ` Executing tools [<first> and <N-1> more]...` — the **first** name and the **count of the remaining
  names**. The label is therefore a **constant-ish width** independent of batch size.
- **Rationale**: the enumerated label is the root cause of the over-wide line; bounding it makes the
  over-wide frame (in practice) impossible on a normal-width terminal and keeps the resource segment
  visible. The first name plus a count preserves the useful "what + how many" signal that a bare
  `[<first>]` would lose.
- **Alternatives considered**:
  - Clamp the full enumeration to a character budget (keeps more names but still needs width math, and
    the result is arbitrary and still unbounded for long names).
  - Show only the first name ` Executing [<first>]...` (loses the count — the operator cannot tell a
    one-tool turn from a ten-tool one).

## Decision 2: Make the clear width-safe by tracking the last frame's occupied rows

- **Decision**: the presenter tracks the number of terminal **rows** its last frame occupied —
  `rows = ceil(visibleWidth / columns)`, clamped to ≥ 1 — and, both before a redraw and on the final
  clear, emits a **multi-row** erase that removes **every** occupied row (move to the frame's last row,
  `\r\x1b[K`, then `\x1b[1A` up per remaining row, ending at the frame's first row). A single-row `\r\x1b[K`
  is the `rows == 1` special case.
- **Rationale**: the round-019 teardown contract requires no spinner residue once the answer is written;
  a single-line clear cannot honour it once a frame soft-wraps across ≥ 2 rows. Erasing every occupied
  row fixes the contract **regardless of why** the line wrapped (a long label, the resource segment, or
  a mid-turn terminal resize). *Known bound (TD-3):* the row count is captured at draw time, so a mid-frame resize leaves it stale and a clear may over-erase one row of prior output — accepted (the reference has no resize handling at all); residue is eliminated.
- **Alternatives considered**:
  - **Clamp the rendered line to the terminal width** (the issue's other option). Prevents wrapping at
    the source, but does not fix residue when the terminal is resized mid-frame, and silently truncates
    the status. (Rejected by the operator in favour of row-tracking.)
  - Save/restore the cursor (`\x1b7`/`\x1b8`) + erase-to-end-of-screen (`\x1b[J`). Width-independent,
    but fragile under scroll (the saved absolute row drifts once content scrolls).

## Decision 3: Terminal-width source — an injected `columns` seam

- **Decision**: add a `columns func() int` seam to `internal/ui.NewSpinner` (beside the round-019
  `now` / `newTicker` seams). The CLI default probes the **diagnostic stream** width via
  `golang.org/x/term.GetSize` on the `stderr` file descriptor (the spinner is already gated on
  `isatty(stderr)`, so the fd is available) and returns **0 when the width is unknown**; on `0` the
  presenter assumes a single row (best-effort). A diagnostic env seam `TELL_ME_FORCE_STDERR_COLS`
  (mirroring the round-019 `TELL_ME_FORCE_STDERR_TTY` seam) overrides the probe so the row-aware clear
  is drivable at a fixed width in E2E and unit tests.
- **Rationale**: the row count needs the width; a seam keeps the presenter deterministic and
  pty-free-testable. `golang.org/x/term` is **already a direct dependency** (round 012, the real isatty
  probe) and already in the module graph — **no new module**.
- **Alternatives considered**:
  - A hard-coded width constant (wrong on real terminals — the whole defect).
  - The `COLUMNS` environment variable (not authoritative; often unset under a real tty).

## Decision 4: Verification split — label bound E2E, row-aware clear unit + E2E bytes

- **Decision**: the **bounded label** is asserted **end-to-end** (the tool-phase spinner line's status
  text). The **row-aware clear** is pinned by a **unit** test with an injected narrow width (render an
  over-wide frame, clear, assert the emitted sequence erases every occupied row — a flat byte capture
  cannot reproduce a terminal grid), **and** witnessed end-to-end as the clear **byte sequence** via the
  `TELL_ME_FORCE_STDERR_COLS` seam (the captured `stderr` carries the cursor-up + per-row erase). The
  round-019 merged-capture teardown assertion (`the progress spinner no longer appears once the answer is
  written`) is retained.
- **Rationale**: mirrors the round-019 precedent (frame advancement is a unit pin; the phase labels are
  E2E) and the round-024 precedent (FR-018 is unit-pinned where a Gherkin `Then` cannot practically
  script it).
- **Alternatives considered**:
  - E2E-only via a real pty (the project forswore a pty harness — round 005 grill Q6 / round 006 Q2).
  - Unit-only for both defects (would leave the label bound without an acceptance witness).

## Residual risks / notes

- The bounded label changes the round-019 several-tool truth row (`the progress spinner names every tool
  it is running` → a bounded assertion); the round-019 plan package stays **frozen** (`fresh-package-per-round`).
- The `-i` submit surface (round 023) draws the same spinner and **inherits** both fixes automatically.
- No new failure class; the frozen class-phrase vocabulary is unchanged; `stdout` stays byte-exact.
