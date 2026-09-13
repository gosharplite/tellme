# Tasks: Agent Tools & the Tool-Call Loop (round 008)

**Plan Package**: `specs/plans/008-agent-tools-and-tool-call-loop`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only — `os`、`path/filepath`、`io`、`encoding/json`、`context`、`time`；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 2: Foundational

**Goal**: 建立 tool domain port、read-only filesystem tool adapter、`llm.Request`/`llm.Response` 工具欄位擴充、`history.Entry` 加寬、`MAX_TOOL_LOOP` 解析、`ToolError` code、CLI loop 落點，以及 E2E 共用元件（fake 供應 tool-call、記錄 tool definitions、stderr 擷取、working-dir 檔案 helper）與 16 個新句 stepdef 落點骨架（Zero Shared Edits 原則）。

- [ ] T001 建立 tool domain port 落點骨架 `internal/domain/tools/tools.go`
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 1
    - `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools / Agent tool loop）
  - 只做：宣告 `internal/domain/tools` 套件與 `Tool` value type（`Name`, `Description`, `Parameters`）及 `Registry` 型別（名稱→executor 解析）與 executor 介面簽名 stub；`Tool` 實作介面 MUST 帶 `context.Context`（`Execute(ctx context.Context, arguments string) (string, error)`），供 loop 以 `context.WithTimeout` 派生 per-tool timeout（**RF-2**）。
  - 不做：不實作任何工具；不接 `internal/cli`；不寫斷言。

- [ ] T002 建立 tool adapter 落點骨架 `internal/infrastructure/tools/filesystem.go`、`internal/infrastructure/tools/summarize.go`
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools）
    - `internal/domain/tools/tools.go`
  - 只做：宣告 `list_files` 與 `read_files` 兩個工具的 constructor 簽名與 **wire-valid snake_case** 名稱 + JSON-schema 參數；`read_files` 留 `io.LimitReader` 的 **1 MiB** 上界 stub 與截斷標記；另宣告 LLM-backed `summarize_history` 工具的 constructor 簽名（注入 `history.Store` + `llm.Gateway`；空參數 schema；回傳 summary 字串、不改記錄）；不做任何寫入 / 程序啟動 / 網路。
  - 不做：不實作目錄列舉 / 檔案讀取邏輯；不碰 `internal/cli`。

- [ ] T003 擴充 `llm.Request`/`llm.Response` 與 OpenAI adapter 落點骨架
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port / Request assembly / Response normalization）
    - `internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`
  - 只做：`llm.Request` 新增 tool **definitions** 欄位、`llm.Response` 新增結構化 **tool-call requests**（id/name/arguments）欄位之型別骨架；在 `requestBody`/`parseAnswer` 留下送出 `tools` 陣列、assistant `tool_calls` 與 `tool`-role 訊息、解析 `choices[0].message.tool_calls` 的簽名/stub。**向後相容**：空 tools 時 payload 與 round 004–007 byte-for-byte 一致。
  - 不做：不改 transport、錯誤處理或既有 messages 組裝語意。

- [ ] T004 加寬 `history.Entry` 與 file adapter 落點骨架 `internal/domain/history/history.go`、`internal/infrastructure/history/file_store.go`
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 5
    - `specs/truth/data/data-model.dbml` -> `history_entry` / `history_step`
    - `specs/truth/techstack.md` -> CLI Application（Session history store）
  - 只做：`Entry` 新增有序 `Steps`（`[]Step{Tool, Arguments, Result}`）欄位骨架；file adapter 留下把加寬欄位序列化進單一 JSON 列（固定欄位順序、無時間戳/ID）與讀回時的 stub。
  - 不做：不實作 append / load / archive 行為；不碰 `internal/cli`。

- [ ] T005 落點骨架 `internal/agent/`、`internal/cli/cli.go`、`internal/cli/exitcode.go`、`internal/config/config.go`
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 3, 6, 7
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop）、Configuration（Rendered width 之鄰）
  - 只做：`internal/cli/exitcode.go` 新增 `ToolError = 7` 常數；`internal/config` 留下 `MAX_TOOL_LOOP` 解析 stub（env/config，預設 1000）；建立 `internal/agent/` 落點骨架並宣告 `AgentLoop`（注入 `llm.Gateway` + `tools.Registry` + `history.Store`，`MaxLoops int`，`Stderr io.Writer`）之 `Run(ctx, prompt, prior)` 簽名（**RF-1**：loop 不在 `cli.go` 內）；`cli.go` 只留「wiring `AgentLoop`」與 `stderr` 工具迴圈 log 的 seam stub；留「loop 無法完成 → `the tool request failed` + code 7」hook 簽名。
  - 不做：不實作 loop 行為、工具派送或 resume/replay；不改既有 dispatch 與 exit-code（產品行為留 Phase 4）。

- [ ] T006 建立 E2E 共用元件落點骨架
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 8, 9
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner / Local fake provider / Pure-helper unit tests）
    - `tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 只做：擴充 fake provider 以**供應** scripted tool-call 回應並**記錄**送出的 `tools` 定義；新增擷取 subprocess `stderr` 的 harness helper；新增在 subprocess working directory 寫入檔案的 helper；新增設定 `MAX_TOOL_LOOP` 的環境注入。
  - 不做：不寫具體 step 斷言；不碰 `internal/`。

