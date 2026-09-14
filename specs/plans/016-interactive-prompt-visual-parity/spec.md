# Feature Specification: tellme `-i` Interactive TUI Prompt — Strict Visual Parity with tell-me-go (round 016)

**Feature Branch**: `016-interactive-prompt-visual-parity`

**Created**: 2026-09-14

**Status**: Draft — parity decision locked (strict parity)

**Input**: Operator request "I would like to have the `-i` feels of tellme same as tell-me-go" — anchor issue [#39](https://github.com/gosharplite/tellme/issues/39). Round 015 delivered the `-i` / `--interactive` TUI prompt with the right **behaviour** (keybindings; injected-stream containment so `stdout` stays byte-exact; suggestions; shared prompt log) but a **barer chrome** than the reference: the editor has no border/placeholder, the suggestion list is unstyled, there is no width handling, suggestions refresh synchronously (no debounce/async), `Tab` cycles but does not insert, and the prompt renders an inline **dashboard header** the reference does not have. This round aligns the `-i` **prompt surface** to `tell-me-go`'s.

Behaviour intent: **MODIFY** the `-i` prompt surface (its chrome + suggestion interaction) to match the reference; **DELETE** the dashboard header from the `-i` surface (superseding round-015 US3 for the prompt); **NOOP** everywhere else (the shared prompt log, suggestion sources, the non-TTY fallback, the round-012 plain reader, and every non-`-i` path are unchanged).

**Locked decision (2026-09-14, by the operator)**: **Strict parity.** The `-i` prompt MUST match `tell-me-go`'s prompt surface exactly — a **bordered multi-line editor** above a **styled suggestion list**, with **no dashboard header** and **no status/keybinding line** (the keybinding hints live in the editor placeholder).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The prompt surface looks like tell-me-go's (Priority: P1)

As an operator, I want `tellme -i` to present the same prompt surface as `tell-me-go -i` — a framed multi-line editor above a styled suggestion list, with a placeholder that names the submit/abort keys and no dashboard header — so the tool feels like its reference and I can move between them without relearning the screen.

**Why this priority**: it is the round's whole reason to exist — strict chrome parity with the reference; the interaction parity (Story 2) and the responsiveness (Story 3) only matter once the surface is framed correctly.

**Independent verification**: render the prompt hermetically (injected streams + scripted keys, no pty) and assert the chrome — a border is drawn around the editor; the editor shows the placeholder hint; the suggestion list shows a header row and a highlighted selection cursor; and **no** dashboard header row is present — and that the previous round-015 dashboard header no longer appears.

**Acceptance Scenarios**:

1. **Given** `tellme` is invoked with `-i` on a terminal, **When** the prompt renders, **Then** the editor is drawn inside a visible border and the suggestion list is drawn beneath the editor (the reference layout), with the surrounding block padded.
2. **Given** the prompt is open, **When** the editor is rendered, **Then** it shows a placeholder hint naming the submit key and the abort key, supports multi-line editing, and shows no line numbers.
3. **Given** the prompt is open with suggestions, **When** the list is rendered, **Then** it shows a header row and visually distinguishes the selected suggestion (a cursor marker) from the unselected ones.
4. **Given** the prompt is open, **When** it renders, **Then** it shows **no** session dashboard header (no provider/token/turn header row).

**Functional Requirements**:

- **FR-001**: The `-i` prompt MUST render the editor inside a visible border, with the prompt block padded, and the suggestion list directly beneath the editor — the reference layout (editor over suggestions).
- **FR-002**: The editor MUST be multi-line and MUST display a placeholder hint that names the submit keybinding (`Ctrl+S` / `Alt+Enter`) and the abort keybinding (`Esc`); it MUST NOT show line numbers.
- **FR-003**: The suggestion list MUST render a header row and a visible selection cursor that distinguishes the selected suggestion from the unselected ones.
- **FR-004**: The `-i` prompt MUST NOT display a session dashboard header (this supersedes round-015 `FR-012` / `FR-013` for the `-i` surface).

**Non-Functional Requirements**:

- **NFR-001**: The prompt chrome MUST degrade gracefully on a monochrome terminal — the selection MUST remain distinguishable without relying on color (the cursor marker distinguishes it).

---

### User Story 2 - Suggestions behave like tell-me-go's (Priority: P1)

As an operator, I want typing and accepting suggestions to behave like the reference — suggestions settle after a short debounce without blocking my typing, over-long suggestions never appear, and `Tab`/`Shift+Tab` accept a suggestion into the editor — so composing a prompt is as smooth as in `tell-me-go`.

**Why this priority**: it is the second half of "same feel" — the prompt may look right (Story 1) but still feel wrong if suggestions do not settle, block typing, or do not insert on `Tab`.

**Independent verification**: drive the model with scripted keys against an injected suggestion source (no pty); assert suggestions refresh after the debounce and not on every keystroke, that an entry with more than three lines is dropped, that a superseded (stale) fetch never overwrites a newer result, and that `Tab`/`Shift+Tab` insert the selection — replacing only the last token for a single-token suggestion on a multi-word line.

**Acceptance Scenarios**:

1. **Given** the prompt is open and the operator types, **When** the input changes, **Then** the suggestions refresh after a short debounce, computed without blocking the visible input.
2. **Given** a suggested entry spans more than three lines, **When** the suggestions are shown, **Then** that entry is excluded from the list.
3. **Given** a suggestion list is shown, **When** the operator presses `Tab` (`Shift+Tab`), **Then** the next (previous) suggestion is selected **and inserted** into the editor.
4. **Given** the editor holds a multi-word line and the selected suggestion is a single token (for example a file path), **When** the operator accepts it, **Then** only the last token is replaced and the preceding text is preserved.
5. **Given** a fetch for an older keystroke is still in flight, **When** a newer keystroke supersedes it, **Then** the stale result is discarded and never overwrites the newer suggestions.

**Functional Requirements**:

- **FR-005**: As the input changes, the system MUST refresh the suggestions after a short debounce, computed asynchronously so the visible input is never blocked.
- **FR-006**: The system MUST exclude any suggestion that spans more than three lines.
- **FR-007**: `Tab` / `Shift+Tab` MUST cycle forward / backward AND insert the selected suggestion into the editor — replacing only the last token when the input has multiple words and the suggestion is a single token, otherwise replacing the whole line.
- **FR-008**: When the input is path-like, suggestions MUST include matching workspace entries and MUST exclude noisy directories (for example `.git`, `node_modules`) — unchanged from round 015.

**Non-Functional Requirements**:

- **NFR-002**: A superseded suggestion fetch MUST be cancelable, so a stale result cannot overwrite a newer one.

---

### User Story 3 - The prompt fits the terminal (Priority: P2)

As an operator, I want the prompt to use the width of my terminal and reflow when I resize it, so the editor and suggestions read cleanly at any window size.

**Why this priority**: it depends on the surface existing (Story 1) and improves legibility rather than changing the capability; the reference adapts the width on resize.

**Independent verification**: send a `WindowSizeMsg` with a new width to the model (hermetic) and assert the editor width adapts to the new terminal width (reference: `width - 4`), and that a very narrow width degrades without panicking.

**Acceptance Scenarios**:

1. **Given** the prompt is open, **When** the terminal is resized, **Then** the editor width adapts to the new terminal width.
2. **Given** a terminal narrower than the prompt needs, **When** the prompt renders, **Then** it degrades without panicking.

**Functional Requirements**:

- **FR-009**: On a terminal resize, the editor width MUST adapt to the new terminal width.

**Non-Functional Requirements**:

- **NFR-003**: The prompt MUST render without panicking at any terminal width, including widths too narrow for the full layout.

---

### Edge Cases

- **Narrow terminal width**: the layout degrades (clamped width) and MUST NOT panic.
- **Monochrome / no-color terminal**: the composed surface still reads — the selection cursor, not color, distinguishes the selected suggestion (NFR-001).
- **Empty suggestion list**: the list area renders nothing (no header-only artifact) — the editor keeps focus.
- **`-i` with a non-terminal stdin**: no TUI renders — the round-015 fallback (round-015 `FR-014` / `NFR-001`) applies unchanged.
- **Multi-word input + single-token suggestion**: `Tab` replaces only the last token (FR-007).
- **Suggestion longer than the terminal width**: wrapping/truncation is a `/axb-ui-plan` (terminal) detail; the row MUST NOT break the frame.
- **Stale fetch**: a superseded keystroke's fetch result is discarded (NFR-002).

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-010**: The round MUST NOT regress rounds 001–015, with the single exception of the retired dashboard header on the `-i` surface — every other prior acceptance scenario MUST remain green, `stdout` MUST remain byte-exact (the prompt renders to the terminal/diagnostic stream, not into a piped `stdout`), the frozen class-phrase vocabulary MUST be unchanged, and the round-015 keybindings (`Ctrl+S`/`Alt+Enter` submit, `Enter` newline, `Tab`/`Shift+Tab` cycle, `Esc`/`Ctrl+C` abort) MUST be preserved. No new class phrase is introduced.
- **FR-011**: The shared prompt log (`$TELL_ME_HOME/output/global_prompts.jsonl`), its read/write semantics, the suggestion *sources* (shared log + session + workspace + tools), the non-TTY fallback, and the round-012 plain reader MUST be unchanged.

#### Non-Functional Requirements

- **NFR-004**: All new assertions MUST be deterministic (no `time.Sleep`, no real pty); the injected streams, a scripted keystroke source, and direct style/state assertions are the verification surface.

### Key Entities *(include if feature involves data)*

- **Interactive prompt surface**: the `-i` chrome — a bordered multi-line editor over a styled suggestion list, with the reference layout and no dashboard header.
- **Editor**: the multi-line input box — bordered, placeholder-bearing, multi-line, no line numbers.
- **Suggestion list**: the ranked candidate list — a header row plus rows, with a cursor marker distinguishing the selection, refreshed on a debounce and excluding over-long entries.
- **Reference parity contract**: the tell-me-go chrome tokens the surface must reproduce (border style/color, root padding, editor height/width, placeholder text, list header text, selection styles) — the exact values are an `/axb-ui-plan` (terminal) + `/axb-technical-research` determination.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The `-i` prompt renders the reference chrome — a bordered multi-line editor, a placeholder naming the submit/abort keys, and a suggestion list with a highlighted cursor — and renders **no** dashboard header, verified hermetically (SC-001 covers FR-001–FR-004).
- **SC-002**: 100% of `Tab`/`Shift+Tab` acceptances insert the selection, including the last-token replacement for a multi-word input with a single-token suggestion.
- **SC-003**: Suggestions refresh after the debounce (not per keystroke), exclude entries spanning more than three lines, discard stale fetches, and adapt the editor width on resize.
- **SC-004**: Every prior acceptance scenario (rounds 001–015) remains green except the retired dashboard rule, and the frozen class-phrase vocabulary is unchanged.
- **SC-005**: The prompt surface behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (parity target)**: **strict parity** is locked by the operator (#39); the reference is `tell-me-go`'s `internal/ui/tui/prompt/{model,textarea,suggester}.go` (and `docs/user/tui-prompt.md`).
- **A2 (dashboard)**: the round-015 session-dashboard header is removed from the `-i` surface and is **not** relocated (strict parity); the existing round-009 payload status line (emitted on the diagnostic stream, outside the prompt) is unaffected.
- **A3 (unchanged)**: the shared prompt log, the suggestion sources, the non-TTY fallback, the round-012 plain reader, and every non-`-i` path are unchanged.
- **A4 (platform / dependency)**: the round is **POSIX-only** (round-015 precedent) and introduces **no new dependency** — the Bubble Tea family (`bubbletea` + `bubbles` + direct `lipgloss`) is already present.
- **A5 (chrome tokens)**: the exact parity tokens (border style/foreground, root padding, editor fixed height/width, placeholder text, list header text, selection styles) are a `/axb-ui-plan` (terminal mode) + `/axb-technical-research` determination — named here as the parity target, not as FR text.
- **A6 (verification)**: verification is hermetic — injected streams + a scripted keystroke source + style/state assertions; no real pty is required.
