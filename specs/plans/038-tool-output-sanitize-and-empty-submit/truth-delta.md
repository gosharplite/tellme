# Truth Delta: 038-tool-output-sanitize-and-empty-submit

**Plan Package**: `specs/plans/038-tool-output-sanitize-and-empty-submit`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md`:
>   - the **Agent command tool (`execute_command`)** row (and/or the **Agent tool loop** row): the live `[Tool Output]` block sanitizes terminal control sequences from the streamed lines and guarantees a neutral terminal state at block close.
>   - the **Interactive TUI prompt (`-i`)** row: submitting an **empty** editor (`Ctrl+S`/`Alt+Enter`) is a **no-op** — the prompt stays open (reference parity); only a non-empty submission submits.
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/watching-the-tool-loop.feature` (a new Rule: the streamed `[Tool Output]` carries no terminal control sequences) + `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (a new Rule: an empty submit is ignored) + the matching `specs/truth/features/cli/chat/dsl.md` rows.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent command tool (`execute_command`)** row | The streamed `[Tool Output]` **content** lines are sanitized at the `internal/ui` presentation seam (all ANSI escape sequences — CSI/SGR, OSC, other ESC-introduced — and stray C0/DEL removed; TAB kept; bytes `≥ 0x80` untouched), and the block **always closes in a neutral state** (a default-state restore before the closing separator on every close path). Presentation-only — the tool **result** keeps its raw bytes. **Recorded divergence:** the reference has no output sanitizer. | `spec.md` FR-001/FR-002/FR-005; `research.md` D1/D2/D3/D6. |
| MODIFY | `specs/truth/techstack.md` — **Interactive TUI prompt (`-i`)** row | Submitting an **empty** (or whitespace-only) editor with `Ctrl+S`/`Alt+Enter` is a **no-op** — the prompt stays open (reference parity); only a **non-empty** submit submits, and `Esc`/`Ctrl+C` still aborts. | `spec.md` FR-003/FR-004; `research.md` D4. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round changes only a `stderr`-bound **presentation** (the `[Tool Output]` sanitization + neutral close) and the `-i` prompt's empty-submit handling. | `contract-authoritative` holds vacuously; `spec.md` FR-006; `plan.md`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record` and the `~/.tellme/*.jsonl` record shapes | No persisted-state change: command output is streamed to `stderr` and never persisted, and the empty-submit no-op writes nothing. | `spec.md` FR-006; `research.md` D7; `plan.md`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — new `Rule: A shell command's streamed output is free of terminal control sequences` | ADD a Rule with two Examples: a colouring command's streamed output is control-sequence-free (visible text preserved), and a command stopped mid-output leaves the terminal in its default state. | `spec.md` FR-001/FR-002/FR-005; `research.md` D1/D2/D3/D5. |
| MODIFY | `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` — new `Rule: An empty submit at the interactive prompt is ignored` | ADD a Rule: submitting an empty editor is a no-op; the following non-empty submit runs exactly one turn (reuses the existing shown/request/answer Thens). | `spec.md` FR-003/FR-004; `research.md` D4. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — +2 `Given` (colouring commands) + 1 `When` (empty-then-real submit) + 2 `Then` (control-sequence-free; terminal neutral) + a round-038 note | ADD the new rows, each matching exactly one step (`dsl-exact-one-match`); `dsl-single-authority` preserved (new rows, no duplication). | `spec.md` FR-001/FR-003/FR-005; `research.md` D5. |