- [ ] T007 建立 16 個新句 stepdef 獨立檔案骨架 `tests/e2e/steps/step_t008_*.go`–`step_t023_*.go`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（14 個新句）
    - `specs/truth/features/cli/history/dsl.md`（2 個新句）
    - `tests/e2e/steps/register.go`
  - 只做：建立 16 個獨立檔案（各對應 T008–T023 的句），各自 `init()` 自我註冊空白 registrar。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的句與本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：介面根 `specs/truth/features/cli/dsl.md`（跨模組句），或同模組 `specs/truth/features/cli/{chat,history}/dsl.md`。不得掃其他模組。
- 本輪 16 個新句：14 個在 `chat/dsl.md`、2 個在 `history/dsl.md`。介面根 `dsl.md` 僅有 **MODIFY**（frozen class-phrase 詞彙 10→11，新增 `the tool request failed`）；該句 `tellme explains on stderr that "{reason}"` 的**句型與斷言語意未變**（新增的片語是參數值），不需 ALIGN。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**無**。既有句的斷言語意不變（`the request carried the earlier exchange …`、`tellme stored the exchange …`、`tellme sends exactly one request …` 均未被改判；`-l` 既有句語意不變）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 16 個新句（`ADD`）。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decisions 1, 4, 6, 8）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 14 個 `chat` 新句
- `specs/truth/features/cli/history/dsl.md` -> 2 個 `history` 新句
- `truth-delta.md` -> `/axb-dsl-refine` ADD（4 chat features + module rows）、MODIFY（root `cli/dsl.md` 詞彙）、MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 `internal/infrastructure/tools/*_test.go`、`internal/domain/tools/*_test.go`、`internal/infrastructure/history/*_test.go`、`internal/config/*_test.go`、`internal/cli/*_test.go`，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T008–T023 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T024（UNIT）可並行；T025 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T008 [P] [BDD-RED] `Given: the working directory contains a file "{name}" whose text is "{content}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the working directory contains a file "{name}" whose text is "{content}"`
  - Landing: `tests/e2e/steps/step_t008_chat_given_workdir_file.go`
  - 語意：在 subprocess 工作目錄建立名為 `{name}`、內容為 `{content}` 的檔案。

