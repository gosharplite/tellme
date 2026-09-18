# Truth Delta: 049-cli-agent-decoupling

**Plan Package**: `specs/plans/049-cli-agent-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: **initialized by `/axb-specify` (skeleton).** **No truth owner has run** — round 1 **clarify is PENDING (Q1)**, so the pipeline has not advanced. The rows below are the **expected** shape once the round proceeds; they are **not** yet ratified.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Expected **MODIFY**: RULE-E baseline figure **2 → 1** (the `internal/cli → internal/agent` edge removed); the row would cite the round's new ADR and record the port mechanism. | `truth-current`; round-049 FR-008 — *contingent on Q1 and the later research pass.* |
| _(pending)_ | `specs/truth/techstack.md` — **CLI Application** rows naming the turn path / loop construction | Expected **MODIFY**: restated to the injected-domain-port reality. | round-049 FR-008 — *contingent on Q1.* |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/` (**no `contracts/**`**) | Expected **NOOP** (checked): tellme has a single CLI end and no OpenAPI/HTTP surface. | `spec.md` A5 — *ratified when the owner runs.* |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/data-model.dbml` | Expected **NOOP** (checked): no persisted/runtime-state change (a wiring inversion, not data). | `spec.md` A5 — *ratified when the owner runs.* |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Expected **NOOP** (checked): a `cli`-reader refactor is not a CLI-contract change. | `spec.md` A2/A6 — *ratified when the owner runs.* |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `docs/decisions/NNNN-cli-agent-decoupling.md` (+ the `docs/decisions/README.md` index row) | Expected **ADD**: the domain port, its home, the behaviour-preservation claim, the baseline movement **2 → 1**, **what the change is *not***, and its relation to ADR 0011/0016/0013/0015/0017. | round-049 FR-007 — *number confirmed in research.* |
