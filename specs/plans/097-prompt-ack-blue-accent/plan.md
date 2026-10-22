# Plan — round 097 `097-prompt-ack-blue-accent`

Truth owner handoffs: `/axb-technical-research` (→ **ADR 0067**, techstack `NOOP`) ·
`/axb-dsl-refine` (→ `specs/truth/features/cli/chat/**`). Interface inventory: **one** (`cli`).

## 1. Interface inventory

| Interface | Involved? | Why | Truth home |
| --- | --- | --- | --- |
| **`cli`** | **YES** | The accent is a diagnostic-stream (`stderr`) presentation contract on a CLI turn. | `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` + `…/chat/dsl.md` |
| `backend` | no | No provider/loop behaviour changes. | — |
| `frontend` | no | No web surface. | — |
| API (`specs/truth/contracts/**`) | **NOOP** | A standalone CLI has no OpenAPI endpoints. | — |
| data (`specs/truth/data/**`) | **NOOP** | No persisted state changes; the ack was and remains excluded from `turns.log`. | — |
| `ui/**` | **NOOP** | The `-i` editor frame is untouched; the round changes a plain chrome line's accent. | — |

## 2. Analysis waves

**Wave 1 (single interface, single line).** No cross-interface dependency; no parallel batch. The
change is one call-site flag, one adapter receiver, and one pure formatter — all inside the existing
`internal/cli` → `internal/ui` seam.

## 3. The change surface (call chain)

```
renderTurn(...opts, env, dp)                              internal/cli/cli.go
  └ if opts.chrome {                                      :850   (ALL prompt surfaces)
      if opts.echo { Fprintln(env.stderr, prompt) }       :855   (unchanged — stays plain)
      emitInputCaptured(env, dp.NewLines(chromeColour(opts, env)))   :858  ← THE CHANGE
    }
      └ render.Lines.InputCaptured(now)                   internal/domain/render/ports.go:23  (signature unchanged)
          └ ui.Lines.InputCaptured(t)                     internal/ui/render_ports.go:30     ← value receiver + colour
              └ formatInputCapturedColour(t, l.colour)    internal/ui/turn.go               ← new sibling
                  └ blue(line, true) / wrap(line, …)      internal/ui/colour.go:62 / :71    (reused)
```

Gate: `chromeColour(opts, env)` = `spinnerGate` = `opts.chrome && !opts.raw && env.stderrIsTerminal()`
(`internal/cli/cli.go:1116`). **Reused, never duplicated.**

## 4. Blast radius (what must NOT move)

| Surface | Expected change |
| --- | --- |
| `stdout` (the answer) | **none** — byte-identical |
| `turns.log` | **none** — a separate plain writer (`cli.go:177`); the ack never enters it (round 053 F-53-3(i)) |
| The `-r` path's **bytes** | **none** — the ack still prints, merely uncoloured (EC-002) |
| The redirected/piped path's bytes | **none** — byte-identical plain form (EC-001) |
| `-i` prompt echo (`cli.go:855`) | **none** — stays plain (issue #204 rev-1 withdrawn) |
| Every other chrome accent | **none** — green / yellow / grey element sets unchanged |
| Provider request / history / usage | **none** |
| `go.mod` / `go.sum` / `make verify` members | **none** |

## 5. Interfaces handed to the planners

- **`/axb-api-plan` → NOOP** (no API surface).
- **`/axb-data-plan` → NOOP** (no state).
- **`/axb-ui-plan` → NOOP** (plain line-oriented chrome; the `-i` editor is untouched).
- **`/axb-dsl-refine` → the CLI contract owner** (the colouring feature + its DSL rows).

## 6. Test strategy (feeds `/axb-tasks`)

| Tier | Carrier |
| --- | --- |
| **E2E (positive)** | A colouring Example with `TELL_ME_FORCE_STDERR_TTY=1` asserting the blue pair on the ack line (whole line, timestamp inside the wrap). |
| **E2E (cross-surface)** | A second colouring Example on the `-i` surface (both `TELL_ME_FORCE_STDIN_TTY` and `TELL_ME_FORCE_STDERR_TTY`) asserting the same accent — this is what makes FR-005 a witness rather than a claim. |
| **E2E (negative)** | The widened `the session chrome carries no colour` row (now also forbidding `\033[1;34m`) on a piped run, plus a `-r` Example asserting the ack is **present and plain**. |
| **Unit** | The plain formatter's existing pin stays byte-exact; a new pin asserts the colour variant wraps the **whole** line (timestamp included) and that the disabled path returns the plain form **verbatim**. |
| **Artifact** | The existing `the session turn log carries no decoration` row (already run on a colour-enabled turn) covers `turns.log` for this accent at no extra cost. |
