# Tasks: 029 — Agent write tools (`write_file` + `replace_text`)

**Plan Package**: `specs/plans/029-agent-write-tools`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/**`
> `specs/truth/contracts/**` (`/axb-api-plan` NOOP) and `specs/truth/data/**` (`/axb-data-plan` NOOP) are unchanged this round; there is no `ui/**` (the write tools are model-facing).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍決 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋，不得遺留孤立產物（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 只做本輪新增技術的基礎建設、技術環境與最後的 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架。
- Phase 3 `Test Alignment & Implementation` 的目的：在寫產品碼之前，先把本輪所有受影響 DSL 的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

## Phase 1: Setup

**Omitted** — this round introduces **no new technology**: the two write tools use stdlib `os` / `path/filepath` / `strings` / `encoding/json`, and they slot into the existing `domain/tools.Tool` port and the round-024 resource contract; `go.mod`/`go.sum` stay unchanged (`research.md` Decision 8). There is nothing to install and no smoke-test to add.

## Phase 2: Foundational

**Goal**: 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。不寫 Phase 3 測試語意，也不寫 Feature Green。

- [ ] T001 建立 write tools 落點 `internal/infrastructure/tools/writer.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Write filesystem tools row)
    - `specs/plans/029-agent-write-tools/research.md` -> `Decision 1`, `Decision 2`, `Decision 3`, `Decision 4`
    - `internal/infrastructure/tools/filesystem.go` -> the existing `Tool` port implementation pattern (reader)
  - 只做：建立 `writeFile` 與 `replaceText` 兩個工具殼（`Name`/`Description`/`Parameters` JSON schema 帶 `max_output_tokens`/`timeout`、`reason` required；`Contract()` 回 30s；`Execute` 簽名），並留 `NewWriteTools() []domaintools.Tool` 建構子。
  - 不做：不寫建立/atomic/strict-unique 邏輯（留 Phase 4），不碰 `cli.go`，不寫安全/consent/undo（D5）。

- [ ] T002 預留 CLI 工具註冊接點 `internal/cli/cli.go`
  - Read:
    - `truth-delta.md` -> `/axb-technical-research` MODIFY techstack (Write filesystem tools row)
    - `specs/plans/029-agent-write-tools/research.md` -> `Decision 1` (register in `newToolRegistry`)
  - 只做：在 `newToolRegistry` 既有的工具集合組裝處，留一個接點把 `tools.NewWriteTools()` 併入註冊（offer order：readers、write pair、command）。
  - 不做：不寫任何工具邏輯，不改 flag/exit/stdout 契約。

- [ ] T003 預留 Phase 3 stepdef 獨立落點骨架（Zero Shared Edits 原則）
  - Read:
    - `tests/e2e/steps/` -> 既有 `step_r0NN_tNNN_*.go` 獨立檔＋`init()` 自我註冊慣例
    - `specs/truth/features/cli/chat/dsl.md` -> 本輪新增/修改的句列（見 Phase 3 `DSL 參照`）
  - 只做：在 `tests/e2e/steps/` 下，為 Phase 3 每一句建立**獨立**檔案骨架（`step_r029_tNNN_chat_{given|when|then}_<slug>.go`，`init()` 註冊、body 待填），使 Phase 3 並行任務目標檔案互斥。
  - 不做：不寫任何 `[BDD-ALIGN]`/`[BDD-RED]` 語意，不寫 helper 邏輯。

- [ ] T004 預留單元測試落點骨架 `internal/infrastructure/tools/writer_test.go`
  - Read:
    - `specs/plans/029-agent-write-tools/research.md` -> `Decision 2`, `Decision 3`, `Decision 4`, `Decision 8`
  - 只做：建立單元落點檔案骨架（create-only、atomicity、strict-unique、empty `old_text`、missing file、parent creation 的測試函式殼）。
  - 不做：不寫斷言（留 Phase 3 `[UNIT]`）。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 truth-delta 有改的句，以及本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪各句都在同模組 `specs/truth/features/cli/chat/dsl.md`。本輪用到的介面根句（`the operator has a runnable tellme installation`、`the runtime home is "..."`、`tellme exits successfully`、`the operator starts tellme with the prompt "..."`）早已有 stepdef，不列。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，但語意是舊 truth。依 `dsl.md` 該列改測試，讓它表達最新版 `StepDef 實作語意`。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- 兩個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the request offered exactly the agent tools`（集合改為六工具）
  -> 本輪新增的 6 條 Given（working-dir no file / no folder / file whose lines are / file with a line twice / provider creates a file / provider edits a file）
  -> 本輪新增的 10 條 Then（file created / folder created / content exactly / content still exactly / lines are / lines are still / creation refused / edit refused not present / edit refused not unique / file still holds a line twice）
- `truth-delta.md` -> `/axb-dsl-refine` ADD `creating-and-editing-files.feature` + MODIFY `chat/dsl.md`
- `specs/plans/029-agent-write-tools/research.md` -> `Decision 7`（markers/failure shape）

**Boundary**:
- 一條 DSL 一個 task。
- 只改該句的 stepdef / assertion / 直接依賴的 helper。
- 落點檔案採獨立檔案（Zero Shared Edits 原則，SHOULD），使並行派出具備互斥寫入目標。
- 不寫產品碼。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。有 issues 就修正再 review，直到沒有任何問題。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T006–T021 各派一個獨立 subagent，一次整批並行 dispatch（目標檔獨立，符合 Zero Shared Edits）；T022 同批並行；T023 等全部回來再啟動 subagent review。

- [ ] T005 [BDD-ALIGN] `the request offered exactly the agent tools`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools`; `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go`
  - 只做：把該 assertion 的期望集合改為六工具（`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`, `execute_command`）。
  - 附註（PR #61 review finding 7）：期望集合**不**再手抄第三份；stepdef 以 `tellme --tool-usage`（列出 live registry 的工具）導出期望集合，再與 offered set 比對，讓工具新增不會與 `dsl.md` 的 `集合` 及 acceptance prose 三處漂移。
- [ ] T006 [P] [BDD-RED] `the working directory contains no file "{name}"`
  - Read: `tests/e2e/steps/step_r029_t006_chat_given_workdir_no_file.go`
- [ ] T007 [P] [BDD-RED] `the working directory contains no folder "{name}"`
  - Read: `tests/e2e/steps/step_r029_t007_chat_given_workdir_no_folder.go`
- [ ] T008 [P] [BDD-RED] `the working directory contains a file "{name}" whose lines are:`
  - Read: `tests/e2e/steps/step_r029_t008_chat_given_workdir_lines.go`
- [ ] T009 [P] [BDD-RED] `the working directory contains a file "{name}" that contains the line "{line}" twice`
  - Read: `tests/e2e/steps/step_r029_t009_chat_given_workdir_line_twice.go`
- [ ] T010 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint creates the file "{path}" with the content "{content}" and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r029_t010_chat_given_provider_creates_file.go`
- [ ] T011 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint edits the file "{path}" replacing "{old}" with "{new}" and then answers with "{answer}"`
  - Read: `tests/e2e/steps/step_r029_t011_chat_given_provider_edits_file.go`
- [ ] T012 [P] [BDD-RED] `tellme created the file "{name}"`
  - Read: `tests/e2e/steps/step_r029_t012_chat_then_file_created.go`
- [ ] T013 [P] [BDD-RED] `the content of "{name}" is exactly "{content}"`
  - Read: `tests/e2e/steps/step_r029_t013_chat_then_content_exact.go`
- [ ] T014 [P] [BDD-RED] `the content of "{name}" is still exactly "{content}"`
  - Read: `tests/e2e/steps/step_r029_t014_chat_then_content_unchanged.go`
- [ ] T015 [P] [BDD-RED] `the lines of "{name}" are:`
  - Read: `tests/e2e/steps/step_r029_t015_chat_then_lines.go`
- [ ] T016 [P] [BDD-RED] `the lines of "{name}" are still:`
  - Read: `tests/e2e/steps/step_r029_t016_chat_then_lines_unchanged.go`
- [ ] T017 [P] [BDD-RED] `the folder "{name}" was created`
  - Read: `tests/e2e/steps/step_r029_t017_chat_then_folder_created.go`
- [ ] T018 [P] [BDD-RED] `the creation is refused`
  - Read: `tests/e2e/steps/step_r029_t018_chat_then_creation_refused.go`
- [ ] T019 [P] [BDD-RED] `the edit is refused because the block is not present`
  - Read: `tests/e2e/steps/step_r029_t019_chat_then_edit_absent.go`
- [ ] T020 [P] [BDD-RED] `the edit is refused because the block is not unique`
  - Read: `tests/e2e/steps/step_r029_t020_chat_then_edit_ambiguous.go`
- [ ] T021 [P] [BDD-RED] `the file "{name}" still contains the line "{line}" twice`
  - Read: `tests/e2e/steps/step_r029_t021_chat_then_line_still_twice.go`
- [ ] T022 [P] [UNIT] write tools 單元測試
  - Read: `internal/infrastructure/tools/writer.go`, `internal/infrastructure/tools/writer_test.go`, `research.md` -> `Decision 2`, `Decision 3`, `Decision 4`
  - 必查：create-only（既有檔 → error 且內容不變）；atomicity（temp+rename；失敗不留半檔；同目錄 rename）；`replace_text` strict-unique（0 → error、>1 → error、恰好 1 → 替換）；empty `old_text` → error；missing file → error（不建立）；missing parent → `MkdirAll`。
  - 並以 **atomicity witness**（review finding 4）取證：注入一個 `Write` 在第 N byte 失敗的 writer，斷言 destination **不存在或 byte-identical**，且**無 `*.tmp` 殘留**（atomicity 為 unit-tier，E2E 無 fault injection）。
  - 另 pin：建立檔案的 mode 為 **`0644`**（review finding 3）；create-only 以 **atomic move**（`os.Link`/`EEXIST`）成立、**非** `Stat`/`rename` TOCTOU，且既有檔出現時不被 clobber（review finding 2）。
- [ ] T023 subagent review (phase quality gate)

## Phase 4A: ADD Feature File - cli/chat/creating-and-editing-files.feature

**Goal**: 以最小 `write_file` + `replace_text` 產品邏輯（create-only + atomic temp+rename + `MkdirAll`；strict-unique replace）讓此 feature 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/creating-and-editing-files.feature` -> `Feature: Creating and editing files`
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 6 條 Given + 10 條 Then
- `truth-delta.md` -> `/axb-dsl-refine` ADD creating-and-editing-files
- `specs/plans/029-agent-write-tools/research.md` -> `Decision 2`, `Decision 3`, `Decision 4`, `Decision 6`, `Decision 7`

**Boundary**:
- 只處理兩個 write 工具的產品碼；不改 reader trio / `execute_command` / flag / exit / stdout 契約。
- 依 research：`write_file` create-only（既有檔 → tool error，檔案不動）；atomic（temp file in target dir + `rename`，失敗清 temp，不留半檔）；`replace_text` strict-unique（0 / >1 → tool error，檔案不動）；失敗為非終端 tool error。
- Register in `internal/cli/cli.go`（offer order：readers、write pair、command）。

**Test Scope**:
- `specs/truth/features/cli/chat/creating-and-editing-files.feature`

- [ ] T024 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T025 [BDD-REFACTOR] 在綠燈下整理 write/atomic/replace 共用邏輯

## Phase 4B: MODIFY Feature File - cli/chat/offering-the-agent-tools.feature

**Goal**: 讓 offer set 成為六工具（三 reader + write pair + command），且不提供 `pipe_commands`/summarisation。**本 phase 只做產品層的註冊與轉綠** — truth feature 的 prose（header comment + Example 標題）已在 **truth half** 由 `/axb-dsl-refine` 更新，此處不再改 truth 文字。

**Shared Must Read**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature` -> `Feature: Offering the agent tools`
- `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY `offering-the-agent-tools.feature`（prose → 六工具）+ MODIFY `chat/dsl.md`（offered 集合）

**Boundary**:
- 只做**產品層**的工具註冊/轉綠；不新增 `pipe_commands`、不引入安全/consent（D5）；**不改** truth feature 的 prose（已於 truth half 完成）。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature`

- [ ] T026 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T027 [BDD-REFACTOR] 在綠燈下整理工具註冊與 offered-tools 投影

## Phase 4C: REGRESSION

**Goal**: 全 CLI truth feature 回歸 + falsifiability witnesses + `make verify` + 拓樸稽核。

**Shared Must Read**:
- `specs/truth/features/cli/**`（全）
- `specs/plans/029-agent-write-tools/research.md` -> `Decision 8`（witnesses）

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses：(a) 取消 create-only → 覆寫被拒的 Example 失敗；(b) 移除 temp+rename → atomicity 見證失敗；(c) `replace_text` 放寬為 replace-first → 歧義 Example 失敗。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T028 [REGRESSION] 跑全 CLI feature + 見證 + `make verify` + 拓樸稽核

---

## Pre-Delivery Orphan Coverage Sweep

**`truth-delta.md` 非 NOOP rows → task 對應**
- `/axb-technical-research` MODIFY `techstack.md`（Write filesystem tools row + *Not Introduced Yet* bullet）→ T001/T002（Foundational）、T022（UNIT）、Phase 4A/4B。
- `/axb-dsl-refine` ADD `creating-and-editing-files.feature` → Phase 4A。MODIFY `chat/dsl.md`（offered 集合 + 6 Given + 10 Then + module note）→ T005（ALIGN）、T006–T021（RED）、Phase 4A/4B。
- `/axb-api-plan` NOOP、`/axb-data-plan` NOOP → 免建立任務（審計豁免）。

**`research.md` 已拍板 Decisions → task `Read` 覆蓋**
- D1 → T001/T002/Phase 4A；D2 → T001/T022/Phase 4A；D3 → T001/T022/Phase 4A；D4 → T001/T022/Phase 4A；D5 → T001（no security/undo）；D6 → T001（resource params + 30s）；D7 → T009–T021/Phase 4A；D8 → T022/T028（hermetic、witnesses）。Truth impact（techstack / dsl-refine / api+data NOOP）→ 對應 `truth-delta.md` rows。

**`specs/truth/techstack.md` 異動章節 → 建置/驗證 task 覆蓋**
- 本輪**無新增技術**（stdlib-only），故無 Setup 建置/版本/make target 變更；相關章節由 T001/T002 與 T022 承接並以 `Read` 引用。

**孤立產物件數：0。掃描通過，准予交付。**
