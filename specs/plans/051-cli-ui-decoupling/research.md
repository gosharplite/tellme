# Phase 0 Research: close [#101](https://github.com/gosharplite/tellme/issues/101) — de-couple `internal/cli` from `internal/ui` + the F-6/F-7/F-8 deferrals (Round 051)

**Plan Package**: `specs/plans/051-cli-ui-decoupling`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — the terminal R5 slice.
**Predecessors**: rounds 047 (R5.1 gate) · 048 (R5.2 `→ ui/tui/prompt`) · 049 (R5.3 contracts→domain) · 050 (R5.4 the `→ agent` inversion, baseline 2 → 1).

## Decision 1: One round closes #101 (Q1 → D)

Round 051 carries the **whole** remaining #101 scope: the `→ ui` de-coupling (**baseline 1 → 0**) **and** F-6/F-7/F-8. Internal ordering (D2) keeps the behaviour-preserving invariants. **Blast-radius caveat (recorded):** if a review judges the change exceeds a single reviewable PR, the **slices are ordered** (D2) so an internal re-cut still closes #101 across the ordered parts.

## Decision 2: The `→ ui` work is internally ordered (values → ports → objects)

The terminal RULE-E residual is **not edge-sized** (ADR 0017 §Forward). Applied recipe (ADR 0017/0018/0019): **(1)** re-home the crossing **value types** (Q2); **(2)** invert the pure **formatters** behind a domain **lines** port; **(3)** invert the two **lifecycle objects** (spinner / tool-output) behind domain **interfaces** and obtain the answer `Renderer` + `ToolLineRenderer` via **deps factories** (Q3); **(4)** narrow the `Dependencies` seams (F-6); **(5)** the `OutputSink` interface (F-8) and the named `Discovery` (F-7). Behaviour-preserving throughout.

## Decision 3: The three value types re-home into existing domain peers (Q2 → (i))

