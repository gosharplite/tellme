# Tasks: tellme post-turn status lines — per-turn metrics + session summary with cost (round 018)

**Plan Package**: `specs/plans/018-post-turn-status-lines`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方相依**（pricing 為 config-only 的純算術；`tokens.log` 為 stdlib JSON；兩個 formatter 手寫於既有 `internal/ui`；`go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點骨架採獨立檔案設計（Zero Shared Edits 原則），為 Phase 3 並行分派消除同檔衝突。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth（本輪 **16 句為全新句 `ADD`**；另 **3 個 `[UNIT]`**）。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/ui/metrics.go`（the metrics-line + Ready-line formatters）、`internal/ui/pricing.go`（cost 算術）、`internal/config/**`（config-only `MODELS` pricing）、`internal/domain/llm/**`（`Usage` widened with `CachedTokens`/`ThinkingTokens`）、`internal/infrastructure/llm/openai/**`（parse usage details）、`internal/domain/history/**` + `internal/infrastructure/history/usage.go`（the per-mode `tokens.log` store）、`internal/agent/agentloop.go`（accumulate every call's usage）、`internal/cli/cli.go`（emit both lines after the answer；rotate the usage log on `--new`）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。**不動** the round-009 payload line、the round-017 turn chrome、the `-i` TUI surface、the class-phrase vocabulary（仍 11）。
- 本輪有 **1 句跨模組**新句（`the run reports no post-turn status`，落於介面根 `specs/truth/features/cli/dsl.md`，由 `chat`／`diagnostics`／`history` 共用）與 **1 個既有句群**（round-017 的 `the run shows no turn chrome` 被 `diagnostics`／`history` 沿用，語意不變、stepdef 已在）。

## Round-018 locked decisions (implementation constraints — MUST)

> 來自 operator interview + Clarify Q1（operator 拍板）與 `research.md` Decisions 1–9。

- **[LINE 2 = metrics of the call that just returned]** — `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>`（`M = prompt − cached`；`H = cached`；`C = completion`；`Th = thinking`），來源是 **the API call that just returned**（the last completion in a tool loop）。`Th:` **always** 呈現（含 `Th: 0`）。參考/見 T004/T009/T010/T012（research D1/D4/D6）。
- **[LINE 3 = Ready summary]** — `╰─⠿ Ready ($<lastCall> $<turn> $<session> - M: <sM> H: <sH> O: <sO> - <hit%>%)`（costs `$%.4f`、hit `%.1f%%`）。`O = C + Th`；`M/H/O` 為 **session-cumulative**；hit% = `H/(M+H)`。參考/見 T011–T016（research D3/D6/D7）。
- **[THREE COSTS]** — `$#1` = the just-returned call；`$#2` = the **whole turn**（prompt completion + each tool-loop call）；`$#3` = the **session**。`$#1 ≤ $#2 ≤ $#3`（**tool turn 為嚴格 `$#1 < $#2`**）。見 T011/T012/T013（research D3/D4）。
- **[PRICING = CONFIG-ONLY]** — 只有 config `MODELS: { <model>: { PRICING: { HIT, MISS, COMP } } }`；**沒有 built-in rates**；un-priced model → `$0.0000`（`$` group 仍呈現）。見 T005/T016（research D2）。
- **[STORAGE = per-mode `tokens.log`]** — 每次 API call 後 append 一筆 JSON record（`{timestamp, provider, model, cached_tokens, prompt_tokens, response_tokens, total_tokens, thinking_tokens, cost}`）；session 總和讀整檔；`--new` rotate/archive。見 T006/T014/T015/T023（research D5/D7）。
- **[PRESENCE]** — 兩行皆寫 `stderr`；出現在 **every prompt-bearing turn**（positional/piped、round-012 reader、`-i` submit）；**只在 the provider reports no usage 時抑制**；`stdout` byte-exact；plain text（no ANSI）；class-phrase 詞彙維持 **11**。見 T008/T018/T025。

---

