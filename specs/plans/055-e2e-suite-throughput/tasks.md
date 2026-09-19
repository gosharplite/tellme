# Tasks: E2E suite throughput (round 055 — `/axb-tasks`)

**Plan Package**: `specs/plans/055-e2e-suite-throughput`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md` (*Testing & Verification*: E2E runner / Host test harness / Test strategy; *Build & Tooling*: Task runner), `docs/decisions/0024-e2e-suite-throughput.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 test-tooling（developer-throughput）輪，非 BDD feature 輪**：`/axb-spec-by-example`、`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無業務 journey／無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D1/D4；`Concurrency` 已存在於 pinned `godog v0.16.0`；`go.mod`／`go.sum` 不動）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`（`research.md` D3）；驗收由 **`make`／`go test` 的呼叫行為** + 可偽性見證承擔（非 Gherkin）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`tests/e2e/suite_test.go` 的並行預設 + seam** 與 **`Makefile` 的 `test-fast` subset target**；truth 變更（`specs/truth/techstack.md` 四 row）與 **ADR 0024** 已於 plan+truth half 交付，並已記入 `truth-delta.md`。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 0024 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、任何 `internal/**`／`cmd/**` 產品碼、`specs/truth/features/**`、既有 gate（`verify-no-test-sleep`／`verify-no-network`／`vet`／`verify-cross-compile`／`verify-mcp-sdk-confinement`／`verify-architecture`／`lint`／`vulncheck`）的語意、`make test`（= `go test ./...`）的**範圍**。

## Round-055 locked decisions (implementation constraints — MUST)

> 來自 `spec.md`（US1/US2 · FR-001…006 · SC-001…005 · Q1–Q3）、`research.md` D1–D6 與 **ADR 0024** D1–D5。

- **[PARALLEL / Q1 → B, D1]** `tests/e2e/suite_test.go` 的 `godog.Options` 加 `Concurrency: e2eConcurrency()`；`e2eConcurrency()` 讀 `TELL_ME_E2E_CONCURRENCY`（可解析且 ≥ 1 時採用），否則回常數 **`e2eDefaultConcurrency = 4`**。**on by default**（非 opt-in）。
- **[NO-PIN / FR-004, D2]** **不得**把任何 scenario 釘成 serial，**不得**引入靜態 timing-feature 清單；timing scenarios 依既有 bounds 通過，見證為 **N = 5 連續綠**；未來若有 flake，remedy 是 **ADR 0010**（提高**該 test** 的 local margin + 保留 shape 非空轉斷言）。
- **[SUBSET SELECTOR / Q2 → A, D3]** subset 只以 **`godog.paths`** 選取（godog v0.16.0 無 name filter；tags 會改 `specs/truth/**` ⇒ RF-055-1，**不在本輪**）；新 target `test-fast`，預設模組 `configuration history workspace usage diagnostics`（= 43 Examples），`E2E_FAST_MODULES` 可覆寫。
- **[SUBSET BANNER + GUARD / FR-005, D3]** `test-fast` 必須印 `SUBSET — NOT THE GATE` banner（列出選取路徑與 gate 指令），且**在選取等同整份 contract 時拒絕執行**（任一 path 等於 features root，或模組集合等於全部模組）。
- **[GATE SCOPE / FR-002, FR-006, D4]** `make test`（= `go test ./...`）與 gate 指令 `go test -count=1 ./...` **仍跑全部 240 Examples**；`test-fast` **不得**成為 `verify` 成員、**不得**被 `test` 引用；`Strict: true`／`-count=1`／`StopOnFailure`／`Randomize` 不變。
- **[BAR / Q3 → A, D5]** 證據須同時呈現：(i) paired serial baseline（同機同 session）與 (ii) parallel 全量跑；比值 ≤ 60 %；**N = 5** 連續綠；≈25 s 絕對上限僅記於 ADR 0024 作為 backstop，**不**入 gate。
- **[NO-DEP / D4]** 無新相依；`go.mod`／`go.sum` 不動；POSIX-only；`Makefile` 的 hermetic 邊界照常適用於新 target。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D1/D4；`Concurrency` 已在 pinned `godog v0.16.0`；無新相依）。不得把 harness 選項塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點；交付物是既有檔案的**修改**（`tests/e2e/suite_test.go`、`Makefile`），由 Phase 3 直接交付，無新目錄骨架。

## Phase 3: Implementation & Alignment

