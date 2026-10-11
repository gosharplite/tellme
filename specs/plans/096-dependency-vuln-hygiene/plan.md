# System analysis — round 096 `096-dependency-vuln-hygiene`

## Project structure

### Document structure (this feature)

```text
specs/plans/096-dependency-vuln-hygiene/
├── spec.md
├── checklists/requirements.md
├── research.md
├── plan.md
├── tasks.md
└── truth-delta.md
```

**Structure decision**: a **dependency/toolchain hygiene** round — no `features/acceptance/`
(`spec-by-example` NOOP), no contracts/data, no CLI surface.

## Analysis flow

### System interface inventory

This requirement inventories **1** system interface.

1. `dependency / toolchain vulnerability floor`
   - Endpoint type: `Toolchain / Make / go.mod` (not a shipped user interface)
   - Main surfaces: `go.mod` / `go.sum` (the module floor) + `Makefile` (`vulncheck` gate) +
     `specs/truth/techstack.md` (the *Dependency vulnerability scan* truth row) +
     `docs/decisions/0066-*.md` (the decision record)
   - Requirement basis: operator directive 2026-10-11 (bump `golang.org/x/net ≥ v0.60.0`; the
     `vulncheck` gate red on the new environment)

### Wave plan

#### Wave 1

- Interface: `dependency / toolchain vulnerability floor`
- Analysis focus: the module floor (`x/net ≥ v0.60.0`), the toolchain floor (≥ 1.26.9), the gate's
  toolchain-sensitivity, the not-modelled boundary
- Rationale: a single interface; its truth owner is `/axb-technical-research` (the **Techstack** owner).

## Wave delegation

- **`/axb-technical-research`** — the **Techstack** truth owner: `specs/truth/techstack.md`
  (*Dependency vulnerability scan* row) **MODIFY**; authored the ADR 0066 decision record
  (`research.md` D1–D6).
- **`/axb-api-plan`** — **NOOP** (`truth-delta.md`): no API surface.
- **`/axb-data-plan`** — **NOOP** (`truth-delta.md`): no state.
- **`/axb-ui-plan`** — **skipped**: a line-oriented CLI, and this round has no UX surface.
- **`/axb-dsl-refine`** — **NOOP** (`truth-delta.md`): no `.feature`/`dsl.md` change (no observable CLI
  contract change).

## §5 — `docs/domain-model/**` not modelled (ADR 0041 escape hatch)

This round changes **no modelled behaviour**: a dependency **version** and an **ambient toolchain** are
not any model's entity / relationship / attribute / enum / glossary term / scenario. So
`docs/domain-model/*.modelith.{yaml,md}` are **NOT** updated (`make modelith-check` must stay green).
This is the ADR 0041 "or record in the plan package why it is not modelled" case.

**Boundary note:** the product model's `deterministic-and-hermetic` invariant (*"`make verify` is hermetic
(ADR 0012)"*) is **not** engaged — ADR 0012 **R1** records that *"hermeticity is a `make`-boundary
property, not a toolchain property"*, and this round changes a **toolchain version**, which is exactly the
input *outside* that boundary — so the invariant's `(ADR 0012)` reference is unaffected and the same-PR
model rule does not apply.

## Artifacts touched

- `go.mod` / `go.sum` — `golang.org/x/net v0.38.0 → v0.60.0` (+ the transitive `golang.org/x/*` rise);
  `go mod tidy` normalises the `go` directive `1.26 → 1.26.0` (inert — research D4).
- `specs/truth/techstack.md` — the *Dependency vulnerability scan* (`govulncheck`) row (**MODIFY**).
- `docs/decisions/0066-dependency-vuln-hygiene.md` (**new**) + `docs/decisions/README.md` (a new index row).
- `STATUS.md` + `docs/session-summary/2026/10/11/session-summary.md` (live state + day log; closeout).
