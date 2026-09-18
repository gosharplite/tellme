# System Analysis Plan: round 051 — close #101 (de-couple `internal/cli` from `internal/ui` + F-6/F-7/F-8)

**Plan Package**: `specs/plans/051-cli-ui-decoupling`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — the terminal R5 slice.
**Created**: 2026-09-18 · **Skill**: `/axb-system-analysis`

## 1. Interface inventory

Internal structural refactor of the `internal/cli` ↔ `internal/ui` seam + three seam refactors. **No** new system boundary.

| Interface | Kind | Classification | Planner |
| --- | --- | --- | --- |
| CLI end (`tellme` command) | `cli` | unchanged — no new Rule/Example/step/`DSLRow` | `/axb-dsl-refine` (**NOOP**) |
| API end | `backend` | none (no OpenAPI/HTTP surface) | `/axb-api-plan` (**NOOP**) |
| Data end | — | none (no state change) | `/axb-data-plan` (**NOOP**) |
| UI end | — | none (line-oriented CLI) | none |

## 2. Waves

**No delegated planner wave applies.** Per the CLI-streamlined workflow the `cli` end is carried forward to its contract owner `/axb-dsl-refine` (recorded **NOOP** — an internal refactor is not a CLI-contract change). api/data **NOOP**.

| Wave | Order | Delegation |
| --- | --- | --- |
| — | — | none |

## 3. The seam under change (grounded @ `dev` `310def4`)

1. **Value types** → `internal/domain/**` (Q2 → i): `Pricing`+`ComputeCost`/`HitRate` → `domain/llm`; `UsageCounts` → `domain/metrics`; `ToolUsageRow` → `domain/history` (folded onto `ToolUsageCounts`).
2. **Lines port** → a domain interface for the 8 `Format*` formatters + the idle-gap value; adapter in `internal/ui`.
3. **`ProgressIndicator` port** → a domain interface for the spinner lifecycle; adapter `ui.Spinner`.
4. **`ToolOutput` port** → a domain interface (Begin/Writer/End) the coordinator satisfies; used directly by F-8.
5. **Answer `Renderer`** + **`ToolLineRenderer`** → obtained via `deps` factories (already interface-seamed).
6. **F-8** — `domaintools.OutputSink` → interface; delete the struct bridge + the `BindToolOutput` rebind.
7. **F-7** — `deps.Discovery{Tools, Warnings, Closer io.Closer}`.
8. **F-6** — narrow the `Dependencies` seams; only `run`/`runTurn` hold the bag.
9. **Composition root** (`cmd/tellme/deps.go`) wires the new factories/ports (tier-exempt).
10. **Test surface** — CLI tests adopt **domain-typed fakes** for the new ports (RULE-E is merged-graph — the round-050 Q4 → A precedent).

## 4. Invariants the plan must preserve

- `make verify` green: RULE-A/B/C **0**; RULE-E baseline **0**, **0 new / 0 stale**; **0** cycles; **RULE-F** green with **no `→ ui` key**; cross-compile 4/4.
- Zero behavioural change (byte-contracts, exit codes, flags, DSL vocabulary); no new dependency; no new Makefile target.
- `internal/domain/**` 100 % pure (RULE-C); the ports carry domain types only.
- No cycle; downward imports (`internal/ui` tier 5 → `domain` 0; `internal/cli` tier 6 → `domain` 0).
- ui-owned bytes/caps/sanitize (ADR 0006/0008/0015) and the spinner epochs + coordinator lock order preserved.
- **Two ratchet removals** (the RULE-E line + the RULE-F `→ ui` key).

## 5. Handoff

- `/axb-api-plan` → **NOOP** · `/axb-data-plan` → **NOOP** · `/axb-dsl-refine` → **NOOP**.
- Next: `/axb-tasks` → `/axb-implement` (ordered per D2; the gate + the F-6/F-7/F-8 acceptances are the witness; witnesses reproduced then reverted).
