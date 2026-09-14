# Tasks: tellme turn surface — operator chrome parity (round 017)

**Plan Package**: `specs/plans/017-turn-chrome-parity`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **本輪沒有新增技術**：the chrome is hand-written Go in the existing `internal/ui` (`stdlib` only); `go.mod`/`go.sum` are untouched. 依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（本輪 **7 句為全新句 `ADD`**；另 1 個 `[UNIT]`）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/turn.go`（新 formatter — the input-capture acknowledgement + the 80-column `─` rule + the `╭─⠿ Turn <N> - <mode>` header + the blank-line spacing）與 `internal/cli/cli.go`（在 surfaces (A)/(B) 上發出 chrome；以 *chrome* switch 排除 `-i` submit path (C) 與非 prompt paths）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**不動** the round-009 payload-line formatter、post-turn path、keybindings、TUI surface（round 016）。
- 本輪**無跨模組**新句：root `specs/truth/features/cli/dsl.md` 為 **NOOP**（class-phrase 詞彙維持 **11**）；所有新句皆 `chat` 模組專屬。

## Round-017 locked decisions (implementation constraints — MUST)

> 來自 `/axb-clarify` Round 1（operator 拍板）與 `research.md` Decisions 1–7。

- **[SURFACE SCOPE = A+B]** — the chrome 只發在 the **positional/piped prompt turn (A)** 與 the **round-012 plain reader (B)**；the `-i` TUI surface (round 016) 與 every non-prompt path (boot / `-d` / `-l` / `--version` / prompt-less `--new`) **不得**出現 chrome（`FR-007`）。見 T002、T009、T012。
- **[STARTUP SET = input-capture line only]** — 只加 `[HH:MM:SS] Input captured. Processing...`；the reference 的第二行 `[Info] Starting chat...` **out of scope**（`A3`）。見 T003、T004。
- **[TURN NUMBER = persisted + 1]** — `<N>` = the session's persisted turn count + 1（`len(prior)+1`；`--new` → `Turn 1`），`<mode>` = the run's effective mode（`FR-004`）。見 T006、T007、T010。
- **[PLAIN TEXT]** — 結構化 chrome 以 plain text 呈現（**no ANSI** this round；research D3）；the round-009 payload line text **不變**，僅被包進 frame。見 T005、T008、T010、T013。
- **[SPACING]** — 一個 leading blank line before the rule、一個 trailing blank line after the payload line（research D4）。見 T007、T008、T010。
- **[NO REGRESSION]** — `stdout` 維持 byte-exact（chrome 只寫 `stderr`）；class-phrase 詞彙維持 **11**；the payload line format、post-turn lines、exit-code 表、offline paths 不變。見 T012、T014。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）與產品碼落點（the chrome formatter），讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫 chrome 行為。

- [ ] T001 建立 7 個 stepdef 獨立檔骨架（7 新句；Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 7 個新句）
    - `tests/e2e/steps/register.go`、`tests/e2e/harness/`（the merged-stream witness + the subprocess runner）
  - 只做：建立 7 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t003_chat_then_input_capture_announced.go`
    - `tests/e2e/steps/step_t004_chat_then_input_capture_before_frame.go`
    - `tests/e2e/steps/step_t005_chat_then_opens_with_rule.go`
    - `tests/e2e/steps/step_t006_chat_then_headed_turn_number.go`
    - `tests/e2e/steps/step_t007_chat_then_chrome_before_answer.go`
    - `tests/e2e/steps/step_t008_chat_then_frame_separated.go`
    - `tests/e2e/steps/step_t009_chat_then_no_turn_chrome.go`
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [ ] T002 落點 the turn-chrome formatter 與其 `[UNIT]` 落點檔骨架
  - Read:
    - `specs/plans/017-turn-chrome-parity/research.md` -> Decision 1, Decision 3, Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator)）
    - `internal/ui/status.go`（the sibling round-009 formatter — same package/home）
  - 只做：
    - 產品碼骨架：在 `internal/ui/turn.go` 宣告 the chrome formatter 的簽名 stub（輸出：the input-capture acknowledgement line、the 80-column `─` rule、the `╭─⠿ Turn <N> - <mode>` header、the leading／trailing blank lines），以及 token 常數（the 80× `─` rule literal、the `╭─⠿` glyph）留位；對齊 `internal/ui/status.go` 的 formatter 風格。
    - 測試落點骨架：`internal/ui/turn_test.go`（stdlib `testing` 空殼）。
  - 不做：不寫渲染結果（除空字串 stub）、不接線 `cli.go`、不改 the round-009 payload-line formatter、不引入任何相依。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 **7 句為全新句（`ADD`）**。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪新句皆 `chat` 模組專屬；features 僅重用其既有 root row，語意不變且 stepdef 已在）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪 chrome Then 的斷言形式：the capture line／the rule／the header **presence** on `stderr`；ordering（Before the frame／Before the answer）與 the blank-gap separation 依 **merged**（`stdout`+`stderr`）witness；the negative Then 對 `stdout`+`stderr` 的 **absence**。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **7 句**（`ADD` 的 DSL rows）。
