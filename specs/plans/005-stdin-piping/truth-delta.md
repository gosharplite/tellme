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
