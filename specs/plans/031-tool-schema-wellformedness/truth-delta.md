# Truth Delta: 031-tool-schema-wellformedness

**Plan Package**: `specs/plans/031-tool-schema-wellformedness`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/techstack.md` | _Placeholder — the tool-schema / agent-tool rows to be updated by `/axb-technical-research` during the round._ | _Round-031 scope: correct the advertised tool schemas to be well-formed (`required ⊆ properties`) and record the well-formedness gate._ |

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
| _pending_ | `specs/truth/features/cli/**` | _Placeholder — expected NOOP (no new business journey), unless a schema-well-formedness interface rule is added._ | _To be confirmed by `/axb-dsl-refine` during the round._ |
