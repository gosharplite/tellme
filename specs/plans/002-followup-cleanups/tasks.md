# Tasks: Follow-up Cleanups (round 002)

**Plan Package**: `specs/plans/002-followup-cleanups`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task（如 build 參數、gate 目標），不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 只做本輪新增技術的基礎建設、技術環境與最後的 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

## Phase 1: Setup

**Goal**: 引入本輪新增的驗證技術——`golangci-lint`（啟用 `errcheck`）與 `govulncheck`——建立 `.golangci.yml` 政策檔與 Makefile 目標（`lint`、`vulncheck`）並接入 `verify`／`check`，最後以 smoke-test 驗證兩個 gate 可運行。不寫 DSL 語意、不寫產品行為。

- [X] T001 建立 lint 聚合與漏洞掃描 gate
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling（Lint aggregator `golangci-lint`（with `errcheck`）；Dependency vulnerability scan `govulncheck`）
    - `specs/plans/002-followup-cleanups/research.md` -> Decision 2
  - 新增 `.golangci.yml`（至少啟用 `errcheck`），在 `Makefile` 新增 `lint`（`golangci-lint run ./...`，找不到時明確報錯）與 `vulncheck`（`govulncheck ./...`）目標，並將兩者接入 `verify` 與 `check` 聚合目標。
  - 工具解析沿用既有慣例（`command -v` + `$GOPATH/bin` fallback），不引入新的安裝步驟。

- [X] T002 Gate smoke-test
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling
  - 在乾淨樹上執行 `make lint`、`make vulncheck` 與 `make verify`，確認兩 gate 可運行且現有測試不受影響。不修產品碼。

## Phase 2: Foundational

**Goal**: 為 F9 的 pure-helper 單元測試預留互斥落點檔案骨架（Zero Shared Edits），供 Phase 3 並行分派。

- [X] T003 建立 pure-helper 單元測試落點骨架
  - Read:
    - `specs/plans/002-followup-cleanups/research.md` -> Decision 1 & 3（pure helpers, stdlib `testing`）
    - `specs/plans/002-followup-cleanups/plan.md` -> Source-code structure
  - 只做：建立三個互斥測試檔骨架 `internal/config/config_test.go`、`internal/home/home_test.go`、`internal/cli/cli_test.go`（僅 package 宣告，**不**寫測試邏輯）。
  - 不做：不寫任何 test function 或斷言，不碰任何產品檔。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪受影響的自動化測試對齊最新版 truth；含 `truth-delta.md` 的 `ADD` / `MODIFY` / `DELETE` 句，以及 F9 的 pure-helper 單元測試。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/{module}/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given/When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在但語意是舊 truth；依 `dsl.md` 該列改測試。本輪 MODIFY 為「釘住」：把「distinct/相應訊息」改為**精確值**斷言（exit code `2/3/4/5`；stderr 前綴 `tellme: {reason}`）。
- `[BDD-REMOVE]`：`DELETE`。此句已不是 truth；移除仍綁該句的 stepdef／assertion，不得留下保護舊行為的測試。
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 stepdef。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 1 & 3）。只動測試層，不寫產品碼。
- 四個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/dsl.md`
- `specs/truth/features/cli/diagnostics/dsl.md`
- `specs/truth/features/cli/usage/dsl.md`
- `specs/truth/features/cli/configuration/dsl.md`
- `specs/truth/features/cli/workspace/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine`（`DELETE` / `ADD` / `MODIFY` 各列）
- `tests/e2e/steps/`

**Boundary**:
- 一條 DSL 一個 task，落入獨立單一檔案（檔案的 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 任務落入 Phase 2 預留的互斥檔案，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T014 各派一個獨立 subagent；T015 等全部回來再啟動 subagent 來 review。

### DELETE（`--json` 已非 truth）

- [X] T004 [P] [BDD-REMOVE] `When: the operator runs tellme's diagnostic with "--json"`（diagnostics）
  - Read: `specs/truth/features/cli/diagnostics/dsl.md`（該句已移除）；既有落點 `tests/e2e/steps/step_t042_diag_when_runs_diagnostic_json.go`

- [X] T005 [P] [BDD-REMOVE] `Then: tellme emits the resolution status as structured output`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md`（該句已移除）；既有落點 `tests/e2e/steps/step_t049_diag_then_emits_structured_resolution.go`

- [X] T006 [P] [BDD-REMOVE] `Then: tellme emits the unresolved status as structured output`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md`（該句已移除）；既有落點 `tests/e2e/steps/step_t050_diag_then_emits_structured_unresolved.go`

### ADD（`--json` 為未識別 flag → usage）

- [X] T007 [P] [BDD-RED] `When: the operator runs tellme's diagnostic with "--json"`（usage）
  - Read: `specs/truth/features/cli/usage/dsl.md` -> `the operator runs tellme's diagnostic with "--json"`
  - Landing: `tests/e2e/steps/step_t007_usage_when_diagnostic_json.go`

### MODIFY（釘住訊息與退出碼）

- [X] T008 [P] [BDD-ALIGN] `Then: tellme explains on stderr that "{reason}"`（介面根；釘住 `tellme: {reason}` 前綴）
  - Read: `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`；既有落點 `tests/e2e/steps/step_t017_root_then_explains_stderr.go`

- [X] T009 [P] [BDD-ALIGN] `Then: tellme exits with the configuration error code`（釘住 `3`）
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `tellme exits with the configuration error code`；既有落點 `tests/e2e/steps/step_t026_config_then_exits_config_error.go`

- [X] T010 [P] [BDD-ALIGN] `Then: tellme exits with the environment error code`（釘住 `4`）
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `tellme exits with the environment error code`；既有落點 `tests/e2e/steps/step_t038_workspace_then_exits_environment_error.go`

