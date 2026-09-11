# usage module DSL

Module-specific rows for the `usage` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> Exit-code contract (**pinned**, round 002 / FR-005): usage error = `2`, distinct from success (0)
> and from configuration (3), environment (4), and diagnostic (5).
>
> `--json` (**removed**, round 002): the `--json` flag no longer exists, so any use of it — alone or
> with `-d` — is an unrecognized-flag usage error (never silently ignored).

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"` | `config_path`: string; path relative to the runtime home. `flag`: string; an unrecognized command-line flag. | 不支援 | `旗標順序`: the unrecognized flag may appear with `-c`; argument parsing precedes any resolution. | `怎麼做`: run `tellme -c $TELL_ME_HOME/{config_path} {flag}`. `權威狀態落地`: parsing runs before resolution; the unrecognized flag is a usage error even when a valid configuration is present. `回寫`: captured exit code, stdout, stderr. |
| `the operator runs tellme's diagnostic with "--json"` | 無 | 不支援 | `旗標`: the run is invoked with `-d --json`. | `怎麼做`: run `tellme -d --json`. `權威狀態落地`: parsing runs before the diagnostic dispatch; `--json` is not a flag, so the run is a usage error even though `-d` is present. `回寫`: captured exit code, stdout, stderr. |
| `the operator starts tellme with "--version" and the unrecognized flag "--json"` | 無 | 不支援 | `旗標`: the run is invoked with `--version --json`. | `怎麼做`: run `tellme --version --json`. `權威狀態落地`: parsing runs before the version path; `--json` is not a flag, so the run is a usage error even though `--version` is present. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme exits with the usage error code` | 無 | 不支援 | `碼值`: `2` (**pinned**, round 002 / FR-005) — a non-zero code distinct from success (0) and from every other error class (configuration 3, environment 4, diagnostic 5). | `必查`: `呈現結果`: the exit code **equals the pinned usage error code `2`**. `不該發生`: it must not collapse to the success code or to any other error class. |
