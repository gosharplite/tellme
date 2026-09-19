# Tasks — Provider-wire projection of MCP-relayed tool schemas (round 061)

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Execution**: `/axb-implement` — Red → Green → Refactor, one task at a time, marked `[X]` after verification.

## Phase 1 — Setup

- [X] **T001** — Branch `061-mcp-schema-provider-projection` off `dev` `768dbec`; plan package created (spec + checklist + truth-delta skeleton); `STATUS.md` phase-gate update.
- [X] **T002** — `/axb-clarify` closed (CQ-1 → C · CQ-2 → ii · CQ-3 → i) + CQ-4…CQ-7 (probe approved · gate = unit pin · floor cross-family · silent) recorded in `spec.md` (S-1…S-7).

## Phase 2 — Foundational

- [X] **T003** — `/axb-spec-by-example`: acceptance journey written (`features/acceptance/keeping-a-gemini-turn-usable-with-an-annotated-mcp-server.feature`; 3 Rules; no residual `# [need clarification]`).
- [X] **T004** — `/axb-technical-research`: the **live probe** (declaration-only `generateContent`, `ait-comment` → `websc-dev-433809`, `gemini-3.8-flash`); the measured accept/reject table recorded in **ADR 0031**; `techstack.md` rows updated (the floor row, the qualified envelope row, the projection row, the gate row); ADR 0025 gains the amendment pointer; index rows added.
- [X] **T005** — `/axb-system-analysis`: `plan.md` (1 interface, 1 wave; api/data NOOP, ui skipped).

## Phase 3 — Test Alignment & Implementation

- [X] **T006** (`/axb-dsl-refine`) — truth rows: the round-056 declaration row gained the **carve-out** qualification; `chat/using-tools-from-a-remote-mcp-server.feature` gained the new Rule + 3 Examples; `chat/dsl.md` gained the round-061 Given (2) + Then (3) rows. Topology audit: no new findings.
- [X] **T007 (RED)** — Unit pins written first: `internal/infrastructure/mcp/schema_test.go::TestNormalizeMCPSchema_StripsVendorExtensions` and `internal/infrastructure/llm/gemini/schema_test.go` (3 pins incl. the envelope-shaped gate). Observed RED (no floor / no projection).
- [X] **T008 (GREEN)** — Floor implemented (`isVendorExtension`, recursive `stripVendorExtensions`, wired into `NormalizeMCPSchema`) and projection implemented (`supportedSchemaKeys`, `projectSchema` recursive + fail-closed, wired into `buildToolDeclarations`). Both packages green.
- [X] **T009 (RED → GREEN)** — E2E: `mcptest.AnnotatedSchema()`, the five new steps (`step_r061_mcp_schema.go`), `writeGeminiConfig` carrying the MCP block; the truth feature's new Rule drives both the **gemini** leg and the **tolerant-family** leg; the E2E contract ran green (the three new scenarios executed).
- [X] **T010 (REFACTOR)** — `projectSchema` failed closed on a non-object schema (caught by `TestProjectSchema_FailsClosed`), documented at the site.

## Phase 4 — Verification

- [X] **T011** — **Witness A**: removing the projection call turned `TestProjectToolDeclarations_NoUnsupportedKeywordReachesTheWire` **RED** (unsupported keywords on the wire) → reverted.
- [X] **T012** — **Witness B**: removing `stripVendorExtensions(obj)` turned `TestNormalizeMCPSchema_StripsVendorExtensions` **RED** (vendor keys kept at root/nested) → reverted.
- [X] **T013** — Gates: `gofmt` clean · `go vet ./...` clean · `go build ./...` clean · `go test -count=1 ./...` green (incl. the godog E2E) · `make verify` (see `STATUS.md`/the round record for the run).
- [X] **T014** — Docs/ledger: `research.md`, `plan.md`, this `tasks.md`, `truth-delta.md` owner rows.

## Notes / ledger conventions

- Witness executions are recorded here (TD-6 convention); the red output of both witnesses is reproducible from the two call sites (`buildToolDeclarations`'s projection call, `NormalizeMCPSchema`'s floor call).
- No product file outside the five sites in `plan.md` was touched; `go.mod`/`go.sum` unchanged.
