# Phase 0 Research: tellme turn surface — operator chrome parity with tell-me-go (Round 017)

Topic: make `tellme`'s **non-interactive prompt surfaces** open a turn the way `tell-me-go` does — add the reference's **input-capture acknowledgement**, its **horizontal rule + `╭─⠿ Turn <N> - <mode>` header** wrapping the existing pre-flight payload line, and its **blank-line spacing** — so a turn reads the same as the reference.

Scope note: the language (Go 1.26), module, the CLI/config layers, the `llm.Gateway` port + adapters, the session-history store, the agent tool loop, the interactive TUI prompt, and the rest of the pipeline were locked in rounds 001–016. The system still has **one CLI end**; the round reaches **no new endpoint** and persists **no new state**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; `godog` running the built binary; E2E acceptance + pure-helper units) and are **not re-decided**. **IN**: the input-capture line, the turn frame (rule + header + payload line + spacing) on the **positional/piped prompt turn** and the **round-012 interactive plain reader**. **OUT**: the `-i` TUI surface (round 016), every non-prompt path (boot / `-d` / `-l` / `--version` / prompt-less `--new`), the reference's second startup line (`[Info] Starting chat...`), and the **post-turn** lines (the measured payload line, the metrics line, the final summary) — explicitly out of scope this round.

Reference: `tell-me-go` `internal/ui/capture.go` (`finalizePrompt`: `[HH:MM:SS] Input captured. Processing...`) and `internal/ui/renderer_metrics.go` (`renderTurnHeader`: a leading `\n`, an **80-column** `─` rule, `╭─⠿ Turn <N> - <mode>`, the payload line, a trailing blank).

---

## Decision 1: Emit the reference turn chrome from a hand-written stderr formatter (no new dependency)

- **Decision**: On a prompt-bearing turn of surfaces (A)/(B), emit a fixed structural block to the **diagnostic stream (`stderr`)**: the acknowledgement line `[HH:MM:SS] Input captured. Processing...`, a leading blank line, an **80-column** horizontal rule (the reference's `─` × 80 literal), the header `╭─⠿ Turn <N> - <mode>`, the **existing** pre-flight payload line (round-009 format, unchanged), and a trailing blank line. The chrome is **plain text** (UTF-8 glyphs; no ANSI this round — see Decision 3), produced by a hand-written Go formatter living beside the existing status formatter (`internal/ui`), reusing the injected clock seam for both timestamps.
- **Rationale**: strict parity here is a *presentation* change needing no capability and no module; the structural tokens (rule, glyph, header text, spacing) are exactly what the operator sees. Keeping the formatter dependency-free and next to `ui.FormatPayloadStatus` makes the whole block unit-testable without a terminal.
- **Alternatives considered**:
  - **Style the chrome with `lipgloss`** (already present via the TUI family): the chrome is a plain glyph block on a non-TUI stream, so a styling library adds nothing — rejected.
  - **Route the chrome through the glamour renderer**: glamour renders the *answer*, not operator chrome — rejected.

## Decision 2: The turn number is the session's completed-turn count + 1

- **Decision**: `<N>` is the count of the session's **completed turns**, plus one (the reference's `SessionTurns + 1`). tellme persists **one `history_entry` line per completed turn** in the active `history.jsonl`, so the count is the number of such lines — **not** a message count and **not** the reference's `entries/2` (whose store keeps 2 messages per turn). It is read from the already-loaded prior history (`len(prior) + 1`); `--new` archives **before** the turn, so a fresh session shows `Turn 1`, and a resumed session continues the count. (Today a session archives only on `--new`, which resets the count to 1; a future summarisation/archive path must preserve this.) `<mode>` is the effective mode (the resolved `MODE`).
- **Rationale**: the reference numbers turns **per session**; a resumed session must continue (not restart at 1), and the history already carries the count — no new counter.
- **Alternatives considered**:
  - **Always `Turn 1`** (per-invocation): diverges from the reference on any resumed session — rejected.
  - **A new persisted turn counter**: duplicates what the history already provides — rejected.

## Decision 3: Plain structural chrome (no ANSI colour this round); the payload line stays as round 009 pinned

