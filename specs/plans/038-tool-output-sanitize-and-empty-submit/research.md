# Technical Research: tool-output sanitization + empty-submit no-op (round 038)

**Plan Package**: `specs/plans/038-tool-output-sanitize-and-empty-submit`

**Status**: Decisions D1–D7 (ratifies the operator Q1–Q3 locked in `spec.md`).

## Context

Two folded operator issues, both on the `stderr`-bound interactive surface:

- **#78** — the live `[Tool Output]` block (`internal/ui/tooloutput.go`, fed by `internal/infrastructure/tools/command.go`'s `teeSink`) writes a shell command's **raw** bytes as `[HH:MM:SS] [Tool Output] <line>` lines to `env.stderr`. A command that sets a colour/attribute and never resets (or is killed mid-output) leaves the terminal in that state, tinting everything printed afterwards.
- **#76** — `internal/ui/tui/prompt/model.go`'s `handleKey` sets `submitted = TrimSpace(value) != ""` but **always** returns `tea.Quit`; `Run` → `("", false, nil)` → the CLI returns `Success`. So an empty `Ctrl+S` ends the session (exit 0), where the reference's `submit()` returns false and the model keeps running.

## Decisions

### D1 — Sanitize at the `internal/ui` presentation seam (presentation-only)

The control-sequence removal lives in the pure `ui` layer, applied to the **assembled** line inside `FormatToolOutputLine` (the same function that adds the `[HH:MM:SS] [Tool Output] ` framing). Reasons:

- **Presentation-only (FR-002).** The tool **result** fed back to the model is a *different* consumer of the same child bytes (`boundedBuffer` → the loop's result), so sanitizing the `ui` formatter cannot touch it. The result keeps its raw bytes.
- **Assembled-line scope.** The writer already buffers until a `\n`, so a sequence split across two `Write` calls is assembled before sanitization (edge case covered).
- **Testability.** A pure formatter is unit-pinned over a hostile fixture with no pty and no process.

Rejected: sanitizing in `command.go` (would be near the result path and would have to duplicate the line-assembly the writer owns); sanitizing in the CLI wiring (not a pure function; harder to pin).

### D2 — The removed class: all ESC-introduced sequences + stray C0/DEL, UTF-8-safe

`sanitizeControl(s)` scans bytes and drops:

- every **ESC-introduced** sequence: CSI (`ESC [ … final 0x40–0x7E`), OSC (`ESC ] … BEL | ESC \`), and any other `ESC x` two-byte sequence (e.g. `ESC ( B`, `ESC 7`);
- every stray **C0** control byte (`< 0x20`) except `TAB` (`0x09`);
- **DEL** (`0x7f`).

It **never** drops bytes `≥ 0x80`, so multibyte UTF-8 text is preserved intact (only ASCII control space is touched). This covers the reported colour leak and the sibling cursor-hiding / OSC-title classes in one rule, and cannot strand a terminal state because no escape survives to the stream.

Rejected: SGR-only stripping (leaves `\x1b[?25l`/OSC to leak); a full terminal-emulator library (a dependency for a one-rule need); regex (hand-rolled scanner is deterministic, dependency-free, and avoids backtracking concerns).

### D3 — Neutral-close restore (belt-and-braces, FR-005)

Independently of D2, the writer emits a single **default-state restore** (`\x1b[0m`, a new `ToolOutputReset` constant) immediately before the **closing separator** in `End()`. `End()` is already reached on every exit path (`runCaptured` calls it after `cmd.Wait()`, and on a `cmd.Start()` failure), so the normal, start-failure, trim (byte-budget) and timeout paths all get the restore. This covers the residuals D2 cannot: a well-behaved child's reset landing in the deliberately-**dropped** trailing partial line, and a child **killed mid-output** before it could reset.

The restore is written by the writer (which owns the block) and is **not** on a `[Tool Output]` content line, so the D2-only "content lines carry no control sequence" predicate stays clean.

### D4 — Empty submit is a no-op (mirror the reference)

`handleKey`'s `Ctrl+S` and `Alt+Enter` branches consult a new `trySubmit()` helper: it returns `tea.Quit` **only** when the trimmed editor value is non-empty; otherwise it returns `nil` (the model keeps running), mirroring the reference's `submit()` → `handleSubmissionKeys` → `(m, nil)`. `Esc`/`Ctrl+C` abort is unchanged (`tea.Quit`), and a non-empty submit is behaviourally identical. The `Model.submitted` flag is therefore only ever set on a real submit, so `Run`'s `!fm.WasSubmitted()` early return now only happens on abort — which the CLI still maps to `Success` (unchanged).

### D5 — Witnesses (falsifiability)

| Layer | Witness |
| --- | --- |
| Unit (`internal/ui`) | `TestFormatToolOutputLine_Sanitizes` (hostile fixture: SGR set/no-reset, SGR set+reset preserving text, OSC title, cursor-hide, stray BEL, split-across-writes via the writer, tab preserved, UTF-8 preserved); `TestToolOutputWriter_EndRestoresNeutralState` (a restore precedes the closing separator on the normal and the never-Begun/`End`-only paths). |
| Unit (`internal/ui/tui/prompt`) | `TestModelEmptySubmitDoesNotQuit` (the `Update` command is nil for an empty `Ctrl+S`/`Alt+Enter`); the existing `TestModelEmptySubmitKeepsFrame` still holds; `TestModelNonEmptySubmitQuits`. |
| E2E (`tests/e2e`) | `the run streamed the command's output free of terminal control sequences` (content lines carry no ESC/control byte); `the terminal is left in its default state` (a restore follows the last content line); the empty-then-real-submit Example (exactly one request ⇒ the empty submit did not quit). |

Falsifiability: removing `sanitizeControl` (identity) fails the sanitize pin + E2E; restoring the unconditional `tea.Quit` fails the empty-submit pin + E2E; removing the `End()` restore fails the neutral-state pin + E2E.

### D6 — Recorded divergence from the reference

`tell-me-go` has **no** output sanitizer — it forwards command control bytes to its diagnostic stream too. Round 038 deliberately **diverges** (it fixes a real terminal-state leak present in both). This is a recorded divergence, not a silent parity break: the round's parity intent applies only to the **empty-submit** behaviour (#76), which *is* reference-exact.

### D7 — Scope guard

Unchanged: the `[Tool Output]` header/separator literals, the block's bound/stop semantics and spinner yield, `output_file` handling, the tool **result**, `stdout`, `-r`, flags, exit codes, the class-phrase vocabulary, the provider transport, and every persisted record. Stdlib-only; POSIX-only; no new dependency.

## Truth impact (ratified)

- `specs/truth/techstack.md` — **MODIFY** two rows (Agent command tool `execute_command`; Interactive TUI prompt `-i`).
- `specs/truth/features/cli/chat/**` + `chat/dsl.md` — **MODIFY** (two new Rules + their `DSLRow`s).
- `/axb-api-plan` — **NOOP** (no HTTP surface). `/axb-data-plan` — checked **NOOP** (no persisted state).

## Residual risks (forward)

- The scanner is deliberately **conservative**: an exotic multi-byte escape it does not model is consumed to the end of the sequence by the CSI/OSC scanners, or to two bytes for a generic `ESC x`; an unmapped case could in principle leave a visible artefact — acceptable for a presentation line, and the D3 restore bounds the terminal-state risk.
- D2 does not reinterpret the text (no "strip then re-colour"); a command's output is shown as plain text in the `[Tool Output]` block. A future "preserve safe styling" is out of scope.
