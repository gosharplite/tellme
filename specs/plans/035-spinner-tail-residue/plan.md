# Plan — round 035 (spinner tail residue)

Plan package: `specs/plans/035-spinner-tail-residue`

## Theme

Pin the spinner **tail residue** fix: the round-034 per-call tail (grouped `[Tool Reason]` lines + measured payload + metrics + `Ready`) must be written with the progress indicator **yielded** (a phase-boundary clear, no immediate resume), so the tail's lines begin on their own rows and no braille-frame residue survives into the finished tool-using turn (issue [#72](https://github.com/gosharplite/tellme/issues/72)). An **implementation-vs-truth** correction — a rendering-only, `stderr`-diagnostic change (`spec.md` FR-005/FR-006).

## System interface inventory

| # | Interface | Kind | Notes |
| --- | --- | --- | --- |
| 1 | **CLI end (terminal diagnostics)** — the `chat` module's progress-spinner surface | `cli` | The subject of the round: **when/how** the per-call tail is written relative to the live spinner on `stderr`. `stdout` (the answer) stays byte-exact; no line format or cadence changes. |

No other interface is touched: there is **no HTTP/API surface** and **no persisted-state change**.

## Wave plan (1 wave)

- **Wave 1 — CLI end → `/axb-dsl-refine`.** The CLI end has **no analysis planner** (no API/data/UI planner applies), so it is carried forward to its **contract owner** `/axb-dsl-refine`, which owns `specs/truth/features/cli/**` + `dsl.md`. `/axb-dsl-refine` MODIFIES `chat/presenting-the-progress-spinner.feature` (+ a `chat/dsl.md` note).

## Planner delegation / NOOPs

- **`/axb-api-plan` → NOOP.** tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response contract — the fix changes only a `stderr` diagnostic write (`truth-delta.md`).
- **`/axb-data-plan` → NOOP (checked).** Inspected `specs/truth/data/data-model.dbml` (`history_entry`/`history_step`/`usage_record`) + the `output/<mode>/tokens.log` cadence: no persisted-state change — the fix alters only when/how the tail is written to the terminal (`spec.md` FR-005/FR-006; `research.md` D6).
- **`/axb-ui-plan` → skipped.** A line-oriented CLI diagnostic: no web/HTML prototype and no TUI screen change; the terminal interaction lives directly in the CLI interface Gherkin (the standing CLI streamlining).

## Analysis focus (handoff to `/axb-dsl-refine`)

`/axb-dsl-refine` must MODIFY the executable CLI truth to exercise the residue rule on the gated **tool-tail** path (`spec.md` FR-001/FR-002/FR-004; `research.md` D3/D4):

- `chat/presenting-the-progress-spinner.feature` — **add a gated tool-tail Example**: `the diagnostics are shown at a terminal` (forces the spinner gate on) + a scripted provider that reads `notes.txt` then answers + the whole-stream residue row `the run shows no progress spinner`. Under the defect it fails (`⠋ Executing …` survives); after the fix it passes. It reuses the interface-root row (no new `DSLRow`).
- `chat/dsl.md` — a **round-035 note** recording the phase-boundary yield (clear-before, no-resume) and that the whole-stream residue row now covers the gated tool-tail path.

`truth-delta.md` carries the explicit MODIFY rows; `specs/truth/techstack.md` is MODIFY-owned by `/axb-technical-research` (already folded).
