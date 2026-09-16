# Truth Delta: 035-spinner-tail-residue

**Plan Package**: `specs/plans/035-spinner-tail-residue`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/techstack.md` — **Turn progress spinner (operator)** row | Expected: note that the per-call tail is written with the indicator yielded (the tail no longer shares a row with a live frame); the row's existing "yields the line to the answer / leaves no residue" intent holds for the tool-tail path. NOOP if the row already covers it. | Issue #72; `spec.md` FR-001/FR-002. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | Expected NOOP (checked): tellme has a single CLI end and no OpenAPI/HTTP surface; this round touches only a `stderr` diagnostic write. | `spec.md` FR-006. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/data-model.dbml` | Expected NOOP (checked): no persisted state changes — the fix alters only when/how the tail is written to the terminal. | `spec.md` FR-005/FR-006. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (+ `chat/dsl.md` if a row is needed) | Expected MODIFY: add/re-anchor a carrier for the **gated tool-tail** journey (diagnostics at a terminal + a scripted tool round) so the existing residue rule is exercised on the tool-tail path; re-use existing rows where possible. | Issue #72; `spec.md` FR-001/FR-002/FR-004. |
