# Tasks: 033 — skills system (`list_skills`: load & list the workspace's skills)

**Plan Package**: `specs/plans/033-skills-system`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `specs/truth/features/cli/**`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**) and `specs/truth/data/**` (`/axb-data-plan` **NOOP**) are unchanged this round. There is no `ui/**` (the skills surface is model-facing).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是 BDD feature 輪**：`/axb-dsl-refine` **ADD** 一個 interface feature (`chat/listing-the-available-skills.feature`) + 11 DSL rows，並 **MODIFY** `chat/offering-the-agent-tools.feature`。`/axb-api-plan` 與 `/axb-data-plan` 皆為 **NOOP**。
- **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` Decision 9；stdlib-only，`go.mod`／`go.sum` 不動）。
- Phase 2 `Foundational` 只建立落點骨架（stepdef 檔案、產品骨架、unit 測試骨架），每則寫「只做／不做」。
- Phase 3 `Test Alignment & Implementation`：把本輪所有受影響 DSL 的自動化測試對齊最新版 truth，**不寫產品碼**。
- Phase 4A/4B：ADD / MODIFY feature 的 `[BDD-GREEN] -> [BDD-REFACTOR]`。
- Truth 參照必須使用 `specs/truth/**` 路徑；plan 參照才使用當前 plan package 相對路徑。
- **不動** the class-phrase vocabulary 或 exit code、`stdout`/`stderr` 既有契約、provider transport，或已 shipped 的六個工具的行為。

## Round-033 locked decisions (implementation constraints — MUST)

> 來自 `research.md` Decisions 1–9 與 spec US1/US2（FR-001..FR-009、NFR-001..NFR-006）。

- **[SOURCE]** 只從 `<TELL_ME_HOME>/docs/skills/` 讀取（單一來源；**無** `.skills/`、**無** `skillssh` tool、**無** 網路）；透過既有 runtime-home resolver + 一個注入的 home seam（`research.md` Decision 1）。
- **[DISCOVERY]** 遞迴 walk；一個檔案是 skill ⟺ 內容以 `---` frontmatter 區塊宣告 **`name`** 與 **`description`**（CRLF 正規化）；無有效 frontmatter 的 `.md`（如 `rules/*.md`）**不是** skill；重名 → 首見者勝 + warning；缺/空/不可讀目錄 → **空 catalog**；個別壞檔 → best-effort skip，**不得** fail the run（`research.md` Decision 2；`specs/truth/techstack.md` -> Skills catalog row）。
- **[SURFACE-ONLY]** **on-demand only**：一個 **read-only** `list_skills` tool，回傳每個 skill 的 **name** + **description** + **location**（path），**paths 排序** 決定性輸出；空 catalog 回報為**結果**（非 error）；**不注入** 任何 skill 內容（`research.md` Decision 3/6；Q1/Q3）。
- **[TOOL]** `list_skills` 是既有 `domain/tools.Tool` port 的 adapter：以 shared `resourceSchema` builder 建 schema（帶 `reason` + resource params；`required ⊆ properties`），宣告 **local-reader 30 s** `ToolContract` default、在 source 端 bound 自身輸出、觀測到 deadline 回 **nil-error timeout result**（FR-018）；註冊進 **production assembler `agentTools()`**（`research.md` Decision 7）。
- **[LOAD-PATH]** catalog 只在 **prompt-bearing turn path** 載入；offline paths（`--version`/`-d`/`--tool-usage`/prompt-less `--new`/boot）**不變**、不觸及 `docs/skills`（`research.md` Decision 4；FR-009）。
- **[CONTENT]** 讀取 skill 內容 **reuse 既有 `read_files`**（**無** 新 read tool；NFR-002）。
- **[OUT-OF-SCOPE]** **無** skills.sh `.skills/`、`search/install/remove_skill`、cross-source merging、automatic injection（recorded divergence；`research.md` Decision 8）。
- **[NO-DEP]** 只用 Go 標準庫（`os`、`path/filepath`、`strings`、`encoding/json`）；無新相依；`go.mod`／`go.sum` 不動；POSIX-only。
- **[VERIFY]** 驗證 **hermetic**（離線；`research.md` Decision 9）；falsifiability witnesses：(a) 移除 frontmatter 規則 → non-skill-`.md` scenario 失敗；(b) 注入 skill 內容進 request → no-injection 斷言失敗。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` Decision 9；無新相依，`go.mod`／`go.sum` 不動）。不得把落點骨架或 helper 塞進 Setup。

## Phase 2: Foundational

**Goal**: 建立後續 Phase 3 測試層與 Phase 4 產品碼的落點骨架（零共享寫入衝突）。

**Shared Must Read**:
- `specs/truth/features/cli/chat/listing-the-available-skills.feature` -> `Feature: Listing the available skills`
- `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 033)` + `## Then (round 033)`
- `specs/plans/033-skills-system/research.md` -> `Decision 1`–`Decision 7`, `Decision 9`
- `specs/truth/techstack.md` -> Skills catalog (load) row + Skills listing tool (`list_skills`) row

**Boundary**:
- 只建立骨架 / 落點 / 入口；**不寫** DSL 語意、**不寫** 產品行為、**不寫** 測試 assertion。

- [ ] T001 stepdef 落點骨架（一 task 一檔案，Zero Shared Edits）
  - Read:
    - `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 033)` + `## Then (round 033)` 各列
    - `tests/e2e/steps/` 既有 stepdef 檔（命名與 `init()` 自動註冊慣例）
  - 只做：為 `## Given (round 033)`（5 句）與 `## Then (round 033)`（6 句）各建立**獨立** stepdef 落點檔案（例如 `tests/e2e/steps/step_r033_t004..t014.go`）與一個 `tests/e2e/steps/wire_skills.go` helper 落點；每檔僅含 `gherkin` step 註冊骨架（可回傳 pending），**不含**邏輯。
  - 不做：不填 assertion（留 Phase 3）；不改既有 stepdef 檔案；不碰產品碼。

- [ ] T002 產品落點骨架
  - Read:
    - `specs/plans/033-skills-system/research.md` -> `Decision 5`, `Decision 7`
    - `internal/domain/tools/tools.go` -> `Tool` port（`Name`/`Description`/`Parameters`/`Contract`/`Execute`）
    - `internal/cli/cli.go` -> `agentTools()`（production assembler）與 `newToolRegistry` seam
    - `internal/infrastructure/tools/filesystem.go` -> `resourceSchema` builder + `readerDefaultTimeout`
  - 只做：建立落點骨架 — `internal/domain/skills/skill.go`（`Skill{Name, Description, Location}` 型別）、`internal/infrastructure/skills/loader.go`（loader 函式簽名 + TODO）、`internal/infrastructure/tools/skills.go`（`listSkills` tool 型別 + 五個 method 空殼，用 `resourceSchema`）、並在 `agentTools()` 預留 `list_skills` 註冊位置（空殼）。
  - 不做：不實作 loader 邏輯或 tool 行為（留 Phase 4）；不改 provider/loop/其他工具；不新增相依。

- [ ] T003 unit 測試落點骨架
  - Read:
    - `internal/infrastructure/tools/filesystem_test.go`（測試風格慣例）
    - `specs/plans/033-skills-system/research.md` -> `Decision 2`, `Decision 6`
  - 只做：建立 `internal/infrastructure/skills/loader_test.go` 與 `internal/infrastructure/tools/skills_test.go` 空殼（table-driven 骨架、`t.TempDir()` fixture 慣例），**不含** assertion。
  - 不做：不寫 assertion（留 Phase 3）；不動既有測試檔。

## Phase 3: Test Alignment & Implementation

**Goal**: 把本輪 Feature 用到的 DSL 自動化測試對齊最新版 truth — 含 truth-delta `ADD` 的 11 句（`[BDD-RED]`）、`MODIFY` 的 1 句（`[BDD-ALIGN]`），以及 loader/tool 的 `[UNIT]`。**不寫產品行為。**

**DSL 參照**:
- 每一句只屬於一個權威 `dsl.md`：同模組 `specs/truth/features/cli/chat/dsl.md`。介面根 `specs/truth/features/cli/dsl.md` 本輪 **NOOP**（無 cross-module row）。
- 讀法：用 task title 的句型對到該檔那一列，以 `StepDef 實作語意` 當作測試程式碼語意。Given / When 讀 `怎麼做`、`權威狀態落地`、`回寫`；Then 讀 `必查`（`呈現結果`、`權威狀態`、`再讀確認`）。
- `truth-delta.md` 只告訴這句是 ADD / MODIFY。語意以 `dsl.md` 那一列為準，不得用 feature 措辭或舊 stepdef 自行發明。

**Markers**:
- `[BDD-ALIGN]`：`MODIFY`。既有 stepdef 還在但語意是舊 truth。依 `dsl.md` 該列改測試，表達最新版 `StepDef 實作語意`。
- `[BDD-RED]`：`ADD`。依 `dsl.md` 該列寫出 stepdef。完成時這句可被跑到，失敗只能是 assertion 或產品行為，不能是 undefined step。
- `[UNIT]`：非 Gherkin 的單元驗證（loader / tool）。只動測試層。
- 三個 marker 都只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 033)` + `## Then (round 033)`
- `truth-delta.md` -> `/axb-dsl-refine`（ADD feature + 11 rows；MODIFY offered-set row）
- `tests/e2e/steps/` 既有 stepdef 檔（`reason` echo / tool-call recording 慣例）

**Boundary**:
- 一條 DSL 一個 task。
- 只改該句的 stepdef / assertion / 直接依賴的 helper（`wire_skills.go`）。
- 落點檔案採獨立檔案（Zero Shared Edits；一 task 一檔），消除並行寫入衝突。
- 不寫產品碼。
- review 啟動 subagent；本輪所有 Test Scope 不得再有 undefined step，失敗只能是 assertion 或產品行為。有 issues 就修正再 review，直到沒有任何問題。通過前不解鎖 Phase 4。

**Parallel Hint**:
- T004–T017 各派一個獨立 subagent（依 `ParallelHint平行Subagent與同檔調度判準.md` 調度）；T018 等全部回來再啟動 subagent 來 review。

- [ ] T004 [P] [BDD-ALIGN] `Then: the request offered exactly the agent tools`
  - Read: `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools` row（`集合` 現為七工具）
  - 只做：更新既有的 offered-set stepdef，使其斷言七個工具（新增 `list_skills`）。
- [ ] T005 [P] [BDD-RED] `Given: the workspace holds a skill "{name}" described as "{description}"`
- [ ] T006 [P] [BDD-RED] `Given: the workspace holds a skill "{name}" whose folder also holds a reference file "{relpath}"`
- [ ] T007 [P] [BDD-RED] `Given: the workspace holds no skills`
- [ ] T008 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to list its skills and then answers with "{answer}"`
- [ ] T009 [P] [BDD-RED] `Given: a configured provider "{provider}" whose endpoint asks tellme to list its skills, then to read the skill "{name}", and then answers with "{answer}"`
- [ ] T010 [P] [BDD-RED] `Then: tellme listed the skills using its list_skills tool`
- [ ] T011 [P] [BDD-RED] `Then: the listing includes the skill "{name}"`
- [ ] T012 [P] [BDD-RED] `Then: the listing does not include "{text}"`
- [ ] T013 [P] [BDD-RED] `Then: the listing reports that no skills are available`
- [ ] T014 [P] [BDD-RED] `Then: tellme read the skill "{name}" using its read_files tool`
- [ ] T015 [P] [BDD-RED] `Then: the request carried none of the text "{text}"`
- [ ] T016 [P] [UNIT] loader 測試（`internal/infrastructure/skills/loader_test.go`）
  - Read: `research.md` -> `Decision 2`；`specs/truth/techstack.md` -> Skills catalog (load) row
  - 只做：pin — frontmatter `name`/`description` 解析、遞迴發現、無 frontmatter 的 `.md` 不算 skill、重名首見者勝、缺/空目錄 → 空 catalog、CRLF 正規化、個別壞檔 best-effort skip。以 `t.TempDir()` 造 fixture。
- [ ] T017 [P] [UNIT] tool 測試（`internal/infrastructure/tools/skills_test.go`）
  - Read: `research.md` -> `Decision 6`, `Decision 7`；`specs/truth/features/cli/chat/dsl.md` -> round-033 Then rows
  - 只做：pin — `list_skills` 輸出格式（name + description + location）、paths 排序決定性、空 catalog 回報、schema well-formed（`required ⊆ properties`，由 round-031 gate 一併涵蓋）、未觀測 deadline 時回一般結果。
- [ ] T018 subagent review (phase quality gate)
  - Read: `tests/e2e/steps/step_r033_*.go`、`tests/e2e/steps/wire_skills.go`、`internal/infrastructure/skills/loader_test.go`、`internal/infrastructure/tools/skills_test.go`、`truth-delta.md`
  - 檢驗：11 句各有獨立 stepdef 且可被跑到（非 undefined）；offered-set ALIGN 已對齊七工具；loader/tool unit 測試就位；只動測試層。有 issues 修正再 review，直到零問題。通過前不解鎖 Phase 4。

## Phase 4A: ADD Feature File - cli/chat/listing-the-available-skills.feature

**Goal**: 交付 read-only `list_skills` tool 的產品實作，讓 Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/listing-the-available-skills.feature` -> `Feature: Listing the available skills`
- `specs/truth/features/cli/chat/dsl.md` -> `## Given (round 033)` + `## Then (round 033)`
- `truth-delta.md` -> `/axb-dsl-refine` ADD rows
- `specs/plans/033-skills-system/research.md` -> `Decision 1`–`Decision 7`

**Boundary**:
- 只在 prompt-bearing turn path 載入 catalog、構造並註冊 `list_skills`；**不注入** skill 內容；不改 provider/loop/其他工具；reuse `read_files` 供內容讀取；stdlib-only。

**Test Scope**:
- `specs/truth/features/cli/chat/listing-the-available-skills.feature`

- [ ] T019 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T020 [BDD-REFACTOR] 在綠燈下整理 loader 與 tool 的共用結構（frontmatter 解析、輸出格式），保持行為不變

## Phase 4B: MODIFY Feature File - cli/chat/offering-the-agent-tools.feature

**Goal**: 讓 offered tool set 更新為七工具，Test Scope 全綠。

**Shared Must Read**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature` -> `Feature: Offering the agent tools`
- `specs/truth/features/cli/chat/dsl.md` -> `the request offered exactly the agent tools` row
- `truth-delta.md` -> `/axb-dsl-refine` MODIFY row（offered set）

**Boundary**:
- `list_skills` 必須出現在 offered set（在 reader trio 之後、或既定順序），且不得改變其他工具的名稱/描述/schema 語意。

**Test Scope**:
- `specs/truth/features/cli/chat/offering-the-agent-tools.feature`

- [ ] T021 [BDD-GREEN] 讓 Test Scope 全綠
- [ ] T022 [BDD-REFACTOR] 在綠燈下整理 `agentTools()` 的註冊（無語意變更）

## Phase 5: Verification & Regression

**Goal**: 全量回歸 + 可偽性見證 + `make verify` + 拓樸稽核，並確認 class phrase／exit code／`stdout`/`stderr`／offline paths 皆未變。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Skills rows + Agent tool schemas row + Agent tool-schema gate row
- `specs/plans/033-skills-system/research.md` -> `Decision 4`, `Decision 9`

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses（可偽性）：(a) 移除 frontmatter 規則（把無 frontmatter 的 `.md` 也當 skill）→ non-skill-`.md` 的 Then 失敗；(b) 暫時把 skill 內容注入 request（system block）→ no-injection 的 Then 失敗。觀察到即還原，再跑一次確認綠燈。

- [ ] T023 [REGRESSION] 跑 Test Scope（`specs/truth/features/cli/**`）+ 見證 + `make verify` + 拓樸稽核
  - Read: `research.md` -> `Decision 4`, `Decision 9`；`specs/truth/techstack.md` -> Skills rows
  - 做：
    - `go test -count=1 ./...` 全綠（unit + godog E2E；本輪不需 undefined step）。
    - 見證 (a)/(b) 如上，觀察到即還原，重跑確認綠。
    - `make verify` **OK**；`gofmt -l .` clean；Gherkin/DSL 拓樸稽核 **PASSED**（44 features · 16 root + **299** module rows · **1533** steps）。
    - 確認 offline paths 不觸及 `docs/skills`；class phrase vocabulary 與 exit code **未變**；六個既有工具行為未變；`go.mod`／`go.sum` 不變（stdlib-only）。
  - 不做：不放寬任何 assertion；不為轉綠而移除見證。
- [ ] T024 subagent review (round quality gate)
  - Read: `internal/domain/skills/skill.go`、`internal/infrastructure/skills/loader.go`、`internal/infrastructure/tools/skills.go`、`internal/cli/cli.go`（`agentTools()` + prompt-path 載入）、`specs/truth/features/cli/chat/listing-the-available-skills.feature`、`specs/plans/033-skills-system/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：`list_skills` 為 read-only 且 no-injection；載入僅在 prompt path；六個既有工具與 provider/loop 未改；`techstack.md` 與 `truth-delta.md` 一致；無新相依；class phrase / exit code / `stdout`/`stderr` 未變；audit PASSED。

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Skills catalog (load) row | T001/T002（落點）、T016（unit）、T019（交付）、T023 | PASS |
| `specs/truth/techstack.md` -> Skills listing tool (`list_skills`) row | T002（骨架）、T017（unit）、T019（交付）、T024 | PASS |
| `specs/truth/techstack.md` -> Agent tool schemas row（builder 現支撐 `list_skills`） | T002、T017、T023 | PASS |
| `specs/truth/techstack.md` -> Agent tool-schema gate row（六 → 七工具） | T002、T017、T023、T024 | PASS |
| `specs/truth/techstack.md` -> Tool-usage accounting row（live-registry 七工具） | T023（offered/registry 未漂移）、T024 | PASS |
| `specs/truth/techstack.md` -> Not Introduced Yet（skills.sh + injection 非目標） | T024（確認未引入） | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY rows | T002、T016、T019、T023、T024 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` ADD feature + 11 rows | T001、T005–T015、T019–T020 | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` MODIFY offered-set row | T004、T021–T022 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP (`specs/truth/contracts/**`) | 豁免（NOOP 不建任務；T024 確認未觸及） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP (`specs/truth/data/**`) | 豁免（NOOP 不建任務；catalog 不持久化） | PASS |
| `research.md` -> Decision 1（單一來源 `<home>/docs/skills`） | T002、T005–T007、T019 | PASS |
| `research.md` -> Decision 2（frontmatter 遞迴發現；best-effort） | T002、T016、T019 | PASS |
| `research.md` -> Decision 3（on-demand；no injection） | T015、T019、T023（見證 b） | PASS |
| `research.md` -> Decision 4（載入僅 prompt path；offline 不變） | T019、T023 | PASS |
| `research.md` -> Decision 5（minimal domain type + loader；無 framework） | T002、T016 | PASS |
| `research.md` -> Decision 6（name+description+location；paths 排序） | T017、T019 | PASS |
| `research.md` -> Decision 7（`list_skills` 用 resourceSchema；30 s；agentTools） | T002、T017、T021 | PASS |
| `research.md` -> Decision 8（無 skills.sh / 無 injection） | T024（確認未引入；負向決策明示） | PASS |
| `research.md` -> Decision 9（hermetic 驗證；stdlib；POSIX） | T016、T017、T023（`make verify`、go.mod 不變） | PASS |
| `spec.md` -> US1（FR-001–FR-004, NFR-001） | T005–T007、T010–T013、T016、T019 | PASS |
| `spec.md` -> US2（FR-005–FR-006, NFR-002） | T009、T014、T015、T019、T023（見證 b） | PASS |
| `spec.md` -> 全域（FR-007–FR-009, NFR-003–NFR-006） | T019、T023、T024 | PASS |
| `spec.md` -> 邊界情況（nested dirs、malformed frontmatter、duplicates、absent dir、large catalog、offline） | T006、T007、T012、T016、T023 | PASS |
| `spec.md` -> SC-001–SC-005 | T019（SC-001/002）、T016（SC-003）、T023（SC-004/005、見證） | PASS |
| `plan.md` -> Structure（domain/skills、infrastructure/skills、tools/skills.go、cli、features/cli、techstack；no new dep） | T001–T003、T019–T023 | PASS |
| `plan.md` -> Scope notes（api/data NOOP；ui skipped） | T023（不觸及）、T024 | PASS |
| operator 拍板（Q1 → on-demand only；Q3 → `list_skills` tool） | T019、T023（no-injection 見證） | PASS |
| 交接來源（system-analysis handoff：CLI end → `/axb-dsl-refine`） | T001、T005–T015、T019（interface truth 已由 dsl-refine 產出） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
