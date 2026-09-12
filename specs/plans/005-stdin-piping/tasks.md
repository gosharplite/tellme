# Tasks: Prompt Piping (round 005)

**Plan Package**: `specs/plans/005-stdin-piping`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 省略（本輪未引入新的第三方依賴）。
- Phase 2 `Foundational` 只建立落點骨架與測試共用元件；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 1: Setup

*(省略 — 本輪未引入新的第三方依賴：TTY 偵測使用 Go 標準庫 `os.File` + `os.ModeCharDevice`（不採用 `golang.org/x/term`），stdin 讀取使用標準庫 `io.LimitReader`，`go.mod` 不變，完全沿用既有 Go 1.26 工具鏈與 Makefile。)*

---

## Phase 2: Foundational

**Goal**: 建立 E2E harness 的 scripted-stdin 能力、scenario 的 stdin 管線、9 個新 stepdef 獨立檔案落點骨架，以及 CLI I/O seam 與純函式落點骨架（Zero Shared Edits 原則）。

- [X] T001 擴充 E2E harness 支援 scripted stdin `tests/e2e/harness/cmd_helper.go`
  - Read:
    - `specs/plans/005-stdin-piping/research.md` -> Decision 2（bounded stdin read）、Decision 4（E2E 策略）
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner / step definitions）
  - 只做：新增一個可注入 scripted stdin 的執行入口（例如 `RunWithStdin(args []string, stdin string, set map[string]string, unset []string) RunResult`），以 `strings.NewReader` → os/exec 建立的 **pipe** 作為 child 的 stdin（非 char device，讓子行程判定為「非 TTY」）；保留既有 `Run`／`RunBinary` 行為不變（nil stdin = `/dev/null`）。
  - 不做：不寫任何 stepdef 或斷言；不碰 `internal/`。

- [X] T002 建立 scenario stdin 管線 `tests/e2e/steps/scenario_context.go`
  - Read:
    - `specs/plans/005-stdin-piping/research.md` -> Decision 4
    - `tests/e2e/steps/scenario_context.go`、`tests/e2e/steps/register.go`
  - 只做：在 `scenarioContext` 新增 `stdin string` 與 `stdinSet bool` 欄位，以及一個 setter（例如 `pipeStdin(content string)`）；`run()` 改為在 `stdinSet` 時呼叫 T001 的 `RunWithStdin`，否則沿用既有 `harness.Run`。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案或 `harness/`。

- [X] T003 建立 9 個 piped stepdef 獨立檔案骨架 `tests/e2e/steps/step_t006_*.go`–`step_t014_*.go`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪新增 4 個 When + 5 個 Then 句型）
    - `tests/e2e/steps/register.go`
  - 只做：在 `tests/e2e/steps/` 下建立 9 個獨立檔案（`step_t006_chat_when_pipes_content.go`、`step_t007_chat_when_pipes_with_instruction.go`、`step_t008_chat_when_pipes_nothing.go`、`step_t009_chat_when_diagnostic_with_piped.go`、`step_t010_chat_then_request_piped_content.go`、`step_t011_chat_then_request_instruction_then_content.go`、`step_t012_chat_then_captured_stdout_exact.go`、`step_t013_chat_then_captured_stdout_no_decoration.go`、`step_t014_chat_then_completes_without_waiting.go`），各自宣告 `init()` 並將空白 registrar 加入 `registrars` 切片。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案（檔名 suffix 與既有 `step_t006_config_*` 等互不衝突）。

- [X] T004 建立 CLI I/O seam 與 prompt 組合落點骨架 `internal/cli/prompt.go`（+ `cmd/tellme/main.go`、`internal/cli/cli.go` 簽名）
  - Read:
    - `specs/plans/005-stdin-piping/research.md` -> Decision 1（TTY 偵測）、Decision 3（prompt 組合）、Decision 4（I/O seam）、Decision 6（僅 prompt-turn 讀 stdin）
    - `specs/truth/techstack.md` -> CLI Application（Prompt input / Terminal detection / Piped stdin read）
    - `internal/cli/cli.go`、`cmd/tellme/main.go`
  - 只做：宣告純函式簽名 `combinePrompt(args []string, piped []byte) string`（回傳 stub）；宣告 TTY 偵測 seam 型別（例如 `terminalDetector func(any) bool`）與 stdin/stdout/stderr 注入參數，讓 `cli.Run`／`runTurn` 接受注入；`cmd/tellme/main.go` 以真實 `os.Stdin/Stdout/Stderr` 與 `os.ModeCharDevice` 偵測器呼叫。
  - 不做：不實作組合邏輯、不實作 TTY 判斷、不改既有 dispatch 行為（產品行為留 Phase 4A）。

