# Round 096 — `096-dependency-vuln-hygiene`

**Theme**: dependency / toolchain **vulnerability hygiene** for the new `kiosk` environment (GKE
container managed by ConfigHub). Two halves in one PR — (1) **bump**: raise the shipped dependency
set past the known `golang.org/x/net` vulnerabilities so the pipeline's `vulncheck` gate no longer
reports a *called* vulnerability; (2) **record**: make the gate's **patched-toolchain precondition**
discoverable on a live surface, so a fresh environment knows the minimum Go toolchain rather than
discovering it only via a red gate. **No product behaviour change.**

**Branch base**: off `dev` `99f1114` (session-87 closeout head).

**Input (operator directive, 2026-10-11)**: *"upgrade the kiosk Go toolchain to ≥ 1.26.9 and bump
`golang.org/x/net` to ≥ v0.60.0"*; the environment/tooling side (installing the `deadcode` tool, the
user-scoped Go 1.26.9) is **environment provisioning**, not a repo change.

**Anchor issue**: none — the theme is an **operator-value** round (Bootstrap Agent Rule 11).

---

## 1. Why this round

The repo's `make verify` includes a zero-tolerance **`vulncheck`** gate (`govulncheck`, a
`QualityGate` of kind `security`). On the new `kiosk` environment (Go **1.26.6**, a fresh checkout of
`dev` `99f1114`) the gate is **RED**:

```
go vulncheck: Your code is affected by 9 vulnerabilities from the Go standard library.
This scan also found 9 vulnerabilities in packages you import and 7 vulnerabilities in modules
you require, but your code doesn't appear to call these vulnerabilities.
```

Two axes are in play, and only **one** is a repo artifact:

- **stdlib** — the fixes are in **`go1.26.9`** (e.g. `GO-2026-6617` HPACK race, `GO-2026-6613`,
  `GO-2026-6605` — all `net/http@go1.26.9`). This is the **ambient Go toolchain** — an
  *environment* input, not a `go.mod` line.
- **module** — `golang.org/x/net@v0.38.0` (indirect) is flagged (`GO-2026-6617/6612/6611/6610/6603`),
  fixed in **`v0.60.0`**. This **is** a repo artifact (`go.mod`/`go.sum`).

So the repo-side fix alone does not make the gate green on an unpatched host, and a patched host
alone does not clear the module finding. This round **bumps the module** (the repo half) and
**records the toolchain floor** (the discoverability half), and verifies the gate returns to
**green** once both hold.

**Measured after the fix** (this environment, 2026-10-11): Go **1.26.9** (user-scoped) +
`golang.org/x/net@v0.60.0` ⇒ `govulncheck` → **"No vulnerabilities found."**; `make verify` **OK**;
`make check-full` **exit 0**. See §5 / `research.md` for the reproduced evidence.

## 2. The change (surfaces)

1. **`go.mod` / `go.sum`** — bump the module graph to clear the flagged module:
   `golang.org/x/net v0.38.0 → v0.60.0` plus the transitive rise it requires
   (`golang.org/x/sys v0.41.0 → v0.48.0` · `golang.org/x/term v0.36.0 → v0.46.0` ·
   `golang.org/x/sync v0.21.0 → v0.23.0` · `golang.org/x/text v0.39.0 → v0.42.0`), via
   `go get golang.org/x/net@v0.60.0` + `go mod tidy`. `go mod tidy` normalises the `go` directive
   `1.26 → 1.26.0` (semantically identical).
2. **`specs/truth/techstack.md`** — record the toolchain **floor** the `vulncheck` gate requires
   (a patched stdlib, ≥ **1.26.9**) so the gate's environmental precondition is a discoverable truth
   clause rather than an implicit host fact. (Exact home/placement = RD-owned, §5; the *Language*
   row or a *Build & Tooling* note both satisfy the intent.)
3. **Records** — the round closes with the standard closeout artifacts (day log / `STATUS.md`
   environment note). Whether a durable **ADR** is owed for "the `vulncheck` gate is
   toolchain-sensitive" is an **RD research decision** (§5, A3) — not assumed here.

**Environment-only (NOT a repo change, out of scope)**: the user-scoped Go **1.26.9** provisioning
and the `deadcode` dev-tool install (`go install golang.org/x/tools/cmd/deadcode@v0.47.0`) are
environment steps; the `deadcode` tool is already governed by **ADR 0064**.

