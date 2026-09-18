# Feature Specification: E2E suite throughput — parallel scenarios + a fast inner-loop subset (round 055)

**Feature Branch**: `055-e2e-suite-throughput`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify round 1, asked one question at a time**: **Q1 LOCKED → Option B** (parallelism **on by default** at a small fixed level + an override seam; the timing-sensitive scenarios protected per **ADR-0010** rather than pinned serial); **Q2 LOCKED → Option A** (`make test-fast` over `godog.paths`, banner + never-the-gate guard); **Q3 OPEN**. `NEEDS CLARIFICATION` markers below are **non-blocking for the plan shape** but **blocking for the acceptance/implementation decisions** they name.

**Input (operator, 2026-09-19)**: *"Many time I see you do `full test` and takes more than 60 sec. Why so long?"* … *"Can these two be included in 055 together?"* → **Start 055 with both** (both = **US1** parallel scenario execution and **US2** the fast inner-loop subset).

**Behaviour intent**: **MODIFY (developer-facing, test tooling)** — the round changes **how the executable contract is executed**, never **what** it asserts. No product code, no acceptance/DSL semantics, no user-visible behaviour. Measured baseline (this host, darwin/arm64, Go 1.26.6): `go test -count=1 ./...` ≈ **54 s**, of which `tests/e2e` ≈ **50 s** (`TestFeatures` ≈ 49.5 s; 240 Examples / 1746 steps / 46 feature files; ≈205 ms per scenario).

---

## Grounded in the current system

