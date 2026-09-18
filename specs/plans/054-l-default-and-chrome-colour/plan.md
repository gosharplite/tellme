# System Analysis: round 054 — `-l` default-1 + green chrome accents

**Feature Branch**: `054-l-default-and-chrome-colour`
**Created**: 2026-09-19
**Status**: Draft

## Interfaces

| # | Interface | Kind | Surface | Delegate |
| --- | --- | --- | --- | --- |
| 1 | `tellme` CLI — the `-l` count default + the terminal-gated green chrome accents | `cli` | flags, exit codes, `stdout`/`stderr` | **carried forward to the contract owner `/axb-dsl-refine`** (a line-oriented CLI end has no analysis planner) |

**No API interface** (`/axb-api-plan` = **NOOP**). **No data interface** (`/axb-data-plan` = **NOOP** — `turns.log` stays plain; no persisted-schema change). **No UI** (chrome colours are plain `stderr` text, not an interactive surface).

## Waves

| Wave | Order | Interface | Planner | Note |
| --- | --- | --- | --- | --- |
| 1 | 1 | CLI end (the `-l` flag + the chrome colour) | none — carried to `/axb-dsl-refine` | `wave-covers-interfaces`: the only interface is carried forward to its contract owner |

`/axb-dsl-refine` adds: (a) the bare-`-l` Rule/Example + row (`history` module); (b) the colour Rules/Examples + rows (`chat` module).

## Scope notes

- **The colour gate** — one expression (`chromeColour`) = `opts.chrome && !opts.raw && stderrIsTerminal()`; the bytes stay owned by `internal/ui`; the RULE-E baseline stays **0**.
- **`stdout` byte-exact**; `turns.log` plain (round-053 RF-53-1).
- **No new dependency**; no Makefile target; `go.mod`/`go.sum` unchanged.
- **The reference's green code** (`\033[0;32m`) is the colour; the element set is tellme's own (a recorded divergence).
