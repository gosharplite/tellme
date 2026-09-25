# Tasks — round 088 `088-mcp-cache-header-routing-and-mode-location`

**Anchor**: issue [#182](https://github.com/gosharplite/tellme/issues/182) (DoD = close it).
Executed One-Shot via `/axb-implement` (Red → Green → Refactor).

## Phase 1 — Setup

- [X] **T001** Branch `088-mcp-cache-header-routing-and-mode-location` off `dev`; plan package.
- [X] **T002** `/axb-specify` — `spec.md` + `checklists/requirements.md` + `truth-delta.md`.

## Phase 2 — Foundational

- [X] **T003** `/axb-spec-by-example` — `features/acceptance/keeping-remembered-mcp-tools-working.feature`.
- [X] **T004** `/axb-technical-research` — `research.md` + **ADR 0059** (+ index; ADR 0058 back-pointer); the `techstack.md` MCP rows.
- [X] **T005** `/axb-system-analysis` — `plan.md`.
- [X] **T006** `/axb-data-plan` (light) — `data-model.dbml` cache location.
- [X] **T007** `/axb-dsl-refine` — the `chat` feature Rules/Examples + `dsl.md` rows.

## Phase 3 — Test Alignment & Implementation

- [X] **T008** *(test alignment)* the `mcptest.Options.HeaderRouted` fake + the `step_r088_toolcache.go` Givens/Thens + the feature Examples, watched RED (the header-routed Example fails `did not record a call to "create_issue"`; the location Example fails `no such file`).
- [X] **T009** *(fix A)* `lazyclient.go` — warm `tools/list` once on the first `CallTool` (best-effort).
- [X] **T010** *(fix B)* `toolcache.go` + the `MCPDiscoverer` seam (`res.Workspace`) + the composition root — the cache lives at `output/<mode>/mcp-toolcache.json`.
- [X] **T011** *(unit pins)* `lazyclient_warm_test.go` (warm-once, best-effort, no-warm-when-unused, workspace path).
- [X] **T012** *(refactor)* `gofmt`/`goimports`; the E2E cache helper reads/writes the workspace.

## Phase 4 — Verification / delivery

- [X] **T013** Reproduce the witnesses (warm-up removed ⇒ unit + E2E red; home-root path ⇒ the location Example red).
- [X] **T014** `make check` + `make test-race`; `modelith-check`; the topology audit.
- [X] **T015** Open the round PR.

## Claim → Witness ledger

| Claim | Carrier | Mutation that reddens it (reproduced + reverted) |
| --- | --- | --- |
| **FR-001 / A** a warm-cached header-routed tool is called successfully | E2E `A remembered header-routed tool is actually run` (`tellme called the tool "create_issue" on the MCP server "gh"`); unit `TestLazyClient_WarmsToolsListOnceBeforeCall` | Remove the `warmToolsList` call in `lazyClient.CallTool` ⇒ the unit pin reports `lists=0 calls=1`, and the E2E fails at `did not record a call to "create_issue"` (the server 400s — no `Mcp-Param-owner`). |
| **FR-002** a cache hit that runs no tool dials nothing | E2E `tellme never contacted …` (warm-reuse Examples); unit `TestLazyClient_NoWarmWhenUnused` | (existing round-087 carriers; the warm-up is inside `CallTool`, so an unused lazy client dials/lists nothing) |
| **FR-003 / B** the cache lives in the per-mode workspace | E2E `tellme remembered the tools of the MCP server "shop" in the session workspace`; unit `TestFileToolCache_LivesInTheGivenWorkspace` | Root the store at `filepath.Dir(workspace)` (home) ⇒ the E2E fails at `the cache must live in the session workspace … no such file`. |
| **FR-004** `--new` keeps the cache | E2E `A fresh session keeps the remembered tool list` (unchanged) | — |
| **EC-001** a warm-up failure does not block the call | unit `TestLazyClient_WarmFailureDoesNotBlockCall` | Make the warm-up fatal on error ⇒ the pin reports a surfaced error. |

## Falsifiability witnesses (reproduced then reverted)

- **A (header routing)** remove the `tools/list` warm-up ⇒ unit RED (`lists=0`) + the header-routed E2E Example RED (server does not record the call).
- **B (location)** root the cache at the home ⇒ the location E2E Example RED (`no such file`).

## Recorded narrowings (non-blocking)

- The warm-up adds one `tools/list` round-trip **on the first cached call** (RF-088-1); a cache hit that calls no tool still dials nothing.
- A warm-up failure is **best-effort** (RF-088-2): the call proceeds; a header-routed server then fails recoverably.
- The offered (cached) schemas remain the round-061 vendor-extension-stripped ones; the SDK recovers the annotation from its own `tools/list` (RF-088-3).
