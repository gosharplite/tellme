# tellme — Status

**Last updated**: 2026-09-14 (session 9) — **round 019 `019-turn-spinner` IN PROGRESS**: the **plan + truth half is DONE + APPROVED** (PR [#44](https://github.com/gosharplite/tellme/pull/44) open → `dev`, human-merge pending); **`/axb-implement` is PAUSED** after Phase-1 reconnaissance — **no product code written yet**. Branch `019-turn-spinner` (10 commits, pushed). Topology audit **PASSED** (33 features · 956 steps). Round-018 detail relocated to the archive (Rule 12).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `019-turn-spinner` — round 019 in progress; PR #44 (plan + truth) open + approved; implementation pending. (The next round's work starts here, not off `dev`.)
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–018).

## Round 019 — `019-turn-spinner` (in progress)

**Status**: ⏳ **PLAN + TRUTH HALF DONE + APPROVED; IMPLEMENTATION PENDING.** Branch `019-turn-spinner` (off `dev`); PR [#44](https://github.com/gosharplite/tellme/pull/44) **open** → `dev` (plan + truth only; **human-merge pending**). `/axb-implement` was started and **paused after Phase-1 reconnaissance** (harness/product reading only) — **no product code yet**. Head `be5171b`.

**Scope**: a live **progress spinner** on the **non-TUI** prompt surfaces — between prompt capture and terminal-control regain, so a run is never silently waiting.

**Locked decisions (operator clarify + reviews)**: reference-parity **stop/resume**; **full reference-parity labels with identifiers** (`⠋ Thinking [<model>]... (3s)` / `⠋ Executing [<tool>]... (3s)` / `⠋ Executing tools [<a>, <b>]... (3s)`); keep the tool-execution **CPU/MEM** segment (**machine-wide**); gate = **`isatty(stderr) && !-r`** (the `stderr` spinner gated on `stderr`, like the reference — the stdout probe is **NOT** wired, so **Obs 1 stays OPEN**); `-i` TUI out of scope; **Linux/macOS only**; **no new dependency**; a **`LoopObserver`** seam on `AgentLoop`; synchronous `Stop()`/`Clear()` + one I/O mutex.

**Review trail (PR #44)**: PLAN+TRUTH **REQUEST CHANGES** (B1 `stderr` gate · B2 machine-wide · TD1 port · TD2 failure carrier · R1–R3) → folded `a5d950b` → re-review APPROVE + residual (`FR-008(a)` falsifiability) folded `80525c7` → nits folded `1638ecc` → impl-half review (B1 `LoopObserver` seam · TD1 terminology · TD2 sweep · R1 I/O contract) folded `be5171b` → **FULL APPROVAL — CERTIFIED READY FOR IMPLEMENTATION** ([#5660991544](https://github.com/gosharplite/tellme/pull/44#issuecomment-5660991544)).

**Artifacts / pipeline**:
- [x] plan package: `spec.md` (FR-001–011 · NFR-001–005 · SC-001–005 · A1–A10), `checklists/requirements.md`, `features/acceptance/**` ×3, `research.md` (D1–10), `plan.md` (1 interface / 1 wave; `/axb-api-plan` = NOOP, `/axb-data-plan` = NOOP, `/axb-ui-plan` skipped), `truth-delta.md`, `tasks.md` (**21 tasks**; Setup omitted — no new dependency).
- [x] truth: `specs/truth/techstack.md` MODIFY (spinner row + `stderr` gate + *System metrics provider (telemetry)* row + `LoopObserver` seam + I/O contract); ADD `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (5 Rules); MODIFY `chat/dsl.md` (**+8 rows**), root `cli/dsl.md` (**+2 rows** — the gate Given `the diagnostics are shown at a terminal` + the cross-module negative `the run shows no progress spinner`), `chat/reporting-a-failed-provider-request.feature` (failure carrier), `chat/presenting-the-turn.feature` + `diagnostics`/`history` (boundary carriers); `contracts/**` NOOP; `data/**` NOOP.
- [ ] implementation (`/axb-implement` T001–T021): **not started** (paused after Phase-1 recon: `internal/ui/spinner.go`, `internal/domain/metrics` port + `internal/infrastructure/telemetry` adapters, `internal/domain/agent` `LoopObserver` + `internal/agent/agentloop.go` seam, `internal/cli/cli.go` gate/lifecycle; 10 stepdefs + 3 `[UNIT]`).
- [ ] verification: pending implementation.
- **Audit**: Gherkin/DSL topology **PASSED** (33 features · **15 root + 185 module rows** · **956 steps**).
- **PR #44**: open, plan+truth approved; head `be5171b`.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–018 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `018-post-turn-status-lines` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `019-turn-spinner` | **in progress** | Round-019 working branch (off `dev`); plan+truth committed (**10 commits**, pushed to `origin/019-turn-spinner`); implementation pending; PR [#44](https://github.com/gosharplite/tellme/pull/44) open → `dev`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 018):** `018-post-turn-status-lines → dev` (PR [#43](https://github.com/gosharplite/tellme/pull/43), `9927287`) `→ main` — DONE (no-ff).
> **Propagation (round 019): PENDING** — the round is **not mergeable** (implementation not started); PR #44 (plan+truth) awaits human merge; propagation deferred.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–018** | — | Provider-registry completeness → … → post-turn status lines. | ✅ **Delivered** (see the delivered-rounds index) |
| **019 — turn spinner** | — | Non-TUI live progress spinner: braille frames + phase labels + machine-wide CPU/MEM, gated on `stderr` terminality, with a `LoopObserver` seam. | ⏳ **In progress** — plan+truth approved (PR [#44](https://github.com/gosharplite/tellme/pull/44)); `/axb-implement` over 21 tasks pending |
| **future slices (candidates)** | [#36](https://github.com/gosharplite/tellme/issues/36) | The remaining #34 candidates — the **Google Gemini API family** (inline key), **Application Default Credentials**, and **concurrent tool-call matching**; plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 019 (`019-turn-spinner`)** — **plan + truth APPROVED** (PR [#44](https://github.com/gosharplite/tellme/pull/44), head `be5171b`); **`/axb-implement` over the 21 tasks is the next step** (paused after reconnaissance). PR #44 still needs a **human merge**.
- **Round-019 forward items** — the failed-turn carrier proves *absence* (mid-wait *clear-before-the-class-phrase* deferred); the `-i` / non-prompt carriers now force a terminal (`FR-008(a)` falsifiable).
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines (plain text); the `tokens.summary.json` roll-up is best-effort (self-heals by recompute).
- **Future-slice candidates** — issue [#36](https://github.com/gosharplite/tellme/issues/36) (Gemini API family / ADC / concurrent tool-call matching); **(d)** coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13); **(e)** the renderer/`-r` forward items.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN** (this round did **not** close it — the spinner is a `stderr` diagnostic); round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (estimation-heuristic constants; persona seam; **N-2**); a future **`history.Store.Count()`** should replace `len(prior)+1`.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh**: the installed `$(go env GOPATH)/bin/tellme` was last rebuilt from the **round-018** tree (`26257b4`) — it does **not** carry round 019 (no product code yet); `tellme --version` reports `dev`. Rebuild with `go install ./cmd/tellme` after round-019 implementation.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`); the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree.
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 019.
- **Round-019 plan+truth closeout (2026-09-14, session 9)**: the plan+truth half was taken through **four** review rounds on PR [#44](https://github.com/gosharplite/tellme/pull/44) (2 blockers + 2 TD + 3 R, then a residual + 2 nits, then an impl-half B1/TD1/TD2/R1) — all folded (`a5d950b`, `80525c7`, `1638ecc`, `be5171b`); **FULL APPROVAL — CERTIFIED READY FOR IMPLEMENTATION**. `/axb-implement` was started and paused after reconnaissance (no product code). `gofmt` clean · `go vet` OK · secret scan clean · topology audit **PASSED** (956 steps).
