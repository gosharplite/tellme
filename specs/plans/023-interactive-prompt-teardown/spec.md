# Feature Specification: tellme Interactive Prompt Teardown & Submit-Surface Parity (round 023)

**Feature Branch**: `023-interactive-prompt-teardown`

**Created**: 2026-09-15

**Status**: Draft — operator-locked decisions Q1–Q3

**Input**: Operator request: "I want `-i` look and feel like [the Ctrl+D surface] … The TUI should disappear after CTRL+S, show prompt again just like others. Current `-i` has no spinner, this is not correct. I believe tell-me-go `-i` TUI disappear after CTRL+S. Please check." — After the operator confirmed the reference behaviour, the request is: once the operator submits the `-i` interactive prompt, the editor must **tear down** (the bordered box disappears) and the run must continue on the **standard turn surface** (echoed prompt · input-capture acknowledgement · turn chrome · live spinner · post-turn status), exactly like the positional and Ctrl+D prompt surfaces.

**Operator-locked decisions (clarify)** — the operator resolved all three scope decisions up front:

- **Q1 → 1a — clear the frame on submit (reference parity).** On submit (`Ctrl+S` / `Alt+Enter`) and on abort (`Esc` / `Ctrl+C`) the interactive editor frame MUST be cleared, exactly as `tell-me-go`'s `-i` does (`View()` returns empty once the model is submitted/aborted). It MUST NOT leave the last frame on screen. This reverses the round-016 "the frame is always rendered" choice (which existed only so the E2E harness could capture the final frame).
- **Q2 → 2a — resume the full standard surface.** The `-i` submit MUST continue on the **same** surface as the positional / Ctrl+D prompt: the input-capture acknowledgement, the 80-column `─` rule, the `╭─⠿ Turn <N> - <mode>` header, the **live progress spinner**, and the post-turn status lines. This reverses the round-017 rule "the interactive prompt shows no turn chrome" and the round-019 "`-i` is out of scope for the spinner" exclusion.
- **Q3 → A′ — echo the submitted prompt, keep tellme's single captured line.** Because the editor box is erased on submit (Q1), the operator can no longer see what they sent; the submitted prompt MUST therefore be **echoed on its own diagnostic line before** the captured line, so the `-i` submit reads:

  ```text
  <submitted prompt>                        ← echoed (new)
  [HH:MM:SS] Input captured. Processing...   ← tellme's existing single line
  ```

  tellme's single-line captured form is kept (the reference's two-line `Input captured:` / `<prompt>` / `Processing...` layout is **not** adopted). The echo applies **only** to the `-i` submit surface: the positional and Ctrl+D surfaces already show the typed text (shell command line / terminal line echo) and MUST stay unchanged.

> **Scope note**: this is a **presentation-surface** round for the CLI end's `-i` interactive prompt, not a capability round. It changes how the `-i` surface renders when the operator submits or aborts — it does **not** change the suggestion engine, the shared prompt log, the opt-in gating, the provider transport, the turn result, the tool surface, any CLI flag, any exit code, or any `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The interactive editor releases the terminal on submit (Priority: P1)

As an operator using `tellme -i`, I want the bordered editor to **disappear** the moment I press `Ctrl+S`, so the rest of the turn reads exactly like every other prompt surface instead of the box lingering on screen.

**Why this priority**: it is the operator's primary request. Today the editor frame is always rendered (round-016, a test-harness convenience), so the box persists after submission and the turn output is drawn below a stale frame — visibly unlike the positional / Ctrl+D surfaces.

**Independent verification**: drive the `-i` surface with a scripted key sequence (open → type → submit) over the forced-terminal seam; assert the editor frame is **gone** from the rendered output after submission (no bordered box / no editor row survives the submit) and that a clean abort (`Esc`) likewise clears the frame.

**Acceptance Scenarios**:

1. **Given** the operator is at a terminal with the interactive prompt open and has typed `hi`, **When** the operator submits (`Ctrl+S`), **Then** the editor frame is cleared — the bordered editor does not remain on screen.
2. **Given** the operator is at a terminal with the interactive prompt open, **When** the operator aborts (`Esc` / `Ctrl+C`), **Then** the editor frame is cleared, no request is sent, and the process exits success.
3. **Given** the operator submits, **When** the frame is cleared, **Then** the teardown leaves no residue — the following output begins cleanly at the start of a line (no orphaned box border or editor prompt).

**Functional Requirements**:

