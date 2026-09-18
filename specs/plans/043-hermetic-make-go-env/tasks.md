# Tasks: a hermetic `make` Go-toolchain environment (round 043 — `/axb-tasks`)

**Plan Package**: `specs/plans/043-hermetic-make-go-env`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0012-hermetic-make-go-env.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 build/quality-pipeline（tooling）輪，非 BDD feature 輪**：`/axb-spec-by-example`、`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無業務 journey／無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D10；僅 `make` + Go toolchain，`go.mod`／`go.sum` 不動）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`；驗收由 **`make` 的 hostile-env 呼叫行為** + 可偽性見證承擔（非 Gherkin，`spec.md` A2 / `research.md` D9）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`Makefile` 的 hermetic invocation boundary（一個 top-of-file `export`/`unexport` block）**；truth 變更（`specs/truth/techstack.md` 的 **Hermetic toolchain invocation** row + **Task runner** row）與 **ADR 0012** 已由 `/axb-technical-research` 於 plan+truth half 交付（PR #97），並已記入 `truth-delta.md`。
  - **本輪不新增 Makefile target**（`plan.md` Structure Decision）：既有 targets 只改「如何啟動 toolchain」；`verify` aggregate 名單不變。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 0012 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 gate（`verify-no-test-sleep`／`verify-no-network`／`vet`／`verify-cross-compile`／`verify-mcp-sdk-confinement`／`verify-architecture`／`lint`／`vulncheck`）的語意、任何 `internal/**`／`cmd/**`／`tests/**` 產品碼。

## Round-043 locked decisions (implementation constraints — MUST)

> 來自 `research.md` D1–D12、`spec.md`（FR-001–013 / NFR-001–004 / SC-001–005，Q1–Q8）與 **ADR 0012**（D1–D9）。

