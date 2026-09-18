# Tasks: round 050 — invert the `AgentLoop` construction/execution into an injected domain port

**Plan Package**: `specs/plans/050-agentloop-construction-inversion`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — sub-slice 2; the **baseline-moving** round (**2 → 1**).
**Created**: 2026-09-18 · **Skill**: `/axb-tasks`
**Decisions**: **Q1 → A** (port-only inversion; the `ui` wiring stays in the CLI) · **Q2 → (i)** (domain `Loop` interface + `LoopSpec` + a func-typed factory; adapter `agent.NewLoop`) · **Q3 → (a)** (the CLI keeps supplying `Lines` + the observer, domain-typed) · **Q4 → A** (the CLI test surface adopts an in-package fake `agentport.Loop`; RULE-E is merged-graph).
**Witness**: the gate (`make verify`) + the identifier-count check + the existing turn unit pins — **not** the E2E suite (#92 AC5).

> **Setup**: **omitted** — stdlib-only, no new dependency, no scaffolding, no new Makefile target (the gate rides `verify-architecture`).
> **Foundational**: see below (the domain port must exist before the CLI can consume it).

---

## Phase 1 — Foundational (the domain port + the adapter; must precede the CLI change)

- [x] **T001** `internal/domain/agent` — declare the loop port (RULE-C-pure; stdlib + domain only):
  - `type Loop interface { Run(ctx context.Context, prompt string, prior []history.Entry) (Result, error) }`
  - `type LoopSpec struct { Gateway llm.Gateway; Registry tools.Registry; MaxLoops int; Stderr io.Writer; Now func() time.Time; EffectiveBudget int; ToolUsage history.ToolUsageSink; Lines ToolLineRenderer; Observer LoopObserver }`
  - `type LoopFactory func(LoopSpec) Loop`
  - Package-doc note: this is the loop's domain **construction/execution** contract (peer of the observer/renderer ports).
- [x] **T002** `internal/domain/agent` — unit test: the port is a pure contract (compile-time `var _ Loop …`; a `LoopSpec` round-trip of a fake `Loop` proves the factory signature). (RULE-C-purity is verified by the gate.)
- [x] **T003** `internal/agent` — add the adapter `func NewLoop(spec agentport.LoopSpec) agentport.Loop { return &AgentLoop{Gateway: spec.Gateway, Registry: spec.Registry, MaxLoops: spec.MaxLoops, Stderr: spec.Stderr, Now: spec.Now, EffectiveBudget: spec.EffectiveBudget, ToolUsage: spec.ToolUsage, Lines: spec.Lines, Observer: spec.Observer} }`. `internal/agent` keeps `AgentLoop`/`Run` **unchanged** (it already satisfies `agentport.Loop`).

## Phase 2 — Test alignment (RED-first; no production change yet)

- [x] **T004** `internal/app/deps` — add the func-typed field `LoopFactory agentport.LoopFactory` (with a doc note: the loop's construction is domain-typed; a func-typed field is covered by `Validate()`'s `Kind()==reflect.Func` predicate — no interface-seam assertion needed). Assert `Validate()` still reports the first unbound func seam.
- [x] **T005** `internal/cli` — **RED**: change `runTurn` to build the loop through `spec := agentport.LoopSpec{…}; loop := dp.LoopFactory(spec)` (replacing `&agent.AgentLoop{…}`); the CLI's only loop type is `agentport.Loop`. Compile fails until the deps/fixture gain `LoopFactory` (the intended RED).
- [x] **T006** `internal/cli` test fixture — add `LoopFactory` to `defaultTestDeps` bound to an **in-package fake `agentport.Loop`** (Q4 → A). The fake is script-driven so a `runTurn` test can assert CLI orchestration (Result/error handling, persistence, stream order).
- [x] **T007** `internal/cli` — rework the three loop-dependent `runTurn` test files (`turn_test.go`, `persistence_invariant_test.go`, `tui_submit_chrome_test.go`) from "real loop + fake gateway" to "fake loop": assert **CLI orchestration** (what `runTurn` does with the `Result`/error), not loop internals. (`composite_observer_test.go` already tests the observer in isolation — unchanged.)
- [x] **T008** `cmd/tellme/deps.go` — bind `LoopFactory: func(spec agentport.LoopSpec) agentport.Loop { return agent.NewLoop(spec) }`; update the composition-root smoke test if it asserts the seam set.
- [x] **T009** `internal/cli/cli.go` — remove the `internal/agent` import; delete the `&agent.AgentLoop{…}` literal; the loop is `agentport.Loop` from `dp.LoopFactory(spec)`. (Production references to `internal/agent` → **0**.)

## Phase 3 — Gate alignment (the two ratchet removals — the DoD)

- [x] **T010** `tools/arch/baseline.txt` — regenerate (`make verify-architecture-update`) so the `internal/cli -> internal/agent` line is **removed**; the file is exactly **1** line (`internal/cli -> internal/ui`) + the header. **RULE-E removal #1.**
- [x] **T011** `tools/arch/arch_test.go` — remove the `couplingSurface["internal/cli -> internal/agent"]` key (`{AgentLoop}`); keep the `→ ui` entry (19 identifiers) unchanged; the RULE-F coverage invariant must stay green (the remaining governed edge is still surface-tracked). **RULE-E removal #2** (the RULE-F key — no regeneration affordance, ADR 0018 D6).

## Phase 4 — Verification & regression

- [x] **T012** `make verify` — green: RULE-A/B/C **0**; RULE-E baseline **1**, 0 new / 0 stale; **0** cycles; **RULE-F** `→ ui` = 19 (0 new / 0 stale), no `→ agent` key; lint 0; govulncheck clean; cross-compile 4/4.
- [x] **T013** `go test -count=1 ./...` — green (incl. the godog E2E); `gofmt`/`go vet ./...` clean; identifier-count check: an `internal/cli` production scan finds **0** `internal/agent` references.
- [x] **T014** **Falsifiability witnesses** (reproduced then reverted, ADR 0010): (a) **compile-level** — delete/rename the `Loop` seam ⇒ the CLI turn path fails to compile; (b) **count-level** — reintroduce a direct `agent.*` selector (or a second construction) in the CLI production sources ⇒ RULE-E/RULE-F **FAILS** while the `→ ui` surface stays green; (c) **baseline-level (two-removal)** — with only ONE of T010/T011 applied, the gate reds once (the un-folded half); with both, green. Revert every witness.
- [x] **T015** confirm `go.mod`/`go.sum` unchanged; no new Gherkin/DSL row; the topology audit unchanged.

---

## Traceability

| Requirement | Tasks |
| --- | --- |
| FR-001 (port + CLI refs = 0) | T001, T003, T005, T009 |
| FR-002 (RULE-C purity; adapter tier) | T001, T002, T003, T012 |
| FR-003 (behaviour-preserving) | T007, T013 |
| FR-004 (deps injection) | T004, T008 |
| FR-005 (two ratchet removals; `→ ui` untouched) | T010, T011, T012 |
| FR-006 (regeneration vs guard-table edit) | T010, T011 |
| FR-007 (ADR 0019) | (ADR landed in `/axb-technical-research`) |
| FR-008 (truth MODIFY) | (truth landed in `/axb-technical-research`) |
| FR-009 (`Validate()` seam) | T004 |
| FR-010 (no new dep/DSL) | T015 |
| FR-011 (witnesses) | T014 |
| FR-012 (no re-open) | all |
| NFR-001 (no new CLI coupling) | T009, T012, T013 |
| NFR-002 (atomic single PR) | all |

**Pre-Delivery Orphan Sweep**: T001–T015 each map to a requirement or the round's DoD; **0** orphans.
