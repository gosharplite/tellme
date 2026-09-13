# Tasks: Stream-Ordering Observability (round 010)

**Plan Package**: `specs/plans/010-stream-ordering-observability`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only — `os/exec`；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪為 contract + oracle（無產品碼變更）**：`internal/cli` 的 emit order 已於 `7bcb2d3` 正確；round 010 只新增 **測試層見證**（合併流擷取）與 **可執行的 ordering 契約**（CLI truth 已於 `/axb-dsl-refine` pin 入），故 Feature phase 的 `[BDD-GREEN]` 為「以見證確認全綠」，`[BDD-REFACTOR]` 為測試層整理。

---

## Phase 2: Foundational

**Goal**: 建立 E2E **合併流擷取見證**（merged-stream witness）與 3 個新句 stepdef 落點骨架（Zero Shared Edits 原則），讓後續 Phase 3 不各自發明擷取方式、不各自重開檔。不寫 DSL 語意、不寫產品行為。

- [ ] T001 在 E2E harness 新增 merged-stream 擷取變體
  - Read:
    - `specs/plans/010-stream-ordering-observability/research.md` -> Decision 1, 6, 7
    - `specs/truth/techstack.md` -> Testing & Verification（Cross-stream ordering witness / E2E runner）
    - `tests/e2e/harness/cmd_helper.go`
  - 只做：新增一個 run 變體，把 child 的 `stdout` 與 `stderr` 指向**同一個** `*bytes.Buffer`（comparable 指標，`os/exec` 對同一 writer 序列化寫入 → 保序），並回傳**合併後**輸出；保留既有分開擷取路徑不動。
  - 不做：不改既有 `RunResult{Stdout,Stderr}` 語意；不引入新相依（stdlib `os/exec`）；不寫 step 斷言。

- [ ] T002 scenario context 走 merged-run 並暴露合併輸出
  - Read:
    - `specs/plans/010-stream-ordering-observability/research.md` -> Decision 1, 4, 6
    - `specs/truth/techstack.md` -> Testing & Verification（Cross-stream ordering witness）
    - `tests/e2e/steps/scenario_context.go`
  - 只做：新增一個以**合併流**執行當前 arranged command 的入口，並在 scenario context 暴露合併輸出欄位供 ordering Then 讀取；維持既有 `run()` 與 hermeticity（`envUnset`）不變。
  - 不做：不改既有 step 行為；不碰 `internal/`。

- [ ] T003 建立 3 個新句 stepdef 獨立檔案骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（3 個新 Then 句）
    - `tests/e2e/steps/register.go`
  - 只做：建立 3 個獨立檔案（`step_t004_*`、`step_t005_*`、`step_t006_*`），各自 `init()` 自我註冊空白 registrar。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 Feature 用到、尚無 stepdef 的 3 個 ordering 句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`（本輪 3 句皆在此）。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（ordering 句為 `chat` 模組專屬；frozen class-phrase 詞彙維持 10）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Then 讀 `必查`（`呈現結果`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**無**。既有句（`tellme reports the estimated payload status for the turn`、`tellme reports the measured payload status for the turn`、`the run reported the tool call "…" on its diagnostic output` 等）語意不變；本輪兩個特徵檔僅**新增**步驟。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 3 個新句（`ADD` 的 DSL rows）。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 `runTurn` emit-order 單元斷言（research Decision 4）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the estimated payload status is reported before the answer`
  -> `the measured payload status is reported after the answer`
  -> `the tool activity is reported before the answer`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/reporting-the-payload-status.feature`、`chat/watching-the-tool-loop.feature`）+ ADD（`chat/dsl.md` 3 rows）; NOOP（root `cli/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/harness/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- ordering Then **以合併流**（T001/T002 見證）斷言位置；不得靠分開擷取猜順序。
- `[UNIT]` 落入 `internal/cli/turn_test.go`，採 stdlib `testing`（擴充既有 `TestRunTurn_PostTurnStatusFollowsAnswer`）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T006 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T007（UNIT）可並行；T008 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T004 [P] [BDD-RED] `Then: the estimated payload status is reported before the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload status is reported before the answer`
  - Landing: `tests/e2e/steps/step_t004_chat_then_estimated_before_answer.go`
  - 語意：在**合併流**擷取中，斷言 pre-flight estimated（`~`）行出現在答案位元組**之前**；答案在 `stdout`。

- [ ] T005 [P] [BDD-RED] `Then: the measured payload status is reported after the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the measured payload status is reported after the answer`
  - Landing: `tests/e2e/steps/step_t005_chat_then_measured_after_answer.go`
  - 語意：合併流中，measured 行出現在答案位元組**之後**。

- [ ] T006 [P] [BDD-RED] `Then: the tool activity is reported before the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the tool activity is reported before the answer`
  - Landing: `tests/e2e/steps/step_t006_chat_then_tool_before_answer.go`
  - 語意：合併流中，tool-loop log 行出現在答案位元組**之前**。

### UNIT（非 DSL 的單元斷言）

- [ ] T007 [P] [UNIT] `runTurn` emit-order 單元斷言（擴充既有 emit-order 測試）
  - Read:
    - `specs/plans/010-stream-ordering-observability/research.md` -> Decision 4
    - `specs/truth/techstack.md` -> Testing & Verification（Cross-stream ordering witness / Pure-helper unit tests）
    - `internal/cli/turn_test.go`
  - Landing: `internal/cli/turn_test.go`
  - 撰寫：以單一合併 buffer 斷言 `pre-flight < answer < measured`（擴充 `TestRunTurn_PostTurnStatusFollowsAnswer`）；為 tool-loop 場景加對應的 emit-order 斷言。

