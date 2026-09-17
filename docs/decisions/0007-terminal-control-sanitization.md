# ADR 0007 — Terminal-control sanitization of the `[Tool Output]` block: a new stderr presentation invariant, and a deliberate reference divergence

- **Status:** Superseded by [ADR 0008](0008-terminal-safe-lines-and-blank-line-grouping.md)
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** issues [#78](https://github.com/gosharplite/tellme/issues/78) (the leak this ADR fixes) and [#76](https://github.com/gosharplite/tellme/issues/76) (the co-delivered `-i` empty-submit parity fix);
  round 034 (`specs/plans/034-tool-call-log-parity` — introduced the live `[Tool Output]` block and the per-line `[HH:MM:SS] [Tool Output] <line>` rendering);
  ADR 0005 (tool-log rendering policy + the rune-cap divergence family);
  ADR 0006 (the sibling sanitize+cap family for the `[Tool Reason]` value);
  the round-019 `isatty(stderr) && !-r` diagnostic gate (`newTurnSpinner`, ADR 0003 lineage);
  round 038 (`specs/plans/038-tool-output-sanitize-and-empty-submit` — this ADR's round)

## Context

Round 034 added a **live `[Tool Output]` block**: while a shell-class `execute_command` runs, the child's stdout/stderr are teed (`command.go` `teeSink` → `io.MultiWriter(buf, sink.Writer)`) into a per-line writer (`internal/ui/tooloutput.go`) that renders `[HH:MM:SS] [Tool Output] <line>` to the **shared diagnostic stream `stderr`**.

Those bytes were forwarded **verbatim**. So a command that sets a terminal attribute and never resets it — `ls --color=always`, `grep --color=always`, an `npm`/`npx` child, or a command killed mid-output at the byte budget or its `timeout` — left the terminal in that state, tinting the rest of the run (the closing separator, the following `[Tool Result]`/reason/metrics/`Ready` lines, and the glamour-rendered answer). Issue [#78](https://github.com/gosharplite/tellme/issues/78) reported it; the reference (`tell-me-go`) has **no** output sanitizer at all, so it has the same leak.

Two project-level items need a citable home that a future round (which reads issues + ADRs, not frozen plan packages) can depend on:

1. **A new invariant over the stderr diagnostic surface** — the streamed `[Tool Output]` content is control-free, and the block closes neutral.
2. **A deliberate reference divergence** — tellme sanitizes; the reference does not.

## Decision

**D1 — Sanitization lives at the `internal/ui` presentation seam, on the assembled line.** `sanitizeControl` is applied inside the pure `FormatToolOutputLine`, i.e. *after* the writer's line assembly (so a sequence split across two `Write` calls is still removed) and purely in the presentation layer. The tool **result** fed to the model is a *separate* consumer of the same child bytes — the other arm of `io.MultiWriter` → `boundedBuffer` — so sanitization **cannot** reach it. That asymmetry is the point: the model keeps raw bytes; the human-facing line is safe.

**D2 — The removed class is precisely defined (7-bit control space).**

Removed:
- every **7-bit** escape sequence introduced by ESC (`0x1b`): CSI (ESC `[` … final `0x40–0x7E`), OSC (ESC `]` … BEL `0x07` | ST `ESC \`), and a generic ESC + optional intermediates (`0x20–0x2F`) + **one ASCII final byte**;
- every stray C0 control byte (`0x00–0x1F`) **except TAB** (`0x09`) — this includes an interior **CR** (`0x0D`), so in-place `\r` progress output concatenates rather than repainting;
- DEL (`0x7f`).

**Out of scope (untouched, recorded):** the **8-bit C1** control range (`0x80–0x9F`) — e.g. `0x9B` as the 8-bit CSI form passes through — and every other byte `≥ 0x80`.

Two correctness constraints that are part of the class definition:
- **The ESC consumption is ASCII-gated**, so ESC followed by a multi-byte rune drops only the ESC and leaves the rune intact; the sanitizer therefore **never introduces** invalid UTF-8 (`"\x1b日本"` → `"日本"`). This was review BLOCKER **B1** (the first cut consumed one arbitrary byte and decapitated the rune). The sanitizer only *removes* bytes and forwards the rest verbatim, so a genuinely binary source can still arrive as invalid UTF-8 (C1 `0x9B`, raw continuation bytes, truncated runes pass through) — the guarantee is *never introduces*, not *always valid*.
- **Consumption is bounded, per kind** (`csiScanLimit = 128`, `oscScanLimit = 1024`; review **TD-3** fixed an over-narrow first cut that used a single 64-byte window). A sequence whose terminator lies **within** its window is removed **in full**, however long — an OSC-8 hyperlink URL is `len(url)+7` bytes and is legitimately long, as is a window title, so the OSC window is deliberately wide; a sequence whose terminator lies **beyond** the window, or an **unterminated** one, drops only the ESC and continues, so a mangled/binary fragment cannot swallow a whole line's visible text (review **RF-2**). The generic path (`genericEscLen`) reuses the **CSI** window to bound its intermediate run (a shared constant, commented at the function) — its final byte is a single ASCII byte, so the run is the only unbounded part.

**D3 — The block always closes in a neutral state.** Independently of D2, `End()` writes a single default-state restore (`\x1b[0m`) before the closing separator, on **every** close path (normal, `cmd.Start` failure, byte-budget trim, timeout). `FR-005` therefore holds by construction — all three paths reach `sink.End()` and `End()` is idempotent + a no-op when never `Begin`ed — and it covers the two residuals D2 cannot: a well-behaved child's reset landing in the deliberately-**dropped** trailing partial line, and a child **killed mid-output**.

**D4 — The restore is written unconditionally, including on a non-terminal `stderr`** (review **TD-2**). The alternative — gating the restore on the round-019 `isatty(stderr) && !-r` seam — was rejected: it would make the invariant depend on *where the stream goes*, and it would force the E2E neutral-close Example to force the stderr-terminal seam (pulling the spinner onto the path). The chosen shape is a single unconditional write that is a **visual no-op on a non-terminal** (a CI log, `2>file`, a `-r` pipe); it adds four bytes to a *redirected* `stderr`. Recorded explicitly here and in the `techstack.md` row; `stdout` stays byte-exact.

**D5 — The `-i` empty-submit behaviour is aligned to the reference (issue #76), independently of the sanitize policy.** `Ctrl+S`/`Alt+Enter` on an empty (or whitespace-only) editor returns **no command** (the prompt stays open), mirroring the reference's `submit()` → `handleSubmissionKeys` → `(m, nil)`. This is *not* a divergence; the sanitize policy is. The two are co-delivered but independent invariants.

**D6 — Recorded divergence.** tell-me-go sanitizes nothing; tellme deliberately removes the control class from its `[Tool Output]` presentation and restores neutral on close. Divergences of this class are recorded in this ADR and in the `techstack.md` `execute_command` row (the ADR 0005 D6 precedent).

## Alternatives considered

1. **Sanitize at the command tool (`command.go`)** — rejected: it is upstream of the line assembly the writer owns, and it risks reaching the model-facing result arm.
2. **SGR-only stripping** — rejected: leaves cursor-hiding (`\x1b[?25l`) and OSC titles to leak, and cannot guarantee a neutral terminal.
3. **Reset-on-close only (no stripping)** — rejected: fixes the tint but leaves the control bytes in the stream and does not touch the sibling classes.
4. **Gate the neutral restore on `isatty(stderr) && !-r`** — rejected (see D4).
5. **Sanitize a *wider* surface now** (the sibling `[Tool Reason]` / `[Tool Result]` / `[Tool Action]` values) — rejected **for this round** as scope creep; homed on a live issue (see Consequences).
6. **A terminal-emulator library** — rejected: a dependency for a one-rule need; the hand-rolled scanner is deterministic and dependency-free.

## Consequences

- The streamed `[Tool Output]` **content** lines are control-free and **never introduce** invalid UTF-8; the block always closes neutral; the tool **result** fed to the model is unchanged (raw). `stdout`, `-r`, flags, exit codes, the frozen class-phrase vocabulary, the `[Tool Output]` header/separator literals, the block's bound/stop/spinner-yield semantics and every persisted record are unchanged.
- **Scope boundary recorded (review TD-1):** the leak class is closed on the `[Tool Output]` surface **only**. Sibling `stderr` formatters render text of the same provenance class and remain unsanitized — `FormatToolReason` (model-authored), `FormatToolResult` (a tool-result snippet, e.g. a `read_files` of an ESC-bearing log/binary or a `list_files` of an escape-bearing filename), and the `[Tool Action]` argument values. The scalable shape is a single-owned **“terminal-safe line” policy** in `internal/ui` applied by **every** `[Tool …]` formatter; that is homed on live issue **[#80](https://github.com/gosharplite/tellme/issues/80)** (round-035 G3 lesson: a durable surface, not a frozen plan package). This ADR's invariant applies to the `[Tool Output]` content lines, not to the stream as a whole (the block's own close restore is a deliberate escape on the same stream).
- **Recorded residual risk (review N-6/N-7):** the scanner is deliberately conservative and **window-bounded** — a sequence it does not model, or whose terminator lies **beyond its kind's window** (CSI 128 / OSC 1024), drops only the ESC and passes the **rest of the line** through as visible text. The window bounds how much is *removed*, not what is *printed*: the printed remainder is the rest of the line, so this residual is bounded by the **line** — ultimately by the block's byte budget (round 024), which bounds the whole `[Tool Output]` stream — and the D4 restore bounds the terminal-state risk.
- Immutable once `Accepted`; a future change to the removed class, the neutral-close policy, or the sanitize seam supersedes this ADR rather than editing it.
