# Truth Delta: 067-toolcall-id-followups

**Plan Package**: `specs/plans/067-toolcall-id-followups`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` RECORDED** (2026-09-20) — `research.md` (D1–D8) + **ADR 0037** + the `techstack.md` *Vertex/Gemini adapter* row MODIFY landed. Decisions: **S-3 = a returned-value accessor** (`UnpairedCallIDs`, not a `stderr` diagnostic), **S-4 = keep the deterministic `call_<n>` fallback**, **S-5 = (i) Gemini-local by construction** (a session is single-family, so the OpenAI-compatible wire is untouched). Skeleton originally initialized by `/axb-specify`. **Clarify NOT escalated** (0 questions).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the response id's **provenance** is the provider's own `functionCall.id` when present/non-empty, else the deterministic `call_<n>` fallback (Gemini-local by construction; the OpenAI-compatible wire unaffected); the `M < N` **boundary drop** is **observable** via the single-owned `UnpairedCallIDs` accessor (never only a silent drop) — the emitted batched turn is unchanged. | `spec.md` US1 (FR-001…FR-005) / US2 (FR-006…FR-010); `research.md` D1/D3/D4 |
| ADD | `docs/decisions/0037-gemini-toolcall-id-provenance.md` (+ the `docs/decisions/README.md` index row; an annotation on ADR 0036's `Status` + §Forward **RF-066-2** (delivered) / **RF-066-7** (delivered) / **RF-066-8** (retired)) | the id-provenance + unpaired-call accounting + cross-family decision. **Extends ADR 0036**; closes [#136](https://github.com/gosharplite/tellme/issues/136). | `spec.md` S-3/S-4/S-5, A4; `research.md` D8 |
| NOOP (checked) | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* row | the round adds no shape change to the media/batch wire. | `spec.md` I-2/I-3; `research.md` D6 |
| NOOP (checked) | the OpenAI-compatible rows *(adapter + image wire)* | unchanged by this round (byte-frozen; I-1 holds by construction — a session is single-family). | `spec.md` I-1; `research.md` D4 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A4 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked, expected) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the `Turn` line and its `steps` are unchanged; the wire `id` is derived at build time, not a persisted field. | `spec.md` A4 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (recorded) | `specs/truth/features/cli/chat/**` + `chat/dsl.md` | **NOOP** — the round is not user-visible: the unpaired-call observability is a **code accessor** (S-3 → returned value, not a `[Tool …]` diagnostic), and the provider-id preference has no shipped user surface. The round-065/066 journeys are unchanged. | `spec.md` A2/A5; `plan.md` §3 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): *expected NOOP (checked)* — this round changes no modelled entity or invariant (the wire `id` provenance/pairing is not a domain concept); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
