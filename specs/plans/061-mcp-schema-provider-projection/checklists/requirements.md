# 規格品質檢查清單：provider-wire projection of MCP-relayed tool schemas (round 061)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/061-mcp-schema-provider-projection`

**Spec 路徑**: `specs/plans/061-mcp-schema-provider-projection/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況常數僅為可追蹤性與重現依據）
- [x] 邊界情況已涵蓋主要高風險情境（含「標準但 Gemini 不支援」的關鍵字、巢狀關鍵字、無標註 server、非物件 schema）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 turn 完成 — P1；US2 宣告保真 — P2；US3 防回歸 — P3）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-003；US2: FR-004…FR-006；US3: FR-007…FR-009）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-5 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（CQ-1 修正位置、CQ-2 投影範圍、CQ-3 truth/ADR 治理）
- [x] 本輪 clarify 題數控制在 1 至 3 題（3 題，一題一問）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A7）
- [ ] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **CQ-1/CQ-2/CQ-3 皆為阻塞**：未收斂前不得進入 `/axb-dsl-refine` / `/axb-tasks`（會改變驗收標準與既有 truth 行為）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（SC-001 有可重現的紅→綠 witness）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **缺陷本體**：MCP server 宣告的 schema 內含 vendor extension `x-mcp-header`（實測 GitHub server 45 個 tool 中 37 個帶此標註），tellme 依 round-056/ADR-0025 以 **verbatim** 方式放入 `MCP_PAYLOAD`；Gemini 的 `functionDeclarations[].parameters` 是 **closed proto** `Schema`，未知欄位即 400，**整個** request 失敗（索引對位已證實：`properties[0].value.properties[2]`＝`MCP_PAYLOAD`→`properties.owner`）。
- **索引對位證據**：`add_comment_to_pending_review`（字母序第一個 tool = `function_declarations[7]`，native 7 個之後）排序後 properties 為 `body(0), line(1), owner(2)←x-mcp-header, path(3), pullNumber(4), repo(5)←x-mcp-header` — 與錯誤訊息兩個索引完全一致。
- **家族差異**：OpenAI-compatible（`client.go:156`）容忍，所以只有 Gemini 掛 — 這使修法有「家族範圍」的取捨。
- **#64/round 031 的復發軸線**：round 031 修的是 tellme **自己** tool 的 `required ⊆ properties`，gate 只覆蓋 `agentTools()`，不含 MCP relay、也不含未知關鍵字。
- **參考實作**：`tell-me-go` 以 **typed per-family schema model** 轉換，vendor 標註根本進不了 Vertex payload（parity precedent）。
- **實測殘留風險**：同一 GitHub server 另有 `anyOf` ×2、`additionalProperties` ×2 — 若投影只處理 `x-*`（CQ-2 → i），這 4 個 tool 仍是潛在 400。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口

**備註**: spec 已完整記錄 defect、重現證據與不變式；但 **CQ-1（投影位置 A/B/C/D）、CQ-2（投影範圍 i/ii）、CQ-3（ADR 治理）尚未拍板**，且三者都會改變正式驗收標準與既有 truth（round-056 verbatim row）。依 `axb-specify` Phase 2，**必須先 `/axb-clarify`（一題一問）**，未收斂前停止，不得自行假設答案，也不得進入 `/axb-spec-by-example` / `/axb-technical-research`。