| Site | Current shape (measured 2026-09-19) |
| --- | --- |
| `tests/e2e/suite_test.go` | one `godog.TestSuite` with `Options{Format:"pretty", Paths:["../../specs/truth/features/cli"], TestingT:t, Strict:true}` — **no `Concurrency`** ⇒ scenarios run **serially** |
| `github.com/cucumber/godog` **v0.16.0** (pinned in `go.mod`) | `Options` fields: `Format`, `Tags`, `Concurrency`, `ShowStepDefinitions`, `StopOnFailure`, `Strict`, `NoColors`, `Randomize`, `Paths`, `Output`. Flags: `godog.paths`, `godog.tags`, `godog.concurrency` (`-c`), `godog.format`, `godog.strict`, … **There is no name-filter flag in this version.** `Concurrency < 1` is clamped to `1` (`run.go`). `Tags` is the only scenario-selector; there are **zero Gherkin tags** in the tree today. |
| Scenario isolation | state is per-scenario (`context.WithValue(ctx, scenarioKey{}, sc)`); each scenario gets a fresh temp **`TELL_ME_HOME`** *and* a fresh temp **`HOME`** (round 026) ⇒ the user-global `~/.tellme/` stores cannot race |
| Shared state in `tests/e2e/steps` | only immutable package-level values (regexps, a tool-name slice); `registrars` is appended at **init** time; the binary is built **once** per process (`sync.Once`, temp-dir output) ⇒ parallel-safe |
| Timing-sensitive scenarios | round-040 idle-gap (child `sleep 2`, forced 50 ms), **4×** `execute_command "timeout":1`, `pollProcessGone(…, 2*time.Second)`, harness `markerDeadline = 10s`, `pipedRunTimeout = 60s` |
| Gates that must survive | `Strict: true` (round-040 TD-1: undefined steps FAIL), `-count=1` at the phase gate (load-bearing — the arch gate's `go list` cannot be cached), `verify-no-test-sleep` (no `time.Sleep` in tests), ADR-036 determinism parity, ADR-0010 (a test must not hardcode a tight wall-clock budget that is not its subject) |
| Task runner | `Makefile` `test: verify-mcp-sdk-confinement` + `go test ./...` (cacheable); `verify` aggregates the static gates. **No subset target exists.** |

**Precedent (recorded):** the round-041 `di` flake (issue #87) is the exact failure class parallelism can aggravate — a subprocess under whole-suite resource pressure exceeded a wall-clock budget and was `signal: killed`. ADR-0010 is the durable response to that class.

---

## Clarify round 1 *(asked one question at a time; ≤3 questions; ≤5 per session)*

| # | Question | Status |
| --- | --- | --- |
| **Q1** | **What is the parallelism default, and how do timing-sensitive scenarios stay green?** | ✅ **LOCKED → Option B** (operator, 2026-09-19): `Concurrency` is **on by default** at a small fixed level (`e2eDefaultConcurrency = 4`, clamped to ≥1) with an override seam `TELL_ME_E2E_CONCURRENCY`; the timing-sensitive scenarios are protected by **generous host-speed margins + shape-based non-vacuity (ADR-0010)**, **not** by serial pinning or a static timing-feature list. |
| **Q2** | **Is the fast subset a *gate-excluded* convenience, and what selects it?** | ✅ **LOCKED → Option A** (operator, 2026-09-19): a `Makefile` target `test-fast` selects a subset via **`godog.paths`** (no truth edits); it prints a loud `SUBSET — NOT THE GATE` banner and **refuses to run** if the selected path resolves to the full `features/cli` root. **Micro-decision (vetoable):** the default subset is the **non-`chat` modules** (`configuration`+`history`+`workspace`+`usage`+`diagnostics` = 43 Examples, ≈10 s measured), because `chat` alone is 197/240 Examples (≈40 s) and is therefore *not* a fast subset; `MODULES=chat` (or any comma list) is available. The real chat-inner-loop fix is the **tag** selector → a recorded forward item (Option C, `/axb-dsl-refine`). |
| **Q3** | **What is the *measurable* success bar?** — e.g. "the full gate is ≤ X s on the reference host with all 240 Examples still executed" and "N consecutive green runs to call it stable". | ⏳ **OPEN** — [NEEDS CLARIFICATION: the target and the stability evidence] |

**Non-negotiable invariant (proposed, not open):** *a subset may select for convenience but MUST NEVER exclude from the gate — the phase gate always runs **all** of `Paths` (all 240 Examples).* This exists to prevent the round-040 TD-1 failure mode (scenarios silently dropping while the suite still exits 0).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the gate runs the scenarios in parallel (Priority: P1)

As an **RD engineer**, I want the E2E suite to execute independent scenarios concurrently, so the phase gate's wall-clock drops without losing a single assertion.

**Why this priority**: it is the only option that shortens the **gate**; US2 only shortens the inner loop.

**Independent verification**: with the parallelism setting enabled, a full suite run reports **the same 240 Examples** and exits 0; a serial (default/opt-off) run is unchanged; the wall-clock is measurably lower.

**Acceptance Scenarios**:

1. **Given** the E2E suite, **When** it is run with parallelism enabled, **Then** it still executes **every** Example under `specs/truth/features/cli` (no skips; `Strict` still on) and exits 0.
2. **Given** a timing-sensitive scenario (a `[Tool Output]` idle gap; a tool timeout), **When** the suite runs with parallelism enabled, **Then** it still passes **non-vacuously** (its witness power is unchanged — ADR-0010).
3. **Given** parallelism is off (default/opt-out), **When** the suite runs, **Then** behaviour is **byte-for-byte as today** (a serial run).
4. **Given** N consecutive parallel full runs, **When** they complete, **Then** all are green (the stability evidence for Q3).

**Functional Requirements**:

- **FR-001**: The suite MUST execute scenarios concurrently **by default** — godog `Concurrency = 4` (a small fixed default, clamped to ≥1), overridable by the `TELL_ME_E2E_CONCURRENCY` env seam (Q1 → B).
- **FR-002**: Every Example under `Paths` MUST still be executed on a gate run — parallelism MUST NOT drop, filter, or skip any scenario.
- **FR-003**: `Strict: true` MUST remain in force; an undefined/ambiguous step MUST still fail the suite.
- **FR-004**: Timing-sensitive scenarios MUST remain **non-vacuous** under parallelism (Q1 → B): a test whose subject is *not* the bound takes a **generous test-local margin** above measured host speed, and a bound-firing test asserts **non-vacuity by the failure's shape** (`exec.ExitError` with `ExitCode() == -1`) per **ADR-0010**; **no** scenario is pinned serial and **no** static timing-feature list is introduced.

---

### User Story 2 - a fast inner-loop subset (Priority: P2)

As an **RD engineer**, I want a convenience target that runs a **subset** of the contract quickly while iterating, so the edit→test loop is short — without any risk that the subset becomes the gate.

**Why this priority**: pure developer ergonomics; it saves nothing at the gate, and part of its value is redundant if US1 already makes the full suite fast.

**Independent verification**: the target runs a strict subset (fewer than 240 Examples) and exits 0; the **gate** target still runs all 240.

**Acceptance Scenarios**:

1. **Given** the fast target, **When** it runs, **Then** it executes a **strict subset** of the contract and exits 0.
2. **Given** the gate target/command, **When** it runs, **Then** it still executes **all** 240 Examples (the subset never becomes the gate).
3. **Given** the fast target's output, **When** read, **Then** it is visibly identified as a subset (so a green subset is never mistaken for a green gate).

**Functional Requirements**:

- **FR-005**: The subset MUST be selected **without editing `specs/truth/**`** — via a **`Makefile` target `test-fast`** that passes a `godog.paths` selection into the suite (Q2 → A). The target MUST print a `SUBSET — NOT THE GATE` banner naming the selected path(s) and **MUST refuse to run** when the selection resolves to the full `features/cli` root. The default selection is the **non-`chat` modules** (43 Examples, ≈10 s); `MODULES=<comma list>` overrides it.
- **FR-006**: The subset MUST be strictly additive: the gate command (`go test -count=1 ./...`) MUST keep executing every Example (all 240); no `Makefile`/CI path may substitute the subset for the gate.

---

## Edge Cases

- **Timing observers under load** — the idle-gap / timeout / `pollProcessGone` scenarios are the flake frontier (the #87 class); their protection is Q1's call.
- **A single shared binary** — the one `sync.Once` build is process-wide and read-only ⇒ safe under concurrency; the temp build dir is unique per process.
- **Per-scenario resources** — each scenario's temp `TELL_ME_HOME`/`HOME` and its fake-provider port are per-scenario ⇒ no cross-talk; the harness must not introduce a shared mutable sink.
- **Fail-fast interaction** — `StopOnFailure` stays **off** (an off-by-default field); parallelism must not change failure reporting.
- **`-count=1`** — stays the gate's invocation (disables the test cache; required by the arch gate's whole-module `go list`); the subset target should NOT weaken it for the gate.
- **Concurrency on a 1-core host / CI** — the level must degrade sanely (godog clamps `< 1` to 1); the default must not assume a large machine.
- **`Randomize`** — remains off; ordering entropy would muddy reproducibility.

## Key Entities

- **Suite options** — the single `godog.Options` in `tests/e2e/suite_test.go` (gains `Concurrency` + its override seam).
- **Timing-sensitive scenario set** — the named group whose witness power parallelism could dilute.
- **The fast subset target** — a `Makefile` convenience target (not part of `verify`/the gate).
- **The gate command** — unchanged in scope: all Examples, `Strict`, `-count=1`.

## Success Criteria

- **SC-001**: With parallelism enabled, the suite executes **all 240 Examples** and exits 0; `Strict` still fails an undefined step.
- **SC-002**: The full gate's wall-clock is **measurably lower** than the ≈54 s baseline — the exact bar is **Q3**.
- **SC-003**: N consecutive parallel full runs are green (**N = Q3**), including the timing-sensitive scenarios, which stay non-vacuous.
- **SC-004**: The fast subset target runs a strict subset and exits 0, while the gate target still runs all 240 (the FR-002/FR-006 invariant is *demonstrable*, not just stated).
- **SC-005**: The round's own gates hold: `gofmt`/`go vet` clean · `make verify` **OK** · `verify-no-test-sleep` green · `go.mod`/`go.sum` unchanged.

## Assumptions

- **A1**: This is a **tooling-only** round: no product code, no acceptance/DSL change ⇒ `/axb-spec-by-example` **NOOP**, `/axb-system-analysis` **0 interfaces** (api/data/dsl-refine **NOOP**) — the rounds-042/043/047 precedent.
- **A2**: Truth impact is limited to `specs/truth/techstack.md` rows (*E2E runner / step definitions*, *Host test harness*, *Test strategy*, *Task runner*), recorded via `truth-delta.md`.
- **A3**: An **ADR (0024)** records the parallelism policy + the timing-protection decision + the subset/gate invariant.
- **A4**: The measured baseline is host-specific; the SC uses a **ratio or a generous ceiling**, never a tight absolute (ADR-0010).
- **A5**: Scenario isolation holds as measured (per-scenario `TELL_ME_HOME`+`HOME`; init-only shared state); a scenario found not to be isolated becomes **part of Q1's protection scope**, not a silent exclusion.

## Out of scope (recorded forward items)

- The **tag-based** subset (would require editing `specs/truth/features/**` ⇒ `/axb-dsl-refine`, its own round).
- A **pty-capable harness** (still a named pin, unchanged).
- Reducing the deliberate waits (`sleep 2`, `"timeout":1`) — they are load-bearing witnesses; trading them for seconds is not this round.
- The locked exclusions: no security layer · no Windows · sequential tool calls (#47 `not_planned`).
