# Tasks: 032 — remote MCP client

**Plan Package**: `specs/plans/032-mcp-client`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `features/acceptance/*.feature`, `specs/truth/techstack.md`, `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`, `specs/truth/features/cli/chat/dsl.md`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**) and `specs/truth/data/**` (`/axb-data-plan` **NOOP**) are unchanged this round. There is no `ui/**` (the MCP surface is model-facing; the operator `stderr` chrome is unchanged).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision；或對應 `specs/truth/features/cli/**` 的一條 DSL row。
- **本輪是「新增能力」輪**：`/axb-api-plan` 與 `/axb-data-plan` 為 **NOOP**；`/axb-technical-research` **MODIFY** `techstack.md`（已交付）；`/axb-dsl-refine` **ADD** `specs/truth/features/cli/chat/**`（已交付）。因此：
  - **Phase 1 `Setup` 存在** —— 本輪新增技術（`github.com/modelcontextprotocol/go-sdk` v1.7.0）。
  - **Phase 3 `Test Alignment`** 把 **15** 條新 DSL row 各落一條 `[BDD-RED]` stepdef，並把純函式（config validation、auth resolution、naming、discovery bound、fold additions）落 `[UNIT]`。
  - **Phase 4 有 Feature phase** —— `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`（ADD）走 `[BDD-GREEN] → [BDD-REFACTOR]`。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision、對應 `specs/truth/**` 與 plan artifacts。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** the native offered-tool set（六工具）、the sequential tool-execution model、the class-phrase vocabulary 或 exit code、`stdout` byte-exactness，或用 offline-path no-network 保證。

## Round-032 locked decisions (implementation constraints — MUST)

> 來自 `research.md` Decisions 1–11 與 spec US1/US2/US3（FR-001..FR-021、NFR-001..NFR-005）與 Q1–Q5。