- **Decision**: reproduce the **structure** (rule, glyph, header text, spacing) in plain text; introduce **no** ANSI colour/styling this round. The reference's gray styling is recorded as a **forward item** (a named non-goal). The pre-flight payload line keeps its round-009 pinned **plain** format and is emitted **inside** the frame, unchanged.
- **Rationale**: the operator's three named items — input-capture line, turn framing, spacing — are **structural**; `tellme`'s diagnostic status lines are plain by convention (round 009); a plain chrome avoids wiring a new terminal probe solely to gate colour and keeps the round-009 payload-line contract byte-for-byte intact. Colour parity is a cheap follow-up if wanted.
- **Alternatives considered**:
  - **Reproduce the reference's gray styling, gated on a terminal**: would require a new colour probe on the diagnostic stream and would leave the (pinned, plain) payload line as the one uncoloured element of the block — a mixed-styling artefact; deferred as a forward item.
  - **Colour unconditionally**: pollutes a piped `stderr` with ANSI escapes — rejected.

## Decision 4: Blank-line spacing mirrors the reference

- **Decision**: emit a **leading blank line** before the rule and a **trailing blank line** after the payload line, so the frame reads as the reference block and the answer that follows is separated from it. The answer continues on `stdout` immediately after the gap.
- **Rationale**: spacing is one of the three named items; the reference writes `"\n────…\n"`, then the header and payload line, then a blank line (`fmt.Fprintln`).
- **Alternatives considered**:
  - **No blank lines** (compact frame): diverges from the reference's spacing — rejected.
  - **Extra blank lines**: over-spaced versus the reference — rejected.

## Decision 5: One seam — the chrome is emitted for surfaces (A) and (B) only

- **Decision**: the acknowledgement and the frame are emitted from **one** turn-path seam, enabled for the **positional/piped prompt turn** (A) and the **round-012 plain reader** (B), and disabled for the **`-i` TUI submit path** (C) and every non-prompt path. Concretely, the turn runner carries a *chrome* switch (or an equivalent boolean) that the CLI sets true for (A)/(B) and false for (C); the runner writes the acknowledgement + frame when it is set, before the provider request. No non-prompt path (boot, `-d`, `-l`, `--version`, prompt-less `--new`) enters the seam.
- **Rationale**: (C) currently funnels through the **same** turn runner as (A)/(B); without an explicit switch the chrome would leak onto the round-016 surface. One switch keeps the decision in one place and makes the boundary unit-testable (`FR-007`).
- **Alternatives considered**:
  - **Emit at each call site**: duplicates the sequence across (A) and (B) and risks drift — rejected.
  - **Emit inside the runner unconditionally**: leaks the frame onto the `-i` surface — rejected (violates `FR-007`).

## Decision 6: Hermetic (no-pty) verification via the injected clock seam and stream order/presence/absence assertions

- **Decision**: verify with the **injected clock seam** + **injected streams**: assert the stderr sequence (acknowledgement → blank → rule → `Turn <N> - <mode>` → payload line → blank), the turn-number rule (fresh vs resumed), that `stdout` carries **only** the answer (byte-exact), and that the `-i` and non-prompt paths carry **no** acknowledgement and **no** frame. The E2E layer drives the built binary through the existing `TELL_ME_FORCE_STDIN_TTY` seam + a scripted stdin for the reader case. No real pty.
- **Rationale**: determinism (no `time.Sleep`, no pty) and `TERM`-independence; asserting structural presence/order/absence (not ANSI bytes) matches the round-006/010/016 precedent.
- **Alternatives considered**:
  - **Exact stderr byte comparison incl. timestamps**: non-deterministic — rejected (match the timestamp by pattern, round-009 precedent).
  - **A pty harness**: heavy and previously foreclosed — rejected (a named pin).

## Decision 7: No new dependency; POSIX-only; the three AIxBDD must-ask questions remain settled

- **Decision**: the round adds **no** module (stdlib + the existing `internal/ui`), is **POSIX-only** (round-012/015/016 precedent), and does not re-open the must-asks — single CLI end; `godog` running the built binary; E2E acceptance + fast pure-helper units — per the standing `techstack.md`. No `/axb-clarify` is owed for them.
- **Rationale**: the round changes only a non-TUI turn's presentation; it introduces no new interface, runner, or end.
- **Alternatives considered**:
  - **Re-open the techstack/test questions**: no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **Post-turn lines (`A7`)**: the reference's measured payload line, its `M:`/`H:`/`C:` metrics line, and the final `╰─⠿ Ready` summary are **out of scope** this round — a forward item.
- **Colour/styling parity (Decision 3)**: the reference's gray rule/glyph styling is a recorded forward item (a named non-goal this round).
- **Fixed rule width (Decision 1)**: the 80-column rule is a fixed literal (reference parity); terminal-width reflow is a forward item if ever wanted.
- **Second startup line (`A3`)**: the reference's `[Info] Starting chat...` line is out of scope this round.
- **Anchor issue**: [#42](https://github.com/gosharplite/tellme/issues/42) — the round-017 anchor, opened retroactively (PR #41 review nit).
