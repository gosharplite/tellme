# 規格品質檢查清單：Roll back the last N turns of the session history (`-b`/`--back`) (round 081)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/081-back-rollback-turns`

**Spec 路徑**: `specs/plans/081-back-rollback-turns/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚（`-b [N]` 離線回退；`-b [N] "p"` 先回退再執行；一次移除整回合；不寫 archive）
- [x] 沒有把實作技術、框架或程式細節寫成需求（Store 形狀與實作細節置於「Grounded」與研究決策）
- [x] 邊界情況已涵蓋主要高風險情境（N 超量、`N ≤ 0`、空 session、`-b`×`-l`/`--new`/`-t`、`-c missing`、不可解碼 history、檔案不存在）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 離線回退 P1 > US2 回退後送出新提示 P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-008 單一 capability、FR-009 pre-pass 一般化、FR-010 schema 不變、NFR-003 詞彙/結束碼凍結、NFR-004 離線路徑、NFR-005 回合編號）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`
- [x] 本輪 clarify 題數為 **0**（不高於 1 至 3 題上限）
- [x] 低風險未定細節已用假設（A2）揭露，並指向 `/axb-technical-research`
- [x] 仍保留的未定項目皆為**研究層級**（S-1…S-8），**不阻塞**後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（`K → K − N` 行、預設 1、最後 N 個被移除、`-b "p"` 先回退再執行、archive 不變、離線零請求）
- [x] 成功標準可量測、可驗證且技術中立（檔案行數、被移除的是最後 N 個、archive 檔案不變、provider 請求計數、結束碼）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **與 `--new` 的界線已具名**：回退只丟棄**最後 N 個**回合並保留其餘；`--new` 是**封存全部**並開新 session。本輪以 **FR-004 / I-3b** 明確禁止回退寫入 `history.archive.jsonl`，避免兩者語意混淆。
- **比 reference 更簡單的機制已具名**：`tellme` 每回合存**一行**（`history.Entry`，`steps[]` 內嵌），故回退為行截斷，不需要 reference 的「N×2 訊息」算術與奇偶修補，也無拆散 `FunctionCall`/`FunctionResponse` 配對之虞（I-1/I-2）。
- **RF-54-4 觸發已具名**：ADR 0023 曾預告 `-l` 的 adjacent-integer 前置處理應在出現 `-b`/`--back` 類旗標時一般化；本輪即其觸發點（FR-009 / S-7），並要求一併調和已記錄的邊界案例。
- **acceptance-shaping 決議已標為操作者可否決**：`N` 超量（clamp vs refuse，S-2）、`N ≤ 0`（usage error vs no-op，S-3）、`-b`×`--new` 組合（S-4）皆可能改變正式驗收；本輪以 **建議值 + A2** 揭露並交由 `/axb-technical-research` 收斂。
- **凍結面已具名**：無新片語、結束碼集合維持十個（NFR-003）；回報走 `stdout`；失敗沿用環境類片語 + exit 4（I-6）。
- **範圍紀律**：`--retry`、`-e`/`--update-turn`、`browse`、TUI `--back`、互動確認、`--json` 皆不在本輪（S-9 / A6）。

## Ready 判定

- [x] 已可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）
- [ ] 仍需先補高影響需求缺口

**備註**: clarify **未升級（0 題）** —— 核心行為由 issue #163 鎖定、由操作者指示開輪；issue 的 *Open design decisions* 交由 `/axb-technical-research`。可直接進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
