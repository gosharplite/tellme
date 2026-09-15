# tellme — Status

**Last updated**: 2026-09-15 (day close, session 9: **#47 closed `not_planned`** — decision recorded in [`techstack.md`](specs/truth/techstack.md) → *Not Introduced Yet*) — **round 026 `026-tool-usage-accounting` DELIVERED / FROZEN**: PR [#57](https://github.com/gosharplite/tellme/pull/57) **MERGED** into `dev` (`9d62379`, by `gosharplite`, 2026-09-15T11:15:01Z); round-026 head frozen at **`672a226`**; propagated `dev → main` (no-ff). Round-025 detail relocated to the archive (Rule 12); rounds 001–024 live in the archives.
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 026 delivered; the next round starts a fresh `027-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/15/session-summary.md`](docs/session-summary/2026/09/15/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–025).

## Round 026 — `026-tool-usage-accounting` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-15) — PR [#57](https://github.com/gosharplite/tellme/pull/57) **MERGED** into `dev` (`9d62379`, by `gosharplite`, 2026-09-15T11:15:01Z); round-026 head frozen at **`672a226`** (10 commits); propagated `dev → main` (no-ff). Anchor issue [#53](https://github.com/gosharplite/tellme/issues/53) — a **measurement** slice: count per-tool invocations + outcome (`ok`/`error`/`timeout`) to see which tools the AI uses and which fail.

**Scope**: per-tool accounting for the agent tool surface — classify each **executed** invocation `ok`/`error`/`timeout` from the loop's **structural** signals only (`err` + the per-call `ctx` deadline; **no result-text sniffing**), persist to a **user-global, append-only** `~/.tellme/tools-count.jsonl` (**never reset by `--new`**; best-effort), and surface an **offline `--tool-usage` report** (**`--version`-class**: no `-c`, no `TELL_ME_HOME`, no workspace). Operator-locked: **Q1 → 1** (three-way outcome) · **Q2 → Others** (global JSONL log) · **Q3 → 1** (dedicated offline report).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · data-plan ✅ · dsl-refine ✅ · tasks ✅ · **implement ✅** (T001–T021 all `[X]`).

**Artifacts**: `spec.md` (US1–US2 · FR-001–012 · NFR-001–004), `checklists/requirements.md` (ready), `features/acceptance/*.feature` (2 journeys), `research.md` (D1–D6 + forward notes), `plan.md` (2 interfaces → `/axb-dsl-refine` + `/axb-data-plan`; api NOOP; ui skipped), `tasks.md` (T001–T021), `truth-delta.md`. **Truth**: `techstack.md` MODIFY; `data/data-model.dbml` ADD (`tool_usage_record` + `tool_usage_outcome`); `chat/accounting-for-the-tool-use.feature` ADD; `chat/dsl.md` MODIFY (+9 rows + note); root `cli/dsl.md` MODIFY (`the runtime home is not set` promoted); `workspace/dsl.md` MODIFY (row removed). **Code**: `internal/domain/history/tool_usage.go` (**NEW** — the `ToolUsageSink` port + the `ToolUsageReader`/`ToolUsageStore` read+write port + `ToolUsageCounts`), `internal/agent/agentloop.go` (classification seam + sink), `internal/infrastructure/history/tool_usage_store.go` (**NEW** adapter), `internal/ui/toolusage.go` (**NEW** formatter), `internal/cli/cli.go` (`--tool-usage` + `dispatchReporting` + home seam).

