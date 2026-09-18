# Truth Delta: 049-cli-agent-decoupling

**Plan Package**: `specs/plans/049-cli-agent-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: **initialized by `/axb-specify` (skeleton).** Round 1 clarify is **CLOSED** — **Q1 → (B) re-cut** · **Q2 → (i)** (all three crossing contracts → `internal/domain/agent`) · **Q3 → (a)** (reference + delete; no alias). **No truth owner has run yet** (the next pipeline step is `/axb-technical-research`); the rows below are the **expected** shape once the round proceeds and are **not** yet ratified.
>
> **Round shape (Q1 → B)**: this is the **re-cut sub-slice 1** of the `cli → agent` de-coupling. **The RULE-E baseline does NOT move this round** (still **2**); the baseline-moving round is **sub-slice 2** (the `AgentLoop` construction inversion, **2 → 1**).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | Expected **MODIFY**: records the re-cut (the loop contracts moved to `internal/domain/**`) and states the RULE-E baseline figure is **unchanged (2)** pending sub-slice 2; cites the round's new ADR. | `truth-current`; round-049 FR-008 — *contingent on Q2 and the later research pass.* |
| _(pending)_ | `specs/truth/techstack.md` — rows naming the CLI turn path / loop contracts | Expected **MODIFY**: restated to the domain-owned-contract reality. | round-049 FR-008 — *contingent on Q2.* |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/` (**no `contracts/**`**) | Expected **NOOP** (checked): tellme has a single CLI end and no OpenAPI/HTTP surface. | `spec.md` A5 — *ratified when the owner runs.* |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/data-model.dbml` | Expected **NOOP** (checked): no persisted/runtime-state change (a contract relocation, not data). | `spec.md` A5 — *ratified when the owner runs.* |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Expected **NOOP** (checked): a `cli`-reader refactor is not a CLI-contract change. | `spec.md` A2/A6 — *ratified when the owner runs.* |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `docs/decisions/NNNN-cli-agent-contracts-extraction.md` (+ the `docs/decisions/README.md` index row) | Expected **ADD**: the re-cut rationale, the extracted `internal/domain/**` contracts, the **deferred** construction inversion (sub-slice 2, baseline **2 → 1**), the **baseline unchanged at 2** statement, **what the change is *not***, and its relation to ADR 0011/0016/0017. | round-049 FR-007 — *number confirmed in research.* |
