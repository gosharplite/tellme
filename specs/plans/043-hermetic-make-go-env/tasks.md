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
- **[NEUTRALISE / Q3 → V1, D2]** **drop/neutralise**：`GOENV` → **`off`**（**load-bearing** —— Go 對「unset **或** empty」都會 fallback 到 env **file**，故單純 unset `GOFLAGS` **無法**中和 persisted `go env -w GOFLAGS=-mod=vendor`；round-042 F-2 / ADR 0011 D5）；`GOWORK` → `off`；`GOFLAGS`／`GO111MODULE`／`GOEXPERIMENT` → **unset**；ambient `GOOS`／`GOARCH`／`GOARM` → **unset**（host-native）。
- **[PRESERVE / Q3 → V1, D3]** 原樣傳遞：`PATH`／`HOME`／`GOPATH`／`GOMODCACHE`／`GOCACHE`（warm-cache／runtime set）**與** `GOPROXY`／`GOSUMDB`／`GOPRIVATE`／`GONOSUMDB`／`GOINSECURE`（network／checksum set）—— cold-cache 或 proxy 環境仍可解析（通用化 `verify-cross-compile` 的 PR #46 TD1 教訓）。
- **[CGO / Q3 → V1, D4]** `CGO_ENABLED` **維持 host default**；**唯一**的 cgo pin 仍是 `verify-cross-compile` recipe 內的**行內** `CGO_ENABLED=0`（**僅作用於該 recipe**）。block **不得**覆蓋 recipe 的行內 per-target 賦值（recipe 的行內賦值對該 recipe 勝出，FR-012）。
- **[SCOPE / Q4 → S1, D5]** **無條件、全 targets**（含 mutating 的 `fmt`／`tidy`）；無 per-target opt-out。
- **[OWNERSHIP / Q6 → D1, D6]** `Makefile` block 為**唯一主要 owner**；round-042 的 `tools/arch/arch_test.go` `childEnv`（ADR 0011 D5）**保留**為 **defence-in-depth**，涵蓋 gate 的**直接**呼叫路徑（繞過 `make`）。兩處**互相 cross-reference 註解**；兩份字面集合**不得 silent drift**（見 T004 的 drift witness 檢查）。
- **[VERIFY / D7, Q7 → R2]** 驗收證據 = (a) **四個 hostile-env 見證**（`GOENV=<file: GOFLAGS=-mod=vendor>`、`GOFLAGS=-trimpath`、`GO111MODULE=off`、stray `GOWORK=<path>`；**red before / green after**，重現後還原，ADR 0010 doctrine）+ (b) **兩個 positive control**：`make verify-cross-compile` 在 hermetic env 下仍 **4/4**（證明只移除 ambient 輸入、未覆蓋行內 per-target 覆寫）；clean env 下 `make tidy`／`make fmt` **no-diff**、`go.mod`／`go.sum` 不變（證明未破壞 mutating targets）。
- **[RESIDUALS / Q7 → R1+R3, D8]** 需記錄於 ADR 0012／truth row（已於 plan half 寫入）：(R1) **繞過 `make` 的裸 `go`** 仍非 hermetic（hermeticity 是 `make`-boundary 性質）；(R3) `GOENV=off` 亦忽略**正當**的 operator `go env -w`，escape hatch = 每次呼叫顯式 export（`GOPROXY=… make verify`）。
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
    - 以暫存檔建立 4 個 hostile ambient 環境，各對一個**便宜**的 toolchain target 取值（`vet`；`GOENV` 案另加 `verify-architecture` 對照）：
      1. `printf 'GOFLAGS=-mod=vendor\n' > /tmp/goenv043; GOENV=/tmp/goenv043 make vet`（及 `… make verify-architecture`）
      2. `GOFLAGS=-trimpath make vet`
      3. `GO111MODULE=off make vet`
      4. `printf 'go 1.26\n' > /tmp/gowork043; GOWORK=/tmp/gowork043 make vet`
    - **期望 RED**：至少 `GOENV=<file -mod=vendor>` 案必須非零失敗並出現 *inconsistent vendoring*（即 #96 的 headline 重現）；其餘記錄其實際結果（可能綠——如 `-trimpath` 在不影響 build graph 時）。
    - 把 4 案的原始輸出與 exit code 記入本任務（供 Phase 4 對照 `green after`）。
  - 不做：不為讓它綠而改樹；不建立任何 committed 見證檔（本輪**不新增 artifact** —— `plan.md`）。
  - **Red-first 期望**：`GOENV` 案 **FAIL**（exit ≠ 0）；此為 T002 的對照基線。