- [X] T005 建立純函式單元測試落點骨架 `internal/cli/prompt_test.go`
  - Read:
    - `specs/plans/005-stdin-piping/research.md` -> Decision 3、Decision 4（folding issue #14）
    - `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests）
  - 只做：建立表驅動測試骨架（空 cases 切片與 `t.Run` 迴圈）與一個可注入 TTY 判斷 / stdin 的測試替身骨架，供 Phase 3 `[UNIT]` 填表。
  - 不做：不寫斷言內容、不碰產品碼。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的句（`chat/dsl.md` 新增 4 個 When + 5 個 Then）與 piped 行為的 pure-helper 單元測試。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：本輪 9 個新句全在同模組 `specs/truth/features/cli/chat/dsl.md`；不得掃其他模組。
- 本輪 Feature 用到的既有句（`the operator has a runnable tellme installation`、`the runtime home is "{home}"`、`a configured provider "{provider}" whose endpoint answers with "{answer}"`、`the operator starts tellme with the prompt "{prompt}"`、`tellme sends exactly one request to the provider "{provider}"`、`tellme prints the provider's answer "{answer}"`、`tellme performs no network access`、`tellme exits successfully`）stepdef 已在且語意不變，不列。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-OFF]`：無。本輪無 `MODIFY`／`DELETE` 既有句，無 `[BDD-ALIGN]`／`[BDD-REMOVE]`。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 3/4）。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine` ADD ×2（features）與 MODIFY（`chat/dsl.md` 9 句）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案的 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 `internal/cli/prompt_test.go`，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T006–T015 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T016 等全部回來再啟動 subagent 執行 review。

### BDD-RED（chat 模組新增句型）

- [X] T006 [P] [BDD-RED] `When: the operator pipes "{content}" into tellme`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator pipes "{content}" into tellme`
  - Landing: `tests/e2e/steps/step_t006_chat_when_pipes_content.go`
  - 語意：無 `-c`、無 positional prompt；以 `{content}` 作為 piped stdin，執行 `tellme`（`sc.pipeStdin({content})`）。

- [X] T007 [P] [BDD-RED] `When: the operator pipes "{content}" into tellme with the instruction "{instruction}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator pipes "{content}" into tellme with the instruction "{instruction}"`
  - Landing: `tests/e2e/steps/step_t007_chat_when_pipes_with_instruction.go`
  - 語意：以 `{instruction}` 為 positional argument，`{content}` 為 piped stdin，執行 `tellme "{instruction}"`。

- [X] T008 [P] [BDD-RED] `When: the operator pipes nothing into tellme`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator pipes nothing into tellme`
  - Landing: `tests/e2e/steps/step_t008_chat_when_pipes_nothing.go`
  - 語意：無 `-c`、無 positional prompt；以**空** piped stdin（立即 EOF）執行 `tellme`。

- [X] T009 [P] [BDD-RED] `When: the operator runs tellme's diagnostic with "{content}" piped in`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator runs tellme's diagnostic with "{content}" piped in`
  - Landing: `tests/e2e/steps/step_t009_chat_when_diagnostic_with_piped.go`
  - 語意：以 `{content}` 為 piped stdin，執行 `tellme -d`；`-d` dispatch 先於任何 prompt 處理。

- [X] T010 [P] [BDD-RED] `Then: the request carried the piped content "{content}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried the piped content "{content}"`
  - Landing: `tests/e2e/steps/step_t010_chat_then_request_piped_content.go`
  - 語意：取 fake 最後一次請求 body，解碼 `messages[0].content` 後 `TrimSpace`，斷言等於 `{content}`。

