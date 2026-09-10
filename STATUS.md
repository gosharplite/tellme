# tellme — Status

**Last updated**: 2026-09-10 (end of day — session-continuity tooling added)
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)
**Daily log**: [`docs/2026/09/10/session-summary.md`](docs/2026/09/10/session-summary.md)

## Branch model

| Branch | Head | Role |
| --- | --- | --- |
| `main` | `6b0fdb7` | Stable / released line |
| `dev` | `505e474` | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | `cca7c0e` (+ closeout) | This session's working branch |

> **Propagation pending:** `dev`/`main` predate the round-1 RD artifacts. The two-step merge
> `working → dev → main` has not been run yet.

## Current round — `001-cli-bootstrap-and-config`

**Scope (narrow foundation, locked)**: CLI boot; YAML configuration load + validation
(`config-valid-provider`); runtime home (`TELL_ME_HOME`) + per-mode session workspace
(`output/<mode>/`); build version (`--version`); offline setup diagnostic (`-d`, `-d --json`).

> **Scope note (updated by clarify Q2)**: the original "no `data/**` truth this round" lock was
> **released** — this round now owes a **minimal** data truth (config input contract + workspace
> lifecycle).

### Artifacts

- [x] `specs/plans/001-cli-bootstrap-and-config/spec.md` *(pending a PM edit — see PM follow-ups)*
- [x] `specs/plans/001-cli-bootstrap-and-config/checklists/requirements.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/truth-delta.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/features/acceptance/*.feature` (4 journey features) — `/axb-spec-by-example`
- [x] `specs/plans/001-cli-bootstrap-and-config/research.md` — `/axb-technical-research` (7 decisions; **post-grill + post-clarify**)
- [x] `specs/truth/techstack.md` — `/axb-technical-research` (**post-grill**)
- [x] `specs/plans/001-cli-bootstrap-and-config/plan.md` — `/axb-system-analysis` (2 interfaces, 2 waves)
- [ ] `specs/truth/data/**` — `/axb-data-plan` (Wave 1; minimal) ← **next**
- [ ] `specs/truth/features/**` + `dsl.md` — `/axb-dsl-refine` (Wave 2)
- [ ] `specs/plans/001-cli-bootstrap-and-config/tasks.md` — `/axb-tasks`
- [ ] Implementation — `/axb-implement`

### System analysis (`plan.md`)

- **Interfaces (2)**: `CLI end (operator terminal interface)`; `Configuration & workspace persistence interface`.
- **Waves (2)**: **Wave 1** = persistence → `/axb-data-plan`; **Wave 2** = CLI contract → `/axb-dsl-refine`.
- **`/axb-api-plan` = NOOP** (single CLI end, no OpenAPI); **`/axb-ui-plan` skipped** (CLI).

### Pipeline position

