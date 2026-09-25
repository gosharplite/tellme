# Requirements checklist — round 091 `091-record-hygiene-tail`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/091-record-hygiene-tail/`) — `fresh-package-per-round`.
- [x] The `Spec` is PM-authored (a docs/truth-record round; no PM-owned requirement gap).
- [x] Anchor issue locked: [#188](https://github.com/gosharplite/tellme/issues/188) (**DoD = close it**).
- [x] Clarify **not escalated (0 questions)** — the issue locks the goals and the two decisions (drop the
      ordinal; name the count's subject); the operator's round instruction grants the intent (A1).

## Judgeable criteria

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-1 | `techstack.md`'s `search_files` row carries no bare ordinal | `specs/truth/techstack.md:36` reads *"the `search_files` tool — the missing half of the reader trio …"*, no "ninth" |
| FR-2 | No **live** bare tool ordinal remains; the Accepted ADR is verbatim | `grep -rn ninth` over live surfaces returns only frozen history (session summaries, archives) + `docs/decisions/0043-*.md` (immutable); `docs/decisions/README.md`'s 0043 index row is ordinal-free |
| FR-3 | The live "ten" surfaces name their subject and drop the stale count | `techstack.md:31` + `:105` and `features/cli/chat/dsl.md` (round-079 note) carry no stale "ten"; the ADRs are byte-identical to `dev` |
| FR-4 | (B) bound **or** deferred with a recorded reason on a durable home | `plan.md` §5 records the reason; a **live issue** homes the deferral (or a carrier binds the enumerator) |
| NFR-1 | No product/`go.mod`/`go.sum` change | `git diff --stat` touches no `.go`/`go.mod`/`go.sum` |

## Boundaries

- **In**: `specs/truth/techstack.md` (rows `:31`, `:36`, `:105`); `specs/truth/features/cli/chat/dsl.md`
  (the round-079 note); `docs/decisions/README.md` (the 0043 index row); the plan package; a home for (B).
- **Out**: product code; `.feature`/step/row-semantics changes; **editing an Accepted ADR body**
  (0043 + the `0051`/`0052`/`0053`/`0054`/`0055`/`0056`/`0057`/`0058` counts stay verbatim); rewriting frozen `specs/plans/NNN-*/**`;
  a *general* docs-prose gate (the larger RF-089-6/TD-090-1 decision).

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the pre-fix stale text is present-tense findable on day one
(`techstack.md:36` "ninth"; `techstack.md:31`/`:105` + `chat/dsl.md` "ten"), and the standing gates
(`verify-adr-index`, `modelith-check`) cover the record surfaces the round edits.
