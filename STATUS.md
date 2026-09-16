# tellme — Status

**Last updated**: 2026-09-16 (day close, session 21: **round 034 `034-tool-call-log-parity` — DELIVERED / FROZEN**). Implementation PR [#71](https://github.com/gosharplite/tellme/pull/71) **merged** into `dev` (`806cede`); plan+truth half via PR [#70](https://github.com/gosharplite/tellme/pull/70) (`dd488e9`). Round 034 re-cut tellme's `stderr` tool-call log to the reference's **decomposed** shape (`[Tool Engine]`/`[Tool Reason]`/`[Tool Action]`/`[Tool Output]`/`[Tool Result]`) and the turn output to a **per-AI-endpoint-call** frame cadence, plus a live **bounded-and-stopped** `[Tool Output]` stream. ✅ **FINAL ARCHITECTURAL APPROVAL + concurrence — review loop CLOSED.** Round-034 detail moved to the archive; rounds 001–034 in the delivered index. Open issues: **#69** (CLI composition root) · **#60** (dogfooding umbrella) · **#13** (coverage tooling). **Propagation: `dev → main` DONE (no-ff).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 034 delivered; the next round starts a fresh `035-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–034).

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001–034 | `001-*` … `034-tool-call-log-parity` | see the archives + the per-round PRs; round 034 via PR [#71](https://github.com/gosharplite/tellme/pull/71) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–034 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **034 is the most recent delivered round.**

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | round 034 delivered (`806cede`) | Integration line (round work lands here before `main`) |
| `034-tool-call-log-parity` | **merged into `dev`** (frozen) | Round 034's working branch — merged via PR [#71](https://github.com/gosharplite/tellme/pull/71) (`806cede`). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026–034 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–034** | — | Provider-registry completeness → … → tool-call log parity. | ✅ **Delivered** (see the archives) |
| **future slices (candidates)** | [#69](https://github.com/gosharplite/tellme/issues/69) · [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **composition-root refactor** ([#69](https://github.com/gosharplite/tellme/issues/69)); the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)); **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). | ⏳ **Candidates** |

## Open items (non-blocking)

- **Round-034 forward items** — (a) the failed-turn **display-only `Ready` overstatement** (G2) and the **numbering skew** (per-call frames reach `prior + k`, next prompt restarts at `prior + 1`) are recorded divergences; (b) `[TECHNICAL DEBT]` `BindToolOutput` registry mutation → constructor injection at the composition root (→ [#69](https://github.com/gosharplite/tellme/issues/69#issuecomment-5698460385)); (c) `[REFACTOR]` `LoopObserver` interface segregation (would **supersede ADR 0005 G1** → a future ADR-superseding decision); (d) the round-022 row→feature audit blind spot → [#60](https://github.com/gosharplite/tellme/issues/60#issuecomment-5697192786).
- **Round-033 forward items** — loader silently best-effort; recursive walk linear; large-catalog bounded by the resource contract; path-sort ordering.
- **Round-032 forward items** — local stdio MCP transport; cross-invocation tool caching; MCP-backed MEMORY/PLUR; MCP `-d` diagnostic; `mcptest/` → [#13](https://github.com/gosharplite/tellme/issues/13); stale `make help` text.
- **Older forward items** — rounds 018–031 forward items live per-round in the archives.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** / **no `flock`**; round-011 forward items.
- **Issue tracker (2026-09-16 closeout, session 21)** — [#69](https://github.com/gosharplite/tellme/issues/69) open (CLI composition root; now also carries the round-034 `BindToolOutput` constructor-injection forward item); [#60](https://github.com/gosharplite/tellme/issues/60) open (dogfooding umbrella; also the round-022 row→feature audit guard note); [#13](https://github.com/gosharplite/tellme/issues/13) open (coverage tooling). Round 034 has **no anchor issue** (operator request); its work has **landed** (PR #71) → no closes this closeout (no open issue was satisfied by round 034).

## Environment notes

- **Dev tooling — `tellme.sh` (external)**: the Niffler-style manager driving the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/{beta-niffler,mbp-johndoe-niffler}/tellme.sh`; invoke via `source tellme.sh` or the `tm` alias. Its banner is round-agnostic (current-state pointer = this `STATUS.md`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` (**refreshed** from the merged round-034 head `806cede`). Sandbox: `unshare -n` unavailable → the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (`verify-mcp-sdk-confinement`); the hermetic fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020).
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, `~/.bashrc`; read for `$TELL_ME_HOME`; read for `…/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` unavailable for this repo; closeout scans are diff-level pattern greps — **clean this closeout**.
