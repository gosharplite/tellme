# Tasks: tellme spinner width safety — a bounded tool label and a residue-free clear (round 025)

**Plan Package**: `specs/plans/025-spinner-width-safety`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方模組**（the width probe reuses `golang.org/x/term`, already a direct dependency since round 012; `go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（**1 句 `MODIFY` → `[BDD-ALIGN]`**；**2 句 `ADD` → `[BDD-RED]`**；另 **2 個 `[UNIT]`**）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/spinner.go`（the bounded several-tool label + the row-aware clear + a `columns` width seam）、`internal/cli/cli.go`（inject the `columns` seam；honour the `TELL_ME_FORCE_STDERR_COLS` diagnostic seam）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**不動** `stdout` bytes、the model-phase / single-tool / no-names labels、the resource segment、the `stderr`-terminal gate、the phase lifecycle、the turn-scoped elapsed、the class-phrase vocabulary（仍 11）。

## Round-025 locked decisions (implementation constraints — MUST)

> 來自 operator 拍板（Q1 → 1：fix both defects；Q2 → 1：track the last frame's rows and erase all of them）與 `research.md` Decisions 1–4。

- **[BOUNDED LABEL]** the several-tool phase label is **bounded** — ` Executing tools [<first> and <N-1> more]...` (the **first** tool name + the **count of the remaining** names); the single-tool ` Executing [<name>]...` 與 no-names ` Executing tools...` forms are **unchanged**；the label no longer enumerates every name. 見 T003/T006（research D1）。
- **[ROW-AWARE CLEAR]** the presenter **tracks the rendered-row count** of its last frame (`rows = ceil(visibleWidth / columns)`, clamped ≥ 1) and erases **every** occupied row — before each redraw and on the final clear — so a soft-wrapped frame leaves no residue (the round-019 teardown row still holds). 見 T005/T006（research D2）。
- **[WIDTH SEAM]** the terminal width comes from an injected `columns func() int` seam (default: probe the `stderr` fd via `golang.org/x/term.GetSize`; `0` when unknown → single-row best effort); the diagnostic `TELL_ME_FORCE_STDERR_COLS` env seam overrides it (mirroring `TELL_ME_FORCE_STDERR_TTY`). 見 T002/T004/T007（research D3）。
- **[VERIFICATION SPLIT]** the bounded **label** is asserted **E2E**; the row-aware **clear** is a **unit** pin (injected width; a flat capture cannot reproduce a terminal grid) **and** witnessed E2E as the clear byte sequence (cursor-up). 見 T006/T007/T011（research D4）。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）與產品碼落點，讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫行為。

- [ ] T001 建立 3 個 stepdef 獨立檔骨架 + 2 個 `[UNIT]` 落點檔骨架（Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 2 個新句 + 1 個改語意句）
    - `tests/e2e/steps/`（the self-registering stepdef pattern — `init()` registrar）
    - `internal/ui/spinner_test.go`（the round-019 unit landing — same package/home）
  - 只做：建立獨立落點檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t003_chat_then_bounded_tools.go`
    - `tests/e2e/steps/step_t004_chat_given_narrow_terminal.go`
    - `tests/e2e/steps/step_t005_chat_then_cleared_rows.go`
    - `internal/ui/spinner_width_test.go`（the `[UNIT]` landing for the bounded label + the row-aware clear; 空殼）
    - `internal/cli/spinner_cols_test.go`（the `[UNIT]` landing for the `columns`/`TELL_ME_FORCE_STDERR_COLS` seam; 空殼）
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔或既有 `spinner_test.go`。

- [ ] T002 落點產品碼骨架（`internal/ui/spinner.go` + `internal/cli/cli.go`）
  - Read:
    - `specs/plans/025-spinner-width-safety/research.md` -> Decision 1, Decision 2, Decision 3
    - `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)；Terminal detection）
    - `internal/ui/spinner.go`（the current `ExecutingToolsLabel` + `renderLocked`/`clearLocked` + `clearControl`）
    - `internal/cli/cli.go`（the round-019 spinner construction + the `stderr` terminal gate seam）
  - 只做：
    - `internal/ui/spinner.go`：the `ExecutingToolsLabel` 幾個分支的簽名不變（只預留 the bounded several-tool branch 的落點）；`Spinner` 新增 `columns func() int` 欄位（`NewSpinner` 參數或 option，預設 nil → single-row）；`renderLocked`/`clearLocked` 預留 the row-count 追蹤與 the rows-aware clear 的簽名（空輸出）。
    - `internal/cli/cli.go`：construct the spinner 時注入 the `columns` seam（default: `golang.org/x/term.GetSize` on the `stderr` fd，0 when unknown）；讀 `TELL_ME_FORCE_STDERR_COLS`（nil-safe）。
  - 不做：不真的 bound the label、不真的算 rows、不真的 erase the wrapped rows、不改 the model/single/no-names labels、不改 the resource segment、不改 the gate／lifecycle／elapsed、不加相依。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；**1 句 `MODIFY`（`[BDD-ALIGN]`）+ 2 句 `ADD`（`[BDD-RED]`）**。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪 3 句都在 `specs/truth/features/cli/chat/dsl.md`（the narrow-terminal Given 由 round-025 demote 至模組層）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪斷言形式：the several-tool spinner line 的 status 為 **bounded** 形式；the narrow-terminal Given 設 `TELL_ME_FORCE_STDERR_TTY=1` **與** `TELL_ME_FORCE_STDERR_COLS`；the clear Then 於 `stderr` 帶 rows-aware clear（cursor-up）。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，但語意是舊 truth。依 `dsl.md` 該列改測試（the row pattern 由 `… names every tool it is running` → `… names the first tool and counts the remaining tools`）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 stepdef（the narrow Given + the cleared-rows Then）。
