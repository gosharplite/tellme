# 規格品質檢查清單：Preserve completed tool steps on an interrupted turn (round 079)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/079-interrupted-turn-partial-persistence`

**Spec 路徑**: `specs/plans/079-interrupted-turn-partial-persistence/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（中斷 → 有步驟則合成收尾寫入／無步驟則不寫 → 續接）
- [x] 沒有把實作技術、框架或程式細節寫成需求（線路細節置於「Grounded」與研究決策，非需求本體）
- [x] 邊界情況已涵蓋主要高風險情境（工具執行中取消、零步驟、append 失敗、重複寫入、round-078 重試互動、MCP 步驟、`--new`/`-i`、usage 計帳）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 保留並續接 P1 > US2 零步驟乾淨中止 P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 單一接縫、FR-007 schema 不變、FR-008 計數語意、NFR-003 stdout 不變、NFR-004 詞彙凍結）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`
- [x] 本輪 clarify 題數為 **0**（不高於 1 至 3 題上限）
- [x] 低風險未定細節已用假設（A2）揭露，並指向 `/axb-technical-research`
- [x] 仍保留的未定項目皆為**研究層級**（S-5…S-10、S-12），**不阻塞**後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（恰好一筆 entry、steps/calls 相等、續接以 assistant 收尾且角色交替合法、零步驟不寫）
- [x] 成功標準可量測、可驗證且技術中立（entry 數、steps 相等、角色序列、`stdout`／詞彙不變）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **本輪的核心風險已具名：角色交替（I-1）。** `BuildMessages` 直接把 `history.jsonl` 重播進下一輪 context；一個停在 `tool` 結果的部分回合會造成連續兩個 `user` 角色，Gemini/Vertex 會以 HTTP 400 拒絕。spec 以 **I-1 + FR-003** 明確要求合成 entry MUST 以 assistant 收尾，並以 in-process fake 於**兩家族**驗證。
- **與既有 invariant 的衝突已具名：** `history-append-after-complete` 與 `turn-answer-stored-verbatim`（`docs/domain-model/tellme.modelith.{yaml,md}`）需要 **MODIFY**（合成答案例外）。spec 以 **S-10 / A4** 指向研究與 truth-owner，並要求以 ADR 收斂，而非默默矛盾。
- **結束碼的張力已具名（S-7）：** `runTurn` 今日把 `context.Canceled` 對映到 `the provider request failed` + exit 6，但 `run` 的互動讀取器註解聲稱中斷的提示回合有「runTurn convention」為 exit 0 —— 兩者互相矛盾。spec 把成功部分保存的結束碼與診斷列列為研究決策，並要求 `TestExitCodesMatchPinnedContract` 保持一致。
- **觸發條件的精確語意已揭露（S-12）：** 被取消的工具仍會被迴圈 `append` 成一個 step（其 `result` 為錯誤文字），因此 `len(steps) > 0` 可能包含一個「被殺掉」的步驟；是否應要求一個**成功**步驟為研究決策。
- **範圍紀律：** 限縮在 provider／工具迴圈回合；MCP 步驟同構；`-b`/`--retry` 與離線讀取器不在本輪（S-11）。

## Ready 判定

- [x] 已可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）
- [ ] 仍需先補高影響需求缺口

**備註**: clarify **未升級（0 題）** —— 核心行為由 issue #159 鎖定、由操作者指派開輪；issue 的 *Open Design Decisions* 由**操作者明確交由 `/axb-technical-research`**。可直接進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
