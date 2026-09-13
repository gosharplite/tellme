# Truth Delta: 010-stream-ordering-observability

**Plan Package**: `specs/plans/010-stream-ordering-observability`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner` row with a **merged-stream capture variant** (both streams into one shared ordered buffer — the `2>&1` equivalent) and added a **Cross-stream ordering witness** row; extended the `Test strategy` and `Pure-helper unit tests` rows with the layered (unit + E2E) ordering assertions. Nothing else changed — one CLI end, no new dependency. | Round-010 research Decisions 1–8: the E2E harness could not represent an inter-stream interleave (round 009's missing oracle); ordering becomes an assertable contract at unit + E2E. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; the outbound provider request/response shape is unchanged this round. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. No persisted state changes this round — per-turn token counts stay display-only and are recomputed; the session-history model (`history_entry`) is untouched. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/features/cli/**` | _Pending — round 010._ | _Placeholder; `/axb-dsl-refine` replaces it (expected MODIFY — ordering semantics in the chat DSL + feature)._ |
