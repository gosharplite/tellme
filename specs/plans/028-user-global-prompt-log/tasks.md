# Tasks: user-global interactive prompt log (round 028)

**Plan Package**: `specs/plans/028-user-global-prompt-log`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 落點或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方模組**（stdlib only；`go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立測試共用元件、後續實作落點與落點骨架；每則寫「只做／不做」；測試落點採獨立檔案設計（Zero Shared Edits 原則）。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪受影響的自動化測試對齊最新版 truth。**本輪無 `DELETE` 句（無 `[BDD-REMOVE]`）**；有 **3 個 `MODIFY` 句 → `[BDD-ALIGN]`**（the three `chat` shared-prompt-log rows）與 **4 個 `ADD` 句 → `[BDD-RED]`**（the new seed rows）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/infrastructure/history/global_prompt_tracker.go`（the log path → `~/.tellme/global_prompts.jsonl` via `os.UserHomeDir()`；the seed-on-absent copy）＋ `internal/cli/cli.go`（the tracker construction seam）。**不動** the record shape、the newest-first-dedup read、the compaction policy、the `-i`-only write rule、the TUI chrome（`internal/ui`）、the tool surface、the round-018/026 logs。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。

## Round-028 locked decisions (implementation constraints — MUST)

> 來自 operator 拍板（read from / save to `~/.tellme/global_prompts.jsonl`；if absent, copy the existing `output/global_prompts.jsonl`）與 `research.md` Decisions 1–7。

- **[USER-GLOBAL PATH]** the shared prompt log lives at **`~/.tellme/global_prompts.jsonl`** （resolved via the **CLI-injected user-home resolver** — `userHomeDir = os.UserHomeDir`, mirroring the round-026 `newToolUsageStore`；the child's `HOME` is a per-scenario temp dir — the round-026 precedent）。Both the suggestion **read** and the `-i` **record write** target it；`$TELL_ME_HOME/output/global_prompts.jsonl` is **no longer read or written** by tellme. 見 T001、T002、T012。
- **[SEED-ON-ABSENT]** when `~/.tellme/global_prompts.jsonl` is **absent**, copy `$TELL_ME_HOME/output/global_prompts.jsonl` into it **verbatim**（copy, **not** move；the source is left in place）；**never overwrite** an existing destination；a **missing source** starts empty（no error）。Best-effort — a seed I/O failure never aborts the prompt. Implemented as an **explicit `Seed(ctx) error`** invoked **once** at the composition root（**not** a constructor side effect — the tracker is built twice per `-i` run）。見 T002、T008、T014。
- **[SHAPE UNCHANGED]** the record `{"timestamp":"<RFC3339>","prompt":"<text>"}` per line、the append-only write、the newest-first-deduped read、and the compaction policy are **unchanged**；the `-i`-only write rule（a one-shot / piped run writes nothing）is **unchanged**；the TUI chrome、`stdout` byte-exactness、the class-phrase vocabulary（11）、and the round-018/026 logs are **unchanged**. 見 T012、T016。
- **[DIVERGENCE]** tellme **no longer shares** the log with `tell-me-go`（which keeps `$TELL_ME_HOME/output/...`）and skips its legacy locations — a recorded, operator-accepted divergence. 見 T012、T016。
- **[STDLIB / POSIX]** no new dependency；POSIX-only.

---

## Phase 2: Foundational

**Goal**: 建立本輪的測試共用元件（the user-global prompt-log path + the seed helpers）與產品／stepdef 落點骨架，讓 Phase 3／Phase 4 不各自發明落點。只建立載體與落點，不寫 stepdef 語意、不寫產品行為。

