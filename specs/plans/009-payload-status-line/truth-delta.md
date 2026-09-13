# Truth Delta: 009-payload-status-line

**Plan Package**: `specs/plans/009-payload-status-line`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/techstack.md` | Expected MODIFY — a token-counting capability (heuristic, stdlib-only) + the `MAX_HISTORY_TOKENS` payload budget, and (if actual tokens are in scope) provider `usage` parsing. | Round-009 payload status line. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/contracts/**` | Expected `NOOP` — tellme has a single CLI end and no OpenAPI surface of its own. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/data/data-model.dbml` | Expected `NOOP` — per-turn payload token counts are computed for display, not persisted (recomputed from history text on resume); pending clarify. | Round-009 payload status line. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/features/cli/**` | Expected ADD — a `status` feature (or an extension of `chat/`) carrying the per-turn payload status line plus its DSL rows. | Round-009 acceptance. |
