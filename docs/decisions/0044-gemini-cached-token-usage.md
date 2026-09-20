# ADR 0044 — Gemini/Vertex usage decodes the cached-token and thinking-token counts

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0015](0015-loop-presentation-port.md) (loop presentation port — the metrics line is a presentation surface fed by `llm.Usage`), round 018 (`specs/plans/018-*` — the usage/cost accounting + the OpenAI family's exclusive-completion rule), round 013 (`specs/plans/013-*` — the Vertex/Gemini adapter), round 072 (`specs/plans/072-gemini-cached-token-usage` — this ADR's round), issue [#149](https://github.com/gosharplite/tellme/issues/149)

## Context

The Gemini/Vertex adapter (`internal/infrastructure/llm/gemini/client.go`) decoded only three fields of the response `usageMetadata` — `promptTokenCount`, `candidatesTokenCount`, `totalTokenCount` — and built `llm.Usage` without `CachedTokens` or `ThinkingTokens`. Downstream, the miss is `prompt − cached` and `ComputeCost` bills `miss·Miss + hit·Hit + (completion+thinking)·Comp`; with `cached == 0` the **entire prompt** was charged at the **miss** rate on every Gemini turn.

With the config's `gemini-3.8-flash` rates (`HIT 0.075 / MISS 0.75 / COMP 3.75` — MISS is **10×** HIT) this inflated the reported cost by up to **~10×** and made the metrics line read `H: 0` always. The OpenAI-compatible sibling already parsed its analogue (`prompt_tokens_details.cached_tokens`, `completion_tokens_details.reasoning_tokens`), so the two families disagreed — a reachable, expensive, silent defect (issue [#149](https://github.com/gosharplite/tellme/issues/149)).

The reference (`tell-me-go/internal/infrastructure/llm/gemini/metrics.go:17-28`) maps `CachedContentTokenCount` and `ThoughtsTokenCount` and treats them as **disjoint** from `CandidatesTokenCount`.

## Decision

**D1 — The Gemini adapter decodes `cachedContentTokenCount` into `llm.Usage.CachedTokens`.**

**D2 — It decodes `thoughtsTokenCount` into `llm.Usage.ThinkingTokens`.** Both are zero-suppressed by the existing presentation path when `0`.

**D3 — Disjointness is DISJOINT (no subtraction) for the Gemini family, and this is deliberately family-specific.** Gemini's `candidatesTokenCount` **excludes** `thoughtsTokenCount` (the provider's documented identity — a *reading*, not a measured round invariant; the wire `totalTokenCount` has no downstream consumer), so `C` and `Th` are the provider's raw figures and are additive (`O = C + Th`). This is the **opposite** of the OpenAI-compatible family, where the wire `completion_tokens` **includes** `reasoning_tokens`, so that adapter stores `max(0, completion − reasoning)` (round-018 FR-002). The two rules must **not** be unified: unifying them would either double-count Gemini thinking tokens or drop OpenAI-visible output.

**D4 — Every decoded count is floored at 0, and the cached count is additionally CAPPED at the prompt count.** Mirrors the OpenAI adapter's `max(0, …)` precedent; the cap (review B1 / **TD-072-1**) guarantees the miss (`prompt − cached`) can never go negative and the hit-rate can never exceed 100 %, even if a provider reports `cachedContentTokenCount > promptTokenCount`.

**D5 — The cost formula is unchanged.** `ComputeCost` and `miss = prompt − cached` are reused; the fix restores the correct **input** to a correct formula.

**D6 — Family-local.** Only `internal/infrastructure/llm/gemini/**` changes (plus the truth row, this ADR, and tests). The request body, the emitted turn, and the OpenAI-compatible wire are byte-identical. No new dependency; stdlib-only.

**D7 — The four accounting surfaces stay single-sourced.** The metrics line, the `Ready` summary totals, and the persisted `UsageRecord` (`tokens.log`) all derive from the one `llm.Usage` value, so a correct decode fixes all of them at once.

## Consequences

- The fix can **raise** the reported cost on a thinking turn: `Th` (newly decoded) is billed at the `COMP` rate. That is correct — Vertex bills thoughts as output, and it matches the reference — but the round is not uniformly "cheaper": it moves cost from the `MISS` input rate onto the true `HIT` input rate **and** the thought output rate (review NIT-072-2).

- A Gemini/Vertex turn now reports real `H`/`Th` figures; a reused prefix is billed at the HIT rate and the miss shrinks to the genuinely-new input.
- The `tokens.log` records and the session `H`/`M` totals become correct for the Gemini family.
- The family split in the disjointness rule becomes an explicit, documented invariant rather than an accident of “no code”: a future adapter must declare which side it is on.
- **Recorded divergence from the reference:** the field **mapping** matches `tell-me-go` exactly (`metrics.go:19-27`); the two deliberate divergences are (a) tellme **floors** every count at 0 and (b) tellme additionally **caps** the cached count at the prompt count (D4) — the reference assigns `CachedContentTokenCount` verbatim, so a provider reporting `cached > prompt` would give it a negative miss. Both are safety refinements, not behavioural mismatches on a well-formed response.

## Verification

- **Unit**: a `generateContent` response with `usageMetadata.promptTokenCount/cachedContentTokenCount/candidatesTokenCount/thoughtsTokenCount` decodes into `Usage{CachedTokens, ThinkingTokens, …}`; `CandidatesTokenCount` and `ThoughtsTokenCount` are preserved **verbatim** (disjoint); an omitted/absent field stays 0 and produces no negative miss.
- **E2E**: the fake Vertex provider scripts `usageMetadata`; the run's metrics line shows the cached figure (not `H: 0`) — the assertion. The Example is now **priced** and asserts that the cost/`Ready` status is emitted (the numeric cost values remain a *derived* consequence of the unchanged, separately-pinned `ComputeCost`/`HitRate` arithmetic, not independently E2E-asserted — review B1 / **TD-072-2**).
- **Falsifiability**: reverting the field mapping REDs the unit pin **and** the E2E Example.
- `make verify` green (layer gate 0 · `modelith-check` · `verify-fmt` · `verify-adr-index` · lint 0 · govulncheck); `go.mod`/`go.sum` unchanged; the OpenAI-compatible wire byte-identical.

## Forward

- **RF-072-1** — the exact Vertex `thinkingLevel`/`thinkingBudget` semantics (the config sets both) is unmodelled here; the round only reads the **reported** token count.
- **RF-072-2** — a cache-hit-ratio *display* or cost breakdown remains out of scope (only the mis-decode is fixed).
- **RF-072-3** — an explicit per-family usage-decode capability (so a third family cannot silently omit a field) is a candidate refactor, not taken here.
