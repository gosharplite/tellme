# Truth Delta: 069-toolset-spec-capability-seam

**Plan Package**: `specs/plans/069-toolset-spec-capability-seam`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` RUN (2026-09-20)** — `research.md` D1...D7 + **ADR 0039** + `techstack.md` MODIFY x2. Clarify **not escalated** (0 questions). The other owners record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` - *Composition root (dependency injection)* | the agent tool-registry build func now takes **one named `deps.ToolSetSpec`** (the former positional `(sink, vision bool, providerType)` scalars are gone) - round 069 / **ADR 0039** | `spec.md` FR-001/FR-004; `research.md` D1/D2/D3 |
| MODIFY | `specs/truth/techstack.md` - *Image filesystem tool (`read_image`)* | the `ToolSetSpec` seam it named as a next-round refactor (ADR 0032 **RF-062-10** / F-062-4) is **landed by round 069 / ADR 0039**; only the media-channel half of RF-062-10 remains | `spec.md` SC-004; `research.md` D7 |
| ADD | `docs/decisions/0039-toolset-spec-capability-seam.md` (+ index row) | the structural ADR (the ADR 0013/0019/0021 lineage); **delivers RF-062-10** (the seam half) + **RF-063-6** | `spec.md` SC-004; `research.md` D7 |
| NOOP (checked) | every other techstack row | no behaviour, no new dependency, no config key, no wire change | `spec.md` I-1/I-2/I-3 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §2 W1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-record change (I-3); the spec is in-memory construction state. | `plan.md` §2 W2 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` | No user-visible behaviour change - the 042/043/047/049 structural-round precedent; the existing journeys (`offering-the-agent-tools`, `reading-a-local-image`, `accounting-for-the-tool-use`) stay green **unchanged**. | `plan.md` §2 W4; `spec.md` US2/SC-002 |
