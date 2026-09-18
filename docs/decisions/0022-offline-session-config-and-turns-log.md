# ADR 0022 — Offline session commands honour `-c`; a per-session `turns.log` and a `-t` reader

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** issue [#103](https://github.com/gosharplite/tellme/issues/103) (the `-l` ignores `-c` bug + no `-t`);
  [ADR 0013](0013-composition-root-injection.md) (the injected `Dependencies` seam — where the `turns.log` writer is bound),
  [ADR 0020](0020-cli-ui-decoupling.md) (RULE-E baseline **0** — the writer must be a **port**, not an `internal/cli → internal/ui` import);
  round 007 (the offline session commands), round 017/018/022/040 (the frozen turn-chrome contracts), round 026 (the `tokens.log` usage-log precedent);
  round 053 (`specs/plans/053-offline-session-config-and-turns-flag` — this ADR's round)

## Context

**Gap 1 — the offline session commands ignore `-c`.** A session is keyed by **mode** (`home.EnsureWorkspace(homeDir, mode)` → `output/<mode>/`). The **prompt** path resolves the mode from the `-c` config (`resolve()` → `res.Mode = cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))`, `internal/cli/cli.go:476,569`). The **offline session commands** (`-l`, prompt-less `--new`) instead go through `resolveWorkspace(homeDir)` → `historyMode(homeDir)`, which reads `TELL_ME_MODE` → else `config.Load(defaultConfigPath(homeDir))` → else `"butler"` — **never the `-c` path** (`cli.go:887,930,950,967,981,1154`). So with `TELL_ME_MODE` unset, `-l 1 -c architect.yaml` and `-l 1 -c butler.yaml` return **identical** output (both read `output/butler/`) — exit 0, silently wrong. This breaks the documented `tmg-chat-ingroup` / `tmg-grill-round` retrieve recipe (`env -u TELL_ME_MODE … -l 1 -r -c "<target>.yaml"`), blocking agent-to-agent relay on tellme.

**Gap 2 — no `-t`.** The reference's `-t`/`--turns` prints the current session's `turns.log` and exits. tellme has no such flag, and no `turns.log` artifact: a session workspace holds `history.jsonl`, `history.archive.jsonl`, `tokens.log`, `tokens.archive.jsonl`, `tokens.summary.json` (`grep` for `turns.log`/`TurnsLog` → 0 hits). The reference's `turns.log` persists the turn chrome (header + payload-status + metrics) via an event-logger; tellme **already renders** those exact lines (`internal/ui/turn.go`, `status.go`, `metrics.go`) — it just writes them to the diagnostic stream and discards them.

## Terminology

- **offline session command** — a command that resolves a session from `TELL_ME_HOME`/mode only (no provider resolution, no network): `-l`, prompt-less `--new`, and now `-t`.
- **mode** — the session key; the `MODE` of the effective config, overridden by `TELL_ME_MODE`.
- **chrome** — tellme's rendered per-turn operator lines (turn rule + header, payload status, metrics).

## Decision

**D1 — One mode-resolution precedence; `-c` supplies the mode, env still wins.** The offline session commands resolve the mode as: **(1)** `TELL_ME_MODE` (env) → **(2)** the `-c` config's `MODE` → **(3)** the default config's `MODE` → **(4)** `"butler"`. The `-c` read is **mode-only** (`config.Load` is a pure parse), so the commands stay **offline** — routing them through `resolve()` is rejected (it would add provider/registry/pricing resolution and change error semantics).

**D2 — An explicit `-c` that cannot be honoured FAILS.** When `-c <path>` is given and cannot be read/parsed, `-l`/`--new`/`-t` fail with an **existing** config-error phrase + non-zero exit (mirroring `reasonConfigMissing`/`reasonConfigInvalid`); no new phrase. Without `-c`, the tolerant default fallback (D1 steps 3–4) is preserved (round 007). Rationale: a named `-c` silently reading the default session is exactly the "plausible-but-wrong" failure the round fixes.

**D3 — The shared seam widens to `resolveWorkspace(homeDir, configPath)`.** One seam serves `-l`, `--new`, and `-t`; the callers thread `f.configPath`.

**D4 — A `-t` / `--turns` flag.** Registered in `parseFlags`; handled by `dispatchReporting` as an offline session command ordered **after `-l`** (`-d` → `-l` → `-t` → `--tool-usage`); prints `output/<mode>/turns.log` verbatim to `stdout` and exits `Success`; a missing/empty file is success.

**D5 — `tellme` writes its own per-session `turns.log` (Q1 → C1).** A plain-text file `output/<mode>/turns.log` holding tellme's **rendered chrome** — the same lines the diagnostic stream emits (turn rule + header, payload status, metrics), so no new format is invented and the E2E-pinned chrome contracts are inherited verbatim. The writer is an injected **port** (`internal/infrastructure/history` store, bound via `deps.Dependencies` at `cmd/tellme`; round 044 / ADR 0013), **best-effort** (a write failure never fails a turn), and `--new` archives it alongside `history`/`tokens`. Rejected: (A) print `tokens.log` (a usage log); (B) project `history.jsonl` (loses the chrome); (C2) reproduce the reference's exact bytes (contradicts the frozen chrome contracts).

## Consequences

- `-l`, `--new`, and `-t` become `-c`-honouring and mutually consistent; the documented `env -u TELL_ME_MODE … -l 1 -c <target>.yaml` retrieve works unchanged ([#103](https://github.com/gosharplite/tellme/issues/103) AC1–AC4).
- The offline property and the env-wins golden rule are preserved; an explicit bad `-c` now fails loudly.
- A new persisted session artifact (`turns.log`) enters the data model; `--new` archives it.
- The RULE-E baseline stays **0** — the writer must be a domain port, injected at the composition root.

## Forward items

- **RF-53-1** — `turns.log` follows the chrome: a future chrome change (a new header field) must decide `turns.log` in the same change.
- **RF-53-2** — the `--new` archive name for `turns.log` (`turns.archive.jsonl` vs `.log`) is an implementation choice, recorded at implementation.
- **RF-53-3** — bare `-l` defaulting to `1` (reference parity) remains **out of scope** ([#103](https://github.com/gosharplite/tellme/issues/103) does not ask for it; the SOP passes `-l 1`).
- **RF-53-4** — a self-diagnosing retrieve ([#103](https://github.com/gosharplite/tellme/issues/103) "also consider": print the resolved session/mode, or a `--json` listing) is a candidate, not required.
