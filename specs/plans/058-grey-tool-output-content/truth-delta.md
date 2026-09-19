# Truth Delta: 058-grey-tool-output-content

**Plan Package**: `specs/plans/058-grey-tool-output-content`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1** — **Q1 (the trailing partial line) is OPEN, non-blocking** (assumption A1: it is never printed and stays dropped). Owner rows are recorded per phase.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Turn chrome (operator)** row | **Round 058 (ADR 0028):** element (v) broadens from the `[Tool Output]` header + both separators to the whole **block** — the header, **every streamed content line**, and both separators grey. | `spec.md` US1 / FR-001…FR-006; `research.md` D1 |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | **Round 058:** the colour qualification now covers the streamed content lines (the sanitizer still governs the value; the grey wrap sits outside it). | `research.md` D2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface change. | `spec.md` A3 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Presentation-only; no persisted/in-memory state change. | `spec.md` A3 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (+ `chat/dsl.md`) | **Round 058 (ADR 0028):** the Rule *"the tool output frame is shown in grey"* broadens to the whole block (header + every content line + both separators); the `dsl.md` row's 必查 updated. | `spec.md` FR-001…FR-006 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0028-grey-tool-output-block.md` (+ index row) | Records the block-wide grey; **supersedes round 057's A3** (the round-057 package stays frozen); §Forward RF-058-x. | `spec.md` A4/A5; `research.md` D8 |
