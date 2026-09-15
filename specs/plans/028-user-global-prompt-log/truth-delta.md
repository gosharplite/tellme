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
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the prompt-log relocation authors no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/data/data-model.dbml` | The `prompt_log_entry` table's durable home changes from the environment-scoped `$TELL_ME_HOME/output/global_prompts.jsonl` to the **user-global** `~/.tellme/global_prompts.jsonl` (resolved via `os.UserHomeDir()`; the same `~/.tellme/` root as the round-026 `tool_usage_record`). Added the **seed-on-absent** lifecycle — when the file is absent it is populated by a verbatim copy of the environment-scoped file (copy, not move; never overwrites an existing destination; a missing source starts empty). Recorded the `tell-me-go` divergence. Updated the table Note and the Project Note (the log's location + seed). The record shape, the append-only write, the newest-first-deduped read, and the compaction policy are **unchanged**. | Round-028 Decisions 1/3/6: the log's durable home moves to the user-global root and gains a first-use seed; the data truth must reflect the current location + lifecycle (`data-model-covers-all-state`, `truth-current`). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/carrying-over-the-environment-prompt-log.feature` | New interface feature (chat module) with 3 atomic Rules — (1) on first use the environment prompt log is carried into the absent shared prompt log; (2) an existing shared prompt log is not overwritten; (3) with no environment prompt log the shared log starts empty. | `acceptance-coverage` for the round-028 acceptance journey `carrying-over-the-existing-prompt-history.feature` (the seed). |
| MODIFY | `specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature` | Header comment: the record target is now the per-user `~/.tellme/global_prompts.jsonl` (round 028 moved it out of `$TELL_ME_HOME/output/`), and the acceptance-journey reference points at `sharing-prompts-across-environments.feature`. The Rules/Examples are unchanged (they assert the "shared prompt log" abstractly). | The interface truth must name the current target and the round-028 acceptance journey. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | The three shared-log rows (`the shared prompt log already holds` Given · `records the prompt` / `still holds only` Thens) retarget their `工作區` from `$TELL_ME_HOME/output/global_prompts.jsonl` to `~/.tellme/global_prompts.jsonl`. **Added** 4 rows: the Given `the environment prompt log already holds "{prompt}"` (the seed source at `$TELL_ME_HOME/output/global_prompts.jsonl`), the Given `the shared prompt log has not been created yet` (ensures the user-global log is absent), the Then `the shared prompt log also holds the earlier prompt "{prompt}"` (the carry-over witness), and the Then `the shared prompt log does not hold "{prompt}"` (the no-overwrite witness). Updated the round-015 module note's path reference and **added** the round-028 module note (the user-global relocation + seed + divergence). No row retired or moved; `dsl-single-authority` holds. | The read/write target and the arranged given must match the relocated user-global log, and the seed must be executable; the module note records the relocation. |

**Verification (this half)**: Gherkin/DSL topology audit — `audit_feature_dsl_topology.py --root specs/truth/features/cli` **PASSED** (40 features · 6 modules · 16 root + 250 module rows · 1312 steps). All remaining references to `$TELL_ME_HOME/output/global_prompts.jsonl` in `specs/truth/**` are the **seed source** (never a read/write target): the `techstack.md` row, the `chat/dsl.md` round-028 note + the `the environment prompt log already holds` row, the new feature's header, and the `data-model.dbml` notes.
