# Feature Specification: layer-discipline gate + violation baseline — the pipeline detects a new illegal layer import (round 042)

**Feature Branch**: `042-layer-discipline-gate-plan`

**Created**: 2026-09-17

**Status**: Draft — plan half only (spec → clarify → spec-by-example → technical-research → system-analysis → dsl-refine). Anchor issue [#93](https://github.com/gosharplite/tellme/issues/93); parent umbrella [#92](https://github.com/gosharplite/tellme/issues/92) **R1**.

**Input**: Operator request to run **R1 of [#92](https://github.com/gosharplite/tellme/issues/92)** as its own plan package — *"Create a detail new issue just to do R1"* → [#93](https://github.com/gosharplite/tellme/issues/93). R1 is: *add a `verify-architecture` analogue (an import-direction check over `domain → app → infrastructure → cli/ui`) to `make verify`, shipping with a **baseline file** that records today's 7 violations so `dev` stays green.* Origin: round-039 review **TD-3** (PR [#81](https://github.com/gosharplite/tellme/pull/81)) — *the missing layer-discipline gate*.

**Behaviour intent**: **ADD** a **layer-discipline verification gate** to `tellme`'s quality pipeline, plus a **committed baseline** of the layer violations that already exist, so that (a) a *new* illegal layer import fails the standard gate instead of sailing through, and (b) the existing violation count can only be **ratcheted down** (7 → 0) by the later rounds (R2). It is a **tooling/gate slice** in the round-020 `verify-cross-compile` / round-032 `verify-mcp-sdk-confinement` lineage: it changes **no** user-facing behaviour and **no** product code — it adds a gate, its baseline, and its truth record.

---

## Grounded in the current system

- **`internal/cli` is both the application layer and the composition root.** `internal/cli/cli.go` (~1236 lines) directly imports five `internal/infrastructure/*` packages (`history`, `llm`, `skills`, `telemetry`, `tools`) and holds package-level factory vars (`newGateway`, `newHistoryStore`, `newUsageStore`, `newToolUsageStore`, `newToolRegistry`, `newRenderer`, `newTUIPromptRunner`, `userHomeDir`); `internal/cli/mcp_discovery.go` imports `internal/infrastructure/di` + `internal/infrastructure/mcp`. Static analysis reports **7 layer violations (0 cycles)** — these exact 7 edges:

  | # | File | Illegal import |
  | --- | --- | --- |
  | 1 | `internal/cli/cli.go` | `internal/infrastructure/history` |
  | 2 | `internal/cli/cli.go` | `internal/infrastructure/llm` |
  | 3 | `internal/cli/cli.go` | `internal/infrastructure/skills` |
  | 4 | `internal/cli/cli.go` | `internal/infrastructure/telemetry` |
  | 5 | `internal/cli/cli.go` | `internal/infrastructure/tools` |
  | 6 | `internal/cli/mcp_discovery.go` | `internal/infrastructure/di` |
  | 7 | `internal/cli/mcp_discovery.go` | `internal/infrastructure/mcp` |

  (Measured 2026-09-17 at `dev` `3f8ec08` from the module's own import graph; the baseline MUST be generated from the gate's own output, never hand-transcribed.)

- **`make verify` detects no layer violation today.** The aggregate is `verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + lint + vulncheck` (`Makefile`). None of these checks import direction, so an eighth illegal import would pass every gate — only a human reviewer would notice.

- **Two established gate shapes already exist** to model on: the **Makefile grep** (`verify-mcp-sdk-confinement` — `grep -rn` for a forbidden import path outside an allow-listed dir) and the **make → `go test -run …`** delegation (`verify-no-network`). The reference `tell-me-go` ships a **build-tagged Go guard** (`make verify-architecture` → `go test -tags=arch -run TestVerifyRealArchitecture ./internal/tools/analysis -args -strict-arch=true` + `modelith-layers`) — `SESSION-BOOTSTRAP.md` §2.2.

- **The real import graph (measured, whole module)** shows edges a *broad* direction check would also have to rule on — e.g. `internal/agent → internal/ui` and `internal/infrastructure/llm|mcp → internal/config` — but **no** current violation inside `internal/domain/**` (it is pure today) and **no** infra import from any `internal/cli/*_test.go`. The `tests/e2e/**` harness **does** import `internal/infrastructure/**`.

- **The measured baseline is exactly the 7 `internal/cli → internal/infrastructure` edges** — *provided* the rule model and the package scope are the ones assumed below; both are **open** (see *Open clarify gaps*).

- **`specs/truth/techstack.md` (Build & Tooling)** already records the gate catalog and the `verify` aggregate list, so adding a gate is a real **truth MODIFY** (not a NOOP).

---

## Open clarify gaps (round-042 `/axb-clarify`, one question at a time — budget 1–3)

> These three gaps are **high-impact** (they change the scope boundary, the rule definition, and the formal success criteria). Each is marked `[NEEDS CLARIFICATION]` at its requirement below and is asked in `/axb-clarify` before the spec is frozen. Lower-impact details (the guard's exact implementation form; whether R1 warrants its own ADR) are **not** escalated — they are RD decisions in `/axb-technical-research`.

- **Q1 — Rule model.** Is the gate **narrow** (the target rule: `internal/cli` (and `internal/ui`?) MUST NOT import `internal/infrastructure`, plus `internal/domain` purity — baseline = the 7 edges) or **broad** (a full import-direction check over the ordering `domain → config/home → app → infrastructure → agent/ui → cli → cmd`, which additionally flags today's `internal/agent → internal/ui` edge unless the ranking permits it)? The reference's `verify-architecture` is the broad form.
- **Q2 — Package scope of the gate.** Which packages does the gate govern: production `internal/**` only — or also `_test.go` imports and the `tests/**` harness (which imports infrastructure) and `cmd/**`? This sets the baseline's exact contents.
- **Q3 — Stale-entry policy.** When a baselined entry no longer violates (e.g. R2 fixes one), does the gate **fail** (forcing the baseline to shrink — gives the ratchet teeth) or **report a warning**?

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A new illegal layer import fails the standard gate (Priority: P1)

As a maintainer, I want `make verify` to fail when a change introduces a new illegal cross-layer import, so that layer discipline is enforced mechanically instead of relying on a human reviewer noticing.

**Why this priority**: it is the round's entire reason to exist and the precondition for every later round in [#92](https://github.com/gosharplite/tellme/issues/92) — without it, R2's "7 → 0" claim is unfalsifiable. Today **no** gate in `make verify` checks import direction.

**Independent verification**: add a *new* illegal import somewhere the gate governs, run the aggregate `make verify` — it exits non-zero and names the offending `path → import`; revert and it exits 0.

**Acceptance Scenarios**:

1. **Given** the current tree, **When** `make verify` runs, **Then** the layer-discipline gate runs, reports **0** violations beyond the baseline, and the aggregate exits 0.
2. **Given** a change that adds an 8th illegal import under the gate's scope, **When** `make verify` runs, **Then** the gate exits non-zero and its output **names the offending source path and import**.
3. **Given** the aggregate command, **When** it runs, **Then** it includes the layer-discipline gate (a violation fails the standard gate, not a side command).

**Functional Requirements**:

- **FR-001**: The quality pipeline MUST include a **layer-discipline gate** that computes each governed package's internal imports and flags any that violate the pinned layer rule.
- **FR-002**: On any violation **beyond the baseline**, the gate MUST exit non-zero and MUST identify the offending source path and import (machine- and human-readable).
- **FR-003**: The gate MUST be a member of the aggregate verification command (`make verify`), so a new violation fails the standard gate.
- **FR-004**: The gate MUST be **hermetic and host-independent** — it MUST NOT depend on the host `GOOS`/`GOARCH`, on an ambient environment export, or on the network; it MUST add no dependency or tool beyond the Go toolchain and `make`.

**Non-Functional Requirements**:

- **NFR-001**: The gate MUST be **deterministic** (a stable, sorted, diff-friendly report) and MUST NOT use `time.Sleep` or otherwise depend on wall-clock/timing.
- **NFR-002**: **stdlib-only** — no new module dependency; `go.mod` / `go.sum` unchanged.

---

### User Story 2 - The existing violations are baselined, and the count only ratchets down (Priority: P2)

As a maintainer, I want the 7 violations that already exist recorded in a committed baseline so that `dev` is green **today**, the gate can land immediately, and each later fix (R2) must **remove** a baseline entry — so the violation count can only decrease.

**Why this priority**: it is what makes the gate *landable now* while the refactor (R2) is still pending. Without a baseline the gate would redden `dev` on day one; without the ratchet the baseline would silently rot.

**Independent verification**: read the baseline — it lists exactly today's violations in the gate's own format; run the gate — green; remove one baseline entry while its violation still exists and re-run — the gate **fails**; fix a violation but leave its entry and re-run — the gate behaves per the pinned stale-entry policy (Q3).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** it is **green** and the baseline lists the current violations (7, under the assumed model/scope).
2. **Given** the baseline and an unfixed violation whose entry was removed, **When** the gate runs, **Then** it **fails** (the baseline is not a blanket allow-list of "anything").
3. **Given** a fixed violation (its code no longer violates), **When** the gate runs, **Then** it behaves per the pinned **stale-entry policy** (Q3): **fail** (forcing the entry's removal) or **report**.
4. **Given** the gate's baseline format, **When** R2 fixes one violation, **Then** exactly one baseline entry is removed and the change is visible in the diff.

**Functional Requirements**:

- **FR-005**: A **baseline file** MUST be committed, listing the currently-known violations in the **gate's own output format**, so the gate is green on `dev` at delivery.
- **FR-006**: The baseline MUST NOT be a blanket allow-list: a violation **not** present in the baseline MUST fail the gate (FR-002).
- **FR-007**: The gate's stale-entry behaviour MUST be **pinned** — an entry whose violation no longer exists is **fail** or **report** per the Q3 decision — and the baseline MUST support removing **exactly one** entry per fixed violation.
- **FR-008**: The gate MUST NOT forbid the legitimate, recorded seams (e.g. it MUST NOT flag `agentTools()` — the parameterless, read-free assembler consumed by the round-031 well-formedness gate).

**Non-Functional Requirements**:

- **NFR-003**: The baseline MUST be stable (sorted, line-based) and MUST be generated from the gate's own output, not hand-transcribed (repo lesson: no unverified transcription).

---

### User Story 3 - The new gate is recorded in the technology-stack truth (Priority: P3)

As a maintainer/operator, I want the layer-discipline gate recorded in `specs/truth/techstack.md` (Build & Tooling) and named in the `make verify` aggregate list, so the gate is discoverable and does not become an orphan Makefile target.

**Why this priority**: it bounds the round in truth. It is a documentation/wiring constraint on Stories 1–2, not an independent capability.

**Independent verification**: read `specs/truth/techstack.md` (Build & Tooling) — the gate and the baseline are recorded, and the `verify` aggregate list includes the new member; assert every pre-round gate's behaviour is unchanged.

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** the Build & Tooling section records the layer-discipline gate and the baseline policy.
2. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing gate's behaviour is unchanged and the `verify` aggregate only **gains** a member.

**Functional Requirements**:

- **FR-009**: The gate and its baseline policy MUST be recorded in `specs/truth/techstack.md` (Build & Tooling), and the **Task runner** row's `verify` aggregate list MUST name the new member.
- **FR-010**: The round MUST NOT change the behaviour of any existing gate or the meaning of the aggregate `verify` beyond adding the new member; it MUST NOT change any `stdout`/`stderr` behaviour of the `tellme` binary, any flag, or any exit code.

**Non-Functional Requirements**:

- **NFR-004**: **POSIX-only** (Linux/macOS); no Windows variant. The round adds **no** new third-party dependency (`go.mod` / `go.sum` unchanged).

---

### Edge Cases

- **Gate lands before its baseline** → forbidden: the gate and its baseline MUST land as **one atomic delivery** (round-040 **TD-1** precedent — a stricter gate that lands before its enablers reddens `dev`).
- **Host `GOOS`/`GOARCH` or ambient `CGO_ENABLED`** → the gate's verdict MUST be identical on every host (mirrors `verify-cross-compile`'s hermeticity rationale).
- **A *new* violation while the baseline already lists others** → the gate MUST distinguish "new" from "baselined" and fail on the new one (FR-006).
- **A *fixed* violation still in the baseline** → resolved by the Q3 stale-entry policy (FR-007).
- **Test files / the `tests/**` harness and `cmd/**`** → whether these are in scope is Q2; if out of scope, the gate MUST say so explicitly (no silent blind spot).
- **Build tags / files excluded on the host** → the gate MUST evaluate the module's package imports (not a naive per-file grep) so a build-tagged illegal import cannot hide.
- **The round's own gate files** → the guard and its self-test MUST NOT themselves introduce a violation under the rule they enforce.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-011**: The round MUST be **tooling + truth only**: it MUST add the gate, its self-tests, the baseline, and the truth/spec records, and MUST NOT modify any production Go behaviour or any existing gate's semantics.
- **FR-012**: The round MUST pin the **rule model** (Q1) and the **package scope** (Q2) explicitly in `research.md`/`plan.md`, and MUST re-measure the baseline from the gate itself before freezing it. [NEEDS CLARIFICATION: Q1 rule model — narrow vs broad; Q2 package scope — production only vs incl. `_test.go`/`tests`/`cmd`]

#### Non-Functional Requirements

- **NFR-005**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows unless `/axb-dsl-refine` decides the gate's dev-observable behaviour is a CLI-truth interface.

### Key Entities *(include if feature involves data)*

- **Layer rule**: the pinned allowed import-direction model (the ordering of `domain`, `config`/`home`, `app`, `infrastructure`, `agent`, `ui`, `cli`, `cmd`) and which edges are forbidden.
- **Layer violation**: a governed package's internal import that violates the rule — reported as `source-path → imported-path`.
- **Baseline**: the committed, sorted file listing the currently-known violations in the gate's format; a ratchet that only shrinks.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make verify` runs the layer-discipline gate; on `dev` it is **green** with the baseline listing the current violations (the baselined count, to be re-measured at freeze). (covers FR-001, FR-003, FR-005)
- **SC-002**: A deliberately added **new** illegal import makes the gate **fail** and name the offending `path → import` — reproduced as a falsifiability witness, then reverted (ADR 0010 doctrine). (covers FR-002, FR-006)
- **SC-003**: The baseline is a **ratchet**: removing an entry for an unfixed violation fails the gate; a fixed violation is handled per the pinned stale-entry policy (Q3). (covers FR-006, FR-007)
- **SC-004**: The gate's verdict is identical across hosts (no dependence on host `GOOS`/`GOARCH`/`CGO_ENABLED`), is deterministic/sorted, and adds no dependency (`go.mod`/`go.sum` unchanged). (covers FR-004, NFR-001, NFR-002, NFR-004)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the gate + the `verify` aggregate member, every pre-round gate's behaviour is unchanged, **no production Go behaviour** changes, and the Gherkin/DSL topology audit is unchanged/green. (covers FR-009, FR-010, FR-011, NFR-005)

## Assumptions

- **A1 (intent)** — R1 is a **ratchet**, not a switch: today's violations are **baselined** (not fixed); fixing them is R2. `dev` MUST be green at delivery.
- **A2 (atomicity)** — the gate lands **together with** its baseline and wiring (round-040 TD-1 precedent).
- **A3 (weak E2E carrier)** — because truth/DSL is expected ~NOOP, the **E2E suite is a weak acceptance carrier** here; the witness is **the gate + unit seams**, not "the suite is green" (the round-009 *"green suite = false confidence"* trap, restated in [#92](https://github.com/gosharplite/tellme/issues/92)).
- **A4 (guard form)** — the exact guard form (a build-tagged Go import-graph test vs a Makefile grep vs a script) is an **RD** decision (`/axb-technical-research`), not escalated to clarify; precedent: round-020 A3.
- **A5 (ADR, optional)** — R1 likely needs **no** new ADR (the ADR obligation in [#92](https://github.com/gosharplite/tellme/issues/92) attaches to **R3**'s yield policy); a short layer-model ADR is an RD decision.
- **A6 (no other interfaces)** — `/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted state — the baseline is a **repo artifact**, not runtime state; the `/axb-data-plan` owner ratifies); `/axb-ui-plan` skipped (no TUI surface). Whether `/axb-dsl-refine` is NOOP or adds a dev-observable contract is an RD decision.

## Out of scope (recorded)

- **Fixing** any of the 7 violations — that is **R2** (composition-root extraction).
- Presentation-policy single-ownership (**R3**/**R4**) and the ride-alongs (suggestion-selection policy; `BindToolOutput` ctor injection).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting infrastructure adapters or domain business logic.
- The reference's `modelith-layers` analogue (no domain-model-as-code toolchain here) — recorded, not adopted.
- The topology audit's **stale-semantics** blind spot (a renamed symbol/string leaving a stale row green) — a **recorded forward item** in [#92](https://github.com/gosharplite/tellme/issues/92) (its own round).
