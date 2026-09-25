# ADR 0059 — MCP cache: restore header-routed argument calls and move the cache to the per-mode workspace

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue #182)
- **Related:** **ADRs 0058** (the cross-invocation MCP tool cache this repairs — supersedes its **D2**, qualifies its **D4**), **0031** (the `x-mcp-header` measurement: 45 tools, 37 annotated), **0032** (MCP client), **0041** (the load-bearing domain model), **0006** (upstream — no norm without a witness); issue [#182](https://github.com/gosharplite/tellme/issues/182); MCP Go SDK v1.7.0 (`mcp/client.go` `lookupTool`, `mcp/streamable_headers.go` `generateParamHeaders`); round 088 (`specs/plans/088-mcp-cache-header-routing-and-mode-location`)

## Context

Round 087 (ADR 0058) added a cross-invocation MCP tool cache: a warm, fresh entry is served with
**zero** MCP connections, and a cached tool binds a **lazy client** that defers connect to the first
`CallTool`. Two defects shipped:

1. **Header routing (severe).** The MCP Go SDK v1.7.0 implements the `x-mcp-header` extension: when a
   tool's `inputSchema` annotates a property, `CallTool` must move that argument onto the HTTP header
   `Mcp-Param-<Name>`. The SDK resolves the annotation from its **`tools/list` cache** —
   `ClientSession.lookupTool` (*"the most recently seen definition of the tool … across all cached
   ListTools results … used by CallTool to inject the tool definition"*). Round 087's
   `lazyClient.CallTool` **connects only** and never issues `tools/list`, so the session cache is
   empty, no header is emitted, and a server that requires one rejects the call. Measured live: the
   **GitHub MCP server** (`api.githubcopilot.com/mcp/`, 37/45 tools annotated) fails every
   parameter-bearing call with `400 missing Mcp-Param-owner header`; only param-less tools work. The
   pre-087 live path issued `ListTools` on the same session, so it worked — the warm cache silently
   disabled header routing.
2. **Location (design).** The cache was written to `$TELL_ME_HOME/mcp-toolcache.json` (ADR 0058
   **D2**). Every other per-mode artifact lives under `output/<mode>/`; the MCP server set is
   per-mode config, so the cache is mode-scoped knowledge.

Both shipped because of a **witness gap**: the hermetic fake MCP server did not require a
header-routed argument, so no test reddened while the real endpoint 400'd (ADR 0006's class).

## Decision

**D1 — Restore header routing by warming the SDK `tools/list` cache on the first cached call.** The
SDK reads the annotation only from its own `tools/list` result, so the lazy client issues **one
`ListTools`** on the client session before delegating the first `CallTool`. (Verified against the
SDK source; there is no way to replay a schema through the domain port.)

**D2 — Warm-up placement and policy.** The warm-up runs **once**, lazily, inside the first
`CallTool` (after `connect`), bounded by the **call timeout**, and is **best-effort** — a warm-up
failure is ignored and the call is still attempted (a header-routed server then fails recoverably,
as pre-fix). A cache hit that runs **no** tool still makes **zero** connections (round 087 FR-001
preserved): connect/warm happen only inside `CallTool`.

**D3 — The cache lives in the per-mode workspace.** `NewFileToolCache(workspace)` →
`$TELL_ME_HOME/output/<mode>/mcp-toolcache.json`. **This supersedes ADR 0058 D2.** The
`deps.Dependencies.MCPDiscoverer` seam takes the resolved **workspace** instead of the home; the
call site (`internal/cli` `augmentRegistryWithMCP`) already holds `res.Workspace`.

**D4 — `--new` still does not clear the cache.** `--new` archives only the known session files; the
per-mode cache file is untouched (unchanged from round 087).

**D5 — Witness the gap (the round's core obligation).** The hermetic fake MCP server
(`internal/infrastructure/mcp/mcptest/`, SDK server API) already **enforces** `Mcp-Param-*`; an
`Options.HeaderRouted` field advertises a tool whose named argument carries `x-mcp-header`, so the
fake reproduces the failure automatically. Carriers: an **E2E** Example (a warm cache + a
header-routed tool + a call ⇒ the tool is reached and the turn completes) and a **unit pin** that
the lazy client issued a `tools/list` before calling. The pre-fix lazy client reddens both.

**D6 — No other observable change.** The discovery bound, the cache key/TTL/refresh, the offered
set/order, the offered (vendor-extension-stripped) schemas, the naming/envelope/call-time-error
contracts, and the offline-path network-freedom are unchanged; **no credential** is persisted;
stdlib-only; POSIX-only.

**D7 — Records.** ADR 0059 (+ index); ADR 0058 `Status`/D2/D4 gain a back-pointer; `techstack.md`
(the *MCP tool cache* row — location + warm-up; the *MCP tool discovery* row); `data-model.dbml`
(the entry's location); the `chat` feature + `dsl.md`; `docs/domain-model/**` **NOT** modelled
(ADR 0041 escape hatch).

## Consequences

- A warm-cached tool once again reaches a server that routes arguments as headers — the GitHub MCP
  server is usable again (issue #182).
- The steady-state zero-dial guarantee holds: a cache hit that runs no tool contacts nothing; a tool
  that **is** called pays one extra `tools/list` (the live path already did) — a small, bounded,
  call-time cost, recorded.
- The cache is now mode-scoped and consistent with the other per-mode artifacts; `--new` still
  preserves it.
- The `x-mcp-header` regression now has a **witness** (the fake server enforces the header), closing
  the gap that let it ship.
- No new class phrase, no new exit code; `go.mod`/`go.sum` unchanged.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Persist the raw annotated schema and replay it into the SDK | Impossible via the `tools.MCPClient` domain port; the SDK reads only its own `tools/list` cache. |
| Issue `tools/list` at discovery time on the cached path | Re-introduces a **prelude dial** every run — defeats round 087's purpose. |
| Revert the cache (round 087) | The cache is the right shape; only the two defects are wrong. |
| Warm **eagerly when the cached tools are assembled** (`DiscoverCached`) — an unconditional prelude dial | Violates the zero-dial guarantee (FR-002): a no-tool hit would dial. (Warming inside `connect` is *not* this — `connect` is reached only from `CallTool`, so it cannot dial on a no-tool hit.) |

## Forward (non-blocking)

> **⚠ Not open work.** A decision deferred to a trigger, or a recorded divergence — not tasking.

- **RF-088-1** — the warm-up adds one `tools/list` round-trip on the first cached call; a
  per-session memo is not persisted across processes (each run warms once if it calls a tool).
- **RF-088-2** — a warm-up failure is best-effort; a header-routed server then fails recoverably
  rather than retrying the warm-up.
- **RF-088-3** — the offered (cached) schemas remain the vendor-extension-stripped ones (round 061
  floor); the SDK recovers the annotation from its own `tools/list`, not from the offered schema.
- **RF-088-4** — the cache path is fixed to `output/<mode>/`; a future multi-mode shared cache would
  be a separate decision.
- **RF-088-5** — the warm-up is charged to the **call's own deadline** (`agentloop` wraps each tool call
  in its per-call timeout, and that ctx is passed into `CallTool`), so a slow `tools/list` can consume the
  budget and tip the first cached call into the structural `timeout`/"stopped" path — a small regression
  vs pre-087, where listing ran under the separate 3 s discovery bound. A mitigation (bound the warm-up by
  its own deadline, e.g. `min(mcpDiscoveryBound, remaining)`) is deferred.
