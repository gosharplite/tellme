# ADR 0019 — De-couple `internal/cli` from the turn loop (sub-slice 2): invert the `AgentLoop` construction/execution into an injected domain port (R5.4 of [#92](https://github.com/gosharplite/tellme/issues/92))

- **Status**: Accepted
- **Date**: 2026-09-18
- **Round**: 050 `050-agentloop-construction-inversion` (the **sub-slice 2** of the `cli → agent` de-coupling; [#101](https://github.com/gosharplite/tellme/issues/101) — **R5**); the **baseline-moving** round (**2 → 1**)
- **Relates to**: **ADR 0011** (the layer-discipline gate + the pinned tier table + the fail-on-stale ratchet — *extended*, not superseded) · **ADR 0016** (RULE-E, the application import ceiling — this round **shrinks** it by one line) · **ADR 0017** (the R5.2 slice — the port-in-domain recipe this round reuses) · **ADR 0018** (the R5.3 re-cut sub-slice 1 — whose §Forward prescribes *this* round as the baseline mover and names its **two** ratchet removals; fold F-4) · **ADR 0013** (the injected `Dependencies` seam this round gains a field on) · **ADR 0014/0015** (the observer/presentation ports reused, not re-litigated)

## Context

[#92](https://github.com/gosharplite/tellme/issues/92) **AC2** second clause: *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities."* After round 049 the RULE-E baseline (ADR 0016 + 0018) records **2** residual unsanctioned edges:

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
```

Round 049 (ADR 0018) extracted the loop's crossing **contracts** (`Result`/`ErrIncomplete`/`ToolDefs`) into `internal/domain/agent`, dropping the CLI's `→ agent` references **4 → 1** and making this edge **provably edge-sized**. Measured 2026-09-18 @ `dev` `684e41e`:

- `internal/cli`'s **only** `internal/agent` reference is the **construction** `&agent.AgentLoop{…}` (`cli.go:699`, 9 fields); the `Run` result/error types are already domain-owned.
- The construction site shares the ~47-line wiring block `cli.go:699-745` with **three** `→ ui` references (`ui.ToolLineRenderer{}` L716 · `*ui.Spinner` L721 · `ui.NewToolOutputCoordinator(...)` L739) — the **R-1 cross-slice coupling** ADR 0018 §Forward flagged.

## Decision

1. **The `→ agent` construction is inverted into an injected domain port, this round.** It is the **baseline-moving** round of R5 (**2 → 1**). (Clarify **Q1 → A**.)

2. **The R-1 split is port-only (Q1 → A).** Only the **`→ agent`** construction inverts; the **`ui`** wiring (the `Lines` renderer, the spinner, the tool-output coordinator, and the composite-observer assembly) **stays in `internal/cli`** by design. The `internal/cli → internal/ui` line **and** its RULE-F surface (19 identifiers) stay **byte-identical**; that edge is the later slice (ADR 0017 §Forward).

3. **The port is a domain interface fed by a func-typed factory (Q2 → (i)).** Declared in `internal/domain/agent` (peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`): **`Loop`** (`Run(ctx, prompt, prior) (Result, error)` — `Result`/`ErrIncomplete` already domain-owned by ADR 0018), **`LoopSpec`** (the loop's 9 construction inputs, all already domain/stdlib-typed), and **`LoopFactory func(LoopSpec) Loop`**. The adapter is **`agent.NewLoop(spec agentport.LoopSpec) agentport.Loop`** in `internal/agent` (tier 4; imports `internal/domain/**` **downward**), bound at the tier-exempt composition root (`cmd/tellme`) via a new **func-typed** `deps.Dependencies` field. Because the field is func-typed, `Dependencies.Validate()`'s existing `Kind()==reflect.Func` predicate covers a nil seam — **no** interface-seam assertion is required (ADR 0017 §Forward stays latent). The port is **RULE-C-pure**: stdlib + domain types only; no `agent`/`ui` type crosses.

4. **The CLI supplies `Lines` and the observer, domain-typed (Q3 → (a)).** `LoopSpec.Lines` (`agentport.ToolLineRenderer`) and `LoopSpec.Observer` (`agentport.LoopObserver`) are passed by the CLI exactly as today (`ui.ToolLineRenderer{}` + its `compositeObserver{…}`). Behaviour is preserved: identical tool-line bytes, identical waiting-phase/observer ordering, unchanged nil-safe defaults. This is **why** the `→ ui` edge survives — the CLI still legitimately holds `ui.ToolLineRenderer{}` and the `*ui.Spinner`-holding composite, crossing the boundary only through **domain-typed** ports.

5. **Two ratchet removals, not one (fold F-4).** When the `→ agent` edge disappears the gate reds **twice** — the **RULE-E** baseline line (`internal/cli -> internal/agent`) **and** the **RULE-F** `couplingSurface["internal/cli -> internal/agent"]` key. Both are removed in the **same PR**, deliberately. The `RULE-E` line is regenerated (`make verify-architecture-update`); the **RULE-F** key is a **guard-table edit** — the table has no regeneration affordance (ADR 0018 D6).

6. **Behaviour-preserving.** The adapter assigns the identical 9 fields to the same `AgentLoop` struct; the loop body, the observer injection point (`LoopSpec.Observer` = the same `compositeObserver`), the `errors.As(*ErrIncomplete)` classification, the `ToolDefs`-based pre-flight estimate, and the deferred final tail are **unchanged**. `stdout`/`stderr` byte-contracts, flags, and exit codes are unchanged; **no** Gherkin/DSL row; **no** new dependency (`go.mod`/`go.sum` unchanged); **no** new Makefile target.

## What this change is *not*

- It does **not** touch the `internal/cli → internal/ui` edge (Q1 → A — the CLI keeps its `ui` wiring).
- It does **not** move/weaken RULE-E's sanctioned set, the tier table, the rule mechanism, or any other rule's verdict (ADR 0011/0016 stand, except the baseline **2 → 1**).
- It does **not** re-open R2's frozen decisions (composition-root home `cmd/tellme`, the `Dependencies` shape, `agentTools()` relocation, MCP orchestration) nor the ADR 0014/0015 observer/presentation ports.
- It is **not** a user-facing change and adds **no** new dependency.

## Consequences

- **Positive**: `internal/cli` production references **zero** `internal/agent` identifiers → the RULE-E ratchet moves **2 → 1**; the `cli → agent` de-coupling is **complete**; the port reuses the domain `Result`/`ErrIncomplete`/`ToolLineRenderer` verbatim, so it is RULE-C-pure with no new domain cross-edge.
- **The `→ ui` edge remains** (by design) — the last RULE-E residual, ≥ edge-sized, a later slice.
- **Cost — root knowledge**: `deps.Dependencies` gains one func-typed field; the `AgentLoop` field shape now flows through `agentport.LoopSpec` (a new domain value type). Small, review-visible.
- **Witness**: the gate + the identifier-count check + the existing turn unit pins. RULE-E baseline **1**, 0 new / 0 stale; RULE-A/B/C **0**; 0 cycles; RULE-F `→ ui` **19** identifiers (0 new / 0 stale), no `→ agent` key; cross-compile 4/4.

## Forward

- **The `internal/cli → internal/ui` edge** — the last RULE-E residual; ≥ edge-sized (≈20 call sites, 3 crossing value types, 4 stateful objects, 8 pure formatters; ADR 0017 §Forward). Likely a **re-cut** (value types → `internal/domain/**` first, then the renderer factory). At baseline **0** after it, the ratchet has **no release valve** (ADR 0011/0016).
- **`cmd/tellme` single-assembly-site** is upheld by **review discipline**, not a machine check (round-044 forward item (d) — the root is tier-table-exempt).
- **F-6/F-7/F-8** remain recorded on [#101](https://github.com/gosharplite/tellme/issues/101).
- **Build-tag scope**: the RULE-E/RULE-F scans cover **production** sources only (ADR 0018 F-3).
