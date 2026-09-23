# Truth Delta: 082-listing-backward-turn-indices

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-23)** — plan package + `spec.md` + checklist; clarify **not escalated (0 questions)** (the theme is anchored by issue [#165](https://github.com/gosharplite/tellme/issues/165); the issue's *Proposed Implementation Touch Points* are routed to `/axb-technical-research`). The truth-owner sections below are **skeletons**, to be filled at the corresponding phase.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | — | — | — |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/contracts/**` | _(expected NOOP — single CLI end; no OpenAPI/HTTP surface)_ | `spec.md` §Grounding |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/**` | _(expected NOOP — no persisted-state shape change; `history_entry`/`history_step` DBML unchanged)_ | `spec.md` FR-008 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/history/**` | _(expected MODIFY — the `-l` listing header Rules/Examples + `history/dsl.md` rows gain the `- N` suffix)_ | `spec.md` US1/US2, FR-001…FR-006 |
