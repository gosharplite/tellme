# Tasks: MCP tool-call reason + a universal "no reason, no go" gate (round 056 — `/axb-tasks`)

**Plan Package**: `specs/plans/056-mcp-tool-call-reason`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md` (*MCP Client*: **MCP tool-call reason envelope**; *Reasoning & Provider Transport*: **Agent tool loop**, **Agent tool schemas**), `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`, `specs/truth/features/cli/chat/requiring-a-reason-to-call-a-tool.feature`, `specs/truth/features/cli/chat/watching-the-tool-loop.feature`, `specs/truth/features/cli/chat/dsl.md`, `specs/plans/056-mcp-tool-call-reason/features/acceptance/**`, `docs/decisions/0025-mcp-tool-call-reason.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- 本輪 `research.md` 已拍板 Decisions（D1–D7）與 `specs/truth/**` 非 NOOP 項目（`techstack.md` 三 row + `chat/**` feature 與 `chat/dsl.md`）必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：system prompt / persona（S-4）；remote MCP server 的 advertised definition（name/description/input schema）**逐字不變**（S-1）；native tools 的**宣告式 schema**（`resourceSchema` / inline command schema）；native `Execute`；reason 的渲染（fold/trim/sanitize/cap；空 reason 不出 row）；`[Tool Action]` / payload estimate / `turns.log`；round-032 的 MCP 行為（命名、discovery fast-fail、disabled server、malformed-schema skip、token 解析、失敗不 abort）；`go.mod` / `go.sum`。

## Round-056 locked decisions (implementation constraints — MUST)

> 來自 `spec.md`（S-1…S-7 · I-1…I-5 · FR-001…FR-010 · SC-001…SC-008）、`research.md` D1–D7 與 **ADR 0025** D1–D5。

- **[ENVELOPE / Q1 → A, ADR 0025 D1]** 對每個 MCP tool，**offered declaration** 改為 tellme **自己的** schema：`{ type:"object", properties:{ reason:{type:"string",…}, MCP_PAYLOAD: <server 的 advertised input schema **逐字**> }, required:["reason"] }`。`reason` 必填、由 tellme 擁有；`MCP_PAYLOAD` **不必填**（server tool 可無參數）。**server schema 只被「定位」為 `MCP_PAYLOAD` 的 subschema，不得增/刪/改名任何 property**。**無** system-prompt 變更。
- **[FORWARD-ONLY-PAYLOAD / S-3, I-2, ADR 0025 D2]** `mcp.Tool.Execute` 只把 **`MCP_PAYLOAD` 物件**交給 `CallTool`；外層 `reason` / `MCP_PAYLOAD` key **不得**送到 server。**缺 `MCP_PAYLOAD` 且 `reason` 有效 ⇒ 空 payload `{}`**。**其他任何 top-level key**，或 `MCP_PAYLOAD` 非 JSON object ⇒ **shape violation ⇒ 拒絕**（**不接觸 server**；回可恢復結果要求模型以 envelope 重試）。
- **[UNIVERSAL GATE / Q2 → B 一般化, S-5/S-7, I-5, #121 fold]** 在 loop 的**單一執行點**（`for _, tc := range resp.ToolCalls`）對**每個** tool call 套用 *no reason, no go*：`reason` 缺失或**不可渲染** ⇒ **不執行**該 tool，回一個**可恢復**的 `tool` 角色結果（nil-error，要求模型帶 `reason` 重試；**非** terminal request-level failure）。
- **[SINGLE OWNER / FR-009, I-5]** gate **必須重用** round-046 的**單一擁有者**：loop 已持有的 `agentport.ToolLineRenderer`（`AgentLoop.Lines`）之 `ReasonLine` → 由 **一次** `toolReasonText` 求值同時回傳「行」與「renders」。**不得**新增第二個 reason-presence predicate；MCP adapter 只加 envelope check。**gating by `renders`** ⇒ escape-only（present 但不可渲染）的 reason 亦**被拒**。
- **[NIL RENDERER / ADR 0025 boundary]** `Lines == nil`（round-031 assembler gate / offline `--tool-usage`，無 prompt turn）⇒ **不套 gate**；production prompt turn 一定有 `Lines`。
- **[NO-RENDER-DIVERGENCE / FR-002]** reason 仍由既有單一路徑渲染（`ui.ToolLineRenderer` / `FormatToolReason`）；**不得**加 MCP 專屬 renderer。
- **[FAILED-NOT-ABORTED]** MCP 的 call-time failure（含**新的拒絕**）一律為**可恢復**結果（nil-error），run 不 abort（FR-007）。拒絕的 call **不執行** tool：其 round-026 accounting 分類為 recorded forward item（RF-056-3），本輪**不**新增分類語意。
- **[NO-DEP]** 無新相依；`go.mod` / `go.sum` 不動；POSIX-only。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D5/D6；無新相依；envelope 與 gate 皆為現有 Go 檔的修改）。不得把 helper/fixture 塞進 Setup。

## Phase 2: Foundational

**Goal**: 建立測試落點骨架與共用 fixture/helper（Zero Shared Edits），供 Phase 3 並行分派；不寫產品行為、不寫 assertion。

- [X] T001 [FOUNDATIONAL] 建立本輪新 stepdef 落點骨架（Zero Shared Edits）
  - Read:
    - `truth-delta.md` -> `/axb-dsl-refine` 的 ADD / MODIFY rows
    - `specs/truth/features/cli/chat/dsl.md` -> 本輪新句（S2/S3/S4/S5/S6/S7）
  - 只做：為每條新句建立**獨立** `tests/e2e/steps/step_r056_t0NN_*.go` 骨架（每檔一個 sentence 的 `ctx.Given/Then` 空實作 + `// T0NN [BDD-RED]` 標頭），彼此不同檔，消除並行同檔衝突。
  - 不做：不寫 `dsl.md` 語意、不寫 assertion、不改產品碼、不改既有 stepdef 檔。

- [X] T002 [FOUNDATIONAL] 擴充 MCP 測試 helper：記錄 server 收到的 arguments + 供 reason
  - Read:
    - `tests/e2e/steps/mcp_helpers.go`（現況 fake server helper）
    - `specs/truth/features/cli/chat/dsl.md` -> `the MCP server "{server}" received only the arguments its tool expects`
  - 只做：在 fake MCP server helper 上新增「記錄每次 tool call 收到的 arguments（原始 JSON）」的能力與讀取器；保留既有行為。
  - 不做：不改 production `internal/infrastructure/mcp/**`；不改既有 MCP stepdef 的語意。

- [X] T003 [FOUNDATIONAL] 建立 loop gate 的單元測試落點 + fake renderer helper
  - Read:
    - `internal/agent/agentloop.go`（現況：`Lines` port、`logAction`、`reasonsOf`、`toolReason`）
    - `internal/domain/agent/`（`ToolLineRenderer` port 定義）
  - 只做：新增 `internal/agent/agentloop_reason_gate_test.go` 骨架 + 一個可注入的 **fake `ToolLineRenderer`**（可回傳 `renders=false`），使 Phase 3/4 能 pin `Lines` 驅動的拒絕決策。
  - 不做：不改產品碼；不改 `internal/ui`。

## Phase 3: Test Alignment & Implementation

**Goal**: 在本輪 Feature 之前，先把受影響 DSL 的自動化測試對齊最新版 truth（含 truth-delta 有改的句，以及本輪 Feature 用到、尚無 stepdef 的句）。不寫產品行為。

**DSL 參照**:
- 本輪各句都在同模組 `specs/truth/features/cli/chat/dsl.md`。**不使用**介面根共用 DSL row。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`、`再讀確認`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE；語意以 `dsl.md` 該列為準。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`（S1：既有 MCP 呼叫 Given 的 arguments 改為 round-056 envelope）。
- `[BDD-REMOVE]`：`DELETE`（S8：round-039 的 `the action of the call without a reason begins after a blank line` 已不可達 ⇒ 移除其 stepdef）。
- `[BDD-RED]`：`ADD` 或本輪 Feature 用到、尚無 stepdef 的句（S2–S7）。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 本輪全部 8 句（S1–S8）
- `truth-delta.md` -> `/axb-dsl-refine` 的 ADD / MODIFY / DELETE 句
- `tests/e2e/steps/step_r032_t013.go`（S1 的既有落點）
- `tests/e2e/steps/step_r039_*.go`（S8 的既有落點）

**Boundary**:
- 一條 DSL 一個 task；只改該句的 stepdef / assertion / 直接依賴的 helper。
- 落點檔案採獨立檔案（Zero Shared Edits）。
- 不寫產品碼。review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step。

**Parallel Hint**:
- T004–T011 各派一個獨立 subagent；T012 等全部回來再啟動 subagent review。

- [X] T004 [P] [BDD-ALIGN] `a configured provider "{provider}" whose endpoint asks tellme to use the MCP tool "{tool}" from the server "{server}" with the reason "{reason}" and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r032_t013.go`（現況：`Arguments: "{}"`）→ 改為 round-056 envelope `{"reason":"{reason}","MCP_PAYLOAD":{}}`。

- [X] T005 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint first asks tellme to use the MCP tool "{tool}" from the server "{server}" without a reason and then with the reason "{reason}" and then answers with "{answer}"`
  - Read: T001 的落點檔。

- [X] T006 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint first asks tellme to create the file "{path}" without a reason, then with the content "{content}" and the reason "{reason}", and then answers with "{answer}"`
  - Read: T001 的落點檔。

- [X] T007 [P] [BDD-RED] `the MCP server "{server}" received only the arguments its tool expects`
  - Read: T002 的 helper + T001 的落點檔。

- [X] T008 [P] [BDD-RED] `the request offered the tool "{tool}" from the MCP server "{server}" with a reason and the server's declared schema`
  - Read: T001 的落點檔；`specs/truth/features/cli/chat/dsl.md` 該列。

- [X] T009 [P] [BDD-RED] `the run asked for a reason before running a tool`
  - Read: T001 的落點檔。

- [X] T010 [P] [BDD-RED] `the run reported the reason "{reason}" for the MCP tool "{tool}" on the server "{server}"`
  - Read: T001 的落點檔；`tests/e2e/steps/step_r021_t027_chat_then_reason_echo.go`（既有 reason-row matcher 可重用）。

- [X] T011 [P] [BDD-REMOVE] `the action of the call without a reason begins after a blank line`
  - Read: 其現況 stepdef 檔；移除該句的 stepdef（Example 已於 `/axb-dsl-refine` 刪除，該句不可達）。

- [X] T012 subagent review (phase quality gate)
  - 確認 S1–S8 全部落地、無 undefined step；無產品碼。

## Phase 4A: ADD Feature File - cli/chat/requiring-a-reason-to-call-a-tool.feature

**Goal**: 實作**通用** reason gate，使新 feature 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/requiring-a-reason-to-call-a-tool.feature` -> `Feature: Requiring a reason to call a tool`
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 8 句
- `truth-delta.md` -> `/axb-dsl-refine` ADD rows
- `docs/decisions/0025-mcp-tool-call-reason.md` -> D2/D3

**Boundary**:
- 產品變更**限** `internal/agent/agentloop.go`（gate）；重用 `AgentLoop.Lines` 的 `ReasonLine`。
- `Lines == nil` ⇒ 不套 gate；reason 缺失或 `renders=false` ⇒ 不執行 + 回可恢復結果。
- **不得**新增第二個 reason-presence predicate；**不得**改渲染、`[Tool Action]`、estimate、`turns.log`。

**Test Scope**:
- `specs/truth/features/cli/chat/requiring-a-reason-to-call-a-tool.feature`

- [X] T013 [BDD-GREEN] 讓 Test Scope 全綠（loop gate + MCP envelope 的拒絕路徑）
- [X] T014 [BDD-REFACTOR] 收斂 gate 實作（單一謂詞呼叫點、註解指向 ADR 0025 D2/D3），測試保持全綠；並補 `internal/agent/agentloop_reason_gate_test.go` 的單元 pin（fake renderer：`renders=false` ⇒ 拒絕）

## Phase 4B: MODIFY Feature File - cli/chat/using-tools-from-a-remote-mcp-server.feature

**Goal**: 讓 MCP feature 全綠：offered declaration 為 tellme 的 envelope；server 只收到 payload。

**Shared Must Read**:
- `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` -> `Feature: Using tools from a remote MCP server`
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 8 句
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY rows
- `docs/decisions/0025-mcp-tool-call-reason.md` -> D1/D2

**Boundary**:
- 產品變更**限** `internal/infrastructure/mcp/tool.go`（envelope 建構 + `Execute` 只轉送 `MCP_PAYLOAD` + envelope shape check）。
- server 的 advertised schema **逐字**作為 `MCP_PAYLOAD` subschema；round-032 的 normalization / skip-with-warning 契約不變。
- **不得**修改 server definition、system prompt、native schemas。

**Test Scope**:
- `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature`

- [X] T015 [BDD-GREEN] 讓 Test Scope 全綠（envelope + forward-only-payload）
- [X] T016 [BDD-REFACTOR] 收斂 `mcp/tool.go`（envelope builder 單一函式；shape check 明確），測試保持全綠；並補 envelope 的單元 pin（`Parameters()` 的 `MCP_PAYLOAD` subschema 與 server schema 逐字相等）

## Phase 4C: DELETE Feature / DSL Truth - cli/chat/watching-the-tool-loop.feature (the retired round-039 Example)

**Goal**: 確認被移除的 Example 與其 stepdef 不留殘骸，且理由條（reason row）不再被 reason-less 情境保護。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` -> `Feature: Watching the tool loop work`
- `specs/truth/features/cli/chat/dsl.md` -> `the action of the call without a reason begins after a blank line`（RETIRED）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY row（watching feature）

**Boundary**:
- 只移除已不可達的 stepdef / 支援 helper（若無其他引用）。
- **不得**改動其他 tool-loop 行（reason/result/action/engine/output）的語意。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [X] T017 [CODE-REMOVE] 移除 `the action of the call without a reason begins after a blank line` 的 stepdef 及其專用 helper（若無其他引用）
- [X] T018 [REGRESSION] 跑 Test Scope + `go test -count=1 ./tests/e2e/`，確認新版 truth 成立（無 undefined step、無殘餘斷言）

## Phase 5: Verification & Evidence

**Goal**: 產出可偽性見證與 gate 全綠證據；更新 `STATUS.md`。

- [X] T019 [EVIDENCE] 可偽性見證（重現後還原，不改 committed 檔）
  - Read: `research.md` D7；`spec.md` SC-001/SC-002/SC-006/SC-007。
  - 做：
    - (a) MCP reason row：freeze envelope ⇒ `using-tools-…` 的 reason Then 紅。
    - (b) payload purity：把 `Execute` 改成轉送整個 map ⇒ `received only the arguments…` 紅。
    - (c) verbatim schema：在 envelope 內改動 server schema ⇒ `with a reason and the server's declared schema` 紅。
    - (d) MCP refusal：移除 adapter 的 envelope check ⇒ `requiring-…` 的第一個 Example 紅。
    - (e) universal refusal：移除 loop gate ⇒ `requiring-…` 的第二個 Example 紅（`write_file` 在無 reason 時被執行）。
    - (f) single-owner：gate 由 `Lines.ReasonLine` 決策（fake renderer `renders=false` ⇒ 拒絕）。
  - 不做：不改 committed 檔；不以 flaky 手段求綠。

- [X] T020 [GATE] round 自身 gate 與不變量檢查
  - Read: `spec.md` SC-004/SC-005/SC-008；`plan.md` I-1…I-5。
  - 做：`gofmt -l .` 乾淨；`go vet ./...` 乾淨；`make verify` **OK**（含 `verify-architecture` baseline header-only）；`go test -count=1 ./...` 綠（含 `tests/e2e`，全部 Examples）；`git diff origin/dev -- go.mod go.sum` 為空。
  - 不做：不弱化任何 gate。

- [X] T021 [CLOSE] 更新 `STATUS.md` 並交付 PR
  - Read: `STATUS.md`（現況）。
  - 做：以最少量更新記 round 056 in flight（branch、PR）、roadmap 一列、並把 #121 的 fold 記入 open items；commit + push；開 PR（human merge）。
  - 不做：不改 frozen plan packages；不代為 merge。

## Pre-Delivery Orphan Coverage Sweep

**0 orphans.** `truth-delta.md` 的非 NOOP 項目（`techstack.md` 三 row：**MCP tool-call reason envelope**、**Agent tool loop**、**Agent tool schemas**）皆由 T013/T015 的交付或 `research.md` D1–D4 承載；`/axb-dsl-refine` 的 feature/dsl 變更（`using-tools-…` MODIFY、`requiring-…` ADD、`watching-…` MODIFY、`dsl.md` 8 句）由 T004–T018 全量覆蓋；`research.md` 已拍板 D1–D7 全部被 T001–T021 引用；governance（**ADR 0025**）由 T013/T015 的 `Read` 引用。無孤立產物。


## Outcome (`/axb-implement`, 2026-09-19)

**Product**
- **T013/T014** — `internal/agent/agentloop.go`: `refuseReasonless()` applies the universal *no reason, no go* gate at the loop's single execution site — a call whose reason does not render is REFUSED (a recoverable nil-error `tool` message; the tool does not execute). It **reuses the round-046 single owner** (`Lines.ReasonLine`) — no second predicate; a nil `Lines` renderer means no gate (ADR 0025 boundary). Unit pins: `agentloop_reason_gate_test.go` (refuse + happy path + nil-renderer).
- **T015/T016** — `internal/infrastructure/mcp/tool.go`: `Parameters()` now returns tellme's own envelope (`reason` required + `MCP_PAYLOAD` = the server's advertised schema **verbatim**); `Execute` validates the envelope (`unwrapEnvelope`), forwards **only** the `MCP_PAYLOAD` object, treats an absent payload as `{}`, and REFUSES a shape violation (extra key / non-object payload) **without contacting the server**. Unit pins: `tool_r056_test.go` (verbatim-schema, forward-only-payload, absent-payload, refusal).

**Tests / alignment (Phase 3)**
- T004 — `step_r032_t013.go` retargeted to the reason-carrying envelope.
- T005–T010 — new stepdefs: the MCP reason envelope, the refusal-then-retry (MCP + native), the payload-purity Then, the offered-envelope Then, the "asked for a reason" Then, and the MCP-aware reason row.
- T011 — the retired round-039 sentence's stepdef removed (`step_r039_blank_lines.go`).
- Shared read fixture (`readFilesArgs`) now carries a reason (the gate would otherwise refuse reason-less reads); mcptest records tool-call arguments (`ReceivedArguments`/`AllReceivedArguments`).
- **Truth correction during implementation**: `requiring-a-reason-to-call-a-tool.feature`'s MCP Example asserted the *native* reason sentence for a namespaced tool; corrected to the MCP-aware sentence (S7).

**Evidence (T019)** — falsifiability witnesses reproduced then reverted:
- (b) forwarding the `reason` to the server ⇒ the payload-purity + refusal Examples RED (2 scenarios).
- (e) removing the loop gate ⇒ BOTH refusal Examples RED.
- (d) weakening the envelope check (accepting a stray key / non-object payload) ⇒ `TestTool_ExecuteRefusesShapeViolationWithoutContactingServer` RED.
- (f) the gate is driven by `Lines.ReasonLine` (unit pin: a fake renderer's `renders=false` refuses; a nil renderer does not).

**Gate (T020)** — `gofmt -l .` clean · `go vet ./...` clean · `make verify` **OK** (arch baseline header-only; lint 0; govulncheck clean; cross-compile 4/4) · `go test -count=1 ./...` green (**243 scenarios · 1796 steps**, 0 undefined, no failures) · the Gherkin/DSL topology audit is back to the **5 pre-existing errors** (round 056 adds none) · `go.mod`/`go.sum` unchanged.

**T021** — `STATUS.md` updated (round 056 in flight) + **PR [#122](https://github.com/gosharplite/tellme/pull/122)** opened (human merge).
