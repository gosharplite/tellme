# tellme — Status

**Last updated**: 2026-09-16 (day close, session 19: **round 033 `033-skills-system` — DELIVERED / FROZEN** (PR [#68](https://github.com/gosharplite/tellme/pull/68) merged into `dev` `86bab47`; propagated `dev → main` `d900bd5`)). tellme now ships a **minimal, on-demand skills system** (`list_skills` over `<TELL_ME_HOME>/docs/skills/`; no injection; no skills.sh). Round-032 detail + earlier live in the archives; rounds 001–033 in the delivered index. Open issues: **#60** (dogfooding-enablement umbrella) · **#13** (coverage tooling). **Propagation this session: `033-skills-system → dev` (PR #68 merge `86bab47`) `→ main` (`d900bd5`) — DONE (no-ff).**
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `dev` (round 033 delivered/frozen; the next round starts fresh off `dev`)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–032).

## Round 033 — `033-skills-system` — DELIVERED / FROZEN (merged into `dev`)

**Status**: ✅ **DELIVERED / FROZEN** — PR [#68](https://github.com/gosharplite/tellme/pull/68) **merged** into `dev` (`86bab47`, by `thptcnec`, 2026-09-16T08:25:38Z); frozen head `4e2d96b` (12 commits). A **new-capability** round (operator request: a skills system without the `skillssh` complication).

**Scope (delivered)**: a **minimal, on-demand skills system** — tellme loads skill definitions from `<TELL_ME_HOME>/docs/skills/` (recursive; a skill = a Markdown file with valid `name`/`description` frontmatter; non-skill Markdown skipped; best-effort) and surfaces them via a read-only **`list_skills`** tool (name + description + location, path-sorted). **No injection** (the agent opens a listed skill with the existing `read_files`); **no skills.sh** (`.skills/`, `search/install/remove_skill`) — a recorded divergence from `tell-me-go`.

**Pipeline (all ✅)**: specify · clarify (Q1 → 1 on-demand only; Q2 → moot; Q3 → 1 a `list_skills` tool) · spec-by-example (1 journey) · research (D1–D9) · system-analysis (1 CLI interface → `/axb-dsl-refine`; api/data NOOP; ui skipped) · dsl-refine (ADD `chat/listing-the-available-skills.feature` + 11 `chat/dsl.md` rows; MODIFY `chat/offering-the-agent-tools.feature` six → **seven** tools) · tasks (T001–T024) · implement (T001–T024 all `[X]`) · **merged** · **propagated `033 → dev` → `main`**.

**Decisions locked (round 033)**: single source `<home>/docs/skills`; a skill = `.md` with valid `name`/`description` frontmatter (recursive); on-demand only; loaded on the prompt path only (offline untouched); minimal shape (`internal/domain/skills` type + infra loader; no repository/selector framework); output = name + description + location (path-sorted); `list_skills` = an ordinary agent tool (shared `resourceSchema` + `reason`; reader-class 30 s default; the round-031 gate covers it). **Review folds (PR #68)**: 🔴 catalog→tool wiring seam pinned (**FR-009**; `agentTools()` parameterless + read-free; lazy `Execute`-only seam set in `runTurn`; `newToolRegistry` unchanged); 🔴 `truth-delta.md` `/axb-api-plan` + `/axb-data-plan` explicit **NOOP** rows; 🟡 `plan.md` (spec-by-example ✅ done) + `spec.md` (clarification markers resolved-by-research); nits (sentinel `REFERENCE-SENTINEL`; `runtime home` vocabulary; `internal/home` helper; six→seven comment reconciliation in T022); the **implementation-review fold** reworded the truth to the shipped **silent** best-effort loader (`spec.md` NFR-001 + edge case; `techstack.md` Skills row; `research.md` D2/D5 + rationale), strengthened the loader-test negative, and synced `STATUS.md`. **Principal architecture review: ✅ FULL ARCHITECTURAL APPROVAL — CERTIFIED READY TO MERGE.**

**Commits** (branch `033-skills-system`, then merged): `c4e85b5` (spec) · `829aa1f` (research + techstack) · `182724d` (plan) · `8219150` (acceptance) · `3b619ca` (interface truth) · `b405c60` (tasks) · `c019370` (review fold 1) · `448b2b9` (re-review fold) · `45db613` (cosmetic tidy) · `320a4a9` (implementation T001–T024) · `4e2d96b` (implementation-review fold). Frozen head **`4e2d96b`**.

**Propagation**: `033-skills-system → dev` **DONE** (PR #68 merge `86bab47`) `→ main` **DONE (no-ff, `d900bd5`)**.

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
| 033 | `033-skills-system` | PR [#68](https://github.com/gosharplite/tellme/pull/68) |

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–032 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **033 is the most recent delivered round.**

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `033-skills-system` | **delivered / frozen** | Round 033's working branch — merged into `dev` via PR [#68](https://github.com/gosharplite/tellme/pull/68) (`86bab47`), then propagated `dev → main` (`d900bd5`). |
| `001-*` … `033-skills-system` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026/027/028/029/030/031 — **DONE (no-ff)**; round **032** — `032-mcp-client → dev` (`4376f79`) `→ main` (`5b9d5fd`) — **DONE (no-ff)**; round **033** — `033-skills-system → dev` (PR #68 merge `86bab47`) `→ main` (`d900bd5`) — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **031** ([#64](https://github.com/gosharplite/tellme/issues/64)) = the tool-schema well-formedness fix + gate; **032** = the remote MCP client (delivered); **033** = the minimal on-demand skills system (`list_skills`).

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–032** | — | Provider-registry completeness → … → the remote MCP client. | ✅ **Delivered** (see the delivered-rounds index) |
| **033** | (operator request) | The **minimal, on-demand skills system** — load `<TELL_ME_HOME>/docs/skills/` + a read-only `list_skills` tool (no injection, no skills.sh). | ✅ **Delivered** (PR [#68](https://github.com/gosharplite/tellme/pull/68) merged into `dev`; propagated `dev → main`) |
| **future slices (candidates)** | [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds; **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Round 033 — delivered / frozen** (PR [#68](https://github.com/gosharplite/tellme/pull/68) merged into `dev` `86bab47`; propagated `dev → main` `d900bd5`; frozen head `4e2d96b`).
- **Round-033 forward items** — (a) a very large catalog is bounded by the round-024 resource contract on the tool's **result** (no paging); (b) the `list_skills` result ordering is a fixed path sort; (c) the loader is **silently** best-effort (operator-visible skip diagnostics belong to the `#60` dogfooding track if ever wanted); (d) the recursive walk scales linearly (an mtime cache/indexer could later go behind the `catalog` seam without changing tool contracts).
- **Round-032 forward items** — (a) **local stdio MCP transport** (`COMMAND`); (b) **cross-invocation tool caching**; (c) **MCP-backed MEMORY/PLUR** integration; (d) **MCP `-d` diagnostic** (non-dialing); (e) `mcptest/` to issue [#13](https://github.com/gosharplite/tellme/issues/13)'s coverage exclusion list; (f) the stale `make help` `verify-no-network` text; (g) a **dedicated credential-resolution bound** (the `gh` spawn shares the discovery fast-fail bound — a deliberate choice; a named bound is an option).
- **Round-031 forward items (recorded on [#60](https://github.com/gosharplite/tellme/issues/60))** — (a) **typed schema construction** ([#60 · 5690778272](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690778272)); (b) **recursive schema walk** ([#60 · 5690604951](https://github.com/gosharplite/tellme/issues/60#issuecomment-5690604951)); (c) **SC-002** — the **manual** live Vertex/Gemini confirmation (open, non-gating).
- **Older forward items** — rounds 018–031 forward items are recorded per-round in the archives (`2026-09-15.md` for 020–026; `2026-09-16.md` for 027–032).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.
- **Issue tracker (2026-09-16 closeout, session 19)** — [#60](https://github.com/gosharplite/tellme/issues/60) open (dogfooding-enablement umbrella); [#13](https://github.com/gosharplite/tellme/issues/13) open (coverage tooling). Round 033 has no anchor issue (operator request); its work has landed → **no closes/revises** this closeout.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (to be refreshed from the merged round-033 head `4e2d96b` — the skills system). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **MCP client dependency (round 032, delivered)**: `github.com/modelcontextprotocol/go-sdk` **v1.7.0** (vendored), confined to `internal/infrastructure/mcp/**` behind the `tools.MCPClient` port (a `verify-mcp-sdk-confinement` Makefile gate). The hermetic SDK-built fake lives in `internal/infrastructure/mcp/mcptest/`.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — **clean this closeout**.
