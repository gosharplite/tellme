# Truth Delta: 059-darwin-metrics-and-1hz-cadence

**Plan Package**: `specs/plans/059-darwin-metrics-and-1hz-cadence`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify CLOSED** (Q1–Q5 → 1, recorded in `spec.md` S-1…S-5). Owner rows are recorded per phase.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/techstack.md` — **System metrics provider (telemetry)** row | **Round 059 (ADR 0029):** macOS CPU lands (cgo Mach + !cgo `runtime/metrics` fallback — the recorded "pending a cgo mach sampler" gap closes); sysctl reads move to `golang.org/x/sys/unix`; MEM is the reference-exact per-platform definition; the 1 Hz sample cadence is recorded. | `spec.md` FR-001…FR-003, S-2/S-4 |
| MODIFY (expected) | `specs/truth/techstack.md` — **Go dependency set** (as applicable) | **Round 059:** add `golang.org/x/sys` (direct). | `spec.md` A2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface change. | `spec.md` A4 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted/in-memory domain-state change (the sample cache is spinner-local presentation state). | `spec.md` A4 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md`) | **Round 059:** the tool-phase resource Rule gains the 1 Hz refresh qualification (the figures refresh once per second; the wheel still animates). | `spec.md` FR-006/FR-007 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD (expected) | `docs/decisions/0029-darwin-metrics-and-sample-cadence.md` (+ index row) | Records the darwin sampler split (cgo Mach + !cgo `runtime/metrics`), the `x/sys` adoption, the reference-exact MEM definitions, and the 1 Hz sample throttle; §Forward RF-059-x. | `spec.md` A5 |
