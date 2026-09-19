# Technical Research — round 057 (`057-tool-chrome-colour-and-payload-delta`)

**Owner of this artifact**: `/axb-technical-research` (plan-side) + a `specs/truth/techstack.md` **MODIFY** (truth-side).
**Inputs**: `spec.md` (US1 colour · US2 payload delta · US3 cap), the clarify locks (Q1 → 2, Q2 → 1, Q3 → 1), the measured current system (`dev` @ `63a54bc`).

## Decisions

| # | Decision | Rationale |
| --- | --- | --- |
| **D1** | **Extend the round-054 palette in `internal/ui/colour.go`.** Add `colorGray = "\033[0;90m"` and `colorYellow = "\033[0;33m"` (the reference's codes) and factor `green`/`grey`/`yellow` over one `wrap(s, code, enabled)` helper (empty string never wrapped; disabled ⇒ verbatim). | One policy, one place; the round-054 gate (terminal **and** `-r` off) is reused unchanged — so off-terminal/`-r` bytes stay byte-identical (I-2). |
| **D2** | **The grey `[Tool Output]` accent is threaded through the progress factory.** `render.ProgressFactory` gains a `colour bool` param; `ui.NewTurnProgress` → `NewToolOutputCoordinator(…, colour)` → `ToolOutputWriter.Colour`. `Begin` writes the grey header + grey opening separator; `EndWith` writes the grey closing separator (after the unconditional reset). | The block is written by `internal/ui` directly (not via `render.Lines`), so the flag must ride the factory seam. The plain path returns the pinned literals, keeping the byte-identical guarantee. |
| **D3** | **The yellow `[Tool Action]` accent lives with the other tool lines.** `formatToolActionColour(t, tool, args, colour)` wraps the already-sanitized `FormatToolAction`; `ToolLineRenderer.ActionLine` applies it with the adapter's colour flag. | Mirrors the round-054 `formatToolReasonColour`/`ReasonLine` pattern; the wrap is outside the sanitizer, so an argument cannot smuggle a control byte out of it. |
| **D4** | **`argValueCap` 189 → 500.** One constant; the mechanic is unchanged (fold → sanitize → rune-boundary cap, one U+2026 inside the cap; keys/list uncapped). | Operator request. Extends the round-034 divergence (reference: 189 **bytes**); the magnitude now diverges too — recorded. |
| **D5** | **A distinct `PayloadEstimate` port method** — `render.Lines.PayloadEstimate(t, tokens, delta, mode, model)`, rendered by `FormatPayloadEstimate` as `[HH:MM:SS] Payload: +<delta> ~<tokens> tokens - <mode> - <model>` (round-054 green MODE accent; estimated token plain). `PayloadStatus` (measured) is **untouched**. | Keeps one shape per line and the measured line byte-identical (Q1 → 2). A delta parameter on the shared method would have made the measured path carry an unused value. |
| **D6** | **The delta baseline is in-memory and session-scoped** — a `prevEstimate`/`haveEstimate` pair on `callRenderer` (one renderer per process = per turn). No predecessor ⇒ `+0` (A7); a smaller payload ⇒ the true signed value (A8). | Q2 → 1. A fresh process has no predecessor; `--new`/restart reset it naturally. **No data-model change** ⇒ `/axb-data-plan` NOOP. |
| **D7** | **`turns.log` needs no change.** The per-call `emit(newline, build)` already renders each line twice from the **same** closure — the coloured leg (`r.lines`) and the plain file leg (`r.fileLines`) — so the file inherits the new content (the delta-bearing estimated line) with no escapes. | Q3 → 1 (content parity, plain). Confirms the round-054 B-54-1 "rendered plain by construction" rule holds for the new content. |
| **D8** | **Governance: ADR 0027.** Records the two palette elements, the cap re-parameterisation, and the increment semantics, with a §Forward. | House convention (each behaviour-bearing chrome round gets an ADR). |
| **D9** | **No contract/data change.** `/axb-api-plan` = **NOOP** (no API surface); `/axb-data-plan` = **NOOP** (the delta baseline is in-memory; no persisted state moves). | The round is presentation-only plus one constant. |

## Truth impact (owner: `/axb-technical-research`)

**MODIFY** `specs/truth/techstack.md`:
- **Agent tool loop** row — the argument-value cap 189 → **500** (cap set `{500, 200, 200}`); the `[Tool Action]` line gains the terminal-gated yellow accent.
- **Turn chrome (operator)** row — the colour element list gains (v) the grey `[Tool Output]` header + both separators and (vi) the yellow `[Tool Action]` line; the pre-flight payload line becomes the signed-increment form (budget dropped).

*(`/axb-api-plan` NOOP · `/axb-data-plan` NOOP · `/axb-dsl-refine` MODIFY — see `truth-delta.md`.)*

## Residual risks / forward items (RF-057-x)

- **RF-057-1** — the delta token is **plain** (no new colour); colouring it is a small follow-on.
- **RF-057-2** — only argument **values** are capped; the `[Tool Action]` **line** (many args / long keys) remains uncapped, and a 500-rune cap enlarges the worst case.
- **RF-057-3** — no non-terminal colour mode (`--color=always`) — the gate stays terminal-only (round-054 RF-54-2).
- **RF-057-4** — the `[Tool Output]` **content** lines stay plain (A3); the reference tints its whole block.
- **RF-057-5** — a resumed session's first estimate is `+0` (no persisted predecessor); if a real delta across processes is ever wanted it needs a persisted field (rejected here as disproportionate).

## Verification plan

- **Unit**: `internal/ui` (the new wrappers; the grey header/separators; the yellow action line; `formatPayloadEstimateColour`; the 501-rune cap), `internal/cli` (the delta chaining: `+0` then a positive delta; the file leg plain).
- **E2E** (`specs/truth/features/cli/chat/**`): the grey frame + yellow action on a terminal; the estimated `+<delta> ~<n>` with no allowance, `+0` first then a positive delta; the 500-rune action cap; `turns.log` plain; the no-colour path unchanged.
- **Falsifiability witnesses** (reproduce → revert): (a) un-gate the colour ⇒ the no-colour path reds; (b) drop the delta ⇒ the increment Example reds; (c) cap 500 → 499 ⇒ the cap Example/pin reds.
