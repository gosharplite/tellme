# Phase 0 Research: de-couple `internal/cli` from the turn loop — sub-slice 2: invert the `AgentLoop` construction/execution into an injected domain port (Round 050)

**Plan Package**: `specs/plans/050-agentloop-construction-inversion`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)) — the **sub-slice 2** of the re-cut `cli → agent` de-coupling; the **baseline-moving** round (**2 → 1**).
**Predecessors**: round 047 (R5.1, RULE-E + baseline) · round 048 (R5.2, the `→ ui/tui/prompt` port) · round 049 (R5.3 / sub-slice 1, the crossing contracts → `internal/domain/agent`; **ADR 0018**).

## Decision 1: The inversion is the round; the baseline **does** move (contrast round 049)

Round 049's DoD was *"contracts domain-owned, baseline unchanged"* (a preparatory slice). **This round's DoD is the ratchet move:** `internal/cli` production references **zero** `internal/agent` identifiers, so the RULE-E baseline drops **2 → 1** (`internal/cli -> internal/agent` gone; `internal/cli -> internal/ui` retained). ADR 0018 §Forward (fold **F-4**) records that the removal reds **twice** — the RULE-E baseline line **and** the RULE-F `couplingSurface["internal/cli -> internal/agent"]` key — so **both** are removed deliberately in the same PR.

## Decision 2: The R-1 cross-slice coupling splits **port-only** (Q1 → A)

Measured @ `dev` `684e41e`: the surviving `→ agent` site (`&agent.AgentLoop{…}`, `cli.go:699`) shares the ~47-line block `cli.go:699-745` with **three** `→ ui` references (`ui.ToolLineRenderer{}` L716, `*ui.Spinner` L721, `ui.NewToolOutputCoordinator(...)` L739). **Locked: invert only the `→ agent` construction into a domain port; the `ui` wiring (the `Lines` renderer, the spinner, the tool-output coordinator, and the composite-observer assembly) stays in `internal/cli` by design.** Consequence: the `→ ui` line **and** its **19**-identifier RULE-F surface stay **byte-identical**; the `→ ui` edge is the later slice (ADR 0017 §Forward — not edge-sized). Rejected: **B** (relocate the whole block to the tier-exempt `cmd/tellme` — grows the root and the `runTurn` signature while the CLI still needs the `*ui.Spinner`-holding composite, so no `→ ui` removal for the size); **C** (re-sequence the `→ ui` slice first — defers the baseline move ADR 0018 assigns to this round).

## Decision 3: The port is a domain **interface** fed by a **func-typed factory** (Q2 → (i))

