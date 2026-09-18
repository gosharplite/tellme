# Phase 0 Research: de-couple `internal/cli` from the turn loop (re-cut sub-slice 1) — extracting the loop's domain-facing contracts (Round 049)

Topic: advance [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause — *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities"*) by executing the **first sub-slice** of the re-cut `cli → agent` de-coupling: move the turn loop's **crossing contracts** (the turn-result type, the incomplete-turn error type, and the `ToolDefs` wire-def projection) into **`internal/domain/agent`**, so a later slice can invert the loop **construction/execution** into an injected domain port **edge-sized** ([#101](https://github.com/gosharplite/tellme/issues/101) — R5, the programme).

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` on the built binary; stdlib `testing` for units), provider transports, skills, MCP, the **Bubble Tea** TUI family, and the presentation packages were locked in rounds 001–048. This round adds **no new system end**, **no external service**, and **no new third-party dependency**; the system keeps **one CLI end**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; BDD techstack = `godog` over the built binary; E2E black-box + unit strategy) and are **not re-decided** (D8). **IN**: the extracted `internal/domain/agent` contracts, the `internal/agent` de-duplication (reference + delete), the CLI re-read, ADR 0018, the truth row (re-cut recorded; baseline **unchanged**), and the re-cut's edge-sizing evidence. **OUT**: the `AgentLoop` construction/execution inversion (**sub-slice 2** — the baseline-moving round), the `cli → ui` edge, F-6/F-7/F-8, and any `tellme` binary behaviour.

> **Provenance of the locked scope (#101 + this round's clarify):** round 049 is the **re-cut sub-slice 1** of the `cli → agent` de-coupling. Clarify **Q1 → B** (re-cut into ordered sub-slices), **Q2 → (i)** (extract **all three** crossing contracts), **Q3 → (a)** (reference + delete — **no alias, no forwarder**). The mandatory questions (D8) are unchanged from the standing truth.

> **⚠️ Round shape — the baseline does NOT move this round.** After sub-slice 1, `internal/cli` still imports `internal/agent` (the `AgentLoop` construction), so the RULE-E edge persists and `tools/arch/baseline.txt` is **byte-identical** (still **2**). The round's DoD is **not** "baseline 2 → 1"; it is *"the crossing contracts are domain-owned, behaviour-preservingly, and the surviving `cli → agent` coupling is reduced to a single construction call site — the gate green with 0 new / 0 stale"*. **Sub-slice 2** is the baseline-moving round (**2 → 1**).

---

## Decision 1: The re-cut is the round; the baseline does **not** move (Q1 → B, recorded up front)

- **Decision**: the round executes **sub-slice 1** of the re-cut `cli → agent` de-coupling — it **extracts the crossing contracts to `internal/domain/agent`** and leaves the `AgentLoop` construction in `internal/cli`. The RULE-E baseline stays exactly as it is (**2** — `internal/cli -> internal/agent`, `internal/cli -> internal/ui`), **byte-identical**, with the gate green (0 new / 0 stale).
- **Rationale**: ADR 0017 §Forward classifies the `cli → agent` edge as *"the deepest slice"* — the measured surface is **4 crossing identifiers over 2 files** (`agent.AgentLoop` constructed with 9 fields; `agent.AgentResult`; `agent.ErrIncomplete` type-asserted; `agent.ToolDefs`), which is **not** edge-sized like round-048's single seam. ADR 0017's own recipe for a >edge-sized slice is *"value types → `internal/domain/**` first, then the … factory"*; the re-cut applies that recipe to `→ agent`. Splitting keeps each round behaviour-preserving, gate-proven, and independently reviewable.
- **Alternatives considered**: **the full `cli → agent` inversion in one round** (Q1-A) — a much larger review surface (a 9-field stateful construction + a result type + an error type + a free function) than any R5.x slice so far; **a different/smaller ride-along** (Q1-C) — the operator chose to advance the `→ agent` edge and to keep the re-cut explicit — both rejected.

## Decision 2: Extract **all three** crossing contracts to `internal/domain/agent` (Q2 → (i))

- **Decision**: move into **`internal/domain/agent`** (the existing port family — peer of `LoopObserver`, `CallObserver`, `ToolLineRenderer`):

  ```go
  // The turn result of one prompt run (was agent.AgentResult).
  type Result struct {
      Answer string
      Steps  []history.Step
      Usage  llm.Usage
      Calls  []llm.Usage
  }

  // The incomplete-turn error (was agent.ErrIncomplete): the loop exhausted its
  // bounded cycle with a tool call still outstanding.
  type ErrIncomplete struct { /* … same fields … */ }
  func (e *ErrIncomplete) Error() string
  func (e *ErrIncomplete) Unwrap() error

  // The wire-def projection (was agent.ToolDefs) — a pure function.
  func ToolDefs(reg tools.Registry) []llm.ToolDef
  ```

  After the round, `internal/cli` production references **exactly one** `internal/agent` identifier: `agent.AgentLoop` (the construction).
- **Rationale**: the value types + the pure projection are exactly what a domain port must eventually carry so the port can be **RULE-C-pure** and leak no `agent` type; extracting **all three** (not just the value types) makes sub-slice 2's claim — *"the surviving coupling is the single construction call site"* — **machine-checkable** by an identifier count, and follows the ADR 0017 re-cut recipe. Naming is RD's (`Result` chosen over `TurnResult` to match the existing `internal/domain/**` noun style; either passes FR-002).
- **Alternatives considered**: **value types only** (Q2-ii) — leaves `agent.ToolDefs` on `internal/agent`, so the CLI keeps **2** identifiers and sub-slice 2 is **not** demonstrably edge-sized — rejected (defeats the re-cut's purpose); **also pre-place an unused `LoopRunner` port** (Q2-iii) — ships dead surface now (a governance/dead-code cost) for a de-risk that sub-slice 2 can do in its own round — rejected.

## Decision 3: `internal/agent` **references** the domain contracts directly — no alias (Q3 → (a))

- **Decision**: `internal/agent` **deletes** its own `AgentResult`/`ErrIncomplete`/`ToolDefs` declarations and **references** the domain types directly, e.g.:

  ```go
  func (a *AgentLoop) Run(ctx context.Context, prompt string, prior []history.Entry) (agentport.Result, error) { … }
  req := llm.Request{Tools: agentport.ToolDefs(a.Registry)}
  ```

  `internal/agent`'s three test files (`agentloop_test.go`, `callhooks_test.go`, `usage_calls_test.go`) are **repointed** to the domain types (permitted by FR-003's "moved symbol" carve-out; the asserted behaviour is unchanged). Import direction: `internal/agent` (tier 4) → `internal/domain/agent` (tier 0) is **downward / RULE-A-clean**.
- **Rationale**: alias re-export (`type AgentResult = domainagent.Result`) would keep the three names **exported** from `internal/agent` — two spellings for one identity — and blur the surviving-coupling measurement. **One concept → one name → one home** (the griller canon). Blast radius is bounded: **25 references across 4 files**, **no** consumers outside `internal/agent` + `internal/cli` (measured 2026-09-18 @ `dev` `12964d6`).
- **Alternatives considered**: **alias re-export** (Q3-b) — zero test churn but leaves the names exported and muddies the edge-sizing evidence — rejected.

## Decision 4: The contracts are **RULE-C-pure** — domain types only, no `agent` type crosses

- **Decision**: the extracted contracts reference only **stdlib** + **domain** types. `Result` uses `history.Step` (`internal/domain/history`) + `llm.Usage` (`internal/domain/llm`); `ToolDefs` takes `tools.Registry` (the loop's existing `internal/domain/tools` import) and returns `[]llm.ToolDef`. The incomplete-turn error carries only primitives + an optional wrapped `error`. **No** `agent`-package type appears in any domain declaration.
- **Rationale**: RULE-C (ADR 0011) requires `internal/domain/**` to import only domain + stdlib; the port family the loop already consumes (`LoopObserver`/`CallObserver`/`ToolLineRenderer`) is the precedent — all domain-declared, all structurally satisfied by non-domain implementations.
- **Alternatives considered**: **place the contracts in a new `internal/domain/agent/…` sub-package** — unnecessary; the existing `internal/domain/agent` is the correct home (Q2-i) and avoids a new normative package — rejected.

## Decision 5: The **loop behaviour is unchanged** — the contracts are moved verbatim, not reinterpreted

- **Decision**: the extraction is a **pure relocation**: `Result`'s four fields keep their exact names/types (`Answer string`, `Steps []history.Step`, `Usage llm.Usage`, `Calls []llm.Usage`); the incomplete-turn error keeps its **identical** `errors.As`-based classification and its `emitToolError` `stderr` bytes; `ToolDefs` keeps its exact semantics (`nil` registry → `nil`; otherwise one `llm.ToolDef{Name,Description,Parameters}` per registered tool, preserving registration order) so the **pre-flight estimate is byte-identical** (the round-011 RF-1 "counts exactly what the loop sends" invariant).
- **Rationale**: the round is behaviour-preserving by construction (FR-003/NFR-004); the CLI's turn output (`stdout` answer, `stderr` tail/`Ready`, exit codes) must not change, and the E2E suite runs as **regression**.
- **Alternatives considered**: **fold the extraction into a semantic change** — forbidden: it would break the behaviour-preservation contract and re-open frozen decisions — rejected.

## Decision 6: ADR **0018** + the truth MODIFY (the durable record)

- **Decision**: record the re-cut durably in a **new ADR 0018** (`docs/decisions/0018-cli-agent-contracts-extraction.md` + the `docs/decisions/README.md` index row): the re-cut rationale, the extracted `internal/domain/agent` contracts, the **Q3 (a)** reference-not-alias choice, the **deferred** construction inversion (**sub-slice 2**, baseline **2 → 1**), the **baseline unchanged (2)** statement, **what the change is *not*** (no user-facing change; no new normative source; no baseline movement; the `→ ui` edge untouched), and its relation to ADR **0011** (the tier table + ratchet), **0016** (RULE-E — unchanged this round), and **0017** (the §Forward classification + re-cut recipe). Update `specs/truth/techstack.md`: the **Layer-discipline gate** row records the re-cut and states the baseline figure is **unchanged (2)** pending sub-slice 2 (a real MODIFY, no unevidenced NOOP); any CLI-architecture row naming the loop contracts is restated to the domain-owned reality.
- **Rationale**: per `docs/decisions/README.md`, a structural change future slices must cite gets an ADR; sub-slice 2 must cite the contract home + the extraction decision. The gate row is the sole truth home for the baseline figure.
- **Alternatives considered**: **truth prose only, no ADR** — leaves the re-cut non-citable and would force sub-slice 2 to re-derive the contract home — rejected (mirrors rounds 042/047/048).

## Decision 7: The DoD is **the gate + RULE-F (the coupling-surface allow-list) + unit seams** — not the E2E suite

- **Decision**: prove the round by (i) the gate (RULE-E reports the baseline **2**, 0 new / 0 stale; RULE-A/B/C **0**; cycles **0**) and (ii) the **new RULE-F** — a fail-on-stale **coupling-surface allow-list** in `tools/arch` (`couplingSurface`): the guard parses `internal/cli` production sources, resolves the target package's alias, collects the selected identifiers, and asserts the set equals the allow-list (a new identifier FAILS; an allow-list entry no longer selected FAILS). It tracks **both** baselined application edges (`internal/cli -> internal/agent` ⇒ `{AgentLoop}`; `internal/cli -> internal/ui` ⇒ 19 identifiers) and, via `assertSurfaceCoversBaseline`, requires **every** baselined application edge to be surface-tracked (fold-review **TD-2**) — so a shrunk edge cannot silently re-inflate and no baselined edge sits unprotected. This promotes the round's *"CLI names exactly `AgentLoop`"* claim from a human grep to a **machine check** (round-049 review **TD-1**); and (iii) the existing turn unit pins + the godog E2E run green as **regression**, **plus** three falsifiability witnesses reproduced then reverted (ADR 0010): (a) **compile-level** — delete a domain declaration ⇒ the seam fails to compile; (b) a flipped `Result` field name ⇒ compile break; (c) **count-level (RULE-F)** — add a *surviving* `internal/agent` identifier (`var _ = agent.DefaultToolTimeout`) ⇒ RULE-F FAILS with the RULE-E baseline still green.
- **Rationale**: #92 **AC5** — the witness is the gate + unit seams; a green E2E suite says nothing about which package the CLI imports (the round-009 trap), and RULE-E's **edge** granularity is structurally **blind to the identifier count**, so the round's central claim needed a machine carrier (TD-1). RULE-F generalises: any future shrunk edge can be tracked with one allow-list entry.
- **Alternatives considered**: **rely on the green E2E suite** — rejected; **a stdlib source-scan unit test next to the round's pins** (TD-1 option b) — works but is local to this round, not a general ratchet — rejected in favour of the in-guard RULE-F (option a); **record the count as a manual grep** (option c) — the anti-falsifiable status quo — rejected; **author a new plan-side acceptance Feature** — no user-facing behaviour to express (the round-042 A3/A6 precedent) — rejected.

## Decision 8: Testing & BDD techstack unchanged; `/axb-api-plan`/`/axb-data-plan`/`/axb-dsl-refine` are `NOOP`

- **Decision**: no new system end, no BDD-techstack change, no test-strategy change. `/axb-spec-by-example` = **NOOP** (no user-facing business journey); `/axb-api-plan` = **NOOP** (no OpenAPI surface); `/axb-data-plan` = **NOOP** (a contract relocation, not persisted/runtime state); `/axb-dsl-refine` = **NOOP** (a `cli`-reader refactor is not the `tellme` CLI contract); `/axb-ui-plan` = skipped. The three AIxBDD must-ask questions stay answered by the standing `techstack.md` and are **not** re-decided.
- **Rationale**: the round changes no observable CLI behaviour, so no Gherkin scenario is the carrier; the E2E suite runs green as **regression**. Authoring CLI Gherkin for an internal refactor would risk an `acceptance-coverage` mismatch.
- **Alternatives considered**: **author a plan-side acceptance Feature** — no user-facing behaviour to express — rejected.

## Decision 9: Inherited mechanism, determinism, and the ratchet are unchanged

- **Decision**: RULE-E and its machinery are **not** re-litigated (ADR 0011 D2–D9, ADR 0016): module-root-anchored `go list`, the `CROSS_TARGETS` union + filtered child env, the merged production+test graph, the production-only SCC pass, the sorted baseline, the fail-on-stale ratchet, default-deny. The **sanctioned set** (ADR 0016) and the **baseline lines** are **unchanged** this round (no edge is removed — the edge move is sub-slice 2). Every rule's verdict is unchanged.
- **Rationale**: the round only **relocates contracts**; re-deriving the rule or editing the baseline would re-open settled decisions and risk a spurious ratchet move.
- **Alternatives considered**: **edit the baseline to reflect the extraction** — there is **nothing to remove** (the edge persists via the construction) — rejected; an accidental baseline edit would trip the fail-on-stale ratchet.

## Decision 10: `internal/cli` gains no new coupling; the surviving edge is provably edge-sized for sub-slice 2

- **Decision**: after the round, `internal/cli`'s `internal/agent` usage is **exactly** the 9-field `&agent.AgentLoop{…}` construction + `loop.Run(...)` (whose result/error types are now domain types). The round **MUST NOT** leave a second `agent.*` reference (a partial extraction is a scope failure, FR-005), and MUST NOT add a non-sanctioned import (FR-002/NFR-001). The construction call site + the `Run` call become the **sole** surface sub-slice 2 must invert.
- **Rationale**: the re-cut's whole point is that sub-slice 2 is **edge-sized** (comparable to round-048's single seam) so it can be delivered as one behaviour-preserving round; proving the count now de-risks it.
- **Alternatives considered**: **leave `ToolDefs` for sub-slice 2 to move alongside the inversion** — blurs the sizing split and makes sub-slice 2 >edge-sized again — rejected.

---

## Residual risks / forward links

- **The `AgentLoop` construction/execution inversion (sub-slice 2)** — the baseline-moving round (`internal/cli` → `internal/agent` removed; RULE-E **2 → 1**). The surviving surface after sub-slice 1 is the 9-field construction + the `Run` call, whose result/error types are already domain types — so the port can be declared **RULE-C-pure** in `internal/domain/agent` (peer of the current family) with a tier-≥4 adapter (or the exempt `cmd/tellme`).
- **The `internal/cli → internal/ui` edge** — the last RULE-E residual (also ≥ edge-sized per ADR 0017 §Forward: ~20 call sites · 3 crossing value types · 4 stateful objects · 8 formatters); a later slice.
- **F-6/F-7/F-8** (PR #104 review) — out of scope; recorded on **#101**.
- **The ADR 0017 §Forward `Validate()` note** — `Dependencies.Validate()`'s predicate is `Kind()==reflect.Func`, structurally blind to **interface** seams; if sub-slice 2 adds an interface-typed `Options`/`Dependencies` field it MUST get its own `Validate()` assertion. Recorded for sub-slice 2.
- **Test churn** — `internal/agent`'s 3 test files repoint to the domain types (25 refs / 4 files); no *behavioural* assertion changes. Recorded as a test-only adaptation (Q3-a).
- **No behaviour change** — `stdout`/`stderr`, class-phrase vocabulary, flags, exit codes, the turn chrome, and every existing rule's verdict are unchanged; the round touches `internal/domain/agent/**`, `internal/agent/**`, `internal/cli/**`, an ADR, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's summary. **No** `tools/arch/baseline.txt` edit.
- **Custom build-tag-gated files** stay out of scope (ADR 0011 D6).
