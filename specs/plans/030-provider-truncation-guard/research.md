# Phase 0 Research: provider-transport truncation guard (round 030)

**Topic**: how tellme stops a provider from silently truncating a response at its output cap — reading the **finish reason** in **both** transports and turning a `length` / `MAX_TOKENS` finish into a **loud provider failure** (never a returned response). Each decision supports `spec.md` (US1/US2 · FR-001..FR-010) and the operator-locked clarifications (**Q1 → universal trigger**; **Q2 → reuse the provider class**).

## Decision 1: The guard lives in the transports — read the finish reason at decode

- **Decision**: read the finish reason inside each adapter's response decode — OpenAI-compatible `choices[0].finish_reason`; Vertex/Gemini `candidates[0].finishReason`. On an output-cap truncation the adapter returns a `*llm.ProviderError` **instead of** the `llm.Response`. No new domain type, no loop change.
- **Rationale**: the finish reason exists only on the wire, so the adapter is the sole place that sees it. `*llm.ProviderError` already maps to the frozen class phrase + exit 6, so the loud failure reuses an existing contract rather than inventing one — the minimal transport-layer change issue #62 asks for.
- **Alternatives considered**:
  - Expose a finish reason on `llm.Response` and fail in the loop — rejected: spreads a transport concern into the agent loop and adds a second decision site for no gain.
  - A dedicated new error type — rejected: `*llm.ProviderError` already carries the provider name + cause and is the CLI's mapped failure surface.
- **Recorded consequence (TD-1)**: a truncation failure is a **failed** turn, so the call's decoded `usage` is **discarded** — the loop records usage only for a **completed** call (round 018), so a truncated call contributes **no** token/cost record. The reference returns `(content, metrics, err)` and still accounts the call; tellme keeps the loss deliberately (carrying usage on the error path would need a `Gateway`/loop change out of this round's scope). Recorded in `techstack.md` so it is a conscious decision, not an unknown.

## Decision 2: The trigger is universal, and only the two truncation values

- **Decision**: fire on **any** finish reason that denotes an output-cap truncation — `"length"` (OpenAI-compatible), `"MAX_TOKENS"` (Vertex/Gemini) — whether the response carries a tool call **or only text** (Q1 → universal). Do **not** fire on `"stop"` / `"tool_calls"` (OpenAI-compatible), `"STOP"` (Vertex/Gemini), or an **absent/empty** finish reason. Other finish reasons (`SAFETY`, `RECITATION`, `MALFORMED_FUNCTION_CALL`, `content_filter`) are **out of scope** (TD-4 — `MALFORMED_FUNCTION_CALL` named explicitly; a tracked forward item, not silently ignored).
- **Rationale**: mirrors the reference (its `checkGeminiTruncation` is universal and pins a text-only case; its OpenAI check fires on any `length`), and it is the conservative reading — a cut-off answer can mislead, not only a cut-off tool call. The rule is false-positive-free on the healthy values, including Gemini's commonly-absent `finishReason`.
- **Alternatives considered**:
  - Tool-call-only — rejected (Q1 → universal): it misses a cut-off answer, and the reference is universal.
  - Fire on every non-`stop` finish reason — rejected: conflates truncation with content filtering/safety, which need different handling and are out of scope.

## Decision 3: Reuse the provider class — no new phrase, no new exit code

- **Decision**: the truncation failure is a `*llm.ProviderError`, surfaced by the CLI as the frozen phrase `tellme: the provider request failed: <detail>` with exit code **6** (Q2). The `<detail>` names the truncation (and the tool when a `functionCall` is the truncation site). No new class phrase; no new exit code; the frozen vocabulary stays at **eleven** phrases.
- **Rationale**: minimal and consistent; the issue itself calls it a "provider error". Operationally a truncation **is** a provider-response failure, and tellme's operator contract already buckets provider/transport failures under one phrase + code 6.
- **Alternatives considered**:
  - A new class phrase + a new exit code (e.g. `the provider response was truncated`, exit 8) — rejected (Q2): grows the frozen vocabulary and the exit-code contract for a failure that is, operationally, a provider failure.

## Decision 4: The failure is terminal by construction — no retry layer is added

