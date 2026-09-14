# tellme — Status

**Last updated**: 2026-09-14 (session 7, day close) — **round 017 DELIVERED / FROZEN** (`017-turn-chrome-parity`, anchor [#42](https://github.com/gosharplite/tellme/issues/42)); PR [#41](https://github.com/gosharplite/tellme/pull/41) **MERGED** into `dev` (`ecf3980`, by `thptcnec`); round-017 head frozen at **`2aead09`**; **propagated `dev → main`**. `make verify` OK · `go test ./...` green · E2E green · topology audit **PASSED** (793 steps). PR #41 review trail: PLAN + TRUTH APPROVED (`dc1a300`) → review folds (`330d567`) → terminology pass (`6220f5f`, `babeee3`) → implementation (`587020a`) → implementation-review nits folded (`d260f70`) → architect-review findings folded (`2aead09`) → merged `ecf3980`. Round 016 detail relocated to the archive ([`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md)).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` — round 017 **DELIVERED / FROZEN**; propagated `dev → main`; the next round starts a fresh `018-*` off `dev`.
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012 detail) · [`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–016 detail).

## Round 017 — `017-turn-chrome-parity` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-14, session 7) — PR [#41](https://github.com/gosharplite/tellme/pull/41) **MERGED** into `dev` (`ecf3980`, by `thptcnec`); propagated `dev → main`. Round-017 head frozen at **`2aead09`**. `make verify` OK · `go test ./...` green · E2E green · topology audit **PASSED** (793 steps). Anchor issue [#42](https://github.com/gosharplite/tellme/issues/42) **closed (completed)**.

**Scope**: make `tellme`'s **non-interactive prompt turn** open like `tell-me-go` — an **input-capture acknowledgement** (`[HH:MM:SS] Input captured. Processing...`), a **80-column `─` rule** + **`╭─⠿ Turn <N> - <mode>` header** over the existing pre-flight payload line, and the reference's **blank-line spacing** — on the **positional/piped prompt turn** and the **round-012 interactive plain reader** only.

**Locked decisions (operator, `/axb-clarify` round 1)**: **surface scope = (A) + (B)** (the `-i` TUI surface and every non-prompt path unchanged); **input-capture line only** (the reference's `[Info] Starting chat...` is out of scope); **post-turn lines** out of scope.

**Review trail**: PLAN + TRUTH APPROVED at `dc1a300` ([#5658345642](https://github.com/gosharplite/tellme/pull/41#issuecomment-5658345642)) → review folds D1–D4 + nits (`330d567`) → terminology pass (`6220f5f`, `babeee3`) → implementation `587020a` → implementation review **FULL ARCHITECTURAL APPROVAL** + nits folded (`d260f70`) → Principal-Architect review **MERGE READY** + findings folded (`2aead09`) → merged `ecf3980`.

**Artifacts / pipeline** — all phases **done**:
- [x] plan package: `spec.md`, `checklists/requirements.md`, `features/acceptance/**` ×3, `research.md` (D1–7), `plan.md` (1 interface / 1 wave; `/axb-api-plan` + `/axb-data-plan` = NOOP; `/axb-ui-plan` skipped), `truth-delta.md`, `tasks.md` (14 tasks; Setup omitted — no new tech).
- [x] truth: `specs/truth/techstack.md` MODIFY (new *Turn chrome (operator)* row); ADD `specs/truth/features/cli/chat/presenting-the-turn.feature`; MODIFY `chat/dsl.md` (**+6 rows**), root `cli/dsl.md` (**+1 cross-module row** `the run shows no turn chrome`), `diagnostics/version-and-setup-diagnostic.feature` + `history/starting-a-fresh-session.feature` (no-chrome carriers); `contracts/**`, `data/**` NOOP.
- [x] implementation (`/axb-implement` T001–T014): `internal/ui/turn.go` (the chrome formatter), `internal/cli/cli.go` (the `chrome` seam — surfaces A/B on, C off), 7 stepdefs + `internal/ui/turn_test.go`; both falsifiability witnesses reproduced.

**Verification (2026-09-14)**: `make verify` **OK** (0 lint · 0 vulns · no `time.Sleep` · offline witness) · `go test -count=1 ./...` green · E2E green · topology audit **PASSED** (31 features · 12 root + 162 module rows · **793 steps**) · falsifiability witnesses (a)/(b) reproduced.

**Open (non-blocking)**: none for the round — the next round starts a fresh `018-*` off `dev` (candidates in Open items below).

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–016 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `017-turn-chrome-parity` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `017-turn-chrome-parity` | delivered / frozen | Round-017 working branch — merged into `dev` via PR [#41](https://github.com/gosharplite/tellme/pull/41) (`ecf3980`); head frozen at `2aead09`; propagated `dev → main`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 016):** `016-interactive-prompt-visual-parity → dev` (PR [#40](https://github.com/gosharplite/tellme/pull/40), `8fef0f8`) `→ main` — DONE (no-ff).
> **Propagation (round 017):** `017-turn-chrome-parity → dev` (PR [#41](https://github.com/gosharplite/tellme/pull/41), `ecf3980`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–016** | — | Provider-registry completeness → … → `-i` strict visual parity. | ✅ **Delivered** (see the delivered-rounds index) |
| **017 — turn-surface operator chrome parity** | [#42](https://github.com/gosharplite/tellme/issues/42) | Make the non-TUI prompt turn open like `tell-me-go`: input-capture line + 80-column `─` rule + `╭─⠿ Turn <N> - <mode>` header over the payload line + blank-line spacing; surfaces (A)+(B) only. | ✅ **Delivered** (PR [#41](https://github.com/gosharplite/tellme/pull/41), `ecf3980`; propagated `dev → main`) |
| **future slices (candidates)** | [#36](https://github.com/gosharplite/tellme/issues/36) | The remaining #34 candidates — the **Google Gemini API family** (inline key), **Application Default Credentials**, and **concurrent tool-call matching**; plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 017 (`017-turn-chrome-parity`)** — **DELIVERED / FROZEN** ([#42](https://github.com/gosharplite/tellme/issues/42)); PR [#41](https://github.com/gosharplite/tellme/pull/41) **MERGED** into `dev` (`ecf3980`, by `thptcnec`); frozen head `2aead09`; `make verify` OK · E2E green · topology audit PASSED (793 steps). **Propagated `dev → main`** — the next round starts a fresh `018-*` off `dev`.
- **Future-slice candidates** — issue [#36](https://github.com/gosharplite/tellme/issues/36): the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials for Vertex, and concurrent tool-call matching (the round-014 E2E helper matches replayed calls by **order** — ready for it).
- **Round-017 forward items** — **post-turn lines** (the reference's measured payload line, the metrics line, the final `╰─⠿ Ready` summary) and the reference's **`[Info] Starting chat...`** line are out of scope; the reference's **gray styling** for the rule/header is a recorded forward item (plain text this round); the rule is a **fixed 80-column literal** (no reflow); a future **`history.Store.Count()`** (per the architect review finding 3) should replace `len(prior)+1` once summarisation/archival lands.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred to multi-turn; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (**estimation-heuristic constants**; the persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`).
- **Future-package candidates**: **(d) coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (low-priority tooling); **(e)** the renderer/`-r` forward items (PR #16 Obs 1/2). (`make verify` in a pipeline platform — [#15](https://github.com/gosharplite/tellme/issues/15) withdrawn `not_planned`, gate stays manual.)

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (round 017)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from the round-017 tree (`d260f70`) — it carries the non-TUI turn chrome; `tellme --version` reports `dev` (no `-ldflags` version stamp). (The `2aead09` refactor is behaviour-preserving.)
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`) — single-prompt and tool-loop turns green; the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 017.
- **Round-017 delivery + closeout (2026-09-14, session 7)**: `/axb-implement` delivered all 14 tasks (the non-TUI turn chrome); PR [#41](https://github.com/gosharplite/tellme/pull/41) **merged** into `dev` (`ecf3980`); two architectural reviews certified the head, with all directives/nits/findings folded (`d260f70`, `2aead09`). `make verify` OK · E2E green · topology audit PASSED (793 steps). Propagated `dev → main`.
