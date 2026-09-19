# Feature Specification: tellme domain model + drift gate — `docs/domain-model/` (round 060)

**Feature Branch**: `060-domain-model-and-drift-gate`

**Created**: 2026-09-19

**Status**: Draft — plan half only (spec → clarify → technical-research → system-analysis → tasks → implement). Clarify is **in flight** (see *Pending clarify decisions* below); this spec is refined as answers land.

**Input**: Operator request (verbatim, this session): *"`/Users/…/tellme/docs/domain-model` — I want tellme to have domain model. Does this need a aixbdd round?"* → after grounding (below) the operator said **proceed**. The request is: give `tellme` its **own** canonical domain model under `docs/domain-model/` (the folder exists but is **empty**), in the lineage of the reference's `tell-me-go.modelith.*` / `quality.modelith.*` models.

**Behaviour intent**: **ADD** a **domain model** for `tellme` — a plain-language description of what the system *is* (its entities, relationships, invariants, scenarios) — authored as a `*.modelith.yaml` source and rendered to Markdown (`*.modelith.md`) with embedded Mermaid ER diagrams by **modelith**. Ship it together with a **drift gate** (`modelith-check`: fail if the committed `.md` is stale) so the model cannot silently rot. It is a **docs/tooling slice** in the round-042/043/055 zero-product-code lineage: it changes **no** user-facing behaviour and **no** product code — it adds a model, its toolchain wiring, a gate, and its truth/ADR records.

**Recorded divergence this round REOPENS**: **ADR 0011 D10** records *"tellme adopts the Go-guard form and has **no** modelith toolchain (no `modelith-layers` analogue)"*, and the round-042 forward item (d) repeats *"no `modelith-layers` analogue"*. This round **supersedes that position** (a modelith toolchain is adopted for the model + a drift gate) ⇒ a new **ADR 0030** that amends the ADR 0011 D10 stance, plus a `specs/truth/techstack.md` MODIFY.

---

## Pending clarify decisions (round 1 — asked one at a time; answers fold here)

> The clarify questions are **dependent** (scope → toolchain/gate form → gate strictness), so they are asked **one at a time**. Answers are folded into this spec and the checklist as they arrive.

