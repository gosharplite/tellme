# Tasks — the `-l` listing presentation (round 073)

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [ ] **T001** Confirm the surfaces: `internal/cli/cli.go` `renderHistoryList` (the flat `role: content` line), `toMessages`, `dispatchReporting`, `runtimeEnv`; `internal/ui/colour.go` (`wrap`, the palette); `internal/domain/render/ports.go` (the `Lines`/`Answer` family); `cmd/tellme/deps.go` (`NewAnswer`/`NewLines`).
- [ ] **T002** Confirm glamour reuse: `internal/ui/renderer.go` `NewRenderer().Render(markdown, width)` (style/sanitizer/word-wrap/degraded fallback) is the one rendering policy.

## Phase 2 — Foundational

- [ ] **T003** `internal/ui/colour.go`: add `colorBlue` (`\033[1;34m`) / `colorMagenta` (`\033[1;35m`) + `blue(s, on)` / `magenta(s, on)` wrappers (via the existing `wrap`).

## Phase 3 — Test Alignment & Implementation

- [ ] **T004 (RED)** `internal/ui/listing_test.go`: `TestListingReportsRoleHeadersAndSeparators` (headers always; one blank line after each message; last message ends with a blank line) · `TestListingRendersOnlyTheModelBody` (the model body is rendered — markers absent/words present; the operator body is verbatim) · `TestListingRawPrintsVerbatim` (raw ⇒ both bodies verbatim, no ANSI) · `TestListingColourGate` (colour on ⇒ blue/magenta SGR around the headers; off ⇒ zero `\033`).
- [ ] **T005 (RED)** `internal/cli/list_render_test.go` (or extend `dispatch_test.go`): `renderHistoryList` maps `user → operator`/`assistant → model` and calls the injected `render.Listing`; the stdout-terminal gate resolves the colour; a `-l N` slice is unchanged; the width resolution is best-effort (an unreadable config ⇒ width 0, listing still succeeds). Update `dispatch_test.go`'s expected `-l` bytes to the new shape.
- [ ] **T006 (RED)** `internal/cli/cli_test.go`: `TestStdoutTerminalDetector` (`TELL_ME_FORCE_STDOUT_TTY` forces the probe; unset ⇒ the real probe).
- [ ] **T007 (GREEN)** Implement the port: `internal/domain/render/ports.go` — `ListingRole{ListingOperator,ListingModel}`, `ListingMessage{Role,Body}`, `ListingSpec{Colour,Raw,Width,Warn}`, `Listing` interface.
- [ ] **T008 (GREEN)** Implement the adapter: `internal/ui/listing.go` — `NewListing()` returning the adapter (header → body → separator; model body via `NewRenderer` unless raw; operator body verbatim; the degrade warning best-effort).
- [ ] **T009 (GREEN)** Wire the deps: `internal/app/deps/deps.go` `NewListing func() render.Listing` (+ `Validate`) and `cmd/tellme/deps.go` `NewListing: func() render.Listing { return ui.NewListing() }`; update `internal/cli/testdeps_test.go` with a fake.
- [ ] **T010 (GREEN)** Wire the CLI: `runtimeEnv` gains `stdoutTTY` + `stdoutIsTerminal()`; `stdoutTerminalDetector()` (the `TELL_ME_FORCE_STDOUT_TTY` seam); `renderHistoryList` gains `raw`/`width`/the listing factory and emits the new bytes; `dispatchReporting` threads them; `-r` reaches the listing; the `-r` flag help text names the listing.
- [ ] **T011 (RED)** E2E steps: `tests/e2e/steps/step_r073_listing.go` — the Given `the output is shown at a terminal` (`TELL_ME_FORCE_STDOUT_TTY=1`), the When `… as raw output`, and the Thens (heads / rendered prose / verbatim prompt / raw source / blank separator / accents / no accents).
- [ ] **T012 (RED)** Update the existing listing Thens to the new shape: `step_t015_history_then_lists_last.go`, `step_t023_history_then_operator_only.go`, `step_r053_history.go` (`thenListsAssistantMessage`).
- [ ] **T013 (GREEN)** Run: the round-073 Examples in `specs/truth/features/cli/history/inspecting-the-session-history.feature` green; the existing `-l` Examples green.

## Phase 4 — Feature (Green / Refactor)

- [ ] **T014 (witness)** Reproduce then revert three mutants: (a) drop the separator ⇒ the blank-line Example RED; (b) render the operator body ⇒ the verbatim-prompt Example RED; (c) un-gate the colour ⇒ the no-accents Example RED.
- [ ] **T015 (witness)** Ignore `-r` in the listing ⇒ the raw-source Example RED (reproduced then reverted).
- [ ] **T016** Truth: `specs/truth/features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` (round-073 rows) — **done at dsl-refine**; `specs/truth/techstack.md` ×4 + the domain model — **done at research**.
- [ ] **T017 (refactor)** Keep `renderHistoryList` under the `cyclop` gate (extract the message mapping / width resolution into named helpers if needed); no second rendering policy.

## Phase 5 — Delivery

- [ ] **T018** `make verify` + `go test -count=1 ./...` green; `go.mod`/`go.sum` unchanged; the topology audit adds no new error; `turns.log` stays control-free.
