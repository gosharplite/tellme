# Plan — Prompt suggestion parity: the newest-50 candidate pool (round 064)

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Truth root**: `specs/truth` · **Interface kind**: `cli` (plain line CLI — no `ui/**`; the `-i` prompt's *rendered* surface is unchanged, only the suggestion *content* changes)

## Interfaces inventoried

| Interface | Kind | Planner | Wave |
| --- | --- | --- | --- |
| The CLI chat end (`specs/truth/features/cli/chat/**`) | `cli` | **carried to its contract owner** `/axb-dsl-refine` (a line CLI has no API/data/UI planner) | 1 |

- `/axb-api-plan` — **NOOP** (no OpenAPI/HTTP surface; a single CLI end).
- `/axb-data-plan` — **NOOP** (checked): the shared prompt-log record (`prompt_log_entry`) is **unchanged**; only the **read depth** into it moves. No persisted shape change.
- `/axb-ui-plan` — **skipped** (the `-i` visual surface — border, placeholder, styling, keybindings — is unchanged; only the candidate pool deepens).

## Waves

**Wave 1 (single, no dependencies)** — the CLI contract owner refines the round-064 acceptance journey into executable truth:
- the existing **`chat/prompting-with-suggestions.feature`** MODIFY-ed with a **depth-distinguishing Rule/Example** (a prompt recorded **beyond the newest 10** but **within the newest 50** is offered; the surfaced list stays ≤10);
- **`chat/dsl.md`** MODIFY-ed: the step rows the new Example needs (the Given that seeds a deep log; the Then that asserts the offered prompt / the ten-suggestion count);
- **no** new source, config key, port, or `-i` surface change (Q1 → A).

## The change surface (RD)

| # | Site | Change |
| --- | --- | --- |
| 1 | `internal/app/suggestions/service.go` | split the single `maxSuggestions = 10` into **`promptPoolDepth = 50`** (the depth the history source is asked for) + **`maxSuggestions = 10`** (the surfaced cap, unchanged); `addPrompts` requests `promptPoolDepth` |
| 2 | `internal/app/suggestions/service_test.go` | a unit pin: the engine asks the source for the deepened depth and still caps the surfaced list at 10 (a fake `PromptSource` records the requested `n`) |
| 3 | `tests/e2e/steps/*` (suggestion Givens/Whens) | a Given that seeds the shared log beyond the newest-10 window; reuse the existing `-i` seams (`TELL_ME_FORCE_STDIN_TTY`, `TELL_ME_TUI_DEBOUNCE=0`) |
| 4 | `docs/domain-model/tellme.modelith.{yaml,md}` | **MODIFY** (review TD-3): correct the stale "seeds from the `PromptLog`, **the session**, the workspace, and the tool registry" clause to "no separate session source"; `make modelith-render` + `modelith-check` green (ADR 0030 §D4 — truth wins, the descriptive docs are corrected) |

## Layer/architecture notes

- The change is **one constant split + one argument** inside `internal/app/suggestions` (the application coordinator) — no new package, no cross-layer edge; `verify-architecture` untouched.
- **No dependency change** (`go.mod`/`go.sum` unchanged); stdlib only; POSIX-only.
- The `internal/cli` construction seam is **untouched** (the CLI still wires the same three sources; the CLI's *no-startup-disk-I/O* property is preserved — D3 of the research).
- The `-i` **rendered** surface is unchanged (round-016 territory); only the *content* of the suggestion list moves.

## Test strategy

| Layer | Carrier |
| --- | --- |
| Unit (engine) | a fake `PromptSource` recording the requested `n` — proves the pool depth is the deepened constant while the surfaced list stays capped at 10; a table over match counts proves the ≤10 cap and the subsequence rule |
| E2E | the truth feature's round-064 Rules: the **depth** Example (a scripted `-i` run whose shared-log pool holds a match **beyond the newest 10** offers it) and the **cap** Example (20 matching prompts: the newest is offered, the eleventh is *not*) |
| Regression | the existing suggestion Examples (recent-prompt, workspace path, tool name, accept, >3-line drop) stay GREEN — the change is depth-only |

Witnesses (reproduced then reverted): (a) restoring the depth constant to **10** turns the unit depth pin **RED** (`promptPoolDepth (10) must be strictly deeper than the surface cap (10)`) and the E2E depth Example **RED**; (b) raising the **surface cap** (`maxSuggestions = 20`) turns the **E2E cap Example RED** (`The ten nearest matches are offered and the surplus is dropped`, `267/268`) while the unit pin stays GREEN — the unit pin's `len(got) == maxSuggestions` clause is self-referential (it moves with the mutation), so the **E2E cap Example is the value-binding cap carrier** (review R-4).
