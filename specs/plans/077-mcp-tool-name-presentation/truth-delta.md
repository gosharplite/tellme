# Truth Delta: 077-mcp-tool-name-presentation

**Plan Package**: `specs/plans/077-mcp-tool-name-presentation`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-22)** — the verify-first step (a live `tools/list` against the GitHub MCP server) is done; `research.md` D1–D8 + **ADR 0049** + `techstack.md` **ADD** (a *MCP tool-name presentation* row). Clarify resolved at specify time (**0 questions** — not escalated). `/axb-api-plan` + `/axb-data-plan` record `NOOP`. **Architect review 1 fold (`APPROVE WITH REQUIRED FOLDS`, F-1…F-4 + TD-1 + nits):** the folds added byte-exact/envelope-set/fallback-segment witnesses (F-1…F-3), added the PM-side acceptance journey `features/acceptance/discovering-the-callable-mcp-tool-name.feature` + reconciled the records (F-4), and made US2/FR-004 E2E-reachable via an `mcptest` empty-description knob (TD-1) — no truth *semantic* change beyond these.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — *MCP tool-name presentation (callable-name discoverability)* | a new row: the offered MCP declaration's **description** positively states the callable wire name (a tellme-authored **prefix** note naming `t.name`; the server's text relayed unchanged), and the empty-description **fallback names the callable name** — description-only, schema unchanged, no prompt change; a recorded divergence *beyond* the reference | `spec.md` US1/US2, FR-001…FR-006; `research.md` D1–D8 |
| ADD | `docs/decisions/0049-mcp-tool-name-discoverability.md` (+ index row) | the decision record: the verify-first result (45/45 GitHub tools have descriptions → fallback not in play), the call-name note, the fallback fix, the beyond-reference divergence, the declined alternatives | `spec.md` SC-003/SC-004; `research.md` |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state change (the offered-description text is computed, never stored). | `spec.md` 關鍵實體; `plan.md` §5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` | a new Rule with **two** Examples: the offered MCP declaration's **description** names the callable wire name (`mcp_<server>_<tool>`) — one for a description-bearing server, one for a server that ships **no** description (the fallback), observed on the fake-provider wire | `spec.md` US1/US2, FR-001…FR-006 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **+2** Then rows — the callable-name note, and the force-bearing empty-description fallback (via `mcptest.Options{EmptyDescription: true}`) | `spec.md` US1/US2 |
