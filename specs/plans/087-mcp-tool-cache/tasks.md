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

- [X] **T008** *(test alignment)* the E2E steps (`step_r087_toolcache.go`, `scenario_context.go` helpers) + the truth feature Examples, RED against the un-cached prelude.
- [X] **T009** *(feature green)* Domain: `internal/domain/tools/toolcache.go` (`MCPToolCache` + `MCPToolCacheEntry`) + JSON tags on `MCPToolDefinition`.
- [X] **T010** *(feature green)* Infra: `internal/infrastructure/mcp/toolcache.go` (file store through the `cacheFS` seam) + `lazyclient.go` (the deferred connect).
- [X] **T011** *(feature green)* Infra: `DiscoverCached` in `discovery.go` (fresh → serve, stale → serve + refresh, cold → discover + write), with `discoverKeys`/`enabledRemoteKeys` extraction.
- [X] **T012** *(feature green)* Seams: `deps.Discovery.Refresh` + `MCPDiscoverer(ctx, home, servers)`; the CLI post-answer refresh hook; the composition-root wiring + `mcpToolCacheTTL`.
- [X] **T013** *(unit pins)* `toolcache_test.go` + `discovery_cached_test.go`.
- [X] **T014** *(refactor)* `gofmt`/`goimports`; package docs; the `cyclop` refactor (CC 24 → ≤15).

## Phase 4 — Review folds (the `architect` peer, PR #181)

- [X] **T015** Fold **F-087-1** — add the cached-tool **success-path** carrier (a live-server Example) + `TestLazyClient_DelegatesOnFirstCall`.
- [X] **T016** Fold **F-087-2** — restate the witness figures (this ledger + the day summary + the PR body).
- [X] **T017** Fold **F-087-3** — reword FR-006 to its actual mechanism (a recoverable `error: …`) + add the dropped-tool Example.
- [X] **T018** Fold **F-087-4** — the `cacheFS` durability seam + `TestFileToolCache_SaveUsesTempThenRename` + `TestFileToolCache_SaveFailureKeepsPrior`.
- [X] **T019** Fold **F-087-5** — the write-error pin (`TestDiscoverCached_WriteErrorIsBestEffort`), the sibling-warm pin (`TestDiscoverCached_MismatchedSiblingStaysWarm`), the E2E credential scan, and the corrected checklist carriers.
- [X] **T020** Fold **F-087-6** — record the deferred/warn-less credential resolution (techstack row + ADR RF-087-9); reconcile `research.md` ↔ `truth-delta.md`.
- [X] **T021** Fold **TD-087-1** — record the repeated post-answer re-dial (§Consequences + RF-087-10).
- [X] **T022** Fold the nits — **N-087-2** (memoised failed connect in D4), **N-087-3** (the EC-003 pin asserts the requirement, not preservation), **N-087-4** (re-normalize the cached schema), **N-087-6** (`issue #180`), **N-087-1** (record the probabilistic FR-005 mutation).
- [X] **T023** Re-run the gates; post the fold comment on the PR.

## Phase 5 — Delivery

- [X] **T024** Re-run `make verify` / `go test ./...` / `make test-race`; the topology audit.
- [X] **T025** Post the fold comment on PR #181; leave the PR for a human to merge.

## Claim → Witness ledger

