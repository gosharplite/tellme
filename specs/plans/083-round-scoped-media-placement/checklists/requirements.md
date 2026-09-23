# 規格品質檢查清單：Round-scoped media placement (round 083)

**建立日期**: 2026-09-23

**Feature Directory**: `specs/plans/083-round-scoped-media-placement`

**Spec 路徑**: `specs/plans/083-round-scoped-media-placement/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（多媒體回合的送出順序修正）
- [x] 沒有把實作技術、框架或程式細節寫成需求（wire 形狀屬系統真相，已以行為語言描述）
- [x] 邊界情況已涵蓋主要高風險情境（N=0/1/部分媒體、重播、provider 錯誤）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 修復 → US2 相容性）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-007/008、NFR-003/004）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（本輪未升級）
- [x] 本輪 clarify 題數控制在 1 至 3 題（0 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（S-1…S-4 交由研究決議，A2）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（無）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（多媒體回合完成 + 相鄰性 + 單圖片凍結 + Gemini）
- [x] 成功標準可量測、可驗證且技術中立（SC-001…SC-005）
- [x] 假設只表達前提與邊界，沒有偷渡新需求（A1…A6）
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- 本輪的缺陷目前**無任何測試可變紅**（既有 `toolExchangeChronologyOK` 只斷言先行、未斷言相鄰，且 fixture 未用媒體工具）—— 見 A4；本輪 MUST 補上 loop-tier 單元 pin 與 E2E 相鄰性見證（NFR-004）。
- 修正 MUST NOT 回歸單圖片（已出貨）路徑或 Gemini 擺放 —— 見 I-3/I-4、FR-005/FR-006。
- 媒體不落地為既有契約（A3）—— 續接 session 的重播對話不含媒體，屬既有行為。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: clarify **未升級（0 題）** —— 核心行為由 issue [#167](https://github.com/gosharplite/tellme/issues/167) 鎖定；殘留技術選擇（S-1…S-4）交由 `/axb-technical-research`。可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