- [X] T011 [P] [BDD-RED] `Then: the request carried the instruction "{instruction}" followed by the piped content "{content}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request carried the instruction "{instruction}" followed by the piped content "{content}"`
  - Landing: `tests/e2e/steps/step_t011_chat_then_request_instruction_then_content.go`
  - 語意：斷言記錄到的 request prompt 等於 `{instruction}` + `"\n"` + `{content}`（instruction 在前、piped content 在後）。

- [X] T012 [P] [BDD-RED] `Then: the captured standard output is exactly "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the captured standard output is exactly "{answer}"`
  - Landing: `tests/e2e/steps/step_t012_chat_then_captured_stdout_exact.go`
  - 語意：斷言 `sc.stdout` 等於 `{answer}` 加恰好一個結尾換行；不得有前置或額外結尾裝飾。

- [X] T013 [P] [BDD-RED] `Then: the captured standard output carries no terminal decoration`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the captured standard output carries no terminal decoration`
  - Landing: `tests/e2e/steps/step_t013_chat_then_captured_stdout_no_decoration.go`
  - 語意：斷言 `sc.stdout` 不含 ANSI escape 序列（例如不含 `\x1b[`）。

- [X] T014 [P] [BDD-RED] `Then: tellme completes the turn without waiting for terminal input`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme completes the turn without waiting for terminal input`
  - Landing: `tests/e2e/steps/step_t014_chat_then_completes_without_waiting.go`
  - 語意：斷言本次執行已完成（`sc.runErr` 為 nil；exit code 已擷取），且 prompt 來自 piped stdin（`sc.stdinSet`），即未等待 TTY。

### UNIT（pure-helper 單元測試，folding issue #14）

- [X] T015 [P] [UNIT] prompt 組合與 I/O-mode 選擇單元測試 `internal/cli/prompt_test.go`
  - Read:
    - `specs/plans/005-stdin-piping/research.md` -> Decision 3、Decision 4
    - `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests）
    - `internal/cli/prompt.go`（T004 骨架）
  - 撰寫表驅動測試覆蓋 `combinePrompt`：僅 args、僅 piped、args + `"\n"` + piped、多 args 以單一空格 join、最終 `TrimSpace`、空 piped + 無 arg 回空字串；以及 I/O-mode 選擇（TTY vs pipe × arg vs stdin）以注入的 TTY 判斷 / stdin 替身驅動，不需真實終端。

### Phase Review Gate

- [X] T016 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/piping-a-prompt.feature`
    - `specs/truth/features/cli/chat/piping-the-answer-out.feature`
    - `specs/truth/features/cli/chat/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
    - `internal/cli/prompt_test.go`
  - 檢驗：無 undefined step；9 個新 stepdef 皆已註冊；測試能編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/piping-a-prompt.feature

**Goal**: 實作 stdin piped prompt 讀取與組合（`args (joined) + "\n"` + piped content，再 trim；僅 prompt-turn 讀 stdin），使 `piping-a-prompt.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/piping-a-prompt.feature` -> `Feature: Piping a prompt into tellme`
- `specs/truth/features/cli/chat/dsl.md` -> When / Then（pipes content、pipes content with instruction、pipes nothing、diagnostic with piped、request carried piped content、request carried instruction followed by piped content）
- `specs/truth/features/cli/dsl.md` -> `tellme performs no network access`、`tellme exits successfully`
- `truth-delta.md` -> `/axb-dsl-refine`（chat ADD + chat/dsl.md MODIFY）
- `specs/truth/techstack.md` -> CLI Application（Prompt input / Terminal detection / Piped stdin read）
- `specs/plans/005-stdin-piping/research.md` -> Decision 1, 2, 3, 6
- `internal/cli/cli.go`、`internal/cli/prompt.go`、`cmd/tellme/main.go`

**Boundary**:
- 實作 `combinePrompt`（`strings.Join(args, " ")` + 非空 piped 時 `"\n"` + piped → `TrimSpace`）；`internal/cli` 僅在 **prompt-turn** 路徑讀 stdin（`io.LimitReader(stdin, 1<<20)`，1 MiB 上限），且僅當 stdin 非 TTY；`--version`／`-d` 不讀 stdin、不觸網；而「stdin 被 pipe 且無 positional prompt」時，boot fall-through 前仍會讀 stdin（有界、立即 EOF），故精確敘述為「非顯式模式 dispatch 路徑」。[PR #16 評審 grill Q2/Q1 修正 — 見 plan.md 同註]
- 不引入新依賴；不實作輸出契約（屬 Phase 4B）；不碰 provider 路徑。

**Test Scope**:
- `specs/truth/features/cli/chat/piping-a-prompt.feature`

- [X] T017 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T018 [BDD-REFACTOR] 在綠燈下整理 prompt 組合、stdin 讀取與 mode 選擇

## Phase 4B: ADD Feature File - cli/chat/piping-the-answer-out.feature

**Goal**: 實作 TTY-aware 輸出契約（answer 僅寫 stdout、恰好一個結尾換行、非 TTY 時抑制呈現、不等待終端輸入），使 `piping-the-answer-out.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/piping-the-answer-out.feature` -> `Feature: Piping the answer out`
- `specs/truth/features/cli/chat/dsl.md` -> Then（captured stdout is exactly、captured stdout carries no terminal decoration、completes the turn without waiting for terminal input）
- `specs/truth/features/cli/dsl.md` -> `tellme prints the provider's answer "{answer}"`、`tellme exits successfully`
- `truth-delta.md` -> `/axb-dsl-refine`（chat/dsl.md MODIFY）
- `specs/truth/techstack.md` -> CLI Application（Terminal detection）
- `specs/plans/005-stdin-piping/research.md` -> Decision 1, 5
- `internal/cli/cli.go`、`internal/cli/prompt.go`

**Boundary**:
- 本輪未計算 `isTTY(stdout)`（無呈現可抑制）；TTY seam 本輪接在 stdin，stdout probe 待引入呈現時才接。output contract pin 為「answer bytes verbatim ＋ 一個 CLI 追加的結尾換行、僅寫 stdout、絕不寫 stderr」。本輪無呈現可抑制，故可觀察位元不變，只 pin 契約。[PR #16 評審 grill Q1/Q5 修正；T019/T020 維持 [X]（Test Scope 已達成）]
- 不新增 renderer 或 `-r` flag；不碰 stdin 組合（屬 Phase 4A）。

**Test Scope**:
- `specs/truth/features/cli/chat/piping-the-answer-out.feature`

- [X] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T020 [BDD-REFACTOR] 在綠燈下整理 TTY-aware 輸出閘控與 stdout/stderr 分流

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、configuration、workspace、diagnostics、usage 全模組）

- [X] T021 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6` 與既有 CLI 契約（round 001/002/003/004）100% 維持綠燈；offline paths（`--version`、`-d`、no-prompt boot）不觸網；既有 E2E 場景（無 stdin 的子行程）行為不變。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Prompt input = positional + piped stdin / Terminal detection / Piped stdin read） | T004、T006–T009、T017（Foundational / Phase 3 / Phase 4A） | PASS |
| `specs/truth/techstack.md` -> CLI Application（Terminal detection 閘控輸出） | T004、T013、T019（Foundational / Phase 3 / Phase 4B） | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner scripted stdin） | T001、T002、T006–T009（Foundational / Phase 3） | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests 擴充 prompt 組合 + I/O-mode） | T005、T015（Foundational / Phase 3） | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（`golang.org/x/term` / renderer + `-r`） | 負向決策 -> T004、T019 Boundary（不採 x/term、不加 renderer/`-r`） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY ×3（`specs/truth/techstack.md`） | T001、T004、T005、T017、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD ×2（`chat/piping-a-prompt.feature`、`chat/piping-the-answer-out.feature`） | T003、T006–T016、T017–T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY ×1（`chat/dsl.md` 4 When + 5 Then） | T003、T006–T014、T016 | PASS |
| `research.md` -> Decision 1（stdlib char-device TTY 偵測） | T004、T013、T015、T019 | PASS |
| `research.md` -> Decision 2（bounded stdin read，1 MiB） | T001、T017 | PASS |
| `research.md` -> Decision 3（prompt combine） | T004、T007、T011、T015、T017 | PASS |
| `research.md` -> Decision 4（I/O-mode seam + unit tests，folding #14） | T004、T005、T015、T016 | PASS |
| `research.md` -> Decision 5（TTY-aware 輸出契約，pin） | T013、T019 | PASS |
| `research.md` -> Decision 6（僅 prompt-turn 讀 stdin；不加 `-r`） | T008、T009、T017 | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP、`/axb-data-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
