# ADR 0045 — The `-l` history listing presents messages as a conversation

**Status**: Accepted (round 073)

**Date**: 2026-09-21

**Related**: round 007 (`-l` listing; the operator-messages-only contract) · round 006 (rendered vs `-r`/`--raw`, the glamour renderer, `WRAP_WIDTH`) · round 053 / ADR 0022 (offline session commands honour `-c`) · round 054 / ADR 0023 (the chrome colour policy + the `-l` count default) · round 051 / ADR 0020 (the `render` ports) · round 057 / ADR 0027, round 058 / ADR 0028 (colour elements) · [SESSION-BOOTSTRAP.md](../SESSION-BOOTSTRAP.md) · reference `tell-me-go/internal/ui/history.go`, `internal/ui/colors.go`, `internal/domain/ports/chat_service.go`.

## Context

`-l N` (round 007) printed the last N persisted messages as flat `role: content`
lines (`user:` / `assistant:`, lowercase, colon, no header, no rendering, no
separator). The reference's listing instead prints a `[ROLE]` **header line**
(uppercased, accented blue for the operator and magenta for the model on a
terminal stdout) followed by the body **rendered as Markdown** by glamour, with a
blank line after every message.

Round 073 was opened as an **operator request** (no anchor issue): *"tellme needs
to have `[USER] / [MODEL]` header lines (blue/magenta on a TTY stdout), body
rendered as Markdown by glamour, blank line between messages."*

Three high-impact gaps were settled by `/axb-clarify` at specify time (one
question at a time):

- **Q1 → A** — the listing **honours `-r`/`--raw`** (reference parity).
- **Q2 → A** — the listing stays **operator-messages-only** (no `[Tool Call]` /
  `[Tool Response]` lines; the round-007 contract is preserved).
- **Q3 → B** — **only the model body is rendered**; the operator prompt body is
  printed verbatim (a recorded divergence from the reference, which renders
  both).

## Decision

1. **A dedicated presentation port.** `internal/domain/render` gains `Listing`
   (+ `ListingMessage`, `ListingRole{ListingOperator,ListingModel}`,
   `ListingSpec{Colour,Raw,Width,Warn}`), implemented by `internal/ui/listing.go`
   (`ui.NewListing()`), injected through `deps.Dependencies.NewListing`. The
   bytes stay single-owned in `internal/ui` (round-051 / ADR 0020); `internal/cli`
   names no `internal/ui` type (the layer-discipline gate, ADR 0011). The role is
   a **domain enum**, not the stored/derived `user`/`assistant` string — the
   `[MODEL]` naming is presentation, and the mapping (`user → operator`,
   `assistant → model`) is one explicit table in the CLI.

2. **The per-message shape.** One **role header line** (`[USER]` / `[MODEL]`,
   uppercase, bracketed — printed **always**), the body normalised to end in
   exactly one newline, then **one additional newline** (the separator). So
   consecutive messages are separated by exactly one blank line, the output ends
   with a blank line, and a body that already ends in a newline cannot produce a
   double gap (the reference's `Fprintln`-after-every-block shape).

3. **Colour is `stdout`-gated.** `[USER]` is accented `\033[1;34m` (blue) and
   `[MODEL]` `\033[1;35m` (magenta) — the reference's codes, added to
   `internal/ui/colour.go` and applied through the shared, empty-safe `wrap` —
   **only when `stdout` is a terminal and `-r` is off**. A redirected/piped
   `stdout` therefore carries no *header accent* — but, exactly like the answer
   path (whose rendering is gated by `-r` **alone**, round 006), a rendered
   model body still carries glamour's own style output; only the `-r` path is
   escape-free end to end. No accent ever reaches `stderr` and none enters
   `turns.log`. The gate uses a **dedicated `stdout` terminal probe**
   (`stdoutTerminalDetector()` + the `TELL_ME_FORCE_STDOUT_TTY` diagnostic seam,
   mirroring the stderr seam) — tellme's first stdout probe, scoped to the
   listing.

4. **One rendering policy.** The listed **model** body is rendered by the same
   `ui.Renderer` the answer path uses (the same `GLAMOUR_STYLE`, the LaTeX
   sanitizer, `WithWordWrap`, and the degraded fallback + its one-time `[WARN]`).
   Under `-r` the model body is verbatim (no rendering, no colour). The operator
   body is verbatim in both modes.

5. **The width is resolved best-effort.** The listing resolves `WRAP_WIDTH` /
   `TELL_ME_WRAP_WIDTH` through the existing composition, but an unreadable or
   invalid width **degrades to the renderer default** instead of failing the
   listing — `-l` stays a read-only reporter (no new failure mode).

6. **Invariants preserved.** The listing stays strictly offline (no provider
   request) and terminal; the precedence (`-d` → `-l` → `-t` → `--tool-usage`),
   the `-l N` count semantics (last N **messages**), the round-053 `-c`-mode
   resolution, and the frozen `history.jsonl` format are unchanged; tool activity
   stays omitted; no new dependency.

## Consequences

- The listing reads like a conversation on a terminal and stays machine-plain
  when redirected; the in-group relay recipe keeps working byte-plain under `-r`.
- The operator-messages-only contract (round 007) is preserved; a future
  tool-activity parity request would supersede it (RF-073-4).
- tellme now has a **stdout** terminal probe; it is scoped to the listing colour,
  so the answer-path **Obs 1** (stdout *chrome* on a prompt turn) remains **OPEN**
  (RF-073-5).
- Divergences from the reference are recorded: the prompt body is **not**
  rendered (Q3 → B); the empty-session sentence `No history found.` is **not**
  adopted; the reference's `-l N "prompt"` composability and its `-b` pairing are
  **not** adopted (terminal/offline shape retained).

## Forward (non-blocking)

- **RF-073-1** — render the operator prompt body too (reference parity); gated on
  an operator request.
- **RF-073-2** — the empty-session `No history found.` sentence (not adopted).
- **RF-073-3** — the reference's `-l N "prompt"` list-then-chat composability and
  the `-b/--back` pairing (not adopted).
- **RF-073-4** — `[Tool Call]` / `[Tool Response]` listing parity (not adopted;
  supersedes only if requested).
- **RF-073-5** — the answer-path stdout-chrome observation (Obs 1) stays OPEN;
  the new stdout probe serves the listing only.
- **RF-073-6** — the best-effort width resolution on the `-l` path (a broken
  `WRAP_WIDTH` silently uses the renderer default).
- **RF-073-7** — the header is emitted on every message including the last; a
  trailing-blank-line policy change is a one-line adapter edit.
