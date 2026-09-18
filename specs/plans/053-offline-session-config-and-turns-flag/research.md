# Technical Research: round 053 — close [#103](https://github.com/gosharplite/tellme/issues/103) (offline session commands honour `-c`; add `-t`)

**Feature Branch**: `053-offline-session-config-and-turns-flag`
**Created**: 2026-09-19
**Status**: Draft (clarify round 1 CLOSED — Q1 → (C1) · Q2 → (A))

## Terminology

- **offline session commands** — `-l`, prompt-less `--new`, and (new) `-t`: commands that resolve a session from `TELL_ME_HOME`/mode only (no provider resolution, no network). Code: `resolveWorkspace` (`internal/cli/cli.go:967`).
- **chrome** — the rendered per-turn operator lines on the diagnostic stream: the turn rule + `╭─⠿ Turn <N> - <mode>` header (`internal/ui/turn.go`), the `[HH:MM:SS] Payload: …` status (`internal/ui/status.go`), the `[HH:MM:SS] [<provider>] M: …` metrics (`internal/ui/metrics.go`).
- **turn log** — the new per-session `output/<mode>/turns.log` recording that chrome (Q1 → (C1)).

## Decisions

### D1 — one mode-resolution seam; `-c` supplies the mode (env still wins)

Replace `historyMode(homeDir)`'s `defaultConfigPath(homeDir)`-only read with an **explicit-config-aware** resolution. Precedence (exactly [#103](https://github.com/gosharplite/tellme/issues/103) AC1/AC2):

1. `TELL_ME_MODE` (env) → wins if set (the golden rule).
2. else the **`-c` config's `MODE`** when `-c` is given.
3. else the **default** config's `MODE` (`defaultConfigPath`).
4. else `"butler"`.

