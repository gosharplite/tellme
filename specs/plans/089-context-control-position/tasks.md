# Tasks — round 089 `089-context-control-position`

Constraint-ordered execution. The round is truth/record-only; the task list is short.

## Phase 1 — Setup

- [X] **T001** Create the branch `089-context-control-position` off `dev` and the plan package
      `specs/plans/089-context-control-position/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational

- [X] **T002** Reproduce the witness (W1): confirm the stale clause in `specs/truth/techstack.md`
      (the *History summarisation* bullet lists `-b`/`--retry` as out of scope) and confirm
      `--retry` has **no** code references. The pre-fix baseline.

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (a — truth)** Reconcile `specs/truth/techstack.md` (*History summarisation* bullet):
      remove **`-b`/`--back`** from the settled-exclusion list with a forward pointer to the *Session
      lifecycle flags* row (ADR 0053); keep **`--retry`** as a real non-introduction, stated
      accurately; leave history pinning / token-budget pruning / `SafePath`-consent unchanged.
- [X] **T004 (b — record)** Write **ADR 0060** (`docs/decisions/0060-context-control-position.md`):
      the decision (no `summarize_history` port; operator-manual `-l`/`-b` context control; the
      remaining exclusions), the rationale (compression vs coarse operator trim; the round-021
      precedent; the determinism/operator-in-the-loop posture), the **non-equivalence** statement, the
      **existing** carrier citation (`offering-the-agent-tools.feature`), and a **revisit trigger**.
- [X] **T005 (index)** Add the **ADR 0060** row to `docs/decisions/README.md` (status `Accepted`) and
      a back-pointer on the **ADR 0053** index row.
- [X] **T006 (witness, re-run)** Confirm SC-001: `specs/truth/techstack.md` no longer lists `-b` as an
      exclusion; `--retry` still recorded.

## Phase 4 — Green / verify

- [X] **T007 (Green)** `make verify` green (incl. `verify-adr-index` + `modelith-check`);
      `go test -count=1 ./...` green with unchanged E2E scenario/step counts; `gofmt`/`goimports`
      clean; `go.mod`/`go.sum` unchanged.

## Phase 5 — Records

- [X] **T008** `STATUS.md` — round 089 in flight (→ delivered at closeout); the day summary §6; the PR.

---

## Claim → Witness ledger

| Claim | Witness | Mutant that reddens it |
| --- | --- | --- |
| (FR-001) `techstack.md` no longer lists `-b` as an exclusion | **carried, manual inspection of `specs/truth/techstack.md:173`** (the *History summarisation* bullet): `-b`/`--back` appears only as **delivered** (ADR 0053 / the *Session lifecycle flags* row), never in the "settled exclusion / out of scope" list. **Discriminating manual check:** the exact stale idiom ``the undo/retry flags (`-b`/`--retry`)`` is **absent from the exclusion position** (it survives only inside the trailing *"Corrected (round 089 …)"* clause as a **quoted historical reference** — so the check must target the exclusion sentence, not the word). **No mechanical carrier** — a docs-prose *status* claim has no `make verify` member (`topology-audit-not-a-gate`; recorded as RF-089-6 / TD-089-1); a bare `grep -- '-b'` is **non-discriminating** (37 matches in both base and head). | re-adding `-b` to the settled-exclusion list ⇒ the clause reads as a false exclusion again (caught by the manual check, not a test) |
| (FR-002) `--retry` stays a genuine non-introduction | the *History summarisation* bullet names `--retry` (roll back the last user message + resend); **`--retry` has no code anywhere** (`grep` over `internal/`, `cmd/`, `tests/`, `specs/truth/features` finds none) | implementing `--retry` (adding the flag/behaviour) without a new record ⇒ the bullet becomes a false non-introduction |
| (FR-003) The decision is durable + indexed once | `make verify-adr-index` (a **standing** `make verify` member) over `docs/decisions/0060-context-control-position.md` + its index row | deleting the index row, or duplicating `# ADR 0060` ⇒ `verify-adr-index` reds |
| (FR-004) tellme offers no summarisation tool (already carried) | the **negative (`不該發生`) clause of the DSL row** `the request offered exactly the agent tools` (`chat/dsl.md:208`), exercised by `chat/offering-the-agent-tools.feature` (comment `:7`, Example `:42`) — **not** a standalone `Rule` | a re-offered summarisation tool ⇒ the offered set is eight, not seven ⇒ that row reds (existing carrier, cited not re-created) |
| (FR-005) Each surviving exclusion stays **true** of the shipped system | `--retry`: no code (grep) · **token-budget pruning**: no prune seam in `internal`/`cmd` · **history pinning**: no pin seam in `internal`/`cmd` · **`SafePath`/consent**: absent by the settled *no-security-layer* direction (the model records `NoSecurityLayer`; no auth/consent surface in the code) | building any one of them without updating this ADR + the truth row ⇒ the exclusion becomes a false claim |
| (I-4) The behaviour is unchanged | `go test -count=1 ./...` green, E2E counts unchanged (330 scenarios · 2487 steps) | a `.feature`/step-definition edit ⇒ count drift |



