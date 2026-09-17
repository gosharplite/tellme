# Feature Specification: layer-discipline gate + violation baseline — the pipeline detects a new illegal layer import (round 042)

**Feature Branch**: `042-layer-discipline-gate-plan`

**Created**: 2026-09-17

**Status**: Draft — plan half only (spec → clarify → spec-by-example → technical-research → system-analysis → dsl-refine). Anchor issue [#93](https://github.com/gosharplite/tellme/issues/93); parent umbrella [#92](https://github.com/gosharplite/tellme/issues/92) **R1**. Clarify round 1 locked Q1/Q2/Q3 (below).

**Input**: Operator request to run **R1 of [#92](https://github.com/gosharplite/tellme/issues/92)** as its own plan package — *"Create a detail new issue just to do R1"* → [#93](https://github.com/gosharplite/tellme/issues/93). R1 is: *add a `verify-architecture` analogue (an import-direction check over `domain → app → infrastructure → cli/ui`) to `make verify`, shipping with a **baseline file** that records today's violations so `dev` stays green.* Origin: round-039 review **TD-3** (PR [#81](https://github.com/gosharplite/tellme/pull/81)) — *the missing layer-discipline gate*.

**Behaviour intent**: **ADD** a **layer-discipline verification gate** to `tellme`'s quality pipeline, plus a **committed baseline** of the layer violations that already exist, so that (a) a *new* illegal layer import fails the standard gate instead of sailing through, and (b) the existing violation count can only be **ratcheted down** to 0 by the later rounds (R2–R4). It is a **tooling/gate slice** in the round-020 `verify-cross-compile` / round-032 `verify-mcp-sdk-confinement` lineage: it changes **no** user-facing behaviour and **no** product code — it adds a gate, its baseline, and its truth record.

---

## Locked decisions (clarify round 1 — plus one resolved by measurement)

- **Q1 → Option 2 — a BROAD import-direction rule (stated in TWO parts).** The gate enforces a general rule over the module's layer ordering (the reference's `verify-architecture` shape), **not** a narrow `cli → infrastructure` prohibition. It is a **two-part predicate** — **(A)** no *upward* import (imported tier > importer tier) **and** **(B)** the application tiers (`internal/app/**`, `internal/cli`) MUST NOT import `internal/infrastructure/**` — plus **(C)** `internal/domain/**` purity and **(D)** default-deny for an unranked `internal/**` package. **Why two parts:** the 7 `cli → infrastructure` edges are **downward** imports that a one-part "no upward import" rule leaves **legal**, so the one-part form cannot produce the 7-row baseline it must; part (B) is the target rule [#93](https://github.com/gosharplite/tellme/issues/93) mechanised ("`internal/cli`: domain + stdlib + application utilities only"). Consequence: `internal/agent → internal/ui` **is a violation** (part A), so the baseline is **8** (see the pinned ranking and the worked examples in **ADR 0011**).
- **Q2 → resolved by measurement (RD-recordable).** Whether the gate governs `_test.go` imports and the `tests/**` harness **does not change the baseline**: measured 2026-09-17, **no** `internal/**` test file imports anything upward, and the only outward harness importer is `tests/e2e` (+ `tests/e2e/steps`), which is test-support. **Scope is therefore pinned as**: the gate governs **production `internal/**`**; `_test.go` imports in `internal/**` are also governed (they add no entries); **`cmd/tellme` (composition top)** and **`tests/**` (test-support)** are **unranked/exempt**. Pinned in `research.md`.
- **Q3 → Option 1 — a STALE baseline entry FAILS the gate.** When a committed baseline entry no longer corresponds to a violation, the gate exits non-zero and names the entry ("remove it from the baseline"). The baseline can only shrink toward truth; each fix *must* edit it (visible in the diff). This is what gives [#92](https://github.com/gosharplite/tellme/issues/92)'s "count decreases → 0" claim teeth (falsifiable).

**Pinned layer ranking (low → high), with the resulting violations:**

| Tier | Packages | Notes |
| --- | --- | --- |
| 0 Pure | `internal/domain/**` | imports only `internal/domain/**` + stdlib |
| 1 Shared utilities | `internal/config`, `internal/home` | bottom shared; `infrastructure → config/home` is **legal** |
| 2 Application | `internal/app/**` | |
| 3 Infrastructure | `internal/infrastructure/**` | |
| 4 Agent | `internal/agent` | |
| 5 Presentation | `internal/ui`, `internal/ui/tui/**` | ranked **above** `agent` ⇒ `agent→ui` is a violation |
| 6 CLI | `internal/cli` | |
| 7 Composition | `cmd/tellme` | top (unranked/exempt from the gate) |
| — | `tests/**` | test-support (exempt) |

**Resulting baseline = 8** — derivable from the **two-part** predicate (B-1): **RULE-B** yields the 7 `internal/cli → internal/infrastructure/{history, llm, skills, telemetry, tools, di, mcp}` edges (downward imports the upward-only rule would leave legal); **RULE-A** yields `internal/agent → internal/ui`; RULE-C/RULE-D yield 0 today. Measured 2026-09-17: 30 packages, **8** violations, **0** unranked governed packages, **0** cycles. Normative rule + worked examples: **ADR 0011**; the gate's own **tier table** is the normative machine-readable ranking source.

> **Ratchet refinement (from Q1)**: [#92](https://github.com/gosharplite/tellme/issues/92) AC1 says the count "decreases as **R2** lands (→ 0)". Under the broad rule the 8th entry (`agent → ui`) is **R3/R4's** workstream, so the ratchet reaches **0 across R2–R4**, not R2 alone. R2 removes the 7 `cli → infrastructure` entries; R3/R4 removes the `agent → ui` entry. Recorded so AC1 is read correctly.

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

- **+ the 8th edge (from the Q1 broad rule)**: `internal/agent → internal/ui` — the agent loop importing the presentation layer. **Baseline = 8.** (Re-measure from the gate's own output before freezing; never hand-transcribe.)

- **`make verify` detects no layer violation today.** The aggregate is `verify-no-test-sleep + verify-no-network + vet + verify-cross-compile + verify-mcp-sdk-confinement + lint + vulncheck` (`Makefile`). None checks import direction, so a ninth illegal import would pass every gate — only a human reviewer would notice.

- **Measured package scope (2026-09-17, whole module)**: no `internal/**` test file imports upward; the outward `tests/e2e` harness imports are test-support. The broad rule's other candidate edges (`infrastructure/* → config`) are **legal** under the pinned tier-1 ranking.

- **Two established gate shapes already exist** to model on: the **Makefile grep** (`verify-mcp-sdk-confinement`) and the **make → `go test -run …`** delegation (`verify-no-network`). The reference `tell-me-go` ships a **build-tagged Go guard** (`make verify-architecture` → `go test -tags=arch -run TestVerifyRealArchitecture ./internal/tools/analysis -args -strict-arch=true` + `modelith-layers`) — `SESSION-BOOTSTRAP.md` §2.2. Guard form is an RD decision (A4).

- **`specs/truth/techstack.md` (Build & Tooling)** already records the gate catalog and the `verify` aggregate list, so adding a gate is a real **truth MODIFY** (not a NOOP).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A new illegal layer import fails the standard gate (Priority: P1)

As a maintainer, I want `make verify` to fail when a change introduces a new illegal cross-layer import (an upward/forward import under the pinned layer ranking), so layer discipline is enforced mechanically instead of relying on a human reviewer noticing.

**Why this priority**: it is the round's entire reason to exist and the precondition for every later round in [#92](https://github.com/gosharplite/tellme/issues/92) — without it, R2's "→ 0" claim is unfalsifiable. Today **no** gate in `make verify` checks import direction.

**Independent verification**: add a *new* illegal import somewhere the gate governs (upward, or an application-tier → `internal/infrastructure/**`), run the aggregate `make verify` — it exits non-zero and names the offending `package -> import`; revert and it exits 0.

**Acceptance Scenarios**:

1. **Given** the current tree, **When** `make verify` runs, **Then** the layer-discipline gate runs, reports **0** violations beyond the baseline, and the aggregate exits 0.
2. **Given** a change that adds a new illegal import under the gate's scope, **When** `make verify` runs, **Then** the gate exits non-zero and its output **names the offending source package and import**.
3. **Given** the aggregate command, **When** it runs, **Then** it includes the layer-discipline gate (a violation fails the standard gate, not a side command).

**Functional Requirements**:

- **FR-001**: The quality pipeline MUST include a **layer-discipline gate** that computes the module's import graph and flags, over the **pinned ranking** (below), every **violation** of the **two-part predicate**: **(A)** an import of a **higher** tier; **(B)** an import of `internal/infrastructure/**` by an application tier (`internal/app/**`, `internal/cli`); **(C)** a `internal/domain/**` package importing outside the domain; **(D)** an `internal/**` package matching **no** tier (*default-deny*). It MUST also assert **0 import cycles** (a direction-only rule cannot detect a same-tier cycle).
- **FR-002**: On any violation **beyond the baseline**, the gate MUST exit non-zero and MUST identify the offending source package and import in a **machine- and human-readable** form (one sorted line per violation: `<source> -> <import>`, module-relative, ASCII delimiter).
- **FR-003**: The gate MUST be a member of the aggregate verification command (`make verify`), so a new violation fails the standard gate.
- **FR-004**: The gate MUST be **hermetic and host-independent**: it MUST evaluate the import graph as the **union over the supported targets** (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`) so an OS-gated illegal import cannot hide, MUST NOT depend on host `GOOS`/`GOARCH`/`CGO_ENABLED`, and MUST NOT use the network with a **warm module cache** (the `verify` aggregate runs it after `vet`/`verify-cross-compile`, which warm the cache). It MUST add no dependency or tool beyond the Go toolchain and `make`.
- **FR-005**: The gate MUST anchor its enumeration to the **module root** (resolve it once; run the child `go list` with `cmd.Dir = <moduleRoot>`), never to the process CWD, and MUST treat a child error or an empty graph as a **failure** (never a vacuous pass).

**Non-Functional Requirements**:

- **NFR-001**: The gate MUST be **deterministic** (a stable, sorted, diff-friendly report) and MUST NOT use `time.Sleep` or depend on wall-clock/timing.
- **NFR-002**: **stdlib-only** — no new module dependency; `go.mod` / `go.sum` unchanged.

---

### User Story 2 - The existing violations are baselined, the count only ratchets down, and a stale entry fails (Priority: P2)

As a maintainer, I want the violations that already exist recorded in a committed baseline so `dev` is green **today** and the gate can land immediately; each later fix (R2–R4) must **remove** a baseline entry (its removal visible in the diff); and a baseline entry left behind after a fix must **fail** the gate, so the baseline can never silently rot.

**Why this priority**: it is what makes the gate *landable now* while the refactor chain (R2–R4) is still pending. Without a baseline the gate reddens `dev` on day one; without the fail-on-stale ratchet the baseline rots.

**Independent verification**: read the baseline — it lists exactly today's violations in the gate's own format; run the gate — green; remove a baseline entry while its violation still exists and re-run — the gate **fails**; fix a violation but leave its entry and re-run — the gate **fails** (stale).

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the gate runs on `dev`, **Then** it is **green** and the baseline lists the current violations (8 under the pinned ranking).
2. **Given** the baseline and an unfixed violation whose entry was removed, **When** the gate runs, **Then** it **fails** (the baseline is not a blanket allow-list).
3. **Given** a fixed violation whose baseline entry remains, **When** the gate runs, **Then** it **fails** and names the stale entry (Q3).
4. **Given** the ratchet, **When** a later round fixes a violation, **Then** exactly one baseline entry is removed and the change is visible in the diff.

**Functional Requirements**:

- **FR-006**: A **baseline file** MUST be committed, listing the currently-known violations in the **gate's own output format** — one **`sort.Strings`-ordered** (byte-wise, never shell `sort`, whose collation is locale-dependent), **ASCII-delimited** (` -> `), **module-relative-package** line per violation — so the gate is green on `dev` at delivery; the baseline is **generated from the gate**, never transcribed.
- **FR-007**: The baseline MUST NOT be a blanket allow-list: a violation **not** present in the baseline MUST fail the gate (FR-002).
- **FR-008**: A **stale** baseline entry (no longer a violation) MUST **fail** the gate and name the entry (Q3 = fail); a fixed violation's entry is removed one-for-one, so the baseline only shrinks.
- **FR-009**: The gate MUST NOT forbid the legitimate, recorded seams (e.g. it MUST NOT flag `agentTools()` — the parameterless, read-free assembler consumed by the round-031 well-formedness gate); the baseline is the mechanism for the existing-but-unfixed edges.

**Non-Functional Requirements**:

- **NFR-003**: The baseline MUST be stable (sorted, line-based) and MUST be generated from the gate's own output, not hand-transcribed (repo lesson: no unverified transcription).

---

### User Story 3 - The new gate is recorded in the technology-stack truth (Priority: P3)

As a maintainer/operator, I want the layer-discipline gate recorded in `specs/truth/techstack.md` (Build & Tooling) and named in the `make verify` aggregate list, so the gate is discoverable and does not become an orphan Makefile target.

**Why this priority**: it bounds the round in truth. It is a documentation/wiring constraint on Stories 1–2, not an independent capability.

**Independent verification**: read `specs/truth/techstack.md` (Build & Tooling) — the gate, the layer ranking, and the baseline policy are recorded, and the `verify` aggregate list includes the new member; assert every pre-round gate's behaviour is unchanged.

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** the Build & Tooling section records the layer-discipline gate, the layer ranking, and the baseline (fail-on-stale) policy.
2. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing gate's behaviour is unchanged and the `verify` aggregate only **gains** a member.

**Functional Requirements**:

- **FR-010**: The gate, the layer ranking, and the baseline policy MUST be recorded in `specs/truth/techstack.md` (Build & Tooling) — the new gate row stating the **predicate** (citing **ADR 0011**) — and the **Task runner** row's `verify` aggregate list MUST name the new member. A new **ADR 0011** (`docs/decisions/0011-layer-discipline-gate.md`) MUST record the rule (what it is / is not) + the worked 8-entry baseline, with its `docs/decisions/README.md` index row.
- **FR-011**: The round MUST NOT change the behaviour of any existing gate or the meaning of the aggregate `verify` beyond adding the new member; it MUST NOT change any `stdout`/`stderr` behaviour of the `tellme` binary, any flag, or any exit code.

**Non-Functional Requirements**:

- **NFR-004**: **POSIX-only** (Linux/macOS); no Windows variant. The round adds **no** new third-party dependency (`go.mod` / `go.sum` unchanged).

---

### Edge Cases

- **Gate lands before its baseline** → forbidden: the gate and its baseline MUST land as **one atomic delivery** (round-040 **TD-1** precedent).
- **Host `GOOS`/`GOARCH` or ambient `CGO_ENABLED`** → the gate's verdict MUST be identical on every host (mirrors `verify-cross-compile`'s hermeticity rationale).
- **A *new* violation while the baseline lists others** → fail on the new one (FR-006).
- **A *fixed* violation still in the baseline (stale)** → **fail**, naming the entry (FR-007, Q3).
- **`_test.go` imports and the `tests/**` harness** → `internal/**` test files are governed (they add no entries); `tests/**` and `cmd/tellme` are exempt (scope pinned in `research.md`).
- **OS/build-tag-gated files** → the gate MUST evaluate the **union over the supported targets** (`CROSS_TARGETS`) so an OS-gated illegal import (`//go:build linux`/`darwin`) cannot hide; **custom build-tag**-gated files (e.g. `//go:build arch`) are a **recorded out-of-scope residual** (none today — `research.md` D6), not a claim that "build tags are respected".
- **Unranked `internal/**` package** → MUST **fail** (default-deny, RULE-D): the gate never treats "unknown" as "allowed".
- **A cycle** (same-tier or any-tier) → the SCC pass **fails** (cycles have no baseline; must be 0).
- **The round's own gate files** → the guard and its self-test MUST NOT themselves introduce a violation under the rule they enforce.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-012**: The round MUST be **tooling + truth only**: it MUST add the gate, its self-tests, the baseline, the ADR, and the truth/spec records, and MUST NOT modify any production Go behaviour or any existing gate's semantics.
- **FR-013**: The round MUST pin the **layer ranking** (Q1, above) and the **package scope** (Q2: production `internal/**`; `cmd/**` + `tests/**` + `tools/**` exempt) explicitly in `research.md`/`plan.md`, and MUST re-measure the baseline from the gate itself before freezing it. Shape: **ADD** — the round adds a gate + baseline + truth/ADR records; it changes no existing truth row's meaning beyond the `verify` aggregate member.

#### Non-Functional Requirements

- **NFR-005**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `verify-mcp-sdk-confinement`, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows (`/axb-dsl-refine` NOOP — A6).

### Key Entities *(include if feature involves data)*

- **Layer ranking**: the pinned tier order (`domain` → `config`/`home` → `app` → `infrastructure` → `agent` → `ui` → `cli`; `cmd`/`tests`/`tools` exempt) that defines a legal import direction. The gate's **tier table** is the normative machine-readable source.
- **Layer violation**: a governed package's import that breaks the **two-part predicate** (upward import; application-tier → `internal/infrastructure/**`; domain-purity breach; unranked governed package) — reported as `source -> import`.
- **Baseline**: the committed, sorted file listing the currently-known violations in the gate's format; a **ratchet** that only shrinks; a stale entry is a **failure**.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make verify` runs the layer-discipline gate; on `dev` it is **green** with the baseline listing the current **8** violations (the count is re-measured from the gate at freeze, and is **derivable from the two-part predicate**). (covers FR-001, FR-003, FR-006)
- **SC-002**: A deliberately added **new** illegal import — upward (RULE-A) **or** an application-tier → `internal/infrastructure/**` (RULE-B) — makes the gate **fail** and name the offending `package -> import` — reproduced as a falsifiability witness, then reverted (ADR 0010 doctrine). (covers FR-002, FR-007)
- **SC-003**: The baseline is a self-policing **ratchet**: removing an entry for an unfixed violation fails the gate, **and** a stale entry (a fixed violation left in the baseline) **fails** the gate and names the entry. (covers FR-007, FR-008)
- **SC-004**: The gate's verdict is identical across hosts — it is the **union over `CROSS_TARGETS`**, so an OS-gated illegal import is caught — is deterministic/sorted, does not depend on host `GOOS`/`GOARCH`/`CGO_ENABLED`, asserts **0 cycles**, and adds no dependency (`go.mod`/`go.sum` unchanged). (covers FR-004, FR-001, NFR-001, NFR-002, NFR-004)
- **SC-005**: `specs/truth/techstack.md` (Build & Tooling) records the gate + predicate + baseline policy and the `verify` aggregate member; **ADR 0011** + its index row are present; every pre-round gate's behaviour is unchanged; **no production Go behaviour** changes; the Gherkin/DSL topology audit is unchanged/green. (covers FR-010, FR-011, FR-012, NFR-005)

## Assumptions

- **A1 (intent)** — R1 is a **ratchet**, not a switch: today's violations are **baselined** (not fixed); fixing them is R2–R4. `dev` MUST be green at delivery.
- **A2 (atomicity)** — the gate lands **together with** its baseline and wiring (round-040 TD-1 precedent).
- **A3 (weak E2E carrier)** — because truth/DSL is expected ~NOOP, the **E2E suite is a weak acceptance carrier** here; the witness is **the gate + its self-tests**, not "the suite is green" (the round-009 *"green suite = false confidence"* trap, restated in [#92](https://github.com/gosharplite/tellme/issues/92)). Per this, **`/axb-spec-by-example` is NOOP** (no user-facing business journey) — precedent: rounds 020/031/036/041.
- **A4 (guard form)** — the guard is a **build-tagged Go guard test** (`//go:build arch`) that shells stdlib `go list` from the module root; the exact file layout is an **RD** decision (`research.md` D2/D4), not an FR.
- **A5 (ADR required)** — **ADR 0011** records the rule + baseline policy (per `docs/decisions/README.md`: a project-level rule other rounds must cite — R2–R4 will). *(This **replaces** the earlier "no ADR" position — review RF-1.)*
- **A6 (no other interfaces)** — `/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted state — the baseline is a **repo artifact**, not runtime state; the `/axb-data-plan` owner ratifies); `/axb-dsl-refine` = **NOOP** (the gate is a **dev surface**, not the `tellme` CLI contract); `/axb-ui-plan` skipped (no TUI surface).

## Out of scope (recorded)

- **Fixing** any baselined violation — that is **R2** (the 7 `cli → infrastructure` edges) and **R3/R4** (the `agent → ui` edge).
- Presentation-policy single-ownership (**R3**/**R4**) and the ride-alongs (suggestion-selection policy; `BindToolOutput` ctor injection).
- Any **user-facing** change: flags, exit codes, stream contracts, formatting, DSL vocabulary — **unchanged**.
- Rewriting infrastructure adapters or domain business logic.
- The reference's `modelith-layers` analogue (no domain-model-as-code toolchain here) — recorded, not adopted.
- The topology audit's **stale-semantics** blind spot (a renamed symbol/string leaving a stale row green) — a **recorded forward item** in [#92](https://github.com/gosharplite/tellme/issues/92) (its own round).
