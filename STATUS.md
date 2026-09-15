# tellme — Status

**Last updated**: 2026-09-15 (day close) — **round 023 `023-interactive-prompt-teardown` DELIVERED / FROZEN**: PR [#51](https://github.com/gosharplite/tellme/pull/51) **MERGED** into `dev` (`97e36c6`, by `thptcnec`, 2026-09-15T03:05:19Z); round-023 head frozen at **`14d7567`**; propagated `dev → main` (no-ff). Round 022 stays delivered/frozen (detail in the archive; Rule 12 — older rounds 001–021 also live in the archives).
**Planning + round 024 (2026-09-15, session 4)**: the design direction (*no security · no Windows · **bash-first** · small surface*) is recorded in [`README.md`](README.md#-design-intent--direction-operator-declared); slices **024** ([#52](https://github.com/gosharplite/tellme/issues/52) — tool resource contract + `execute_command` + reader retrofit) and **025** ([#53](https://github.com/gosharplite/tellme/issues/53) — tool-usage accounting) scoped. **Round 024 plan + truth half COMPLETE and APPROVED** — PR [#54](https://github.com/gosharplite/tellme/pull/54) open → `dev` (head **`9e68598`**): the pipeline (spec → acceptance → research/techstack → plan → interface truth), then a **grill round** (architect vs griller) and an **architecture review**, both folded → **PLAN + TRUTH — APPROVED**; implementation (`/axb-tasks` → `/axb-implement`) pending.
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `024-tool-resource-contract-and-execute-command`
**Daily log**: [`docs/session-summary/2026/09/15/session-summary.md`](docs/session-summary/2026/09/15/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–022).

## Round 024 — `024-tool-resource-contract-and-execute-command` (plan + truth half — in progress)

**Status**: 🔵 **PLAN + TRUTH COMPLETE / APPROVED — IMPLEMENTATION PENDING** (2026-09-15). PR [#54](https://github.com/gosharplite/tellme/pull/54) open → `dev` (branch head **`9e68598`**): the pipeline (spec → acceptance → research/techstack → plan → interface truth), then a **grill round** (architect vs griller — 8 questions, verdict *proceed with changes*) whose fold `7bdeede` and an **architecture review** on `7bdeede` (verdict *approved with required folds* — T1–T4 · R1–R5) whose fold `c6d0366` + nit `9e68598` were folded → **PLAN + TRUTH — APPROVED** ([#5675900803](https://github.com/gosharplite/tellme/pull/54#issuecomment-5675900803), re-review closed at [#5675924254](https://github.com/gosharplite/tellme/pull/54#issuecomment-5675924254)). Topology audit PASSED (38 features · 15 root + **235** module rows · **1194** steps) · no product code yet · `go.mod`/`go.sum` unchanged.

**Scope**: the tool resource contract + a bash-first `execute_command` + a reader retrofit. `execute_command` runs `bash -c` (no `pipe_commands`, no security/consent, no Windows); a non-zero exit is a **successful result** carrying the exit status; a timed-out command is terminated as a **process group** (`Setpgid` + `kill(-pgid)` + `WaitDelay`) and surfaced as a **nil-error timeout result** (never the loop's `error: ` path — **FR-018**, uniform across every tool); optional `output_file`/`append` binds stdout **and** stderr directly to the file. Every agent tool takes uniform `max_output_tokens` + `timeout` (default → param → ceiling), loop-enforced, bounding at the source. The bound derives from the **effective budget** = `min(MAX_HISTORY_TOKENS, the model's configured MODELS.<model>.CONTEXT_WINDOW)` → default `÷4`, ceiling `÷2` (issue #49, config-gated); the token bound is realised as a **byte** budget (contract-owned `bytesPerToken = 4`; default `1000000 B`), clamped on raw length. Readers read files **whole** up to the aggregate bound (fixed 100000 B / 1 MiB retired); a skip marker names unread files.

**Locked decisions**: D1 no security · D2 no Windows · D3 bash-first · D4 small surface · D5 three-tier contract · D6 one aggregate bound · D7 scope; clarify Q1 (non-zero exit = success result) · Q2 (`output_file`/`append` in scope) · Q3 (mechanism locked, numbers in research).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · dsl-refine ✅ · **tasks ✅ (T001–T049; orphan sweep 0)** · implement ⏳ next.

**Artifacts**: `spec.md` (US1–US2 · FR-001–**018** · SC-001–006), `checklists/requirements.md` (ready), `features/acceptance/*.feature` ×3, `research.md` (D1–D8 + D1a), `plan.md` (1 interface → `/axb-dsl-refine`; api/data NOOP), `truth-delta.md`. Truth: `techstack.md` MODIFY (Tool resource contract + Agent command tool + `CONTEXT_WINDOW` row; reader caps retired; grill/review folds); `chat/running-a-shell-command.feature` ADD (+ a process-tree Rule); `chat/offering-the-reader-tools.feature` → `offering-the-agent-tools.feature`; `chat/reading-several-files.feature` + `chat/listing-a-directory.feature` + `chat/surveying-a-folder-tree.feature` MODIFY (byte wording + the list/tree bound witnesses); `chat/dsl.md` (command + bound/process-tree rows + note).

**Review trail (PR #54)**: plan+truth **REQUEST CHANGES** (B1 model-derived bound · B2 process-group · D1–D4 · coverage) → fold `124f345` → residual fold `2fe29cf`; **grill round** (architect ⚔️ griller, 8 Q) → *proceed with changes* (8 shipped-half defects) → fold **`7bdeede`**; **architecture review** on `7bdeede` → *approved with required folds* (T1 trim lifecycle · T2 process-tree witness · T3 displayed budget · T4 reader timeout-result · R1–R5) → fold **`c6d0366`** → re-review **PLAN + TRUTH — APPROVED** → residual-nit **`9e68598`** (payload-line budget pointer) → **APPROVED; ready for human merge**.

**Propagation**: **PENDING** — PR [#54](https://github.com/gosharplite/tellme/pull/54) open → `dev` (plan + truth half); not merged (human-only); `main` unchanged.

## Round 023 — `023-interactive-prompt-teardown` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-15) — PR [#51](https://github.com/gosharplite/tellme/pull/51) **MERGED** into `dev` (`97e36c6`, by `thptcnec`, 2026-09-15T03:05:19Z); round-023 head frozen at **`14d7567`**. `make verify` OK · `go test ./...` green · godog **156/156** (0 undefined) · topology audit PASSED (37 features · 15 root + **216** module rows · **1124** steps) · `go.mod`/`go.sum` unchanged.

**Scope**: after the operator submits (or aborts) the `-i` interactive prompt, the editor frame must **clear** (reference parity) and the run must continue on the **standard turn surface** — the submitted prompt **echoed** (on `stderr`), then the input-capture acknowledgement, the `─` rule + `╭─⠿ Turn N - <mode>` header, the live spinner, and the post-turn status. Reverses the round-016 always-render, the round-017 "no turn chrome", and the round-019 "`-i` excluded" rules.

**Locked decisions**: Q1 → 1a (clear the frame on submit/abort) · Q2 → 2a (resume the standard surface) · Q3 → A′ (echo the prompt **and** keep tellme's single captured line, echo before it). **Round-022 ripple (Option 1)**: the round-022 negative is re-anchored to `the pre-flight payload line is separated from the answer by a single blank line` (the `-i` submit is now a chrome surface).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · analysis ✅ · ui-plan (terminal) ✅ · dsl-refine ✅ · tasks ✅ · implement ✅ · **delivered**.

**Artifacts**: `spec.md` (US1–US2 · FR-001–014 · SC-001–005), `checklists/requirements.md`, `features/acceptance/*.feature` ×2, `research.md` (D1–D8), `plan.md` (1 interface → `/axb-dsl-refine`; api/data NOOP; ui terminal reviewed — not re-planned), `ui/ui-plan.md` + `ui/screens/*.txt` ×3, `tasks.md` (T001–T016; all `[X]`; orphan sweep 0), `truth-delta.md`. Truth: `techstack.md` MODIFY (6 rows); `chat/continuing-the-interactive-prompt.feature` ADD; `presenting-the-turn.feature` MODIFY (delete the `-i` no-chrome rule); `watching-the-tool-loop.feature` / `presenting-the-progress-spinner.feature` MODIFY; `chat/dsl.md` (+4 / −1 rows + notes); root `cli/dsl.md` (2 rows re-scoped). **Implementation**: `internal/ui/tui/prompt/model.go` (clear the editor frame on submit/abort) · `internal/cli/cli.go` (`turnOptions.echo`; `-i` submit → `chrome:true, echo:true`; echo the prompt verbatim before the input-capture line); unit pins (`model_teardown_test.go`, `tui_submit_chrome_test.go`); E2E `step_r023_t004/005/006` + the re-anchored `step_r022_t009` + the terminal-reduction helper.

**Harness note**: bubbletea **coalesces** frames when all keys arrive at once, so the editor box never reached the captured stream — the E2E TUI driver now **synchronizes on the editor's output** (`harness.RunInWithSyncedStdin`: write the compose keys, then the terminal key only once the frame paints — no wall-clock sleep).

**Review trail (PR #51)**: plan+truth **APPROVED** → review-2 verified → fold `98b9a3a` → residual-nit fold `67e077f` → principal review **ARCHITECTURALLY APPROVED** → fold `491800a` → implementation **APPROVED** → fold `5a4f4a8` (deterministic synced handshake + non-vacuous cleared witness + ledger/notes) → fold `14d7567` (`Wait`-ordering) → **FINAL ARCHITECTURAL APPROVAL — IMPLEMENTATION FULLY CERTIFIED & READY TO MERGE** → **MERGED** `97e36c6`.

**Propagation**: `023-interactive-prompt-teardown → dev` (PR [#51](https://github.com/gosharplite/tellme/pull/51), `97e36c6`, merged by `thptcnec`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–022 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md)); 023 stays here as the most recent delivered round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `023-interactive-prompt-teardown` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `024-tool-resource-contract-and-execute-command` | in progress | Round 024 (plan + truth **APPROVED**; PR [#54](https://github.com/gosharplite/tellme/pull/54) open → `dev`; implementation pending). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 020):** `020-cross-compile-gate → dev` (PR [#46](https://github.com/gosharplite/tellme/pull/46), `642583b`) `→ main` — DONE (no-ff).
> **Propagation (round 021):** `021-tool-surface-parity → dev` (PR [#48](https://github.com/gosharplite/tellme/pull/48), `3877053`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 022):** `022-tool-loop-log-line → dev` (PR [#50](https://github.com/gosharplite/tellme/pull/50), `05278a5`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 023):** `023-interactive-prompt-teardown → dev` (PR [#51](https://github.com/gosharplite/tellme/pull/51), `97e36c6`, merged by `thptcnec`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> **Propagation (round 024):** `024-tool-resource-contract-and-execute-command → dev` (PR [#54](https://github.com/gosharplite/tellme/pull/54), head `9e68598`) — **PENDING** (plan + truth half, **approved**; human-only merge; implementation follows).
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Slice **024** ([#52](https://github.com/gosharplite/tellme/issues/52)) = the tool resource contract + `execute_command` + reader retrofit; slice **025** ([#53](https://github.com/gosharplite/tellme/issues/53)) = tool-usage accounting.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–023** | — | Provider-registry completeness → … → the `-i` teardown & submit-surface parity. | ✅ **Delivered** (see the delivered-rounds index) |
| **023 interactive prompt teardown** | [#51](https://github.com/gosharplite/tellme/pull/51) | Clear the `-i` editor frame on submit/abort and resume the standard turn surface (echoed prompt + chrome + spinner + post-turn status). | ✅ **Delivered** — round 023 (PR [#51](https://github.com/gosharplite/tellme/pull/51)) |
| **024 tool resource contract + `execute_command` + reader retrofit** | [#52](https://github.com/gosharplite/tellme/issues/52) | The cross-cutting **token bound + timeout** contract (uniform per-tool params: **default + param + ceiling**, loop-enforced; bound from the **effective budget** = `min(MAX_HISTORY_TOKENS, the model's configured window)` → default `÷4`, ceiling `÷2`) + **add `execute_command`** (bash-first `bash -c`, bounded, **process-group** timeout, no `pipe_commands`, no security) + **retrofit the readers** (whole-file reads; fixed 1 MiB/100000 caps retired). Issue [#49](https://github.com/gosharplite/tellme/issues/49) resolved (**config-gated**). | 📋 **Plan+truth APPROVED** — PR [#54](https://github.com/gosharplite/tellme/pull/54); implementation `/axb-tasks` next |
| **025 tool-usage accounting** | [#53](https://github.com/gosharplite/tellme/issues/53) | Count per-tool **pass / fail** (and invocation) usage across a run/session — the empirical instrument for seeing which tools the AI actually uses and which fail, and for tuning defaults / pruning the surface with evidence. | 📋 **Queued** (after 024) |
| **future slices (candidates)** | [#47](https://github.com/gosharplite/tellme/issues/47) | **Concurrent tool-call matching** — parallel tool execution in the agent loop (`MAX_CONCURRENT_TOOLS`-bounded), one-batch feedback; the survivor of [#36](https://github.com/gosharplite/tellme/issues/36) (closed — Gemini API family + ADC dropped). Plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round-023 forward item** — none new; the teardown + echo landed clean. (The reference echoes the prompt via its `provideFeedback` two-line `Input captured:` form; tellme keeps its single captured line **plus** a verbatim echo block — a recorded divergence. The `tellme: `-leading-prompt residual on the diagnostic stream is recorded in `research.md` + the echo DSL row.)
- **Round-022 forward item** — none new; the reshape + separation landed clean. (The reference's `[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Result]` decomposition remains a recorded divergence — tellme uses the single compact `[Tool] <name> - <reason>` line.)
- **Round-021 forward item** — issue [#49](https://github.com/gosharplite/tellme/issues/49): tie the fixed 1 MiB reader-result cap to the resolved `MAX_HISTORY_TOKENS` budget (PR #48 carried-forward item; the `readAggregateCap` constant still exceeds a 200 k-token window).
- **Round-020 forward item** — none new; the cross-compile gate is delivered. (Watch: if a supported target ever needs cgo, the `CGO_ENABLED=0` pin must be revisited.)
- **Round-019 forward items** — the failed-turn carrier proves *absence* (mid-wait *clear-before-the-class-phrase* deferred); the macOS **CPU** leg is pending a cgo `mach` sampler and reports `0.0%` (the memory leg uses sysctl).
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines (plain text); the `tokens.summary.json` roll-up is best-effort (self-heals by recompute).
- **Future-slice candidates** — issue [#47](https://github.com/gosharplite/tellme/issues/47) (**concurrent tool-call matching**); **(d)** coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13); **(e)** the renderer/`-r` forward items.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN** (round 019 did **not** close it — the spinner is a `stderr` diagnostic); round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (estimation-heuristic constants; persona seam; **N-2**); a future **`history.Store.Count()`** should replace `len(prior)+1`.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (refreshed round 023 — carries the `-i` teardown + echo). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Host (this workspace)**: **darwin/arm64** — this session runs on the MacBook Pro niffler env (`…/mbp-johndoe-niffler/ait-tellme`); `go env` reports `darwin/arm64`, so `make verify` compiles macOS natively and the **Linux** path is the cross-compile weak spot here (round 020). The other dev host (`…/beta-niffler/`) is Linux and mirrors the opposite.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`); the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree.
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
