# Truth Delta: 037-interactive-prompt-no-selection

**Plan Package**: `specs/plans/037-interactive-prompt-no-selection`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md` (the interactive-prompt / suggestions row: the suggestion selection starts at **no choice** and resets on refresh, aligned to the reference).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (the at-rest `Then` flips from *marks one suggestion as the current choice* to *marks no suggestion as the current choice*) + **MODIFY** the matching row in `specs/truth/features/cli/chat/dsl.md`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` | — | `spec.md` FR-001/FR-002. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/contracts/**` | — | `spec.md` FR-005/FR-006. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/data-model.dbml` | — | `spec.md` FR-005/FR-006. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` + `chat/dsl.md` | — | `spec.md` FR-001/FR-003. |
