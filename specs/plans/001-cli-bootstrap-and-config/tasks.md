# Tasks: CLI Bootstrap and Configuration

**Plan Package**: `specs/plans/001-cli-bootstrap-and-config`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- Phase 1 `Setup` 只做本輪新增技術的基礎建設、技術環境與最後的 smoke-test；不寫 DSL 語意、不寫產品行為。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架。
- Phase 3 `Test Alignment & Implementation` 的目的：在寫產品碼之前，先把本輪所有受影響 DSL 的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

## Phase 1: Setup

**Goal**: 初始化 Go 1.26 模組 `github.com/gosharplite/tellme`，引入 CLI 與測試所需依賴套件 (`github.com/spf13/pflag`, `gopkg.in/yaml.v3`, `github.com/cucumber/godog`)，配置 Makefile 構建、靜態分析與驗證目標（含 `verify-no-test-sleep` 與 `verify` 聚合目標），並以 smoke-test 驗證 Go 測試與 godog 運行環境。不寫 DSL 語意、不寫產品行為。

- [X] T001 初始化 Go 模組並引入核心相依套件
  - Read:
    - `specs/truth/techstack.md` -> CLI Application / Configuration / Testing & Verification
  - 執行 `go mod init github.com/gosharplite/tellme`，將 Go 版本設為 1.26。
  - 引入 `github.com/spf13/pflag`、`gopkg.in/yaml.v3` 與 `github.com/cucumber/godog` 套件並完成 `go mod tidy`。

- [X] T002 建立標準 Makefile 構建、檢驗與驗證目標
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling (Task runner: `make | fmt, tidy, build, test, verify`)
    - `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 4 & 7 (verify-no-test-sleep parity)
  - 建立 `Makefile`，提供 `build`（預設 `VERSION ?= dev` 注入版本至 `main.version`）、`fmt`、`vet`、`staticcheck`、`tidy`、`test`、`verify-no-test-sleep`（grep 掃描測試程式碼禁止使用非同步 `time.Sleep`）與 `verify`（聚合驗證目標：執行 `verify-no-test-sleep` 與依賴圖防護驗證）。
  - Makefile 備註說明：`build` 的 `dev` 預設值僅供本機與正式 release 構建，不可作為 E2E 測試 harness 的構建途徑（harness 嚴格走帶 sentinel 的顯式構建）。

- [X] T003 建立 Go 與 godog runner 初始化 smoke-test
  - Read:
    - `specs/truth/techstack.md` -> Testing & Verification
  - 建立基本測試檔驗證 `go test` 與 `godog.TestSuite` 能夠正常編譯與被 Go 測試工具鏈執行。
  - 不跑本輪 Feature，不實作 stepdef 語意。

## Phase 2: Foundational

**Goal**: 建立 E2E 測試框架入口、Stepdef 登錄機制與場景狀態結構、CLI 子行程執行 helper（帶 sentinel 注入）、退出碼常數定義與成對相異性斷言、產品程式碼進入點骨架，以及離線依賴圖防護，為後續測試與實作提供確定性共用設施。

- [X] T004 建立 E2E 測試套件入口與 Suite 佈線
  - Read:
    - `specs/truth/techstack.md` -> Testing & Verification (E2E runner / step definitions)
    - `specs/plans/001-cli-bootstrap-and-config/plan.md` -> Source-code structure
  - 只做：在 `tests/e2e/suite_test.go` 建立 `godog.TestSuite` 入口（宣告 `TestFeatures`），設置 `ScenarioInitializer` 為 `func(ctx *godog.ScenarioContext) { steps.RegisterAll(ctx) }`，並指定 features 路徑指向 `../../specs/truth/features/cli`。
  - 不做：不實作各步驟的 StepDef 實作語意，不宣告 scenarioContext 實體（由 steps 套件管理）。

- [X] T005 建立 Stepdef 自我註冊機制、ScenarioContext 與執行期掛鉤骨架
  - Read:
    - `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 4 (Runner/step-definition location)
    - `specs/truth/features/cli/dsl.md` -> Given / When / Then 總表
  - 只做：在 `tests/e2e/steps/register.go` 建立 `var registrars []func(*godog.ScenarioContext)` 與 `func RegisterAll(ctx *godog.ScenarioContext)` 走訪執行；建立 `tests/e2e/steps/scenario_context.go` 定義 `scenarioContext` 結構（記錄暫存 `TELL_ME_HOME`、執行參數、環境變數、捕獲的 exitCode/stdout/stderr），並透過 `ctx.Before(...)` 掛鉤在每個場景建立獨立 context 與注入 `context.Context`，於場景結束時清理暫存目錄。
  - 不做：不撰寫各別 DSL 句型邏輯，不實作產品行為。

