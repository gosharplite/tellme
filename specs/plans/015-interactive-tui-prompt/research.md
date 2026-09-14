# Phase 0 Research: tellme Interactive TUI Prompt — Suggestions, Session Dashboard, and the Shared Global Prompt Log (Round 015)

Topic: add the **`-i` / `--interactive` Interactive TUI Prompt** to `tellme` — a rich terminal prompt with a **live suggestion engine** (recent prompts + workspace paths + registered tools), a **session dashboard** (provider/model · tokens · turns), a **multi-line editor**, and terminal **keybindings** — interoperable with the shared Niffler-env prompt log at `$TELL_ME_HOME/output/global_prompts.jsonl`.

Today `tellme` has only the round-012 **plain** interactive multi-line reader (`Ctrl+D`, hint on `stderr`, real isatty via `x/term`). The rich prompt is a new, opt-in surface on top of it, and the shared prompt log is a **new data artifact** (distinct from the per-session `history.jsonl`).

Scope note: the language (Go 1.26), module, CLI/config layers, the `llm.Gateway` port, both adapters, the session-history store (rounds 007–014), the agent tool loop, and the rest of the pipeline were locked in rounds 001–014. The system still has **one CLI end**; the round reaches **no new endpoint**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; `godog` running the built binary; E2E acceptance + pure-helper units) and are **not re-decided**. Clarify Round 1 (in `spec.md`) locked: **Q1** = read + write the shared log, recorded **only under `-i`**; **Q2** = the TUI **coexists** with the round-012 plain reader; **Q3** = **POSIX-only**. **IN**: the `-i` TUI prompt, the suggestion engine, the dashboard, the shared prompt log, and the opt-in/non-TTY guarantees. **OUT**: the round-012 reader itself (unchanged), streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent.

---

## Decision 1: Adopt the Bubble Tea family as tellme's TUI layer

- **Decision**: Build the interactive prompt on `github.com/charmbracelet/bubbletea` (the Elm-style TUI runtime) with the `github.com/charmbracelet/bubbles/textarea` multi-line editor component, using `lipgloss` for styling. `lipgloss` is **already in the module graph** transitively via `glamour` (round 006), so it is promoted from indirect to direct; `bubbletea` and `bubbles` are the new modules. This is `tellme`'s **first TUI dependency**.
- **Rationale**: A rich terminal UI (focusable panes, a live suggestion list with a selection cursor, a redraw loop, an overlay) is exactly what Bubble Tea provides and what is impractical to hand-roll with stdlib alone within a disciplined round. It is the **same family** the reference uses, so behaviour parity is straightforward, and it shares the `charmbracelet` lineage with the existing `glamour` renderer (consistent dependency family, `lipgloss` already present). The non-TTY fallback keeps the dependency off the piped/boot paths (it is only imported by the interactive branch).
- **Alternatives considered**:
  - **Hand-rolled raw-terminal TUI (stdlib `golang.org/x/term` raw mode + ANSI)** — zero new modules, but re-implements cursor/redraw/key decoding; high risk and no parity with the reference's keybindings — rejected.
  - **A different TUI library (`tcell`, `gocui`, `termenv`-only)** — `tcell` is a lower-level cell engine; `gocui` is view-based and less idiomatic for this model; neither matches the reference's behaviour — rejected.

## Decision 2: The suggestion engine is a multi-source, subsequence-matched, debounced aggregator

