# Truth Delta: 056-mcp-tool-call-reason

**Plan Package**: `specs/plans/056-mcp-tool-call-reason`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 CLOSED** — **Q1 → A** (the offered envelope: required top-level `reason` + `MCP_PAYLOAD` carrying the server schema verbatim) · **Q2 → B, STRICT** (a non-envelope MCP call is refused; the server is never contacted). **Scope extended (operator, decision (ii)):** the refusal generalises to a **universal, single-owned *no reason, no go* gate** for every call (native **and** MCP); **[#121](https://github.com/gosharplite/tellme/issues/121) folded into this round** (closed as superseded). **Owner rows recorded:** `/axb-technical-research` = 3 MODIFY (recorded) · governance = **ADR 0025** ADD (recorded); `/axb-api-plan` + `/axb-data-plan` = NOOP (recorded); `/axb-dsl-refine` = 1 MODIFY + 1 ADD + 1 MODIFY (recorded) + 1 **fold** MODIFY (recorded). |

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **MCP tool-call reason envelope** (a **new** MCP Client row) | **Round 056**: an MCP tool is offered with tellme's own declaration — a required `reason` + `MCP_PAYLOAD` carrying the server's advertised schema **verbatim** (never mutated); a call forwards **only** `MCP_PAYLOAD`; absent payload (with a reason) = `{}`; a shape violation is refused (server not contacted); the adapter adds only the envelope check. | `spec.md` S-1…S-6 / FR-001…FR-006; `research.md` D1/D3 |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | **Round 056**: the loop is the single enforcement point of the **universal *no reason, no go* gate** — a call whose reason does not render is refused (recoverable nil-error result; the tool does not execute); the gate **reuses** the round-046 `ReasonLine` owner (no new predicate); a nil `Lines` renderer means no gate. | `spec.md` FR-002 / FR-008 / FR-009 / I-5; `research.md` D2/D4 |
| MODIFY | `specs/truth/techstack.md` — **Agent tool schemas** row | **Round 056**: the declared mandatory `reason` is now **enforced**, not declaration-only; the declared schemas are unchanged. | `spec.md` S-5 / FR-008; `research.md` D5 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | tellme exposes a single CLI end; no OpenAPI/HTTP surface changes. | `spec.md` A1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted or in-memory state model changes; the MCP call shape is transient. | `spec.md` Edge Cases |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `chat/dsl.md`) | **Round 056**: a new Rule/Example — an MCP tool call renders a `[Tool Reason]` row, and the server receives only its payload; an envelope-less MCP call is refused. | `spec.md` US1/US2 / FR-001…FR-006 |
| ADD + MODIFY | `specs/truth/features/cli/chat/requiring-a-reason-to-call-a-tool.feature` (ADD) + `chat/watching-the-tool-loop.feature` (MODIFY — one obsolete Example removed) + `chat/dsl.md` | **Round 056**: a Rule/Example for the **universal** gate — a native tool call with no/blank `reason` is refused (recoverable retry), a conforming call renders its row; the round-022 schema-nonconforming `read_files`-without-reason fixture is updated. | `spec.md` US3 / FR-008 / FR-010 |

| MODIFY (fold) | `specs/truth/features/cli/chat/requiring-a-reason-to-call-a-tool.feature` (+ `chat/dsl.md`) | **Round 056 fold (PR [#122](https://github.com/gosharplite/tellme/pull/122) review TD-056-1):** a new **Rule/Example** for a shape-violating MCP call (`the MCP server "shop" received no call`) and a `received exactly one call` Then on the reasonless Example — the force-bearing carriers for SC-006/FR-003 at the server boundary; three new `dsl.md` rows. | `spec.md` SC-006 / I-4; PR [#122](https://github.com/gosharplite/tellme/pull/122) review TD-056-1 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0025-mcp-tool-call-reason.md` (+ the `docs/decisions/README.md` index row) | Records the tellme-owned `reason` envelope for MCP calls (server schema relayed verbatim), the **universal, single-owned *no reason, no go* gate** (the [#121](https://github.com/gosharplite/tellme/issues/121) fold), and a §Forward. | `spec.md` A3; `research.md` D1–D6 |
