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
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` | Added three Rules — a call that states no reason (`[Tool] <name>`), each tool call on its own line, and the answer set apart from the tool report — carrying the round-022 acceptance (US1/US2). | FR-003/FR-005/FR-006; acceptance `reporting-each-tool-use` / `separating-the-tools-from-the-answer`. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Reshaped the `the run reported the reason …` row to the `[HH:MM:SS] [Tool] <tool name> - <reason>` line and the `the run reported the tool call …` row to the `[HH:MM:SS] [Tool] <tool name>` shape (no `arguments=`/`result=`); added the no-reason, one-line-per-call, and separation rows; added the round-022 module note. | FR-001–FR-006; research Decisions 1–6. |
