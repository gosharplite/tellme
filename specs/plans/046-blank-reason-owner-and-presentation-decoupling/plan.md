# System Analysis Plan: blank-reason owner + the loop's presentation de-coupling (round 046)

**Plan Package**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`
**Anchor**: [#108](https://github.com/gosharplite/tellme/issues/108) — R4 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> `/axb-system-analysis` output: the interface inventory, the classification, and the dependency-ordered delegation waves. It **orchestrates** analysis only; it never creates or modifies a `TruthArtifact` (`analysis-plan-never-writes-truth`).

## Interface inventory

| # | Interface | Kind | Touched by this round? | Planner | Disposition |
| --- | --- | --- | --- | --- | --- |
| 1 | `tellme` CLI end (flags, exit codes, `stdout`/`stderr` streams, formatting) | `cli` | **No behaviour change** (structural refactor) | `/axb-dsl-refine` (contract owner) | **NOOP** — no feature/DSL row changes; the class-phrase vocabulary is unchanged; the unit pins + the ADR are the acceptance carrier |
| 2 | Session-history store (`history.jsonl`), usage log (`tokens.log`), tool-usage log (`~/.tellme/tools-count.jsonl`), prompt log (`~/.tellme/global_prompts.jsonl`) | data | **No** | `/axb-data-plan` | **NOOP** — no record/location/lifecycle change |
| 3 | Provider HTTP surface (Vertex/Gemini, OpenAI-compatible) | backend | **No** | `/axb-api-plan` | **NOOP** — no request/response shape change |

**New structural element (not an interface):** an injected **`agentport.ToolLineRenderer`** port (declared in `internal/domain/agent`, implemented by `internal/ui`) that replaces the loop's direct `internal/ui` import, plus the **single-owned blank-reason predicate** (`ui.ToolLineRenderer.ReasonLine`) and the removal of the dead `callRenderer.OnCallEnd` re-check. Both are recorded in `specs/truth/techstack.md` (CLI Application; Build & Tooling) and **ADR 0015** — not an interface change.

## Wave plan

| Wave | Scope | Delegation |
| --- | --- | --- |
| **0** | Truth owners check their areas against the round | `/axb-technical-research` (completed — MODIFY techstack ×2 + **ADR 0015**); `/axb-api-plan` → **NOOP**; `/axb-data-plan` → **NOOP**; `/axb-dsl-refine` → **NOOP** |
| **1** | No dependency-ordered interface waves — this round has **no interface change** to sequence. The only planning output is the truth/architecture record (Wave 0) and the task list (`/axb-tasks`) | — |

`wave-covers-interfaces`: the CLI end (interface #1) is carried **forward to its contract owner `/axb-dsl-refine`** with a **NOOP** disposition (no Rule/Example/step/`DSLRow` change); interfaces #2/#3 are NOOP. No interface is left un-owned.

## Truth handoff

- `/axb-technical-research` → `specs/truth/techstack.md` MODIFY (the **Agent tool loop** row + the **Layer-discipline gate** row) + `docs/decisions/0015-loop-presentation-port.md`; recorded in `truth-delta.md`.
- `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-dsl-refine` = **NOOP** — each recorded as an evidenced NOOP row in `truth-delta.md`.
- `/axb-ui-plan` **skipped** (not user-facing).

## Interface classification summary

- **1 CLI end** — no planner wave; carried to `/axb-dsl-refine` as a **NOOP** (the CLI contract is unchanged; the executable truth tree is untouched).
- **0 backend / 0 frontend / 0 new data** interfaces.

## The mechanism (boundary the tasks implement)

Per `research.md` D1–D7 (clarify **C-R4-1 = A**):

- **The port** — `agentport.ToolLineRenderer` in `internal/domain/agent` (co-located with `LoopObserver`): `EngineLine`, `ActionLine`, `ResultLine`, and `ReasonLine(t, reason) (line string, renders bool)`.
- **The loop** — `internal/agent/agentloop.go` drops `import "…/internal/ui"`, gains a nil-safe `Lines agentport.ToolLineRenderer` field, and routes `logEngine`/`logAction`/`logResult` through it; **the write schedule, the per-call blanks, and the round-045 yield bracket (`withToolLog` → `YieldIndicator`/`RestoreIndicator`) are unchanged** (ADR 0014 untouched). `reasonsOf` becomes a method that consults the port's `ReasonLine` and appends the **raw** reason when it renders.
- **The predicate's single owner** — `ui.ToolLineRenderer.ReasonLine` (ONE `ui.toolReasonText` evaluation returns the line **and** the decision). The **dead** third site (`cli/call_renderer.go` `OnCallEnd`) is **deleted** (C-R4-2); the tail prints the already-filtered `roundReasons`.
- **The wiring** — `internal/cli/cli.go` `runTurn` (the composition site, ADR 0013) injects `Lines: ui.ToolLineRenderer{}` alongside `Now`/`ToolUsage`. No new `deps.Dependencies` seam.
- **The ratchet** — `tools/arch/baseline.txt` regenerates **header-only (0)**; the gate must report **0 new / 0 stale / 0 cycles**.

## Notes for `/axb-tasks`

- The round's **witness is unit pins + the gate** (C-R4-3) — **not** the E2E suite (`spec.md` A1 / FR-008). The real formatting stays E2E-asserted through the production wiring.
- The **`[BDD-*]` markers do not apply** (no truth feature/DSL row changes): Phase 3 collapses to `[UNIT]` pins + a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases.
- The loop's own `_test.go` files **may not** import `internal/ui` (the gate governs test imports) → they inject an in-package **fake renderer** and assert the **schedule**; the **formatting** is asserted at the `ui` tier + E2E.
- The task list must include the **falsifiability witnesses** (a)/(b)/(c) (re-add the import; diverge the predicate from the format; add a stale baseline line at 0).
- The **ratchet reaches 0** this round: regenerate the baseline **in the same commit-range** as the compliant code (the gate ships with its enabler).
