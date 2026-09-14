# System Analysis Plan — round 020 (`020-cross-compile-gate`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/020-cross-compile-gate/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling) ✓ done
```

*(No `features/acceptance/**` (no user-facing CLI behaviour — `/axb-spec-by-example` skipped), no `features/cli/**` change
(`/axb-dsl-refine` `NOOP`), no `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change
(`/axb-data-plan` `NOOP`), and no `ui/**` artifact.)*

### Repository structure (root)

```text
Makefile                            # CHANGED — add the `verify-cross-compile` target; wire it into `verify`
SESSION-CLOSEOUT.md                 # CHANGED — reference the gate in the Step 2 quality gates
specs/truth/techstack.md            # MODIFY — Build & Tooling: the cross-compile gate row + the target matrix
go.mod / go.sum                     # unchanged — no new dependency
internal/** , cmd/** , tests/**     # unchanged — no product code / no CLI behaviour change
```

**Structure Decision**: Round 020 is a **build-pipeline change**, not a runtime-interface change. It adds
a **host-independent cross-compile gate** (`Makefile` `verify-cross-compile`) that compiles and vets the
module for every supported POSIX target — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` —
and is a member of `make verify`, so build-tagged platform code can never silently fail to compile for a
platform the project ships. There is **no** new endpoint, **no** persisted state, **no** CLI behaviour
change, and **no** new dependency, consistent with `research.md` Decisions 1–6. The **only** truth change
is `specs/truth/techstack.md` (Build & Tooling), owned by `/axb-technical-research`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes the **build / quality pipeline**,
which is **not** a system boundary (backend / frontend / CLI): the CLI end's observable behaviour is
unchanged, so there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the build pipeline authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (no persisted or in-memory state; the gate reads nothing at runtime).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — no new/changed Gherkin or DSL rows; an existing gate's exit code is not a CLI contract).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **skipped** (no user-facing CLI journey — a build gate is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's only truth change (`techstack.md`)
is RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/020-cross-compile-gate`; truth root
`specs/truth`; truth-delta `specs/plans/020-cross-compile-gate/truth-delta.md`; interfaces: **none**;
the round's delivery is the `Makefile` gate + the closeout reference + the `techstack.md` truth row.

---

### Gating blockers

*(none — the operator locked the theme ("do B now") and confirmed the POSIX target matrix (A1);
`research.md` Decisions 1–6 settle the matrix, the mechanism, the build+vet scope, the wiring, the
dependency stance, and the unchanged must-asks. No open decision gates the round.)*
