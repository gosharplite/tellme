# Tasks: resolver test load tolerance (round 041)

**Plan Package**: `specs/plans/041-di-resolver-test-load-tolerance`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0010-test-deadline-decoupling.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 test-harness determinism（tooling）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D5；stdlib-only，`go.mod`／`go.sum` 不動）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`；驗證由 Go unit test 自身與可偽性見證承擔（非 Gherkin，`research.md` D6）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是**重調的 resolver harness**（`internal/infrastructure/di/mcp_factory_test.go`）；truth 變更（`specs/truth/techstack.md` Testing rows + **ADR 0010**）已由 `/axb-technical-research` 於 plan/truth half 交付並記入 `truth-delta.md`。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：`internal/infrastructure/di/mcp_factory.go`（產品碼）、任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 gate 語意。

## Round-041 locked decisions (implementation constraints — MUST)

> 來自 `research.md` D1–D8 與 `spec.md` 的 operator-locked Q1–Q6。

- **[D1 / Q1]** 正向測試 `TestNewGhTokenResolver_TrimsToken` 改用**寬鬆的 test-local bound**（`generousResolverBound = 30 * time.Second`），**不得**再用生產 2s fast-fail 常數。
- **[D2 / Q2]** `writeFakeGh` 給 shim **dominant PATH**（shim 目錄在前 + 繼承 PATH；繼承值須在 `t.Setenv` 前取得）；hanging shim 化簡為 `exec sleep 3`，並移除 round-032 N1 的 in-shim PATH 還原與其註解。
- **[D3 / Q3]** `TestNewGhTokenResolver_Bounded` 維持 `200 ms` bound，新增 **non-vacuity pin**（`elapsed >= bound`），ceiling 由 `1s` 改為 `boundedCeiling = 2 * time.Second`。
- **[D4]** `TestNewGhTokenResolver_MissingGh` = **recorded non-change**（不 spawn child）。
- **[D5]** 無 `t.Skip`／無 retry／無 vacuous assertion；無 Go `time.Sleep`（shim 的 shell `sleep` 不算）；stdlib-only。
- **[D6 / Q4]** 見證：`go test -count=20 ./...` **under contention** 綠；可偽性 (a)/(b)/(c) 各自重現後還原。
- **[D7 / Q5]** truth（`techstack.md` Testing rows + ADR 0010）由 `/axb-technical-research` 交付。
- **[D8 / Q6]** 同類 wall-clock assertion 的 sibling 掃描 = **out of scope**（forward item，另開一輪）。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D5；無新相依，`go.mod`／`go.sum` 不動）。不得把 harness 骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無測試落點骨架（無新 DSL 句／stepdef）；harness 落點即既有 `mcp_factory_test.go`（test file），由 T001 直接交付。

## Phase 3: Test Alignment & Implementation (test harness, RED-first)

**Goal**: 讓 `internal/infrastructure/di` 的 resolver unit tests 對 host contention 免疫（正向測試與生產常數解耦；shim 用 dominant PATH；bounded test 保留為唯一 falsifiability carrier 並加 vacuity pin），且**不**弱化任何既有斷言。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗證語言為 Go unit test。

- [ ] T001 [UNIT] 重調 resolver harness 的三處斷言與 fixture（RED-first）
  - Read:
    - `specs/plans/041-di-resolver-test-load-tolerance/research.md` -> D1, D2, D3, D4, D5
    - `specs/truth/techstack.md` -> Testing & Verification（Host test harness；Pure-helper unit tests）
    - `docs/decisions/0010-test-deadline-decoupling.md`
    - `internal/infrastructure/di/mcp_factory_test.go`（現況）
  - 做（**測試檔內**，不動產品碼）：
    - 新增 test-local 常數：`generousResolverBound = 30 * time.Second`、`boundedCeiling = 2 * time.Second`（與 `boundedResolverBound = 200 * time.Millisecond` 並列，附 ADR 0010 一行註解）。
    - `TestNewGhTokenResolver_TrimsToken`：bound 由 `2 * time.Second` → `generousResolverBound`（D1）。
    - `TestNewGhTokenResolver_Bounded`：ceiling 由 `time.Second` → `boundedCeiling`，**新增 non-vacuity pin** `if elapsed < boundedResolverBound { t.Fatalf(...) }`（D3）。
    - `writeFakeGh`：`t.Setenv("PATH", dir + string(os.PathListSeparator) + inherited)`（`inherited := os.Getenv("PATH")` 於 `t.Setenv` **之前**取得）（D2）；更新 fixture 註解（shadow-vs-resolve）。
    - `TestNewGhTokenResolver_Bounded` 的 shim 由 `PATH="/usr/bin:/bin"; exec sleep 3` 化簡為 `exec sleep 3`；移除 N1 的 PATH 還原與其註解（D2）。
    - `TestNewGhTokenResolver_MissingGh`：確認**不需**變更（recorded non-change，D4）。
  - RED-first 見證（在 T002–T004 逐一被要求重現）：T001 完成後，先以**未**放寬的姿態確認既有斷言語意未被削弱（ceiling 放寬仍保留「unbounded ⇒ red」的可偽性——由 T003 見證）。
  - 不做：不改 `mcp_factory.go`／任何產品碼；不放寬／移除任何既有 assertion；不引入 `time.Sleep`、`t.Skip` 或 retry；不加相依。

## Phase 4: Verification & Regression

**Goal**: 證明 harness 對 host contention 免疫（under contention 綠）、可偽性三向保留（(a) bound 未放寬即紅、(b) unbounded ⇒ ceiling 紅、(c) vacuous shim ⇒ vacuity pin 紅），且既有 gates／行為未受影響。

**Test Scope**: `internal/infrastructure/di/mcp_factory_test.go`（unit）；whole-suite `go test ./...`；`./tests/e2e`（作為 contention 來源）。

- [ ] T002 [UNIT] 可偽性見證 (a) —— 正向 bound 未放寬即紅（under load）
  - Read: `research.md` -> D1, D6(a)；`internal/infrastructure/di/mcp_factory_test.go`
  - 做：暫時把 `TestNewGhTokenResolver_TrimsToken` 的 bound 還原為 `2 * time.Second`，在 **contention** 下（同時跑 `go test ./tests/e2e/`）執行 whole-suite（或至少 `./internal/infrastructure/di/` 於 whole-suite 情境）→ 觀察 `TestNewGhTokenResolver_TrimsToken` 以 `signal: killed` / `2.00s` **紅**；**還原**為 `generousResolverBound` 後再跑一次確認綠。
  - 不做：不改任何其他斷言以「讓它過」；不用 retry 掩蓋。

- [ ] T003 [UNIT] 可偽性見證 (b) —— unbounded resolver ⇒ ceiling 紅
  - Read: `research.md` -> D3, D6(b)；`internal/infrastructure/di/mcp_factory.go`（唯讀）、`mcp_factory_test.go`
  - 做：暫時讓 resolver 忽略 `bound`（unbounded，例如 `context.WithTimeout(ctx, time.Hour)` 的 local 變體 **僅在測試** 或以測試內等效方式模擬），使 shim 真的 `sleep 3` 完成 → `TestNewGhTokenResolver_Bounded` 量得 ≈3s `> boundedCeiling (2s)` → **紅**；**還原**後確認綠。
  - 不做：不提交任何產品碼變更（見證後必須還原到 HEAD）。

- [ ] T004 [UNIT] 可偽性見證 (c) —— vacuous shim ⇒ vacuity pin 紅
  - Read: `research.md` -> D3, D6(c)；`internal/infrastructure/di/mcp_factory_test.go`
  - 做：暫時把 bounded test 的 shim 改成**立即失敗**（例如 `exit 1` 或一個 child 解析不到的 `sleep`），使 resolver 幾乎立即回錯、`elapsed < boundedResolverBound` → **non-vacuity pin** 觸發 **紅**（證明 pin 非真空）；**還原**後確認綠。
  - 不做：不留任何 vacuous／放寬的 assertion。

- [ ] T005 [REGRESSION] under-contention acceptance + 全 gate + 範圍檢查
  - Read:
    - `specs/plans/041-di-resolver-test-load-tolerance/research.md` -> D5, D6, D8
    - `specs/truth/techstack.md` -> Testing & Verification（Host test harness；Pure-helper unit tests）
  - 做：
    - **under contention**：並行執行一個 `go test -count=1 ./tests/e2e/`（壓力來源）與 `go test -count=20 ./internal/infrastructure/di/`（或 `go test -count=20 ./...`），全程 `internal/infrastructure/di` **綠**（SC-001）。
    - `go test -count=1 ./...` 綠；`make verify`（`verify-no-test-sleep` + `verify-no-network` + `vet` + `verify-cross-compile` 4/4 + `verify-mcp-sdk-confinement` + `lint` + `vulncheck`）**OK**。
    - Gherkin/DSL topology audit 重跑 → **unchanged**（44 features · 16 root + 327 module rows · 1674 steps）。
    - `gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 確認**未動任何產品 Go 檔**（僅 `mcp_factory_test.go` + docs）；`go.mod`／`go.sum` 不變。
  - 不做：不為了讓 gate 綠而放寬 assertion 或改產品碼。

