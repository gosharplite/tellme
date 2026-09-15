# Terminal / TUI Prototype Plan — tellme `-i` submit teardown & handoff (round 023)

## Interface scope

- Target interface: `tellme -i / --interactive` interactive prompt (the `cli` interface that ships a TUI)
- medium: `terminal`
- Requirement source: `spec.md` (US1–US2; FR-001–FR-014); `features/acceptance/**` (2 journeys)
- Upstream basis: `spec.md`, `features/acceptance/handing-back-the-terminal.feature`, `features/acceptance/continuing-on-the-standard-surface.feature`
- Reference (parity target): `tell-me-go`'s `internal/ui/tui/prompt/model.go` (`View()` returns empty once submitted/aborted) and `prompt_capturer.go` (`provideFeedback` after the TUI closes)
- Output order: finalize this `ui/ui-plan.md` first, then produce `ui/screens/*.txt` frames from it.

## Terminal visual direction

- Style source: the round-016 `-i` chrome (a bordered multi-line editor above a suggestion list, no dashboard header) for the entry frame, plus the round-017/018/019 **standard turn surface** the submit hands to.
- Conclusion: the interactive prompt is a **transient** surface. On submit (`Ctrl+S`/`Alt+Enter`) or abort (`Esc`/`Ctrl+C`) the editor frame is **cleared** — no residue — and the terminal then shows the standard turn surface (the echoed prompt · input-capture acknowledgement · `─` rule + `╭─⠿ Turn <N>` header · live spinner · answer · post-turn status).
- Visual focus: the editor border; the **cleared** frame (the box does not linger); the handoff to the standard surface. The operator must be able to see the prompt they sent (it is echoed, because the box that held it is gone).
- Aesthetic principle: each frame looks like the real terminal render. The keybinding table and the state-transition list live **in this plan**, never inside a frame.

## Screens and flow

### 1. `entry` — the editor (unchanged chrome)

- Frame file: `ui/screens/entry.txt`
- Goal: the operator composes a prompt in the bordered editor while the suggestion list shows recent prompts — the surface's entry point.
- Entry condition: the operator starts `tellme -i` (`-i` / `USE_TUI_PROMPT`) on a terminal with no prompt.
- Primary keys: type text; `Tab` / `Shift+Tab` accept/cycle a suggestion; `Ctrl+S` / `Alt+Enter` submit; `Enter` newline; `Esc` / `Ctrl+C` abort.
- Success transition: a submit **clears the frame** (this round) and the run continues on the standard turn surface (frame `10-submitted`).
- Failure feedback: no inline error row (strict parity) — an empty submit simply does not submit; the operator keeps editing.

### 2. `10-submitted` — after `Ctrl+S`: the editor is gone, the standard surface continues

- Frame file: `ui/screens/10-submitted.txt`
- Goal: show the post-submit terminal — the bordered editor is **cleared**, the submitted prompt is **echoed**, and the run continues exactly like the positional / Ctrl+D surfaces.
- Entry condition: the operator submits a non-empty prompt at the interactive prompt.
- Primary keys: (none — the prompt has closed); the run proceeds on its own.
- Success transition: the standard turn surface runs to completion (answer on `stdout`, status on `stderr`).
- Failure feedback: a provider/history failure surfaces on the normal `stderr` class-phrase path (outside the prompt), exactly as on the other surfaces.

### 3. `20-abandoned` — after `Esc`: the editor is gone, no request

- Frame file: `ui/screens/20-abandoned.txt`
- Goal: show that aborting the prompt also clears the editor and sends nothing.
- Entry condition: the operator aborts (`Esc` / `Ctrl+C`) with or without text.
- Primary keys: (none — the prompt has closed).
- Success transition: the editor is cleared with no residue; tellme exits `0` with no provider request.
- Failure feedback: none.

## Keybinding table

| Key | Action | Target frame | Expected outcome |
| --- | --- | --- | --- |
| `Ctrl+S` / `Alt+Enter` | Submit the prompt | (clear) → `10-submitted` | the editor frame is **cleared**; the submitted (trimmed) prompt is **echoed**; exactly one reasoning turn runs on the standard surface |
| `Esc` / `Ctrl+C` | Abort the prompt | (clear) → `20-abandoned` | the editor frame is **cleared**; no provider request; tellme exits `0` |
| `Tab` / `Shift+Tab` | Cycle / accept a suggestion | `entry` | the selected suggestion is inserted into the editor (last token replaced for a single-token suggestion) |
| `Enter` | Insert a newline | `entry` | the editor gains a line; the line break is preserved |
| (empty submit) | Submit with empty input | `entry` | no transition; no turn runs; the editor keeps its content |

## State-transition list

1. `entry` --`Ctrl+S`/`Alt+Enter` (non-empty)--> **clear frame** → `10-submitted` (one turn runs; the answer prints to `stdout`)
2. `entry` --`Esc`/`Ctrl+C`--> **clear frame** → `20-abandoned` (no request; exit `0`)
3. `entry` --`Ctrl+S`/`Alt+Enter` (empty)--> `entry` (no transition; content kept)
4. `entry` --resize--> `entry` (border + suggestion list reflow to the terminal width)
5. `10-submitted` --turn completes--> the terminal holds the last answer + post-turn status (no frame to clear)

## Interaction & fake-data principles

- Fake-data strategy: render the editor, placeholder, and suggestion list from fake recent prompts; render the post-submit frame from a fake answer + fake usage figures so a reviewer can follow the handoff.
- Interaction principles: every frame maps back to an operable key; the teardown (frame cleared) and the handoff to the standard surface are the load-bearing behaviours.
- Content principles: frames contain only what the terminal would render; the keybinding/state lists stay in this plan.

## State & information disclosure

- Operator-visible information: the typed prompt; the echoed prompt after submit; the suggestion list; the standard turn surface (input-capture, frame, payload, answer, post-turn status).
- Must-be-hidden information: any session metrics header inside the prompt (absent — strict parity); internal tool arguments; the full filesystem index.
- Main UI states: editing, suggestions visible, submitted (cleared → standard surface), abandoned (cleared).
- Role / permission differences: single operator; none.

## Validation & error feedback

- Input validation: a submit is checked non-empty; an empty submit keeps the content (no inline error row — strict parity).
- State-conflict handling: a superseded suggestion fetch never overwrites a newer result (unchanged).
- User-understandable error messages: not inside the prompt; failures surface on the normal `stderr` class-phrase path.

## Prototype output plan

- Entry frame: `ui/screens/entry.txt`
- Planned frame files: `ui/screens/entry.txt`, `ui/screens/10-submitted.txt`, `ui/screens/20-abandoned.txt`
- Frame-transition principle: switch the whole frame only when the terminal would really redraw; the teardown is a full clear-to-standard-surface transition.
- Review goal: a reviewer can trace **key → state → outcome** and confirm that (a) the editor frame is cleared on submit/abort, (b) the submitted prompt is echoed, and (c) the run continues on the standard turn surface.

## Terminal implementation split (suggested)

- Pane / component split: unchanged from round 016 (bordered editor + suggestion list); this round adds the model's **clear-on-submit/abort** and the CLI's **chrome/echo** handoff.
- Shared blocks: the editor border, the placeholder, the suggestion list (unchanged); the standard turn-surface emitters (rounds 017/018/019) are reused.
- Dependencies on the backend contract: the suggestion engine (unchanged); the prompt emits only the composed text; the echoed prompt + chrome are `stderr`-only.
- Acceptance focus: on submit/abort the editor frame is gone; on submit the prompt is echoed and the standard turn surface (chrome + spinner + status) follows, identical to the other prompt surfaces.