- [ ] T009 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t009_chat_given_provider_read_then_answer.go`
  - 語意：寫可解析 config 選 `{provider}`；fake 先回一個 `read_files` 的 tool-call 回應（arguments 帶 `{path}`），下一個請求回 `{answer}`。

- [ ] T010 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint always asks tellme to read "{path}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint always asks tellme to read "{path}"`
  - Landing: `tests/e2e/steps/step_t010_chat_given_provider_always_read.go`
  - 語意：fake 每次請求都回同一個 `read_files` tool-call 回應（永不給最終答案），使 loop 撞到上界。

- [ ] T011 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks for a tool that is not available`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint asks for a tool that is not available`
  - Landing: `tests/e2e/steps/step_t011_chat_given_provider_unknown_tool.go`
  - 語意：fake 回一個指名未宣告工具的 tool-call 回應。

- [ ] T012 [P] [BDD-RED] `Given: the tool-loop limit is "{limit}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the tool-loop limit is "{limit}"`
  - Landing: `tests/e2e/steps/step_t012_chat_given_tool_loop_limit.go`
  - 語意：把 `MAX_TOOL_LOOP` 設為 `{limit}` 進 subprocess 環境。

- [ ] T013 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to summarise the conversation and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint asks tellme to summarise the conversation and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t013_chat_given_provider_summarise_then_answer.go`
  - 語意：fake 先回一個 summarise tool-call 回應，下一個請求回 `{answer}`。

- [ ] T014 [P] [BDD-RED] `Then: tellme read "{path}" using its read_files tool`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme read "{path}" using its read_files tool`
  - Landing: `tests/e2e/steps/step_t014_chat_then_read_tool.go`
  - 語意：斷言 fake 記錄到模型對 `{path}` 的 `read_files` tool-call，且該工具結果被餵回對話。

- [ ] T015 [P] [BDD-RED] `Then: tellme used no tool`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme used no tool`
  - Landing: `tests/e2e/steps/step_t015_chat_then_used_no_tool.go`
  - 語意：斷言該 run 未呼叫任何工具（request 未帶 tool result；fake 未記錄任何 tool round）。

- [ ] T016 [P] [BDD-RED] `Then: the run reported the tool call "{tool}" on its diagnostic output`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reported the tool call "{tool}" on its diagnostic output`
  - Landing: `tests/e2e/steps/step_t016_chat_then_reported_tool_call.go`
  - 語意：斷言擷取的 **stderr** 帶有指名 `{tool}` 的工具迴圈 log 行；且該 log **不在** stdout。

- [ ] T017 [P] [BDD-RED] `Then: the request carried the read-tool error for "{path}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried the read-tool error for "{path}"`
  - Landing: `tests/e2e/steps/step_t017_chat_then_read_tool_error.go`
  - 語意：斷言 fake 記錄到一個請求，帶有 `read_files` 對 `{path}` 的**錯誤**結果且被當成非終結結果餵回。

- [ ] T018 [P] [BDD-RED] `Then: tellme stopped after {count} tool iterations`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme stopped after {count} tool iterations`
  - Landing: `tests/e2e/steps/step_t018_chat_then_stopped_after.go`
  - 語意：斷言 fake 恰好記錄 `{count}` 個 tool-iteration round。

- [ ] T019 [P] [BDD-RED] `Then: tellme exits with the tool error code`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme exits with the tool error code`
  - Landing: `tests/e2e/steps/step_t019_chat_then_tool_error_code.go`
  - 語意：斷言 exit code 等於 `7`（distinct from 0/2/3/4/5/6）。

- [ ] T020 [P] [BDD-RED] `Then: tellme summarised the earlier conversation using its summarise tool`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme summarised the earlier conversation using its summarise tool`
  - Landing: `tests/e2e/steps/step_t020_chat_then_summarised.go`
  - 語意：斷言 fake 記錄到模型的 summarise tool-call，且產出的摘要被餵回對話。

- [ ] T021 [P] [BDD-RED] `Then: the earlier conversation records are unchanged`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the earlier conversation records are unchanged`
  - Landing: `tests/e2e/steps/step_t021_chat_then_records_unchanged.go`
  - 語意：斷言 history 檔先前已存的列 byte 不變。

