# Tasks — Bounded provider transport retry (round 078)

**Plan Package**: `specs/plans/078-provider-transport-retry`
**Inputs**: `spec.md`, `plan.md`, `research.md`, `features/acceptance/…`

> Executed One-Shot (red → green → refactor via `/axb-bdd`). Each task is marked `[X]` after its verification passes. The Test-Alignment layer (Phase 3) carries the interface carrier (the truth Rule/Examples); the Feature phases carry the green/refactor work.

## Phase 1 — Setup & Foundational

- [X] **T001** Add the retryable predicate to the domain: `llm.ProviderError` gains `Status int` / `Transport bool` (`internal/domain/llm/gateway.go`), plus `Retryable(err) bool` (`Transport || Status == 429 || 500≤Status≤599`); the `errors` import.
- [X] **T002** Set the typed facts at the adapters: `openai/client.go` and `gemini/client.go` gain `wrapTransport`/`wrapStatus` and route `Do`/`io.ReadAll` → `Transport`, non-2xx → `Status`, decode/truncation → neither (non-retryable).

## Phase 2 — Foundational (the retry seam)

- [X] **T003** Add the CLI retry decorator `internal/cli/retry_gateway.go`: `retryDelays = {1s, 3s}` (a literal-pinned constant), `retryingGateway` (bounded attempts, `ctx`-aware, non-retryable short-circuit), `realRetrySleep`, `withProviderRetry`, the `TELL_ME_FORCE_RETRY_DELAY_MS` seam (`resolveRetryDelays`), and `retryNotifier` (a chrome-styled `stderr` line, spinner yield/restore, never the class phrase).
- [X] **T004** Wire it into `runTurn` (`internal/cli/cli.go`): build the chain **after** the indicator exists — retry **inner**, `withUnpairedDiagnostic` **outer** — so the retry re-sends the SAME request and the notify closure can yield the live spinner frame.

## Phase 3 — Test Alignment & Implementation

- [X] **T005** (Test Alignment) Extend the truth contract: a new `Rule` + 4 Examples in `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` (drop-once → 2 calls; drop-twice → 3 calls; always-drop → 3 calls + phrase + exit 6; reject → 1 call); the `chat/dsl.md` Given/Then rows (round 078) + the round-030 note reconciliation.
- [X] **T006** (Test Alignment) Extend the hermetic E2E harness: `fakeprovider.Provider.DropFirst(n)` (a transport drop via a hijacked connection, recorded then closed); the scenario default `TELL_ME_FORCE_RETRY_DELAY_MS=0`; the round-078 Givens/Thens (`step_r078_retry_givens.go`, `step_r078_retry_thens.go`).
- [X] **T007** (Green) **US1** — a retryable failure is retried (1s, 3s) and the turn succeeds on a later attempt; the retry line is emitted on `stderr` only.
- [X] **T008** (Green) **US2** — a non-retryable failure (400 / decode / truncation) is attempted once and fails immediately.
- [X] **T009** (Green) The exhaust path fails with `the provider request failed` + exit 6 (unchanged), with exactly 3 attempts.
- [X] **T010** (Green) Accounting: a retried call is one `calls`/usage/frame; a failed attempt writes no history — **witnessed by the E2E Then `the turn was recorded as a single provider call`** (the persisted history line's `calls == 1` **and** exactly one `tokens.log` record), added at fold F-078-2 (the original "proven by the unit pin + existing carriers" claim was false — the unit pins never observe the loop's accounting).
- [X] **T011** (Refactor under green) Extract `arrangeFlakyProvider`; keep the adapters' `wrap`/`wrapTransport`/`wrapStatus` minimal; keep `retryingGateway` a small value type.
- [X] **T012** (Regression) `make verify` + `go test -count=1 ./...` + `make test-race` green; the round-030 truncation Rule stays terminal (never retried); `stdout` byte-exact.

## Unit pins

