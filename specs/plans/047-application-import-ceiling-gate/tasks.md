# Tasks: application import-ceiling gate (RULE-E) + application-tier import baseline (round 047 — implementation half)

**Plan Package**: `specs/plans/047-application-import-ceiling-gate`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0016-application-import-ceiling.md`, `docs/decisions/0011-layer-discipline-gate.md`, `tools/arch/arch_test.go`, `tools/arch/baseline.txt`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 build/quality-pipeline（tooling）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D2/D10；stdlib-only，`go.mod`／`go.sum` 不動；無新 Makefile target）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`；驗證由 **gate 自身的 exit code** 與可偽性見證承擔（非 Gherkin，`spec.md` A2 / `research.md` D8/D9）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`tools/arch/arch_test.go` 新增 RULE-E + `tools/arch/baseline.txt` 新增 3 行**；truth 變更（`specs/truth/techstack.md` Layer-discipline gate row）與 **ADR 0016** 已由 `/axb-technical-research` 於 plan/truth half 交付，並已記入 `truth-delta.md`。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 RULE-A/B/C/D 的判定、既有 gate 語意、任何 `internal/**`／`cmd/**` 產品碼、`Makefile`。

## Round-047 locked decisions (implementation constraints — MUST)

> 來自 `research.md` D1–D11、`spec.md`（FR-001–013 / SC-001–006）與 **ADR 0016**。

