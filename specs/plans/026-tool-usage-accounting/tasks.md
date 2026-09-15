# Tasks: tellme tool-usage accounting — per-tool outcomes + an offline roll-up (round 026)

**Plan Package**: `specs/plans/026-tool-usage-accounting`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方模組**（the log uses the Go stdlib — `encoding/json` over `os`; `go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（**9 句 `ADD` → `[BDD-RED]`**；另 **3 個 `[UNIT]`**）。**無 `MODIFY`、無 `DELETE` 句。**
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/domain/history/usage.go`（the `ToolUsageSink` port + the `ToolUsageRecord`/outcome types）、`internal/agent/agentloop.go`（classify each executed call `ok`/`error`/`timeout` and record it via the injected sink）、`internal/infrastructure/history/tool_usage_store.go`（the file-backed adapter over `~/.tellme/tools-count.jsonl`）、`internal/ui/toolusage.go`（the pure report formatter）、`internal/cli/cli.go`（inject the sink + the `--tool-usage` offline path + the user-home seam）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**不動** the tool execution semantics/set、the round-018 token store、the post-turn status lines、the spinner、`stdout` bytes、the class-phrase vocabulary（仍 11）。

## Round-026 locked decisions (implementation constraints — MUST)

> 來自 operator 拍板（Q1 → 1：three-way `ok`/`error`/`timeout`；Q2 → Others：a global append-only `~/.tellme/tools-count.jsonl`；Q3 → 1：a dedicated offline reporting flag）與 `research.md` Decisions 1–6。

