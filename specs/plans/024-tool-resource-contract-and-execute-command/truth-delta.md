# Truth Delta: 024-tool-resource-contract-and-execute-command

**Plan Package**: `specs/plans/024-tool-resource-contract-and-execute-command`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added a **Tool resource contract (bounds)** row (`max_output_tokens` + `timeout`; the bound derives from the **effective budget** = `min(MAX_HISTORY_TOKENS, the model's configured context window)`, default `effectiveBudget ÷ 4`, ceiling `effectiveBudget ÷ 2`; shell timeout 300 s / readers 30 s; ceiling 7200 s; **loop-enforced**) and an **Agent command tool (`execute_command`)** row (bash-first `bash -c`, no `pipe_commands`, no security, non-zero exit = success result, `output_file`/`append` bound **directly to the file**, process-group termination, stdout binding). Reworked the **Read-only filesystem tools** row to retire the fixed 100000-byte / 1 MiB caps in favour of the parameterized aggregate bound (whole-file reads + skip marker; ≤50 kept). Extended the **Model pricing** row with an optional per-model **`CONTEXT_WINDOW`**. Noted the loop as the single contract enforcement point; extended the pure-helper unit-tests row; updated Not-Introduced-Yet. | Round-024 research Decisions 1–8 + review fixes B1/B2/D1/D2/D3. |
| MODIFY | `specs/truth/techstack.md` | **Grill fold (PR #54):** requalified the misleading "displayed, not enforced / displayed only" lines (`Payload budget`, `Not Introduced Yet → token-budget pruning`) so the scalar is *displayed **and** the cap on the effective budget that drives the tool-result bound* (conversation pruning stays a settled exclusion); softened "tracks the model" to `min(MAX_HISTORY_TOKENS, configured window)` — **model-tracking only when `CONTEXT_WINDOW` is set**; corrected "retiring the fixed 1 MiB cap" (default byte bound `1000000 B`, replacing — not identical to — `1048576 B`); recorded the contract-owned `bytesPerToken` byte realisation (raw-byte clamp, never the estimator); stated a timeout as a **nil-error result** and the capture as **bounded pipes**. | The shipped truth contradicted itself on the same scalar (grill Q3), and Q1/Q4/Q8 fixes had to reach the truth. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the tool contract authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked — no new persisted state; the tool-step record already carries `{tool, arguments, result[, signature]}` and accommodates `execute_command`. | No record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/running-a-shell-command.feature` | New interface feature — the `execute_command` tool: offered; runs; a non-zero exit is reported, not fatal; a command that exceeds its time is stopped; output captured to a file. | FR-001..FR-007. |
| ADD | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` | The offered tool set is now the four agent tools (readers + `execute_command`). | FR-013. |
| DELETE | `specs/truth/features/cli/chat/offering-the-reader-tools.feature` | Superseded by `offering-the-agent-tools.feature` (the surface now includes `execute_command`). | FR-013. |
| MODIFY | `specs/truth/features/cli/chat/reading-several-files.feature` | The read bound is the parameterized `max_output_tokens` (whole-file reads; the fixed 100000-byte cap retired); added a Rule for a request that cannot return every file (skip marker). | FR-008..FR-012. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | +4 Given (execute_command scripting) +5 Then (command ran / exit status / stopped / output target / file not read); the offering row → `the request offered exactly the agent tools`; the read-limit Given → `longer than the read bound`; a round-024 note. | FR-001..FR-017. |
| NOOP (shape) + ADD (bound witness) | `specs/truth/features/cli/chat/listing-a-directory.feature`, `surveying-a-folder-tree.feature` | **Grill Q5:** the two tools keep their **output shape** (NOOP), but the shared bound is a **tool-level, at-source** behavior (both `list_files` and `get_tree` `truncateToCap` at the source), so each gains an **ADD** Rule + Example witnessing the truncation marker (FR-011) via a small `max_output_tokens`. | The earlier "a loop concern, no new step" rationale was **false** (the bound is set by the tool, not the loop); FR-011 must be falsifiable. |
| MODIFY | `specs/truth/features/cli/chat/running-a-shell-command.feature` | **Grill Q6:** added a Rule + Example asserting the **process-tree** stop — a command that spawns a long-lived descendant is stopped with it and the descendant is gone afterwards (the shipped `sleep 60` fixture could not falsify `kill(-pgid)`/`WaitDelay`). | FR-003's "process tree" was unfalsifiable. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **Grill fold:** +3 Given (list/tree with a small result budget; a descendant-spawning command) and +3 Then (listing/tree trimmed; descendant no longer running); the read-bound Given restated **in bytes** (`effectiveBudget` bytes, default `1000000 B`); the stopped/trimmed Then rows distinguish the **time** stop (a nil-error result, not `error: `) from the **byte** trim; the round-024 note records the byte conversion, the pipe capture, and the new witnesses. | Q1/Q4/Q5/Q6/Q8 — the loop-clamp vs at-source units and the trim-vs-stop outcomes were under-specified, and FR-011/FR-003 had no witness. |
