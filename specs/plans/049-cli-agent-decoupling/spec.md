# Feature Specification: de-couple `internal/cli` from the turn loop — R5.3, the `cli → agent` slice (round 049)

**Feature Branch**: `049-cli-agent-decoupling`

**Created**: 2026-09-18

**Status**: Draft — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)). This round is the **next de-coupling slice** (round 047 = **R5.1**, the **RULE-E** gate + baseline; round 048 = **R5.2**, the `cli → ui/tui/prompt` edge). **Clarify round 1 is PENDING — Q1 (the slice + its sizing) is the round's first question** (see §Clarify pending).

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5** of the [#92](https://github.com/gosharplite/tellme/issues/92) gate-first split. After round 048 the layer-discipline guard (`tools/arch`, ADR 0011/0016) enforces **RULE-E — the application import ceiling** for the application tiers (`internal/app/**`, `internal/cli`) and the committed baseline records the **2** residual unsanctioned edges (measured 2026-09-18 @ `dev` `12964d6`):

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
```

R5's remaining work removes these edges one de-coupling slice at a time (the baseline ratchets **2 → 0**).

**Behaviour intent**: **MODIFY (behaviour-preserving refactor)** — invert the CLI's use of the turn loop (`internal/agent`) into an **injected domain port**, so `internal/cli` no longer imports `internal/agent`; the RULE-E baseline loses exactly that edge. **No** user-facing behaviour change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary) and no domain business logic — the round-044/045/046/048 lineage (gate-proven, behaviour-preserving structural refactors).

---

## ⚠️ Clarify pending (round 1)

> **Q1 — slice selection + sizing (PENDING).** The operator's initial direction was the **`internal/cli → internal/agent`** edge (the **turn-path loop construction**). Grounding measurement (2026-09-18 @ `dev` `12964d6`) corrects an earlier shorthand: **ADR 0017 §Forward classifies this edge as *"the deepest slice"***, while the **`internal/cli → internal/ui`** edge is described there as *"the next natural slice"* (though itself **not edge-sized** — ~20 call sites · 3 crossing value types · 4 stateful objects · 8 formatters → *"likely not a single-edge-sized round; consider re-cutting it"*). So both remaining edges exceed the round-048 (R5.2) sizing, and the `→ agent` edge is the deeper of the two. Q1 therefore decides:
>
> - **(A)** do the **full `cli → agent` inversion** in **one** round (baseline **2 → 1**) — the deepest slice, largest review surface; or
> - **(B)** **re-cut** `cli → agent` into ordered sub-slices (e.g. first move the **loop result/error + tool-def projection** to `internal/domain/**`, then invert the **loop construction/runner** in a later round); or
> - **(C)** switch the round's slice to a **smaller ride-along** from [#92](https://github.com/gosharplite/tellme/issues/92) / [#103](https://github.com/gosharplite/tellme/issues/103) and leave both >edge-sized de-couplings for a dedicated (possibly multi-round) programme.
>
> **Q2 and Q3 are held** (asked one at a time, only after Q1 lands): **Q2** = the port home/shape (the existing `internal/domain/agent` port family vs. a new domain package); **Q3** = the ride-along choice (the ADR 0017 §Forward `Validate()` interface-seam note; the F-6/F-7/F-8 items).

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `12964d6` (static import scan; the gate re-measures at implementation).

### The selected residual edge and its call sites (`internal/cli` production files)

| Item | Detail |
| --- | --- |
| Edge | `internal/cli -> internal/agent` |
| Import sites | `internal/cli/cli.go`, `internal/cli/call_renderer.go` (production); **no** `internal/cli/*_test.go` imports `internal/agent` |
| Identifiers used (4) | `agent.AgentLoop` (the **constructed** loop type) · `agent.AgentResult` (the returned turn result type) · `agent.ErrIncomplete` (the incomplete-turn error type, type-asserted) · `agent.ToolDefs` (the wire-def projection function) |
| Call sites | `&agent.AgentLoop{…}` (`cli.go:699`, the turn-path loop construction — 9 fields: `Gateway`/`Registry`/`MaxLoops`/`EffectiveBudget`/`Stderr`/`Now`/`ToolUsage`/`Lines`/`Observer`) · `loop.Run(ctx, prompt, prior)` → `(agent.AgentResult, error)` (`cli.go:~749`) · `var inc *agent.ErrIncomplete; errors.As(err, &inc)` (`cli.go:754`) · `agent.ToolDefs(r.reg)` (`call_renderer.go:77`, the CLI-computed pre-flight estimate) · `result agent.AgentResult` (`call_renderer.go:175`, `persistTurnUsage` parameter) |
| The loop contract the port must carry | `func (a *AgentLoop) Run(ctx context.Context, prompt string, prior []history.Entry) (AgentResult, error)` (`agentloop.go:112`); `type AgentResult struct { Answer string; Steps []history.Step; Usage llm.Usage; Calls []llm.Usage }` (`agentloop.go:52`) |
| Existing domain ports the loop already consumes | `agentport.LoopObserver` (`internal/domain/agent/observer.go`) · `agentport.CallObserver` (`internal/domain/agent/call.go`) · `agentport.ToolLineRenderer` (`internal/domain/agent/presenter.go`) — the loop already **imports downward** into `internal/domain/**`, so the inversion has a natural, established seam |

### Tier constraint that shapes the port

ADR 0011 **D1** ranks `domain` 0 · `config`/`home` 1 · `app` 2 · `infrastructure` 3 · **`agent` 4** · `ui`/`ui/tui/**` 5 · `cli` 6 (higher = more upward). **RULE-A** forbids an *upward* import, so the adapter that constructs/satisfies the loop port must live at a tier **≥ 4** (or the exempt composition root `cmd/tellme`, tier-table-exempt). The **port declaration** lands in `internal/domain/**` (Q2 pending); `internal/cli` then imports only the domain port. `internal/domain/**` must stay **RULE-C-pure** (stdlib + domain types only — **no** `agent` type crosses in).

### The gate that measures the DoD (round 047 / ADR 0016)

- **RULE-E** (ADR 0016): a governed application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`; the sanctioned set is a **normative** section of the gate's tier table (ADR 0011 **D7**), **default-deny**, **fail-on-stale**.
- The committed **baseline** (`tools/arch/baseline.txt`) is a **fail-on-stale ratchet** (ADR 0011 **D3**): a new violation fails; a stale entry fails; removing an entry while its violation persists fails.
- **DoD is the gate**, not the E2E suite (#92 **AC5**, the round-009 false-confidence trap).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale**; 0 cycles; cross-compile 4/4.
- **Zero behavioural change**: the turn path (positional / piped / `-i` submit / `--retry`) produces identical `stdout`/`stderr` byte-contracts, identical exit codes, unchanged flags, and **no** new Gherkin/DSL row.
- **`internal/domain/**` stays 100 % pure**: the new port references only stdlib + domain types; no `agent` type crosses the boundary (RULE-C).
- **The loop's own inward direction is preserved**: `internal/agent` keeps importing `internal/domain/**` (its existing `agentport.*` ports) and never the reverse.
- Tests stay hermetic (no mutable package globals; round-029/030 precedent).

---

## User Scenarios & Testing *(mandatory)*

> **Scope note**: the story split below is written for the operator's stated slice (Q1 option **A** — the full `cli → agent` inversion in one round). If Q1 lands on **(B)** a re-cut or **(C)** a different slice, US1/US2 are re-scoped to the chosen sub-slice and the baseline movement changes accordingly.

### User Story 1 - the turn loop is reached through an injected domain port, so `internal/cli` no longer imports `internal/agent` (Priority: P1)

As a maintainer, I want the turn loop (`AgentLoop` construction + `Run`) to be invoked through an **injected domain port** wired at the composition root, so `internal/cli` no longer imports `internal/agent` — while the turn's behaviour (answer, tool steps, usage, error classification) is **byte-identical**.

**Why this priority**: it is the round's entire reason to exist — a de-coupling step of R5.

**Independent verification**: after the round, `go list`/`grep` shows **no** `internal/cli` (production or test) import of `internal/agent`; `make verify-architecture` reports the RULE-E baseline with that edge gone; the godog E2E + the turn unit pins stay green; a deliberately re-introduced import reds the gate.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** `make verify` runs, **Then** the gate is **green** and `baseline.txt` no longer lists `internal/cli -> internal/agent`.
2. **Given** a prompt-bearing turn runs (positional / piped / `-i` / `--retry`), **When** the turn completes, **Then** the answer, the tool steps, the usage records, the exit codes, and the `ErrIncomplete` tool-error path are **unchanged** (godog E2E + unit pins green).
3. **Given** a change that re-introduces `internal/cli -> internal/agent`, **When** `make verify` runs, **Then** RULE-E **fails** and names the edge.
4. **Given** the new domain port, **When** it is declared, **Then** it references only stdlib + domain types (RULE-C-pure); the adapter satisfying it lives at a tier ≥ 4 (or the exempt `cmd/tellme`); the port is wired at the composition root.

**Functional Requirements**:

- **FR-001**: The **`internal/cli → internal/agent`** import MUST be removed from the production **and test** import graph of `internal/cli`; the turn loop MUST be reached through an **injected domain port** declared in `internal/domain/**` and wired at the composition root.
- **FR-002**: The port MUST be a **domain-owned contract** referencing only stdlib + domain types (RULE-C purity). It MUST carry at least: the loop **execution** (`ctx`/prompt/prior → result) and the **wire-def projection** (`ToolDefs`) and the **incomplete-turn error** signal — expressed with **domain** types (`AgentResult`-equivalent with `Answer`/`Steps`/`Usage`/`Calls`; an `ErrIncomplete`-equivalent domain error), **not** `agent` types. The exact port package/name/shape is an **RD** decision recorded in `research.md` + the ADR; the natural home is the existing `internal/domain/agent` port family (peer of `LoopObserver`/`CallObserver`/`ToolLineRenderer`).
- **FR-003**: The implementation MUST be **behaviour-preserving**: the loop's `Run` semantics, the `AgentResult` field semantics, the `ErrIncomplete` classification (the `emitToolError` path), and the `ToolDefs`-based pre-flight estimate MUST be unchanged. Existing tests MUST pass **unmodified** except where a test directly names the inverted seam.
- **FR-004**: The adapter that satisfies the port (constructing the `AgentLoop` with its 9 fields and calling `Run`) MUST live at a tier **≥ 4** (`internal/agent` itself, or a tier-≥4 package) or in the exempt composition root `cmd/tellme` — because RULE-A forbids a lower tier importing `internal/agent`.

**Non-Functional Requirements**:

- **NFR-001**: The composition root (`cmd/tellme`) MUST remain the single assembly site for the port; `internal/cli` MUST NOT construct the loop implementation itself and MUST NOT retain a nil-defaulted in-package fallback that imports `internal/agent`.

---

### User Story 2 - the RULE-E baseline shrinks 2 → 1 — a gate-proven ratchet (Priority: P1)

As a maintainer, I want the committed baseline to lose exactly the removed edge and to fail if it drifts, so this de-coupling is machine-verified and the ratchet can only shrink toward **0**.

**Why this priority**: the DoD is the gate (#92 AC5); a green E2E suite alone is false confidence.

**Independent verification**: read `tools/arch/baseline.txt` — it lists the 2 residuals **minus** `internal/cli -> internal/agent`; run the gate — green; re-add the import — red; leave the now-fixed edge in the baseline — red (stale).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** `baseline.txt` is **regenerated from the gate** and lists exactly `internal/cli -> internal/ui`.
2. **Given** the removed edge is re-introduced, **When** the gate runs, **Then** it **fails** (RULE-E, not baselined).
3. **Given** a stale entry (the `→ agent` edge left in the baseline), **When** the gate runs, **Then** it **fails** and names the entry.
4. **Given** a governance change to the port that would re-open the edge, **When** the gate runs, **Then** it **fails** — no release valve (ADR 0011/0016).

**Functional Requirements**:

- **FR-005**: `tools/arch/baseline.txt` MUST be regenerated from the gate (`make verify-architecture-update`) after the de-coupling, dropping exactly `internal/cli -> internal/agent`; no RULE-A/B/C entry appears; the gate is **green on `dev` at delivery**.
- **FR-006**: The round MUST NOT add any baseline entry, weaken RULE-E, or exempt the removed edge; the sanctioned set (ADR 0016) is unchanged.

**Non-Functional Requirements**:

- **NFR-002**: The de-coupling + the baseline regeneration MUST land as **one atomic delivery** (the round-040 TD-1 precedent: a stricter baseline must never land before the code is compliant).

---

### User Story 3 - the de-coupling and the port are recorded in truth and an ADR (Priority: P2)

As a maintainer/operator, I want the new domain port, its home, and the shrinking baseline recorded in `specs/truth/techstack.md` (Build & Tooling) and a new ADR, so the remaining R5 slices can cite the pattern and the project's layered-architecture statement stays current.

**Why this priority**: it bounds the round in truth/governance and makes the pattern reusable.

**Independent verification**: read the new ADR + its index row; read the Layer-discipline gate row — it records the new baseline figure (**1**) and cites the ADR; the guard self-test still passes.

**Acceptance Scenarios**:

1. **Given** the round lands, **Then** a new **ADR** records the domain port, its home, and the ratchet movement (2 → 1), citing ADR **0011/0016/0013/0015/0017**.
2. **Given** the technology-stack truth, **Then** the Layer-discipline gate row records the baseline figure **1** and cites the ADR; any architectural row naming the inverted seam (the CLI turn path / loop construction) reflects the injected-domain-port reality; the `verify` aggregate member is unchanged.
3. **Given** the remaining residual edge, **Then** it stays recorded on the live issue [#101](https://github.com/gosharplite/tellme/issues/101) (not a frozen package).

**Functional Requirements**:

- **FR-007**: A new **ADR** (next free number, confirmed in research; `docs/decisions/NNNN-<slug>.md` + the `docs/decisions/README.md` index row) MUST record the de-coupling: the **domain** port, its home, the behaviour-preservation claim, the baseline movement **2 → 1**, and its relation to ADR 0011/0016/0013/0015/0017 — including **what the change is *not*** (no user-facing change; no new normative source; the `→ ui` edge untouched).
- **FR-008**: `specs/truth/techstack.md` (**Build & Tooling**) MUST be **MODIFY**d: the Layer-discipline gate row records the new RULE-E baseline figure (**1**) and cites the ADR; any architectural row naming the turn-path loop construction reflects the injected-domain-port reality; the `verify` aggregate member is unchanged.
- **FR-009**: The round MUST leave the de-coupling of the **other** residual edge (`→ ui`) and the F-6/F-7/F-8 items **out of scope** (later slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)), and MUST state the round's slice boundary in `research.md`/`plan.md`.

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only**; **stdlib-only** for the port (no new module dependency; `go.mod`/`go.sum` unchanged).

### Edge Cases

- **A domain port that would leak an `agent` type** → forbidden: the port MUST reference only stdlib + domain types (RULE-C); `AgentResult`/`ErrIncomplete` semantics are re-expressed as domain types, else the port home would have to change.
- **The adapter that constructs/satisfies the loop port** → it MUST live at a tier ≥ 4 (`internal/agent` or a tier-≥4 package) or in the exempt `cmd/tellme`, because RULE-A forbids a lower tier importing `internal/agent`.
- **A test file (not production) importing `internal/agent`** → the merged production+test graph governs; the test import must be inverted/moved too, else RULE-E stays red.
- **A partially de-coupled round** (some call sites inverted, some not) → forbidden: `internal/cli` either imports the package (edge present → baseline unchanged) or it does not (edge gone).
- **The `ErrIncomplete` classification** → the `emitToolError` path MUST survive the inversion with an identical `stderr` contract (the tool-error line bytes unchanged).
- **The `ToolDefs`-based pre-flight estimate** → the estimate MUST stay byte-identical (the round-011 RF-1 "counts exactly what the loop sends" invariant).
- **The removed edge re-appearing via a new call site in a later round** → **fails** RULE-E (not baselined) — the ratchet has teeth.
- **A stale baseline entry** after the de-coupling → **fails** (ADR 0011 D3 preserved).
- **The composition root not wiring the port** → a compile/harness failure (a missing injection is caught by the build + the CLI smoke/`Validate()` seam), not a silent behavioural drift; per ADR 0017 §Forward, an **interface-typed** seam added to `Options`/`Dependencies` MUST get its own `Validate()` assertion (`Kind()==reflect.Func` is blind to interface seams).
- **A cycle introduced by the new domain package** → the SCC pass **fails** (cycles have no baseline).
- **OS/build-tag-gated files** → the `CROSS_TARGETS` union catches an OS-gated import; custom build-tag-gated files remain a recorded residual (ADR 0011 D6).

## Requirements *(mandatory)*

> Story-specific FR/NFR are attached under each story; this section holds only cross-story constraints.

### Global Requirements

#### Functional Requirements

- **FR-010**: The round MUST preserve **zero behavioural change** (FR-003) and MUST ship the de-coupling + baseline regeneration + truth/ADR records as **one PR** (NFR-002); it MUST NOT touch the `→ ui` residual edge, any flag/exit-code/stream contract, or any domain business logic.
- **FR-011**: The round MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010): (a) re-introduce `internal/cli → internal/agent` ⇒ RULE-E reds; (b) leave the now-fixed edge in the baseline ⇒ the gate reds (stale); (c) a missing/false port injection ⇒ the CLI `Validate()`/smoke seam (or the build) fails loudly.
- **FR-012**: The round MUST NOT modify a delivered `specs/plans/NNN-*` package and MUST record the remaining slice + F-6/F-7/F-8 on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).

#### Non-Functional Requirements

- **NFR-004**: The witness MUST be **the gate + unit seams**, NOT the E2E suite (#92 AC5); `make verify` (incl. `verify-architecture`, cross-compile 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL row.
- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.

### Key Entities

- **The removed edge** — `internal/cli -> internal/agent` (production; no test import).
- **The injected domain port** — the `internal/domain/**` seam carrying the loop execution + `ToolDefs` projection + incomplete-turn error, expressed with domain types (Q2 pending).
- **The adapter** — the tier-≥4 value (or exempt `cmd/tellme` wiring) that constructs the `AgentLoop` and satisfies the port.
- **The loop** — `internal/agent.AgentLoop` (`Run(ctx, prompt, prior) (AgentResult, error)`), remaining the single owner of the turn cycle.
- **The baseline** — `tools/arch/baseline.txt`; loses exactly the `→ agent` edge (**2 → 1**), a fail-on-stale ratchet (ADR 0011 D3).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go list`/`grep` shows **no** `internal/cli` (production + test) import of `internal/agent`; the RULE-E baseline lists exactly the one remaining residual (`internal/cli -> internal/ui`); RULE-A/B/C stay **0**; cycles **0**. (covers FR-001, FR-005)
- **SC-002**: `make verify` is **green** at delivery; `go test -count=1 ./...` green (incl. the godog E2E); the Gherkin/DSL topology audit unchanged. (covers FR-003, NFR-004)
- **SC-003**: A deliberately re-introduced `internal/cli → internal/agent` import makes RULE-E **fail** and names the edge; a stale baseline entry makes the gate **fail** — both reproduced as falsifiability witnesses, then reverted. (covers FR-006, FR-011)
- **SC-004**: The port lives in **`internal/domain/**`** referencing only stdlib + domain types (RULE-C preserved; no `agent` type crosses in); the adapter stays at a tier ≥ 4; a missing injection fails loudly; any interface-typed `Options`/`Dependencies` seam gets its own `Validate()` assertion. (covers FR-002, FR-004, FR-011)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the baseline figure **1** (citing the new ADR); the ADR + its index row are present; **no production behaviour** changes; `go.mod`/`go.sum` unchanged. (covers FR-007, FR-008, NFR-003, NFR-005)
- **SC-006**: The remaining residual edge + F-6/F-7/F-8 are recorded on [#101](https://github.com/gosharplite/tellme/issues/101); no frozen plan package is touched. (covers FR-009, FR-012)

## Assumptions

- **A1 (programme slice)** — this round is **R5.3**; the slice is **`internal/cli → internal/agent`** (operator direction), **subject to Q1** (full inversion vs. re-cut vs. a smaller ride-along). The other edge is a later slice.
- **A2 (no E2E change)** — the refactor changes no `tellme` CLI behaviour, so the E2E suite is a **regression** carrier, not the witness; the witness is **the gate + unit seams** (NFR-004). `/axb-spec-by-example` is therefore **NOOP** — precedent: rounds 020/031/036/041–048.
- **A3 (port design is RD)** — the port's exact package/name/shape are **RD** decisions (`research.md` D-x), subject to FR-002 / RULE-A·C.
- **A4 (ADR required)** — a new ADR records the de-coupling (FR-007); the number is confirmed in research (next free after `0017`).
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP** (no persisted/runtime-state change); `/axb-ui-plan` **skipped** (no new user-facing surface).
- **A6 (`/axb-dsl-refine` NOOP)** — a `cli`-reader refactor is not a CLI-contract change: no feature Rule, Example, step, or `DSLRow` change; the topology audit stays unchanged.
- **A7 (scope guard)** — the round touches `internal/cli/**`, the new `internal/domain/**` port surface, `internal/agent/**` (the adapter, if any), `cmd/tellme` (wiring), `tools/arch/baseline.txt`, `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`.

## Out of scope (recorded)

- The other residual edge — `internal/cli → internal/ui` (later R5 slice, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)).
- **F-6/F-7/F-8** (the remaining PR #104 deferrals: narrow seams ≤2 fields; named `Discovery{Closer}`; `OutputSink` → interface) — later slices.
- Re-opening R2's decisions (composition-root home `cmd/tellme`, `internal/app/deps.Dependencies`, `agentTools()` relocation, MCP orchestration into `internal/infrastructure/mcp`).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting the loop's business logic; the ADR 0014 (yield policy) and ADR 0015 (presentation port) patterns are reused, not re-litigated.
- A **re-ruling** of RULE-E's sanctioned set.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#101](https://github.com/gosharplite/tellme/issues/101) — the round's anchor (**R5**; the programme + the 2 residual edges + F-6/F-7/F-8).
- [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause) + **AC5** (witness = gate + unit seams).
- Round **047** ([#101](https://github.com/gosharplite/tellme/issues/101), ADR **0016**) — delivered **R5.1**: the RULE-E gate + the baseline this round shrinks.
- Round **048** ([#101](https://github.com/gosharplite/tellme/issues/101), ADR **0017**) — delivered **R5.2** (the `cli → ui/tui/prompt` edge) and recorded the §Forward classification (the `→ agent` edge = *"the deepest slice"*).
- Round **046** ([#108](https://github.com/gosharplite/tellme/issues/108), ADR **0015**) — the loop-presentation port precedent.
- Round **044** ([#100](https://github.com/gosharplite/tellme/issues/100), ADR **0013**) — the injected-`Dependencies` pattern + the composition root.

## References

- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the guard, tier table, and baseline.
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) · [`0016-application-import-ceiling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0016-application-import-ceiling.md) · [`0015-loop-presentation-port.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0015-loop-presentation-port.md) · [`0017-cli-tui-prompt-decoupling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0017-cli-tui-prompt-decoupling.md) (§Forward).
- [`internal/agent/agentloop.go`](https://github.com/gosharplite/tellme/blob/dev/internal/agent/agentloop.go) (`AgentLoop`, `Run`, `AgentResult`, `ErrIncomplete`, `ToolDefs`) · [`internal/cli/cli.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/cli.go) (the loop construction + `ErrIncomplete` handling) · [`internal/cli/call_renderer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/call_renderer.go) (`ToolDefs`, `persistTurnUsage`).
- [`internal/domain/agent/`](https://github.com/gosharplite/tellme/tree/dev/internal/domain/agent) — the existing port family (`LoopObserver`, `CallObserver`, `ToolLineRenderer`).
- [`specs/truth/techstack.md`](https://github.com/gosharplite/tellme/blob/dev/specs/truth/techstack.md) — Build & Tooling (the gate row).
