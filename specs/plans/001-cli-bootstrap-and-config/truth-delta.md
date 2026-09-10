# Truth Delta: 001-cli-bootstrap-and-config

**Plan Package**: `specs/plans/001-cli-bootstrap-and-config`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` | Created the initial techstack truth: Go 1.26, module `github.com/gosharplite/tellme`, layout `cmd/tellme/` + `internal/…`; `spf13/pflag` for CLI flags; `gopkg.in/yaml.v3` + a hand-written resolver for config; `godog` as the CLI BDD runner; E2E black-box test strategy; `go build` `-ldflags -X` version injection; `gofmt`/`go vet`/`make` tooling. Explicitly excludes cobra, viper, golangci-lint, testify, provider SDKs, TUI, MCP, SQLite, and web frontend. | First plan package (starting project): `spec.md` fixed Go but deferred the toolchain/test stack to this phase; three AIxBDD must-ask decisions locked — system ends = one CLI end, BDD techstack = godog, test strategy = E2E (black-box CLI). |
| MODIFY | `specs/truth/techstack.md` | Corrected the techstack **representation** after the grill round ([issue #1](https://github.com/gosharplite/tellme/issues/1), verdict *proceed with changes*): (i) layout row justified by separation of concerns / E2E-observable FR mapping, **not** unit-testability; (ii) **`internal/version` dropped** — single version source is `main.version`; (iii) `godog` labelled **adopted this round; first instantiated in `/axb-dsl-refine`** via an explicit "Adopted, Not Yet Instantiated" bucket (executed = interface Gherkin; acceptance = carried); (iv) added a **no-network verification** row (`unshare -n` sandbox + build-graph capability guard) proving SC-004 as a host-harness assertion; (v) added **`staticcheck`** to the static gate; (vi) version-injection row + **`VERSION=0.0.0-harness` sentinel** test row (single target `main.version`); (vii) deferral reasons for `golangci-lint`/`govulncheck` changed from install-cost to policy-artifact; (viii) title de-round-scoped. **Deliberately NOT placed in truth:** the `-d` diagnostic exit-code contract and the resolver algorithm (behaviour, not technology). | Grill round 001 review gate: the core stack choices held, but nearly every *representation* of them overstated or mis-scoped. Recorded as `MODIFY` (not a silent edit) per the round's outcome. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/contracts/**` | To be recorded by `/axb-api-plan` | CLI has a single end — no OpenAPI/HTTP surface; expected `NOOP`, to be confirmed by the owner. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/data/**` | **PM ruling (clarify Q2 → Option 1): the "no `data/**` truth" scope lock is RELEASED — a minimal data truth is owed.** `/axb-data-plan` (delegated by `/axb-system-analysis`) records an expected `ADD` covering (a) the config input contract (`configs/<mode>.yaml`: `MODE`/`PERSON`/`SELECTED_PROVIDER`/`PROVIDERS` + `TELL_ME_*` precedence) and (b) the workspace/state lifecycle (`output/<mode>/` created-on-first-run, reused, must-be-a-directory, idempotent). | The grill found the earlier "expected NOOP" an unexamined premise; the AIxBDD CLI rule (`aixbdd-tmg` README line 37) triggers `/axb-data-plan` for a CLI that manages persistent configuration and system-owned state, and the benchmark (`tell-me-go`) persists and models both. PM ruled the lock released; authoring belongs to the `/axb-data-plan` owner (`truth-single-owner`). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| PENDING | `specs/truth/features/**` | To be recorded by `/axb-dsl-refine` | The CLI contract (config resolution, workspace init, version, **`-d` diagnostics**) is decomposed into executable interface Gherkin + DSL here. Per grill round #1: the **executed** artifact is this interface Gherkin (acceptance journeys are **carried**, not executed); the runner/step-def package is pinned at `tests/e2e/`, so dsl-refine fixes the concrete module subpaths and the `And tellme performs no network access` step's harness-mediated `StepDef 實作語意`. **⟦clarify Q1⟧** the ratified `-d` contract (always report; dedicated non-zero "diagnostic: unresolved" code; `FR-014` codes bind the boot path only) must be carried as an interface feature. |
