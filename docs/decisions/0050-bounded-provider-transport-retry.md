# ADR 0050 — Bounded provider transport retry (transient failures)

**Status**: Accepted (round 078)

**Date**: 2026-09-22

**Related**: operator request (2026-09-22, in-session) · round 004 (the provider gateway port + the frozen provider phrase / exit 6) · round 030 (the output-cap truncation guard — the `no retry layer` claim this ADR reconciles) · round 034 (the per-call turn frame + "an internal retry does not count") · round 068 / [ADR 0038](0038-unpaired-call-diagnostic.md) (the gateway-decorator precedent) · round 070 / [ADR 0040](0040-media-channel-in-band.md) · `tell-me-go`'s `resilientClient` / `FailoverGateway` / `DefaultRetryPolicy` (the reference resilience stack — deliberately not re-created) · `specs/truth/techstack.md` §Reasoning & Provider Transport · `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature`.

## Context

tellme sends **exactly one** provider request per inference round (`internal/infrastructure/llm/openai/client.go`, `internal/infrastructure/llm/gemini/client.go`); any failure is wrapped as `*llm.ProviderError` and returned unchanged by the loop, and the CLI maps it to the frozen class phrase `tellme: the provider request failed: …` + exit code **6**. There is **no retry layer** — nothing absorbs a momentary network blip, so the operator must re-run the command. Observed: a connection drop by the remote server aborts the turn, while a re-run usually succeeds.

tellme's error type carries **no typed failure facts**: `llm.ProviderError{Provider, Err}`, with an HTTP status living only inside a formatted message (`fmt.Errorf("provider returned status %d: %s", …)`). The reference's resilience stack (`resilientClient` 2-attempt auth retry + connection reset, `FailoverGateway` cross-provider iteration, `RecoveryStep` + `DefaultRetryPolicy` exponential backoff with jitter, plus a `--retry` flag) was **deliberately not re-created** (a settled scope decision — cf. the no-security-layer / no-runaway-detection directions).

The operator asked for a **simple, bounded** fix: *wait 1 s → retry → wait 3 s → retry → fail* — retrying only **transient** failures, so a real error is not hidden behind pointless waiting.

## Decision

