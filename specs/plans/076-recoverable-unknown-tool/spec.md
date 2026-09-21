# Feature Specification: A recoverable, per-turn-bounded unknown tool name (round 076)

**Feature Branch**: `076-recoverable-unknown-tool`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; residual technical choices → `/axb-technical-research`; see §Clarify strategy)

**Anchor**: **[#154](https://github.com/gosharplite/tellme/issues/154)** — *"Agent loop aborts on an unknown tool name — an off-list name should fold back a recoverable 'no such tool' result (exit 7 today)"*. **The round's DoD is closing #154.**

**Related (out of scope)**: **[#155](https://github.com/gosharplite/tellme/issues/155)** — MCP tool **presentation** does not surface the `mcp_<server>_<tool>` wire name (a separate, research-gated presentation issue). This round may *shrinks* it if the fold-back message names the available tools, but #155's own fix is **not** in scope here.

**Input (issue #154, grounded live)**: a butler session (`b`, provider `deepseek-flash`) on the `misc` repo executing that repo's `SESSION-BOOTSTRAP.md` asked for the GitHub MCP tool by its **bare** upstream name `get_me` (the callable wire name is `mcp_github_get_me`); tellme aborted the whole run: `tellme: the tool request failed: tool "get_me" is not available` (**exit 7**).

**Behaviour intent**: **MODIFY (agent tool loop)** — an **unknown tool name** becomes a **recoverable fold-back** (the model receives a `tool`-role result and the loop continues), **bounded per turn** by a fixed cap; the frozen tool-error phrase + **exit 7** remain for the genuine incomplete cases. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The defect (grounded).** `internal/agent/agentloop.go:133-136` — `Registry.Lookup(tc.Name)`; a miss **returns terminally**:
  ```go
  tool, ok := a.Registry.Lookup(tc.Name)
  if !ok {
      return agentport.Result{Steps: steps, Calls: calls}, &agentport.ErrIncomplete{Reason: fmt.Sprintf("tool %q is not available", tc.Name)}
  }
  ```
  `emitToolError` (`internal/cli/cli.go:1392-1395`) renders the frozen phrase and returns `ToolError = 7` (`internal/cli/exitcode.go:28`; round-008 FR-010). One bad name kills the turn.
- **The existing recoverable convention (the pattern to reuse).** In the *same* loop: (a) a real tool that returns an error is folded back as a `tool`-role result `error: …` and the loop continues (`agentloop.go:171-176`); (b) `refuseReasonless` (round 056 / ADR 0025 D3) logs a result, appends a `tool` message with `ToolCallID`, and `continue`s — **never terminal**. An unknown name is the same shape.
- **The pairing invariant (round-065 / [#132](https://github.com/gosharplite/tellme/issues/132)).** The model's assistant message carries *N* calls and the loop appends one `tool` result per call; on the **Gemini/Vertex** wire each `functionCall` needs a matching `functionResponse` part in the same turn or the request 400s. The fold-back **MUST** append a result for **every** unknown call in the round — never skip one.
- **The bound is load-bearing.** tellme has **no repetition/runaway detection** (the reference's SHA-256 / tool-repetition guards were deliberately not re-created — verified absent in `internal/`). The only guard is the round bound `MaxLoops = MAX_TOOL_LOOP` (`internal/config/config.go:23`, `DefaultMaxToolLoop = 1000`; resolved `cli.go:596-603`), and **each round is one paid provider call**. An unbounded fold-back could spin up to ~1000 paid calls — a money/latency hang. The recoverable path **MUST** therefore be **bounded per turn**.
- **The contract to preserve.** The frozen phrase `tellme: the tool request failed: …` + **exit 7** stay for the genuine incomplete cases: the tool-loop **bound reached**, the per-turn **cap exhausted**, and the **`no tools are registered`** case. Round-008 FR-010 is **narrowed, not removed** (a likely ADR + truth-row change).
- **The reason gate ordering.** Today the registry lookup (`:133`) precedes the reason gate (`refuseReasonless`, `:143`). Whether an unknown-name call with a blank reason is classified as *unknown* or *reason-less* is a **research decision** (see §Design).

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `86d0deb`)*

| Site | Current shape |
| --- | --- |
| `internal/agent/agentloop.go:129-136` | `for _, tc := range resp.ToolCalls { … Registry.Lookup(tc.Name); if !ok { return ErrIncomplete } … }` — an unknown name is **terminal**. |
| `internal/agent/agentloop.go:143-146` | `refuseReasonless` — the **precedent** recoverable fold-back: `logResult` + append `tool` message + `continue`. |
| `internal/agent/agentloop.go:171-176` | a real tool's call-time error is folded back as `error: …` (recoverable). |
| `internal/agent/agentloop.go:183` | `steps = append(steps, history.Step{…})` — recorded **only for an executed** call. |
| `internal/cli/cli.go:1392-1395` | `emitToolError` — the frozen phrase + `ToolError`. |
| `internal/cli/exitcode.go:28` | `ToolError = 7`. |
| `internal/config/config.go:23,233-251` | `DefaultMaxToolLoop = 1000`; `EffectiveMaxToolLoop` (env `MAX_TOOL_LOOP` → file → default). |
| Observed evidence | butler/`deepseek-flash` on `misc`: `tellme: the tool request failed: tool "get_me" is not available` → exit 7. |

---

## Design (shape is largely a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | An unknown tool name **MUST** be a **recoverable fold-back** (the model gets a `tool`-role result and the loop continues), not a terminal abort. | **locked (operator / #154)** |
| **S-2** | The recoverable path **MUST be bounded per turn**: after **N** unknown-name fold-backs within one `Turn`, the loop stops folding back and returns the incomplete error (phrase + exit 7). `N = 3` proposed. | **locked cap; N value research-decided** |
| **S-3** | The fold-back message **content** — whether it names the unknown name, a bounded list of **available tool names**, and/or a near-miss suggestion. | **research decision** (D-x) |
| **S-4** | **Counter home** — a loop-local `run` variable versus a `TurnState` field; and that it resets per `Turn`. | **research decision** (D-x) |
| **S-5** | **Reason-gate ordering** — an unknown call with a blank reason: unknown-name first (today's lookup order) or reason-first. | **research decision** (D-x) |
| **S-6** | The **ADR + truth-row** shape: which `specs/truth/**` rows own the narrowed FR-010 contract and the fold-back rule. | **research decision** (D-x); truth written by the owner skills |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Every requested call gets a paired `tool` result.** The fold-back appends a result for **every** unknown call in the round (never skip one), so the Gemini/Vertex wire never 400s (round-065 / #132 class).
- **I-2 — Nothing executes and nothing is recorded for an unknown name.** No tool runs; no `history.Step`/`Signature` is appended (matching the reason-less refusal's recording); no tool-usage record is written (the call did not execute).
- **I-3 — The per-turn cap is enforced.** At most **N** unknown-name fold-backs per `Turn`; reaching it ends the run with the frozen phrase + exit 7. The counter resets each `Turn`.
- **I-4 — The genuine-incomplete contract is preserved.** The tool-loop **bound reached**, the per-turn **cap exhausted**, and **`no tools are registered`** still produce `ErrIncomplete` → the frozen phrase + exit 7 (round-008 FR-010 narrowed, not removed).
- **I-5 — Family-local & wire-stable.** The change is confined to the agent loop; the provider adapters/wire envelopes are untouched (no byte-identity claim changes); the tool registry, the reason gate (ADR 0025 D3), and the loop clamp are untouched.
- **I-6 — stdlib-only, POSIX-only, hermetic.** No new dependency; the behaviour is provable through the existing fake-provider/E2E seam.

---

## Clarify strategy

**Not escalated (0 questions).** No gap changes story slicing, requirement attribution, the main flow, or the acceptance/success criteria:

- the **goal** (close #154) and the **locked design** (recoverable fold-back + per-turn cap) are unambiguous from the operator's directive and the issue body;
- the residual choices — *the cap value N* (**S-2**), *the message content* (**S-3**), *the counter home* (**S-4**), *the reason-gate ordering* (**S-5**), *the ADR/truth-row shape* (**S-6**) — are **technical**, and per the round-074/075 precedent defer to `/axb-technical-research`;
- the available-tools listing (the only observable-output question) is an implementation detail of the *"tell the model no such tool"* intent the operator endorsed, not a story/acceptance change.

**Disclosed assumptions (not asked):**

- **A1** — **N = 3** unknown-name fold-backs per turn (research may adjust; the value is pinned by a test).
- **A2** — the frozen phrase + **exit 7** stay for the genuine incomplete cases (A2 is a contract statement, not a new behaviour).
- **A3** — the fold-back result **SHOULD** name the unknown tool and a **bounded** list of available tool names, to make the first retry self-correcting; exact text at research.
- **A4** — the `Tool`/`ToolCall` domain value shapes are **unchanged**; an unexecuted call records nothing new.
- **A5** — **#155** (MCP namespacing discoverability) is **out of scope**; this round does not change how MCP tools are named or declared.

**No `NEEDS CLARIFICATION` remains.**

---

## 使用者情境與測試 *(必填)*

### 使用者故事 1 - 未知工具名稱可回復，迴圈續行 (Priority: P1)

作為一個使用 tellme 的操作者，當模型請求一個不在註冊表中的工具名稱（打字錯誤、別名、或把 MCP 工具以裸名呼叫）時，我希望 tellme 不要中止整個回合，而是把「無此工具」的可回復結果回饋給模型並繼續，讓模型能自我修正並完成任務。

**為何為此優先級**: 這正是本次缺陷本身 —— 沒有它，任何一次工具命名失误都會讓整個回合以 exit 7 失敗；這是使用者直接感受到的價值，也是 #154 的 DoD。

**獨立驗證方式**: 以 fake provider 讓模型先請求一個不存在的工具名稱，再請求一個有效工具並給出最終答案；斷言整個回合未中止、被摺回未知名稱的結果、且最終 exit 0。

**驗收情境**:

1. **Given** the model requests a tool name that is **not** in the registry, **When** the loop dispatches the round, **Then** the loop does **not** abort — it folds back a recoverable `tool`-role result and continues.
2. **Given** the folded-back result follows an unknown-name call, **When** the model next requests a **valid** tool, **Then** that tool executes normally and the run can reach a final answer (exit 0).
3. **Given** an unknown-name call, **When** it is folded back, **Then** **no** tool executes for it and **no** history step is recorded for it.

**功能需求（FR）**:

- **FR-001**: 系統 MUST NOT 因未知工具名稱而中止回合；MUST 以可回復的 `tool`-role 結果回饋並續行。
- **FR-002**: 系統 MUST 為該回合中**每一個**未知呼叫附加一筆配對的 `tool` 結果（不得跳過任何一個），以維持 Gemini/Vertex 線路的配對不變式。
- **FR-003**: 系統 SHOULD 在回饋結果中指名未知的工具名稱，並列出**有界**的可用工具名稱清單，使模型能在第一次重試即自我修正（確切文字由研究決定）。

### 使用者故事 2 - 可回復重試以每回合上限收斂 (Priority: P2)

作為一個在意成本與延遲的操作者，我希望未知工具名稱的可回復重試在同一個回合內有固定上限，一旦模型固執地重複請求未知名稱，回合會在上限後以既有的工具錯誤片語與 exit 7 結束，而不是無限地消耗付費的 provider 呼叫。

**為何為此優先級**: 這是 US1 的安全護欄；tellme 沒有重複/失控偵測，唯一的界限是 `MAX_TOOL_LOOP`（預設 1000，每次為付費呼叫），因此若無上限，US1 會把「快速失敗」變成「昂貴空轉」。US1 成立後才能證明此上限的意義。

**獨立驗證方式**: 以 fake provider 在同一回合內連續請求同一個未知工具名稱 N 次；斷言摺回次數不超過 N，且到達上限後回合以工具錯誤片語與 exit 7 結束；另以「未達上限即出現有效工具或最終答案」的情境斷言正常結束（exit 0）。

**驗收情境**:

1. **Given** the model requests unknown names **N** times within one turn, **When** the cap is reached, **Then** the loop stops folding back and the run ends with the frozen tool-error phrase and **exit 7**.
2. **Given** fewer than **N** unknown-name fold-backs, **When** the model eventually requests a valid tool or returns a final answer, **Then** the run completes normally (**exit 0**).

**功能需求（FR）**:

- **FR-004**: 系統 MUST 將單一 `Turn` 內未知名稱的摺回次數上限設為 **N**（提議 N = 3；最終值由研究決定並以測試釘住）。
- **FR-005**: 系統 MUST 讓該計數器**逐回合重置**（而非逐 session），使其不跨回合累積。

**非功能需求（NFR）**:

- **NFR-001**: 上限 MUST 確保最壞情況付費 provider 呼叫數在單一回合內有界（不隨模型固執而退化為 `MAX_TOOL_LOOP` 級別的呼叫數）。

---

### 邊界情況

- 當一個回合同時含未知與有效呼叫時，系統 MUST 為每個呼叫各附加一筆配對結果，且有效呼叫 MUST 照常執行（維持配對與執行一致性）。
- 當未知名稱的呼叫同時沒有可渲染的 `reason` 時，系統 MUST 以**單一歸屬**判定其類別（未知名稱優先或 reason 優先），且該判定 MUST 由研究拍板並由測試釘住（不得同時套用兩條規則）。
- 當註冊表為空（`no tools are registered`）時，系統 MUST 維持既有行為（終止 + 片語 + exit 7），不受本輪影響。
- 當模型在同一回合內未達上限即修正（改用有效工具或給出答案）時，系統 MUST 正常結束（exit 0），且上限計數不得跨回合殘留。
- 當未知名稱的呼叫被摺回時，系統 MUST NOT 寫入任何 `history.Step`/`Signature`，且 MUST NOT 產生該呼叫的 tool-usage 記錄。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 系統 MUST 為「真實的未完成回合」保留既有的工具錯誤片語（`the tool request failed`）與 exit 7 —— 涵蓋工具迴圈**達到上限（bound reached）**、**每回合未知名稱上限耗盡（cap exhausted）**，以及 **`no tools are registered`** 三種情形（round-008 FR-010 為**收窄**，非移除）。

#### 非功能需求

- **NFR-002**: 變更 MUST 侷限於 agent loop；provider adapter、wire envelope、tool registry、reason gate（ADR 0025 D3）與迴圈 clamp 皆 MUST 維持不變（零線路變更）。
- **NFR-003**: 新增的每回合狀態 MUST NOT 跨回合洩漏；行為 MUST 具確定性，並可透過既有 fake-provider / E2E 縫隙驗證。
- **NFR-004**: 修正 MUST 維持 stdlib-only、POSIX-only、hermetic（不新增依賴）。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`ToolCall`**: 一次工具呼叫（`tool`、`reason`、`outcome`）—— **形狀不變**；一個未知名稱的呼叫**不執行**，因此**不**記錄新的 `outcome`、`history.Step` 或 usage 記錄。
- **`AgentLoop` 的每回合狀態**（實作層）: 一個新的**每回合**未知名稱計數（US2 的上限）—— 屬執行期狀態，非持久化實體；是否需反映於領域模型（`Turn`/`ToolCall`）由 `/axb-technical-research` 依 ADR 0041 決定。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 在 E2E 中，**100%** 的未知名稱呼叫（未達上限者）產生一筆配對的摺回結果並讓回合續行；**0** 次在未達上限前出現終止性中止。
- **SC-002**: 一個在同一回合內連續請求同一未知名稱 **N** 次的模型，其回合以工具錯誤片語 + **exit 7** 結束（且摺回次數恰為 N）。
- **SC-003**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），且 provider adapter 無變更（零線路差異）。

## 假設

- **A1**: 每回合未知名稱摺回上限 **N = 3**（研究可調整，最終以測試釘住）。
- **A2**: 既有工具錯誤片語 + exit 7 保留給真正的未完成情形（bound reached / cap exhausted / no tools registered）。
- **A3**: 摺回結果 SHOULD 指名未知工具並列出**有界**的可用工具名稱（確切文字由研究決定）。
- **A4**: `Tool` / `ToolCall` 的領域值形狀不變；未執行的呼叫不記錄任何新狀態。
- **A5**: [#155](https://github.com/gosharplite/tellme/issues/155)（MCP 命名呈現）不在本輪範圍內；本輪不改變 MCP 工具的名稱或宣告方式。
