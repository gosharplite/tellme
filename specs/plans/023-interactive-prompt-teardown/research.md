# Phase 0 Research: tellme Interactive Prompt Teardown & Submit-Surface Parity (round 023)

**Topic**: how tellme's `-i` interactive prompt hands the terminal back when the operator submits or abandons the prompt, and how its submitted turn resumes the standard operator surface (echoed prompt · input-capture acknowledgement · turn chrome · live spinner · post-turn status). Each decision supports `spec.md` (US1/US2 · FR-001..FR-014) and the round's operator-locked decisions Q1 (1a) · Q2 (2a) · Q3 (A′).

## Decision 1: Clear the editor frame on the submit/abort transition (reference parity)

- **Decision**: the interactive prompt's model **clears its rendered frame** the moment the operator submits (`Ctrl+S` / `Alt+Enter`) or aborts (`Esc` / `Ctrl+C`) — the model reports submitted/aborted and `View()` returns empty, matching `tell-me-go`'s `promptModel.View()` (`if m.submitted || m.aborted { return "" }`). This replaces round-016's "the frame is always rendered" comment/behaviour.
- **Rationale**: it is the operator's primary request and the reference's behaviour; a frame that persists after submission draws the turn's output beneath a stale bordered box, visibly unlike the positional / Ctrl+D surfaces.
- **Alternatives considered**:
  - Keep rendering the frame and print a separator line — rejected: the box still lingers, which is exactly the reported defect.
  - Scroll the frame off by padding the view — rejected: terminal-height dependent and non-deterministic; the reference simply clears.

## Decision 2: The `-i` submit joins the existing chrome surface (no bespoke chrome)

- **Decision**: the `-i` submit path passes the **same chrome flag** the positional / Ctrl+D prompt surfaces already use (`turnOptions.chrome = true`) so the input-capture acknowledgement, the 80-column `─` rule, the `╭─⠿ Turn <N> - <mode>` header, and the post-turn status render through the **shared** `internal/ui` emitters — no `-i`-specific chrome code path.
- **Rationale**: the chrome and post-turn status emitters already exist (rounds 017/018); reusing them keeps a single rendering path and a single set of truth rows to maintain.
- **Alternatives considered**:
  - Re-implement the chrome inline on the `-i` path — rejected: duplicates the formatters and forks the truth into two shapes that can drift.

## Decision 3: The spinner gate is unchanged; `-i` becomes a positive surface

- **Decision**: the round-019 spinner gate (`a chrome surface` AND `-r` off AND `isatty(stderr)`) is **unchanged**; joining the chrome surface (Decision 2) makes the `-i` submit a **positive** spinner surface. This supersedes the round-019 "`-i` excluded" and removes `-i` from the spinner's negative set; the round-023 truth (DSL) drops `-i` from the spinner negatives and adds it to the positives.
- **Rationale**: the spinner is already gated on the chrome flag plus a terminal diagnostic stream; a separate `-i` gate would be redundant and would contradict Decision 2. The operator explicitly flagged the missing spinner as incorrect.
- **Alternatives considered**:
  - A `-i`-specific spinner gate — rejected: redundant, and it would reintroduce a second gate to keep in sync.

## Decision 4: Echo the submitted prompt on the `-i` surface only (Q3 A′)

- **Decision**: because the editor box is cleared (Decision 1), tellme prints the **submitted (trimmed) prompt on its own diagnostic (`stderr`) line immediately before** the input-capture acknowledgement, **on the `-i` submit surface only**. tellme **keeps its existing single-line** `[HH:MM:SS] Input captured. Processing...` form (the reference's two-line `Input captured:` / `<prompt>` / `Processing...` layout is **not** adopted).
- **Rationale**: the operator must still see what they sent — the box that held the text is gone. The positional and Ctrl+D surfaces already show the typed text (shell command line / terminal line echo) and must not double-print it.
- **Alternatives considered**:
  - Adopt the reference's two-line `Input captured:` form — rejected: Q3 A′ chose tellme's uniform single captured line.
  - Emit no echo — rejected: the operator would lose sight of the prompt (the round-021 style defect the operator raised).
  - Echo on every prompt surface — rejected: the other surfaces already show the text; double-printing is noise.
- **Reserved-prefix note (PR #51 implementation-review NIT)**: the echoed prompt is written verbatim with **no** `tellme: ` prefix. The frozen reservation (`exactly one stderr line begins with `tellme: ``) binds the **class-phrase** line, which the echo never carries; a submitted prompt that itself begins with `tellme: ` would nonetheless read like a class phrase on that stream — a recorded, accepted residual (very low likelihood).

## Decision 5: E2E witness rework — assert the frame is gone and the chrome is present

- **Decision**: the round-016 witness ("the final rendered frame is captured by the E2E harness") is **replaced**. The `-i` submit witness asserts (a) the editor frame is **absent** after submit/abort — a model-level pin on the cleared `View()` — and (b) the **standard chrome / echoed prompt / spinner is present** on the `-i` submit surface, asserted on the captured streams (E2E). The harness keeps the `TELL_ME_FORCE_STDIN_TTY` seam + scripted keys.
- **Rationale**: once the frame is cleared, the "final rendered frame" premise no longer holds; the assertion must move to "the frame is gone" **and** "the standard surface took over", which is exactly what the round's success criteria measure.
- **Alternatives considered**:
  - Keep always-render and assert the box is gone by a sentinel marker — rejected: impossible, the frame would persist.

## Decision 6: The turn-scoped spinner epoch is unchanged (starts at capture)

- **Decision**: the `-i` submit's spinner epoch is the **same prompt-capture moment** the other surfaces use (the input-capture acknowledgement); no new timing seam is introduced.
- **Rationale**: joining the chrome surface (Decision 2) reuses the existing turn-scoped epoch (`turnStart`); the round-019 "counts from prompt capture" semantics carry over unchanged.
- **Alternatives considered**:
  - A `-i`-specific epoch — rejected: inconsistent with the other surfaces and unneeded.

## Decision 7: No new dependency; POSIX-only; hermetic seams unchanged

- **Decision**: no new module is added; the `TELL_ME_FORCE_STDIN_TTY` and `TELL_ME_TUI_DEBOUNCE` diagnostic seams and the scripted-key E2E driver stay; **no pty** is used.
- **Rationale**: the teardown is a model-state change; the echo/chrome/spinner reuse existing code. The reference ships no pty either.
- **Alternatives considered**:
  - A pty-capable harness — rejected: forsworn (round-005 grill Q6 / round-006 Q2); the forced-terminal seam plus the cleared-frame pin suffice.

## Decision 8: Scope — presentation-only; contracts unchanged

- **Decision**: the round changes **only** the `-i` surface's submit/abort rendering. The opt-in gating, the suggestion engine, the shared prompt log, the provider request, the one-turn contract, the exit codes, the frozen class-phrase vocabulary (11), and `stdout` are **unchanged**.
- **Rationale**: the operator-locked Q1/Q2 scope and `spec.md` FR-010..FR-014; the round is a presentation-surface round, not a capability round.
- **Alternatives considered**:
  - Widen the round to also touch the positional / Ctrl+D surfaces — rejected: out of scope; those surfaces already behave correctly (and must stay echo-free).
