# 規格品質檢查清單：Gemini image vision — `inline_data` on the Gemini/Vertex wire (round 063)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/063-gemini-image-vision`

**Spec 路徑**: `specs/plans/063-gemini-image-vision/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；wire 形狀的**放置位置**留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（上限邊界雙向、非圖片內容、0-byte、`VISION` off、functionCall/functionResponse 配對、無 reason、base64 膨脹、`thoughtSignature`、宣告錯誤、history replay）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 Gemini 讀圖 — P1；US2 單一能力鍵、雙家族對等 — P2；US3 防回歸／文字路徑不變 — P3）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-006；US2: FR-007…FR-009；US3: FR-010…FR-012）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-5 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [ ] 只有高影響缺口才升級到 `/axb-clarify` — **Q1**（能力鍵是否沿用 `VISION`）、**Q2**（Gemini inline 尺寸上限）已升級；**Q3**（wire 放置位置）刻意留給 `/axb-technical-research`（不改變驗收標準）
- [ ] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題 — **OPEN：Q1、Q2 尚未拍板，一題一問進行中**
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A6）
- [ ] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **Q1、Q2 為阻塞項**（決定 config 介面、提供面語意與一個編號驗收標準）；收斂前不進入 handoff

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（SC-002 有可重現的紅→綠 witness；SC-003 為 byte-identity pin）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **需求本體**：operator 問「reference 的 vertex gemini 有 vision 嗎」→ 經實查，**有**（Gemini adapter 以 `inline_data` blob 原生帶圖，且**不受** `SupportsVision` 旗標限制；該旗標只 gate OpenAI transport），接著要求「讓 tellme 的 gemini 也能 read_image」。
- **本輪即 `RF-062-1`**：round 062 / ADR 0032 D4 刻意留下家族不對稱（Gemini 大聲拒絕），並在 §Forward 記錄 `RF-062-1` 為 forward item。本輪落地該項。
- **tellme 現況（實測）**：`llm.Message.Media` 與 per-call media channel 已存在（round 062 D5/D7a）；`read_image` 已存在且**已**由 `VISION` gate（提供面不需改）；**唯一缺的是 Gemini adapter 的 `inline_data` 序列化**，以及隨之要 retire 的大聲拒絕。
- **關鍵風險**：Gemini parser 對 `functionCall` 未被 `functionResponse` 回應會整包 400；reference 以 `normalizeUserTurnParts` 將 user turn 重排為 `[InlineData][FunctionResponse][other]`（#1441）。放置位置為本輪最高風險，列為 A2／Q3 交給 research。
- **治理**：預期新增 **ADR（延伸 ADR 0032、supersede `RF-062-1`）**；確切編號於 research 決定（S-8）。

## Ready 判定

- [ ] 已可進入後續規劃
- [x] 仍需先補高影響需求缺口 — **Q1（能力鍵）、Q2（Gemini inline 上限）** 尚未拍板

**備註**: spec 已完整記錄需求來源、現況實測與不變式；US/FR/SC 骨架已就位，等待 Q1／Q2 拍板後即可進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
