# 規格品質檢查清單：The `ToolSetSpec` capability seam (round 069)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/069-toolset-spec-capability-seam`

**Spec 路徑**: `specs/plans/069-toolset-spec-capability-seam/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（以一個具名 capability value 取代 registry 建構的 positional scalars）
- [x] 沒有把實作技術、框架或程式細節寫成需求（型別落點 S-4／carried 值 S-5 留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（離線 `--tool-usage` 兩變體／nil sink／non-vision／不同 family ceiling）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用
- [x] **已明確界定純重構**：config schema 不變、UX 不變、行為 byte-identical

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 seam，P1；US2 behaviour-identity，P1）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-001 grep 可驗、SC-002 現有測試全綠且斷言不變、SC-003 gates）
- [x] 邊界與拒絕條件已明列（I-1 無 config 變更、I-2 無 UX 變更、I-4 範圍僅 seam）
- [x] **無 `NEEDS CLARIFICATION`**（目標明確，未升級）

## 範圍與邊界

- [x] 明確列出 out-of-scope（RF-062-10 的 media-channel 半部／任何新 capability／tool-call concurrency 已 settled off）
- [x] 已標示對既有 truth 的 MODIFY 意圖（*Agent tool loop*／composition-root seam row；新 ADR 遞交 RF-062-10/RF-063-6）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- **Round-number 說明**：`069` 原指派給已**撤回**的 `069-concurrent-tool-dispatch`（未落地、分支已刪、#139 closed `not_planned`）。依命名規則（既有最大編號 +1；`dev` 上最大為 `068`），`069` 為正確且可用的下一號，本 package 沿用之。
- **Slice 風險**：`RF-062-10` 實為兩項（seam + media-channel refactor）。本輪**只做 seam**；媒體通道重構留在 §Forward（避免 scope creep）。

## 結論

- **Ready for `/axb-technical-research`**（clarify 未升級，0 問題）。`/axb-spec-by-example` 預期 **NOOP**（無使用者可見行為變更 — 042/043/047/049 結構輪前例）。
