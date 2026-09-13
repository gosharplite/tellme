# Truth Delta: 008-agent-tools-and-tool-call-loop

**Plan Package**: `specs/plans/008-agent-tools-and-tool-call-loop`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added `Read-only filesystem tools` (`list_files`/`read_files`, wire-valid snake_case, 1 MiB read cap, no path boundary) + `Session-summarisation tool` (`summarize_history`, LLM-backed) + `Agent tool loop` (`AgentLoop` orchestrator; `MAX_TOOL_LOOP` default 1000, per-tool timeout, non-streaming, logs to `stderr`); extended `Session history store` (widened line `{prompt,answer,steps:[…]}`, replay on resume) and `Session lifecycle flags` (`-l` prompt/answer only). **Reasoning & Provider Transport**: widened `Provider gateway port` (`Request` tool definitions; `Response` tool-call requests), `Request assembly` (`tools` + `tool_calls` + `tool` messages), `Response normalization` (`tool_calls`). **Testing & Verification**: extended the `E2E runner` + `Local fake provider` (serve/record tool calls) and `Pure-helper unit tests`. **Not Introduced Yet**: summarisation now introduced as an on-demand agent tool; minimal tool-call shape noted; pruning/write-tools/concurrency listed. | Round-008 research Decisions 1–9; Clarify Q1/Q2, R2/Q1, extra Q2. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/contracts/**` | Expected `NOOP` — tellme has a single CLI end and no OpenAPI surface of its own. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/data/data-model.dbml` | Widened the persisted turn to carry its tool activity: `history_entry` now embeds an ordered steps array, modelled as a new `history_step` table (`location, position, step` → `history_entry.(location, position)`) with `tool`/`arguments`/`result`. Append-only lifecycle and byte-determinism preserved; `-l` still reads only `prompt`/`answer`. | Round-008 research Decision 5; Clarify Q2 (widen the persisted turn). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/**` | New `chat` features `using-a-tool`, `watching-the-tool-loop`, `failing-the-tool-loop`, `summarising-the-conversation` + module DSL rows (working-dir file fixture; tool-loop provider fakes; read-tool call/result; tool-loop log; loop bound; `the tool request failed` / exit-7; summarise + unchanged-records). | Round-008 tool loop; acceptance `answering-with-a-declared-tool` / `watching-the-tool-loop` / `bounding-and-failing-the-tool-loop` / `summarising-the-conversation`. |
| MODIFY | `specs/truth/features/cli/dsl.md` | Interface root: frozen class-phrase vocabulary 10→11 — added `the tool request failed`. | Round-008 failure contract (exit 7). |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` | Added Rule *Listing after a tool-using turn shows only the operator's messages* + the `the session history already holds a tool-using exchange` Given and the `tellme lists only the operator's messages` Then. | Round-008 FR-017 (`-l` unchanged under the widened record). |
