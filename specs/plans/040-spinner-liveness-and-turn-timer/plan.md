# Plan — round 040 (spinner liveness while a command streams + a per-turn elapsed timer)

Plan package: `specs/plans/040-spinner-liveness-and-turn-timer`

## Theme

Two folded workstreams on the round-019/025 **progress spinner** (`internal/ui/spinner.go`, drawn on the `stderr` diagnostic stream):

1. **#82 — liveness while a command's `[Tool Output]` streams.** Today the spinner is **paused for the whole live output block** (round-034 FR-012/G8, ADR 0005 D7), so a quiet/slow command shows no animation at all. Round 040 keeps the block's single-writer `stderr` ownership but **resumes the indicator after an idle gap** (no new output line for N seconds) and **synchronously clears it before the next output line**.
2. **#83 — a dual elapsed timer.** The line gains a **second** whole-second figure — the **current turn's** duration, reset at each **AI-endpoint call** — alongside the existing total since prompt capture: `({total}s {turn}s)`.

Presentation only — no flag, exit code, frozen vocabulary, tool result, transport, persisted record, or `stdout` byte changes (`spec.md` FR-009–FR-011).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (diagnostic stream)** — the `chat` module's progress spinner (`internal/ui/spinner.go`) and the `[Tool Output]` block it yields to | `cli` | The subject of the round. The spinner line carries **two** figures (total + per-AI-endpoint-call turn time); during a `[Tool Output]` block the indicator **resumes after an idle gap** and is **cleared before the next line**. `stdout` stays byte-exact; the block literals, the bound/stop semantics, the `isatty(stderr) && !-r` gate, the `-i` surface, the class-phrase vocabulary, and the tool schemas are unchanged. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change** (the spinner and the block are `stderr`-only presentation and are never persisted).

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies: `/axb-api-plan` and `/axb-data-plan` NOOP; `/axb-ui-plan` is **skipped** — no TUI chrome/screen change), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` **MODIFIES** `chat/presenting-the-progress-spinner.feature`: it **amends** `Rule: The spinner is paused while a command's output streams` (the indicator is not hidden for the whole block — it resumes after an idle gap and clears on the next line) and **adds** a Rule for the dual elapsed timer, plus their `chat/dsl.md` rows.

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `~/.tellme/*.jsonl` record shapes: no persisted-state change — the spinner and the `[Tool Output]` block are `stderr`-only presentation and are never persisted (`spec.md` FR-009; `research.md` D7).
- **`/axb-ui-plan` → skipped.** The spinner is a single `stderr` status line, not TUI chrome; no screen/keybinding/state-transition artifact applies (`spec.md` assumptions).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the two axes (`spec.md` FR-001–FR-008; `research.md` D1–D5):

- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — **MODIFY** `Rule: The spinner is paused while a command's output streams`:
  - the rule is narrowed: the indicator is **not** hidden for the whole block — it **resumes** during a quiet stretch (no new output line for the idle threshold) and is **cleared before the next output line** (`spec.md` FR-001–FR-004).
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — **ADD** a Rule: *the spinner shows the total time and the current turn's time*:
  - `Then the progress indicator shows the time since the prompt and the current turn's time` (single-call and tool-using cases);
  - `Then the current turn's time restarts when a new model call begins` (`spec.md` FR-005–FR-008).
- `specs/truth/features/cli/chat/dsl.md` — **ADD** the rows above (a new `Given` for the quiet-command fixture + the new `Then`s), each matching exactly one step (`dsl-exact-one-match`); `dsl-single-authority` is preserved (new/amended rows, no duplication).

The witness set is **unit** pins (the two-figure arithmetic under an injected `now`; the idle-gap resume/clear under an injected ticker; the row-aware clear for the longer line) **plus** the E2E `Then`s (a forced idle threshold + a scripted quiet command; a multi-call turn's two figures) — `research.md` D6.

`truth-delta.md` carries the explicit rows; `specs/truth/techstack.md` (**Turn progress spinner** row) and **ADR 0009** (superseding ADR 0005 **D7**, amending round-019 D4) are owned by `/axb-technical-research` (already folded).
