# Truth Delta: 066-toolcall-id-pairing

**Plan Package**: `specs/plans/066-toolcall-id-pairing`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify NOT escalated** (0 questions) — the goal (close [#134](https://github.com/gosharplite/tellme/issues/134)) and the fix are unambiguous; the residual choices (id provenance/fallback, unmatched accounting) are left to `/axb-technical-research` (S-4/S-6). Owner rows below are **expected** shapes, not yet recorded.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the adapter emits an **`id` on every `functionCall`/`functionResponse` part** (when non-empty) and binds each result to its call **by `ToolCallID`** (order-independent), with **FIFO name matching retained as the fallback** for a result with no id; the round-065 clause "`ToolCallID` optional" is updated. Live + replay paths. | `spec.md` FR-001…FR-009; `research.md` (TBD) |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | the batched-function-response rule gains the **id axis** (the batched `user` turn's parts carry ids); the round-065 *qualification* is extended, not replaced. | `spec.md` FR-002/FR-009; `research.md` (TBD) |
| ADD | `docs/decisions/00NN-*.md` (+ the `docs/decisions/README.md` index row; an annotation on ADR 0035's `Status` + its **D2** FIFO note + its §Forward **RF-065-1**) | the id-keyed pairing decision. **Extends ADR 0035**; homes #134. | `spec.md` A4; `research.md` (TBD) |
| NOOP (checked, expected) | the OpenAI-compatible row | unchanged by this round (the `tool_call_id` wire stands, byte-frozen — I-1). | `spec.md` I-1 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A4 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked, expected) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the round-055 `Turn` line and its `steps` are unchanged; the wire `id` is derived at build time, not a persisted field. | `spec.md` A4 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP **or** MODIFY (expected — research decides) | `specs/truth/features/cli/chat/calling-several-tools-in-one-round.feature` + `chat/dsl.md` | **NOOP** if the carrier stays unit-level (the round-065 journeys are unchanged — no user-visible behaviour change); **MODIFY** (a small id-observing Then + a `dsl.md` row) **iff** the fake provider can observe the wire ids. | `spec.md` A5; `research.md` (TBD) |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): **NOOP** (checked) — this round changes no modelled entity or invariant (the wire `id` pairing is not a domain concept); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
