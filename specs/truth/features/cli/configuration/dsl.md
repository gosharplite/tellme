# configuration module DSL

Module-specific rows for the `configuration` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `no configuration exists at "{config_path}"` | `config_path`: string; path relative to the runtime home that must be absent. | 不支援 | 無 | `怎麼做`: ensure nothing exists at `{home}/{config_path}`. `權威狀態落地`: the path is absent. `回寫`: none. |
| `a malformed configuration "{config_path}"` | `config_path`: string; path relative to the runtime home. | 不支援 | 無 | `怎麼做`: write a file at `{home}/{config_path}` containing invalid YAML. `權威狀態落地`: the file exists but cannot be parsed. `回寫`: the malformed file. |
| `no configuration exists at the default location` | 無 | 不支援 | `預設路徑`: `$TELL_ME_HOME/configs/<effective mode>.yaml`; the effective mode's file is absent. | `怎麼做`: ensure the default configuration path for the effective mode does not exist. `權威狀態落地`: no default configuration is discoverable. `回寫`: none. |
| `the effective mode is "{mode}"` | `mode`: string; the effective mode to arrange. | 不支援 | `來源`: arranged via `TELL_ME_MODE` (environment over file). | `怎麼做`: set `TELL_ME_MODE` to `{mode}`. `權威狀態落地`: the effective mode resolves to `{mode}`. `回寫`: none. |
| `a well-formed configuration "{config_path}" whose selected provider "{provider}" is in its registry` | `config_path`: string; path relative to the runtime home. `provider`: string; a provider key that must be present in the file registry. | 不支援 | `PROVIDERS`: contains `{provider}`; `SELECTED_PROVIDER`: defaults to `{provider}`. | `怎麼做`: write a usable YAML whose `SELECTED_PROVIDER` is `{provider}` and whose `PROVIDERS` registry contains `{provider}`. `權威狀態落地`: the file's selected provider resolves to a registry entry. `回寫`: the configuration file. |
| `a well-formed configuration "{config_path}" whose provider registry is empty` | `config_path`: string; path relative to the runtime home. | 不支援 | 無 | `怎麼做`: write a usable YAML whose `PROVIDERS` registry is empty. `權威狀態落地`: the registry holds no entries. `回寫`: the configuration file. |
| `the selected provider override is "{provider}"` | `provider`: string; the provider name supplied by the environment override. | 不支援 | `來源`: arranged via `TELL_ME_SELECTED_PROVIDER`. | `怎麼做`: set `TELL_ME_SELECTED_PROVIDER` to `{provider}`. `權威狀態落地`: the effective selected provider is taken from the environment first. `回寫`: none. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme reports the configuration is ready` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the configuration is ready. `權威狀態`: the run validated the file and the effective selected provider. `不該發生`: no configuration-error exit code. |
| `tellme exits with the configuration error code` | 無 | 不支援 | `碼值`: a dedicated, deterministic non-zero code distinct from the success code and from every other error class (usage, environment, diagnostic). | `必查`: `呈現結果`: the exit code is non-zero and equals **its** code dedicated to configuration errors. `不該發生`: it must not collapse to the success code or to any other error class. |
