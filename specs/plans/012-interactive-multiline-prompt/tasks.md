# Tasks: Interactive Multi-line Prompt Capture (round 012)

**Plan Package**: `specs/plans/012-interactive-multiline-prompt`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪為「pin + unit」特性**：互動式多行讀取是 **POSIX 終端**能力，pty-less E2E harness 無法驅動（round-005 named pin）。**正面路徑**（hint→stderr、讀到 EOF、一次 reasoning turn、empty/cancel 不送 request）由 **`[UNIT]`**（注入 `isTTY` seam + scripted reader）驗證；**E2E** 只承載可觀察的**負向**（pipe → 不輸出 reading announcement）。產品碼（interactive reader）於 Phase 4A 的 `[BDD-GREEN]` 落地，並被 Phase 3 的 `[UNIT]` 驅動轉綠。

---

## Phase 2: Foundational

**Goal**: 建立本輪新句 stepdef 落點骨架與互動讀取的單元測試落點（Zero Shared Edits 原則），讓 Phase 3 不各自發明檔。不寫產品行為。

- [X] T001 建立負向句 stepdef 獨立檔案骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（新 Then 句 `no reading announcement is reported`）
    - `tests/e2e/steps/register.go`
  - 只做：建立 `tests/e2e/steps/step_t016_chat_then_no_reading_announcement.go`，`init()` 自我註冊空白 registrar。
  - 不做：不寫具體斷言邏輯；不碰既有 step 檔案；不碰 `internal/`。

- [X] T002 建立互動讀取的單元測試落點
  - Read:
    - `specs/plans/012-interactive-multiline-prompt/research.md` -> Decision 1, 2, 4, 6
    - `internal/cli/cli.go`（`runtimeEnv.isTTY` seam）、`internal/cli/turn_test.go`
  - 只做：建立 `internal/cli/prompt_multiline_test.go` 骨架：一個「以注入的 `isTTY` 回報 stdin 是終端 + 提供 scripted `io.Reader`（含 EOF）」的 helper 與 table 骨架。
  - 不做：不寫產品碼；不啟用互動分支。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含本輪新句 `no reading announcement is reported`，以及互動讀取的 `[UNIT]` 斷言。不寫產品行為。

**DSL 參照**:
- 該句屬同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（該句為 `chat` 模組專屬；frozen class-phrase 詞彙維持 10）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Then 讀 `必查`（`呈現結果`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：**無**。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 1 句（`ADD` 的 DSL row）。
- `[UNIT]`：互動讀取（正面路徑 + empty/cancel），透過注入的 `isTTY` seam 驅動（research Decision 6）。
- 兩者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `no reading announcement is reported`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reading-a-multi-line-prompt.feature`）+ MODIFY（`chat/dsl.md`）；NOOP（root `cli/dsl.md`）
- `tests/e2e/steps/`、`internal/cli/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（`init()` 自我註冊）。
- `[UNIT]` 於 `internal/cli/prompt_multiline_test.go`，採 stdlib `testing`；以注入的 `isTTY` seam（stdin=true）＋ scripted stdin 驅動互動分支（無 pty）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003、T004 各派一個獨立 subagent；T005 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [X] T003 [P] [BDD-RED] `Then: no reading announcement is reported`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `no reading announcement is reported`
  - Landing: `tests/e2e/steps/step_t016_chat_then_no_reading_announcement.go`
  - 語意：斷言 captured stdout 與 stderr **皆不**包含多行 reading hint（`[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]`）。

### UNIT（非 DSL 的單元斷言）

- [X] T004 [P] [UNIT] 互動多行讀取（注入的 `isTTY` seam）
  - Read:
    - `specs/plans/012-interactive-multiline-prompt/research.md` -> Decision 1, 2, 4, 6
    - `specs/truth/techstack.md` -> CLI Application（Interactive prompt read）+ Testing & Verification（Pure-helper unit tests / Test strategy）
    - `internal/cli/cli.go`（`runtimeEnv`、`run`）、`internal/cli/prompt_multiline_test.go`
  - 撰寫：以 `isTTY` seam 回報 stdin 是終端 + scripted stdin（多行 + EOF），斷言：hint 寫到 `stderr`（且不在 `stdout`）；captured 多行文字成為唯一 turn 的 prompt；**empty（立即 EOF）** 與 **cancel（context 取消）** 皆**不送 request**、不印答案。

