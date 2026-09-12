# Tasks: Rendered Output & Raw Flag (round 006)

**Plan Package**: `specs/plans/006-rendered-output-and-raw-flag`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 只做本輪新增技術（glamour 相依）的基礎建設與 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 1: Setup

**Goal**: 引入本輪唯一的第三方渲染相依（glamour），使 `go.mod` / `go.sum` 一致且可編譯，並以一個最小 smoke-test 證明版本可用。

- [ ] T001 加入 glamour 相依並跑 smoke-test
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Output rendering / Raw output flag）
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 1
  - 只做：`go get github.com/charmbracelet/glamour@v1.0.0`；`go mod tidy`；確認 `go.mod` / `go.sum` 一致；寫一個最小 smoke（例如在既有單元測試或臨時程式中 `glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithEmoji())` 對 `"**bold**"` 渲染且輸出非空）並執行 `go build ./...`。
  - 不做：不寫任何 renderer adapter、config 欄位、flag 或 stepdef（屬 Phase 2/3/4）。

---

## Phase 2: Foundational

**Goal**: 建立 renderer adapter、config 欄位、CLI wiring、harness predicate 與 6 個 stepdef 的落點骨架（Zero Shared Edits 原則，各步獨立檔案）。

- [ ] T002 建立 renderer adapter 落點骨架 `internal/ui/renderer.go`
  - Read:
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 1, 2, 4, 5, 6
    - `specs/truth/techstack.md` -> CLI Application（Output rendering）
  - 只做：宣告 `internal/ui` 套件與一個 `Renderer` seam（例如 `Render(markdown string, width int) (string, error)` 及/或 `Degraded() bool`），帶 glamour build options（`WithStandardStyle(resolveGlamourStyle())` + `WithEmoji()`）與 `resolveGlamourStyle()`（讀 `GLAMOUR_STYLE`）的 stub；宣告 `sanitizeForTerminal` 與 rendered/raw byte-handling 的函式簽名（stub）。
  - 不做：不實作渲染邏輯、不接 `internal/cli`、不寫斷言。

- [ ] T003 建立 config `WrapWidth` 落點骨架 `internal/config/config.go`
  - Read:
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 7
    - `specs/truth/techstack.md` -> Configuration（Rendered width）
    - `internal/config/config.go`
  - 只做：在 `Config` 新增 `WrapWidth int`（`yaml:"WRAP_WIDTH"`）；在 config 載入/解析處綁定環境覆寫 `TELL_ME_WRAP_WIDTH`（沿用既有 `TELL_ME_*` env-over-file precedence）；留下 `>= 0` 驗證的 hook 簽名（stub）。
  - 不做：不實作寬度套用（屬 T002/renderer）；不寫驗證錯誤字串內容（屬 Phase 4D）。

- [ ] T004 建立 CLI wiring 落點骨架 `internal/cli/cli.go`
  - Read:
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 3, 4
    - `specs/truth/techstack.md` -> CLI Application（Raw output flag / Terminal detection）
    - `internal/cli/cli.go`、`cmd/tellme/main.go`
  - 只做：新增 `-r`/`--raw` boolean flag 的解析骨架；宣告 renderer seam 與 stdout TTY probe seam 的注入參數（讓 `cli.Run`／`runTurn` 可注入），`cmd/tellme/main.go` 以真實 `os.*` 與真實偵測器呼叫；render/raw 模式選擇與 answer 輸出改以注入的 renderer 為介面的 stub。
  - 不做：不實作渲染/raw 行為、不變更既有 dispatch 與 exit-code（產品行為留 Phase 4）。

- [ ] T005 建立 E2E harness ANSI-aware predicate 骨架 `tests/e2e/harness/cmd_helper.go`
  - Read:
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Residual risks（ANSI-dependent assertions）
    - `specs/truth/techstack.md` -> Testing & Verification（E2E runner）
    - `tests/e2e/harness/cmd_helper.go`
  - 只做：新增可判斷「渲染後」與「原樣」的 stdout 斷言輔助（例如 `stripANSI`、`containsLiteralMarkdown`），供 Phase 3 stepdef 使用；保留既有 `Run`／`RunWithStdin` 行為。
  - 不做：不寫具體 stepdef 斷言；不碰 `internal/`。

