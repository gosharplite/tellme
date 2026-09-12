# Truth Delta: 006-rendered-output-and-raw-flag

**Plan Package**: `specs/plans/006-rendered-output-and-raw-flag`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: added an `Output rendering` row (glamour v1.0.0 — default Markdown→ANSI; `WithStandardStyle` + `GLAMOUR_STYLE` + `WithEmoji`; LaTeX→Unicode sanitize; ADR-007 raw degradation) and a `Raw output flag` row (`-r`/`--raw`; rendering gated by `-r` alone); reworded the `Terminal detection` row (the seam is now wired to **stdout** as well, gating tellme's own presentation, `UseColor = isTTY && !raw`); extended the `CLI flag parsing` row to list `-r/--raw`. | Round-006 research Decisions 1–3, 5–6 (renderer parity, render/raw gate, stdout probe, sanitization, degradation); Clarify Q1/Q2. |
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
| _(pending)_ | `specs/truth/features/cli/**` | | |
