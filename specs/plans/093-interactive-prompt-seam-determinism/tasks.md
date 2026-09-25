# Tasks — round 093 `093-interactive-prompt-seam-determinism`

Constraint-ordered execution.

## Phase 1 — Setup

- [X] **T001** Create the branch `093-interactive-prompt-seam-determinism` off `dev` and the plan package
      `specs/plans/093-interactive-prompt-seam-determinism/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `features/acceptance/**` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational (reproduce the defect)

- [X] **T002** Reproduce the round-191 defect. (a) **Root cause A** — the round-023 handshake's gate is the
      generic border `┌`: run the shipped seam at 16-way concurrency ⇒ **~62 % red** on the authoring host
      (host/load-specific — **F-093-2**; the composed frame is dropped). (b) **Root cause B** — a scripted
      `\n` is decoded by bubbletea as `KeyCtrlJ`, which the bubbles `textarea` does not bind to
      `InsertNewline`; drive the real runtime with a raw LF ⇒ the editor renders `line oneline two` (one
      joined row) deterministically, and the pre-093 substring assertion passes vacuously.

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
- [X] **T009 (witness, re-run)** Reproduce W-A (revert the gate to a constant `┌`) ⇒ the `paintGate` pin
      reddens and the flake re-opens (**~62 %** at 16-way concurrency, authoring host — **F-093-2**); W-A′
      (revert the **call site** `marker := paintGate(…)` → `marker := fallbackMarker`) ⇒ the pin stays
      **green**, only the repetition witness reds (**F-093-1** — the two mutations are not equivalent);
      W-B (revert the Enter byte to LF) ⇒ the multi-line Example reddens (joined row); W-C (drop the
      separate-rows clause) ⇒ the joined state passes — all then reverted.
- [X] **T010 (Green)** After the fix: the feature is green **50 / 50** serially and **0 red** at 16-way
      concurrency (was ~62 %, authoring host — **F-093-2**); the full E2E suite is green.

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
| (FR-001) the handshake releases the terminal key only after the composed text painted — the gate **rule** | `TestSyncedStdinGateWaitsForTheComposedText` (the `paintGate` pin) | **gut `paintGate`** (return the fallback) ⇒ the pin reds **and** the E2E flakes (W-A, ~62 % at 16-way concurrency, authoring host — **F-093-2**) |
| (FR-001) the call-site **wiring** — `runExecSynced` uses the rule's result as the marker | the **E2E repetition witness** (probabilistic; there is no deterministic carrier) | **revert the call site** (`marker := paintGate(…)` → `marker := fallbackMarker`) ⇒ the pin stays **green**, only the repetition witness reds (W-A′, **F-093-1**; RF-093-1) |
| (FR-002) a scripted line break is the terminal Enter byte (CR) | `TestTypeTextDeliversTheTerminalEnterByte` + the multi-line E2E Example | revert `typeText` to LF ⇒ the pin reds **and** the Example reds (joined row) |
| (FR-003) the assertion is discriminating (each line on its own row) | `TestJoinedRowCatchesTheJoinedState` + the strengthened `thenKeepsTypedText` | drop the separate-rows clause ⇒ the joined state passes (the pinned regression returns green) |
| (SC-002) determinism under repetition | 50/50 green serially; 0 red at 16-way concurrency | the pre-fix seam: intermittent red (~62 % authoring host; ~2.8 % review host — host/load-specific, **F-093-2**) |
| (NFR-001 / I-2) no product change, no sleep/retry | `internal/**` + `cmd/**` unchanged; `make verify-no-test-sleep`; `go.mod`/`go.sum` unchanged | a product edit ⇒ the "no `internal/**`/`cmd/**` diff" review + a behaviour suite red |
| (I-4) behaviour identity of the whole suite | `go test -count=1 ./...` green; E2E 330 · 2487 unchanged | an E2E count/behaviour change ⇒ red |

---

## Fold ledger (architect review — PR [#193](https://github.com/gosharplite/tellme/pull/193))

Review: [`pull/193#issuecomment-5834874802`](https://github.com/gosharplite/tellme/pull/193#issuecomment-5834874802)
— **APPROVE WITH REQUIRED FOLDS** (no `[ARCHITECTURAL BLOCKER]`).

| Fold | Resolution |
| --- | --- |
| **F-093-1** (carrier/record) the `paintGate` pin exercises the gate **rule**, not its **call site** — reverting `runExecSynced`'s wiring reproduces the flake yet leaves the pin green (W-A′ ≠ W-A) | Restated the claim to name the covered surface (the rule) across the ledger (FR-001 rows split: rule vs wiring), ADR 0063 D4 + §Verification + §Forward (RF-093-1 is now the **wiring** gap, not a vague "ordering"), `research.md` D4 + §5; the W-A description now states the two mutations are not equivalent. |
| **F-093-2** (record accuracy) the determinism magnitudes are host/load-specific | Qualified the pre-fix/~62 % figures as authoring-host measurements (review host ~2.8 %: 4/144 vs 0/288; direction reproduces) across `research.md` §1/D4, `spec.md` §2/SC-002, ADR 0063 Context/Consequences/Verification, this ledger, the day log, and the `cmd_helper_test.go` comment (N-093-1). |
| **F-093-3** (record/live-state) `STATUS.md` Candidates self-contradiction ("0 open while round 093 carries #191") | Restated to "the tracker holds the in-flight anchor #191" — F-092-6 lineage. |
| **TD-093-1** the `paintGate` fallback path has no direct carrier | Recorded as **RF-093-4** (ADR 0063 §Forward + `research.md` §5). |
| **N-093-1** the `cmd_helper_test.go` comment repeats the unreproduced `~63 %` | Folded with F-093-2 (the comment now states host/load-specificity) + the pin's coverage scope (F-093-1). |
| **N-093-2** this ledger heading read `PR #TBD` | Named **#193** (this heading). |
