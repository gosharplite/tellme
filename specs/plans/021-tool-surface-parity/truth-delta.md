# Truth Delta: 021-tool-surface-parity

**Plan Package**: `specs/plans/021-tool-surface-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: reshaped `Read-only filesystem tools` to **three** tools (`list_files` — `Contents of <path>:` + `[d]/[f]` lines, optional `path` default `.`; `read_files` — multi-file `filepaths: string[]`, `--- File: <path> ---` framing, 100000-byte cap + `... (truncated)`, binary/directory/≤50 handling; `get_tree` — connector tree, default `max_depth` 2, `.git` not recursed) all **requiring `reason`**; **removed** the `Session-summarisation tool` row; **Agent tool loop** row notes the `reason` echo into the `[tool] …` `stderr` line; **Testing & Verification** (pure-helper unit tests) extended for the three tool contracts + the `reason` echo; **Not Introduced Yet** records history summarisation as removed in round 021. | Round-021 research Decisions 1–8; operator-locked D1–D5. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/contracts/**` | Expected `NOOP` — tellme has a single CLI end and no OpenAPI surface of its own. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/data/data-model.dbml` | Expected `NOOP` — the persisted tool step shape (`{tool, arguments, result[, signature]}`) is unchanged; only the `read_files` argument content now carries `filepaths[]`. | FR-015; no record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/using-a-tool.feature` + `chat/dsl.md` | Re-point the `read_files` Given/Then rows to the multi-file `filepaths[]` signature; add `reason` and the reshaped-output assertions. | FR-001–FR-007. |
| ADD | `specs/truth/features/cli/chat/**` | New `get_tree` interface feature + module DSL rows (connector tree; `max_depth` default 2; `.git` not recursed) and the `list_files` reference-format rows. | FR-007–FR-009 (D5). |
| DELETE | `specs/truth/features/cli/chat/summarising-the-conversation.feature` + its `chat/dsl.md` rows | Remove the summarisation interface feature and every step pattern that referenced the retired tool. | FR-010/FR-011 (remove `summarize_history`). |
