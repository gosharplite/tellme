# Truth Delta: 031-tool-schema-wellformedness

**Plan Package**: `specs/plans/031-tool-schema-wellformedness`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added an **Agent tool schemas** row (CLI Application): the shared schema builder declares the resource params **and the mandatory `reason` property**, so every tool it backs satisfies `required ⊆ properties`; `execute_command` builds inline and already complies; records the defect's regression origin (round-024 `cfa005c`, carried by round-029 `eb0367c`). Added an **Agent tool-schema gate** row (Testing & Verification): a hermetic unit check over the production registry asserting `required ⊆ properties` + schema-parses for **every** registered tool. Added a **Not Introduced Yet** bullet recording the deferred stronger guards (fake schema-validation; a live pipeline leg; a structural builder). | Round-031 Decisions 1–6: correct the advertised tool schemas and add the recurrence gate; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | _Placeholder — expected NOOP (single CLI end; no OpenAPI/HTTP surface)._ | _Expected `contract-authoritative` holds vacuously._ |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/**` | _Placeholder — expected NOOP (no persisted state changes)._ | _Expected `data-model-covers-all-state` holds — no new state._ |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/**` | _Placeholder — expected NOOP (no new business journey; the verification is a registry unit gate — round-020 precedent)._ | _To be confirmed by `/axb-dsl-refine` during the round._ |
