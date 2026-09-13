# tellme — Status

**Last updated**: 2026-09-13 — **round 010 `010-stream-ordering-observability` DELIVERED / FROZEN** (session 25; PR [#29](https://github.com/gosharplite/tellme/pull/29) merged into `dev` (`7c6d793`); propagated `dev → main`; see the *Round 010* section). Round 009 `009-payload-status-line` DELIVERED + PROPAGATED (session 24; see the *Round 009* section). Prior rounds' detail is in the archive. **Post-round closeout**: `STATUS.md` **split** into `docs/archives/status/2026-09-13.md` (rounds 003–008 detail), and the split procedure added to `SESSION-CLOSEOUT.md` (Rule 12 + Step 3 item 8); propagated `dev → main`. **Ordering fix**: the post-turn payload status line now trails the answer (`7bcb2d3`); the round-010 anchor opened as [#28](https://github.com/gosharplite/tellme/issues/28) (cross-stream ordering observability).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` (round 010 delivered — PR [#29](https://github.com/gosharplite/tellme/pull/29) merged into `dev` (`7c6d793`); propagated `dev → main`). Next round starts a fresh `011-*` off `dev`.
**Daily log**: [`docs/session-summary/2026/09/13/session-summary.md`](docs/session-summary/2026/09/13/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–008 detail + header/review-response/propagation history).

## Round 010 — `010-stream-ordering-observability` (active)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-13, session 25) — PR [#29](https://github.com/gosharplite/tellme/pull/29) merged into `dev` (`7c6d793`); propagated `dev → main`; `make verify` OK. No product change (a **contract + oracle** round): the deliverable is the merged-stream **witness** + the **executable ordering contract**.

**PR**: [#29](https://github.com/gosharplite/tellme/pull/29) (base `dev`) — open, awaiting owner review/merge.

**Review trail (round 010)**: PR [#29](https://github.com/gosharplite/tellme/pull/29#issuecomment-5652030345) — **FULL ARCHITECTURAL APPROVAL — READY TO MERGE** (no blockers); both non-blocking findings fixed in-round: **[TECHNICAL DEBT]** trace-free merged capture (the merged re-run now executes against copies of home/workdir and restores each fake via `Snapshot`/`Restore`, so no history/request-count leak) and **[REFACTOR]** defensive `sc.merged = ""` reset in `run()` → re-review [#5652056493](https://github.com/gosharplite/tellme/pull/29#issuecomment-5652056493) **FINAL APPROVAL — CERTIFIED READY TO MERGE** (commit `7ea79eb`).

**Scope**: make **cross-stream output ordering** (the interleave of the diagnostic stream `stderr` with the answer stream `stdout`) a **first-class, checkable contract** — responding to round 009's ordering defect (the post-turn payload line printed *before* the answer) that **every gate was structurally unable to see**. Required orderings: (a) the payload status brackets the answer (`pre-flight < answer < measured`); (b) the tool-loop log precedes the answer. The **witness** is a merged (`2>&1`) single-buffer capture in the E2E harness — the missing oracle. Behaviour intent **MODIFY** (contract + oracle; no content change; no new dependency).

**Clarify decisions locked (Round 1)**: (Q1) required orderings = **payload-status + tool-loop** (the round-006 degrade warning is incidental); (Q2) E2E witness = **merged single-buffer capture**; (Q3) assert at **both** the unit and E2E layers.

**Artifacts / pipeline**:
- [x] plan package (`spec.md`, `checklists/requirements.md`, `truth-delta.md` skeleton, `features/acceptance/` ×2) — on branch `010-stream-ordering-observability`.
- [x] `research.md` + `specs/truth/techstack.md` MODIFY (cross-stream ordering witness + layered assertions).
- [x] `plan.md` (1 interface / 1 wave; `/axb-api-plan` + `/axb-data-plan` = NOOP).
- [x] truth (`features/cli/chat/**` MODIFY — ordering semantics; topology audit **PASSED**, 467 steps).
- [x] `tasks.md` (13 tasks; Setup omitted — stdlib-only; orphan sweep 0).
- [x] implementation (merged-stream witness + 3 ordering stepdefs + unit emit-order + falsifiability witness); **all 13/13 tasks `[X]`**.

**Pipeline position**: all phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (13/13 tasks `[X]`)**.

**Verification (2026-09-13)**: `make verify` **OK** (0 lint · 0 reachable vulns · no test-sleep · offline witness) · `go test ./...` green · godog **70/70 scenarios** · topology audit **PASSED** (467 steps) · no new dependency · **falsifiability witness** confirmed (a temporarily inverted emit order failed at **both** the unit and E2E layers).

**Open (non-blocking)**: research/DSL-level determinations only — merged-capture plumbing, no-final-answer ordering semantics, where the incidental-interleave note lives, and the unit emit-order helper.

## Round 009 — `009-payload-status-line` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-13, session 24) — plan/truth PR [#26](https://github.com/gosharplite/tellme/pull/26) merged into `dev` (`98c0fb3`); implementation PR [#27](https://github.com/gosharplite/tellme/pull/27) merged into `009-payload-status-line` (`5b744e8`); propagated `009-payload-status-line → dev` (`d4fd911`) `→ main`; `make verify` OK.

**Scope**: per-turn **payload-budget visibility** — a prompt turn reports the payload's **pre-flight estimate** (`~<est>/<max>`) before the request and the **post-turn measured** size (`<actual>/<max>`) after it, on the **diagnostic stream (`stderr`)**, against a configurable **payload budget** (`MAX_HISTORY_TOKENS`, default **1000000**). Observe-only — **no** pruning. Reproduces the reference's `[HH:MM:SS] Payload: … tokens - <mode> - <model>` line.

**Clarify decisions locked (Round 1)**: (Q1) status line → **`stderr`** (`stdout` byte-exact); (Q2) **estimate + actual** (widen `llm.Response` with the provider's reported `usage`); (Q3) **always-on** (not TTY-gated, not `-r`-suppressed); (locked default) `MAX_HISTORY_TOKENS` = **1000000**.

**Artifacts / pipeline**:
- [x] plan package (`spec.md`, `checklists/requirements.md`, `research.md`, `plan.md`, `features/acceptance/` ×2, `tasks.md`, `truth-delta.md`) — merged PR [#26](https://github.com/gosharplite/tellme/pull/26).
- [x] truth (`techstack.md` MODIFY; `features/cli/chat/reporting-the-payload-status.feature` ADD + `chat/dsl.md` rows; `features/cli/history/inspecting-the-session-history.feature` + `history/dsl.md` MODIFY; `/axb-api-plan` + `/axb-data-plan` + root `cli/dsl.md` = NOOP) — merged PR [#26](https://github.com/gosharplite/tellme/pull/26).
- [x] implementation (`internal/domain/llm/token.go`, `internal/ui/status.go`, `internal/config` `MAX_HISTORY_TOKENS`, `llm.Usage` + adapter parse, `AgentLoop.Run` → `AgentResult`, exported `BuildMessages`, CLI `stderr` emitter) + 4 unit files + 9 step files + E2E hermeticity fix — merged PR [#27](https://github.com/gosharplite/tellme/pull/27).
- [x] **all 23/23 tasks `[X]`**.

**Pipeline position**: all phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (23/23 tasks `[X]`)**.

**Verification (2026-09-13)**: `gofmt`/`go vet` clean · `make verify` **OK** (0 lint · 0 reachable vulns · no test-sleep · offline witness) · `go test ./...` green · Gherkin/DSL topology audit **PASSED** (463 steps) · class-phrase vocabulary **10**; `data/**` NOOP; no new dependency.

**Review trail (round 009)**: PR [#26](https://github.com/gosharplite/tellme/pull/26) — review → **REQUEST CHANGES** (BLOCKER-1 `_TBD_`; BLOCKER-2 `AgentLoop.Run` usage seam; TD-1/TD-2; RF-1) → fixed in-round (`e8182a6`) → **CERTIFIED READY TO MERGE** → merged. PR [#27](https://github.com/gosharplite/tellme/pull/27) — **FULL ARCHITECTURAL APPROVAL — READY TO MERGE** → merged.

**Open (non-blocking)**: none new — carried items unchanged (PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / no pruning / no `flock`).


## Delivered rounds (001–008)

Detail lives in the archives (003–008 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md)). Quick index:

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

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`. Delivered round branches
> (`001`–`009`) remain frozen history and never receive post-round commits.
>
> **Propagation (round 009):** `009-payload-status-line → dev` (`d4fd911`) `→ main` (`32074ed`) — DONE; closeout docs on `dev` (`43d3507`).
> **Propagation (round 010):** `010-stream-ordering-observability → dev` (PR [#29](https://github.com/gosharplite/tellme/pull/29), `7c6d793`) `→ main` — DONE; closeout docs on `dev`.
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

## Open items (non-blocking)

- **Round 010 delivered / frozen — [#28](https://github.com/gosharplite/tellme/issues/28)**: **stream-ordering observability** — make cross-stream (`stdout`/`stderr`) ordering assertable (DSL ordering semantics + a merged-stream witness in the E2E harness). PR [#29](https://github.com/gosharplite/tellme/pull/29) merged into `dev` (`7c6d793`); propagated `dev → main`; issue [#28](https://github.com/gosharplite/tellme/issues/28) **closed**. **Next round starts a fresh `011-*` off `dev`.**

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
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).

- **Secret scanning (session 19)**: `mcp_github_run_secret_scanning` is **unavailable for this repo** — GitHub reports *"Repository does not have GitHub Advanced Security enabled."* Closeout secret scans are therefore **diff-level** (pattern grep over `git diff`), as run on 2026-09-12 (session 19 — clean; session 18 — clean; session 17 — clean).
