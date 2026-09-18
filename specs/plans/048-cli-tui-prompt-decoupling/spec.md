# Feature Specification: de-couple `internal/cli` from the TUI prompt — R5.2, first de-coupling slice (round 048)

**Feature Branch**: `048-cli-tui-prompt-decoupling`

**Created**: 2026-09-18

**Status**: Draft — plan package created by `/axb-specify`. Anchor issue [#101](https://github.com/gosharplite/tellme/issues/101) (**R5** of [#92](https://github.com/gosharplite/tellme/issues/92)). This round is **R5.2** — the **first de-coupling slice** (round 047 = **R5.1**, the **RULE-E** gate + baseline). Clarify round 1 **LOCKED**: **Q1 → B** (slice = the **TUI prompt** edge `internal/cli → internal/ui/tui/prompt`) · **Q2 → A** (the port lives in **`internal/domain/**`**) · **Q3 → A** (**F-4 is folded into this round**).

**Input**: Issue [#101](https://github.com/gosharplite/tellme/issues/101) — **R5** of the [#92](https://github.com/gosharplite/tellme/issues/92) gate-first split. After round 047 the layer-discipline guard (`tools/arch`, ADR 0011/0016) enforces **RULE-E — the application import ceiling** for the application tiers (`internal/app/**`, `internal/cli`), and the committed baseline records the **3** residual unsanctioned edges:

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
internal/cli -> internal/ui/tui/prompt
```

R5's remaining work is to remove these edges one de-coupling slice at a time (the baseline ratchets **3 → 0**). **Q1 → B**: this round removes exactly **`internal/cli -> internal/ui/tui/prompt`** (baseline **3 → 2**).

**Behaviour intent**: **MODIFY (behaviour-preserving refactor)** — invert the CLI's call into the interactive TUI prompt into an **injected domain port**, so `internal/cli` no longer imports `internal/ui/tui/prompt`; the RULE-E baseline loses exactly that edge. **No** user-facing behaviour change (`stdout`/`stderr` byte-contracts, flags, exit codes, DSL vocabulary) and no domain business logic — the round-044/045/046 lineage (gate-proven, behaviour-preserving structural refactors).

---

## Locked decisions (clarify round 1 — Q1 … Q3)

> Locked one at a time. Q1 fixes the slice; Q2 fixes the port home; Q3 folds F-4. All requested clarifications for this round are answered — no `NEEDS CLARIFICATION` remains.

| # | Decision |
| --- | --- |
| **Q1 → (B) the TUI prompt edge** | Round 048 de-couples exactly **`internal/cli → internal/ui/tui/prompt`** → the RULE-E baseline ratchets **3 → 2**. The other two residual edges (`→ agent`, `→ ui`) are **later slices** (tracked on [#101](https://github.com/gosharplite/tellme/issues/101)). Slug/branch: **`048-cli-tui-prompt-decoupling`** (renamed from the provisional `048-cli-strict-decoupling`). Rationale: the smallest, most self-contained residual edge — a single subsystem already behind the `tuiPromptRunner` func-type seam (`cli.go:173`/`180`), so the inversion is a one-seam change with the cleanest behaviour-preservation proof and the smallest review surface. |
| **Q2 → (A) the port lives in `internal/domain/**`** | The injected port is a **domain-owned contract** (a small new domain seam, e.g. `internal/domain/tui`, holding the `Run`-shaped interface + a domain-declared `Source{ Suggest(ctx, query) []string }`); it references **only stdlib** (`context`/`io`/`time`) + domain types (RULE-C-pure — no `ui` type). The TUI side **satisfies it structurally**; the wiring/adapter sits in a tier ≥ 5 (`internal/ui/**`) or in the exempt composition root (`cmd/tellme`). Rationale: a port is a domain-facing contract; consistent with ADR 0015 (the loop-presentation port declared in `internal/domain/agent`) and with [#101](https://github.com/gosharplite/tellme/issues/101) open-decision 3. Rejected: **(B)** a home in `internal/app/**` (an application *utility*, whereas this is a plain outbound contract). |
| **Q3 → (A) F-4 folded in** | The PR #104 deferral **F-4** (`cli.Options.RunTUIPrompt` is an exported field of an unexported type `tuiPromptRunner`) is **removed for free** by Q1's rewrite: the field and the `tuiPromptRunner` func type MUST be **deleted**; the port is injected instead (`cli.Options` is reduced to `Deps`, or the runner is supplied as the port). F-4 therefore **closes** in this round — it does **not** affect the RULE-E baseline (still 3 → 2). Rationale: F-4 is the same seam being rewritten; leaving it would keep a contradictory exported-field-of-unexported-type interim state and require a second pass. |

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `e5db873` (static import scan; the gate re-measures at implementation).

### The selected residual edge and its call sites (`internal/cli` production files)

| Item | Detail |
| --- | --- |
| Edge | `internal/cli -> internal/ui/tui/prompt` |
| Import sites | `internal/cli/cli.go` only (aliased `tuiprompt`) |
| Call sites | `tuiprompt.Run(ctx, env.stdin, env.stderr, src, tuiDebounceDuration())` (`cli.go:202`, inside `defaultRunTUIPrompt`) · `tuiprompt.DefaultDebounceDuration` (`cli.go:214`, inside `tuiDebounceDuration`) |
| Existing seam | `type tuiPromptRunner func(ctx context.Context, res resolution, env runtimeEnv, dp deps.Dependencies) (string, bool, error)` (`cli.go:173`) + `defaultRunTUIPrompt` (`cli.go:180`), selected in `runTUIPrompt` via `opts.RunTUIPrompt` (nil ⇒ default) |
| The call the port must carry | `Run(ctx, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)` — `internal/ui/tui/prompt/run.go:18`; `Source` = `Suggest(ctx, query) []string` (`model.go:37`) |
| The test seam | `Options.RunTUIPrompt` is set by `tui_dispatch_test.go` / `tui_submit_chrome_test.go` (the seam adapted to the port, Q3) |

### Tier constraint that shapes the port

ADR 0011 **D1** ranks `domain` 0 · `config`/`home` 1 · `app` 2 · `infrastructure` 3 · `agent` 4 · `ui`/`ui/tui/**` 5 · `cli` 6 (higher = more upward). **RULE-A** forbids an *upward* import, so **only tiers ≥ 5 may import `internal/ui/**`** — the adapter that actually calls `tuiprompt.Run` stays **inside `internal/ui/**`** (or in the exempt `cmd/tellme`). The **port declaration** lands in `internal/domain/**` (Q2-A); `internal/cli` then imports only the domain port + its own adapter plumbing. `internal/domain/**` must stay **RULE-C-pure** (stdlib + domain types only — no `ui` type crosses in).

### The gate that measures the DoD (round 047 / ADR 0016)

- **RULE-E** (ADR 0016): a governed application tier (`internal/app/**`, `internal/cli`) may import, beyond stdlib, only `internal/domain/**`, `internal/config`, `internal/home`, `internal/app/**`; the sanctioned set is a **normative** section of the gate's tier table (ADR 0011 **D7**), **default-deny**, **fail-on-stale**.
- The committed **baseline** (`tools/arch/baseline.txt`) is a **fail-on-stale ratchet** (ADR 0011 **D3**): a new violation fails; a stale entry fails; removing an entry while its violation persists fails.
- **DoD is the gate**, not the E2E suite (#92 **AC5**, the round-009 false-confidence trap).

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E reports **0 new / 0 stale**; 0 cycles; cross-compile 4/4.
- **Zero behavioural change**: identical TUI prompt chrome/submit/abort behaviour, `stdout` byte-exact, `stderr` line contracts unchanged, flags/exit codes unchanged, no new Gherkin/DSL row.
- **`internal/domain/**` stays 100 % pure**: the new port references only stdlib + domain types; no `ui` type crosses the boundary (RULE-C).
- **Tests stay hermetic** (no mutable package globals; round-029/030 precedent); the existing `tuiPromptRunner` test seam (`tui_dispatch_test.go`, `tui_submit_chrome_test.go`) is adapted to inject a **port** double (Q3).
- The **existing RULE-E tier table** is extended in place — no new Makefile target, no new normative source.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the TUI prompt is reached through an injected domain port, so `internal/cli` no longer imports `internal/ui/tui/prompt` (Priority: P1)

As a maintainer, I want the interactive TUI prompt to be invoked through an **injected domain port** wired at the composition root, so `internal/cli` no longer imports `internal/ui/tui/prompt` and the RULE-E baseline can drop that edge — with **no** behavioural change to `tellme -i`.

**Why this priority**: it is the round's entire reason to exist — the first de-coupling step of R5.

**Independent verification**: after the round, `go list`/`grep` shows **no** `internal/cli` (production or test) import of `internal/ui/tui/prompt`; `make verify-architecture` reports the RULE-E baseline with that edge gone; the godog E2E + the TUI unit pins stay green; a deliberately re-introduced import reds the gate.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** `make verify` runs, **Then** the gate is **green** and `baseline.txt` no longer lists `internal/cli -> internal/ui/tui/prompt`.
2. **Given** `tellme -i` runs (submit and abort paths), **When** the prompt engages, **Then** the diagnostic hint, the suggestion chrome, the composed-prompt return, the abort/no-op behaviour, and the absence of `stdout` writes are **unchanged** (godog E2E + the TUI unit pins green).
3. **Given** a change that re-introduces `internal/cli -> internal/ui/tui/prompt`, **When** `make verify` runs, **Then** RULE-E **fails** and names the edge.
4. **Given** the new domain port, **When** it is declared, **Then** it references only stdlib + domain types (RULE-C-pure) and the adapter that calls `tuiprompt.Run` stays inside a tier ≥ 5; the port is wired at `cmd/tellme`.

**Functional Requirements**:

- **FR-001**: The **`internal/cli → internal/ui/tui/prompt`** import MUST be removed from the production **and test** import graph of `internal/cli`; the TUI prompt MUST be reached through an **injected domain port** declared in `internal/domain/**` and wired at the composition root.
- **FR-002**: The port MUST be a **domain-owned contract** referencing only stdlib (`context`/`io`/`time`) + domain types (RULE-C purity — **no** `ui` type crosses in); the TUI side satisfies it **structurally**, and the adapter that calls `tuiprompt.Run` MUST live in a tier ≥ 5 (`internal/ui/**`) or in the exempt composition root. The exact port home (a new small package such as `internal/domain/tui`, or an existing domain owner) and its name are **RD** decisions recorded in `research.md` + the ADR.
- **FR-003**: The implementation MUST be **behaviour-preserving**: the diagnostic hint (`TUIHint`), the suggestion-engine wiring, the debounce resolution (`TELL_ME_TUI_DEBOUNCE` seam + `DefaultDebounceDuration` fallback), the `(text, ok, err)` contract, the "never writes to `stdout`" rule, and the abort/empty-submit no-op MUST be unchanged. The existing tests MUST pass **unmodified** except where a test directly names the inverted seam (itemised as test-only adaptation; the dispatch/submit tests are re-pointed to inject a port double).
- **FR-004**: The PR #104 deferral **F-4** MUST be **removed**: `cli.Options.RunTUIPrompt` (an exported field of the unexported type `tuiPromptRunner`) and the `tuiPromptRunner` func type MUST be **deleted**, replaced by the injected port; `cli.Options` exposes no field of an unexported type. F-4 closes this round.

**Non-Functional Requirements**:

- **NFR-001**: The composition root (`cmd/tellme`) MUST remain the single assembly site for the port; `internal/cli` MUST NOT construct the TUI implementation itself, and MUST NOT retain a nil-defaulted in-package fallback that imports the TUI package.

### User Story 2 - the RULE-E baseline shrinks 3 → 2 — a gate-proven ratchet (Priority: P1)

As a maintainer, I want the committed baseline to lose exactly the removed edge and to fail if it drifts, so this de-coupling is machine-verified and the ratchet can only shrink toward **0**.

**Why this priority**: the DoD is the gate (#92 AC5); a green E2E suite alone is false confidence.

**Independent verification**: read `tools/arch/baseline.txt` — it lists the 3 residuals **minus** `internal/cli -> internal/ui/tui/prompt`; run the gate — green; re-add the import — red; leave the now-fixed edge in the baseline — red (stale).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** `baseline.txt` is **regenerated from the gate** and lists exactly `internal/cli -> internal/agent` + `internal/cli -> internal/ui` (the TUI edge gone).
2. **Given** the removed edge is re-introduced, **When** the gate runs, **Then** it **fails** (RULE-E, not baselined).
3. **Given** a stale entry (the TUI edge left in the baseline), **When** the gate runs, **Then** it **fails** and names the entry.
4. **Given** the eventual baseline at **0** (after the later slices), **When** a stale entry is added, **Then** the gate **fails** (the round-046 anti-bypass rule holds).

**Functional Requirements**:

- **FR-005**: `tools/arch/baseline.txt` MUST be regenerated from the gate (`make verify-architecture-update`) after the de-coupling, dropping exactly `internal/cli -> internal/ui/tui/prompt`; no RULE-A/B/C entry appears; the gate is **green on `dev` at delivery**.
- **FR-006**: The round MUST NOT add any baseline entry, weaken RULE-E, or exempt the removed edge; the sanctioned set (ADR 0016) is unchanged.

**Non-Functional Requirements**:

- **NFR-002**: The de-coupling + the baseline regeneration MUST land as **one atomic delivery** (the round-040 TD-1 precedent: a stricter baseline must never land before the code is compliant).

### User Story 3 - the de-coupling and the port are recorded in truth and an ADR (Priority: P2)

As a maintainer/operator, I want the new domain port, its home, and the shrinking baseline recorded in `specs/truth/techstack.md` (Build & Tooling) and a new ADR, so the remaining R5 slices can cite the pattern and the project's layered-architecture statement stays current.

**Why this priority**: it bounds the round in truth/governance and makes the pattern reusable for the remaining slices.

**Independent verification**: read the new ADR + its index row; read the Layer-discipline gate row — it records the new baseline figure (2) and cites the ADR; the guard self-test still passes.

**Acceptance Scenarios**:

1. **Given** the round lands, **Then** a new **ADR** records the domain port, its home, and the ratchet movement (3 → 2), citing ADR **0011/0016/0013/0015**.
2. **Given** the technology-stack truth, **Then** the Layer-discipline gate row records the baseline figure **2** and cites the ADR; the *Interactive TUI prompt* row (if any) reflects the injected-port wiring; the `verify` aggregate member is unchanged.
3. **Given** the remaining residual edges, **Then** they stay recorded on the live issue [#101](https://github.com/gosharplite/tellme/issues/101) (not a frozen package).

**Functional Requirements**:

- **FR-007**: A new **ADR** (next free number, confirmed in research; `docs/decisions/NNNN-<slug>.md` + the `docs/decisions/README.md` index row) MUST record the de-coupling: the **domain** port, its home, the behaviour-preservation claim, the baseline movement **3 → 2**, the **F-4 closure**, and its relation to ADR 0011/0016/0013/0015 — including **what the change is *not*** (no user-facing change; no new normative source; the other two edges untouched).
- **FR-008**: `specs/truth/techstack.md` (**Build & Tooling**) MUST be **MODIFY**d: the Layer-discipline gate row records the new RULE-E baseline figure (**2**) and cites the ADR; any architectural row that names the inverted seam (e.g. CLI Application *Interactive TUI prompt* / *Composition root*) MUST be updated to the injected-domain-port reality; the `verify` aggregate member is unchanged.
- **FR-009**: The round MUST leave the de-coupling of the **other two** residual edges (`→ agent`, `→ ui`) and **F-6/F-7/F-8** **out of scope** (later slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)), and MUST state the round's slice boundary in `research.md`/`plan.md`.

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only**; **stdlib-only** for the port (no new module dependency; `go.mod`/`go.sum` unchanged).

### Edge Cases

- **A domain port that would leak a `ui` type** → forbidden: the port MUST reference only stdlib + domain types (Q2-A / RULE-C); if a `ui` type proved necessary the port home would have to change (not the case here — the seam is `Run(ctx, io.Reader, io.Writer, Source, time.Duration)`).
- **The adapter that calls `tuiprompt.Run`** → it MUST live in a tier ≥ 5 (`internal/ui/**`) or in the exempt `cmd/tellme`, because RULE-A forbids a lower tier importing `internal/ui/**`.
- **A test file (not production) importing the TUI package** → the merged production+test graph governs; the test import must be inverted/moved too, else RULE-E stays red.
- **The `tuiPromptRunner` test seam** (`tui_dispatch_test.go`, `tui_submit_chrome_test.go`) → it MUST be adapted to inject a **port double**, not deleted silently; the field `Options.RunTUIPrompt` and the func type are removed (F-4).
- **A partially de-coupled round** (some call sites inverted, some not) → forbidden: `internal/cli` either imports the package (edge present → baseline unchanged) or it does not (edge gone).
- **The removed edge re-appearing via a new call site in a later round** → **fails** RULE-E (not baselined) — the ratchet has teeth.
- **A stale baseline entry** after the de-coupling → **fails** (ADR 0011 D3 preserved).
- **The composition root not wiring the port** → a compile/harness failure (a missing injection is caught by the build + the TUI dispatch unit seam), not a silent behavioural drift.
- **A cycle introduced by the new domain package** → the SCC pass **fails** (cycles have no baseline).
- **OS/build-tag-gated files** → the `CROSS_TARGETS` union catches an OS-gated import; custom build-tag-gated files remain a recorded residual (ADR 0011 D6).

## Requirements *(mandatory)*

> Story-specific FR/NFR are attached under each story; this section holds only cross-story constraints.

### Global Requirements

#### Functional Requirements

- **FR-010**: The round MUST preserve **zero behavioural change** (FR-003) and MUST ship the de-coupling + baseline regeneration + truth/ADR records as **one PR** (NFR-002); it MUST NOT touch the other two residual edges, any flag/exit-code/stream contract, or any domain business logic.
- **FR-011**: The round MUST include **falsifiability witnesses** reproduced then reverted (ADR 0010): (a) re-introduce `internal/cli → internal/ui/tui/prompt` ⇒ RULE-E reds; (b) leave the now-fixed TUI edge in the baseline ⇒ the gate reds (stale); (c) a missing/false port injection ⇒ the TUI dispatch unit seam (or the build) fails loudly.
- **FR-012**: The round MUST NOT modify a delivered `specs/plans/NNN-*` package and MUST record the remaining slices + **F-6/F-7/F-8** on the live issue [#101](https://github.com/gosharplite/tellme/issues/101) (F-4 **closes** this round, Q3).

#### Non-Functional Requirements

- **NFR-004**: The witness MUST be **the gate + unit seams**, NOT the E2E suite (#92 AC5); `make verify` (incl. `verify-architecture`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL row.
- **NFR-005**: **stdlib-only** — no new module dependency; `go.mod`/`go.sum` unchanged.

### Key Entities

- **The removed edge** — `internal/cli -> internal/ui/tui/prompt` (production + test).
- **The injected domain port** — the `internal/domain/**` seam carrying `Run(ctx, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)` semantics + a domain `Source{ Suggest(ctx, query) []string }` (Q2-A).
- **The adapter** — the TUI-side value (inside `internal/ui/**`) or the exempt `cmd/tellme` wiring that satisfies the port.
- **The removed seam (F-4)** — `cli.Options.RunTUIPrompt` + the `tuiPromptRunner` func type (deleted).
- **The baseline** — `tools/arch/baseline.txt`; loses exactly the TUI edge (**3 → 2**), a fail-on-stale ratchet (ADR 0011 D3).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go list`/`grep` shows **no** `internal/cli` (production + test) import of `internal/ui/tui/prompt`; the RULE-E baseline lists exactly the two remaining residuals; RULE-A/B/C stay **0**; cycles **0**. (covers FR-001, FR-005)
- **SC-002**: `make verify` is **green** at delivery; `go test -count=1 ./...` green (incl. the godog E2E and the adapted TUI unit pins); the Gherkin/DSL topology audit unchanged. (covers FR-003, NFR-004)
- **SC-003**: A deliberately re-introduced `internal/cli → internal/ui/tui/prompt` import makes RULE-E **fail** and names the edge; a stale baseline entry makes the gate **fail** — both reproduced as falsifiability witnesses, then reverted. (covers FR-006, FR-011)
- **SC-004**: The port lives in **`internal/domain/**`** referencing only stdlib + domain types (RULE-C preserved; no `ui` type crosses in); the adapter stays in a tier ≥ 5; a missing injection fails loudly; **F-4 is closed** (`Options` exposes no field of an unexported type). (covers FR-002, FR-004, FR-011)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the baseline figure **2** (citing the new ADR); the ADR + its index row are present; **no production behaviour** changes; `go.mod`/`go.sum` unchanged. (covers FR-007, FR-008, NFR-003, NFR-005)
- **SC-006**: The two remaining residual edges + F-6/F-7/F-8 are recorded on [#101](https://github.com/gosharplite/tellme/issues/101) (F-4 closed); no frozen plan package is touched. (covers FR-009, FR-012)

## Assumptions

- **A1 (programme slice)** — this round is **R5.2**; the slice is **the TUI prompt edge** (Q1-B). The other two edges are later slices.
- **A2 (no E2E change)** — the refactor changes no `tellme` CLI behaviour, so the E2E suite is a **regression** carrier, not the witness; the witness is **the gate + unit seams** (NFR-004). `/axb-spec-by-example` is therefore **NOOP** — precedent: rounds 020/031/036/041–047.
- **A3 (port design is RD)** — the port's exact package/name/shape are **RD** decisions (`research.md` D-x), subject to FR-002 / RULE-A·C.
- **A4 (ADR required)** — a new ADR records the de-coupling (FR-007); the number is confirmed in research (next free after `0016`).
- **A5 (no other interfaces)** — `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP** (no persisted/runtime state change); `/axb-ui-plan` **skipped** (no new user-facing surface).
- **A6 (`/axb-dsl-refine` NOOP)** — a `cli`-reader refactor is not a CLI-contract change: no feature Rule, Example, step, or `DSLRow` change; the topology audit stays unchanged.
- **A7 (scope guard)** — the round touches `internal/cli/**`, the new `internal/domain/**` port package, `internal/ui/**` (the adapter, if any), `cmd/tellme` (wiring), `tools/arch/baseline.txt`, `docs/decisions/**`, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's `session-summary.md`.

## Out of scope (recorded)

- The other two residual edges — `internal/cli → internal/agent` and `internal/cli → internal/ui` (later R5 slices, tracked on [#101](https://github.com/gosharplite/tellme/issues/101)).
- **F-6/F-7/F-8** (the remaining PR #104 deferrals: narrow seams ≤2 fields; named `Discovery{Closer}`; `OutputSink` → interface) — later slices. **F-4 is folded in** (Q3).
- Re-opening R2's decisions (composition-root home `cmd/tellme`, `internal/app/deps.Dependencies`, `agentTools()` relocation, MCP orchestration into `internal/infrastructure/mcp`).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting adapters or domain business logic; the yield policy (ADR 0014) and the presentation port (ADR 0015) patterns are reused, not re-litigated.
- A **re-ruling** of RULE-E's sanctioned set.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#101](https://github.com/gosharplite/tellme/issues/101) — the round's anchor (**R5**; the programme + the 3 residual edges + F-4/F-6/F-7/F-8).
- [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** (second clause) + **AC5** (witness = gate + unit seams).
- Round **047** ([#101](https://github.com/gosharplite/tellme/issues/101), ADR **0016**) — delivered **R5.1**: the RULE-E gate + the 3-entry baseline this round shrinks.
- Round **044** ([#100](https://github.com/gosharplite/tellme/issues/100), ADR **0013**) — the injected-`Dependencies` pattern + the composition root; PR [#104](https://github.com/gosharplite/tellme/pull/104) review `5243584043` — F-4/F-6/F-7/F-8.
- Round **046** ([#108](https://github.com/gosharplite/tellme/issues/108), ADR **0015**) — the loop-presentation port precedent (invert a direct import into an injected domain port, behaviour-preserving).

## References

- [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) · [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) — the guard, tier table, and baseline.
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) (RULE-A/B/C/D + D3 ratchet + D7 normative table) · [`0016-application-import-ceiling.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0016-application-import-ceiling.md) (RULE-E) · [`0013`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0013-composition-root-injection.md) · [`0015`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0015-loop-presentation-port.md).
- [`internal/cli/cli.go`](https://github.com/gosharplite/tellme/blob/dev/internal/cli/cli.go) (the `tuiPromptRunner` seam + `defaultRunTUIPrompt`) · [`internal/ui/tui/prompt/run.go`](https://github.com/gosharplite/tellme/blob/dev/internal/ui/tui/prompt/run.go) · [`internal/ui/tui/prompt/model.go`](https://github.com/gosharplite/tellme/blob/dev/internal/ui/tui/prompt/model.go) (`Source`).
- [`specs/truth/techstack.md`](https://github.com/gosharplite/tellme/blob/dev/specs/truth/techstack.md) — Build & Tooling (the gate row).