- [X] T006 建立 CLI 子行程編譯 helper 與 Sentinel 版本注入
  - Read:
    - `specs/truth/techstack.md` -> Testing & Verification (Version assertion: `VERSION=0.0.0-harness`)
    - `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 5 & 6 (E2E black-box harness, version injection)
    - `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme prints the build version`
  - 只做：在 `tests/e2e/harness/cmd_helper.go`（或 `tests/e2e/steps/cmd_helper.go` 獨立 leaf 套件）定義常數 `const SentinelVersion = "0.0.0-harness"`，實作單次編譯 helper 嚴格執行 `go build -ldflags "-X main.version=0.0.0-harness" -o <tmp>/tellme ./cmd/tellme`（禁止走會落入 `dev` 預設值的 `make build`）；提供以指定命令列參數與環境變數執行該 binary、完整收集 stdout、stderr 與 exit code 的黑箱執行 helper。
  - 不做：不斷言 exit code 語意或輸出內容，不寫業務流程。

- [X] T007 定義退出碼常數與成對相異性斷言
  - Read:
    - `specs/plans/001-cli-bootstrap-and-config/spec.md` -> FR-014
    - `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 1 & 7
  - 只做：在 `internal/cli/exitcode.go` 定義並匯出 5 個退出碼常數：`Success` (0)、`UsageError`、`ConfigError`、`EnvironmentError` 以及 `DiagnosticUnresolvedError`；在 `tests/e2e/exitcode_test.go` 撰寫 host-harness 測試 `TestExitCodesAreDistinct`，斷言 4 個非零錯誤碼兩兩互不相同且皆大於 0（非零）。E2E harness 匯入 `internal/cli` 作為預期碼值 oracle，執行步驟則黑箱比對子行程產生的實際退出碼。
  - 不做：不實作參數解析、配置檔案讀取或診斷邏輯。

- [X] T008 建立產品程式碼進入點與核心套件骨架
  - Read:
    - `specs/plans/001-cli-bootstrap-and-config/plan.md` -> Source-code structure
    - `specs/truth/techstack.md` -> Project layout
  - 只做：建立 `cmd/tellme/main.go`（宣告 `var version = "dev"`），以及 `internal/cli`、`internal/config`、`internal/home` 的基本 package 宣告與型別骨架，確保專案能順利被 T006 編譯成 binary。
  - 不做：不實作具體參數解析、YAML 載入、路徑解析或產品行為。

- [X] T009 建立 E2E 離線檢驗與依賴圖防護 helper
  - Read:
    - `specs/truth/techstack.md` -> No-network verification
    - `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 5
  - 只做：在 `tests/e2e/network_guard_test.go` 實作 helper，透過 `go list -deps cmd/tellme` 驗證二進位檔依賴圖未引入 `net` 或 `net/http` 套件，並支援在封鎖網路環境（hostile DNS/proxy）下比對診斷結果。
  - 不做：不實作 `-d` 診斷核心功能，不寫 stepdef。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 truth-delta 有改的句，以及本輪 Feature 用到、尚無 stepdef 的句。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/{module}/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪各句位於介面根 `specs/truth/features/cli/dsl.md` 與四個模組：`configuration/dsl.md`、`workspace/dsl.md`、`diagnostics/dsl.md`、`usage/dsl.md`。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`、`再讀確認`）。
