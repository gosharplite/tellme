# Truth Delta: 031-tool-schema-wellformedness

**Plan Package**: `specs/plans/031-tool-schema-wellformedness`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added an **Agent tool schemas** row (CLI Application): the shared schema builder declares the resource params **and the mandatory `reason` property**, so every tool it backs satisfies `required ⊆ properties`; `execute_command` builds inline and already complies; records the defect's regression origin (round-024 `cfa005c`, carried by round-029 `eb0367c`). Added an **Agent tool-schema gate** row (Testing & Verification): a hermetic unit check over the production assembler `agentTools()` asserting `required ⊆ properties` + schema-parses for **every** registered tool. Added a **Not Introduced Yet** bullet recording the deferred stronger guards (fake schema-validation; a live pipeline leg; a structural builder). | Round-031 Decisions 1–6: correct the advertised tool schemas and add the recurrence gate; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the schema fix authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked — the round changes no persisted or in-memory state; the existing persisted shapes (`history.jsonl`, `tokens.log`, the prompt logs) are unchanged. | `data-model-covers-all-state` holds — no new state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**` | Checked — no new or changed CLI interface truth: the tool-schema well-formedness is asserted by the **production-assembler (`agentTools()`) gate**, and the offered-tool-set Gherkin/DSL rows (`offering-the-agent-tools.feature` + `chat/dsl.md`) are unchanged. | `acceptance-coverage` / `dsl-exact-one-match` hold — nothing to add or change (round-020 precedent). **Recorded divergence (ARCH-5)**: the round's acceptance (US1/US2) is carried by the production-assembler (`agentTools()`) gate + a manual live Vertex/Gemini check, **not** by executable Gherkin (deliberate — the schemas are not observable through the built binary hermetically). |
