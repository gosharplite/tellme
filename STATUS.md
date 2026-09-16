# tellme — Status

**Last updated**: 2026-09-16 (day close, session 13: **round 030 `030-provider-truncation-guard` OPEN** — plan + truth + **implementation** delivered on the round branch; **PR [#63](https://github.com/gosharplite/tellme/pull/63) awaiting human merge**). Round-029 detail relocated to the archive (Rule 12); rounds 001–029 live in the archives. **#60** = the dogfooding-enablement umbrella; **#62** = the transport-hardening slice (this round).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation this phase)
**Active branch**: `030-provider-truncation-guard` (round open; PR [#63](https://github.com/gosharplite/tellme/pull/63) pending human merge; `dev` is the integration line)
**Daily log**: [`docs/session-summary/2026/09/16/session-summary.md`](docs/session-summary/2026/09/16/session-summary.md)
**Archive**: [`2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012) · [`2026-09-14.md`](docs/archives/status/2026-09-14.md) (rounds 013–019) · [`2026-09-15.md`](docs/archives/status/2026-09-15.md) (rounds 020–026) · [`2026-09-16.md`](docs/archives/status/2026-09-16.md) (rounds 027–029).

## Round 030 — `030-provider-truncation-guard` (open — plan + truth + implementation; PR #63 awaiting merge)

**Status**: 🔵 **OPEN / AWAITING HUMAN MERGE** — branch `030-provider-truncation-guard` (off `dev`); PR [#63](https://github.com/gosharplite/tellme/pull/63) carried through **three review rounds** (round-1 fold → re-review **fold ACCEPTED** → residual fold → **fold ACCEPTED — review loop CLOSED, no open findings**) and then the full **implementation** (`/axb-implement`, T001–T015 all `[X]`). Head **`ca873d9`**. **Resolves issue [#62](https://github.com/gosharplite/tellme/issues/62)** on merge. The provider-transport **truncation guard**: the transports read the **finish reason** and turn an **output-cap truncation** into a loud provider failure instead of returning a (possibly corrupted) reply — the silent data-corruption hazard round 029 (multi-KB tool args) exposed.

**Scope**: OpenAI-compatible `choices[0].finish_reason == "length"`; Vertex/Gemini `candidates[0].finishReason == "MAX_TOKENS"`. **Universal** trigger (tool call **or** text); the failure reuses the frozen `the provider request failed` + exit **6** (**no** new class phrase, **no** new exit code); the Gemini detail is **function-call-aware** (names the tool); the guard runs **before** the generic "no usable answer" check (a recorded divergence); **no** retry layer; the **request side is unchanged** (unset `MAX_TOKENS` → provider default).

**Pipeline**: specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · dsl-refine ✅ · tasks ✅ · **implement ✅** (T001–T015 all `[X]`); merge ⏳ (human-only).

**Artifacts**: `spec.md` (US1 tool call · US2 text · FR-001–010 · NFR-001–002 · SC-001–004), `checklists/requirements.md` (2 clarify, both Option 1), `features/acceptance/refusing-a-cut-off-reply.feature` (2 Rules), `research.md` (D1–7 incl. TD-1 usage loss · TD-3 ordering divergence · TD-4 `MALFORMED_FUNCTION_CALL`), `plan.md` (1 interface → `/axb-dsl-refine`; api/data NOOP; ui skipped), `tasks.md` (T001–T015; orphan sweep 0), `truth-delta.md`. **Truth**: `techstack.md` MODIFY (Provider output-cap truncation guard row + adapter/normalization/port rows; the write-tools row's `#62` forward item → "delivered"); `chat/refusing-a-cut-off-reply.feature` **ADD** (2 Rules · 5 Examples — both families at both truncation sites, incl. a Gemini `functionCall`) + `chat/dsl.md` **MODIFY** (5 Given + 2 Then + note) + `chat/reporting-a-failed-provider-request.feature` **MODIFY** (doc-only header cross-reference); `contracts/**` + `data/**` NOOP. **Code**: `internal/infrastructure/llm/openai/client.go` (reads `finish_reason`; `checkTruncation` + `openAIFinishReasonLength`), `internal/infrastructure/llm/gemini/client.go` (reads `finishReason`; `geminiTruncationError` / `geminiFunctionCallTruncationError` + `geminiFinishReasonMaxTokens`); tests: `openai/truncation_test.go` + `gemini/truncation_test.go` + the fake-provider truncation seam (`tests/e2e/fakeprovider`) + 7 `step_r030_*` stepdefs.

**Verification (2026-09-16)**: topology audit **PASSED** (42 features · 16 root + **273** module rows · **1403** steps). `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `golangci-lint` 0 issues · `govulncheck` clean); full E2E `ok` (**193 scenarios**, 0 undefined) + unit pins green. **Falsifiability witnesses** reproduced then reverted — (a) disable the OpenAI guard → the 3 OpenAI cut-off scenarios fail · (b) Gemini fires only on function-calls → the cut-off **answer** scenario + 2 unit pins fail · (c) fire on the healthy *absent* finish reason → many existing scenarios fail. `stdout` byte-exact; `go.mod`/`go.sum` unchanged (stdlib-only).

**Review trail (PR #63)**: plan+truth **APPROVED WITH REQUIRED FOLDS** (B1 Gemini `functionCall` Example · B2 truth-delta NOOPs · TD-1..TD-4 · RF-1/RF-2/N1) → fold `71b5d3e` → re-review **fold ACCEPTED** (residuals R-1/R-2) → fold `008d08b` → **fold ACCEPTED — review loop CLOSED** → implementation `ca873d9`.

**Commits**: `880d3bb` (plan package + spec) · `62f9a0c` (acceptance + research + techstack) · `19d80e4` (system-analysis plan) · `5ea4c23` (interface truth) · `5b9fdc4` (tasks) · `71b5d3e` (review fold: B1/B2/TD/RF/N) · `008d08b` (re-review fold: R-1/R-2) · `ca873d9` (implementation).

**Propagation**: ⏳ **PENDING** — `030-provider-truncation-guard → dev → main` runs **after a human merges PR [#63](https://github.com/gosharplite/tellme/pull/63)** (operator: human-only merge).

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013–019 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md); 020–026 in [`2026-09-15.md`](docs/archives/status/2026-09-15.md); 027–029 in [`2026-09-16.md`](docs/archives/status/2026-09-16.md)); 029 is the most recent **delivered** round.

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `029-agent-write-tools` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history. |
| `030-provider-truncation-guard` | **open** | The current round's working branch off `dev`; PR [#63](https://github.com/gosharplite/tellme/pull/63) **awaiting human merge**; not yet propagated. |
| (next) `031-*` | not started | The next round's working branch off `dev` (after 030 merges). |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation history**: rounds 026/027/028/029 — **DONE (no-ff)**. **Round 030 — PENDING (human merge of PR [#63](https://github.com/gosharplite/tellme/pull/63)).**
> Read live heads with `git rev-parse --short main dev 030-provider-truncation-guard`.

## Roadmap — next slices

> **Direction (2026-09-15)** — no security · no Windows · **bash-first** · a deliberately small tool surface (see [`README.md`](README.md#-design-intent--direction-operator-declared)). Recent slices: **029** ([#60](https://github.com/gosharplite/tellme/issues/60) track) = the first agent write tools; **030** ([#62](https://github.com/gosharplite/tellme/issues/62)) = the provider-transport truncation guard.

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–029** | — | Provider-registry completeness → … → the agent write tools. | ✅ **Delivered** (see the delivered-rounds index) |
| **030 provider-transport truncation guard** | [#62](https://github.com/gosharplite/tellme/issues/62) | Read the finish reason in both transports; fail an output-cap truncation (`length`/`MAX_TOKENS`) as a loud provider error (frozen `the provider request failed` + exit 6); universal trigger; function-call-aware Gemini. | 🔵 **Open** — round 030 (PR [#63](https://github.com/gosharplite/tellme/pull/63) awaiting human merge) |
| **future slices (candidates)** | [#60](https://github.com/gosharplite/tellme/issues/60) · [#13](https://github.com/gosharplite/tellme/issues/13) | the **dogfooding track** ([#60](https://github.com/gosharplite/tellme/issues/60)) skills/context rounds; **coverage tooling** ([#13](https://github.com/gosharplite/tellme/issues/13)); plus the carried forward items below. | ⏳ **Candidates** (not started) |

## Open items (non-blocking)

- **Issue tracker (2026-09-16 closeout)** — **#13** open (coverage tooling; still accurate); **#60** open (dogfooding-enablement umbrella); **#62** open (this round's slice — **to be closed on merge of PR [#63](https://github.com/gosharplite/tellme/pull/63)**). `#47` `not_planned`, `#53`/`#55` completed (earlier). **No closes/revises this closeout** (round 030 not yet merged).
- **Round-030 forward items** — (a) **#62** is implemented by round 030 but stays **open until PR #63 merges**; (b) a truncation failure **accounts no usage** for the call (TD-1, recorded); (c) the guard runs **before** the generic "no usable answer" path (TD-3 — a deliberate divergence from the reference); (d) Vertex/Gemini `MALFORMED_FUNCTION_CALL` is **out of scope** (TD-4, a tracked forward item).
- **Round-029 forward items** — (a) `write_file` survival to be **measured** via `--tool-usage` after dogfooding; (b) `replace_text` reads the **whole** file (an unbounded input read — a forward item); (c) **TD-3** — `replace_text` through a symlink replaces the link. (Round-029 detail → [`2026-09-16.md`](docs/archives/status/2026-09-16.md).)
- **Older forward items** — rounds 018–028 forward items are recorded per-round in the archives (`2026-09-15.md` for 020–026; `2026-09-16.md` for 027–029).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3** renderer lifecycle deferred; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items.

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin` (run with `$GOPATH/bin` on `PATH`); the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme` (refreshed round 029 — carries the two write tools; **round 030 unmerged, so the guard is not yet in the installed `dev`-line binary**). Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Host (this session)**: **Linux** (`…/beta-niffler/ait-tellme`); the **darwin** path is the cross-compile weak spot here (round 020). The other dev host (`…/mbp-johndoe-niffler/`) is macOS and mirrors the opposite.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias); read for `$TELL_ME_HOME` — the vendored skill tree; read for `/home/pos/tmp/github/gosharplite/tell-me-go` and `…/aixbdd-tmg` (bootstrap reference trees; registered read-only this session).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`).
