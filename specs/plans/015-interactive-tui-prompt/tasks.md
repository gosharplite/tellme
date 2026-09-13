# Tasks: Interactive TUI Prompt — Suggestions, Session Dashboard, and the Shared Global Prompt Log (round 015)

**Plan Package**: `specs/plans/015-interactive-tui-prompt`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **本輪有新增技術**：`tellme` 的**第一個 TUI 相依家族**（`bubbletea` + `bubbles/textarea`；`lipgloss` 由 indirect 升為 direct）。依 SOP 建立 Phase 1 `Setup`（套件名、配置、技術環境、smoke-test），不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（本輪 16 句皆為全新句，`ADD`）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：新增 `internal/ui/tui/prompt/**`（Bubble Tea 提示）、`internal/app/suggestions/**`（建議引擎）、`internal/domain/suggestions/**` + prompt-tracker port、`internal/infrastructure/history/global_prompt_tracker.go`（共享提示紀錄 store）；`internal/config/config.go` 加 `USE_TUI_PROMPT`、`internal/cli/cli.go` 加 `-i` 與 TUI dispatch。相位順序仍為「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。

## PR #38 Review Directives (implementation constraints — MUST)

> 來自 PR [#38](https://github.com/gosharplite/tellme/pull/38) 的架構審查（[#5657097770](https://github.com/gosharplite/tellme/pull/38#issuecomment-5657097770)，evaluated head `c66f500`，verdict **PLAN + TRUTH APPROVED — PROCEED WITH ARCHITECTURAL DIRECTIVES**）。以下為實作約束，已寫入對應 task 的 `Read`／`Boundary`。

- **[BLOCKER] TUI output stream containment** — 以 `tea.NewProgram(model, tea.WithOutput(env.stderr), tea.WithInput(env.stdin))` 綁定；TUI 只寫 `env.stderr`（或注入的測試 writer），**`stdout` 保持 byte-exact**（`FR-015`）。見 T004、T008、T029、T032–T039。
- **[TECHNICAL DEBT] package/domain-boundary alignment** — **不得**建立 `internal/domain/config` 或集中式 `internal/domain/ports`；config 直接改 `internal/config/config.go`；port 落專責子領域（`internal/domain/suggestions/`、prompt-tracker port）；協調者 `internal/app/suggestions/service.go`；adapter `internal/infrastructure/history/global_prompt_tracker.go`。見 T005–T008。
- **[TECHNICAL DEBT] suggestion-engine I/O bounds** — 不做無界 `filepath.WalkDir`；以 `filepath.Split(query)` 收斂範圍、`ReadDir` 分批（≤100）、`ctx.Err()` 立即讓出、跳過 ignore 目錄（`.git`、`node_modules`）、收滿 10 筆即停。見 T006、T027、T032。
- **[REFACTOR] TUI-runner DI seam** — `internal/cli/cli.go` 導入 `tuiPromptRunner` factory 變數（比照 `gatewayFactory`／`historyStoreFactory`），使 `-i`／`USE_TUI_PROMPT`／非 TTY 的 dispatch 可在無終端事件迴圈下單元測試。見 T008、T030。
- **[REFACTOR] optimistic concurrency / async append safety** — `Append` 以 `O_APPEND|O_CREATE|O_WRONLY` 開檔；compaction 於背景非同步執行，先快照檔案大小、寫回前驗證 `newSize == initialSize` 才 `AtomicWrite`（否則指數退避重試）；`GlobalPromptTracker` 與 `SuggestionService` 皆實作 `Close(ctx context.Context) error`，以 `sync.WaitGroup` 於行程退出前排空。見 T007、T028、T036。

---

## Phase 1: Setup

**Goal**: 引入本輪新增技術（Bubble Tea 家族），備妥 TUI 執行環境與單一來源公告字面，最後以最小 smoke-test 確認 TTY-gated `-i` 路徑可建置連結。不寫 DSL 語意、不寫提示／建議產品行為。

- [ ] T001 加入 `bubbletea` + `bubbles`（textarea）並將 `lipgloss` 升為 direct
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Interactive TUI prompt (`-i`)）、Not Introduced Yet（TUI libraries — introduced in round 015）
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 1, Decision 7
  - `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles`；於 `go.mod` 將 `github.com/charmbracelet/lipgloss` 由 indirect 升為 direct；`go mod tidy`。
  - 只加這三個 TUI 家族模組，不加其他新相依（無 fuzzy／glob 套件）。

- [ ] T002 備妥 TUI 執行環境與單一來源公告字面 `cli.TUIHint`
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Terminal detection — 重用 round-012 real-isatty seam）、Interactive prompt enable（`USE_TUI_PROMPT` + `-i`）
    - `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is shown`（其 `來源`: `cli.TUIHint`）
    - `internal/cli/cli.go`
  - 於 `internal/cli` 定義單一來源公告字面 `const TUIHint = "[Interactive prompt. Type a prompt; Tab accepts a suggestion, Ctrl+S sends, Esc cancels]"`；確認注入式 I/O seam（`env.stdin`/`env.stderr`）與 `golang.org/x/term` real-isatty seam 已可重用。
  - 只定義字面與 seam 掛點；不實作提示渲染、不讀 stdin、不寫提示語意。

- [ ] T003 smoke-test：TTY-gated `-i` 路徑可建置與連結
  - Read:
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner / step definitions；Test strategy）
    - `Makefile`（`build`、`verify`）
  - `go build ./...` 綠，且 `make verify` 的既有 gate（gofmt/vet/staticcheck/golangci-lint/govulncheck）維持綠；確認新模組不破壞 offline-path witness。
  - 不跑本輪 Feature，不寫 stepdef 語意，不寫任何提示行為。

---

## Phase 2: Foundational

**Goal**: 建立本輪產品碼與測試層的落點骨架（Zero Shared Edits），讓後續 Phase 3／Phase 4 不各自發明檔案、seam 或套件路徑。只建立落點與載體，不寫建議／提示／紀錄行為。

- [ ] T004 建立 `internal/ui/tui/prompt/` 套件骨架（Bubble Tea 模型 + 建議清單 + 編輯器）
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 1, Decision 6
    - `specs/plans/015-interactive-tui-prompt/plan.md` -> Source-code structure
    - `specs/plans/015-interactive-tui-prompt/ui/ui-plan.md`（terminal 模式；畫面／panel 切分）
  - 只做：建立 `internal/ui/tui/prompt/model.go`（`tea.Model` 骨架；注入 `input io.Reader`、`output io.Writer`、`suggestion source`；**output 綁定 `env.stderr`**，見 review blocker）、`suggester.go`（selection cursor 骨架）、`textarea.go`（`bubbles/textarea` 包裝骨架）。所有注入點以參數形式存在，供單元測試注入。
  - 不做：不寫 seed／refresh／accept 行為；不寫 keybinding 邏輯；不寫 dashboard；不碰 `env.stdout`。

- [ ] T005 於專責子領域建立 suggestion-source port 與 prompt-tracker port（**不得**用 `internal/domain/ports`）
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 2, Decision 3
    - `specs/plans/015-interactive-tui-prompt/plan.md` -> Source-code structure；PR #38 review directive ②
    - `specs/truth/data/data-model.dbml` -> `prompt_log_entry`
  - 只做：建立 `internal/domain/suggestions/suggestions.go`（`Suggestion` 值型別 + `SuggestionService` port）與 prompt-tracker port（`internal/domain/history/` 下，例如 `tracker.go`：`Append`／`Recent`／`Close(ctx) error`）。
  - 不做：不建立 `internal/domain/config` 或集中式 `internal/domain/ports`；不寫引擎或 store 實作。

- [ ] T006 建立 `internal/app/suggestions/service.go` 協調者骨架（多來源引擎外殼）
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> CLI Application（Prompt suggestion engine）
    - `internal/domain/suggestions/suggestions.go`
  - 只做：建立 `multiSourceSuggestionService` 外殼，留出 recent-prompt / workspace-path / tool 三來源與 debounce 掛點；workspace 掃描以 `filepath.Split(query)` 收斂、`ReadDir` 分批、`ctx.Err()` 讓出、ignore 目錄跳過、10 筆上限（見 review directive ③）的函式殼。
  - 不做：不寫 subsequence 比對結果、不寫 dedupe、不寫 debounce 計時語意。

- [ ] T007 建立 `internal/infrastructure/history/global_prompt_tracker.go` store 骨架（append-only JSONL）
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 3
    - `specs/truth/data/data-model.dbml` -> `prompt_log_entry`
    - `specs/truth/techstack.md` -> CLI Application（Shared global prompt log）
  - 只做：建立 store 骨架——`output/` **root** 的 `global_prompts.jsonl` 路徑解析、`Append`（以 `O_APPEND|O_CREATE|O_WRONLY` 開檔，見 review directive ⑤）、`Recent`（newest-first + dedupe 讀取）與 `Close(ctx) error`（`sync.WaitGroup` 排水）的函式殼；保留 async compaction（`newSize == initialSize` 才 `AtomicWrite`）掛點。
  - 不做：不寫實際寫入／compaction 邏輯、不寫 `{timestamp,prompt}` 序列化斷言。

- [ ] T008 於 `internal/config` 加 `USE_TUI_PROMPT`、於 `internal/cli` 加 `-i` 解析與 `tuiPromptRunner` DI seam
  - Read:
    - `specs/truth/techstack.md` -> Configuration（Interactive prompt enable）、CLI Application（CLI flag parsing — `-i`/`--interactive`）
    - `specs/truth/techstack.md` -> CLI Application（Terminal detection）
    - `internal/config/config.go`、`internal/cli/cli.go`
    - PR #38 review directive ②, ④
  - 只做：`internal/config/config.go` 加 `USE_TUI_PROMPT` 欄位與解析；`internal/cli/cli.go` 新增 `-i`/`--interactive` flag 解析與 `tuiPromptRunner` factory 變數宣告（比照 `gatewayFactory`／`historyStoreFactory`），並在 dispatch 骨架留出 gating（`-i`/`USE_TUI_PROMPT` AND terminal stdin）分支，暫回非 TUI 路徑。
  - 不做：不寫真正的 TUI 呼叫與 rendering；不改既有非 `-i` dispatch 行為；不寫 gating 斷言。

- [ ] T009 建立 16 個新句 stepdef 獨立檔骨架（Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 16 個新句）
    - `tests/e2e/steps/register.go`、`tests/e2e/harness/`（`TELL_ME_FORCE_STDIN_TTY` seam + scripted stdin 既有能力）
  - 只做：建立 16 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），對到 T011–T026：
    - `tests/e2e/steps/step_t011_chat_given_shared_log_holds.go`
    - `tests/e2e/steps/step_t012_chat_when_open_prompt.go`
    - `tests/e2e/steps/step_t013_chat_when_open_and_type.go`
    - `tests/e2e/steps/step_t014_chat_when_submit_at_prompt.go`
    - `tests/e2e/steps/step_t015_chat_when_abort_prompt.go`
    - `tests/e2e/steps/step_t016_chat_when_pipe_with_tui.go`
    - `tests/e2e/steps/step_t017_chat_then_prompt_shown.go`
    - `tests/e2e/steps/step_t018_chat_then_prompt_not_shown.go`
    - `tests/e2e/steps/step_t019_chat_then_offers_recent_prompt.go`
    - `tests/e2e/steps/step_t020_chat_then_offers_workspace_entry.go`
    - `tests/e2e/steps/step_t021_chat_then_offers_tool.go`
    - `tests/e2e/steps/step_t022_chat_then_reports_provider.go`
    - `tests/e2e/steps/step_t023_chat_then_reports_tokens_turns.go`
    - `tests/e2e/steps/step_t024_chat_then_log_records.go`
    - `tests/e2e/steps/step_t025_chat_then_log_unchanged.go`
    - `tests/e2e/steps/step_t026_chat_then_no_request.go`
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [ ] T010 建立 4 個 `[UNIT]` 落點檔骨架
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 2, 3, 4, 6
    - `specs/truth/techstack.md` -> Testing & Verification（Interactive TUI prompt harness；Pure-helper unit tests）
  - 只做：預留 4 個 `[UNIT]` 測試落點空殼——`internal/app/suggestions/service_test.go`（建議引擎）、`internal/infrastructure/history/global_prompt_tracker_test.go`（共享 store）、`internal/ui/tui/prompt/model_test.go`（注入 I/O 的 TUI 模型）、`internal/cli/tui_dispatch_test.go`（gating 矩陣，獨立檔避免與既有 `cli_test.go` 衝突）。採 stdlib `testing`。
  - 不做：不寫任何斷言或 fixture 語意。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 16 句皆為全新句（`ADD`）。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪 16 句皆 `chat` 模組專屬；features 僅重用其既有 `the operator has a runnable tellme installation`／`the runtime home is "…"` 等既有 root row，語意不變且 stepdef 已在）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given／When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪顯示型 Then（`the interactive prompt offers …`／`… reports …`）依 `dsl.md` 是對「captured output 之 **presence**」斷言（非精確 ANSI bytes）；shown／not-shown Then 以單一來源字面 `cli.TUIHint` 斷言。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **16 句**（`ADD` 的 DSL rows）。
- `[UNIT]`：4 個非 DSL 單元斷言（建議引擎／共享 store／TUI 模型／gating 矩陣）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the shared prompt log already holds "{prompt}"`
  -> `the operator opens the interactive prompt`
  -> `the operator opens the interactive prompt and types "{query}"`
  -> `the operator submits the prompt "{prompt}" at the interactive prompt`
  -> `the operator aborts the interactive prompt`
  -> `the operator pipes "{content}" into tellme with the interactive prompt enabled`
  -> `the interactive prompt is shown`
  -> `the interactive prompt is not shown`
  -> `the interactive prompt offers the recent prompt "{prompt}"`
  -> `the interactive prompt offers the workspace entry "{entry}"`
  -> `the interactive prompt offers the available tool "{tool}"`
  -> `the interactive prompt reports the active provider "{provider}"`
  -> `the interactive prompt reports the session's token usage and turn count`
  -> `the shared prompt log records the prompt "{prompt}"`
  -> `the shared prompt log still holds only "{prompt}"`
  -> `tellme sends no request to the provider "{provider}"`
- `specs/truth/features/cli/dsl.md` -> **NOOP**（root 詞彙不變，維持 11 片語）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/prompting-with-suggestions.feature`、`using-the-interactive-prompt.feature`、`recording-the-shared-prompt-log.feature`、`choosing-the-interactive-prompt.feature`）+ MODIFY（`chat/dsl.md`）；NOOP（root `cli/dsl.md`、`contracts/**`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/app/suggestions/service_test.go`、`internal/infrastructure/history/global_prompt_tracker_test.go`、`internal/ui/tui/prompt/model_test.go`、`internal/cli/tui_dispatch_test.go`。
- 不寫產品碼（除 T004–T008 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T011–T030 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；目標檔案互斥，符合 Zero Shared Edits）；T031 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T011 [P] [BDD-RED] `Given: the shared prompt log already holds "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log already holds "{prompt}"`
  - Landing: `tests/e2e/steps/step_t011_chat_given_shared_log_holds.go`
  - 語意：於 `$TELL_ME_HOME/output/`（**root**，非 `output/<mode>/`）建立 `global_prompts.jsonl`（若缺），append 一行 `{"timestamp":"<RFC3339>","prompt":"{prompt}"}`。

- [ ] T012 [P] [BDD-RED] `When: the operator opens the interactive prompt`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator opens the interactive prompt`
  - Landing: `tests/e2e/steps/step_t012_chat_when_open_prompt.go`
  - 語意：以 `TELL_ME_FORCE_STDIN_TTY=1` 與 scripted key sequence（open → abort）跑 `tellme -i`，captured 結果：exit code、stdout、stderr、fake 記錄的請求。

- [ ] T013 [P] [BDD-RED] `When: the operator opens the interactive prompt and types "{query}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator opens the interactive prompt and types "{query}"`
  - Landing: `tests/e2e/steps/step_t013_chat_when_open_and_type.go`
  - 語意：跑 `tellme -i`（forced-terminal + scripted：open → type `{query}` → abort），captured 結果；prompt 已對 `{query}` 計算建議。

- [ ] T014 [P] [BDD-RED] `When: the operator submits the prompt "{prompt}" at the interactive prompt`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator submits the prompt "{prompt}" at the interactive prompt`
  - Landing: `tests/e2e/steps/step_t014_chat_when_submit_at_prompt.go`
  - 語意：跑 `tellme -i`（forced-terminal + scripted：open → type `{prompt}` → submit），captured 結果 + fake 記錄的請求（提交文字成為一次推理回合的 prompt）。

- [ ] T015 [P] [BDD-RED] `When: the operator aborts the interactive prompt`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator aborts the interactive prompt`
  - Landing: `tests/e2e/steps/step_t015_chat_when_abort_prompt.go`
  - 語意：跑 `tellme -i`（forced-terminal + scripted：open → abort），captured 結果（無 prompt 送出）。

- [ ] T016 [P] [BDD-RED] `When: the operator pipes "{content}" into tellme with the interactive prompt enabled`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator pipes "{content}" into tellme with the interactive prompt enabled`
  - Landing: `tests/e2e/steps/step_t016_chat_when_pipe_with_tui.go`
  - 語意：以 piped stdin 跑 `tellme -i`（無 positional prompt），captured 結果 + fake 記錄的請求（非終端輸入走 piped 路徑）。

- [ ] T017 [P] [BDD-RED] `Then: the interactive prompt is shown`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is shown`
  - Landing: `tests/e2e/steps/step_t017_chat_then_prompt_shown.go`
  - 語意：captured stderr 帶有單一來源公告字面 `cli.TUIHint`；**不得**出現在 stdout。

- [ ] T018 [P] [BDD-RED] `Then: the interactive prompt is not shown`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is not shown`
  - Landing: `tests/e2e/steps/step_t018_chat_then_prompt_not_shown.go`
  - 語意：stdout 與 stderr 皆**不**帶 `cli.TUIHint` 字面（非終端輸入不得渲染 TUI）。

- [ ] T019 [P] [BDD-RED] `Then: the interactive prompt offers the recent prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt offers the recent prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t019_chat_then_offers_recent_prompt.go`
  - 語意：captured output（rendered prompt）含 `{prompt}` 作為建議（presence）。

- [ ] T020 [P] [BDD-RED] `Then: the interactive prompt offers the workspace entry "{entry}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt offers the workspace entry "{entry}"`
  - Landing: `tests/e2e/steps/step_t020_chat_then_offers_workspace_entry.go`
  - 語意：captured output 含 workspace entry `{entry}` 作為建議。

- [ ] T021 [P] [BDD-RED] `Then: the interactive prompt offers the available tool "{tool}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt offers the available tool "{tool}"`
  - Landing: `tests/e2e/steps/step_t021_chat_then_offers_tool.go`
  - 語意：captured output 含註冊工具名 `{tool}` 作為建議。

- [ ] T022 [P] [BDD-RED] `Then: the interactive prompt reports the active provider "{provider}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt reports the active provider "{provider}"`
  - Landing: `tests/e2e/steps/step_t022_chat_then_reports_provider.go`
  - 語意：captured output 含作用中 provider/model `{provider}`（dashboard）。

- [ ] T023 [P] [BDD-RED] `Then: the interactive prompt reports the session's token usage and turn count`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt reports the session's token usage and turn count`
  - Landing: `tests/e2e/steps/step_t023_chat_then_reports_tokens_turns.go`
  - 語意：captured output 帶 token 用量與 turn 數兩個 figure（dashboard）。

- [ ] T024 [P] [BDD-RED] `Then: the shared prompt log records the prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log records the prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t024_chat_then_log_records.go`
  - 語意：讀 `$TELL_ME_HOME/output/global_prompts.jsonl`，斷言其含 `{prompt}`。

- [ ] T025 [P] [BDD-RED] `Then: the shared prompt log still holds only "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log still holds only "{prompt}"`
  - Landing: `tests/e2e/steps/step_t025_chat_then_log_unchanged.go`
  - 語意：讀 `global_prompts.jsonl`，斷言恰只含 `{prompt}`（非 `-i` 執行未 append）。

- [ ] T026 [P] [BDD-RED] `Then: tellme sends no request to the provider "{provider}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme sends no request to the provider "{provider}"`
  - Landing: `tests/e2e/steps/step_t026_chat_then_no_request.go`
  - 語意：`{provider}` 的 fake 記錄 **0** 次請求（abort／非互動空提交不得觸及 provider）。

### UNIT（非 DSL 的單元斷言）

- [ ] T027 [P] [UNIT] 建議引擎：subsequence / dedupe / cap / 空查詢 / path-like gating / ignore 目錄
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 2；PR #38 review directive ③
    - `specs/truth/techstack.md` -> CLI Application（Prompt suggestion engine）、Testing & Verification（Pure-helper unit tests）
    - `internal/app/suggestions/service.go`
  - 撰寫：空查詢回 newest recent prompts；查詢以 subsequence 命中、dedupe、≤10；path-like 查詢收斂 `filepath.Split(query)`、分批讀取、跳過 `.git`／`node_modules`、收滿 10 即停；`ctx` 取消即時讓出。
  - 落點：`internal/app/suggestions/service_test.go`。

- [ ] T028 [P] [UNIT] 共享 store：append-only、newest-first dedupe 讀取、byte-identical round-trip、size-checked compaction、`Close`
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 3；PR #38 review directive ⑤
    - `specs/truth/data/data-model.dbml` -> `prompt_log_entry`
    - `internal/infrastructure/history/global_prompt_tracker.go`
  - 撰寫：`Append` 以 `O_APPEND|O_CREATE|O_WRONLY` 且不截斷；一行 `{"timestamp":"<RFC3339>","prompt":"<text>"}` byte-identical round-trip；`Recent` newest-first + dedupe；compaction 僅在 `newSize == initialSize` 才 `AtomicWrite`；`Close(ctx)` 排空背景工作。
  - 落點：`internal/infrastructure/history/global_prompt_tracker_test.go`。

- [ ] T029 [P] [UNIT] TUI 模型（注入 I/O）：seed / refresh / accept / submit-vs-newline-vs-abort / dashboard / 只寫 stderr
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 6；PR #38 review directive ①
    - `specs/truth/techstack.md` -> Testing & Verification（Interactive TUI prompt harness）
    - `internal/ui/tui/prompt/model.go`
  - 撰寫：以 scripted key 驅動模型——初始建議 seed、輸入後（debounce）refresh、`Tab`/`Shift+Tab` 接受／循環、`Ctrl+S`/`Alt+Enter` submit vs `Enter` newline vs `Esc`/`Ctrl+C` abort、dashboard 欄位；輸出只寫**注入的 stderr writer**，stdout 保持空。
  - 落點：`internal/ui/tui/prompt/model_test.go`。

- [ ] T030 [P] [UNIT] `-i`/`USE_TUI_PROMPT`/非 TTY gating 矩陣（經 `tuiPromptRunner` DI seam）
  - Read:
    - `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 4；PR #38 review directive ④
    - `specs/truth/techstack.md` -> Configuration（Interactive prompt enable）、CLI Application（Terminal detection）
    - `internal/cli/cli.go`
  - 撰寫：以替換 `tuiPromptRunner` 變數的方式，斷言 dispatch 矩陣——`-i` 且 stdin 為終端 → 走 TUI runner；否則（無旗標／非終端）→ 走既有 round-012 plain reader／piped 路徑，**不**呼叫 TUI runner。
  - 落點：`internal/cli/tui_dispatch_test.go`。

### Phase Review Gate

- [ ] T031 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/prompting-with-suggestions.feature`、`using-the-interactive-prompt.feature`、`recording-the-shared-prompt-log.feature`、`choosing-the-interactive-prompt.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
    - `internal/ui/tui/prompt/`、`internal/app/suggestions/`、`internal/infrastructure/history/global_prompt_tracker.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；16 個新句 stepdef + 4 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（TUI 未渲染、建議引擎未回結果、store 未寫入、gating 未接線）。

---

## Phase 4A: ADD Feature File - cli/chat/prompting-with-suggestions.feature

**Goal**: 讓 `prompting-with-suggestions.feature` 全綠 — `-i` 提示提供三來源建議（recent prompts 含共享紀錄、workspace paths 排除 noisy 目錄、registered tools），隨輸入更新並可以 keystroke 接受。

**Shared Must Read**:
- `specs/truth/features/cli/chat/prompting-with-suggestions.feature` -> `Feature: Prompting with live suggestions`
- `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log already holds "…"`、`the operator opens the interactive prompt and types "…"`、`the interactive prompt offers the recent prompt "…"`、`the interactive prompt offers the workspace entry "…"`、`the interactive prompt offers the available tool "…"`、`the working directory contains a file "…" whose text is "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/prompting-with-suggestions.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Interactive TUI prompt / Prompt suggestion engine / Shared global prompt log）
- `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 1, 2, 3, 6, 7
- `specs/plans/015-interactive-tui-prompt/ui/ui-plan.md` -> `10-suggestion` 畫面

**Boundary**:
- 產品碼：`internal/ui/tui/prompt/**`（Bubble Tea 模型；`tea.WithOutput(env.stderr)`／`tea.WithInput(env.stdin)` — review blocker①）、`internal/app/suggestions/service.go`（多來源引擎，bounded traversal — review directive③）、`internal/infrastructure/history/global_prompt_tracker.go` 的 `Recent` 讀取（newest-first + dedupe）。
- 建議只斷言 **presence**；非精確 ANSI bytes。
- `stdout` byte-exact；class-phrase 詞彙維持 **11**；exit-code 不變；POSIX-only。
- 不得引入 fuzzy/glob 新相依（research Decision 7）。

**Test Scope**:
- `specs/truth/features/cli/chat/prompting-with-suggestions.feature`

- [ ] T032 [BDD-GREEN] 讓 Test Scope 全綠（並使 T027 的建議引擎 `[UNIT]` 轉綠）
- [ ] T033 [BDD-REFACTOR] 在綠燈下整理建議來源聚合與 debounce 路徑

## Phase 4B: ADD Feature File - cli/chat/using-the-interactive-prompt.feature

**Goal**: 讓 `using-the-interactive-prompt.feature` 全綠 — submit 恰跑一次推理回合、abort 不送請求、開啟回報 session dashboard。

**Shared Must Read**:
- `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` -> `Feature: Using the interactive prompt`
- `specs/truth/features/cli/chat/dsl.md` -> `the operator submits the prompt "…" at the interactive prompt`、`the operator aborts the interactive prompt`、`the operator opens the interactive prompt`、`the interactive prompt is shown`、`the interactive prompt reports the active provider "…"`、`the interactive prompt reports the session's token usage and turn count`、`tellme sends exactly one request to the provider "…"`、`tellme sends no request to the provider "…"`、`tellme prints the provider's answer "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/using-the-interactive-prompt.feature`）
- `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 4, 5, 6
- `specs/plans/015-interactive-tui-prompt/ui/ui-plan.md` -> `entry` / `20-dashboard` 畫面

**Boundary**:
- 產品碼：`internal/cli/cli.go` 的 `-i` dispatch（submit → 既有單回合路徑；abort → 無請求、exit 0）、`internal/ui/tui/prompt/**` dashboard header（重用 provider/model + round-009 token figure + history turn count — 無新帳務，research Decision 5）。
- dashboard 只讀既有 session／metrics 狀態；不改 status line／estimator。
- `stdout` byte-exact；TUI 只寫 `env.stderr`；class-phrase 詞彙維持 **11**。

**Test Scope**:
- `specs/truth/features/cli/chat/using-the-interactive-prompt.feature`

- [ ] T034 [BDD-GREEN] 讓 Test Scope 全綠（並使 T029 的 TUI 模型 `[UNIT]` 轉綠）
- [ ] T035 [BDD-REFACTOR] 在綠燈下整理 submit／abort 與 dashboard 注入路徑

## Phase 4C: ADD Feature File - cli/chat/recording-the-shared-prompt-log.feature

**Goal**: 讓 `recording-the-shared-prompt-log.feature` 全綠 — `-i` 提交在共享提示紀錄 append 恰一行；one-shot 與 piped 執行不寫。

**Shared Must Read**:
- `specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature` -> `Feature: Recording the shared prompt log`
- `specs/truth/features/cli/chat/dsl.md` -> `the operator submits the prompt "…" at the interactive prompt`、`the operator starts tellme with the prompt "…"`、`the operator pipes "…" into tellme`、`the shared prompt log already holds "…"`、`the shared prompt log records the prompt "…"`、`the shared prompt log still holds only "…"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/recording-the-shared-prompt-log.feature`）+ `/axb-data-plan` ADD（`data-model.dbml` `prompt_log_entry`）+ `/axb-technical-research` MODIFY（`techstack.md` Shared global prompt log）
- `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 3；PR #38 review directive ⑤

**Boundary**:
- 產品碼：`internal/infrastructure/history/global_prompt_tracker.go`（`Append` 以 `O_APPEND|O_CREATE|O_WRONLY`；寫入**僅在 `-i`**；detached/bounded 不阻塞 prompt；async compaction size-snapshot-checked；`Close(ctx)` 排水）；`internal/cli/cli.go` 僅在 `-i` 提交時觸發寫入。
- 形狀 byte-identical `{"timestamp":"<RFC3339>","prompt":"<text>"}`；不取跨行程 `flock`（carried item）；位置為 `output/` **root**。
- 非 `-i`（positional／piped／non-interactive）一律不寫。

**Test Scope**:
- `specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature`

- [ ] T036 [BDD-GREEN] 讓 Test Scope 全綠（並使 T028 的 store `[UNIT]` 轉綠）
- [ ] T037 [BDD-REFACTOR] 在綠燈下整理 append／compaction 與 `Close` 排水路徑

## Phase 4D: ADD Feature File - cli/chat/choosing-the-interactive-prompt.feature

**Goal**: 讓 `choosing-the-interactive-prompt.feature` 全綠 — TUI 為 opt-in：無旗標的終端維持 round-012 plain reader；非終端輸入永不開啟 TUI。

**Shared Must Read**:
- `specs/truth/features/cli/chat/choosing-the-interactive-prompt.feature` -> `Feature: Choosing the interactive prompt`
- `specs/truth/features/cli/chat/dsl.md` -> `the operator pipes "…" into tellme`、`the operator pipes "…" into tellme with the interactive prompt enabled`、`the interactive prompt is not shown`、`the reading announcement is reported on the diagnostic output`、`the interactive prompt is shown`
- `specs/truth/features/cli/dsl.md` -> `the operator is working at an interactive terminal`（既有 root row；語意不變）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/choosing-the-interactive-prompt.feature`）
- `specs/plans/015-interactive-tui-prompt/research.md` -> Decision 4, 8

**Boundary**:
- 產品碼：`internal/config/config.go`（`USE_TUI_PROMPT`）、`internal/cli/cli.go`（`-i`/`--interactive` 解析；engage 僅當 `-i`/`USE_TUI_PROMPT` AND stdin 為終端；否則既有 round-012 plain reader／piped 路徑不變）；重用 round-012 real-isatty seam。
- 不得改 round-012 既有公告與非 `-i` 行為；class-phrase 詞彙維持 **11**；`stdout` byte-exact。
- POSIX-only；無 Windows 變體。

**Test Scope**:
- `specs/truth/features/cli/chat/choosing-the-interactive-prompt.feature`

- [ ] T038 [BDD-GREEN] 讓 Test Scope 全綠（並使 T030 的 gating `[UNIT]` 轉綠）
- [ ] T039 [BDD-REFACTOR] 在綠燈下整理 `-i`／`USE_TUI_PROMPT`／isatty gating 與 fallback 路徑

## Phase 4E: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、exit-code 表、offline paths、rounds 001–014 全綠），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T040 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（review blocker①）**：暫時讓 TUI runner 以預設 `os.Stdout` 輸出（或把 `tea.WithOutput` 指向 stdout），確認 `choosing-the-interactive-prompt`／`using-the-interactive-prompt` 的 `stdout` byte-exact 斷言（或 `the interactive prompt is not shown`）失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（共享紀錄）**：暫時停用 `-i` 提交的 append，確認 `recording-the-shared-prompt-log.feature` 的 `the shared prompt log records the prompt "…"` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact（TUI 只寫 `stderr`）；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；rounds 001–014 既有場景全綠；`go mod tidy` 後 module graph 僅新增 TUI 家族（bubbletea/bubbles；lipgloss 升 direct）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/data/data-model.dbml` -> `prompt_log_entry`（ADD） | T005、T007、T011、T024、T025、T028、T036 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Interactive TUI prompt (`-i`)） | T001、T002、T004、T008、T029、T032、T034 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Prompt suggestion engine） | T005、T006、T019–T021、T027、T032 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Shared global prompt log） | T005、T007、T011、T024、T025、T028、T036 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session dashboard） | T023、T034 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Terminal detection — reuses the isatty seam） | T002、T008、T018、T030、T038 | PASS |
| `specs/truth/techstack.md` -> CLI Application（CLI flag parsing — `-i`/`--interactive`） | T008、T030、T038 | PASS |
| `specs/truth/techstack.md` -> Configuration（Interactive prompt enable — `USE_TUI_PROMPT` + `-i`） | T008、T030、T038 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Interactive TUI prompt harness） | T006、T010、T029 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner / step definitions） | T003、T009、T012–T026、T031 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Local fake provider — recorded request） | T014、T016、T026、T034 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests） | T010、T027–T030 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（TUI libraries — introduced round 015） | T001 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001–T003、T032–T037 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T040 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` ADD（`specs/truth/data/data-model.dbml`） | T005、T007、T011、T024、T025、T028、T036 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/prompting-with-suggestions.feature`） | T019–T021、T032、T033 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/using-the-interactive-prompt.feature`） | T012、T014、T015、T017、T022、T023、T024、T026、T034、T035 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/recording-the-shared-prompt-log.feature`） | T011、T024、T025、T036、T037 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/choosing-the-interactive-prompt.feature`） | T016、T018、T038、T039 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — +16 rows） | T009、T011–T026、T031 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T040 驗證詞彙維持 11） | PASS |
| `research.md` -> Decision 1（adopt the Bubble Tea family） | T001、T004、T010 | PASS |
| `research.md` -> Decision 2（multi-source suggestion engine） | T005、T006、T027、T032 | PASS |
| `research.md` -> Decision 3（append-only shared log; `-i` only; compaction） | T005、T007、T011、T024、T025、T028、T036 | PASS |
| `research.md` -> Decision 4（opt-in; round-012 default; non-TTY fallback） | T008、T018、T030、T038 | PASS |
| `research.md` -> Decision 5（dashboard reuses existing state） | T023、T034 | PASS |
| `research.md` -> Decision 6（hermetic verification, no pty） | T003、T010、T029 | PASS |
| `research.md` -> Decision 7（dependency footprint — TUI family only） | T001、T040 | PASS |
| `research.md` -> Decision 8（POSIX-only） | T004（Boundary）、T038（Boundary）、T040 | PASS |
| `research.md` -> Decision 9（must-ask questions unchanged） | T040（不改 runner／端） | PASS |
| `spec.md` -> US1–US4、`FR-001`–`FR-015`、`NFR-001`–`NFR-005` | T011–T030（對齊）、T032–T039（交付）、T040 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/tui/prompt`、`internal/app/suggestions`、`internal/domain/suggestions`、`internal/infrastructure/history/global_prompt_tracker.go`、`internal/config`、`internal/cli`、`tests/e2e/**`） | T004–T010、T027–T030、T032–T039 | PASS |
| `plan.md` -> Scope notes（api NOOP；data ADD；UI reviewed；CLI end → `/axb-dsl-refine`） | T031、T040 | PASS |
| PR #38 review directives（① blocker TUI stream containment；② package/domain-boundary；③ suggestion I/O bounds；④ DI seam；⑤ async append safety） | ① T004、T008、T029、T032–T039 Boundary、T040 witness；② T005–T008；③ T006、T027、T032 Boundary；④ T008、T030；⑤ T007、T028、T036 Boundary | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