**Goal**: 讓 godog suite 以並行預設執行（`Concurrency = 4`，可經 `TELL_ME_E2E_CONCURRENCY` 覆寫），並新增 `make test-fast` subset target（banner + never-the-gate guard）—— gate 範圍與契約內容完全不變。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗收語言為「`go test`／`make` 的呼叫行為與輸出」。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Testing & Verification（**E2E runner / step definitions**、**Host test harness**、**Test strategy** rows）+ Build & Tooling（**Task runner** row）
- `docs/decisions/0024-e2e-suite-throughput.md`（D1–D5）
- `docs/decisions/0010-test-deadline-decoupling.md`（timing doctrine）
- `specs/plans/055-e2e-suite-throughput/research.md` -> D1–D6
- `specs/plans/055-e2e-suite-throughput/plan.md` -> The invariants this round introduces (I-1…I-5) + 預期變更
- `tests/e2e/suite_test.go`（現況：`godog.TestSuite`／`Strict: true`）
- `Makefile`（現況：`test: verify-mcp-sdk-confinement` + `go test ./...`；top-of-file hermetic block）

**Boundary**:
- 只改 `tests/e2e/suite_test.go` 與 `Makefile`；**不**碰產品碼、**不**改 `specs/truth/features/**`、**不**改 `make test`／`verify` 的範圍與語意、**不**改 truth。
- `test-fast` **不**得成為 `verify`（或 `test`）的依賴。
- review 啟動 subagent；通過前不解鎖 Phase 4。

- [X] T001 [IMPL] 在 `tests/e2e/suite_test.go` 落並行預設與 seam
  - Read:
    - `specs/plans/055-e2e-suite-throughput/research.md` -> D1, D4
    - `docs/decisions/0024-e2e-suite-throughput.md` -> D1, D4
    - `tests/e2e/suite_test.go`（現況）
  - 做：
    - 新增常數 `e2eDefaultConcurrency = 4` 與 `e2eConcurrency() int`（讀 `TELL_ME_E2E_CONCURRENCY`；`strconv.Atoi` 成功且 ≥ 1 → 採用；否則回常數；附註解指向 ADR 0024）。
    - `godog.Options` 加 `Concurrency: e2eConcurrency()`。
    - **不改**：`Format`、`Paths`、`TestingT`、`Strict: true`。
  - 不做：不改 `Paths`（gate 範圍）；不引入 tags／name filter；不改任何 scenario／feature／stepdef。
  - **驗證**：`go test -count=1 ./tests/e2e/` 綠；`TELL_ME_E2E_CONCURRENCY=1` 與未設時皆綠，且未設時 wall-clock 明顯低於 `=1`。

- [X] T002 [IMPL] 在 `Makefile` 新增 `test-fast` subset target（banner + guard）
  - Read:
    - `specs/plans/055-e2e-suite-throughput/research.md` -> D3
    - `docs/decisions/0024-e2e-suite-throughput.md` -> D3
    - `Makefile`（現況 targets／`test:`／`.PHONY`／top-of-file hermetic block）
  - 做：
    - 新增 `E2E_FAST_MODULES ?= configuration history workspace usage diagnostics`（可覆寫）。
    - 新增 `test-fast` recipe：以 shell 組出 `godog.paths`（`specs/truth/features/cli/<module>` 串接）；若組出的選取等於 features root，**或**模組集合等於 `specs/truth/features/cli/` 下全部模組 → 印錯誤並 `exit 1`（never-the-gate guard）；若某模組目錄不存在 → `exit 1`（fail loud）；否則印 `SUBSET — NOT THE GATE` banner（列選取路徑與 gate 指令 `make test`），再執行 `go test -count=1 ./tests/e2e/ -args -godog.paths="<paths>"`。
    - 把 `test-fast` 加入 `.PHONY`。**不**把 `test-fast` 加進 `verify`／`test`。
  - 不做：不改 `test:`／`verify:` recipe 與其成員；不改 hermetic block；不改任何既有 target 語意。
  - **驗證**：`make test-fast` 綠且印 banner、執行的 Example 數 < 240；`E2E_FAST_MODULES="$(all modules)"` 被 guard 拒絕（exit 1）。

- [X] T003 [REVIEW] 啟動 subagent 對 T001／T002 的 diff 做 review（Phase 3 gate）
  - Read: T001／T002 的 diff；`spec.md` FR-001…006；`research.md` D1–D4；ADR 0024 D1–D4。
  - 做：確認 (i) 並行預設與 seam 如 D1；(ii) 無 scenario 被排除、`Strict`／`Paths` 未動；(iii) `test-fast` 不影響 gate；(iv) 無產品碼／truth 變更。
  - **Red-first 期望**：無——本任務為審查 gate。

## Phase 4: Verification & Evidence

**Goal**: 產出可量測、可偽性見證與 gate 全綠證據；並更新 `STATUS.md`。