1. **A retryable provider failure is retried, bounded, before failing (D1).** Retryable = a **transport** failure (dial failure, `EOF`/`unexpected EOF`, connection reset, broken pipe, HTTP/2 `GOAWAY`, TLS failure, request timeout) **or** HTTP **429**/**5xx**. Non-retryable = 4xx, auth, content-filter rejection, decode/parse error, and the round-030 output-cap truncation. *Rejected:* retrying every `*ProviderError` (re-sends doomed 400/401 requests; re-sends an unchangeable truncation); retrying nothing (the defect).

2. **Typed classification: facts at the adapter, the predicate in the domain (D2).** `llm.ProviderError` gains `Status int` and `Transport bool`; the adapters set them where the reason is known (`Do`/`io.ReadAll` → `Transport`; non-2xx → `Status`; decode/truncation → neither). The predicate `llm.Retryable(err)` (`Transport || Status == 429 || 500 ≤ Status ≤ 599`) lives in `internal/domain/llm`. **No string matching** (NFR-003). *Rejected:* parsing the message; baking a `Retryable bool` at the adapter (policy in infrastructure).

3. **The schedule and bound are fixed constants, not config (D3).** `retryDelays = {1 s, 3 s}` → at most **2 retries / 3 attempts**; then the frozen phrase + exit **6**. Pinned by a **literal** unit test. A hermetic seam `TELL_ME_FORCE_RETRY_DELAY_MS` (ms; unset/invalid = the real delays) mirrors `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS` so the E2E asserts **count/order**, not wall-clock. *Rejected:* config keys (out of scope for "simple"); exponential backoff + jitter (the reference's policy — not adopted).

4. **The seam is a `llm.Gateway` decorator in `internal/cli`, inner to `withUnpairedDiagnostic` (D4).** Constructed in `runTurn` like the round-068 unpaired wrapper (it needs the diagnostic stream + the clock, both on `runtimeEnv`; `internal/cli` names only domain types, so the RULE-B baseline stays 0). Ordering: `withUnpairedDiagnostic(retrying(gw), …)` — the unpaired check must inspect the **post-retry** response. *Rejected:* a per-adapter retry (duplicated across two transports); a composition-root wrap (no access to the CLI's diagnostic-stream seam).

5. **The retry is announced once per retry, on `stderr` only, and never with the class phrase (D5).** `tellme: retrying the provider request in <delay> (attempt <n> of <max>): <folded detail>` — control-free, best-effort, emitted through the CLI's terminal-safe line/yield seam (ADR 0014/0015) so it cannot tear the spinner frame; never on `stdout`, never in `turns.log`. Deliberately **not** the frozen phrase, so `the provider request failed` appears exactly once — on final failure. *Rejected:* reusing the class phrase; silence.

6. **Cancellation aborts immediately (D6).** The decorator checks `ctx.Err()` before each sleep/attempt; a parent cancellation (SIGINT/SIGTERM) stops at once. A **client-side** 300 s timeout (`*url.Error` with `ctx` **not** cancelled) is retryable; a **parent** cancellation (`ctx.Err() != nil`) is not.

7. **A retried call is one AI-endpoint call (D7).** The decorator returns a single `llm.Response` on eventual success → one `calls` entry, one `usage` record, one turn frame; a failed attempt writes no history. This honours round 034's "an internal retry does not count".

8. **Records; not modelled (D8).** ADR 0050 + index; `techstack.md` MODIFY ×2 + ADD ×1; the CLI feature + DSL rows. **The change is not modelled** (ADR 0041 escape hatch): it introduces no modelled entity/attribute/relationship/invariant — `docs/domain-model/**` is unchanged (reason in `plan.md` §5).

9. **Scope: provider path only (D9).** MCP is out (discovery is warn+skip; an MCP tool-call failure is already a recoverable fold-back). `-b`/`--retry` is unrelated (settled exclusion). No new class phrase; no new exit code.

## Consequences

- A momentary connection drop no longer costs a re-run: tellme retries the same request up to twice (1 s, 3 s) and, on eventual success, completes normally.
- The final failure surface is **byte-identical** to today: `the provider request failed` + exit 6, `stdout` empty; the class-phrase vocabulary and the exit-code set are unchanged.
- A non-retryable failure (4xx/auth/decode/truncation) fails **immediately**, with **one** request — no hidden waits.
- A worst case adds 1 s + 3 s before the failure surfaces; the retry is bounded and cancellation-aware (SIGINT never sleeps through).
- The turn counter, the usage log, the history, and the per-call frame are unaffected by a retry (one call).
- Both provider families are covered by **one** seam; the adapters gain only typed facts (no behavioural change); MCP and the offline readers are untouched.
- **Reconciliation:** the round-030 row's "tellme adds **no** retry layer (there is none to change)" is now false and is **MODIFY**-ed to state the retry's scope (transport/status only; a truncated reply stays terminal).

## Forward

- **RF-078-1** — the delays (1 s, 3 s) and the retry count (2) are constants; a config knob is out of scope.
- **RF-078-2** — no jitter (a deliberate divergence from the reference's jittered backoff).
- **RF-078-3** — the retry covers transport/status only; a successful-but-unusable body (empty/garbled) stays terminal (the reference's empty-response retry is not re-created).
- **RF-078-4** — added worst-case latency is 1 s + 3 s (+ up to 3 × the unchanged 300 s client timeout).
- **RF-078-5** — the retry line's exact spinner yield/restore wiring is pinned at the implementation tier (the residual).
- **RF-078-6** — the MCP path is unchanged.
- **RF-078-7** — the offline reader paths are untouched.
- **RF-078-8** — a byte-identity pin proving each attempt re-sends the same request (the fake records bodies).