- [ ] T006 建立 6 個新句 stepdef 獨立檔案骨架 `tests/e2e/steps/step_t007_*.go`–`step_t012_*.go`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（新句：`the operator starts tellme with the prompt "{prompt}" and the raw flag`、`the rendered width is "{width}"`、`the captured standard output is the rendered answer, not its raw Markdown`、`the captured standard output is wrapped so that no line is wider than {width} columns`、`the captured standard output contains the answer on a single line`）
    - `specs/truth/features/cli/configuration/dsl.md`（新句：`a well-formed configuration "{config_path}" whose rendered width is "{width}"`）
    - `tests/e2e/steps/register.go`
  - 只做：建立 6 個獨立檔案（各對應 T007–T012 的句），各自 `init()` 自我註冊空白 registrar。
  - 不做：不寫具體 step 實作邏輯；不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的句與本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：`chat` 新句在 `specs/truth/features/cli/chat/dsl.md`；`configuration` 新句在 `specs/truth/features/cli/configuration/dsl.md`。不得掃其他模組。
- 介面根 `specs/truth/features/cli/dsl.md`：`tellme explains on stderr that "{reason}"` 的 `片語詞彙` 新增第 10 個片語（純詞彙擴充，stepdef 語意不變，無需 ALIGN）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：無。本輪無既有句語意被改到需要改既有 stepdef（`chat/dsl.md` 既有句語意不變；`is exactly` / `carries no terminal decoration` 的 `必查` 僅為 rendered-vs-raw 分流的敘述澄清，斷言行為不變）。
- `[BDD-REMOVE]`：無。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef；完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 4/7）。
- 三者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> 本輪 5 個新句
- `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" whose rendered width is "{width}"`
- `truth-delta.md` -> `/axb-dsl-refine` ADD ×2（features）+ MODIFY（features/dsl rows）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 `internal/config/*_test.go`（寬度解析）與 `internal/cli/*_test.go`（render/raw 模式選擇），採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T007–T013 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T014 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T007 [P] [BDD-RED] `When: the operator starts tellme with the prompt "{prompt}" and the raw flag`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator starts tellme with the prompt "{prompt}" and the raw flag`
  - Landing: `tests/e2e/steps/step_t007_chat_when_start_with_raw_prompt.go`
  - 語意：無 `-c`；以 `-r`（raw）加 positional prompt 執行 `tellme -r "{prompt}"`，擷取 exit code、stdout、stderr 與 fake 記錄的請求。

- [ ] T008 [P] [BDD-RED] `Given: the rendered width is "{width}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the rendered width is "{width}"`
  - Landing: `tests/e2e/steps/step_t008_chat_given_rendered_width.go`
  - 語意：在子行程環境設 `TELL_ME_WRAP_WIDTH={width}`，讓有效渲染寬度解析為 `{width}`。

- [ ] T009 [P] [BDD-RED] `Then: the captured standard output is the rendered answer, not its raw Markdown`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the captured standard output is the rendered answer, not its raw Markdown`
  - Landing: `tests/e2e/steps/step_t009_chat_then_captured_is_rendered.go`
  - 語意：斷言擷取的 stdout **不含** 腳本答案攜帶的字面 Markdown 強調記號（`**`、`_`），且答案的文字仍在（即已被渲染）。

- [ ] T010 [P] [BDD-RED] `Then: the captured standard output is wrapped so that no line is wider than {width} columns`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the captured standard output is wrapped so that no line is wider than {width} columns`
  - Landing: `tests/e2e/steps/step_t010_chat_then_wrapped_width.go`
  - 語意：去除 ANSI 後逐行計算顯示寬度，斷言每行不寬於 `{width}` 欄。

- [ ] T011 [P] [BDD-RED] `Then: the captured standard output contains the answer on a single line`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the captured standard output contains the answer on a single line`
  - Landing: `tests/e2e/steps/step_t011_chat_then_single_line.go`
  - 語意：斷言擷取的 stdout 以**單一未折行**帶出腳本答案（答案內不得被 tellme 插入換行）。

- [ ] T012 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" whose rendered width is "{width}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" whose rendered width is "{width}"`
  - Landing: `tests/e2e/steps/step_t012_config_given_rendered_width.go`
  - 語意：在 `{home}/{config_path}` 寫出一個可解析的 YAML，頂層 `WRAP_WIDTH` 為 `{width}`。

### UNIT（pure-helper 單元測試）