- [ ] T001 retarget the shared prompt-log test helper to the user-global path + add the seed helpers（`tests/e2e/steps/shared_prompt_log.go`）
  - Read:
    - `specs/truth/data/data-model.dbml` -> `prompt_log_entry`（the user-global location + the seed lifecycle）
    - `specs/plans/028-user-global-prompt-log/research.md` -> Decision 1, Decision 3, Decision 4, Decision 5
    - `tests/e2e/steps/shared_prompt_log.go`、`tests/e2e/steps/scenario_context.go`（`sc.userHomeDir`、`sc.home`、`userToolUsagePath` 的 round-026 seam）
  - 只做：
    - `promptLogPath(sc)` → `filepath.Join(sc.userHomeDir, ".tellme", "global_prompts.jsonl")`（the user-global log；the round-026 `userHomeDir` seam）。
    - 新增 `envPromptLogPath(sc)` = `filepath.Join(sc.home, "output", "global_prompts.jsonl")`（the seed source）；`appendEnvPromptLog(sc, prompt)`（append the seed-source line）；`ensureSharedPromptLogAbsent(sc)`（remove `~/.tellme/global_prompts.jsonl` if present）。
  - 不做：不改 stepdef 檔、不改 feature、不改產品碼、不加相依。

- [ ] T002 建立產品落點骨架 — the tracker path (injected resolver) + the explicit seed seam（`internal/infrastructure/history/global_prompt_tracker.go`、`internal/cli/cli.go`）
  - Read:
    - `specs/truth/techstack.md` -> CLI Application（Shared global prompt log）
    - `specs/plans/028-user-global-prompt-log/research.md` -> Decision 1, Decision 4, Decision 5, Decision 7
    - `internal/infrastructure/history/global_prompt_tracker.go`（the round-015 adapter）、`internal/domain/history/tracker.go`（the port doc）、`internal/cli/cli.go`（the `userHomeDir` seam `:176`、`newToolUsageStore` `:188`、the `NewGlobalPromptTracker(res.Home)` call sites `:210`,`:283`）
  - 只做：
    - 把 adapter 的 **destination** 改成經 **injected user-home resolver** 解析的 `~/.tellme/global_prompts.jsonl` — `NewGlobalPromptTracker(home string, userHome func() (string, error))`，mirroring the round-026 `newToolUsageStore`（`userHomeDir = os.UserHomeDir`，`cli.go:176`）。the **path join** 留在 adapter，只注入 resolver；the CLI 的 two call sites 傳入 `res.Home`（seed source）+ the shared `userHomeDir` seam，so CLI unit tests stay hermetic（`cli_test.go` overrides `userHomeDir`）。
    - 新增 **explicit** `Seed(ctx context.Context) error` 方法殼（best-effort、暫無實作行為）；the CLI 於 composition root（the interactive read 之前）呼叫**一次**（TD-2：the tracker is built **twice** per `-i` run — `cli.go:210` seeds suggestions, `cli.go:283` appends — 故 seed **不**放在 constructor）。
    - 更新 stale 產品檔的 doc comments（RF-2）：`internal/infrastructure/history/global_prompt_tracker.go`（the `globalPromptLogFile` const + the `GlobalPromptTracker`/`NewGlobalPromptTracker` type docs）、`internal/domain/history/tracker.go`（the port doc）→ the user-global path + the seed + the divergence。
  - 不做：不實作 the seed copy（Phase 4）、不改 the record shape / read / compaction / `-i`-only rule、不改 any other surface、不加相依。

