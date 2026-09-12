# Tasks: Session History Persistence (round 007)

**Plan Package**: `specs/plans/007-session-history-persistence`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only — `encoding/json`、`os`；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- **PR #20 Architectural Review 指引已納入**（verdict：FULL ARCHITECTURAL APPROVAL、無 blocker）：**TD-1**（`historyStoreFactory` DI seam）、**TD-2**（整檔讀取的有界 reader）、**RF-1**（`--new` archive 冪等/非破壞）、**RF-2**（`-l N` 契約：正整數驗證、存量 clamp、離線短路、`role: content` 格式）、**RF-3**（`llm.Request` 擴充向後相容）——分別落在對應 task 的 `只做` / `Boundary`。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 2: Foundational

**Goal**: 建立 history domain port、file adapter、`llm.Request` prior-message 擴充、CLI wiring、E2E 共用元件（fake 記錄 messages、history 檔案 helper、arranged-exchanges 場景狀態）與 11 個新句 stepdef 落點骨架（Zero Shared Edits 原則）。

- [X] T001 建立 history domain port 落點骨架 `internal/domain/history/history.go`
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 1, 2
    - `specs/truth/data/data-model.dbml` -> `history_entry` / `history_location`
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Conversation context）
  - 只做：宣告 `internal/domain/history` 套件與 `Entry` value type（`Prompt`, `Answer`）及 `Store` 介面簽名 stub（例如 `Load() ([]Entry, error)`、`Append(Entry) error`、`Archive() error`）。
  - 不做：不實作任何檔案 I/O；不接 `internal/cli`；不寫斷言。

- [X] T002 建立 history file adapter 落點骨架 `internal/infrastructure/history/file_store.go`
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 1, 3
    - `specs/truth/techstack.md` -> CLI Application（Session history store）
    - `internal/domain/history/history.go`
  - 只做：宣告 adapter 型別與 constructor 簽名，留下 `<workspace>/history.jsonl` 與 `<workspace>/history.archive.jsonl` 路徑解析 stub；讀取骨架採**有界 reader**（`json.Decoder`／`bufio.Reader`，或 `scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)`），避免 `bufio.Scanner` 預設 64KB token 上限（**TD-2**）。
  - 不做：不實作 append / load / archive；不碰 `internal/cli`。

- [X] T003 擴充 `llm.Request` 與 OpenAI adapter 落點骨架
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Conversation context / Request assembly）
    - `internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`
  - 只做：在 `llm.Request` 新增 prior-message 欄位（`[]Message{Role, Content}`）之型別骨架；在 `requestBody` 留下把「resumed history + current prompt」組成 `messages` 陣列之簽名/stub（**RF-3**：`req.Messages` 為空時 `messages` 恰為當前單一 user 訊息，與 round 004–006 byte-for-byte 一致）。
  - 不做：不改 transport、錯誤處理或 response 解析。

- [X] T004 CLI wiring 落點骨架 `internal/cli/cli.go`
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 4, 5, 6
    - `specs/truth/techstack.md` -> CLI Application（Session lifecycle flags）
    - `internal/cli/cli.go`
  - 只做：新增 `--new`（bool）與 `-l`（int）flag 解析骨架；宣告注入 `history.Store` 的 seam（DI，採 `historyStoreFactory` 工廠型別，mirror `gatewayFactory` / `newGateway`，**TD-1**）；在 `run` 留下 `--version` → `-d` → `-l` → (`--new`) prompt turn → boot 的 dispatch stub；宣告「history I/O 失敗 → environment class phrase + code 4」的 hook 簽名；`-l` 於 `N ≤ 0` 立即 `emitUsageError`（code 2）且在任何網路/stdin 之前短路（**RF-2**）。
  - 不做：不實作 resume / persist / archive / list 行為；不改既有 dispatch 與 exit-code（產品行為留 Phase 4）。

