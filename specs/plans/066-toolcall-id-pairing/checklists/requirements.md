# 規格品質檢查清單：Gemini/Vertex — pair a round's tool results to their calls by `ToolCallID` (round 066)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/066-toolcall-id-pairing`

**Spec 路徑**: `specs/plans/066-toolcall-id-pairing/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（Gemini/Vertex 一輪的 `functionCall`/`functionResponse` **id 連動** + 依 `ToolCallID` 配對；FIFO 名稱配對保留為 fallback）
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；id 來源、fallback 拼法、unmatched 記帳方式留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（N=1／N=0 保真、結果亂序、空 `ToolCallID`、`M < N`、未知 id、部分/全部 media、tool error／reason-less、同 turn 兩輪、無 provider id、OpenAI-compatible 不受影響）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 一輪的 call/result 在 wire 上 id 連動 — P1；US2 依 id 配對、與到達序無關 — P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-005；US2: FR-006…FR-010）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-7 為不變式，不是新需求）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify` — **本輪不升級**：目標明確（關閉 [#134](https://github.com/gosharplite/tellme/issues/134)）、現況與修法已由 issue 載明並在地測量；殘餘選擇為技術性（id 來源/fallback/unmatched 記帳），依 round-063/065 前例交由 `/axb-technical-research`
- [x] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題 — **未使用**（0 題）
- [x] 低風險未定細節已用假設揭露（A1–A7；S-4/S-6 明確標為 proposed／research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（US1 以「每個 call/result part 帶 id、response id = call id」為紅→綠見證；US2 以「亂序結果仍配到自己的 call」為紅→綠見證，另以 replay 的 id-primary 配對（FIFO fallback 為防禦路徑）保真）
- [x] 成功標準可量測、可驗證且技術中立（SC-001…SC-003 為 shape 層可紅 pin；SC-004 為 byte/shape 保真回歸 pin；SC-006 為 falsifiability witness）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **性質**：本輪是 **hardening/parity round，不是 defect fix** — tellme 的 loop **循序** 執行 tool call 且按呼叫序 append 結果，故 round 065 的位置式配對（ADR 0035 D2）**今天永遠正確**；id-keyed 配對是為 **concurrent** dispatch（[#36](https://github.com/gosharplite/tellme/issues/36) item 3）預備。**無 user-visible 行為改變**，見證為 request-body shape。
- **現況（grounded 2026-09-20 @ `dev` `babeff7`）**：`gemini/client.go` `buildContents` 以 `pending []string`（呼叫 **名稱** FIFO）配對，emit 的 part 為 `{"functionResponse":{"name":…,"response":…}}`，**無 `id`**；`parseResponse` 以 `call_<n>` 合成 id；`agentloop.go` 每個呼叫都設 `ToolCallID`（live: `tc.ID`；replay: `call_step_<n>`）。OpenAI-compatible wire **已**用 `tool_call_id`（`:136-137`）——Gemini 家族是落後的一方。
- **參考實作（`tell-me-go`）**：`FunctionCall.ID`/`FunctionResponse.ID` 端到端存在，**provider id 優先**、否則 `gemini-call-<index>-<name>` 決定性 fallback；**空 id = invalid**（`isInvalidToolPart` 剝除）。其 executor **併發**執行並把 `call.ID` 拷進每個 response。註：參考仍是 **位置式組裝**，真正的 `map[id]` 查找是**超出**參考的強化。
- **修法（issue 已載明）**：wire 上帶 id；`buildContents` 依 id 配對（FIFO fallback 為防禦路徑）；採用參考的 id 紀律。unmatched 記帳**未**交付（邊界丟棄不變 — F-066-1，改列 RF-066-7/RF-066-8）。
- **既有 pin 需擴充**：round-065 的 `TestRequestBody_MultiCallRound_BatchesFunctionResponses` / `…_NoMedia_BatchesResults` / `TestRequestBody_ThreeCallRound_BatchesResults` / `TestRequestBody_ShortRound_DropsUnpairedNames` 目前 pin 住 **無 id 的 batched 形狀**，本輪加上 id 軸（且 S-6 可能讓 short-round pin 改以 exact unmatched id 記帳）。
- **治理**：預期 **新增 ADR**（id-keyed 配對；extend ADR 0035、標註其 D2 的 FIFO 註記與 §Forward RF-065-1）；`/axb-api-plan` NOOP；`/axb-data-plan` NOOP；`/axb-spec-by-example` 預期 NOOP（無 user-visible 行為）；`/axb-dsl-refine` 預期 NOOP 或小幅 MODIFY（視 fake 能否觀察 id）。預期不新增 capability/config/dependency。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: **Clarify 未升級（0 題）** — 目標明確（關閉 [#134](https://github.com/gosharplite/tellme/issues/134)）、現況已在地測量、修法已載明；殘餘為技術選擇（S-4/S-6），交由 `/axb-technical-research`。**無殘留 `NEEDS CLARIFICATION`**，可進入後續規劃（`/axb-spec-by-example` 預期 NOOP + `/axb-technical-research`）。
