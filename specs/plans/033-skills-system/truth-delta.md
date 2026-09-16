# Truth Delta: 033-skills-system

**Plan Package**: `specs/plans/033-skills-system`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` -> `### Skills` (new category) | Added a **Skills** category (CLI Application): a **Skills catalog (load)** row (single source `<TELL_ME_HOME>/docs/skills/`, recursive `name`/`description` frontmatter parse, non-skill Markdown skipped, duplicate first-wins, missing dir → empty, best-effort, not persisted) and a **`list_skills`** read-only agent tool row (name + description + location, deterministic path order, honours the round-024 resource contract + the round-031 `reason` schema). | Round-033 Q1 → on-demand only + Q3 → a `list_skills` tool; the truth must reflect the new capability (`truth-current`). |
| MODIFY | `specs/truth/techstack.md` -> `### CLI Application` (Agent tool schemas row + Agent tool-schema gate row) + `### Testing & Verification` (Tool-usage accounting row) | The shared schema builder now also backs `list_skills`; the round-031 `required ⊆ properties` gate covers **seven** tools (was six); the round-026 `--tool-usage` report's live-registry enumeration is corrected to the current set (`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`, `execute_command`, `list_skills`) — repairing a stale four-tool list. | `truth-current`: the tool surface grew; the truth describes the shipped set. |
| MODIFY | `specs/truth/techstack.md` -> `### Not Introduced Yet` | Added an explicit **Skills.sh ecosystem & automatic skill injection** non-goal bullet (the reference's `.skills/` source, the `skillssh` tools, cross-source merging, and relevance-based injection). | `truth-current` + recorded divergence: the reference ships these; tellme deliberately does not (operator scope, Q1/Q3). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/contracts/**` | (to be completed by `/axb-api-plan`) | — |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/data/**` | (to be completed by `/axb-data-plan`) | — |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| TBD | `specs/truth/features/cli/**` | (to be completed by `/axb-dsl-refine`) | — |