| Claim | Carrier | Mutation that reddens it (reproduced + reverted) |
| --- | --- | --- |
| **FR-001 / W1** a warm fresh cache dials nothing | E2E `tellme never contacted the MCP server "shop"` (2 Examples: reuse + `--new`) + unit `TestDiscoverCached_WarmFreshMakesNoDial` (1 pin) + `TestLazyClient_NoConnectWhenUnused` | Force every key cold in `DiscoverCached` ⇒ the unit pin reports `dialed [u-shop]`; the 2 `never contacted` Examples + (when the server is closed/failing) 2 more Examples at `the request offered the tool …` redden (**4 Examples** total). |
| **FR-001 success / F-087-1** a cached tool the model calls reaches its server | E2E `A remembered tool against a live server is actually run` (`tellme called the tool …`) + unit `TestLazyClient_DelegatesOnFirstCall` | Make `lazyClient.CallTool` never delegate (return a fixed result) ⇒ the unit pin reports a non-`ok` text + the E2E Example fails at `tellme called the tool`. |
| **FR-002 / W3** a cold key discovers once + writes | E2E `tellme contacted the MCP server "shop"` + `tellme remembered the tools of the MCP server "shop"` (1 Example) + unit `TestDiscoverCached_ColdDiscoversAndWrites` + `TestDiscoverCached_CorruptCacheIsCold` (2 pins) | Skip the cache write (`_ = merged`) ⇒ the unit reports `saves=0` (2 pins) + the E2E fails at `tellme remembered the tools`. |
| **FR-003 / W2** a stale entry is served, not re-fetched | E2E `An aged tool list is still offered to the model` (1 Example, server closed) + unit `TestDiscoverCached_StaleServesThenRefreshes` + `TestDiscoverCached_FailedRefreshKeepsPrior` (2 pins) | Treat stale as cold / force a pre-request revalidation ⇒ the unit pins redden and the E2E fails at `the request offered the tool …` (the closed server drops it). |
| **FR-003 refresh** a failed refresh keeps the prior entry | unit `TestDiscoverCached_FailedRefreshKeepsPrior` | Overwrite on failure ⇒ the pin reports the changed `fetched_at`. |
| **FR-004 / EC-001 / F-087-5** a changed declaration is cold; the sibling stays warm | unit `TestDiscoverCached_DeclarationMismatchIsCold` + `TestDiscoverCached_MismatchedSiblingStaysWarm` | Ignore the declaration match ⇒ the pins report dial sets that include the warm key. |
| **FR-005** the warm offer set/order equals live | unit `TestDiscoverCached_WarmOrderEqualsLive` | Rebuild `run.Tools` in map-iteration order ⇒ the pin's name lists diverge (**N-087-1**: discriminating in ~1 of 8 single runs with 2 keys — `-count=20` makes it reliable). |
| **FR-006 / F-087-3** a dropped/renamed cached tool fails softly | E2E `A remembered tool the server has since dropped fails softly` + `the run continued past the failed MCP tool call` — a **companion guard**, not a discriminating witness (see the note) | The recoverable fold is **structural / loop-owned** (round-032 TD1/R3, 076/080): no in-round mutation of this feature reddens it (measured — a tool-level-error mutant leaves all scenarios green). The Example pins the *observable* outcome; the fold *mechanism* stays owned by the round-032/076 unit tier. |
| **FR-007 / W5** a call-time connect failure is recoverable | unit `TestLazyClient_ConnectFailureIsRecoverable` + E2E `A remembered tool against a server that has stopped` | Return the connect error ⇒ the unit pin reports `got boom`. |
| **FR-008 / F-087-5** a corrupt cache / a write error is best-effort | unit `TestDiscoverCached_CorruptCacheIsCold` + `TestDiscoverCached_WriteErrorIsBestEffort` | Ignore `Load`'s error / fail on the save ⇒ the pins' dial sets / tool sets differ. |
| **FR-009** no credential persisted | unit `TestFileToolCache_NeverStoresCredential` + E2E `tellme remembered the tools …` (the step scans the cache file) | Add a `token` field to the entry ⇒ the banned-substring scans fail (`found "token"`). |
| **FR-010** `--new` keeps the cache | E2E `A fresh session keeps the remembered tool list` | Have `--new` delete the cache file ⇒ the Example fails at `never contacted`. |
| **trust boundary (N-087-4 / RES-087-FV-3)** a cached name/schema is untrusted | unit `TestDiscoverCached_CachedNameAndSchemaReValidated` (an invalid name + an unsafe schema + a good tool ⇒ only the good tool offered) | Trust the cached name + skip schema normalization ⇒ the pin reports `[mcp_shop_ok_tool mcp_shop_bad name mcp_shop_unsafe]`. |
| **F-087-4** the write is a temp file + rename, not in place | unit `TestFileToolCache_SaveUsesTempThenRename` + `TestFileToolCache_SaveFailureKeepsPrior` | Replace `Save` with an in-place `os.WriteFile` ⇒ the recording seam sees no `createtemp`/`rename` (pin reddens); a failed rename leaves the prior file byte-intact. |
| **EC-003** an unknown server's entry is ignored | unit `TestDiscoverCached_IgnoresEntryForUnknownServer` (asserts the unconfigured key is never offered) | Offer from any entry ⇒ the pin's dial/tool sets grow. |
| **EC-004** no servers ⇒ inert | unit `TestDiscoverCached_NoServersIsInert` | Dial unconditionally ⇒ the pin reports a dial. |

