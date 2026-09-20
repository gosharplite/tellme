# Tasks — the `-h`/`-v` CLI shorthands (round 074)

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [X] **T001** Confirm the surfaces: `internal/cli/cli.go` `parseFlags` / `run` / `emitUsageError`; the `usage` and `diagnostics` truth features + DSL; the reference's cobra `-h`/`-v` (verified on the installed binary).

## Phase 2 — Foundational

- [X] **T002** `internal/cli/cli.go`: add `flags.help` + `flags.helpText`; register `-h/--help` and the `-v` shorthand on `--version`.

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (RED)** `internal/cli/help_version_test.go`: `TestParseFlagsHelpShorthand` · `TestParseFlagsVersionShorthand` · `TestParseFlagsUnknownFlagStillRefused` · `TestRunHelpPrecedenceAndStream` · `TestRunVersionShorthand`.
- [X] **T004 (RED)** `tests/e2e/steps/step_r074_help.go`: the `tellme prints its flag list` and `the help is reported as a success` Thens.
- [X] **T005 (RED)** Truth feature `specs/truth/features/cli/usage/requesting-help.feature` (+2 Rules) and the `-v` Example in `diagnostics/version-and-setup-diagnostic.feature`.
- [X] **T006 (GREEN)** `parseFlags` captures `"Usage of tellme:\n" + fs.FlagUsages()` when help is set; `run` prints it to **`stdout`** and returns `Success`, checked **before** the `--version` path.

## Phase 4 — Feature (Green / Refactor)

- [X] **T007 (witness — reproduced, then reverted)** (a) drop the explicit `-h` flag (pflag's implicit path): the E2E `The operator asks for help with "-h"` reddens at `tellme prints its flag list` (stdout empty; exit 2) **and** the unit pins; (b) drop the `-v` shorthand: the E2E `… with the short flag` reddens at `tellme prints the build version` **and** the unit pins.
- [X] **T008** Truth: `specs/truth/techstack.md` *CLI flag parsing* row MODIFY; the root `cli/dsl.md` gains the cross-module `runs tellme with "{flag}"` row (moved from `diagnostics/dsl.md`, which keeps a pointer note); `usage/dsl.md` gains the two Thens; **ADR 0046** + index.
- [X] **T009** Domain model: **not modelled** — the round changes a CLI flag surface, not a modelled entity/invariant; recorded in `plan.md` §5 (ADR 0041's escape hatch).

## Phase 5 — Delivery

- [X] **T010** `make verify` + `go test -count=1 ./...` green; `go.mod`/`go.sum` unchanged; the topology audit adds no new error.

## Review fold ledger (PR [#152](https://github.com/gosharplite/tellme/pull/152) — the `architect` peer)

- **F-1 (required)** — the truth row owning the **dispatch precedence** (`techstack.md` *Prompt input*) was left stale; the round now MODIFYs it to `--help` → `--version` → `-d` → `-l` → `-t` → `--tool-usage`, adds `-h`/`--help` to the never-read-stdin list, and records that a **parse error pre-empts help**.
- **F-2 (required)** — the durable record mis-described the error path: ADR 0046 §2 + the `cli.go` comment now say the block is what pflag's **implicit help path** wrote to stderr *before this round* (the unrecognized-flag path prints only the phrase, no block); the §Consequences "minus `Usage:` sections" wording is corrected (the block's first line *is* `Usage of tellme:`).
- **TD-3** — the flag-list Then now asserts the **full** accepted flag set (claim == carrier); the success Then's row is made self-contained.
- **TD-4** — `-h`/`--help` added to the `tellme performs no network access` scope enumeration + the two help Examples now carry the no-network Then.
- **TD-5** — the new Examples gain `And the diagnostics are shown at a terminal`, so the `no progress spinner` negative has teeth.
- **RF-1** — recorded as **RF-074-5** (`helpText` cached on the parse result).
- **RF-2** — folded: ADR 0046 §3 + `spec.md` record the parse-error pre-emption, `techstack.md` *Prompt input* states it.
- **N-1** — the *Version injection* row now reads `-v`/`--version`.
- **N-2** — `spec.md` marks the `-h`+prompt/`-c` and `-h -v` edge cases as unit-carried.
- **N-3** — addressed by TD-4 (the help Examples now carry the offline Then).
