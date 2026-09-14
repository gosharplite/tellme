# Terminal / TUI Prototype Plan — tellme `-i` interactive prompt

## Interface scope

- Target interface: `tellme -i / --interactive` interactive prompt (the `cli` interface that ships a TUI)
- medium: `terminal`
- Requirement source: `spec.md` US1–US4; FR-001–FR-015; NFR-001–NFR-005; `features/acceptance/**` (4 journeys)
- Upstream basis: `spec.md`, `specs/plans/015-interactive-tui-prompt/features/acceptance/composing-a-prompt-with-live-suggestions.feature`, `…/sharing-the-prompt-log.feature`, `…/seeing-the-session-dashboard.feature`, `…/choosing-between-the-interactive-prompt-and-plain-input.feature`
- Output order: finish this `ui/ui-plan.md` first, then produce `ui/screens/*.txt` frames from it.

## Terminal visual direction

- Style source: reuse tellme's existing CLI vocabulary (single screen, monospace, no color assumptions — the render degrades gracefully on a monochrome terminal).
- Conclusion: one dense-but-scannable frame — a thin session-dashboard header, a multi-line editor, a suggestion list beneath the editor, and a single status + keybinding line at the bottom.
- Visual focus: the suggestion **selection cursor** (`▸`), the editor **caret** (`▏`), a clear separation between the editor and the suggestion list, and a muted style for the disabled/empty state.
- Aesthetic principle: each frame must look like the real terminal render. The keybinding table and the state-transition list live **in this plan**, never inside a frame.

## Screens and flow

### 1. `entry` — start / edit frame

- Frame file: `ui/screens/entry.txt`
- Goal: the operator composes a prompt while seeing the live suggestion list and the session dashboard — the entry point of the whole interactive flow.
- Entry condition: the operator starts `tellme -i` (or `USE_TUI_PROMPT` is enabled) on a terminal with no prompt.
- Primary keys: type text; `Tab` / `Shift+Tab` accept/cycle a suggestion; `Ctrl+S` / `Alt+Enter` submit; `Enter` newline; `Esc` / `Ctrl+C` abort; `?` help.
- Success transition: a submit runs one reasoning turn; on return the frame is redrawn with the dashboard updated (`20-dashboard`).
- Failure feedback: an empty submit shows an inline error row **in place** (no transition), keeping the typed content.

### 2. `10-suggestion` — path-completion frame

- Frame file: `ui/screens/10-suggestion.txt`
- Goal: when the input looks like a path, show matching workspace entries in the suggestion list (and exclude noisy directories).
- Entry condition: the operator types a path-like query (`.`, `./`, `/tmp/`).
- Primary keys: `Tab` / `Shift+Tab` cycle the selection; `Esc` closes the list.
- Success transition: accepting inserts the selected path into the editor and returns to `entry`.
- Failure feedback: with no matches, the list shows a muted "no suggestions" placeholder and the editor content is unchanged.

### 3. `20-dashboard` — post-turn frame

- Frame file: `ui/screens/20-dashboard.txt`
- Goal: show the session dashboard after a turn completes (turn count and token usage updated), ready for the next prompt.
- Entry condition: a turn submitted from `entry` has completed.
- Primary keys: `Ctrl+S` submit; `?` help; `Esc` abort.
- Success transition: back to `entry` behaviour with the refreshed dashboard.
- Failure feedback: a provider failure shows an inline error row and stays on this frame.

### 4. `30-error` — inline error frame

- Frame file: `ui/screens/30-error.txt`
- Goal: show an inline, user-understandable error row without discarding the operator's input.
- Entry condition: an empty submit, or a provider failure surfaced at submit.
- Primary keys: continue editing; `Ctrl+S` re-submit; `Esc` abort.
- Success transition: correcting the input and submitting returns to the normal flow.
- Failure feedback: the error row names the cause and the next step.

### 5. `40-help` — keybinding overlay

- Frame file: `ui/screens/40-help.txt`
- Goal: show the available keybindings over the editor so the operator can discover them.
- Entry condition: the operator presses `?` from any interactive frame.
- Primary keys: `Esc` closes the overlay and returns to the previous frame.
- Success transition: `Esc` → back to `entry`.
- Failure feedback: none (a read-only overlay).

