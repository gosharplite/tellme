# Truth Delta: 012-interactive-multiline-prompt

**Plan Package**: `specs/plans/012-interactive-multiline-prompt`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added an *Interactive prompt read* row (read a multi-line prompt from a terminal to EOF, bounded 1 MiB, cancellable; hint on `stderr`; empty/cancel sends no request; POSIX-only) and extended the *Prompt input* row (the interactive path) and the *Terminal detection* row (the seam now also gates the reader). **Testing & Verification**: extended *E2E runner* (the negative assertion: a piped run prints no reading announcement), *Test strategy* (interactive read asserted at the unit layer via the seam + observable negatives E2E), and *Pure-helper unit tests* (the interactive read). **Not Introduced Yet**: noted the seam reuse (no `x/term`). Nothing else — one CLI end, no new dependency. | Round-012 research Decisions 1–8: add the interactive multi-line prompt reader (POSIX-only, reference-shaped) with no new dependency. |

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
| ADD | `specs/truth/features/cli/chat/reading-a-multi-line-prompt.feature` | New CLI `chat` feature pinning the one E2E-observable fact: a **piped** run prints **no** reading announcement (the interactive reader never engages on a non-terminal input); a header note documents the interactive read + empty/cancel contract as the round-005-style named pin (unit-verified via the terminal seam). | Round 012 — the interactive multi-line reader is a POSIX-terminal capability the pty-less harness cannot drive; pin what the E2E *can* observe and document the pin (`FR-004`, `FR-008`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added the Then row `no reading announcement is reported` and a round-012 note (interactive read + empty/cancel = a named pin; POSIX-only, no Windows variant). | Round 012 — carry the observable negative into the executable DSL and record the pin. |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the reading row is `chat`-module-specific; the frozen class-phrase vocabulary stays **10**. | No cross-module row was needed. |