- **`Pricing`** (`{Hit, Miss, Comp float64}`) + the pure **`ComputeCost`**/**`HitRate`** → **`internal/domain/llm`**. Cost is a provider/model concern; `ComputeCost` feeds **persistence** (`usageRecordOf`), not only display — so it cannot live inside a *render* port.
- **`UsageCounts`** → **`internal/domain/metrics`**.
- **`ToolUsageRow`** → **`internal/domain/history`**, **folded onto** the existing `history.ToolUsageCounts` (one concept, one name).

**Siting rationale**: each type sits with its domain peers, adding **no** new package and **no** new domain cross-edge (contrast: a fresh `domain/pricing` would be over-structure for ~2 functions).

## Decision 4: Narrow lifecycle-grouped ports (Q3 → (ii))

Declared under `internal/domain/**`, implemented by `internal/ui`, wired at `cmd/tellme`:

- **A pure lines port** — the status/tail **byte** formatters: `FormatInputCaptured`, `FormatTurnOpening`, `FormatTurnGap`, `FormatPayloadStatus`, `FormatMetrics`, `FormatReady`, `FormatToolReason`, `FormatToolUsage` (+ the `DefaultToolOutputIdleGap` constant as a domain value). The bytes/caps/sanitize stay **single-owned in `internal/ui`** (ADR 0006/0008/0015) — the domain port names the *shape*, `internal/ui` owns the *bytes*.
- **A `ProgressIndicator` port** — the spinner lifecycle (Start/Stop/Clear/Resume) as a domain interface; the CLI owns the turn-scoped lifecycle epoch, the indicator owns the frames.
- **A `ToolOutput` port** — the `[Tool Output]` block lifecycle (Begin/Writer/End) — the same interface shape as F-8's `domaintools.OutputSink` (D7), so the coordinator satisfies one interface for both uses.
- **Deps factories** — the answer `Renderer` (already behind the CLI's `answerRenderer` interface) and the `ToolLineRenderer` (already an `agentport.ToolLineRenderer` implementation) are obtained via `deps` factories; near-zero churn.

**Why not one broad port (rejected (i))**: it flattens distinct lifecycles and over-bundles. **Why not relocate `call_renderer` wholesale (rejected (iii))**: the frame **schedule** is CLI orchestration (the ADR 0015 precedent keeps the loop's schedule in `internal/agent`); moving it risks byte drift.

## Decision 5: `OutputSink` becomes an interface (F-8)

`domaintools.OutputSink` moves from a struct-of-funcs (`{Begin func(); Writer io.Writer; End func()}`) to an **interface** `{ Begin(); Writer() io.Writer; End(); Enabled() bool }`. `ui.ToolOutputCoordinator` satisfies it **directly**, deleting the struct bridge and the `BindToolOutput` per-turn rebind (the "partial-binding hole" the PR #104 review flagged). **RULE-C-clean**: `internal/domain/tools` already imports `io`.

## Decision 6: A named `Discovery` for MCP (F-7)

`deps.MCPDiscoverer` returns a named **`deps.Discovery{Tools []domaintools.Tool; Warnings []string; Closer io.Closer}`** instead of the bare trailing `func()`. `deps` (tier 2) already imports `domaintools`; `io.Closer` is stdlib. The close becomes a **visible field** (a reader cannot miss it).

## Decision 7: Narrow `Dependencies` seams (F-6)

Leaf functions receive the **narrow seam(s)** they use (a func-typed field or a domain port), not the wide bag; only `run`/`runTurn` (the composition-bearing functions) hold `Dependencies`. Falsifiable acceptance (PR #104 review): *"every function receiving `Dependencies` reads ≤2 of its fields"* — pinned by a unit test that reflects/audits the call sites (or a recorded grep). The **interface-seam caveat** (ADR 0017 §Forward) applies: a new **interface**-typed `Dependencies` field needs its own `Validate()` assertion; a func-typed field does not.

## Decision 8: ADR **0020** + the truth MODIFY

A new **ADR 0020** records: the `→ ui` de-coupling (value-type homes + the four ports/factories), the F-6/F-7/F-8 resolutions, the **two ratchet removals** (baseline **1 → 0** + the RULE-F `→ ui` key), what the change is *not*, and the relation to ADR 0006/0008/0013/0015/0016/0017/0018/0019. Truth: `specs/truth/techstack.md` **MODIFY** — the **Layer-discipline gate** row (baseline **1 → 0**; the RULE-F `→ ui` key removed; the round-051 note) + any row naming the CLI's ui rendering seams; the **Task runner** row is **NOOP**. At baseline **0** the ratchet has **no release valve** (ADR 0011/0016) — recorded.

## Decision 9: The DoD is the gate + the F-6/F-7/F-8 acceptances

Witness: `make verify` (RULE-A/B/C **0**; RULE-E baseline **0**, 0 new / 0 stale; **0** cycles; RULE-F coverage green with **no `→ ui` key**; cross-compile 4/4) + the identifier-count check (`internal/cli` production `→ ui` refs = **0**) + the F-6 (`≤2 fields`), F-7 (named `Discovery`), F-8 (interface `OutputSink`) acceptances + the existing unit/E2E pins. Falsifiability witnesses (reproduced then reverted, ADR 0010): (a) compile-level (delete a port seam ⇒ the CLI fails to compile); (b) count-level (a stray `ui.*` selector ⇒ RULE-E new + RULE-F coverage red); (c) baseline-level (the removal reds **twice** until both halves fold); (d) F-8 (a struct `OutputSink` ⇒ the interface acceptance reds).

## Decision 10: Testing & BDD techstack unchanged; api/data/dsl-refine are NOOP

`/axb-api-plan` — no `contracts/**`. `/axb-data-plan` — no state change. `/axb-dsl-refine` — a `cli`-side rendering de-coupling is not a tellme CLI-contract change. `/axb-spec-by-example` — **NOOP** (no user-facing journey). CLI tests may need to adopt domain-typed fakes for the new ports (the round-050 Q4 → A precedent: RULE-E is merged-graph).

## Decision 11: No new dependency, no new Makefile target, no cycle

`go.mod`/`go.sum` unchanged; the gate rides `verify-architecture`; the ports keep the import direction **downward** (`internal/ui` (tier 5) → `internal/domain/**` (tier 0); the CLI (tier 6) → `internal/domain/**`).

## Decision 12: Behaviour preservation is the hard invariant

Every `stdout`/`stderr` byte, exit code, flag, and the DSL vocabulary is unchanged; the sanitize/cap policy stays single-owned in `internal/ui`; the spinner epochs (rounds 019/025/040) and the coordinator's lock order / mutual-exclusion + join are preserved.

## Residual risks / forward links

- **Blast radius** — the largest single round of R5; the D2 ordering is the mitigation, and a review may require an internal re-cut (recorded).
- **At baseline 0** the ratchet has **no release valve** (ADR 0011/0016).
- **Port-shape leak** — a domain port must carry domain types only (ADR 0015 recorded the risk); the new ports must not name a `ui` type.
- **`Validate()` interface-seam caveat** (ADR 0017 §Forward) — a new interface-typed `Dependencies` field needs its own assertion.
