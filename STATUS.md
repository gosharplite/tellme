# tellme — Status

**Last updated**: 2026-09-11 (**session 10 — upstream closeout**: grill-#5 methodology issues `aixbdd-tmg#9` and `#10` **resolved upstream** by [PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (Zero Shared Edits + ParallelHint arbitration, ADR 0003) and [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (mandatory Pre-Delivery Orphan Coverage Sweep, ADR 0004); `tellme` round-001 `tasks.md` verified already conformant — no local change; `tellme` **issue [#5](https://github.com/gosharplite/tellme/issues/5)** (Grill Round #5) **closed as completed**; `SESSION-CLOSEOUT.md` executed). *Prior — session 9: `/axb-tasks` executed & grilled (Grill #5); `tasks.md` delivered (63 tasks, 45 per-task stepdef files, sweep 19/19 PASSED).* Next = **`/axb-implement`**)
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)
**Daily log**: [`docs/2026/09/11/session-summary.md`](docs/2026/09/11/session-summary.md)

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from the working branch | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | session tip (moves per commit) | This session's working branch |

> **Heads are intentionally not pinned here** — the working branch is the moving tip and every commit
> on it is followed by the two-step merge `working → dev → main`. Read live heads with
> `git rev-parse --short main dev HEAD` rather than trusting a snapshot.
> **Propagation status (2026-09-11, session 9 closeout):** **DONE.** The two-step merge
> `working → dev → main` carried the **session-9** closeout (`/axb-tasks` delivery, Grill #5, Q5 truth fix,
> 5 ratified decisions) to `dev` and `main`. All three lines now carry the round-001 artifacts through
> the session-9 closeout.
>
> **Propagation status (2026-09-11, session 10):** **PENDING** — the session-10 closeout commit on the
> working branch `001-cli-bootstrap-and-config` (upstream grill-#5 resolution + issue-#5 closure +
> closeout) is committed but **not yet pushed/propagated**; awaiting user approval for the two-step
> merge `working → dev → main`.

## Current round — `001-cli-bootstrap-and-config`

**Scope (narrow foundation, locked)**: CLI boot; YAML configuration load + validation
(`config-valid-provider`); runtime home (`TELL_ME_HOME`) + per-mode session workspace
(`output/<mode>/`); build version (`--version`); offline setup diagnostic (`-d`, `-d --json`).

