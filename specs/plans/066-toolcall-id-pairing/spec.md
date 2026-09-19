# Feature Specification: Gemini/Vertex — pair a round's tool results to their calls by `ToolCallID` (round 066)

**Feature Branch**: `066-toolcall-id-pairing`

**Created**: 2026-09-20

**Status**: **Draft — `/axb-specify` output (round opened)**. Produced from **anchor issue [#134](https://github.com/gosharplite/tellme/issues/134)**, which homes **ADR 0035 §Forward RF-065-1** (the round-065 follow-up). No `specs/truth/**` file is written by this skill. Clarify: **not escalated** (see *Clarify strategy* — the goal and the fix are unambiguous; the remaining choices are technical, deferred to `/axb-technical-research`).

**Input (operator, 2026-09-20, this session)**:

> *"Open round 066-*, **the goal is to close issue #134**."*

**Anchor issue**: [#134](https://github.com/gosharplite/tellme/issues/134) — *"Gemini/Vertex: pair a round's tool results to calls by `ToolCallID` (ADR 0035 RF-065-1; FIFO fallback)"*. **The round's definition of done is closing #134.**

**Behaviour intent**: **MODIFY (the Gemini/Vertex `functionCall`/`functionResponse` serialization) — a *hardening / parity* round, NOT a defect fix.** Today the Gemini adapter emits `functionResponse{name, response}` and pairs each result to a call by a **positional FIFO of names** — correct only because the loop is **sequential**. The round makes the wire **id-linked** (`functionCall.id` and `functionResponse.id`) and binds each result to its call **by id** (order-independent), keeping the FIFO name match as a **fallback** for a result with no id (the replay path). **Observable user-visible behaviour is unchanged today** (a sequential loop already pairs correctly); the round is measured by the *request-body shape* and by **byte/shape preservation** on the untouched paths, not by a user-visible before/after.

---

## Grounded in the current system *(measured 2026-09-20, `dev` @ `babeff7`)*

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/llm/gemini/client.go` `buildContents` | `pending := make([]string, 0)` — a **FIFO of call names**; a `model` turn appends every `tc.Name`; each `tool` result pops `pending[0]` and emits `{"functionResponse":{"name":name,"response":{"content":…}}}` — **no `id`** on either the `functionCall` or the `functionResponse` part. At a round boundary any names still pending are dropped (the `M < N` case). |
| `internal/infrastructure/llm/gemini/client.go` `parseResponse` | Builds `llm.ToolCall{ID: fmt.Sprintf("call_%d", callIdx), …}` — a **synthetic, positional** id; the Vertex response's own `functionCall.id`, if any, is unread. |
| `internal/agent/agentloop.go` | The live path appends one `llm.Message{Role:"tool", Content:result, ToolCallID:tc.ID}` per call (`~:145/171`); the **replay** path synthesises a deterministic `call_step_<n>` id per replayed step (`~:413-426`). So **every producer sets a `ToolCallID`**. |
| `internal/infrastructure/llm/openai/client.go` | The OpenAI-compatible wire **already** emits `tool_call_id: m.ToolCallID` on `role:"tool"` messages (`:136-137`) — the id axis already exists for that family; the Gemini family is the outlier. |
| `internal/domain/llm/gateway.go` | `ToolCall{ID,…}` and `Message.ToolCallID` are **present and set**; the Gemini adapter ignores them today. |
| `internal/infrastructure/llm/gemini/client_image_test.go` + the round-065 unit pins | `TestRequestBody_MultiCallRound_BatchesFunctionResponses`, `…_NoMedia_BatchesResults`, `TestRequestBody_ThreeCallRound_BatchesResults`, `TestRequestBody_ShortRound_DropsUnpairedNames` pin the round-065 **batched shape** (names in call order; no ids). Round 066 extends them with an id axis. |
| `specs/truth/techstack.md` | the *Vertex/Gemini adapter* row records the batched rule + **"FIFO name pairing retained, `ToolCallID` optional"** — round 066 changes that clause. |
| `tell-me-go` (reference) | `llm.FunctionCall{ID,Name,Args}` / `llm.FunctionResponse{ID,Name,Response}` carry ids end-to-end; provenance = **provider id first**, else a deterministic `gemini-call-<index>-<name>` fallback; **empty id = invalid** (`isInvalidToolPart` strips such parts — "these parts cause API errors"). Its executor runs tools **concurrently** and copies `call.ID` into every response. *(Note: the reference still assembles positionally — `calls[i] ↔ results[i]`; a `map[id]` lookup is a hardening **beyond** it.)* |

**Why this is not a defect today:** the `AgentLoop` dispatches tool calls **sequentially** and appends results in call order, so the FIFO never mispairs. The failure it guards against — a result bound to the **wrong** call because results arrive **out of order** — needs **concurrent** tool dispatch, a settled exclusion recorded as [#36](https://github.com/gosharplite/tellme/issues/36) item 3 ("concurrent tool-call matching"). ADR 0035 **D2** deliberately kept the FIFO and deferred the id map to that slice; **#134** is its durable home.

---

## Design (S-1 is the round's settled goal; the rest is proposed — technical choices pending `/axb-technical-research`)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | **A Gemini/Vertex round's tool calls and results are id-linked on the wire, and each result is bound to its call by `ToolCallID`.** This is the change that closes [#134](https://github.com/gosharplite/tellme/issues/134). | **locked (round goal / #134)** |
| **S-2** | **The pairing is order-independent:** the i-th result is matched by the result's `ToolCallID`, not by queue position. A result whose id matches a call in the round takes that call's `name`; results are emitted in **call order** (the round-065 batched turn shape is unchanged). | proposed |
| **S-3** | **FIFO name matching is retained as the fallback** for a result with **no** `ToolCallID` (older persisted `steps`); a result with an unmatched id also falls back deterministically rather than dropping silently. The round-014 replay-fidelity invariant is preserved. | proposed |
| **S-4** | **Id provenance is a technical choice** — whether to prefer a **provider-issued** id (parse `functionCall.id` from the Vertex response) with a deterministic fallback, and the exact fallback spelling (`call_<n>` today vs the reference's `gemini-call-<index>-<name>`), is settled by `/axb-technical-research`. **The OpenAI-compatible wire MUST stay byte-identical** whatever is chosen. | proposed (research) |
| **S-5** | **The id is emitted only when non-empty**; an empty id is **omitted** from the part (never sent as `"id":""`, which the reference treats as invalid) — and the position falls back to FIFO (S-3). | proposed |
| **S-6** | **Exact unmatched accounting:** when a round yields `M < N` results (some calls unpaired), the drop is accounted by **call identity**, replacing the positional drop of the round-065 `pending[:0]` boundary (this also lets the deferred `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual — the pin cannot kill the pre-fold partial-drop mutant — be re-anchored to exact unmatched-id accounting). | proposed |
| **S-7** | **No new dependency / no new tool / no new config key / no port change** — stdlib-only (`encoding/json`); the change is confined to **`internal/infrastructure/llm/gemini`** (plus pins, plus possibly the round-066 truth rows). | proposed |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — The OpenAI-compatible wire is byte-preserved.** Its request serialization is correct today (one `role:"tool"` message per `tool_call_id`); the round MUST NOT change it, and its existing tests stay green and unchanged.
- **I-2 — The round-065 batched shape is preserved.** A round of **N** calls still serializes to **one** `user` turn carrying **N** `functionResponse` parts (in call order), with the round's media turns after it (ADR 0035 D1/D3). Round 066 **adds ids**; it does not change the turn/part structure.
- **I-3 — The media-free text path is byte-preserved** on **both** families (no tool parts ⇒ no id change).
- **I-4 — No silent media loss** (ADR 0032 D8 / ADR 0033): every image a round attached is carried on the wire or the turn fails loudly; never dropped (a fix MUST NOT satisfy a shape check by dropping media).
- **I-5 — Round-014 replay fidelity is preserved:** a resumed turn's persisted `steps` (which carry no stored wire id) still pair deterministically via the FIFO fallback; a replayed body is stable.
- **I-6 — No security layer, POSIX-only, hermetic.** No new dependency; every gate offline (ADR 0012).
- **I-7 — The universal reason gate (ADR 0025) and the round-024 tool resource contract are untouched.**

> **Byte-identity note (disclosed, not hidden):** because the round **adds an `id` key** to the Gemini `functionCall`/`functionResponse` parts, a **tool-bearing** Gemini body is **shape-identical** (same turns, roles, order, part counts, blob bytes) but **not byte-identical**. Byte-identity is claimed **only** where no tool part is emitted: the **media-free text path** (I-3) and the **OpenAI-compatible** wire (I-1). The round-065 I-2/I-3 byte claims are therefore **narrowed** to that scope — see *Assumptions A-B*.

---

## Clarify strategy

**Not escalated.** Per the clarify-escalation rule, only a gap that changes the user-story split, requirement attribution, the main flow, scope, or the formal acceptance criteria warrants `/axb-clarify`. Here the goal is unambiguous (**close #134**), the change is fully grounded (issue #134 carries the current behaviour, the reference comparison, and the proposed change), and the surviving choices are **technical** — the id provenance, the fallback spelling, and whether the drop accounting changes. Those are disclosed as **S-4/S-6** and in *Assumptions* rather than asked. No `NEEDS CLARIFICATION` remains.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - a round's tool results carry their calls' ids on the Gemini wire (Priority: P1)

As the **operator** running `tellme` against a **Gemini/Vertex** provider, I want each `functionResponse` part to carry the **id** of the `functionCall` it answers (and each `functionCall` part to carry an id), so that the round is **id-linked** on the wire — matching the reference and safe against a provider that requires it — instead of relying on a positional name queue.

**Why this priority**: it is the round's headline change (#134 item 1/3): the reference's whole design is *"tool parts carry ids; a part without one is an API error"*, and tellme's Gemini adapter is the outlier that drops them. It is also the prerequisite for US2.

**Independent verification**: build the request body for a tool round and assert every `functionCall` part and every `functionResponse` part carries an `id`, with the response's id equal to its call's id. A falsifiability witness omits the id and the check goes red.

**Acceptance Scenarios**:

1. **Given** a Gemini round whose model turn carries **N ≥ 1** `functionCall` parts, **When** the request body is built, **Then** every `functionCall` part carries an `id` and every `functionResponse` part carries the `id` of the call it answers.
2. **Given** a `tool` result whose `ToolCallID` is **empty** (a replay step), **When** the body is built, **Then** the `functionResponse` part emits **no** `id` key (never `"id":""`) and pairs via the FIFO fallback.
3. **Given** the **single-call** round and the **media-free text** round, **When** the body is built, **Then** the text path is **byte-identical** to today and the single-call shape is unchanged apart from the added id (I-2/I-3).

**Functional Requirements**:

- **FR-001**: Every `functionCall` part emitted to the Gemini/Vertex wire MUST carry an `id`.
- **FR-002**: Every `functionResponse` part emitted MUST carry the `id` of the call it answers (the two ids are equal for a matched pair).
- **FR-003**: When a result's `ToolCallID` is empty, the adapter MUST **omit** the `id` key on the response part — it MUST NOT emit `"id":""` (the reference treats an empty id as invalid).
- **FR-004**: The id-carrying MUST apply on **both** the live path and the **replay** (resumed-history) path.
- **FR-005**: The id MUST be **stable for a given round** so a replayed/resumed body is deterministic (no per-build random id).

### User Story 2 - a round's results bind to their calls by id, not by arrival order (Priority: P2)

As the **operator**, I want each tool result bound to the call it answers **by id**, so that a round's results are correctly paired **regardless of the order the results arrive in** — the property that keeps tellme correct if tool calls are ever dispatched concurrently — while a result with no id still pairs deterministically (FIFO) on the replay path.

**Why this priority**: it is the substantive robustness gain (#134 item 2), but it is not user-visible today (the loop is sequential), so it ranks below the id-carrying that makes it meaningful.

**Independent verification**: drive a round whose results are presented **out of call order** and assert each `functionResponse` carries its **own** call's `name` and `id` (not the arrival-order pairing); and drive a **replayed** round (no stored id) and assert the FIFO fallback still pairs correctly. Witnesses: a wrong-id pairing and a dropped-fallback each go red.

**Acceptance Scenarios**:

1. **Given** a Gemini round with **two** calls (`A`, `B`) whose results are presented in the order **`B, A`**, **When** the body is built, **Then** the response for `B`'s result carries `B`'s name and id and `A`'s carries `A`'s — i.e. the pairing follows **identity**, not position.
2. **Given** a **replayed** round whose `tool` steps carry only the synthesised deterministic id (or none), **When** the body is built, **Then** the results pair via the **FIFO fallback** and the body is stable (round-014 fidelity).
3. **Given** a round that yields **fewer** results than calls (`M < N`), **When** the body is built, **Then** the unpaired calls are dropped/accounted **by identity** (not by a positional truncation), and no result is silently mispaired.

**Functional Requirements**:

- **FR-006**: The adapter MUST bind each result to its call **by `ToolCallID`** when the result carries a non-empty id, **independent of arrival order**.
- **FR-007**: When a result carries no id (or an id matching no call of the round), the adapter MUST fall back to the **FIFO name match** deterministically — never dropping silently and never mispairing a later round (I-5).
- **FR-008**: When a round yields `M < N` results, the adapter MUST account the **unpaired calls by identity** (the emitted batched turn carries the `M` parts the round produced; the `N − M` unpaired calls are identified, replacing the round-065 positional `pending[:0]` drop).
- **FR-009**: The round-065 batched shape MUST be preserved — **one** `user` turn with **N** `functionResponse` parts in call order, media turns after it (I-2).
- **FR-010**: Every image a round attached MUST still reach the model (I-4); a shape fix MUST NOT drop or merge media.

---

## Edge Cases

- **N = 1** (single call) — id-carrying applies; the turn shape is unchanged (I-2).
- **N = 0** — no tool round; the text path is **byte-preserved** (I-3).
- **Results arrive out of order** — paired by id (US2 Scenario 1).
- **A result with no `ToolCallID`** (replay) — FIFO fallback; no `id` key emitted (FR-003).
- **`M < N`** (some calls unpaired) — accounted by identity (FR-008); the emitted turn carries the `M` produced parts.
- **An id that matches no call of the round** — falls back to FIFO deterministically (FR-007); recorded, never a silent mispair.
- **Media on only some calls / every call** — only those calls contribute `inlineData`; the batched turn carries all `M`/`N` results; no media loss (I-4).
- **A tool error / a reason-less refusal** — the (error/refusal) text is a `functionResponse` like any other; the count is preserved.
- **Two rounds in a turn** — each round is paired and batched independently; the round boundary is respected.
- **An OpenAI-compatible provider** — unchanged (I-1).
- **A provider that returns no `functionCall.id`** — the deterministic fallback id is used (S-4), never an empty id.

## Key Entities

- **A model round** — one assistant turn carrying N `functionCall` parts plus the N tool results (and any media) it produces; the unit whose call↔result pairing this round makes id-keyed.
- **A call id** — the identity linking a `functionCall` part to its `functionResponse` part; the round's new wire axis.
- **The Gemini request `contents`** — the serialized wire body; its construction is the round's subject (an added `id` key; unchanged turn structure).
- **A `functionResponse` part** — one tool result, bound to its call by id (fallback: FIFO name).
- **A media (`inlineData`) part** — an attached image (unbounded by this round; the family-aware ceiling standing).

## Success Criteria

- **SC-001**: Every `functionCall` and `functionResponse` part in a Gemini tool round carries an `id`; a response's id equals its call's id (US1). *(Unit pin over the built request body.)*
- **SC-002**: An **out-of-order** round pairs each result to its own call (name + id) — the identity pairing (US2 Scenario 1). *(Unit pin.)*
- **SC-003**: A **replayed** round (no stored id / empty `ToolCallID`) still pairs via the FIFO fallback and produces a **stable** body (I-5). *(Unit pin.)*
- **SC-004**: The **media-free text** path and the **OpenAI-compatible** wire are **byte-identical** to the pre-round serialization (I-1/I-3); the **tool-bearing Gemini** path is **shape-identical** apart from the added id (I-2). *(Regression pins.)*
- **SC-005**: `M < N` unpaired calls are accounted by identity; no result is mispaired, and the round-065 short-round residual is re-anchored to exact unmatched-id accounting if the pairing change reaches it (FR-008). *(Unit pin.)*
- **SC-006**: A red-capable carrier proves the change: removing the id from the wire reds SC-001; a positional (arrival-order) pairing reds SC-002; dropping the FIFO fallback reds SC-003.
- **SC-007**: `make verify` + `go test -count=1 ./...` green (incl. the E2E contract); the topology/DSL audit adds **no** new findings; no new dependency; `go.mod`/`go.sum` unchanged.

## Assumptions

- **A1**: The change is **family-local** — `internal/infrastructure/llm/gemini` (+ pins, + the round-066 truth rows). The OpenAI-compatible wire is correct as-is (I-1) and is **not** touched.
- **A2**: **This is a hardening/parity round, not a defect fix** — no user-visible behaviour changes today; the round is measured by request-body shape + byte/shape preservation, not a before/after UX diff. The falsifiability witness is **shape-level**.
- **A3**: The exact **id provenance and fallback spelling** (provider id first vs the synthetic `call_<n>`; the fallback format; whether the round-014 replay synthesised `call_step_<n>` is reused) is a **technical** decision for `/axb-technical-research` (S-4), bounded by I-1 (OpenAI wire frozen) and I-5 (replay stable).
- **A4**: **Truth impact expected** — `specs/truth/techstack.md` (the *Vertex/Gemini adapter* row: the id-linked pairing rule; the `ToolCallID` clause updated from "optional" to "carried when present, FIFO fallback otherwise"), and a **new ADR** (the id-keyed pairing decision; it **extends ADR 0035** and annotates its D2 "FIFO is sound for the sequential loop" note + its §Forward RF-065-1; the ADR-0033/0035 forward-annotation precedent). `/axb-api-plan` **NOOP**; `/axb-data-plan` **NOOP**.
- **A5**: **Testable hermetically.** The carrier is a **unit pin over the built request body** (the round-065 D5 primary carrier); the round-065 E2E journeys drive a multi-call round end to end unchanged. `/axb-spec-by-example` is **expected NOOP** (no user-visible behaviour change — the non-BDD-tooling/structure-round precedent, e.g. rounds 042/043/045–052); `/axb-dsl-refine` is **expected NOOP or a small MODIFY** to the round-065 `chat/calling-several-tools-in-one-round.feature` **iff** the fake provider can observe ids (RF-065-3 is a separate forward item) — the research decides. **No** real network/credential.
- **A6**: No new tool, no config key, no port change, no new dependency (`encoding/json` only); POSIX-only; hermetic (ADR 0012).
- **A7**: Byte-identity is claimed **only** for the media-free text path and the OpenAI-compatible wire; the tool-bearing Gemini body is **shape-identical** (the added `id` key is the sole difference) — the round-065 I-2/I-3 byte claims are narrowed accordingly (see *Byte-identity note*).

## Out of scope (recorded forward items)

- **Concurrent tool dispatch itself** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — this round makes the wire *ready* for concurrency; it does **not** add concurrent execution (sequential tools remain a settled exclusion). Closing #134 does **not** require it.
- **A fake-side contract check** (the E2E fake rejecting a `functionResponse`/`functionCall` count mismatch) — **RF-065-3**, a separate forward item.
- **One merged media turn per round** (ADR 0035 D3 variant (b)) — **RF-065-2**.
- **A family-agnostic round-batching/id concept** in the domain — **RF-065-4**.
- **The reference's exact `normalizeUserTurnParts` shape** (media reordered before the merged function responses) — tellme keeps standalone media turns (ADR 0033 D2).
- **`RF-065-5`** (the RF-063-x items: RF-063-1 the derived 14 MiB ceiling · RF-063-3 the aggregate bound · RF-063-4 the Files-API leg · RF-063-5 the dimension guard · RF-063-8 the family-blind oversize fixture · RF-063-9 the ceiling placement · RF-063-10 *(PM-owned)* the meta-Rule clean-up).
- **The `ToolSetSpec` seam** (RF-062-10 / RF-063-6) — a separate structural round.
- **Windows** (locked exclusion) · **a security/consent layer** (locked exclusion).
