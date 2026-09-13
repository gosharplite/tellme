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
| MODIFY | `specs/truth/features/cli/chat/reporting-the-payload-status.feature` | Added the ordering Thens `the estimated payload status is reported before the answer` (Rule 1 and the no-usage variant) and `the measured payload status is reported after the answer` (Rule 2). | Round 010 — pin the payload-status **ordering** (round 009 pinned only presence on `stderr`). |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` | Added the ordering Then `the tool activity is reported before the answer`. | Round 010 — pin the tool-loop log ordering (the same `stderr`-diagnostic class). |
| ADD | `specs/truth/features/cli/chat/dsl.md` | Added three Then rows — `the estimated payload status is reported before the answer` / `the measured payload status is reported after the answer` / `the tool activity is reported before the answer` — plus the round-010 ordering-rows note (asserted against the merged capture). | Round 010 — the ordering facet dropped from the executable truth in round 009 (the `acceptance-coverage` leak). |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the ordering rows are `chat`-module-specific; the frozen class-phrase vocabulary stays **10**. | No cross-module row was needed. |
