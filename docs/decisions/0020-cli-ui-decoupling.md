# ADR 0020 — Close [#101](https://github.com/gosharplite/tellme/issues/101): de-couple `internal/cli` from `internal/ui` (baseline 1 → 0) + the F-6/F-7/F-8 seam resolutions (R5.5 of [#92](https://github.com/gosharplite/tellme/issues/92))

- **Status**: Accepted
- **Date**: 2026-09-18
- **Round**: 051 `051-cli-ui-decoupling` — the **terminal R5 slice** ([#101](https://github.com/gosharplite/tellme/issues/101) — **R5**); the **last** RULE-E residual + the PR [#104](https://github.com/gosharplite/tellme/pull/104) deferrals
- **Relates to**: **ADR 0011** (the layer-discipline gate + the pinned tier table + the fail-on-stale ratchet — *completed*: the baseline reaches **0**) · **ADR 0016** (RULE-E — this round **empties** its baseline) · **ADR 0017** (the port-in-domain recipe) · **ADR 0018/0019** (the R5.3/R5.4 `→ agent` re-cut + the two-removal discipline) · **ADR 0013** (the injected `Dependencies` seam) · **ADR 0006/0008/0015** (the ui-owned tool-line bytes/caps/sanitize policy — *untouched*) · **ADR 0014** (the yield policy — *untouched*)

## Context

[#92](https://github.com/gosharplite/tellme/issues/92) **AC2** second clause: *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities."* After round 050 the RULE-E baseline records **1** residual edge:

```text
internal/cli -> internal/ui
```

Measured 2026-09-18 @ `dev` `310def4`: the CLI names **19** `internal/ui` identifiers (RULE-F `couplingSurface`) over ~33 call sites — 3 crossing **value types** (`Pricing`/`UsageCounts`/`ToolUsageRow`), 8 pure **formatters**, 4 stateful **objects/factories** (`Renderer`/`Spinner`/`ToolOutputCoordinator`/`ToolLineRenderer`). The edge is **not edge-sized** (ADR 0017 §Forward). Separately, the PR [#104](https://github.com/gosharplite/tellme/pull/104) review deferrals **F-6/F-7/F-8** remain open.

## Decision

1. **One round closes #101 (Q1 → D).** This round removes the `→ ui` edge (**baseline 1 → 0**) **and** resolves F-6/F-7/F-8. **Blast-radius caveat**: if a review judges the change exceeds a single PR, the parts are **ordered** (D2) so an internal re-cut still closes #101 across the ordered slices.

2. **The value types re-home into existing domain peers (Q2 → (i)).** `Pricing` + the pure `ComputeCost`/`HitRate` → **`internal/domain/llm`** (cost is a provider concern; persistence consumes `ComputeCost` too); `UsageCounts` → **`internal/domain/metrics`**; `ToolUsageRow` → **`internal/domain/history`**, **folded onto** `history.ToolUsageCounts`. No new package; no new domain cross-edge.

3. **Narrow lifecycle-grouped ports (Q3 → (ii)).** Declared in `internal/domain/**`, implemented by `internal/ui`, wired at `cmd/tellme`: a **pure lines port** (the 8 status/tail formatters `Format*` + the `DefaultToolOutputIdleGap` value — the domain names the *shape*, `internal/ui` keeps the *bytes*); a **`ProgressIndicator`** port (the spinner lifecycle); a **`ToolOutput`** port (the `[Tool Output]` block); and the answer `Renderer` + `ToolLineRenderer` obtained via **deps factories** (both already interface-seamed). The CLI keeps its frame **schedule** (orchestration) — the ADR 0015 precedent (the loop owns the schedule; `internal/ui` owns the bytes).

4. **F-8 — `domaintools.OutputSink` becomes an interface** `{ Begin(); Writer() io.Writer; End(); Enabled() bool }`; `ui.ToolOutputCoordinator` satisfies it **directly** (the struct-of-funcs bridge + the `BindToolOutput` partial-binding hole are deleted). RULE-C-clean (`internal/domain/tools` already imports `io`).

5. **F-7 — a named `deps.Discovery{Tools, Warnings, Closer io.Closer}`** replaces `MCPDiscoverer`'s bare trailing `func()` (the close is a visible field).

6. **F-6 — the `Dependencies` seams are narrowed**: a leaf receives the narrow seam(s) it uses; only `run`/`runTurn` hold the bag; acceptance *"every function receiving `Dependencies` reads ≤2 of its fields."* A new **interface**-typed `Dependencies` field gets its own `Validate()` assertion (ADR 0017 §Forward); a func-typed field does not.

7. **Behaviour-preserving.** Every `stdout`/`stderr` byte, exit code, flag, and the DSL vocabulary is unchanged. The sanitize/cap policy stays single-owned in `internal/ui` (ADR 0006/0008/0015 — **untouched**); the spinner epochs (rounds 019/025/040) and the coordinator's lock order / mutual-exclusion + join are preserved. No new dependency; no new Makefile target; no Gherkin/DSL change.

## What this change is *not*

- It does **not** re-open R2/`#101` frozen decisions (composition-root home `cmd/tellme`, the `Dependencies` shape, `agentTools()` relocation, MCP orchestration) nor the ADR 0014/0015/0017/0018/0019 ports (reused, not re-litigated).
- It does **not** change the RULE-E **sanctioned set**, the tier table, or the rule mechanism (ADR 0016 stands, except the baseline reaches **0**).
- It does **not** move the ui-owned bytes/caps/sanitize policy into the domain (the ports carry shapes; `internal/ui` owns the bytes).
- It is **not** a user-facing change and adds **no** new dependency.

## Consequences

- **Positive**: the RULE-E ratchet reaches its **terminal state (0)** → [#92](https://github.com/gosharplite/tellme/issues/92) **AC2 clause 2 holds**; [#101](https://github.com/gosharplite/tellme/issues/101) closes; the CLI names only `internal/domain/**` + stdlib + the sanctioned set; F-6/F-7/F-8 are code-resolved.
- **⚠️ No release valve** *(in part superseded — see §Fold review folds, **R-51-1**: the valve is a guarded, review-visible two-edit path, now enforced)*: at baseline **0** a future legitimate ceiling violation must be **refactored** (or the rule amended), not baselined (ADR 0011/0016).
- **Cost — a large single round**: the ui de-coupling + F-6/F-7/F-8 in one PR; the D2 ordering is the mitigation and a review may require an internal re-cut.
- **Witness**: the gate (RULE-E baseline **0**; RULE-A/B/C **0**; 0 cycles; RULE-F no `→ ui` key) + the F-6/F-7/F-8 acceptances + unit/E2E pins.

## Forward

- **No further RULE-E slices remain** — the residual set is empty. Future layer-discipline growth must be prevented by the ratchet (fail-on-new) with **no** baselining.
- The **#92 records** (permanent E2E narrowing; one-concurrent-block coordinator; `End`-while-write-stalled) stay records.
- [#101](https://github.com/gosharplite/tellme/issues/101) **closes** on delivery; any residual nits (e.g. a recorded `≤2 fields` audit boundary) are noted in the round's `truth-delta.md`.

## Fold review folds (PR [#114](https://github.com/gosharplite/tellme/pull/114), review `5729783315`)

- **R-51-1 (the terminal-state valve).** The claim *"at 0 the ratchet has no release valve"* was **overstated**: `-update-baseline` (`writeBaseline`) had **no growth guard**, so the reachable path was a **two-review-visible-edit valve** (regenerate + add a `couplingSurface` key). **Fixed by enforcement**: the `-update-baseline` branch now **refuses to GROW** the baseline (`len(violations) > len(prev)` ⇒ `t.Fatalf`), so a new ceiling violation must be **refactored**, never baselined. The three surfaces (ADR · `techstack.md` · PR) are restated to *"the valve is a loud two-edit path, now guarded"*.
- **R-51-2 (the F-6 exemption record).** The `bagHolders` exemption set was inaccurate (`run` takes `Options`, not `deps.Dependencies`; `renderTurn` is the live forwarder) and had **no liveness assertion**. Folded: the exemption set is **`{renderTurn, runTurn}`** and the audit test **fails on a stale exemption** (the `assertSanctionedInUse` analogue). **Metric boundary (recorded)**: F-6 counts **direct distinct selector reads** on a `deps.Dependencies` parameter — a **pass-through** (forwarding the bag) and **closure** frames are exempt by construction; so the claim is a *read* bound, not an ownership bound.
- **R-51-3 (F-8 typed-nil).** The interface form moved the struct-of-funcs' inherent zero-value safety onto callers; a **typed-nil** `*ToolOutputCoordinator` in a non-nil interface defeated `sink != nil`. Folded: `Enabled`/`Begin`/`End`/`Writer` carry **nil-receiver guards** (`c != nil && …`), pinned by `internal/ui/TestOutputSinkTypedNilIsDisabled`.
- **R-51-4 (test-double spelling).** The CLI's `fakeLines` **re-authored the production byte formats**, so the CLI chrome assertions pinned a *copy*. Folded: `fakeLines` returns **distinguishable sentinels** (no production spelling); the CLI tests assert **orchestration** (presence/order/count/blank rules) and the **real bytes** are pinned in `internal/ui` + the godog E2E (the round-046 fold F-1 pattern applied to the port).
- **R-51-5 (STATUS self-contradiction).** `STATUS.md` asserted both *"no round is in flight"* and *"051 in flight, PR #114 open"*. Folded.

## Decision (RF-51-2) — the render port family: adjudicated, not by precedent

`render.Lines` returns **finished presentation bytes** (glyphs, `$%.4f`, caps), so the **domain declares the spelling** and no domain-side pin can hold it — inverting the round-046 fold N-3 principle (*"port = invariant · adapter = the conforming impl · ui pin = the adapter's spelling"*). **This round adjudicates it explicitly: the ports are kept in `internal/domain/render`, and the bytes stay owned + pinned by `internal/ui`** (its `Format*` unit pins) + the godog E2E. The trade-off accepted: the domain now names the *shape* of a presentation surface (a **presentation-shape leak**, the same class ADR 0015 recorded for `ToolLineRenderer`), bought for the terminal de-coupling. **Forward options (for a later round, not this one):** (i) move the pure formatters **and** `sanitizeControl` into `internal/domain/render` (pure stdlib builders, already policy-pinned by the ADR 0006/0008/0015 tests) and delete the `Lines` interface; or (ii) keep them in `internal/ui` and expose **per-formatter func-typed** deps seams. Also recorded: `DefaultToolOutputIdleGap` is a **value smuggled through the interface** — a future round may expose it as a plain deps value.

## Forward (round-051 review non-blocking items)

- **RF-51-1** — the round-044 fix-5 *gate-before-provider-read* discipline is dropped: `NewProgress` builds the metrics provider unconditionally. Harmless today (`NewSystemMetricsProvider()` is pure), **recorded** here so a future round either passes a `func() metrics.SystemMetricsProvider` (built only when the indicator is enabled) or re-records the constructor as pure.
- **RF-51-3** — `dp.NewLines()` is called ~4× per run (a stateless zero-sized adapter); resolve once and thread if it ever stops being cheap.
- **RF-51-4** — `internal/domain/render` ships no package-side pin and `internal/domain/metrics/usage_counts.go` none; the moved `Pricing` test travelled ✅. Add small pins or record the disposition (the round-021 TD2' *witness the contract* precedent).
- **RF-51-5** — at baseline 0 the RULE-F coverage clause is only load-bearing on the `-update-baseline` path (now guarded by R-51-1); the truth row carries a clause.
- **RF-51-6** — F-6's letter is met but trades a *named* bag for **positional parameter lists** (`runInteractiveTUI` 7 args, `renderToolUsage` 5, `dispatchReporting` 5); if it grows again, a small named per-path seam struct is cheaper.
