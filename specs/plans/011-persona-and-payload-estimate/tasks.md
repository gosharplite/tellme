# Tasks: Persona on the Wire & Wire-Faithful Payload Estimate (round 011)

**Plan Package**: `specs/plans/011-persona-and-payload-estimate`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**（與 round 010 的 contract+oracle 不同）：`internal/cli` 需把 resolved `PERSON` 帶入並送給 transport；`internal/infrastructure/llm` 於請求體最前面送 `system` persona；`internal/domain/llm` 的估量器輸入放寬為 wire payload。相位順序仍為「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。

---

## Phase 2: Foundational

**Goal**: 建立 persona 注入與 wire-payload 估量的**落點骨架與 seam**（Zero Shared Edits 原則），讓後續 Phase 3 / Phase 4 不各自發明 seam、不各自重開檔。只建立落點與載體，不寫請求體行為、不寫估量行為。

- [X] T001 在 CLI resolution 帶上 resolved `PERSON`
  - Read:
    - `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 2, 3
    - `specs/truth/techstack.md` -> Configuration（Persona instruction）
    - `internal/cli/cli.go`
  - 只做：在 `resolution` 增加 `Person` 欄位，並於 `resolve()` 從已載入的 config 填入；不改任何輸出、不改估量、不改 gateway 建構。
  - 不做：不寫 request 行為；不改 `runTurn` 行為；不碰 truth。

- [X] T002 在 transport seam 加上 persona 載體
  - Read:
    - `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Request assembly）
    - `internal/infrastructure/llm/factory.go`、`internal/infrastructure/llm/openai/client.go`
  - 只做：在 `openai.Config` 增加 `Persona` 欄位，於 `NewGateway`/factory 由傳入值填入；新增 factory → adapter 的參數／來源（gateway-level 注入）。
  - 不做：**不**改 `requestBody` 內容（leading system message 留給 Phase 4A GREEN）；不改 transport 其他行。

- [X] T003 放寬估量器入口為 wire-payload 形狀
  - Read:
    - `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 4, 5
    - `specs/truth/techstack.md` -> CLI Application（Token estimator）
    - `internal/domain/llm/token.go`
  - 只做：新增一個 wire-payload 估量入口（`persona string`、tools 宣告、messages 三段），保留既有 `EstimateTokens(messages)` 作為 message 項；並在需要處暴露 CLI 可取得的 tool 宣告（例如由 registry 投影 tool defs）。
  - 不做：不決定啟發式常數語意之外的行為；不改既有 message-term 結果；不寫 stepdef。

- [X] T004 建立 9 個新句 + 1 個 ALIGN 落點骨架與 E2E persona helper
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪新句與 reworded 句）
    - `tests/e2e/steps/register.go`、`tests/e2e/steps/scenario_context.go`
  - 只做：建立 9 個獨立新句 stepdef 檔案（各自 `init()` 自我註冊），並在 `scenario_context.go` 加一個「把 persona 寫進預設 config」的 helper（測試共用元件）供後續 Given 使用。
  - 不做：不寫具體 step 斷言邏輯；不碰既有 step 檔案；不碰 `internal/`。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 truth-delta 有改的句（`the request carried no earlier exchange`），以及本輪 Feature 用到、尚無 stepdef 的 9 個新句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`（本輪 10 句皆在此）。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（persona／estimate 句為 `chat` 模組專屬；frozen class-phrase 詞彙維持 10）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**1 句** — `the request carried no earlier exchange`（MODIFY）。既有 stepdef（`step_t010_chat_then_request_carried_nothing.go`）斷言 messages 恰為單一 user 訊息；本輪起請求會多一則 leading `system` persona 訊息，依 `dsl.md` 該列改測試以容許之。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **9 句**（`ADD` 的 DSL rows）。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 wire-payload 估量單元測試（research Decision 4、5）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the runtime home holds a configuration whose persona is "{persona}"`
  -> `the runtime home holds a configuration with no persona`
  -> `a previous run with the persona "{persona}" and the prompt "{prompt}" reported an estimated payload status`
  -> `the operator uses the persona "{persona}" and starts tellme with the prompt "{prompt}"`
  -> `the request carried the persona "{persona}"`
  -> `the request carried no persona`
  -> `the estimated payload exceeds the conversation messages alone`
  -> `the estimated payload is larger than the previous run's`
  -> `the estimated payload matches the previous run's`
  -> `the request carried no earlier exchange`（reworded）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/sending-the-configured-persona.feature`、`chat/estimating-the-wire-payload.feature`）+ MODIFY（`chat/dsl.md`）；NOOP（root `cli/dsl.md`）
- `tests/e2e/steps/`、`tests/e2e/harness/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `the request carried the earlier exchange …` stepdef 掃描成對訊息、不要求 exactness，故不需 ALIGN；於 review 時確認其在 leading system message 下仍成立。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T005–T014 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T015（UNIT）可並行；T016 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN（MODIFY 句型）

