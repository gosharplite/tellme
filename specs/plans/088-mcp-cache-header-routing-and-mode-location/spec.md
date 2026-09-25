# Round 088 — `088-mcp-cache-header-routing-and-mode-location`

**Theme**: repair the two defects shipped by round 087's cross-invocation MCP tool cache —
(1) a **warm cache hit breaks any MCP server whose tools route arguments as HTTP headers**
(the whole GitHub MCP server: `missing Mcp-Param-<name> header`), because the lazy client
connects **without** populating the SDK's `tools/list` cache; and (2) the cache file is written
to the **`$TELL_ME_HOME` root** instead of the **per-mode workspace** (`output/<mode>/`).

**Anchor**: live issue **#182** (DoD = close it). Both defects are in the frozen round-087
package (**ADR 0058**) — this is a **fresh round** (`fresh-package-per-round`), never a re-open.

---

## 1. Why this round

Round 087 made the prompt prelude local in the steady state: a warm, fresh cache entry is served
with **zero** MCP connections, and a cached tool binds a **lazy client** that defers connect (and
credential resolution) to the first `CallTool`.

Two defects shipped with it:

1. **Header routing (severe).** The client SDK (MCP Go SDK v1.7.0) implements the `x-mcp-header`
   extension: when a tool's `inputSchema` annotates a property, `CallTool` must move that argument
   onto an HTTP header `Mcp-Param-<Name>`. The SDK resolves the annotation from its **`tools/list`
   cache** (`mcp/client.go` `lookupTool`: *"the most recently seen definition of the tool … across
   all cached ListTools results … used by CallTool to inject the tool definition"*). Round 087's
   `lazyClient.CallTool` **connects only** — it never calls `tools/list` — so the session cache is
   empty, no `Mcp-Param-*` header is emitted, and a server that requires one (GitHub: 37 of 45
   tools annotated, ADR 0031) rejects every parameter-bearing call. The pre-087 live path called
   `ListTools`, so it worked; the warm cache silently disabled it.
2. **Location (design).** `NewFileToolCache(home)` writes `$TELL_ME_HOME/mcp-toolcache.json`. Every
   other per-mode artifact lives under `output/<mode>/`; the MCP server set is per-mode config, so
   the cache is mode-scoped knowledge.

The reason both shipped is a **witness gap**: the hermetic fake MCP server did not require a
header-routed argument, so no test reddened while the real endpoint 400'd.

## 2. The change

1. **`internal/infrastructure/mcp/lazyclient.go`** — on the first `CallTool`, after connecting, the
   lazy client **populates the session's `tools/list` cache** (one `ListTools` call, bounded by the
   call timeout) before delegating. A cache hit that runs **no** tool still connects nothing
   (FR-002 preserved); a tool that **is** called pays one `tools/list` round-trip (the live path
   already did).
2. **`internal/infrastructure/mcp/toolcache.go` + the composition root** — the cache path becomes
   **`<workspace>/mcp-toolcache.json`** (`$TELL_ME_HOME/output/<mode>/…`); the discoverer seam takes
   the resolved workspace instead of the home. `--new` must **not** clear it (it is not a known
   archive target).
3. **Witness** — extend the hermetic fake MCP server (`internal/infrastructure/mcp/mcptest/`) to
   advertise a tool with an `x-mcp-header`-annotated argument (the SDK server enforces the header
   and rejects a call without it, reproducing the real failure), then add the warm-cache E2E + unit
   pins that redden when the lazy client skips `tools/list`, plus the cache-location pins.
