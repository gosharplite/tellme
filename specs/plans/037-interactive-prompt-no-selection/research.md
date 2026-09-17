# Technical Research: interactive prompt no-selection (round 037)

**Plan package**: `specs/plans/037-interactive-prompt-no-selection`
**Inputs**: `spec.md` (US1/US2, FR-001–006, operator-locked Q1–Q4), the existing `techstack.md` interactive-prompt rows, and the live `internal/ui/tui/prompt/**` code + the reference `tell-me-go/internal/ui/tui/prompt/**`.

## AIxBDD must-ask questions (already settled — no clarify round needed)

Per `rules/AIxBDD必問問題與起始專案介面澄清判準.md`, the three must-asks are answered by the **existing** `specs/truth/techstack.md` and this round does not re-judge them: **BDD techstack** = `godog` over `specs/truth/features/**` + Go `testing`; **Test strategy** = E2E (black-box) + fast unit tests; **System ends** = a single CLI end (the `-i` TUI prompt is a CLI surface), no HTTP/API, no web frontend. No new must-ask gap arises from a selection-state parity change, so `/axb-technical-research` proceeds without a clarify round (the Q1–Q4 decisions were resolved with the operator during `/axb-specify`).

---

## 決策 1：No-selection is the `cursor` sentinel `-1`; `set()` resets to it on every refresh

