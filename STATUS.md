# tellme — Status

**Last updated**: 2026-09-17 (day close, session 2: **round 036 `036-tool-reason-sanitize` — DELIVERED / FROZEN**). Implementation PR [#75](https://github.com/gosharplite/tellme/pull/75) **merged** into `dev` (`ddd6f7d`); frozen round head `11356db` (12 commits). Round 036 fixed issue [#74](https://github.com/gosharplite/tellme/issues/74): `FormatToolReason` now **folds** (`\n`/`\r` → space) + **trims** + **caps** the model-authored reason at **200 rendered runes** (`reasonValueCap`, one U+2026 inside the cap, rune-safe) like its siblings, and a **blank** reason emits **no** `[Tool Reason]` line — restoring round-022 **B1**'s fold guarantee that round 034 dropped **unrecorded** (recorded now in **ADR 0006**). Pure display fix; the sibling caps, flags, exit codes, line formats, cadence, schemas, transport, persisted records and `stdout` are unchanged. ✅ **Review loop CLOSED** (architectural review → fold → fold review → residual sweep → final certification; **truth half re-certified byte-identical across all heads**). Open issues: **#69** · **#60** · **#13**. **Propagation: `dev → main` DONE (no-ff).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 036 delivered; the next round starts a fresh `037-*` off `dev`)
**Daily log**: [`docs/session-summary/2026/09/17/session-summary.md`](docs/session-summary/2026/09/17/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–034) · [`2026-09-17.md`](docs/archives/status/2026-09-17.md) (round 035).

> **Split note (Rule 12)**: the **round-035** detail section was relocated **verbatim** into [`2026-09-17.md`](docs/archives/status/2026-09-17.md) this closeout (round 036 is now the current round), keeping `STATUS.md` to a single current-round section.

## Delivered rounds (index)

| Round | Branch | Delivered via |
| --- | --- | --- |
| 001–036 | `001-*` … `036-tool-reason-sanitize` | see the archives + the per-round PRs; round 035 via PR [#73](https://github.com/gosharplite/tellme/pull/73); round 036 via PR [#75](https://github.com/gosharplite/tellme/pull/75) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–034 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md); 035 in [`2026-09-17.md`](docs/archives/status/2026-09-17.md)); **036 is the most recent delivered round.**

## Current round — 036 `036-tool-reason-sanitize` (DELIVERED / FROZEN)

**036-tool-reason-sanitize** — the `stderr` tool log's model-authored `[Tool Reason]` value is now **sanitized and capped like its siblings**: `FormatToolReason` = `capRunes(oneLine(strings.TrimSpace(reason)), reasonValueCap)` with `reasonValueCap = 200` (one U+2026 inside the cap, rune-boundary cut), and the **three** blank-reason suppression sites emit **no** reason line for an empty/whitespace-only reason (the loop's `logAction` + `reasonsOf` = the production filters; a defensive guard in the `callRenderer.OnCallEnd` `emit` closure). Single-site fold in the pure formatter; `oneLine` relocated out of the retired orphan `internal/ui/toollog.go`. Restores round-022 **B1**'s fold guarantee (dropped **unrecorded** by round 034); recorded in the new **ADR 0006** (ADR 0005 stays immutable). Truth: `techstack.md` Agent-tool-loop row + `chat/dsl.md` reason row (explicit single-line guarantee + the `reasonValueCap` constant) + a round-036 note; **no new `DSLRow`/Example** (`/axb-spec-by-example` NOOP). Witness: **hostile-fixture unit pins** (`\n`, `\r`, trim, combined, 201-rune cap, mid-rune boundary) + loop/tail blank-reason pins — a deliberate **permanent narrowing** (the E2E surface is structurally blind to this class). Merged via PR [#75](https://github.com/gosharplite/tellme/pull/75) (`ddd6f7d`); frozen head `11356db` (12 commits).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | round 036 delivered (`ddd6f7d`) | Integration line (round work lands here before `main`) |
| `036-tool-reason-sanitize` | **merged into `dev`** (frozen) | Round 036's working branch — merged via PR [#75](https://github.com/gosharplite/tellme/pull/75) (`ddd6f7d`). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026–036 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–036** | — | Provider-registry completeness → … → tool-reason sanitize. | ✅ **Delivered** (see the archives) |
| **future slices (candidates)** | [#69](https://github.com/gosharplite/tellme/issues/69) · [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **composition root + spinner-yield/blank-reason-predicate ownership** refactor ([#69](https://github.com/gosharplite/tellme/issues/69)); the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)); **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)). | ⏳ **Candidates** |

## Open items (non-blocking)

- **Round-036 forward items** — (a) `[TECHNICAL DEBT]` the blank-reason suppression predicate has **three sites**, one **provably dead on the production path** (the `callRenderer.OnCallEnd` `emit` closure; `agent.reasonsOf` filters upstream) → **single ownership** parked on [#69](https://github.com/gosharplite/tellme/issues/69) (no in-round refactor); (b) the **permanent E2E narrowing** (the `\n`/`\r` reason-collision class has no E2E carrier; the hostile-fixture unit pin is the deterministic carrier) — recorded on [#69](https://github.com/gosharplite/tellme/issues/69) **and** [#74](https://github.com/gosharplite/tellme/issues/74) so it outlives the package freeze; (c) `oneLine` relocated into `toolcall.go` (orphan `toollog.go` deleted).
- **Round-035 forward items** — (a) `[TECHNICAL DEBT]` the observer port hook overload (`BeforeToolLog`/`AfterToolLog` serving two resume policies) → [#69](https://github.com/gosharplite/tellme/issues/69); (b) `[REFACTOR]` the turn's spinner-yield policy has **three homes / four call sites** with **no named owner** → [#69](https://github.com/gosharplite/tellme/issues/69) (see the round-035 detail in [`2026-09-17.md`](docs/archives/status/2026-09-17.md)).
- **Round-034 forward items** — (a) the failed-turn **display-only `Ready` overstatement** (G2) and the **numbering skew** (per-call frames reach `prior + k`, next prompt restarts at `prior + 1`) are recorded divergences; **G2 now also couples to round 035** (TD-2: `final` is load-bearing for both the deferred tail and the round-035 yield) — its round must re-check the round-035 phase-boundary yield; (b) `[TECHNICAL DEBT]` `BindToolOutput` registry mutation → constructor injection at the composition root (→ [#69](https://github.com/gosharplite/tellme/issues/69#issuecomment-5698460385)); (c) `[REFACTOR]` `LoopObserver` interface segregation (would **supersede ADR 0005 G1**); (d) the round-022 row→feature audit blind spot → [#60](https://github.com/gosharplite/tellme/issues/60#issuecomment-5697192786).
- **Round-033 forward items** — loader silently best-effort; recursive walk linear; large-catalog bounded by the resource contract; path-sort ordering.
- **Round-032 forward items** — local stdio MCP transport; cross-invocation tool caching; MCP-backed MEMORY/PLUR; MCP `-d` diagnostic; `mcptest/` → [#13](https://github.com/gosharplite/tellme/issues/13); stale `make help` text.
- **Older forward items** — rounds 018–031 forward items live per-round in the archives.
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** / **no `flock`**; round-011 forward items.
- **Issue tracker (2026-09-17 closeout, session 2)** — [#74](https://github.com/gosharplite/tellme/issues/74) **CLOSED (completed)** — delivered by round 036 (PR [#75](https://github.com/gosharplite/tellme/pull/75) merged `ddd6f7d`); [#69](https://github.com/gosharplite/tellme/issues/69) open (**body extended** to carry the spinner-yield-policy ownership + the port hook pair + the tool-log blank-reason-predicate single ownership + the permanent E2E narrowing record); [#60](https://github.com/gosharplite/tellme/issues/60) open (dogfooding umbrella); [#13](https://github.com/gosharplite/tellme/issues/13) open (coverage tooling). [#72](https://github.com/gosharplite/tellme/issues/72) closed (completed) by round 035.

## Environment notes

- **Dev tooling — `tellme.sh` (external)**: the Niffler-style manager driving the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/{beta-niffler,mbp-johndoe-niffler}/tellme.sh`; invoke via `source tellme.sh` or the `tm` alias. Its banner is round-agnostic (current-state pointer = this `STATUS.md`).
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` (**refreshed** from the merged round-036 head `11356db`). Sandbox: `unshare -n` unavailable → the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (`verify-mcp-sdk-confinement`); the hermetic fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020).
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, `~/.bashrc`; read for `$TELL_ME_HOME`; read for `…/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` unavailable for this repo; closeout scans are diff-level pattern greps — **clean this closeout**.
