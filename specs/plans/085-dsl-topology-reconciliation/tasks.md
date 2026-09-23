# Tasks — round 085 `085-dsl-topology-reconciliation`

Constraint-ordered execution. The round is docs/truth-only; the task list is short.

## Phase 1 — Setup

- [X] **T001** Create the branch `085-dsl-topology-reconciliation` off `dev` and the plan package
      `specs/plans/085-dsl-topology-reconciliation/` (`spec.md` · `checklists/requirements.md` ·
      `truth-delta.md` · `research.md` · `plan.md` · `tasks.md`).

## Phase 2 — Foundational

- [X] **T002** Reproduce the witness: `audit_feature_dsl_topology.py --root specs/truth/features/cli`
      ⇒ **11 errors** (the pre-fix baseline; the `STATUS.md` "6" is stale — corrected in T008).

## Phase 3 — Test Alignment & Implementation

- [X] **T003 (Class A — promote)** Move the four cross-module rows to the interface root
      `specs/truth/features/cli/dsl.md` (Given: `a configured provider … answers with …`, `the
      effective mode is …`; When: `the operator starts tellme with the prompt …`; Then: `tellme exits
      with the configuration error code`) and remove them from `chat/dsl.md` /
      `configuration/dsl.md` (single authority).
- [X] **T004 (Class B — restore)** Add the `## Given (round 083)` / `## Then (round 083)` headings +
      `DSL 句型` header rows in `chat/dsl.md` so the five round-083 rows are parsed.
- [X] **T005 (Class C — add)** Add the composite projection row
      (`… read "{path}" with the reason "{reason}" … and reports the token usage:`) with the
      `prompt/cached/completion/thinking` DataTable to `chat/dsl.md`.
- [X] **T006 (#174 / R8d)** Correct the `Makefile` `make help` `verify-no-network` line and the target
      header comment; retire the `techstack.md` R8d bullet; annotate the ADR 0012 **index row** (the §Forward body is left verbatim).
- [X] **T007 (witness, re-run)** `audit_feature_dsl_topology.py` ⇒ **0 errors**; warnings not
      regressed; the four promoted phrases each appear exactly once (at the root).

## Phase 4 — Green / verify

- [X] **T008 (Green)** `make check` (verify + test) green; `go test -count=1 ./...` green with the
      same E2E scenario/step counts; `make modelith-check` green; `gofmt`/`goimports` clean.

## Phase 5 — Records

- [X] **T009** `STATUS.md` — round 085 in flight → delivered; correct the topology env-note (11 → 0);
      record the round; the day summary; the PR.

---

## Claim → Witness ledger

| Claim | Witness | Mutant that reddens it |
| --- | --- | --- |
| The topology audit reports 0 errors | `audit_feature_dsl_topology.py --root specs/truth/features/cli` exits 0 (0 errors; 53 features · 6 modules · 21 root rows · 457 module rows · 2328 steps) | re-introduce a module-scoped shared row, or strip a round-083 header ⇒ the corresponding step surfaces as "找不到 DSL row" (back to 11) |
| The four shared rows have one home (root) | each phrase appears exactly once across `dsl.md` + `*/dsl.md`, at the root | leaving a copy in a module ⇒ the audit's duplicate-authority error |
| The behaviour is unchanged | `go test -count=1 ./...` green, E2E counts unchanged (314 scenarios · 2354 steps) | a `.feature`/step-definition edit ⇒ count drift |
| The help text describes the witness | `make help \| grep verify-no-network` | restoring the retired "build-graph capability guard" wording |