`/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` (+ **grill round #1**, **clarify**) →
`/axb-system-analysis` **(done — `plan.md`)** → **next: `/axb-data-plan` (Wave 1) → `/axb-dsl-refine`
(Wave 2)** → `/axb-tasks` → `/axb-implement`.

**Pending decision:** run a **grill round on `plan.md`** (recommended, cap ~6) or proceed directly to Wave 1.

## Grill round #1 — round-001 techstack & research (CLOSED)

- **Issue**: [#1](https://github.com/gosharplite/tellme/issues/1) — **closed** (completed) · subject `architect` · griller `griller` · orchestrator `butler` · **10/10 questions**
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/0b6775f580c4ec66ada984e1bbe023d6>
- **Verdict**: **proceed with changes** — core stack choices held; representations were corrected.
- **Committed corrections** (⟦grill R1⟧) landed in `research.md` / `specs/truth/techstack.md` /
  `truth-delta.md`: SC-004 proven by the **host harness** (sandbox + build-graph guard); `godog` labelled
  *adopted, first instantiated in `/axb-dsl-refine`*; **6-step resolver**; **dropped `internal/version`**;
  `VERSION=0.0.0-harness` sentinel; **`staticcheck` adopted**; three-state truth discipline.

## Clarify round — PM-boundary rulings (CLOSED)

Resolved the two items grill round #1 routed to `/axb-clarify`, **before** `/axb-system-analysis`
(grounded against `tell-me-go`'s current behavior):

- **Q1 → Option 3** — `-d` diagnostic semantics: always produce the report (resolved *or* unresolved) and
  exit a **dedicated, distinct non-zero "diagnostic: unresolved"** code; the `FR-014` four codes
  (success/usage/config/environment) bind the **boot** path only.
- **Q2 → Option 1** — the "no `data/**` truth this round" scope lock is **released**; `/axb-data-plan`
  authors a **minimal** data truth (config input contract + workspace lifecycle).

## Feasibility verification pass (done)

- ✓ `go1.26.6`; ✓ `staticcheck`/`golangci-lint`/`govulncheck` present; ✓ single `main.version` target;
  ✓ 6-step resolver non-circular.
- ✗ **`unshare -n` / `-rn` fail `Operation not permitted`** on this dev host → `research.md` Decision 5
  now names the portable **unprivileged** no-egress fallback (hostile DNS/proxy) alongside the privileged
  netns; the build-graph guard remains the decisive witness. (`6951d17`)

## Decisions locked this session

- Round 1 = **narrow foundation**; **not everything in tell-me-go will appear in tellme**.
- **`/axb-constitution` skipped** — the default constitution is used.
- **Niffler shell-env alignment** folded into `spec.md` (`TELL_ME_*` env-over-file precedence;
  `FR-003`/`FR-007`/`FR-015`); default config path `$TELL_ME_HOME/configs/<mode>.yaml`.
- **Round-001 acceptance Gherkin written** (4 journey features; plan-side only).
- Artifacts written in **English**; butler runs the `axb-*` skills directly with a per-phase review gate
  (the grill round was the technical-research gate; the clarify round resolved its routings).
- **Self-starting bootstrap** (`SESSION-BOOTSTRAP.md` Step 7 + `STATUS.md`) on `main`/`dev`/working branch.
- **Working style**: session work on a local branch; updates flow up via the two-step merge–merge.
- **Data scope released** (clarify Q2); **`-d` = reporting path** (clarify Q1).
- **Two-wave analysis** (data → CLI contract): the persistence model underpins the executable contract.
- **Session-continuity tooling**: `SESSION-BOOTSTRAP.md` gains **Step 8** (read the last 5 days of
  `docs/…/session-summary.md`); new **`SESSION-CLOSEOUT.md`** defines the end-of-day procedure
  (review tree → quality gates → `STATUS.md` → daily summary → reconcile → commit → propagate).

## PM follow-ups (spec/acceptance are PM-owned — not written by the RD/butler flow)

- **`spec.md`**: update the Assumptions line — the "no `data/**` truth this round" lock is released (Q2).
- **`features/acceptance/**`**: add an edge case + Example for **`-d` on a broken/unresolved setup** (Q1),
  and the missing **no-`-c` + `MODE≠butler`** default-path Example.
- **Proposal (optional)**: smallest vertical addition for user value — `tellme init` (or `tellme config show`).

## Open items (non-blocking)

- Exact `-d --json` output schema.
- Exit-code numeric values (incl. the new dedicated "diagnostic: unresolved" code).
- Error-message wording (`NFR-004`).
- **Unchecked-error coverage** (round-001 residual): closed next slice by `golangci-lint` + `errcheck`.
- **Session-lifecycle docs**: optionally cross-link `SESSION-BOOTSTRAP.md` ↔ `SESSION-CLOSEOUT.md`.

## Environment notes

- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for read **and** write.
- **2026-09-10**: session restarted following the `deepseek-flash` (v4.1) provider upgrade — a
  session/environment event only; nothing in round 001 depends on the provider version.
- **No-network sandbox limitation (verified 2026-09-10)**: `unshare -n` / `unshare -rn` fail
  `Operation not permitted` on this dev host — the privileged netns sandbox is a CI/privileged-Linux
  mechanism; local SC-004 uses the unprivileged hostile-DNS/proxy fallback + the build-graph guard.
