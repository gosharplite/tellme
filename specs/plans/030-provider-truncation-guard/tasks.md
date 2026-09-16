# Tasks: 030 — Provider-transport truncation guard (issue #62)

**Plan Package**: `specs/plans/030-provider-truncation-guard`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/**`
> `specs/truth/contracts/**` (`/axb-api-plan` NOOP) and `specs/truth/data/**` (`/axb-data-plan` NOOP) are unchanged this round; there is no `ui/**` (the guard is transport-facing, with no UX surface change).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍決 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋，不得遺留孤立產物（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 只做本輪新增技術的基礎建設、技術環境與最後的 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架。
- Phase 3 `Test Alignment & Implementation` 的目的：在寫產品碼之前，先把本輪所有受影響 DSL 的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

## Phase 1: Setup

**Omitted** — this round introduces **no new technology**: the guard reads the finish reason with stdlib `encoding/json` only, and it slots into the existing `domain/llm` `Gateway` and the existing `*llm.ProviderError`; `go.mod`/`go.sum` stay unchanged (`research.md` Decision 7). There is nothing to install and no smoke-test to add.

## Phase 2: Foundational

**Goal**: 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。不寫 Phase 3 測試語意，也不寫 Feature Green。

- [ ] T001 建立 Phase 3 stepdef 獨立落點骨架（Zero Shared Edits 原則）
  - Read:
    - `tests/e2e/steps/` -> 既有 `step_r0NN_tNNN_*.go` 獨立檔＋`init()` 自我註冊慣例（見 `step_r029_t012_chat_then_file_created.go`）
    - `specs/truth/features/cli/chat/dsl.md` -> 本輪新增的句列（見 Phase 3 `DSL 參照`）
  - 只做：在 `tests/e2e/steps/` 下，為 Phase 3 每一句建立**獨立**檔案骨架（`step_r030_t004…t009_…`，`init()` 註冊、body 待填），使 Phase 3 並行任務目標檔案互斥。
  - 不做：不寫任何 `[BDD-RED]` 語意，不寫 fake 腳本或產品邏輯。

- [ ] T002 擴充 fake provider 的 truncation 腳本接點
  - Read:
    - `tests/e2e/fakeprovider/fakeprovider.go` -> `Reply`、`scriptedBody`、`answerBody`、`toolCallBody`、`vertexAnswerBody`、`vertexToolCallBody`
    - `specs/plans/030-provider-truncation-guard/research.md` -> `Decision 1`（finish reason 在 wire 上）
  - 只做：為 `Reply` 增加一個 **finish-reason** 欄位，並讓 body builders 在該欄位非空時輸出 `"finish_reason":"<value>"`（OpenAI-compatible `choices[0]`）或 `"finishReason":"<value>"`（Vertex `candidates[0]`）——一個**共用**的truncation 腳本接點，供 Phase 3 的 Given 使用（`"length"` / `"MAX_TOKENS"`）。
  - 不做：不寫任一 Given 的腳本內容，不碰 stepdef，不改產品碼。

- [ ] T003 建立兩個 adapter 的 finish-reason guard 單元測試落點骨架
  - Read:
    - `specs/plans/030-provider-truncation-guard/research.md` -> `Decision 1`, `Decision 6`, `Decision 7`
    - `internal/infrastructure/llm/openai/client.go` -> `parseResponse`
    - `internal/infrastructure/llm/gemini/client.go` -> `parseResponse`
  - 只做：建立單元落點骨架 `internal/infrastructure/llm/openai/truncation_test.go` 與 `internal/infrastructure/llm/gemini/truncation_test.go`（測試函式殼：truncated → error；healthy `stop`/`tool_calls`/`STOP`/absent → ok；空-body truncation 的 precedence）。
  - 不做：不寫斷言（留 Phase 3 `[UNIT]`），不寫產品碼。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 truth-delta 有改的句，以及本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪新增各句都在同模組 `specs/truth/features/cli/chat/dsl.md`（`## Given (round 030)` / `## Then (round 030)`）。本輪用到的介面根句（`the operator has a runnable tellme installation`、`the runtime home is "..."`、`tellme refuses to proceed`、`tellme explains on stderr that "..."`）與既有 chat 句（`the operator starts tellme with the prompt "..."`、`the working directory contains no file "..."`、`the working directory contains a file "..." whose lines are:`、`the lines of "..." are still:`、`tellme exits with the provider error code`）**早已有 stepdef**，不列。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，但語意是舊 truth。依 `dsl.md` 該列改測試，讓它表達最新版 `StepDef 實作語意`。
- `[BDD-REMOVE]`：`DELETE`。此句已不是 truth。移除或改寫仍綁這句的 stepdef / assertion，不得留下保護舊行為的測試。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- 三個 marker 都只動測試層，不寫產品碼。（本輪 6 句全為 `ADD` → 全 `[BDD-RED]`；無 `[BDD-ALIGN]`／`[BDD-REMOVE]`。）

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `## Given (round 030)`：4 條（provider cut off while **creating** a file／while **editing** a file／before finishing its answer／a **Gemini** provider before finishing its answer）
  -> `## Then (round 030)`：2 條（`tellme creates no file "{name}"`／`tellme prints no answer`）
- `truth-delta.md` -> `/axb-dsl-refine` ADD `refusing-a-cut-off-reply.feature` + MODIFY `chat/dsl.md`
- `tests/e2e/fakeprovider/fakeprovider.go` -> the truncation scripting seam（T002）

**Boundary**:
- 一條 DSL 一個 task。
- 只改該句的 stepdef / assertion / 直接依賴的 helper（含 T002 的 fake 接點）。
- 落點檔案採獨立檔案（Zero Shared Edits 原則，SHOULD），使並行派出具備互斥寫入目標。
- 不寫產品碼。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。有 issues 就修正再 review，直到沒有任何問題。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T010 各派一個獨立 subagent，一次整批並行 dispatch（目標檔獨立，符合 Zero Shared Edits）；T011 等全部回來再啟動 subagent review。

- [ ] T004 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint is cut off at the output limit while creating the file "{path}" with the content "{content}"`
  - Read: `tests/e2e/steps/step_r030_t004_chat_given_provider_cutoff_create.go`（+ `tests/e2e/fakeprovider/fakeprovider.go` 的 T002 接點）
  - 語意：script the fake to return a single response whose OpenAI-compatible `finish_reason` is `"length"` and whose message carries a `write_file` tool call (`filepath`={path}, `content`={content}, `reason` set) — a reply cut off mid-tool-call, no final answer.
- [ ] T005 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint is cut off at the output limit while editing the file "{path}" replacing "{old}" with "{new}"`
  - Read: `tests/e2e/steps/step_r030_t005_chat_given_provider_cutoff_edit.go`（+ fake 接點）
  - 語意：同上，tool call 改為 `replace_text`（`filepath`={path}, `old_text`={old}, `new_text`={new}, `reason` set），`finish_reason` = `"length"`。
- [ ] T006 [P] [BDD-RED] `a configured provider "{provider}" whose endpoint is cut off at the output limit before finishing its answer`
  - Read: `tests/e2e/steps/step_r030_t006_chat_given_provider_cutoff_answer.go`（+ fake 接點）
  - 語意：script the fake to return a single response whose `finish_reason` is `"length"` carrying partial text and **no** tool call.
- [ ] T007 [P] [BDD-RED] `a configured Gemini provider "{provider}" whose endpoint is cut off at the output limit before finishing its answer`
  - Read: `tests/e2e/steps/step_r030_t007_chat_given_gemini_cutoff_answer.go`（+ fake 接點）
  - 語意：arrange a `gemini` Vertex-shaped provider（key file 等，照既有 Gemini Given），script the fake for a single Vertex `:generateContent` response whose `candidates[0].finishReason` is `"MAX_TOKENS"` carrying partial text (no `functionCall`).
- [ ] T008 [P] [BDD-RED] `tellme creates no file "{name}"`
  - Read: `tests/e2e/steps/step_r030_t008_chat_then_no_file.go`
  - 語意（必查 權威狀態）：the file `{name}` does **not** exist on disk after the run.
- [ ] T009 [P] [BDD-RED] `tellme prints no answer`
  - Read: `tests/e2e/steps/step_r030_t009_chat_then_no_answer.go`
  - 語意（必查 呈現結果）：the captured standard output is **empty** — the failed turn emitted no answer bytes.
- [ ] T010 [P] [UNIT] provider finish-reason guard 單元測試
  - Read: `internal/infrastructure/llm/openai/client.go` (`parseResponse`), `internal/infrastructure/llm/gemini/client.go` (`parseResponse`), `internal/infrastructure/llm/openai/truncation_test.go`, `internal/infrastructure/llm/gemini/truncation_test.go`, `research.md` -> `Decision 1`, `Decision 2`, `Decision 6`
  - 必查：OpenAI-compatible adapter 對 `choices[0].finish_reason == "length"` 回 **error**（`*llm.ProviderError`），Vertex/Gemini adapter 對 `candidates[0].finishReason == "MAX_TOKENS"` 回 **error**；trigger 為 **universal**（tool call 或純 text 皆然，`Decision 2`）；healthy 值（`stop`/`tool_calls`/`STOP`/absent）**不**觸發；truncation error 對空 body 的 precedence 勝過 `no usable answer`（`Decision 6`）；Gemini 訊息在 `functionCall` 在場時點名該 tool（function-call-aware）。
- [ ] T011 subagent review (phase quality gate)

## Phase 4A: ADD Feature File - cli/chat/refusing-a-cut-off-reply.feature

**Goal**: 以最小 finish-reason guard 產品邏輯（兩個 adapter 於 decode 讀 finish reason，遇到 output-cap truncation 回 `*llm.ProviderError`）讓此 feature 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/refusing-a-cut-off-reply.feature` -> `Feature: Refusing a reply cut off at the output limit`
- `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 030)` 4 條 + `## Then (round 030)` 2 條
- `truth-delta.md` -> `/axb-dsl-refine` ADD `refusing-a-cut-off-reply.feature`
- `specs/truth/techstack.md` -> `Provider output-cap truncation guard` row
- `specs/plans/030-provider-truncation-guard/research.md` -> `Decision 1`, `Decision 2`, `Decision 3`, `Decision 4`, `Decision 5`, `Decision 6`

**Boundary**:
- 只處理兩個 adapter 的 finish-reason guard；**不改** request side（`MAX_TOKENS` 未設時仍不送 cap — `Decision 5`）；不改 CLI dispatch、flag、exit code、`stdout` 契約；**不加** retry/failover/classify 層（`Decision 4`）。
- 產品落點（in-place edit，**無新檔**）：`internal/infrastructure/llm/openai/client.go` 的 `parseResponse` 讀 `choices[0].finish_reason`；`internal/infrastructure/llm/gemini/client.go` 的 `parseResponse` 讀 `candidates[0].finishReason`。truncation error 以 `*llm.ProviderError` 回傳（沿用既有型別 → CLI 的 frozen `the provider request failed` + exit 6），並**優先於**既有 `no usable answer` error。
- `[BDD-GREEN]` 讓 `refusing-a-cut-off-reply.feature` 全綠後，方得 `[BDD-REFACTOR]`。

**Test Scope**:
- `specs/truth/features/cli/chat/refusing-a-cut-off-reply.feature`

- [ ] T012 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T013 [BDD-REFACTOR] 在綠燈下整理兩個 adapter 的 truncation 判斷（各自抽出本地判斷函式＋具名常數，收斂訊息），行為不變

## Phase 4C: REGRESSION

**Goal**: 全 CLI truth feature 回歸 + falsifiability witnesses + `make verify` + 拓樸稽核。（本 phase 一併覆蓋 `/axb-dsl-refine` 對 `reporting-a-failed-provider-request.feature` 的 **doc-only** header MODIFY — 其步驟未變，僅需回歸確認仍為綠。）

**Shared Must Read**:
- `specs/truth/features/cli/**`（全）
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` -> `Feature: Reporting a failed provider request`（header 已加 truncation cross-reference）
- `specs/plans/030-provider-truncation-guard/research.md` -> `Decision 7`（witnesses）

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses：(a) 移除 truncation guard → cut-off file Example 失敗（檔案被建立）或 cut-off answer Example 失敗（答案被印出）；(b) 只對 tool-call truncation 觸發 → cut-off **answer** Example 失敗（universal 被破壞）；(c) 對 healthy finish reason 也觸發 → 既有 provider/answer 場景失敗。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T014 [REGRESSION] 跑全 CLI feature + 見證 + `make verify` + 拓樸稽核

---

## Pre-Delivery Orphan Coverage Sweep

**`truth-delta.md` 非 NOOP rows → task 對應**
- `/axb-technical-research` MODIFY `techstack.md`（Provider output-cap truncation guard row + adapter/normalization/port rows + write-tools row 更新）→ Phase 4A 的 `Read`（`specs/truth/techstack.md -> Provider output-cap truncation guard`）；實作落點由 T012 交付。
- `/axb-dsl-refine` ADD `refusing-a-cut-off-reply.feature` → Phase 4A。MODIFY `chat/dsl.md`（4 Given + 2 Then + module note）→ T004–T009（RED）、T010（UNIT）；Phase 4A。MODIFY `reporting-a-failed-provider-request.feature`（**doc-only** header cross-reference）→ Phase 4C 回歸（其步驟未變）。
- `/axb-api-plan` NOOP、`/axb-data-plan` NOOP → 免建立任務（審計豁免）。

**`research.md` 已拍決 Decisions → task `Read` 覆蓋**
- D1 → T002/T004–T007/T010/Phase 4A；D2 → T004–T009/T010/Phase 4A；D3 → T008/T009/Phase 4A（frozen phrase + exit 6）；D4 → T010/Phase 4A（no retry layer）；D5 → Phase 4A `Boundary`（request side unchanged）；D6 → T010/Phase 4A（precedence）；D7 → T002/T010/T014（hermetic、both families、witnesses）。Truth impact（techstack / dsl-refine / api+data NOOP）→ 對應 `truth-delta.md` rows。

**`specs/truth/techstack.md` 異動章節 → 建置/驗證 task 覆蓋**
- 本輪**無新增技術**（stdlib-only），故無 Setup 建置/版本/make target 變更；相關章節（Provider output-cap truncation guard row）由 Phase 4A 的 `Read` 引用並由 T012 承接。

**孤立產物件數：0。掃描通過，准予交付。**