- [ ] T003 建立 4 個新 stepdef 獨立落點骨架（Zero Shared Edits 原則）
  - Read:
    - `tests/e2e/steps/step_t011_chat_given_shared_log_holds.go`（the existing stepdef pattern）
    - `specs/truth/features/cli/chat/dsl.md`（the 4 new rows）
  - 只做：新增獨立落點骨架檔 `step_r028_given_env_log_holds.go`、`step_r028_given_shared_log_absent.go`、`step_r028_then_shared_log_also_holds.go`、`step_r028_then_shared_log_not_holds.go`（各含 `init()` 註冊與空殼函式）。使 Phase 3 的 4 個 `[P]` `[BDD-RED]` 任務目標檔案互斥。
  - 不做：不寫 stepdef 語意（Phase 3 寫）、不改既有 stepdef。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪受影響的自動化測試對齊最新版 truth：**3 個 `MODIFY` 句 → `[BDD-ALIGN]`** + **4 個 `ADD` 句 → `[BDD-RED]`**。不寫產品行為（除 T001–T003 的骨架）。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：the 7 句都在同模組 `specs/truth/features/cli/chat/dsl.md`。不得掃其他模組。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有的 3 句 stepdef 還在，但 `工作區` 已由 `$TELL_ME_HOME/output/global_prompts.jsonl` 改為 `~/.tellme/global_prompts.jsonl`（the shared helper 於 T001 已 retarget）；更新該句 stepdef 的過時註解並確認它經 retargeted helper 讀寫 the user-global log。
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 4 個新句的 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[BDD-REMOVE]`：**無**（無 `DELETE`）。
- 所有 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the shared prompt log already holds "{prompt}"`（Round 028）
  -> `the shared prompt log records the prompt "{prompt}"`（Round 028）
  -> `the shared prompt log still holds only "{prompt}"`（Round 028）
  -> `the environment prompt log already holds "{prompt}"`（Round 028）
  -> `the shared prompt log has not been created yet`（Round 028）
  -> `the shared prompt log also holds the earlier prompt "{prompt}"`（Round 028）
  -> `the shared prompt log does not hold "{prompt}"`（Round 028）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（the 3 retargeted rows）+ ADD（the 4 new rows）；`/axb-data-plan` MODIFY；`/axb-technical-research` MODIFY
- `tests/e2e/steps/shared_prompt_log.go`（T001 的 retargeted helper + the seed helpers）

**Boundary**:
- 一條 DSL 一個 task；只改該句的 stepdef／assertion／直接依賴的 helper。
- T004 另負責 the shared path helper 的 retarget（`shared_prompt_log.go`）；**該檔為三句共用**，故 T004 是唯一動該檔的任務，T005／T006 只動各自的 stepdef 檔（目標檔案互斥，可與 T007–T010 並行）。
- 不寫產品碼（除 T001–T003 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或尚未實作的產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T010 各派一個獨立 subagent（目標檔案互斥：T004 = `shared_prompt_log.go` + `step_t011_*`；T005 = `step_t024_*`；T006 = `step_t025_*`；T007–T010 = the four new `step_r028_*` files）；T011 等全部回來再啟動 subagent 來 review。

### BDD-ALIGN

- [ ] T004 [P] [BDD-ALIGN] `Given: the shared prompt log already holds "{prompt}"`（+ the shared path helper retarget）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log already holds "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_t011_chat_given_shared_log_holds.go`、`tests/e2e/steps/shared_prompt_log.go`
  - Landing: `tests/e2e/steps/shared_prompt_log.go`（**三句共用；唯一寫者**）+ `tests/e2e/steps/step_t011_chat_given_shared_log_holds.go`
  - 語意：the Given 現在寫入 `~/.tellme/global_prompts.jsonl`（the user-global log；creating it，so the first-use seed is skipped）。確認 `appendPromptLog` 經 retargeted `promptLogPath` 寫入；更新 stepdef／helper 的過時註解。

- [ ] T005 [P] [BDD-ALIGN] `Then: the shared prompt log records the prompt "{prompt}"`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log records the prompt "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_t024_chat_then_log_records.go`
  - Landing: `tests/e2e/steps/step_t024_chat_then_log_records.go`
  - 語意：`工作區` 改為讀 `~/.tellme/global_prompts.jsonl`（via retargeted helper）。更新過時註解；不改 assertion 邏輯。

