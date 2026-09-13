# ADR 0003 — Terminal detection: a real isatty (`golang.org/x/term`)

- **Status:** Accepted
- **Date:** 2026-09-13
- **Deciders:** tellme owner
- **Supersedes:** —
- **Related:** round 012 (`specs/plans/012-interactive-multiline-prompt`); PR #31 architectural review **BLOCKER B1**;
  round-005 research Decision 1 (superseded); round-012 research Decision 1 (the "hold `x/term`" ruling, revised here)

## Context

Rounds 001–005 detected a terminal with a stdlib-only heuristic —
`info.Mode()&os.ModeCharDevice != 0` — deliberately avoiding `golang.org/x/term` (round-005 research
Decision 1). While the probe only gated *whether piped stdin was read*, the heuristic's imprecision was
benign: reading a character device yields EOF.

Round 012 **promotes the same probe to a behaviour gate**: it decides whether the interactive
multi-line reader engages. The heuristic is true for *any* character device — `/dev/null`, `/dev/zero`,
`/dev/random`, `/dev/tty` — not only a terminal. So `tellme < /dev/null` (and any cron / process
manager / CI that opens `/dev/null` as stdin) was treated as an interactive terminal: it printed the
reader hint instead of the boot report, and — because the empty/cancel path returns success without
resolving setup — **exited 0 on a missing configuration, masking the failure** (PR #31 review BLOCKER
B1). The heuristic is unsound for this use.

## Problem

tellme needs a probe that answers "is this stream a terminal?" *correctly* for a redirected device —
false for `/dev/null`, `/dev/zero`, a pipe, a regular file; true only for a real terminal — without a
pty, and while keeping the round-005 target of no new *module*.

## Decision

Adopt **`golang.org/x/term`** and its `IsTerminal(fd int) bool` as tellme's terminal probe, replacing
the `os.ModeCharDevice` heuristic.

- The probe lives behind the existing injected `isTTY func(any) bool` seam
  (`internal/cli.defaultIsTerminal`), so it stays unit-testable and swappable.
- Only `*os.File` streams can be a terminal; any other reader/writer (an injected fake, a
  `bytes.Buffer`) is reported non-terminal.
- `golang.org/x/term` is **already in the module graph transitively** (via glamour's tree), so adopting
  it **adds no new module** — it moves from the indirect to the direct `require` block. This preserves
  the round-005 spirit ("no new dependency") in substance while fixing the correctness defect.
- A diagnostic environment seam **`TELL_ME_FORCE_STDIN_TTY`** (truthy ⇒ report every stream as a
  terminal) is added so the interactive read is drivable end-to-end over a pipe without a pty
  (round-012 review RF1).

## Alternatives considered

1. **Keep `os.ModeCharDevice` and merely stop masking the error** (round-012 review option (b)) —
   rejected: it leaves the `tellme < /dev/null` *boot-report* regression in place (the reader still
   engages), so it is a partial fix only.
2. **A hand-rolled POSIX ioctl (`TCGETS` / `TIOCGETA`)** — dependency-free, but needs per-GOOS
   build-tagged files (the request constant and `syscall.Termios` differ across Linux/BSD/Darwin) for no
   benefit over `x/term`, which is already present. Rejected as needless surface.
3. **`github.com/mattn/go-isatty`** — also already in the graph and equally correct, but `x/term` is the
   more standard choice the review named. Rejected on preference.

## Consequences

- tellme carries one more **direct** dependency (`golang.org/x/term`), whose module was already
  downloaded transitively; `go.mod` moves it out of the indirect block. The "dependency-free TTY probe"
  wording in the round-005/012 artifacts is superseded (see `specs/truth/techstack.md`).
- `tellme < /dev/null` (and other non-terminal character devices) now take the **boot path**: a
  configuration failure is reported (exit 3) instead of masked as success, and no reading hint is
  printed.
- The E2E suite keeps a **hermetic** empty-pipe stdin for non-piped runs and gains an explicit
  **null-device** scenario (review RF2) so the character-device path cannot silently regress.
- Immutable once `Accepted`; a future probe change supersedes this ADR rather than editing it.