- [ ] T006 subagent review (round quality gate)
  - Read:
    - `internal/infrastructure/di/mcp_factory_test.go`
    - `specs/truth/techstack.md`（Testing & Verification）、`docs/decisions/0010-test-deadline-decoupling.md`
    - `specs/plans/041-di-resolver-test-load-tolerance/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：正向測試已與生產常數解耦；bounded test 為唯一 falsifiability carrier（200ms + vacuity pin + 2s ceiling，且 2s < unbounded ≈3s）；dominant PATH 且 `gh` 仍被 shadow；無 `t.Skip`／retry／`time.Sleep`；**無產品碼變更**；`techstack.md` 與 `truth-delta.md` 一致；既有 gate 語意未變。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Testing & Verification（Host test harness row：ADR 0010 規則） | T001（依規則實作）、T005、T006 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests row：di resolver harness） | T001、T005、T006 | PASS |
| `docs/decisions/0010-test-deadline-decoupling.md` | T001（Read + 依 D1–D4 實作）、T006 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md` ×2 rows） | T001、T005、T006 | PASS |
| `truth-delta.md` -> `/axb-technical-research` ADD（ADR 0010 + index） | T001（Read）、T006 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T005 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`specs/truth/features/cli/**`） | 豁免（NOOP 不建任務；T005 驗證 topology audit 不變） | PASS |
| `research.md` -> D1（generous test-local bound） | T001、T002 | PASS |
| `research.md` -> D2（dominant PATH；retire N1） | T001 | PASS |
| `research.md` -> D3（non-vacuity pin + 2s ceiling） | T001、T003、T004 | PASS |
| `research.md` -> D4（missing-gh non-change） | T001 | PASS |
| `research.md` -> D5（no skip/retry/sleep；stdlib-only） | T001、T002、T003、T004、T005 | PASS |
| `research.md` -> D6（witness under contention + (a)/(b)/(c)） | T002、T003、T004、T005 | PASS |
| `research.md` -> D7（governance：ADR 0010 + techstack） | T001、T006 | PASS |
| `research.md` -> D8（scope：no production change；siblings out） | T005（範圍檢查）、T006 | PASS |
| `spec.md` -> US1（FR-001–FR-005, NFR-001–NFR-002） | T001–T005 | PASS |
| `spec.md` -> US2（FR-006–FR-008, NFR-003） | T001（Read）、T006 | PASS |
| `spec.md` -> 全域（FR-009, NFR-004–NFR-005） | T005（no production file；go.mod 不變；make verify + audit） | PASS |
| `spec.md` -> SC-001–SC-004 | T005（SC-001/004）、T002（SC-002）、T003+T004（SC-003） | PASS |
| `plan.md` -> Structure（`mcp_factory_test.go` CHANGED；`mcp_factory.go` UNCHANGED；`techstack.md`/ADR MODIFY+ADD） | T001、T005 | PASS |
| `plan.md` -> Scope notes（api/data/dsl NOOP；ui/spec-by-example skipped） | T005、T006 | PASS |
| operator 拍板 Q1–Q6（delegated recommendations） | T001（Q1–Q3）、T002–T004（Q4）、T001/T006（Q5）、T005/T006（Q6） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
