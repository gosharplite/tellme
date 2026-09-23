# 規格品質檢查清單：Backward-counting turn indices in the `-l`/`--list` role headers (`[USER] - N` / `[MODEL] - N`) (round 082)

**建立日期**: 2026-09-23

**Feature Directory**: `specs/plans/082-listing-backward-turn-indices`

**Spec 路徑**: `specs/plans/082-listing-backward-turn-indices/spec.md`

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

- [x] 使用者故事依商業價值與交付順序排序
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`
- [x] 本輪 clarify 題數控制在 1 至 3 題（**未升級 — 0 題**；issue 已鎖定行為與設計原則）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（S-1…S-4 交由 `/axb-technical-research`；A2 明示）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（**無阻塞項**）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- 已將 issue 的 *Design Principles*（turn-based / backward / true distance / colour+raw / presentation-only）逐條落為 I-1…I-7 與 FR-001…FR-006，並把 *Proposed Implementation Touch Points* 轉為待研究決議 S-1…S-4（未在本 skill 決定）。
- 奇數 `-l N` 的前導半回合（真實距離）為本輪最易誤作的高風險點，已明確寫成 FR-004 + SC-003 + 邊界情況。
- `-l N` 語意維持「最後 N **messages**」不變（I-7 / NFR-001），避免與「N 回合」混淆。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: 本輪為純呈現增強（`-l` listing 表頭索引），核心行為由 issue 鎖定、clarify 未升級（0 題）；可直接進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
