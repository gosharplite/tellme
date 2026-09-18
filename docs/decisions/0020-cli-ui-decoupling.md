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
- **⚠️ No release valve**: at baseline **0** a future legitimate ceiling violation must be **refactored** (or the rule amended), **never** baselined (ADR 0011/0016).
- **Cost — a large single round**: the ui de-coupling + F-6/F-7/F-8 in one PR; the D2 ordering is the mitigation and a review may require an internal re-cut.
- **Witness**: the gate (RULE-E baseline **0**; RULE-A/B/C **0**; 0 cycles; RULE-F no `→ ui` key) + the F-6/F-7/F-8 acceptances + unit/E2E pins.

## Forward

- **No further RULE-E slices remain** — the residual set is empty. Future layer-discipline growth must be prevented by the ratchet (fail-on-new) with **no** baselining.
- The **#92 records** (permanent E2E narrowing; one-concurrent-block coordinator; `End`-while-write-stalled) stay records.
- [#101](https://github.com/gosharplite/tellme/issues/101) **closes** on delivery; any residual nits (e.g. a recorded `≤2 fields` audit boundary) are noted in the round's `truth-delta.md`.
