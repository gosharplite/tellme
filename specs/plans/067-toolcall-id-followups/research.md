# Technical Research — Round 067: Gemini/Vertex tool-call id follow-ups (surface unpaired calls + provider-issued `functionCall.id`)

**Plan package**: `specs/plans/067-toolcall-id-followups`
**Truth owner**: `/axb-technical-research` → updates `specs/truth/techstack.md`
**Anchor**: [#136](https://github.com/gosharplite/tellme/issues/136) (ADR 0036 §Forward **RF-066-2** + **RF-066-7**; retires **RF-066-8**)
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1/US2 · FR-001…FR-010 · SC-001…SC-007 · S-1…S-7 · I-1…I-8 · A1…A7); the grounded current shape (spec § *Grounded*, measured `dev` @ `30f54a1`); `specs/truth/techstack.md`; `tell-me-go` (the reference, read locally).

**Goal**: close [#136](https://github.com/gosharplite/tellme/issues/136) — the small, self-contained continuation of round 066 (**ADR 0036**): (A) **surface the unpaired calls** of a Gemini round instead of the silent drop (**RF-066-7**; retires the `N=2 M=1` residual **RF-066-8**), and (B) **prefer the provider-issued `functionCall.id`** over the synthetic `call_<n>` (**RF-066-2**). A **hardening/parity** round (no user-visible behaviour change).

---

## D1 — Unpaired calls: a single-owned accessor on the round (US1 / RF-066-7)

`roundBuilder` gains a **single-owned accessor** `unpaired() []string` returning, **in call order**, the ids of the round's calls that received no result (`pending[i].used == false`). `flush()` — the boundary where the round-066 drop happens — records those ids on the builder (`b.dropped`) **before** clearing `pending`, so the accounting is done **once**, at the one place the drop occurs, and cannot drift between call sites (FR-002).

The accessor is surfaced as a **returned value**, not a `stderr` diagnostic (S-3 → the returned-value form): a package-level `UnpairedCallIDs(prior []llm.Message) []string` builds the round(s) and returns the accumulated `dropped` ids. **Why not a diagnostic:** the adapter owns no logging/stderr seam today (the CLI owns the chrome), and a `[Tool …]`-class note or a `stderr` write from infrastructure would be a new side-effect that could pollute the E2E stdout/stderr assertions — it is deferred (a forward item) and the returned-value accessor is the honest, hermetic observability. **Why exported (`UnpairedCallIDs`) and not test-only:** it is a legitimate, documented seam for a future diagnostic/consumer, and exporting keeps it out of the `unused` lint class — the `internal/…` package boundary already makes it non-public to the world.

The emitted batched turn is **unchanged**: it still carries the `M` parts the round produced (FR-003, ADR 0035's recorded shape). Round 067 makes the drop **observable**, it does not change the wire.

## D2 — `UnpairedCallIDs` retires the `N=2 M=1` residual (RF-066-8 → closed)

The round-066 residual was that `TestRequestBody_ShortRound_DropsUnpairedNames` (`N=2 M=1`) **cannot kill a partial-drop mutant** — the emitted turn is the same single `functionResponse` part whether the adapter drops **one** or **both** unpaired calls (inherent `M == N/2` arithmetic equivalence). With exact-identity accounting, a pin can now assert the **identity** of the unpaired call: `UnpairedCallIDs(prior)` for a 2-call / 1-result round returns exactly `["call_2"]` (call order). A mutant that drops the **wrong** call — or drops a paired call — changes the returned set and the pin reds. **RF-066-8 is therefore retired** (SC-007): the residual is replaced by a pin over the exact unpaired identity, and the old arithmetic-equivalence gap no longer applies to the accounting claim.

## D3 — Provider-issued id preference (US2 / RF-066-2)

`parseResponse` reads `candidates[0].content.parts[].functionCall.id` (a new `ID string` field on the decode struct's `FunctionCall`) and **prefers** it when non-empty; otherwise the existing deterministic `call_<n>` is used:

```go
func callID(callIdx int, providerID string) string {
    if strings.TrimSpace(providerID) != "" {
        return providerID
    }
    return fmt.Sprintf("call_%d", callIdx)
}
```

- **Why prefer it:** the reference's `fromSDKFunctionCall` reads `f.ID` first; a provider that keys its own bookkeeping on the id it issued must see it echoed. tellme's **OpenAI-compatible adapter already prefers the provider id** (`openai/client.go` sets `ToolCall.ID` from the response), so Gemini is the outlier this round aligns.
- **Empty provider id ⇒ fallback (FR-008).** A blank/whitespace provider id is treated as **absent** (never an empty `id` on the wire — ADR 0036 D2/D4). A provider-id-**absent** response keeps **today's** behaviour exactly (S-6) — **no new failure mode**.
- **Determinism (FR-007 / I-5).** The fallback is the existing positional `call_<n>`, deterministic per turn; a provider id is provider-stable for a given response. Both are stable across builds, so a replayed/resumed body is deterministic (round-014 fidelity holds). The **fallback spelling stays `call_<n>`** (S-4) — the reference's `gemini-call-<index>-<name>` is **not** adopted (a recorded divergence: tellme's existing spelling is already deterministic and in place; changing it would churn the round-066 fixtures for no behavioural gain).

## D4 — Cross-family id scope (S-5): **Gemini-local by construction — the OpenAI-compatible wire is untouched**

The id value flows, at the loop, into `llm.Message.ToolCallID`, which the OpenAI-compatible adapter serializes as `tool_call_id` (`openai/client.go:137`) and the Gemini adapter as `functionResponse.id`. So the concern in #136 ("the id value flows to the OpenAI wire too") is real **for a shared convention**.

It resolves **cleanly as S-5 (i)**, with **no loop/port change**, because **one session drives exactly one family**: the loop is family-blind, and the provider (and therefore the adapter whose `parseResponse` produced the id) is fixed for the session. **The Gemini adapter's `parseResponse` change can only ever feed the Gemini adapter's own next request**; an OpenAI-compatible session's ids come from the **OpenAI adapter's** `parseResponse` (already the provider's own id). Hence **no OpenAI-compatible wire is affected** by this round — **I-1 holds trivially**, and the change is **family-local** (D5). *(Decision recorded, not deferred: the alternative (ii) — a unified id convention — is neither needed nor desirable; it would couple the two families for no gain.)*

## D5 — Scope: family-local, adapter-only

The change is confined to `internal/infrastructure/llm/gemini` (`client.go`: `parseResponse` id provenance + the `roundBuilder` accessor/`consume` factor). The **OpenAI-compatible** adapter, the loop, the ports, the tools, the config, and the persisted records are **untouched** (I-1/D4). No new dependency (`encoding/json`, `fmt` only); POSIX-only; hermetic. The `roundBuilder` refactor factors the message switch into `consume()` so `buildContents` and `UnpairedCallIDs` share one walk and each function stays well under the `cyclop` max (15).

## D6 — Invariants held

- **I-1** OpenAI-compatible wire byte-preserved (D4 — family-locality makes this hold by construction; its tests stay green and unchanged).
- **I-2** the round-066 batched shape is preserved (the id **values** are the only possible delta; the produced `M` parts, call order, and media placement are unchanged).
- **I-3** the media-free Gemini text path is byte-identical (no tool parts ⇒ no id).
- **I-4** no silent media loss — the media-bearing-`tool`-message carriage (F-066-2) is untouched.
- **I-5** replay stays id-primary (the loop's `call_step_<n>` on both sides); the provider-id preference does not touch the replay synthesiser, and the FIFO fallback stays defensive.
- **I-8** an empty id is omitted, a foreign id is omitted (ADR 0036 D2/D4) — unchanged.

## D7 — Verification: unit pins over the built request body + the accessor

- **Primary carrier — a unit pin over the built request body** (the round-065/066 precedent) **plus** a direct accessor pin:
  - (a) a response carrying `functionCall.id` ⇒ the emitted parts carry **that** id (provider preference);
  - (b) a response **without** a provider id ⇒ the deterministic `call_<n>` fallback (and no empty id);
  - (c) an `M < N` round ⇒ the produced `M` parts are named correctly **and** `UnpairedCallIDs` returns exactly the unpaired call's id(s) (call order);
  - (d) `M = N` ⇒ `UnpairedCallIDs` is empty;
  - (e) the existing byte/shape controls (I-1/I-2/I-3/I-8) stay green.
- **E2E** — the round-065/066 journeys are **unchanged** (not user-visible); `/axb-dsl-refine` is **NOOP** (the accessor is a code seam, not a user-visible diagnostic).
- **Falsifiability** — removing the provider-id preference reds (a); suppressing the accessor reds (c); a non-deterministic fallback reds (b).

## D8 — Governance: a new ADR extending ADR 0036

**ADR 0037** records the provider-id preference + the unpaired-call accounting + the cross-family decision (D4). It **extends ADR 0036** and annotates its §Forward **RF-066-2** (delivered), **RF-066-7** (delivered) and **RF-066-8** (retired), plus the README index row. ADR 0036's body is otherwise not edited (the annotation precedent).

---

## Truth impact (semantic units)

| Action | Truth Spec | Summary |
| --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | append the round-067 clause: the response id's **provenance** is the provider's own `functionCall.id` when present, else the deterministic `call_<n>` (OpenAI-compatible adapter already prefers the provider id; a session is single-family so the OpenAI wire is unaffected); the `M < N` boundary drop is **observable** via the single-owned `UnpairedCallIDs` accessor (never only a silent drop). |
| ADD | `docs/decisions/0037-*.md` (+ the `docs/decisions/README.md` index row; an annotation on ADR 0036's `Status` + §Forward RF-066-2/RF-066-7/RF-066-8) | the id-provenance + unpaired-call accounting + cross-family decision. |
| NOOP (checked) | *Image content on the provider wire (Gemini/Vertex)* row | the round adds no shape change to the media/batch wire. |
| NOOP (checked) | the OpenAI-compatible rows | unchanged (byte-frozen; I-1). |

## Residual risks / forward items (non-blocking)

- **RF-067-1** — the unpaired-call observability is a **returned-value accessor**, not a user-visible/`stderr` diagnostic; surfacing it as a diagnostic (a `[Tool …]`-class note) is a forward item (no adapter logging seam today).
- **RF-067-2** — the provider-id preference is **live-unverified** hermetically (the E2E fake scripts no `functionCall.id`), so a branch-built live Vertex check is a closeout item (RF-066-10 convention: record `go version -m` provenance).
- **RF-067-3** — an E2E carrier for the wire `id` (RF-066-9) stays open; the pins remain fixture-based.
- **RF-067-4** — concurrent tool **execution** ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) is still not added; the id axis is merely more faithful.
- **RF-067-5** — the reference's fallback spelling (`gemini-call-<index>-<name>`) is not adopted (recorded divergence, D3).

## References

- `internal/infrastructure/llm/gemini/client.go` (`buildContents`/`roundBuilder`, `parseResponse`) · `internal/infrastructure/llm/openai/client.go` (`tool_call_id`) · `internal/agent/agentloop.go` (replay `call_step_<n>`).
- `docs/decisions/0036-toolcall-id-pairing.md` (D2/D3/D4 + §Forward RF-066-2/RF-066-7/RF-066-8) · `specs/plans/066-toolcall-id-pairing/` (the shape this extends).
- Reference: `tell-me-go/internal/infrastructure/llm/gemini/adapter.go` (`fromSDKFunctionCall` — `f.ID` first).
- [#136](https://github.com/gosharplite/tellme/issues/136) · [#36](https://github.com/gosharplite/tellme/issues/36) item 3.