- [X] T005 建立 E2E 共用元件落點骨架（fake 記錄 messages、history 檔案 helper、arranged-exchanges 場景狀態）
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 8
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner / Local fake provider）
    - `tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 只做：擴充 fake provider 以解析並記錄 request 的 `messages` 陣列；新增讀 `<workspace>/history.jsonl` / `history.archive.jsonl` 的 harness helper；在 scenario context 保存 Given 安排的 exchanges。
  - 不做：不寫具體 step 斷言；不碰 `internal/`。

- [X] T006 建立 11 個新句 stepdef 獨立檔案骨架 `tests/e2e/steps/step_t007_*.go`–`step_t017_*.go`
  - Read:
    - `specs/truth/features/cli/history/dsl.md`（新句：`the operator starts a fresh session with "--new"`、`the operator asks tellme to list the last {count} messages`、`the active session history holds no exchanges`、`the archived session history holds the exchange "{prompt}" and "{answer}"`、`tellme lists the last {count} messages`、`tellme lists no messages`、`tellme sends no request to any provider`）
    - `specs/truth/features/cli/chat/dsl.md`（新句：`tellme stored the exchange "{prompt}" and "{answer}" in the session history`、`the request carried the earlier exchange "{prompt}" and "{answer}"`、`the request carried no earlier exchange`）
    - `specs/truth/features/cli/dsl.md`（新句：`the session history already holds the exchanges:`）
    - `tests/e2e/steps/register.go`
  - 只做：建立 11 個獨立檔案（各對應 T007–T017 的句），各自 `init()` 自我註冊空白 registrar。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的句與本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：介面根 `specs/truth/features/cli/dsl.md`（跨模組句），或同模組 `specs/truth/features/cli/{chat,history}/dsl.md`。不得掃其他模組。
- 本輪 11 個新句：`the session history already holds the exchanges:` 在**介面根**（chat 與 history 共用）；3 個 `chat` 句在 `chat/dsl.md`；7 個 `history` 句在 `history/dsl.md`。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：**無**。本輪無既有句語意被改到需要改既有 stepdef（既有 `chat`/`configuration`/`workspace` 句語意不變；`answering-a-single-prompt.feature` 為 NOOP）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 11 個新句（`ADD`）。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 1, 2, 4）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/dsl.md` -> `the session history already holds the exchanges:`
- `specs/truth/features/cli/chat/dsl.md` -> `tellme stored the exchange …` / `the request carried the earlier exchange …` / `the request carried no earlier exchange`
- `specs/truth/features/cli/history/dsl.md` -> 7 個 `history` 新句
- `truth-delta.md` -> `/axb-dsl-refine` ADD ×3（features）+ MODIFY（chat/dsl.md、root dsl.md）+ ADD（history/dsl.md）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 `internal/infrastructure/history/*_test.go`、`internal/domain/llm/*_test.go`、`internal/cli/*_test.go`，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T007–T018 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T019 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [X] T007 [P] [BDD-RED] `Given: the session history already holds the exchanges:`
  - Read: `specs/truth/features/cli/dsl.md` -> `the session history already holds the exchanges:`
  - Landing: `tests/e2e/steps/step_t007_history_given_holds_exchanges.go`
  - 語意：建立 per-mode session workspace 目錄（若不存在），依 DataTable 逐列寫一行 JSON（`prompt`、`answer`）進 `history.jsonl`（表序）；把 exchanges 存入 scenario context。

- [X] T008 [P] [BDD-RED] `Then: tellme stored the exchange "{prompt}" and "{answer}" in the session history`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme stored the exchange "{prompt}" and "{answer}" in the session history`
  - Landing: `tests/e2e/steps/step_t008_chat_then_stored_exchange.go`
  - 語意：斷言 `<workspace>/history.jsonl` 存在一列，其 prompt/answer 為 `{prompt}`/`{answer}`。

- [X] T009 [P] [BDD-RED] `Then: the request carried the earlier exchange "{prompt}" and "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried the earlier exchange "{prompt}" and "{answer}"`
  - Landing: `tests/e2e/steps/step_t009_chat_then_request_carried_exchange.go`
  - 語意：斷言 fake 記錄到**恰好一個**請求，其 `messages` 依序在最前帶有 user=`{prompt}`、assistant=`{answer}` 的先前訊息（在當前 prompt 之前）。

