# Truth Delta: 042-layer-discipline-gate

**Plan Package**: `specs/plans/042-layer-discipline-gate`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md` (**Build & Tooling**): add a **layer-discipline gate** row and name the new member in the **Task runner** row's `verify` aggregate list. Also **MODIFY** the row(s) that reference the gate catalog if needed. *No new third-party dependency.*
> - `/axb-api-plan` — **NOOP** (no HTTP/OpenAPI surface).
> - `/axb-data-plan` — checked **NOOP** (the baseline is a **repo artifact**, not persisted runtime state; no `specs/truth/data/**` change).
> - `/axb-dsl-refine` — **NOOP-or-add** (RD decision, spec A6): the gate's dev-observable behaviour *may* warrant a CLI-truth interface row (e.g. "`make verify` fails on a new layer violation"). If NOOP, name what was inspected.
> - Governance — an **ADR is likely NOT needed** for R1 (the #92 ADR obligation attaches to R3's yield policy); a short layer-model ADR is an RD decision (spec A5).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — **Build & Tooling** | TODO: new **layer-discipline gate** row (rule model Q1, package scope Q2, baseline + stale-entry policy Q3, hermetic/host-independent, gate form A4) | round 042 FR-009; spec §Grounded |
| *(pending)* | `specs/truth/techstack.md` — **Task runner** row | TODO: extend the `verify` aggregate list with the new member | round 042 FR-009 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/contracts/**` | TODO: expected NOOP — tellme has a single CLI end and no OpenAPI surface; inspect and record. | spec A6 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/data-model.dbml` | TODO: expected checked NOOP — the baseline is a repo artifact, not persisted runtime state. | spec A6 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/**` + `chat/dsl.md` | TODO: NOOP-or-add — decide whether the gate's dev-observable behaviour is a CLI-truth interface (a `Rule`/`Example`/`DSLRow`) or carried by the gate + unit seams. | spec A3/A6 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `docs/decisions/*` | TODO: likely **no ADR** for R1 (the #92 ADR obligation attaches to R3); if the layer model needs a durable home, add a short ADR + index row. | spec A5 |
