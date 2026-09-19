# 規格品質檢查清單：MCP tool-call reason — a tellme-owned `reason` alongside the server payload (round 056)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/056-mcp-tool-call-reason`

**Spec 路徑**: `specs/plans/056-mcp-tool-call-reason/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求
- [x] 邊界情況已涵蓋主要高風險情境
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 reason visible · US2 server untouched — both P1）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…004 → US1；FR-005…007 → US2）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；S-1…S-6 為跨故事設計前提）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1 如何向模型請求 reason；Q2 未包裝/legacy 呼叫的處置）
- [x] 本輪 clarify 題數控制在 1 至 3 題（2 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **Q1 → A LOCKED；Q2 → B（STRICT）LOCKED**；本輪無殘留 `NEEDS CLARIFICATION`

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **設計已於 spec 收斂**：S-1…S-6 記述 operator 於本輪會話確認的設計（server definition 不可變；tellme 請求並渲染 `reason`；呼叫 JSON 為 `{reason, MCP_PAYLOAD}`，僅轉送 `MCP_PAYLOAD`；不動 system prompt；MCP-only；server 自帶 `reason` 不衝突）。
- **Q1（LOCKED → A）** — tellme 以自身宣告向模型請求 `reason`：頂層 `reason`（required，tellme 擁有）+ `MCP_PAYLOAD`（原樣承載 server 的 schema）；server definition 不改，system prompt 不改。
- **Q2（LOCKED → B, STRICT）** — 未使用 envelope 的 MCP 呼叫（缺 `reason`/空白 `reason`/非物件 `MCP_PAYLOAD`）一律拒絕：不接觸 server，回可恢復結果要求模型以 envelope 重試；缺 `MCP_PAYLOAD` 但 `reason` 有效 = 空 payload `{}`。
- **A4** — round-032 既有 MCP E2E fixture（scripted `Arguments: "{}"`）需於實作半更新；本 skill 不寫入 `specs/truth/**`。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: Q1/Q2 皆已 LOCKED；spec 主體、故事切分、FR 歸戶、成功標準與 clarify 皆已收斂，可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
