# Truth Delta: 037-interactive-prompt-no-selection

**Plan Package**: `specs/plans/037-interactive-prompt-no-selection`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md` (the interactive-prompt / suggestions row: the suggestion selection starts at **no choice** and resets on refresh, aligned to the reference).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (the at-rest `Then` flips from *marks one suggestion as the current choice* to *marks no suggestion as the current choice*) + **MODIFY** the matching row in `specs/truth/features/cli/chat/dsl.md`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Interactive TUI prompt (`-i`)** row + **Prompt suggestion engine** row | The suggestion list opens with **no** suggestion pre-selected: the selection starts at a **no-choice** sentinel (`cursor == -1`, mirroring the reference's `suggester{Index: -1}`) and **resets to no-choice on every suggestions refresh** (the reference's `Update(msg, -1)`); the first `Tab` still selects the **first** suggestion. **Supersedes** round 016's pre-selection (which followed the reference's mockup; the reference's code never pre-selects). No dependency, ordering, glyph, styling, or debounce change. | `spec.md` FR-001/FR-002/FR-003; `research.md` D1/D2/D3. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; this round changes only the at-rest **selection state** of the `-i` TUI prompt (a `stderr`-bound terminal frame). | `contract-authoritative` holds vacuously; `spec.md` FR-005; `plan.md`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record` and the `~/.tellme/global_prompts.jsonl` record shape | No persisted-state change: a suggestion's **selection index** is in-memory TUI state and is never written; the shared prompt log's record shape and the session/usage records are untouched. | `spec.md` FR-005; `research.md` D6; `plan.md`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` — the at-rest `Then` | The at-rest assertion flips from *the interactive prompt marks **one** suggestion as the current choice* to *… marks **no** suggestion as the current choice* (zero `>` cursor rows), in place; the header comment notes the round-037 supersession of round 016 and lists the new acceptance journey. | `spec.md` FR-001; `research.md` D1/D3/D5. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — the current-choice row + a round-037 note | The row `the interactive prompt marks one suggestion as the current choice` is replaced by `the interactive prompt marks no suggestion as the current choice` (its `不該發生` clause restored: no row carries the cursor at rest; the after-`Tab` selection is carried by the accept journey). A round-037 note records the no-choice sentinel + reset-on-refresh and the supersession of round 016. **No new `DSLRow`** (the row is edited in place — `dsl-single-authority` preserved). | `spec.md` FR-001/FR-003; `research.md` D1/D2/D4. |
