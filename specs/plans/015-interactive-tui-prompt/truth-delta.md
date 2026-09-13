# Truth Delta: 015-interactive-tui-prompt

**Plan Package**: `specs/plans/015-interactive-tui-prompt`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application → Interactive TUI prompt (`-i`)**: `bubbletea` + `bubbles` (textarea) + direct `lipgloss` — tellme's **first TUI dependency**; opt-in, gated by `-i`/`USE_TUI_PROMPT` + a terminal stdin. **Prompt suggestion engine**: multi-source (shared log + session + workspace + tools), subsequence, deduped, ≤10, ~100 ms debounce. **Shared global prompt log**: append-only JSONL at `output/global_prompts.jsonl` (`{timestamp, prompt}`, RFC3339), read newest-first + deduped, written **only under `-i`**, compaction ≈150 KiB/≤1200 unique. **Session dashboard**: reuses the provider/model + round-009 token figures + history turn count. **Terminal detection**: the round-012 real-isatty seam now also gates the TUI. **CLI flag parsing**: adds `-i`/`--interactive`; **Configuration**: adds `USE_TUI_PROMPT`. **Testing & Verification**: the injected-I/O + scripted-key TUI harness (no pty). **Not Introduced Yet**: TUI libraries now introduced; Windows variant excluded. | Round-015 research Decisions 1–9: adopt the Bubble Tea family, the multi-source suggestion engine, the append-only shared prompt log (write only under `-i`), the opt-in TUI with the round-012 default preserved, the reused dashboard state, and the hermetic no-pty verification. |

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
