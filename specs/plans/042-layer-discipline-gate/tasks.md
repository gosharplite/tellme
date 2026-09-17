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
- **[N-2]** **absent / unreadable / empty / 不可解析** 的 baseline → **fail**（**never** 當「未設定 baseline」）；self-test 斷言 committed baseline 非空且每行可解析為 ` -> `。
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

- [ ] T001 建立 `tools/arch/` guard（`//go:build arch` test + untagged `doc.go`）
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

- [ ] T002 產生並提交 `tools/arch/baseline.txt`（generated，非手抄）
  - Read:
    - `specs/plans/042-layer-discipline-gate/research.md` -> D3, D10
    - `docs/decisions/0011-layer-discipline-gate.md` -> D3, D9（worked example）
    - `specs/plans/042-layer-discipline-gate/spec.md` -> SC-001（8 violations）
  - 做：以 `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline` **生成** `tools/arch/baseline.txt`（不得手抄）；確認內容為 sorted、ASCII ` -> `、module-relative 的 **8** 行（7 `internal/cli -> internal/infrastructure/*` + `internal/agent -> internal/ui`），且與 ADR 0011 的 worked example 逐行一致；檔頭含 `make verify-architecture-update` 生成法（N-1）。
  - 驗證：重跑 `go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch` → **綠**（baseline 與 gate 輸出相符）。
  - 不做：不手改任何 baseline 行；不為了綠而移除違規。

- [ ] T003 `Makefile`：新增 `verify-architecture` + `verify-architecture-update` 並接入 `verify`
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling（Layer-discipline gate row；Task runner row）
    - `docs/decisions/0011-layer-discipline-gate.md` -> Consequences（New）
    - `Makefile`（既有 gate 風格：`verify-cross-compile`、`verify` aggregate、`.PHONY`、`help`）
  - 做：
    - 新增 `verify-architecture` target：`@go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch`（+ echo 說明）。
    - 新增 `verify-architecture-update` target：`@go test -tags=arch -run TestVerifyRealArchitecture ./tools/arch -args -update-baseline`（N-1 affordance）。
    - 將 **`verify-architecture`** 加入 `verify` aggregate（置於 `verify-cross-compile`/`verify-mcp-sdk-confinement` 之後，維持 warm cache）、`.PHONY` 與 `help` 文字；`verify-architecture-update` 只進 `.PHONY` + `help`，**不**進 `verify`。
  - 不做：不改既有 target 行為；不把 `-update-baseline` 放進 `verify`（避免 gate 自我改寫 baseline）；不加相依。

- [ ] T004 subagent review (phase quality gate)
  - Read: `tools/arch/doc.go`、`tools/arch/arch_test.go`、`tools/arch/baseline.txt`、`Makefile`、`research.md` -> D1–D12、`docs/decisions/0011-layer-discipline-gate.md`
  - 檢驗：predicate 為兩段式且 **能導出** 8 行 baseline；enumeration 錨定 module root 且 self-test 斷言 graph 本身；union over `CROSS_TARGETS`；env 過濾含 drop **與** preserve 兩集；baseline 缺／空／不可解析 → fail；entry test 涵蓋三性質並顯式斷言；無 undefined 行為；未動產品碼／truth／既有 gate。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4: Verification & Regression

**Goal**: 證明 gate 在 `dev` 上綠、對**新**違規與**stale** 條目皆紅（可偽性）、既有 gates／行為未受影響。

**Test Scope**: `tools/arch/**`（unit；`-tags=arch`）；whole-suite `go test ./...`；`make verify`；Gherkin/DSL topology audit。

- [ ] T005 [REGRESSION] 可偽性見證 (a) 新違規 ⇒ 紅 / (b) stale 條目 ⇒ 紅
  - Read: `research.md` -> D3, D11；`docs/decisions/0010-test-deadline-decoupling.md`（見證 doctrine）；`tools/arch/arch_test.go`
  - 做：
    - **(a) 新違規**：暫時在一個受治理套件加入一個 illegal import（例如於 `internal/domain/llm` 暫時 import `internal/config`；或於 `internal/cli` 暫時 import 另一個 `internal/infrastructure/*`）→ `make verify-architecture` **非零失敗並列出** 該 `src -> dst`；**還原**後確認綠。
    - **(b) stale**：暫時自 `tools/arch/baseline.txt` 移除一行（其違規仍在）→ gate 回報**新 violation** 而紅；改成暫時修掉一個 baselined 違規（或自 baseline 保留一行卻讓該行不再違規）→ gate 回報 **stale** 而紅；兩種皆觀察到即**還原**，重跑確認綠。
  - 不做：不放寬任何 assertion 以「讓它過」；見證後必須還原到 HEAD。

- [ ] T006 [REGRESSION] 全量回歸 + `make verify` + topology audit + 範圍檢查
  - Read: `research.md` -> D3, D7, D12；`specs/truth/techstack.md` -> 兩 row
  - 做：
    - `make verify-architecture` → 綠（8 baselined、0 new、0 stale、0 cycles）。
    - `go test -count=1 ./...` 全綠（unit + godog E2E；`tools/arch` 於預設 tags 下為 `[no test files]`）。
    - `make verify` **OK**（既有 gates + 新成員 `verify-architecture`）。
    - Gherkin/DSL topology audit 重跑 → **unchanged**（44 features · 16 root + 327 module rows · 1674 steps）。
    - `gofmt -l .` clean；`git diff --name-only origin/dev..HEAD` 確認僅動 `Makefile` + `tools/arch/**`（+ 文件）；`go.mod`／`go.sum` 不變；`internal/**`／`cmd/**` 產品碼未動。
  - 不做：不為了綠而放寬 assertion、改產品碼或 baseline 手抄。

- [ ] T007 subagent review (round quality gate)
  - Read:
    - `tools/arch/doc.go`、`tools/arch/arch_test.go`、`tools/arch/baseline.txt`、`Makefile`
    - `specs/truth/techstack.md`（Layer-discipline gate row；Task runner row）、`docs/decisions/0011-layer-discipline-gate.md`
    - `specs/plans/042-layer-discipline-gate/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：`dev` 上綠；**新違規 ⇒ 紅**；**stale ⇒ 紅**；baseline absent/empty ⇒ fail；三性質皆有跑（N-3）；tier 表為唯一權威且與 `techstack.md` predicate 一致；env 過濾含 drop+preserve；union 四 target；無新相依；既有 gate 語意與 `stdout`/`stderr` 未變；**無產品碼變更**；`techstack.md` 與 `truth-delta.md` 一致。

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
