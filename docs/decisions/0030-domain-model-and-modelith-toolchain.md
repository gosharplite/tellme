# ADR 0030 — tellme domain model + the modelith toolchain (amends ADR 0011 D10)

- **Status:** Accepted. *(Its D5 / RF-060-3 — "the reference's advisory code↔model gates are not adopted" — is qualified by [0041](0041-domain-model-drift-guard.md): tellme now ships its **own**, narrower **advisory** `modelith-drift` (a stale-entry check; **not** the reference's name-diff gates, which remain unadopted on measurement).)*
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** operator request (no anchor issue) · round 060 (`specs/plans/060-domain-model-and-drift-gate` — this ADR's round) ·
  **ADR 0011 D10** (its "no modelith toolchain (no `modelith-layers` analogue)" position is **amended here** for the model + drift gate) ·
  reference `tell-me-go` (`docs/domain-model/tell-me-go.modelith.*`, `quality.modelith.*`, `docs/architect/environments/domain-model/environment-management.modelith.*`, and its `make modelith-lint|render|check` gates) ·
  `SESSION-BOOTSTRAP.md` §2.3/§2.4/§2.6 (the reference models read at bootstrap) ·
  **ADR 0012** (hermetic `make` env — the "a gate must not be a spurious-red generator" lineage) ·
  **ADR 0026** (governance: superseding a recorded position is a new ADR + an index row).

## Context

`tellme` has **no** domain model: `docs/domain-model/` exists but is **empty**, and `git log --all -- docs/domain-model` is empty. The reference `tell-me-go` ships three modelith models (product, quality, environment-management) and gates their rendered Markdown, but **ADR 0011 D10** recorded tellme's deliberate divergence: *"the reference ships a `verify-architecture` Go guard **plus** `modelith-layers`; tellme adopts the Go-guard form and has **no** modelith toolchain (no `modelith-layers` analogue)."*

The operator requested (*"I want tellme to have domain model"*) that tellme obtain its own model. Grounding confirmed the fork toolchain (`gosharplite/modelith@feat/self-domain-model`) is available as a dev-tool binary and the `domain-model-*` authoring skills are pre-loaded. Adopting the toolchain **reopens** the D10 position, so this ADR records the adoption and **amends D10** rather than silently contradicting it.

## Decision

**D1 — tellme adopts a domain model: three models under `docs/domain-model/`.** Authored as `*.modelith.yaml` **sources**, **rendered** to `*.modelith.md` (Markdown + Mermaid ER) by modelith:
- **Product** — `tellme.modelith.*` (the shipped system; entity set per `research.md` D3).
- **Quality** — `quality.modelith.*` (tellme's actual process: the `make verify` gate catalog, E2E, topology audit, ADR governance, triage per `research.md` D4).
- **Environment management** — `environment-management.modelith.*` (the **external** Niffler manager `tellme.sh` + `ait-<tag>` environments + personas + hot-swap per `research.md` D2).

**D2 — the toolchain is a dev-tool binary, not a `go.mod` dependency.** `make modelith-lint|render|check` are added (POSIX-only; `command -v modelith`, **no** network fallback). `go.mod`/`go.sum` are **unchanged**.

*Acquisition (the fork is not installable from its GitHub path).* The `gosharplite/modelith` fork **declares the upstream module path** (`github.com/stacklok/modelith`), so `go install github.com/gosharplite/modelith/...@<ref>` fails at **any** ref — the branch query form is rejected (`invalid version … disallowed version string`) and the pseudo-version form fails (`module declares its path as: github.com/stacklok/modelith`). The working route is a **local clone of the fork, checked out at an immutable commit, built locally** (a local module build is unaffected by the declared-path mismatch):

```sh
git clone https://github.com/gosharplite/modelith && cd modelith && git checkout b4153541cee8 && go install ./cmd/modelith
```

- **Single source:** this route lives in `docs/domain-model/README.md`; the `Makefile`'s `$(MODELITH_INSTALL)` and the `modelith-check` failure message quote it, and `specs/truth/techstack.md` cites it — never restated in five prose copies.
- **Pinned identifier (immutable):** commit **`b4153541cee8`** (the `feat/self-domain-model` tip this round was authored against). Branch-tracking is an explicit, documented **upgrade**, not the default. A fork move changes the renderer ⇒ a committed `.md` renders differently ⇒ `verify` reds on a re-install with **no repo change** (the ADR-0012 "spurious-red generator" class, here a *tool-version* non-hermeticity) — the observable consequence of the pin (**TD-060-1**).
- **Known-good provenance:** `github.com/stacklok/modelith v0.0.0-20260815121344-b4153541cee8`, commit `b4153541cee8`, `vcs.modified=false` — the build the three committed `.md` files were rendered with (recorded as provenance, not as a claim about which tool the gate "supports").
- **Enumeration is a tree property (TD-060-2):** `MODELITH_MODELS = $(wildcard docs/domain-model/*.modelith.yaml)` with a non-empty assertion, so a newly added model is covered by construction and an empty set fails (never a vacuous green) — the round-042 B-2 / round-047 RULE-D default-deny lineage.

**D3 — the drift gate is a zero-tolerance `make verify` member.** `modelith-check` (`modelith render --check`) joins the aggregate `verify`: **drift ⇒ fail**; an **absent `modelith` binary ⇒ fail**, naming the install route (D2). The gate **never** silently skips. (Recorded consequence: `make verify` requires the modelith dev tool.)

**D4 — the models are descriptive docs, subordinate to truth.** They are **not** AIxBDD `TruthArtifact`s; on any conflict with `specs/truth/**`, **truth wins** and the model is corrected. The drift gate protects model-internal consistency (YAML ↔ `.md`), **not** model-vs-code.

**D5 — ADR 0011 D10 is amended, not repealed.** tellme now **has** a modelith toolchain (the model + drift gate) — so D10's "no modelith toolchain" clause is **superseded for the model**. tellme still ships **no** `modelith-layers` architecture gate: the Go-guard form of `verify-architecture` (ADR 0011) is **retained**. The reference's advisory `modelith-drift`/`modelith-layers` code↔model gates remain **not adopted**.

**D6 — lifecycle.** No scheduled refresh pass: the models are refreshed alongside a round's truth changes (same PR); the drift gate is the safety net. The models represent the *current* system; they are not plan packages and have no `delivered` freeze.

## Consequences

- A `*.modelith.yaml` is the single source; the committed `*.modelith.md` is **generated** (never hand-edited) — the `domainmodel-md-generated` discipline the reference records.
- `make verify` gains a member and a **new host prerequisite** (`modelith`); a host without it fails by design (D3). This is the accepted cost of a non-rotting model.
- ADR 0011's *Layer-discipline gate* truth row and D10's prose are corrected in the same round to stop asserting "no modelith toolchain" (truth-current).
- The environment model describes an **external** system (the Niffler `tellme.sh`), recorded as a divergence (RF-060-2); the quality model records tellme's genuine divergence from the reference (no `NonFixCatalog`; a lighter gate catalog) (RF-060-4).

## Forward (non-blocking)

- **RF-060-1** — tool-version drift: the pin is now the immutable commit `b4153541cee8` (D2); a deliberate fork upgrade is an explicit, documented step, and a fork move makes `verify` red on re-install (a visible red, not silent rot).
- **RF-060-2** — the environment model is a model *about* an external system, not this repo's runtime.
- **RF-060-3** — the reference's advisory code↔model gates (`modelith-drift`, `modelith-layers`) are not adopted. *(Qualified by [ADR 0041](0041-domain-model-drift-guard.md): tellme ships its **own** advisory `modelith-drift` — a stale-entry check, not the reference's name-diff gates, which a measurement rejected as ~79 % false-positive.)*
- **RF-060-4** — the quality model documents tellme's reality (divergences recorded, not smoothed).
- **RF-060-5** — `make verify` now requires the modelith dev tool (D3); install it via the clone+pinned-build route (D2).
