# Tasks: Strict Visual Parity for the `-i` Interactive Prompt (round 016)

**Plan Package**: `specs/plans/016-interactive-prompt-visual-parity`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **本輪沒有新增技術**：`tellme` 的 TUI 家族（`bubbletea` + `bubbles/textarea` + direct `lipgloss`）已在 round 015 引入，本輪不動 `go.mod`。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（本輪 11 句為全新句 `ADD`；2 句為退役 `DELETE`）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/tui/prompt/{model,textarea,suggester}.go`（strict-parity chrome + debounced refresh + `Tab`-inserts + resize；**移除** dashboard 標頭）與 `internal/cli/cli.go`（移除 dashboard wiring；`-i` gating 不變）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN / CODE-REMOVE）」。
- 本輪**無跨模組**新句：root `specs/truth/features/cli/dsl.md` 為 **NOOP**（class-phrase 詞彙維持 **11**）；所有新句皆 `chat` 模組專屬。

## Round-016 locked decision + PR #40 review folds (implementation constraint — MUST)

> 來自 issue [#39](https://github.com/gosharplite/tellme/issues/39)（operator 拍板）與 PR [#40](https://github.com/gosharplite/tellme/pull/40) 的架構審查（[#5657665895](https://github.com/gosharplite/tellme/pull/40#issuecomment-5657665895)，verdict **PLAN + TRUTH APPROVED**）。

- **[STRICT PARITY] 提示面與 `tell-me-go` 一致** — `-i` 提示只呈現**有框線的多行編輯器**（`lipgloss.NormalBorder`、fg `240`、固定高度 10、寬度隨終端、reference placeholder、無行號）與在其下方的**有樣式建議清單**（`Suggestions:` 標頭；選中列 bold fg `205` / bg `235`、未選中 fg `245`）；**移除** dashboard 標頭、**移除**底部狀態／keybinding 列、**無** `?` help overlay（提示語在 placeholder 內）。見 T003、T009、T022、T023、T024。
- **[INTERACTION PARITY] 建議刷新與接受** — 建議刷新**去抖動（debounce）且可取消**（stale fetch 不得覆寫較新結果），>3 行的建議**不得提供**；`Tab`/`Shift+Tab` **插入**選中建議（多詞且建議為單一 token 時只換最後一個 token）。見 T003、T011、T013、T014、T016、T018、T020、T021。
- **[RESPONSIVE] 終端寬度** — 處理 `tea.WindowSizeMsg`，編輯器寬度 = `msg.Width - 4`；極窄寬度不得 panic。見 T003、T012、T017、T022、T023。
- **[NO REGRESSION] 單一來源** — `stdout` 維持 byte-exact（TUI 只寫注入的 `stderr`）；keybindings 不變；shared log、建議來源、非 TTY fallback、round-012 plain reader 不變；class-phrase 詞彙維持 **11**。見 T023、T024、T026。
- **[PR #40 F1 — acceptance-coverage]** — F1.3（多行保留 + 無行號）以新增 interface Rule 承接（T015/T017/T022）；F1.1（debounce timing）與 F1.2（resize/reflow + narrow degrade）為 **unit pin**，已在 `features/acceptance/**` 與 `chat/dsl.md` 明示標註。
- **[PR #40 F2 — last-token]** — 最後一個 token 置換分支為 **unit pin**（已在 acceptance + interface 標註；E2E accept 使用整行置換）。
- **[PR #40 F3 — negative predicate]** — `the interactive prompt shows no session metrics header` 以**具體樣式**（`provider .* | tokens .* | turns .*`）斷言；T026(a) 為非真空見證。

## Round-016 architect review directives (implementation constraints — MUST)

> 來自 PR [#40](https://github.com/gosharplite/tellme/pull/40) 的 Principal Software Architect 審查（[#5657823910](https://github.com/gosharplite/tellme/pull/40#issuecomment-5657823910)，evaluated head `e9f7083`，verdict **PLAN + TRUTH APPROVED WITH ARCHITECTURAL DIRECTIVES**）。以下為 `/axb-implement` 的實作約束，已寫入對應 task 的 `Read`／`Boundary`。

- **[TECHNICAL DEBT 1 — cancelable refresh] `prompt.Source` carries `ctx`** — 將 `Source.Suggest(query string) []string` 改為 `Suggest(ctx context.Context, query string) []string`；`prompt.Model` 在排定 debounce fetch 時以 `context.WithCancel(parent)` 建可取消子 context，新按鍵即 `cancel()` 前一查詢；`tuiSource`（`cli.go`）把 ctx 下傳至 `appsuggestions.Service.Suggest(ctx, query)`，使 `OSSWorkspace.Entries` 於 `ctx.Err() != nil` 立即讓出。見 T003、T018、T020。
- **[TECHNICAL DEBT 2 — event-loop fidelity] delegate editing to `textarea.Model.Update`** — 命令鍵（`Ctrl+S`/`Alt+Enter`/`Esc`/`Ctrl+C`/`Tab`/`Shift+Tab`）在外層攔截；其餘（Backspace/Delete/方向鍵/`Ctrl+A`/`Ctrl+E`/word movement）交由 `bubbles/textarea` 處理；僅在 `m.ed.value()` 變動時排 debounce。見 T003、T022。
- **[REFACTOR 1 — remove startup disk I/O] drop `store.Load()`** — 從 `defaultRunTUIPrompt` 移除 `prior, _ := store.Load()`，並移除 `tuiprompt.Run`/`New` 的 `dash Dashboard` 參數（dashboard 退役，`FR-004`）。見 T024。
- **[REFACTOR 2 — cursor init] `suggester.cursor` 預設 0** — 建議清單載入時游標為 `0`（不是 `-1`），對齊 Gherkin「marks one suggestion as the current choice」與 `ui/screens/entry.txt`。見 T003、T008。

### Implementation gate checklist（Phase 4 完成前逐項確認）

- [X] **Context lifecycle**：取消能終止 `appsuggestions.Service` 進行中的目錄迭代（`OSSWorkspace.Entries`）。
- [X] **Stream containment**：TUI 渲染只寫 `env.stderr`；`stdout` 未被觸碰。
- [X] **Startup optimization**：`defaultRunTUIPrompt` 已無 `store.Load()`。
- [X] **Dual falsifiability witnesses (T026)**：(a) 暫時重繪 metrics header → `no session metrics header` 失敗；(b) 暫時讓 `Tab` 只循環 → `holds the accepted suggestion` 失敗。
- [X] **Quality gates**：`make verify` 乾淨（0 lint、0 test-sleep、0 vulns）。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）與產品碼落點（reference chrome tokens + debounced refresh seam），讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫 chrome／debounce／insert 行為。

- [X] T001 建立 13 個 stepdef 獨立檔骨架（11 新句 + 2 退役句；Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 11 個新句 + 2 個退役句）
    - `tests/e2e/steps/register.go`、`tests/e2e/harness/`（`TELL_ME_FORCE_STDIN_TTY` seam + scripted stdin）
  - 只做：建立 13 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t004_chat_remove_reports_provider.go`
    - `tests/e2e/steps/step_t005_chat_remove_reports_tokens_turns.go`
    - `tests/e2e/steps/step_t006_chat_then_framed_editor.go`
    - `tests/e2e/steps/step_t007_chat_then_suggestions_beneath.go`
    - `tests/e2e/steps/step_t008_chat_then_marks_current_choice.go`
    - `tests/e2e/steps/step_t009_chat_then_no_metrics_header.go`
    - `tests/e2e/steps/step_t010_chat_then_advertises_keys.go`
    - `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go`
    - `tests/e2e/steps/step_t012_chat_then_frames_within_width.go`
    - `tests/e2e/steps/step_t013_chat_then_holds_accepted.go`
    - `tests/e2e/steps/step_t014_chat_then_no_overlong_suggestion.go`
    - `tests/e2e/steps/step_t015_chat_then_no_line_numbers.go`
    - `tests/e2e/steps/step_t016_chat_when_accept_suggestion.go`
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [X] T002 建立 2 個 `[UNIT]` 落點檔骨架
  - Read:
    - `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 1, Decision 2, Decision 5, Decision 6
    - `specs/truth/techstack.md` -> Testing & Verification（Interactive TUI prompt harness）
  - 只做：預留 2 個 `[UNIT]` 空殼——`internal/ui/tui/prompt/model_chrome_test.go`（chrome 存在／dashboard 不存在／resize 寬度／`Tab` 插入含 last-token）、`internal/ui/tui/prompt/refresh_test.go`（debounce + cancel + >3 行 drop）。採 stdlib `testing`。
  - 不做：不寫任何斷言或 fixture 語意。

- [X] T003 落點 `internal/ui/tui/prompt/` 的 reference chrome tokens 與 debounced refresh seam 骨架
  - Read:
    - `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 1, Decision 2, Decision 3, Decision 5
    - `specs/plans/016-interactive-prompt-visual-parity/ui/ui-plan.md`（terminal 視覺方向；frame fidelity；`entry` / `10-suggestion` / `20-composed` / `30-narrow` frames）
    - `internal/ui/tui/prompt/{model,textarea,suggester}.go`（現況）
  - 只做：在 `internal/ui/tui/prompt/` 建立 **chrome tokens 落點**（editor border/placeholder/固定高度、root padding、`Suggestions:` 標頭 + 選中／未選中樣式的常數與套用函式殼）與 **debounced／cancelable refresh seam 骨架**（`tea.Tick` + `context` cancel 的 hook 留位、>3 行過濾掛點）。可為新檔（例如 `styles.go`、`refresh.go`）或於既有 `model.go`/`suggester.go`/`textarea.go` 留位。
  - 不做：不寫 chrome 渲染結果、不寫 debounce／insert 行為、不寫 resize 邏輯、不移除 dashboard（該動作為 T024）。
  - 架構指令（#5657823910）：落點即帶 `Source.Suggest(ctx, query)` 簽章骨架、`editor.Update` 委派 `textarea.Model.Update` 的落點、`suggester.cursor` 預設 `0` 的落點。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 11 句為全新句（`ADD`），2 句為退役句（`DELETE`）。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪新句皆 `chat` 模組專屬；features 僅重用其既有 root row，語意不變且 stepdef 已在）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪 chrome Then（`framed`／`lists suggestions beneath`／`marks one suggestion …`／`no session metrics header`／`advertises … keys`／`keeps the typed text`／`shows no line numbers`／`frames the editor within the terminal width`）依 `dsl.md` 是對「captured output 之 **presence／absence**」斷言（非精確 ANSI bytes）；`no session metrics header` 用具體 dashboard 樣式（非裸 substring）。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：本輪 **2 句**（退役的 dashboard Thens）。
- `[BDD-RED]`：本輪 **11 句**（`ADD` 的 DSL rows）。
- `[UNIT]`：2 個非 DSL 單元斷言（TUI chrome/resize/insert；debounce/cancel/drop）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the interactive prompt is framed around a multi-line editor`
  -> `the interactive prompt lists suggestions beneath the editor`
  -> `the interactive prompt marks one suggestion as the current choice`
  -> `the interactive prompt shows no session metrics header`
  -> `the interactive prompt advertises the submit and abort keys`
  -> `the interactive prompt keeps the typed text "{text}"`
  -> `the interactive prompt frames the editor within the terminal width`
  -> `the interactive prompt holds the accepted suggestion "{text}"`
  -> `no suggestion offered to the operator spans more than three lines`
  -> `the interactive prompt shows no line numbers`
  -> `the operator opens the interactive prompt, types "{query}", and accepts the current suggestion`
  -> `the interactive prompt reports the active provider "{provider}"`（退役）
  -> `the interactive prompt reports the session's token usage and turn count`（退役）
- `specs/truth/features/cli/dsl.md` -> **NOOP**（root 詞彙不變，維持 11 片語）
- `truth-delta.md` -> `/axb-dsl-refine` DELETE（dashboard Thens）+ ADD（`chat/presenting-the-interactive-prompt.feature`）+ MODIFY（`chat/prompting-with-suggestions.feature`、`chat/dsl.md`）；NOOP（root `cli/dsl.md`、`contracts/**`、`data/**`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/ui/tui/prompt/model_chrome_test.go`、`internal/ui/tui/prompt/refresh_test.go`。
- 不寫產品碼（除 T001–T003 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T018 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；目標檔案互斥，符合 Zero Shared Edits）；T019 等全部回來再啟動 subagent 執行 review。

### BDD-REMOVE（退役句型）

- [X] T004 [P] [BDD-REMOVE] `Then: the interactive prompt reports the active provider "{provider}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該列已刪（退役）
  - Landing: `tests/e2e/steps/step_t004_chat_remove_reports_provider.go`
  - 語意：移除綁這句的既有 stepdef／assertion（若其註冊於既有共用檔，改寫為不再註冊此句），不得留下保護 dashboard 的測試。

- [X] T005 [P] [BDD-REMOVE] `Then: the interactive prompt reports the session's token usage and turn count`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該列已刪（退役）
  - Landing: `tests/e2e/steps/step_t005_chat_remove_reports_tokens_turns.go`
  - 語意：同上——移除綁這句的既有 stepdef／assertion。

### BDD-RED（本輪新增句型）

- [X] T006 [P] [BDD-RED] `Then: the interactive prompt is framed around a multi-line editor`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is framed around a multi-line editor`
  - Landing: `tests/e2e/steps/step_t006_chat_then_framed_editor.go`
  - 語意：captured output 於編輯器外畫出框線（存在編輯器 frame 的水平邊框線）。

- [X] T007 [P] [BDD-RED] `Then: the interactive prompt lists suggestions beneath the editor`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt lists suggestions beneath the editor`
  - Landing: `tests/e2e/steps/step_t007_chat_then_suggestions_beneath.go`
  - 語意：captured output 在編輯器 frame **下方**帶 `Suggestions:` 標頭。

- [X] T008 [P] [BDD-RED] `Then: the interactive prompt marks one suggestion as the current choice`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt marks one suggestion as the current choice`
  - Landing: `tests/e2e/steps/step_t008_chat_then_marks_current_choice.go`
  - 語意：captured output 恰有**一列**建議被標為當前選項（`>` cursor）。
  - 架構指令（#5657823910）：建議載入時 `suggester.cursor` 預設 `0`（非 -1），使恰有一列被標為當前選項。

- [X] T009 [P] [BDD-RED] `Then: the interactive prompt shows no session metrics header`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt shows no session metrics header`
  - Landing: `tests/e2e/steps/step_t009_chat_then_no_metrics_header.go`
  - 語意：captured output **不**帶 session metrics 標頭——即**無**符合具體樣式 `provider .* | tokens .* | turns .*` 的列（round-015 dashboard 格式）——strict parity；**不**用裸 `tokens`／`turns` substring。

- [X] T010 [P] [BDD-RED] `Then: the interactive prompt advertises the submit and abort keys`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt advertises the submit and abort keys`
  - Landing: `tests/e2e/steps/step_t010_chat_then_advertises_keys.go`
  - 語意：captured output 帶編輯器 placeholder，命名 submit key（`Ctrl+S`）與 abort key（`Esc`）。

- [X] T011 [P] [BDD-RED] `Then: the interactive prompt keeps the typed text "{text}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt keeps the typed text "{text}"`
  - Landing: `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go`
  - 語意：captured output 於編輯器內帶所打字串 `{text}`（多行值以 `\n` escape 表示）。

- [X] T012 [P] [BDD-RED] `Then: the interactive prompt frames the editor within the terminal width`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt frames the editor within the terminal width`
  - Landing: `tests/e2e/steps/step_t012_chat_then_frames_within_width.go`
  - 語意：captured output 的編輯器 frame 落在終端寬度內（邊框不超出該次執行寬度）。

- [X] T013 [P] [BDD-RED] `Then: the interactive prompt holds the accepted suggestion "{text}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt holds the accepted suggestion "{text}"`
  - Landing: `tests/e2e/steps/step_t013_chat_then_holds_accepted.go`
  - 語意：captured output 於編輯器內帶 `{text}`（接受的建議已插入）。

- [X] T014 [P] [BDD-RED] `Then: no suggestion offered to the operator spans more than three lines`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `no suggestion offered to the operator spans more than three lines`
  - Landing: `tests/e2e/steps/step_t014_chat_then_no_overlong_suggestion.go`
  - 語意：captured output 每一列建議至多三行。

- [X] T015 [P] [BDD-RED] `Then: the interactive prompt shows no line numbers`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt shows no line numbers`
  - Landing: `tests/e2e/steps/step_t015_chat_then_no_line_numbers.go`
  - 語意：captured output 的編輯器**不**繪出逐行行號欄（無前導行號欄）。

- [X] T016 [P] [BDD-RED] `When: the operator opens the interactive prompt, types "{query}", and accepts the current suggestion`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator opens the interactive prompt, types "{query}", and accepts the current suggestion`
  - Landing: `tests/e2e/steps/step_t016_chat_when_accept_suggestion.go`
  - 語意：以 `TELL_ME_FORCE_STDIN_TTY=1` 與 scripted key（open → type `{query}` → `Tab` accept → abort）跑 `tellme -i`，captured 結果。

### UNIT（非 DSL 的單元斷言）

- [X] T017 [P] [UNIT] TUI 模型：chrome 存在／無 dashboard／resize 寬度／`Tab` 插入（含 last-token）
  - Read:
    - `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 1, Decision 3, Decision 5, Decision 6
    - `specs/truth/techstack.md` -> CLI Application（Interactive TUI prompt (`-i`)；Session dashboard — removed）、Testing & Verification（Interactive TUI prompt harness）
    - `internal/ui/tui/prompt/model.go`、`textarea.go`、`suggester.go`
  - 撰寫：以注入 I/O 直接驅動模型，斷言——渲染含 editor 框線／placeholder／`Suggestions:` 標頭／`>` cursor；渲染**不含** dashboard 列、**不含**行號欄；`tea.WindowSizeMsg` 設編輯器寬度為 `msg.Width - 4`；`Tab`/`Shift+Tab` **插入**選中建議，且多詞輸入 + 單一 token 建議只換最後一個 token（F2 unit pin）。
  - 落點：`internal/ui/tui/prompt/model_chrome_test.go`。

- [X] T018 [P] [UNIT] 建議刷新：debounce + cancel（stale 丟棄）+ >3 行 drop
  - Read:
    - `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 2, Decision 6
    - `specs/truth/techstack.md` -> CLI Application（Prompt suggestion engine）
    - `internal/ui/tui/prompt/model.go`
  - 撰寫：查詢變更後刷新經 debounce（不每次按鍵即刷新；F1.1 unit pin）；被較新按鍵取代的 fetch 結果被丟棄（不覆寫較新清單）；>3 行的建議被過濾。
  - 落點：`internal/ui/tui/prompt/refresh_test.go`。
  - 架構指令（#5657823910）：取消須下傳至 `appsuggestions.Service`（`OSSWorkspace.Entries` 於 `ctx.Err() != nil` 讓出），非只在 model 丟棄 stale 回傳值。

### Phase Review Gate

- [X] T019 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature`、`prompting-with-suggestions.feature`、`using-the-interactive-prompt.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
    - `internal/ui/tui/prompt/`
  - 檢驗：無 undefined step；11 個新句 stepdef + 2 個 `[UNIT]` 皆已註冊；退役 2 句的 stepdef 已移除；測試可編譯；失敗僅限尚未實作之產品行為（chrome 未渲染、dashboard 未移除、debounce／insert 未實作、resize 未接線）。

---

## Phase 4A: MODIFY Feature File - cli/chat/prompting-with-suggestions.feature

**Goal**: 讓 `prompting-with-suggestions.feature` 全綠 — 額外涵蓋「接受建議會插入編輯器」與「>3 行建議不提供」，維持既有三來源建議。

**Shared Must Read**:
- `specs/truth/features/cli/chat/prompting-with-suggestions.feature` -> `Feature: Prompting with live suggestions`
- `specs/truth/features/cli/chat/dsl.md` -> `the operator opens the interactive prompt, types "…", and accepts the current suggestion`、`the interactive prompt holds the accepted suggestion "…"`、`no suggestion offered to the operator spans more than three lines`、`the shared prompt log already holds "…"`、`the interactive prompt offers the recent prompt "…"`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/prompting-with-suggestions.feature`、`chat/dsl.md`）
- `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 2, Decision 3
- `specs/plans/016-interactive-prompt-visual-parity/ui/ui-plan.md` -> `10-suggestion` 畫面

**Boundary**:
- 產品碼：`internal/ui/tui/prompt/model.go`（`Tab`/`Shift+Tab` **插入**選中建議，last-token 規則）、`internal/ui/tui/prompt/refresh.go` 或 `model.go`（debounce + cancel + >3 行 drop）、`suggester.go`（`>` cursor 套用）。
- 架構指令（#5657823910）：debounce fetch 以 `context.WithCancel` 包裝並把 `ctx` 經 `Source.Suggest(ctx, query)` 下傳（T018/T020）；非命令鍵編輯委派 `textarea.Model.Update`（T022）。
- 建議只斷言 **presence**；非精確 ANSI bytes；`stdout` byte-exact；class-phrase 詞彙維持 **11**；不得引入 fuzzy/glob 新相依。

**Test Scope**:
- `specs/truth/features/cli/chat/prompting-with-suggestions.feature`

- [X] T020 [BDD-GREEN] 讓 Test Scope 全綠（並使 T018 的刷新 `[UNIT]` 轉綠）
- [X] T021 [BDD-REFACTOR] 在綠燈下整理建議插入與 debounce／cancel 路徑

## Phase 4B: ADD Feature File - cli/chat/presenting-the-interactive-prompt.feature

**Goal**: 讓 `presenting-the-interactive-prompt.feature` 全綠 — `-i` 提示呈現 reference chrome（有框線編輯器 + 其下建議清單 + 單一 `>` cursor + 無 metrics 標頭 + placeholder 提示鍵 + 保留多行文字 + 無行號 + frame 在終端寬度內）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` -> `Feature: Presenting the interactive prompt`
- `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is framed around a multi-line editor`、`the interactive prompt lists suggestions beneath the editor`、`the interactive prompt marks one suggestion as the current choice`、`the interactive prompt shows no session metrics header`、`the interactive prompt advertises the submit and abort keys`、`the interactive prompt keeps the typed text "…"`、`the interactive prompt shows no line numbers`、`the interactive prompt frames the editor within the terminal width`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-interactive-prompt.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Interactive TUI prompt）
- `specs/plans/016-interactive-prompt-visual-parity/research.md` -> Decision 1, Decision 4, Decision 5
- `specs/plans/016-interactive-prompt-visual-parity/ui/ui-plan.md` -> `entry` / `20-composed` / `30-narrow` 畫面

**Boundary**:
- 產品碼：`internal/ui/tui/prompt/textarea.go`（`lipgloss.NormalBorder` + fg `240` + 固定高度 10 + reference placeholder + 無行號）、`model.go`（root padding；`tea.WindowSizeMsg` → 寬度 = `msg.Width - 4`；View 只組編輯器 + 建議清單）、`suggester.go`（`Suggestions:` 標頭 + 選中／未選中樣式）。
- 架構指令（#5657823910）：`editor.Update` 委派 `textarea.Model.Update`（Backspace/Delete/方向鍵/`Ctrl+A`/`Ctrl+E` 可運作，僅攔截命令鍵 — T022）；`suggester.cursor` 載入時預設 `0`（T008）。
- **不得**渲染 dashboard 標頭、底部 status／keybinding 列或 `?` overlay；`stdout` byte-exact（TUI 只寫注入 `stderr`）；class-phrase 詞彙維持 **11**。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature`

- [X] T022 [BDD-GREEN] 讓 Test Scope 全綠（並使 T017 的 chrome `[UNIT]` 轉綠）
- [X] T023 [BDD-REFACTOR] 在綠燈下整理 chrome tokens 與 resize 路徑

## Phase 4C: DELETE DSL Truth - retire the `-i` session dashboard rule

**Goal**: 移除 `-i` 提示面的 session dashboard 標頭（strict parity），並確認 `using-the-interactive-prompt.feature` 在無 dashboard 下仍成立。（Truth 層已於 PR #40 前的 dsl-refine 完成：`using-the-interactive-prompt.feature` 的 dashboard Rule/Example 已刪；本 phase 只退役產品行為。）

**Shared Must Read**:
- `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` -> `Feature: Using the interactive prompt`
- `specs/truth/features/cli/chat/dsl.md` -> 退役列（已刪）
- `truth-delta.md` -> `/axb-dsl-refine` DELETE（dashboard Thens）
- `internal/ui/tui/prompt/model.go` -> the `Dashboard` header render + wiring（obsolete）

**Boundary**:
- 產品碼：從 `internal/ui/tui/prompt/model.go` 移除 dashboard 標頭的渲染與 `Dashboard` 注入路徑；`internal/cli/cli.go` 移除 dashboard wiring。
- 架構指令（#5657823910，T024）：移除 `defaultRunTUIPrompt` 的 `prior, _ := store.Load()`，並移除 `tuiprompt.Run`/`New` 的 `dash Dashboard` 參數（無啟動期磁碟 I/O）。
- 不得改動 submit／abort、keybindings、shared log、建議來源、非 TTY fallback 或 round-012 plain reader；`stdout` byte-exact；class-phrase 詞彙維持 **11**。

**Test Scope**:
- `specs/truth/features/cli/chat/using-the-interactive-prompt.feature`

- [X] T024 [CODE-REMOVE] 移除 dashboard 標頭的產品分支與 `Dashboard` 注入路徑
- [X] T025 [REGRESSION] 跑 Test Scope，確認新版 truth 成立（無 dashboard）且 submit／abort 不受影響

## Phase 4D: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、exit-code 表、offline paths、rounds 001–015 全綠），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T026 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（chrome／dashboard；非真空）**：暫時讓 `model.go` 的 View 再次渲染 dashboard 標頭列（`provider … | tokens … | turns …`），確認 `presenting-the-interactive-prompt.feature` 的 `the interactive prompt shows no session metrics header` 失敗；觀察到失敗即還原（證明此 Then 非真空）。
  - **可偽性見證 (b)（`Tab` 插入）**：暫時讓 `Tab` 只循環不插入，確認 `prompting-with-suggestions.feature` 的 `the interactive prompt holds the accepted suggestion "…"` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact（TUI 只寫 `stderr`）；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；rounds 001–015 既有場景全綠；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Interactive TUI prompt (`-i`) — strict-parity chrome） | T003、T017、T022、T023 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session dashboard — removed from `-i`) | T004、T005、T009、T017、T024 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Prompt suggestion engine — debounce/cancel/drop/insert） | T003、T014、T018、T020 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Interactive TUI prompt harness — style/state assertions） | T002、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T003、T017、T018、T020–T024 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T026 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` DELETE（`chat/using-the-interactive-prompt.feature` dashboard rule） | T004、T005、T024、T025 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-interactive-prompt.feature`） | T006–T012、T015、T022、T023 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/prompting-with-suggestions.feature`） | T013、T014、T016、T020、T021 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — +11 rows / −2 rows） | T001、T004–T016、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T026 驗證詞彙維持 11） | PASS |
| `research.md` -> Decision 1（reproduce chrome on the existing family; no new dep） | T003、T017、T022、T023、T026 | PASS |
| `research.md` -> Decision 2（debounced/async/cancelable refresh + >3-line drop） | T003、T014、T018、T020 | PASS |
| `research.md` -> Decision 3（`Tab`/`Shift+Tab` insert; last-token heuristic） | T011、T013、T016、T017、T020 | PASS |
| `research.md` -> Decision 4（remove the dashboard header — strict parity） | T009、T017、T024 | PASS |
| `research.md` -> Decision 5（terminal-width handling + graceful narrow） | T003、T012、T017、T023 | PASS |
| `research.md` -> Decision 6（hermetic no-pty verification; style/state assertions） | T002、T017、T018、T019 | PASS |
| `research.md` -> Decision 7（no new dependency; POSIX-only; keybindings unchanged） | T003（Boundary）、T023（Boundary）、T026 | PASS |
| `research.md` -> Decision 8（must-ask questions unchanged） | T026（不改 runner／端） | PASS |
| `spec.md` -> US1–US3、`FR-001`–`FR-011`、`NFR-001`–`NFR-004` | T004–T018（對齊）、T020–T025（交付）、T026 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/tui/prompt/**`、`internal/cli/cli.go`；no new pkg/module） | T003、T017、T018、T020–T024 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；UI reviewed；CLI end → `/axb-dsl-refine`） | T019、T026 | PASS |
| `ui/ui-plan.md` + `ui/screens/*.txt`（terminal-mode PM artifact；fixed-height/width-driven frames） | T003、T022、T023 | PASS |
| Round-016 locked decision (#39) + PR #40 folds (F1 acceptance-coverage / F2 last-token unit pin / F3 specific predicate) | chrome → T003/T017/T022/T023/T024；interaction → T003/T011–T018/T020/T021；responsive → T003/T012/T017/T023；no-regression → T023/T024/T026；F1 → T015/T017/T022（+annotations in acceptance + `chat/dsl.md`）；F2 → T017（+annotations）；F3 → T009/T026(a) | PASS |
| PR #40 architect directives (#5657823910): ctx-cancelable `Source` · `textarea.Update` delegation · `store.Load` removal · cursor=0 · gate checklist | D1 → T003/T018/T020；D2 → T003/T022；D3 → T024；D4 → T003/T008；gate checklist → T026 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
