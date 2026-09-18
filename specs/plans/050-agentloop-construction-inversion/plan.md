# System Analysis Plan: round 050 — invert the `AgentLoop` construction/execution into an injected domain port

**Plan Package**: `specs/plans/050-agentloop-construction-inversion`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — the **sub-slice 2** of the re-cut `cli → agent` de-coupling; the **baseline-moving** round (**2 → 1**).
**Created**: 2026-09-18 · **Skill**: `/axb-system-analysis`

## 1. Interface inventory

This round is an **internal structural refactor** of the `internal/cli` ↔ `internal/agent` seam. It ships **no** new system boundary.

| Interface | Kind | Classification | Planner |
| --- | --- | --- | --- |
| `internal/cli` turn path (the loop seam) | dev-internal | **not** a user-facing CLI end — the round does not change flags, arguments, exit codes, `stdin`/`stdout`/`stderr` shapes, or the DSL vocabulary | **carried forward to its contract owner** |
| CLI end (the `tellme` command surface) | `cli` | unchanged — **no** new Rule/Example/step/`DSLRow` | `/axb-dsl-refine` (**NOOP**) |
| API end | `backend` | none — tellme has **no** OpenAPI/HTTP surface | `/axb-api-plan` (**NOOP**) |
| Data end | `—` | none — **no** persisted/runtime-state change | `/axb-data-plan` (**NOOP**) |
| UI end | `—` | none (a line-oriented CLI; `/axb-ui-plan` skipped) | none |

## 2. Waves

**No delegated planner wave applies** (0 interfaces invoke a planner). Per the CLI-streamlined workflow (aixbdd-tmg README §Developing CLI Applications) the `cli` end is **carried forward to its contract owner, `/axb-dsl-refine`**, which this round records as **NOOP** (an internal refactor is not a `tellme` CLI-contract change). The `api`/`data` ends are **NOOP** (no API surface; no state change).

| Wave | Order | Delegation |
| --- | --- | --- |
| — | — | none (no planner invoked) |

## 3. The seam under change (grounded @ `dev` `684e41e`)

- **Production**: `internal/cli/cli.go` — the single `&agent.AgentLoop{…}` construction (`cli.go:699`) inside the ~47-line block `cli.go:699-745`, entangled with three `→ ui` refs (R-1). After the round the CLI builds the loop through **`deps.Dependencies.LoopFactory`** and holds the loop as the domain **`agentport.Loop`**.
- **Domain**: `internal/domain/agent` gains `Loop` (interface), `LoopSpec` (value), `LoopFactory` (func type).
- **Adapter**: `internal/agent` gains `NewLoop(spec agentport.LoopSpec) agentport.Loop`.
- **Composition root**: `cmd/tellme/deps.go` binds `LoopFactory: func(spec agentport.LoopSpec) agentport.Loop { return agent.NewLoop(spec) }`.
- **Test surface (Q4)**: `internal/cli`'s fixture (`defaultTestDeps`) binds `LoopFactory` to an in-package **fake `agentport.Loop`**; the real-loop coverage stays in `internal/agent` + the godog E2E.

## 4. Invariants the plan must preserve

- `make verify` green: RULE-A/B/C **0**; RULE-E baseline **1** (`internal/cli -> internal/ui`), **0 new / 0 stale**; **0** cycles; **RULE-F** `→ ui` = 19 identifiers (0 new / 0 stale) and **no** `→ agent` key; cross-compile 4/4.
- Zero behavioural change (byte-contracts, exit codes, flags, DSL vocabulary); no new dependency; no new Makefile target.
- `internal/domain/**` 100 % pure (RULE-C): `Loop`/`LoopSpec`/`LoopFactory` reference stdlib + domain only.
- No cycle: `internal/agent` (tier 4) → `internal/domain/agent` (tier 0) is downward.
- **Both** ratchet removals land (RULE-E line + RULE-F key — ADR 0018 fold F-4).

## 5. Handoff

- `/axb-api-plan` → **NOOP** (no `contracts/**`).
- `/axb-data-plan` → **NOOP** (no state change).
- `/axb-dsl-refine` → **NOOP** (carried-forward CLI end; no contract change).
- Next: `/axb-tasks` → `/axb-implement` (TDD over the seam; the gate is the witness; the three falsifiability witnesses reproduced then reverted).
