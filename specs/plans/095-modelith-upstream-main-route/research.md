# Phase 0 研究：`095-modelith-upstream-main-route`

Truth owner: `/axb-technical-research` (the **Techstack** owner). The round has one truth change
(`specs/truth/techstack.md` *Domain model* row) and one decision record (**ADR 0065**).

## 決策 1：Install route = upstream `stacklok/modelith` `main` HEAD

- **Decision**: the route is an ordinary `go install` of the upstream module's own path at the `main`
  branch HEAD — `go install github.com/stacklok/modelith/cmd/modelith@main` — single-sourced in
  `docs/domain-model/README.md`, quoted by `$(MODELITH_INSTALL)`.
- **Rationale**: the operator directive (2026-09-27). The upstream module *is* `github.com/stacklok/modelith`,
  so `go install <module>/cmd/<bin>@<ref>` works — the fork's declared-path mismatch (which forced round
  060's clone + local build) does not apply. **Verified** on the host: `git ls-remote`/`gh api` →
  `main` HEAD `9008354f19ff13f24273a7c71395c63698c7fbac`; `go install …@main` into a temp `GOBIN` built the
  same `v0.5.1-0.20260927062055-9008354f19ff`; `make modelith-check` is **green** with that build.
- **Alternatives considered**:
  - The **fork clone + pin** route (round 060 / ADR 0030 D2) — *superseded* by the directive.
  - `go install github.com/gosharplite/modelith/...@<ref>` — **not viable** (fork declares the upstream
    module path; query form rejected, pseudo-version form fails).

## 決策 2：Track `@main` (HEAD), not a frozen commit

- **Decision**: the tracked ref is **`main`** (`MODELITH_REF := main`); the route string is `…@main`. The
  **current HEAD hash** (`9008354f19ff`) is recorded as the **observed / last-render provenance**, not as a
  pin.
- **Rationale**: the directive names *"HEAD of `stacklok/modelith` main branch"*; tracking `@main` is the
  literal reading. Recording the HEAD hash keeps the "which build rendered the committed `.md`" fact
  discoverable without re-introducing a pin.
- **Alternatives considered**:
  - **Pin `@9008354f19ff`** (the current HEAD hash) — restores hermeticity, but contradicts "HEAD of
    `main`"; if the operator wants this it is a **one-line** change (`MODELITH_REF := 9008354f19ff`), and
    A1 in `spec.md` records it as the explicit fallback.
  - **No ref / `@latest`** — meaningless for an untagged dependency; rejected.

## 決策 3：The accepted cost of tracking `main` (recorded, not fixed)

- **Decision**: **accept** that tracking `main` makes the tool version non-hermetic; record it in the README
  install section, the `Makefile` comment, the truth row, and ADR 0065.
- **Rationale**: ADR 0030 D2 already named the exact exposure (*"a fork move ⇒ a committed `.md` renders
  differently ⇒ `verify` reds on a re-install with no repo change"* — the ADR-0012 spurious-red class). The
  operator re-adopts it deliberately. Gate behaviour is unchanged: a `main` move is a **visible red**, never
  silent rot.
- **Alternatives considered**:
  - **Suppress the red** by pinning (Decision 2's alternative) — rejected (contradicts the directive).
  - **Pin *and* document `@main`** — incoherent (two owners of one fact); rejected.

## 決策 4：Single source stays `docs/domain-model/README.md`

- **Decision**: the route's **designation** is single-sourced in `docs/domain-model/README.md`; the
  `Makefile` (`$(MODELITH_INSTALL)`) and the gate's failure message **quote** it and derive it from
  `$(MODELITH_REF)`; `techstack.md`, ADR 0065, and the ADR index row each carry a **self-contained one-line
  restatement** (the round-060 **R-060-1** precedent) — the single-source discipline governs the designation,
  not verbatim re-use.
- **Rationale**: the round-060 **B-060-1** discipline (one owner, no drifting prose copies); `RF-060-1`
  already flagged the ADR↔Makefile drift risk.
- **Alternatives considered**: restate the command in every surface — rejected (the B-060-1 drift class).

## 決策 5：Supersede via a new ADR + an index pointer (ADR 0026)

- **Decision**: **ADR 0065** records the decision and **supersedes ADR 0030 D2**; ADR 0030's **body stays
  verbatim**; its **index row** gains a forward pointer to 0065. ADR 0065 does **not** touch D3/D4/D5/D6 of
  0030 (the gate, the descriptive-docs boundary, the model lifecycle are unchanged).
- **Rationale**: ADR 0026 — a recorded position is superseded by a new ADR + an index row, never by editing
  an Accepted ADR's body.
- **Alternatives considered**: edit ADR 0030 D2 in place — rejected (violates the immutability convention).

## 決策 6：No `docs/domain-model/**` YAML/MD change (not modelled)

- **Decision**: the install *route* is **not** a modelled entity/enum/glossary term/invariant; no behaviour
  changes ⇒ no `*.modelith.{yaml,md}` edit. `make modelith-check` must stay green.
  **Boundary note (F-095-3):** the product model's `deterministic-and-hermetic` invariant (*"`make verify`
  is hermetic (ADR 0012)"*) is **not** engaged — ADR 0012's hermetic boundary governs the **ambient Go-env
  invocation** (ADR 0012 D1/D5/R1), whereas this round changes a **dev-tool version**, which is outside it.
- **Rationale**: ADR 0041 load-bearing rule applies only to **modelled behaviour**; the escape hatch
  ("record why not modelled") is recorded in `plan.md` §5.
- **Alternatives considered**: add a glossary term for the toolchain — rejected (a dev-tool install route is
  process, not a product concept; the model's glossary is product/quality/environment concepts).

## 決策 7：`spec-by-example` + `dsl-refine` are NOOP

- **Decision**: no plan-side acceptance `*.feature`; no `specs/truth/features/**` change; no CLI surface.
- **Rationale**: the round changes no observable CLI contract (a dev-tool install route + its records).
- **Alternatives considered**: author a (vacuous) acceptance feature — rejected (a Make target is not
  godog-drivable; the honest carrier is the `[WITNESS]` grep/absent-binary predicate).

## Known limits / residual risks (→ ADR 0065 §Forward)

- The route tracks a **moving branch**; a `main` advance may red `modelith-check` with no repo change
  (accepted; EC-001).
- Frozen-history surfaces (`specs/plans/060-*`, the 09/19 archive/summary) legitimately still name the fork
  route — they are history, excluded from the W3 predicate.