## Keybinding table

| Key | Action | Target frame | Expected outcome |
| --- | --- | --- | --- |
| `Enter` | Insert a newline | `entry` | the editor gains a line; the line break is preserved |
| `Tab` / `Shift+Tab` | Cycle forward / backward through suggestions | `entry` | the selected suggestion is inserted into the editor (last token replaced for a single-token suggestion) |
| `Ctrl+S` / `Alt+Enter` | Submit the prompt | `entry` | exactly one reasoning turn runs with the trimmed text; on return the dashboard updates (`20-dashboard`) |
| `Esc` / `Ctrl+C` | Abort the prompt | (exit) | no provider request; tellme exits `0` |
| `?` | Show the keybinding overlay | `40-help` | the overlay opens; `Esc` returns to `entry` |
| (empty submit) | Submit with empty input | `30-error` | an inline error row appears; the editor content is kept |

## State-transition list

1. `entry` --type text--> `entry` (the suggestion list refreshes after a short debounce; newest-first, deduped)
2. `entry` --`Tab`--> `entry` (the selected suggestion is inserted into the editor)
3. `entry` --`Ctrl+S`/`Alt+Enter` (non-empty)--> running -> `20-dashboard` (one turn runs; the dashboard shows the updated turn count and tokens)
4. `entry` --`Ctrl+S`/`Alt+Enter` (empty)--> `30-error` (inline error row; no transition, content kept)
5. `entry` --`?`--> `40-help` --`Esc`--> `entry`
6. any frame --provider failure--> `30-error` (inline error row naming the cause)
7. any frame --`Esc`/`Ctrl+C`--> exit `0` (no request)

## Interaction & fake-data principles

- Fake-data strategy: render the suggestion list, the editor and the dashboard from fake recent prompts, fake workspace entries and fake session metrics so a reviewer can feel the flow without a real provider.
- Interaction principles: every frame maps back to an operable key; the state changes (suggestion selection, editor content, dashboard values) are shown with fake data.
- Content principles: frames contain only what the terminal would render to the operator; the keybinding table and state-transition list stay in this plan.

## State & information disclosure

- Operator-visible information: the typed prompt, the suggestion list, the focused region, the session dashboard (provider/model, tokens, turns), and the error/hint row.
- Must-be-hidden information: internal tool arguments, the full filesystem index, and any provider/server internals that should not leak into the surface.
- Main UI states: start, editing, suggestions visible, submitted/running, post-turn summary, error.
- Role / permission differences: single operator; no multi-role differences.

## Validation & error feedback

- Input validation: a submit is checked non-empty; an empty submit shows an inline error row and keeps the content.
- State-conflict handling: while a turn is running the submitted content remains visible and no new submit is accepted until the flow returns.
- User-understandable error messages: the error row states the cause and the next step (for example, "input must not be empty — type a prompt and submit again").

## Prototype output plan

- Entry frame: `ui/screens/entry.txt`
- Planned frame files: `ui/screens/entry.txt`, `ui/screens/10-suggestion.txt`, `ui/screens/20-dashboard.txt`, `ui/screens/30-error.txt`, `ui/screens/40-help.txt`
- Frame-transition principle: switch the whole frame only when the terminal would really redraw the screen; otherwise prefer to swap a region inside the same frame.
- Review goal: a reviewer can trace **key → state → outcome** from this plan plus the frames, without reading Markdown gloss inside a frame.

## Terminal implementation split (suggested)

- Pane / component split: the suggestion-list panel, the multi-line editor panel, the dashboard header, and the status/keybinding line.
- Shared blocks: the selection cursor, the status line, the error row, and the panel borders.
- Dependencies on the backend contract: the suggestion engine's inputs/outputs (shared prompt log + session + filesystem + tools), the shared-log write, and the session metrics for the dashboard.
- Acceptance focus: the flow from start → type → accept a suggestion → submit → dashboard/answer must be continuous and understandable, and must not expose internal tool arguments or the full filesystem index.