Declare in `internal/domain/agent` (peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`):

- **`Loop`** — an interface: `Run(ctx context.Context, prompt string, prior []history.Entry) (Result, error)`. `Result`/`ErrIncomplete` are already domain-owned (round 049), so the interface is **RULE-C-pure**.
- **`LoopSpec`** — a value struct carrying the loop **struct's 9 fields** (the struct's fields, not the loop's whole config surface — package-level defaults like `agent.DefaultToolTimeout` are not expressible here), every one already domain/stdlib-typed: `Gateway llm.Gateway`, `Registry tools.Registry`, `MaxLoops int`, `Stderr io.Writer`, `Now func() time.Time`, `EffectiveBudget int`, `ToolUsage history.ToolUsageSink`, `Lines ToolLineRenderer`, `Observer LoopObserver`.
- **`LoopFactory func(LoopSpec) Loop`** — the injected factory type.

The adapter is a new exported constructor **`agent.NewLoop(spec agentport.LoopSpec) agentport.Loop`** in `internal/agent` (tier 4; imports `internal/domain/**` **downward** — RULE-A-clean); it assigns the spec's fields to the existing `AgentLoop` literal. `deps.Dependencies` gains a **func-typed** `LoopFactory` field bound at `cmd/tellme` (tier-table-exempt). **`Validate()` impact (FR-009):** the added field is func-typed, so the existing `Kind()==reflect.Func` predicate already covers it — **no** explicit interface-seam assertion is needed; ADR 0017 §Forward's interface-seam caveat stays latent. Rejected: **(ii)** a domain struct-of-funcs (the F-8 smell minted fresh); **(iii)** a `cmd/tellme` closure (hides the loop's field contract in `main`).

**Siting**: `internal/domain/agent` is the only home that adds **no** new domain cross-edge — `LoopSpec` names `llm.Gateway`, `tools.Registry`, `history.ToolUsageSink`, `agentport.ToolLineRenderer`, `agentport.LoopObserver`, all already domain-owned and already named by that package (mirrors ADR 0018's `ToolDefs` siting rationale).

## Decision 4b: The CLI **test** surface also leaves `internal/agent` (Q4 → A)

RULE-E evaluates the **merged** (production + in-process test) import graph (ADR 0016 D4; `tools/arch/arch_test.go` — *"test imports ARE governed"*, review N-3), and RULE-D's "no unranked package" check also reads the merged graph. So a `_test.go` import of `internal/agent` in `internal/cli` would keep the `internal/cli -> internal/agent` edge **alive**, the baseline line would **not** go stale, and the DoD (**2 → 1**) would be unreachable. **Locked (Q4 → A):** the CLI test fixture's `LoopFactory` returns an **in-package fake `agentport.Loop`**; the real-loop coverage stays in **`internal/agent` unit tests + the godog E2E**; the three loop-dependent `runTurn` test files are reworked to assert **CLI orchestration** (persistence, exit codes, stream ordering) rather than loop internals. Recorded cost: honest test-only churn (`turn_test.go`, `persistence_invariant_test.go`, `tui_submit_chrome_test.go`). Rejected: **B** (keep a `_test.go` `internal/agent` import → DoD fails); **C** (amend RULE-E to production-only → a rule change contradicting ADR 0016/review N-3). Note: RULE-F itself is **production-only** (ADR 0018 F-3), so the RULE-F `→ agent` key removal is unaffected by test files — it is RULE-E that forces the test-surface move.

## Decision 4: The CLI keeps `Lines` + the observer, domain-typed (Q3 → (a))

`LoopSpec.Lines`/`LoopSpec.Observer` are **domain-typed**; the CLI passes `ui.ToolLineRenderer{}` and its `compositeObserver{…}` **exactly as today**. This is **behaviour-preserving**: identical `[Tool Reason]`/yield tool-line bytes, identical waiting-phase/observer ordering, and the loop's documented **nil-safe** defaults are unchanged. It is **why** the `→ ui` edge survives (Q1 → A): `ui.ToolLineRenderer{}` and the `*ui.Spinner`-holding composite still cross, but only through **domain-typed** ports — no `agent` type leaks into the CLI, RULE-C intact. Rejected: **(b)** factory-supplied defaults (a silent behaviour change; and the domain port cannot build them without importing `ui` — a RULE-C breach); **(c)** moving the observer assembly into the port/adapter (drags `ui`/CLI types into the domain).

## Decision 5: The port is **RULE-C-pure**

`Loop`/`LoopSpec`/`LoopFactory` reference only **stdlib + domain** types — `context`, `io`, `time`, and the domain `llm`/`tools`/`history`/`agent` packages. **No** `internal/agent` type crosses (the loop struct stays private to `internal/agent`); **no** `internal/ui` type crosses (the `Lines`/`Observer` fields are the domain ports). Verified: the loop's 9 constructed fields are **already** all domain/stdlib-typed (there is nothing to re-type — the extraction is a pure seam).

## Decision 6: The inversion is **behaviour-preserving** (the struct's 9 fields, the hooks, the error, the tail)

The adapter assigns the same 9 fields to the same struct; the loop body is **unchanged**. The observer injection point (`loop.Observer = compositeObserver{…}`) is replaced by `LoopSpec.Observer` — the same value, so the call-hook ordering (per-call begin/end, final-tail deferral) is identical. The `errors.As(*agentport.ErrIncomplete)` classification (`internal/cli/cli.go:754`) and the `ToolDefs`-based pre-flight estimate are **already** domain-typed (round 049) and untouched. `stdout`/`stderr` stay **byte-identical**; exit codes and flags are unchanged. No behavioural assertion changes except where a test names a moved/renamed symbol.

## Decision 7: ADR **0019** + the truth MODIFY (the durable record)

A new **ADR 0019** (`docs/decisions/0019-agentloop-construction-inversion.md` + the `README.md` index row) records: the port-only inversion (Q1-A), the domain `Loop`/`LoopSpec`/`LoopFactory` shape + the `agent.NewLoop` adapter (Q2-i), the CLI-supplied domain-typed `Lines`/`Observer` (Q3-a), the **two** ratchet removals (RULE-E line + RULE-F key; F-4), the retained `→ ui` edge, **what the change is *not***, and the relation to ADR 0011/0016/0017/0018. Truth: `specs/truth/techstack.md` **MODIFY ×2** — the **Layer-discipline gate** row (baseline **2 → 1**; the removed RULE-F key; the round-050 note) and the **Agent tool loop** row (the construction home = the injected domain port + `agent.NewLoop`). The **Task runner** row is **NOOP** (no new Makefile target — the gate rides `verify-architecture`).

## Decision 8: The DoD is **the gate** — the two removals + the byte-stable `→ ui` surface

The witness is `make verify` (RULE-E **0 new / 0 stale** at baseline **1**; RULE-A/B/C **0**; **0** cycles; **RULE-F** reports `→ ui` at **19** identifiers, 0 new / 0 stale, and no `→ agent` key) + the identifier-count check (an `internal/cli` production scan finds **0** `internal/agent` references) + the existing turn unit pins — **not** the E2E suite (#92 **AC5**). Falsifiability witnesses (reproduced then reverted, ADR 0010): (a) **compile-level** — delete/rename the port seam ⇒ the CLI turn path fails to compile; (b) **count-level** — reintroduce a direct `agent.*` selector in the CLI (or a second construction) ⇒ RULE-F **/ **RULE-E **FAILS** while the `→ ui` surface stays green; (c) **baseline-level** — before both removals are folded the gate reds **twice** (RULE-E line **and** RULE-F key).

## Decision 9: Testing & BDD techstack unchanged; `/axb-api-plan` / `/axb-data-plan` / `/axb-dsl-refine` are **NOOP**

`/axb-api-plan` — no `specs/truth/contracts/**` (a single CLI end). `/axb-data-plan` — no persisted/runtime-state change (the round moves a construction, not data). `/axb-dsl-refine` — an internal refactor is not a `tellme` CLI-contract change; **no** Rule/Example/step/`DSLRow` changes (the rounds 020/031/036/041–049 non-BDD-tooling precedent). `/axb-spec-by-example` — **NOOP** (no user-facing journey).

## Decision 10: Inherited mechanism, determinism, and the ratchet are unchanged

No new dependency (`go.mod`/`go.sum` unchanged); **no** new Makefile target; the `-count=1`/`-tags=arch` invocation and the `CROSS_TARGETS`/child-env hermeticity stay as shipped (ADR 0011/0012); the RULE-F table has **no** regeneration affordance (ADR 0018 D6) — the `→ agent` key is a deliberate guard-table edit; `internal/domain/**` gains only domain-typed surface.

## Decision 11: The `Validate()` seam is a **func-typed** field (no interface assertion)

The only new `deps.Dependencies` field is `LoopFactory` (func-typed), so `Dependencies.Validate()`'s `Kind()==reflect.Func` predicate covers a nil seam; **no** new `Validate()` assertion is required. Recorded so a future interface-typed seam re-adjudicates (ADR 0017 §Forward).

## Decision 12: `internal/cli` gains no new coupling; the surviving edge is removed in one PR

The round adds **no** new `internal/cli → {agent, ui}` reference; the single `→ agent` construction is inverted, and the rotation lands **atomically** with the truth/ADR update and the two ratchet removals (NFR-002).

## Residual risks / forward links

- **The `internal/cli → internal/ui` edge** (the last RULE-E residual; ≥ edge-sized — ADR 0017 §Forward; likely a re-cut) — the next slice; **not** this round (Q1 → A).
- **`cmd/tellme` single-assembly-site** upheld by **review discipline**, not a machine check (round-044 forward item (d): the root is tier-table-exempt).
- **At baseline 0** (after this round + the `→ ui` slice) the ratchet has **no release valve** (ADR 0011/0016).
- **F-6** (narrow `Dependencies` seams) · **F-7** (a named `Discovery{Tools, Warnings, Closer io.Closer}`) · **F-8** (`domaintools.OutputSink` → interface) remain on [#101](https://github.com/gosharplite/tellme/issues/101).
- Build-tag scope: the RULE-E/RULE-F scans cover **production** sources only (ADR 0018 F-3).
