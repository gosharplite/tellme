# Feature Specification: close [#103](https://github.com/gosharplite/tellme/issues/103) — the offline session commands honour `-c` (`-l`, prompt-less `--new`) + a `-t` (turns-log) flag (round 053)

**Feature Branch**: `053-offline-session-config-and-turns-flag`

**Created**: 2026-09-19

**Status**: Draft — **clarify round 1 OPEN** (Q1 = the `-t` semantics; Q2 = `-c` load-failure behaviour). Produced by `/axb-specify`.

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

`$TELL_ME_HOME/output/<mode>/` = `history.jsonl`, `history.archive.jsonl`, `tokens.log`, `tokens.archive.jsonl`, `tokens.summary.json` — **there is no `turns.log`** in `tellme`. The token/usage log (`tokens.log`, round-026) and the persisted turn record (`history.jsonl`) are the two candidate carriers for a `tellme` "turns log". **Which one `-t` prints is clarify Q1.**

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
| **Q1** | **What does `-t` print?** `tellme` has no `turns.log`; the reference's `-t` prints its own detailed turn trace. | **(A)** Print the session's **usage/token log** (`tokens.log`, the round-026 artifact) — cheapest, exact file parity with the SOP's monitoring intent. **(B)** Print a **turn trace** derived from `history.jsonl` (e.g. one `[HH:MM:SS]`-style block per persisted turn) — closest to the reference's "turns log", but a new projection with its own format decisions. **(C)** Add a **new `turns.log`** written per turn (like the reference) and have `-t` print it — largest scope. | **OPEN** |
| **Q2** | **`-c` load failure on an offline session command**: when `-c <file>` is given but the file is missing/invalid, should `-l`/`--new` **fail** (like the prompt path's `reasonConfigMissing`/`reasonConfigInvalid`) or **fall back** to the default/`butler` (today's tolerant behaviour)? | **(A)** Fail on an **explicit** `-c` (the user named it); still tolerate an **absent default**. **(B)** Always fall back (status quo tolerance). | **OPEN** |

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

### User Story 2 - a `-t` flag prints the session's turns/trace log (Priority: P2)

As an **orchestrator agent**, I want `tellme -t -c "<target>.yaml"` to print the target session's turns/trace log and exit, so the SOP's optional monitoring step (`tell-me-go -t -c … | tail -5`) has a `tellme` equivalent that honours `-c` like US1.

