# Tasks — round 091 `091-record-hygiene-tail`

One task per unit of work; the round is docs/records-only (no product code). A `[WITNESS]` task pins a
non-observable claim (planned); `[X]` marks a verified task.

## Setup

- [X] **T001** — Create the round branch `091-record-hygiene-tail` off `dev` `9f9cd9c`; create the plan
  package `specs/plans/091-record-hygiene-tail/`.

## Foundational

- [X] **T002** — Verify the evidence the issue surveyed: the offer order in `cmd/tellme/deps.go`
  `assembleAgentTools` (`search_files` 4th; base set eight); the class-phrase authority
  (`specs/truth/features/cli/dsl.md`, **eleven** phrases); the exit-code set (`internal/cli/exitcode.go`,
  **seven**); the live "ten"/"ninth" surfaces.

## Test Alignment & Implementation

- [X] **T003** — **(A)** `specs/truth/techstack.md:36` — drop the "ninth" ordinal (name + round/ADR only).
- [X] **T004** — **(A)** `docs/decisions/README.md` — drop the "ninth" from the ADR 0043 index row; leave
  the ADR 0043 body verbatim.
- [X] **T005** — **(C)** `specs/truth/techstack.md:105` — name the subject ("the class-phrase vocabulary is
  unchanged"), drop the stale "ten".
- [X] **T006** — **(C)** `specs/truth/techstack.md:31` — "the ten-value exit-code set" → "the exit-code set".
- [X] **T007** — **(C)** `specs/truth/features/cli/chat/dsl.md` (round-079 note) — "the ten-value exit-code
  set" → "the exit-code set" (prose only).

## Witnesses

- [X] **T008** — **[WITNESS] W-A** — carried, manual: the predicate `grep -rn 'ninth agent tool'` over the
  live tree minus (frozen history · the immutable ADR 0043 body · a self-referential defect mention) returns
  **empty**; `techstack.md:36` + `docs/decisions/README.md:69` + the `search.go` comment carry no ordinal.
- [X] **T009** — **[WITNESS] W-C** — carried, manual: the three live count surfaces name their subject and
  carry no stale "ten"; the **eight** ADRs (`0051`–`0058`) are byte-identical to `dev`.
- [X] **T010** — **(B)** defer with a recorded reason (research D5.1) and home the deferral on a live issue —
  [#189](https://github.com/gosharplite/tellme/issues/189).

## Verification

- [X] **T011** — `make verify` **OK** (incl. `verify-adr-index`, `modelith-check` no drift); `go test -count=1 ./...`
  **green** (E2E **330 scenarios · 2487 steps — unchanged**); `go.mod`/`go.sum` **unchanged**; **no**
  product-code diff.

## Delivery

- [X] **T012** — Commit + push the round branch; open the round PR (no Copilot review; only a human merges).

---

## Claim → Witness ledger

| Claim | Carrier / witness | Kind | Status |
| --- | --- | --- | --- |
| **CLM-001 (A)** — no **live** surface carries a bare `ninth agent tool` ordinal | **predicate** `grep -rn 'ninth agent tool'` over the live tree minus the exclusion list (frozen history · the immutable ADR 0043 body · a self-referential mention of the defect) → **empty** after the fold; product code (`search.go`), truth, the ADR index row, and STATUS fixed | carried, manual (RF-089-6) | verified |
| **CLM-002 (A)** — no live bare tool ordinal remains; ADR 0043 verbatim | `docs/decisions/README.md:69` ordinal-free; `git diff` shows no `docs/decisions/0043-*.md` change | carried, manual + `git diff` | verified |
| **CLM-003 (C)** — the live count surfaces name their subject, no stale "ten" | manual inspection of `techstack.md:31`/`:105` + `chat/dsl.md` round-079 note; `git diff` shows no `0051`–`0058` body change | carried, manual + `git diff` | verified |
| **CLM-004 (record)** — the ADR index stays consistent after the row edit | `make verify-adr-index` | mechanical | verified |
| **CLM-005 (B)** — the enumerator binding is deferred with a recorded reason, homed durably | `research.md` D5.1 + `plan.md` §5 + the linked live issue [#189](https://github.com/gosharplite/tellme/issues/189) | record + a live issue | verified |
| **CLM-006** — no product **behaviour** change (behaviour-scoped: the fold adds comment-only edits — N-091-3) | `git diff --stat` touches no `.go` logic/`go.mod`/`go.sum` (only `.go` **comments**); `make verify` **OK** + `go test -count=1 ./...` green (E2E 330 · 2487, unchanged) | mechanical | verified |

## Fold ledger

**Review 1** (PR [#190](https://github.com/gosharplite/tellme/pull/190), the `architect` peer — init once with `SESSION-BOOTSTRAP.md`, continuations): `review` **`APPROVE WITH REQUIRED FOLDS`** — no `[ARCHITECTURAL BLOCKER]`; the folds are **record accuracy** (the round's own subject class): **F-091-1** the "no live bare ordinal" claim was falsified on the live tree — `internal/infrastructure/tools/search.go:20` (a product-code comment, "the ninth agent tool") **and** `STATUS.md:4` (a self-referential mention) matched the recorded W-A sweep, which was also non-discriminating; **F-091-2** the immutable-ADR enumeration named **4** of **8** ("ten" occurrences: also `0053:44`, `0054:26`, `0056:53`, `0057:30`); **F-091-3** checklist FR-1 / `spec.md` §2 quoted a non-shipped expected string. Plus **TD-091-1** (`filesystem_test.go:235` stale "five tools" comment), **N-091-1** (research D1 over-claim "nothing states"), **N-091-2** (why `eleven` survives), **N-091-3** (behaviour-scope CLM-006), **N-091-4** (name the "quotes-the-defect" exclusion), **R-091-1** (the eight ADR bodies carry the stale figure permanently), **R-091-2** (#189 unaffected), **R-091-3** (W-A/W-C honestly carried; only their scope was wrong).

**Fold 1** (`f58ce9e`): F-091-1 → dropped the ordinal from the `search.go` comment + the self-referential literal from `STATUS.md` + restated W-A as a predicate with an explicit exclusion list (frozen history · the immutable ADR 0043 body · a self-referential defect mention); F-091-2 → restated the ADR enumeration to all **eight** (`research.md` D3/D7, `spec.md`, `checklists/requirements.md`, `truth-delta.md`, `plan.md` §3, `tasks.md` T009/CLM-003) + recorded R-091-1; F-091-3 → corrected the FR-1 expected string + `spec.md` §2 to the shipped text; TD-091-1 → dropped the stale "five tools" count from the `filesystem_test.go` comment; N-091-1/2/3/4 folded (research D1 softened; the "eleven-survives" note added; CLM-006 behaviour-scoped; the exclusion list names the self-referential shape).

**Fold-verification 1** (the `architect` peer): **`FOLDS VERIFIED WITH RESIDUALS`** — the substantive folds verified (the predicate sweep **empty** · the eight-ADR enumeration cross-checked against the tree · all eight ADR bodies md5-byte-identical to `dev` · the shipped FR-1 string · the `five tools` count gone · the gates green · every changed `.go` line a comment); residuals **RES-091-FV-1** (F-091-2 incomplete on `plan.md:35` + `tasks.md` T009 — the "four" enumeration survived), **RES-091-FV-2** (F-091-1's restatement missed `checklists/requirements.md` FR-2 + `tasks.md` T008 — the bare-word `grep -rn ninth` survived, still factually wrong), **RES-091-FV-3** (N-091-3 folded `CLM-006` only — the checklist `NFR-1` + the boundaries still claimed no `.go` touched), **RES-091-FV-4** (an observation: the round-082 interactive-prompt flake recurred → filed as live issue [#191](https://github.com/gosharplite/tellme/issues/191); not this round's defect).

**Fold 2 (residuals)** (`<pending>`): folded RES-091-FV-1/2/3 (record-only): the ADR enumeration completed to eight on `plan.md` §3 + `tasks.md` T009/CLM-003 (and the fold ledger's surface list corrected); the W-A predicate restated in `checklists/requirements.md` FR-2 + `tasks.md` T008 (with the N-091-4 exclusion); the checklist `NFR-1` + the boundaries restated **behaviour-scoped** to name the two comment-only `.go` edits. RES-091-FV-4 → homed on live issue [#191](https://github.com/gosharplite/tellme/issues/191).
