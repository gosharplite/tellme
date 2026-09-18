# Feature Specification: close [#101](https://github.com/gosharplite/tellme/issues/101) — de-couple `internal/cli` from `internal/ui` (the last RULE-E residual) + the F-6/F-7/F-8 deferrals (round 051)

**Feature Branch**: `051-cli-ui-decoupling`

**Created**: 2026-09-18

**Status**: Draft (clarify round 1 in progress — **Q1 → D** closed; Q2/Q3/Q4 pending) — plan package created by `/axb-specify`.

**Input**: [#101](https://github.com/gosharplite/tellme/issues/101) — **R5** of [#92](https://github.com/gosharplite/tellme/issues/92). The committed baseline records the **1** residual unsanctioned edge (measured 2026-09-18 @ `dev` `310def4`, post-round-050):

```text
internal/cli -> internal/ui
```

**Programme goal (operator-declared)**: **close [#101](https://github.com/gosharplite/tellme/issues/101)** — i.e. (i) remove the **last** RULE-E residual (`cli → ui`; baseline **1 → 0**), so `internal/cli` names **only** `internal/domain/**` + stdlib + the RULE-E sanctioned set; and (ii) resolve the three PR [#104](https://github.com/gosharplite/tellme/pull/104) deferrals **F-6 / F-7 / F-8**.

**Behaviour intent**: **MODIFY (behaviour-preserving structural refactor)** — no user-facing change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary); the rounds 044–050 lineage (gate-proven, behaviour-preserving de-couplings). The `→ ui` edge is **not edge-sized** (ADR 0017 §Forward), so it is **re-cut** — Q1 decides this round's place in the closing programme.

---

## Clarify round 1 — OPEN (asked one at a time)

> **Q1 → D is CLOSED** (the one-round close; recorded below). Q2/Q3/Q4 remain — asked **one at a time**.

| # | Question (asked one at a time) | Status |
| --- | --- | --- |
| **Q1** | **The closing-programme shape: what is round 051?** The remaining #101 work is 3 items of very different size. Options: **(A) Re-cut — round 051 = sub-slice 1** (the 3 crossing **value types** → `internal/domain/**`; behaviour-preserving; the RULE-F `→ ui` surface shrinks but the **edge remains** so the baseline **stays 1**; later sub-slices then do the formatters + factories). **(B) Re-cut — round 051 = the `→ ui` collapse** (invert the CLI's ui rendering — the `call_renderer` + the `ui.Format*` calls + the stateful factories — behind injected **domain ports** implemented by `internal/ui`; baseline **1 → 0**), with **F-6/F-7/F-8 in a following round 052**. **(C) F-6/F-7/F-8 first** (small, self-contained; the `→ ui` collapse in round 052). **(D) One mega-round** closing all of #101 at once (the `→ ui` collapse **and** F-6/F-7/F-8) — highest risk. | ✅ **D** |
| **Q2** | **Value-type homes**: **(i)** reuse existing domain packages (`ui.Pricing`→`internal/domain/llm`, `ui.UsageCounts`→`internal/domain/metrics`, `ui.ToolUsageRow`→`internal/domain/history`, folded onto `history.ToolUsageCounts`; `ComputeCost`/`HitRate` move with `Pricing` since persistence consumes them) · **(ii)** a new `internal/domain/pricing` for money · **(iii)** keep them ui-owned and invert everything through the port. | ✅ **(i)** |
| **Q2** | **Value-type homes** (if any option above needs them): `ui.Pricing` → `internal/domain/llm` vs a new `internal/domain/pricing`; `ui.UsageCounts` → `internal/domain/metrics`; `ui.ToolUsageRow` → `internal/domain/history` (it mirrors `history.ToolUsageCounts`). | ⏳ TBD |
| **Q3** | **The `→ ui` inversion mechanism**: a **single** injected presentation port (e.g. `domain/render.StatusRenderer` covering the turn frame/tail + tool-usage) vs **several** narrow ports (status / metrics / tool-usage) vs moving `call_renderer` wholesale into `internal/ui` behind one port. | ⏳ TBD |
| **Q4** | **F-6/F-7/F-8**: fold them into this round (if the shape allows) or keep them a separate round; and for F-8 (`domaintools.OutputSink` struct-of-funcs → interface) — is the RULE-C-clean domain change acceptable now. | ⏳ TBD |

### Q1 → D (LOCKED) — one round that closes [#101](https://github.com/gosharplite/tellme/issues/101)

Round 051 carries the **whole** remaining #101 scope in **one** delivery: (i) the `→ ui` de-coupling (baseline **1 → 0**, with the two ratchet removals) **and** (ii) the F-6/F-7/F-8 deferrals. Internal **sub-slice ordering is an RD decision** (Q3), and the operator's recorded caveat holds: if `/axb-system-analysis` shows the blast radius exceeds a single reviewable PR, the round is re-cut internally (still closing #101 across the ordered slices). Rejected (recorded): **A/C** (do not close #101 this round); **B** (closes only the edge, defers F-6/F-7/F-8).

### Q2 → (i) (LOCKED) — reuse the existing domain packages; fold the tool-usage row

The three crossing **value types** re-home into existing domain peers (no new package): `Pricing` (+ the pure `ComputeCost`/`HitRate` arithmetic) → **`internal/domain/llm`** (cost is a provider/model concern; persistence consumes it too, so it cannot live only behind a render port); `UsageCounts` → **`internal/domain/metrics`**; `ToolUsageRow` → **`internal/domain/history`**, **folded onto** the existing `history.ToolUsageCounts` (or an alias — one concept, one name). Rejected: **(ii)** a new `internal/domain/pricing` (over-structure for ~2 functions); **(iii)** ui-owned + fully port-inverted (pushes pricing arithmetic behind a *render* port and breaks `ComputeCost`'s use in persistence).

> **Assumptions (not escalated — low impact, disclosed):** the ADR number (**0020**); the api/data/dsl-refine **NOOP** set; the `-count=1`/`-tags=arch` invocation stays as shipped; the sanctioned set (domain/config/home/app) is **not** re-ruled (no `internal/pkg` is introduced — ADR 0016 D1 governs).

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `310def4` (post-round-050 static read; the gate re-measures at implementation).

### The residual edge and its exact surface (RULE-F: **19 identifiers**, ~33 call sites)

| Kind | Identifiers the CLI names | Detail |
| --- | --- | --- |
| **Value types (3)** | `ui.Pricing` · `ui.UsageCounts` · `ui.ToolUsageRow` | `Pricing` is built from `config.PricingRates` (`Hit`/`Miss`/`Comp`); `UsageCounts` is a metrics projection; `ToolUsageRow` mirrors the domain `history.ToolUsageCounts` |
| **Pure formatters / consts (11)** | `ui.ComputeCost` · `ui.HitRate` · `ui.FormatInputCaptured` · `ui.FormatMetrics` · `ui.FormatPayloadStatus` · `ui.FormatReady` · `ui.FormatToolReason` · `ui.FormatToolUsage` · `ui.FormatTurnGap` · `ui.FormatTurnOpening` · `ui.DefaultToolOutputIdleGap` | all pure (bytes); return `string`/`float64`; the caps/sanitize policy is single-owned in `internal/ui` (ADR 0006/0008/0015) |
| **Stateful objects / factories (4)** | `ui.Renderer`+`ui.NewRenderer` · `ui.Spinner`+`ui.NewSpinner` · `ui.ToolOutputCoordinator`+`ui.NewToolOutputCoordinator` · `ui.ToolLineRenderer{}` | the answer renderer, the progress spinner, the `[Tool Output]` coordinator, and the loop's tool-line renderer (already an `agentport.ToolLineRenderer` **port implementation**) |
| **Call sites** | `internal/cli/cli.go` + `internal/cli/call_renderer.go` (production); **no** `internal/cli/*_test.go` imports `internal/ui` | `call_renderer.go` (190 lines) is the CLI-side per-call renderer that calls the `ui.Format*` formatters |

### Why removing the edge is not edge-sized (ADR 0017 §Forward)

≈20 call sites · **3 crossing value types** · **4 stateful objects** · **8 pure formatters**. The recipe ADR 0017 prescribes — and ADR 0018/0019 applied to `→ agent` — is **values/types first, then the factory**. Hence the Q1 re-cut options.

### Tier + sanctioned set (ADR 0011 D1 / ADR 0016)

`domain` 0 · `config`/`home` 1 · `app` 2 · `infrastructure` 3 · `agent` 4 · `ui`/`ui/tui/**` 5 · `cli` 6. **RULE-E** allows a governed application tier (`internal/app/**`, `internal/cli`), beyond stdlib, **only** `internal/domain/**` + `internal/config` + `internal/home` + `internal/app/**` — a **normative, default-deny, fail-on-stale** allow-list. There is **no `internal/pkg`** in tellme (so the "move a pure utility to `internal/pkg`" escape does not exist here). `internal/cli → internal/ui` (tier 6 → 5) is **downward but unsanctioned** → the last RULE-E violation.

### The F-6/F-7/F-8 deferrals (PR [#104](https://github.com/gosharplite/tellme/pull/104) review `5243584043`; still open, verified)

| ID | Item | Current state |
| --- | --- | --- |
| **F-6** | `deps.Dependencies` is threaded as a **wide bag** into ~10 functions (`newCallRenderer`, `persistTurnUsage`, `newTurnSpinner`, `renderToolUsage`, `renderHistoryList`, `dispatchReporting`, `runInteractiveTUI`, `renderTurn`, `runTurn`) | still a wide bag; acceptance: *"every function receiving `Dependencies` reads ≤2 fields."* |
| **F-7** | `MCPDiscoverer`'s third result is a bare `func()` | `MCPDiscoverer func(...) (tools []domaintools.Tool, warnings []string, close func())` — should return a named `Discovery{Tools, Warnings, Closer io.Closer}` |
| **F-8** | `domaintools.OutputSink` is a **struct-of-funcs in the domain**, rebuilt per turn | `type OutputSink struct { Begin func(); Writer io.Writer; End func() }` — should be an **interface** the coordinator satisfies directly (+ `Enabled()`); the gate is **structurally blind** here (domain→`io` is RULE-C-clean) |

### The gate that measures the DoD (rounds 047/049 + ADR 0016/0018)

- **RULE-E**: the application import ceiling (default-deny; fail-on-stale) — the `cli → ui` line is the **only** baselined violation.
- **RULE-F**: the application **coupling surface** — a fail-on-stale per-edge identifier allow-list (`couplingSurface["internal/cli -> internal/ui"]` = the 19 identifiers), with a coverage invariant.
- Both removals land together when the edge disappears (the **two-removal** precedent — ADR 0018 fold F-4, exercised by round 050).
- **DoD is the gate**, not the E2E suite (#92 **AC5**).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale** at whatever baseline the round leaves; **RULE-F** consistent; 0 cycles; cross-compile 4/4.
- **Zero behavioural change**: identical `stdout`/`stderr` byte-contracts, identical exit codes, unchanged flags, **no** new Gherkin/DSL row.
- **`internal/domain/**` stays 100 % pure** (RULE-C).
- **No cycle**; the `internal/ui` package keeps the byte/sanitize/cap single-ownership (ADR 0006/0008/0015).
- **At baseline 0** the ratchet has **no release valve** (ADR 0011/0016) — a future legitimate ceiling violation must be refactored, never baselined.

---

## User Scenarios & Testing *(mandatory)*

> **Scope note**: the stories are written at the **programme level** (#101 closes); Q1 decides the round's slice and therefore which stories this round carries. Every story is behaviour-preserving.

### User Story 1 - `internal/cli` names no `internal/ui` identifier, so the RULE-E ratchet reaches 0 (Priority: P1)

As a maintainer, I want the CLI to obtain every ui-provided value, formatter, and object through **sanctioned** packages or injected **domain ports**, so the `internal/cli → internal/ui` edge disappears and the RULE-E baseline drops **1 → 0** — satisfying [#92](https://github.com/gosharplite/tellme/issues/92) AC2's second clause.

**Why this priority**: it is the last RULE-E residual; without it the gate can never be green at 0 and AC2 clause 2 cannot hold.

**Independent verification**: after the round(s), an import scan of `internal/cli` production sources finds **0** `internal/ui` references; the gate reports the baseline at **0** with 0 new / 0 stale; godog E2E + unit pins green.

**Acceptance Scenarios**:

1. **Given** the closing round has landed, **When** the CLI production sources are scanned, **Then** **no** `internal/ui` identifier is referenced.
2. **Given** a prompt-bearing turn runs, **When** it completes, **Then** the answer, the tool steps, the usage records, the exit codes, the status frames/tails, and the tool-usage report are **byte-unchanged** (godog E2E + unit pins green).
3. **Given** the port(s), **When** declared in `internal/domain/**`, **Then** they reference only stdlib + domain types (RULE-C-pure); the `internal/ui` adapters construct them (tier 5) and no cycle is introduced.
4. **Given** the baseline, **When** the gate runs, **Then** the `internal/cli -> internal/ui` line is **gone** (baseline **0**) and the RULE-F `couplingSurface` key for that edge is **removed**.

**Functional Requirements**:

- **FR-001**: The CLI's production `internal/ui` references (the 19 identifiers) MUST be removed — by re-homing the crossing **value types** to `internal/domain/**` and/or inverting the formatters + stateful objects behind injected **domain port(s)** implemented by `internal/ui`.
- **FR-002**: Any new port MUST be declared under `internal/domain/**` and reference only stdlib + domain types (RULE-C purity); the `internal/ui` adapter (tier 5) is wired at the composition root (`cmd/tellme`, tier-exempt); **no** new import cycle.
- **FR-003**: The implementation MUST be **behaviour-preserving**: byte-identical `stdout`/`stderr`, identical exit codes, unchanged flags; the sanitize/cap policy stays single-owned in `internal/ui` (ADR 0006/0008/0015).
- **FR-004**: The seam(s) MUST be wired through the existing **`internal/app/deps.Dependencies`** injection (round 044 / ADR 0013); the CLI MUST receive domain-typed seams.

**Non-Functional Requirements**:

- **NFR-001**: `internal/cli` MUST NOT gain any new non-sanctioned import; the round introduces **no** new `internal/cli → {ui, agent, infrastructure}` reference.

---

### User Story 2 - the two ratchet removals are made deliberately (Priority: P1)

As a maintainer, I want the removal to fold **both** the RULE-E baseline line **and** the RULE-F `couplingSurface` key in the same PR (the ADR 0018 fold-F-4 pattern), and the gate to stay green, so the ratchet cannot be laundered.

**Why this priority**: the DoD is the gate (#92 AC5); the two-removal discipline is the recorded lesson.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs, **Then** the RULE-E baseline is **0** lines and reports **0 new / 0 stale**.
2. **Given** the edge is gone, **When** the RULE-F table is inspected, **Then** `couplingSurface["internal/cli -> internal/ui"]` is **removed** and the RULE-F coverage invariant stays green.
3. **Given** a partial removal, **When** the gate runs, **Then** it is **red** (RULE-E stale line **and/or** a RULE-F surface/F coverage mismatch) — the round must present a complete removal.

**Functional Requirements**:

- **FR-005**: The round MUST remove the RULE-E baseline line **and** the RULE-F `couplingSurface` key (two distinct, deliberate edits — `verify-architecture-update` regenerates only the baseline; the RULE-F table is a guard-table edit).

---

### User Story 3 - the F-6/F-7/F-8 deferrals are resolved (Priority: P2)

As a maintainer, I want the three PR #104 review deferrals closed so [#101](https://github.com/gosharplite/tellme/issues/101) can close.

**Why this priority**: they are the second half of #101's remaining scope; each is a small, self-contained refactor.

**Acceptance Scenarios**:

1. **Given** the round(s) have landed, **When** the `Dependencies`-threaded functions are inspected, **Then** each reads **≤2** fields (F-6).
2. **Given** the MCP discovery seam, **When** inspected, **Then** it returns a named `Discovery{Tools, Warnings, Closer io.Closer}` (F-7).
3. **Given** `domaintools.OutputSink`, **When** inspected, **Then** it is an **interface** the coordinator satisfies directly (F-8).

**Functional Requirements**:

- **FR-006**: F-6 — narrow the `Dependencies` seams so a leaf function reads ≤2 fields (or a recorded equivalent), with a falsifiable acceptance.
- **FR-007**: F-7 — replace the bare `func()` close result with a named `Discovery` type exposing the closer as a visible field.
- **FR-008**: F-8 — make `domaintools.OutputSink` an **interface** (with `Enabled()`), removing the per-turn struct-of-funcs bridge.

---

### Global Requirements *(cross-story only)*

- **FR-009**: A new **ADR 0020** MUST record the round's decision(s) (the `→ ui` de-coupling mechanism + any F-6/F-7/F-8 decisions), indexed in `docs/decisions/README.md`; truth `specs/truth/techstack.md` rows updated through `truth-delta.md`.
- **FR-010**: The round MUST introduce **no** new dependency (`go.mod`/`go.sum` unchanged) and **no** new Gherkin/DSL row or topology change.
- **FR-011**: Falsifiability witnesses MUST be reproduced then reverted (ADR 0010): (a) **compile-level** — delete/rename the port seam ⇒ the CLI path fails to compile; (b) **count-level** — reintroduce a direct `ui.*` selector in CLI production ⇒ RULE-E/RULE-F **FAILS**; (c) baseline-level — the removal reds **twice** until both halves are folded.
- **FR-012**: The round MUST NOT re-open R2/`#101` frozen decisions (composition-root home `cmd/tellme`, the `Dependencies` shape, `agentTools()` relocation, MCP orchestration) nor the ADR 0014/0015/0017/0018/0019 ports (reused, not re-litigated).

---

## Edge Cases

- **Partial removal** — a leftover `ui.*` selector ⇒ RULE-E new/stale or RULE-F surface mismatch ⇒ rejected (US2 scenario 3).
- **Port-shape leak** — a domain port carrying a `ui` type ⇒ RULE-C violation; the port must carry domain types only (round-046 ADR 0015 recorded the port-shape-leak risk — do not repeat it).
- **Formatter ownership** — the sanitize/cap policy must stay single-owned in `internal/ui`; inverting the call must not duplicate the bytes logic into the CLI or the domain.
- **Stateful-object lifecycle** — the spinner (round 019/025/040 epochs) and the `[Tool Output]` coordinator (lock order, mutual exclusion + join) must keep their exact lifecycle when moved behind a port.
- **Render degraded path** — `ui.Renderer.Render` returns `(string, bool)`; the warn-degraded path must be preserved.
- **Interface-seam `Validate()`** — a new **interface**-typed `Dependencies` field is invisible to `Kind()==reflect.Func` (ADR 0017 §Forward) ⇒ needs its own `Validate()` assertion; a func-typed field does not.
- **No release valve** — after baseline 0 a future violation must be refactored, never baselined.

## Key Entities

- **The `internal/ui` surface** (19 identifiers) — value types, pure formatters, stateful objects/factories.
- **Domain presentation port(s) (new)** — the seam(s) declared under `internal/domain/**` through which the CLI obtains ui-provided rendering; RULE-C-pure.
- **`internal/ui` adapters** — the tier-5 implementations of those port(s), wired at `cmd/tellme`.
- **The `deps.Dependencies` seam** — the injection point (round 044 / ADR 0013).
- **The layer-discipline ratchet** — `tools/arch/baseline.txt` (→ **0**) + the RULE-F `couplingSurface`.
- **Governance** — ADR 0020 (new) + `specs/truth/techstack.md` rows.

## Success Criteria

- **SC-001**: An import scan of `internal/cli` production sources finds **0** `internal/ui` references.
- **SC-002**: `make verify` green with the RULE-E baseline at **0** lines (**0 new / 0 stale**); RULE-A/B/C **0**; **0** cycles; cross-compile 4/4.
- **SC-003**: RULE-F reports no `internal/cli -> internal/ui` key; the coverage invariant stays green.
- **SC-004**: `go test -count=1 ./...` green (incl. the godog E2E) with **no** behavioural assertion changed except where a test names a moved symbol.
- **SC-005**: `go.mod`/`go.sum` unchanged; **no** new Gherkin/DSL row; the topology audit unchanged.
- **SC-006**: F-6/F-7/F-8 each have a falsifiable acceptance met (per US3).

## Assumptions

- **A1**: The `→ ui` surface is the measured 19 identifiers / ~33 call sites (re-measured at implementation; the gate is the authority).
- **A2**: The rounds 020/031/036/041–050 **non-BDD-tooling** precedent holds → `/axb-spec-by-example` is **NOOP**; the carrier is the gate + unit seams (#92 AC5).
- **A3**: Port/type names are RD decisions (exact identifiers not fixed here).
- **A4**: The ADR number is **0020** (next free); adjust if taken at implementation.
- **A5**: `/axb-api-plan` and `/axb-data-plan` are **NOOP** (no API surface; no persisted/runtime-state change).
- **A6**: `/axb-dsl-refine` is **NOOP** (a `cli`-side refactor is not a tellme CLI-contract change).
- **A7**: The RULE-E sanctioned set is **not** re-ruled; no `internal/pkg` is introduced.

---

## Out of scope (recorded forward items)

- Any **user-facing** change: flags, exit codes, stream contracts, formatting, the 11-phrase DSL vocabulary.
- Rewriting infrastructure adapters or domain business logic.
- The #92 **records** (permanent E2E narrowing; one-concurrent-block; `End`-while-write-stalled) — stay records, not work.
- Settled exclusions: no Windows, no security/consent layer, no conversation pruning, no tool-call concurrency.