- [X] T010 [P] [BDD-RED] `Then: the request carried no earlier exchange`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried no earlier exchange`
  - Landing: `tests/e2e/steps/step_t010_chat_then_request_carried_nothing.go`
  - 語意：斷言 fake 記錄到**恰好一個**請求，其 `messages` 僅為當前 user prompt。

- [X] T011 [P] [BDD-RED] `When: the operator starts a fresh session with "--new"`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `the operator starts a fresh session with "--new"`
  - Landing: `tests/e2e/steps/step_t011_history_when_start_fresh.go`
  - 語意：以 `tellme --new` 執行（無 `-c`、無 positional prompt）；擷取 exit code、stdout、stderr 與產生的 workspace 檔案。

- [X] T012 [P] [BDD-RED] `When: the operator asks tellme to list the last {count} messages`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `the operator asks tellme to list the last {count} messages`
  - Landing: `tests/e2e/steps/step_t012_history_when_list_last.go`
  - 語意：以 `tellme -l {count}` 執行（無 `-c`、無 positional prompt）；擷取 exit code、stdout、stderr。

- [X] T013 [P] [BDD-RED] `Then: the active session history holds no exchanges`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `the active session history holds no exchanges`
  - Landing: `tests/e2e/steps/step_t013_history_then_active_empty.go`
  - 語意：斷言 `<workspace>/history.jsonl` 不含任何 exchange 列。

- [X] T014 [P] [BDD-RED] `Then: the archived session history holds the exchange "{prompt}" and "{answer}"`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `the archived session history holds the exchange "{prompt}" and "{answer}"`
  - Landing: `tests/e2e/steps/step_t014_history_then_archived_holds_exchange.go`
  - 語意：斷言 `<workspace>/history.archive.jsonl` 存在一列，其 prompt/answer 為 `{prompt}`/`{answer}`。

- [X] T015 [P] [BDD-RED] `Then: tellme lists the last {count} messages`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `tellme lists the last {count} messages`
  - Landing: `tests/e2e/steps/step_t015_history_then_lists_last.go`
  - 語意：依 scenario context 的 arranged exchanges，斷言 stdout 依序列出最後 `{count}` 個訊息（每行 `role: content`，與 `internal/cli` 的 `"%s: %s\n"` 格式一致，**RF-2**）。

- [X] T016 [P] [BDD-RED] `Then: tellme lists no messages`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `tellme lists no messages`
  - Landing: `tests/e2e/steps/step_t016_history_then_lists_none.go`
  - 語意：斷言 stdout 不含任何訊息。

- [X] T017 [P] [BDD-RED] `Then: tellme sends no request to any provider`
  - Read: `specs/truth/features/cli/history/dsl.md` -> `tellme sends no request to any provider`
  - Landing: `tests/e2e/steps/step_t017_history_then_no_provider_request.go`
  - 語意：斷言所有設定的 fake provider 都記錄到**零**請求。

### UNIT（pure-helper 單元測試）

- [X] T018 [P] [UNIT] history store 與 request 組裝單元測試
  - Read:
    - `specs/plans/007-session-history-persistence/research.md` -> Decision 1, 2, 4
    - `specs/truth/data/data-model.dbml` -> `history_entry`
    - `specs/truth/techstack.md` -> CLI Application（Session history store）、Testing & Verification（Pure-helper unit tests）
    - `internal/domain/history/history.go`、`internal/infrastructure/history/file_store.go`、`internal/domain/llm/gateway.go`
  - Landing: `internal/infrastructure/history/file_store_test.go`、`internal/domain/llm/gateway_test.go`、`internal/cli/cli_test.go`（表驅動）
  - 撰寫：history 路徑解析、append（一 turn 一列）、reload（列→有序 messages）、`--new` archive 搬移（含 active 不存在時 no-op 回 `nil`，**RF-1**）；`llm.Request` 的 messages 組裝（空 history = 單一 user 訊息，byte-identical，**RF-3**）；`-l` 模式選擇 + 正整數 count 驗證（`N ≤ 0` 為 usage error；存量 M < N 時印全部，**RF-2**）；大型 prompt（至 1 MiB）reload 不得因 reader 上限失敗（**TD-2**）。
  - 註：CLI 單元測試以 `historyStoreFactory` 注入 in-memory fake store（**TD-1**），零磁碟 I/O。

### Phase Review Gate

- [X] T019 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/remembering-the-conversation.feature`、`specs/truth/features/cli/history/starting-a-fresh-session.feature`、`specs/truth/features/cli/history/inspecting-the-session-history.feature`
    - `specs/truth/features/cli/dsl.md`、`specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/history/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
  - 檢驗：無 undefined step；11 個新 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/remembering-the-conversation.feature

**Goal**: 實作「每個完成的 prompt turn 被持久化」與「prompt run 以已持久化的對話作為 provider context」與「無持久化對話時不帶先前訊息」，使 `remembering-the-conversation.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/remembering-the-conversation.feature` -> `Feature: Remembering the conversation across runs`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`tellme stored the exchange …`、`the request carried the earlier exchange …`、`the request carried no earlier exchange`）
- `specs/truth/features/cli/dsl.md` -> `the session history already holds the exchanges:`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/remembering-the-conversation.feature`）+ MODIFY（`chat/dsl.md`）
- `specs/truth/data/data-model.dbml` -> `history_entry`
- `specs/truth/techstack.md` -> CLI Application（Session history store）、Reasoning & Provider Transport（Conversation context）
- `specs/plans/007-session-history-persistence/research.md` -> Decision 1, 2, 7
- `internal/domain/history/`、`internal/infrastructure/history/`、`internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`、`internal/cli/cli.go`

