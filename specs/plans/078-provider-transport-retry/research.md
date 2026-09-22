# Technical Research: Bounded provider transport retry (round 078)

**Plan Package**: `specs/plans/078-provider-transport-retry`
**Spec**: `specs/plans/078-provider-transport-retry/spec.md`
**Anchor**: operator request (2026-09-22, in-session) — no GitHub issue (rounds 073/074/075 precedent) · **ADR**: [0050](../../decisions/0050-bounded-provider-transport-retry.md)
**Status**: complete (2026-09-22) — decisions locked; the retryability predicate + the schedule are **operator-locked**, the remaining calls are technical.

---

## 0. The problem (grounded, `dev` @ `dd9b94b`)

An operator sees `tellme: the provider request failed: …` (**exit 6**) on a momentary network blip; re-running the same command succeeds. tellme has **no retry layer at all**: one `http.Client.Do`, one failure, terminal.

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/llm/openai/client.go:63-104` | `Complete` sends exactly **one** request; every failure → `c.wrap(err)` → `*llm.ProviderError`. `defaultTimeout = 300s` (line 23). |
| `internal/infrastructure/llm/gemini/client.go:72-114` | identical shape for Vertex/Gemini (`defaultTimeout = 300s`, line 31). |
| `internal/domain/llm/gateway.go:95-112` | `ProviderError{Provider, Err}` — **no status / transport field**; a non-2xx status exists only inside a formatted string (`fmt.Errorf("provider returned status %d: %s", …)`). |
| `internal/agent/agentloop.go` (`Run`) | `resp, err := a.Gateway.Complete(ctx, req)` → on `err != nil` returns it **unchanged**. |
| `internal/cli/cli.go:743` | `gw = withUnpairedDiagnostic(gw, …)` — the **existing gateway-decorator precedent** (`unpaired_gateway.go`, round 068). |
| `internal/cli/cli.go` → `emitProviderError` | `*agentport.ErrIncomplete` → tool phrase + 7; everything else → `the provider request failed` + 6. |

**The honest hazard**: the discriminator for "retryable" (transport vs. status) is **not typed today**. A predicate written now would have to parse the error string — which this round forbids (NFR-003). The round therefore **owes a typed classification** (D2).

---

## 1. Reference comparison (`tell-me-go`)

The reference ships a **4-layer resilience stack** — and it is exactly what tellme deliberately did **not** re-create:

| Layer | Reference (`tell-me-go`) | tellme today |
| --- | --- | --- |
| Transport wrapper | `resilientClient` — 2 attempts, one auth-refresh retry, connection-pool reset | none |
| Cross-provider failover | `FailoverGateway` over an ordered provider chain | none (one selected provider) |
| Engine recovery | `RecoveryStep` + `DefaultRetryPolicy` — **exponential backoff** (2 s / 5 s, ×6, capped 2 min) + jitter; empty-response retry (3) | none |
| User flag | `--retry` (re-run the last message, with confirmation) | settled exclusion (`-b`/`--retry` out of scope) |

**This round adopts a deliberate subset of layer 3 only** — a **fixed, small, non-configurable** backoff (1 s, 3 s; two retries) with **no** failover, **no** auth-refresh, **no** jitter, and **no** user-facing flag. It is **not** a re-creation of the reference's resilience program; it is the minimal fix for the observed symptom, and it is a **recorded divergence** (the reference's exponential/jittered policy is not adopted).

---

## 2. Where the retry must attach (grounded)

- The transport adapters are the only place that **knows** the failure kind (a `c.http.Do` error vs. `resp.StatusCode` vs. a decode error vs. the round-030 truncation). So the **typed facts** must be attached **at the adapter** (D2).
- The retry loop, the sleep, and the operator-visible line are **call-orchestration + presentation** — the CLI already owns a gateway decorator at exactly the right spot (`withUnpairedDiagnostic`, `internal/cli/cli.go:743`) and the diagnostic stream (`env.stderr`) + the injected clock (`env.now`). So the **loop** attaches there (D4).
- **Order matters**: the round-068 unpaired diagnostic inspects the **final** response, so the retry MUST be the **inner** wrapper (retries resolve first, then unpaired inspects) → `withUnpairedDiagnostic(retrying(gw), …)`.

---

## 3. Decisions

**D1 — The retryable class (operator-locked, 2026-09-22).** Retry **transport failures** (dial failure, `EOF`/`unexpected EOF`, connection reset, broken pipe, HTTP/2 `GOAWAY`, TLS failure, request timeout) **and HTTP 429 + 5xx**. Do **not** retry 4xx (400/401/403/404/422…), content-filter rejections, decode/parse errors, or the round-030 output-cap truncation (a successful-but-truncated reply stays terminal). *Rejected:* retry every `*ProviderError` (would re-send doomed 400/401 requests, and re-send on a truncation that cannot change); retry nothing (the defect).

**D2 — Typed classification, owned by the domain; facts supplied by the adapter.** `llm.ProviderError` gains typed **facts** — `Status int` (the HTTP status, 0 when the failure was not an HTTP-status failure) and `Transport bool` (the failure came from the connection, not the status/body/decode). The **predicate** lives in `internal/domain/llm` (e.g. `Retryable(err) bool` = `Transport || Status == 429 || 500 ≤ Status ≤ 599`) so the policy has a single domain owner and the CLI names only domain types. The adapters set the facts where both are known (`Do`/`io.ReadAll` → `Transport`; non-2xx → `Status`; decode/truncation → neither). **String matching is forbidden (NFR-003).** *Rejected:* keeping an untyped `ProviderError` and parsing the message (fragile, forbidden); a boolean `Retryable` baked at the adapter (puts policy in infrastructure, splits the single owner).

**D3 — Schedule and bound: named constants, no config.** `retryDelays = {1s, 3s}` (an array) → **at most 2 retries / 3 attempts**, then the frozen phrase + exit 6. Fixed constants (mirroring round-076's `maxUnknownToolFolds`), pinned by a **literal** unit test (the round-064/076 literal-pin precedent; the *order* 1 s → 3 s is a constant-pin, not an observation). A hermetic seam `TELL_ME_FORCE_RETRY_DELAY_MS` (milliseconds; unset/invalid = the real delays) mirrors `TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS` (`internal/cli/cli.go:1001`, `internal/ui/coordinator.go:44`), **scoped to the 078 E2E scenarios**, so the E2E asserts the request **count**, not wall-clock. *Rejected:* config keys (the operator asked for "simple"; a knob is a forward item); exponential/jitter (the reference's policy — not adopted).

**D4 — The seam: a `llm.Gateway` decorator in `internal/cli`, inner to `withUnpairedDiagnostic`.** Constructed in `runTurn` exactly like `withUnpairedDiagnostic` (D-section above): it needs the diagnostic stream + the clock, both already on `runtimeEnv`; `internal/cli` names only domain types (`llm.Gateway`, `llm.ProviderError`), so the RULE-B baseline stays 0. *Rejected:* a per-adapter retry (duplicated in two transports — violates I-7/FR-006); a composition-root wrap (no access to the CLI's diagnostic-stream seam → the retry line would escape the CLI's captured stderr in unit tests).

**D5 — The retry diagnostic: one plain, non-class line on `stderr` only.** `[HH:MM:SS] retrying the provider request in <delay> (attempt <n> of <max>): <folded detail>` — chrome-styled and deliberately carrying **neither** the `tellme: ` class prefix **nor** the frozen phrase (so `tellme: the provider request failed` appears **exactly once**, on final failure — the round-017 *exactly-one-`tellme:`-line* contract holds, protecting the E2E assertion and operator comprehension). Control-free, best-effort, never on `stdout`, never in `turns.log`. Emitted with the spinner **yielded/restored** through the CLI's existing terminal-safe line/yield seam (ADR 0014/0015) so it cannot tear the spinner frame; TTY-accent-free. *Rejected:* reusing the class phrase (collides with the failure assertion, confuses operators); a `tellme:`-prefixed line (breaks the exactly-one-class-line contract); silence (a 4 s stall with no explanation).

**D6 — Cancellation aborts immediately.** Before each sleep and each attempt the decorator checks `ctx.Err()`; a parent cancellation (SIGINT/SIGTERM) stops the loop at once and returns the (already-provider-classified) failure — never sleeping through, never retrying after, a cancellation (I-3). Note the discriminator: a **client-side 300 s timeout** surfaces as a `*url.Error` with `ctx` **not** cancelled → retryable; a **parent** cancellation has `ctx.Err() != nil` → not retryable.

**D7 — Accounting: a retried call is one AI-endpoint call.** The decorator returns a single `llm.Response` on eventual success, so the loop sees one `Complete` return → one `calls` entry, one `usage` record, one turn frame. The round-034 chrome rule ("an internal retry does not count") is thereby honoured; a failed attempt writes no history. *Rejected:* counting each attempt (would inflate `Σ calls` and the turn header).

**D8 — Records.** A new **ADR 0050** + index; `specs/truth/techstack.md` — **MODIFY** the round-030 row (delete the now-false "tellme adds **no** retry layer" clause and point it at the new retry row), **ADD** a *Provider request retry (transient transport)* row, **MODIFY** the *Provider gateway port* row (`ProviderError` carries typed facts); the CLI feature `reporting-a-failed-provider-request.feature` gains the retry Rule/Examples (via `/axb-dsl-refine`). **Not modelled** (ADR 0041 escape hatch): the change adds no modelled entity/attribute/relationship/invariant — no `docs/domain-model/**` change (reason in `plan.md` §5).

**D9 — Scope.** MCP is **out** (a discovery failure is already warn+skip; an MCP tool-call failure is already a recoverable fold-back — no change). `-b`/`--retry` is unrelated (settled exclusion). No new class phrase; no new exit code (S-10).

---

## 4. Truth & domain-model impact

- `specs/truth/techstack.md` — MODIFY ×2 (the round-030 truncation row; the Provider gateway port row) + ADD ×1 (the retry row). See §7.
- `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` + `chat/dsl.md` — the retry Rule/Examples (owned by `/axb-dsl-refine`).
- `docs/decisions/0050-*.md` + index.
- `docs/domain-model/**` — **unchanged** (D8; not modelled).

---

## 5. Falsifiability (the witnesses the round MUST carry)

| Witness | Mutation (reproduce then revert) | Must redden |
| --- | --- | --- |
| **W1** | neutralize the retry (one attempt, no loop) | the E2E `the endpoint drops the connection once and then answers` — `served == 1`, and the unit pin for the retry count |
| **W2** | remove the bound (retry forever) | the E2E exhausted Example — `served == 3` (it would exceed, or never exit) |
| **W3** | make the predicate always-true | the E2E non-retryable Example (400) — `served == 1` (it would be 3) |
| **W4** | swap a delay constant (1s → 2s) | the **literal** unit pin on `retryDelays` |

(W1–W3 are E2E-observable through the fake provider's `RequestCount()`; W4 is a unit literal pin.)

---

## 6. Residual risks / forward items

- **RF-078-1** — the delays (1 s, 3 s) and the count (2) are constants; making them configurable is out of scope (an operator knob if demand appears).
- **RF-078-2** — **no jitter** (the reference jitters). Deliberate: two fixed delays at two retries carry negligible thundering-herd risk for a single-operator CLI.
- **RF-078-3** — the retry covers the **transport/status** class only; a provider that returns a *successful but unusable* body (empty/garbled) is still terminal (the reference's empty-response retry is a separate, non-existent mechanism).
- **RF-078-4** — worst-case added latency is bounded by 1 s + 3 s (+ up to 3 × the 300 s client timeout) before the failure surfaces; the timeout itself is unchanged.
- **RF-078-5** — **RESOLVED at implementation** (not deferred): the retry line's spinner interaction is now wired (the notifier **yields/restores** the indicator around the line) and pinned by the round-019 *no-residue* E2E carrier. Two carrier contracts also surfaced and were folded (`plan.md` §6): the line must **not** carry the `tellme: ` prefix (the round-017 *exactly one `tellme:` line* carrier), so it is chrome-styled.
- **RF-078-6** — the MCP path is unchanged (D9).
- **RF-078-7** — the retry is provider-path only; the offline readers (`-l`, `-t`, `--tool-usage`, `--new`) never build a gateway and are untouched.
- **RF-078-8** — the retry re-sends the **same** request (the decorator calls `inner.Complete` with the same `llm.Request`); a byte-identity pin across attempts (the fake records bodies) is a small optional add — recorded, not required for delivery.

---

## 7. Techstack truth changes (semantic units)

| Action | Row | Summary |
| --- | --- | --- |
| MODIFY | *Provider output-cap truncation guard* (row 102) | delete the "tellme adds **no** retry layer (there is none to change)" clause; state that the **transport/status** retry (new row) applies to transient connection failures only, and a successful-but-truncated reply stays **terminal** (never retried) |
| MODIFY | *Provider gateway port* (row 94) | `llm.ProviderError` now carries typed facts (`Status`, `Transport`) so retryability is classified **without** string matching; the domain owns the predicate |
| ADD | *Provider request retry (transient transport)* (new row in *Reasoning & Provider Transport*) | bounded automatic retry of a retryable provider failure: transport + 429 + 5xx; fixed delays 1 s → 3 s; ≤2 retries; exhaustion reuses the frozen phrase + exit 6; one AI-endpoint call; `TELL_ME_FORCE_RETRY_DELAY_MS` hermetic seam; a recorded divergence (no failover / auth-refresh / jitter) |
