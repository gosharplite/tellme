# diagnostics module DSL

Module-specific rows for the `diagnostics` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> **Round 074:** the `the operator runs tellme with "{flag}"` row (used for `--version` and now `-v`)
> lives at the **interface root** ([`../dsl.md`](../dsl.md)) — it is shared with the `usage` module
> (`-h`/`--help`), so it is a cross-module row.

> `tellme performs no network access` (now a **cross-module** row in [`../dsl.md`](../dsl.md)) is a
> host-harness assertion: the offline paths leave a recording sink untouched (zero connections) and
> complete identically under blocked egress. Round 004 **retired** the whole-binary capability guard
> ("no `net/http` in the closure") because the prompt-bearing chat path legitimately links `net/http`
> — see the root DSL row.
>
> Path normalization convention (**pinned**, grill #5 fix Q5):
> - `{home}` is the operator-facing name standing for the actual runtime home (`TELL_ME_HOME`).
> - `{workspace_path}` is an operator-facing path expressed with `{home}` as its root (e.g. `"ait-tmg/output/butler"`).
>
> Diagnostic output contract (**round 002**): the diagnostic has **one plain-text form** only — there
> is **no** `--json` / machine-readable mode. The `--json` flag was **removed**; passing it (alone or
> with `-d`) is an **unrecognized-flag usage error**, handled by the `usage` module.

## Given

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `a configuration "{config_path}" that does not resolve` | `config_path`: string; path relative to the runtime home. | Not supported | `Failure reason`: defaults to a missing or malformed file so the diagnostic reports unresolved. | `How`: arrange a configuration at `{home}/{config_path}` that does not reach a ready state. `State landing`: setup resolution is not ready. `Write-back`: the arranged (unusable) configuration. |

## When

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `the operator runs tellme's diagnostic` | none | Not supported | `Flags`: the run is invoked with `-d` (the reporting path; `--json` has been removed). | `How`: run `tellme -d` under the current environment. `State landing`: the diagnostic report is produced regardless of the resolution outcome. `Write-back`: captured exit code, stdout, stderr. |

## Then

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `tellme prints the build version` | none | Not supported | none | `Must-check`: `Presented Result`: stdout contains the build version string (the E2E build uses a distinctive sentinel so a missed injection fails). |
| `tellme reports the configuration resolved` | none | Not supported | none | `Must-check`: `Presented Result`: stdout reports the configuration resolved. `Authoritative State`: the resolution matches a ready configuration + effective provider. |
| `tellme reports the runtime home resolved to "{home}"` | `home`: string; the operator-facing runtime home name. | Not supported | none | `Must-check`: `Presented Result`: stdout reports the runtime home resolved to `{home}` (the arranged `TELL_ME_HOME`). |
| `tellme reports the session workspace resolved to "{workspace_path}"` | `workspace_path`: string; the expected operator-facing path reported on stdout (e.g. "ait-tmg/output/butler"). | Not supported | none | `Must-check`: `Presented Result`: stdout reports the session workspace resolved to `{workspace_path}`. |
| `tellme reports the configuration did not resolve` | none | Not supported | none | `Must-check`: `Presented Result`: stdout reports the configuration did not resolve. `Authoritative State`: the reported status matches the arranged unresolved setup. |
| `tellme reports the reason the configuration did not resolve` | none | Not supported | none | `Must-check`: `Presented Result`: stdout names the unresolved category (one of `config-missing`, `config-invalid`, `provider-mismatch`, `home-unset`, `home-unusable`) consistent with the arranged setup. |
| `tellme exits with the diagnostic error code` | none | Not supported | `Code`: `5` (**pinned**, round 002 / FR-005) — a dedicated non-zero code distinct from success (0) and from every other error class (usage 2, configuration 3, environment 4). | `Must-check`: `Presented Result`: the report was produced and the exit code **equals the pinned diagnostic code `5`**. `Should Not Happen`: it must not collapse to the success code or to any other error class. |
