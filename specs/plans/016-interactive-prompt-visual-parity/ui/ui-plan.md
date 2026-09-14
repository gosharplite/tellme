# Terminal / TUI Prototype Plan — tellme `-i` interactive prompt (strict parity)

## Interface scope

- Target interface: `tellme -i / --interactive` interactive prompt (the `cli` interface that ships a TUI)
- medium: `terminal`
- Requirement source: `spec.md` (US1–US3; FR-001–FR-011; NFR-001–NFR-004); `features/acceptance/**` (3 journeys); anchor issue [#39](https://github.com/gosharplite/tellme/issues/39)
- Upstream basis: `spec.md`, `features/acceptance/seeing-a-prompt-that-matches-the-reference.feature`, `…/composing-with-settled-suggestions.feature`, `…/fitting-the-terminal-window.feature`
- Reference (parity target): `tell-me-go`'s `internal/ui/tui/prompt/{model,textarea,suggester}.go` and `docs/user/tui-prompt.md`
- Output order: finish this `ui/ui-plan.md` first, then produce `ui/screens/*.txt` frames from it.

## Terminal visual direction

- Style source: reproduce `tell-me-go`'s prompt chrome exactly — a **bordered** multi-line editor above a **suggestion list**; **no** dashboard header and **no** status/keybinding line (the keybinding hints live in the editor placeholder).
- Conclusion: the prompt is two stacked blocks — the editor inside a square-corner border, then the suggestion list — with padding around the whole block.
- Visual focus: the editor border, the placeholder hint, the suggestion list's selected row (a `>` cursor), and the deliberate absence of any metrics header.
- Aesthetic principle: each frame must look like the real terminal render. The keybinding table and the state-transition list live **in this plan**, never inside a frame.
- Frame fidelity: the frames show the reference chrome at the editor's **fixed height** (`lipgloss` editor height 10) and a representative terminal width; the live render sets the width from the terminal (`tea.WindowSizeMsg` → `width - 4`) and clamps on a narrow terminal — `30-narrow` illustrates the clamped form.

## Screens and flow

### 1. `entry` — start / edit frame

- Frame file: `ui/screens/entry.txt`
- Goal: the operator composes a prompt in the bordered editor while the suggestion list shows the recent prompts — the entry point of the whole interactive flow.
- Entry condition: the operator starts `tellme -i` (or `USE_TUI_PROMPT` is enabled) on a terminal with no prompt.
- Primary keys: type text; `Tab` / `Shift+Tab` accept/cycle a suggestion; `Ctrl+S` / `Alt+Enter` submit; `Enter` newline; `Esc` / `Ctrl+C` abort.
- Success transition: a submit closes the prompt and runs one reasoning turn; the answer prints to `stdout` outside the prompt.
- Failure feedback: there is no inline error row (strict parity) — an empty submit simply does not submit; the operator keeps editing.

### 2. `10-suggestion` — path-completion frame

- Frame file: `ui/screens/10-suggestion.txt`
- Goal: when the input looks like a path, show matching workspace entries in the suggestion list (and exclude noisy directories).
- Entry condition: the operator types a path-like query (`.`, `./`, `/tmp/`).
- Primary keys: `Tab` / `Shift+Tab` cycle + insert the selection; `Esc` aborts.
- Success transition: accepting inserts the selected path into the editor (replacing the last token).
- Failure feedback: with no matches, the list shows nothing; the editor content is unchanged.

### 3. `20-composed` — multi-line composed frame

- Frame file: `ui/screens/20-composed.txt`
- Goal: show a multi-line prompt being composed (Enter inserts newlines), with the suggestions still offering matches.
- Entry condition: the operator has typed across several lines.
- Primary keys: `Ctrl+S` / `Alt+Enter` submit; `Enter` newline; `Tab` accept; `Esc` abort.
- Success transition: a submit closes the prompt and runs one turn.
- Failure feedback: none (the editor keeps every line; no error row).

### 4. `30-narrow` — narrow-terminal frame

- Frame file: `ui/screens/30-narrow.txt`
- Goal: show the prompt on a terminal narrower than the reference default — the editor border and the suggestion list shrink to the terminal width without breaking.
- Entry condition: the terminal is narrower than the editor's default width.
- Primary keys: same as `entry`.
- Success transition: widening the terminal reflows back to `entry`.
- Failure feedback: none — the prompt degrades without panicking.

## Keybinding table

| Key | Action | Target frame | Expected outcome |
| --- | --- | --- | --- |
| `Enter` | Insert a newline | `entry` | the editor gains a line; the line break is preserved |
| `Tab` / `Shift+Tab` | Cycle forward / backward through suggestions | `entry` | the selected suggestion is **inserted** into the editor (the last token replaced for a single-token suggestion on a multi-word line) |
| `Ctrl+S` / `Alt+Enter` | Submit the prompt | (close) | exactly one reasoning turn runs with the trimmed text; the prompt closes and the answer prints to `stdout` |
| `Esc` / `Ctrl+C` | Abort the prompt | (exit) | no provider request; tellme exits `0` |
| (empty submit) | Submit with empty input | `entry` | no turn runs; the editor keeps its content (no error row — strict parity) |

*(There is no `?` help overlay — strict parity with the reference, which has none.)*

## State-transition list

1. `entry` --type text--> `entry` (the suggestion list refreshes after a short debounce; newest-first, deduped)
2. `entry` --`Tab`/`Shift+Tab`--> `entry` (the selected suggestion is inserted into the editor)
3. `entry` --`Ctrl+S`/`Alt+Enter` (non-empty)--> close → one turn runs (the answer prints to `stdout`)
4. `entry` --`Ctrl+S`/`Alt+Enter` (empty)--> `entry` (no transition; content kept)
5. `entry` --resize (narrow)--> `30-narrow` --resize (wider)--> `entry`
6. any frame --`Esc`/`Ctrl+C`--> exit `0` (no request)

## Interaction & fake-data principles

- Fake-data strategy: render the editor, the placeholder, and the suggestion list from fake recent prompts and fake workspace entries so a reviewer can feel the flow without a real provider.
- Interaction principles: every frame maps back to an operable key; the state changes (suggestion selection, editor content, reflow) are shown with fake data.
- Content principles: frames contain only what the terminal would render to the operator; the keybinding table and state-transition list stay in this plan.

## State & information disclosure

- Operator-visible information: the typed prompt, the suggestion list, the selected suggestion (the `>` cursor), and the placeholder hint.
- Must-be-hidden information: any session metrics header (removed by strict parity), internal tool arguments, and the full filesystem index.
- Main UI states: start/editing, suggestions visible, composed multi-line, narrow-terminal.
- Role / permission differences: single operator; no multi-role differences.

## Validation & error feedback

- Input validation: a submit is checked non-empty; an empty submit does not submit and keeps the content — there is no inline error row (strict parity removes the round-015 dashboard/error chrome).
- State-conflict handling: a superseded (stale) suggestion fetch never overwrites a newer result.
- User-understandable error messages: not applicable inside the prompt (errors surface on the normal `stderr` class-phrase path outside the prompt).

## Prototype output plan

- Entry frame: `ui/screens/entry.txt`
- Planned frame files: `ui/screens/entry.txt`, `ui/screens/10-suggestion.txt`, `ui/screens/20-composed.txt`, `ui/screens/30-narrow.txt`
- Frame-transition principle: switch the whole frame only when the terminal would really redraw the screen; otherwise prefer to swap a region inside the same frame.
- Review goal: a reviewer can trace **key → state → outcome** from this plan plus the frames, and confirm the surface matches the reference — a framed editor over a suggestion list, and no metrics header.

## Terminal implementation split (suggested)

- Pane / component split: the bordered editor panel and the suggestion-list panel (the reference's two stacked blocks).
- Shared blocks: the editor border, the placeholder, the selection cursor, and the suggestion list header.
- Dependencies on the backend contract: the suggestion engine's inputs/outputs (shared prompt log + session + filesystem + tools) — unchanged from round 015; the prompt emits only the composed text.
- Acceptance focus: the flow from start → type → accept a suggestion → submit must read like the reference, must let the editor border and suggestion list reflow on resize, and must not expose a metrics header or internal tool arguments.
