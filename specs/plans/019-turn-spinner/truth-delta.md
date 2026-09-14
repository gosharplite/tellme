# Truth Delta: 019-turn-spinner

**Plan Package**: `specs/plans/019-turn-spinner`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Add the **Turn progress spinner (operator)** row (CLI Application): a hand-written `internal/ui` live spinner on `stderr` (`{frame}{status} ({n}s)`, ~200 ms braille frames), phase labels (` Thinking [<model>]...` / ` Executing [<tool>]...` / ` Executing tools [<a>, <b>]...`) + the tool-execution ` [CPU: … \| MEM: …]` segment. **Wire the standard-output terminal probe** in the Terminal detection row (the spinner is tellme's first own presentation chrome — closing round-006 / PR #16 Obs 1) and add the `TELL_ME_FORCE_STDOUT_TTY` seam. Extend the Testing & Verification rows (round-019 spinner formatter + CPU/mem samplers; E2E via the forced stdout seam) and the pty bullet under Not Introduced Yet. | Round-019 research D1–D9: reproduce the reference's spinner as a dependency-free hand-written presenter (labels + host CPU/mem), gated by the newly-wired standard-output probe. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | _pending_ | _pending_ |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/**` | _pending_ | _pending_ |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/**` | _pending_ | _pending_ |
