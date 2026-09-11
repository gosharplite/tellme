# Tasks: Provider Registry Completeness (round 003)

**Plan Package**: `specs/plans/003-provider-registry-completeness`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 省略（本輪未引入新第三方套件或外部工具鏈，沿用 Go 1.26 與既有工具）。
- Phase 2 `Foundational` 只建立落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 1: Setup

*(省略 — 本輪未引入新的第三方依賴、函式庫或外部工具，變數展開使用 Go 標準庫 `regexp` 與 `os.LookupEnv`，完全沿用現有 Go 1.26 工具鏈與 Makefile 驗證目標。)*

---

## Phase 2: Foundational

**Goal**: 建立變數展開器 `internal/config/expand.go`、單元測試 `internal/config/expand_test.go` 與 8 個 Phase 3 獨立 stepdef 檔案骨架（Zero Shared Edits 原則），為並行分派消除同檔衝突。

- [X] T001 建立變數展開器函式骨架 `internal/config/expand.go`
  - Read:
    - `specs/plans/003-provider-registry-completeness/research.md` -> Decision 2（Environment Variable Expansion Engine）
    - `specs/plans/003-provider-registry-completeness/plan.md` -> Source-code structure
  - 只做：建立 `internal/config/expand.go`，宣告 package `config`，定義 `ExpandString(s string) (string, error)` 與相關 error 常數或變數宣告（回傳 stub `return s, nil`）。
  - 不做：不實作正規表示式解析、不碰 `config.go`、不寫產品邏輯。

- [X] T002 建立變數展開單元測試檔骨架 `internal/config/expand_test.go`
  - Read:
    - `specs/plans/003-provider-registry-completeness/research.md` -> Decision 4（Pure-Helper Unit Testing Architecture）
  - 只做：建立 `internal/config/expand_test.go`（宣告 package `config_test` 或 `config` 與 `testing` 引用）。
  - 不做：不寫測試案例或斷言、不更動既有測試。

- [X] T003 建立 8 個獨立 stepdef 檔案骨架 `tests/e2e/steps/step_t004_*.go` 至 `step_t011_*.go`
  - Read:
    - `specs/truth/features/cli/configuration/dsl.md`（8 個新增 Given 句型）
    - `tests/e2e/steps/register.go`
  - 只做：在 `tests/e2e/steps/` 下建立 8 個獨立檔案（`step_t004_config_given_provider_specifies.go`、`step_t005_config_given_provider_headers.go`、`step_t006_config_given_provider_mandatory_only.go`、`step_t007_config_given_env_var_set.go`、`step_t008_config_given_env_var_unset.go`、`step_t009_config_given_provider_missing_field.go`、`step_t010_config_given_provider_negative_field.go`、`step_t011_config_given_provider_unset_var.go`），各自宣告 `init()` 並將空白 registrar 加入 `registrars` 切片。
  - 不做：不寫具體 step 實作邏輯、不碰既有 step 檔案。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪受影響的自動化測試對齊最新版 truth；包含 `truth-delta.md` 的新增 Given 句型、介面根 class phrase 詞彙驗證，以及 pure-helper 單元測試。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/configuration/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`。
- `truth-delta.md` 指引 ADD / MODIFY。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 stepdef，此時因為產品碼尚未實作完整 schema 與驗證，測試應能編譯但執行紅燈。
- `[BDD-ALIGN]`：`MODIFY`。確認既有 stepdef 符合最新 class phrase 詞彙。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 4）。只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/dsl.md`
- `specs/truth/features/cli/configuration/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine`
- `tests/e2e/steps/`

**Boundary**:
- 一條 DSL 一個 task，落入獨立單一檔案（檔案的 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 任務落入 Phase 2 預留的檔案，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T014 各派一個獨立 subagent；T015 等全部完成後再啟動 subagent 執行 review。

### BDD-RED（configuration 模組新增 Given 句型）

