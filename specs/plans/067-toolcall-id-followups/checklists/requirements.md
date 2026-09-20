# 規格品質檢查清單：Gemini/Vertex tool-call id follow-ups (round 067)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/067-toolcall-id-followups`

**Spec 路徑**: `specs/plans/067-toolcall-id-followups/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（Gemini/Vertex 一輪的**未配對呼叫可觀測** + `functionCall.id` **provider 優先**；皆限 `internal/infrastructure/llm/gemini`）
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；可觀測形式 S-3、fallback 拼法 S-4、跨家族 id 範圍 S-5 留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（`M = N`／`M < N`／`M = 0`、provider id 缺失/空字串、重複 provider id、replay、空 `ToolCallID`、部分/全部 media、OpenAI-compatible 不受影響）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 未配對呼叫可觀測 — P1（#136 item A / RF-066-7，並退役 RF-066-8）；US2 provider `functionCall.id` 優先 — P2（#136 item B / RF-066-2））
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-005 + NFR-001；US2: FR-006…FR-010 + NFR-002）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-8 為不變式，不是新需求）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify` — **本輪不升級**：目標明確（關閉 [#136](https://github.com/gosharplite/tellme/issues/136)）、兩項變更已由 issue 載明並在地測量；殘餘選擇為技術性（S-3 可觀測形式、S-4 fallback 拼法、S-5 跨家族 id 範圍），依 round-063/065/066 前例交由 `/axb-technical-research`
- [x] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題 — **未使用**（0 題）
- [x] 低風險未定細節已用假設揭露（A1–A7；S-3/S-4/S-5 明確標為 proposed／research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（US1 以「`M < N` 時未配對呼叫 id 可觀測」為紅→綠見證；US2 以「provider id 在場則沿用、缺席則決定性 fallback」為紅→綠見證）
- [x] 成功標準可量測、可驗證且技術中立（SC-001/SC-002 為 shape 層可紅 pin；SC-003/SC-004 為 byte/shape 與回歸 pin；SC-005 為 falsifiability witness；SC-007 為 RF-066-8 退役條件）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **性質**：本輪是 **hardening/parity round，不是 defect fix** — tellme 的 loop **循序** 執行 tool call 且按呼叫序 append 結果，故合成 `call_<n>` 與位置式配對**今天永遠正確**；id **值**不經任何 shipped surface 對使用者可見。RF-066-7 是 ADR 0036 §Forward 記錄的記帳缺口；RF-066-2 是 round 066 刻意延後的 reference-parity 軸。
- **現況（grounded 2026-09-20 @ `dev` `30f54a1`；#136 於 `3637ec2` 同述）**：`gemini/client.go` `parseResponse` 以 `call_<n>` 合成 id，**未讀** provider 的 `functionCall.id`（decode struct 僅 `Name`+`Args`）；`roundBuilder.flush` 只 emit `M` 個 `functionResponse` 並清空 `pending`，未配對呼叫**靜默丟棄**（無 error/log/accessor）；`roundBuilder.bind` **已**依 `ToolCallID` 配對（FIFO fallback）— round 066 核心。`agentloop.go` 每個呼叫都設 `ToolCallID`（live: `tc.ID`；replay: `call_step_<n>`）。OpenAI-compatible wire `"id"`（`:126`）+ `tool_call_id`（`:136-137`）**已**存在 —— id **值**因此流向**兩個**家族。
- **參考實作（`tell-me-go`）**：`fromSDKFunctionCall` 先讀 provider `f.ID`、否則 `gemini-call-<index>-<name>` 決定性 fallback；**空 id = invalid**（`isInvalidToolPart` 剝除）。
- **既有 pin**：round-066 `client_ids_test.go`（`…ToolPartsCarryIDs`/`…OutOfOrderResultsPairByIdentity`/`…EmptyToolCallIDOmitsID`/`…UnmatchedToolCallIDFallsBackToFIFO`/`…ReplayedStepIDsPairByIdentity`）+ round-065 `TestRequestBody_ShortRound_DropsUnpairedNames`（`N=2 M=1` 殘餘 — 本輪以 SC-007 退役）。
- **治理**：預期 **新增 ADR**（id 來源 + 跨家族決定；extend ADR 0036、標註其 §Forward RF-066-2/RF-066-7/RF-066-8）或 ADR-0036 D-note；`/axb-api-plan` NOOP；`/axb-data-plan` NOOP；`/axb-spec-by-example` 預期 NOOP（無 user-visible 行為，除非 S-3 選為 user-visible diagnostic）；`/axb-dsl-refine` 預期 NOOP 或小幅 MODIFY。預期不新增 capability/config/dependency。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: **Clarify 未升級（0 題）** — 目標明確（關閉 [#136](https://github.com/gosharplite/tellme/issues/136)）、現況已在地測量、兩項變更已載明；殘餘為技術選擇（S-3/S-4/S-5），交由 `/axb-technical-research`。**無殘留 `NEEDS CLARIFICATION`**，可進入後續規劃（`/axb-spec-by-example` 預期 NOOP + `/axb-technical-research`）。
