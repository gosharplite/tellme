# tellme — Status

**Last updated**: 2026-09-15 (day close, session 10: **round 027 `027-ai-call-turn-counter` DELIVERED / FROZEN** — PR [#58](https://github.com/gosharplite/tellme/pull/58) **MERGED** into `dev` (`ecd4d44`, by `gosharplite`, 2026-09-15T13:35:59Z); round-027 head frozen at **`87643d2`** (9 commits); propagated `dev → main` (no-ff)). Round-026 detail relocated to the archive (Rule 12); rounds 001–026 live in the archives.
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 027 delivered; the next round starts a fresh `028-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/15/session-summary.md`](docs/session-summary/2026/09/15/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026).

## Round 027 — `027-ai-call-turn-counter` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-15) — PR [#58](https://github.com/gosharplite/tellme/pull/58) **MERGED** into `dev` (`ecd4d44`, by `gosharplite`, 2026-09-15T13:35:59Z); round-027 head frozen at **`87643d2`** (9 commits); propagated `dev → main` (no-ff). A **presentation-accounting** slice (operator request): the `╭─⠿ Turn <N> - <mode>` header's `<N>` now counts the session's **AI-endpoint calls** (inference rounds), matching `tell-me-go`'s counter value at each prompt's first call.

**Scope**: keep tellme's chrome **as-is** (one header + one `╰─⠿ Ready` per prompt; plain text; no denominator), but redefine `<N>` = **Σ of the session's prior-turn call counts + 1**. A tool-less turn advances it by one; a tool-using turn by its inference-round count; a provider-internal retry does **not** count; `--new` restarts at `Turn 1`. Persist each completed turn's call count on `history_entry` (`calls`; a legacy/field-less line counts as 1). **Operator fold**: `--new` now also archives on the **`-i`** TUI surface (it previously dropped `--new`, so `tellme --new -i` opened at the un-archived history's count — `Turn 2` instead of `Turn 1`).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · data-plan ✅ · dsl-refine ✅ · tasks ✅ · **implement ✅** (T001–T014 all `[X]`).

**Artifacts**: `spec.md` (US1–US2 · FR-001–007 · NFR-001–004 · SC-001–004), `checklists/requirements.md` (ready), `features/acceptance/*.feature` (2 journeys), `research.md` (D1–D6 + residual risks), `plan.md` (2 interfaces → `/axb-dsl-refine` + `/axb-data-plan`; api NOOP; ui skipped), `tasks.md` (T001–T014), `truth-delta.md`. **Truth**: `techstack.md` MODIFY (Turn chrome `<N>` + Session history store `calls`); `data/data-model.dbml` MODIFY (`history_entry.calls`); `chat/presenting-the-turn.feature` MODIFY (counting Rule reshaped + 3 Examples) + `chat/continuing-the-interactive-prompt.feature` MODIFY (the `--new -i` Example) + `chat/dsl.md` MODIFY (header row + 2 arrange Givens + 1 new When + note) + `history/dsl.md` MODIFY; `contracts/**` NOOP. **Code**: `internal/domain/history/history.go` (`Entry.Calls` + `TotalCalls`), `internal/infrastructure/history` (serialize `calls`), `internal/cli/cli.go` (`turnNumber` = `history.TotalCalls(prior)+1`; persist `Calls: len(result.Calls)`; `--new` archives before **both** terminal readers); `internal/ui/turn.go` + `internal/agent/agentloop.go` **untouched**.