## Phase 2: Foundational

**Goal**: 建立本輪測試層落點骨架（Zero Shared Edits）與產品碼落點，讓 Phase 3／Phase 4 不各自發明檔案或落點。只建立落點與載體，不寫行為。

- [ ] T001 建立 16 個 stepdef 獨立檔骨架（16 新句；Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 16 個新句）
    - `specs/truth/features/cli/dsl.md`（the root row）
    - `tests/e2e/steps/register.go`、`tests/e2e/harness/`（the merged-stream witness + the subprocess runner）
  - 只做：建立 16 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），檔名對應 Phase 3 task id：
    - `tests/e2e/steps/step_t003_chat_given_provider_reports_usage.go`
    - `tests/e2e/steps/step_t004_chat_given_tool_provider_reports_usage.go`
    - `tests/e2e/steps/step_t005_chat_given_pricing.go`
    - `tests/e2e/steps/step_t006_chat_given_usage_log_prior_call.go`
    - `tests/e2e/steps/step_t007_chat_when_new_with_prompt.go`
    - `tests/e2e/steps/step_t008_chat_then_metrics_reported.go`
    - `tests/e2e/steps/step_t009_chat_then_metrics_values.go`
    - `tests/e2e/steps/step_t010_chat_then_cost_reported.go`
    - `tests/e2e/steps/step_t011_chat_then_costs_equal.go`
    - `tests/e2e/steps/step_t012_chat_then_turn_cost_gt_request.go`
    - `tests/e2e/steps/step_t013_chat_then_session_includes_earlier.go`
    - `tests/e2e/steps/step_t014_chat_then_session_tokens.go`
    - `tests/e2e/steps/step_t015_chat_then_cache_share.go`
    - `tests/e2e/steps/step_t016_chat_then_zero_cost.go`
    - `tests/e2e/steps/step_t017_chat_then_status_trails_answer.go`
    - `tests/e2e/steps/step_t018_chat_then_no_post_turn_status.go`
  - 不做：不寫具體 arrange／斷言邏輯；不碰既有 step 檔。

- [ ] T002 落點產品碼骨架與 `[UNIT]` 落點檔骨架
  - Read:
    - `specs/plans/018-post-turn-status-lines/research.md` -> Decision 1, Decision 2, Decision 3, Decision 5, Decision 6
    - `specs/truth/techstack.md` -> CLI Application（Post-turn status lines (operator)；Session usage log）、Configuration（Model pricing）、Reasoning & Provider Transport
    - `specs/truth/data/data-model.dbml` -> `usage_record`
    - `internal/ui/status.go`（the sibling round-009 formatter — same package/home）
  - 只做：
    - 產品碼骨架（簽名 stub，空輸出/未接線）：`internal/ui/metrics.go`（the metrics-line + Ready-line formatter 簽名 + 常數 `Ready` glyph / `%` 格式）、`internal/ui/pricing.go`（cost 算術簽名 + `MODELS` rate 型別）、`internal/domain/llm`（`Usage` 加 `CachedTokens`/`ThinkingTokens` 欄位）、`internal/config`（`MODELS` pricing 型別宣告）、`internal/domain/history`（usage-record 型別 + store 介面）、`internal/infrastructure/history/usage.go`（append/sum/rotate 簽名）、`internal/agent`（accumulate 落點簽名）、`internal/cli/cli.go`（emit seam + `--new` rotate 落點）。
    - 測試落點骨架：`internal/ui/pricing_test.go`、`internal/ui/metrics_test.go`、`internal/infrastructure/history/usage_test.go`（stdlib `testing` 空殼）。
  - 不做：不寫渲染結果、不算 cost、不解析 usage details、不接線 `cli.go`、不改 the round-009 payload-line formatter、不引入任何相依。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 **16 句為全新句（`ADD`）**。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪 **15 句**在 `specs/truth/features/cli/chat/dsl.md`；**1 句**（`the run reports no post-turn status`）在介面根 `specs/truth/features/cli/dsl.md`。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。
