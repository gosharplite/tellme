# 規格品質檢查清單：Make the callable MCP wire name discoverable in the offered declaration (round 077)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/077-mcp-tool-name-presentation`

**Spec 路徑**: `specs/plans/077-mcp-tool-name-presentation/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（讓 MCP 工具的可呼叫線路名稱在提供宣告中可被正面發現；無描述時合成文字指名線路名稱；或研究判定 NOOP）
- [x] 沒有把實作技術、框架或程式細節寫成需求（S-2/S-3/S-4/S-5 明列為 research 決策，非需求）
- [x] 邊界情況已涵蓋主要高風險情境（空／非空描述、hash 截斷名稱、封閉線路投影、NOOP）
- [x] 關鍵實體與成功標準已補齊（`MCPToolDefinition` 形狀不變、僅描述層級、schema 位元組不變）

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 P1 為可發現性總體；US2 P2 為無描述時的具體子案）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…003 / FR-004…005 + NFR-001）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 見證／NOOP 記錄；NFR-002…004）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（本輪判定：目標明確、設計由 ADR 0025 D1 + issue 選項鎖定，無高影響缺口 → 0 題）
- [x] 本輪 clarify 題數控制在 1 至 3 題（0 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A5；S-1…S-5 交 research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（無殘留；不阻塞）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（宣告可發現性 + 備援指名線路名稱 + schema 不變）
- [x] 成功標準可量測、可驗證且技術中立（SC-001 100%/0、SC-002 逐位元組指名、SC-003 gates + 零 schema 差異、SC-004 NOOP 由可觀察事實支撐）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **研究先行、NOOP 可接受（最重要）**：#155 明示此為研究導向、可能 NOOP。spec 必須讓「先驗證是否確有缺口、再決定是否變更」成為一等公民（S-1、FR-006、SC-004），不得投機性改動。
- **不得竄改伺服器定義（I-1）**：任何 tellme 說明 MUST 附加於伺服器文字之外（ADR 0025 D1）；提供宣告是 tellme 自己的信封。
- **線路形狀中性（I-3）**：變更僅限描述層級；封閉線路投影與 vendor-extension floor 不受影響；schema 位元組必須不變。
- **Option 3 已交付**：回合 076 已讓未知名稱摺回結果列出可用工具名稱；#155 剩餘面為 Option 1（備援文字）與 Option 2（命名空間說明）。
- **無系統提示變更（A4 / ADR 0025 D4）**：ask 為宣告承載。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: 可進入 `/axb-technical-research`（**先驗證** fallback 是否在起作用 + S-2…S-5 決策 + ADR/truth row + domain-model 判定）。`/axb-spec-by-example` 視研究結果而定（NOOP 則預期無新 acceptance journey）。E2E interface truth 交由 `/axb-dsl-refine`（預期以 fake MCP server / fake provider 觸發並觀察提供宣告；NOOP 則 NOOP）。
