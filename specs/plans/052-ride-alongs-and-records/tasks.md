# Tasks: 052-ride-alongs-and-records

> 來自 `research.md` Decisions D1–D7 與 `spec.md` (US1/US2/US3 · FR-001…FR-010 · Q1/Q2 locked)。
> 一輪恰好完成一批再標記 `[X]`（`task-strict-ordering`）。Red → Green → Refactor。
> **Setup**：略（stdlib-only；無新依賴）。**Foundational**：T001–T002。

## Phase 1 — Foundational

- [X] **T001** [FOUND] `internal/infrastructure/tools/command.go`: `NewCommandTool(sink domaintools.OutputSink) domaintools.Tool`; `executeCommand{output domaintools.OutputSink}`; delete `toolOutputBox` + `sink()` + `BindToolOutput`; `Execute` passes `c.output`.
- [X] **T002** [FOUND] External constructors: `cmd/tellme/deps.go` (`assembleAgentTools(sink)` → `NewCommandTool(sink)`) and `tests/e2e/steps/tool_usage.go` (`NewCommandTool(nil)`).

## Phase 3 — Test Alignment & Implementation

- [X] **T003** [UNIT-RED] New `internal/infrastructure/tools/command_sink_test.go` — a fake `domaintools.OutputSink` passed to `NewCommandTool(sink)` receives the block (Begin/End + the child bytes); a nil sink is a no-op.
- [X] **T004** [UNIT-RED] New `internal/ui/tui/prompt/suggester_set_test.go` — `set(items, noChoice)` selects nothing; `set(items, 0)` selects the first row; a refresh replaces items+cursor; `cycle` from `noChoice` lands on index 0.
- [X] **T005** [GREEN] `internal/infrastructure/tools/command.go`: the box/accessor/`BindToolOutput` deleted (same file as T001).
- [X] **T006** [GREEN] `cmd/tellme/deps.go`: `assembleAgentTools(sink)`; `agentTools()` delegates with `nil` (parameterless + read-free); `newToolRegistry(sink)`; the `BindToolOutput` wiring dropped.
- [X] **T007** [GREEN] `internal/app/deps/deps.go`: `NewToolRegistry func(sink domaintools.OutputSink) domaintools.Registry`; the `BindToolOutput` field deleted.
- [X] **T008** [GREEN] `internal/cli/cli.go`: `prog` built before `reg`; `reg := dp.NewToolRegistry(prog.ToolOutput)`; the `dp.BindToolOutput` call deleted; `renderToolUsage` widened + calls its factory with `nil`.
- [X] **T009** [GREEN] `internal/ui/tui/prompt/suggester.go` + `model.go`: `set(items, cursor)`; both callers pass `noChoice`.
- [X] **T010** [ALIGN] Fixtures: `internal/cli/{cli,testdeps}_test.go`, `cmd/tellme/deps_test.go` (`newToolRegistry(nil)`), `tests/e2e/steps/tool_usage.go`.
- [X] **T011** [DOCS] `docs/decisions/0021-ride-alongs-and-records.md` + the `docs/decisions/README.md` index row; `specs/truth/techstack.md` MODIFY ×3.
- [X] **T012** [DOCS] The three `#116` records relocated into ADR 0021 §Records (copied verbatim + provenance); the frozen `036-*`/`040-*` packages untouched.

## Phase 4 — Verification & Regression

