# Tasks — Gemini cached-token usage (round 072)

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`

Legend: `[ ]` pending · `[X]` done.

## Phase 1 — Setup

- [X] **T001** `internal/infrastructure/llm/gemini/client.go`: add `cachedContentTokenCount` + `thoughtsTokenCount` to the `usageMetadata` decode struct.
- [X] **T002** Map them into `llm.Usage.CachedTokens`/`ThinkingTokens` (each `max(0, …)`), family-disjoint (no subtraction).

## Phase 2 — Foundational

- [X] **T003** Confirm the cost path is untouched (`call_renderer.go` `miss = prompt − cached`; `pricing.go` `ComputeCost`) — correctness comes from populating the input.

## Phase 3 — Test Alignment & Implementation

- [X] **T004 (RED)** `gemini/client_test.go`: `TestParseResponse_DecodesCachedAndThinkingTokens` (cached+thinking decode; disjoint) + `TestParseResponse_AbsentCachedFieldsStayZero` (degenerate).
- [X] **T005 (RED)** `tests/e2e/fakeprovider`: extend the Vertex answer body to emit `cachedContentTokenCount`/`thoughtsTokenCount` when `detailed` (disjoint `candidatesTokenCount`).
- [X] **T006 (RED)** `tests/e2e/steps/step_r072_gemini_usage.go`: the Gemini detailed-usage Given.
- [X] **T007 (RED)** `presenting-the-post-turn-status.feature`: a Gemini metrics Example (`1026 missed, 389538 cached, 100 completed, 4096 reasoning`).
- [X] **T008 (GREEN)** Implement + run: unit pins + the E2E Example green.

## Phase 4 — Feature (Green / Refactor)

- [X] **T009 (witness)** Revert `CachedTokens` to 0 ⇒ the unit pin AND the E2E Example RED (reproduced then reverted).
- [X] **T010** Truth: `specs/truth/techstack.md` (Vertex/Gemini adapter + Post-turn status rows) MODIFY; `chat/dsl.md` Given row.
- [X] **T011** Domain model: `docs/domain-model/tellme.modelith.yaml` `UsageRecord` (the family-decoded cached/thinking counts) + `make modelith-render`.
- [X] **T012** **ADR 0044** + index row.

## Phase 5 — Delivery

- [X] **T013** `make verify` + `go test -count=1 ./...` green; `go.mod`/`go.sum` unchanged; OpenAI-compatible wire byte-identical.
