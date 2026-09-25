# System analysis — round 091 `091-record-hygiene-tail`

## 1. Interface inventory

| # | Interface | Kind | Change |
| --- | --- | --- | --- |
| I-1 | The CLI end (tellme's operator terminal) | `cli` | **NOOP** behaviour — no command, flag, argument, exit code, stream, or persisted record changes |

A **plain line-oriented CLI**; no new terminal UX surface → `/axb-ui-plan` **skipped**.

## 2. Wave / planner delegation

| Wave | Interface | Planner | Verdict |
| --- | --- | --- | --- |
| 1 | I-1 `cli` end | `/axb-api-plan` | **NOOP** (no API surface — a CLI end) |
| 1 | I-1 `cli` end | `/axb-data-plan` | **NOOP** (no persisted-state change) |
| 1 | I-1 `cli` end | `/axb-dsl-refine` | **MODIFY** — the CLI contract owner; one **prose note** reconciled in `chat/dsl.md` (no `.feature`/row semantics change) |
| 1 | I-1 `cli` end | `/axb-ui-plan` | **SKIPPED** (plain CLI) |

The `cli` end is carried forward to its contract owner `/axb-dsl-refine` (the CLI-streamlined rule).

## 3. The change surface

| Surface | Owner | Edit |
| --- | --- | --- |
| `specs/truth/techstack.md:36` | `/axb-technical-research` | Drop the "ninth" ordinal from the `search_files` row |
| `specs/truth/techstack.md:31` | `/axb-technical-research` | "the ten-value exit-code set" → "the exit-code set" |
| `specs/truth/techstack.md:105` | `/axb-technical-research` | "the vocabulary stays ten" → "the class-phrase vocabulary is unchanged" |
| `specs/truth/features/cli/chat/dsl.md` (round-079 note) | `/axb-dsl-refine` | "the ten-value exit-code set" → "the exit-code set" |
| `docs/decisions/README.md:69` | record (index) | Drop the "ninth" ordinal from the ADR 0043 index row |

## 4. Boundaries honoured

- **No** product code; no `.feature` step text; no `.feature` row semantics; no `go.mod`/`go.sum`.
- **No** Accepted-ADR body edited (ADR 0043 + the `0051`/`0052`/`0055`/`0058` counts stay verbatim).
- **No** frozen `specs/plans/NNN-*/**` touched.
- The round-090 offered-set carrier (`cmd/tellme/deps_offered_set_test.go`) is untouched.

## 5. Domain model — not modelled (ADR 0041 escape hatch)

No modelled behaviour changes: nothing in `docs/domain-model/**` carries a tool ordinal or a class-phrase
count (the `Tool` entity lists tools by name; the product model has no phrase-count invariant). The model
files are **not** touched → `modelith-check` stays green with no re-render.

## 6. Falsifiability plan (Witness ledger)

| Claim | Witness | Kind | Reddens if |
| --- | --- | --- | --- |
| **W-A** (A) no live bare ordinal | manual inspection of `techstack.md:36` + `docs/decisions/README.md:69` (no "ninth") + `grep -rn ninth` over live surfaces = only frozen history + the immutable ADR 0043 body | **carried, manual** | the ordinal is reintroduced on a live surface |
| **W-C** (C) no stale "ten" on a live surface | manual inspection of the three live surfaces (`techstack.md:31`/`:105`, `chat/dsl.md` round-079 note); the ADRs byte-identical to `dev` | **carried, manual** | a stale count is reintroduced (no `make verify` member — RF-089-6) |
| **W-B** (B) | **deferred** — homed on [#189](https://github.com/gosharplite/tellme/issues/189) (D5.1); no carrier added | n/a | n/a |
| **Standing** | `make verify-adr-index` (the index edit), `modelith-check` (no drift), `go test -count=1 ./...` | mechanical | the index row breaks ADR-index consistency |

> **Honest narrowing (the F-089-1 lesson):** W-A and W-C are **carried, manual** checks — a docs-prose
> claim has no mechanical carrier (ADR 0060 §Forward RF-089-6 / round-090 TD-090-1). This is recorded, not
> dressed as a mechanical witness.
