# System Analysis — Bounded provider transport retry (round 078)

**Plan Package**: `specs/plans/078-provider-transport-retry`
**Anchor**: operator request (2026-09-22, in-session) — no GitHub issue

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The provider-failure surface (a failed request's error class + the retry) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module's *reporting a failed provider request* feature gains a retry Rule/Examples; new Given/Then DSL rows observe the fake provider's request count |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state change (a retried call still writes one usage record / one history entry on success; a failed attempt writes nothing) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (one diagnostic line on `stderr`, no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The typed failure facts + the domain predicate | the implementation | `internal/domain/llm/gateway.go` — `ProviderError` gains `Status`/`Transport` + `Retryable`; the adapters (`openai/client.go`, `gemini/client.go`) set the facts (`Do`/`io.ReadAll` → `Transport`; non-2xx → `Status`) |
| **W2** | The retry loop + the diagnostic line | the implementation | `internal/cli/retry_gateway.go` — a `llm.Gateway` decorator (`retryingGateway`), the fixed `retryDelays` constant, the `TELL_ME_FORCE_RETRY_DELAY_MS` seam, the chrome-styled `stderr` line with the spinner yield/restore |
| **W3** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` — *Provider gateway port* MODIFY, *Provider output-cap truncation guard* MODIFY (reconcile), *Provider request retry* ADD; **ADR 0050** + index |
| **W4** | The executable CLI contract | `/axb-dsl-refine` | a Rule/Examples on `reporting-a-failed-provider-request.feature` + `chat/dsl.md` rows (Givens scripting the fake's drop; Thens for the count + the frozen phrase/exit 6) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` | `features/acceptance/…` (a recovery journey) |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is a **failure-handling** behaviour (retry then fail), so `/axb-spec-by-example` is **NOT** NOOP (a short recovery journey).

## 4. Unchanged surfaces (invariants)

- The frozen class phrase `the provider request failed` and exit code **6** are unchanged; no new phrase, no new code; the vocabulary stays ten (`spec.md` I-1/S-10).
- The retry is bounded (≤2 retries / 3 attempts) (`spec.md` I-2) and cancellation-aware (SIGINT aborts at once; `spec.md` I-3).
- A retried call is one AI-endpoint call — one `calls`/usage/frame, no history write for a failed attempt (`spec.md` I-4/D7).
- A non-retryable failure is attempted once (`spec.md` I-5).
- Retryability is **typed**, never string-matched (`spec.md` NFR-003).
- `stdout` byte-exact; the retry line is `stderr`-only; the MCP path and the offline readers are untouched (`spec.md` I-6/D9).
- One attachment point covers both provider families (`spec.md` I-7/FR-006).

## 5. Domain model (ADR 0041)

**Not modelled, and that is recorded here** (the same-PR rule's escape hatch): the round changes an **error-handling policy** (retry then fail) on the already-modelled `Provider` failure surface. It introduces no modelled entity, attribute, relationship, invariant, or scenario — the provider failure is not a modelled entity, and the retry adds no persisted state. `docs/domain-model/**` is **unchanged** and `modelith-check` stays green. (`techstack.md`'s new *Provider request retry* row is the truth home.)

## 6. Behavioural notes locked during implementation (the record)

Two carrier contracts surfaced while wiring the retry line; both are now satisfied and pinned:

- **N-1 — the retry line must NOT carry the `tellme: ` prefix.** The round-017 *failed provider request* carrier asserts **exactly one** `tellme: `-prefixed line (the class phrase). A retry line prefixed `tellme: ` would break it, so the retry line is **chrome-styled** (`[HH:MM:SS] retrying …`) — consistent with the turn chrome (which carries no `tellme:` prefix) and with the class-phrase vocabulary (`dsl.md`). Pinned by the unit test + the E2E.
- **N-2 — the retry line must YIELD/RESTORE the spinner frame.** The round-019 *no progress spinner residue* carrier would otherwise see an un-cleared frame (the retry line written while the frame is live). The notifier clears the indicator (ADR 0014's yield route) before the line and restores it after, so the frame never survives into the line. Pinned by the E2E.