- [ ] T022 [P] [BDD-RED] `Given: the session history already holds a tool-using exchange`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `the session history already holds a tool-using exchange`
  - Landing: `tests/e2e/steps/step_t022_history_given_tool_using_exchange.go`
  - 語意：建立 per-mode workspace 目錄，append 一行加寬的 JSON（`prompt`、`answer`、`steps` 含一步）進 `history.jsonl`。

- [ ] T023 [P] [BDD-RED] `Then: tellme lists only the operator's messages`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `tellme lists only the operator's messages`
  - Landing: `tests/e2e/steps/step_t023_history_then_operator_only.go`
  - 語意：斷言 stdout 帶有存的 prompt 與 answer，且**不含** tool-step 文字。

### UNIT（pure-helper 單元測試）

- [ ] T024 [P] [UNIT] tool registry / tools / loop / 加寬 history 單元測試
  - Read:
    - `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 1, 4, 6, 8
    - `specs/truth/data/data-model.dbml` -> `history_entry` / `history_step`
    - `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools / Agent tool loop）、Testing & Verification（Pure-helper unit tests）
  - Landing: `internal/domain/tools/*_test.go`、`internal/infrastructure/tools/*_test.go`、`internal/infrastructure/history/file_store_test.go`、`internal/config/*_test.go`、`internal/cli/*_test.go`（表驅動）
  - 撰寫：tool registry 派送；`list_files`/`read_files`（含 1 MiB 讀取上界與 missing-path 錯誤）；loop bound（`MAX_TOOL_LOOP` 解析 + 預設 1000）與失敗契約（`the tool request failed` / code 7）；`history.Entry` 加寬的 append + reload（固定欄位順序、無時間戳/ID）；`-l` 僅讀 prompt/answer。

### Phase Review Gate

- [ ] T025 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/using-a-tool.feature`、`watching-the-tool-loop.feature`、`failing-the-tool-loop.feature`、`summarising-the-conversation.feature`、`specs/truth/features/cli/history/inspecting-the-session-history.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/history/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 檢驗：無 undefined step；16 個新 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/using-a-tool.feature

