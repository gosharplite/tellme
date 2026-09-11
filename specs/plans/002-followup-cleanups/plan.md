# System Analysis Plan — round 002 (`002-followup-cleanups`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/002-followup-cleanups/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── single-diagnostic-output.feature
│       └── failure-reporting-contract.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` in the document structure — see the `NOOP` `/axb-api-plan` note below.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # entrypoint; `var version` is the single `-X main.version` target

internal/
├── cli/                           # flag parsing (pflag), exit-code classification, `-d` dispatch
│   ├── cli.go                     # `--json` flag removed this round
│   ├── exitcode.go                # exit-code values FROZEN this round (FR-005)
│   └── *_test.go                  # NEW: pure-helper unit tests (stdlib testing)
├── config/                        # YAML load + validate + effective-value resolver
│   └── *_test.go                  # NEW: precedence unit tests (EffectiveMode / EffectiveSelectedProvider)
└── home/                          # TELL_ME_HOME resolution + `output/<mode>/` workspace lifecycle
    └── *_test.go                  # NEW: EnsureWorkspace idempotency unit tests

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
Makefile                           # + golangci-lint (errcheck) + govulncheck gates (round 002)
.golangci.yml                      # NEW: lint policy artifact (round-002 research Decision 2)
go.mod / go.sum
```

**Structure Decision**: the CLI-only structure is unchanged from round 001 — one Go module
(`github.com/gosharplite/tellme`), entrypoint `cmd/tellme/`, non-public logic in
`internal/{cli,config,home}`, E2E Gherkin driven from `tests/e2e/`. Round 002 adds **no new package**;
it (a) removes the `--json` flag, (b) freezes the messages/exit codes (interface truth), and
(c) adds two **verification** layers — per-helper unit tests (co-located `*_test.go`, stdlib
`testing`) and the `golangci-lint` (`errcheck`) + `govulncheck` gates (`.golangci.yml` + Makefile
targets) — plus no API surface and no `data/**` model.

## Analysis Plan

### System interface inventory

This requirement yields **2** system interfaces — the same two as round 001; the round-002 cleanup does
not add, split, or remove a system interface.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: command flags (`-c`/`--config`, `-d`, `--version` — the **`--json` flag is
     removed** this round); the exit-code taxonomy (success / usage / configuration / environment /
     diagnostic-unresolved — the **numeric values are frozen** this round, FR-005); stdout / stderr;
     and the console output (readiness, resolved workspace path, build version, diagnostic report —
     the **stderr wording is frozen** this round, FR-004).
   - Requirement evidence: `FR-001`/`FR-002`/`FR-003` (US1 — remove `--json`; the plain diagnostic is
     retained) and `FR-004`/`FR-005`/`FR-006` (US2 — the frozen message/exit-code contract); the two
     round-002 acceptance features (`single-diagnostic-output`, `failure-reporting-contract`).
   - Planner: **none** — no `axb-system-analysis` planner takes a terminal endpoint. Its **contract
     owner** is **`/axb-dsl-refine`**, which produces the executable contract in the next pipeline
     phase, **carried forward at delivery** (a forward handoff — **not** a `Wave` delegation). Covered
     by `wave-covers-interfaces` via its *carried-forward-to-its-contract-owner **at delivery*** branch.

2. `Configuration & workspace persistence interface`
   - Endpoint type: `local-file / state endpoint`
   - Primary interface: the config input contract (`$TELL_ME_HOME/configs/<mode>.yaml` + `TELL_ME_*`
     env precedence) and the workspace/state lifecycle (`$TELL_ME_HOME/output/<mode>/`) — **unchanged**
     by this round.
   - Requirement evidence: `spec.md` Assumptions (no `data/**` model); round 002 introduces no new
     persisted or in-memory state.
   - Planner: **`/axb-data-plan`** = **`NOOP`** (no `data/**` model this round).

> **Scope notes.** No **API/HTTP** interface — a CLI has a single end, so `/axb-api-plan` is **`NOOP`**.
> No **UI** interface — `/axb-ui-plan` is **skipped** (CLI-streamlined workflow). The `--json` removal
> and the message/exit-code freeze are **CLI-contract behaviour** owned by `/axb-dsl-refine`
> (`specs/truth/features/cli/**`).

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Configuration & workspace persistence interface`
- Analysis focus:
  - **Persistence** → **`NOOP`**: no `data/**` model this round; round 002 changes no persisted or
    in-memory state.
  - **CLI end** → produce the **executable CLI-contract delta**: **`DELETE`** the `--json` flag together
    with its two `diagnostics` DSL rows, the pinned JSON key-schema clause, and the `--json`
    structured-output Examples (reverses round-001 FR-013); **pin** the exact operator-facing stderr
    messages (FR-004) and the numeric exit-code values (FR-005) in the DSL rows; and ensure the
    **usage** rule treats `--json` as an **unrecognized flag** (the residual of the removal).
- Scheduling rationale: unchanged from round 001 — the two interfaces are **information-independent**
  (both read the same `spec.md`/`research.md` sources; neither consumes the other's output), so per
  `Wave依賴排序與平行分組判準.md` Rule 2 they share a **single wave**.

### Delegation order

1. **`/axb-data-plan`** — Wave 1 (`Configuration & workspace persistence interface`) → **`NOOP`**
   (no `data/**` model).
2. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` (no api/data/ui
   planner applies; **not** a `Wave` delegation) → `specs/truth/features/cli/**` + `dsl.md`.

Not delegated: `/axb-ui-plan` (skipped — no UI) and `/axb-api-plan` (`NOOP` — single CLI end, no
OpenAPI surface).

### Gating blockers

*(none — this round introduces no gating blocker; the round-001 CLI-seat blocker remains resolved and
the `/axb-dsl-refine` slice is ungated.)*
