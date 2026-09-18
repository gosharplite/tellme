# Technical Research: round 054 — `-l` defaults to `1` + green chrome accents

**Feature Branch**: `054-l-default-and-chrome-colour`
**Created**: 2026-09-19
**Status**: Draft (clarify round 1 CLOSED — Q1 gate · Q2 elements/divergence · Q3 `turns.log` plain)

## Terminology

- **chrome** — the rendered per-turn operator lines on the diagnostic stream: the turn rule + `╭─⠿ Turn <N> - <mode>` header, the `[HH:MM:SS] Payload: …` status lines, the `[HH:MM:SS] [Tool Reason] …` lines, the `[HH:MM:SS] [<provider>] M: …` metrics line, and `╰─⠿ Ready`.
- **green** — the reference's `colorGreen = "\033[0;32m"` (tell-me-go `internal/ui/colors.go`).

## Decisions

### D1 — `-l` takes an optional value: `NoOptDefVal="1"` + an args pre-pass

`tellme`'s `-l` is an int with no `NoOptDefVal`, so a bare `-l` is a pflag usage error. The reference sets `NoOptDefVal = "1"` **and** normalizes argv (`consumeOptionalIntFlag`) because pflag would otherwise parse `-l 5` as `-l=1` + positional `5`. tellme mirrors that: `fs.Lookup("list").NoOptDefVal = "1"` plus `consumeListValue(args)` that consumes an adjacent **integer** (`-l 5` → `-l=5`), leaves a non-integer token (a prompt), and stops at `--`. A non-positive explicit value stays a usage error (round-007).

### D2 — the colour gate: terminal `stderr` AND `-r` off (Q1)

The chrome is plain-text today (round-017 Decision 3). Colour is emitted only when the **diagnostic stream (`stderr`) is a terminal AND `-r` is off AND the surface is prompt-bearing** — the exact round-019 spinner gate (`chromeColour(opts, env) = opts.chrome && !opts.raw && env.stderrIsTerminal()`). Consequence: redirected/piped `stderr`, `-r`, and every offline path stay **byte-identical** to today, so the E2E's regex assertions over non-terminal `stderr` are unchanged. This **supersedes** round-017 D3 for the four elements (recorded in ADR 0023).

### D3 — the four green elements (Q2 — operator-locked, divergence accepted)

1. every `[Tool Reason]` line — **the whole line**;
2. the `MODE` token in **both** `Payload` lines (estimated + measured);
3. the **measured** token number (`...Payload: <green>438165</green>/1000000 tokens...`; the `~`-estimated number stays plain);
4. the third (**session**) cost in `╰─⠿ Ready`.

**Recorded divergence** — tell-me-go's own layout greens only (4); its `[Tool Reason]` line is **gray**, its `Payload` mode is **gray**, and its token number goes **yellow/red by budget ratio** (never green). tellme reuses the reference's **code** on **tellme's** element set.

### D4 — `turns.log` stays plain: the FILE leg is rendered with colour OFF (Q3; round-053 RF-53-1)

The per-call renderer holds **two** `render.Lines`: the coloured one (stderr) and a **plain** one (`dp.NewLines(false)`) for the file leg, and `emit(build)` writes each line to both sinks — the file leg is rendered plain **by construction**, not stripped afterwards. The round-053 tee is the raw file writer (no `MultiWriter`). So no ANSI enters the session artifact, while the `stderr` leg is coloured. (This was the PR #119 review **B-54-1** blocker: the first implementation teed the *coloured* string to both sinks.)

### D5 — implementation shape: colour lives in the adapters, not the pure formatters

The pure formatters (`FormatPayloadStatus`, `FormatReady`, `FormatToolReason`) keep their signatures and gain colour-aware siblings (`formatPayloadStatusColour`, `formatReadyColour`, `formatToolReasonColour`, plain `false` path byte-identical). The **adapters** (`ui.Lines`, `ui.ToolLineRenderer`) carry a `colour bool`; their constructors widen to `ui.NewLines(colour)` / `ui.ToolLines(colour)`, and the deps seams become `NewLines func(colour bool)` / `NewToolLines func(colour bool)` — so the bytes stay owned by `internal/ui` (RULE-E baseline stays **0**) and the CLI only resolves the gate. `chromeColour` is the one gate expression.

### D6 — Witness plan (falsifiable, reproduced then reverted)

- **(a)** revert `NoOptDefVal`/`consumeListValue` ⇒ bare `-l` is a usage error (the unit pin + the E2E Example red).
- **(b)** invert the colour gate (colour on a non-terminal) ⇒ the negative Example (`the session chrome carries no colour`) reds.
- **(c)** drop a colour site (e.g. the measured-token accent) ⇒ the unit pin + the positive E2E Example red.

### D7 — Truth impact & governance

- `specs/truth/techstack.md`: **CLI flag parsing** + **Session lifecycle flags** (`-l` optional value), **Turn chrome** + **Post-turn status lines** (the colour policy + the round-017 D3 supersession), **Not Introduced Yet** (the round-017 no-ANSI clause).
- `specs/truth/features/cli/history/inspecting-the-session-history.feature` (+ `history/dsl.md`): the bare-`-l` Rule/row.
- `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (+ `chat/dsl.md`): the colour Rules/rows.
- **ADR 0023** (this round), indexed.
- `specs/truth/data/data-model.dbml`: **NOOP** (the `turns_log_line` Note already says control-free — re-affirmed).
- No new dependency; POSIX-only.

### D8 — Scope guard

In scope: the `-l` optional value; the four green accents behind the gate; the truth/DSL/ADR. Out of scope: colouring any other element; ANSI in `stdout` or `turns.log`; `--json`; Windows; a security layer; conversation pruning; tool concurrency.

## Risks & residual items

- **E2E regex sensitivity** — the gate keeps the E2E's non-terminal assertions valid; the colour is asserted through the `TELL_ME_FORCE_STDERR_TTY` seam only. A future round that colours an element on a **non-terminal** path would break many assertions — recorded as a caution.
- **`\r`-redrawn spinner + colour interleave** — the spinner frames carry no colour, so the terminal read is unaffected; recorded.
- **`[Tool Reason]` whole-line wrap vs the reason transform** — colour wraps the ALREADY sanitized/capped line, so a model-authored escape cannot escape the wrapper (round 036/039 preserved).
- **`Ready` absent when usage is unreported** — element (4) only renders on a usage-reported turn (existing gate); the unit pin covers the colour, the E2E uses a usage-reporting provider.
