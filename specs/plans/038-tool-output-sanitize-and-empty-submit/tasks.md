# Tasks — round 038 (tool-output sanitization + empty-submit no-op)

Plan package: `specs/plans/038-tool-output-sanitize-and-empty-submit`

Task markers: `[UNIT]` unit test · `[BDD-RED]` new/aligned stepdef that must fail before the product change · `[BDD-GREEN]`/`[BDD-REFACTOR]` feature green/refactor · `[REGRESSION]` regression + falsifiability.

## Setup

_(none — stdlib-only; no new dependency, no SDL/build change.)_

## Foundational

- [X] **T001** — Add the sanitizer + neutral-restore seam in `internal/ui/tooloutput.go`: a `ToolOutputReset` constant (`"\x1b[0m"`), a pure `sanitizeControl(string) string`, applied inside `FormatToolOutputLine`, and the restore emitted before the closing separator in `End()`. No wiring change (the CLI already binds the writer). `[scaffold]`
- [X] **T002** — Add the `trySubmit()` seam in `internal/ui/tui/prompt/model.go` (returns `tea.Quit` only for a non-empty editor; `nil` otherwise) and route the `Ctrl+S`/`Alt+Enter` branches through it. `[scaffold]`

## Phase 3 — Test alignment & implementation (Red first)

- [X] **T003** `[UNIT]` — `internal/ui/tooloutput_sanitize_test.go`: the hostile fixture — SGR-set-no-reset removed; SGR-set+reset removed but visible text preserved; OSC title removed; cursor-hide (`\x1b[?25l`) removed; stray BEL removed; TAB preserved; UTF-8 text preserved; a sequence split across two `Write` calls removed (assembled-line scope). **RED** before T001.
- [X] **T004** `[UNIT]` — `internal/ui/tooloutput_test.go` (extend): `End()` writes the neutral restore before the closing separator (normal close + the `End()`-only/never-`Begin` path). **RED** before T001.
- [X] **T005** `[UNIT]` — `internal/ui/tui/prompt/model_submit_test.go`: `TestModelEmptySubmitDoesNotQuit` (the `Update` command is `nil` for an empty `Ctrl+S` and an empty `Alt+Enter`), `TestModelNonEmptySubmitQuits`, and `TestModelAbortStillQuits` (Esc/Ctrl+C). **RED** before T002.
- [X] **T006** `[BDD-RED]` — E2E `Given`s for the two colouring commands (in `tests/e2e/steps/step_r038.go`): `…runs a colouring command and then answers with "{answer}"` (SGR set, no reset, newline-terminated) and `…runs a colouring command that is stopped at its time limit and then answers with "{answer}"` (colour escape then `sleep`, small `timeout`). **RED** (the steps are undefined) before implementation.
- [X] **T007** `[BDD-RED]` — E2E `Then` `the run streamed the command's output free of terminal control sequences` + `the terminal is left in its default state` (helpers in `tests/e2e/steps/r038_terminal.go`): content lines carry no ESC/C0 byte; a restore follows the last content line. **RED** before T001.
- [X] **T008** `[BDD-RED]` — E2E `When` `the operator submits an empty prompt then submits the prompt "{prompt}" at the interactive prompt` (scripted keys `submit → type → submit`). Reuses the existing shown/request/answer Thens. **RED** before T002.
- [X] **T009** — Review gate: after T003–T008, every new step matches exactly one `DSLRow` and the suite reports **0 undefined** steps.

## Phase 4 — Feature green / refactor

- [X] **T010** `[BDD-GREEN]` — `watching-the-tool-loop.feature` (the new Rule) green: sanitizer + neutral restore (T001).
- [X] **T011** `[BDD-REFACTOR]` — tidy the sanitizer (single responsibility, no allocation on the happy path — `strings.ContainsRune` fast-path) under green protection.
- [X] **T012** `[BDD-GREEN]` — `using-the-interactive-prompt.feature` (the new Rule) green: `trySubmit` (T002).
- [X] **T013** `[BDD-REFACTOR]` — tidy the prompt key handling under green protection.

## Regression & falsifiability

- [X] **T014** `[REGRESSION]` — `make verify` + the full godog suite green; `stdout` byte-exact; the existing `[Tool Output]` / interactive-prompt Examples unchanged-green.
- [X] **T015** — Falsifiability witnesses (each reproduced then reverted): remove `sanitizeControl` (identity) → T003 + the tool-output E2E fail; remove the `End()` restore → T004 + `the terminal is left in its default state` fail; restore the unconditional `tea.Quit` → T005 + the empty-submit E2E fail.
- [X] **T016** — Update `STATUS.md` + the daily summary; open the PR; close **#78** and **#76** on delivery.

## Outcomes (implementation)

