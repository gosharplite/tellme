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
├── data/                          # minimal data truth — /axb-data-plan
│   └── *.dbml
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
`specs/truth/features/**` (godog, driven from `tests/e2e/`), and the round's minimal persisted-state
truth lives in `specs/truth/data/**`.

## Analysis Plan

> **Revised by grill round #2** (issue [#2](https://github.com/gosharplite/tellme/issues/2); subject
> `architect`, griller `griller`, 6/6 questions; verdict **proceed with changes**). Three corrections
> landed below: (a) the two-wave split **collapses to a single wave** (no information dependency —
> `/axb-dsl-refine`'s READ contract excludes `specs/truth/data/**`); (b) the interface count is
> retained at **2** with the **CLI end's unassigned planner recorded as a gap** (not deleted);
> (c) the **`-d`-unresolved / default-path scope is marked not-yet-delegable** pending PM acceptance
> Examples. See *Gating blockers* at the end of this section.

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
   - Planner: **UNASSIGNED — recorded gap** (consolidated blocker). No `axb-system-analysis` planner
     takes a terminal endpoint: `/axb-api-plan` is `NOOP`, `/axb-ui-plan` is skipped, and
     `/axb-data-plan` is N/A (this is not entity/field/lifecycle/storage). Its executable contract is
     produced by the **next pipeline phase `/axb-dsl-refine`** — a phase/truth owner, **not** an
     api/data/ui planner. Under the invariant's strict wording (`wave-covers-interfaces` — "every
     interface … delegated to a **planner**"), this interface is therefore **not covered**; see
     *Gating blockers*.

2. `Configuration & workspace persistence interface`
   - Endpoint type: `local-file / state endpoint`
   - Primary interface: the **config input contract** (`$TELL_ME_HOME/configs/<mode>.yaml`: `MODE`,
     `PERSON`, `SELECTED_PROVIDER`, `PROVIDERS` + `TELL_ME_*` env precedence) and the **workspace/state
     lifecycle** (`$TELL_ME_HOME/output/<mode>/` — created on first run, reused, must be a directory,
     idempotent).
   - Requirement evidence: `FR-002`/`003`/`005`/`006`/`007`/`008`/`009`/`015`, `NFR-001`/`NFR-002`;
     `spec.md` Assumptions.
   - Planner: **`/axb-data-plan`** → minimal `specs/truth/data/**`.

> **Scope notes.** `spec.md` originally locked "no `data/**` truth this round"; that lock was
> **released** (clarify Q2 → Option 1), so interface 2 now delegates a **minimal** data truth.
> There is **no API/HTTP interface** — a CLI has a single end, so `/axb-api-plan` is **`NOOP`** —
> and **no UI interface** — `/axb-ui-plan` is **skipped** (CLI-streamlined workflow).

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Configuration & workspace persistence interface`
- Analysis focus:
  - **Persistence** → model the **minimal data truth**: the config input contract (keys, types,
    `TELL_ME_*` precedence) and the workspace/state lifecycle (`output/<mode>/`), in
    `specs/truth/data/**`.
  - **CLI end** → produce the **executable CLI contract** — interface Gherkin + DSL for the
    config-resolution order (the 6-step contract, `research.md` Decision 3), workspace initialization,
    `--version`, and the **resolved** `-d` / `-d --json` reporting — driven E2E against the built
    binary. The unresolved `-d` and default-path behaviour is **gated** (see *Gating blockers*).
- Scheduling rationale: the two interfaces are **information-independent** — both read the *same*
  upstream sources (`spec.md` `FR-003`/`005`/`007`/`015` and `research.md` Decision 3), and neither
  consumes the other's output (`/axb-dsl-refine`'s READ contract does **not** include
  `specs/truth/data/**`). Per `Wave依賴排序與平行分組判準.md` Rule 2 they therefore share a **single
  wave**. *Grill round #2 retracted the earlier Wave 1 → Wave 2 split: it was a semantic hand-wave,
  not an information-supply dependency.*

### Delegation order

1. **`/axb-data-plan`** — Wave 1 (`Configuration & workspace persistence interface`) → minimal
   `specs/truth/data/**`.
2. **`/axb-dsl-refine`** *(next phase)* — Wave 1 (`CLI end`) → `specs/truth/features/**` + `dsl.md`.

> The order above is the **CLI-streamlined pipeline's phase sequence** (`aixbdd-tmg/README.md`:
> `/axb-system-analysis` + `/axb-data-plan`, then the next phase `/axb-dsl-refine`) — **not** a Rule-2
> information dependency within the single wave.

Not delegated: `/axb-ui-plan` (skipped — no UI) and `/axb-api-plan` (`NOOP` — single CLI end, no
OpenAPI surface).

### Gating blockers (gate `/axb-dsl-refine` — the CLI-end executable contract)

1. **Cross-repo blocker (`aixbdd-tmg` truth-model owner) — no seat for a CLI end.** The typed model
   has neither a valid `InterfaceKind` value (the enum is `{backend, frontend}`, both web-bound) nor an
   api/data/ui planner for a terminal endpoint. If left unratified, `/axb-dsl-refine` must either
   invent an out-of-enum subpath (e.g. `features/cli/**`) or mis-file CLI features under `backend`.
   **Proposed resolution (for the owner):** extend `InterfaceKind` with a CLI value
   (`cli` → `features/cli/**`), **or** declare `/axb-dsl-refine` the CLI end's planner-of-record.
   *(Consolidates grill Q1 + Q3 + Q6.)* **Tracked upstream:**
   [`gosharplite/aixbdd-tmg#1`](https://github.com/gosharplite/aixbdd-tmg/issues/1).
2. ~~**PM-owned acceptance gap #1 — `-d` on a broken/unresolved setup.**~~ **RESOLVED** (session 5, PM
   role): `version-and-setup-diagnostic.feature` now carries the `-d`-unresolved Example (plain +
   `--json`) with the dedicated non-zero "diagnostic: unresolved" exit, and `spec.md` adds the matching
   edge case. *(Was grill Q5; closed as PM-1.)*
3. ~~**PM-owned acceptance gap #2 — default-path discovery.**~~ **RESOLVED** (session 5, PM role):
   `starting-with-a-configuration.feature` now carries the positive no-`-c` + `MODE≠butler`
   found-default Example. *(Was grill Q5; closed as PM-2.)*

**Only blocker 1 (the cross-repo `aixbdd-tmg` decision) remains** — the CLI-end slice is gated on that
alone. The persistence half (single wave → `/axb-data-plan`) is **not** gated.
