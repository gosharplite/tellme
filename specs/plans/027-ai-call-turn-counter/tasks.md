# Tasks: tellme turn counter counts AI-endpoint calls (round 027)

**Plan Package**: `specs/plans/027-ai-call-turn-counter`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/data/data-model.dbml`, `specs/truth/features/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task 不強制對應 `truth-delta.md` row。
- **本輪沒有新增第三方模組**（stdlib only；`go.mod`/`go.sum` 不動）。依 SOP **省略 Phase 1 `Setup`**；不得把 helper、fixture 或落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口與落點骨架；每則寫「只做／不做」；測試落點採獨立檔案設計（Zero Shared Edits 原則）。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪受影響的自動化測試對齊最新版 truth。**本輪無 `ADD` 句（無 `[BDD-RED]`）、無 `DELETE` 句（無 `[BDD-REMOVE]`）**；只有 **3 個 `MODIFY` 句** → **`[BDD-ALIGN]`**（the two chat tool-using exchange Givens + the history-module tool-using exchange Given），另 **1 個 `[UNIT]`**。the header Then row（`the turn is headed "Turn {number}"`）的 **stepdef 是泛型的**（解析 `{number}` 後比對；語意由 feature + 產品承擔）→ **只更新過時註解，不列為獨立 task**。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：`internal/domain/history/history.go`（`Entry` 新增整數 `Calls`）、`internal/infrastructure/history/*`（(de)serialize `calls`）、`internal/cli/cli.go`（header 號數 = Σ `prior[].Calls` + 1；append 時寫入 `Calls: len(result.Calls)`）。**不動** `internal/ui/turn.go`（formatter 不變）、`internal/agent/agentloop.go`（`AgentResult.Calls` 已足夠）。相位順序仍為「先對齊測試層（Phase 3），再於 Feature phase 補產品碼（GREEN）」。

## Round-027 locked decisions (implementation constraints — MUST)

> 來自 operator 拍板（tellme as-is；`N` = AI-endpoint-call index；a call = an inference round，retry 不計）與 `research.md` Decisions 1–6。

- **[UNIT = INFERENCE ROUND]** `<N>` = **Σ over the active session's `history_entry` `calls` + 1**（the running index of this prompt's first call）。A **call** = one `Gateway.Complete` inference round: a tool-less turn = 1；a turn that runs `k` tool rounds = `1 + k`；**provider-internal retries (backoff / empty-response) 不計**。the loop already exposes the turn's count as `len(AgentResult.Calls)`（round 018；one entry appended per successful call）。見 T003、T007、T010。
- **[PERSIST]** `history_entry` 新增整數 **`calls`**（the turn's inference-round count；written for every completed turn, ≥ 1）。a line **without** `calls`（legacy/arranged plain line）**counts as 1**。`--new` 會 archive the active file → Σ 歸零 → `Turn 1`。見 T001、T002、T007。
- **[SURFACE UNCHANGED]** exactly **one** header + **one** `╰─⠿ Ready` per prompt；plain text；no `/MAX_HISTORY_TURNS` denominator；` - <mode>` suffix；**不對每個 provider call 各發一組 header/footer**（the reference does；an operator-accepted divergence）。`stdout` byte-exact；the class-phrase vocabulary（11）、the pre-flight/measured payload lines、the post-turn metrics/`Ready`、the spinner、the tool-loop log line、the tool set/semantics 皆**不變**。見 T010、T011。
- **[STDLIB / POSIX]** no new dependency；POSIX-only。

---

## Phase 2: Foundational

**Goal**: 建立本輪產品承載（the persisted `calls` field + the counter seam）與一個 `[UNIT]` 落點，讓 Phase 3／Phase 4 不各自發明落點。只建立落點與載體，不寫行為。