- [X] T004 [EVIDENCE] paired baseline（serial）＋ parallel 全量跑，N = 5
  - Read: `research.md` D1/D5；ADR 0024 D1/D5；`spec.md` SC-002/SC-003。
  - 做（同機同 session）：
    - 記 serial baseline：`TELL_ME_E2E_CONCURRENCY=1 go test -count=1 ./tests/e2e/`（時間）。
    - 記 parallel：預設（`Concurrency = 4`）`go test -count=1 ./tests/e2e/` **×5**，全數綠並記時間。
    - 計算比值（≤ 60 % 為門檻）。把數據記入 `research.md`（owner）與 ADR 0024（若需補）。
  - 不做：不以 `test-fast` 取代任何一次全量跑；不調 margin 以求綠。

- [X] T005 [EVIDENCE] subset 行為見證（banner 可見 + guard 拒絕整份 contract）
  - Read: `research.md` D3；ADR 0024 D3；`spec.md` SC-004。
  - 做：
    - `make test-fast` → 綠、印 `SUBSET — NOT THE GATE`、Example 數 < 240（記下實際數，預期 43）。
    - guard 見證：以覆寫把選取設成全部模組 → `exit 1`，訊息說明「that is the gate」；重現後還原。
  - 不做：不為讓 guard 綠而調整判準。

- [X] T006 [EVIDENCE] 可偽性見證（seam 有效 + 未設 seam 時確實並行）
  - Read: `research.md` D1；ADR 0024 D1；`spec.md` SC-001。
  - 做（重現後還原）：
    - (a) `TELL_ME_E2E_CONCURRENCY=1` 的 wall-clock 明顯高於未設時（seam 有效）。
    - (b) 以一個**故意失敗**的臨時情境（或 `-godog.stop-on-failure` 觀察）確認並行下失敗仍被回報（`Strict`／失敗語意不因並行改變）——**不改 committed 檔案**，重現後還原。
  - 不做：不改 committed 檔案；不以 flaky 手段求快。

- [X] T007 [GATE] round 自身 gate 與不變量檢查
  - Read: `spec.md` SC-001/SC-005；`plan.md` I-1…I-5。
  - 做：
    - `gofmt -l .` 乾淨；`go vet ./...` 乾淨。
    - `make verify` **OK**（含 `verify-no-test-sleep`；`verify-architecture` baseline 仍 header-only）。
    - 全量 gate：`go test -count=1 ./...` 綠（含 `tests/e2e`，全 240 Examples）。
    - `go.mod`／`go.sum` 未變（`git diff --stat origin/dev -- go.mod go.sum` 為空）。
    - 確認 `specs/truth/features/**` 未變。
  - 不做：不弱化任何 gate 以求綠。

- [X] T008 [CLOSE] 更新 `STATUS.md` 並交付 PR
  - Read: `STATUS.md`（現況 header／branch model／roadmap／open items）。
  - 做：以最少量更新記 round 055 in flight（branch、PR）、roadmap 一列；commit + push；開 PR（human merge）。
  - 不做：不改 frozen plan packages；不代為 merge。

## Pre-Delivery Orphan Coverage Sweep

**0 orphans.** `truth-delta.md` 的非 NOOP 項目（`techstack.md` 四 row）皆由 T001–T002 的交付或 `research.md` D1–D5 承載，且四 row 的內容是**描述**（非被 task 讀取的輸入）；`research.md` 的已拍板 Decisions D1–D5 全部被 T001–T007 引用；`specs/truth/techstack.md` 本輪異動段落（四 row）皆被 T001/T002 的 `Read` 或 ADR 0024 引用；governance（ADR 0024）由 T001–T002 的 `Read` 引用。無孤立產物。


## Outcome (`/axb-implement`, 2026-09-19)

