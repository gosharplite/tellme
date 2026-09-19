# Truth Delta: 065-gemini-parallel-tool-calls

**Plan Package**: `specs/plans/065-gemini-parallel-tool-calls`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` rows RECORDED** (2026-09-20). Clarify **not escalated** (0 questions). `api`/`data` **NOOP**; `dsl-refine` row still **pending** (that phase has not run).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | the round-063 TD-063-1 *qualification* (the interleaved multi-call body) is replaced by the **batched** rule: a round's `functionResponse` parts share **one** `user` turn (call order), the round's **media turns follow** the batch; the pre-065 interleave and the pin rewrite are recorded. | `spec.md` FR-001/FR-002/FR-004; `research.md` D1/D2/D3 |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the adapter **batches a model round's tool results** — N `functionCall` parts ⇒ N `functionResponse` parts in one `user` turn (the `len(functionResponse) == len(functionCall)`-per-turn invariant), FIFO name pairing retained, `ToolCallID` optional, live + replay paths. | `spec.md` FR-001…FR-003; `research.md` D1/D2/D6 |
| MODIFY | `specs/truth/techstack.md` — *Local fake provider* | record that the fake's existing **multi-call scripting** (round 018) is the carrier for the parallel-tool-call acceptance journeys. | `spec.md` A5; `research.md` D5(c) |
| ADD | `docs/decisions/0035-gemini-parallel-tool-call-batching.md` (+ the `docs/decisions/README.md` index row; an annotation on ADR 0033's `Status` + its D2 note + its §Forward RF-063-7) | the decision: a round's tool results share one `user` turn (batched `functionResponse` parts; media turns after). **Supersedes ADR 0033 RF-063-7**; annotates ADR 0033 D2's multi-call scope note; closes [#132](https://github.com/gosharplite/tellme/issues/132). | `spec.md` A4; `research.md` D8 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the `Turn` line and its `steps` are unchanged; media is in-flight only (ADR 0032 D7). | `spec.md` A3 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/chat/*.feature` + `chat/dsl.md` | a parallel-round Rule/Example (a Gemini turn with ≥2 tool calls completes; the recorded request carried the round's results together; a two-image round reaches both images). | `spec.md` US1/US2, FR-001…FR-009; `research.md` D5(c) |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): **NOOP** (checked) — this round changes no modelled entity or invariant (the wire serialization is not a domain concept).

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
