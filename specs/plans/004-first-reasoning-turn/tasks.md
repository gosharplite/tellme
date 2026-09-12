# Tasks: First Reasoning Turn (round 004)

**Plan Package**: `specs/plans/004-first-reasoning-turn`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- Phase 1 `Setup` 省略（本輪未引入新的第三方依賴）。
- Phase 2 `Foundational` 只建立落點骨架與測試共用元件；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。

---

## Phase 1: Setup

*(省略 — 本輪未引入新的第三方依賴：provider 傳輸使用 Go 標準庫 `net/http`（+ `encoding/json`），E2E fake provider 使用標準庫 `net/http/httptest`，`go.mod` 不變，完全沿用既有 Go 1.26 工具鏈與 Makefile。)*

---

## Phase 2: Foundational

**Goal**: 建立 provider gateway port、OpenAI-compatible adapter 與 E2E fake provider helper 的落點骨架，預留 10 個 chat stepdef 獨立檔案（Zero Shared Edits 原則），並重構 no-network guard 為 offline-path witness。

- [ ] T001 建立 provider gateway port 骨架 `internal/domain/llm/gateway.go`
  - Read:
    - `specs/plans/004-first-reasoning-turn/research.md` -> Decision 1（Provider gateway as a domain port）
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port）
    - `specs/plans/004-first-reasoning-turn/plan.md` -> Source-code structure
  - 只做：宣告 package `llm`；定義 `Gateway` 介面（`Complete(ctx context.Context, req Request) (Response, error)`）與 `Request`、`Response`、`ProviderError` 型別（值型別；`ProviderError` 至少承載 provider 名稱與底層原因）。不含任何 `net/http` import。
  - 不做：不引入 `net/http`、不寫 request assembly、不碰 `internal/cli`。

- [ ] T002 建立 OpenAI-compatible adapter 骨架 `internal/infrastructure/llm/openai/client.go`
  - Read:
    - `specs/plans/004-first-reasoning-turn/research.md` -> Decision 1, 2, 3, 4
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（OpenAI-compatible adapter / HTTP transport / request assembly / response normalization）
  - 只做：宣告 package `openai`；定義 adapter struct（持有 `*http.Client` 與 base URL）並以 stub 實作 `llm.Gateway`；宣告純函式簽名 `buildRequest(...)`（request assembly）與 `parseAnswer(...)`（response normalization），回傳 stub（`buildRequest` 回零值、`parseAnswer` 回 `""`/error）。
  - 不做：不實作 HTTP 呼叫、不寫 request 組裝或回應解析邏輯、不碰 `internal/cli`。

- [ ] T003 建立 E2E fake provider helper `tests/e2e/fakeprovider/fakeprovider.go`
  - Read:
    - `specs/plans/004-first-reasoning-turn/research.md` -> Decision 6
    - `specs/truth/techstack.md` -> Testing & Verification（Local fake provider）
  - 只做：建立以 `net/http/httptest` 為底的 fake provider helper，提供**完整** API（供 Phase 3 直接呼叫，不需再改本檔）：啟動/關閉、`URL()`/`Host()`、設定回應（固定 answer、error status、no usable answer）、記錄收到的請求與請求數、以及寫入「指向該 fake 的預設 configuration」helper（單一 provider 與雙 provider 兩種，可指定 `SELECTED_PROVIDER`）。
  - 不做：不寫任何 stepdef 或斷言；不碰既有 `tests/e2e/harness/`。

- [ ] T004 建立 10 個 chat stepdef 獨立檔案骨架 `tests/e2e/steps/step_t010_*.go`–`step_t019_*.go`
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪新增句型）
    - `tests/e2e/steps/register.go`
  - 只做：在 `tests/e2e/steps/` 下建立 10 個獨立檔案（`step_t010_chat_given_provider_answers.go`、`step_t011_chat_given_two_providers.go`、`step_t012_chat_given_provider_error_status.go`、`step_t013_chat_given_provider_unreachable.go`、`step_t014_chat_given_provider_no_answer.go`、`step_t015_chat_when_starts_with_prompt.go`、`step_t016_chat_when_diagnostic_with_prompt.go`、`step_t017_chat_then_one_request.go`、`step_t018_chat_then_prints_answer.go`、`step_t019_chat_then_provider_error_code.go`），各自宣告 `init()` 並將空白 registrar 加入 `registrars` 切片。
  - 不做：不寫具體 step 實作邏輯、不碰既有 step 檔案。