- [X] T004 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" where provider "{provider}" specifies:` (DataTable)
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" where provider "{provider}" specifies:`
  - Landing: `tests/e2e/steps/step_t004_config_given_provider_specifies.go`
  - 語意：依 DataTable（`field`, `value`）寫出 YAML，包含 `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `THINKING_BUDGET`, `THINKING_LEVEL`，並將 `SELECTED_PROVIDER` 預設為 `{provider}`。

- [X] T005 [P] [BDD-RED] `Given: the provider "{provider}" in configuration "{config_path}" includes custom headers:` (DataTable)
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `the provider "{provider}" in configuration "{config_path}" includes custom headers:`
  - Landing: `tests/e2e/steps/step_t005_config_given_provider_headers.go`
  - 語意：依 DataTable（`header`, `value`）在指定 configuration 的 `{provider}` 下更新 `HEADERS` 鍵值對。

- [X] T006 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" where provider "{provider}" specifies only mandatory fields:` (DataTable)
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" where provider "{provider}" specifies only mandatory fields:`
  - Landing: `tests/e2e/steps/step_t006_config_given_provider_mandatory_only.go`
  - 語意：依 DataTable 僅寫入 `TYPE`, `MODEL`, `URL` 三個必填欄位，省略所有可選欄位。

- [X] T007 [P] [BDD-RED] `Given: the environment variable "{var_name}" is set to "{var_value}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `the environment variable "{var_name}" is set to "{var_value}"`
  - Landing: `tests/e2e/steps/step_t007_config_given_env_var_set.go`
  - 語意：在 scenario 的 subprocess 環境中設定 `{var_name}={var_value}`。

- [X] T008 [P] [BDD-RED] `Given: the environment variable "{var_name}" is unset`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `the environment variable "{var_name}" is unset`
  - Landing: `tests/e2e/steps/step_t008_config_given_env_var_unset.go`
  - 語意：在 scenario 的 subprocess 環境中確保 `{var_name}` 被移除。

- [X] T009 [P] [BDD-RED] `Given: a configuration "{config_path}" where selected provider "{provider}" is missing "{field}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a configuration "{config_path}" where selected provider "{provider}" is missing "{field}"`
  - Landing: `tests/e2e/steps/step_t009_config_given_provider_missing_field.go`
  - 語意：寫入缺少 mandatory 欄位 `{field}` 的 provider 設定，並選定該 provider。

- [X] T010 [P] [BDD-RED] `Given: a configuration "{config_path}" where selected provider "{provider}" has negative "{field}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a configuration "{config_path}" where selected provider "{provider}" has negative "{field}"`
  - Landing: `tests/e2e/steps/step_t010_config_given_provider_negative_field.go`
  - 語意：寫入 `{field}` 為負數（如 `-1`）的 provider 設定，並選定該 provider。

- [X] T011 [P] [BDD-RED] `Given: a well-formed configuration "{config_path}" where provider "{provider}" references unset variable "{var_name}"`
  - Read: `specs/truth/features/cli/configuration/dsl.md` -> `a well-formed configuration "{config_path}" where provider "{provider}" references unset variable "{var_name}"`
  - Landing: `tests/e2e/steps/step_t011_config_given_provider_unset_var.go`
  - 語意：寫入 `API_KEY` 為 `${{var_name}}` 且環境中 `{var_name}` 為 unset 的 provider 設定。

### BDD-ALIGN（介面根 class phrase 詞彙對齊）

- [X] T012 [P] [BDD-ALIGN] `Then: tellme explains on stderr that "{reason}"`（介面根；驗證包含 `the provider configuration is invalid`）
  - Read:
    - `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`
    - `tests/e2e/steps/step_t017_root_then_explains_stderr.go`
  - 確認 `step_t017` 前綴比對涵蓋 `tellme: the provider configuration is invalid` 失敗類別。

### UNIT（pure-helper 單元測試）

- [X] T013 [P] [UNIT] `internal/config/expand_test.go` 變數展開表驅動單元測試
  - Read:
    - `specs/plans/003-provider-registry-completeness/research.md` -> Decision 2 & 4
  - 撰寫表驅動測試覆蓋：`${VAR}` 正確展開、`${VAR:-default}` fallback、`${VAR:-}` 空 fallback、多重變數展開、未宣告且無預設之變數拋出錯誤、格式不合規 `${` 拋出錯誤。

