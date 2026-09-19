# Tasks — round 057 (`057-tool-chrome-colour-and-payload-delta`)

**Inputs**: `spec.md` (US1/US2/US3) · `research.md` (D1–D9) · `plan.md` (1 CLI end; api/data NOOP) · `truth-delta.md`.
**Setup**: omitted (stdlib-only; no dependency change).

## Phase 1 — Foundational

- [X] **T001** — `internal/ui/colour.go`: add `colorGray`/`colorYellow` and factor `green`/`grey`/`yellow` over `wrap`.
- [X] **T002** — `internal/ui/tooloutput.go` + `coordinator.go` + `render_ports.go` + `internal/domain/render/ports.go` + `cmd/tellme/deps.go`: thread the `colour` flag through the progress factory into `ToolOutputWriter`.
- [X] **T003** — `internal/domain/render/ports.go` + `internal/ui/status.go` + `render_ports.go`: add `Lines.PayloadEstimate` + `FormatPayloadEstimate`(+colour).

## Phase 2 — Truth alignment

- [X] **T004** — `specs/truth/techstack.md` (MODIFY ×2 rows) — owner `/axb-technical-research`.
- [X] **T005** — `specs/truth/features/cli/chat/**` (MODIFY: the colour Rule; the payload increment Rule + budget re-scope; the cap Example) + `chat/dsl.md` — owner `/axb-dsl-refine`.
- [X] **T006** — **ADR 0027** + the `docs/decisions/README.md` index row.

## Phase 3 — Implementation (red-first where a carrier exists)

- [X] **T007** [BDD] — `internal/ui/colour_test.go`: pin the grey header/separators + the yellow action line (colour on/off byte-identity).
- [X] **T008** [BDD] — `internal/ui/status_test.go`/`colour_test.go`: pin `FormatPayloadEstimate` (`+100 ~203148 tokens - butler - model`; `+0`; `-1234`) and its green MODE accent.
- [X] **T009** [BDD] — `internal/ui/toolcall_test.go`: move the cap fixture to 501 (cap 500) + update the cap-set expectations.
- [X] **T010** [CODE] — `internal/ui/toolcall.go`: `argValueCap = 500` + `formatToolActionColour`; `toolrenderer.go`: `ActionLine` yellow.
- [X] **T011** [CODE] — `internal/ui/tooloutput.go`/`coordinator.go`/`render_ports.go`/`ports.go`/`deps.go`: the grey frame plumbing.
- [X] **T012** [CODE] — `internal/ui/status.go`/`render_ports.go`: `FormatPayloadEstimate`(+colour).
- [X] **T013** [CODE] — `internal/cli/call_renderer.go` + `cli.go`: the in-memory prev-estimate tracker; emit `PayloadEstimate`.
- [X] **T014** [BDD] — `tests/e2e/steps/*`: the payload helpers (delta) + the new colour/delta stepdefs; the no-colour step covers grey/yellow.
- [X] **T015** [BDD] — `internal/cli` (fake `Lines`): the `PayloadEstimate` double + the turn-order pins.

## Phase 4 — Verification & regression

- [X] **T016** [REGRESSION] — `gofmt` · `go vet ./...` · `make verify` · `go test -count=1 ./...` (all Examples).
- [X] **T017** [WITNESS] — falsifiability witnesses (a) un-gate the colour → the no-colour path reds; (b) drop the delta → the increment step reds; (c) cap 500 → 499 → the cap pin reds. Reproduce then revert.
- [X] **T018** — `truth-delta.md` owner rows + STATUS + the round branch/PR.

**Pre-Delivery Orphan Sweep**: 0 orphans (every new sentence has exactly one `DSLRow`; no feature removed).
