# Feature Specification: tellme Interactive TUI Prompt — Suggestions, Session Dashboard, and the Shared Global Prompt Log (round 015)

**Feature Branch**: `015-interactive-tui-prompt`

**Created**: 2026-09-14

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "tellme needs `tell-me-go`'s `-i` Interactive TUI Prompt" — anchor issue [#37](https://github.com/gosharplite/tellme/issues/37). Today `tellme` has only the round-012 **plain** interactive multi-line reader (`Ctrl+D`, hint on `stderr`, real isatty via `x/term`). This round adds the rich **`-i` / `--interactive` TUI prompt**: a live **suggestion engine** (recent prompts + filesystem paths + registered tools), a **session dashboard** (provider/model, token usage, turn count), a **multi-line editor**, and terminal **keybindings**. It must **interoperate** with the shared Niffler-env prompt log at `$TELL_ME_HOME/output/global_prompts.jsonl` (the same file `tell-me-go` uses). Behaviour intent: **ADD** a new prompt surface + a new planner/repeated interaction, **ADD** a data-model artifact for the shared prompt log, and **MODIFY** nothing of the existing non-`-i` paths.

This round also produces a **plan-side TUI UX artifact** (`ui/ui-plan.md` + `ui/screens/*.txt`) via `/axb-ui-plan` in **terminal mode** (never `specs/truth/**`) — the upstream gate [aixbdd-tmg#13](https://github.com/gosharplite/aixbdd-tmg/issues/13) is **RESOLVED** (PR [#14](https://github.com/gosharplite/aixbdd-tmg/pull/14) merged), so that step is un-gated.

**Clarify Round 1 (2026-09-14)** resolved three high-impact decisions:

- **Q1 → Option 1 (global-log record semantics)**: tellme **reads + writes** the shared `output/global_prompts.jsonl`, and records **only under `-i`** — mirroring `tell-me-go` (which writes on the TUI path). Every non-`-i` run writes nothing, so rounds 001–014 behavior is byte-stable.
- **Q2 → Option 1 (round-012 relationship)**: the TUI **coexists** with the round-012 plain reader — a bare terminal invocation keeps the plain reader; `-i` / `USE_TUI_PROMPT` opts into the TUI. Both keep the non-TTY fallback.
- **Q3 → Option 1 (platform scope)**: **POSIX-only** this round (round-012 precedent); no Windows keybinding/build variant.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Compose a prompt with live suggestions at the terminal (Priority: P1)

As an operator, I want to launch `tellme -i` (or enable `USE_TUI_PROMPT`) and, as I type, see live suggestions drawn from my recent prompts, matching workspace paths, and available tools — and accept one with a keystroke — so I can compose a prompt quickly and accurately without retyping.

**Why this priority**: it is the round's core capability — the rich prompt with suggestions is the whole reason the round exists; the dashboard (Story 3) and the coexist guarantee (Story 4) only matter once this surface exists.

**Independent verification**: force the injected terminal seam and feed scripted keystrokes against an injected suggestion source (no real pty); assert the initial suggestions are the empty-query top-N (deduped, newest-first), that input refreshes them (debounced, subsequence match, ≤ cap), that a path-like query yields filesystem matches while excluding noisy directories, that `Tab`/`Shift+Tab` insert the selection, and that a submit runs exactly one reasoning turn on the trimmed composed text.

**Acceptance Scenarios**:

1. **Given** `tellme` is invoked with `-i` (or `USE_TUI_PROMPT` is enabled) on a terminal, **When** the prompt opens with no input, **Then** it shows the top recent prompts from the shared global log (newest-first, deduplicated) as the initial suggestions.
2. **Given** the prompt is open and the operator types text, **When** the input changes, **Then** the suggestions refresh (after a short debounce) to entries matching the query (subsequence match), deduplicated, up to a fixed maximum.
3. **Given** the operator types a path-like query (`.`, `./`, `/tmp/`), **When** suggestions are computed, **Then** they include matching workspace entries and exclude noisy directories (for example `.git`, `node_modules`).
4. **Given** a suggestion list is shown, **When** the operator presses `Tab` (`Shift+Tab`), **Then** the next (previous) suggestion is selected and inserted into the editor (replacing the last token for a single-token suggestion).
5. **Given** the operator has composed a prompt in the editor, **When** they submit (`Ctrl+S` or `Alt+Enter`), **Then** exactly one reasoning turn runs with the trimmed composed text as the prompt.

**Functional Requirements**:

- **FR-001**: The system MUST enable the interactive TUI prompt via the `-i` / `--interactive` flag and via the `USE_TUI_PROMPT` configuration key.
- **FR-002**: When the TUI prompt opens, the system MUST show initial suggestions seeded from recent prompts (the shared global log plus the active session's prompts), newest-first, deduplicated, up to a fixed maximum.
- **FR-003**: As the input changes, the system MUST refresh the suggestions (after a short debounce) to entries matching the current input by subsequence, deduplicated, up to the maximum; an empty input MUST show the top recent prompts.
- **FR-004**: When the input is path-like, the system MUST include matching workspace file/directory entries as suggestions and MUST exclude noisy directories (for example `.git`, `node_modules`).
- **FR-005**: `Tab` / `Shift+Tab` MUST cycle forward / backward through the suggestions and insert the selected suggestion into the editor.
- **FR-006**: `Ctrl+S` (or `Alt+Enter`) MUST submit the prompt; `Enter` MUST insert a newline (multi-line editing); `Esc` / `Ctrl+C` MUST abort without sending a prompt.
- **FR-007**: The submitted text (trimmed) MUST become the prompt of exactly one reasoning turn — the identical downstream of a positional/piped prompt (one provider request; the answer printed to `stdout`), with no request on an empty/aborted submission.

**Non-Functional Requirements**:

- **NFR-001**: The TUI MUST fall back to the existing non-interactive path when stdin is not a terminal (pipes/CI) — no TUI rendering is attempted.
- **NFR-002**: Suggestion computation MUST NOT block the visible input (debounced/asynchronous), and MUST NOT scan noisy directories.

---

### User Story 2 - Share the prompt log with the rest of the Niffler env (Priority: P1)

As an operator, I want `tellme`'s prompts to live in the same shared, append-only prompt log as `tell-me-go` (and other personas) — read to seed suggestions and (under `-i`) written in the identical shape — so what I type is remembered across runs, personas, and modes, and lines round-trip.

**Why this priority**: it is the explicit compatibility requirement that makes the suggestion engine shared and useful; it is a **data-model** addition and stands alongside Story 1 as a core value (a `-i` prompt is remembered; suggestions come from a shared pool).

**Independent verification**: point the runtime home at a directory holding a `global_prompts.jsonl`, open the TUI and submit a prompt, then assert (a) suggestions were seeded from the existing lines and (b) exactly one line was appended in the identical shape `{"timestamp","prompt"}`; conversely, assert a non-`-i` run appends nothing.

**Acceptance Scenarios**:

1. **Given** the shared global prompt log already holds entries, **When** the `-i` prompt opens, **Then** it seeds its suggestions from those entries (newest-first, deduplicated).
2. **Given** an `-i` interactive prompt is submitted, **When** the submission is recorded, **Then** `tellme` appends exactly one line in the identical shape `{"timestamp":"<RFC3339>","prompt":"<text>"}` to `$TELL_ME_HOME/output/global_prompts.jsonl`.
3. **Given** a non-`-i` run (a positional prompt or piped stdin), **When** it completes, **Then** the shared global prompt log is unchanged (nothing is appended).
4. **Given** an entry appended by `tellme`, **When** it is read back by the same reader, **Then** it round-trips and can appear in suggestions.

**Functional Requirements**:

- **FR-008**: The TUI suggestion engine MUST read the shared global prompt log at `$TELL_ME_HOME/output/global_prompts.jsonl` (the `output/` root, shared across modes) to seed suggestions.
- **FR-009**: An `-i` interactive submission MUST append exactly one line to that file in the **identical shape** (`{"timestamp":"<RFC3339>","prompt":"<text>"}`), preserving its append-only, shared semantics.
- **FR-010**: A non-`-i` run MUST NOT write the shared global prompt log.
- **FR-011**: An entry written by `tellme` MUST round-trip — the identical reader reads it back and it participates in suggestion seeding.

**Non-Functional Requirements**:

- **NFR-003**: The shared-log write MUST be append-only (`O_APPEND|O_CREATE`) so a file also written by other personas/modes is never corrupted or truncated.
- **NFR-004**: Recording the prompt to the shared log MUST NOT block the visible prompt (bounded, detached from the request lifecycle).

---

### User Story 3 - See the session dashboard at the prompt (Priority: P2)

As an operator, I want the active provider/model, my token usage, and the turn count shown at the top of the prompt, so I know which model I am talking to and how much of the budget I have used before I send.

**Why this priority**: it depends on the interactive surface existing (Story 1) and adds visibility rather than a new capability; it is valuable but not the round's core.

**Independent verification**: render the prompt with injected session metrics and assert the dashboard fields (active provider/model, tokens used vs budget, turn count) and that they reflect changed inputs.

**Acceptance Scenarios**:

1. **Given** an `-i` session on a terminal, **When** the prompt is shown, **Then** the dashboard displays the active provider/model, the token usage (used vs budget), and the turn count.
2. **Given** the provider/model or the session metrics change between turns, **When** the dashboard is rendered, **Then** it reflects the current values.

**Functional Requirements**:

- **FR-012**: The TUI MUST display a session dashboard (active provider/model, token usage, turn count) at the top of the prompt.
- **FR-013**: The dashboard MUST reflect the current session's provider/model and metrics (not a stale snapshot).

---

### User Story 4 - Coexist with the plain reader; preserve the non-TTY fallback (Priority: P2)

As an operator, I want the new TUI to be opt-in — a bare terminal invocation keeps the round-012 plain reader, and any non-terminal stdin keeps the existing plain/piped behavior — so upgrading `tellme` never changes my existing workflows.

**Why this priority**: it protects existing workflows and keeps the change trustworthy; it depends on Story 1 (there is a new surface only because of Story 1).

**Independent verification**: assert that a bare terminal invocation with no `-i` uses the round-012 plain reader (the TUI does not engage), that a non-terminal stdin never renders the TUI, and that rounds 001–014 acceptance remains green.

**Acceptance Scenarios**:

1. **Given** a bare `tellme` on a terminal with no prompt and no `-i`, **When** the interactive reader engages, **Then** it is the round-012 plain reader — the TUI does not engage.
2. **Given** `-i` is set but stdin is **not** a terminal, **When** the run starts, **Then** the TUI does not render and the existing plain/piped/non-interactive behavior applies.
3. **Given** any of the rounds 001–014 paths (positional prompt, piped stdin, `-l`, `-d`, `--version`, `--new`), **When** it runs, **Then** its behavior is unchanged.

**Functional Requirements**:

- **FR-014**: The TUI MUST engage **only** when `-i` (or `USE_TUI_PROMPT`) is set **and** stdin is a terminal; otherwise the existing round-012 plain reader / piped / non-interactive paths MUST apply unchanged.

---

### Edge Cases

- **`-i` with a positional prompt**: the argument is the prompt; the TUI does **not** engage for that run (consistent with the positional-prompt precedence across rounds 001–014).
- **`-i` with a piped (non-terminal) stdin**: no TUI rendering — the existing piped import/combination path applies (NFR-001/FR-014).
- **Empty or aborted TUI submission** (immediate submit with no content, or `Esc`/`Ctrl+C`): **no** provider request, **no** append to the shared log.
- **Missing shared log**: suggestions still work from the active session; reading a non-existent file yields nothing (not an error surfaced to the operator).
- **Shared log written concurrently** (another persona/mode appends while tellme does): tellme's append is atomic (`O_APPEND`) and MUST NOT corrupt or truncate others' lines; the known **no `flock`** carried item still applies (a full cross-process lock is out of scope).
- **Noisy directories**: `.git`, `node_modules`, and similar MUST NOT appear in path suggestions.
- **Large shared log**: reads are newest-first and deduplicated, bounded by the suggestion cap; a compaction policy is a `/axb-data-plan` determination (Assumption A5).
- **Abort keybinding**: `Esc`/`Ctrl+C` aborts the prompt and exits without a request (consistent with round-012 cancel semantics).

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-015**: The round MUST NOT regress rounds 001–014 — every prior acceptance scenario MUST remain green, `stdout` MUST remain byte-exact (the TUI renders to the terminal, not into a piped `stdout`), and the frozen class-phrase vocabulary MUST be unchanged. No new class phrase is introduced.

#### Non-Functional Requirements

- **NFR-005**: All new assertions MUST be deterministic (no `time.Sleep`); the injected terminal seam, a scripted keystroke/input source, and the injected suggestion source are the verification surface — the E2E path MUST NOT require a real pty.

### Key Entities *(include if feature involves data)*

- **Interactive TUI prompt**: the rich terminal prompt surface — a multi-line editor, a live suggestion list, and a session dashboard. Present only when `-i`/`USE_TUI_PROMPT` is set and stdin is a terminal.
- **Suggestion**: one candidate completion offered by the suggestion engine — a recent prompt (from the shared log or the active session), a workspace path, or a registered tool name.
- **Session dashboard**: the live header of the prompt showing the active provider/model, the token usage (used vs budget), and the turn count.
- **Global prompt log**: the shared, append-only file `$TELL_ME_HOME/output/global_prompts.jsonl` holding one `{"timestamp","prompt"}` record per line; shared across personas/modes with `tell-me-go`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of `-i` interactive submissions perform exactly one reasoning turn on the trimmed composed text (no request on empty/aborted submissions).
- **SC-002**: 100% of `-i` submissions append exactly one line in the identical shape to the shared global prompt log; 100% of non-`-i` runs append none.
- **SC-003**: A prompt appended by `tellme` round-trips — the identical reader reads it back and it can appear in suggestions.
- **SC-004**: 100% of `-i` runs on a non-terminal stdin render no TUI and take the existing fallback path.
- **SC-005**: Every prior acceptance scenario (rounds 001–014) remains green, and the frozen class-phrase vocabulary is unchanged.
- **SC-006**: The interactive prompt behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.
- **SC-007**: The shared log is appended without corrupting or truncating concurrent lines (append-only), verified by a read-back after a concurrent append.

## Assumptions

- **A1 (record semantics)**: the shared global prompt log is **read + written**, and recorded **only under `-i`** (Q1 = Option 1); non-`-i` runs never write it.
- **A2 (relationship)**: the TUI **coexists** with the round-012 plain reader — the default interactive surface is unchanged, and `-i`/`USE_TUI_PROMPT` opts into the TUI (Q2 = Option 1).
- **A3 (platform)**: the TUI targets **POSIX** terminals only this round; there is no Windows keybinding/build variant (Q3 = Option 1).
- **A4 (tool suggestions)**: the tool-suggestion source is `tellme`'s own registered tool registry (the tool names it exposes); the exact projection is a research/DSL detail.
- **A5 (shared-log policy)**: the log is read newest-first and deduplicated, bounded by the suggestion cap; a **compaction policy** (mirroring the reference's size/unique-entry thresholds or a simpler policy) is a `/axb-data-plan` determination, and the known **no `flock`** carried item remains out of scope.
- **A6 (dependency)**: the round introduces `tellme`'s **first TUI-grade presentation dependency** (as the reference does) — a `/axb-technical-research` decision feeding the Setup phase; the non-TTY fallback keeps the no-TUI path dependency-light.
- **A7 (unchanged)**: rounds 001–014 semantics are unchanged except the new prompt surface — single-turn execution, stdin piping, rendered/raw output, durable history, the agent tool loop, the payload status line, cross-stream ordering, persona-on-the-wire, the round-012 plain reader, and the Vertex/Gemini transport. No streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, or `SafePath`/consent. The shared global prompt log is a **new** data artifact, distinct from the per-session `history.jsonl`.