- **Decision**: One suggestion engine aggregates three sources — (a) recent prompts (the shared global log seeded newest-first + the active session's prompts), (b) workspace paths (only when the query is path-like — contains a separator or starts with `.`), and (c) registered tool names from `tellme`'s own tool registry. Matching is **fuzzy subsequence** (`IsSubsequence`), results are **deduplicated** and capped at **10**; a query change triggers a **~100 ms debounce** before recomputation; an empty query returns the newest recent prompts. Path scanning **excludes noisy directories** (`.git`, `node_modules`, and similar) via a workspace-ignore policy and yields to cancellation between directory batches.
- **Rationale**: This mirrors the reference's `multiSourceSuggestionService` exactly, which is the behaviour the round is re-creating; the debounce keeps suggestion computation off the visible-input path (`NFR-002`), and the cap + ignore policy bound cost (`NFR-002`). Subsequence matching (not prefix) matches the reference's feel.
- **Alternatives considered**:
  - **Exact/prefix match only** — cheaper but does not match the reference's fuzzy feel — rejected.
  - **No debounce (recompute on every keystroke)** — stalls the UI on large directories — rejected.

## Decision 3: The shared global prompt log is an append-only JSONL store, written only under `-i`

- **Decision**: The shared prompt log is an **append-only JSON-Lines file** at `$TELL_ME_HOME/output/global_prompts.jsonl` (the `output/` **root**, shared across modes/personas — not `output/<mode>/`), one record per line: `{"timestamp":"<RFC3339>","prompt":"<text>"}`, both fields strings. Reads are **newest-first and deduplicated** (reverse chunked scan), bounded by the suggestion cap. Writes use **`O_APPEND|O_CREATE`** (atomic append; never truncate a file other personas also write) and are **detached/bounded** so they never block the prompt (`NFR-003`/`NFR-004`). Per Clarify Q1, the write happens **only for an `-i` interactive submission**; every non-`-i` run writes nothing. **Compaction mirrors the reference** (a size threshold ≈150 KiB triggers an async dedupe pass keeping ≤1200 unique entries, newest-first), performed best-effort behind the append lock. The exact DBML column/lifecycle is the **data owner's** (`/axb-data-plan`) determination.
- **Rationale**: Byte-for-byte interop with `tell-me-go` requires the identical record shape and the identical file location; `O_APPEND` is the only safe write against a file shared by concurrent personas (the known **no `flock`** item stays out of scope, but append-only avoids corruption/truncation). Recording only under `-i` mirrors the reference and keeps rounds 001–014 byte-stable (`SC-002`). Mirroring the reference's compaction keeps the file from growing without bound while preserving round-trip lines.
- **Alternatives considered**:
  - **Record always (every prompt)** — Clarify Q1 Option 2; would touch the shared file on piped/positional runs and change rounds 001–014 behaviour — rejected.
  - **Read-only (never write)** — Clarify Q1 Option 3; loses the round-trip value and the reference parity — rejected.
  - **A per-mode log, or a database** — breaks the shared-across-modes requirement; the file is intentionally plain JSONL — rejected.

## Decision 4: The TUI is opt-in; the round-012 reader stays the default; the non-TTY fallback is preserved

- **Decision**: The TUI engages **only** when `-i`/`--interactive` (or `USE_TUI_PROMPT: true`) is set **and** stdin is a terminal (reusing the round-012 `golang.org/x/term` real-isatty seam, behind the injected probe). Otherwise the existing behaviour is unchanged: a bare terminal invocation keeps the round-012 plain reader; piped/non-terminal input keeps the piped/non-interactive path; a positional prompt is used directly; `--version`/`-d`/`-l`/prompt-less-`--new` are untouched.
- **Rationale**: Clarify Q2 = Option 1 (coexist) + `NFR-001`/`FR-014`; opt-in preserves every existing workflow and mirrors the reference's default-off flag. Reusing the existing isatty seam keeps the gating consistent with round 012 (ADR 0003) and avoids a second terminal probe.
- **Alternatives considered**:
  - **Supersede the plain reader (TUI is the terminal default)** — Clarify Q2 Option 2; changes the default surface for every operator — rejected.
  - **A new `-tui`-style flag distinct from `-i`** — diverges from the reference's flag name — rejected.

## Decision 5: The dashboard reuses the session's existing provider/model, token, and turn state

- **Decision**: The session dashboard shows the active **provider/model** (the resolved provider the turn uses), the **token usage vs the configured budget** (the same payload estimate/actual the round-009 status line already computes against `MAX_HISTORY_TOKENS`), and the **turn count** (from the session history). These are read from the existing session/metrics state and injected into the TUI model — the dashboard adds **no** new accounting.
- **Rationale**: `tellme` already knows the active provider/model, the per-turn token figures (round 009), and the persisted history length (round 007); re-presenting them satisfies `FR-012`/`FR-013` with no new source of truth and keeps the dashboard deterministic and fake-injectable for tests.
- **Alternatives considered**:
  - **A new live token counter in the TUI** — duplicates the round-009 accounting and risks divergence — rejected.
  - **Show only the provider (no tokens/turns)** — under-delivers the spec's dashboard (`FR-012`) — rejected.

## Decision 6: Hermetic verification via injected input/output and a fake suggestion source (no pty)

- **Decision**: The TUI model takes an **injected input reader, output writer, and suggestion source**, so it runs under `godog`/unit tests **without a real pty** (a scripted key sequence drives it; the fake suggestion source returns canned suggestions). Verification is at **both layers**: (a) fast **unit** tests drive the Bubble Tea model with scripted keys (suggestion seed/refresh/accept, submit vs newline vs abort, dashboard fields); (b) **E2E** acceptance drives the built binary through the round-012 `TELL_ME_FORCE_STDIN_TTY` seam plus a scripted stdin, asserting the observable outcomes — the shared-log append (under `-i` only) and the single reasoning turn (the fake provider's recorded request). Real-pty stream fidelity remains a **named pin** (rounds 005/006/012 precedent).
- **Rationale**: `NFR-005` requires determinism and no real pty; the project already forswore a pty harness (round 005 grill Q6; round 006 Q2). Injecting the I/O + suggestion source makes the model unit-testable exactly as the reference's `model_test.go` does, while the E2E layer pins the CLI wiring (the append + the turn) through the existing seam.
- **Alternatives considered**:
  - **A pty-based E2E harness** — heavy, platform-fragile, and previously foreclosed — rejected (named pin).
  - **Assert by exact rendered ANSI bytes** — `TERM`/profile-dependent (round 006 precedent) — rejected in favour of behavioural assertions.

## Decision 7: Dependency footprint — first TUI modules; no others

- **Decision**: The round adds `github.com/charmbracelet/bubbletea` and `github.com/charmbracelet/bubbles` and **promotes `lipgloss` to direct**; no other new module. `go.mod` changes; `go.sum` gains the new modules' checksums. The suggestion engine, the shared-log store, and the dashboard are hand-written over stdlib + the existing layers.
- **Rationale**: All UI machinery lives in the Bubble Tea family; matching is a hand-written subsequence check; the store is `encoding/json` over `os`. This keeps the dependency growth minimal and confined to the interactive branch.
- **Alternatives considered**:
  - **Add a fuzzy-match library for suggestions** — the subsequence check is ~10 lines; a module is unnecessary — rejected.
  - **Add a file-watcher/glob library for path suggestions** — stdlib `os.ReadDir` + an ignore policy suffices — rejected.

## Decision 8: POSIX-only; no Windows variant

- **Decision**: The TUI targets **POSIX** terminals this round (keybindings `Tab`/`Shift+Tab`, `Ctrl+S`/`Alt+Enter`, `Esc`, `Ctrl+C`); there is **no** Windows-specific keybinding branch or build tag, matching the round-012 reader.
- **Rationale**: Clarify Q3 = Option 1; the reference's keybindings are POSIX-oriented and round 012 already established the POSIX-only precedent, keeping the scope and the test surface bounded.
- **Alternatives considered**:
  - **POSIX + Windows** — an extra keybinding/build branch and a wider test surface for no round-015 requirement — rejected.

## Decision 9: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog` running the built binary; strategy = E2E acceptance + fast pure-helper units — unchanged this round, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round adds an opt-in prompt surface to the single CLI end; it introduces no new interface, service, runner, or second end, and does not change the runner or the strategy.
- **Alternatives considered**:
  - **Re-open the techstack/test questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **TUI dependency growth (Decision 1/7)**: `bubbletea`/`bubbles` are new modules (the project's first TUI family). If a future slice wants a leaner dependency set, the TUI is confined to the interactive branch and can be swapped behind the same model seam.
- **Shared-log shape + lifecycle (Decision 3)**: the record shape (`{timestamp, prompt}`), the file location, the ordering/dedupe, and the compaction policy (≈150 KiB / ≤1200 unique) are proposed here; the **exact DBML column names/types and the lifecycle/retention wording are the data owner's** (`/axb-data-plan` MODIFY) determination — held open.
- **No `flock` (Decision 3)**: append-only avoids corruption, but there is no cross-process lock; concurrent compaction by two personas could race (best-effort, last-writer-wins). A `ModeLocker`/`flock` remains a carried forward item.
- **Real-pty fidelity (Decision 6)**: stream fidelity of a real TTY is not asserted (a named pin); the round proves behaviour through injected I/O + the `TELL_ME_FORCE_STDIN_TTY` seam.
- **Dashboard semantics (Decision 5)**: the dashboard reuses the round-009 estimate/actual; the **estimator-ignores-replayed-arguments** forward item (round 011, N-2) still applies to the displayed figure.
- **Compaction threshold tuning (Decision 3)**: the ≈150 KiB threshold is adopted for parity; if real usage shows churn, it is a tuning knob (a forward item).
- **Deferred, still out of scope**: streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent, the Google Gemini API family (inline key), Application Default Credentials, and a Windows variant.
