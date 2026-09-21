# 規格品質檢查清單：Skill frontmatter block-scalar descriptions (round 075)

**建立日期**: 2026-09-21

**Feature Directory**: `specs/plans/075-skill-frontmatter-block-scalars`

**Spec 路徑**: `specs/plans/075-skill-frontmatter-block-scalars/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（S-2/S-3 明列為 research 決策，非需求）
- [x] 邊界情況已涵蓋主要高風險情境（inline `>`、空 block、`name` block、CRLF、引號、重名）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用（`Skill` 形狀不變，已說明）

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 P1 為缺陷本身；US2 P2 為讀取器完整覆蓋）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下（FR-001…003 / FR-004…005）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 未知鍵忽略；NFR-001…003）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（本輪判定：無高影響缺口 → 0 題）
- [x] 本輪 clarify 題數控制在 1 至 3 題（0 題）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A5；S-2/S-3/S-4 交 research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃（無殘留；不阻塞）

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（SC-001 100%/0、SC-002 byte-identical、SC-003 gates+zero-dep）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **參考實作同缺陷（重要揭露）**：`tell-me-go` 的 `parseSkill` 亦為 line-based，同樣無法解析 block scalar；本輪因此為**刻意偏離參考的穩健化**，須在 truth row 與 ADR 記錄為 divergence，不得宣稱 parity。
- **範圍界定**：只修 `tellme` 的 loader；三個 `domain-model-*` 技能非本方擁有，列為 out of scope。
- **truth owner**：`techstack.md` → §Skills → *Skills catalog (load)* row 由本輪 `/axb-technical-research` MODIFY。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: 可進入 `/axb-spec-by-example`（預期 1 條 acceptance journey：block-scalar 技能的列表顯示）與 `/axb-technical-research`（S-2/S-3/S-4 決策 + techstack MODIFY + 新 ADR）。E2E interface truth 交由 `/axb-dsl-refine`（預期需一個可撰寫 block-scalar 技能的 Given）。
