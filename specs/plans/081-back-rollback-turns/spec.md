# Feature Specification: Roll back the last N turns of the session history (`-b`/`--back`) (round 081)

**Feature Branch**: `081-back-rollback-turns`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the theme is anchored by issue [#163](https://github.com/gosharplite/tellme/issues/163) and the operator's directive *"Open a new round, the goal is to close #163"*, and the issue's *Open design decisions* are routed to `/axb-technical-research`)

**Input (anchor issue [#163](https://github.com/gosharplite/tellme/issues/163), 2026-09-22; operator directive)**: *"Add `-b`/`--back`: roll back the last N turns of the session history (offline, optional count, optional follow-up prompt)"* — port the reference's `-b`/`--back [N]` (*"Go back / delete the last N turns from history"*), which `tellme` today rejects (`tellme -b` → `tellme: the command-line usage is invalid`, exit 2).

**Observed context (`dev` @ `0008fd0`, round 080 closeout)**: `internal/cli/cli.go` `parseFlags` registers `-c -d -h -i -l --new -r --tool-usage -t -v` — there is **no `-b`**. The offline session commands (`-l`, `-t`, prompt-less `--new`) already resolve the session **workspace** through `resolveWorkspace` + `offlineConfigAndMode` (mode = `TELL_ME_MODE` → the `-c` config's `MODE` → the default config's `MODE` → `butler`). The domain port `` internal/domain/history.Store `` exposes only `Load()`, `Append(Entry)`, `Archive()` — **no truncate/rewrite capability**. The adapter `internal/infrastructure/history/file_store.go` stores `output/<mode>/history.jsonl` (+ `history.archive.jsonl`), one `history.Entry` line per completed turn; `Archive()` moves the **whole** active file (that is `--new`, not a rollback).

**Behaviour intent**: **ADD (a CLI offline session command + a durable store capability)** — `tellme -b [N]` rolls back the **last N turns** (default **1**) of the session history and exits offline; `tellme -b [N] "prompt"` rolls back first and then runs `"prompt"` against the trimmed history. A **turn** is one `history.Entry` line — `{prompt, answer, calls, steps[]}` — so removing a turn removes the operator prompt, the model answer, **and every tool step recorded in that turn**, atomically. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **A turn is exactly one `history.Entry` line (I-1).** `tellme` stores one line per turn with its tool steps embedded in `steps[]` (round 008); rolling back drops the **last N lines**. No message-pair arithmetic, and no risk of splitting a `FunctionCall`/`FunctionResponse` pair — the line is atomic (the reference's message-pair math and its odd-parity repair are **not** needed here).
- **Only complete turns exist in the file (I-2, load-bearing).** Rounds 079/080 (ADR 0051/0052) guarantee `history.jsonl` holds only **complete** turns — a partial/failed turn is closed with a synthetic answer *before* it is written. So after a rollback the file still replays as a valid `user … assistant` sequence on **both** provider families (Gemini/Vertex included).
- **Durable & atomic (I-3).** The rewrite MUST be write-to-temp + `fsync` + atomic rename over `history.jsonl`; a crash mid-rollback MUST leave the **prior** history intact (never truncate the live file in place before the new bytes are durable).
- **Rollback ≠ `--new` (I-3b).** A rollback MUST **not** write `history.archive.jsonl` — it drops the last N turns and keeps the rest; `--new` archives everything and starts fresh. The two must not be conflated.
- **Standalone `-b` is offline (I-5).** `tellme -b` / `tellme -b N` make **no** provider request; they join the offline-path set (the `tellme performs no network access` truth row). The prompt-bearing form (`-b "p"`) does contact the provider and shows the normal turn chrome.
- **Frozen surfaces (I-6).** No new `tellme: ` class phrase; the exit-code set stays **ten**; the standalone confirmation is written to **`stdout`** (the offline-report channel, like `-l`/`--version`); a genuine failure (e.g. an unreadable/undecodable history) reuses the existing environment-class phrase + exit **4** (the round-007 `emitHistoryError` precedent).
- **Session selection is identical to `-l`/`-t` (I-7).** mode = `TELL_ME_MODE` → the `-c` config's `MODE` → the default config's `MODE` → `butler`; an explicit `-c` that cannot be honoured **refuses** (round-053 Q2 → A).
- **The next turn's number follows the trimmed history (I-8).** The `╭─⠿ Turn N` header derives from `history.TotalCalls(prior) + 1`; after a rollback it MUST reflect the **reduced** history.
- **Hermetic + stdlib-only + POSIX-only (NFR-2).** No new dependency; `verify-no-network` intact; the E2E needs no pty and no live network.

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `0008fd0`)*

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go` (`parseFlags`) | registers `-c -d -h -i -l --new -r --tool-usage -t -v`; **no `-b`**. `-l` uses `pflag` `NoOptDefVal = "1"` + a bespoke `consumeListValue(args)` pre-pass that consumes an **adjacent integer** (`-l 5` → `-l=5`), leaves a non-integer token untouched, stops at `--`. |
| `internal/cli/cli.go` (`run`, dispatch) | precedence: `--help` → `--version` → `-d` → `-l` → `-t` → `--tool-usage` → prompt → prompt-less `--new` → boot. |
| `internal/cli/cli.go` (`resolveWorkspace`, `offlineConfigAndMode`) | resolve only home + effective mode + workspace for `-l`/`-t`/prompt-less `--new`; the mode = `TELL_ME_MODE` → `-c` MODE → default MODE → `butler`; an explicit unreadable `-c` refuses. |
| `internal/cli/cli.go` (`renderNewSession`, `renderTurnsLog`, `emitHistoryError`) | the offline actions; `emitHistoryError` maps a history read/write failure to `the runtime home is not usable (session history: …)` + exit **4**. |
| `internal/domain/history/history.go` (`Store`) | `Load() ([]Entry, error)` · `Append(Entry) error` · `Archive() error`; `Entry{Prompt, Answer, Calls, Steps[]}`; `TotalCalls(entries) int`. |
| `internal/infrastructure/history/file_store.go` | active `history.jsonl` (append-only, `O_APPEND\|O_CREATE`, `Sync`); archive `history.archive.jsonl`; `Archive()` moves the whole active file then removes it. **No trim.** |
| `internal/cli/exitcode.go` | `Success=0`, `UsageError=2`, `ConfigError=3`, `EnvironmentError=4`, … `ProviderError=6`, `ToolError=7`; the ten-value set is pinned by `TestExitCodesMatchPinnedContract`. |
| `internal/cli/cli.go` (`turnNumber`) | `history.TotalCalls(prior) + 1` — the turn-header index. |
| `docs/decisions/0023-…md` (`RF-54-4`) | *"`-l`'s optional value is implemented via a bespoke pre-pass; if more optional-int flags appear (`-b`/`--back`-style), the pre-pass should be generalised."* — **this round is its trigger.** |
| `tests/e2e/steps/` | the Given `the session history already holds the exchanges:` (a DataTable of `prompt`/`answer` rows appended to `output/<mode>/history.jsonl`); a shared `readSessionEntries` helper (round 080). |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **L-1** | The flag surface: `-b`/`--back [N]`, default **1** when the count is omitted; a following **non-integer** token is a **prompt** (not a count), so `tellme -b "p"` = roll back 1 then run `"p"`. | **locked** (issue + operator: *"tellme -b \"new prompt\" will work? → yes"*) |
| **L-2** | Standalone `-b` / `-b N` is an **offline** action (no provider request) that reports and exits; `-b [N] "p"` rolls back **then** runs the prompt. | **locked** |
| **L-3** | A rollback removes whole `history.Entry` lines (prompt + answer + all `steps[]`) atomically; it MUST NOT write the archive. | **locked** (issue N-1/N-3/`--new` contrast) |
| **S-1** | **The Store capability** — (a) add `Rollback(n int) (removed int, err error)` to the domain port; (b) a generic `Rewrite([]Entry) error`; (c) `Load` + trim + rewrite at the CLI layer. | research decision (D-x) |
| **S-2** | **Cardinality when `N` exceeds the available turns** — **clamp** to all turns and report the actual removed count (reference behaviour), or **refuse** with an error? | research decision (D-x); **proposed: clamp** (operator-vetoable) |
| **S-3** | **`N ≤ 0` boundary** — `-b 0` / `-b -1` as a **usage error** (exit 2, symmetric with `-l ≤ 0`), or a no-op? | research decision (D-x); **proposed: usage error** (operator-vetoable; the reference treats `backN ≤ 0` as a silent no-op) |
| **S-4** | **Composition / precedence** — does `-b` compose with `-l` (reference: **list then roll back**) and with `--new` (roll back then fresh session, or refuse)? Exact position of `-b` in the dispatch order. | research decision (D-x) |
| **S-5** | **The confirmation line** — exact text + stream (`stdout`, plain, **no** `tellme: ` prefix, not in `turns.log`); pluralisation. | research decision (D-x) |
| **S-6** | **Side logs** — whether a rollback also trims `turns.log` / `tokens.log`. | research decision (D-x); **proposed: no** (out of scope) |
| **S-7** | **The optional-int pre-pass** — generalise `consumeListValue` into a shared pre-pass covering `-l`/`--list` **and** `-b`/`--back` (honours **RF-54-4**; reconcile the recorded boundary cases: positional-agnostic; combined short flags like `-rl 5` are not handled). | research decision (D-x) |
| **S-8** | **Records** — a new **ADR** + index; `techstack.md` (the CLI-flags row + the offline-path set); the `history` CLI feature + `history/dsl.md` rows; the `docs/domain-model` `History` entity (a rollback action/invariant; same-PR, ADR 0041); `data/**` likely NOOP (schema unchanged). | research decision (D-x) |
| **S-9** | **Scope excludes** — `--retry` (roll back the last user message + resend), `-e`/`--update-turn`, the interactive `browse`, `--back` in the TUI prompt, an interactive confirmation, and any `--json`. | **locked** |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — One line per turn, atomically removed.** A rollback removes whole `history.Entry` lines; prompt, answer, and all `steps[]` go together; a turn is never split.
- **I-2 — Only complete turns exist.** Post-079/080 the active file holds only complete turns, so after `-b` it still replays as a valid `user … assistant` sequence on **both** families.
- **I-3 — Durable & atomic; archive untouched.** Temp-file + `fsync` + rename; a crash mid-rollback leaves the prior history intact; `history.archive.jsonl` is never written by a rollback.
- **I-4 — Schema unchanged.** `history.Entry` / `history.Step` JSON shapes are untouched; no id/timestamp is introduced (bytes stay deterministic, NFR-002 of the store).
- **I-5 — Offline when standalone.** `tellme -b` / `tellme -b N` make **no** provider request; the prompt-bearing form does (with the normal turn chrome + spinner).
- **I-6 — Frozen surfaces.** No new class phrase; the exit-code set stays **ten**; the standalone report is `stdout`; a genuine failure reuses the environment-class phrase + exit 4.
- **I-7 — Session selection identical to `-l`/`-t`** (`TELL_ME_MODE` → `-c` MODE → default → `butler`); `-b -c missing.yaml` refuses.
- **I-8 — Turn numbering follows the trimmed history** (`TotalCalls(remaining) + 1`).
- **I-9 — Deterministic, hermetic, stdlib-only, POSIX-only.** No new dependency; the E2E arranges history and asserts files/streams without a pty or live network.

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 離線回退最後 N 個回合 (Priority: P1)

作為一個操作者，當我對最近幾個回合不滿意時，我希望用 `tellme -b`（或 `tellme -b N`）**離線**地把最後 N 個回合（預設 1）從 session history 中移除並得到回報，使我能乾淨地「撤銷」而不必重建整個 session；過程中 tellme **不得**發出任何 provider 請求。

**為何為此優先級**: 這是 issue 的核心 —— 補上 `tellme` 目前完全缺失的「撤銷」動作；它是 prompt-bearing 形式的先決條件，且風險面最小（純離線、字串比對回退為行截斷）。

**獨立驗證方式**: 以 Given `the session history already holds the exchanges:` 佈置 K 個回合，執行 `tellme -b N`，斷言 `output/<mode>/history.jsonl` 恰好剩 `K − N` 行、`stdout` 出現回報列、結束碼 0、且所有 fake provider 記錄 **零** 次請求；再以 `tellme -l` 讀回驗證被移除的是**最後** N 個回合。

**驗收情境**:

1. **Given** a session whose active history holds 3 completed exchanges, **When** the operator asks tellme to roll back the last 1 turn, **Then** `history.jsonl` holds exactly 2 exchanges (the last one removed), `stdout` carries a rollback confirmation, and the run exits **0** and sends no request to any provider.
2. **Given** the same 3-exchange session, **When** the operator asks tellme to roll back the last 2 turns, **Then** `history.jsonl` holds exactly 1 exchange and the run exits **0** offline.
3. **Given** a session whose last exchange used a tool, **When** the operator rolls it back, **Then** the removed turn's prompt, answer, **and** its tool `steps` are gone together (the remaining file is unchanged for the earlier turns).

**功能需求（FR）**:

- **FR-001**: 當操作者執行 `tellme -b`（未給數值）時，系統 MUST 移除 session history 的**最後 1 個** `history.Entry`。
- **FR-002**: 當操作者執行 `tellme -b N`（N 為正整數）時，系統 MUST 移除 session history 的**最後 N 個** `history.Entry`（維持其餘既有順序）。
- **FR-003**: 獨立形式的 `-b` MUST 為**離線**動作 —— MUST NOT 對任何 provider 發出請求。
- **FR-004**: 回退 MUST NOT 寫入 `history.archive.jsonl`（回退 ≠ `--new`）。
- **FR-005**: 系統 MUST 在 `stdout` 印出一行回報，說明實際移除的回合數（其餘 stream/schema 見 NFR）。

**非功能需求（NFR）**:

- **NFR-001**: 回退 MUST 為耐久的（durable）與原子的（atomic）：先寫入暫存檔並 `fsync`，再以原子替換覆蓋 `history.jsonl`；中途失敗 MUST 保留原有 history 不變。

---

### 使用者情境 2 - 回退後立即送出新提示 (Priority: P2)

作為一個操作者，當我想「撤銷上一回合並改問」時，我希望 `tellme -b "new prompt"`（或 `tellme -b N "new prompt"`）先回退最後 N 個回合，**再**以修剪後的 history 送出新的提示，讓一次指令完成「undo 並重問」。

**為何為此優先級**: 這是 `-b` 的高價值用法（reference 的 `bb` 別名精神），但建立在 US1 之上（US1 未成立前無法交付），故為 P2。

**獨立驗證方式**: 以 Given 佈置 K 個回合，執行 `tellme -b 1 "…"`（由 in-process fake provider 腳本化一個純文字回答），斷言：新回合以**修剪後**的 history 為前文、`history.jsonl` 最終恰好 K 行（移除 1、新增 1）、`stdout` 印出新的答案、結束碼 0。

**驗收情境**:

1. **Given** a session whose active history holds 2 completed exchanges, **When** the operator runs `tellme -b "try again"`, **Then** the last exchange is removed first and the new prompt `"try again"` is then answered against the trimmed history (the resulting file holds 1 old + 1 new exchange).
2. **Given** a session of 3 exchanges, **When** the operator runs `tellme -b 2 "fresh"`, **Then** the last 2 exchanges are removed and `"fresh"` runs against the 1 remaining exchange.

**功能需求（FR）**:

- **FR-006**: 當 `-b`（或 `-b N`）**伴隨**一個非整數的位置參數（提示）時，系統 MUST 先完成回退，**再**以修剪後的 history 執行該提示的一個正常推理回合。
- **FR-007**: prompt-bearing 形式 MUST 顯示正常的回合 chrome（此形式**允許** provider 請求）。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`verify-no-network` MUST NOT 被破壞；E2E MUST NOT 需要 pty 或真實網路。

---

### 邊界情況

- 當 `N` **超過**現有回合數時，系統 MUST 依 S-2 的決議處理（建議：clamp 至全部並回報實際移除數）—— MUST NOT 損毀 history。
- 當 `N ≤ 0`（`-b 0` / `-b -1`）時，系統 MUST 依 S-3 的決議處理（建議：usage error + exit 2）。
- 當 session history **為空**（0 回合）時，系統 MUST 依 S-2/S-3 精神安全處理（建議：exit 0 且回報移除 0，不報錯）。
- 當 `-b` 與 `-l` 併用時，系統 MUST 依 S-4 的決議順序處理（reference：先 `-l` 列出、後 `-b` 回退）。
- 當 `-b` 與 `--new` 併用時，系統 MUST 依 S-4 的決議處理（MUST NOT 造成 archive 與 rollback 的語意衝突）。
- 當 `-b -c <missing>` 指向不可讀取的配置時，系統 MUST 拒絕（沿用既有 config 類片語 + exit 3，round-053 Q2 → A）。
- 當 `history.jsonl` 存在但**無法解碼**時，系統 MUST 沿用以**環境類**片語（exit 4）回報（`emitHistoryError` 前例），MUST NOT 崩潰、MUST NOT 留下半截檔案。
- 當既有 history 檔**不存在**時，系統 MUST 視為空 session（不報錯）並依空 session 語意處理。
- 當 `-b` 與 `-t`／`--tool-usage`／`--version`／`-h` 同時出現時，系統 MUST 依既有 precedence 與 S-4 的決議處理，且 MUST NOT 產生相互矛盾的行為。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-008**: 回退 MUST 透過**單一**、具耐久語意的 store capability 實作（S-1），MUST NOT 在多處重複截斷邏輯。
- **FR-009**: 系統 MUST 依 **RF-54-4** 將 `-l` 的 adjacent-integer 前置處理**一般化**為同時涵蓋 `-l`/`--list` 與 `-b`/`--back`（S-7）。
- **FR-010**: 回退 MUST NOT 改變 `history.Entry` / `history.Step` 的既有 JSON schema，且 MUST NOT 觸碰被移除回合以外的行（包含 `signature` 欄位）。

#### 非功能需求

- **NFR-003**: 結束碼集合 MUST 維持十個；片語詞彙 MUST 維持凍結；MUST NOT 因本輪新增任何片語或結束碼。
- **NFR-004**: 獨立 `-b` MUST 屬於離線路徑集（併入 `tellme performs no network access` 的範圍），並 MUST NOT 讀取 stdin。
- **NFR-005**: 回合編號 MUST 依修剪後的 history 推導（`TotalCalls(remaining) + 1`；I-8）。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`history.Entry`（`internal/domain/history`）**: 既有形狀 —— **schema 不變**；一個回合 = 一行；回退即移除整行。
- **Store 回退能力（新增）**: 依 S-1 決議的耐久截斷/重寫介面（domain port + 檔案配接器實作）。
- **回報列（rollback confirmation）**: 印到 `stdout` 的單行純文字（確切文字 = S-5），**不得**是 `tellme: ` 類別片語。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: hermetic E2E 中，對 K 個已佈置回合執行 `tellme -b N` 後，`history.jsonl` 恰有 `K − N` 行、`stdout` 有一行回報、結束碼 0，且**所有** fake provider 記錄零請求。
- **SC-002**: `tellme -b` 等同 `tellme -b 1`（預設值），且被移除的是**最後**一個回合（可由 `-l` 讀回驗證）。
- **SC-003**: `tellme -b 1 "p"` 先回退再執行：最終 `history.jsonl` = `K − 1 + 1` 行，且新回合以前文（修剪後 history）執行（由 fake provider 記錄的請求內容佐證）。
- **SC-004**: 回退**不寫入** archive（`history.archive.jsonl` 不變）。
- **SC-005**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），`TestExitCodesMatchPinnedContract` 綠，且既有 `-l`/`--new` 契約位元組不變。

## 假設

- **A1**: 本輪 anchor = issue [#163](https://github.com/gosharplite/tellme/issues/163)；**DoD = 關閉它**。
- **A2**: 核心行為 —— flag 形式（`-b [N]`，預設 1）、獨立離線、`-b [N] "prompt"` 先回退再執行、一次移除整回合、不寫 archive —— 由 issue 與操作者指令鎖定；issue 的 *Open design decisions*（S-1…S-8：Store capability、clamp/refuse、`N ≤ 0`、composition、回報文字、附屬 log、pre-pass 一般化、records）由 `/axb-technical-research` 決議，故本輪 clarify **未升級（0 題）**。研究決議若改變正式驗收契約（例如 clamp vs refuse），MUST 回寫 truth 並於 `truth-delta.md` 記錄。
- **A3**: `history.jsonl` 於本輪起點僅含**完整**回合（rounds 079/080 保證），故回退為對 `user … assistant` 邊界的乾淨截斷。
- **A4**: 回退可安全重用 `resolveWorkspace`/`offlineConfigAndMode` 的 session 選擇；`-b` 的離線形式屬 `--version`/`-l` 類的離線路徑。
- **A5**: 預期新增一枚 ADR、`techstack.md` 的 CLI 旗標列 + 離線路徑列更新、`history` CLI feature 的 Rules/Examples + `history/dsl.md` rows，以及 `docs/domain-model` 的 `History` 實體更新（same-PR，ADR 0041）；`data/**` 預期 NOOP（schema 不變）。
- **A6**: 本輪**不**實作 `--retry`；`-e`/`--update-turn`、`browse`、TUI 內 `--back`、互動確認與 `--json` 皆不在範圍（S-9）。
