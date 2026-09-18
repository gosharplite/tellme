# Tasks: yield-policy owner + `LoopObserver` hook split (round 045)

**Plan Package**: `specs/plans/045-yield-policy-owner`
**Anchor**: [#105](https://github.com/gosharplite/tellme/issues/105) — R3 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> `/axb-implement` execution control plane. Read `spec.md` + `truth-delta.md` before starting. This is a **non-BDD structural round** (no truth feature/DSL row changes), so Phase 3 carries `[UNIT]` ordering pins and a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases — the round-042/043/044 non-BDD precedent.

## Core Inputs

- `spec.md` (Locked decisions **C-R3-1 … C-R3-6**; US1–US3; FR-001…FR-012)
- `research.md` (D1–D9)
- `plan.md` (0 interfaces; `/axb-api-plan` + `/axb-data-plan` + `/axb-dsl-refine` NOOP)
- `truth-delta.md` (techstack MODIFY ×2; ADR 0014)
- `specs/truth/techstack.md` — the **Agent tool loop** row + the **Turn progress spinner** row
- `specs/truth/features/cli/chat/dsl.md` — the round-035/036 notes (grep evidence that no row goes stale)
- `docs/decisions/0005-tool-call-log-parity.md` (**D1** — amended by reference, **not** edited) · `docs/decisions/0009-…` · `docs/decisions/0010-…` · `docs/decisions/0011-…`
- `tools/arch/baseline.txt` (must stay **1** — R4's DoD)

## Setup

_(omitted — stdlib-only; no new technology; `go.mod`/`go.sum` unchanged)_

## Phase 2 — Foundational

- [X] **T001** — Create `internal/ui/yield.go`: the `YieldController` owner. **只做**：the type, `NewYieldController(*Spinner)`, `Enabled()`, `Yield()`/`Restore()`/`Admit()` (nil-safe), and the **single** policy doc comment. **不做**：no callers, no change to `spinner.go`/`coordinator.go`, no new behaviour.
- [X] **T002** — Land the test skeletons: an empty `internal/ui/yield_test.go`; the `composite_observer_test.go` harness re-typed to the new vocabulary (compiles only after T003). **只做**：landing files. **不做**：assertions yet (they are T003).

## Phase 3 — Test Alignment & Implementation (unit-only)

> No DSL rows: the round changes no feature step. The "alignment" is the harness re-typing; the "RED" is the ordering/route pins failing before the split.

- [X] **T003 `[UNIT-RED]`** — Rewrite `internal/cli/composite_observer_test.go`'s `recordingSpinner` to the split vocabulary and pin the orderings: `[spinner.YieldIndicator, call.OnCallEnd]` (non-final Y2), `[call.OnCallEnd]` (final), and the nil-safe no-op. **RED**: `compositeObserver`/`recordingSpinner` do not yet expose `YieldIndicator`/`RestoreIndicator`.
  - **DSL 參照**: n/a (no DSL row). **Boundary**: test file only.
- [X] **T004 `[UNIT-RED]`** — New `internal/ui/yield_test.go`: pin that `YieldController` is the single route — a recording `*Spinner` counts `Yield`/`Restore`/`Admit`, `Enabled()` is false for `{nil}`, and every method is a nil-safe no-op. **RED**: `YieldController` has no callers yet (and the coordinator still holds a bare `*Spinner`).
  - **Boundary**: `internal/ui/yield_test.go` only.
- [X] **T005** — Review gate (subagent review): the split is complete, no `BeforeToolLog`/`AfterToolLog` remains in production code (only explanatory notes), and the ordering pins are non-vacuous.

## Phase 4 — Green / Refactor

- [X] **T006 `[GREEN]`** — The split + the owner wiring (one atomic compile-breaking change): `internal/domain/agent/observer.go` replaces `BeforeToolLog`/`AfterToolLog` with `YieldIndicator`/`RestoreIndicator` (+ the route/policy doc); `internal/ui/spinner.go` implements the pair by delegating to `YieldController`; `internal/cli/composite_observer.go` implements the pair and routes the tail yield through `YieldIndicator()`; `internal/agent/agentloop.go` `withToolLog` calls the new pair; `internal/ui/coordinator.go` holds a `YieldController` and routes its clears/resumes/admits through it.
  - **Boundary**: only the files named above (+ the port). No behaviour change.
- [X] **T007 `[REFACTOR]`** — Sweep the yield rationale out of H1/H2/H3 into the owner + the port doc; update the `coordinator.go` header (drop the "#69 ledger item / still has these homes" text) and the `spinner.go` doc; leave exactly **one** policy statement.
- [X] **T008 `[CODE-REMOVE]`** — Delete the old pair (`Spinner.BeforeToolLog`/`AfterToolLog`, the composite's old methods) and the coordinator's `clearIndicator`→`Stop()` yield route (a yield now calls `Yield()`; `Stop()` stays only at the turn teardown).
- [X] **T009 `[REGRESSION]`** — `make verify` (gate **0 new / 0 stale**, baseline **1**) · `go test -count=1 ./...` green · **`go test -race ./internal/ui/...`** green · topology audit unchanged · `gofmt`/`go vet` clean. **Falsifiability witnesses** (a)/(b)/(c) reproduced then reverted. **Boundary**: verification only.

## Phase 5 — Truth, governance, close-out

- [X] **T010** — Land the truth + governance: `specs/truth/techstack.md` (the two MODIFY rows), `docs/decisions/0014-yield-policy-owner.md` + the index row, `tasks.md` `[X]`, `STATUS.md` + the daily summary, and open the PR.

## Pre-Delivery Orphan Coverage Sweep

| Artifact | Carrier |
| --- | --- |
| `truth-delta` non-NOOP: techstack **Agent tool loop** row | T010 (landed) + T007 (comment sweep) |
| `truth-delta` non-NOOP: techstack **Turn progress spinner** row | T001/T006/T010 |
| `research.md` D1–D9 | T001 (D2), T003/T004 (D2/D3/D7), T006 (D3/D4), T007 (D5), T009 (D7/D9), T010 (D5/D6) |
| `truth-delta` governance: ADR 0014 + index | T010 |
| `chat/dsl.md` stale-row check (D6) | T010 (grep evidence recorded in `truth-delta.md`) |
| Baseline stays **1** (C-R3-4) | T009 (gate) |

Orphans: **0**.