- [X] T005 [P] [BDD-ALIGN] `Then: the request carried no earlier exchange`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried no earlier exchange`; `tests/e2e/steps/step_t010_chat_then_request_carried_nothing.go`, `step_t010_chat_then_request_piped_content.go`, `step_t011_chat_then_request_instruction_then_content.go`
  - 語意：斷言 recorded request 的 messages 為「當前 user prompt，前面僅可有（當 persona 已設定時的）leading `system` 訊息」；不得接受任何 earlier turn。

### BDD-RED（本輪新增句型）

- [X] T006 [P] [BDD-RED] `Given: the runtime home holds a configuration whose persona is "{persona}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the runtime home holds a configuration whose persona is "{persona}"`
  - Landing: `tests/e2e/steps/step_t006_chat_given_persona.go`
  - 語意：把預設 config 的 `PERSON` 設為 `{persona}`（config 需已由 provider Given 建好）。

- [X] T007 [P] [BDD-RED] `Given: the runtime home holds a configuration with no persona`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the runtime home holds a configuration with no persona`
  - Landing: `tests/e2e/steps/step_t007_chat_given_no_persona.go`
  - 語意：清除預設 config 的 `PERSON`。

- [X] T008 [P] [BDD-RED] `Given: a previous run used the persona "{persona}" and reported an estimated payload status`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a previous run with the persona "{persona}" and the prompt "{prompt}" reported an estimated payload status`
  - Landing: `tests/e2e/steps/step_t008_chat_given_previous_run.go`
  - 語意：**對 runtime home 的拷貝**設定 persona、跑一次、解析 pre-flight `~<n>` 並存入 scenario；真 home（history/config）不受影響。

- [X] T009 [P] [BDD-RED] `When: the operator uses the persona "{persona}" and starts tellme with the prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator uses the persona "{persona}" and starts tellme with the prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t009_chat_when_persona_run.go`
  - 語意：設 persona 後跑 `tellme "{prompt}"`（單一 Act）。

- [X] T010 [P] [BDD-RED] `Then: the request carried the persona "{persona}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried the persona "{persona}"`
  - Landing: `tests/e2e/steps/step_t010_chat_then_request_persona.go`
  - 語意：recorded request 的 messages 首則為 `system`-role，內容等於 `{persona}`。

- [X] T011 [P] [BDD-RED] `Then: the request carried no persona`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried no persona`
  - Landing: `tests/e2e/steps/step_t011_chat_then_request_no_persona.go`
  - 語意：recorded request 的 messages 不含任何 `system`-role 訊息。

- [X] T012 [P] [BDD-RED] `Then: the estimated payload exceeds the conversation messages alone`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload exceeds the conversation messages alone`
  - Landing: `tests/e2e/steps/step_t012_chat_then_estimate_exceeds.go`
  - 語意：pre-flight `~<n>` 嚴格大於估量器套用在 recorded request 的 messages（去除 leading `system`）上的值。

