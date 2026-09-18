# System Analysis Plan: composition-root extraction (round 044)

**Plan Package**: `specs/plans/044-composition-root-extraction`
**Anchor**: [#100](https://github.com/gosharplite/tellme/issues/100) — R2 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> `/axb-system-analysis` output: the interface inventory, the classification, and the dependency-ordered delegation waves. It **orchestrates** analysis only; it never creates or modifies a `TruthArtifact` (`analysis-plan-never-writes-truth`).

## Interface inventory

| # | Interface | Kind | Touched by this round? | Planner | Disposition |
| --- | --- | --- | --- | --- | --- |
| 1 | `tellme` CLI end (flags, exit codes, `stdout`/`stderr` streams, formatting) | `cli` | **No behaviour change** (structural refactor) | `/axb-dsl-refine` (contract owner) | **NOOP** — no feature/DSL row changes; the vocabulary stays 11; the gate + unit seams are the acceptance carrier |
| 2 | Session-history store (`history.jsonl`), usage log (`tokens.log`), tool-usage log (`~/.tellme/tools-count.jsonl`), prompt log (`~/.tellme/global_prompts.jsonl`) | data | **No** | `/axb-data-plan` | **NOOP** — no record/location/lifecycle change |
| 3 | Provider HTTP surface (Vertex/Gemini, OpenAI-compatible) | backend | **No** | `/axb-api-plan` | **NOOP** — no request/response shape change |

**New structural element (not an interface):** the composition root moves to `cmd/tellme`; the injected port value is `internal/app/deps.Dependencies` + a cli-local `cli.Options`. This is an **architecture** change recorded in `specs/truth/techstack.md` (CLI Application) and **ADR 0013** — not an interface change.

## Wave plan

| Wave | Scope | Delegation |
| --- | --- | --- |
| **0** | Truth owners check their areas against the round | `/axb-technical-research` (completed — MODIFY techstack + **ADR 0013**); `/axb-api-plan` → **NOOP**; `/axb-data-plan` → **NOOP**; `/axb-dsl-refine` → **NOOP** |
| **1** | No dependency-ordered interface waves — this round has **no interface change** to sequence. The only planning output is the truth/architecture record (Wave 0) and the task list (`/axb-tasks`, a later branch) | — |

`wave-covers-interfaces`: the CLI end (interface #1) is carried **forward to its contract owner `/axb-dsl-refine`** with a **NOOP** disposition (no Rule/Example/step/`DSLRow` change); interfaces #2/#3 are NOOP. No interface is left un-owned.

## Truth handoff

- `/axb-technical-research` → `specs/truth/techstack.md` MODIFY (the rows naming the moved symbols) + `docs/decisions/0013-composition-root-injection.md`; recorded in `truth-delta.md`.
- `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-dsl-refine` = **NOOP** — each recorded as an evidenced NOOP row in `truth-delta.md`.
- `/axb-ui-plan` **skipped** (not user-facing).

## Interface classification summary

- **1 CLI end** — no planner wave; carried to `/axb-dsl-refine` as a **NOOP** (the CLI contract is unchanged; the executable truth tree is untouched).
- **0 backend / 0 frontend / 0 new data** interfaces.

## Notes for `/axb-tasks`

- The round's **witness is the R1 layer-discipline gate** (`make verify-architecture`) + the unit seams — **not** the E2E suite (`spec.md` A1 / FR-012).
- The round carries a **ratchet** discipline: each removed edge + its baseline regeneration land in the **same commit** (FR-008 / `research.md` D6).
- The task list must include the **falsifiability witness** (re-introduce one `cli → infrastructure` import ⇒ the gate reddens; revert).
