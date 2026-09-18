# Tasks: 052-ride-alongs-and-records

> 來自 `research.md` Decisions D1–D7 與 `spec.md` (US1/US2/US3 · FR-001…FR-010 · Q1/Q2 locked)。
> 一輪恰好完成一批再標記 `[X]`（`task-strict-ordering`）。Red → Green → Refactor。
> **Setup**：略（stdlib-only；無新依賴）。**Foundational**：T001–T002。

## Phase 1 — Foundational

- [ ] **T001** [FOUND] `internal/infrastructure/tools/command.go`: change `NewCommandTool` to `NewCommandTool(sink domaintools.OutputSink) domaintools.Tool`; make `executeCommand` hold `output domaintools.OutputSink` directly; pass `c.output` in `Execute`. (No behaviour change; `sinkOpen`/`sinkClose`/`teeSink` already nil-guard.)
- [ ] **T002** [FOUND] Update the two external constructors: `cmd/tellme/deps.go` (`assembleAgentTools(sink)` → `NewCommandTool(sink)`) and `tests/e2e/steps/tool_usage.go` (`NewCommandTool(nil)`).

## Phase 3 — Test Alignment & Implementation

- [ ] **T003** [UNIT-RED] New `internal/infrastructure/tools/command_sink_test.go`: a fake `domaintools.OutputSink` (records `Begin`/`End`, exposes a `Writer`) passed to `NewCommandTool(sink)` receives the `[Tool Output]` block when a command runs (and **not** when the sink is nil). Reproduce the red by reverting the injection.
- [ ] **T004** [UNIT-RED] New `internal/ui/tui/prompt/suggester_set_test.go`: `set(items, noChoice)` leaves nothing selected; `set(items, 0)` selects the first item; `cycle` from `noChoice` lands on index 0 (the round-037 arithmetic unchanged).
- [ ] **T005** [GREEN] `internal/infrastructure/tools/command.go`: delete `toolOutputBox`, the `sink()` accessor, and `BindToolOutput`. (Zero-shared-edits: the deletion is in the same file as T001.)
- [ ] **T006** [GREEN] `cmd/tellme/deps.go`: extract `assembleAgentTools(sink domaintools.OutputSink) []domaintools.Tool`; `agentTools()` delegates with `nil` (stays parameterless + read-free); `newToolRegistry(sink)` uses `assembleAgentTools`; drop `BindToolOutput: infratools.BindToolOutput`.
- [ ] **T007** [GREEN] `internal/app/deps/deps.go`: widen `NewToolRegistry func(sink domaintools.OutputSink) domaintools.Registry`; delete the `BindToolOutput` field.
- [ ] **T008** [GREEN] `internal/cli/cli.go`: build `prog` before `reg`; `reg := dp.NewToolRegistry(prog.ToolOutput)`; delete `dp.BindToolOutput(reg, prog.ToolOutput)`; `renderToolUsage` calls its factory param with `nil` (`reg := newToolRegistry(nil)`), and its parameter type widens to `func(domaintools.OutputSink) domaintools.Registry`.
- [ ] **T009** [GREEN] `internal/ui/tui/prompt/suggester.go` + `model.go`: `set(items []string, cursor int)`; the two callers pass `noChoice`.
- [ ] **T010** [ALIGN] Update fixtures/tests for the widened seam: `internal/cli/cli_test.go`, `internal/cli/testdeps_test.go`, `cmd/tellme/deps_test.go` (`newToolRegistry(nil)`), and any `NewCommandTool` call.
- [ ] **T011** [DOCS] `docs/decisions/0021-ride-alongs-and-records.md` (already drafted in the plan half) + the `docs/decisions/README.md` index row; `specs/truth/techstack.md` MODIFY ×3.
- [ ] **T012** [DOCS] Relocate the three `#116` records into ADR 0021 §Records (verbatim + provenance); no edit to the frozen `036-*`/`040-*` packages.

## Phase 4 — Verification & Regression

- [ ] **T013** [VERIFY] `gofmt -l .` clean · `go vet ./...` clean · `make verify` **OK** (RULE-A/B/C 0; **RULE-E baseline 0**, 0 new / 0 stale; RULE-F consistent; 0 cycles; lint 0; cross-compile 4/4).
- [ ] **T014** [VERIFY] `go test -count=1 ./...` green (incl. the godog E2E); the topology audit unchanged; `go.mod`/`go.sum` unchanged.
- [ ] **T015** [WITNESS] Reproduce then revert: (a) the command-tool sink injection pin; (b) the `set(items, 0)` policy pin; (c) a governed-import probe ⇒ RULE-E red (the gate still has teeth at 0).
- [ ] **T016** [CLOSE] `STATUS.md` + the daily summary; then close [#115](https://github.com/gosharplite/tellme/issues/115) and [#116](https://github.com/gosharplite/tellme/issues/116) (comment naming ADR 0021 for #116).

## Orphan sweep

- Pre-delivery orphan sweep: **0** at plan time (the round deletes `toolOutputBox`/`BindToolOutput` and adds two files + two unit pins).
