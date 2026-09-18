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

---

## Fold ledger — PR [#119](https://github.com/gosharplite/tellme/pull/119) review (`5737154767`) + fold-verification (`5737246516`)

Append-only record of the review folds (kept so the pre-fold execution log above survives). Fold head **`74e0eb2`**.

- **B-54-1 `[ARCHITECTURAL BLOCKER]` — the colour leaked into `turns.log`.** The renderer produced the **coloured** string and the round-053 tee (`io.MultiWriter(stderr, file)`) sent it to **both** sinks, so the artifact carried SGR bytes and `-t | …` ceased to be plain (violating `FR-006`/Q3/ADR-0023-D4, with an inverted NOOP in the ledger). **Fold:** the per-call renderer now holds **two** `render.Lines` — coloured for `stderr`, a **plain** one (`dp.NewLines(false)`) for the file leg — and a single `emit(build)` helper writes each chrome line to both sinks, building the file leg from the plain renderer; the tee is the **raw file writer** (no `MultiWriter`), and `diag()` is deleted. `turns.log` is control-free **by construction** (render-don't-strip, the stronger contract). Carried by the new acceptance Rule + interface Rule + `chat/dsl.md` row + E2E `thenTurnLogNoDecoration`; witnessed by re-introducing the coloured file leg (red).
- **F-54-1 `--last`** — the dead `arg == \"--last\"` branch is deleted; FR-001 + the `cli.go` comments corrected to `-l`/`--list`.
- **F-54-2** — the round-039 "control-free `[Tool …]`" policy is **qualified** (it governs the **model-authored value**; tellme's own `[Tool Reason]` accent is a distinct, terminal-gated class; the control-free rows hold for their non-terminal fixtures) in `techstack.md` + `chat/dsl.md`; the wrong **NOOP** for the Turn-log row is now a **MODIFY**.
- **F-54-3** — `chromeColour` now delegates to `spinnerGate` (one predicate, one home).
- **F-54-4** — the dropped `FR-006` requirement re-carried (acceptance Rule + interface Rule + DSL row + carrier); the checklist line corrected.
- **RF-54-1 (closed)** — the plain `FormatPayloadStatus`/`FormatReady`/`FormatToolReason` are the **single entry points** the colour siblings return on the colour-off path (production call edges; no orphans).
- **RF-54-2 (closed)** — the terminal gate is resolved **once** in `runTurn` (`colourOn`) and threaded to the three consumers.
- **RF-54-3 / nit 1 (recorded)** — the `consumeListValue` boundary cases + the `-l`-beats-a-positional precedence in **ADR 0023 RF-54-4**.
- **R-54-1 / R-54-2 (residual, non-blocking)** — the new carrier and the negative Example are tool-less; the `emit(build)` seam is uniform and mutation-proved, so a tool-using fixture is a strengthening, recorded for when the file is next touched.
- **R-54-3 (this section)** — the fold ledger, appended per the round-053 precedent.

**Fold-head verification:** `gofmt`/`go vet` clean · `make verify` **OK** · `go test -count=1 ./...` green (incl. the godog E2E, **Strict**, 240 scenarios) · `go.mod`/`go.sum` unchanged. Fold-verification `5737246516` — **ALL FOLDS VERIFIED, CLEARED FOR MERGE**.
