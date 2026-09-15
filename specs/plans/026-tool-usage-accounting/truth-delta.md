# Truth Delta: 026-tool-usage-accounting

**Plan Package**: `specs/plans/026-tool-usage-accounting`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added the **Tool-usage accounting** row (global append-only JSONL `~/.tellme/tools-count.jsonl`, one record per executed invocation `{timestamp, tool, outcome}`, `ok`/`error`/`timeout`, best-effort, never reset by `--new`, via an injected `ToolUsageSink`; the `--tool-usage` offline report) + added `--tool-usage` to the **CLI flag parsing** row + noted the round-026 **counting seam** in the **Agent tool loop** row; **Testing & Verification**: added the round-026 assertions to the **E2E runner**, **Test strategy**, **Local fake provider**, and **Pure-helper unit tests** rows; added the unbounded-log forward item to **Not Introduced Yet**. | Round-026 decisions 1–6: the three-way outcome, the global log, the injected counting seam, and the offline report. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the tool-usage accounting authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/data/data-model.dbml` | New table `tool_usage_record` — one JSON object per line `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"<ok|error|timeout>"}` appended to the **user-global** `~/.tellme/tools-count.jsonl` (resolved via `os.UserHomeDir()`; deliberately OUTSIDE `$TELL_ME_HOME`). One record per EXECUTED invocation; `outcome` ∈ `ok` (incl. a bounded/truncated result) / `error` (non-nil tool error) / `timeout` (nil-error result at the per-call deadline). Append-only (`O_APPEND|O_CREATE`), best-effort, and **never reset by `--new`**. Added the `tool_usage_outcome` enum and extended the Project Note. | Round-026 Q2 (global append-only log) + research Decisions 1–2. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/accounting-for-the-tool-use.feature` | New interface feature (3 Rules): each executed tool call is recorded with its outcome (`ok`/`error`/`timeout`); the record is kept across sessions (survives `--new`); the operator reviews the per-tool roll-up offline. Carries both round-026 acceptance journeys. | FR-001–FR-011; the outcome classification + the global log + the `--tool-usage` report. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added 3 Givens (the tool-usage log arrange + two new fake-provider tool behaviours), 1 When (the `--tool-usage` review), and 5 Thens (the log outcome counts + "no tool used"; the report outcome counts + "never used" + "every tool with no uses"), plus the round-026 note. | The record + the report need executable rows; no existing row was retired or moved. |
| NOOP | `specs/truth/features/cli/dsl.md` | Checked — the round-026 rows are used only by the `chat` module (the interface-root vocabulary and the class-phrase vocabulary are unchanged); the report reuses the existing root `tellme exits successfully` row. | `dsl-single-authority` — the root keeps only cross-module rows. |