- **T001** — `tests/e2e/suite_test.go`: `e2eDefaultConcurrency = 4` + `e2eConcurrency()` (the `TELL_ME_E2E_CONCURRENCY` seam) + a harness-owned `-godog.paths` flag (registered on the stdlib flag set in `init()` so `go test -args -godog.paths=…` reaches the suite; godog's `TestSuite`/`Options` path binds no flags — see ADR 0024 D3 note) + `e2ePaths()`. `Strict`/`Paths` default unchanged.
- **T002** — `Makefile`: `test-fast` (+ `E2E_FAST_MODULES`/`E2E_FEATURES_ROOT`/`E2E_FEATURES_REL`); banner + never-the-gate guard + fail-loud missing-module; added to `.PHONY`; **not** a member of `verify`/`test`.
- **T003** — review gate: the diff is confined to `tests/e2e/suite_test.go` + `Makefile`; no product code, no `specs/truth/features/**` change, gate semantics untouched.
- **T004** — paired evidence (same host/session): **serial `tests/e2e` ≈ 50.4–51.0 s → default ≈ 16.66–17.07 s (5/5 green, ~3.0×)**; **paired full gate 55.0 s → 20.7 s (38 %, 2.7×)** ⇒ the ≤ 60 % bar is met. All 240 Examples executed (`-v`: `240 scenarios (240 passed)`).
- **T005** — `make test-fast` green in **~2.0 s**, prints the `SUBSET — NOT THE GATE` banner, runs the 43-Example non-`chat` subset (spot-check `history` alone: 15 scenarios); guard witnesses: all-modules selection **refused**, unknown module **refused (exit 1)**.
- **T006** — falsifiability: (a) the seam is effective (serial `TELL_ME_E2E_CONCURRENCY=1` ≈ 50.7 s vs default ≈ 16.9 s on the same host/session); (b) a bad selection still **fails loudly** under concurrency (`feature path … is not available` → non-zero) — failures are not swallowed by parallelism.
- **T007** — `gofmt -l .` clean; `go vet ./...` clean; `make verify` **OK**; `go test -count=1 ./...` green (25.5 s); `go.mod`/`go.sum` unchanged; `specs/truth/features/**` unchanged.
- **T008** — `STATUS.md` updated (round 055 in flight) + PR opened.

### Note on the selector plumbing (recorded)

`godog v0.16.0`'s `TestSuite`/`Options` path **never binds its command-line flags** (`TestSuite.Run` skips `getDefaultOptions` when `Options != nil`), so godog's own `-godog.paths` cannot reach the suite through `go test -args`. The harness therefore registers the **same flag name/semantics** on the stdlib flag set in `init()`. Effect: `make test-fast` passes a genuine **path-based subset** (godog's documented selector, minus tags/name which are unavailable or truth-writing), and the default (no flag) is the full contract. Recorded in ADR 0024 D3.

## Round-055 fold ledger

| # | Reviewed head | Fold | Note |
| --- | --- | --- | --- |
| 1 | `42f942a` (PR [#120](https://github.com/gosharplite/tellme/pull/120) review [#issuecomment-5737679923](https://github.com/gosharplite/tellme/pull/120#issuecomment-5737679923) — **REQUEST CHANGES**: **B-055-1** + TD-055-1…TD-055-7 + R-A/R-B/R-055-x) | *(this fold commit)* | **B-055-1**: the never-the-gate guard now runs in the **harness on RESOLVED paths** (`guardSelection`/`featureSet`) + a harness banner — **both** entry points obey it (traversal + all-modules caught); **TD-055-1**: the dead `paths = root` clause + the mislabelled empty message replaced by one canonical containment check + an explicit "no modules selected" fail; **TD-055-2**: new `tests/e2e/suite_guard_test.go` pins `e2eConcurrency`/`splitFeaturePaths`/`e2ePaths`/`guardSelection`; **TD-055-3**: `spec.md` FR-005/Q2 now name the shipped `E2E_FAST_MODULES` (and the measured ≈2 s); **TD-055-4**: ADR 0024 reconciled to `research.md` (240 Examples / **1770** steps; ≈**55 s** baseline); **TD-055-5**: the three glued `techstack.md` rows get their `. ` separator; **TD-055-6**: `make help` lists `test-fast`; **TD-055-7**: `STATUS.md` carries the RF-055-x pointer; **R-A**: the `-race` witness recorded in `research.md` D2 (+ RF-055-5); **R-B**: the ratio recorded as a **range (33–47 %)**; **R-055-x**: the flag-namespace/`-args` caveats recorded in ADR D3 + the harness comment (+ RF-055-6) |
| 2 | `27c6b42` (fold-verification [#issuecomment-5737827816](https://github.com/gosharplite/tellme/pull/120#issuecomment-5737827816) — **FOLDS VERIFIED, CLEARED FOR MERGE** + R-055-1…R-055-4) | *(this fold commit)* | **R-055-1**: `guardSelection` now refuses a selection that **covers** the contract (**equality or superset**), not just an equal set — the root-plus-extras and parent `…/features` spellings are refused (`TestGuardSelection` gains the superset row); **R-055-2**: ADR 0024 **D5** records the observed **range (33–47 %)**; **R-055-3**: ADR D3's banner sentence notes it is visible under `-v` / on refusal; **R-055-4**: the `Makefile` pre-refuses a module that resolves to the root, so the refusal precedes the `SUBSET` banner |
| 3 | `1561281` (final verification [#issuecomment-5737871344](https://github.com/gosharplite/tellme/pull/120#issuecomment-5737871344) — **FOLDS VERIFIED — CERTIFIED MERGE-READY**; no blocking items) | *(this doc-only commit)* | **R-055-4b recorded** (informational, per the reviewer's own recommendation): a union/parent spelling (`E2E_FAST_MODULES=".."`) still prints the banner before the harness refuses — the Makefile's per-module pre-check cannot see a whole-set union without duplicating the Go guard in shell (two authorities for one invariant, which is worse); the run **fails**, so the invariant holds and the single authority stays the harness. Doc-only: ADR 0024 D3 + this ledger. |
