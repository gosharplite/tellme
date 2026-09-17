# Tasks — round 038 (tool-output sanitization + empty-submit no-op)

Plan package: `specs/plans/038-tool-output-sanitize-and-empty-submit`

Task markers: `[UNIT]` unit test · `[BDD-RED]` new/aligned stepdef that must fail before the product change · `[BDD-GREEN]`/`[BDD-REFACTOR]` feature green/refactor · `[REGRESSION]` regression + falsifiability.

## Setup

_(none — stdlib-only; no new dependency, no SDL/build change.)_

## Foundational

- [ ] **T001** — Add the sanitizer + neutral-restore seam in `internal/ui/tooloutput.go`: a `ToolOutputReset` constant (`"\x1b[0m"`), a pure `sanitizeControl(string) string`, applied inside `FormatToolOutputLine`, and the restore emitted before the closing separator in `End()`. No wiring change (the CLI already binds the writer). `[scaffold]`
- [ ] **T002** — Add the `trySubmit()` seam in `internal/ui/tui/prompt/model.go` (returns `tea.Quit` only for a non-empty editor; `nil` otherwise) and route the `Ctrl+S`/`Alt+Enter` branches through it. `[scaffold]`

## Phase 3 — Test alignment & implementation (Red first)

- [ ] **T003** `[UNIT]` — `internal/ui/tooloutput_sanitize_test.go`: the hostile fixture — SGR-set-no-reset removed; SGR-set+reset removed but visible text preserved; OSC title removed; cursor-hide (`\x1b[?25l`) removed; stray BEL removed; TAB preserved; UTF-8 text preserved; a sequence split across two `Write` calls removed (assembled-line scope). **RED** before T001.
- [ ] **T004** `[UNIT]` — `internal/ui/tooloutput_test.go` (extend): `End()` writes the neutral restore before the closing separator (normal close + the `End()`-only/never-`Begin` path). **RED** before T001.
- [ ] **T005** `[UNIT]` — `internal/ui/tui/prompt/model_submit_test.go`: `TestModelEmptySubmitDoesNotQuit` (the `Update` command is `nil` for an empty `Ctrl+S` and an empty `Alt+Enter`), `TestModelNonEmptySubmitQuits`, and `TestModelAbortStillQuits` (Esc/Ctrl+C). **RED** before T002.
- [ ] **T006** `[BDD-RED]` — E2E `Given`s for the two colouring commands (in `tests/e2e/steps/step_r038.go`): `…runs a colouring command and then answers with "{answer}"` (SGR set, no reset, newline-terminated) and `…runs a colouring command that is stopped at its time limit and then answers with "{answer}"` (colour escape then `sleep`, small `timeout`). **RED** (the steps are undefined) before implementation.
- [ ] **T007** `[BDD-RED]` — E2E `Then` `the run streamed the command's output free of terminal control sequences` + `the terminal is left in its default state` (helpers in `tests/e2e/steps/r038_terminal.go`): content lines carry no ESC/C0 byte; a restore follows the last content line. **RED** before T001.
- [ ] **T008** `[BDD-RED]` — E2E `When` `the operator submits an empty prompt then submits the prompt "{prompt}" at the interactive prompt` (scripted keys `submit → type → submit`). Reuses the existing shown/request/answer Thens. **RED** before T002.
- [ ] **T009** — Review gate: after T003–T008, every new step matches exactly one `DSLRow` and the suite reports **0 undefined** steps.

## Phase 4 — Feature green / refactor

- [ ] **T010** `[BDD-GREEN]` — `watching-the-tool-loop.feature` (the new Rule) green: sanitizer + neutral restore (T001).
- [ ] **T011** `[BDD-REFACTOR]` — tidy the sanitizer (single responsibility, no allocation on the happy path — `strings.ContainsRune` fast-path) under green protection.
- [ ] **T012** `[BDD-GREEN]` — `using-the-interactive-prompt.feature` (the new Rule) green: `trySubmit` (T002).
- [ ] **T013** `[BDD-REFACTOR]` — tidy the prompt key handling under green protection.

## Regression & falsifiability

- [ ] **T014** `[REGRESSION]` — `make verify` + the full godog suite green; `stdout` byte-exact; the existing `[Tool Output]` / interactive-prompt Examples unchanged-green.
- [ ] **T015** — Falsifiability witnesses (each reproduced then reverted): remove `sanitizeControl` (identity) → T003 + the tool-output E2E fail; remove the `End()` restore → T004 + `the terminal is left in its default state` fail; restore the unconditional `tea.Quit` → T005 + the empty-submit E2E fail.
- [ ] **T016** — Update `STATUS.md` + the daily summary; open the PR; close **#78** and **#76** on delivery.
