# Feature Specification: spinner liveness while a command streams + a per-model-call elapsed timer (round 040)

**Feature Branch**: `040-spinner-liveness-and-turn-timer`

**Created**: 2026-09-17

**Status**: Draft — two folded spinner workstreams. **QB1** (reset boundary) is **operator-locked**; **QA1–QA3 + QB2–QB4** are the round's proposed readings, **confirmed by the architect review** (PR [#84](https://github.com/gosharplite/tellme/pull/84) comment) and folded here; the round goes to the operator at the pre-`/axb-tasks` review gate.

**Input**: Two spinner workstreams folded into one round (both touch `internal/ui/spinner.go` and `presenting-the-progress-spinner.feature` — the round-038/039 fold precedent):

- **Issue [#82](https://github.com/gosharplite/tellme/issues/82)** — *spinner liveness while a command's `[Tool Output]` streams*. Today the round-019/025 spinner is **paused for the whole live output block** (round-034 FR-012/G8, ADR 0005 D7): the block owns `stderr` alone from `Begin` (before the child starts) to `End` (after the closing separator). On a **quiet/slow** command (a line, then a long computation with no output) there is **no animation at all** in that window.
- **Issue [#83](https://github.com/gosharplite/tellme/issues/83)** — *a second elapsed figure*. The spinner shows one figure, whole seconds since the **prompt** was captured (turn-scoped, never resets — round-019 D4): `⠋ Thinking [<model>]... (120s)`. The operator wants a **second** figure — the **current model call's** elapsed — appended: `⠋ Thinking [<model>]... (120s 21s)`.

## Terminology (added by the PR #84 review TD-6)

- **turn** — one prompt exchange: a submitted prompt and everything the run does for it (1..k AI-endpoint calls). The **total** elapsed is turn-scoped.
- **model call** (AI-endpoint call) — one request/response round to the provider. The **second** elapsed figure measures the **current model call**, and resets at each new call.

Two names for one concept is a defect; this round uses **"the current model call's elapsed"** for the second figure so the truth vocabulary does not contradict round 027's turn semantics.

## Operator-locked / review-confirmed decisions (clarify)

### Workstream B — the second elapsed figure (issue #83)

- **QB1 → Reset boundary = per AI-endpoint call (option A).** The second figure resets at the start of each **AI-endpoint call**, matching the round-027 turn semantics. Rejected: per waiting phase (B), per tool execution (C). *Mechanically clean:* `AgentLoop.Run` fires `notifyInferenceStart()` **exactly once per AI-endpoint call**, and the spinner already reacts to `OnInferenceStart`, so resetting the second timer there is per-call **by construction** with **no loop change** (verified against `agentloop.go:134-137` by the review). *(Operator-locked 2026-09-17.)*
- **QB2 → Both phases carry both figures.** The second figure appears in the model-wait label and in the tool-execution label (which also keeps its ` [CPU: x% | MEM: y%]` segment). In the tool phase the second figure also contains that call's model wait — accepted. *(Review-confirmed.)*
- **QB3 → Format `({total}s {call}s)`** — one ASCII space, both `s`-suffixed, inside the existing single parentheses. **Both figures are unlabelled** (matching the reference's unlabelled single figure); recorded as a third divergence in **ADR 0009** (QB4). *(Review-confirmed; QB4: unlabelled by design.)*

### Workstream A — liveness while the output block streams (issue #82)

- **QA1 → Mechanism: idle-gap resume.** Keep the block's `stderr` ownership, but do not stay hidden for the **whole** block: after **N** seconds with **no new output line**, the indicator resumes; the next output line clears it again. The two rejected options were verified: a permanent output-progress marker would mutate the pinned `[Tool Output]` block literals, and a per-line yield is round-034 G8's stated blocker. *(Review-confirmed.)*
- **QA2 → The idle gap N is a small fixed value** (assume **3 s**) with a **hermetic env seam** (`TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`; milliseconds; **`0` = admit immediately**; unset/invalid = the 3 s default). The **watcher poll period P** is pinned to the spinner's existing ~200 ms cadence, so the resume lands within `N … N+P`. *(Review-confirmed with the seam + poll period pinned — TD-4/TD-5.)*
- **QA3 → Scope guard.** Only the tool-execution phase of a **streaming `[Tool Output]` block** is affected; the model-wait phase, the block literals, the bounded-and-stopped semantics, the non-TTY/`-r` gates, and the `-i` surface are unchanged. *(Review-confirmed.)*

**Scope note**: two `stderr`-presentation changes to the round-019/025 progress spinner. Neither changes a CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record, the tool **result** fed to the model, or a `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A long-running command still shows a live indicator (Priority: P1)

As an operator, when a shell command runs for a while and streams its `[Tool Output]` block, I want the progress spinner to reappear during quiet stretches, so that I can tell the run is still alive instead of staring at a static header.

**Why this priority**: it is issue #82 — the anchor; the pause is the anchor's stated liveness gap.

**Independent verification**: unit-pin the spin cycle's resume-after-idle / clear-on-next-line behaviour under an injected `newTicker` seam (no `time.Sleep`) — including the **anti-vacuity** case (continuous output ⇒ **no** frame between lines) and the **race** case (a line arriving as the resume is admitted ⇒ no interleave, no residue); script an E2E run whose command prints one line and then stays quiet past the idle gap (a **child** `sleep`, not a Go sleep), and assert a spinner frame is drawn **within the block** on the captured `stderr`.

**Acceptance Scenarios**:

1. **Given** a terminal whose diagnostics are shown, **When** a command streams an output line and then produces no further output for at least the idle gap, **Then** the run shows the progress spinner again during that quiet stretch, and the next output line clears it before being written.
2. **Given** the same run, **When** the block closes, **Then** the indicator leaves **no** residue and the turn-scoped total elapsed is preserved.

**Functional Requirements**:

- **FR-001**: While a `[Tool Output]` block is open, the indicator MUST NOT remain hidden for the whole block: after **no new output line for the idle gap N** (default **3 s**, seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`, `0` = admit immediately, unset/invalid = the default; the watcher polls on the spinner's ~200 ms cadence, so the resume lands within `N … N+P`), the indicator MUST reappear; when the next output line arrives, the indicator MUST be cleared synchronously **before** that line is written.
- **FR-002**: The resume/clear MUST be **mutually exclusive + join** safe (the reviewer's TD-2 wording, not "single-writer"): (i) the resume is admitted only while holding the block writer's mutex; (ii) the clear is the presenter's synchronous `deactivate()` (a join: it waits for the redraw goroutine to exit, then erases every occupied row); (iii) lock order is **block-writer mutex → spinner mutex**. It MUST NOT interleave a spinner frame with a partial output line and MUST NOT drop or reorder any output byte. The unit stress MUST pin **both** directions: a line/clear race ⇒ no interleave, and **continuous** output ⇒ **no** frame between lines (anti-vacuity — proves the resume cannot decay into the rejected per-line yield).
- **FR-003**: The resumed indicator MUST keep the **turn-scoped total** elapsed (round-019 D4) and MUST leave **no** residue on block close (round-019/025/035).
- **FR-004**: The block literals (header `... (Output shown below)` + the 60-hyphen separator), the bounded-and-stopped semantics (round-024), and the non-TTY / `-r` gates MUST be unchanged.

### User Story 2 - The spinner shows the total time and the current model call's time (Priority: P2)

As an operator, I want the spinner to show both how long ago I submitted my prompt and how long the current model call has been running, so that I can separate a slow single call from a slow series of calls.

**Why this priority**: it is issue #83, an additive readability change to the same line; independent of US1.

**Independent verification**: unit-pin `FormatSpinnerLine` / the two-epoch arithmetic under an injected `now` seam (the second figure resets at each `OnInferenceStart`; the total never decreases); script a multi-call tool turn and assert the two figures in the captured frames.

**Acceptance Scenarios**:

1. **Given** a prompt that makes one AI-endpoint call, **When** the spinner renders, **Then** the line shows two whole-second figures — the total since prompt capture and the current model call's elapsed — and for a single-call turn the two are approximately equal.
2. **Given** a prompt that makes several AI-endpoint calls, **When** each call begins, **Then** the second figure **resets** to that call's own elapsed while the first keeps growing from the prompt epoch.

**Functional Requirements**:

- **FR-005**: The spinner line MUST carry **two** whole-second figures inside the existing parentheses: the **total** since prompt capture first, then the **current model call's** elapsed — `({total}s {call}s)`, both unlabelled.
- **FR-006**: The **total** figure MUST measure whole seconds since the turn's prompt-capture epoch and MUST **not** reset within a prompt (round-019 D4 preserved).
- **FR-007**: The **second** figure MUST measure whole seconds since the **current AI-endpoint call** began and MUST **reset at the start of each AI-endpoint call** (the per-call boundary of QB1; round-027 AI-call semantics).
- **FR-008**: Both figures MUST appear in the model-wait label **and** the tool-execution label; the tool-phase resource segment ` [CPU: <c>% | MEM: <m>%]` MUST be unchanged.

### Edge cases

- **A tool-less turn** (exactly one AI-endpoint call) → the two figures are approximately equal; no special-casing.
- **A quiet streaming command** (WS-A) → the indicator reappears after the idle gap; the total keeps growing and the second figure keeps counting **within its call**.
- **Continuous output** (WS-A) → no idle gap elapses, so the indicator stays cleared for the whole burst — a frame MUST NOT appear between lines (the anti-vacuity property; FR-002).
- **The AI-endpoint boundary coinciding with a resume** → the second figure resets on the next call's `OnInferenceStart`; the total is unaffected.
- **A resume attempt with no phase label** (reviewer RF-3) → `resume()` MUST be a defined **no-op** (the presenter has no status to draw), so the coordinator cannot silently disable the feature; the block always begins during a tool phase today (the loop fires `OnToolsStart` before the batch), which is where a label is present.
- **A very long turn** (both figures reach 3+ digits, e.g. `1234s 567s`) → the line is longer, so the round-025 **rune-based** `rowsForLine`/`eraseRows` bound MUST still erase every occupied row (no residue); re-witnessed for the two-figure 3+-digit frame (QB3).
- **A terminal resize mid-frame** → the round-025 recorded limitation (a stale row count may over-erase one row) is unchanged and out of scope.
- **Non-terminal `stderr`, `-r`, or the `-i` submit surface** → the spinner gate is unchanged; `-i`/`-r`/non-TTY behaviour is untouched.
- **A provider that reports no usage** → the spinner is orthogonal (the spinner is not the post-turn status group); unchanged.

### Key entities

- **Spinner line** — one rendered frame: `{braille}{status} ({total}s {call}s){resource}`.
- **Total elapsed epoch** — the turn's **prompt-capture** time; drives the first figure; never resets.
- **Call elapsed epoch** — the **current AI-endpoint call's** start time; drives the second figure; reset per call.
- **`[Tool Output]` block** — the live streamed shell output; the spinner must resume/clear around its quiet stretches (WS-A).

## Requirements *(mandatory)*

### Global requirements

- **FR-009**: The round MUST NOT change any CLI flag, exit code, the frozen class-phrase vocabulary, the tool surface, the provider transport, a persisted record shape (`history.jsonl` / `tokens.log` / `global_prompts.jsonl` / `tools-count.jsonl`), the tool **result** fed to the model, or `stdout` bytes.
- **FR-010**: The offline paths (`--version`, `-d`, `-l`, `--tool-usage`) MUST remain unchanged.
- **FR-011**: The spinner MUST remain gated to `isatty(stderr) && !-r`; a non-terminal diagnostic stream MUST draw no spinner, and the WS-A resume/clear MUST be a no-op when the spinner is gated off.

### Out of scope (recorded)

- Per-line spinner redraw inside the block (G8's stated blocker). WS-A resumes on an **idle gap**, not per line.
- Any change to the block literals, the bounded-and-stopped semantics, `output_file` handling, or the `-i` TUI surface.
- The reference divergences: `tell-me-go` shows a **single, per-phase-restarting** figure and does not re-activate a spinner inside its streamed output — both recorded divergences (ADR 0009 D5), plus the **unlabelled** two-figure format (ADR 0009, QB4).

## Success criteria *(mandatory)*

- **SC-001**: Unit pins (injected clock) show the two-figure arithmetic resets the **second** figure per AI-endpoint call while the total never decreases (US2); and (injected ticker) the spinner reappears after the idle gap and clears on the next output line, with no residue and the total preserved, **including the anti-vacuity case (continuous output ⇒ no frame between lines)** (US1).
- **SC-002**: A scripted E2E run with a **real** idle gap — a small **nonzero** forced threshold (e.g. `50` ms) plus a scripted command that prints one line then a **child** `sleep` (not a Go `time.Sleep`) whose quiet stretch **exceeds `N + 2·P` with margin** (e.g. a 2 s child `sleep` at `N=50 ms`, `P≈200 ms`) — shows a spinner frame **inside** the block window on the captured `stderr` (US1).
- **SC-003**: A scripted E2E multi-call turn shows the **two-figure shape** (`({total}s {call}s)`); the per-call **reset arithmetic** (the second figure dropping at a call boundary) is carried by the **unit** pin (an injected clock) — it is **not** E2E-observable on a fast scripted turn (whole seconds need a wait a test may not take; there is no subprocess clock seam). *(Reworded by round-040 review **TD-5** to match the accepted witness plan — `chat/dsl.md` / `research.md` D6.)*
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit (re-run after the truth MODIFY) are green **at delivery** — the **godog `Strict: true`** flag (undefined steps fail) lands in the **implementation half** with the stepdefs, witnessed by the FAIL→PASS transition; the plan half leaves the harness unchanged so `dev` stays green.
- **SC-005**: The change has **falsifiability witnesses**: removing the idle-gap resume fails the WS-A Example; freezing the second figure (no per-call reset) fails the WS-B Example; widening the line to two 3+-digit figures without the row-aware clear strands a residue frame — each reproduced then reverted.
- **SC-006**: The row-aware clear is re-witnessed on the **two-figure, 3+-digit** frame (`1234s 567s`) — the erase-all-rows pin exercised on the longer line (QB3).

## Assumptions

- **Truth ownership.** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md` are owned by `/axb-dsl-refine`; `techstack.md` by `/axb-technical-research`; **ADR 0009** is governance, not truth. `/axb-api-plan` is a checked **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted-state change). All edits are recorded in this round's `truth-delta.md`.
- **Supersession.** Round-034 FR-012/G8 and **ADR 0005 D7** (the whole-block pause) are `Accepted`/frozen → WS-A's change is recorded via **ADR 0009**, which **supersedes ADR 0005 D7 only** (the rest of 0005 stands; its body is not edited and its overall `Status` stays `Accepted`; the supersession is named in ADR 0009 **and annotated in the decisions index** — reviewer RF-2). Round-019's D4 ("the elapsed epoch … never reset") is **amended** by WS-B.
- **Frozen history.** Every earlier plan package (incl. `019-turn-spinner`, `025-spinner-width-safety`, `034-tool-call-log-parity`, `035-spinner-tail-residue`) is never edited; 040 is a fresh package.
- **Enabler.** The idle-gap seam (`TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`) exists so E2E can force the threshold without a wall-clock wait (round-016 `TELL_ME_TUI_DEBOUNCE` precedent); the E2E's own quiet stretch is a **child** `sleep`, so `verify-no-test-sleep` is untouched.
