# Tasks: Tool-Loop Log Line Reshape (round 022)

**Plan Package**: `specs/plans/022-tool-loop-log-line`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

> **PR #50 review fold** (B1/TD1/TD2/TD3/R1/R2): this revision adds the negative Rule/Then (B1), the `-i` tool-using Example (TD1), the ordered Rule/Then + sequential two-tool Given (TD3), and the `formatClock` factor (R2). B1/TD1/TD3 land as new `[BDD-RED]` sentences (T008–T010) + the `-i` Examples; TD2 is a documentation note; R2 is a Foundational/product-refactor note. **(review 2)**: `R-1` extends `formatClock` to `FormatMetrics` (T001/T011); implementation directives 1–3 (preserve `LoopObserver` `Before/AfterToolLog` hooks; blank line after `sp.Stop()` + `store.Append` and immediately before `writeAnswer`; `AgentLoop.now()` nil→`time.Now` fallback) are folded into T014.

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions（1–8）+ PR #50 review fold（B1/TD1/TD2/TD3）與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod`／`go.sum` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/agent/agentloop.go`（重塑 `logStep` + `Now` clock seam）、`internal/ui/toollog.go`（新增 formatter）、`internal/cli/cli.go`（`loop.Now` + tool-using turn 於答案前輸出一個空行）。維持「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。**`stdout` 維持 byte-exact；class-phrase 詞彙不變（11）。**

---

## Phase 2: Foundational

**Goal**: 建立本輪 formatter／clock seam 與 stepdef／`[UNIT]` 落點骨架（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔案或 seam。只建立落點與載體，不寫行為。

- [X] T001 建立 `internal/ui` tool-log formatter 落點、`formatClock` 共用、與 `AgentLoop` clock seam 骨架
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Pure-helper unit tests）
    - `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2；`## Review fold — PR #50`（R2）
    - `internal/ui/status.go`、`internal/ui/turn.go`（既有 clock-seam formatter 樣式）、`internal/agent/agentloop.go`、`internal/cli/cli.go`（`env.now`）
  - 只做：新增 `internal/ui/toollog.go`，宣告 `func FormatToolLog(t time.Time, name, reason string) string` 空殼（`return ""`）；抽出 `func formatClock(t time.Time) string`（R2）並讓既有 `FormatPayloadStatus`（`status.go`）／`FormatInputCaptured`（`turn.go`）／`FormatMetrics`（`metrics.go`）共用（行為不變，PR #50 review 2 R-1 要求一併納入 `FormatMetrics`）；在 `AgentLoop` 新增 `Now func() time.Time` 欄位（含 nil→`time.Now` 的 `now()` helper 空殼）。
  - 不做：不實作 log 行格式字串；不改 `logStep` 輸出；不動 CLI blank-line；不寫斷言。

- [X] T002 建立 stepdef／`[UNIT]` 落點檔骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 2 個 MODIFY 句 + 6 個新句）
    - `tests/e2e/steps/register.go`、`tests/e2e/steps/scenario_context.go`
    - `internal/ui/toollog.go`、`internal/agent/agentloop_reason_test.go`
  - 只做：建立 6 個新句的獨立 stepdef 檔骨架（`tests/e2e/steps/step_r022_t005_chat_then_tool_no_reason.go`、`step_r022_t006_chat_then_one_line_per_call.go`、`step_r022_t007_chat_then_tool_report_separated.go`、`step_r022_t008_chat_then_tool_calls_in_order.go`、`step_r022_t009_chat_then_tool_no_blank.go`、`step_r022_t010_chat_given_tree_then_read.go`，各自 `init()` 自我註冊空白 registrar，對到 T005–T010）；建立 `[UNIT]` 落點 `internal/ui/toollog_test.go` 空殼。
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth：先對齊兩個 MODIFY 的 tool-loop log 句到 `[HH:MM:SS] [Tool] …` 形狀，再為 6 個新句建立失敗訊號；最後補非 DSL 的 `[UNIT]`。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪無跨模組新句；frozen class-phrase 詞彙維持 **11**）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`／`權威狀態落地`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：本輪 **2 句**（`MODIFY` 的既有 tool-loop log Thens）。
- `[BDD-RED]`：本輪 **6 句**（`ADD` 的新句 — 5 個 Then + 1 個 Given）。
- `[UNIT]`：`internal/ui` formatter 與 `AgentLoop.logStep` 行形狀（非 DSL）。
- 兩者都只動測試層，不寫產品碼（除 T001 已預留的骨架）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 8 句：`the run reported the reason "…" for the tool call "…"`、`the run reported the tool call "…" on its diagnostic output`、`the run reported the tool call "…" without a reason`、`the run reported one tool-loop log line for each tool call`、`the run reported the tool calls in order "…" and "…"`、`the tool report is separated from the answer`、`the tool loop added no blank line before the answer`、`a configured provider "…" whose endpoint shows the folder tree and then reads "…" and then answers with "…"`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/fakeprovider.go`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點採 stdlib `testing`；`internal/ui/toollog_test.go`、`internal/agent/agentloop_reason_test.go`。
- 不寫產品碼（除 T001 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T012 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；T003–T004 觸及的既有 stepdef 檔與其他 task 不同檔，可並行）；T013 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN（既有 tool-loop log 句對齊新形狀）