- [X] T002 [MECH] 落 `Makefile` 的 hermetic `export`/`unexport` block
  - Read:
    - `specs/plans/043-hermetic-make-go-env/research.md` -> D1, D2, D3, D4, D5, D10
    - `docs/decisions/0012-hermetic-make-go-env.md` -> D1, D2, D3, D4, D5
    - `specs/truth/techstack.md` -> Build & Tooling（Hermetic toolchain invocation row）
    - `Makefile`（`VERSION ?= dev` 之下、`STATICCHECK := $(shell …)` **之上**的位置）
  - 做（在 `Makefile` 頂部、`VERSION ?= dev` 之後、`$(shell command -v …)` probes 之前，加一段帶註解的 block）：
    - `export GOENV := off`（load-bearing；註解說明 env-file fallback 對 unset **與** empty 都生效 —— round-042 F-2）
    - `export GOWORK := off`
    - `unexport GOFLAGS GO111MODULE GOEXPERIMENT GOOS GOARCH GOARM`（僅 **ambient** build-context 被清除；host 維持 native）
    - 一段區塊註解：**目的**（#96；generalises ADR 0011 D5 + round-020 TD1）、**neutralise set**、**preserve set**（`PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` + `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE` —— **不**被 touch）、**`CGO_ENABLED` 維持 host default** 且 recipe 行內賦值仍勝出、**cross-reference** 指向 `tools/arch` 的 `childEnv`（defence-in-depth）與 **ADR 0012**。
    - 確認 `$(shell command -v …)` probes 仍能解析 `PATH`（block 在其上、且不 touch `PATH`）。
  - 驗證：`gofmt` n/a（Makefile）；clean env 下 `make help` 正常；T001 的 4 案（至少 `GOENV` 案）由 **FAIL → PASS**（Phase 4 逐一對照）。
  - 不做：不新增 target；不改 `verify` aggregate 名單；不 touch `CGO_ENABLED`（全域）；不覆蓋 `verify-cross-compile` 的行內 per-target 賦值；不改任何 recipe 指令本身。

- [X] T003 [DOC] `tools/arch/arch_test.go` 加一行 cross-reference 註解（選配；**零行為變更**）
  - Read:
    - `docs/decisions/0012-hermetic-make-go-env.md` -> D6（ownership: primary = Makefile block；childEnv = defence-in-depth）
    - `tools/arch/arch_test.go` -> `childEnv` / `droppedBuildEnv` 上方的註解塊
  - 做：在 `droppedBuildEnv`（或 `childEnv`）註解處補一句——*the `Makefile` hermetic block (ADR 0012) is the primary owner for `make`-launched invocations; this filter covers the gate's documented direct invocation*。**僅註解**，不改任何判斷式／常數。
  - 驗證：`make verify-architecture` 仍綠；`git diff` 該檔**僅註解行**變動。
  - 不做：不改 `droppedBuildEnv`／`childEnv` 的任何值或邏輯（round-042 frozen 行為）。
  - **可折疊**：若 review 認為此註解非必要，可整條移除而不影響本輪交付。

- [X] T004 subagent review (phase quality gate)
  - Read: `Makefile`（新 block）、`tools/arch/arch_test.go`（註解）、`research.md` -> D1–D12、`docs/decisions/0012-hermetic-make-go-env.md`
  - 檢驗：block 落在 `$(shell …)` 之上且**不 touch `PATH`**；neutralise set 含 `GOENV=off`（load-bearing）且 preserve set 未被觸及；`CGO_ENABLED` 未被全域 pin；**無新 target**、`verify` 名單不變；recipe 行內 per-target 賦值未被覆蓋；無產品碼／truth／既有 gate 語意變更。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 hostile-env 由 **red → green**、clean env 行為**不變**、兩個 positive controls 成立、既有 gates 未受影響。

**Test Scope**: `make` 呼叫（hostile／clean env）；`make verify`；`go test ./...`；Gherkin/DSL topology audit。

- [X] T005 [REGRESSION] 可偽性見證：4 個 hostile-env 由 RED → GREEN；還原 block ⇒ 再 RED
  - Read: `research.md` -> D4, D7；`docs/decisions/0010-test-deadline-decoupling.md`；`spec.md` -> SC-001
  - 做：
    - (a) 對 T001 的 4 案逐一在**已落 block** 的樹上重跑 → 全部 **exit 0**（`GOENV=<file -mod=vendor>` 由 *inconsistent vendoring* → green）；另跑 `GOENV=/tmp/goenv043 make verify-architecture` → green。
    - (b) **還原**（暫時移除 block）→ `GOENV` 案**再度 FAIL**；還原回 block → green（證明非空轉、且見證綁定於 block 而非環境偶然）。
    - (c) **drift witness**：以 `grep` 斷言 `Makefile` block 中和的變數集合（`GOENV`/`GOWORK`/`GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOOS`/`GOARCH`/`GOARM`）**涵蓋** `tools/arch` `droppedBuildEnv` 的變數集合（不 silent drift；Q6/D6）——記錄為人工／scripted 檢查，**不** commit 見證檔。
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
- **T003 kept** (not folded) — the one-line cross-reference comment is landed so both ownership sites are greppable.

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
