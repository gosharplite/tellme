# 規格品質檢查清單：A recoverable, per-turn-bounded unknown tool name (round 076)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/076-recoverable-unknown-tool`

**Spec 路徑**: `specs/plans/076-recoverable-unknown-tool/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（未知名稱 → 可回復摺回 → 續行；以每回合上限收斂）
- [x] 沒有把實作技術、框架或程式細節寫成需求（S-3/S-4/S-5/S-6 明列為 research 決策，非需求）
- [x] 邊界情況已涵蓋主要高風險情境（混合回合、reason 疊加、空註冊表、未達上限即修正、未執行不記錄）
- [x] 關鍵實體與成功標準已補齊（`ToolCall` 形狀不變、未執行不記錄；領域模型是否需更新交 research）

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 P1 為缺陷本身；US2 P2 為安全護欄，建立於 US1 之上）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…003 / FR-004…005 + NFR-001）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 片語/exit 7 契約；NFR-002…004）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（本輪判定：設計已由上鎖定，無高影響缺口 → 0 題）
- [x] 本輪 clarify 題數控制在 1 至 3 題（0 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A5；S-2…S-6 交 research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（無殘留；不阻塞）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（摺回 + 續行 + 上限終止）
- [x] 成功標準可量測、可驗證且技術中立（SC-001 100%/0、SC-002 上限 N 次 + exit 7、SC-003 gates + 零線路差異）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **上限為護欄核心（重要）**：tellme 無重複/失控偵測，唯一界限為 `MAX_TOOL_LOOP`（預設 1000，每次付費呼叫）；若不以每回合上限收斂，US1 會把快速失敗變成昂貴空轉。US2 與 NFR-001 必須據此成立。
- **片語/exit 7 契約是收窄**：round-008 FR-010 保留給 bound reached / cap exhausted / no tools registered；此為 likely **ADR + truth row** 變更，交 truth owner 記錄。
- **範圍界定**：#154 為本輪唯一 anchor（DoD = 關閉 #154）；#155 為獨立、research-gated 的呈現議題，**out of scope**。
- **reason 疊加歸屬**：未知 vs 空 reason 的優先序為單一歸屬問題，須由研究拍板並以測試釘住（邊界情況已列）。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: 可進入 `/axb-spec-by-example`（預期 acceptance journeys：未知名稱可回復續行、每回合上限終止）與 `/axb-technical-research`（S-2…S-6 決策 + 片語/exit 7 契約收窄的 ADR + truth row + domain-model 判定）。E2E interface truth 交由 `/axb-dsl-refine`（預期需以 fake provider 觸發未知名稱的 Given）。
