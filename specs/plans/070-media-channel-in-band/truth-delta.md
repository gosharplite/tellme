# Truth Delta: 070-media-channel-in-band

**Plan Package**: `specs/plans/070-media-channel-in-band`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify not escalated by default (0 questions)** — the goal is unambiguous; the shape choice defers to `/axb-technical-research` (S-1). No truth owner has run yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — *Image filesystem tool (`read_image`)* | expected MODIFY: the media **channel** is now **in-band** (returned from `Execute` via a `domain/tools` value; the loop translates); the per-call `context` collector (ADR 0032 D7a) is removed | `spec.md` US1/FR-001…FR-004 |
| _(pending)_ | `specs/truth/techstack.md` — *Agent tool loop* (or the tool-contract row) | expected MODIFY if shape (a): the `Tool.Execute` contract carries media | `spec.md` S-1/I-4 |
| _(pending)_ | `docs/decisions/00NN-*.md` (+ index) | expected ADD: a new ADR recording the in-band media channel; **annotate ADR 0032 `RF-062-10` as fully delivered** and **ADR 0039 `RF-069-1` as delivered**; the rejected shape recorded as a **settled rejection** (S-4) | `spec.md` SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | anchor #142 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/data/**` | No persisted-record change (I-1); media is never persisted (ADR 0032). | `spec.md` I-1 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/features/cli/**` | No user-visible behaviour change (the 042/043/047/049/069 structural-round precedent); `reading-a-local-image.feature` stays green **unchanged**. | `spec.md` US2/SC-002 |