- 路徑正規化約定（Grill #5 Q5）：`{home}` 為 operator 名稱，代表實際 `TELL_ME_HOME`；`{workspace_path}` 以 `{home}` 為根（如 `"ait-tmg/output/butler"`），stepdef 在執行檔案系統檢查前必須將前導 `{home}` 名稱替換為實際暫存目錄路徑。`{config_path}` 則為 home 相對路徑。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY / DELETE。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，但語意是舊 truth。依 `dsl.md` 該列改測試，讓它表達最新版 `StepDef 實作語意`。
- `[BDD-REMOVE]`：`DELETE`。此句已不是 truth。移除或改寫仍綁這句的 stepdef / assertion，不得留下保護舊行為的測試。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成條件：這句非 undefined，且 stepdef body 依 `dsl.md` 該列完整實作其通道（Given/When 的 `怎麼做`/`權威狀態落地`/`回寫`，Then 的 `必查`），執行失敗只能是 assertion 或產品行為，不能是 placeholder 或空實作。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/dsl.md`
- `specs/truth/features/cli/configuration/dsl.md`
- `specs/truth/features/cli/workspace/dsl.md`
- `specs/truth/features/cli/diagnostics/dsl.md`
- `specs/truth/features/cli/usage/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine`
- `tests/e2e/steps/`

**Boundary**:
- 一條 DSL 一個 task，落入獨立單一檔案（45 個獨立檔案），在檔案的 `init()` 中將自我註冊函式加入 `registrars` 列表。
- 零檔案衝突：每個 subagent 只建立/修改自己專屬的 step 檔案，不修改共用檔案。
- 單一 task 自我驗證條件：本句不在 undefined step 清單中，**且** stepdef body 確實依據 `dsl.md` 實作其通道斷言，非無操作佔位符。
- 不寫產品碼。
- review 啟動 subagent；審查 45 個 stepdef 實體是否忠實對齊 `dsl.md` 通道規格，且本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T010–T054 各派一個獨立 subagent；T055 等全部回來再啟動 subagent 來 review。

### Interface Root DSL Steps

- [X] T010 [P] [BDD-RED] `Given: the operator has a runnable tellme installation`
  - Read: `specs/truth/features/cli/dsl.md` -> `the operator has a runnable tellme installation`
  - Landing: `tests/e2e/steps/step_t010_root_given_runnable_installation.go`

- [X] T011 [P] [BDD-RED] `Given: the runtime home is "{home}"`
  - Read: `specs/truth/features/cli/dsl.md` -> `the runtime home is "{home}"`
  - Landing: `tests/e2e/steps/step_t011_root_given_runtime_home.go`

- [X] T012 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}"`
  - Read: `specs/truth/features/cli/dsl.md` -> `a well-formed configuration "{config_path}"`
  - Landing: `tests/e2e/steps/step_t012_root_given_well_formed_config.go`

- [X] T013 [P] [BDD-RED] `When: the operator starts tellme`
  - Read: `specs/truth/features/cli/dsl.md` -> `the operator starts tellme`
  - Landing: `tests/e2e/steps/step_t013_root_when_starts_tellme.go`

- [X] T014 [P] [BDD-RED] `When: the operator starts tellme pointing at the configuration "{config_path}"`
  - Read: `specs/truth/features/cli/dsl.md` -> `the operator starts tellme pointing at the configuration "{config_path}"`
  - Landing: `tests/e2e/steps/step_t014_root_when_starts_with_config.go`

- [X] T015 [P] [BDD-RED] `Then: tellme exits successfully`
  - Read: `specs/truth/features/cli/dsl.md` -> `tellme exits successfully`
  - Landing: `tests/e2e/steps/step_t015_root_then_exits_successfully.go`

- [X] T016 [P] [BDD-RED] `Then: tellme refuses to proceed`
  - Read: `specs/truth/features/cli/dsl.md` -> `tellme refuses to proceed`
  - Landing: `tests/e2e/steps/step_t016_root_then_refuses_to_proceed.go`

- [X] T017 [P] [BDD-RED] `Then: tellme explains on stderr that "{reason}"`
  - Read: `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`
  - Landing: `tests/e2e/steps/step_t017_root_then_explains_stderr.go`

### Configuration Module DSL Steps

- [X] T018 [P] [BDD-RED] `Given: no configuration exists at "{config_path}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `no configuration exists at "{config_path}"`
  - Landing: `tests/e2e/steps/step_t018_config_given_no_config_at_path.go`

- [X] T019 [P] [BDD-RED] `Given: a malformed configuration "{config_path}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a malformed configuration "{config_path}"`
  - Landing: `tests/e2e/steps/step_t019_config_given_malformed_config.go`

- [X] T020 [P] [BDD-RED] `Given: no configuration exists at the default location`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `no configuration exists at the default location`
  - Landing: `tests/e2e/steps/step_t020_config_given_no_default_config.go`