- [X] **T013** [VERIFY] `gofmt -l .` clean · `go vet ./...` clean · `make verify` **OK** (RULE-A/B/C 0; **RULE-E baseline 0**, 0 new / 0 stale; RULE-F consistent; 0 cycles; lint 0; govulncheck clean; cross-compile 4/4).
- [X] **T014** [VERIFY] `go test -count=1 ./...` **green** (24 pkgs incl. the godog E2E); `go.mod`/`go.sum` unchanged; no new Gherkin/DSL row.
- [X] **T015** [WITNESS] Reproduced then reverted: **(a)** dropping the sink injection ⇒ `TestNewCommandToolCarriesInjectedOutputSink` FAILs (`Begin/End = 0/0`); **(b)** resetting the cursor inside `set` ⇒ `TestSuggesterSetOwnsTheCursor/an_explicit_cursor…` FAILs; **(c)** a governed `cli → ui` import at baseline 0 ⇒ the gate FAILs (RULE-F coverage + *"an emptied baseline MUST fail"*). All reverted clean.
- [X] **T016** [CLOSE] `STATUS.md` + the daily summary; close [#115](https://github.com/gosharplite/tellme/issues/115) and [#116](https://github.com/gosharplite/tellme/issues/116) (comment naming ADR 0021 for #116).

## Outcome (implementation)

- **Product** — `internal/infrastructure/tools/command.go` (ctor-injected sink; box/rebind deleted) · `internal/app/deps/deps.go` (widened `NewToolRegistry`; `BindToolOutput` deleted) · `cmd/tellme/deps.go` (`assembleAgentTools` + sink-aware `newToolRegistry`; wiring) · `internal/cli/cli.go` (reorder + inject; rebind deleted) · `internal/ui/tui/prompt/{suggester,model}.go` (caller-owned cursor).
- **Tests** — new `command_sink_test.go` + `suggester_set_test.go`; fixtures aligned.
- **Docs/truth** — ADR 0021 (+ index) · `techstack.md` ×3 · this package.
- **Gates** — `make verify` OK · `go test -count=1 ./...` green (24 pkgs) · witnesses (a)/(b)/(c) reproduced + reverted · `go.mod`/`go.sum` unchanged.

## Orphan sweep

- Post-delivery orphan sweep: **0** (no orphan symbols introduced; `BindToolOutput`/`toolOutputBox` removed with their sole callers).

## Fold — PR [#117](https://github.com/gosharplite/tellme/pull/117) review `5730498811` (APPROVE WITH REQUIRED FOLDS)

| Finding | Fold |
| --- | --- |
| **F-52-1** `[TECHNICAL DEBT]` — the **Agent tool loop** truth row still asserts *"`BindToolOutput` MUST run before `Run`"*, a clause R-2 makes false; the `truth-delta.md` never inspected that row | `specs/truth/techstack.md` Agent tool loop row: the clause is **retired, not deleted** — *"the `BindToolOutput`-before-`Run` half is retired by construction (ctor-injected sink; ADR 0021) … the surviving pre-`Run` requirement binds `BindSkillsCatalog`"*; a **MODIFY** row added to `truth-delta.md` (the row is now inspected + recorded) |
| **F-52-2** `[TECHNICAL DEBT]` — `STATUS.md` carried two live-state contradictions (3rd occurrence of the class) | **(a)** the delivered-rounds index line no longer asserts liveness (*"the delivered-rounds index above records delivery, not liveness"*); **(b)** the roadmap `future slices` row dropped [#115](https://github.com/gosharplite/tellme/issues/115)/[#116](https://github.com/gosharplite/tellme/issues/116) (they are the in-flight round); **durable remedy** — a new **SESSION-CLOSEOUT** Step-3 rule + closeout rule #14 (*no liveness contradictions*) |
| **RF-52-1** `[REFACTOR]` — `BindSkillsCatalog` is now the only pointer-based seam and fails silently | ADR 0021 §Forward **RF-52-1** (migrate it in the same change that converts `listSkills` to a value type) |
| **RF-52-2** `[REFACTOR]` — the records' ADR home is write-once | ADR 0021 §Records **lifecycle clause** (a changed record is superseded by a new ADR citing this one) |
| **RF-52-3** `[REFACTOR]` — the caller-trust cursor is bounded but unpinned | `suggester_set_test.go` gains the **out-of-range** case (`set(nil, 3)` ⇒ no selection, no panic) |
| **Nit** — state which layer owns what | ADR 0021 §Consequences: the **two-layer witness** split (ctor pin vs the E2E wiring) + **no-drift by construction** (one builder) + **interface-copy survival through `augmentRegistryWithMCP`** |
| **Nit** — witness (c) re-runs round 051's witness | labelled in this package as a **ratchet regression check**, not round-052 evidence |
| **Witness (d)** — the reviewer's **unclaimed** witness: mutate the prompt path to `dp.NewToolRegistry(nil)` | **Reproduced** — 4 godog E2E `[Tool Output]` scenarios FAIL (`no [Tool Output] block to inspect` / `no [Tool Output] content line`); reverted clean. Recorded in ADR 0021 §Consequences (the wiring is owned by the E2E, one layer above the ctor pin). |
