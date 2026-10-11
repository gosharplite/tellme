# ADR 0066 — Dependency / toolchain vulnerability hygiene: the `vulncheck` gate's floors

- **Status:** Accepted
- **Date:** 2026-10-11
- **Deciders:** tellme owner (operator directive, 2026-10-11)
- **Related:** **ADR 0012** (hermetic `make` Go-env — its **R1** already records *"hermeticity is a `make`-boundary property, not a toolchain property"*; this ADR is the concrete instance and **cites**, not amends, 0012) ·
  **ADR 0042** (the `verify` aggregate; `vulncheck` is a member, its **body** is unchanged — no forward pointer owed) ·
  **ADR 0041** (the model is load-bearing; this round is **not modelled**) ·
  **ADR 0064**/**ADR 0065** (the same-shape precedent: a durable quality/tooling fact recorded as an ADR) ·
  `Makefile` (`vulncheck`) · `specs/truth/techstack.md` (*Dependency vulnerability scan* row) · `go.mod`/`go.sum`

## Context

`make verify` includes a **zero-tolerance `vulncheck`** gate (`govulncheck`). On a fresh `kiosk`
environment (a GKE container managed by ConfigHub; Go **1.26.6**) the gate is **RED** on a clean `dev`
tree:

```
Your code is affected by 9 vulnerabilities from the Go standard library.
This scan also found 9 vulnerabilities in packages you import and 7 vulnerabilities in modules you
require, but your code doesn't appear to call these vulnerabilities.
```

The red has **two** independent axes, only one of which is a repo artifact:

- **stdlib** — the fixes are in **`go1.26.9`** (e.g. `GO-2026-6617` HPACK race, `GO-2026-6613`,
  `GO-2026-6605` — all `net/http@go1.26.9`). This is the **ambient Go toolchain** — an *environment*
  input, invisible in `go.mod`.
- **module** — `golang.org/x/net@v0.38.0` (indirect, via the MCP SDK / `x/oauth2`) is flagged
  (`GO-2026-6617/6612/6611/6610/6603`), fixed in **`v0.60.0`**. This **is** a repo artifact.

The operator directed (2026-10-11): raise the kiosk toolchain to **≥ 1.26.9** and bump
`golang.org/x/net` to **≥ v0.60.0**. **Verified:** Go 1.26.9 + `x/net v0.60.0` ⇒ `govulncheck` →
**"No vulnerabilities found."**; `make verify` **OK**; `make check-full` **exit 0**.

## Decision

**D1 — the module floor: `golang.org/x/net ≥ v0.60.0`.** The shipped `go.mod` MUST carry
`golang.org/x/net ≥ v0.60.0` (the first version clearing the flagged advisories), with `go.sum`
consistent; the transitive `golang.org/x/*` rise the version requires is accepted. No module outside the
`golang.org/x/*` family is introduced.

**D2 — the toolchain floor: the environment MUST provide a patched Go (≥ 1.26.9).** The `vulncheck`
gate's stdlib half is clean only on a patched toolchain; the floor is recorded in
`specs/truth/techstack.md` (the *Dependency vulnerability scan* row — the gate's own truth home). The
floor is an **environment** prerequisite (a ConfigHub / image concern), **not** a `go.mod` directive:
the `go` directive stays `1.26` (its `go mod tidy` normalisation to `go 1.26.0` is **inert** —
D4), because raising the directive would over-constrain the module's declared build minimum.

**D3 — the standing method.** When `vulncheck` reds, clear the **module** half in `go.mod` (bump past the
recorded floor) and satisfy the **toolchain** half in the environment (≥ 1.26.9); never conclude from
memory (the repo's live-fetch discipline). The two floors are recorded in the truth row so a fresh host
learns them from the docs rather than only from a red gate.

**D4 — the `go` directive normalisation is inert.** `go mod tidy`'s `go 1.26 → go 1.26.0` rewrite is a
no-op spelling normalisation (a minor version with no patch component); it MUST NOT be read as raising
the minimum toolchain. The floor is a *gate* precondition (D2), not a module-build precondition.

**D5 — not modelled.** No `docs/domain-model/*.modelith.{yaml,md}` change: a dependency version and an
ambient toolchain are process/environment facts, not product concepts, and no behaviour changes (ADR 0041
escape hatch). The product model's `deterministic-and-hermetic` invariant (*"`make verify` is hermetic
(ADR 0012)"*) is **not** engaged — ADR 0012 **R1** places the toolchain *outside* the make-hermetic
boundary, exactly the input this round changes.

**D6 — this ADR does not amend any Accepted ADR body.** It **cites** ADR 0012 R1 (the toolchain is
outside the make-hermetic boundary) and relates to ADR 0042 (the `verify` aggregate) without editing
either; no back-pointer is owed (ADR 0042 does not own the pre-existing `vulncheck` gate).

## Consequences

- The shipped tree no longer trips `vulncheck`'s **module** half; the **stdlib** half is clean on any
  Go ≥ 1.26.9 host (the environment floor is now recorded).
- The bump is **behaviour-neutral**: no product code, no `.feature`, no new `make verify` member; the E2E
  contract is unchanged (**330 scenarios · 2487 steps**).
- A host with Go < 1.26.9 reds `vulncheck` with **no repo change** — a **visible red**, never silent rot
  (the ADR-0012 "a gate must not be a spurious-red generator" concern, here mitigated by *recording* the
  floor rather than by pinning an unavailable toolchain in `go.mod`).

## Forward (non-blocking)

- **RF-066-1** — the toolchain floor (≥ 1.26.9) is **recorded, not machine-checked**: a `go.mod` `go`
  directive is the wrong lever (D4), so a maintainer/host must read the truth row. A future `make` guard
  could parse `go env GOVERSION` and warn — deferred (no carrier today; the ADR-0060 §Forward RF-089-6
  class — a docs claim with no mechanical carrier).
- **RF-066-2** — the **environment** provisioning of Go ≥ 1.26.9 (image / ConfigHub) is outside the repo;
  the repo records *that* it is required, not *how* it is provided.
- **RF-066-3** — the module floor is stated as `≥`; a future bump past `v0.60.0` stays valid and is
  recorded by the then-current `go.mod` (the truth row cites the floor, not an exact pin).
- **RF-066-4** — the `golang.org/x/net` version is **not** recorded in `techstack.md` as a dependency row
  (it is indirect); the floor lives in the gate's row + `go.mod`. If the dependency ever becomes direct,
  it earns its own row (the `golang.org/x/term` precedent).
