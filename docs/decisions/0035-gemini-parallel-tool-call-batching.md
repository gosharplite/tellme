# ADR 0035 — Gemini/Vertex: a round's tool results share one `user` turn (batched `functionResponse` parts)

- **Status:** Accepted **(its D2 FIFO note + §Forward RF-065-1 delivered by [ADR 0036](0036-toolcall-id-pairing.md), round 066)**
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0033](0033-gemini-image-vision.md) (Gemini image vision — this ADR **lands its §Forward RF-063-7** and annotates its **D2** multi-call scope note), [ADR 0032](0032-agent-image-vision.md) (the per-call media message that sits *between* a round's tool results), `tell-me-go` (the parity precedent: its Gemini adapter normalizes a user turn so a function-call turn's responses share one turn), [ADR 0025](0025-mcp-tool-call-reason.md) (the universal reason gate — unchanged here), round 008 (the per-call tool message), round 013 (the Vertex adapter), round 019 (the fake provider's multi-call scripting), round 065 (`specs/plans/065-gemini-parallel-tool-calls` — this ADR's round), issue [#132](https://github.com/gosharplite/tellme/issues/132)

## Context

A live check during the round-063 closeout (2026-09-20) found that a **Gemini/Vertex** turn in which the model requests **two or more tools at once** fails on the **follow-up** request:

```text
tellme: the provider request failed: provider dev: provider returned status 400:
Please ensure that the number of function response parts is equal to the
number of function call parts of the function call turn.
```

It reproduces with **2 × `read_files`** (no media) and with **2 × `read_image`**; a **single** `read_image` succeeds. The defect is **media-agnostic** and **pre-existing** (per-call tool message: round 008; per-message `functionResponse` `user` turn: round 013; round 062 added the per-call media message; round 063 added the `inlineData` branch). The **OpenAI-compatible** family is unaffected — its wire uses separate `role:"tool"` messages per `tool_call_id`.

The cause is a split between the loop and the adapter:

- the loop appends **one** `tool` message **per call** (and one extra `user` media message **per media call**, immediately after its own tool result);
- `gemini.buildContents` maps **each** `tool` message to its **own** `user` `contents` entry carrying **one** `functionResponse` part, and each media message to its own entry.

A two-call round therefore serializes to `[model: fcA, fcB] [user: frA] [user: inlineData(A)] [user: frB] [user: inlineData(B)]`, and Vertex rejects it because the responses to a function-call turn must share **one** turn. ADR 0033 had already named the fix — **RF-063-7**, "round-scoped placement" — but deferred it.

## Decision

**D1 — The Gemini/Vertex adapter batches a round's tool results into ONE `user` turn.** When a `model` turn carries **N** `functionCall` parts, the adapter takes the **next N** `tool` messages as that round's results and emits them as **one** `user` `contents` entry carrying **N** `functionResponse` parts, **in call order**. The rule is stateless and identical on the **live** and **replay** paths. The change is **confined to `internal/infrastructure/llm/gemini`** — the loop, the message model, the ports, the tools and the config are unchanged; the **OpenAI-compatible wire is untouched** (it is correct as-is). The fix closes **[#132](https://github.com/gosharplite/tellme/issues/132)**.

**D2 — The pairing stays the existing FIFO name match; a `ToolCallID` is not required.** The i-th `functionCall`'s name pairs with the i-th result of the round (the adapter already keeps a `pending` name queue). A `tool` message without a `ToolCallID` (the replay path) pairs the same way. An id-keyed map is a recorded forward item (matching [#36](https://github.com/gosharplite/tellme/issues/36) item 3, concurrent matching). **(Round 066 / [ADR 0036](0036-toolcall-id-pairing.md) delivers this forward item: the adapter now id-links the parts and pairs by `ToolCallID`, keeping this FIFO as the fallback — so "a `ToolCallID` is not required" still holds, now because the id-less result falls back to FIFO.)**

**D3 — Media placement is round-scoped: the round's media turns follow the batched function-response turn.** A media message stays its **own** `user` `contents` entry (ADR 0033 D2/D3 — every media entry is standalone, so the `functionResponse`-before-`inlineData` hazard cannot arise), but the round's media entries are emitted **after** the single batched function-response turn. A two-call round with two images serializes to `[model: fcA, fcB] [user: frA, frB] [user: inlineData(A)] [user: inlineData(B)]`. This is exactly the "round-scoped placement" RF-063-7 named. (The reference's alternative — merging the media into the function-response turn, media-first — is **not** adopted: it would change the single-call shape and reintroduce the ordering hazard.)

**D4 — The wire contract is stated as an invariant: `len(functionResponse parts in a turn) == len(functionCall parts in the driving model turn)`.** And two byte-identity guarantees hold: a round with **N = 1** serializes exactly as before, and a media-free conversation keeps its current shape exactly.

**D5 — Verification: a unit pin over the request body (primary) + the superseded round-063 pin + the existing multi-call E2E.** A unit pin asserts, for a two-call round, one `user` turn carrying two `functionResponse` parts (names in call order) with the media turns after it; a three-call case; and the **N = 1**/**N = 0** byte-identity controls. `TestRequestBody_MultiCallRound_MediaTurnsInterleave` (the round-063 TD-063-1 pin of the interleaved shape) is **rewritten** to the batched shape. The Vertex-shaped fake provider already scripts a multi-call model turn (round 019), so the acceptance journeys drive it end to end. Witnesses: reverting the batching reds the unit pin (non-vacuity); dropping media to satisfy the count reds the two-image journey. No new `make verify` member.

**D6 — Governance: this ADR extends ADR 0033 and supersedes its forward item RF-063-7.** It annotates ADR 0033 **D2**'s multi-call scope note (the "pre-existing trait this round does not introduce" clause is resolved) and the §Forward RF-063-7 entry. ADR 0033's body is otherwise not edited; it remains an `Accepted` historical record (the ADR-0027/0028/0032 annotation precedent).

**D7 — No new dependency; stdlib-only; hermetic.** `encoding/json` only; POSIX-only; every gate offline (ADR 0012). A live re-check runs at the closeout (the round-059/061/063 precedent).

## Consequences

### Positive

- A Gemini/Vertex turn with any number of parallel tool calls completes; the 400 is gone. [#132](https://github.com/gosharplite/tellme/issues/132) closes.
- The OpenAI-compatible wire is **byte-identical** (a loop-side fix would have changed it).
- The single-call Gemini path (with or without media) and the text path are **byte-identical**.
- Every media entry stays standalone, so the ordering hazard the reference heals cannot arise.
- The fix is one adapter function — a pure transform provable by a unit pin.

### Negative / Accepted Trade-offs

- The adapter gains a **round-grouping rule** (it now derives a round boundary from a driving model turn) — a modest increase in adapter state, documented and unit-pinned.
- **Media from a multi-call round is not merged**; a future wire issue would revisit D3 variant (b) (one merged media turn) — recorded.
- The pairing remains **FIFO name-based** (D2); an explicit id map stays a forward item.

### Neutral

- No domain, config, history, estimator, or tool change; the loop's messages, the persisted `steps`, and the reason gate (ADR 0025) are untouched.
- The **concurrent tool-call matching** question ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) is unchanged — tellme's loop remains sequential.

## Alternatives Considered

1. **Fix it in the loop** (append one combined tool message per round; defer media). Rejected (D1) — it changes the OpenAI wire and the persisted step shape, for a concern the adapter owns.
2. **A family-agnostic "batch the round" domain concept.** Rejected (D1) — the OpenAI family neither needs nor wants it; recorded as a forward item.
3. **Merge the media into the batched function-response turn** (the reference's normalized shape). Rejected (D3) — breaks the single-call byte-identity and reintroduces the ordering hazard.
4. **One merged media turn per round** (D3 variant (b)). Rejected for the smallest diff; recorded as a forward item.
5. **Pair by `ToolCallID`.** Deferred (D2) — the FIFO is sound for the sequential loop; an id map is a forward item.
6. **Teach the fake provider to reject a malformed Vertex body.** Deferred (D5) — the unit pin is sharper; recorded as a forward item.
7. **E2E-only verification.** Rejected (D5) — weaker red-capability than a unit pin.

## Verification

- **Unit** — a pin over the built Vertex request body: a two-call round ⇒ one `user` turn with two `functionResponse` parts (call order) and the media turns after it; a three-call round; the **N = 1** and **N = 0** byte-identity controls; the superseded TD-063-1 pin rewritten.
- **E2E** — the Vertex-shaped fake provider scripts a multi-call model turn; the acceptance journeys assert the turn completes and the recorded request carried the round's results together (and, for images, that both pictures reached the fake). Rides `go test` / `make test`.
- **Falsifiability** — reverting to per-message turns reds the unit pin; dropping media to satisfy the count reds the two-image journey.
- **Live (non-gating, closeout)** — a real Vertex turn with two tool calls (the round-063 live-check follow-up). **PERFORMED 2026-09-20 (session 43) — PASSED.** Via the `coder` peer on the `dev` provider (`gemini-3.8-flash`, Vertex, workspace `ait-tellme`), a single turn issued **two tool calls in one round** — `read_files([note1.txt, note2.txt])` and `list_files(/tmp/tellme-livecheck)`, both dispatched at the same instant (`05:44:39`) from one model round; the follow-up request completed normally (`Payload: 11890/1000000 tokens`, `exit 0`, self-reported `ANOMALY: None`, and **zero** `400`s in the chrome). This is exactly the *media-free, two-tool-call* scenario that 400'd as *Run C* in the round-063 live check (Context above) — now green on a real endpoint, confirming the D1 batching on the wire. Evidence (transient): `/tmp/tellme-livecheck/r065_send.{stdout,stderr}`.

## References

- Live check (2026-09-20, round-063 closeout): 2 × `read_files` and 2 × `read_image` both 400; a single `read_image` succeeds — [#132](https://github.com/gosharplite/tellme/issues/132).
- `internal/infrastructure/llm/gemini/client.go` (`buildContents`), `internal/agent/agentloop.go` (the per-call appends), `internal/infrastructure/llm/gemini/client_image_test.go` (the superseded pin).
- [ADR 0033](0033-gemini-image-vision.md) (RF-063-7 — the fix named) · [ADR 0032](0032-agent-image-vision.md) (the per-call media message) · `specs/plans/065-gemini-parallel-tool-calls/` (`spec.md` US1/US2; `research.md` D1…D9).

## §Forward (deferred, non-blocking)

- **RF-065-1** — an explicit **`ToolCallID`-keyed** name pairing (with a FIFO fallback), the natural companion of the concurrent-tools slice ([#36](https://github.com/gosharplite/tellme/issues/36) item 3). **Homed → [#134](https://github.com/gosharplite/tellme/issues/134)** (the live issue: carry the call id on the Vertex wire + pair by it, reference-parity scope). **DELIVERED by [ADR 0036](0036-toolcall-id-pairing.md) (round 066)** — closes [#134](https://github.com/gosharplite/tellme/issues/134).
- **RF-065-2** — **one merged media turn per round** (D3 variant (b)) if a wire issue ever appears.
- **RF-065-3** — a **fake-side contract check** (the E2E fake rejects a `functionResponse`/`functionCall` count mismatch) for belt-and-braces; the unit pin is the current carrier.
- **RF-065-4** — a **family-agnostic round-batching concept** in the domain if a third family ever needs it.
- **RF-065-5** — the other round-063 forward items stay open (RF-063-1 the derived ceiling · RF-063-3 the aggregate bound · RF-063-4 the Files-API leg · RF-063-5 the dimension guard · RF-063-8 the family-blind fixture · RF-063-9 the ceiling placement · RF-063-10 *(PM-owned)* the meta-Rule clean-up).
- **Live check — CLOSED (verified)** — D7's closeout live re-check (the `## Verification` *Live* bullet) was **performed 2026-09-20 (session 43) and passed**; no residual.
