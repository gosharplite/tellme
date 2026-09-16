# Truth Delta: 032-mcp-client

**Plan Package**: `specs/plans/032-mcp-client`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` -> `### MCP Client` (new category) | Added a **MCP Client** category (CLI Application): the protocol library `github.com/modelcontextprotocol/go-sdk` **v1.7.0** **confined** to `internal/infrastructure/mcp/` behind the `tools.MCPClient` domain port (a `verify-mcp-sdk-confinement` gate); the typed `MCP_SERVERS` registry + its deterministic validation; credential resolution by auth mode (`auto`/`gh`/`bearer`/`basic`/`none`, token never logged); and the **non-stall** tool discovery (a small fixed fast-fail bound + per-server `ENABLED`, prompt-path-only, deterministic `mcp_<server>_<tool>` naming). Moved **"MCP client SDK"** out of *Not Introduced Yet*. | Round-032 Decisions 1–7: tellme gains a remote (Streamable HTTP) MCP client; the truth must reflect the current system (`truth-current`). |
| MODIFY | `specs/truth/techstack.md` -> `### Testing & Verification` / `### Not Introduced Yet` | Added a **Local fake MCP server** test edge (`httptest`, hermetic; a "never answers" fake witnesses the fast-fail bound; a manual live check is a closeout step, not a gate). Added *Not Introduced Yet* bullets for the deferred **local stdio transport**, **cross-invocation tool caching**, and **MCP-backed MEMORY/PLUR**; annotated the round-047 tool-call-concurrency note as unchanged despite the new MCP client. | Round-032 Decisions 1/3/7: record the test surface and the explicitly deferred MCP sub-capabilities (Q1/Q2/Q5). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _placeholder_ | `specs/truth/contracts/**` | To be filled by `/axb-api-plan` (expected NOOP — tellme has a single CLI end). | Truth owner fills during the round. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _placeholder_ | `specs/truth/data/**` | To be filled by `/axb-data-plan` (expected NOOP — no persisted state this round; caching deferred). | Truth owner fills during the round. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _placeholder_ | `specs/truth/features/cli/**` | To be filled by `/axb-dsl-refine` (expected ADD — the executable MCP capability interface truth). | Truth owner fills during the round. |
