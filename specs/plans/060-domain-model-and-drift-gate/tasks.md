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
- **[GATE / Q3, D5]** `Makefile` 新增 `modelith-lint`／`modelith-render`／`modelith-check`；`modelith-check` 為 `verify` 的 **zero-tolerance** 成員：**drift ⇒ fail**；**binary 缺席 ⇒ fail**（具名 `go install …@feat/self-domain-model`）。**無** `go run …@branch` fallback（hermetic）。
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

- [ ] T001 `Makefile` 接線：`modelith-lint`／`modelith-render`／`modelith-check`（`modelith-check` 進 `verify`；binary 缺席 → 具名安裝指令 + exit 1；無 network fallback）
  - Read: `research.md` D5/D6/D11；`spec.md` FR-004/FR-006；`Makefile`（既有 gate 風格 + `.PHONY`/`help`）
  - RED-first 期望：尚無任何 `docs/domain-model/*.modelith.yaml` 時，`make modelith-check` 報錯（缺口徑）——證明 gate 未接線即不綠。

- [ ] T002 產品模型 `docs/domain-model/tellme.modelith.yaml`（skeleton→behaviour→refinement；實體：`Session`/`Turn`/`Provider`/`Tool`/`ToolCall`/`Context`/`History`/`Skill`/`MCPServer`/`MCPTool`/`Config`/`Chrome`；含 invariants + scenarios）
  - Read: `research.md` D3；`spec.md` FR-001/FR-002/FR-003(a)；the shipped tree（`internal/**`）以為實證；README *Design Intent & Direction*（**不得**含 security/Windows/`pipe_commands`）
  - 做: 3-pass；`modelith lint` 0/0；`modelith render` 產出 `.md`。

- [ ] T003 品質模型 `docs/domain-model/quality.modelith.yaml`（`QualityPipeline` + 既有 gate catalog + ADR governance + triage；**記錄**無 `NonFixCatalog` 之 divergence）
  - Read: `research.md` D4；`spec.md` FR-003(b)；`Makefile`（實際 gate 清單）；`docs/decisions/README.md`
  - 做: 3-pass；lint 0/0；render。

- [ ] T004 環境模型 `docs/domain-model/environment-management.modelith.yaml`（`Environment`/`Group`/`Persona`/`SecretSet`/`TellMeHome`；hot-swap；**描述外部 Niffler `tellme.sh`**，`description` 記錄 divergence）
  - Read: `research.md` D2；`spec.md` FR-003(c)；reference `environment-management.modelith.*`（作形狀參考，非抄）；`$TELL_ME_HOME` 佈局
  - 做: 3-pass；lint 0/0；render。

- [ ] T005 `docs/domain-model/README.md`（作者慣例 + 釘住 `go install …@feat/self-domain-model` + 觀測版本 + `make modelith-*` 用法 + 「descriptive docs, not truth」）
  - Read: `research.md` D2/D6/D8；reference `docs/domain-model/README.md`（形狀參考）

- [ ] T006 全量驗證 + 見證（`modelith lint` 三檔 0/0 · `modelith render --check` 三檔 up-to-date · `make verify` 綠 · drift 見證：改 YAML 不 re-render ⇒ `modelith-check` 紅 → revert · absent-binary 見證：以 path-shim 令 `command -v modelith` 空 ⇒ `modelith-check` 紅並具名安裝指令 → revert）
  - Read: `spec.md` SC-001…SC-005；`research.md` D5/D7；ADR 0010（見證再現後 revert）

## Phase 4: Regression & Verification

- [ ] T007 Regression：`make verify`（含所有既有成員）+ `go test -count=1 ./...` + Gherkin/DSL topology audit 皆綠；`go.mod`／`go.sum` 未動；產品碼未動
  - Read: `spec.md` NFR-004；`research.md` D11/D12
- [ ] T008 回寫 `tasks.md` `[X]`；`truth-delta.md` 無新增（plan half 已交付）；準備交付記錄
