# Truth Delta: 028-user-global-prompt-log

**Plan Package**: `specs/plans/028-user-global-prompt-log`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Shared global prompt log** row: relocated from `$TELL_ME_HOME/output/global_prompts.jsonl` to the per-user `~/.tellme/global_prompts.jsonl` (resolved via `os.UserHomeDir()`), read + written only there; added the seed-on-absent migration (a verbatim copy of the environment-scoped file — copy, not move; never overwrites an existing destination; a missing source starts empty) and the unresolvable/unwritable-home no-op; recorded the `tell-me-go` divergence (sharing unit changes from *per-`TELL_ME_HOME`* to *per-user*; tellme skips tell-me-go's legacy locations). | Round-028 Decisions 1–6: operator-directed relocation to a user-global file (reusing the round-026 `~/.tellme/` root) + the first-use seed + the recorded sharing divergence. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/contracts/**` | TBD — expected NOOP (no API/HTTP surface in tellme). | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/data/data-model.dbml` | TBD — expected MODIFY the `prompt_log_entry` location (`$TELL_ME_HOME/output/global_prompts.jsonl` → `~/.tellme/global_prompts.jsonl`) + the seed-on-absent migration note + the tell-me-go divergence. | Round-028: the log's durable home changes from environment-scoped to user-global. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/features/cli/chat/**` | TBD — expected MODIFY the DSL/feature rows that name `$TELL_ME_HOME/output/global_prompts.jsonl` (write/read targets → `~/.tellme/global_prompts.jsonl`) + a seed Example/row + the module note. | `acceptance-coverage` for the round-028 relocation + migration. |
