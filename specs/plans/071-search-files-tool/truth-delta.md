# Truth Delta: 071-search-files-tool

**Plan Package**: `specs/plans/071-search-files-tool`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify` (2026-09-20). **Clarify ESCALATED** (Q1 search mode · Q2 scope/ignore · Q3 bound/shape — asked one at a time). No truth owner has run yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — *Read-only filesystem tools* (or a new reader row) | expected MODIFY/ADD: a ninth tool `search_files` — a bounded, deterministic in-file content search; its walk, skip policy (per Q2), bound (per Q3), and mode (per Q1) | `spec.md` US1/US2/US3, FR-001…FR-009 |
| _(pending)_ | `specs/truth/techstack.md` — *Agent tool schemas* | expected MODIFY: the ninth tool's schema (shared `resourceSchema`; `required ⊆ properties`) | `spec.md` FR-005/I-7 |
| _(pending)_ | `specs/truth/techstack.md` — *Tool-usage accounting* | expected MODIFY: the tool joins the recordable union | `spec.md` FR-006/I-6 |
| _(pending)_ | `docs/decisions/00NN-*.md` (+ index) | expected ADD: a new ADR recording the tool, the design-intent justification, and the three recorded divergences (no `SafePath` · no `WorkspacePolicy` · deterministic) | `spec.md` §Read first, SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | anchor [#147](https://github.com/gosharplite/tellme/issues/147) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/data/**` | No persisted-record change; the tool result is transient (in-flight only). | `spec.md` A1 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD (expected) | `specs/truth/features/cli/chat/**` + `chat/dsl.md` | a new executable journey for the tool (bounded, deterministic search; the "0 matches" and invalid-pattern cases) + the step rows | `spec.md` US1/US2/US3, FR-010 |
