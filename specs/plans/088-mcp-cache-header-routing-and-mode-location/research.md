# Technical research — round 088 `088-mcp-cache-header-routing-and-mode-location`

**Owner**: `specs/truth/techstack.md`. **Inputs**: `spec.md`, issue **#182**, the frozen round-087
package (`specs/plans/087-mcp-tool-cache`), `docs/decisions/0058-mcp-tool-cache.md`, and the MCP Go
SDK **v1.7.0** source.

## Problem restatement (verified)

- **Header routing.** The SDK implements `x-mcp-header`: `CallTool` must move an annotated argument
  onto the HTTP header `Mcp-Param-<Name>`. The annotation is resolved from the SDK's **`tools/list`
  cache**:

  - `mcp/client.go` `ClientSession.lookupTool(name)` — *"returns the most recently seen definition of
    the tool … across all cached ListTools results … It is used by CallTool to inject the tool
    definition into the outgoing request context for transport-layer features (e.g. x-mcp-header
    param annotations)"*.
  - `mcp/streamable_headers.go` `generateParamHeaders(tool, params)` reads the annotation from
    `tool.InputSchema` and builds `Mcp-Param-*`.

  So a session that never issued `tools/list` emits **no** header. Round 087's `lazyClient.CallTool`
  connects only → the header is missing → the server's `validateMcpHeaders` returns
  `CodeHeaderMismatch` (`streamable.go`) → `400 missing Mcp-Param-<name> header`. The pre-087 live
  path called `ListTools` on the same session, so it worked.
- **Location.** `NewFileToolCache(home)` → `$TELL_ME_HOME/mcp-toolcache.json` (ADR 0058 D2); every
  other per-mode artifact is under `output/<mode>/` (`internal/home/home.go`).

## Decisions

### D1 — Fix header routing by warming the SDK `tools/list` cache on the first cached call

The SDK reads the annotation **only** from its own `tools/list` result; there is no way to replay a
schema through the domain port. So the lazy client must issue **one `ListTools`** on the client
session before delegating the first `CallTool`. (Confirmed by source, not inferred.)

### D2 — Warm-up placement and policy

`lazyClient` warms **once**, lazily, on the **first `CallTool`** (after `connect`), bounded by the
**call timeout**; the warm-up is **best-effort** — a warm-up failure is ignored and the call is still
attempted (EC-001). This keeps **FR-002** (a cache hit that runs no tool makes **zero** connections):
`connect`/warm happen only inside `CallTool`.

```go
func (c *lazyClient) CallTool(ctx, name, args) (domaintools.MCPToolResult, error) {
    cl, err := c.connect(ctx)
    if err != nil { return recoverable(err) }
    c.warmToolsList(ctx, cl)   // best-effort: issue tools/list so x-mcp-header routing works
    return cl.CallTool(ctx, name, args)
}
```

### D3 — Location: the per-mode workspace

`NewFileToolCache(workspace)` → `<workspace>/mcp-toolcache.json` = `$TELL_ME_HOME/output/<mode>/…`.
The `MCPDiscoverer` seam takes the resolved **workspace** (`res.Workspace`) instead of `res.Home`
(the call site already holds `res` — `internal/cli/cli.go` `augmentRegistryWithMCP`). Rationale:
the MCP server set is per-mode config; consistency with `history.jsonl`/`turns.log`/`tokens.log`.

### D4 — `--new` still does not clear it

`--new` archives only the known session files; it never touches `output/<mode>/mcp-toolcache.json`.
Unchanged from round 087 (the placement change is orthogonal).

### D5 — The witness (the gap that let this ship)

The hermetic fake MCP server (`internal/infrastructure/mcp/mcptest/`, built with the SDK **server**
API) **enforces** `Mcp-Param-*`: its `validateMcpHeaders` rejects a call missing a required header.
So advertising a tool whose argument carries `x-mcp-header` makes the fake reproduce the real
failure **automatically**. Add an `mcptest.Options.HeaderRouted string` (the property name to
annotate) and:
- an **E2E** Example (warm cache + a header-routed tool + a call ⇒ success);
- a **unit pin** that the lazy client issued a `tools/list` (e.g. a fake client recording it) — the
  pre-fix lazy client reddens both.

### D6 — Rejected alternatives

| Alternative | Why rejected |
| --- | --- |
| Persist the raw annotated schema and replay it into the SDK | Not possible through the `tools.MCPClient` domain port; the SDK reads only its own `tools/list` cache. |
| Issue `tools/list` at discovery time on the cached path | Re-introduces a **prelude dial** every run — defeats round 087's whole purpose (FR-001). |
| Revert the cache (round 087) | The cache is the right shape; only the two defects are wrong. |
| Warm **eagerly when the cached tools are assembled** (`DiscoverCached`) — an unconditional prelude dial | Violates FR-002: a cache hit that runs no tool would dial. (Warming inside `connect` is *not* this: `connect` is itself reached only from `CallTool`, so warming there cannot dial on a no-tool hit.) |

### D7 — Records

**ADR 0059** — supersedes ADR 0058 **D2** (home-root placement → per-mode workspace) and qualifies
**D4** (the lazy connect must also warm `tools/list` before calling); ADR 0058 gets `Status`/
`§Forward` back-pointers. `techstack.md` MODIFY (the *MCP tool cache* + *MCP tool discovery* rows);
`data-model.dbml` MODIFY (the entry's location); the `chat` feature + `dsl.md`; `docs/domain-model/**`
**NOT** modelled.
