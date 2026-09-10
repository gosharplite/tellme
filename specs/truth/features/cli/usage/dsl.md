# usage module DSL

Module-specific rows for the `usage` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"` | `config_path`: string; path relative to the runtime home. `flag`: string; an unrecognized command-line flag. | 不支援 | `旗標順序`: the unrecognized flag may appear with `-c`; argument parsing precedes any resolution. | `怎麼做`: run `tellme -c $TELL_ME_HOME/{config_path} {flag}`. `權威狀態落地`: parsing runs before resolution; the unrecognized flag is a usage error even when a valid configuration is present. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme exits with the usage error code` | 無 | 不支援 | `碼值`: a dedicated, deterministic non-zero code distinct from the success code and from every other error class (configuration, environment, diagnostic). | `必查`: `呈現結果`: the exit code is non-zero and equals **its** code dedicated to usage errors. `不該發生`: it must not collapse to the success code or to any other error class. |
