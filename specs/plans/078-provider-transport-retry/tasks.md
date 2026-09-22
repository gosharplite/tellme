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
- [X] **T010** (Green) Accounting: a retried call is one `calls`/usage/frame; a failed attempt writes no history (the loop persists only a completed turn) — proven by the unit pin (one `Complete` return) + the existing round-027/034 carriers.
- [X] **T011** (Refactor under green) Extract `arrangeFlakyProvider`; keep the adapters' `wrap`/`wrapTransport`/`wrapStatus` minimal; keep `retryingGateway` a small value type.
- [X] **T012** (Regression) `make verify` + `go test -count=1 ./...` + `make test-race` green; the round-030 truncation Rule stays terminal (never retried); `stdout` byte-exact.

## Unit pins

- [X] **U1** `internal/domain/llm/retryable_test.go` — the predicate truth table (transport; 429/500/503/599; 400/401/404/decode/nil/plain) + `errors.As` unwrap.
- [X] **U2** `internal/cli/retry_gateway_test.go` — the literal `retryDelays` pin; drop-once→2; two-drops→3; exhaust→3; non-retryable→1; notify (attempt/total/delay); cancellation-during-wait aborts; parent-cancel → 0 calls; `realRetrySleep`; `resolveRetryDelays` override; the retry line is plain (no `tellme:` prefix, no class phrase).

## Fold ledger

- **F-078-1** (self-found during implementation, `plan.md` §6 N-1) — the retry line must **not** carry the `tellme: ` prefix (the round-017 *exactly one `tellme:` line* carrier). Folded: chrome-styled line; pinned by U2 + the E2E.
- **F-078-2** (self-found, `plan.md` §6 N-2) — the retry line must **yield/restore** the spinner frame (the round-019 *no residue* carrier). Folded: `retryNotifier(env, ind)` clears+restores; pinned by the E2E.
