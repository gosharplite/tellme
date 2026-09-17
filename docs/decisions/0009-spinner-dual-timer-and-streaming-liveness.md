# ADR 0009 — Spinner dual elapsed timer + streaming liveness: a per-AI-call turn timer, and an idle-gap resume during a `[Tool Output]` block

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Supersedes:** **ADR 0005 D7** (the whole-block spinner pause during a streaming `[Tool Output]` block) — **D7 only**; the rest of ADR 0005 (D1–D6, D8…) stands, so this ADR does **not** flip ADR 0005's overall `Status`
- **Related:** issue [#82](https://github.com/gosharplite/tellme/issues/82) (streaming liveness) and issue [#83](https://github.com/gosharplite/tellme/issues/83) (dual timer);
  ADR 0005 (the tool-call log policy whose D7 pause this ADR narrows), round 034 (the per-call frame + the pause), round 035 (the per-call tail yield), round 025 (the row-aware clear), round 019 (the turn-scoped elapsed this ADR extends), round 027 (the AI-endpoint-call turn semantics the reset boundary follows);
  round 040 (`specs/plans/040-spinner-liveness-and-turn-timer` — this ADR's round)

## Context

Two operator-requested spinner changes arrive together (they touch `internal/ui/spinner.go` and `presenting-the-progress-spinner.feature`, so they are folded into one round — the round-038/039 precedent):

1. **Dual elapsed timer (issue [#83](https://github.com/gosharplite/tellme/issues/83)).** The spinner shows **one** figure, whole seconds since **prompt capture**, turn-scoped and **never reset** (round-019 D4): `⠋ Thinking [<model>]... (120s)`. The operator wants a **second** figure — the **current turn's** own duration — reset at each **AI-endpoint call** (the round-027 turn semantics): `⠋ Thinking [<model>]... (120s 21s)`.
2. **Streaming liveness (issue [#82](https://github.com/gosharplite/tellme/issues/82)).** While a shell call streams its live `[Tool Output]` block, the round-019/025 spinner is **paused for the whole block** (round-034 FR-012 / G8, **ADR 0005 D7**). On a quiet/slow command there is **no** animation for the block's duration, so the operator cannot tell the run is alive.

Two items need a citable home:

- an **amendment** to round-019 D4 (the total figure's never-reset rule is preserved; a *second* figure is added and *does* reset per call);
- a **supersession of ADR 0005 D7** (the whole-block pause is narrowed to an idle-gap pause).

## Decision

**D1 — The elapsed display is a dual timer: `({total}s {turn}s)`.** The first figure is the **total** elapsed since the turn's **prompt-capture** epoch — turn-scoped, **never reset** within a prompt (round-019 D4 preserved). The second is the **current turn's** duration, **reset at the start of each AI-endpoint call**. Both are whole seconds, floored at 0; a single-call turn shows two approximately-equal figures (no special-casing).

**D2 — The reset boundary is the AI-endpoint call, set in `OnInferenceStart`; no loop change.** `AgentLoop.Run` fires `notifyInferenceStart()` **exactly once per AI-endpoint call** (call `i`: `OnCallBegin` → `OnInferenceStart` → … → `OnCallEnd`), and the spinner already reacts to it, so the second epoch is stamped there and every other hook (`OnInferenceEnd`, `OnToolsStart/End`, `Before/AfterToolLog`, the WS-A resume) preserves it. `internal/agent` stays unaware of the timer; no new observer hook is added.

**D3 — During a `[Tool Output]` block the spinner is no longer hidden for the whole block; it resumes after an idle gap.** The block keeps its **single-writer** `stderr` ownership (round-034 G8 — the yield is not per line). After **N seconds with no new output line** (default **3 s**, exposed through a hermetic seam) the indicator **resumes**; the **next complete output line** is preceded by a **synchronous clear** of the indicator, then written, and output continues. The block literals, the bound/stop semantics, and the non-TTY/`-r` gates are unchanged.

**D4 — The resume/clear reuses the presenter's existing synchronous primitives, serialized with the block writer.** The idle watcher and the writer's clear+write are serialized on the **block writer's mutex**, and the clear (`deactivate()`) is goroutine-joined and erases every occupied row **after** the redraw goroutine exits, so a frame can neither interleave an output line nor be stranded. The no-residue contract (rounds 019/025/035) holds.

**D5 — Recorded divergences (two).** `tell-me-go`'s spinner shows a **single** figure that **restarts per phase** (`startTime` reset on each `spinner.start`; `internal/ui/tui/progress/model.go`), and it does **not** re-activate a live spinner over its streamed tool output (its `warnWriter` appends `ToolOutputStreamEvent`s to a TUI log viewport). tellme deliberately diverges on both: the dual timer separates a slow call from a slow series of calls, and the idle-gap liveness removes a silent window. Divergences of this class are recorded here and in the `techstack.md` row (the ADR 0005 D6 / ADR 0007 D6 precedent).

## Alternatives considered

1. **A permanent output-progress marker on the paused block** — rejected: it changes the `[Tool Output]` block's pinned literals/shape and its sanitize/structure pins.
2. **A safe per-line spinner yield** — rejected: the listener's background redraw would interleave output lines (round-034 **G8**'s stated blocker); the idle-gap approach resumes only after a **quiet** stretch.
3. **A separate clock/ticker seam for the turn timer** — rejected: two epochs on the existing `now` seam suffice.
4. **Reset the turn figure at every waiting-phase transition (per phase)** — rejected: it would not match the header's **AI-endpoint-call** turn semantics (round 027); the operator locked the per-call boundary.
5. **Flip ADR 0005's whole `Status` to superseded** — rejected: only **D7** is superseded; D1–D6/D8 stand, so the supersession is named explicitly here rather than retiring the whole ADR.

## Consequences

- The spinner line carries two whole-second figures; the **total** is turn-scoped and never reset, the **turn** figure resets per AI-endpoint call.
- A quiet streaming command shows a live indicator again; output lines are never interleaved by a frame and no residue survives (the round-019/025/035 contract, re-witnessed for the longer line).
- `stdout`, flags, exit codes, the class-phrase vocabulary, the `[Tool …]` lines and their blank-line grouping, the block literals and bound/stop semantics, the tool result fed to the model, every persisted record, and the `isatty(stderr) && !-r` gate are **unchanged**.
- **Governance**: round-019 D4 is **amended** (a second figure is added; the total keeps its never-reset rule) and **ADR 0005 D7** is **superseded** (the whole-block pause → an idle-gap pause). ADR 0005 remains `Accepted`; its body is not edited. The reference divergences are recorded.