- **FR-001**: On submit (`Ctrl+S` or `Alt+Enter`) the interactive editor frame MUST be cleared before the turn surface renders; the bordered editor MUST NOT remain on screen.
- **FR-002**: On abort (`Esc` or `Ctrl+C`) the interactive editor frame MUST be cleared; no request is sent and the process exits success (unchanged abort contract).
- **FR-003**: The teardown MUST leave no residue — the cursor returns to column 0 of a clean line so the operator's shell prompt and the subsequent output are not interleaved with a stale frame.
- **FR-004**: A submit of an empty editor MUST NOT submit (unchanged) and MUST NOT be treated as a turn.

---

### User Story 2 - The `-i` submit continues on the standard turn surface (Priority: P1)

As an operator, after I submit `-i` I want the run to continue exactly like the positional / Ctrl+D prompt — my prompt echoed, then the input-capture acknowledgement, the turn frame, the live spinner, and the post-turn status lines — so the interactive prompt is just another way to enter a prompt, not a separate output mode.

**Why this priority**: it completes the "look and feel like the others" request. Today the `-i` submit path passes `chrome: false`, so it shows **no** input-capture line, **no** turn chrome, and **no** spinner (round-017/019 pinned this); only the payload line, the answer, and the post-turn status are emitted. The operator flagged the missing spinner as incorrect.

**Independent verification**: submit a prompt via `-i` and assert the post-submit diagnostic stream carries (in order) the echoed prompt, the input-capture acknowledgement, the `─` rule + `╭─⠿ Turn <N> - <mode>` header, and (with a terminal `stderr` and not `-r`) the live spinner, then the payload/answer/post-turn status — matching the positional / Ctrl+D output for the same prompt. Confirm `stdout` is byte-identical to the non-interactive run.

**Acceptance Scenarios**:

1. **Given** the operator submits the prompt `hi` at the interactive prompt, **When** the turn begins, **Then** the diagnostic stream carries the echoed prompt, then the input-capture acknowledgement.
2. **Given** the same submit, **When** the turn surface renders, **Then** the diagnostic stream carries the 80-column `─` rule and the `╭─⠿ Turn <N> - <mode>` header, the pre-flight payload line, the answer, the post-turn payload line, and the post-turn status lines — the same shape as the positional / Ctrl+D surfaces.
3. **Given** a terminal `stderr` and not `-r/--raw`, **When** the `-i` turn runs, **Then** the live progress spinner is shown while waiting and cleared before the answer (as on the other surfaces).
4. **Given** the same prompt entered positionally and via `-i`, **When** both runs complete, **Then** `stdout` is byte-identical and the diagnostic chrome shape matches **apart from the echoed prompt (FR-007)**.

**Functional Requirements**:

- **FR-005**: On `-i` submit, the turn MUST resume the **same** operator surface as the positional / Ctrl+D prompt — the input-capture acknowledgement, the 80-column `─` rule, and the `╭─⠿ Turn <N> - <mode>` header.
- **FR-006**: The `-i` submit MUST be a **chrome surface**: it MUST render the turn chrome, the live progress spinner (gated on a terminal `stderr` and not `-r/--raw`), and the post-turn status lines, identical to the other prompt surfaces.
- **FR-007**: Because the editor box is erased (FR-001), the submitted prompt MUST be **echoed as its own diagnostic block before** the input-capture acknowledgement, so the operator can see the text they sent.
- **FR-008**: The echo (FR-007) MUST apply **only** to the `-i` submit surface. The positional and Ctrl+D surfaces MUST NOT echo (the shell command line / terminal line echo already shows the text) — unchanged.
- **FR-009**: The echoed text MUST be the submitted (trimmed) prompt echoed **verbatim**; a multi-line submission preserves its embedded newlines (the echo is a diagnostic block, not a single line).
- **FR-010**: The `-i` turn MUST remain exactly one reasoning turn per submit — same provider call, same session-history append — as on the other surfaces.

---

### Edge cases

- **Submit an empty editor** → not a submission; no turn, nothing echoed, no chrome (unchanged).
- **Abort (`Esc` / `Ctrl+C`)** → frame cleared, no echo, no request, exit success (FR-002).
- **`-r/--raw` on `-i`** → the spinner is suppressed (raw), but the echo, the input-capture line, and the turn chrome still render — matching the other surfaces under `-r`.
- **Non-terminal input** → the interactive prompt never engages (opt-in gating unchanged); the piped / positional path is unaffected.
- **Multi-line submission** → echoed as submitted (FR-009); the turn runs on the whole prompt.
- **Streams split** (`stdout` vs `stderr` redirected separately) → the echo, the input-capture line, the chrome, the spinner, and the tool-loop lines are on `stderr`; `stdout` is byte-exact (FR-011).