**Not modelled** — a dependency *version* and an *ambient* toolchain are neither entities nor
invariants; the product model's `deterministic-and-hermetic` invariant is **not** engaged (ADR 0012
governs the *ambient Go-env invocation*, not the toolchain version, ADR 0012 D1/D5/R1) ⇒ no
`docs/domain-model/*.modelith.{yaml,md}` change is expected (RD confirms per ADR 0041). This mirrors
round 095's "not modelled" boundary note.

---

## 3. User stories

### Story 1 — the shipped dependency set no longer trips `vulncheck` (P1)

As the tellme operator on a fresh environment, I want the shipped dependency set bumped past the
known `golang.org/x/net` vulnerabilities, so that the pipeline's `vulncheck` gate reports no *called*
vulnerability and `make check-full` is green — instead of a red gate forcing me to triage a module I
do not ship.

**Why this priority**: it is the round's core value — the quality gate must be green on the shipped
tree; a red `vulncheck` blocks delivery.

**Independent verification method**: on a patched toolchain, `make verify` reaches and passes
`vulncheck` ("No vulnerabilities found."); `go.mod` reads `golang.org/x/net v0.60.0`; the E2E count
is unchanged.

**Acceptance shape**: **Given** the bumped `go.mod`/`go.sum` on a Go ≥ 1.26.9 toolchain, **When**
`make verify` runs, **Then** the `vulncheck` gate reports *No vulnerabilities found* and the
aggregate prints `verify: OK`; and **`go test -count=1 ./...`** stays green with the same E2E figures.

**Functional Requirements (FR)**:

- **FR-001**: `go.mod` MUST require `golang.org/x/net ≥ v0.60.0`, and `go.sum` MUST be consistent
  (`go mod verify` / `go mod tidy` no-op). [Verification Intent: unobservable → the `go.mod` diff +
  `go mod tidy` no-diff]
- **FR-002**: on a Go ≥ 1.26.9 toolchain, `make verify`'s `vulncheck` gate MUST report **no called
  vulnerabilities** (exit 0). [Verification Intent: observable → the `vulncheck` gate output
  ("No vulnerabilities found.") — a black-box tool run]
- **FR-003**: the bump MUST be **behaviour-neutral** — `go build ./...` / `go vet ./...` green and
  `go test -count=1 ./...` green with the **E2E counts unchanged** (330 scenarios · 2487 steps).
  [Verification Intent: unobservable → `make check` + the E2E run]

**Non-Functional Requirements (NFR)**:

- **NFR-001**: the bump MUST NOT introduce a **new module** beyond the `golang.org/x/*` family
  already in the graph (no new vendor / no new feature dependency). [Verification Intent:
  unobservable → the `go.mod` diff]

### Story 2 — the gate's patched-toolchain precondition is discoverable (P1)

As the tellme maintainer, I want the `vulncheck` gate's **minimum patched Go toolchain** recorded on
a **live** surface, so that a fresh environment (a new container, another host) learns the
requirement from the docs instead of discovering it only through a red gate.

**Why this priority**: the stdlib half of the red gate is invisible in `go.mod`; without a recorded
floor, every fresh environment re-discovers it. It is independent of Story 1 (the module half) and
independently verifiable (a grep of the truth clause).

**Independent verification method**: grep `specs/truth/techstack.md` for the recorded toolchain
floor; cross-check the `STATUS.md` environment note.

**Acceptance shape**: **Given** the updated truth, **When** a maintainer reads
`specs/truth/techstack.md`, **Then** the toolchain floor (≥ 1.26.9, the `vulncheck` precondition) is
stated, and `make modelith-check` stays drift-free.

**Functional Requirements (FR)**:

- **FR-004**: `specs/truth/techstack.md` MUST state the minimum patched Go toolchain (**≥ 1.26.9**)
  the `vulncheck` gate depends on, as a current-state clause (not an annotated history).
  [Verification Intent: unobservable → grep of `techstack.md` + `make verify-adr-index`/`modelith-check`]
- **FR-005**: the round MUST NOT modify `specs/truth/**` anywhere **except** the single toolchain
  clause of FR-004 (no `data/**`, no `features/**`, no other `techstack.md` row).
  [Verification Intent: unobservable → the truth `git diff --stat`]

