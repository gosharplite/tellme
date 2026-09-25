# Tasks — round 093 `093-interactive-prompt-seam-determinism`

Constraint-ordered execution.

## Phase 1 — Setup

- [X] **T001** Create the branch `093-interactive-prompt-seam-determinism` off `dev` and the plan package
      `specs/plans/093-interactive-prompt-seam-determinism/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `features/acceptance/**` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational (reproduce the defect)

- [X] **T002** Reproduce the round-191 defect. (a) **Root cause A** — the round-023 handshake's gate is the
      generic border `┌`: run the shipped seam at 16-way concurrency ⇒ **599 / 960 red** (the composed frame
      is dropped). (b) **Root cause B** — a scripted `\n` is decoded by bubbletea as `KeyCtrlJ`, which the
      bubbles `textarea` does not bind to `InsertNewline`; drive the real runtime with a raw LF ⇒ the editor
      renders `line oneline two` (one joined row) **960 / 960**, and the pre-093 substring assertion passes
      vacuously.

## Phase 3 — Test Alignment & Implementation

- [X] **T003** `tests/e2e/harness/cmd_helper.go` — add `paintGate(compose, fallback)` (the composition's
      last visible line; control bytes skipped; the fallback is the compose-less marker) and derive the gate
      in `runExecSynced`; rename the marker parameter to `fallbackMarker`.
- [X] **T004** `tests/e2e/steps/tui_keys.go` — add `tuiKeyEnter = "\r"` and `typeText` (`\n → CR`); route
      `tuiKeysTypeAndAbort` / `tuiKeysTypeAndSubmit` / `tuiKeysTypeAcceptAbort` through it; `step_r038.go`'s
      empty-submit sequence likewise.
- [X] **T005** `tests/e2e/steps/step_t011_chat_then_keeps_typed_text.go` — strengthen `thenKeepsTypedText`
      with the multi-line separate-rows clause; extract the pure helper `joinedRow(rows, parts)`.
- [X] **T006 (carrier)** `tests/e2e/harness/cmd_helper_test.go` — `TestSyncedStdinGateWaitsForTheComposedText`
      (the content-aware gate; a revert to a constant marker reddens).
- [X] **T007 (carrier)** `tests/e2e/steps/tui_keys_test.go` — `TestTypeTextDeliversTheTerminalEnterByte`
      (a revert to LF reddens).
- [X] **T008 (carrier)** `tests/e2e/steps/step_t011_chat_then_keeps_typed_text_test.go` —
      `TestJoinedRowCatchesTheJoinedState` (the discriminating clause).
- [X] **T009 (witness, re-run)** Reproduce W-A (revert the gate to `┌`) ⇒ **607 / 960** red at 16-way
      concurrency; W-B (revert the Enter byte to LF) ⇒ the multi-line Example reddens (joined row); W-C
      (drop the separate-rows clause) ⇒ the joined state passes — all then reverted.
- [X] **T010 (Green)** After the fix: the feature is green **50 / 50** serially and **0 / 960** red at 16-way
      concurrency (was 599 / 960); the full E2E suite is green.

## Phase 4 — Verify

- [X] **T011 (Green)** `make check` green; E2E counts unchanged (330 scenarios · 2487 steps);
      `verify-no-test-sleep` green; `verify-architecture` 0 violations; `modelith-check` no drift;
      `gofmt`/`goimports` clean; `go.mod`/`go.sum` unchanged; no `internal/**` or `cmd/**` change.

## Phase 5 — Records

- [X] **T012** **ADR 0063** (`docs/decisions/0063-*.md`) + `docs/decisions/README.md` index row;
      `specs/truth/techstack.md` (the *Interactive TUI prompt harness* row).
- [X] **T013** `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` (header note +
      carrier comment) and `chat/dsl.md` (the `keeps the typed text` row).
- [X] **T014** `STATUS.md` — round 093 in flight (→ delivered at closeout); the day summary; the PR.

---

## Claim → Witness ledger

| Claim | Witness | Mutant that reddens it |
| --- | --- | --- |
| (FR-001) the handshake releases the terminal key only after the composed text painted | `TestSyncedStdinGateWaitsForTheComposedText` (the `paintGate` pin) | revert the gate to the constant `┌` ⇒ the pin reds **and** the E2E flakes (W-A, measured 607/960 red) |
| (FR-002) a scripted line break is the terminal Enter byte (CR) | `TestTypeTextDeliversTheTerminalEnterByte` + the multi-line E2E Example | revert `typeText` to LF ⇒ the pin reds **and** the Example reds (joined row) |
| (FR-003) the assertion is discriminating (each line on its own row) | `TestJoinedRowCatchesTheJoinedState` + the strengthened `thenKeepsTypedText` | drop the separate-rows clause ⇒ the joined state passes (the pinned regression returns green) |
| (SC-002) determinism under repetition | 50/50 green serially; 0/960 red at 16-way concurrency | the pre-fix seam: 599/960 red (measured) |
| (NFR-001 / I-2) no product change, no sleep/retry | `internal/**` + `cmd/**` unchanged; `make verify-no-test-sleep`; `go.mod`/`go.sum` unchanged | a product edit ⇒ the "no `internal/**`/`cmd/**` diff" review + a behaviour suite red |
| (I-4) behaviour identity of the whole suite | `go test -count=1 ./...` green; E2E 330 · 2487 unchanged | an E2E count/behaviour change ⇒ red |

---

## Fold ledger (architect review — PR #TBD)

_(filled during the review-fold loop.)_
