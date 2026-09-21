# Tasks — recoverable unknown tool name (round 076)

**Plan Package**: `specs/plans/076-recoverable-unknown-tool`
**Anchor**: [#154](https://github.com/gosharplite/tellme/issues/154)

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [X] **T001** Confirm the surfaces: `internal/agent/agentloop.go` (`Run` dispatch, `refuseReasonless` precedent, `unknownToolResult` target); `internal/cli/cli.go` `emitToolError` + `exitcode.go` `ToolError=7`; the `chat` truth feature + DSL; the E2E helpers (`wire_tools.go`, `scenarioFrom`/`onlyFake`/`readArgs`); the fake provider's scripted replies. No new technology → no Setup package.

## Phase 2 — Foundational

- [X] **T002** `internal/agent/agentloop.go`: add the `strings` import, the `maxUnknownToolFolds` constant, and the loop-local `unknownFolds` counter. *Only* the loop; no registry/port/wire change.

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (RED)** `internal/agent/agentloop_test.go`: replace `TestRunUnknownToolIsTerminal` with `TestRunUnknownToolIsFedBackAndContinues` (fold back, name the tool, continue to the answer, no step) and `TestRunUnknownToolIsBoundedPerTurn` (the cap terminates at `maxUnknownToolFolds+1` provider calls).
- [X] **T004 (RED)** `tests/e2e/steps/step_r076_unknown_tool.go`: the unknown-then-read-then-answer Given and the `the run reported the unavailable tool "{tool}"` Then.
- [X] **T005 (RED)** Truth: new `specs/truth/features/cli/chat/using-an-unknown-tool-name.feature` (2 Rules); relocate the unavailable-tool Example out of `failing-the-tool-loop.feature`; add the `chat/dsl.md` Given/Then rows.
- [X] **T006 (GREEN)** `Run`: on a registry miss, fold back `unknownToolResult(tc.Name, a.Registry.Tools())` as a `tool`-role result + `continue`; return the incomplete error only once `unknownFolds >= maxUnknownToolFolds`. Add the `unknownToolResult` helper.

## Phase 4 — Feature (Green / Refactor)

- [X] **T007 (witness — reproduced, then reverted)** Restore the terminal `return` on the unknown branch: `TestRunUnknownToolIsFedBackAndContinues` reddens (the run aborts) **and** the E2E `A one-off unknown tool name is recovered and the turn finishes` reddens (`no recoverable result naming the unknown tool "time_travel" was fed back`).
- [X] **T008** Truth: `specs/truth/techstack.md` *Agent tool loop* row MODIFY (+ the *Tool-usage accounting* parenthetical); **ADR 0048** + index.
- [X] **T009** Domain model: **not modelled** — the round changes an error-handling classification, not a modelled entity/invariant; recorded in `plan.md` §5 (ADR 0041's escape hatch).

## Phase 5 — Delivery

- [X] **T010** `gofmt -l` clean; `go build ./...`; `go vet ./...`; `go test -count=1 ./...` green (E2E incl. the pair); `make verify`; `make test-race`; `go.mod`/`go.sum` unchanged; the topology audit adds no new error.
