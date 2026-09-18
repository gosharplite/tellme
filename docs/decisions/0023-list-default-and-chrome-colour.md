# ADR 0023 — `-l` takes an optional count; green chrome accents on a terminal

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [ADR 0006](0006-tool-reason-fold-and-cap.md) (the reason transform the colour wraps), [ADR 0009](0009-spinner-dual-timer-and-streaming-liveness.md) (the spinner gate the colour reuses), [ADR 0022](0022-offline-session-config-and-turns-log.md) (`turns.log`, kept plain); round 017 (the round-017 D3 "no ANSI" chrome decision this round supersedes for four elements); round 018 (the post-turn status lines); round 053 (`turns.log`); round 054 (`specs/plans/054-l-default-and-chrome-colour` — this ADR's round)

## Context

**The `-l` count.** `tellme`'s `-l`/`--list` is an int flag with no `NoOptDefVal`, so a bare `-l` is a pflag usage error. The reference (`tell-me-go`) sets `NoOptDefVal = "1"` **and** normalizes argv (`consumeOptionalIntFlag`), because pflag would otherwise parse `-l 5` as `-l=1` + a positional `5`. The SOP's shorthand (`-l 1`) works, but the bare form does not.

**The chrome is plain.** Round-017 Decision 3 made the non-TUI chrome plain text ("no ANSI this round"). The operator wants four elements accented in **green** for scannability on a long session, using the reference's green code.

## Decision

**D1 — `-l` takes an optional value.** `fs.Lookup("list").NoOptDefVal = "1"` plus a `consumeListValue(args)` pre-pass: it consumes an **adjacent integer** (`-l 5` → `-l=5`), leaves a **non-integer** token (a prompt) untouched, and stops at `--`. A non-positive value stays a usage error (round-007). `tellme -l` ≡ `tellme -l 1`.

**D2 — the colour gate.** The green is emitted only when the diagnostic stream (`stderr`) is a **terminal**, `-r` is **off**, and the surface is prompt-bearing — the round-019 spinner gate (one expression, `chromeColour`). Redirected/piped `stderr`, `-r`, the offline paths, `stdout`, and `turns.log` stay **byte-identical** to the plain form. This **supersedes round-017 Decision 3** for the four elements below.

**D3 — the four elements (operator-locked).** Green the reference's `\033[0;32m`:
1. every `[Tool Reason]` line — **the whole line**;
2. the `MODE` token in **both** `Payload` lines;
3. the **measured** token number in the measured `Payload` line (the `~`-estimated number stays plain);
4. the third (**session**) cost in `╰─⠿ Ready`.

**Recorded divergence** — the reference's own layout greens only (4); its `[Tool Reason]` line is **gray**, its `Payload` mode is **gray**, and its token number goes yellow/red by budget ratio. tellme reuses the reference's **code** on **tellme's** element set.

**D4 — `turns.log` stays plain (the FILE leg is rendered with colour OFF).** The per-call renderer writes each chrome line to **two** sinks: the diagnostic stream via the **coloured** renderer, and the session `turns.log` via a **separate plain** renderer (`dp.NewLines(false)`), so the artifact is plain **by construction** — never produced by stripping colour after the fact (the stronger contract; a strip-at-writer would re-couple the artifact to "we remove our own colour"). `stdout` and the offline paths never carry colour.

**D5 — the colour lives in the adapters.** The pure formatters keep their signatures (colour-aware siblings, plain `false` path byte-identical); `ui.Lines`/`ui.ToolLineRenderer` carry a `colour bool`; the deps seams widen to `NewLines func(colour bool)`/`NewToolLines func(colour bool)` — the bytes stay owned by `internal/ui` (RULE-E baseline stays **0**).

## Consequences

- `tellme -l` = `tellme -l 1`; the flag matches the reference's shorthand.
- On a terminal the four chrome elements are green; off a terminal, under `-r`, in `stdout`, and in `turns.log` **nothing changes** (byte-identical), so the E2E (non-terminal) assertions and the `turns.log` "control-free" spec both hold.
- The round-017 D3 "no ANSI" clause is retired for the four elements (this ADR is the supersession record).
- No new dependency; POSIX-only; `go.mod`/`go.sum` unchanged.

## Forward items

- **RF-54-1** — colouring **any other** element (the `Ready` label, the metrics numbers, the turn header, the spinner) is out of scope; a future round decides per element.
- **RF-54-2** — a **non-terminal** colour mode (e.g. an opt-in `--color=always`) is **not** provided; the gate is terminal-only (a caution: colouring a non-terminal path would break many E2E regex assertions).
- **RF-54-3** — the round-053 self-diagnosing retrieve (RF-53-4) could adopt the same colour discipline if added.
- **RF-54-4** — `-l`'s optional value is implemented via a bespoke pre-pass; if more optional-int flags appear (`-b`/`--back`-style), the pre-pass should be generalised. Its boundary cases (recorded): the pre-pass is **positional-agnostic** (`tellme "p" -l 5` → `-l=5` + prompt `p`) and **misses combined short flags** (`-rl 5` → `-l`=1 + prompt `5`); and `-l` wins the dispatch precedence over a preserved positional token (round-007), so `-l hello` lists and drops the prompt.
