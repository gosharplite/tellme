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

2. `Configuration & workspace persistence interface`
   - Endpoint type: `local-file / state endpoint`
   - Primary interface: the **config input contract** (`$TELL_ME_HOME/configs/<mode>.yaml`: `MODE`,
     `PERSON`, `SELECTED_PROVIDER`, `PROVIDERS` + `TELL_ME_*` env precedence) and the **workspace/state
     lifecycle** (`$TELL_ME_HOME/output/<mode>/` — created on first run, reused, must be a directory,
     idempotent).
   - Requirement evidence: `FR-002`/`003`/`005`/`006`/`007`/`008`/`009`/`015`, `NFR-001`/`NFR-002`;
     `spec.md` Assumptions.

> **Scope notes.** `spec.md` originally locked "no `data/**` truth this round"; that lock was
> **released** (clarify Q2 → Option 1), so interface 2 now delegates a **minimal** data truth.
> There is **no API/HTTP interface** — a CLI has a single end, so `/axb-api-plan` is **`NOOP`** —
> and **no UI interface** — `/axb-ui-plan` is **skipped** (CLI-streamlined workflow).

### Analysis Wave schedule

#### Wave 1

- Parallel-analyzed interfaces:
  - `Configuration & workspace persistence interface`
- Analysis focus:
  - Model the **minimal data truth**: the config input contract (keys, types, `TELL_ME_*` precedence)
    and the workspace/state lifecycle (`output/<mode>/`), in `specs/truth/data/**`.
- Scheduling rationale: the persisted-state model is the shape the executable CLI contract references,
  so it is settled first. This is the only **planner delegation** performed by `/axb-system-analysis`
  for this round (`/axb-data-plan`).

#### Wave 2

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - Produce the **executable CLI contract** — interface Gherkin + DSL for the config-resolution order
    (the 6-step contract, `research.md` Decision 3), workspace initialization, `--version`, and the
    ratified `-d` reporting contract (always report; dedicated non-zero "unresolved" code; `FR-014`
    codes bind the boot path only), driven E2E against the built binary.
- Scheduling rationale: the CLI contract's file / stream / workspace behavior is expressed against the
  model settled in Wave 1, so it follows. This wave is produced by the **next pipeline phase
  `/axb-dsl-refine`** (the CLI contract), not by a `/axb-system-analysis` planner.

### Delegation order

1. **`/axb-data-plan`** — Wave 1 (`Configuration & workspace persistence interface`) → minimal
   `specs/truth/data/**`.
2. **`/axb-dsl-refine`** *(next phase)* — Wave 2 (`CLI end`) → `specs/truth/features/**` + `dsl.md`.

Not delegated: `/axb-ui-plan` (skipped — no UI) and `/axb-api-plan` (`NOOP` — single CLI end, no
OpenAPI surface).
