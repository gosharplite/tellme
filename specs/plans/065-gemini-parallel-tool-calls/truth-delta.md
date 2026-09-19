# Truth Delta: 065-gemini-parallel-tool-calls

**Plan Package**: `specs/plans/065-gemini-parallel-tool-calls`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **initialized by `/axb-specify`** (2026-09-20). Clarify **not escalated** (0 questions). Owner rows below are **placeholders** to be filled by each owner skill when its phase runs — do not treat any as recorded yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` — *Gemini/Vertex provider transport* (exact row name to be confirmed by the owner) | the batched-function-response rule: a model round's N `functionResponse` parts share one `user` turn; media turns follow. | `spec.md` FR-001/FR-002/FR-004, S-1/S-2/S-3 |
| _(pending)_ | `docs/decisions/00NN-*.md` (+ index) | a new ADR recording the placement decision (supersedes ADR 0033 **RF-063-7**; annotates ADR 0033 **D2**'s multi-call scope note). | `spec.md` A4, S-4/S-5 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(expected)_ NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | no API/HTTP surface exists or changes. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(expected)_ NOOP (checked) | `specs/truth/data/data-model.dbml` | no persisted shape changes (the `Turn` line is unchanged; media is in-flight only — ADR 0032 D7). | `spec.md` A3 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending, likely MODIFY)_ | `specs/truth/features/cli/chat/*.feature` + `chat/dsl.md` | a multi-call round Rule/Example over the fake provider (the observed shape: the round completes, the model sees both images); or NOOP if the carrier is unit-level — the owner decides. | `spec.md` US1/US2, FR-001…FR-009, A3 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): to be reviewed by the owner; likely **NOOP** (this round changes no modelled entity/invariant — the wire serialization is not a domain entity).

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
