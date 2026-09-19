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
- [x] 沒有把實作技術、框架或程式細節寫成需求（toolchain/gate 形式已由 Q2/Q3 拍板，spec 以意圖 + 鎖定決策記錄）
- [x] 邊界情況已涵蓋主要高風險情境（drift、hand-edit、absent binary、truth conflict、excluded entity、fork drift）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 three models · US2 drift gate · US3 truth/ADR）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-009/FR-010/NFR-004）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1 scope · Q2 toolchain · Q3 gate strictness）
- [x] 本輪 clarify 題數控制在 1 至 3 題（實際 3 題，one at a time；session 預算 5 未超）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（Q4/Q5 → A6/A7）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（**無**；Q1–Q3 已鎖定，Q4/Q5 已收斂為假設）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（lint 0/0、render --check、drift witness、absent-binary witness、verify 成員、truth/ADR 記錄）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **Clarify round 1 CLOSED** — Q1 → 3（三模型：product + quality + environment）；Q2 → 1（採用 modelith fork + `make modelith-lint|render|check`）；Q3 → 1（`modelith-check` 為 `make verify` 的 zero-tolerance 成員；二進位缺席即硬失敗並具名安裝指令）。Q4 → A6（模型為描述性 docs，truth 優先）；Q5 → A7（隨 truth 變更刷新，drift gate 為安全網）。
- **Reopens ADR 0011 D10**（"no modelith toolchain"）⇒ 新 **ADR 0030**（於 `/axb-technical-research` 產出）。
- US1 範圍已由 Q1 放寬至三模型；FR-001/FR-003/FR-010 已同步。

## Ready 判定

- [x] 已可進入後續規劃（`/axb-technical-research` → ADR 0030 + techstack truth）
- [ ] 仍需先補高影響需求缺口

**備註**: Clarify 已關閉；下一步 `/axb-technical-research`（research.md + ADR 0030 + `specs/truth/techstack.md` MODIFY）。`/axb-spec-by-example` 預期 **NOOP**（docs/tooling round，無使用者旅程）。