**Boundary**:
- Prompt turn 完成後把 `{prompt, answer}` 以單一 JSON 列 append 進 `<workspace>/history.jsonl`（append-after-complete；不寫時間戳/ID）；resume 時整檔讀入，展開為 user/assistant 訊息，前置於當前 prompt 送給 provider。空 history 時 `messages` 與 round-004 單一 user 訊息一致（byte-identical，**RF-3**）。
- 整檔讀取採**有界 reader**（見 T002 / **TD-2**），可處理至 1 MiB 的 prompt（round-005 piped stdin），不得因 `bufio.Scanner` 預設上限而 `ErrTooLong`。
- store 以 `historyStoreFactory` 注入（**TD-1**）；不實作 `--new`（屬 Phase 4B）；不實作 `-l`（屬 Phase 4C）。

**Test Scope**:
- `specs/truth/features/cli/chat/remembering-the-conversation.feature`

- [X] T020 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T021 [BDD-REFACTOR] 在綠燈下整理 history store 落點與 request 組裝

## Phase 4B: ADD Feature File - cli/history/starting-a-fresh-session.feature

**Goal**: 實作 `--new`：開始 fresh session（active history 淨空）並保留前一對話（archive），使 `starting-a-fresh-session.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/history/starting-a-fresh-session.feature` -> `Feature: Starting a fresh session`
- `specs/truth/features/cli/history/dsl.md` -> When（`the operator starts a fresh session with "--new"`）、Then（`the active session history holds no exchanges`、`the archived session history holds the exchange …`）
- `specs/truth/features/cli/dsl.md` -> `the session history already holds the exchanges:`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`history/starting-a-fresh-session.feature`）
- `specs/truth/data/data-model.dbml` -> `history_location`
- `specs/plans/007-session-history-persistence/research.md` -> Decision 3
- `internal/infrastructure/history/`、`internal/cli/cli.go`

**Boundary**:
- `--new`：把 active `history.jsonl` 的列 append 進 `history.archive.jsonl`（不存在則建立），再移除 active 檔（確定性單一 archive；保留、不銷毀）。
- **RF-1**：active 檔不存在時 `Archive()` 為 **no-op 回 `nil`**（非 `os.ErrNotExist`）；archive 以 `os.O_APPEND|os.O_CREATE|os.O_WRONLY` 開啟並在 `os.Remove(activePath)` **之前** flush/sync，避免中途失敗時遺失 active。
- 不實作 `-l`（屬 Phase 4C）；不改 prompt turn 的 resume/persist（屬 Phase 4A）。

**Test Scope**:
- `specs/truth/features/cli/history/starting-a-fresh-session.feature`

- [X] T022 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T023 [BDD-REFACTOR] 在綠燈下整理 archive 搬移與 active 重置

## Phase 4C: ADD Feature File - cli/history/inspecting-the-session-history.feature

**Goal**: 實作 `-l N`：印出最近 N 則訊息且不觸網，使 `inspecting-the-session-history.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature` -> `Feature: Inspecting the session history`
- `specs/truth/features/cli/history/dsl.md` -> When（`the operator asks tellme to list the last {count} messages`）、Then（`tellme lists the last {count} messages`、`tellme lists no messages`、`tellme sends no request to any provider`）
- `specs/truth/features/cli/dsl.md` -> `the session history already holds the exchanges:`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`history/inspecting-the-session-history.feature`）
- `specs/plans/007-session-history-persistence/research.md` -> Decision 4, 5
- `internal/cli/cli.go`、`internal/infrastructure/history/`

