# tellme — Status

**Last updated**: 2026-09-13 — **round 013 `013-vertex-gemini-provider` DELIVERED / FROZEN** (PR [#33](https://github.com/gosharplite/tellme/pull/33) merged into `dev` (`6ed3bbc`, by `thptcnec`); propagated `dev → main` (`9a3587a`); see the *Round 013* section). Prior rounds' detail is in the archives ([2026-09-11](docs/archives/status/2026-09-11.md), [2026-09-13](docs/archives/status/2026-09-13.md)).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` — round 013 delivered / frozen; the next round starts a fresh `014-*` off `dev` (candidate: issue [#34](https://github.com/gosharplite/tellme/issues/34)).
**Daily log**: [`docs/session-summary/2026/09/13/session-summary.md`](docs/session-summary/2026/09/13/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012 detail + header/review-response/propagation history).

## Round 013 — `013-vertex-gemini-provider` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-13) — PR [#33](https://github.com/gosharplite/tellme/pull/33) **MERGED** into `dev` (`6ed3bbc`, by `thptcnec`); propagated `dev → main` (`9a3587a`). Round-013 head frozen at **`6f3f83f`** (the re-certified SHA). `make verify` OK · godog **88/88** · topology audit **PASSED** (595 steps) · **live-verified against Vertex**.

**Scope**: add the **Vertex AI / Gemini provider family** to the CLI end — a new `internal/infrastructure/llm/gemini` adapter (Vertex `:generateContent` request assembly + response normalization, plus a **stdlib** service-account OAuth2 flow), the factory family mapping (`gemini`/`google`), and the E2E support. Behaviour intent **ADD** (a new provider transport). Anchor issue [#32](https://github.com/gosharplite/tellme/issues/32).

**Clarify Round 1 (locked `1,1,1`)**: (Q1) **Vertex AI only** (Gemini API family + ADC deferred); (Q2) **`.json`-suffix service-account detection** (no schema change); (Q3) **stdlib-only OAuth2** — no new module.

**Review / re-certification trail**: FULL ARCHITECTURAL APPROVAL (plan & truth) → directives D1–D4 folded (`d1fc433`) → certified → implementation (`de79fc0`) → certified ready to merge → **two live-usage defects** found running against real Vertex (**`821824f`** — one thinking knob + surfaced Vertex error; **`f7af54f`** — echo the Gemini `thoughtSignature` on replayed tool calls) → **RE-CERTIFIED at `6f3f83f`** → merged `6ed3bbc` → propagated `9a3587a`.

**Artifacts / pipeline** — all phases **done**:
- [x] plan package: `spec.md`, `checklists/requirements.md`, `research.md` (Decisions 1–10), `plan.md` (1 interface / 1 wave; `/axb-api-plan` + `/axb-data-plan` = NOOP; CLI end → `/axb-dsl-refine`), `features/acceptance/**` ×3, `tasks.md` (21 tasks; Setup omitted — stdlib-only; orphan sweep 0), `truth-delta.md`.
- [x] truth: `specs/truth/techstack.md` MODIFY (Vertex/Gemini adapter; Provider family mapping; Service-account authentication; credential resolution); `specs/truth/features/cli/chat/**` ADD ×3 + `chat/dsl.md` MODIFY; root `cli/dsl.md` + `contracts/**` + `data/**` NOOP.
- [x] implementation: `internal/infrastructure/llm/gemini/{client.go,auth.go}`; `factory.go`; unit tests (`client_test.go`, `auth_test.go`, `factory_test.go`); E2E fake (Vertex shape + token exchange) + 7 step files + wire-family-agnostic helpers.

**Verification (2026-09-13)**: `make verify` **OK** (0 lint · 0 reachable vulns · no test-sleep · offline witness) · `go test ./...` green · godog **88/88** · topology audit **PASSED** (595 steps) · **falsifiability witness** reproduced (detaching the token fails the auth scenario) · `go.mod`/`go.sum` unchanged · **live Vertex runs green** (single prompt + tool loop).

**Open (non-blocking)**: **issue [#34](https://github.com/gosharplite/tellme/issues/34)** — the **014 slice candidate**: persist the tool-call signature for faithful **resume** replay (the `thoughtSignature` forward item). Carried: PR #16 **Obs 1** (stdout TTY probe) OPEN; round-006 **Obs 3** (renderer lifecycle) deferred; sequential tool execution / no pruning / no `flock`; round-011 forward items (estimation-heuristic constants; persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`). The next round starts a fresh `014-*` off `dev`.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `013-vertex-gemini-provider` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 013):** `013-vertex-gemini-provider → dev` (PR [#33](https://github.com/gosharplite/tellme/pull/33), `6ed3bbc`) `→ main` (`9a3587a`) — DONE; closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–012** | — | Provider-registry completeness → first reasoning turn → stdin piping → rendered output + `-r` → session-history persistence → agent tools & the tool-call loop → payload status line → stream-ordering observability → persona on the wire + wire-faithful estimate → interactive multi-line prompt capture. | ✅ **Delivered** (see the delivered-rounds index) |
| **013 — Vertex AI / Gemini provider support** | [#32](https://github.com/gosharplite/tellme/issues/32) | Drive a `TYPE: gemini` Vertex provider end-to-end (Vertex `:generateContent` transport + stdlib service-account OAuth2), no new module; `anthropic`/Gemini-API/ADC deferred. | ✅ **Delivered** (PR [#33](https://github.com/gosharplite/tellme/pull/33); propagated `013 → dev → main`) |
| **014 (candidate) — Session-replay fidelity** | [#34](https://github.com/gosharplite/tellme/issues/34) | Persist the tool-call **signature** so a **resumed** session replays its tool steps faithfully (the `thoughtSignature` forward item); also lists further 014 candidates (Gemini API family, ADC, concurrent tool calls). | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 014 candidate** — issue [#34](https://github.com/gosharplite/tellme/issues/34): persist the per-tool-step provider signature for faithful **resume** replay (Gemini `thoughtSignature`); a **data-model** change (owned by `/axb-data-plan`), so it is its own slice — not part of round 013.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred to multi-turn; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (**estimation-heuristic constants**; the persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`).
- **Future-package candidates** (list refreshed 2026-09-12): ~~(a) run `make verify` in a pipeline platform~~ (withdrawn — [#15](https://github.com/gosharplite/tellme/issues/15) closed `not_planned`; the gate stays manual); ~~(b) F9 flag-parsing units~~ (folded into round 005); ~~(c) `tellme init`~~ (dropped — config provisioning stays with the env manager); **(d) coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (low-priority tooling); **(e)** the renderer/`-r` forward items (PR #16 Obs 1/2).

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (round 013)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from the round-013 head (`6f3f83f`); `tellme --version` reports `dev` (no `-ldflags` version stamp).
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`) — single-prompt and tool-loop turns green; the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 013.
