# Research — Gemini/Vertex cached-token usage (round 072)

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`
**Anchor**: issue [#149](https://github.com/gosharplite/tellme/issues/149).
**Method**: read the tellme Gemini adapter + the reference (`tell-me-go/internal/infrastructure/llm/gemini/metrics.go`) + the OpenAI-compatible sibling, and decide the two open mechanics (the `usageMetadata` field names; the `thoughtsTokenCount` disjointness rule).

## Decisions

| # | Decision | Rationale |
| --- | --- | --- |
| **D1** | **Decode `cachedContentTokenCount`** into `llm.Usage.CachedTokens`. | The provider reports the reused-conversation portion under this key (proto-JSON); the reference maps exactly it (`metrics.go:19`), and the OpenAI-compatible sibling maps its analogue (`prompt_tokens_details.cached_tokens`). |
| **D2** | **Decode `thoughtsTokenCount`** into `llm.Usage.ThinkingTokens` (zero-suppressed downstream, as today). | `dev` declares `THINKING_LEVEL: HIGH`; the field is the provider's reasoning count (`metrics.go:26-27`). |
| **D3** | **Disjointness = DISJOINT (no subtraction)** for the Gemini family. Gemini's `candidatesTokenCount` **excludes** `thoughtsTokenCount` (the provider's identity is `totalTokenCount = promptTokenCount + candidatesTokenCount + thoughtsTokenCount`), so `C` and `Th` are reported as-is and are additive. This is the **opposite** of the OpenAI-compatible family, where the wire `completion_tokens` **includes** `reasoning_tokens` and the adapter therefore stores `max(0, completion − reasoning)` (round-018 FR-002). **The rule is family-specific by necessity** — do not unify them. | The reference does exactly this (`ResponseTokens = CandidatesTokenCount`, `ThinkingTokens = ThoughtsTokenCount`, no subtraction). Unifying the two families would either double-count Gemini thinking or drop OpenAI-visible output. |
| **D4** | **Floor every decoded count at 0** (defensive) and leave a missing field at 0. | Mirrors the OpenAI adapter's `max(0, …)` precedent; a malformed/absent field must not yield a negative miss (`miss = prompt − cached`). |
| **D5** | **No change to the cost formula.** `ComputeCost` + `miss = prompt − cached` are reused; correctness comes from populating `CachedTokens`. | The formula is correct; the input was wrong (round-018 shape unchanged). |
| **D6** | **Family-local**: only `internal/infrastructure/llm/gemini/**` changes (plus the truth row + the ADR + tests). The request body, the emitted turn, and the OpenAI-compatible wire are **byte-identical**. | `spec.md` I-1; the defect is a decode omission. |
| **D7** | **Record in a new ADR 0044** (the Gemini usage decode + the family-specific disjointness rule). | Durable decision; `verify-adr-index` gate. |

## Field-name verification (proto-JSON / SDK)

The `generateContent` response's `usageMetadata` object uses **camelCase proto-JSON** keys — the same casing the adapter already uses for its three existing fields:

| Key | Meaning |
| --- | --- |
| `promptTokenCount` | input tokens (already decoded) |
| `candidatesTokenCount` | visible output tokens (already decoded; **excludes** thoughts) |
| `totalTokenCount` | `prompt + candidates + thoughts` (already decoded) |
| **`cachedContentTokenCount`** | **the reused/cached portion of the input (the missing field)** |
| **`thoughtsTokenCount`** | **the reasoning tokens (the missing field)** |

## The exact defect (measured, `dev` @ `5b1bfa1`)

- `internal/infrastructure/llm/gemini/client.go:532-535` decodes only 3 fields.
- `:574-580` builds `llm.Usage` without `CachedTokens`/`ThinkingTokens`.
- `internal/cli/call_renderer.go:205` → `miss = prompt − 0 = prompt`; `ComputeCost` charges the whole prompt at `MISS`.
- Config `gemini-3.8-flash`: `HIT 0.075 / MISS 0.75 / COMP 3.75` → **10× over-charge**. Operator witness: `M: 390564 H: 0 C: 100 ⇒ $0.2933` (vs ~$0.03 with a realistic cache ratio).

## Risks / mitigations

- **Double-counting thinking tokens** — mitigated by D3 (disjoint, pinned by a unit test that asserts `C` and `Th` are the provider's raw figures and `total` matches `prompt + candidates + thoughts`).
- **A provider that omits the cached field** — D4/D5: stays 0 (a legal cold cache), `Reported` gate unchanged.
- **Regression on the OpenAI-compatible path** — none expected (no shared code); pinned by the existing OpenAI usage tests staying green.

## Verification intent (witnesses)

- **Unit**: a `generateContent` response carrying `cachedContentTokenCount`/`thoughtsTokenCount` decodes into `Usage.CachedTokens`/`Usage.ThinkingTokens`; a response omitting them leaves zeros (no negative miss).
- **E2E**: the fake Vertex provider scripts `usageMetadata`; the run's metrics line shows the cached count (not `H: 0`) and the `Ready` totals agree.
- **Falsifiability**: revert the added field mapping ⇒ the unit pin + the E2E Example RED.

## Truth / decisions

- **ADR 0044** (`docs/decisions/0044-gemini-cached-token-usage.md` + index).
- `specs/truth/techstack.md` — MODIFY the *Gemini/Vertex adapter* + the usage/`UsageRecord` accounting rows.
- `specs/truth/features/cli/chat/**` (+ `chat/dsl.md`) — an ADD/MODIFY journey for the metrics figures.
- No `contracts/**`; `data/**` unchanged (the `UsageRecord` **shape** is unchanged — only its values).