- [X] T011 [P] [BDD-ALIGN] `Then: tellme exits with the usage error code`（釘住 `2`）
  - Read: `specs/truth/features/cli/usage/dsl.md` -> `tellme exits with the usage error code`；既有落點 `tests/e2e/steps/step_t054_usage_then_exits_usage_error.go`

- [X] T012 [P] [BDD-ALIGN] `Then: tellme exits with the diagnostic error code`（釘住 `5`）
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme exits with the diagnostic error code`；既有落點 `tests/e2e/steps/step_t052_diag_then_exits_diagnostic_error.go`

### UNIT（pure-helper 單元測試，F9）

- [X] T013 [P] [UNIT] `internal/config` 解析器表驅動單元測試
  - Read: `specs/plans/002-followup-cleanups/research.md` -> Decision 1 & 3；落點 `internal/config/config_test.go`
  - 覆蓋 `EffectiveSelectedProvider`（env 覆寫 vs 檔案）與 `EffectiveMode`（env → 檔案 → `butler`）與 `ProviderInRegistry` 的優先序矩陣。

- [X] T014 [P] [UNIT] `internal/home` 工作區冪等性單元測試
  - Read: `specs/plans/002-followup-cleanups/research.md` -> Decision 1 & 3；落點 `internal/home/home_test.go`
  - 覆蓋 `EnsureWorkspace`：首次建立、重複重用（不破壞）、路徑為一般檔案時回 `ErrNotDirectory`。

### Phase Review Gate

- [X] T015 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/**/*.feature`
    - `specs/truth/features/cli/**/dsl.md`
    - `tests/e2e/steps/*.go`
    - `internal/{config,home}/*_test.go`
  - 檢驗：無 undefined step；`[BDD-REMOVE]` 的舊 json stepdef 已清除；`[BDD-ALIGN]` 已改為精確值斷言（`2/3/4/5` 與 `tellme: {reason}`）；`[UNIT]` 覆蓋 pure helpers；失敗只能是 assertion 或產品行為。

## Phase 4A: DELETE Feature / DSL Truth - cli `--json`（machine-readable diagnostic）

**Goal**: 移除產品碼中 `--json` flag 與 JSON 診斷輸出，使 `--json` 成為未識別 flag（usage error），並確認新版 truth 成立。

**Shared Must Read**:
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` -> `Feature: Checking the build version and diagnosing setup`
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature` -> `Feature: Rejecting unsupported command-line usage`
- `specs/truth/features/cli/diagnostics/dsl.md`、`specs/truth/features/cli/usage/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine` `DELETE`（diagnostics）與 `ADD`（usage）
- `internal/cli/cli.go` -> `--json` flag、`emitDiagnosticJSON`、`diagnosticJSON`

**Boundary**:
- 移除 `internal/cli/cli.go` 的 `json` 選項、`--json` flag 註冊、`diagnosticJSON` 型別與 `emitDiagnosticJSON`；`renderDiagnostic` 不再接受 `asJSON`；移除不再使用的 `encoding/json` import。
- 只做移除；不新增其他 flag 或行為。`-d` 純文字路徑與其 unresolved 退出碼維持不變。

**Test Scope**:
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature`

- [X] T016 [CODE-REMOVE] 移除 `--json` flag 與 JSON 診斷輸出路徑
- [X] T017 [REGRESSION] 跑 Test Scope，確認 `--json` 已是 usage error 且 `-d` 純文字路徑仍成立

## Phase 4B: MODIFY Feature - 釘住的訊息與退出碼契約

**Goal**: 在 Phase 3 強化的精確斷言下，讓受影響 Feature 全綠（產品已符合；確認釘住契約成立），並在綠燈下整理。

**Shared Must Read**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
- `specs/truth/features/cli/workspace/runtime-home-and-session-workspace.feature`
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature`
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`
- `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`
- `truth-delta.md` -> `/axb-dsl-refine` `MODIFY`（釘住訊息與退出碼）

**Boundary**:
- 產品行為預期不需變更（exit code 與 stderr 訊息已符合）；此 phase 只驗證釘住契約成立並整理共用斷言 helper。
- 不擴及 round-001 以外的行為。

**Test Scope**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
- `specs/truth/features/cli/workspace/runtime-home-and-session-workspace.feature`
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature`
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`

- [X] T018 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T019 [BDD-REFACTOR] 在綠燈下整理精確值斷言與共用 helper

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Lint aggregator `golangci-lint`（+`errcheck`）與 Dependency vulnerability scan `govulncheck` | T001、T002 (Setup) | PASS |
| `specs/truth/techstack.md` -> Test strategy（E2E acceptance path + pure-helper unit tests） | T003、T013、T014 (Foundational / Phase 3) | PASS |
| `specs/truth/techstack.md` -> CLI flag list（`--json` removed） | T016 (Phase 4A) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` DELETE（diagnostics `--json` 3 句） | T004–T006 (Phase 3) + T016 (Phase 4A) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（usage `--json` 1 句） | T007 (Phase 3) + T016 (Phase 4A) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（root 訊息 + 4 退出碼 row） | T008–T012 (Phase 3) + T018 (Phase 4B) | PASS |
| `research.md` -> Decision 1（E2E acceptance path + pure-helper unit tests） | T003、T013、T014 | PASS |
| `research.md` -> Decision 2（adopt `golangci-lint` + `govulncheck` + `.golangci.yml`） | T001、T002 (Setup) | PASS |
| `research.md` -> Decision 3（stdlib `testing`, table-driven） | T003、T013、T014 | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP、`/axb-data-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
