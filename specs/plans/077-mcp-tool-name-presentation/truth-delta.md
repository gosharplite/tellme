# Truth Delta: 077-mcp-tool-name-presentation

**Plan Package**: `specs/plans/077-mcp-tool-name-presentation`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-specify` RUN (2026-09-22)** — the package + `spec.md` + this checklist + the truth-delta skeleton are created; clarify **not escalated (0 questions)**; the round is **research-gated (a NOOP verdict is acceptable)**. Owner rows are pending `/axb-technical-research` (the verify-first decision + a likely ADR + the MCP offered-declaration truth row) and `/axb-dsl-refine` (a fake-MCP-server carrier, iff a change ships).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — *MCP tool-call reason envelope* row (or a sibling offered-declaration row) | **pending** — record whether the offered MCP declaration makes the callable wire name discoverable (and how), or a verified NOOP | `spec.md` US1/US2, FR-001…FR-006 |
| *(pending)* | `docs/decisions/00NN-*.md` (+ index) | **pending** — a decision record for the presentation shape (or the recorded NOOP rationale) | `spec.md` SC-003/SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/` (**no `contracts/**`**) | **expected NOOP** — single CLI end; no OpenAPI/HTTP surface | `plan.md` (to be written) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/**` | **expected NOOP** — no persisted-state change (description text only) | `spec.md` 關鍵實體 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` (+ `dsl.md`) | **pending** — a Rule/Example observing the offered MCP declaration's name/description on the fake-provider wire (iff a change ships), else NOOP | `spec.md` US1/US2, FR-001…FR-006 |
