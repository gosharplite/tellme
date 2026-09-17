# Technical Research: spinner liveness while a command streams + a per-model-call elapsed timer (round 040)

**Plan Package**: `specs/plans/040-spinner-liveness-and-turn-timer`

**Status**: Decisions D1–D9 (ratifies the operator's **QB1** lock — the per-AI-endpoint-call reset boundary — and folds the **architect review** of PR [#84](https://github.com/gosharplite/tellme/pull/84): TD-2 (protocol wording), TD-3 (critical-section bound), TD-4 (poll period), TD-5 (seam), TD-6 (terminology), D8 (index annotation), QB3/QB4).

## Context

Two folded spinner workstreams on `internal/ui/spinner.go` + `presenting-the-progress-spinner.feature`:

- **#82** — while a shell call streams its live `[Tool Output]` block, the round-019/025 spinner is **paused for the whole block** (round-034 FR-012/G8, ADR 0005 D7). On a quiet command the operator sees **no** animation for the block's duration.
- **#83** — the spinner shows **one** figure: whole seconds since prompt capture, turn-scoped, never reset (round-019 D4). The operator wants a **second** figure: the current model call's elapsed, reset per AI-endpoint call.

### Terminology (reviewer TD-6)

**turn** = the whole prompt exchange (1..k model calls); **model call** = one AI-endpoint request/response. The second figure is named **"the current model call's elapsed"** so the truth vocabulary does not contradict round 027.

### Reference check (`tell-me-go`)

- **Spinner line**: `internal/ui/tui/progress/model.go` → `fmt.Sprintf("%s %s (%ds)", frame, status, elapsed)` (with a ` [CPU: %.1f%% | MEM: %.1f%%]` suffix when `showMetrics`), where `elapsed = int(time.Since(s.startTime).Seconds())` and `s.startTime` is reset on **each** `spinner.start` (a per-**phase** event). So the reference's single figure **restarts per phase** — the opposite pole from tellme's turn-scoped never-reset (round-019 D4). WS-B's second figure is at **AI-endpoint-call** granularity: neither the reference's per-phase restart nor tellme's never-reset — a **new recorded divergence**. The reference also shows the figure **unlabelled**; tellme keeps both unlabelled (QB4, recorded).
- **Streamed tool output**: the reference's `shellTool` `warnWriter` publishes `ToolOutputStreamEvent`s that its TUI progress model appends to a log viewport; it does **not** redraw a live single-line spinner over them. tellme's line-mode surfaces **pause** the spinner for the whole block (round 034). WS-A's idle-gap liveness is therefore also a **new recorded divergence** (operator-requested).

Both are recorded divergences, not silent parity breaks.

## Decisions

### D1 — WS-B: a second elapsed figure (the current model call), reset per AI-endpoint call

The line gains a **second** whole-second figure: the **total** since prompt capture first (unchanged; turn-scoped; never reset — round-019 D4 preserved), then the **current model call's** elapsed, reset at the start of each **AI-endpoint call** (QB1). The line becomes `{frame}{status} ({total}s {call}s){resource}`. Both figures are whole seconds, floored at 0, **unlabelled** (QB4). A single-call turn shows two approximately-equal figures (no special-casing).

Rejected forms: a byte/millisecond precision (the reference is whole seconds); a third clock seam (unnecessary — two epochs on one `now`); **labelled** figures (QB4 — the operator/reference format is two bare numbers; recording labelling as a divergence would be a format change we do not make).

### D2 — WS-B mechanics: a second epoch set in `OnInferenceStart`; **no** loop change; **no** new constructor param

`Spinner` already carries one epoch (`epoch`, the prompt-capture time). It gains a second, the **current call's** start time (field **`callEpoch`**, initialised to `epoch`). `OnInferenceStart` sets `callEpoch = now()` **before** activating (so the first frame of each call already shows `(… 0s)`), then relabels as today. Every other hook is **unchanged**: `OnInferenceEnd`, `OnToolsStart`/`OnToolsEnd` (relabel within the **same** call — the second figure keeps counting), `BeforeToolLog`/`AfterToolLog` (a clear/resume within the same call — the second figure is preserved), and the WS-A idle resume.

`FormatSpinnerLine(frame, status string, total, call int, resource string)` gains the second figure (the second param is the **current model call's** elapsed — not "turn"). **`NewSpinner` keeps its five positional seams** — the second epoch is **not** a sixth parameter (that would trip round-019's own recorded R-2 nit, which says a sixth positional seam needs a functional-options refactor). Reviewer RF-4.

Why this is per-call **by construction**: `AgentLoop.Run` fires `notifyCallBegin(i, …)` → `notifyInferenceStart()` → (inference) → `notifyInferenceEnd()` → (`notifyToolsStart/End`) → `notifyCallEnd(i, …)` — `OnInferenceStart` fires **exactly once per AI-endpoint call** (verified against `agentloop.go:134-137`), and the spinner already reacts to it. So the reset boundary needs **no** loop change and no new observer hook; `internal/agent` stays unaware of the timer.

### D3 — WS-A: idle-gap liveness during a streaming `[Tool Output]` block (seam + cadence pinned)

Keep the block's `stderr` ownership (round-034 G8) — do **not** go per-line — but do **not** stay hidden for the whole block: after **N seconds with no new output line**, the indicator **resumes**; when the next complete output line arrives, the indicator is **synchronously cleared** immediately before that line is written, and output continues. The block literals, the bound/stop semantics, and the non-TTY/`-r` gates are unchanged.

- **Idle gap N** = a small fixed default (**3 s**), exposed through the hermetic env seam **`TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`** — **milliseconds**; **`0` = admit immediately** (the gap has already elapsed); **unset or invalid = the 3 s default**; a negative value is invalid → the default. This is the round-016 `TELL_ME_TUI_DEBOUNCE` precedent (a diagnostic seam named once, exact unit, exact sentinel).
- **Poll period P** (reviewer TD-4) = the spinner's existing **~200 ms** ticker cadence, so the resume is admitted within **one poll** of N — the observable latency is pinned to `N … N+P` (and the unit assertion uses the injected ticker to make it exact). The watcher is started at `Begin` and stopped at `End`.

Rejected: **whole-block pause** (the status quo — issue #82); **per-line yield** (G8's stated blocker — the spinner's background redraw would interleave output lines); an **output-progress marker** line owned by the block writer (it would change the block's pinned literals/shape and the `[Tool Output]` sanitize/structure pins).

### D4 — WS-A mechanics: mutual exclusion + join (not "single-writer"); the block critical section does not span a frame write

The presenter already has a **synchronous** clear (`deactivate()`, idempotent, joins the redraw goroutine before erasing — ADR-0005 G8 uses it as `BeforeToolLog`) and a resume (`resume()`), used by the sink today. WS-A reuses the pair, block-scoped, under a **stated protocol** (reviewer TD-2 — "single-writer" is no longer literally true once a redraw goroutine and the line writer both touch `stderr`; the invariant is **mutual exclusion + join**):

1. `Begin` (before the child starts): clear the spinner, write the header/separator (unchanged).
2. A **block-scoped idle watcher** (the ~200 ms ticker) resumes the indicator once `now − lastLine ≥ N`.
3. On each **complete** output line: if the indicator is live, **clear it first** (`deactivate()`), then write the line, and stamp `lastLine = now()`.
4. `End`: stop the watcher, write the closing separator, then resume (unchanged).

**Protocol (FR-002):** (i) the resume is **admitted only while holding the block writer's mutex**; (ii) the clear is the presenter's **synchronous join** — `deactivate()` waits for the redraw goroutine to exit before erasing every occupied row; (iii) **lock order = block-writer mutex → spinner mutex** (the spinner never calls into the writer, so no deadlock). The block critical section therefore never emits a frame while holding the block mutex: **the resume is admitted under the mutex but the redraw goroutine draws the first frame** (reviewer TD-3 — the synchronous first frame would otherwise hold the block mutex across a blocking `write(2)` to `stderr`; on a stalled stream that would stall the child's `io.Copy` drains and could push a command into its time limit). **TD-10 scope (reviewer):** this deferral applies **only to the in-block resume**, via a **named resume-only admission path** (e.g. an explicit `admitResume()` — *not* a global change to `activate()`), so `OnInferenceStart`/`OnToolsStart` keep the **synchronous** first frame that the rounds-019/025/034/035 E2E assertions depend on (`the run shows the progress spinner while it waits` on a fast scripted turn). **TD-10 precision:** "draws the first frame" means **immediately on start** — the resumed path renders **once before waiting for the tick** (today `loop()` renders only on `<-tick`, `spinner.go:289-306`, so a literal reading would delay the resumed frame by up to one more poll period). The **clear** stays synchronous (it is the safety property; the erase is a short, non-blocking write to an already-cleared line). This keeps the round-019/025/035 no-residue contract.

The coordinator is a small `internal/ui` type (the `ToolOutputWriter` + a `*Spinner` + the idle seams), so `internal/cli` passes one object to the sink and `ui` owns the coordination (single ownership). This **pays down part of [#69](https://github.com/gosharplite/tellme/issues/69)** (the spinner-yield policy's "three homes / four call sites"; reviewer RF-5 — update #69's body).

**RF-3:** the coordinator's resume path MUST be a defined **no-op** when the presenter has no phase label (`resume()` already returns early on an empty status); the block always begins during a tool phase today (the loop fires `OnToolsStart` before the batch), which is where a label is present — pinned so a future path cannot silently disable the feature.

### D5 — Scope guard

Unchanged: the block literals (header `... (Output shown below)` + 60-hyphen separator) and the neutral-close restore; the bounded-and-stopped semantics (round-024); `output_file` (no block); the tool-log lines and their round-039 blank-line grouping; the spinner's `isatty(stderr) && !-r` gate and the `-i` submit surface; the class-phrase vocabulary; exit codes; flags; the tool **result** fed to the model; every persisted record; `stdout` bytes. The spinner stays hand-written (no dependency); POSIX-only.

### D6 — Witnesses (falsifiability)

| Layer | Witness |
| --- | --- |
| Unit (`internal/ui`) | WS-B: with an injected `now`, a first `OnInferenceStart` then a render shows `(t t)`, a later render shows `(t+k t+k)`, and a second `OnInferenceStart` shows the total grown but the second figure reset. WS-A: with an injected ticker/now, a quiet stretch past N resumes the indicator and the next line clears it before its bytes. |
| Unit (`internal/ui`) | WS-A **race + anti-vacuity** (reviewer TD-2): a line arriving as the resume is admitted ⇒ no interleave, no residue; and **continuous** output ⇒ **no** frame between lines (proves the resume cannot decay into the rejected per-line yield). |
| Unit (`internal/ui`) | the round-025 row-aware clear still erases every row for the **longer two-figure 3+-digit** line (`1234s 567s`) — the erase-all-rows pin exercised on the longer line (reviewer QB3). |
| E2E (`tests/e2e`) | a scripted quiet-then-resuming command with a **real** idle gap — a small **nonzero** forced threshold (e.g. `50` ms) plus a **child** `sleep` (not a Go `time.Sleep`, so `verify-no-test-sleep` is untouched) — shows a spinner frame **inside** the block window on the captured `stderr` (SC-002). **Timing budget (reviewer TD-10):** the child's quiet stretch MUST exceed `N + 2·P` with margin (worst case = admit ≤ `N+P`, first frame ≤ `+P`), so with `N=50 ms`, `P=200 ms` use e.g. a **2 s** child `sleep` (not `0.5 s`, which leaves ~50 ms of slop and can clear the indicator before any frame byte). |
| E2E (`tests/e2e`) | a scripted multi-call turn shows both figures with the total monotonic and the second figure dropping at a call boundary (SC-003). |

Falsifiability: (a) remove the idle-gap resume → the WS-A Example fails; (b) freeze the second figure (no per-call reset) → the WS-B Example fails; (c) drop the row-aware clear → a residue frame survives on the wider line — each reproduced then reverted. **Harness (TD-1):** `tests/e2e/suite_test.go` gains `Strict: true` in the **implementation half** (with the stepdefs), so undefined steps then **fail** the suite rather than being reported-and-ignored; the plan half leaves the harness unchanged so `dev` stays green.

### D7 — Truth & governance impact (ratified)

- `specs/truth/techstack.md` — **MODIFY** the **Turn progress spinner** row: the dual elapsed timer (**total** + the **current model call's** elapsed, per-AI-endpoint-call reset, amending round-019 D4) and the idle-gap liveness during a `[Tool Output]` block (the mutual-exclusion + join protocol; the seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`; the ~200 ms poll; superseding round-034's whole-block pause).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md` — **MODIFY** (amend the "paused while streaming" Rule → the idle-gap liveness Rule; add a dual-timer Rule; the seam Given; their `DSLRow`s) — owned by `/axb-dsl-refine`.
- `docs/decisions/` — **ADD ADR 0009** ("Spinner dual elapsed timer + streaming liveness"), which **supersedes ADR 0005 D7** (the whole-block pause only; the rest of ADR 0005 stands) and records the round-019-D4 amendment + the three recorded divergences (per-phase reference figure; no reference liveness; unlabelled figures — QB4). **Annotate the decisions index row for 0005** (RF-2).
- `/axb-api-plan` — **NOOP** (no HTTP surface). `/axb-data-plan` — checked **NOOP** (no persisted state).

### D8 — ADR lifecycle: supersede a single decision **without rewriting** ADR 0005, and annotate the **index** (RF-2)

ADR 0005 is `Accepted` and immutable **except** its `Status` line **and the index** (`docs/decisions/README.md`). WS-A revises **one** decision (D7, the whole-block pause) and leaves the rest (D1–D6, D8…) in force — a *partial* supersession. Round 040 **writes ADR 0009** and, per reviewer RF-2, annotates the **index row** for 0005 rather than editing its body:

```
| [0005](0005-tool-call-log-parity.md) | … | Accepted (D7 superseded by [0009](0009-spinner-dual-timer-and-streaming-liveness.md)) |
```

That is strictly better than a body edit (no immutability breach, visible at the lookup point). **This resolves the open D8 trade-off flagged for the review gate** — no operator-level decision is needed.

### D9 — Harness: godog `Strict: true` (reviewer TD-1)

`godog.Options.Strict` defaults to `false`, so an undefined/pending/ambiguous step was reported in the pretty output while the suite exited 0, and `make verify` does not run the E2E at all. Round 040 ships **7 new steps** (4 sentences), so its witnesses would otherwise be unenforced. `tests/e2e/suite_test.go` gains `Strict: true` — **in the implementation half, together with the stepdefs** (reviewer correctness note: `make test` is `go test ./...`, which **includes** `tests/e2e`, so enabling Strict on the plan half would make `dev` red on every run; the implement half lands the flag with the steps, witnessed by the FAIL-then-PASS transition — the same plan/implement boundary as the dead stepdef's `[BDD-REMOVE]`). The implement half's burn-down list (4 sentences / 7 occurrences): `the command stays quiet for longer than the spinner's idle gap`; `a configured provider "…" whose endpoint runs a command that prints a line and then stays quiet and then answers with "…"`; `the run shows the progress spinner again while the command stays quiet`; `the progress spinner shows the time since the prompt and the current model call's time` (×4 occurrences). The dead stepdef `step_r034_t016_chat_then_no_spinner_while_streaming.go` (0 matching feature steps after the Rule replacement) is retired by `/axb-tasks` (RF-1).

## Residual risks (forward)

- **Interleaving safety (WS-A)** is the highest-risk area: the design relies on `deactivate()`'s synchronous join plus the block-writer mutex (admit under the mutex; the redraw goroutine draws the first frame). A dedicated unit stress (concurrent line writes + a fast ticker) is required, plus the anti-vacuity case; the E2E must assert no frame residue inside the block. If the stress reveals a window, the fallback is to widen the shared lock to cover the resume as well.
- **Back-pressure (WS-A, TD-3 bounded)**: the fix keeps the block mutex off the frame-write path; a pathological *spinner-mutex* stall could still delay a frame, but never the child's drains (the resume-admit holds the block mutex only for a bookkeeping update).
- **The idle threshold** is a fixed small value with an env seam; a value that is too small risks clear/resume flicker under slow-but-continuous output (recorded; the reference has no equivalent). The E2E uses a **nonzero** forced threshold so a real idle gap is observed.
- **The longer line** (two figures) is a small extra soft-wrap risk on a narrow terminal — bounded by the existing round-025 rune-based row count (re-witnessed at 3+ digits, QB3).
- **The reference divergences** (dual timer; liveness; unlabelled figures) are recorded; aligning to the reference is explicitly **not** a goal here.
