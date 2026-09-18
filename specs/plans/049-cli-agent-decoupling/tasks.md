# Tasks: de-couple `internal/cli` from the turn loop — re-cut sub-slice 1 (round 049)

**Plan Package**: `specs/plans/049-cli-agent-decoupling`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) — R5.3 of [#92](https://github.com/gosharplite/tellme/issues/92) (re-cut sub-slice 1 of the `cli → agent` de-coupling)
**Created**: 2026-09-18

> `/axb-implement` execution control plane. Read `spec.md` + `research.md` + `plan.md` + `truth-delta.md` before starting. This is a **non-BDD structural round** (no truth feature/DSL row changes), so Phase 3 carries `[UNIT-ALIGN]`/`[UNIT-RED]` + a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases — the round-042…048 non-BDD precedent.

## Core Inputs

- `spec.md` (Locked **Q1 → B** / **Q2 → (i)** / **Q3 → (a)**; US1–US3; FR-001…FR-012)
- `research.md` (D1–D10)
- `plan.md` (0 interfaces; no waves; `/axb-api-plan` + `/axb-data-plan` + `/axb-dsl-refine` NOOP)
- `truth-delta.md` (techstack MODIFY ×2 + the Task-runner NOOP; **ADR 0018**)
- `specs/truth/techstack.md` — the **Layer-discipline gate** row (re-cut recorded; baseline **unchanged 2**) + the **Agent tool loop** row
- `docs/decisions/0018-cli-agent-contracts-extraction.md` (the re-cut) · `0017` (§Forward classification + re-cut recipe) · `0016` (RULE-E) · `0011` (tier table D1, ratchet D3) · `0015` (the port-in-domain precedent)
- `tools/arch/baseline.txt` (**must stay byte-identical** — still 2 lines; this round moves no edge)
- `internal/agent/agentloop.go` (`AgentLoop`, `Run`, the local `AgentResult`/`ErrIncomplete`/`ToolDefs` — the declarations being relocated)
- `internal/domain/agent/{observer,call,presenter}.go` (the existing port family the new contracts join)

## Setup

_(omitted — stdlib-only; no new technology; `go.mod`/`go.sum` unchanged)_

## Phase 2 — Foundational

- [X] **T001** — Create `internal/domain/agent/result.go`: relocate the three crossing contracts **verbatim** into the domain package (`package agent`) — the turn-result type `Result{Answer string; Steps []history.Step; Usage llm.Usage; Calls []llm.Usage}`, the incomplete-turn error `ErrIncomplete{Reason string; Err error}` + `Error()`/`Unwrap()`, and the pure projection `func ToolDefs(reg tools.Registry) []llm.ToolDef`. Imports: `domain/history` + `domain/llm` + `domain/tools` (domain-only — **RULE-C-pure**). Doc states the home (peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`) + the round-049 rationale. **只做**：the contract declarations (verbatim copies of the `internal/agent` originals). **不做**：no `internal/agent` edits, no consumers, no behaviour change.
- [X] **T002** — Land `internal/domain/agent/result_test.go` as an independent landing file (the T008 pin location). **只做**：the landing file for the behaviour pins. **不做**：no pins yet, no product edits.

## Phase 3 — Test Alignment & Implementation (unit-only)

> No DSL rows: the round changes no feature step. The "alignment" is repointing `internal/agent`'s tests to the domain contracts; the "RED" is the **compile break** induced by deleting the local declarations before the consumers are repointed (proving the contracts now have exactly **one** home).

- [X] **T003 `[UNIT-ALIGN]`** — Repoint `internal/agent/{agentloop_test.go,callhooks_test.go,usage_calls_test.go}` to the **domain** contracts (`agentport.Result` / `agentport.ErrIncomplete` / `agentport.ToolDefs`) — import `agentport "…/internal/domain/agent"`; the asserted behaviour is **unchanged**. **只做**：the three test files. **不做**：no `agentloop.go` edit yet.
  - **Boundary**: `internal/agent/*_test.go` (3 files) only.
- [X] **T004 `[UNIT-RED]`** — **Delete** the local `AgentResult`, `ErrIncomplete` (+ its two methods), and `ToolDefs` declarations from `internal/agent/agentloop.go`. **RED**: `go build ./...` fails — `agentloop.go`, `internal/cli/{cli,call_renderer}.go`, and the T003-repointed agent tests reference the now-deleted local symbols; this is the machine evidence that the contracts have exactly one home.
  - **Boundary**: `internal/agent/agentloop.go` only.
- [X] **T005** — **Review gate** (subagent): confirm the RED is non-vacuous (a real compile error naming the deleted symbols, not a missing-file artefact); confirm T001 relocated the contracts **verbatim** (field names/types, `Error`/`Unwrap` semantics, `ToolDefs` body) with **no** alias left behind; confirm scope = the 3 contracts only (no other `internal/agent` symbol moved).

## Phase 4 — Green / Refactor

- [X] **T006 `[GREEN]`** — `internal/agent/agentloop.go`: repoint the loop to the domain contracts — `func (a *AgentLoop) Run(...) (agentport.Result, error)`, the four `return AgentResult{…}` sites → `agentport.Result{…}`, the three `&ErrIncomplete{…}` sites → `&agentport.ErrIncomplete{…}`, and `a.toolDefs()`/`ToolDefs(...)` → `agentport.ToolDefs(...)`. Build green for the `agent` package. **做**：the reference+delete. **不做**：no behaviour change, no alias.
  - **Boundary**: `internal/agent/agentloop.go` only.
- [X] **T007 `[GREEN]`** — `internal/cli/call_renderer.go`: **drop** the `internal/agent` import; add `agentport "…/internal/domain/agent"`; `agent.ToolDefs(r.reg)` → `agentport.ToolDefs(r.reg)`; the `persistTurnUsage(… result agent.AgentResult …)` parameter → `agentport.Result`. **不做**：no formatting/estimate change (the pre-flight estimate stays byte-identical).
  - **Boundary**: `internal/cli/call_renderer.go` only.
- [X] **T008 `[GREEN]`** — `internal/cli/cli.go`: `var inc *agent.ErrIncomplete` → `*agentport.ErrIncomplete` (the `agentport` import already exists at `cli.go:24`); **keep** `&agent.AgentLoop{…}` (the surviving, baselined coupling). After this task `internal/cli` production references **exactly one** `internal/agent` identifier (`agent.AgentLoop`). **不做**：no construction/behaviour change.
  - **Boundary**: `internal/cli/cli.go` only.
- [X] **T009 `[REFACTOR]`** — `internal/domain/agent/result_test.go`: add the behaviour-preservation pins — `ToolDefs` (nil registry → nil; one `llm.ToolDef` per registered tool in **registration order**; `Name`/`Description`/`Parameters` mapping); the `Result` field set (compile-checked via a literal); `ErrIncomplete` (`Error()` with/without an `Err`; `Unwrap()`; `errors.As` round-trip). Sweep the moved docs for stale `agent.` citations.

## Phase 5 — Regression, witnesses, close-out

- [X] **T010 `[REGRESSION]`** — `make verify` **green** with `tools/arch/baseline.txt` **byte-identical** (RULE-E **0 new / 0 stale**, baseline **2**; RULE-A/B/C **0**; **0** cycles; lint 0; govulncheck clean; cross-compile 4/4) · `go test -count=1 ./...` green (incl. the godog E2E) · the Gherkin/DSL topology audit **unchanged** · `gofmt`/`go vet` clean. **Identifier-count check**: `grep` shows `internal/cli` production references exactly **one** `internal/agent` identifier (`agent.AgentLoop`), and `internal/agent` declares **none** of `AgentResult`/`ErrIncomplete`/`ToolDefs` (no alias). **Falsifiability witnesses** reproduced then reverted (ADR 0010): (a) leave `agent.ToolDefs` referenced in `internal/cli` ⇒ the count is **2** (incomplete extraction); (b) flip a `Result` field name in the domain contract ⇒ the CLI/loop seam **fails to compile**; (c) a partial extraction leaving a second `agent.*` reference ⇒ rejected by the count check. **Boundary**: verification only.
- [X] **T011** — Land/confirm truth + governance: `specs/truth/techstack.md` (the two MODIFY rows), confirm `docs/decisions/0018-cli-agent-contracts-extraction.md` + the index row, mark `tasks.md` `[X]`, update `STATUS.md` + the daily summary, and open the PR. Record **sub-slice 2** (the construction inversion; baseline **2 → 1**) + the `→ ui` edge + F-6/F-7/F-8 on [#101](https://github.com/gosharplite/tellme/issues/101), and the `Options`/`Dependencies` interface-seam `Validate()` caveat (ADR 0017 §Forward) for sub-slice 2.

## Parallel Hint

- T006 / T007 / T008 touch **disjoint files** (`internal/agent/agentloop.go` · `internal/cli/call_renderer.go` · `internal/cli/cli.go`) and may be dispatched once T006 lands the domain types in the loop; the CLI tasks (T007/T008) depend only on T001.

## Pre-Delivery Orphan Coverage Sweep

| Artifact | Carrier |
| --- | --- |
| `truth-delta` non-NOOP: techstack **Layer-discipline gate** row (re-cut; baseline **unchanged 2**) | T010 + T011 |
| `truth-delta` non-NOOP: techstack **Agent tool loop** row (contracts domain-owned) | T001 + T006 + T009 + T011 |
| `research.md` D1–D10 | T001 (D2/D4), T003 (D3), T004 (D1), T006 (D2/D3/D5), T007 (D2), T008 (D10), T009 (D5), T010 (D7/D9), T011 (D6) |
| `truth-delta` governance: ADR 0018 + index | T011 (recorded — already landed in the research phase) |
| **Q2-(i)** all three contracts extracted (FR-001) | T001 + T006 + T007 + T008 + T010 (count==1) |
| **Q3-(a)** reference + delete, no alias (FR-004) | T004 + T006 + T010 (no-alias check) |
| **CLI `→ agent` count == 1** (FR-005, SC-001) | T008 + T010 |
| Baseline **byte-identical** (FR-006) | T010 |

## Notes

- The round is **behaviour-preserving** — the turn path's `stdout`/`stderr` byte-contracts, flags, exit codes, and the pre-flight estimate are unchanged; the E2E suite is **regression**, not the carrier (NFR-004).
- **The baseline does NOT move** this round (the `AgentLoop` construction keeps the `→ agent` edge); `tools/arch/baseline.txt` is **byte-identical**. The baseline-moving round is **sub-slice 2** (**2 → 1**).
- **RULE-A shapes the homes**: the contracts live at tier 0 (`internal/domain/agent`); `internal/agent` (tier 4) → domain is downward; `internal/cli` (tier 6) → domain is sanctioned under RULE-E.
- **No new Makefile target**; **no new dependency** (`go.mod`/`go.sum` unchanged); **no** `tools/arch` change.

## Outcome (implementation delivered)

- **Product**: `internal/domain/agent/result.go` (**NEW** — the relocated `Result` + `ErrIncomplete` + `ToolDefs`; RULE-C-pure) · `internal/agent/agentloop.go` (CHANGED — the local declarations **deleted**; `Run` returns `agentport.Result`; `agentport.ErrIncomplete` / `agentport.ToolDefs`) · `internal/cli/call_renderer.go` (CHANGED — `internal/agent` import **dropped**; `agentport.ToolDefs` / `agentport.Result`) · `internal/cli/cli.go` (CHANGED — `*agentport.ErrIncomplete`; the surviving `&agent.AgentLoop{…}` kept) · `internal/domain/agent/result_test.go` (**NEW** — the behaviour pins) · `internal/agent/{agentloop,callhooks,usage_calls}_test.go` (CHANGED — repointed to the domain contracts).
- **Contract count**: `internal/cli` production references **exactly one** `internal/agent` identifier — `agent.AgentLoop` (**4 → 1**); `internal/agent` declares **none** of the three contracts (no alias).
- **Gate**: `make verify` **OK** — `verify-architecture` reports RULE-E **0 new / 0 stale** with the baseline **byte-identical** at **2** (`internal/cli -> internal/agent`, `internal/cli -> internal/ui` — **unchanged**, as designed); RULE-A/B/C **0**; **0** cycles; lint 0 issues; govulncheck clean; cross-compile **4/4**.
- **Tests**: `go test -count=1 ./...` green (incl. the godog E2E); `internal/domain/agent` carries the 3 relocation pins (`TestResultFieldSet`, `TestErrIncomplete`, `TestToolDefs`).
- **Falsifiability witnesses (reproduced then reverted, ADR 0010)**: (a) re-adding `agent.ToolDefs` in `internal/cli` raises the identifier count to **2** (and fails to build — the symbol is gone); (b) flipping a domain `Result` field (`Answer`→`AnswerX`) **breaks the loop/CLI seam at compile time**; (c) a partial extraction is caught by the count check.
- **No behaviour change**: `stdout`/`stderr` contracts, flags, exit codes, and the pre-flight estimate are unchanged; `go.mod`/`go.sum` unchanged; `tools/arch/baseline.txt` byte-identical.

