# Tasks: tellme turn spinner — a live progress indicator for prompt turns (round 019)

**Plan Package**: `specs/plans/019-turn-spinner`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方相依**（the spinner is hand-written in `internal/ui`; the CPU/memory sampling is stdlib POSIX; the diagnostic-stream terminal probe reuses `golang.org/x/term`, already a direct dependency since round 012; `go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（本輪 **10 句為全新句 `ADD`**；另 **3 個 `[UNIT]`**）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/spinner.go`（the presenter + the frame set + the label/elapsed formatter）、`internal/domain/metrics/**` + `internal/infrastructure/telemetry/system_metrics_{linux,darwin_cgo,darwin_nocgo}.go`（the dependency-free machine-wide POSIX CPU/memory sampler）、`internal/cli/cli.go`（own the spinner lifecycle around each waiting phase；gate on the diagnostic stream（`isatty(stderr) && !-r`）+ the `TELL_ME_FORCE_STDERR_TTY` seam）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**不動** `stdout` bytes、the round-009 payload line、the round-017 turn chrome、the round-018 post-turn lines、the class-phrase vocabulary（仍 11）、the `-i` TUI surface。
- 本輪有 **1 句跨模組**新句（`the run shows no progress spinner`，落於介面根 `specs/truth/features/cli/dsl.md`，由 `chat`／`diagnostics`／`history` 共用）與 **1 個既有句群**（the round-017 `the run shows no turn chrome` 被 `diagnostics`／`history`／`chat` 沿用，語意不變、stepdef 已在）。

## Round-019 locked decisions (implementation constraints — MUST)

> 來自 operator clarify（Q1=2、Q2=2）與 `research.md` Decisions 1–9。

- **[VISIBLE FORM]** `{frame}{status} ({elapsed}s)` — the reference braille frames (`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`) advancing on a ~200 ms injected ticker; a carriage-return redraw (`\r` + clear-line); the **first frame synchronous**; cleared with **no** trailing newline so the answer starts on the same line. 見 T002/T005/T013（research D1/D2）。
- **[LABELS]** the status tracks the waiting phase, each with a leading space — awaiting the model → ` Thinking [<model>]...` (the bracket omitted when the model is empty); executing tools → ` Executing [<tool>]...` (one), ` Executing tools [<a>, <b>]...` (several), ` Executing tools...` (names unavailable). 見 T006/T008/T009（research D3）。
- **[ELAPSED]** whole seconds; an in-place label update **preserves** the counter, a fresh interval restarts it. 見 T007（research D4）。
- **[RESOURCES]** ` [CPU: <c>% | MEM: <m>%]` (one decimal each) — **machine-wide** host CPU (Δ of `Σcpu − idle`) and memory, dependency-free POSIX sampling (Linux `/proc/stat` + `/proc/meminfo`; macOS `sysctl`/mach — cgo/nocgo), behind a domain port + an `internal/infrastructure/telemetry` adapter; **only** on the tool-execution state. 見 T002/T010/T014（research D5）。
- **[GATE]** drawn **only** when `isatty(stderr) && !-r` — mirroring the reference's `IsTerminalContext()` (`ui.stderr`); **no** stdout probe is wired (round-006 / PR #16 **Obs 1** stays **OPEN**); a diagnostic `TELL_ME_FORCE_STDERR_TTY` seam drives the branch without a pty. 見 T003/T015/T017（research D6）。
- **[LIFECYCLE]** the turn layer owns the spinner — start on each waiting phase, stop before any interleaved write (a tool-trace line, the answer, the post-turn lines), resume if waiting resumes, and stop via a deferred guard before completion / on failure (no residue). 見 T011/T017（research D7）。
- **[SURFACES]** the surface conjunction — prompt surface ∈ {positional, piped, round-012 reader} **AND** the FR-006 gate (`stderr` terminal, `-r` off); **not** the `-i` TUI; **not** the non-prompt paths; `stderr` only; `stdout` byte-exact; POSIX-only; no new dependency. 見 T012/T017/T021（research D6/D9）。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）與產品碼落點，讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫行為。