- [X] T003 [P] [BDD-ALIGN] `Then: the run reported the reason "{reason}" for the tool call "{tool}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（新形狀 `[HH:MM:SS] [Tool] <tool> - <reason>`）
  - Landing: `tests/e2e/steps/step_r021_t027_chat_then_reason_echo.go`
  - 動作：改為斷言 `stderr` 帶 `[HH:MM:SS] [Tool] <tool name> - <reason>` 形狀的行（reason 為 ` - ` 尾），且**不含** `arguments=`／`result=`。

- [X] T004 [P] [BDD-ALIGN] `Then: the run reported the tool call "{tool}" on its diagnostic output`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（新形狀 `[HH:MM:SS] [Tool] <tool name>`）
  - Landing: `tests/e2e/steps/step_t016_chat_then_reported_tool_call.go`
  - 動作：改為斷言 `stderr` 帶 `[HH:MM:SS] [Tool]` 前綴且命名 `{tool}` 的行；不含 `arguments=`／`result=`。

### BDD-RED（本輪新增句型）

- [X] T005 [P] [BDD-RED] `Then: the run reported the tool call "{tool}" without a reason`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r022_t005_chat_then_tool_no_reason.go`
  - 語意：`stderr` 帶 `[HH:MM:SS] [Tool] <tool name>` 行命名 `{tool}` 且**無** reason 尾（無 ` - `）。

- [X] T006 [P] [BDD-RED] `Then: the run reported one tool-loop log line for each tool call`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r022_t006_chat_then_one_line_per_call.go`
  - 語意：`[HH:MM:SS] [Tool] …` 行的數量等於該回合工具呼叫數（每呼叫恰一行，不合併、不重複）。

- [X] T007 [P] [BDD-RED] `Then: the tool report is separated from the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（merged capture）
  - Landing: `tests/e2e/steps/step_r022_t007_chat_then_tool_report_separated.go`
  - 語意：在 merged (`stdout`+`stderr`) capture 中，最後一條 `[HH:MM:SS] [Tool] …` 行與答案位元組之間恰隔一個空行。

- [X] T008 [P] [BDD-RED] `Then: the run reported the tool calls in order "{tool_a}" and "{tool_b}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（TD3）
  - Landing: `tests/e2e/steps/step_r022_t008_chat_then_tool_calls_in_order.go`
  - 語意：`stderr` 帶 `[HH:MM:SS] [Tool] {tool_a}` 行**先於** `[HH:MM:SS] [Tool] {tool_b}` 行。

- [X] T009 [P] [BDD-RED] `Then: the tool loop added no blank line before the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（B1；merged capture）
  - Landing: `tests/e2e/steps/step_r022_t009_chat_then_tool_no_blank.go`
  - 語意：merged capture 中，答案位元組**不**緊接一個空行（答案前的位元組是前一診斷行的換行，而非空行）。

- [X] T010 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint shows the folder tree and then reads "{path}" and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given（TD3；三段式 tree → read → answer）
  - Landing: `tests/e2e/steps/step_r022_t010_chat_given_tree_then_read.go`
  - 語意：fake 依序回 `get_tree` 呼叫、`read_files` 呼叫（for `{path}`）、答案 `{answer}`。

