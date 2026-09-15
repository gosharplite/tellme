# Truth Delta: 024-tool-resource-contract-and-execute-command

**Plan Package**: `specs/plans/024-tool-resource-contract-and-execute-command`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added a **Tool resource contract (bounds)** row (`max_output_tokens` + `timeout`; default = resolved budget ÷ 4, ceiling = budget; shell timeout 300 s / readers 30 s; ceiling 7200 s; **loop-enforced**) and an **Agent command tool (`execute_command`)** row (bash-first `bash -c`, no `pipe_commands`, no security, non-zero exit = success result, `output_file`/`append`). Reworked the **Read-only filesystem tools** row to retire the fixed 100000-byte / 1 MiB caps in favour of the parameterized aggregate bound (whole-file reads + skip marker; ≤50 kept). Noted the loop as the single contract enforcement point; extended the pure-helper unit-tests row; updated Not-Introduced-Yet (shell tool now introduced; write tools still deferred). | Round-024 research Decisions 1–8. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — round 024)_ | `specs/truth/contracts/**` | Expected **NOOP** — tellme has a single CLI end and no OpenAPI/HTTP surface; the tool contract authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — round 024)_ | `specs/truth/data/data-model.dbml` | Expected **NOOP** — no new persisted state; the tool-step record already carries `{tool, arguments, result[, signature]}` and accommodates `execute_command`. | No record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — round 024)_ | `specs/truth/features/cli/chat/**` | Expected **ADD** an `execute_command` interface feature; **MODIFY** the reader features (`reading-several-files`, `listing-a-directory`, `surveying-a-folder-tree`) + `chat/dsl.md` to carry the token-bound/timeout params and the new truncation/skip markers. | FR-001..FR-017. |
