# ADR 0065 — `modelith` dev tool: install from upstream `stacklok/modelith` `main` HEAD (supersedes ADR 0030 D2)

- **Status:** Accepted
- **Date:** 2026-09-27
- **Deciders:** tellme owner (operator directive, 2026-09-27; issue [#200](https://github.com/gosharplite/tellme/issues/200))
- **Related:** **ADR 0030** (the domain model + the toolchain; **supersedes its D2 acquisition route** — its
  §Forward **RF-060-1**/RF-060-5 are qualified) · **ADR 0012** (hermetic `make` — the "a gate must not be a
  spurious-red generator" lineage the pin served) · **ADR 0041** (the model is load-bearing; the
  install route is **not** modelled) · **ADR 0026** (a superseding decision is a new ADR + an index row) ·
  **ADR 0064** (the same-session precedent: a superseding ADR + a dev-tool binary with a provenance caveat)
  · `docs/domain-model/README.md` (the single source of the route) · `Makefile` (`MODELITH_REF` /
  `MODELITH_INSTALL`) · `specs/truth/techstack.md` (*Domain model* row) · upstream
  `github.com/stacklok/modelith`

## Context

ADR 0030 D2 recorded a **fork + immutable-commit clone** acquisition route for the `modelith` dev-tool
binary, because the `gosharplite/modelith` *fork* (branch `feat/self-domain-model`) **declares the upstream
module path** (`github.com/stacklok/modelith`) and is therefore **not** `go install`-able from its own GitHub
path:

```sh
git clone https://github.com/gosharplite/modelith && cd modelith && git checkout b4153541cee8 && go install ./cmd/modelith
```

The route is single-sourced in `docs/domain-model/README.md`, quoted by the `Makefile`'s
`$(MODELITH_INSTALL)` and by the `modelith-check` failure message, and cited by `specs/truth/techstack.md`.
D2 chose an **immutable commit pin** because the renderer affects the committed Markdown — *"a fork move ⇒
a committed `.md` renders differently ⇒ `verify` reds on a re-install with **no repo change**"* (the
ADR-0012 "spurious-red generator" class, here a *tool-version* non-hermeticity).

The operator has directed (2026-09-27) that `modelith` **MUST be installed from the HEAD of the upstream
`stacklok/modelith` `main` branch** — an ordinary `go install` of the upstream module's own path. The fork +
pin route is **superseded.**

**Verified (2026-09-27, authoring host):** upstream `main` HEAD = `9008354f19ff13f24273a7c71395c63698c7fbac`
(pseudo-version `github.com/stacklok/modelith v0.5.1-0.20260927062055-9008354f19ff`);
`go install github.com/stacklok/modelith/cmd/modelith@main` **resolves and builds** that same commit, and
`make modelith-check` is **green** with it (the upstream renderer produces byte-identical output for the three
committed `*.modelith.md`).

## Decision

**D1 — the route is the upstream `main` HEAD.** `modelith` is installed with

```sh
go install github.com/stacklok/modelith/cmd/modelith@main
```

— an ordinary `go install` of the upstream module's own path; the fork + commit-pin clone route is
**superseded**. The route is single-sourced in `docs/domain-model/README.md` (the **single source**); the
`Makefile` (`$(MODELITH_INSTALL)`) and the `modelith-check` failure message **quote** it; `techstack.md` and
this ADR **cite** that section rather than restating the command.

**D2 — the tracked ref is `main` (HEAD); the current HEAD hash is recorded as provenance, not a pin.**
`MODELITH_REF := main`; `MODELITH_INSTALL := go install github.com/stacklok/modelith/cmd/modelith@$(MODELITH_REF)`.
The README records the HEAD hash observed at adoption (`9008354f19ff`) as the build the committed `.md` were
last rendered with. The binary remains a **dev-tool binary** — **not** a `go.mod` dependency (`go.mod`/
`go.sum` unchanged).

**D3 — the accepted cost: tracking `main` makes the tool version non-hermetic.** ADR 0030 D2 named the exact
exposure (a `main` advance can change the renderer ⇒ a committed `.md` renders differently ⇒ `modelith-check`
reds on a re-install with **no repo change**). This is **re-adopted deliberately** by operator decision, and
recorded in the README, the `Makefile` comment, `specs/truth/techstack.md`, and this ADR. The gate behaviour
is unchanged: a `main` move is a **visible red**, never silent rot.

**D4 — ADR 0030 D2 is superseded, not repealed.** D2's acquisition route is replaced by D1/D2 above. ADR
0030's **body stays verbatim**; its **index row** gains a forward pointer to this ADR (ADR 0026). This ADR
does **not** touch ADR 0030's **D3** (the zero-tolerance `modelith-check` gate), **D4** (the models are
descriptive docs, subordinate to truth), **D5** (the ADR 0011 D10 amendment), or **D6** (the model
lifecycle) — all are unchanged.

**D5 — the install route is not modelled.** No `docs/domain-model/*.modelith.{yaml,md}` change: the route is
process, not a product/quality/environment concept, and no behaviour changes (ADR 0041 escape hatch).

**D6 — frozen history is out of scope.** `specs/plans/060-domain-model-and-drift-gate/**`,
`docs/archives/status/2026-09-19.md`, and `docs/session-summary/2026/09/19/**` legitimately still name the
fork route (they record the `B-060-1` / `TD-060-1` fork-era decision) and MUST NOT be edited
(`plan-package-frozen` / Rule 12).

## Consequences

- A fresh host installs the upstream tool with one `go install`; the private fork + commit pin is no longer
  required.
- `make verify`'s `modelith-check` member is unchanged in behaviour (drift ⇒ fail; absent binary ⇒ fail naming
  the route) — only the **named route** changes. The host prerequisite stays: `make verify` requires `modelith`
  (ADR 0030 D3).
- The tool version tracks `main`; a `main` advance may red `modelith-check` with no repo change (accepted, D3).
- Live surfaces no longer reference the fork/pin; the historical records retain them (D6).

## Forward (non-blocking)

- **RF-065-1** — the tracked ref is a **moving** branch (`main`); a future operator may prefer pinning a
  commit — a **one-line** change (`MODELITH_REF`), not a new decision shape.
- **RF-065-2** — the "which build rendered the committed `.md`" provenance is **recorded at adoption** but not
  machine-checked; a `main` advance is detected only by a `modelith-check` red on re-render.
- **RF-065-3** — the upstream-published binary route (a release download) is **not** adopted; `go install` from
  the module path is the chosen acquisition.
- **RF-065-4** — ADR 0030's own prose (D2 + §Forward RF-060-1/RF-060-5) still describes the fork pin
  verbatim; it is **corrected forward** by this ADR's index pointer, not by editing the body.
