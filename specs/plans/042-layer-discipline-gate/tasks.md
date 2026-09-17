# Tasks: layer-discipline gate + violation baseline (round 042 — implementation half)

**Plan Package**: `specs/plans/042-layer-discipline-gate`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0011-layer-discipline-gate.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 build/quality-pipeline（tooling）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D2/D6；stdlib-only，`go.mod`／`go.sum` 不動）。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`；驗證由 **gate 自身的 exit code** 與可偽性見證承擔（非 Gherkin，`spec.md` A3 / `research.md` D12）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`tools/arch/` 的 layer-discipline gate + 其 committed baseline + `Makefile` 接線**；truth 變更（`specs/truth/techstack.md` gate row + Task runner row）與 **ADR 0011** 已由 `/axb-technical-research` 於 plan/truth half 交付（PR #94 已 merged into `dev`），並已記入 `truth-delta.md`。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 對應段落。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 gate（`verify-no-test-sleep`／`verify-no-network`／`vet`／`verify-cross-compile`／`verify-mcp-sdk-confinement`／`lint`／`vulncheck`）的語意、任何 `internal/**`／`cmd/**` 產品碼。

## Round-042 locked decisions (implementation constraints — MUST)

> 來自 `research.md` D1–D12、`spec.md`（FR-001–013 / SC-001–005）與 **ADR 0011**，外加 PR #94 最終認證的三個 implementation-time notes（**N-1…N-3**）。

- **[RULE / D1, ADR D2]** predicate 為**兩段式**：**(A)** 不得 import **更高** tier；**(B)** application tiers（`internal/app/**`、`internal/cli`）不得 import `internal/infrastructure/**`；**(C)** `internal/domain/**` 只可 import domain + stdlib；**(D)** 未分級的 `internal/**` package = violation（default-deny）。tier 表（0 domain · 1 config/home · 2 app · 3 infrastructure · 4 agent · 5 ui/tui · 6 cli）為**唯一權威機讀來源**（ADR D7）。
- **[MECH / D2/D4]** guard 為 **`//go:build arch` 的 Go test**，位於 `tools/arch/`；以 stdlib `go list` 讀模組自身 import graph，**`cmd.Dir = <module root>`**（Go test 的 CWD 是其 package 目錄，`./...` 必須錨定）；self-test 先斷言**enumerated graph 本身**（套件數 + 已知受治理套件）再做規則斷言；**不傳 `-e`**；child 錯誤或空 graph = **fail**（B-2 / FR-005）。
- **[UNION / D5, TD-1]** 對 `CROSS_TARGETS`（`linux/amd64 linux/arm64 darwin/amd64 darwin/arm64`）取 **union**，使 OS-gated illegal import 無所遁形。
- **[ENV / D5, F-1]** child `go list` 用**過濾後的繼承環境**：**drop/neutralise** `GOOS`/`GOARCH`/`GOARM`/`CGO_ENABLED`（改由 per-target 設定）＋ `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOWORK`；**preserve** `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE`（承載 warm module cache）。
- **[TAG / D6, TD-2]** 不傳 `-tags` 給 child；tag-gated（如 `//go:build arch`）檔案 **out of scope**（唯一 custom-tagged 檔即 guard 本身，在 exempt 的 `tools/**`）。
- **[BASELINE / D3, Q3]** committed、**sorted**（Go `sort.Strings`，byte-wise；ASCII ` -> `；module-relative package 路徑）baseline；**ratchet that only shrinks**：baseline 未列的新 violation → **fail**；baseline 列了但已不違規（stale）→ **fail**（"remove it from the baseline"）。
- **[DEFAULT-DENY / D8, TD-3]** 未分級的 `internal/**` package → violation。
- **[ACYCLIC / D11, TD-4]** 額外以 stdlib **SCC** 斷言 **0 cycles**（cycles **無** baseline，必為 0）。
- **[N-1]** baseline **generation affordance**：guard 提供 `-args -update-baseline`（測試旗標），並加 `Makefile` 目標 `verify-architecture-update`；於 `tools/arch/baseline.txt` 檔頭引用之，讓 R2–R4 不必各自發明。
- **[N-2]** **absent / unreadable / 不可解析** 的 baseline → **fail**（**never** 當「未設定 baseline」）；**empty** → 僅在**存在違規時** fail（ratchet 的終點 = 0 違規時，header-only baseline 必須綠，PR #95 review **F-1**）；self-test 斷言 committed baseline 的每行可解析為 ` -> `。
- **[N-3]** entry test（`TestVerifyRealArchitecture`）**必須一次涵蓋三性質**（enumeration · ranking/baseline diff · acyclicity），且**顯式斷言三者都跑到**（`-run TestVerifyRealArchitecture` 是契約的一部分；不得讓某性質落在別的 test 函式而被 `-run` 靜默略過）。
- **[TRUTH / D7]** truth（`specs/truth/techstack.md` 兩 row + ADR 0011）已於 plan half 交付；本輪**不**再改 truth，只交付 gate + baseline + 接線。
- **[NO-DEP]** stdlib-only；`go.mod`／`go.sum` 不動；POSIX-only。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D2/D6；無新相依，`go.mod`／`go.sum` 不動）。不得把 gate 骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點骨架；gate 落點即**新目錄** `tools/arch/`（由 T001 直接交付），`Makefile` 接線由 T003 交付。

## Phase 3: Implementation & Alignment (gate, RED-first)

**Goal**: 交付 `tools/arch/` 的 layer-discipline gate（兩段式 predicate + SCC + union + env 過濾 + baseline diff/ratchet）與其 committed baseline，並完成 `Makefile` 接線，使 `dev` 上綠燈、新違規紅燈、stale 條目紅燈。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗證語言為 Go test（gate 自身）。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Build & Tooling（**Layer-discipline gate** row + **Task runner** row）
- `docs/decisions/0011-layer-discipline-gate.md`（D1–D10；唯一權威機讀 ranking = guard 的 tier 表）
- `specs/plans/042-layer-discipline-gate/research.md` -> D1–D12
- `specs/plans/042-layer-discipline-gate/plan.md` -> Pinned layer ranking + 預期變更（`Makefile` CHANGED；`tools/arch/` NEW）

**Boundary**:
- 只新增 `tools/arch/**`（`doc.go`、`arch_test.go`、`baseline.txt`）與修改 `Makefile`；**不**碰任何產品碼、**不**改既有 gate 語意、**不**改 truth。
- guard 之 tier 表為唯一權威來源；`techstack.md` 以 predicate 指涉並 cite ADR 0011（RF-3）。
- review 啟動 subagent；通過前不解鎖 Phase 4。

- [X] T001 建立 `tools/arch/` guard（`//go:build arch` test + untagged `doc.go`）
  - Read:
    - `specs/plans/042-layer-discipline-gate/research.md` -> D1, D2, D4, D5, D6, D8, D9, D10, D11, D12
    - `docs/decisions/0011-layer-discipline-gate.md` -> D1, D2, D4, D5, D6, D7, D8, D9, D10
    - `specs/truth/techstack.md` -> Build & Tooling（Layer-discipline gate row）
    - `Makefile`（既有 gate 風格；`CROSS_TARGETS` 已在 `verify-cross-compile` 定義）
  - 做（**新增** `tools/arch/doc.go` = untagged package doc，使 `go build ./...`／`go vet ./...` 不因 build constraint 全排除而報錯；**新增** `tools/arch/arch_test.go`，`//go:build arch`，內含 guard 全部邏輯）：
    - **tier 表**（`internal/domain/**`=0 · `internal/config`/`internal/home`=1 · `internal/app/**`=2 · `internal/infrastructure/**`=3 · `internal/agent`=4 · `internal/ui`(+`ui/tui/**`)=5 · `internal/cli`=6；非 `internal/` 一律 **exempt**；`internal/**` 未分級 = **RULE-D**）。
    - **`evaluate(graph) []string`**：回傳 **sorted** canonical violation lines `<src> -> <dst>`（module-relative、ASCII ` -> `），涵蓋 RULE-A/B/C/D。
    - **enumeration**：`moduleRoot()` 以 `runtime.Caller` 向上找 `go.mod`；對每個 `CROSS_TARGETS` 以 `exec.Command("go","list","-f","{{.ImportPath}}|{{join .Imports \" \"}}","./...")`、`cmd.Dir=moduleRoot`、`cmd.Env=childEnv(goos,goarch)` 取值並 **union**；`childEnv` = 繼承 env **過濾**（drop `GOOS`/`GOARCH`/`GOARM`/`CGO_ENABLED`/`GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOWORK`，再由目標設 `GOOS`/`GOARCH`/`CGO_ENABLED=0`；其餘如 `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` **保留**）。
    - **acyclicity**：Tarjan SCC over governed ranked graph；斷言 **0** 個大小 >1 的 SCC（cycles 無 baseline）。
    - **baseline**：`readBaseline` 讀 `tools/arch/baseline.txt`（skip 空行與 `#` 開頭；每行須含 ` -> `；**absent/unreadable/empty/不可解析 → `t.Fatalf`**）；`writeBaseline`（`-args -update-baseline` 時）寫入檔頭註解（引用 `make verify-architecture-update`）＋ sorted violations。
    - **entry test `TestVerifyRealArchitecture`**：依序 (1) enumerate 並**斷言 graph 本身**（套件數 ≥ 20 + 存在 `internal/cli`/`internal/agent`/`internal/ui`/`internal/ui/tui/prompt`/`internal/domain/llm`）；(2) `selfTestPredicate`（synthetic graph 的 table-driven 斷言，覆蓋 RULE-A/B/C/D 與合法邊界）；(3) `assertNoUnrankedGoverned`（default-deny self-test）；(4) `evaluate`；(5) `cycles` 斷言 0；(6) **顯式斷言三性質皆執行**（N-3）；(7) `-update-baseline` 時寫 baseline 後 return；(8) 否則 `readBaseline` + **diff**：新 violation → `t.Errorf`（列出）；stale → `t.Errorf`（"remove it from the baseline"）。
    - `-update-baseline` 以 package-level `flag.Bool("update-baseline", …)` 定義，供測試旗標傳入。
  - RED-first 期望：**未**提交 baseline 前，`readBaseline` 以 absent/empty → **fail**（證明「gate 未自帶 baseline 即紅」，對應 A2 atomic-delivery / N-2）。
  - 不做：不改 `Makefile`（T003）；不提交 baseline（T002 產生）；不碰產品碼／truth／既有 gate；不引入相依。

- [X] T002 產生並提交 `tools/arch/baseline.txt`（generated，非手抄）
  - Read:
    - `specs/plans/042-layer-discipline-gate/research.md` -> D3, D10
    - `docs/decisions/0011-layer-discipline-gate.md` -> D3, D9（worked example）
    - `specs/plans/042-layer-discipline-gate/spec.md` -> SC-001（8 violations）
  - 做：以 `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline` **生成** `tools/arch/baseline.txt`（不得手抄）；確認內容為 sorted、ASCII ` -> `、module-relative 的 **8** 行（7 `internal/cli -> internal/infrastructure/*` + `internal/agent -> internal/ui`），且與 ADR 0011 的 worked example 逐行一致；檔頭含 `make verify-architecture-update` 生成法（N-1）。
  - 驗證：重跑 `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch` → **綠**（baseline 與 gate 輸出相符）。
  - 不做：不手改任何 baseline 行；不為了綠而移除違規。

- [X] T003 `Makefile`：新增 `verify-architecture` + `verify-architecture-update` 並接入 `verify`
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling（Layer-discipline gate row；Task runner row）
    - `docs/decisions/0011-layer-discipline-gate.md` -> Consequences（New）
    - `Makefile`（既有 gate 風格：`verify-cross-compile`、`verify` aggregate、`.PHONY`、`help`）
  - 做：
    - 新增 `verify-architecture` target：`@go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch`（+ echo 說明）。
    - 新增 `verify-architecture-update` target：`@go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline`（N-1 affordance）。
    - 將 **`verify-architecture`** 加入 `verify` aggregate（置於 `verify-cross-compile`/`verify-mcp-sdk-confinement` 之後，維持 warm cache）、`.PHONY` 與 `help` 文字；`verify-architecture-update` 只進 `.PHONY` + `help`，**不**進 `verify`。
  - 不做：不改既有 target 行為；不把 `-update-baseline` 放進 `verify`（避免 gate 自我改寫 baseline）；不加相依。

- [X] T004 subagent review (phase quality gate)
  - Read: `tools/arch/doc.go`、`tools/arch/arch_test.go`、`tools/arch/baseline.txt`、`Makefile`、`research.md` -> D1–D12、`docs/decisions/0011-layer-discipline-gate.md`
  - 檢驗：predicate 為兩段式且 **能導出** 8 行 baseline；enumeration 錨定 module root 且 self-test 斷言 graph 本身；union over `CROSS_TARGETS`；env 過濾含 drop **與** preserve 兩集；baseline 缺／空／不可解析 → fail；entry test 涵蓋三性質並顯式斷言；無 undefined 行為；未動產品碼／truth／既有 gate。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 gate 在 `dev` 上綠、對**新**違規與**stale** 條目皆紅（可偽性）、既有 gates／行為未受影響。

**Test Scope**: `tools/arch/**`（unit；`-tags=arch`）；whole-suite `go test ./...`；`make verify`；Gherkin/DSL topology audit。

- [X] T005 [REGRESSION] 可偽性見證 (a) 新違規 ⇒ 紅 / (b) stale 條目 ⇒ 紅
  - Read: `research.md` -> D3, D11；`docs/decisions/0010-test-deadline-decoupling.md`（見證 doctrine）；`tools/arch/arch_test.go`
  - 做：
    - **(a) 新違規**：暫時在一個受治理套件加入一個 illegal import（例如於 `internal/domain/llm` 暫時 import `internal/config`；或於 `internal/cli` 暫時 import 另一個 `internal/infrastructure/*`）→ `make verify-architecture` **非零失敗並列出** 該 `src -> dst`；**還原**後確認綠。
    - **(b) stale**：暫時自 `tools/arch/baseline.txt` 移除一行（其違規仍在）→ gate 回報**新 violation** 而紅；改成暫時修掉一個 baselined 違規（或自 baseline 保留一行卻讓該行不再違規）→ gate 回報 **stale** 而紅；兩種皆觀察到即**還原**，重跑確認綠。
  - 不做：不放寬任何 assertion 以「讓它過」；見證後必須還原到 HEAD。

- [X] T006 [REGRESSION] 全量回歸 + `make verify` + topology audit + 範圍檢查
  - Read: `research.md` -> D3, D7, D12；`specs/truth/techstack.md` -> 兩 row
  - 做：
    - `make verify-architecture` → 綠（8 baselined、0 new、0 stale、0 cycles）。
    - `go test -count=1 ./...` 全綠（unit + godog E2E；`tools/arch` 於預設 tags 下為 `[no test files]`）。
    - `make verify` **OK**（既有 gates + 新成員 `verify-architecture`）。
    - Gherkin/DSL topology audit 重跑 → **unchanged**（44 features · 16 root + 327 module rows · 1674 steps）。
    - `gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 確認僅動 `Makefile` + `tools/arch/**`（+ 文件）；`go.mod`／`go.sum` 不變；`internal/**`／`cmd/**` 產品碼未動。
  - 不做：不為了綠而放寬 assertion、改產品碼或 baseline 手抄。

- [X] T007 subagent review (round quality gate)
  - Read:
    - `tools/arch/doc.go`、`tools/arch/arch_test.go`、`tools/arch/baseline.txt`、`Makefile`
    - `specs/truth/techstack.md`（Layer-discipline gate row；Task runner row）、`docs/decisions/0011-layer-discipline-gate.md`
    - `specs/plans/042-layer-discipline-gate/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：`dev` 上綠；**新違規 ⇒ 紅**；**stale ⇒ 紅**；baseline absent/empty ⇒ fail；三性質皆有跑（N-3）；tier 表為唯一權威且與 `techstack.md` predicate 一致；env 過濾含 drop+preserve；union 四 target；無新相依；既有 gate 語意與 `stdout`/`stderr` 未變；**無產品碼變更**；`techstack.md` 與 `truth-delta.md` 一致。

---

## Execution outcome (T001–T007)

**Delivered** (`tools/arch/` NEW + `Makefile` CHANGED; **zero product code**; `go.mod`/`go.sum` unchanged):

- `tools/arch/doc.go` — untagged package doc (keeps the dir buildable under default tags).
- `tools/arch/arch_test.go` — `//go:build arch`; the gate: tier table (normative ranking), `evaluate` (two-part predicate A–D), `moduleRoot` (walk-up to `go.mod`), `childEnv` (drop/neutralise + preserve), `enumerate` (`go list` union over `CROSS_TARGETS`), `cycles` (Tarjan SCC), `readBaseline`/`writeBaseline`, `selfTestPredicate`, and the single entry test `TestVerifyRealArchitecture` (enumeration + predicate self-test + default-deny coverage + ranking diff + acyclicity + an explicit “all three properties ran” assertion, N-3).
- `tools/arch/baseline.txt` — **generated** from the gate (`-args -update-baseline`), **8** lines (7 `internal/cli -> internal/infrastructure/*` + `internal/agent -> internal/ui`), matching ADR 0011's worked example line-for-line.
- `Makefile` — `verify-architecture` (member of `verify`, after `verify-mcp-sdk-confinement`), `verify-architecture-update` (baseline regeneration; **not** in `verify`), `.PHONY` + `help`.

### Phase-3 review outcome (T004)

- **Reviewer**: the session (inline self-review — **disclosed deviation**, the round-029/8 and round-041 precedent: no parallel-subagent substrate in this session).
- **Result**: **PASS**. The predicate derives the 8-line baseline by recomputation (7 RULE-B + 1 RULE-A); the enumeration is anchored to the module root and the self-test asserts the graph itself (count + named governed packages) before ranking; the child env carries **both** the drop-set and the preserve-set; an absent/empty/unparsable baseline fails; the entry test covers all three properties.

### Round review outcome (T007)

- **Result**: **PASS** — `dev` green, new violation ⇒ red, stale ⇒ red, baseline absent ⇒ fail, 0 cycles, no new dependency, no product code changed.

### Fold review (PR #95 review #1, comment `5721461410` — APPROVE WITH REQUIRED FOLDS; all folded)

- **F-1 [TECHNICAL DEBT → folded] the ratchet's terminal state is unreachable.** `readBaseline` rejected *any* empty baseline, but `writeBaseline` writes header-only at 0 violations ⇒ the round's own endpoint (count → 0) had no acceptable state. **Fix:** `readBaseline` no longer fails on empty; the caller fails only when `len(violations) > 0 && len(baseline) == 0`. The N-2 anti-bypass intent is preserved exactly (an emptied baseline while 8 violations exist still fails — witnessed), and a genuine zero is green. (Mechanism refinement; **ADR 0011 D3 unchanged**.) **Witness:** header-only baseline with 8 violations ⇒ `lists no violations but 8 exist — an emptied baseline MUST fail`.
- **F-2 [TECHNICAL DEBT → folded] `drop` is not `neutralise`.** Deleting a key does not neutralise a persisted (`go env -w`) setting — Go falls back to the env file for unset **and** empty values. **Fix:** `childEnv` now sets explicit **non-empty** values — `GOFLAGS=-mod=readonly`, `GO111MODULE=on`, `GOWORK=off` — so the artifacts' word *neutralise* is true of the mechanism (chosen over `GOENV=off`, which would discard a persisted `GOMODCACHE`/`GOPATH` and undermine the preserve-set). **Witness:** with a persisted `GOENV` file holding `GOFLAGS=-mod=vendor` and the outer given an explicit `GOFLAGS=-mod=readonly`, the gate is green (the child's explicit value wins over the file). *(The **outer** `go test`/`go vet` inheriting the environment is a repo-wide property, true of every gate in this Makefile — out of scope.)*
- **F-3 [TECHNICAL DEBT → folded] `moduleRoot` broke under `-trimpath`.** `runtime.Caller(0)` returns a module-relative path under `-trimpath`, so the walk-up found no `go.mod`. **Fix:** resolve the root from the test process's **CWD** first (a Go test's CWD is its package dir, inside the module — trimpath-immune), keeping the caller path as a fallback. **Witness:** `GOFLAGS=-trimpath make verify-architecture` ⇒ green.
- **F-4 [REFACTOR → folded] `internal/agent` was exact-match while its siblings are prefix-match.** **Fix:** `tier()` now ranks the `internal/agent` **subtree** by prefix (`internal/agent` or `internal/agent/`), so a future `internal/agent/*` restructure (R3/R4) does not trip RULE-D by accident. (`internal/config`/`internal/home` stay exact-match — leaf packages; RULE-D flags any new subpackage actionably.) **Witness:** adding `internal/agent/zz_tmp_sub` leaves the gate green (tier 4). ADR 0011 D10 already carries the wider "RULE-D will tell you" record; the ADR is merged/immutable, so the mechanism refinement lives here.
- **N-a [CONSISTENCY → recorded]** the exemption is **symmetric and silent**: `evaluate` skips *any* internal → non-internal edge, so `internal/** → cmd|tests|tools` is never flagged (recorded so the choice is explicit, not accidental).
- **N-b [FALSIFIABILITY wording → recorded]** two cycle paths: a **same-target** cycle is caught first by the child `go list` error (`runGoList`'s `t.Fatalf`); only a **cross-target** cycle (a→b on linux, b→a on darwin) reaches the **SCC** pass. Both fail the gate.
- **N-c [HARDENING → folded]** the `verify-architecture` target now runs `go vet -tags=arch ./tools/arch` before the test, so a compile/vet error in the build-tagged guard surfaces locally (the guard file is compiled by no other gate).
- **N-e [ACCOUNT-KEEPING → folded into the PR body]** the verification header names the **code** SHA and notes the docs commit.

### Fold review #2 (PR #95 review #2, comment `5721537481` — APPROVE WITH REQUIRED FOLDS; all folded)

- **Fold 1 [TECHNICAL DEBT → folded, option (a) implement] test-only imports were not evaluated although the artifacts say they are.** `go list`'s `.Imports` **excludes** test imports, so `spec.md` Q2 / the `_test.go` edge case ("`internal/**` test files are **also governed**") asserted coverage the mechanism did not honour — and R2 reworks the `internal/cli` test seams, where a test-only upward import could slip through. **Fix (implement):** `listFormat` now carries `.Imports` **and** `.TestImports` **and** `.XTestImports`, merged into the graph (a test file's import is an edge from the package it lives in). Measured: **0** test-only upward/RULE-B/RULE-C edges exist today ⇒ **no baseline churn** (verified — `make verify-architecture-update` leaves `baseline.txt` byte-identical). **Witness:** a temp `internal/domain/llm/zz_probe_test.go` with `import _ "…/internal/infrastructure/mcp"` ⇒ **FAIL**, naming `internal/domain/llm -> internal/infrastructure/mcp` (previously green); removed ⇒ green. (The guard's own package stays exempt; `//go:build arch` files remain invisible per TD-2.)
- **Fold 2 [TECHNICAL DEBT → filed] the outer-env residual needs a durable home.** F-2 fixed only the **child**; the **parent** `go test`/`go vet` (and every other `Makefile` gate) still inherit a persisted (`go env -w`) / ambient Go env (`GOENV=<file: GOFLAGS=-mod=vendor> make verify-architecture` ⇒ exit 2). Scoping it out of R1 was correct, but it must not live only in this thread. **Filed as [#96](https://github.com/gosharplite/tellme/issues/96)** — *"Makefile gates are not hermetic against a persisted/ambient Go env (`go env -w` / `GOENV` / `GOFLAGS`)"* (same family as round-020's PR #46 **TD1** `CGO_ENABLED=0` pin and this round's F-2).
- **Fold 3 [NIT → recorded] N-d had no record.** The `properties != 3` check in the entry test is a **static counter**, not a proof the properties ran; it is harmless because all three are inlined in `TestVerifyRealArchitecture`. Recorded here so the record matches the claim.

#### Nits (review #2, recorded)

- **`-mod=readonly` has a second effect:** a working tree whose `go.mod`/`go.sum` need updates now fails the child's `go list` (surfaced by `runGoList`'s `t.Fatalf`); consistent with the aggregate (`vet`/`verify-cross-compile` fail first), but stated rather than implied.
- **`GO111MODULE=on` / `GOWORK=off` are deliberate:** the layer rule is **module-scoped**; if a `go.work` is ever adopted, the child ignores it (intended).
- **`moduleRoot`'s CWD-first + caller-fallback shape is correct as delivered** (the only failure case is a directly-invoked test binary from outside the module; the error is actionable).
- **N-a/N-b as recorded** now match the mechanism (exempt is symmetric/silent; the SCC's reachable witness is a **cross-target** cycle, a same-target cycle being caught first by the child `go list`).

### Fold review #3 (PR #95 review #3, comment `5721591781` — APPROVE WITH TWO SMALL REQUIRED FOLDS; both folded)

- **Fold 1 [TECHNICAL DEBT → folded] the test-import merge made the SCC fail on *legal* test-only cycles.** Merging `.TestImports`/`.XTestImports` into **one** graph means the acyclicity assertion also saw test edges — but Go permits cycles that exist only in tests (two same-tier governed packages whose **tests** import each other compile and test cleanly, yet tripped the SCC). **Fix:** `enumerate` now returns **two** edge sets — the **merged** graph (production + test) that the **rule** evaluates (test imports stay governed, Fold 1 of review #2 preserved), and the **production-only** graph that `cycles()` evaluates. The ADR D8 / [#93](https://github.com/gosharplite/tellme/issues/93) AC4 property is about code that must compile; the cross-target cycle the `CROSS_TARGETS` union exists to catch is a **production** phenomenon. **Witness:** two `internal/domain/zzc*` packages with mutually-importing `_test.go` ⇒ `go test` ok **and** `make verify-architecture` **green** (previously `import cycles detected (0 expected)`).
- **Fold 2 [CONSISTENCY → folded] the F-1 `-count=1` trap survived in the two *copied-from* documents.** `tools/arch/doc.go` and the `tools/arch/baseline.txt` header both showed the invocation **without** `-count=1` — i.e. exactly the vacuous path F-1 fixed (a cached run lets an added illegal import pass). **Fix:** both now document `go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch` and add `go vet -tags=arch ./tools/arch`; `tools/arch/baseline.txt` was regenerated (header-only change; the 8 content lines are unchanged).
- **Intended friction (recorded for R3/R4).** With the merged graph the **rule** now also governs an `internal/agent` **test** importing `internal/ui` — it reds the gate as **RULE-A**, which is *desirable*: it forces the loop→`ui` seam through a port rather than reproducing the coupling in test code. Recorded so R3/R4 meets it as a designed rule, not a surprise.

### Fold review #4 (PR #95 review #4 — CERTIFIED MERGE-READY after one fold; comment `5721639279`)

- **Fold 1 [TECHNICAL DEBT → folded] the `-count=1` trap survived in the *authority* artifacts.** `specs/truth/techstack.md` (the **Layer-discipline gate** row — `truth-current`, the artifact future rounds consult first) and `specs/plans/042-layer-discipline-gate/research.md` **D2** still recorded the **pre-F-1** invocation (`go test -tags=arch -run …`, no `-count=1`) — i.e. the vacuous cached path. **Fix (truth-owner `/axb-technical-research`):** the truth row now records the shipped invocation — `go vet -tags=arch ./tools/arch` + `golangci-lint run --build-tags=arch ./tools/arch/...` + `go test -count=1 -tags=arch -run TestVerifyRealArchitecture ./tools/arch` — with a matching **`truth-delta.md`** MODIFY row; `research.md` D2 is corrected in the same commit. **Witness:** the documented command now fails on an illegal import even on a warm cache (no `(cached)`).
- **N-1 [HARDENING → folded] the build-tagged guard was outside lint coverage.** `make lint` runs `golangci-lint run ./...` with no `--build-tags`, so `arch_test.go` was never compiled by the linter (only `go vet -tags=arch` saw it, from N-c). **Fix:** the `verify-architecture` target now also runs `golangci-lint run --build-tags=arch ./tools/arch/...` (guarded by the Makefile's existing `command -v` idiom) — the round owns that target, so it is not a change to an existing gate's semantics. Currently `0 issues`.
- **N-2 [NIT → folded] ledger labelling.** The fold-review sections are renumbered `#1…#4` (each naming its comment id) so a future round citing them does not hit duplicate `Fold 1`/`Fold 2` labels.

### Folds raised during execution (recorded)

- **F-1 (correctness — the Go test cache made the gate vacuous).** The first `Makefile` wiring used the ADR's literal command `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch`. Because the gate's verdict depends on the **whole module** (read via `go list`), which Go's test cache **cannot** see, a *cached* result let an added illegal import pass silently (witness (a) initially printed `ok (cached)` — the round-009 *"vacuous pass"* trap in gate form). **Fix**: the target now runs with **`-count=1`** (`go test -count=1 -tags=arch …`), which disables the test-result cache — mirroring the existing `verify-no-network` gate's `-count=1`. This is a **refinement** of the ADR's command, not a contradiction (the ADR is immutable; the mechanism detail is recorded here). Witness (a) was re-run uncached: **red**, naming `internal/domain/llm -> internal/infrastructure/mcp`.
- **F-2 (test-local).** `min()` (Go 1.21 builtin) is used for the Tarjan `low` update; `maps.Keys`/`slices.Sorted` give deterministic iteration order.

### Evidence

- Witness (a) — a temporary illegal import (`internal/domain/llm` importing `internal/infrastructure/mcp`) → `make verify-architecture` **FAIL**, naming `internal/domain/llm -> internal/infrastructure/mcp` (uncached); removed → green.
- Witness (b1) — a baseline line removed while its violation remains (`internal/cli -> internal/infrastructure/tools`) → **FAIL** (`1 new violation(s) not in the baseline`); restored → green.
- Witness (b2) — a bogus baseline line (`internal/domain/llm -> internal/config`, not a violation) → **FAIL** (`1 stale baseline entr(ies) … remove them from the baseline`); restored → green.
- Red-first: the gate with **no** baseline file → `baseline unreadable … an absent/unreadable baseline MUST fail` (**N-2** atomic-delivery).
- `make verify-architecture` green (8 baselined, 0 new, 0 stale, 0 cycles); `go test -count=1 ./...` green (`tools/arch` = `[no test files]` under default tags); `make verify` **OK** (existing gates + `verify-architecture`; golangci-lint 0 issues; govulncheck clean; cross-compile 4/4); topology audit **PASSED** and unchanged (44 features · 6 modules · 16 root + 327 module rows · 1674 steps); `gofmt -l .` clean; `go.mod`/`go.sum` unchanged; no `internal/**`/`cmd/**` changed.

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Build & Tooling（**Layer-discipline gate** row） | T001（依 predicate/tier 表實作）、T002、T003、T006、T007 | PASS |
| `specs/truth/techstack.md` -> Build & Tooling（**Task runner** row：`verify` aggregate 含 `verify-architecture`） | T003（接線）、T006（`make verify`）、T007 | PASS |
| `docs/decisions/0011-layer-discipline-gate.md`（rule + baseline policy + worked 8） | T001（D1–D10）、T002（worked example 對齊）、T007 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md` ×2 rows） | 已於 plan half 交付；T003／T006 驗證接線一致 | PASS |
| `truth-delta.md` -> Governance ADD（ADR 0011 + index） | 已於 plan half 交付；T001／T007 依 ADR 實作/驗證 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T006 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`specs/truth/features/cli/**`） | 豁免（NOOP 不建任務；T006 驗證 topology audit 不變） | PASS |
| `research.md` -> D1（兩段式 predicate over pinned ranking） | T001、T002、T004 | PASS |
| `research.md` -> D2（build-tagged Go guard + Makefile target） | T001、T003 | PASS |
| `research.md` -> D3（sorted baseline + fail-on-stale ratchet） | T001、T002、T005 | PASS |
| `research.md` -> D4（enumeration anchored to module root + graph self-test, B-2） | T001、T004 | PASS |
| `research.md` -> D5（CROSS_TARGETS union + filtered env, TD-1/F-1/R-4） | T001、T006 | PASS |
| `research.md` -> D6（custom tag scope out; withdrawn claim, TD-2） | T001、T007 | PASS |
| `research.md` -> D7（wiring + truth + ADR） | T001、T003、T007 | PASS |
| `research.md` -> D8（default-deny, TD-3） | T001、T005、T007 | PASS |
| `research.md` -> D9（one normative ranking source + coverage assert, RF-3） | T001、T007 | PASS |
| `research.md` -> D10（deterministic baseline format, RF-2） | T001、T002 | PASS |
| `research.md` -> D11（acyclicity asserted, TD-4） | T001、T005、T006 | PASS |
| `research.md` -> D12（BDD techstack unchanged; dsl-refine NOOP） | T006、T007 | PASS |
| `spec.md` -> US1（FR-001–FR-005, NFR-001–NFR-002） | T001、T003、T005 | PASS |
| `spec.md` -> US2（FR-006–FR-009, NFR-003） | T001、T002、T005 | PASS |
| `spec.md` -> US3（FR-010–FR-011, NFR-004） | T003、T006、T007 | PASS |
| `spec.md` -> 全域（FR-012, FR-013, NFR-005） | T006（no product code；make verify + audit） | PASS |
| `spec.md` -> SC-001（green on dev, 8 baselined） | T002、T006 | PASS |
| `spec.md` -> SC-002（new illegal import ⇒ red, witness） | T005(a) | PASS |
| `spec.md` -> SC-003（ratchet: removed entry ⇒ red；stale ⇒ red） | T005(b) | PASS |
| `spec.md` -> SC-004（host-independent union + 0 cycles + no new dep） | T001（union/SCC）、T006 | PASS |
| `spec.md` -> SC-005（truth + ADR + no product change + audit） | T006、T007 | PASS |
| PR #94 certification **N-1**（baseline-generation affordance） | T001（`-update-baseline`）、T002、T003（`verify-architecture-update`） | PASS |
| PR #94 certification **N-2**（absent/empty baseline ⇒ fail + self-test） | T001（`readBaseline` Fatalf）、T005(b) | PASS |
| PR #94 certification **N-3**（entry test covers all three properties + asserts they ran） | T001（entry test）、T007 | PASS |
| `plan.md` -> Structure（`Makefile` CHANGED；`tools/arch/` NEW；`techstack.md`/ADR 已 MODIFY+ADD；no product code） | T001、T003、T006 | PASS |
| `plan.md` -> Scope notes（api/data/dsl NOOP；ui/spec-by-example skipped） | T006、T007 | PASS |
| operator 拍板 Q1–Q3（broad→兩段式 predicate；scope via measurement；fail-on-stale） | T001（Q1/Q3）、T002、T005 | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
