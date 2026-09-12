# tellme — Status

**Last updated**: 2026-09-12 (**session 17 — tooling/roadmap triage (docs-only): coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13) triaged → low-priority (title de-scoped); F9 flag-parsing opened as [#14](https://github.com/gosharplite/tellme/issues/14) (→ fold into the piping/`-r` slice); CI-platform candidate re-scoped to "run `make verify` in a chosen platform" — **manual for now, platform TBD (GitHub Actions / ADO / Tekton)**. Session 16 — round 004 `004-first-reasoning-turn` DELIVERED**: full pipeline `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement` (25/25 tasks `[X]`); PR [#12](https://github.com/gosharplite/tellme/pull/12) reviewed (**FULL APPROVAL — READY TO MERGE**; F1–F5 fixed at `6f1b6d5` + `cyclop` gate `bf6b983`) → **merged** (`4525b38`) → **propagated `→ dev → main`**; PR-head branch deleted; issue [#10](https://github.com/gosharplite/tellme/issues/10) closed; coverage tooling tracked as [#13](https://github.com/gosharplite/tellme/issues/13)). (Prior — session 15: round 003 `003-provider-registry-completeness` **DELIVERED** — PR [#11](https://github.com/gosharplite/tellme/pull/11) re-reviewed (FULL APPROVAL) → review findings #1–#3 resolved in-round (`05d2e0d`) → merged (`9ab3185`) → propagated `003-provider-registry-completeness → dev → main`; local PR-head branch deleted). *Prior — session 14: round 003 implemented end-to-end (full pipeline `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement`; 18/18 tasks `[X]`), PR [#11](https://github.com/gosharplite/tellme/pull/11) opened. Prior — session 13: 003/004 roadmap and tracking issues ([#9](https://github.com/gosharplite/tellme/issues/9), [#10](https://github.com/gosharplite/tellme/issues/10)) + doc-tree reorganization. Prior — session 12: round 002 delivered & merged ([PR #7](https://github.com/gosharplite/tellme/pull/7)).*
**Review response**: PR [#12](https://github.com/gosharplite/tellme/pull/12) architectural review ([#5642832718](https://github.com/gosharplite/tellme/pull/12#issuecomment-5642832718), verdict *APPROVE with non-blocking follow-ups*; re-check [#5642872784](https://github.com/gosharplite/tellme/pull/12#issuecomment-5642872784), verdict *FULL APPROVAL — READY TO MERGE*; final sign-off [#5643011832](https://github.com/gosharplite/tellme/pull/12#issuecomment-5643011832)) — findings **F1–F5** addressed in-round (`6f1b6d5`); `cyclop` complexity gate added (`bf6b983`); coverage tooling tracked as [#13](https://github.com/gosharplite/tellme/issues/13).
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` (round 004 delivered; the next round starts a fresh `005-*` package off `dev`).
**Daily log**: [`docs/session-summary/2026/09/12/session-summary.md`](docs/session-summary/2026/09/12/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) — historical status (rounds 001 and 002, all closed grill/upstream/clarify records, accumulated decisions log, propagation + environment history), cut at 2026-09-11 (session 13).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | delivered / frozen (round 001) | Round-001 working branch — PR [#6](https://github.com/gosharplite/tellme/pull/6) merged; round 001 is delivered / frozen history |
| `002-followup-cleanups` | delivered / frozen (round 002) | Round-002 base branch — PR [#7](https://github.com/gosharplite/tellme/pull/7) merged (`f2a058f`); propagated `→ dev → main` |
| `003-provider-registry-completeness` | delivered / frozen (round 003) | Round-003 base branch — PR [#11](https://github.com/gosharplite/tellme/pull/11) merged (`9ab3185`); propagated `→ dev → main`; frozen history |
| `004-first-reasoning-turn` | delivered / frozen (round 004) | Round-004 base branch — PR [#12](https://github.com/gosharplite/tellme/pull/12) merged (`4525b38`); propagated `→ dev → main`; frozen history |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`. Delivered round branches
> (`001`, `002`, `003`, `004`) remain frozen history and never receive post-round commits.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Round 003 — `003-provider-registry-completeness` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-12) — PR [#11](https://github.com/gosharplite/tellme/pull/11) merged; propagated to `dev` and `main`.

**Scope**: (1) Expand `PROVIDERS` entry from boot subset to the real, typed provider configuration needed for LLM requests (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`); (2) Deterministic `${VAR}` and `${VAR:-default}` environment variable expansion in `API_KEY`, `URL`, and `HEADERS` string values; (3) Deterministic offline validation and operator-facing failure contract (exit code `3`, frozen class phrase `tellme: the provider configuration is invalid`); (4) Pure-helper unit tests and regression verification.

> **Rounds 001 and 002 (and now 003) remain frozen**: `specs/plans/001-cli-bootstrap-and-config/**`, `specs/plans/002-followup-cleanups/**`, and `specs/plans/003-provider-registry-completeness/**` are immutable history.

### Artifacts

- [x] `specs/plans/003-provider-registry-completeness/spec.md` — `/axb-specify`
- [x] `specs/plans/003-provider-registry-completeness/checklists/requirements.md` — `/axb-specify`
- [x] `specs/plans/003-provider-registry-completeness/truth-delta.md` — `/axb-specify` / `/axb-technical-research` / `/axb-dsl-refine`
- [x] `specs/plans/003-provider-registry-completeness/features/acceptance/*.feature` — `/axb-spec-by-example`
- [x] `specs/plans/003-provider-registry-completeness/research.md` + `specs/truth/techstack.md` — `/axb-technical-research`
- [x] `specs/plans/003-provider-registry-completeness/plan.md` — `/axb-system-analysis`
- [x] `specs/truth/features/cli/**` — `/axb-dsl-refine`
- [x] `specs/plans/003-provider-registry-completeness/tasks.md` — `/axb-tasks`
- [x] Implementation — `/axb-implement` (18/18 tasks `[X]`; godog 27/27 scenarios, `make verify` OK; PR [#11](https://github.com/gosharplite/tellme/pull/11) **merged** `9ab3185`)

### Pipeline position

All phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (18/18 tasks `[X]`)**. Unit tests green, godog **27/27 scenarios · 195/195 steps**, `make verify` OK (zero test-sleep, zero network capability, 0 lint issues, 0 vulnerabilities). PR [#11](https://github.com/gosharplite/tellme/pull/11) reviewed (**FULL APPROVAL**), **merged** (`9ab3185`), and **propagated `003-provider-registry-completeness → dev → main`**. Round 003 is delivered / frozen.

### Decisions locked (round 003)

- **Clarify Q1 -> Option 1 (Core request set)**: Model `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS` (map of string to string), `THINKING_BUDGET` (int), and `THINKING_LEVEL` (string). Defer `USER_ID`, `THINKING_ENABLED`, and top-level `MODELS` pricing tables.
- **Clarify Q2 -> Option 1 (Targeted `${VAR}` expansion with error on unset)**: Expand `${VAR}` and `${VAR:-default}` in `API_KEY`, `URL`, and `HEADERS` values. If an environment variable has no default and is unset/empty, configuration resolution fails deterministically.
- **Clarify Q3 -> Option 1 (Exit code 3 with dedicated class phrase)**: Validation failure exits with code `3` and emits the frozen class phrase `tellme: the provider configuration is invalid`.
- **Top-level key tolerance**: YAML parser maintains tolerance for unknown top-level keys (`MODELS:`, `MCP_SERVERS:`) to ensure compatibility with real-world configs, while strictly validating the resolved provider entry.
- **Context-window / pricing mapping**: Deferred to Slice 004 as an assumption.

### Review response — PR #11 (round 003, in-round)

- Review comment: [#5641984093](https://github.com/gosharplite/tellme/pull/11#issuecomment-5641984093) — verdict **APPROVE WITH NON-BLOCKING FOLLOW-UPS**.
- **F1 (OS coupling in `expand.go`)** — introduced an injectable `EnvLookupFunc` port (`ExpandStringWithLookup`; `ExpandString` delegates to `os.LookupEnv`); the expansion unit table is now in-memory and `t.Parallel()`-safe. Truth: `techstack.md` *Variable expansion* row (MODIFY).
- **F2 (validate ran before expand)** — `resolve()` now **expands then validates**, so a mandatory field whose placeholder resolves to empty (e.g. `URL: "${UNSET:-}"`) is rejected; added executable acceptance Example *"A mandatory field resolves to empty after expansion"* + `TestResolveRejectsEmptyAfterExpansion`. Truth: `techstack.md` resolution row (MODIFY) + `features/cli/configuration` feature (MODIFY).
- **F3 (resolved provider discarded)** — `resolution` now carries the resolved (expanded) `config.Provider` for Slice 004; covered by `TestResolveCarriesExpandedProvider`. No truth change (internal struct).
- Verification: unit tests green; godog **27/27 scenarios · 195/195 steps**; topology audit **PASSED** (172 steps); `make verify` OK; staticcheck clean.

### Delivery & propagation (session 15)

- Re-check comment: [#5642050755](https://github.com/gosharplite/tellme/pull/11#issuecomment-5642050755) — verdict **FULL APPROVAL — READY TO MERGE**.
- PR [#11](https://github.com/gosharplite/tellme/pull/11) merged into `003-provider-registry-completeness` (`9ab3185`); remote PR-head branch deleted.
- Propagated `003-provider-registry-completeness → dev` (`6db276e`, no-ff) `→ main` (no-ff).
- Round 003 `specs/plans/003-provider-registry-completeness/**` is now **delivered / frozen** history.
- Next: **Slice 004 — First reasoning turn** ([#10](https://github.com/gosharplite/tellme/issues/10)) starts a fresh `004-*` package off `dev`; it will consume the provider carried on `resolution` (review F3).

## Round 004 — `004-first-reasoning-turn` (delivered / frozen)

**Status**: ✅ **DELIVERED / FROZEN** (2026-09-12) — PR [#12](https://github.com/gosharplite/tellme/pull/12) merged (`4525b38`); propagated to `dev` and `main`; issue [#10](https://github.com/gosharplite/tellme/issues/10) closed.

> **Rounds 001–003 (and now 004) remain frozen**: `specs/plans/001-cli-bootstrap-and-config/**`, `specs/plans/002-followup-cleanups/**`, `specs/plans/003-provider-registry-completeness/**`, and `specs/plans/004-first-reasoning-turn/**` are immutable history.

**Scope**: (1) `tellme "<prompt>"` → one provider request → printed response; (2) a provider domain port + one concrete adapter (OpenAI-compatible family first); (3) request assembly from the round-003 resolved provider; (4) response normalization to a minimal answer; (5) a deterministic provider/transport failure contract (frozen class phrase `the provider request failed` + exit code `6`); (6) a network-path test strategy (local fake provider) + amended no-network capability guard.

### Artifacts

- [x] `specs/plans/004-first-reasoning-turn/spec.md` — `/axb-specify`
- [x] `specs/plans/004-first-reasoning-turn/checklists/requirements.md` — `/axb-specify`
- [x] `specs/plans/004-first-reasoning-turn/truth-delta.md` — `/axb-specify` / `/axb-technical-research` / `/axb-dsl-refine`
- [x] `features/acceptance/*.feature` — `/axb-spec-by-example` (2 journey features: `answering-a-single-prompt`, `reporting-a-failed-provider-request`)
- [x] `research.md` + `specs/truth/techstack.md` — `/axb-technical-research` (7 decisions; techstack ADD Reasoning & Provider Transport + guard amendment)
- [x] `plan.md` — `/axb-system-analysis` (2 interfaces, 1 wave; `/axb-api-plan` = NOOP, `/axb-data-plan` = NOOP; CLI end + provider gateway → `/axb-dsl-refine` handoff)
- [x] `specs/truth/features/cli/**` — `/axb-dsl-refine` (new `chat` module: 2 features + dsl; root DSL +1 class phrase `the provider request failed`; 2 rows promoted to root; topology audit PASSED — 219 steps)
- [x] `tasks.md` — `/axb-tasks` (25 tasks; Setup omitted — stdlib only; Phase 3 = 3 ALIGN + 1 UNIT + 10 RED + review; orphan sweep 0)
- [x] Implementation — `/axb-implement` (25/25 tasks `[X]`; godog 34/34 scenarios · 242/242 steps; `make verify` OK)

### Pipeline position

All phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (25/25 tasks `[X]`)**. Unit tests green, godog **34/34 scenarios · 242/242 steps**, `make verify` OK (zero test-sleep, offline-path guard, 0 lint issues, 0 vulnerabilities). PR [#12](https://github.com/gosharplite/tellme/pull/12) reviewed (**FULL APPROVAL**), **merged** (`4525b38`), and **propagated `004-first-reasoning-turn → dev → main`**. Round 004 is delivered / frozen.

### Decisions locked (round 004 — Clarify Round 1)

- **Q1 -> Option 1 (single in-memory, non-streaming turn; no persistence)**: one `tellme "<prompt>"` = one provider request → printed response; no `history.jsonl`, no session loop, no streaming.
- **Q2 -> Option 1 (OpenAI-compatible family first)**: the first adapter targets `openai`/`deepseek`/`kimi`; Gemini/Vertex and Anthropic are deferred.
- **Q3 -> Option 1 (new failure class)**: provider/transport failure → frozen class phrase `the provider request failed` + new distinct exit code `6` (table extends `0/2/3/4/5` → `0/2/3/4/5/6`).
- **Deferred to `/axb-technical-research`**: transport = stdlib `net/http` (no SDK); the no-network capability guard is re-scoped to boot/`--version`/`-d` (round-001 research Decision 5 amendment). Recorded as spec assumptions.

### Review response — PR #12 (round 004)

Review comment [#5642832718](https://github.com/gosharplite/tellme/pull/12#issuecomment-5642832718) — verdict **APPROVE WITH NON-BLOCKING FOLLOW-UPS**; re-check [#5642872784](https://github.com/gosharplite/tellme/pull/12#issuecomment-5642872784) — **FULL APPROVAL — READY TO MERGE**; final sign-off [#5643011832](https://github.com/gosharplite/tellme/pull/12#issuecomment-5643011832). All five findings were addressed in-round on the PR branch:

- **F1** (direct infra coupling + unchecked `Provider.Type`) → added the **gateway factory/dispatch seam** `internal/infrastructure/llm/factory.go` (`NewGateway` switches on the family; unsupported → `*llm.ProviderError`) and made `runTurn` injectable; covered by `factory_test.go` + `turn_test.go`.
- **F2** (unbounded hang) → the adapter's default HTTP client now carries a 300s timeout; the turn context is cancelled on `SIGINT`/`SIGTERM` (`signal.NotifyContext`).
- **F3** (discarded non-2xx body) → the provider's structured error message (`{"error":{"message":…}}`) is surfaced in the actionable detail (FR-008).
- **F4** (unbounded read) → response read bounded by `io.LimitReader` (32 MiB).
- **F5** (hardcoded model in a test step) → `step_t017` now asserts a generic JSON `"model"` field instead of a literal.

Verification after the fix set: `gofmt`/`vet`/`staticcheck` clean · `make verify` OK · godog **34/34 · 242/242** · new unit tests for the factory, `runTurn`, and `extractErrorMessage`.

### Delivery & propagation (session 16)

- Re-check comment: [#5642872784](https://github.com/gosharplite/tellme/pull/12#issuecomment-5642872784) — verdict **FULL APPROVAL — READY TO MERGE**; final sign-off: [#5643011832](https://github.com/gosharplite/tellme/pull/12#issuecomment-5643011832).
- PR [#12](https://github.com/gosharplite/tellme/pull/12) merged into `004-first-reasoning-turn` (`4525b38`); remote + local PR-head branch `004-implement-first-reasoning-turn` deleted.
- Propagated `004-first-reasoning-turn → dev` (`ee54e46`, no-ff) `→ main` (no-ff).
- Issue [#10](https://github.com/gosharplite/tellme/issues/10) closed as **completed**; coverage-tooling follow-up tracked as [#13](https://github.com/gosharplite/tellme/issues/13).
- Round 004 `specs/plans/004-first-reasoning-turn/**` is now **delivered / frozen** history.

## Roadmap — next slices

| Slice | Issue | Scope | Status |
| --- | --- | --- | --- |
| **003 — Provider-registry completeness** | [#9](https://github.com/gosharplite/tellme/issues/9) | Grow boot-subset `PROVIDERS` entry to real provider fields (`API_KEY` with `${VAR}` expansion, `HEADERS`, `THINKING_BUDGET`/`THINKING_LEVEL`) + deterministic offline validation. | ✅ **Delivered** (PR [#11](https://github.com/gosharplite/tellme/pull/11) merged; propagated to `dev`/`main`) |
| **004 — First reasoning turn** | [#10](https://github.com/gosharplite/tellme/issues/10) | `tellme "<prompt>"` → one provider request → printed response; provider domain port + one adapter; deterministic failure class; network-path test strategy. **Depends on 003.** | ✅ **Delivered** (PR [#12](https://github.com/gosharplite/tellme/pull/12) merged `4525b38`; propagated to `dev`/`main`; [#10](https://github.com/gosharplite/tellme/issues/10) closed) |

## Open items (non-blocking)

- **Future-package candidates** (list refreshed 2026-09-12):
  - **(a) Run `make verify` in a pipeline platform** — [#15](https://github.com/gosharplite/tellme/issues/15); the CI/CD platform (GitHub Actions / ADO / Tekton) is **not a repo-level choice**, so the gate stays **manual for now** (the `SESSION-CLOSEOUT.md` `make verify`). The Makefile stays the single source of gate truth; whichever platform is chosen later wraps that same command. *(Re-scoped 2026-09-12 from "CI workflow".)*
  - **(b) F9 extension — `internal/cli` flag-parsing unit tests** — tracked as [#14](https://github.com/gosharplite/tellme/issues/14); recommended to **fold into the future piping / raw-output (`-r`) slice** rather than stand alone.
  - ~~**(c) PM-4 `tellme init`**~~ — **DROPPED (2026-09-12)**: config provisioning stays with the environment manager (Niffler / `tellme.sh`); `tellme` remains a **load/validate consumer** working inside that shell — **no second config-writer**. Candidate withdrawn.
  - **(d) Coverage tooling** — [#13](https://github.com/gosharplite/tellme/issues/13) (`make test-coverage` report + `go build -cover` E2E-integration spike; PR [#12](https://github.com/gosharplite/tellme/pull/12) review follow-up). Triaged 2026-09-12 as **low-priority tooling**, not a committed round.
- Pre-existing non-blocking items from rounds 001/002 remain documented in archive.


## Environment notes

- **Dev tooling — `tellme.sh` (external; not a repo/truth artifact)**: the Niffler-style manager that drives the `tellme` binary lives at both `~/tmp/dualnets/seed/notebooks/beta-niffler/tellme.sh` and `…/mbp-johndoe-niffler/tellme.sh`. Its usage banner was made **round-agnostic** this session — the hardcoded `round-001` capability text (and the `bare boot` capability hint) were removed so it no longer needs revising each slice; the current-state pointer is this `STATUS.md`. Invoke via `source tellme.sh` (aliases `b`/`a`/`c`/`g`/`p`/`r` + optional prompt arg) or the `tm` alias in `~/.bashrc`.
- **Host / toolchain**: Go 1.26; `golangci-lint` / `staticcheck` / `govulncheck` present in `$GOPATH/bin`; the `tellme` binary is installed at `$(go env GOPATH)/bin/tellme` via `go install ./cmd/tellme`. Sandbox: the privileged netns (`unshare -n`) is unavailable on the host, so the offline-path guard uses the unprivileged canary + hostile-env differential.
- **Persistent path authorizations**: read+write for `…/beta-niffler/`, `…/mbp-johndoe-niffler/tellme.sh`, and `~/.bashrc` (the `tm` alias).
