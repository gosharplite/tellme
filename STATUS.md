# tellme — Status

**Last updated**: 2026-09-10 (session 4 — grill round #2 on `plan.md` + plan revision)
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)
**Daily log**: [`docs/2026/09/10/session-summary.md`](docs/2026/09/10/session-summary.md)

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from the working branch | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | session tip (moves per commit) | This session's working branch |

> **Heads are intentionally not pinned here** — the working branch is the moving tip and every commit
> on it is followed by the two-step merge `working → dev → main`. Read live heads with
> `git rev-parse --short main dev HEAD` rather than trusting a snapshot.
> **Last propagation (2026-09-10):** `working → dev → main` — `dev`/`main` carry the round-001
> artifacts and the session-continuity tooling. Re-run the merge after any later working-branch commit.

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
- [x] `specs/plans/001-cli-bootstrap-and-config/plan.md` — `/axb-system-analysis` (**revised by grill #2**: 2 interfaces, **1 wave**, gating blockers)
- [ ] `specs/truth/data/**` — `/axb-data-plan` (Wave 1; minimal) ← **next** (not gated)
- [ ] `specs/truth/features/**` + `dsl.md` — `/axb-dsl-refine` (**CLI slice gated** — see gating blockers)
- [ ] `specs/plans/001-cli-bootstrap-and-config/tasks.md` — `/axb-tasks`
- [ ] Implementation — `/axb-implement`

### System analysis (`plan.md`) — **revised by grill round #2**

- **Interfaces (2)**: `CLI end (operator terminal interface)` — planner **UNASSIGNED (recorded gap)**;
  `Configuration & workspace persistence interface` — `/axb-data-plan`.
- **Wave (1, single)**: both interfaces are information-independent → one parallel wave. *(The earlier
  2-wave split was retracted — no Rule-2 information dependency.)*
- **`/axb-api-plan` = NOOP** (single CLI end, no OpenAPI); **`/axb-ui-plan` skipped** (CLI).
- **Gating blockers** (gate the `/axb-dsl-refine` CLI slice — see grill round #2 below).

### Pipeline position

`/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` (+ **grill round #1**, **clarify**) →
`/axb-system-analysis` **(done — `plan.md`, revised by grill round #2)** → **next: `/axb-data-plan`
(single wave; not gated) → `/axb-dsl-refine` (Wave 1 CLI contract — **CLI slice gated** on the
blockers)** → `/axb-tasks` → `/axb-implement`.

**Pending decision:** *(resolved)* the **grill round on `plan.md`** ran as **grill round #2** (6/6,
verdict *proceed with changes*); the edit set is applied. **Remaining gate:** the cross-repo
`aixbdd-tmg` blocker + the two PM acceptance gaps must land before the `/axb-dsl-refine` CLI slice
(the `-d`-unresolved / default-path behaviour) can proceed; the `/axb-data-plan` half is **not** gated.

## Grill round #1 — round-001 techstack & research (CLOSED)

- **Issue**: [#1](https://github.com/gosharplite/tellme/issues/1) — **closed** (completed) · subject `architect` · griller `griller` · orchestrator `butler` · **10/10 questions**
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/0b6775f580c4ec66ada984e1bbe023d6>
- **Verdict**: **proceed with changes** — core stack choices held; representations were corrected.
- **Committed corrections** (⟦grill R1⟧) landed in `research.md` / `specs/truth/techstack.md` /
  `truth-delta.md`: SC-004 proven by the **host harness** (sandbox + build-graph guard); `godog` labelled
  *adopted, first instantiated in `/axb-dsl-refine`*; **6-step resolver**; **dropped `internal/version`**;
  `VERSION=0.0.0-harness` sentinel; **`staticcheck` adopted**; three-state truth discipline.

## Grill round #2 — round-001 `plan.md` system analysis (CLOSED)

- **Issue**: [#2](https://github.com/gosharplite/tellme/issues/2) · subject `architect` · griller `griller` · orchestrator `butler` · **6/6 questions** (cap 6)
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/27c916a78fd3eb988f2ac22f7b4da954>
- **Findings comment**: <https://github.com/gosharplite/tellme/issues/2#issuecomment-5617519522>
- **Verdict**: **proceed with changes** — the plan's structure / NOOP-skip rulings held; its wave
  structure and interface framing were corrected.
- **Applied to `plan.md`** (edit set): (a) **2-wave split → single wave** (no information dependency —
  `/axb-dsl-refine` reads no `specs/truth/data/**`); (b) **count stays 2** with the **CLI end's planner
  UNASSIGNED recorded as a gap** (not deleted); (c) **`-d`-unresolved / default-path marked
  not-yet-delegable**; (d) delegation ordering re-attributed to **pipeline phasing**, not a Rule-2
  dependency; (e) a **Gating blockers** subsection added.
- **Subject retractions** (all conceded under verification): **Q1** `Interface.kind` gap is a present
  blocker, not a "forward risk"; **Q2** the Wave-1→Wave-2 dependency; **Q3**
  `wave-covers-interfaces` ("holds" claim); **Q6** the count-1 fix (definition inverted the skill's
  inventory-then-delegate order). **Q5**: conceded an over-commit.

### Gating blockers (gate the `/axb-dsl-refine` CLI slice)

1. **Cross-repo blocker (`aixbdd-tmg` truth-model owner)** — the typed model has **no seat for a CLI
   end**: no valid `InterfaceKind` value (`{backend, frontend}`, both web-bound) and no api/data/ui
   planner for a terminal endpoint. *Proposed resolution:* extend `InterfaceKind` (`cli` →
   `features/cli/**`) **or** declare `/axb-dsl-refine` the CLI end's planner-of-record.
   *(Consolidates grill Q1 + Q3 + Q6.)* **Tracked upstream:**
   [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1).
2. **PM-owned acceptance gap #1** — an Example for **`-d` on a broken/unresolved setup** + the
   non-zero "diagnostic: unresolved" exit. *(Grill Q5.)*
3. **PM-owned acceptance gap #2** — an Example for **no-`-c` + `MODE≠butler` default-path discovery**.
   *(Grill Q5.)*

The **`/axb-data-plan` half is not gated**; the CLI-contract slice is.

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
- **~~Two-wave analysis~~** (data → CLI contract) — **retracted by grill round #2**: the two
  interfaces are information-independent, so the analysis collapses to a **single wave**; the
  `/axb-data-plan` → `/axb-dsl-refine` ordering is **pipeline phasing**, not a Rule-2 dependency.
- **Session-continuity tooling**: `SESSION-BOOTSTRAP.md` gains **Step 8** (read the last 5 days of
  `docs/…/session-summary.md`); new **`SESSION-CLOSEOUT.md`** defines the end-of-day procedure
  (review tree → quality gates → `STATUS.md` → daily summary → reconcile → commit → propagate).
- **Status de-brittling**: `SESSION-BOOTSTRAP.md` ↔ `SESSION-CLOSEOUT.md` cross-linked; the branch
  model above no longer pins head hashes (read live with `git rev-parse`).

## PM follow-ups (spec/acceptance are PM-owned — not written by the RD/butler flow)

**PM TODO checklist** — owner `pm`; source: clarify Q2 + grill round #2 Q5. **Distinct PM tasks: 4**
= 3 required (**2 blocking**) + 1 optional.

### Blocking (gate the `/axb-dsl-refine` CLI slice)

- [ ] **PM-1** — `features/acceptance/version-and-setup-diagnostic.feature`: add an Example for **`-d`
  on a broken/unresolved setup** + the dedicated **non-zero "diagnostic: unresolved"** exit (clarify
  Q1's other half; grill #2 Q5).
- [ ] **PM-2** — `features/acceptance/starting-with-a-configuration.feature`: add the **positive**
  Example for **no-`-c` + `MODE≠butler`** → default `$TELL_ME_HOME/configs/<mode>.yaml` is found and
  resolves (grill #2 Q5).

### Required (tracked cleanup — not blocking)

- [ ] **PM-3** — `spec.md`: reconcile the stale "slice-local input / no `data/**` truth" wording to the
  **released** lock (clarify Q2). Touches ~3–4 sentences: the `**Input**:` line, the **Key Entities →
  Configuration** note, the **Assumptions** line (and, arguably, **FR-002**).

### Optional

- [ ] **PM-4** — smallest vertical addition for user value — `tellme init` (or `tellme config show`).

## Open items (non-blocking)

- **Blocking (grill #2)** — see the *Gating blockers* section above: the cross-repo `aixbdd-tmg`
  decision + the two PM acceptance gaps gate the `/axb-dsl-refine` CLI slice.
- Exact `-d --json` output schema.
- Exit-code numeric values (incl. the new dedicated "diagnostic: unresolved" code).
- Error-message wording (`NFR-004`).
- **Unchecked-error coverage** (round-001 residual): closed next slice by `golangci-lint` + `errcheck`.

## Environment notes

- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for read **and** write.
- **2026-09-10**: session restarted following the `deepseek-flash` (v4.1) provider upgrade — a
  session/environment event only; nothing in round 001 depends on the provider version.
- **No-network sandbox limitation (verified 2026-09-10)**: `unshare -n` / `unshare -rn` fail
  `Operation not permitted` on this dev host — the privileged netns sandbox is a CI/privileged-Linux
  mechanism; local SC-004 uses the unprivileged hostile-DNS/proxy fallback + the build-graph guard.
