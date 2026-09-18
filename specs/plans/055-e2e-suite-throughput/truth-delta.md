# Truth Delta: 055-e2e-suite-throughput

**Plan Package**: `specs/plans/055-e2e-suite-throughput`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`; **clarify round 1 CLOSED** — **Q1 → B** (parallelism on by default at `4` + the `TELL_ME_E2E_CONCURRENCY` seam; ADR-0010 timing protection) · **Q2 → A** (`make test-fast` over `godog.paths`; `SUBSET — NOT THE GATE` banner + a never-the-gate guard; default = the non-`chat` modules) · **Q3 → A** (gate ≤ 60 % of the paired serial baseline; N = 5 green runs). **Owner rows recorded** (see below): `/axb-technical-research` = 4 MODIFYs; `/axb-api-plan` + `/axb-data-plan` + `/axb-dsl-refine` = NOOP; governance = **ADR 0024** (ADD).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **E2E runner / step definitions** row | **Round 055**: the suite gains parallel scenario execution (godog `Concurrency` + an override seam) and a documented subset/gate invariant. | `spec.md` US1/US2; `research.md` D1/D3 |
| MODIFY | `specs/truth/techstack.md` — **Host test harness** row | **Round 055**: the timing-sensitive scenarios' protection under parallelism (per the Q1 decision). | `spec.md` FR-004; `research.md` D2 |
| MODIFY | `specs/truth/techstack.md` — **Task runner** row | **Round 055**: a fast subset target (convenience; not part of `verify`/the gate). | `spec.md` US2 / FR-005; `research.md` D3 |
| MODIFY | `specs/truth/techstack.md` — **Test strategy** row | **Round 055**: the gate's scope is pinned as *all Examples*, independent of the subset target. | `spec.md` FR-002/FR-006; `research.md` D3 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | tellme has a single CLI end and no OpenAPI/HTTP surface; a test-harness change adds none. | `spec.md` A1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted or in-memory state model changes; per-scenario `TELL_ME_HOME`/`HOME` isolation is unchanged. | `spec.md` Edge Cases |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` + `**/dsl.md` | No acceptance/DSL semantics change: the round alters *how* the contract is executed, not *what* it asserts. | `spec.md` A1 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0024-e2e-suite-throughput.md` (+ the `docs/decisions/README.md` index row) | Records the parallelism default + the timing-protection decision (Q1), the subset selector + the subset-never-the-gate invariant (Q2), and the measurable bar (Q3) + a §Forward. | `spec.md` A3; `research.md` D1–D5 |
