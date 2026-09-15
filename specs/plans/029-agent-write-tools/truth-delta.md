# Truth Delta: 029-agent-write-tools

**Plan Package**: `specs/plans/029-agent-write-tools`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Added a **Write filesystem tools** row to the CLI-application agent-tool surface: the `write_file` **create-only + atomic** contract (**atomic create-only** via `os.Link`/`EEXIST`; created file mode **`0644`**), the `replace_text` **strict-unique + atomic** contract (temp + `rename`; a no-op replace short-circuits), and the **`replace_text` whole-file-read hazard**, a **missing `content`** key rejected (explicit `""` only), a `0755` parent-directory mode, and the tracked provider-truncation forward item (issue #62); no security/consent and no backup/undo; the deferred/non-goal split (`append_text` / `undo_file_change` deferred; `delete_path` / `create_directory` non-goals). Rewrote the *Not Introduced Yet* write-tools bullet so `write_file` / `replace_text` are no longer listed as deferred. | Round-029 D1–D8: the write pair becomes part of tellme's current tool surface; the truth must reflect the current system (`truth-current`). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the write tools author no request/response document. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked — the two write tools persist no state; the existing persisted shapes (`history.jsonl`, `tokens.log`, the prompt logs) are unchanged. | `data-model-covers-all-state` holds — no new state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/creating-and-editing-files.feature` | New interface feature (chat module) with 4 atomic Rules — (1) a new file is created with exactly the requested content (incl. a missing parent folder); (2) an existing file is never overwritten; (3) a uniquely identified block is replaced and nothing else changes; (4) an edit that cannot be uniquely placed is refused, leaving the file untouched (0 / >1 matches). | `acceptance-coverage` for the round-029 acceptance journeys `creating-a-file.feature` + `editing-a-file.feature`. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Updated the `the request offered exactly the agent tools` row to the **six**-tool set (`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`, `execute_command`). **Added** 6 Given rows (working-dir contains no file / no folder; a file whose lines are a table; a file containing a line twice; providers that create / edit a file before answering) and 10 Then rows (file created; folder created; content exactly / still exactly; lines are / are still; creation refused; edit refused — not present / not unique; file still holds a line twice), plus the round-029 module note. | The offered set and the new write-tool steps must be executable and uniquely owned (`dsl-exact-one-match`); the module note records the round. |
| MODIFY | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` | Header comment + Example title updated from the four-tool set (three readers + command) to the **six**-tool set (three readers + the write pair + command), so the file's prose matches its own `the request offered exactly the agent tools` step. | The offered set grew (round 029); the truth feature must name the current set (`truth-current`). The prose edit belongs in the **truth half** (`truth-single-owner` — `/axb-dsl-refine`), not the implementation phase. |
