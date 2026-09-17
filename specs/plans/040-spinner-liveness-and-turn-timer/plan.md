# Plan — round 040 (spinner liveness while a command streams + a per-model-call elapsed timer)

Plan package: `specs/plans/040-spinner-liveness-and-turn-timer`

## Theme

Two folded workstreams on the round-019/025 **progress spinner** (`internal/ui/spinner.go`, drawn on the `stderr` diagnostic stream):

1. **#82 — liveness while a command's `[Tool Output]` streams.** Today the spinner is **paused for the whole live output block** (round-034 FR-012/G8, ADR 0005 D7), so a quiet/slow command shows no animation at all. Round 040 keeps the block's `stderr` ownership but **resumes the indicator after an idle gap** (no new output line for N seconds; the invariant is **mutual exclusion + join**, not "single-writer" — admit under the block-writer mutex, the clear is the goroutine-joined `deactivate()`, lock order block-mutex → spinner-mutex) and **synchronously clears it before the next output line**.
2. **#83 — a dual elapsed timer.** The line gains a **second** whole-second figure — the **current model call's** elapsed, reset at each **AI-endpoint call** — alongside the existing total since prompt capture: `({total}s {call}s)` (both unlabelled).

Presentation only — no flag, exit code, frozen vocabulary, tool result, transport, persisted record, or `stdout` byte changes (`spec.md` FR-009–FR-011).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (diagnostic stream)** — the `chat` module's progress spinner (`internal/ui/spinner.go`) and the `[Tool Output]` block it yields to | `cli` | The subject of the round. The spinner line carries **two** figures (total + per-AI-endpoint-call turn time); during a `[Tool Output]` block the indicator **resumes after an idle gap** and is **cleared before the next line**. `stdout` stays byte-exact; the block literals, the bound/stop semantics, the `isatty(stderr) && !-r` gate, the `-i` surface, the class-phrase vocabulary, and the tool schemas are unchanged. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change** (the spinner and the block are `stderr`-only presentation and are never persisted).

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies: `/axb-api-plan` and `/axb-data-plan` NOOP; `/axb-ui-plan` is **skipped** — no TUI chrome/screen change), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` **MODIFIES** `chat/presenting-the-progress-spinner.feature`: it **replaces** `Rule: The spinner is paused while a command's output streams` with `Rule: A quiet command's output still shows the progress spinner` (the indicator is not hidden for the whole block — it resumes after an idle gap and clears on the next line) and **adds** `Rule: The spinner shows the total time and the current model call's time`, plus their `chat/dsl.md` rows.

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `~/.tellme/*.jsonl` record shapes: no persisted-state change — the spinner and the `[Tool Output]` block are `stderr`-only presentation and are never persisted (`spec.md` FR-009; `research.md` D7).
- **`/axb-ui-plan` → skipped.** The spinner is a single `stderr` status line, not TUI chrome; no screen/keybinding/state-transition artifact applies (`spec.md` assumptions).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the two axes (`spec.md` FR-001–FR-008; `research.md` D1–D5):

- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — **REPLACE** `Rule: The spinner is paused while a command's output streams` with `Rule: A quiet command's output still shows the progress spinner`:
  - the rule is narrowed: the indicator is **not** hidden for the whole block — it **resumes** during a quiet stretch (no new output line for the idle gap `N`; seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`) and is **cleared before the next output line** (`spec.md` FR-001–FR-004).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — **ADD** `Rule: The spinner shows the total time and the current model call's time`:
  - `Then the progress spinner shows the time since the prompt and the current model call's time` (single-call and tool-using cases);
  - the current model call's elapsed resets when a new model call begins (`spec.md` FR-005–FR-008).
- `specs/truth/features/cli/chat/dsl.md` — the streaming row is **modified in place** to `the run shows the progress spinner again while the command stays quiet`; **ADD** a `## Given (round 040)` (the state Given `the command stays quiet for longer than the spinner's idle gap` — seam `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS`, ms, `0` = admit immediately, unset/invalid = 3 s — + a quiet-command provider Given) and a `## Then (round 040)` (`the progress spinner shows the time since the prompt and the current model call's time`), each matching exactly one step (`dsl-exact-one-match`); `dsl-single-authority` is preserved (no duplication).

The witness set is **unit** pins (the two-figure arithmetic under an injected `now`; the idle-gap resume/clear under an injected ticker, incl. the **race** + **anti-vacuity** cases; the row-aware clear for the longer two-figure 3+-digit line) **plus** the E2E `Then`s (a **real** idle gap — a small nonzero forced threshold + a **child** `sleep` whose quiet stretch exceeds `N + 2·P` with margin; a multi-call turn's two figures) — `research.md` D6.

`truth-delta.md` carries the explicit rows; `specs/truth/techstack.md` (**Turn progress spinner** row) and **ADR 0009** (superseding ADR 0005 **D7**, amending round-019 D4) are owned by `/axb-technical-research` (already folded).

## Acceptance-rule → carrier mapping (round-040 review RF-1)

Four of the six new round-040 acceptance rules are carried by a truth Rule with the same subject; the two "*both times while tools run / a single-call answer shows both times*" journeys and one "*the indicator leaves no trace*" leg reuse Examples under Rules whose named subject is the tool label/resource (rounds 034/039 precedent — recorded here because the row→feature audit blind spot is already a `#60` item).

| Acceptance rule (plan package) | Truth carrier (`specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`) |
| --- | --- |
| A quiet stretch of a command's output shows the progress indicator again | `Rule: A quiet command's output still shows the progress spinner` — Example *The indicator returns while a long-running command stays quiet* |
| The indicator leaves no trace once the command finishes | the same Example (the whole-stream `the run shows no progress spinner` Then) |
| The indicator is not shown when the operator is not watching a terminal | the interface-root `the run shows no progress spinner` negative (unchanged) |
| The indicator shows the total wait and the current model call's time | `Rule: The spinner shows the total time and the current model call's time` — Example *A single-call answer shows both times* |
| A tool-using answer restarts the model-call time for each call | the same Rule — Example *A tool-using answer shows both times on each model call* (the per-call reset is a **unit** pin) |
| Both times are shown while tools run, alongside the resource usage | the two amended tool-phase Examples (`Rule: The spinner names the tools and the resources while tools run`) — cross-Rule carrier, recorded |

## Task-list directives (for `/axb-tasks` — reviewer RF-1)

The task list MUST carry, beyond the four new stepdefs:

- **[BDD-REMOVE]** retire the now-dead stepdef `tests/e2e/steps/step_r034_t016_chat_then_no_spinner_while_streaming.go` (0 matching feature steps after the Rule replacement; **keep** the helpers `toolOutputBlockIndexes` / `hasSpinnerStatusBetween` in `toolcall_log.go` — the round-040 liveness Then reuses them at positive polarity, bound on the **closing separator** (PR #85 fold B2/R-1; the helpers are **not** removed).
- **[BDD-ALIGN]/[BDD-RED]** the four new sentences of D9's burn-down list.
- the **coordinator extraction** (the `internal/ui` sink+spinner coordinator), the **lock-order comment**, the **race/no-interleave + anti-vacuity unit stress**, the **dual-timer arithmetic + injected-clock pins**, and the **two-figure 3+-digit row-aware-clear re-witness** (QB3).
- **add `Strict: true` to `tests/e2e/suite_test.go`** (TD-1) **in the implementation half** — the plan half leaves the harness as-is so `dev` never goes red (`make test` = `go test ./...`, which includes `tests/e2e`); the implement half lands it together with the four stepdefs, witnessed by the FAIL-then-PASS transition.
- **TD-10 (mechanism + timing)**: name the **resume-only** admission path (e.g. an explicit `admitResume()` / a resume flag) so `activate()`'s synchronous first frame is preserved for `OnInferenceStart`/`OnToolsStart` (rounds 019/025/034/035 depend on it); the resumed first frame is drawn **immediately on start** (render once before waiting for the tick), i.e. not one more poll later.
- **`Turn progress spinner` row carve-out**: the initial activation stays synchronous; the **in-block resume** draws on the goroutine.
