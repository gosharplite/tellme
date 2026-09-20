# Tasks — Gemini/Vertex tool-call id follow-ups (round 067)

**Plan Package**: `specs/plans/067-toolcall-id-followups`
**Anchor issue**: [#136](https://github.com/gosharplite/tellme/issues/136) (ADR 0036 §Forward RF-066-2 + RF-066-7; retires RF-066-8) — **DoD: close #136**.
**Mode**: one-shot (`/axb-implement`) — a single round, test-aligned first (RED), then GREEN.
**Pipeline position**: specify ✅ · clarify ✅ (not escalated) · spec-by-example **NOOP** · technical-research ✅ (**ADR 0037**) · system-analysis ✅ · dsl-refine **NOOP** · tasks ✅ · implement ✅ (PR open; human merge only)

## Core Inputs

- `specs/plans/067-toolcall-id-followups/spec.md` (US1/US2; FR-001…FR-010; SC-001…SC-007; I-1…I-8)
- `specs/plans/067-toolcall-id-followups/research.md` (D1–D8) · `plan.md`
- `specs/plans/067-toolcall-id-followups/truth-delta.md` (technical-research RECORDED; api/data/dsl-refine NOOP)
- `specs/truth/techstack.md` — *Vertex/Gemini adapter*
- `docs/decisions/0037-gemini-toolcall-id-provenance.md` (+ ADR 0036 annotated)
- `internal/infrastructure/llm/gemini/client_ids_test.go` (the round-066 id pins this round extends)

## Phase 0 — Setup

- [x] **T001** [SETUP] Confirm the branch `067-toolcall-id-followups` (off `dev`) and the plan package present. No new technology ⇒ **Setup omitted** (stdlib `encoding/json`/`fmt` already imported).

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research phase, already landed): `techstack.md` *Vertex/Gemini adapter* row MODIFY + **ADR 0037** + index + the ADR 0036 annotation + `plan.md` + the `truth-delta.md` owner rows. **NOOP confirmed** for `/axb-spec-by-example` (no user-visible change) and `/axb-dsl-refine` (no feature/DSL edit).
- [x] **T003** [SETUP] No new test-landing skeleton needed — the pins extend the existing `internal/infrastructure/llm/gemini/client_ids_test.go` (same package).

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T004** [BDD-RED][UNIT] `client_ids_test.go`: ADD `TestParseResponse_PrefersProviderFunctionCallID` (SC-002 — a Vertex response carrying `functionCall.id` ⇒ `llm.ToolCall.ID` is that id, and the emitted parts carry it) **and** `TestParseResponse_FallsBackToDeterministicID` (SC-002 — a response without a provider id ⇒ `call_<n>`; a blank id ⇒ fallback). *Observed RED against the pre-change tree.*
- [x] **T005** [BDD-RED][UNIT] ADD `TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall` (SC-001/SC-007 — an `N=2 M=1` round ⇒ `UnpairedCallIDs` returns exactly `["call_2"]`, the unpaired call's id, in call order) **and** `TestUnpairedCallIDs_AllPairedIsEmpty` (SC-001 — `M=N` ⇒ empty) **and** `TestUnpairedCallIDs_MultiRound` (call order across two rounds). *Observed RED against the pre-change tree.*

## Phase 3 — Feature

- [x] **T006** [GREEN][IMPL] `internal/infrastructure/llm/gemini/client.go` `parseResponse`: read the provider `functionCall.id`; prefer it when non-empty, else the deterministic `call_<n>` (`callID` helper). `roundBuilder`: add `unpaired()` (the round's unpaired call ids, call order), record them in `flush()` (`b.dropped`) at the one drop site, factor the message switch into `consume()`, and add the package-level `UnpairedCallIDs(prior)` accessor. `buildContents` output unchanged; the OpenAI-compatible wire, the loop, ports, tools and config untouched (I-1).
- [x] **T007** [VERIFY] `gofmt -l .` clean · `go build ./...` · `go vet ./...` · `go test -count=1 ./...` **GREEN** · `make verify` **OK** (arch 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).

## Phase 4 — Regression & witnesses

- [x] **T008** [REGRESSION] The existing pins stay GREEN: the round-066 id pins (`TestRequestBody_ToolPartsCarryIDs`, `…OutOfOrderResultsPairByIdentity`, `…EmptyToolCallIDOmitsID`, `…UnmatchedToolCallIDFallsBackToFIFO`, `…ReplayedStepIDsPairByIdentity`, `…ToolRoleMediaStillCarriesMedia`) · the round-065 batch pins · `TestRequestBody_MediaFreeIsByteIdentical` (I-3) · `TestRequestBody_ShortRound_DropsUnpairedNames` (M < N) · the OpenAI-compatible byte pins (I-1) · the round-065/066 E2E journeys (unchanged).
- [x] **T009** [WITNESS] Reproduce then revert: (a) remove the provider-id preference ⇒ `TestParseResponse_PrefersProviderFunctionCallID` **RED**; (b) suppress the accessor (return nothing / drop without recording) ⇒ `TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall` **RED**; (c) make the fallback non-deterministic ⇒ `TestParseResponse_FallsBackToDeterministicID` **RED**; (d) **the RF-066-8 kill (review F-067-1)** — reinstate the pre-fold *conditional partial drop* (retain a stale `pending` slice instead of the unconditional `b.pending = nil`) ⇒ `TestUnpairedCallIDs_MultiRound` **RED** while `TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall` stays **green** (the `N=2 M=1` arithmetic equivalence the residual named) — reproduced then reverted.
- [x] **T010** [RECORD] Update `truth-delta.md` (implement outcome) · topology audit **5 pre-existing, none new** (no feature/DSL edit) · STATUS updated · commit + PR (human merge only).

## Residuals (non-blocking — ADR 0037 §Forward)

RF-067-1 the observability is a returned-value accessor (no `stderr` diagnostic) · RF-067-2 the provider-id preference is live-unverified hermetically · RF-067-3 no E2E asserts the wire `id` · RF-067-4 concurrent execution not added · RF-067-5 the reference fallback spelling not adopted.

## Conventions

- RED first: T004/T005 were observed failing against the pre-change tree before T006 landed.
- A hardening/parity round — no user-visible behaviour change; the witness is request-body **shape** + the accessor.

## Outcome (2026-09-20)

- **T004/T005** landed in `internal/infrastructure/llm/gemini/client_ids_test.go` (NEW: `TestParseResponse_PrefersProviderFunctionCallID`, `TestParseResponse_FallsBackToDeterministicID`, `TestUnpairedCallIDs_ShortRound_ReportsUnpairedCall`, `TestUnpairedCallIDs_AllPairedIsEmpty`, `TestUnpairedCallIDs_MultiRound`).
- **T006** implemented in `internal/infrastructure/llm/gemini/client.go`: `parseResponse` reads `functionCall.id` and prefers it (`callID`); `roundBuilder` gains `unpaired()` + `dropped` (recorded at the one drop site in `flush()`), the message switch is factored into `consume()`, and the package-level `UnpairedCallIDs(prior)` accessor is added. `buildContents` output unchanged.
- **T007** — `gofmt` clean · `go build ./...` · `go vet ./...` · `go test -count=1 ./...` **GREEN** · `make verify` **OK** (arch gate 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).
- **T008** — all existing pins + the round-065/066 E2E journeys stayed GREEN (the media-free byte-identity pin, the OpenAI-compatible pins, the round-066 id pins).
- **T009** — witnesses reproduced then reverted: (a) reverting the provider-id preference ⇒ the preference pin RED; (b) not recording unpaired ids in `flush()` ⇒ the short-round accessor pin RED; (c) a positional/non-deterministic fallback ⇒ the fallback pin RED; (d) reinstating the pre-fold conditional partial drop ⇒ `TestUnpairedCallIDs_MultiRound` RED (the RF-066-8 kill; F-067-1).
- **T010** — no feature/DSL edit ⇒ the topology audit is unchanged (**5 pre-existing, none new**; no `specs/truth/features/**` change). `go.mod`/`go.sum` unchanged. **RF-066-8 retired** (the exact-identity accessor pin kills the partial-drop mutant).

### Architect-review folds (2026-09-20, reviewer comment `5746659904` — APPROVE WITH REQUIRED FOLDS, 0 blockers)

- **F-067-1 [TECHNICAL DEBT → folded]** — the RF-066-8 retirement was mis-attributed to the short-round pin; the real carrier is the **cross-round** account (`TestUnpairedCallIDs_MultiRound`). Re-attributed on all five surfaces (`client_ids_test.go` doc comment · `research.md` §D2 · ADR 0037 §D5/§Consequences · ADR 0036 §Forward RF-066-8 · the `techstack.md` *Vertex/Gemini adapter* row) + the mutant-kill witness recorded as **T009(d)**.
- **F-067-2 [TECHNICAL DEBT → folded, option (b)]** — "the drop is no longer silent" outran the delivery (`UnpairedCallIDs` has no live consumer). Scoped the wording to **accountable in code; silent at runtime** across `spec.md` (US1 · S-1 · FR-001/FR-002 · SC-001) · ADR 0037 §D4/§Consequences · the `client.go` doc comment · the truth clause; `RF-067-1` records the (operator-gated) user-visible-diagnostic option.
- **F-067-3 [REFACTOR → folded]** — one builder: `buildRound(prompt, prior) (contents, unpaired)` is the single pass; `buildContents` and `UnpairedCallIDs` delegate, so the account and the emitted body cannot disagree by construction.
- **F-067-4 [NIT → folded]** — `callID` now returns `strings.TrimSpace(providerID)` on the non-empty branch (one whitespace normalisation); pinned by `TestParseResponse_TrimsProviderID`.
- **F-067-5 [NIT → folded]** — the `M = 0` **middle**-round edge pinned by `TestUnpairedCallIDs_MiddleRoundZeroResults`.
- **F-067-6 [NIT → folded]** — ADR 0037 D3's locality reason corrected to the load-bearing invariant (**the id is never persisted/replayed**); a guarding note on `history.Step` recorded as RF-067-7.
- **F-067-7 [NIT → folded]** — the degenerate duplicate-provider-id determinism pinned by `TestUnpairedCallIDs_DuplicateProviderIDIsDeterministic` + recorded as RF-067-6.
- **Held / no action** — I-1/I-2/I-3/I-4/I-5/I-8 (the architect verified the change family-local and the shapes preserved); governance hygiene confirmed.
