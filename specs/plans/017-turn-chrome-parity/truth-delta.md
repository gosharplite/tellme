# Truth Delta: 017-turn-chrome-parity

**Plan Package**: `specs/plans/017-turn-chrome-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Operator turn chrome**: add the row documenting the reference's non-TUI per-turn chrome — the input-capture acknowledgement `[HH:MM:SS] Input captured. Processing...`, a leading blank line, an 80-column `─` rule, the `╭─⠿ Turn <N> - <mode>` header (`<N>` = persisted turns + 1), the round-009 pre-flight payload line, and a trailing blank line — emitted on the positional/piped prompt turn and the round-012 plain reader **only** (the `-i` TUI surface and non-prompt paths unchanged; plain text, no ANSI; the payload line's text unchanged; `stdout` byte-exact). | Round-017 research Decisions 1–7: reproduce the reference's turn chrome structurally on a hand-written `internal/ui` formatter, no new dependency; the must-ask questions stay settled. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/contracts/**` | <filled by /axb-api-plan> | <why> |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/data/**` | <filled by /axb-data-plan> | <why> |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| (pending) | `specs/truth/features/cli/**` | <filled by /axb-dsl-refine> | <why> |