- [X] T021 [P] [BDD-RED] `Given: the effective mode is "{mode}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `the effective mode is "{mode}"`
  - Landing: `tests/e2e/steps/step_t021_config_given_effective_mode.go`

- [X] T022 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" whose selected provider "{provider}" is in its registry`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" whose selected provider "{provider}" is in its registry`
  - Landing: `tests/e2e/steps/step_t022_config_given_selected_provider_in_registry.go`

- [X] T023 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" whose provider registry is empty`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" whose provider registry is empty`
  - Landing: `tests/e2e/steps/step_t023_config_given_provider_registry_empty.go`

- [X] T024 [P] [BDD-RED] `Given: the selected provider override is "{provider}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `the selected provider override is "{provider}"`
  - Landing: `tests/e2e/steps/step_t024_config_given_selected_provider_override.go`

- [X] T025 [P] [BDD-RED] `Then: tellme reports the configuration is ready`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `tellme reports the configuration is ready`
  - Landing: `tests/e2e/steps/step_t025_config_then_reports_ready.go`

- [X] T026 [P] [BDD-RED] `Then: tellme exits with the configuration error code`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `tellme exits with the configuration error code`
  - Landing: `tests/e2e/steps/step_t026_config_then_exits_config_error.go`

### Workspace Module DSL Steps

- [X] T027 [P] [BDD-RED] `Given: no session workspace exists under "{home}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `no session workspace exists under "{home}"`
  - Landing: `tests/e2e/steps/step_t027_workspace_given_no_workspace_under_home.go`

- [X] T028 [P] [BDD-RED] `Given: the session workspace "{workspace_path}" already exists`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the session workspace "{workspace_path}" already exists`
  - Landing: `tests/e2e/steps/step_t028_workspace_given_workspace_already_exists.go`

- [X] T029 [P] [BDD-RED] `Given: the workspace "{workspace_path}" already holds a file "{file_name}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the workspace "{workspace_path}" already holds a file "{file_name}"`
  - Landing: `tests/e2e/steps/step_t029_workspace_given_workspace_holds_file.go`

- [X] T030 [P] [BDD-RED] `Given: the configuration "{config_path}" declares the mode "{mode}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the configuration "{config_path}" declares the mode "{mode}"`
  - Landing: `tests/e2e/steps/step_t030_workspace_given_config_declares_mode.go`

- [X] T031 [P] [BDD-RED] `Given: the mode override is "{mode}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the mode override is "{mode}"`
  - Landing: `tests/e2e/steps/step_t031_workspace_given_mode_override.go`

- [X] T032 [P] [BDD-RED] `Given: the runtime home is not set`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the runtime home is not set`
  - Landing: `tests/e2e/steps/step_t032_workspace_given_runtime_home_not_set.go`

- [X] T033 [P] [BDD-RED] `Given: the workspace path "{workspace_path}" already exists as a regular file`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the workspace path "{workspace_path}" already exists as a regular file`
  - Landing: `tests/e2e/steps/step_t033_workspace_given_workspace_is_regular_file.go`

- [X] T034 [P] [BDD-RED] `Then: tellme creates the session workspace "{workspace_path}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `tellme creates the session workspace "{workspace_path}"`
  - Landing: `tests/e2e/steps/step_t034_workspace_then_creates_workspace.go`

- [X] T035 [P] [BDD-RED] `Then: tellme reports the session workspace "{workspace_path}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `tellme reports the session workspace "{workspace_path}"`
  - Landing: `tests/e2e/steps/step_t035_workspace_then_reports_workspace.go`

