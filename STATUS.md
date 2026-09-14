# tellme — Status

**Last updated**: 2026-09-14 (session 11) — **round 020 `020-cross-compile-gate` DELIVERED / FROZEN**: PR [#46](https://github.com/gosharplite/tellme/pull/46) human-**MERGED** into `dev` (`642583b`, by `gosharplite`) and propagated `dev → main`; frozen head `57a3a05`. The round-020 detail stays here as the current-round section (Rule 12 — older rounds 001–019 live in the archives).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` — round 020 delivered/frozen; the next round starts a fresh `021-*` off `dev`.
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019).

## Round 020 — `020-cross-compile-gate` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-14) — PR [#46](https://github.com/gosharplite/tellme/pull/46) **MERGED** into `dev` (`642583b`, by `gosharplite`, 2026-09-14T11:13:47Z); propagated `dev → main`. Round-020 head frozen at **`57a3a05`**. `make verify` OK · cross-compile gate green (4/4) · `gofmt` clean · `go.mod`/`go.sum` unchanged.

**Scope**: a **host-independent cross-compile gate** in the quality pipeline — `make verify` previously compiled only the host `GOOS`/`GOARCH`, so build-tagged, OS-specific production code for any other target was invisible to every gate (round-019 review forward recommendation; the darwin sampler had shipped uncompiled).

**Locked decisions**: **POSIX** target matrix `linux/amd64 · linux/arm64 · darwin/amd64 · darwin/arm64`, **host-independent** (this workspace is darwin/arm64, so the *Linux* path was the weak spot); mechanism = a **`Makefile` `verify-cross-compile`** target (build + vet per target, **`CGO_ENABLED=0`-pinned** for hermeticity, fail fast naming the target) wired into `make verify` + referenced from the closeout checklist; **no new dependency**; **no CLI behaviour change** (non-BDD tooling round). **Roadmap**: Gemini API family + ADC **dropped** → issue [#47](https://github.com/gosharplite/tellme/issues/47) (concurrent tool-call matching) replaces the closed [#36](https://github.com/gosharplite/tellme/issues/36).

**Artifacts / pipeline** — all phases **done**:
- [x] spec: `spec.md` (US1–US2 · FR-001–008 · NFR-001–004 · SC-001–004 · A1–A5), `checklists/requirements.md`, `truth-delta.md`.
- [x] research/plan: `research.md` (D1–6); `plan.md` (**0 interfaces**; `/axb-api-plan` = NOOP, `/axb-data-plan` = NOOP, `/axb-dsl-refine` = NOOP, `/axb-spec-by-example` + `/axb-ui-plan` skipped); `tasks.md` (T001–T004; orphan sweep 0).
- [x] truth: `techstack.md` **MODIFY** (Build & Tooling: *Cross-compile verification* row + *Task runner* aggregate); `contracts/**` NOOP; `data/**` NOOP; `features/cli/**` NOOP.
- [x] implementation: `Makefile` `verify-cross-compile` (+ `verify` wiring, `.PHONY`, `help`); `SESSION-CLOSEOUT.md` reference (Step 2).

**Review trail (PR #46)**: PLAN+IMPLEMENTATION **APPROVED (non-blocking directives)** (TD1 `CGO_ENABLED` pin · REFACTOR host-`vet` order) → fold `57a3a05` → **FINAL ARCHITECTURAL APPROVAL — CERTIFIED READY TO MERGE** → **MERGED** `642583b`. Anchor [#45](https://github.com/gosharplite/tellme/issues/45) closed (completed).

**Verification (2026-09-14)**: `make verify` OK (no-test-sleep · offline witness · cross-compile · `golangci-lint` 0 issues · `govulncheck` 0 reachable vulns) · `make verify-cross-compile` green (4/4) · **TD1 hermeticity proof** (`CGO_ENABLED=1 make verify-cross-compile` green) · **falsifiability witness** (broken non-host `system_metrics_linux.go` → gate exit 2, naming `linux/amd64` + `…:60:9`) · `gofmt -l .` clean · `go.mod`/`go.sum` unchanged.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)); 020 stays here as the most recent delivered round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `020-cross-compile-gate` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 018):** `018-post-turn-status-lines → dev` (PR [#43](https://github.com/gosharplite/tellme/pull/43), `9927287`) `→ main` — DONE (no-ff).
> **Propagation (round 019):** `019-turn-spinner → dev` (PR [#44](https://github.com/gosharplite/tellme/pull/44), `4315e59`) `→ main` — DONE (no-ff).
> **Propagation (round 020):** `020-cross-compile-gate → dev` (PR [#46](https://github.com/gosharplite/tellme/pull/46), `642583b`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–020** | — | Provider-registry completeness → … → the cross-compile gate. | ✅ **Delivered** (see the delivered-rounds index) |
| **cross-compile gate** | — | Host-independent `go build ./...` + `go vet ./...` gate for the POSIX target matrix (`linux/amd64 · linux/arm64 · darwin/amd64 · darwin/arm64`), wired into `make verify` + the closeout checklist (round-019 review forward recommendation). | ✅ **Delivered** — round 020 (PR [#46](https://github.com/gosharplite/tellme/pull/46)) |
| **future slices (candidates)** | [#47](https://github.com/gosharplite/tellme/issues/47) | **Concurrent tool-call matching** — parallel tool execution in the agent loop (`MAX_CONCURRENT_TOOLS`-bounded), one-batch feedback; the survivor of [#36](https://github.com/gosharplite/tellme/issues/36) (closed — Gemini API family + ADC dropped). Plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round-020 forward items** — none new; the cross-compile gate is delivered. (Watch: if a supported target ever needs cgo, the `CGO_ENABLED=0` pin must be revisited.)
- **Round-019 forward items** — the failed-turn carrier proves *absence* (mid-wait *clear-before-the-class-phrase* deferred); the macOS **CPU** leg is pending a cgo `mach` sampler and reports `0.0%` (the memory leg uses sysctl).
- **Round-018 forward items** — the reference's **gray styling** for the post-turn lines (plain text); the `tokens.summary.json` roll-up is best-effort (self-heals by recompute).
- **Future-slice candidates** — issue [#47](https://github.com/gosharplite/tellme/issues/47) (**concurrent tool-call matching**); **(d)** coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13); **(e)** the renderer/`-r` forward items.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN** (round 019 did **not** close it — the spinner is a `stderr` diagnostic); round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (estimation-heuristic constants; persona seam; **N-2**); a future **`history.Store.Count()`** should replace `len(prior)+1`.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Host (this workspace)**: **darwin/arm64** — this session runs on the MacBook Pro niffler env (`…/mbp-johndoe-niffler/ait-tellme`); `go env` reports `darwin/arm64`, so `make verify` compiles macOS natively and the **Linux** path is the cross-compile weak spot here (round 020). The other dev host (`…/beta-niffler/`) is Linux and mirrors the opposite.
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`); the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` (`…/beta-niffler/ait-bdd`) — the vendored skill tree.
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
