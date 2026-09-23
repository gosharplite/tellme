# System analysis — round 085 `085-dsl-topology-reconciliation`

## 1. Interfaces inventory

| Interface | Kind | Disposition |
| --- | --- | --- |
| The operator terminal end | `cli` | **carried forward** to its contract owner `/axb-dsl-refine` |

This is the **pure-CLI / documentation round** case: the round changes the truth-tree's DSL
**documentation rows** (and one `make help` string), not any API, data, or TUI surface. No planner
wave applies.

## 2. Planner delegation (`truth-delta.md`)

- `/axb-api-plan` → **NOOP** (no endpoints; a CLI end) — recorded.
- `/axb-data-plan` → **NOOP** (no persisted-state change) — recorded.
- `/axb-ui-plan` → **skipped** (no user-facing UX surface change; a plain line-oriented CLI) —
  recorded.
- `/axb-dsl-refine` → **owns** the change (`specs/truth/features/cli/**/dsl.md`).

## 3. Waves

| Wave | Order | Work |
| --- | --- | --- |
| W1 | 1 | `/axb-dsl-refine` reconciles the DSL topology (promote / restore / add) |
| W2 | 2 | `/axb-technical-research` retires the R8d truth bullet (+ the ADR 0012 Forward annotation) |

W2 is independent of W1 (different truth artifact) but lands in the same delivery.

## 4. Interface-level changes

None at the executable-contract level: the `.feature` files and the Go step definitions are
**unchanged**. The round changes the **rows** that map each step sentence to its implementation
semantics, and one `make help` line.

## 5. Domain model

**Not modelled** (ADR 0041 escape hatch): a DSL-row authority-level move + a `make help` string change
touch no modelled entity/invariant/scenario. `docs/domain-model/**` is **unchanged**; `modelith-check`
stays green.

## 6. Notes

- The topology audit is a **carried check**, not a `make verify` member (quality model,
  `topology-audit-not-a-gate`); this round reconciles its findings rather than gating it.
- The E2E contract remains the behaviour guard (I-2).
