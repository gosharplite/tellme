# Feature Specification: Round-scoped media placement — a multi-media tool round delivers its images once, after all tool results (round 083)

**Feature Branch**: `083-round-scoped-media-placement`

**Created**: 2026-09-23

**Status**: Draft (specified; clarify **not escalated — 0 questions**; the theme is anchored by issue [#167](https://github.com/gosharplite/tellme/issues/167) and the operator's directive *"Open a new round, the goal is close #167"*, and the issue's *Root cause / Proposed fix / Witnesses* are routed to `/axb-technical-research`)

**Input (anchor issue [#167](https://github.com/gosharplite/tellme/issues/167), 2026-09-23; operator directive)**: *"A turn that makes **more than one media-producing tool call in a single round** (two or three `read_image` calls in one model response) is fundamentally broken on the OpenAI-compatible family — the loop interleaves each image between tool results, so the next request of the turn is rejected 400."* — so a multi-image round must complete: all `tool` results answered **contiguously**, then the round's image(s) delivered to the model.

**Observed context (`dev` @ `426a93f`, round 082 era)**: `internal/agent/agentloop.go` appends the assistant `tool_calls` message once (`:143`), then **inside the same per-call loop** appends a `tool` result (`:211`) **and** — when that call attached media — a `user` media message (`:216`). The OpenAI-compatible adapter (`internal/infrastructure/llm/openai/client.go` `requestBody`) relays each message **verbatim**, preserving the loop's order. So a round with N ≥ 2 media-producing calls produces the wire order `assistant(tool_calls), tool(r1), user(img1), tool(r2), user(img2), …` — an interleaved non-`tool` message that violates the provider's contiguity requirement (`An assistant message with 'tool_calls' must be followed by tool messages responding to each 'tool_call_id' …`), rejecting the next request of the turn (exit **6**; the round-080 failure path then persists the completed steps). The **Gemini/Vertex** adapter (`internal/infrastructure/llm/gemini/client.go` `roundBuilder.consume`/`flush`) buffers media and emits it after the batched function-response turn — accidentally correct by construction. The **single-call** case (`assistant, tool, user(img)`) is a complete block and works.

**Behaviour intent**: **MODIFY (the agent tool loop's media placement)** — accumulate a round's media across its calls and fold it **once**, **after** the per-call loop, so every `tool` message of the round is contiguous. Single-call output stays **byte-identical**. **Clarifies the round-062 ADR 0032 D7:** its *body* is already round-scoped ("**one** `user` message … after the round's `tool`-result message(s)"), but its *heading* ("a `user` message after the tool result") was read **per-call** in the pre-083 implementation — this round makes the implementation match D7's body. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **Contiguity is the wire's mandate (I-1).** A round's `tool` messages MUST be contiguous: **no** `user`/`assistant`/media message may sit between the first and last `tool` result of a round. This is the invariant the OpenAI-compatible family enforces with the 400; it is the invariant the fix restores.
- **Round-scoped media = one message (I-2).** A round's **N** media parts are delivered in **one** `user` message, placed **after** the round's results (never per-call). This is the round-scoped form this round adopts (clarifying ADR 0032 D7 — its body read round-scoped; the pre-083 implementation did not).
- **Single-call bytes are frozen (I-3).** For N = 1 the emitted order stays exactly `assistant(tool_calls), tool(result), user(media)` — **byte-identical** to today, so the shipped single-image path is unchanged.
- **Gemini placement unchanged (I-4).** The Gemini/Vertex adapter still emits the round's media **after** the batched function-response turn. *Recorded consequence:* because the loop now hands it **one** media message per round, the Gemini multi-media path emits **one** media turn (N `inlineData` parts) instead of N turns — a benign shape change decided/recorded by `/axb-technical-research`, never a failure.
- **No storage / presentation change (I-5).** Media is never persisted (the `history.Step` stores the tool's **text** result); a resumed session replays contiguous `assistant(tool_calls), tool(result)…` steps and then delivers no media (the images are in-flight only). `history.jsonl` schema, `history.Store`, `-b`/`--back`, the offline readers, the interactive prompt, and the live chrome are **unchanged**. The exit-code set stays **ten**; the frozen class-phrase vocabulary is unchanged.
- **Deterministic, hermetic, stdlib-only, POSIX-only (I-6).** No new dependency; the E2E arranges a scripted multi-media round and asserts the recorded wire/bytes with no pty or live network.
- **The witness must redden (I-7).** The round adds the **missing tripwire**: a **unit** pin over the loop's emitted message order for a 3-media round, and an **E2E** contiguity witness (the wire never carries a non-`tool` message between a round's `tool` results). Both MUST redden under the **pre-fix** ordering (mutation of the media placement) and pass after. (This is the `aixbdd-tmg#15` / `GAPS.md` "claim-without-a-tripwire" class — the existing `toolExchangeChronologyOK` asserts only *precedence*, never *contiguity*, and its fixtures never use a media tool.)

---

## Grounded in the current system *(measured 2026-09-23, `dev` @ `426a93f`)*

| Site | Current shape |
| --- | --- |
| `internal/agent/agentloop.go` (`Run`) | `turn = append(turn, llm.Message{Role: "assistant", ToolCalls: resp.ToolCalls})` (`:143`); per call: `turn = append(turn, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})` (`:211`), then `if len(media) > 0 { turn = append(turn, llm.Message{Role: "user", Media: media}) }` (`:216`) — **interleaved per call**. |
| `internal/agent/agentloop.go` (`BuildMessages`) | the **replay**-path projection: `user(prompt)`, then per step `assistant(tool_call)`, `tool(result)`, then `assistant(answer)` — already contiguous (no media; media is not persisted). |
| `internal/infrastructure/llm/openai/client.go` (`requestBody`/`messageContent`) | relays each `llm.Message` verbatim (`role` + `content` + `tool_call_id`); `messageContent` renders a media-bearing message as a content **array** (optional text part + one inline base64 `image_url` block per media part). |
| `internal/infrastructure/llm/gemini/client.go` (`roundBuilder.consume`/`flush`) | buffers a media-bearing message into `mediaTurns` and emits it in `flush()` **after** the batched function-response turn. |
| `internal/domain/llm` (`Message`) | `Message{Role, Content, ToolCalls, ToolCallID, Media []tools.MediaPart}` — the loop's wire message shape; `Media` is in-flight only. |
| `internal/domain/tools` (`MediaTool`/`MediaPart`) | `read_image` returns media **in-band** via `ExecuteMedia` (round 070 / ADR 0040); the loop type-asserts `tools.MediaTool`. |
| `tests/e2e/steps/wire_tools.go:169` (`toolExchangeChronologyOK`) | asserts a `tool` message is **preceded** by an assistant-with-tool_calls and a user message — **no** contiguity assertion, and its fixtures never use a media tool. |
| `tests/e2e/fakeprovider` (`Reply.Tools`) | scripts ONE response carrying several tool calls (round 019) — the `<server>` seam the multi-image fixture uses. |
| `specs/truth/features/cli/chat/reading-a-local-image.feature` | the image Rules/Examples (rounds 062/063) — the owning feature for the media journey; no multi-call Rule today. |
| `specs/truth/features/cli/chat/calling-several-tools-in-one-round.feature` | the several-tools-in-one-round feature (rounds 065/066) — the natural sibling carrier for a media round. |
| `docs/decisions/0032-agent-image-vision.md` (D7) | the round-062 media-placement decision — its heading ("a `user` message after the tool result") read per-call, its body round-scoped; this round **clarifies** it (back-pointer added to its `Status` + D7). |
| `GAPS.md` / `aixbdd-tmg#15` | the "claim-without-a-tripwire" class record this round's missing witness is an instance of. |

---

## Design (locked by the issue/operator vs. decided by `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **L-1** | A round's media is accumulated and folded **once**, **after** the per-call loop, so every `tool` message of the round is **contiguous**. | **locked** (issue Proposed fix) |
| **L-2** | The single-call output (`assistant, tool, user(media)`) stays **byte-identical**; the Gemini placement (media after the batched function-response turn) is unchanged. | **locked** (issue Proposed fix / Single-call) |
| **L-3** | **Placement-only**: no `history.jsonl` schema, `history.Store`, `-b`/`--back`, prompt, chrome, or exit-code/phrase change. | **locked** (issue Acceptance / references) |
| **S-1** | **The exact emitted shape for N ≥ 2** — ONE `user` message carrying the round's N media parts (media-first, no text), appended after the round's results; vs. an alternative (e.g. one media message per call appended after the loop). | research decision (D-x); **proposed: one `user` message carrying the round's media in call order** (mirrors the issue's sketch and the multi-part content array) |
| **S-2** | **Which witness tiers** — a loop-tier unit pin over the emitted message order (3-media round) **plus** an E2E contiguity witness (extend `toolExchangeChronologyOK` / a new multi-image fixture). Whether the wire witness rides the `openai` family only or both. | research decision (D-x); **proposed: both tiers; wire witness on the OpenAI-compatible family (the broken one) + a Gemini multi-media companion** |
| **S-3** | **Which truth artifacts change** — a new **ADR** (clarifies ADR 0032 D7); the `techstack.md` *Agent tool loop* / *Image filesystem tool* / *Image content on the provider wire* rows; the `reading-a-local-image` (or `calling-several-tools-in-one-round`) feature + `dsl.md` rows; whether `docs/domain-model/**` is modelled (media placement is a wire detail — possibly **not modelled**, the ADR-0041 escape hatch). | research decision (D-x) |
| **S-4** | **Scope excludes** — media persistence / re-budget (the images still ride the next round's active-turn messages), image dedupe/downscale, a `--image` flag, a config toggle, and any change to `read_image`'s contract or the family-aware ceiling. | **locked** (issue scope) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Contiguity.** A round's `tool` messages are contiguous; no non-`tool` message sits between the first and last `tool` result.
- **I-2 — Round-scoped media.** A round's N media parts are delivered in ONE `user` message after the round's results.
- **I-3 — Single-call bytes frozen.** N = 1 is byte-identical to today.
- **I-4 — Gemini placement unchanged.** Media still follows the batched function-response turn (recorded cardinality consequence for N ≥ 2).
- **I-5 — Placement-only.** No storage/schema/`-b`/prompt/chrome change; exit-code set stays **ten**; frozen phrases unchanged.
- **I-6 — Deterministic, hermetic, stdlib-only, POSIX-only.**
- **I-7 — The witness reddens.** A unit pin + an E2E contiguity witness that fail under the pre-fix ordering.

---

## 使用者情境與測試 *(必填)*

### 使用者情境 1 - 一回合多張圖片的工具呼叫能在 OpenAI-compatible 家族完成 (Priority: P1)

作為一個操作者，當我在同一個回合讓模型讀取**多張**本機圖片（例如三個 `read_image` 呼叫在同一個 model response 中），我希望該回合能順利完成（provider 不回 400），因為所有 `tool` 結果彼此相鄰、回合的圖片在**所有**結果之後才送達模型。

**為何為此優先級**: 這是 issue 的核心缺陷與唯一交付面 —— 在既有的媒體路徑上修正**送出順序**，讓多媒體回合不再被 provider 拒絕（exit 6）；修復直接、風險受限、價值明確。

**獨立驗證方式**: 以 fake provider 佈置一個**單一 model response 攜帶 N（≥2）個 `read_image` 呼叫**的回合，執行 tellme，斷言：(a) 該回合完成（exit 0、印出最終答案）；(b) 記錄到的 wire 中，該回合的所有 `tool` 結果彼此相鄰、其後才是**一個**媒體 `user` 訊息。並以**修復前**的順序（把媒體訊息移回 per-call）驗證見證會變紅。

**驗收情境**:

1. **Given** a vision-capable OpenAI-compatible provider whose ONE endpoint response asks tellme to read three images AND then answers, **When** the operator starts tellme with a prompt asking to inspect them, **Then** the turn completes with the provider's answer and tellme exits successfully.
2. **Given** the same round, **When** the recorded request is inspected, **Then** every `tool` result of the round is immediately adjacent to another `tool` result or to the assistant `tool_calls` message — **no** `user`/media message sits between them — and the round's images ride **one** `user` message after the results.
3. **Given** a session already holding a completed multi-media round, **When** the operator resumes it (a later prompt), **Then** the replayed conversation is contiguous and the new turn completes (exit successfully).

**功能需求（FR）**:

- **FR-001**: 當一個 model round 產生 N（N ≥ 2）個媒體工具呼叫時，系統 MUST 讓該回合的**所有** `tool` 結果在 wire 上 **相鄰**（其間 MUST NOT 插入任何 `user`/媒體訊息）（I-1）。
- **FR-002**: 該回合的媒體 MUST 以**單一** `user` 訊息送達，且 MUST 置於該回合**所有** `tool` 結果**之後**（I-2）。
- **FR-003**: 多媒體回合 MUST 完成（provider 不回 400），tellme MUST 印出最終答案並以成功結束碼結束。
- **FR-004**: 已完成的回合其後在**續接（resumed）**session 中被重播時，重播的對話 MUST 保持相鄰且後續回合 MUST 能完成。

**非功能需求（NFR）**:

- **NFR-001**: 媒體累積與擺放 MUST 為決定性；MUST NOT 依賴 map 走訪或任何非決定性順序（媒體依**呼叫順序**）。

---

### 使用者情境 2 - 單圖片路徑與 Gemini 路徑維持不變 (Priority: P2)

作為一個操作者，當我在一個回合只讀取**一張**圖片時，我希望送出的位元與修正前**完全相同**；當我使用 **Gemini/Vertex** 家族時，我希望媒體仍如以往在批次化的 function-response turn **之後**送達。

**為何為此優先級**: 這是 US1 的相容性保證 —— 修正多媒體順序 MUST NOT 回歸單圖片（已出貨）路徑或改變 Gemini 家族的擺放；它在 US1 之上（US1 未成立前無從交付），故為 P2。

**獨立驗證方式**: (a) 以單一 `read_image` 呼叫的回合，斷言記錄到的 wire 位元與既有 `assistant, tool, user(media)` 形狀一致（既有單圖片 E2E 持續綠）。(b) 以 Gemini（Vertex 形狀）多媒體回合，斷言 function-response turn 仍為批次且媒體其後送達。

**驗收情境**:

1. **Given** a vision-capable provider whose endpoint asks tellme to read exactly one image and then answers, **When** the operator runs the prompt, **Then** the recorded request still carries the image on a single `user` message immediately after the `tool` result (the shipped shape).
2. **Given** a Gemini/Vertex vision provider whose ONE response asks tellme to read two images and then answers, **When** the operator runs the prompt, **Then** the request carries the round's tool results in ONE batched function-response turn and the media after it, and the turn completes.

**功能需求（FR）**:

- **FR-005**: 當一個回合僅有**單一**媒體工具呼叫時，其送出的訊息順序 MUST 維持 `assistant(tool_calls), tool(result), user(media)` 的既有位元（I-3）。
- **FR-006**: Gemini/Vertex 家族的媒體擺放 MUST 維持於批次化 function-response turn **之後**（I-4）。

**非功能需求（NFR）**:

- **NFR-002**: 本輪變更 MUST 維持 stdlib-only、POSIX-only、hermetic —— 不新增依賴；`verify-no-network` MUST NOT 被破壞。

---

### 邊界情況

- 當一個回合的媒體工具呼叫數為 **0** 時，系統 MUST 完全不改變其 wire 形狀（無媒體訊息）。
- 當一個回合的媒體工具呼叫數為 **1** 時，系統 MUST 產生與修正前**位元相同**的順序（FR-005）。
- 當一個回合的部分呼叫產生媒體、部分不產生時，系統 MUST 只把**產生的**媒體併入那**一個**回合媒體訊息，且 MUST NOT 影響其餘 `tool` 結果的相鄰性。
- 當回合的媒體訊息在**重播（replay）**路徑出現時，系統 MUST NOT 重播媒體（媒體不落地 —— `history.Step` 只存文字結果），且重播的對話 MUST 保持相鄰。
- 當 provider 回傳 400 或其他錯誤時，系統 MUST 沿用既有 provider 片語（exit 6）回報（本輪 MUST NOT 改變錯誤呈現）；修復後多媒體回合 MUST NOT 觸發該 400。
- 當訊息帶有非空 `Content` **且**媒體時（未來呼叫端），系統 MUST 依 S-1 的決議處理（媒體訊息的前置文字），MUST NOT 產生兩個媒體訊息。

## 需求 *(必填)*

### 全域需求

#### 功能需求

- **FR-007**: 媒體的累積與落地 MUST 由 `internal/agent` 迴圈單一擁有（在 per-call 迴圈之外折疊一次）；MUST NOT 由 provider 配接器各自重排（配接器維持既有職責）（S-1）。
- **FR-008**: 本輪 MUST NOT 改變 `history.jsonl` schema、`history.Store`、`-b`/`--back`、互動提示、即時 turn chrome、`read_image` 契約與 family-aware 上限（I-5）。

#### 非功能需求

- **NFR-003**: 結束碼集合 MUST 維持十個；provider 片語詞彙 MUST 維持凍結；MUST NOT 因本輪新增任何片語或結束碼。
- **NFR-004**: 本輪 MUST 提供使既有缺陷可被否證的見證（loop-tier 單元 pin + E2E 相鄰性見證），且在**修復前**順序下 MUST 變紅（I-7）。

### 關鍵實體 *(若功能涉及資料，必填)*

- **`llm.Message`（`internal/domain/llm`）**: 既有形狀 —— **不變**；本輪只改變**何時/如何**由迴圈附加媒體訊息（`Media` 仍為 in-flight）。
- **回合媒體訊息（round media message）**: 一個 `user` 訊息，`Media` = 該回合所有呼叫的 `[]tools.MediaPart`（依呼叫順序），置於回合所有 `tool` 結果之後。
- **`history.Step`（`internal/domain/history`）**: 既有形狀 —— **不變**；只存工具**文字**結果（媒體不落地）。

## 成功標準 *(必填)*

### 可量測成果

- **SC-001**: hermetic E2E 中，一個 model round 攜帶 **3** 個 `read_image` 呼叫時，回合完成（exit 0、印出答案），且**未**觸發 provider 400。
- **SC-002**: 該回合記錄到的 wire 中，**沒有**任何非-`tool` 訊息夾在 `tool` 結果之間；回合的 3 個媒體區塊位於**一個** `user` 訊息、在結果之後。
- **SC-003**: 單圖片回合（既有 E2E）位元順序**不變**（`assistant, tool, user(media)`）；`reading-a-local-image.feature` 既有 Examples 全綠。
- **SC-004**: Gemini/Vertex 多媒體回合的 function-response turn 仍為**批次**，媒體其後送達（**載體 = `TestRequestBody_RoundScopedMedia_OneTurnTwoParts`** — 迴圈交一個含兩 parts 的媒體訊息 ⇒ 一個含兩個 `inlineData` parts 的媒體 turn），且回合完成。
- **SC-005**: 新見證在**修復前**的媒體擺放下**變紅**（mutation 復現後回退）；`make verify` 通過、E2E 全綠、`go.mod`/`go.sum` 不變、`TestExitCodesMatchPinnedContract` 綠。

## 假設

- **A1**: 本輪 anchor = issue [#167](https://github.com/gosharplite/tellme/issues/167)；**DoD = 關閉它**。
- **A2**: 核心行為 —— 回合內 `tool` 結果相鄰、媒體以**單一**訊息置於結果之後、單呼叫位元凍結、Gemini 擺放不變、placement-only —— 由 issue 鎖定；issue 的 *Proposed fix* 與 *Witness to add*（S-1…S-4：N≥2 的確切形狀、見證層級、truth/ADR 更新、範圍排除）由 `/axb-technical-research` 決議，故本輪 clarify **未升級（0 題）**。研究若改變正式驗收契約（例如 Gemini 多媒體 cardinality），MUST 回寫 truth 並於 `truth-delta.md` 記錄。
- **A3**: 媒體從未落地（`history.Step` 只存文字結果，rounds 062/070 契約），故續接 session 的重播對話不含媒體；此為既有行為，本輪不改。
- **A4**: 既有 `toolExchangeChronologyOK` 只斷言**先行**關係、未斷言**相鄰**關係，且其 fixture 從未使用媒體工具 —— 因此本輪的相鄰性缺陷目前**無任何測試可變紅**（GAPS.md / aixbdd-tmg#15 類）。本輪 MUST 補上該見證。
- **A5**: 預期新增一枚 ADR（澄清 ADR 0032 D7 的 round-scoped 讀法）、`techstack.md` 相關列更新、`reading-a-local-image`（或 `calling-several-tools-in-one-round`）feature + `dsl.md` rows 更新；`docs/domain-model/**` 是否更新由研究決議（媒体擺放為 wire 細節，可能 not modelled，ADR 0041 escape hatch）；`contracts/**` 與 `data/**` 預期 NOOP。
- **A6**: 本輪**不**實作媒体落地、payload 縮減/去重、`--image` 旗標或 config 開關（S-4）。