- [X] T036 [P] [BDD-RED] `Then: tellme reuses the session workspace "{workspace_path}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `tellme reuses the session workspace "{workspace_path}"`
  - Landing: `tests/e2e/steps/step_t036_workspace_then_reuses_workspace.go`

- [X] T037 [P] [BDD-RED] `Then: the workspace "{workspace_path}" still holds the file "{file_name}"`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `the workspace "{workspace_path}" still holds the file "{file_name}"`
  - Landing: `tests/e2e/steps/step_t037_workspace_then_workspace_still_holds_file.go`

- [X] T038 [P] [BDD-RED] `Then: tellme exits with the environment error code`
  - Read: `specs/truth/features/cli/workspace/dsl.md` -> `tellme exits with the environment error code`
  - Landing: `tests/e2e/steps/step_t038_workspace_then_exits_environment_error.go`

### Diagnostics Module DSL Steps

- [X] T039 [P] [BDD-RED] `Given: a configuration "{config_path}" that does not resolve`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `a configuration "{config_path}" that does not resolve`
  - Landing: `tests/e2e/steps/step_t039_diag_given_unresolved_config.go`

- [X] T040 [P] [BDD-RED] `When: the operator runs tellme with "--version"`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `the operator runs tellme with "--version"`
  - Landing: `tests/e2e/steps/step_t040_diag_when_runs_version.go`

- [X] T041 [P] [BDD-RED] `When: the operator runs tellme's diagnostic`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `the operator runs tellme's diagnostic`
  - Landing: `tests/e2e/steps/step_t041_diag_when_runs_diagnostic.go`

- [X] T042 [P] [BDD-RED] `When: the operator runs tellme's diagnostic with "--json"`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `the operator runs tellme's diagnostic with "--json"`
  - Landing: `tests/e2e/steps/step_t042_diag_when_runs_diagnostic_json.go`

- [X] T043 [P] [BDD-RED] `Then: tellme prints the build version`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme prints the build version`
  - Landing: `tests/e2e/steps/step_t043_diag_then_prints_build_version.go`

- [X] T044 [P] [BDD-RED] `Then: tellme reports the configuration resolved`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme reports the configuration resolved`
  - Landing: `tests/e2e/steps/step_t044_diag_then_reports_config_resolved.go`

- [X] T045 [P] [BDD-RED] `Then: tellme reports the runtime home resolved to "{home}"`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme reports the runtime home resolved to "{home}"`
  - Landing: `tests/e2e/steps/step_t045_diag_then_reports_home_resolved.go`

- [X] T046 [P] [BDD-RED] `Then: tellme reports the session workspace resolved to "{workspace_path}"`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme reports the session workspace resolved to "{workspace_path}"`
  - Landing: `tests/e2e/steps/step_t046_diag_then_reports_workspace_resolved.go`

- [X] T047 [P] [BDD-RED] `Then: tellme reports the configuration did not resolve`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme reports the configuration did not resolve`
  - Landing: `tests/e2e/steps/step_t047_diag_then_reports_config_not_resolved.go`

- [X] T048 [P] [BDD-RED] `Then: tellme reports the reason the configuration did not resolve`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme reports the reason the configuration did not resolve`
  - Landing: `tests/e2e/steps/step_t048_diag_then_reports_unresolved_reason.go`

- [X] T049 [P] [BDD-RED] `Then: tellme emits the resolution status as structured output`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme emits the resolution status as structured output`
  - Landing: `tests/e2e/steps/step_t049_diag_then_emits_structured_resolution.go`

- [X] T050 [P] [BDD-RED] `Then: tellme emits the unresolved status as structured output`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme emits the unresolved status as structured output`
  - Landing: `tests/e2e/steps/step_t050_diag_then_emits_structured_unresolved.go`

- [X] T051 [P] [BDD-RED] `Then: tellme performs no network access`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme performs no network access`
  - Landing: `tests/e2e/steps/step_t051_diag_then_performs_no_network.go`

- [X] T052 [P] [BDD-RED] `Then: tellme exits with the diagnostic error code`
  - Read: `specs/truth/features/cli/diagnostics/dsl.md` -> `tellme exits with the diagnostic error code`
  - Landing: `tests/e2e/steps/step_t052_diag_then_exits_diagnostic_error.go`

### Usage Module DSL Steps

- [X] T053 [P] [BDD-RED] `When: the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"`
  - Read: `specs/truth/features/cli/usage/dsl.md` -> `the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"`
  - Landing: `tests/e2e/steps/step_t053_usage_when_unrecognized_flag.go`

- [X] T054 [P] [BDD-RED] `Then: tellme exits with the usage error code`
  - Read: `specs/truth/features/cli/usage/dsl.md` -> `tellme exits with the usage error code`
  - Landing: `tests/e2e/steps/step_t054_usage_then_exits_usage_error.go`

### Phase Review Gate

