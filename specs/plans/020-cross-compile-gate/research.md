# Phase 0 Research: tellme cross-compile gate — every supported platform compiles (Round 020)

Topic: make **every supported platform compile in the quality pipeline**. `make verify` today compiles only the **host** `GOOS`/`GOARCH`, so build-tagged, OS-specific production code for any other target is invisible to **every** gate. This already bit the project once — round 019's darwin sampler used `syscall.SysctlUint64` (undefined on darwin), a real compile break that slipped through two review folds and was caught only by a manual cross-build. Round 020 adds a **host-independent cross-compile gate** (build + vet per supported target) and wires it into `make verify` + the closeout checklist. It adds **no** dependency and changes **no** CLI behaviour.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` + stdlib `testing`), provider transports (stdlib `net/http`), and the presentation packages were locked in rounds 001–019. The system still has **one CLI end**, adds **no new system end**, **no external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here** (single CLI end; BDD techstack = `godog` on the built binary; E2E + unit strategy). **IN this round**: the gate, its wiring, and its documentation. **OUT**: any change to the `tellme` binary's behaviour, and any unrelated platform fix (round 019's darwin break is already fixed).

---

## Decision 1: A host-independent target matrix (not a single hardcoded cross-target)

- **Decision**: the gate verifies the **full supported POSIX matrix** — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` — **regardless of the host** `GOOS`/`GOARCH`.
- **Rationale**: the blind spot is **host-dependent**. This project is developed on **both** a Linux host and a macOS host (this workspace is **darwin/arm64**): on the Linux host the *darwin* path is invisible; here the *linux* path is invisible. A gate that hardcodes one cross-target (e.g. the round-019 note's `darwin/arm64`) would leave the symmetric hole open. `specs/truth/techstack.md` declares the project **POSIX-only (Linux/macOS)**, which fixes the target set.
- **Alternatives considered**:
  - **Single hardcoded `GOOS=darwin GOARCH=arm64`** (the literal round-019 note) — insufficient; misses Linux on a macOS host and misses arm64/amd64 splits — rejected.
  - **Verify only the host target** — a no-op; `make verify` already does that — rejected.
  - **Add Windows targets** — the project forswears Windows (POSIX-only); a Windows gate would fail on the round-012/015 POSIX-only surfaces — rejected.

## Decision 2: The gate is a `Makefile` target wired into `make verify`

- **Decision**: add a `verify-cross-compile` target to the `Makefile` — a POSIX shell loop over the matrix running `GOOS=<os> GOARCH=<arch> go build ./...` then `go vet ./...`, failing fast and naming the offending target — and add it to the `verify` aggregate (plus `.PHONY` and `help`).
- **Rationale**: the gates live in the `Makefile` (`verify-no-test-sleep` is a pure Makefile check), so the new gate belongs there too. The **build itself is the oracle** — no Go test harness is needed. Unlike `verify-no-network` (which delegates to a Go guard in `tests/e2e` because it has a *scenario step* whose oracle must not drift), the cross-compile gate has **no** parallel scenario step, so there is no drift surface and no reason to add an indirection.
- **Alternatives considered**:
  - **A Go guard test in `tests/e2e` that shells out to `go build`/`go vet` per target** — heavier, requires spawning the Go toolchain from within a test, and buys no drift protection (there is no second definition) — rejected.
  - **A separate `scripts/` shell file** — an extra artifact for a four-line loop already housed by the `Makefile` — rejected.

## Decision 3: Build **and** vet, per target, failing fast

- **Decision**: per target, run `go build ./...` (compile every package) **and** `go vet ./...` (type-check, including test files), stopping at the first failure and naming the target.
- **Rationale**: `go build` catches the round-019 class (a symbol undefined on the target); `go vet` additionally type-checks the packages **and their test files** for the target, so test-only breakage is caught too. The round-019 recommendation explicitly pairs `go build` with `go vet`.
- **Alternatives considered**:
  - **`go build` only** — misses vet-level type errors and cross-target test breakage — rejected.
  - **`go vet` only** — slower than a bare build and does not produce a compiled artifact; build-first is the natural fail-fast order — rejected.

## Decision 4: Wiring and the closeout reference

- **Decision**: the target is a **member of `make verify`**; the session-closeout procedure (`SESSION-CLOSEOUT.md`, Step 2 quality gates) **references** the gate.
- **Rationale**: the round-019 review recommendation named **"the pipeline / closeout checklist"**. Making it a `make verify` member means the standard gate fails on a broken target; referencing it from the closeout keeps it discoverable and run at day close.
- **Alternatives considered**:
  - **A standalone target not in `verify`** — a broken target would pass the standard gate — rejected.
  - **Closeout reference only** — the pipeline must carry it too per the recommendation — rejected.

## Decision 5: No new dependency; deterministic

- **Decision**: the gate uses only the Go toolchain and `make`; it introduces no new dependency and no new network service; it does not use `time.Sleep` nor depend on ambient environment for its verdict. `go.mod` / `go.sum` are unchanged. The gate pins **`CGO_ENABLED=0`** for every target so an ambient `CGO_ENABLED=1` (common in CGO-laden shell/CI environments) cannot make a cross build fail by invoking the host C compiler against target assembly/headers — the gate is hermetic regardless of the shell (PR #46 review TD1).
- **Rationale**: a build loop needs nothing beyond `go`. Like any build, a **cold** module cache may resolve modules once, but the gate adds no new module and no new service.
- **Alternatives considered**:
  - **A cross-compile helper library / a CI action** — needless dependency / out-of-repo coupling — rejected.

## Decision 6: Testing & BDD techstack unchanged (must-ask answers settled)

- **Decision**: no new system end and no change to the BDD techstack or strategy. The three AIxBDD must-ask questions remain answered by the standing `techstack.md`; the round adds no CLI interface truth (`/axb-dsl-refine` is `NOOP`), no API surface (`/axb-api-plan` `NOOP`), and no data model (`/axb-data-plan` `NOOP`).
- **Rationale**: the round is a build-pipeline change; its verification is the gate's own exit code plus a falsifiability witness, not a Gherkin scenario.
- **Alternatives considered**:
  - **Author CLI Gherkin for the gate** — there is no user-facing CLI behaviour to express; forcing Gherkin would be ceremony — rejected.

---

## Residual risks / forward links

- **cgo**: the gate pins **`CGO_ENABLED=0`** (PR #46 review TD1), so an ambient `CGO_ENABLED=1` cannot break a cross build. If a supported target ever requires cgo, the pin must be revisited (a new decision).
- **Cold module cache**: the first build after a clean cache resolves modules like any build; the gate adds no new module and no new network *service* (Decision 5).
- **Runtime**: four builds + four vets add bounded time to `make verify`; acceptable for the coverage gained.
- **Cross-target test breakage**: `go vet ./...` type-checks test files, so a test that fails to compile for a target fails the gate — desirable, but it means a host-only test that references host-only symbols would (correctly) fail the gate; such a test must be build-tagged.
- **No behaviour change**: `stdout`/`stderr` behaviour, the class-phrase vocabulary, and every existing gate's semantics are unchanged; the round touches the build pipeline and a repo doc only.
- **Deferred, still out of scope**: Windows targets, streaming, MCP, memory, write/shell tools, and any unrelated platform change.
