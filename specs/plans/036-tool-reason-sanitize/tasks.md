# Tasks: 036 — tool reason sanitize (fold + trim + cap the model-authored reason)

**Plan Package**: `specs/plans/036-tool-reason-sanitize`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0006-tool-reason-fold-and-cap.md`, `specs/truth/features/cli/chat/dsl.md`, `specs/truth/features/cli/chat/watching-the-tool-loop.feature`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**) and `specs/truth/data/**` (`/axb-data-plan` **NOOP**) are unchanged this round. There is no `ui/**` (a line-oriented CLI diagnostic).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是一個顯示契約缺陷修復（presentation-contract）**：`stderr` 工具日誌的 `[Tool Reason]` 行。沒有新增 / 刪除任何 Gherkin 句、Example 或 `DSLRow`；`/axb-dsl-refine` 只 **MODIFY** `chat/dsl.md` 的 reason row 文案（語意不變，明示先前隱含的單行保證）並加一段 round-036 note。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D7；`go.mod`／`go.sum` 不動）。
  - **省略 Phase 2 `Foundational`** —— 落點（`internal/ui/toolcall.go`、`internal/agent/agentloop.go`、`internal/cli/call_renderer.go`、新檔 `internal/ui/toolcall_reason_test.go`）皆由下方任務直接交付；無 stepdef 落點骨架、無並行調度需求。
  - **Phase 3 只有 `[UNIT]`** —— 本輪沒有待處理 DSL 句；測試層工作是**新增 hostile-fixture unit pins 並讓它們 RED**（operator Q1：unit-pin-only witness；**不新增 E2E Example**）。
  - **Phase 4 只有一個 MODIFY phase** —— 產品端讓 formatter + 兩個 caller 滿足 reason row 的新文案／契約（`[BDD-GREEN] -> [BDD-REFACTOR]`）。
  - 交付物是 **`FormatToolReason` 的 fold + trim + cap**（`internal/ui/toolcall.go`，新常數 `reasonValueCap = 200`）與 **blank-reason 抑制的三個站點**（`internal/agent/agentloop.go` `logAction`、`internal/agent/agentloop.go` `reasonsOf`（tail 的來源清單；生產路徑過濾器）、`internal/cli/call_renderer.go` `OnCallEnd` 內 `emit` closure 的**防禦性**守衛——生產不可達，見 `spec.md` FR-006 與 [#69](https://github.com/gosharplite/tellme/issues/69)）；truth 變更為 `specs/truth/techstack.md`（Agent tool loop row，by `/axb-technical-research`）與 `specs/truth/features/cli/chat/dsl.md`（reason row + note，by `/axb-dsl-refine`）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與 `specs/truth/techstack.md` 對應 section。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** flags、exit codes、class-phrase vocabulary、其他 line formats、frame cadence、tool schemas/results、provider transport、`history.jsonl`、`stdout`；**不動** sibling caps（189/200）與 `FormatToolResult` 的 fold 行為。

## Round-036 locked decisions (implementation constraints — MUST)

> 來自 operator Q1/Q2、`spec.md` FR-001..FR-008 / SC-001..SC-005、`research.md` D1–D7。

- **[SINGLE-SITE]** fold + trim 落在**純 formatter** `FormatToolReason` 內（重用既有 `oneLine`），兩個 caller 不各自重複（`research.md` D1）；單點修復即修好兩條表面（loop action line + per-call tail）。
- **[CAP]** `const reasonValueCap = 200`；render = `capRunes(oneLine(reason), reasonValueCap)`（單一 U+2026 計入 cap、rune 邊界切、對 folded 值評估），與 `FormatToolResult`（200）同機制（`research.md` D2）。
- **[BLANK]** 空或全空白（fold + trim 後）的 reason **不輸出** `[Tool Reason]` 行；抑制在**兩個 caller** 以 trim 後值判斷（`strings.TrimSpace(reason) != ""` 取代現行 raw `reason != ""`），`FormatToolReason` 保持**純**（不得回傳空字串 sentinel —— tail caller 用 `Fprintln`，會印出裸換行）（`research.md` D3）。
- **[WITNESS]** hostile-fixture **unit pins**（`\n`、`\r`、全空白、201-rune over-cap）+ blank-reason 抑制 pins（action line + tail）；**不新增 E2E Example、不新增 `DSLRow`**（Q1 Option 1；E2E 對此缺陷類結構性失明，count 公式各漏一項）。
- **[RED-FIRST]** 新 unit pins 今日必須對現行缺陷**非真空地失敗**（assertion-only，非 undefined、非 parse error），再於 `[BDD-GREEN]` 修復後轉綠。
- **[SCOPE]** 範圍守門（`research.md` D7）：只改 `stderr` 工具日誌的 reason 呈現；sibling formatters、`FormatToolResult` 行為、class-phrase vocabulary、`stdout` 皆**不動**；`agentloop_reason_test.go` 的 loop-tier pin 若被 trim/fold 觸及須同批更新（clean reason 應 byte-identical）。
- **[ADR]** round-034 的未記錄落差 + 本輪決策記於新 **ADR 0006**（`docs/decisions/0006-tool-reason-fold-and-cap.md`）；**不編輯** ADR 0005（immutable）（`research.md` D5）。
- **[TIDY]** 將 `oneLine` 由 `internal/ui/toollog.go` 搬入 `internal/ui/toolcall.go`（其唯一 consumer 家族）並刪除 orphan `toollog.go`（`research.md` D6）。
- **[NO-DEP]** 只用 Go 標準庫；無新相依；`go.mod`／`go.sum` 不動；POSIX-only。
- **[VERIFY]** hermetic（無 pty）；`make verify` + godog E2E + 拓樸稽核皆須綠。
- **[TRUTH]** 唯一 truth 變更 = `specs/truth/techstack.md`（Agent tool loop row）+ `specs/truth/features/cli/chat/dsl.md`（reason row + round-036 note）。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D7；無新相依，`go.mod`／`go.sum` 不動）。不得把落點骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無 stepdef 落點骨架（無新增 DSL 句）；修復落點（`internal/ui/toolcall.go`、`internal/agent/agentloop.go`、`internal/cli/call_renderer.go`）與 unit pin 落點（新檔 `internal/ui/toolcall_reason_test.go`）皆由下方任務直接交付。不得偷做 Phase 3 測試層或 Feature Green。

## Phase 3: Test Alignment (unit)

**Goal**: 先把本輪的驗證落到測試層（**不寫產品碼**）：新增 `FormatToolReason` 的 hostile-fixture unit pins（fold / trim / cap）與 blank-reason 抑制 pins（action line + tail）並讓它們對現行缺陷 **RED**。

**DSL 參照**:
- 本輪**沒有新增或修改任何 DSL 句**：`chat/dsl.md` 的 reason row 是**文案 MODIFY**（明示已隱含的單行保證 + 落上 cap 常數），無新增句型、無新 stepdef；因此 Phase 3 不含 `[BDD-ALIGN]`／`[BDD-REMOVE]`／`[BDD-RED]` 任務。
- `truth-delta.md` 記錄 `/axb-dsl-refine` 為 **MODIFY**（reason row 文案 + note，非 DSL 句新增）。

**Markers**:
- `[UNIT]`：本輪測試層工作是 **unit-tier**（pure formatter pins + emission-path pins）；非 Gherkin stepdef。只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> reason row（`the run reported the reason …`）+ round-036 note
- `specs/truth/techstack.md` -> Agent tool loop row（round-036 reason clause）
- `docs/decisions/0006-tool-reason-fold-and-cap.md` -> D1–D3
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY `chat/dsl.md`；`/axb-technical-research` MODIFY `techstack.md`
- `research.md` -> `Decision 1`, `Decision 2`, `Decision 3`, `Decision 4`

**Boundary**:
- 只動測試層（新檔 `internal/ui/toolcall_reason_test.go`，Zero Shared Edits；不修改既有 `internal/ui/toolcall_test.go`）。不寫產品碼。
- review 啟動 subagent；本輪 unit pins 必須對現行缺陷 **非真空失敗**（assertion-only）；且**不得**新增或修改任何 feature/Example（0 undefined steps 的 E2E 不受影響）。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T001 與 T002 觸及不同斷言面向但可同檔合作；為求 Zero Shared Edits 與簡單性，序列執行：T001 → T002 → T003（subagent review）；T003 等 T001/T002 回來再啟動。

- [X] T001 [UNIT] `FormatToolReason` hostile-fixture pins（fold + trim + cap；RED）
  - Read:
    - `research.md` -> `Decision 1`, `Decision 2`, `Decision 4`
    - `internal/ui/toolcall.go` -> `FormatToolReason`、`argValueCap`/`resultValueCap`、`capRunes`
    - `internal/ui/toollog.go` -> `oneLine`
    - `internal/ui/toolcall_test.go` -> `TestFormatToolReason`（現行 clean-fixture pin）、`TestFormatToolResult_FoldsAndCaps`（sibling 典範）
    - `specs/truth/features/cli/chat/dsl.md` -> reason row（single-line 保證 + cap 常數）
  - 做：新增 `internal/ui/toolcall_reason_test.go`（新檔），以 `r034Clock` 斷言：
    - **fold `\n`**：`FormatToolReason(t, "a\nb")` == `[HH:MM:SS] [Tool Reason] a b`（單行、無 `\n`）；
    - **fold `\r`**：`FormatToolReason(t, "a\rb")` == `[HH:MM:SS] [Tool Reason] a b`（無 `\r`；**prefix 完整保留**，不被截斷）；
    - **trim**：前後空白被去除（例：`"  spaced  "` → `[Tool Reason] spaced`）；
    - **cap**：201-rune reason 的值恰為 **200 runes** 且以單一 U+2026 結尾（`reasonValueCap`）；並以多-byte rune 邊界 case 確認不切裂 rune；
    - **clean 不變**：`"checking the launch code"` 仍 render byte-identical（與既有 `TestFormatToolReason` 相同字串，避免與既有 pin 衝突）。
  - 邊界：這些斷言**今日必須失敗**（現行 `FormatToolReason` 不 fold、不 cap、不 trim）——**不得**放寬 assertion 讓它變綠（RED-first）；不寫產品碼；**不**修改既有 `toolcall_test.go`。
  - 不做：不改 `toolcall.go`（留 Phase 4）；不碰 sibling formatters；不新增 E2E。

- [X] T002 [UNIT] blank-reason 抑制 pins（loop real path + tail defensive；RED）
  - Read:
    - `research.md` -> `Decision 3`
    - `internal/agent/agentloop.go` -> `logAction` 與 `reasonsOf`（現行 `if reason := toolReason(tc.Arguments); reason != ""`；`reasonsOf` 是 tail 的來源清單）
    - `internal/cli/call_renderer.go` -> `OnCallEnd` 內的 `emit` closure（tail 的 grouped reason 輸出）
    - `internal/agent/agentloop_reason_test.go` -> 既有 loop-tier pin（`[12:34:56] [Tool Reason] because`）
    - `specs/truth/features/cli/chat/dsl.md` -> reason row（blank ⇒ no line）
  - 做：新增抑制 pins（放於新檔，Zero Shared Edits）：
    - **loop real path**：以 whitespace-only reason（`"   "`）與 newline-only（`"\n"`）驅動**真實** `AgentLoop.Run`（同時觸及 `logAction` 與 `reasonsOf`），斷言 `stderr` **不含**任何 `[Tool Reason]` 行（也不多出空白行）；
    - **tail defensive**：直接以 whitespace-only reason 呼叫 `callRenderer.OnCallEnd`，斷言無 `[Tool Reason]` 行（此站為防禦性、生產不可達——見 `spec.md` FR-006；其真實路徑由 `reasonsOf` 覆蓋）。
    - **clean 對照**：非空 reason 仍輸出一行（避免把抑制寫成一律不輸出）。
  - 邊界：**今日必須失敗**（現行 guard 只看 raw `reason != ""`，`"   "`／`"\n"` 會通過並輸出 dangling 行）——不得放寬 assertion（RED-first）；不寫產品碼。
  - 不做：不改 `agentloop.go`／`call_renderer.go`（留 Phase 4）；不動 pure formatter。

- [X] T003 subagent review (phase quality gate)
  - Read: `internal/ui/toolcall_reason_test.go`、`internal/ui/toolcall.go`、`internal/agent/agentloop.go`、`internal/cli/call_renderer.go`、`specs/truth/features/cli/chat/dsl.md`、`research.md` -> `Decision 3`, `Decision 4`
  - 檢驗：
    - unit pins **非真空失敗**且點名 fold / trim / cap / blank-suppression；只動測試層。
    - **本輪未新增/修改任何 feature 或 DSL 句**：跑 `go test ./tests/e2e/...` 確認既有 suite 仍可跑到，且 **0 undefined steps**；`chat/dsl.md` 的 reason row 文案與 feature 一致。
    - clean reason 的既有 pin（`TestFormatToolReason`、`agentloop_reason_test.go`）仍成立（未被新檔覆寫）。
    - 有 issues 修正再 review，直到零問題；通過前不解鎖 Phase 4。

> Phase-3 review 由 orchestrator 執行（本 session 無 parallel-subagent substrate）：新 unit pins 以現行 `FormatToolReason`／callers 跑得起來但**失敗**（assertion-only）：`FoldsNewline`／`FoldsCarriageReturn`／`TrimsSurroundingWhitespace`／`CapsAndFoldsCombined`／`ValuesAreRuneCapped`／`CapCutOnRuneBoundary` 全部 RED（實測輸出顯示未 fold 的 `a\nb`、`a\rb`、未 trim 的 `  spaced  `、未 cap 的 201/300 runes）；`TestLogOmitsReasonLineForWhitespaceOnlyReason` RED（`[Tool Reason]    ` dangling row）；`TestCallTailOmitsBlankReasonLine` RED（`[Tool Reason]    ` + 裸換行）。E2E suite 不含新句（**0 undefined steps**，`go test ./tests/e2e/...` OK）。Gate PASSED —— Phase 4 解鎖。

## Phase 4: MODIFY — the reason row contract (formatter + callers)

**Goal**: 以最小產品變更讓 T001/T002 轉綠：`FormatToolReason` fold + trim + cap；兩個 caller 對 blank reason 抑制；並與 `chat/dsl.md` 的 reason row 新文案一致。不改 sibling formatters、不改任何 line format/cadence、`stdout` byte-exact。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> reason row（single-line 保證 + `reasonValueCap`）
- `specs/truth/techstack.md` -> Agent tool loop row（round-036 reason clause）
- `docs/decisions/0006-tool-reason-fold-and-cap.md` -> D1, D2, D3
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY `chat/dsl.md`；`/axb-technical-research` MODIFY `techstack.md`
- `research.md` -> `Decision 1`, `Decision 2`, `Decision 3`, `Decision 6`, `Decision 7`

**Boundary**:
- 改 `internal/ui/toolcall.go`：新增 `const reasonValueCap = 200`；`FormatToolReason` 改為 `fmt.Sprintf("[%s] [Tool Reason] %s", formatClock(t), capRunes(oneLine(strings.TrimSpace(reason)), reasonValueCap))`；更新其 doc comment（fold + trim + cap，200）。
- 改 `internal/agent/agentloop.go`（`logAction` **與** `reasonsOf`）與 `internal/cli/call_renderer.go`（`OnCallEnd` 內 `emit` closure）：把 raw `reason != ""` guard 改為 sanitized-value guard（`strings.TrimSpace(reason) != ""`）；**不**在 caller 內 fold/cap（單點原則）。第三站（`emit` closure）為防禦性冗餘，生產不可達，其單一歸屬收斂記於 [#69](https://github.com/gosharplite/tellme/issues/69)（本輪不重構）。
- **不動** `FormatToolAction`／`FormatToolResult`／`argValueCap`／`resultValueCap`／`capRunes` 語意、class phrase、exit codes、`stdout`、frame cadence、tool schema。
- 不得為了轉綠而放寬 T001/T002 的 assertion。

**Test Scope**:
- `internal/ui/toolcall_reason_test.go`（T001/T002）
- `internal/ui/toolcall_test.go`（既有 `TestFormatToolReason` 續綠）
- `internal/agent/agentloop_reason_test.go`（loop-tier pin 續綠；若 trim 觸及 fixture 則同批更新）

- [X] T004 [BDD-GREEN] 讓 Test Scope 全綠（fold + trim + cap + blank suppression）
  - Read: `research.md` -> `Decision 1`, `Decision 2`, `Decision 3`；`internal/ui/toolcall.go`；`internal/agent/agentloop.go`；`internal/cli/call_renderer.go`
  - 做：
    - `FormatToolReason` 套用 `capRunes(oneLine(strings.TrimSpace(reason)), reasonValueCap)`（`reasonValueCap = 200`）。
    - 兩個 caller 的 guard 改為 sanitized-value（`strings.TrimSpace(reason) != ""`）。
  - 驗證：T001/T002 轉綠；`TestFormatToolReason` 與 `agentloop_reason_test.go` 續綠（clean reason byte-identical）；`go test ./...` 全綠；`stdout` byte-exact。
  - 不做：不改 sibling formatters；不新增相依；不改 format/cadence；不在 caller 重複 fold/cap。

- [X] T005 [BDD-REFACTOR] 在綠燈下 tidy `oneLine` 的歸屬與註解
  - Read: `research.md` -> `Decision 6`, `Decision 7`；`internal/ui/toollog.go`；`internal/ui/toolcall.go`
  - 做：將 `oneLine` 由 `internal/ui/toollog.go` 搬入 `internal/ui/toolcall.go`（其唯一 consumer 家族），刪除 `toollog.go`；確保 `toolcall.go` 頂部註解正確描述三條 formatter 皆 sanitize+cap（arg 189 / reason 200 / result 200）。保持輸出語意不變、gate 續綠。
  - 不做：不擴大重構範圍、不改任何輸出、不改 caller 契約。

## Phase 5: Verification & Regression

**Goal**: 全量回歸 + 可偽性見證 + `make verify` + 拓樸稽核，並確認 flags／exit codes／class phrase／其他 line formats／cadence／`stdout`／sibling caps 皆未變。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Agent tool loop row
- `docs/decisions/0006-tool-reason-fold-and-cap.md`
- `research.md` -> `Decision 4`, `Decision 5`, `Decision 7` + Recorded residual risks

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses（可偽性）：(a) **移除 fold**（`oneLine`）→ `\n`/`\r` pins 必須失敗；(b) **移除 cap** → over-cap pin 必須失敗；(c) **還原 raw guard**（`reason != ""`）→ blank-reason 抑制 pins 必須失敗。觀察到即還原，重跑確認綠燈。

- [X] T006 [REGRESSION] 跑全測試 + 見證 + `make verify` + 拓樸稽核
  - Read: `research.md` -> `Decision 4`, `Decision 5`, `Decision 7`；`specs/truth/techstack.md` -> Agent tool loop row；`docs/decisions/0006-tool-reason-fold-and-cap.md`
  - 做：
    - `go test -count=1 ./...` 全綠（unit + godog E2E）。
    - 見證 (a)/(b)/(c) 如上，觀察到即還原，重跑確認綠。
    - `make verify` **OK**；`gofmt -l .` clean；Gherkin/DSL 拓樸稽核 **PASSED**（44 features · 16 root + 310 module rows · **1576** steps —— 與 `/axb-dsl-refine` 交付時相同；本輪無改 feature/DSL 句）。
    - `go.mod`／`go.sum` 不變（stdlib-only）。
    - 確認 sibling caps（189/200）與 `FormatToolResult` 行為未變；clean reason 輸出 byte-identical。
  - 不做：不放寬任何 assertion；不為轉綠而移除見證。

- [X] T007 subagent review (round quality gate)
  - Read: `internal/ui/toolcall.go`、`internal/ui/toolcall_reason_test.go`、`internal/agent/agentloop.go`、`internal/cli/call_renderer.go`、`specs/truth/features/cli/chat/dsl.md`、`specs/truth/techstack.md`、`docs/decisions/0006-tool-reason-fold-and-cap.md`、`specs/plans/036-tool-reason-sanitize/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：修復僅動 `toolcall.go`（+ caller guards + 新 test 檔）；sibling formatters／caps 未改；`FormatToolReason` 仍為純 formatter（無 sentinel）；blank reason 兩表面皆無行；未新增 E2E/DSL；ADR 0006 存在且與 truth 一致；無新相依；flags／exit codes／其他 formats／`stdout` 未變。

> Round review 由 orchestrator 執行：修復只動 `internal/ui/toolcall.go`（新增 `reasonValueCap = 200`、`FormatToolReason` fold+trim+cap、`oneLine` 由已刪除的 `toollog.go` 遷入）、`internal/agent/agentloop.go`（`logAction` + `reasonsOf` sanitized guard）、`internal/cli/call_renderer.go`（tail emit 的 sanitized guard），與測試檔；`FormatToolAction`／`FormatToolResult`／`argValueCap`／`resultValueCap`／`capRunes` **未改**；`FormatToolReason` 仍為純 formatter（無 sentinel）；`go.mod`/`go.sum` 不變（stdlib-only）；`gofmt -l .` clean；`go test -count=1 ./...` 全綠（E2E OK）；`make verify` **OK**（no-test-sleep · offline witness · cross-compile 4/4 · MCP confinement · golangci-lint **0 issues** · govulncheck **0 reachable**）；拓樸稽核 **PASSED**（44 features · 6 modules · 16 root + 310 module rows · 1576 steps —— 與交付時相同）。可偽性見證 (a)/(b)/(c) 重現後還原（見 T006）。

### PR #75 架構審查 fold（round 036 — APPROVE WITH NON-BLOCKING FOLDS）

> 審查 head `048bb28`；無 `[ARCHITECTURAL BLOCKER]`；單點純 formatter 修復、caller-side 抑制、successor ADR、doc-level truth 編輯皆獲認可。以下 folds 只讓**本輪自己的紀錄與程式碼一致**（**不**在輪內重構程式碼）。

- **TD-2**：抑制站點數由「兩個」更正為**三個**（`logAction`、`reasonsOf`、`OnCallEnd` 內 `emit` closure），並以真實符號取代不存在的 `renderCallTail` — 已 fold 於 `spec.md` FR-006 + Key entities、`research.md` D3、`tasks.md` Phase-4 Boundary、`chat/dsl.md` round-036 note。
- **TD-3**：permanent E2E narrowing 的永久歸宿改指向**活躍 issue**（`research.md` D4 → [#69](https://github.com/gosharplite/tellme/issues/69)，並於 #74 關閉前留言），因 plan package 於 merge 時凍結（round-035 G3 教訓）。
- **TD-1**：`OnCallEnd` 的第三站為**生產不可達的防禦性冗餘** — 已在 code 加註指明上游生產過濾器為 `reasonsOf`，並將「blank-reason 述詞單一歸屬」記於 [#69](https://github.com/gosharplite/tellme/issues/69)（**本輪不重構**）。
- **RF-1**：刪除與既有 `toolcall_test.go::TestFormatToolReason` 逐字重複的 `TestFormatToolReason_CleanReasonUnchanged`，改以註解引用既有 pin（Zero Shared Edits）。
- **RF-2**：`spec.md` 兩條已被 supersede 的敘述加註（Q1 的 pin 檔路徑 → `toolcall_reason_test.go`；ADR 0005 車輛 → research D5 的 ADR 0006）。
- **N-1 / N-3 / N-4**：tail 空輸出改用 `buf.Len() != 0` 直接斷言；`r036Clock` 補上存在理由；`reasonsOf` 補上「以 trimmed 值過濾、append raw 值」的註解。
- 驗證：`gofmt -l .` clean · `go build ./...` OK · `go test -count=1 ./...` 全綠（E2E OK）· 可偽性見證 (a)/(b) 重現後還原。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Agent tool loop row (round-036 reason clause) | T004（交付修復）、T006（`make verify`／稽核）、T007 | PASS |
| `specs/truth/features/cli/chat/dsl.md` -> reason row（single-line 保證 + `reasonValueCap`） | T001（pin 契約）、T004（交付一致）、T006、T007 | PASS |
| `specs/truth/features/cli/chat/dsl.md` -> round-036 note | T001/T003（Read）；T007（一致性檢查） | PASS |
| `docs/decisions/0006-tool-reason-fold-and-cap.md` (ADR) | T001（Read D1–D3）、T005（Read D6/D7）、T007 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY `techstack.md` | T004、T006、T007 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY reason row + note | T001（Read）、T004（交付一致）、T006、T007 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP (`specs/truth/contracts/**`) | 豁免（NOOP 不建任務；T006/T007 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP (`specs/truth/data/**`) | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `research.md` -> Decision 1（single-site fold in the pure formatter） | T001（pin 契約）、T004（交付）、T007 | PASS |
| `research.md` -> Decision 2（`reasonValueCap = 200`；capRunes/oneLine 機制） | T001（cap pin）、T004（交付）、T006、T007 | PASS |
| `research.md` -> Decision 3（blank-reason 抑制於兩 caller；formatter 保持純） | T002（pin）、T004（交付）、T007 | PASS |
| `research.md` -> Decision 4（witness = hostile-fixture unit；無 E2E） | T001/T002（unit pins）、T003（0 undefined 驗證） | PASS |
| `research.md` -> Decision 5（ADR 0006；不編輯 0005） | T005/T007（Read ADR 0006）、T006（一致性） | PASS |
| `research.md` -> Decision 6（tidy：`oneLine` 搬入 `toolcall.go`） | T005（交付 tidy）、T007 | PASS |
| `research.md` -> Decision 7（scope guard；loop-tier pin 同批） | T002（loop pin Read）、T004（交付）、T006（未變 surfaces） | PASS |
| `spec.md` -> US1（FR-001–FR-004） | T001、T004 | PASS |
| `spec.md` -> US2（FR-005、FR-006） | T002、T004 | PASS |
| `spec.md` -> 全域（FR-007、FR-008） | T004（不改 surfaces）、T006（sibling/caps/stdout 未變） 、T007 | PASS |
| `spec.md` -> 邊界情況（both `\n`+`\r`；mid-rune cap；absent reason；blank on tail；tool-less turn） | T001（both + mid-rune）、T002（blank on tail；absent/tool-less 屬既有行為） | PASS |
| `spec.md` -> SC-001..SC-005 | T001（SC-001）、T002（SC-002）、T004/T006（SC-003）、T006（SC-004）、T005/T006（SC-005 見證） | PASS |
| `plan.md` -> Interface inventory / Wave（1 CLI end → `/axb-dsl-refine`；api/data NOOP；ui skipped） | T001/T002/T004（CLI 介面交付）、T006（不觸及 API/資料） | PASS |
| operator 拍板（Q1 unit-pin-only；Q2 blank ⇒ no line） | T001/T002（unit witness）、T004（交付） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
