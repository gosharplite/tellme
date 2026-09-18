# Tasks: blank-reason owner + the loop's presentation de-coupling (round 046)

**Plan Package**: `specs/plans/046-blank-reason-owner-and-presentation-decoupling`
**Anchor**: [#108](https://github.com/gosharplite/tellme/issues/108) — R4 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> `/axb-implement` execution control plane. Read `spec.md` + `research.md` + `truth-delta.md` before starting. This is a **non-BDD structural round** (no truth feature/DSL row changes), so Phase 3 carries `[UNIT]` pins and a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases — the round-042/043/044/045 non-BDD precedent.

## Core Inputs

- `spec.md` (Locked decisions **C-R4-1 = A** / **C-R4-2 = delete the dead site** / **C-R4-3 = unit pins + the gate** / **C-R4-4 = ADR 0015**; US1–US3; FR-001…FR-011)
- `research.md` (D1–D10)
- `plan.md` (0 interfaces; `/axb-api-plan` + `/axb-data-plan` + `/axb-dsl-refine` NOOP)
- `truth-delta.md` (techstack MODIFY ×2; ADR 0015)
- `specs/truth/techstack.md` — the **Agent tool loop** row + the **Layer-discipline gate** row
- `docs/decisions/0015-loop-presentation-port.md` (the policy) · `0005` (D1 reaffirmed) · `0013` (wiring) · `0014` (yield — untouched)
- `tools/arch/baseline.txt` (must regenerate **header-only (0)** — R4's DoD)

## Setup

_(omitted — stdlib-only; no new technology; `go.mod`/`go.sum` unchanged)_

## Phase 2 — Foundational

- [ ] **T001** — Create `internal/domain/agent/presenter.go`: the `ToolLineRenderer` interface (`EngineLine`, `ActionLine`, `ResultLine`, `ReasonLine(t, reason) (string, bool)`) + the route/policy doc comment. **只做**：the port declaration. **不做**：no `internal/ui` types, no callers.
- [ ] **T002** — Land the `internal/ui` adapter: `ui.ToolLineRenderer{}` over `FormatToolEngine`/`FormatToolAction`/`FormatToolResult` + a `ReasonLine` that evaluates `toolReasonText` **once** (returns `("", false)` for a blank/whitespace/escape-only reason, else the `[Tool Reason]` line). Land an empty `internal/ui/toolrenderer_test.go`. **只做**：the adapter + landing file.

## Phase 3 — Test Alignment & Implementation (unit-only)

> No DSL rows: the round changes no feature step. The "alignment" is the renderer pin + the loop-test fake; the "RED" is the compile failure + the scheduling pins.

- [ ] **T003 `[UNIT-RED]`** — Retarget the three `internal/agent` log tests (`agentloop_reason_test.go`, `agentloop_blank_reason_test.go`, `agentloop_spacing_test.go`) to inject an in-package **recording fake renderer** (they must NOT import `internal/ui` — the gate governs test imports) and assert the **schedule**: the four line kinds, their order, the per-call leading blank, and the blank-reason suppression *via the fake's contract*. **RED**: `agentloop.go` still imports `internal/ui` and has no `Lines` field.
  - **DSL 參照**: n/a (no DSL row). **Boundary**: `internal/agent/*_test.go` only.
- [ ] **T004 `[UNIT-RED]`** — `internal/ui/toolrenderer_test.go`: pin the port contract — `ReasonLine` returns `("", false)` for `""`/`"   "`/`"\n"`/`"\u001b[31m"`, and a line **byte-equal** to `FormatToolReason` for a real reason (incl. an over-cap reason), and `renders` ⇔ `ToolReasonRenders`. **RED**: the adapter's `ReasonLine` does not exist yet.
  - **Boundary**: `internal/ui/toolrenderer_test.go` only.
- [ ] **T005** — Review gate: the port + the fake are non-vacuous; the fake asserts the schedule (not the formatting), and the formatting pin lives at the `ui` tier.

## Phase 4 — Green / Refactor

- [ ] **T006 `[GREEN]`** — `internal/agent/agentloop.go`: drop `import "…/internal/ui"`; add the nil-safe `Lines agentport.ToolLineRenderer` field; route `logEngine`/`logAction`/`logResult` through it (the schedule + the per-call blank unchanged); `withToolLog` **unchanged** (ADR 0014); `reasonsOf` → a method that consults `Lines.ReasonLine` and appends the **raw** reason when it renders.
  - **Boundary**: `internal/agent/agentloop.go` only. No behaviour change.
- [ ] **T007 `[GREEN]`** — `internal/cli/cli.go` `runTurn`: inject `Lines: ui.ToolLineRenderer{}` on the loop (the composition site, ADR 0013). **Boundary**: `internal/cli/cli.go` only.
- [ ] **T008 `[REFACTOR]`** — `internal/cli/call_renderer.go`: **delete** the dead `ui.ToolReasonRenders` re-check in the `emit` closure (P3) — the tail prints the already-filtered `roundReasons`; update the comment to name the single owner (`ui.ToolLineRenderer.ReasonLine`). Sweep the `internal/agent` doc comments (the loop renders through the port; the predicate is single-owned).
- [ ] **T009 `[CODE-REMOVE]`** — Regenerate `tools/arch/baseline.txt` to **header-only** via `make verify-architecture-update` (the ratchet's terminal state; the gate ships with its enabler).
- [ ] **T010 `[REGRESSION]`** — `make verify` (gate **0 new / 0 stale / 0 cycles**, baseline **0**) · `go test -count=1 ./...` green · topology audit unchanged · `gofmt`/`go vet` clean. **Falsifiability witnesses** (a)/(b)/(c) reproduced then reverted. **Boundary**: verification only.

## Phase 5 — Truth, governance, close-out

- [ ] **T011** — Land the truth + governance: `specs/truth/techstack.md` (the two MODIFY rows), confirm `docs/decisions/0015-loop-presentation-port.md` + the index row, mark `tasks.md` `[X]`, update `STATUS.md` + the daily summary, and open the PR.

## Pre-Delivery Orphan Coverage Sweep

| Artifact | Carrier |
| --- | --- |
| `truth-delta` non-NOOP: techstack **Agent tool loop** row | T001/T002/T006/T008/T011 |
| `truth-delta` non-NOOP: techstack **Layer-discipline gate** row | T009/T011 |
| `research.md` D1–D10 | T001 (D1), T002 (D2), T006 (D3/D6/D7), T007 (D4), T008 (D5/D6), T009 (D8), T010 (D10) |
| `truth-delta` governance: ADR 0015 + index | T011 (recorded — already landed in the plan half) |
| `chat/dsl.md` stale-row check (D9) | T011 (grep evidence recorded in `truth-delta.md`) |
| Baseline reaches **0** (FR-004) | T009 + T010 (gate) |
