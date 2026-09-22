# Feature Specification: Preserve completed tool steps on an interrupted turn (round 079)

**Feature Branch**: `079-interrupted-turn-partial-persistence`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the core behaviour is anchored by issue [#159](https://github.com/gosharplite/tellme/issues/159) and the operator's directive, and the issue's *Open Design Decisions* are **explicitly routed by the operator to `/axb-technical-research`**)

**Input (anchor issue [#159](https://github.com/gosharplite/tellme/issues/159), 2026-09-22)**: *"When an operator interrupts `tellme` with `Ctrl+C` (`SIGINT`) during a multi-round tool loop (e.g. at step 8 of 10), `runTurn` receives `context.Canceled` and exits with `ProviderError` (exit code 6) without persisting the turn. Per the `history-append-after-complete` invariant, the entire turn — the prompt and all 7 previously completed tool steps — is discarded from `history.jsonl`."* — plus the operator directive: *"Open a new round, the goal is to close issue #159."*

**Observed context (`dev` @ `4f9c96d`)**: `internal/cli/cli.go` (`runTurn`) opens the interrupt context (`signal.NotifyContext(…, os.Interrupt, syscall.SIGTERM)`, line 743); on a loop error it maps `*agentport.ErrIncomplete` → tool phrase + exit 7, **everything else → `emitProviderError` (phrase + exit 6)**, and **skips `store.Append`** (the append is only reached on `err == nil`). The completed steps exist in memory at cancellation (`internal/agent/agentloop.go` — `steps = append(steps, history.Step{…})`) but are returned in `agentport.Result{Steps: steps, Calls: calls}` and then discarded by the caller.

**Behaviour intent**: **MODIFY (session-history persistence)** — an operator-interrupted turn that has completed at least one tool step is **closed with a synthetic assistant answer** so it can be persisted as a structurally valid, resumable history entry; a turn interrupted with **zero** completed steps keeps today's **clean abort** (nothing is written). The frozen class-phrase vocabulary and the exit-code set are **unchanged** (their exact values on the interrupted path are a `/axb-technical-research` decision). **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The wire-role-alternation constraint is the whole reason the synthetic answer exists (I-1).** `history.jsonl` is replayed verbatim into the next turn's context by `BuildMessages` (`internal/agent/agentloop.go`): `user(prompt) → [assistant(tool_calls) → tool(result)]… → assistant(answer)`. A partial save that ended on a `tool` result would produce two consecutive `user` roles on the next request — which **Gemini/Vertex strictly rejects** (`functionResponse` parts ride under `role:"user"`) and **OpenAI-compatible** rejects by alternation. Therefore an interrupted turn is only safe to persist if it is **closed with an assistant message**.
- **A partial turn is persisted as an ordinary history entry.** The synthetic closing message is just the entry's `answer`; the schema of `history.Entry` / `history.Step` is **unchanged** (`{prompt, answer, calls, steps}`), so no new field and (most likely) **no data-model change** (confirm at `/axb-technical-research`).
- **Discard on zero steps (I-2).** If no tool step was recorded (`len(result.Steps) == 0`), **nothing** is written to `history.jsonl` — the current clean-abort behaviour is preserved byte-for-byte.
- **Child processes are still killed immediately (I-3).** In-flight `execute_command` children must still be terminated via the process-group `SIGKILL` the loop already owns (`ctx.Done()` → the per-call deadline path). This round MUST NOT weaken, delay, or route around that kill.
- **The interruption is a *context cancellation* (I-6).** Tellme already cancels the turn context on `SIGINT`/`SIGTERM` (`signal.NotifyContext`). Detection MUST be **typed** (`errors.Is(err, context.Canceled)`) — **never** string-sniffing the message ("context canceled").
- **The exit code / diagnostic on a *successful partial save* is an open decision (S-7).** Today the path prints `tellme: the provider request failed: context canceled` and exits **6**. The issue asks whether a successful partial save should exit 0, 130 (`128+SIGINT`), or keep 6 — and what line (if any) appears on `stderr`. This is a user-facing contract (exit codes are pinned by `TestExitCodesMatchPinnedContract`) → **research decision (S-7)**, to be reconciled with truth; **no new class phrase**.
- **The append is best-effort-safe (edge case).** If the synthetic `store.Append` itself fails, the run MUST fall back to today's failure surface (the original provider failure → phrase + exit per S-7) rather than crash or half-write; the exited process must not leave a corrupt `history.jsonl`.
- **Hermeticity holds (NFR-002).** The E2E simulates the interruption **without** a real network or a pty; how it deterministically lands a `SIGINT` mid-turn (e.g. the fake provider stalls the `N+1`-th request until the signal arrives) is a research/tasks decision (**S-9**).

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `4f9c96d`)*

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go:743` | `ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` — the turn context; `defer cancel()`. |
| `internal/cli/cli.go:828-842` (`runTurn`) | `result, err := loop.Run(ctx, prompt, prior)`; `ind.Stop()`; `if err != nil { … ErrIncomplete → emitToolError; else emitProviderError }` → **`store.Append` is skipped** on any error. On success: `store.Append(history.Entry{Prompt, Answer, Calls: len(result.Calls), Steps: result.Steps})`. |
| `internal/agent/agentloop.go` (`Run`) | accumulates `steps []history.Step` and `calls []llm.Usage`; on `a.Gateway.Complete` error returns `agentport.Result{Steps: steps, Calls: calls}, err` (the completed steps survive in the returned value). A tool that returns (success **or** error) is appended to `steps` after its result is folded back. |
| `internal/domain/agent/result.go` | `Result{Answer, Steps []history.Step, Usage, Calls []llm.Usage}`; `ErrIncomplete{Reason, Err}`. |
| `internal/domain/history/history.go` | `Entry{Prompt, Answer, Calls, Steps []Step}` (JSON `prompt/answer/calls,omitempty/steps,omitempty`); `Step{Tool, Arguments, Result, Signature,omitempty}`. `Store{Load, Append(Entry), Archive}`. |
| `internal/agent/agentloop.go` (`BuildMessages`) | replays each entry as `user(prompt)` → per step `assistant(tool_call)` + `tool(result)` → `assistant(answer)`. The last message of a valid turn is `assistant`. |
| `internal/cli/cli.go` (`emitProviderError` / `emitToolError`) | `tellme: the provider request failed: <detail>` + exit 6 · `tellme: the tool request failed: <detail>` + exit 7 (newlines folded to one line). |
| `internal/cli/exitcode.go:22-28` | `ProviderError = 6`; the ten-value exit-code set pinned by `TestExitCodesMatchPinnedContract`. |
| `internal/cli/call_renderer.go:223` (`persistTurnUsage`) | writes the turn's usage in ONE `AppendBatch` for the **Reported** subset of `result.Calls`, gated on `res.Workspace != "" && result.Usage.Reported`. On an interrupted turn `result.Usage` is zero (`Reported == false`) unless S-8 decides otherwise. |
| `internal/cli/cli.go` (`run`, the interactive reader) | a `SIGTERM` **during the read** cancels and returns `Success` (0) — the *input-capture* convention; the comment there claims a matching "runTurn convention" for an interrupted prompt turn, but `runTurn` today maps `context.Canceled` → **exit 6**. This tension is **S-7**. |
| `tests/e2e/fakeprovider` | scriptable in-process OpenAI-compatible fake: `Answer`, `ErrorStatus`, `Script(replies…)` (tool-call scripts), `DropFirst(n)` (round-078 transport drop), `served`/`RequestCount`/`Snapshot`. **No stall/hang mode yet** — a research/tasks item for the interruption witness (S-9). |
| `tests/e2e/harness/cmd_helper.go` | runs the built binary as a subprocess (`exec.CommandContext`/`cmd.Start`/`cmd.Wait`); `RunInWithSyncedStdin` observes a marker before driving stdin — the precedent for synchronising a signal with a mid-turn state (S-9). |
| `specs/truth/features/cli/chat/**` + `docs/domain-model/tellme.modelith.{yaml,md}` | the provider-failure feature and the `history-append-after-complete` / `turn-answer-stored-verbatim` invariants — the owning surfaces this round likely MODIFIES (see S-10). |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **Persist on interrupt with completed steps** — when the turn context is cancelled and `len(result.Steps) > 0`, append ONE atomic `history.Entry` closing the turn with a **synthetic assistant answer**. | **locked (issue)** |
| **S-2** | **Discard on zero steps** — when `len(result.Steps) == 0`, write **nothing** (today's clean abort). | **locked (issue)** |
| **S-3** | **Child-process termination unchanged** — in-flight `execute_command` children are still `SIGKILL`-ed via the process group immediately (I-3). | **locked (issue)** |
| **S-4** | **Role alternation on resume** — the persisted partial turn MUST replay to a valid `user … assistant` sequence accepted by both families (I-1). | **locked (issue)** |
| **S-5** | **Default vs. gated** — whether `Ctrl+C` always saves when `len(steps) > 0`, or a config toggle / dual-interruption (abort-and-discard) is warranted. | research decision (D-x) |
| **S-6** | **The synthetic answer text** — the standardised closing phrase (e.g. `[Turn interrupted by operator via Ctrl+C]`). | research decision (D-x) |
| **S-7** | **Exit code & diagnostic** — a successful partial save's exit code (0 / 130 / 6) and its `stderr` line; reconciles the `runTurn`-vs-reader tension above; **no new class phrase**. | research decision (D-x) |
| **S-8** | **Token-usage accounting** — whether the completed calls' usage is persisted to `tokens.log` for the interrupted turn (via the existing one-`AppendBatch` shape). | research decision (D-x) |
| **S-9** | **The hermetic interruption seam** — how the E2E deterministically lands a `SIGINT` mid-turn (a fake-provider stall on the `N+1`-th request + a synced signal; a scripted marker), with no pty and no live network. | research decision (D-x) |
| **S-10** | **Records** — a new ADR + index (and/or an amendment to the round-007 history record), the `specs/truth/**` rows to MODIFY, the CLI feature + DSL rows, and the domain-model amendment call (ADR 0041 — the `history-append-after-complete` / `turn-answer-stored-verbatim` synthetic-answer exception; likely MODIFY, not NOOP). | research decision (D-x); truth written by the owner skills |
| **S-11** | **Scope** — the **provider/tool-loop** turn only; MCP-backed tools are ordinary steps on the same seam (no MCP-specific handling). `-b`/`--retry` and the offline readers are untouched. | **locked** |
| **S-12** | **The step trigger's exact meaning** — `len(steps) > 0` counts a step the loop **recorded** (a tool that returned, even with an error result from a killed call). Whether the trigger should require a *successful* (non-error) step is open. | research decision (D-x) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Role alternation on resume.** The persisted partial turn must replay to `… assistant(answer)`, so the next request's role sequence is valid on **both** the OpenAI-compatible and Gemini/Vertex wires (every tool call paired; the turn ends with an assistant message before the next user prompt).
- **I-2 — Discard on zero steps.** `len(result.Steps) == 0` ⇒ **no** `history.jsonl` write.
- **I-3 — Clean child termination.** In-flight child processes are still terminated immediately (process-group `SIGKILL`); this round does not weaken or delay it.
- **I-4 — Deterministic & tested.** Covered by executable Gherkin with a **simulated** interruption, asserting `history.jsonl`, `turns.log`, the exit code, and subsequent-turn resumption.
- **I-5 — Single append-only write.** A partial turn is persisted as **exactly one** line; no schema change to `history.Entry` / `history.Step`; no double write; best-effort (an append failure degrades to today's failure surface, never a corrupt file).
- **I-6 — Typed detection.** Interruption is detected by the **typed** cancellation (`errors.Is(err, context.Canceled)`), never by string matching.
- **I-7 — Frozen vocabulary / surface-neutral.** No new class phrase; `stdout` byte-unchanged (the interruption path writes no `stdout`); any diagnostic is on `stderr` only and control-free.
- **I-8 — Hermetic + stdlib-only + POSIX-only.** No new dependency; `verify-no-network` intact; the E2E needs no pty and no live network.

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 中斷的回合保留已完成的工具步驟並以合成答案收尾 (Priority: P1)

作為一個執行長時間工具回合（大量的程式碼搜尋、多檔分析、透過 `execute_command` 跑測試／建置）的操作者，當我在第 8 步按下 `Ctrl+C` 時，我希望 tellme **保留前 7 個已完成的工具步驟**、以一個可辨識的合成答案把回合收尾，讓我能在同一個工作階段用 *"continue"*／*"summarize what you found so far"* 續接，而不是把所有已收集的上下文通通丟掉、從頭再來。

**為何為此優先級**: 這是本輪唯一的交付目的 —— 把一次昂貴的中斷，從「全部丟棄」降為「保留已完成的工作且工作階段仍可續接」；同時把結構風險（角色交替）以合成答案收斂。

**獨立驗證方式**: 以 in-process fake provider 腳本化一個多輪工具回合，在**第 N 次推論請求時停住**，於此時送出 `SIGINT`；斷言 `history.jsonl` **新增恰好一筆** entry（其 `steps` = 中斷前完成的步驟、`answer` = 合成訊息、`calls` = 已完成的推論輪數），並斷言下一次執行的重播訊息序列以 assistant 收尾、角色交替合法（兩家族）。

**驗收情境**:

1. **Given** a multi-round tool turn in which **at least one** tool step has completed, **When** the operator interrupts the turn with `SIGINT` (the turn context is cancelled), **Then** tellme appends **exactly one** `history.Entry` to `history.jsonl` carrying the original `prompt`, the completed `steps` (1..N), a `calls` count of the completed inference rounds, and a **synthetic** `answer`.
2. **Given** the session after such an interruption, **When** the operator runs the next prompt, **Then** the replayed conversation (`BuildMessages`) ends the prior turn with an **assistant** message and the request's role sequence alternates legally on **both** the OpenAI-compatible and Gemini/Vertex wires.
3. **Given** the interrupted turn's in-memory result, **When** it is persisted, **Then** the entry's `steps` equal **exactly** the steps the loop recorded before cancellation (no fabricated, duplicated, or dropped step), and the entry's `calls` equals `len(result.Calls)`.
4. **Given** the operator interrupted the turn, **When** the process exits, **Then** the completed work is durable (`history.jsonl` holds it) and the surface behaviour (exit code / diagnostic / `stdout`) matches the **S-7** decision, with no new class phrase.

**功能需求（FR）**:

- **FR-001**: 當回合的 context 因操作者中斷（`SIGINT`/`SIGTERM`）而被取消，**且** `len(result.Steps) > 0` 時，系統 MUST 以**一個**原子 `history.Entry` 將該回合持久化到 `history.jsonl`，其 `answer` 為一個**合成的收尾訊息**。
- **FR-002**: 合成 entry 的 `steps` MUST 等於中斷前迴圈**已記錄**的工具步驟（1..N），`calls` MUST 等於已完成的推論輪數（`len(result.Calls)`）；MUST NOT 捏造、複製或漏掉任何步驟。
- **FR-003**: 持久化後的該回合，經 `BuildMessages` 重播 MUST 以 **assistant** 訊息收尾，使下一個請求的角色序列在法律上有效（OpenAI-compatible 與 Gemini/Vertex 兩家族皆然）。
- **FR-004**: 中斷偵測 MUST 為**型別化**的 context 取消判定（`errors.Is(err, context.Canceled)`）；MUST NOT 以錯誤訊息字串比對來判定中斷。

**非功能需求（NFR）**:

- **NFR-001**: 合成的收尾訊息 MUST 為控制碼安全（control-free）且 MUST NOT 洩漏憑證（比照 round-039 / round-056 的既有政策）。

---

### 使用者情境 2 - 零步驟中斷維持乾淨中止 (Priority: P2)

作為一個在**首次推論期間**（尚未有任何工具完成）就改變心意、想放棄一個失控或幻覺迴圈的操作者，我希望 tellme **不要**留下任何半截的 history entry —— 保持現行的「乾淨中止：不寫入」。

**為何為此優先級**: 它是 US1 的安全邊界 —— 沒有它，一次「什麼都還沒完成」的中斷就會寫出一筆以合成答案收尾、卻毫無工具內容的空殼回合，污染後續上下文；它與 US1 是同一個機制的必要另一半。

**獨立驗證方式**: 以 in-process fake provider 腳本化一個**尚未**完成任何工具就停住的回合，送出 `SIGINT`；斷言 `history.jsonl` **不變**（無新 entry）。

**驗收情境**:

1. **Given** a turn interrupted **before any tool step completed** (e.g. during the first inference, or before the first tool returns), **When** the operator sends `SIGINT`, **Then** **no** entry is written to `history.jsonl` (today's clean abort is preserved).
2. **Given** such a zero-step interruption, **When** the process exits, **Then** the surface behaviour (exit code / diagnostic / `stdout`) matches the **S-7** decision and the frozen phrase vocabulary is unchanged.

**功能需求（FR）**:

- **FR-005**: 當回合被中斷且 `len(result.Steps) == 0` 時，系統 MUST NOT 寫入任何 `history.jsonl` entry（維持現行乾淨中止語意）。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`make verify` 的 `verify-no-network` MUST NOT 被破壞；E2E MUST NOT 需要 pty 或真實網路。

---

### 邊界情況

- 當取消發生在**工具執行中**（一個工具已啟動但尚未返回）時，該工具的 context 被取消、其結果為錯誤文字，迴圈**仍會記錄一個 step**（其 `result` 為錯誤訊息）—— 本輪的觸發條件以 `len(result.Steps) > 0` 為準；「是否應要求一個**成功**的 step」為研究決策（S-12）。
- 當取消發生在**提供者呼叫期間且尚無任何工具完成**時，`len(result.Steps) == 0` ⇒ MUST NOT 寫入（US2）。
- 當合成 `history.Entry` 的 `store.Append` **失敗**時，系統 MUST 退回今天的失敗面（原始提供者失敗 → 片語 + S-7 的結束碼），MUST NOT 崩潰、MUST NOT 留下半截或損毀的 `history.jsonl`。
- 當同一回合在取消後又被重試或重播時，系統 MUST NOT 為同一個中斷回合寫入**兩次**（恰好一次，I-5）。
- 當取消與 round-078 的重試等待重疊時，取消 MUST 勝出（不得在取消後仍發出下一次嘗試）—— 沿用 round-078 的 cancellation-aware 契約。
- 當被中斷的回合含 **MCP** 工具步驟時，這些步驟與原生步驟同構（同一接縫、同一記錄方式）。
- 當被中斷的回合是 `--new` 或 `-i`（互動提示）路徑時，同一中斷語意 MUST 成立。
- 當被中斷的回合內**已完成呼叫有 usage**時，是否將其一併寫入 `tokens.log` 為研究決策（S-8）。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 中斷路徑 MUST 涵蓋**兩**個提供者家族（同一個 `runTurn` 接縫），MUST NOT 於各家族分別重複實作。
- **FR-007**: 合成 entry MUST 為單一 append-only 行，且 MUST NOT 改變 `history.Entry` / `history.Step` 的既有 JSON schema（合成答案即為 `answer` 欄位）。
- **FR-008**: 合成 entry 的 `calls` MUST 與既有回合計數語意一致（已完成的 AI-endpoint 推論輪數）；MUST NOT 因中斷而把未完成的推論計入。

#### 非功能需求

- **NFR-003**: 中斷路徑 MUST NOT 改變 `stdout` 的既有契約（不寫任何 `stdout`）；任何診斷 MUST 僅在 `stderr`。
- **NFR-004**: 結束碼與片語詞彙 MUST 維持凍結（除 S-7 對中斷路徑的明確裁定外）；MUST NOT 新增類別片語。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`history.Entry` / `history.Step`（`internal/domain/history`）**: 本輪持久化的既有資料形狀 —— **schema 不變**，合成答案只是 `answer` 的值；`steps` 為已完成步驟。
- **合成收尾訊息（synthetic closing answer）**: 一個固定的、可辨識的字串常數（確切文字 = S-6），作為被中斷回合的 `answer`。
- **中斷偵測接縫（in `runTurn`）**: 在 `loop.Run` 返回後，以 `errors.Is(err, context.Canceled) && len(result.Steps) > 0` 判定「部分回合可持久化」；本輪新增。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 在 hermetic E2E 中，一個在第 N 次推論請求時被 `SIGINT` 中斷的多輪工具回合，使 `history.jsonl` **新增恰好一筆** entry，其 `steps` = 中斷前完成的步驟、`answer` = 合成訊息、`calls` = 已完成的推論輪數。
- **SC-002**: 中斷後的後續執行所送出的訊息序列以 **assistant** 收尾且角色交替合法（兩家族，經 fake provider 斷言）。
- **SC-003**: 一個**零步驟**中斷使 `history.jsonl` **不變**（無新 entry）。
- **SC-004**: 凍結的片語詞彙不變；`TestExitCodesMatchPinnedContract` 綠（S-7 若裁定新碼則一併更新其 pinned contract 並於 truth 記錄）。
- **SC-005**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），且 `stdout` 既有契約位元組不變。

## 假設

- **A1**: 本輪 anchor = issue [#159](https://github.com/gosharplite/tellme/issues/159)；**DoD = 關閉它**。
- **A2**: 核心行為 —— 有步驟則以合成答案收尾寫入、無步驟則不寫、child process 立即終止、續接角色交替 —— 由 issue 鎖定；issue 標示的 *Open Design Decisions*（預設 vs 閘控、合成文字、結束碼與診斷、usage 計帳、domain-model 修訂）**由操作者明確交由 `/axb-technical-research`**，故本輪 clarify **未升級（0 題）**。
- **A3**: MCP 工具步驟與原生步驟同構（同一接縫，無 MCP 專屬處理）。
- **A4**: 本輪可能 MODIFY `docs/domain-model/**`（`history-append-after-complete` 與 `turn-answer-stored-verbatim` 的合成答案例外）與 `specs/truth/**`（history/provider-failure 相關列），視研究而定（ADR 0041 的 same-PR 規則）。
- **A5**: 合成收尾訊息為固定片語；其確切文字為研究決策（S-6）。
- **A6**: 若研究發現某項開放決策會改變正式驗收契約（例如新增結束碼），該項須回寫 truth 並於 `truth-delta.md` 記錄；必要時以釐清／修訂收斂，MUST NOT 默默落地。
