# Tasks: 035 — spinner tail residue (yield the per-call tail to the progress indicator)

**Plan Package**: `specs/plans/035-spinner-tail-residue`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`, `specs/truth/features/cli/chat/dsl.md`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**) and `specs/truth/data/**` (`/axb-data-plan` **NOOP**) are unchanged this round. There is no `ui/**` (a line-oriented CLI diagnostic).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是一個實作對真相（implementation-vs-truth）的顯示缺陷修復**：違反的規則已存在於 executable truth（round-019/025 spinner residue rows），**沒有新的 DSL 句、沒有新的 stepdef**——`/axb-dsl-refine` 對 feature 的改動只是**新增一個 Example 重用既有句**。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D6；`go.mod`／`go.sum` 不動）。
  - **省略 Phase 2 `Foundational`** —— 只需一個 Phase-3 `[UNIT]` 任務（落點為新檔），無 stepdef 落點骨架、無並行調度需求；不得把落點骨架塞進 Setup／Foundational。
  - **Phase 3 只有 `[UNIT]`** —— 本輪沒有待處理 DSL 句（新 Example 的 7 句全部沿用既有 stepdef）；測試層工作是**新增 ordering unit pin**並讓它 RED。
  - **Phase 4 只有一個 MODIFY Feature phase** —— 新增的 Example 需 `[BDD-GREEN] -> [BDD-REFACTOR]`。
  - 交付物是 **`compositeObserver.OnCallEnd` 的 phase-boundary spinner clear**（`internal/cli/composite_observer.go`）；唯一 truth 變更是 `specs/truth/techstack.md`（spinner row，by `/axb-technical-research`）與 `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` + `chat/dsl.md`（by `/axb-dsl-refine`）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與 `specs/truth/techstack.md` 對應 section。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** flags、exit codes、class-phrase vocabulary、line formats、frame cadence、tool schemas/results、provider transport、`history.jsonl`、`stdout`。

## Round-035 locked decisions (implementation constraints — MUST)

> 來自 operator Q1–Q4、`spec.md` FR-001..FR-007 / SC-001..SC-004、`research.md` D1–D6。

- **[MECHANISM]** 修復 = **phase-boundary yield**：per-call tail 寫入前**同步清除** spinner，且**不立即 resume**（`research.md` D1）。**明確拒絕**「clear + resume after tail」——byte trace 顯示它只把殘留**移位**到下一次 frame 寫入（`FormatTurnOpening` 的前導 `\n` 把 resumed frame 留在上一列）（D1 替代方案）。indicator 由**下一個等待階段的 hook**（`OnInferenceStart`）或回合的 `Stop()` 重新啟用；turn-scoped elapsed epoch 不變。
- **[FIX-SITE]** 修復落在 **`compositeObserver.OnCallEnd`**（`internal/cli/composite_observer.go`）：在委派 `call.OnCallEnd` **之前**呼叫 spinner 的 clear，且**只在 `!final`**（final 的 tail 是 deferred，於 `Stop()` 之後才 emit，故跳過以維持 byte-identical）（`research.md` D2）。`callRenderer`、`internal/ui/spinner.go`、`Spinner.OnCallEnd` 皆**不動**（Q1(a)；`Spinner.OnCallEnd` 維持 no-op）。
- **[SCOPE]** 純 bug fix（Q2）：**不**納入 #69 的 composition-root 重構、`BindToolOutput` ctor 注入或 `LoopObserver` 介面切分。**不**納入 G2 `Ready` overstatement 與 per-call numbering skew（Q3）。
- **[WITNESS]** E2E（primary）+ unit（secondary）（Q4）：E2E = 新增的 Example（gate + scripted tool round + whole-stream residue row，`research.md` D3/D4）；unit = `compositeObserver.OnCallEnd` 的 yield ordering pin（D5）。
- **[RED-FIRST]** 新 Example 與 unit pin 今日必須對缺陷**非真空地失敗**（assertion-only，非 undefined、非 parse error），再於 `[BDD-GREEN]` 修復後轉綠。
- **[NO-DEP]** 只用 Go 標準庫；無新相依；`go.mod`／`go.sum` 不動；POSIX-only。
- **[VERIFY]** hermetic（無 pty）；`make verify` + godog E2E + 拓樸稽核皆須綠。
- **[TRUTH]** 唯一 truth 變更 = `specs/truth/techstack.md`（Turn progress spinner row）+ `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` 與 `chat/dsl.md`。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D6；無新相依，`go.mod`／`go.sum` 不動）。不得把落點骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無 stepdef 落點骨架（新 Example 全用既有句）；修復落點（`internal/cli/composite_observer.go`）與 unit pin 落點（新檔 `internal/cli/composite_observer_test.go`）皆由下方任務直接交付。不得偷做 Phase 3 測試層或 Feature Green。

## Phase 3: Test Alignment (unit)

**Goal**: 先把本輪的驗證落到測試層（**不寫產品碼**）：新增 `compositeObserver.OnCallEnd` 的 yield-ordering unit pin 並讓它對現行缺陷 **RED**；同時確認 `[axb-dsl-refine]` 新增的 E2E Example 以**既有 stepdef** 跑得起來且 RED（0 undefined steps）。

**DSL 參照**:
- 本輪**沒有新增或修改任何 DSL 句**：新增 Example 的 7 句全部沿用既有的同模組/介面根 rows 與其 stepdef；因此 Phase 3 不含 `[BDD-ALIGN]`／`[BDD-REMOVE]`／`[BDD-RED]` 任務。
- `truth-delta.md` 記錄 `/axb-dsl-refine` 為 **MODIFY**（feature + note，非 DSL 句改動）。

**Markers**:
- `[UNIT]`：本輪測試層工作是 **unit-tier**（composite yield ordering）；非 Gherkin stepdef。只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` -> `Rule: The spinner leaves no residue on a tool-using turn`
- `specs/truth/features/cli/chat/dsl.md` -> round-035 note + round-019/025 notes
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY feature + `/axb-technical-research` MODIFY `techstack.md`
- `research.md` -> `Decision 2`（fix 落點）、`Decision 5`（unit pin）

**Boundary**:
- 只動測試層（新檔 `internal/cli/composite_observer_test.go`）；不寫產品碼。
- review 啟動 subagent；本輪的 E2E Example 與 unit pin 必須對現行缺陷 **非真空失敗**（assertion-only）；且本輪所有 Test Scope **不得有 undefined step**。通過前不解鎖 Phase 4。

**Parallel Hint**:
- 只有一個 Phase-3 任務（T001）；T002 等 T001 回來再啟動 subagent 來 review。

- [ ] T001 [UNIT] 新增 `compositeObserver.OnCallEnd` yield-ordering pin（RED）
  - Read:
    - `research.md` -> `Decision 2`, `Decision 5`
    - `internal/cli/composite_observer.go` -> `compositeObserver`（`call` / `spinner` 兩半）與 `OnCallEnd`
    - `internal/domain/agent/observer.go` -> `CallObserver` / `LoopObserver` 契約
    - `internal/ui/spinner.go` -> `BeforeToolLog`（synchronous clear）與 `OnCallEnd`（no-op）
    - `internal/cli/cli.go` -> `loop.Observer = compositeObserver{...}`（composition root）與 `sp.Stop()` / `renderer.EmitFinalTail()`（deferral）
  - 做：新增 `internal/cli/composite_observer_test.go`（新檔，Zero Shared Edits），以 **recording doubles**（記錄呼叫順序的 `call` stub 與 `spinner` stub）斷言：
    - `OnCallEnd(_, _, _, final=false)` → spinner 的 clear（`BeforeToolLog`）在 `call.OnCallEnd` **之前**被呼叫，且 `AfterToolLog`（resume）**未**被呼叫；
    - `OnCallEnd(_, _, _, final=true)` → clear 與 resume **皆未**被呼叫，而 `call.OnCallEnd` 仍被呼叫（deferral 由 renderer 內部處理）。
  - 邊界：此測試 **今日必須失敗**（現行 `OnCallEnd` 沒有 clear）——**不得**放寬 assertion 讓它變綠（RED-first）；不寫產品碼。
  - 不做：不改 `composite_observer.go`（留 Phase 4）；不碰 `callRenderer`／spinner 內部；不寫 E2E stepdef（既有句已存在）。

- [ ] T002 subagent review (phase quality gate)
  - Read: `internal/cli/composite_observer_test.go`、`tests/e2e/steps/spinner_helpers.go`、`specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`、`research.md` -> `Decision 3`, `Decision 5`
  - 檢驗：
    - unit pin **非真空失敗**（現行缺陷下 clear 不存在）且點名 ordering；只動測試層。
    - 本輪新 Example 以**既有 stepdef** 可被跑到（**0 undefined steps**），且對現行缺陷 **RED**：跑 `go test ./tests/e2e/...`（或對應 godog 標籤）確認 `Rule: The spinner leaves no residue on a tool-using turn` 的 Example 因 **residue assertion**（`the run shows no progress spinner` 讀到 `⠋ Executing tools […]...`）而失敗——**非** undefined／setup 錯誤。
    - 有 issues 修正再 review，直到零問題；通過前不解鎖 Phase 4。

> Phase-3 review 由 orchestrator 執行（本 session 無 parallel-subagent substrate）：新增的 unit pin 以現行 `compositeObserver` 跑得起來但**失敗**（無 clear 呼叫）；新 Example 由既有 stepdef 跑到 `the run shows no progress spinner` 並因殘留而**失敗**（assertion-only，0 undefined steps）。Gate PASSED —— Phase 4 解鎖。

## Phase 4B: MODIFY Feature File - cli/chat/presenting-the-progress-spinner.feature

**Goal**: 以最小產品變更（`compositeObserver.OnCallEnd` 的 phase-boundary clear）讓新增的 Example 與 T001 的 unit pin 轉綠，且不改任何顯示格式、cadence 或 gated-off 行為。

**Shared Must Read**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` -> `Feature: Presenting the progress spinner`（新 `Rule: The spinner leaves no residue on a tool-using turn`）
- `specs/truth/features/cli/chat/dsl.md` -> round-035 note（phase-boundary yield）
- `specs/truth/techstack.md` -> Turn progress spinner (operator) row（round-035 clause）
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY feature + `chat/dsl.md`；`/axb-technical-research` MODIFY `techstack.md`
- `research.md` -> `Decision 1`, `Decision 2`, `Decision 6`

**Boundary**:
- 只改 `internal/cli/composite_observer.go`（`OnCallEnd` 於 `!final` 時、委派 `call.OnCallEnd` 前，呼叫 spinner half 的 synchronous clear）。
- **不 resume**（phase-boundary yield；下一個 `OnInferenceStart` 或 `Stop()` 重新啟用）；**不動** `callRenderer.OnCallEnd`／deferral、**不動** `internal/ui/spinner.go`（含 `OnCallEnd` no-op）、**不動** 任何 line format／cadence／class phrase／exit code／`stdout`。
- 不得為了轉綠而放寬 T001 的 assertion。

**Test Scope**:
- `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`

- [ ] T003 [BDD-GREEN] 讓 Test Scope 全綠（套用 phase-boundary clear）
  - Read:
    - `research.md` -> `Decision 1`, `Decision 2`
    - `internal/cli/composite_observer.go`
    - `internal/ui/spinner.go` -> `BeforeToolLog`（clear 語意）／`OnCallEnd`（no-op）
  - 做：在 `compositeObserver.OnCallEnd` 中，當 `final == false` 時，於委派 `c.call.OnCallEnd(...)` **之前**呼叫 spinner half 的 clear（`BeforeToolLog`），且**不**呼叫 resume（`AfterToolLog`）。確保 gated-off（`spinner == nil`）與 `final == true` 皆 no-op（維持既有 byte-exact 行為）。
  - 驗證：T001 unit pin 轉綠；`presenting-the-progress-spinner.feature` 新 Example 轉綠；既有 spinner Examples（含 narrow-terminal residue）與 `the progress spinner no longer appears once the answer is written`（t011）保持綠；`stdout` byte-exact。
  - 不做：不 resume；不改格式/cadence；不碰 `callRenderer`／spinner 內部；不新增相依。

- [ ] T004 [BDD-REFACTOR] 在綠燈下整理 clear 呼叫的落點
  - Read: `internal/cli/composite_observer.go`、`research.md` -> `Decision 1`, `Decision 2`
  - 做：若可提升可讀性，將「yield the indicator before a tail write」抽成 composite 上一個具名小 helper（或加清楚註解指出 phase-boundary no-resume 的理據），保持輸出語意不變、gate 續綠。
  - 不做：不擴大重構範圍、不改 `callRenderer` 契約、不改行為。

## Phase 5: Verification & Regression

**Goal**: 全量回歸 + 可偽性見證 + `make verify` + 拓樸稽核，並確認 flags／exit codes／class phrase／line formats／cadence／`stdout` 皆未變。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Turn progress spinner (operator) row
- `research.md` -> `Decision 3`, `Decision 4`, `Decision 6` + Recorded residual risk

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses（可偽性）：(a) **還原** clear（移除 `OnCallEnd` 的 yield）→ T001 unit pin 與新 Example 必須失敗（非真空；`⠋ Executing …` 殘留）；(b) 讓 clear 改為 **resume-after-tail** → 確認下一回合的 frame 寫入把 resumed frame 留在上一列（殘留移位），使新 Example 仍失敗——據此佐證 D1 的 **no-resume** 選擇；觀察到即還原，重跑確認綠燈。

- [ ] T005 [REGRESSION] 跑全測試 + 見證 + `make verify` + 拓樸稽核
  - Read: `research.md` -> `Decision 3`, `Decision 4`, `Decision 6`；`specs/truth/techstack.md` -> Turn progress spinner row
  - 做：
    - `go test -count=1 ./...` 全綠（unit + godog E2E）。
    - 見證 (a)/(b) 如上，觀察到即還原，重跑確認綠。
    - `make verify` **OK**；`gofmt -l .` clean；Gherkin/DSL 拓樸稽核 **PASSED**（44 features · 16 root + 310 module rows · **1576** steps —— 與 `/axb-dsl-refine` 交付時相同；本輪無再改 feature/DSL）。
    - `go.mod`／`go.sum` 不變（stdlib-only）。
    - 確認 gated-off（非 terminal `stderr`／`-r`）路徑 byte-identical（無 spinner、無 clear）。
  - 不做：不放寬任何 assertion；不為轉綠而移除見證。

- [ ] T006 subagent review (round quality gate)
  - Read: `internal/cli/composite_observer.go`、`internal/cli/composite_observer_test.go`、`specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`、`specs/truth/techstack.md`、`specs/plans/035-spinner-tail-residue/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：修復僅動 `composite_observer.go`（`callRenderer`／`spinner.go` 未改）；`Spinner.OnCallEnd` 仍 no-op；final/ gated-off 路徑 byte-identical；flags／exit codes／class phrase／line formats／cadence／`stdout` 皆未變；`techstack.md` 與 `truth-delta.md` 一致；無新相依。

> Round review 由 orchestrator 執行：修復只動 `internal/cli/composite_observer.go`（+ 新 test 檔）；`callRenderer`／`internal/ui/spinner.go` **未改**（`Spinner.OnCallEnd` 仍 no-op）；`go.mod`/`go.sum` 不變（stdlib-only）；`make verify` **OK**；拓樸稽核 **PASSED**（44 features · 16 root + 310 module rows · 1576 steps）；可偽性見證 (a)/(b) 重現後還原 —— (a) 移除 yield → unit pin 與新 Example 失敗（`⠋ …` 殘留），(b) 改為 resume-after-tail → 新 Example 仍失敗（殘留移位），佐證 no-resume。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Turn progress spinner (operator) row (round-035 clause) | T003（交付修復）、T005（`make verify`／稽核）、T006 | PASS |
| `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature`（新 Rule/Example） | T002（RED 見證）、T003（交付 GREEN）、T005、T006 | PASS |
| `specs/truth/features/cli/chat/dsl.md`（round-035 note） | T002、T003（Read）；T006（一致性檢查） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY `techstack.md` | T003、T005、T006 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP (`specs/truth/contracts/**`) | 豁免（NOOP 不建任務；T005/T006 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP (`specs/truth/data/**`) | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY feature + `chat/dsl.md` | T002（新 Example RED）、T003（GREEN）、T005（Test Scope 回歸）、T006 | PASS |
| `research.md` -> Decision 1（phase-boundary clear、no-resume；拒絕 resume-after-tail） | T003（交付）、T005 見證 (b) 佐證、T006 | PASS |
| `research.md` -> Decision 2（fix 於 `compositeObserver.OnCallEnd`，keyed on `!final`） | T001（pin 契約）、T003（交付）、T006 | PASS |
| `research.md` -> Decision 3（E2E witness = gate + tool round + whole-stream row） | T002（RED 見證）、T003（GREEN） | PASS |
| `research.md` -> Decision 4（既有測試為何漏接；witness 缺口） | T002（確認新 Example 覆蓋缺口）、T006 | PASS |
| `research.md` -> Decision 5（unit ordering pin） | T001（交付 pin）、T006 | PASS |
| `research.md` -> Decision 6（範圍守門；無其他變更） | T003、T005、T006 | PASS |
| `research.md` -> Recorded residual risk（boundary fragility: no write between tail and next activation） | T003 邊界註解、T006（確認 invariant 成立） | PASS |
| `spec.md` -> US1（FR-001–FR-005） | T001、T002、T003 | PASS |
| `spec.md` -> 全域（FR-006、FR-007） | T003（不改 surfaces）、T005（gated-off byte-identical）、T006 | PASS |
| `spec.md` -> 邊界情況（multi-call／tool-less／bound-reached／gated-off／narrow terminal） | T003（multi-call trace）、T005（narrow-terminal Example 回歸；gated-off byte-identical） | PASS |
| `spec.md` -> SC-001–SC-004 | T001（SC-004 unit）、T002/T003（SC-001/002）、T005（SC-003/004） | PASS |
| `plan.md` -> Interface inventory / Wave（1 CLI end → `/axb-dsl-refine`；api/data NOOP；ui skipped） | T002/T003（CLI 介面交付）、T005（不觸及 API/資料） | PASS |
| operator 拍板（Q1(a) yield、Q2 純 bug fix、Q3 siblings out、Q4 E2E+unit） | T001（unit）、T002/T003（E2E）、T003（fix 落點）、T006 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
