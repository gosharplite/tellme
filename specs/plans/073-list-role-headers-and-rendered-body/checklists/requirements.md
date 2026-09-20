# 規格品質檢查清單：The history listing reads like the reference (round 073)

**建立日期**: 2026-09-21

**Feature Directory**: `specs/plans/073-list-role-headers-and-rendered-body`

**Spec 路徑**: `specs/plans/073-list-role-headers-and-rendered-body/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（`-l` 列表改為 `[USER]`/`[MODEL]` 標頭 + glamour 渲染內文 + 訊息間空行；顏色僅在 stdout 為終端時出現）
- [x] 沒有把實作技術、框架或程式細節寫成需求（renderer 建構、`WRAP_WIDTH` 套用、stdout 終端探針 seam 皆為技術決策 → research）
- [x] 邊界情況已涵蓋主要高風險情境（空內文、控制序列、`WRAP_WIDTH`、pipe、內文尾端換行）
- [x] 關鍵實體（history message 的 role/body）與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 標頭+渲染+空行 P1；US2 顏色與其 gate P1；US3 選擇/計數語意不變 P2）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-002 施加「pipe 下 0 個 `\033`」的機械判準；SC-003 既有的 offline/selection 斷言保留）
- [x] 邊界與不變量已明列（I-1 offline/terminal、I-2 檔案格式凍結、I-3 顏色只走 stdout、I-5 計數語意不變）
- [ ] **無 `NEEDS CLARIFICATION`** — 待 `/axb-clarify` 收斂 **Q1（`-r` 是否生效）／Q2（是否列出工具活動）／Q3（prompt 內文是否也渲染）** 後才可勾選

## 範圍與邊界

- [x] 明確列出 out-of-scope（不採 reference 的 `-l N "prompt"` 可組合行為、不配 `-b`、不改 `history.jsonl` 格式、不加新依賴）
- [x] 已標示對既有 truth 的 MODIFY 意圖（`specs/truth/features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` 的 listing 列；`techstack.md` 的 CLI flags/history 列；視需要 `tellme.modelith.*`）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- Q1–Q3 依 `/axb-clarify`（一次一題）收斂；A1–A5 為已揭露、可由 operator 否決的假設。
- A2（空 session 靜默 vs reference 的 `No history found.`）刻意不問、列為可否決假設。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口（Q1–Q3）

**備註**: spec 骨架與既有系統對照已完成；待 clarify 收斂後即可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