- [X] T055 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/**/*.feature`
    - `tests/e2e/steps/*.go`
    - `specs/truth/features/cli/**/dsl.md`
  - 四項完整檢驗標準：
    1. 所有 45 條步驟皆已獨立建立 stepdef 檔案並於 `init()` 完成註冊，無任何 undefined step。
    2. 每個 stepdef body 皆依據 authoritative `dsl.md` 的通道規範（Given/When 的 `怎麼做`/`權威狀態落地`/`回寫`，Then 的 `必查`）實作具體行為與斷言，嚴禁無操作佔位符。
    3. 執行測試時，失敗僅限於 assertion 或產品行為（無語法錯誤、無未初始化 panic）。
    4. 退出碼斷言正確引用 `internal/cli` 導出常數，路徑斷言正確依據 `{home}` 替換為實際暫存目錄。

## Phase 4A: ADD Feature File - cli/configuration/starting-with-a-configuration.feature

**Goal**: 實作 configuration 模組與核心解析邏輯，涵蓋配置載入、YAML 語法校驗、預設路徑探索、effective provider 校驗與專用配置錯誤碼退出，讓 Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature` -> `Feature: Starting tellme with a configuration`
- `specs/truth/features/cli/configuration/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 3 (6-step resolver algorithm)
- `truth-delta.md` -> `/axb-dsl-refine` ADD `starting-with-a-configuration.feature`

**Boundary**:
- 嚴格遵守 Decision 3 順序：無論是否提供 `-c`，`TELL_ME_HOME` 皆優先解析；`${REQ_MODE:-butler}` 僅為預設檔案搜尋之檔機種子，檔案內 `MODE` 與 workspace mode 之差異屬合法非致命發散。
- 只實作配置載入、YAML 格式解析、預設配置探索與 provider registry 校驗，不處理 workspace 目錄建立、診斷輸出或非本 feature 定義的行為。

**Test Scope**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`

- [X] T056 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T057 [BDD-REFACTOR] 在綠燈下重構與整理 configuration 模組實作與 YAML 解析邏輯

## Phase 4B: ADD Feature File - cli/workspace/runtime-home-and-session-workspace.feature

**Goal**: 實作 workspace 模組與執行階段目錄生命週期管理，包含 `TELL_ME_HOME` 判定、`output/<mode>/` 首次建立、重用與既有檔案保存、`TELL_ME_MODE` 優先順序、非目錄衝突防護與環境錯誤碼退出，讓 Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/workspace/runtime-home-and-session-workspace.feature` -> `Feature: Resolving the runtime home and session workspace`
- `specs/truth/features/cli/workspace/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 3 (6-step resolver algorithm)
- `truth-delta.md` -> `/axb-dsl-refine` ADD `runtime-home-and-session-workspace.feature`

**Boundary**:
- 嚴格遵守 Decision 3：`TELL_ME_HOME` 優先解析；目錄建立路徑為 `$TELL_ME_HOME/output/<effective_mode>`；路徑以 `{home}` 為根之 operator path（如 `"ait-tmg/output/butler"`），對應實際路徑為 `<runtime_home>/output/butler`。
- 只實作執行期 home 與 session workspace 的建立、重用、衝突判定與環境錯誤碼，不擴及對話歷史記錄讀寫。

**Test Scope**:
- `specs/truth/features/cli/workspace/runtime-home-and-session-workspace.feature`

- [X] T058 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T059 [BDD-REFACTOR] 在綠燈下重構與整理 workspace 模組實作與目錄權限/生命週期管理邏輯

## Phase 4C: ADD Feature File - cli/diagnostics/version-and-setup-diagnostic.feature

**Goal**: 實作 diagnostics 模組與 `--version`、`-d`、`-d --json` 分流處理，支援純文字與 JSON 結構化輸出（pinned schema）、離線執行保證，以及 unresolved 狀態下的專屬診斷錯誤碼退出，讓 Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` -> `Feature: Checking the build version and diagnosing setup`
- `specs/truth/features/cli/diagnostics/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 3 (resolver reason precedence), Decision 5 (offline witness) & Decision 6 (version sentinel)
- `truth-delta.md` -> `/axb-dsl-refine` ADD `version-and-setup-diagnostic.feature`

**Boundary**:
- 結構化 JSON 輸出嚴格遵守 pinned schema 欄位：`status` ("resolved"/"unresolved")、`reason`（unresolved 時存在）、`runtime_home`、`session_workspace`。
- 未解析原因優先級參照 Decision 3 解析順序；未解析 setup 嚴格輸出專屬 `DiagnosticUnresolvedError` 退出碼。
- 嚴格保證不發起任何網路請求，不實作對話推論。

