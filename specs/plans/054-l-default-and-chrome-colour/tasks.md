# Tasks: round 054 — `-l` defaults to `1` + green chrome accents

**Plan package**: `specs/plans/054-l-default-and-chrome-colour`
**Branch**: `054-l-default-and-chrome-colour`
**Status**: `/axb-implement` executed — all tasks `[X]` (one-shot).

## Phase 1 — Foundational

- [X] **T001** — `internal/ui/colour.go`: the reference's `colorGreen`/`colorReset` + the `green(s, enabled)` helper (the round-054 policy + the recorded divergence).
- [X] **T002** — colour-aware formatter siblings: `formatPayloadStatusColour`, `formatReadyColour`, `formatToolReasonColour` (plain `false` path byte-identical to the existing formatters).

## Phase 2 — the `-l` default (US1)

- [X] **T003** — `parseFlags`: `fs.Lookup("list").NoOptDefVal = "1"` + `consumeListValue(args)` (consume an adjacent integer; leave a non-integer token; stop at `--`).
- [X] **T004** — Regression pin `internal/cli/list_default_test.go` (`TestParseFlagsListDefault`).

## Phase 3 — the chrome colour (US2)

- [X] **T005** — `ui.Lines` carries `colour bool` (`NewLines(colour)`) and uses the colour-aware formatters; `ui.ToolLineRenderer` carries `colour bool` (`ToolLines(colour)`) and colours the whole `[Tool Reason]` line.
- [X] **T006** — widen the deps seams `NewLines func(colour bool)` / `NewToolLines func(colour bool)`; wire `cmd/tellme/deps.go`; compute `chromeColour(opts, env)` in `runTurn` and thread it into `newCallRenderer` + the loop spec.
- [X] **T007** — Unit pin `internal/ui/colour_test.go` (`TestChromeColour`).

## Phase 4 — Test alignment, truth & regression

- [X] **T008** — `/axb-dsl-refine`: CLI truth — the bare-`-l` Rule in `history/inspecting-the-session-history.feature` + the new `chat/colouring-the-session-chrome.feature`; the `history/dsl.md` + `chat/dsl.md` rows.
- [X] **T009** — E2E stepdefs `tests/e2e/steps/step_r054_colour.go` (the bare-`-l` When; the green-accent Then; the no-colour Then; the usage-reporting reason provider Given).
- [X] **T010** — ADR 0023 + the `docs/decisions/README.md` index row; `specs/truth/techstack.md` rows (CLI flag parsing, Session lifecycle flags, Turn chrome, Post-turn status).
- [X] **T011** — Falsifiability witnesses reproduced then reverted: (a) drop the `-l` default ⇒ bare `-l` is a usage error; (b) invert the colour gate ⇒ the negative Example reds; (c) drop a colour site ⇒ the unit pin reds.

## Phase 5 — Verification

- [X] **T012** — `gofmt` clean · `go vet ./...` clean · `make verify` **OK** · `go test -count=1 ./...` green (incl. the godog E2E, **Strict**).
- [X] **T013** — `STATUS.md` + `truth-delta.md` updated; `go.mod`/`go.sum` unchanged.

## Pre-Delivery Orphan Sweep — 0 orphans

`green`/`colorGreen` (the adapters), `formatPayloadStatusColour`/`formatReadyColour`/`formatToolReasonColour` (the adapters), `consumeListValue` (parseFlags), `chromeColour` (runTurn), `NewLines`/`ToolLines` colour flag (cmd/tellme + cli). No new Makefile target.
