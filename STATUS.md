# tellme — Status

**Last updated**: 2026-09-10
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)
**Daily log**: [`docs/2026/09/10/session-summary.md`](docs/2026/09/10/session-summary.md)

## Branch model

| Branch | Role |
| --- | --- |
| `main` | Stable / released line |
| `dev` | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | This session's working branch |

## Current round — `001-cli-bootstrap-and-config`

**Scope (narrow foundation, locked)**: CLI boot; YAML configuration load + validation
(`config-valid-provider`); runtime home (`TELL_ME_HOME`) + per-mode session workspace
(`output/<mode>/`); build version (`--version`); offline setup diagnostic (`-d`, `-d --json`).

> **Scope note (updated by clarify Q2)**: the original "no `data/**` truth this round" lock was
> **released** — this round now owes a **minimal** data truth (config input contract + workspace
> lifecycle). Configuration is still **not** modeled as owned system state beyond that minimal surface.

### Artifacts

- [x] `specs/plans/001-cli-bootstrap-and-config/spec.md` *(pending a PM edit — see PM follow-ups)*
- [x] `specs/plans/001-cli-bootstrap-and-config/checklists/requirements.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/truth-delta.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/features/acceptance/*.feature` (4 journey features) — `/axb-spec-by-example` *(pending a PM edit — see PM follow-ups)*
- [x] `specs/plans/001-cli-bootstrap-and-config/research.md` — `/axb-technical-research` (7 decisions; **post-grill + post-clarify**)
- [x] `specs/truth/techstack.md` — `/axb-technical-research` (**post-grill**)
- [ ] `specs/plans/001-cli-bootstrap-and-config/plan.md` — `/axb-system-analysis` ← **next**
- [ ] `specs/truth/data/**` — `/axb-data-plan` (delegated by `/axb-system-analysis`; minimal)
- [ ] `specs/truth/features/**` + `dsl.md` — `/axb-dsl-refine`
- [ ] `specs/plans/001-cli-bootstrap-and-config/tasks.md` — `/axb-tasks`
- [ ] Implementation — `/axb-implement`

### Pipeline position

`/axb-specify` done → `/axb-spec-by-example` done → `/axb-technical-research` done + **grill round #1** +
**clarify round** → **next: `/axb-system-analysis`** (delegates `/axb-data-plan` for the minimal data
truth) → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement`. *(Both former PM-boundary blockers are
resolved.)*

## Grill round #1 — round-001 techstack & research (closed)

- **Issue**: [#1](https://github.com/gosharplite/tellme/issues/1) · **Subject**: `architect` · **Griller**: `griller` · **Orchestrator**: `butler` · **10/10 questions**
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/0b6775f580c4ec66ada984e1bbe023d6>
- **Verdict**: **proceed with changes** — core stack choices held; representations were corrected.
- **Committed corrections landed** into `research.md` / `specs/truth/techstack.md` / `truth-delta.md`
  (flagged **⟦grill R1⟧**): SC-004 proven by the **host harness** (sandbox + build-graph guard); `godog`
  labelled *adopted, first instantiated in `/axb-dsl-refine`* with a pinned `tests/e2e/` harness contract;
  **6-step resolver** pinned; **dropped `internal/version`**; `VERSION=0.0.0-harness` sentinel;
  **`staticcheck` adopted**; three-state truth discipline.

## Clarify round — PM-boundary rulings (closed)

Resolved the two items grill round #1 routed to `/axb-clarify`, **before** `/axb-system-analysis`
(grounded against `tell-me-go`'s current behavior — `-d` renders the report then verdicts in the exit
code; `RunDiagnostics` exits non-zero when unhealthy; the CLI persists config + `output/<mode>/` state
and models both):

- **Q1 → Option 3** — `-d` diagnostic semantics. `-d` **always produces the report** (resolved *or*
  unresolved) and exits a **dedicated, distinct non-zero "diagnostic: unresolved"** code on failure; the
  `FR-014` four codes (success/usage/config/environment) bind the **boot** path only. (Option 1 — exit 0
  on unresolved — rejected as contrary to the benchmark; Option 2 — fail-fast with no report — rejected as
  raw parity with the benchmark's own flaw.) Recorded in `research.md` Decision 2.
- **Q2 → Option 1** — `/axb-data-plan`. The "no `data/**` truth this round" scope lock is **released**;
  the owner authors a **minimal** data truth (config input contract `configs/<mode>.yaml` + workspace
  lifecycle `output/<mode>/`). Recorded in `research.md` + `truth-delta.md`. *(This supersedes the
  `spec.md` assumption — see PM follow-ups.)*

## Decisions locked this session

- Round 1 = **narrow foundation**; **not everything in tell-me-go will appear in tellme**.
- **`/axb-constitution` skipped** — the default constitution is used.
- **Niffler shell-env alignment folded into `spec.md`** (`TELL_ME_*` env-over-file precedence;
  `FR-003`/`FR-007`/`FR-015`); default config path `$TELL_ME_HOME/configs/<mode>.yaml`.
- **Round-001 acceptance Gherkin written** (4 journey features; plan-side only).
- Artifacts written in **English**; butler runs the `axb-*` skills directly with a per-phase review gate
  (the grill round served as the technical-research review gate; the clarify round resolved its routings).
- **Self-starting bootstrap** (`SESSION-BOOTSTRAP.md` Step 7 + `STATUS.md`) on `main`/`dev`/working branch.
- **Working style**: session work on a local branch; updates flow up via the two-step merge–merge
  (`working branch → dev → main`).
- **Data scope**: the round's "no `data/**` truth" lock is **released** (clarify Q2) → a minimal data
  truth is owed.

## PM follow-ups (spec/acceptance are PM-owned — not written by the RD/butler flow)

- **`spec.md`**: update the Assumptions/config-scope line — the "no `data/**` truth this round" lock is
  released (clarify Q2). *(Needed to remove the spec-vs-ruling inconsistency.)*
- **`features/acceptance/**`**: add an edge case + an acceptance Example for **`-d` on a broken/unresolved
  setup** (clarify Q1), and the missing **no-`-c` + `MODE≠butler`** default-path Example.
- **Proposal (optional)**: smallest vertical addition for user value — `tellme init` (or `tellme config
  show`) — see `research.md` residual risks.

## Open items (non-blocking)

- Exact `-d --json` output schema.
- Exit-code numeric values (only distinctness is required) — note the new dedicated "diagnostic:
  unresolved" code (clarify Q1).
- Error-message wording (`NFR-004`).
- **Unchecked-error coverage** (round-001 residual): closed next slice by `golangci-lint` with `errcheck`
  enabled (Decision 7).

## Environment notes

- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for both read and write.
- **2026-09-10**: session restarted following the `deepseek-flash` (v4.1) provider upgrade. The restart is
  a session/environment event only — nothing in round 001 depends on the provider version.
