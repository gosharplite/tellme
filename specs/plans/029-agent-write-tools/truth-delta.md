# Truth Delta: 029-agent-write-tools

**Plan Package**: `specs/plans/029-agent-write-tools`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added a **Write filesystem tools** row to the CLI-application agent-tool surface — the `write_file` **create-only + atomic** contract and the `replace_text` **strict-unique** contract, no security/consent and no backup/undo, with the deferred/non-goal split (`append_text` / `undo_file_change` deferred; `delete_path` / `create_directory` non-goals). Rewrote the *Not Introduced Yet* write-tools bullet so `write_file` / `replace_text` are no longer listed as deferred. | Round-029 D1–D8: the write pair becomes part of tellme's current tool surface; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/contracts/**` | — | Filled by `/axb-api-plan` (expected NOOP: tellme has a single CLI end, no API surface; `contract-authoritative` holds vacuously). |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/**` | — | Filled by `/axb-data-plan` (expected NOOP: the two write tools persist no state; the persisted-file shapes are unchanged). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/**` | — | Filled by `/axb-dsl-refine` (expected ADD: a write-tools interface feature + `dsl.md` rows for the create-only / atomic / strict-unique steps; `acceptance-coverage` for the round-029 journeys). |