- **Decision**: the truncation error simply fails the turn; this round adds **no** retry/failover/classification layer. The reference's "terminal / never auto-retried" wording exists because the reference *has* a `Classify`/resilient client; tellme does not, so there is nothing to retry and no loop to break.
- **Rationale**: tellme has no retry today (verified: zero `Classify`/`retry`/`failover`/`backoff` hits in `internal/`); adding one is far out of scope and unsafe — retrying with the same budget loops forever, which is exactly the reference's scar tissue. Re-running with a larger `MAX_TOKENS` is the operator's call.
- **Alternatives considered**:
  - Add a retry-with-bigger-budget layer — rejected: out of scope and a behaviour change; a silent retry is worse than a loud stop.
  - Say nothing about retries — rejected: the spec/truth must state the posture so a future retry layer cannot silently swallow this failure.

## Decision 5: The request side is unchanged; the unset-`MAX_TOKENS` case is documented

- **Decision**: neither adapter changes what it **sends**. When `MAX_TOKENS` is unset tellme sends **no** output cap and the provider's own default governs; the guard still fires on whatever truncation the provider reports. **No** default cap is added this round.
- **Rationale**: the issue asks only for the documented sentence; adding a default cap is a separate, request-side decision that could surprise operators. The guard's job is to **detect** truncation, not to **prevent** it.
- **Alternatives considered**:
  - Add a generous default cap (reference parity with Anthropic's raised 16384 default) — rejected for this round: a request-side change beyond issue #62; recorded as a possible future hardening.
  - Omit the unset case — rejected: FR-009 / the spec assumption require the honest statement.

## Decision 6: Guard ordering — the truncation error wins over the generic "no usable answer"

- **Decision**: in each adapter the truncation check runs on the decoded response and its error **takes precedence** over the existing generic `provider response carried no usable answer` error, so a truncated response that is also empty is reported as a **truncation**. The OpenAI check runs after the decode regardless of whether tool-call args parse — the adapter stores args as raw strings, so today a truncated args string does **not** error on its own; the finish-reason guard is the residual-class catcher. The Gemini check is **function-call-aware**: when a `functionCall` part is present it names the tool, else it emits a generic truncation message.
- **Rationale**: a truncation-specific message is actionable (raise the budget / break the call up); the generic "no usable answer" would hide the true cause. Mirrors the reference's function-call-aware diagnostic.
- **Recorded divergence (TD-3)**: tellme runs the truncation check **before** the generic `no usable answer` path, so an *empty + truncated* response is reported as a **truncation**; the reference runs its empty-content path (`checkResponse` → `handleEmptyContent`) **before** `checkGeminiTruncation`, reporting `empty response (Finish Reason: MAX_TOKENS)`. Both fail loud (same outcome); the inversion is deliberate and recorded so "mirrors the reference" is not read as ordering parity.
- **Alternatives considered**:
  - Reuse the generic "no usable answer" error — rejected: loses the truncation cause the operator needs.
  - Only check when args fail to parse — rejected: misses the residual well-formed-but-incomplete case the issue explicitly calls out.

## Decision 7: No new dependency; hermetic verification; falsifiability witnesses

- **Decision**: stdlib only; POSIX-only; no new module. **Unit tests** pin each adapter's decode guard (truncated vs healthy, both families — `httptest`/the fake scripts the OpenAI `choices[0].finish_reason` and Vertex `candidates[0].finishReason` shapes). **E2E** drives the built binary with the fake provider scripting a truncated response (tool-call **and** text-only), asserting the refusal + exit 6 and the file/answer effects; no pty. **Falsifiability witnesses**: remove the guard → the truncated-tool-call scenario silently returns/writes; remove a negative-control branch → a healthy response falsely fails.
- **Rationale**: the standing scope (stdlib, POSIX, offline) and the round-008/013/021/029 harness precedent (the fake already serves both wire shapes). Witnesses keep the guard falsifiable at both layers.
- **Alternatives considered**:
  - A third-party JSON/serializer helper — rejected: nothing beyond `encoding/json` is needed.
  - Assert only at the unit layer — rejected: the E2E acceptance (the file not written, the answer not printed) is the PM-visible value and must be witnessed end-to-end.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: the OpenAI-compatible adapter and Vertex/Gemini adapter rows note the finish-reason read; a new **Provider output-cap truncation guard** row records the cross-cutting contract; the response-normalization and provider-gateway-port rows are updated; the write-filesystem-tools row's "#62 forward item" sentence becomes "delivered in round 030".
- `specs/truth/features/cli/**` → **ADD** an interface feature + `dsl.md` rows (owned by `/axb-dsl-refine`).
- `specs/truth/contracts/**` and `specs/truth/data/**` → **NOOP**.
