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
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and authors no HTTP/OpenAPI surface of its own; round 015 adds no tellme-owned request/response shape — the interactive prompt + shared log are a local terminal + local-file concern. | `contract-authoritative` holds vacuously — round 015 changes the CLI prompt surface and adds a local-file store, not a tellme-owned API. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/data/data-model.dbml` | New table `prompt_log_entry` — the shared, append-only global prompt log at `$TELL_ME_HOME/output/global_prompts.jsonl`, one `{"timestamp":"<RFC3339>","prompt":"<text>"}` per line (both fields string). Append-only (`O_APPEND|O_CREATE`), written **only under `-i`**; read newest-first + deduped by `prompt` (bounded by the suggestion cap); best-effort async compaction keeping the newest ≤1200 unique once past ≈150 KiB. Byte-identical shape to `tell-me-go`; no cross-process `flock`. The Project Note was broadened to cover both local-state artifacts (session history + global prompt log). | Round-015 research Decision 3 + spec `FR-008`–`FR-011`/`NFR-003`/`NFR-004`: `tellme` must interoperate with the shared prompt log (a data-model addition — `data-model-covers-all-state`). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/**` | _To be recorded by `/axb-dsl-refine`: the interactive-prompt acceptance rules and DSL rows (invocation, suggestion engine, dashboard, keybindings, non-TTY fallback, global-log interop)._ | _Round 015 carries its acceptance journeys into executable interface truth._ |
