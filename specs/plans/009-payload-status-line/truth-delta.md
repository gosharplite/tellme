# Truth Delta: 009-payload-status-line

**Plan Package**: `specs/plans/009-payload-status-line`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added `Token estimator` (stdlib heuristic) + `Payload status line` (per-turn `stderr` estimate + actual, always-on, no `tellme:` prefix, injectable clock seam, observe-only). **Configuration**: added `Payload budget` (`MAX_HISTORY_TOKENS` env/config, default **1000000**, `>= 0`; negative → configuration error). **Reasoning & Provider Transport**: widened the `Provider gateway port` `Response` with the provider's reported `usage`, and `Response normalization` to extract the `usage` block. **Testing & Verification**: extended the `E2E runner`, `Local fake provider` (report/withhold `usage`), and `Pure-helper unit tests` (estimator, budget resolver, status formatting). **Not Introduced Yet**: the reference metrics line (`M:`/`H:`/`C:`/`$cost`) stays out of scope; the budget is displayed, not enforced. | Round-009 research Decisions 1–8 (Clarify Q1–Q3; user-locked default). |

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