| # | Decision axis | Status |
| --- | --- | --- |
| **Q1** | **Scope** — which model(s)? (product only · + quality · + environment) | ⏳ pending |
| **Q2** | **Toolchain** — adopt the modelith fork + `make modelith-*` targets (reopens ADR 0011 D10 ⇒ ADR 0030), or hand-authored Markdown (no gate)? | ⏳ pending |
| **Q3** | **Gate strictness + hermeticity** — is `modelith-check` a **zero-tolerance `verify` member**? What happens when the fork binary is absent (round-043 hermeticity; the fork branch is unmerged)? | ⏳ pending |
| **Q4** | **Authority boundary** — the model is **descriptive docs**, not an AIxBDD `TruthArtifact`; rule it *must never contradict* `specs/truth/**` (truth wins)? | ⏳ pending (expect: recorded as a spec assumption) |
| **Q5** | **Lifecycle** — who refreshes the model (each round's truth owner vs. a periodic sweep), given it is not a plan package and has no `delivered` freeze? | ⏳ pending (expect: recorded as a spec assumption) |

---

## Grounded in the current system

- `docs/domain-model/` **exists and is empty**; `git log --all -- docs/domain-model` is **empty** ⇒ `tellme` has **never** had a domain model. (The `modelith lint` / `render --check` lines in the **2026-09-10/11** session notes refer to **`aixbdd-tmg`'s own** model, not tellme's — see `docs/session-summary/2026/09/10/session-summary.md:490` and `docs/archives/status/2026-09-11.md:193`.)
- The reference (`tell-me-go`) ships three models under `docs/domain-model/` (`tell-me-go.modelith.*` product · `quality.modelith.*` process) plus `docs/architect/environments/domain-model/environment-management.modelith.*`, and wires three gates (`modelith-lint`, `modelith-render`, `modelith-check`) — `modelith-check` is a member of its `make check`/`check-full`.
- The **modelith toolchain** is the `gosharplite/modelith` fork (`@feat/self-domain-model`); the binary is installed on this host (`$(go env GOPATH)/bin/modelith`, `v0.0.0-20260815121344-b4153541cee8`). Its `domain-model-author` / `domain-model-context` / `domain-model-lint` skills are **pre-loaded** in this session (Step 5 of `SESSION-BOOTSTRAP.md`), so authorship is available.
- `tellme`'s `Makefile` has **no** `modelith-*` target today; `modelith` is **not** a `go.mod` dependency. The `verify` aggregate is `verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + verify-architecture + lint + vulncheck` (`Makefile:257`).
- **ADR 0011 D10** records the recorded divergence (no modelith toolchain); `specs/truth/techstack.md` (Build & Tooling, *Layer-discipline gate* row) also records it. Adding the toolchain is therefore a real **truth MODIFY** (not a NOOP) + a new ADR.
- The model is **descriptive docs**, distinct from the AIxBDD truth tree (`specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/cli/**`) and from the plan packages. Its relationship to the truth tree is **Q4**.

---

## User Scenarios & Testing *(mandatory)*

> Stories are drafted for the **primary scope** (the product model). Story boundaries are **refined by Q1** (if the quality model and/or an environment model are in scope, Story 1 widens; the gate/truth/ADR stories are unchanged).

### User Story 1 - tellme has a canonical domain model it can reason in (Priority: P1)

As a maintainer (and as the agent working in this repo), I want `tellme`'s own domain model — its entities, relationships, invariants, and scenarios — authored under `docs/domain-model/` in the canonical modelith form, so that the team (and the agent) reason in a shared, precise vocabulary instead of re-deriving "what a `Turn`/`ToolCall`/`Session`/`Provider` is" from the code each time.

**Why this priority**: it is the round's reason to exist. Today there is **no** model — the folder is empty; only the reference repos have models.

**Independent verification**: `modelith lint` over the committed YAML reports 0 errors/0 warnings; the rendered `.md` is up to date; the model names the shipped product entities and their invariants; a spot check confirms the named invariants hold against the code.

**Acceptance Scenarios**:

1. **Given** the committed `docs/domain-model/*.modelith.yaml`, **When** `modelith lint` runs, **Then** it reports **0 errors / 0 warnings**.
2. **Given** the committed rendered `*.modelith.md`, **When** `modelith render --check` runs, **Then** the Markdown is **up to date** (drift = fail).
3. **Given** the model, **When** it is read against the shipped code, **Then** every named entity exists in the code and every printed invariant is **true of the shipped system** (no aspirational entity, no contradicted invariant).

**Functional Requirements**:

- **FR-001**: The round MUST author a **canonical domain model** under `docs/domain-model/` as a `*.modelith.yaml` **source** plus its **rendered** `*.modelith.md` (generated by modelith; never hand-edited). Entity keys PascalCase; backtick entity names in freeform text; `cardinality` ∈ {`1:1`,`1:n`,`n:1`,`n:n`}; `ownership` ∈ {`owned`,`referenced`}.
- **FR-002**: The model MUST describe the **shipped** `tellme` (POSIX/bash-only, **no security layer**, deliberately small tool surface) — it MUST NOT carry entities the project deliberately excludes (`SafePath`, `SecurityManager`, `UserInteractor`, Windows paths, `pipe_commands`).
- **FR-003**: The model MUST follow the 3-pass build order (skeleton → behaviour → refinement) and MUST include, for the product model, at least: `Session`, `Turn`, `Provider`, `Tool`, `ToolCall`, the context/`Context`, `History`, `Skill`, `MCP` server/tool, `Config`, and the **presentation/Chrome** surface — with **invariants** (where the real behaviour lives) and **scenarios** exercising every entity.

**Non-Functional Requirements**:

- **NFR-001**: The model MUST be reviewed (lint-clean) and MUST stay **truth-consistent** (FR-008).

---

### User Story 2 - the model cannot silently go stale (Priority: P2)

As a maintainer, I want the rendered model checked for drift in the standard quality pipeline, so that the committed Markdown cannot diverge from its YAML source without failing the build.

**Why this priority**: an un-gated model rots silently (the reference gates it with `modelith-check`; tellme has **no** such gate today).

**Independent verification**: edit the YAML without re-rendering → the drift gate **fails** and names the stale `.md`; re-render → it passes.

**Acceptance Scenarios**:

1. **Given** the YAML and a stale (or missing) committed `.md`, **When** the drift gate runs, **Then** it exits **non-zero** and names the offending file.
2. **Given** a freshly rendered `.md`, **When** the gate runs, **Then** it exits **0**.
3. **Given** the aggregate quality command, **When** it runs, **Then** it includes the drift gate per **Q3**.

**Functional Requirements**:

- **FR-004**: The round MUST provide a **modelith drift gate** (`modelith-check`) wired per **Q3** (its placement in `make verify` and its absent-binary behaviour are pinned by Q3; the round-043 hermeticity principle applies — a gate MUST NOT become a spurious-red generator).
- **FR-005**: On drift the gate MUST exit non-zero and name the stale artifact; the round MUST reproduce the drift witness (edit-and-not-render ⇒ red) **then revert** (ADR 0010 doctrine).
- **FR-006**: The gate MUST be **deterministic** and MUST add **no** third-party Go dependency (`go.mod`/`go.sum` unchanged); the modelith binary is a **dev toolchain** prerequisite, not a module dependency.

**Non-Functional Requirements**:

- **NFR-002**: **POSIX-only** (Linux/macOS); the round adds **no** new `go.mod`/`go.sum` entry.

---

### User Story 3 - the new toolchain + model are recorded (Priority: P3)

As a maintainer/operator, I want the model and its toolchain recorded in `specs/truth/techstack.md` and governed by a new ADR that supersedes ADR 0011 D10's "no modelith toolchain" position, so the change is discoverable and does not contradict a recorded decision.

**Why this priority**: it bounds the round in truth + governance; the earlier position must not be left standing.

**Independent verification**: read `specs/truth/techstack.md` — the model + toolchain + gate are recorded; **ADR 0030** exists with its index row and names the ADR 0011 D10 amendment; no existing gate's behaviour changes.

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** it records the domain model, the modelith toolchain, and the drift gate.
2. **Given** ADR 0011 D10, **When** the round lands, **Then** **ADR 0030** amends the "no modelith toolchain" position and says so explicitly.
3. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing gate's behaviour is unchanged.

**Functional Requirements**:

- **FR-007**: The model, the modelith toolchain, and the drift gate MUST be recorded in `specs/truth/techstack.md`; a new **ADR 0030** (`docs/decisions/0030-*.md`) with its `docs/decisions/README.md` index row MUST record the adoption and the ADR 0011 D10 amendment.
- **FR-008**: The model MUST be ruled **subordinate to truth**: on any conflict with `specs/truth/**` (techstack, data model, CLI interface features/DSL), the **truth** wins and the model is corrected (Q4).

**Non-Functional Requirements**:

- **NFR-003**: The round MUST NOT change any `tellme` binary behaviour: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**; **no production Go behaviour** changes.

---

### Edge Cases

- **YAML edited, `.md` not re-rendered** → the drift gate **fails** (FR-005); the `.md` is regenerated in the same commit.
- **`.md` hand-edited** → forbidden: the rendered file is generated (`domainmodel-md-generated`); the drift gate reds on the next render.
- **modelith binary absent** on a host running `make verify` → resolved by **Q3** (the round-043 hermeticity principle: the gate MUST NOT be a spurious-red generator).
- **A model claim contradicting `specs/truth/**`** → the **truth** wins; the model is corrected (FR-008).
- **The model naming a deliberately excluded entity** (a security/Windows concept) → forbidden (FR-002).
- **Fork-branch drift** — the modelith fork is unmerged; a toolchain version change could change rendering → recorded as a residual (a forward item), pinned in `research.md`.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only cross-story requirements.

### Global Requirements

#### Functional Requirements

- **FR-009**: The round MUST be **docs + tooling + truth only**: it MUST add the model, its toolchain wiring, the drift gate, the ADR, and the truth/spec records, and MUST NOT modify any production Go behaviour or any existing gate's semantics.
- **FR-010**: The round MUST pin, in `research.md`/`plan.md`: the **scope** (Q1), the **toolchain form** (Q2), the **gate placement + absent-binary policy** (Q3), the **authority boundary** vs `specs/truth/**` (Q4 Q→FR-008), and the **refresh lifecycle** (Q5) — and MUST re-verify the model against the code before freeze. Shape: **ADD** — a model + toolchain + gate + truth/ADR records; it amends one recorded ADR position (0011 D10) and adds one techstack area.

#### Non-Functional Requirements

- **NFR-004**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `verify-architecture`, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows (`/axb-dsl-refine` **NOOP** — the model is a **dev/docs surface**, not the `tellme` CLI contract; the round-020/031/041/042/043 non-BDD-tooling precedent).

### Key Entities *(include if feature involves data)*

- **DomainModel**: `docs/domain-model/*.modelith.yaml` (canonical source) + the rendered `*.modelith.md` (generated; never hand-edited) — a **descriptive** artifact, not an AIxBDD `TruthArtifact` (Q4).
- **Modelith toolchain**: the `gosharplite/modelith` fork binary + the `make modelith-lint|render|check` targets (Q2/Q3).
- **Drift gate**: `modelith-check` — fails when the committed `.md` is stale; placement/absent-binary policy pinned by **Q3**.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `modelith lint` over the committed YAML reports **0 errors / 0 warnings**, and `modelith render --check` reports the committed `.md` **up to date**. (covers FR-001)
- **SC-002**: The model covers the shipped product entities (`Session`/`Turn`/`Provider`/`Tool`/`ToolCall`/context/`History`/`Skill`/`MCP`/`Config`/presentation), each with relationships; the model contains **no** excluded entity (no security layer / no Windows). (covers FR-002, FR-003)
- **SC-003**: A YAML edit without a re-render makes the drift gate **fail**, reproduced as a falsifiability witness then reverted; a fresh render makes it pass. (covers FR-004, FR-005)
- **SC-004**: `specs/truth/techstack.md` records the model + toolchain + gate; **ADR 0030** + its index row amend ADR 0011 D10; every pre-round gate's behaviour is unchanged; **no production Go behaviour** changes; `go.mod`/`go.sum` unchanged; the topology audit is unchanged/green. (covers FR-007, FR-008, FR-009, FR-006, NFR-002, NFR-003, NFR-004)

## Assumptions

- **A1 (intent)** — this is a **docs/tooling** slice; it adds a model + a gate, not product behaviour. `dev` MUST stay green at delivery.
- **A2 (author tooling)** — the model is authored with the pre-loaded `domain-model-author` skill in the 3-pass order (skeleton → behaviour → refinement); *"your value is in the questions, not the typing"*.
- **A3 (weak E2E carrier)** — because truth/DSL is expected **NOOP**, the E2E suite is a **weak acceptance carrier**; the witness is **the model + its lint/render + the drift gate** (the round-009 *"green suite = false confidence"* trap). Per this, **`/axb-spec-by-example` is NOOP** (no user-facing business journey) — precedent: rounds 020/031/041/042/043/055.
- **A4 (ADR required)** — **ADR 0030** records the adoption + the ADR 0011 D10 amendment (a project-level rule later rounds cite).
- **A5 (no other interfaces)** — `/axb-api-plan` **NOOP** (no HTTP surface) · `/axb-data-plan` **NOOP** (the model is a docs artifact, not runtime/persisted state) · `/axb-dsl-refine` **NOOP** (a dev surface, not the CLI contract) · `/axb-ui-plan` skipped (no UI surface).
- **A6 (authority boundary — to be confirmed by Q4)** — the model is **descriptive docs**; on any conflict the truth tree wins (FR-008).
- **A7 (lifecycle — to be confirmed by Q5)** — the model is refreshed alongside a round's truth changes rather than by a separate scheduled pass.

## Out of scope (recorded)

- **Changing any product behaviour**: flags, exit codes, stream contracts, formatting, DSL vocabulary — unchanged.
- Adopting the reference's **`modelith-layers`** architecture gate — the round adds the **model drift gate** only (the Go-guard form of `verify-architecture` is retained; ADR 0011 D10 amended **only** for the toolchain + model, not for the layer gate).
- A `specs/truth/**` restructuring — the model does **not** become a `TruthArtifact` (Q4).
- Fixing any pre-existing model/topology residual.
