# diagnostics module DSL

Module-specific rows for the `diagnostics` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

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

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `a configuration "{config_path}" that does not resolve` | `config_path`: string; path relative to the runtime home. | 不支援 | `失效原因`: defaults to a missing or malformed file so the diagnostic reports unresolved. | `怎麼做`: arrange a configuration at `{home}/{config_path}` that does not reach a ready state. `權威狀態落地`: setup resolution is not ready. `回寫`: the arranged (unusable) configuration. |

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator runs tellme with "--version"` | 無 | 不支援 | `旗標`: the run is invoked with `--version`. | `怎麼做`: run `tellme --version`. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr. |
| `the operator runs tellme's diagnostic` | 無 | 不支援 | `旗標`: the run is invoked with `-d` (the reporting path; `--json` has been removed). | `怎麼做`: run `tellme -d` under the current environment. `權威狀態落地`: the diagnostic report is produced regardless of the resolution outcome. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme prints the build version` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout contains the build version string (the E2E build uses a distinctive sentinel so a missed injection fails). |
| `tellme reports the configuration resolved` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the configuration resolved. `權威狀態`: the resolution matches a ready configuration + effective provider. |
| `tellme reports the runtime home resolved to "{home}"` | `home`: string; the operator-facing runtime home name. | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the runtime home resolved to `{home}` (the arranged `TELL_ME_HOME`). |
| `tellme reports the session workspace resolved to "{workspace_path}"` | `workspace_path`: string; the expected operator-facing path reported on stdout (e.g. "ait-tmg/output/butler"). | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the session workspace resolved to `{workspace_path}`. |
| `tellme reports the configuration did not resolve` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the configuration did not resolve. `權威狀態`: the reported status matches the arranged unresolved setup. |
| `tellme reports the reason the configuration did not resolve` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout names the unresolved category (one of `config-missing`, `config-invalid`, `provider-mismatch`, `home-unset`, `home-unusable`) consistent with the arranged setup. |
| `tellme exits with the diagnostic error code` | 無 | 不支援 | `碼值`: `5` (**pinned**, round 002 / FR-005) — a dedicated non-zero code distinct from success (0) and from every other error class (usage 2, configuration 3, environment 4). | `必查`: `呈現結果`: the report was produced and the exit code **equals the pinned diagnostic code `5`**. `不該發生`: it must not collapse to the success code or to any other error class. |
