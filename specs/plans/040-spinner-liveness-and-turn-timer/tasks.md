# Tasks — round 040 (spinner liveness while a command streams + a per-model-call elapsed timer)

Plan package: `specs/plans/040-spinner-liveness-and-turn-timer`

Task markers: `[UNIT]` unit test · `[BDD-ALIGN]` realign an existing stepdef to the current truth · `[BDD-REMOVE]` retire a stepdef/assertion no longer in truth · `[BDD-RED]` new stepdef that must fail before the product change · `[BDD-GREEN]`/`[BDD-REFACTOR]` feature green/refactor · `[CODE-REMOVE]` delete dead code · `[REGRESSION]` regression + falsifiability.

## Core Inputs (must read)

- `spec.md` (FR-001..FR-011, SC-001..SC-006), `research.md` (D1–D9), `plan.md` (incl. the acceptance→carrier table + the task directives), `truth-delta.md`.
- Truth: `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (one Rule **replaced**, one Rule **added**, two tool-phase Examples amended; header refreshed), `specs/truth/features/cli/chat/dsl.md` (the streaming row modified **in place** + `## Given (round 040)` + `## Then (round 040)`), `specs/truth/techstack.md` (**Turn progress spinner** row — MODIFY, incl. the round-019 carve-out).
- Governance: `docs/decisions/0009-spinner-dual-timer-and-streaming-liveness.md` (ADD; supersedes **ADR 0005 D7 only**; amends round-019 D4) + the ADR 0005 index annotation.
- Code seams: `internal/ui/spinner.go` (`Spinner`, `FormatSpinnerLine`, `activate`/`resume`/`deactivate`/`loop`/`renderLocked`/`eraseRows`/`rowsForLine`, the `newTicker`/`now` seams), `internal/ui/tooloutput.go` (`ToolOutputWriter`), `internal/cli/cli.go` (:738–:768 the spinner + the `ToolOutputSink` closure; `newTurnSpinner`/`spinnerGate`/`stderrColumns`), `internal/cli/composite_observer.go`, `internal/infrastructure/tools/command.go` (`runCaptured` Begin/End), `tests/e2e/steps/toolcall_log.go` (`toolOutputBlockIndexes`/`hasSpinnerStatusBetween`).

## Setup

_(none — stdlib-only; no new dependency, no build/SDL change.)_

## Foundational

- [ ] **T001** — Add the round-040 E2E stepdef landing skeletons as **independent files** (Zero Shared Edits): `tests/e2e/steps/step_r040_idle_gap.go` (register the idle-gap `Given`, the quiet-command provider `Given`, and the liveness `Then`) and `tests/e2e/steps/step_r040_dual_timer.go` (register the dual-timer `Then`), each with stub bodies returning `nil`. **只做** registration + stubs; **不做** the assertions or the fixtures. `[scaffold]`
- [ ] **T002** — Add the unit-test landing files + the `internal/ui` coordinator seam **skeleton**: `internal/ui/spinner_round040_test.go` and `internal/ui/coordinator_test.go` (empty test funcs), and a new `internal/ui/coordinator.go` with the type/constructor skeleton (the `ToolOutputWriter` + `*Spinner` + the idle seams; no behaviour). **只做** skeletons; **不做** the resume/clear logic or the assertions. `[scaffold]`

## Phase 3 — Test alignment & implementation (Red first)

