# tellme — Status

**Last updated**: 2026-09-13 — **round 011 `011-persona-and-payload-estimate` DELIVERED / FROZEN** (session 26; PR [#30](https://github.com/gosharplite/tellme/pull/30) merged into `dev` (`fe6d229`); propagated `dev → main`; see the *Round 011* section). Prior rounds' detail is in the archives. **2026-09-13 (review response):** round 012 PR [#31](https://github.com/gosharplite/tellme/pull/31) — review **BLOCKER B1** fixed (real isatty; ADR 0003) + RF1/RF2/TD fixes at `331cf88`.
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `012-interactive-multiline-prompt` (round 012 in progress — PR [#31](https://github.com/gosharplite/tellme/pull/31) `012-interactive-multiline-prompt → dev`; review response at `331cf88`). Next: re-review → human merge → propagate `dev → main`.
**Daily log**: [`docs/session-summary/2026/09/13/session-summary.md`](docs/session-summary/2026/09/13/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–010 detail + header/review-response/propagation history).

## Round 012 — `012-interactive-multiline-prompt` (active)

**Status**: 🔄 **IN PROGRESS** — PR [#31](https://github.com/gosharplite/tellme/pull/31) (`012-interactive-multiline-prompt → dev`); review **REQUEST CHANGES** → **fixed** (`331cf88`) → re-review **✅ APPROVE** ([#5652688392](https://github.com/gosharplite/tellme/pull/31#issuecomment-5652688392)) → spec/research drift closed (`abc49f0`) → closing confirmation ✅ ([#5652709341](https://github.com/gosharplite/tellme/pull/31#issuecomment-5652709341)). **Amendment A8** (`1fb7a0e`): a prompt-less `--new` on a terminal now archives then reads. Fresh re-review of A8 **REQUEST CHANGES** ([#5652804419](https://github.com/gosharplite/tellme/pull/31#issuecomment-5652804419) — a non-hermetic unit test) → **fixed** (`b5cb61c`).

**Scope**: the reference's interactive multi-line prompt reader (`Ctrl+D`; hint to `stderr`; POSIX-only, no Windows variant). Tasks **8/8 `[X]`**; godog **81/81**; topology audit **PASSED** (547 steps).

**Review response (`331cf88`)**: **B1** — a **real isatty** (`golang.org/x/term.IsTerminal`, already in the module graph) replacing the `os.ModeCharDevice` heuristic, so `< /dev/null` no longer masks a config failure as exit `0` (now exit `3`); **ADR 0003** recorded. **RF1** — `TELL_ME_FORCE_STDIN_TTY` seam + an **E2E positive-read** scenario (carries acceptance Rules 1–2 executably). **RF2** — the hermetic empty-pipe stdin default kept + a **null-device** E2E scenario. **TD1** — the hint is single-sourced (`cli.MultiLineHint`). **TD2/TD3/TD4** — documented. Truth updated (`techstack.md`, `chat/dsl.md`, `reading-a-multi-line-prompt.feature`, `truth-delta.md`).

**Amendment A8 (`1fb7a0e`)**: a prompt-less `--new` on a **terminal** now archives the session **first**, then engages the reader (unifying "start fresh and type"); a **non-terminal** prompt-less `--new` keeps its round-007 archive-and-exit behaviour. Spec `FR-009` amended + `FR-012`/`SC-007` added; `reading-a-multi-line-prompt.feature` + `chat/dsl.md` + `truth-delta.md` updated; unit + E2E (`--new` archive-then-read) added; falsifiability witness reproduced. Landed after the approval, so a fresh re-review may be requested.

**Open**: human merge → propagate `012-interactive-multiline-prompt → dev → main`; then STATUS split + daily log.

## Round 011 — `011-persona-and-payload-estimate` (active)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-13, session 26) — PR [#30](https://github.com/gosharplite/tellme/pull/30) merged into `dev` (`fe6d229`); propagated `dev → main`; `make verify` OK.

**PR**: [#30](https://github.com/gosharplite/tellme/pull/30) (base `dev`) — **MERGED** by `thptcnec` (2026-09-13T09:15:08Z).

**Review trail (round 011)**: PR [#30](https://github.com/gosharplite/tellme/pull/30#issuecomment-5652346462) — **APPROVE WITH NON-BLOCKING FOLLOW-UPS** (TD-1 real-home mutation / leak-dependent example; TD-2 `FR-004` construction-only; TD-3 arrange-run trace; RF-1 duplicated tool-def projection; N-1/N-2/N-3) → all in-scope findings fixed in-round (`2015315`) → re-review [#5652372651](https://github.com/gosharplite/tellme/pull/30#issuecomment-5652372651) **FINAL APPROVAL — CERTIFIED READY TO MERGE**.

**Scope**: make the outbound request faithful to its configuration and the pre-flight estimate faithful to the wire. The configured `PERSON` is sent as the **leading `system` message** of every request the turn makes (empty ⇒ none; request-only), and the pre-flight estimate counts the **wire payload** (persona + tool declarations + messages), so the `~` line is comparable to the provider's measured count. Behaviour intent **MODIFY**; no new dependency.

**Clarify decisions locked (Round 1)**: (Q1) estimate scope = **persona + tool declarations + messages**; (Q2) pin the estimate's **inputs + determinism** (no numeric-equality assertion).

**Artifacts / pipeline**:
- [x] plan package (`spec.md`, `checklists/requirements.md`, `research.md`, `plan.md`, `features/acceptance/` ×2, `tasks.md`, `truth-delta.md`).
- [x] `specs/truth/techstack.md` MODIFY (Persona instruction; Request assembly; Token estimator; Payload status line; testing rows).
- [x] `plan.md` (1 interface / 1 wave; `/axb-api-plan` + `/axb-data-plan` = NOOP).
- [x] truth (`features/cli/chat/**` ADD ×2 + `chat/dsl.md` MODIFY; topology audit **PASSED**, 516 steps).
- [x] `tasks.md` (23 tasks; Setup omitted — stdlib-only; orphan sweep 0).
- [x] implementation (transport persona + `EstimatePayload` + CLI wiring + 10 stepdefs + wire-payload unit test); **all 23/23 tasks `[X]`**.

**Pipeline position**: all phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (23/23 tasks `[X]`)**.

**Verification (2026-09-13)**: `make verify` **OK** (0 lint · 0 reachable vulns · no test-sleep · offline witness) · `go test ./...` green · godog **76/76 scenarios** · topology audit **PASSED** (516 steps) · **falsifiability witness** confirmed (suppressing the persona fails the persona scenario) · no new dependency.

**Open (non-blocking)**: estimation heuristic constants; the persona-plumbing seam shape; **N-2** (estimator ignores replayed tool-call `arguments` — a forward item). Carried: PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / no pruning / no `flock`.

## Delivered rounds (001–010)

Detail lives in the archives (003–010 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md)). Quick index:

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

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | delivered / frozen (round 001) | Round-001 working branch — PR [#6](https://github.com/gosharplite/tellme/pull/6) merged; round 001 is delivered / frozen history |
| `002-followup-cleanups` | delivered / frozen (round 002) | Round-002 base branch — PR [#7](https://github.com/gosharplite/tellme/pull/7) merged (`f2a058f`); propagated `→ dev → main` |
| `003-provider-registry-completeness` | delivered / frozen (round 003) | Round-003 base branch — PR [#11](https://github.com/gosharplite/tellme/pull/11) merged (`9ab3185`); propagated `→ dev → main`; frozen history |
| `004-first-reasoning-turn` | delivered / frozen (round 004) | Round-004 base branch — PR [#12](https://github.com/gosharplite/tellme/pull/12) merged (`4525b38`); propagated `→ dev → main`; frozen history |
| `005-stdin-piping` | delivered / frozen (round 005) | Round-005 branch — PR [#16](https://github.com/gosharplite/tellme/pull/16) **merged** into `dev` (`37c0c24`, "Merge pull request #16"); propagated `→ main`; frozen history |
| `006-rendered-output-and-raw-flag` | delivered / frozen (round 006) | Round-006 branch — PR [#19](https://github.com/gosharplite/tellme/pull/19) **merged** into `dev` (`7cc1304`); propagated `→ main`; frozen history (review branch `006-rendered-output-and-raw-flag-r2`) |
| `007-session-history-persistence` | delivered / frozen (round 007) | Round-007 branch — plan/truth PR [#20](https://github.com/gosharplite/tellme/pull/20) merged (`3187584`) + implementation PR [#21](https://github.com/gosharplite/tellme/pull/21) merged (`f36a83b`); propagated `→ dev` (`6f5483b`) `→ main` (`c7b9950`); frozen history (impl branch `007-implement-session-history-persistence` deleted) |
| `008-agent-tools-and-tool-call-loop` | delivered / frozen (round 008) | Round-008 branch — plan/truth PR [#24](https://github.com/gosharplite/tellme/pull/24) merged (`fb382fc`) + implementation PR [#25](https://github.com/gosharplite/tellme/pull/25) merged (`f9b5d74`); propagated `→ dev` (`d376e03`) `→ main`; frozen history (impl branch `008-implement-agent-tools-and-tool-call-loop` deleted) |
| `009-payload-status-line` | delivered / frozen (round 009) | Round-009 branch — plan/truth PR [#26](https://github.com/gosharplite/tellme/pull/26) merged into `dev` (`98c0fb3`) + implementation PR [#27](https://github.com/gosharplite/tellme/pull/27) merged (`5b744e8`); propagated `009-payload-status-line → dev` (`d4fd911`) `→ main`; frozen history (impl branch `009-implement-payload-status-line` deleted) |
| `010-stream-ordering-observability` | delivered / frozen (round 010) | Round-010 branch — PR [#29](https://github.com/gosharplite/tellme/pull/29) merged into `dev` (`7c6d793`); propagated `dev → main`; frozen history |
| `011-persona-and-payload-estimate` | delivered / frozen (round 011) | Round-011 branch — PR [#30](https://github.com/gosharplite/tellme/pull/30) merged into `dev` (`fe6d229`); propagated `dev → main`; frozen history |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`. Delivered round branches
> (`001`–`009`) remain frozen history and never receive post-round commits.
>
> **Propagation (round 009):** `009-payload-status-line → dev` (`d4fd911`) `→ main` (`32074ed`) — DONE; closeout docs on `dev` (`43d3507`).
> **Propagation (round 010):** `010-stream-ordering-observability → dev` (PR [#29](https://github.com/gosharplite/tellme/pull/29), `7c6d793`) `→ main` — DONE; closeout docs on `dev`.
> **Propagation (round 011):** `011-persona-and-payload-estimate → dev` (PR [#30](https://github.com/gosharplite/tellme/pull/30), `fe6d229`) `→ main` — DONE; closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003 — Provider-registry completeness** | [#9](https://github.com/gosharplite/tellme/issues/9) | Grow boot-subset `PROVIDERS` entry to real provider fields (`API_KEY` with `${VAR}` expansion, `HEADERS`, `THINKING_BUDGET`/`THINKING_LEVEL`) + deterministic offline validation. | ✅ **Delivered** (PR [#11](https://github.com/gosharplite/tellme/pull/11) merged; propagated to `dev`/`main`) |
| **004 — First reasoning turn** | [#10](https://github.com/gosharplite/tellme/issues/10) | `tellme "<prompt>"` → one provider request → printed response; provider domain port + one adapter; deterministic failure class; network-path test strategy. **Depends on 003.** | ✅ **Delivered** (PR [#12](https://github.com/gosharplite/tellme/pull/12) merged `4525b38`; propagated to `dev`/`main`; [#10](https://github.com/gosharplite/tellme/issues/10) closed) |
| **005 — Prompt piping (stdin)** | [#14](https://github.com/gosharplite/tellme/issues/14) | Read the prompt from stdin (combined with the positional instruction), adopt the TTY-aware output contract; defer `-r`; fold F9 flag-parsing unit tests. | ✅ **Delivered** (PR [#16](https://github.com/gosharplite/tellme/pull/16) merged `37c0c24`; propagated to `dev`/`main`; [#14](https://github.com/gosharplite/tellme/issues/14) folded) |
| **006 — Rendered output + `-r`** | [#17](https://github.com/gosharplite/tellme/issues/17) | Default rendered output (glamour Markdown→ANSI) + `-r`/`--raw` inverse + `WRAP_WIDTH`/`TELL_ME_WRAP_WIDTH`. | ✅ **Delivered** (PR [#19](https://github.com/gosharplite/tellme/pull/19) merged `7cc1304`; propagated to `dev`/`main`; [#17](https://github.com/gosharplite/tellme/issues/17) closed) |
| **007 — Session history persistence** | — | Durable session history (`history.jsonl`) + auto-resume + `--new` / `-l N`. | ✅ **Delivered** (PRs [#20](https://github.com/gosharplite/tellme/pull/20)/[#21](https://github.com/gosharplite/tellme/pull/21); propagated `007 → dev → main`) |
| **008 — Agent tools & the tool-call loop** | [#23](https://github.com/gosharplite/tellme/issues/23) | Bounded agent tool-call loop (read-only `list_files`/`read_files` + LLM-backed `summarize_history`); widened history + replay; live `stderr` loop log; `the tool request failed` + exit `7`; `MAX_TOOL_LOOP` (1000). | ✅ **Delivered** (PRs [#24](https://github.com/gosharplite/tellme/pull/24)/[#25](https://github.com/gosharplite/tellme/pull/25); propagated `008 → dev → main`) |
| **009 — Payload status line** | — | Per-turn payload status on `stderr` (pre-flight estimate `~est/max` + post-turn measured `actual/max`); `MAX_HISTORY_TOKENS` budget (default 1000000); widen `Response` with the provider's `usage`; estimator + clock seam. Observe-only (no pruning). | ✅ **Delivered** (PRs [#26](https://github.com/gosharplite/tellme/pull/26)/[#27](https://github.com/gosharplite/tellme/pull/27); propagated `009 → dev → main`) |
| **010 — Stream-ordering observability** | [#28](https://github.com/gosharplite/tellme/issues/28) | Make cross-stream (`stdout`/`stderr`) ordering a checkable contract: pin the payload-status + tool-loop ordering in the CLI truth; add the merged-stream E2E witness; assert at unit + E2E. | ✅ **Delivered** (PR [#29](https://github.com/gosharplite/tellme/pull/29); propagated `010 → dev → main`) |
| **011 — Persona on the wire & wire-faithful payload estimate** | — | Send the configured `PERSON` as the leading `system` message of every request; make the pre-flight estimate count the wire payload (persona + tool declarations + messages). | ✅ **Delivered** (PR [#30](https://github.com/gosharplite/tellme/pull/30); propagated `011 → dev → main`) |

## Open items (non-blocking)

- **Round 011 delivered / frozen**: **persona on the wire + wire-faithful payload estimate** — send the configured `PERSON` as the leading `system` message of every request; the pre-flight estimate counts persona + tool declarations + messages. PR [#30](https://github.com/gosharplite/tellme/pull/30) merged into `dev` (`fe6d229`); propagated `dev → main`. **Next round starts a fresh `012-*` off `dev`.**

- **Round 009 delivered / frozen** — plan/truth PR [#26](https://github.com/gosharplite/tellme/pull/26) merged into `dev` (`98c0fb3`); implementation PR [#27](https://github.com/gosharplite/tellme/pull/27) merged into `009-payload-status-line` (`5b744e8`); propagated `009-payload-status-line → dev` (`d4fd911`) `→ main`. **Next round starts a fresh `010-*` off `dev`.**

- **Round 008 delivered / frozen** — plan/truth PR [#24](https://github.com/gosharplite/tellme/pull/24) (`fb382fc`) + implementation PR [#25](https://github.com/gosharplite/tellme/pull/25) (`f9b5d74`) merged; propagated `008-agent-tools-and-tool-call-loop → dev` (`d376e03`) `→ main`. **Next round starts a fresh `009-*` off `dev`.**

- **Round 007 delivered** — plan/truth PR [#20](https://github.com/gosharplite/tellme/pull/20) (`3187584`) + implementation PR [#21](https://github.com/gosharplite/tellme/pull/21) (`f36a83b`) merged; propagated `007-session-history-persistence → dev` (`6f5483b`) `→ main` (`c7b9950`). **Round 008 followed** (agent tools / the tool-call loop), landing **history summarisation as an on-demand agent tool**; **token-budget pruning is a settled exclusion** (out of `tellme` scope, not carried by that round).

- **Round 006 delivered** — PR [#19](https://github.com/gosharplite/tellme/pull/19) merged (`7cc1304`) + propagated `dev → main`; issue [#17](https://github.com/gosharplite/tellme/issues/17) closed.

- **Round 005 delivered** — PR [#16](https://github.com/gosharplite/tellme/pull/16) merged (`37c0c24`) + propagated `dev → main`. **Next round starts a fresh `006-*` off `dev`** (candidates below).
- **Future-package candidates** (list refreshed 2026-09-12):
  - ~~**(a) Run `make verify` in a pipeline platform**~~ — **WITHDRAWN → CLOSED `not_planned` (2026-09-12)**: [#15](https://github.com/gosharplite/tellme/issues/15) closed — the platform is **not a repo-level choice** and won't be picked any time soon, so the gate stays **manual** (the `SESSION-CLOSEOUT.md` `make verify`). The Makefile stays the single source of gate truth; refile if a platform is ever chosen.
  - ~~**(b) F9 extension — `internal/cli` flag-parsing unit tests**~~ — **FOLDED INTO ROUND 005** ([#14](https://github.com/gosharplite/tellme/issues/14)): the piping slice added `internal/cli/prompt_test.go` (flag parsing + I/O-mode selection). Resolved.
  - ~~**(c) PM-4 `tellme init`**~~ — **DROPPED (2026-09-12)**: config provisioning stays with the environment manager (Niffler / `tellme.sh`); `tellme` remains a **load/validate consumer** working inside that shell — **no second config-writer**. Candidate withdrawn.
  - **(d) Coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (`make test-coverage` report + `go build -cover` E2E-integration spike; PR [#12](https://github.com/gosharplite/tellme/pull/12) review follow-up). Triaged 2026-09-12 as **low-priority tooling**, not a committed round.
  - **(e) Rendered output + `-r` slice** — carries the two forward items from PR #16's **Final Architectural Review** ([#5644538637](https://github.com/gosharplite/tellme/pull/16#issuecomment-5644538637)): **Obs 1** — wire the symmetric `isTTY(stdout)` probe with the renderer (defers with it), and **Obs 2** — consolidate the CLI stream parameters into a `RuntimeEnv` struct as the flag/env surface grows.
- Pre-existing non-blocking items from rounds 001/002 remain documented in archive.


## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at both `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner was made **round-agnostic** this session — the hardcoded `round-001` capability text (and the `bare boot` capability hint) were removed so it no longer needs revising each slice; the current-state pointer is this `STATUS.md`. Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (session 26, round 011)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from `dev`; `tellme --version` reports `dev` (no `-ldflags` version stamp).
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).

- **Secret scanning (session 19)**: `mcp_github_run_secret_scanning` is **unavailable for this repo** — GitHub reports *"Repository does not have GitHub Advanced Security enabled."* Closeout secret scans are therefore **diff-level** (pattern grep over `git diff`), as run on 2026-09-13 (session 26 — clean, round-011 diff; session 25/24 — clean).
