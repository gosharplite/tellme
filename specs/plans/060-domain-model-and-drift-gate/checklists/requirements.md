# 規格品質檢查清單：tellme domain model + drift gate (round 060)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/060-domain-model-and-drift-gate`

**Spec 路徑**: `specs/plans/060-domain-model-and-drift-gate/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（toolchain/gate 形式以 Q2/Q3 拍板，spec 只寫意圖）
- [x] 邊界情況已涵蓋主要高風險情境（drift、hand-edit、absent binary、truth conflict、excluded entity、fork drift）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 model · US2 drift gate · US3 truth/ADR）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-009/FR-010/NFR-004）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [ ] 只有高影響缺口才升級到 `/axb-clarify`
- [ ] 本輪 clarify 題數控制在 1 至 3 題（本輪 Q1–Q3；Q4/Q5 為低風險，預期以 spec 假設收斂）
- [ ] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（Q4/Q5）
- [ ] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（Q1–Q3 阻塞；Q4/Q5 不阻塞）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（lint 0/0、render --check、drift witness、truth/ADR 記錄）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **Pending clarify** — Q1（scope）· Q2（toolchain）· Q3（gate strictness/hermeticity）；Q4（authority boundary）· Q5（lifecycle）待 Q1–Q3 後以假設收斂。
- Spec 目前為 draft；Q1–Q3 拍板後回刷 US1 範圍、FR-004 與 Key Entities。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口（Q1–Q3）

**備註**: 等待 clarify round 1（Q1–Q3，one at a time）拍板；`/axb-technical-research` 於收斂後進行（ADR 0030）。