### Phase Review Gate

- [X] T005 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature`、`specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/step_t016_chat_then_no_reading_announcement.go`、`internal/cli/prompt_multiline_test.go`
    - `internal/cli/cli.go`
  - 檢驗：無 undefined step；新 stepdef 已註冊；`[UNIT]` 已驅動互動分支並全綠；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: ADD Feature File - cli/chat/reading-a-multi-line-prompt.feature

**Goal**: 實作互動式多行讀取（產品碼）並讓 `reading-a-multi-line-prompt.feature` 全綠；同時讓 Phase 3 的 `[UNIT]` 轉綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature` -> `Feature: Reading a multi-line prompt`
- `specs/truth/features/cli/chat/dsl.md` -> `no reading announcement is reported` + round-012 note
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reading-a-multi-line-prompt.feature`）+ MODIFY（`chat/dsl.md`）；`/axb-technical-research` MODIFY（`techstack.md` Prompt input / Terminal detection / Interactive prompt read）
- `specs/plans/012-interactive-multiline-prompt/research.md` -> Decision 1, 2, 3, 4, 5（seam / bounded EOF read / hint 到 stderr / empty-cancel / POSIX-only）

**Boundary**:
- 產品碼：`internal/cli` 在「resolved prompt 為空、無 positional args、無 piped stdin、且 stdin 為終端、且無 terminal-less 子命令（`--version`/`-d`/`-l`/prompt-less `--new`）」時，印出多行 hint 到 **`stderr`** 並將 stdin **讀到 EOF**（`io.ReadAll`/`io.LimitReader`，上限 1 MiB），captured 文字（trim）成為一次 reasoning turn 的 prompt；`Ctrl+C`（SIGINT context）取消；empty/whitespace ⇒ 不送 request。
- 提示必須在 `stderr`；`stdout` 保持 byte-exact；不得引入 class-phrase prefix；**POSIX-only，不得有 Windows 分支**。
- 不得改動 pipe/positional/non-prompt 路徑與既有 behavior。

**Test Scope**:
- `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature`

- [X] T006 [BDD-GREEN] 讓 Test Scope 全綠（並實作 interactive reader，使 Phase 3 `[UNIT]` 轉綠）
- [X] T007 [BDD-REFACTOR] 在綠燈下整理互動讀取 helper 與 dispatch 分支

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（pipe/positional 路徑、byte-exact `stdout`、exit-code 與 class-phrase 契約、status/ordering 契約），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [X] T008 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**：暫時讓互動 reader 在非終端也輸出 hint（或讓 empty 仍送 request），確認相關 unit/E2E 斷言失敗；觀察到失敗即還原。
  - 確認：exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **10**；pipe/positional/non-prompt 路徑與 `stdout` byte 不變；offline paths 不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Interactive prompt read） | T004、T006、T007 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Prompt input；Terminal detection） | T004、T006 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner / Test strategy / Pure-helper unit tests） | T003、T004、T008 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T004、T006、T008 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/reading-a-multi-line-prompt.feature`） | T003、T006、T007 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md`） | T001、T003、T006 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T008 驗證詞彙維持 10） | PASS |
| `research.md` -> Decision 1（reuse the dependency-free TTY seam; no `x/term`） | T002、T004、T006 | PASS |
| `research.md` -> Decision 2（read to EOF；bounded 1 MiB） | T004、T006 | PASS |
| `research.md` -> Decision 3（hint to `stderr`；no class prefix） | T004、T006 | PASS |
| `research.md` -> Decision 4（empty/cancel ⇒ no request） | T004、T006 | PASS |
| `research.md` -> Decision 5（POSIX-only；no Windows variant） | T006、T007 | PASS |
| `research.md` -> Decision 6（unit seam + E2E negatives；named pin） | T002、T003、T004、T005 | PASS |
| `research.md` -> Decision 7（no new dependency；techstack unchanged） | T008（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 8（must-ask questions settled） | T003–T008（不改 runner/端） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-011`、`NFR-001`–`NFR-005` | T003、T004（對齊）、T006–T007（交付）、T008 | PASS |
| `plan.md` -> Source-code structure（`internal/cli` interactive read + unit seam；E2E negative step） | T001、T002、T003、T004、T006 | PASS |
| `plan.md` -> Gating/verification（named pin；pty-less） | T003、T004、T005 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
