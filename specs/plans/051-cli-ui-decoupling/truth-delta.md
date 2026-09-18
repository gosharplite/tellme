# Truth Delta: 051-cli-ui-decoupling

**Plan Package**: `specs/plans/051-cli-ui-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: **skeleton — initialized by `/axb-specify`.** Owner rows are **not yet filled**: the round is **paused at clarify round 1 (Q1 OPEN)**. The owner skills run after Q1–Q4 close.
>
> **Programme goal**: **close [#101](https://github.com/gosharplite/tellme/issues/101)** — remove the last RULE-E residual (`internal/cli → internal/ui`; baseline **1 → 0**) **and** resolve the F-6/F-7/F-8 deferrals. Q1 decides how much of the programme this round carries.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Expected: record the round-051 outcome (the baseline figure → **0** if the `→ ui` edge is removed this round; the RULE-F `→ ui` key removal; the round-051 note). | round 051 FR-009 |
| (pending) | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row and any row naming the CLI's ui rendering seams | Expected: record any moved symbols (value types, presentation port(s), the renderer/spinner/coordinator homes). | round 051 FR-001/FR-002 |
| (pending) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Expected: NOOP (no new Makefile target — the gate rides `verify-architecture`). | round 051 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/` (**no `contracts/**`**) | Expected: NOOP — a single CLI end, no OpenAPI/HTTP surface; the round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/data-model.dbml` | Expected: NOOP — no persisted/runtime-state change (a rendering de-coupling + seam refactors). | `spec.md` A5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/**` and `specs/truth/features/cli/**/dsl.md` | Expected: NOOP — a `cli`-side rendering de-coupling is not a tellme CLI-contract change (no Rule/Example/step/`DSLRow` change). | `spec.md` A2/A6; the rounds 020/031/036/041–050 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `docs/decisions/0020-*.md` (+ the `docs/decisions/README.md` index row) | Expected: record the `→ ui` de-coupling decision (value-type homes + the presentation port(s)) and the F-6/F-7/F-8 decisions; state what the change is *not* and its relation to ADR 0006/0008/0013/0015/0016/0017/0018/0019. | round 051 FR-009 |
