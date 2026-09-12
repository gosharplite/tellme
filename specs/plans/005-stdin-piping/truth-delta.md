# Truth Delta: 005-stdin-piping

**Plan Package**: `specs/plans/005-stdin-piping`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: reworded the `Prompt input` row to cover piped standard input combined with the positional argument (`args (joined by spaces) + "\n"` + stdin, then trim; stdin read only on the prompt-turn path); added a `Terminal detection` row (stdlib `os.File` + `os.ModeCharDevice`, behind an injected seam) and a `Piped stdin read` row (`io.LimitReader`, fixed 1 MiB cap). | Round-005 research Decisions 1–3 & 6: the dependency-free TTY check, the bounded read, and the combine rule. |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner / step definitions` row (the subprocess runner injects a scripted stdin for piped scenarios) and the `Pure-helper unit tests` row (adds prompt combination + input/output-mode selection, folding issue #14). | Round-005 research Decisions 3–4: the piped-stdin E2E strategy and the injectable I/O-mode seam + unit tests. |
| MODIFY | `specs/truth/techstack.md` | **Not Introduced Yet**: added `golang.org/x/term` (replaced by the stdlib char-device check) and a renderer with the `-r`/raw-output flag (deferred; tellme's output already equals the reference's `-r` output). | Round-005 research Decisions 1 & 6 (Clarify Q1): explicit exclusions. |
| MODIFY | `specs/truth/techstack.md` | **CLI Application · Terminal detection**: reworded so the row no longer over-claims — the `isTTY` seam is a general stream probe wired to **stdin** this round; the **stdout** probe is wired when presentation is introduced (no presentation exists yet). | Round-005 PR #16 review finding **F2** (phantom `isTTY(stdout)` documentation) — reconciled docs to the implementation. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; round 005 changes only prompt ingestion and the output contract, not any API surface (the provider request shape is unchanged from round 004). | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 005 adds no persisted or in-memory state — it reads stdin into an ephemeral prompt string and prints the answer; no state model is introduced. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/piping-a-prompt.feature` | New interface feature — Rule *A prompt piped on standard input is sent to the selected provider and the answer printed* (2 Examples: piped question with no instruction; piped content with an instruction) + Rule *A pipe without prompt content makes no provider request* (2 Examples: pipes nothing; diagnostic flag precedence over piped input). | Carries acceptance `piping-a-prompt.feature` under the CLI `chat` module. |
| ADD | `specs/truth/features/cli/chat/piping-the-answer-out.feature` | New interface feature — Rule *A redirected answer is the answer text alone* (1 Example) + Rule *A piped run does not wait for interactive terminal input* (1 Example). | Carries acceptance `piping-the-answer-out.feature` under the CLI `chat` module. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added 4 `When` rows (`the operator pipes "{content}" into tellme`; `the operator pipes "{content}" into tellme with the instruction "{instruction}"`; `the operator pipes nothing into tellme`; `the operator runs tellme's diagnostic with "{content}" piped in`) and 5 `Then` rows (`the request carried the piped content "{content}"`; `the request carried the instruction "{instruction}" followed by the piped content "{content}"`; `the captured standard output is exactly "{answer}"`; `the captured standard output carries no terminal decoration`; `tellme completes the turn without waiting for terminal input`). | Round-005 piping behaviour + TTY-aware output contract; `chat` module reused (the prompt-turn boundary). Topology audit **PASSED** (261 steps). |
