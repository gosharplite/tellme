# Tasks: Payload Status Line (round 009)

**Plan Package**: `specs/plans/009-payload-status-line`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only — `time`、`strings`／`unicode/utf8`；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 2: Foundational

**Goal**: 建立 token estimator 落點、status-line formatter 落點、`llm.Response` usage 欄位、`MAX_HISTORY_TOKENS` 解析 stub、CLI 狀態行落點，以及 E2E 共用元件（fake 供應/隱藏 usage、stderr 擷取、budget 環境注入）與 9 個新句 stepdef 落點骨架（Zero Shared Edits 原則）。

- [ ] T001 建立 token estimator 落點骨架 `internal/domain/llm/token.go`
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 1, 6
    - `specs/truth/techstack.md` -> CLI Application（Token estimator）
  - 只做：宣告 `EstimateTokens(messages []Message) int` 純函式簽名 stub（stdlib-only、deterministic）與固定 bytes-per-token 比例常數槽；不呼叫 provider、不涉網路。
  - 不做：不實作估計算式；不接 `internal/cli`。

- [ ] T002 建立 status-line formatter 落點骨架 `internal/ui/status.go`
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 3, 4
    - `specs/truth/techstack.md` -> CLI Application（Payload status line）
  - 只做：宣告 payload status line formatter 的簽名 stub（輸入 tokens/budget/mode/model + **注入的 clock seam**；輸出 pre-flight `~` 與 post-turn 兩型）；對齊 `[HH:MM:SS] Payload: <n>/<max> tokens - <mode> - <model>` 形狀；**不得**加 `tellme:` 前綴。
  - 不做：不決定何時輸出；不接 CLI 分流。

- [ ] T003 擴充 `llm.Response` usage 欄位與 OpenAI adapter 落點骨架
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port / Response normalization）
    - `internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`
  - 只做：`llm.Response` 新增 `Usage`（prompt/completion/total tokens）型別骨架；在 `parseResponse` 留下解析回應 `usage` 區塊的 stub（無 `usage` 時為 zero-value）。**向後相容**：不帶 usage 的既有回應語意不變。
  - 不做：不改 transport、錯誤處理或既有 content/tool_calls 解析。

- [ ] T004 落點骨架 `internal/config/config.go`（`MAX_HISTORY_TOKENS`）
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 5
    - `specs/truth/techstack.md` -> Configuration（Payload budget）
    - `internal/config/config.go`（`EffectiveMaxToolLoop` 鄰）
  - 只做：`Config` 新增 `MaxHistoryTokens int`（`yaml:"MAX_HISTORY_TOKENS"`）與 `DefaultMaxHistoryTokens = 1000000` 常數；留下 `EffectiveMaxHistoryTokens(override string) (int, error)` 的 stub（env-over-file、預設 1000000、`>= 0`、非整數/負值 → `ErrInvalidValue`）。
  - 不做：不接 `resolve()`；不寫斷言（產品行為留 Phase 4）。

- [ ] T005 落點骨架 `internal/cli/cli.go`（狀態行接線）
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 3, 4, 5, 7
    - `specs/truth/techstack.md` -> CLI Application（Payload status line / Token estimator）、Configuration（Payload budget）
    - `internal/cli/cli.go`
  - 只做：在 `resolution` 落點新增 `MaxHistoryTokens`；在 `resolve()` 留下 `EffectiveMaxHistoryTokens` 呼叫的 stub；在 prompt turn 落點留下「送出前寫 pre-flight 行、完成後寫 post-turn 行到 `stderr`」的 hook 簽名（皆為 stub）。狀態行只落 `stderr`。
  - 不做：不實作估計算式／formatter 內容；不改既有 dispatch、`stdout` 答案流、exit-code 或離線路徑。

