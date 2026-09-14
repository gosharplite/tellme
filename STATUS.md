# tellme — Status

**Last updated**: 2026-09-14 (session 8, day close) — **round 018 DELIVERED / FROZEN** (`018-post-turn-status-lines`); PR [#43](https://github.com/gosharplite/tellme/pull/43) **MERGED** into `dev` (`9927287`, by `thptcnec`); round-018 head frozen at **`26257b4`**; **propagated `dev → main`**. `make verify` OK · `go test ./...` green · E2E green · topology audit **PASSED** (883 steps). PR #43 review trail: PLAN + TRUTH APPROVED (`7f5f335`, #5659599275) → implementation APPROVED (`decc4a1`, #5659712977) → doc-comment fold (`a9cdcb4`, #5659733926) → Principal-Architect findings folded (`26257b4`, #5659960243) → merged `9927287`. Round 017 detail relocated to the archive ([`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md)).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` — round 018 **DELIVERED / FROZEN**; propagated `dev → main`; the next round starts a fresh `019-*` off `dev`.
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012 detail) · [`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–017 detail).

## Round 018 — `018-post-turn-status-lines` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-14, session 8) — PR [#43](https://github.com/gosharplite/tellme/pull/43) **MERGED** into `dev` (`9927287`, by `thptcnec`); propagated `dev → main`. Round-018 head frozen at **`26257b4`**. `make verify` OK · `go test ./...` green · E2E green · topology audit **PASSED** (883 steps).

**Scope**: add `tellme`'s **post-turn status** to the prompt surfaces (the counterpart round 017 deliberately deferred) — a per-turn **metrics line** `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>` (from the API call that just returned; `Th` always shown) and a **`╰─⠿ Ready`** session summary `($<lastCall> $<turn> $<session> - M: … H: … O: … - <hit%>%)` — backed by a **config-only `MODELS` pricing table** and a **per-mode API-call usage log**.

**Locked decisions (operator interview + clarify Q1)**: `Th` **always** shown (incl. `Th: 0`); three `$` = **last-returned call / whole turn / session**; line-2's `M/H/C/Th` + `$#1` from **the call that just returned**; **pricing config-only** (no built-in rates; un-priced → `$0.0000`, `$` group still renders); storage = per-mode `output/<mode>/tokens.log` (one JSON record per call); presence = **every prompt-bearing turn**, `stderr`, plain text, suppressed only when the provider reports no usage; `stdout` byte-exact; vocabulary unchanged (11).

**Review trail**: PLAN + TRUTH APPROVED at `7f5f335` ([#5659599275](https://github.com/gosharplite/tellme/pull/43#issuecomment-5659599275)) → implementation APPROVED at `decc4a1` ([#5659712977](https://github.com/gosharplite/tellme/pull/43#issuecomment-5659712977)) → doc-comment fold `a9cdcb4` ([#5659733926](https://github.com/gosharplite/tellme/pull/43#issuecomment-5659733926)) → Principal-Architect findings folded `26257b4` ([#5659960243](https://github.com/gosharplite/tellme/pull/43#issuecomment-5659960243)) → merged `9927287`.

**Artifacts / pipeline** — all phases **done**:
- [x] plan package: `spec.md` (FR-001–014 · NFR-001–004 · SC-001–005 · A1–A10), `checklists/requirements.md`, `features/acceptance/**` ×3, `research.md` (D1–9), `plan.md` (2 interfaces / 1 wave; `/axb-api-plan` = NOOP, `/axb-data-plan` = ADD, `/axb-ui-plan` skipped), `truth-delta.md`, `tasks.md` (25 tasks; Setup omitted — no new dependency).
- [x] truth: `specs/truth/techstack.md` MODIFY (post-turn status-lines + `MODELS` pricing + `tokens.log` rows); ADD `specs/truth/data/data-model.dbml` `usage_record` (+ `tokens.summary.json` roll-up companion); ADD `specs/truth/features/cli/chat/presenting-the-post-turn-status.feature` (9 Rules); MODIFY `chat/dsl.md` (**+15 rows**), root `cli/dsl.md` (**+1 cross-module row** `the run reports no post-turn status`), `diagnostics`/`history` features (boundary carriers); `contracts/**` NOOP.
- [x] implementation (`/axb-implement` T001–T025): `internal/domain/llm` (`Usage` widened with cached/reasoning; exclusive `C`), `internal/infrastructure/llm/openai` (usage details + `max(0,…)` floor), `internal/config` (config-only `MODELS`), `internal/ui/{pricing,metrics}.go`, `internal/domain/history` + `internal/infrastructure/history/usage_store.go` (per-mode `tokens.log` + `tokens.summary.json` O(1) roll-up + batch append), `internal/agent/agentloop.go` (per-call accumulate), `internal/cli/cli.go` (emit + `--new` rotate); 16 stepdefs + 3 UNIT files; both falsifiability witnesses reproduced.

**Verification (2026-09-14)**: `make verify` **OK** (0 lint · 0 vulns · no `time.Sleep` · offline witness) · `go test -count=1 ./...` green · E2E **126/126 scenarios · 907 steps** · topology audit **PASSED** (32 features · 13 root + 177 module rows · **883 steps**) · falsifiability witnesses (a)/(b) reproduced · **no new dependency**.

**Open (non-blocking)**: none for the round — the next round starts a fresh `019-*` off `dev` (candidates in Open items below).

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–017 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `018-post-turn-status-lines` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `018-post-turn-status-lines` | delivered / frozen | Round-018 working branch — merged into `dev` via PR [#43](https://github.com/gosharplite/tellme/pull/43) (`9927287`); head frozen at `26257b4`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 017):** `017-turn-chrome-parity → dev` (PR [#41](https://github.com/gosharplite/tellme/pull/41), `ecf3980`) `→ main` — DONE (no-ff).
> **Propagation (round 018):** `018-post-turn-status-lines → dev` (PR [#43](https://github.com/gosharplite/tellme/pull/43), `9927287`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–017** | — | Provider-registry completeness → … → non-TUI turn chrome parity. | ✅ **Delivered** (see the delivered-rounds index) |
| **018 — post-turn status lines** | — | Non-TUI post-turn status: metrics line + `╰─⠿ Ready` summary with config-only pricing + per-mode `tokens.log`. | ✅ **Delivered** (PR [#43](https://github.com/gosharplite/tellme/pull/43), `9927287`; propagated `dev → main`) |
| **future slices (candidates)** | [#36](https://github.com/gosharplite/tellme/issues/36) | The remaining #34 candidates — the **Google Gemini API family** (inline key), **Application Default Credentials**, and **concurrent tool-call matching**; plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 018 (`018-post-turn-status-lines`)** — **DELIVERED / FROZEN**; PR [#43](https://github.com/gosharplite/tellme/pull/43) **MERGED** into `dev` (`9927287`, by `thptcnec`); frozen head `26257b4`; `make verify` OK · E2E 126/126 · topology audit PASSED (883 steps). **Propagated `dev → main`** — the next round starts a fresh `019-*` off `dev`.
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines is a recorded forward item (plain text this round); the session roll-up (`tokens.summary.json`) is best-effort (a crash between append and the atomic summary write can understate; a missing summary self-heals by recompute).
- **Future-slice candidates** — issue [#36](https://github.com/gosharplite/tellme/issues/36): the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials for Vertex, and concurrent tool-call matching (the round-014 E2E helper matches replayed calls by **order** — ready for it).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred to multi-turn; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (**estimation-heuristic constants**; the persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`); a future **`history.Store.Count()`** should replace `len(prior)+1` once summarisation/archival lands.
- **Future-package candidates**: **(d) coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (low-priority tooling); **(e)** the renderer/`-r` forward items (PR #16 Obs 1/2). (`make verify` in a pipeline platform — [#15](https://github.com/gosharplite/tellme/issues/15) withdrawn `not_planned`, gate stays manual.)

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (round 018)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from the round-018 tree (`26257b4`) — it carries the post-turn status lines; `tellme --version` reports `dev` (no `-ldflags` version stamp). The `tm` alias invokes the same binary.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`) — single-prompt and tool-loop turns green; the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree.
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 018.
- **Round-018 delivery + closeout (2026-09-14, session 8)**: `/axb-implement` delivered all 25 tasks (the post-turn status lines); PR [#43](https://github.com/gosharplite/tellme/pull/43) **merged** into `dev` (`9927287`); three review rounds certified the head, with all directives/nits/findings folded (`d39c996`, `decc4a1`, `a9cdcb4`, `26257b4`). `make verify` OK · E2E 126/126 · topology audit PASSED (883 steps). Propagated `dev → main`.
