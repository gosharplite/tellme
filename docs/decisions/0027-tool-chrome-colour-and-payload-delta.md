# ADR 0027 — Tool-chrome colour accents, a 500-rune `[Tool Action]` cap, and a pre-flight payload increment

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0023](0023-list-default-and-chrome-colour.md) (the round-054 chrome-colour policy this ADR extends), [ADR 0022](0022-offline-session-config-and-turns-log.md) (the `turns.log` file leg this ADR keeps plain), [ADR 0005](0005-tool-call-log-parity.md) (the decomposed tool-log rendering + the rune-safe caps this ADR re-parameterises), [ADR 0008](0008-terminal-safe-lines-and-blank-line-grouping.md) (the sanitize policy the colour wraps around), [ADR 0014](0014-yield-policy-owner.md) (the `[Tool Output]` coordinator this ADR re-parameterises), round 057 (`specs/plans/057-tool-chrome-colour-and-payload-delta`)

## Context

The operator inspected `tell-me-go`'s palette and asked for three chrome changes, then answered three clarify questions:

1. **`[Tool Output]` and the two horizontal lines will be grey. `[Tool Action]` will be yellow.** (the reference's `colorGray = "\033[0;90m"`, `colorYellow = "\033[0;33m"`)
2. **`argValueCap` will become 500.** (today 189)
3. **`~203148/1000000` will become `+100 ~203148`** — *"`+100` is the increment of last `~203048`"* — and, on a later remark, **"`-t` needs to still be the same as what is shown on the terminal."**

Clarify (one question at a time) settled: **Q1 → only the estimated pre-flight line changes** (the measured line keeps `<tokens>/<budget>`); **Q2 → the increment baseline is the last estimate emitted in-process, held in memory** (no persistence); **Q3 → `turns.log` stays content-equal but plain**. Two boundary forms were exposed as operator-vetoable assumptions: no predecessor ⇒ `+0`; a shrunken payload ⇒ the true signed value.

## Decision

**D1 — Extend the round-054 colour policy with two more elements.** `internal/ui/colour.go` gains `colorGray` and `colorYellow` (the reference's codes) plus `grey`/`yellow` wrappers, factored over a shared `wrap` helper with `green`. The **element set remains tellme's own** (the round-054 convention): after this ADR the terminal-gated accents are the four round-054 greens **plus** the whole `[Tool Output]` **header** line, **both** `ToolOutputSeparator` lines (grey), and the whole `[Tool Action]` line (yellow). The colour gate is unchanged — the diagnostic stream is a terminal **and** `-r` is off — and colour never enters `stdout` or `turns.log`.

**D2 — The grey `[Tool Output]` accent is threaded through the progress factory.** `render.ProgressFactory` (and `ui.NewTurnProgress` → `NewToolOutputCoordinator` → `ToolOutputWriter.Colour`) gains a `colour bool`. `Begin` writes `formatToolOutputHeaderColour` + `ToolOutputSeparatorColour`; `EndWith` writes `ToolOutputSeparatorColour` after the unconditional reset. The **streamed content lines stay plain** — the round-038 sanitizer owns them, and the operator named only the header and the two separators (assumption A3). The colour-off path returns the pinned literals verbatim, so redirected/`-r` output is byte-identical.

**D3 — The yellow `[Tool Action]` accent lives in the existing renderer.** `formatToolActionColour` wraps the already-sanitized `FormatToolAction` line; `ToolLineRenderer.ActionLine` applies it with the adapter's colour flag (the round-054 `ReasonLine` pattern). The colour wraps the sanitized line, so no control byte from an argument can escape the wrapper.

**D4 — `argValueCap` 189 → 500.** One constant. The mechanic is unchanged (fold → sanitize → rune-boundary cap; one U+2026 counted **inside** the cap; keys and the joined list remain uncapped). This extends the round-034 recorded divergence: the reference caps at 189 **bytes**, tellme now caps at 500 **runes** — the **magnitude** diverges as well as the unit.

**D5 — The estimated payload line becomes a signed increment and drops the budget.** A new `render.Lines` method `PayloadEstimate(t, tokens, delta, mode, model)` renders `[HH:MM:SS] Payload: +<delta> ~<tokens> tokens - <mode> - <model>`; the round-054 green MODE accent applies, the estimated token number stays plain. `PayloadStatus` (the measured line) is **unchanged**. The delta is computed in the CLI call renderer from an **in-memory, session-scoped** previous estimate (a `prevEstimate`/`haveEstimate` pair on `callRenderer`): the **first** estimate of a process renders `+0`; a smaller payload renders the true negative value. This means **no persistence** and **no data-model change** — a fresh process has no predecessor, exactly as clarified.

**D6 — `turns.log` stays content-equal but plain.** No file-leg change: the per-call `emit` already renders each line twice — the coloured diagnostic leg and the plain file leg from the *same* closure — so the file inherits the new content (the `+<delta> ~<tokens>` estimated line; and, when the loop's lines were in the file, the 500-rune cap) with no escapes. The `[Tool …]`/`[Tool Output]` lines remain stderr-only (round-053 F-53-3).

**D7 — What is unchanged.** The measured payload line's shape; the round-024 effective budget (still shown, on the measured line); the round-026 accounting; the round-038 sanitizer and the neutral close; the round-040 coordinator's locking/idle semantics; the round-056 reason gate; exit codes; `go.mod`/`go.sum`.

## Consequences

- **Positive**: the tool chrome is faster to scan (the reference's colour language, scoped to the operator's named elements) and the pre-flight line now reports **motion** (how much this request grew) instead of repeating a constant; a long argument value keeps most of its text.
- **Recorded divergences**: the colour **element set** is tellme's own (the reference does not grey its separators or yellow its action line); the cap **magnitude** now diverges (500 runes vs 189 bytes); the payload-line delta has no reference analogue.
- **Boundary forms** (operator-vetoable assumptions, A7/A8): no predecessor ⇒ `+0`; a shrunken payload ⇒ the true signed value. The round's clarify budget (1–3 questions) was spent, so these were exposed rather than asked.
- **Forward** (RF-057-x): (1) the delta is plain — colouring it is a small follow-on; (2) a whole-line cap for `[Tool Action]` (only values are capped); (3) a non-terminal colour mode (`--color=always`) stays excluded (round-054 RF-54-2); (4) the `[Tool Output]` **content** lines stay plain (A3) — the reference tints its whole block.

## Alternatives considered

- **Colouring the streamed content lines**: rejected — the sanitizer/neutral-close policy owns the content, and the operator named only the header + separators.
- **Persisting the previous estimate** (a round-027-`Calls`-style `Entry` field): rejected — a display number does not justify a data-model change; in-memory is sufficient and matches "the same run".
- **A `--color=always` mode**: out of scope (round-054 RF-54-2).
- **Reusing `PayloadStatus` with a delta parameter**: rejected — the measured line must stay byte-identical; a distinct `PayloadEstimate` method keeps one shape per line.
