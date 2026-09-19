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

## Fold ledger — PR [#123](https://github.com/gosharplite/tellme/pull/123) review (`5254577583`, APPROVE WITH REQUIRED FOLDS — no blocker)

| # | Finding | Fold |
| --- | --- | --- |
| **F-057-1** (required) | The acceptance Rule *"A payload that shrank shows a negative increase"* (A8) had **no** interface carrier (no Rule/`DSLRow`/stepdef). | **Recorded as a deliberate permanent narrowing (the reviewer's sanctioned alternative)** — the shrink is un-constructible end-to-end (the baseline is per-process and an in-turn request only grows): ADR 0027 §Forward names the carriers, and a **new CLI-tier pin** `TestOnCallBeginEmitsNegativeIncrement` drives the signed arithmetic to a negative (a clamp mutant reds it). The round-036 precedent. |
| **F-057-2** (required) | The `turns.log` content-parity claim had no real carrier (`Payload:` matched the old shape too). | New truth `Then` *"the saved turn log carries the pre-flight payload with its increment and no allowance"* + stepdef (`reTurnLogEstimate` present, `Payload: ~<n>/<n>` absent, no ESC) + a `dsl.md` row; and the round-053 assertion `thenSessionTurnsLogHoldsProgress` now pins the round-057 shape instead of the bare `Payload:` substring (+ its `history/dsl.md` row). |
| **TD-057-1** | `Lines.PayloadStatus` kept a now-unreachable `estimated bool` branch; `status.go`'s doc described the retired shape. | Folded the reviewer's scalable form: `PayloadStatus(t, …, estimated)` → **`PayloadMeasured(t, tokens, budget, mode, model)`**; the `~`-branch is gone; `FormatPayloadStatus` → **`FormatPayloadMeasured`** (+ `formatPayloadMeasuredColour`); the stale doc comment rewritten; the two `estimated == true` unit pins removed (`status_test.go` rewritten; `turn_test.go`'s pin moved to `FormatPayloadEstimate`). |
| **TD-057-2** | `NewProgress(…, colourOn, colourOn)` — two adjacent positional bools with different policies. | Folded the reviewer's named-field form (the repo's own `LoopSpec` pattern): `render.ProgressSpec{Stream, Now, Model, Epoch, Columns, IdleGap, Spinner, Colour}` + `ProgressFactory func(ProgressSpec)`; the `cli.go` call site reads by name. |
| **R-057-1** | `ToolOutputSeparatorColour` exported for in-package use. | Unexported to `toolOutputSeparatorColour` (matching its unexported sibling `formatToolOutputHeaderColour`). |
| **R-057-2** | Two unrecorded consequences. | Recorded in **ADR 0027 §Forward**: (a) the effective budget is now invisible on a usage-less provider (round-024's display decision narrowed to *"iff the provider reports usage"*); (b) the baseline is *per turn == per process* today — a future multi-prompt process must re-scope it deliberately. |

**Re-verification after the folds**: `gofmt` clean · `go vet ./...` clean · `make verify` **OK** · `go test -count=1 ./...` green · topology audit unchanged (same 5 pre-existing errors; 357 module rows) · witnesses re-run: (a) force the colour gate on ⇒ the no-colour Example reds; (b) clamp the delta to `+0` ⇒ the new negative-increment pin reds; (c) cap 500 → 499 ⇒ the cap pin reds; (d) render the retired allowance form on the estimated line ⇒ the new turns.log-parity step reds (non-vacuous).