- [ ] T001 建立 10 個 stepdef 獨立檔骨架（10 新句；Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 9 個新句）
    - `specs/truth/features/cli/dsl.md`（the root row）
    - `tests/e2e/steps/register.go`、`tests/e2e/harness/`（the merged-stream witness + the subprocess runner）
  - 只做：建立 10 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t003_given_watching_terminal.go`
    - `tests/e2e/steps/step_t004_chat_given_two_tool_provider.go`
    - `tests/e2e/steps/step_t005_chat_then_spinner_shown.go`
    - `tests/e2e/steps/step_t006_chat_then_names_model.go`
    - `tests/e2e/steps/step_t007_chat_then_shows_elapsed.go`
    - `tests/e2e/steps/step_t008_chat_then_names_tool.go`
    - `tests/e2e/steps/step_t009_chat_then_names_tools.go`
    - `tests/e2e/steps/step_t010_chat_then_reports_resources.go`
    - `tests/e2e/steps/step_t011_chat_then_spinner_cleared.go`
    - `tests/e2e/steps/step_t012_chat_then_no_spinner.go`
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [ ] T002 落點產品碼骨架與 `[UNIT]` 落點檔骨架
  - Read:
    - `specs/plans/019-turn-spinner/research.md` -> Decision 1, Decision 2, Decision 5, Decision 6, Decision 7
    - `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)；Terminal detection）
    - `internal/ui/turn.go`（the sibling round-017 helper — same package/home）
    - `internal/cli/cli.go`（the turn flow + the existing stdin probe seam）
  - 只做：
    - 產品碼骨架（簽名 stub，空輸出/未接線）：`internal/ui/spinner.go`（the spinner presenter 簽名 + the frame set 常數 + the label/elapsed formatter 簽名）、`internal/domain/metrics/**`（the `SystemMetricsProvider` port 簽名）、`internal/infrastructure/telemetry/system_metrics_linux.go`、`system_metrics_darwin_cgo.go`、`system_metrics_darwin_nocgo.go`（the machine-wide CPU/memory sampler 簽名，behind the port seam）、`internal/cli/cli.go`（the spinner start/stop seam + the diagnostic-stream gate hook + the `TELL_ME_FORCE_STDERR_TTY` read）。
    - 測試落點骨架：`internal/ui/spinner_test.go`、`internal/infrastructure/telemetry/system_metrics_test.go`（stdlib `testing` 空殼）。
  - 不做：不畫任何 frame、不算 elapsed、不 sample CPU/mem、不接線 the turn flow、不改 the round-009/017/018 formatters、不引入任何相依。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 **10 句為全新句（`ADD`）**。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪 **8 句**在 `specs/truth/features/cli/chat/dsl.md`；**2 句**（`the diagnostics are shown at a terminal`、`the run shows no progress spinner`）在介面根 `specs/truth/features/cli/dsl.md`。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪斷言形式：the spinner line 於 `stderr` 的 **presence + label + elapsed + resources**（frame 以 set 比對，容忍 `\r` redraw）；clearing 的 **merged ordering**（answer 之後無 frame）；the negative Then 對 `stdout`+`stderr` 的 **absence**。**frame advancement over ticks** 為 `[UNIT]`（無 pty；a fast turn may render a single synchronous frame）。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **10 句**（`ADD`；2 Given + 7 Then + 1 root Then）。
- `[UNIT]`：**3 個**非 DSL 單元斷言（the spinner formatter + advancement；the CPU/memory samplers + resource segment；the gate resolution）。
- 兩種 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `a configured provider "{provider}" whose endpoint asks tellme to read "{path_a}" and "{path_b}" and then answers with "{answer}"`
  -> `the run shows the progress spinner while it waits`
  -> `the progress spinner names the model it is waiting for`
  -> `the progress spinner shows how long it has waited`
  -> `the progress spinner names the tool it is running`
  -> `the progress spinner names every tool it is running`
  -> `the progress spinner reports the machine's resource usage`
  -> `the progress spinner no longer appears once the answer is written`
