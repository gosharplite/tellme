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

- [X] **T008** — **[WITNESS] W-A** — carried, manual: `techstack.md:36` + `docs/decisions/README.md:69`
  carry no ordinal; `grep -rn ninth` over live surfaces returns only frozen history + the immutable ADR
  0043 body.
- [X] **T009** — **[WITNESS] W-C** — carried, manual: the three live count surfaces name their subject and
  carry no stale "ten"; the four ADRs are byte-identical to `dev`.
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
| **CLM-001 (A)** — `techstack.md`'s `search_files` row carries no bare ordinal | manual inspection of `techstack.md:36`; `grep -rn ninth` over live surfaces | carried, manual (RF-089-6) | verified |
| **CLM-002 (A)** — no live bare tool ordinal remains; ADR 0043 verbatim | `docs/decisions/README.md:69` ordinal-free; `git diff` shows no `docs/decisions/0043-*.md` change | carried, manual + `git diff` | verified |
| **CLM-003 (C)** — the live count surfaces name their subject, no stale "ten" | manual inspection of `techstack.md:31`/`:105` + `chat/dsl.md` round-079 note; `git diff` shows no `0051`/`0052`/`0055`/`0058` change | carried, manual + `git diff` | verified |
| **CLM-004 (record)** — the ADR index stays consistent after the row edit | `make verify-adr-index` | mechanical | verified |
| **CLM-005 (B)** — the enumerator binding is deferred with a recorded reason, homed durably | `research.md` D5.1 + `plan.md` §5 + the linked live issue [#189](https://github.com/gosharplite/tellme/issues/189) | record + a live issue | verified |
| **CLM-006** — no product behaviour change | `git diff --stat` touches no `.go`/`go.mod`/`go.sum`; `make verify` **OK** + `go test -count=1 ./...` green (E2E 330 · 2487, unchanged) | mechanical | verified |

## Fold ledger

_(appended during the architect review-fold loop, if run)_