**Goal**: 實作「prompt run 可使用已宣告的 read-only 工具並把結果併入答案」與「不需工具的 prompt 不呼叫工具」，使 `using-a-tool.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/using-a-tool.feature` -> `Feature: Using a declared tool`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`tellme read …`、`tellme used no tool`）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/using-a-tool.feature` + module rows）
- `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools / Agent tool loop）、Reasoning & Provider Transport（Provider gateway port / Request assembly / Response normalization）
- `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 1, 2, 4
- `internal/domain/tools/`、`internal/infrastructure/tools/`、`internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`、`internal/cli/cli.go`

**Boundary**:
- 提供 `list_files` 與 `read_files`（read-only、1 MiB 讀取上界、無 path/safety boundary）；送出 `tools` 定義；解析模型回傳的 tool-call 並執行；把工具結果以 `tool`/assistant tool-call 訊息餵回 provider；loop 有界（`MAX_TOOL_LOOP`）。
- 不需工具的 prompt 維持**單一** provider 請求，且 payload 與 round 004–007 byte-for-byte 一致。
- 不實作 summarise / 失敗契約細節（屬 4C/4D boundary 內的其他 phase）。

**Test Scope**:
- `specs/truth/features/cli/chat/using-a-tool.feature`

- [ ] T026 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T027 [BDD-REFACTOR] 在綠燈下整理 tool registry 派送與 request/response 工具欄位

## Phase 4B: ADD Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: 實作工具迴圈的即時可見性（工具迴圈 log 行寫到 `stderr`），使 `watching-the-tool-loop.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` -> `Feature: Watching the tool loop work`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`the run reported the tool call …`）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/watching-the-tool-loop.feature`）
- `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 7

**Boundary**:
- 每個工具迴圈步驟寫一條**離散** log 行到 **standard error**（工具名、arguments、結果/錯誤），`loop` 進行時即時輸出；**stdout** 保持為答案流（rendered / raw 不變）；非 token streaming。
- 不把 log 寫到 stdout；不改最終答案輸出契約。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [ ] T028 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T029 [BDD-REFACTOR] 在綠燈下整理工具迴圈 log 的輸出與 stream 分流

## Phase 4C: ADD Feature File - cli/chat/failing-the-tool-loop.feature

**Goal**: 實作 loop 的有界終結與失敗契約（可回復工具錯誤餵回、無法完成的 loop 回 `the tool request failed` + code 7），使 `failing-the-tool-loop.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/failing-the-tool-loop.feature` -> `Feature: Bounding and failing the tool loop`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`the request carried the read-tool error …`、`tellme stopped after …`、`tellme exits with the tool error code`）
- `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`（詞彙 10→11）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/failing-the-tool-loop.feature`）+ MODIFY（root `cli/dsl.md` 詞彙）
- `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 3, 6

**Boundary**:
- 工具**回錯誤** → 當作工具結果餵回模型（非終結）；loop 以 `MAX_TOOL_LOOP`（env/config，預設 1000）為上界；未宣告工具 / 上界到達等**無法完成**情況 → 新 frozen class phrase `tellme: the tool request failed` + exit code **7**（vocabulary 10→11）。
- 不新增第二條 `tellme: ` 行；不得 hang。

**Test Scope**:
- `specs/truth/features/cli/chat/failing-the-tool-loop.feature`

- [ ] T030 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T031 [BDD-REFACTOR] 在綠燈下整理 loop 上界、timeout 與失敗分類

## Phase 4D: ADD Feature File - cli/chat/summarising-the-conversation.feature

**Goal**: 實作 history summarisation 作為一個 **on-demand agent tool**（LLM-backed），且不重寫既有 history 記錄，使 `summarising-the-conversation.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/summarising-the-conversation.feature` -> `Feature: Summarising the conversation with an agent tool`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`tellme summarised the earlier conversation using its summarise tool`、`the earlier conversation records are unchanged`）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/summarising-the-conversation.feature`）
- `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 1, 3, 5
- `specs/truth/techstack.md` -> Not Introduced Yet（History summarisation introduced as an agent tool；pruning excluded）

**Boundary**:
- summarise 工具為 LLM-backed、受相同 loop 上界與失敗契約約束；受測對象是「模型可透過工具摘要既有對話」與「既有記錄不被改寫」。
- **不得**引入 token-budget pruning（settled exclusion）；summary 不得覆寫既有 history 列。

**Test Scope**:
- `specs/truth/features/cli/chat/summarising-the-conversation.feature`

- [ ] T032 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T033 [BDD-REFACTOR] 在綠燈下整理 summarise 工具的接線與記錄不變保證

## Phase 4E: MODIFY Feature File - cli/history/inspecting-the-session-history.feature

**Goal**: 在加寬記錄下維持 `-l` 只列 operator 訊息，使 `inspecting-the-session-history.feature`（含新 Rule）全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature` -> `Feature: Inspecting the session history`
- `specs/truth/features/cli/history/dsl.md` -> Given（`the session history already holds a tool-using exchange`）、Then（`tellme lists only the operator's messages`）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`）
- `specs/truth/data/data-model.dbml` -> `history_entry` / `history_step`
- `specs/plans/008-agent-tools-and-tool-call-loop/research.md` -> Decision 5

**Boundary**:
- `-l N` 讀加寬列時，只輸出 `prompt`/`answer`（`role: content` 一行一則），**不**輸出 tool steps；`-l` 契約與 round 007 一致（正整數、clamp、離線短路）。
- 不改 `internal/cli` 的 `-l` 既有語意；不把 tool activity 帶進 listing。

**Test Scope**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature`

- [ ] T034 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T035 [BDD-REFACTOR] 在綠燈下整理加寬列的讀取投影（operator-only）

