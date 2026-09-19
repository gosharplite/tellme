# Tasks — Gemini/Vertex parallel tool calls: a round's tool results share one turn (round 065)

**Plan Package**: `specs/plans/065-gemini-parallel-tool-calls`
**Anchor issue**: [#132](https://github.com/gosharplite/tellme/issues/132)
**Mode**: one-shot (`/axb-implement`) — a single round, test-aligned first (RED), then GREEN.
**Pipeline position**: specify ✅ · clarify ✅ (not escalated) · spec-by-example ✅ · technical-research ✅ (ADR 0035) · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · implement ✅

## Core Inputs

- `specs/plans/065-gemini-parallel-tool-calls/spec.md` (US1/US2; FR-001…FR-009; I-1…I-6; SC-001…SC-006)
- `specs/plans/065-gemini-parallel-tool-calls/research.md` (D1…D9) · `plan.md`
- `specs/plans/065-gemini-parallel-tool-calls/truth-delta.md` (all owner rows RECORDED)
- `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* · *Vertex/Gemini adapter* · *Local fake provider*
- `docs/decisions/0035-gemini-parallel-tool-call-batching.md`
- `specs/truth/features/cli/chat/calling-several-tools-in-one-round.feature` + `specs/truth/features/cli/chat/dsl.md` (`## Given (round 065)` / `## Then (round 065)`)

## Phase 0 — Setup

- [x] **T001** [SETUP] Confirm the branch `065-gemini-parallel-tool-calls` (off `dev`) and the plan package present. No new technology ⇒ **Setup omitted** (stdlib `encoding/json` already imported).

## Phase 1 — Foundational

- [x] **T002** [SETUP] Record the truth/ADR (research + dsl-refine phases): `techstack.md` ×3 rows MODIFY + **ADR 0035** + index + the ADR 0033 annotation + the truth feature + the `dsl.md` round-065 rows/note.
- [x] **T003** [SETUP] E2E landing skeleton: `tests/e2e/steps/step_r065_multi.go` (the multi-tool Givens + the batched-shape Then registered via `init()`).

## Phase 2 — Test Alignment (RED, no product change)

- [x] **T004** [BDD-RED][UNIT] Rewrote `TestRequestBody_MultiCallRound_MediaTurnsInterleave` → `TestRequestBody_MultiCallRound_BatchesFunctionResponses` (two-call + two-image round ⇒ 4 `contents`: model · one `user` turn with TWO `functionResponse` parts · mediaA · mediaB); **added** `TestRequestBody_MultiCallRound_NoMedia_BatchesResults` (the #132 shape) and `TestRequestBody_ThreeCallRound_BatchesResults`. *Observed RED against the pre-change tree — see the witness.*
- [x] **T005** [BDD-RED][E2E] The new Givens + Then in `tests/e2e/steps/step_r065_multi.go`. *Observed RED against the pre-change tree — see the witness.*

## Phase 3 — Feature

- [x] **T006** [GREEN][IMPL] `internal/infrastructure/llm/gemini/client.go` `buildContents`: buffer a round's `functionResponse` parts + media; **flush** on the next model/plain-text turn and before the prompt ⇒ one batched `user` turn, then the media turns. `inlineDataParts` + the `pending` FIFO reused; the OpenAI-compatible wire, the loop, ports, tools and config untouched (I-1).
- [x] **T007** [VERIFY] `gofmt -l .` clean · `go build ./...` · `go vet ./...` · `go test -count=1 ./...` **GREEN** (271 scenarios incl. the 3 new ones) · `make verify` **OK** (arch 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4).

## Phase 4 — Regression & witnesses

- [x] **T008** [REGRESSION] The existing pins stayed GREEN: `TestRequestBody_MediaBecomesInlineData` (N = 1 ⇒ 3 turns — I-2) · `TestRequestBody_MediaFreeIsByteIdentical` (I-3) · the OpenAI-compatible pins · the round-019 multi-tool E2E.
- [x] **T009** [WITNESS] Reproduced then reverted: re-introducing the pre-065 per-message emission turned **T004 RED** (3 unit pins, e.g. `contents len = 5, want 4`) **and** the E2E Then RED (`found 2 response turn(s), 0 batched`), the recorded body showing the split turns Vertex rejects. Reverted ⇒ all green.
- [x] **T010** [RECORD] `truth-delta.md` dsl-refine row RECORDED · topology audit **5 pre-existing, none new** (49 features · 384 module rows · 1989 steps) · STATUS updated · commit + PR (human merge only).

## Residuals (non-blocking — ADR 0035 §Forward)

RF-065-1 `ToolCallID`-keyed pairing · RF-065-2 one merged media turn · RF-065-3 a fake-side contract check · RF-065-4 a family-agnostic batching concept · RF-065-5 the untouched RF-063-x items.

## Conventions

- RED first: T004/T005 were observed failing against the pre-change tree before T006 landed.