## Falsifiability witnesses (reproduced then reverted — measured on a scratch copy of the fold head)

- **W1** every key cold ⇒ **unit** `TestDiscoverCached_WarmFreshMakesNoDial` reports `dialed [u-shop]`; **E2E 4 Examples** redden — 2 at `tellme never contacted the MCP server "shop"` and 2 at `the request offered the tool "lookup_price" … alongside the agent tools` (the cold dial against the closed/failing server drops the tool).
- **W2** treat stale as cold ⇒ **unit 2 pins** (`TestDiscoverCached_StaleServesThenRefreshes`, `TestDiscoverCached_FailedRefreshKeepsPrior`); **E2E 1 Example** at `the request offered the tool …`.
- **W3** skip the cache write ⇒ **unit 2 pins** (`TestDiscoverCached_ColdDiscoversAndWrites`, `TestDiscoverCached_CorruptCacheIsCold`); **E2E 1 Example** at `tellme remembered the tools …`.
- **W5** return the lazy connect error ⇒ **unit** `TestLazyClient_ConnectFailureIsRecoverable` reports `got boom` (the E2E Example still passes — the loop also folds a non-nil tool error back, so the E2E carries the observable outcome while the fold *mechanism* is unit-pinned).
- **F-087-1** `lazyClient.CallTool` never delegates ⇒ **unit** `TestLazyClient_DelegatesOnFirstCall` + **E2E** `A remembered tool against a live server is actually run`.
- **F-087-4** in-place `os.WriteFile` ⇒ **unit** `TestFileToolCache_SaveUsesTempThenRename` (no `createtemp`/`rename` events).
- Extra probes (affirmations): always-schedule-refresh ⇒ the warm Examples red; a CLI-without-refresh ⇒ the aged Example reds.

## Recorded narrowings (non-blocking)

- **FR-006's Example is a companion guard, not a discriminating witness** (fold-verification RES-087-FV-2): making `CallTool` return a non-nil error on a tool-level error leaves the whole suite green — the recoverable fold is structural/loop-owned (round-032 TD1/R3 + 076/080), so no in-round mutation of *this* feature reddens it. The Example still pins the observable outcome end-to-end.

- The E2E arranges the warm cache **directly** (a Given), so `ConnectionCount() == 0` is asserted against a never-dialed server; there is no two-run baseline.
- The **post-answer refresh** is witnessed by the aged Example's `could not be reached` warning (the refresh ran), not by a wall-clock assertion.
- The **TTL seam** is the injected clock (unit-only); no `TELL_ME_FORCE_*TTL*` env seam ships (the E2E arranges a stale `fetched_at` directly).
- **TD-087-1**: a stale entry that cannot be refreshed is re-dialed after **every** subsequent run (bounded, post-answer, one warning each) — a negative-cache/backoff policy is deferred (RF-087-10).
- **RF-087-8 / RF-087-9**: the failed lazy connect is memoised for the run (no reconnect) and the cached path's credential warning is not surfaced — recorded in ADR 0058 §Forward.
- **N-087-1**: the FR-005 order pin is probabilistically discriminating (~1 in 8 single runs with 2 keys; `-count=20` reliable).
