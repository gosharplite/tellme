# Truth Delta: 006-rendered-output-and-raw-flag

**Plan Package**: `specs/plans/006-rendered-output-and-raw-flag`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added an `Output rendering` row (glamour v1.0.0 — default Markdown→ANSI; `WithStandardStyle` + `GLAMOUR_STYLE` + `WithEmoji`; LaTeX→Unicode sanitize; ADR-007 raw degradation) and a `Raw output flag` row (`-r`/`--raw`; rendering gated by `-r` alone); reworded the `Terminal detection` row (the stdout side uses the same probe, ready to gate tellme's own presentation, `UseColor = isTTY && !raw` — inert today since no chrome ships; a named pin); extended the `CLI flag parsing` row to list `-r/--raw`. | Round-006 research Decisions 1–3, 5–6 (renderer parity, render/raw gate, stdout probe, sanitization, degradation); Clarify Q1/Q2. |
| MODIFY | `specs/truth/techstack.md` | **Configuration**: added a `Rendered width` row (`WRAP_WIDTH` config + `TELL_ME_WRAP_WIDTH` env, env-over-file, `>= 0`, `0` = renderer default, rendered-only). | Round-006 research Decision 7; Clarify Q3. |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: extended the `E2E runner / step definitions` row (an ANSI-aware rendered-vs-raw output predicate) and the `Pure-helper unit tests` row (render/raw mode selection + wrap-width resolution). | Round-006 research Decisions 3, 4, 7 (unit-testable mode + width resolution; robust E2E assertions). |
| MODIFY | `specs/truth/techstack.md` | **Not Introduced Yet**: removed the "Markdown/ANSI renderer and the `-r` flag" bullet (now introduced); reworded the TUI-libraries bullet (glamour is now the direct renderer; `lipgloss` only transitive). | Round-006 research Decision 1 — the renderer is no longer deferred. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; round 006 changes only how the answer is displayed, not any API surface (the provider request shape is unchanged from round 004). | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 006 introduces no persisted or in-memory state; `WRAP_WIDTH` is configuration input, not state. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/rendering-the-answer.feature` | New interface feature — Rule *A successful turn renders the answer as formatted text by default* (1 Example). | Carries acceptance `rendering-the-answer` Rule 1 under the CLI `chat` module (default rendered output). |
| ADD | `specs/truth/features/cli/chat/controlling-the-rendered-width.feature` | New interface feature — Rule *Rendered output wraps at the configured width* + Rule *The rendered width applies only to rendered output* (1 Example each). | Carries acceptance `controlling-the-rendered-width` Rules 1–2 (wrap-width effect; rendered-only). |
| MODIFY | `specs/truth/features/cli/chat/piping-the-answer-out.feature` | Rule *A redirected answer is the answer text alone* → *A raw ("-r") answer is the answer text alone*; the three Examples now run with the raw flag, so the byte-exact/verbatim assertion is the raw contract; header records the FR-007 amendment. | Round-006 FR-007 amendment: under reference parity the default output is rendered, so the byte-fidelity assertion moves to the `-r` path. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added Given `the rendered width is "{width}"`, When `the operator starts tellme with the prompt "{prompt}" and the raw flag`, and Then rows `the captured standard output is the rendered answer, not its raw Markdown`, `the captured standard output is wrapped so that no line is wider than {width} columns`, `the captured standard output contains the answer on a single line`; reworded the output note for the rendered-vs-raw split. | Round-006 rendering + raw + width behaviour; the `chat` module is reused (the prompt-turn boundary). |
| MODIFY | `specs/truth/features/cli/configuration/starting-with-a-configuration.feature` | Added Rule *A run stops when the rendered width is invalid* (1 Example: a negative rendered width is refused). | Round-006 FR-006 (a negative `WRAP_WIDTH` is a configuration error); config validity lives in the `configuration` module. |
| MODIFY | `specs/truth/features/cli/configuration/dsl.md` | Added Given `a well-formed configuration "{config_path}" whose rendered width is "{width}"`. | Arrange the rendered-width configuration value for the invalid-width rule. |
| MODIFY | `specs/truth/features/cli/dsl.md` | Extended the `tellme explains on stderr that "{reason}"` row's `片語詞彙` to **ten** phrases — added the general `the configuration is invalid` (a well-formed file whose resolved values fail validation, e.g. a negative rendered width). | Round-006: the first non-provider validated config field needs a general configuration-invalid class phrase. |
