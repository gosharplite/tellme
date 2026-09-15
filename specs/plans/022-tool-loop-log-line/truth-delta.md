# Truth Delta: 022-tool-loop-log-line

**Plan Package**: `specs/plans/022-tool-loop-log-line`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Agent tool loop** row: reshaped the `stderr` tool-loop log to a single timestamped line `[HH:MM:SS] [Tool] <tool name> - <reason>` (no-reason form `[HH:MM:SS] [Tool] <tool name>`), dropping the raw `arguments`/`result` segments, plus **one blank line** before the answer of a tool-using turn; the timestamp comes from the injected clock seam (an **operator-chosen** shape vs. the reference's decomposed `[Tool …]` lines). **Pure-helper unit tests** row: adds the round-022 tool-loop log-line formatter. | Round-022 research Decisions 1–8 (operator-locked Q1–Q3 in `spec.md`); a pure diagnostic-rendering change — no new dependency. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the tool-loop log line authors no request/response contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked — the persisted tool-step record (`{tool, arguments, result[, signature]}`) and the `history.jsonl` shape are unchanged; the tool-loop log line is operator-facing output, not stored. | FR-008/FR-010; no record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` | Reshaped the header (acceptance pointer → `reporting-each-tool-use` + `separating-the-tools-from-the-answer`); the Rules now cover: the tool call, its reason, a call with **no reason**, one line per call, **tools in order**, the **positive separation** (positional + `-i`), and the **negative separation** (used-no-tool, on `-i`). | FR-001–FR-007; PR #50 review B1/TD1/TD3. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | MODIFY the `the run reported the reason …` + `the run reported the tool call …` rows to the `[HH:MM:SS] [Tool] …` shapes (no `arguments=`/`result=`); ADD the no-reason, one-line-per-call, positive-separation, **ordered** (`…in order "…" and "…"`), and **negative** (`the tool loop added no blank line before the answer`) Then rows + the **sequential two-tool Given** (`…shows the folder tree and then reads "…" and then answers with "…"`); expanded the round-022 module note (ungated witness, negative isolation from the frame gap, the TD2 schema-nonconforming note, the order row). | FR-001–FR-007; research Decisions 1–6 + PR #50 review B1/TD1/TD2/TD3. |
