# Tasks: round 053 — offline session commands honour `-c` + a `-t` turn-log flag

**Plan package**: `specs/plans/053-offline-session-config-and-turns-flag`
**Branch**: `053-offline-session-config-and-turns-flag`
**Status**: `/axb-implement` executed — all tasks `[X]` (one-shot).

## Phase 1 — Foundational

- [X] **T001** — Domain port `history.TurnsLogStore` (`internal/domain/history/turns_log.go`): `Writer() (io.WriteCloser, error)` / `Read() (string, error)` / `Archive() error`.
- [X] **T002** — Infrastructure adapter `internal/infrastructure/history/turns_log_store.go` (`NewTurnsLogStore`; active `turns.log`, archive `turns.archive.log`; missing-file tolerance; best-effort).
- [X] **T003** — `deps.Dependencies.NewTurnsLogStore` seam + wiring in `cmd/tellme/deps.go`; fixture `internal/cli/testdeps_test.go` (`fakeTurnsLogStore`).

## Phase 2 — `-c`-honouring session selection (US1)

- [X] **T004** — `historyMode(homeDir, configPath) (string, error)`: `TELL_ME_MODE` → else the `-c` config's `MODE` → else default → else `"butler"`; an explicit unreadable `-c` errors (Q2 → A).
- [X] **T005** — `resolveWorkspace(homeDir, configPath) (resolution, *resolveError)`; `renderHistoryList`, `renderNewSession` take the config path; the `--new` dispatch sites thread `f.configPath`.
- [X] **T006** — Regression pin `internal/cli/history_mode_test.go` (`TestHistoryModeHonoursConfigPath`).

## Phase 3 — the `-t` flag + the `turns.log` writer (US2)

- [X] **T007** — `-t`/`--turns` flag in `parseFlags`; `renderTurnsLog` prints the resolved session's `turns.log`; `dispatchReporting` order `-d` → `-l` → `-t` → `--tool-usage`.
- [X] **T008** — `runtimeEnv.chrome` tee + `diag()`: on the prompt path the chrome sites (`call_renderer.go` frame/tail/metrics; `emitInputCaptured`; the `-i` echo) write through `diag()` → `io.MultiWriter(stderr, turns.log)`; `--new` archives the turn log.
- [X] **T009** — Infra unit pin `internal/infrastructure/history/turns_log_store_test.go`.

## Phase 4 — Test alignment, truth & regression

- [X] **T010** — `/axb-dsl-refine`: CLI truth — new Rules/Examples in `history/inspecting-the-session-history.feature` + new `history/reviewing-the-turn-log.feature`; new `history/dsl.md` rows.
- [X] **T011** — E2E stepdefs `tests/e2e/steps/step_r053_history.go` (config-session seed, `-t` seed, `-l -c`, `-t -c`, the three Thens).
- [X] **T012** — `specs/truth/data/data-model.dbml` `turns_log_line` artifact.
- [X] **T013** — Falsifiability witnesses reproduced then reverted: (a) revert `-c`'s mode read ⇒ the differential `-l` examples return identical output; (b) revert `-t`'s `-c` resolution ⇒ `-t` reads the default session; (c) drop the `turns.log` write ⇒ `-t` finds nothing.

## Phase 5 — Verification

- [X] **T014** — `gofmt` clean · `go vet ./...` clean · `make verify` **OK** (lint 0; arch gate 0 new/0 stale; cross-compile 4/4; vulncheck clean) · `go test -count=1 ./...` green (incl. the godog E2E, **Strict**).
- [X] **T015** — `STATUS.md` + `truth-delta.md` updated; ADR 0022 indexed.

## Pre-Delivery Orphan Sweep — 0 orphans

Every produced symbol has a consumer: `TurnsLogStore` (reader `-t`, archiver `--new`, writer the prompt path), `historyMode`/`resolveWorkspace` (the three offline session commands), the `-t` flag (the reporting batch), the `turns_log_line` artifact (the store). No new Makefile target; `go.mod`/`go.sum` unchanged.