- `[UNIT]`：**2 個**非 DSL 單元斷言（the bounded label formatter + the row-aware clear；the `columns` seam）。
- 三種 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the diagnostics are shown at a terminal narrower than the indicator line`
  -> `the progress spinner names the first tool and counts the remaining tools`
  -> `the progress indicator is cleared from every row it occupied`
  -> `the progress spinner no longer appears once the answer is written`（沿用）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/presenting-the-progress-spinner.feature` + `chat/dsl.md`）；`/axb-api-plan` NOOP；`/axb-data-plan` NOOP；`/axb-technical-research` MODIFY（`techstack.md`）
- `internal/ui/spinner.go`、`internal/cli/cli.go`（T002 落點）

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T007 各派一個獨立 subagent（目標檔案互斥，符合 Zero Shared Edits）；T008 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN（本輪改語意句；1 句）

- [ ] T003 [P] [BDD-ALIGN] `Then: the progress spinner names the first tool and counts the remaining tools`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner names the first tool and counts the remaining tools`
  - Landing: `tests/e2e/steps/step_t003_chat_then_bounded_tools.go`
  - 語意：`stderr` 的 spinner line status 為 **bounded** several-tool 形式 `Executing tools [<first> and <N-1> more]...`（first name + 其餘 count）；**不該** every-name 枚舉、**不該** collapse 成 single-tool、**不該** 漏 count。（the retired round-019 row `… names every tool it is running` 的 stepdef 於此改語意。）

### BDD-RED（本輪新增句型；1 Given + 1 Then）

- [ ] T004 [P] [BDD-RED] `Given: the diagnostics are shown at a terminal narrower than the indicator line`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the diagnostics are shown at a terminal narrower than the indicator line`
  - Landing: `tests/e2e/steps/step_t004_chat_given_narrow_terminal.go`
  - 語意：把 `TELL_ME_FORCE_STDERR_TTY=1` **與** `TELL_ME_FORCE_STDERR_COLS`（窄寬，例 `40`）設進 subprocess 環境，使 tellme 的 `stderr` terminal probe 回報 terminal **且** width probe 回報該窄寬（tool-phase frame 會 soft-wrap）。

