# 規格品質檢查清單：Gemini/Vertex parallel tool calls — a round's tool results must share one turn (round 065)

**建立日期**: 2026-09-20

**Feature Directory**: `specs/plans/065-gemini-parallel-tool-calls`

**Spec 路徑**: `specs/plans/065-gemini-parallel-tool-calls/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（Gemini/Vertex 一輪多個 tool call 的結果序列化：N 個 `functionResponse` 必須共用同一個 `user` turn）
- [x] 沒有把實作技術、框架或程式細節寫成需求（引用檔案位置／現況形狀僅為可追蹤性與重現依據；batch 落點、media 併入或分開、loop 順序是否更動等留待 `/axb-technical-research`）
- [x] 邊界情況已涵蓋主要高風險情境（N=1 保真、N=0 文字路徑保真、部分/全部帶 media、tool error／reason-less 拒絕、同一 turn 兩輪、replay/resume、缺 `ToolCallID` 的 FIFO 名稱配對、OpenAI-compatible 不受影響）
- [x] 關鍵實體與成功標準已補齊，或已明確說明為何不適用

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 多 tool call 的一輪能完成 — P1；US2 多圖仍送達模型 — P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR 已直接掛在故事底下（US1: FR-001…FR-008；US2: FR-009）
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（無獨立全域條目；I-1…I-6 為不變式）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify` — **本輪不升級**：目標明確（關閉 [#132](https://github.com/gosharplite/tellme/issues/132)）、缺陷已重現且修法已在 issue 內載明；殘餘選擇為技術性（落點與 media 併法），依 round-063 前例交由 `/axb-technical-research`
- [x] 本輪 clarify 題數控制在 1 至 3 題／session 上限 5 題 — **未使用**（0 題）
- [x] 低風險未定細節已用假設揭露（A1–A6；S-2/S-3/S-4 明確標為 proposed／research）
- [x] 仍保留的 `NEEDS CLARIFICATION` 已標示是否阻塞後續規劃 — **無殘留缺口**，可進入後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（US1 以「兩次 tool call 的一輪不再 400」為紅→綠見證；US2 以「模型描述兩張圖」為結果層見證）
- [x] 成功標準可量測、可驗證且技術中立（SC-001/002 具可重現紅→綠 witness；SC-003/004 為 byte-identical 回歸 pin）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **需求本體**：round-063 closeout **live check**（2026-09-20）以 `coder` peer、`dev` provider（`gemini-3.8-flash`／Vertex）實測：**2 × `read_files`**（無圖）與 **2 × `read_image`** 皆在下一輪 Vertex 請求得到 **400**（`function response parts` 數 ≠ `function call parts` 數）；**單張 `read_image`** 則 `exit 0` 且模型描述出圖內容。缺陷為 **media-agnostic**。
- **根因**：loop **每次呼叫各 append 一個 `tool` message**（並於 media 呼叫後各 append 一個 `user` media message），Gemini adapter `buildContents` **逐訊息**映射為各帶單一 `functionResponse` 的 **多個 `user` turn**；Vertex 要求同一 function-call turn 的 responses 併於 **同一 turn**。
- **來源判定（pre-existing）**：per-call `tool` message = round 008（`5a37fe4`）；per-message `functionResponse` `user` turn = round 013（`de79fc0`）；round 062 加 per-call media message、round 063（`e8e0880`）加 `inlineData` 分支。**非 round 063 引進**；OpenAI-compatible 家族使用獨立 `role:"tool"` messages，不受影響。
- **修法（issue 已載明）**：把一輪的 tool 結果批次成 **單一 `user` turn**（所有 `functionResponse` 相鄰、依呼叫序），media turns 置後（round-scoped placement，即 ADR 0033 RF-063-7）。
- **既有 pin 需更新**：`internal/infrastructure/llm/gemini/client_image_test.go` 的 `TestRequestBody_MultiCallRound_MediaTurnsInterleave` 目前 **pin 住 interleave 形狀**，本輪須改為新形狀（S-5）。
- **治理**：預期 **新增 ADR**（落點決策，supersede ADR 0033 RF-063-7、標註其 D2 範圍註記）；`/axb-api-plan` NOOP；`/axb-data-plan` NOOP；`/axb-dsl-refine` 視研究決定（新增一則 multi-call Rule/Example 或 NOOP）。預期不新增 capability/config/dependency。

## Ready 判定

- [x] 已可進入後續規劃
- [ ] 仍需先補高影響需求缺口

**備註**: **Clarify 未升級（0 題）** — 目標明確（關閉 [#132](https://github.com/gosharplite/tellme/issues/132)），缺陷已重現、修法已載明；殘餘為技術選擇（S-2/S-3/S-4），交由 `/axb-technical-research`。**無殘留 `NEEDS CLARIFICATION`**，可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）。
