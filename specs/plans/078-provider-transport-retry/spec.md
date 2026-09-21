# Feature Specification: Retry a transient provider transport failure before failing (round 078)

**Feature Branch**: `078-provider-transport-retry`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the retryability predicate + the schedule are **operator-locked** in-session; residual technical choices → `/axb-technical-research`)

**Input (operator request, 2026-09-22, in-session)**: *"Sometimes see connection drop by remote server in tellme. … Many times it is just network disruption, redo command will just work. Can we have a simple wait 1 sec → retry → wait 3 sec → retry → error?"* — and the operator's follow-up agreement on the retryability predicate: **retry transport failures + HTTP 429/5xx only.**

**Observed symptom (operator)**: a prompt turn aborts with `tellme: the provider request failed: …` (**exit 6**) on a momentary network blip; re-running the same command succeeds. Today tellme has **no retry layer at all** — one provider call, one failure, terminal.

**Behaviour intent**: **MODIFY (provider failure handling)** — a *retryable* provider failure is retried automatically (**at most twice**, 1 s then 3 s) before the run fails with the frozen provider phrase + exit **6**; a *non-retryable* failure is **unchanged** (fails immediately, one call). The frozen class-phrase vocabulary and the exit-code set are **unchanged**. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The frozen phrase and exit code do NOT change (I-1).** The exhaustion path is exactly today's: `tellme: the provider request failed: <detail>` on `stderr`, exit code **6** (`internal/cli/cli.go` `emitProviderError`; `internal/cli/exitcode.go:27`). No new class phrase; the vocabulary stays ten.
- **tellme has no error taxonomy, and the HTTP status is not typed (the honest hazard).** `llm.ProviderError` carries only `{Provider string; Err error}` (`internal/domain/llm/gateway.go:95-112`). Today a non-2xx status exists **only inside a formatted string** — `fmt.Errorf("provider returned status %d: %s", resp.StatusCode, msg)` (`openai/client.go:88-93`, `gemini/client.go:100-105`) — so a status-based retry predicate would otherwise force **string sniffing**, which this round forbids (**NFR-003**). The round therefore owes a **typed** classification (a status field / a retryability query on the error, or a dedicated retryable sentinel) — a `/axb-technical-research` decision (**S-6**).
- **Retryable vs. not (operator-locked).** Retry the **transport** class (dial failure, `EOF`/`unexpected EOF`, connection reset by peer, broken pipe, HTTP/2 `GOAWAY`, TLS handshake failure, request timeout) **and HTTP 429 + 5xx**. Do **not** retry 4xx (400/401/403/404/422…), content-filter rejections, decode/parse errors, or the round-030 **output-cap truncation guard** (a successful-but-truncated reply is terminal — tellme "adds **no** retry layer"; this round must reconcile that statement, not silently contradict it).
- **The schedule is fixed (operator-locked).** attempt 1 → **1 s** → attempt 2 → **3 s** → attempt 3 → fail. **Two** retries, **three** attempts total — bounded, never unbounded (**I-2**).
- **Cancellation is respected (I-3).** The turn's context is cancelled on SIGINT/SIGTERM (`signal.NotifyContext`). A parent-context cancellation MUST abort immediately — never sleep through it, never retry after it.
- **A retry is invisible to accounting (I-4).** A retried call re-sends the **same** request; the round-027 turn counter counts **AI-endpoint calls** and the round-030 wording already says "an internal retry … does not count". A retry MUST be **one** `calls` entry, **one** `usage` record, **no** second turn frame, and **no** history write for a failed attempt.
- **MCP is out of scope (S-9).** A remote MCP failure already has its own behaviour (discovery warn+skip; an MCP tool-call failure is a *recoverable* fold-back the model can retry). This round is the **provider** path only.
- **`-b`/`--retry` is unrelated.** That is the reference's user-facing *"re-run the last message"* flag — a settled exclusion (`specs/truth/techstack.md`). This round is an automatic **transport** retry; the two must not be conflated.
- **Hermeticity holds (NFR-002).** The sleeps live in **production** code, so `verify-no-test-sleep` (which bans `time.Sleep` in *test* files) is **not** implicated. The retry must not break `verify-no-network`; the E2E proves it through the in-process fake provider only. Whether the E2E can run the real 1 s/3 s or needs a **fast-retry seam** (an injected sleeper / clock port) is a research decision (**S-7**).

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `dd9b94b`)*