### Key entities

- **Interactive prompt (`-i`)** — tellme's opt-in Bubble Tea editor (round 015/016); the surface this round changes on submit/abort.
- **Submit transition** — the moment `Ctrl+S` / `Alt+Enter` ends the editor and the reasoning turn begins (and the analogous abort transition for `Esc` / `Ctrl+C`).
- **Turn surface** — the shared `stderr` rendering of a turn (echoed prompt · input-capture acknowledgement · `─` rule + `╭─⠿ Turn <N>` header · live spinner · tool-loop lines · post-turn status); the `-i` submit now uses it.
- **Echoed prompt** — the submitted prompt re-printed as its own diagnostic block on the `-i` surface (verbatim; embedded newlines preserved), because the editor box that held it is erased.
- **Diagnostic stream (`stderr`)** — the stream that carries all of the above; the answer stays on `stdout`.

### Global requirements

- **FR-011**: `stdout` MUST stay byte-exact; the echoed prompt, the input-capture acknowledgement, the turn chrome, the spinner, and the tool-loop lines MUST be written to the diagnostic stream (`stderr`).
- **FR-012**: The `-i` opt-in gating (`-i` / `USE_TUI_PROMPT` AND a terminal stdin), the suggestion engine, the shared prompt log (round-015), and the abort/empty contract MUST be unchanged.
- **FR-013**: The frozen class-phrase vocabulary MUST stay unchanged (11) and no new failure class MUST be introduced; the exit-code contract is unchanged.
- **FR-014**: The turn result contract MUST be unchanged — one reasoning turn per submit, the same provider request, the same session-history/usage persistence (rounds 018/019).

### Success criteria

- **SC-001**: After an `-i` submit, the rendered output no longer contains the editor frame (the bordered box / editor row does not survive the submit); after an `-i` abort, likewise.
- **SC-002**: An `-i` submit's post-submit `stderr` carries, in order, the echoed prompt line, the input-capture acknowledgement, the `─` rule + `╭─⠿ Turn <N> - <mode>` header, and (terminal `stderr`, not `-r`) the live spinner — the same shape as a positional / Ctrl+D run of the same prompt.
- **SC-003**: `stdout` is byte-identical between the `-i` submit and the positional run of the same prompt (and to the pre-change run).
- **SC-004**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, and falsifiability witnesses fail the matching scenario — (a) reverting the teardown leaves frame residue; (b) disabling the `-i` chrome drops the header/spinner.
- **SC-005**: The positional and Ctrl+D surfaces are unchanged (no echo introduced); the interactive prompt's suggestions, opt-in gating, and shared prompt log are unchanged.

### Assumptions

- The teardown is achieved by clearing the editor frame on the submit/abort transition (the reference parity route, `View()` → empty once submitted/aborted); the exact mechanism is a `/axb-technical-research` detail. The round-016 "always render" existed solely so the E2E harness could capture the final frame — the harness witness is reworked in `/axb-dsl-refine` + `/axb-tasks` (assert the frame is **absent** post-submit and the standard chrome **present**).
- The `-i` submit path passes the same chrome flag the positional / Ctrl+D surfaces use; the spinner gate is unchanged (`chrome` AND a terminal `stderr` AND not `-r`) — the `-i` path simply joins that set. The echoed prompt is emitted on `stderr` as part of the `-i` submit chrome (before the input-capture line).
- The reference's two-line `Input captured:` / `<prompt>` / `Processing...` layout is **not** adopted; tellme keeps its single-line captured form and adds the echo (Q3 A′).
- No new dependency; POSIX-only (unchanged); the hermetic forced-terminal seam + scripted keys remain the E2E driver; no pty.
- This rewrites in place the round-015/016/017/019 truth rules for the `-i` surface ("always render", "no turn chrome", "`-i` excluded from the spinner"); the prior plan packages stay frozen (`fresh-package-per-round`).
- The positional / Ctrl+D surfaces, the opt-in gating, the suggestion engine, the shared prompt log, the payload line, the tool-loop line, and the post-turn status lines are **out of scope** and unchanged.