### UNIT（非 DSL 的單元斷言）

- [X] T011 [P] [UNIT] `internal/ui.FormatToolLog` formatter（+ `formatClock`）
  - Read:
    - `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2；`## Review fold — PR #50`（R2）
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Pure-helper unit tests）
    - `internal/ui/toollog.go`、`internal/ui/status.go`
  - 撰寫：以注入的固定時間斷言 `[HH:MM:SS] [Tool] <name> - <reason>`（含 reason）；無 reason 時為 `[HH:MM:SS] [Tool] <name>`（無 ` - ` 尾）；時間格式為 `15:04:05`；`formatClock` 與 `FormatPayloadStatus`／`FormatInputCaptured`／`FormatMetrics` 共用同一時鐘格式（R2 + review 2 R-1）。
  - 落點：`internal/ui/toollog_test.go`。

- [X] T012 [P] [UNIT] `AgentLoop.logStep` 行形狀（新 `[Tool]` 行 + clock seam）
  - Read: `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1, 2, 3, 4；`internal/agent/agentloop.go`
  - 撰寫：`AgentLoop` 搭配注入的 `Now`，`stderr` 帶 `[HH:MM:SS] [Tool] <name> - <reason>`；**不含** `arguments=`／`result=`；無 reason 時為 `[HH:MM:SS] [Tool] <name>`。
  - 落點：`internal/agent/agentloop_reason_test.go`（更新既有檔至新形狀）。

### Phase Review Gate

- [X] T013 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/watching-the-tool-loop.feature`、`specs/truth/features/cli/chat/dsl.md`
    - `tests/e2e/steps/*.go`（含 `step_r022_t0{05,06,07,08,09,10}_*.go`、`step_r021_t027_*.go`、`step_t016_*.go`）
    - `internal/ui/toollog.go`、`internal/agent/agentloop.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；2 個 ALIGN 句已對齊、6 個新句 stepdef 已註冊；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: MODIFY Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: 讓 `watching-the-tool-loop.feature` 全綠 — tool-loop log 行重塑為 `[HH:MM:SS] [Tool] <tool name> - <reason>`（無 reason 時無 ` - ` 尾；不再 echo `arguments`／`result`），tool-using turn 於答案前輸出一個空行（**ungated**），且 used-no-tool run 不加空行。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `the run reported the reason "…" for the tool call "…"`、`the run reported the tool call "…" on its diagnostic output`、`the run reported the tool call "…" without a reason`、`the run reported one tool-loop log line for each tool call`、`the run reported the tool calls in order "…" and "…"`、`the tool report is separated from the answer`、`the tool loop added no blank line before the answer`、`a configured provider "…" whose endpoint shows the folder tree and then reads "…" and then answers with "…"`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`）+ `/axb-technical-research` MODIFY（`techstack.md` Agent tool loop）
- `specs/plans/022-tool-loop-log-line/research.md` -> Decision 1–6；`## Review fold — PR #50`

