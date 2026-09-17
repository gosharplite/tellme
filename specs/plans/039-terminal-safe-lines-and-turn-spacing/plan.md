# Plan — round 039 (terminal-safe `[Tool …]` lines + turn-output blank-line grouping)

Plan package: `specs/plans/039-terminal-safe-lines-and-turn-spacing`

## Theme

Two folded workstreams on the `stderr`-bound diagnostic surface:

1. **#80** — round 038 sanitized **one** `[Tool …]` formatter (`[Tool Output]`); the siblings render the same provenance class and remain raw, so a model-authored `reason` / an escape-bearing result / an escape-bearing argument can still tint or alter the operator's terminal. Round 039 generalizes the sanitization into a **single-owned** `internal/ui` policy applied by **every** `[Tool …]` formatter (`[Tool Reason]`, `[Tool Result]`, `[Tool Action]` keys+values), preserving the round-036 fold/trim/cap semantics and the round-038 removed class.
2. **Operator spacing request** — the live turn output is grouped: one blank before **each** call's begin block (the `[Tool Reason]` line, else the `[Tool Action]` line), one blank before the trailing grouped `[Tool Reason]` block (none inside it), and one blank before the post-status group (`Payload:` + metrics + `Ready`).

Presentation only — no flag, exit code, frozen vocabulary, tool result, transport, persisted record, or `stdout` byte changes (`spec.md` FR-009–FR-011).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (diagnostic stream)** — the `chat` module's decomposed `[Tool …]` tool-loop presentation (reason / action / result lines + the grouped tail + the post-status group) | `cli` | The subject of the round. The `[Tool Reason]`/`[Tool Result]`/`[Tool Action]` lines become **control-free** (same class as `[Tool Output]`), and the live turn output gains the three blank-line separations. `stdout` stays byte-exact; the block literals, the block bound/stop/spinner-yield, the caps, the class-phrase vocabulary, and the tool schemas are unchanged. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change** (the affected text is streamed to `stderr` and never persisted; blanks and stripped escapes are presentation-only).

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies: `/axb-api-plan` and `/axb-data-plan` NOOP; `/axb-ui-plan` is **skipped** — no chrome/screen change), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` **ADDS** two Rules in `chat/watching-the-tool-loop.feature`: (a) *every tool-loop line is free of terminal control sequences*; (b) *the live turn output is grouped by blank lines*, plus their `chat/dsl.md` rows.

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `~/.tellme/*.jsonl` record shapes: no persisted-state change — the affected text is streamed to `stderr` and never persisted (`spec.md` FR-009/FR-011; `research.md` D6).
- **`/axb-ui-plan` → skipped.** Plain-CLI text-stream change; no TUI chrome or screen change (`spec.md` assumptions).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the two axes (`spec.md` FR-001/FR-005/FR-006/FR-007/FR-008; `research.md` D1–D5):

- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — **ADD** a Rule: *a tool-using run's reason, result, and action lines are free of terminal control sequences*:
  - an escape-bearing `reason` ⇒ the `[Tool Reason]` line is control-free (reuses `the run reported the reason … for the tool call …`, strengthened / or a new `Then the run reported the reason for the tool call … free of terminal control sequences`);
  - an escape-bearing result ⇒ the `[Tool Result]` line is control-free;
  - an escape-bearing argument ⇒ the `[Tool Action]` line is control-free.
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` — **ADD** a Rule: *the live turn output is grouped by blank lines*:
  - `Then each tool call's report begins after a blank line` (per-call begin block; a reason-less call's action line included);
  - `Then the trailing reason summary follows a blank line`;
  - `Then the turn's closing status follows a blank line`.
- `specs/truth/features/cli/chat/dsl.md` — **ADD** the rows above (the new `Given`s for escape-bearing fixtures + the new `Then`s), each matching exactly one step (`dsl-exact-one-match`); `dsl-single-authority` is preserved (new rows, no duplication).

The witness set is **unit hostile fixtures** for the three sibling formatters (control-free + valid UTF-8, caps preserved; plus a `[Tool Output]` **byte-identical** regression pin and an escape-only-reason suppression pin), **plus** the E2E `Then`s (the escape byte and the blank positions are directly observable in the captured `stderr`) — `research.md` D7.

`truth-delta.md` carries the explicit rows; `specs/truth/techstack.md` (Agent tool loop + `execute_command` rows) and **ADR 0008** (superseding ADR 0007) are owned by `/axb-technical-research` (already folded).