- **[MECH / Q2 → A1, D1]** neutralisation 只定義**一次**：`Makefile` 頂部（**在 `$(shell command -v …)` probes 之上**、在任何 recipe 之前）的一個 `export`/`unexport` block。這使**每個** recipe、以及 scripted recipe 內**巢狀 spawn** 的 `go`（`make test` 的 E2E harness、`verify-no-network` 的見證）都繼承到已消毒環境；**不得**改成 per-recipe `$(HERMETIC_ENV)` 前綴。
- **[NEUTRALISE / Q3 → V1, D2（criterion-derived；PR #97 review B-4）]** **drop/neutralise**（依 ADR D2 inclusion criterion）：`GOENV` → **`off`**（**load-bearing** —— Go 對「unset **或** empty」都會 fallback 到 env **file**，故單純 unset `GOFLAGS` **無法**中和 persisted `go env -w GOFLAGS=-mod=vendor`；round-042 F-2 / ADR 0011 D5）；`GOWORK` → `off`；`GOFLAGS`／`GO111MODULE`／`GOEXPERIMENT`／`GOTOOLCHAIN`／`GOFIPS140`／`GODEBUG` → **unset**；ambient **target triple** —— `GOOS`／`GOARCH` **及** micro-arch family（`GOARM`／`GOARM64`／`GOAMD64`／`GO386`／`GOMIPS`／`GOMIPS64`／`GOPPC64`／`GORISCV64`／`GOWASM`）→ **unset**（host-native）。
- **[PRESERVE / Q3 → V1, D3]** 原樣傳遞：`PATH`／`HOME`／`GOPATH`／`GOMODCACHE`／`GOCACHE`（warm-cache／runtime set）**與** `GOPROXY`／`GOSUMDB`／`GOPRIVATE`／`GONOSUMDB`／`GOINSECURE`（network／checksum set）—— cold-cache 或 proxy 環境仍可解析（通用化 `verify-cross-compile` 的 PR #46 TD1 教訓）。
- **[CGO / Q3 → V1, D4（R-1(1) 校正）]** `CGO_ENABLED` **preserved from the caller**（block 既不 set 也不 unset —— **非**全域 pin）；**唯一**的 cgo pin 仍是 `verify-cross-compile` recipe 內的**行內** `CGO_ENABLED=0`（**僅作用於該 recipe**）。block **不得**覆蓋 recipe 的行內 per-target 賦值（recipe 的行內賦值對該 recipe 勝出，FR-012）。
- **[SCOPE / Q4 → S1, D5]** **無條件、全 targets**（含 mutating 的 `fmt`／`tidy`）；無 per-target opt-out。
- **[OWNERSHIP / Q6 → D1, D6（B-2/B-3 校正）]** `Makefile` block 為**唯一主要 owner**；round-042 的 `tools/arch/arch_test.go` `childEnv`（ADR 0011 D5）**保留**為 **defence-in-depth**，只使 gate 的**verdict** 在其**直接**呼叫路徑（繞過 `make`）保持 hermetic —— 該路徑的**外層** `go test`/`go vet` 仍非 hermetic（R1）。兩處**互相 cross-reference 註解**；兩者為**互補機制**（block：關 env file + unset；`childEnv`：重設顯式值），故其不變量為 **coverage**（每個 neutralised 名已被 `childEnv` 重設、或被記錄為已知未覆蓋名 —— 如 frozen round-042 下的 `GOARM`/`GOEXPERIMENT`，R4），**非**集合相等（見 T005(c)）。
- **[VERIFY / D7（B-1 校正）, Q7 → R2]** 驗收證據 = (a) **四個 red→green 見證**（`GOENV=<file: GOFLAGS=-mod=vendor>`、`GO111MODULE=off`、stray `GOWORK=<path>`、`GOTOOLCHAIN=go1.99.9`；**red before / green after**，重現後還原，ADR 0010 doctrine）**＋ neutralisation assertions**（green-before 的 `GOFLAGS=-trimpath` 與長尾名 —— 斷言其**不在 recipe env** 中；**不**計為見證）**＋ aggregate-level**（`GOENV=<file> make verify` ⇒ exit 0，全員綠）+ (b) **兩個 positive control**：`make verify-cross-compile` 在 hermetic env 下仍 **4/4**（證明只移除 ambient 輸入、未覆蓋行內 per-target 覆寫）；clean env 下 `make tidy`／`make fmt` **no-diff**、`go.mod`／`go.sum` 不變（證明未破壞 mutating targets）。
- **[RESIDUALS / Q7 → R1+R3；PR #97 review → R4+R5, D8]** 已於 plan half 寫入 ADR 0012／truth row：**(R1)** 繞過 `make` 的裸 `go`（含 gate 直接呼叫路徑的**外層** `go test`/`go vet`）仍非 hermetic（hermeticity 是 `make`-boundary 性質）；**(R3)** escape hatch 為**逐變數類別**（已 export 名靠 command-line assignment；plumbing 名靠顯式 export；unexported 名只能直接呼叫或改 block）；**(R4)** frozen round-042 guard 內 `childEnv` **drop 但不 re-set** `GOARM`/`GOEXPERIMENT`（及本輪新增名），故直接路徑上 env-file 值仍可達其 child `go list` —— 記錄、不修；**(R5)** parse-time `$(shell …)`/`$(eval …)`/command-line 變數在 boundary 之外（keep `go` out of `$(shell …)`）。
- **[GOVERNANCE / Q5 → G1, D6]** rule 已記於 **ADR 0012** + `specs/truth/techstack.md`（Build & Tooling 兩 row）；本輪**不**再改 truth，只交付 `Makefile` block（+ 選配的一行 cross-reference 註解）。
- **[ORDERING / Q8 → O1]** 本輪 = `043-*`；**R2 of [#92](https://github.com/gosharplite/tellme/issues/92) 順延為 `044-*`**（`STATUS.md` roadmap 已於 plan half 更新）。
- **[NO-DEP / D10]** 僅 `make` + Go toolchain；**無新相依**；`go.mod`／`go.sum` 不動；POSIX-only。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D10；無新相依）。不得把 hermetic block 塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點；交付物為 `Makefile` 內一段既有檔案的**修改**（由 T002 直接交付），無新目錄骨架。

## Phase 3: Implementation & Alignment (RED-first)

**Goal**: 在 `Makefile` 頂部落一個 hermetic invocation boundary，使**每個** target（含巢狀 spawn）都以消毒後的環境啟動 Go toolchain —— hostile ambient／persisted Go env 不再能讓 `verify` aggregate 非因樹而紅；clean env 行為完全不變。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗收語言為「在 hostile／clean 環境下的 `make` 呼叫」。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Build & Tooling（**Hermetic toolchain invocation** row + **Task runner** row）
- `docs/decisions/0012-hermetic-make-go-env.md`（D1–D9）
- `specs/plans/043-hermetic-make-go-env/research.md` -> D1–D12
- `specs/plans/043-hermetic-make-go-env/plan.md` -> The boundary the round installs + 預期變更（`Makefile` CHANGED）
- `Makefile`（現況：`VERSION ?= dev`、`$(shell command -v …)` probes、`test:`／`verify:` recipe 樣式）
- `tools/arch/arch_test.go` -> `droppedBuildEnv` + `childEnv`（round-042 的 child-env 紀律 —— cross-reference 對象，**行為不得改**）

**Boundary**:
- 只改 `Makefile`（新增 block，**不新增 target**）；可選地在 `tools/arch/arch_test.go` 加**一行註解**（cross-reference；**零行為變更**）；**不**碰任何產品碼、**不**改既有 gate 語意、**不**改 truth。
- review 啟動 subagent；通過前不解鎖 Phase 4。

- [X] T001 [WITNESS-RED] 先行擷取 hostile-env 的 RED 基線（未修復的樹上）
  - Read:
    - `specs/plans/043-hermetic-make-go-env/research.md` -> D4, D7
    - `docs/decisions/0010-test-deadline-decoupling.md`（見證 doctrine）；`docs/decisions/0011-layer-discipline-gate.md` -> D5（child-env 紀律）
    - `Makefile`（現況 recipe）
  - 做（**不改任何檔案**，只觀察並記錄；RED-first 證明非空轉）：
    - 以暫存檔建立 **4 個 RED 見證** 的 hostile ambient 環境，各對一個**便宜**的 toolchain target 取值（`vet`；`GOENV` 案另加 `verify-architecture` 對照）：
      1. `printf 'GOFLAGS=-mod=vendor\n' > /tmp/goenv043; GOENV=/tmp/goenv043 make vet`（及 `… make verify-architecture`）
      2. `GO111MODULE=off make vet`
      3. `printf 'go 1.26\n' > /tmp/gowork043; GOWORK=/tmp/gowork043 make vet`
      4. `GOTOOLCHAIN=go1.99.9 make vet`
    - 另記錄 **green-before** 輸入（`GOFLAGS=-trimpath`）—— 它**不**是真見證（B-1），其非空轉性由 T005 的 `make`-level neutralisation assertion 承擔。
    - **期望 RED**：四案皆非零失敗（`GOENV` → *inconsistent vendoring*；`GO111MODULE=off` → 解析失敗；`GOWORK` → *does not contain modules listed in go.work*；`GOTOOLCHAIN` → *toolchain not available*）。
    - 把四案（及 `-trimpath`）的原始輸出與 exit code 記入本任務（供 Phase 4 對照 `green after`）。
  - 不做：不為讓它綠而改樹；不建立任何 committed 見證檔（本輪**不新增 artifact** —— `plan.md`）。
  - **Red-first 期望**：`GOENV` 案 **FAIL**（exit ≠ 0）；此為 T002 的對照基線。

- [X] T002 [MECH] 落 `Makefile` 的 hermetic `export`/`unexport` block（criterion-derived 全集合）
  - Read:
    - `specs/plans/043-hermetic-make-go-env/research.md` -> D1, D2, D3, D4, D5, D10
    - `docs/decisions/0012-hermetic-make-go-env.md` -> D1, D2, D3, D4, D5
    - `specs/truth/techstack.md` -> Build & Tooling（Hermetic toolchain invocation row）
    - `Makefile`（`VERSION ?= dev` 之下、`STATICCHECK := $(shell …)` **之上**的位置）
  - 做（在 `Makefile` 頂部、`VERSION ?= dev` 之後、`$(shell command -v …)` probes 之前，加一段帶註解的 block）：
    - `export GOENV := off`（load-bearing；註解說明 env-file fallback 對 unset **與** empty 都生效 —— round-042 F-2）
    - `export GOWORK := off`
    - `unexport GOFLAGS GO111MODULE GOEXPERIMENT GOTOOLCHAIN GOFIPS140 GODEBUG` + `unexport GOOS GOARCH GOARM GOARM64 GOAMD64 GO386 GOMIPS GOMIPS64 GOPPC64 GORISCV64 GOWASM`（ADR D2 criterion-derived；僅 **ambient** build-context 被清除；host 維持 native）
    - 一段區塊註解：**目的**（#96；generalises ADR 0011 D5 + round-020 TD1）、**neutralise set**、**preserve set**（`PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` + `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE` —— **不**被 touch）、**`CGO_ENABLED` preserved from the caller**（非全域 pin）且 recipe 行內賦值仍勝出、**boundary 範圍**（recipes + descendants；parse-time `$(shell …)` 在外）、**coverage** 不變量（非 equality）、**cross-reference** 指向 `tools/arch` 的 `childEnv`（defence-in-depth；verdict-only）與 **ADR 0012**。
    - 確認 `$(shell command -v …)` probes 仍能解析 `PATH`（block 在其上、且不 touch `PATH`）。
  - 驗證：`gofmt` n/a（Makefile）；clean env 下 `make help` 正常；T005 的 **4 案**（`GOENV`／`GO111MODULE`／`GOWORK`／`GOTOOLCHAIN`）由 **FAIL → PASS**（Phase 4 逐一對照）。
  - 不做：不新增 target；不改 `verify` aggregate 名單；不 touch `CGO_ENABLED`（全域）；不覆蓋 `verify-cross-compile` 的行內 per-target 賦值；不改任何 recipe 指令本身。

- [X] T003 [DOC] `tools/arch/arch_test.go` cross-reference 註解（**FR-008 MUST**；僅註解、**零行為變更**）
  - Read:
    - `docs/decisions/0012-hermetic-make-go-env.md` -> D6（ownership: primary = Makefile block；childEnv = defence-in-depth）
    - `tools/arch/arch_test.go` -> `childEnv` / `droppedBuildEnv` 上方的註解塊
  - 做：在 `droppedBuildEnv` 註解處寫明——*the `Makefile` hermetic block (ADR 0012) is the primary owner; this filter keeps **this gate's VERDICT** hermetic on the documented **direct** invocation, and does **not** make that path's outer `go test`/`go vet` hermetic (R1); the relation is **coverage**, not equality (B-2/B-3)*。**僅註解**，不改任何判斷式／常數。
  - 驗證：`make verify-architecture` 仍綠；`git diff` 該檔**僅註解行**變動。
  - 不做：不改 `droppedBuildEnv`／`childEnv` 的任何值或邏輯（round-042 frozen 行為）。

- [X] T004 subagent review (phase quality gate)
  - Read: `Makefile`（新 block）、`tools/arch/arch_test.go`（註解）、`research.md` -> D1–D12、`docs/decisions/0012-hermetic-make-go-env.md`
  - 檢驗：block 落在 `$(shell …)` 之上且**不 touch `PATH`**；neutralise set 含 `GOENV=off`（load-bearing）且 preserve set 未被觸及；`CGO_ENABLED` 未被全域 pin；**無新 target**、`verify` 名單不變；recipe 行內 per-target 賦值未被覆蓋；無產品碼／truth／既有 gate 語意變更。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 hostile-env 由 **red → green**、clean env 行為**不變**、兩個 positive controls 成立、既有 gates 未受影響。

**Test Scope**: `make` 呼叫（hostile／clean env）；`make verify`；`go test ./...`；Gherkin/DSL topology audit。

- [X] T005 [REGRESSION] 可偽性見證：4 個 red→green hostile-env（`GOENV`／`GO111MODULE`／`GOWORK`／`GOTOOLCHAIN`）+ neutralisation assertions；還原 block ⇒ 再 RED
  - Read: `research.md` -> D4, D7；`docs/decisions/0010-test-deadline-decoupling.md`；`spec.md` -> SC-001
  - 做：
    - (a) 對 T001 的 **4 案** 逐一在**已落 block** 的樹上重跑 → 全部 **exit 0**（`GOENV=<file -mod=vendor>` 由 *inconsistent vendoring* → green；`GOTOOLCHAIN` 由 *toolchain not available* → green）；另跑 `GOENV=/tmp/goenv043 make verify-architecture` 與 `GOENV=/tmp/goenv043 make verify`（aggregate，全員綠）→ green；並以 `make`-level probe 斷言 neutralise 集合的名字**不在 recipe env**（neutralisation assertion，非見證）。
    - (b) **還原**（暫時移除 block）→ `GOENV` 案**再度 FAIL**；還原回 block → green（證明非空轉、且見證綁定於 block 而非環境偶然）。
    - (c) **coverage 檢查**（非 equality drift witness，B-3）：以 `grep` 斷言 `Makefile` block 的 neutralise 集合（`GOENV`/`GOWORK`/`GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOTOOLCHAIN`/`GOFIPS140`/`GODEBUG`/`GOOS`/`GOARCH`/micro-arch family）**覆蓋** `tools/arch` `droppedBuildEnv` 的集合，**除去** `CGO_ENABLED`（機制差異，已記錄）——記錄為人工／scripted 檢查，**不** commit 見證檔。
  - 不做：不放寬任何斷言以「讓它過」；見證後必須還原到 HEAD（僅保留 T002 block）。

- [X] T006 [REGRESSION] 正向控制 + 全量回歸 + 範圍檢查
  - Read: `research.md` -> D7, D9；`spec.md` -> SC-002–SC-005；`specs/truth/techstack.md` -> 兩 row
  - 做：
    - **正向控制 1（cross-compile）**：clean env 下 `make verify-cross-compile` → **4/4**（行內 `CGO_ENABLED=0 GOOS=… GOARCH=…` 未被 block 覆蓋）。
    - **正向控制 2（mutating targets）**：clean env 下 `make tidy`（→ `go.mod`／`go.sum` 不變）、`make fmt`（→ 無 diff）。
    - `make verify` **OK**（既有 gates 全部；`verify` aggregate **名單不變**）。
    - `go test -count=1 ./...` 全綠（unit + godog E2E）。
    - Gherkin/DSL topology audit 重跑 → **unchanged**（44 features · 6 modules · 16 root + 327 module rows · 1674 steps）。
    - `gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 確認僅動 `Makefile`（+ 選配的 `tools/arch` 註解 + plan/truth 文件）；`go.mod`／`go.sum` 不變；`internal/**`／`cmd/**`／`tests/**` 產品碼未動。
  - 不做：不為了綠而改產品碼或放寬斷言。

- [X] T007 subagent review (round quality gate)
  - Read:
    - `Makefile`（新 block）、`tools/arch/arch_test.go`（註解，若有）、`specs/truth/techstack.md`（Hermetic toolchain invocation + Task runner rows）、`docs/decisions/0012-hermetic-make-go-env.md`
    - `specs/plans/043-hermetic-make-go-env/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：4 案 red→green 且還原後再 red；clean env 全綠；`verify-cross-compile` 4/4；`tidy`/`fmt` no-diff；`verify` 名單不變且無新 target；`GOENV=off` 存在且 preserve set 未被觸及；`CGO_ENABLED` 未全域 pin；drift witness 通過；與 `techstack.md`／`truth-delta.md`／ADR 0012 一致；**無產品碼變更**；`go.mod`／`go.sum` 不變。

---

## Execution outcome (T001–T007) — `/axb-implement` complete

**Delivered** (`Makefile` CHANGED + `tools/arch/arch_test.go` comment-only; **zero product code**; no new target; `go.mod`/`go.sum` unchanged):

- **`Makefile`** — a labelled `export`/`unexport` block at the top (after `VERSION ?= dev`, **above** the `$(shell command -v …)` probes): `export GOENV := off`, `export GOWORK := off`, `unexport GOFLAGS GO111MODULE GOEXPERIMENT GOOS GOARCH GOARM` + a comment block stating the neutralise set, the preserve set, the `GOENV=off` rationale, the host-default `CGO_ENABLED` (recipe-inline wins), and the `tools/arch` `childEnv` cross-reference (ADR 0012).
- **`tools/arch/arch_test.go`** — a **comment-only** cross-reference (`droppedBuildEnv`): the `Makefile` block is the primary owner; this filter covers the gate's documented **direct** invocation; the two sets must not drift silently. **No logic/value change** (round-042 frozen behaviour preserved).
- **No new target; the `verify` aggregate list is unchanged** (`plan.md` Structure Decision).

### Evidence

**T001 — RED baseline (unfixed tree, `make vet`):**

| # | Hostile env | exit (before) | note |
| --- | --- | --- | --- |
| 1 | `GOENV=<file: GOFLAGS=-mod=vendor>` | **2** | *inconsistent vendoring* — the #96 headline |
| 2 | `GOFLAGS=-trimpath` | 0 | harmless on this tree (recorded; still neutralised) |
| 3 | `GO111MODULE=off` | **2** | GOPATH-mode resolution failure |
| 4 | `GOWORK=/tmp/gowork043` | **2** | *directory prefix . does not contain modules listed in go.work* |
| 1b | `GOENV=<file>` + `make verify-architecture` | **2** | matches #96's reproduction verbatim |

**T005(a) — GREEN after the block:** cases 1, 2, 3, 4 and 1b all **exit 0** (case 1b: `0 issues` + `ok tools/arch` + the gate's success line).

**T005(b) — revert witness (non-vacuity):** with the block temporarily removed (`git checkout -- Makefile`), case 1 returned **exit 2**; with the block restored, **exit 0**. The witness is bound to the block, not to environment accident.

**T005(c) — drift witness:** the `Makefile` neutralise set = `{GOENV, GOWORK, GOFLAGS, GO111MODULE, GOEXPERIMENT, GOOS, GOARCH, GOARM}`; `tools/arch` `droppedBuildEnv` = the same set **+ `CGO_ENABLED`**. The only difference is `CGO_ENABLED` — the **documented host-default exception** (ADR 0012 D4; the gate pins it per target because it must cross-evaluate). **No silent drift.**

**T006 — positive controls + regression:**

- `make verify-cross-compile` → **exit 0** (linux/amd64 · linux/arm64 · darwin/amd64 · darwin/arm64 built + vetted — the recipe's inline `CGO_ENABLED=0 GOOS=… GOARCH=…` still wins).
- `make tidy` → exit 0, **no diff**; `make fmt` → exit 0, **no diff** (`git status` shows only the two intended files).
- `make verify` → **OK** (all existing gates + `verify-architecture`; golangci-lint 0 issues; govulncheck 0 reachable).
- `go test -count=1 ./...` → **green** (22 packages `ok`, **0** `FAIL`; incl. godog E2E).
- Gherkin/DSL topology audit → **PASSED & unchanged** (44 features · 6 modules · 16 root + 327 module rows · **1674** steps).
- `gofmt -l .` clean; `go.mod`/`go.sum` unchanged; the working-tree change set is exactly `Makefile` + `tools/arch/arch_test.go` (comment-only); no `internal/**`/`cmd/**`/`tests/**` product code touched.
- Tool probes intact: `STATICCHECK`/`GOLANGCI`/`GOVULNCHECK` still resolve from `PATH` (the block is above the `$(shell …)` probes and does not touch `PATH`).

### Reviews (T004 / T007)

- **Execution**: the session (inline self-review — **disclosed deviation**: no parallel-subagent substrate; round-029/041/042 precedent).
- **T004 (phase gate)**: **PASS** — the block sits above the `$(shell …)` probes; the neutralise set includes `GOENV=off` (load-bearing) and does not touch the preserve set; `CGO_ENABLED` is not globally pinned; **no new target** and the `verify` list is unchanged; recipe-inline per-target assignments are not clobbered; no product code / truth / existing-gate semantics changed.
- **T007 (round gate)**: **PASS** — 4 hostile envs red→green with a revert witness; clean env unchanged; `verify-cross-compile` 4/4; `tidy`/`fmt` no-diff; drift witness passes with the documented `CGO_ENABLED` exception; `tasks.md`/`techstack.md`/ADR 0012 consistent.
- **Folds raised during execution**: **none** (no plan/truth change was required).

### Deviations (recorded)

- **No parallel-subagent substrate in this session** — T004/T007 ran **inline self-review** (round-029/041/042 precedent).
- **No committed witness artifact** — per `plan.md`'s Structure Decision (no new target/artifact), the four hostile-env witnesses are **reproduced-then-reverted** (recorded here), the round-042 witness style. The temp env files (`/tmp/goenv043`, `/tmp/gowork043`) are scratch, not committed.
- **T003 kept** (not folded) — the one-line cross-reference comment is landed so both ownership sites are greppable. *(Post-fold: the comment now records the **coverage-not-equality** invariant and the ADR 0012 **R1** scope — see the fold note below.)*
- **`make test` under the hermetic env** — **exit 0**, **22 packages `ok`, 0 `FAIL`** (the E2E harness builds the binary as a child of `make` → a **nested `go build`**, the case A1 was chosen to cover; this is the strongest evidence that the block does not disturb the suite).

### Neutralisation assertion (ADR 0012 D7) — `make`-level probe, non-vacuous

Green-before inputs (`GOFLAGS=-trimpath`, the long-tail names) are **not** witnesses (B-1); the property they stand for is asserted directly against the **recipe environment**:

```
$ GOFLAGS=-trimpath GO111MODULE=off GOOS=plan9 GOTOOLCHAIN=go1.99.9 \
    make --eval='.PHONY: __probe' \
         --eval='__probe: ; @env | grep -E "^(GOFLAGS|GO111MODULE|GOEXPERIMENT|GOTOOLCHAIN|GOFIPS140|GODEBUG|GOOS|GOARCH|GOARM|GOARM64|GOAMD64|GO386|GOMIPS|GOMIPS64|GOPPC64|GORISCV64|GOWASM)=" || echo "NEUTRALISED: none present"' __probe
NEUTRALISED: none present        # WITH the block
```

**Non-vacuity** — the same probe against a **block-stripped** `Makefile` (`make -f /tmp/Makefile.stripped …`) reports the hostile names **present** (`GOOS=plan9`, `GOTOOLCHAIN=go1.99.9`, `GOFLAGS=-trimpath`, `GO111MODULE=off`). The two neutral **exports** are present in the recipe env: `GOENV=off GOWORK=off`.

### Coverage check (ADR 0012 D6) — coverage, not set-equality

The `Makefile` neutralise set **covers** `tools/arch`'s `droppedBuildEnv` set, **except `CGO_ENABLED`** — the documented mechanism difference (the block *preserves from the caller*; the gate's `childEnv` *re-sets* it per target). Verified by scripted `grep`; recorded here, not committed as a witness file.

### Widened-set validation (fold `d3c437f`, read-only)

| Ambient env (hostile) | `make vet` |
| --- | --- |
| `GOTOOLCHAIN=go1.99.9` | exit 0 |
| `GOAMD64=v4` | exit 0 |
| `GODEBUG=inittrace=1` | exit 0 |
| `GOFIPS140=latest` | exit 0 |
| `GOENV=<file: GOFLAGS=-mod=vendor>` | exit 0 |
| `GOFLAGS=-trimpath` · `GO111MODULE=off` · `GOWORK=<stray>` | exit 0 |

**Escape hatch holds:** an explicit per-invocation assignment still wins — `make GOENV=/tmp/goenv043 vet` ⇒ **exit 2** (*inconsistent vendoring*).



The operator's review of PR [#97](https://github.com/gosharplite/tellme/pull/97) (**B-1…B-4, TD-1, R-1…R-3**, plan branch `d014da0`) replaced the **incident-list** set with a **criterion-derived** one (ADR 0012 **D2**): *neutralise the ambient **build context** (what / which toolchain builds), preserve the **plumbing***. Consequences for this task file:

1. **The neutralise set is wider than T002's original** — it now also unsets `GOTOOLCHAIN`, `GOFIPS140`, `GODEBUG`, and the **micro-architecture family** (`GOARM64`, `GOAMD64`, `GO386`, `GOMIPS`, `GOMIPS64`, `GOPPC64`, `GORISCV64`, `GOWASM`) alongside `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOOS`/`GOARCH`/`GOARM`. `CGO_ENABLED` remains **preserved-from-the-caller** (not globally pinned — V1 unchanged).
2. **T005(c)'s drift witness is superseded** by the **coverage invariant** (ADR 0012 **D6**): the two sites neutralise by *different mechanisms* (the `Makefile` block disables the env file + unsets; `tools/arch`'s `childEnv` re-sets explicit values), so the relation is **coverage** — every name the block neutralises is either re-set by `childEnv` or **recorded as a non-covered class (R4; 14 exact names)** — **not** set-equality.
3. **T001's table stands**; re-validated against the folded block (see below).

**Re-validation against the folded block (read-only; ambient vars ⇒ `make vet`):**

| Ambient env | `make vet` |
| --- | --- |
| `GOTOOLCHAIN=go1.99.9` | exit 0 |
| `GOAMD64=v4` | exit 0 |
| `GODEBUG=inittrace=1` | exit 0 |
| `GOFIPS140=latest` | exit 0 |
| `GOENV=<file: GOFLAGS=-mod=vendor>` | exit 0 |
| `GOFLAGS=-trimpath` · `GO111MODULE=off` · `GOWORK=<stray>` | exit 0 |

**Escape hatch (ADR 0012 scope note) holds:** an explicit per-invocation assignment still wins — `make GOENV=/tmp/goenv043 vet` ⇒ **exit 2** (*inconsistent vendoring*), i.e. the boundary is deliberately overridable by the caller.

### Superseded: the original "considered and deliberately out of scope" note (kept for the record)

> The note below was written at `/axb-implement` time, **before** the PR #97 review. Its `GOTOOLCHAIN` / micro-arch line is **superseded** by the criterion-derived set above (the review's R-1(1)/B-4); the `CGO_ENABLED` line still stands.

- **`CGO_ENABLED`** — deliberately **not** neutralised (Q3→V1; ADR 0012 D4): a global pin would be a behaviour change vs future cgo-tagged code. The `tools/arch` gate pins it **per target** because it cross-evaluates; that is the only divergence in the drift witness.
- **`GOTOOLCHAIN`** — a *toolchain-selection* input, not a build-context one. Neutralising it as `local` could break a module that requires a newer toolchain (it would fail instead of fetching); leaving it ambient keeps the property incomplete (recorded). Out of #96's stated scope (`go env -w` / `GOENV` / `GOFLAGS` / `GO111MODULE` / `GOWORK`); a candidate for a future hardening round if the operator wants it.
- **Arch micro-tuning vars** (`GOAMD64`, `GOARM64`, `GOPPC64`, …) — same family as `GOOS`/`GOARCH`/`GOARM`, which **are** neutralised; the set is kept to the round-042 `childEnv` set + `GOENV` (plus `GOEXPERIMENT`) so the two owners stay mutually checkable (the drift witness). Adding more vars is a one-line edit if a reviewer wants it.

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Build & Tooling（**Hermetic toolchain invocation** row） | T002（依 neutralise/preserve set 實作）、T004、T006、T007 | PASS |
| `specs/truth/techstack.md` -> Build & Tooling（**Task runner** row：hermetic note；aggregate 名單不變） | T002、T004（名單不變）、T006（`make verify`） | PASS |
| `docs/decisions/0012-hermetic-make-go-env.md`（D1–D9） | T002（D1–D5）、T003（D6）、T005（D7–D8）、T007 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md` ×2 rows） | 已於 plan half 交付；T002／T006 驗證一致 | PASS |
| `truth-delta.md` -> Governance ADD（ADR 0012 + index） | 已於 plan half 交付；T002／T003／T007 依 ADR 實作/驗證 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T006 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`specs/truth/features/cli/**`） | 豁免（NOOP 不建任務；T006 驗證 topology audit 不變） | PASS |
| `research.md` -> D1（一個 top-of-`Makefile` block；非 per-recipe） | T002、T004 | PASS |
| `research.md` -> D2（neutralise set + preserve set + host-default `CGO_ENABLED`） | T002、T004 | PASS |
| `research.md` -> D3（unconditional、全 targets） | T002、T006 | PASS |
| `research.md` -> D4（`GOENV=off` load-bearing） | T001（RED 案 1）、T002、T004 | PASS |
| `research.md` -> D5（ownership：Makefile primary；childEnv defence-in-depth） | T002、T003、T005(c) | PASS |
| `research.md` -> D6（governance ADR 0012 + truth） | 已於 plan half 交付；T007 驗證一致 | PASS |
| `research.md` -> D7（4 見證 + 2 positive controls） | T001、T005、T006 | PASS |
| `research.md` -> D8（residuals R1/R3 記錄） | 已於 plan half 寫入 ADR 0012/truth；T007 檢核 | PASS |
| `research.md` -> D9（BDD techstack 不變；spec-by-example/api/data/dsl NOOP） | T006、T007 | PASS |
| `research.md` -> D10（no new dep；POSIX-only） | T002、T006 | PASS |
| `research.md` -> D11（ordering：043 now；R2 → 044） | 已於 plan half 寫入 STATUS；T006 範圍檢查 | PASS |
| `research.md` -> D12（BDD techstack unchanged） | T006、T007 | PASS |
| `spec.md` -> US1（FR-001–FR-005, NFR-001–NFR-002） | T001、T002、T005 | PASS |
| `spec.md` -> US2（FR-007, FR-011） | 已於 plan half 交付（ADR/truth）；T007 | PASS |
| `spec.md` -> US3（FR-006, FR-012） | T006（positive controls）、T007 | PASS |
| `spec.md` -> 全域（FR-008–FR-010, FR-013, NFR-003–NFR-004） | T002（FR-008）、T005(c)（FR-008 drift）、T006（FR-010/NFR-004）、T007 | PASS |
| `spec.md` -> SC-001（4 hostile env ⇒ green，red→green witness） | T001、T005(a)(b) | PASS |
| `spec.md` -> SC-002（clean env 行為不變；no new dep） | T006 | PASS |
| `spec.md` -> SC-003（positive control：cross-compile 4/4） | T006 | PASS |
| `spec.md` -> SC-004（positive control：`tidy`/`fmt` no-diff） | T006 | PASS |
| `spec.md` -> SC-005（truth + ADR + no product change + audit） | 已於 plan half 交付；T006、T007 | PASS |
| `plan.md` -> Structure（`Makefile` CHANGED；`techstack.md`/ADR 已 MODIFY+ADD；no new artifact；no product code） | T002、T003、T006 | PASS |
| `plan.md` -> Scope notes（api/data/dsl NOOP；ui/spec-by-example skipped） | T006、T007 | PASS |
| operator 拍板 Q1–Q8（A / A1 / V1 / S1 / G1 / D1 / R1+R2+R3 / O1） | T002（Q1–Q4）、T003（Q6）、T005（Q7）、T006（Q7-R2）、T007 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。

---

## Fold — PR [#99](https://github.com/gosharplite/tellme/pull/99) review (head `5fc8794`)

| ID | Class | Fold applied |
| --- | --- | --- |
| **N-1** | false **measured** claim in ADR 0012 **D1** (+ `research.md` D1) | Re-measured myself (`$(shell …)` child vs recipe, GNU Make 4.3): a `$(shell …)`/`$(eval …)` child receives make's **original environment verbatim** — a makefile `export` does *not* reach it and `unexport` does *not* strip it, so parse-time expansion is outside the boundary **for every name**. ADR D1's scope clause now states exactly that (with the three-line measurement), the "neutralised on both paths" claim is dropped, and the `Makefile`/`research.md` D1 scope clauses are mirrored; the residual **R5** and the operational rule (*keep `go` out of `$(shell …)`*) are unchanged. The edge-case clause in `spec.md` (which implied only unexported names leak) is corrected the same way. |
| **N-2** | stale round-020 TD1 framing (4 sites) | Reworded to *the `CGO_ENABLED=0` pin is the **precedent** generalised as an **invocation** rule; the pin itself stays **recipe-local** (D4)* in: ADR 0012 **Related** line · `docs/decisions/README.md` index row · `spec.md` §Behaviour intent · `research.md` §Lineage. (D4/FR-004 — `CGO_ENABLED` preserved from the caller — is unchanged.) |
| **N-3** | coverage-sentence precision | The `Makefile` comment and the `techstack.md` row now say *"…or **recorded as a non-covered class** (ADR 0012 R4)"* rather than implying every name is individually re-set/recorded. |

**Re-verification at the fold head** (the two probes the review named):

```
recipe env (14 hostile names set):  GOENV=off GOWORK=off  ; CGO_ENABLED=1 preserved ; no other GO* name
parse-time  (same ambient set):     GOENV=<file> GOFLAGS=-mod=vendor GOTOOLCHAIN=go1.99.9  ← ambient (ADR D1 corrected)
four red→green witnesses:           GOENV=<file> 0 · GO111MODULE=off 0 · GOWORK=<stray> 0 · GOTOOLCHAIN=go1.99.9 0   (each was 2)
escape hatch:                       make GOENV=<file> vet 2  ·  make GOFLAGS=-mod=vendor vet 0  ·  make GOTOOLCHAIN=go1.99.9 vet 0
```

## Fold — PR [#99](https://github.com/gosharplite/tellme/pull/99) re-review (`/pullrequestreview-5242647672`, FOLD-ACCEPTED at `8e144de`)

Three **non-blocking nits**, folded:

- **a (name vs class, 5 sites)** — ADR 0012 **D6**, `spec.md` **FR-008**, `research.md` **D5**, `plan.md` and the `tasks.md` T005-fold note now all read *"...re-set explicitly by `childEnv`, or recorded as a **non-covered class (R4)**"*; and **R4 enumerates the exact, closed list** — the **14** D2 names `childEnv` does not re-set: `GOENV`, `GOEXPERIMENT`, `GOTOOLCHAIN`, `GOFIPS140`, `GODEBUG` + the micro-arch family `GOARM`/`GOARM64`/`GOAMD64`/`GO386`/`GOMIPS`/`GOMIPS64`/`GOPPC64`/`GORISCV64`/`GOWASM` (every D2 name except `GOFLAGS`/`GO111MODULE`/`GOWORK`/`GOOS`/`GOARCH`, which `childEnv` re-sets).
- **b (ADR Related citation)** — the ADR **Related** line now also cites **PR #99** review `5242608457` (N-1...N-3) and the re-review `5242647672` (FOLD-ACCEPTED), so the immutable ADR carries its full fold provenance.
- **c (STATUS 043 row)** — the roadmap summary now states the folded set (criterion-derived; `GOENV`/`GOWORK` + `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOTOOLCHAIN`/`GOFIPS140`/`GODEBUG` + the ambient target triple; `CGO_ENABLED` preserved from the caller).

**Scope:** docs only (no `Makefile`/logic change); `go.mod`/`go.sum` untouched.

## Fold — PR [#99](https://github.com/gosharplite/tellme/pull/99) re-review #2 (`/pullrequestreview-5242675172` at `ad2ee7d`)

- **N-4 (new REFACTOR)** — `research.md` **D5**'s stale symmetric-difference sentence (*"`{GOENV}` on one side and `{CGO_ENABLED}` on the other"* — true of the **8-name** set, false after B-4 widened it to **19**) is replaced with the **derivation**: the *literal* symmetric difference is **12 Makefile-only names** + `{CGO_ENABLED}` on the `childEnv` side; *semantically* the **non-covered set is 14** — `GOEXPERIMENT` and `GOARM` sit in **both** literal sets (the filter *drops* them) but are never *re-set*, and **drop ≠ neutralise** (F-2). Verified mechanically: `Makefile(19) − childEnv-literal(8)` → 12; `Makefile(19) − childEnv-re-set(5)` → **14**. D5 is now consistent with **R4**'s exact list.
- **b2 (minor)** — the **two-example** form of R4 (*"drops but does not re-set `GOARM`/`GOEXPERIMENT` (and the names added this round)"*) is aligned to *"the **14** non-covered names (ADR 0012 R4)"* at the three remaining sites: `research.md` **D8**, `spec.md` **FR-009(c)**, `specs/truth/techstack.md` (the truth row). `grep -rn 'drops but does not re-set'` now shows only the ADR **R4** (the enumerated home) and these aligned pointers.

**Scope:** docs only (no `Makefile`/logic change); `go.mod`/`go.sum` untouched.

## Fold — PR [#99](https://github.com/gosharplite/tellme/pull/99) re-review #3 (`/pullrequestreview-5242697414` at `694ba91`)

- **N-5 (new; recorded as R6, option (i))** — make's **`-e` mode** (`make -e`, or an ambient `MAKEFLAGS=-e` / `GNUMAKEFLAGS=-e`) lets the ambient environment **override** the makefile, so it **resurrects the two `export`ed names** (`GOENV`, `GOWORK`) — reproduced here: `GOENV=<file> make -e vet` ⇒ **exit 2**; `MAKEFLAGS=-e …` ⇒ **exit 2**; `GNUMAKEFLAGS=-e …` ⇒ **exit 2**; the recipe env under `-e` shows `GOENV=[<file>] GOWORK=[<file>]` while `GOFLAGS`/`GO111MODULE`/`GOOS`/`GOTOOLCHAIN` stay neutralised (contrast: normal run `GOENV=[off] GOWORK=[off]`). This narrows D5's "unconditional" to **ambient *Go-env* input** — `-e` is a **make-mode** input. **Recorded, not closed**: option (ii) (`override export`) closes it but destroys the **R3** hatch (`make GOENV=<file> …`), and option (iii) (fail-loud `$(error …)`) adds a new failure mode — both a future decision, not this round's. Folded: **ADR 0012 D5** (scope qualifier) + **D8 (R6)** · `spec.md` **FR-009(e)** + **Q4** · `research.md` **D8** · the `techstack.md` truth row.

**Scope:** docs only (no `Makefile`/logic change); `go.mod`/`go.sum` untouched.

## Fold — PR [#99](https://github.com/gosharplite/tellme/pull/99) re-review #4 (`/pullrequestreview-5242720941` at `cc93051`)

- **N-6 (optional; folded via option (a))** — D5's *"(one exception — `-e`)"* undercounted: **`MAKEFILES`** is a second, **contrived** ambient make-level channel (an injected `override export GOENV=<file>` re-enables the load-bearing file). Reproduced here: `MAKEFILES=evil2.mk` (`override export GOENV=/tmp/goenv043`) ⇒ `make vet` **exit 2**; `MAKEFILES=evil.mk` (`override export GOFLAGS=-mod=vendor`) ⇒ **exit 0** (the `unexport` still defeats it); `MAKEOVERRIDES='GOENV=…'` ⇒ **exit 0** (no channel — no `$(MAKE)` recursion); `MAKEFLAGS='-- GOENV=…'` ⇒ **exit 2** (the already-documented R3 command-line-var carrier). Folded: **ADR D5** now says *"the exceptions are **make-level** inputs — `-e` (R6, reproduced) and a `MAKEFILES`-injected `override export`, same class but contrived"*; **ADR R6** gains the breadth note (MAKEFILES channel + the two positives); `spec.md` **FR-009(e)**, `research.md` **D8**, the `techstack.md` truth row aligned.

**Scope:** docs only (no `Makefile`/logic change); `go.mod`/`go.sum` untouched.
