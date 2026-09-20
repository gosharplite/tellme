# ADR 0037 — Gemini/Vertex tool-call id provenance (provider-issued `functionCall.id` preference) + unpaired-call accounting

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0036](0036-toolcall-id-pairing.md) (id-link + id-keyed pairing — this ADR **delivers its §Forward RF-066-2 + RF-066-7** and **retires RF-066-8**), [ADR 0035](0035-gemini-parallel-tool-call-batching.md) (the batched round shape), round 066 (`specs/plans/066-toolcall-id-pairing`), round 067 (`specs/plans/067-toolcall-id-followups` — this ADR's round), issue [#136](https://github.com/gosharplite/tellme/issues/136), issue [#36](https://github.com/gosharplite/tellme/issues/36) item 3 (concurrent tool-call matching), `tell-me-go` (the parity precedent: `fromSDKFunctionCall` reads `f.ID` first, else a deterministic fallback)

## Context

Round 066 (ADR 0036) id-linked a Gemini/Vertex round's `functionCall`/`functionResponse` parts and paired each result to its call by `ToolCallID`. It **deliberately deferred** two things, recorded in its §Forward:

- **RF-066-2** — the id **provenance**: round 066 kept the synthetic positional `call_<n>` (`parseResponse`) and did **not** read the provider's own `functionCall.id`; the reference prefers the provider id (`f.ID` first).
- **RF-066-7 / RF-066-8** — the **unpaired-call accounting**: at a round boundary a call that received no result (`M < N`) contributes no part and is **silently dropped** (no error, log, or accessor), leaving the `TestRequestBody_ShortRound_DropsUnpairedNames` `N=2 M=1` residual un-killable (the emitted turn is identical whether one or both unpaired calls are dropped — inherent `M == N/2` arithmetic equivalence).

[#136](https://github.com/gosharplite/tellme/issues/136) homes both as round 067. This is a **hardening/parity** change — **no user-visible behaviour changes today** (the loop is sequential; the id value has no shipped user surface).

## Decision

**D1 — The Gemini/Vertex adapter prefers the provider-issued `functionCall.id`.** `parseResponse` reads `candidates[0].content.parts[].functionCall.id`; when it is non-empty it **becomes** the call's `llm.ToolCall.ID`; otherwise the existing deterministic `call_<n>` (positional) is used. A blank/whitespace provider id is treated as **absent** (never an empty id on the wire — ADR 0036 D2/D4). A provider-id-**absent** response keeps today's behaviour exactly — **no new failure mode**.

**D2 — The fallback spelling stays `call_<n>`.** The reference's `gemini-call-<index>-<name>` is **not** adopted: tellme's existing spelling is already deterministic and in place, and changing it would churn the round-066 fixtures for no behavioural gain (a recorded divergence). Both the provider id and the fallback are **deterministic per turn**, so a replayed/resumed body is stable (round-014 fidelity; ADR 0036 D3).

**D3 — Cross-family scope: Gemini-local by construction; the OpenAI-compatible wire is untouched.** The id value flows, at the loop, into `llm.Message.ToolCallID` — which the OpenAI-compatible adapter serializes as `tool_call_id`. The safety argument is **not** "one session drives exactly one family" (a resumed session's provider is whatever the *current* config selects — the environment model's hot-swap — so that alone does not hold); the load-bearing invariant is that **the id is never persisted and never replayed**: `history.Step` carries no id field, and the replay path synthesises its own deterministic `call_step_<n>` on both the call and the result (`internal/agent/agentloop.go`). A Gemini-produced id therefore lives only inside the round that produced it and can never reach a later (possibly OpenAI-compatible) request — **no OpenAI-compatible wire changes** and **I-1 holds by construction**. The unification alternative (a single id convention across families) is **not** taken — it would couple the families for no gain. *(Reason corrected at the round-067 architect review F-067-6; a guarding comment on `history.Step`'s id-less shape is a recorded forward item, RF-067-7.)*

**D4 — The unpaired calls of an `M < N` round are made accountable (a single-owned accessor).** `roundBuilder` gains `unpaired() []string` — the ids, in call order, of the round's calls that received no result — computed at the **one** place the drop happens (`flush()`), so the accounting cannot drift between call sites. It is surfaced as a **returned value** (`UnpairedCallIDs(prior []llm.Message) []string`, a documented package-level accessor), **not** a `stderr`/`[Tool …]` diagnostic (the adapter owns no logging seam; a diagnostic is a forward item). **Scope of the claim (F-067-2):** the accessor has **no live consumer** today, so **at runtime the boundary drop remains exactly as silent as before round 067**; what round 067 lands is that the drop is **accountable in code** through a single owner, not that it is surfaced to the operator. The emitted batched turn is **unchanged** — it still carries the `M` parts the round produced (ADR 0035's recorded shape). `UnpairedCallIDs` and `buildContents` share **one build pass** (`buildRound`), so the account and the emitted body cannot disagree by construction (F-067-3).

**D5 — RF-066-8 is retired by the cross-round account.** With the `M < N` boundary drop accounted, a pin can assert the **identity** of the unpaired calls. **The carrier is `TestUnpairedCallIDs_MultiRound`** (the cross-round account), **not** the short-round pin: at `N=2 M=1` the pre-fold *conditional partial drop* is arithmetically equivalent to the correct form (the very residual RF-066-8 named), so the short-round pin stays green under that mutant. Reinstating the pre-fold conditional drop (retaining a stale `pending` slice instead of the unconditional `b.pending = nil`) reds `TestUnpairedCallIDs_MultiRound` — the retained stale calls are re-accounted on the next flush (`[call_r1b call_r1b call_r2a]` ≠ `[call_r1b call_r2a]`). *(Attribution corrected at the round-067 architect review F-067-1; the witness is recorded as `tasks.md` T009(d).)*

**D6 — Scope is family-local: `internal/infrastructure/llm/gemini`.** The OpenAI-compatible adapter, the loop, the ports, the tools, the config, and the persisted records are untouched. No new dependency (`encoding/json`/`fmt` only); POSIX-only; hermetic. The message switch is factored into `consume()` so `buildContents` and `UnpairedCallIDs` share one walk and stay under the `cyclop` max.

**D7 — Invariants held.** The round-066 batched shape is preserved (I-2); the media-free text path is byte-identical (I-3); no silent media loss (I-4); replay stays id-primary with the FIFO fallback defensive (I-5); an empty/foreign id stays omitted (I-8).

**D8 — Verification: unit pins over the built request body + the accessor.** (a) a response carrying `functionCall.id` ⇒ the emitted parts carry that id; (b) a response without a provider id ⇒ the deterministic `call_<n>` fallback; (c) an `M < N` round ⇒ the produced `M` parts named correctly **and** `UnpairedCallIDs` returns exactly the unpaired id(s); (d) `M = N` ⇒ `UnpairedCallIDs` empty; (e) the byte/shape controls stay green. Falsifiability: removing the preference reds (a); suppressing the accessor reds (c); a non-deterministic fallback reds (b). No new `make verify` member.

**D9 — Governance: this ADR extends ADR 0036 and delivers its §Forward RF-066-2 and RF-066-7 (and retires RF-066-8).** It annotates those entries and the README index row. ADR 0036's body is otherwise not edited; it remains an `Accepted` historical record (the ADR-0035/0036 annotation precedent).

## Consequences

### Positive

- **Reference parity on id provenance**: the Gemini/Vertex family now prefers the provider's own `functionCall.id` — matching the reference and the OpenAI-compatible adapter's existing behaviour.
- **The boundary drop is accountable in code**: the unpaired call ids are single-owned and observable through the accessor; the round-066 residual (**RF-066-8**) is retired by the **cross-round** account pin (D5). *(It remains silent at runtime — the accessor has no live consumer; surfacing it is RF-067-1.)*
- The **OpenAI-compatible wire is untouched** by construction (the id is never persisted/replayed; D3); the round-066 batched shape and the media-free text path are preserved.
- One adapter file + pins — small and provable.

### Negative / Accepted Trade-offs

- The observability is a **code accessor**, not a user-visible diagnostic (deferred; no adapter logging seam).
- The provider-id preference is **live-unverified** hermetically (the E2E fake scripts no `functionCall.id`) — a branch-built live check is a closeout item.
- `parseResponse` gains a small id-selection branch; `roundBuilder` gains an accumulation field + `unpaired()` (modest, unit-pinned).

### Neutral

- **No user-visible behaviour change today** — the witness is request-body shape + the accessor.
- No domain, config, history, estimator, or tool change; the persisted `steps`, the reason gate (ADR 0025), and the round-024 resource contract are untouched.
- Concurrent tool **execution** is **not** added (a settled exclusion).

## Alternatives Considered

1. **Do nothing (keep the synthetic id; keep the silent drop).** Rejected — #136 exists to close both; the provider-id axis is reference parity and the silent drop is an accountability gap.
2. **Adopt the reference's fallback spelling (`gemini-call-<index>-<name>`).** Rejected (D2) — churns fixtures for no behavioural gain; the existing `call_<n>` is already deterministic.
3. **Unify the id convention across families (S-5 (ii)).** Rejected (D3) — couples two families for no gain; the Gemini-local preference already yields parity without touching the OpenAI wire.
4. **Surface the unpaired calls as a user-visible `[Tool …]`/`stderr` diagnostic.** Deferred (D4) — the adapter owns no logging seam; the returned-value accessor is the honest hermetic observability (RF-067-1).
5. **Fix the accounting in the loop.** Rejected (D6) — the concern is the adapter's round building; the loop already carries the id on every result.
6. **Fold this into ADR 0036 instead of a new ADR.** Rejected (D9) — the annotation precedent: 0036 is an `Accepted` record; the new decision is its own.

## Verification

- **Unit** — a pin over the built Vertex request body: (a) a response carrying `functionCall.id` ⇒ the emitted `functionCall`/`functionResponse` parts carry **that** id; (b) a response without a provider id ⇒ the deterministic `call_<n>` fallback (and no empty id key); (c) an `M < N` round ⇒ the produced `M` parts are named correctly **and** `UnpairedCallIDs` returns exactly the unpaired call's id (call order); (d) `M = N` ⇒ `UnpairedCallIDs` empty; (e) the existing byte/shape pins (`TestRequestBody_MediaFreeIsByteIdentical`, the round-065/066 batch + id pins, the OpenAI-compatible byte pins) stay green.
- **E2E** — the round-065/066 `chat/calling-several-tools-in-one-round.feature` journeys are unchanged and stay green (not user-visible); rides `go test` / `make test`.
- **Falsifiability** — removing the provider-id preference reds (a); suppressing the accessor reds (c); a non-deterministic fallback reds (b).
- **Live (non-gating)** — a branch-built binary against a real Vertex provider (the round-065/066 precedent; the RF-066-10 convention — record `go version -m` provenance), performed at the closeout and recorded here.

## References

- `internal/infrastructure/llm/gemini/client.go` (`buildContents`/`roundBuilder`, `parseResponse`) · `internal/infrastructure/llm/openai/client.go` (`tool_call_id`) · `internal/agent/agentloop.go` (replay `call_step_<n>`).
- [ADR 0036](0036-toolcall-id-pairing.md) (§Forward RF-066-2/RF-066-7/RF-066-8) · [ADR 0035](0035-gemini-parallel-tool-call-batching.md) · `specs/plans/067-toolcall-id-followups/` (`spec.md` FR-001…FR-010; `research.md` D1…D8).
- Reference: `tell-me-go/internal/infrastructure/llm/gemini/adapter.go` (`fromSDKFunctionCall`).
- [#136](https://github.com/gosharplite/tellme/issues/136) · [#36](https://github.com/gosharplite/tellme/issues/36) item 3.

## §Forward (deferred, non-blocking)

- **RF-067-1** — the unpaired-call observability is a **returned-value accessor** with **no live consumer**; the boundary drop stays silent at runtime. Surfacing it as a user-visible `[Tool …]`/`stderr` diagnostic is deferred (no adapter logging seam; a scope addition requiring operator confirmation).
- **RF-067-2** — the provider-id preference is **live-unverified** hermetically; a branch-built live Vertex check is a closeout item (record `go version -m`; RF-066-10).
- **RF-067-3** — an E2E carrier for the wire `id` (RF-066-9) stays open; the pins are fixture-based.
- **RF-067-4** — concurrent tool **execution** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) is still not added.
- **RF-067-5** — the reference's fallback spelling is not adopted (recorded divergence, D2).
- **RF-067-6** — a **degenerate duplicate-provider-id** response binds deterministically (first unused identity match, `bind`); carried by a pin and recorded (the spec Edge Case), but a provider that legitimately reuses an id within a round is not otherwise handled.
- **RF-067-7** — a guarding comment/doc on `history.Step`'s **id-less** shape (the invariant D3's locality rests on), so a future `ID` field cannot silently break it.
