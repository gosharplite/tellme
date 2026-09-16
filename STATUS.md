# tellme — Status

**Last updated**: 2026-09-16 (day close, session 14: **round 030 `030-provider-truncation-guard` DELIVERED / FROZEN** — PR [#63](https://github.com/gosharplite/tellme/pull/63) merged by `thptcnec` (frozen head `1c2dcb6` → `dev` `9ecf845`) and propagated `dev → main`). Round-029 detail lives in the archive (Rule 12); rounds 001–029 live in the archives. **#60** = the dogfooding-enablement umbrella; **#13** = coverage tooling; **#64** = a NEW provider/tool-schema bug (opened this session).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `031-tool-schema-wellformedness` (round 031 **in progress** — off `dev`; rounds 001–030 delivered/frozen; `dev` is the integration line)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–029).

## Round 031 — `031-tool-schema-wellformedness` (IN PROGRESS)

**Status**: ⏳ **IN PROGRESS** — branch `031-tool-schema-wellformedness` (off `dev`); **PR [#65](https://github.com/gosharplite/tellme/pull/65)** open (human-only merge). Resolves issue [#64](https://github.com/gosharplite/tellme/issues/64): a tool-schema `required`/`properties` defect from round 024 — 5 of 6 tools list `reason` in `required` but never declare its property, so a strict provider (Vertex/Gemini) 400s **every** request.

**Pipeline**: specify ✅ · research ✅ · spec-by-example (skipped — no new journey) · system-analysis ✅ (0 interfaces) · api-plan / data-plan / dsl-refine (NOOP) · tasks ✅ · implement ⏳.

**Decisions locked (round 031)**: **Q1 → 1** (hermetic registry unit gate + a **manual** live Vertex/Gemini closeout check; keeps `make verify` offline); **Q2 → 1** (minimal fix — the shared `resourceSchema` declares the `reason` property; `execute_command` left inline, already compliant — plus a for-every-registered-tool `required ⊆ properties` unit gate).

**Review (PR #65)**: plan+truth **APPROVED** with 5 findings to fold — **ARCH-1** (gate reads the non-overridable `agentTools()`, not the `newToolRegistry` DI seam) and **ARCH-3** (T002 must not re-hand-enumerate all six) folded into `tasks.md`; **ARCH-2** (flat-schema precondition + `#60` forward item), **ARCH-4** (this block reconciled), **ARCH-5** (acceptance-carrier divergence recorded) folded into the artifacts.

**Commits**: `0180f9c` (spec) · `cec72db` (research + techstack) · `f3dc75d` (system-analysis) · `c7efcfe` (tasks) · + review-fold commit.

**Next**: `/axb-implement` (T001–T007) → round review → manual live Vertex/Gemini confirmation → closeout.

## Round 030 — `030-provider-truncation-guard` (DELIVERED / FROZEN)

**Status**: ✅ **DELIVERED / FROZEN** — branch `030-provider-truncation-guard` (off `dev`); PR [#63](https://github.com/gosharplite/tellme/pull/63) carried through **three plan+truth review rounds** + an **implementation review** (`APPROVED`, two doc/perf notes) + a **principal architecture review** (**CERTIFIED READY TO MERGE**) — every fold accepted, no open findings — then **merged** by `thptcnec` (2026-09-16T01:05:42Z) as `dev` `9ecf845`; frozen head **`1c2dcb6`** (10 commits). **Resolves issue [#62](https://github.com/gosharplite/tellme/issues/62)** (now `closed`). The provider-transport **truncation guard**: the transports read the **finish reason** and turn an **output-cap truncation** into a loud provider failure instead of returning a (possibly corrupted) reply — the silent data-corruption hazard round 029 (multi-KB tool args) exposed.

**Scope**: OpenAI-compatible `choices[0].finish_reason == "length"`; Vertex/Gemini `candidates[0].finishReason == "MAX_TOKENS"`. **Universal** trigger (tool call **or** text); the failure reuses the frozen `the provider request failed` + exit **6** (**no** new class phrase, **no** new exit code); the Gemini detail is **function-call-aware** (names the tool); the guard runs **before** the generic content-empty "no usable answer" check (a recorded divergence); **no** retry layer; the **request side is unchanged** (unset `MAX_TOKENS` → provider default).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · implement ✅ (T001–T015 all `[X]`) · **merged ✅** · **propagated ✅**.

**Review trail (PR #63)**: plan+truth **APPROVED WITH REQUIRED FOLDS** (B1 Gemini `functionCall` Example · B2 truth-delta NOOPs · TD-1..TD-4 · RF-1/RF-2/N1) → fold `71b5d3e` → **fold ACCEPTED** (R-1/R-2) → fold `008d08b` → **review loop CLOSED** → implementation `ca873d9` → **impl review APPROVED** (N-1 comment accuracy · N-2 `Complete`-layer `ProviderError` type pin) → fold `1c2dcb6` → **fold ACCEPTED** → **principal review CERTIFIED READY TO MERGE**.

**Commits**: `880d3bb` (plan package + spec) · `62f9a0c` (acceptance + research + techstack) · `19d80e4` (system-analysis plan) · `5ea4c23` (interface truth) · `5b9fdc4` (tasks) · `71b5d3e` (review fold: B1/B2/TD/RF/N) · `008d08b` (re-review fold: R-1/R-2) · `ca873d9` (implementation) · `da969de` (day-close docs) · `1c2dcb6` (impl-review fold: N-1/N-2) → PR [#63](https://github.com/gosharplite/tellme/pull/63) merge into `dev` (`9ecf845`).

**Propagation**: ✅ **DONE (no-ff)** — `030-provider-truncation-guard → dev` (PR #63, `9ecf845`) `→ main`.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–029 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); **030 is the most recent delivered round**.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `030-provider-truncation-guard` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history. |
| (next) `031-*` | not started | The next round's working branch off `dev`. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026/027/028/029/030 — **DONE (no-ff)**.
> Read live heads with `git rev-parse --short main dev`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **029** ([#60](https://github.com/gosharplite/tellme/issues/60) track) = the first agent write tools; **030** ([#62](https://github.com/gosharplite/tellme/issues/62)) = the provider-transport truncation guard.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–030** | — | Provider-registry completeness → … → the provider-transport truncation guard. | ✅ **Delivered** (see the delivered-rounds index) |
| **tool-schema fix (candidate 031)** | [#64](https://github.com/gosharplite/tellme/issues/64) | **`required` lists `reason` but `properties` never declares it** on 5 of 6 agent tools — Vertex/Gemini rejects the request (`400`); declare the `reason` property + add a `required ⊆ properties` well-formedness gate for every registered tool. | 🆕 **Filed** (found by real-endpoint dogfooding; not started) |
| **future slices (candidates)** | [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds; **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-16 closeout)** — **[#62](https://github.com/gosharplite/tellme/issues/62) CLOSED (completed)** — delivered by round 030 (PR [#63](https://github.com/gosharplite/tellme/pull/63) merged `9ecf845`); **[#64](https://github.com/gosharplite/tellme/issues/64)** open (**NEW this session** — tool-schema `required`/`properties` defect; candidate round 031); **[#60](https://github.com/gosharplite/tellme/issues/60)** open (dogfooding-enablement umbrella); **[#13](https://github.com/gosharplite/tellme/issues/13)** open (coverage tooling; still accurate). `#47` `not_planned`, `#53`/`#55` completed (earlier).
- **New this session — [#64](https://github.com/gosharplite/tellme/issues/64)** (tool-schema bug) — real-endpoint dogfooding surfaced that Vertex/Gemini 400s **every** request: `required fields ['reason'] are not defined in the schema properties`. Root cause: the shared `resourceSchema` helper (introduced round-024 fold `cfa005c`, renamed round-029 fold `eb0367c`) emits `properties:{<tool props>, max_output_tokens, timeout}` + `required:[…,"reason"]` but **never declares a `reason` property** — so 5 of 6 tools violate `required ⊆ properties` (only `execute_command` complies, built inline). OpenAI-compatible tolerates it; Vertex does not. Fix + a well-formedness gate scoped in the issue.
- **Round-030 forward items** — (a) a truncation failure **accounts no usage** for the call (TD-1, recorded); (b) the guard runs **before** the generic content-empty "no usable answer" path (TD-3 — a deliberate divergence from the reference); (c) Vertex/Gemini `MALFORMED_FUNCTION_CALL` is **out of scope** (TD-4, a tracked forward item); (d) the guard matches `"length"`/`"MAX_TOKENS"` **exactly** (a proxy variant spelling would evade it — recorded forward item).
- **Round-029 forward items** — (a) `write_file` survival to be **measured** via `--tool-usage` after dogfooding; (b) `replace_text` reads the **whole** file (an unbounded input read — a forward item); (c) **TD-3** — `replace_text` through a symlink replaces the link. (Round-029 detail → [`2026-09-16.md`](docs/archives/status/2026-09-16.md).)
- **Older forward items** — rounds 018–028 forward items are recorded per-round in the archives (`2026-09-15.md` for 020–026; `2026-09-16.md` for 027–029).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (refreshed this session from `1c2dcb6` — the merged round-030 head, so it carries the truncation guard). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Real-endpoint note (this session)**: dogfooding a **Vertex/Gemini** provider (`gemini-3.8-flash`) surfaced bug [#64](https://github.com/gosharplite/tellme/issues/64) — the tool-schema defect fails **every** request with a `400`, so the Vertex path is currently unusable against a real endpoint (the hermetic fakes never schema-validate, which is why rounds 024–030 stayed green). The OpenAI-compatible path (e.g. `deepseek-flash`) is unaffected.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees; registered read-only this session).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
