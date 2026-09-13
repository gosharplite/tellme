# Truth Delta: 015-interactive-tui-prompt

**Plan Package**: `specs/plans/015-interactive-tui-prompt`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/techstack.md` | _To be recorded by `/axb-technical-research`: the interactive TUI prompt dependency (the round's first TUI-grade library), the suggestion engine, the dashboard, and the shared global prompt log store._ | _Round 015 introduces the `-i` interactive TUI prompt._ |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | _Expected NOOP — tellme authors no HTTP/OpenAPI surface of its own._ | _Round 015 changes the CLI prompt surface, not a tellme-owned API._ |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/data-model.dbml` | _Expected MODIFY — model the shared global prompt log (`$TELL_ME_HOME/output/global_prompts.jsonl`): record shape, append-and-read lifecycle, dedupe/newest-first ordering, and compaction policy._ | _The suggestion engine must interoperate with the shared Niffler-env prompt log (a data-model addition)._ |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/**` | _To be recorded by `/axb-dsl-refine`: the interactive-prompt acceptance rules and DSL rows (invocation, suggestion engine, dashboard, keybindings, non-TTY fallback, global-log interop)._ | _Round 015 carries its acceptance journeys into executable interface truth._ |
