# Technical Research — Round 066: pair a round's tool results to their calls by `ToolCallID` (Gemini/Vertex)

**Plan package**: `specs/plans/066-toolcall-id-pairing`
**Truth owner**: `/axb-technical-research` → updates `specs/truth/techstack.md`
**Anchor**: [#134](https://github.com/gosharplite/tellme/issues/134) (ADR 0035 §Forward **RF-065-1**)
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1/US2 · FR-001…FR-010 · SC-001…SC-007 · S-1…S-7 · I-1…I-7 · A1…A7); the grounded current shape (§ *Grounded* in `spec.md`); `specs/truth/techstack.md`; `tell-me-go` (the reference, read locally).

**Goal**: close [#134](https://github.com/gosharplite/tellme/issues/134) — make the Gemini/Vertex adapter **id-link** a round's `functionCall`/`functionResponse` parts and pair each result to its call **by `ToolCallID`**, keeping the positional FIFO name match as a **fallback**. A **hardening/parity** round (no user-visible behaviour change today).

---

## D1 — The wire gains an `id` on both part kinds (the id-link)

The adapter emits `id` on **every** `functionCall` part (from the model turn's `llm.ToolCall.ID`) and on **every** `functionResponse` part (from the result's `llm.Message.ToolCallID`), so a call and its answer are id-linked on the Vertex `contents` wire. This is the reference's shape (`toSDKFunctionCall`/`toSDKFunctionResponse` both carry `ID`) and the prerequisite for D2.

- **Key spelling**: `"id"` (proto-JSON), consistent with the adapter's other camelCase keys (`functionCall`, `functionResponse`, `inlineData`, `mimeType`, `maxOutputTokens`).
- **Consequence — shape-identical, not byte-identical.** The added key changes a **tool-bearing** Gemini body's bytes; the turn/part **structure** (roles, order, counts, blob bytes) is unchanged. Byte-identity is therefore claimed **only** where no tool part is emitted: the media-free **text** path (I-3) and the **OpenAI-compatible** wire (I-1). Round 065's I-2/I-3 byte claims are **narrowed** to that scope (spec A7). Verified: the round-065 batch pins (`TestRequestBody_MultiCallRound_BatchesFunctionResponses` etc.) assert on decoded parts (names/roles/counts), not raw bytes, so the added key does not red them — they are **extended** with the id assertion.

## D2 — Pair by `ToolCallID` first, FIFO name as the fallback

`buildContents` keeps the round-scoped buffering (ADR 0035 D1/D3) but replaces the "pop `pending[0]`" pairing with a **per-round call index**: when a `model` turn arrives, its calls are recorded as ordered `(id → name)` entries; each `tool` result binds to the call whose `id == result.ToolCallID` — **order-independent**. A result whose `ToolCallID` is empty, or matches no call of the round, binds by the **FIFO name match** (the round-065 behaviour) — deterministic, never a silent mispair.

- The emitted batched `user` turn still carries the results **in call order** (a result set presented out of order is emitted ordered by its matched call's position), preserving the round-065 wire shape.
- **Why**: the loop is **sequential** today, so position and identity agree; the id binding is the property that survives a future **concurrent** dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — the exact motivation ADR 0035 **D2** recorded when it deferred this.

## D3 — Id provenance: keep the existing deterministic ids (the smallest, safest choice)

The ids the adapter emits are the **already-produced** ones — the live path's `call_<n>` (from `parseResponse`) and the replay path's `call_step_<n>` (from `agentloop.go`) — **unchanged**. The adapter **does not** mint a new id and **does not** change the id format.

- **Why not "provider id first" (the reference's `f.ID`)?** tellme parses the Vertex response itself (no SDK) and does not read `candidates[].content.parts[].functionCall.id` today; reading and **preferring** it would change the live id value, which flows to the Gemini wire `id` **and** — for the OpenAI-compatible family — to `tool_call_id`. That is a behaviour change beyond #134's scope and risks the OpenAI byte-freeze (I-1). It is recorded as a **forward item (RF-066-2)**, not taken.
- **Replay stability (I-5)** holds trivially: `call_step_<n>` is deterministic per step, so a replayed body is stable and differs from the pre-066 body only by the added `id` values (which are themselves deterministic).

## D4 — Emit the id only when non-empty (never `"id":""`)

A part emits `id` **iff** its source id is non-empty; an empty id emits the key **omitted** and the position falls back to FIFO (D2). The reference treats an empty id as **invalid** (`isInvalidToolPart` strips it); tellme neither strips nor sends an empty id — it simply omits it, so the round can never turn a present-but-untagged result into a provider 400.

## D5 — The round boundary: unmatched calls are identified, output unchanged

At a round boundary the adapter still emits the `M` parts the round produced (ADR 0035's recorded shape). With the per-round call index (D2) the **unpaired** calls are now known **by identity** (the ids with no result); the **observable output is unchanged** (no part is emitted for an unpaired call either way). This is an internal clarity gain; the deferred `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual (its pin cannot kill the pre-fold partial-drop mutant) becomes expressible as exact unmatched-id accounting — **re-anchored, not re-scoped** (SC-005). No new failure mode.

## D6 — Scope: family-local, adapter-only

The change is confined to `internal/infrastructure/llm/gemini` (`client.go` `buildContents` (+ its id threading) and `parseResponse` only if D3 needs it — it does not). The **OpenAI-compatible** adapter, the loop, the ports, the tools, the config, and the persisted records are **untouched** (I-1). No new dependency (`encoding/json` only); POSIX-only; hermetic.

## D7 — Verification: unit pins + the unchanged E2E journeys

- **Primary carrier — a unit pin over the built request body** (the round-065 D5 precedent): every `functionCall`/`functionResponse` part carries an `id`; a response's id equals its call's id; an **out-of-order** result set pairs by identity; an **empty-`ToolCallID`** result omits the id and pairs by FIFO; the media-free text path is byte-identical; the OpenAI-compatible wire is byte-identical.
- **E2E** — the round-065 `chat/calling-several-tools-in-one-round.feature` journeys are **unchanged** (the change is not user-visible); no new E2E assertion is required unless the fake is taught to expose ids (RF-065-3, a separate item) — `/axb-dsl-refine` is therefore **NOOP**.
- **Falsifiability** — omit the id ⇒ the shape pin reds; pair positionally ⇒ the out-of-order pin reds; drop the FIFO fallback ⇒ the empty-id pin reds.

## D8 — Governance: a new ADR, extending ADR 0035

**ADR 0036** records the id-link + id-keyed pairing decision. It **extends ADR 0035** and annotates (README index + `Status`) its **D2** "the FIFO is sound for today's sequential loop; an id-keyed map is a forward item" note and its §Forward **RF-065-1** (now delivered / homed #134). ADR 0035's body is otherwise not edited (the ADR-0027/0028/0032/0033/0035 annotation precedent).

---

## Truth impact (semantic units)

| Action | Truth Spec | Summary |
| --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the adapter now emits an `id` on every `functionCall`/`functionResponse` part (when non-empty) and binds each result to its call by `ToolCallID` (order-independent), with FIFO name matching retained as the fallback; the round-065 *"`ToolCallID` is not required"* clause is updated. |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | the round-065 batched-function-response rule gains the **id axis** (the batched turn's parts carry ids). |
| ADD | `docs/decisions/0036-toolcall-id-pairing.md` (+ index row; an annotation on ADR 0035's `Status` + its D2 note + its §Forward RF-065-1) | the id-link + id-keyed pairing decision. |
| NOOP (checked) | *OpenAI-compatible* rows | unchanged (byte-frozen; I-1). |

## Residual risks / forward items (non-blocking)

- **RF-066-1** — the added `id` key narrows the byte-identity claim for a **tool-bearing** Gemini body to **shape**-identity (recorded; the round-065 byte claims are scoped).
- **RF-066-2** — **provider-issued id preference** (read `candidates[].content.parts[].functionCall.id` and prefer it, reference parity) — deferred (D3); would change the live id value and touch the OpenAI wire, so it is its own decision.
- **RF-066-3** — a **fake-side contract check** (the E2E fake rejects a count mismatch) — RF-065-3; still deferred.
- **RF-066-4** — concurrent tool **execution** itself ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — this round makes the wire *ready*; it does not add concurrency (sequential tools remain a settled exclusion).

## References

- `internal/infrastructure/llm/gemini/client.go` (`buildContents`, `parseResponse`) · `internal/agent/agentloop.go` (replay `call_step_<n>`) · `internal/infrastructure/llm/openai/client.go` (`tool_call_id`).
- `specs/plans/065-gemini-parallel-tool-calls/` (the batched shape this round extends) · `docs/decisions/0035-gemini-parallel-tool-call-batching.md` (D2 + RF-065-1).
- Reference: `tell-me-go/internal/infrastructure/llm/gemini/adapter.go` (`fromSDKFunctionCall`/`fromSDKFunctionResponse`), `tell-me-go/internal/agent/session/context/transformers.go` (`isInvalidToolPart`), `tell-me-go/internal/agent/executor/executor.go` (`buildFunctionResponse`).
- [#134](https://github.com/gosharplite/tellme/issues/134) · [#36](https://github.com/gosharplite/tellme/issues/36) item 3.
