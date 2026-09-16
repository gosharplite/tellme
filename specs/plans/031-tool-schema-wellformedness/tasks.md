# Tasks: 031 — tool-schema well-formedness (`required ⊆ properties` + registry gate)

**Plan Package**: `specs/plans/031-tool-schema-wellformedness`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`
> `specs/truth/contracts/**` (`/axb-api-plan` **NOOP**), `specs/truth/data/**` (`/axb-data-plan` **NOOP**), and `specs/truth/features/**` (`/axb-dsl-refine` **NOOP**) are unchanged this round. There is no `ui/**` (the tool schemas are model-facing).

## Task Binding Contract

- 每個**開發任務**都必須對應 `truth-delta.md` 中的 ADD / MODIFY / DELETE / NOOP 語意，或 `research.md` 已拍板的 Decision。
- **本輪是「schema 正確性 + 驗證 gate」輪，非 BDD feature 輪**：`/axb-api-plan`、`/axb-data-plan`、`/axb-dsl-refine` 皆為 **NOOP**（無 API／無資料／無 CLI interface truth 變更）。因此：
  - **省略 Phase 1 `Setup`** —— 本輪不新增技術（`research.md` D5；`go.mod`／`go.sum` 不動）。
  - **Phase 3 `Test Alignment` 只有 `[UNIT]`** —— 本輪沒有新的 DSL 句，也沒有 stepdef；驗證由 **production-assembler（`agentTools()`）unit gate** 承擔（非 Gherkin，`research.md` D2）。
  - **沒有 Feature phase** —— 沒有 `specs/truth/features/**` 變更；產品修復以 **Phase 4 Implementation** 直接交付。
  - 交付物是 **shared schema builder 的修復**（`internal/infrastructure/tools/filesystem.go`）與 **production-assembler（`agentTools()`）well-formedness gate**（`internal/cli/tool_registry_test.go`）；唯一 truth 變更是 `specs/truth/techstack.md`（by `/axb-technical-research`）。
- 每個開發任務的 `Read` 須涵蓋 `research.md` 對應 Decision 與 `specs/truth/techstack.md` 對應 section。
- Truth 參照使用 `specs/truth/**` 路徑；plan 參照使用當前 plan package 相對路徑。
- **不動** the offered tool set、the `tellme` binary 的 `stdout`／`stderr` 行為、the class-phrase vocabulary 或 exit code、任何工具的名稱/描述/執行語意。

## Round-031 locked decisions (implementation constraints — MUST)

> 來自 `research.md` Decisions 1–6 與 spec US1/US2（FR-001..FR-009、SC-001..SC-004）。

- **[FIX-SITE]** 只改 **shared schema builder**（`resourceSchema`）：在 `properties` 中宣告 **`reason`** property（每個 caller 都已把 `reason` 放進 `required`，且 `reason` 對每個工具都是 mandatory）。修好它所支撐的五個工具（`list_files`、`read_files`、`get_tree`、`write_file`、`replace_text`）。**`execute_command` 不動**（inline schema 已宣告 `reason`，是唯一合規者）。
- **[GATE]** 在 **非可覆寫的 production assembler `agentTools()`** 上斷言（**不**用可被測試覆寫的 `newToolRegistry` var——ARCH-1）：對 **每一個** 註冊工具，schema 的 `required` 每個名稱都必須在 `properties` 中宣告，且 schema 可解析為 JSON object。**覆蓋全部工具**（含 write pair 與 `execute_command`）。root-only 走訪（flat-schema 前置條件）；巢狀物件自帶 `required` 的遞迴形式列為 **`#60`** forward item（ARCH-2）。
- **[RED-FIRST]** gate 必須先對現行缺陷 **非真空地失敗**（今日 5 個工具違規），再於 `[GREEN]` 修復後轉綠。
- **[NO-DEP]** 只用 Go 標準庫（`encoding/json`、`testing`）；無新相依；`go.mod`／`go.sum` 不動；POSIX-only。
- **[VERIFY]** 驗證為 **hermetic**（離線）；真實 Vertex/Gemini 確認是 **closeout 手動步驟**（非 gate 成員）。
- **[TRUTH]** 唯一 truth 變更 = `specs/truth/techstack.md`（Agent tool schemas row + Agent tool-schema gate row）。

---

## Phase 1: Setup

**Omitted** —— 本輪無新增技術（`research.md` D5；無新相依，`go.mod`／`go.sum` 不動）。不得把落點骨架塞進 Setup。

## Phase 2: Foundational

**Omitted** —— 無新 DSL 句／stepdef 落點骨架；修復落點即 `resourceSchema`，gate 落點即既有 `internal/cli/tool_registry_test.go`，皆由下方任務直接交付。

## Phase 3: Test Alignment (unit)

**Goal**: 先把本輪的驗證落到測試層（**不寫產品碼**）：新增 **production-assembler（`agentTools()`）層** well-formedness gate 並讓它對現行缺陷 **RED**，同時強化只斷言 `required` 的盲點測試。

**Markers**:
- `[UNIT]`：本輪的驗證是 **unit tier**（production-assembler 層 schema 檢查）；非 Gherkin、非 `[BDD-*]`（本輪無 DSL 句）。只動測試層，不寫產品碼。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Agent tool-schema gate row（Testing & Verification）
- `truth-delta.md` -> `/axb-technical-research` MODIFY `techstack.md`
- `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 2`（gate 位置與非真空要求）

**Boundary**:
- 只動測試層（`internal/cli/tool_registry_test.go`、`internal/infrastructure/tools/filesystem_test.go`）；不寫產品碼。
- review 啟動 subagent；本輪 gate 必須對現行缺陷 **非真空失敗**（非 undefined、非 parse error）；通過前不解鎖 Phase 4。

- [X] T001 [UNIT] 新增 production-assembler well-formedness gate（RED）
  - Read:
    - `specs/truth/techstack.md` -> Agent tool-schema gate row
    - `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 2`
    - `internal/cli/cli.go` -> `agentTools()`（production assembler；gate 對象）與 `newToolRegistry` var（DI seam；**不得**作為 gate 對象）
    - `internal/cli/tool_registry_test.go` -> 既有 `TestNewToolRegistryOffersAgentTools`（同一套工具集）
    - `internal/domain/tools/tools.go` -> `Tool.Parameters()` 契約
  - 做：新增 `TestAgentToolSchemasAreWellFormed`（名稱可調），迭代 **非可覆寫的 production assembler `agentTools()`**（ARCH-1：gate 不得坐在可被測試覆寫的 `newToolRegistry` var 上，否則未來一個忘記 `t.Cleanup` 還原的覆寫會讓 gate 讀到 fake registry 而**空過**）；對每個工具解析 `Parameters()` 為 `{properties map[string]json.RawMessage, required []string}`，斷言：schema 可解析、為 JSON object、且 **每個 `required` 名稱都在 `properties` 中**（`required ⊆ properties`）。
  - 邊界（ARCH-2，flat-schema 前置條件）：本 gate 只走 **root** 的 `properties`/`required`；六個 schema 今日皆為 flat（`read_files` 只在 `items` 內層帶 **properties**、內層無 `required`），故此檢查正確——以註解 **明示此前置條件**。巢狀物件自帶 `required` 的**遞迴**形式列為 **`#60` forward item**，不得靜默假設。
  - 邊界（nits）：failure message 需含 **不變式名稱 `required ⊆ properties`** 與違規工具名（diagnostic 指向**類別**，不只工具名）；另 **明確斷言** spec 的兩個 edge case——**zero-`required`** 工具須 **vacuously pass**，**非 object／不可解析** 的 schema 須 **fail**（FR-007）。
  - 邊界（optional belt-and-braces，PR #65 fold-review 建議）：另以 **一行**斷言 `agentTools()` 的**工具名集合**與 `newToolRegistry().Tools()` 的工具名集合**相同**，把 assembler 與 registry 綁在一起，使兩處測試不會漂移；此比較只讀 registry 的**名稱**（**不**作為 gate 對象——不變式仍只對 `agentTools()` 斷言，ARCH-1）。
  - 邊界：此測試 **今日必須失敗**（5 個工具違規）——**不得**放寬 assertion 讓它變綠（RED-first）；不寫產品碼。
  - 不做：不改 `resourceSchema`（留 Phase 4）；不碰 offered-set 測試。

- [X] T002 [UNIT] 強化盲點測試 `TestToolSchemasRequireReason`（`reason`-specific；**不**重抄全工具清單）
  - Read:
    - `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 2`
    - `internal/infrastructure/tools/filesystem_test.go` -> 既有 `TestToolSchemasRequireReason`（只斷言 `required` **包含** `reason`，且只涵蓋 3 個 reader）
  - 做：把該測試強化為斷言 **`reason` 是 `properties` 中宣告的 property**（而不只是出現在 `required`），並涵蓋 **builder-backed** 的工具（`list_files`、`read_files`、`get_tree`、`write_file`、`replace_text`——即 `resourceSchema` 所支撐者；`execute_command` 為 inline 且已宣告 `reason`），使「required 有、property 沒有」不再能通過。
  - 邊界（ARCH-3）：**完整性的責任在 T001**（production-assembler 的 `required ⊆ properties`，覆蓋**每一個**工具）。**T002 不得**再手抄第二份「全部六工具」清單——該 package（`internal/infrastructure/tools`）**無法 import `internal/cli`**，兩份清單會漂移。T002 只保留 `reason` 的專項敘述與 builder-backed 範圍（用**單一本地 table**，或直接對 builder 支撐的工具斷言），不與 production-assembler gate 競爭完整性。
  - 不做：不重寫其他既有 reader 測試；不建立第二份全工具清單。

- [X] T003 subagent review (phase quality gate)
  - Read: `internal/cli/tool_registry_test.go`、`internal/infrastructure/tools/filesystem_test.go`、`research.md` -> `Decision 2`
  - 檢驗：gate 覆蓋 **每一個** 註冊工具（非子集）；今日 **非真空失敗** 且點名違規工具；無 undefined／parse error；只動測試層。有 issues 修正再 review，直到零問題。

> Phase-3 review executed by the orchestrator (no parallel-subagent substrate in this session): `go test ./internal/cli/ ./internal/infrastructure/tools/ -run 'WellFormed|RequireReason'` → `TestAgentToolSchemasAreWellFormed` **FAILED non-vacuously**, naming exactly the 5 violators (`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`); `TestSchemaWellFormedEdgeCases` PASSED; `TestToolSchemasRequireReason` FAILED for the same 5. Failures are assertion-only (no undefined/parse errors). Gate PASSED — Phase 4 unlocked.

## Phase 4: Implementation (product fix)

**Goal**: 以最小產品變更（shared schema builder 宣告 `reason` property）讓 T001 的 gate 轉綠，且不改任何工具的行為或 offered set。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Agent tool schemas row（CLI Application）
- `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 1`, `Decision 3`
- `internal/infrastructure/tools/filesystem.go` -> `resourceSchema` 與其五個 caller

**Boundary**:
- 只改 `resourceSchema`：在其輸出的 `properties` 中加入 `reason` property（description 可先內嵌或單一常數化）。
- **不動** `execute_command`（`internal/infrastructure/tools/command.go`）：其 inline schema 已宣告 `reason`，保留其 `output_file`/`append` props 與 process-tree `timeout` 措辭（`research.md` D3）。
- 不改任何工具的名稱、描述、執行語意、offered set、flag／exit／`stdout`／`stderr` 契約。
- 不得為了轉綠而放寬 T001／T002 的 assertion。

- [X] T004 [GREEN] 修復 shared schema builder：宣告 `reason` property
  - Read:
    - `specs/truth/techstack.md` -> Agent tool schemas row
    - `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 1`, `Decision 3`
    - `internal/infrastructure/tools/filesystem.go` -> `resourceSchema` + 五個 caller
  - 做：讓 `resourceSchema` 在 `properties` 中宣告 `reason`（型別 `string`、簡短 description），使其支撐的五個工具全部滿足 `required ⊆ properties`。確認 `execute_command` 仍合規且未改。
  - 驗證：T001／T002 轉綠；`go test ./...`（含 E2E）維持綠；確認無任一既有測試 pin 了「舊的、缺 `reason` 的」schema（若有，於 T006 回報；預期無——既有測試只斷言 `required` 包含 `reason`）。
  - 不做：不改工具行為／offered set；不碰 `command.go`；不新增相依。

- [X] T005 [REFACTOR] 在綠燈下整理 schema builder 的 `reason` 描述來源
  - Read: `internal/infrastructure/tools/filesystem.go`、`research.md` -> `Decision 1`
  - 做：若可提升可讀性，將 `reason` 的 description 抽成單一常數（與 `maxOutputTokensDesc` 同層風格）；保持輸出語意不變，gate 續綠。
  - 不做：不擴大重構範圍、不動其他 row 的措辭、不改行為。

## Phase 5: Verification & Regression

**Goal**: 全量回歸 + 可偽性見證 + `make verify` + 拓樸稽核，並確認 offered set／class phrase／exit code／`stdout`/`stderr` 皆未變。

**Shared Must Read**:
- `specs/truth/techstack.md` -> Agent tool schemas row + Agent tool-schema gate row
- `specs/plans/031-tool-schema-wellformedness/research.md` -> `Decision 4`, `Decision 6`

**Boundary**:
- 不改產品碼；只跑回歸與見證。
- Witnesses（可偽性）：(a) **還原** builder 修復 → T001 gate 必須失敗（非真空）；(b) 暫時新增一個 **required 有、property 沒有** 的工具（或暫時移除某工具的 `reason` property）→ gate **非零失敗並點名**；觀察到即還原，再跑一次確認綠燈。

- [X] T006 [REGRESSION] 跑全測試 + 見證 + `make verify` + 拓樸稽核
  - Read: `research.md` -> `Decision 4`, `Decision 6`；`specs/truth/techstack.md` -> 兩 row
  - 做：
    - `go test -count=1 ./...` 全綠（unit + godog E2E）。
    - 見證 (a)/(b) 如上，觀察到即還原，重跑確認綠。
    - `make verify` **OK**；`gofmt -l .` clean；Gherkin/DSL 拓樸稽核 **PASSED**（本輪無 feature 變更，步數應與 round 030 相同：42 features · 16 root + 273 module rows · 1403 steps）。
    - `go.mod`／`go.sum` 不變（stdlib-only）。
    - 記錄 **SC-002 的手動 Vertex/Gemini 確認** 為 closeout 步驟（非 gate）：以真實 Vertex/Gemini provider 跑一個 plain prompt 與一個 tool-using prompt，確認無 `400 required fields ... not defined`。
  - 不做：不放寬任何 assertion；不為轉綠而移除見證。

- [X] T007 subagent review (round quality gate)
  - Read: `internal/infrastructure/tools/filesystem.go`、`internal/cli/tool_registry_test.go`、`specs/truth/techstack.md`、`specs/plans/031-tool-schema-wellformedness/{spec.md,research.md,plan.md,truth-delta.md}`
  - 檢驗：gate 覆蓋 **每一個** 註冊工具且非真空；修復僅動 shared builder（`command.go` 未改）；offered set（六工具）/ class phrase / exit code / `stdout`/`stderr` 皆未變；`techstack.md` 與 `truth-delta.md` 一致；無新相依；`execute_command` 仍合規。

> Round review executed by the orchestrator: the fix touches only `internal/infrastructure/tools/filesystem.go` (+ the two test files); `command.go` is **unchanged** (still inline-compliant); `go.mod`/`go.sum` unchanged (stdlib-only); the offered set is still the six tools; `make verify` **OK**; topology audit **PASSED** (42 features · 16 root + 273 module rows · 1403 steps — unchanged); falsifiability witnesses reproduced then reverted — (a) reverting the builder fix → the gate fails non-vacuously, (b) a mandatory-but-undeclared argument → the gate fails and names the tool.

---

## Pre-Delivery Orphan Coverage Sweep

| Truth Spec / Plan Decision | Covered By Task Read / Delivery | Status |
|:---|:---|:---:|
| `specs/truth/techstack.md` -> Agent tool schemas row | T004（交付修復）、T001/T002（gate）、T007 | PASS |
| `specs/truth/techstack.md` -> Agent tool-schema gate row | T001（交付 gate）、T003（phase review）、T006（見證）、T007 | PASS |
| `truth-delta.md` -> `/axb-technical-research` MODIFY `techstack.md` | T001、T004、T007 | PASS |
| `truth-delta.md` -> `/axb-api-plan` NOOP (`specs/truth/contracts/**`) | 豁免（NOOP 不建任務；T006 驗證不觸及 API surface） | PASS |
| `truth-delta.md` -> `/axb-data-plan` NOOP (`specs/truth/data/**`) | 豁免（NOOP 不建任務；無資料變更） | PASS |
| `truth-delta.md` -> `/axb-dsl-refine` NOOP (`specs/truth/features/cli/**`) | 豁免（NOOP 不建任務；無 CLI interface 變更；T007 確認 offered-set rows 未變） | PASS |
| `research.md` -> Decision 1（fix 於 shared builder） | T004、T005、T007 | PASS |
| `research.md` -> Decision 2（gate 於 production assembler `agentTools()`，覆蓋全工具；ARCH-1/2/3） | T001、T002、T003、T006 | PASS |
| `research.md` -> Decision 3（`execute_command` 不動） | T004、T007 | PASS |
| `research.md` -> Decision 4（hermetic gate + 手動 live 確認） | T006（hermetic + 手動 closeout 記錄） | PASS |
| `research.md` -> Decision 5（no new dependency；stdlib；POSIX） | T004/T006（`go.mod` 不變） | PASS |
| `research.md` -> Decision 6（deferred：fake 驗 schema、live pipeline leg、結構化 builder） | 負向決策；T004/T007 邊界明示不引入 | PASS |
| `spec.md` -> US1（FR-001–FR-004, NFR-001） | T004、T001 | PASS |
| `spec.md` -> US2（FR-005–FR-007, NFR-002） | T001、T002、T003、T006 | PASS |
| `spec.md` -> 全域（FR-008, FR-009, NFR-003） | T004（不改 offered set/phrase）、T006（`make verify`；live 為手動） | PASS |
| `spec.md` -> 邊界情況（zero-required vacuous；非 JSON object 失敗；新工具自動覆蓋） | T001（gate 邏輯）、T007 | PASS |
| `spec.md` -> SC-001–SC-004 | T001（SC-001/003）、T006（SC-003/004）、T007（SC-004）；SC-002 手動 closeout | PASS |
| `plan.md` -> Structure（filesystem.go CHANGED；tool_registry_test.go ADDED；command.go 未動；techstack.md MODIFY；no new dep） | T001、T004、T006 | PASS |
| `plan.md` -> Scope notes（api/data/dsl-refine NOOP；`/axb-ui-plan`、`/axb-spec-by-example` skipped） | T006（不觸及 API/資料/CLI）、T007 | PASS |
| operator 拍板（Q1 → 1：hermetic gate + 手動 live；Q2 → 1：minimal fix + gate） | T001（hermetic gate）、T004（minimal fix）、T006（手動 live 記錄） | PASS |

> 孤立產物件數：0。掃描通過，准予交付。
