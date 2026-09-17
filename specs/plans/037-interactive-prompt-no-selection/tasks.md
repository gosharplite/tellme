# Tasks: 037 — interactive prompt no-selection (align the `-i` suggestion cursor to the reference)

**Plan Package**: `specs/plans/037-interactive-prompt-no-selection`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature`, `specs/truth/features/cli/chat/dsl.md`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**) and `specs/truth/data/**` (`/axb-data-plan` **NOOP**) are unchanged. Plan-side `ui/**` exists (terminal mode) but is review-only.

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是一個 TUI 選擇狀態的 presentation-parity 修正**：`-i` 提示的建議清單**在起始與每次 refresh 後皆無選中項**（no-choice sentinel `cursor == -1`；第一個 `Tab` 仍選中**第一項**）。`/axb-dsl-refine` **MODIFY** `presenting-the-interactive-prompt.feature` 的 at-rest `Then`（原地字句改寫）與 `chat/dsl.md` 的**同一列**（句型改寫，非新增）並加一段 round-037 note。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D1–D3；`go.mod`／`go.sum` 不動，stdlib + 既有 bubbletea/lipgloss）。
  - **省略 Phase 2 `Foundational`** —— 落點（`internal/ui/tui/prompt/suggester.go`、`model_chrome_test.go`、`refresh_test.go`、`tests/e2e/steps/step_t008_chat_then_marks_current_choice.go`）皆由下方任務直接交付；無 stepdef 落點骨架、無並行調度需求。
  - **Phase 3 含 `[BDD-ALIGN]`（E2E stepdef 字句對齊）+ `[UNIT]`（unit pins）** —— 測試層工作是把 at-rest 斷言翻成「**零** 個 cursor 列」並讓它對現行行為 **RED**（現行起始會標記 1 列）。
  - **Phase 4 一個 MODIFY phase** —— 產品端讓 suggester 起始為 no-choice 並於 refresh 重設（`[BDD-GREEN] -> [BDD-REFACTOR]`）。
  - 交付物是 **`suggester` 的 no-choice 起始 + refresh 重設**（`internal/ui/tui/prompt/suggester.go`）；truth 變更為 `specs/truth/techstack.md`（interactive-prompt / suggestion-engine rows，by `/axb-technical-research`）與 `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` + `chat/dsl.md`（at-rest Then / 列 + note，by `/axb-dsl-refine`）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與 `specs/truth/techstack.md` 對應 section。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** suggestion sources/ordering、`Suggestions:` header、selected/unselected 樣式、`> ` glyph、editor、placeholder、debounce、no-metrics-header、round-023 teardown、flags、exit codes、class-phrase vocabulary、`history.jsonl`、`stdout`。

## Round-037 locked decisions (implementation constraints — MUST)

> 來自 operator Q1–Q4、`spec.md` FR-001..FR-006 / SC-001..SC-005、`research.md` D1–D7。

- **[NO-CHOICE]** 選擇狀態以 `cursor == -1` 表示「無選中項」，並於 `suggester.set(items)` **每次**重設為 `-1`（對齊參考 `Update(msg, -1)`）；`selected()` 對負游標已回傳 `""`（`research.md` D1）。
- **[ARITHMETIC]** `cycle(delta)` 公式不變（`(cursor+delta+len)%len`），不對 `-1` 特判；從 `-1` 的 `cycle(1)` → `0`（第一項），`cycle(-1)` → `len-2`（參考自身算術）（`research.md` D2）。
- **[ACCEPT]** `accept(delta)` 仍為 `cycle` + 插入 `selected()`；第一個 `Tab`（`accept(1)`）從 no-choice 選中並插入**第一項**——既有 accept journey 不改（`research.md` D2）。
- **[WITNESS]** **unit pins**（at-rest `View()` 零 cursor 列；`selected()` 預設 `""`；refresh 重設；`Tab` 從 no-choice 選中第一項）**＋ E2E** `Then` 翻成斷言 **0** 列（cursor 列在截取輸出可直接觀測）（Q4；`research.md` D4）。
- **[RED-FIRST]** 新／改的 unit pins 與 E2E stepdef 今日必須對現行行為**非真空地失敗**（現行起始 1 列 → 期望 0 列），再於 `[BDD-GREEN]` 修復後轉綠。
- **[SCOPE]** 範圍守門（`research.md` D6）：只改選択の at-rest/refresh 狀態；`stdout`、flags、exit codes、class-phrase、樣式、glyph、排序、debounce 皆**不動**。**空白提交（empty `Ctrl+S`）的落差**屬 issue [#76](https://github.com/gosharplite/tellme/issues/76)，**不**在本輪處理。
- **[SUPERSEDE]** 本輪 at-rest 規則**取代** round 016 的 at-rest 規則；round-016 plan package（含其 `ui/screens/entry.txt`）為**凍結歷史**，**不編輯**（`research.md` D5）。
- **[NO-DEP]** 只用 Go 標準庫 + 既有相依；`go.mod`／`go.sum` 不動；POSIX-only。
- **[VERIFY]** hermetic（無 pty）；`make verify` + godog E2E + 拓樸稽核皆須綠。
- **[TRUTH]** 唯一 truth 變更 = `specs/truth/techstack.md`（2 rows）+ `specs/truth/features/cli/chat/{presenting-the-interactive-prompt.feature,dsl.md}`。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D1–D3；無新相依）。

## Phase 2: Foundational

**Omitted** —— 無 stepdef 落點骨架（句型為**原地改寫**，非新增）；修復落點（`internal/ui/tui/prompt/suggester.go`）與 pin 落點皆由下方任務直接交付。

## Phase 3: Test Alignment (E2E align + unit)

**Goal**: 先把本輪的驗證落到測試層（**不寫產品碼**）：把 at-rest E2E `Then` 與 unit pins 翻成「無選中項」，並讓它們對現行起始「標記第一項」**RED**。

**Checklist**:
- [x] T001 [BDD-ALIGN] `tests/e2e/steps/step_t008_chat_then_marks_current_choice.go`：把註冊的句型由 `… marks one suggestion as the current choice` 改為 `… marks no suggestion as the current choice`，斷言由「`tuiCursorRows(out) < 1` → 失敗」改為「`tuiCursorRows(out) != 0` → 失敗」（at-rest 必須零 cursor 列）。函式註解更新（round 037）。**RED**：現行起始 1 列 → `!= 0` 失敗。
- [x] T002 [P][UNIT] `internal/ui/tui/prompt/model_chrome_test.go`：
  - 將 `TestModelCursorDefaultsToFirstItem` 改為 `TestModelCursorStartsNoChoice`：`selected()` 預設為 `""`。
  - 將 `TestModelRendersReferenceChrome` 的 `strings.Count(view, "> ") != 1` 改為期望 **0**。
  - 新增 `TestTabFromNoChoiceSelectsFirst`：從 no-choice 按一次 `Tab` → `selected() == 第一項`（並可插入）。
- [x] T003 [P][UNIT] `internal/ui/tui/prompt/refresh_test.go`：新增 `TestRefreshResetsSelection`（先 `Tab` 選中，再送一筆 current-value `debounceMsg` 觸發 refresh → `selected()` 回到 `""`）；`TestRefreshUsesDebounce` 改以 `m.sug.items` 斷言重算（不再用 `selected()` 當身分探針）。
- [x] T004 Review gate：`go test ./internal/ui/tui/prompt/...` 對現行碼 **RED**（非 undefined、非 parse error）；`go test ./tests/e2e/...` 對 at-rest Example **RED**。記錄 RED 證據。

## Phase 4: Feature (Green → Refactor → Regression)

**Goal**: 讓 `suggester` 滿足 no-choice 起始 + refresh 重設，直到 Phase 3 的 pins 與 E2E 全綠。

**Checklist**:
- [x] T005 [BDD-GREEN] `internal/ui/tui/prompt/suggester.go`：新增具名常數 `noChoice = -1`；`newSuggester()` 回傳 `suggester{cursor: noChoice}`，且 **`set(items)` 一律將 `cursor` 設為 `noChoice`**（每次 refresh 重設為無選中項），移除「空清單 → 0 / 超界 → 0」的特判。`selected()`／`cycle()` 維持不變（負游標已安全回傳 `""`，且 `cycle` 對 `noChoice` 的模算術即為參考值）。
- [x] T006 [BDD-GREEN] 驗證 Phase 3 的 unit pins + E2E at-rest Example 轉綠；既有 accept E2E（`prompting-with-suggestions`）與其他 `-i` Examples 仍綠。
- [x] T007 [BDD-REFACTOR] 收斂：更新 `suggester` 的 doc 註解（cursor 預設與 `set` 重設語意由「first item」改為「no choice」），確認無殘留死碼。
- [x] T008 [REGRESSION] `gofmt -l .` 乾淨、`go vet ./...` 乾淨、`make verify` OK、`go test -count=1 ./...` 全綠、拓樸稽核 PASSED。
- [x] T009 [WITNESS] 可偽性見證：把 `set` 的重設改回 `0` → at-rest unit pin 與 E2E `Then` 失敗；移除重設 → refresh pin 失敗；各自重現後還原。

## Phase 5: Review & Delivery

- [ ] T010 開 PR（`037-interactive-prompt-no-selection` → `dev`）；送 architectural review（plan+truth 與 implementation 一併），依 fold 收斂。
- [ ] T011 人類 merge；propagate `dev → main`；`SESSION-CLOSEOUT.md`（STATUS refresh、daily summary、issue tracker：新 [#76](https://github.com/gosharplite/tellme/issues/76) 已開、無 closed）。

---

## Delivery outcomes

- **Phase 3 RED (recorded)**: `go test ./internal/ui/tui/prompt/...` → 4 pins fail non-vacuously (assertion-only): `TestModelRendersReferenceChrome` (`cursor count = 1, want exactly 0`), `TestModelCursorStartsNoChoice` (`cursor default = "a"`), `TestTabFromNoChoiceSelectsFirst` (`expected no pre-selection`), `TestRefreshResetsSelection` (`expected no pre-selection`). E2E: `go test ./tests/e2e/` → **215 scenarios (214 passed, 1 failed)**; the single failure is `The prompt presents the reference surface` — `the interactive prompt marked 1 suggestion(s) as the current choice at rest, want 0` (the flipped `Then`). All other `-i`/accept Examples pass.
- **Phase 4 GREEN (recorded)**: `internal/ui/tui/prompt/suggester.go` — `const noChoice = -1`; `newSuggester()` → `suggester{cursor: noChoice}`; `set(items)` resets `cursor = noChoice` on every refresh; `selected()`/`cycle()` unchanged. `go test ./tests/e2e/` → **ok** (215/215). `go test ./internal/ui/tui/prompt/...` → ok.
- **Phase 4 REGRESSION (recorded)**: `gofmt -l .` clean · `go vet ./...` clean · `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `verify-mcp-sdk-confinement` · golangci-lint 0 issues · govulncheck 0 reachable) · `go test -count=1 ./...` green (21 packages) · topology audit **PASSED** (44 features · 6 modules · 16 root + 310 module rows · 1576 steps — unchanged, as intended).
- **Phase 4 WITNESS (recorded)**: (A) re-defaulting `set`'s reset to `0` fails the at-rest unit pins; (B) removing the reset-on-refresh (clamping instead) fails `TestRefreshResetsSelection` (`a refresh did not reset the selection to no-choice: "alpha"`). Each reproduced then reverted.

---

## T010 — PR #77 review fold (2026-09-17)

Architectural review of PR [#77](https://github.com/gosharplite/tellme/pull/77) (`pullrequestreview-5230205073`): **APPROVE WITH REQUIRED FOLDS** — no product-code/truth-semantics change; the reference-parity premise was verified against `tell-me-go`, and the gates were independently reproduced. All folds landed:

| # | Finding | Resolution |
| --- | --- | --- |
| **F-1** | `step_t008` predicate scope wider than the rule (whole stream, both streams) | Added `atRestFrame(out)` (`tests/e2e/steps/tui_chrome.go`) — the first painted frame, delimited by the editor border `┌`; `thenMarksNoCurrentChoice` now asserts `tuiCursorRows(atRestFrame(...)) == 0`. Re-verified non-vacuous (witness A still fails the at-rest Example, `-count=1`). ***(Intermediate state — withdrawn by G-1 below; superseded by the exact-prefix-predicate fold.)*** |
| **F-2** | `research.md` D2 residual-risk claim arithmetically false (`len == 1` → "no selection") | Reworded to the true statement: `Shift+Tab` from no-choice → `len-2` for `len ≥ 2` **and `0` (the sole item) for `len == 1`** (Go `%` mod 1 = 0); now carried by the new table-driven pin. |
| **F-3** | Selection policy hardcoded in `suggester.set` (no named owner) | (a) stated the `cursor ∈ {noChoice} ∪ [0, len)` invariant in `suggester.go`'s doc; (b) recorded the shape (mirror the reference's `Update(items, index)`) as a **single-ownership item on [#69](https://github.com/gosharplite/tellme/issues/69)** — no in-round refactor (scope guard). |
| **F-4** | FR-003's wrap / `Shift+Tab`-from-no-choice clause had **no carrier** | Added `TestCycleArithmetic` (table-driven: `len ∈ {1,2,3}`, both deltas, both wrap directions) — closes FR-003 and witnesses the F-2 correction. |
| **F-5** | DSL row wording vs assertion scope not isomorphic | Narrowed the assertion to the at-rest frame (F-1) and rewrote the `chat/dsl.md` row to scope to the **at-rest frame** and name the **unit pin** as the first-`Tab` authority. ***(Intermediate state — the at-rest-frame scoping was withdrawn by G-1; the final row scopes to the rendered suggestion block.)*** |
| **F-6** | `STATUS.md` not updated (Step 7 would have checked out `dev`, dropping round-037 artifacts) | Refreshed `STATUS.md`: Active branch `037-interactive-prompt-no-selection`; round 037 current-round section + pipeline position (T001–T009 `[X]`; T010 = this review; T011 open); `#76` recorded; round-036 detail relocated **verbatim** to the [2026-09-17 archive](docs/archives/status/2026-09-17.md). |

**Also folded (non-numbered):** the evidence-carrier overstatement (US2 / research D3–D4) now names the **unit pin** `TestTabFromNoChoiceSelectsFirst` (two items) as the first-`Tab`-index authority, with the accept journey carrying *insertion* only; the cross-round note landed as a **body edit on [#76](https://github.com/gosharplite/tellme/issues/76)** (display-only sharpening + unambiguous fix direction); the plan-side nit (`ui/screens/entry.txt` seeds `version 037`).

**Re-verification after the folds:** `gofmt -l .` clean · `go vet ./...` clean · `make verify` **OK** · `go test -count=1 ./...` green (21 packages) · E2E **ok (215/215)** · topology audit **PASSED** (44 features · 16 root + 310 module rows · 1576 steps).

---

## T010 (cont.) — PR #77 fold review (`pullrequestreview-5230289661`)

The fold review confirmed **F-2 / F-3 / F-4 / F-6** correct and verified, and raised three follow-ups (the reviewer also corrected **their own F-1 mechanism** — the `┌`-delimited-frame premise was wrong for this harness):

| # | Finding | Resolution |
| --- | --- | --- |
| **G-1** | `atRestFrame` is the **identity** in every measured capture: the harness writes the compose keys (incl. any `Tab`) **before** the sync-marker wait (`tui_keys.go:45-50`; marker `┌`, `scenario_context.go:260`), so a navigated capture carries exactly **one** painted frame and the `next >= 0` branch never fires. The doc premise was false and the truth claimed a guarantee the code could not enforce. | **Dropped** `atRestFrame` + its dead branch; the E2E `Then` asserts `tuiCursorRows(renderedOutput(...)) == 0`. Corrected the `chat/dsl.md` row (scope = the rendered **suggestion block**), `research.md` D3/D4, and `STATUS.md`. |
| **G-2** | The flipped predicate had a reproduced **false failure**: `TrimLeft` + `HasPrefix("> ")` counted any `> `-leading line, so a suggestion whose *text* is `> quoted reply` (rendered with the 4-space unselected prefix) miscounted as a cursor row. | `tuiCursorRows` now matches the exact rendered cursor row prefix `"  > "` (2-space model padding + the `> ` cursor prefix); an unselected `> `-leading text renders `"    > "` and is not counted. Added the unit pin `tests/e2e/steps/tui_chrome_test.go`. |
| **G-3** | `STATUS.md` item (c) (plan-side `entry.txt` seeds `version 016`) was **stale in the commit that fixed it**. | Dropped item (c). |

**Re-verification after the G-fold:** `gofmt -l .` clean · `go vet ./...` clean · `go test -count=1 ./tests/e2e/steps/` (new pin) green · E2E **ok (215/215)** · `make verify` OK · topology audit **PASSED**.
