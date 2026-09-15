# Truth Delta: 028-user-global-prompt-log

**Plan Package**: `specs/plans/028-user-global-prompt-log`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/techstack.md` | TBD — expected MODIFY the *Shared global prompt log* row (location → `~/.tellme/global_prompts.jsonl`) + the `tell-me-go` divergence. | Round-028 relocation of the `-i` shared prompt log to a user-global file. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/contracts/**` | TBD — expected NOOP (no API/HTTP surface in tellme). | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/data/data-model.dbml` | TBD — expected MODIFY the `prompt_log_entry` location (`$TELL_ME_HOME/output/global_prompts.jsonl` → `~/.tellme/global_prompts.jsonl`) + the seed-on-absent migration note. | Round-028: the log's durable home changes from environment-scoped to user-global. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/features/cli/chat/**` | TBD — expected MODIFY the DSL/feature rows that name `$TELL_ME_HOME/output/global_prompts.jsonl` (write/read targets → `~/.tellme/global_prompts.jsonl`) + a seed Example. | `acceptance-coverage` for the round-028 relocation + migration. |
