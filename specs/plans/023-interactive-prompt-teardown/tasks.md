# Tasks: Interactive Prompt Teardown & Submit-Surface Parity (round 023)

**Plan Package**: `specs/plans/023-interactive-prompt-teardown`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

> **Operator-locked decisions**: Q1 → 1a (clear the frame on submit/abort) · Q2 → 2a (resume the standard turn surface) · Q3 → A′ (echo the prompt, keep tellme's single captured line). **Round-022 ripple (Option 1)**: the round-022 negative `the tool loop added no blank line before the answer` is re-anchored to `the pre-flight payload line is separated from the answer by a single blank line`, because round 023 makes the `-i` submit a **chrome** surface (the round-017 frame gap is now the single blank).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 1–8 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證任務不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib + 既有 Bubble Tea family；`go.mod`／`go.sum` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立落點骨架、共用元件與 helper；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/tui/prompt/model.go`（提交/中止時清除編輯框）、`internal/cli/cli.go`（`-i` submit 取 chrome 旗標 + 於 acknowledgement 前 echo 提交的 prompt）。維持「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**`stdout` 維持 byte-exact；class-phrase 詞彙不變（11）。**

---

## Phase 2: Foundational

**Goal**: 建立本輪 stepdef／`[UNIT]` 落點骨架與產品碼落點（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔案或 seam。只建立落點與載體，不寫行為。

- [ ] T001 建立 stepdef 落點骨架（獨立檔、`init()` 自我註冊）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> 本輪 3 個新句 + 1 個改寫句
    - `tests/e2e/steps/register.go`、`tests/e2e/steps/scenario_context.go`、`tests/e2e/steps/tui_keys.go`
  - 只做：建立獨立 stepdef landing 檔 — `tests/e2e/steps/step_r023_t004_chat_then_prompt_echoed.go`、`..._t005_chat_then_prompt_not_echoed.go`、`..._t006_chat_then_prompt_cleared.go`；於 `step_r022_t009_chat_then_tool_no_blank.go` 建立/預留改寫落點（對到 T003）。每檔註冊空白 registrar。
  - 不做：不寫 arrange／斷言邏輯；不碰其他既有 step 檔。

- [ ] T002 建立產品碼落點骨架
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Interactive TUI prompt (`-i`)；Turn chrome；Turn progress spinner）
    - `internal/ui/tui/prompt/model.go`、`internal/cli/cli.go`（`runTUIPrompt`、`runTurn`、`turnOptions`）
  - 只做：在 `internal/ui/tui/prompt/model.go` 預留「submitted/aborted ⇒ frame cleared」的落點（`View()` 的分支 + 註解，實作留待 T010）；在 `internal/cli/cli.go` 的 `turnOptions` 新增 `echo bool` 欄位（空殼；**僅** `-i` 提交為 `true`，PR #51 principal review 🔵 REFACTOR）與 `runTUIPrompt` 傳 `chrome: true, echo: true` 的註解（實作留待 T010）；建立 `internal/ui/tui/prompt/model_teardown_test.go` 空殼。
  - 不做：不實作清除行為、不輸出 echo、不改任何輸出。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth：先改寫一個 MODIFY 句（round-022 negative 重錨），再為 3 個新句建立失敗訊號；最後補非 DSL 的 `[UNIT]`。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 本輪為 **MODIFY**（`the run shows no turn chrome` / `the run shows no progress spinner` 的 `不該發生` 去掉 `-i` clause）——這兩個 root row 仍被 `diagnostics`／`history` 模組使用，**不需新增 stepdef**（其 stepdef 沿用，只改語意）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`／`權威狀態落地`；Then 讀 `必查`。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：**1 句**（round-022 negative 重錨句）。
- `[BDD-RED]`：**3 句**（新增的 echo／not-echo／cleared Then）。
- `[UNIT]`：2 個（model teardown；CLI `-i` echo/chrome wiring）（非 DSL）。
- 三者都只動測試層，不寫產品碼（除 T002 已預留骨架）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> `the pre-flight payload line is separated from the answer by a single blank line`、`the interactive prompt is cleared`、`the submitted prompt "…" is echoed on the diagnostic output`、`the prompt is not echoed on the diagnostic output`
- `truth-delta.md` -> `/axb-dsl-refine` 的 ADD/MODIFY rows
- `tests/e2e/steps/tui_keys.go`、`tests/e2e/steps/tui_chrome.go`、`tests/e2e/steps/scenario_context.go`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點採 stdlib `testing`；`internal/ui/tui/prompt/model_teardown_test.go`、`internal/cli/cli_test.go`。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T008 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；T003 觸及的 `step_r022_t009_*` 與其他 task 不同檔）；T009 等全部回來再啟動 subagent 執行 review。

- [ ] T003 [P] [BDD-ALIGN] `Then: the pre-flight payload line is separated from the answer by a single blank line`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（merged capture；恰一空行）
  - Landing: `tests/e2e/steps/step_r022_t009_chat_then_tool_no_blank.go`
  - 動作：改寫 round-022 的 `thenToolNoBlank`，於 merged (`stdout`+`stderr`) capture 中斷言 pre-flight payload 行與答案位元組間**恰**一個空行（非「答案前不得有空行」）；round-022 的 positive separation 句不變。

- [ ] T004 [P] [BDD-RED] `Then: the submitted prompt "{prompt}" is echoed on the diagnostic output`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r023_t004_chat_then_prompt_echoed.go`
  - 語意：`stderr` 在 input-capture acknowledgement 之前，以**獨立區塊**逐字帶提交的 `{prompt}`（`fmt.Fprintln(env.stderr, prompt)`；**不得折疊**內嵌換行，符合 FR-009）；echo 不得落在 `stdout`。

- [ ] T005 [P] [BDD-RED] `Then: the prompt is not echoed on the diagnostic output`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then
  - Landing: `tests/e2e/steps/step_r023_t005_chat_then_prompt_not_echoed.go`
  - 語意：positional／Ctrl+D 表面不在 acknowledgement 前單獨重印 prompt。

- [ ] T006 [P] [BDD-RED] `Then: the interactive prompt is cleared`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> 該 Then（讀法：以終端還原，套用 renderer 的 teardown clear）
  - Landing: `tests/e2e/steps/step_r023_t006_chat_then_prompt_cleared.go`
  - 語意：以終端讀法還原 merged capture 後，編輯框（`┌…└`）**沒有**倖存於 submit/abort 之後；另由 T007 的 `[UNIT]` 釘住模型清除。**不得**用天真的 `!strings.Contains(sc.stderr, "┌")` — bubbletea 的 erase 序列（carriage-return 與 line-erase）不會移除先前的 draw bytes，raw buffer 仍留著編輯期 bytes；須先套用終端還原（處理 carriage-return 與行清除序列，參照 `step_t017_root_then_explains_stderr.go` 的讀法），並以 T007 的模型層 `m.View() == ""` 為權威釘（PR #51 principal review TD2）。

- [ ] T007 [P] [UNIT] 模型於 submit/abort 清除 frame（`internal/ui/tui/prompt`）
  - Read:
    - `specs/plans/023-interactive-prompt-teardown/research.md` -> Decision 1, 5
    - `specs/truth/techstack.md` -> Interactive TUI prompt (`-i`)；Interactive TUI prompt harness
    - `internal/ui/tui/prompt/model.go`
  - 撰寫：以 scripted keys 驅動模型，斷言 submit（`Ctrl+S`）與 abort（`Esc`）後 `View()` 為空；並斷言 submit 前 `View()` 非空。
  - 落點：`internal/ui/tui/prompt/model_teardown_test.go`。

- [ ] T008 [P] [UNIT] `-i` submit 的 echo/chrome wiring（`internal/cli`）
  - Read:
    - `specs/plans/023-interactive-prompt-teardown/research.md` -> Decision 2, 3, 4, 6
    - `internal/cli/cli.go`（`runTUIPrompt`、`runTurn`、`turnOptions`）、`internal/cli/cli_test.go`
  - 撰寫：注入 fake `tuiPromptRunner`，斷言 `-i` submit 後 `renderTurn` 收到 `chrome:true`（標準 chrome 會輸出）且 echo 於 acknowledgement 前輸出到 `stderr`；positional 表面 `echo=false`。
  - 落點：`internal/cli/cli_test.go`。

- [ ] T009 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature`、`specs/truth/features/cli/chat/dsl.md`
    - `tests/e2e/steps/step_r023_t0{04,05,06}_*.go`、`tests/e2e/steps/step_r022_t009_*.go`
    - `internal/ui/tui/prompt/*`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；1 個 ALIGN + 3 個新句 stepdef 已註冊；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: ADD Feature File - cli/chat/continuing-the-interactive-prompt.feature

**Goal**: 讓 `continuing-the-interactive-prompt.feature` 全綠 — `-i` submit/abort 清除編輯框；提交後 prompt 被 echo、回合以標準 chrome 開場並（終端 `stderr`、非 `-r`）顯示 spinner；positional 表面不 echo。

**Shared Must Read**:
- `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `the interactive prompt is cleared`、`the submitted prompt "…" is echoed on the diagnostic output`、`the prompt is not echoed on the diagnostic output`、`the input capture is announced for the turn`、`the turn opens with a horizontal rule`、`the turn is headed "Turn N" for the active mode`、`the run shows the progress spinner while it waits`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`continuing-the-interactive-prompt.feature`）+ `/axb-technical-research` MODIFY
- `specs/plans/023-interactive-prompt-teardown/research.md` -> Decision 1–6

**Boundary**:
- 產品碼：`internal/ui/tui/prompt/model.go` — 提交/中止時 `View()` 回空（清除編輯框），其餘 chrome（borders、suggestions、placeholder）不變；`internal/cli/cli.go` — `runTUIPrompt` 以 `turnOptions{raw: opts.raw, chrome: true, echo: true}` 呼叫 `renderTurn`，且 `runTurn` 在 `emitInputCaptured` 之前以 `fmt.Fprintln(env.stderr, prompt)` 把提交的 prompt **逐字**寫入 `env.stderr`（僅當 `opts.echo`；**不得折疊**內嵌換行，符合 FR-009）。**positional／Ctrl+D 路徑維持 `echo:false`。**
- 不得改：opt-in gating、suggestion engine、共享 prompt log、provider request、one-turn 契約、exit codes、class-phrase 詞彙（11）；`stdout` byte-exact；`internal/ui` 的 chrome/spinner/status emitters 沿用（不重寫）。
- **E2E witness 重製（research D5）**：round-016「final rendered frame is captured」前提失效；本 phase 以 cleared `View()` 的 `[UNIT]` pin（T007）+ 標準 surface 的 E2E Thens 取代；不得為舊前提保留 always-render。
- 保留 `TELL_ME_FORCE_STDIN_TTY`／`TELL_ME_TUI_DEBOUNCE` seam；無 pty；無新相依。
- 更新 `internal/cli/cli.go` 中已過時的註解（`…including the \`-i\` submit path **and the non-chrome path**` → 去掉 non-chrome 字樣；PR #51 review NIT）。

**Test Scope**:
- `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature`

- [ ] T010 [BDD-GREEN] 讓 Test Scope 全綠（並使 T006–T008 轉綠）
- [ ] T011 [BDD-REFACTOR] 在綠燈下整理 `-i` submit 的 teardown／echo／chrome 落點

## Phase 4B: MODIFY Feature File - cli/chat/presenting-the-turn.feature

**Goal**: `presenting-the-turn.feature` 全綠 — `-i` no-chrome Rule 已刪除（`-i` 改為 chrome surface）；其餘 positional／reader chrome Rules 不變。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature`
- `specs/truth/features/cli/chat/dsl.md` -> turn-chrome Thens（`the input capture is announced for the turn`、`the turn opens with a horizontal rule`、`the turn is headed "Turn N" for the active mode`、`the turn chrome is shown before the answer`）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`presenting-the-turn.feature`）

**Boundary**:
- 此 feature 為 MODIFY（刪除一條 Rule）；無新產品碼（`-i` chrome 由 Phase 4A 交付）。`stdout` byte-exact。
- `the run shows no turn chrome` root row 仍供 `diagnostics`／`history` 使用，**不得移除該 row 或 stepdef**。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature`

- [ ] T012 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T013 [BDD-REFACTOR] 整理 feature header note 與 Rule 群

## Phase 4C: MODIFY Feature File - cli/chat/watching-the-tool-loop.feature

**Goal**: `watching-the-tool-loop.feature`（含 `presenting-the-progress-spinner.feature` 的 header note）全綠 — round-022 negative 重錨為 chrome-aware 句；`-i` 從 spinner negatives 移除。

**Shared Must Read**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`、`specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`
- `specs/truth/features/cli/chat/dsl.md` -> `the pre-flight payload line is separated from the answer by a single blank line`、`the tool report is separated from the answer`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`watching-the-tool-loop.feature`、`presenting-the-progress-spinner.feature`）

**Boundary**:
- 此 feature 為 MODIFY（negative 重錨；spinner header note）；產品碼已於 Phase 4A 落地，本 phase 只調整 feature/assertion 語意。
- 不得改 round-022 的 positive separation 行為；`stdout` byte-exact。

**Test Scope**:
- `specs/truth/features/cli/chat/watching-the-tool-loop.feature`

- [ ] T014 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T015 [BDD-REFACTOR] 整理重錨後的註解與停用字詞

## Phase 4D: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞，並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T016 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）與 Gherkin/DSL topology audit。
  - **可偽性見證**（各觀察失敗後還原）：
    1. 還原 `model.go` 的 always-render → `the interactive prompt is cleared` 失敗（`View()` 非空）。
    2. 移除 `runTUIPrompt` 的 `chrome:true` → `the input capture is announced for the turn`（`-i`）失敗 / spinner 消失。
    3. 移除 `-i` 的 echo → `the submitted prompt "hi" is echoed on the diagnostic output` 失敗。
    4. 讓 positional surface 也 echo → `the prompt is not echoed on the diagnostic output` 失敗。
    5. 讓 used-no-tool `-i` run 多印一個空行 → `the pre-flight payload line is separated from the answer by a single blank line` 失敗。
  - 確認：`stdout` byte-exact；exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；offline paths 不變；`go mod tidy` 後 module graph 不變（無新相依）；topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`：TTY prompt / Turn chrome / Turn progress spinner / E2E runner / TUI harness / Pure-helper unit tests） | T002、T007、T008、T010 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務；T016 驗證詞彙維持 11） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務；T016 驗證 record 形狀不變） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`continuing-the-interactive-prompt.feature`） | T010、T011 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`presenting-the-turn.feature` + `watching-the-tool-loop.feature` + `presenting-the-progress-spinner.feature` + `chat/dsl.md` + root `cli/dsl.md`） | T003、T006、T012–T015 | PASS |
| `research.md` -> Decision 1（clear the frame on submit/abort） | T002、T006、T007、T010、T016 | PASS |
| `research.md` -> Decision 2（`-i` joins the chrome surface） | T002、T008、T010、T012、T016 | PASS |
| `research.md` -> Decision 3（spinner gate unchanged；`-i` positive） | T008、T012、T014、T016 | PASS |
| `research.md` -> Decision 4（echo on `-i` only；Q3 A′） | T002、T004、T005、T008、T010、T016 | PASS |
| `research.md` -> Decision 5（E2E witness rework — assert cleared + standard surface） | T006、T007、T010、T011、T016 | PASS |
| `research.md` -> Decision 6（turn-scoped spinner epoch unchanged） | T008、T010、T016 | PASS |
| `research.md` -> Decision 7（no new dependency；POSIX-only；seams unchanged） | T010、T016 | PASS |
| `research.md` -> Decision 8（presentation-only；contracts unchanged） | T010、T012、T014、T016 | PASS |
| **Round-022 ripple (Option 1)** — negative re-anchored to the chrome surface | T003、T014、T016 | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-014`、`SC-001`–`SC-005` | T003–T008（對齊）、T010–T015（交付）、T016 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/tui/prompt`、`internal/cli`、`internal/ui`）與 planner delegation（api/data NOOP；ui terminal reviewed；CLI end → /axb-dsl-refine） | T002、T010、T016 | PASS |
| `ui/**`（terminal mode：entry／10-submitted／20-abandoned） | T010（`10-submitted` 為 handoff 目標畫面；reviewed by `/axb-system-analysis`） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
