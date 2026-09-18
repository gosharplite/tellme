# Feature Specification: de-couple `internal/cli` from the turn loop — sub-slice 2: invert the `AgentLoop` construction/execution into an injected domain port (round 050)

**Feature Branch**: `050-agentloop-construction-inversion`

**Created**: 2026-09-18

**Status**: Draft (clarify round 1 in progress — **Q1 → A CLOSED**; Q2/Q3 pending) — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)), the **sub-slice 2** of the re-cut `cli → agent` de-coupling. Round 047 = **R5.1** (the RULE-E gate + baseline) · round 048 = **R5.2** (`cli → ui/tui/prompt`) · round 049 = **R5.3 / sub-slice 1** (the loop's crossing contracts → `internal/domain/agent`).

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5**, the **baseline-moving** round (**2 → 1**). The committed baseline records the **2** residual unsanctioned edges (measured 2026-09-18 @ `dev` `684e41e`, post-round-049):

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
```

**Behaviour intent**: **MODIFY (behaviour-preserving refactor)** — the round inverts the surviving `AgentLoop` **construction/execution** into an injected **domain port**, so `internal/cli` production references **zero** `internal/agent` identifiers. Per ADR 0018 §Forward, this round's DoD names **TWO** ratchet removals, not one (fold F-4): the RULE-E baseline line **and** the RULE-F `couplingSurface` key. **No** user-facing behaviour change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary); the round-044/045/046/048/049 lineage (gate-proven, behaviour-preserving structural refactors).

---

## Clarify round 1 — OPEN (asked one at a time)

> **Q1 → A is CLOSED** (recorded below); Q2/Q3 remain. Per the `/axb-specify` → `/axb-clarify` gate, questions are asked **one at a time**.

| # | Question (asked one at a time) | Status |
| --- | --- | --- |
| **Q1** | **How is the R-1 cross-slice coupling split?** The surviving `→ agent` site (`&agent.AgentLoop{…}`, `cli.go:699`) shares one ~47-line wiring block (`cli.go:699-745`) with **three** `→ ui` references (`ui.ToolLineRenderer{}` L716 · `*ui.Spinner` L721 · `ui.NewToolOutputCoordinator(...)` L739). Options: **(A)** invert only the `→ agent` construction into a domain port and **keep the `ui` wiring in the CLI** (baseline **2 → 1**; `→ ui` edge intentionally retained; edge-sized) · **(B)** move the **whole block** to the exempt `cmd/tellme` (pulls the `ui` construction into the root; the CLI's turn signature grows; ≥ edge-sized) · **(C)** **re-sequence** — do the `→ ui` value-types extraction **first**, then sub-slice 2 (this round is re-scoped / deferred). | ✅ **A** |
| **Q2** | **Port shape + adapter home** (under Q1 → A): a domain interface (`agentport.Loop` with `Run(ctx, prompt, prior) (Result, error)`) fed by an injected **factory** — adapter as an exported `internal/agent` constructor injected via `deps` — vs a domain **struct-of-funcs**, vs a `cmd/tellme` closure. | ⏳ TBD |
| **Q3** | **`Lines`/observer ownership**: does the CLI keep supplying `ui.ToolLineRenderer{}` as the domain `ToolLineRenderer` port (keeping one `→ ui` ref inside the block) and the composite observer, or does the port supply defaults? | ⏳ TBD |

### Q1 → A (LOCKED) — port-only inversion; the `ui` wiring stays in the CLI

The `→ agent` **construction** is inverted into a domain port; the `ui` wiring (the `Lines` renderer, the spinner, the tool-output coordinator, and the composite observer assembly) **stays in `internal/cli`** by design. Consequence (recorded up front): this round moves the RULE-E ratchet **2 → 1** (the `internal/cli -> internal/agent` line **and** its RULE-F `couplingSurface` key both go — ADR 0018 fold **F-4**), while the `internal/cli -> internal/ui` line **and** its 19-identifier surface stay **byte-identical**. The `→ ui` edge is the later slice (ADR 0017 §Forward — not edge-sized), **not** this round. Rejected: **B** (relocates 3 `ui` refs at the cost of growing the tier-exempt root and the `runTurn` signature, while the CLI still needs the `*ui.Spinner`-holding composite — no `→ ui` removal for the size); **C** (re-sequences away the baseline move ADR 0018 assigns to *this* round).

> **Assumptions (not escalated — low impact, disclosed):** exact port/type names; the ADR number (**0019**); the api/data/dsl-refine **NOOP** set; the `-count=1`/`-tags=arch` invocation stays as shipped.

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `684e41e` (post-round-049 static import scan; the gate re-measures at implementation).

### The surviving edge, its site, and its crossing identifier

| Item | Detail |
| --- | --- |
| Edge | `internal/cli -> internal/agent` — the **last** `→ agent` residual after round 049 |
| Import sites | `internal/cli/cli.go` (production) **only**; **no** `internal/cli/*_test.go` imports `internal/agent` (verified) |
| Crossing identifier (**1**) | `agent.AgentLoop` — the **constructed** loop, 9 fields: `Gateway`/`Registry`/`MaxLoops`/`EffectiveBudget`/`Stderr`/`Now`/`ToolUsage`/`Lines`/`Observer`; `cli.go:699` |
| The call | `loop.Run(ctx, prompt, prior)` → `(agentport.Result, error)`; the incomplete-turn path type-asserts `*agentport.ErrIncomplete` (both **already domain-owned** by round 049; `internal/cli/cli.go:754`) |
| The entangled block (R-1) | `cli.go:699-745` (~47 lines): the `&agent.AgentLoop{…}` literal, then `loop.Observer = compositeObserver{call: renderer, spinner: spinner}`, interleaved with **three** `→ ui` refs — `ui.ToolLineRenderer{}` (L716, the `Lines` field), `*ui.Spinner` (L721, via `newTurnSpinner`), `ui.NewToolOutputCoordinator(...)` (L739, the tool-output sink) |
| The `→ ui` edge (untouched target) | 19 production identifiers (`call_renderer.go` + `cli.go`); RULE-F `couplingSurface["internal/cli -> internal/ui"]` = those 19 |

### Why this round is edge-sized (post-sub-slice-1)

Round 049 (ADR 0018) extracted the loop's crossing **contracts** (`Result`/`ErrIncomplete`/`ToolDefs`) to `internal/domain/agent`, so the CLI's `→ agent` surface dropped **4 → 1** and is **machine-pinned** by RULE-F (`{AgentLoop}`). The only remaining coupling is **one construction call site** plus the `Run` call, whose result/error types are **already domain types** — so the inversion is a single-seam round (ADR 0017's port recipe: domain interface + tier-≥4 adapter + composition-root injection), *unless* Q1 chooses B/C.

### Tier constraint

ADR 0011 **D1** ranks `domain` 0 · `config`/`home` 1 · `app` 2 · `infrastructure` 3 · **`agent` 4** · `ui`/`ui/tui/**` 5 · **`cli` 6**. **RULE-A** forbids an *upward* import; `internal/cli` (tier 6) → `internal/domain/**` (tier 0) is **sanctioned** (downward), while `internal/cli → internal/agent` (tier 4) and `internal/cli → internal/ui` (tier 5) are **downward but unsanctioned** (RULE-E). `internal/domain/**` must stay **RULE-C-pure**; the adapter (tier 4 `internal/agent`, or the tier-table-exempt `cmd/tellme`) may import `internal/domain/**` (downward).

### The gate that measures the DoD (rounds 047/049 / ADR 0016 + 0018)

- **RULE-E**: a governed application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`; the sanctioned set is a **normative** section of the gate's tier table (ADR 0011 **D7**), **default-deny**, **fail-on-stale**.
- **RULE-F**: for a governed application edge, the set of identifiers the tier may **select** from the target package is a normative **fail-on-stale** allow-list (`couplingSurface`), with a coverage invariant (`assertSurfaceCoversBaseline`) and a synthetic self-test. **Metric**: distinct package-qualified **selectors**, not call sites.
- The committed **baseline** (`tools/arch/baseline.txt`) is a **fail-on-stale ratchet** (ADR 0011 **D3**).
- **DoD is the gate**, not the E2E suite (#92 **AC5**).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale** at baseline **1**; the `→ ui` line is byte-identical; **RULE-F** reports the `→ ui` surface as its **19** identifiers (0 new / 0 stale) and the `→ agent` key is **removed**; 0 cycles; cross-compile 4/4.
- **Zero behavioural change**: identical `stdout`/`stderr` byte-contracts, identical exit codes, unchanged flags, **no** new Gherkin/DSL row, **no** new dependency (`go.mod`/`go.sum` unchanged).
- **`internal/domain/**` stays 100 % pure**: the port references only stdlib + domain types; no `agent`/`ui` type crosses in (RULE-C).
- **No cycle**: the loop's inward direction (`internal/agent` → `internal/domain/**`) is preserved.
- **Single assembly site upheld by review discipline**: `cmd/tellme` is tier-table-exempt (ADR 0013; round-044 forward item (d)).

---

## User Scenarios & Testing *(mandatory)*

> **Scope note**: stories are written at the **outcome** level of the "invert the construction" family (hold under Q1 → A, and are the target of B); the **mechanism split** (where the `ui` wiring lives) is **Q1**, and the port shape/home is **Q2**.

### User Story 1 - the CLI runs a turn through an injected domain port, so it references zero `internal/agent` identifiers (Priority: P1)

As a maintainer, I want `internal/cli` to obtain the turn loop from an injected **domain** port (constructed at the composition root) rather than building `&agent.AgentLoop{…}` itself, so the `internal/cli → internal/agent` edge disappears and the RULE-E ratchet moves **2 → 1**.

**Why this priority**: it is the round's reason to exist — the baseline-moving slice of R5, and the last `→ agent` residual.

**Independent verification**: after the round, an import scan of `internal/cli` production sources finds **zero** `internal/agent` references; the gate reports the baseline at **1** with 0 new / 0 stale; the godog E2E + the turn unit pins stay green.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the `internal/cli` production sources are scanned, **Then** **no** `internal/agent` identifier is referenced (the loop is consumed through a domain-declared port).
2. **Given** a prompt-bearing turn runs, **When** it completes, **Then** the answer, the tool steps, the usage records, the exit codes, and the incomplete-turn tool-error path are **unchanged** (godog E2E + unit pins green).
3. **Given** the injected port, **When** declared in `internal/domain/**`, **Then** it references only stdlib + domain types (RULE-C-pure); the adapter constructs `agent.AgentLoop` (downward import) and no cycle is introduced.
4. **Given** the baseline, **When** the gate runs, **Then** the `internal/cli -> internal/agent` line is **gone** and the `internal/cli -> internal/ui` line is **unchanged** (baseline **2 → 1**, 0 new / 0 stale).

**Functional Requirements**:

- **FR-001**: The loop's **construction/execution** MUST be inverted into an injected **domain** port declared under `internal/domain/**`; the CLI obtains and runs the loop through the port and MUST reference **zero** `internal/agent` identifiers in production sources.
- **FR-002**: The port MUST reference only stdlib + domain types (RULE-C purity — no `agent`/`ui` type crosses the boundary); the adapter that constructs `agent.AgentLoop` MUST live at tier ≥4 (`internal/agent`) or at the composition root (`cmd/tellme`, tier-table-exempt), importing `internal/domain/**` **downward**; **no** new import cycle.
- **FR-003**: The implementation MUST be **behaviour-preserving**: the 9 constructed fields' semantics, the call-hook/observer ordering, the `errors.As` incomplete-turn classification, the `ToolDefs`-based pre-flight estimate, and the deferred final tail MUST be unchanged (byte-identical `stdout`/`stderr`, identical exit codes, unchanged flags). Existing tests MUST pass **unmodified** except where a test directly names a moved symbol.
- **FR-004**: The seam MUST be wired through the existing **`internal/app/deps.Dependencies`** injection (round 044 / ADR 0013), not a new global; the CLI MUST receive it as a domain-typed field.

**Non-Functional Requirements**:

- **NFR-001**: `internal/cli` MUST NOT gain any new non-sanctioned import; the round introduces **no** new `internal/cli → {agent, ui}` reference.

---

### User Story 2 - the two ratchet removals are made deliberately and the `→ ui` edge is provably untouched (Priority: P1)

As a maintainer, I want the baseline move to be the **expected two-part red** (the RULE-E line **and** the RULE-F key) and the `→ ui` surface to be **byte-stable**, so the round cannot launder the remaining edge or drift the untouched one.

**Why this priority**: the DoD is the gate (#92 AC5); ADR 0018 §Forward (fold F-4) warns the removal reds **twice** — the round must fold **both**, deliberately.

**Independent verification**: re-read `tools/arch/baseline.txt` — **1** line (`internal/cli -> internal/ui`); run the gate — the `→ ui` surface reports 19 identifiers, 0 new / 0 stale; the `→ agent` `couplingSurface` key is absent.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs, **Then** the RULE-E baseline is exactly **1** line (`internal/cli -> internal/ui`) and reports **0 new / 0 stale**.
2. **Given** the removed edge, **When** the RULE-F table is inspected, **Then** `couplingSurface["internal/cli -> internal/agent"]` is **absent** and `couplingSurface["internal/cli -> internal/ui"]` is **unchanged** (19 identifiers).
3. **Given** a change that leaves a **second** `→ agent` coupling (a partial inversion), **When** the gate runs, **Then** it is **red** — the round must present a single, complete inversion.

**Functional Requirements**:

- **FR-005**: The round MUST remove **both** ratchets in the same PR: the RULE-E baseline line `internal/cli -> internal/agent` **and** the RULE-F `couplingSurface` entry for that edge; the `internal/cli -> internal/ui` line **and** its 19-identifier surface MUST be **unchanged**.
- **FR-006**: Regenerating the baseline (`make verify-architecture-update`) MUST be the mechanism for the **RULE-E** line only; the **RULE-F** key is a **guard-table edit** (there is no regeneration affordance — ADR 0018 D6), so the two removals are distinct, deliberate edits.

**Non-Functional Requirements**:

- **NFR-002**: The inversion + the truth/ADR update + the two ratchet removals MUST land as **one PR** (atomicity); **no** new Makefile target.

---

### User Story 3 - the port home, the ADR, and the follow-on caveats are recorded (Priority: P2)

As a maintainer, I want the port's home, the ADR, and the two carried caveats (the `Validate()` interface-seam blindness; the sequencing of the remaining `→ ui` edge) recorded, so the next slice inherits an accurate state.

**Why this priority**: value is in *not re-litigating* the settled seams; the round does not change user-visible behaviour.

**Independent verification**: the ADR exists + is indexed; `specs/truth/techstack.md` records the new row state; the `truth-delta.md` records the owners' entries.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the ADR index is checked, **Then** the new ADR (**0019**) is present and unique.
2. **Given** the port, **When** it adds an **interface-typed** field to `deps.Dependencies`/`cli.Options`, **Then** that field has its own `Validate()` assertion (ADR 0017 §Forward — `Kind()==reflect.Func` is blind to interface seams).

**Functional Requirements**:

- **FR-007**: A new **ADR 0019** MUST record the inversion decision (port home, adapter home, the two removals, what the change is *not*, and the relation to ADR 0011/0016/0017/0018), and be indexed in `docs/decisions/README.md`.
- **FR-008**: `specs/truth/techstack.md` MUST be updated (the layer-discipline gate row's baseline figure **2 → 1**, and the agent-tool-loop row's construction-home note) through the round's `truth-delta.md`.
- **FR-009**: If the seam adds an **interface-typed** field to `deps.Dependencies` or `cli.Options`, it MUST carry a `Validate()` assertion; a **func-typed** field is covered by the existing `Kind()==reflect.Func` predicate.

---

### Global Requirements *(cross-story only)*

- **FR-010**: The round MUST introduce **no** new dependency (`go.mod`/`go.sum` unchanged) and **no** new Gherkin/DSL row or topology change.
- **FR-011**: Falsifiability witnesses MUST be reproduced then reverted (ADR 0010): (a) **compile-level** — delete/rename the port seam ⇒ the CLI turn path fails to compile; (b) **count-level (RULE-F)** — reintroduce a direct `agent.*` selector in the CLI (or a second construction) ⇒ RULE-F/RULE-E **FAILS** while the rest stays green; (c) baseline-level — the removal reds **twice** (RULE-E line + RULE-F key) until both are folded.
- **FR-012**: The round MUST NOT re-open R2's frozen decisions (composition-root home `cmd/tellme`, the `internal/app/deps.Dependencies` shape, `agentTools()` relocation, MCP orchestration) nor ADR 0014/0015 observer/presentation ports, nor the `→ ui` slice's scope.

---

## Edge Cases

- **Partial inversion** — a second `→ agent` coupling survives (e.g. a leftover `agent.X` helper) ⇒ RULE-E reports **0 new / 0 stale** while RULE-F **FAILS** with *"NEW coupling-surface identifier(s)"* ⇒ rejected (US2 scenario 3).
- **Interface seam blindness** — if the port is stored as an **interface**-typed field, `Dependencies.Validate()`'s `Kind()==reflect.Func` predicate cannot see a nil ⇒ an explicit assertion is required (FR-009; ADR 0017 §Forward).
- **`ui` drift** — moving the block touches `ui.*` call sites ⇒ the `→ ui` line/surface must remain **19 / byte-identical**; any drift FAILS RULE-F (staleness).
- **Observer ordering** — the composite (`call` renderer + spinner) fires the call hooks; the inversion must preserve the observer's injection point (`loop.Observer = …`) exactly, or the per-call frame/tail bytes drift.
- **Checksum/tool-usage seams** — `Now` and `ToolUsage` (the `~/.tellme` sink) are injected today; they must cross the port as domain/stdlib-typed seams (no `ui`/`agent` leak).
- **Build-tag scope** — the RULE-F/RULE-E scans cover **production** sources only; a `_test.go` reference is unguarded (recorded residual, ADR 0018 F-3).
- **Cycles** — the adapter (tier 4) → domain is downward; the port must not be declared in a package the loop imports upward.

## Key Entities

- **Turn loop (the object under inversion)** — `agent.AgentLoop` + `Run(ctx, prompt, prior) (agentport.Result, error)`; 9 constructed fields.
- **Domain loop port (new)** — the injected seam declared under `internal/domain/**` (peer of `agentport.LoopObserver`/`CallObserver`/`ToolLineRenderer`); RULE-C-pure.
- **Command emission spec** — unchanged (flags, exit codes, `stdout`/`stderr` byte-contracts).
- **Layer-discipline ratchet** — `tools/arch/baseline.txt` (**2 → 1**) + RULE-F `couplingSurface` (`→ agent` key removed).
- **Governance** — ADR 0019 (new) + `specs/truth/techstack.md` rows.

## Success Criteria

- **SC-001**: An import scan of `internal/cli` **production** sources finds **0** `internal/agent` references.
- **SC-002**: `make verify` green with the RULE-E baseline at exactly **1** line (`internal/cli -> internal/ui`), **0 new / 0 stale**; RULE-A/B/C **0**; **0** cycles; cross-compile 4/4.
- **SC-003**: RULE-F reports `→ ui` at **19** identifiers (0 new / 0 stale) and no `→ agent` key.
- **SC-004**: `go test -count=1 ./...` green (including the godog E2E) with **no** behavioural assertion changed except a test naming a moved symbol.
- **SC-005**: `go.mod`/`go.sum` unchanged; **no** new Gherkin/DSL row; the topology audit is unchanged.
- **SC-006**: The two removals red **twice** pre-fold (RULE-E line + RULE-F key) and both are folded deliberately.

## Assumptions

- **A1**: The R-1 block is genuinely **one** construction site (re-measured at implementation; the gate + RULE-F are the authority).
- **A2**: Follows the rounds 020/031/036/041–049 **non-BDD-tooling** precedent → `/axb-spec-by-example` is **NOOP**; the carrier is the gate + unit seams (#92 AC5).
- **A3**: Port/type names are RD decisions (exact identifiers not fixed here).
- **A4**: The ADR number is **0019** (next free); adjust if taken at implementation.
- **A5**: `/axb-api-plan` and `/axb-data-plan` are **NOOP** (no API surface; no persisted/runtime-state change).
- **A6**: `/axb-dsl-refine` is **NOOP** (an internal refactor is not a `tellme` CLI-contract change; the Gherkin/DSL topology is unchanged).
- **A7**: Q1 → A keeps the `→ ui` edge (by design); B grows the root + the turn signature; C re-sequences the round.

---

## Out of scope (recorded forward items)

- The **`internal/cli → internal/ui`** edge (the last RULE-E residual; ≥ edge-sized — ADR 0017 §Forward; likely a re-cut) — unless **Q1 → C** re-sequences.
- **F-6** (narrow `Dependencies` seams; "≤2 fields" acceptance) · **F-7** (a named `Discovery{Tools, Warnings, Closer io.Closer}`) · **F-8** (`domaintools.OutputSink` struct-of-funcs → interface) — remain on [#101](https://github.com/gosharplite/tellme/issues/101).
- No user-facing CLI change, no adapter/domain rewrites, no Windows, no security/consent layer, no conversation pruning, no tool-call concurrency.
