# Truth Delta: 024-tool-resource-contract-and-execute-command

**Plan Package**: `specs/plans/024-tool-resource-contract-and-execute-command`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — round 024)_ | `specs/truth/techstack.md` | Expected **MODIFY**: the agent tool surface description gains the tool resource contract (`max_output_tokens` + `timeout`, default/param/ceiling) and `execute_command` (bash-first); the reader rows move from the fixed 1 MiB / 100000 caps to the params. | Round-024 research decisions (D5–D7). |

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
