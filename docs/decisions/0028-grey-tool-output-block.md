# ADR 0028 — Every `[Tool Output]` line is grey (the whole block)

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0027](0027-tool-chrome-colour-and-payload-delta.md) (the round-057 grey that this ADR **extends** — specifically its assumption **A3**, which this ADR supersedes), [ADR 0023](0023-list-default-and-chrome-colour.md) (the terminal colour gate), [ADR 0008](0008-terminal-safe-lines-and-blank-line-grouping.md) (the sanitize policy the grey wraps around), round 034 (`specs/plans/034-tool-call-log-parity` — the block's shape + the drop-a-partial rule)

## Context

Round 057 greyed the `[Tool Output]` **header** line and the **two** horizontal separators and deliberately left the streamed **content** lines plain (its assumption **A3**). The operator then asked, in a follow-up session:

> *"I want all the `[Tool Output]` lines grey, not just the first line and the two horizontal lines."*

So the block currently reads as a grey frame with plain content inside it; the operator wants the **whole block** grey.

## Decision

**D1 — Every `[Tool Output]` line is grey on a colour-enabled stream.** The header, **each streamed content line**, and **both** separators are wrapped in the reference's grey `\033[0;90m…\033[0m`. This is implemented beside the round-057 wraps: `internal/ui/tooloutput.go` gains `formatToolOutputLineColour(t, line, colour)` (which the writer's `WriteWith` calls with the block's `Colour` flag), so the content path consults the **same** flag as the header and separators.

**D2 — The wrap is applied OUTSIDE the sanitizer.** `FormatToolOutputLine` sanitizes the content first (`sanitizeControl`), and the grey pair wraps that result. Consequence: a command's control bytes are removed **before** the wrap is added, so (i) no command output can escape the wrapper and (ii) the grey pair is the line's **only** escape. The round-038 sanitize policy is unchanged; the round-057 wording *"the sanitizer governs the model-authored value, not the whole line on a terminal"* generalises to this block.

**D3 — Nothing else changes.** The content **text** (timestamp/prefix framing + the sanitize result) is byte-identical; the neutral-close restore (`\x1b[0m` before the closing separator) is unchanged; the plain path (off a terminal / `-r`) is byte-identical; the file leg (`turns.log`/`-t`) stays plain; the yellow `[Tool Action]`, the 500-rune cap, and the payload increment are untouched.

**D4 — The trailing partial line stays dropped.** A partial line (no terminating newline) is never printed (round-034 FR-010), so there is nothing to grey; flushing it would be a behaviour change outside the request. *(The operator-vetoable assumption A1.)*

**D5 — Round 057's assumption A3 is superseded; the round-057 package stays frozen.** This ADR records the reversal. ADR 0027's body is **immutable** (an `Accepted` ADR is not edited); the extension is recorded here and linked from ADR 0027's lineage and from `specs/truth/techstack.md`.

## Consequences

- **Positive**: the `[Tool Output]` block reads as one visually distinct region at a terminal, which is what the operator asked for; the change is one formatter + one call site.
- **Unchanged guarantees**: the sanitizer boundary (the wrap cannot be escaped), the neutral close, the plain file leg, and off-terminal/`-r` byte-identity.
- **Recorded divergence**: the **element set** is tellme's own (the reference greys its whole output block too, but with a different sanitizer/neutral-close lineage).
- **Forward** (RF-058-x): (1) the content text is **not** rune-capped (unchanged — a very long line is grey in full); (2) the trailing partial line stays dropped (a future flush must decide its colour in the same change); (3) the block is stderr-only — a future turn-log inclusion of `[Tool Output]` (round-053 RF-53-1) must render the file leg plain.

## Alternatives considered

- **Also flushing the trailing partial line** (grey it): rejected — a round-034 FR-010 behaviour change outside the request (D4).
- **Wrapping inside the sanitizer**: rejected — the sanitizer removes control bytes, so wrapping inside would be stripped; the wrap must sit outside (D2).
- **A per-block "whole region" escape (one wrap around all lines)**: rejected — the writer emits line-by-line under a mutex with per-line clears from the coordinator, so per-line wraps are the correct, race-free form.