**Boundary**:
- `-l N`：讀 active `history.jsonl`，把最後 N 則訊息以純文字（`role: content`，每則一行）印到 stdout 後 exit；**不發 provider 請求**。輸出為純文字（history 的 rendering 屬後續 slice）。
- **RF-2**：`N ≤ 0`（或缺值）於 `cli.Run` 立即 `emitUsageError`（既有 usage class phrase + code `2`）；存量 `M < N`（含 `M == 0`）時印出全部 `M` 則並以 code `0` 結束；`-l` 在任何網路呼叫與 stdin 讀取**之前**求值（嚴格離線，`NFR-001`）；每行格式固定為 `"%s: %s\n"`（`role: content`），且 `tests/e2e/steps/step_t015` 與 `internal/cli` 必須一致。
- 不改 stdin 組合；不實作 `--new`（屬 Phase 4B）。

**Test Scope**:
- `specs/truth/features/cli/history/inspecting-the-session-history.feature`

- [X] T024 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T025 [BDD-REFACTOR] 在綠燈下整理 `-l` 模式選擇與輸出格式

## Phase 4D: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（含 `-l`/`--new` 離線路徑與既有 CLI 契約）。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T026 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6` 與既有 CLI 契約（round 001–006）維持綠燈；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不觸網；fresh-workspace 的既有 chat 場景行為不變；`go mod tidy` 後 module graph 不變（本輪無新相依）。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/data/data-model.dbml` -> `history_entry`（prompt/answer；(location,position) pk） | T001、T005、T008、T013、T014、T018、T020 | PASS |
| `specs/truth/data/data-model.dbml` -> `history_location`（active/archived） | T001、T002、T018、T022、T023 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session history store：append-only JSON-Lines `history.jsonl`） | T002、T018、T020、T021 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session lifecycle flags：`--new` archive + `-l N`） | T004、T011、T012、T022、T024、T025 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Conversation context：`llm.Request` prior messages） | T003、T018、T020、T021 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner filesystem 斷言 / fake 記錄 messages / pure-helper 擴充 / offline 集合含 `-l`、prompt-less `--new`） | T005、T009、T010、T017、T018、T026 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（SQLite/db 措辭、summarisation/pruning deferred、out-of-scope） | 負向決策 -> T002/T021/T023 Boundary、T026（不引入 SQLite/streaming/pinning） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY ×4（`specs/truth/techstack.md`） | T002、T003、T005、T018、T020、T022、T024 | PASS |
| `truth-delta.md` -> `/axb-data-plan` ADD（`specs/truth/data/data-model.dbml`） | T001、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/remembering-the-conversation.feature`） | T006、T008–T010、T019、T020、T021 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` 3 新 Then 句） | T006、T008–T010、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`history/starting-a-fresh-session.feature`） | T006、T011、T013、T014、T019、T022、T023 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`history/inspecting-the-session-history.feature`） | T006、T012、T015、T016、T017、T019、T024、T025 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`history/dsl.md` 2 When + 5 Then） | T006、T011–T017、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（root `cli/dsl.md` 新 Given 句） | T006、T007、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`chat/answering-a-single-prompt.feature`） | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |
| `research.md` -> Decision 1（history store：append-only JSON-Lines） | T001、T002、T018、T020 | PASS |
| `research.md` -> Decision 2（resume：`llm.Request` prior messages） | T003、T018、T020、T021 | PASS |
| `research.md` -> Decision 3（`--new` 單一 deterministic archive） | T002、T022、T023 | PASS |
| `research.md` -> Decision 4（`-l N` 純文字、無 provider 請求、正整數） | T012、T015、T016、T024、T025 | PASS |
| `research.md` -> Decision 5（flag 介面 + dispatch precedence） | T004、T011、T012、T024 | PASS |
| `research.md` -> Decision 6（failure contract：environment phrase + code 4） | T004、T026 | PASS |
| `research.md` -> Decision 7（持久化內容為 provider answer text） | T008、T020 | PASS |
| `research.md` -> Decision 8（fake 記錄 messages；offline 擴充） | T005、T009、T010、T017 | PASS |
| `spec.md` -> US1/US2/US3 驗收 + `FR-001`–`FR-012`、`NFR-001`–`NFR-003` | T007–T017（對齊）、T020–T025（交付）、T026 | PASS |
| PR #20 Architectural Review -> TD-1 / TD-2 / RF-1 / RF-2 / RF-3 | T002（TD-2）、T003（RF-3）、T004（TD-1、RF-2）、T018（TD-1/TD-2/RF-1/RF-2/RF-3）、Phase 4A（TD-1/TD-2）、Phase 4B（RF-1）、Phase 4C（RF-2）、T015（RF-2 格式） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
