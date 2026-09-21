# Feature Specification: Make the callable MCP wire name discoverable in the offered declaration (round 077)

**Feature Branch**: `077-mcp-tool-name-presentation`

**Created**: 2026-09-22

**Status**: Draft (specified; clarify **not escalated — 0 questions**; **research-gated — a NOOP verdict is an acceptable outcome**; residual technical choices → `/axb-technical-research`; see §Clarify strategy)

**Anchor**: **[#155](https://github.com/gosharplite/tellme/issues/155)** — *"MCP tool presentation doesn't surface the namespaced wire name — fallback description names the bare tool; `mcp_` is unadvertised (research-gated, may be NOOP)"*. **The round's DoD is closing #155** — by shipping the presentation fix **or** by recording a research-verified NOOP and closing it as settled.

**Related (delivered, the loop-side half)**: **[#154](https://github.com/gosharplite/tellme/issues/154)** — the *loop-side* unknown-name consequence was fixed by **round 076 / ADR 0048** (an unknown name is now a recoverable, per-turn-bounded fold-back whose result **names the available wire names** — a partial mitigation of #155's discoverability symptom). #155 is the *presentation-side* hypothesis for *why* a weak model chose the bare name in the first place.

**Input (issue #155, grounded on `dev` @ `86d0deb`)**: a butler session on the `misc` repo (`deepseek-flash`) emitted a call to `get_me` while the callable wire name is `mcp_github_get_me`. Two upstream sites in `internal/infrastructure/mcp`:
1. `tool.go:53-56` — the **synthesized fallback description names the BARE tool** when the server ships no description: `"MCP tool " + def.Name + " from server " + server` → `MCP tool get_me from server github`.
2. `tool.go:97` (`Parameters()`) / `tool.go:60` (`name`) — the offered declaration's **name** is the namespaced one, but the **description** is not enriched with it; **no** system-prompt change advertises the `mcp_<server>_<tool>` convention (**ADR 0025 D4**), so namespacing is an **unadvertised wire convention**.

**Behaviour intent**: **MODIFY (MCP tool presentation)** — the callable wire name becomes **positively discoverable** to the model in the offered declaration, and any tellme-synthesized text **names the callable wire name**, not the bare upstream name — **or**, if research verifies the gap is not in play, a recorded **NOOP**. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **The gap is a *discoverability* problem, not a proven cause.** Issue #155 is explicit: the fallback fires **only** when the server's description is empty (GitHub MCP tools usually ship descriptions, so this path may not have triggered the observed run); a well-behaved model reads `function.name` (which *is* the namespaced name) rather than the description. So the round **MUST verify first** whether the fallback is in play and whether a presentation change is warranted — **a NOOP verdict is an acceptable outcome**.
- **The server definition is NEVER mutated (I-1; ADR 0025 D1).** The remote server's **name / description / input schema** are relayed unchanged. The offered declaration is **tellme's own** envelope (round 056 / ADR 0025): a **tellme-authored** note is added *outside* the server's own text — the issue's Option 2 ("still not touching the server text") — never a rewrite of the server's words.
- **The namespaced wire name is the contract (I-2).** `internal/infrastructure/mcp/naming.go:14-16` — `mcpNamePrefix = "mcp_"`; `NamespacedName(server, tool)` → `mcp_<server>_<tool>` (64-byte hash-truncated). The offered declaration's `name` is that namespaced name (`tool.go:60`); the bare name is **never** offered. This round does **not** change the naming rule.
- **The offered declaration is assembled once, in-domain.** `internal/domain/agent/result.go:75` — `llm.ToolDef{Name: t.Name(), Description: t.Description(), Parameters: t.Parameters()}`. The description tellme offers is `Tool.Description()` — the server's text, else the synthesized fallback. Any change to *what the model reads* lands here (or in `mcp.Tool`'s description).
- **The `description` field is schema-projection-safe.** `description` is an **accepted** keyword on the closed Vertex/Gemini wire (`supportedSchemaKeys`, round 061 / ADR 0031); the vendor-extension floor (`NormalizeMCPSchema`) drops `x-…`/`$schema` — neither touches the tool-level `description`. A description-only change is **wire-shape-neutral** (no schema diff).
- **No system-prompt / persona change (ADR 0025 D4).** The ask stays **declaration-carried**, exactly like the native `reason` mechanism. A prompt/persona change is **out of scope**.
- **#155's Option 3 is already delivered.** An "available-tools listing in a recoverable unknown-name result" now exists (round 076): the unknown-name fold-back result reads `error: no tool named "<name>"; available tools: <n1>, <n2>, …` from `Registry.Tools()`. So the remaining #155 surface is **Options 1 and 2** (the fallback text; a namespacing note on the offered declaration).

---

## Grounded in the current system *(measured 2026-09-22, `dev` @ `ab8fab7`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/mcp/naming.go:14-16,39-55` | `mcpNamePrefix = "mcp_"`; `NamespacedName(server, tool)` → `mcp_<server>_<tool>` (deterministic, 64-byte hash-truncated). |
| `internal/infrastructure/mcp/tool.go:53-56` | `desc := def.Description; if strings.TrimSpace(desc) == "" { desc = "MCP tool " + def.Name + " from server " + server }` — the **fallback names the bare tool**. |
| `internal/infrastructure/mcp/tool.go:60` | `name: NamespacedName(server, def.Name)` — the offered declaration's **name** is the namespaced wire name. |
| `internal/infrastructure/mcp/tool.go:76-78` | `func (t *Tool) Name() string { return t.name }` · `Description()` returns the server text (or fallback). |
| `internal/domain/agent/result.go:68-77` | `ToolDefs(reg)` → `[]llm.ToolDef{Name, Description, Parameters}` in **registration order** — the one projection the model sees. |
| `internal/infrastructure/mcp/schema.go` | the vendor-extension floor + `required ⊆ properties`; **no** tool-level description handling. |
| `internal/infrastructure/llm/gemini/schema.go` | the closed-wire projection: `description` IS an accepted keyword; the projection never touches the tool-level `description`. |
| Observed evidence | butler/`deepseek-flash` on `misc`: emitted `get_me` (callable `mcp_github_get_me`) — the presentation-side hypothesis. |
| Existing falsifiable seam | the hermetic fake MCP server (`internal/infrastructure/mcp/mcptest/`, round 032) + the fake-provider E2E (`using-tools-from-a-remote-mcp-server.feature`) can observe the **offered declaration bytes** a fake provider records. |

---

## Design (shape is largely a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **Verify first, then decide** — confirm whether the fallback is in play (an empty server description) **and** whether the presentation gap warrants a change; a **NOOP verdict is acceptable** and closes #155 as settled. | **locked (issue #155: "research-gated, may be NOOP")** |
| **S-2** | **The mechanism** — (a) fix the synthesized fallback so it names the callable wire name; (b) add a **tellme-authored namespacing note** to the offered declaration (a description preamble); (c) both; (d) **NOOP**. | **research decision** (D-x) |
| **S-3** | The **exact wording / placement** of any note (e.g. a `[tellme: call this tool as "mcp_github_get_me"]` preamble vs. a suffix), and whether it applies to **every** MCP tool or only the fallback case. | **research decision** (D-x) |
| **S-4** | Whether the fix touches `mcp.Tool.Description()` (the adapter) or the `ToolDefs` projection, and how it interacts with the schema-projection/normalizer surfaces (must stay **wire-shape-neutral**). | **research decision** (D-x) |
| **S-5** | The **ADR + truth-row** shape (the MCP offered-declaration row; a new ADR vs. an amendment to ADR 0025/0031), and the domain-model call (ADR 0041). | **research decision** (D-x); truth written by the owner skills |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The server's own definition is never mutated.** The remote server's `name` / `description` / `input schema` are relayed unchanged (ADR 0025 D1); any tellme-authored note is **added outside** the server's text, never a rewrite of it.
- **I-2 — The namespaced wire name is the contract.** The offered declaration's `name` stays `mcp_<server>_<tool>`; the bare upstream name is **never** offered. The naming rule is unchanged.
- **I-3 — Wire-shape-neutral.** A description-only change MUST NOT alter the offered **schema** bytes on any family; the vendor-extension floor and the closed-wire projection are untouched (no new schema keyword; no byte-identity regression on the text path).
- **I-4 — No system-prompt / persona change** (ADR 0025 D4). The ask stays declaration-carried.
- **I-5 — Family-local, stdlib-only, POSIX-only, hermetic.** No new dependency; the behaviour is provable through the existing hermetic fake MCP server + the fake-provider E2E (no live network in the gate).
- **I-6 — A NOOP is a first-class outcome.** If research verifies the gap is not in play, the round records the verification and closes #155 **without** a product change; no speculative churn.

---

## 使用者情境與測試 *(必填)*

### 使用者故事 1 - MCP 工具的可呼叫線路名稱可被正面發現 (Priority: P1)

作為一個使用 tellme 的操作者，當 tellme 向模型提供一個 MCP 工具時，我希望模型能在宣告本身中**正面發現**該工具的可呼叫線路名稱（`mcp_<server>_<tool>`），使得較弱的模型不會因為只記得裸上游名稱（例如 `get_me`）而呼叫一個註冊表中不存在的名稱。

**為何為此優先級**: 這正是 #155 的呈現面假設 —— 裸名被誤用的根源；若模型能從宣告中讀到正確的可呼叫名稱，第一次就呼叫正確，無須依賴回合 076 的可回復摺回來收拾。

**獨立驗證方式**: 以 hermetic fake MCP server 提供一個工具，讓 fake provider 記錄 tellme 實際提供的宣告；斷言該宣告讓可呼叫線路名稱可被發現（且伺服器本身的文字未被竄改）。

**驗收情境**:

1. **Given** a remote MCP server advertises a tool, **When** tellme offers it to the model, **Then** the **callable wire name** `mcp_<server>_<tool>` is **positively discoverable** in the offered declaration (not merely implied by an unadvertised convention).
2. **Given** the tool's server-authored description, **When** tellme offers the declaration, **Then** the server's own words are **relayed unchanged** (any tellme note is added outside them).
3. **Given** the offered declaration, **When** it reaches the provider wire, **Then** the offered **schema** bytes are unchanged (the change is description-only / wire-shape-neutral).

**功能需求（FR）**:

- **FR-001**: 系統 MUST 使每個 MCP 工具的可呼叫線路名稱（`mcp_<server>_<tool>`）在其提供給模型的宣告中**正面可被發現**（若研究判定需要變更）。
- **FR-002**: 系統 MUST NOT 竄改遠端伺服器自己的定義（`name`/`description`/`input schema`）；任何 tellme 撰寫的說明 MUST 附加於伺服器文字之外（ADR 0025 D1）。
- **FR-003**: 系統 MUST 維持提供宣告的 `name` 為命名空間化的線路名稱，且 MUST NOT 提供裸上游名稱。

### 使用者故事 2 - tellme 合成文字指名可呼叫線路名稱（無描述時）(Priority: P2)

作為一個操作者，當一個 MCP 伺服器未提供工具描述時，我希望 tellme 合成的備援描述指名**可呼叫的線路名稱**（例如 `mcp_github_get_me`），而不是裸上游名稱（`get_me`），使備援文字本身不再誤導模型。

**為何為此優先級**: 這是 #155 的地面真值 —— 一個明確、可證偽、低風險的文字修正；它只在伺服器描述為空時觸發，是 US1 的一個具體子案。

**獨立驗證方式**: 以 hermetic fake MCP server 提供一個**無描述**的工具，斷言備援描述指名可呼叫的線路名稱（且不指名裸名稱為可呼叫名稱）。

**驗收情境**:

1. **Given** a remote MCP server advertises a tool with an **empty** description, **When** tellme synthesizes the fallback, **Then** the fallback names the **callable wire name**, not the bare upstream name.
2. **Given** a remote MCP server advertises a tool **with** a description, **When** tellme offers it, **Then** the fallback is not used and the server's text is relayed unchanged.

**功能需求（FR）**:

- **FR-004**: 系統 MUST 讓「伺服器無描述」時的合成文字指名可呼叫的線路名稱（`mcp_<server>_<tool>`）。
- **FR-005**: 系統 MUST NOT 在提供給模型的宣告中，把裸上游名稱呈現為可呼叫名稱。

**非功能需求（NFR）**:

- **NFR-001**: 本輪變更 MUST 為描述層級、與線路 schema 形狀無關；提供宣告的 schema 位元組 MUST NOT 改變。

---

### 邊界情況

- 當伺服器描述為空時，合成文字 MUST 指名可呼叫線路名稱，且 MUST 仍然標示其來源伺服器。
- 當伺服器描述**非**空時，系統 MUST 原樣轉述伺服器文字；若加入 tellme 說明，MUST 附加於伺服器文字之外、不改寫其內容。
- 當可呼叫線路名稱因 64 位元組上限而被截斷（hash 後綴）時，任何指名的名稱 MUST 與實際提供的 `name` 逐位元組一致。
- 當所選 provider 屬於封閉線路家族（Vertex/Gemini）時，描述層級的變更 MUST NOT 引入任何未支援的 schema 關鍵字（維持投影結果不變）。
- 當研究判定呈現缺口**不在**起作用時（NOOP），系統 MUST NOT 產生投機性變更；#155 以「研究已驗證、維持現狀」結案。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-006**: 若研究判定需要變更，系統 MUST 以 hermetic 方式（fake MCP server + fake provider）可證偽地見證「可呼叫線路名稱可被發現」；若研究判定 NOOP，系統 MUST 記錄驗證所依據的可觀察事實（伺服器描述是否為空、宣告 `name` 已為線路名稱等）。

#### 非功能需求

- **NFR-002**: 變更 MUST 侷限於 MCP 工具的呈現（`internal/infrastructure/mcp` 的 `Tool` 描述，及／或 `internal/domain/agent.ToolDefs` 的描述投影）；MCP 名稱規則、schema 正規化、封閉線路投影與 reason 信封皆 MUST 維持不變。
- **NFR-003**: 修正 MUST 維持 stdlib-only、POSIX-only、hermetic（不新增依賴；`make verify` 的 `verify-no-network` 不得被破壞；不加入 live-network 閘）。
- **NFR-004**: 任何 tellme 撰寫的文字 MUST 為控制碼安全（control-free）且不洩漏憑證（與 round-039 sanitize 家族一致）。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`MCPToolDefinition`**: 一個遠端伺服器提供的工具（`Name`、`Description`、`InputSchema`）—— **形狀不變**；本輪不修改伺服器自身的值。
- **`Tool`（MCP adapter）的提供宣告**: 對模型提供時使用的 `Name()`/`Description()`/`Parameters()` —— 本輪可能修改 **`Description()`** 的內容（合成備援及／或附加的 tellme 說明）；`Name()`/`Parameters()` 不變。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: 若實作，則在 hermetic E2E 中，**100%** 的 MCP 工具提供宣告讓可呼叫線路名稱可被發現，且 **0** 個提供宣告把裸上游名稱呈現為可呼叫名稱。
- **SC-002**: 一個**無描述**的 MCP 工具，其合成描述**逐位元組**指名實際提供的命名空間線路名稱（hash 截斷者亦然）。
- **SC-003**: `make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變（零新依賴），且提供宣告的 **schema 位元組不變**（零 schema 差異）；伺服器自身的定義不變。
- **SC-004**: 若研究判定 **NOOP**，該判定 MUST 由可觀察事實支撐（記錄於 `research.md` 與相關 ADR/truth），且 #155 以「研究已驗證」結案 —— 不得產生投機性程式碼變更。

## 假設

- **A1**: #155 為本輪唯一 anchor（DoD = 關閉 #155，可能以 NOOP 結案）；#154 的回合 076 修復（可回復摺回）已交付，且其結果已列出可用工具名稱。
- **A2**: 提供宣告的 `name` 已是命名空間化線路名稱（`mcp_<server>_<tool>`）；本輪不改變命名規則。
- **A3**: 「可被發現」的機制（合成備援修正 vs. 提供宣告的命名空間說明 vs. 兩者 vs. NOOP）由 `/axb-technical-research` 拍板。
- **A4**: 任何 tellme 撰寫的說明為宣告承載，無系統提示／人設變更（ADR 0025 D4）。
- **A5**: 變更若發生，為描述層級、schema 形狀無關；不引入任何 schema 關鍵字。
