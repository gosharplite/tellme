# Truth Delta: 019-turn-spinner

**Plan Package**: `specs/plans/019-turn-spinner`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | Add the **Turn progress spinner (operator)** row (CLI Application): a hand-written `internal/ui` live spinner on `stderr` (`{frame}{status} ({n}s)`, ~200 ms braille frames), phase labels (` Thinking [<model>]...` / ` Executing [<tool>]...` / ` Executing tools [<a>, <b>]...`) + the tool-execution ` [CPU: … \| MEM: …]` segment. **Gate the spinner on the diagnostic stream** in the Terminal detection row (`isatty(stderr) && !-r`; the spinner is a `stderr` diagnostic like the reference) and add the `TELL_ME_FORCE_STDERR_TTY` seam — **no** standard-output probe is wired (round-006 / PR #16 Obs 1 stays OPEN). Add the **System metrics provider (telemetry)** row (a domain port + `internal/infrastructure/telemetry` POSIX adapters). Extend the Testing & Verification rows (round-019 spinner formatter + machine-wide CPU/mem samplers; E2E via the forced stderr seam) and the pty bullet under Not Introduced Yet. | Round-019 research D1–D9: reproduce the reference's spinner as a dependency-free hand-written presenter (labels + **machine-wide** host CPU/mem), gated on the diagnostic stream (`stderr`) like the reference. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. | Standalone CLI; the spinner is terminal presentation — no OpenAPI/HTTP surface of tellme's own, and no request/response shape is authored. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. | The spinner is ephemeral runtime presentation — no persisted state, no entity/field/lifecycle/store. It reads only the active model name (config) and the current tool-call names. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` | New feature carrying the turn-spinner contract as atomic Rules: presence while waiting, the model label + elapsed counter, the tool label(s) + the resources segment, the line yielding to the answer, and the terminal gate. | Round 019 — carry the three acceptance journeys into executable interface truth (`FR-001`–`FR-011`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Add **2 Given rows** (`the operator is watching a terminal` — the round-019 stdout-terminal seam; the one-round two-tool read) and **7 Then rows** (spinner presence; the model label; the elapsed; the single-tool label; the several-tool label; the resources segment; the clearing) + the round-019 note. | Round 019 — the turn-spinner step vocabulary (its users are the `chat` features). |
| MODIFY | `specs/truth/features/cli/dsl.md` | Add **1 root Then row** `the run shows no progress spinner` — the spinner-absence predicate is used by the `chat` (`-i`), `diagnostics`, and `history` modules, so it lives at the interface root (single authority). | Round 019 — the negative boundary is cross-module; no new class phrase (still 11). |
| MODIFY | `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | Add `the run shows no progress spinner` to the version Example. | Round 019 — carry "no spinner on `--version`". |
| MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | Add `the run shows no progress spinner` to the `-l` listing Example. | Round 019 — carry "no spinner on `-l`". |
| MODIFY | `specs/truth/features/cli/history/starting-a-fresh-session.feature` | Add `the run shows no progress spinner` to the prompt-less `--new` Example. | Round 019 — carry "no spinner on a prompt-less `--new`". |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-turn.feature` | Add `the run shows no progress spinner` to the `-i` interactive-prompt Example. | Round 019 — carry "no spinner on the `-i` surface". |
| MODIFY | `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | Add the **failure Rule** `A failed turn still reports the failure and leaves no progress indicator` (spinner cleared + class phrase unchanged on a failing turn). | Round 019 review TD2 — close the `acceptance-coverage` gap for `FR-009` / `SC-004` (the failed-run spinner-clear had no interface carrier). |
