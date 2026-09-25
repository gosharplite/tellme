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
| **FR-002 / FR-004** a cache hit that runs no tool dials nothing; `--new` keeps the cache | E2E `tellme never contacted the MCP server "shop"` (warm-reuse + the `--new` Example); unit `TestLazyClient_NoWarmWhenUnused` | Warm **eagerly at assembly** (`lc.ListTools(ctx)` inside `DiscoverCached` after the lazy client is built) ⇒ **2** E2E scenarios red (`never contacted … "shop"`, feature lines 254 / 265) — the mutation that violates the zero-dial guarantee (warming inside `connect` cannot: `connect` is reached only from `CallTool`). |
| **FR-003 / B** the cache lives in the per-mode workspace | E2E `tellme remembered the tools of the MCP server "shop" in the session workspace`; unit `TestFileToolCache_LivesInTheGivenWorkspace` | Root the store one level up (`filepath.Dir(workspace)` = `<home>/output`, not the workspace) ⇒ **6** E2E scenarios red at `the cache must live in the session workspace … no such file`. |
| **EC-003** no servers ⇒ inert (no dial, no tools, **no write**) | unit `TestDiscoverCached_NoServersIsInert` (asserts `saves == 0`) | Write on the no-servers path ⇒ the pin reports `saves != 0`. |
| **EC-001** a warm-up failure does not block the call | unit `TestLazyClient_WarmFailureDoesNotBlockCall` | Make the warm-up fatal on error ⇒ the pin reports a surfaced error. |

## Falsifiability witnesses (reproduced then reverted)

- **A (header routing)** remove the `tools/list` warm-up ⇒ unit RED (`lists=0`) + the header-routed E2E Example RED (server does not record the call).
- **B (location)** root the cache one level up (`filepath.Dir(workspace)`) ⇒ **6** E2E scenarios RED at `the cache must live in the session workspace … no such file`.
- **(FR-002/FR-004, the review's mutation C)** warm eagerly at assembly ⇒ 2 zero-contact E2E Examples RED — the mutation that violates the zero-dial guarantee.
- **(EC-003)** write on the no-servers path ⇒ `TestDiscoverCached_NoServersIsInert` would report `saves != 0` (the pin now asserts `saves == 0`).

## Recorded narrowings (non-blocking)

- The warm-up adds one `tools/list` round-trip **on the first cached call** (RF-088-1); a cache hit that calls no tool still dials nothing.
- A warm-up failure is **best-effort** (RF-088-2): the call proceeds; a header-routed server then fails recoverably.
- The offered (cached) schemas remain the round-061 vendor-extension-stripped ones; the SDK recovers the annotation from its own `tools/list` (RF-088-3).
- The warm-up is charged to the **call's own deadline**; a slow `tools/list` can tip the first cached call into the structural timeout path — a small regression vs pre-087, recorded (RF-088-5).

## Review-fold ledger (the `architect` peer, PR #183)

| Finding | Resolution |
| --- | --- |
| **F-088-1** truth self-contradiction in `techstack.md` (*shared across modes*) + a stale path on line 168 | Replaced with the per-mode statement; the retired-item path corrected. |
| **F-088-2** the round-087 `chat/dsl.md` rows documented the old `$TELL_ME_HOME/mcp-toolcache.json` | Annotated the three rows + the note with the per-mode path (history preserved, not deleted). |
| **F-088-3** EC-003's carrier over-claimed (no write half) + a stale `EC-004` label | `TestDiscoverCached_NoServersIsInert` now asserts `saves == 0`; the label corrected to EC-003. |
| **F-088-4** the rejected alternative mis-stated (`connect` is itself lazy) | Restated as an **assembly-time** warm (the real FR-002 violator) in `research.md` + ADR §Alternatives. |
| **F-088-5** the superseded home-root claim still present-tense on `STATUS.md` + the ADR 0058 index row | One-line qualifiers added. |
| **TD-088-1** the warm-up is charged to the call's deadline | Recorded as ADR 0059 §Forward **RF-088-5** (an optional mitigation is deferred). |
| **N-088-1** hardcoded `butler` in the E2E cache helpers | Derive the mode (`sc.effectiveMode()`). |
| **N-088-2** the plan-side Example omitted the routed-argument clause | Added. |
| **N-088-3** the `listed` memo is not a strict once-guard | Commented (benign — sequential dispatch). |
| **N-088-4** the FR-003 mutation cell mis-labelled `filepath.Dir(workspace)` as “home” | Corrected to “`<home>/output`”. |
