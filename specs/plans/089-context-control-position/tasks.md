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
| `techstack.md` no longer lists `-b` as an exclusion | reproducible `grep -- '--retry' specs/truth/techstack.md` — the *History summarisation* bullet names `-b` as delivered (ADR 0053) | re-adding `-b` to the settled-exclusion list ⇒ the clause reads as a false exclusion again |
| The decision is durable + indexed once | `make verify-adr-index` (standing `make verify` member) over `docs/decisions/0060-context-control-position.md` + its index row | deleting the index row, or duplicating `# ADR 0060` ⇒ `verify-adr-index` reds |
| tellme offers no summarisation tool (already carried) | the E2E `chat/offering-the-agent-tools.feature` Rule *"a summarisation tool must not be offered"* | a re-offered summarisation tool ⇒ that Rule reds (existing carrier, cited not re-created) |
| The behaviour is unchanged | `go test -count=1 ./...` green, E2E counts unchanged | a `.feature`/step-definition edit ⇒ count drift |


