# ADR 0008 — Terminal-safe `[Tool …]` line policy (generalized) + live turn-output blank-line grouping: an extended stderr presentation invariant, and two recorded reference divergences

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Supersedes:** [ADR 0007](0007-terminal-control-sanitization.md) (its recorded scope boundary — "the leak class is closed on the `[Tool Output]` surface **only**" — is closed by this ADR)
- **Related:** issue [#80](https://github.com/gosharplite/tellme/issues/80) (the generalized policy this ADR records);
  round 038 / PR [#79](https://github.com/gosharplite/tellme/pull/79) (the `[Tool Output]` sanitizer + neutral close this ADR generalizes);
  ADR 0007 (the superseded scope boundary), ADR 0006 (the reason fold/cap this ADR extends with sanitization), ADR 0005 (the tool-log rendering policy + the cap divergence family);
  issue [#69](https://github.com/gosharplite/tellme/issues/69) (the single-ownership theme this policy joins); issue [#72](https://github.com/gosharplite/tellme/issues/72) / round 035 (the spinner-yield policy on the same tail);
  round 039 (`specs/plans/039-terminal-safe-lines-and-turn-spacing` — this ADR's round)

## Context

Round 038 (ADR 0007) removed terminal control data from **one** `stderr` surface — the live `[Tool Output]` block — and explicitly recorded a scope boundary: the leak class was closed on `[Tool Output]` **only**, with the sibling formatters (`FormatToolReason`, `FormatToolResult`, `FormatToolAction`) still rendering raw model-authored / externally-sourced text. Issue [#80](https://github.com/gosharplite/tellme/issues/80) is the durable home for that boundary; round 039 closes it.

Separately, an operator request asked for the live turn output to be **grouped by blank lines** so a tool-using turn is scannable rather than a dense wall.

Two project-level items need a citable home:

1. **An extended invariant** — the control-sanitization policy is **single-owned** and applies to **every** `[Tool …]` formatter, and the live turn output has a defined **blank-line grouping**.
2. **Two recorded reference divergences** — the reference (`tell-me-go`) sanitizes nothing, and it does not space these lines.

ADR 0007 recorded its scope boundary as a consequence and stated it is superseded (not edited) by a future change to the removed class / neutral-close policy / sanitize seam. Generalizing the policy's scope revises that recorded consequence, so this ADR **supersedes** ADR 0007 rather than editing it (the ADR 0005 → 0006 precedent).

## Decision

**D1 — One single-owned terminal-safe-line policy, consumed by every `[Tool …]` formatter.** `sanitizeControl` and its helpers (`isControlRune`, `escSequenceLen`, `csiLen`, `oscLen`, `genericEscLen`, the per-kind scan windows) move into a dedicated `internal/ui/sanitize.go` — **one** definition of the class — consumed by `FormatToolOutputLine`, `FormatToolReason`, `FormatToolResult`, and `FormatToolAction` (argument **keys and values**). `FormatToolEngine` renders only literals + integers and is a non-consumer by nature. Rationale: the class is defined once (the [#69](https://github.com/gosharplite/tellme/issues/69) single-ownership theme), the `[Tool Output]` behavior stays **byte-identical** after the move, and the policy is separable from any one surface.

**D2 — The removed class and the application order are unchanged.** The class is exactly ADR 0007 D2 (every **7-bit** ESC-introduced sequence — CSI/SGR, OSC, other; every stray C0/DEL; TAB kept; interior CR dropped; bytes `≥ 0x80` untouched; ASCII-gated ESC consumption; per-kind bounded window; **never introduces** invalid UTF-8). For the reason and result the order is **fold + trim → sanitize → rune cap**, so the existing rune caps (`reasonValueCap`/`resultValueCap` = 200, `argValueCap` = 189) bound the **visible** output and the one-line/cap contract of round 036 / ADR 0006 is preserved. Argument **keys** are sanitized and remain uncapped; the composed list otherwise keeps its shape (sorted ascending, `reason` excluded).

**D3 — The block's neutral-close restore stays `[Tool Output]`-scoped.** No per-line reset is added to the single-line formatters. The round-038 restore exists for the `[Tool Output]` **block** (a multi-line, byte-budgeted, killable stream); a reset line per tool-log row would be new visible noise, not a fix, and those formatters open no block state. This is a deliberate **non**-change, recorded so it is not silently re-litigated (the ADR 0007 D4 unconditional-restore rationale still holds for the block).

**D4 — The blank-reason suppression reads the sanitized value.** The round-036 blank-reason guard (`agentloop.logAction`, `agentloop.reasonsOf`, and the defensive `callRenderer.OnCallEnd` site) is widened to test the **folded + sanitized + trimmed** value, so a reason that is *only* control data (non-blank before removal, blank after — e.g. `"\x1b[31m"`) emits **no** `[Tool Reason]` row. This extends the guard's **input**, not its three sites; consolidating the sites remains on [#69](https://github.com/gosharplite/tellme/issues/69).

**D5 — Live turn-output blank-line grouping (a defined presentation policy).**

| # | Site | Rule |
| --- | --- | --- |
| 1 | call begin | exactly one blank line before the call block's **first** line — the `[Tool Reason]` line, else the `[Tool Action]` line — **per call** (a `k`-call round emits `k` blanks) |
| 2 | post-call tail | exactly one blank line before the grouped `[Tool Reason]` block, and **no** blank line between its lines (only when the block is non-empty) |
| 3 | post-status | exactly one blank line before the measured `Payload:` line + metrics line + `Ready` footer (only when that group is written) |

The blanks are plain `\n` on the `stderr` diagnostic stream; nothing is added to a non-tool turn, `stdout`, a persisted record, or the offline paths.

**D6 — Recorded divergences (two).** tell-me-go **sanitizes nothing** in its tool log (its only `sanitizeForTerminal` is a LaTeX→Unicode helper) **and** writes its tool-log lines **without** these blanks (`internal/ui/renderer_metrics.go` `LogToolCall`/`LogToolResult`/`renderPostCallStatus`). tellme deliberately diverges on both counts: the sanitize closes a real terminal-affecting leak present in both; the spacing is an operator-requested readability choice. Divergences of this class are recorded here and in the `techstack.md` rows (the ADR 0005 D6 / ADR 0007 D6 precedent).

## Alternatives considered

1. **Leave `sanitizeControl` in `tooloutput.go` and import it into the siblings** — rejected: a `[Tool Output]`-named file owning a general policy is the wrong home (the issue's own "current vs. scalable" note).
2. **Sanitize in `internal/agent`** — rejected: it sits near the model-facing result and the stored steps; the rule belongs at the `internal/ui` presentation seam (presentation-only, FR-002).
3. **Add a per-line reset to the single-line formatters** — rejected (D3): noise, and no block state to restore.
4. **Sanitize *after* the rune cap** — rejected (D2): the cap would bound the wrong string, letting a removed-sequence value print "short".
5. **A round-level blank only (no per-call blank)** — rejected (D5 #1): the operator explicitly wants each call separated.
6. **Blanks between the grouped tail's lines** — rejected (D5 #2): the operator explicitly wants the block kept tight.
7. **Gate the blank lines on `isatty(stderr) && !-r`** — rejected: the tool-log lines themselves are not TTY-gated, so gating only the blanks would make the grouping depend on where the stream goes (the ADR 0007 D4 rationale).

## Consequences

- **Every** `[Tool …]` line (Output / Reason / Result / Action) is control-free and **never introduces** invalid UTF-8; the tool **result** fed to the model, the call's raw/stored arguments, `stdout`, `-r`, flags, exit codes, the frozen class-phrase vocabulary, the block literals, the block's bound/stop/spinner-yield semantics, and every persisted record are **unchanged**. The live turn output is blank-line grouped on the diagnostic stream.
- **ADR 0007 is superseded**, not edited: its D1–D6 remain the accurate record of *round 038's* `[Tool Output]` decisions, but its recorded scope boundary ("closed on `[Tool Output]` only") no longer describes the system. The authoritative policy is now this ADR.
- **Recorded residuals (forward):** sanitizing argument **keys** is a small extension over issue #80's literal "argument values" (included because a key is equally model-authored); the blank-line grouping is a reference divergence (the recorded place to revisit if tool-log parity is ever re-sought); the three-site blank-reason predicate is only *extended*, not consolidated — its single ownership stays on [#69](https://github.com/gosharplite/tellme/issues/69).
- The scanner's window-bounded residual (ADR 0007 D2 / N-6 / N-7) is unchanged: an unmapped or beyond-window sequence drops only the ESC and passes the rest of the line as visible text, bounded by the line / the block's byte budget.
- Immutable once `Accepted`; a future change to the removed class, the grouping policy, or the sanitize seam supersedes this ADR rather than editing it.