**Verification**: topology audit **PASSED** (39 features · 16 root + **245** module rows · **1262** steps). `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `golangci-lint` 0 issues · `govulncheck` clean); `go test -count=1 ./...` green (unit + godog E2E). **Falsifiability witnesses** reproduced then reverted — (a) classification · (b) `--new` non-reset · (c) report footprint.

**Review trail (PR #57)**: plan+truth **APPROVED WITH REQUIRED FOLDS** → fold `64a9fa9` (8 findings) → **FOLD ACCEPTED** (5 nits) → fold `3170588` → **FINAL ARCHITECTURAL APPROVAL**. Implementation `aa6a8dd` → **APPROVED WITH NON-BLOCKING FOLDS** (A–E) → fold `fb660ff` → **FINAL APPROVAL — CERTIFIED READY TO MERGE** → coverage `bd758f2` (final-review micro-note). **Principal-architect review** (independent second review) on `5a87d27` → **FULL ARCHITECTURAL APPROVAL** (DEBT port-asymmetry · REFACTOR map-allocation-guard) → fold **`672a226`** (domain reader port + allocation guard, unit-pinned) → **RE-CERTIFIED READY TO MERGE**. `stdout` byte-exact; vocabulary 11.

**Commits**: `31a72de` (plan package + truth) · `64a9fa9` (plan+truth folds) · `3170588` (re-review nits) · `aa6a8dd` (implementation) · `fb660ff` (implementation folds A–E) · `bd758f2` (read-error coverage) · `38a021b` (day close) · `5a87d27` (closeout Step 8) · `672a226` (principal-architect fold) · `9d62379` (PR #57 merge into `dev`).

**Propagation**: `026-tool-usage-accounting → dev` (PR [#57](https://github.com/gosharplite/tellme/pull/57), `9d62379`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`. **Also**: anchor issue [#53](https://github.com/gosharplite/tellme/issues/53) **closed (completed)**.

---

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001 | `001-cli-bootstrap-and-config` | PR [#6](https://github.com/gosharplite/tellme/pull/6) |
| 002 | `002-followup-cleanups` | PR [#7](https://github.com/gosharplite/tellme/pull/7) |
| 003 | `003-provider-registry-completeness` | PR [#11](https://github.com/gosharplite/tellme/pull/11) |
| 004 | `004-first-reasoning-turn` | PR [#12](https://github.com/gosharplite/tellme/pull/12) |
| 005 | `005-stdin-piping` | PR [#16](https://github.com/gosharplite/tellme/pull/16) |
| 006 | `006-rendered-output-and-raw-flag` | PR [#19](https://github.com/gosharplite/tellme/pull/19) |
| 007 | `007-session-history-persistence` | PRs [#20](https://github.com/gosharplite/tellme/pull/20)/[#21](https://github.com/gosharplite/tellme/pull/21) |
| 008 | `008-agent-tools-and-tool-call-loop` | PRs [#24](https://github.com/gosharplite/tellme/pull/24)/[#25](https://github.com/gosharplite/tellme/pull/25) |
| 009 | `009-payload-status-line` | PRs [#26](https://github.com/gosharplite/tellme/pull/26)/[#27](https://github.com/gosharplite/tellme/pull/27) |
| 010 | `010-stream-ordering-observability` | PR [#29](https://github.com/gosharplite/tellme/pull/29) |
| 011 | `011-persona-and-payload-estimate` | PR [#30](https://github.com/gosharplite/tellme/pull/30) |
| 012 | `012-interactive-multiline-prompt` | PR [#31](https://github.com/gosharplite/tellme/pull/31) |
| 013 | `013-vertex-gemini-provider` | PR [#33](https://github.com/gosharplite/tellme/pull/33) |
| 014 | `014-session-replay-fidelity` | PR [#35](https://github.com/gosharplite/tellme/pull/35) |
| 015 | `015-interactive-tui-prompt` | PR [#38](https://github.com/gosharplite/tellme/pull/38) |
| 016 | `016-interactive-prompt-visual-parity` | PR [#40](https://github.com/gosharplite/tellme/pull/40) |
| 017 | `017-turn-chrome-parity` | PR [#41](https://github.com/gosharplite/tellme/pull/41) |
| 018 | `018-post-turn-status-lines` | PR [#43](https://github.com/gosharplite/tellme/pull/43) |
| 019 | `019-turn-spinner` | PR [#44](https://github.com/gosharplite/tellme/pull/44) |
| 020 | `020-cross-compile-gate` | PR [#46](https://github.com/gosharplite/tellme/pull/46) |
| 021 | `021-tool-surface-parity` | PR [#48](https://github.com/gosharplite/tellme/pull/48) |
| 022 | `022-tool-loop-log-line` | PR [#50](https://github.com/gosharplite/tellme/pull/50) |
| 023 | `023-interactive-prompt-teardown` | PR [#51](https://github.com/gosharplite/tellme/pull/51) |
| 024 | `024-tool-resource-contract-and-execute-command` | PR [#54](https://github.com/gosharplite/tellme/pull/54) |
| 025 | `025-spinner-width-safety` | PR [#56](https://github.com/gosharplite/tellme/pull/56) |
| 026 | `026-tool-usage-accounting` | PR [#57](https://github.com/gosharplite/tellme/pull/57) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–025 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md)); 026 stays here as the most recent delivered round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `026-tool-usage-accounting` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| (next) `027-*` | not started | The next round's working branch off `dev` (per the branch convention). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 020):** `020-cross-compile-gate → dev` (PR [#46](https://github.com/gosharplite/tellme/pull/46), `642583b`) `→ main` — DONE (no-ff).
> **Propagation (round 021):** `021-tool-surface-parity → dev` (PR [#48](https://github.com/gosharplite/tellme/pull/48), `3877053`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 022):** `022-tool-loop-log-line → dev` (PR [#50](https://github.com/gosharplite/tellme/pull/50), `05278a5`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 023):** `023-interactive-prompt-teardown → dev` (PR [#51](https://github.com/gosharplite/tellme/pull/51), `97e36c6`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 024):** `024-tool-resource-contract-and-execute-command → dev` (PR [#54](https://github.com/gosharplite/tellme/pull/54), `a59ccad`, merged by `thptcnec`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.
> **Propagation (round 025):** `025-spinner-width-safety → dev` (PR [#56](https://github.com/gosharplite/tellme/pull/56), `a6fb921`, merged by `thptcnec`); `dev → main` — **DONE (no-ff)**; closeout docs on `dev`.
> **Propagation (round 026):** `026-tool-usage-accounting → dev` (PR [#57](https://github.com/gosharplite/tellme/pull/57), `9d62379`, merged by `gosharplite`); `dev → main` — **DONE (no-ff)**; closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Slice **024** ([#52](https://github.com/gosharplite/tellme/issues/52), **delivered**) = the tool resource contract + `execute_command` + reader retrofit; slice **025** ([#55](https://github.com/gosharplite/tellme/issues/55), **delivered**) = the spinner-label/bounded-clear fix; slice **026** ([#53](https://github.com/gosharplite/tellme/issues/53), **delivered**) = tool-usage accounting.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–023** | — | Provider-registry completeness → … → the `-i` teardown & submit-surface parity. | ✅ **Delivered** (see the delivered-rounds index) |
| **023 interactive prompt teardown** | [#51](https://github.com/gosharplite/tellme/pull/51) | Clear the `-i` editor frame on submit/abort and resume the standard turn surface (echoed prompt + chrome + spinner + post-turn status). | ✅ **Delivered** — round 023 (PR [#51](https://github.com/gosharplite/tellme/pull/51)) |
| **024 tool resource contract + `execute_command` + reader retrofit** | [#52](https://github.com/gosharplite/tellme/issues/52) | The cross-cutting **token bound + timeout** contract (uniform per-tool params: **default + param + ceiling**, loop-enforced; bound from the **effective budget** = `min(MAX_HISTORY_TOKENS, the model's configured window)` → default `÷4`, ceiling `÷2`) + **add `execute_command`** (bash-first `bash -c`, bounded, **process-group** timeout, no `pipe_commands`, no security) + **retrofit the readers** (whole-file reads; fixed 1 MiB/100000 caps retired). Issue [#49](https://github.com/gosharplite/tellme/issues/49) resolved (**config-gated**). | ✅ **Delivered** — round 024 (PR [#54](https://github.com/gosharplite/tellme/pull/54)) |
| **025 spinner width-safety** | [#55](https://github.com/gosharplite/tellme/issues/55) | The spinner tool-phase label enumerates every tool name → over-wide line (clipped CPU/MEM; a wrapped frame defeats the single-row teardown clear). Fix: **bound** the several-tool label (` Executing tools [<first> and <N-1> more]...`) + a **width-safe clear** (row-aware erase). | ✅ **Delivered** — round 025 (PR [#56](https://github.com/gosharplite/tellme/pull/56), `a6fb921`); issue [#55](https://github.com/gosharplite/tellme/issues/55) **closed** |
| **026 tool-usage accounting** | [#53](https://github.com/gosharplite/tellme/issues/53) | Count per-tool **invocations + outcome** (`ok`/`error`/`timeout`) across all sessions in a **user-global** log (`~/.tellme/tools-count.jsonl`; never reset by `--new`), surfaced by the offline **`--tool-usage`** report — the empirical instrument for pruning the surface with evidence. | ✅ **Delivered** — round 026 (PR [#57](https://github.com/gosharplite/tellme/pull/57), `9d62379`); issue [#53](https://github.com/gosharplite/tellme/issues/53) **closed** |
| **future slices (candidates)** | [#13](https://github.com/gosharplite/tellme/issues/13) | **Coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). Plus the carried forward items below. ~~Concurrent tool-call matching [#47](https://github.com/gosharplite/tellme/issues/47)~~ — **CLOSED `not_planned`** (2026-09-15); rationale in [`specs/truth/techstack.md`](specs/truth/techstack.md) → *Not Introduced Yet*. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-15, reconciled post-closeout)** — **#53** closed (completed — round 026, PR [#57](https://github.com/gosharplite/tellme/pull/57) `9d62379`); **#47** closed (`not_planned`) — the decision is recorded in the truth tree: [`specs/truth/techstack.md`](specs/truth/techstack.md) → *Not Introduced Yet*; **#13** open (future candidate, still accurate).
- **Round-026 forward items** — the slice landed clean; recorded in `research.md`: (1) **recoverable inline failures count as `ok`** (`read_files` renders a missing/unreadable file as a nil-error inline `ERROR: …` → recorded `ok`; correct by the operator-locked Q1, but a blind spot for "which tools fail" — a future `error` axis could admit a tool-declared recoverable-failure signal); (2) **pruning erases the prune signal** (the report enumerates the live registry, so a later-removed tool drops its historical counts — a future "no longer registered" line for log-only names would preserve it); (3) the **`~/.tellme/tools-count.jsonl` log grows unbounded** (compaction deferred — the report streams O(tools) memory, so scan cost is linear). **PR [#57](https://github.com/gosharplite/tellme/pull/57) merged; propagation DONE.**
- **Round-025 forward item** — the fix landed clean; the **mid-frame-resize over-erase bound** (TD-3) is recorded (accepted). **FD-1**: the narrow-terminal residue witness depends on the tool-phase frame wrapping at the forced width (fails loudly, not vacuously) — the durable answer is a wider scripted batch. (Round 025 detail relocated to [`2026-09-15.md`](docs/archives/status/2026-09-15.md).)
- **Round-023 forward item** — none new; the teardown + echo landed clean. (The reference echoes the prompt via its `provideFeedback` two-line `Input captured:` form; tellme keeps its single captured line **plus** a verbatim echo block — a recorded divergence. The `tellme: `-leading-prompt residual on the diagnostic stream is recorded in `research.md` + the echo DSL row.)
- **Round-022 forward item** — none new; the reshape + separation landed clean. (The reference's `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Result]` decomposition remains a recorded divergence — tellme uses the single compact `[Tool] <name> - <reason>` line.)
- **Round-024 forward items** — the recorded divergences/limitations: the `CONTEXT_WINDOW` bound is **config-gated** (model-blind without a window); `output_file` redirects **both** streams (no inline preview); the sequential-tools worst-case wall-clock (`MAX_TOOL_LOOP × 7200 s` ceiling; the closed [#47](https://github.com/gosharplite/tellme/issues/47) would have changed the shape); the `ProcessRunner` port extraction trigger ("the first **write** tool, or any second managed-process consumer"). New bug/candidate: **[#55](https://github.com/gosharplite/tellme/issues/55)** (spinner label over-width → clipped CPU/MEM + wrap defeats the teardown clear; round-019 territory) — ✅ **resolved / closed by round 025**.
- **Round-021 forward item** — ✅ **resolved / closed by round 024**: issue [#49](https://github.com/gosharplite/tellme/issues/49) (reader cap → `MAX_HISTORY_TOKENS`) is now **closed (completed)** — the fixed 1 MiB/100000 caps are retired (config-gated `CONTEXT_WINDOW`).
- **Round-020 forward item** — none new; the cross-compile gate is delivered. (Watch: if a supported target ever needs cgo, the `CGO_ENABLED=0` pin must be revisited.)
- **Round-019 forward items** — the failed-turn carrier proves *absence* (mid-wait *clear-before-the-class-phrase* deferred); the macOS **CPU** leg is pending a cgo `mach` sampler and reports `0.0%` (the memory leg uses sysctl).
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines (plain text); the `tokens.summary.json` roll-up is best-effort (self-heals by recompute).
- **Future-slice candidates** — **(d)** coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13); **(e)** the renderer/`-r` forward items. (Concurrent tool-call matching [#47](https://github.com/gosharplite/tellme/issues/47) closed `not_planned` — see *Not Introduced Yet* in [`techstack.md`](specs/truth/techstack.md).)
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN** (round 019 did **not** close it — the spinner is a `stderr` diagnostic); round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (estimation-heuristic constants; persona seam; **N-2**); a future **`history.Store.Count()`** should replace `len(prior)+1`.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (refreshed round 026 — carries the `--tool-usage` offline report + the domain reader port). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Host (this workspace)**: **darwin/arm64** — this session runs on the MacBook Pro niffler env (`…/mbp-johndoe-niffler/ait-tellme`); `go env` reports `darwin/arm64`, so `make verify` compiles macOS natively and the **Linux** path is the cross-compile weak spot here (round 020). The other dev host (`…/beta-niffler/`) is Linux and mirrors the opposite.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`); the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree; read for `/Users/johndoe/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
