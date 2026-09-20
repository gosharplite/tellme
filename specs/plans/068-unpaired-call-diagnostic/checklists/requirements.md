# 規格品質檢查清單：Surface a Gemini/Vertex round's unpaired tool calls as a diagnostic (round 068)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/068-unpaired-call-diagnostic`

**Spec 路徑**: `specs/plans/068-unpaired-call-diagnostic/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（把 Gemini/Vertex 一輪的未配對呼叫以 user-visible diagnostic 呈現；stderr、受 chrome gate 管）
- [x] 沒有把實作技術、框架或程式細節寫成需求（偵測 seam S-6 留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（`M == N`／`M = 0`／多輪／media／非終端 `-r`／OpenAI-compatible 不受影響）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 唯一故事，P1）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…FR-005 + NFR-001/002）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-7 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [ ] 只有高影響缺口才升級到 `/axb-clarify` — **本輪升級 2 題（Q1 surface/routing · Q2 loudness）**：RF-067-1 本身即 operator-gated，因其新增 **user-visible surface**，會改變正式驗收標準
- [x] 本輪 clarify 題數控制在 1 至 3 題
- [x] 低風險未定細節已用假設揭露（A1–A6；S-2/S-3/S-4 標為 proposed／clarify；S-6 標為 research）
- [ ] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **FR-006（Q1）/ FR-007（Q2）仍開放，且阻塞 `/axb-spec-by-example`**

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（`M < N` 時 stderr 出現 diagnostic；`M == N` 時不出現）
- [x] 成功標準可量測、可驗證且技術中立（SC-001 為可紅載體；SC-002 為負向；SC-003 為 byte/shape 保真；SC-004 為 falsifiability）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **性質**：本輪**不是** hardening/parity NOOP round — 它新增一個 **user-visible `stderr` diagnostic**，因此 `/axb-spec-by-example` + `/axb-dsl-refine` **不是 NOOP**（與 rounds 066/067 相反）。
- **誠實揭露（重要）**：現行 loop **每呼叫必 append 一個 `tool` 結果**（`agentloop.go`），故 production 今日 **恆 `M == N`**；`M < N` 只可能來自 hand-built/partial/corrupted `prior` 或**未來** out-of-order/concurrent dispatch（[#36](https://github.com/gosharplite/tellme/issues/36) item 3）。本輪診斷為 **defensive safety-net**，不得誇稱為高頻 user surface。
- **現況（grounded 2026-09-20 `dev` @ `e864d9a`）**：`roundBuilder.flush`/`unpaired()`/`UnpairedCallIDs` 已在 code 層記帳但**無 live consumer**；adapter 無 logging seam；`tools.OutputSink`（ADR 0021 ctor 注入）為既有注入前例；chrome + `turns.log`（ADR 0022）為 CLI-owned；colour terminal-gated（ADR 0023）。
- **治理**：預期 `techstack.md` MODIFY + **新增 ADR**（或 ADR 0037 forward-annotation）；`/axb-api-plan` NOOP；`/axb-data-plan` NOOP。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口 — **阻塞於 Q1（surface/routing）+ Q2（loudness）之 operator 拍板**

**備註**: RF-067-1 之 operator gate 已由 operator 覆蓋（*"the goal is to resolve RF-067-1"*）；但**診斷的表面與 loudness** 仍屬正式驗收標準層級，需 Q1/Q2 拍板後方可進入 `/axb-spec-by-example`。偵測 seam（S-6）交 `/axb-technical-research`。
