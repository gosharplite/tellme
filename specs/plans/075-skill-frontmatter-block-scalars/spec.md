# Feature Specification: Skill frontmatter block-scalar descriptions (round 075)

**Feature Branch**: `075-skill-frontmatter-block-scalars`

**Created**: 2026-09-21

**Status**: Draft (specified; clarify **not escalated — 0 questions**; residual technical choices → `/axb-technical-research`; see §Clarify strategy)

**Anchor**: **operator request** — no anchor issue.

**Input (operator, 2026-09-21)**:

> *"Now about fixing tellme …"* — after observing that six in-tree skills render as `- <name>: > (<path>)` in `list_skills` because their `description:` uses a YAML **folded block scalar** (`>`), the operator directed that **tellme itself** be fixed to read block-scalar frontmatter (the docs-side fix to the three `tm-*` template skills was done separately, in the seed repo). The three `domain-model-*` skills are **out of scope** (not owned by us).

**Behaviour intent**: **MODIFY (skill loading)** — the skills loader resolves a frontmatter scalar written as a YAML **block scalar** (`>` / `|`, with chomping indicators) to its textual value, instead of the literal indicator character. **No other behaviour change.** **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The defect (grounded).** `internal/infrastructure/skills/loader.go` → `parseFrontmatter` is a **line-based** reader: for `description: >` it takes only the text after `:` on *that same line* → the value becomes the literal string `">"`. `renderSkills` (`internal/infrastructure/tools/skills.go`) then prints `- <name>: > (<location>)`. The frontmatter is **valid YAML**; the reader is simply not YAML-aware.
- **The reference has the *same* limitation.** `tell-me-go` `internal/infrastructure/skills/file_repo.go` → `parseSkill` is also line-based (`strings.HasPrefix(line, "description:")` → `strings.TrimPrefix`), so it would also yield `">"`. **This round is therefore a deliberate divergence *beyond* the reference — a robustness improvement, not parity.** It MUST be recorded as a divergence in the truth row and the ADR (not claimed as reference behaviour).
- **Where the truth lives.** `specs/truth/techstack.md` → §Skills → **Skills catalog (load)** row (owned this round by `/axb-technical-research`) pins: *"a Markdown file is a skill iff its content starts with a `---` frontmatter block declaring `name` and `description` (the reference's `parseSkill` rule)"*. The invariants to preserve: **best-effort** (never fails a turn), **first-wins** on a duplicate name, **non-skill Markdown skipped**, **empty catalog** on a missing/empty dir.
- **The listing contract is unchanged.** `list_skills` returns each skill's **name + description + location**, path-sorted, with an empty catalog reported as a *result* (not an error) — `specs/truth/features/cli/chat/listing-the-available-skills.feature`. This round changes the **parsed description value**, not the tool's surface, ordering, or bounds.
- **The value type is unchanged.** `internal/domain/skills.Skill` (`Name`, `Description`, `Location`) is untouched.

---

## Grounded in the current system *(measured 2026-09-21, `dev` @ `3dfc3a4`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/skills/loader.go` — `parseFrontmatter` | normalises CRLF; requires a leading `---\n`; splits the block on lines; `key, val, _ := strings.Cut(line, ":")`; `description = unquote(strings.TrimSpace(val))` — i.e. **the same line only**, so `description: >` ⇒ `">"`; a file needs non-empty `name` **and** `description` to be a skill. |
| `internal/infrastructure/tools/skills.go` — `renderSkills` | `Skills (<n>):` header, then one `- <name>: <description> (<location>)` line per skill; empty ⇒ `No skills are available.` |
| `specs/truth/techstack.md` (row 54) | *"starts with a `---` frontmatter block declaring `name` and `description` (the reference's `parseSkill` rule)"* |
| `specs/truth/features/cli/chat/dsl.md` (row 389) | the E2E Given writes **inline** frontmatter `---\nname: {name}\ndescription: {description}\n---\n` |
| Observed evidence | a `SKILL.md` with `description: >` + indented lines lists as `- <name>: > (<path>)` |

---

