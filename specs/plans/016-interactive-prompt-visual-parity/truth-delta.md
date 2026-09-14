# Truth Delta: 016-interactive-prompt-visual-parity

**Plan Package**: `specs/plans/016-interactive-prompt-visual-parity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Interactive TUI prompt (`-i`)**: strict-parity chrome — a **bordered** editor (`NormalBorder`, fg `240`, height 10, width from the terminal, the reference placeholder) over a **styled suggestion list** (`Suggestions:` header; selected bold fg `205`/bg `235`, unselected fg `245`); **no** dashboard header / status line / `?` overlay. **Session dashboard**: the prompt header is **removed** from the `-i` surface (the figures stay on the round-009 status line). **Prompt suggestion engine**: the refresh is debounced + async + cancelable; over-3-line entries dropped; `Tab`/`Shift+Tab` **insert** the selection (last-token heuristic). **Testing & Verification**: the harness adds style/state assertions (no pty). No new dependency; POSIX-only. | Round-016 research Decisions 1–8: reproduce the reference chrome on the existing Bubble Tea family, debounce/cancel the suggestion refresh, insert-on-`Tab`, and retire the round-015 dashboard header (strict parity, issue #39). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and authors no OpenAPI/HTTP surface of its own; round 016 changes only the `-i` prompt's presentation/interaction — no tellme-owned request/response shape. | `contract-authoritative` holds vacuously — the round is a local terminal-presentation change. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 016 persists no new state — the shared global prompt log (`prompt_log_entry`) and every other data artifact are unchanged. | `data-model-covers-all-state` is already satisfied by round 015's model; the round adds no state. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| DELETE | `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` (a rule + its Example) | Retire the rule "the interactive prompt reports the session dashboard" — the dashboard header is removed from the `-i` surface (the submit/abort rules are unchanged). | Round 016 — strict parity (`FR-004`); the round-015 dashboard requirement is superseded. |
| ADD | `specs/truth/features/cli/chat/presenting-the-interactive-prompt.feature` | New atomic rules: the prompt is framed around a bordered editor and lists suggestions beneath; it shows no session metrics header; the editor advertises the submit/abort keys; it keeps a multi-line prompt and shows no line numbers; the editor frame stays within the terminal width. | Round 016 — carry the chrome parity into executable truth (`FR-001`–`FR-004`). |
| MODIFY | `specs/truth/features/cli/chat/prompting-with-suggestions.feature` | Added the "accepting a suggestion inserts it into the editor" rule and the "an over-long suggestion is not offered" rule. | Round 016 — `Tab`-inserts + the >3-line drop become executable truth (`FR-006`/`FR-007`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Add the chrome + insertion step vocabulary (**+11 rows**: framed editor; suggestions beneath; current-choice cursor; no metrics header; key hints; keeps typed text; no line numbers; bounded frame; holds accepted suggestion; no over-long suggestion; accept-selection When). **−2 rows** (the dashboard Thens). | Round 016 — the strict-parity Given/When/Then vocabulary (its only users are the `chat` features). |