The read is **mode-only** — `config.Load` is a pure `os.ReadFile`+`yaml.Unmarshal` (`internal/config/config.go:168-178`), so this stays **offline** ([#103](https://github.com/gosharplite/tellme/issues/103) AC3). Routing through `resolve()` is **rejected** (provider/registry/pricing resolution; would change error semantics — a provider-invalid config must not fail `-l`).

### D2 — the failure policy for an explicit `-c` (Q2 → (A))

When `-c <path>` is **given** and `config.Load` fails, the offline session command **fails** with an existing config-error phrase + non-zero exit (mirroring `resolve()`'s `reasonConfigMissing`/`reasonConfigInvalid`). When `-c` is **absent**, the tolerant fallback (steps 3–4 of D1) is preserved (round-007). The failure reuses the frozen `tellme: {phrase}` vocabulary — no new phrase.

### D3 — the shared seam: `resolveWorkspace(homeDir, configPath)`

Widen the offline helper to take the config path — `resolveWorkspace(homeDir, configPath string) (string, *resolveError)` → `home.EnsureWorkspace(homeDir, historyMode(homeDir, configPath))`. Update the two existing callers (`renderHistoryList` `cli.go:930`, `renderNewSession` `cli.go:951`) and thread `f.configPath` from `dispatchReporting` (`cli.go:887`) and the `--new` dispatch sites (`cli.go:407, 430`). One seam covers `-l`, `--new`, and the new `-t` (A1).

### D4 — the `-t` flag: `-t` / `--turns`

Register `fs.BoolVarP(&o.turns, "turns", "t", false, "Print the current session's turns log and exit.")` in `parseFlags`. Handle it in `dispatchReporting` as an **offline session command**, ordered **after `-l`** in the precedence batch (`-d` → `-l` → `-t` → `--tool-usage`) — `-d` and `-l` keep their round-004/007 precedence; `--tool-usage` stays last (needs no session). `-t` resolves the session with D1/D3, opens `output/<mode>/turns.log`, copies it to `stdout` verbatim, exits `Success`; a missing/empty file is success (mirrors `-l` on an empty session). Sanitization: the file is written chrome (already control-free) — see D5.

### D5 — the `turns.log` artifact + writer seam (Q1 → (C1))

**What**: a per-session plain-text file `output/<mode>/turns.log` holding **the subset of diagnostic lines routed through the chrome sink**: the per-call renderer's `─` turn rule, `╭─⠿ Turn <N> - <mode>` header, `[HH:MM:SS] Payload: …` status lines (estimated + measured), grouped `[HH:MM:SS] [Tool Reason] …` lines, the `[HH:MM:SS] [<provider>] M: …` metrics line, and the `╰─⠿ Ready` summary. The input-capture acknowledgement, the `-i` prompt echo, error phrases, the spinner, and the `[Tool …]`/`[Tool Output]` block are **not** persisted (fold F-53-3(i) — narrow the sink). No new format (the renderers are the formatter).

**Writer seam**: a new domain port `history.TurnsLogStore` (or a func-typed `deps` seam) — `Open(workspace) (io.WriteCloser, error)` / an `Append(line)` — implemented in `internal/infrastructure/history/turns_log.go` (a `turnsLogStore`, mirroring `usage_store.go`). The CLI **tees** each chrome line it already writes to `stderr` into the store; the store is **best-effort** (a write failure never fails the turn — A3). Bound at `cmd/tellme` through the injected `deps.Dependencies` (round 044 / ADR 0013 precedent; keeps `internal/cli` naming no infrastructure type).

**Lines recorded** (C1 — whatever the diagnostic stream emits): the turn rule + `╭─⠿ Turn …` header, the `[HH:MM:SS] Payload: …` lines, the `[HH:MM:SS] [<provider>] M: …` metrics line. The turn/status/metrics bytes are pinned by the E2E, so `turns.log` inherits their contracts verbatim.

**`--new`**: archives `turns.log` alongside `history`/`tokens` (a `turns.archive.jsonl`-style sibling, or the same archive mechanism — RD detail at implementation).

**Rejected**: (A) print `tokens.log` (a usage log, not a turn trace — not what `-t` names); (B) project `history.jsonl` (loses the chrome; a new format); (C2) reproduce `tell-me-go`'s exact bytes (contradicts tellme's frozen chrome contracts, rounds 017/018/022/040).

### D6 — Witness plan (falsifiable, reproduced then reverted)

- **(a)** Revert D1's `-c` mode read ⇒ `-l 1 -c a.yaml` and `-l 1 -c b.yaml` return **identical** output (the #103 repro).
- **(b)** Revert D5's write ⇒ a run produces **no** `turns.log`; `-t` finds nothing.
- **(c)** Revert D4's `-c` resolution for `-t` ⇒ `-t -c <non-default>` reads the default session.
- **The writer's carrier (F-53-1)** is a CLI-tier pin — `internal/cli/turns_log_test.go` (`TestRunTurnTeesChromeIntoTurnsLog`): the positive asserts the chrome reaches the injected store's buffer; the negative asserts an empty workspace tees nothing. **The `-t` precedence pin (F-53-4)** is `internal/cli/dispatch_test.go` (`TestDispatchReportingPrecedence`): `-l` beats `-t` beats `--tool-usage`.

### D7 — Truth impact & governance

- **techstack truth** (`specs/truth/techstack.md`): MODIFY the **CLI flag parsing** row (`-t`), the **Session lifecycle flags** row (the `-c`-honouring session selection), the **Turn chrome** row (a persisted `turns.log` tee), and ADD a **Turn log (`turns.log`)** row.
- **data truth** (`specs/truth/data/data-model.dbml`): ADD the `turns_log` artifact (per-session plain text; archived on `--new`).
- **CLI interface truth** (`specs/truth/features/cli/**`): the `-t` read + the `-l -c` selection gain Examples/`DSLRow`s (owner `/axb-dsl-refine`).
- **ADR 0022**: the round's decision record (D1–D5), indexed in `docs/decisions/README.md`.
- **No** new dependency; POSIX-only; no `flock`.

### D8 — Scope guard

In scope: the `-c` mode resolution for `-l`/`--new`/`-t`; the `turns.log` artifact + writer; the `-t` reader. Out of scope: the prompt path (already honours `-c`); bare `-l` defaulting to `1`; `tokens.log` changes; the reference's exact `turns.log` bytes; conversation pruning; tool concurrency; the `-i` terminal requirement.

## Risks & residual items

- **Chrome-format coupling** — `turns.log` records tellme's rendered chrome; a future chrome change (e.g. a new header field) must decide whether `turns.log` follows (it should — the file is the stream's shadow). Recorded forward.
- **Widening the offline helper** — every caller of `resolveWorkspace` must pass the config path; a missed caller silently keeps the bug (the witnesses (a)/(c) catch it).
- **Write-site location** — the tee site (CLI composite observer vs the `internal/ui` renderers) is an implementation detail; the `internal/cli`→`internal/ui` RULE-E baseline is **0** (ADR 0020), so a `ui`-side writer must be injected as a **domain port**, not imported.
- **`-t` precedence** — a combined `-l -t …` resolves by the recorded order (`-l` first); pinned by a unit test.
- **Archive naming** — the `--new` archive name for `turns.log` is `turns.archive.log` (RD choice, recorded). The archive is **best-effort** (ADR 0022 forward RF-53-2).