## Design (shape is a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | tellme reads a block-scalar `description` as its **textual value**; the listing must **never** show a bare `>`/`|`. | **locked (operator)** |
| **S-2** | **How** to implement: a **minimal stdlib** block-scalar extension to the line reader, versus adopting a real YAML parser (`gopkg.in/yaml.v3`) — a **dependency / techstack** decision. | **research decision** (D-x) |
| **S-3** | The **exact folding / chomping semantics** and the **malformed-block** outcome (an indicator with no indented body). | **research decision** (D-x) |
| **S-4** | Whether the E2E gains a **new/extended Given** that can author a block-scalar skill (the current Given writes inline frontmatter only). | **research decision** (D-x); the interface truth is `/axb-dsl-refine`'s to write. |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — No regression.** Inline (quoted and unquoted) `name`/`description` values resolve unchanged (byte-identical listing entries); a non-skill Markdown file is still skipped; a duplicate name is still first-wins; a missing/empty/unreadable directory still yields an empty catalog.
- **I-2 — Best-effort only.** The loader never fails a turn; a file whose frontmatter cannot be resolved to a valid skill is skipped silently (the run continues).
- **I-3 — The listing contract is unchanged** — name + description + location, path-sorted; an empty catalog is a *result*, not an error.
- **I-4 — stdlib-only, POSIX-only, hermetic** *(proposed)* — prefer no new dependency (matching the round-033 "minimal, stdlib-only" loader); adopting a YAML library is the S-2 alternative and carries its own justification in `techstack.md`.
- **I-5 — Single source.** The catalog stays single-source (`<TELL_ME_HOME>/docs/skills/`); no `.skills/`, no `skillssh` tools.

---

## Clarify strategy

**Not escalated (0 questions).** No gap changes story slicing, requirement attribution, the main flow, or the acceptance/success criteria:

- the **goal** (read block scalars) and the **scope** (the frontmatter values the loader reads) are unambiguous from the operator's request;
- the residual choices — *stdlib vs a YAML library* (**S-2**), *exact folding/chomping + malformed-block semantics* (**S-3**), *the E2E authoring seam* (**S-4**) — are **technical**, and per the round-074 precedent defer to `/axb-technical-research`;
- the reference-parity finding (the reference shares the limitation) is a **disclosure**, resolved by the operator's explicit intent to fix tellme regardless.

**Disclosed assumptions (not asked):**

- **A1** — **stdlib-only** is preferred over a new dependency; the YAML-library option is available to research but must justify itself.
- **A2** — the fix applies to **every frontmatter scalar the loader reads** (`name` and `description`), for a uniform reader — though the motivating case is `description`.
- **A3** — folding / chomping semantics follow **YAML 1.2** (`>` folds intra-block line breaks to single spaces; a blank line becomes a newline; `-` strips the trailing newline; `|` is literal).
- **A4** — the `Skill` value type, the `list_skills` tool, its bounds, and its ordering are **unchanged**.
- **A5** — the six in-tree `>` skills are the **motivating evidence** (three `tm-*` already fixed docs-side; three `domain-model-*` are **out of scope**), but the fix is **general** — it must handle any block-scalar frontmatter, including future / third-party (`skills.sh`-style) skills.

**No `NEEDS CLARIFICATION` remains.**

---

## 使用者情境與測試 *(必填)*

### 使用者故事 1 - A folded block-scalar description is listed as its real text (Priority: P1)

作為一個 `$TELL_ME_HOME` 的技能由人手或其他 agent 撰寫的 tellme 操作者，我希望一個以 YAML folded block scalar（`description: >`）撰寫的技能的 `description` 能被 `list_skills` 以其真正摘要文字列出，讓 agent 能單憑列表判斷相關性，而不必逐一開啟每個檔案。

**為何為此優先級**: 這正是本次缺陷本身 —— 沒有它，任何以 `>` 撰寫的技能在目錄裡都只顯示 `>`，發現性等於零；其餘變體（`|`、chomping）都是同一修正的延伸。

**獨立驗證方式**: 在 runtime home 放置一個以 `description: >` + 縮排內文撰寫的 `SKILL.md`，執行 tellme 並觸發 `list_skills`；結果必須顯示折疊後的真實文字，且不得出現裸 `>`。

**驗收情境**:

1. **Given** the runtime home holds a skill whose `description` is a folded block scalar (`>` + indented lines), **When** the agent lists skills, **Then** the listing carries that skill's folded description text and **not** the bare indicator `>`.
2. **Given** the same folded block whose source spans several lines, **When** the agent lists skills, **Then** each intra-block line break is folded to a single space (the entry has no stray line break).

**功能需求（FR）**:

