# Tasks — Surface a Gemini/Vertex round's unpaired tool calls as a diagnostic (round 068)

**Plan Package**: `specs/plans/068-unpaired-call-diagnostic`
**Anchor**: **ADR 0037 §Forward RF-067-1** (operator-confirmed) — **DoD: resolve RF-067-1**.
**Mode**: one-shot (`/axb-implement`).
**Pipeline position**: specify ✅ · clarify ✅ (Q1 → A · Q2 → A) · spec-by-example **NOOP (narrowing)** · technical-research ✅ (**ADR 0038**) · system-analysis ✅ · dsl-refine **NOOP (narrowing)** · tasks ✅ · implement ✅ (PR open; human merge only)

## Core Inputs

- `specs/plans/068-unpaired-call-diagnostic/spec.md` (US1; FR-001…FR-007; I-1…I-7) · `research.md` (D1–D7) · `plan.md` · `truth-delta.md`
- `specs/truth/techstack.md` — *Vertex/Gemini adapter* · `docs/decisions/0038-unpaired-call-diagnostic.md`
- `internal/infrastructure/llm/gemini/client_ids_test.go` (the round-067 pins that must stay green under the delegation)

## Phase 0 — Setup

- [x] **T001** [SETUP] Branch `068-unpaired-call-diagnostic` (off `dev`) + the plan package present. No new technology ⇒ Setup omitted (stdlib only).

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research phase): `techstack.md` *Vertex/Gemini adapter* MODIFY + **ADR 0038** + index + the ADR 0037 annotation. **NOOP-with-narrowing** confirmed for `/axb-spec-by-example` + `/axb-dsl-refine` (no hermetic producer for `M < N`).
- [x] **T003** [SETUP] The render port + ui adapter are the landing skeleton for the formatter.

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T004** [BDD-RED][UNIT] `internal/domain/llm/unpaired_test.go`: `TestUnpairedToolCalls` (all-paired · short-round · zero-results · multi-round · out-of-order · id-less-FIFO · duplicate-id · text-closes-round). *Observed RED against the pre-change tree (the function did not exist).*
- [x] **T005** [UNIT] `internal/ui/toolcall_test.go`: `TestFormatUnpairedCalls` (the exact `[Tool Warning]` line; plain). *RED pre-change.*
- [x] **T006** [UNIT] `internal/cli/unpaired_gateway_test.go`: `TestUnpairedGateway_EmitsOnShortRound` · `…_SilentOnHappyPath` · `TestUnpairedEmitter_WritesToStderr` · `…_NilEmitterPassthrough`. *RED pre-change.*

## Phase 3 — Feature

- [x] **T007** [GREEN][IMPL] `internal/domain/llm/unpaired.go` — the family-neutral single owner `UnpairedToolCalls` + `roundPairing`.
- [x] **T008** [IMPL] `internal/infrastructure/llm/gemini/client.go` — `UnpairedCallIDs` delegates to `llm.UnpairedToolCalls`; delete the `dropped`/`unpaired()` machinery; `buildRound` returns contents only.
- [x] **T009** [IMPL] `internal/domain/render/ports.go` (`Lines.UnpairedCalls`) + `internal/ui/render_ports.go` + `internal/ui/toolcall.go` (`FormatUnpairedCalls`).
- [x] **T010** [IMPL] `internal/cli/unpaired_gateway.go` (decorator + `unpairedEmitter`) + `internal/cli/cli.go` `runTurn` wiring.
- [x] **T011** [VERIFY] `gofmt`/`go vet`/`go build` clean · `go test -count=1 ./...` **GREEN** · `make verify` **OK** (arch 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).

## Phase 4 — Regression & witnesses

- [x] **T012** [REGRESSION] The round-065/066/067 pins (`client_ids_test.go`, the batch pins) stay GREEN under the delegation; the media-free byte-identity + OpenAI-compatible pins stay GREEN; the E2E journeys are unchanged; `stdout` byte-exact (I-4).
- [x] **T013** [WITNESS] Reproduce then revert: (a) suppress the emit ⇒ `TestUnpairedGateway_EmitsOnShortRound` RED; (b) suppress the account (`roundPairing.close` a no-op) ⇒ `TestUnpairedToolCalls` RED.
- [x] **T014** [RECORD] Topology audit **5 pre-existing, none new** (no feature/DSL edit); the narrowing recorded at the acceptance/DSL layers; STATUS + `truth-delta.md`; commit + PR (human merge only).

