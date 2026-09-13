# Tasks: Session-Replay Fidelity — Persist the Tool-Call Signature (round 014)

**Plan Package**: `specs/plans/014-session-replay-fidelity`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod`／`go.sum` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/domain/history` 的 `Step` 新增 `Signature`；`internal/agent/agentloop.go` 於記錄與 `BuildMessages` 重播時帶入它。相位順序仍為「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。

---

## Phase 2: Foundational

**Goal**: 建立本輪 stepdef／`[UNIT]` 的落點骨架與一個資料形狀骨架（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔案或 seam。只建立落點與載體，不寫記錄／重播行為。

- [X] T001 建立 4 個新句 stepdef 獨立檔骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 4 個新句）
    - `tests/e2e/steps/register.go`
  - 只做：建立 4 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），對到 T003–T006 四句。
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [X] T002 建立 `Step.Signature` 資料形狀骨架與 `[UNIT]` 落點
  - Read:
    - `specs/plans/014-session-replay-fidelity/research.md` -> Decision 1, 3
    - `specs/truth/data/data-model.dbml` -> `history_step`
    - `internal/domain/history/history.go`、`internal/infrastructure/history/file_store_widened_test.go`、`internal/agent/agentloop_test.go`
  - 只做：在 `internal/domain/history/history.go` 的 `Step` 加上 `Signature string \`json:"signature,omitempty"\`` 欄位（型別骨架，讓測試層可編譯）；在 `file_store_widened_test.go` 與 `agentloop_test.go` 預留 `[UNIT]` 測試落點（空殼）。
  - 不做：不在 `AgentLoop` 記錄或重播 signature；不寫 store 行為；不寫斷言。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 4 句皆為全新句（`ADD`）。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪 4 句皆 `chat` 模組專屬；frozen class-phrase 詞彙維持 **11**）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **4 句**（`ADD` 的 DSL rows）。