- [ ] T013 [P] [UNIT] 渲染寬度解析與 render/raw 模式選擇單元測試
  - Read:
    - `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 4, 7
    - `specs/truth/techstack.md` -> Configuration（Rendered width）、Testing & Verification（Pure-helper unit tests）
    - `internal/config/config.go`、`internal/cli/cli.go`
  - Landing: `internal/config/config_test.go`、`internal/cli/cli_test.go`（表驅動）
  - 撰寫：`WRAP_WIDTH` 解析（檔案值、`TELL_ME_WRAP_WIDTH` 覆寫優先、`0` = 預設、負值為錯誤）；render/raw 模式選擇（`-r` 與 `--raw` 等價；未給 `-r` = rendered）。

### Phase Review Gate

- [ ] T014 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/rendering-the-answer.feature`、`controlling-the-rendered-width.feature`、`piping-the-answer-out.feature`
    - `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/configuration/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`
  - 檢驗：無 undefined step；6 個新 stepdef 皆已註冊；測試可編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/rendering-the-answer.feature

**Goal**: 實作預設渲染輸出（glamour、全流向渲染、LaTeX 淨化、初始化失敗時降級為 raw），使 `rendering-the-answer.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/rendering-the-answer.feature` -> `Feature: Rendering the answer`
- `specs/truth/features/cli/chat/dsl.md` -> Then（captured standard output is the rendered answer）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（rendering-the-answer.feature）+ chat/dsl.md MODIFY
- `specs/truth/techstack.md` -> CLI Application（Output rendering）
- `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 1, 2, 4, 5, 6
- `internal/ui/renderer.go`、`internal/cli/cli.go`

**Boundary**:
- 預設路徑以 glamour 渲染 Markdown→ANSI；rendered 位元處理沿用 reference（trim 前後換行、非空則輸出 `out + "\n\n"`）；LaTeX→Unicode 淨化在渲染前套用；`glamour.NewTermRenderer` 失敗時降級為 raw 並在 stderr 至多一行 warn，不得讓 turn 失敗（ADR-007 parity）。
- 不實作 `-r`（屬 Phase 4C）；不實作寬度套用（屬 Phase 4B）。

**Test Scope**:
- `specs/truth/features/cli/chat/rendering-the-answer.feature`

- [ ] T015 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T016 [BDD-REFACTOR] 在綠燈下整理 renderer adapter、build options 與降級路徑

## Phase 4B: ADD Feature File - cli/chat/controlling-the-rendered-width.feature

**Goal**: 實作 `WRAP_WIDTH` / `TELL_ME_WRAP_WIDTH` 對渲染輸出的折行，使 `controlling-the-rendered-width.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/controlling-the-rendered-width.feature` -> `Feature: Controlling the rendered width`
- `specs/truth/features/cli/chat/dsl.md` -> Given（the rendered width is）、Then（wrapped so that no line is wider、contains the answer on a single line）、When（… and the raw flag）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（controlling-the-rendered-width.feature）+ chat/dsl.md MODIFY
- `specs/truth/techstack.md` -> Configuration（Rendered width）
- `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 7
- `internal/config/config.go`、`internal/ui/renderer.go`、`internal/cli/cli.go`

**Boundary**:
- `width > 0` 時以 `glamour.WithWordWrap(width)` 套用；`0`/未設 = 渲染器預設（80）；寬度只作用於 rendered 輸出，`-r` 下忽略。
- 不實作 `-r` 的分流本身（屬 Phase 4C，本 phase 只需讓「raw 不折行」的 Test Scope 成立，必要時可依賴 Phase 4C 的 flag 行為；若先行，則以最小 `-r` 旁路滿足）；不改既有 dispatch 與 exit-code。

**Test Scope**:
- `specs/truth/features/cli/chat/controlling-the-rendered-width.feature`

- [ ] T017 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T018 [BDD-REFACTOR] 在綠燈下整理寬度解析與折行套用

## Phase 4C: MODIFY Feature File - cli/chat/piping-the-answer-out.feature

