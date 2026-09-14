# Phase 0 Research: post-turn status lines — request metrics, cost, and the session summary (round 018)

## Decision 1: Capture cached and reasoning tokens from the provider usage block

- **Decision**: Widen `llm.Usage` with `CachedTokens` and `ThinkingTokens` and parse them from the OpenAI-compatible usage details (`prompt_tokens_details.cached_tokens`, `completion_tokens_details.reasoning_tokens`); the metrics line's `M` is derived as `prompt − cached`. The transport exposes **disjoint, additive** counters — `Th = completion_tokens_details.reasoning_tokens` and `C = completion_tokens − reasoning_tokens` (the OpenAI-compatible wire `completion_tokens` **includes** reasoning) — so `C`/`Th` never double-count and `O = C + Th = completion_tokens`.
- **Rationale**: the metrics line's `H`/`Th` and the cache-hit rate need the provider's cached/reasoning counts, which the round-009 `Usage` (prompt/completion/total only) does not carry; the reference's `tokens.log` records exactly these fields.
- **Alternatives considered**:
  - Estimate cached/thinking locally — impossible: they are provider-side facts.
  - Leave them zero — would make `H`/`Th`/hit-rate always 0, defeating the line.

## Decision 2: Pricing comes only from a config `MODELS` table (no built-in rates)

- **Decision**: add a config `MODELS: { <model>: { PRICING: { HIT, MISS, COMP } } }` map; the effective pricing for the active model is `MODELS[<provider.MODEL>].PRICING`; there is **no** built-in rate table; the table is **config-only** (the nested map is not env-overrideable — the standing `TELL_ME_*` precedence applies to scalar keys only).
- **Rationale**: the operator locked "no built-in defaults" (a wrong built-in rate is worse than a visible `$0.0000`); rates are model-specific and change often; keeping them in config keeps the binary rate-free and the line format stable.
- **Alternatives considered**:
  - Built-in defaults + override (the reference) — rejected: shipped rates rot and mislead.
  - Config-only with env override — rejected: a nested rate map has no clean scalar env form.

## Decision 3: Per-call cost formula and the three-cost semantics

- **Decision**: a call's cost = `(miss·MISS + hit·HIT + (completion + thinking)·COMP) / 1,000,000` USD; `$#1` = the last-returned call; `$#2` = the summed cost of every call in the turn (the prompt completion plus each tool-loop call); `$#3` = the session's cumulative cost. `completion` and `thinking` are the **disjoint** counters from D1 (reasoning billed at the `COMP` rate, matching the reference's separate `response_tokens`/`thinking_tokens`).
- **Rationale**: matches the reference's cached/miss/output billing and the operator's three-slot definition; reasoning is billed as output (the reference convention).
- **Alternatives considered**:
  - Bill reasoning at the MISS rate — rejected: reasoning is output-side.
  - Fewer than three costs — rejected: the operator kept the three-slot shape.

## Decision 4: The agent loop records every call's usage (accumulate + expose)

- **Decision**: widen `AgentLoop` to accumulate every provider call's usage for the turn and to expose the per-call sequence on `AgentResult`; the metrics line and `$#1` use the **last** entry, `$#2` the **sum**, and each call appends one record to the per-mode usage log.
- **Rationale**: today `AgentResult.Usage` carries only the final completion's usage (round 009); the operator's `$#2` (turn total) and the `tokens.log` require every call.
- **Alternatives considered**:
  - Keep only the final usage — cannot produce `$#2`.
  - Re-derive the turn total from the log — the log is per-mode, not per-turn; in-loop accumulation is exact and simpler.

## Decision 5: A per-mode `tokens.log` usage log (one JSON line per API call)

- **Decision**: after each call, append one JSON record — `{timestamp, provider, model, cached_tokens, prompt_tokens, response_tokens, total_tokens, thinking_tokens, cost}` — to `$TELL_ME_HOME/output/<mode>/tokens.log` (the same workspace dir as `history.jsonl`); the session summary reads the whole file; `--new` archives/rotates it alongside `history.jsonl`.
- **Rationale**: the reference's mechanism; one per-mode file = the session, so the summary total is a simple sum; survives resume, resets on `--new`.
- **Alternatives considered**:
  - Add cost/token fields to `history.jsonl` entries — rejected: mixes presentation metrics into the conversation record and complicates the frozen entry shape.
  - A session counter only — rejected: loses per-call detail and the reference shape.

## Decision 6: Render the two lines with a hand-written `internal/ui` formatter

- **Decision**: two formatters alongside `FormatPayloadStatus` — the metrics line `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>` (`Th:` **always**, incl. 0) and the summary `╰─⠿ Ready ($<call> $<turn> $<session> - M: <sM> H: <sH> O: <sO> - <hit%>%)` (costs `$%.4f`, hit-rate `%.1f%%`); plain text, `stderr`, timestamp from the injected clock seam.
- **Rationale**: dependency-free, deterministic, mirrors the round-009/017 formatter pattern (`internal/ui`).
- **Alternatives considered**:
  - Reuse glamour — rejected: these are plain status lines, not Markdown.

## Decision 7: Session totals read the usage log; the turn accumulates in memory

- **Decision**: the session's `M/H/O` totals and hit-rate are the sums over the session's `tokens.log`; the turn's cost is accumulated in-loop during the run.
- **Rationale**: exact, resume-safe, and `--new`-reset; needs no state beyond the log.
- **Alternatives considered**: a persisted running total — redundant with the log.

## Decision 8: Suppress the lines only when the just-returned call reports no usage

- **Decision**: emit both lines on every prompt-bearing turn unless the just-returned call's usage is absent (`Reported == false`), in which case both are skipped; failure paths (provider / tool / history) keep their class phrase + exit code and emit no post-turn line.
- **Rationale**: matches the reference's suppress-when-no-usage and the operator's presence rule; never fabricates zeros for a non-measured turn.
- **Alternatives considered**: always emit zeros — rejected: misleading.

## Decision 9: Verification — unit formatters + E2E via the fake provider; no new dependency; no pty

- **Decision**: pure-helper unit tests for the pricing arithmetic, the two formatters, and the usage-record JSON round-trip; E2E via the existing fake provider extended to report (or withhold) the usage **details** (cached + reasoning) and to script a **multi-call** tool turn; hermetic (injected clock + streams); no new dependency.
- **Rationale**: matches the round-011/014 pattern (unit pure helpers + E2E acceptance), keeps `make verify` dependency-free, and asserts the `$#1 < $#2 ≤ $#3` relationship from the fake's recorded requests.
- **Alternatives considered**: a pty — unnecessary (no TUI surface change this round).

## Must-ask questions (already settled by existing truth — unchanged this round)

- **BDD techstack**: `godog` (existing `specs/truth/techstack.md`).
- **Test strategy**: E2E (black-box) for the acceptance path + pure-helper unit tests.
- **System ends**: a single CLI end.
