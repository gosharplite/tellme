# tellme — Status

**Last updated**: 2026-09-17 (day close, session 1: **round 035 `035-spinner-tail-residue` — DELIVERED / FROZEN**). Implementation PR [#73](https://github.com/gosharplite/tellme/pull/73) **merged** into `dev` (`18cd947`); frozen round head `11cc9a3` (9 commits). Round 035 fixed issue [#72](https://github.com/gosharplite/tellme/issues/72): the round-034 per-call **tail** is now written with the round-019/025 spinner **phase-boundary yielded** (a synchronous clear before the tail's first line, **no resume** — the next waiting phase re-activates). Pure display fix; the offered tool set / flags / exit codes / line formats / cadence / `stdout` are unchanged. ✅ **Review loop CLOSED** (4 review rounds, 3 folds; fix approved since `af41102`; truth half certified byte-identical through all folds). Open issues: **#74** (new `FormatToolReason` defect) · **#69** · **#60** · **#13**. **Propagation: `dev → main` DONE (no-ff).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 035 delivered; the next round starts a fresh `036-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/17/session-summary.md`](docs/session-summary/2026/09/17/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–034).

> **Split note (Rule 12)**: no split this closeout — `STATUS.md` was already lean (the round-034 detail went to [`2026-09-16.md`](docs/archives/status/2026-09-16.md) last closeout and no per-round detail section accumulated since), so there is nothing to relocate to a `2026-09-17.md` archive.

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001–035 | `001-*` … `035-spinner-tail-residue` | see the archives + the per-round PRs; round 035 via PR [#73](https://github.com/gosharplite/tellme/pull/73) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–034 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **035 is the most recent delivered round.**

## Current round — 035 `035-spinner-tail-residue` (DELIVERED / FROZEN)

**035-spinner-tail-residue** — the per-call tail is written with the spinner **phase-boundary yielded** (a synchronous clear before the tail, no resume; the next waiting phase re-activates). Pure display fix; `Spinner.OnCallEnd` stays a no-op (ADR 0005 D1 untouched); `callRenderer` / `internal/ui` unchanged; the final deferred tail is byte-identical. Truth: `techstack.md` spinner row + a new `presenting-the-progress-spinner` Rule reusing existing rows (`dsl.md` note; **no new `DSLRow`**). Witness: E2E (gate + tool round + whole-stream residue row, previously unpaired) + a unit yield-ordering pin. Merged via PR [#73](https://github.com/gosharplite/tellme/pull/73) (`18cd947`); frozen head `11cc9a3`.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | round 035 delivered (`18cd947`) | Integration line (round work lands here before `main`) |
| `035-spinner-tail-residue` | **merged into `dev`** (frozen) | Round 035's working branch — merged via PR [#73](https://github.com/gosharplite/tellme/pull/73) (`18cd947`). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026–035 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–035** | — | Provider-registry completeness → … → spinner tail residue. | ✅ **Delivered** (see the archives) |
| **future slices (candidates)** | [#74](https://github.com/gosharplite/tellme/issues/74) · [#69](https://github.com/gosharplite/tellme/issues/69) · [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **`FormatToolReason` sanitize/cap fix** ([#74](https://github.com/gosharplite/tellme/issues/74)); the **composition root + spinner-yield-policy ownership** refactor ([#69](https://github.com/gosharplite/tellme/issues/69)); the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)); **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). | ⏳ **Candidates** |

## Open items (non-blocking)

- **Round-035 forward items** — (a) `[TECHNICAL DEBT]` the observer port hook overload (`BeforeToolLog`/`AfterToolLog` serving two resume policies) → [#69](https://github.com/gosharplite/tellme/issues/69); (b) `[REFACTOR]` the turn's spinner-yield policy now has **three homes / four call sites** with **no named owner** → [#69](https://github.com/gosharplite/tellme/issues/69) (reframed to "single ownership of the yield policy"); (c) the SC-002 line-count narrowing is **permanent** (the whole-stream residue row is the carrier) — its future literal-count candidate is recorded in the round-035 `research.md` residual risk; (d) the G1 `FormatToolReason` defect → [#74](https://github.com/gosharplite/tellme/issues/74).
- **Round-034 forward items** — (a) the failed-turn **display-only `Ready` overstatement** (G2) and the **numbering skew** (per-call frames reach `prior + k`, next prompt restarts at `prior + 1`) are recorded divergences; **G2 now also couples to round 035** (TD-2: `final` is load-bearing for both the deferred tail and the round-035 yield) — its round must re-check the round-035 phase-boundary yield; (b) `[TECHNICAL DEBT]` `BindToolOutput` registry mutation → constructor injection at the composition root (→ [#69](https://github.com/gosharplite/tellme/issues/69#issuecomment-5698460385)); (c) `[REFACTOR]` `LoopObserver` interface segregation (would **supersede ADR 0005 G1**); (d) the round-022 row→feature audit blind spot → [#60](https://github.com/gosharplite/tellme/issues/60#issuecomment-5697192786).
- **Round-033 forward items** — loader silently best-effort; recursive walk linear; large-catalog bounded by the resource contract; path-sort ordering.
- **Round-032 forward items** — local stdio MCP transport; cross-invocation tool caching; MCP-backed MEMORY/PLUR; MCP `-d` diagnostic; `mcptest/` → [#13](https://github.com/gosharplite/tellme/issues/13); stale `make help` text.
- **Older forward items** — rounds 018–031 forward items live per-round in the archives.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** / **no `flock`**; round-011 forward items.
- **Issue tracker (2026-09-17 closeout, session 1)** — [#72](https://github.com/gosharplite/tellme/issues/72) **CLOSED (completed)** — delivered by round 035 (PR [#73](https://github.com/gosharplite/tellme/pull/73) merged `18cd947`); [#74](https://github.com/gosharplite/tellme/issues/74) **OPEN (new)** — `FormatToolReason` unsanitized/uncapped (a verified, test-pinned round-022 B1 regression; its own round); [#69](https://github.com/gosharplite/tellme/issues/69) open (composition root; **reframed** to also carry the spinner-yield-policy ownership + the port hook pair); [#60](https://github.com/gosharplite/tellme/issues/60) open (dogfooding umbrella); [#13](https://github.com/gosharplite/tellme/issues/13) open (coverage tooling).

## Environment notes

- **Dev tooling — `tellme.sh` (external)**: the Niffler-style manager driving the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/{beta-niffler,mbp-johndoe-niffler}/tellme.sh`; invoke via `source tellme.sh` or the `tm` alias. Its banner is round-agnostic (current-state pointer = this `STATUS.md`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` (**refreshed** from the merged round-035 head `11cc9a3`). Sandbox: `unshare -n` unavailable → the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (`verify-mcp-sdk-confinement`); the hermetic fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020).
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, `~/.bashrc`; read for `$TELL_ME_HOME`; read for `…/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` unavailable for this repo; closeout scans are diff-level pattern greps — **clean this closeout**.
