# 規格品質檢查清單：every `[Tool Output]` line grey (round 058)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/058-grey-tool-output-content`

**Spec 路徑**: `specs/plans/058-grey-tool-output-content/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用現況常數／檔案位置僅為可追蹤性）
- [x] 邊界情況已涵蓋主要高風險情境
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 the block-wide grey — P1；單一故事）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…FR-006）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；S-1…S-6 為跨故事設計前提；I-1…I-3 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1：trailing partial line 是否屬於「all lines」）
- [x] 本輪 clarify 題數控制在 1 至 3 題（1 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1 trailing partial 維持 dropped）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **Q1 為 OPEN，惟可先以 A1 推進**（trailing partial 從不輸出，維持 drop 不影響本輪驗收）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **本輪為 round 057 的單點延伸**：round 057 只灰化 `[Tool Output]` 標頭 + 兩條分隔線（假設 A3 明確保留內容行 plain）；本輪要求**所有** `[Tool Output]` 行（含內容行）灰化，**明示反轉 round 057 A3**。round 057 package 已 frozen，不修改。
- **安全性**：內容行先 `sanitizeControl` 再包 grey（S-2 / I-3），指令輸出的控制序列無法逃出 wrapper。
- **檔案腿**：`turns.log`/`-t` 維持 plain（S-4 / I-2，round-053/054 規則）。
- **A4/A5** — 新增 **ADR 0028** 記錄 block-wide grey（supersede round 057 A3）。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: Q1（trailing partial line）為唯一缺口，已以 A1 揭露（從不輸出、維持 drop）；可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
