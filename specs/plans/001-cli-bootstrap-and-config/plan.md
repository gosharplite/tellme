# System Analysis Plan — round 001 (`001-cli-bootstrap-and-config`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/001-cli-bootstrap-and-config/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── starting-with-a-configuration.feature
│       ├── runtime-home-and-session-workspace.feature
│       ├── version-and-setup-diagnostic.feature
│       └── unsupported-cli-usage.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — done
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── <interface>/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` in the document structure — see the `NOOP` /axb-api-plan note below.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # entrypoint; `var version` is the single `-X main.version` target

internal/
├── cli/                           # flag parsing (pflag), exit-code classification, `-d` dispatch
├── config/                        # YAML load + validate + effective-value resolver (TELL_ME_* precedence)
└── home/                          # TELL_ME_HOME resolution + `output/<mode>/` workspace lifecycle

tests/e2e/                         # godog suite driving the built binary (per-scenario isolated TELL_ME_HOME)
Makefile                           # fmt, tidy, build, test, verify (staticcheck, no-network guard)
go.mod / go.sum
```

**Structure Decision**: a **single CLI end** — Go 1.26 module `github.com/gosharplite/tellme`, entrypoint
`cmd/tellme/`, non-public logic in `internal/{cli,config,home}` (one FR cluster each — `research.md`
Decision 1). There is **no web UI and no HTTP/OpenAPI server**. The executable CLI contract lives in
`specs/truth/features/**` (godog, driven from `tests/e2e/`). There is **no `data/**` truth this round**
(`/axb-data-plan` = **NOOP** — grill round #3 + user ratification; see *Scope notes* below).

## Analysis Plan

> **Revised by grill round #2** (issue [#2](https://github.com/gosharplite/tellme/issues/2); subject
> `architect`, griller `griller`, 6/6 questions; verdict **proceed with changes**). Three corrections
> landed below: (a) the two-wave split **collapses to a single wave** (no information dependency —
> `/axb-dsl-refine`'s READ contract excludes `specs/truth/data/**`); (b) the interface count is
> retained at **2** with the **CLI end's planner recorded as unassigned** (not deleted);
> (c) the **`-d`-unresolved / default-path scope is marked not-yet-delegable** pending PM acceptance
> Examples. See *Gating blockers* at the end of this section.
>
> **Follow-up — CLI-seat blocker RESOLVED (upstream, 2026-09-10).** The `aixbdd-tmg` truth-model owner
> closed the CLI-seat gap:
> [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1) → PR
> [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2) added **`InterfaceKind: cli`**
> (`features/cli/**`) and reworded `wave-covers-interfaces` to *"…either delegated to a planner **in
> at least one `Wave`** or carried forward to its contract owner **at delivery**."* The CLI end is
> thus a **contract-owner handoff** to `/axb-dsl-refine` (a forward handoff at delivery — **not** a
> `Wave` delegation and **not** a planner). The two PM acceptance Examples (correction **c**) have
> also landed (**PM-1/PM-2**), so **all three gating blockers are now resolved** and the
> `/axb-dsl-refine` CLI slice is **ungated**.

### System interface inventory

This requirement yields **2** system interfaces.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: command invocation and flags (`-c`/`--config`, `-d`, `--json`, `--version`);
     the exit-code taxonomy (success / usage / configuration / environment, plus the dedicated
     "diagnostic: unresolved" code); stdout / stderr / stdin streams; and the console output
     (readiness, resolved workspace path, build version, diagnostic report).
   - Requirement evidence: `FR-001`–`FR-015`, `NFR-003`/`NFR-004`; the four acceptance features
     (`starting-with-a-configuration`, `runtime-home-and-session-workspace`,
     `version-and-setup-diagnostic`, `unsupported-cli-usage`).
   - Planner: **none** — no `axb-system-analysis` planner takes a terminal endpoint: `/axb-api-plan`
     is `NOOP`, `/axb-ui-plan` is skipped, and `/axb-data-plan` is N/A (this is not
     entity/field/lifecycle/storage). Instead, its **contract owner** is **`/axb-dsl-refine`**, which
     produces the executable contract in the next pipeline phase. **Covered** by
     `wave-covers-interfaces` via its *carried-forward-to-its-contract-owner **at delivery*** branch
     (a forward handoff — **not** a `Wave` delegation). *Resolved by upstream
     [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1) → PR
     [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2).*

2. `Configuration & workspace persistence interface`
   - Endpoint type: `local-file / state endpoint`
   - Primary interface: the **config input contract** (`$TELL_ME_HOME/configs/<mode>.yaml`: `MODE`,
     `PERSON`, `SELECTED_PROVIDER`, `PROVIDERS` + `TELL_ME_*` env precedence) and the **workspace/state
     lifecycle** (`$TELL_ME_HOME/output/<mode>/` — created on first run, reused, must be a directory,
     idempotent).
   - Requirement evidence: `FR-002`/`003`/`005`/`006`/`007`/`008`/`009`/`015`, `NFR-001`/`NFR-002`;
     `spec.md` Assumptions.
   - Planner: **`/axb-data-plan`** = **`NOOP`** (grill round #3 + user ratification: no `data/**` model
     this round). The config input contract and the workspace lifecycle are owned by `/axb-dsl-refine`
     (`specs/truth/features/cli/**`).

> **Scope notes.** `spec.md` originally locked "no `data/**` truth this round"; clarify Q2 **released**
> that lock (→ Option 1) and authorized a **minimal** data truth, but grill round #3 argued its content
> out of `data/**` and the user **ratified reversing clarify Q2** (2026-09-11) — so round 001 owes **no**
> `data/**` model and `/axb-data-plan` = **`NOOP`** (see the truth-delta `/axb-data-plan` section). The
> config input contract + workspace lifecycle are owned by `/axb-dsl-refine`. There is **no API/HTTP
> interface** — a CLI has a single end, so `/axb-api-plan` is **`NOOP`** — and **no UI interface** —
> `/axb-ui-plan` is **skipped** (CLI-streamlined workflow).

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Configuration & workspace persistence interface`
- Analysis focus:
  - **Persistence** → **`NOOP`**: no `data/**` model this round (grill round #3 + user ratification).
    The config input contract (keys, types, `TELL_ME_*` precedence) and the workspace/state lifecycle
    (`output/<mode>/`) are executable CLI-contract behaviour owned by `/axb-dsl-refine`.
  - **CLI end** → produce the **executable CLI contract** — interface Gherkin + DSL for the
    config-resolution order (the 6-step contract, `research.md` Decision 3), workspace initialization,
    `--version`, and the **resolved** `-d` / `-d --json` reporting — driven E2E against the built
    binary. The CLI end is **carried forward at delivery** to its contract owner `/axb-dsl-refine`;
    the unresolved `-d` and default-path behaviour is **ungated** (PM acceptance Examples landed —
    PM-1/PM-2).
- Scheduling rationale: the two interfaces are **information-independent** — both read the *same*
  upstream sources (`spec.md` `FR-003`/`005`/`007`/`015` and `research.md` Decision 3), and neither
  consumes the other's output (`/axb-dsl-refine`'s READ contract does **not** include
  `specs/truth/data/**`). Per `Wave依賴排序與平行分組判準.md` Rule 2 they therefore share a **single
  wave**. *Grill round #2 retracted the earlier Wave 1 → Wave 2 split: it was a semantic hand-wave,
  not an information-supply dependency.*

### Delegation order

1. **`/axb-data-plan`** — Wave 1 (`Configuration & workspace persistence interface`) → **`NOOP`**
   (no `data/**` model: grill round #3 + user ratification); the config input contract + workspace
   lifecycle move to the CLI contract owner below.
2. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` (no api/data/ui
   planner applies; **not** a `Wave` delegation) → `specs/truth/features/cli/**` + `dsl.md`.

> The order above is the **CLI-streamlined pipeline's phase sequence** (`aixbdd-tmg/README.md`:
> `/axb-system-analysis` + `/axb-data-plan`, then the next phase `/axb-dsl-refine`) — **not** a Rule-2
> information dependency within the single wave. The CLI end is covered by `wave-covers-interfaces`
> via its **"carried forward to its contract owner at delivery"** branch.

Not delegated: `/axb-ui-plan` (skipped — no UI) and `/axb-api-plan` (`NOOP` — single CLI end, no
OpenAPI surface).

### Gating blockers (gate `/axb-dsl-refine` — the CLI-end executable contract)

1. ~~**Cross-repo blocker (`aixbdd-tmg` truth-model owner) — no seat for a CLI end.**~~ **RESOLVED**
   (upstream, 2026-09-10): [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1)
   is **closed by PR [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2)** — `InterfaceKind` gained
   **`cli`** (`features/cli/**`) and `wave-covers-interfaces` now reads *"…either delegated to a
   planner **in at least one `Wave`** or carried forward to its contract owner **at delivery**."*
   `/axb-dsl-refine` is the CLI end's **contract owner** (a forward handoff at delivery, **not** a
   planner). *(Consolidated grill Q1 + Q3 + Q6.)*
2. ~~**PM-owned acceptance gap #1 — `-d` on a broken/unresolved setup.**~~ **RESOLVED** (session 5, PM
   role): `version-and-setup-diagnostic.feature` now carries the `-d`-unresolved Example (plain +
   `--json`) with the dedicated non-zero "diagnostic: unresolved" exit, and `spec.md` adds the matching
   edge case. *(Was grill Q5; closed as PM-1.)*
3. ~~**PM-owned acceptance gap #2 — default-path discovery.**~~ **RESOLVED** (session 5, PM role):
   `starting-with-a-configuration.feature` now carries the positive no-`-c` + `MODE≠butler`
   found-default Example. *(Was grill Q5; closed as PM-2.)*

**All three gating blockers are resolved** — the CLI-end slice is **ungated**. The persistence half
(single wave → `/axb-data-plan`) was never gated.
