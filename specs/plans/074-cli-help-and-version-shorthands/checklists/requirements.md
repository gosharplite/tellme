# 規格品質檢查清單：`-h` help and `-v` version shorthands (round 074)

**建立日期**: 2026-09-21

**Feature Directory**: `specs/plans/074-cli-help-and-version-shorthands`

**Spec 路徑**: `specs/plans/074-cli-help-and-version-shorthands/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（`-h` 印出 flag 用法並成功結束；`-v` 等同 `--version`）
- [x] 沒有把實作技術、框架或程式細節寫成需求（help 內容/diispatch 優先序 → research）
- [x] 邊界情況已涵蓋主要高風險情境（`-h` 與 prompt/`-c` 併用、`-h` 與 `-v` 同時、pipe）
- [x] 關鍵實體（action flag）與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 `-h`，P1；US2 `-v`，P1；US3 usage-error 不回歸，P2）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-001 E2E 載體；SC-002 既有 usage-error/`--version` 場景保留不變）
- [x] 邊界與不變量明列（I-1 usage-error 不回歸、I-2 offline/prompt-less、I-3 片語詞彙不變、I-4 `--version` 不變）
- [ ] **無 `NEEDS CLARIFICATION`** — 待 `/axb-clarify` 收斂 **Q1（`-h` 是否附 `--help` 長形式）** 與 **Q2（help 走 stdout+0 或 stderr+專用碼）**

## 範圍與邊界

- [x] 明確列出 out-of-scope（不做子命令、不做 cobra、不擴充片語詞彙、不改 `--version` 行為）
- [x] 已標示對既有 truth 的 MODIFY 意圖（`specs/truth/features/cli/usage/**` 的 flag-parsing 列；`techstack.md` 的 CLI-flags 列；必要時 `diagnostics/version-and-setup-diagnostic.feature`）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- Q1/Q2 依 `/axb-clarify`（一次一題）收斂；A1–A4 為已揭露、可由 operator 否決的假設。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口（Q1–Q2）

**備註**: spec 骨架與既有系統對照已完成；待 clarify 收斂後進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
