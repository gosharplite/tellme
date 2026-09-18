# Tasks: round 051 — close #101 (de-couple `internal/cli` from `internal/ui` + F-6/F-7/F-8)

**Plan Package**: `specs/plans/051-cli-ui-decoupling`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — the terminal R5 slice; **baseline 1 → 0**.
**Created**: 2026-09-18 · **Skill**: `/axb-tasks`
**Decisions**: **Q1 → D** (one round closes #101) · **Q2 → (i)** (value types → existing domain peers) · **Q3 → (ii)** (narrow lifecycle-grouped ports) · **Q4 → (A)** (F-8 interface `OutputSink`; F-7 named `Discovery`; F-6 narrow seams).
**Witness**: the gate + the F-6/F-7/F-8 acceptances + unit/E2E pins — not the E2E suite alone (#92 AC5).
**Ordering (D2)**: values → lines port → indicator/tool-output ports → F-8 → F-7 → F-6 → gate removals.

> **Setup**: **omitted** — stdlib-only; no new dependency; no new Makefile target.

---

## Phase 1 — Foundational (domain value types + ports; must precede the CLI change)

- [x] **T001** `internal/domain/llm` — add `Pricing{Hit, Miss, Comp float64}` + pure `ComputeCost(pricing, miss, hit, completion, thinking int) float64` + `HitRate(hit, miss int) float64` (moved verbatim from `internal/ui`; RULE-C-pure).
- [x] **T002** `internal/domain/metrics` — add `UsageCounts{Miss, Hit, Completion, Thinking int}` (moved verbatim).
- [x] **T003** `internal/domain/history` — fold `ToolUsageRow{Tool string; OK, Error, Timeout int}` onto `ToolUsageCounts` (one concept): add the `Tool`/`Error`/`Timeout` shape needed by the report, or alias to the existing counts; export a `ToolUsageRow` equivalent + `Total()`.
- [x] **T004** `internal/domain/render` (NEW) — declare the pure **lines port**: an interface with the 8 `Format*` methods (`InputCaptured(t)`, `TurnOpening(turn, mode)`, `TurnGap()`, `PayloadStatus(t, tokens, budget, mode, model, estimated)`, `Metrics(t, provider, UsageCounts)`, `Ready(...)`, `ToolReason(t, reason)`, `ToolUsage([]history.ToolUsageRow)`) + a `DefaultToolOutputIdleGap` domain value. (RULE-C-pure; no `ui` type.)
- [x] **T005** `internal/domain/render` — declare the **`ProgressIndicator`** port (domain interface: the spinner lifecycle the CLI drives) and the **`ToolOutput`** port (`Begin()`, `Writer() io.Writer`, `End()`).
- [x] **T006** unit tests for the domain value types + ports (shape/compile pins; RULE-C purity via the gate).

## Phase 2 — Adapters (internal/ui implements the ports; no CLI change yet)

- [x] **T007** `internal/ui` — implement the `render.Lines` port (a `Lines` adapter delegating to the existing `Format*` funcs — the bytes stay here); implement `ProgressIndicator` (wrapping `Spinner`) and `ToolOutput` (the coordinator satisfies it — F-8).
- [x] **T008** `internal/ui` — make `ui.ToolOutputCoordinator` satisfy `domaintools.OutputSink` **directly** (F-8): the interface `{Begin(); Writer() io.Writer; End(); Enabled() bool}`; delete the struct-of-funcs use on the sink path.
- [x] **T009** `internal/ui` — implement the `agentport.ToolLineRenderer` + the answer renderer as before (already interface-seamed); expose constructors for the deps factories.

## Phase 3 — Test alignment (RED-first; the CLI/deps seam changes)

- [x] **T010** `internal/app/deps` — add the new seams: `Lines`/`NewProgressIndicator`/`NewToolOutput`/`NewAnswerRenderer` factories (func- or interface-typed as chosen), and the F-7 **`Discovery`** type + `MCPDiscoverer` signature change.
- [x] **T011** `internal/cli` — **RED**: replace every `ui.*` reference with the domain types/ports (`domain/llm.Pricing`/`ComputeCost`/`HitRate`, `domain/metrics.UsageCounts`, `domain/history.ToolUsageRow`, the `render.*` ports + the deps factories); F-6 narrow the seams. Compile fails until T010 lands (the intended RED).
- [x] **T012** `internal/cli` tests — adopt **domain-typed fakes** for the new ports (no `_test.go` import of `internal/ui` — RULE-E merged-graph).
- [x] **T013** `cmd/tellme` — wire the new factories/ports (tier-exempt); the `Discovery` result (F-7).

## Phase 4 — The two ratchet removals (the DoD)

- [x] **T014** `tools/arch/baseline.txt` — regenerate so the `internal/cli -> internal/ui` line is **removed** (baseline **0** — header-only). **Removal #1.**
- [x] **T015** `tools/arch/arch_test.go` — remove the `couplingSurface["internal/cli -> internal/ui"]` key (the 19 identifiers) + update the synthetic coverage self-test; the coverage invariant stays green. **Removal #2.**

## Phase 5 — Verification & regression

- [x] **T016** `make verify` — green: RULE-A/B/C **0**; RULE-E baseline **0**, 0 new / 0 stale; **0** cycles; RULE-F coverage green with no `→ ui` key; cross-compile 4/4.
- [x] **T017** `go test -count=1 ./...` green (incl. the godog E2E); `gofmt`/`go vet` clean; identifier count: `internal/cli` production `→ ui` refs = **0**.
- [x] **T018** F-6/F-7/F-8 acceptances: (F-6) every function receiving `Dependencies` reads ≤2 fields — pinned by a unit audit; (F-7) a named `Discovery`; (F-8) `OutputSink` is an interface.
- [x] **T019** Falsifiability witnesses (reproduced then reverted, ADR 0010): (a) compile-level; (b) count-level (stray `ui.*` ⇒ RULE-E + RULE-F red); (c) baseline-level (the removal reds twice until both halves fold); (d) F-8 (a struct `OutputSink` ⇒ the interface acceptance reds).
- [x] **T020** confirm `go.mod`/`go.sum` unchanged; no new Gherkin/DSL row; topology audit unchanged; truth + ADR + the round's `truth-delta.md` updated.

---

## Traceability

| Requirement | Tasks |
| --- | --- |
| FR-001 (CLI `→ ui` refs = 0) | T001–T013 |
| FR-002 (RULE-C ports, downward) | T004/T005/T007/T009/T016 |
| FR-003 (behaviour-preserving) | T011/T012/T017 |
| FR-004 (deps injection) | T010/T013 |
| FR-005 (two ratchet removals) | T014/T015/T016 |
| FR-006/F-6 | T010/T011/T018 |
| FR-007/F-7 | T010/T013/T018 |
| FR-008/F-8 | T008/T018 |
| FR-009 (ADR 0020 + truth) | T020 (+ research/ADR landed) |
| FR-010 (no dep/DSL) | T020 |
| FR-011 (witnesses) | T019 |
| FR-012 (no re-open) | all |

**Pre-Delivery Orphan Sweep**: T001–T020 each map to a requirement or the DoD; **0** orphans.
