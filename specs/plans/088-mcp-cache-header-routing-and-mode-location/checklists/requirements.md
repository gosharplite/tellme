# Requirements checklist — round 088 `088-mcp-cache-header-routing-and-mode-location`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/088-mcp-cache-header-routing-and-mode-location/`) — `fresh-package-per-round` (never a re-open of the frozen round 087).
- [x] Anchor: live issue **#182** (DoD = close it) — the round-087 regression (header routing + cache location).
- [x] Clarify **not escalated** — the issue fixes the goal and the fix; residual mechanics are RD-owned (A3).
- [x] Every normative clause (FR / NFR / EC / SC) carries a **Verification Intent**.

## Requirements → judgeable carriers

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | warm-cached header-routed tool is called successfully | E2E `A remembered header-routed tool is actually run` + a unit pin (the warm-up happens) |
| FR-002 | warm + no tool ⇒ zero connections | the existing warm-cache `tellme never contacted …` Example + unit pins |
| FR-003 | cache at `output/<mode>/mcp-toolcache.json` | E2E cache-path Then + unit pin (`NewFileToolCache(workspace)`) |
| FR-004 | `--new` keeps the cache | the existing `--new` warm-cache Example |
| FR-005 | no other surface change; offline network-free | the existing MCP feature Examples + `make verify-no-network` |
| EC-001 | a warm-up failure does not block the call | unit pin (a failing `ListTools` still allows the call) |
| EC-002 | dropped cached tool ⇒ recoverable | the existing `A remembered tool the server has since dropped fails softly` Example |
| EC-003 | no servers ⇒ inert (no dial, no tools, **no write**) | unit pin `TestDiscoverCached_NoServersIsInert` (asserts `saves == 0`) |
| SC-001…003 | the two fixes + gates | the carriers above + `make check` |

## Boundaries

- **In**: `lazyclient.go` (warm `tools/list`); the cache path + the `MCPDiscoverer` seam argument + the composition root; the `mcptest` header-routed tool; unit/E2E pins; `techstack.md` / `data-model.dbml` / the `chat` feature + `dsl.md`; **ADR 0059** (+ index).
- **Out**: the discovery bound, the cache key/TTL/refresh, the offered set/order, the offered schemas, stdio, the `-d` diagnostic; `go.mod`/`go.sum`.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable carriers exist (the header-routed E2E + unit pin; the cache-path pin).