- [X] T001 在 persisted `Entry` 新增整數欄位 `Calls`（`internal/domain/history/history.go`）
  - Read:
    - `specs/truth/data/data-model.dbml` -> `history_entry`（the widened record `{prompt, answer, calls, steps:[…]}`）
    - `specs/truth/techstack.md` -> CLI Application（Session history store）
    - `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 2
    - `internal/domain/history/history.go`（the existing `Entry`/`Step` types）
  - 只做：為 `history.Entry` 新增 `Calls int`（JSON tag `calls,omitempty`）。純欄位，無邏輯。
  - 不做：不改 store、不改 CLI、不改 `Step`。

- [X] T002 在 history store 序列化／解析 `calls`（`internal/infrastructure/history/*`）
  - Read:
    - `specs/truth/data/data-model.dbml` -> `history_entry`
    - `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 2, Decision 5
    - `internal/infrastructure/history/`（the JSONL entry struct + Load/Append）
  - 只做：讓 persisted record 帶上 `calls`（`omitempty`）；`Load` 時一個沒有 `calls` 的行解析為 `0`（the round-027 rule 把它視為 1 — the `0` sentinel is resolved by the counter，見 T003）。the append path 寫入 `Entry.Calls` 的值。
  - 不做：不改 `history_step`、不改 `-l` 投影、不改 archive 行為、不改 round-018 usage store。

- [X] T003 在 `internal/cli/cli.go` 建立 the counter seam 並接線（`Σ calls + 1`；persist `len(result.Calls)`）
  - Read:
    - `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 3, Decision 4, Decision 5
    - `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator)；Session history store）
    - `internal/cli/cli.go`（`runTurn` — the `emitTurnOpening(env, len(prior)+1, res.Mode)` call site and the `store.Append(history.Entry{…})` call site）
  - 只做：
    - 新增一個 pure helper `turnNumber(prior []history.Entry) int` = `Σ (e.Calls > 0 ? e.Calls : 1) + 1`（a missing/zero `calls` counts as 1 — Decision 5）。空殼/骨架先行（Phase 3 由 T007 unit-pin）。
    - 把 `emitTurnOpening` 的引數由 `len(prior)+1` 改為 `turnNumber(prior)`。
    - 把 `store.Append(history.Entry{Prompt: …, Answer: …, Steps: …})` 改為同時帶 `Calls: len(result.Calls)`（the turn's inference-round count）。
  - 不做：不改 `internal/ui/turn.go`（formatter 不變）、不改 chrome 的順序/格式、不改 payload/post-turn/spinner、不改 `--new` 行為（archive 本身即重置 the sum）、不改任何 exit code。

- [X] T004 建立 `[UNIT]` 落點骨架
  - Read:
    - `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 2, Decision 5
    - `internal/cli/turn_test.go`（the `internal/cli` unit-test precedent）
    - `internal/infrastructure/history/`（the existing history store test file）
  - 只做：新增獨立單元測試落點 `internal/cli/turn_counter_test.go`（空殼）；於既有 history-store 測試檔預留 the `calls` round-trip 落點（不新增共用檔衝突）。**不**在 Foundational 寫斷言邏輯。
  - 不做：不寫斷言、不改既有測試、不加相依。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪受影響的自動化測試對齊最新版 truth：**3 個 `MODIFY` 句 → `[BDD-ALIGN]`** + **1 個 `[UNIT]`**。不寫產品行為（除 T001–T003 的骨架）。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：`the session history already holds a tool-using exchange carrying the provider token "{token}"` 與 `… with no provider token` 在 `specs/truth/features/cli/chat/dsl.md`；`the session history already holds a tool-using exchange` 在 `specs/truth/features/cli/history/dsl.md`。the header Then row（`the turn is headed "Turn {number}" for the active mode`）在 `specs/truth/features/cli/chat/dsl.md`（stepdef 泛型 — 只更新註解）。不得掃其他模組。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。把既有 stepdef 對齊新版 `dsl.md` 該列的語意（the arrange 現在要帶 `calls: 2`）。stepdef 檔案已存在，不新增。
- `[BDD-REMOVE]`：**無**（無 `DELETE`）。
- `[BDD-RED]`：**無**（無 `ADD` 句 — the new feature Example 重用既有句）。
- `[UNIT]`：**1 個**（the counter computation + the `calls` round-trip）。
- 所有 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `the session history already holds a tool-using exchange carrying the provider token "{token}"`（Round 027）
  -> `the session history already holds a tool-using exchange with no provider token`（Round 027）
  -> `the turn is headed "Turn {number}" for the active mode`（Round 027）
- `specs/truth/features/cli/history/dsl.md`
  -> `the session history already holds a tool-using exchange`（Round 027）
- `specs/truth/features/cli/dsl.md` -> the interface-root row `the session history already holds the exchanges:`（unchanged；its plain lines count as 1）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（chat + history rows；feature reshape）；`/axb-data-plan` MODIFY（`history_entry.calls`）；`/axb-technical-research` MODIFY
- `internal/domain/history/history.go`、`internal/infrastructure/history/*`、`internal/cli/cli.go`（T001–T003 落點）

**Boundary**:
- 只改該句的 stepdef／arrange 欄位；不寫產品碼（除 T001–T003 已預留的骨架）。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或尚未實作的產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T005–T007 各派一個獨立 subagent（目標檔案互斥）；T008 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN

- [X] T005 [P] [BDD-ALIGN] `Given: the session history already holds a tool-using exchange …`（chat：token variant + no-token variant）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `the session history already holds a tool-using exchange carrying the provider token "{token}"`、`the session history already holds a tool-using exchange with no provider token`（both Round 027）
  - Landing: `tests/e2e/steps/step_t003_chat_given_signed_exchange.go`（+ `step_t004_chat_given_unsigned_exchange.go` reuses `writeToolUsingExchange`）
  - 語意：the written line must carry `calls: 2`（a tool-using turn makes two inference rounds）。將 `signedEntryFixture` 加上 `Calls int \`json:"calls"\`` 並在 `writeToolUsingExchange` 設為 `2`。順帶把 `step_t006_chat_then_headed_turn_number.go` 的過時註解（「completed-turn count + 1」）更新為 the AI-endpoint-call count（純註解，行為不變）。

- [X] T006 [P] [BDD-ALIGN] `Given: the session history already holds a tool-using exchange`（history module）
  - Read:
    - `specs/truth/features/cli/history/dsl.md` -> `the session history already holds a tool-using exchange`（Round 027）
  - Landing: `tests/e2e/steps/step_t022_history_given_tool_using_exchange.go`
  - 語意：the written `widenedEntry` line must carry `calls: 2`（a tool-using turn makes two inference rounds）；加上 `Calls int \`json:"calls"\`` 欄位並設為 `2`。（`-l` 投影仍只讀 prompt/answer — 不變。）

### UNIT

- [X] T007 [P] [UNIT] the counter computation + the persisted `calls` round-trip
  - Read:
    - `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 2, Decision 4, Decision 5
    - `specs/truth/data/data-model.dbml` -> `history_entry`
    - `internal/cli/cli.go`（the `turnNumber` helper）、`internal/domain/history/history.go`、`internal/infrastructure/history/*`
  - 撰寫（landing `internal/cli/turn_counter_test.go` + the history-store test）：`turnNumber` — empty `prior` → `1`；two entries `calls:1` → `3`；one entry `calls:2` → `3`；an entry with `Calls==0`（absent/legacy）counts as `1`（→ `2` for one such entry）；a mix。the store round-trip — an `Entry{Calls:2}` serialises a `calls` field and `Load` reads it back; a line **without** `calls` loads as `Calls==0`（→ counted as 1）。**不**改 `AgentResult`。
  - 不做：不改產品的 header 順序/格式。

### Phase Review Gate

- [X] T008 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/presenting-the-turn.feature`、`specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/history/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/step_t003_*`、`step_t004_*`、`step_t006_*`、`step_t022_*`
    - `internal/domain/history/history.go`、`internal/infrastructure/history/*`、`internal/cli/cli.go`
  - 檢驗：無 undefined step；3 個 `MODIFY` stepdef 已對齊（the two chat Givens + the history Given 帶 `calls: 2`）；`[UNIT]` 已註冊且可編譯；the header Then stepdef 泛型且註解已更新；the interface-root row 未動；失敗僅限尚未實作之產品行為（the counter 尚未接線、the store 尚未帶 `calls`）。

---

## Phase 4A: MODIFY Feature File - cli/chat/presenting-the-turn.feature

**Goal**: 讓 `presenting-the-turn.feature` 全綠 — the turn header 的號數以 **AI-endpoint calls**（Σ `calls` + 1）計；a tool-using history advances by its inference-round count；`--new` 自 `Turn 1` 起。**其餘 chrome 行為完全不變**（one header + one `Ready` per prompt；plain；`stdout` byte-exact；class-phrase 詞彙 11；payload/post-turn/spinner/tool-log 不變）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature` -> the Rule `The turn header counts the model requests made so far`（both Examples）
- `specs/truth/features/cli/chat/dsl.md` -> `the turn is headed "Turn {number}" for the active mode`（Round 027）+ the two tool-using exchange Givens（Round 027）
- `specs/truth/features/cli/history/dsl.md` -> `the session history already holds a tool-using exchange`（Round 027）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY（feature + chat/history rows）
- `specs/plans/027-ai-call-turn-counter/research.md` -> Decision 1–5
- `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator)；Session history store）
- `specs/truth/data/data-model.dbml` -> `history_entry`

**Boundary**:
- 產品碼：`internal/domain/history/history.go`（`Entry.Calls`）、`internal/infrastructure/history/*`（serialize/parse `calls`）、`internal/cli/cli.go`（`turnNumber` = Σ `calls` + 1；persist `Calls: len(result.Calls)`）。
- 依 `research.md` Decisions 1–5；a call = an inference round（`len(AgentResult.Calls)`）；retries 不計；a line without `calls` counts as 1；`--new` archive resets the sum。
- **不動** `internal/ui/turn.go`（the `╭─⠿ Turn <N> - <mode>` formatter 不變）、`internal/agent/agentloop.go`（`AgentResult.Calls` 已足夠）、the tool semantics/set、the round-018 token store、post-turn lines、spinner、`stdout` bytes、class-phrase 詞彙。**不加相依**。
- 落點採零共用編輯。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-turn.feature`

- [X] T009 [BDD-GREEN] 讓 Test Scope 全綠（並使 T007 的 `[UNIT]` 轉綠）
- [X] T010 [BDD-REFACTOR] 在綠燈下整理 `turnNumber`／store `calls`／CLI wiring 落點

## Phase 4B: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（`stdout` byte-exact；class-phrase 詞彙 11；exit-code 表；the chrome cadence/format（one header + one `Ready` per prompt）；the payload/post-turn lines；the spinner；the tool-loop log line；the tool semantics/set；the round-018 token store；rounds 001–026 全綠），並提供可偽性見證。

**Shared Must Read**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）
- `truth-delta.md` -> all owner rows（含 `/axb-api-plan` NOOP）

**Boundary**:
- 只跑回歸與見證，不新增產品行為；不得為了讓回歸通過而放寬任何既有 assertion。

**Test Scope**:
- `specs/truth/features/cli/**`

- [X] T011 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、`verify-cross-compile`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證 (a)（the number counts calls, not turns；非真空）**：暫時把 `turnNumber` 改回 `len(prior)+1`，確認 the `The turn header counts the model requests made so far` 的第二個 Example（a single tool-using earlier turn → `Turn 3`）失敗（會得到 `Turn 2`）；觀察到失敗即還原。
  - **可偽性見證 (b)（`--new` 重置）**：暫時讓 the header 用 lifetime `Σ`（跨 archive），確認 the fresh-session Example（`Turn 1`）失敗；觀察到失敗即還原。
  - 確認：`stdout` byte-exact；the chrome 仍為 **one** header + **one** `Ready` per prompt；the payload/post-turn/spinner/tool-log 不變；the tool semantics/set 不變；class-phrase 詞彙維持 **11**；exit-code 表 `0/2/3/4/5/6/7` 不變；the round-018 token store 不變；`go.mod`/`go.sum` 不變（無新相依）；Gherkin/DSL topology audit **PASSED**。
  - 若任何既有測試 pin 了 history 行的**逐位元組**形狀（round-008/014 的 byte-exact 主張），確認 `calls` 只在 completed turns 上出現且既有 fixture 已對齊（the `calls` field is a new top-level entry field — a recorded, expected round-027 line change）。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> CLI Application（Turn chrome (operator) — the call-based `<N>`） | T003、T009、T010、T011 | PASS |
| `specs/truth/techstack.md` -> CLI Application（Session history store — the `calls` field） | T001、T002、T007、T009 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T001–T003、T007、T009 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T011 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` MODIFY（`history_entry.calls`） | T001、T002、T007、T009 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/presenting-the-turn.feature` reshape） | T005、T009、T010 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md` — the header row + the 2 tool-using Givens） | T005、T008、T009 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`history/dsl.md` — the tool-using Given） | T006、T008 | PASS |
| `research.md` -> Decision 1（unit = inference round；retries excluded） | T003、T007、T010 | PASS |
| `research.md` -> Decision 2（persist the per-turn count on the entry） | T001、T002、T007、T009 | PASS |
| `research.md` -> Decision 3（header emitted pre-turn → prior+1） | T003、T009 | PASS |
| `research.md` -> Decision 4（reuse `AgentResult.Calls`） | T003、T007 | PASS |
| `research.md` -> Decision 5（session-scoped；`--new` reset；legacy=1） | T002、T003、T007、T011 | PASS |
| `research.md` -> Decision 6（stdlib / POSIX / surface unchanged） | T003、T010、T011 | PASS |
| `spec.md` -> US1–US2、`FR-001`–`FR-007`、`NFR-001`–`NFR-004` | T001–T007（對齊）、T009–T011（交付） | PASS |
| `spec.md` -> 邊界情況（failed turn persists nothing；non-prompt paths no header；resumed session continues；legacy entry=1） | failed turn → 現行行為（T011 回歸確認）；non-prompt → 現行行為；resumed → T007/T011；legacy → T007 | PASS |
| `plan.md` -> Source-code structure（`domain/history/history.go`、`infrastructure/history`、`cli/cli.go`；`ui`/`agent` unchanged；no new module） | T001–T003、T009、T010 | PASS |
| `plan.md` -> Scope notes（api NOOP；data MODIFY；`/axb-ui-plan` skipped；CLI end → `/axb-dsl-refine`） | T001、T008、T009、T011 | PASS |
| acceptance -> `counting-how-often-the-model-is-asked.feature`、`starting-the-count-over-on-a-fresh-session.feature` | T005、T006、T009（carried by the reshaped interface Rule） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。


---

## Folds — operator-reported `--new -i` defect (PR #58 review/re-review follow-on)

> The operator reported that `tellme --new -i` on a populated session opened at `Turn 2` instead of `Turn 1`. Root cause: the `-i` (TUI) submit path **dropped `--new`**, so the session was never archived and the header counted the prior history. Folded into round 027 (it violates the round's `--new`-reset contract on the `-i` surface).

- [X] T012 [BDD-RED] `When: the operator starts a fresh session at the interactive prompt and submits the prompt "{prompt}"` (new row in `specs/truth/features/cli/chat/dsl.md`)
  - Landing: `tests/e2e/steps/step_r027_chat_when_fresh_submit_at_prompt.go` + `launchTUIFresh` (`tests/e2e/steps/tui_keys.go` — runs `tellme --new -i`).

- [X] T013 [BDD-GREEN] `--new` honored on the `-i` surface + the interface Example
  - `internal/cli/cli.go`: `--new` now archives **before** the interactive read for **both** terminal readers (moved above the `tuiRequested` branch).
  - `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature`: Example `The operator starts a fresh session at the interactive prompt` (arranged tool-using history → `Turn 1`).

- [X] T014 [REGRESSION] witness + gates
  - Witness: restoring the pre-fix ordering makes the new Example fail — `the header carried Turn 3, want 1`.
  - `make verify` + `go test -count=1 ./...` + topology audit green.
  - Forward note (recorded): `--new` on the `-i` surface archives **before** the prompt (A8-style, matching the plain reader), so an aborted/empty `--new -i` still starts a fresh session.
