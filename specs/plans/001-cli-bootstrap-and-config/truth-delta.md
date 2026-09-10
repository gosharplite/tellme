# Truth Delta: 001-cli-bootstrap-and-config

**Plan Package**: `specs/plans/001-cli-bootstrap-and-config`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/techstack.md` | To be recorded by `/axb-technical-research` | Skeleton initialized by `/axb-specify`; the Go toolchain and test stack for this round are set here. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/contracts/**` | To be recorded by `/axb-api-plan` | CLI has no OpenAPI surface; expected `NOOP`, to be confirmed by the owner. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/data/**` | To be recorded by `/axb-data-plan` | Configuration is a slice-local input, not persisted state; expected `NOOP`, to be confirmed by the owner. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/features/**` | To be recorded by `/axb-dsl-refine` | The CLI contract (config resolution, workspace init, version, diagnostics) is decomposed into executable interface Gherkin + DSL here. |
