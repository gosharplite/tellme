# Tasks: tellme cross-compile gate — every supported platform compiles (round 020)

**Plan Package**: `specs/plans/020-cross-compile-gate`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 build-pipeline（tooling）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D5；`go.mod`／`go.sum` 不動）。
  - **Phase 3 `Test Alignment` 為空** —— 本輪沒有新的 DSL 句，也沒有 stepdef；驗證由 gate 自身的 exit code 與可偽性見證承擔（非 Gherkin，`research.md` D6）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`Makefile` 的 `verify-cross-compile` gate**（含 `verify` aggregate 接線）與 **`SESSION-CLOSEOUT.md` 的引用**；唯一 truth 變更是 `specs/truth/techstack.md`（Build & Tooling，by `/axb-technical-research`）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與 `specs/truth/techstack.md` 對應 section。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** the `tellme` binary 的任何 `stdout`／`stderr` 行為、the class-phrase vocabulary、既有 gate（`verify-no-test-sleep`／`verify-no-network`／`vet`／`lint`／`vulncheck`）的語意。

## Round-020 locked decisions (implementation constraints — MUST)

> 來自 `research.md` Decisions 1–6 與 spec A1–A5。

- **[MATRIX]** `linux/amd64`、`linux/arm64`、`darwin/amd64`、`darwin/arm64`（POSIX；無 Windows），且 **host-independent**（不因 host 不同而少驗一個 target）。
- **[MECHANISM]** `Makefile` 的 `verify-cross-compile`：對每個 target 跑 `GOOS=<os> GOARCH=<arch> go build ./...` 後 `go vet ./...`，失敗即停並指出 target；接進 `verify` aggregate（+ `.PHONY` + `help`）。
- **[NO-DEP]** 只依賴 Go toolchain 與 `make`；無新相依；`go.mod`／`go.sum` 不動。
- **[WIRING]** `verify-cross-compile` 為 `make verify` 成員；`SESSION-CLOSEOUT.md` Step 2 引用之。
- **[TRUTH]** 唯一 truth 變更 = `specs/truth/techstack.md`（Build & Tooling：Cross-compile verification row + Task runner row）。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D5；無新相依，`go.mod`／`go.sum` 不動）。不得把 gate 骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無測試落點骨架（無新 DSL 句／stepdef）；gate 落點即 `Makefile`，由 T001 直接交付。

## Phase 3: Implementation (build pipeline)

**Goal**: 交付 host-independent cross-compile gate 並完成接線與文件，使 build-tagged 平台碼不可能對某個 shipped target 靜默編譯失敗。

- [X] T001 在 `Makefile` 新增 `verify-cross-compile` gate 並接進 `verify`
  - Read:
    - `specs/truth/techstack.md` -> Build & Tooling（Cross-compile verification；Task runner）
    - `specs/plans/020-cross-compile-gate/research.md` -> Decision 1, Decision 2, Decision 3, Decision 5
    - `Makefile`（既有 gate 的風格：`verify-no-test-sleep`、`verify` aggregate、`.PHONY`、`help`）
  - 做：
    - 新增 `CROSS_TARGETS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64`。
    - 新增 `verify-cross-compile` target：對 `$(CROSS_TARGETS)` 逐一 `GOOS=<os> GOARCH=<arch> go build ./...` 後 `go vet ./...`；任一失敗即 `exit 1` 並在訊息中指出該 `<os>/<arch>`。
    - 將 `verify-cross-compile` 加入 `verify` aggregate、`.PHONY` 與 `help` 文字（與既有 target 一致）。
  - 不做：不改既有 target 的行為；不加相依；不碰 `cmd/**`／`internal/**` 產品碼；不引入 `time.Sleep`。

- [X] T002 在 `SESSION-CLOSEOUT.md` 引用 cross-compile gate
  - Read:
    - `SESSION-CLOSEOUT.md` -> Step 2「Quality Gates」項目
    - `specs/plans/020-cross-compile-gate/research.md` -> Decision 4
  - 做：在 Step 2 的 gate 清單加入 cross-compile gate（`make verify-cross-compile` / 其為 `make verify` 成員）。
  - 不做：不重構 closeout 文件其餘內容。

