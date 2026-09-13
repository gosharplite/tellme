# Truth Delta: 012-interactive-multiline-prompt

**Plan Package**: `specs/plans/012-interactive-multiline-prompt`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added an *Interactive prompt read* row (read a multi-line prompt from a terminal to EOF, bounded 1 MiB, cancellable; hint on `stderr`; empty/cancel sends no request; POSIX-only) and extended the *Prompt input* row (the interactive path) and the *Terminal detection* row (the seam now also gates the reader). **Testing & Verification**: extended *E2E runner* (the negative assertion: a piped run prints no reading announcement), *Test strategy* (interactive read asserted at the unit layer via the seam + observable negatives E2E), and *Pure-helper unit tests* (the interactive read). **Not Introduced Yet**: noted the seam reuse (no `x/term`). Nothing else — one CLI end, no new dependency. | Round-012 research Decisions 1–8: add the interactive multi-line prompt reader (POSIX-only, reference-shaped) with no new dependency. **(Superseded by the review-response row below — B1: real isatty, ADR 0003.)** |

| MODIFY | `specs/truth/techstack.md` | **Review response (PR #31 B1/RF1/RF2):** *Terminal detection* is now `golang.org/x/term` (`IsTerminal`) — a **real isatty** replacing the `os.ModeCharDevice` heuristic (ADR 0003, `docs/decisions/0003-terminal-detection-isatty.md`), which was true for `/dev/null` and masked a missing-configuration failure; added the `TELL_ME_FORCE_STDIN_TTY` diagnostic seam (RF1). *E2E runner* + *Test strategy* updated (the interactive read is driven E2E; a pipe **and** the null device are pinned as non-terminal). *Not Introduced Yet* reconciled (`x/term` is now introduced; the stale `"/dev/null as a terminal"` stand-in text removed). | PR #31 architectural review — BLOCKER B1 + RF1/RF2 + the `techstack.md` self-inconsistency. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; the interactive prompt read is a CLI input concern recorded in `techstack.md`, not an authored contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. No persisted-state change — the interactive prompt read is input-only; the session-history model is untouched. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature` | New CLI `chat` feature pinning the one E2E-observable fact: a **piped** run prints **no** reading announcement (the interactive reader never engages on a non-terminal input); a header note documents the interactive read + empty/cancel contract as the round-005-style named pin (unit-verified via the terminal seam). | Round 012 — the interactive multi-line reader is a POSIX-terminal capability the pty-less harness cannot drive; pin what the E2E *can* observe and document the pin (`FR-004`, `FR-008`). **(Superseded by the review-response row below — RF1: E2E positive path via `TELL_ME_FORCE_STDIN_TTY`.)** |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added the Then row `no reading announcement is reported` and a round-012 note (interactive read + empty/cancel = a named pin; POSIX-only, no Windows variant). | Round 012 — carry the observable negative into the executable DSL and record the pin. |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the reading row is `chat`-module-specific; the frozen class-phrase vocabulary stays **10**. | No cross-module row was needed. |
| MODIFY | `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature` + `chat/dsl.md` | **Review response (PR #31 RF1/RF2/TD1):** the feature gains an **executable positive** Rule (the interactive read driven E2E via `TELL_ME_FORCE_STDIN_TTY`) and a **null-device** Rule (a non-terminal character device never announces) — carrying acceptance Rules 1–2 executably. `chat/dsl.md` gains the `the operator is working at an interactive terminal` Given, the `the operator starts tellme with the null device on standard input` When, and the `the reading announcement is reported on the diagnostic output` Then; the `no reading announcement is reported` row now asserts against the single-sourced `cli.MultiLineHint` (TD1). | PR #31 review — RF1 (close the positive-path blind spot), RF2 (pin the char-device path), TD1 (single-source the hint). |
| MODIFY | `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature` + `chat/dsl.md` | **Amendment A8 (in-round):** a prompt-less `--new` on a **terminal** now archives the session **first** and *then* engages the reader (unifying "start fresh and type"); a **non-terminal** prompt-less `--new` keeps its round-007 archive-and-exit behaviour. The feature gains the `A prompt-less --new at the terminal starts fresh, then reads` Rule; `chat/dsl.md` gains the `the operator starts a fresh session with "--new" and pipes "{content}"` When row. | In-round amendment requested by the operator: `tellme --new` should capture input rather than silently exit. |
