# Tasks — round 090 `090-record-hygiene-offered-set-and-direction`

Constraint-ordered execution.

## Phase 1 — Setup

- [X] **T001** Create the branch `090-record-hygiene-offered-set-and-direction` off `dev` and the plan package
      `specs/plans/090-record-hygiene-offered-set-and-direction/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational

- [X] **T002** Reproduce the (A) baseline: `chat/dsl.md:208` says **seven** (no `search_files`); the feature
      says **eight**; `agentTools()`/`registeredToolNames()` return **eight**. Confirm no carrier reddens.

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (A — truth)** Reconcile `specs/truth/features/cli/chat/dsl.md:208`: the `集合` cell → the eight
      base tools (add `search_files`), drop the stale hand-count, name the single owner (the live registry);
      keep the `不該發生` clause. No `.feature` change (already current).
- [X] **T004 (A — carrier)** Add the doc-consistency test in `cmd/tellme` (test-only): parse the `chat/dsl.md`
      offered-set cell, assert set-equality with `agentTools()`.
- [X] **T005 (A rider)** Correct the stale history comment in
      `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go` (record round 071's `search_files`).
- [X] **T006 (B)** Reconcile `README.md`'s *"deliberately small tool surface"* enumeration to the shipped set.
- [X] **T007 (C)** Write **ADR 0061** (`docs/decisions/0061-operator-declared-direction.md`) + its index row;
      demote the `README.md` direction paragraph to a summary pointing at the ADR; point `STATUS.md`'s
      direction line at the ADR.
- [X] **T008 (D)** Correct `specs/truth/techstack.md:155` (`retry` is not a subcommand).
- [X] **T009 (witness, re-run)** The carrier passes at head; reproduce W-A (remove a tool from `agentTools()`,
      or edit the documented list) ⇒ the carrier reddens; revert.

## Phase 4 — Green / verify

- [X] **T010 (Green)** `make check` green; E2E counts unchanged (330 · 2487); `make verify-adr-index` green;
      `modelith-check` no drift; `gofmt`/`goimports` clean; `go.mod`/`go.sum` unchanged.

## Phase 5 — Records

- [X] **T011** `STATUS.md` — round 090 in flight (→ delivered at closeout); the day summary §8; the PR.

---

## Claim → Witness ledger

| Claim | Witness | Mutant that reddens it |
| --- | --- | --- |
| (FR-001) the offered-set claim is bound to (checked against) the live registry | the `cmd/tellme` doc-consistency carrier: `set(documented 集合 cell) == set(agentTools())` | remove a tool from `agentTools()` (without editing the doc) ⇒ the carrier reds; edit the documented `集合` cell ⇒ the carrier reds |
| (FR-002) `dsl.md` ↔ feature agree; eight tools incl. `search_files` | both surfaces name the same eight; the carrier's live set is eight | the doc omits `search_files` ⇒ the carrier reds |
| (FR-003) the step comment records round 071 | the comment names round 071 / `search_files`; no "grew to seven" | — (manual; a comment claim) |
| (FR-004) README enumeration matches the shipped set | `README.md` lists readers + `search_files` + write pair + `list_skills` (+ `read_image` noted) | — (manual, carried) |
| (FR-005) the direction has a durable ADR home | `make verify-adr-index` (standing) over ADR 0061 + README/STATUS point at it | delete the index row / duplicate `# ADR 0061` ⇒ `verify-adr-index` reds |
| (FR-006) the `cobra` note does not call `retry` a subcommand | `techstack.md:155` re-reads correctly | — (manual, carried) |
| (I-3) the behaviour is unchanged | `go test -count=1 ./...` green, E2E counts unchanged (330 · 2487) | a `.feature`/step-behaviour edit ⇒ count drift |

---

## Fold ledger (review-fold loop — the `architect` peer, PR #187)

Dispatched the `architect` peer per `tm-chat-ingroup`: initialized **once** with `SESSION-BOOTSTRAP.md`
(`--new`), then continuations. The architect reviewed PR #187 and posted
[`pull/187#issuecomment-5829481252`](https://github.com/gosharplite/tellme/pull/187#issuecomment-5829481252)
— **`APPROVE WITH REQUIRED FOLDS`**, no `[ARCHITECTURAL BLOCKER]`; it reproduced `make verify` /
`go test` (E2E 330 · 2487) on a scratch copy and **attacked the carrier from both directions** (doc edit ⇒
`doc (7) vs live (8)`; assembler edit ⇒ `doc (8) vs live (7)`), confirmed non-vacuity and that the parse
target is the `集合` cell.

| Fold | Resolution |
| --- | --- |
| **F-090-1** ADR 0061 §Consequences over-claimed ("the 'small surface' cannot silently grow") and contradicted its own §Forward RF-061-3 | Reworded: the offered set is **bound to** `agentTools()` (doc↔registry **consistency**), so the *documented* set cannot silently **drift**; the *size* of the surface stays a per-round judgement (RF-061-1), **not** mechanically gated (RF-061-3). |
| **F-090-2** the reconciled row kept a second, unchecked copy of the eight-tool set in its `必查` cell | **Dropped** the inline list (option (a)): the `必查` cell now refers to the row's `集合` cell / the live registry — one enumeration, the checked one. |
| **N-090-1** ADR §Context "never recorded in an ADR" over-claimed | Reworded to the honest form: the direction had **no dedicated home** (several ADRs **cite** it), only a README narrative. |
| **N-090-2** "single-sourced" read as *generated* | Restated across the round artifacts as **bound to / checked against** (a checked mirror, not a derived source). |
| **N-090-4** ADR mis-quoted `docs/decisions/README.md` | Quote corrected ("…other artifacts **(or)** future rounds…"). |
| **N-090-3 / N-090-5 / N-090-6** | Accepted, not actioned: N-090-3 (the carrier is a `go test` tier, not a `make verify` member — the plan states this honestly); N-090-5 (the `backtickedName` regex is fine today); N-090-6 (a **pre-existing** `techstack.md` "ninth agent tool" ordinal, out of this round's scope). |