**Goal**: 依 FR-007 修訂實作 `-r`/`--raw` raw 輸出（rendering 由 `-r` 單獨決定），使 `piping-the-answer-out.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/piping-the-answer-out.feature` -> `Feature: Piping the answer out`（Rule 1 已改為 raw 路徑）
- `specs/truth/features/cli/chat/dsl.md` -> When（… and the raw flag）、Then（is exactly、carries no terminal decoration）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（piping-the-answer-out.feature）
- `specs/truth/techstack.md` -> CLI Application（Raw output flag / Terminal detection）
- `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 3, 4
- `internal/cli/cli.go`

**Boundary**:
- 實作 `-r`/`--raw`：rendering 由 `-r` **單獨**決定（非由 TTY）；raw 路徑原樣輸出答案並在缺結尾換行時補一個；將 stdout TTY probe 接到 tellme **自身**呈現（`UseColor = isTTY && !raw`）。FR-007 修訂：非終端抑制僅涵蓋 tellme 自身 chrome，非答案渲染。**Round-005 既定 `isTTY(stdout)`「未接」敘述在此被取代**（T019/T020 為本 phase 的 GREEN/REFACTOR）。
- 不改 stdin 組合；不實作寬度（屬 Phase 4B）。

**Test Scope**:
- `specs/truth/features/cli/chat/piping-the-answer-out.feature`

- [ ] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 raw/rendered 閘控與 stdout/stderr 分流

## Phase 4D: MODIFY Feature File - cli/configuration/starting-with-a-configuration.feature

**Goal**: 實作無效渲染寬度（負值）的設定錯誤拒絕與一般設定錯誤片語，使 `starting-with-a-configuration.feature` 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature` -> `Rule: A run stops when the rendered width is invalid`
- `specs/truth/features/cli/configuration/dsl.md` -> Given（a well-formed configuration whose rendered width is）
- `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`（片語詞彙新增第 10 個 `the configuration is invalid`）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（configuration feature + dsl + root dsl）
- `specs/truth/techstack.md` -> Configuration（Rendered width）
- `specs/plans/006-rendered-output-and-raw-flag/research.md` -> Decision 7
- `internal/config/config.go`、`internal/cli/cli.go`、`internal/cli/exitcode.go`

**Boundary**:
- 於設定解析/驗證階段對 `WRAP_WIDTH < 0` 失敗，exit code 維持 `3`，並以新的第 10 個 frozen class phrase `the configuration is invalid` 回報（provider 專屬片語不變）。此為本輪唯一新增的 class phrase；root `dsl.md` 詞彙 9→10 已記錄。
- 不改既有 provider 驗證語意；不重排既有片語。

**Test Scope**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`

- [ ] T021 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T022 [BDD-REFACTOR] 在綠燈下整理寬度驗證與設定錯誤片語對應

## Phase 4E: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（含 FR-007 修訂後對 round-005 行為的覆蓋）。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、configuration、workspace、diagnostics、usage 全模組）

- [ ] T023 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6` 與既有 CLI 契約（round 001–005）維持綠燈；offline paths（`--version`、`-d`、no-prompt boot）不觸網；既有 E2E 場景（無 stdin 的子行程）行為不變；`go mod tidy` 後 module graph 一致。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Output rendering = glamour；`WithStandardStyle`+`GLAMOUR_STYLE`+`WithEmoji`；sanitize；ADR-007 degrade） | T001、T002、T015、T016 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Raw output flag `-r`/`--raw`；rendering 由 `-r` 單獨決定） | T004、T007、T019、T020 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Terminal detection 接 stdout，閘控自身呈現） | T004、T019 | PASS |
| `specs/truth/techstack.md` -> Configuration（Rendered width：`WRAP_WIDTH` + `TELL_ME_WRAP_WIDTH`，`>=0`，`0`=default，rendered-only） | T003、T008、T010、T012、T013、T017、T021 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（E2E runner ANSI-aware predicate；pure-helper 擴充 width/mode） | T005、T009、T013、T014 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（pty harness；TUI 修訂；glamour 已引入） | 負向決策 -> T004/T005 Boundary、T014（不採 pty） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY ×4（`specs/truth/techstack.md`） | T001–T005、T013、T015、T017、T019、T021 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD ×2（`chat/rendering-the-answer.feature`、`chat/controlling-the-rendered-width.feature`） | T006、T009–T011、T014、T015–T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/piping-the-answer-out.feature`） | T007、T019、T020、T023 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` 5 新句） | T006–T011、T014 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`configuration/starting-with-a-configuration.feature` + `configuration/dsl.md`） | T012、T021、T022 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（root `cli/dsl.md` 片語 9→10） | T021、T022 | PASS |
| `research.md` -> Decision 1（glamour renderer） | T001、T002、T015、T016 | PASS |
| `research.md` -> Decision 2（build options + GLAMOUR_STYLE） | T002、T015、T016 | PASS |
| `research.md` -> Decision 3（render/raw gate = `-r` only；stdout probe 閘控自身呈現） | T004、T019、T020 | PASS |
| `research.md` -> Decision 4（rendered/raw byte handling） | T002、T015、T019 | PASS |
| `research.md` -> Decision 5（LaTeX→Unicode sanitize） | T002、T015 | PASS |
| `research.md` -> Decision 6（ADR-007 degrade） | T002、T015、T016 | PASS |
| `research.md` -> Decision 7（wrap width：`WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH`，`>=0`，rendered-only） | T003、T008、T010、T012、T013、T017、T021 | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP、`/axb-data-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
