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
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and authors no OpenAPI/HTTP surface of its own; round 017 changes only a non-TUI turn's presentation — no tellme-owned request/response shape. | `contract-authoritative` holds vacuously — the round is a local terminal-presentation change. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 017 persists no new state — `history.jsonl` / `history.archive.jsonl` and the shared prompt log (`prompt_log_entry`) are unchanged; the turn number is derived from the already-loaded history, not stored. | `data-model-covers-all-state` is already satisfied; the round adds no state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/presenting-the-turn.feature` | New feature carrying the turn-chrome contract as atomic rules: a prompt run announces the captured input (argument · pipe · terminal reader); the turn opens with an 80-column rule and a `╭─⠿ Turn <N> - <mode>` header over the pre-flight payload line; the header counts the session's turns; the frame is separated from the answer; and the chrome appears **only** on the non-interactive prompt turn (no chrome on the `-i` prompt or a non-prompt run). | Round 017 — carry the acceptance journeys into executable interface truth (`FR-001`–`FR-008`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Add **6 Then rows**: `the input capture is announced for the turn`, `… before the turn frame`, `the turn opens with a horizontal rule`, `the turn is headed "Turn {number}" for the active mode`, `the turn chrome is shown before the answer`, `the turn frame is separated from the answer`. | Round 017 — the turn-chrome step vocabulary (its users are the `chat` features). |
| MODIFY | `specs/truth/features/cli/dsl.md` | Add **1 root Then row**: `the run shows no turn chrome` — the chrome-absence predicate is used by the `chat`, `diagnostics`, and `history` modules, so it lives at the interface root (single authority). | Round 017 — the negative boundary is cross-module (PR #41 review **D2**); no new class phrase (still 11). |
| MODIFY | `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | Add `the run shows no turn chrome` to the version Example. | Round 017 — carry "no chrome on `--version`" (PR #41 review **D2**). |
| MODIFY | `specs/truth/features/cli/history/starting-a-fresh-session.feature` | Add `the run shows no turn chrome` to the prompt-less `--new` Example. | Round 017 — carry "no chrome on a prompt-less `--new`" (PR #41 review **D2**). |