- [X] T013 [P] [BDD-RED] `Then: the estimated payload is larger than the previous run's`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload is larger than the previous run's`
  - Landing: `tests/e2e/steps/step_t013_chat_then_estimate_larger.go`
  - 語意：當前 `~<n>` 嚴格大於 scenario 儲存的 previous estimate。

- [X] T014 [P] [BDD-RED] `Then: the estimated payload matches the previous run's`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload matches the previous run's`
  - Landing: `tests/e2e/steps/step_t014_chat_then_estimate_matches.go`
  - 語意：當前 `~<n>` 等於 scenario 儲存的 previous estimate。

### UNIT（非 DSL 的單元斷言）

- [X] T015 [P] [UNIT] wire-payload 估量單元測試
  - Read:
    - `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 4, 5
    - `specs/truth/techstack.md` -> CLI Application（Token estimator）+ Testing & Verification（Pure-helper unit tests）
    - `internal/domain/llm/token.go`
  - Landing: `internal/domain/llm/token_test.go`
  - 撰寫：估量器對「persona + tool 宣告 + messages」輸入為確定性、可回應（payload 增加則估計增加）、且 > 僅 messages 的估計；涵蓋 persona 為空與非空。

### Phase Review Gate

- [X] T016 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/sending-the-configured-persona.feature`、`specs/truth/features/cli/chat/estimating-the-wire-payload.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/cmd_helper.go`
    - `internal/cli/cli.go`、`internal/domain/llm/token.go`、`internal/infrastructure/llm/factory.go`、`internal/infrastructure/llm/openai/client.go`
  - 檢驗：無 undefined step；10 個新／改 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之產品行為；`the request carried the earlier exchange …` stepdef 在 leading `system` 訊息下仍成立。

---

## Phase 4A: ADD Feature File - cli/chat/sending-the-configured-persona.feature

**Goal**: 讓 `sending-the-configured-persona.feature` 全綠 — 實作把 resolved `PERSON` 作為每個請求的 leading `system` 訊息送出。

**Shared Must Read**:
- `specs/truth/features/cli/chat/sending-the-configured-persona.feature` -> `Feature: Sending the configured persona`
- `specs/truth/features/cli/chat/dsl.md` -> `the request carried the persona "…"`、`the request carried no persona`、`the runtime home holds a configuration whose persona is "…"`、`the runtime home holds a configuration with no persona`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/sending-the-configured-persona.feature`）+ MODIFY（`chat/dsl.md`）；`/axb-technical-research` MODIFY（`techstack.md` Request assembly / Persona instruction）
- `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 1, 2, 3（persona shape／transport 注入／空 persona）

**Boundary**:
- 產品碼：`openai.requestBody`（或等效組裝點）在 `Persona` 非空時，於 messages 最前面插入 `{"role":"system","content":Persona}`；`Persona` 由 T001/T002 的 seam 帶入。空 persona 不插。
- 只改 persona 組裝與其 seam；**不**改估量（Phase 4B）與 `stdout` 契約。

**Test Scope**:
- `specs/truth/features/cli/chat/sending-the-configured-persona.feature`

- [X] T017 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T018 [BDD-REFACTOR] 在綠燈下整理 persona 注入組裝與 seam

## Phase 4B: ADD Feature File - cli/chat/estimating-the-wire-payload.feature

**Goal**: 讓 `estimating-the-wire-payload.feature` 全綠 — 讓 pre-flight 估量覆蓋 wire payload（persona + tool 宣告 + messages）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/estimating-the-wire-payload.feature` -> `Feature: Estimating the wire payload`
- `specs/truth/features/cli/chat/dsl.md` -> `the estimated payload exceeds the conversation messages alone`、`the estimated payload is larger than the previous run's`、`the estimated payload matches the previous run's`、`a previous run used the persona "…" and reported an estimated payload status`、`the operator uses the persona "…" and starts tellme with the prompt "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/estimating-the-wire-payload.feature`）；`/axb-technical-research` MODIFY（`techstack.md` Token estimator / Payload status line）
- `specs/plans/011-persona-and-payload-estimate/research.md` -> Decision 4, 5（估量輸入 = persona + tool 宣告 + messages；釘輸入與確定性，不釘數值相等）