**Boundary**:
- 產品碼：`internal/agent/agentloop.go`（`logStep` 改用 `ui.FormatToolLog` + `Now` seam；移除 `arguments=`／`result=`；**`logStep` 必須持續包裹 `a.Observer.BeforeToolLog()`／`AfterToolLog()`** 以維持 round-019 spinner — PR #50 review 2 directive 1）；`internal/ui/toollog.go`（實作 `FormatToolLog`；R2 `formatClock` 含 `FormatMetrics`）；`internal/cli/cli.go`（`loop.Now = env.now`；`len(result.Steps) > 0` 時輸出 `fmt.Fprintln(env.stderr)` 的**恰好一個 `\n`**，位置**在 `sp.Stop()` 與 `store.Append` 之後、緊接 `env.writeAnswer(...)` 之前** — review 2 directive 2，以免被 spinner 清除吞掉）。
- `AgentLoop.now()` fallback：`if a.Now != nil { return a.Now() } return time.Now()`（review 2 directive 3）。
- 空行**不受 `chrome` 條件限制**（跟隨 tool-log 行；任何 tool-using 表面，含 `-i` submit 與非 chrome 路徑）——由 `-i` tool-using Example 見證（TD1）。
- `stdout` 契約不變；class-phrase 詞彙不變（11）；不動 payload line／spinner；`go mod` 不變。
- review 回應指出：本輪改變**每個** tool-using turn 的 `stderr`，故 T016 回歸須重檢既有 round-010/017 ordering Thens。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [X] T014 [BDD-GREEN] 讓 Test Scope 全綠（並使 T011／T012 的 `[UNIT]` 轉綠）
- [X] T015 [BDD-REFACTOR] 在綠燈下整理 `logStep`／`FormatToolLog`／blank-line emit 落點

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞，並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T016 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **重檢既有 ordering Thens**（round-010/017）：`the tool activity is reported before the answer`、`the turn frame is separated from the answer`（因本輪改了每個 tool-using turn 的 `stderr`）。
  - **可偽性見證**（各觀察失敗後還原）：
    1. 還原 `logStep` 為舊的 `[tool] … arguments=… result=…` 行 → `watching-the-tool-loop.feature` 的 reason Then 失敗。
    2. 移除 `len(result.Steps) > 0` 時於答案前的空行 → `the tool report is separated from the answer` 失敗（**含 `-i` 那條**，即 ungated 見證）。
    3. 讓空行**多印一次**（`\n\n`）於 tool-using run → `the tool report is separated from the answer`（恰一空行）失敗。
    4. 讓 used-no-tool 的 `-i` run 也印空行 → `the tool loop added no blank line before the answer` 失敗。
    5. 把空行 gate 在 `chrome` → `-i` 的 `the tool report is separated from the answer` 失敗（TD1 見證）。
    6. 讓無 reason 的呼叫仍印出 ` - ` 尾 → `the run reported the tool call "read_files" without a reason` 失敗。
    7. 讓同回合的兩個呼叫只記一行 → `the run reported one tool-loop log line for each tool call` 失敗。
    8. 反轉 `get_tree`／`read_files` 的 log 次序 → `the run reported the tool calls in order "get_tree" and "read_files"` 失敗（TD3 見證）。
  - 確認：`stdout` byte-exact；exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；offline paths 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`：Agent tool loop + Pure-helper unit tests） | T001、T011、T012、T014 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務；T016 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務；T016 驗證 record 形狀不變） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`） | T003–T010、T014、T015 | PASS |
| `research.md` -> Decision 1（單行 shape） | T001、T003–T005、T011、T012、T014、T015 | PASS |
| `research.md` -> Decision 2（injected clock seam / `internal/ui` formatter / `AgentLoop.Now`） | T001、T011、T012、T014 | PASS |
| `research.md` -> Decision 3（drop `arguments=`／`result=`） | T003、T004、T012、T014 | PASS |
| `research.md` -> Decision 4（no-reason ⇒ `[Tool] <name>`；generic extraction kept） | T005、T012、T014 | PASS |
| `research.md` -> Decision 5（CLI 於 tool-using turn 答案前輸出一個空行） | T007、T009、T014、T015、T016 | PASS |
| `research.md` -> Decision 6（`stderr`；`stdout` byte-exact） | T007、T014、T016 | PASS |
| `research.md` -> Decision 7（testing strategy；no pty） | T002、T011、T012、T016 | PASS |
| `research.md` -> Decision 8（no new dependency；no schema/loop-contract change） | T014、T016（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Review fold **B1**（negative carrier isolated on `-i`） | T009、T014、T016 | PASS |
| `research.md` -> Review fold **TD1**（ungated witnessed on `-i`） | T007（`-i` Example）、T014、T016 | PASS |
| `research.md` -> Review fold **TD2**（no-reason fixture documented） | T005、T014 | PASS |
| `research.md` -> Review fold **TD3**（ordered Then + sequential two-tool Given） | T008、T010、T014 | PASS |
| `research.md` -> Review fold **R1**（acceptance pointer） | T014（feature header） | PASS |
| `research.md` -> Review fold **R2**（`formatClock` factor） | T001、T011、T015 | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-012`、`SC-001`–`SC-005` | T003–T012（對齊）、T014–T015（交付）、T016 | PASS |
| `plan.md` -> Source-code structure（`internal/agent`、`internal/ui`、`internal/cli`） | T001、T014、T015 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；CLI end → /axb-dsl-refine；ui skipped） | T001、T016 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
