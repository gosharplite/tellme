# Tasks: Agent Tool Surface Parity (round 021)

**Plan Package**: `specs/plans/021-tool-surface-parity`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

> **PR #48 review fold** (B1/B2/TD1/TD2/TD3/R1/R3 + nits): this revision adds four `[BDD-RED]` tasks (T028–T031 — the tree negatives, the directory branch, the read framing), extends the `[UNIT]` set (T032 — aggregate/list/tree caps + directory), and fixes the T047 falsifiability witness #3.

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions（1–8）+ D3a（B2 bounded results）與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod`／`go.sum` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（含 `[BDD-REMOVE]` 與 `[BDD-ALIGN]`）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/infrastructure/tools`（重塑 `list_files`／多檔 `read_files`、新增 `get_tree`、1 MiB aggregate cap）、`internal/infrastructure/tools/summarize.go`（刪除）、`internal/cli/cli.go`（registry factory 只提供三個工具）、`internal/agent/agentloop.go`（`reason` echo）。維持「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。

---

## Phase 2: Foundational

**Goal**: 建立本輪 stepdef／`[UNIT]` 與產品碼落點骨架（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔案或 seam。只建立落點與載體，不寫工具行為。

- [ ] T001 建立 25 個新句 stepdef 獨立檔骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 25 個新句）
    - `tests/e2e/steps/register.go`、`tests/e2e/steps/scenario_context.go`
  - 只做：建立 25 個獨立 stepdef 檔（命名 `tests/e2e/steps/step_r021_t0NN_<slug>.go`，各自 `init()` 自我註冊空白 registrar），對到 T007–T031。
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [ ] T002 建立 `[UNIT]` 落點與產品碼骨架 seam
  - Read:
    - `specs/plans/021-tool-surface-parity/research.md` -> Decision 3, 5, 7
    - `internal/infrastructure/tools/filesystem.go`、`internal/infrastructure/tools/filesystem_test.go`、`internal/agent/agentloop.go`
  - 只做：新增 `internal/infrastructure/tools/get_tree.go`（`get_tree` 工具型別骨架，`Name/Description/Parameters/Execute` 空殼）與 `internal/infrastructure/tools/binary.go`（stdlib binary-probe helper 骨架）；建立 `[UNIT]` 落點檔（`internal/infrastructure/tools/get_tree_test.go`、`internal/agent/agentloop_reason_test.go`）為空殼。
  - 不做：不實作工具行為；不改 registry factory；不寫 `reason` echo；不寫斷言。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth：先移除已退場工具的步驟、對齊既有 `read_files` 句到多檔 `filepaths`，再為 25 個新句建立失敗訊號；最後補非 DSL 的 `[UNIT]`。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪無跨模組新句；frozen class-phrase 詞彙維持 **11**）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-REMOVE]`：本輪 **3 句**（`DELETE` 的 DSL rows — summarise Given + 兩個 summarise Then）。
- `[BDD-ALIGN]`：既有 `read_files` 句的 stepdef 需對齊多檔 `filepaths`（fake scripting + read-result assertion）。
- `[BDD-RED]`：本輪 **25 句**（`ADD` 的 DSL rows — 21 個功能句 + B1/TD2/R1 的 4 個）。
- `[UNIT]`：三個工具的 schema／輸出／上限（含 1 MiB aggregate）、`get_tree` 格式、`reason` echo、registry 工具集合（非 DSL）。
- 除 T002 的型別骨架外，兩者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 25 個新句（見 T007–T031）+ 既有 `read_files` 句（`a configured provider "…" whose endpoint asks tellme to read "…" and then answers with "…"`、`… always asks tellme to read "…"`、`a configured Gemini provider "…" whose endpoint asks tellme to read "…" …`、`tellme read "…" using its read_files tool`、`the request carried the read-tool error for "…"`）
- `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`）+ `/axb-dsl-refine` ADD / MODIFY / DELETE + NOOP（root `cli/dsl.md`、`contracts/**`、`data/**`）
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/fakeprovider.go`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點採 stdlib `testing`；`internal/infrastructure/tools/*_test.go`、`internal/agent/*_test.go`。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T034 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；`[BDD-REMOVE]`／`[BDD-ALIGN]` 觸及的既有檔若同檔則序列化）；T035 等全部回來再啟動 subagent 執行 review。

### BDD-REMOVE（退場工具的步驟）

- [ ] T003 [BDD-REMOVE] 移除 Given `a configured provider "{provider}" whose endpoint asks tellme to summarise the conversation and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md`（該句已刪除）；`tests/e2e/steps/step_t013_chat_given_provider_summarise_then_answer.go`
  - 動作：刪除該 stepdef 檔。
  - 不做：不動讀檔／清單／樹狀工具步驟。

- [ ] T004 [BDD-REMOVE] 移除 Then `tellme summarised the earlier conversation using its summarise tool` 與 Then `the earlier conversation records are unchanged`
  - Read: `tests/e2e/steps/step_t020_chat_then_summarised.go`、`tests/e2e/steps/step_t021_chat_then_records_unchanged.go`
  - 動作：刪除這兩個 stepdef 檔。
  - 不做：不動其他 Then。

### BDD-ALIGN（既有 `read_files` 句對齊多檔 `filepaths`）

- [ ] T005 [BDD-ALIGN] 既有 fake-scriing Givens 改送 `filepaths: ["…"]`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "…" whose endpoint asks tellme to read "…" and then answers with "…"`、`… always asks tellme to read "…"`、`a configured Gemini provider "…" whose endpoint asks tellme to read "…" and then answers with "…"`
  - Landing: `step_t009_chat_given_provider_read_then_answer.go`、`step_t010_chat_given_provider_always_read.go`、`step_t007_config_given_gemini_read_then_answer.go`、`step_t004_chat_given_two_tool_provider.go`
  - 動作：讓這些 Given 對 fake 腳本化的 `read_files` 呼叫改用新參數 `{"filepaths":["…"]}`（單檔仍為長度 1 的陣列）。

- [ ] T006 [BDD-ALIGN] 既有 read-result Thens 對齊 `filepaths`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme read "{path}" using its read_files tool`、`the request carried the read-tool error for "{path}"`
  - Landing: `step_t014_chat_then_read_tool.go`、`step_t017_chat_then_read_tool_error.go`
  - 動作：讓斷言以 `filepaths` 內的 `{path}` 判定 `read_files` 呼叫與其結果回饋。

### BDD-RED（本輪新增句型 — Given）

- [ ] T007 [P] [BDD-RED] `Given: the working directory contains a sub-folder "{name}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t007_chat_given_workdir_subfolder.go`
  - 語意：在工作目錄建立空的子資料夾 `{name}`（支援巢狀如 `src/pkg`、`.git`）。

- [ ] T008 [P] [BDD-RED] `Given: the working directory contains a file "{name}" whose text is longer than the read limit`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t008_chat_given_workdir_large_file.go`
  - 語意：建立 `{name}`，內容超過 100000 bytes。

- [ ] T009 [P] [BDD-RED] `Given: the working directory contains a binary file "{name}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t009_chat_given_workdir_binary_file.go`
  - 語意：建立含 NUL byte 的 `{name}`（非 UTF-8 文字）。

- [ ] T010 [P] [BDD-RED] `Given: the working directory contains more files than tellme reads in one request`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t010_chat_given_workdir_too_many_files.go`
  - 語意：建立 >50 個檔案於工作目錄。

- [ ] T011 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path_a}" and "{path_b}" in one request and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t011_chat_given_provider_read_two_one_request.go`
  - 語意：fake 一次回應帶一個 `read_files` 呼叫，`filepaths` = [`{path_a}`, `{path_b}`]，次回覆答案 `{answer}`。

- [ ] T012 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read more files than one request allows and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t012_chat_given_provider_read_too_many.go`
  - 語意：fake 回應一個 `read_files` 呼叫，`filepaths` 列出工作目錄全部（>50）檔案。

- [ ] T013 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" with the reason "{reason}" and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t013_chat_given_provider_read_with_reason.go`
  - 語意：fake 回應 `read_files` 呼叫，args 帶 `filepaths` = [`{path}`] 與 `reason` = `{reason}`。

- [ ] T014 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint lists the current directory and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t014_chat_given_provider_lists_dir.go`
  - 語意：fake 回應 `list_files` 呼叫（無 path），次回覆答案 `{answer}`。

- [ ] T015 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint shows the folder tree and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t015_chat_given_provider_shows_tree.go`
  - 語意：fake 回應 `get_tree` 呼叫，次回覆答案 `{answer}`。

- [ ] T016 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint reports the offered tools and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Given
  - Landing: `tests/e2e/steps/step_r021_t016_chat_given_provider_reports_tools.go`
  - 語意：fake 記錄請求帶的工具定義並回覆答案 `{answer}`。

### BDD-RED（本輪新增句型 — Then）

- [ ] T017 [P] [BDD-RED] `Then: the run made a single read_files request carrying "{path_a}" and "{path_b}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r021_t017_chat_then_single_read_two.go`
  - 語意：fake 恰記錄一次請求，其 `read_files` 呼叫 `filepaths` = [`{path_a}`, `{path_b}`]。

- [ ] T018 [P] [BDD-RED] `Then: the part of "{name}" that tellme read ends with a truncation marker`
  - Landing: `tests/e2e/steps/step_r021_t018_chat_then_truncated.go`
  - 語意：`read_files` 對 `{name}` 的結果以 `... (truncated)` 結尾。

- [ ] T019 [P] [BDD-RED] `Then: tellme reports that "{name}" is a binary file that cannot be shown as text`
  - Landing: `tests/e2e/steps/step_r021_t019_chat_then_binary.go`
  - 語意：`read_files` 結果帶 `(Binary file, cannot display as text)`。

- [ ] T020 [P] [BDD-RED] `Then: tellme reports that too many files were requested`
  - Landing: `tests/e2e/steps/step_r021_t020_chat_then_too_many.go`
  - 語意：`read_files` 結果帶 too-many-files 訊息。

- [ ] T021 [P] [BDD-RED] `Then: tellme listed the directory using its list_files tool`
  - Landing: `tests/e2e/steps/step_r021_t021_chat_then_listed.go`
  - 語意：fake 記錄 `list_files` 呼叫且結果回饋。

- [ ] T022 [P] [BDD-RED] `Then: the listing shows "{name}" as a file`
  - Landing: `tests/e2e/steps/step_r021_t022_chat_then_listing_file.go`
  - 語意：`list_files` 結果帶 `[f] {name}` 行。

- [ ] T023 [P] [BDD-RED] `Then: the listing shows "{name}" as a folder`
  - Landing: `tests/e2e/steps/step_r021_t023_chat_then_listing_folder.go`
  - 語意：`list_files` 結果帶 `[d] {name}` 行。

- [ ] T024 [P] [BDD-RED] `Then: tellme showed the folder tree using its get_tree tool`
  - Landing: `tests/e2e/steps/step_r021_t024_chat_then_tree.go`
  - 語意：fake 記錄 `get_tree` 呼叫且結果回饋。

- [ ] T025 [P] [BDD-RED] `Then: the tree shows "{entry}"`
  - Landing: `tests/e2e/steps/step_r021_t025_chat_then_tree_entry.go`
  - 語意：`get_tree` 結果帶一條 connector 行，其 entry 名為 `{entry}`。

- [ ] T026 [P] [BDD-RED] `Then: the request offered exactly the reader tools`
  - Landing: `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go`
  - 語意：記錄的請求恰提供 `list_files`／`read_files`／`get_tree` 三者，無其他。

- [ ] T027 [P] [BDD-RED] `Then: the run reported the reason "{reason}" for the tool call "{tool}"`
  - Landing: `tests/e2e/steps/step_r021_t027_chat_then_reason_echo.go`
  - 語意：`stderr` 的 tool-loop log 行命名 `{tool}` 且帶 `reason={reason}`。

- [ ] T028 [P] [BDD-RED] `Then: the tree does not show "{entry}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then (B1)
  - Landing: `tests/e2e/steps/step_r021_t028_chat_then_tree_absent.go`
  - 語意：`get_tree` 結果**沒有**任何 entry 名為 `{entry}` 的 connector 行（超深度的 entry 不得出現）。

- [ ] T029 [P] [BDD-RED] `Then: the tree does not descend into "{name}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then (B1)
  - Landing: `tests/e2e/steps/step_r021_t029_chat_then_tree_no_descend.go`
  - 語意：`get_tree` 結果列出 `{name}`，但**沒有**其任何子項的 connector 行（`.git` 列出但不遞迴）。

- [ ] T030 [P] [BDD-RED] `Then: tellme reports that "{path}" is a directory`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then (TD2)
  - Landing: `tests/e2e/steps/step_r021_t030_chat_then_reports_directory.go`
  - 語意：`read_files` 對 `{path}` 的結果帶 `ERROR: path is a directory, use list_files instead`。

- [ ] T031 [P] [BDD-RED] `Then: the read result frames "{name}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then (R1)
  - Landing: `tests/e2e/steps/step_r021_t031_chat_then_read_framed.go`
  - 語意：`read_files` 結果帶一行 `--- File: {name} ---`。

### UNIT（非 DSL 的單元斷言）

- [ ] T032 [P] [UNIT] 三個工具的 schema、輸出格式與結果上限
  - Read:
    - `specs/plans/021-tool-surface-parity/research.md` -> Decision 2, 3, 5（含 D3a aggregate cap）
    - `specs/truth/techstack.md` -> CLI Application（Read-only filesystem tools）
    - `internal/infrastructure/tools/filesystem.go`、`internal/infrastructure/tools/filesystem_test.go`
  - 撰寫：`list_files` 輸出 `Contents of <path>:` + `[d]/[f]`（含預設 path）；`read_files` 多檔 framing（`--- File: … ---`）／request 順序／100000-byte 截斷／binary／directory `ERROR:`／≤50 上限／空 args 錯誤／**整份結果 1 MiB aggregate 上限**；`get_tree` connector 輸出／預設 `max_depth` 2／不遞迴 `.git`／**1 MiB 上限**；`list_files` **1 MiB 上限**；三者 schema 皆含 required `reason`（**schema-only**，工具不額外驗證）。
  - 落點：`internal/infrastructure/tools/filesystem_test.go`、`internal/infrastructure/tools/get_tree_test.go`。

- [ ] T033 [P] [UNIT] `reason` 被 echo 進 tool-loop log 行
  - Read: `specs/plans/021-tool-surface-parity/research.md` -> Decision 4；`internal/agent/agentloop.go`
  - 撰寫：`AgentLoop` 的 `logStep` 從工具 args JSON 取頂層 `reason` 並使 log 行帶 `reason=<value>`；無 `reason` 時省略。
  - 落點：`internal/agent/agentloop_reason_test.go`。

- [ ] T034 [P] [UNIT] registry factory 只提供三個檔案工具
  - Read: `specs/plans/021-tool-surface-parity/research.md` -> Decision 6；`internal/cli/cli.go`
  - 撰寫：`newToolRegistry` 回應的工具名集合恰為 `{list_files, read_files, get_tree}`（無 `summarize_history`）。
  - 落點：`internal/cli/tool_registry_test.go`。

### Phase Review Gate

- [ ] T035 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/**`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/fakeprovider/fakeprovider.go`
    - `internal/infrastructure/tools/*.go`、`internal/cli/cli.go`、`internal/agent/agentloop.go`
  - 檢驗：無 undefined step；3 個移除句已消失、25 個新句 stepdef + 3 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: MODIFY Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: 讓 `watching-the-tool-loop.feature` 全綠 — `AgentLoop` 於 `stderr` 的 `[tool]` 行附上該呼叫的 `reason`（D2）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "…" whose endpoint asks tellme to read "…" with the reason "…" and then answers with "…"`、`the run reported the reason "…" for the tool call "…"`、`the run reported the tool call "…" on its diagnostic output`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature` + `chat/dsl.md`）
- `specs/plans/021-tool-surface-parity/research.md` -> Decision 4

**Boundary**:
- 產品碼：`internal/agent/agentloop.go` 的 `logStep` 解析 args JSON 的頂層 `reason` 並加 `reason=<value>`；`stdout` 不變；class-phrase 詞彙不變（11）。
- 不動 provider transport、session history、CLI flags。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [ ] T036 [BDD-GREEN] 讓 Test Scope 全綠（並使 T033 的 `[UNIT]` 轉綠）
- [ ] T037 [BDD-REFACTOR] 在綠燈下整理 `logStep` 的 reason 擷取

## Phase 4B: ADD Feature File - cli/chat/reading-several-files.feature

**Goal**: 讓 `reading-several-files.feature` 全綠 — `read_files` 改為多檔 `filepaths`，含 `--- File: … ---` framing、截斷／binary／directory／≤50 邊界，與整份結果 1 MiB 上限。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reading-several-files.feature`
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 `read_files` 相關 Given/Then（T005/T006/T007–T013/T017–T020/T030/T031）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`reading-several-files.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Read-only filesystem tools）
- `specs/plans/021-tool-surface-parity/research.md` -> Decision 1, 5（含 D3a aggregate cap）

**Boundary**:
- 產品碼：`internal/infrastructure/tools/filesystem.go` 重塑 `read_files`（`filepaths`、framing、100000 per-file cap、**1 MiB aggregate cap**、binary/dir/≤50、inline `ERROR:`）。
- `stdout` 契約不變；class-phrase 詞彙不變；`go mod` 不變。

**Test Scope**:
- `specs/truth/features/cli/chat/reading-several-files.feature`

- [ ] T038 [BDD-GREEN] 讓 Test Scope 全綠（並使 T032 的 read 部分 `[UNIT]` 轉綠）
- [ ] T039 [BDD-REFACTOR] 在綠燈下整理 `read_files` 的 framing 與邊界處理

## Phase 4C: ADD Feature File - cli/chat/listing-a-directory.feature

**Goal**: 讓 `listing-a-directory.feature` 全綠 — `list_files` 輸出 `Contents of <path>:` + `[d]/[f]`，`path` 可省略預設 `.`，結果 1 MiB 上限。

**Shared Must Read**:
- `specs/truth/features/cli/chat/listing-a-directory.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "…" whose endpoint lists the current directory and then answers with "…"`、`tellme listed the directory using its list_files tool`、`the listing shows "…" as a file`、`the listing shows "…" as a folder`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`listing-a-directory.feature`）

**Boundary**:
- 產品碼：`internal/infrastructure/tools/filesystem.go` 重塑 `list_files`（輸出格式 + 預設 path + required `reason` + 1 MiB 結果上限）。
- `stdout` 契約不變。

**Test Scope**:
- `specs/truth/features/cli/chat/listing-a-directory.feature`

- [ ] T040 [BDD-GREEN] 讓 Test Scope 全綠（並使 T032 的 list 部分 `[UNIT]` 轉綠）
- [ ] T041 [BDD-REFACTOR] 在綠燈下整理 `list_files` 的輸出構造

## Phase 4D: ADD Feature File - cli/chat/surveying-a-folder-tree.feature

**Goal**: 讓 `surveying-a-folder-tree.feature` 全綠 — 新增 `get_tree` 工具（connector tree、`max_depth` 預設 2、不遞迴 `.git`、結果 1 MiB 上限）；B1 的負向斷言（不顯示超深度 entry、不遞迴 `.git`）可偽。

**Shared Must Read**:
- `specs/truth/features/cli/chat/surveying-a-folder-tree.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "…" whose endpoint shows the folder tree and then answers with "…"`、`tellme showed the folder tree using its get_tree tool`、`the tree shows "…"`、`the tree does not show "…"`、`the tree does not descend into "…"`、`the working directory contains a sub-folder "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`surveying-a-folder-tree.feature`）；registry 的三工具組合由 T038 交付
- `specs/plans/021-tool-surface-parity/research.md` -> Decision 3

**Boundary**:
- 產品碼：新增 `internal/infrastructure/tools/get_tree.go`（connector tree、預設 path/depth、`.git` 不遞迴、1 MiB 上限）。
- `stdout` 契約不變；`go mod` 不變。

**Test Scope**:
- `specs/truth/features/cli/chat/surveying-a-folder-tree.feature`

- [ ] T042 [BDD-GREEN] 讓 Test Scope 全綠（並使 T032 的 tree 部分 `[UNIT]` 轉綠）
- [ ] T043 [BDD-REFACTOR] 在綠燈下整理 `get_tree` 的樹狀構造

## Phase 4E: ADD Feature File - cli/chat/offering-the-reader-tools.feature

**Goal**: 讓 `offering-the-reader-tools.feature` 全綠 — registry 恰提供三個讀取工具；退場工具請求走既有 unknown-tool 契約。

**Shared Must Read**:
- `specs/truth/features/cli/chat/offering-the-reader-tools.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "…" whose endpoint reports the offered tools and then answers with "…"`、`the request offered exactly the reader tools`、`a configured provider "…" whose endpoint asks for a tool that is not available`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`offering-the-reader-tools.feature`）；registry factory 由 T040 交付
- `specs/plans/021-tool-surface-parity/research.md` -> Decision 6

**Boundary**:
- 產品碼：`internal/cli/cli.go` 的 `newToolRegistry` 只組合三個檔案工具。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-reader-tools.feature`

- [ ] T044 [BDD-GREEN] 讓 Test Scope 全綠（並使 T034 的 `[UNIT]` 轉綠）
- [ ] T045 [BDD-REFACTOR] 在綠燈下整理 registry factory

## Phase 4F: CODE-REMOVE — retire `summarize_history`

**Goal**: 移除 `summarize_history` 產品碼與其 registry 註冊，使 truth 的 DELETE 成立（工具與其 truth feature / DSL 皆已移除）。

**Shared Must Read**:
- `truth-delta.md` -> `/axb-dsl-refine` DELETE（`summarising-the-conversation.feature` + 3 rows）；`/axb-technical-research` MODIFY（`techstack.md` 已移除 Session-summarisation tool row）
- `specs/plans/021-tool-surface-parity/research.md` -> Decision 6
- `internal/infrastructure/tools/summarize.go`、`internal/cli/cli.go`

**Boundary**:
- 刪除 `internal/infrastructure/tools/summarize.go`；`newToolRegistry` 不再 append `NewSummarizeHistoryTool`。
- 不得改動 loop 的 bound／failure 契約（`the tool request failed` / exit 7 不變）。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-reader-tools.feature`（reader-tool 集合 + 退場工具請求）

- [ ] T046 [CODE-REMOVE] 移除 `summarize_history` 產品碼與其註冊

## Phase 4G: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞，並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T047 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**（各觀察失敗後還原）：
    1. 讓 `read_files` 只讀 `filepaths` 的第一個路徑 → `chat/reading-several-files.feature` 的雙檔場景失敗。
    2. 讓 `list_files` 維持舊的「names, dirs suffixed /」格式 → `chat/listing-a-directory.feature` 失敗。
    3. 讓 `get_tree` **忽略 `max_depth`（遞迴超過預設深度）** → `surveying-a-folder-tree.feature` 的 `the tree does not show "leaf.go"` 失敗；另**讓 `get_tree` 遞迴 `.git`** → `the tree does not descend into ".git"` 失敗（B1 修正後的雙見證）。
    4. 還原 `summarize_history` 註冊 → `chat/offering-the-reader-tools.feature` 的 offered-tools Then 失敗。
    5. 移除 `logStep` 的 `reason` echo → `chat/watching-the-tool-loop.feature` 的 reason Then 失敗。
    6. 令 `read_files` 的整份結果不受 1 MiB aggregate 上限約束（或移除上限）→ T032 的 aggregate `[UNIT]` 失敗（B2 見證）。
  - 確認：`stdout` byte-exact；exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；offline paths 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T007–T013、T028–T032、T038、T040、T042、T046 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務；T047 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務；T047 驗證 record 形狀不變） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（4 個新 feature） | T007–T031、T038–T045 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop` + `sending-the-configured-persona` + `chat/dsl.md`） | T013、T027、T028–T029、T030、T031、T036、T037 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` DELETE（`summarising-the-conversation` + 3 rows） | T003、T004、T046、T047 | PASS |
| `research.md` -> Decision 1（multi-file `filepaths`） | T005、T011、T017、T031、T032、T038、T039 | PASS |
| `research.md` -> Decision 2（`list_files` shape + default path） | T014、T021–T023、T032、T040、T041 | PASS |
| `research.md` -> Decision 3（add `get_tree`） | T002、T015、T024、T025、T028、T029、T032、T042、T043 | PASS |
| `research.md` -> Decision 4（`reason` required + echoed） | T013、T027、T033、T036、T037 | PASS |
| `research.md` -> Decision 5（cap/binary/dir/≤50 + D3a aggregate 1 MiB） | T008–T010、T012、T018–T020、T030、T032、T038、T047 | PASS |
| `research.md` -> Decision 6（remove `summarize_history`） | T003、T004、T016、T026、T034、T044、T046 | PASS |
| `research.md` -> Decision 7（testing strategy；no pty） | T001、T032–T034、T047 | PASS |
| `research.md` -> Decision 8（no new dependency；no security layer） | T032、T047（`go mod tidy` graph 不變；無 `SafePath`） | PASS |
| `spec.md` -> US1/US2/US3/US4、`FR-001`–`FR-016`、`NFR-001`、`SC-001`–`SC-006` | T007–T034（對齊）、T036–T046（交付）、T047 | PASS |
| `plan.md` -> Source-code structure（`internal/infrastructure/tools`、`internal/cli/cli.go`、`internal/agent/agentloop.go`） | T002、T036–T046 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；CLI end → /axb-dsl-refine） | T002、T047 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
