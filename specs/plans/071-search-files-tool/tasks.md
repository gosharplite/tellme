# Tasks — `search_files` (round 071)

**Plan Package**: `specs/plans/071-search-files-tool`

Legend: `[ ]` pending · `[X]` done. Phases per `/axb-tasks`: Setup → Foundational → (Phase 3) Test Alignment & Implementation → Feature (Green/Refactor).

## Phase 1 — Setup

- [X] **T001** Add `internal/infrastructure/tools/search.go`: the `searchFiles` adapter — `Name`/`Description`/`Contract`/`Parameters` (through the shared `resourceSchema`) — plus `NewSearchTool()`.
- [X] **T002** Add the constants (`searchMaxMatches=100`, `searchScanCap=1000`, `searchLineCap=500`, `searchMaxLineBytes=10MB`, `searchBinProbe=8000`) and the three recorded divergences in the file header.

## Phase 2 — Foundational

- [X] **T003** Offer the tool: append `infratools.NewSearchTool()...` in `assembleAgentTools` (after `NewFilesystemTools()`, before the write pair).
- [X] **T004** Record it: add `NewSearchTool()` to `registeredToolNames()` and `recordableToolNames()` (`tests/e2e/steps/tool_usage.go`).

## Phase 3 — Test Alignment & Implementation

- [X] **T005 (ALIGN/RED)** Add `internal/infrastructure/tools/search_test.go`: deterministic matches, no-match result, empty-query error, invalid-regex error, literal-vs-regex, binary skip, budget trim, 100-match cap, line trim.
- [X] **T006 (GREEN)** Implement `Execute` — args parse, `newSearchMatcher` (literal/regex), `collectMatches` (walk + scanFile), `sortMatches`, `formatMatches`, `truncateToBudget`.
- [X] **T007 (GREEN)** Wire the E2E: `specs/truth/features/cli/chat/searching-file-contents.feature` + `tests/e2e/steps/step_r071_search.go`.
- [X] **T008 (ALIGN)** Extend the schema/name unit tests (`TestToolSchemasRequireReason`, `TestToolNamesWireValid`) and the base offered-set assertion (`cmd/tellme/deps_test.go`).

## Phase 4 — Feature (Green / Refactor)

- [X] **T009 (GREEN)** `go test -count=1 ./...` green (incl. the new E2E journey: 275 scenarios).
- [X] **T010 (REFACTOR)** Fold the scan-cap early stop into the `collector` (no slice-length leakage into `walkDir`).
- [X] **T011 (RED→GREEN witness)** Reproduce then revert a falsifiability witness (remove the sort ⇒ the determinism pin reds; drop the cap ⇒ the capped marker pin reds) — recorded in the PR.

## Phase 5 — Truth / delivery

- [X] **T012** `specs/truth/techstack.md`: ADD a *file-content search tool* row; MODIFY the *Agent tool schemas* + *Tool-usage accounting* enumerations.
- [X] **T013** `specs/truth/features/cli/chat/dsl.md`: ADD the round-071 Given/Then rows.
- [X] **T014** **ADR 0043** + index row.
- [X] **T015** Domain model: `docs/domain-model/tellme.modelith.yaml` `Tool` entity (nine-tool surface) + `make modelith-render`.
- [X] **T016** `make verify` + `gofmt`/`go vet` clean; topology audit unchanged.