- `[UNIT]`：1 個非 DSL 單元斷言（the chrome formatter — deterministic formatting）。
- 兩者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the input capture is announced for the turn`
  -> `the input capture is announced before the turn frame`
  -> `the turn opens with a horizontal rule`
  -> `the turn is headed "Turn {number}" for the active mode`
  -> `the turn chrome is shown before the answer`
  -> `the turn frame is separated from the answer`
  -> `the run shows no turn chrome`
- `specs/truth/features/cli/dsl.md` -> **NOOP**（root 詞彙不變，維持 11 片語）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-turn.feature`）+ MODIFY（`chat/dsl.md`，+7 rows）；NOOP（root `cli/dsl.md`、其他 `cli` 模組、`contracts/**`、`data/**`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
- `internal/ui/status.go`、`internal/ui/turn.go`（T002 落點）

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/ui/turn_test.go`。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T010 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；目標檔案互斥，符合 Zero Shared Edits）；T011 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T003 [P] [BDD-RED] `Then: the input capture is announced for the turn`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the input capture is announced for the turn`
  - Landing: `tests/e2e/steps/step_t003_chat_then_input_capture_announced.go`
  - 語意：captured `stderr` 帶一行符合 `[HH:MM:SS] Input captured. Processing...`（timestamp 以 pattern 比對）；**不**在 `stdout`。

- [ ] T004 [P] [BDD-RED] `Then: the input capture is announced before the turn frame`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the input capture is announced before the turn frame`
  - Landing: `tests/e2e/steps/step_t004_chat_then_input_capture_before_frame.go`
  - 語意：merged capture 中，the acknowledgement 出現在 the horizontal rule **之前**。

- [ ] T005 [P] [BDD-RED] `Then: the turn opens with a horizontal rule`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the turn opens with a horizontal rule`
  - Landing: `tests/e2e/steps/step_t005_chat_then_opens_with_rule.go`
  - 語意：captured `stderr` 帶一行，其可見內容為 the reference 的 **80 欄** `─` rule。

- [ ] T006 [P] [BDD-RED] `Then: the turn is headed "Turn {number}" for the active mode`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the turn is headed "Turn {number}" for the active mode`
  - Landing: `tests/e2e/steps/step_t006_chat_then_headed_turn_number.go`
  - 語意：captured `stderr` 帶一行符合 `╭─⠿ Turn {number} - <mode>`，`<mode>` 為該次執行 effective mode；**不**在 `stdout`，且 number 不得不同。

- [ ] T007 [P] [BDD-RED] `Then: the turn chrome is shown before the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the turn chrome is shown before the answer`
  - Landing: `tests/e2e/steps/step_t007_chat_then_chrome_before_answer.go`
  - 語意：merged capture 中，the header（the rule 與 the `╭─⠿ Turn …` line）出現在 the answer bytes **之前**。

- [ ] T008 [P] [BDD-RED] `Then: the turn frame is separated from the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the turn frame is separated from the answer`
  - Landing: `tests/e2e/steps/step_t008_chat_then_frame_separated.go`
  - 語意：merged capture 中，一個空行分隔 the frame 的末行（the pre-flight payload line）與 the answer bytes。

- [ ] T009 [P] [BDD-RED] `Then: the run shows no turn chrome`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run shows no turn chrome`
  - Landing: `tests/e2e/steps/step_t009_chat_then_no_turn_chrome.go`
  - 語意：`stdout` 與 `stderr` 皆**不**帶 the input-capture acknowledgement、the 80-column `─` rule、或 `╭─⠿ Turn …` header；用於 `-i` submit path 與 non-prompt path。

### UNIT（非 DSL 的單元斷言）

- [ ] T010 [P] [UNIT] the turn-chrome formatter — deterministic formatting
  - Read:
    - `specs/plans/017-turn-chrome-parity/research.md` -> Decision 1, Decision 3, Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator)）
    - `internal/ui/status.go`（the sibling formatter — the round-009 payload line text must stay unchanged）
  - 撰寫：斷言 formatter 輸出——the input-capture acknowledgement line（`[HH:MM:SS] Input captured. Processing...`，時刻由注入 clock 決定）、the **80 欄** `─` rule、the `╭─⠿ Turn <N> - <mode>` header、以及 the leading／trailing blank lines；並斷言 the round-009 payload-line 文字**未**被改變。
  - 落點：`internal/ui/turn_test.go`。

### Phase Review Gate

- [ ] T011 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-turn.feature`、`reporting-the-payload-status.feature`、`using-the-interactive-prompt.feature`、`reading-a-multi-line-prompt.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
    - `internal/ui/turn.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；7 個新句 stepdef + `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（chrome 未發出、turn number 未計算、`-i`/non-prompt 尚未隔離）。

---

## Phase 4A: ADD Feature File - cli/chat/presenting-the-turn.feature

