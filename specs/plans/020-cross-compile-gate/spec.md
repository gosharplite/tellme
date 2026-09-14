# Feature Specification: tellme cross-compile gate — every supported platform compiles (round 020)

**Feature Branch**: `020-cross-compile-gate`

**Created**: 2026-09-14

**Status**: Draft — the target matrix and pipeline wiring are grounded in existing truth; no clarify round proposed

**Input**: Operator request, verbatim: **"Let's just do B now."** — where **B** is the round-019 review forward recommendation and the `STATUS.md` "cross-compile gate" candidate:

> *"Add a `GOOS=darwin GOARCH=arm64 go build ./...` (+ `go vet`) cross-compile gate to the pipeline / closeout checklist — build-tagged platform code is invisible to `make verify`."*

Behaviour intent: **ADD** a **host-independent cross-compile verification gate** to `tellme`'s quality pipeline, so build-tagged, OS-specific production code is compiled (and type-checked) for **every supported target** — not only the host's — closing the blind spot that let the round-019 darwin sampler ship uncompiled.

**Grounded in the current system**:
- `make verify` aggregates only `verify-no-test-sleep`, `verify-no-network`, `vet`, `lint`, `vulncheck` — **none** compiles a non-host `GOOS`/`GOARCH` (`Makefile`).
- Round 019 ships **build-tagged** production code: `internal/infrastructure/telemetry/system_metrics_linux.go` (`//go:build linux`) and `system_metrics_darwin.go` (`//go:build darwin`).
- The pipeline blind spot is **host-dependent**: this workspace runs on **darwin/arm64**, so the *Linux* path is invisible to `make verify` here; on the Linux dev host the reverse holds. The gate must therefore be **host-independent**, not a single hardcoded cross-target.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Every supported platform compiles in the pipeline (Priority: P1)

As a maintainer, I want the quality pipeline to compile and vet the module for **every supported OS/arch combination** — not just the one I happen to be running on — so that platform-tagged code can never silently fail to compile (or type-check) for a target the project ships.

**Why this priority**: it is the round's entire reason to exist. A platform that cannot be built is broken for its users, and today such a break escapes **every** gate because they only ever build the host `GOOS`/`GOARCH`. This already happened once (round 019's darwin sampler used `syscall.SysctlUint64`, undefined on darwin) and was caught only by a manual cross-build.

**Independent verification**: run the gate on the current tree — it builds and vets every supported target and exits 0; then introduce a deliberate compile error into a **non-host** target's build-tagged file — the gate exits non-zero and names that target; revert.

**Acceptance Scenarios**:

1. **Given** the current tree, **When** the cross-compile gate runs, **Then** the module compiles and vets successfully for every supported target and the gate exits 0.
2. **Given** a supported target whose build-tagged code does not compile (or fails `go vet`), **When** the gate runs, **Then** the gate exits non-zero and identifies the failing target.
3. **Given** the aggregate verification command, **When** it runs, **Then** it includes the cross-compile gate.

**Functional Requirements**:

