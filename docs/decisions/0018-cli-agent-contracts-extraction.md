# ADR 0018 — De-couple `internal/cli` from the turn loop (re-cut sub-slice 1): extract the loop's crossing contracts to `internal/domain/agent` (R5.3 of [#92](https://github.com/gosharplite/tellme/issues/92))

- **Status**: Accepted
- **Date**: 2026-09-18
- **Round**: 049 `049-cli-agent-decoupling` (the **re-cut sub-slice 1** of the `cli → agent` de-coupling; [#101](https://github.com/gosharplite/tellme/issues/101) — **R5**)
- **Relates to**: **ADR 0011** (the layer-discipline gate + the pinned tier table + the fail-on-stale ratchet — *extended*, not superseded) · **ADR 0016** (RULE-E, the application import ceiling this round **does not** shrink — the edge persists; **unchanged**) · **ADR 0017** (the R5.2 slice, whose §Forward classifies this edge as *"the deepest slice"* and prescribes the **value-types-first re-cut** recipe this round applies) · **ADR 0015** (the port-in-domain precedent — the loop already consumes `agentport.LoopObserver`/`CallObserver`/`ToolLineRenderer`).

## Context

[#92](https://github.com/gosharplite/tellme/issues/92) **AC2** second clause: *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities."* After round 048 (R5.2) the RULE-E baseline (ADR 0016) records **2** residual unsanctioned edges:

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
```

The `cli → agent` edge is **not** edge-sized. Measured 2026-09-18 @ `dev` `12964d6`, `internal/cli` touches `internal/agent` through **4 crossing identifiers over 2 files**:

- `agent.AgentLoop` — the **constructed** loop (9 fields: `Gateway`/`Registry`/`MaxLoops`/`EffectiveBudget`/`Stderr`/`Now`/`ToolUsage`/`Lines`/`Observer`; `cli.go:699`);
- `agent.AgentResult` — the returned turn-result type (`cli.go:~749`, `call_renderer.go:175`);
- `agent.ErrIncomplete` — the incomplete-turn error, type-asserted (`cli.go:754`);
- `agent.ToolDefs` — the wire-def projection free function (`call_renderer.go:77`).

Round 048 (ADR 0017 §Forward) therefore classified this edge as *"the deepest slice"* and prescribed the re-cut recipe it had used informally for the `→ ui` edge: **value types → `internal/domain/**` first, then the … factory**.

## Decision

1. **Re-cut, this round is sub-slice 1.** The `cli → agent` de-coupling is split into ordered sub-slices. **This round extracts the loop's crossing *contracts***; **sub-slice 2** (a later round) inverts the loop **construction/execution** into an injected domain port — and is the **baseline-moving** round (**2 → 1**). (Clarify **Q1 → B**.)

2. **The three crossing contracts move to `internal/domain/agent`.** Sub-slice 1 relocates — into `internal/domain/agent`, the existing port family (peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`) — (i) the **turn-result** type (`Answer string`/`Steps []history.Step`/`Usage llm.Usage`/`Calls []llm.Usage`), (ii) the **incomplete-turn error** type (`Error()`/`Unwrap()`, same `errors.As` classification), and (iii) the **`ToolDefs`** pure projection (`ToolDefs(reg tools.Registry) []llm.ToolDef`). The contracts reference only **stdlib + domain types** — **RULE-C-pure**; no `agent` type crosses in. (Clarify **Q2 → (i)**.)

   **Siting rationale (why `internal/domain/agent`, not `domain/tools` or `domain/llm`).** `ToolDefs(reg tools.Registry) []llm.ToolDef` bridges `domain/tools` (the argument) and `domain/llm` (the result). Siting the projection in `domain/tools` would mint a **new** `tools → llm` domain edge; siting it in `domain/llm` would mint a **new** `llm → tools` edge (neither exists today — verified empty). `internal/domain/agent` is the existing package that already names **both** (it hosts `LoopObserver`/`CallObserver`/`ToolLineRenderer`), so it is the only home that adds **no** new domain cross-edge. The projection consumes only `reg.Tools()` (public `Registry` surface), so it reintroduces **no** coupling to the tool loop.

3. **`internal/agent` references the domain contracts directly — no alias, no forwarder.** `internal/agent` **deletes** its own `AgentResult`/`ErrIncomplete`/`ToolDefs` declarations and uses the domain types verbatim (`Run(...) (agentport.Result, error)`, `agentport.ToolDefs(a.Registry)`); its three test files are repointed to the domain types. One concept → one name → one home. (Clarify **Q3 → (a)**.) Import direction `internal/agent` (tier 4) → `internal/domain/agent` (tier 0) is **downward / RULE-A-clean**.

4. **The RULE-E baseline does NOT move this round.** After sub-slice 1, `internal/cli` still imports `internal/agent` — for the surviving `AgentLoop` **construction** — so `tools/arch/baseline.txt` is **byte-identical** (still **2**; 0 new / 0 stale). The round's DoD is *"the crossing contracts are domain-owned and the surviving `cli → agent` coupling is a single construction call site"*, **not** a ratchet move. Sub-slice 2 ratchets **2 → 1**.

5. **Behaviour-preserving.** The contracts are moved **verbatim, not reinterpreted**: the result field semantics, the `errors.As` incomplete-turn classification (identical `emitToolError` `stderr` bytes), and the `ToolDefs`-based pre-flight estimate (the round-011 RF-1 "counts exactly what the loop sends" invariant) are unchanged. No user-facing behaviour change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary); **no** Gherkin/DSL row; **no** new dependency (`go.mod`/`go.sum` unchanged).

6. **RULE-F — the application coupling surface (machine carrier for the round's DoD).** RULE-E's granularity is the **edge**; it is structurally blind to *how many identifiers* cross a baselined edge, so a shrunk edge could silently re-inflate (a second construction site, a new `agent.*` helper, a package-level `var _ agent.X`) with 0 new / 0 stale. Adding to the guard the **RULE-F** allow-list closes that gap: for a governed application tier, the set of identifiers it may **select** from a listed target package is a normative allow-list (**fail-on-stale**), so this round's headline claim (*"`internal/cli` names exactly `AgentLoop` from `internal/agent`"*) becomes **machine-checked**, not prose. The normative table lives in `tools/arch/arch_test.go` (`couplingSurface`), mirroring the sanctioned-set design (ADR 0016 D1); the gate runs it **before** the `-update-baseline` branch so regenerating the baseline cannot launder a re-inflated surface into a green gate. RULE-F generalises beyond this round (any future shrunk edge can be tracked); its edits land here because this round's DoD required it (review **TD-1**).

## What this change is *not*

- It does **not** invert the `AgentLoop` construction/execution — that is **sub-slice 2** (the baseline-moving round, **2 → 1**).
- It does **not** move the RULE-E baseline, weaken RULE-E, or re-rule its sanctioned set (ADR 0016 is **unchanged**); it does **not** change the tier table, the rule mechanism, or any existing rule's verdict (ADR 0011 stands).
- It does **not** touch the `internal/cli → internal/ui` edge.
- It is **not** a user-facing change and adds **no** new dependency.
- It does **not** re-open any frozen decision (the observer/presentation ports of ADR 0014/0015 are reused, not re-litigated).

## Consequences

- **Positive**: the loop's crossing **contracts** become domain-owned; `internal/cli`'s `internal/agent` references drop **4 → 1** (`agent.AgentLoop` only), so **sub-slice 2 is provably edge-sized** (one construction call site + one `Run` call, whose result/error types are already domain types) — the port pattern (domain interface + tier-≥4 adapter + composition-root injection) is directly reusable from ADR 0017.
- **Cost — test churn**: `internal/agent`'s three test files (`agentloop_test.go`, `callhooks_test.go`, `usage_calls_test.go`) are repointed to the domain types (25 references / 4 files); no *behavioural* assertion changes (a test-only adaptation).
- **Cost — a preparatory round with no visible ratchet move**: this round advances AC2 without shrinking the baseline; its value is the *edge-sizing* of sub-slice 2, which must be read from the identifier count, not the baseline.
- **Faithfulness**: the extraction is a pure relocation; the domain `Result`/`ErrIncomplete`/`ToolDefs` mirror the `internal/agent` originals verbatim.
- **Witness**: the gate (RULE-E reports the baseline **2**, 0 new / 0 stale; RULE-A/B/C **0**; 0 cycles; **RULE-F** reports the `internal/cli -> internal/agent` surface as exactly `{AgentLoop}`, 0 new / 0 stale) + the existing turn unit pins, plus falsifiability witnesses (a) a partial extraction ⇒ the **RULE-F count** reports the extra identifier (reproduced: `agent.DefaultToolTimeout` ⇒ *"NEW coupling-surface identifier(s): [DefaultToolTimeout]"*, gate FAILS); (b) flipping a domain `Result` field ⇒ the CLI seam **fails to compile**; (c) an incomplete extraction leaving a second coupling ⇒ caught by RULE-F — reproduced then reverted (ADR 0010). The E2E suite is regression, not the carrier (#92 AC5).

## Forward

- **Alias note (R-2)**: `internal/domain/agent` is imported as `agentport` at 9 sites, an alias that now under-describes the package (it holds **3 ports** — `CallObserver`/`LoopObserver`/`ToolLineRenderer` — and, after this round, **3 non-port contracts** — `Result`/`ErrIncomplete`/`ToolDefs`). **Decision: keep `agentport`** as the package's conventional short name and read the file doc as *"the loop's domain contract package"* — renaming 9 sites buys no behaviour and would churn every R5.x slice. Recorded so the mismatch is not re-raised each round.
- **Forward — cross-slice coupling at `internal/cli/cli.go:699-745` (R-1)**: the surviving `→ agent` call site (`&agent.AgentLoop{…}`) and the three `→ ui` references (`ui.ToolLineRenderer{}`, `ui.NewToolOutputCoordinator(...)`, `*ui.Spinner`) live in the **same ~47-line wiring block**. So sub-slice 2 **cannot** mechanically move that block to the exempt `cmd/tellme` without deciding what happens to the interleaved `ui.*` references — either the loop port's assembly moves to the root while the `ui` wiring stays in the CLI (the port's assembly is then **split across two tiers** and the `→ ui` edge remains by design), or sub-slice 2 drags part of the `→ ui` slice in and becomes ≥ edge-sized again. Sub-slice 2 must therefore be **scoped from this measured surface** (and must state which of the two remaining edges is sequenced first).
- **Sub-slice 2** (the baseline-moving round): invert the surviving `AgentLoop` construction/execution into an injected domain port. It can now be **RULE-C-pure** in `internal/domain/agent` (the result/error types are domain-owned), with a tier-≥4 adapter (or the exempt `cmd/tellme`); the ratchet moves **2 → 1**. If the assembly moves to `cmd/tellme`, the composition root grows by ~9 field wirings — the *intended* ADR-0013 shape, but the "single assembly site" property is upheld by **review discipline**, not a machine check (the round-044 forward item (d): `cmd/tellme` is tier-table-exempt).
- **`Options`/`Dependencies` interface-seam caveat** (ADR 0017 §Forward): `Dependencies.Validate()`'s predicate is `Kind()==reflect.Func`, structurally **blind to interface seams** — if sub-slice 2 adds an interface-typed field, it MUST get its own `Validate()` assertion.
- The **`internal/cli → internal/ui`** edge remains the last RULE-E residual (also ≥ edge-sized — ADR 0017 §Forward); a later slice.
- **F-6/F-7/F-8** remain recorded on [#101](https://github.com/gosharplite/tellme/issues/101).
- At baseline **0** (after sub-slice 2 + the `→ ui` slice) the ratchet has **no release valve** (ADR 0011/0016): a future legitimate ceiling violation must be refactored (or the rule amended), never baselined.
