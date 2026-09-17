# Plan — round 038 (tool-output sanitization + empty-submit no-op)

Plan package: `specs/plans/038-tool-output-sanitize-and-empty-submit`

## Theme

Two folded operator issues on the `stderr`-bound interactive surface:

1. **#78** — the live `[Tool Output]` block forwards a shell command's raw bytes, so a colouring/control-emitting command (or one killed mid-output) can tint or alter the operator's terminal for the rest of the run. Round 038 **sanitizes** the streamed **content** lines at the `internal/ui` presentation seam (all ANSI escape sequences + stray C0/DEL removed; TAB kept; UTF-8 preserved) and makes the block **always close in a neutral state** (a default-state restore before the closing separator on every close path).
2. **#76** — `Ctrl+S`/`Alt+Enter` on an **empty** `-i` editor quits the prompt (exit 0); round 038 makes it a **no-op**, mirroring the reference (the prompt stays open; only a non-empty submit submits; `Esc`/`Ctrl+C` still aborts).

Presentation/parity only — no flag, exit code, frozen vocabulary, tool result, transport, persisted record, or `stdout` byte changes (`spec.md` FR-001–007).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (diagnostic stream + interactive terminal surface)** — the `chat` module's `[Tool Output]` presentation and the `-i` submit handling | `cli` | The subject of the round. The `[Tool Output]` **content** lines become control-sequence-free and the block closes neutral; the `-i` empty submit becomes a no-op. `stdout` stays byte-exact; the header/separator literals, the block's bound/stop/spinner-yield semantics, `output_file` handling, the suggestion engine, the editor, the placeholder and the other keybindings are unchanged. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change** (streamed output is never persisted; the no-op writes nothing).

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies: `/axb-api-plan` and `/axb-data-plan` NOOP; `/axb-ui-plan` is **skipped** — no chrome change), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` **ADDS** two Rules: one in `watching-the-tool-loop.feature` (*the streamed `[Tool Output]` carries no terminal control sequences; the block closes neutral*) and one in `using-the-interactive-prompt.feature` (*an empty submit is ignored*), plus their `chat/dsl.md` rows.

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `~/.tellme/*.jsonl` record shapes: no persisted-state change — command output is streamed to `stderr` and never persisted; the empty-submit no-op writes nothing (`spec.md` FR-006; `research.md` D7).
- **`/axb-ui-plan` → skipped.** Plain-CLI text-stream + key-handling change; no TUI chrome or screen change (`spec.md` assumptions).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the two axes (`spec.md` FR-001/FR-003/FR-005; `research.md` D1–D5):

- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — **ADD** a Rule: *a shell command's streamed output is free of terminal control sequences*:
  - `Then the run streamed the command's output free of terminal control sequences` (each `[Tool Output]` **content** line carries no ESC/C0 byte; the visible text preserved).
  - `Then the terminal is left in its default state` (a default-state restore follows the last content line).
- `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` — **ADD** a Rule: *an empty submit at the interactive prompt is ignored*:
  - `When the operator submits an empty prompt then submits the prompt "<prompt>" at the interactive prompt` (reuses the existing `the interactive prompt is shown` / `tellme sends exactly one request …` Thens).
- `specs/truth/features/cli/chat/dsl.md` — **ADD** the rows above (2 `Then` + 1 `When` + the supporting `Given`s for the colouring commands), each matching exactly one step (`dsl-exact-one-match`); `dsl-single-authority` is preserved (new rows, no duplication).

The witness set is **unit hostile fixtures + the writer neutral-close pin + the model empty-submit pin**, **plus** the E2E `Then`s (the escape byte and the neutral restore are directly observable in the captured `stderr`) — `research.md` D5.

`truth-delta.md` carries the explicit rows; `specs/truth/techstack.md` is MODIFY-owned by `/axb-technical-research` (already folded: the `execute_command` row + the `-i` row).
