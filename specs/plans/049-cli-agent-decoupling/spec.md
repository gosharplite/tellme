# Feature Specification: de-couple `internal/cli` from the turn loop — re-cut sub-slice 1: extract the loop's domain-facing contracts (round 049)

**Feature Branch**: `049-cli-agent-decoupling`

**Created**: 2026-09-18

**Status**: Draft — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)). **Clarify round 1**: **Q1 → B** (re-cut; locked) · **Q2 PENDING** (the sub-slice boundary / contract set). Round 047 = **R5.1** (the RULE-E gate + baseline); round 048 = **R5.2** (the `cli → ui/tui/prompt` edge).

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5**. The committed baseline records the **2** residual unsanctioned edges (measured 2026-09-18 @ `dev` `12964d6`):

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
```

**Behaviour intent**: **MODIFY (behaviour-preserving refactor)** — the round executes the **first sub-slice** of the re-cut `cli → agent` de-coupling; it **moves the loop's domain-facing crossing contracts** into `internal/domain/**` so a later slice can invert the loop construction **edge-sized**. **No** user-facing behaviour change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary) and no domain business logic — the round-044/045/046/048 lineage (gate-proven, behaviour-preserving structural refactors).

---

## Locked decisions (clarify round 1)

> Locked **one at a time**. **Q1 is locked**; **Q2 is pending**. No `NEEDS CLARIFICATION` remains **outside** the pending Q2.

| # | Decision |
| --- | --- |
| **Q1 → (B) re-cut the `cli → agent` de-coupling** | The `cli → agent` edge is **not** inverted in one round. ADR 0017 §Forward classifies it as *"the deepest slice"* (measured: **4** crossing identifiers over 2 files — `agent.AgentLoop` constructed with 9 fields, `agent.AgentResult`, `agent.ErrIncomplete` type-asserted, `agent.ToolDefs` — versus round-048's single edge-sized seam). The de-coupling is therefore **re-cut into ordered sub-slices**, mirroring ADR 0017's own recipe for the `→ ui` edge (*"value types → `internal/domain/**` first, then the … factory"*): **sub-slice 1 (this round) = extract the loop's crossing *contracts* to `internal/domain/**`**; **sub-slice 2 (later) = invert the loop *construction/execution* into an injected domain port** (the round that actually drops the edge; baseline **2 → 1**). |

### ⚠️ Consequence of Q1 → B (recorded up front)

Under the re-cut, **this round does NOT move the RULE-E baseline**: after sub-slice 1, `internal/cli` still imports `internal/agent` — for the `AgentLoop` **construction** — so the edge persists and the baseline stays exactly as it is (**2**). The round's DoD is therefore **not** "baseline 2 → 1" but **"the crossing *contracts* are domain-owned, behaviour-preservingly, and the surviving `cli → agent` coupling is reduced to a single construction call site — gate still green (0 new / 0 stale, baseline byte-identical)"**. This is a deliberate, recorded **preparatory** round; the baseline-moving round is sub-slice 2.

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `12964d6` (static import scan; the gate re-measures at implementation).

### The re-cut edge, its files, and its crossing identifiers

| Item | Detail |
| --- | --- |
| Edge | `internal/cli -> internal/agent` |
| Import sites | `internal/cli/cli.go`, `internal/cli/call_renderer.go` (production); **no** `internal/cli/*_test.go` imports `internal/agent` |
| Crossing identifiers (4) | `agent.AgentLoop` (the **constructed** loop, 9 fields: `Gateway`/`Registry`/`MaxLoops`/`EffectiveBudget`/`Stderr`/`Now`/`ToolUsage`/`Lines`/`Observer`; `cli.go:699`) · `agent.AgentResult` (the returned turn-result type; `cli.go:~749`, `call_renderer.go:175`) · `agent.ErrIncomplete` (the incomplete-turn error, type-asserted; `cli.go:754`) · `agent.ToolDefs` (the wire-def projection free function; `call_renderer.go:77`) |
| **Contracts** (move in sub-slice 1) | `agent.AgentResult` — `struct { Answer string; Steps []history.Step; Usage llm.Usage; Calls []llm.Usage }` (`agentloop.go:52`) · `agent.ErrIncomplete` — `struct { … }; Error()/Unwrap()` (`agentloop.go:30-44`) · `agent.ToolDefs(reg tools.Registry) []llm.ToolDef` (`agentloop.go:209`, a **pure** projection) |
| **The stateful object** (invert in sub-slice 2) | `agent.AgentLoop` + `func (a *AgentLoop) Run(ctx, prompt, prior) (AgentResult, error)` (`agentloop.go:112`) — the surviving single call site after sub-slice 1 |
| Existing domain ports the loop already consumes | `agentport.LoopObserver` (`internal/domain/agent/observer.go`) · `agentport.CallObserver` (`internal/domain/agent/call.go`) · `agentport.ToolLineRenderer` (`internal/domain/agent/presenter.go`) — the loop already imports **downward** into `internal/domain/**`, the natural home for the extracted contracts |

### Why the contracts are the right first cut

`internal/cli` consumes the loop through **values** (a result struct, an error type) and a **pure projection** (`ToolDefs`), plus **one stateful construction**. The values + projection are exactly what a domain port must eventually carry (so the port can be **RULE-C-pure** and leak no `agent` type); extracting them **first** (a) shrinks sub-slice 2 to a single, edge-sized construction inversion, and (b) is independently shippable and behaviour-preserving. This is the ADR 0017 `→ ui` re-cut recipe applied to `→ agent`.

### Tier constraint

ADR 0011 **D1** ranks `domain` 0 · `config`/`home` 1 · `app` 2 · `infrastructure` 3 · **`agent` 4** · `ui`/`ui/tui/**` 5 · `cli` 6. **RULE-A** forbids an *upward* import: `internal/cli` (tier 6) may import `internal/domain/**` (tier 0) — so after sub-slice 1 the CLI's contract imports are **sanctioned**; only the `AgentLoop` construction keeps the `→ agent` edge. `internal/domain/**` must stay **RULE-C-pure**; `internal/agent` (tier 4) may import `internal/domain/**` (downward).

### The gate that measures the DoD (round 047 / ADR 0016)

- **RULE-E**: a governed application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`; the sanctioned set is a **normative** section of the gate's tier table (ADR 0011 **D7**), **default-deny**, **fail-on-stale**.
- The committed **baseline** (`tools/arch/baseline.txt`) is a **fail-on-stale ratchet** (ADR 0011 **D3**).
- **DoD is the gate**, not the E2E suite (#92 **AC5**).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale**; the baseline is **byte-identical** (still the 2 lines); 0 cycles; cross-compile 4/4.
- **Zero behavioural change**: the turn path produces identical `stdout`/`stderr` byte-contracts, identical exit codes, unchanged flags, **no** new Gherkin/DSL row.
- **`internal/domain/**` stays 100 % pure**: the extracted contracts reference only stdlib + domain types; no `agent` type crosses in (RULE-C).
- **The loop's inward direction is preserved** (`internal/agent` → `internal/domain/**`), and **no cycle** is introduced.

---

## User Scenarios & Testing *(mandatory)*

> **Scope note**: stories written for Q1 option **B**, sub-slice 1 (contracts → domain). The exact contract set + the alias/reference choice are **Q2 / RD**.

### User Story 1 - the loop's crossing contracts are domain-owned, so the CLI reads them from `internal/domain/**` (Priority: P1)

As a maintainer, I want the turn-result type, the incomplete-turn error type, and the tool-def projection to live in `internal/domain/**`, so `internal/cli` obtains those contracts from the **domain** (sanctioned) rather than `internal/agent`, leaving only the `AgentLoop` construction on the `→ agent` edge.

**Why this priority**: it is the round's reason to exist — re-cut sub-slice 1, shrinking the surviving coupling to one edge-sized call site.

**Independent verification**: after the round, the `agent.*` identifiers referenced by `internal/cli` production drop from **4 → 1** (`agent.AgentLoop` only); the turn-result/error/projection are imported from `internal/domain/**`; the godog E2E + the turn unit pins stay green; the gate is green with the baseline unchanged.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the CLI's production sources are scanned, **Then** the only surviving `internal/agent` identifier is `agent.AgentLoop` (the construction); the turn-result/error/`ToolDefs` are consumed from `internal/domain/**`.
2. **Given** a prompt-bearing turn runs, **When** it completes, **Then** the answer, the tool steps, the usage records, the exit codes, and the incomplete-turn tool-error path are **unchanged** (godog E2E + unit pins green).
3. **Given** the extracted contracts, **When** declared in `internal/domain/**`, **Then** they reference only stdlib + domain types (RULE-C-pure); `internal/agent` consumes them (downward import) and no cycle is introduced.
4. **Given** the baseline, **When** the gate runs, **Then** it is **unchanged** (still `internal/cli -> internal/agent` + `internal/cli -> internal/ui`), reporting 0 new / 0 stale — no accidental movement.

**Functional Requirements**:

- **FR-001**: The loop's crossing **contracts** — the turn-result type (`Answer`/`Steps`/`Usage`/`Calls`), the incomplete-turn error type, and the wire-def projection (`ToolDefs(reg) []llm.ToolDef`) — MUST be declared in `internal/domain/**` and consumed by **both** `internal/agent` and `internal/cli` from there; the CLI's production `internal/agent` identifier references MUST reduce to **`AgentLoop` only**.
- **FR-002**: The extracted contracts MUST reference only stdlib + domain types (RULE-C purity — no `agent` type crosses the boundary; the projection takes a domain `tools.Registry` and returns domain `llm.ToolDef`).
- **FR-003**: The implementation MUST be **behaviour-preserving**: the result field semantics, the incomplete-turn error classification (`errors.As`), and the `ToolDefs`-based pre-flight estimate MUST be unchanged (byte-identical). Existing tests MUST pass **unmodified** except where a test directly names a moved symbol.
- **FR-004**: `internal/agent` (`AgentLoop`, `Run`) MUST consume the **domain** contracts (downward import; RULE-A-clean) and MUST NOT retain a duplicate declaration of any extracted contract.

**Non-Functional Requirements**:

- **NFR-001**: `internal/cli` MUST NOT gain any new non-sanctioned import; the round introduces **no** new `internal/cli → internal/agent` reference beyond the surviving `AgentLoop` construction, and **no** new domain-package cycle.

---

### User Story 2 - the surviving `cli → agent` coupling is proven edge-sized for the next slice (Priority: P1)

As a maintainer, I want the round to leave exactly **one** `→ agent` call site (the `AgentLoop` construction) and a **green, unchanged** baseline, so the next slice (the port inversion) is demonstrably edge-sized and cannot hide a second coupling.

**Why this priority**: the DoD is the gate (#92 AC5); this round's measurable outcome is *reduced coupling*, not a baseline move.

**Independent verification**: count the `internal/cli` production references to `internal/agent` — exactly one (`agent.AgentLoop`); run the gate — green, baseline unchanged; re-read `tools/arch/baseline.txt` — still the 2 lines.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** `tools/arch/baseline.txt` is **unchanged** and the gate reports **0 new / 0 stale**.
2. **Given** the surviving coupling, **When** enumerated, **Then** it is exactly one construction call site (`&agent.AgentLoop{…}`) — the sizing evidence for sub-slice 2.
3. **Given** a change that leaves a *second* `→ agent` coupling (a partial extraction), **When** reviewed, **Then** it is **rejected** — sub-slice 2 must see a single, edge-sized call site.

**Functional Requirements**:

- **FR-005**: `internal/cli` production MUST reference **exactly one** `internal/agent` identifier (`agent.AgentLoop`); any other `agent.*` reference is a **scope failure**.
- **FR-006**: The round MUST NOT add/modify/stale a baseline entry, weaken RULE-E, or exempt the surviving edge; `tools/arch/baseline.txt` is **byte-identical** at delivery.

**Non-Functional Requirements**:

- **NFR-002**: The extraction + truth/ADR records MUST land as **one atomic delivery**; because the baseline does not move this round there is no "stricter-baseline-before-compliant-code" window, but the round MUST still verify the baseline is untouched.

---

### User Story 3 - the re-cut (sub-slice 1) is recorded in truth and an ADR (Priority: P2)

As a maintainer/operator, I want the re-cut decision, the extracted domain contracts, and the *deferred* inversion recorded in `specs/truth/techstack.md` (Build & Tooling) and a new ADR, so sub-slice 2 can cite the pattern, and the project's layered-architecture statement stays current.

**Why this priority**: it bounds the round in truth/governance and makes the two-step re-cut explicit (so no future reader expects a baseline move here).

**Independent verification**: read the new ADR + its index row; read the Layer-discipline gate row — it records the re-cut and states that the baseline figure is **unchanged (2)** pending sub-slice 2.

**Acceptance Scenarios**:

1. **Given** the round lands, **Then** a new **ADR** records the re-cut: the extracted domain contracts, the alias-vs-reference choice (Q2), the **deferred** construction inversion, the **baseline unchanged at 2** statement, and its relation to ADR 0011/0016/0017.
2. **Given** the technology-stack truth, **Then** the Layer-discipline gate row records the re-cut + the **unchanged (2)** figure, and any CLI-architecture row naming the loop contracts reflects the domain-owned reality; the `verify` aggregate member is unchanged.
3. **Given** the remaining slices, **Then** they stay recorded on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).

**Functional Requirements**:

- **FR-007**: A new **ADR** (next free number, confirmed in research; `docs/decisions/NNNN-<slug>.md` + the `docs/decisions/README.md` index row) MUST record the round: the re-cut rationale, the extracted contracts, the **deferred** construction inversion (sub-slice 2, baseline **2 → 1**), what the change is **not**, and its relation to ADR 0011/0016/0017.
- **FR-008**: `specs/truth/techstack.md` (**Build & Tooling**) MUST be **MODIFY**d: the Layer-discipline gate row records the re-cut and the **unchanged (2)** figure; any architectural row naming the loop contracts reflects the domain-owned reality; the `verify` aggregate member is unchanged.
- **FR-009**: The round MUST leave the **`→ agent` construction inversion** and the **`→ ui`** edge and F-6/F-7/F-8 **out of scope** (later slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)), and MUST state the sub-slice boundary in `research.md`/`plan.md`.

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only**; **stdlib-only** (no new module dependency; `go.mod`/`go.sum` unchanged).

### Edge Cases

- **A domain contract that would leak an `agent` type** → forbidden (FR-002); `AgentResult`/`ErrIncomplete` are re-expressed as domain types; `ToolDefs` returns `[]llm.ToolDef` (domain).
- **Alias vs. two names** → if `internal/agent` keeps `type AgentResult = domainagent.TurnResult` (an **alias**) the concept keeps **one** identity; a *distinct* type would add a second name and a conversion — **the Q2 / RD decision** (one name per concept).
- **The `errors.As` type-assertion** → moving the incomplete-turn error to the domain MUST keep the *same* `errors.As`-based classification and the identical `emitToolError` `stderr` bytes.
- **A partial extraction** (some contracts moved, one left) → forbidden (FR-005): it would leave a second `→ agent` coupling and defeat the edge-sizing of sub-slice 2.
- **A test file importing `internal/agent`** → none today; if a test needs a moved symbol it must import the **domain** package, not re-introduce the edge.
- **A cycle** introduced by the extracted contracts → the SCC pass **fails** (cycles have no baseline).
- **Accidental baseline drift** (a moved symbol changes the import graph) → the gate reports 0 new / 0 stale; any movement is a **scope failure** (FR-006).
- **OS/build-tag-gated files** → the `CROSS_TARGETS` union catches an OS-gated import; custom build-tag-gated files remain a recorded residual (ADR 0011 D6).

## Requirements *(mandatory)*

> Story-specific FR/NFR are attached under each story; this section holds only cross-story constraints.

### Global Requirements

#### Functional Requirements

- **FR-010**: The round MUST preserve **zero behavioural change** (FR-003), leave the RULE-E baseline **byte-identical**, and ship the extraction + truth/ADR records as **one PR**; it MUST NOT touch the `AgentLoop` construction, the `→ ui` edge, any flag/exit-code/stream contract, or any domain business logic.
- **FR-011**: The round MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010): (a) leave one of the extracted contracts referenced from `internal/agent` in `internal/cli` ⇒ the identifier count is **2+**, not 1 (the extraction is incomplete); (b) flip a domain contract's field so the CLI's read breaks ⇒ a compile failure at the seam; (c) a partial extraction that leaves a second `→ agent` identifier ⇒ rejected by the identifier-count check.
- **FR-012**: The round MUST NOT modify a delivered `specs/plans/NNN-*` package and MUST record the deferred inversion + F-6/F-7/F-8 on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).

#### Non-Functional Requirements

- **NFR-004**: The witness MUST be **the gate + the identifier-count check + unit seams**, NOT the E2E suite (#92 AC5); `make verify` (incl. `verify-architecture`, cross-compile 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL row.
- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.

### Key Entities

- **The extracted contracts** — the turn-result type, the incomplete-turn error type, and the `ToolDefs` projection, declared in `internal/domain/**`.
- **The surviving coupling** — the single `agent.AgentLoop` construction (9 fields) in `internal/cli/cli.go`.
- **The baseline** — `tools/arch/baseline.txt`; **unchanged** (2 lines) this round; the ratchet moves in the deferred sub-slice 2.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `internal/cli` production references exactly **one** `internal/agent` identifier (`agent.AgentLoop`); the turn-result/error/`ToolDefs` are imported from `internal/domain/**`; RULE-A/B/C stay **0**; cycles **0**. (covers FR-001, FR-005)
- **SC-002**: `make verify` is **green** at delivery with `tools/arch/baseline.txt` **byte-identical** (0 new / 0 stale); `go test -count=1 ./...` green (incl. the godog E2E); the topology audit unchanged. (covers FR-003, FR-006, NFR-004)
- **SC-003**: The falsifiability witnesses are reproduced then reverted — an incomplete extraction shows **2+** identifiers (not 1); a flipped domain contract breaks the CLI seam at compile time. (covers FR-010, FR-011)
- **SC-004**: The extracted contracts live in **`internal/domain/**`** referencing only stdlib + domain types (RULE-C preserved; no `agent` type crosses in); `internal/agent` consumes them downward; no cycle. (covers FR-002, FR-004)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the re-cut + the **unchanged (2)** figure (citing the new ADR); the ADR + its index row are present; `go.mod`/`go.sum` unchanged. (covers FR-007, FR-008, NFR-003, NFR-005)
- **SC-006**: The deferred inversion + the `→ ui` edge + F-6/F-7/F-8 are recorded on [#101](https://github.com/gosharplite/tellme/issues/101); no frozen plan package is touched. (covers FR-009, FR-012)

## Assumptions

- **A1 (programme slice)** — Q1 → **B**: this round is **re-cut sub-slice 1** (contracts → domain); the construction inversion (baseline 2 → 1) is **sub-slice 2**.
- **A2 (no E2E change)** — no `tellme` CLI behaviour changes; `/axb-spec-by-example` is **NOOP** — precedent: rounds 020/031/036/041–048.
- **A3 (contract home/naming is RD)** — the exact domain package/type names and the **alias vs. reference** choice are **RD** decisions (`research.md` D-x), subject to FR-002 / RULE-A·C (natural home: `internal/domain/agent`, peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`).
- **A4 (ADR required)** — a new ADR records the re-cut (FR-007); the number is confirmed in research (next free after `0017`).
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**; `/axb-ui-plan` **skipped**.
- **A6 (`/axb-dsl-refine` NOOP)** — a `cli`-reader refactor is not a CLI-contract change; no `DSLRow` change.
- **A7 (scope guard)** — the round touches `internal/domain/agent/**` (the extracted contracts), `internal/agent/**` (consumes them), `internal/cli/**` (reads them), `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`. **No** baseline edit.

## Out of scope (recorded)

- **Sub-slice 2** — the `AgentLoop` **construction/execution inversion** into an injected domain port (the round that drops the `→ agent` edge; baseline **2 → 1**).
- The other residual edge — `internal/cli → internal/ui` (a later slice; itself ≥ edge-sized per ADR 0017 §Forward).
- **F-6/F-7/F-8** (later slices).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting the loop's business logic; ADR 0014/0015 patterns are reused, not re-litigated.
- A **re-ruling** of RULE-E's sanctioned set.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#101](https://github.com/gosharplite/tellme/issues/101) — the round's anchor (**R5**; the programme + the 2 residual edges + F-6/F-7/F-8).
- [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause) + **AC5** (witness = gate + unit seams).
- Round **047** (ADR **0016**) — **R5.1**: the RULE-E gate + baseline.
- Round **048** (ADR **0017**) — **R5.2**; §Forward classifies the `→ agent` edge as *"the deepest slice"* and prescribes the **value-types-first re-cut** recipe.
- Round **046** (ADR **0015**) — the loop-presentation port precedent (a port declared in `internal/domain/**`).

## References

- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt).
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) · [`0016-application-import-ceiling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0016-application-import-ceiling.md) · [`0015-loop-presentation-port.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0015-loop-presentation-port.md) · [`0017-cli-tui-prompt-decoupling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0017-cli-tui-prompt-decoupling.md) (§Forward).
- [`internal/agent/agentloop.go`](https://github.com/gosharplite/tellme/blob/dev/internal/agent/agentloop.go) (`AgentLoop`, `Run`, `AgentResult`, `ErrIncomplete`, `ToolDefs`) · [`internal/cli/cli.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/cli.go) · [`internal/cli/call_renderer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/call_renderer.go).
- [`internal/domain/agent/`](https://github.com/gosharplite/tellme/tree/dev/internal/domain/agent) — the existing port family.
- [`specs/truth/techstack.md`](https://github.com/gosharplite/tellme/blob/dev/specs/truth/techstack.md) — Build & Tooling (the gate row).
