# Truth Delta: 056-mcp-tool-call-reason

**Plan Package**: `specs/plans/056-mcp-tool-call-reason`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 OPEN** — **Q1** (how `reason` is elicited in the offered declaration) · **Q2** (un-wrapped / legacy call disposition). Owner rows below are **expected** (not yet recorded); the truth-owner skills fill them in their phases.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/techstack.md` — **MCP client** row | **Round 056**: the model-visible MCP tool declaration requests a tellme-owned `reason`; a call is `{reason, MCP_PAYLOAD}`; only `MCP_PAYLOAD` is forwarded; the server's advertised schema is relayed verbatim. | `spec.md` S-1…S-6 / FR-001…FR-006 |
| MODIFY (expected) | `specs/truth/techstack.md` — **Agent tool loop** row | **Round 056**: the existing single-owned reason path now reaches MCP calls (the reason is tellme's field, never forwarded to the server). | `spec.md` FR-002 / I-2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/` (**no `contracts/**`**) | tellme exposes a single CLI end; no OpenAPI/HTTP surface changes. | `spec.md` A1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (expected) | `specs/truth/data/data-model.dbml` | No persisted or in-memory state model changes; the MCP call shape is transient. | `spec.md` Edge Cases |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `chat/dsl.md`) | **Round 056**: a new Rule/Example — an MCP tool call renders a `[Tool Reason]` row, and the server receives only its payload. | `spec.md` US1/US2 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD (expected) | `docs/decisions/0025-mcp-tool-call-reason.md` (+ the `docs/decisions/README.md` index row) | Records the tellme-owned `reason` envelope for MCP calls, the untouched-server-definition invariant, the elicitation mechanism (no prompt change), and a §Forward. | `spec.md` A3 |
