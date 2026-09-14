# Phase 0 Research: tellme `-i` Interactive Prompt — Strict Visual Parity with tell-me-go (Round 016)

Topic: make `tellme`'s **`-i` / `--interactive`** TUI prompt **feel identical to `tell-me-go`'s** — reproduce the reference's prompt chrome (a **bordered** multi-line editor above a **styled suggestion list**, with the keybinding hints in the placeholder) and its suggestion interaction (a debounced, cancelable refresh; `Tab`/`Shift+Tab` **insert** the selection), and **remove** the round-015 dashboard header that the reference does not have.

Anchor issue: [#39](https://github.com/gosharplite/tellme/issues/39). Round 015 delivered the `-i` surface with the right **behaviour** (keybindings; injected-stream containment so `stdout` stays byte-exact; suggestions; the shared prompt log) but a **barer chrome** than the reference. This round is a bounded **presentation-parity** change to that surface.

Scope note: the language (Go 1.26), module, the CLI/config layers, the `llm.Gateway` port + adapters, the session-history store, the agent tool loop, and the rest of the pipeline were locked in rounds 001–015. The system still has **one CLI end**; the round reaches **no new endpoint** and persists **no new state**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; `godog` running the built binary; E2E acceptance + pure-helper units) and are **not re-decided**. **IN**: the `-i` prompt chrome, the suggestion interaction (debounce/async/cancel + over-long drop + `Tab`-inserts), terminal-width handling, and the retirement of the dashboard header. **OUT**: the shared prompt log, the suggestion *sources*, the non-TTY fallback, the round-012 plain reader, the keybindings (already at parity), streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent, and a Windows variant.

---

## Decision 1: Reproduce the reference chrome on the existing Bubble Tea family (no new dependency)

- **Decision**: Build the reference chrome on the **already-present** `bubbletea` + `bubbles/textarea` + direct `lipgloss` stack (round 015) — **no new module**. Reproduce the reference's chrome tokens: the editor wrapped in a `lipgloss.NormalBorder()` with `BorderForeground` colour `240`; the root block padded `padding(1, 1)`; the editor fixed at **height 10**, width set from the terminal (`width - 4`) with the reference placeholder `Type your message here... (Alt+Enter or Ctrl+S to submit, Esc to abort)` and `ShowLineNumbers = false`; the suggestion list styled with `padding(0, 1)`, a `Suggestions:` header, a selected row (`> ` prefix, bold, fg `205` on bg `235`) and unselected rows (fg `245`), all rendered **beneath** the bordered editor.
- **Rationale**: Strict parity is a *presentation* change; it needs no new capability and no new module — the reference uses the same family the round-015 core already adopted. Reproducing the exact tokens is what makes the two prompts feel the same.
- **Alternatives considered**:
  - **Hand-roll the border/styling with ANSI** — re-implements what `lipgloss` already provides and diverges from the reference's rendering — rejected.
  - **A different border/palette** — breaks the parity the round exists to achieve — rejected.

## Decision 2: Debounced, asynchronous, cancelable suggestion refresh

- **Decision**: A query change schedules the suggestion refresh after a short **debounce** (~100 ms, matching the reference) and computes it **asynchronously**, so the visible input is never blocked; a superseded keystroke's fetch is **canceled** (`context`), so a stale result never overwrites a newer one. Suggestions spanning **more than three lines** are dropped from the list.
- **Rationale**: The round-015 model refreshed synchronously on every keystroke and had no cancel/drop logic; the reference debounces, computes off the input path, and filters over-long entries. This is the interaction half of "same feel" and satisfies `FR-005`/`FR-006`/`NFR-002`.
- **Alternatives considered**:
  - **Keep synchronous refresh** — stalls the UI on large directories and diverges from the reference — rejected.
  - **No cancel (last-writer-wins)** — a stale result can flicker/replace a newer list — rejected.

## Decision 3: `Tab` / `Shift+Tab` accept inserts the selection (last-token heuristic)

