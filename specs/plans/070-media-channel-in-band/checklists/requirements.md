# 規格品質檢查清單：The media channel in-band (round 070)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/070-media-channel-in-band`

**Spec 路徑**: `specs/plans/070-media-channel-in-band/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（讓工具的媒體效果 in-band：由 `Execute` 回傳、loop 翻譯；刪除 per-call context collector）
- [x] 沒有把實作技術、框架或程式細節寫成需求（shape (a)/(b) 留待 `/axb-technical-research`；S-1）
- [x] 邊界情況已涵蓋主要高風險情境（非媒體工具／`execute_command` 並存／無媒體／直呼工具的單元測試／fold 順序）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用
- [x] **已明確界定為純結構重構**：config 不變、UX 不變、行為 byte-identical

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 in-band，P1；US2 behaviour-identity，P1）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-001 grep 可驗、SC-002 現有測試全綠且斷言不變、SC-003 gates）
- [x] 邊界與拒絕條件已明列（I-1 byte-identical、I-2 無 config/UX 變更、I-3 gate/timeout 不變）
- [x] **無 `NEEDS CLARIFICATION`**（目標明確；shape 為技術決策 → research，若改變正式驗收才升級 clarify）

## 範圍與邊界

- [x] 明確列出 out-of-scope（新媒體能力／video/documents／inline ceilings 序列化／`ToolSetSpec`）
- [x] 已標示對既有 truth 的 MODIFY 意圖（*Image filesystem tool* row + 若 shape (a) 的 tool-contract row；新 ADR 遞交 RF-062-10 + RF-069-1）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- **Lineage 收束**：本輪**完成** `ADR 0032 §Forward RF-062-10` 的 media-channel 半部（seam 半部已由 round 069 / ADR 0039 交付），並**退役** `ADR 0039 §Forward RF-069-1`。此後 RF-062-10 不再有殘餘半部。
- **Anti-muse 承諾**：未採用的 shape 以 **settled rejection** 記入新 ADR，**不得**留成 live forward item（這正是 RF-062-10 變成多輪線的原因）。
- **No-halving**（I-6）：媒體通道**一次**落地；若必須切分，屬 **STOP-and-re-decide**，不得產生 "part 2"。
- **用語校正**：這不是 Go `string` 的限制（image bytes 可以塞進 string），而是 **contract/type-honesty** 限制——結果字串的所有文字型消費者都會破壞它，且對話模型需要具型的 media part。

## 結論

- **Ready for `/axb-clarify` (not escalated) → `/axb-spec-by-example` (NOOP) → `/axb-technical-research`**（shape 決策 + 新 ADR + `techstack.md` MODIFY）。`/axb-dsl-refine` 預期 **NOOP**（無使用者可見變更 — 042/043/047/049/069 結構輪前例）。
