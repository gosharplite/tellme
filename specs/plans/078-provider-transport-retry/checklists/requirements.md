# 規格品質檢查清單：Retry a transient provider transport failure before failing (round 078)

**建立日期**: 2026-09-22

**Feature Directory**: `specs/plans/078-provider-transport-retry`

**Spec 路徑**: `specs/plans/078-provider-transport-retry/spec.md`

## 使用方式

- 依目前 `spec.md` 的內容逐項檢查。
- 若項目未通過，請在「問題與修正紀錄」補充具體落差與修正方向。
- 若仍保留 `NEEDS CLARIFICATION`，請明確說明它是否阻塞後續規劃。

## 內容完整性

- [x] 已完成所有必填章節
- [x] 功能主題、範圍與主要流程已表達清楚
- [x] 沒有把實作技術、框架或程式細節寫成需求
- [x] 邊界情況已涵蓋主要高風險情境（取消、部分成功、429 vs 4xx、截斷守衛、上限邊界）
- [x] 關鍵實體與成功標準已補齊

## 使用者故事與需求歸戶

- [x] 使用者故事依商業價值與交付順序排序（US1 自動重試 P1 > US2 不重試非可重試類別 P2）
- [x] 每個使用者故事都可被獨立驗證
- [x] 每個使用者故事都包含驗收情境
- [x] 可歸屬單一故事的 FR / NFR 已直接掛在故事底下
- [x] 全域需求只保留跨故事或無法合理歸戶的條目（FR-006 單一接縫、FR-007 計數語意、NFR-003 型別化、NFR-004 控制碼安全）
- [x] 正式需求沒有在故事區與全域需求區重複列出

## 缺口與澄清策略

- [x] 只有高影響缺口才升級到 `/axb-clarify`
- [x] 本輪 clarify 題數為 **0**（不高於 1 至 3 題上限）
- [x] 低風險未定細節已用假設（A2）揭露，並指向 `/axb-technical-research`
- [x] 仍保留的未定項目皆為**研究層級**（機制／接縫／型別化分類／測試接縫／記錄形狀），**不阻塞**後續規劃

## 可驗證性與成功標準

- [x] 驗收情境足以驗證主要成功路徑（恰好 2／3 次呼叫、片語 + exit 6、非可重試零重試）
- [x] 成功標準可量測、可驗證且技術中立（呼叫次數、exit code、位元組精確）
- [x] 假設只表達前提與邊界，沒有偷渡新需求
- [x] 需求、邊界情況、關鍵實體與成功標準彼此一致

## 問題與修正紀錄

- **本輪的核心風險已具名**：`llm.ProviderError` 目前不攜帶 HTTP 狀態，狀態只存在於格式化字串中 —— spec 以 **NFR-003** 明確禁止字串比對，並把「型別化分類」列為研究決策（S-6），避免實作以脆弱的方式落地。
- **與既有 truth 的衝突已具名**：`specs/truth/techstack.md:102`（round-030 列）斷言「tellme adds **no** retry layer」，本輪使其失效 —— spec 要求在該列 **MODIFY**（而非只附加），並以 ADR 0050（或修正 round-030 記錄）收斂。
- **計數語意已錨定在既有契約上**：round-034 的「an internal retry does not count」措辭即為 FR-007 的掛勾，避免與 turn 計數器產生新矛盾。
- **範圍紀律**：MCP 與 `-b`/`--retry` 明確排除（S-9），避免與既有的可回復摺回機制重疊。

## Ready 判定

- [x] 已可進入後續規劃（`/axb-spec-by-example` + `/axb-technical-research`）
- [ ] 仍需先補高影響需求缺口

**備註**: clarify **未升級（0 題）** —— 主題與排程為操作者給定，可重試性判定為操作者在 session 中鎖定；其餘為研究層級決策。可直接進入 `/axb-spec-by-example` 與 `/axb-technical-research`。
