# tellme — Status

**Last updated**: 2026-09-16 (**round 029 `029-agent-write-tools` — plan + truth half reviewed & CERTIFIED READY** — PR [#61](https://github.com/gosharplite/tellme/pull/61); the implementation half (`/axb-implement`, T001–T028) is next; **propagation pending the operator's merge**. Prior: round 028 delivered/frozen; rounds 001–027 live in the archives. **#60** = the dogfooding-enablement umbrella; **#62** = a new transport-hardening slice).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `029-agent-write-tools` (round 029 — plan + truth half **CERTIFIED READY**; PR [#61](https://github.com/gosharplite/tellme/pull/61) awaiting the operator's merge)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (round 027).

## Round 029 — `029-agent-write-tools` (plan + truth half; PR #61 open — awaiting merge)

**Status**: 🟡 **PLAN + TRUTH HALF CERTIFIED READY** (2026-09-16) — branch `029-agent-write-tools`; **PR [#61](https://github.com/gosharplite/tellme/pull/61)** passed a **4-round review/fold loop** and is **CERTIFIED READY TO MERGE**; **awaiting the operator's merge** (propagation pending). First round of the **dogfooding-enablement track** (umbrella [#60](https://github.com/gosharplite/tellme/issues/60)).

**Scope**: add tellme's first **write** capability to the agent tool surface — `write_file` (**create-only**, itself **atomic** via `os.Link`/`EEXIST`; `MkdirAll` mode `0755`; created-file mode `0644`; error if the path exists) and `replace_text` (**strict-unique**: `0` → error, `>1` → error, exactly `1` → replace; **atomic** temp+`rename`; no-op short-circuit). Both writes atomic; **no security/consent gate, no undo**; `reason` required; round-024 resource contract (30 s default); a **missing** `content` key is rejected (explicit `""` only). **Omitted:** `append_text`, `undo_file_change`, `delete_path`, `create_directory` (the shell covers them). `write_file`'s survival is to be **measured** via `--tool-usage` after dogfooding.

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · **implement ⏳ (next)**.

**Artifacts**: `spec.md` (US1 `replace_text` P1 · US2 `write_file` P2 · FR-001–014 · NFR-001–002 · SC-001–005), `checklists/requirements.md` (ready; 0 clarify), `features/acceptance/*.feature` ×3, `research.md` (D1–8 + an R1 forward item), `plan.md` (1 interface → `/axb-dsl-refine`; api/data NOOP; ui skipped), `tasks.md` (T001–T028; orphan sweep 0), `truth-delta.md`. **Truth**: `techstack.md` MODIFY (Write filesystem tools row); `chat/creating-and-editing-files.feature` **ADD** (4 Rules) + `chat/dsl.md` **MODIFY** (offered set → 6 tools; +6 Given +10 Then; note) + `chat/offering-the-agent-tools.feature` **MODIFY** (4→6 tools); `contracts/**` + `data/**` NOOP.

**Review (PR #61)**: 1 architectural blocker (atomicity asymmetry — `replace_text` now atomic) + 3 technical-debt (atomic create-only · mode `0644` · a unit-tier atomicity witness) + 3 refactor (stale offered-set truth · read hazard recorded · single-sourced expected set) + 4 consistency (A–D) + a **reference cross-check** (R1 → **#62**; R2 missing-`content` reject + R3 dir mode `0755` folded) — **all closed**; **CERTIFIED READY** (no open findings).

**Verification (half)**: topology audit **PASSED** (41 features · 16 root + 266 module rows · 1360 steps); no product code → `make verify` N/A.

**Commits**: `6a541e0` (plan package + spec) · `035cbcc` (acceptance + research + techstack) · `af3b2ac` (system-analysis plan) · `303c529` (interface truth) · `2785d3e` (tasks) · `a684148` (status) · `2cfc9c6` (review fold) · `6583bc7` (re-review fold) · `8b7d078` (cross-check fold R2/R3) · `d22d2cd` (cross-check nit fold).

## Round 028 — `028-user-global-prompt-log` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-16; merged 2026-09-15) — PR [#59](https://github.com/gosharplite/tellme/pull/59) **MERGED** into `dev` (`9e13f9e`, by `gosharplite`, 2026-09-15T20:05:06Z); round-028 head frozen at **`39821fd`** (10 commits); propagated `dev → main` (no-ff). An **operator-request** state-location slice: the `-i` shared prompt log moves from the environment-scoped `<TELL_ME_HOME>/output/global_prompts.jsonl` to the **per-user** `~/.tellme/global_prompts.jsonl`, with a one-time `Seed(ctx)` migration.

**Scope**: read **and** write the prompt log at the user-global `~/.tellme/global_prompts.jsonl` (resolved via the CLI-injected `userHomeDir` resolver — mirroring the round-026 `newToolUsageStore` seam); on first use, when the file is absent, copy the environment-scoped file **verbatim** (copy, **not** move; never overwrite — `O_EXCL`; a blank runtime home skips it; **best-effort**). The record shape, the newest-first deduped read, the compaction policy, and the `-i`-only write rule are **unchanged**; the TUI chrome is unchanged. **Recorded divergence:** tellme no longer shares the log with `tell-me-go` (sharing unit *per-`TELL_ME_HOME`* → *per-user*); the seed publish is **not atomic** (non-blocking forward note). **ADR 0004**.

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · data-plan ✅ · dsl-refine ✅ · tasks ✅ · **implement ✅** (T001–T016 all `[X]`).

**Artifacts**: `spec.md` (US1–US2 · FR-001–008 · NFR-001–004 · SC-001–004), `checklists/requirements.md` (ready), `features/acceptance/*.feature` ×2, `research.md` (D1–7 + residual risks), `plan.md` (2 interfaces → `/axb-dsl-refine` + `/axb-data-plan`; api NOOP; ui skipped), `tasks.md` (T001–T016; orphan sweep 21/21), `truth-delta.md`. **Truth**: `techstack.md` MODIFY (Shared global prompt log); `data/data-model.dbml` MODIFY (`prompt_log_entry` location + seed lifecycle + divergence); `chat/carrying-over-the-environment-prompt-log.feature` **ADD** (3 Rules) + `chat/recording-the-shared-prompt-log.feature` MODIFY (header) + `chat/dsl.md` MODIFY (3 rows retargeted + 4 new seed rows + the round-028 note); `contracts/**` NOOP. **Code**: `internal/infrastructure/history/global_prompt_tracker.go` (destination via the injected resolver; pure ctor + explicit `Seed(ctx)`; `O_EXCL` no-overwrite), `internal/domain/history/tracker.go` (+ the segregated `Seeder` capability port + doc), `internal/cli/cli.go` (both call sites hold the `PromptTracker` port; `Seed` via the `Seeder` type-assert, once, before the interactive read); `internal/ui` + `internal/agent` **untouched**; `docs/decisions/0004-user-global-prompt-log.md` **NEW**.

**Verification**: topology audit **PASSED** (40 features · 6 modules · 16 root + **250** module rows · **1312** steps). `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `golangci-lint` 0 issues · `govulncheck` clean); `go test -count=1 ./...` green (unit + godog E2E, **182 scenarios · 1336 steps**, 0 undefined). **Falsifiability witnesses** reproduced then reverted — (a) write-target revert (env-scoped) → 7 scenario failures · (b) seed source disabled → the carry-over Example fails · (c) never-overwrite guard disabled → the no-carry-over Example fails. `stdout` byte-exact; vocabulary 11; `go.mod`/`go.sum` byte-exact (stdlib-only).

**Review trail (PR #59)**: plan+truth **APPROVED WITH REQUIRED FOLDS** (TD-1 resolver seam · TD-2 explicit `Seed` · TD-3 contention escalation · RF-1 ADR · RF-2 stale docs · RF-3 nits) → folds `0eceef2` (TD/RF set) + `f67f934` (blank-home micro-note) → **FINAL ARCHITECTURAL APPROVAL**; implementation **CERTIFIED READY TO MERGE** (`d41405d`; impl-review forward note `eb682b6`); **principal-architect review** → **CERTIFIED** (TD-1 compaction forward note · TD-2 dormant `wg` documented · REFACTOR-1 `Seeder` interface segregation) → fold **`39821fd`** → **RE-CERTIFIED — FINAL ARCHITECTURAL APPROVAL** → **MERGED** `9e13f9e`.

**Commits**: `43000ce` (plan package + spec) · `1d0fe95` (acceptance + research + techstack) · `4d24c12` (system-analysis plan) · `65820ad` (data + interface truth) · `b09eec1` (tasks) · `0eceef2` (plan-review fold) · `f67f934` (micro-note fold) · `d41405d` (implementation) · `eb682b6` (impl-review forward note) · `39821fd` (architect-review fold) · `9e13f9e` (PR #59 merge into `dev`).

**Propagation**: `028-user-global-prompt-log → dev` (PR [#59](https://github.com/gosharplite/tellme/pull/59), `9e13f9e`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.

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
| 028 | `028-user-global-prompt-log` | PR [#59](https://github.com/gosharplite/tellme/pull/59) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); 028 stays here as the most recent delivered round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `028-user-global-prompt-log` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `029-agent-write-tools` | in flight | The round-029 working branch off `dev` — **plan + truth half CERTIFIED READY**; **PR [#61](https://github.com/gosharplite/tellme/pull/61) open, awaiting the operator's merge**; propagation pending. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 026):** `026-tool-usage-accounting → dev` (PR [#57](https://github.com/gosharplite/tellme/pull/57), `9d62379`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**.
> **Propagation (round 027):** `027-ai-call-turn-counter → dev` (PR [#58](https://github.com/gosharplite/tellme/pull/58), `ecd4d44`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**.
> **Propagation (round 028):** `028-user-global-prompt-log → dev` (PR [#59](https://github.com/gosharplite/tellme/pull/59), `9e13f9e`, merged by `gosharplite`) `→ main` — **DONE (no-ff)**; closeout docs on `dev`.
> **Propagation (round 029):** **PENDING** — awaits the operator's merge of PR [#61](https://github.com/gosharplite/tellme/pull/61).
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **026** ([#53](https://github.com/gosharplite/tellme/issues/53), delivered) = tool-usage accounting; **027** (operator request, delivered) = the AI-call turn counter; **028** (operator request, delivered) = the user-global prompt log; **029** (operator request; part of the dogfooding track [#60](https://github.com/gosharplite/tellme/issues/60)) = the first agent write tools.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–024** | — | Provider-registry completeness → … → the tool resource contract + `execute_command`. | ✅ **Delivered** (see the delivered-rounds index) |
| **025 spinner width-safety** | [#55](https://github.com/gosharplite/tellme/issues/55) | **Bound** the several-tool spinner label + a **width-safe (row-aware) clear**. | ✅ **Delivered** — round 025 (PR [#56](https://github.com/gosharplite/tellme/pull/56)); issue [#55](https://github.com/gosharplite/tellme/issues/55) **closed** |
| **026 tool-usage accounting** | [#53](https://github.com/gosharplite/tellme/issues/53) | Count per-tool **invocations + outcome** in a **user-global** log (`~/.tellme/tools-count.jsonl`), surfaced by the offline **`--tool-usage`** report. | ✅ **Delivered** — round 026 (PR [#57](https://github.com/gosharplite/tellme/pull/57)); issue [#53](https://github.com/gosharplite/tellme/issues/53) **closed** |
| **027 AI-call turn counter** | — (operator request) | Redefine the `╭─⠿ Turn <N>` header's `<N>` to count the session's **AI-endpoint calls**; persist per-turn `calls`. | ✅ **Delivered** — round 027 (PR [#58](https://github.com/gosharplite/tellme/pull/58), `ecd4d44`) |
| **028 user-global prompt log** | — (operator request) | Move the `-i` shared prompt log to `~/.tellme/global_prompts.jsonl` + a first-use `Seed(ctx)`; record the `tell-me-go` divergence (ADR 0004). | ✅ **Delivered** — round 028 (PR [#59](https://github.com/gosharplite/tellme/pull/59), `9e13f9e`) |
| **029 agent write tools** | — (operator request; dogfooding track [#60](https://github.com/gosharplite/tellme/issues/60)) | tellme's first **write** capability: `write_file` (create-only, atomic) + `replace_text` (strict-unique, atomic). | 🔵 **In flight** — round 029 (PR [#61](https://github.com/gosharplite/tellme/pull/61)); plan + truth half **CERTIFIED READY**, implementation next |
| **future slices (candidates)** | [#13](https://github.com/gosharplite/tellme/issues/13) · [#62](https://github.com/gosharplite/tellme/issues/62) | **Coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); **provider-transport truncation guard** ([#62](https://github.com/gosharplite/tellme/issues/62), a data-corruption hazard surfaced at round 029); the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds; plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-16, reconciled post-closeout)** — **#13** open (coverage tooling; future candidate, still accurate); **#60** open (dogfooding-enablement umbrella — round 029 is its first slice); **#62** open (provider-transport truncation guard; surfaced at round 029). #47 closed (`not_planned`), #53/#55 completed (earlier). **No changes this closeout.**
- **Round-029 forward items** — (a) **merge PR [#61](https://github.com/gosharplite/tellme/pull/61)** → propagate `029 → dev → main`, then run `/axb-implement` over **T001–T028**; (b) **#62** — the provider-transport guard (silent truncation of large tool args) is a **deliberate deferral** to its own transport round; (c) **`write_file` survival** is to be **measured** via `--tool-usage` after real dogfooding (its marginal value is lower than `replace_text`'s); (d) `replace_text` reads the **whole** file (an unbounded **input** read — recorded hazard; a future input bound is a forward item).
- **Round-028 forward items** — (a) the **seed publish is not atomic** (`O_EXCL` create + one `Write`); (b) the **user-global log grows unbounded** and `Recent()` reads the whole file — a future `Compact(maxEntries)` must take an advisory **`flock`** (shared with round 026); (c) the **`tell-me-go` divergence** is recorded. **PR [#59](https://github.com/gosharplite/tellme/pull/59) merged; propagation DONE.**
- **Round-027 forward items** — (a) **legacy entries count as 1**; (b) a future **summarisation/archive** path must preserve the counter; (c) the reference's **per-call chrome cadence** remains a recorded divergence. (Round-027 detail relocated to [`2026-09-16.md`](docs/archives/status/2026-09-16.md).)
- **Round-026 forward items** — (1) recoverable inline failures count as `ok`; (2) pruning erases the prune signal; (3) the `~/.tellme/tools-count.jsonl` log grows unbounded. (Round-026 detail → [`2026-09-15.md`](docs/archives/status/2026-09-15.md).)
- **Round-025 forward item** — the **mid-frame-resize over-erase bound** (TD-3) is recorded (accepted). **FD-1**: the narrow-terminal residue witness depends on the tool-phase frame wrapping at the forced width.
- **Round-024 forward items** — the `CONTEXT_WINDOW` bound is **config-gated**; `output_file` redirects **both** streams; the `ProcessRunner` port extraction trigger.
- **Round-023/022 forward items** — none new (recorded divergences).
- **Round-021 forward item** — ✅ **resolved / closed by round 024**: [#49](https://github.com/gosharplite/tellme/issues/49) closed (completed).
- **Round-020/019/018 forward items** — the cross-compile gate is delivered (watch the `CGO_ENABLED=0` pin); the failed-turn carrier proves *absence* + the macOS **CPU** leg pending a cgo `mach` sampler; the post-turn gray styling + the best-effort `tokens.summary.json`.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`. The operator's live test env for round 028 was `…/mbp-johndoe-niffler/ait-test/` (tag `test`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (**last refreshed round 028**). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential. **Round 029 added no product code** (plan + truth only) — the installed binary is unchanged this round.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees; registered read-only this session).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
