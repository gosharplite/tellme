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
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the MCP client authors no request/response contract of tellme's own (the remote MCP server is an outbound dependency, not an exposed end). | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked — the round persists no new state: MCP tool lists are discovered live each run (caching deferred, Q2 Option 2); the existing persisted shapes (`history.jsonl`, `tokens.log`, the prompt logs) are unchanged. The `MCP_SERVERS` block is configuration, not persisted system state. | `data-model-covers-all-state` holds — no new state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` -> Rules (a server's tool is offered · offered **alongside** the native tools · a prompt is answered by calling it · an unreachable server is skipped without stalling · a reachable server still works while another is down · several unresponsive servers still do not add up to a long wait · a server marked off is never contacted · a malformed schema is not offered · a failed tool call does not abort the run · the resolved token is sent) | Added the executable MCP interface truth: on a prompt-bearing turn a configured, enabled remote MCP server's tools are offered **alongside** the native tools under the deterministic namespaced name `mcp_<server>_<tool>`; a prompt may call one and its result is fed back; discovery is **non-stall** (a server that never answers is skipped with a single-sourced warning; several unresponsive servers still answer); a server marked `ENABLED: false` is never contacted; a **malformed schema is not offered** (B1); a **failed tool call (tool-level or transport) does not abort the run** (TD1); a resolved token is sent. | Round-032 US1/US2/US3 + FR-018/FR-019: the MCP capability's observable CLI behaviour is executable truth (`acceptance-coverage`); the review folds B1/TD1 add carriers. |
| ADD | `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 032)` (8 rows) + `## Then (round 032)` (8 rows) | Added the module DSL rows for the new MCP Given/Then sentences (configure a reachable / never-answering / off / token-protected / malformed-schema / tool-error / transport-failing fake MCP server; a provider that asks for the namespaced MCP tool; offered-namespaced / offered-**alongside** / offered-none / called / unreachable-warning / never-contacted / token-received / **the run continued past the failed MCP tool call**). | Round-032: every new Gherkin step needs exactly one authoritative DSL row (`dsl-exact-one-match`); the topology audit PASSED (43 features · 289 module rows · 1498 steps). |
