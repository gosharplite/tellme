# tellme — Status

**Last updated**: 2026-09-16 (day close, session 17: **round 032 `032-mcp-client` — DELIVERED / FROZEN**; PR [#66](https://github.com/gosharplite/tellme/pull/66) **merged** into `dev` (`4376f79`)). tellme ships a **remote (Streamable HTTP) MCP client** (plan + truth + implementation); the **SC-002 live check** found and fixed the `MCP_SERVERS` `${VAR}` gap ([#67](https://github.com/gosharplite/tellme/issues/67)). Round-031 detail lives in the archive (Rule 12); rounds 001–030 live in the archives. Open issues: **#60** (dogfooding-enablement umbrella) · **#13** (coverage tooling). **Propagation `dev → main`: PENDING (awaiting approval).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 032 delivered; next: start the `033-*` round via `/axb-specify`)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–031).

## Round 032 — `032-mcp-client` — DELIVERED / FROZEN (merged into `dev`)

**Status**: ✅ **DELIVERED / FROZEN** — PR [#66](https://github.com/gosharplite/tellme/pull/66) **merged** into `dev` (`4376f79`, by `thptcnec`, 2026-09-16T06:20:44Z); frozen head `c370433` (18 commits). A **new-capability** round (operator request *"Let tellme support MCP"*) motivated by the reference's startup stall: `tell-me-go` discovers MCP tools on **every** CLI invocation and `wg.Wait()`s under a 30 s cap, so one offline server (`hf`) delayed every run. tellme now ships the **remote (Streamable HTTP) MCP client** end-to-end (plan + truth + implementation).

**Scope (delivered)**: a typed `MCP_SERVERS` registry, a `tools.MCPClient` domain port, the official SDK adapter confined to `internal/infrastructure/mcp/`, credential resolution (`auto`/`gh`/`bearer`/`basic`/`none`), **non-stall** discovery (a fixed fast-fail bound + per-server `ENABLED`), and deterministic `mcp_<server>_<tool>` names offered **alongside** the native tools. Stdio transport, cross-invocation caching, and MEMORY/PLUR remain **deferred** (recorded forward items).

**Pipeline (all ✅)**: specify · spec-by-example (3 journeys) · research (D1–D12) · system-analysis (1 CLI interface; api/data NOOP) · dsl-refine (ADD `chat/using-tools-from-a-remote-mcp-server.feature` + 15 `chat/dsl.md` rows) · tasks (T001–T032) · implement (all `[X]`) · **merged** · **propagated `032 → dev`**; **`dev → main` PENDING (approval)**.

**Decisions locked (round 032)**: operator Q1–Q5 — **Q1 → 1** remote HTTP only; **Q2 → 1+3** fixed fast-fail bound + per-server `ENABLED`; **Q3 → 1** official MCP Go SDK (`v1.7.0`), confined behind the domain port; **Q4 → 1** full auth parity + reference naming; **Q5 → 1** MCP client only. Review folds: **B1–B3** (schema normalization; SDK-built fake in `mcptest/`; bounded injectable `tokenResolver`), **TD1–TD8**, **R1–R8**, the principal-review **tasks.md Phase-3 realignment**, the implementation-review **F1–F9 + N1**, and the **SC-002 live-check fix ([#67](https://github.com/gosharplite/tellme/issues/67))** — `MCP_SERVERS` `${VAR}` expansion (best-effort, non-fatal) + self-diagnosing skip warnings (`FR-022`; research D12).

**Review trail (PR #66)**: plan+truth NOT APPROVED → folds `912010d`/`313e11f`/`0b926c4`/`cb407ed` → CERTIFIED READY → principal review (`tasks.md` DSL drift) → fold `4b4ca46` → CERTIFIED READY FOR IMPLEMENTATION; implementation review → fold `0a3ad9c` (F1–F9) → N1 `a14e6f4`; principal review (**`RoundTrip` `defer cancel()` blocker**) → fold `7a13a53` → CERTIFIED READY TO MERGE; **SC-002 live check** → issue [#67](https://github.com/gosharplite/tellme/issues/67) → fold `2ced555` (`${VAR}` expansion) → VERIFIED → diagnostic fold `c370433` (self-diagnosing skips) → verified, no further review items → **merged**.

**Commits** (branch `032-mcp-client`, then merged): `cd882a2` (spec) · `6c622d2` (acceptance) · `37a12f3` (research+techstack) · `4bd6bad` (system-analysis) · `c63633a` (interface truth) · `c94c5ac` (tasks) · `912010d`/`313e11f`/`0b926c4`/`cb407ed` (plan+truth folds) · `4b4ca46` (principal fold — tasks realignment) · `3fa2a96` (implementation T001–T032) · `0a3ad9c` (impl-review F1–F9) · `a14e6f4` (N1) · `7a13a53` (RoundTrip blocker fold) · `2ced555` (SC-002 `${VAR}` fix, #67) · `c370433` (self-diagnosing skips). Frozen head **`c370433`**.

**Propagation**: `032-mcp-client → dev` **DONE** (PR #66 merge `4376f79`). `dev → main` ⏳ **PENDING** (awaiting approval — Rule 8).

---

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
| 021 | `021-tool-surface-parity` | PR [#48](https://github.com/gosharplite/tellme/pull/48) |
| 022 | `022-tool-loop-log-line` | PR [#50](https://github.com/gosharplite/tellme/pull/50) |
| 023 | `023-interactive-prompt-teardown` | PR [#51](https://github.com/gosharplite/tellme/pull/51) |
| 024 | `024-tool-resource-contract-and-execute-command` | PR [#54](https://github.com/gosharplite/tellme/pull/54) |
| 025 | `025-spinner-width-safety` | PR [#56](https://github.com/gosharplite/tellme/pull/56) |
| 026 | `026-tool-usage-accounting` | PR [#57](https://github.com/gosharplite/tellme/pull/57) |
| 027 | `027-ai-call-turn-counter` | PR [#58](https://github.com/gosharplite/tellme/pull/58) |
| 028 | `028-user-global-prompt-log` | PR [#59](https://github.com/gosharplite/tellme/pull/59) |
| 029 | `029-agent-write-tools` | PR [#61](https://github.com/gosharplite/tellme/pull/61) |
| 030 | `030-provider-truncation-guard` | PR [#63](https://github.com/gosharplite/tellme/pull/63) |
| 031 | `031-tool-schema-wellformedness` | PR [#65](https://github.com/gosharplite/tellme/pull/65) |
| 032 | `032-mcp-client` | PR [#66](https://github.com/gosharplite/tellme/pull/66) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–031 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **032 is the most recent delivered round.**

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `032-mcp-client` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026/027/028/029/030/031 — **DONE (no-ff)**; round **032** — `032-mcp-client → dev` **DONE** (PR #66 merge `4376f79`), **`dev → main` PENDING** (approval).
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **031** ([#64](https://github.com/gosharplite/tellme/issues/64)) = the tool-schema well-formedness fix + gate; **032** = the remote MCP client (delivered) + the SC-002 `MCP_SERVERS` `${VAR}` fix ([#67](https://github.com/gosharplite/tellme/issues/67)).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–031** | — | Provider-registry completeness → … → the tool-schema well-formedness fix + gate. | ✅ **Delivered** (see the delivered-rounds index) |
| **032** | (operator request) · [#67](https://github.com/gosharplite/tellme/issues/67) | The **remote MCP client** (Streamable HTTP) + non-stall discovery; the SC-002 live check found + fixed the `MCP_SERVERS` `${VAR}` gap ([#67](https://github.com/gosharplite/tellme/issues/67)). | ✅ **Delivered** (PR [#66](https://github.com/gosharplite/tellme/pull/66) merged into `dev`) |
| **future slices (candidates)** | [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds; **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Round 032 — delivered** (PR [#66](https://github.com/gosharplite/tellme/pull/66) merged into `dev` `4376f79`; frozen head `c370433`). **`dev → main` propagation: PENDING** (awaiting approval — Rule 8).
- **Round-032 forward items** — (a) **local stdio MCP transport** (`COMMAND`); (b) **cross-invocation tool caching**; (c) **MCP-backed MEMORY/PLUR** integration; (d) **MCP `-d` diagnostic** (non-dialing); (e) `mcptest/` to issue [#13](https://github.com/gosharplite/tellme/issues/13)'s coverage exclusion list; (f) the stale `make help` `verify-no-network` text; (g) a **dedicated credential-resolution bound** (the `gh` spawn shares the discovery fast-fail bound — a deliberate choice; a named bound is an option).
- **Round-031 forward items (recorded on [#60](https://github.com/gosharplite/tellme/issues/60))** — (a) **typed schema construction** ([#60 · 5690778272](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690778272)); (b) **recursive schema walk** ([#60 · 5690604951](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690604951)); (c) **SC-002** — the **manual** live Vertex/Gemini confirmation (open, non-gating).
- **Issue tracker (2026-09-16 closeout, session 17)** — **[#67](https://github.com/gosharplite/tellme/issues/67)** **CLOSED (completed)** — the `MCP_SERVERS` `${VAR}` gap, delivered by round 032 (folds `2ced555`/`c370433`; PR #66 merged `4376f79`); **[#60](https://github.com/gosharplite/tellme/issues/60)** open (dogfooding-enablement umbrella); **[#13](https://github.com/gosharplite/tellme/issues/13)** open (coverage tooling). No revisions.
- **Older forward items** — rounds 018–030 forward items are recorded per-round in the archives (`2026-09-15.md` for 020–026; `2026-09-16.md` for 027–031).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (refreshed this session from the merged round-032 head `c370433` — the MCP client + the SC-002 `${VAR}` fix). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032, delivered)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (a `verify-mcp-sdk-confinement` Makefile gate). The hermetic SDK-built fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — **clean this closeout**.