- **[RULE-E / D1, ADR 0016 D1]** 對**受治理 application tier**（`internal/app/**` tier 2、`internal/cli` tier 6），一個 `internal/**` import 為 **violation**，**除非**被 import 的 package 屬於 **sanctioned set**：`internal/domain/**`、`internal/config`、`internal/home`、`internal/app/**`。**stdlib** 依建構即允許；**third-party** module import **不在 RULE-E 範圍**（recorded residual，ADR D6）。
- **[DEDUP / D11]** violation 以**邊（edge）**去重（與 RULE-A/B/C 共用一份 sorted edge list，不重複計數、無順序依賴）。
- **[BOTH TIERS / D6, Q3]** RULE-E 同時治理 `internal/app/**` 與 `internal/cli`；sanctioned set 共用。故 `cli → app/**` 與 `app/** → config` 為 **sanctioned**，`internal/app/**` 目前 **0** residual。
- **[NORMATIVE SET / D3, Q2]** sanctioned set 為 guard **tier 表**內的**規範性**段落（ADR 0011 D7 唯一機讀來源）；**default-deny**（未列即 violation）。
- **[FAIL-ON-STALE ALLOW-LIST / D3, Q2]** self-test 斷言**每個 sanctioned entry 都仍被使用**（至少一個受治理 application tier import 它）；未使用的 sanctioned entry → **fail**。
- **[BASELINE / D4, Q1]** committed `tools/arch/baseline.txt`（既有、唯一 ratchet）**新增 3 行** RULE-E residual：`internal/cli -> internal/agent`、`internal/cli -> internal/ui`、`internal/cli -> internal/ui/tui/prompt`（Go `sort.Strings`、ASCII ` -> `、module-relative；**generated 非手抄**）。RULE-A/B/C 維持 **0**；ratchet 語意不變（新 violation→fail；stale→fail；移除仍在違規的行→fail）。
- **[SCOPE / D5, ADR D6]** RULE-E 只管 application tier 對 `internal/**` 的 import；**不改** RULE-A/B/C/D；**不做** third-party allow-list；`cmd/**`／`tests/**`／`tools/**` exempt。
- **[INHERIT MECH / D10]** 沿用既有機制，**不重造**：module-root 錨定 `go list`（ADR 0011 D4）、`CROSS_TARGETS` union + child env 過濾（D5）、merged production+test graph 供 rule、**production-only** graph 供 SCC/cycles（D8）、deterministic baseline（D9）、`-count=1`。
- **[NO NEW TARGET / D2]** RULE-E **rides** 既有 `verify-architecture`（已是 `make verify` 成員）；`Makefile` **不動**（no new member）。
- **[WITNESS / D9, #92 AC5]** 見證 = **gate + unit self-tests**，非 E2E；可偽性見證 (a)/(b)/(c) reproduced then reverted（ADR 0010）。
- **[TRUTH / D7]** truth（`specs/truth/techstack.md` Layer-discipline gate row + ADR 0016）已於 plan half 交付；本輪**不**再改 truth（Task runner row 為 NOOP）。
- **[NO-DEP]** stdlib-only；`go.mod`／`go.sum` 不動；POSIX-only。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D2/D10；無新相依、無新 Makefile target，`go.mod`／`go.sum` 不動）。不得把 RULE-E 骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點骨架；RULE-E 直接落點於**既有** `tools/arch/arch_test.go`（由 T001 交付），baseline 由 T002 產生。`Makefile` 不動（D2）。

## Phase 3: Implementation & Alignment (RULE-E, RED-first)

**Goal**: 在既有 `tools/arch` guard 內新增 **RULE-E**（sanctioned-set 規範性段落 + allow-list ceiling 判定 + default-deny + fail-on-stale allow-list 覆蓋自檢），產生並提交 **3** 行 RULE-E baseline，使 `dev` 上綠燈、新 unsanctioned import 紅燈、unused sanctioned entry 紅燈、stale 條目紅燈；既有 RULE-A/B/C/D 判定與 `Makefile` 不變。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗證語言為 Go test（gate 自身）。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Build & Tooling（**Layer-discipline gate** row；Task runner row 之 `verify` 成員不變）
- `docs/decisions/0016-application-import-ceiling.md`（D1–D9；sanctioned set 為規範性機讀來源）
- `docs/decisions/0011-layer-discipline-gate.md`（D1–D10；tier 表、ratchet、機制）
- `specs/plans/047-application-import-ceiling-gate/research.md` -> D1–D11
- `specs/plans/047-application-import-ceiling-gate/plan.md` -> RULE-E 表 + 預期變更（`tools/arch/**` CHANGED；`Makefile` unchanged）
- `tools/arch/arch_test.go`、`tools/arch/baseline.txt`（既有 guard 實作與 baseline 格式）

**Boundary**:
- 只改 `tools/arch/arch_test.go` 與 `tools/arch/baseline.txt`（+ 文件）；**不**碰產品碼、**不**改 RULE-A/B/C/D 與既有 gate 語意、**不**改 `Makefile`、**不**改 truth。
- sanction 集為規範性機讀來源；`techstack.md` 以 predicate 指涉並 cite ADR 0016（RF-3 精神）。
- review 啟動 subagent；通過前不解鎖 Phase 4。

- [ ] T001 [P] 在 `tools/arch/arch_test.go` 實作 RULE-E（sanctioned-set 段落 + ceiling 判定 + 覆蓋自檢）；更新 `baseline.txt` 檔頭註解
  - Read:
    - `specs/plans/047-application-import-ceiling-gate/research.md` -> D1, D3, D4, D5, D6, D11
    - `docs/decisions/0016-application-import-ceiling.md` -> D1, D2, D3, D4, D5, D6, D7
    - `docs/decisions/0011-layer-discipline-gate.md` -> D2, D3, D4, D5, D7, D8, D9
    - `tools/arch/arch_test.go`（既有 `tier()`/`evaluate()`/`enumerate()`/`cycles()`/`readBaseline()`/`writeBaseline()`/`selfTestPredicate()`/entry test）
    - `specs/truth/techstack.md` -> Build & Tooling（Layer-discipline gate row）
  - 做：
    - 於 guard 的 **tier 表**新增一個**規範性 sanctioned-set** 段落（每個受治理 application tier → sanctioned import **prefix** 清單）：`internal/cli` 與 `internal/app/**` 皆 → `internal/domain/`, `internal/config`, `internal/home`, `internal/app/`。
    - 在 `evaluate(graph)` 內新增 **RULE-E**：對 **tier 2 與 tier 6** 的每個 `internal/**` import，若其 package 不匹配該 tier 的 sanctioned prefix → 產生 violation 行 `<src> -> <dst>`；**default-deny**；與 RULE-A/B/C/D 的結果**以 edge 去重**後一起 sorted 回傳。
    - 新增 **allow-list 覆蓋自檢**：斷言**每個** sanctioned entry 至少被一個受治理 application-tier package 使用（否則 `t.Errorf` — fail-on-stale allow-list）；並斷言 tier 2 / tier 6 對 sanctioned 集**無**豁免（default-deny 正面斷言）。
    - 於 entry test `TestVerifyRealArchitecture` **內**補上 RULE-E 與其覆蓋自檢的呼叫與「已執行」顯式斷言（維持 N-3：三+性質皆在單一 entry test 內，`-run` 不得靜默略過）。
    - 更新 `tools/arch/baseline.txt` **檔頭註解**：說明 RULE-A/B/C/D **+ RULE-E**，並保留 `make verify-architecture-update` 生成法與 `-count=1` 契約（承襲 round-042 fold-review #3 Fold 2）。**不得**手改內容行（內容行由 T002 生成）。
  - RED-first 期望：完成 T001、**未**更新 baseline 前，`make verify-architecture` **非零失敗**，列出 **3** 條 RULE-E edge（`internal/cli -> internal/agent`、`internal/cli -> internal/ui`、`internal/cli -> internal/ui/tui/prompt`）；證明規則**非空轉**（對應 SC-002）。
  - 不做：不改 `Makefile`（D2）；不手抄 baseline 內容行（T002）；不碰產品碼／truth／既有 RULE-A/B/C/D 判定；不引入相依；不動 `cmd/**`／`tests/**`。

- [ ] T002 以 gate 重新生成並提交 `tools/arch/baseline.txt`（generated，非手抄；+3 RULE-E 行）
  - Read:
    - `specs/plans/047-application-import-ceiling-gate/research.md` -> D4
    - `docs/decisions/0016-application-import-ceiling.md` -> D5（worked example）
    - `specs/plans/047-application-import-ceiling-gate/spec.md` -> SC-001（3 RULE-E residuals）
    - `tools/arch/baseline.txt`（既有 header-only）
  - 做：以 `make verify-architecture-update`（或 `go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline`）**生成** `tools/arch/baseline.txt`；確認內容為 sorted、ASCII ` -> `、module-relative 的 **3** 行（`internal/cli -> internal/agent`、`internal/cli -> internal/ui`、`internal/cli -> internal/ui/tui/prompt`），與 ADR 0016 的 worked example 逐行一致；**RULE-A/B/C 維持 0**。
  - 驗證：重跑 `make verify-architecture` → **綠**（3 baselined、0 new、0 stale、0 cycles）。
  - 不做：不手改任何 baseline 行；不為了綠而移除違規或放寬 RULE-E。

- [ ] T003 subagent review (phase quality gate)
  - Read: `tools/arch/arch_test.go`、`tools/arch/baseline.txt`、`research.md` -> D1–D11、`docs/decisions/0016-application-import-ceiling.md`
  - 檢驗：RULE-E 為 **allow-list ceiling**（非 direction 規則），同時治理 tier 2 與 tier 6；sanctioned set 為唯一機讀來源且 **default-deny**；**fail-on-stale allow-list** 覆蓋自檢存在且非空轉；violation 與 RULE-A/B/C **以 edge 去重**；3 行 baseline **可由規則導出**；未動產品碼／truth／RULE-A/B/C/D／`Makefile`；`-count=1` 契約保留。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 RULE-E 在 `dev` 上綠、對**新** unsanctioned import 與**未使用 sanctioned entry** 與**stale** 條目皆紅（可偽性），且既有 RULE-A/B/C/D、既有 gate、產品行為未受影響。

**Test Scope**: `tools/arch/**`（unit；`-tags=arch`）；whole-suite `go test ./...`；`make verify`；Gherkin/DSL topology audit。

- [ ] T004 [REGRESSION] 可偽性見證 (a) 新 unsanctioned import ⇒ 紅 / (b) unused sanctioned entry ⇒ 紅 / (c) stale baseline 條目 ⇒ 紅
  - Read: `research.md` -> D1, D3, D4, D9；`docs/decisions/0010-test-deadline-decoupling.md`（見證 doctrine）；`docs/decisions/0016-application-import-ceiling.md` -> D4, D5；`tools/arch/arch_test.go`
  - 做：
    - **(a) 新 unsanctioned import**：暫時於 `internal/cli` 加入一個不在 sanctioned set 的 `internal/**` import（例如 `internal/infrastructure/mcp` 或一個新的 `internal/telemetry`）→ `make verify-architecture` **非零失敗並列出** 該 `src -> dst`（RULE-E）；**還原**後確認綠。
    - **(b) unused sanctioned entry**：暫時於 sanctioned-set 段落加入一個**未被使用**的 entry（例如 `internal/infrastructure/…` 或一個不存在於應用層 import 集的 prefix）→ 覆蓋自檢 **fail**（fail-on-stale allow-list）；**還原**後確認綠。
    - **(c) stale**：暫時自 `tools/arch/baseline.txt` 移除一行（其違規仍在）→ gate 回報**新 violation** 而紅；改成暫時讓某一 baselined 行不再違規（例如以注入的 sanctioned entry 讓該 edge 合法化）→ gate 回報 **stale** 而紅；兩種皆觀察到即**還原**，重跑確認綠。
  - 不做：不放寬任何 assertion 以「讓它過」；見證後必須還原到 HEAD。

- [ ] T005 [REGRESSION] 全量回歸 + `make verify` + topology audit + 範圍檢查
  - Read: `research.md` -> D2, D7, D10；`specs/truth/techstack.md` -> Build & Tooling（Layer-discipline gate row；Task runner row）
  - 做：
    - `make verify-architecture` → 綠（3 RULE-E baselined、RULE-A/B/C **0**、0 new、0 stale、0 cycles）。
    - `go test -count=1 ./...` 全綠（unit + godog E2E；`tools/arch` 於預設 tags 下為 `[no test files]`）。
    - `make verify` **OK**（既有 gates 不變；`verify-architecture` 未被新增成員）。
    - Gherkin/DSL topology audit 重跑 → **unchanged**（44 features · 6 modules · 16 root + 310 module rows · 1576 steps）。
    - `gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 確認僅動 `tools/arch/**`（+ 文件：`docs/decisions/**`、`specs/truth/techstack.md`、plan package、`STATUS.md`）；`go.mod`／`go.sum` 不變；`Makefile` 未動；`internal/**`／`cmd/**` 產品碼未動。
  - 不做：不為了綠而放寬 assertion、改產品碼或 baseline 手抄。

- [ ] T006 subagent review (round quality gate)
  - Read:
    - `tools/arch/arch_test.go`、`tools/arch/baseline.txt`
    - `specs/truth/techstack.md`（Layer-discipline gate row；Task runner row）、`docs/decisions/0016-application-import-ceiling.md`
    - `specs/plans/047-application-import-ceiling-gate/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：`dev` 上綠；**新 unsanctioned import ⇒ 紅**；**unused sanctioned entry ⇒ 紅**；**stale ⇒ 紅**；RULE-A/B/C 仍 0；sanctioned set 為唯一權威且與 `techstack.md` predicate 一致；RULE-E 以 edge 與 RULE-A/B/C 去重；無新相依；`Makefile` 未動；既有 gate 語意與 `stdout`/`stderr` 未變；**無產品碼變更**；`techstack.md` 與 `truth-delta.md` 一致。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Build & Tooling（**Layer-discipline gate** row，含 RULE-E + sanctioned set + 3-residual ratchet） | T001（依 predicate/tier/sanctioned set 實作）、T002、T005、T006 | PASS |
| `specs/truth/techstack.md` -> Build & Tooling（**Task runner** row：`verify` aggregate 不變） | T005（`make verify`）、T006 | PASS |
| `docs/decisions/0016-application-import-ceiling.md`（RULE-E + sanctioned set + default-deny + fail-on-stale + scope + worked 3） | T001（D1–D6）、T002（worked example 對齊）、T006 | PASS |
| `docs/decisions/0011-layer-discipline-gate.md`（tier 表、ratchet、機制；被 RULE-E 沿用） | T001（沿用 D4/D5/D7/D8/D9）、T006 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（techstack Layer-discipline gate row） | 已於 plan half 交付；T001／T005 驗證一致 | PASS |
| `truth-delta.md` -> `/axb-technical-research` NOOP（techstack Task runner row） | 豁免（NOOP；T005 驗證 `make verify` 成員不變） | PASS |
| `truth-delta.md` -> Governance ADD（ADR 0016 + index） | 已於 plan half 交付；T001／T006 依 ADR 實作/驗證 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T005 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`specs/truth/features/cli/**`） | 豁免（NOOP 不建任務；T005 驗證 topology audit 不變） | PASS |
| `research.md` -> D1（RULE-E ceiling over the tier table） | T001、T002、T003 | PASS |
| `research.md` -> D2（rides existing guard/target；no new Makefile target） | T001、T005 | PASS |
| `research.md` -> D3（normative sanctioned set + default-deny + fail-on-stale allow-list） | T001、T003、T004(b) | PASS |
| `research.md` -> D4（baseline +3；ratchet reused） | T001、T002、T004(c) | PASS |
| `research.md` -> D5（scope boundary: application tiers only; stdlib allowed; third-party out） | T001、T006 | PASS |
| `research.md` -> D6（binds both application tiers；`app→config` sanctioned） | T001、T002（`app/**` 0 residual） | PASS |
| `research.md` -> D7（ADR 0016 + truth MODIFY） | 已於 plan half 交付；T001／T006 | PASS |
| `research.md` -> D8（BDD techstack unchanged; api/data/dsl NOOP） | T005、T006 | PASS |
| `research.md` -> D9（witness = gate + unit self-tests + falsifiability a/b/c） | T004、T006 | PASS |
| `research.md` -> D10（inherited mechanism: anchor/union/merged+production graphs/deterministic/`-count=1`） | T001、T005 | PASS |
| `research.md` -> D11（dedup by edge；guard self-compliance） | T001、T003 | PASS |
| `spec.md` -> US1（FR-001–FR-004, NFR-001） | T001、T004(a) | PASS |
| `spec.md` -> US2（FR-005–FR-007, NFR-002） | T001、T002、T004(c) | PASS |
| `spec.md` -> US3（FR-008–FR-010, NFR-003） | 已於 plan half 交付；T001（覆蓋自檢）、T006 | PASS |
| `spec.md` -> 全域（FR-011–FR-013, NFR-004–NFR-005） | T004、T005（no product code；make verify + audit） | PASS |
| `spec.md` -> SC-001（green on dev, 3 RULE-E baselined；RULE-A/B/C 0） | T002、T005 | PASS |
| `spec.md` -> SC-002（new unsanctioned import ⇒ red, witness a） | T004(a) | PASS |
| `spec.md` -> SC-003（removed entry ⇒ red；stale ⇒ red；unused sanctioned ⇒ red） | T004(b)（unused sanctioned）、T004(c)（removed/stale） | PASS |
| `spec.md` -> SC-004（host-independent union + 0 cycles + default-deny + no new dep） | T001（union/SCC）、T005 | PASS |
| `spec.md` -> SC-005（truth + ADR + no product change + audit） | 已於 plan half 交付；T005、T006 | PASS |
| `spec.md` -> SC-006（residuals + F-4/F-6/F-7/F-8 記於 #101） | closeout Step 8（round-closeout；非 implement task，於 PR/closeout 記錄） | PASS |
| operator 拍板 Q1–Q5（gate-first slice；normative allow-list + fail-on-stale；both tiers；sanctioned set as measured；ADR 0016 + slug） | T001（Q2/Q3/Q4）、T002（Q1）、T006 | PASS |
| `plan.md` -> Structure（`tools/arch/**` CHANGED；`Makefile` unchanged；truth/ADR 已 MODIFY+ADD；no product code） | T001、T002、T005 | PASS |
| `plan.md` -> Scope notes（api/data/dsl NOOP；ui/spec-by-example skipped） | T005、T006 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
