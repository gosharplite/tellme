# Truth Delta: 064-prompt-suggestion-parity

**Plan Package**: `specs/plans/064-prompt-suggestion-parity`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **Clarify OPEN — Q1** (the scope of "behaviour only": pool-only vs. the reference's other source behaviours) is pending the operator's answer. Owner rows are recorded once their phases run.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/techstack.md` — *Prompt suggestion engine* | *(to be recorded: the candidate-pool depth 10 → newest 50; the cap stays 10)* | `spec.md` FR-001…FR-004, SC-001…SC-003 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/` (**no `contracts/**`**) | expected **NOOP (checked)** — single CLI end; no OpenAPI/HTTP surface changes. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/data/data-model.dbml` | expected **NOOP (checked)** — the shared prompt-log record shape is unchanged; only the read depth into it changes. | `spec.md` A3 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| *(pending)* | `specs/truth/features/cli/chat/prompting-with-suggestions.feature` (+ `chat/dsl.md`) | *(to be recorded: a depth-distinguishing Example — a prompt beyond the newest 10 but within the newest 50 is offered)* | `spec.md` US1; `acceptance-coverage` |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): expected unchanged unless a suggestion-source fact's wording changes; `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
