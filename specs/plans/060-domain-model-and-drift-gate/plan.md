# System Analysis Plan — round 060 (`060-domain-model-and-drift-gate`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/060-domain-model-and-drift-gate/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling) ✓ done

docs/decisions/
├── 0030-domain-model-and-modelith-toolchain.md  # governance — ADD ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (the model + gate are a dev/docs surface, not the
`tellme` CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
docs/domain-model/                          # the models (NEW content; the folder already exists, empty)
├── README.md                               # NEW — authoring conventions + the pinned modelith install/@branch/version
├── tellme.modelith.yaml / .md              # NEW — the PRODUCT model (source + rendered)
├── quality.modelith.yaml / .md             # NEW — the QUALITY process model (source + rendered)
└── environment-management.modelith.yaml / .md  # NEW — the ENVIRONMENT model (external Niffler; source + rendered)
Makefile                                    # CHANGED — add `modelith-lint`/`modelith-render`/`modelith-check`; wire `modelith-check` into `verify`; `.PHONY` + `help`
docs/decisions/0030-domain-model-and-modelith-toolchain.md  # NEW — adoption + ADR 0011 D10 amendment; index row ✓ done
specs/truth/techstack.md                    # MODIFY — Build & Tooling: the Domain model row + the `verify` aggregate member ✓ done
go.mod / go.sum                             # unchanged — no new dependency (modelith is a dev-tool binary)
internal/** , cmd/** , tests/**             # unchanged — no product code / no CLI behaviour change
```

**Structure Decision**: Round 060 is a **docs + build/quality-pipeline change**, not a runtime-interface
change. It authors `tellme`'s canonical domain model (`docs/domain-model/*.modelith.yaml` → rendered
`*.modelith.md`) and adds a **modelith drift gate** (`make modelith-check` → `modelith render --check`,
wired as a **zero-tolerance** member of `make verify`). There is **no** new endpoint, **no** persisted
state, **no** CLI behaviour change, and **no** new dependency, consistent with `research.md` D1–D12. The
truth changes are `specs/truth/techstack.md` (Build & Tooling) + the governance **ADR 0030**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round changes the **docs set + the build /
quality pipeline**, which is **not** a system boundary (backend / frontend / CLI): the `tellme` CLI end's
observable behaviour is unchanged, so there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/runtime state; the models are **docs artifacts**, not runtime state).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the model + gate are a dev/docs surface (`make verify`), not the `tellme` binary's CLI contract).
> - `/axb-ui-plan` = **skipped** (no UX surface).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — a docs model + a build gate is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes (`techstack.md` +
ADR 0030) are RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/060-domain-model-and-drift-gate`; truth
root `specs/truth`; truth-delta `specs/plans/060-domain-model-and-drift-gate/truth-delta.md`; interfaces:
**none**; the round's delivery is the three models + `docs/domain-model/README.md` + the `Makefile`
`modelith-*` targets (with `modelith-check` in `verify`) + the ADR 0030 + the `techstack.md` truth rows.

---

### Pinned decisions the round implements (from clarify round 1)

| # | Decision | Carrier |
| --- | --- | --- |
| **Q1 → 3** | Author **three** models: product · quality · environment-management | `docs/domain-model/*.modelith.*`; FR-001/FR-003 |
| **Q2 → 1** | Adopt the **modelith fork** (`@feat/self-domain-model`) + `make modelith-lint\|render\|check`; `modelith` is a **dev-tool binary**, not a `go.mod` dep | `Makefile`; `docs/domain-model/README.md`; FR-004/FR-006 |
| **Q3 → 1** | `modelith-check` is a **zero-tolerance `verify` member**; an **absent binary hard-fails** naming the install command | `Makefile`; FR-004; SC-005 |
| **Q4 (A6)** | The models are **descriptive docs, not truth** — on conflict, `specs/truth/**` wins | the model `description`s; FR-008 |
| **Q5 (A7)** | Refresh **alongside a round's truth changes**; the drift gate is the safety net; no scheduled pass | FR-008/FR-010; `docs/domain-model/README.md` |

**Model file locations (Q1 + research D2).** All three sit under `docs/domain-model/` (tellme has no
`docs/architect/**` tree): `tellme.modelith.*`, `quality.modelith.*`, `environment-management.modelith.*`.
The environment model models the **external** Niffler `tellme.sh` (recorded divergence, RF-060-2).

**Gate wiring (Q2/Q3 + research D5).** POSIX-only; `MODELITH := $(shell command -v modelith 2>/dev/null)`
(no `go run …@branch` fallback); `modelith-check` requires the binary (else a named install instruction +
`exit 1`) and runs `modelith render --check` over the three YAMLs; it is added to the `verify` aggregate.

**Fork pin (research D6).** `docs/domain-model/README.md` + the truth row pin
`go install github.com/gosharplite/modelith/cmd/modelith@feat/self-domain-model` and the observed version.

---

### Gating blockers

*(none — the operator confirmed *proceed* and answered clarify Q1 (all three models), Q2 (adopt the
fork), Q3 (zero-tolerance gate; absent binary hard-fails); Q4/Q5 converged as assumptions A6/A7. No open
decision gates the round. The **plan half is complete**; `/axb-tasks` and `/axb-implement` follow.)*
