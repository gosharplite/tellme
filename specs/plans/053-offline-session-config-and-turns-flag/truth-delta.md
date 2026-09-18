# Truth Delta: 053-offline-session-config-and-turns-flag

**Plan Package**: `specs/plans/053-offline-session-config-and-turns-flag`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 OPEN** (Q1 = `-t` carrier/format; Q2 = explicit-`-c` load-failure policy). No truth has been written yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — CLI session commands / offline reporting rows | _to be recorded by `/axb-technical-research`_ | `spec.md` US1/US2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected NOOP)_ | `specs/truth/` (**no `contracts/**`**) | tellme has a single CLI end and no OpenAPI/HTTP surface. | `spec.md` A4 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected NOOP)_ | `specs/truth/data/data-model.dbml` | The `-t` carrier read is read-only; no persisted-schema change. | `spec.md` A4 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected MODIFY)_ | `specs/truth/features/cli/**` and `specs/truth/features/cli/**/dsl.md` | A user-facing session-selection correction + a new `-t` flag likely add/align a CLI Example + `DSLRow`. | `spec.md` A5 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `docs/decisions/00NN-*.md` (+ the `docs/decisions/README.md` index row) | The round's decision record (mode-resolution precedence; the `-t` carrier; the explicit-`-c` policy). | `spec.md` FR-008; `research.md` |
