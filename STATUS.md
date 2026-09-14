# tellme — Status

**Last updated**: 2026-09-14 (session 4) — **round 015 plan + truth reviewed APPROVED; `/axb-tasks` delivered** (`015-interactive-tui-prompt`, anchor [#37](https://github.com/gosharplite/tellme/issues/37)); PR [#38](https://github.com/gosharplite/tellme/pull/38) **open → `dev`** (plan + truth only — not ready to merge); the architectural review ([#5657097770](https://github.com/gosharplite/tellme/pull/38#issuecomment-5657097770)) certified the plan + truth at head `c66f500` (**PLAN + TRUTH APPROVED — PROCEED WITH ARCHITECTURAL DIRECTIVES**, 1 blocker + 4 directives); **`tasks.md` is now delivered** (40 tasks, embedding those directives); next is `/axb-implement`. The upstream gate [aixbdd-tmg#13](https://github.com/gosharplite/aixbdd-tmg/issues/13) is **RESOLVED** (PR [#14](https://github.com/gosharplite/aixbdd-tmg/pull/14) merged) — the `/axb-ui-plan` terminal-mode step is **un-gated**. Round 014 remains **DELIVERED / FROZEN** (see the *Round 014* section). Prior rounds' detail is in the archives ([2026-09-11](docs/archives/status/2026-09-11.md), [2026-09-13](docs/archives/status/2026-09-13.md), [2026-09-14](docs/archives/status/2026-09-14.md)).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `015-interactive-tui-prompt` — round 015 **IN PROGRESS** (plan + truth done; implementation next). Branch pushed; PR [#38](https://github.com/gosharplite/tellme/pull/38) **open → `dev`**.
**Daily log**: [`docs/session-summary/2026/09/14/session-summary.md`](docs/session-summary/2026/09/14/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) (rounds 001–002 + grill/clarify history) · [`docs/archives/status/2026-09-13.md`](docs/archives/status/2026-09-13.md) (rounds 003–012 detail + header/review-response/propagation history) · [`docs/archives/status/2026-09-14.md`](docs/archives/status/2026-09-14.md) (round 013 detail).

## Round 015 — `015-interactive-tui-prompt` (in progress — plan + truth done)

**Status**: 🔄 **IN PROGRESS — PLAN + TRUTH REVIEWED; `tasks.md` DELIVERED** (2026-09-14, session 4) — theme anchored at issue [#37](https://github.com/gosharplite/tellme/issues/37). The **plan + truth half** landed on branch `015-interactive-tui-prompt` (pushed); PR [#38](https://github.com/gosharplite/tellme/pull/38) is **open → `dev`** (plan + truth only — not ready to merge). The plan + truth were **architecturally reviewed APPROVED**; `tasks.md` is delivered. Remaining: `/axb-implement`.

**Architectural review (2026-09-14, session 4)** — PR [#38](https://github.com/gosharplite/tellme/pull/38) review ([#5657097770](https://github.com/gosharplite/tellme/pull/38#issuecomment-5657097770]), evaluated head **`c66f500`** (== current HEAD): **PLAN + TRUTH APPROVED — PROCEED TO IMPLEMENTATION WITH ARCHITECTURAL DIRECTIVES**. Directives, embedded in `tasks.md`:

1. **🔴 BLOCKER (implementation)** — **TUI output stream containment**: bind Bubble Tea to the diagnostic stream (`tea.WithOutput(env.stderr)` + `tea.WithInput(env.stdin)`); `stdout` stays byte-exact for piping (FR-015). The TUI runner must write **only** to `env.stderr` (or the injected test writer).
2. **🟡 TECHNICAL DEBT — package/domain-boundary alignment**: `plan.md`'s `internal/domain/config/config.go` and `internal/domain/ports/` are **stale** — modify `internal/config/config.go` (never an artificial `internal/domain/config`); place ports in focused subdomains (`internal/domain/suggestions/`, `internal/domain/history/…`), the coordinator at `internal/app/suggestions/service.go`, the adapter at `internal/infrastructure/history/global_prompt_tracker.go`. No centralized `internal/domain/ports/`.
3. **🟡 TECHNICAL DEBT — suggestion-engine I/O bounds**: scope traversal to `filepath.Split(query)`, read directory entries in bounded chunks (≤100), honour `ctx.Err()` for debounce abort, skip ignore-listed dirs (`.git`, `node_modules`), stop at 10 candidates (no unbounded `WalkDir`).
4. **🔵 REFACTOR — TUI-runner DI seam** in `cli.go`: a `tuiPromptRunner` factory var (mirroring `gatewayFactory`/`historyStoreFactory`) so the `-i`/`USE_TUI_PROMPT`/non-TTY dispatch matrix is unit-testable without a terminal loop.
5. **🔵 REFACTOR — optimistic concurrency / async append safety**: `Append` opens `O_APPEND|O_CREATE|O_WRONLY`; compaction runs asynchronously and is size-snapshot-checked (`newSize == initialSize`) before `AtomicWrite`, with backoff retry; `GlobalPromptTracker` + `SuggestionService` implement `Close(ctx) error` draining via a `sync.WaitGroup`.

**Scope**: add the **`-i` / `--interactive` Interactive TUI Prompt** (re-creating `tell-me-go`'s TUI) — a live **suggestion engine** (history + filesystem + tools), a **session dashboard** (tokens / turns / provider), a multi-line editor, and terminal keybindings. Distinct from round 012's *plain* multi-line reader; the **non-TTY fallback is preserved**.

**Compatibility requirement (shared Niffler env)**: the suggestion engine must interoperate with `$TELL_ME_HOME/output/global_prompts.jsonl` — the shared global prompt log at the `output/` **root** (format `{"timestamp":"<RFC3339>","prompt":"<text>"}`, append-only, dedupe, newest-first; source: `tell-me-go`'s `globalPromptTracker`). A **data-model** addition → `/axb-data-plan`.

**Gating dependency — RESOLVED**: the plan-side **TUI UI artifact** (`ui/`) needed the upstream **`axb-ui-plan` terminal mode**. [aixbdd-tmg#13](https://github.com/gosharplite/aixbdd-tmg/issues/13) is **closed/completed** by [PR #14](https://github.com/gosharplite/aixbdd-tmg/pull/14) (merged): `axb-ui-plan` now selects **HTML mode** (`ui/*.html`) / **terminal mode** (`ui/screens/*.txt`, no HTML) / **skipped** by interface surface, with `PrototypeMedium` + `Prototype.medium` in the domain model. The vendored skills are **synced** — the `/axb-ui-plan` terminal-mode step for the `ui/` artifact is **un-gated**; the rest of the pipeline was never gated.

**Clarify Round 1 (locked `1,1,1`)**: (Q1) the shared log is **read + written, recorded only under `-i`**; (Q2) the TUI **coexists** with the round-012 plain reader (default unchanged; `-i`/`USE_TUI_PROMPT` opts in); (Q3) **POSIX-only**.

**Artifacts / pipeline** — plan + truth done, implementation next:
- [x] plan package: `spec.md`, `checklists/requirements.md`, `features/acceptance/**` ×4, `ui/ui-plan.md` + `ui/screens/*.txt` ×5 (terminal mode), `research.md` (D1–9), `plan.md` (CLI end + TUI UX surface + shared prompt log; 1 wave; `/axb-api-plan` = NOOP; `/axb-data-plan` = ADD), `truth-delta.md`.
- [x] truth: `specs/truth/techstack.md` MODIFY (TUI prompt + suggestion engine + shared log + dashboard + TUI test harness); `specs/truth/data/data-model.dbml` ADD (`prompt_log_entry`; project → `tellme_local_state`); `specs/truth/features/cli/chat/**` ADD ×4 + `chat/dsl.md` MODIFY (+16 rows); root `cli/dsl.md` + `contracts/**` NOOP.
- [x] `/axb-tasks` — `tasks.md` (40 tasks; Setup = the Bubble Tea family; Phase 3 Test Alignment = 16 `[BDD-RED]` + 4 `[UNIT]` + subagent review; 4 ADD Feature phases + regression; Pre-Delivery orphan sweep 0) — embeds the PR #38 review directives.
- [ ] implementation (`/axb-implement`) — carries the review directives (TUI stream containment → `env.stderr`; focused-subdomain ports; bounded suggestion I/O; `tuiPromptRunner` DI seam; async append + `Close(ctx)`).

**Verification (2026-09-14)**: Gherkin/DSL topology audit **PASSED** (29 features · 11 root + 147 module DSL rows · **690 steps**, 0 errors). No product code this half → `make verify` N/A.

**Open (non-blocking)**: the `/axb-tasks` → `/axb-implement` implementation determinations (the TUI package layout; the exact injected-I/O + suggestion-source seams); the carried round-006/011 items below.

## Round 014 — `014-session-replay-fidelity` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-14) — PR [#35](https://github.com/gosharplite/tellme/pull/35) **MERGED** into `dev` (`e4dac2d`, by `thptcnec`); propagated `dev → main` (`e523c59`). Round-014 head frozen at **`e9239a8`** (the re-certified SHA). `make verify` OK · godog **90/90** · topology audit **PASSED** (609 steps).

**Scope**: make a **resumed** session replay its tool steps faithfully by persisting the per-tool-step provider **signature** (the Gemini 3 `thoughtSignature`). `history.Step` gains an optional, provider-agnostic `Signature`; `AgentLoop` records it per step and `BuildMessages` replays it into the synthesised `llm.ToolCall`. Behaviour intent **MODIFY** (the persisted tool-step record) + **ADD** (the resume-with-tools contract). Anchor issue [#34](https://github.com/gosharplite/tellme/issues/34) — the primary 014 candidate.

**Clarify Round 1 (locked `1,1`)**: (Q1) a dedicated nullable, provider-agnostic **`signature`** step field (no generic metadata container); (Q2) an absent signature on resume keeps the **unchanged best-effort** replay (`the provider request failed`, exit 6) — no pre-flight detection.

**Review / re-certification trail**: FULL ARCHITECTURAL APPROVAL at `cb146bf` (one non-blocking forward consideration — a name-keyed replay map in the E2E helper) → fixed in-round `e9239a8` (match replayed calls **by order**) → **RE-CERTIFIED at `e9239a8`** → merged `e4dac2d` → propagated `e523c59`.

**Artifacts / pipeline** — all phases **done**:
- [x] plan package: `spec.md`, `checklists/requirements.md`, `research.md` (Decisions 1–8), `plan.md` (CLI end + session-history store; 1 wave; `/axb-api-plan` = NOOP; `/axb-data-plan` = MODIFY; CLI end → `/axb-dsl-refine`), `features/acceptance/**` ×2, `tasks.md` (12 tasks; Setup omitted — stdlib-only; orphan sweep 0), `truth-delta.md`.
- [x] truth: `specs/truth/data/data-model.dbml` MODIFY (`history_step` + nullable `signature`; record Note reconciled); `specs/truth/techstack.md` MODIFY (Session history store / Agent tool loop / Vertex-Gemini adapter / E2E runner / Local fake provider / Pure-helper unit tests); `specs/truth/features/cli/chat/replaying-a-tool-using-conversation.feature` ADD + `chat/dsl.md` MODIFY (+ `history/dsl.md` note); root `cli/dsl.md` + `contracts/**` NOOP.
- [x] implementation: `internal/domain/history` (`Step.Signature`, `omitempty`); `internal/agent/agentloop.go` (record `Signature: tc.Signature`; replay `Signature: s.Signature`); unit tests (`file_store_widened_test.go`, `agentloop_test.go`); 4 E2E step files + 2 scenarios.

**Verification (2026-09-14)**: `make verify` **OK** (0 lint · 0 reachable vulns · no test-sleep · offline witness) · `go test ./...` green · godog **90/90 scenarios · 633/633 steps** · topology audit **PASSED** (609 steps) · **falsifiability witness** reproduced (detaching the replayed signature fails the resume scenario) · `go.mod`/`go.sum` unchanged (stdlib-only).

**Open (non-blocking)**: the remaining issue-[#34](https://github.com/gosharplite/tellme/issues/34) candidates (the Google Gemini API family, Application Default Credentials, concurrent tool-call matching) are future slices; the round-014 E2E helper now matches replayed calls by **order** (the review forward-item resolved). Carried: PR #16 **Obs 1** (stdout TTY probe) OPEN; round-006 **Obs 3** (renderer lifecycle) deferred; sequential tool execution / no pruning / no `flock`; round-011 forward items (estimation-heuristic constants; persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`). The next round starts a fresh `015-*` off `dev`.

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

Per-round detail lives in the archives (001–002 in [`2026-09-11.md`](docs/archives/status/2026-09-11.md); 003–012 in [`2026-09-13.md`](docs/archives/status/2026-09-13.md); 013 in [`2026-09-14.md`](docs/archives/status/2026-09-14.md)).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-*` … `014-session-replay-fidelity` | delivered / frozen | Each round's working branch — merged into `dev` via its PR, then propagated `dev → main`; frozen history (never receives post-round commits). |
| `015-interactive-tui-prompt` | in progress (plan + truth) | Round-015 working branch; plan + truth landed and pushed; PR [#38](https://github.com/gosharplite/tellme/pull/38) open → `dev`; implementation next. |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`; only a human merges the PR. Propagation is the no-ff merge `dev → main`.
> **Propagation (round 014):** `014-session-replay-fidelity → dev` (PR [#35](https://github.com/gosharplite/tellme/pull/35), `e4dac2d`) `→ main` (`e523c59`) — DONE; closeout docs on `dev`.
> **Propagation (2026-09-14 session 2 closeout):** `dev → main` — DONE (docs-only; round-015 scoping + issue #37).
> **Propagation (round 015):** **PENDING** — PR [#38](https://github.com/gosharplite/tellme/pull/38) open → `dev` (plan + truth only); not merged; `main` unchanged.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003–012** | — | Provider-registry completeness → first reasoning turn → stdin piping → rendered output + `-r` → session-history persistence → agent tools & the tool-call loop → payload status line → stream-ordering observability → persona on the wire + wire-faithful estimate → interactive multi-line prompt capture. | ✅ **Delivered** (see the delivered-rounds index) |
| **013 — Vertex AI / Gemini provider support** | [#32](https://github.com/gosharplite/tellme/issues/32) | Drive a `TYPE: gemini` Vertex provider end-to-end (Vertex `:generateContent` transport + stdlib service-account OAuth2), no new module. | ✅ **Delivered** (PR [#33](https://github.com/gosharplite/tellme/pull/33); propagated `013 → dev → main`) |
| **014 — Session-replay fidelity** | [#34](https://github.com/gosharplite/tellme/issues/34) | Persist the per-tool-step provider signature (Gemini `thoughtSignature`) so a **resumed** session replays its tool steps faithfully — a data-model change. | ✅ **Delivered** (PR [#35](https://github.com/gosharplite/tellme/pull/35); propagated `014 → dev → main`) |
| **015 — Interactive TUI prompt (`-i`)** | [#37](https://github.com/gosharplite/tellme/issues/37) | Rich terminal prompt (Bubble Tea): suggestion engine (history + FS + tools), session dashboard, multi-line editor, keybindings; compatible with the shared `output/global_prompts.jsonl`. Upstream gate [aixbdd-tmg#13](https://github.com/gosharplite/aixbdd-tmg/issues/13) **RESOLVED** (PR #14) — `/axb-ui-plan` terminal mode available. | 🔄 **In progress** (plan + truth reviewed APPROVED; `tasks.md` delivered; PR [#38](https://github.com/gosharplite/tellme/pull/38)) |
| **future slices (candidates)** | [#36](https://github.com/gosharplite/tellme/issues/36) | The remaining #34 candidates — the **Google Gemini API family** (inline key), **Application Default Credentials**, and **concurrent tool-call matching**; plus the carried forward items below. | ⏳ **Candidate** (not started) |

## Open items (non-blocking)

- **Round 015 (`015-interactive-tui-prompt`)** — **IN PROGRESS (plan + truth reviewed APPROVED; `tasks.md` delivered)** ([#37](https://github.com/gosharplite/tellme/issues/37)); PR [#38](https://github.com/gosharplite/tellme/pull/38) open → `dev` (plan + truth only). The plan + truth were reviewed APPROVED at head `c66f500` (PR #38 review [#5657097770](https://github.com/gosharplite/tellme/pull/38#issuecomment-5657097770)) with **1 implementation blocker + 4 directives** — embedded in `tasks.md`. Implementation (`/axb-implement`) is the next step. Upstream gate [aixbdd-tmg#13](https://github.com/gosharplite/aixbdd-tmg/issues/13) **RESOLVED** (PR #14).
- **Future-slice candidates** — issue [#36](https://github.com/gosharplite/tellme/issues/36): the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials for Vertex, and concurrent tool-call matching (the round-014 E2E helper matches replayed calls by **order** — ready for it).
- **Carried forward items** — PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3** renderer lifecycle deferred to multi-turn; sequential tool execution / **no pruning** (a settled exclusion) / **no `flock`**; round-011 forward items (**estimation-heuristic constants**; the persona-plumbing seam; **N-2** estimator ignores replayed tool-call `arguments`).
- **Future-package candidates** (list refreshed 2026-09-12): ~~(a) run `make verify` in a pipeline platform~~ (withdrawn — [#15](https://github.com/gosharplite/tellme/issues/15) closed `not_planned`; the gate stays manual); ~~(b) F9 flag-parsing units~~ (folded into round 005); ~~(c) `tellme init`~~ (dropped — config provisioning stays with the env manager); **(d) coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (low-priority tooling); **(e)** the renderer/`-r` forward items (PR #16 Obs 1/2).

## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner is round-agnostic (the current-state pointer is this `STATUS.md`). Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Binary refresh (round 014)**: `go install ./cmd/tellme` rebuilt `$(go env GOPATH)/bin/tellme` from the round-014 head (`e9239a8`); `tellme --version` reports `dev` (no `-ldflags` version stamp).
- **Live Vertex verification (round 013)**: the installed binary was run against the real Vertex API (`tm` → provider `dev`, `TYPE: gemini`, `gemini-3.8-flash`) — single-prompt and tool-loop turns green; the `/home/pos/tmp/dualnets/seed/notebooks/beta-niffler/ait-test/` niffler env (config + `secrets/key.json`) is the live test workspace.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).
- **Secret scanning**: `mcp_github_run_secret_scanning` is **unavailable for this repo** (no GitHub Advanced Security). Closeout secret scans are **diff-level** (pattern grep over `git diff`) — clean through round 014.
- **Round-015 review + `/axb-tasks` (2026-09-14, session 4)**: PR [#38](https://github.com/gosharplite/tellme/pull/38) architecturally reviewed at head `c66f500` — **PLAN + TRUTH APPROVED** with 1 blocker + 4 directives, embedded in `tasks.md`. No `specs/truth/**` change this session (plan-side `tasks.md` + `STATUS.md`/daily log only → no `truth-delta.md` update).
