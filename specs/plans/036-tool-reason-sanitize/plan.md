# Plan — round 036 (tool reason sanitize)

Plan package: `specs/plans/036-tool-reason-sanitize`

## Theme

Make the model-authored tool `reason` obey the tool log's one-line contract: `FormatToolReason` must **fold** (`\n`/`\r` → space) + **trim** the reason and **cap** it at a named rune bound (`reasonValueCap = 200`) like its siblings, and a **blank** reason must emit **no** reason line. A presentation-contract correction — a rendering-only, `stderr`-diagnostic change that restores round-022 B1's fold guarantee (dropped unrecorded by round 034); issue [#74](https://github.com/gosharplite/tellme/issues/74) (`spec.md` FR-001–008).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (terminal diagnostics)** — the `chat` module's tool-call log (`[Tool Reason]` line) | `cli` | The subject of the round: the rendering of the per-call `[Tool Reason]` line on `stderr` (both emission surfaces — the loop action line and the per-call tail — via the one pure formatter). `stdout` stays byte-exact; no flag, exit code, cadence, schema, transport, or persisted-record change. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change**.

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` MODIFIES the `chat/dsl.md` **reason row** prose (state the single-line guarantee explicitly + land the `reasonValueCap` constant) and adds a round-036 note. **No new `DSLRow`, no feature Example** (the operator's Q1 narrowing).

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract — the fix changes only a `stderr` diagnostic rendering (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `output/<mode>/tokens.log` cadence: no persisted-state change — the fix alters only how the reason is **rendered** to the terminal (`spec.md` FR-007; `research.md` D7).
- **`/axb-ui-plan` → skipped.** A line-oriented CLI diagnostic: no web/HTML prototype and no TUI screen change; the terminal interaction lives directly in the CLI interface Gherkin (the standing CLI streamlining). (`/axb-spec-by-example` was NOOP; the one-line promise is carried at the contract level.)

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the reason axis (`spec.md` FR-001/FR-002/FR-005; `research.md` D2/D4):

- `chat/dsl.md` — **MODIFY the reason row** (`the run reported the reason "{reason}" for the tool call "{tool}"`): state the **single-line guarantee** explicitly (as the sibling result row already does — fold `\n`/`\r`, trim, one line) and record the `reasonValueCap` constant **200** next to the siblings' documented `189`/`200`; the blank-reason behaviour (no line) is noted. Also add a **round-036 note** recording the fold+trim+cap + blank suppression and that the guarantee restores round-022 **B1** (ADR 0006).
- **No new `DSLRow`; no feature Example** — the witness is a hostile-fixture **unit pin** (operator Q1); `chat/watching-the-tool-loop.feature` is unchanged.

`truth-delta.md` carries the explicit MODIFY row; `specs/truth/techstack.md` is MODIFY-owned by `/axb-technical-research` (already folded).