4. **Truth + records** — `specs/truth/techstack.md` (the *MCP tool cache* row + the *MCP tool
   discovery* row), `specs/truth/data/data-model.dbml` (the cache entry's location), the `chat` MCP
   feature + `dsl.md`; **ADR 0059** (+ index; supersedes ADR 0058 D2's home-root placement and
   qualifies its D4 lazy-connect).

## 3. Requirements

### 使用者故事 1 — 暖快取的工具仍能呼叫需要 header 路由的伺服器 (Priority: P1)

As an operator with a warm MCP tool cache, I want a cached tool to still call its server correctly,
so a server that routes arguments as HTTP headers (the GitHub MCP server) keeps working.

**驗收情境**

1. **Given** a warm, fresh cache entry for an MCP server that routes a tool argument via a header,
   and the model calls that tool, **When** the turn runs, **Then** the call reaches the server with
   the required header and the turn completes.

**功能需求（FR）**

- **FR-001**: On a **warm cache hit**, a call to a cached tool whose server routes an argument as an
  HTTP header MUST succeed — the lazy client MUST populate the client session's `tools/list` cache
  before delegating the call (so the SDK emits the `Mcp-Param-*` header).
  [Verification Intent: observable → the E2E "a remembered header-routed tool is actually run" + a unit pin]
- **FR-002**: A cache hit that executes **no** tool MUST still make **zero** MCP connections (round
  087 FR-001 preserved). [Verification Intent: observable → the existing warm-cache zero-contact Example + unit pins]
- **FR-003**: The cache file MUST live under the **per-mode workspace** —
  `$TELL_ME_HOME/output/<mode>/mcp-toolcache.json`. [Verification Intent: observable → a cache-path Then]
- **FR-004**: `--new` MUST **not** clear the cache (it is not a session-history archive target).
  [Verification Intent: observable → the existing `--new` warm-cache Example]
- **FR-005**: The offered set/order, the naming/envelope/schema contracts, the call-time error
  contract, and the offline-path network-freedom are **unchanged**; **no credential** is persisted.
  [Verification Intent: observable → the existing MCP feature Examples + `verify-no-network`]

### 邊界情況

- **EC-001**: If the `tools/list` warm-up fails at call time (e.g. a flaky server), the call MUST
  still be attempted (best-effort) and any server-side failure MUST remain a **recoverable** result
  — never an abort. [Verification Intent: unobservable → unit pin]
- **EC-002**: A cached tool the server has since dropped MUST remain a **recoverable** result (round
  087 behaviour unchanged). [Verification Intent: observable → the existing dropped-tool Example]
- **EC-003**: With **no** `MCP_SERVERS`, no cache is read for effect and no file is created.
  [Verification Intent: unobservable → unit pin]

### 關鍵實體

- **`MCPToolCacheEntry`** — unchanged shape; its **location** moves to the per-mode workspace.

## 4. Invariants

- **I-1 — header routing restored on the cached path** (FR-001).
- **I-2 — warm-no-tool still dials nothing** (round 087 FR-001, FR-002).
- **I-3 — the cache is mode-scoped** and survives `--new` (FR-003, FR-004).
- **I-4 — no observable surface change** beyond the above (FR-005).
- **I-5 — offline paths stay network-free**; **stdlib-only**; **POSIX-only**.

## 5. Scope

**In**: `lazyclient.go` (warm `tools/list` on first call); the cache path + the `MCPDiscoverer`
seam's workspace argument + the composition root; the `mcptest` fake's header-routed tool; the
unit + E2E pins; the truth rows (`techstack.md`, `data-model.dbml`, the `chat` feature + `dsl.md`);
**ADR 0059** (+ index).

**Out**: the discovery bound, the cache key/TTL/refresh design, the offered set/order, the current
(vendor-extension-stripped) offered schemas, the stdio transport, the MCP `-d` diagnostic;
`go.mod`/`go.sum`.

## 6. Success criteria

- **SC-001** A warm-cached header-routed MCP tool is called successfully end-to-end (FR-001).
- **SC-002** The cache lives at `output/<mode>/mcp-toolcache.json` and survives `--new` (FR-003/4).
- **SC-003** `make check` green; `make test-race` clean; `go.mod`/`go.sum` unchanged.

## 7. Assumptions

- **A1** The SDK requirement is exact: `lookupTool` reads only `tools/list` results — so the fix is
  to perform a `tools/list` on the same session before the call (no schema can be replayed through
  the domain port).
- **A2** The warm-up is **best-effort**: a warm-up failure does not block the call (EC-001); a
  header-routed server then fails recoverably, as it would have pre-fix.
- **A3** No `NEEDS CLARIFICATION`: the issue fixes the goal and the fix; the residual mechanics
  (warm-up placement, best-effort) are RD-owned.
- **A4** `docs/domain-model/**` is **not** modelled (an internal optimization; ADR 0041 escape
  hatch) — recorded in `plan.md` §5.
