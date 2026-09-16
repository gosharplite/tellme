# System Analysis Plan — round 032 (`032-mcp-client`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/032-mcp-client/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (3 journeys: use a server's tool · non-stall · ENABLED)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done (MCP Client category + test edge)
└── features/cli/**                # /axb-dsl-refine — ADD (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact — the MCP surface is model-facing and its CLI behaviour is carried by the CLI end.)*

### Repository structure (root)

```text
internal/domain/tools/mcp_client.go        # ADDED   — the tools.MCPClient domain port (ListTools / CallTool / Close) + MCPToolDefinition; network-free, protocol-free
internal/infrastructure/mcp/                # ADDED   — the remote (Streamable HTTP) adapter over github.com/modelcontextprotocol/go-sdk; credential resolution (auth modes); the ONLY package allowed to import the SDK
internal/config/                            # CHANGED — typed MCP_SERVERS registry (MCPServerConfig) + deterministic validation (key format, URL required / COMMAND rejected, auth mode + credentials, TIMEOUT ≥ 0, ENABLED default true)
internal/infrastructure/di/                 # CHANGED — an MCP client factory (builds a client per enabled server; injectable token resolver; memoized gh/token)
internal/cli/cli.go                         # CHANGED — on the prompt-bearing path, discover enabled servers under the fixed fast-fail bound, register their tools into the tool registry (deterministic mcp_<server>_<tool> names), and thread the MCP tools through the existing tool loop
internal/*/mcp_discovery*.go                # ADDED   — the concurrent, fixed-bound discovery + sorted registration (namespacing)
Makefile                                    # CHANGED — add `verify-mcp-sdk-confinement` (SDK import confined to internal/infrastructure/mcp/) and wire it into the gates
tests/e2e/steps/*, suite, (fake MCP server) # ADDED   — a hermetic fake MCP server (httptest) + steps: use a tool; a "never answers" server witnesses the fast-fail bound; ENABLED skip
specs/truth/techstack.md                    # MODIFY  — MCP Client category + Local fake MCP server edge + Not-Introduced-Yet bullets ✓ done
specs/truth/features/cli/**                 # ADD     — the executable MCP interface truth (/axb-dsl-refine, contract owner)
go.mod / go.sum                             # CHANGED — add github.com/modelcontextprotocol/go-sdk (v1.7.0)
```

**Structure Decision**: Round 032 adds the **remote MCP client** capability *inside* the existing CLI end —
it adds **no** new system boundary. A new domain port (`tools.MCPClient`) and a new adapter package
(`internal/infrastructure/mcp`) isolate the MCP protocol library behind an injection seam (the same
discipline used for the provider transports), and the SDK is **confined** to that adapter by a Makefile
gate. The MCP tools flow through the **unchanged** `internal/domain/tools.Tool` port and the **unchanged**
agent tool loop, so there is **no** change to the native tool surface or the sequential execution model.
The round **does change the CLI end's observable behaviour** (a prompt run now offers and can call the
tools of a configured remote MCP server) and it persists **no** new state (`/axb-data-plan` `NOOP`;
caching is deferred). It reaches an **external dependency** (the remote MCP server) — an *outbound*
endpoint tellme consumes, exactly as it consumes provider endpoints — which is **not** a new system
interface tellme exposes. Consistent with `research.md` Decisions 1–7.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the agent tool surface a prompt run
offers and may call). The round **changes the CLI end's observable behaviour**: on a prompt-bearing turn,
tellme now offers the tools of each configured, enabled remote MCP server (under deterministic namespaced
names) and can call them, and it must do so **without ever stalling** on an unreachable server.

1. `CLI end — the agent tool surface (incl. remote MCP tools)`
   - 端點類型: `CLI 端點`
   - 主要介面: the offered tool set a prompt run presents to the model (now including each enabled remote MCP server's tools), the namespacing/order of those tools, the discovery non-stall bound + `ENABLED` behaviour, the MCP tool-call outcome, and the `MCP_SERVERS` configuration surface + its validation.
   - 需求原文依據: spec US1 (FR-001…007 — load/validate `MCP_SERVERS`, discover + offer + call a server's tools), US2 (FR-008…010 — a server that does not answer never stalls a run), US3 (FR-011…012 — a server marked `ENABLED: false` is never contacted).

> **Scope notes**:
> - The **remote MCP server** is an **outbound external dependency** tellme consumes (like a provider endpoint), **not** a system interface tellme exposes; its integration decisions live in `research.md` (transport, SDK, auth, non-stall discovery) and `specs/truth/techstack.md` — there is **no** analysis planner for it, and it is not invented as an end.
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the MCP client authors no request/response contract of tellme's own).
> - `/axb-data-plan` = **`NOOP`** (the round persists no new state — no cache; the existing persisted shapes are unchanged).
> - `/axb-dsl-refine` = **contract owner** (ADD the executable MCP interface truth under `specs/truth/features/cli/**` — offering a server's tools, calling one, the non-stall behaviour, the `ENABLED` switch — + the `chat/dsl.md` rows).
> - `/axb-ui-plan` = **skipped** (no UX-surface change; the operator `stderr` chrome is unchanged — the MCP integration adds no new visible chrome).
> - `/axb-spec-by-example` = **done** (3 acceptance journeys).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change; caching deferred).
3. **`/axb-dsl-refine`** — **contract owner**: ADD the MCP interface truth under `specs/truth/features/cli/**` — that a configured remote server's tools are offered to the model alongside the native tools and can be called (their results fed back); that an unreachable server is skipped within the fixed bound and never stalls the run; that a server marked off is never contacted — plus the matching `chat/dsl.md` rows. *(Exact file/rule/row names are `/axb-dsl-refine`'s call.)*

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX-surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/032-mcp-client`;
truth root `specs/truth`; truth-delta `specs/plans/032-mcp-client/truth-delta.md`;
interfaces: **1 (CLI end)** carried to `/axb-dsl-refine`; the round's delivery is the MCP domain port + the
confined SDK adapter + the typed `MCP_SERVERS` config + the fast-fail discovery/registration wired into the
prompt path, plus the CLI interface truth.

---

### Gating blockers

*(none — the operator locked the design in-session before `/axb-specify` (Q1 remote HTTP only; Q2 fast-fail
bound + `ENABLED`; Q3 official SDK confined; Q4 full auth parity + reference naming; Q5 MCP client only);
`research.md` Decisions 1–7 settle the transport, the SDK/domain-port confinement, the non-stall discovery,
the auth set, the naming, the config/validation, and the verification posture. No open decision gates the
round.)*