- **FR-001**: The quality pipeline MUST include a cross-compile gate that, for every supported target, (a) compiles the module (`go build ./...`) and (b) type-checks it (`go vet ./...`).
- **FR-002**: The **supported targets** MUST be the POSIX set the project ships — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` — and the gate MUST be **host-independent**: it verifies targets other than the host's own, and behaves correctly on either a Linux or a macOS host.
- **FR-003**: When any supported target fails to compile or vet, the gate MUST fail with a non-zero exit status and MUST identify the offending target in its output.
- **FR-004**: The gate MUST be a member of the aggregate verification command (`make verify`), so a broken target fails the standard gate rather than a side command.

**Non-Functional Requirements**:

- **NFR-001**: The gate MUST be deterministic and MUST add no dependency or tool beyond the Go toolchain and `make`; it uses the same module set as a normal build, introduces no new network service, and pins `CGO_ENABLED=0` so cross builds are hermetic regardless of the shell environment.
- **NFR-002**: The gate's added runtime MUST be bounded (a fixed, small number of target builds) and it MUST NOT use `time.Sleep` or depend on ambient environment for its verdict.

---

### User Story 2 - The new gate is recorded in truth and runs at closeout (Priority: P2)

As a maintainer/operator, I want the cross-compile gate recorded in the technology-stack truth and referenced by the session-closeout procedure, so the gate is discoverable and actually runs when the day closes — not an orphan Makefile target.

**Why this priority**: it bounds the round-019 forward item ("pipeline **/ closeout checklist**"). It is a documentation + wiring constraint on Story 1, not an independent capability.

**Independent verification**: read `specs/truth/techstack.md` (Build & Tooling) and `SESSION-CLOSEOUT.md`; assert the gate is recorded and the closeout requires it; assert every pre-round gate's semantics are unchanged.

**Acceptance Scenarios**:

1. **Given** the technology-stack truth, **When** the round lands, **Then** the Build & Tooling section records the cross-compile gate and the supported target matrix.
2. **Given** the closeout procedure, **When** a session is closed out, **Then** the procedure runs/references the cross-compile gate.
3. **Given** the pre-round gate catalog, **When** the round lands, **Then** every existing gate's behaviour is unchanged.

**Functional Requirements**:

- **FR-005**: The cross-compile gate MUST be recorded in `specs/truth/techstack.md` (Build & Tooling), including the supported target matrix.
- **FR-006**: The session-closeout procedure (`SESSION-CLOSEOUT.md`) MUST reference the cross-compile gate, so it runs at day close.
- **FR-007**: The round MUST NOT change the behaviour of any existing gate (`verify-no-test-sleep`, `verify-no-network`, `vet`, `lint`, `vulncheck`) or the meaning of the aggregate `verify` beyond adding the new member.

**Non-Functional Requirements**:

- **NFR-003**: The behaviour MUST be POSIX-only (Linux/macOS); no Windows target is introduced.

---

### Edge Cases

- **Host equals one of the supported targets** → the gate still verifies the other targets; it MAY also re-verify the host's own (harmless).
- **A target whose required toolchain pieces are unavailable** (e.g. a missing cross stdlib) → the gate MUST fail loudly with the target named, never pass silently.
- **Test-only breakage under a cross target** → `go vet ./...` type-checks test files, so a non-compiling test for a target fails the gate (desired).
- **A future target that needs cgo** → out of scope; today all supported targets build pure-Go (a cgo requirement would be a new decision).
- **The round-019 darwin break repeated** → on a Linux host, the gate catches the darwin path that `make verify` alone would miss (and vice versa).

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-008**: The round MUST add the cross-compile gate and its wiring/documentation only; it MUST NOT fix unrelated platform defects (the round-019 darwin break was already fixed in-round) and MUST NOT alter any `stdout`/`stderr` behaviour of the `tellme` binary.

#### Non-Functional Requirements

- **NFR-004**: The round MUST add **no new third-party dependency** and MUST leave `go.mod` / `go.sum` unchanged.

### Key Entities *(include if feature involves data)*

- **Target matrix**: the ordered set of supported `GOOS/GOARCH` pairs the gate verifies (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`).
- **Cross-compile gate**: a quality-pipeline step that builds and vets the module once per target and fails on the first broken target.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The cross-compile gate passes on the current tree for all four supported targets (covers FR-001, FR-002, NFR-001).
- **SC-002**: A deliberately broken **non-host** build-tagged file makes the gate fail and name the target — reproduced as a falsifiability witness, then reverted (covers FR-003).
- **SC-003**: The aggregate `make verify` includes the gate and fails when a supported target is broken (covers FR-004).
- **SC-004**: `specs/truth/techstack.md` and `SESSION-CLOSEOUT.md` record/reference the gate, and every pre-round gate's behaviour is unchanged (covers FR-005–FR-008, NFR-003, NFR-004).

## Assumptions

- **A1 (targets)**: the supported set is the POSIX matrix the project ships — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` — per `specs/truth/techstack.md` ("POSIX-only (Linux/macOS)"); **no Windows** target.
- **A2 (host-independence)**: the project is developed on **both** a Linux host and a macOS (darwin/arm64) host; the gate must catch a break on either, so it verifies the full matrix regardless of host.
- **A3 (mechanism)**: the gate is a pipeline step wired into `make verify`; the exact mechanism — a `Makefile` recipe versus a Go guard test the target delegates to (mirroring `verify-no-network`) — is an RD (`/axb-technical-research`) decision.
- **A4 (vet scope)**: `go vet ./...` type-checks test files as well, so cross-target test breakage is caught.
- **A5 (scope-bound)**: no CLI interface behaviour changes; `/axb-dsl-refine` is expected `NOOP`, and `/axb-api-plan` / `/axb-data-plan` are `NOOP` (no API surface, no persisted/in-memory state).