- `[UNIT]`：`Step` signature 的 JSON round-trip + `BuildMessages` 重播輸出 signature（非 DSL）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the session history already holds a tool-using exchange carrying the provider token "{token}"`
  -> `the session history already holds a tool-using exchange with no provider token`
  -> `the request replayed the earlier tool step "{tool}" carrying the provider token "{token}"`
  -> `the request replayed the earlier tool step "{tool}"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/replaying-a-tool-using-conversation.feature`）+ MODIFY（`chat/dsl.md`）；NOOP（root `cli/dsl.md`、`contracts/**`）
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/infrastructure/history/file_store_widened_test.go`、`internal/agent/agentloop_test.go`；採 stdlib `testing`。
- 不寫產品碼（除 T002 已預留的型別欄位）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T008 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T009 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [X] T003 [P] [BDD-RED] `Given: the session history already holds a tool-using exchange carrying the provider token "{token}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the session history already holds a tool-using exchange carrying the provider token "{token}"`
  - Landing: `tests/e2e/steps/step_t003_chat_given_signed_exchange.go`
  - 語意：建立 per-mode workspace，寫入一行 JSON，`steps` 含一個 step `{tool: "read_files", arguments, result, signature: "{token}"}`。

- [X] T004 [P] [BDD-RED] `Given: the session history already holds a tool-using exchange with no provider token`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the session history already holds a tool-using exchange with no provider token`
  - Landing: `tests/e2e/steps/step_t004_chat_given_unsigned_exchange.go`
  - 語意：同 T003，但 step 不含 `signature`（provider-neutral 形狀）。

- [X] T005 [P] [BDD-RED] `Then: the request replayed the earlier tool step "{tool}" carrying the provider token "{token}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request replayed the earlier tool step "{tool}" carrying the provider token "{token}"`
  - Landing: `tests/e2e/steps/step_t005_chat_then_replayed_with_token.go`
  - 語意：fake 恰記錄一次請求，其 assistant tool-call message 為 `{tool}` 且其 provider token 等於 `{token}`（verbatim），後接對應 tool result。

- [X] T006 [P] [BDD-RED] `Then: the request replayed the earlier tool step "{tool}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request replayed the earlier tool step "{tool}"`
  - Landing: `tests/e2e/steps/step_t006_chat_then_replayed.go`
  - 語意：fake 恰記錄一次請求，其 assistant tool-call message 為 `{tool}`，後接對應 tool result。

### UNIT（非 DSL 的單元斷言）

- [X] T007 [P] [UNIT] `history.Step` 的 signature JSON round-trip + omitempty 相容
  - Read:
    - `specs/plans/014-session-replay-fidelity/research.md` -> Decision 3, 4
    - `specs/truth/techstack.md` -> CLI Application（Session history store）+ Testing & Verification（Pure-helper unit tests）
    - `internal/infrastructure/history/file_store_widened_test.go`
  - 撰寫：append → reload 後 `Step.Signature` 保留；signature 為空時序列化與 round-008 形狀 byte-identical（無新欄位）。
  - 落點：`internal/infrastructure/history/file_store_widened_test.go`。

- [X] T008 [P] [UNIT] `BuildMessages` 重播輸出 `Signature`
  - Read:
    - `specs/plans/014-session-replay-fidelity/research.md` -> Decision 2
    - `internal/agent/agentloop.go`、`internal/agent/agentloop_test.go`
  - 撰寫：`BuildMessages` 對帶 signature 的 `history.Step` 產生的 `llm.ToolCall` 帶 `Signature`；`call_step_<n>` id 不變。
  - 落點：`internal/agent/agentloop_test.go`。

### Phase Review Gate

- [X] T009 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/replaying-a-tool-using-conversation.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/history/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/fakeprovider/fakeprovider.go`
    - `internal/domain/history/history.go`、`internal/agent/agentloop.go`
  - 檢驗：無 undefined step；4 個新句 stepdef + 2 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（`BuildMessages` 尚未帶 signature）。

---

## Phase 4A: ADD Feature File - cli/chat/replaying-a-tool-using-conversation.feature

**Goal**: 讓 `replaying-a-tool-using-conversation.feature` 全綠 — 於 `AgentLoop` 記錄每個 tool step 的 provider signature，並在 `BuildMessages` 重播進合成的 `llm.ToolCall`，使 resumed Gemini session 忠實重播其 tool step。

**Shared Must Read**:
- `specs/truth/features/cli/chat/replaying-a-tool-using-conversation.feature` -> `Feature: Replaying a tool-using conversation`
- `specs/truth/features/cli/chat/dsl.md` -> `the session history already holds a tool-using exchange carrying the provider token "…"`、`… with no provider token`、`the request replayed the earlier tool step "…" carrying the provider token "…"`、`the request replayed the earlier tool step "…"`、`a configured Gemini provider "…" whose endpoint answers with "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/replaying-a-tool-using-conversation.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Session history store / Agent tool loop / Vertex/Gemini adapter）+ `/axb-data-plan` MODIFY（`data-model.dbml` `history_step`）
- `specs/plans/014-session-replay-fidelity/research.md` -> Decision 1, 2, 3, 5（per-step signature；記錄／重播點；store round-trip + omitempty；legacy best-effort）

**Boundary**:
- 產品碼：`internal/agent/agentloop.go` 於建立 `history.Step` 時記錄 `Signature: tc.Signature`，並在 `BuildMessages` 合成的 `llm.ToolCall` 帶入 `Signature: s.Signature`；`internal/domain/history.Step` 的 `Signature`（T002 已加）以 `omitempty` 序列化。
- store（`internal/infrastructure/history/file_store.go`）**不需**變更（struct round-trip 免費）；gemini adapter **不需**變更（已於 round 013 捕捉／echo `thoughtSignature`）。
- `call_step_<n>` deterministic id 不變；OpenAI-family 請求 byte 不變；`stdout` 契約不變；class-phrase 詞彙不變（11）；exit-code 不變。
- **Legacy best-effort（Q2 = Option 1）**：缺失 signature 時照樣重播，不加任何 pre-flight 偵測。

**Test Scope**:
- `specs/truth/features/cli/chat/replaying-a-tool-using-conversation.feature`

- [X] T010 [BDD-GREEN] 讓 Test Scope 全綠（並使 T008 的 `BuildMessages` `[UNIT]` 轉綠）
- [X] T011 [BDD-REFACTOR] 在綠燈下整理 signature 的記錄／重播路徑

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（OpenAI-family 請求 byte 不變、`stdout` byte-exact、exit-code 與 class-phrase 契約、offline paths、非工具回合與空 signature 行 byte-identical），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T012 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**：暫時讓 `BuildMessages` 的重播不帶 `Signature`，確認 `chat/replaying-a-tool-using-conversation.feature` 的 gemini 場景失敗（fake 記錄的重播 tool call 缺 token）；觀察到失敗即還原。
  - 確認：`history.jsonl` 對非工具回合與空 signature 步驟 byte-identical；OpenAI-family 請求 byte 不變；`stdout` byte-exact；exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/data/data-model.dbml` -> `history_step`（signature 欄；record Note） | T002、T003、T004、T007、T010 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session history store — per-step signature） | T002、T003、T007、T010、T011 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Agent tool loop — records the signature） | T002、T008、T010、T011 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Vertex/Gemini adapter — persisted） | T005、T010、T012 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner — two-process resume witness） | T001、T003、T005、T006、T009、T012 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Local fake provider — replayed signature） | T005、T006、T012 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests — Step round-trip + replay） | T002、T007、T008、T012 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T002、T005、T007、T008、T010–T012 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務；T012 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` MODIFY（`specs/truth/data/data-model.dbml`） | T002、T003、T004、T007、T010 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/replaying-a-tool-using-conversation.feature`） | T003–T006、T010、T011 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md`） | T003–T006、T009 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/dsl.md` — round-014 note） | T002、T009 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T012 驗證詞彙維持 11） | PASS |
| `research.md` -> Decision 1（persist per tool step） | T002、T003、T004、T007、T010 | PASS |
| `research.md` -> Decision 2（capture in the loop；replay in `BuildMessages`） | T002、T008、T010、T011 | PASS |
| `research.md` -> Decision 3（store round-trip + omitempty） | T002、T007、T010、T012 | PASS |
| `research.md` -> Decision 4（provider-neutral, opaque, verbatim） | T005、T010 | PASS |
| `research.md` -> Decision 5（legacy best-effort） | T010（Boundary）、T012 | PASS |
| `research.md` -> Decision 6（hermetic two-process resume witness） | T001、T003、T005、T006、T009、T012 | PASS |
| `research.md` -> Decision 7（no new module；BDD techstack unchanged） | T012（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 8（must-ask questions settled） | T010–T012（不改 runner/端） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-008`、`NFR-001`–`NFR-003` | T003–T008（對齊）、T010–T011（交付）、T012 | PASS |
| `plan.md` -> Source-code structure（`internal/domain/history`、`internal/agent/agentloop.go`、`tests/e2e/**`） | T001、T002、T007、T008、T010、T011 | PASS |
| `plan.md` -> Scope notes（api NOOP；data MODIFY；CLI end → /axb-dsl-refine） | T009、T012 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
