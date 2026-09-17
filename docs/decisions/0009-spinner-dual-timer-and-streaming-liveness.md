# ADR 0009 — Spinner dual elapsed timer + streaming liveness: a per-model-call elapsed figure, and an idle-gap resume during a `[Tool Output]` block

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Supersedes:** **ADR 0005 D7** (the whole-block spinner pause during a streaming `[Tool Output]` block) — **D7 only**; the rest of ADR 0005 (D1–D6, D8…) stands, so this ADR does **not** flip ADR 0005's overall `Status`. The supersession is annotated on the **decisions index** row for 0005 (the index is exempt from immutability).
- **Related:** issue [#82](https://github.com/gosharplite/tellme/issues/82) (streaming liveness) and issue [#83](https://github.com/gosharplite/tellme/issues/83) (dual timer);
  ADR 0005 (the tool-call log policy whose D7 pause this ADR narrows), round 034 (the per-call frame + the pause), round 035 (the per-call tail yield), round 025 (the row-aware clear), round 019 (the turn-scoped elapsed this ADR extends), round 027 (the AI-endpoint-call semantics the reset boundary follows);
  round 040 (`specs/plans/040-spinner-liveness-and-turn-timer` — this ADR's round)

## Context

Two operator-requested spinner changes arrive together (they touch `internal/ui/spinner.go` and `presenting-the-progress-spinner.feature`, so they are folded into one round — the round-038/039 precedent):

1. **A second elapsed figure (issue [#83](https://github.com/gosharplite/tellme/issues/83)).** The spinner shows **one** figure, whole seconds since **prompt capture**, turn-scoped and **never reset** (round-019 D4): `⠋ Thinking [<model>]... (120s)`. The operator wants a **second** figure — the **current model call's** elapsed — reset at each **AI-endpoint call**: `⠋ Thinking [<model>]... (120s 21s)`.
2. **Streaming liveness (issue [#82](https://github.com/gosharplite/tellme/issues/82)).** While a shell call streams its live `[Tool Output]` block, the round-019/025 spinner is **paused for the whole block** (round-034 FR-012 / G8, **ADR 0005 D7**). On a quiet/slow command there is **no** animation for the block's duration.

Two items need a citable home: an **amendment** to round-019 D4 (the total figure keeps its never-reset rule; a *second* figure is added and *does* reset per call), and a **supersession of ADR 0005 D7** (the whole-block pause → an idle-gap pause).

## Terminology

- **turn** — the whole prompt exchange (a submitted prompt and everything the run does for it): **1..k model calls**.
- **model call** (AI-endpoint call) — one request/response round to the provider.

The second figure measures the **current model call**, hence its name **"the current model call's elapsed"** — the truth vocabulary must not call it a "turn" (round 027 already defines a turn as the whole exchange; two names for one concept is a defect).

## Decision

**D1 — The elapsed display is a dual timer: `({total}s {call}s)`.** The first figure is the **total** elapsed since the turn's **prompt-capture** epoch — turn-scoped, **never reset** within a prompt (round-019 D4 preserved). The second is the **current model call's** elapsed, **reset at the start of each AI-endpoint call**. Both are whole seconds, floored at 0, **unlabelled** (D5, QB4); a single-call turn shows two approximately-equal figures (no special-casing).

**D2 — The reset boundary is the AI-endpoint call, set in `OnInferenceStart`; no loop change; no new constructor seam.** `AgentLoop.Run` fires `notifyInferenceStart()` **exactly once per AI-endpoint call** (call `i`: `OnCallBegin` → `OnInferenceStart` → … → `OnCallEnd`), and the spinner already reacts to it, so the second epoch is stamped there and every other hook (`OnInferenceEnd`, `OnToolsStart/End`, `Before/AfterToolLog`, the WS-A resume) preserves it. `internal/agent` stays unaware of the timer; no new observer hook is added; `NewSpinner` keeps its five positional seams (the second epoch is **not** a sixth parameter — round-019 R-2).

**D3 — During a `[Tool Output]` block the spinner is no longer hidden for the whole block; it resumes after an idle gap.** The block keeps its `stderr` ownership (round-034 G8 — the yield is not per line). After **N seconds with no new output line** (default **3 s**, the seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS` — milliseconds; `0` = admit immediately; unset/invalid = the default; the watcher polls on the spinner's **~200 ms** cadence, so the resume lands within `N … N+P`) the indicator **resumes**; the **next complete output line** is preceded by a **synchronous clear** of the indicator, then written, and output continues. The block literals, the bound/stop semantics, and the non-TTY/`-r` gates are unchanged.

**D4 — The resume/clear is mutual exclusion + join, and the block critical section never spans a frame write.** The invariant is **not** "single-writer" (after WS-A the redraw goroutine and the line writer both touch `stderr`); it is: (i) the resume is **admitted only while holding the block writer's mutex**; (ii) the clear is the presenter's **synchronous join** — `deactivate()` waits for the redraw goroutine to exit, then erases every occupied row; (iii) **lock order = block-writer mutex → spinner mutex**. The resume is admitted under the mutex but the **redraw goroutine draws the resumed first frame** — via a **named resume-only admission path** (an explicit `admitResume()`, *not* a global change to `activate()`) and it renders **immediately on start** (once before waiting for the tick), so no blocking `write(2)` to `stderr` is held inside the block critical section (a synchronous first frame there could stall the child's `io.Copy` drains on a stalled stream and push a command into its time limit), while `OnInferenceStart`/`OnToolsStart` keep their **synchronous** first frame (rounds 019/025/034/035 depend on it). The no-residue contract (rounds 019/025/035) holds; a resume with no phase label is a defined no-op.

**D5 — Recorded divergences (three).** `tell-me-go`'s spinner shows a **single** figure that **restarts per phase** (`startTime` reset on each `spinner.start`; `internal/ui/tui/progress/model.go`); it does **not** re-activate a live spinner over its streamed tool output (its `warnWriter` appends `ToolOutputStreamEvent`s to a TUI log viewport); and its figure is **unlabelled**. tellme deliberately diverges: the dual timer separates a slow call from a slow series of calls, the idle-gap liveness removes a silent window, and the two figures stay **unlabelled** (QB4 — matching the reference's unlabelled form; labelling would be a format change we do not make). Divergences of this class are recorded here and in the `techstack.md` row (the ADR 0005 D6 / ADR 0007 D6 precedent).

## Alternatives considered

1. **A permanent output-progress marker on the paused block** — rejected: it changes the `[Tool Output]` block's pinned literals/shape and its sanitize/structure pins.
2. **A safe per-line spinner yield** — rejected: the listener's background redraw would interleave output lines (round-034 **G8**'s stated blocker); the idle-gap approach resumes only after a **quiet** stretch.
3. **A separate clock/ticker seam for the second figure** — rejected: two epochs on the existing `now` seam suffice; and a sixth positional constructor seam would trip round-019 R-2.
4. **Reset the second figure at every waiting-phase transition (per phase)** — rejected: it would not match the header's **AI-endpoint-call** semantics; the operator locked the per-call boundary.
5. **Flip ADR 0005's whole `Status` to superseded** — rejected: only **D7** is superseded; D1–D6/D8 stand, so the supersession is named here and annotated on the **index** row (the index is exempt from immutability — no body edit).
6. **Admit the resume with a synchronous first frame** — rejected (D4): it holds the block mutex across a blocking frame write (back-pressure coupling to the child's drains); the redraw goroutine draws the first frame instead.

## Consequences

- The spinner line carries **two unlabelled whole-second figures**; the **total** is turn-scoped and never reset, the **second** figure (the current model call's elapsed) resets per AI-endpoint call.
- A quiet streaming command shows a live indicator again; output lines are never interleaved by a frame and no residue survives (the round-019/025/035 contract, re-witnessed for the longer two-figure line, incl. 3+-digit frames).
- `stdout`, flags, exit codes, the class-phrase vocabulary, the `[Tool …]` lines and their blank-line grouping, the block literals and bound/stop semantics, the tool result fed to the model, every persisted record, and the `isatty(stderr) && !-r` gate are **unchanged**.
- **Governance**: round-019 D4 is **amended** (a second figure is added; the total keeps its never-reset rule) and **ADR 0005 D7** is **superseded** (the whole-block pause → an idle-gap pause). ADR 0005 remains `Accepted`; its body is not edited (the index row carries the pointer). The three reference divergences are recorded.
- **Harness**: `tests/e2e/suite_test.go` sets `Strict: true`, so the round's new truth Examples fail the suite until `/axb-implement` defines their steps (previously reported-and-ignored — round 040 review TD-1).
