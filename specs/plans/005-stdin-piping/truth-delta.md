# Truth Delta: 005-stdin-piping

**Plan Package**: `specs/plans/005-stdin-piping`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.
>
> **Grill-correction pass (pre-merge, PR #16 review):** the rows below include the round's original deltas *and* the corrective `MODIFY` rows from the architectural grill round (`architect` vs `griller`). The corrections fix: the `Prompt input` read-scope predicate (Q2), the decorative `必查` over-broad predicate (Q4), the output-predicate wording (Q5), the newly-expressible input class (Q6), and the environment class phrase's third cause with F1's "no truth change" claim retracted (Q7).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: reworded the `Prompt input` row to cover piped standard input combined with the positional argument (`args (joined by spaces) + "\n"` + stdin, then trim; standard input is read on the **non-explicit-mode dispatch path** — the reasoning turn **and** the empty→boot fall-through; `--version`/`-d` never read it); added a `Terminal detection` row (stdlib `os.File` + `os.ModeCharDevice`, behind an injected seam) and a `Piped stdin read` row (`io.LimitReader`, fixed 1 MiB cap). | Round-005 research Decisions 1–3 & 6: the dependency-free TTY check, the bounded read, and the combine rule. **Corrected (grill Q2):** the round's own behavior reads stdin on the empty→boot fall-through too, so "prompt-turn path" was a false predicate. |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner / step definitions` row (the subprocess runner injects a scripted stdin for piped scenarios) and the `Pure-helper unit tests` row (adds prompt combination + input/output-mode selection, folding issue #14). | Round-005 research Decisions 3–4: the piped-stdin E2E strategy and the injectable I/O-mode seam + unit tests. |
| MODIFY | `specs/truth/techstack.md` | **Not Introduced Yet**: added `golang.org/x/term` (replaced by the stdlib char-device check) and a renderer with the `-r`/raw-output flag (deferred; tellme's output already equals the reference's `-r` output). | Round-005 research Decisions 1 & 6 (Clarify Q1): explicit exclusions. |
| MODIFY | `specs/truth/techstack.md` | **CLI Application · Terminal detection**: reworded so the row no longer over-claims — the `isTTY` seam is a general stream probe wired to **stdin** this round; the **stdout** probe is wired when presentation is introduced (no presentation exists yet). *(Cross-ref, plan-side not truth: the same F2 withdrawal is annotated in `plan.md` and `tasks.md` Phase-4B Boundary in the grill-correction pass — Q1.)* | Round-005 PR #16 review finding **F2** (phantom `isTTY(stdout)` documentation) — reconciled docs to the implementation. |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner / step definitions` row to note the runner now expresses **newline/ANSI-bearing answers** (via the DSL escape convention) and runs piped turns under a **bounded deadline** (a hang fails explicitly rather than being inferred). **Not Introduced Yet**: added a **pty-capable harness** (real-TTY fidelity) as a deferred dependency. | Grill round **Q6** (root cause): the harness/DSL input surface made several pinned assertions unfalsifiable; the representable half is fixed, and the pty branch is named as unverifiable-this-round. |

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
| ADD | `specs/truth/features/cli/chat/piping-the-answer-out.feature` | New interface feature — Rule *A redirected answer is the answer text alone* (now **3** Examples: plain answer; newline-terminated answer; control-byte/ANSI-bearing answer) + Rule *A piped run does not wait for interactive terminal input* (1 Example). | Carries acceptance `piping-the-answer-out.feature` under the CLI `chat` module. **Extended (grill Q5/Q4):** the two new Examples make the output predicate falsifiable and demonstrate the narrowed decoration predicate does not misfire on a correct ANSI-bearing answer. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added 4 `When` rows (`the operator pipes "{content}" into tellme`; `the operator pipes "{content}" into tellme with the instruction "{instruction}"`; `the operator pipes nothing into tellme`; `the operator runs tellme's diagnostic with "{content}" piped in`) and 5 `Then` rows (`the request carried the piped content "{content}"`; `the request carried the instruction "{instruction}" followed by the piped content "{content}"`; `the captured standard output is exactly "{answer}"`; `the captured standard output carries no terminal decoration`; `tellme completes the turn without waiting for terminal input`). | Round-005 piping behaviour + TTY-aware output contract; `chat` module reused (the prompt-turn boundary). Topology audit **PASSED** (261 steps). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **Grill corrections:** narrowed the decoration row's `必查` from the content-blind "contains no ANSI escape sequences" to the actual FR-007 contract — *no presentation control codes introduced by tellme* (the redirected stream equals the answer bytes verbatim); disambiguated the `is exactly "{answer}"` row's `必查` to "the answer bytes verbatim, then exactly one newline **appended by the CLI**"; made the `completes the turn without waiting` row's `必查` assert a **bounded deadline**; documented the `\n`/`\t`/`\\`/`\xHH` parameter-escape convention in the module header. | Grill round **Q4/Q5/Q6**: the decoration `必查` contradicted its own `不該發生`/FR-007 and would misfire on a correct ANSI-bearing answer; the newline wording misnamed the predicate; and the falsifying input class is now expressible. |
| MODIFY | `specs/truth/features/cli/dsl.md` | Extended the `tellme explains on stderr that "{reason}"` row's `片語詞彙` so the environment phrase's cause enumeration records the **third** cause — standard input unreadable — alongside the not-usable and unset paths. | Grill round **Q7**: F1 attached a third cause to a frozen phrase whose enumerated mapping is truth; "standard input unreadable" is an *environment* condition, so it belongs under this phrase (exit `4` value unchanged, meaning widened). |