**Boundary**:
- 產品碼：`internal/cli` 的 pre-flight 改呼叫 wire-payload 估量入口（T003），傳入 resolved `PERSON`、當前 registry 的 tool 宣告、與 messages。估量仍為 dependency-free 啟發式。
- 只改估量輸入與呼叫點；**不**改 post-turn measured 行（仍以 provider 的 `usage.prompt_tokens`）、不改狀態行格式。

**Test Scope**:
- `specs/truth/features/cli/chat/estimating-the-wire-payload.feature`

- [X] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T020 [BDD-REFACTOR] 在綠燈下整理估量輸入組裝與 tool-defs 投影

## Phase 4C: MODIFY Feature File - cli/chat/remembering-the-conversation.feature

**Goal**: 確認 `remembering-the-conversation.feature` 在 leading persona `system` 訊息下全綠（`the request carried no earlier exchange` 已於 T005 ALIGN），無產品碼變更。

**Shared Must Read**:
- `specs/truth/features/cli/chat/remembering-the-conversation.feature` -> `Feature: Remembering the conversation across runs`
- `specs/truth/features/cli/chat/dsl.md` -> `the request carried no earlier exchange`（reworded）、`the request carried the earlier exchange "…" and "…"`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` reworded row）

**Boundary**:
- 本 phase **無產品碼變更**：persona 由 Phase 4A 送入；此處僅確認 reworded assertion 全綠。紅燈只修測試層。

**Test Scope**:
- `specs/truth/features/cli/chat/remembering-the-conversation.feature`

- [X] T021 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T022 [BDD-REFACTOR] 在綠燈下整理該模組受影響的測試落點

## Phase 4D: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（persona 契約、wire-payload 估量、byte-exact `stdout`、exit-code 與 class-phrase 契約），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T023 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**：暫時讓 persona 不送入（或估量退回僅 messages），確認相關 E2E/unit 兩層都失敗；觀察到失敗即還原。
  - 確認：exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **10**；prompt turn 的 `stdout` byte 不變；post-turn measured 行仍以 provider usage 為源；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Configuration（Persona instruction） | T001、T002、T017 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Request assembly：leading system message） | T002、T017、T018 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Token estimator：wire payload inputs） | T003、T015、T019、T020 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Payload status line：pre-flight over wire payload） | T003、T019、T020 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner / Pure-helper unit tests） | T006–T014、T015、T023 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001–T003、T017、T019、T023 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/sending-the-configured-persona.feature`） | T006–T011、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/estimating-the-wire-payload.feature`） | T012–T015、T019、T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md`，含 reworded row） | T005、T006–T014、T016、T021 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T023 驗證詞彙維持 10） | PASS |
| `research.md` -> Decision 1（persona = leading `system` message） | T002、T017、T023 | PASS |
| `research.md` -> Decision 2（transport-level injection — every request） | T002、T017、T018 | PASS |
| `research.md` -> Decision 3（sourced from resolved PERSON；empty ⇒ none） | T001、T006、T007、T017 | PASS |
| `research.md` -> Decision 4（estimator counts persona + tool declarations + messages） | T003、T015、T019 | PASS |
| `research.md` -> Decision 5（pin inputs + determinism; no numeric equality） | T003、T012–T015、T019、T023 | PASS |
| `research.md` -> Decision 6（reuse fake; assert persona + estimate） | T008–T014、T016、T023 | PASS |
| `research.md` -> Decision 7（no new dependency；BDD techstack unchanged） | T023（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 8（must-ask questions settled） | T017–T023（不改 runner/端） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-013`、`NFR-001`–`NFR-005` | T006–T015（對齊）、T017–T022（交付）、T023 | PASS |
| `plan.md` -> Source-code structure（`internal/cli`、`internal/domain/llm/token.go`、`internal/infrastructure/llm/{factory,openai/client}.go`、E2E steps） | T001–T003、T006–T014、T017–T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` reworded `the request carried no earlier exchange` | T005、T016、T021 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
