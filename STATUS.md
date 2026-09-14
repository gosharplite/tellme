# tellme — Status

**Last updated**: 2026-09-14 (session 6, day close) — **round 016 DELIVERED / FROZEN** (`016-interactive-prompt-visual-parity`, anchor [#39](https://github.com/gosharplite/tellme/issues/39)); PR [#40](https://github.com/gosharplite/tellme/pull/40) **MERGED** into `dev` (`8fef0f8`, by `thptcnec`); round-016 head frozen at **`3906b57`** (the re-certified SHA); **propagated `dev → main`**. `make verify` OK · godog **107/107 scenarios · 754 steps** · topology audit **PASSED** (730 steps). PR #40 review trail: PLAN + TRUTH APPROVED (`35dd2cb`) → folds (`9c12cbf`) → architect directives folded (`bd8e11a`) → control plane certified → implementation certified (`3906b57`) → merged. Round 015 detail relocated to the archive ([`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md)).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` — round 016 **DELIVERED / FROZEN**; propagated `dev → main`; the next round starts a fresh `017-*` off `dev`.
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012 detail) · [`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–015 detail).

## Round 016 — `016-interactive-prompt-visual-parity` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-14, session 6) — PR [#40](https://github.com/gosharplite/tellme/pull/40) **MERGED** into `dev` (`8fef0f8`, by `thptcnec`); propagated `dev → main`. Round-016 head frozen at **`3906b57`** (the re-certified SHA). `make verify` OK · godog **107/107 scenarios · 754 steps** · topology audit **PASSED** (730 steps). Anchor issue [#39](https://github.com/gosharplite/tellme/issues/39) **closed (completed)** (2026-09-14).

**Scope**: make `tellme -i` a **strict visual-parity** re-creation of `tell-me-go -i` — a **bordered** multi-line editor above a **styled suggestion list**; **remove** the round-015 dashboard header; debounced/cancelable suggestions + `Tab`-inserts; terminal-width reflow.

**Locked decision (operator, #39)**: **strict parity** — bordered editor + styled `Suggestions:` list, keybinding hints in the placeholder; **no** dashboard header / status line / `?` overlay. Supersedes round-015 US3 for the `-i` surface.

**Review trail**: PLAN + TRUTH APPROVED at `35dd2cb` ([#5657665895](https://github.com/gosharplite/tellme/pull/40#issuecomment-5657665895)) → F1–F3 + nits folded (`9c12cbf`, re-certified) → architect directives **D1–D4** folded (`bd8e11a`, control plane certified) → implementation `08b03ad` → implementation review: **BLOCKER** (unit cursor exactness + truth revert) + **TD1** (run-ctx/`Destroy`) + **TD2** (debounce seam) folded (`3906b57`) → **FULL ARCHITECTURAL APPROVAL — CERTIFIED READY TO MERGE** ([#5657995195](https://github.com/gosharplite/tellme/pull/40#issuecomment-5657995195)) → merged `8fef0f8`.

**Artifacts / pipeline** — all phases **done**:
- [x] plan package: `spec.md`, `checklists/requirements.md`, `features/acceptance/**` ×3, `ui/ui-plan.md` + `ui/screens/*.txt` ×4 (terminal mode), `research.md` (D1–8), `plan.md` (2 interfaces / 1 wave; `/axb-api-plan` + `/axb-data-plan` = NOOP), `truth-delta.md`, `tasks.md` (26 tasks; Setup omitted — no new tech).
- [x] truth: `specs/truth/techstack.md` MODIFY (strict-parity chrome; dashboard removed; debounce/insert; `TELL_ME_TUI_DEBOUNCE` seam); `specs/truth/features/cli/chat/**` — DELETE the dashboard rule, ADD `presenting-the-interactive-prompt.feature`, MODIFY `prompting-with-suggestions.feature` + `chat/dsl.md` (**+11 / −2** rows); `contracts/**`, `data/**`, root `cli/dsl.md` NOOP.
- [x] implementation (`/axb-implement` T001–T026): strict-parity chrome (`model`/`textarea`/`suggester`), debounced/cancelable refresh (D1), `textarea.Model.Update` delegation (D2), `store.Load` removed (D3), cursor `0` (D4), `Tab`-inserts, resize; the two falsifiability witnesses reproduced.

**Verification (2026-09-14)**: `make verify` **OK** (0 lint · 0 vulns · no `time.Sleep` · offline witness) · godog **107/107 scenarios · 754 steps** · topology audit **PASSED** (30 features · 11 root + 156 module rows · **730 steps**) · falsifiability witnesses (a)/(b) reproduced. **Soft-pin**: F1.1/F1.2/F2 unit-pinned (annotations); round-015 `seeing-the-session-dashboard.feature` superseded (un-carried by design, `FR-004`).

**Open (non-blocking)**: none for the round — the next round starts a fresh `017-*` off `dev` (candidates in Open items below).

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–015 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `016-interactive-prompt-visual-parity` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `016-interactive-prompt-visual-parity` | delivered / frozen | Round-016 working branch — merged into `dev` via PR [#40](https://github.com/gosharplite/tellme/pull/40) (`8fef0f8`); head frozen at `3906b57`; propagated `dev → main`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 015):** `015-interactive-tui-prompt → dev` (PR [#38](https://github.com/gosharplite/tellme/pull/38), `a3df102`) `→ main` (`60bbf72`) — DONE (no-ff).
> **Propagation (round 016):** `016-interactive-prompt-visual-parity → dev` (PR [#40](https://github.com/gosharplite/tellme/pull/40), `8fef0f8`) `→ main` — DONE (no-ff); closeout docs on `dev`.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–015** | — | Provider-registry completeness → … → interactive TUI prompt capture. | ✅ **Delivered** (see the delivered-rounds index) |
| **016 — `-i` prompt strict visual parity** | [#39](https://github.com/gosharplite/tellme/issues/39) | Make `tellme -i` feel identical to `tell-me-go -i` (strict parity): bordered editor + styled suggestion list; **remove** the dashboard header; debounced suggestions + `Tab`-inserts; resize. | ✅ **Delivered** (PR [#40](https://github.com/gosharplite/tellme/pull/40), `8fef0f8`; propagated `dev → main`) |
| **future slices (candidates)** | [#36](https://github.com/gosharplite/tellme/issues/36) | The remaining #34 candidates — the **Google Gemini API family** (inline key), **Application Default Credentials**, and **concurrent tool-call matching**; plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 016 (`016-interactive-prompt-visual-parity`)** — **DELIVERED / FROZEN** ([#39](https://github.com/gosharplite/tellme/issues/39)); PR [#40](https://github.com/gosharplite/tellme/pull/40) **MERGED** into `dev` (`8fef0f8`, by `thptcnec`); frozen head `3906b57`; `make verify` OK · godog **107/107** · topology audit PASSED (730 steps). **Propagated `dev → main`** — the next round starts a fresh `017-*` off `dev`.
- **Future-slice candidates** — issue [#36](https://github.com/gosharplite/tellme/issues/36): the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials for Vertex, and concurrent tool-call matching (the round-014 E2E helper matches replayed calls by **order** — ready for it).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred to multi-turn; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (**estimation-heuristic constants**; the persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`).
- **Future-package candidates**: **(d) coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (low-priority tooling); **(e)** the renderer/`-r` forward items (PR #16 Obs 1/2). (`make verify` in a pipeline platform — [#15](https://github.com/gosharplite/tellme/issues/15) withdrawn `not_planned`, gate stays manual.)

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (round 016)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from the round-016 head (`3906b57`) — it now carries the strict-parity `-i` prompt; `tellme --version` reports `dev` (no `-ldflags` version stamp).
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`) — single-prompt and tool-loop turns green; the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 016.
- **Round-016 delivery + closeout (2026-09-14, session 6)**: `/axb-implement` delivered all 26 tasks (strict-parity `-i` prompt); PR [#40](https://github.com/gosharplite/tellme/pull/40) **merged** into `dev` (`8fef0f8`); the implementation review's BLOCKER + TD1/TD2 folded (`3906b57`). `make verify` OK · godog **107/107** · topology audit PASSED (730 steps). Propagated `dev → main`.
