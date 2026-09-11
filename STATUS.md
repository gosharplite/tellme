# tellme — Status

**Last updated**: 2026-09-11 (**session 14 — round 003 `003-provider-registry-completeness` launched**; checked out working branch `003-provider-registry-completeness` off `dev`; `/axb-specify` executed; Clarify Round 1 resolved with user [Q1 core request set, Q2 targeted `${VAR}` expansion with error on unset, Q3 exit 3 with `tellme: the provider configuration is invalid`]; `spec.md`, `checklists/requirements.md`, and `truth-delta.md` scaffolded). *Prior — session 13: 003/004 roadmap and tracking issues ([#9](https://github.com/gosharplite/tellme/issues/9), [#10](https://github.com/gosharplite/tellme/issues/10)) + doc-tree reorganization. Prior — session 12: round 002 delivered & merged ([PR #7](https://github.com/gosharplite/tellme/pull/7)).*
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `003-provider-registry-completeness` (working branch off `dev`).
**Daily log**: [`docs/session-summary/2026/09/11/session-summary.md`](docs/session-summary/2026/09/11/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) — historical status (rounds 001 and 002, all closed grill/upstream/clarify records, accumulated decisions log, propagation + environment history), cut at 2026-09-11 (session 13).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from delivered round branches | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | delivered / frozen (round 001) | Round-001 working branch — PR [#6](https://github.com/gosharplite/tellme/pull/6) merged; round 001 is delivered / frozen history |
| `002-followup-cleanups` | delivered / frozen (round 002) | Round-002 base branch — PR [#7](https://github.com/gosharplite/tellme/pull/7) merged (`f2a058f`); propagated `→ dev → main` |
| `003-provider-registry-completeness` | active (working branch) | Round-003 working branch — scoped to provider-registry completeness (Issue [#9](https://github.com/gosharplite/tellme/issues/9)) |

> **Branch convention**: each round works on its own `NNN-*` branch off `dev`. Delivered round branches
> (`001`, `002`) remain frozen history and never receive post-round commits.
> Read live heads with `git rev-parse --short main dev HEAD`.

## Current round — `003-provider-registry-completeness`

**Scope**: (1) Expand `PROVIDERS` entry from boot subset to the real, typed provider configuration needed for LLM requests (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`); (2) Deterministic `${VAR}` and `${VAR:-default}` environment variable expansion in `API_KEY`, `URL`, and `HEADERS` string values; (3) Deterministic offline validation and operator-facing failure contract (exit code `3`, frozen class phrase `tellme: the provider configuration is invalid`); (4) Pure-helper unit tests and regression verification.

> **Rounds 001 and 002 remain frozen**: `specs/plans/001-cli-bootstrap-and-config/**` and `specs/plans/002-followup-cleanups/**`
> are immutable history.

### Artifacts

- [x] `specs/plans/003-provider-registry-completeness/spec.md` — `/axb-specify`
- [x] `specs/plans/003-provider-registry-completeness/checklists/requirements.md` — `/axb-specify`
- [x] `specs/plans/003-provider-registry-completeness/truth-delta.md` (skeleton) — `/axb-specify`
- [x] `specs/plans/003-provider-registry-completeness/features/acceptance/*.feature` — `/axb-spec-by-example`
- [x] `specs/plans/003-provider-registry-completeness/research.md` + `specs/truth/techstack.md` — `/axb-technical-research`
- [x] `specs/plans/003-provider-registry-completeness/plan.md` — `/axb-system-analysis`
- [x] `specs/truth/features/cli/**` — `/axb-dsl-refine`
- [x] `specs/plans/003-provider-registry-completeness/tasks.md` — `/axb-tasks`
- [ ] Implementation — `/axb-implement`

### Pipeline position

Phase 1 (`/axb-specify`) **complete** (`spec.md`, `checklists/requirements.md`, `truth-delta.md` authored). Clarify Round 1 resolved with user. Ready for parallel / sequential Phase 2: `/axb-spec-by-example` (PM acceptance Gherkin) and `/axb-technical-research` (RD technical decisions).

### Decisions locked (round 003)

- **Clarify Q1 -> Option 1 (Core request set)**: Model `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS` (map of string to string), `THINKING_BUDGET` (int), and `THINKING_LEVEL` (string). Defer `USER_ID`, `THINKING_ENABLED`, and top-level `MODELS` pricing tables.
- **Clarify Q2 -> Option 1 (Targeted `${VAR}` expansion with error on unset)**: Expand `${VAR}` and `${VAR:-default}` in `API_KEY`, `URL`, and `HEADERS` values. If an environment variable has no default and is unset/empty, configuration resolution fails deterministically.
- **Clarify Q3 -> Option 1 (Exit code 3 with dedicated class phrase)**: Validation failure exits with code `3` and emits the frozen class phrase `tellme: the provider configuration is invalid`.
- **Top-level key tolerance**: YAML parser maintains tolerance for unknown top-level keys (`MODELS:`, `MCP_SERVERS:`) to ensure compatibility with real-world configs, while strictly validating the resolved provider entry.
- **Context-window / pricing mapping**: Deferred to Slice 004 as an assumption.

## Roadmap — next slices

| Slice | Issue | Scope | Primary truth owners |
| --- | --- | --- | --- |
| **003 — Provider-registry completeness** | [#9](https://github.com/gosharplite/tellme/issues/9) | Grow boot-subset `PROVIDERS` entry to real provider fields (`API_KEY` with `${VAR}` expansion, `HEADERS`, `THINKING_BUDGET`/`THINKING_LEVEL`) + deterministic offline validation. **(In Progress)** | `/axb-dsl-refine` (MODIFY `features/cli` configuration), `/axb-technical-research` (techstack) |
| **004 — First reasoning turn** | [#10](https://github.com/gosharplite/tellme/issues/10) | `tellme "<prompt>"` → one provider request → printed response; provider domain port + one adapter; deterministic failure class; network-path test strategy. **Depends on 003.** | `/axb-dsl-refine` (new chat/turn module), `/axb-technical-research` (transport amendment) |

## Open items (non-blocking)

- **Future-package candidates**: (a) CI workflow for `make verify`; (b) F9 extension for flag parsing unit tests; (c) PM-4 `tellme init`.
- Pre-existing non-blocking items from rounds 001/002 remain documented in archive.