- 本輪斷言形式：the metrics line 與 the Ready line 於 `stderr` 的 **presence + values**（timestamp 以 pattern）；costs 的 **關係**（equal / greater / includes-earlier / zero）；the two lines 的 **ordering**（trailing the answer，依 **merged** witness）；the negative Then 對 `stdout`+`stderr` 的 **absence**。Given 的 arrange 會把 usage（含 cached/reasoning）與 config-only pricing 寫進 fixture。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **16 句**（`ADD`；4 Given + 1 When + 11 Then，含 1 root row）。
- `[UNIT]`：**3 個**非 DSL 單元斷言（pricing 算術；兩個 formatter；usage-record round-trip + accumulation）。
- 兩種 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the run reports the token metrics of the request that just completed`
  -> `the reported metrics line shows {miss} missed, {cached} cached, {completion} completed, and {thinking} reasoning tokens`
  -> `the run reports the cost of the request, the turn, and the session`
  -> `the reported request, turn, and session costs are equal`
  -> `the reported turn cost is greater than the request cost`
  -> `the reported session summary includes the earlier call`
  -> `the run reports the session's missed, cached, and output tokens`
  -> `the run reports the share of the prompt served from cache`
  -> `the run reports a zero cost`
  -> `the post-turn status trails the answer`
  -> （Givens）`a configured provider "{provider}" whose endpoint answers with "{answer}" and reports the token usage:`、`… asks tellme to read "{path}" and then answers with "{answer}" and reports the token usage:`、`the configuration prices the active model with hit "{hit}", miss "{miss}", and completion "{comp}" per million tokens`、`the session's usage log already records a prior call`
  -> （When）`the operator starts a fresh session with "--new" and the prompt "{prompt}"`