## Phase 4F: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（含 prompt-turn loop 與既有 CLI 契約，含 `-l`/`--new` 離線路徑）。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T036 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6/7` 與既有 CLI 契約（round 001–007）維持綠燈；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不觸網；無工具 prompt 的既有 chat 場景 payload 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；class-phrase 詞彙為 11、`-l` 仍只列 prompt/answer。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/data/data-model.dbml` -> `history_entry`（加寬，嵌 steps） | T004、T020–T022、T024、T026、T034 | PASS |
| `specs/truth/data/data-model.dbml` -> `history_step`（location, position, step） | T004、T022、T024、T034 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools：`list_files`/`read_files`、1 MiB cap、no boundary；+ LLM-backed `summarize_history`） | T002、T013、T014、T020、T024、T026、T032 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Agent tool loop：`MAX_TOOL_LOOP` 1000、per-tool timeout、logs to stderr） | T005、T012、T016、T018、T024、T026、T028、T030、T031 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port / Request assembly / Response normalization 工具欄位） | T003、T009–T013、T026、T027 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session history store：加寬列 + resume replay；`-l` prompt/answer only） | T004、T021–T023、T034、T035 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（fake 供應/記錄 tool calls；pure-helper 擴充） | T006、T009–T010、T014、T024、T025 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（summarisation introduced as an agent tool；pruning/write-tools/concurrency excluded） | 負向決策 -> T032/T033 Boundary、T036（不引入 pruning / write tools / concurrency） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001–T006、T026–T036 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` MODIFY（`specs/truth/data/data-model.dbml`） | T004、T024 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/using-a-tool.feature` + module rows） | T007–T009、T014、T015、T025、T026、T027 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/watching-the-tool-loop.feature` + row） | T007、T016、T025、T028、T029 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/failing-the-tool-loop.feature` + rows） | T007、T010–T012、T017–T019、T025、T030、T031 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/summarising-the-conversation.feature` + rows） | T007、T013、T020、T021、T025、T032、T033 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（root `cli/dsl.md` 詞彙 10→11） | T019（`the tool request failed` 的 exit-7）；句型語意未變故不另建 ALIGN | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`） | T007、T022、T023、T025、T034、T035 | PASS |
| `research.md` -> Decision 1（tool domain port + registry） | T001、T024、T026、T027 | PASS |
| `research.md` -> Decision 2（provider port 工具欄位擴充） | T003、T026、T027 | PASS |
| `research.md` -> Decision 3（有界 loop、`MAX_TOOL_LOOP` 1000） | T005、T012、T018、T030、T031 | PASS |
| `research.md` -> Decision 4（兩個 read-only 工具 + 1 MiB 讀取上界） | T002、T014、T024、T026 | PASS |
| `research.md` -> Decision 5（加寬 `history_entry` + resume replay） | T004、T021、T022、T034 | PASS |
| `research.md` -> Decision 6（失敗契約 `the tool request failed` + exit 7） | T005、T019、T030、T031 | PASS |
| `research.md` -> Decision 7（工具迴圈 log → stderr，非 streaming） | T005、T016、T028、T029 | PASS |
| `research.md` -> Decision 8（fake 供應/記錄 tool calls；offline 集合不變） | T006、T009–T010、T014、T036 | PASS |
| `research.md` -> Decision 9（無新相依） | T036（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 10（`summarize_history`：名稱、空 schema、注入 `Store`+`Gateway`、non-mutating） | T002、T013、T020、T032、T033 | PASS |
| `research.md` -> Decision 4/5 修訂（wire-valid snake_case 名稱；`arguments` 欄位；`tool_call_id` 合成） | T002、T004、T014、T019、T024、T036 | PASS |
| `spec.md` -> US1–US4、`FR-001`–`FR-018`、`NFR-001`–`NFR-007` | T008–T023（對齊）、T026–T035（交付）、T036 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