**Verification**: topology audit **PASSED** (39 features · 16 root + **246** module rows · **1284** steps). `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `golangci-lint` 0 issues · `govulncheck` clean); `go test -count=1 ./...` green (unit + godog E2E, 179 scenarios). **Falsifiability witnesses** reproduced then reverted — (a) call-vs-turn counter (`Turn 3`/`Turn 2`) · (b) `--new` reset (archive) · (c) `-i` `--new` reset.

**Review trail (PR #58)**: plan+truth **APPROVED WITH REQUIRED FOLDS** → fold `88342be` (the `--new` acceptance Example · `runTurn` call-count assertions · domain `TotalCalls`) → **FINAL ARCHITECTURAL APPROVAL**. Operator-reported defect (`--new -i` opened at `Turn 2`) → fold **`87643d2`** (archive `--new` before both terminal readers; new `-i` When row + Example; witness) → **RE-CERTIFIED — CERTIFIED READY TO MERGE** → **MERGED** `ecd4d44`. `stdout` byte-exact; vocabulary 11.

**Commits**: `75f925c` (plan package + spec + acceptance) · `20cef97` (research + techstack) · `b3ecb75` (system-analysis plan) · `b57abe8` (data truth) · `b348ea5` (CLI interface truth) · `8a0e74b` (tasks) · `f089c03` (implementation) · `88342be` (review fold) · `87643d2` (operator-defect fold) · `ecd4d44` (PR #58 merge into `dev`).

**Propagation**: `027-ai-call-turn-counter → dev` (PR [#58](https://github.com/gosharplite/tellme/pull/58), `ecd4d44`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.

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
| 027 | `027-ai-call-turn-counter` | PR [#58](https://github.com/gosharplite/tellme/pull/58) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md)); 027 stays here as the most recent delivered round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `027-ai-call-turn-counter` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| (next) `028-*` | not started | The next round's working branch off `dev` (per the branch convention). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 024):** `024-tool-resource-contract-and-execute-command → dev` (PR [#54](https://github.com/gosharplite/tellme/pull/54), `a59ccad`, merged by `thptcnec`) `→ main` — **DONE (no-ff)**.
> **Propagation (round 025):** `025-spinner-width-safety → dev` (PR [#56](https://github.com/gosharplite/tellme/pull/56), `a6fb921`, merged by `thptcnec`) `→ main` — **DONE (no-ff)**.
> **Propagation (round 026):** `026-tool-usage-accounting → dev` (PR [#57](https://github.com/gosharplite/tellme/pull/57), `9d62379`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**.
> **Propagation (round 027):** `027-ai-call-turn-counter → dev` (PR [#58](https://github.com/gosharplite/tellme/pull/58), `ecd4d44`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **024** ([#52](https://github.com/gosharplite/tellme/issues/52), delivered) = the tool resource contract + `execute_command`; **025** ([#55](https://github.com/gosharplite/tellme/issues/55), delivered) = spinner width-safety; **026** ([#53](https://github.com/gosharplite/tellme/issues/53), delivered) = tool-usage accounting; **027** (operator request, delivered) = the AI-call turn counter.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–023** | — | Provider-registry completeness → … → the `-i` teardown & submit-surface parity. | ✅ **Delivered** (see the delivered-rounds index) |
| **024 tool resource contract + `execute_command` + reader retrofit** | [#52](https://github.com/gosharplite/tellme/issues/52) | The cross-cutting **token bound + timeout** contract + **add `execute_command`** (bash-first) + **retrofit the readers**. Issue [#49](https://github.com/gosharplite/tellme/issues/49) resolved (**config-gated**). | ✅ **Delivered** — round 024 (PR [#54](https://github.com/gosharplite/tellme/pull/54)) |
| **025 spinner width-safety** | [#55](https://github.com/gosharplite/tellme/issues/55) | **Bound** the several-tool spinner label + a **width-safe (row-aware) clear**. | ✅ **Delivered** — round 025 (PR [#56](https://github.com/gosharplite/tellme/pull/56)); issue [#55](https://github.com/gosharplite/tellme/issues/55) **closed** |
| **026 tool-usage accounting** | [#53](https://github.com/gosharplite/tellme/issues/53) | Count per-tool **invocations + outcome** across all sessions in a **user-global** log (`~/.tellme/tools-count.jsonl`), surfaced by the offline **`--tool-usage`** report. | ✅ **Delivered** — round 026 (PR [#57](https://github.com/gosharplite/tellme/pull/57)); issue [#53](https://github.com/gosharplite/tellme/issues/53) **closed** |
| **027 AI-call turn counter** | — (operator request) | Redefine the `╭─⠿ Turn <N>` header's `<N>` to count the session's **AI-endpoint calls** (inference rounds), matching `tell-me-go`; persist each turn's call count on `history_entry` (`calls`); `--new` restarts at `Turn 1`. **Operator fold:** `--new` also archives on the `-i` surface. | ✅ **Delivered** — round 027 (PR [#58](https://github.com/gosharplite/tellme/pull/58), `ecd4d44`) |
| **future slices (candidates)** | [#13](https://github.com/gosharplite/tellme/issues/13) | **Coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). Plus the carried forward items below. ~~Concurrent tool-call matching [#47](https://github.com/gosharplite/tellme/issues/47)~~ — **CLOSED `not_planned`** (2026-09-15); rationale in [`specs/truth/techstack.md`](specs/truth/techstack.md) → *Not Introduced Yet*. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-15, reconciled post-closeout)** — **#53** closed (completed — round 026, PR [#57](https://github.com/gosharplite/tellme/pull/57)); **#47** closed (`not_planned`) — recorded in [`specs/truth/techstack.md`](specs/truth/techstack.md) → *Not Introduced Yet*; **#13** open (future candidate, still accurate).
- **Round-027 forward items** — (a) **legacy entries count as 1**: a pre-027 resumed session may undercount its tool-using turns (a field-less line counts as one) — acceptable for a dev-stage tool, no migration; (b) a future **summarisation/archive** path that drops entries must preserve the counter (`Σ` over the active entries); (c) the reference's **per-call chrome cadence** remains a recorded divergence — tellme emits one header + one `╰─⠿ Ready` per prompt, not per provider call (only the number's **unit** is aligned). **PR [#58](https://github.com/gosharplite/tellme/pull/58) merged; propagation DONE.**
- **Round-026 forward items** — (1) **recoverable inline failures count as `ok`** (a nil-error inline `ERROR: …` result); (2) **pruning erases the prune signal** (the report enumerates the live registry); (3) the **`~/.tellme/tools-count.jsonl` log grows unbounded** (compaction deferred). (Round-026 detail relocated to [`2026-09-15.md`](docs/archives/status/2026-09-15.md).)
- **Round-025 forward item** — the **mid-frame-resize over-erase bound** (TD-3) is recorded (accepted). **FD-1**: the narrow-terminal residue witness depends on the tool-phase frame wrapping at the forced width.
- **Round-023 forward item** — none new; the teardown + echo landed clean (the reference's two-line `Input captured:` echo vs tellme's single captured line **plus** a verbatim echo block is a recorded divergence).
- **Round-022 forward item** — none new (the reference's decomposed `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Result]` shape remains a recorded divergence).
- **Round-024 forward items** — the `CONTEXT_WINDOW` bound is **config-gated**; `output_file` redirects **both** streams; the `ProcessRunner` port extraction trigger ("the first **write** tool, or any second managed-process consumer").
- **Round-021 forward item** — ✅ **resolved / closed by round 024**: [#49](https://github.com/gosharplite/tellme/issues/49) closed (completed).
- **Round-020 forward item** — none new; the cross-compile gate is delivered. (Watch: if a supported target ever needs cgo, the `CGO_ENABLED=0` pin must be revisited.)
- **Round-019 forward items** — the failed-turn carrier proves *absence*; the macOS **CPU** leg is pending a cgo `mach` sampler (`0.0%`; memory via sysctl).
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines (plain text); the `tokens.summary.json` roll-up is best-effort.
- **Future-slice candidates** — **(d)** coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13); **(e)** the renderer/`-r` forward items.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN** (the spinner is a `stderr` diagnostic); round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (estimation-heuristic constants; persona seam; **N-2**).

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`. The operator's live test env for round 027 is `…/mbp-johndoe-niffler/ait-test/` (tag `test`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (**refreshed round 027** — carries the call-based turn counter + the `-i` `--new` fix). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Host (this workspace)**: **darwin/arm64** — this session runs on the MacBook Pro niffler env (`…/mbp-johndoe-niffler/ait-tellme`); `go env` reports `darwin/arm64`, so `make verify` compiles macOS natively and the **Linux** path is the cross-compile weak spot here (round 020). The other dev host (`…/beta-niffler/`) is Linux and mirrors the opposite.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`); the `…/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree; read for `/Users/johndoe/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
