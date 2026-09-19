# Technical Research — round 058 (`058-grey-tool-output-content`)

**Owner of this artifact**: `/axb-technical-research` (plan-side) + a `specs/truth/techstack.md` **MODIFY** (truth-side).
**Inputs**: `spec.md` (US1), assumption A1 (the trailing partial line), the measured current system (`dev` @ `50ace10`).

## Decisions

| # | Decision | Rationale |
| --- | --- | --- |
| **D1** | **Wrap the content line in the writer, beside the header/separator wraps.** `internal/ui/tooloutput.go` gains `formatToolOutputLineColour(t, line, colour)` = `grey(FormatToolOutputLine(t, line), colour)`; `WriteWith` calls it with `w.Colour`. | The block is written by `internal/ui` directly (not via `render.Lines`), and the round-057 `Colour` flag already rides the writer — so the content path consults the same flag as the header/separators. One `wrap`, one owner. |
| **D2** | **The wrap is applied OUTSIDE the sanitizer.** `FormatToolOutputLine` sanitizes first (`sanitizeControl(line)`); `grey(...)` wraps that result. | A command's control bytes are removed **before** the grey pair is added, so (i) no output can escape the wrapper and (ii) the wrap is the line's **only** escape (the round-038 policy is unchanged, and the round-057 "the sanitizer governs the value, not the whole line" wording generalises cleanly). |
| **D3** | **No text change.** The timestamp/prefix framing and the sanitize result are byte-identical; only the surrounding escapes differ. | The request is about colour only (I-1); the content text is asserted unchanged. |
| **D4** | **The neutral-close restore is unchanged.** `EndWith` still writes `\x1b[0m` then the grey closing separator. | Round 038's invariant is orthogonal to which lines are grey; the reset must still precede the closing separator. |
| **D5** | **The plain path is byte-identical.** `colour=false` returns the plain line verbatim (one `wrap`). | The round-054 gate already guarantees off-terminal/`-r` byte-identity; nothing here can regress it. |
| **D6** | **The file leg stays plain.** The per-call `emit` renders each chrome line twice from one closure — the coloured diagnostic leg and the plain file leg — but the `[Tool Output]` block is written by the coordinator **only** to the diagnostic stream (`stream`), never to `turns.log`. | Round 053/054 rule (S-4); the block was never in the file leg, so nothing changes there. |
| **D7** | **The trailing partial line stays dropped** (Q1 → A). A partial line is never printed, so there is nothing to grey; flushing it would be a round-034 FR-010 behaviour change, out of the request's spirit. | A1; recorded as an operator-vetoable assumption. |
| **D8** | **Governance: a new ADR 0028** (the round-057 ADR 0027 body is immutable) recording the block-wide grey and **superseding round 057's A3**; the round-057 **package** stays frozen. | House convention; a reversal of a recorded assumption must be recorded, not silently applied. |
| **D9** | **No contract/data change.** `/axb-api-plan` = **NOOP**; `/axb-data-plan` = **NOOP**. | Presentation-only. |

## Truth impact (owner: `/axb-technical-research`)

**MODIFY** `specs/truth/techstack.md`:
- **Turn chrome (operator)** row — element (v) becomes the whole `[Tool Output]` **block** grey (header + every content line + both separators); round 058 recorded.
- **Agent tool loop** row — the colour qualification: the grey now covers the streamed content lines too (the sanitizer still governs the value; the wrap is outside it).

*(`/axb-api-plan` NOOP · `/axb-data-plan` NOOP · `/axb-dsl-refine` MODIFY — see `truth-delta.md`.)*

## Residual risks / forward items (RF-058-x)

- **RF-058-1** — the content text is **not** rune-capped (unchanged); a very long content line is grey in full.
- **RF-058-2** — the trailing partial line stays dropped (D7); a future round wanting it flushed must decide its colour in the same change.
- **RF-058-3** — the block is stderr-only; a future turn-log inclusion of `[Tool Output]` (round-053 RF-53-1) must render the file leg plain.

## Verification plan

- **Unit** (`internal/ui`): the content line is grey when enabled and **byte-identical plain** when off; the header/separators unchanged; the round-038 sanitizer still removes a command's escape, with the grey pair as the line's only escape.
- **E2E** (`specs/truth/features/cli/chat/**`): the `the tool output frame is shown in grey` Example now also requires **content** lines grey (≥2 grey `[Tool Output] …` lines); the no-colour path unchanged.
- **Falsifiability witness** (reproduce → revert): drop the content wrap ⇒ the grey Example / the unit pin reds.
