# Tasks: composition-root extraction (round 044 — implementation half)

**Plan Package**: `specs/plans/044-composition-root-extraction`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0013-composition-root-injection.md`, `docs/decisions/0011-layer-discipline-gate.md`, `tools/arch/baseline.txt`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision，或 **ADR 0013**。
- **本輪是結構重構（refactor）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更，`research.md` D9）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪**不新增技術**（stdlib-only；`go.mod`／`go.sum` 不動，`research.md` D11 / `spec.md` NFR-005）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— `specs/truth/features/**` **完全不變**（`spec.md` A5）；因此本輪的 Phase 3 依 **round-042 先例** 改為 **「Implementation（結構抽取）」**：交付 `cmd/tellme` 組裝根、`internal/app/deps`、`domaintools.OutputSink`、MCP 抽取，並在**同一變更**內把 `internal/cli` 的 `internal/infrastructure/*` import 邊逐條移除 + 重生成 `tools/arch/baseline.txt`。
  - **沒有 Feature phase（Gherkin）** —— 驗收載體是 **R1 layer-discipline gate 的 exit code** 與 **unit seam**，**不是** E2E（`spec.md` A1 / FR-012：truth/DSL ~NOOP ⇒ 綠 suite 是 false confidence，round-009 陷阱）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 gate 語意、`specs/truth/features/**`、`internal/domain` 業務邏輯、`internal/infrastructure` adapter 行為。

## Round-044 locked decisions + review directives (implementation constraints — MUST)

> 來自 `spec.md`（FR-001–013 / NFR-001–006 / SC-001–007 + 八條 grill folds）、`research.md` D1–D12b、**ADR 0013**，以及 PR #102 最終認證 **review 5243391406** 的 **TD-1 / RF-1 / RF-2**。

- **[ROOT / Q1]** 組裝根放 **`cmd/tellme`**（`internal/` 之外 ⇒ tier table 豁免）。**不可**放 `internal/app/**`（tier 2 會重造 RULE-B 違規）。
- **[INJECT / Q2, G1, ADR D2]** 新套件 **`internal/app/deps`**（tier 2）定義 **domain-typed** 的 `Dependencies`；欄位：`NewGateway`、`NewHistoryStore`、`NewUsageStore`、**`NewToolUsageStore func(func() (string, error)) history.ToolUsageStore`**、**`NewPromptTracker func(home string, userHome func() (string, error)) history.PromptTracker`**、`NewToolRegistry`（**7-tool agent**）、**`NewTUIRegistry`（3-reader `-i` 用，fix-1）**、`BindToolOutput`、`BindSkillsCatalog`、`NewMetricsProvider`、`MCPDiscoverer`、`UserHomeDir`。**每個欄位都要有具名 consumer**（fix-5/fix-6）。接受寬 bag（G1）；interface segregation 記於 ADR 0013 *Alternatives considered*（**非本輪**）。
- **[ONE SEAM / Q6, fix-2, TD-1]** `cli.Options = { Deps deps.Dependencies ; RunTUIPrompt tuiPromptRunner }`。`newRenderer` **刪除**（改 inline `ui.NewRenderer()`；renderer 注入點仍是 `runtimeEnv.renderer`）。runner 型別**加寬**為 **`type tuiPromptRunner func(ctx context.Context, res resolution, env runtimeEnv, dp deps.Dependencies) (string, bool, error)`**（**`dp`，不是 `opts Options`**——後者自我指涉）。runner 於 `cli` 內 **nil-default**；`cmd/tellme` 只填 `Options.Deps`。
- **[THREAD / fix-4/fix-5, G2]** 注入值以參數名 **`dp`**（`dp deps.Dependencies`）貫穿 `Run`→`run`→`runTurn`/`renderTurn`/`dispatchReporting`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newCallRenderer`/`persistTurnUsage`/`runTUIPrompt` **及** leaf helpers `newTurnSpinner`/`augmentRegistryWithMCP`。flags struct 更名 **`options` → `flags`**。`newTurnSpinner` 的 `dp.NewMetricsProvider` 讀取**必須在 `spinnerGate` 短路之後**（fix-5）。
- **[ASSEMBLY / Q3, fix-3]** `agentTools()` + `newToolRegistry` **搬到 `cmd/tellme`**（`agentTools()` 仍是 **parameterless、read-free、non-overridable**；round-031 `TestAgentToolSchemasAreWellFormed` **一起搬**）。`ToolOutputSink` → **`domaintools.OutputSink`**（`{ Begin func() ; Writer io.Writer ; End func() }` **＋ method `Enabled() bool { return s.Writer != nil }`**——`command.go` 4 處呼叫，`{}`-is-disabled 契約 load-bearing）。
- **[MCP / Q4]** discovery 編排**搬進** `internal/infrastructure/mcp`（cohesive `Discover(ctx, servers, bound, newClient, resolveToken) ([]domaintools.Tool, []string, func())`）；`di.NewRemoteClient`／`di.NewGhTokenResolver` 由 `cmd/tellme` **以參數傳入**；`deps.MCPDiscoverer` 為 func-typed port。`verify-mcp-sdk-confinement` 必須維持綠。
- **[TESTS / fix-7, RF-1]** 刪除**全部 8 個** package-level var（`newGateway`/`newHistoryStore`/`newUsageStore`/`newToolUsageStore`/`newToolRegistry`/`newRenderer`/`newTUIPromptRunner`/`userHomeDir`）＋ `defaultMCPDiscovery`。測試改建 `Options`；**新增 Foundational 共用 fixture `defaultTestDeps(modifiers ...func(*deps.Dependencies)) deps.Dependencies`**（RF-1，**不得** import `internal/infrastructure/*`）。**新增 NFR-003 invariant：任何 `internal/cli` 測試檔不得 import `internal/infrastructure/*`**（gate 併入 `.TestImports`/`.XTestImports`，`arch_test.go:39`）。`cli_test.go` 的 `TestRenderToolUsageDiagnosesReadError` **就地**改用 failing `ToolUsageStore` double（**不**搬到 `cmd/tellme`；`cli.Run` 硬綁 `os.Std*`，無法觀測其 stderr/stdout）。`mcp_discovery_test.go` 的 fakes 搬到 `internal/infrastructure/mcp` 測試。
- **[RF-2]** `cmd/tellme` **無條件**接 `UserHomeDir: os.UserHomeDir`；每個 double（含 `defaultTestDeps`）供 **non-nil** resolver（`GlobalPromptTracker.destPath()` 無 nil 守衛）。adapter 端對稱**非本輪**範圍。
- **[CMD-TEST / fix-8]** `cmd/tellme` 測試 = **relocated assembler gate + deps-construction smoke**；**不得**做 stream 斷言（`cli.Run` 硬綁 `os.Std*`；本輪不加 stream seam、不做 `os.Std*` swap）。
- **[RATCHET / FR-008]** 每移除一條 `internal/cli -> internal/infrastructure/*` 邊，**同一 commit** 內以 `make verify-architecture-update` **重生成** baseline（不得手改）；`dev` 在**每個 commit** 綠燈。終點：`internal/cli` 的 7 條 RULE-B → **0**（baseline 8 → 1，餘 `internal/agent -> internal/ui` 屬 R3/R4）。
- **[WITNESS / FR-012, SC-007]** 可偽性見證：重新引入一條 `internal/cli -> internal/infrastructure/*` import（或還原一條已移除 baseline 行）⇒ gate **紅**；還原後綠。
- **[NO-DEP / NFR-005/NFR-006]** stdlib-only；`go.mod`／`go.sum` 不動；POSIX-only。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D11；stdlib-only，`go.mod`／`go.sum` 不動）。不得把 `deps`／`Options` 骨架塞進 Setup。

## Phase 2: Foundational

**Goal**: 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架，讓 Phase 3 不各自重開檔、不各自發明注入方式。**不寫**產品行為語意、**不**移除任何 import 邊。

- [X] T001 新套件 `internal/app/deps`：定義 domain-typed `Dependencies` 骨架
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY（Composition root row）
    - `specs/plans/044-composition-root-extraction/research.md` -> D2；`docs/decisions/0013-composition-root-injection.md` -> D2
    - `specs/truth/techstack.md` -> CLI Application（Composition root row）
  - 只做：新增 `internal/app/deps/deps.go`，宣告 `type Dependencies struct { … }`，欄位與型別**依 D2 清單**（domain-typed only：`internal/domain/**`、`internal/config`、`internal/home`、stdlib），**含** `NewTUIRegistry` 與 argument-taking 的 `NewToolUsageStore`/`NewPromptTracker` 簽名。
  - 不做：不寫 `cmd/tellme` 的組裝（T008）；不移除任何 var/import（Phase 3）；不引用 `internal/ui`/`internal/agent`（tier-2 上限）。

- [X] T002 `internal/domain/tools`：新增 `OutputSink` port 骨架
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY（execute_command row）
    - `research.md` -> D3；`docs/decisions/0013-composition-root-injection.md` -> D3
    - `internal/infrastructure/tools/tooloutput.go`（現行 `ToolOutputSink` + `Enabled()`）
  - 只做：在 `internal/domain/tools` 宣告 `type OutputSink struct { Begin func() ; Writer io.Writer ; End func() }` 與 `func (s OutputSink) Enabled() bool { return s.Writer != nil }`（`io` 為 stdlib，RULE-C 允許）。
  - 不做：不改 `internal/infrastructure/tools` 的 binder（T006）；不改任何呼叫端（Phase 3）；不動 round-034/038/039/040 渲染行為。

- [X] T003 `internal/cli`：`Options` 骨架 + `Run(args, version, opts)` 簽名 + `flags` 更名落點
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY（Composition root row）
    - `research.md` -> D2；`docs/decisions/0013-composition-root-injection.md` -> D2；`spec.md` -> Q6 / FR-007
    - `internal/cli/cli.go`（`type options` @ :50、`type tuiPromptRunner` @ :206、`Run` @ :312、`run` @ :371、`runTUIPrompt` @ :286）
  - 只做：新增 `type Options struct { Deps deps.Dependencies ; RunTUIPrompt tuiPromptRunner }`；把 `Run` 簽名改為 `Run(args []string, version string, opts Options) int`；**加寬** `tuiPromptRunner` 至 `(ctx, res resolution, env runtimeEnv, dp deps.Dependencies) (string, bool, error)`；`Options.RunTUIPrompt == nil` 時在 `cli` 內 fallback 到 `defaultRunTUIPrompt`；把 parsed-flags struct `options` **更名為 `flags`**（機械更名，全檔一致）。
  - 不做：不移除任何 package-level var（T010）；不改 `agentTools()`／registry（T007）；不改任何 flag 行為或輸出。

- [X] T004 `internal/cli`：共用 in-memory 測試 fixture `defaultTestDeps`（**RF-1**）
  - Read:
    - `research.md` -> D5（migration table + RF-1）；`spec.md` -> NFR-003
    - `internal/cli/turn_test.go`（既有 `fakeStore`/`factoryReturning`/`env` helpers）、`internal/cli/persistence_invariant_test.go`（`noopUsageStore`/`capturingUsageStore`）、`internal/cli/cli_test.go`
  - 只做：新增 `internal/cli/testdeps_test.go`（in-package）內 `func defaultTestDeps(modifiers ...func(*deps.Dependencies)) deps.Dependencies`，回傳**完整填充、no-op、in-memory** 的 bag（fake gateway；`fakeStore`/`capturingUsageStore`/`noopUsageStore`/`noopPromptTracker`；兩個 registry func 回空 `domaintools.NewRegistry()`；no-op binders；nil metrics provider；no-op `MCPDiscoverer`；**non-nil** `UserHomeDir`），可被 `modifiers` 覆寫。
  - 不做：**不得** import `internal/infrastructure/*`（NFR-003）；不寫產品行為；不改既有測試（Phase 3 才改建）。

- [X] T005 `internal/infrastructure/mcp`：`Discover` 落點骨架
  - Read:
    - `research.md` -> D4；`docs/decisions/0013-composition-root-injection.md` -> D4
    - `internal/cli/mcp_discovery.go`（`discover`/`discoverServer`/`serverResult`/`mcpRun`/`defaultMCPDiscovery`）
  - 只做：在 `internal/infrastructure/mcp` 新增 `discovery.go`，宣告 `func Discover(ctx context.Context, servers map[string]config.MCPServerConfig, bound time.Duration, newClient … , resolveToken …) ([]domaintools.Tool, []string, func())`（型別由 `internal/cli` 現用者決定，供 `deps.MCPDiscoverer` 對接）。
  - 不做：不移除 `internal/cli/mcp_discovery.go`（T009）；不改 3 s bound／concurrency／warn+skip 語意；不讓 `mcp` import `di`。

- [X] T006 `internal/infrastructure/tools`：`BindToolOutput` 改接 `domaintools.OutputSink` 落點
  - Read:
    - `research.md` -> D3/D11；`truth-delta.md` -> execute_command row
    - `internal/infrastructure/tools/command.go`（`ToolOutputSink` 5 處參照）、`tooloutput.go`
  - 只做：把 `command.go`／`tooloutput.go` 的 `ToolOutputSink` 型別改名為 `domaintools.OutputSink`（欄位/method 一對一）；`BindToolOutput` 簽名改接 `domaintools.OutputSink`。
  - 不做：不改 4 個 `Enabled()` 呼叫點的**語意**（仍 `sink.Writer != nil` 判 disabled）；不改渲染行為（rounds 034/038/039/040）；不改 `internal/cli` 呼叫端（Phase 3）。

- [X] T007 `cmd/tellme`：組裝根落點骨架（含 `agentTools()`/`newToolRegistry` 搬入點）
  - Read:
    - `research.md` -> D3/D5；`docs/decisions/0013-composition-root-injection.md` -> D1/D3
    - `cmd/tellme/main.go`、`internal/cli/cli.go`（`agentTools()` @ :1152、`newToolRegistry` @ :1124、`bindSkillsCatalog` @ :1169）
  - 只做：在 `cmd/tellme` 建立組裝函式落點（例如 `deps.go`：`buildDeps(...) deps.Dependencies`、`buildOptions(...) cli.Options`）與 `agentTools()`／7-tool registry／TUI 3-reader registry 的搬入骨架；`main` 維持可編譯（暫可保留舊路徑，Phase 3 完成切換）。
  - 不做：不移除 `internal/cli` 的 var（Phase 3）；不改 assembler 的參數/read-free 性質；**不**做 stream 斷言（fix-8）。

## Phase 3: Implementation (結構抽取；follows the round-042 precedent for a non-BDD structural round)

**Goal**: 交付組裝根抽取，逐條移除 `internal/cli -> internal/infrastructure/*` 的 **7** 條 RULE-B 邊並重生成 baseline（8 → 1）；注入值以 `dp` 貫穿；交付 MCP 抽取與 `OutputSink` port。**不**改任何使用者可見行為。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP，`spec.md` A5）；驗收語言為 **R1 layer-discipline gate**（`-tags=arch`）與 unit seam。

**Shared Must Read**:
- `specs/truth/techstack.md` -> CLI Application（**Composition root** row / **Project layout** row / **Interactive TUI prompt** row / **Agent command tool** row）、Skills（**Skills listing tool** row）、MCP Client（**MCP client protocol library** / **MCP tool discovery (non-stall)** / **MCP credential resolver seam** rows）、Testing（**Agent tool-schema gate** row）
- `docs/decisions/0013-composition-root-injection.md`（D1–D5 + *Alternatives considered*）
- `docs/decisions/0011-layer-discipline-gate.md`（RULE-A/B/C/D + tier table + ratchet）
- `tools/arch/baseline.txt`（8 條；終點移除 7 條）
- `specs/plans/044-composition-root-extraction/{research.md,spec.md}` -> D1–D12b / Q1–Q7 / FR-001–013

**Boundary**:
- 一條「seam」一 task；只動該 seam 的建構點、消費端與**直接依賴的測試**。
- 每個移除 import 邊的 commit **必須**同 commit `make verify-architecture-update` 重生成 baseline。
- review 啟動 subagent；通過前不解鎖 Phase 4。
- **不動**：flags／exit code／stream contract／格式／`specs/truth/features/**`／adapter 行為／domain 業務邏輯。

**Parallel Hint**:
- T008–T012 具**共享產品檔**（`internal/cli/cli.go`、`call_renderer.go`、`cmd/tellme`）⇒ **同檔序列調度**（不可全並行）；T013/T014 可各自獨立檔並行；T015 review 等全部回來再啟動。

- [X] T008 gateway + stores seam：注入 `dp.NewGateway`/stores/prompt-tracker
  - Read:
    - `research.md` -> D2/D5；`spec.md` -> FR-003
    - `internal/cli/cli.go`（`newGateway` @ :619、`newHistoryStore`/`newUsageStore`/`newToolUsageStore`/`userHomeDir`、`renderTurn` @ :642、`resolve` @ :485）、`internal/cli/call_renderer.go`（`newUsageStore` @ :65/:196）
  - 做：將 gateway/stores/prompt-tracker 的建構改由 `dp` 提供（`renderTurn` 由 `dp.NewGateway` 取 `factory`）；stores 由 `dp.NewHistoryStore`/`dp.NewUsageStore`/`dp.NewToolUsageStore(dp.UserHomeDir)`/`dp.NewPromptTracker(res.Home, dp.UserHomeDir)`；`renderTurn`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newCallRenderer`/`persistTurnUsage` 收 `dp`。移除 `infrhistory`/`infrallm` 之行內建構。
  - 不做：不動 flags/輸出；不動 `agentTools()`（T007）；不改 adapter 行為。

- [X] T009 MCP discovery 抽取：`internal/cli/mcp_discovery.go` → `internal/infrastructure/mcp.Discover`
  - Read:
    - `research.md` -> D4/D8；`truth-delta.md` -> MCP rows；`spec.md` -> FR-006
    - `internal/cli/mcp_discovery.go`、`internal/infrastructure/mcp/**`、`internal/infrastructure/di/mcp_factory.go`
  - 做：把 `discover`/`discoverServer`/`serverResult`/`mcpRun`/`mcp.*` 呼叫整段搬入 `internal/infrastructure/mcp`（`Discover`，`di` 建構子以參數傳入）；刪除 `internal/cli/mcp_discovery.go`；`augmentRegistryWithMCP` 改收 `dp` 並呼叫 `dp.MCPDiscoverer`（保留 merge + warn 列印）；`cmd/tellme` 以 `di.NewRemoteClient`/`di.NewGhTokenResolver(bound)` 綁 `MCPDiscoverer`。移除 `internal/cli` 的 `infrastructure/di`＋`infrastructure/mcp` import。
  - 不做：不改 3 s bound／並行／warn+skip／prompt-path-only 語意；不讓 SDK 外洩（`verify-mcp-sdk-confinement` 保持綠）。

- [X] T010 刪除 `internal/cli` 全部 8 個 package-level var + `defaultMCPDiscovery`；renderer inline
  - Read:
    - `research.md` -> D5；`spec.md` -> FR-007/SC-002；`docs/decisions/0013-composition-root-injection.md` -> D2/D3
    - `internal/cli/cli.go`（:155–619, :1124 各 var）
  - 做：刪 `newGateway`/`newHistoryStore`/`newUsageStore`/`newToolUsageStore`/`newToolRegistry`/`newRenderer`/`newTUIPromptRunner`/`userHomeDir` 與 `mcp_discovery.go` 的 `defaultMCPDiscovery`；`newRenderer` 改 inline `ui.NewRenderer()`（單一 site @ `cli.go:319`）；移除失敗的 doc comment。
  - 不做：不改任何 flag/輸出；不留任何 package-level var；不引入新相依。

- [X] T011 tool assembly seam：`agentTools()`/`newToolRegistry` 搬 `cmd/tellme`；注入 `NewToolRegistry`/`NewTUIRegistry`/`BindToolOutput`/`BindSkillsCatalog`
  - Read:
    - `research.md` -> D2/D3；`spec.md` -> FR-003/FR-004/FR-005、fix-1；`truth-delta.md` -> Skills row / execute_command row
    - `internal/cli/cli.go`（`agentTools()` @ :1152、`newToolRegistry` @ :1124、`bindSkillsCatalog` @ :1169、`-i` registry @ :229）
  - 做：把 `agentTools()`（**parameterless、read-free、non-overridable**）＋ 7-tool `newToolRegistry` ＋ TUI **3-reader** registry 搬入 `cmd/tellme`；`runTurn` 改 `opts.Deps.NewToolRegistry()`、`dp.BindToolOutput(reg, domaintools.OutputSink{Begin/Writer/End})`（sink 仍由 `ui.NewToolOutputCoordinator` 建，legal）、`dp.BindSkillsCatalog(reg, home.SkillsDir(res.Home))`；移除 `infratools`/`infrskills` import。
  - 不做：不讓 `agentTools()` 吃參數或讀檔；不讓 TUI 用 7-tool registry（語意會變）；不改 `-i`/offline `--tool-usage` 契約。

- [X] T012 threading + TUI runner 加寬 + metrics leaf
  - Read:
    - `research.md` -> D2（threading）/D12；`spec.md` -> FR-007、fix-5；`docs/decisions/0013-composition-root-injection.md` -> D2
    - `internal/cli/cli.go`（`run` @ :371、`runTurn` @ :665、`runTUIPrompt` @ :286、`defaultRunTUIPrompt` @ :216、`newTurnSpinner` @ :865）
  - 做：`dp` 貫穿 `run`/`runTurn`/`dispatchReporting`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newCallRenderer`/`persistTurnUsage`/`runTUIPrompt`；`runTUIPrompt` 與 `defaultRunTUIPrompt` 改收 `dp`（runner 加寬）；`defaultRunTUIPrompt` 用 `dp.NewPromptTracker`/`dp.UserHomeDir`/`dp.NewTUIRegistry`（移除行內 `infrhistory`/`infratools`）；`newTurnSpinner` 收 `dp`，**`dp.NewMetricsProvider` 讀取保留在 `spinnerGate` 之後**；移除 `infratelemetry` import。
  - 不做：不動 spinner 呈現/時序；不在 `spinnerGate` 之前建 metrics provider；不改 `-i` teardown/resume。

- [X] T013 測試 seam 遷移（`dp`/`Options`/`defaultTestDeps`），**不** import infra
  - Read:
    - `research.md` -> D5（migration table）、RF-1/RF-2；`spec.md` -> FR-007 / NFR-003
    - 各 `internal/cli/*_test.go`（`turn_test.go`、`prompt_multiline_test.go`、`persistence_invariant_test.go`、`tui_submit_chrome_test.go`、`tui_dispatch_test.go`、`cli_test.go`）
  - 做：以 `defaultTestDeps` 改建所有曾改 var 的測試：`turn_test.go`（8 個 `runTurn` 站）、`persistence_invariant_test.go`（3）、`tui_submit_chrome_test.go`/`tui_dispatch_test.go`（`run`/`runTurn`/gateway/TUI runner 站）、`prompt_multiline_test.go`（5 個 `run` 站；兩個 archive-path 站建 `fakeStore`/`capturingUsageStore` double）；`cli_test.go` 的 `TestRenderToolUsageDiagnosesReadError` **就地**改注入 failing `ToolUsageStore` double（**不搬**）。`env()`/`factoryReturning` 改傳 `dp`。
  - 不做：不讓任何測試檔 import `internal/infrastructure/*`（NFR-003）；不改測試斷言的**行為語意**；不新增/刪除 scenario。

- [X] T014 round-031 assembler gate 搬 `cmd/tellme` + `mcp_discovery_test.go` fakes 搬 `internal/infrastructure/mcp`
  - Read:
    - `research.md` -> D5；`spec.md` -> FR-004/SC-003、fix-8
    - `internal/cli/tool_registry_test.go`、`internal/cli/mcp_discovery_test.go`、`cmd/tellme/**`
  - 做：把 `tool_registry_test.go`（含 `schemaWellFormed` + `TestAgentToolSchemasAreWellFormed` + registry 名稱 cross-check）**搬**到 `cmd/tellme`，仍 iterate **非可覆寫** `agentTools()`；`mcp_discovery_test.go` 的 fakes 搬為 `internal/infrastructure/mcp` 對 `Discover` 的測試；`cmd/tellme` 測試 = assembler gate + **deps-construction smoke**（**無** stream 斷言，fix-8）。
  - 不做：不弱化 assembler gate（仍 parameterless/read-free/non-overridable + 名稱 cross-check）；不在 `cmd/tellme` 測試做 stream 斷言。

- [X] T015 baseline ratchet：重生成 `tools/arch/baseline.txt`（8 → 1）
  - Read:
    - `spec.md` -> FR-008 / SC-001；`research.md` -> D6；`tools/arch/baseline.txt`；`docs/decisions/0011-layer-discipline-gate.md` -> baseline policy
    - `Makefile` -> `verify-architecture` / `verify-architecture-update`
  - 做：以 `make verify-architecture-update` **重生成** baseline（不得手改）；確認 7 條 `internal/cli -> internal/infrastructure/*` 逐條消失、僅餘 `internal/agent -> internal/ui`；`make verify-architecture` **綠**（0 new / 0 stale / 0 cycles）。
  - 不做：不手改任一 baseline 行；不為綠而保留已不違規的 stale 條目。

- [X] T016 subagent review (phase quality gate)
  - Read: 本輪全部變更檔；`research.md` D1–D12b；`docs/decisions/0013-composition-root-injection.md`；`tools/arch/baseline.txt`
  - 檢驗：`internal/cli` 已**無** `internal/infrastructure/*` import（gate 證據）；`agentTools()` 性質保持；`OutputSink.Enabled()` 存在且 4 站語意不變；`dp` threading 完整（含 `newTurnSpinner` 之 gate-after-read）；TUI 3-reader registry；測試檔未 import infra；無 flag/輸出變更。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 gate 7 → 0、行為零變化（`stdout` byte-exact + E2E）、既有 gates 未受影響，並完成可偽性見證。

**Test Scope**: `tools/arch/**`（`-tags=arch`）；`go test -count=1 ./...`（unit + godog E2E）；`make verify`；Gherkin/DSL topology audit。

- [X] T017 [REGRESSION] 可偽性見證：重引 1 條 edge ⇒ 紅；還原 ⇒ 綠（SC-007）
  - Read: `spec.md` -> FR-012/SC-007；`research.md` -> D10；`docs/decisions/0010-test-deadline-decoupling.md`（見證 doctrine）
  - 做：(a) 暫時在 `internal/cli` 重新引入一個 `internal/infrastructure/*` import ⇒ `make verify-architecture` **紅**並指名 `internal/cli -> internal/infrastructure/*`；還原 → 綠。(b) 暫時還原 1 條已移除的 baseline 行（其違規已不存在）⇒ gate 回報 **stale** 紅；還原 → 綠。
  - 不做：不放寬 assertion；見證後必須還原到 HEAD。

- [X] T018 [REGRESSION] 全量回歸 + `make verify` + topology audit + 行為/範圍檢查
  - Read: `research.md` -> D3/D10；`spec.md` -> FR-009/NFR-004；`specs/truth/techstack.md`（本輪 MODIFY rows）
  - 做：`make verify` **OK**（`verify-architecture` 7→0、`verify-mcp-sdk-confinement` 綠、cross-compile 4/4、lint 0、govulncheck clean）；`go test -count=1 ./...` 全綠（含 E2E）；Gherkin/DSL topology audit **unchanged**（44 · 6 · 16+327 · 1674）；`stdout` byte-exact（spot + E2E）；`gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 僅動 `internal/cli/**`、`internal/app/deps/**`、`internal/domain/tools/**`、`internal/infrastructure/{mcp,tools}/**`、`cmd/tellme/**`、`tools/arch/baseline.txt`、`docs/decisions/**`、`specs/truth/techstack.md`、plan package；`go.mod`/`go.sum` 不變。
  - 不做：不為綠而放寬 assertion／改產品碼／手改 baseline。

- [X] T019 subagent review (round quality gate)
  - Read: 本輪全部變更檔；`spec.md`/`research.md`/`truth-delta.md`；`docs/decisions/0013-composition-root-injection.md`；`tools/arch/baseline.txt`
  - 檢驗：SC-001…SC-007 逐條成立；truth MODIFY rows 與實作一致；ADR 0013 一致；無產品行為變更；每個 `deps` 欄位有 consumer；NFR-003 invariant 成立；見證重現。零問題後交付。

- [X] T020 [CLOSE] 更新 STATUS + PR + 關閉 #100（`STATUS.md` 已在本 PR 更新；PR #104 open；`gh issue close 100` 於 **merge 後** 由 closeout 執行）
  - Read: `specs/plans/044-composition-root-extraction/spec.md` -> Provenance；`STATUS.md`
  - 做：`STATUS.md` 記 round 044 實作交付（7→0）；`tasks.md` 勾 `[X]`；開 PR `044-implement-composition-root-extraction` → `dev`；`gh issue close 100`（交付後）。
  - 不做：不手改已 frozen 的 plan package 內容（僅回寫 task 勾選與 outcome）。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（**Composition root** row） | T001、T003、T007、T008、T011、T018 | PASS |
| `specs/truth/techstack.md` -> CLI Application（**Project layout** row） | T001、T007、T018 | PASS |
| `specs/truth/techstack.md` -> CLI Application（**Interactive TUI prompt** row — 3-reader registry） | T003、T011、T014 | PASS |
| `specs/truth/techstack.md` -> CLI Application（**Agent command tool** row — `OutputSink`） | T002、T006、T011、T018 | PASS |
| `specs/truth/techstack.md` -> Skills（**Skills listing tool** row — assembler 位置） | T011、T014 | PASS |
| `specs/truth/techstack.md` -> MCP Client（**MCP tool discovery** row） | T005、T009、T018 | PASS |
| `specs/truth/techstack.md` -> MCP Client（**MCP client protocol library** row — confinement） | T009、T016、T018 | PASS |
| `specs/truth/techstack.md` -> MCP Client（**MCP credential resolver seam** row） | T005、T009 | PASS |
| `specs/truth/techstack.md` -> Testing（**Agent tool-schema gate** row — assembler 搬移） | T011、T014、T019 | PASS |
| `docs/decisions/0013-composition-root-injection.md`（D1–D5 + Alternatives） | T001、T002、T003、T007、T008–T012、T019 | PASS |
| `docs/decisions/0011-layer-discipline-gate.md`（tier table + baseline policy） | T015、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY ×8 rows + TUI row | T001、T002、T003、T005、T006、T011、T018 | PASS |
| `truth-delta.md` -> Governance ADD（ADR 0013） | 已於 plan half 交付；T019 驗證一致 | PASS |
| `truth-delta.md` -> `/axb-api-plan` / `/axb-data-plan` / `/axb-dsl-refine` NOOP | 豁免（NOOP 不建任務；T018 驗證 topology audit 與 API/data 不變） | PASS |
| `research.md` D1（`cmd/tellme` root） | T007、T008 | PASS |
| `research.md` D2（`deps.Dependencies` + one seam + threading + signatures） | T001、T003、T008、T011、T012、T013 | PASS |
| `research.md` D3（assembly + `OutputSink`） | T002、T006、T011 | PASS |
| `research.md` D4（MCP 搬移 + `MCPDiscoverer`） | T005、T009 | PASS |
| `research.md` D5（var 刪除 + 測試遷移 + RF-1/RF-2） | T004、T010、T013、T014 | PASS |
| `research.md` D6（ratchet same-commit） | T015、T017 | PASS |
| `research.md` D7（ADR 0013） | 已交付；T019 | PASS |
| `research.md` D8（confinement + di boundary） | T009、T016 | PASS |
| `research.md` D9（api/data/dsl NOOP） | T018（audit 不變） | PASS |
| `research.md` D10（witness = gate + unit seams） | T016、T017、T018 | PASS |
| `research.md` D11（scope guard） | T018（diff 範圍檢查） | PASS |
| `research.md` D12b + fix-1…fix-8 | T003（fix-1/2）、T006（fix-3）、T012（fix-4/5）、T008（fix-6）、T013（fix-7）、T014（fix-8） | PASS |
| `spec.md` Q1–Q7（含 G1/G2） | T001、T003、T007、T008、T010、T012 | PASS |
| **TD-1**（widened runner signature） | T003（簽名）、T012（threading） | PASS |
| **RF-1**（`defaultTestDeps` 先建） | T004（Foundational）、T013（消費） | PASS |
| **RF-2**（`UserHomeDir` non-nil 接線） | T008（`dp.NewToolUsageStore(dp.UserHomeDir)`）、T004（fixture non-nil） | PASS |
| `spec.md` FR-001…FR-013 / NFR-001…NFR-006 | T008–T015（FR）、T002（NFR-002）、T013（NFR-003）、T018（NFR-004/005/006） | PASS |
| `spec.md` SC-001…SC-007 | T015（SC-001）、T010/T013（SC-002）、T014（SC-003）、T018（SC-004/005/006）、T017（SC-007） | PASS |
| PR #102 certificaton handoff（merge → impl branch → tasks → implement） | T020（PR/close #100） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。

---

## Execution outcome (T001–T020)

**Delivered** (round-044 implementation half; **behaviour-preserving**; `go.mod`/`go.sum` unchanged):

- **`internal/app/deps/deps.go`** (NEW) — the domain-typed `Dependencies` (12 fields; `NewTUIRegistry`; argument-taking `NewToolUsageStore`/`NewPromptTracker`).
- **`internal/domain/tools/outputsink.go`** (NEW) — `domaintools.OutputSink` struct + `Enabled()` method.
- **`internal/infrastructure/mcp/discovery.go`** (NEW) — `mcp.Discover` (the orchestration; the `di` constructors passed in); `mcp.ClientFactory`.
- **`cmd/tellme/deps.go`** (NEW) + **`main.go`** — the composition root (`buildDeps`/`buildOptions`, `agentTools()` relocated, `newToolRegistry`/`newTUIRegistry`); `main` calls `cli.Run(args, version, buildOptions())`.
- **`internal/cli/cli.go`** — `Options{Deps; RunTUIPrompt}`; `Run(args, version, opts)`; the `tuiPromptRunner` + `defaultRunTUIPrompt`/`runTUIPrompt` widened with `dp`; `flags` rename; `dp` threaded through `run`/`runTurn`/`renderTurn`/`dispatchReporting`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newTurnSpinner`; **all 8 package-level factory vars + `defaultMCPDiscovery` deleted**; `newRenderer` inlined (`ui.NewRenderer()`).
- **`internal/cli/call_renderer.go`** — `newCallRenderer`/`persistTurnUsage` take `dp`.
- **`internal/infrastructure/tools/`** — `command.go` re-signed on `domaintools.OutputSink`; **`tooloutput.go` deleted** (the type moved to the domain).
- **Tests**: `cmd/tellme/deps_test.go` (relocated assembler gate + registry tests + deps smoke); `internal/infrastructure/mcp/discovery_test.go` (the discovery fakes); `internal/cli/testdeps_test.go` (**RF-1** `defaultTestDeps`); migrated `turn_test.go`/`persistence_invariant_test.go`/`cli_test.go`/`prompt_multiline_test.go`/`tui_dispatch_test.go`/`tui_submit_chrome_test.go`; deleted `internal/cli/tool_registry_test.go` + `internal/cli/mcp_discovery_test.go`.
- **`tools/arch/baseline.txt`** — regenerated from the gate: **8 → 1** (only `internal/agent -> internal/ui` remains; the 7 `internal/cli -> internal/infrastructure/*` edges are gone).

### Verification (evidence)

- **The gate proves it**: `make verify-architecture` **green** with the 7 `internal/cli -> internal/infrastructure/*` baseline entries **removed** (0 new / 0 stale / 0 cycles).
- **`make verify`** — **OK** (verify-architecture · verify-mcp-sdk-confinement · cross-compile 4/4 · lint 0 · govulncheck clean).
- **`go test -count=1 ./...`** — **green** (unit + godog E2E; 23 packages, 0 FAIL).
- **Falsifiability witness (SC-007)** — (a) re-introducing a `internal/cli -> internal/infrastructure/history` import ⇒ gate **FAIL** naming it; reverted ⇒ green. (b) a **stale** baseline line (`internal/cli -> internal/infrastructure/history`) ⇒ gate **FAIL** ("remove them from the baseline"); reverted ⇒ green.
- **`gofmt -l .`** clean; **`go.mod`/`go.sum` unchanged**; `specs/truth/features/**` untouched (topology audit unchanged: 44 · 6 · 16+327 · 1674).

### Deviation (disclosed)

The Parallel-Hint subagent dispatch for Phase 3 (`T008`–`T015`) ran **inline** (sequential, single orchestrator) — the round-029/041/042 precedent: this session has no parallel-subagent substrate. The `[P]`/Hint semantics are otherwise honoured (independent-file test migrations were completed and reviewed).
