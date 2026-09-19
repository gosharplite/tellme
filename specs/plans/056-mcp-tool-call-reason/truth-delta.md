# Truth Delta: 056-mcp-tool-call-reason

**Plan Package**: `specs/plans/056-mcp-tool-call-reason`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 CLOSED** — **Q1 → A** (the offered envelope: required top-level `reason` + `MCP_PAYLOAD` carrying the server schema verbatim) · **Q2 → B, STRICT** (a non-envelope MCP call is refused; the server is never contacted). **Scope extended (operator, decision (ii)):** the refusal generalises to a **universal, single-owned *no reason, no go* gate** for every call (native **and** MCP); **[#121](https://github.com/gosharplite/tellme/issues/121) folded into this round** (closed as superseded). Owner rows below are **expected** (not yet recorded); the truth-owner skills fill them in their phases. |

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/techstack.md` — **MCP client** row | **Round 056**: the model-visible MCP tool declaration requests a tellme-owned `reason`; a call is `{reason, MCP_PAYLOAD}`; only `MCP_PAYLOAD` is forwarded; the server's advertised schema is relayed verbatim. | `spec.md` S-1…S-6 / FR-001…FR-006 |
| MODIFY (expected) | `specs/truth/techstack.md` — **Agent tool loop** row | **Round 056**: (a) the existing single-owned reason path now reaches MCP calls; (b) a **universal, single-owned *no reason, no go* gate** — a call (native **or** MCP) whose top-level `reason` is missing/blank is refused (recoverable retry result; the tool does not execute). | `spec.md` FR-002 / FR-008 / FR-009 / I-2 / I-5 |
| MODIFY (expected) | `specs/truth/techstack.md` — **Agent tool schemas** row | **Round 056**: clarity note — the mandatory `reason` is now **enforced**, not declaration-only (the `resourceSchema` declaration is unchanged). | `spec.md` S-5 / FR-008 |

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
| MODIFY (expected) | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `chat/dsl.md`) | **Round 056**: a new Rule/Example — an MCP tool call renders a `[Tool Reason]` row, and the server receives only its payload; an envelope-less MCP call is refused. | `spec.md` US1/US2 |
| MODIFY (expected) | `specs/truth/features/cli/chat/**` (a new Rule) + `chat/dsl.md` | **Round 056**: a Rule/Example for the **universal** gate — a native tool call with no/blank `reason` is refused (recoverable retry), a conforming call renders its row; the round-022 schema-nonconforming `read_files`-without-reason fixture is updated. | `spec.md` US3 / FR-008 / FR-010 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD (expected) | `docs/decisions/0025-mcp-tool-call-reason.md` (+ the `docs/decisions/README.md` index row) | Records the tellme-owned `reason` envelope for MCP calls, the untouched-server-definition invariant, the elicitation mechanism (no prompt change), and a §Forward. | `spec.md` A3 |
