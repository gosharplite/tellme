# 規格品質檢查清單：tool-chrome colour, a 500-rune `[Tool Action]` cap, and a payload increment (round 057)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/057-tool-chrome-colour-and-payload-delta`

**Spec 路徑**: `specs/plans/057-tool-chrome-colour-and-payload-delta/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用的常數／檔案位置僅作為「現況量測」與可追蹤性）
- [x] 邊界情況已涵蓋主要高風險情境
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 chrome colour · US2 payload increment — both P1；US3 the 500-rune cap — P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…005 → US1；FR-006…009 → US2；FR-010…012 → US3）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；S-1…S-5 為跨故事設計前提；I-1…I-4 為跨故事不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1 哪條 payload line 帶 delta；Q2「前一個 payload」的定義與儲存；Q3 budget 移除／無前值／負 delta）
- [x] 本輪 clarify 題數控制在 1 至 3 題（3 題，一次問一題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1「兩條水平線」= 開／閉分隔線；A3 內容行維持 plain；A2 顏色碼與 whole-line wrap）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **Q1/Q2/Q3 為 OPEN，且阻塞後續規劃**（payload 行形狀屬正式驗收標準）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **三項請求皆為 chrome 變更**：(1) `[Tool Output]` 標頭行 + 兩條分隔線改 grey、`[Tool Action]` 改 yellow；(2) `argValueCap` 189 → 500；(3) payload 行的 token 段由 `<tokens>/<budget>` 改為 `+<delta> <tokens>`（例：`+100 ~203148`）。
- **唯一真正的設計缺口在 (3)**：`history.Entry` 目前未持久化任何 payload/token 數；delta 需有「前一個 payload」的來源（in-memory 或 persisted）——見 Q2。
- **顏色沿用 round-054 政策**：terminal-only + `-r` off 才輸出；`stdout` 與 `turns.log` 維持 plain（I-2）。顏色碼取 reference（`colorGray`／`colorYellow`），元素集合為 tellme 自有（round-054 慣例）。
- **A4** — 除非 Q2 → B（新增 persisted `Entry` 欄位），`/axb-data-plan` 與 `/axb-api-plan` 為 NOOP。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口

**備註**: Q1/Q2/Q3 尚待 operator 拍板（一次一題），收斂後即可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