- **T001/T002** done — `sanitizeControl` + `ToolOutputReset` + the `End()` restore in `internal/ui/tooloutput.go`; `trySubmit()` in `internal/ui/tui/prompt/model.go`. `escSequenceLen` was split (`csiLen`/`oscLen`/`genericEscLen`) to satisfy the `cyclop` lint gate (CC ≤ 15).
- **T003–T005** RED→GREEN — `internal/ui/tooloutput_sanitize_test.go` (9-case hostile fixture + split-across-writes + neutral-close + start-failure) and `internal/ui/tui/prompt/model_submit_test.go` (empty/whitespace no-op; non-empty quits; abort quits).
- **T006–T008** RED→GREEN — `tests/e2e/steps/step_r038.go` (2 colouring-command Givens, the empty-then-real-submit When, the 2 Thens + helpers).
- **T009** — every new step matched exactly one `DSLRow`; suite **218 scenarios / 0 undefined** (was 215).
- **T010–T013** — the two new Rules green; sanitizer/key-handling refactored under green.
- **T014** — `make verify` **OK** (lint 0 issues) · `go test -count=1 ./...` green (E2E **218/218**).
- **T015** — **falsifiability witnesses** (each reproduced then reverted): (a) sanitizer disabled → the split-escape unit pin + both tool-output E2E Examples fail (the pre-fix stream showed `[Tool Output] \x1b[31mERROR: failed` verbatim); (b) the `End()` restore removed → the start-failure unit pin + `the terminal is left in its default state` fail; (c) the unconditional `tea.Quit` restored → the whitespace-only unit pin + the empty-submit E2E Example fail.
- **T016** — `STATUS.md` + daily summary updated; PR opened; #78/#76 close at merge (closeout Step 8).


## Round-038 review fold (PR #79)

- **B1** [blocker] — ASCII-gated the ESC final byte (`genericEscLen` consumes a final byte only when `< 0x80`), so ESC before a multi-byte rune can no longer decapitate it into invalid UTF-8; added the `"日本"` / `"aé"` fixtures with `utf8.ValidString`, and added `utf8.ValidString` to the E2E content-line predicate (the colouring command now carries a multibyte word).
- **B2** [blocker, governance] — added **ADR 0007** (`docs/decisions/0007-terminal-control-sanitization.md`) + the index row, carrying D1–D6, the precise class (7-bit ESC/C0/DEL; C1 out of scope; interior CR dropped), the ASCII gate + bounded consumption, and the TD-1/TD-2 scope boundaries.
- **TD-1** — homed the broader "sanitize every `[Tool …]` formatter" policy on live issue **[#80](https://github.com/gosharplite/tellme/issues/80)** (cited in ADR 0007 + the `techstack.md` row).
- **TD-2** — decided to keep the neutral restore **unconditional** (stated in ADR 0007 D4, the `ToolOutputReset` code comment, and the `techstack.md` row; the `isatty`-gating alternative is recorded as rejected).
- **RF-1/RF-2** — bounded the CSI/OSC scanners (`escScanLimit = 64`) so an unterminated sequence cannot swallow a line's text; the class definition (7-bit only; C1 untouched; interior CR dropped) is stated in the code comment + ADR.
- **RF-3** — tightened the round-023 `TestModelEmptySubmitKeepsFrame` pin to assert the returned command is nil (the observable that carries the control-flow half), and documented the ownership split.
- **N-1/N-2/N-3** — fixed the `truth-delta.md` cross-reference (`using-the-interactive-prompt.feature`), corrected the witness names in `research.md`, and anchored the E2E neutral-close predicate to the block's close line.

## Round-038 fold review #2 (PR #79)

- **TD-3** [blocker-node, folded] — the single `escScanLimit = 64` capped **terminated** sequences too, so a long-but-valid sequence (an OSC-8 hyperlink URL longer than ~57 chars, a long OSC title, a long SGR run) was no longer removed and its parameters printed as visible text — contradicting ADR 0007 D2. Replaced with **per-kind** bounds (`csiScanLimit = 128`, `oscScanLimit = 1024`): a terminated sequence inside its window is removed in full; only a terminator-beyond-window / unterminated sequence falls back to dropping the ESC. Added `TestFormatToolOutputLineStripsLongTerminatedSequences` (OSC-8 long URL, 400-byte OSC title, long SGR run); `TestToolOutputWriterBoundsUnterminatedSequence` stays green (the 200-byte blob finds no terminator within the CSI window). ADR 0007 D2 updated to the per-kind wording.
- **N-4** — narrowed "the result is always valid UTF-8" to the true claim **"never introduces invalid UTF-8"** (the sanitizer only removes bytes; it forwards the rest verbatim, so C1 `0x9B` / raw continuation bytes / truncated runes pass through) in the `sanitizeControl` comment, the `techstack.md` row, and the `dsl.md` round-038 note; the note also scopes the E2E predicate's `utf8.ValidString` clause to the round's **text** fixtures.
- **N-5** — (a) corrected the fold-record wording: the unterminated-CSI fixture was **added** in the fold and corrected within the same change (no pre-existing fixture was modified); (b) **added real end-to-end coverage** for `genericEscLen`'s adjacency path — the E2E colouring command now emits an ESC **directly before a multibyte rune** (`printf '\033[31mFAILED\033失敗\n'`) alongside the SGR colour, so the E2E sanitize predicate exercises the adjacency path (no longer a unit-only narrowing).
- **Housekeeping** — resolved the eight first-review threads (fold replies attached).

## Round-038 fold review #3 (PR #79) — N-6

- **N-6** — the ADR 0007 residual-risk bullet now states the consumption is **window-bounded** and names the residual class TD-3 bounded (a sequence whose terminator lies beyond its kind's window prints its parameter text, bounded ≤ 1024 bytes); the D2 bounded-consumption bullet names that `genericEscLen` reuses the **CSI** window for its intermediate run; the ADR consequences bullet uses the precise "never introduces invalid UTF-8" claim (N-4 consistency); and the `genericEscLen` code comment names the shared CSI-window constant.
