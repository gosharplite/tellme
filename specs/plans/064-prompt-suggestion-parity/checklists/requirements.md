# 規格品質檢查清單：Prompt suggestion parity — the reference's newest-50 candidate pool (round 064)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/064-prompt-suggestion-parity`

**Spec 路徑**: `specs/plans/064-prompt-suggestion-parity/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（`-i` 建議清單的**候選池深度**：10 → 最新 50 個去重提示詞）
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；pre-load vs per-query、常數命名等留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（恰 10 筆、<50 筆、深度 >50 不進池、空／缺／不可讀 log、跨視窗去重、無命中、>3 行丟棄、長檔界線）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 候選池加深至 50 — P1；US2 其餘來源行為對齊 — P2，**條件式**）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-005；US2: FR-006…FR-008，條件式；**Q1 → A 後 US2／FR-006…FR-008 移除**）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-5 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify` — **Q1 已答（A，單純加深候選池）**：僅把 recent-prompt 候選池加深至最新 **50**，輸出上限維持 10；空輸入筆數與 session 最後提示詞來源**不納入**。
- [x] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題 — **CLOSED：Q1 → A**（一題一問）
- [x] 低風險未定細節已用假設揭露（A1–A6；深度值 50 為可單點調整的常數）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，Q1 已拍板，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（US1 以「最新 10 之外的提示詞被提出」為紅→綠見證）
- [x] 成功標準可量測、可驗證且技術中立（SC-001/002 具可重現紅→綠 witness；SC-003 為 ≤10 與 subsequence 不動的 pin）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **需求本體**：operator 實測 `-i` 輸入 `commit` 僅得 **3** 筆建議 → 追因：tellme 的候選池為最新 **10** 個去重提示詞；reference 為最新 **50**（`LoadTopN(ctx, 50)`），同為 subsequence、同上限 10。以真實 `~/.tellme/global_prompts.jsonl`（471 筆 / 259 去重；43 筆含 "commit"）模擬：池 10 → 3 筆命中；池 50 → 10 筆（達上限）。
- **差距定位**：差距在**池深度**，非匹配規則；operator 明示「behaviour only」且要「看到 10 而非 3」。
- **對齊面 vs 不對齊面**：匹配規則、dedup、上限 10、debounce、>3 行丟棄、無預選、路徑類查詢才查 workspace、tool 名稱來源長 — 皆已對齊或不動；**不對齊且刻意不移植**者記為 forward item（`WorkspacePolicy`、≈150 KiB compaction、log 位置）。
- **高影響未定**：Q1 — 「behaviour only」是否含 reference 的空輸入前 5 與 session 最後提示詞來源（後者會打破 round-016 的「啟動不讀磁碟」性質）。留待 operator 一題拍板。
- **治理**：預期不新增 capability/config；`/axb-api-plan` NOOP；`/axb-data-plan` 預期 NOOP；若 Q1 → B，需記錄 round-016 D3 的 supersession。是否新增 ADR 視研究而定。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: **Clarify CLOSED（1 題，一題一問）— Q1 → A**：僅加深候選池（10 → 最新 50），輸出上限不變 10；空輸入筆數與 session 來源**不納入**（列為 forward item），US2／FR-006…FR-008 移除。**無殘留 `NEEDS CLARIFICATION`**，可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）。