**Why this priority**: it is [#103](https://github.com/gosharplite/tellme/issues/103) gap 2 — the issue itself marks it **lower priority** and a **separate decision**; it does not block the round's relay loop, but it is part of the operator's tasking.

**Independent verification**: seed a session; `tellme -t -c "<file>.yaml"` prints the chosen carrier's contents for that session's `output/<mode>/` and exits 0; running it with a different `-c` prints the other session's log.

**Acceptance Scenarios**:

1. **Given** a session with a non-empty turns/trace log, **When** `tellme -t -c "<file>.yaml"` runs, **Then** it writes that session's log to stdout and exits success. **(exact content/format = clarify Q1)**
2. **Given** `-t` with a `-c` whose mode differs from the default, **When** it runs (`TELL_ME_MODE` unset), **Then** it reads the `-c` session (US1 resolution), not the default.
3. **Given** an empty/absent log for the resolved session, **When** `-t` runs, **Then** it exits success without error (offline tolerance, mirroring `-l`).

**Functional Requirements**:

- **FR-004**: A new flag `-t` / `--turns` MUST be registered, documented in the `--help` usage, and handled as an **offline reporting command** in the `dispatchReporting` precedence order. **(the log source/format = clarify Q1)**
- **FR-005**: `-t` MUST resolve the session with the **same** `-c`-honouring, offline mode resolution as US1 (FR-001/FR-002).

---

### Global Requirements *(cross-story only)*

- **FR-006**: `-l`, prompt-less `--new`, and `-t` MUST keep the offline session-command contract: no configuration/provider **resolution** beyond the mode read, no network, no `TELL_ME_MODE` override regression.
- **FR-007**: The change MUST NOT widen the frozen `tellme: {phrase}` stderr vocabulary; any new diagnostic (if Q2 → fail) MUST reuse an existing phrase.
- **FR-008**: A new **ADR** MUST record the round's decisions (the mode-resolution precedence; the `-t` carrier per Q1; the explicit-`-c` failure policy per Q2), indexed in `docs/decisions/README.md`; `specs/truth/techstack.md` + the CLI interface truth (`specs/truth/features/cli/**`) updated through `truth-delta.md`.
- **FR-009**: The round MUST introduce **no** new dependency (`go.mod`/`go.sum` unchanged).
- **FR-010**: Falsifiability witnesses MUST be reproduced then reverted (ADR 0010): (a) revert `-c`'s mode read ⇒ the differential retrieve fails (identical output for two configs); (b) revert `-t`'s `-c` resolution ⇒ `-t` reads the default session.
- **FR-011**: A regression test MUST pin [#103](https://github.com/gosharplite/tellme/issues/103) AC4: seed A (`-c a.yaml`) + B (`-c b.yaml`), then `-l 1 -c a.yaml` returns A's last message and `-l 1 -c b.yaml` returns B's.

---

## Edge Cases

- **`-t` combined with `-l`** — the `dispatchReporting` precedence order (`-d` → `-l` → `--tool-usage`, `cli.go:876-891`) must define where `-t` sits; a combined invocation MUST resolve deterministically (proposed: `-t` joins the offline-session group; precedence recorded at implementation — see Q1/`plan.md`).
- **`-t` + `--new`** — archiving interacts with printing; define whether they compose or `--new` wins.
- **`TELL_ME_MODE` set + `-c`** — env MUST still win (AC2).
- **`-c` file absent/invalid** — behaviour per clarify Q2; must not silently reintroduce the gap-1 failure.
- **Absent default config** — `-l`/`--new` MUST keep working (round-007 tolerance).
- **`-c` path** — `-c` may be relative or absolute; resolution MUST match the prompt path's handling of `opts.configPath` (whatever `config.Load` accepts).

## Key Entities

- **The session-mode resolution** — `historyMode`/`resolveWorkspace` widened to take the config path; precedence `TELL_ME_MODE` → `-c` `MODE` → default → `butler`.
- **The `-t` flag** — `--turns`; its carrier (per Q1) under `output/<mode>/`.
- **`dispatchReporting`** — the offline reporting precedence order (`-d` → `-l` → `--tool-usage`), plus the new `-t`.
- **The ADR** — the round's decision record (mode precedence + `-t` carrier + explicit-`-c` policy).

## Success Criteria

- **SC-001**: With `TELL_ME_MODE` unset, `-l 1 -c a.yaml` and `-l 1 -c b.yaml` return their **distinct** sessions' last messages ([#103](https://github.com/gosharplite/tellme/issues/103) AC1/AC4).
- **SC-002**: `-l`/`--new` remain **offline** (no provider resolution, no network) and the env override still wins ([#103](https://github.com/gosharplite/tellme/issues/103) AC2/AC3).
- **SC-003**: `-t` exists, is documented, prints the Q1-chosen carrier for the `-c`-resolved session, and exits success.
- **SC-004**: `make verify` green (RULE-A/B/C 0; RULE-E baseline 0, 0 new/0 stale; 0 cycles; lint 0; cross-compile 4/4) · `go test -count=1 ./...` green (incl. the godog E2E).
- **SC-005**: `go.mod`/`go.sum` unchanged; the topology/DSL audit green; the frozen `tellme: {phrase}` vocabulary unchanged.

## Assumptions

- **A1**: The `-c`-honouring fix is a single, shared `resolveWorkspace`-family change (US1 and `-t` use it).
- **A2**: `-l` remains a **message list** (`role: content` lines, `toMessages`) — unchanged output shape; only the **session** changes.
- **A3**: The fix touches only `internal/cli` (resolution wiring) + the persistence read path; **no** `internal/domain/**` behaviour change beyond any port widening.
- **A4**: `/axb-api-plan` and `/axb-data-plan` are **NOOP** (no API surface; no persisted-schema change — the carrier read is read-only).
- **A5**: `/axb-spec-by-example` is **invoked** (a user-facing, observable behaviour change: a new flag + a corrected session selection). Whether `/axb-dsl-refine` adds new `DSLRow`s (a new `-t` Example; an `-l -c` selection Example) is an RD decision at `/axb-system-analysis`/`/axb-dsl-refine`.
- **A6**: The ADR number is the next free one after 0021.

---

## Out of scope (recorded forward items)

- **Bare `-l` defaulting to `1`** (the `tell-me-go` `NoOptDefVal="1"` + `sanitizeArgs` parity) — **not** part of [#103](https://github.com/gosharplite/tellme/issues/103); the SOP already passes `-l 1`. Recorded as a forward item only.
- The **self-diagnosing retrieve** ([#103](https://github.com/gosharplite/tellme/issues/103) "also consider": print the resolved session/mode to stderr, or a `--json` listing) — a candidate, not required; decide in the round or record forward.
- Changing the **prompt** path (it already honours `-c`); making `-l` network-capable; the `-i` terminal requirement.
- Settled exclusions: no Windows, no security/consent layer, no conversation pruning, no tool-call concurrency.