- `specs/truth/features/cli/dsl.md`（interface root）-> `the diagnostics are shown at a terminal`、`the run shows no progress spinner`（本輪新增的跨模組 root rows；class-phrase 詞彙維持 11）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-progress-spinner.feature`）+ MODIFY（`chat/dsl.md` +8 rows、root `cli/dsl.md` +2 rows、`diagnostics/version-and-setup-diagnostic.feature`、`history/inspecting-the-session-history.feature`、`history/starting-a-fresh-session.feature`、`chat/presenting-the-turn.feature`）；`/axb-api-plan` NOOP（`contracts/**`）；`/axb-data-plan` NOOP（`data/**`）；`/axb-technical-research` MODIFY（`techstack.md`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
- `internal/ui/turn.go`、`internal/ui/spinner.go`、`internal/domain/metrics/**`、`internal/infrastructure/telemetry/system_metrics_*.go`（T002 落點）

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/ui/spinner_test.go`、`internal/infrastructure/telemetry/system_metrics_test.go`。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T015 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；目標檔案互斥，符合 Zero Shared Edits）；T016 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型；2 Given + 7 Then + 1 root Then）

- [ ] T003 [P] [BDD-RED] `Given: the diagnostics are shown at a terminal`
  - Read: `specs/truth/features/cli/dsl.md`（interface root）-> `the diagnostics are shown at a terminal`
  - Landing: `tests/e2e/steps/step_t003_given_watching_terminal.go`
  - 語意：把 `TELL_ME_FORCE_STDERR_TTY=1` 設進 subprocess 環境，使 tellme 的 **standard-error (`stderr`)** terminal probe 回報該 stream 為 terminal（round-019 seam）。

- [ ] T004 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path_a}" and "{path_b}" and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint asks tellme to read "{path_a}" and "{path_b}" and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t004_chat_given_two_tool_provider.go`
  - 語意：寫 config（選 `{provider}`、endpoint 指 fake）；script fake 回 **一個** response 帶 **兩個** `read_files` tool call（`{path_a}`、`{path_b}`），下一 request 回 final answer `{answer}`（one-round two-tool exchange）。

- [ ] T005 [P] [BDD-RED] `Then: the run shows the progress spinner while it waits`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run shows the progress spinner while it waits`
  - Landing: `tests/e2e/steps/step_t005_chat_then_spinner_shown.go`
  - 語意：`stderr` 帶一行 spinner line — 一個 braille frame（`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`）後接 phase status 與 `(<n>s)` elapsed segment（容忍 `\r` redraw）；**不**在 `stdout`。

- [ ] T006 [P] [BDD-RED] `Then: the progress spinner names the model it is waiting for`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner names the model it is waiting for`
  - Landing: `tests/e2e/steps/step_t006_chat_then_names_model.go`
  - 語意：spinner line 的 status 為 `Thinking [<model>]...`，`<model>` = active provider 的 configured `MODEL` attribute。

- [ ] T007 [P] [BDD-RED] `Then: the progress spinner shows how long it has waited`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner shows how long it has waited`
  - Landing: `tests/e2e/steps/step_t007_chat_then_shows_elapsed.go`
  - 語意：spinner line 帶 `(<n>s)` elapsed segment（whole seconds）。

- [ ] T008 [P] [BDD-RED] `Then: the progress spinner names the tool it is running`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner names the tool it is running`
  - Landing: `tests/e2e/steps/step_t008_chat_then_names_tool.go`
  - 語意：spinner line 的 status 為 single-tool 形式 `Executing [<tool>]...`，named tool = the run 正在執行的 tool。

- [ ] T009 [P] [BDD-RED] `Then: the progress spinner names every tool it is running`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner names every tool it is running`
  - Landing: `tests/e2e/steps/step_t009_chat_then_names_tools.go`
  - 語意：spinner line 的 status 為 several-tool 形式 `Executing tools [<a>, <b>]...`，命名 every tool（不得 collapse 成 single-tool 形式或漏一個）。

- [ ] T010 [P] [BDD-RED] `Then: the progress spinner reports the machine's resource usage`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner reports the machine's resource usage`
  - Landing: `tests/e2e/steps/step_t010_chat_then_reports_resources.go`
  - 語意：spinner line 的 tool-execution status 另帶一個 resource segment — 一個 `CPU` 百分比與一個 `MEM` 百分比（各一位小數）；**不**在 model-phase spinner 出現。

- [ ] T011 [P] [BDD-RED] `Then: the progress spinner no longer appears once the answer is written`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the progress spinner no longer appears once the answer is written`
  - Landing: `tests/e2e/steps/step_t011_chat_then_spinner_cleared.go`
  - 語意：**merged** capture 中，answer bytes **之後** 不再出現任何 braille frame（多帶 phase status）。依 round-010 merged-stream witness。

- [ ] T012 [P] [BDD-RED] `Then: the run shows no progress spinner`
  - Read: `specs/truth/features/cli/dsl.md`（interface root）-> `the run shows no progress spinner`
  - Landing: `tests/e2e/steps/step_t012_chat_then_no_spinner.go`
  - 語意：`stdout` 與 `stderr` 皆**不**帶 spinner frame。用於 non-terminal `stderr`、`-r`、`--version`、`-l`、prompt-less `--new`、`-i` submit path。

### UNIT（非 DSL 的單元斷言）

- [ ] T013 [P] [UNIT] the spinner formatter + frame advancement
  - Read:
    - `specs/plans/019-turn-spinner/research.md` -> Decision 1, Decision 2, Decision 3, Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)）
    - `internal/ui/turn.go`（the sibling formatter — same package/home）
  - 撰寫：斷言 the spinner line 的文字（`{frame}{status} ({elapsed}s)`，label 為 ` Thinking [<model>]...`，model 空時省略 bracket）、the frame 隨每個 tick 前進（注入 clock/ticker，**無** `time.Sleep`）、the elapsed 為 whole seconds 且 in-place label update 不歸零。
  - 落點：`internal/ui/spinner_test.go`。

- [ ] T014 [P] [UNIT] the machine-wide CPU/memory samplers + the resource segment
  - Read:
    - `specs/plans/019-turn-spinner/research.md` -> Decision 5
    - `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)；System metrics provider (telemetry)）
  - 撰寫：以注入的 readings 驅動 machine-wide sampler parser（Linux `/proc/stat` host `cpu ` 線 + `/proc/meminfo` 的固定取樣；macOS `sysctl`/mach 的固定取樣）→ 機器 CPU 百分比（Δ of `Σcpu − idle`）與 MEM 百分比（各 `%.1f`），以及 the resource segment 的字串；darwin cgo/nocgo 以 build tag 隔離。
  - 落點：`internal/infrastructure/telemetry/system_metrics_test.go`。

- [ ] T015 [P] [UNIT] the spinner gate resolution (forced stderr seam + `-r`)
  - Read:
    - `specs/plans/019-turn-spinner/research.md` -> Decision 6
    - `specs/truth/techstack.md` -> CLI Terminal detection（the diagnostic-stream gate）
    - `internal/cli/cli.go`（the existing stdin probe seam — mirror it for `stderr`）
  - 撰寫：斷言 gate 的解析 — `isatty(stderr) && !-r`；`TELL_ME_FORCE_STDERR_TTY` 覆寫生效；非 terminal `stderr` 或 `-r` 時 spinner 不啟。
  - 落點：`internal/ui/spinner_test.go`（或 `internal/cli` 的對應單元檔，若 gate 在 cli）。

### Phase Review Gate

- [ ] T016 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`、`presenting-the-turn.feature`、`diagnostics/version-and-setup-diagnostic.feature`、`history/inspecting-the-session-history.feature`、`history/starting-a-fresh-session.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`
    - `internal/ui/spinner.go`、`internal/domain/metrics/**`、`internal/infrastructure/telemetry/system_metrics_*.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；10 個新句 stepdef + 3 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（spinner 未畫、label/elapsed/resources 未產生、gate 未接線、clearing 未實作）。

---

## Phase 4A: ADD Feature File - cli/chat/presenting-the-progress-spinner.feature

**Goal**: 讓 `presenting-the-progress-spinner.feature` 全綠 — 在非 TUI prompt turn 的每個 waiting phase，於 `stderr` 發出 the live spinner（braille frame + phase label + elapsed；tool-execution 另帶 machine-wide CPU/MEM），只在 `isatty(stderr) && !-r` 時畫，於 answer 前讓位且無殘留；`stdout` byte-exact。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` -> `Feature: Presenting the progress spinner`
- `specs/truth/features/cli/chat/dsl.md` -> the 9 round-019 `chat` rows
- `specs/truth/features/cli/dsl.md` -> `the run shows no progress spinner`
- `truth-delta.md` -> `/axb-dsl-refine` ADD + MODIFY；`/axb-technical-research` MODIFY
- `specs/plans/019-turn-spinner/research.md` -> Decision 1, Decision 2, Decision 3, Decision 4, Decision 5, Decision 6, Decision 7
- `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)；Terminal detection）

**Boundary**:
- 產品碼：`internal/ui/spinner.go`（the presenter + the frame set + the label/elapsed formatter）、`internal/domain/metrics/**` + `internal/infrastructure/telemetry/system_metrics_{linux,darwin_cgo,darwin_nocgo}.go`（the dependency-free machine-wide POSIX CPU/memory sampler）、`internal/cli/cli.go`（own the lifecycle：每個 waiting phase 啟動、interleaved output 前停止、deferred stop；接線 the diagnostic-stream gate + `TELL_ME_FORCE_STDERR_TTY`）。
- 依 `research.md` Decisions 1–9；labels 具 leading space；elapsed 為 whole seconds；resources `%.1f`；gate `isatty(stderr) && !-r`；surface 需符合 FR-008 的 conjunction。
- **不動** `stdout` bytes、the round-009 payload line 文字、the round-017 turn chrome、the round-018 post-turn lines、the class-phrase 詞彙（維持 **11**）、the `-i` TUI surface；plain text（no ANSI）；**不加相依**。
- 落點採零共用編輯；the presenter 與 the gate seam 各自獨立檔。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`

- [ ] T017 [BDD-GREEN] 讓 Test Scope 全綠（並使 T013–T015 的 `[UNIT]` 轉綠）
- [ ] T018 [BDD-REFACTOR] 在綠燈下整理 the presenter／the sampler／the gate seam 與 the lifecycle 落點

## Phase 4B: MODIFY Feature Files - the boundary carriers

**Goal**: 讓五個被 MODIFY 的 carrier features 全綠 — 每個新增的 `the run shows no progress spinner`（含 failed-turn carrier）都由同一個產品改動（the diagnostic-stream gate + the surface exclusion + the deferred stop / clear-on-failure）滿足。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` -> `Feature: Reporting a failed provider request`（the new failed-turn spinner-clear Rule）
- `specs/truth/features/cli/chat/presenting-the-turn.feature` -> `Feature: Presenting the turn`（the `-i` Example）
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`
- `specs/truth/features/cli/history/inspecting-the-session-history.feature`
- `specs/truth/features/cli/history/starting-a-fresh-session.feature`
- `specs/truth/features/cli/dsl.md` -> `the run shows no progress spinner`
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（the five carriers）

**Boundary**:
- 只讓這五個 carrier 的新負向斷言轉綠；不改其餘既有斷言；產品改動仍限於 the spinner gate／surface／lifecycle 落點。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature`
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature`
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`
- `specs/truth/features/cli/history/inspecting-the-session-history.feature`
- `specs/truth/features/cli/history/starting-a-fresh-session.feature`

- [ ] T019 [BDD-GREEN] 讓本 phase 的 Test Scope 全綠
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 gate／surface／lifecycle 落點

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、the round-009 payload line 文字不變、the round-017 turn chrome 不變、the round-018 post-turn lines 不變、exit-code 表、offline paths、rounds 001–018 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP、`/axb-data-plan` NOOP）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T021 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 亦確認新的 no-spinner carriers：`diagnostics/version-and-setup-diagnostic.feature`（`--version`）、`history/inspecting-the-session-history.feature`（`-l`）、`history/starting-a-fresh-session.feature`（prompt-less `--new`）、`chat/presenting-the-turn.feature`（`-i`）、以及 `chat/reporting-a-failed-provider-request.feature`（failed turn）皆斷言 `the run shows no progress spinner`。
  - **可偽性見證 (a)（spinner presence；非真空）**：暫時停畫 the spinner，確認 `presenting-the-progress-spinner.feature` 的 `the run shows the progress spinner while it waits` 失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（gate）**：暫時讓 the spinner 在 non-terminal `stderr` 時仍畫出，確認 `presenting-the-progress-spinner.feature` 的 `the run shows no progress spinner` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact（the spinner 只寫 `stderr`）；the round-009 payload line 文字不變；the round-017 turn chrome 不變；the round-018 post-turn lines 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Turn progress spinner (operator)） | T002、T013、T014、T017、T018 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Terminal detection — the diagnostic-stream gate + the forced stderr seam） | T002、T015、T017 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（spinner assertions + samplers） | T001、T013、T014、T019 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T002、T013、T014、T015、T017 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T019 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-progress-spinner.feature`） | T003–T012、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` +8 rows；root `cli/dsl.md` +2 rows；`diagnostics`/`history`/`presenting-the-turn` features） | T001、T003–T012、T016、T019 | PASS |
| `research.md` -> Decision 1（hand-written `internal/ui` spinner） | T002、T013、T017 | PASS |
| `research.md` -> Decision 2（frames / cadence / the drawing primitive） | T002、T013、T017 | PASS |
| `research.md` -> Decision 3（phase labels with identifiers） | T006、T008、T009、T013、T017 | PASS |
| `research.md` -> Decision 4（the elapsed counter） | T007、T013、T017 | PASS |
| `research.md` -> Decision 5（dependency-free POSIX CPU/memory sampling） | T002、T010、T014、T017 | PASS |
| `research.md` -> Decision 6（the standard-output gate + the forced seam） | T003、T012、T015、T017、T019 | PASS |
| `research.md` -> Decision 7（the lifecycle placement） | T011、T017 | PASS |
| `research.md` -> Decision 8（unit + E2E；no pty） | T001、T013–T015、T019 | PASS |
| `research.md` -> Decision 9（POSIX-only；no new dependency） | T002、T014、T019 | PASS |
| `research.md` -> Must-ask questions（BDD techstack godog；E2E + units；single CLI end） | T001、T019（既有 truth 未改） | PASS |
| `spec.md` -> US1–US2、`FR-001`–`FR-011`、`NFR-001`–`NFR-005` | T003–T015（對齊）、T017–T019（交付） | PASS |
| `spec.md` -> A1–A4、A6–A9、A10 | A3/A4 → T013/T014；A6 → T003/T012；A7 → T005；A8 → T007/T013；A9 → T014/T019 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/spinner.go`、`internal/domain/metrics/**`、`internal/infrastructure/telemetry/system_metrics_*.go`、`internal/cli/cli.go`；no new module） | T002、T017、T018 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T016、T018、T019 | PASS |
| `/axb-clarify` Q1 拍板（full reference-parity label with identifiers） | T006、T008、T009、T013 | PASS |
| `/axb-clarify` Q2 拍板（keep the tool-execution CPU/MEM segment） | T010、T014 | PASS |
| operator 拍板（stop/resume A；`-i` out of scope；POSIX-only） | A → T011/T017；scope → T012/T021；POSIX → T014/T021 | PASS |
| `specs/truth/techstack.md` -> CLI Application（System metrics provider (telemetry)） | T002、T014、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/reporting-a-failed-provider-request.feature` failure carrier） | T011、T019、T021 | PASS |
| `research.md` -> Decision 6（the `stderr` diagnostic gate + the forced `stderr` seam; no stdout probe; Obs 1 open） | T003、T012、T015、T017、T021 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