> **Scope note (updated by grill round #3 + user ratification, 2026-09-11)**: clarify-Q2's "minimal
> `data/**` truth" is **reversed by ratification** — round 001 owes **no** `data/**` model. The config
> input contract and the workspace lifecycle are owned by `/axb-dsl-refine` (the CLI contract owner).

### Artifacts

- [x] `specs/plans/001-cli-bootstrap-and-config/spec.md` *(PM edits landed — PM-1..PM-3; **session 8**: grill-#3 data-scope wording reconciled — no `data/**` model)*
- [x] `specs/plans/001-cli-bootstrap-and-config/checklists/requirements.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/truth-delta.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/features/acceptance/*.feature` (4 journey features) — `/axb-spec-by-example`
- [x] `specs/plans/001-cli-bootstrap-and-config/research.md` — `/axb-technical-research` (7 decisions; **post-grill + post-clarify**; **session 8**: clarify-Q2 data-scope ruling annotated ⟦grill #3⟧ REVERSED)
- [x] `specs/truth/techstack.md` — `/axb-technical-research` (**post-grill**)
- [x] `specs/plans/001-cli-bootstrap-and-config/plan.md` — `/axb-system-analysis` (**revised by grill #2**: 2 interfaces, **1 wave**; **all gating blockers resolved** — CLI-seat blocker closed by `aixbdd-tmg` PR #2; **session 8**: `/axb-data-plan` reconciled to the ratified **NOOP**)
- [x] `specs/truth/data/**` — `/axb-data-plan` = **NOOP** (authored as `data-model.dbml`, then **deleted** by the user-ratified reversal of clarify Q2 — grill round #3)
- [x] `specs/truth/features/cli/**` + `dsl.md` — `/axb-dsl-refine` (CLI **contract owner**; **ungated**) — **4 modules**, interface-root + module DSL; **audit PASSED** (0 errors/0 warnings) — **grill round #4 corrections applied** (restored W1/W2/W5 arrange; pinned the `--json` key schema; error-code rows assert distinctness); **session 8**: W2/D2 re-decided under the upstream **entailment** criterion ([PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7)) → **fold** (identical verdict), recorded as **NOOP**
- [x] `decisions/0001-project-language.md` + `decisions/README.md` — project-language declaration (**named home**, English; upstream PR #8 / R2)
- [x] `specs/plans/001-cli-bootstrap-and-config/tasks.md` — `/axb-tasks` (63 tasks: revised post-grill #5 — per-task stepdef files, sentinel in T006, exit-code oracle, Decision 3 reads, verify targets in T002, and pre-delivery orphan sweep)
- [ ] Implementation — `/axb-implement`

### System analysis (`plan.md`) — **revised by grill round #2**

- **Interfaces (2)**: `CLI end (operator terminal interface)` — **contract owner `/axb-dsl-refine`**
  (carried forward at delivery; no api/data/ui planner); `Configuration & workspace persistence
  interface` — `/axb-data-plan` = **NOOP** (grill #3 + ratification).
- **Wave (1, single)**: both interfaces are information-independent → one parallel wave. *(The earlier
  2-wave split was retracted — no Rule-2 information dependency.)*
- **`/axb-api-plan` = NOOP** (single CLI end, no OpenAPI); **`/axb-ui-plan` skipped** (CLI).
- **Gating blockers** — **ALL RESOLVED** (PM-1/PM-2 landed; the CLI-seat blocker closed by upstream
  `aixbdd-tmg` PR #2). The `/axb-dsl-refine` CLI slice is **ungated**.

### Pipeline position

`/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` (+ **grill round #1**, **clarify**) →
`/axb-system-analysis` **(done — `plan.md`, revised by grill round #2)** → `/axb-data-plan`
**(done — NOOP, grill #3 + ratification)** → `/axb-dsl-refine` **(done — CLI executable contract;
audit PASSED; W2/D2 re-decided → fold)** → `/axb-tasks` **(done — `tasks.md`, 63 tasks, grill #5 closed, Q5 truth fix landed, orphan sweep passed)** → **next: `/axb-implement`**.

**Pending decision:** *(none)* the **grill round on `plan.md`** ran as **grill round #2** (6/6,
verdict *proceed with changes*); its edit set is applied. **All gates cleared:** the two PM acceptance
gaps landed (PM-1/PM-2) and the cross-repo `aixbdd-tmg` CLI-seat blocker is **closed** by upstream PR
[#2](https://github.com/gosharplite/aixbdd-tmg/pull/2) — the `/axb-dsl-refine` CLI slice is
**ungated**.

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
  UNASSIGNED recorded as a gap** (not deleted) — *since resolved: CLI end mapped to its **contract
  owner** `/axb-dsl-refine`*; (c) **`-d`-unresolved / default-path marked not-yet-delegable** —
  *since un-gated (PM-1/PM-2 landed)*; (d) delegation ordering re-attributed to **pipeline phasing**,
  not a Rule-2 dependency; (e) a **Gating blockers** subsection added. *(See the
  **Upstream — `aixbdd-tmg` CLI-seat resolution** section below for the closures.)*
- **Subject retractions** (all conceded under verification): **Q1** `Interface.kind` gap is a present
  blocker, not a "forward risk"; **Q2** the Wave-1→Wave-2 dependency; **Q3**
  `wave-covers-interfaces` ("holds" claim); **Q6** the count-1 fix (definition inverted the skill's
  inventory-then-delegate order). **Q5**: conceded an over-commit.

### Gating blockers (gate the `/axb-dsl-refine` CLI slice)

1. ~~**Cross-repo blocker (`aixbdd-tmg` truth-model owner)**~~ — **RESOLVED** (2026-09-10): upstream
   [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1) is **closed by
   PR [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2)** — `InterfaceKind` gained **`cli`**
   (`features/cli/**`) and `wave-covers-interfaces` now reads *"…either delegated to a planner **in at
   least one `Wave`** or carried forward to its contract owner **at delivery**."* `/axb-dsl-refine` is
   the CLI end's **contract owner** (a forward handoff at delivery, not a planner).
   *(Consolidated grill Q1 + Q3 + Q6.)*
2. ~~**PM-owned acceptance gap #1**~~ — **RESOLVED** (session 5, PM role): the `-d`-unresolved Example
   (plain + `--json`) + the non-zero "diagnostic: unresolved" exit added; a `spec.md` edge case added.
   *(PM-1.)*
3. ~~**PM-owned acceptance gap #2**~~ — **RESOLVED** (session 5, PM role): the positive
   no-`-c` + `MODE≠butler` found-default Example added. *(PM-2.)*

**All three gating blockers are resolved** — the `/axb-dsl-refine` CLI slice is **ungated**.

## Grill round #3 — round-001 data truth (CLOSED)

- **Issue**: [#3](https://github.com/gosharplite/tellme/issues/3) · subject `architect` · griller `griller` · orchestrator `butler` · **6/6 questions** (cap 6)
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/8744cdb9d2f748c070e985cf7d42ce3d>
- **Findings comment**: <https://github.com/gosharplite/tellme/issues/3#issuecomment-5624893374>
- **Verdict**: **proceed with changes** — both the *diagnosis* ("provably meets the trigger") and the
  *design* (`data-model.dbml`) failed under the artifact's own input-vs-state test.
- **Subject retractions/concessions** (all under verification): **Q1** trigger not "provable" — clarify
  Q2 is the sole binding warrant; **Q2** `config-valid-provider` is **file-only** and the **effective
  provider is unmodelled** (false-pass on `TELL_ME_SELECTED_PROVIDER=ghost`); **Q3** the `Ref` is a
  **relational FK over non-relational storage** (a PK enforces uniqueness, not existence); **Q4**
  collapse the two tables → one document-shaped unit, no `Ref`; **Q5** the artifact's content is
  **argued out of `data/**`**; **Q6** **no self-NOOP** — escalate to `/axb-clarify`.
- **Applied (user-ratified 2026-09-11)**: `specs/truth/data/data-model.dbml` **deleted**; the
  `/axb-data-plan` owner section records **`NOOP`**; the CLI input contract + workspace lifecycle hand
  to **`/axb-dsl-refine`** as their single owner. *(Reverse of clarify Q2 → Option 1.)*

## Grill round #4 — round-001 CLI interface truth (CLOSED)

- **Issue**: [#4](https://github.com/gosharplite/tellme/issues/4) · subject `architect` · griller `griller` · orchestrator `butler` · **8/8 questions** (cap 8)
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/7adcc308e12ff10f91cf676fd391b8b5>
- **Findings comment**: <https://github.com/gosharplite/tellme/issues/4#issuecomment-5625335168>
- **Verdict**: **proceed with changes** — the core shapes held (the `cli` kind, root/module DSL split, one-Act-per-Example; coverage genuinely complete 7→17), but two contract defects and two host-rule matters surfaced.
- **In-round fixes applied** (`/axb-dsl-refine`, ungated): (a) restored the dropped configuration `Given` in `workspace` W1/W2/W5 — without it W1/W2 are **unsatisfiable** and W5 asserts the **wrong exit class** under the ratified resolver (a Rule-5/6 arrange loss); (b) **pinned the `--json` key schema** in the two `diagnostics` structured-output rows; (c) reworded the four error-code rows to assert **distinctness** (FR-014), not a literal. Audit re-run → **PASSED**.
- **Routed residuals** (host-rule matters — not locally editable): **(R1) the atomicity convention** (Rule 2's fold/split boundary is not mechanically derivable) and **(R2) the English override** (STANDARDS §2/§3 unconditional; no warrant in either repo). Both route `/axb-clarify` → `aixbdd-tmg` issue → PR. **Opened upstream**: [aixbdd-tmg#5](https://github.com/gosharplite/aixbdd-tmg/issues/5) (R1), [aixbdd-tmg#6](https://github.com/gosharplite/aixbdd-tmg/issues/6) (R2). *See Open items.* **⟦RESOLVED session 8 — R1 by [PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7), R2 by [PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8).⟧**
- **Subject retractions/concessions** (under verification): **Q1** "six" was a **miscount** (7); coverage holds via the enumerated 7→17 map; **Q2/Q3** the "subject" atomicity discriminator is **not mechanical** (W2 ≡ D2) → the D2 split was retracted and the boundary routed upstream (R1); **Q5** the axis-6 config "hole" was **wrong** (the shape is owned by `/axb-dsl-refine` via `cli/dsl.md`); **Q6** exit-code **values = implementation** (blocker retracted), `--json` **schema = contract** (fixed); **Q7** confirmed the root `When` is fine and **W1/W2/W5 under-arranged** (fixed); **Q8** the English override has **no ratified home** (routed, R2).

## Grill round #5 — round-001 `tasks.md` task breakdown (CLOSED)

- **Issue**: [#5](https://github.com/gosharplite/tellme/issues/5) · subject `architect` · griller `griller` · orchestrator `butler` · **8/8 questions** (cap 8)
- **Full transcript (gist)**: <https://gist.github.com/gosharplite/b04e559d2451d8965a87e24c7a62b0d9>
- **Findings comment**: <https://github.com/gosharplite/tellme/issues/5#issuecomment-5627065931>
- **Verdict**: **proceed with changes** — the *structural* diagnosis held (62 tasks; 45-row ↔ 8+9+12+14+2 coverage; W2/D2 NOOP; one-batch/review-last Hint; every parity claim verified), but the *readiness* diagnosis did **not**: `tasks.md` is not executable as-is.
- **Bound edit set** (applies to `tasks.md`):
  - **Q1/Q2** — Phase-3 stepdef landing/registration was unbound; bind **one step file per Phase-3 task** (45 files) with `init()`-based self-registration into a `registrars` slice + `RegisterAll`; `scenario_context.go`/run-helper/sentinel in a **leaf package** (avoid the `steps → e2e → steps` cycle); new Foundational pin task; every `T009–T053` `Read` gains its landing file.
  - **Q3** — import `internal/cli` as the **gray-box exit-code oracle** + `TestExitCodesAreDistinct` (pairwise + distinct-from-zero), bound to **T006**; record in `research.md` D4/D5.
  - **Q4** — `research.md -> Decision 3` into 4A+4B(+4C) `Shared Must Read`; quote the two clauses into Boundary; disclose the order-sensitive overlap as a residual.
  - **Q5 — TRUTH DEFECT → `/axb-dsl-refine`** — `{workspace_path}` is declared home-*relative* but composed `{home}/{workspace_path}` against `"ait-tmg/output/butler"`; bind reading **(b)** (`{home}`-rooted operator paths); **8 rows** affected; re-word `workspace/dsl.md` + `diagnostics/dsl.md`; **4B/4C Green block behind it**; `tasks.md` must not encode a normalizer until settled.
  - **Q6** — body-conformance criterion in **T054** + Phase-3 Boundary + the **Markers** `[BDD-RED]` clause (the gate proved only *wiring*).
  - **Q7** — the `VERSION=0.0.0-harness` sentinel is orphaned; bind it into **T005** (value + explicit `go build -ldflags` + assertion + Decision 6 Read); bar any build path that can omit the sentinel.
  - **Q8** — `verify-no-test-sleep` (D7) + the techstack `verify` target are orphans; bind **T002**; reconcile D5(2)'s "Makefile gate" wording vs T008's Go test; **run the orphan-coverage sweep as the pre-delivery gate**.
- **Systemic finding**: three distinct **orphaned artifacts** (Q4 D3; Q7 D6 + "Version assertion"; Q8 D7 + "Task runner") — a truth row / plan Decision that binds behaviour no task `Read` reaches or delivers. Sweep recommended as a `tasks.md` pre-delivery gate, and upstream for the `axb-tasks` Phase-5 self-check (R1/R2 pattern).
- **Status**: **DELIVERED** — Q5 truth fix applied via `/axb-dsl-refine` (`workspace/dsl.md` + `diagnostics/dsl.md`, audit passed); `tasks.md` revised (63 tasks, Q1–Q4/Q6–Q8 applied); all 5 open questions decided by user (reading b ratified, leaf package `harness`, Go test for build guard, host-harness exit-code test, upstream issues routed); pre-delivery orphan-coverage sweep **PASSED** (19/19); upstream issues [#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) and [#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) created in `aixbdd-tmg`. Next: `/axb-implement`. **⟦RESOLVED session 10: [#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) → upstream [PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (ADR 0003); [#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) → [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (ADR 0004).⟧**

## Upstream — `aixbdd-tmg` CLI-seat resolution (session 6, CLOSED)

`aixbdd-tmg#1` (the round-001 cross-repo Gating blocker #1) is **closed** by
[PR #2](https://github.com/gosharplite/aixbdd-tmg/pull/2) — *"give the CLI end a truth-tree home and a
contract owner"* — merged to `main` (2 commits, 8 files): the issue's **Option B**, with two design
amendments.

- **`InterfaceKind` gains `cli`** → executable truth lives under `specs/truth/features/cli/**`
  (`Interface.definition` now `backend (API), frontend (web), or CLI (terminal)`).
- **`wave-covers-interfaces` reworded** → *"…either delegated to a planner **in at least one `Wave`**
  or carried forward to its contract owner **at delivery**."*
- **Locus pinned to delivery** — the CLI end is a **forward handoff at delivery**, **not** a `Wave`
  delegation; **`/axb-dsl-refine` is the CLI end's *contract owner*, not a "planner."**
- Cross-file coherence landed in `axb-system-analysis/SKILL.md` (+ `分析介面委派與planner對應判準.md`
  Rule 2 → `planner／contract-owner 對應`, using 承接／交棒), `axb-dsl-refine/SKILL.md`, `README.md` §3,
  and `axb-tasks/templates/tasks.md` (`Core Inputs` → root-agnostic `specs/truth/features/**`).
- **Verification**: `modelith lint` → 0 errors/0 warnings; `modelith render --check` → up to date.

**Paired `tellme` follow-up (this session)**: round-001 `plan.md` dropped the CLI-end
"Planner: **UNASSIGNED — recorded gap**" (now the **contract-owner** mapping) and moved Gating blocker
#1 → **resolved**; `STATUS.md` synced. **The `/axb-dsl-refine` CLI slice is un-gated.**

## Upstream — R1/R2 resolution (session 8, CLOSED)

Both round-001 host-rule residuals (raised by grill round #4) are **resolved upstream**, merged, and the paired tellme follow-ups are applied.

- **R1 → [aixbdd-tmg PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7)** (closes [#5](https://github.com/gosharplite/aixbdd-tmg/issues/5)): Rule 2 now states a single **entailment** criterion — *one named subject; a `Then`/`And` folds in iff entailed by the subject's outcome over **valid domain states** (never injected bugs)* — plus a **calibration set**; the repo gained a `decisions/` **ADR mechanism** (`0001-atomicity-fold-split-criterion`). `STANDARDS.md` §2 and `axb-spec-by-example` Rule 4 aligned (title test demoted to a smell).
- **R2 → [aixbdd-tmg PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8)** (closes [#6](https://github.com/gosharplite/aixbdd-tmg/issues/6)): `STANDARDS.md` gains a **Project Language** clause — default 繁體中文, but language is **project-declared** with a **named home** (`.agents/constitution/shared.md` → project ADR → `spec.md` constraint); the §4/§5 DSL meta-schema tokens + §5.2 channel labels stay **fixed** (the topology audit requires the first header to be exactly `DSL 句型`).

**Paired `tellme` follow-ups applied (session 8):**

- **R2 home** — created [`decisions/0001-project-language.md`](decisions/0001-project-language.md) (+ [`decisions/README.md`](decisions/README.md)) declaring **English** as tellme's artifact language (the named home; previously only a status-log note). tellme's `dsl.md` headers already keep the fixed Chinese meta-schema tokens, so the audit stays green.
- **R1 re-decision → NOOP** — `/axb-dsl-refine` re-decided **W2/D2** under the entailment criterion → **fold** for both (structurally isomorphic ⇒ identical verdict, as the calibration set requires); recorded in `truth-delta.md`. **No feature/DSL body change.**
- **Doc reconcile** — `spec.md` (revision note + Key Entities + Assumptions), `plan.md`, and `research.md` reconciled to the ratified `data/**` NOOP (grill #3).

## Upstream — grill #5 methodology resolution (session 10, CLOSED)

Both upstream methodology issues that grill #5 routed to `aixbdd-tmg` (round-001 `tasks.md` review) are **resolved** — merged to `main` — and the paired `tellme` conformance check is confirmed.

- **[aixbdd-tmg#9](https://github.com/gosharplite/aixbdd-tmg/issues/9)** → **[PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11)** — `fix(rules): ParallelHint concurrency arbitration and Zero Shared Edits + ADR 0003`. Removed the `後寫入者自己 merge` contradiction. Two-pronged fix: upstream **Zero Shared Edits** (`axb-tasks` Rule 5, `SHOULD`: 1 task = 1 independent landing file) + downstream **arbitration** (`axb-implement`: disjoint files dispatched in parallel; shared files serialized per-file or grouped into one subagent; **review is audit-only**). Rule file **renamed** `ParallelHint平行Subagent與衝突Merge判準.md` → **`ParallelHint平行Subagent與同檔調度判準.md`**; `平行執行與檔案衝突判準.md` Rule 2 + Bad Example updated; `axb-tasks` templates now model 7 independent landing files. **ADR [0003](https://github.com/gosharplite/aixbdd-tmg/blob/main/decisions/0003-parallel-hint-concurrency-arbitration.md)** (Accepted).
- **[aixbdd-tmg#10](https://github.com/gosharplite/aixbdd-tmg/issues/10)** → **[PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12)** — `fix(tasks): Pre-Delivery Orphan Coverage Sweep for tasks.md + ADR 0004`. Made the **Pre-Delivery Orphan Coverage Sweep** a mandatory `axb-tasks` Phase 5 gate: non-NOOP `truth-delta.md` rows, decided `research.md` Decisions, and changed `techstack.md` sections must all be task-`Read`-covered or task-delivered (NOOP / empty-set exempt) or delivery is blocked. New rule `Pre-Delivery覆蓋檢驗與孤立產物盤點判準.md`. **ADR [0004](https://github.com/gosharplite/aixbdd-tmg/blob/main/decisions/0004-pre-delivery-orphan-coverage-sweep.md)** (Accepted).
- **`tellme` conformance (verified, no artifact change)**: round-001 `tasks.md` **already embodies both** — Phase 3 lands **45 independent per-task stepdef files** (`tests/e2e/steps/step_t###_*.go`) = the endorsed **Zero Shared Edits** pattern (Parallel Hint T010–T054 dispatched in parallel; review T055 last), and its **Pre-Delivery Orphan Coverage Sweep** section passed **19/19**. No stale reference to the renamed rule file. **Neither issue blocks `/axb-implement`; both are now closed.**
- **Tracking issue** — [`tellme#5`](https://github.com/gosharplite/tellme/issues/5) (Grill Round #5) **closed as completed** (2026-09-11, session 10): findings applied, `tasks.md` delivered (63 tasks), orphan sweep **19/19 PASSED**; closing summary comment posted.

## Clarify round — PM-boundary rulings (CLOSED)

Resolved the two items grill round #1 routed to `/axb-clarify`, **before** `/axb-system-analysis`
(grounded against `tell-me-go`'s current behavior):

- **Q1 → Option 3** — `-d` diagnostic semantics: always produce the report (resolved *or* unresolved) and
  exit a **dedicated, distinct non-zero "diagnostic: unresolved"** code; the `FR-014` four codes
  (success/usage/config/environment) bind the **boot** path only.
- **Q2 → Option 1** — the "no `data/**` truth this round" scope lock is **released**; `/axb-data-plan`
  authors a **minimal** data truth (config input contract + workspace lifecycle).
  **⟦REVERSED 2026-09-11 — grill round #3 + user ratification: round 001 owes NO `data/**` model;
  `specs/truth/data/**` deleted, `/axb-data-plan` = NOOP.⟧**

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
- **Data scope reversed** (2026-09-11, grill #3 + user ratification): the "minimal `data/**` truth"
  owed by Q2 is **withdrawn** — round 001 owes **no** data model; `specs/truth/data/**` **deleted**,
  `/axb-data-plan` = **NOOP**; the CLI input contract + workspace lifecycle are owned by
  `/axb-dsl-refine`.
- **~~Two-wave analysis~~** (data → CLI contract) — **retracted by grill round #2**: the two
  interfaces are information-independent, so the analysis collapses to a **single wave**; the
  `/axb-data-plan` → `/axb-dsl-refine` ordering is **pipeline phasing**, not a Rule-2 dependency.
- **Session-continuity tooling**: `SESSION-BOOTSTRAP.md` gains **Step 8** (read the last 5 days of
  `docs/…/session-summary.md`); new **`SESSION-CLOSEOUT.md`** defines the end-of-day procedure
  (review tree → quality gates → `STATUS.md` → daily summary → reconcile → commit → propagate).
- **Status de-brittling**: `SESSION-BOOTSTRAP.md` ↔ `SESSION-CLOSEOUT.md` cross-linked; the branch
  model above no longer pins head hashes (read live with `git rev-parse`).
- **CLI-seat blocker resolved upstream** (`aixbdd-tmg` PR #2): `InterfaceKind` gains `cli`
  (`features/cli/**`); `wave-covers-interfaces` reworded to allow a **contract-owner handoff at
  delivery**; `/axb-dsl-refine` is the CLI end's **contract owner** (not a planner). Round-001
  `plan.md` Gating blocker #1 → resolved; the CLI slice is **ungated**.
- **Project-language declaration home** (session 8, R2): `decisions/0001-project-language.md` declares **English** as tellme's artifact language — the **named home** the upstream Project Language clause requires (previously recorded only in `STATUS.md` / the daily log). The fixed §4/§5 DSL meta-schema tokens are unchanged.
- **W2/D2 re-decided → fold** (session 8, R1): under upstream Rule 2's **entailment** criterion the workspace-reuse pair (W2) and the config-resolution pair (D2) both **fold** (identical verdict) — grounded in the contract's domain definitions (FR-008 + NFR-002 for W2; FR-001/FR-003 for D2); recorded as a **NOOP** in `truth-delta.md`.
- **`tasks.md` 63-task plan delivered** (session 9, post-grill #5): Phase 1 Setup (T001–T003), Phase 2 Foundational (T004–T009), Phase 3 Test Alignment (T010–T054 + T055 review), Phase 4A–4D (T056–T063).
- **Stepdef landing convention** (session 9, Q1/Q2): one step file per task in `tests/e2e/steps/` with `init()` self-registration into `registrars` slice + `RegisterAll` (zero shared edits under concurrent batch dispatch).
- **Acyclic test harness graph** (session 9, Q2 ratified): leaf package `tests/e2e/harness/` for process helper and `SentinelVersion = "0.0.0-harness"` constant; `steps` and `suite_test.go` both import `harness`.
- **Exit-code oracle & distinctness** (session 9, Q3/Q4 ratified): `internal/cli` imported as expected exit-code oracle + `TestExitCodesAreDistinct` (pairwise + distinct-from-zero) in `tests/e2e/exitcode_test.go` (T007).
- **Build-graph guard form** (session 9, Q3 ratified): Go test in `tests/e2e/network_guard_test.go` invoked via `make verify` (T002).
- **`{workspace_path}` reading (b) ratified** (session 9, Q5): operator-facing `{home}`-rooted paths normalized via step definition name-substitution; Q5 truth fix landed in `workspace/dsl.md` and `diagnostics/dsl.md`.
- **Pre-delivery orphan-coverage sweep** (session 9, Q8): 19/19 truth rows and research decisions verified covered by task `Read` or delivery.
- **Upstream methodology tracking** (session 9, Q5; **⟦resolved session 10⟧**): issues [#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) and [#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) opened in `gosharplite/aixbdd-tmg` — both **closed** by [PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (**ADR 0003**, Zero Shared Edits + ParallelHint concurrency arbitration) and [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (**ADR 0004**, Pre-Delivery Orphan Coverage Sweep). No `tellme` artifact change required.

## PM follow-ups (spec/acceptance are PM-owned) — **CLOSED** (session 5, PM role)

**PM TODO checklist** — owner `pm`; source: clarify Q2 + grill round #2 Q5. **4/4 closed** on
2026-09-10 (PM role): PM-1..PM-3 landed, PM-4 decided (deferred).

### Blocking (gated the `/axb-dsl-refine` CLI slice) — CLOSED

- [x] **PM-1** — `features/acceptance/version-and-setup-diagnostic.feature`: added the **`-d`-unresolved**
  Example (plain + `--json`) with the non-zero "diagnostic: unresolved" exit; `spec.md` gained the
  matching edge case.
- [x] **PM-2** — `features/acceptance/starting-with-a-configuration.feature`: added the positive
  **no-`-c` + `MODE≠butler`** found-default Example.

### Required (tracked cleanup) — CLOSED

- [x] **PM-3** — `spec.md`: reconciled the stale "slice-local / no `data/**` truth" wording to the
  released lock (clarify Q2) — the `**Input**:` line, **FR-002**, **Key Entities → Configuration**, and
  the **Assumptions** line.

### Optional — DECIDED (deferred)

- [x] **PM-4** — **deferred to a future round** (candidate: `tellme init` — first-run value via a
  generated default config). **Not** added to round 001: its scope is locked and
  `fresh-package-per-round` requires a new plan package for new scope.

## Open items (non-blocking)

- **Blocking (grill #2)** — **CLEARED.** Both PM acceptance gaps are resolved (PM-1/PM-2) and the
  cross-repo `aixbdd-tmg` CLI-seat decision is **closed** (PR
  [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2)). No gating blockers remain.
- Exact `-d --json` output schema.
- Exit-code numeric values (incl. the new dedicated "diagnostic: unresolved" code).
- Error-message wording (`NFR-004`).
- **Unchecked-error coverage** (round-001 residual): closed next slice by `golangci-lint` + `errcheck`.
- **Host-rule residuals (grill #4)** — **(R1) the atomicity convention** (Rule 2's fold/split boundary is not mechanically derivable) and **(R2) the English override** (STANDARDS §2/§3 are unconditional; no ratified warrant in either repo). Both are **out of `/axb-dsl-refine`'s writ** → route `/axb-clarify` → `aixbdd-tmg` issue → PR (the PR #2 route). **Opened upstream 2026-09-11**: [aixbdd-tmg#5](https://github.com/gosharplite/aixbdd-tmg/issues/5) (R1) and [aixbdd-tmg#6](https://github.com/gosharplite/aixbdd-tmg/issues/6) (R2). They **do not gate** `/axb-tasks`; round-001 ships the CLI contract unchanged on these two points. **⟦RESOLVED (session 8)⟧** — R1 closed by [aixbdd-tmg PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7) (entailment criterion), R2 by [PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8) (Project Language clause); tellme followed up with `decisions/0001-project-language.md` (English home) and the W2/D2 re-decision (fold → recorded as NOOP in `truth-delta.md`). Nothing outstanding here.
- **Upstream methodology tracking (grill #5)** — [aixbdd-tmg#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) (`ParallelHint` same-file merge risk under concurrent dispatch) and [aixbdd-tmg#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) (`axb-tasks` Phase 5 missing Pre-Delivery Orphan Coverage Sweep gate). **⟦RESOLVED (session 10)⟧** — closed by [aixbdd-tmg PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (Zero Shared Edits + ParallelHint concurrency arbitration, ADR 0003) and [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (mandatory Pre-Delivery Orphan Coverage Sweep, ADR 0004). `tellme` round-001 `tasks.md` verified already conformant (45 independent stepdef files; sweep 19/19 PASSED) — no local change required.
- **DECIDED (grill #4)**: exit-code **numeric values** = an **implementation** choice (FR-014 requires only *distinct + deterministic*); the `--json` **key schema** is now **pinned** in `specs/truth/features/cli/diagnostics/dsl.md`.

## Environment notes

- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for read **and** write.
- **2026-09-10**: session restarted following the `deepseek-flash` (v4.1) provider upgrade — a
  session/environment event only; nothing in round 001 depends on the provider version.
- **No-network sandbox limitation (verified 2026-09-10)**: `unshare -n` / `unshare -rn` fail
  `Operation not permitted` on this dev host — the privileged netns sandbox is a CI/privileged-Linux
  mechanism; local SC-004 uses the unprivileged hostile-DNS/proxy fallback + the build-graph guard.
- **2026-09-11 (session 7)**: grill rounds **#3** and **#4** ran via `tell-me-go` sub-agents (`architect`,
  `griller`) seeded with `SESSION-BOOTSTRAP.md`; both authorized the reference repos via `register_readpath`.
  Issue/gist/PR workflow via `gh` (token scopes `gist` + `repo`). The feature/DSL topology audit ran via
  `python3` (script at `$TELL_ME_HOME/docs/skills/axb-gherkin-and-dsl/scripts/`); `uv` is present.
- **2026-09-11 (session 8)**: upstream `aixbdd-tmg` [PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7) (R1 — atomicity entailment criterion + ADR mechanism) and [PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8) (R2 — Project Language clause) read and confirmed **merged**; local skill copies under `$TELL_ME_HOME/docs/skills/` verified **synced** with both. Created tellme's `decisions/` ADR home (project language).
- **2026-09-11 (session 10)**: upstream `aixbdd-tmg` [PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (grill-#5 [#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) — Zero Shared Edits + ParallelHint concurrency arbitration, ADR 0003) and [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (grill-#5 [#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) — Pre-Delivery Orphan Coverage Sweep, ADR 0004) read and confirmed **merged**; `tellme` round-001 `tasks.md` verified **already conformant** (45 independent per-task stepdef files; orphan sweep 19/19 PASSED) — **no artifact change required**. `SESSION-BOOTSTRAP.md` re-executed; active branch `001-cli-bootstrap-and-config` confirmed current. `tellme` [#5](https://github.com/gosharplite/tellme/issues/5) closed as completed; `SESSION-CLOSEOUT.md` executed.
