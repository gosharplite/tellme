# Tasks — Roll back the last N turns of the session history (`-b`/`--back`) (round 081)

**Plan Package**: `specs/plans/081-back-rollback-turns`
**Inputs**: `spec.md`, `plan.md`, `research.md`, `features/acceptance/…`

> Executed One-Shot (red → green → refactor). Each task is marked `[X]` after its verification passes. The Test-Alignment layer (Phase 3) carries the interface carrier (the truth Rule/Examples); the Feature phases carry the green/refactor work.

## Phase 1 — Setup & Foundational

- [X] **T001** Extend the domain port `internal/domain/history.Store` with `Rollback(n int) (removed int, err error)` (durable, clamp, archive untouched) and document it.
- [X] **T002** Implement `fileStore.Rollback` in `internal/infrastructure/history/file_store.go`: load the active entries; `removed = min(n, len)`; write the survivors to a same-dir temp file + `fsync` + `os.Rename` over `history.jsonl` (never truncate in place); a missing file ⇒ 0 removed; a decode failure ⇒ an error (no partial write).

## Phase 2 — Foundational (the CLI seam)

- [X] **T003** Register `-b`/`--back [N]` in `internal/cli/cli.go` (`IntVarP`, `NoOptDefVal = "1"`) and generalise the round-054 optional-int argv pre-pass (`consumeListValue` → a shared helper covering `-l`/`--list` and `-b`/`--back`) — resolving ADR 0023 RF-54-4.
- [X] **T004** Add `renderRollback(...)` (reusing `resolveWorkspace`) and wire `-b` into `dispatchReporting` **after `-l`** and **before `-t`**: a standalone `-b`/`-b N` rolls back, prints the plain confirmation to `stdout`, exits 0; `-b 0`/`-b -1` → usage error; `-b` + `--new` (no prompt) → usage error.
- [X] **T005** Wire the prompt-bearing form: when `-b [N]` is followed by a positional prompt, roll back **first**, then run the prompt through `renderTurn` (the rollback happens before the prompt phase).

## Phase 3 — Test Alignment & Implementation

- [X] **T006** (Test Alignment) Write the truth contract: the NEW `specs/truth/features/cli/history/rolling-back-the-session-history.feature` Rules/Examples; the `history/dsl.md` Given/When/Then rows + the round-081 note; the root `cli/dsl.md` offline-path scope.
- [X] **T007** (Test Alignment) Implement the step definitions (`tests/e2e/steps/step_r081_rollback.go`): the new Givens/Whens/Thens; reuse the shared `readSessionEntries`.
- [X] **T008** (Green) **US1** — `tellme -b` removes the last turn (file holds K−1), reports 1, exits 0, records zero provider requests.
- [X] **T009** (Green) **US1** — `tellme -b 2` removes the last 2 (file holds K−2), reports 2.
- [X] **T010** (Green) **US1** — the tool-using exchange survives while the later plain exchange is undone (the surviving line keeps its `steps`).
- [X] **T011** (Green) **US1** — `-b 5` on a 1-exchange session clears it; the archive stays absent.
- [X] **T012** (Green) **US2** — `tellme -b "third Q"` rolls back then runs the prompt (file = K−1+1; the last exchange asks/answers the new turn).
- [X] **T013** (Green) **US1** — `-b 0` refuses (usage error) and leaves the history unchanged; `-b --new` refuses and leaves the history unchanged.
- [X] **T014** (Refactor under green) Keep `Rollback` small + single-owned; the pre-pass one helper; the confirmation text a single constant.
- [X] **T015** (Regression) `make verify` + `go test -count=1 ./...` green; `-l`/`--new`/`-t` unchanged; `stdout` byte-exact for the existing paths.

## Unit pins

- [X] **U1** `internal/infrastructure/history/file_store_test.go` — `Rollback`: normal (K→K−N), clamp (N>available ⇒ 0 lines), `n ≤ 0` no-op, missing file ⇒ 0, the archive is NOT created, the surviving bytes are the original lines (unchanged), a decode failure returns an error and leaves the file intact.
- [X] **U2** `internal/cli/cli_test.go` — the optional-int pre-pass covers `-b` (bare `-b` ⇒ 1; `-b 3` ⇒ 3; `-b "p"` ⇒ 1 + prompt `p`); `-b 0` ⇒ usage error; `-b --new` ⇒ usage error; `renderRollback` prints the confirmation to stdout.

## Deliberate narrowings (recorded)

- **N-1 — no `-b` × `-t`/`--tool-usage` E2E.** The order is recorded (ADR 0053 RF-081-5); the round does not add a scenario for the unlikely combination.
- **N-2 — no concurrent-rollback E2E.** No `flock` (the repo's standing policy); recorded as RF-081-3.

## Fold ledger

- **F-081-1** (self-found at implementation) — `run`'s cyclomatic complexity exceeded the `cyclop` gate (19 > 15) after adding the `-b` validation + the prompt-bearing rollback. Folded: the `-b` usage validation moved into `dispatchReporting` (the offline-dispatch owner) and the prompt-bearing rollback moved into `renderTurn` via a `turnOptions.rollbackTurns` field; `run` is back under the gate and no behaviour changed.
- **F-081-2** (self-found at implementation) — the topology audit (`--root specs/truth/features/cli`) initially reported 11 errors: (a) a **blank line** split my new Given row out of the `history/dsl.md` Given table (fix: removed the blank line); (b) **two When rows** (`… the last {count} turns` and `… {count} turns`) both matched `the operator rolls back the last 2 turns` (dsl-exact-one-match) — folded by renaming the invalid-count form to `the operator rolls back a count of {count} turns`; (c) the refusal Thens used the `usage`-module `tellme exits with the usage error code` row from a `history` feature — switched to the **interface-root** `tellme explains on stderr that "the command-line usage is invalid"` row. **Recorded carried-check delta:** the audit now reports **6** errors (the same **5 pre-existing** + **1** of the already-documented cross-module class: a `chat`-module `a configured provider …` Given used in the `history` rollback feature — exactly the `history/reviewing-the-turn-log.feature:12` class). The audit is a **carried** check, not a `make verify` member.

*(The architect review-fold ledger — F-081-3 … — is appended after the review-fold loop.)*
