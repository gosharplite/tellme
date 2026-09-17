# Technical Research: spinner liveness while a command streams + a per-turn elapsed timer (round 040)

**Plan Package**: `specs/plans/040-spinner-liveness-and-turn-timer`

**Status**: Decisions D1–D8 (ratifies the operator's **QB1** lock — the per-AI-endpoint-call reset boundary). **QA1–QA3** (WS-A mechanism/threshold/scope) and **QB2/QB3** (format/phases) are the round's **proposed** readings, pending the operator's confirmation at the pre-`/axb-tasks` review gate.

## Context

Two folded spinner workstreams on `internal/ui/spinner.go` + `presenting-the-progress-spinner.feature`:

- **#82** — while a shell call streams its live `[Tool Output]` block, the round-019/025 spinner is **paused for the whole block** (round-034 FR-012/G8, ADR 0005 D7). On a quiet command the operator sees **no** animation for the block's duration.
- **#83** — the spinner shows **one** figure: whole seconds since prompt capture, turn-scoped, never reset (round-019 D4). The operator wants a **second** figure: the current **turn's** duration, reset per AI-endpoint call.

### Reference check (`tell-me-go`)

- **Spinner line**: `internal/ui/tui/progress/model.go` → `fmt.Sprintf("%s %s (%ds)", frame, status, elapsed)` (with a ` [CPU: %.1f%% | MEM: %.1f%%]` suffix when `showMetrics`), where `elapsed = int(time.Since(s.startTime).Seconds())` and `s.startTime` is reset on **each** `spinner.start` (a per-**phase** event). So the reference's single figure **restarts per phase** — the opposite pole from tellme's turn-scoped never-reset (round-019 D4). WS-B's second figure is at **AI-endpoint-call** granularity: neither the reference's per-phase restart nor tellme's never-reset — a **new recorded divergence**.
- **Streamed tool output**: the reference's `shellTool` `warnWriter` publishes `ToolOutputStreamEvent`s that its TUI progress model appends to a log viewport; it does **not** redraw a live single-line spinner over them. tellme's line-mode surfaces **pause** the spinner for the whole block (round 034). WS-A's idle-gap liveness is therefore also a **new recorded divergence** (operator-requested).

Both are recorded divergences, not silent parity breaks.

## Decisions

### D1 — WS-B: a second elapsed figure, reset per AI-endpoint call

The line gains a **second** whole-second figure: the **total** since prompt capture first (unchanged; turn-scoped; never reset — round-019 D4 preserved), then the **current turn's** duration, reset at the start of each **AI-endpoint call** (QB1). The line becomes `{frame}{status} ({total}s {turn}s){resource}`. Both figures are whole seconds, floored at 0. A single-call turn shows two approximately-equal figures (no special-casing).

Rejected forms: a byte/millisecond precision (the reference is whole seconds); a third clock seam (unnecessary — two epochs on one `now`).

### D2 — WS-B mechanics: a second epoch set in `OnInferenceStart`; **no** loop change

`Spinner` already carries one epoch (`epoch`, the prompt-capture time). It gains a second, the **current call's** start time (`turnEpoch`). `OnInferenceStart` sets `turnEpoch = now()` **before** activating (so the first frame of each call already shows `(… 0s)`), then relabels as today. Every other hook is **unchanged**: `OnInferenceEnd` (no-op), `OnToolsStart`/`OnToolsEnd` (relabel within the **same** call — the turn figure keeps counting), `BeforeToolLog`/`AfterToolLog` (a clear/resume within the same call — the turn figure is preserved), and the WS-A idle resume (D4 — also within the same call).

Why this is per-call **by construction**: `AgentLoop.Run` fires `notifyCallBegin(i, …)` → `notifyInferenceStart()` → (inference) → `notifyInferenceEnd()` → (`notifyToolsStart/End`) → `notifyCallEnd(i, …)` — `OnInferenceStart` fires **exactly once per AI-endpoint call**, and the spinner already reacts to it. So the reset boundary needs **no** loop change and no new observer hook. `internal/agent` stays unaware of the timer.

`FormatSpinnerLine(frame, status string, total, turn int, resource string)` gains the second figure; `NewSpinner(…, epoch time.Time, …)` keeps its single inbound epoch (the turn epoch is initialised to it, so call 1's turn figure starts from prompt capture, and reset per call thereafter).

### D3 — WS-A: idle-gap liveness during a streaming `[Tool Output]` block

Keep the block's **single-writer** `stderr` ownership (round-034 G8) — do **not** go per-line — but do **not** stay hidden for the whole block: after **N seconds with no new output line**, the indicator **resumes** (re-drawn on the cleared line); when the next complete output line arrives, the indicator is **synchronously cleared** immediately before that line is written, and output continues. The block literals, the bound/stop semantics, and the non-TTY/`-r` gates are unchanged. Default **N = 3 s** (the round-019 research fast-fail precedent), exposed through a **hermetic env seam** so E2E can force it without a wall-clock sleep (round-016 `TELL_ME_TUI_DEBOUNCE=0` precedent).

Rejected: **whole-block pause** (the status quo — issue #82); **per-line yield** (G8's stated blocker — the spinner's background redraw would interleave output lines); an **output-progress marker** line owned by the block writer (it would change the block's pinned literals/shape and the `[Tool Output]` sanitize/structure pins).

### D4 — WS-A mechanics: reuse the synchronous clear/resume, serialized with the block writer

The presenter already has a **synchronous** clear (`deactivate()`, idempotent, waits for the redraw goroutine to exit before erasing; ADR-0005 G8 uses it as `BeforeToolLog`) and a resume (`resume()`), used by the sink today. WS-A reuses **the same pair** block-scoped:

1. `Begin` (before the child starts): clear the spinner, write the header/separator (unchanged).
2. A **block-scoped idle watcher** (a `time.Ticker` on the injected ticker/now seams, started at `Begin`, stopped at `End`) resumes the indicator (`resume()`) once `now − lastLine ≥ N`.
3. On each **complete** output line: if the indicator is live, **clear it first** (`deactivate()`), then write the line, and stamp `lastLine = now()` (the writer stays the only `stderr` writer for the block).
4. `End`: stop the watcher, write the closing separator, then resume (unchanged).

**Single-writer safety** rests on two properties of the existing primitives: (a) `deactivate()` is synchronous and erases every row the last frame occupied **after** the redraw goroutine has exited, so a line written immediately after a clear cannot be interleaved by a frame; (b) the watcher's resume and the writer's clear+write are serialized on the **block writer's mutex** (the watcher re-checks `lastLine` under that mutex before resuming), so a resume cannot slip between a clear and the following line. This keeps the round-019/025/035 no-residue contract.

The coordinator is a small `internal/ui` type (the `ToolOutputWriter` + a `*Spinner` + the idle seams), so `internal/cli` passes one object to the sink and `ui` owns the coordination (single ownership — the #69 theme).

### D5 — Scope guard

Unchanged: the block literals (header `... (Output shown below)` + 60-hyphen separator) and the neutral-close restore; the bounded-and-stopped semantics (round-024); `output_file` (no block); the tool-log lines and their round-039 blank-line grouping; the spinner's `isatty(stderr) && !-r` gate and the `-i` submit surface; the class-phrase vocabulary; exit codes; flags; the tool **result** fed to the model; every persisted record; `stdout` bytes. The spinner stays hand-written (no dependency); POSIX-only.

### D6 — Witnesses (falsifiability)

| Layer | Witness |
| --- | --- |
| Unit (`internal/ui`) | WS-B: with an injected `now`, a first `OnInferenceStart` then a render shows `(t t)`, a later render shows `(t+k t+k)`, and a second `OnInferenceStart` shows the total grown but the turn figure reset. WS-A: with an injected ticker/now, a quiet stretch past N resumes the indicator and the next line clears it before its bytes (no interleave, no residue). |
| Unit (`internal/ui`) | the round-025 row-aware clear still erases every row for the **longer** two-figure line (no residue). |
| E2E (`tests/e2e`) | a scripted quiet-then-resuming command (forced idle threshold) shows a spinner frame **inside** the block window on the captured `stderr` (SC-002); a scripted multi-call turn shows both figures with the total monotonic and the turn figure dropping at a call boundary (SC-003). |

Falsifiability: (a) remove the idle-gap resume → the WS-A Example fails; (b) freeze the turn figure (no per-call reset) → the WS-B Example fails; (c) drop the row-aware clear → a residue frame survives on the wider line — each reproduced then reverted.

### D7 — Truth & governance impact (ratified)

- `specs/truth/techstack.md` — **MODIFY** the **Turn progress spinner** row: the dual elapsed timer (total + per-AI-endpoint-call, amending round-019 D4's "never reset") and the idle-gap liveness during a `[Tool Output]` block (superseding round-034's whole-block pause).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md` — **MODIFY** (amend the "paused while streaming" Rule; add a dual-timer Rule; their `DSLRow`s) — owned by `/axb-dsl-refine`.
- `docs/decisions/` — **ADD ADR 0009** ("Spinner dual elapsed timer + streaming liveness"), which **supersedes ADR 0005 D7** (the whole-block pause only; the rest of ADR 0005 stands) and records the round-019-D4 amendment.
- `/axb-api-plan` — **NOOP** (no HTTP surface). `/axb-data-plan` — checked **NOOP** (no persisted state).

### D8 — ADR lifecycle: supersede a single decision **without rewriting** ADR 0005

ADR 0005 is `Accepted` and immutable but for its `Status` line (decisions README). WS-A revises **one** of its decisions (D7, the whole-block pause) and leaves the rest (D1–D6, D8…) in force — a *partial* supersession. Rather than flipping 0005's whole `Status` (which would falsely retire D1–D6), round 040 **writes ADR 0009** and marks the supersession **there** and in the decisions **index** (a "supersedes ADR 0005 D7" note on 0009's row); ADR 0005's body is **not** edited. `tellme`'s precedent for a full supersede is ADR 0008 → 0007; this partial case is handled by naming the superseded decision explicitly in the new ADR. *(If the operator/architect prefers a `D7`-line forward pointer inside 0005, that is a one-line, immutability-clause-blessed `Status`-line change — flagged for the review gate.)*

## Residual risks (forward)

- **Interleaving safety (WS-A)** is the highest-risk area: the design relies on `deactivate()`'s synchronous, goroutine-joined clear plus a shared mutex. A dedicated unit stress (concurrent line writes + a fast ticker) is required, and the E2E must assert no frame residue and no byte loss/reorder. If the stress reveals a window, the fallback is to widen the shared lock to cover the resume as well (already planned) or, worst case, fall back to D3's rejected output-progress-marker (recorded).
- **The idle threshold** is a fixed small value with an env seam; a value that is too small risks frequent clear/resume flicker under slow-but-continuous output (recorded; the reference has no equivalent).
- **The longer line** (two figures) is a small extra soft-wrap risk on a narrow terminal — bounded by the existing round-025 rune-based row count.
- **The reference divergences** (dual timer; liveness) are recorded; aligning to the reference is explicitly **not** a goal here.