- [ ] T005 重構 no-network guard 為 offline-path witness
  - Read:
    - `specs/plans/004-first-reasoning-turn/research.md` -> Decision 6（Guard amendment）
    - `specs/truth/techstack.md` -> Testing & Verification（No-network verification (offline paths)）
    - `tests/e2e/harness/network.go`、`tests/e2e/network_guard_test.go`、`Makefile` -> `verify-no-network`
  - 只做：退役 whole-binary capability guard（移除 `NetworkCapabilityViolation`、`networkCapableClosurePackages`、`networkDialPatterns` 的 capability 斷言與 Makefile 對它的委派）；改以 **offline-path witness** 證明 `--version`、`-d`、no-prompt boot 不觸網——no-dial canary（指向 recording sink 的 provider URL 在這些路徑下必須維持 **零連線**）加上既有 differential no-egress sandbox；更新 `verify-no-network` 目標與 `TestDependencyGraphHasNoNetworkCapability` 為 canary-based assertion 並重新命名。
  - 不做：不碰 chat 產品碼、不碰 `tests/e2e/steps/` 的 stepdef。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；含 `truth-delta.md` 有改的 3 句（MODIFY／搬移）與本輪 Feature 用到、尚無 stepdef 的 `chat` 模組 10 句，以及 adapter pure-helper 單元測試。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`，或介面根 `specs/truth/features/cli/dsl.md`。不得掃其他模組。
- 本輪新增句型在 `specs/truth/features/cli/chat/dsl.md`；3 條搬移／改語意的句在介面根 `specs/truth/features/cli/dsl.md`。本輪其它 Feature 用到的既有根句（`the operator has a runnable tellme installation`、`the runtime home is "{home}"`、`the operator starts tellme`、`tellme exits successfully`、`tellme refuses to proceed`）stepdef 已在，不列。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在，依 `dsl.md` 該列對齊精確參照與語意。
- `[BDD-RED]`：`ADD`，或本輪 Feature 用到、尚無 stepdef 的句。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 DSL 的 pure-helper 單元測試（research Decision 1/3/4）。只動測試層。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
- `specs/truth/features/cli/dsl.md`
- `truth-delta.md` -> `/axb-dsl-refine` 有對應 ADD / MODIFY 的句
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案的 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper；不寫產品碼。
- `[UNIT]` 落入 adapter 的獨立 `_test.go`，採 stdlib `testing` 表驅動。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T006–T019 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T020 等全部回來再啟動 subagent 執行 review。

### BDD-ALIGN（介面根 MODIFY／搬移句）

- [ ] T006 [P] [BDD-ALIGN] `Then: tellme explains on stderr that "{reason}"`（介面根；新增第 9 條 class phrase `the provider request failed`）
  - Read:
    - `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`
    - `tests/e2e/steps/step_t017_root_then_explains_stderr.go`
  - 對齊：確認既有前綴比對 stepdef 涵蓋 `tellme: the provider request failed` 失敗類別（片語詞彙由 8 條增為 9 條）。

- [ ] T007 [P] [BDD-ALIGN] `Given: the selected provider override is "{provider}"`（由 configuration 模組搬移為介面根跨模組 row；語意不變）
  - Read:
    - `specs/truth/features/cli/dsl.md` -> `the selected provider override is "{provider}"`
    - `tests/e2e/steps/step_t024_config_given_selected_provider_override.go`
  - 對齊：確認既有 stepdef 精確對應介面根那列，configuration 與 chat 兩模組皆可命中（單一權威，不得重複）。

- [ ] T008 [P] [BDD-ALIGN] `Then: tellme performs no network access`（由 diagnostics 模組搬移為介面根跨模組 row，語意改為 offline-path canary）
  - Read:
    - `specs/truth/features/cli/dsl.md` -> `tellme performs no network access`
    - `tests/e2e/steps/step_t051_diag_then_performs_no_network.go`
    - `tests/e2e/harness/network.go`
  - 對齊：把 stepdef 由 whole-binary capability guard 改為 offline-path witness（recording sink 零連線 + differential）；沿用 T005 的 harness API。

### UNIT（adapter pure-helper 單元測試）

- [ ] T009 [P] [UNIT] `internal/infrastructure/llm/openai/*_test.go` request assembly 與 response normalization
  - Read:
    - `specs/plans/004-first-reasoning-turn/research.md` -> Decision 1, 3 & 4
  - 撰寫表驅動測試覆蓋：endpoint 由 base URL 加上 `/chat/completions`、`Authorization: Bearer`、`MODEL`、`MAX_TOKENS`（>0 才帶）、`HEADERS` 合併；response 取 `choices[0].message.content`，空／缺欄／不可解析一律回 error。

### BDD-RED（chat 模組新增句型）

- [ ] T010 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t010_chat_given_provider_answers.go`
  - 語意：寫入預設 configuration 選定 `{provider}`，其 endpoint 指向 fake provider，並 script fake 回覆 `{answer}`。

- [ ] T011 [P] [BDD-RED] `Given: configured providers "{provider_a}" and "{provider_b}" whose endpoints answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `configured providers "{provider_a}" and "{provider_b}" whose endpoints answer`
  - Landing: `tests/e2e/steps/step_t011_chat_given_two_providers.go`
  - 語意：寫入含兩個 provider、皆指向 fake 的預設 configuration。

- [ ] T012 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with an error status`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with an error status`
  - Landing: `tests/e2e/steps/step_t012_chat_given_provider_error_status.go`
  - 語意：寫入選定 `{provider}` 的預設 configuration，並 script fake 回非 2xx（如 `500`）。

- [ ] T013 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint is unreachable`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint is unreachable`
  - Landing: `tests/e2e/steps/step_t013_chat_given_provider_unreachable.go`
  - 語意：寫入選定 `{provider}` 的預設 configuration，endpoint 指向關閉的埠／不可達位址。

- [ ] T014 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint answers with no usable answer`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" whose endpoint answers with no usable answer`
  - Landing: `tests/e2e/steps/step_t014_chat_given_provider_no_answer.go`
  - 語意：寫入選定 `{provider}` 的預設 configuration，並 script fake 回不可解讀的 body（如空 `choices`）。

- [ ] T015 [P] [BDD-RED] `When: the operator starts tellme with the prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator starts tellme with the prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t015_chat_when_starts_with_prompt.go`
  - 語意：在當前環境下執行 `tellme "{prompt}"`（無 `-c`），擷取 exit code / stdout / stderr 與 fake 收到的請求。

- [ ] T016 [P] [BDD-RED] `When: the operator runs tellme's diagnostic with the prompt "{prompt}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the operator runs tellme's diagnostic with the prompt "{prompt}"`
  - Landing: `tests/e2e/steps/step_t016_chat_when_diagnostic_with_prompt.go`
  - 語意：執行 `tellme -d "{prompt}"`；`-d` dispatch 先於 prompt turn，產生診斷報告。

- [ ] T017 [P] [BDD-RED] `Then: tellme sends exactly one request to the provider "{provider}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme sends exactly one request to the provider "{provider}"`
  - Landing: `tests/e2e/steps/step_t017_chat_then_one_request.go`
  - 語意：斷言 `{provider}` 的 fake 恰好收到 1 次請求，且攜帶 prompt 與解析後的 model/credential；無其它 provider 收到請求。

- [ ] T018 [P] [BDD-RED] `Then: tellme prints the provider's answer "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme prints the provider's answer "{answer}"`
  - Landing: `tests/e2e/steps/step_t018_chat_then_prints_answer.go`
  - 語意：斷言 stdout 含有 provider 的 answer `{answer}`，且未被吞掉或只出現在 stderr。

- [ ] T019 [P] [BDD-RED] `Then: tellme exits with the provider error code`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `tellme exits with the provider error code`
  - Landing: `tests/e2e/steps/step_t019_chat_then_provider_error_code.go`
  - 語意：斷言 exit code 等於 pin 定的 provider error code `6`，且不得退化為其它錯誤類別。

### Phase Review Gate

- [ ] T020 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/answering-a-single-prompt.feature`
    - `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature`
    - `specs/truth/features/cli/chat/dsl.md`
    - `specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/fakeprovider/`
    - `internal/infrastructure/llm/openai/*_test.go`
  - 檢驗：無 undefined step；10 個新 stepdef 皆已註冊；測試能編譯；失敗僅限於尚未實作之產品碼。

---

## Phase 4A: ADD Feature File - cli/chat/answering-a-single-prompt.feature

**Goal**: 實作 provider gateway（domain port + OpenAI-compatible adapter）與 CLI prompt-turn dispatch，使 chat 的成功 Rule 與 offline-path Rule 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/answering-a-single-prompt.feature` -> `Feature: Answering a single prompt`
- `specs/truth/features/cli/chat/dsl.md` -> Given / When / Then（configured provider、starts tellme with the prompt、sends exactly one request、prints the provider's answer）
- `specs/truth/features/cli/dsl.md` -> `tellme performs no network access`、`the selected provider override is "{provider}"`、`tellme exits successfully`
- `truth-delta.md` -> `/axb-dsl-refine`（chat ADD + root MODIFY）
- `specs/truth/techstack.md` -> Reasoning & Provider Transport
- `specs/plans/004-first-reasoning-turn/research.md` -> Decision 1, 2, 3, 4 & 7
- `internal/cli/cli.go`、`internal/domain/llm/gateway.go`、`internal/infrastructure/llm/openai/client.go`

**Boundary**:
- 實作 `internal/domain/llm`（`Gateway` port 與 `Request`/`Response`/`ProviderError` value types；network-free）。
- 實作 `internal/infrastructure/llm/openai`（stdlib `net/http`；`buildRequest`、`parseAnswer`、`Complete`）。
- `internal/cli`：讀取 positional prompt（`fs.Args()` 首個），維持 dispatch 順序 `--version` → `-d` → prompt turn → boot；以 `resolution.Provider` 組請求、印出 answer；no-prompt boot 與 `-d` 皆不觸網。
- 不引入 provider SDK、不加入任何新依賴；不實作 provider failure 映射（屬 Phase 4B）。

**Test Scope**:
- `specs/truth/features/cli/chat/answering-a-single-prompt.feature`

- [ ] T021 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T022 [BDD-REFACTOR] 在綠燈下整理 request assembly、response normalization 與 CLI dispatch

## Phase 4B: ADD Feature File - cli/chat/reporting-a-failed-provider-request.feature

**Goal**: 實作 provider/transport failure 映射，使 chat 的失敗 Rule 全綠，並在綠燈保護下重構。

**Shared Must Read**:
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` -> `Feature: Reporting a failed provider request`
- `specs/truth/features/cli/chat/dsl.md` -> Then（`tellme exits with the provider error code`）
- `specs/truth/features/cli/dsl.md` -> `tellme explains on stderr that "{reason}"`、`tellme refuses to proceed`
- `truth-delta.md` -> `/axb-dsl-refine`（root MODIFY：第 9 條 class phrase）
- `specs/plans/004-first-reasoning-turn/research.md` -> Decision 5
- `internal/cli/cli.go`、`internal/cli/exitcode.go`、`internal/infrastructure/llm/openai/client.go`

**Boundary**:
- 將 `llm.ProviderError`（transport / 非 2xx / 不可解讀 body）映射為單一 stderr 行 `tellme: the provider request failed: <detail>` 與 exit code `6`；`internal/cli/exitcode.go` 擴充 `ProviderError = 6`，維持 `0/2/3/4/5` 不變。
- 不得把 provider failure 併入 configuration class（`3`）。

**Test Scope**:
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature`

- [ ] T023 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T024 [BDD-REFACTOR] 在綠燈下整理 failure 映射與 exit-code 分類

## Phase 4C: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、configuration、workspace、diagnostics、usage 全模組）

- [ ] T025 [REGRESSION] 執行全域回歸，確認零破壞
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - 確認 exit-code 表 `0/2/3/4/5/6` 與既有 CLI 契約（round 001/002/003）100% 維持綠燈，且 offline paths 不觸網。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider gateway port / OpenAI-compatible adapter / HTTP transport / Request assembly / Response normalization） | T001、T002、T009、T021（Foundational / Phase 3 / Phase 4A） | PASS |
| `specs/truth/techstack.md` -> CLI Application（Project layout、Prompt input） | T001、T002、T021（Phase 2 / Phase 4A） | PASS |
| `specs/truth/techstack.md` -> Configuration（Effective-value resolution & validation） | T021（Phase 4A，消費 `resolution.Provider`） | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Local fake provider / No-network verification (offline paths) / Pure-helper unit tests） | T003、T005、T008、T009（Foundational / Phase 3） | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（provider SDK / Gemini·Anthropic adapter / streaming / Thought） | 負向決策 -> T021、T023 Boundary（不引入 SDK／串流／Thought） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY ×5（`specs/truth/techstack.md`） | T001–T005、T009、T021、T023 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD ×3（`chat/dsl.md` + 2 chat features） | T004、T010–T020、T021–T024 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY ×3（`cli/dsl.md`、`configuration/dsl.md`、`diagnostics/dsl.md`） | T006、T007、T008（+ T004、T010–T019 用到根 class phrase） | PASS |
| `research.md` -> Decision 1（provider gateway port + adapter） | T001、T002、T021 | PASS |
| `research.md` -> Decision 2（stdlib `net/http`，無 SDK） | T002、T021 | PASS |
| `research.md` -> Decision 3（request assembly） | T002、T009、T021 | PASS |
| `research.md` -> Decision 4（response normalization） | T002、T009、T021 | PASS |
| `research.md` -> Decision 5（deterministic failure contract） | T023 | PASS |
| `research.md` -> Decision 6（local fake provider + guard amendment） | T003、T005、T008、T010–T014 | PASS |
| `research.md` -> Decision 7（prompt handling & precedence） | T015、T016、T021 | PASS |
| `truth-delta.md` -> `/axb-api-plan` = NOOP、`/axb-data-plan` = NOOP | 豁免（NOOP 不建任務） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
