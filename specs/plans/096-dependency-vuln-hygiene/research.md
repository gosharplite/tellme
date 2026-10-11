# Research — round 096 `096-dependency-vuln-hygiene`

Truth owner: `/axb-technical-research` (the **Techstack** owner). One truth change
(`specs/truth/techstack.md` — the *Dependency vulnerability scan* row) + one decision record (**ADR 0066**)
+ the `go.mod`/`go.sum` bump.

## Decision 1 — bump `golang.org/x/net` to ≥ v0.60.0 (the module half)

- **Decision**: raise `golang.org/x/net` from `v0.38.0` → **`v0.60.0`** (the first version clearing the
  flagged advisories), via `go get golang.org/x/net@v0.60.0` + `go mod tidy`.
- **Rationale**: `govulncheck` flags `golang.org/x/net@v0.38.0` with **`GO-2026-6617/6612/6611/6610/6603`**
  (fixed in `v0.60.0`). The module is an **indirect** dependency (via `modelcontextprotocol/go-sdk` /
  `x/oauth2`), so the bump also rises the transitive `golang.org/x/*` set it requires (`x/sys` v0.41.0 →
  v0.48.0 · `x/term` v0.36.0 → v0.46.0 · `x/sync` v0.21.0 → v0.23.0 · `x/text` v0.39.0 → v0.42.0).
- **Verified**: after the bump, `govulncheck ./...` → **"No vulnerabilities found."** (both halves clear
  given a ≥ 1.26.9 toolchain; see Decision 2).
- **Alternatives considered**:
  - **Do nothing** — the gate reds; a red `vulncheck` blocks delivery. Rejected.
  - **Pin a newer-but-not-minimal version** — needless churn; `≥ v0.60.0` (the fix floor) is the honest
    target (EC-002 allows a later version).

## Decision 2 — record the toolchain floor (≥ 1.26.9) on the `vulncheck` gate's own truth row

- **Decision**: state the gate's **patched-toolchain precondition** in `specs/truth/techstack.md` as a
  **current-state** clause on the *Dependency vulnerability scan* (`govulncheck`) row — the single truth
  home for that gate: a **Go ≥ 1.26.9** toolchain (the stdlib fixes land in `go1.26.9`) + the shipped
  **module floor** `golang.org/x/net ≥ v0.60.0`.
- **Placement choice**: the *vulncheck* row (not the *Language* row) — it is the gate's own row, so the
  environmental precondition sits beside the gate it constrains; a single clause (FR-005).
- **Rationale**: the stdlib half of the red gate is an **environment** input invisible in `go.mod`, so
  without a recorded floor every fresh host re-discovers it via a red gate. This is the *same shape* as
  round 095 (a dev-tool/gate route is a **truth clause**, not just prose).
- **Alternatives considered**:
  - **Record it in the *Language* row** — also valid; rejected in favour of the gate's own row (single owner
    = the gate). The *Language* row still reads `Go 1.26` (still true for 1.26.9).
  - **Leave the floor implicit** — rejected (re-discovery per host; the round-095 lesson).

## Decision 3 — a durable ADR (0066) records the gate's toolchain-sensitivity

- **Decision**: **ADR 0066** records: the `vulncheck` gate is **toolchain-sensitive** (its stdlib half
  depends on the ambient Go patch level), the **module floor** (`x/net ≥ v0.60.0`) and the **toolchain
  floor** (≥ 1.26.9), and the standing method ("when `vulncheck` reds: clear the module half in `go.mod`,
  satisfy the toolchain half in the environment; do not conclude from memory").
- **Rationale**: the repo records durable quality decisions as ADRs (ADR 0042 = the gate set; ADR 0064 =
  the advisory carrier). A gate whose greenness depends on an *ambient* input is a durable fact future
  sessions must know.
- **Relates to (not supersedes)**:
  - **ADR 0012** — its **R1** already states *"hermeticity is a `make`-boundary property, not a toolchain
    property"*; this round is the concrete instance (the toolchain version is outside the make-hermetic
    boundary), so ADR 0066 **cites** 0012, it does **not** amend it.
  - **ADR 0042** — the `verify` aggregate that carries `vulncheck`; **no body change** (no forward pointer
    owed — 0042 owns `verify-fmt`/`verify-adr-index`/`test-race`, not the pre-existing `vulncheck` gate).
  - **ADR 0041** — the load-bearing model rule; the round is **not modelled** (Decision 5).
