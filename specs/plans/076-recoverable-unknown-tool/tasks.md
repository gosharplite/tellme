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

## Review fold ledger — architect review 1 (`APPROVE WITH REQUIRED FOLDS`)

| # | Fold | Where |
| --- | --- | --- |
| **F-1** | pin the fold-back's **pairing** (`ToolCallID`) + the **no-usage-record** half + a **mixed-round** witness | `internal/agent/agentloop_test.go` (`TestRunUnknownToolIsFedBackAndContinues`, `TestRunUnknownToolRecordsNoUsage`, `TestRunMixedRoundUnknownAndValid`) |
| **F-2** | make the cap witness **discriminating** (assert the fold-back happened) + pin the value `3` | `using-an-unknown-tool-name.feature` + `offering-the-agent-tools.feature` (the `the run reported the unavailable tool` Then) · `TestMaxUnknownToolFoldsValuePinned` |
| **F-3** | pin the **reason × unknown ordering** (unknown wins; single-owned) | `TestRunUnknownToolWithBlankReasonClassifiesAsUnknown` |
| **F-4** | correct the false "bounded by the byte clamp" claim | `research.md` D2 · ADR 0048 Decision 2 + RF-076-2 |
| **F-5** | the fold-back emits the **action block** before the result (chrome parity) | `internal/agent/agentloop.go` (`logAction`) + `TestRunUnknownFoldBackEmitsActionLine` |
| **F-6** | sync the stale `offering-the-agent-tools.feature` unknown-tool Example | `offering-the-agent-tools.feature` |
| **TD-076-1** | assert the **available-tools list** | `TestRunUnknownToolIsFedBackAndContinues` |
| **TD-076-2** | the per-turn **reset** witness | `TestRunUnknownFoldBackCounterResetsPerRun` |
| **TD-076-4** | the cap pin fails on the cap, not fake exhaustion | `TestRunUnknownToolIsBoundedPerTurn` (20 scripted replies) |
| **N-1** | ADR Related conflation | ADR 0048 |
| **N-2** | scope the tool-usage sentence | `specs/truth/techstack.md` |
| **N-4** | annotate the relocated-Example pointer | `failing-the-tool-loop.feature` |
| TD-076-3 / N-3 / N-5 | recorded / accepted as-is / closed by F-1 | — |

**Witnesses re-run (reproduced then reverted):** drop `ToolCallID` ⇒ `TestRunUnknownToolIsFedBackAndContinues` red (`ToolCallID = ""`) · `maxUnknownToolFolds = 10` ⇒ `TestMaxUnknownToolFoldsValuePinned` red · `maxUnknownToolFolds = 0` (the old terminal behaviour) ⇒ the three E2E unknown-tool Examples red at `no recoverable result naming the unknown tool "time_travel" was fed back` (the cap witness is now discriminating).


## Residual fold ledger — fold-verification 1 (`FOLDS VERIFIED WITH RESIDUALS — CLEARED FOR HUMAN MERGE`)

| # | Residual | Fold |
| --- | --- | --- |
| **R1** | TD-076-4 — the cap unit pin still coupled to fake exhaustion for a large N | `repeatingGateway` (repeats its last reply, like the E2E fake) used by `TestRunUnknownToolIsBoundedPerTurn`; robust for any cap value |
| **R2** | the fold-back pairing was unit-tier only | `thenUnknownToolReported` now asserts the E2E fold-back carries a non-empty `tool_call_id` (the round-065 / #132 pairing at the E2E tier) |
| **R3** | `offering-the-agent-tools.feature` Example prose named the "summarisation tool" while the fixture scripts `time_travel` | Example title reworded to "a tool tellme does not offer" |

**Witness (reproduced then reverted):** `maxUnknownToolFolds = 25` ⇒ the cap pin **passes** (the repeating gateway removes the fake-exhaustion coupling).