| Site | Current shape |
| --- | --- |
| `internal/domain/llm/gateway.go:95-112` | `ProviderError{Provider string; Err error}` — the single typed provider/transport failure; `Error()`/`Unwrap()`; **no status field**. |
| `internal/infrastructure/llm/openai/client.go:63-104` | `Complete` sends **exactly one** request (`c.http.Do`); every failure → `c.wrap(err)` → `*ProviderError`; a non-2xx becomes `fmt.Errorf("provider returned status %d: %s", …)`. `defaultTimeout = 300s` (line 23). |
| `internal/infrastructure/llm/gemini/client.go:72-114` | the same shape for the Gemini/Vertex family (`defaultTimeout = 300s`, line 31). |
| `internal/agent/agentloop.go` (Run) | `resp, err := a.Gateway.Complete(ctx, req)`; on `err != nil` it returns the error **unchanged** — no classification, no retry. |
| `internal/cli/cli.go:743` | `gw = withUnpairedDiagnostic(gw, unpairedEmitter(…))` — the **existing gateway-decorator precedent** (`unpaired_gateway.go`, round 068); a retry wrapper slots at the **same seam** (one place, both families). |
| `internal/cli/cli.go` (`runTurn`) → `emitProviderError` | the loop error is mapped: `*agentport.ErrIncomplete` → tool phrase + **7**; everything else → `the provider request failed` + **6**. |
| `internal/cli/exitcode.go:22-28` | `ProviderError = 6`; the exit-code set (ten values) is pinned by `TestExitCodesMatchPinnedContract`. |
| `specs/truth/techstack.md:102` (round-030 row) | *"tellme adds **no** retry layer (there is none to change)"* — **becomes false** this round; MODIFY, not append. |
| `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | the existing provider-failure feature — the natural home for the new retry Rule/Examples. |
| `tests/e2e/fakeprovider` | scriptable in-process OpenAI-compatible fake: `Answer`, `ErrorStatus(code)`, and a **`served` request counter**; **no transport-drop / fail-once-then-succeed mode yet** — a research/tasks item for the retry witness. |
| `specs/truth/features/cli/chat/dsl.md` (§ round 030) + `techstack.md:45` | the round-034 "an internal retry (backoff / empty-response) does **not** count" wording — the hook this round's accounting semantics hang on. |

---

## Design (locked by the operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **What is retryable** — transport failures + HTTP **429** + **5xx**; **not** 4xx / auth / content-filter / decode / the round-030 truncation guard. | **locked (operator)** |
| **S-2** | **The schedule** — at most **two** retries: wait **1 s**, retry; wait **3 s**, retry; then fail with the frozen phrase + exit **6**. | **locked (operator)** |
| **S-3** | **Observability** — how (and whether) the wait is surfaced; the default is a plain line on the **diagnostic stream** (`stderr`) only — never `stdout`, never `turns.log`. | research decision (D-x) |
| **S-4** | **Accounting semantics** — one AI-endpoint call, one `usage` record, no second turn frame, no history write for a failed attempt (an internal retry does not count). | research decision (D-x); consistent with rounds 027/034 |
| **S-5** | **The seam** — a single **gateway decorator** at the `runTurn` attachment point (the `withUnpairedDiagnostic` precedent), covering **both** families; the alternative (per-adapter retry) is the rejected shape (duplication in two transports). | research decision (D-x) |
| **S-6** | **The classification mechanism** — a **typed** retryability signal (a `Status` field / a `Retryable()` query on `ProviderError`, or a dedicated sentinel), **never** string matching. | research decision (D-x) |
| **S-7** | **The hermetic test seam** — how the E2E proves retry count/order without paying the real 1 s/3 s (an injected sleeper / clock port vs. accepting the delay), and how the fake provider scripts a transport drop / fail-once-then-succeed. | research decision (D-x) |
| **S-8** | **Records** — a new **ADR 0050** + index (or an amendment to the round-030 record), the `techstack.md` rows to MODIFY, the CLI feature + DSL rows, and the domain-model call (ADR 0041 — likely **not modelled**, recorded in `plan.md` §5). | research decision (D-x); truth written by the owner skills |
| **S-9** | **MCP out of scope**; `-b`/`--retry` unrelated (settled exclusion). | **locked** |
| **S-10** | **No new class phrase, no new exit code** — exhaustion reuses `the provider request failed` + exit **6**. | **locked** |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The frozen phrase and exit 6 are unchanged.** Exhaustion is byte-identical to today's failure line; the vocabulary stays ten; no new code.
- **I-2 — Bounded.** At most two retries (three attempts); never unbounded; the bound is a fixed constant (the round-076 `maxUnknownToolFolds` precedent).
- **I-3 — Cancellation-aware.** A parent-context cancellation (SIGINT/SIGTERM) aborts immediately — the round MUST NOT sleep through, nor retry after, a cancellation.
- **I-4 — Accounting-invisible.** A retried call is **one** AI-endpoint call: one `calls` entry, one `usage` record, no second turn frame, no history write for a failed attempt (the round-034 "an internal retry does not count" wording is honoured).
- **I-5 — Retry only the retryable class.** A non-retryable failure (4xx / auth / content-filter / decode / truncation) is attempted **once** and surfaces immediately — no pointless waits hiding a real error.
- **I-6 — Hermetic + surface-neutral.** stdlib-only, POSIX-only, no new dependency; `stdout` byte-exact; retry diagnostics on `stderr` only; `verify-no-network` intact (the gate makes no live call). The classification is **typed**, never string-based (**NFR-003**).
- **I-7 — One attachment point.** Both provider families are covered by **one** seam; the retry is not duplicated per transport.

---

## 使用者情境與測試 *(必填)*

### 使用者故事 1 - 一時的提供者連線中斷會自動重試 (Priority: P1)

作為一個使用 tellme 的操作者，當一次提供者請求因為**一時的網路中斷**（連線被重置、EOF、暫時的 429/5xx、請求逾時）而失敗時，我希望 tellme 會自動等待並重試（先 1 秒、再 3 秒），使得一次網路抖動不會逼我重新輸入並重跑整個指令；若三次都失敗，我才看到與今天**完全相同**的錯誤訊息。

**為何為此優先級**: 這是操作者的原始痛點與本輪的唯一交付目的 —— 把「常常只是網路抖動」從一次昂貴的失敗，降為一次自動恢復；同時把最終失敗面維持在現行的凍結片語 + exit 6，風險面最小。

**獨立驗證方式**: 以 in-process fake provider 腳本化「先傳輸失敗、再成功」、「連續失敗」，斷言 fake 被呼叫的次數（恰好 2 / 恰好 3）與最終結果（成功 / 片語 + exit 6），完全不需真實網路。

**驗收情境**:

1. **Given** a configured provider whose endpoint **drops the connection once** (a transport failure) and then answers, **When** the operator runs a prompt, **Then** tellme retries after **1 s**, the retry succeeds, the turn completes normally, and the endpoint was called **exactly twice**.
2. **Given** a configured provider whose endpoint **fails twice** (transport) and then answers, **When** the operator runs a prompt, **Then** tellme waits **1 s**, retries (fails), waits **3 s**, retries (succeeds) — the endpoint was called **exactly three times** and the turn completes.
3. **Given** a configured provider whose endpoint **always fails** with a retryable failure, **When** the operator runs a prompt, **Then** the endpoint is called **exactly three times** (2 retries, no more) and the run fails with `the provider request failed` on `stderr`, exit **6**, `stdout` empty and byte-exact.
4. **Given** a retryable failure was retried and then succeeded, **When** the turn completes, **Then** the turn's persisted `calls` count is **1** (the retry does not advance the counter) and exactly **one** usage record is written.

**功能需求（FR）**:

- **FR-001**: 系統 MUST 在提供者請求因**可重試的失敗類別**（傳輸層失敗、HTTP 429、HTTP 5xx）而失敗時，自動重試**同一個**請求；重試**至多兩次**。
- **FR-002**: 重試排程 MUST 為固定的 **1 秒**（第 1 次失敗後）與 **3 秒**（第 2 次失敗後）；第 3 次失敗即為最終失敗。
- **FR-003**: 重試耗盡後，系統 MUST 以現行的凍結類別片語 `the provider request failed`（`stderr`）與結束碼 **6** 失敗；MUST NOT 新增片語或結束碼。
- **FR-004**: 重試 MUST 送出與原請求**內容相同**的請求（相同的 messages / tools / persona / 參數）；MUST NOT 在重試時改變請求語意。

**非功能需求（NFR）**:

- **NFR-001**: 重試的可觀察性 MUST 僅限**診斷流（`stderr`）**；`stdout` MUST 維持位元組精確；`turns.log` MUST NOT 增加重試行（或 MUST NOT 改變其既有內容契約）。

---

### 使用者故事 2 - 不可重試的失敗不被無謂重試 (Priority: P2)

作為一個使用 tellme 的操作者，當一次提供者請求因為**真正的原因**（例如 400/401 之類的請求或認證錯誤、回應解碼失敗、輸出上限截斷）而失敗時，我希望 tellme **立刻**回報，而不是先等 4 秒、再重送一次註定失敗的請求。

**為何為此優先級**: 它是 US1 的安全邊界 —— 沒有它，重試會把一個快速、明確的錯誤變成一段無謂的等待，並對遠端造成無效負載；它是與 US1 同一個機制的必要另一半。

**獨立驗證方式**: 以 in-process fake provider 腳本化一個 **400**（非可重試）回應，斷言 fake 被呼叫**恰好一次**且立即以片語 + exit 6 失敗（零重試、零等待）；以相同方式腳本化輸出上限截斷，斷言不重試。

**驗收情境**:

1. **Given** a configured provider whose endpoint answers a **non-retryable** failure (e.g. HTTP 400), **When** the operator runs a prompt, **Then** the endpoint is called **exactly once** and the run fails immediately with the frozen phrase + exit **6** (zero retries, zero waits).
2. **Given** a configured provider whose reply is a successful-but-**truncated** reply (the round-030 output-cap guard), **When** the operator runs a prompt, **Then** the failure is surfaced immediately — it is **not** retried.

**功能需求（FR）**:

- **FR-005**: 系統 MUST NOT 重試**非可重試**的失敗：4xx（含 400/401/403/404/422…）、內容過濾拒絕、回應解碼／解析錯誤，以及輸出上限截斷守衛（round 030）。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`make verify` 的 `verify-no-network` MUST NOT 被破壞；MUST NOT 新增 live-network 閘。

---

### 邊界情況

- 當父 context（SIGINT/SIGTERM）在**請求期間或等待期間**被取消時，系統 MUST 立即停止並走現行的失敗路徑（不新增片語；MUST NOT 在取消後繼續等待或重試）。
- 當某一次重試**成功**時，系統 MUST 正常完成該 turn（正常持久化、正常 `stdout` 輸出），且 `calls` 計為 **1**。
- 當遠端在**回應主體中途**斷線（`io.ReadAll` 得到 `EOF`/`unexpected EOF`）時，系統 MUST 將其視為可重試的傳輸失敗。
- 當失敗為 **429** 時，系統 MUST 依 1 s／3 s 排程重試（與其他可重試類別相同）；當失敗為 **一般 4xx** 時 MUST NOT 重試。
- 當「等待」與取消同時發生時，取消 MUST 勝出（系統 MUST NOT 在取消後仍發出下一次嘗試）。
- 當 fake provider 於**第 3 次**嘗試才成功時，系統 MUST 成功且 MUST NOT 再發出第 4 次嘗試（上限為硬性邊界）。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 重試 MUST 於**單一**接縫涵蓋**兩個**提供者家族（OpenAI-compatible 與 Gemini/Vertex），MUST NOT 於各轉接器分別重複實作。
- **FR-007**: 一次被重試的呼叫 MUST 計為**一次** AI-endpoint call（`calls`）、MUST 只寫入**一筆** usage 記錄、MUST NOT 觸發第二個 turn frame，且失敗的嘗試 MUST NOT 寫入 session history。

#### 非功能需求

- **NFR-003**: 可重試性的判定 MUST 為**型別化**的（例如錯誤型別上的狀態欄位或可重試性查詢，或專屬 sentinel）；MUST NOT 以錯誤**字串**比對來判定可重試性。
- **NFR-004**: 重試相關的診斷文字 MUST 為控制碼安全（control-free）且 MUST NOT 洩漏憑證（與 round-039／round-056 的既有政策一致）。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`ProviderError`（`internal/domain/llm`）**: 單一的提供者／傳輸失敗型別 —— 本輪可能擴充其**可分類性**（狀態欄位／可重試性查詢），但 MUST 維持其對 CLI 的對外語意不變（仍是「片語 + exit 6」的來源）。
- **重試裝飾器（gateway wrapper）**: 在 CLI 的 `runTurn` 接縫包住 `llm.Gateway` 的一層（類似既有的 `withUnpairedDiagnostic`）—— 本輪新增；重試次數與排程為其內部契約。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 在 hermetic E2E 中，一個「先傳輸失敗、再成功」的腳本化 fake provider 被呼叫**恰好 2 次**，且 turn 正常完成（答案輸出、`stdout` 正確）。
- **SC-002**: 一個「失敗、失敗、再成功」的腳本被呼叫**恰好 3 次**（在 1 s→3 s 排程之後成功），且**零**第 4 次嘗試。
- **SC-003**: 一個「持續可重試失敗」的腳本被呼叫**恰好 3 次**，之後以 `the provider request failed` + exit **6** 失敗，`stdout` 為空且位元組精確。
- **SC-004**: 一個非可重試失敗（如 HTTP 400）被呼叫**恰好 1 次**並**立即**以片語 + exit **6** 失敗（零重試、零等待）；輸出上限截斷亦不重試。
- **SC-005**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），且 `stdout` 既有契約位元組不變。

## 假設

- **A1**: 本輪為**操作者請求**（無 anchor issue；比照 rounds 073/074/075 的前例）。
- **A2**: 可重試性判定（傳輸 + 429/5xx）與排程（1 s → 3 s，最多兩次重試）為**操作者在 session 中鎖定**的決策；其餘（機制／接縫／型別化分類／測試接縫／記錄形狀）由 `/axb-technical-research` 拍板。
- **A3**: MCP 路徑**不在本輪範圍**；`-b`/`--retry`（使用者主動重跑）為既有的 settled exclusion，與本輪的自動傳輸重試無關。
- **A4**: 不新增類別片語、不新增結束碼；最終失敗面與今天逐位元組相同。
- **A5**: 重試的等待為**production** 行為（非測試 `time.Sleep`），故 `verify-no-test-sleep` 不受影響；E2E 是否需 fast-retry 接縫（注入 sleeper／clock port）為研究決策。