## Residuals (non-blocking — ADR 0038 §Forward)

RF-068-1 no live producer / E2E carrier · RF-068-2 the line is plain · RF-068-3 the loud-failure variant not taken · RF-068-4 eager per-`Complete` walk.

## Conventions

- RED first: T004–T006 were observed failing before T007–T010 landed.
- **Narrowing** (round 059 class): no godog Example can drive `M < N` (no hermetic producer) ⇒ the carriers are unit/CLI-tier pins.

## Outcome (2026-09-20)

- **T004–T006** landed as the three new pin files (domain / ui / cli).
- **T007–T010** implemented: `llm.UnpairedToolCalls` (+ `roundPairing`, factored under the `cyclop` gate: CC 18 → ≤15); the gemini `UnpairedCallIDs` delegation + dead-code removal; the `render.Lines.UnpairedCalls` port + `ui.FormatUnpairedCalls`; the `internal/cli/unpaired_gateway.go` decorator + the `runTurn` wiring (wrapping the resolved gateway, emitting **before** the request, informational).
- **T011** — `gofmt`/`go vet`/`go build` clean · `go test -count=1 ./...` **GREEN** (24 pkgs incl. the godog E2E) · `make verify` **OK** (arch 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).
- **T012** — all existing pins + the E2E journeys stayed GREEN (the delegation preserves round-067 behaviour).
- **T013** — witnesses reproduced then reverted: (a) emit suppressed ⇒ the cli pin RED (`emit got [], want [b]`); (b) account suppressed ⇒ the domain pin RED (`UnpairedToolCalls = [], want [a]`).
- **T014** — no feature/DSL edit ⇒ the topology audit is unchanged (**5 pre-existing, none new**). `go.mod`/`go.sum` unchanged. The **narrowing** is recorded (spec A5 · research D5 · plan §3 · this ledger): no godog Example can drive `M < N`.

### Architect-review folds (2026-09-20, reviewer comment `5746878141` — REQUEST CHANGES, 1 blocker)

- **B-068-1 [ARCHITECTURAL BLOCKER → folded]** — the `[Tool Warning]` id value was interpolated **raw** (no fold/sanitize/cap), violating `sanitize.go`'s formatter contract and the `chrome-tool-values-control-free` invariant (the ids are provider-sourced since round 067). Fixed: `FormatUnpairedCalls` now applies the single-owned policy `capRunes(sanitizeControl(oneLine(strings.Join(ids, ", "))), unpairedIDsCap)` (`unpairedIDsCap = 200`); `sanitize.go` names `[Tool Warning]` in its enumeration; `TestFormatUnpairedCalls_HostileAndCapped` pins the hostile (ESC/CSI/CR/BEL) + 400-rune cases; ADR 0038 D3 + the techstack clause amended (colour ≠ control class, ADR 0008).
- **TD-068-1 [TECHNICAL DEBT → folded]** — "cannot drift by construction" over-claimed (round-067 F-067-3's shared pass is gone). Folded by the **property pin** `TestUnpairedCallIDs_AgreesWithEmittedBody` (for each fixture `totalCalls − len(unpaired) == #functionResponse parts`; every reported id absent from the body) **+** the missing cases in `TestUnpairedToolCalls` (media-before-result, media-after-round, tool-role-media) **+** narrowed wording in ADR 0038 D1/Positive + the techstack clause. Witness: the media-as-boundary mutant now reds the tie pin (was green before).
- **TD-068-2 [TECHNICAL DEBT → folded]** — stale truth/source: the techstack row's round-067 clause retargeted to `llm.UnpairedToolCalls` + a *"superseded by ADR 0038"* pointer; the gemini `UnpairedCallIDs` doc comment corrected (RF-067-1 is delivered; the live consumer is the CLI decorator).
- **N-068-1 [NIT → folded]** — the `[Tool Warning]` line added to the `Chrome` domain entity (`tellme.modelith.yaml` + re-render; `modelith-check` green) — the disposition is now closed.
- **N-068-2 [NIT → folded]** — ADR 0038 D2 documents the per-`Complete` re-reporting (a seen-set is RF-068-4); the spec edge case is finalised (one aggregated, call-order line).
