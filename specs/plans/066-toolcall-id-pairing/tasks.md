# Tasks — Gemini/Vertex: pair a round's tool results to their calls by `ToolCallID` (round 066)

**Plan Package**: `specs/plans/066-toolcall-id-pairing`
**Anchor issue**: [#134](https://github.com/gosharplite/tellme/issues/134) (ADR 0035 §Forward RF-065-1) — **DoD: close #134**.
**Mode**: one-shot (`/axb-implement`) — a single round, test-aligned first (RED), then GREEN.
**Pipeline position**: specify ✅ · clarify ✅ (not escalated) · spec-by-example **NOOP** · technical-research ✅ (**ADR 0036**) · system-analysis ✅ · dsl-refine **NOOP** · tasks ✅ · implement ✅ (PR [#135](https://github.com/gosharplite/tellme/pull/135) open; architect review folded)

## Core Inputs

- `specs/plans/066-toolcall-id-pairing/spec.md` (US1/US2; FR-001…FR-010; SC-001…SC-007; I-1…I-7)
- `specs/plans/066-toolcall-id-pairing/research.md` (D1–D8) · `plan.md`
- `specs/plans/066-toolcall-id-pairing/truth-delta.md` (technical-research RECORDED; api/data/dsl-refine NOOP)
- `specs/truth/techstack.md` — *Vertex/Gemini adapter* · *Image content on the provider wire (Gemini/Vertex)*
- `docs/decisions/0036-toolcall-id-pairing.md` (+ ADR 0035 annotated)
- `internal/infrastructure/llm/gemini/client_image_test.go` (the round-065 batch pins this rounds extends)

## Phase 0 — Setup

- [x] **T001** [SETUP] Confirm the branch `066-toolcall-id-pairing` (off `dev`) and the plan package present. No new technology ⇒ **Setup omitted** (stdlib `encoding/json` already imported).

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research phase, already landed): `techstack.md` ×2 rows MODIFY + **ADR 0036** + index + the ADR 0035 annotation + the `plan.md` + the `truth-delta.md` owner rows. **NOOP confirmed** for `/axb-spec-by-example` (no user-visible change) and `/axb-dsl-refine` (no feature/DSL edit).
- [x] **T003** [SETUP] No new test-landing skeleton needed — the pins extend the existing `internal/infrastructure/llm/gemini/client_image_test.go` (same package).

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T004** [BDD-RED][UNIT] `client_image_test.go`: ADD `TestRequestBody_ToolPartsCarryIDs` (SC-001/SC-002 — no media: every `functionCall`/`functionResponse` part carries an `id`; a response's id equals its call's id) **and** `TestRequestBody_OutOfOrderResultsPairByIdentity` (SC-002 — results presented `B, A` bind to their own call's name+id, emitted in call order) **and** `TestRequestBody_EmptyToolCallIDOmitsID` (FR-003/SC-003 — a result with an empty `ToolCallID` emits **no** `id` key and pairs by FIFO). *Observed RED against the pre-change tree.*
- [x] **T005** [UNIT] Extend the round-065 pins with the id assertion: `TestRequestBody_MultiCallRound_NoMedia_BatchesResults` asserts each `functionResponse` part's `id`; `TestRequestBody_ThreeCallRound_BatchesResults` likewise. *Observed RED against the pre-change tree.*

## Phase 3 — Feature

- [x] **T006** [GREEN][IMPL] `internal/infrastructure/llm/gemini/client.go` `buildContents`: emit `id` on each `functionCall` part (from `tc.ID`, omitted when empty) and each `functionResponse` part (from `m.ToolCallID`, omitted when empty); bind each result to its call **by `ToolCallID`** (a per-round ordered call index), with the FIFO name match retained as the **fallback**; emit the batched turn's parts **in call order** (the round-065 shape unchanged). `parseResponse` unchanged (D3); the OpenAI-compatible wire, the loop, ports, tools and config untouched (I-1).
- [x] **T007** [VERIFY] `gofmt -l .` clean · `go build ./...` · `go vet ./...` · `go test -count=1 ./...` **GREEN** · `make verify` **OK** (arch 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).

## Phase 4 — Regression & witnesses

- [x] **T008** [REGRESSION] The existing pins stay GREEN: `TestRequestBody_MediaBecomesInlineData` (N = 1) · `TestRequestBody_MultiCallRound_BatchesFunctionResponses` (the two-image round *shape* — roles/order/blob bytes; the added id is the sole difference) · `TestRequestBody_MediaFreeIsByteIdentical` (I-3, **no** id key on the text path) · `TestRequestBody_ShortRound_DropsUnpairedNames` (M < N) · the OpenAI-compatible byte pins (I-1) · the round-065 E2E journeys (unchanged).
- [x] **T009** [WITNESS] Reproduce then revert: (a) drop the `id` from the wire ⇒ `TestRequestBody_ToolPartsCarryIDs` **RED**; (b) pair positionally (ignore `ToolCallID`) ⇒ `TestRequestBody_OutOfOrderResultsPairByIdentity` **RED**; (c) drop the FIFO fallback ⇒ `TestRequestBody_EmptyToolCallIDOmitsID` **RED**.
- [x] **T010** [RECORD] Re-anchor the `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual note to the exact unmatched-identity accounting (SC-005) · topology audit **5 pre-existing, none new** (no feature/DSL edit) · STATUS updated · commit + PR (human merge only).

## Residuals (non-blocking — ADR 0036 §Forward)

RF-066-1 the narrowed byte-identity claim (shape-identical tool-bearing body) · RF-066-2 provider-issued id preference · RF-066-3 the fake-side contract check · RF-066-4 concurrent tool execution · RF-066-5 the other round-065 forward items.

## Conventions

- RED first: T004/T005 were observed failing against the pre-change tree before T006 landed.
- A hardening/parity round — no user-visible behaviour change; the witness is request-body **shape**.

## Outcome (2026-09-20)

- **T004/T005** landed as `internal/infrastructure/llm/gemini/client_ids_test.go` (NEW: `TestRequestBody_ToolPartsCarryIDs`, `TestRequestBody_OutOfOrderResultsPairByIdentity`, `TestRequestBody_EmptyToolCallIDOmitsID`, `TestRequestBody_UnmatchedToolCallIDFallsBackToFIFO`) + the id assertion extended into the two round-065 pins.
- **T006** implemented: `buildContents` was refactored into a `roundBuilder` (modelTurn/result/bind/flush/textTurn + `functionCallPart`) — both to carry the id link and to stay under the `cyclop` lint gate (the refactor was forced by `make verify`, CC 26 → ≤15).
- **T007** — `gofmt` clean · `go build ./...` · `go vet ./...` · `go test -count=1 ./...` **GREEN** (24 pkgs incl. the godog E2E) · `make verify` **OK** (arch gate 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).
- **T008** — all existing pins stayed GREEN (the media N=1 pin, the two-image batch *shape* pin, the media-free byte-identity pin, the M<N short-round pin, the OpenAI-compatible pins, the round-065 E2E journeys).
- **T009** — witnesses reproduced then reverted: (a) dropping the wire id ⇒ the id pins RED; (b) disabling the id match ⇒ the out-of-order pin RED (`part 0 = {id:call_B name:read_files …}`, the FIFO mispair); (c) disabling the FIFO fallback ⇒ the two fallback pins RED.
- **T010** — the `TestRequestBody_ShortRound_DropsUnpairedNames` residual is now **exact unmatched-identity accounting** (the round-065 `pending[:0]` positional drop is replaced by the per-round index; the deferred `N=2 M=1` note is re-anchored, not re-scoped). No feature/DSL edit ⇒ the topology audit is unchanged (**5 pre-existing, none new**; no `specs/truth/features/**` change). `go.mod`/`go.sum` unchanged.
- **Design note (recorded)** — FR-003 made reachable: the result case is now `m.ToolCallID != "" || m.Role == "tool"`, so an id-less `tool`-role message serializes as a FIFO-paired `functionResponse` (no `id` key) instead of a stray text turn. No live producer emits one (both the loop and the replay set `ToolCallID`); it closes the spec's id-less case.

### Architect-review folds (2026-09-20, reviewer `5258004457` — APPROVE WITH REQUIRED FOLDS)

- **F-066-1** — the "exact unmatched-identity accounting" claim was not delivered; **corrected** in `spec.md` (S-6 / FR-008 / SC-005) + `research.md` D5 → the boundary drop is unchanged; the residual is re-homed as **RF-066-7/RF-066-8** (ADR 0036 §Forward).
- **F-066-2** — the id-less-`tool` widening is now restricted to a **media-free** `tool` message (`client.go` + a comment), so a media-bearing `tool`-role message is still carried; pinned by `TestRequestBody_ToolRoleMediaStillCarriesMedia`.
- **TD-066-1** — a **foreign** response `id` (not equal to its bound call's id) is omitted (wire self-consistency); pinned by `TestRequestBody_UnmatchedToolCallIDFallsBackToFIFO` (updated) + ADR D2/D4 + research D4.
- **TD-066-2** — the replay mechanism is **id-primary** (the loop sets the same `call_step_<n>` on both sides); a pin `TestRequestBody_ReplayedStepIDsPairByIdentity` added, and the wording corrected in ADR D3/D6 + the truth row.
- **N-066-1** — `spec.md` Status + this `tasks.md` pipeline line updated. **N-066-2** — the *Vertex/Gemini adapter* truth row restores round-065's clause and **appends** the round-066 sentence.
- **R-066-1** — the live check is performed at the closeout and recorded in ADR 0036 `## Verification` (*Live*).
