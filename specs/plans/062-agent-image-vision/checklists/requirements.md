# 規格品質檢查清單：agent image vision — `read_image` to the provider wire (round 062)

**建立日期**: 2026-09-19

**Feature Directory**: `specs/plans/062-agent-image-vision`

**Spec 路徑**: `specs/plans/062-agent-image-vision/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；wire 形狀與放置位置留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（32 MiB 邊界雙向、非圖片內容、0-byte、`VISION` off 的落空呼叫、media-first 排序、無 reason、base64 膨脹、宣告錯誤、history replay）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 讀圖 — P1；US2 能力閘門 — P2；US3 防回歸／文字路徑不變 — P3）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-005；US2: FR-006…FR-009；US3: FR-010…FR-012）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-5 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`（Q1 家族範圍、Q2 能力判定、Q3 工具範圍、Q4 提供面、Q5 尺寸上限與超限行為 — 皆改變範圍邊界或驗收標準）
- [x] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題（**5 題，一題一問**）— **CLOSED：Q1 → 1**（僅 OpenAI-compatible 家族）· **Q2 → 1**（每 provider 顯式 `VISION` 設定鍵，預設關）· **Q3 → 1**（僅 `read_image`）· **Q4 → 1**（`VISION` off 時不提供 `read_image`）· **Q5 → 1**（inline 上限 32 MiB；超限為大聲 tool error）
- [x] 低風險未定細節已用 `NEEDS CLARIFICATION` 或假設揭露（A1–A7）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，Q1–Q5 皆已拍板，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑
- [x] 成功標準可量測、可驗證且技術中立（SC-002 有可重現的紅→綠 witness；SC-003 為 byte-identity pin）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **需求本體**：operator 提供官方 DeepSeek 文件 — `deepseek-flash` **已支援影像**（JPEG/PNG/GIF/WebP，格式由**檔案內容**判定），且 `deepseek-v4-flash-vision-exp` 已退役（仍接受，由最新 Flash 服務）。這使 `tell-me-go` 以 model-id 子字串（`vision`）判定能力的做法**過時**（其 ADR-070 假設 `deepseek-v4-flash` 為純文字）。
- **tellme 現況（實測）**：端到端皆為純文字 — `llm.Message.Content string`、兩個 transport 只送文字、`tools.Tool` 回傳 `string`、無 media tool、無 capability 設定鍵。因此目前**無任何路徑**可讓模型看到圖片。
- **家族差異**：OpenAI-compatible（本輪範圍）以 inline base64 `image_url` block；Gemini 以 `inline_data`（本輪 out of scope，forward item）。
- **參考實作**：`tell-me-go` 有 `read_image`（回傳 `BinaryData`）＋ `llm.Part{InlineData}`，並由 `executor.AssembleResponse` 以 **media-first** 摺回、Gemini 用 `role:user`。DeepSeek 文件則稱「images in user messages only」— 放置位置的取捨留待 `/axb-technical-research`（A2）。
- **Q4 的 truth 影響**：`VISION` off 時不提供 `read_image`，使「提供面」成為能力的函數 — 須同步 MODIFY round-029/033 的 offered-set truth（FR-009）。
- **治理**：新增 **ADR 0032**（能力鍵、範圍、Forward）；不修訂既有 ADR（S-8）。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: spec 已完整記錄需求來源、現況實測與不變式。**Clarify CLOSED（5 題，一題一問）**：家族範圍（僅 OpenAI-compatible）· 能力判定（顯式 `VISION` 設定鍵）· 工具範圍（僅 `read_image`）· 提供面（能力決定）· 尺寸上限（inline 32 MiB、超限大聲錯誤）。**無殘留 `NEEDS CLARIFICATION`**，可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）。
