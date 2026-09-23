# Feature Specification: Backward-counting turn indices in the `-l`/`--list` role headers (`[USER] - N` / `[MODEL] - N`) (round 082)

**Feature Branch**: `082-listing-backward-turn-indices`

**Created**: 2026-09-23

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the theme is anchored by issue [#165](https://github.com/gosharplite/tellme/issues/165) and the operator's directive *"Open a new round, the goal is to close #165"*, and the issue's *Design Principles / Implementation Touch Points* are routed to `/axb-technical-research`)

**Input (anchor issue [#165](https://github.com/gosharplite/tellme/issues/165), 2026-09-23; operator directive)**: *"Add backward-counting turn indices to `-l`/`--list` role headers (`[USER] - N` / `[MODEL] - N`) to align with `-b`/`--back [N]`"* — so the operator can read `-l`, see the exact suffix, and pass that same number to `tellme -b [N]` without counting turns by hand.

**Observed context (`dev` @ `f9aa96f`, round 081 closeout)**: `render.ListingMessage` carries `{Role, Body}` only — **no turn index**. `internal/cli/cli.go` `listingMessages(entries, n)` projects each `history.Entry` into two messages (`ListingOperator` = prompt, `ListingModel` = answer) and truncates to the last **n** **messages**; it never computes a turn distance. `internal/ui/listing.go` `header(role, colour)` emits a bare `[USER]` / `[MODEL]` (accented blue/magenta when colour is on). Round 073 (ADR 0045) added the headers and the stdout-terminal colour gate; round 081 (ADR 0053) added `-b`/`--back [N]`, which removes whole `history.Entry` lines.

**Behaviour intent**: **MODIFY (presentation of the offline `-l`/`--list` listing)** — append a **backward-counting turn index** to each role header, `[USER] - N` / `[MODEL] - N`, where `N = 1` is the **most recent** `history.Entry`. The index is a function of the operator's request and the session (**turn**-based, not message-based), carries **no** new storage, and leaves `-b` and `history.jsonl` **untouched**. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **Turn-based, not message-based (I-1).** `tellme -b [N]` removes whole `history.Entry` lines (prompt **and** answer together). Therefore the `[USER]` and the `[MODEL]` of one exchange **MUST** carry the **same** index `- K`; a header pair that disagreed would invite the operator to roll back a half-turn.
- **Backward counting (I-2).** Index `1` = the **most recent** turn in the active session history; index `K` = the K-th turn counting back from the end. There is **no** absolute / session-start numbering (the issue excludes `[USER] #3`).
- **True distance, not slice-relative (I-3).** `-l N` selects the last **N messages** — an **odd** N can lead with a `[MODEL]` whose partner `[USER]` is outside the slice. That leading message **MUST** keep its **true** turn distance from the end of history (e.g. a lone leading `[MODEL] - 2`), never be renumbered to `- 1` by its position in the printed slice.
- **Colour & raw discipline (I-4).** The whole label — `[USER] - N` / `[MODEL] - N` — is the colour unit: `blue("[USER] - N")` / `magenta("[MODEL] - N")` on a terminal `stdout` with `-r` off; under `-r`/`--raw` **or** a non-terminal `stdout`, the header prints **plain** text with **no** ANSI escapes. The existing stdout-terminal gate (round 073) is reused unchanged.
- **Presentation-only (I-5).** The change is confined to the `-l`/`--list` listing path: `history.jsonl` schema, `history.Store` contracts, `-b`/`--back`, the interactive prompt, and the live turn chrome (`╭─⠿ Turn N`) are **unchanged**. The exit-code set stays **ten**; no new phrase; the listing stays **offline** (no provider request, no stdin read).
- **Deterministic, hermetic, stdlib-only, POSIX-only (I-6).** No new dependency; the E2E arranges history and asserts streams without a pty or live network.
- **Backward-compatibility of the bytes (I-7).** The header text change is the **intended** contract change; the body rendering (model body via glamour, operator body verbatim), the blank-line separator, and the `-r` behaviour are **unchanged**.

---

## Grounded in the current system *(measured 2026-09-23, `dev` @ `f9aa96f`)*

| Site | Current shape |
| --- | --- |
| `internal/domain/render/ports.go` (`ListingMessage`) | `struct { Role ListingRole; Body string }` — **no** index/offset field. `ListingRole` = `ListingOperator` \| `ListingModel`. `Listing.Render(w, msgs, ListingSpec)`; `ListingSpec{Colour, Raw, Width, Warn}`. |
| `internal/ui/listing.go` (`header`) | `func (l listing) header(role render.ListingRole, colour bool) string` → `blue("[USER]", colour)` / `magenta("[MODEL]", colour)`. `Render` prints `header` then `body` then one blank line per message. |
| `internal/cli/cli.go` (`listingMessages`) | `func listingMessages(entries []history.Entry, n int) []render.ListingMessage` — emits (operator, model) per entry, then `if len(msgs) > n { msgs = msgs[len(msgs)-n:] }`. **No** index computed. |
| `internal/cli/cli.go` (`renderHistoryList`) | resolves the workspace, `store.Load()`, calls `newListing().Render(env.stdout, listingMessages(entries, n), ListingSpec{Colour: env.stdoutIsTerminal(), Raw: raw, Width: res.WrapWidth, Warn: env.stderr})`; returns `Success`. |
| `internal/cli/cli.go` (`-l` parsing) | `-l`/`--list` via `NoOptDefVal = "1"` + the shared optional-int pre-pass (round 081 generalised it to `-l` + `-b`); `-l 0`/`-l -1` is a usage error (exit 2). `-l` = last **N messages**. |
| `internal/ui/colour.go` | `blue(s, enabled)` / `magenta(s, enabled)` (the reference's `\033[1;34m` / `\033[1;35m`), empty-safe, plain-when-disabled. |
| `internal/domain/history/history.go` | `Entry{Prompt, Answer, Calls, Steps[]}`; `Load()` returns the complete-turn slice oldest→newest; `TotalCalls(entries)`. |
| `specs/truth/features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` | the `-l` listing Rules/Examples and their DSL rows — the owning truth of the header shape (round 073; ADR 0045). |
| `docs/decisions/0045-list-role-headers-and-rendered-body.md` | the round-073 decision (role headers, rendered model body, blank separator, stdout-gated accents) — this round **amends** it. |
| `docs/decisions/0053-back-rollback-turns.md` | the round-081 decision (`-b`/`--back [N]`; whole-turn removal) — the alignment target this round matches. |
| `tests/e2e/steps/` | the Given `the session history already holds the exchanges:` (a DataTable appended to `output/<mode>/history.jsonl`); the round-073 listing stepdefs. |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **L-1** | The header label gains a **trailing ` - N`**: `[USER] - N` / `[MODEL] - N`; both roles of one exchange share the **same** N. | **locked** (issue; operator directive) |
| **L-2** | `N` is a **backward** index: `1` = the most recent `history.Entry`; index = **true** turn distance from the end of the loaded history, **not** the position inside the printed slice. | **locked** (issue Design Principle 1–3) |
| **L-3** | The colour wrap encloses the **whole** label including the suffix; `-r`/non-terminal prints plain `[USER] - N` with no escapes. | **locked** (issue Design Principle 4) |
| **L-4** | **Presentation-only**: no change to `history.jsonl`, `history.Store`, `-b`/`--back`, the interactive prompt, or the live chrome; no new exit code or phrase. | **locked** (issue Invariant 5 / Out of Scope) |
| **S-1** | **Where the index is carried** — add a field to `render.ListingMessage` (e.g. `TurnIndex int`) computed in `listingMessages`, vs. compute entirely in the adapter from the slice length, vs. a richer `ListingMessage{..., TurnIndex}` populated by the CLI. | research decision (D-x); **proposed: the CLI computes the true distance into a new `ListingMessage` field** (keeps the adapter slice-agnostic and honest under odd slices) |
| **S-2** | **The exact label format string** — the separator (`" - "` vs `" – "`), and the fallback when an index is absent/zero (a non-history caller, or a future direct port use): plain `[USER]`? | research decision (D-x); **proposed: `fmt.Sprintf("[%s] - %d", …)`, with `TurnIndex <= 0` falling back to the bare `[USER]`/`[MODEL]`** (operator-vetoable) |
| **S-3** | **Which truth artifacts change** — the `history` CLI feature + `history/dsl.md` rows; a new **ADR**; the `techstack.md` *Session lifecycle flags* / listing row; whether `docs/domain-model/**` is modelled (the listing is a presentation surface — possibly **not modelled**, the ADR-0041 escape hatch). | research decision (D-x) |
| **S-4** | **Scope excludes** — absolute numbering (`[USER] #3`), a `--json` listing, a config toggle, `-l` count semantics (stays last-N **messages**), the `-b`-then-`-l` interplay, and any live/interactive surface. | **locked** (issue Out of Scope) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Turn-based.** Both messages of one `history.Entry` carry the **same** index; a turn is never halved in the display.
- **I-2 — Backward.** Index `1` = the most recent turn; index increments toward older turns.
- **I-3 — True distance.** The index is the distance from the end of the **loaded** history, unaffected by message-count truncation; a lone leading message keeps its true distance.
- **I-4 — Colour unit = whole label.** The accent wraps `[USER] - N` / `[MODEL] - N`; `-r`/non-terminal = no escapes.
- **I-5 — Presentation-only.** No storage/schema/`-b`/prompt/chrome change; exit-code set stays **ten**; no new phrase.
- **I-6 — Deterministic, hermetic, stdlib-only, POSIX-only.** No new dependency; E2E needs no pty or network.
- **I-7 — Unchanged neighbours.** The body rendering, the blank separator, the `-l` last-N-**messages** selection, and `-r` behaviour are unchanged.

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 在 `-l` 表頭看到可對應 `-b` 的回退索引 (Priority: P1)

作為一個操作者，當我用 `tellme -l [N]` 檢視最近的歷史時，我希望每個角色表頭附上一個由尾端倒數的回合索引（`[USER] - K` / `[MODEL] - K`，`1` = 最近一回合），使我能直接把看到的 `K` 交給 `tellme -b K`，不必再用肉眼從底部往上數。

**為何為此優先級**: 這是 issue 的核心與唯一交付面 —— 在既有的 `-l` listing 上補一個純呈現的索引，風險最小、價值直接（把「看」與「回退」對齊）。

**獨立驗證方式**: 以 Given `the session history already holds the exchanges:` 佈置 K 個回合，執行 `tellme -l M`，斷言輸出中每個表頭為 `[USER] - i` / `[MODEL] - i`，且同一回合的兩個角色索引相同、索引由 1 起往回遞增；再以奇數 M（造成前導半回合）驗證前導訊息保留其真實距離。

**驗收情境**:

1. **Given** a session whose active history holds 2 completed exchanges, **When** the operator lists the last 4 messages, **Then** the headers read `[USER] - 2`, `[MODEL] - 2`, `[USER] - 1`, `[MODEL] - 1` in that order.
2. **Given** the same 2-exchange session, **When** the operator lists the last 1 message, **Then** the single header is `[MODEL] - 1`.
3. **Given** the same 2-exchange session, **When** the operator lists the last 3 messages, **Then** the headers are `[MODEL] - 2`, `[USER] - 1`, `[MODEL] - 1` (the lone leading `[MODEL]` keeps true distance `2`).
4. **Given** a session whose last exchange used a tool, **When** the operator lists it, **Then** its `[USER]` and `[MODEL]` headers carry the same index (the turn is never split in the display).

**功能需求（FR）**:

- **FR-001**: 系統 MUST 在 `-l`/`--list` listing 的每個角色表頭附加由尾端倒數的回合索引，格式為 `[USER] - N` / `[MODEL] - N`（N 為正整數；確切分隔字串 = S-2）。
- **FR-002**: 同一 `history.Entry` 的 `[USER]` 與 `[MODEL]` MUST 帶**相同**的索引（I-1）。
- **FR-003**: 索引 MUST 為**倒數**：`1` = 最近一回合，往較舊方向遞增（I-2）。
- **FR-004**: 索引 MUST 為相對**完整已載入 history** 尾端的**真實距離**；message-count 截斷 MUST NOT 改變索引（奇數 `-l N` 的前導訊息保留其真實距離；I-3）。

**非功能需求（NFR）**:

- **NFR-001**: 索引計算 MUST 為決定性且與既有 `-l` 選取語意（最後 N **messages**）一致 —— MUST NOT 改變 `-l` 所選取的訊息集合（I-7）。

---

### 使用者情境 2 - 索引遵守色彩與 raw 紀律 (Priority: P2)

作為一個操作者，當我把 `-l` 輸出導向管線或在 `-r` 模式下使用時，我希望表頭是**純文字** `[USER] - N` / `[MODEL] - N`（無 ANSI 逃逸）；在終端且未加 `-r` 時，我希望整個標籤（含後綴）被上色。

**為何為此優先級**: 這是 US1 的表現層必要條件（既有的 stdout-terminal 色彩閘必須持續成立）；它在 US1 之上（US1 未成立前無從交付），故為 P2。

**獨立驗證方式**: 對同一 listing 分別以 colour 開/關（或 `-r` 開/關）執行，斷言：開啟時 `[USER] - N` 為 `\033[1;34m…\033[0m`、`[MODEL] - N` 為 `\033[1;35m…\033[0m`；關閉/`-r`/非終端時無任何 `\033`。

**驗收情境**:

1. **Given** a terminal `stdout` with `-r` off, **When** the operator lists the session, **Then** the `[USER] - N` label (including the suffix) is wrapped in the blue SGR pair and the `[MODEL] - N` label in the magenta pair.
2. **Given** `-r`/`--raw` **or** a non-terminal `stdout`, **When** the operator lists the session, **Then** the headers print as plain `[USER] - N` / `[MODEL] - N` with **no** ANSI escapes.

**功能需求（FR）**:

- **FR-005**: 當 stdout 為終端且 `-r` 關閉時，系統 MUST 以既有藍/洋紅 SGR 包住**整個**標籤（含後綴）。
- **FR-006**: 當 `-r`/`--raw` 開啟**或** stdout 非終端時，系統 MUST 印出純文字標籤且 MUST NOT 產生任何 ANSI 逃逸。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`verify-no-network` MUST NOT 被破壞；listing MUST 保持離線（無 provider 請求、不讀 stdin）。

---

### 邊界情況

- 當 session history **為空**（0 回合）時，系統 MUST 維持既有空 listing 行為（既有 `No history found.` 契約不變；MUST NOT 產生索引列）。
- 當 `-l N` 的前導訊息為**半個回合**（奇數 N）時，該前導訊息 MUST 保留其真實回合距離（FR-004）。
- 當 `history.jsonl` 無法解碼時，系統 MUST 沿用既有環境類片語（exit 4）回報（`emitHistoryError` 前例），MUST NOT 崩潰。
- 當 `-l` 與 `-r` 併用時，索引 MUST 以純文字呈現（FR-006）。
- 當索引欄位缺省或 `<= 0`（例如未來非 history 的呼叫端）時，系統 MUST 依 S-2 的決議處理（建議：回退為既有裸表頭 `[USER]`/`[MODEL]`，MUST NOT 產生 `- 0`）。
- 當 listing 被 `--new`／`-t`／其他離線指令伴隨時，系統 MUST 依既有 precedence 處理，且 MUST NOT 與索引呈現互相矛盾。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-007**: 索引 MUST 由 CLI 端的既有 `listingMessages` 投影計算並經 `render.ListingMessage` 傳遞（單一來源，MUST NOT 在配接器內以切片長度另算第二套；S-1）。
- **FR-008**: 本輪 MUST NOT 改變 `history.jsonl` schema、`history.Store` 契約、`-b`/`--back` 行為、互動提示與即時 turn chrome。

#### 非功能需求

- **NFR-003**: 結束碼集合 MUST 維持十個；片語詞彙 MUST 維持凍結；MUST NOT 因本輪新增任何片語或結束碼。
- **NFR-004**: listing MUST 保持離線（屬 `tellme performs no network access` 範圍），並 MUST NOT 讀取 stdin。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`render.ListingMessage`（`internal/domain/render`）**: 既有 `{Role, Body}`；本輪依 S-1 決議新增回合索引欄位（domain-owned shape，bytes 仍單一擁有於 `internal/ui`）。
- **`history.Entry`（`internal/domain/history`）**: 既有形狀 —— **schema 不變**；索引由其尾端距離導出，**不**新增欄位。
- **表頭標籤（listing header label）**: `[USER] - N` / `[MODEL] - N` 的呈現字串（確切格式 = S-2）；色彩為既有 stdout-terminal 閘（round 073）。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: hermetic E2E 中，對 K 個已佈置回合執行 `tellme -l 2K`（或 `-l 4` 於 2 回合）後，表頭序列為 `[USER] - K`, `[MODEL] - K`, … , `[USER] - 1`, `[MODEL] - 1`；同一回合兩角色索引相等。
- **SC-002**: `tellme -l 1` 於 2 回合 session 只印 `[MODEL] - 1`（最近一回合的答案）。
- **SC-003**: 奇數 `-l 3` 於 2 回合 session 印出 `[MODEL] - 2`, `[USER] - 1`, `[MODEL] - 1`（前導訊息保留真實距離）。
- **SC-004**: `-r`/非終端下輸出**不含**任何 ANSI 逃逸；終端且非 `-r` 下 `[USER] - N`（含後綴）以 `\033[1;34m…\033[0m`、`[MODEL] - N` 以 `\033[1;35m…\033[0m` 呈現。
- **SC-005**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），`TestExitCodesMatchPinnedContract` 綠，且 `-b`/`--back` 與 `history.jsonl` 契約位元組不變。

## 假設

- **A1**: 本輪 anchor = issue [#165](https://github.com/gosharplite/tellme/issues/165)；**DoD = 關閉它**。
- **A2**: 核心行為 —— 表頭格式 `[USER] - N` / `[MODEL] - N`、倒數（1 = 最近）、同回合同索引、真實距離、色彩包整個標籤、`-r`/非終端純文字、純呈現 —— 由 issue 鎖定；issue 的 *Proposed Implementation Touch Points*（S-1…S-4：索引乘載位置、確切格式/缺省、truth/ADR/技術棧更新、範圍排除）由 `/axb-technical-research` 決議，故本輪 clarify **未升級（0 題）**。研究決議若改變正式驗收契約（例如採「配接器由切片長度推算」），MUST 回寫 truth 並於 `truth-delta.md` 記錄。
- **A3**: `history.jsonl` 於本輪起點僅含**完整**回合（rounds 079/080 保證），故尾端距離與回合邊界一致。
- **A4**: listing 的既有 stdout-terminal 色彩閘（round 073 D3）與 `-r` 行為可直接重用，無須新增閘。
- **A5**: 預期新增一枚 ADR、`history` CLI feature 的 Rules/Examples + `history/dsl.md` rows 更新，以及 `techstack.md` 的 CLI/listing 列更新；`docs/domain-model/**` 是否更新由研究決議（listing 為呈現面，可能 not modelled，ADR 0041 escape hatch）；`contracts/**` 與 `data/**` 預期 NOOP。
- **A6**: 本輪**不**實作绝对編號、`--json` listing、config 開關，也不改變 `-l` 的 last-N-**messages** 選取語意（S-4）。
