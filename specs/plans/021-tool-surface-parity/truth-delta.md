# Truth Delta: 021-tool-surface-parity

**Plan Package**: `specs/plans/021-tool-surface-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: reshaped `Read-only filesystem tools` to **three** tools (`list_files` — `Contents of <path>:` + `[d]/[f]` lines, optional `path` default `.`; `read_files` — multi-file `filepaths: string[]`, `--- File: <path> ---` framing, 100000-byte per-file cap + `... (truncated)`, binary/directory/≤50 handling, **1 MiB aggregate result cap**; `get_tree` — connector tree, default `max_depth` 2, `.git` not recursed, 1 MiB cap) all **requiring `reason`**; **removed** the `Session-summarisation tool` row; **Agent tool loop** row notes the `reason` echo into the `[tool] …` `stderr` line; **Testing & Verification** (pure-helper unit tests) extended for the three tool contracts + the `reason` echo; **Not Introduced Yet** records history summarisation as removed in round 021. | Round-021 research Decisions 1–8 + D3a; operator-locked D1–D5; PR #48 review B2/TD1. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the agent tool surface authors no request/response contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked — the persisted tool-step shape (`{tool, arguments, result[, signature]}`) is unchanged; only the `read_files` argument **content** now carries `filepaths[]` (stored as an opaque string, as today). | FR-015; no record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/{reading-several-files,listing-a-directory,surveying-a-folder-tree,offering-the-reader-tools}.feature` | Four new `chat` interface features carrying US1–US4 — multi-file `read_files` (one request, several files; truncation/binary/too-many handled inside the result), `list_files` `Contents of …` + `[d]/[f]` shape (default `path`), `get_tree` connector tree, and the exactly-three-reader-tools offer (a removed-tool request fails under the existing unknown-tool contract). | FR-002–FR-011; acceptance US1–US4 (D1/D3/D5). |
| MODIFY | `specs/truth/features/cli/chat/{watching-the-tool-loop,sending-the-configured-persona,surveying-a-folder-tree,reading-several-files}.feature` + `chat/dsl.md` | Added the reason-echo Rule; **deepened the tree fixture** (`src/pkg/deep/leaf.go` + `.git/config`) with the negative Thens `the tree does not show "…"` / `the tree does not descend into "…"` (B1); added the directory Example + `tellme reports that "…" is a directory` (TD2) and the framing Thens `the read result frames "…"` (R1); restored the prior-history Given (R3); `chat/dsl.md` **+25 rows** (10 Given + 15 Then) and **−3 rows**. | FR-001/FR-005/FR-006/FR-007/FR-009/FR-012 (D1–D5); PR #48 review B1/TD2/R1/R3. |
| DELETE | `specs/truth/features/cli/chat/summarising-the-conversation.feature` + its 3 `chat/dsl.md` rows | Removed the summarisation interface feature and every step pattern that referenced the retired tool. | FR-010/FR-011 (remove `summarize_history`, D5/D6). |
