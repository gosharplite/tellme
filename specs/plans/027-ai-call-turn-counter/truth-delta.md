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
| MODIFY | `specs/truth/features/cli/chat/presenting-the-turn.feature` | Reshaped the header-counting Rule to **`The turn header counts the model requests made so far`**: the two-plain-exchange Example still reads `Turn 3`, and a **new** Example proves a **single** earlier turn that consulted a tool also reads `Turn 3` (its two inference rounds) — not `Turn 2`. A **third** Example proves a **fresh session after an earlier tool-using turn** reads `Turn 1` (the `--new` reset; PR #58 review fold, carrying `starting-the-count-over-on-a-fresh-session.feature`). | `acceptance-coverage` for the round-027 counting semantics (a tool-using turn advances the number by its call count). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | The `the turn is headed "Turn {number}" for the active mode` Then row: `{number}` is now the session's **AI-endpoint-call count** + 1 (Σ of the active `history_entry` `calls` + 1; a line without `calls` counts as 1; a tool-using history advances by its inference-round count, not by one). The two arranged tool-using history Givens (`… carrying the provider token "{token}"` / `… with no provider token`) now write `calls: 2`. Added the round-027 module note. No row retired or moved; the interface-root `the session history already holds the exchanges:` row is unchanged (its plain lines count as 1). | The header definition and the arranged-history Givens must match the new unit; `dsl-single-authority` holds (no duplicate row). |
| MODIFY | `specs/truth/features/cli/history/dsl.md` | The `the session history already holds a tool-using exchange` Given now writes `calls: 2` (the persisted shape records a tool-using turn's two inference rounds). | Keeps the arranged tool-using history consistent with the round-027 persisted shape. |
| MODIFY | `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature` + `specs/truth/features/cli/chat/dsl.md` | **`--new` on the `-i` surface (review/operator fold):** a new When row `the operator starts a fresh session at the interactive prompt and submits the prompt "{prompt}"` (runs `tellme --new -i`) and an Example asserting `Turn 1` after an arranged tool-using history. The product now archives `--new` **before** the interactive read for **both** terminal readers (the `-i` TUI path previously dropped `--new`, so the header counted the prior history). | Carries the round-027 `--new`-reset contract on the `-i` submit surface. |
