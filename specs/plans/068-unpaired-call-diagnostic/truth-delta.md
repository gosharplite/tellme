# Truth Delta: 068-unpaired-call-diagnostic

**Plan Package**: `specs/plans/068-unpaired-call-diagnostic`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify ESCALATED — 2 questions open (Q1 surface/routing · Q2 loudness)**, and the round **blocks on them** (RF-067-1 is an operator-gated user-visible surface addition). Residual technical choice (the detection seam) defers to `/axb-technical-research`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — *Vertex/Gemini adapter* (and/or *Agent tool loop*) | *expected MODIFY*: the `M < N` boundary drop is surfaced as a user-visible diagnostic (not only an in-code account). | `spec.md` US1 (FR-001…FR-007) |
| *(pending)* | `docs/decisions/0038-*.md` (or an ADR-0037 forward-annotation) | *expected ADD*: the diagnostic decision + the detection seam. | `spec.md` A4/S-6 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `spec.md` A4 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/data-model.dbml` | No persisted shape change (the diagnostic is transient; `turns.log` is a rendered-text file, not a modelled record). | `spec.md` A4 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/**` + `dsl.md` | *expected MODIFY*: a small interface Rule/Example for the unpaired-call diagnostic (`M < N` ⇒ a `stderr` diagnostic; `M == N` ⇒ none). | `spec.md` A5 (user-visible → not NOOP) |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive): *expected MODIFY/NOOP* — decided when the diagnostic's domain concept (if any) is fixed; `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; every new/changed step must match exactly one `DSLRow`.
