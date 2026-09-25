# System analysis — round 090 `090-record-hygiene-offered-set-and-direction`

## 1. Interfaces inventory

| Interface | Kind | Disposition |
| --- | --- | --- |
| The operator terminal end | `cli` | **carried forward** to its contract owner `/axb-dsl-refine` |

This is a **record/direction round + a test-only carrier**: the changes are the `chat/dsl.md` offered-set
row (reconciled + carried), the `README.md` surface enumeration + direction pointer, **ADR 0061**, the
`STATUS.md` direction line, and the `techstack.md` `cobra` note. No API, data, or TUI surface changes.

## 2. Planner delegation (`truth-delta.md`)

- `/axb-api-plan` → **NOOP** (no endpoints; a CLI end) — recorded.
- `/axb-data-plan` → **NOOP** (no persisted-state change) — recorded.
- `/axb-ui-plan` → **skipped** (no user-facing UX surface change) — recorded.
- `/axb-dsl-refine` → **owns** `specs/truth/features/cli/chat/dsl.md` (the offered-set reconciliation).
- `/axb-technical-research` → **owns** `specs/truth/techstack.md` (the `cobra` note) and produces **ADR 0061**.

## 3. Waves

| Wave | Order | Work |
| --- | --- | --- |
| W1 | 1 | `/axb-dsl-refine` reconciles the offered-set row + the step comment; the (test-only) carrier lands with it |
| W2 | 2 | `/axb-technical-research` fixes the `cobra` note + writes **ADR 0061** + the README/STATUS direction pointers |

W1 and W2 touch different artifacts; both land in the same delivery.

## 4. Interface-level changes

None at the executable-contract level: the `.feature` files and the Go step definitions are **unchanged**
(the offered-set row's *sentence pattern* is identical; only its documented `集合`/prose cell changes). The
carrier is a unit test.

## 5. Domain model

**Not modelled** (ADR 0041 escape hatch): a doc *count* fix, a decision record, a comment, and a
test-only carrier touch no modeled entity/invariant/scenario. `docs/domain-model/**` is **unchanged**;
`modelith-check` stays green.

## 6. Notes

- The carrier rides `go test` (`make test`), **not** `make verify` — like the round-031/061 schema gates.
- The E2E contract remains the behaviour guard (I-3): unchanged counts, green.
- Frozen `specs/plans/NNN-*/**` history is left verbatim (`plan-package-frozen`).