- [ ] T006 建立 E2E 共用元件落點骨架
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 8, 9
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner / Local fake provider / Pure-helper unit tests）
    - `tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 只做：擴充 fake provider 以**供應**帶 `usage` 的回應並可**隱藏** `usage`；新增擷取 subprocess `stderr` 的 harness helper；新增設定 `MAX_HISTORY_TOKENS` 的環境注入。
  - 不做：不寫具體 step 斷言；不碰 `internal/`。

- [ ] T007 建立 9 個新句 stepdef 獨立檔案骨架 `tests/e2e/steps/step_t008_*.go`–`step_t016_*.go`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（8 個新句）
    - `specs/truth/features/cli/history/dsl.md`（1 個新句）
    - `tests/e2e/steps/register.go`
  - 只做：建立 9 個獨立檔案（各對應 T008–T016 的句），各自 `init()` 自我註冊空白 registrar。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的句與本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/{chat,history}/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪 9 個新句：8 個在 `chat/dsl.md`、1 個在 `history/dsl.md`。介面根 `dsl.md` 為 **NOOP**（狀態行不帶 `tellme:` 前綴，frozen class-phrase 詞彙維持 10），本輪不新增根句。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**無**。既有句（含 `the operator starts tellme with the prompt …`、`the captured standard output is exactly …`、`tellme explains on stderr that …`）語意不變。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 9 個新句（`ADD`）。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decisions 1, 2, 4, 5）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 8 個 `chat` 新句
- `specs/truth/features/cli/history/dsl.md` -> 1 個 `history` 新句
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reporting-the-payload-status.feature` + module rows）與 MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`）；NOOP（root `cli/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 `internal/domain/llm/token_test.go`、`internal/config/*_test.go`、`internal/ui/status_test.go`、`internal/infrastructure/llm/openai/client_test.go`，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T008–T016 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T017（UNIT）可並行；T018 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T008 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with "{answer}" and reports its usage`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with "{answer}" and reports its usage`
  - Landing: `tests/e2e/steps/step_t008_chat_given_provider_usage.go`
  - 語意：寫可解析 config 選 `{provider}`；fake 回 `{answer}` 並附 `usage` 物件（scripted prompt-token 數）。

- [ ] T009 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with "{answer}" and reports no usage`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with "{answer}" and reports no usage`
  - Landing: `tests/e2e/steps/step_t009_chat_given_provider_no_usage.go`
  - 語意：寫可解析 config 選 `{provider}`；fake 回 `{answer}` 且**不**附 `usage`。

- [ ] T010 [P] [BDD-RED] `Given: the payload budget is "{budget}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the payload budget is "{budget}"`
  - Landing: `tests/e2e/steps/step_t010_chat_given_payload_budget.go`
  - 語意：把 `MAX_HISTORY_TOKENS` 設為 `{budget}` 進 subprocess 環境。

- [ ] T011 [P] [BDD-RED] `Then: tellme reports the estimated payload status for the turn`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme reports the estimated payload status for the turn`
  - Landing: `tests/e2e/steps/step_t011_chat_then_estimated_status.go`
  - 語意：斷言擷取的 **stderr** 帶有 pre-flight 行（`[HH:MM:SS] Payload: ~<n>/<max> tokens - <mode> - <model>`，`~` 前綴）；且該行**不在** stdout。

- [ ] T012 [P] [BDD-RED] `Then: tellme reports the measured payload status for the turn`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme reports the measured payload status for the turn`
  - Landing: `tests/e2e/steps/step_t012_chat_then_measured_status.go`
  - 語意：斷言 stderr 帶有 post-turn 行（`[HH:MM:SS] Payload: <n>/<max> tokens - <mode> - <model>`，無 `~`）。

- [ ] T013 [P] [BDD-RED] `Then: tellme reports no measured payload status`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme reports no measured payload status`
  - Landing: `tests/e2e/steps/step_t013_chat_then_no_measured_status.go`
  - 語意：斷言 stderr 不含無 `~` 的 payload 行。

- [ ] T014 [P] [BDD-RED] `Then: the payload status measures against a budget of {budget} tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the payload status measures against a budget of {budget} tokens`
  - Landing: `tests/e2e/steps/step_t014_chat_then_budget_value.go`
  - 語意：斷言回報的 payload 狀態行 `<max>` 欄等於 `{budget}`。

- [ ] T015 [P] [BDD-RED] `Then: the payload status names the active mode and model`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the payload status names the active mode and model`
  - Landing: `tests/e2e/steps/step_t015_chat_then_mode_model.go`
  - 語意：斷言回報的 payload 狀態行以 ` - <mode> - <model>` 結尾（本輪有效 mode 與作用 provider 標籤）。

