# Truth Delta: 030-provider-truncation-guard

**Plan Package**: `specs/plans/030-provider-truncation-guard`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added the cross-cutting **Provider output-cap truncation guard** row (both transports read the finish reason; **universal** trigger; the truncation rides the existing `*llm.ProviderError` → frozen `the provider request failed` + exit 6; no new phrase / no new exit code; **no** retry layer; request side unchanged — no default cap; other finish reasons out of scope) and updated the **OpenAI-compatible adapter**, **Vertex/Gemini adapter**, **response-normalization**, and **provider-gateway-port** rows to point at it. The **write-filesystem-tools** row's "#62 forward item" sentence now reads "delivered in round 030"; it also records the TD-1 usage loss (a truncation failure accounts no usage), the TD-3 guard-ordering divergence (the truncation check wins over the generic no-usable-answer path), and the TD-4 `MALFORMED_FUNCTION_CALL` boundary. | Round-030 D1–D7: an output-cap truncation becomes a loud provider failure in both transports; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the truncation guard authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked — the guard persists no state; the existing persisted shapes (`history.jsonl`, `tokens.log`, the prompt logs) are unchanged. | `data-model-covers-all-state` holds — no new state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/refusing-a-cut-off-reply.feature` | New interface feature (chat module) with 2 atomic Rules — (1) a reply cut off while asking tellme to change a file is refused, and the file is untouched (OpenAI-compatible create **and** edit, **and a Gemini `functionCall` create** — the tool-call site carried for **both** families); (2) a reply cut off before the answer is finished is refused, not shown as complete (a cut-off answer is not printed — default and Gemini families). | `acceptance-coverage` for the round-030 acceptance rule `refusing-a-cut-off-reply.feature`. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **Added** a round-030 section: **5 Given** rows (a provider whose **reply** is cut off while **creating** / **editing** a file; a default provider cut off before finishing its answer; a **Gemini** provider cut off before finishing its answer; **and a Gemini `functionCall` create** — the tool-call site for both families) + **2 Then** rows (`tellme creates no file "{path}"`; `tellme prints no answer`). | The new guard steps must be executable and uniquely owned (`dsl-exact-one-match`); the module note records the round. |
| MODIFY | `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | Header comment cross-references the truncation guard (a reply cut off at the output limit is a further cause of the same frozen provider class). | The truncation failure is a further cause of the existing `the provider request failed` surface (`truth-current`). |