- **[TRANSPORT]** Remote **Streamable HTTP** only；`COMMAND`（stdio）entry 由 config validation **warn+skip**（非 fatal；FR-013；Q1）。
- **[SDK/CONFINEMENT]** MCP 協定由官方 `github.com/modelcontextprotocol/go-sdk` v1.7.0 實作；**只有** `internal/infrastructure/mcp/` 可 import SDK，其餘只消費 `tools.MCPClient` domain port；由 `verify-mcp-sdk-confinement` Makefile gate 強制（Q3）。
- **[NON-STALL]** Discovery 併發、每 server 一個 **小固定 fast-fail deadline**（預設 3 s，獨立於 tool-call timeout）；未回應者 warn+skip；整體延遲有界；註冊順序 **sorted by server key**；`ENABLED: false` 者 **完全不連線**（Q2）。
- **[PROMPT-PATH-ONLY]** Discovery 只在 **prompt-bearing turn** 發生；offline paths（`--version`/`-d`/`-l`/prompt-less `--new`/boot）**不連網**（FR-016）。
- **[AUTH]** 依 auth mode 解析（`auto`/`gh`/`bearer`/`basic`/`none`）：explicit `TOKEN` 優先；GitHub-hosted 主機 fallback 至 token 來源（`gh auth token` → `GITHUB_TOKEN`，memoized），否則匿名；token **不得被記錄**（FR-006/017；Q4）。`gh` spawn **受同一 fast-fail bound 限制** + injectable seam（B3/R6）。
- **[NAMING]** 以確定性名稱 `mcp_<server>_<tool>` 提供；超過 64 bytes 時以**推導預算**（bytes）截斷 tool 段並加 8-hex SHA-256（FR-004；Q4/TD3）。
- **[VALIDATION/CONTRACT]** typed `MCP_SERVERS` + 確定性驗證（key `^[a-z0-9-]{1,24}$`、URL 必填、auth 合法 + 對應憑證、`TIMEOUT ≥ 0`、`ENABLED` 預設 true）；**unmodelled sub-keys 容忍**；**`COMMAND` warn+skip**；只有 malformed **remote** entry fatal；失敗 **重用既有 class phrase / exit code**，**不新增** failure class（FR-002/015；TD4）。
- **[NO-DEP-ELSE]** 除 SDK 外無其他新相依；POSIX-only；MCP tool 逾時走既有 FR-018 nil-error timeout result；任何 call-time failure 皆為 recoverable tool result。
- **[REVIEW FOLDS (PR #66)]** **B1** — MCP tool schemas are normalized/verified (`required ⊆ properties`) before offering; unsafe → skip+warn (FR-019; T003/T008/T028). **B2** — the e2e fake is **SDK-built** in `internal/infrastructure/mcp/mcptest/`; the confinement gate covers production **and** test files (T007/T008). **B3** — token resolution is bounded + an injectable `tokenResolver` seam; no `gh` spawn in tests (FR-020; T006/T028). **TD1** — a tool-call failure (tool-level / transport / closed client) is **recoverable**, never aborts the run (FR-018; T028). **TD2** — the port types `InputSchema json.RawMessage` / `args map[string]interface{}` with the nil→`{}` invariant (T002). **TD3** — the name budget is **derived** (bytes) (T005/T028). **TD4** — unmodelled sub-keys tolerated; `COMMAND` warn+skip (FR-002/013; T024). **TD5** — the MCP timeout follows the resource contract (default 300 s; clamped by the fixed 7200 s ceiling) (FR-021; T028). **TD7** — `ENABLED` addition / `REQUIRES_CONSENT` drop recorded. **TD8** — `-d` MCP diagnostics deferred. **R1** — structure paths pinned in `plan.md`. **R3** — gate hygiene (`.PHONY` + `help` + aggregate).

---

## Phase 1: Setup

**Goal**: 引入 MCP 協定相依並以一個 smoke test 證明可編譯 —— 不寫 DSL 語意、不寫產品行為。

- [X] T001 [SETUP] Add the MCP Go SDK dependency + smoke test
  - Read: `specs/plans/032-mcp-client/research.md` -> `Decision 2`；`go.mod`
  - 做：`go get github.com/modelcontextprotocol/go-sdk@v1.7.0`（pin v1.7.0，與 reference 一致）並 `go mod tidy`；在 `internal/infrastructure/mcp/` 新增一個最小 smoke test —— import `github.com/modelcontextprotocol/go-sdk/mcp` 並建構一個 `mcp.NewClient`（或等義的編譯期 smoke），證明 SDK 可載入且 `go build ./...` 通過。
  - 只做：相依 + smoke test。不做：任何 DSL 語意、任何 MCP 產品行為、discovery/auth/registration。
  - 邊界：`go.mod`/`go.sum` 是本輪唯一新增相依的落點；不得引入 SDK 之外的新模組。

## Phase 2: Foundational

**Goal**: 建立後續實作與測試的落點骨架（Zero Shared Edits；獨立檔案），**不寫 Phase 3 測試層、不寫 Feature Green**。

- [X] T002 [FOUNDATION] Domain port `tools.MCPClient`
  - Read: `specs/truth/techstack.md` -> MCP Client / MCP client protocol library row；`research.md` -> `Decision 2`；`internal/domain/tools/`（既有 `Tool` port）
  - 做：新增 `internal/domain/tools/mcp_client.go` —— `MCPToolDefinition{Name, Description, InputSchema json.RawMessage}` 與 `MCPClient` interface（`ListTools(ctx) ([]MCPToolDefinition, error)`、`CallTool(ctx, name string, args map[string]interface{}) (ToolResult, error)`、`Close() error`）。**TD2：** 型別固定為 `InputSchema json.RawMessage` 與 `args map[string]interface{}`，並在 port 註解寫明 **nil args → `{}`** 的不變式（MCP wire 不得攜帶 `"arguments": null`，strict server 會拒）。**R3：** port 註解並寫明 **任何 call-time 失敗（tool-level error、transport failure、以及對「已關閉 client」的呼叫）一律以一個 `ToolResult`（nil error，carrying the error text）回傳**，使 loop 一律走 **recoverable** 路徑；`CallTool` **不因任何 call-time 失敗回非 nil error**（`error` 回傳僅保留給介面一致性／未來真正 terminal 條件，今日 call-time 失敗不會用到）。network-free、protocol-free。
  - 只做：型別 + interface（含兩條不變式註解）。不做：任何 SDK import、任何 transport、normalization（T028）。

- [X] T003 [FOUNDATION] Adapter package `internal/infrastructure/mcp/` (SDK confined here)
  - Read: `research.md` -> `Decision 2`；`internal/domain/tools/mcp_client.go`
  - 做：新增 `internal/infrastructure/mcp/client.go` —— 一個 constructor 與 `ListTools`/`CallTool`/`Close` 的骨架，實作 `tools.MCPClient`（**R2**：`CallTool` 骨架須在發出 wire 前把 **nil args 正規化為 `{}`**）；並新增 `internal/infrastructure/mcp/schema.go` 的骨架（**B1**：`normalizeMCPSchema(raw json.RawMessage) (json.RawMessage, error)` —— 非 object/absent/unparseable → freeform；確保 `required ⊆ properties`；無法安全化 → 回 error 供 skip+warn）。SDK import **只** 出現在此 package（`client.go`）。
  - 只做：package + 編譯骨架（`client.go` + `schema.go`）。不做：credential 解析細節、discovery 接線、schema 邏輯測試（T028）。

- [X] T004 [FOUNDATION] Typed `MCP_SERVERS` config + validation skeleton
  - Read: `research.md` -> `Decision 6`；`specs/truth/techstack.md` -> MCP server registry (config) row；`internal/config/`
  - 做：新增 `MCPServerConfig`（`URL`/`TOKEN`/`USERNAME`/`AUTH`/`TIMEOUT`/`ENABLED`）並把 `MCP_SERVERS` 解碼進 `Config`；放一個 validation stub。
  - 只做：型別 + decode + stub。不做：validation 邏輯（T024 unit）。

- [X] T005 [FOUNDATION] Deterministic tool-name helper + UNIT landing
  - Read: `research.md` -> `Decision 5`
  - 做：新增 pure function 產出 `mcp_<server>_<tool>`；**TD3** 超 64 bytes 時以**推導預算**截斷 tool 段（`maxPrefixLen = 64 − 4 − len(server) − 1 − 1 − 8`，cap 40，floor 0）並接 8-hex SHA-256；預算以 **bytes（`len()`）** 計，不以 runes。與其 `[UNIT]` landing 檔。
  - 只做：pure helper + 空測試檔。不做：實作測試斷言（T026/T028）。

- [X] T006 [FOUNDATION] Discovery / registration + credential seams in `internal/cli`
  - Read: `research.md` -> `Decision 3`, `Decision 4`；`specs/plans/032-mcp-client/plan.md`（Structure）
  - 做：新增（骨架）一個在 prompt path 上被呼叫的 discovery 函式（給定 config + client factory → 併發探測 enabled server、以小固定 bound 收集 tools、normalize schema（B1，無法安全化者 skip+warn）、sorted 回傳），與一個 credential-resolution seam。**B3：** credential seam 是一個 **injectable `tokenResolver func(ctx) (string, error)`**（default = 受 **同一個 fast-fail bound** 限制的 `gh auth token` spawn；timeout/失敗 → warn + anonymous）；註冊時使用 T005 的 naming。
  - 只做：seams。不做：bound 值調校/測試（T027）、真實 auth 解析（T025）、token 邊界測試（T028）。

- [X] T007 [FOUNDATION] Makefile gate `verify-mcp-sdk-confinement`
  - Read: `research.md` -> `Decision 2`；`Makefile`（既有 `verify-*` 樣式）
  - 做：新增 target 斷言 `github.com/modelcontextprotocol/go-sdk` 的 import **只** 出現在 `internal/infrastructure/mcp/**`（含 `mcptest/`），並接入 `make verify`/`test`。**R3（gate hygiene）：** 比照 round-020 樣式 —— `.PHONY` + `help` 條目 + 接入與其他 `verify-*` 相同的 aggregate；**明確涵蓋 production 與 `_test.go`**（reference 如此），故 `mcptest/` 內的 SDK-built fake 合規。
  - 只做：gate + 接線。不做：其他工具鏈變更。

- [X] T008 [FOUNDATION] Fake MCP server harness + stepdef/UNIT landing files
  - Read: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`；`chat/dsl.md` -> `## Given (round 032)` + `## Then (round 032)`；`tests/e2e/`（既有 fake provider 樣式）
  - 做：新增一個 in-process **fake MCP server**；**B2：** 它以 **SDK 的 server API** 建構（不手刻 Streamable-HTTP wire），住在 **`internal/infrastructure/mcp/mcptest/`**（exported，供 e2e import）—— mirroring the reference（其 fake 在 `internal/infrastructure/mcp/` 內、「starts a real SDK MCP server (same wire…)」）。提供可 script 的：一個 tool、never-answers 模式、token 要求、recording 模式、**malformed-schema** 模式、**tool-error** 模式、**transport-failure** 模式（B1/TD1）。**為 15 條新 DSL row 各建獨立 stepdef landing 檔**（Zero Shared Edits）；為純函式建 `[UNIT]` landing 檔。
  - 只做：harness（SDK-built）+ 空落點。不做：step 實作（Phase 3）、任何 hand-rolled protocol。

## Phase 3: Test Alignment & Implementation (test layer)

**Test Scope**: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`.

**DSL 參照**: `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 032)`（8 rows）與 `## Then (round 032)`（7 rows）。讀取每列的 **StepDef 實作語意** 欄以取得該 step 的落地契約。

**Markers**: `[P]` `[BDD-RED]`（每條新 DSL row 一條，共 **15**）；`[P]` `[UNIT]`（純函式）。**只動測試層，不寫產品碼。**

**Shared Must Read**: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`；`specs/truth/features/cli/chat/dsl.md` -> `## Given/Then (round 032)`；`specs/plans/032-mcp-client/research.md` -> `Decision 3`–`Decision 5`, `Decision 8`–`Decision 10`。

**Boundary**: 只動 `tests/e2e/**` 與 `internal/**/*_test.go`；不寫產品碼；review 啟動 subagent；0 undefined steps 前不解鎖 Phase 4。

**Parallel Hint**: 下方 `[P]` 任務落點互不重疊（每條 DSL row 一個獨立 stepdef 檔；純函式各自獨立測試檔），可平行分派。

- [X] T009 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that offers a tool "{tool}" answering "{result}"`
  - Read: `chat/dsl.md` -> 對應 row（StepDef 實作語意）
  - 做：寫入 `MCP_SERVERS.{server}`（URL = fake）並啟動 fake MCP server（list `{tool}`；call 回 `{result}`）。
- [X] T010 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that never answers`
  - 做：寫入 `MCP_SERVERS.{server}` 並啟動 never-answers fake。
- [X] T011 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that is marked off`
  - 做：寫入 `MCP_SERVERS.{server}` with `ENABLED: false` + recording fake。
- [X] T012 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that requires the token "{token}" and offers a tool "{tool}" answering "{result}"`
  - 做：寫入 `MCP_SERVERS.{server}` with `TOKEN`；啟動要求 `Authorization: Bearer {token}` 的 fake。
- [X] T013 [P] [BDD-RED] stepdef: `a configured provider "{provider}" whose endpoint asks tellme to use the MCP tool "{tool}" from the server "{server}" and then answers with "{answer}"`
  - 做：script fake provider 回一個 tool-call（名稱 `mcp_{server}_{tool}`），再一次回 `{answer}`。
- [X] T014 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that advertises a tool "{tool}" with a malformed schema`
  - 做：啟動 fake 使其對 `{tool}` 廣告一個 **malformed** schema（非 object，或 `required` 無對應 `properties`）。
- [X] T015 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" whose tool "{tool}" reports a tool error`
  - 做：啟動 fake 使其對 `{tool}` 回 **MCP-level tool error**（`isError: true`）。
- [X] T016 [P] [BDD-RED] stepdef: `a remote MCP server "{server}" that fails the tool call`
  - 做：啟動 fake，於 tool call 時在 **transport** 層失敗。
- [X] T017 [P] [BDD-RED] stepdef: `the request offered the tool "{tool}" from the MCP server "{server}" alongside the agent tools`
  - 做：斷言 fake provider 記錄的 offered 定義含 `mcp_{server}_{tool}` **且** 含 native agent tools。
- [X] T018 [P] [BDD-RED] stepdef: `the request offered no tool from the MCP server "{server}"`
  - 做：斷言無 offered 名稱以 `mcp_{server}_` 起頭。
- [X] T019 [P] [BDD-RED] stepdef: `tellme called the tool "{tool}" on the MCP server "{server}"`
  - 做：斷言 fake MCP server 記錄了對 `{tool}` 的呼叫。
- [X] T020 [P] [BDD-RED] stepdef: `tellme reported on stderr that the MCP server "{server}" could not be reached`
  - 做：斷言 `stderr` 含單一來源的 skip warning（點名 `{server}`）；且非 frozen class phrase。
- [X] T021 [P] [BDD-RED] stepdef: `tellme never contacted the MCP server "{server}"`
  - 做：斷言 recording fake 記錄 **零** 連線。
- [X] T022 [P] [BDD-RED] stepdef: `the MCP server "{server}" received the token "{token}"`
  - 做：斷言 fake 記錄 `Authorization: Bearer {token}`。
- [X] T023 [P] [BDD-RED] stepdef: `the run continued past the failed MCP tool call`
  - 做：斷言失敗的 MCP tool call 以 **recoverable tool result** 回饋、loop 續行至最終答案（run 不中止；**非** `the tool request failed`）。
- [X] T024 [P] [UNIT] Config validation: `MCP_SERVERS` rules
  - Read: `research.md` -> `Decision 6`
  - 做：表驅動 unit —— key 格式、URL 必填、auth mode 合法 + 憑證要求、`TIMEOUT ≥ 0`、`ENABLED` 預設 true；**TD4**：**unmodelled sub-keys（`REQUIRES_CONSENT`/`ARGS`/`DIR`/`ENV`）被容忍（忽略）**；**`COMMAND`-shaped entry → warn+skip（非 fatal）**；只有 malformed **remote** entry → 既有 configuration-invalid class phrase（fatal）。
- [X] T025 [P] [UNIT] Auth resolution + token-not-logged
  - Read: `research.md` -> `Decision 4`
  - 做：表驅動 unit —— `auto`/`gh`/`bearer`/`basic`/`none` 的解析（explicit token 優先；GitHub-hostname→token 來源；匿名）；斷言 token 不落入任何 log。
- [X] T026 [P] [UNIT] Deterministic naming (namespacing + 64-byte hash truncation)
  - Read: `research.md` -> `Decision 5`；T005 helper
  - 做：unit —— `mcp_<server>_<tool>`；超長 → **bytes（`len()`）** 推導預算截斷 + 8-hex SHA-256；`server=24 + very long tool` 的結果 `len() ≤ 64`；兩個同名 tool 跨 server 不碰撞。
- [X] T027 [P] [UNIT] Discovery fast-fail bound + skip + sorted registration
  - Read: `research.md` -> `Decision 3`
  - 做：unit —— enabled server 併發探測；never-answers → 在 bound 內 skip 並 warning；off → 不連線；註冊順序 sorted by server key。
- [X] T028 [P] [UNIT] Fold additions: schema normalization (B1), tool-call failure semantics (TD1), bounded token resolution (B3), name byte-budget (TD3), timeout clamp (TD5), nil-args normalization (TD2/R2)
  - Read: `research.md` -> `Decision 8`–`Decision 10`；`spec.md` -> FR-018..FR-021
  - 做：表驅動 unit —— (a) `normalizeMCPSchema` 對 malformed/absent/non-object → freeform，輸出滿足 `required ⊆ properties`；不可安全化 → error（skip+warn）；(b) 一次 MCP tool call 的 **tool-level error**、**transport failure**、與 **對已關閉 client 的呼叫** 皆回一個 **nil-error `ToolResult`**（run 不中止；**非** `the tool request failed`）——並 pin **`CallTool` 對任何 call-time 失敗皆 nil-error** 的簽章契約（R3）；(c) `tokenResolver` 逾時/失敗 → warn + anonymous，且測試**不 spawn `gh`**（inject fake resolver）；(d) name 預算：`server=24 + very long tool` 的結果 `len() ≤ 64`；(e) MCP tool timeout default **300 s**、server `TIMEOUT` 由 round-024 **fixed 7200 s** timeout ceiling clamp（**R1**：非 token-bound ceiling）；(f) **nil args → `{}`** 於 wire 前正規化（R2）。
  - 不做：不寫產品碼（產品在 Phase 4）；不為轉綠放寬 assertion。

- [X] T029 subagent review (phase quality gate)
  - Read: `tests/e2e/**`、`internal/**/*_test.go`、`chat/dsl.md` -> round-032 rows
  - 檢驗：**15** 條 row 各有 stepdef 且 **0 undefined steps**；`[UNIT]` 覆蓋 **T024–T028**；只動測試層。有 issues 修正再 review，直到零問題。

> Phase-3 review executed by the orchestrator (no parallel-subagent substrate in this session): `<test command>` reports 0 undefined steps; `[UNIT]` suites present for validation/auth/naming/discovery/fold-additions. (Recorded at implementation time.)

## Phase 4: Feature GREEN & Refactor

### Phase 4A — Feature: using tools from a remote MCP server (ADD)

**Test Scope**: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`.

**Shared Must Read**: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`；`chat/dsl.md` -> `## Given/Then (round 032)`；`specs/truth/techstack.md` -> MCP Client rows；`research.md` -> `Decision 1`–`Decision 11`。

**Boundary**: 產品碼限於 `internal/domain/tools/mcp_client.go`、`internal/infrastructure/mcp/**`、`internal/config/**`、`internal/infrastructure/di/**`、`internal/cli/**` 與 `Makefile`（gate）；SDK import 只可在 `internal/infrastructure/mcp/`；discovery 只在 prompt path；不得改 native tool set / 順序模型 / class phrase / exit code / `stdout`。

- [X] T030 [BDD-GREEN] Implement the MCP client + discovery + registration to pass the ADD'ed feature
  - Read: `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`；`research.md` -> `Decision 1`–`Decision 11`；`specs/truth/techstack.md` -> MCP Client rows
  - 做：實作 T002–T006 的骨架：SDK adapter（Streamable HTTP）+ credential 解析（含 **bounded + injectable `tokenResolver`** seam，B3）；typed `MCP_SERVERS` + 驗證（TD4 容忍）；**schema normalization/well-formedness 於 offering 前**（B1）；在 prompt path 上以小固定 fast-fail bound 併發 discovery（skip+warn、`ENABLED` 跳過、prompt-path-only），以 sorted 順序把 tools 以 `mcp_<server>_<tool>` 註冊進既有 tool registry，並讓 loop 可呼叫之（**任何 call-time failure 皆為 recoverable result**，逾時走 FR-018）；MCP tool timeout 依 round-024 契約（default **300 s**，server `TIMEOUT` 由 **fixed 7200 s** ceiling clamp，TD5）。確認 `verify-mcp-sdk-confinement` 綠。
  - 驗證：`specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` 全綠；offline paths 仍不連網（root `tellme performs no network access` row）。
  - 不做：不改 native 六工具、不加 tool-call 併發、不為轉綠放寬 assertion。

- [X] T031 [BDD-REFACTOR] Refactor under green (naming/validation/discovery seams)
  - Read: `research.md` -> `Decision 3`, `Decision 5`, `Decision 6`, `Decision 8`
  - 做：在綠燈下整理 —— 把 fast-fail bound 抽為單一具名常數；single-source 的 skip-warning 訊息（不可達與 schema-skip 各一）；把 validation 子檢查分組；保持語意不變、gate 續綠。
  - 不做：不擴大重構範圍、不改行為。

### Phase 4B — Regression

- [X] T032 [REGRESSION] Full regression + falsifiability witnesses + `make verify` + topology audit
  - Read: `research.md` -> `Decision 7`；`specs/truth/techstack.md` -> MCP Client + Local fake MCP server rows
  - 做：
    - `go test -count=1 ./...` 全綠（unit + godog E2E）。
    - Witnesses（可偽性，觀察後還原）：(a) 令 never-answers 的 fast-fail bound 失效 → non-stall 案例失敗（涵蓋 FR-010 的時間界）；(b) 移除 `ENABLED` 跳過 → off 案例失敗；(c) 移除 namespacing → offered 案例失敗。
    - `make verify` **OK**（含 `verify-mcp-sdk-confinement`）；`gofmt -l .` clean；Gherkin/DSL 拓樸稽核 **PASSED**（`--root specs/truth/features/cli`：43 features · 288 module rows · 1492 steps）。
    - offline paths 不連網（差異見證）；native 六工具 / class phrase / exit code / `stdout` byte-exact 不變。
    - 記錄對**真實** remote MCP endpoint 的**手動** closeout 確認（非 gate）。
  - 不做：不放寬 assertion；不為轉綠移除見證。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> `### MCP Client` (protocol library / registry / credential / discovery / schema / timeout rows) | T001（SDK）、T003、T004、T006、T024–T027、T030 | PASS |
| `specs/truth/techstack.md` -> `### Testing & Verification` (Local fake MCP server row) | T008（fake）、T009–T023、T030 | PASS |
| `specs/truth/techstack.md` -> `### Not Introduced Yet` (stdio / caching / MEMORY / `-d`) | 負向決策；T030 邊界明示不引入 | PASS |
| `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`（ADD） | T009–T023（RED）、T030/T031（GREEN/REFACTOR）、T032 | PASS |
| `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 032)`（8 rows） | T009–T016、T024 | PASS |
| `specs/truth/features/cli/chat/dsl.md` -> `## Then (round 032)`（7 rows） | T017–T023 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY `techstack.md` | T001、T003、T004、T024–T027、T030、T032 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP (`specs/truth/contracts/**`) | 豁免（NOOP 不建任務；T030 邊界不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP (`specs/truth/data/**`) | 豁免（NOOP 不建任務；無資料變更；caching deferred） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD (`specs/truth/features/cli/chat/**`) | T009–T023、T030 | PASS |
| `research.md` -> Decision 1（remote HTTP only） | T004、T024（COMMAND warn+skip）、T030 | PASS |
| `research.md` -> Decision 2（SDK confined behind port + gate） | T001、T002、T003、T007、T030 | PASS |
| `research.md` -> Decision 3（non-stall bound + ENABLED + prompt-path-only） | T006、T010、T011、T020、T021、T027、T030、T032 | PASS |
| `research.md` -> Decision 4（auth parity；token not logged；bounded seam） | T012、T022、T025、T028、T030 | PASS |
| `research.md` -> Decision 5（deterministic naming） | T005、T017、T018、T026、T030 | PASS |
| `research.md` -> Decision 6（typed config + validation；no new failure class） | T004、T024、T030 | PASS |
| `research.md` -> Decision 7（hermetic fake + manual live check） | T008、T032（手動 closeout 記錄） | PASS |
| `research.md` -> Decision 8（MCP schema normalization/well-formedness；B1） | T003（schema.go）、T008（malformed mode）、T014、T028、T030 | PASS |
| `research.md` -> Decision 9（tool-call failure recoverable；TD1） | T015、T016、T023、T028、T030 | PASS |
| `research.md` -> Decision 10（timeout default 300 s / fixed 7200 s ceiling；TD5） | T028、T030 | PASS |
| `research.md` -> Decision 11（SDK-built fake home；B2；TD8；TD7） | T007（gate scope）、T008（mcptest）、T030 邊界 | PASS |
| `research.md` -> TD4（tolerant sub-keys + `COMMAND` warn+skip） | T004、T024、T030 | PASS |
| `research.md` -> B3（bounded `tokenResolver` + seam） | T006、T028、T030 | PASS |
| `spec.md` -> US1（FR-001–007, NFR-001） | T004、T005、T006、T009–T018、T024–T026、T030 | PASS |
| `spec.md` -> US2（FR-008–010, NFR-002） | T006、T010、T020、T027、T030、T032 | PASS |
| `spec.md` -> US3（FR-011–012, NFR-003） | T011、T018、T021、T024、T030 | PASS |
| `spec.md` -> 全域（FR-013–017, NFR-004–005） | T004、T007、T025（token not logged）、T030（prompt-path-only、no new failure class）、T032 | PASS |
| `spec.md` -> FR-018..FR-021（tool-call failure / schema / token / timeout） | T015、T016、T023（FR-018）、T028、T030 | PASS |
| `spec.md` -> 邊界情況（malformed server data、long name、同名 cross-server、no `MCP_SERVERS` or all-off、tool error、timeout） | T014、T024、T026、T027、T030、T032 | PASS |
| `spec.md` -> SC-001–SC-005 | T017–T023、T027、T030（SC-001/002/003）、T025（SC-004）、T032（SC-005） | PASS |
| `plan.md` -> Structure（port / adapter / config / di / cli / Makefile / fake；`go.mod` CHANGED） | T001–T008、T030 | PASS |
| `plan.md` -> Scope notes（api/data NOOP；`/axb-ui-plan` skipped） | T030 邊界、T032 | PASS |
| operator 拍板（Q1 remote only；Q2 fast-fail + ENABLED；Q3 SDK confined；Q4 auth parity + naming；Q5 MCP only） | T001（SDK）、T004/T010/T011/T020/T021/T027（Q2）、T007（Q3）、T012/T022/T025（Q4）、T030 邊界（Q5） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
