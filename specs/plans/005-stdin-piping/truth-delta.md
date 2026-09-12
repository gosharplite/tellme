# Truth Delta: 005-stdin-piping

**Plan Package**: `specs/plans/005-stdin-piping`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/techstack.md` | _To be recorded: the TTY-detection mechanism, the bounded stdin read, and the input/output-mode selection seam._ | Round-005 stdin piping + TTY-aware output contract (Clarify Q1–Q3). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/contracts/**` | _Expected NOOP — tellme has a single CLI end and no HTTP/OpenAPI surface of its own._ | To be confirmed by `/axb-api-plan`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/data/**` | _Expected NOOP — no persisted or in-memory state is introduced._ | To be confirmed by `/axb-data-plan`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending)_ | `specs/truth/features/cli/**` | _To be recorded: the piping feature/Rules and the DSL rows for piped stdin, the combined prompt, and the TTY-aware output contract._ | Carries the round-005 acceptance journeys (`features/acceptance/**`) under the CLI interface. |