- [ ] **T003** `[BDD-REMOVE]` — retire the dead stepdef `tests/e2e/steps/step_r034_t016_chat_then_no_spinner_while_streaming.go` (0 matching feature steps after the Rule replacement) **and** its now-unused helpers `toolOutputBlockIndexes` / `hasSpinnerStatusBetween` in `tests/e2e/steps/toolcall_log.go` (grep-confirmed unused after T003). **RED**: the suite must compile and report **no** undefined step from the retired sentence. No product code.
- [ ] **T004** `[BDD-RED]` — `tests/e2e/steps/step_r040_idle_gap.go`: implement the `Given` `the command stays quiet for longer than the spinner's idle gap` (set `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS` to a small **nonzero** value, e.g. `50`, alongside `TELL_ME_FORCE_STDERR_TTY`; the child's quiet stretch **exceeds `N + 2·P` with margin** — a **child** `sleep 2`, never a Go `time.Sleep`, so `verify-no-test-sleep` is untouched), the `Given` `a configured provider "…" whose endpoint runs a command that prints a line and then stays quiet and then answers with "…"`, and the `Then` `the run shows the progress spinner again while the command stays quiet` (a braille frame + a phase status **inside** the `[Tool Output]` block's span; reuse the whole-stream `the run shows no progress spinner` for the no-residue leg). **RED** before T009.
- [ ] **T005** `[BDD-RED]` — `tests/e2e/steps/step_r040_dual_timer.go`: implement the `Then` `the progress spinner shows the time since the prompt and the current model call's time` (the elapsed segment is **two** whole-second figures inside one parentheses, `({total}s {call}s)`, the first the total since prompt capture, the second the current model call's elapsed; both unlabelled). Covers the ×4 occurrences (2 amended tool-phase Examples + the 2 new Examples). **RED** before T011.
- [ ] **T006** `[UNIT]` — `internal/ui/spinner_round040_test.go`: the **dual-timer arithmetic** under an injected `now` — a first `OnInferenceStart` then a render shows `(t t)`; a later render shows `(t+k t+k)`; a second `OnInferenceStart` shows the **total grown** but the **second figure reset** (per-call `callEpoch`; the total never decreases). **RED** before T011.
- [ ] **T007** `[UNIT]` — `internal/ui/coordinator_test.go`: the WS-A **race + anti-vacuity + no-residue** stress under an injected ticker — (a) a line arriving as the resume is admitted ⇒ the frame is cleared **before** the line (no interleave); (b) **continuous** output ⇒ **no** frame between lines (anti-vacuity — the resume cannot decay into the rejected per-line yield); (c) no residue survives the block close. Plus the round-025 **row-aware clear on the two-figure 3+-digit** frame `1234s 567s` (SC-006). **RED** before T009/T010.
- [ ] **T008** — Review gate: after T003–T007, every new step matches exactly one `DSLRow` and the suite reports **0 undefined** steps. **Add `Strict: true` to `tests/e2e/suite_test.go` in this same change** (TD-1 — `godog.Options.Strict` defaults to `false`, so undefined steps are reported-and-ignored; enabling it makes the round's new Examples fail the suite until the stepdefs land — the FAIL→PASS transition is the witness). The plan half left the harness unchanged so `dev` stays green; this task lands the flag with the stepdefs.

## Phase 4 — Feature green / refactor

- [ ] **T009** `[BDD-GREEN]` — `presenting-the-progress-spinner.feature` Rule *A quiet command's output still shows the progress spinner* green (WS-A): extract the `internal/ui` **coordinator** (the `ToolOutputWriter` + the `*Spinner` + the idle seams) so `internal/cli` passes **one** object to the sink (the `#69` pay-down); a block-scoped **idle watcher** polls on the spinner's ~200 ms cadence and **resumes** after `N` with no new line; on each complete output line the indicator is **cleared synchronously first** (`deactivate()`), then the line is written; `End` stops the watcher, writes the closing separator, and resumes. The resume/clear is **mutual exclusion + join** — the resume is **admitted only under the block-writer's mutex**; the clear is the goroutine-joined `deactivate()`; lock order block-mutex → spinner-mutex. The resume uses a **named resume-only admission path** (`admitResume()`), the **redraw goroutine draws the resumed first frame immediately on start** (once before waiting for the tick), and `activate()` keeps its **synchronous** first frame (rounds 019/025/034/035 depend on it). The idle seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS` (ms; `0` = admit immediately; unset/invalid = the 3 s default) resolves once at construction. Test Scope: the Rule's Example + the unchanged `the run streamed the command's output …` Examples.
- [ ] **T010** `[BDD-REFACTOR]` — tidy the coordinator + the `internal/cli` wiring under green protection: one blank-free resume path, the **lock-order comment**, no frame write inside the block critical section, and the single-owner seam (a no-op when the spinner is gated off). Test Scope: the WS-A Rule green + the round-019/025/034/035 spinner Examples remain green.
- [ ] **T011** `[BDD-GREEN]` — `presenting-the-progress-spinner.feature` Rule *The spinner shows the total time and the current model call's time* green (WS-B): `FormatSpinnerLine(frame, status string, total, call int, resource string)` renders `({total}s {call}s)`; the `Spinner` gains an internal **`callEpoch`** (initialised to `epoch`, **not** a sixth ctor param — round-019 R-2) set in `OnInferenceStart` **before** activating; `OnToolsStart`/`BeforeToolLog`/`AfterToolLog`/the WS-A resume preserve it. Test Scope: the Rule's 2 Examples + the 2 amended tool-phase Examples.
- [ ] **T012** `[BDD-REFACTOR]` — tidy the formatter + the epoch handling under green protection (one elapsed-formatting site; no allocation regression; the resource segment unchanged). Test Scope: the WS-B Rule green + the tool-phase Examples green.
- [ ] **T013** `[CODE-REMOVE]` — remove the retired **whole-block pause** product path: the unconditional per-call pause in the `internal/cli` sink closure is superseded by the coordinator (T009); delete the now-dead closure wiring and any branch that assumed the spinner stays hidden for the whole block, keeping the block literals + the bound/stop semantics intact. Test Scope: the retired Rule is gone (no feature step requests it); the `[Tool Output]` Examples stay green.

## Regression & falsifiability

- [ ] **T014** `[REGRESSION]` — `make verify` + the full godog suite green (now `Strict: true`); `stdout` byte-exact; every existing spinner/tool-loop Example unchanged-green (the round-019/025/034/035/039 pins). Re-run the Gherkin/DSL topology audit (`--root specs/truth/features/cli`) and confirm PASSED.
- [ ] **T015** — Falsifiability witnesses (each reproduced then reverted): (a) remove the idle-gap resume → the WS-A Example fails; (b) freeze the second figure (no per-call reset) → the WS-B Example fails; (c) drop the row-aware clear → a residue frame survives the 3+-digit frame → the SC-006 pin fails.
- [ ] **T016** — Update `STATUS.md` + the daily summary; open the implementation PR; close **#82** + **#83** at closeout (delivery).

## Boundary

- In scope: `internal/ui/spinner.go`, a new `internal/ui/coordinator.go`, `internal/ui/tooloutput.go`, `internal/cli/cli.go` (the sink/spinner wiring + the idle seam), `tests/e2e/steps/step_r040_*.go` (+ the `[BDD-REMOVE]` of the dead stepdef + its helpers), `tests/e2e/suite_test.go` (`Strict: true`), and the `internal/ui` unit tests.
- Out of scope (unchanged): the `[Tool Output]` header/separator literals, the block bound/stop semantics, `output_file`, the tool **result** fed to the model, the call's stored arguments, `stdout`, `-r`, flags, exit codes, the class-phrase vocabulary, the provider transport, every persisted record, and the offline paths (`spec.md` FR-009/FR-010/FR-011). No new dependency (stdlib-only; POSIX-only).

## Parallel Hint

- T001/T002 are independent landing files (Zero Shared Edits). T006/T007 are independent unit files. T004/T005 are independent step files. T003 (the removal) must land with T008 (`Strict: true`) so the suite is defined. T009–T013 touch product files (`internal/ui`, `internal/cli`) and must serialize (T009 → T010; T011 → T012; T013 after T009).

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP rows: `techstack.md` spinner row (Read by T009/T011 + delivered there), the two feature Rules + the `chat/dsl.md` rows (Read by T004/T005 and delivered by T009–T013), ADR 0009 (governance, ADD — no task action). **Covered.**
- `research.md` D1–D9: D1/D2 (WS-B timer) → T006/T011; D3/D4/D5 (WS-A mechanism/protocol/scope) → T007/T009/T010; D6 (witnesses) → T004–T007/T014/T015; D7 (truth scope) → T009–T013; D8 (ADR lifecycle) → no task action (governance); D9 (harness `Strict`) → T008. **Covered.**
- `techstack.md` round-040 section (the dual timer + the idle-gap liveness + the carve-out) → T009/T011 + the `techstack.md` Read in Core Inputs. **Covered.**
- **Orphans: 0.**
