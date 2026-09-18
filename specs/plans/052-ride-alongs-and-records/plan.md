# System Analysis Plan: 052-ride-alongs-and-records

**Round**: `052-ride-alongs-and-records` — close [#115](https://github.com/gosharplite/tellme/issues/115) (ride-alongs) + [#116](https://github.com/gosharplite/tellme/issues/116) (records)

**Inputs**: `spec.md` (US1/US2/US3 · FR-001…FR-010 · SC-001…SC-006 · Q1/Q2 locked) · `research.md` (D1–D7) · `specs/truth/techstack.md` (the round-052 MODIFY rows) · `truth-delta.md`.

---

## 1. Interface inventory

This round is a **dev-surface structural refactor + a records disposition**. It adds/changes **no** system interface: no CLI command, no flag, no argument, no stdin/stdout/stderr byte-contract, no exit code, no persisted record, no config key.

| Interface | Kind | Present? | Reason |
| --- | --- | --- | --- |
| CLI (terminal) end | `cli` | **contract unchanged** | The `[Tool Output]` block literals + the suggestion-list rendering are **byte-identical**; no new Rule/Example/step/`DSLRow`. |
| Backend API | `backend` | **absent** | tellme has a single CLI end and no OpenAPI/HTTP surface. |
| Frontend web | `frontend` | **absent** | A line-oriented CLI with a TUI prompt; the round touches no UX surface. |

**Conclusion**: **0 interfaces require a planner wave.** The round carries the CLI end forward to its contract owner `/axb-dsl-refine` — which records a **NOOP** (a `cli`-side refactor is not a CLI-contract change; the rounds 020/031/036/041–051 non-BDD-tooling precedent).

---

## 2. Waves

| Wave | Content | Delegation |
| --- | --- | --- |
| 1 | The round's whole surface (the two ride-alongs + the records relocation) | `/axb-api-plan` = **NOOP** · `/axb-data-plan` = **NOOP** · `/axb-dsl-refine` = **NOOP** · `/axb-ui-plan` = **not run** (no UX surface) |

`wave-covers-interfaces` holds vacuously (0 interfaces); `analysis-plan-never-writes-truth` — this plan writes no truth.

---

## 3. Change surface (RD map)

| # | Site | Change | Spec |
| --- | --- | --- | --- |
| 1 | `internal/infrastructure/tools/command.go` | `NewCommandTool(sink domaintools.OutputSink)`; `executeCommand{output domaintools.OutputSink}`; delete `toolOutputBox` + `sink()` + `BindToolOutput` | US1 · FR-001/FR-002 |
| 2 | `cmd/tellme/deps.go` | `agentTools()` (parameterless, read-free) + a new `assembleAgentTools(sink)`; `newToolRegistry(sink)`; drop the `BindToolOutput` wiring | US1 · FR-002/FR-003 |
| 3 | `internal/app/deps/deps.go` | widen `NewToolRegistry func(sink domaintools.OutputSink) domaintools.Registry`; **delete** `BindToolOutput` | US1 · FR-002 |
| 4 | `internal/cli/cli.go` | build `prog` before `reg`; `reg := dp.NewToolRegistry(prog.ToolOutput)`; delete the rebind call; `renderToolUsage` calls the factory with `nil` | US1 · FR-003 |
| 5 | `internal/ui/tui/prompt/suggester.go` | `set(items []string, cursor int)` | US2 · FR-005 |
| 6 | `internal/ui/tui/prompt/model.go` | the two callers pass `noChoice` | US2 · FR-005 |
| 7 | `docs/decisions/0021-ride-alongs-and-records.md` (+ index) | the round's ADR + the three relocated `#116` records | US3 · FR-006/FR-007 |
| 8 | `specs/truth/techstack.md` | MODIFY ×3 (command tool · `-i` · suggestion engine); the task-runner row NOOP | FR-007 |
| 9 | tests | new unit pins + the seam/fixture updates | FR-009 |

**Fixtures/tests touched by the seam-width change** (each passes `nil`): `internal/cli/cli_test.go`, `internal/cli/testdeps_test.go`, `cmd/tellme/deps_test.go`, `tests/e2e/steps/tool_usage.go`.

---

## 4. Gate & invariants (the DoD)

- **`make verify`** green: RULE-A/B/C **0**; **RULE-E baseline stays 0** (0 new / 0 stale — the terminal state; **no release valve**, ADR 0011/0016); RULE-F consistent; 0 cycles; lint 0; cross-compile 4/4.
- **Behaviour-preserving**: `go test -count=1 ./...` green (incl. the godog E2E); **no** behavioural assertion changed except where a test names a changed seam.
- **Assembler gate**: `agentTools()` stays parameterless + read-free; `agentTools()` ≡ the registry's tool set.
- **stdlib-only**: `go.mod`/`go.sum` unchanged; no new Gherkin/DSL row; the topology audit unchanged.
- **Witnesses (FR-009)** reproduced then reverted.

## 5. NOOP set (recorded in `truth-delta.md`)

`/axb-api-plan` → `specs/truth/contracts/**` NOOP · `/axb-data-plan` → `specs/truth/data/data-model.dbml` NOOP · `/axb-dsl-refine` → `specs/truth/features/cli/**` NOOP.
