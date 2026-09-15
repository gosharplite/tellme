# Tasks: Tool-Loop Log Line Reshape (round 022)

**Plan Package**: `specs/plans/022-tool-loop-log-line`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions（1–8）與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod`／`go.sum` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/agent/agentloop.go`（重塑 `logStep` + `Now` clock seam）、`internal/ui/toollog.go`（新增 formatter）、`internal/cli/cli.go`（`loop.Now` + tool-using turn 於答案前輸出一個空行）。維持「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。**`stdout` 維持 byte-exact；class-phrase 詞彙不變（11）。**

---

## Phase 2: Foundational

**Goal**: 建立本輪 formatter／clock seam 與 stepdef／`[UNIT]` 落點骨架（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔案或 seam。只建立落點與載體，不寫行為。

- [ ] T001 建立 `internal/ui` tool-log formatter 落點與 `AgentLoop` clock seam 骨架
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Pure-helper unit tests）
    - `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2
    - `internal/ui/status.go`、`internal/ui/turn.go`（既有 clock-seam formatter 樣式）、`internal/agent/agentloop.go`、`internal/cli/cli.go`（`env.now`）
  - 只做：新增 `internal/ui/toollog.go`，宣告 `func FormatToolLog(t time.Time, name, reason string) string` 空殼（`return ""`）；在 `AgentLoop` 新增 `Now func() time.Time` 欄位（含 nil→`time.Now` 的 `now()` helper 空殼）。
  - 不做：不實作格式字串；不改 `logStep` 輸出；不動 CLI blank-line；不寫斷言。

- [ ] T002 建立 stepdef／`[UNIT]` 落點檔骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 3 個新 Then 句 + 2 個 MODIFY 句）
    - `tests/e2e/steps/register.go`、`tests/e2e/steps/scenario_context.go`
    - `internal/ui/toollog.go`、`internal/agent/agentloop_reason_test.go`
  - 只做：建立 3 個新 Then 的獨立 stepdef 檔骨架（`tests/e2e/steps/step_r022_t005_chat_then_tool_no_reason.go`、`step_r022_t006_chat_then_one_line_per_call.go`、`step_r022_t007_chat_then_tool_report_separated.go`，各自 `init()` 自我註冊空白 registrar，對到 T005–T007）；建立 `[UNIT]` 落點 `internal/ui/toollog_test.go` 空殼。
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth：先對齊兩個 MODIFY 的 tool-loop log 句到 `[HH:MM:SS] [Tool] …` 形狀，再為 3 個新句建立失敗訊號；最後補非 DSL 的 `[UNIT]`。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪無跨模組新句；frozen class-phrase 詞彙維持 **11**）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：本輪 **2 句**（`MODIFY` 的既有 tool-loop log Thens）。
- `[BDD-RED]`：本輪 **3 句**（`ADD` 的新 Thens）。
- `[UNIT]`：`internal/ui` formatter 與 `AgentLoop.logStep` 行形狀（非 DSL）。
- 兩者都只動測試層，不寫產品碼（除 T001 已預留的骨架）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 5 句：`the run reported the reason "…" for the tool call "…"`、`the run reported the tool call "…" on its diagnostic output`、`the run reported the tool call "…" without a reason`、`the run reported one tool-loop log line for each tool call`、`the tool report is separated from the answer`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/fakeprovider.go`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點採 stdlib `testing`；`internal/ui/toollog_test.go`、`internal/agent/agentloop_reason_test.go`。
- 不寫產品碼（除 T001 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T009 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；T003–T004 觸及的既有 stepdef 檔與其他 task 不同檔，可並行）；T010 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN（既有 tool-loop log 句對齊新形狀）

- [ ] T003 [P] [BDD-ALIGN] `Then: the run reported the reason "{reason}" for the tool call "{tool}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（新形狀 `[HH:MM:SS] [Tool] <tool> - <reason>`）
  - Landing: `tests/e2e/steps/step_r021_t027_chat_then_reason_echo.go`
  - 動作：改為斷言 `stderr` 帶 `[HH:MM:SS] [Tool] <tool name> - <reason>` 形狀的行（reason 為 ` - ` 尾），且**不含** `arguments=`／`result=`。

- [ ] T004 [P] [BDD-ALIGN] `Then: the run reported the tool call "{tool}" on its diagnostic output`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（新形狀 `[HH:MM:SS] [Tool] <tool name>`）
  - Landing: `tests/e2e/steps/step_t016_chat_then_reported_tool_call.go`
  - 動作：改為斷言 `stderr` 帶 `[HH:MM:SS] [Tool]` 前綴且命名 `{tool}` 的行；不含 `arguments=`／`result=`。

### BDD-RED（本輪新增句型 — Then）

- [ ] T005 [P] [BDD-RED] `Then: the run reported the tool call "{tool}" without a reason`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r022_t005_chat_then_tool_no_reason.go`
  - 語意：`stderr` 帶 `[HH:MM:SS] [Tool] <tool name>` 行命名 `{tool}` 且**無** reason 尾（無 ` - `）。

- [ ] T006 [P] [BDD-RED] `Then: the run reported one tool-loop log line for each tool call`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r022_t006_chat_then_one_line_per_call.go`
  - 語意：`[HH:MM:SS] [Tool] …` 行的數量等於該回合工具呼叫數（每呼叫恰一行，不合併、不重複）。

- [ ] T007 [P] [BDD-RED] `Then: the tool report is separated from the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（merged capture）
  - Landing: `tests/e2e/steps/step_r022_t007_chat_then_tool_report_separated.go`
  - 語意：在 merged (`stdout`+`stderr`) capture 中，最後一條 `[HH:MM:SS] [Tool] …` 行與答案位元組之間恰隔一個空行。

