# Tasks: Vertex AI / Gemini Provider Support (round 013)

**Plan Package**: `specs/plans/013-vertex-gemini-provider`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/contracts/**`, `specs/truth/data/**`, `specs/truth/features/**`, `ui/**`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意。
- 本輪 `research.md` 已拍板 Decisions 與 `specs/truth/**` 非 NOOP 項目，必須在 tasks 的 `Read` 或交付目標中被全量涵蓋（Pre-Delivery Orphan Coverage Sweep）。
- 由 `research.md` Decision 衍生的 Setup / Foundational 建置或驗證 task，不強制對應 `truth-delta.md` row。
- **Phase 1 `Setup` 省略**：本輪無新增技術（stdlib-only；`go.mod` 不變），依 SOP 略過 Setup，且不得把 helper／fixture／落點骨架塞進 Setup。
- Phase 2 `Foundational` 只建立後續實作程式、測試共用元件、入口、fixture、helper 與落點骨架；每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation` 在寫產品碼之前，先把本輪所有受影響的自動化測試對齊最新版 truth。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 內相對路徑。
- **本輪有產品碼變更**：新增 `internal/infrastructure/llm/gemini` adapter（Vertex `:generateContent` + service-account OAuth2），並在 `internal/infrastructure/llm/factory.go` 把 `gemini`/`google` 對到該 adapter。相位順序仍為「先對齊測試層（Phase 3），再在 Feature phase 補產品碼（GREEN）」。

---

## Phase 2: Foundational

**Goal**: 建立 Vertex/Gemini adapter 與其單元測試、E2E step、fake provider 的落點骨架與共用元件（Zero Shared Edits 原則），讓 Phase 3 / Phase 4 不各自發明檔或 seam。只建立落點與載體，不寫請求組裝、token 流程、family 對應等產品行為。

- [ ] T001 建立 `gemini` adapter 套件落點骨架
  - Read:
    - `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 1, 2, 3
    - `internal/infrastructure/llm/factory.go`、`internal/infrastructure/llm/openai/client.go`
  - 只做：建立 `internal/infrastructure/llm/gemini/client.go` 與 `auth.go`，各放最小型別／建構子 stub（例如 `New` 與 credential 型別的空殼），實作留空。
  - 不做：不寫請求組裝、回應正規化、JWT 簽章或 token 快取；不改 factory 對應。

- [ ] T002 擴充 E2E fake provider 以服務 Vertex 形狀與 token 交換
  - Read:
    - `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 2, 6
    - `tests/e2e/fakeprovider/fakeprovider.go`
  - 只做：在 fake provider 增加一個可切換的 Vertex 形狀回應（`candidates[0].content.parts` 文本／`functionCall`、`usageMetadata`）與一個 token 端點（回 `access_token`），以 request path 分派（`…/token` → OAuth2 JSON；`…:generateContent` → Vertex JSON；其餘 → 既有 OpenAI handler）；保留既有 OpenAI 形狀與既有 recorder。
  - 不做：不改產品碼；不改既有 OpenAI 場景行為。

- [ ] T003 新增 E2E「gemini provider + service-account key」共用 arrange helper
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider …` / `… whose key file is missing` / `the configured Gemini provider entry allows at most …` / `a configured provider … of a family tellme cannot drive`
    - `tests/e2e/steps/scenario_context.go`
  - 只做：在 `scenario_context.go` 增加 helper，寫出預設 config 的 `gemini` provider 條目（Vertex 形狀 URL 指向 fake、`API_KEY` 為 `.json` key 檔）與 service-account key 檔（其 `token_uri` 指向 fake token 端點）。
  - 不做：不寫 step 斷言邏輯；不碰 `internal/`。

- [ ] T004 建立 7 個新句 stepdef 獨立檔骨架
  - Read:
    - `specs/truth/features/cli/chat/dsl.md`（本輪 7 個新句）
    - `tests/e2e/steps/register.go`
  - 只做：建立 7 個獨立 stepdef 檔（各自 `init()` 自我註冊空白 registrar），對到 T006–T012 七句。
  - 不做：不寫具體斷言／arrange 邏輯；不碰既有 step 檔。

- [ ] T005 建立 `[UNIT]` 落點骨架
  - Read:
    - `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 2, 3, 7
    - `internal/infrastructure/llm/gemini/client.go`、`auth.go`、`internal/infrastructure/llm/factory_test.go`
  - 只做：建立 `internal/infrastructure/llm/gemini/client_test.go`、`auth_test.go`（table 骨架 / fake HTTP + fake token 端點 helper），並預留 `factory_test.go` 的 `gemini`/`google` 對應斷言骨架。
  - 不做：不寫產品碼；不啟用任何斷言。

---

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth；本輪 7 句皆為全新句（`ADD`）。不寫產品行為。

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 為 **NOOP**（本輪 7 句皆 `chat` 模組專屬；frozen class-phrase 詞彙維持 **11**）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`／`權威狀態`）。
- `truth-delta.md` 只告訴這句是 ADD。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：**無**（本輪未 reword 既有句）。
- `[BDD-REMOVE]`：**無**。
- `[BDD-RED]`：本輪 **7 句**（`ADD` 的 DSL rows）。
- `[UNIT]`：Vertex 請求／回應對應、service-account JWT + in-memory token cache、family 對應（非 DSL）。
- 兩者都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md`
  -> `a configured Gemini provider "{provider}" whose endpoint answers with "{answer}"`
  -> `a configured Gemini provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"`
  -> `a configured Gemini provider "{provider}" whose key file is missing`
  -> `the configured Gemini provider entry allows at most {tokens} output tokens`
  -> `a configured provider "{provider}" of a family tellme cannot drive`
  -> `the request to the provider "{provider}" carried the service-account access token`
  -> `the request to the provider "{provider}" allows at most {tokens} output tokens`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/driving-a-vertex-gemini-model.feature`、`chat/authenticating-to-a-vertex-gemini-model.feature`、`chat/refusing-a-provider-family-tellme-cannot-drive.feature`）+ MODIFY（`chat/dsl.md`）；NOOP（root `cli/dsl.md`、`contracts/**`、`data/**`）
- `tests/e2e/steps/`、`tests/e2e/fakeprovider/`

**Boundary**:
- 一條 DSL 一個 task；落入獨立單一檔案（檔案 `init()` 自我註冊）。
- 只改該句的 stepdef／assertion／直接依賴的 helper（T003 的 arrange helper、T002 的 fake）。
- `[UNIT]` 於 `internal/infrastructure/llm/gemini/{client_test.go,auth_test.go}`；採 stdlib `testing` + `net/http/httptest`。
- 不寫產品碼。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T006–T013 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T014 等全部回來再啟動 subagent 執行 review。

### BDD-RED（本輪新增句型）

- [ ] T006 [P] [BDD-RED] `Given: a configured Gemini provider "{provider}" whose endpoint answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider "{provider}" whose endpoint answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t006_config_given_gemini_provider_answers.go`
  - 語意：寫出可解析的預設 config，`SELECTED_PROVIDER` = `{provider}`、`PROVIDERS.{provider}` 為 `gemini` 條目（Vertex 形狀 URL 指向 fake、`API_KEY` 為 service-account `.json` key 檔）；寫出 key 檔（`token_uri` 指向 fake token 端點）；script fake 以 Vertex 形狀回 `{answer}` 並服務 token 交換。

- [ ] T007 [P] [BDD-RED] `Given: a configured Gemini provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider "{provider}" whose endpoint asks tellme to read "{path}" and then answers with "{answer}"`
  - Landing: `tests/e2e/steps/step_t007_config_given_gemini_read_then_answer.go`
  - 語意：如上；script fake 兩步 Vertex 交換（先 `functionCall` 要求 `read_files` 讀 `{path}`，再回 Vertex 答案 `{answer}`）。

- [ ] T008 [P] [BDD-RED] `Given: a configured Gemini provider "{provider}" whose key file is missing`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider "{provider}" whose key file is missing`
  - Landing: `tests/e2e/steps/step_t008_config_given_gemini_key_missing.go`
  - 語意：寫出 config 選 `{provider}` 為 `gemini` 條目，`API_KEY` 為一個不存在（結尾 `.json`）的檔案。

- [ ] T009 [P] [BDD-RED] `Given: the configured Gemini provider entry allows at most {tokens} output tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the configured Gemini provider entry allows at most {tokens} output tokens`
  - Landing: `tests/e2e/steps/step_t009_config_given_gemini_max_tokens.go`
  - 語意：把預設 config 選中 provider 條目的 `MAX_TOKENS` 設為 `{tokens}`。

- [ ] T010 [P] [BDD-RED] `Given: a configured provider "{provider}" of a family tellme cannot drive`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" of a family tellme cannot drive`
  - Landing: `tests/e2e/steps/step_t010_config_given_unsupported_family.go`
  - 語意：寫出可解析的預設 config，選中 `{provider}` 且其條目 `TYPE` 為 tellme 無法驅動的 family（例：`anthropic`）。

- [ ] T011 [P] [BDD-RED] `Then: the request to the provider "{provider}" carried the service-account access token`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request to the provider "{provider}" carried the service-account access token`
  - Landing: `tests/e2e/steps/step_t011_chat_then_service_account_token.go`
  - 語意：fake 對 `{provider}` 恰記錄一次請求，且帶 `Authorization: Bearer <token>`；該 token 由 fake token 端點自 service-account 憑證簽發。

- [ ] T012 [P] [BDD-RED] `Then: the request to the provider "{provider}" allows at most {tokens} output tokens`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request to the provider "{provider}" allows at most {tokens} output tokens`
  - Landing: `tests/e2e/steps/step_t012_chat_then_output_budget.go`
  - 語意：fake 對 `{provider}` 恰記錄一次請求，且其 generation config 上限為 `{tokens}` tokens。

### UNIT（非 DSL 的單元斷言）

- [ ] T013 [P] [UNIT] Vertex 請求／回應對應 + service-account JWT + token cache + family 對應
  - Read:
    - `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 1, 2, 3, 5, 7
    - `specs/truth/techstack.md` -> Reasoning & Provider Transport（Vertex/Gemini adapter；Service-account authentication；Provider family mapping）+ Testing & Verification（Pure-helper unit tests）
    - `internal/infrastructure/llm/gemini/client_test.go`、`auth_test.go`、`internal/infrastructure/llm/factory_test.go`
  - 撰寫：以 `httptest` + 注入的 token 端點驗 — Vertex 請求體（`contents`／`systemInstruction`／`generationConfig`／`functionDeclarations`）與回應正規化（`text`／`functionCall`／`usageMetadata`）；service-account JWT 簽章與 token 取得 + 快取重用；factory 對 `gemini`/`google` 對到 Gemini adapter、其他 family 維持 refuse。鍵於 `NewGateway` 介面（已存在）：以一個只實作 `llm.Gateway` 的 fake，斷言 adapter 存在，或直接測 adapter 的請求組裝。

### Phase Review Gate

- [ ] T014 subagent review (phase quality gate)
  - Read:
    - `specs/truth/features/cli/chat/driving-a-vertex-gemini-model.feature`、`specs/truth/features/cli/chat/authenticating-to-a-vertex-gemini-model.feature`、`specs/truth/features/cli/chat/refusing-a-provider-family-tellme-cannot-drive.feature`
    - `specs/truth/features/cli/chat/dsl.md`、`specs/truth/features/cli/dsl.md`
    - `tests/e2e/steps/*.go`、`tests/e2e/fakeprovider/fakeprovider.go`
    - `internal/infrastructure/llm/gemini/*.go`、`internal/infrastructure/llm/factory.go`
  - 檢驗：無 undefined step；7 個新句 stepdef + `[UNIT]` 皆已註冊；測試可編譯；失敗僅限尚未實作之產品行為。

---

## Phase 4A: ADD Feature File - cli/chat/driving-a-vertex-gemini-model.feature

**Goal**: 讓 `driving-a-vertex-gemini-model.feature` 全綠 — 實作 Vertex/Gemini adapter 的請求組裝與回應正規化，並在 factory 把 `gemini`/`google` 對到它。

**Shared Must Read**:
- `specs/truth/features/cli/chat/driving-a-vertex-gemini-model.feature` -> `Feature: Driving a Vertex Gemini model`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider "…" whose endpoint answers with "…"`、`… asks tellme to read "…" and then answers with "…"`、`the configured Gemini provider entry allows at most …`、`the request to the provider "…" allows at most …`、`the request carried the persona "…"`、`tellme read "…" using its read_files tool`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/driving-a-vertex-gemini-model.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Vertex/Gemini adapter / Provider family mapping）
- `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 1, 2（transport + wire mapping）

**Boundary**:
- 產品碼：`internal/infrastructure/llm/gemini` 實作 `llm.Gateway`：以設定 `URL` 的 project/location/publisher 路徑 + `MODEL` 組 Vertex `:generateContent` 請求（`contents`／`systemInstruction`／`generationConfig`（`maxOutputTokens`、`thinkingConfig`）／`tools[].functionDeclarations`），並正規化 `candidates[0].content.parts`（text + `functionCall`）與 `usageMetadata`。`factory.go` 把 `gemini`/`google` 對到此 adapter；其他 family 維持既有 unsupported 失敗（不 silent 擴張）。
- `internal/cli` 不需變更（已透過 factory seam）。
- OpenAI-family 請求 byte 不變；`stdout` 契約不變。
- **多輪工具交換的 Vertex wire format（review D1）**：把 `AgentLoop` 帶入的 prior messages re-map 成 Vertex `contents` — assistant 的 tool call 用 `role: "model"` + `functionCall`，tool result 用 `role: "user"` + `functionResponse`（Vertex 拒絕 `role: "tool"`）。
- **endpoint 由設定 `URL` 直組（review D2）**：不做 hostname 驗證（以利 loopback fake）；僅驗 path 結構符合 Vertex 佈局。

**Test Scope**:
- `specs/truth/features/cli/chat/driving-a-vertex-gemini-model.feature`

- [ ] T015 [BDD-GREEN] 讓 Test Scope 全綠（並使 T013 的請求／回應與 family 對應 `[UNIT]` 轉綠）
- [ ] T016 [BDD-REFACTOR] 在綠燈下整理 Gemini 請求組裝與回應正規化

## Phase 4B: ADD Feature File - cli/chat/authenticating-to-a-vertex-gemini-model.feature

**Goal**: 讓 `authenticating-to-a-vertex-gemini-model.feature` 全綠 — 實作 service-account OAuth2（JWT-RS256 → token 端點 → `Authorization: Bearer`，in-memory 快取；缺失／不可用 → frozen provider phrase + exit 6）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/authenticating-to-a-vertex-gemini-model.feature` -> `Feature: Authenticating to a Vertex Gemini model`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured Gemini provider "…" whose endpoint answers with "…"`、`a configured Gemini provider "…" whose key file is missing`、`the request to the provider "…" carried the service-account access token`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/authenticating-to-a-vertex-gemini-model.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Service-account authentication / Provider credential resolution）
- `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 3, 4, 5（stdlib OAuth2／credential `.json` + failure taxonomy／in-memory cache）

**Boundary**:
- 產品碼：`internal/infrastructure/llm/gemini/auth.go` 讀 service-account JSON（`client_email`／`private_key`／`token_uri`），以 stdlib `crypto/*` 簽 JWT-RS256，`net/http` POST token 端點取 access token，附 `Authorization: Bearer`，並在 process 內快取重用。缺失／不可讀／非 `.json` → `the provider request failed`（exit 6），**不** silent fallback。
- 只改 gemini auth 路徑；不改 OpenAI-family 與 `-d`／boot（不建 transport）。
- **並發安全 token 快取（review D3）**：以 `sync.RWMutex` + double-checked locking + ~60s 到期安全邊際；到期／401 才重新 mint。

**Test Scope**:
- `specs/truth/features/cli/chat/authenticating-to-a-vertex-gemini-model.feature`

- [ ] T017 [BDD-GREEN] 讓 Test Scope 全綠（並使 T013 的 JWT／token cache `[UNIT]` 轉綠）
- [ ] T018 [BDD-REFACTOR] 在綠燈下整理 service-account 憑證讀取與 token 快取

## Phase 4C: ADD Feature File - cli/chat/refusing-a-provider-family-tellme-cannot-drive.feature

**Goal**: 讓 `refusing-a-provider-family-tellme-cannot-drive.feature` 全綠 — 確認新增 `gemini`/`google` 對應未擴張支援：無法驅動的 family（例 `anthropic`）維持既有 refuse。

**Shared Must Read**:
- `specs/truth/features/cli/chat/refusing-a-provider-family-tellme-cannot-drive.feature` -> `Feature: Refusing a provider family tellme cannot drive`
- `specs/truth/features/cli/chat/dsl.md` -> `a configured provider "{provider}" of a family tellme cannot drive`
- `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/refusing-a-provider-family-tellme-cannot-drive.feature`）+ `/axb-technical-research` MODIFY（`techstack.md` Provider family mapping）
- `specs/plans/013-vertex-gemini-provider/research.md` -> Decision 1, 7（family 對應；不 silent 擴張）

**Boundary**:
- 產品碼：`factory.go` 的 family 對應只新增 `gemini`/`google`；其餘 family 一律走既有 unsupported 分支（`ProviderError`）。
- 不改 OpenAI-family；`stdout`／class-phrase／exit-code 不變。

**Test Scope**:
- `specs/truth/features/cli/chat/refusing-a-provider-family-tellme-cannot-drive.feature`

- [ ] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 family 對應的預設分支

## Phase 4D: Regression

**Goal**: 證明新版 truth 成立且既有行為未被破壞（OpenAI-family 請求 byte 不變、`stdout` byte-exact、exit-code 與 class-phrase 契約、offline paths 不變），並提供可偽性見證。

**Test Scope**:
- `specs/truth/features/cli/**`（chat、history、configuration、workspace、diagnostics、usage 全模組）

- [ ] T021 [REGRESSION] 執行全域回歸 + falsifiability witness
  - 執行 `make verify`（`gofmt`、`go vet`、`staticcheck`、`golangci-lint`、`govulncheck`）與 `go test -count=1 ./...`（含 godog）。
  - **可偽性見證**：暫時讓 factory 把 `gemini` 也視為 unsupported（或讓請求不帶 service-account token），確認對應 gemini E2E 場景失敗；觀察到失敗即還原。
  - 確認：exit-code 表 `0/2/3/4/5/6/7` 與 class-phrase 詞彙維持 **11**；OpenAI-family 請求 byte 不變；`stdout` byte-exact；offline paths（`--version`、`-d`、no-prompt boot、`-l`、prompt-less `--new`）不變；`go mod tidy` 後 module graph 不變（本輪無新相依）；Gherkin/DSL topology audit **PASSED**。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Vertex/Gemini adapter） | T001、T013、T015、T016 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Provider family mapping） | T010、T013、T015、T019、T020 | PASS |
| `specs/truth/techstack.md` -> Reasoning & Provider Transport（Service-account authentication） | T001、T013、T017、T018 | PASS |
| `specs/truth/techstack.md` -> Configuration（Provider entry schema — API_KEY per-family） | T003、T006、T008、T015、T017 | PASS |
| `specs/truth/techstack.md` -> Configuration（Provider credential resolution） | T008、T013、T017、T018 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Local fake provider — Vertex + token exchange） | T002、T006、T007、T011、T021 | PASS |
| `specs/truth/techstack.md` -> Testing & Verification（Pure-helper unit tests） | T005、T013、T021 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`specs/truth/techstack.md`） | T001–T005、T013、T015–T018、T021 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP | 豁免（NOOP 不建任務） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/driving-a-vertex-gemini-model.feature`） | T006、T007、T009、T011、T012、T015、T016 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/authenticating-to-a-vertex-gemini-model.feature`） | T006、T008、T011、T017、T018 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD（`chat/refusing-a-provider-family-tellme-cannot-drive.feature`） | T010、T019、T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY（`chat/dsl.md`） | T003、T004、T006–T014 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（root `cli/dsl.md`） | 豁免（NOOP 不建任務；T021 驗證詞彙維持 11） | PASS |
| `research.md` -> Decision 1（new Vertex/Gemini adapter；factory 對應） | T001、T013、T015、T019 | PASS |
| `research.md` -> Decision 2（Vertex request/response mapping） | T001、T002、T006、T013、T015、T016 | PASS |
| `research.md` -> Decision 3（stdlib service-account OAuth2） | T001、T013、T017、T018 | PASS |
| `research.md` -> Decision 4（`.json` credential + failure taxonomy exit 6） | T008、T013、T017、T018 | PASS |
| `research.md` -> Decision 5（in-memory token cache） | T013、T017、T018 | PASS |
| `research.md` -> Decision 6（hermetic fake + credential `token_uri`） | T002、T003、T006、T011、T021 | PASS |
| `research.md` -> Decision 7（no new module；BDD techstack unchanged） | T021（`go mod tidy` graph 不變） | PASS |
| `research.md` -> Decision 8（must-ask questions settled） | T015–T021（不改 runner/端） | PASS |
| `spec.md` -> US1/US2、`FR-001`–`FR-013`、`NFR-001`–`NFR-006` | T006–T013（對齊）、T015–T020（交付）、T021 | PASS |
| `plan.md` -> Source-code structure（`internal/infrastructure/llm/gemini`、`factory.go`、`tests/e2e/**`） | T001–T005、T013、T015–T020 | PASS |
| `plan.md` -> Scope notes（api/data NOOP；CLI end → /axb-dsl-refine） | T014、T021 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