- [ ] T005 [P] [BDD-RED] `Then: the progress indicator is cleared from every row it occupied`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress indicator is cleared from every row it occupied`
  - Landing: `tests/e2e/steps/step_t005_chat_then_cleared_rows.go`
  - 語意：`stderr` 帶一個 clear sequence，erase **every** terminal row the last frame occupied — 在窄寬 seam 迫使 frame 寬於 terminal 時，clear 含 cursor-up（`\x1b[<n>A`）後接 per-row erase；讀為 terminal 後無 braille-frame phase-status 殘留。

### UNIT（非 DSL 的單元斷言）

- [ ] T006 [P] [UNIT] the bounded several-tool label + the row-aware clear
  - Read:
    - `specs/plans/025-spinner-width-safety/research.md` -> Decision 1, Decision 2, Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)）
    - `internal/ui/spinner.go`（`ExecutingToolsLabel` + `clearControl` + `renderLocked`/`clearLocked`）
  - 撰寫：斷言 the label formatter — `N=0` → ` Executing tools...`、`N=1` → ` Executing [<name>]...`（unchanged）、`N>=2` → ` Executing tools [<first> and <N-1> more]...`（bounded；不得 every-name 枚舉）；並斷言 the row-aware clear — 注入窄寬，畫一 over-wide frame，clear 後 emit 的序列 erase **every** occupied row（`rows = ceil(width/columns)`；含 cursor-up），且 `rows==1` 時退化為單列 clear（**無** `time.Sleep`）。
  - 落點：`internal/ui/spinner_width_test.go`。

- [ ] T007 [P] [UNIT] the `columns` seam + the `TELL_ME_FORCE_STDERR_COLS` override
  - Read:
    - `specs/plans/025-spinner-width-safety/research.md` -> Decision 3
    - `specs/truth/techstack.md` -> CLI Terminal detection（the diagnostic-stream gate + the cols seam）
    - `internal/cli/cli.go`（the spinner construction + the round-019 stderr gate seam）
  - 撰寫：斷言 width 解析 — 注入的 `columns` seam 生效；`TELL_ME_FORCE_STDERR_COLS` 覆寫 seam；寬度未知（`0`）時 the presenter 退化為 single-row best effort（不 panic、不 over-erase）。
  - 落點：`internal/cli/spinner_cols_test.go`。

### Phase Review Gate

- [ ] T008 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`
    - `internal/ui/spinner.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；1 個 `[BDD-ALIGN]` stepdef 已改語意；2 個新句 stepdef + 2 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（the label 尚未 bound、the clear 尚未 rows-aware、the `columns` seam 尚未接線）。

---

## Phase 4A: MODIFY Feature File - cli/chat/presenting-the-progress-spinner.feature

**Goal**: 讓 `presenting-the-progress-spinner.feature` 全綠 — the several-tool label 為 bounded 形式，且 the width-safe clear 於窄 terminal 無殘留；其餘 spinner 行為（model/single/no-names label、resource segment、gate、lifecycle、elapsed）不變。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` -> `Feature: Presenting the progress spinner`
- `specs/truth/features/cli/chat/dsl.md` -> the 3 round-025 rows
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY；`/axb-technical-research` MODIFY
- `specs/plans/025-spinner-width-safety/research.md` -> Decision 1, Decision 2, Decision 3, Decision 4
- `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)；Terminal detection）

**Boundary**:
- 產品碼：`internal/ui/spinner.go`（`ExecutingToolsLabel` 的 several-tool branch → bounded；the presenter 追蹤 rendered-row count 並 erase every occupied row；a `columns` seam）、`internal/cli/cli.go`（inject the `columns` seam；honour `TELL_ME_FORCE_STDERR_COLS`）。
- 依 `research.md` Decisions 1–4；`N=0`/`N=1` label 不動；the resource segment 不動；the gate／lifecycle／elapsed 不動；an unknown width 退化為 single-row best effort。
- **不動** `stdout` bytes、the class-phrase 詞彙（維持 **11**）、the `-i` TUI surface；plain text（no ANSI）；**不加相依**。
- 落點採零共用編輯。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`

- [ ] T009 [BDD-GREEN] 讓 Test Scope 全綠（並使 T006/T007 的 `[UNIT]` 轉綠）
- [ ] T010 [BDD-REFACTOR] 在綠燈下整理 the label branch／the row-count clear／the width seam 落點

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、the model/single/no-names labels 不變、the resource segment 不變、the round-009 payload line 文字不變、the round-017 turn chrome 不變、the round-018 post-turn lines 不變、exit-code 表、offline paths、rounds 001–024 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP、`/axb-data-plan` NOOP）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T011 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（bounded label；非真空）**：暫時恢復 the every-name 枚舉 label，確認 `presenting-the-progress-spinner.feature` 的 `the progress spinner names the first tool and counts the remaining tools` 失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（row-aware clear）**：暫時把 the clear 退回單列 `\r\x1b[K`，確認 the narrow-terminal Example 的 `the progress indicator is cleared from every row it occupied` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact；the model-phase / single-tool / no-names labels 不變；the resource segment 不變；the round-009 payload line 文字不變；the round-017 turn chrome 不變；the round-018 post-turn lines 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths 不變；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator) — bounded label + row-aware clear） | T002、T006、T009、T010 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Terminal detection — the `TELL_ME_FORCE_STDERR_COLS` seam） | T002、T007、T009 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（round-025 unit pins） | T001、T006、T007、T011 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T002、T006、T007、T009 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T011 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/presenting-the-progress-spinner.feature`） | T003、T004、T005、T009、T010 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` +2 rows、−1 row、round-025 note） | T001、T003、T004、T005、T008 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md` — narrow Given demoted to the module） | 豁免（NOOP 不建任務；T008 review 確認 root 未被誤改） | PASS |
| `research.md` -> Decision 1（bounded several-tool label） | T002、T003、T006、T009 | PASS |
| `research.md` -> Decision 2（row-aware clear） | T002、T005、T006、T009 | PASS |
| `research.md` -> Decision 3（the `columns` seam + `TELL_ME_FORCE_STDERR_COLS`） | T002、T004、T007、T009 | PASS |
| `research.md` -> Decision 4（verification split：label E2E · clear unit + bytes） | T001、T005、T006、T007、T011 | PASS |
| `spec.md` -> US1–US2、`FR-001`–`FR-007`、`NFR-001`–`NFR-003` | T003–T007（對齊）、T009–T011（交付） | PASS |
| `spec.md` -> 邊界情況（2/1/0 tools；exact-width；unknown width） | 2 tools → T003/T006；1 tool → T006；0 tools → T006；unknown width → T007 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/spinner.go`、`internal/cli/cli.go`；no new module） | T002、T009、T010 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T008、T010、T011 | PASS |
| operator 拍板（Q1 → 1：fix both defects；Q2 → 1：row-tracking clear） | Q1 → T003/T005；Q2 → T005/T006 | PASS |
| `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` -> the existing teardown Then（沿用） | T005、T011 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