- **Alternatives considered**: a plan-side note only — rejected (does not outlive the frozen package;
  Bootstrap Agent Rule 11 homes a durable disclosure in an ADR or a live issue).

## Decision 4 — the `go` directive normalisation is inert

- **Decision**: accept `go mod tidy`'s `go 1.26 → go 1.26.0` rewrite as a **no-op** normalisation (the
  two spellings are semantically identical: a minor version with no patch component).
- **Rationale**: it MUST NOT be read as raising the *minimum* toolchain to `1.26.9` (the floor is a
  `vulncheck` **environment** requirement, recorded in truth, **not** a `go.mod` directive — raising the
  directive would change the module's declared minimum and could break older builders unnecessarily).
- **Alternatives considered**: set `go 1.26.9` in `go.mod` — rejected (over-constrains the module; the
  floor is a *gate* precondition, not a build precondition).

## Decision 5 — no `docs/domain-model/**` change (not modelled)

- **Decision**: a dependency **version** and an **ambient toolchain** are neither entities, enums,
  glossary terms, nor invariants; no behaviour changes ⇒ **no** `*.modelith.{yaml,md}` edit.
  **Boundary note:** the product model's `deterministic-and-hermetic` invariant (*"`make verify` is
  hermetic (ADR 0012)"*) is **not** engaged — ADR 0012 **R1** places the toolchain *outside* the
  make-hermetic boundary, exactly the class this round touches.
- **Rationale**: ADR 0041 load-bearing rule applies only to **modelled behaviour**; the escape hatch
  ("record why not modelled") is in `plan.md` §5.
- **Alternatives considered**: add a dependency/version glossary term — rejected (process/environment, not
  a product concept).

## Decision 6 — `spec-by-example` + `dsl-refine` are NOOP

- **Decision**: no plan-side acceptance `*.feature`; no `specs/truth/features/**` change; no CLI surface.
- **Rationale**: a behaviour-neutral dependency bump has no observable CLI contract change; the honest
  carrier is the relocated gate run + the `go.mod`/`go.sum` diff + the truth grep (`[WITNESS]`).
- **Alternatives considered**: author a (vacuous) acceptance feature — rejected (not godog-drivable).

## Environment facts (reproduced, 2026-10-11 — the kiosk GKE container)

- Baseline red: Go **1.26.6**, `dev` `99f1114` ⇒ `make check-full` stops at `vulncheck`
  (*"your code is affected by 9 vulnerabilities from the Go standard library"* + `x/net@v0.38.0`).
- Fix: Go **1.26.9** (user-scoped SDK `~/sdk/go1.26.9` + a `~/.local/bin/go` wrapper that pins
  `GOROOT`, since the container pins `GOROOT=/usr/local/go` and `/usr/local/go` is root-only) +
  `golang.org/x/net@v0.60.0` ⇒ `govulncheck` **"No vulnerabilities found."**; `make verify` **OK**;
  `make check-full` **exit 0** (E2E **330 scenarios · 2487 steps**).
- Reproduce the module half anywhere: `go get golang.org/x/net@v0.60.0 && go mod tidy`.
- Reproduce the gate: `go install golang.org/x/vuln/cmd/govulncheck@latest; govulncheck ./...`.

## Known limits / residual risks (→ ADR 0066 §Forward)

- The toolchain floor is **recorded, not machine-checked** (a `go.mod` `go` directive is the wrong lever —
  Decision 4); a maintainer must read it (the ADR-0060 §Forward RF-089-6 class — a docs claim with no
  carrier). A host with Go < 1.26.9 reds `vulncheck` with no repo change.
- The **environment** provisioning (the 1.26.9 install) is a ConfigHub/image concern, outside the repo;
  the repo records *that* it is required, not *how* it is provided.
