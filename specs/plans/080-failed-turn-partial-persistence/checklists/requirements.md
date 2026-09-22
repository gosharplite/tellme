# 規格品質檢查清單：Persist a failed turn's completed tool steps (round 080)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/080-failed-turn-partial-persistence`

**Spec 路徑**: `specs/plans/080-failed-turn-partial-persistence/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（失敗 → 有步驟則合成收尾寫入、失敗面不變；無步驟則不寫）
- [x] 沒有把實作技術、框架或程式細節寫成需求（線路細節置於「Grounded」與研究決策）
- [x] 邊界情況已涵蓋主要高風險情境（ErrIncomplete/exit 7、重試等待期間的 Ctrl+C hole #2、append 失敗、重複寫入、`--new`/`-i`、usage）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 失敗保留 P1 > US2 零步驟乾淨中止 P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 單一接縫、FR-007 schema 不變、NFR-003 詞彙/結束碼凍結、NFR-004 結構化/型別化判定）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`
- [x] 本輪 clarify 題數為 **0**（不高於 1 至 3 題上限）
- [x] 低風險未定細節已用假設（A2）揭露，並指向 `/axb-technical-research`
- [x] 仍保留的未定項目皆為**研究層級**（S-1…S-7），**不阻塞**後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（恰好一筆、steps/calls、合成 close、片語 + 結束碼不變、續接以 assistant 收尾、零步驟不寫）
- [x] 成功標準可量測、可驗證且技術中立（entry 數、steps 相等、角色序列、結束碼、片語、`stdout` 不變）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **與 round 079 的關係已具名**：本輪重用 round 079 的持久化形狀（`persistInterruptedTurn`、合成 close、`history.Entry` schema 不變），只是把觸發條件從「僅操作者中斷」擴為「失敗且有完成步驟」；**失敗面（片語 + 結束碼）保持不變**是本輪最重要的風險邊界。
- **issue hole #2 已具名**：round-078 重試裝飾器在其放棄路徑回傳 `lastErr`（前一嘗試的 transport 錯誤，未包裹 `context.Canceled`），故 round 079 的 `errors.Is(err, context.Canceled)` 測不到「重試等待期間的 Ctrl+C」。spec 以 **S-2** 提出改用 `ctx.Err() != nil`（嚴格更廣）一併修補，並要求以 Gherkin 見證。
- **範圍（S-1）A vs B 為**操作者可否決的決策**：spec 以 **proposed (B)**（任一失敗回合皆適用，含 exit 7）為預設並標為假設 A2，避免默默落地。
- **`docs/domain-model` 例外延伸已具名**：round 079 修訂了 `history-append-after-complete` / `turn-answer-stored-verbatim`；本輪把同一例外延伸到**失敗**回合（same-PR，ADR 0041）。
- **範圍紀律**：MCP 步驟同構；`-b`/`--retry` 與離線讀取器不在本輪（S-8）。

## Ready 判定

- [x] 已可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）
- [ ] 仍需先補高影響需求缺口

**備註**: clarify **未升級（0 題）** —— 核心行為由 issue #161 鎖定、由操作者指示開輪；issue 的 *Open Design Decisions* 由**操作者指示交由 `/axb-technical-research`**。可直接進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
