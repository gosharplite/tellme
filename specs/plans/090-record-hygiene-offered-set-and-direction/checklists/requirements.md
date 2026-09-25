# Requirements checklist — round 090 `090-record-hygiene-offered-set-and-direction`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/090-record-hygiene-offered-set-and-direction/`) — `fresh-package-per-round`.
- [x] The `Spec` is PM-authored (a docs/truth-record round + a test-only carrier; no PM-owned requirement gap).
- [x] Anchor issue locked: [#186](https://github.com/gosharplite/tellme/issues/186) (**DoD = close it**).
- [x] Clarify **not escalated (0 questions)** — #186 records the two operator-locked decisions (A1).

## Judgeable criteria

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | Offered-set claim single-sourced + mechanically checked | the `cmd/tellme` doc-consistency test reddens under a tool add/remove (W-A) and passes at head |
| FR-002 | `dsl.md` ↔ feature agree | both name the **eight** base tools incl. `search_files` |
| FR-003 | Step comment corrected | `step_r021_t026_chat_then_offered_tools.go` records round 071 (no "grew to seven") |
| FR-004 | README enumeration matches the shipped set | `README.md` lists readers + `search_files` + write pair + `list_skills` (+ `read_image` noted) |
| FR-005 | Direction recorded in an ADR; README demoted | ADR 0061 exists + indexed; README/STATUS point at it |
| FR-006 | `cobra` note corrected | `techstack.md:155` no longer calls `retry` a subcommand |
| SC-003 | ADR indexed once / unique | `make verify-adr-index` green |
| SC-006 | No regressions | `make check` green; E2E 330 · 2487 unchanged; `go.mod`/`go.sum` unchanged |

## Boundaries

- **In**: `chat/dsl.md` (offered-set) + the test-only carrier + the step comment; `README.md` (surface + direction);
  **ADR 0061** + index; `STATUS.md` direction line; `techstack.md` `cobra` note.
- **Out**: product code; `.feature` step text; relitigating the direction; a general docs-prose gate; frozen packages.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable witnesses exist on day one (W-A is a real reddening carrier —
the round's whole point; W-C is the standing `verify-adr-index` gate).
