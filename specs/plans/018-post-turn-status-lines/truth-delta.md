# Truth Delta: 018-post-turn-status-lines

**Plan Package**: `specs/plans/018-post-turn-status-lines`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Post-turn status lines**: add the metrics-line + `╰─⠿ Ready` chrome (CLI Application) and the per-mode `tokens.log` usage log; add the **config-only `MODELS` pricing** row (Configuration); widen the provider **usage** with `cached`/`reasoning` tokens (Reasoning & Provider Transport); extend the round-018 test coverage (unit formatters + fake-provider usage details/multi-call); retire the "post-turn metrics line not introduced" bullet. | Round-018 research Decisions 1–9: reproduce the reference's post-turn status on a hand-written `internal/ui` formatter, dependency-free, with a config-only pricing table and a per-call usage log. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
|  |  |  |  |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
|  |  |  |  |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
|  |  |  |  |
