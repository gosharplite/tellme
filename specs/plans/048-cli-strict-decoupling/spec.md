# Feature Specification: `internal/cli` strict de-coupling — R5.2, the first de-coupling slice (round 048)

**Feature Branch**: `048-cli-strict-decoupling`

**Created**: 2026-09-18

**Status**: Draft — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)); this round is **R5.2** (the round-047 **R5.1** slice delivered the **RULE-E** gate + its 3-entry baseline). **Clarify round 1 is OPEN** — **Q1 (slice selection)** is the round's pivotal decision and gate on the package slug/scope; per `/axb-specify` Phase 2 no answer is assumed until the operator locks it (see §Open decision). This draft is slice-parameterised: every story/FR is written for **the selected residual edge(s)**, not for a pre-chosen one.

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5** of the [#92](https://github.com/gosharplite/tellme/issues/92) gate-first split. After round 047 the layer-discipline guard (`tools/arch`, ADR 0011) enforces **RULE-E — the application import ceiling** for the application tiers (`internal/app/**`, `internal/cli`), and the committed baseline (`tools/arch/baseline.txt`) records the **3** residual unsanctioned edges:

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
internal/cli -> internal/ui/tui/prompt
```

R5's remaining work is to **remove these residual edges** (one de-coupling slice at a time), so the baseline ratchets **3 → 0** and [#92](https://github.com/gosharplite/tellme/issues/92) **AC2**'s second clause holds in fact.

**Behaviour intent**: **MODIFY (behaviour-preserving refactor)** — invert the selected residual edge(s) into injected **ports** so `internal/cli` no longer imports the unsanctioned internal package; the RULE-E baseline loses exactly the removed edge(s). It changes **no** user-facing behaviour (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary) and no domain business logic — the round-044/045/046 lineage (gate-proven, behaviour-preserving structural refactors).

---

## Open decision (clarify round 1 — Q1)

> The round's scope is **not** assumed. #101 states the full de-coupling "is a multi-round programme, not a slice"; the operator selects which residual edge(s) this round inverts. **Q1 (slice selection)** is asked first; the answer fixes the slug, the user-story split, and the acceptance surface. Subsequent questions (port home per edge; whether a companion **F-4/F-6/F-7/F-8** item rides along) are asked **one at a time** after Q1.

**Q1 options (recommended: B).**

| # | Slice | RULE-E baseline | Why / risk |
| --- | --- | --- | --- |
| **A** | All **3** edges in one round (`agent` + `ui` + `ui/tui/prompt`) | 3 → 0 | Closes R5 in one shot, but inverts essentially **every** `cli` wiring call across the turn path, the renderer/observer stack, and the TUI — the largest behavioural surface; #101 itself calls this "not a slice". |
| **B** | **One** edge — the **TUI prompt** (`internal/cli → internal/ui/tui/prompt`) | 3 → 2 | Smallest, most self-contained: a single subsystem behind the existing `tuiPromptRunner` seam (`cli.go`); one injected port; low behavioural risk; a clean falsifiable DoD. **Recommended.** |
| **C** | **One** edge — the **agent loop** (`internal/cli → internal/agent`) | 3 → 2 | The turn-path loop construction (`agent.AgentLoop`, `agent.AgentResult`, `agent.ToolDefs`, `agent.ErrIncomplete`) — a deep, wide seam. |
| **D** | **One** edge — the **ui presentation** (`internal/cli → internal/ui`) | 3 → 2 | The renderer/spinner/coordinator/formatters (`ui.NewRenderer`, `ui.NewSpinner`, `ui.Format*`, `ui.Pricing`, …) — broad but mechanical; a large surface of small call sites. |
| **E** | **Two** edges as one coherent slice (e.g. `ui` + `ui/tui/prompt`, the whole presentation tier) | 3 → 1 | Groups the presentation ports; a mid-size slice between B and A. |

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `e5db873` (static import scan; the gate re-measures at implementation).

### The 3 residual edges and their call sites (`internal/cli` production files)

| Edge | Import sites | Representative call sites |
| --- | --- | --- |
| `internal/cli → internal/agent` | `cli.go`, `call_renderer.go` | `agent.AgentLoop{…}` (`cli.go:681`), `agent.ErrIncomplete` (`cli.go:736`), `agent.AgentResult` + `agent.ToolDefs` (`call_renderer.go`) |
| `internal/cli → internal/ui` | `cli.go`, `call_renderer.go` | `ui.NewRenderer`, `ui.NewSpinner`, `ui.NewToolOutputCoordinator`, `ui.ToolLineRenderer`, `ui.FormatTurnOpening/Gap/PayloadStatus/ToolReason/Metrics/Ready/InputCaptured/ToolUsage`, `ui.Pricing`, `ui.UsageCounts`, `ui.HitRate`, `ui.ComputeCost`, `ui.ToolUsageRow`, `ui.DefaultToolOutputIdleGap` |
| `internal/cli → internal/ui/tui/prompt` | `cli.go` | `tuiprompt.Run` (`cli.go:202`, behind the `tuiPromptRunner` seam `defaultRunTUIPrompt`), `tuiprompt.DefaultDebounceDuration` (`cli.go:214`) |

### The gate that measures the DoD (round 047 / ADR 0016)

- **RULE-E** (ADR 0016): a governed application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`; the sanctioned set is a **normative** section of the gate's tier table (ADR 0011 **D7**), **default-deny**, **fail-on-stale**.
- The committed **baseline** (`tools/arch/baseline.txt`) is a **fail-on-stale ratchet** (ADR 0011 **D3**): a new violation fails; a stale entry fails; removing an entry while its violation persists fails.
- **DoD is the gate**, not the E2E suite (#92 **AC5**, the round-009 false-confidence trap).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale**; 0 cycles; the cross-compile gate 4/4.
- **Zero behavioural change**: `stdout` byte-exact, `stderr` line contracts unchanged, flags/exit codes unchanged, no new Gherkin/DSL row (a `cli`-reading refactor is not a CLI-contract change).
- **`internal/domain/**` stays 100 % pure**; the new port(s) declare only domain/stdlib types (no `ui`/`agent` type leakage into the domain).
- **Tests stay hermetic** (no mutable package globals; the round-029/030 precedent).
- The **existing `tuiPromptRunner` seam** and the **RULE-E tier table** are extended in place — no new Makefile target, no new normative source.

---

## User Scenarios & Testing *(mandatory)*

> The user stories are written for **the selected residual edge(s)** (Q1). Wherever a story says "the selected edge(s)", read it as the operator's Q1 choice; the stories themselves are slice-independent.

### User Story 1 - the selected residual edge(s) are inverted into an injected port, so `internal/cli` no longer imports the unsanctioned package (Priority: P1)

As a maintainer, I want the `internal/cli` wiring for the selected residual (agent / ui / ui-tui-prompt) to go through an **injected port** declared in a permitted tier, so that `internal/cli` no longer imports the unsanctioned internal package and RULE-E's baseline can shrink without any behavioural change.

**Why this priority**: it is the round's entire reason to exist — #101's remaining work; without it there is no de-coupling.

**Independent verification**: after the round, `go list`/`grep` shows no `internal/cli` import of the selected package; `make verify-architecture` reports the RULE-E baseline reduced by exactly the removed edge; `make verify` + the godog E2E remain green; a deliberately re-introduced import reds the gate (falsifiability witness).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** `make verify` runs, **Then** the gate is **green** and the RULE-E baseline no longer lists `internal/cli -> <selected package>`.
2. **Given** the de-coupling is in place, **When** the `tellme` prompt/turn/TUI surfaces run, **Then** every rendered line, byte on `stdout`, exit code, and flag behaviour is **unchanged** (godog E2E green; the existing unit pins unchanged).
3. **Given** a change that re-introduces `internal/cli -> <selected package>`, **When** `make verify` runs, **Then** RULE-E **fails** and names the edge.
4. **Given** the new port, **When** it is declared, **Then** it lives in a **permitted tier** (an `internal/domain/**` port that drags no `ui`/`agent` type, or an `internal/app/**` application-tier port) and is wired at the composition root (`cmd/tellme`).

**Functional Requirements**:

- **FR-001**: The selected residual edge(s) MUST be removed from the **production** import graph of `internal/cli`: every wiring call the tier makes into the selected package MUST be routed through an **injected port** (a small interface or function value) instead of a direct import.
- **FR-002**: The new port MUST be declared in a **permitted tier** and MUST NOT drag an unsanctioned type across its boundary: an `internal/domain/**` port may reference only domain/stdlib types (RULE-C purity preserved); if the seam cannot be expressed without leaking a `ui`/`agent` type, the port MUST live in `internal/app/**` (an application-tier utility the application tiers may import). The home is an RD decision recorded in `research.md` and the round's ADR.
- **FR-003**: The implementation MUST be **behaviour-preserving**: identical `stdout`, identical `stderr` line contracts, identical exit codes and flags; the existing tests (unit + godog E2E) MUST pass **unmodified** except where a test directly names the inverted type (test-only adaptation permitted and itemised).

**Non-Functional Requirements**:

- **NFR-001**: The refactor MUST keep the composition root (`cmd/tellme`) the single assembly site for the port; `internal/cli` MUST NOT construct the implementation itself.

### User Story 2 - the RULE-E baseline shrinks by exactly the removed edge(s) — a gate-proven ratchet (Priority: P1)

As a maintainer, I want the committed baseline to lose exactly the removed edge(s) and to fail if it drifts, so the de-coupling is machine-verified and the ratchet can only shrink toward **0**.

**Why this priority**: the DoD is the gate (#92 AC5); a green E2E suite alone is false confidence.

**Independent verification**: read `tools/arch/baseline.txt` — it lists the 3 residuals **minus** the removed edge(s); run the gate — green; re-add the removed edge's import — red; leave a now-fixed edge in the baseline — red (stale).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** `baseline.txt` is **regenerated from the gate** and lists exactly the remaining residual edges (the removed ones gone).
2. **Given** the removed edge is re-introduced, **When** the gate runs, **Then** it **fails** (RULE-E, not baselined).
3. **Given** a stale entry (a now-fixed edge left in the baseline), **When** the gate runs, **Then** it **fails** and names the entry.
4. **Given** the eventual baseline at **0** (after the later slices), **When** a stale entry is added, **Then** the gate **fails** (the round-046 anti-bypass rule holds).

**Functional Requirements**:

- **FR-004**: `tools/arch/baseline.txt` MUST be regenerated from the gate (`make verify-architecture-update`) after the de-coupling, dropping exactly the removed edge(s); no RULE-A/B/C entry appears; the gate is **green on `dev` at delivery**.
- **FR-005**: The round MUST NOT add any baseline entry, weaken RULE-E, or exempt the removed edge; the sanctioned set (ADR 0016) is unchanged unless a **re-ruling** is explicitly locked in clarify (fail-on-stale allow-list preserved).

**Non-Functional Requirements**:

- **NFR-002**: The de-coupling + the baseline regeneration MUST land as **one atomic delivery** (the round-040 TD-1 precedent: a stricter baseline must never land before the code is compliant).

### User Story 3 - the de-coupling and the port are recorded in truth and an ADR (Priority: P2)

As a maintainer/operator, I want the new port, its home tier, and the shrinking baseline recorded in `specs/truth/techstack.md` (Build & Tooling) and a new ADR, so the remaining R5 slices can cite the pattern and the project's layered architecture statement stays current.

**Why this priority**: it bounds the round in truth/governance and makes the pattern reusable for the remaining slices.

**Independent verification**: read the new ADR + its index row; read the Layer-discipline gate row — it records the new baseline figure and cites the ADR; the guard self-test still passes.

**Acceptance Scenarios**:

1. **Given** the round lands, **Then** a new **ADR** records the port, its home tier, and the ratchet movement (3 → N), citing ADR **0011/0016/0013**.
2. **Given** the technology-stack truth, **Then** the Layer-discipline gate row records the new baseline figure and cites the ADR; the `verify` aggregate member is unchanged.
3. **Given** the remaining residual edges, **Then** they stay recorded on the live issue [#101](https://github.com/gosharplite/tellme/issues/101) (not a frozen package).

**Functional Requirements**:

- **FR-006**: A new **ADR** (next free number, confirmed in research; `docs/decisions/NNNN-<slug>.md` + the `docs/decisions/README.md` index row) MUST record the de-coupling: the port, its home tier, the behaviour-preservation claim, the baseline movement, and its relation to ADR 0011/0016/0013 — including **what the change is *not*** (no user-facing change; no new normative source).
- **FR-007**: `specs/truth/techstack.md` (**Build & Tooling**) MUST be **MODIFY**d: the Layer-discipline gate row records the new RULE-E baseline figure (the removed edge(s) gone); any architectural-structure note (e.g. the CLI Application *Composition root* / *Project layout* row, or the *Interactive TUI prompt* row) that names the inverted seam MUST be updated to the injected-port reality; the `verify` aggregate member is unchanged.
- **FR-008**: The round MUST leave the de-coupling of the **remaining** residual edges **out of scope** (later slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)) and MUST state the round's slice boundary in `research.md`/`plan.md`.

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only**; **stdlib-only** for the port (no new module dependency; `go.mod`/`go.sum` unchanged).

### Edge Cases

- **A port that would leak a `ui`/`agent` type into `internal/domain/**`** → the port MUST live in `internal/app/**` instead (FR-002); a domain-purity breach fails RULE-C.
- **A test file (not production) importing the selected package** → the merged production+test graph governs; the import must be inverted there too (or the test moved), else RULE-E stays red.
- **A partially de-coupled round** (some call sites inverted, some not) → forbidden: `internal/cli` either imports the package (edge present → baseline unchanged) or it does not (edge gone); a half-refactor cannot shrink the baseline.
- **The removed edge re-appearing via a new call site in a later round** → **fails** RULE-E (not baselined) — the ratchet has teeth.
- **A stale baseline entry** after the de-coupling → **fails** (ADR 0011 D3 preserved).
- **The composition root no longer wiring the port** → a compile/harness failure (a missing injection is caught by the build + the `runTurn`/TUI unit seams), not a silent behavioural drift.
- **A cycle introduced by the new port home** → the SCC pass **fails** (cycles have no baseline).
- **OS/build-tag-gated files** → the `CROSS_TARGETS` union catches an OS-gated import; custom build-tag-gated files remain a recorded residual (ADR 0011 D6).
- **A `ui/tui/prompt` slice** → the injected port is the **prompt runner** seam (the existing `tuiPromptRunner` func type at `cli.go:173` is the natural shape; F-4's unexport may ride along if locked).

## Requirements *(mandatory)*

> Story-specific FR/NFR are attached under each story; this section holds only cross-story constraints.

### Global Requirements

#### Functional Requirements

- **FR-009**: The round MUST preserve **zero behavioural change** (FR-003) and MUST ship the de-coupling + baseline regeneration + truth/ADR records as **one PR** (NFR-002); it MUST NOT touch any other residual edge, any flag/exit-code/stream contract, or any domain business logic.
- **FR-010**: The round MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010): (a) re-introduce the removed edge ⇒ RULE-E reds; (b) leave a now-fixed edge in the baseline ⇒ the gate reds (stale); (c) a missing/false port injection ⇒ the relevant unit seam (or the build) fails loudly.
- **FR-011**: The round MUST NOT modify a delivered `specs/plans/NNN-*` package and MUST record the remaining slices + any deferred **F-4/F-6/F-7/F-8** item on the live issue [#101](https://github.com/gosharplite/tellme/issues/101).

#### Non-Functional Requirements

- **NFR-004**: The witness MUST be **the gate + unit seams**, NOT the E2E suite (#92 AC5); `make verify` (incl. `verify-architecture`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL row.
- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.

### Key Entities

- **The residual edge(s)** — the selected `internal/cli → internal/<pkg>` import(s) removed this round (Q1).
- **The injected port** — the domain- or application-tier seam that replaces the direct import (FR-002).
- **The baseline** — `tools/arch/baseline.txt`; loses exactly the removed edge(s) (3 → N), a fail-on-stale ratchet (ADR 0011 D3).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go list`/`grep` shows **no** `internal/cli` (production + test) import of the selected package(s); the RULE-E baseline lists exactly the remaining residual edges (regenerated from the gate); RULE-A/B/C stay **0**; cycles **0**. (covers FR-001, FR-004)
- **SC-002**: `make verify` is **green** at delivery; `go test -count=1 ./...` green (incl. the godog E2E, unmodified in behaviour); the Gherkin/DSL topology audit unchanged. (covers FR-003, NFR-004)
- **SC-003**: A deliberately re-introduced `internal/cli → <selected package>` import makes RULE-E **fail** and names the edge; a stale baseline entry makes the gate **fail** — both reproduced as falsifiability witnesses, then reverted. (covers FR-005, FR-010)
- **SC-004**: The new port lives in a **permitted tier** and drags no unsanctioned type across its boundary (`internal/domain/**` purity preserved); a missing/false injection fails loudly. (covers FR-002, FR-010, RULE-C)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the new baseline figure (citing the new ADR); the ADR + its index row are present; **no production behaviour** changes; `go.mod`/`go.sum` unchanged. (covers FR-006, FR-007, NFR-003, NFR-005)
- **SC-006**: The remaining residual edges + F-4/F-6/F-7/F-8 are recorded on [#101](https://github.com/gosharplite/tellme/issues/101); no frozen plan package is touched. (covers FR-008, FR-011)

## Assumptions

- **A1 (programme slice)** — this round is **R5.2** (the first de-coupling slice); the remaining edges are later slices. The slice selection is **Q1** (open — see §Open decision); the round is written slice-parameterised.
- **A2 (no E2E change)** — the refactor changes no `tellme` CLI behaviour, so the E2E suite is a **regression** carrier, not the witness; the witness is **the gate + unit seams** (NFR-004). `/axb-spec-by-example` is therefore **NOOP** — precedent: rounds 020/031/036/041/042/043/044/045/046/047.
- **A3 (port design is RD)** — the port's exact shape, its name, and its home tier are **RD** decisions (`research.md` D-x), subject to FR-002's purity constraint.
- **A4 (ADR required)** — a new ADR records the de-coupling (FR-006); the number is confirmed in research (next free after `0016`).
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP** (no persisted/runtime state change); `/axb-ui-plan` **skipped** (not a new user-facing surface).
- **A6 (`/axb-dsl-refine` NOOP)** — a `cli`-reader refactor is not a CLI-contract change: no feature Rule, Example, step, or `DSLRow` change; the topology audit stays unchanged.
- **A7 (scope guard)** — the round touches `internal/cli/**`, the new port's home tier, `cmd/tellme` (wiring), `tools/arch/baseline.txt`, `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`.

## Out of scope (recorded)

- The **de-coupling of the other residual edges** (the ones not selected by Q1) — later R5 slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101).
- **F-4/F-6/F-7/F-8** (the PR #104 review deferrals) — later slices, unless explicitly ridden along (a clarify question).
- Re-opening R2's decisions (composition-root home `cmd/tellme`, `internal/app/deps.Dependencies`, `agentTools()` relocation, MCP orchestration into `internal/infrastructure/mcp`).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting adapters or domain business logic; the yield policy (ADR 0014) and the presentation port (ADR 0015) patterns are reused, not re-litigated.
- A **re-ruling** of RULE-E's sanctioned set (unless explicitly locked in clarify).
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#101](https://github.com/gosharplite/tellme/issues/101) — the round's anchor (**R5**; the programme + the 3 residual edges + F-4/F-6/F-7/F-8).
- [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause) + **AC5** (witness = gate + unit seams).
- Round **047** ([#101](https://github.com/gosharplite/tellme/issues/101), ADR **0016**) — delivered **R5.1**: the RULE-E gate + the 3-entry baseline this round shrinks.
- Round **044** ([#100](https://github.com/gosharplite/tellme/issues/100), ADR **0013**) — the injected-`Dependencies` pattern + the composition root; PR [#104](https://github.com/gosharplite/tellme/pull/104) review `5243584043` — F-4/F-6/F-7/F-8.
- Round **046** ([#108](https://github.com/gosharplite/tellme/issues/108), ADR **0015**) — the loop-presentation port precedent (invert a direct import into an injected port, behaviour-preserving).

## References

- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the guard, tier table, and baseline.
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) (RULE-A/B/C/D + D3 ratchet + D7 normative table) · [`0016-application-import-ceiling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0016-application-import-ceiling.md) (RULE-E) · [`0013`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0013-composition-root-injection.md) · [`0015`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0015-loop-presentation-port.md).
- [`internal/cli/cli.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/cli.go) · [`internal/cli/call_renderer.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/call_renderer.go) — the application tier under RULE-E.
- [`specs/truth/techstack.md`](https://github.com/gosharplite/tellme/blob/dev/specs/truth/techstack.md) — Build & Tooling (the gate row).
