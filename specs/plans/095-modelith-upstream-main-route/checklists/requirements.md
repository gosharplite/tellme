# 規格品質檢查清單：Round 095 `095-modelith-upstream-main-route`

**建立日期**: 2026-09-27

**Feature Directory**: `specs/plans/095-modelith-upstream-main-route/`

**Spec 路徑**: `specs/plans/095-modelith-upstream-main-route/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（route 為文件/工具鏈事實，非產品實作）
- [x] 邊界情況已涵蓋主要高風險情境（EC-001 main 前進的 accepted cost；EC-002 既有拒絕行為）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US-1 唯一）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…007、NFR-001）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（本輪為空）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（本輪 0 題；directive 已鎖定 route）
- [x] 本輪 clarify 題數控制在 1 至 3 題（N/A — 未進入 clarify）
- [x] 低風險未定細節已用假設揭露（A1 route 拼寫 `@main` vs pin）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞（無；A1 為假設，不阻塞）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（SC-001…004）
- [x] 所有規範性條目（FR / NFR / SC / EC）皆已標註驗證意圖（Verification Intent）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- 本輪為 toolchain/record round：多數 FR 為「文件/工具鏈表面」條目，**無 Gherkin 外部可觀察性**；
  故以 `carried manual` witness（grep / git-diff / absent-binary 訊息）承接，並在 `tasks.md` 標為
  `[WITNESS]`。EC-001 為 `accepted-unwitnessed`（ADR-0012 spurious-red class，ADR 0065 明文接受）。
- FR-003 / SC-002 對應一個真正可觀察的載體：`make modelith-check` 的 absent-binary 訊息（W1）。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: Ready for `/axb-spec-by-example`（預期 NOOP）→ `/axb-technical-research`（本輪唯一的 truth owner）。