### UNIT（非 DSL 的單元斷言）

- [ ] T008 [P] [UNIT] `internal/ui.FormatToolLog` formatter
  - Read:
    - `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Pure-helper unit tests）
    - `internal/ui/toollog.go`、`internal/ui/status.go`
  - 撰寫：以注入的固定時間斷言 `[HH:MM:SS] [Tool] <name> - <reason>`（含 reason）；無 reason 時為 `[HH:MM:SS] [Tool] <name>`（無 ` - ` 尾）；時間格式為 `15:04:05`。
  - 落點：`internal/ui/toollog_test.go`。

- [ ] T009 [P] [UNIT] `AgentLoop.logStep` 行形狀（新 `[Tool]` 行 + clock seam）
  - Read: `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2, 3, 4；`internal/agent/agentloop.go`
  - 撰寫：`AgentLoop` 搭配注入的 `Now`，`stderr` 帶 `[HH:MM:SS] [Tool] <name> - <reason>`；**不含** `arguments=`／`result=`；無 reason 時為 `[HH:MM:SS] [Tool] <name>`。
  - 落點：`internal/agent/agentloop_reason_test.go`（更新既有檔至新形狀）。

### Phase Review Gate

- [ ] T010 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/watching-the-tool-loop.feature`、`specs/truth/features/cli/chat/dsl.md`
    - `tests/e2e/steps/*.go`（含 `step_r022_t00{5,6,7}_*.go`、`step_r021_t027_*.go`、`step_t016_*.go`）
    - `internal/ui/toollog.go`、`internal/agent/agentloop.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；2 個 ALIGN 句已對齊、3 個新句 stepdef 已註冊；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: MODIFY Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: 讓 `watching-the-tool-loop.feature` 全綠 — tool-loop log 行重塑為 `[HH:MM:SS] [Tool] <tool name> - <reason>`（無 reason 時無 ` - ` 尾；不再 echo `arguments`／`result`），且 tool-using turn 於答案前輸出一個空行。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `the run reported the reason "…" for the tool call "…"`、`the run reported the tool call "…" on its diagnostic output`、`the run reported the tool call "…" without a reason`、`the run reported one tool-loop log line for each tool call`、`the tool report is separated from the answer`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`）+ `/axb-technical-research` MODIFY（`techstack.md` Agent tool loop）
- `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1–6

**Boundary**:
- 產品碼：`internal/agent/agentloop.go`（`logStep` 改用 `ui.FormatToolLog` + `Now` seam；移除 `arguments=`／`result=`）；`internal/ui/toollog.go`（實作 `FormatToolLog`）；`internal/cli/cli.go`（`loop.Now = env.now`；`len(result.Steps) > 0` 時於 `writeAnswer` 前輸出一個空行到 `stderr`）。
- 空行**不受 `chrome` 條件限制**（跟隨 tool-log 行；任何 tool-using 表面，含 `-i` submit 與非 chrome 路徑）。
- `stdout` 契約不變；class-phrase 詞彙不變（11）；不動 payload line／spinner；`go mod` 不變。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [ ] T011 [BDD-GREEN] 讓 Test Scope 全綠（並使 T008／T009 的 `[UNIT]` 轉綠）
- [ ] T012 [BDD-REFACTOR] 在綠燈下整理 `logStep`／`FormatToolLog`／blank-line emit 落點

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞，並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T013 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**（各觀察失敗後還原）：
    1. 還原 `logStep` 為舊的 `[tool] … arguments=… result=…` 行 → `watching-the-tool-loop.feature` 的 reason Then 失敗。
    2. 移除 `len(result.Steps) > 0` 時於答案前的空行 → `the tool report is separated from the answer` 失敗。
    3. 讓無 reason 的呼叫仍印出 ` - ` 尾 → `the run reported the tool call "read_files" without a reason` 失敗。
    4. 讓同回合的兩個呼叫只記一行 → `the run reported one tool-loop log line for each tool call` 失敗。
  - 確認：`stdout` byte-exact；exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；offline paths 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`：Agent tool loop + Pure-helper unit tests） | T001、T008、T009、T011 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務；T013 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務；T013 驗證 record 形狀不變） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`） | T003–T007、T011、T012 | PASS |
| `research.md` -> Decision 1（單行 shape） | T001、T003、T004、T005、T008、T009、T011、T012 | PASS |
| `research.md` -> Decision 2（injected clock seam / `internal/ui` formatter / `AgentLoop.Now`） | T001、T008、T009、T011 | PASS |
| `research.md` -> Decision 3（drop `arguments=`／`result=`） | T003、T004、T009、T011 | PASS |
| `research.md` -> Decision 4（no-reason ⇒ `[Tool] <name>`；generic extraction kept） | T005、T009、T011 | PASS |
| `research.md` -> Decision 5（CLI 於 tool-using turn 答案前輸出一個空行） | T007、T011、T012、T013 | PASS |
| `research.md` -> Decision 6（`stderr`；`stdout` byte-exact） | T007、T011、T013 | PASS |
| `research.md` -> Decision 7（testing strategy；no pty） | T002、T008、T009、T013 | PASS |
| `research.md` -> Decision 8（no new dependency；no schema/loop-contract change） | T011、T013（`go mod tidy` graph 不變） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-012`、`SC-001`–`SC-005` | T003–T009（對齊）、T011–T012（交付）、T013 | PASS |
| `plan.md` -> Source-code structure（`internal/agent`、`internal/ui`、`internal/cli`） | T001、T011、T012 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；CLI end → /axb-dsl-refine；ui skipped） | T001、T013 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
