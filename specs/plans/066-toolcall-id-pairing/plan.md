# System Analysis Plan — Round 066: pair a round's tool results to their calls by `ToolCallID` (Gemini/Vertex)

**Plan package**: `specs/plans/066-toolcall-id-pairing`
**Owner**: `/axb-system-analysis`
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1/US2 · FR-001…FR-010 · SC-001…SC-007 · I-1…I-7 · A1…A7) · `research.md` (D1–D8) · `truth-delta.md` · `specs/truth/**`.

## 1. Interface inventory

| Interface | Kind | Analysis planner | Disposition |
| --- | --- | --- | --- |
| The CLI end (the `tellme` binary) | `cli` | none (a line-oriented CLI) | **Carried forward to its contract owner `/axb-dsl-refine`** at delivery (the round-065 precedent). |

**One** CLI interface. There is **no** HTTP/API surface and **no** new persisted state, so:

- `/axb-api-plan` → **NOOP** (no `specs/truth/contracts/**`). `contract-authoritative` holds vacuously.
- `/axb-data-plan` → **NOOP** (no `specs/truth/data/**` change). The round adds no persisted field — the wire `id` is derived at build time from the existing `llm.ToolCall.ID` / `llm.Message.ToolCallID`, so `data-model-covers-all-state` holds.
- `/axb-ui-plan` → **skipped** (a plain line-oriented CLI; no TUI surface changes).

## 2. Waves

**No dependency waves.** The round is **adapter-local** (`internal/infrastructure/llm/gemini`): a pure request-body transform plus unit pins. There is no shared interface to sequence and nothing to fan out; the work is one coherent edit.

## 3. Contract owner hand-off — `/axb-dsl-refine`

The CLI end is carried to **`/axb-dsl-refine`**. Expected disposition: **NOOP**.

- The round changes no **user-visible** behaviour (the loop is sequential, so the pre-066 positional pairing already produced the correct wire; SC-001…SC-007 are request-body **shape** assertions, not new user journeys).
- The round-065 interface journeys (`specs/truth/features/cli/chat/calling-several-tools-in-one-round.feature`) are **unchanged** — they assert that a multi-call round completes and the recorded request carried the round's results together, which still holds.
- A new **id-observing** Then would require the E2E fake to expose the wire ids (RF-065-3 / RF-066-3, a separate forward item) — so `/axb-dsl-refine` records a **NOOP** unless that carrier is taken in-round (it is not).
- The topology audit (`axb-gherkin-and-dsl` over `specs/truth/features/cli`) must keep the **same 5 pre-existing errors**, **none new** (no feature/DSL edit).

## 4. Component / tier view (the changed surface)

| Site | Tier | Change |
| --- | --- | --- |
| `internal/infrastructure/llm/gemini/client.go` `buildContents` | infrastructure (tier 4) | emit `id` on each emitted `functionCall`/`functionResponse` part (from `tc.ID` / `m.ToolCallID`, omitted when empty); pair each result to its call **by `ToolCallID`** (per-round `(id → name)` index), FIFO name fallback retained; the round-065 batching + media-after-batch placement unchanged. |
| `internal/infrastructure/llm/gemini/client.go` `parseResponse` | infrastructure (tier 4) | **no change** (D3 keeps the existing `call_<n>` ids). |
| `internal/agent/agentloop.go` | agent (tier 5) | **no change** (it already sets `ToolCallID` on every result; the replay id is `call_step_<n>`). |
| `internal/infrastructure/llm/openai/**` | infrastructure (tier 4) | **no change** (I-1: byte-frozen). |

**Gate impact:** none — no new import edge, no `internal/cli` change, so `verify-architecture`'s RULE-A…F baseline stays header-only/empty. No `make verify` member is added.

## 5. Artifact tree (plan-side + truth-side)

```
specs/plans/066-toolcall-id-pairing/
  spec.md            (axb-specify)
  checklists/requirements.md
  research.md        (axb-technical-research)
  plan.md            (this file — axb-system-analysis)
  tasks.md           (axb-tasks — next)
  truth-delta.md
specs/truth/techstack.md                      MODIFY (×2 rows)
docs/decisions/0036-toolcall-id-pairing.md    ADD (+ README index; ADR 0035 annotated)
```

## 6. Task directives (handed to `/axb-tasks`)

1. **Phase 3 (RED first)** — land the unit pins over the built request body *before* the change: (a) every `functionCall`/`functionResponse` part carries an `id` (response id == call id); (b) an out-of-order result set pairs by identity; (c) an empty-`ToolCallID` result omits the id and pairs by FIFO; (d) the media-free text body is byte-identical; (e) the OpenAI-compatible body is byte-identical. Extend the round-065 batch pins (`TestRequestBody_MultiCallRound_BatchesFunctionResponses` etc.) with the id assertion.
2. **Phase 4 (GREEN)** — implement the id-link + id-keyed pairing in `buildContents`; keep the round-065 batch shape and media placement; re-anchor the `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual to exact unmatched-id accounting if the pairing change reaches it (SC-005).
3. **Regression** — the round-065 E2E journeys stay green unchanged; `make verify` + `go test -count=1 ./...`; topology audit unchanged (no feature/DSL edit); `go.mod`/`go.sum` unchanged.
4. **`[BDD-REMOVE]` applicant** — none (no dead stepdef).
