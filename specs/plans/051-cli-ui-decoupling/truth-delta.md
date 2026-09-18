# Truth Delta: 051-cli-ui-decoupling

**Plan Package**: `specs/plans/051-cli-ui-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: `/axb-technical-research` (D1–D12 + **ADR 0020** + the `techstack.md` MODIFY rows below) has **landed**. **Clarify round 1 CLOSED**: **Q1 → D** (one round closes #101) · **Q2 → (i)** (value types → existing domain peers) · **Q3 → (ii)** (narrow lifecycle-grouped ports) · **Q4 → (A)** (F-8 interface `OutputSink`; F-7 named `Discovery`; F-6 narrow seams).
>
> **Round shape**: the **terminal R5 slice** — the last RULE-E residual `internal/cli → internal/ui` removed (**baseline 1 → 0**, **two** removals: the RULE-E line + its RULE-F key) + **F-6/F-7/F-8** resolved. [#101](https://github.com/gosharplite/tellme/issues/101) closes on delivery.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | **Round 051 / ADR 0020**: the RULE-E baseline figure goes **1 → 0** (the **terminal state** — the last edge `internal/cli → internal/ui` is removed). The **RULE-F** `couplingSurface` table is now **EMPTY** (its last key removed); a future governed application edge MUST be surface-tracked. Records the terminal state + "no release valve at 0". | `truth-current`; round 051 FR-009; `research.md` D8/D9. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | Notes that the CLI's presentation now flows through injected **domain ports** (`render.Lines` for the status/tail lines, `render.Indicator` + `deps.NewProgress` for the spinner + `[Tool Output]` coordinator, deps factories for the answer renderer + the loop's `ToolLineRenderer`), so `internal/cli` names no `internal/ui` type. Behaviour unchanged; the bytes/caps/sanitize policy stays single-owned in `internal/ui`. | round 051 FR-001/FR-002/FR-009; `research.md` D3/D4. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the `verify` aggregate already names `verify-architecture`. The round adds **no** Makefile target — the gate rides that target. | round 051 NFR (no new target). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and no OpenAPI/HTTP surface. The round authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`: no persisted/runtime-state change (a rendering de-coupling + seam refactors). | `spec.md` A5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/**/dsl.md` | Inspected: a `cli`-side rendering de-coupling is not a tellme CLI-contract change (no Rule/Example/step/`DSLRow` change). | `spec.md` A2/A6; the rounds 020/031/036/041–050 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0020-cli-ui-decoupling.md` (+ the `docs/decisions/README.md` index row) | Records the `→ ui` de-coupling (value-type homes + the ports/factories), the F-6/F-7/F-8 resolutions, the **two ratchet removals** (baseline **1 → 0** + the RULE-F key), what the change is *not*, and its relation to ADR 0006/0008/0013/0015/0016/0017/0018/0019. | round 051 FR-009; `research.md` D8 |

## Fold review (PR [#114](https://github.com/gosharplite/tellme/pull/114), review `5729783315`)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row (fold) | Records **R-51-1**: the `-update-baseline` path now **refuses to grow** the baseline (enforcement, not policy) — the terminal state's valve is a guarded two-edit path, not an absence. **RF-51-5**: at baseline 0 the RULE-F coverage clause is load-bearing on that guarded `-update-baseline` path. | round-051 fold R-51-1/RF-51-5. |