- **Decision**: In `internal/ui/tui/prompt/suggester.go`, the **no-choice** state is `cursor == -1`, and `set(items)` **always resets `cursor` to `-1`** (the reference's `Update(msg, -1)`), instead of clamping to `0`. `selected()` already returns `""` for a negative cursor, so no other read path changes.
- **Rationale**: The reference (`tell-me-go/internal/ui/tui/prompt/suggester.go`) constructs `suggester{Index: -1}` and its `suggestionsMsg` handler calls `Update(msg, -1)` — no selection at rest **and** no selection after each refresh. Round 016 deviated (cursor default `0`, "first item is the current choice") to satisfy a Gherkin line drawn from the reference's *mockup*; the reference's *code* never pre-selects. Resetting in `set()` (the one mutation point that receives a fresh list) makes both the initial render and every refresh consistent from a single site.
- **Alternatives considered**:
  - Reset only at construction: leaves a refresh re-selecting `0` (a second deviation) — rejected; the reference resets on refresh.
  - Reset to `0` when the list is non-empty (status quo): the defect under fix.

## 決策 2：`cycle` arithmetic is kept exactly; the first `Tab` lands on index `0` from `-1`

- **Decision**: Keep `cycle(delta)` as `(cursor+delta+len)%len` with **no special-casing** of `-1`. With `len ≥ 1`: `cycle(1)` from `-1` → `0` (the **first** suggestion); `cycle(-1)` from `-1` → `len-2` (the reference's own `(index-1+n)%n` quirk on `Prev`).
- **Rationale**: The round is **parity**; a "sanity" deviation (e.g. `Prev` from no-choice → last item) would introduce a **new** divergence requiring its own justification. The modulo arithmetic already yields the reference's exact numbers because `-1 ≡ len-1 (mod len)`. `accept(delta)` calls `cycle(delta)` then inserts `selected()`, so `Tab` (`accept(1)`) from no-choice selects the first item and inserts it — the existing accept journey holds.
- **Edge case `len == 1` (corrected — the earlier "residual risk" claim was arithmetically false; round-037 review F-2)**: `Shift+Tab` from no-choice follows the reference's own `(i-1+n)%n`, landing on `len-2` for `len ≥ 2` and on the **sole item (`0`)** for `len == 1` — Go's `%` truncates toward zero and any integer mod 1 is `0`, so a one-item list **is** selectable by both `Tab` and `Shift+Tab`, with no special-casing. This is frozen by the table-driven `TestCycleArithmetic` pin (`model_chrome_test.go`).
- **Alternatives considered**: normalise `-1` to `0` before `cycle` (removes the quirk but diverges from the reference); leave the arithmetic but add a `Prev`-from-no-choice special case (new divergence).

## 決策 3：Keep the "current choice" vocabulary; invert the at-rest assertion; no new rule for the accept path

- **Decision**: The concept name "current choice" stays. The at-rest `Then` inverts to *the interactive prompt **marks no** suggestion as the current choice* (the rendered **suggestion block** carries **zero** `>` cursor rows). The after-`Tab` selection is **not** given a second acceptance rule — `prompting-with-suggestions.feature`'s existing accept journey (`… accepts the current suggestion` → *the interactive prompt holds the accepted suggestion "{text}"*) continues to carry *insertion* of the selected text.
- **Carrier authority (round-037 review)**: the accept journey proves **insertion**, not **which index** was selected — its hermetic log holds a **single** candidate, so index `0` and any other landing index yield the same inserted text. The authority for *"the first `Tab` selects the **first** item"* is therefore the **unit pin `TestTabFromNoChoiceSelectsFirst`** (a two-item fixture, the only carrier that can distinguish `-1 → 0`), reinforced by the table-driven `TestCycleArithmetic` (`-1 → 0` for every `len ≥ 1`).
- **Rationale**: `acceptance-coverage` requires every rule to be carried by ≥1 `InterfaceFeature` and vice-versa in effect; minting a second at-rest/after-`Tab` rule would add a near-duplicate carrier. The inversion is a **MODIFY** of the existing rule/Example (the DSLRow is edited in place), keeping the CLI interface truth's shape stable.
- **Alternatives considered**: a brand-new `Rule: The prompt selects nothing until the operator navigates` plus keeping the old rule (contradiction — rejected); dropping "current choice" entirely (loses the accept-path vocabulary the accept journey depends on).

## 決策 4：Witness = updated unit pins + an E2E `Then` flip (the E2E carrier is valid here)

- **Decision**: The guarantee is witnessed by (a) **unit pins** in `internal/ui/tui/prompt/model_chrome_test.go` (at-rest view has **0** cursor rows; `selected()` is `""` by default; `TestTabFromNoChoiceSelectsFirst` — the two-item authority that the first `Tab` lands on index `0`; `TestCycleArithmetic` — FR-003's wrap + `Shift+Tab`-from-no-choice arithmetic incl. `len == 1`) and `refresh_test.go` (a refresh resets the selection); and (b) the **E2E** `Then` `step_t008_chat_then_marks_current_choice.go` **flipped** to assert **0** cursor rows in the rendered **suggestion block**, via the precise cursor-row predicate (`tuiCursorRows`, an exact `"  > "`-prefix match — round-037 review G-2). *(The fold attempt to scope the assertion to an "at-rest frame" via `atRestFrame` was withdrawn: review G-1 measured that the harness writes the compose keys — including any `Tab` — before the sync-marker wait (`tui_keys.go:45-50`; marker `┌`, `scenario_context.go:260`), so a navigated capture carries exactly one painted frame and the helper was the identity.)* Falsifiability witnesses are reproduced-then-reverted: re-defaulting the cursor to `0` fails the at-rest unit pin **and** the E2E `Then`; removing the reset-on-refresh fails the refresh pin.
- **Rationale**: Unlike round 036's class (structurally invisible to the E2E surface), the **cursor row is directly observable** in the capture — `tuiCursorRows` matches the exact rendered cursor-row prefix (`"  > "`, not a `TrimLeft` + `> ` proxy; round-037 review G-2) — so the E2E carrier is genuine here and is kept (operator Q4). Per-frame exactness still lives in the unit layer: the harness delivers the compose keys (incl. any `Tab`) **before** the first paint, so a navigated capture carries a **single** painted frame (round-037 review G-1), and the unit pins are the deterministic per-frame carrier (the round-016 precedent).
- **Alternatives considered**: unit-only (round-036 style) — unnecessary, since the observable E2E carrier exists and its loss would be a needless narrowing; E2E-only — loses the deterministic per-frame pin.

## 決策 5：Supersede the frozen round-016 at-rest acceptance; never edit the frozen package

- **Decision**: 037's at-rest rule **supersedes** the round-016 rule *exactly one suggestion is marked as the current choice at rest* for the **current** system. The round-016 package (`specs/plans/016-interactive-prompt-visual-parity/**`, including `ui/screens/entry.txt` which drew `>` on row 1) is **frozen history** and is **never edited**; 037 carries a fresh package with its own `ui/`.
- **Rationale**: `plan-package-frozen` — a delivered package is frozen; the current system's behaviour is carried by `specs/truth/**` (`truth-current`), which 037 modifies through its owner skills and records in `truth-delta.md`. The round-016 plan-side mockup was an **input** that mis-stated the reference; correcting the *truth* (not the frozen plan) is the disciplined move.
- **Alternatives considered**: edit the round-016 package's screens (violates `plan-package-frozen`); leave the contradiction (violates `truth-current`).

## 決策 6：Scope guard + the empty-submit divergence is a separate live issue

- **Decision**: This round changes **only** the suggestion selection's at-rest/refresh state (the `cursor` default + reset) and its truth/tests. The verified **empty-`Ctrl+S` divergence** (reference: no-op, stays in the prompt; tellme: quits with exit `0`) is **out of scope**, filed as live issue [#76](https://github.com/gosharplite/tellme/issues/76) (durable surface — round-035 G3 lesson), and recorded as a round forward item.
- **Rationale**: Mixing a turn-lifecycle change into a presentation-parity round would enlarge the blast radius (exit codes / turn dispatch) beyond the parity fix; a live issue guarantees the finding outlives the package freeze. FR-005/FR-006 record the unchanged surfaces (ordering/sources, glyph, styling, editor, placeholder, debounce, flags, exit codes, vocabulary, records, `stdout`).
- **Alternatives considered**: fold the empty-submit fix in (scope creep across the turn lifecycle); leave it entirely unrecorded (loses the finding at freeze).

## 決策 7：Terminal-mode `ui/` re-renders the at-rest frame without the cursor

- **Decision**: `/axb-ui-plan` (terminal mode) supplies plan-side screens under `specs/plans/037-interactive-prompt-no-selection/ui/screens/` showing the at-rest frame with **no** `>` cursor (and the suggestion list otherwise byte-identical). `/axb-system-analysis` **reviews** these (never redraws them); they are plan-side artifacts, never `TruthArtifact`s.
- **Rationale**: The `-i` prompt is a CLI that ships a TUI, so `/axb-ui-plan` runs in terminal mode (per the standing CLI streamlining); the visible at-rest surface is the round's subject, so a reviewable screen belongs in the plan package. `prototype-plan-side-only` holds.
- **Alternatives considered**: skip the `ui/` (the round's whole visible surface would lack a reviewable artifact); HTML mode (n/a — no web frontend).