- [ ] T016 [P] [BDD-RED] `Then: no payload status is reported`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `no payload status is reported`
  - Landing: `tests/e2e/steps/step_t016_history_then_no_status.go`
  - 語意：斷言擷取的 stderr 不含任何 payload 狀態行（既無 `~` 估計行，也無 measured 行）。

### UNIT（pure-helper 單元測試）

- [ ] T017 [P] [UNIT] estimator / budget resolver / status formatting / usage parsing 單元測試
  - Read:
    - `specs/plans/009-payload-status-line/research.md` -> Decision 1, 2, 4, 5
    - `specs/truth/techstack.md` -> CLI Application（Token estimator / Payload status line）、Configuration（Payload budget）、Reasoning & Provider Transport（Response normalization）、Testing & Verification（Pure-helper unit tests）
  - Landing: `internal/domain/llm/token_test.go`、`internal/config/*_test.go`、`internal/ui/status_test.go`、`internal/infrastructure/llm/openai/client_test.go`（表驅動）
  - 撰寫：`EstimateTokens` 的決定性（固定 payload → 固定值）；`EffectiveMaxHistoryTokens`（環境覆寫、預設 1000000、非整數/負值 → `ErrInvalidValue`）；status-line formatter（pre-flight `~` 與 post-turn 兩型、clock seam 注入、無 `tellme:` 前綴）；OpenAI adapter 解析 `usage`（有/無 usage 兩型）。

### Phase Review Gate

- [ ] T018 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/reporting-the-payload-status.feature`、`specs/truth/features/cli/history/inspecting-the-session-history.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/history/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 檢驗：無 undefined step；9 個新 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/reporting-the-payload-status.feature

**Goal**: 實作「prompt turn 送出前在 `stderr` 報出 payload 的**估計**大小、完成後報出 provider **實測**大小，皆對照可配置 budget」與「答案流（stdout）不受影響」，使 `reporting-the-payload-status.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reporting-the-payload-status.feature` -> `Feature: Reporting the payload status`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`tellme reports the estimated payload status for the turn`、`tellme reports the measured payload status for the turn`、`tellme reports no measured payload status`、`the payload status measures against a budget of {budget} tokens`、`the payload status names the active mode and model`）+ Given（`… reports its usage` / `… reports no usage` / `the payload budget is "{budget}"`）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reporting-the-payload-status.feature` + module rows）
- `specs/truth/techstack.md` -> CLI Application（Token estimator / Payload status line）、Configuration（Payload budget）、Reasoning & Provider Transport（Provider gateway port / Response normalization）
- `specs/plans/009-payload-status-line/research.md` -> Decision 1, 2, 3, 4, 5, 6
- `internal/domain/llm/token.go`、`internal/ui/status.go`、`internal/cli/cli.go`、`internal/config/config.go`、`internal/infrastructure/llm/openai/client.go`

**Boundary**:
- 只用 stdlib 估算 payload（resumed conversation + 當前 prompt 文字；不含 tool 定義）；pre-flight 行以 `~` 標示估計；post-turn 行的 `<actual>` 來自 provider 回報的 `usage`（缺則整行省略）。
- 兩行皆寫 **`stderr`**；**always-on**（不因非終端而抑制、不因 `-r` 而抑制）；**不帶** `tellme:` 前綴（frozen 詞彙維持 10）；timestamp 由注入的 clock seam 產生。
- `stdout`（答案流）byte 不變；不得引入任何 pruning 或自動摘要（settled exclusion）。
- 不實作 `-l` 無狀態行（屬 4B boundary）。

**Test Scope**:
- `specs/truth/features/cli/chat/reporting-the-payload-status.feature`

- [ ] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 estimator、usage 解析與狀態行輸出（stream 分流）

## Phase 4B: MODIFY Feature File - cli/history/inspecting-the-session-history.feature

**Goal**: 讓 `-l` 離線路徑不輸出 payload 狀態行，使 `inspecting-the-session-history.feature`（含新 Rule *Listing reports no payload status*）全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature` -> `Feature: Inspecting the session history`
- `specs/truth/features/cli/history/dsl.md` -> Then（`no payload status is reported`）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`）
- `specs/plans/009-payload-status-line/research.md` -> Decision 3, 7

**Boundary**:
- `-l N` 為非 prompt 路徑：**不**輸出任何 payload 狀態行；`-l` 既有契約（正整數、clamp、離線短路、只列 prompt/answer）不變。
- 不改 `internal/cli` 既有 `-l` 語意；離線路徑不得觸網。

**Test Scope**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature`

