# Tasks — round 087 `087-mcp-tool-cache`

**Anchor**: issue [#180](https://github.com/gosharplite/tellme/issues/180) (DoD = close it).
Executed One-Shot via `/axb-implement` (Red → Green → Refactor per feature).

## Phase 1 — Setup

- [X] **T001** Create the round branch `087-mcp-tool-cache` off `dev`; create the plan package.
- [X] **T002** `/axb-specify` — `spec.md` + `checklists/requirements.md` + `truth-delta.md` skeleton.

## Phase 2 — Foundational

- [X] **T003** `/axb-spec-by-example` — plan-side acceptance `features/acceptance/reusing-the-discovered-mcp-tools.feature`.
- [X] **T004** `/axb-technical-research` — `research.md` (D1–D9) + **ADR 0058** (+ the ADR index row); the techstack MCP rows + the retired *Not Introduced Yet* item.
- [X] **T005** `/axb-system-analysis` — `plan.md` (one CLI end; api NOOP; data conditional → the cache file; ui skipped; the CLI end carried to `/axb-dsl-refine`).
- [X] **T006** `/axb-data-plan` (light) — `specs/truth/data/data-model.dbml` gains `mcp_tool_cache_entry`.
- [X] **T007** `/axb-dsl-refine` — the `chat` MCP feature Rules/Examples + the `chat/dsl.md` round-087 rows.

## Phase 3 — Test Alignment & Implementation

- [X] **T008** *(test alignment)* Add the E2E steps (`step_r087_toolcache.go`, `scenario_context.go` helpers) + the truth feature Examples, and watch them RED against the un-cached prelude (the warm Examples fail `tellme never contacted`).
- [X] **T009** *(feature green)* Domain: `internal/domain/tools/toolcache.go` (`MCPToolCache` + `MCPToolCacheEntry`) + JSON tags on `MCPToolDefinition`.
- [X] **T010** *(feature green)* Infra: `internal/infrastructure/mcp/toolcache.go` (file store, temp+fsync+rename) + `lazyclient.go` (the deferred connect).
- [X] **T011** *(feature green)* Infra: `DiscoverCached` in `discovery.go` (fresh → serve, stale → serve + refresh, cold → discover + write), with `discoverKeys`/`enabledRemoteKeys` extraction.
- [X] **T012** *(feature green)* Seams: `deps.Discovery.Refresh` + `MCPDiscoverer(ctx, home, servers)`; the CLI post-answer refresh hook; the composition-root wiring + `mcpToolCacheTTL`.
- [X] **T013** *(unit pins)* `toolcache_test.go` + `discovery_cached_test.go` (the falsifiable carriers below).
- [X] **T014** *(refactor)* `gofmt`/`goimports`; package docs; extend the discovery-row comment.

## Phase 4 — Verification / delivery

- [X] **T015** Reproduce each falsifiability witness (mutate → redden → revert) — see the ledger.
- [X] **T016** `make verify` + `go test -count=1 ./...` + `make test-race`; `modelith-check` (no drift); the topology audit.
- [X] **T017** Open the round PR (no Copilot review; only a human merges).

## Claim → Witness ledger

| Claim | Carrier | Mutation that reddens it (reproduced + reverted) |
| --- | --- | --- |
| **FR-001 / W1** a warm fresh cache dials nothing | E2E `tellme never contacted the MCP server "shop"` (2 Examples) + unit `TestDiscoverCached_WarmFreshMakesNoDial` + `TestLazyClient_NoConnectWhenUnused` | Force every key cold in `DiscoverCached` ⇒ the unit pin reports `dialed [u-shop]` and the 3 cache-served E2E Examples fail at `never contacted`. |
| **FR-002 / W3** a cold key discovers once + writes | E2E `tellme contacted the MCP server "shop"` + `tellme remembered the tools of the MCP server "shop"` + unit `TestDiscoverCached_ColdDiscoversAndWrites` | Skip the cache write (`_ = merged`) ⇒ the unit reports `saves=0` and the E2E fails at `tellme remembered the tools`. |
| **FR-003 / W2** a stale entry is served, not re-fetched | E2E `An aged tool list is still offered to the model` + unit `TestDiscoverCached_StaleServesThenRefreshes` | Force a pre-request revalidation ⇒ the E2E fails at `the request offered the tool` (the stopped server drops it) and the unit reports a non-empty pre-request dial. |
| **FR-003 / refresh** a failed refresh keeps the prior entry | unit `TestDiscoverCached_FailedRefreshKeepsPrior` | Overwrite on failure ⇒ the pin reports the changed `fetched_at`. |
| **FR-004 / EC-001** a changed declaration is cold | unit `TestDiscoverCached_DeclarationMismatchIsCold` | Ignore the declaration match ⇒ the pin reports zero dials. |
| **FR-005** the warm offer set/order equals live | unit `TestDiscoverCached_WarmOrderEqualsLive` | Rebuild in map-iteration order ⇒ the pin's name lists diverge. |
| **FR-007 / W5** a call-time connect failure is recoverable | unit `TestLazyClient_ConnectFailureIsRecoverable` + E2E `A remembered tool against a server that has stopped` | Return the connect error ⇒ the unit pin reports `got boom`. |
| **FR-008** a corrupt cache is cold | unit `TestDiscoverCached_CorruptCacheIsCold` | Ignore `Load`'s error ⇒ the pin reports zero dials. |
| **FR-009** no credential persisted | unit `TestFileToolCache_NeverStoresCredential` + `TestFileToolCache_SaveLoadRoundTrip` | Add a `token` field to the entry ⇒ the banned-substring scan fails. |
| **FR-010** `--new` keeps the cache | E2E `A fresh session keeps the remembered tool list` | Have `--new` delete the cache file ⇒ the Example fails at `never contacted`. |
| **EC-003** an unknown server's entry is ignored | unit `TestDiscoverCached_IgnoresEntryForUnknownServer` | Offer from any entry ⇒ the pin's dial set grows. |
| **EC-004** no servers ⇒ inert | unit `TestDiscoverCached_NoServersIsInert` | Dial unconditionally ⇒ the pin reports a dial. |

## Falsifiability witnesses (reproduced then reverted)

- **W1** every key cold ⇒ unit RED + 3 E2E Examples RED.
- **W3** no cache write ⇒ unit RED + 1 E2E Example RED.
- **W5** lazy connect error returned ⇒ unit RED (the E2E Example still passes — the loop folds a non-nil tool error
  back too, so the E2E carries the *observable* outcome while the *fold mechanism* is unit-pinned).

## Recorded narrowings (non-blocking)

- The E2E cannot assert a **zero** connection count *baseline* for a two-run scenario; the round arranges the warm cache
  directly (a Given), so `ConnectionCount() == 0` is asserted against a never-dialed server.
- The **post-answer refresh** is witnessed by the stale Example's `could not be reached` warning (the refresh ran), not by
  a wall-clock assertion.
- The **TTL seam** is the injected clock (unit-only); no `TELL_ME_FORCE_*TTL*` env seam ships (the E2E arranges a stale
  `fetched_at` directly).
