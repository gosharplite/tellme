# ADR 0036 — Gemini/Vertex: id-link a round's tool calls and results, and pair them by `ToolCallID`

- **Status:** Accepted (**§Forward RF-066-2 and RF-066-7 delivered, and RF-066-8 retired, by [ADR 0037](0037-gemini-toolcall-id-provenance.md) — round 067**)
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0035](0035-gemini-parallel-tool-call-batching.md) (Gemini parallel tool-call batching — this ADR **delivers its §Forward RF-065-1** and annotates its **D2** FIFO note), [ADR 0033](0033-gemini-image-vision.md) (the Gemini media wire), round 008 (the per-call `tool` message), round 013 (the Vertex adapter), round 065 (`specs/plans/065-gemini-parallel-tool-calls` — the batched shape this extends), round 066 (`specs/plans/066-toolcall-id-pairing` — this ADR's round), issue [#134](https://github.com/gosharplite/tellme/issues/134), issue [#36](https://github.com/gosharplite/tellme/issues/36) item 3 (concurrent tool-call matching), `tell-me-go` (the parity precedent: its Gemini adapter carries `FunctionCall.ID`/`FunctionResponse.ID` end-to-end)

## Context

Round 065 (ADR 0035) made a Gemini/Vertex round's tool results **batch into one `user` turn**, closing the ≥2-tool-call 400 ([#132](https://github.com/gosharplite/tellme/issues/132)). Its **D2** deliberately kept the pairing **positional** — a `pending []string` FIFO of call **names**, popped one per result — and recorded the id-keyed pairing as **§Forward RF-065-1** (the "concurrent-tool-call matching" companion of [#36](https://github.com/gosharplite/tellme/issues/36) item 3).

The FIFO is **correct today** only because the `AgentLoop` dispatches tool calls **sequentially** and appends results in call order. It stops being correct the moment results can arrive **out of order** (concurrent dispatch): the i-th result would take the i-th queued name, binding the wrong name to the wrong result. The adapter also emits **no `id`** on either part, while the **OpenAI-compatible** family already sends `tool_call_id` and the reference carries `FunctionCall.ID`/`FunctionResponse.ID` end-to-end (its `isInvalidToolPart` even treats an empty id as invalid — "these parts cause API errors").

[#134](https://github.com/gosharplite/tellme/issues/134) homes RF-065-1: carry the call id on the Vertex wire and bind each result to its call by it, keeping the FIFO as a fallback. This is a **hardening/parity** change — **no user-visible behaviour changes today**.

## Decision

**D1 — The Gemini/Vertex adapter emits an `id` on every `functionCall` and every `functionResponse` part (when non-empty).** The `functionCall` part's id is the model turn's `llm.ToolCall.ID`; the `functionResponse` part's id is the result message's `llm.Message.ToolCallID`. The key is proto-JSON `"id"`, consistent with the adapter's other camelCase keys. A part whose source id is **empty** emits the key **omitted** (never `"id":""` — the reference's invalid shape).

**D2 — The adapter binds each result to its call by `ToolCallID`, keeping the FIFO name match as the fallback.** While buffering a round (ADR 0035 D1/D3), the adapter records the round's calls as ordered `(id → name)` entries and binds a result to the call whose `id == result.ToolCallID` — **order-independent**. A result whose `ToolCallID` is empty, or matches no call of the round, binds by the existing FIFO name match (deterministic; never a silent mispair). The emitted batched `user` turn is unchanged: the round's results in **call order** (an out-of-order result set is emitted ordered by its matched call's position). **On the fallback path the response `id` is emitted only when it equals the bound call's id** — a *foreign* id (an unmatched result's id, present on no `functionCall` part of the round) is **omitted**, preserving the reference's `response.id == call.id` invariant and avoiding a wire that could itself 400 a strict provider (the auditor review's TD-066-1).

**D3 — The ids are the existing deterministic ones; no new id is minted.** The adapter uses `call_<n>` (live, from `parseResponse`) and the replay path's `call_step_<n>` (from `agentloop.go`) **unchanged**. Preferring a **provider-issued** id (reading `candidates[].content.parts[].functionCall.id`) is **not** taken — it would change the live id value, which flows to the OpenAI-compatible `tool_call_id`, and is a separate decision (recorded as a forward item). **The replay path is id-primary, not FIFO:** the loop's `BuildMessages` synthesises the **same** `call_step_<n>` on both the assistant call and the tool result, so a replayed round binds by identity and the FIFO fallback is **defensive** (for a future producer that omits an id) — corrected at the review's TD-066-2.

**D4 — Scope is family-local: `internal/infrastructure/llm/gemini`.** The OpenAI-compatible adapter, the loop, the ports, the tools, the config, and the persisted records are **untouched**. No new dependency (`encoding/json` only); POSIX-only; hermetic.

**D5 — The round boundary (and its drop) is unchanged.** At a round boundary the adapter still emits the `M` parts the round produced (ADR 0035's recorded shape); a call left unpaired contributes no part and `pending` is cleared — behaviourally identical to the round-065 `pending[:0]`. The **`N - M` unpaired calls are not surfaced** (no error, log, or accessor); surfacing them is a forward item (RF-066-7). No new failure mode. *(Corrected at the fold verification's F-066-3: this clause previously claimed the unpaired calls are identified by identity, which the code does not deliver.)*

**D6 — Invariants held.** The OpenAI-compatible wire is **byte-preserved**; the round-065 batched **shape** is preserved (one `user` turn, N `functionResponse` parts, media turns after) — the added `id` key is the sole difference, so a **tool-bearing** Gemini body is **shape-identical, not byte-identical**; the media-free **text** path is **byte-identical**; no silent media loss (ADR 0032/0033) — the id-less-`tool`-message widening is restricted to a **media-free** `tool` message, so a media-bearing one is still carried (the review's F-066-2); round-014 replay fidelity holds via the **id-primary** match (D3), with the FIFO fallback defensive.

**D7 — Verification: a unit pin over the built request body (primary) + the unchanged E2E journeys.** A unit pin asserts (a) every `functionCall`/`functionResponse` part carries an `id` and a response's id equals its call's id; (b) an **out-of-order** result set pairs by identity; (c) an **empty-`ToolCallID`** result omits the id and pairs by FIFO; (d) the media-free text path is byte-identical; (e) the OpenAI-compatible wire is byte-identical. The round-065 E2E journeys are unchanged (the change is not user-visible). Falsifiability: omitting the id reds (a); a positional pairing reds (b); dropping the fallback reds (c). No new `make verify` member.

**D8 — Governance: this ADR extends ADR 0035 and delivers its §Forward RF-065-1.** It annotates ADR 0035's **D2** "the FIFO is sound for today's sequential loop; an id-keyed map is a forward item" note and its RF-065-1 entry. ADR 0035's body is otherwise not edited; it remains an `Accepted` historical record (the ADR-0027/0028/0032/0033/0035 annotation precedent).

## Consequences

### Positive

- A round's calls and results are **id-linked** on the Vertex wire — reference parity, and safe against a provider that requires it.
- Pairing is **order-independent**: the adapter stays correct if tool calls are ever dispatched concurrently ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — the property ADR 0035 D2 anticipated.
- The **id-primary** pairing preserves the **replay** path (the loop's `BuildMessages` sets the same `call_step_<n>` on both sides) — the FIFO fallback is the defensive path for an id-less result.
- The OpenAI-compatible wire and the media-free text path are **byte-identical**; the round-065 batched shape is **preserved**.
- One adapter function + pins — small and provable.

### Negative / Accepted Trade-offs

- The added `id` key makes a **tool-bearing** Gemini body **shape-identical, not byte-identical** (narrowing the round-065 byte claim for that path) — recorded.
- The adapter gains a small per-round call index (a modest, unit-pinned state increase alongside ADR 0035's round-grouping rule).
- The id provenance keeps tellme's own deterministic ids rather than the provider's; parity on that axis (provider-id preference) is a recorded forward item.

### Neutral

- **No user-visible behaviour change today** — the loop is sequential, so position and identity agree; the witness is request-body **shape**.
- No domain, config, history, estimator, or tool change; the persisted `steps`, the reason gate (ADR 0025), and the round-024 resource contract are untouched.
- Concurrent tool **execution** is **not** added (a settled exclusion); the wire is merely made ready.

## Alternatives Considered

1. **Keep the FIFO only (do nothing).** Rejected — it is correct only while dispatch is sequential; #134 exists to remove that latent coupling.
2. **Emit no id, pair by id anyway (derive an id internally).** Rejected (D1) — the id-link (a) is the round's headline and (b) is what makes the pairing legible on the wire (reference parity).
3. **Prefer the provider-issued id (reference's `f.ID`).** Deferred (D3) — changes the live id value and touches the OpenAI wire; its own decision.
4. **Pair by id only, no FIFO fallback.** Rejected (D2/D6) — an id-less result (a hypothetical legacy/future producer; the replay path itself is id-primary) would be left unpaired; the fallback is the defensive, deterministic path (I-5).
5. **Fix it in the loop** (attach the id to the message differently). Rejected (D4) — the concern is the adapter's wire; the loop already sets `ToolCallID` on every result.
6. **Teach the E2E fake to reject a count mismatch** (RF-065-3). Deferred (D7) — the unit pin is sharper; a separate forward item.
7. **Merge the id into ADR 0035 instead of a new ADR.** Rejected (D8) — the ADR-annotation precedent: 0035 is an `Accepted` record; the new decision is its own.

## Verification

- **Unit** — a pin over the built Vertex request body: every `functionCall`/`functionResponse` part carries an `id` (response id == call id); an **out-of-order** result set pairs by identity; an **empty-`ToolCallID`** result omits the id and pairs by FIFO; the media-free text body is byte-identical; the OpenAI-compatible body is byte-identical.
- **E2E** — the round-065 `chat/calling-several-tools-in-one-round.feature` journeys are unchanged and stay green (the change is not user-visible). Rides `go test` / `make test`.
- **Falsifiability** — omitting the id reds the shape pin; a positional pairing reds the out-of-order pin; dropping the FIFO fallback reds the empty-id pin.
- **Live (non-gating)** — a real Vertex turn with two tool calls (the round-063/065 live-check precedent). The `id` key is the round's *only wire-visible* delta and is live-unverified hermetically (the review's R-066-1), so a round-065-style live check is **performed at the closeout** and its outcome recorded here; the key is independently revertible if it fails. **PERFORMED 2026-09-20 (session 45) — PASSED, with recorded provenance** (see the Outcome note below). *(The first attempt was **superseded**: it was dispatched through the GOPATH-installed `tellme`, which was still the round-065 build — fold-verification **F-066-4** — so it did not exercise the new key; the re-run below used the branch binary.)*

## References

- `internal/infrastructure/llm/gemini/client.go` (`buildContents`, `parseResponse`), `internal/agent/agentloop.go` (replay `call_step_<n>`), `internal/infrastructure/llm/openai/client.go` (`tool_call_id`).
- [ADR 0035](0035-gemini-parallel-tool-call-batching.md) (the batched shape; D2 + RF-065-1) · [ADR 0033](0033-gemini-image-vision.md) (the media wire) · `specs/plans/066-toolcall-id-pairing/` (`spec.md` FR-001…FR-010; `research.md` D1…D8).
- Reference: `tell-me-go/internal/infrastructure/llm/gemini/adapter.go` (`fromSDKFunctionCall`/`fromSDKFunctionResponse`), `tell-me-go/internal/agent/session/context/transformers.go` (`isInvalidToolPart`), `tell-me-go/internal/agent/executor/executor.go` (`buildFunctionResponse`).
- [#134](https://github.com/gosharplite/tellme/issues/134) · [#36](https://github.com/gosharplite/tellme/issues/36) item 3.

## §Forward (deferred, non-blocking)

- **RF-066-1** — the added `id` key narrows a **tool-bearing** Gemini body's byte-identity claim to **shape**-identity (recorded; the round-065 byte claims are scoped to the media-free text path and the OpenAI-compatible wire).
- **RF-066-2** — **provider-issued id preference** (read the Vertex `functionCall.id` and prefer it — reference parity); changes the live id value and touches the OpenAI wire, so it is its own decision. **Homed → [#136](https://github.com/gosharplite/tellme/issues/136)** (the round-067 anchor). **DELIVERED by [ADR 0037](0037-gemini-toolcall-id-provenance.md) (round 067)** — the provider id is preferred, Gemini-local by construction (the OpenAI-compatible wire is untouched).
- **RF-066-3** — a **fake-side contract check** (the E2E fake rejects a `functionResponse`/`functionCall` count mismatch) — RF-065-3; still deferred.
- **RF-066-4** — concurrent tool **execution** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) — the wire is now ready; the execution change is a separate, future decision. **⚠ TRIGGER-GATED — not open work; do not re-raise absent the trigger.**
- **RF-066-5** — the other round-065 forward items (RF-065-2 one merged media turn per round · RF-065-4 a family-agnostic round-batching concept · RF-065-5 the RF-063-x items) stay open.
- **RF-066-6** — an **order-independence carrier**: US2's order-independent property has no in-system producer today (the loop is sequential; a replayed round is one call/one result), so it is exercised only by a hand-built fixture; a real carrier arrives with concurrent dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3). *(the review's R-066-2.)* **⚠ TRIGGER-GATED — do not re-raise absent the [#36](https://github.com/gosharplite/tellme/issues/36) item-3 trigger.**
- **RF-066-7** — surface the **unpaired** calls of an `M < N` round (an id accessor) to replace the still-open `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual; this round leaves the boundary drop unchanged. *(the review's F-066-1 option (i).)* **Homed → [#136](https://github.com/gosharplite/tellme/issues/136)** (the round-067 anchor). **DELIVERED by [ADR 0037](0037-gemini-toolcall-id-provenance.md) (round 067)** — the single-owned `UnpairedCallIDs` accessor surfaces the unpaired call ids.
- **RF-066-8** — the still-open `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual (the pin cannot kill the pre-fold partial-drop mutant — inherent `M == N/2` arithmetic equivalence); carried from round 065, unchanged this round. **Homed → [#136](https://github.com/gosharplite/tellme/issues/136)** (retired by RF-066-7's accessor). **RETIRED by [ADR 0037](0037-gemini-toolcall-id-provenance.md) (round 067)** — the **cross-round** account pin (`TestUnpairedCallIDs_MultiRound`) kills the partial-drop mutant; the short-round pin alone cannot (arithmetic equivalence), per the round-067 review F-067-1.
- **RF-066-9** *(fold-verification R-066-6)* — the replay id-primary claim is pinned with a **hand-built `prior`**, not `agent.BuildMessages` output, and no E2E asserts the wire `id`; adapter↔`BuildMessages` coherence rests on fixture convergence.
- **RF-066-10** *(fold-verification R-066-5, convention)* — the in-group live-check harness resolves the closeout-refreshed **GOPATH** binary, so a **mid-round** live check silently exercises the last install; a mid-round live check must build the **branch** binary and record its `go version -m` provenance.

## Outcome

- **The architect review (round 066, reviewer `5258004457`) — `APPROVE WITH REQUIRED FOLDS` (0 blockers).** Folded in-round: **F-066-1** (the "exact unmatched-identity accounting" claim corrected → the boundary drop is unchanged; the residual re-homed as RF-066-7/RF-066-8), **F-066-2** (the id-less-`tool` widening restricted to a media-free `tool` message — one line + a pin — so a media-bearing `tool`-role message is still carried), **TD-066-1** (a foreign response `id` is omitted — wire self-consistency), **TD-066-2** (the replay mechanism is id-primary, not FIFO; a replayed-parts-carry-equal-ids pin added; the mechanism wording corrected here + in the truth row), **N-066-1/N-066-2** (spec Status + the truth-row rewrite → an appended sentence).
- **`## Verification` *Live*** — **PERFORMED 2026-09-20 (session 45) and PASSED, provenance recorded**: the `coder` peer ran on the `dev` provider (`gemini-3.8-flash`, Vertex, workspace `ait-tellme`) with the **branch-built binary** `/tmp/tellme-066/tellme` — `go version -m` → `github.com/gosharplite/tellme v0.0.0-20260919221828-45239e5757f6` (branch head `45239e5`, clean tree, i.e. **round-066 code incl. the `id` key**). One turn issued **two tool calls in one round** (`read_files([note1,note2])` + `list_files(/tmp/tellme-livecheck)`) and completed (`exit 0`, `ANOMALY: None`, **zero** `400`s) — the added `id` key is accepted by the live Vertex parser. Evidence (transient): `/tmp/tellme-livecheck/r066b_send.{stdout,stderr}`. *(The earlier `r066_send.*` run used the GOPATH round-065 binary and is **superseded** — F-066-4.)*
