# System Analysis: round 053 — offline session commands honour `-c` + a `-t` flag

**Feature Branch**: `053-offline-session-config-and-turns-flag`
**Created**: 2026-09-19
**Status**: Draft

## Interfaces

| # | Interface | Kind | Surface | Delegate |
| --- | --- | --- | --- | --- |
| 1 | `tellme` CLI — offline session commands (`-l`, prompt-less `--new`, `-t`) + the `turns.log` session artifact | `cli` | flags, exit codes, `stdout`/`stderr`, the per-session `output/<mode>/` files | **carried forward to the contract owner `/axb-dsl-refine`** (a line-oriented CLI end has no analysis planner) |

**No API interface** (no OpenAPI/HTTP surface — `/axb-api-plan` = **NOOP**). **No standalone data interface** (a plain per-session text artifact; `/axb-data-plan` = **MODIFY** — records the `turns.log` artifact in `specs/truth/data/data-model.dbml`, a data-truth update, not a planner wave). **No UI** (the `-t` output is plain text on `stdout`; there is no new interactive surface).

## Waves

| Wave | Order | Interface | Planner | Note |
| --- | --- | --- | --- | --- |
| 1 | 1 | CLI end (the `-c`/`-t`/`turns.log` surface) | none — carried to `/axb-dsl-refine` | `wave-covers-interfaces`: the only interface is carried forward to its contract owner |

`/axb-dsl-refine` is the contract owner for the CLI end: it adds the executable Examples/`DSLRow`s for (a) the `-c`-honouring session selection and (b) the `-t` turn-log read, under `specs/truth/features/cli/**`.

## Scope notes

- **Offline guarantee** — the CLI end touches no provider/network; `-l`/`--new`/`-t` stay offline (a mode-only config read).
- **Persisted artifact** — `turns.log` is a new per-session file; the data truth records it (`specs/truth/data/data-model.dbml`), the DAO is best-effort.
- **No new Makefile target**; `go.mod`/`go.sum` unchanged.
- **RULE-E baseline stays 0** — the `turns.log` writer is injected as a **domain port** at `cmd/tellme` (no `internal/cli → internal/ui` import).
