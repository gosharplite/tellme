# Truth Delta: 067-toolcall-id-followups

**Plan Package**: `specs/plans/067-toolcall-id-followups`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify NOT escalated** (0 questions) — the goal (close [#136](https://github.com/gosharplite/tellme/issues/136)) and the two changes are unambiguous; the residual choices — the observability form (S-3), the fallback spelling (S-4), and the cross-family id scope (S-5) — are left to `/axb-technical-research`. Owner rows land in the phases below.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | *expected MODIFY*: the id **provenance** clause (provider `functionCall.id` preferred, deterministic fallback) + the **unpaired-call accounting** (surfaced, never a silent drop). | `spec.md` US1/US2 (FR-001…FR-010); research D-n |
| *(pending)* | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | *expected MODIFY* (iff the id axis touches the batched-shape wording). | `spec.md` I-2/I-3 |
| *(pending)* | `docs/decisions/0037-*.md` (+ the `docs/decisions/README.md` index row; an annotation on ADR 0036's `Status` / §Forward **RF-066-2**, **RF-066-7**, **RF-066-8**) | *expected ADD*: the id-provenance + unpaired-call accounting + **cross-family** decision. | `spec.md` S-5, A4 |
| *(pending)* | the OpenAI-compatible row | *expected NOOP (checked)* under the S-5 default **(i)** — the wire is byte-frozen. | `spec.md` I-1 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A4 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/data-model.dbml` | No persisted shape changes: the `Turn` line and its `steps` are unchanged; the wire `id` is derived at build time, not a persisted field. | `spec.md` A4 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/**` + `chat/dsl.md` | *expected NOOP* — the round is not user-visible (no shipped surface observes the id value or the unpaired accounting) **unless** S-3 surfaces the observability as a user-visible diagnostic, in which case a small interface Rule/Example lands. | `spec.md` A2/A5; `plan.md` §3 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): *expected NOOP (checked)* — this round changes no modelled entity or invariant (the wire `id` provenance/pairing is not a domain concept); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
