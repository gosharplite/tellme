# 規格品質檢查清單：`search_files` — a bounded, deterministic in-file search tool (round 071)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/071-search-files-tool`

**Spec 路徑**: `specs/plans/071-search-files-tool/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（新增第九個 tool `search_files`：在目錄子樹內做 **bounded、deterministic** 的檔案內容搜尋 — reader trio 缺的那一半）
- [ ] 沒有把實作技術、框架或程式細節寫成需求（搜尋模式／ignore policy／bound 機制的 **正式驗收面** 已明確升級為 clarify Q1–Q3；其餘機制留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（0 matches／invalid regex／binary／unreadable／超大檔與超長行／path 為檔案／空 query／timeout）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 找到內容，P1；US2 bounded，P1；US3 模式明確，P1 — 由 Q1 定案）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-001 assembler／report／schema gate；SC-002 E2E；SC-003 gates；SC-004 bash 對照的 boundedness witness）
- [x] 邊界與拒絕條件已明列（I-1 reason gate、I-2 resource contract、I-3 no-security-layer、I-4 deterministic、I-8 無 config/UX 變更）
- [ ] **`NEEDS CLARIFICATION` 未收斂**：Q1 搜尋模式、Q2 scope/ignore、Q3 bound/shape — **阻塞 `/axb-spec-by-example`**（工具 schema 與輸出即驗收契約）

## 範圍與邊界

- [x] 明確列出 out-of-scope（Go/AST 24 tools；`find_file` 檔名搜尋；enterprise/MCP 候選）
- [x] 已標示對既有 truth 的 ADD/MODIFY 意圖（`techstack.md` reader row；`specs/truth/features/cli/chat/**` + `dsl.md`；`Tool` 實體）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- **與 reference 的三個刻意分歧**：`(1)` **不安裝 `SafePath`/consent 閘**（no-security-layer 為既定排除）；`(2)` **不採用 reference 的 `defaultWorkspacePolicy` ignore 清單**（那是 secret-scanning 關注；tellme 無此 policy → 改由 Q2 決定本工具自身 scope）；`(3)` **輸出必須 deterministic**（reference 是 worker/channel 順序 — 工具的輸出是 executable truth，須可重現）。三者在 spec §「Read first」與 §Design 明列。
- **Bar 對照（README design intent）**：工具必須在真實軸上勝過 bash — 本工具勝在 **context-boundedness**（`grep -rn` 無界）與 **determinism/testable contract**，且補上 reader trio 的缺口。
- **Scope 護欄**：`search_files` 是 #147 唯一的 Filesystem 候選；Go/AST 24 tools 與 `find_file`（excluded, #146）明確排除，避免本輪膨脹。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口 → `/axb-clarify`（Q1–Q3，一次一題）

**備註**: `/axb-clarify` 收斂並回折 spec 後，方可進 `/axb-spec-by-example`（預期 **ADD**：新增一個 CLI journey，非 NOOP — 本輪有使用者可見（模型可見）的新行為）與 `/axb-technical-research`（工具實作決策 + `techstack.md` MODIFY + 新 ADR）。
