# 規格品質檢查清單：Gemini/Vertex usage drops the cached-token count (round 072)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/072-gemini-cached-token-usage`

**Spec 路徑**: `specs/plans/072-gemini-cached-token-usage/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（Gemini/Vertex adapter 解出 `cachedContentTokenCount`/`thoughtsTokenCount`，修正 `H: 0` 與 ~10× 溢價）
- [x] 沒有把實作技術、框架或程式細節寫成需求（殘餘的 disjointness 規則 S-4 為技術決策 → research）
- [x] 邊界情況已涵蓋主要高風險情境（無 `usageMetadata`／欄位為 0／`cached == prompt`／`thoughtsTokenCount` 缺漏／負值／工具輪）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 真實 cache hits，P1；US2 thinking tokens，P1；US3 一致性的會計表面，P2）
- [x] 每個 FR / NFR 歸戶到對應的使用者故事或全域需求
- [x] 每個使用者故事至少有一個可獨立驗證的驗收情境

## 需求明確性與可驗證性

- [x] FR 使用 MUST 等可驗證語言
- [x] 成功標準可量測（SC-001 unit pin + E2E witness；SC-002 對照 reference；SC-003 gates；SC-004 溢價 witness）
- [x] 邊界與拒絕條件已明列（I-1 byte-identity、I-2 公式不動、I-3 退化安全、I-5 四表面一致）
- [x] **無 `NEEDS CLARIFICATION`**（缺陷與修法明確；S-4 為技術決策 → research，若改變 operator 可見數字才升級）

## 範圍與邊界

- [x] 明確列出 out-of-scope（新的 pricing/accounting 功能；OpenAI-compatible adapter 已正確；provider 端 cache 管理）
- [x] 已標示對既有 truth 的 MODIFY 意圖（`specs/truth/techstack.md` 的 usage / `UsageRecord` 會計 row；Gemini adapter row）
- [x] 未改寫任何既有 plan package；未寫入 `specs/truth/**`

## 問題與修正紀錄

- **Defect 錨定**：issue [#149](https://github.com/gosharplite/tellme/issues/149)（本輪 DoD 即關閉它）。實測 `dev` @ `5b1bfa1`：`gemini/client.go:532-535` 只解 3 個 `usageMetadata` 欄位，`:574-580` 未設 `CachedTokens`/`ThinkingTokens` ⇒ `miss = prompt − 0`。
- **Reference 對照**：`tell-me-go/internal/infrastructure/llm/gemini/metrics.go:17-28`（`CachedContentTokenCount` + `ThoughtsTokenCount`）。
- **In-tellme 一致性論證**：OpenAI-compatible adapter 已正確解析（`openai/client.go:278-281,299-309`），故本輪是讓兩 adapter 一致，非新增能力。
- **溢價量化**：`gemini-3.8-flash` `HIT 0.075 / MISS 0.75 / COMP 3.75`，MISS=10×HIT；operator 終端 `M: 390564 H: 0 C: 100 ⇒ $0.2933`。

## Ready 判定

- [x] 已可進入後續規劃（clarify 未升級）
- [ ] 仍需先補高影響需求缺口

**備註**: `/axb-spec-by-example`（**ADD** — 使用者可見的 metrics 修正）→ `/axb-technical-research`（S-4 決策 + 新 ADR + `techstack.md` MODIFY）→ `/axb-system-analysis` → `/axb-dsl-refine`（ADD/MODIFY 一個 CLI journey）→ `/axb-tasks` → `/axb-implement`。