- [ ] T006 [P] [BDD-ALIGN] `Then: the shared prompt log still holds only "{prompt}"`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log still holds only "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_t025_chat_then_log_unchanged.go`
  - Landing: `tests/e2e/steps/step_t025_chat_then_log_unchanged.go`
  - 語意：`工作區` 改為讀 `~/.tellme/global_prompts.jsonl`（via retargeted helper）。更新過時註解；不改 assertion 邏輯。

### BDD-RED

- [ ] T007 [P] [BDD-RED] `Given: the environment prompt log already holds "{prompt}"`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the environment prompt log already holds "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_r028_given_env_log_holds.go`（T003 skeleton）
  - 語意：寫入 `$TELL_ME_HOME/output/global_prompts.jsonl`（the seed source；create `output/` if absent，append one `{timestamp,prompt}` line via `appendEnvPromptLog`）。

- [ ] T008 [P] [BDD-RED] `Given: the shared prompt log has not been created yet`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log has not been created yet`（Round 028）
    - `tests/e2e/steps/step_r028_given_shared_log_absent.go`（T003 skeleton）
  - 語意：ensure `~/.tellme/global_prompts.jsonl` does not exist（remove it if present via `ensureSharedPromptLogAbsent`）。

- [ ] T009 [P] [BDD-RED] `Then: the shared prompt log also holds the earlier prompt "{prompt}"`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log also holds the earlier prompt "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_r028_then_shared_log_also_holds.go`（T003 skeleton）
  - 語意：`必查 權威狀態`：讀 `~/.tellme/global_prompts.jsonl`，確認它持有 `{prompt}`（the carry-over）。

