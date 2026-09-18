# Feature Specification: close [#103](https://github.com/gosharplite/tellme/issues/103) — the offline session commands honour `-c` (`-l`, prompt-less `--new`) + a `-t` (turns-log) flag (round 053)

**Feature Branch**: `053-offline-session-config-and-turns-flag`

**Created**: 2026-09-19

**Status**: Draft — **clarify round 1 CLOSED**: **Q1 → (C1)** (`tellme` writes its **own** `turns.log` — its rendered turn chrome — and `-t` prints it); **Q2 → (A)** (an **explicit** `-c` that cannot be honoured **fails**; an absent default stays tolerant). Produced by `/axb-specify`; Q1/Q2 folded via `/axb-clarify`.

**Input**: [#103](https://github.com/gosharplite/tellme/issues/103) — *"`-l` ignores `-c` for session selection (+ no `-t`) — blocks the documented agent-to-agent grill/chat plumbing on `tellme`"* — plus the operator's tasking: *"let fix below items: (1) `-c` needs to work when using `-l` or `--new`; (2) need `-t` (turns-log) flag."*

**Programme goal (operator-declared)**: **close [#103](https://github.com/gosharplite/tellme/issues/103)** — make the `tmg-chat-ingroup` / `tmg-grill-round` retrieve recipe work **as documented** on `tellme` (`env -u TELL_ME_MODE … -l 1 -r -c "<target>.yaml"` must read the **target's** session), and give the SOP's optional monitoring step a `tellme` equivalent (`-t`).

**Behaviour intent**: **MODIFY (user-facing)** — the offline session commands (`-l`, prompt-less `--new`) change **which session** they resolve when `-c` is given with `TELL_ME_MODE` unset; a **new flag** `-t`/`--turns` is added. The **prompt** path, the **env-wins** golden rule, the **offline** property of the session commands, and every existing stream/exit-code contract are **unchanged**.

---

## Grounded in the current system

Measured 2026-09-19 @ `dev` `f9253a7` (a static read; the round re-measures at implementation).

### Gap 1 — the offline session commands drop `-c`

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go:887` | `dispatchReporting`: the `-l` branch calls `renderHistoryList(homeDir, f.list, env, newHistoryStore)` — **`f.configPath` is not passed** |
| `internal/cli/cli.go:930` | `renderHistoryList` → `ws, rerr := resolveWorkspace(homeDir)` |
| `internal/cli/cli.go:950` | `renderNewSession` (prompt-less `--new`; called at `:407`, `:430`) → `resolveWorkspace(homeDir)` — **also drops `-c`** |
| `internal/cli/cli.go:967-971` | `resolveWorkspace(homeDir)` → `home.EnsureWorkspace(homeDir, historyMode(homeDir))` |
| `internal/cli/cli.go:981-986` | `historyMode(homeDir)`: `TELL_ME_MODE` env → else `config.Load(defaultConfigPath(homeDir)).EffectiveMode("")` → else `"butler"` |
| `internal/cli/cli.go:1154-1160` | `defaultConfigPath(homeDir)` = `$TELL_ME_HOME/configs/<TELL_ME_MODE\|butler>.yaml` |

**Contrast — the paths that honour `-c`** all call `resolve(homeDir, configPath)`, which loads `-c` and sets `res.Mode = cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))` (`cli.go:476, 569`): boot (`:580`), turn (`:611`), `-i` (`:251`), `-d` (`:1130`).

**Consequence (reproduced on the issue)**: with `TELL_ME_MODE` unset, `-l 1 -c architect.yaml` and `-l 1 -c butler.yaml` return **identical** output (both read `output/butler/`) — exit 0, silently wrong.

### Gap 2 — no `-t` (turns-log) flag

`parseFlags` (`internal/cli/cli.go:442-455`) registers `-c -d --version -r --new -l -i --tool-usage`; there is **no `-t`**. The reference's `-t`/`--turns` prints the current session's `turns.log` and exits (`tell-me-go internal/cli/chat_command.go:92`).

### What a session workspace actually holds (measured)

`$TELL_ME_HOME/output/<mode>/` = `history.jsonl`, `history.archive.jsonl`, `tokens.log`, `tokens.archive.jsonl`, `tokens.summary.json` — **there is no `turns.log`** in `tellme` (`grep` for `turns.log`/`TurnsLog` in tellme → 0 hits). A `turns.log` exists only in `tell-me-go`, where `internal/infrastructure/logging/async_turns_logger.go` persists the session's turn chrome (the turn header, payload-status lines, metrics line) and its `-t` (`StreamTurnsLog`) prints that file.

**But tellme already *renders* exactly those chrome lines** (they are just written to the diagnostic stream and discarded): the turn rule + header (`internal/ui/turn.go`), the payload status (`internal/ui/status.go`), the metrics line (`internal/ui/metrics.go`). So **Q1 → (C1)** adds a per-session **sink that persists that same chrome** to `output/<mode>/turns.log`; `-t` reads it. The formatter is the existing renderer; no new format, no re-opening of the round-017/018/022/040 chrome contracts. The write site/seam is an RD decision (`research.md`).

### Invariants that must survive

- **`-l` stays offline** ([#103](https://github.com/gosharplite/tellme/issues/103) AC3): the fix MUST NOT route the session commands through `resolve()` (no provider/registry/pricing resolution, no network). The `-c` read MUST be a **mode-only** parse — `config.Load` is already a pure `os.ReadFile`+`yaml.Unmarshal` (`internal/config/config.go:168-178`), so this is within the offline boundary.
- **`TELL_ME_MODE` still wins** ([#103](https://github.com/gosharplite/tellme/issues/103) AC2 / the `tmg-chat-ingroup` golden rule): precedence becomes `TELL_ME_MODE` → else `-c`'s `MODE` → else default config → else `"butler"`.
- **Round-007 tolerance**: `-l`/`--new` work when the configuration is **absent**; the fix MUST NOT turn an absent default config into a hard failure (the exact behaviour for an absent/invalid **explicit** `-c` is clarify Q2).
- **Closed phrase vocabulary**: the frozen `tellme: {phrase}` stderr vocabulary (`specs/truth/features/cli/**/dsl.md`) MUST NOT be widened by the new flag/paths.

---

## Clarify round 1 — OPEN

> Per `/axb-specify` Phase 2: two gaps would change the acceptance criteria and are therefore escalated rather than assumed. Asked **one at a time** (the `/axb-clarify` rule).

| # | Question | Options | Status |
| --- | --- | --- | --- |
| **Q1** | **What does `-t` print, and who writes it?** `tellme` has no `turns.log` today. | **(A)** print the existing `tokens.log` (per-call usage) · **(B)** print a turn trace projected from `history.jsonl` · **(C)** `tellme` **writes its own `turns.log`** and `-t` prints it — with **(C1)** recording **tellme's own rendered chrome** (the existing, E2E-pinned turn/status/metrics lines) or **(C2)** reproducing `tell-me-go`'s exact bytes. | ✅ **CLOSED — (C1)** |
| **Q2** | **`-c` load failure on an offline session command**: when `-c <file>` is given but the file is missing/invalid, should `-l`/`--new`/`-t` **fail** or **fall back** to the default/`butler` (today's tolerant behaviour)? | **(A)** Fail on an **explicit** `-c` (the user named it); still tolerate an **absent default**. **(B)** Always fall back (status quo tolerance). | ✅ **CLOSED — (A)** |

### Q2 → (A) (LOCKED) — an explicit `-c` that cannot be honoured FAILS

**Decision**: when `-c <path>` is **explicitly given** and cannot be read/parsed, the offline session commands (`-l`, prompt-less `--new`, `-t`) **fail** with an existing config-error phrase and a non-zero exit — mirroring the prompt path's `reasonConfigMissing` / `reasonConfigInvalid`. When **no** `-c` is given, the tolerant default-config fallback (`TELL_ME_MODE` → default → `butler`) is **preserved** (round-007).

**Rationale**: exactly the failure mode [#103](https://github.com/gosharplite/tellme/issues/103) is about — a named `-c` silently reading the **default** session with exit 0 is "plausible-but-wrong". An explicit `-c` is a user assertion; honouring it or failing loudly are the only honest outcomes. `TELL_ME_MODE` still wins over `-c` (the golden rule, [#103](https://github.com/gosharplite/tellme/issues/103) AC2) — so the failure applies only when the mode is actually needed from `-c` (env unset).

### Q1 → (C1) (LOCKED) — `tellme` writes its own `turns.log`; `-t` prints it

**Decision**: `tellme` gains a **per-session `turns.log`** under `output/<mode>/`, written on the turn path, containing **tellme's own rendered turn chrome** — the same plain-text lines it already renders to the diagnostic stream and which the E2E already pins byte-for-byte:

- the `──…──` turn rule (`internal/ui/turn.go` `TurnRule()` / `FormatTurnOpening`),
- the `╭─⠿ Turn <N> - <mode>` header (`internal/ui/turn.go` `FormatTurnHeader`),
- the `[HH:MM:SS] Payload: …` status lines (`internal/ui/status.go` `FormatPayloadStatus`),
- the `[HH:MM:SS] [<provider>] M: …` metrics line (`internal/ui/metrics.go` `FormatMetrics`).

**Rationale**: this is the reference's *role* (a persisted turn trace `-t` prints) realised with **tellme's already-frozen presentation** — no re-opening of the round-017/018/022/040 formatting (rejected option **(C2)**: reproducing `tell-me-go`'s exact bytes would contradict tellme's settled chrome contracts). No new *format* is invented: the chrome renderers are the formatter.

**Scope consequence (accepted)**: the round now has three parts — US1 (the `-c` fix) + the **`turns.log` writer** + US2 (the `-t` reader). The writer is an **RD mechanism** (`research.md` D-series: the sink seam + the write site), and the artifact is a **persisted session file** → `specs/truth/data/data-model.dbml` gains a `turns_log` artifact (like the `history`/usage artifacts) and `--new` archives it.

**Disclosed assumptions (low-impact; not escalated)** — see *Assumptions* A1–A6.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the offline session commands select the session named by `-c` (Priority: P1)

As an **orchestrator agent**, I want `tellme -l 1 -c "<target>.yaml"` (with `TELL_ME_MODE` unset) to read the **target agent's** session — the same session the matching `tellme -c "<target>.yaml" < prompt` written to — so the `tmg-chat-ingroup` / `tmg-grill-round` **retrieve** returns the target's reply instead of my own.

**Why this priority**: it is the **load-bearing** gap ([#103](https://github.com/gosharplite/tellme/issues/103) gap 1); without it the agent-to-agent round cannot be relayed at all.

**Independent verification**: seed session A (`-c a.yaml`, MODE `alpha`) and session B (`-c b.yaml`, MODE `beta`); `-l 1 -c a.yaml` returns A's last message and `-l 1 -c b.yaml` returns B's ([#103](https://github.com/gosharplite/tellme/issues/103) AC4).

**Acceptance Scenarios**:

1. **Given** `TELL_ME_MODE` is **unset** and `configs/a.yaml` declares `MODE: alpha`, **When** `tellme -l N -c configs/a.yaml` runs, **Then** it lists the last N messages of `output/alpha/` (the session the prompt path writes for `-c configs/a.yaml`).
2. **Given** two configs `a.yaml` (`alpha`) and `b.yaml` (`beta`) with distinct seeded sessions, **When** `-l 1 -c a.yaml` then `-l 1 -c b.yaml` run, **Then** the two outputs are their respective last messages (not identical) ([#103](https://github.com/gosharplite/tellme/issues/103) AC1/AC4).
3. **Given** `TELL_ME_MODE=gamma` **is set**, **When** `-l N -c a.yaml` runs, **Then** the session read is `output/gamma/` — the env still wins ([#103](https://github.com/gosharplite/tellme/issues/103) AC2).
4. **Given** a prompt-less `tellme --new -c configs/a.yaml`, **When** it runs, **Then** it archives `output/alpha/`'s session — not the default/`butler` session.
5. **Given** `-l`/`--new`, **When** it resolves the session, **Then** it performs **no** provider/registry/pricing resolution and **no** network access ([#103](https://github.com/gosharplite/tellme/issues/103) AC3).

**Functional Requirements**:

- **FR-001**: `-l` and prompt-less `--new` MUST resolve the session **mode** as `TELL_ME_MODE` (if set) → else the **`-c` config's `MODE`** → else the default config's `MODE` → else `"butler"` — replacing the current `defaultConfigPath(homeDir)`-only read.
- **FR-002**: The `-c` read on this path MUST be **mode-only** and stay **offline** (no full `resolve()`, no provider/registry/pricing resolution, no network); `config.Load` (a pure parse) is the permitted mechanism.
- **FR-003**: The `resolveWorkspace` / `historyMode` seam MUST be widened to receive the config path (both `renderHistoryList` and `renderNewSession` callers updated); `-l` and prompt-less `--new` MUST share the same resolution.

---

### User Story 2 - a `-t` flag prints the session's `turns.log` (Priority: P2)

As an **orchestrator agent**, I want `tellme -t -c "<target>.yaml"` to print the target session's **`turns.log`** and exit, so the SOP's optional monitoring step (`tell-me-go -t -c … | tail -5`) has a `tellme` equivalent that honours `-c` like US1. `tellme` gains its own per-session `turns.log` — the rendered turn chrome — written on the turn path.

**Why this priority**: it is [#103](https://github.com/gosharplite/tellme/issues/103) gap 2 — the issue marks it **lower priority** and a **separate decision**; it does not block the relay loop, but it is part of the operator's tasking.

**Independent verification**: run a turn (a `turns.log` appears under the resolved `output/<mode>/` carrying the turn chrome byte-identically to what the diagnostic stream showed); `tellme -t -c "<file>.yaml"` prints that file for the `-c`-resolved session and exits 0; a different `-c` prints the other session's log.

**Acceptance Scenarios**:

1. **Given** a turn has run for a session, **When** `output/<mode>/turns.log` is inspected, **Then** it holds the session's rendered turn chrome (the turn rule, the `╭─⠿ Turn …` header, the `[HH:MM:SS] Payload: …` lines, the `[HH:MM:SS] [<provider>] M: …` metrics line) in the same plain text the diagnostic stream emitted (C1 — tellme's own chrome).
2. **Given** a session with a non-empty `turns.log`, **When** `tellme -t -c "<file>.yaml"` runs, **Then** it writes that file's contents to stdout verbatim and exits success.
3. **Given** `-t` with a `-c` whose mode differs from the default, **When** it runs (`TELL_ME_MODE` unset), **Then** it reads the `-c` session (US1 resolution), not the default.
4. **Given** `tellme --new -c "<file>.yaml"` from a session with a `turns.log`, **When** it runs, **Then** the `turns.log` is **archived** with the session (alongside `history`/`tokens`), leaving a fresh session.
5. **Given** an empty/absent `turns.log` for the resolved session, **When** `-t` runs, **Then** it exits success without error (offline tolerance, mirroring `-l`).

**Functional Requirements**:

- **FR-004**: `tellme` MUST write a per-session **`turns.log`** at `output/<mode>/turns.log` on the turn path, recording the session's **rendered turn chrome** (C1: tellme's existing plain-text turn/status/metrics lines — the formatter is the existing renderer, no new format). The write site/seam is an RD decision (`research.md`).
- **FR-005**: A new flag `-t` / `--turns` MUST be registered, documented in the `--help` usage, and handled as an **offline reporting command** in the `dispatchReporting` precedence order; it MUST print `output/<mode>/turns.log` verbatim to stdout and exit success (a missing/empty file is success).
- **FR-006**: `-t` MUST resolve the session with the **same** `-c`-honouring, offline mode resolution as US1 (FR-001/FR-002); `--new` MUST archive `turns.log` alongside the session's history and usage logs.
- **FR-007**: `turns.log` MUST be plain text on the diagnostic stream's model (no ANSI/control bytes) — sanitized, mirroring the round-038 `[Tool Output]` discipline — so `-t` output is safe to pipe.

---

### Global Requirements *(cross-story only)*

- **FR-008**: `-l`, prompt-less `--new`, and `-t` MUST keep the offline session-command contract: no configuration/provider **resolution** beyond the mode read, no network, no `TELL_ME_MODE` override regression.
- **FR-009**: When `-c` is **explicitly given** and cannot be read/parsed, `-l` / prompt-less `--new` / `-t` MUST **fail** with an existing config-error phrase and a non-zero exit (Q2 → (A)); the failure MUST NOT widen the frozen `tellme: {phrase}` stderr vocabulary. When **no** `-c` is given, the tolerant default fallback MUST be preserved (round-007).
- **FR-010**: A new **ADR** MUST record the round's decisions (the mode-resolution precedence; the `turns.log` artifact + write site; the explicit-`-c` policy per Q2), indexed in `docs/decisions/README.md`; `specs/truth/techstack.md`, the data model (`specs/truth/data/data-model.dbml`), and the CLI interface truth (`specs/truth/features/cli/**`) updated through `truth-delta.md`.
- **FR-011**: The round MUST introduce **no** new dependency (`go.mod`/`go.sum` unchanged).
- **FR-012**: Falsifiability witnesses MUST be reproduced then reverted (ADR 0010): (a) revert `-c`'s mode read ⇒ the differential retrieve fails (identical output for two configs); (b) revert `turns.log`'s write ⇒ `-t` finds no file; (c) revert `-t`'s `-c` resolution ⇒ `-t` reads the default session.
- **FR-013**: A regression test MUST pin [#103](https://github.com/gosharplite/tellme/issues/103) AC4: seed A (`-c a.yaml`) + B (`-c b.yaml`), then `-l 1 -c a.yaml` returns A's last message and `-l 1 -c b.yaml` returns B's.

---

## Edge Cases

- **`-t` combined with `-l`** — the `dispatchReporting` precedence order (`-d` → `-l` → `--tool-usage`, `cli.go:876-891`) must define where `-t` sits; a combined invocation MUST resolve deterministically (proposed: `-t` joins the offline-session group, ordered after `-l`; recorded in `plan.md`).
- **`-t` + `--new`** — `--new` archives the session (including `turns.log`) **before** the `-t` read resolves; the resulting empty/fresh log is success (documented; `--new` wins the mutation).
- **`turns.log` write failure** — a session whose log cannot be written MUST NOT fail the turn (best-effort, like the usage log); recorded forward.
- **`TELL_ME_MODE` set + `-c`** — env MUST still win (AC2); the explicit-`-c` failure (FR-009) applies only when the mode is actually needed from `-c`.
- **`-c` file absent/invalid** — an **explicit** `-c` that cannot be honoured **fails** with an existing config-error phrase (Q2 → (A)); it MUST NOT silently fall back to the default session.
- **Absent default config (no `-c`)** — `-l`/`--new`/`-t` MUST keep working (round-007 tolerance).
- **`-c` path** — `-c` may be relative or absolute; resolution MUST match the prompt path's handling of `opts.configPath` (whatever `config.Load` accepts).
- **Chrome sanitization** — `turns.log` MUST NOT carry ANSI/control bytes (round-038 discipline); a sanitized line is emitted byte-identically to the diagnostic stream's plain text.

## Key Entities

- **The session-mode resolution** — `historyMode`/`resolveWorkspace` widened to take the config path; precedence `TELL_ME_MODE` → `-c` `MODE` → default → `butler`.
- **`turns.log`** — the new per-session artifact at `output/<mode>/turns.log`, holding the rendered turn chrome (C1); archived by `--new`.
- **The `-t` flag** — `--turns`; reads `turns.log` for the resolved session.
- **`dispatchReporting`** — the offline reporting precedence order (`-d` → `-l` → `--tool-usage`), plus the new `-t`.
- **The ADR** — the round's decision record (mode precedence + the `turns.log` artifact/write site + explicit-`-c` policy).

## Success Criteria

- **SC-001**: With `TELL_ME_MODE` unset, `-l 1 -c a.yaml` and `-l 1 -c b.yaml` return their **distinct** sessions' last messages ([#103](https://github.com/gosharplite/tellme/issues/103) AC1/AC4).
- **SC-002**: `-l`/`--new`/`-t` remain **offline** (no provider resolution, no network) and the env override still wins ([#103](https://github.com/gosharplite/tellme/issues/103) AC2/AC3).
- **SC-003**: A turn writes `output/<mode>/turns.log` carrying the rendered chrome; `-t` prints that file for the `-c`-resolved session and exits success; `--new` archives it.
- **SC-004**: `make verify` green (RULE-A/B/C 0; RULE-E baseline 0, 0 new/0 stale; 0 cycles; lint 0; cross-compile 4/4) · `go test -count=1 ./...` green (incl. the godog E2E).
- **SC-005**: `go.mod`/`go.sum` unchanged; the topology/DSL audit green; the frozen `tellme: {phrase}` vocabulary unchanged.

## Assumptions

- **A1**: The `-c`-honouring fix is a single, shared `resolveWorkspace`-family change (US1, prompt-less `--new`, and `-t` all use it).
- **A2**: `-l` remains a **message list** (`role: content` lines, `toMessages`) — unchanged output shape; only the **session** changes.
- **A3**: `turns.log` is written **per session** (always, like `history`/`tokens` — not gated), appended on the turn path; a write failure is best-effort (never fails the turn). The **sink seam and write site** are RD decisions (`research.md`); the **content = tellme's existing rendered chrome** (Q1 C1: no new format).
- **A4**: `turns.log` is a **persisted session artifact** → `specs/truth/data/data-model.dbml` gains it, and `--new` archives it alongside `history`/`tokens`.
- **A5**: `/axb-api-plan` is **NOOP** (no API surface); `/axb-data-plan` is **MODIFY** (the `turns.log` artifact); `/axb-spec-by-example` is **invoked** (a user-facing new flag + a corrected session selection); `/axb-dsl-refine` adds the CLI Examples/`DSLRow`s for the `-t` read and the `-l -c` selection.
- **A6**: The ADR number is the next free one after 0021 (likely **0022**).

---

## Out of scope (recorded forward items)

- **Bare `-l` defaulting to `1`** (the `tell-me-go` `NoOptDefVal="1"` + `sanitizeArgs` parity) — **not** part of [#103](https://github.com/gosharplite/tellme/issues/103); the SOP already passes `-l 1`. Recorded as a forward item only.
- The **self-diagnosing retrieve** ([#103](https://github.com/gosharplite/tellme/issues/103) "also consider": print the resolved session/mode to stderr, or a `--json` listing) — a candidate, not required; decide in the round or record forward.
- Changing the **prompt** path (it already honours `-c`); making `-l` network-capable; the `-i` terminal requirement.
- Settled exclusions: no Windows, no security/consent layer, no conversation pruning, no tool-call concurrency.