- [ ] T021 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T022 [BDD-REFACTOR] 在綠燈下整理「僅 prompt turn 輸出狀態行」的分流判斷

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（含 prompt-turn 狀態行、離線路徑、exit-code 與 class-phrase 契約）。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T023 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **10**（狀態行不帶 `tellme:` 前綴）；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不觸網且**不**輸出狀態行；prompt turn 的 `stdout` 答案流 byte 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；`MAX_HISTORY_TOKENS` 預設 1000000。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Token estimator：stdlib heuristic，deterministic） | T001、T011、T017、T019、T020 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Payload status line：pre-flight `~est/max` + post-turn `actual/max`，stderr，always-on，no `tellme:` prefix，clock seam，observe-only） | T002、T005、T011–T015、T019、T020、T022 | PASS |
| `specs/truth/techstack.md` -> Configuration（Payload budget：`MAX_HISTORY_TOKENS` env/config，default 1000000，>= 0，negative → config error） | T004、T010、T014、T017、T019、T021、T023 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port `Response` + Usage；Response normalization 解析 `usage`） | T003、T008、T009、T012、T017、T019、T020 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner 斷言 stderr 狀態行 + byte-exact stdout；fake 供應/隱藏 usage；pure-helper 擴充） | T006、T008–T016、T017、T018、T019 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（metrics line `M/H/C/$` 不引入；budget 只顯示不強制） | 負向決策 -> T019 Boundary、T023（不引入 metrics / pruning） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001–T006、T019–T023 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reporting-the-payload-status.feature` + 8 module rows） | T007–T015、T018、T019、T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/inspecting-the-session-history.feature` + `history/dsl.md`） | T007、T016、T018、T021、T022 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T023 驗證詞彙維持 10） | PASS |
| `research.md` -> Decision 1（stdlib heuristic estimator + injectable seam） | T001、T011、T017、T019 | PASS |
| `research.md` -> Decision 2（widen `Response` + parse `usage`；no usage → omit line） | T003、T008、T009、T012、T013、T017、T019 | PASS |
| `research.md` -> Decision 3（stderr + reference-parity format + always-on；non-prompt paths emit none） | T002、T005、T011–T015、T016、T019、T022 | PASS |
| `research.md` -> Decision 4（line format + injectable clock seam） | T002、T015、T017、T019、T023 | PASS |
| `research.md` -> Decision 5（`MAX_HISTORY_TOKENS` default 1000000 env-over-file, >= 0） | T004、T010、T014、T017、T021、T023 | PASS |
| `research.md` -> Decision 6（estimate counts resumed conversation + prompt；excludes tool defs） | T001、T017、T019 | PASS |
| `research.md` -> Decision 7（no persistence — recompute; `history_entry` unchanged） | T023（DataModel NOOP 驗證）、T019 Boundary | PASS |
| `research.md` -> Decision 8（stdlib-only；no new dependency） | T001、T023（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 9（testing unchanged; godog E2E + stdlib unit） | T006、T007、T017、T018、T019 | PASS |
| `spec.md` -> US1–US3、`FR-001`–`FR-014`、`NFR-001`–`NFR-005` | T008–T016（對齊）、T019–T022（交付）、T023 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
