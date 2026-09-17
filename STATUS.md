# tellme — Status

**Last updated**: 2026-09-17 (day close, session 3: **round 037 `037-interactive-prompt-no-selection` — DELIVERED / FROZEN**). Implementation PR [#77](https://github.com/gosharplite/tellme/pull/77) **merged** into `dev` (`3aacf32`, by `thptcnec`); frozen round head `5754cf3` (13 commits). Round 037 aligns the `-i` prompt's suggestion **selection** to the reference: the list opens with **no** suggestion pre-selected (`cursor == noChoice`, mirroring the reference's `suggester{Index: -1}`) and **resets to no-choice on every suggestions refresh** (the reference's `Update(msg, -1)`), so an empty editor never shows a "chosen" hint; the first `Tab` still selects the **first** suggestion. Full AIxBDD pipeline run (specify → spec-by-example → research → system-analysis → ui-plan(terminal) → dsl-refine → tasks → implement) then a **four-round review/fold chain to FINAL CERTIFICATION** (F-1…F-6 → G-1…G-3 → H-1…H-3 → J-1; **no product-code/truth-semantics fold** — all documentation/evidence/predicate). Truth: `techstack.md` (interactive-prompt + suggestion-engine rows) + `presenting-the-interactive-prompt.feature` (at-rest `Then` flipped **in place**) + `chat/dsl.md` (row rewritten + round-037 note); **no new `DSLRow`/Example**; `contracts/**` + `data/**` NOOP. **Supersedes** round 016's pre-selection (which followed the reference's mockup; the reference's code never pre-selects). New issue **[#76](https://github.com/gosharplite/tellme/issues/76)** filed (empty-`Ctrl+S` divergence — out of round scope). Open issues: **#76** · **#69** · **#60** · **#13**. **Propagation: `dev → main` DONE (no-ff).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 037 delivered; the next round starts a fresh `038-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/17/session-summary.md`](docs/session-summary/2026/09/17/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–034) · [`2026-09-17.md`](docs/archives/status/2026-09-17.md) (rounds 035–036).

> **Split note (Rule 12)**: the **round-036** detail section was relocated **verbatim** into [`2026-09-17.md`](docs/archives/status/2026-09-17.md) when round 037 opened (joining round 035), keeping `STATUS.md` to a single current-round section.

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001–037 | `001-*` … `037-interactive-prompt-no-selection` | see the archives + the per-round PRs; round 035 via PR [#73](https://github.com/gosharplite/tellme/pull/73); round 036 via PR [#75](https://github.com/gosharplite/tellme/pull/75); round 037 via PR [#77](https://github.com/gosharplite/tellme/pull/77) (`3aacf32`) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–034 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md); 035–036 in [`2026-09-17.md`](docs/archives/status/2026-09-17.md)); **037 is the most recent delivered round.**

## Current round — 037 `037-interactive-prompt-no-selection` (DELIVERED / FROZEN)

**037-interactive-prompt-no-selection** — the `-i` prompt's suggestion list no longer pre-selects a suggestion: the selection starts at a **no-choice** sentinel (`cursor == noChoice == -1`) and **resets to no-choice on every refresh** (`set(items)` → `cursor = noChoice`, mirroring the reference's `Update(msg, -1)`); the first `Tab` still selects the **first** suggestion (the reference's own `(cursor+delta+len)%len` — `-1 → 0`, and `Shift+Tab` from no-choice → `len-2` for `len ≥ 2` / the sole item for `len == 1`). `selected()`/`cycle()`/`view()` are unchanged; the `> ` glyph, styling, ordering/sources, editor, placeholder, debounce, no-metrics-header and the round-023 teardown are unchanged. **Supersedes** round 016's at-rest "first item is the current choice" (which followed the reference's mockup; the reference's code never pre-selects). Witness = updated **unit pins** (`model_chrome_test.go` incl. the table-driven `TestCycleArithmetic` + `TestModelUnselectedQuoteTextIsNotACursorRow`; `refresh_test.go`) + the flipped **E2E** `Then` asserting zero cursor rows in the rendered **suggestion block** via the exact `"  > "` cursor-row predicate (pinned at two layers: `tests/e2e/steps/tui_chrome_test.go` + the product-rendering pin). Operator Q1–Q4 locked (strict cursor-only scope; exact reference arithmetic; rule restated; witness set). Truth: `techstack.md` + `presenting-the-interactive-prompt.feature` + `chat/dsl.md`; **no new `DSLRow`/Example** (the row is edited in place). Reference parity **verified against `tell-me-go`** by the review. Merged via PR [#77](https://github.com/gosharplite/tellme/pull/77) (`3aacf32`); frozen head `5754cf3` (13 commits).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | round 037 delivered (`3aacf32`) | Integration line (round work lands here before `main`) |
| `037-interactive-prompt-no-selection` | **merged into `dev`** (frozen) | Round 037's working branch — merged via PR [#77](https://github.com/gosharplite/tellme/pull/77) (`3aacf32`). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026–037 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–037** | — | Provider-registry completeness → … → interactive-prompt no-selection. | ✅ **Delivered** (see the archives) |
| **future slices (candidates)** | [#76](https://github.com/gosharplite/tellme/issues/76) · [#69](https://github.com/gosharplite/tellme/issues/69) · [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **empty-`Ctrl+S` divergence** ([#76](https://github.com/gosharplite/tellme/issues/76)); the **composition root + single-ownership** refactor ([#69](https://github.com/gosharplite/tellme/issues/69)); the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)); **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). | ⏳ **Candidates** |

## Open items (non-blocking)

- **Round-037 forward items** — (a) the **empty-`Ctrl+S` divergence** (reference: no-op, stays in the prompt; tellme: quits with exit `0`) → live issue **[#76](https://github.com/gosharplite/tellme/issues/76)** (round-035 G3 lesson: durable surface = a live issue); (b) `[REFACTOR]` the "a refresh resets the selection" policy is hardcoded inside `suggester.set` (the reference names it at the call site) → **[#69](https://github.com/gosharplite/tellme/issues/69)** (single-ownership theme).
- **Round-036 forward items** — (a) `[TECHNICAL DEBT]` the blank-reason suppression predicate has **three sites**, one **provably dead on the production path** → **single ownership** on [#69](https://github.com/gosharplite/tellme/issues/69); (b) the **permanent E2E narrowing** (the `\n`/`\r` reason-collision class has no E2E carrier) — recorded on [#69](https://github.com/gosharplite/tellme/issues/69) **and** [#74](https://github.com/gosharplite/tellme/issues/74); (c) `oneLine` relocated into `toolcall.go` (orphan `toollog.go` deleted).
- **Round-035 forward items** — the observer port hook overload + the spinner-yield-policy ownership (three homes / four call sites) → [#69](https://github.com/gosharplite/tellme/issues/69).
- **Round-034 forward items** — the failed-turn **display-only `Ready` overstatement** (G2) + the **numbering skew**; `BindToolOutput` ctor injection → [#69](https://github.com/gosharplite/tellme/issues/69#issuecomment-5698460385); `LoopObserver` segregation (would **supersede ADR 0005 G1**); the round-022 row→feature audit blind spot → [#60](https://github.com/gosharplite/tellme/issues/60#issuecomment-5697192786).
- **Round-033 forward items** — loader silently best-effort; recursive walk linear; large-catalog bounded by the resource contract; path-sort ordering.
- **Round-032 forward items** — local stdio MCP transport; cross-invocation tool caching; MCP-backed MEMORY/PLUR; MCP `-d` diagnostic; `mcptest/` → [#13](https://github.com/gosharplite/tellme/issues/13); stale `make help` text.
- **Older forward items** — rounds 018–031 forward items live per-round in the archives.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** / **no `flock`**; round-011 forward items.
- **Issue tracker (2026-09-17 closeout, session 3)** — [#76](https://github.com/gosharplite/tellme/issues/76) **OPEN (new this round)** — the empty-`Ctrl+S` divergence (body sharpened by the round-037 review: display-only coupling, lifecycle-only residual); [#69](https://github.com/gosharplite/tellme/issues/69) open (**title refreshed** this closeout to enumerate all five single-ownership items; body carries the spinner-yield ownership + the port hook pair + the blank-reason-predicate single ownership + the suggestion-selection-policy ownership + the permanent E2E narrowing); [#60](https://github.com/gosharplite/tellme/issues/60) open (dogfooding); [#13](https://github.com/gosharplite/tellme/issues/13) open (coverage tooling). **No issues closed this closeout** (round 037 delivered no issue-tracked slice; nothing superseded). [#74](https://github.com/gosharplite/tellme/issues/74) closed (completed) by round 036; [#72](https://github.com/gosharplite/tellme/issues/72) closed by round 035.

## Environment notes

- **Dev tooling — `tellme.sh` (external)**: the Niffler-style manager driving the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/{beta-niffler,mbp-johndoe-niffler}/tellme.sh`; invoke via `source tellme.sh` or the `tm` alias. Its banner is round-agnostic (current-state pointer = this `STATUS.md`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` (**refreshed** from the merged round-037 head `5754cf3`). Sandbox: `unshare -n` unavailable → the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (`verify-mcp-sdk-confinement`); the hermetic fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020).
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, `~/.bashrc`; read for `$TELL_ME_HOME`; read for `…/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` unavailable for this repo; closeout scans are diff-level pattern greps — **clean this closeout**.