- [X] **U1** `internal/domain/llm/retryable_test.go` — the predicate truth table (transport; 429/500/503/599; 400/401/404/decode/nil/plain) + `errors.As` unwrap.
- [X] **U2** `internal/cli/retry_gateway_test.go` — the literal `retryDelays` pin (array); drop-once→2; two-drops→3; exhaust→3; non-retryable→1; notify (attempt/total/delay); cancellation-during-wait aborts; parent-cancel → 0 calls; **no announcement when the parent ctx is cancelled mid-call (TD-078-1)**; `realRetrySleep`; `resolveRetryDelays` override; the retry line is plain (no `tellme:` prefix, no class phrase); **the notifier yields→writes→restores around the line (fold F-078-3)**.

## Fold ledger

- **F-078-1** (self-found during implementation, `plan.md` §6 N-1) — the retry line must **not** carry the `tellme: ` prefix (the round-017 *exactly one `tellme:` line* carrier). Folded: chrome-styled line; pinned by U2 + the E2E.
- **F-078-2** (self-found, `plan.md` §6 N-2) — the retry line must **yield/restore** the spinner frame (the round-019 *no residue* carrier). Folded: `retryNotifier(env, ind)` clears+restores; pinned by a unit pin on the `[yield, write, restore]` order (fold F-078-3 re-anchored this from a false "pinned by the E2E" claim).

## Architect review 1 fold (PR #158, `APPROVE WITH REQUIRED FOLDS`; F-078-1…4 + TD-1…3 + RF/N)

- **F-078-1** (record) — the `tellme: ` prefix is removed from the retry-line quote on **all three** record surfaces (ADR D5, `research.md` D5, the `techstack.md` row); the line is documented as `[HH:MM:SS] retrying …` with the round-017 reason inline. Folded.
- **F-078-2** (witness) — the accounting MUST gains a real witness: the E2E Then `the turn was recorded as a single provider call` (`calls == 1` + one usage record); T010's false citation corrected. Folded.
- **F-078-3** (witness + record) — the notifier's yield/restore is now genuinely pinned by a **unit** pin (`retryYield` seam + a `[yield, write, restore]` order assertion); `plan.md` N-2 and ADR RF-078-5 corrected. Folded.
- **F-078-4** (witness) — the retry line gains an E2E carrier (`tellme announces on stderr that it is retrying the provider request` on the `drop once` Example); `plan.md` N-1 corrected. Folded.
- **TD-078-1** (bug) — the decorator now checks `ctx.Err()` **before** announcing a retry, so a SIGINT mid-call cannot emit a spurious "retrying …"; pinned by `TestRetryingGateway_NoAnnounceWhenCancelledMidCall`. Folded.
- **TD-078-2** — the delay seam is now **scoped to the 078 scenarios** (set in the 078 Givens, not globally) and added to the harness `envUnset`; "count/order" corrected to **count** everywhere (the order is a constant-pin). Folded.
- **TD-078-3** — the round-004 "exactly one provider request" note in `chat/dsl.md` is reconciled with a parenthetical pointing at the round-078 note. Folded.
- **RF-078-1** — a shared `newRetryingGateway` constructor now serves production + tests (no drift). Folded.
- **RF-078-2** — `retryDelays` is an **array** (`[2]time.Duration`), not a mutable slice. Folded.
- **RF-078-3** — the `len(delays) == 0` guard is commented as **defensive**. Folded.
- **RF-078-4** — observation only (the split retry policy is deliberate + documented); recorded as `RF-078-9`. Acknowledged.
- **N-078-1** — the `DropFirst` hijack fallback is commented as defensive (unreachable on httptest). Folded.
- **N-078-2** — the exhaustion Example now carries `the run wrote nothing to standard output`. Folded.
- **N-078-3** — the `drop once` Example now carries `the retry re-sent the same request`. Folded.
- **N-078-4** — cosmetic notify wording; **not** folded (the "attempt n of max" reading is intentional — the announcement precedes the wait).
