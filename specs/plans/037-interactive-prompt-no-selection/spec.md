# Feature Specification: interactive prompt no-selection — align the `-i` suggestion cursor to the reference (round 037)

**Feature Branch**: `037-interactive-prompt-no-selection`

**Created**: 2026-09-17

**Status**: Draft — behavior/parity change. Operator-locked decisions Q1–Q4 (see below).

**Input**: Operator request: *"align to the reference's `-1` no-selection."* In `tellme -i` the suggestion list is rendered with the **first item already highlighted** (the `> ` cursor on row 1) while the editor is still empty — so an empty box looks like it has a "chosen" hint. The reference (`tell-me-go`) starts with **no selection** (`suggester{Index: -1}`) and resets to no selection on every suggestions refresh (`Update(msg, -1)`). This round aligns `tellme` to the reference's no-selection behaviour.

## Operator-locked decisions (clarify — resolved one at a time)

- **Q1 → Strict scope (cursor only).** This round aligns **only** the suggestion-cursor no-selection behaviour and its truth/tests. The **separate** verified divergence found while investigating — with an empty box the reference `Ctrl+S` is a **no-op** (the operator stays in the prompt) while `tellme` **quits** and exits `0` with no turn — is **out of scope**, recorded as a forward item and filed as a **live issue** so it outlives the package freeze (round-035 G3 lesson: durable surface = a live issue, not a frozen plan package).
- **Q2 → Exact reference arithmetic.** The selection starts at **no choice** (`Index == -1`) and **every suggestions refresh resets to no choice** (mirroring `tellme`'s reference `suggestionsMsg → Update(msg, -1)`). The **first `Tab`** selects the **first** suggestion (`Next`: `-1 → 0`); `Shift+Tab` from no-choice follows the reference's own `(index-1+n)%n` arithmetic (`-1 → n-2` for `n ≥ 2`). No "sanity" deviation is introduced: the round's whole point is parity, and a deviation would be a *new* deliberate divergence needing its own justification.
- **Q3 → Keep the "current choice" vocabulary; restate the rule.** "Current choice" stays the concept name; the at-rest assertion is **inverted** to *no* suggestion is the current choice (exactly **zero** `>` cursor rows). The after-`Tab` selection is carried by the **existing** accept journey (`prompting-with-suggestions.feature` — *the operator … accepts the current suggestion* → the accepted text lands in the editor, which only holds if the first `Tab` selected row 0); **no second acceptance rule is minted** for it.
- **Q4 → Witness set = unit pins + an E2E flip.** Unit pins (`internal/ui/tui/prompt/model_chrome_test.go`, `refresh_test.go`) are updated to expect **no** default selection and a **reset on refresh**; the E2E `Then` (`step_t008_chat_then_marks_current_choice.go`) flips to expect **0** cursor rows at rest. Unlike round 036's structurally-blind E2E class, the cursor row is **directly observable** in the captured frames (`tuiCursorRows` counts `> ` rows), so the E2E carrier is valid here alongside the unit pins.

**Scope note**: a **presentation-parity** change (`truth-current`), **not** a spec gap and **not** a new capability. It changes the *initial selection state* of an existing surface. It changes **no** CLI flag, exit code, frozen class-phrase vocabulary, suggestion **ordering/sources**, editor, placeholder, debounce, the `> ` glyph, tool surface, transport, persisted record, or `stdout` byte.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - No suggestion is pre-chosen (Priority: P1)

As an operator opening the interactive prompt, I want the suggestion list to show **no** highlighted entry until I navigate — so an empty box never looks like it already has a "chosen" hint.

**Why this priority**: it is the operator's reported concern and the round's whole substance; it restores reference parity for the at-rest surface.

**Independent verification**: open the prompt through the forced-terminal seam and assert the rendered output carries **zero** `>` cursor rows; unit-pin the `suggester` so a fresh list selects nothing.

**Acceptance Scenarios**:

1. **Given** the prompt has a populated suggestion list, **When** it first renders (no `Tab`/`Shift+Tab` yet), **Then** **no** suggestion row carries the `>` cursor.
2. **Given** an empty suggestion list, **When** it renders, **Then** it shows no cursor rows (unchanged from today).

**Functional Requirements**:

- **FR-001**: With a populated suggestion list, the at-rest rendered prompt MUST mark **no** suggestion as the current choice (exactly **zero** `>` cursor rows).

- **FR-002**: The selection state MUST start at **no choice** and MUST reset to **no choice** on every suggestions refresh (mirroring the reference's `Update(msg, -1)`), so typing that re-fetches suggestions clears any prior highlight.

### User Story 2 - The first navigation chooses the first suggestion (Priority: P2)

As an operator, when I press `Tab` once on the no-selection list, I want the **first** suggestion chosen and — if I then accept it — inserted into the editor.

**Why this priority**: it is the required companion to no-selection; without it, "no selection" would be indistinguishable from a broken cursor. It is carried by the existing accept journey rather than a new rule.

**Independent verification**: the **unit pin** `TestTabFromNoChoiceSelectsFirst` (a **two-item** fixture — the only carrier that can distinguish `-1 → 0` from another landing index) is the authority that the first `Tab` selects the **first** suggestion; the table-driven `TestCycleArithmetic` freezes the wrap/`Shift+Tab` arithmetic. The existing E2E accept journey (`… accepts the current suggestion` → *the interactive prompt holds the accepted suggestion "{text}"*) carries **insertion** of the selected text (its hermetic one-candidate log cannot distinguish which index was selected).

**Acceptance Scenarios**:

1. **Given** the prompt shows no current choice, **When** the operator presses `Tab`, **Then** the **first** suggestion becomes the current choice.
2. **Given** the operator accepted a suggestion, **When** the run completes, **Then** the accepted suggestion is present in the editor (today's accept journey — it proves *insertion* of the selected text; *which* index is selected is proven by the unit pin).

**Functional Requirements**:

- **FR-003**: The first `Tab` on the no-choice list MUST select the **first** suggestion; further `Tab`/`Shift+Tab` MUST wrap through the list; `Shift+Tab` from no-choice MUST follow the reference's arithmetic (no special-casing).
- **FR-004**: The existing accept journey MUST be preserved — `type query → Tab → abort` still inserts the selected suggestion into the editor, and the suggest/accept/abort keybindings and the debounce MUST be unchanged.

---

### Edge cases

- **A single-item list** → still no pre-selection; one `Tab` selects that item.
- **A refresh while a highlight exists** → the refresh resets to no choice (mirroring the reference), not to the previously cycled item.
- **An empty list** → no cursor rows (unchanged); `Tab`/`Shift+Tab` are no-ops.
- **Submit (`Ctrl+S`/`Alt+Enter`) and abort (`Esc`/`Ctrl+C`)** → the round-023 teardown clears the frame; unchanged.
- **The `-i` surface disabled / non-terminal** → no prompt; unchanged (round 015/016).

### Key entities

- **Current choice (the `> ` cursor row)** — the single suggestion row the operator has selected; the round's subject. Its at-rest state becomes "none".
- **Selection state (`Index`)** — the cursor position, now carrying the same `-1` no-choice sentinel as the reference.
- **Suggestion list** — the ordered, styled rows (`Suggestions:` header) — ordering, styling and the `> ` glyph are **unchanged**.

## Requirements *(mandatory)*

### Global requirements

- **FR-005**: The round MUST NOT change the suggestion list **ordering/sources**, the `Suggestions:` header, the selected/unselected styling, the `> ` glyph, the editor, the placeholder, the debounce, the no-metrics-header behaviour, tellme's CLI flags, exit codes, frozen class-phrase vocabulary, tool surface, provider transport, persisted `history.jsonl` shape, or `stdout` bytes; the **only** affected surface is the at-rest **cursor state** of the `-i` prompt (a `stderr`-bound TUI frame).
- **FR-006**: The offline paths (`--tool-usage`, `--version`, `-d`) MUST remain unchanged.

### Out of scope (recorded)

- The **empty-`Ctrl+S` divergence** (reference: no-op, stays in the prompt; tellme: quits with exit `0`) → **live issue [#76](https://github.com/gosharplite/tellme/issues/76)** (forward item); **not** in this round (Q1).
- Any change to the accept **heuristic** (whole-line vs last-token replacement, round-016 FR-007) or to the suggestion **sources/engines**.
- The round-036 tool-reason and the round-035 spinner-yield surfaces.

## Success criteria *(mandatory)*

- **SC-001**: At rest, a populated prompt renders **zero** `>` cursor rows (US1).
- **SC-002**: The first `Tab` selects the **first** suggestion and the existing accept journey inserts it into the editor (US2).
- **SC-003**: A suggestions refresh resets the highlight to none; the `> ` glyph, list ordering, styling, editor, placeholder, debounce and the no-metrics-header behaviour are **byte-identical** apart from the at-rest cursor (FR-005).
- **SC-004**: `make verify`, the E2E suite and the Gherkin/DSL topology audit (re-run after the truth MODIFY) are green.
- **SC-005**: The change has a **falsifiability witness**: re-defaulting the cursor to `0` fails the at-rest pin (and the E2E `Then`); removing the reset-on-refresh fails the refresh pin; each reproduced then reverted.

## Assumptions

- **Truth ownership.** `presenting-the-interactive-prompt.feature` and `chat/dsl.md` are `TruthArtifact`s owned by `/axb-dsl-refine` (`truth-single-owner`); `techstack.md` is owned by `/axb-technical-research`; `/axb-api-plan` is **NOOP** (no HTTP surface); `/axb-data-plan` is a checked **NOOP** (no persisted state). All truth edits are recorded in this round's `truth-delta.md`.
- **Frozen history.** `specs/plans/016-interactive-prompt-visual-parity/**` (including its `ui/screens/entry.txt`, which drew `>` on row 1) is **never edited** — 037 is a fresh package with its own `ui/`.
- **`/axb-ui-plan` (terminal mode).** The TUI end gets a terminal-mode `ui/` (textual screens) re-rendering the **at-rest** frame with **no** `>` cursor, supplied to `/axb-system-analysis` for review (not redone there). Plan-side only; never truth.
- **`/axb-spec-by-example` is NOT a NOOP.** The at-rest highlight is a **user-visible acceptance** change, so 037 writes a new plan-side acceptance rule (*the prompt pre-selects no suggestion*), which `/axb-dsl-refine` maps 1:1 onto the flipped interface `Then` (`acceptance-coverage`). This is the correct contrast to round 036, whose violated guarantee was a code behaviour already carried at the contract level.
- **Supersession.** The frozen round-016 acceptance (*exactly one suggestion is marked as the current choice at rest*) is **superseded** for the current system by 037's at-rest rule; the round-016 package stays frozen history and is never rewritten.
- **Forward item (durable surface).** The empty-`Ctrl+S` divergence is filed as a **new live issue [#76](https://github.com/gosharplite/tellme/issues/76)** this round (its statement, not a comment), so it survives the package freeze (round-035 G3 lesson).
- **Cursor-sentinel semantics.** The no-choice state is represented by a negative cursor index (the reference's `-1`), and `selected()` already returns the empty string for it — so the change is confined to the **default** and the **reset-on-refresh**, plus the pins that assert them.
