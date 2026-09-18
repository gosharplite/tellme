# Truth Delta: 050-agentloop-construction-inversion

**Plan Package**: `specs/plans/050-agentloop-construction-inversion`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: **skeleton — initialized by `/axb-specify`.** Owner rows are **not yet filled**: the round is **paused at clarify round 1 (Q1 OPEN)**. The owner skills run after Q1–Q3 close.
>
> **Round shape (sub-slice 2, pending Q1)**: the **baseline-moving** round — invert the surviving `AgentLoop` **construction/execution** into an injected domain port, so the `internal/cli → internal/agent` edge disappears and the RULE-E ratchet moves **2 → 1**. **Two** ratchet removals (RULE-E baseline line + RULE-F `couplingSurface` key; ADR 0018 fold F-4); the `internal/cli → internal/ui` line/surface stays **byte-identical**.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Expected: record the round-050 note + the baseline figure **2 → 1** (+ the removed RULE-F `→ agent` key). | round 050 FR-008; `research.md` |
| (pending) | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | Expected: record the loop's construction home (the injected domain port + its adapter). | round 050 FR-008 |
| (pending) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Expected: NOOP (no new Makefile target — the gate rides `verify-architecture`). | round 050 NFR-002 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/` (**no `contracts/**`**) | Expected: NOOP — tellme has a single CLI end and no OpenAPI/HTTP surface; the round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/data-model.dbml` | Expected: NOOP — no persisted/runtime-state change (the round moves a construction, not data). | `spec.md` A5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Expected: NOOP — an internal refactor is not a `tellme` CLI-contract change (no Rule/Example/step/`DSLRow` change). | `spec.md` A2/A6; rounds 020/031/036/041–049 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `docs/decisions/0019-agentloop-construction-inversion.md` (+ the `docs/decisions/README.md` index row) | Expected: record the inversion (port home, adapter home, the two ratchet removals, what the change is *not*, and the relation to ADR 0011/0016/0017/0018). | round 050 FR-007 |