- [ ] T010 [P] [BDD-RED] `Then: the shared prompt log does not hold "{prompt}"`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log does not hold "{prompt}"`（Round 028）
    - `tests/e2e/steps/step_r028_then_shared_log_not_holds.go`（T003 skeleton）
  - 語意：`必查 權威狀態`：讀 `~/.tellme/global_prompts.jsonl`，確認它不含 `{prompt}`（the no-overwrite witness）。

### Phase Review Gate

- [ ] T011 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature`、`specs/truth/features/cli/chat/carrying-over-the-environment-prompt-log.feature`
    - `tests/e2e/steps/shared_prompt_log.go`、`tests/e2e/steps/step_t011_*`、`step_t024_*`、`step_t025_*`、`tests/e2e/steps/step_r028_*`
    - `internal/infrastructure/history/global_prompt_tracker.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；3 個 `MODIFY` stepdef 已對齊（the shared log 讀寫目標為 `~/.tellme/global_prompts.jsonl`）；4 個 `ADD` stepdef 已註冊且可編譯；失敗僅限尚未實作之產品行為（the relocation / the seed 尚未接線）。

---

## Phase 4A: MODIFY Feature File - cli/chat/recording-the-shared-prompt-log.feature

**Goal**: 讓 `recording-the-shared-prompt-log.feature` 全綠 — an `-i` submission records to the **user-global** `~/.tellme/global_prompts.jsonl`；a one-shot / piped run writes nothing. **Shape、read semantics、the `-i`-only rule、class-phrase 詞彙、`stdout` 皆不變**。

**Shared Must Read**:
- `specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature` -> `Feature: Recording the shared prompt log`
- `specs/truth/features/cli/chat/dsl.md` -> `the shared prompt log already holds "{prompt}"`、`the shared prompt log records the prompt "{prompt}"`、`the shared prompt log still holds only "{prompt}"`（Round 028）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`recording-the-shared-prompt-log.feature` + the 3 rows）；`/axb-data-plan` MODIFY；`/axb-technical-research` MODIFY
- `specs/plans/028-user-global-prompt-log/research.md` -> Decision 1, Decision 2, Decision 5, Decision 6
- `specs/truth/techstack.md` -> CLI Application（Shared global prompt log）
- `specs/truth/data/data-model.dbml` -> `prompt_log_entry`

**Boundary**:
- 產品碼：`internal/infrastructure/history/global_prompt_tracker.go`（the log path → `~/.tellme/global_prompts.jsonl` via `os.UserHomeDir()`；the `seedIfAbsent` call before reads）＋ `internal/cli/cli.go`（the tracker construction seam）。
- 依 `research.md` Decisions 1／2／5／6；the record shape / read / compaction / `-i`-only rule 不變；an unresolvable/unwritable home degrades to a no-op。
- **不動** `internal/ui`、`internal/agent`、the tool surface、the round-018 token log、the round-026 tool-usage log、`stdout` bytes、class-phrase 詞彙。**不加相依**。
- 落點採零共用編輯。

**Test Scope**:
- `specs/truth/features/cli/chat/recording-the-shared-prompt-log.feature`

- [ ] T012 [BDD-GREEN] 讓 Test Scope 全綠（relocate the shared prompt log to the user-global path）
- [ ] T013 [BDD-REFACTOR] 在綠燈下整理 the tracker path resolution 與 the CLI construction seam

## Phase 4B: ADD Feature File - cli/chat/carrying-over-the-environment-prompt-log.feature

**Goal**: 讓 `carrying-over-the-environment-prompt-log.feature` 全綠 — on first use, the environment prompt log is carried into the absent shared prompt log（verbatim copy）；an existing shared prompt log is not overwritten；with no environment prompt log the shared log starts empty.

**Shared Must Read**:
- `specs/truth/features/cli/chat/carrying-over-the-environment-prompt-log.feature` -> `Feature: Carrying over the environment prompt log`
- `specs/truth/features/cli/chat/dsl.md` -> `the environment prompt log already holds "{prompt}"`、`the shared prompt log has not been created yet`、`the shared prompt log also holds the earlier prompt "{prompt}"`、`the shared prompt log does not hold "{prompt}"`（Round 028）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`carrying-over-the-environment-prompt-log.feature` + the 4 rows）；`/axb-data-plan` MODIFY
- `specs/plans/028-user-global-prompt-log/research.md` -> Decision 3, Decision 4, Decision 5, Decision 6
- `specs/truth/data/data-model.dbml` -> `prompt_log_entry`（the seed lifecycle）

**Boundary**:
- 產品碼：`internal/infrastructure/history/global_prompt_tracker.go` — implement the explicit `Seed(ctx)`（the seed-on-absent: when the user-global file is absent, copy `$TELL_ME_HOME/output/global_prompts.jsonl` verbatim；copy, not move；never overwrite；a missing source starts empty；best-effort — **swallow** any I/O error so the prompt is never aborted）。The CLI invokes `Seed(ctx)` **once** at the composition root（before the interactive read）。
- 依 `research.md` Decisions 3／4；the seed fires only on the `-i` path（the tracker is built there）；the copy is byte-identical；an existing destination is never replaced。
- **不動** the record shape / read / compaction、the relocation semantics（Phase 4A）、any other surface。**不加相依**。

**Test Scope**:
- `specs/truth/features/cli/chat/carrying-over-the-environment-prompt-log.feature`

- [ ] T014 [BDD-GREEN] 讓 Test Scope 全綠（implement the explicit `Seed(ctx)` + the first-use seed）
  - Read:
    - `specs/truth/features/cli/chat/carrying-over-the-environment-prompt-log.feature`
    - `specs/truth/features/cli/chat/dsl.md` -> the 4 seed rows（Round 028）
    - `specs/plans/028-user-global-prompt-log/research.md` -> Decision 3, Decision 4
    - `internal/infrastructure/history/global_prompt_tracker.go`
  - 只做：implement `Seed(ctx)`（the verbatim copy + the no-overwrite guard + the missing-source-is-fine）並接上 the single composition-root call；另加一個 **hermetic unit pin** 證明 the seed fires with **no submission**（i.e. on **opening** the prompt — RF-3；inject the resolver + a temp runtime home），使 the abort/empty-open seed path 也被釘住。
  - 不做：不改 the record shape / read / compaction、不改 the relocation（Phase 4A）、不改 any other surface。
- [ ] T015 [BDD-REFACTOR] 在綠燈下整理 the seed（best-effort copy + no-overwrite guard）

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact；class-phrase 詞彙 11；exit-code 表；the record shape / read / compaction / `-i`-only rule；the TUI chrome；the round-018/026 logs；the tool surface；rounds 001–027 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP）
- `specs/truth/techstack.md` -> CLI Application（Shared global prompt log / Tool-usage accounting）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T016 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（the read/write target moved；非真空）**：暫時把 `promptLogPath` 改回 `$TELL_ME_HOME/output/global_prompts.jsonl`（或把 the adapter path 改回），確認 the `recording`／`carrying-over` 的 Examples 失敗（a submission 不再被 the user-global read 觀察到）；觀察到失敗即還原。
  - **可偽性見證 (b)（the seed-on-absent）**：暫時移除 the seed copy，確認 the `carrying-over` 的第一個 Example（the `also holds the earlier prompt` Then）失敗；觀察到失敗即還原。
  - **可偽性見證 (c)（the never-overwrite guard）**：暫時讓 the seed 覆寫一個既有 the user-global log，確認 the `carrying-over` 的第二個 Example（the `does not hold` Then）失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact；the record shape / read / compaction / `-i`-only rule 不變；the TUI chrome 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；the round-018 token log / the round-026 tool-usage log 不變；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**；the user-global log never touches the operator's real `~/.tellme`（the `HOME` seam）。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Shared global prompt log — user-global path + seed + divergence） | T002、T012、T013、T016 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T002、T012、T016 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T016 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` MODIFY（`prompt_log_entry` location + seed lifecycle） | T001、T002、T008、T014、T016 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/recording-the-shared-prompt-log.feature` header） | T012、T004 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — the 3 retargeted rows） | T004、T005、T006、T011 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/carrying-over-the-environment-prompt-log.feature`） | T014、T007–T010 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/dsl.md` — the 4 new rows + the round-028 note） | T007、T008、T009、T010、T011 | PASS |
| `research.md` -> Decision 1（user-global path） | T001、T002、T004、T012 | PASS |
| `research.md` -> Decision 2（read + write move；env file no longer used） | T002、T012、T013 | PASS |
| `research.md` -> Decision 3（seed-on-absent; verbatim; copy-not-move; no-overwrite; missing→empty） | T001、T002、T008、T009、T010、T014 | PASS |
| `research.md` -> Decision 4（seed in the adapter, best-effort, before the first read） | T002、T014、T015 | PASS |
| `research.md` -> Decision 5（unresolvable/unwritable home → no-op） | T002、T014、T016 | PASS |
| `research.md` -> Decision 6（tell-me-go divergence；shape unchanged） | T012、T016 | PASS |
| `research.md` -> Decision 7（stdlib / POSIX / hermetic；temp HOME） | T001、T016 | PASS |
| `spec.md` -> US1（FR-001–FR-004）、US2（FR-005–FR-007）、FR-008 | T001–T010（對齊）、T012–T015（交付）、T016（回歸） | PASS |
| `spec.md` -> 邊界情況（multiple envs single seed；`~/.tellme/` created；deleted log re-seeds；`--new` does not clear the log） | the seed logic（T014）；`~/.tellme/` creation（T002）；T016 | PASS |
| `plan.md` -> Source-code structure（`infrastructure/history/global_prompt_tracker.go`、`cli/cli.go`；`ui`/`agent` unchanged；no new module） | T002、T012、T013 | PASS |
| `plan.md` -> Scope notes（api NOOP；data MODIFY；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T002、T011、T014、T016 | PASS |
| acceptance -> `sharing-prompts-across-environments.feature` | T004、T005、T006（the record + the untouched rules carry it） | PASS |
| acceptance -> `carrying-over-the-existing-prompt-history.feature` | T014（carried by the ADD feature Rules） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
