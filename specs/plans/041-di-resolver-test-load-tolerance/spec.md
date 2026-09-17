# Feature Specification: resolver test load tolerance — a test must not hardcode a tight wall-clock budget that is not its subject (round 041)

**Feature Branch**: `041-di-resolver-test-load-tolerance`

**Created**: 2026-09-17

**Status**: Draft — test-only determinism fix. Operator-locked decisions Q1–Q6 (see below; the operator delegated the recommendations). Anchor issue [#87](https://github.com/gosharplite/tellme/issues/87).

**Input**: Operator request: *"proceed to open 041-\* via /axb-specify"* with the round's decisions delegated to the agent's recommendations. Anchor issue [#87](https://github.com/gosharplite/tellme/issues/87) reports that `internal/infrastructure/di`'s bounded `gh`-token-resolver unit tests fail **non-deterministically** during a whole-suite run (`go test ./...` / `make test`) with `signal: killed` at exactly `2.00s`.

**Behaviour intent**: **MODIFY** the resolver's **unit-test harness** so the tests are **load-tolerant** — a saturated host cannot red them — **without** weakening the resolver's boundedness falsifiability. This is a **test-harness determinism** slice (the ADR-036 / round-036 determinism-standard family), **not** a product-behaviour change.

---

## Grounded in the current system

- **Production seam (unchanged).** `internal/infrastructure/di/mcp_factory.go` — `NewGhTokenResolver(bound)` spawns `gh auth token` in its **own process group**, bounded by `bound`; on deadline it SIGKILLs the group (`cmd.Cancel = kill(-pgid, SIGKILL)`) with `ghWaitDelay = 2s`, so `cmd.Output()` returns an `*exec.ExitError` whose text is exactly `signal: killed`.
- **The failing test.** `internal/infrastructure/di/mcp_factory_test.go`:
  - `TestNewGhTokenResolver_TrimsToken` — subject: **token trimming**; the shim is `echo "tok-123"` (**instant**); but it **hardcodes its own `2 * time.Second`** budget — **not** the caller's bound (production's fast-fail bound is **`mcpDiscoveryBound = 3 * time.Second`**, `internal/cli/mcp_discovery.go`; the 2 s literal mirrors the sibling `ghWaitDelay = 2 s`).
  - `TestNewGhTokenResolver_Bounded` — subject: **boundedness**; the shim restores a real PATH then `exec sleep 3`; bound `200 ms`; ceiling assertion `elapsed > time.Second ⇒ fail`.
  - `TestNewGhTokenResolver_MissingGh` — subject: **absent `gh`**; PATH holds no `gh`, so `exec.LookPath` fails instantly and **no child is spawned**; bound `1 s`.
  - `writeFakeGh` sets `PATH` to **only** the shim dir (round-032 implementation-review N1), which forced the hanging shim to restore a real PATH internally.
- **Observed.** Whole-suite runs reddened ~50% on the contended darwin host (3 of 5 runs across three heads, **including the pre-PR base `dev` `802e51c`**); the same package is green **standalone** (0.14–0.84 s). The failure tracks **suite load**, not the diff; the `2.00s` duration is the deadline to the millisecond (a ~14× host-speed factor).

---

## Operator-locked decisions (clarify — delegated; the operator accepted the recommendations)

- **Q1 → widen the positive test's bound (the load-bearing fix).** `TestNewGhTokenResolver_TrimsToken` (subject = trimming) takes a **generous** resolver bound (e.g. `30 * time.Second`); its subject is trimming, so it must never be paced by a tight hardcoded budget that is not its subject.
- **Q2 → dominant PATH (issue option B; hygiene).** `writeFakeGh` sets `PATH = <shim-dir> + ":" + <inherited PATH>`, so the shim still **shadows** `gh` (its dir is first) while its children **resolve** normally — retiring the round-032 N1 in-shim PATH-restoration contortion. This is *not* the load-bearing fix (an instant shim can still blow a tight deadline).
- **Q3 → bounded-test ceiling margin.** `TestNewGhTokenResolver_Bounded`'s wall-clock ceiling moves `> 1s → > 2s` (falsifiability survives: an unbounded resolver measures ≈3 s; the margin over the 200 ms bound grows 5× → 10×).
- **Q4 → witness under contention.** Acceptance = `go test -count=20 ./...` green **under contention** (run concurrently with a full `./tests/e2e` run); **no `t.Skip`, no retry** masking. A **falsifiability witness** is required (revert the bound → the bounded test fails).
- **Q5 → record the rule as ADR 0010.** A new ADR records the generalizable convention — *a test must not hardcode a tight wall-clock budget that is not its subject; real-time assertions must clear a host-speed margin* — authored by `/axb-technical-research`.
- **Q6 → siblings recorded, not swept.** Other wall-clock assertions in the same class are **out of scope**; recorded as a forward item (its own round), not swept here.

**Scope note**: this round changes **no** production code — not `NewGhTokenResolver`, not `ghWaitDelay`, not the MCP factory, the CLI, any `stdout`/`stderr`, any flag/exit code, or any persisted record. It retunes the **test fixture** (`writeFakeGh`) and **two test bounds/ceilings**, and records the rule in truth (ADR + `techstack.md`).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The resolver unit tests are load-tolerant (Priority: P1)

As a maintainer running `go test ./...` (or `make test`) on a loaded host, I want the `di` resolver unit tests to be robust to host contention, so the whole-suite gate never reddens for a reason unrelated to my change.

**Why this priority**: it is the round's entire substance. The flake reddens the standard gate — and the round-040 closeout gate is `go test -count=1 ./...` (`SESSION-CLOSEOUT.md` Step 2), where **Rule 1 forbids closing out on a red gate** — at roughly a coin flip on the contended host.

**Independent verification**: run `go test -count=20 ./...` while a full `./tests/e2e` run saturates the host — the `di` package stays green in every iteration; then revert Q1's bound widening — the positive test reddens (or reproduces the `2.00s` `signal: killed` signature) under the same load.

**Acceptance Scenarios**:

1. **Given** a saturated host (a full `./tests/e2e` run in flight), **When** `go test -count=20 ./...` runs, **Then** the `di` package passes every iteration with no `signal: killed` failure.
2. **Given** the positive trimming test, **When** it runs on a loaded host, **Then** it is **not** paced by a tight hardcoded budget — its resolver deadline is generous (e.g. `30s`) — so an instant shim cannot exceed it.
3. **Given** the bounded test, **When** a correct resolver is measured under load, **Then** it passes (the kill+reap path clears the ceiling margin); **and** **When** the resolver's bound is disabled (an unbounded resolver), **Then** the test still **fails** (falsifiability preserved).
4. **Given** the shim fixture, **When** it is written, **Then** `gh` is shadowed (the shim dir is first on PATH) **and** the shim's own children resolve on a normal PATH — no in-shim PATH restoration is required.

**Functional Requirements**:

- **FR-001**: The positive resolver test (`TestNewGhTokenResolver_TrimsToken`) MUST take a resolver bound that is **decoupled from host speed** — a generous value, not a tight hardcoded literal — so an instant shim cannot exceed it under whole-suite contention.
- **FR-002**: `TestNewGhTokenResolver_Bounded` MUST remain the **sole falsifiability carrier** of the resolver's boundedness: it keeps a tight bound (200 ms), and its assertions MUST still fail for an unbounded/incorrect resolver.
- **FR-003**: The bounded test's wall-clock **ceiling** MUST clear a host-speed margin (raise `> 1s` to a value that is comfortably above the observed kill+reap cost) while preserving falsifiability (an unbounded resolver still exceeds it).
- **FR-004**: The shim fixture (`writeFakeGh`) MUST give the shim a **dominant** PATH — the shim dir first, then the inherited PATH — so `gh` is shadowed while the shim's children resolve normally.
- **FR-005**: The round MUST NOT mask the nondeterminism: no `t.Skip`, no retry loop, no relaxed/vacuous assertion; and it MUST introduce no `time.Sleep` (the determinism gate `verify-no-test-sleep` stays green).

**Non-Functional Requirements**:

- **NFR-001**: **stdlib-only** — no new module dependency; `go.mod` / `go.sum` unchanged.
- **NFR-002**: **Hermetic and host-portable** — no real `gh`, no network; the tests pass both **standalone** and **under load**, on a Linux host and on a macOS host.

---

### User Story 2 - The determinism rule is recorded (ADR 0010 + techstack) (Priority: P2)

As a maintainer / future round author, I want the "a test must not hardcode a tight wall-clock budget that is not its subject" rule recorded in the decision records and the technology-stack truth, so this class of coupling is not repeated.

**Why this priority**: it bounds the fix with a **durable, citable home** (not a frozen plan package — the round-035 session lesson), so a future test does not re-introduce the same host-speed coupling. It is a documentation + governance constraint on Story 1, not an independent capability.

**Independent verification**: read `docs/decisions/0010-*.md`, the `docs/decisions/README.md` index row, and `specs/truth/techstack.md` (Testing & Verification); assert the rule is recorded, the ADR is indexed with a unique number, and every pre-round gate's behaviour is unchanged.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the decision records are read, **Then** ADR 0010 records the test-deadline rule and is listed in the index with a unique number (`adr-index-consistent`).
2. **Given** the technology-stack truth, **When** it is read, **Then** the Testing & Verification section records the resolver harness's load-tolerance (the positive/generic bound split; the dominant-PATH shim).
3. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing gate's behaviour (`verify-no-test-sleep`, `verify-no-network`, `verify-cross-compile`, `vet`, `lint`, `vulncheck`) is unchanged.

**Functional Requirements**:

- **FR-006**: The rule MUST be recorded in a new ADR (`docs/decisions/0010-<slug>.md`) with the `docs/decisions/README.md` index row added (zero-padded, ascending, unique number).
- **FR-007**: `specs/truth/techstack.md` (Testing & Verification) MUST record the resolver harness's load-tolerance change.
- **FR-008**: The round MUST NOT change any production behaviour (the resolver, `ghWaitDelay`, the MCP factory, the CLI) or the semantics of any existing Makefile gate.

**Non-Functional Requirements**:

- **NFR-003**: **POSIX-only** (Linux/macOS); no Windows variant.

---

### Edge Cases

- **The bounded test under extreme load** → the kill+reap path must clear the new ceiling with margin; the *falsifiability* margin (an unbounded resolver measures ≈3 s) caps how much slack the ceiling can carry — pinned in `/axb-technical-research`.
- **The missing-`gh` test** → it spawns **no** child (`exec.LookPath` fails instantly), so it is not load-coupled to its `1 s` bound; whether to widen it for uniformity is a research decision (the default is a **recorded non-change**).
- **`t.Parallel()`** → the `di` tests are **not** parallel; the flake is host-load, not intra-package concurrency. Marking them parallel is out of scope (the Q6 class).
- **`verify-no-test-sleep`** → the shim's shell `sleep 3` is **not** a Go `time.Sleep`; the gate stays green.
- **A host where `/bin/sh` is absent** → out of scope (POSIX-only; the shim relies on `#!/bin/sh`).

### Key Entities *(include if feature involves data)*

- **Resolver test harness** — `writeFakeGh` + the `gh` shim (the test-side fixture; PATH-bound).
- **Bounded resolver** — `NewGhTokenResolver(bound)` (the production seam under test; **unchanged**).
- **Load-tolerance rule** — the recorded convention (ADR 0010) that a test must not hardcode a tight wall-clock budget that is not its subject.

---

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-009**: The round MUST be **test-only + documentation**: it MUST NOT modify any production Go file (including `internal/infrastructure/di/mcp_factory.go`), any `stdout`/`stderr` behaviour, any CLI flag or exit code, or any truth behaviour beyond the `techstack.md` Testing & Verification note and the new ADR.

#### Non-Functional Requirements

- **NFR-004**: The round MUST add **no new third-party dependency** and MUST leave `go.mod` / `go.sum` unchanged.
- **NFR-005**: `make verify` (including `verify-no-test-sleep`, `verify-cross-compile` 4/4, `lint`, `govulncheck`) and the Gherkin/DSL topology audit MUST be green and **unchanged** (no new feature/DSL rows).

### Out of scope (recorded)

- The **sibling wall-clock-assertion class** elsewhere in the suite (Q6) — recorded as a forward item (its own round), not swept here.
- Any **product** change. The `internal/cli` composition-root refactor ([#69](https://github.com/gosharplite/tellme/issues/69)), the dogfooding track ([#60](https://github.com/gosharplite/tellme/issues/60)), and coverage tooling ([#13](https://github.com/gosharplite/tellme/issues/13)) remain their own rounds.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `go test -count=20 ./...` is green **under contention** (concurrently with a full `./tests/e2e` run) — the **discriminating** criterion; a quiet-host run cannot distinguish "fixed" from "quiet". **Fold (measured):** an unqualified `-count=20 ./...` cannot pass on tellme because the `tests/e2e` package re-runs 20× and exceeds Go's **default 10-minute** test timeout (`panic: test timed out after 10m0s`) — the criterion therefore carries an explicit **`-timeout 30m`** (`go test -count=20 -timeout 30m ./...`), and the affected package is additionally witnessed as `go test -count=20 ./internal/infrastructure/di/` under contention. (covers FR-001, FR-004)
- **SC-002**: The positive test no longer hardcodes a tight budget that is not its subject, while the bounded test keeps its tight bound and remains the falsifiability carrier: reverting the positive bound reproduces the flake / red **under load** — reproduced as a falsifiability witness, then reverted. (covers FR-002)
- **SC-003**: The bounded test's ceiling clears the host-speed margin **and** still fails an unbounded resolver — falsifiability preserved in both directions. (covers FR-003)
- **SC-004**: ADR 0010 + the decisions index row + the `specs/truth/techstack.md` (Testing & Verification) note are present; every pre-round gate's behaviour is unchanged; **no production Go file** changed; `go.mod`/`go.sum` unchanged. (covers FR-005–FR-009, NFR-003–NFR-005)

---

## Assumptions

- **A1 (mechanism)**: the failure is the resolver's **own deadline** SIGKILL (confirmed by the exact `2.00s` duration + the `signal: killed` text + the ~14× standalone-vs-loaded factor) — **not** PATH resolution and not an environmental spawn defect (per issue #87's refinement comments).
- **A2 (falsifiability carrier)**: `TestNewGhTokenResolver_Bounded` is and remains the **sole** carrier of the boundedness contract; the positive test's subject is **trimming**, so widening its bound costs no coverage.
- **A3 (dominant PATH preserves shadowing)**: a shim-dir-first PATH still shadows `gh`; the missing-`gh` test uses a dedicated no-`gh` PATH and is unaffected.
- **A4 (serial tests)**: the `di` tests are not parallel; the flake is host-load, not intra-package concurrency.
- **A5 (proposed values)**: the exact generous bound (e.g. `30s`) and the exact ceiling (e.g. `2s`) are the operator-recommended targets; the precise numbers are pinned in `/axb-technical-research` against the observed host-speed factor. [NEEDS CLARIFICATION: none blocking — the numbers are implementation details bounded by the falsifiability constraint (SC-003).]
- **A6 (no CLI interface truth)**: `/axb-spec-by-example` is **NOOP** (no new PM journey); `/axb-api-plan` + `/axb-data-plan` are **NOOP** (no HTTP surface, no persisted state); `/axb-dsl-refine` is **NOOP** (no new/changed CLI interface truth); `/axb-ui-plan` is skipped (plain CLI, no TUI surface).