- **Decision**: `Tab` / `Shift+Tab` cycle the selection **and insert** the selected suggestion into the editor. Matching the reference heuristic: if the input has more than one word **and** the suggestion is a single token (no internal space — e.g. a path), **replace only the last token** and preserve the preceding text; otherwise **replace the whole line**. After insertion, place the cursor at the end of the editor value.
- **Rationale**: The round-015 model cycled the cursor but did **not** insert — so `Tab` did not "accept" a suggestion as the reference does. Restoring the insert (with the reference's last-token heuristic) is required by `FR-007`.
- **Alternatives considered**:
  - **Cycle-only (no insert)** — the reference accepts; cycle-only is a visible behavioural divergence — rejected.
  - **Always replace the whole line** — loses the multi-word context the reference preserves — rejected.

## Decision 4: Remove the dashboard header from the `-i` surface (strict parity)

- **Decision**: The `-i` prompt renders **only** the bordered editor and the suggestion list — the round-015 **session dashboard header** (provider/model · tokens · turns) is **removed** from the prompt surface (and not relocated). The same figures remain observable only through the round-009 payload **status line** on the diagnostic stream, outside the prompt. This **supersedes round-015 `FR-012`/`FR-013`** for the `-i` surface and **retires** the truth rule "the interactive prompt reports the session dashboard" in `specs/truth/features/cli/chat/using-the-interactive-prompt.feature` (`/axb-dsl-refine`).
- **Rationale**: `tell-me-go`'s prompt has **no** dashboard — strict parity (the operator's locked decision, #39) requires removing it. The reference's "dashboard" is a separate concern (its own progress/TUI surface), not part of the prompt box.
- **Alternatives considered**:
  - **Keep the dashboard (parity-plus)** — the operator chose strict parity — rejected.
  - **Move the dashboard elsewhere in the prompt (status line inside the box)** — still absent in the reference — rejected.

## Decision 5: Terminal-width handling and graceful narrow degradation

- **Decision**: Handle `tea.WindowSizeMsg` to set the editor width to `msg.Width - 4` (the reference's rule), and render without panicking at any width, including widths too narrow for the full layout (the reference clamps/degrades).
- **Rationale**: The round-015 model ignored resize; the reference adapts. Satisfies `FR-009`/`NFR-003`.
- **Alternatives considered**:
  - **Fixed width** — diverges from the reference and breaks on narrow terminals — rejected.

## Decision 6: Hermetic (no-pty) verification via injected I/O, scripted keys, and style/state assertions

- **Decision**: Keep the round-015 injected I/O + scripted-key model, and add **deterministic style/state assertions**: the rendered prompt contains the editor border and the placeholder; the list contains the `Suggestions:` header and a highlighted selection; the prompt contains **no** dashboard header; `Tab` inserts (including the last-token case); the refresh is debounced and drops over-long entries; a resize changes the width. E2E drives the built binary through the round-012 `TELL_ME_FORCE_STDIN_TTY` seam + a scripted stdin. No real pty.
- **Rationale**: `NFR-004` requires determinism and no pty (project precedent: round 005 grill Q6, round 006 Q2, round 015 Decision 6). Asserting chrome **presence/absence** (not exact ANSI bytes) keeps it `TERM`/profile-independent (round 006 precedent).
- **Alternatives considered**:
  - **A pty harness** — heavy, platform-fragile, previously foreclosed — rejected (named pin).
  - **Exact rendered ANSI byte comparison** — `TERM`/profile-dependent — rejected.

## Decision 7: No new dependency; POSIX-only; keybindings unchanged

- **Decision**: The round adds **no** module (`bubbletea`/`bubbles`/`lipgloss` already present); it is **POSIX-only** (round-012/015 precedent), and the keybindings are unchanged (`Enter` newline; `Ctrl+S`/`Alt+Enter` submit; `Tab`/`Shift+Tab` cycle+insert; `Esc`/`Ctrl+C` abort) — no `?` help overlay and no inline error row (strict parity with the reference).
- **Rationale**: Parity is chrome + interaction; it needs nothing new. Keeping the keybindings and the platform scope bounded preserves rounds 001–015 behaviour (`FR-010`).
- **Alternatives considered**:
  - **Add a styling/box library** — `lipgloss` already covers it — rejected.
  - **Add a `?` help overlay / error row** — the reference has neither — rejected.

## Decision 8: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog` running the built binary; strategy = E2E acceptance + fast pure-helper units — unchanged, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round changes only the `-i` prompt's presentation/interaction; it introduces no new interface, service, runner, or second end.
- **Alternatives considered**:
  - **Re-open the techstack/test questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **Dashboard retirement (Decision 4)**: a high-impact **MODIFY/DELETE** of existing truth — the round-015 dashboard rule in `using-the-interactive-prompt.feature` and the acceptance feature `seeing-the-session-dashboard.feature` are retired for the `-i` surface. The decision is **locked by the operator (#39)**, so no `/axb-clarify` is owed; the truth change itself is `/axb-dsl-refine`'s to make.
- **Chrome-token drift (Decision 1)**: the parity tokens mirror the reference at a point in time; if `tell-me-go` changes its prompt styling, parity would need re-checking (a forward-link, not a gate).
- **Real-pty fidelity (Decision 6)**: exact stream fidelity of a real TTY is not asserted (a named pin, round-005/006/012 precedent).
- **Debounce timing (Decision 2)**: the ~100 ms debounce is adopted for parity; it is a tuning knob (a forward item).
- **Deferred, still out of scope**: streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent, the Google Gemini API family (inline key), Application Default Credentials, concurrent tool-call matching, and a Windows variant.