- `specs/truth/features/cli/dsl.md` -> `the run reports no post-turn status`（本輪新增的跨模組 root row；class-phrase 詞彙維持 11）
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-post-turn-status.feature`）+ MODIFY（`chat/dsl.md` +15 rows、root `cli/dsl.md` +1 row、`diagnostics/version-and-setup-diagnostic.feature`、`history/inspecting-the-session-history.feature`、`history/starting-a-fresh-session.feature`）；`/axb-data-plan` ADD（`data/data-model.dbml` `usage_record`）；`/axb-technical-research` MODIFY（`techstack.md`）；`/axb-api-plan` NOOP（`contracts/**`）
- `tests/e2e/steps/`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
- `internal/ui/status.go`、`internal/ui/metrics.go`、`internal/ui/pricing.go`（T002 落點）

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper。
- `[UNIT]` 落點：`internal/ui/pricing_test.go`、`internal/ui/metrics_test.go`、`internal/infrastructure/history/usage_test.go`。
- 不寫產品碼（除 T002 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T003–T021 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度；目標檔案互斥，符合 Zero Shared Edits）；T022 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型；4 Given + 1 When + 11 Then）

- [ ] T003 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with "{answer}" and reports the token usage:`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with "{answer}" and reports the token usage:`
  - Landing: `tests/e2e/steps/step_t003_chat_given_provider_reports_usage.go`
  - 語意：寫 config（選 `{provider}`、endpoint 指 fake）；script fake 回 `{answer}` 並帶 JSON `usage`（`prompt_tokens`=prompt、`completion_tokens`=completion、`total_tokens`=prompt+completion+thinking、details `prompt_tokens_details.cached_tokens`=cached、`completion_tokens_details.reasoning_tokens`=thinking）。

- [ ] T004 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}" and reports the token usage:`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `… asks tellme to read "{path}" and then answers with "{answer}" and reports the token usage:`
  - Landing: `tests/e2e/steps/step_t004_chat_given_tool_provider_reports_usage.go`
  - 語意：script fake 先回 `read_files` tool-call（`{path}`）**並帶** usage block，再回 final answer `{answer}` 並帶同一 usage block（同 T003）；**兩個** response 都帶 usage，使該 turn 累積兩筆 usage-bearing call，故 `$#1 < $#2`。

- [ ] T005 [P] [BDD-RED] `Given: the configuration prices the active model with hit "{hit}", miss "{miss}", and completion "{comp}" per million tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the configuration prices the active model with hit "{hit}", miss "{miss}", and completion "{comp}" per million tokens`
  - Landing: `tests/e2e/steps/step_t005_chat_given_pricing.go`
  - 語意：把 `MODELS: { <active model>: { PRICING: { HIT: {hit}, MISS: {miss}, COMP: {comp} } } }` 寫進 default config。

- [ ] T006 [P] [BDD-RED] `Given: the session's usage log already records a prior call`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the session's usage log already records a prior call`
  - Landing: `tests/e2e/steps/step_t006_chat_given_usage_log_prior_call.go`
  - 語意：在 `$TELL_ME_HOME/output/<mode>/tokens.log` append 一筆 prior call（固定 counts prompt 10 / cached 6 / completion 3 / thinking 2 + 非零 cost）。

- [ ] T007 [P] [BDD-RED] `When: the operator starts a fresh session with "--new" and the prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator starts a fresh session with "--new" and the prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t007_chat_when_new_with_prompt.go`
  - 語意：run `tellme --new "{prompt}"`；擷取 exit code、`stdout`、`stderr`、fake 記錄的 requests。

- [ ] T008 [P] [BDD-RED] `Then: the run reports the token metrics of the request that just completed`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reports the token metrics of the request that just completed`
  - Landing: `tests/e2e/steps/step_t008_chat_then_metrics_reported.go`
  - 語意：`stderr` 帶一行符合 `[HH:MM:SS] [<provider>] M: <n> H: <n> C: <n> Th: <n>`（timestamp 以 pattern；`<provider>` 為 active provider key）；**不**在 `stdout`。

- [ ] T009 [P] [BDD-RED] `Then: the reported metrics line shows {miss} missed, {cached} cached, {completion} completed, and {thinking} reasoning tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the reported metrics line shows {miss} missed, {cached} cached, {completion} completed, and {thinking} reasoning tokens`
  - Landing: `tests/e2e/steps/step_t009_chat_then_metrics_values.go`
  - 語意：metrics line 的 `M/H/C/Th` 等於 `{miss}/{cached}/{completion}/{thinking}`（`M = prompt − cached`）；`Th` 即使為 0 仍存在。

- [ ] T010 [P] [BDD-RED] `Then: the run reports the cost of the request, the turn, and the session`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reports the cost of the request, the turn, and the session`
  - Landing: `tests/e2e/steps/step_t010_chat_then_cost_reported.go`
  - 語意：`stderr` 帶一行符合 `╰─⠿ Ready ($<a> $<b> $<c> - M: <n> H: <n> O: <n> - <pct>%)`；**不**在 `stdout`。

- [ ] T011 [P] [BDD-RED] `Then: the reported request, turn, and session costs are equal`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the reported request, turn, and session costs are equal`
  - Landing: `tests/e2e/steps/step_t011_chat_then_costs_equal.go`
  - 語意：Ready line 的三個 `$` 相等（fresh session：request == turn == session）。

- [ ] T012 [P] [BDD-RED] `Then: the reported turn cost is greater than the request cost`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the reported turn cost is greater than the request cost`
  - Landing: `tests/e2e/steps/step_t012_chat_then_turn_cost_gt_request.go`
  - 語意：Ready line 的第二個 `$`（turn）嚴格大於第一個 `$`（last request）。

- [ ] T013 [P] [BDD-RED] `Then: the reported session summary includes the earlier call`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the reported session summary includes the earlier call`
  - Landing: `tests/e2e/steps/step_t013_chat_then_session_includes_earlier.go`
  - 語意：Ready line 的 session（`$` #3）與 token 總和包含 arranged prior call（嚴格大於當前 call alone）。

- [ ] T014 [P] [BDD-RED] `Then: the run reports the session's missed, cached, and output tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reports the session's missed, cached, and output tokens`
  - Landing: `tests/e2e/steps/step_t014_chat_then_session_tokens.go`
  - 語意：Ready line 帶 session token totals `M: <n> H: <n> O: <n>`。

- [ ] T015 [P] [BDD-RED] `Then: the run reports the share of the prompt served from cache`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reports the share of the prompt served from cache`
  - Landing: `tests/e2e/steps/step_t015_chat_then_cache_share.go`
  - 語意：Ready line 帶 cache-hit 百分比（`<pct>%`）。

- [ ] T016 [P] [BDD-RED] `Then: the run reports a zero cost`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the run reports a zero cost`
  - Landing: `tests/e2e/steps/step_t016_chat_then_zero_cost.go`
  - 語意：Ready line 的 `$` 為 `$0.0000`（該 model 無 `MODELS` pricing entry）。

- [ ] T017 [P] [BDD-RED] `Then: the post-turn status trails the answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the post-turn status trails the answer`
  - Landing: `tests/e2e/steps/step_t017_chat_then_status_trails_answer.go`
  - 語意：merged capture 中，the metrics line 與 the Ready line 皆出現在 the answer bytes **之後**。

- [ ] T018 [P] [BDD-RED] `Then: the run reports no post-turn status`
  - Read: `specs/truth/features/cli/dsl.md`（interface root）-> `the run reports no post-turn status`
  - Landing: `tests/e2e/steps/step_t018_chat_then_no_post_turn_status.go`
  - 語意：`stdout` 與 `stderr` 皆**不**帶 metrics line 或 `╰─⠿ Ready` line；用於 no-usage path 與 non-prompt path（`--version`、`-l`、prompt-less `--new`）。

### UNIT（非 DSL 的單元斷言）

- [ ] T019 [P] [UNIT] the cost arithmetic — `M/H/C/Th` → the three costs + hit-rate
  - Read:
    - `specs/plans/018-post-turn-status-lines/research.md` -> Decision 3
    - `specs/truth/techstack.md` -> Configuration（Model pricing）
    - `specs/truth/data/data-model.dbml` -> `usage_record`
  - 撰寫：由 config `MODELS` rates 與一筆 usage 算 per-call cost（`miss·MISS + hit·HIT + (completion+thinking)·COMP`/1e6）、turn 加總、session 加總、hit-rate `H/(M+H)` `%.1f%%`；un-priced model → 0；`O = C + Th`。
  - 落點：`internal/ui/pricing_test.go`。

- [ ] T020 [P] [UNIT] the two formatters — deterministic rendering
  - Read:
    - `specs/plans/018-post-turn-status-lines/research.md` -> Decision 6
    - `specs/truth/techstack.md` -> CLI Application（Post-turn status lines (operator)）
    - `internal/ui/status.go`（the sibling formatter — the round-009 payload line must stay unchanged）
  - 撰寫：斷言 the metrics line 的文字（`[HH:MM:SS] [<provider>] M: … H: … C: … Th: …`，`Th` 即使 0 仍呈現）與 the Ready line 的文字（`╰─⠿ Ready ($… $… $… - M: … H: … O: … - …%)`，costs `$%.4f`、hit `%.1f%%`），時刻由注入 clock 決定。
  - 落點：`internal/ui/metrics_test.go`。

- [ ] T021 [P] [UNIT] the usage record — JSON round-trip + per-call accumulation
  - Read:
    - `specs/plans/018-post-turn-status-lines/research.md` -> Decision 4, Decision 5
    - `specs/truth/data/data-model.dbml` -> `usage_record`
    - `specs/truth/techstack.md` -> CLI Application（Session usage log）
  - 撰寫：斷言 usage-record 的 append/read round-trip（欄位、排序、`total = prompt + completion + thinking`）、session 總和，及 the loop 對每次 call 累積 usage（`$#1 < $#2`）。
  - 落點：`internal/infrastructure/history/usage_test.go`。

### Phase Review Gate

- [ ] T022 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature`、`presenting-the-turn.feature`、`reporting-the-payload-status.feature`、`answering-a-single-prompt.feature`、`using-a-tool.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/harness/`、`tests/e2e/fakeprovider/`
    - `internal/ui/metrics.go`、`internal/ui/pricing.go`、`internal/config/**`、`internal/agent/agentloop.go`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；16 個新句 stepdef + 3 個 `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為（the lines 未發出、costs 未計算、the usage log 未落地、`--new` 未 rotate、no-usage 未抑制）。

---

## Phase 4A: ADD Feature File - cli/chat/presenting-the-post-turn-status.feature

**Goal**: 讓 `presenting-the-post-turn-status.feature` 全綠 — 在 prompt turn 的 answer 之後，於 `stderr` 發出 the metrics line（just-returned call 的 M/H/C/Th；`Th` 恆呈現）與 the Ready summary（three costs + session token totals + cache-hit rate），由 config-only `MODELS` pricing 與 per-mode `tokens.log` 支撐；無 usage 或 non-prompt path 不發出；`stdout` byte-exact。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature` -> `Feature: Presenting the post-turn status`
- `specs/truth/features/cli/chat/dsl.md` -> the 15 round-018 `chat` rows（T003–T017）
- `specs/truth/features/cli/dsl.md` -> `the run reports no post-turn status`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-post-turn-status.feature`）+ MODIFY（`chat/dsl.md`）+ `/axb-data-plan` ADD（`data/data-model.dbml`）+ `/axb-technical-research` MODIFY（`techstack.md`）
- `specs/plans/018-post-turn-status-lines/research.md` -> Decision 1, Decision 4, Decision 5, Decision 6, Decision 7, Decision 8
- `specs/truth/data/data-model.dbml` -> `usage_record`
- `specs/truth/techstack.md` -> CLI Application（Post-turn status lines (operator)；Session usage log）+ Configuration（Model pricing）

**Boundary**:
- 產品碼：`internal/domain/llm`（`Usage` + `CachedTokens`/`ThinkingTokens`）+ `internal/infrastructure/llm/openai`（parse `prompt_tokens_details`/`completion_tokens_details`）+ `internal/agent/agentloop.go`（accumulate every call's usage 並 expose）+ `internal/domain/history`/`internal/infrastructure/history/usage.go`（per-mode `tokens.log`：append per call、sum、rotate）+ `internal/ui/pricing.go` + `internal/ui/metrics.go`（the two formatters）+ `internal/cli/cli.go`（在 prompt turn 的 answer 之後發出兩行；`--new` rotate the usage log；無 usage 時抑制）+ `internal/config`（`MODELS` pricing）。
- `$#1` = the last-returned call、`$#2` = the whole turn、`$#3` = the session；`<provider>` = active provider key；costs `$%.4f`、hit `%.1f%%`。
- **不動** the round-009 payload line 的文字、the round-017 turn chrome、the `-i` TUI surface、the class-phrase 詞彙（維持 **11**）；`stdout` 維持 byte-exact；plain text（no ANSI）；**不加相依**。
- 落點採零共用編輯；the emit seam 只加在 the turn 的 post-answer 位置。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature`

- [ ] T023 [BDD-GREEN] 讓 Test Scope 全綠（並使 T019–T021 的 `[UNIT]` 轉綠）
- [ ] T024 [BDD-REFACTOR] 在綠燈下整理 formatter、pricing、usage-store 與 emit seam（含 `--new` rotate 與 no-usage 抑制的落點）

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact、class-phrase 詞彙 11、the round-009 payload line 文字不變、the round-017 turn chrome 不變、exit-code 表、offline paths、rounds 001–017 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [ ] T025 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 亦確認新的 no-post-turn-status carriers：`diagnostics/version-and-setup-diagnostic.feature`（`--version`）、`history/inspecting-the-session-history.feature`（`-l`）、`history/starting-a-fresh-session.feature`（prompt-less `--new`）皆斷言 the run reports no post-turn status。
  - **可偽性見證 (a)（metrics line；非真空）**：暫時停發 the metrics line，確認 `presenting-the-post-turn-status.feature` 的 `the run reports the token metrics of the request that just completed` 失敗；觀察到失敗即還原。
  - **可偽性見證 (b)（no-usage 抑制）**：暫時讓 the lines 在 the provider reports no usage 時仍發出，確認 `presenting-the-post-turn-status.feature` 的 `the run reports no post-turn status` 失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact（兩行只寫 `stderr`）；the round-009 payload line 文字不變；the round-017 turn chrome 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Post-turn status lines (operator)） | T002、T020、T023、T024 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session usage log） | T002、T021、T023、T024 | PASS |
| `specs/truth/techstack.md` -> Configuration（Model pricing） | T002、T005、T019、T023、T024 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（usage widened） | T002、T003、T004、T023 | PASS |
| `specs/truth/data/data-model.dbml` -> `usage_record`（ADD） | T002、T006、T019、T021、T023 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T002、T005、T019、T020、T021、T023 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T024 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` ADD（`data/data-model.dbml` `usage_record`） | T002、T006、T021、T023 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/presenting-the-post-turn-status.feature`） | T003–T018、T023、T024 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` +15 rows；root `cli/dsl.md` +1 row；`diagnostics`/`history` features） | T001、T003–T018、T022、T025 | PASS |
| `research.md` -> Decision 1（capture `cached`/`reasoning` from usage details） | T002、T003、T004、T009、T023 | PASS |
| `research.md` -> Decision 2（config-only `MODELS`；un-priced → 0） | T002、T005、T016、T019、T024 | PASS |
| `research.md` -> Decision 3（cost formula；three-cost semantics） | T010–T013、T019、T023 | PASS |
| `research.md` -> Decision 4（loop accumulates every call's usage） | T002、T017、T021、T023、T024 | PASS |
| `research.md` -> Decision 5（per-mode `tokens.log`；`--new` rotates） | T002、T006、T013、T021、T023、T024 | PASS |
| `research.md` -> Decision 6（hand-written formatters；`Th` always） | T002、T008、T009、T020、T023 | PASS |
| `research.md` -> Decision 7（session totals read the log；turn accumulates） | T013、T014、T015、T021、T023 | PASS |
| `research.md` -> Decision 8（suppress only on no usage） | T018、T023、T024、T025 | PASS |
| `research.md` -> Decision 9（unit + E2E；no new dependency；no pty） | T001、T019–T021、T025 | PASS |
| `research.md` -> Must-ask questions（BDD techstack godog；E2E + units；single CLI end） | T001、T025（既有 truth 未改） | PASS |
| `spec.md` -> US1–US3、`FR-001`–`FR-014`、`NFR-001`–`NFR-004` | T003–T021（對齊）、T023–T025（交付） | PASS |
| `spec.md` -> A3（config-only pricing）、A5（`Th` always）、A6（fallbacks）、A7（derived）、A10（`$` group always） | A3 → T005/T016/T019；A5 → T009/T020；A6 → T003/T004/T009；A7 → T019/T020；A10 → T016/T023 | PASS |
| `plan.md` -> Source-code structure（`internal/ui/{metrics,pricing}.go`、`internal/infrastructure/history/usage.go`、`internal/config`、`internal/agent`、`internal/cli`；no new module） | T002、T023、T024 | PASS |
| `plan.md` -> Scope notes（api NOOP；data ADD；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T022、T024、T025 | PASS |
| `/axb-clarify` Q1 拍板（config-only `MODELS`；un-priced → `$0.0000`） | T005、T016、T019 | PASS |
| operator interview 拍板（line 2/3 shapes；三 `$` 語意；line-2 source；`tokens.log` storage；presence；`Th` always） | T008–T018、T023、T024 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
