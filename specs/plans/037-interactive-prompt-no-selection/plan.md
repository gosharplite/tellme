# Plan — round 037 (interactive prompt no-selection)

Plan package: `specs/plans/037-interactive-prompt-no-selection`

## Theme

Align the `-i` interactive prompt's suggestion **selection** to the reference: the list opens with **no** suggestion pre-selected and **resets to no-choice on every refresh** (`cursor == -1`; the reference's `suggester{Index: -1}` + `Update(msg, -1)`), so an empty editor never shows a "chosen" hint; the first `Tab` still selects the **first** suggestion. A presentation-parity correction on a CLI (`stderr`-bound TUI) surface — no flag, exit code, ordering, glyph, styling, debounce, persisted record, or `stdout` byte changes (`spec.md` FR-001–006).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (interactive terminal surface)** — the `chat` module's `-i` prompt suggestion **selection state** | `cli` | The subject of the round: the at-rest/refresh selection of the suggestion list rendered by the round-015/016 TUI prompt. `stdout` stays byte-exact; the suggestion sources/ordering, the `Suggestions:` header, the selected/unselected styling, the `> ` glyph, the editor, the placeholder, the debounce, the no-metrics-header behaviour and the round-023 teardown are unchanged. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change** (a selection index is in-memory).

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies: `/axb-api-plan` and `/axb-data-plan` NOOP; `/axb-ui-plan` runs in **terminal mode** for the plan-side screens only), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` **MODIFIES** the at-rest `Then` **in place** (`the interactive prompt marks **no** suggestion as the current choice`) in `presenting-the-interactive-prompt.feature` and the matching row in `chat/dsl.md` (the `不該發生` clause is restored to its originally-drafted sense: *no row must carry the cursor* at rest). **No new `DSLRow`; no new feature file.**

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract — it changes only the at-rest selection state of a `stderr` terminal frame (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `~/.tellme/global_prompts.jsonl` record shape: no persisted-state change — a suggestion's selection index is in-memory TUI state, never written (`spec.md` FR-005; `research.md` D6).
- **`/axb-ui-plan` → terminal mode.** The `-i` prompt ships a rich terminal UI, so a plan-side `ui/` is produced: `specs/plans/037-interactive-prompt-no-selection/ui/screens/entry.txt` re-renders the **at-rest** frame with **no** `>` cursor (the suggestion list otherwise byte-identical). `/axb-system-analysis` reviews it, never redraws it; it is plan-side only (`prototype-plan-side-only`).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must reconcile the executable CLI truth for the selection axis (`spec.md` FR-001/FR-003; `research.md` D1/D3):

- `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` — **MODIFY** the at-rest rule/Example in place: `Then the interactive prompt marks no suggestion as the current choice`. The after-`Tab` selection stays carried by the **existing** accept journey in `prompting-with-suggestions.feature` (no second rule — keeps `acceptance-coverage` exact).
- `specs/truth/features/cli/chat/dsl.md` — **MODIFY** the row `the interactive prompt marks one suggestion as the current choice` → `the interactive prompt marks no suggestion as the current choice` with its `不該發生` clause restored (*no* row carries the cursor at rest). Also add a **round-037 note** recording the no-choice sentinel + reset-on-refresh and the supersession of round 016.
- **No new `DSLRow`**; `watching-the-tool-loop.feature`, the suggestion-source features and `prompting-with-suggestions.feature` are unchanged.

The witness is the updated **unit pins** (`model_chrome_test.go`, `refresh_test.go`) **plus** the flipped **E2E** `Then` (`step_t008`) — the cursor row is directly observable in the captured frames (`research.md` D4).

`truth-delta.md` carries the explicit rows; `specs/truth/techstack.md` is MODIFY-owned by `/axb-technical-research` (already folded).
