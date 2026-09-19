# Tasks: tellme domain model + drift gate (round 060)

**Plan Package**: `specs/plans/060-domain-model-and-drift-gate`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/0030-domain-model-and-modelith-toolchain.md`

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 docs + build/quality-pipeline（tooling）輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增 Go 相依（`research.md` D11；`go.mod`／`go.sum` 不動）。`modelith` 為 **dev-tool 二進位**，非 module 相依。
  - **Phase 3 沒有新的 DSL 句／stepdef** —— 本輪不改 `specs/truth/features/**`；驗證由 **`modelith lint`／`render --check`** 與可偽性見證（drift ⇒ 紅；binary 缺席 ⇒ 紅）承擔（非 Gherkin，`spec.md` A3 / `research.md` D10）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更。
  - 交付物是 **`docs/domain-model/` 三份模型 + `Makefile` `modelith-*` 接線**；truth 變更（`specs/truth/techstack.md` Domain model row + Task runner row + Layer-discipline gate row 修正）與 **ADR 0030** 已由 `/axb-technical-research` 於 plan half 交付。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與（若引用）`specs/truth/techstack.md` / ADR 對應段落。
- **不動**：任何 `stdout`／`stderr` 行為、CLI flag／exit code、既有 gate 語意、任何 `internal/**`／`cmd/**`／`tests/**`。

## Round-060 locked decisions (implementation constraints — MUST)

> 來自 `research.md` D1–D12、`spec.md`（FR-001–010 / SC-001–005）與 **ADR 0030**。

- **[SCOPE / Q1, D1/D2]** 產出**三份**模型於 `docs/domain-model/`：**product**（`tellme.modelith.*`）· **quality**（`quality.modelith.*`）· **environment-management**（`environment-management.modelith.*`，描述**外部** Niffler `tellme.sh`，記錄為 divergence）。
- **[TOOLCHAIN / Q2, D1]** YAML 為唯一真相源；`.md` 由 modelith **產生**（never hand-edit）。modelith = `gosharplite/modelith` fork `@feat/self-domain-model`，**dev-tool 二進位**（`$GOPATH/bin`），**不入 `go.mod`**。
- **[GATE / Q3, D5]** `Makefile` 新增 `modelith-lint`／`modelith-render`／`modelith-check`；`modelith-check` 為 `verify` 的 **zero-tolerance** 成員：**drift ⇒ fail**；**binary 缺席 ⇒ fail**（具名 **安裝路徑**：clone + pinned build，見 `docs/domain-model/README.md`；非 `go install @path`）。**無** `go run …@branch` fallback（hermetic）。
- **[BOUNDARY / Q4, D8]** 模型為**描述性 docs，非 truth**；與 `specs/truth/**` 衝突時 **truth 勝**。drift gate 只保 **model-internal**（YAML ↔ `.md`）一致；reference 的 `modelith-drift`／`modelith-layers` **不採用**。
- **[LIFECYCLE / Q5, D9]** 隨 truth 變更刷新；無排程 pass；模型無 `delivered` freeze。
- **[CONVENTIONS / D1]** PascalCase entity keys；freeform 以 backtick 包 entity；`cardinality ∈ {1:1,1:n,n:1,n:n}`；`ownership ∈ {owned,referenced}`；3-pass build order。
- **[NO-DEP]** `go.mod`／`go.sum` 不動；POSIX-only；不改任何既有 gate 語意。

---

## Phase 1: Setup

**Omitted** —— 無新增技術／相依（`research.md` D11）。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點骨架；模型落點即 `docs/domain-model/`（由 T002–T004 直接交付）。

## Phase 3: Test Alignment & Implementation (RED-first: the gate must red on a missing model / drift)

**Goal**: 交付三份模型 + `docs/domain-model/README.md` + `Makefile` `modelith-*` 接線，使 `make verify` 在具備 modelith 的主機上綠燈、YAML 漂移紅燈、binary 缺席紅燈。

**DSL 參照**: 本輪無 Gherkin／`dsl.md` row（`/axb-dsl-refine` NOOP）；驗證語言為 `modelith` 指令 + gate exit code。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Build & Tooling（**Domain model** row + **Task runner** row + Layer-discipline gate row 修正）
- `docs/decisions/0030-domain-model-and-modelith-toolchain.md`（D1–D6；adoption + ADR 0011 D10 amendment）
- `specs/plans/060-domain-model-and-drift-gate/research.md` -> D1–D12
- `specs/plans/060-domain-model-and-drift-gate/plan.md` -> Pinned decisions + 預期變更（`docs/domain-model/**` NEW content；`Makefile` CHANGED）

**Boundary**:
- 只新增 `docs/domain-model/**` 與修改 `Makefile`；**不**碰產品碼、**不**改既有 gate 語意、**不**改既存 truth（除已於 plan half 交付者）。
- modelith 版本／`@branch` 釘在 `docs/domain-model/README.md`（D6）。
- review 啟動 subagent；通過前不解鎖 Phase 4。

- [X] T001 `Makefile` 接線：`modelith-lint`／`modelith-render`／`modelith-check`（`modelith-check` 進 `verify`；binary 缺席 → 具名安裝指令 + exit 1；無 network fallback）
  - Read: `research.md` D5/D6/D11；`spec.md` FR-004/FR-006；`Makefile`（既有 gate 風格 + `.PHONY`/`help`）
  - RED-first 期望：尚無任何 `docs/domain-model/*.modelith.yaml` 時，`make modelith-check` 報錯（缺口徑）——證明 gate 未接線即不綠。

- [X] T002 產品模型 `docs/domain-model/tellme.modelith.yaml`（skeleton→behaviour→refinement；實體：`Session`/`Turn`/`Provider`/`Tool`/`ToolCall`/`Context`/`History`/`Skill`/`MCPServer`/`MCPTool`/`Config`/`Chrome`；含 invariants + scenarios）
  - Read: `research.md` D3；`spec.md` FR-001/FR-002/FR-003(a)；the shipped tree（`internal/**`）以為實證；README *Design Intent & Direction*（**不得**含 security/Windows/`pipe_commands`）
  - 做: 3-pass；`modelith lint` 0/0；`modelith render` 產出 `.md`。

- [X] T003 品質模型 `docs/domain-model/quality.modelith.yaml`（`QualityPipeline` + 既有 gate catalog + ADR governance + triage；**記錄**無 `NonFixCatalog` 之 divergence）
  - Read: `research.md` D4；`spec.md` FR-003(b)；`Makefile`（實際 gate 清單）；`docs/decisions/README.md`
  - 做: 3-pass；lint 0/0；render。

- [X] T004 環境模型 `docs/domain-model/environment-management.modelith.yaml`（`Environment`/`Group`/`Persona`/`SecretSet`/`TellMeHome`；hot-swap；**描述外部 Niffler `tellme.sh`**，`description` 記錄 divergence）
  - Read: `research.md` D2；`spec.md` FR-003(c)；reference `environment-management.modelith.*`（作形狀參考，非抄）；`$TELL_ME_HOME` 佈局
  - 做: 3-pass；lint 0/0；render。

- [X] T005 `docs/domain-model/README.md`（作者慣例 + 釘住 `go install …@feat/self-domain-model` + 觀測版本 + `make modelith-*` 用法 + 「descriptive docs, not truth」）
  - Read: `research.md` D2/D6/D8；reference `docs/domain-model/README.md`（形狀參考）

- [X] T006 全量驗證 + 見證（`modelith lint` 三檔 0/0 · `modelith render --check` 三檔 up-to-date · `make verify` 綠 · drift 見證：改 YAML 不 re-render ⇒ `modelith-check` 紅 → revert · absent-binary 見證：以 path-shim 令 `command -v modelith` 空 ⇒ `modelith-check` 紅並具名安裝指令 → revert）
  - Read: `spec.md` SC-001…SC-005；`research.md` D5/D7；ADR 0010（見證再現後 revert）

## Phase 4: Regression & Verification

- [X] T007 Regression：`make verify`（含所有既有成員）+ `go test -count=1 ./...` + Gherkin/DSL topology audit 皆綠；`go.mod`／`go.sum` 未動；產品碼未動
  - Read: `spec.md` NFR-004；`research.md` D11/D12
- [X] T008 回寫 `tasks.md` `[X]`；`truth-delta.md` 無新增（plan half 已交付）；準備交付記錄

---

## Implementation ledger (round 060)

- **T001** — `Makefile`: `MODELITH`/`MODELITH_INSTALL`/`MODELITH_MODELS` vars; `.PHONY` + `help` entries; `modelith-lint`/`modelith-render`/`modelith-check` targets (POSIX, no network fallback); `modelith-check` added to the `verify` aggregate.
- **T002–T004** — three models authored (3-pass) — all `modelith lint` → **0 errors / 0 warnings**; all rendered:
  - `docs/domain-model/tellme.modelith.yaml`/.md (15 entities: `Config`/`Provider`/`Pricing`/`Session`/`Turn`/`ToolCall`/`Tool`/`MCPServer`/`MCPTool`/`Skill`/`Context`/`Persona`/`History`/`UsageRecord`/`TurnLog`/`ToolUsageRecord`/`PromptLog`/`Chrome`/`PromptInput`; 9 scenarios).
  - `docs/domain-model/quality.modelith.yaml`/.md (`QualityPipeline`/`QualityGate`/`E2ESuite`/`TopologyAudit`/`ADR`/`DecisionIndex`/`Residual`/`Round`; `NonFixCatalog` recorded as a divergence; 5 scenarios).
  - `docs/domain-model/environment-management.modelith.yaml`/.md (`Environment`/`Group`/`Persona`/`SecretSet`/`InGroupChat`; the **external** Niffler manager recorded; 4 scenarios).
- **T005** — `docs/domain-model/README.md` (toolchain install + observed version pin; conventions incl. the "plain scalar must not start with a backtick" rule; the descriptive-docs-not-truth boundary; lifecycle).
- **T006 — verification + witnesses** (reproduced then reverted, ADR 0010 doctrine):
  - **drift witness (a)**: edit `tellme.modelith.yaml` without re-render ⇒ `make modelith-check` **FAILS** (`… .md is out of date — regenerate it …`) — reverted ⇒ green.
  - **absent-binary witness (b)**: `PATH=/usr/bin:/bin make modelith-check` ⇒ **FAILS** naming `go install github.com/gosharplite/modelith/cmd/modelith@feat/self-domain-model` — restored ⇒ green. *(**Superseded by the PR #126 fold `2f59f91`**: the message now names the clone + pinned-build route; the quoted form does not resolve. Kept as the pre-fold witness history.)*
  - `make modelith-lint` → 0/0 (three models); `make modelith-check` → all three up to date.
- **T007 — regression**: `make verify` **OK** (incl. the new `modelith-check` member; `golangci-lint` 0 issues; `govulncheck` no reachable vulns; cross-compile 4/4); `go test -count=1 ./...` **green** (all packages incl. the godog E2E); `gofmt -l .` clean; `go.mod`/`go.sum` **unchanged**; no product code touched; no `specs/truth/features/**` change (topology audit unchanged).
- **T008** — markers flipped; `truth-delta.md` unchanged from the plan half (truth + ADR already delivered there).
- **Companion (post-tasks, operator-requested)** — `SESSION-BOOTSTRAP.md` updated so **Step 1** reads tellme's **three domain models** immediately after `README.md` (a new step-1 detail section + Agent Rule 10). This **must land with round 060**: the bootstrap references `docs/domain-model/*.modelith.md`, which exist only once this round merges.

---

## Fold ledger — PR #126 review (5254902427; REQUEST CHANGES)

**Verdict folded**: B-060-1 (blocker) + TD-060-1…TD-060-4 (+ 3 nits).

| # | Fold |
| --- | --- |
| **B-060-1** (blocker) | The documented (and gate-printed) `go install github.com/gosharplite/modelith/cmd/modelith@feat/self-domain-model` **does not resolve** — reproduced independently: the query form is rejected (`invalid version … disallowed version string`) and the pseudo-version form fails (`module declares its path as: github.com/stacklok/modelith`). Folded to a **single-sourced clone + pinned-build route** (`git clone … && git checkout b4153541cee8 && go install ./cmd/modelith`), **executed** (built the tool from the clone: `modelith version v0.0.0-20260815121344-b4153541cee8`), single-sourced in `docs/domain-model/README.md`, quoted by `$(MODELITH_INSTALL)` + the gate's failure message, cited by ADR 0030 D2 and the truth row (5 surfaces re-aligned). The upstream-path fact is recorded in ADR 0030 (why the clone route is needed). |
| **TD-060-1** | Immutable pin (commit `b4153541cee8`) named by the gate's failure message; branch-tracking is an explicit documented upgrade; ADR note records that a fork move reds `verify` on re-install. Folded with B-060-1. |
| **TD-060-2** | `MODELITH_MODELS := $(wildcard docs/domain-model/*.modelith.yaml)` + a **non-empty assertion** in every target (default-deny). Witness: adding a 4th `*.modelith.yaml` is covered by construction (the gate reds on its missing `.md`); an empty set fails. |
| **TD-060-3** | `SESSION-BOOTSTRAP.md` now **owned by a requirement** — added to `plan.md`'s tree + **FR-011** + `truth-delta.md` companion rows. |
| **TD-060-4** | ADR 0011's **Status** line + its index row carry a **forward pointer** to ADR 0030 (body untouched). |
| **nit 1** | `skill-unique-name` restated to the shipped mechanism (duplicate ⇒ first discovered wins, dropped silently). |
| **nit 2** | README provenance row reframed (provenance, not a claim about the gate's supported tool). |
| **nit 3** | Pre-existing `staticcheck` truth row corrected (`command -v` only — the `$GOPATH/bin` fallback was never in the Makefile). |

**Re-verified after the fold**: `make modelith-lint` 0/0 ×3 · `make modelith-check` all up to date · **witness (b)** absent binary ⇒ fails naming the **clone route** · **TD-060-2 witness** a 4th model auto-covered (gate reds on its missing `.md`) ⇒ reverted green · `make verify` **OK** · `go test -count=1 ./...` green · `gofmt` clean · `go.mod`/`go.sum` unchanged.

### Fold-verification fold-back (PR #126 comment `5740054490`; **TF-060-1**)

**FOLDS VERIFIED 5/5 (B-060-1 + TD-060-1…4 + 3 nits)**, with one required claim-surface fold-back: the fold re-aligned 5 **outward** surfaces but left in-package ones naming the falsified command.

| # | Surface | Fold |
| --- | --- | --- |
| 1 | `spec.md` Edge Cases | install command → the **clone + pinned-build** route (FR-004). |
| 2 | `plan.md` Fork pin | pin restated as the **immutable commit `b4153541cee8`** + the clone route (cites the README single owner). |
| 3 | `tasks.md` locked-decisions block | the MUST constraint now names the **install route** (clone + pinned build), not `go install @path`. |
| 4 | `plan.md` Gate wiring | `$(wildcard docs/domain-model/*.modelith.yaml)` + a **non-empty assertion** (TD-060-2 drift). |
| — | `spec.md` FR-004 (live requirement) | aligned: names the **install route** + the **immutable commit pin**. |
| hist | `spec.md` Q3 line · `tasks.md` pre-fold witness · `research.md` D5 fold note | **not rewritten**; a *"superseded by the fold `2f59f91`"* forward pointer added where it names the falsified form. |

**Residuals (recorded, non-blocking)**: **R-060-1** the ADR 0030 copy of the route is static while the Makefile's derives from `$(MODELITH_PIN)` (defensible — an ADR is self-contained) · **R-060-2** `$(wildcard)` returns directory order (harmless) · **R-060-3** witness executions live in this ledger (TD-6 convention).

**Re-verified after TF-060-1**: `make verify` **OK** · `go test -count=1 ./...` **green** · `gofmt` clean · `go.mod`/`go.sum` unchanged. Fold-back head: *see the PR comment*.