- **[OUTCOME TAXONOMY]** every **executed** invocation is classified **`ok` / `error` / `timeout`** from **structural** loop signals only — `error` = a non-nil tool error; `timeout` = a nil-error result whose per-call `ctx` deadline was exceeded (the round-024 FR-018 "stopped" path); `ok` = otherwise (incl. a bounded/truncated result). **No result-text sniffing.** 見 T006/T010/T015/T019（research D1）。
- **[GLOBAL LOG]** one JSON line per invocation `{"timestamp":"<RFC3339>","tool":"<wire name>","outcome":"<ok|error|timeout>"}` appended (`O_APPEND|O_CREATE`, 0o644) to **`~/.tellme/tools-count.jsonl`** (`os.UserHomeDir()`; the ONLY artifact outside `TellMeHome` — a **recorded divergence**). **Global** (across repos/envs/personas) and **never reset by `--new`**; **best-effort** (an I/O failure or an unresolvable home is a silent no-op that never breaks a turn; emits nothing on `stdout`/`stderr`). 見 T002/T007/T019（research D2）。
- **[COUNTING SEAM]** the loop records through an injected domain-port **`ToolUsageSink`** (`internal/domain/history`; **nil = no-op**) — the CLI builds the file-backed adapter and injects it (mirrors `gatewayFactory` / `newUsageStore`); `internal/agent` stays filesystem-free. 見 T001/T004/T019（research D3）。
- **[OFFLINE REPORT]** the **`--tool-usage`** flag (provisional name; the CLI contract owner pins the wording) prints, per **registered** tool in **registry order** (`list_files`, `read_files`, `get_tree`, `execute_command`), the invocations + the `ok`/`error`/`timeout` breakdown to `stdout` and exits — **strictly offline** (no provider, no stdin), dispatched **before** any prompt/stdin access. One streaming pass (O(N)). 見 T003/T009/T012/T013/T014/T017/T020（research D4）。
- **[HERMETIC + SPLIT]** the write resolves the user home, so the E2E harness and unit tests point **`HOME`** at a temporary directory (never the operator's real `~/.tellme`). The record + report are asserted **E2E**; the outcome classification (esp. the `timeout` leg) is **unit-pinned**. 見 T005/T006/T015/T021（research D5）。
- **[STDLIB / POSIX]** no new dependency; POSIX-only; a single small `O_APPEND` write is atomic across processes — **no `flock`**. 見 T002/T004/T021（research D6）。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）、the domain port + the store adapter + the report formatter + the CLI/loop wiring 落點，以及 the harness `HOME` seam，讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫行為。

- [ ] T001 建立 the `ToolUsageSink` port + the `ToolUsageRecord`/outcome 型別骨架
  - Read:
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 1, Decision 3
    - `specs/truth/data/data-model.dbml` -> `tool_usage_record`、`tool_usage_outcome`
    - `specs/truth/techstack.md` -> CLI Application（Tool-usage accounting）
    - `internal/domain/history/usage.go`（the round-018 `UsageStore` port — the port shape precedent）
  - 只做：於 `internal/domain/history/usage.go`（或新增 `tool_usage.go`）定義 `ToolUsageRecord`（`Timestamp`, `Tool`, `Outcome`；JSON tags 對齊 DBML）、an outcome 型別／列舉（`ok`/`error`/`timeout`），以及 port `ToolUsageSink interface { Record(ToolUsageRecord) error }`。純型別與介面，無 I/O。
  - 不做：不實作檔案 adapter、不寫 loop 分類邏輯、不加相依。

- [ ] T002 建立 the file-backed `ToolUsageSink` adapter 骨架（`internal/infrastructure/history/tool_usage_store.go`）
  - Read:
    - `specs/truth/data/data-model.dbml` -> `tool_usage_record`
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 2
    - `specs/truth/techstack.md` -> CLI Application（Tool-usage accounting）
    - `internal/infrastructure/history/usage_store.go`（the round-018 adapter — the JSON-Lines + best-effort precedent）
  - 只做：新增 `tool_usage_store.go` — resolve `~/.tellme/tools-count.jsonl` via `os.UserHomeDir()`; `Record`（`MkdirAll ~/.tellme` then `O_APPEND|O_CREATE|O_WRONLY` append one JSON line）; `Load()`（read the whole file into records; missing file = empty, not an error）; a pure `Aggregate(records, registeredTools)` helper placeholder（deterministic registry order）。Nil-safe / best-effort signature 就位。
  - 不做：不接上 loop、不寫報告格式、不加相依；不觸碰 `usage_store.go`（round-018 log 不動）。

- [ ] T003 建立 the pure report formatter 骨架（`internal/ui/toolusage.go`）
  - Read:
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 4
    - `internal/ui/metrics.go`、`internal/ui/status.go`（the pure-formatter + injected-clock precedent）
  - 只做：新增 `internal/ui/toolusage.go` — 一個 pure function 的簽名（輸入：per-tool aggregate + 固定工具順序；輸出：the report text），空殼；不改既有 `internal/ui` 檔。
  - 不做：不真的排版、不讀檔、不加相依。

- [ ] T004 建立 the CLI + loop wiring 骨架（`internal/cli/cli.go`、`internal/agent/agentloop.go`）
  - Read:
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 1, Decision 3, Decision 4
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Tool-usage accounting）
    - `internal/agent/agentloop.go`（`AgentLoop` struct — the `Observer`/`Now` seam precedent；the per-call `context.WithTimeout`）
    - `internal/cli/cli.go`（`newUsageStore` / `parseFlags` / the dispatch precedence `--version` → `-d` → `-l` → prompt → boot）
  - 只做：
    - `internal/agent/agentloop.go`：`AgentLoop` 新增 `ToolUsage history.ToolUsageSink` 欄位（nil = no-op）；在 `Run` 的 per-call 區塊預留 classify+record 的落點（空動作、不寫判定）。
    - `internal/cli/cli.go`：新增 `--tool-usage` flag 於 `options` + `parseFlags`；在 dispatch 加入 the offline report branch（落點函式空殼，置於 prompt/stdin 之前）；一個 user-home seam（`os.UserHomeDir` 包一層 var，供測試注入）；construct `AgentLoop` 時注入 `newToolUsageStore` 的 sink（一新 var factory）。
  - 不做：不真的分類、不真的寫檔、不真的排版報告、不改 the tool semantics/set、不改 the round-018 usage store / the post-turn lines / the spinner、不改 `stdout` bytes 或 class-phrase。

- [ ] T005 建立 stepdef + `[UNIT]` 落點骨架，以及 the E2E harness `HOME` seam
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 9 個新句：3 Given + 1 When + 5 Then）
    - `tests/e2e/steps/`（the self-registering stepdef pattern — `init()` registrar）
    - `tests/e2e/harness/`（the per-scenario temp home + subprocess env）
    - `internal/agent/agentloop_test.go`、`internal/infrastructure/history/usage_test.go`、`internal/ui/metrics_test.go`（the `[UNIT]` landing precedent）
  - 只做：建立獨立落點檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t006_chat_given_tool_usage_log.go`
    - `tests/e2e/steps/step_t007_chat_given_lists_missing_folder.go`
    - `tests/e2e/steps/step_t008_chat_given_hanging_command.go`
    - `tests/e2e/steps/step_t009_chat_when_reviews_tools.go`
    - `tests/e2e/steps/step_t010_chat_then_log_outcomes.go`
    - `tests/e2e/steps/step_t011_chat_then_log_empty.go`
    - `tests/e2e/steps/step_t012_chat_then_review_outcomes.go`
    - `tests/e2e/steps/step_t013_chat_then_review_never_used.go`
    - `tests/e2e/steps/step_t014_chat_then_review_all_zero.go`
    - `internal/agent/tool_usage_test.go`（the `[UNIT]` landing for the loop outcome classification; 空殼）
    - `internal/infrastructure/history/tool_usage_test.go`（the `[UNIT]` landing for the store round-trip + aggregation; 空殼）
    - `internal/ui/toolusage_test.go`（the `[UNIT]` landing for the report formatter; 空殼）
    並於 `tests/e2e/harness/` 加上 the **`HOME` seam**：把子程序的 `HOME` 指向 scenario 的臨時目錄（與 `TELL_ME_HOME` 並列），使 `~/.tellme` 落在 scenario 內且測試隱密（不觸碰真實 `~/.tellme`）；提供一個讀取 `$HOME/.tellme/tools-count.jsonl` 的 helper。
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；**9 句 `ADD`（`[BDD-RED]`）**。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：本輪 9 句都在同模組 `specs/truth/features/cli/chat/dsl.md`（**Round 026** 區塊）。介面根 `specs/truth/features/cli/dsl.md` **未改**（NOOP）。不得掃其他模組。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭自行發明。
- 本輪斷言形式：the log 位於 `$HOME/.tellme/tools-count.jsonl`（the harness `HOME` seam）；the report 為 `--tool-usage` 的 `stdout`；outcome 對應 `succeeded`→`ok`、`failed`→`error`、`ran out of time`→`timeout`。

**Markers**:
- `[BDD-ALIGN]`：**無**（round 026 無 `MODIFY` 句）。
- `[BDD-REMOVE]`：**無**（round 026 無 `DELETE` 句）。
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 stepdef（3 Given + 1 When + 5 Then）。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：**3 個**非 DSL 單元斷言（the loop outcome classification；the store round-trip + aggregation；the report formatter）。
- 所有 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the tool usage already records that the tool "{tool}" was used {count} times with the outcome "{outcome}"`
  -> `a configured provider "{provider}" whose endpoint lists a folder that does not exist and then answers with "{answer}"`
  -> `a configured provider "{provider}" whose endpoint runs a command that never returns within a short limit and then answers with "{answer}"`
  -> `the operator reviews how the tools have been used`
  -> `the tool usage shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  -> `the tool usage shows no tool has been used`
  -> `the review shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  -> `the review shows the tool "{tool}" was never used`
  -> `the review shows every tool with no uses`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/accounting-for-the-tool-use.feature`）+ MODIFY（`chat/dsl.md`）；`/axb-data-plan` ADD（`tool_usage_record`）；`/axb-technical-research` MODIFY（`techstack.md`）；`/axb-api-plan` NOOP
- `internal/domain/history/usage.go`、`internal/infrastructure/history/tool_usage_store.go`、`internal/ui/toolusage.go`、`internal/agent/agentloop.go`、`internal/cli/cli.go`（T001–T004 落點）

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- 不寫產品碼（除 T001–T004 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T006–T017 各派一個獨立 subagent（目標檔案互斥，符合 Zero Shared Edits）；T018 等全部回來再啟動 subagent 執行 review。

### BDD-RED — Given（新增 arrange 句；3 句）

- [ ] T006 [P] [BDD-RED] `Given: the tool usage already records that the tool "{tool}" was used {count} times with the outcome "{outcome}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the tool usage already records that the tool "{tool}" was used {count} times with the outcome "{outcome}"`
  - Landing: `tests/e2e/steps/step_t006_chat_given_tool_usage_log.go`
  - 語意：create `$HOME/.tellme/` if absent，append `{count}` 條 JSON lines `{"timestamp":"<RFC3339>","tool":"{tool}","outcome":"<mapped>"}` 到 `tools-count.jsonl`（`succeeded`→`ok`、`failed`→`error`、`ran out of time`→`timeout`）。

- [ ] T007 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint lists a folder that does not exist and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint lists a folder that does not exist and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t007_chat_given_lists_missing_folder.go`
  - 語意：script the fake 回一個 `list_files` tool-call（`path` 不存在 → 工具回非 nil error），再回答案 `{answer}`。

- [ ] T008 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint runs a command that never returns within a short limit and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint runs a command that never returns within a short limit and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t008_chat_given_hanging_command.go`
  - 語意：script the fake 回一個 `execute_command` tool-call（args 帶短 `timeout`（例 `1`）與不返回的 `command`（例 `sleep 60`）），再回答案 `{answer}`。

### BDD-RED — When（新增動作句；1 句）

- [ ] T009 [P] [BDD-RED] `When: the operator reviews how the tools have been used`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator reviews how the tools have been used`
  - Landing: `tests/e2e/steps/step_t009_chat_when_reviews_tools.go`
  - 語意：跑 `tellme --tool-usage`（no `-c`、no positional prompt、不讀 stdin）；擷取 exit code / stdout / stderr 與 the fake 的 recorded requests。

### BDD-RED — Then（新增斷言句；5 句）

- [ ] T010 [P] [BDD-RED] `Then: the tool usage shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the tool usage shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  - Landing: `tests/e2e/steps/step_t010_chat_then_log_outcomes.go`
  - 語意：讀 `$HOME/.tellme/tools-count.jsonl`，斷言其中 `{tool}` 恰有 `{ok}` 條 `ok`、`{error}` 條 `error`、`{timeout}` 條 `timeout`，且無其他 outcome。

- [ ] T011 [P] [BDD-RED] `Then: the tool usage shows no tool has been used`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the tool usage shows no tool has been used`
  - Landing: `tests/e2e/steps/step_t011_chat_then_log_empty.go`
  - 語意：讀 `$HOME/.tellme/tools-count.jsonl`，斷言檔案不存在或為 **零** 條記錄。

- [ ] T012 [P] [BDD-RED] `Then: the review shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the review shows the tool "{tool}" with {ok} successes, {error} failures, and {timeout} timeouts`
  - Landing: `tests/e2e/steps/step_t012_chat_then_review_outcomes.go`
  - 語意：於 `--tool-usage` 的 `stdout` 斷言列出 `{tool}` 且其 `{ok}`/`{error}`/`{timeout}` 對應。

- [ ] T013 [P] [BDD-RED] `Then: the review shows the tool "{tool}" was never used`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the review shows the tool "{tool}" was never used`
  - Landing: `tests/e2e/steps/step_t013_chat_then_review_never_used.go`
  - 語意：於 `--tool-usage` 的 `stdout` 斷言列出 `{tool}` 且其使用次數為零（a registered-but-unused tool 不得被省略）。

- [ ] T014 [P] [BDD-RED] `Then: the review shows every tool with no uses`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the review shows every tool with no uses`
  - Landing: `tests/e2e/steps/step_t014_chat_then_review_all_zero.go`
  - 語意：於 `--tool-usage` 的 `stdout` 斷言列出**每個** registered agent tool（`list_files`、`read_files`、`get_tree`、`execute_command`）且使用次數為零。

### UNIT（非 DSL 的單元斷言）

- [ ] T015 [P] [UNIT] the loop outcome classification（`ok` / `error` / `timeout`）
  - Read:
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 1, Decision 3
    - `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Tool-usage accounting）
    - `internal/agent/agentloop.go`（the per-call `context.WithTimeout` + the `AgentLoop` struct）
  - 撰寫：以注入的 fake gateway + fake registry + fake sink，斷言 classification — 工具回 error → `error`；工具回 nil-error **且** per-call deadline 逾時 → `timeout`（the FR-018 leg；產品回 `timeoutMarker`）；工具回一般結果 → `ok`（含 bounded/truncated 視為 `ok`）；並斷言每 call 恰一筆 `Record`、順序為 call order。
  - 落點：`internal/agent/tool_usage_test.go`。

- [ ] T016 [P] [UNIT] the store round-trip + the aggregation
  - Read:
    - `specs/truth/data/data-model.dbml` -> `tool_usage_record`、`tool_usage_outcome`
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 2, Decision 4
    - `internal/infrastructure/history/usage_store.go`（the round-018 round-trip precedent）
  - 撰寫：以臨時 `HOME`（injected user-home seam）斷言 — `Record` append 一行且 `Load` 逐字回來（JSON round-trip）；`Aggregate` 對固定工具順序（含零使用工具）給出正確的 `ok`/`error`/`timeout` counts；缺檔 → 空（非 error）；a read-only/不可寫的 home → `Record` best-effort 回錯且**不** panic。
  - 落點：`internal/infrastructure/history/tool_usage_test.go`。

- [ ] T017 [P] [UNIT] the report formatter（deterministic, every registered tool）
  - Read:
    - `specs/plans/026-tool-usage-accounting/research.md` -> Decision 4
    - `internal/ui/metrics.go`（the pure-formatter precedent）
  - 撰寫：斷言 the report text — 每個工具以 **registry 順序**出現（含零使用），含其 invocations 與 `ok`/`error`/`timeout`；輸出對同一 aggregate 決定性（無雜湊序）；空 aggregate → 全零。
  - 落點：`internal/ui/toolusage_test.go`。

### Phase Review Gate

- [ ] T018 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/accounting-for-the-tool-use.feature`
    - `specs/truth/features/cli/chat/dsl.md`（Round 026 區塊）、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`
    - `internal/domain/history/usage.go`、`internal/infrastructure/history/tool_usage_store.go`、`internal/ui/toolusage.go`、`internal/agent/agentloop.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；9 個新句 stepdef + 3 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（the classification 尚未接線、the store 尚未寫檔、the report 尚未排版、the `--tool-usage` path 尚未實作）。

---

## Phase 4A: ADD Feature File - cli/chat/accounting-for-the-tool-use.feature

**Goal**: 讓 `accounting-for-the-tool-use.feature` 全綠 — each executed tool call 被以 `ok`/`error`/`timeout` 記入 the global log；the record 跨 session（`--new` 不重置）；the operator 以 `--tool-usage` 離線檢視 roll-up。其餘行為（tool semantics/set、round-018 token store、post-turn lines、spinner、`stdout` bytes、class-phrase 詞彙）不變。

**Shared Must Read**:
- `specs/truth/features/cli/chat/accounting-for-the-tool-use.feature` -> `Feature: Accounting for how the tools are used`
- `specs/truth/features/cli/chat/dsl.md` -> the 9 round-026 rows（Round 026 區塊）
- `truth-delta.md` -> `/axb-dsl-refine` ADD + MODIFY；`/axb-data-plan` ADD；`/axb-technical-research` MODIFY
- `specs/plans/026-tool-usage-accounting/research.md` -> Decision 1–6
- `specs/truth/techstack.md` -> CLI Application（Agent tool loop；Tool-usage accounting）；Testing & Verification
- `specs/truth/data/data-model.dbml` -> `tool_usage_record`

**Boundary**:
- 產品碼：`internal/domain/history/usage.go`（port + record/outcome types）、`internal/agent/agentloop.go`（classify `ok`/`error`/`timeout` from `terr` + the per-call deadline；record via the injected sink；best-effort，silent）、`internal/infrastructure/history/tool_usage_store.go`（the `~/.tellme/tools-count.jsonl` adapter：append + Load + Aggregate）、`internal/ui/toolusage.go`（the report formatter）、`internal/cli/cli.go`（inject the sink；the `--tool-usage` offline path **before** prompt/stdin；the user-home seam）。
- 依 `research.md` Decisions 1–6；classification 只讀 `terr` 與 `ctx` deadline（**不** sniff 結果文字）；the log write 為 best-effort 且**不**輸出到 `stdout`/`stderr`；the report 為 registry 順序、strictly offline、單一 streaming pass。
- **不動** the tool execution semantics/set、the round-018 token store（`tokens.log` / `tokens.summary.json` / `tokens.archive.jsonl`）、the post-turn lines、the spinner、`stdout` bytes、the class-phrase 詞彙（維持 **11**）、the `-i` TUI surface；**不加相依**（stdlib）。
- 落點採零共用編輯。

**Test Scope**:
- `specs/truth/features/cli/chat/accounting-for-the-tool-use.feature`

- [ ] T019 [BDD-GREEN] 讓 Test Scope 全綠（並使 T015/T016/T017 的 `[UNIT]` 轉綠）
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 classification／store／report／CLI wiring 落點

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、the round-018 token store 不變、the post-turn lines 不變、the spinner 不變、the tool semantics/set 不變、exit-code 表、offline paths、`--new` 不重置 the global log、rounds 001–025 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP、`/axb-dsl-refine` root `cli/dsl.md` NOOP）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T021 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（outcome classification；非真空）**：暫時把 `timeout` 一律當 `ok`，確認 `accounting-for-the-tool-use.feature` 的 `the tool usage shows the tool "execute_command" with 0 successes, 0 failures, and 1 timeouts` 失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（the `--new` 不重置）**：暫時讓 `--new` 也清空 the global log，確認 `The record survives a fresh session` 的 `the tool usage shows the tool "read_files" with 3 successes, 0 failures, and 0 timeouts` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact；the round-018 token store 不變；the post-turn lines 不變；the spinner 不變；the tool semantics/set 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths 不變；the global log 於 `--new` 不被重置；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Tool-usage accounting） | T001、T002、T003、T004、T019、T020 | PASS |
| `specs/truth/techstack.md` -> CLI Application（CLI flag parsing — `--tool-usage`） | T004、T009、T019 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Agent tool loop — the counting seam） | T001、T004、T015、T019 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（round-026 assertions + HOME hermeticity） | T005、T015、T016、T017、T021 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（unbounded-log forward item） | 豁免（forward item；T021 只確認檔不重置） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T001–T004、T015、T016、T019 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T021 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` ADD（`tool_usage_record` + `tool_usage_outcome`） | T001、T002、T016、T019 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/accounting-for-the-tool-use.feature`） | T006–T014、T018、T019、T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — the 9 round-026 rows + note） | T005、T006–T014、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T018 review 確認 root 未被誤改） | PASS |
| `research.md` -> Decision 1（three-way outcome from structural signals） | T001、T006、T010、T015、T019 | PASS |
| `research.md` -> Decision 2（global append-only `~/.tellme/tools-count.jsonl`） | T002、T007、T016、T019 | PASS |
| `research.md` -> Decision 3（the injected `ToolUsageSink` counting seam） | T001、T004、T015、T019 | PASS |
| `research.md` -> Decision 4（the offline `--tool-usage` report） | T003、T009、T012、T013、T014、T017、T020 | PASS |
| `research.md` -> Decision 5（hermetic `HOME` + unit-pin the timeout leg） | T005、T006、T015、T021 | PASS |
| `research.md` -> Decision 6（stdlib / POSIX / best-effort / no `flock`） | T002、T004、T019、T021 | PASS |
| `spec.md` -> US1–US2、`FR-001`–`FR-011`、`NFR-001`–`NFR-004` | T006–T014（對齊）、T019–T021（交付） | PASS |
| `spec.md` -> 邊界情況（no tool；repeated tool；truncated=`ok`；unavailable tool；unresolvable home；empty report；write failure） | no tool → T011；repeated → T006；truncated → T015；unavailable → 現行行為（T018 確認）；unresolvable home → T016；empty report → T014/T017；write failure → T016 | PASS |
| `plan.md` -> Source-code structure（`domain/history`、`agent`、`infrastructure/history`、`ui`、`cli`；no new module） | T001–T004、T019、T020 | PASS |
| `plan.md` -> Scope notes（api NOOP；data ADD；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T002、T018、T020、T021 | PASS |
| operator 拍板（Q1 → 1 three-way；Q2 → Others global log；Q3 → 1 offline report） | Q1 → T001/T006/T010/T015；Q2 → T002/T007/T016；Q3 → T003/T009/T012–T014/T017 | PASS |
| `specs/truth/features/cli/chat/accounting-for-the-tool-use.feature` -> the reused root/chat rows（`the runtime home is …`、`tellme sends no request to the provider …`、`tellme exits successfully`） | T009、T018 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
