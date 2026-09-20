# Truth Delta: 072-gemini-cached-token-usage

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify not escalated by default (0 questions)** — the defect and fix are unambiguous; the residual choice (the `thoughtsTokenCount` disjointness rule) defers to `/axb-technical-research` (S-4). No truth owner has run yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — *Gemini/Vertex adapter* row | expected MODIFY: the adapter decodes `cachedContentTokenCount` into `Usage.CachedTokens` and `thoughtsTokenCount` into `Usage.ThinkingTokens` | `spec.md` US1/US2, FR-001/FR-002 |
| _(pending)_ | `specs/truth/techstack.md` — the usage/`UsageRecord` accounting row | expected MODIFY: the Gemini family reports `H`/`Th` (previously always 0) | `spec.md` US3, FR-007 |
| _(pending)_ | `docs/decisions/00NN-*.md` (+ index) | expected ADD: the Gemini usage decode + the pinned `thoughtsTokenCount` disjointness rule (S-4) | `spec.md` SC-002/SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | anchor [#149](https://github.com/gosharplite/tellme/issues/149) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/**` | check: the persisted `UsageRecord` shape (`tokens.log`) is unchanged; only the **values** (`cached`/`thinking`) become correct | `spec.md` I-5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/chat/**` + `chat/dsl.md` | expected ADD/MODIFY: a metrics Example asserting the cached/thinking figures on a Gemini turn | `spec.md` US1/US2 |
