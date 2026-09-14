# Truth Delta: 021-tool-surface-parity

**Plan Package**: `specs/plans/021-tool-surface-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _TBD_ | `specs/truth/techstack.md` | Expected `MODIFY` — desktop tool contracts for `list_files`/`read_files`/`get_tree` (multi-file `filepaths[]`, required `reason`, 100 KB cap, ≤50 files, default `max_depth` 2) and removal of the summarisation tool row. | Round-021 aligned tool surface (D1–D5). |

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
