# Requirements checklist — round 085 `085-dsl-topology-reconciliation`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/085-dsl-topology-reconciliation/`) — `fresh-package-per-round`.
- [x] The `Spec` is PM-authored (a docs/truth reconciliation; no PM-owned requirement gap).
- [x] Anchor issues locked: [#173](https://github.com/gosharplite/tellme/issues/173) (the topology
      audit) + [#174](https://github.com/gosharplite/tellme/issues/174) (the R8d help text).
- [x] Clarify **not escalated (0 questions)** — the audit output locks the goal; promotion is the
      only `dsl-single-authority`-legal fix (A2).

## Judgeable criteria

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | Audit reports 0 errors | `audit_feature_dsl_topology.py --root specs/truth/features/cli` exits 0 |
| FR-002 | Each shared row has one home (root) | `grep -c` the four phrases across `*/dsl.md` + `dsl.md` ⇒ each appears once, at root |
| FR-003 | The round-083 block is parseable | the audit parses the five rows (they no longer surface as "找不到 DSL row") |
| FR-004 | The composite row exists | the composite Given in `colouring-the-session-chrome.feature` matches exactly one row |
| FR-005 | The help line describes the witness | `make help \| grep verify-no-network` |
| FR-006 | R8d retired | the `Not Introduced Yet` bullet is gone from `techstack.md` |

## Boundaries

- **In**: `dsl.md` reconciliation; `Makefile` help string/comment; the R8d truth/ADR records.
- **Out**: product code; `.feature`/step-definition changes; new ADRs; other non-fix-catalog items.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable witness (the audit 11 → 0) exists on day one.