- **FR-001**: 系統 MUST 將 `description:` 的值為 folded block scalar（`>`，含 `>-`）者解析為其折疊後文字。
- **FR-002**: 系統 MUST 將 block 內部的換行折疊為單一空白，並去除 block 結尾的換行（YAML folding）。
- **FR-003**: 系統 MUST 讓 `list_skills` 的項目顯示解析後的真實 description，**不得**顯示裸的 `>` / `|` 指示字元。

### 使用者故事 2 - Literal and chomping indicators are honoured (Priority: P2)

作為一個可能以 `|`（literal）或 chomping 變體（`|-` / `>-`）撰寫技能的 tellme 操作者，我希望這些變體也被正確解析，讓任何指示字元都不會滲入列表。

**為何為此優先級**: 這是同一讀取器的完整覆蓋；它建立在 US1 的解析器之上，價值較低但成本也低，屬收尾性的穩健化。

**獨立驗證方式**: 分別以 `|`、`|-`、`>-` 與一個含空行的 `>` 段落撰寫技能，檢查解析出的 description 值符合 YAML 語意。

**驗收情境**:

1. **Given** a literal block scalar (`|`), **When** the agent lists skills, **Then** the resolved description preserves the source's line breaks.
2. **Given** a chomping indicator (`>-` or `|-`), **When** the agent lists skills, **Then** the block's trailing newline is stripped.

**功能需求（FR）**:

- **FR-004**: 系統 MUST 依 YAML 語意解析 literal（`|`）與 chomping（`>-`、`|-`、`>+`、`|+` 若適用）變體。*(fold F-2：最終 `TrimSpace` 會正規化結尾換行，故對 frontmatter 純量而言各 chomping 變體等價；解析器接受其拼法，但不宣稱可由測試分辨。)*
- **FR-005**: 系統 MUST 支援多段落 folded block（block 內的單一空行轉為一個換行）。

### 邊界情況

- 當一個 inline 值本身含 `>`（例如 `description: a > b`）時，系統 MUST 將其視為一般值並原樣解析（**不得**當作 block scalar）。
- 當某個 block-scalar 指示字元後沒有任何縮排內文時，系統 MUST 解析出空值，因而該檔案**不**構成技能（維持 best-effort skip）。
- 當 `name` 亦以 block scalar 撰寫時，系統 MUST 以相同規則解析。
- 當檔案使用 CRLF 換行時，系統 MUST 正規化後再解析（維持既有行為）。
- 當值以單引號或雙引號包覆時，系統 MUST 維持既有去引號行為，不得退化。
- 當兩個技能同名時，系統 MUST 維持 first-wins（維持既有行為）。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 系統 MUST 讓 frontmatter 讀取器對未知/非純量行維持 line-based、best-effort 忽略（例如 `disable-model-invocation: true` 或其他額外鍵），使更豐富的 frontmatter 永不破壞載入。

#### 非功能需求

- **NFR-001**: 讀取器 MUST 維持 stdlib-only、POSIX-only、hermetic；若採用外部 YAML 函式庫，MUST 在 `techstack.md` 明列並附理由（預設不新增依賴）。
- **NFR-002**: 讀取器 MUST 維持 best-effort —— 任何 malformed / 無法解析的 frontmatter 只會使該檔案被跳過，**永不**使一個 turn 失敗。
- **NFR-003**: 解析 MUST 具確定性，且載入目錄的順序與 `list_skills` 的 path-sorted 排序維持不變。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`Skill`**: 載入的技能值型別（`Name`、`Description`、`Location`）—— **形狀不變**；本輪只改變 `Description` 的**解析來源**（支援 block scalar），不新增欄位或實體。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 對一個同時含 block-scalar（`>`、`|` 及 chomping 變體）與 inline description 的技能目錄，`list_skills` 的每個條目 **100%** 顯示其預期 description 文字，**0** 個條目顯示裸的 `>`/`|`。
- **SC-002**: 既有 inline / 引號 description 的列表條目維持 **byte-identical**（零回歸）。
- **SC-003**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴）。

## 假設

- **A1**: 以 stdlib 實作為優先；採用外部 YAML 函式庫為 S-2 的備選，需於 `techstack.md` 附理由。
- **A2**: 修正適用於讀取器所讀取的每個 frontmatter 純量（`name`、`description`）。
- **A3**: folding / chomping 依 YAML 1.2 語意。
- **A4**: `Skill` 值型別、`list_skills` 工具、其 bounds 與排序皆不變。
- **A5**: 六個 in-tree `>` 技能為動機證據，但修正為通則（含未來 / 第三方 `skills.sh` 風格技能）。
