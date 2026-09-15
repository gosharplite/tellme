# Tasks: 024 — Tool resource contract, `execute_command` & reader retrofit

**Plan Package**: `specs/plans/024-tool-resource-contract-and-execute-command`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/**`
> `specs/truth/contracts/**` (`/axb-api-plan` NOOP) and `specs/truth/data/**` (`/axb-data-plan` NOOP) are unchanged this round; there is no `ui/**` (the tool surface is model-facing).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍決 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋，不得遺留孤立產物（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task（如 build 參數、version、make target），不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 只做本輪新增技術的基礎建設、技術環境與最後的 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架。
- Phase 3 `Test Alignment & Implementation` 的目的：在寫產品碼之前，先把本輪所有受影響 DSL 的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

## Phase 1: Setup

**Omitted** — this round introduces **no new technology**: the command tool uses stdlib `os/exec`, the readers stay `os`/`io`, and the contract is hand-written Go; `go.mod`/`go.sum` stay unchanged (`research.md` Decision 8; `techstack.md` *Not Introduced Yet*). There is nothing to install and no smoke-test to add.

## Phase 2: Foundational

**Goal**: 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。不寫 Phase 3 測試語意，也不寫 Feature Green。

- [x] T001 建立純契約解析器落點 `internal/agent/tool_contract.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Tool resource contract row)
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 4` (loop enforcement + pure resolver)
  - 只做：建立檔案與未實作的函式簽名 `resolveBound(param, ceiling, def int) int` 與 `clampBytes(result string, byteBudget int) string`，讓 `internal/agent` 之後可呼叫。
  - 不做：不寫解析/夾取/裁剪邏輯（留 Phase 4），不碰 `agentloop.go`，不寫任何 DSL 語意。

- [x] T002 建立視窗設定落點 `internal/config/config.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Model pricing & context window row)
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 5` (effective budget)
  - 只做：在 `ModelPricing` 加 `ContextWindow int`（yaml `CONTEXT_WINDOW`）欄位，並留 `ContextWindowFor(model string) (int, bool)` 的函式殼（與 `PricingFor` 分開）。
  - 不做：不接進 `resolve()`，不寫讀取/解析邏輯（留 Phase 4），不改 env 覆寫面。

- [x] T003 預留 `Tool` port 兩向加寬 `internal/domain/tools/tools.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Agent tool loop row)
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 4` (two-way port; upward default descriptor + downward byte budget)
  - 只做：加一個向上契約描述元（如 `Contract() ToolContract{ DefaultTimeout time.Duration }`）與把 `Execute` 改成 `Execute(ctx, arguments string, budget ByteBudget)`；同步更新兩個既有 test double（`internal/agent/agentloop_test.go` 的 `fakeTool`、`internal/domain/tools/tools_test.go` 的 `stubTool`）使其可編譯。
  - 不做：不實作任何工具的 `Contract`/`Execute` 邏輯，不改 registry 行為。

- [x] T004 建立命令工具落點 `internal/infrastructure/tools/command.go`
  - Read:
    - `truth-delta.md` -> `/axb-dsl-refine` ADD `chat/running-a-shell-command.feature`
    - `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint runs a command and then answers with "{answer}"`
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 1`, `Decision 8`
  - 只做：建立 `execute_command` 工具殼（`Name`/`Description`/`Parameters` JSON schema + `Contract`），註冊入口留給 `cli.go`。
  - 不做：不寫 `os/exec`/process-group/pipe 擷取邏輯（留 Phase 4），不寫安全/consent（D1）。

- [x] T005 預留 reader retrofit 落點 `internal/infrastructure/tools/filesystem.go` + `get_tree.go`
  - Read:
    - `truth-delta.md` -> `/axb-dsl-refine` MODIFY `chat/reading-several-files.feature`
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 6` (one aggregate bound; retire fixed caps)
  - 只做：把 `readMaxPerFile`/`readAggregateCap` 固定常數換成可傳入的 aggregate **byte** bound 之介面接點，讓後續能以參數化 bound 讀取。
  - 不做：不寫增量讀取/skip marker/truncation marker 語意（留 Phase 4）。

- [x] T006 預留 loop 契約 enforcement seam `internal/agent/agentloop.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Agent tool loop row)
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 4`
  - 只做：在 `Run` 的工具執行處留呼叫 T001 純解析器與 `Execute(ctx, arguments, byteBudget)` 的接點（單一解析/強制點）。
  - 不做：不寫解析/夾取/裁剪邏輯，不改 wire 時序或 observer hooks。

- [x] T007 預留 CLI resolution 與工具註冊接點 `internal/cli/cli.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Payload budget / Payload status line rows)
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 5` (default/ceiling) + `Decision 5` forward (payload-line displays the effective budget)
  - 只做：`resolution` 加 `EffectiveBudget int` 欄位並在 `resolve()` 留計算接點；`newToolRegistry` 留 `execute_command` 註冊接點；留一次性「無視窗」log 接點；把 loop literal 指向帶 `EffectiveBudget` 的接點。
  - 不做：不寫 `min(...)`/÷4/÷2 與 log 邏輯（留 Phase 4），不改 flag/exit/stdout 契約。

- [x] T008 預留 Phase 3 stepdef 獨立落點骨架（Zero Shared Edits 原則）
  - Read:
    - `tests/e2e/steps/` -> 既有 `step_r0NN_tNNN_*.go` 的獨立檔＋`init()` 自我註冊慣例
    - `specs/truth/features/cli/chat/dsl.md` -> 本輪新增/修改的句列（見 Phase 3 `DSL 參照`）
  - 只做：在 `tests/e2e/steps/` 下，為 Phase 3 每一句建立**獨立**檔案骨架（`step_r024_tNNN_chat_{given|when|then}_<slug>.go`，`init()` 註冊、body 待填），使 Phase 3 並行任務目標檔案互斥。
  - 不做：不寫任何 `[BDD-ALIGN]`/`[BDD-REMOVE]`/`[BDD-RED]` 語意，不寫 helper 邏輯。

- [x] T009 預留單元測試落點骨架 `internal/**`（contract/config/command/reader）
  - Read:
    - `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 4`, `Decision 5`, `Decision 8`
  - 只做：建立單元落點檔案骨架（契約解析器、視窗解析、命令工具、reader bound、reader timeout-result）。
  - 不做：不寫斷言（留 Phase 3 `[UNIT]`）。

- [x] T010 建立 process-tree witness 的量測 helper 落點
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the command's descendant process is no longer running`
    - `tests/e2e/harness/` -> 既有 E2E helper 掛載點
  - 只做：建立「執行後以 bounded poll 檢查某 pid 是否存活（`syscall.Kill(pid,0)`→`ESRCH`）＋檢查 sentinel 是否缺席」的量測 helper 殼。
  - 不做：不寫命令工具產品碼，不決定 fixture 內容（fixture 由 Given 決定）。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 truth-delta 有改的句，以及本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪各句都在同模組 `specs/truth/features/cli/chat/dsl.md`。本輪未使用介面根新增句（用到的 `the operator has a runnable tellme installation`、`the runtime home is "..."`、`tellme exits successfully` 早已有 stepdef，不列）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`、`再讀確認`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，但語意是舊 truth。依 `dsl.md` 該列改測試，讓它表達最新版 `StepDef 實作語意`。
- `[BDD-REMOVE]`：`DELETE`。此句已不是 truth。移除或改寫仍綁這句的 stepdef / assertion，不得留下保護舊行為的測試。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the working directory contains a file "{name}" whose text is longer than the read bound`
  -> `the request offered exactly the agent tools`
  -> 本輪新增的 8 條 `a configured provider "..." whose endpoint ...` Givens
  -> 本輪新增的 10 條 command/reader/list/tree Thens
- `truth-delta.md` -> `/axb-dsl-refine` 有對應 ADD / MODIFY / DELETE 的句
- `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 7`（markers）、`Decision 4`/`Decision 5`（bound）

**Boundary**:
- 一條 DSL 一個 task。
- 只改該句的 stepdef / assertion / 直接依賴的 helper。
- 落點檔案採獨立檔案（Zero Shared Edits 原則，SHOULD），使並行派出具備互斥寫入目標。
- 不寫產品碼。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。有 issues 就修正再 review，直到沒有任何問題。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T011–T031 各派一個獨立 subagent，一次整批並行 dispatch（目標檔獨立，符合 Zero Shared Edits）；T032–T036 同批並行；T037 等全部回來再啟動 subagent review。

- [x] T011 [P] [BDD-ALIGN] `the working directory contains a file "{name}" whose text is longer than the read bound`
  - Read: `tests/e2e/steps/step_r021_t008_chat_given_workdir_large_file.go`
- [x] T012 [P] [BDD-ALIGN] `the request offered exactly the agent tools`
  - Read: `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go`（舊句 `the request offered exactly the reader tools`）
- [x] T013 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t013_chat_given_provider_runs_command.go`
- [x] T014 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command that exits non-zero and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t014_chat_given_provider_command_nonzero.go`
- [x] T015 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command that never returns and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t015_chat_given_provider_command_hangs.go`
- [x] T016 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command that spawns a long-lived descendant and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t016_chat_given_provider_command_descendant.go`
- [x] T017 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command producing a great deal of output and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t017_chat_given_provider_command_output.go`
- [x] T018 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint runs a command that writes its output to a file and then reads it and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t018_chat_given_provider_command_outputfile.go`
- [x] T019 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint lists the current directory with a small result budget and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t019_chat_given_provider_list_smallbudget.go`
- [x] T020 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint shows the folder tree with a small result budget and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r024_t020_chat_given_provider_tree_smallbudget.go`
- [x] T021 [P] [BDD-RED] `the working directory contains a file "{name}" whose text is longer than the old per-file cap`
  - Read: `tests/e2e/steps/step_r024_t021_chat_given_workdir_oldcap_file.go`
- [x] T022 [P] [BDD-RED] `tellme ran the command using its execute_command tool`
  - Read: `tests/e2e/steps/step_r024_t022_chat_then_ran_command.go`
- [x] T023 [P] [BDD-RED] `the command result carried the exit status "{code}"`
  - Read: `tests/e2e/steps/step_r024_t023_chat_then_command_exit.go`
- [x] T024 [P] [BDD-RED] `the command result recorded that the command was stopped`
  - Read: `tests/e2e/steps/step_r024_t024_chat_then_command_stopped.go`
- [x] T025 [P] [BDD-RED] `the command's descendant process is no longer running`
  - Read: `tests/e2e/steps/step_r024_t025_chat_then_descendant_gone.go`
- [x] T026 [P] [BDD-RED] `the command result was trimmed to what the run can hold`
  - Read: `tests/e2e/steps/step_r024_t026_chat_then_command_trimmed.go`
- [x] T027 [P] [BDD-RED] `the command result reported that its output went to "{path}"`
  - Read: `tests/e2e/steps/step_r024_t027_chat_then_command_output_target.go`
- [x] T028 [P] [BDD-RED] `the listing was trimmed to what the run can hold`
  - Read: `tests/e2e/steps/step_r024_t028_chat_then_listing_trimmed.go`
- [x] T029 [P] [BDD-RED] `the tree was trimmed to what the run can hold`
  - Read: `tests/e2e/steps/step_r024_t029_chat_then_tree_trimmed.go`
- [x] T030 [P] [BDD-RED] `the read result reports that "{name}" was not read`
  - Read: `tests/e2e/steps/step_r024_t030_chat_then_read_skipped.go`
- [x] T031 [P] [BDD-RED] `the part of "{name}" that tellme read does not end with a truncation marker`
  - Read: `tests/e2e/steps/step_r024_t031_chat_then_read_not_truncated.go`
- [x] T032 [P] [UNIT] 純契約解析器 `resolveBound` + `clampBytes`
  - Read: `internal/agent/tool_contract.go`, `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 4`
  - 必查：param→ceiling→default 與「恰好在 bound / 超出一個 byte」邊界；loop clamp 用 raw `len`，不呼叫 estimator。
- [x] T033 [P] [UNIT] 視窗解析 `ContextWindowFor` + `effectiveBudget` + 一次性無視窗 log
  - Read: `internal/config/config.go`, `internal/cli/cli.go`, `research.md` -> `Decision 5`
- [x] T034 [P] [UNIT] reader timeout-as-nil-error-result（FR-018 / T4）
  - Read: `internal/infrastructure/tools/filesystem.go`, `spec.md` -> `FR-018`
  - 必查：reader 觀察到 deadline 時回 nil-error 的 timeout result（不再 `ctx.Err()`→error）。此為 FR-018 reader 路徑的**錄定 witness**（Gherkin `Then` 無法實用腳本化 30s 預設）。
- [x] T035 [P] [UNIT] 命令工具：byte-trim 生命週期、exit-status、process group
  - Read: `internal/infrastructure/tools/command.go`, `research.md` -> `Decision 1a`, `Decision 8`（T1 lifecycle）
  - 必查：stop→close read-ends→kill(-pgid) if alive→Wait；trimmed 不把 141/137 當命令 exit status。
- [x] T036 [P] [UNIT] reader aggregate byte bound + skip marker + 固定 cap 退場
  - Read: `internal/infrastructure/tools/filesystem.go`, `get_tree.go`, `research.md` -> `Decision 6`
- [x] T037 subagent review (phase quality gate)

## Phase 4A: ADD Feature File - cli/chat/running-a-shell-command.feature

**Goal**: 以最小 `execute_command` 產品邏輯（`bash -c`、bounded pipes、process-group timeout、`output_file`/`append`、非零 exit＝成功結果、nil-error timeout result）讓此 feature 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/running-a-shell-command.feature` -> `Feature: Running a shell command`
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 6 條 command Given + 相關 Then
- `truth-delta.md` -> `/axb-dsl-refine` ADD running-a-shell-command
- `specs/plans/024-tool-resource-contract-and-execute-command/research.md` -> `Decision 1`, `Decision 1a`, `Decision 2`, `Decision 3`, `Decision 8`；`Decision 7`（trim vs stopped）

**Boundary**:
- 只處理命令工具與其資源契約；不處理 reader/list/tree 的產品碼，不改 flag/exit/stdout 契約。
- 依 research Decision 8：`StdoutPipe`/`StderrPipe` 有界讀取；達 byte budget 走 **T1 生命週期**；`output_file` 直綁檔案；`stdout` 維持 byte-exact。

**Test Scope**:
- `specs/truth/features/cli/chat/running-a-shell-command.feature`

- [x] T038 [BDD-GREEN] 讓 Test Scope 全綠
- [x] T039 [BDD-REFACTOR] 在綠燈下整理命令擷取與 process-group 生命週期共用邏輯

## Phase 4B: MODIFY Feature File - cli/chat/reading-several-files.feature

**Goal**: 讓 reader 以參數化 aggregate **byte** bound 讀整檔、超過時以 marker 與 skip 回報。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reading-several-files.feature` -> `Feature: Reading several files in one request`
- `specs/truth/features/cli/chat/dsl.md` -> `the working directory contains a file "{name}" whose text is longer than the read bound`, `the part of "{name}" that tellme read does not end with a truncation marker`, `the read result reports that "{name}" was not read`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY reading-several-files
- `research.md` -> `Decision 6`（one aggregate bound；cap 退場）

**Boundary**:
- 只改 `read_files` 的 bound 與 marker/skip；保留 ≤50 與 `--- File: … ---` framing；不改 `list_files`/`get_tree`（各自 phase）。

**Test Scope**:
- `specs/truth/features/cli/chat/reading-several-files.feature`

- [x] T040 [BDD-GREEN] 讓 Test Scope 全綠
- [x] T041 [BDD-REFACTOR] 在綠燈下整理聚合 byte bound 與 skip marker 共用邏輯

## Phase 4C: ADD Feature File - cli/chat/offering-the-agent-tools.feature

**Goal**: 讓 registry 提供恰好四個 agent tools（三 reader + `execute_command`），且不提供 `pipe_commands`/summarisation。

**Shared Must Read**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature` -> `Feature: Offering the agent tools`
- `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools`
- `truth-delta.md` -> `/axb-dsl-refine` ADD offering-the-agent-tools

**Boundary**:
- 只改工具集合的組成；不新增 `pipe_commands`、不引入安全/consent（D1）。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature`

- [x] T042 [BDD-GREEN] 讓 Test Scope 全綠
- [x] T043 [BDD-REFACTOR] 在綠燈下整理工具註冊與 offered-tools 投影

## Phase 4D: DELETE Feature / DSL Truth - cli/chat/offering-the-reader-tools.feature

**Goal**: 舊的 reader-only 提供面已由 `offering-the-agent-tools.feature` 取代；移除僅保護舊工具集合的斷言/分支。

**Shared Must Read**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature`（取代者）
- `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools`
- `truth-delta.md` -> `/axb-dsl-refine` DELETE offering-the-reader-tools
- `internal/infrastructure/tools/` -> 可能仍只列 reader 的舊 offered-tools 分支

**Boundary**:
- 只移除「恰好只有 reader」的舊提供面斷言/投影；不移除四個工具本身，也不移除未知工具契約。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature`

- [x] T044 [CODE-REMOVE] 移除舊 reader-only 提供面的斷言/產品分支
- [x] T045 [REGRESSION] 跑 Test Scope，確認新版 truth（四工具）成立

## Phase 4E: MODIFY Feature File - cli/chat/listing-a-directory.feature

**Goal**: `list_files` 以共享 byte bound 綁定，超出時帶 truncation marker（FR-011 witness）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/listing-a-directory.feature` -> `Feature: Listing a directory`
- `specs/truth/features/cli/chat/dsl.md` -> `the listing was trimmed to what the run can hold`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY listing-a-directory
- `research.md` -> `Decision 6`, `Decision 7`

**Boundary**:
- 只改 `list_files` 的 bound 與 marker；輸出形狀（`Contents of …`、`[d]`/`[f]`）不變。

**Test Scope**:
- `specs/truth/features/cli/chat/listing-a-directory.feature`

- [x] T046 [BDD-GREEN] 讓 Test Scope 全綠
- [x] T047 [BDD-REFACTOR] 在綠燈下整理共享 truncateToCap 路徑

## Phase 4F: MODIFY Feature File - cli/chat/surveying-a-folder-tree.feature

**Goal**: `get_tree` 以同一 byte bound 綁定，超出時帶 truncation marker（FR-011 witness）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/surveying-a-folder-tree.feature` -> `Feature: Surveying a folder tree`
- `specs/truth/features/cli/chat/dsl.md` -> `the tree was trimmed to what the run can hold`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY surveying-a-folder-tree
- `research.md` -> `Decision 6`, `Decision 7`

**Boundary**:
- 只改 `get_tree` 的 bound 與 marker；connector 樹狀與 `max_depth`/`.git` 行為不變。

**Test Scope**:
- `specs/truth/features/cli/chat/surveying-a-folder-tree.feature`

- [x] T048 [BDD-GREEN] 讓 Test Scope 全綠
- [x] T049 [BDD-REFACTOR] 在綠燈下整理樹狀輸出與共享 bound

---

## Pre-Delivery Orphan Coverage Sweep

**`truth-delta.md` 非 NOOP rows → task 對應**
- `/axb-technical-research` MODIFY `techstack.md`（Tool resource contract · Agent command tool · Read-only filesystem tools · Model pricing & context window · Payload budget · Payload status line · Agent tool loop · pure-helper unit tests）→ T002/T003/T004/T005/T006/T007（Foundational）、T032–T036（Phase 3 UNIT）、Phase 4A/4B/4E/4F。
- `/axb-dsl-refine` ADD `running-a-shell-command.feature` → Phase 4A。ADD `offering-the-agent-tools.feature` → Phase 4C。DELETE `offering-the-reader-tools.feature` → T012（ALIGN 改名列）、Phase 4D。MODIFY `reading-several-files.feature` → T011/T031、Phase 4B。MODIFY `listing-a-directory.feature` / `surveying-a-folder-tree.feature` → T028/T029、Phase 4E/4F。MODIFY `chat/dsl.md` → T011–T031。
- `/axb-api-plan` NOOP、`/axb-data-plan` NOOP → 免建立任務（審計豁免）。

**`research.md` 已拍板 Decisions → task `Read` 覆蓋**
- D1 → T004/T038；D1a → T035/T038；D2 → T023/T038；D3 → T018/T027/T038；D4 → T001/T003/T006/T032；D5 → T002/T007/T033；D6 → T005/T036/Phase 4B/4E/4F；D7 → T024/T026/T028/T029；D8 → T004/T010/Phase 4A（hermetic 無新依賴）。Forward（wall-clock / 常數獨立 / payload-line display）→ T007/T033。

**`specs/truth/techstack.md` 異動章節 → 建置/驗證 task 覆蓋**
- 本輪**無新增技術**（stdlib-only），故無 Setup 建置/版本/make target 變更；相關章節由 T002–T007 與 Phase 3 UNIT 承接並以 `Read` 引用。

**孤立產物件數：0。掃描通過，准予交付。**
