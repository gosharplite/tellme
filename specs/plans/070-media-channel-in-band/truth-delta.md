# Truth Delta: 070-media-channel-in-band

**Plan Package**: `specs/plans/070-media-channel-in-band`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` RUN (2026-09-20)** - `research.md` D1...D7 + **ADR 0040** + `techstack.md` MODIFY x1. Clarify **not escalated** (0 questions). The other owners record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` - *Image filesystem tool (`read_image`)* | the media **channel** is now **in-band** (a `tools.MediaPart` returned via the optional `tools.MediaTool` capability); the round-062 `context` collector (ADR 0032 D7a) is **removed**; `internal/infrastructure/tools` no longer imports `internal/domain/llm` - round 070 / **ADR 0040** | `spec.md` FR-001...FR-004; `research.md` D1/D2/D3 |
| ADD | `docs/decisions/0040-media-channel-in-band.md` (+ index row) | the in-band media channel; **completes ADR 0032 RF-062-10** + **delivers ADR 0039 RF-069-1**; the rejected shape recorded **settled** | `spec.md` SC-004; `research.md` D7 |
| MODIFY | `docs/decisions/0032-agent-image-vision.md` - `Status` + `D7a` + `§Forward RF-062-10` | `D7a` **SUPERSEDED** by ADR 0040; `RF-062-10` **FULLY DELIVERED** | the per-round annotation convention |
| MODIFY | `docs/decisions/0039-toolset-spec-capability-seam.md` - `Status` + `§Forward RF-069-1` | `RF-069-1` **DELIVERED** by ADR 0040 | the per-round annotation convention |
| NOOP (checked) | every other techstack row | no behaviour, no new dependency, no config key, no wire change | `spec.md` I-1/I-2/I-3 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §2 W1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-record change; media is never persisted (ADR 0032). | `plan.md` §2 W2 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` | No user-visible behaviour change - the 042/043/047/049/069 structural-round precedent; `reading-a-local-image.feature` stays green **unchanged**. | `plan.md` §2 W4; `spec.md` US2/SC-002 |
