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
| ADD | `specs/truth/features/cli/chat/listing-the-available-skills.feature` -> Rules (a prompt can list the workspace's skills · a skill's reference material is not listed as a skill · a workspace with no skills lists none · a listed skill can be opened on demand · skill content is not added to the request unless a skill is opened) | Added the executable skills interface truth: on a prompt-bearing turn the agent can list the workspace's skills (name + description) through a read-only `list_skills` tool; only frontmatter skill files are listed (a skill's `rules/*.md` is not); an empty catalog lists none; a listed skill is opened with the existing `read_files`; and **no** skill content is injected into the request. | Round-033 US1/US2 + FR-001…FR-006: the on-demand skills capability's observable CLI behaviour is executable truth (`acceptance-coverage`). |
| ADD | `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 033)` (5 rows) + `## Then (round 033)` (6 rows) | Added the module DSL rows for the new Given/Then sentences (arrange a skill / a skill with a reference file / no skills; a provider that asks tellme to list skills, and to list-then-read a skill; then: listed via `list_skills`, the listing includes/omits a name, reports none, read via `read_files`, and the request carried none of a text). | Round-033: every new Gherkin step needs exactly one authoritative DSL row (`dsl-exact-one-match`); the topology audit PASSED (`--root specs/truth/features/cli`: 44 features · 299 module rows · 1533 steps). |
| MODIFY | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` + `specs/truth/features/cli/chat/dsl.md` (`the request offered exactly the agent tools` row) | The offered agent-tool set is now **seven** tools — the reader trio, the write pair, `execute_command`, and `list_skills` (round 033). | `truth-current`: the tool surface grew; the offered set must include the new read-only `list_skills` tool. |
