# Truth Delta: 030-provider-truncation-guard

**Plan Package**: `specs/plans/030-provider-truncation-guard`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added the cross-cutting **Provider output-cap truncation guard** row (both transports read the finish reason; **universal** trigger; the truncation rides the existing `*llm.ProviderError` → frozen `the provider request failed` + exit 6; no new phrase / no new exit code; **no** retry layer; request side unchanged — no default cap; other finish reasons out of scope) and updated the **OpenAI-compatible adapter**, **Vertex/Gemini adapter**, **response-normalization**, and **provider-gateway-port** rows to point at it. The **write-filesystem-tools** row's "#62 forward item" sentence now reads "delivered in round 030". | Round-030 D1–D7: an output-cap truncation becomes a loud provider failure in both transports; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/contracts/**` | To be recorded by `/axb-api-plan` (expected NOOP — no OpenAPI/HTTP surface). | Expected: `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/**` | To be recorded by `/axb-data-plan` (expected NOOP — the guard persists no state). | Expected: `data-model-covers-all-state` holds — no new state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/**` | To be recorded by `/axb-dsl-refine` (expected ADD: an interface feature + `dsl.md` rows for the truncation guard; possibly MODIFY `reporting-a-failed-provider-request.feature`). | The truncation failure must be executable interface truth (`acceptance-coverage`, `dsl-exact-one-match`). |
