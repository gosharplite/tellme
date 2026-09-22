# Feature Specification: Persist a failed turn's completed tool steps (round 080)

**Feature Branch**: `080-failed-turn-partial-persistence`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the theme is anchored by issue [#161](https://github.com/gosharplite/tellme/issues/161) and the operator's directive, and the issue's *Open Design Decisions* are routed to `/axb-technical-research`)

**Input (anchor issue [#161](https://github.com/gosharplite/tellme/issues/161), 2026-09-22; operator directive)**: *"retry succeeds → all executed steps persist; retry fails → none do. We need to fix the above. When retry fails and exits during long tool calls, previous tool calls should be persisted similar to Ctrl+C."*

**Observed context (`dev` @ `b1c929e`)**: round 079 (ADR 0051) made an **operator interruption** (`SIGINT`/`SIGTERM` → `context.Canceled`) preserve a partial turn (one `history.Entry` closed with a synthetic answer). But a **genuine turn failure** still discards the work: in `internal/cli/cli.go` (`runTurn`), when `loop.Run` returns an error the only persistence branch is `errors.Is(err, context.Canceled) && len(result.Steps) > 0`; any other error maps to `emitProviderError` (phrase + **exit 6**) or `emitToolError` (**exit 7**) and **skips the append**. The completed steps exist in memory (`agentport.Result{Steps, Calls}` is returned **alongside** the error) but are discarded.

**Behaviour intent**: **MODIFY (session-history persistence)** — a failed turn that completed **at least one** tool step is **persisted** (the same shape round 079 uses: one `history.Entry` closed with a synthetic assistant answer), while the **failure surface is unchanged** (the frozen class phrase + the existing exit code, plus one informational `stderr` line). A turn that fails with **zero** completed steps keeps today's clean abort. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The failure surface does NOT change (I-2).** A failed-but-kept turn still exits **6** (provider) / **7** (tool loop) with its frozen class phrase. Exit **0** would misreport a genuine failure (round 079's exit 0 is right *only* because the operator chose to stop). No new class phrase; the exit-code set stays **ten**.
- **Role alternation is the whole reason for the synthetic answer (I-1).** `history.jsonl` is replayed verbatim by `BuildMessages`: a partial turn ending on a `tool` result would produce two consecutive `user` roles — rejected by Gemini/Vertex (its `functionResponse` parts ride under `role:"user"`) and ill-formed on the OpenAI-compatible wire. A persisted partial turn MUST be **closed with an assistant message**.
- **The interruption predicate is the round-079 seam, broadened (I-5 / S-2).** `runTurn` holds `ctx`; a `SIGINT`/`SIGTERM` shows as `ctx.Err() != nil` (the `defer cancel()` runs only after this branch). Using `ctx.Err() != nil` instead of `errors.Is(err, context.Canceled)` is **strictly broader**: it also catches a `Ctrl+C` that landed **during a round-078 retry wait**, where the retry decorator returns the previous attempt's provider error (`lastErr`, a transport `*llm.ProviderError`) that does **not** wrap the cancellation (issue **hole #2**).
- **The trigger is `len(result.Steps) > 0` (I-3).** Same as round 079: a step is recorded whenever a tool **returned** a result (success, or a nil-error kill/timeout marker) — not necessarily a successful one. Zero steps ⇒ **nothing** is written, for every error class.
- **The persisted entry schema is unchanged (I-4).** One append-only line of the existing `{prompt, answer, calls, steps}` shape; `calls` = the **completed** inference rounds. The synthetic answer is the `answer` value.
- **A completed-tool-step turn is bounded by the tool loop (edge case).** On the exit-7 path (`ErrIncomplete`: the loop bound reached / the unknown-tool cap exhausted / no tools registered) the same rule applies — if steps were completed, they are kept; the run still reports the tool failure + exit 7.
- **Hermeticity holds (NFR-2).** The failure is produced by the in-process fake provider (an always-drop transport → retry exhaustion; and an outright rejection → non-retryable), no live network, no pty.

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `b1c929e`)*

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go` (`runTurn`, the single persistence seam) | `result, err := loop.Run(ctx, prompt, prior)`; `ind.Stop()`; `if err != nil { if errors.Is(err, context.Canceled) && len(result.Steps) > 0 { return persistInterruptedTurn(...) }; ErrIncomplete → emitToolError; else emitProviderError }`; `store.Append(...)` only when `err == nil`. |
| `internal/cli/cli.go` (`persistInterruptedTurn`) | round 079: builds `history.Entry{Prompt, Answer: history.InterruptedTurnAnswer, Calls: len(result.Calls), Steps: result.Steps}`, appends, emits the informational `stderr` line, returns `Success`. |
| `internal/agent/agentloop.go` (`Run`) | on a `Complete` error returns `agentport.Result{Steps: steps, Calls: calls}, err` — the completed steps survive in the returned value; on the bound/cap it returns `Result{Steps, Calls}, &ErrIncomplete{…}`. |
| `internal/cli/retry_gateway.go` (`retryingGateway.Complete`) | on its two abort paths returns `lastErr` (the previous attempt's provider error), **not** the cancellation → the round-079 `errors.Is(err, context.Canceled)` test misses a `Ctrl+C` during the retry wait (**hole #2**). |
| `internal/domain/history/history.go` | `Entry{Prompt, Answer, Calls, Steps}`; `InterruptedTurnAnswer` (round 079). |
| `internal/cli/exitcode.go` | `ProviderError = 6`, `ToolError = 7`; the ten-value set pinned by `TestExitCodesMatchPinnedContract`. |
| `docs/domain-model/tellme.modelith.yaml` | `history-append-after-complete` + `turn-answer-stored-verbatim` amended in round 079 for the operator-interrupted exception. |
| `tests/e2e/fakeprovider` | `DropFirst(n)` (transport drop; a large `n` = always), `ErrorStatus(400)` (non-retryable), `Script(...)` (tool-call scripts), `RequestCount`. |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **Scope** — (A) provider failures only, or (B) **any** failed turn with `len(steps) > 0` (also the tool-loop **bound-reached / unknown-tool-cap**, exit 7). | research decision (D-x); **proposed (B)** — one predicate covers both; the loss is identical. Operator-vetoable. |
| **S-2** | **The predicate** — replace round 079's `errors.Is(err, context.Canceled)` with the strictly broader `ctx.Err() != nil` for the interruption branch (folds hole #2). | research decision (D-x); **proposed: yes**. |
| **S-3** | **Exit code / diagnostic** — keep the real failure surface (phrase + exit 6/7) and add one informational `stderr` line; no new code. | research decision (D-x); **proposed: keep 6/7**. |
| **S-4** | **The synthetic answer for the failure case** — a distinct, cause-naming string vs reusing the round-079 `[Turn interrupted by operator via Ctrl+C]`. | research decision (D-x); **proposed: distinct** (the `answer` is the only marker). |
| **S-5** | **Usage** — persist the completed calls' usage on the failed-but-kept path, or keep round 079's D6 (none)? | research decision (D-x); **proposed: none** for both partial paths (one decision). |
| **S-6** | **The informational line** — exact text; `stderr` only, control-free, **no** `tellme: ` prefix, not in `turns.log`. | research decision (D-x). |
| **S-7** | **Records** — a new **ADR 0052** + index; `techstack.md` (the *Interrupted-turn persistence* row → a broader *partial-turn persistence* row); the CLI feature + `dsl.md` rows; the `docs/domain-model` invariant extension (same-PR, ADR 0041). | research decision (D-x); truth written by the owner skills. |
| **S-8** | **Scope excludes** — MCP steps are ordinary steps on the same seam; `-b`/`--retry` and the offline readers untouched; the in-flight child `SIGKILL` untouched. | **locked** |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Role alternation on resume.** A persisted partial turn MUST replay as a valid `user … assistant` sequence on **both** provider families (the synthetic close is load-bearing).
- **I-2 — The failure surface is unchanged.** A failed turn still exits **6** (provider) / **7** (tool loop) with its frozen class phrase. No new phrase; the exit-code set stays **ten**.
- **I-3 — Zero steps ⇒ nothing written**, for every error class.
- **I-4 — One append-only entry; schema unchanged.** `history.Entry` / `history.Step` JSON shapes untouched; `calls` = the completed inference rounds.
- **I-5 — Typed/structural detection, never string matching.** The interruption classification is `ctx.Err()` (structural); the failure classification is the existing typed error path.
- **I-6 — Surface-neutral.** `stdout` byte-unchanged (a failed turn writes no answer); the informational line is `stderr`-only and control-free.
- **I-7 — Hermetic + stdlib-only + POSIX-only.** No new dependency; `verify-no-network` intact; the E2E needs no pty and no live network.
- **I-8 — The in-flight child kill is untouched** (`execute_command` children still `SIGKILL`-ed via the process group).

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 失敗的回合保留已完成的工具步驟 (Priority: P1)

作為一個執行長時間工具回合的操作者，當回合因為**提供者失敗**（重試耗盡，或不可重試的 4xx）而中止時，我希望 tellme **保留**中斷前已完成的工具步驟（以一個可辨識的合成答案收尾），使我不必重跑整個回合；同時我仍看到與今天**完全相同**的失敗訊息與結束碼。

**為何為此優先級**: 這是操作者的原始痛點與本輪的唯一交付目的 —— 把一次昂貴的失敗，從「全部丟棄」降為「保留已完成的工作且工作階段仍可續接」，同時把失敗面維持在現行的凍結片語 + 結束碼，風險面最小。

**獨立驗證方式**: 以 in-process fake provider 腳本化「先完成一個工具步驟、下一次推論持續失敗（重試耗盡）」、「下一次推論以 400 拒絕（不可重試）」，斷言 `history.jsonl` 新增恰好一筆（`steps` = 已完成的步驟、`answer` = 合成訊息、`calls` = 已完成的推論輪數）、`stderr` 出現資訊列、類別片語仍在、結束碼不變（6），且下一次執行的重播以 assistant 收尾。

**驗收情境**:

1. **Given** a turn in which a tool step completed and the next provider request **keeps failing** (a retryable failure the retry cannot absorb), **When** the operator runs the prompt, **Then** `history.jsonl` gains **exactly one** entry carrying the original `prompt`, the completed `steps`, `calls` = the completed inference rounds, and a **synthetic** `answer`; and `stderr` carries the unchanged `the provider request failed` phrase plus an informational "kept … step(s)" line; and the run exits **6**.
2. **Given** a turn in which a tool step completed and the next provider request is **rejected outright** (a non-retryable 4xx), **When** the operator runs the prompt, **Then** the completed `steps` are persisted (one entry, synthetic close) and the run still exits **6** with the frozen phrase.
3. **Given** the session after such a failed-but-kept turn, **When** the operator runs the next prompt, **Then** the replayed conversation ends the earlier turn with an **assistant** message (the synthetic answer) so role alternation holds on **both** families.
4. **Given** a turn that fails with **zero** completed tool steps, **When** the operator runs the prompt, **Then** **no** entry is written and the failure surface is unchanged (phrase + exit 6).

**功能需求（FR）**:

- **FR-001**: 當回合以 `err != nil` 結束，**且** `len(result.Steps) > 0`，**且**失敗**不是**操作者中斷，系統 MUST 以**一個**原子 `history.Entry` 將該回合持久化，其 `answer` 為一個**合成的收尾訊息**，其 `steps` 為已完成步驟、`calls` 為已完成的推論輪數。
- **FR-002**: 該失敗回合的失敗面 MUST 維持不變：凍結類別片語（`the provider request failed` 或 `the tool request failed`）與結束碼（**6** / **7**）；MUST NOT 新增片語或結束碼。
- **FR-003**: 持久化後的該回合，經 `BuildMessages` 重播 MUST 以 **assistant** 訊息收尾（角色交替在兩家族皆合法）。
- **FR-004**: 失敗回合 MUST NOT 寫入任何 `stdout`；資訊列 MUST 僅在 `stderr`、控制碼安全、且 MUST NOT 攜帶 `tellme: ` 類別前綴。

**非功能需求（NFR）**:

- **NFR-001**: 合成收尾訊息 MUST 為控制碼安全（control-free）且 MUST NOT 洩漏憑證。

---

### 使用者情境 2 - 零步驟失敗維持乾淨中止 (Priority: P2)

作為一個操作者，當回合在**尚未完成任何工具步驟**時就失敗（例如第一次推論即失敗），我希望 tellme **不要**留下任何半截的 history entry —— 保持現行的「乾淨中止」。

**為何為此優先級**: 它是 US1 的安全邊界 —— 沒有它，一次「什麼都還沒完成」的失敗就會寫出一筆以合成答案收尾、卻毫無工具內容的空殼回合，污染後續上下文。

**獨立驗證方式**: 以 in-process fake provider 腳本化一個在**首個工具完成前**即失敗的回合，斷言 `history.jsonl` **不變**。

**驗收情境**:

1. **Given** a turn that fails **before any tool step completed**, **When** the operator runs the prompt, **Then** **no** entry is written to `history.jsonl` and the failure surface is unchanged.

**功能需求（FR）**:

- **FR-005**: 當回合失敗且 `len(result.Steps) == 0` 時，系統 MUST NOT 寫入任何 `history.jsonl` entry。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`verify-no-network` MUST NOT 被破壞；E2E MUST NOT 需要 pty 或真實網路。

---

### 邊界情況

- 當失敗為 **`agentport.ErrIncomplete`**（`MAX_TOOL_LOOP` 到達、未知工具上限耗盡、或無工具註冊）**且**已有完成的步驟時，系統 MUST 依同一規則保留（S-1 → B），且失敗面維持不變（`the tool request failed` + exit 7）。
- 當操作者在**round-078 重試等待期間**按 `Ctrl+C` 時，系統 MUST 視為**操作者中斷**（`ctx.Err() != nil`），走中斷路徑（round 079 的 exit 0 + 中斷資訊列）—— 修補 issue 的 hole #2。
- 當失敗回合的合成 `store.Append` **失敗**時，系統 MUST 退回現行的失敗面（原始錯誤 + 片語 + 原結束碼），MUST NOT 崩潰、MUST NOT 留下半截或損毀的 `history.jsonl`。
- 當**同一次執行**因失敗而寫入了一筆部分回合後，MUST NOT 再為同一回合寫入第二筆（恰好一次）。
- 當失敗回合內已完成的呼叫有 usage 時，是否寫入 `tokens.log` 為研究決策（S-5）。
- 當回合在 `--new` 或 `-i`（互動提示）路徑上失敗時，同一語意 MUST 成立。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 中斷（`ctx.Err() != nil`）與失敗兩種部分回合持久化 MUST 共用**同一個** `runTurn` 接縫（單一實作），MUST NOT 於多處重複。
- **FR-007**: 合成 entry MUST 為單一 append-only 行，且 MUST NOT 改變 `history.Entry` / `history.Step` 的既有 JSON schema。

#### 非功能需求

- **NFR-003**: 結束碼集合 MUST 維持十個；片語詞彙 MUST 維持凍結；MUST NOT 因本輪新增任何片語或結束碼。
- **NFR-004**: 中斷偵測 MUST 為結構化（`ctx.Err()`），失敗分類 MUST 為型別化（既有錯誤路徑）；MUST NOT 以錯誤訊息字串比對判定。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`history.Entry` / `history.Step`（`internal/domain/history`）**: 既有形狀 —— **schema 不變**；合成答案僅是 `answer` 的值；`steps` 為已完成步驟。
- **合成收尾訊息（synthetic closing answer）**: 固定、可辨識的字串常數（確切文字 = S-4 / S-6）—— 一個用於操作者中斷（round 079 現有），一個用於失敗（本輪新增，若 S-4 採 distinct）。
- **`runTurn` 的失敗持久化分支**: 在既有中斷分支之外，新增「失敗且 `len(steps) > 0`」的持久化，再走既有 `emitProviderError` / `emitToolError`；本輪新增。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 在 hermetic E2E 中，一個「已完成一個工具步驟、隨後推論持續失敗（重試耗盡）」的回合使 `history.jsonl` **新增恰好一筆**（`steps` = 1、`answer` = 合成訊息、`calls` = 已完成的推論輪數），且結束碼為 **6**、`stderr` 同時出現凍結片語與資訊列。
- **SC-002**: 一個不可重試失敗（400）的等效回合同樣被持久化（恰好一筆）且結束碼為 **6**。
- **SC-003**: 一個在首個工具步驟前失敗的回合使 `history.jsonl` **不變**（無新 entry）。
- **SC-004**: 中斷（含 round-078 重試等待期間的 `Ctrl+C`）仍走 round 079 的中斷路徑（exit 0 + 中斷資訊列）—— hole #2 修補後仍成立。
- **SC-005**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），且 `TestExitCodesMatchPinnedContract` 綠、`stdout` 既有契約位元組不變。

## 假設

- **A1**: 本輪 anchor = issue [#161](https://github.com/gosharplite/tellme/issues/161)；**DoD = 關閉它**。
- **A2**: 核心行為 —— 失敗且 `len(steps) > 0` 則以合成答案收尾寫入、失敗面（片語 + 結束碼）不變、零步驟不寫、續接角色交替 —— 由 issue 與操作者指令鎖定；issue 的 *Open Design Decisions*（範圍 A/B、結束碼、合成文字、usage、資訊列文字、`ctx.Err()` 改寫）由**操作者指示交由 `/axb-technical-research`**，故本輪 clarify **未升級（0 題）**。研究決議若改變正式驗收契約（例如範圍 A vs B），MUST 回寫 truth 並於 `truth-delta.md` 記錄。
- **A3**: `ctx.Err() != nil` 可用於辨識操作者中斷（`runTurn` 的 `defer cancel()` 在此分支之後才執行）。
- **A4**: 本輪延伸 round 079 的 `docs/domain-model` 例外（`history-append-after-complete`、`turn-answer-stored-verbatim`）至**失敗**回合（same-PR，ADR 0041）；預期新增 **ADR 0052**。
- **A5**: MCP 工具步驟與原生步驟同構；`-b`/`--retry` 與離線讀取器不受影響。
