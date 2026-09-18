# System Analysis Plan: yield-policy owner + `LoopObserver` hook split (round 045)

**Plan Package**: `specs/plans/045-yield-policy-owner`
**Anchor**: [#105](https://github.com/gosharplite/tellme/issues/105) — R3 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> `/axb-system-analysis` output: the interface inventory, the classification, and the dependency-ordered delegation waves. It **orchestrates** analysis only; it never creates or modifies a `TruthArtifact` (`analysis-plan-never-writes-truth`).

## Interface inventory

| # | Interface | Kind | Touched by this round? | Planner | Disposition |
| --- | --- | --- | --- | --- | --- |
| 1 | `tellme` CLI end (flags, exit codes, `stdout`/`stderr` streams, formatting) | `cli` | **No behaviour change** (structural refactor) | `/axb-dsl-refine` (contract owner) | **NOOP** — no feature/DSL row changes; the vocabulary stays **11**; the unit ordering pins + the ADR are the acceptance carrier |
| 2 | Session-history store (`history.jsonl`), usage log (`tokens.log`), tool-usage log (`~/.tellme/tools-count.jsonl`), prompt log (`~/.tellme/global_prompts.jsonl`) | data | **No** | `/axb-data-plan` | **NOOP** — no record/location/lifecycle change |
| 3 | Provider HTTP surface (Vertex/Gemini, OpenAI-compatible) | backend | **No** | `/axb-api-plan` | **NOOP** — no request/response shape change |

**New structural element (not an interface):** a single-owned yield policy — `internal/ui/yield.go`'s `YieldController` — plus a renamed observer hook pair (`YieldIndicator`/`RestoreIndicator`) on `internal/domain/agent.LoopObserver`. This is an **architecture** change recorded in `specs/truth/techstack.md` (CLI Application) and **ADR 0014** — not an interface change.

## Wave plan

| Wave | Scope | Delegation |
| --- | --- | --- |
| **0** | Truth owners check their areas against the round | `/axb-technical-research` (completed — MODIFY techstack + **ADR 0014**); `/axb-api-plan` → **NOOP**; `/axb-data-plan` → **NOOP**; `/axb-dsl-refine` → **NOOP** |
| **1** | No dependency-ordered interface waves — this round has **no interface change** to sequence. The only planning output is the truth/architecture record (Wave 0) and the task list (`/axb-tasks`) | — |

`wave-covers-interfaces`: the CLI end (interface #1) is carried **forward to its contract owner `/axb-dsl-refine`** with a **NOOP** disposition (no Rule/Example/step/`DSLRow` change); interfaces #2/#3 are NOOP. No interface is left un-owned.

## Truth handoff

- `/axb-technical-research` → `specs/truth/techstack.md` MODIFY (the **Agent tool loop** + **Turn progress spinner** rows) + `docs/decisions/0014-yield-policy-owner.md`; recorded in `truth-delta.md`.
- `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-dsl-refine` = **NOOP** — each recorded as an evidenced NOOP row in `truth-delta.md`.
- `/axb-ui-plan` **skipped** (not user-facing).

## Interface classification summary

- **1 CLI end** — no planner wave; carried to `/axb-dsl-refine` as a **NOOP** (the CLI contract is unchanged; the executable truth tree is untouched).
- **0 backend / 0 frontend / 0 new data** interfaces.

## Notes for `/axb-tasks`

- The round's **witness is the unit ordering pins + ADR 0014** — **not** the E2E suite (`spec.md` A1 / FR-009).
- The **`[BDD-*]` markers do not apply** (no truth feature/DSL row changes): Phase 3 collapses to `[UNIT]` ordering-pin tasks + a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases.
- The task list must include the **falsifiability witnesses** (a)/(b)/(c) (drop the Y2 clear; resume after the Y2 tail; a self-locking owner).
- The **ratchet** stays at **1**: do **not** remove the `internal/agent -> internal/ui` baseline line (R4's DoD, C-R3-4); the gate must report **0 new / 0 stale**.
- **Stale-row guard**: `/axb-dsl-refine`'s NOOP is evidenced by a grep of the truth tree for the old hook names — a task `Read` must carry the `chat/dsl.md` notes so the round does not silently orphan them.
- `go test -race ./internal/ui/...` is a **required** gate (the coordinator stress is the anti-vacuity guard for H3).