## 4. Requirements (global)

> No cross-story global FR/NFR; Stories 1–2 carry the whole round.

### Key entities

- **Dependency floor** — the minimum version of `golang.org/x/net` (and the `golang.org/x/*`
  transitive set) the shipped module graph must carry to clear the `vulncheck` gate.
- **Toolchain floor** — the minimum patched Go toolchain (≥ 1.26.9) the `vulncheck` gate's stdlib
  half requires; an *environment* prerequisite, recorded as truth, not held in a repo file.

## 5. Success criteria

### Measurable outcomes

- **SC-001**: `go.mod` carries `golang.org/x/net v0.60.0` (≥ v0.60.0); `go.sum` consistent.
  [Verification Intent: unobservable → the `go.mod` diff + `go mod tidy` no-diff]
- **SC-002**: on Go ≥ 1.26.9, `make verify` is **OK**, with `govulncheck` reporting *"No
  vulnerabilities found."* [Verification Intent: observable → the `vulncheck` gate run]
- **SC-003**: `go test -count=1 ./...` green, E2E **330 scenarios · 2487 steps** (unchanged);
  `make test-race` no races. [Verification Intent: unobservable → `make check-full`]
- **SC-004**: `specs/truth/techstack.md` states the toolchain floor (≥ 1.26.9); the truth diff touches
  **no other** clause. [Verification Intent: unobservable → the truth `git diff`]
- **SC-005**: no new module outside the `golang.org/x/*` family; no new `make verify` member.
  [Verification Intent: unobservable → the `go.mod` diff + the `verify:` aggregate]

## 6. Edge cases

- **EC-001**: on an **unpatched** toolchain (< 1.26.9), `vulncheck` still reports the stdlib
  vulnerabilities **even after** the module bump ⇒ the round must NOT claim "the gate is always
  green"; it claims "green **on a ≥ 1.26.9 toolchain**", and the floor is recorded (FR-004).
  [Verification Intent: unobservable → the recorded floor + the red-run reproduction]
- **EC-002**: a strictly newer `golang.org/x/net` (> v0.60.0) satisfies the floor and MUST NOT fail
  the requirement. [Verification Intent: unobservable → the `≥` predicate]
- **EC-003**: the `go mod tidy` `go`-directive normalisation (`1.26 → 1.26.0`) MUST be
  **semantically identical** (no minimum bump to `go 1.26.9`). [Verification Intent: unobservable →
  the `go.mod` diff review]

## 7. Invariants

- **I-1** — the bump is **behaviour-neutral**: no product code changes; the E2E contract is
  byte/step-identical.
- **I-2** — `specs/truth/**` is edited **only** by the single toolchain clause (FR-004/FR-005);
  frozen `specs/plans/NNN-*/**` packages are never modified.
- **I-3** — no Accepted-ADR **body** is edited (immutable — supersede, do not edit).
- **I-4** — the environment provisioning (Go 1.26.9 install, `deadcode` install) is **out of scope**
  for repo files; the round records its *requirement*, not its mechanism.

## 8. Assumptions

- **A1** — the kiosk Go toolchain can be raised to ≥ 1.26.9 (done here via a **user-scoped** 1.26.9
  SDK + wrapper, since the container pins `GOROOT=/usr/local/go` and `/usr/local/go` is root-only; a
  system-wide bump needs a ConfigHub/image change — recorded, not a repo artifact).
- **A2** — the "bump `golang.org/x/net ≥ v0.60.0`" directive is read **literally** (a floor, not an
  exact pin); a later compiler bump past v0.60.0 still satisfies it (EC-002). No `/axb-clarify`
  escalation — the directive locks the goal, and the residual (truth placement, ADR-or-not) is
  RD-owned.
- **A3** — whether a durable **ADR** is owed (the `vulncheck` gate is toolchain-sensitive) is decided
  in `/axb-technical-research`; this spec assumes neither. If owed, ADR 0042 (the `vulncheck` gate's
  home) is annotated by **forward pointer only** (its body stays verbatim).
- **A4** — `docs/domain-model/**` is **not** updated (no modelled entity/invariant; the boundary note
  in §2 holds) unless `/axb-technical-research` finds a modelled clause engaged (ADR 0041).