- [X] T014 [P] [UNIT] `internal/config/config_test.go` Provider 結構與驗證單元測試
  - Read:
    - `specs/plans/003-provider-registry-completeness/research.md` -> Decision 1, 3 & 4
  - 撰寫表驅動測試覆蓋：完整欄位 YAML 解析、可選欄位省略與零值、必填欄位缺失校驗、負數限制校驗、未選用 provider 之環境變數未設不報錯。

### Phase Review Gate

- [X] T015 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
    - `specs/truth/features/cli/configuration/dsl.md`
    - `specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`
    - `internal/config/*_test.go`
  - 檢驗：無 undefined step；所有 stepdef 註冊完備；測試能編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4: Feature Implementation (MODIFY Feature - Configuration)

**Goal**: 實作完整 `Provider` 結構、`${VAR}` 與 `${VAR:-default}` 展開器以及 active provider 驗證管線，使所有單元測試與 Gherkin 驗收測試轉為全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
- `specs/truth/features/cli/configuration/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine`
- `specs/truth/techstack.md` -> Configuration
- `specs/plans/003-provider-registry-completeness/research.md` -> Decisions 1, 2, 3
- `internal/config/config.go`
- `internal/config/expand.go`

**Boundary**:
- `internal/config/config.go`：擴充 `Provider` struct（`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`）。
- `internal/config/expand.go`：以標準庫 `regexp` 與 `os.LookupEnv` 實作字串展開函式與錯誤定義。
- `internal/config/config.go`：在 `resolve()` 呼叫 provider 驗證與展開，若校驗失敗回傳格式化錯誤，使 CLI exit code 為 `3` 並輸出 `tellme: the provider configuration is invalid: <detail>`。
- 維持 top-level YAML decode 寬容度（不對 root `Config` 開啟 `KnownFields(true)`）。

**Test Scope**:
- `specs/truth/features/cli/configuration/starting-with-a-configuration.feature`
- `internal/config/expand_test.go`
- `internal/config/config_test.go`

- [X] T016 [BDD-GREEN] 實作 Provider 結構、變數展開器與驗證邏輯，使 Test Scope 全綠
  - 實作 `expand.go` 與 `config.go` 中的欄位擴充與校驗邏輯。
  - 執行 `go test ./internal/config/...` 與 godog 驗收測試確認全數通過。

- [X] T017 [BDD-REFACTOR] 在綠燈保護下整理展開器與校驗邏輯
  - 檢視代碼清晰度、錯誤訊息可讀性與函數圈複雜度。
  - 確保無代碼重複且所有 exported/unexported 識別碼符合 Go 慣例。

- [X] T018 [REGRESSION] 執行全域回歸檢驗確認零破壞
  - 執行 `make verify`（含 `gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`、unit tests、e2e tests）。
  - 確認既有 Round 001/002 退出碼與所有 4 個 CLI 模組功能 100% 維持綠燈。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Configuration（`Provider entry schema`） | T001、T014、T016 (Foundational / Phase 3 / Phase 4) | PASS |
| `specs/truth/techstack.md` -> Configuration（`Variable expansion`） | T001、T013、T016 (Foundational / Phase 3 / Phase 4) | PASS |
| `specs/truth/techstack.md` -> Configuration（`Effective-value resolution & validation`） | T014、T016 (Phase 3 / Phase 4) | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（`Pure-helper unit tests`） | T002、T013、T014 (Foundational / Phase 3) | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T016、T018 (Phase 4) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`specs/truth/features/cli/dsl.md`） | T012 (Phase 3) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`specs/truth/features/cli/configuration/dsl.md`） | T003–T011 (Foundational / Phase 3) | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`starting-with-a-configuration.feature`） | T015 (Review) + T016 (Phase 4) | PASS |
| `research.md` -> Decision 1（Provider Entry Schema & Struct Representation） | T001、T014、T016 | PASS |
| `research.md` -> Decision 2（Environment Variable Expansion Engine） | T001、T013、T016 | PASS |
| `research.md` -> Decision 3（Provider Validation Pipeline and Failure Reporting） | T012、T014、T016 | PASS |
| `research.md` -> Decision 4（Pure-Helper Unit Testing Architecture） | T002、T013、T014 | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP、`/axb-data-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