### Phase Review Gate

- [ ] T008 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/reporting-the-payload-status.feature`、`specs/truth/features/cli/chat/watching-the-tool-loop.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/cmd_helper.go`
    - `internal/cli/turn_test.go`
  - 檢驗：無 undefined step；3 個新 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之斷言；**合併見證確實記錄跨流順序**（分開擷取不得用來猜順序）。

---

## Phase 4A: MODIFY Feature File - cli/chat/reporting-the-payload-status.feature

**Goal**: 確認 `reporting-the-payload-status.feature`（含新增 ordering Then）在合併流見證下全綠，並在綠燈保護下整理測試層落點（無產品碼變更）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reporting-the-payload-status.feature` -> `Feature: Reporting the payload status`
- `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload status is reported before the answer`、`the measured payload status is reported after the answer` + 既有 payload-status rows
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/reporting-the-payload-status.feature`）+ ADD（`chat/dsl.md`）
- `specs/plans/010-stream-ordering-observability/research.md` -> Decision 4, 5
- `tests/e2e/steps/`、`tests/e2e/harness/`

**Boundary**:
- 本輪**無產品碼變更**（research Decision 5；`internal/cli` 的 emit order 已於 `7bcb2d3` 正確）。`[BDD-GREEN]` 以合併流見證執行 Test Scope 並確認全綠；若因見證／stepdef 落點導致紅燈，只修正測試層，不改產品。
- `stdout` 保持 byte-exact；兩條狀態行皆在 `stderr`。

**Test Scope**:
- `specs/truth/features/cli/chat/reporting-the-payload-status.feature`

- [ ] T009 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T010 [BDD-REFACTOR] 在綠燈下整理 ordering stepdef 與合併見證 helper

## Phase 4B: MODIFY Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: 確認 `watching-the-tool-loop.feature`（含新增 tool-ordering Then）在合併流見證下全綠，並在綠燈保護下整理測試層落點（無產品碼變更）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature` -> `Feature: Watching the tool loop work`
- `specs/truth/features/cli/chat/dsl.md` -> `the tool activity is reported before the answer` + 既有 tool-loop rows（`the run reported the tool call "…" on its diagnostic output`）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/watching-the-tool-loop.feature`）
- `specs/plans/010-stream-ordering-observability/research.md` -> Decision 3, 5

**Boundary**:
- 本輪**無產品碼變更**；tool-loop log 已於 round 008 在 `stderr` 於答案前輸出。`[BDD-GREEN]` 以合併流見證執行 Test Scope 並確認全綠；紅燈只修測試層。
- tool-loop log 維持在 `stderr`；答案流不變。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [ ] T011 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T012 [BDD-REFACTOR] 在綠燈下整理 tool-ordering stepdef 與共用見證 helper

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（含 ordering 契約、byte-exact `stdout`、exit-code 與 class-phrase 契約），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T013 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證（SC-003）**：暫時反轉 emit order（或暫時改壞 ordering 斷言），確認 **unit 與 E2E 兩層都失敗**；觀察到失敗即還原。
  - 確認：exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **10**；prompt turn 的 `stdout` byte 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變且不輸出 ordering 行；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Testing & Verification（Cross-stream ordering witness：merged capture，stdlib `os/exec`） | T001、T002、T004–T006、T008、T010、T012 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner：merged-stream capture variant） | T001、T002、T008、T009–T012 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Test strategy：layered unit + E2E ordering assertions） | T004–T007、T008、T013 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests：turn emit-order assertion） | T007、T013 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001、T002、T007、T013 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/reporting-the-payload-status.feature`） | T004、T005、T009、T010 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/watching-the-tool-loop.feature`） | T006、T011、T012 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/dsl.md` 3 ordering rows） | T003–T006、T008 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T013 驗證詞彙維持 10） | PASS |
| `research.md` -> Decision 1（E2E witness = merged single-buffer capture） | T001、T002、T004–T006、T008 | PASS |
| `research.md` -> Decision 2（ordering pinned as executable interface truth） | T004–T006（對齊已 pin 入的 DSL rows）、T013 | PASS |
| `research.md` -> Decision 3（required = payload-status + tool-loop；degrade warning incidental） | T004–T006、T008、T013 | PASS |
| `research.md` -> Decision 4（assert at both layers — unit + E2E） | T004–T007、T013 | PASS |
| `research.md` -> Decision 5（no product change — contract + oracle） | T009、T011 Boundary；T013 | PASS |
| `research.md` -> Decision 6（deterministic — no sleeps; shared-buffer serialization；hermetic） | T001、T002、T007、T013 | PASS |
| `research.md` -> Decision 7（stdlib-only witness；no new dependency） | T001、T013（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 8（testing/BDD techstack unchanged） | T001–T013（不改 runner/策略） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-011`、`NFR-001`–`NFR-005` | T004–T006（對齊）、T009–T012（交付）、T013 | PASS |
| `plan.md` -> Source-code structure（test-side only：harness merged capture；scenario context；new step files；`internal/cli/turn_test.go`） | T001、T002、T003、T007 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` 之 ordering 契約（round 009 僅 pin presence） | T004–T006、T008 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