**Test Scope**:
- `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature`

- [X] T060 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T061 [BDD-REFACTOR] 在綠燈下重構與整理 diagnostics 模組實作與 structured output / offline 檢查邏輯

## Phase 4D: ADD Feature File - cli/usage/unsupported-cli-usage.feature

**Goal**: 實作 CLI 參數解析與未支援 flag 拒絕機制，確保未識別參數在配置解析前即早觸發 usage 錯誤並以專屬 usage 錯誤碼退出，讓 Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature` -> `Feature: Rejecting unsupported command-line usage`
- `specs/truth/features/cli/usage/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `specs/plans/001-cli-bootstrap-and-config/research.md` -> Decision 2 (pflag usage handling)
- `truth-delta.md` -> `/axb-dsl-refine` ADD `unsupported-cli-usage.feature`

**Boundary**:
- 只實作 flag 未識別錯誤的早期偵測與 usage 錯誤處理（退出碼為 `UsageError`），不影響已知合法 flag（`-c`, `-d`, `--json`, `--version`）的傳遞。

**Test Scope**:
- `specs/truth/features/cli/usage/unsupported-cli-usage.feature`

- [X] T062 [BDD-GREEN] 讓 Test Scope 全綠
- [X] T063 [BDD-REFACTOR] 在綠燈下重構與整理 CLI flag 解析與 usage error 分流邏輯

---

## Pre-Delivery Orphan Coverage Sweep

為防止規範、決策與任務產生斷層，在交付 `/axb-implement` 前必須執行 Orphan Coverage Sweep，驗證每條 truth row 與 research decision 皆被 task `Read` 涵蓋或由 task 交付：

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Go 1.26, pflag, yaml.v3, godog | T001 (Setup) | PASS |
| `specs/truth/techstack.md` -> Makefile build, fmt, vet, staticcheck, verify | T002 (Setup) | PASS |
| `specs/truth/techstack.md` -> E2E runner, godog.TestSuite | T004, T005 (Foundational) | PASS |
| `specs/truth/techstack.md` -> Test strategy (E2E black-box) | T006 (Foundational) | PASS |
| `specs/truth/techstack.md` -> Version assertion (`VERSION=0.0.0-harness`) | T006, T043 (Foundational / Phase 3) | PASS |
| `specs/truth/techstack.md` -> No-network verification & build-graph guard | T009, T051 (Foundational / Phase 3) | PASS |
| `specs/truth/techstack.md` -> Determinism (no `time.Sleep` ADR-036) | T002 (Setup `verify-no-test-sleep`) | PASS |
| `specs/truth/features/cli/dsl.md` -> 8 root rows | T010–T017 (Phase 3) | PASS |
| `specs/truth/features/cli/configuration/dsl.md` -> 9 rows | T018–T026 (Phase 3) | PASS |
| `specs/truth/features/cli/workspace/dsl.md` -> 12 rows | T027–T038 (Phase 3) | PASS |
| `specs/truth/features/cli/diagnostics/dsl.md` -> 14 rows | T039–T052 (Phase 3) | PASS |
| `specs/truth/features/cli/usage/dsl.md` -> 2 rows | T053–T054 (Phase 3) | PASS |
| `research.md` -> Decision 1 (cmd/tellme + internal layout) | T008 (Foundational) | PASS |
| `research.md` -> Decision 2 (pflag) | T001, Phase 4D (Setup / Feature) | PASS |
| `research.md` -> Decision 3 (6-step resolver algorithm) | Phase 4A, Phase 4B, Phase 4C (Features) | PASS |
| `research.md` -> Decision 4 (godog runner & harness) | T002, T004, T005 (Setup / Foundational) | PASS |
| `research.md` -> Decision 5 (offline witness & hostile-env fallback) | T009, T051 (Foundational / Phase 3) | PASS |
| `research.md` -> Decision 6 (version injection & sentinel) | T006, T043, Phase 4C (Foundational / Feature) | PASS |
| `research.md` -> Decision 7 (exit-code taxonomy, staticcheck, verify gates) | T002, T007 (Setup / Foundational) | PASS |
