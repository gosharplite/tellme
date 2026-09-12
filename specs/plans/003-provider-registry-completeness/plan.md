# System Analysis Plan — round 003 (`003-provider-registry-completeness`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/003-provider-registry-completeness/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── provider-configuration-loading.feature
│       └── provider-validation-contract.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` or `data/**` truth artifacts in this round — see the `NOOP` notes for `/axb-api-plan` and `/axb-data-plan` below.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # entrypoint; `var version` is the single `-X main.version` target

internal/
├── cli/                           # flag parsing (pflag), exit-code classification, -d dispatch
│   ├── cli.go
│   ├── exitcode.go                # exit-code values (0/2/3/4/5)
│   └── *_test.go
├── config/                        # YAML load + validate + effective-value resolver
│   ├── config.go                  # expanded: Provider struct with core request fields; active provider validation
│   ├── expand.go                  # NEW: stdlib regex-based ${VAR} and ${VAR:-default} expansion engine
│   ├── expand_test.go             # NEW: unit tests for variable expansion
│   └── config_test.go             # updated: unit tests for complete provider schema, expansion, and validation
└── home/                          # TELL_ME_HOME resolution + `output/<mode>/` workspace lifecycle
    └── *_test.go

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
Makefile                           # fmt, tidy, build, test, lint, vulncheck, verify
.golangci.yml                      # lint policy
go.mod / go.sum
```

**Structure Decision**: Retains the clean 3-package architecture (`internal/{cli,config,home}`) established in Round 001. Round 003 modifies `internal/config/config.go` to expand the `Provider` struct with the core request fields, adds `internal/config/expand.go` for deterministic `${VAR}` and `${VAR:-default}` string expansion, and adds comprehensive table-driven unit tests in `internal/config/expand_test.go` and `internal/config/config_test.go`. No new top-level package or external dependencies are introduced.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** system interfaces (consistent with rounds 001 and 002):

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: command flags (`-c`/`--config`, `-d`, `--version`); stdout/stderr console reporting; exit codes (`0` success, `3` configuration error); and the frozen stderr class phrase (`tellme: the provider configuration is invalid`).
   - Requirement evidence: `FR-007`, `FR-008`, `FR-009` (US3 — deterministic validation & failure reporting); acceptance feature `provider-validation-contract.feature`.
   - Planner: **none** — terminal endpoints do not have an analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Configuration & workspace persistence interface`
   - Endpoint type: `local-file / state endpoint`
   - Primary interface: Configuration input file (`$TELL_ME_HOME/configs/<mode>.yaml` or explicit `-c`), `PROVIDERS` schema (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`), environment variable expansion (`${VAR}` and `${VAR:-default}`), and active provider validation.
   - Requirement evidence: `FR-001`..`FR-006` (US1, US2); acceptance feature `provider-configuration-loading.feature`.
   - Planner: **`/axb-data-plan`** = **`NOOP`** (configuration is loaded input; no persistent state entities or database models are introduced).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - `/axb-data-plan` = **`NOOP`** (no persisted database model or state store).

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Configuration & workspace persistence interface`
- Analysis focus:
  - **Persistence** → **`NOOP`**: The configuration is read-only input loaded at boot time. No database schema or state persistence changes.
  - **CLI end** → Handoff to contract owner **`/axb-dsl-refine`** to update `specs/truth/features/cli/configuration/**` (adding scenarios for full provider schema, optional field omission, `${VAR}` expansion, and failure handling) and `specs/truth/features/cli/dsl.md` (adding the frozen class phrase `the provider configuration is invalid` to the failure vocabulary).
- Scheduling rationale: The two interfaces are information-independent and share a single wave per `Wave依賴排序與平行分組判準.md` Rule 2.

---

### Delegation order

1. **`/axb-data-plan`** — Wave 1 (`Configuration & workspace persistence interface`) → **`NOOP`** (no data model).
2. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → update `specs/truth/features/cli/configuration/**` and interface root `specs/truth/features/cli/dsl.md`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` (`NOOP` — no OpenAPI contract).

---

### Gating blockers

*(none — all requirement questions were settled during Clarify Round 1; no cross-repo or architectural blockers remain).*
