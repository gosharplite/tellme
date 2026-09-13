# System Analysis Plan — round 010 (`010-stream-ordering-observability`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/010-stream-ordering-observability/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── ordering-the-payload-status-around-the-answer.feature
│       └── reporting-tool-activity-before-the-answer.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (merged-stream witness)
├── data/**                        # /axb-data-plan — NOOP this round (no persisted-state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/{*.feature, dsl.md}   # MODIFY — ordering semantics (not mere presence)
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **`NOOP`**: no persisted state changes — per-turn token counts remain display-only and the session model is untouched.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── cli/
│   ├── cli.go                     # unchanged — the emit order (write the answer, THEN emit the post-turn
│   │                              #   line) is already correct after 7bcb2d3; round 010 adds NO product change
│   └── turn_test.go               # EXTENDED — the unit emit-order assertion (the payload status brackets
│                                  #   the answer: pre-flight < answer < measured) at the `runTurn` layer
├── agent/agentloop.go             # unchanged — the live tool-loop log already precedes the answer
└── ...                            # unchanged — no product surface changes
                                     #   (no new domain port, no adapter, no config field, no dependency)

tests/e2e/
├── harness/cmd_helper.go          # CHANGED — add a merged-stream capture variant (the child's `stdout`
│                                  #   and `stderr` into ONE shared ordered buffer — the `2>&1` equivalent),
│                                  #   returning the merged output
├── steps/scenario_context.go      # CHANGED — run the ordering scenarios through the merged capture and
│                                  #   expose the merged output for assertions
└── steps/                         # NEW step files for the ordering scenarios (payload-status bracketing;
                                   #   tool-loop-precedes-answer)

specs/truth/features/cli/chat/
├── reporting-the-payload-status.feature   # MODIFY — assert the ordering (pre-flight < answer < measured)
└── dsl.md                                  # MODIFY — ordering semantics on the payload-status + tool-log rows

go.mod / go.sum                     # unchanged — stdlib-only (os/exec); no new dependency
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 010 is a **contract + oracle** round — it changes **no product code** (the payload-status emit order is already correct after `7bcb2d3`; the tool-loop log already precedes the answer). Every source change is **test-side**: the E2E harness (`tests/e2e/harness`) gains a **merged-stream capture variant** (both streams into one shared ordered buffer) and the scenario context runs the ordering scenarios through it; new step files assert the interleave; and the `runTurn` unit emit-order test is extended. The **ordering contract** itself is pinned in the CLI interface truth (`specs/truth/features/cli/chat/**`) by `/axb-dsl-refine`. The round introduces **no** new domain port, adapter, persisted-state change, config field, or dependency — consistent with `research.md` Decisions 1–8.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface. Round 010 does not introduce a new system end; it makes the **CLI end**'s diagnostic-output ordering an assertable contract.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **interleave** of the diagnostic stream (`stderr`) with the answer stream (`stdout`) becomes the contract: **(a)** the payload status **brackets** the answer (`pre-flight < answer < measured`); **(b)** the live tool-loop log **precedes** the answer. The answer stream (`stdout`) stays **byte-exact** (rounds 005/006 contracts); the frozen `tellme: {phrase}` class-phrase vocabulary and the exit-code table (`0/2/3/4/5/6/7`) are **unchanged**; non-prompt paths emit no ordering-relevant line. The interleave is observed end-to-end through a **merged-stream capture** (both streams into one ordered buffer — the `2>&1` equivalent), the oracle round 009 lacked. The round-006 degraded-render warning relative to the answer is **incidental** — documented as not guaranteed, **not** asserted.
   - Requirement evidence: `FR-001`–`FR-011`, `NFR-001`–`NFR-005`; acceptance features `ordering-the-payload-status-around-the-answer.feature`, `reporting-tool-activity-before-the-answer.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own). The outbound provider request and its response shape are unchanged this round.
> - `/axb-data-plan` = **`NOOP`** — no persisted state changes; the session-history model (`history_entry`) is untouched (research Decision 5/7).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **merged-stream witness** is **not** a separate system interface: it is test-harness tooling with no schema/contract artifact and no persisted state; its purpose (making the CLI end's interleave observable) belongs to the **CLI end**'s verification and is carried forward to `/axb-dsl-refine` (executable truth) and `/axb-tasks` (the witness task).
> - The **provider** remains an **external dependency reached outbound by the CLI**; round 010 adds no outbound request/response change, so it introduces no new tellme-authored contract.
> - **Truth amendments carried to `/axb-dsl-refine`**: **MODIFY** `specs/truth/features/cli/chat/reporting-the-payload-status.feature` so the payload-status **ordering** is asserted (an interface Rule, not prose); **MODIFY** `specs/truth/features/cli/chat/dsl.md` so the payload-status rows and the tool-loop log row pin **ordering** (`pre-flight < answer < measured`; tool-log precedes answer), not merely presence; and confirm **every round-010 acceptance rule is carried** by an interface feature (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10**; a `-l`/`-d`/`--version`/boot run stays byte-identical to rounds 001–009.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound provider request/response shape is unchanged.
  - **Data** → **`NOOP`**: no persisted-state change; token counts stay display-only and are recomputed.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the **ordering** in `specs/truth/features/cli/chat/**` — (a) the payload status brackets the answer (`pre-flight < answer < measured`), including the no-usage variant; (b) the tool-loop log precedes the answer — so the ordering is **mechanically assertable** (an interface Rule, not prose), while the answer stream stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: there is **one** interface and the behaviour is **already fixed** by `research.md` Decisions 1–8 and Clarify Round 1, so there is nothing to sequence; a single wave suffices. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a sole interface forms one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → pin the ordering semantics in `specs/truth/features/cli/chat/{reporting-the-payload-status.feature, dsl.md}` (payload-status bracketing; tool-loop-precedes-answer) and confirm `acceptance-coverage` for both round-010 acceptance features.

*Handoff payload (for the next phase)*: plan package `specs/plans/010-stream-ordering-observability`; truth root `specs/truth`; truth-delta `specs/plans/010-stream-ordering-observability/truth-delta.md`; interfaces `CLI end`; analysis focus as above; acceptance features `features/acceptance/ordering-the-payload-status-around-the-answer.feature` + `features/acceptance/reporting-tool-activity-before-the-answer.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the scope (payload-status + tool-loop ordering; degrade warning incidental), the witness (merged single-buffer capture), and the layering (unit + E2E); `research.md` Decisions 1–8 settled the witness mechanism, the ordering-as-truth decision, determinism, and the no-product-change/no-dependency posture. **Open (non-blocking):** the merged-capture plumbing API, the no-final-answer ordering wording, the incidental-interleave note location, and the unit emit-order helper shape — all held as research/DSL-level defaults, none gating this round.)*
