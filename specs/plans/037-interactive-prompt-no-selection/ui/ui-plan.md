# UI Plan — round 037 (interactive prompt no-selection, terminal mode)

Plan-side terminal screens for the `-i` prompt's at-rest selection state. The `-i`
prompt ships a rich terminal UI, so `axb-ui-plan` runs in **terminal mode** (no HTML).

## Screens

- `screens/entry.txt` — the at-rest frame with **no** suggestion pre-selected, plus the
  after-`Tab` frame where the first suggestion becomes the current choice.

## Scope

- Only the **selection state** changes: the at-rest frame carries **no** `>` cursor row;
  the after-`Tab` frame carries the cursor on the **first** row.
- Byte-identical otherwise: the bordered editor, the placeholder, the `Suggestions:`
  header, the styled rows, and the row order are unchanged.

## Supersession

Supersedes the round-016 at-rest mockup (`specs/plans/016-interactive-prompt-visual-parity/ui/screens/entry.txt`),
which drew `>` on the first row. That package is **frozen history** and is not edited.

## Not truth

These screens are plan-side artifacts supplied for `/axb-system-analysis` review; they
never become `TruthArtifact`s (`prototype-plan-side-only`). The executable contract lives
in `specs/truth/features/cli/chat/**`.