**Goal**: 讓 `presenting-the-turn.feature` 全綠 — 在 (A)/(B) 的 prompt turn 上發出 reference chrome：the input-capture acknowledgement、the 80-column `─` rule、the `╭─⠿ Turn <N> - <mode>` header（over the pre-flight payload line）、the blank-gap spacing；且 the `-i` prompt 與 non-prompt runs **不**出現 chrome。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature` -> `Feature: Presenting the turn`
- `specs/truth/features/cli/chat/dsl.md` -> `the input capture is announced for the turn`、`… before the turn frame`、`the turn opens with a horizontal rule`、`the turn is headed "Turn {number}" for the active mode`、`the turn chrome is shown before the answer`、`the turn frame is separated from the answer`、`the run shows no turn chrome`、`tellme reports the estimated payload status for the turn`（既有，被包進 frame）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-turn.feature`）+ MODIFY（`chat/dsl.md`）+ `/axb-technical-research` MODIFY（`techstack.md` Turn chrome）
- `specs/plans/017-turn-chrome-parity/research.md` -> Decision 1, Decision 2, Decision 4, Decision 5

**Boundary**:
- 產品碼：`internal/ui/turn.go`（the chrome formatter — acknowledgement + 80-column rule + `╭─⠿ Turn <N> - <mode>` header + spacing；plain text）+ `internal/cli/cli.go`（在 surfaces (A)/(B) 的 turn 上發出 the acknowledgement 與 the frame，包住既有 pre-flight payload line；以 *chrome* switch 讓 `-i` submit path (C) 與 non-prompt paths 不發出）。
- `<N>` = the session's persisted turn count + 1（`len(prior)+1`）；`<mode>` = the resolved effective mode。
- **不動** the round-009 payload-line formatter（其文字不變，僅被包進 frame）、the post-turn path、the keybindings、the round-016 TUI surface；`stdout` 維持 byte-exact（chrome 只寫 `stderr`）；class-phrase 詞彙維持 **11**；plain text（**no ANSI**）；不加相依。
- 落點採零共用編輯；the emission seam 只加在 the turn entry，不複製到多處。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature`

- [ ] T012 [BDD-GREEN] 讓 Test Scope 全綠（並使 T010 的 `[UNIT]` 轉綠）
- [ ] T013 [BDD-REFACTOR] 在綠燈下整理 formatter 與 emission seam（含 *chrome* switch 的落點）

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、the payload line 文字不變、exit-code 表、offline paths、rounds 001–016 全綠），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T014 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（rule／frame；非真空）**：暫時停發 the horizontal rule（或 the header）於 (A)/(B)，確認 `presenting-the-turn.feature` 的 `the turn opens with a horizontal rule` 失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（surface scope）**：暫時讓 the chrome 洩漏到 the `-i` submit path，確認 `presenting-the-turn.feature` 的 `the run shows no turn chrome` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact（chrome 只寫 `stderr`）；the round-009 payload-line 文字不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；rounds 001–016 既有場景全綠；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator)） | T002、T010、T012、T013 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T002、T010、T012、T013 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T014 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-turn.feature`） | T003–T009、T012、T013 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — +7 rows） | T001、T003–T009、T011 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md` + 其他 `cli` 模組） | 豁免（NOOP 不建任務；T014 驗證詞彙維持 11） | PASS |
| `research.md` -> Decision 1（chrome 由 `internal/ui` 手寫 formatter 發出；no new dependency） | T002、T010、T012、T013 | PASS |
| `research.md` -> Decision 2（turn number = 已持久化 turn 數 + 1） | T006、T007、T012 | PASS |
| `research.md` -> Decision 3（plain text；the payload line 文字不變） | T005、T008、T010、T012、T013 | PASS |
| `research.md` -> Decision 4（blank-line spacing） | T007、T008、T010、T012 | PASS |
| `research.md` -> Decision 5（one seam；只 surfaces (A)/(B)） | T002、T009、T012、T013 | PASS |
| `research.md` -> Decision 6（hermetic 無 pty 驗證） | T001、T010、T011 | PASS |
| `research.md` -> Decision 7（no new dependency；POSIX-only；must-ask 不變） | T002（Boundary）、T013（Boundary）、T014 | PASS |
| `spec.md` -> US1–US3、`FR-001`–`FR-009`、`NFR-001`–`NFR-003` | T003–T010（對齊）、T012–T014（交付） | PASS |
| `spec.md` -> A2（surface scope A+B）、A3（input-capture line only）、A4（turn number）、A5（chrome tokens；the payload line 不變）、A7（post-turn out of scope） | A2 → T002/T009/T012；A3 → T003/T004；A4 → T006/T007/T012；A5 → T005/T010/T013；A7 → T012（Boundary）/T014 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/turn.go`；`internal/cli/cli.go`；no new pkg/module） | T002、T012、T013 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T011、T014 | PASS |
| `/axb-clarify` Round 1 locked decisions（Q1 → A+B；Q2 → input-capture line only） | Q1 → T002/T009/T012/T013；Q2 → T003/T004 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
