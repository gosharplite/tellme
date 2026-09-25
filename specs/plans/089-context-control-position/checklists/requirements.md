# Requirements checklist — round 089 `089-context-control-position`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/089-context-control-position/`) — `fresh-package-per-round`.
- [x] The `Spec` is PM-authored (a docs/truth-record round; no PM-owned requirement gap).
- [x] Anchor issue locked: [#184](https://github.com/gosharplite/tellme/issues/184) (**DoD = close it**).
- [x] Clarify **not escalated (0 questions)** — the issue locks the goal and the decision content; the
      operator's round instruction grants the intent the issue recorded as pending (A1).

## Judgeable criteria

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | `-b` is no longer an exclusion | `specs/truth/techstack.md` *History summarisation* bullet names `-b` as **delivered** (ADR 0053) and no longer lists it as out of scope |
| FR-002 | `--retry` stays a non-introduction | the bullet records `--retry` (roll back + resend); the reference ships it, tellme does not |
| FR-003 | An ADR is recorded + indexed | `docs/decisions/0060-context-control-position.md` exists; `# ADR 0060` unique; `make verify-adr-index` green |
| FR-004 | No unfalsifiable equivalence claim | the ADR states a **non-equivalence** and cites `offering-the-agent-tools.feature` (the "no summarisation tool offered" carrier) |
| FR-005 | Surviving exclusions stay true | all four: `--retry` (no code — `grep`), token-budget pruning (no prune seam in `internal`/`cmd`), history pinning (no pin seam in `internal`/`cmd`), `SafePath`/consent (absent by the settled no-security-layer direction) |

## Boundaries

- **In**: `techstack.md` clause reconciliation; ADR 0060 + its index row; the ADR-0053 index back-pointer.
- **Out**: product code; `.feature`/step changes; introducing summarisation / pruning / pinning;
  porting `--retry`; other `Not Introduced Yet` items; rewriting frozen `specs/plans/NNN-*/` history.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable witnesses exist on day one (the stale clause is
present-tense findable pre-fix; `verify-adr-index` is a standing gate; the summarisation-absence
carrier already ships in the E2E).