## Phase 4: Verification & Regression

**Goal**: 證明 gate 對全部 supported target 綠燈、接線生效、且可被證偽（非真空），並確認既有行為未受影響。

- [X] T003 [REGRESSION] 執行 gate、aggregate 與可偽性見證
  - Read:
    - `specs/plans/020-cross-compile-gate/research.md` -> Decision 1, Decision 3, Decision 4, Decision 5
    - `Makefile`
  - 做：
    - `make verify-cross-compile` → 對四個 target 全綠。
    - `make verify` → 綠（含新成員）。
    - **可偽性見證**：在一個 **non-host** target 的 build-tagged 檔（host = `darwin/arm64`，故取 `internal/infrastructure/telemetry/system_metrics_linux.go`）暫時注入一個編譯錯誤 → `make verify-cross-compile` 必須 **非零退出** 並指出 `linux/*`；觀察到失敗即**還原**，再跑一次確認綠燈。
    - `gofmt -l .` clean；確認 `go.mod`／`go.sum` 不變。
  - 不做：不為了讓 gate 綠而放寬任何既有 assertion 或移除任何 target。

- [ ] T004 subagent review (round quality gate)
  - Read:
    - `Makefile`、`SESSION-CLOSEOUT.md`、`specs/truth/techstack.md`
    - `specs/plans/020-cross-compile-gate/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：gate 覆蓋四個 target 且 host-independent；`verify` 含之；失敗會指出 target；無新相依；`techstack.md` 與 `truth-delta.md` 一致；既有 gate 語意未變；`stdout`／`stderr` 行為未變。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Build & Tooling（Cross-compile verification row） | T001（交付 gate）、T003（驗證）、T004（review） | PASS |
| `specs/truth/techstack.md` -> Build & Tooling（Task runner row：aggregate 含 `verify-cross-compile`） | T001、T004 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY（`techstack.md`） | T001、T003、T004 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP（`specs/truth/contracts/**`） | 豁免（NOOP 不建任務；T003 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP（`specs/truth/data/**`） | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP（`specs/truth/features/cli/**`） | 豁免（NOOP 不建任務；無 CLI interface 變更） | PASS |
| `research.md` -> Decision 1（host-independent POSIX matrix） | T001、T003 | PASS |
| `research.md` -> Decision 2（`Makefile` target wired into `verify`） | T001 | PASS |
| `research.md` -> Decision 3（build + vet per target, fail fast） | T001、T003 | PASS |
| `research.md` -> Decision 4（wiring + closeout reference） | T001、T002 | PASS |
| `research.md` -> Decision 5（no new dependency；deterministic） | T001、T003 | PASS |
| `research.md` -> Decision 6（must-asks / no interface change） | T004、T003（既有行為不變） | PASS |
| `spec.md` -> US1（FR-001–FR-004, NFR-001–NFR-002） | T001、T003 | PASS |
| `spec.md` -> US2（FR-005–FR-007, NFR-003） | T001、T002、T004 | PASS |
| `spec.md` -> 全域（FR-008, NFR-004） | T001（no new dep）、T003（既有行為不變） | PASS |
| `spec.md` -> SC-001–SC-004 | T001、T003（SC-001/002/003）、T004（SC-004） | PASS |
| `plan.md` -> Structure（`Makefile` CHANGED；`SESSION-CLOSEOUT.md` CHANGED；`techstack.md` MODIFY；no product code） | T001、T002 | PASS |
| `plan.md` -> Scope notes（api NOOP；data NOOP；dsl-refine NOOP；`/axb-ui-plan`、`/axb-spec-by-example` skipped） | T003（不觸及 API/資料/CLI）、T004 | PASS |
| operator 拍板（A1 POSIX matrix；"do B now"） | T001（matrix）、T003（witness） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