---

## Fold ledger (review-fold loop — the `architect` peer, PR #185)

Dispatched the `architect` peer per `tm-chat-ingroup`: initialized **once** with
`SESSION-BOOTSTRAP.md` (`--new`), then continuations (no `--new`). The architect reviewed PR #185 and
posted [`pull/185#issuecomment-5828815049`](https://github.com/gosharplite/tellme/pull/185#issuecomment-5828815049)
— **`APPROVE WITH REQUIRED FOLDS`**, no `[ARCHITECTURAL BLOCKER]`; it reproduced `make verify` /
`go test` (E2E 330 · 2487) on a scratch copy of the head and confirmed ADR 0060 indexed once, unique,
non-contradictory, and the domain model correctly unmodelled.

| Fold | Resolution |
| --- | --- |
| **F-089-1** the recorded W1 carrier is not discriminating (`grep -- '--retry'` matches both states; `grep -c -- '-b'` is 37 in both) | Restated W1 honestly as a **carried, manual inspection** of `techstack.md:173` with a **discriminating manual check** (the stale idiom is absent from the *exclusion position*; it survives only as a quoted historical reference) — in `tasks.md` (ledger row 1), `research.md` D5, and `spec.md` SC-001; no mechanical carrier exists (recorded RF-089-6 / TD-089-1). |
| **F-089-2** FR-005 (and FR-002) had no ledger row; the FR-005 cell dropped `SafePath`/consent | Added ledger rows for **FR-002** and **FR-005** (all four exclusions name a witness) and completed the FR-005 cell in `checklists/requirements.md`. |
| **F-089-3** the "already carried" half was cited as a `Rule`; it is the `不該發生` clause of a DSL row | Cited precisely (the negative clause of the DSL row `the request offered exactly the agent tools`, `chat/dsl.md:208`, exercised by `offering-the-agent-tools.feature`) in ADR 0060 D6 and `research.md` D4. |
| **N-089-1** ADR `Related` mislabelled ADR 0023 as "the `-l` listing" | Relabelled (0023 = optional-count + chrome colour; the listing's record is ADR 0045, with 0054/0057). |
| **N-089-2** `--retry` grammatically grouped as a "settled exclusion" | Disambiguated in `techstack.md` (a separate sentence: `--retry` is a **genuine non-introduction**, not a settled exclusion). |
| **N-089-3** `STATUS.md` Daily log still read "(rounds 087 + 088)" | Refreshed to "(rounds 087–089)". |
| **N-089-4** RF-089-5 asserted "not engaged" without naming the model anchors | Named the two anchors (`History.history-rollback-removes-complete-turns`; `Session.session-rollback-stays-offline`). |
| **TD-089-1** a truth-prose *status* claim has no mechanical guard | Recorded as **RF-089-6** in ADR 0060 §Forward (a carried check; the F-088-1 class recurs and is caught by review). Accepted, not actioned in-round. |
