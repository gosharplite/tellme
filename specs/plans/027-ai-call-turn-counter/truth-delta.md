# Truth Delta: 027-ai-call-turn-counter

**Plan Package**: `specs/plans/027-ai-call-turn-counter`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Turn chrome (operator)** row: redefined `<N>` as the session's **AI-endpoint-call count + 1** (the running sum of the prior turns' inference rounds) instead of the completed-turn count, and appended the round-027 clause (call-based unit; cadence/format unchanged — one header + one `╰─⠿ Ready` per prompt, not per call; a tool-less turn advances by one, a tool-using turn by its inference-round count; an internal retry does not count; `--new` restarts at `Turn 1`). **Session history store** row: added the integer **`calls`** field (the turn's inference-round count, summed to derive the header number; a legacy line without it counts as 1). | Round-027 Decision 1 (unit = inference round), Decision 2 (persist the per-turn count), Decision 5 (`--new` resets; legacy floor), Decision 6 (surface unchanged). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the turn counter authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/data/data-model.dbml` | `history_entry` gains an integer **`calls`** attribute (always ≥ 1): the turn's AI-endpoint-call count (provider inference rounds; 1 for a tool-less turn, 1 + tool rounds otherwise; an internal retry is not counted). It is summed across the active session (`Σ calls`), and the round-017 turn header shows `Σ calls + 1`; a legacy line without the field counts as 1; `--new` archives the active file, so the sum restarts. Updated the `history_entry` Note, the record shape to `{prompt, answer, calls, steps:[…]}`, and the Project Note. | Round-027 Decision 2 (persist the per-turn count) + Decision 5 (`--new` reset; legacy floor); the session-history record is the durable home for the count. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/features/cli/**` | (to be filled by `/axb-dsl-refine`) | Skeleton initialized by `/axb-specify`. |
