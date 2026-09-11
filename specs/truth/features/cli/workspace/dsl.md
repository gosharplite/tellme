# workspace module DSL

Module-specific rows for the `workspace` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> Path normalization convention (**pinned**, grill #5 fix Q5):
> - `{home}` is the operator-facing name standing for the actual runtime home (`TELL_ME_HOME`).
> - `{workspace_path}` is an operator-facing path expressed with `{home}` as its root (e.g. `"ait-tmg/output/butler"`). Stepdefs resolve the leading `{home}` name segment to the actual `TELL_ME_HOME` directory path before performing filesystem assertions.
> - Contrast with `{config_path}` (e.g. `"configs/butler.yaml"`), which is a home-relative input path resolved under `TELL_ME_HOME`.

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `no session workspace exists under "{home}"` | `home`: string; the operator-facing runtime home name. | 不支援 | `實際路徑`: the arranged `TELL_ME_HOME` temp directory carries no `output/` yet. | `怎麼做`: ensure the arranged runtime home has no `output/<mode>/` directory. `權威狀態落地`: no session workspace exists. `回寫`: none. |
| `the session workspace "{workspace_path}" already exists` | `workspace_path`: string; operator-facing path rooted in `{home}` (e.g. "ait-tmg/output/butler"). | 不支援 | 無 | `怎麼做`: resolve `{home}` to the arranged `TELL_ME_HOME` and create the directory (e.g. `<runtime_home>/output/<mode>`). `權威狀態落地`: the workspace directory exists. `回寫`: the directory. |
| `the workspace "{workspace_path}" already holds a file "{file_name}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. `file_name`: string; the name of a file already stored in the workspace. | 不支援 | `內容`: a non-empty sentinel body. | `怎麼做`: write a sentinel file `{file_name}` into the resolved workspace directory `<runtime_home>/output/<mode>`. `權威狀態落地`: the workspace holds pre-existing content. `回寫`: the file. |
| `the configuration "{config_path}" declares the mode "{mode}"` | `config_path`: string; path relative to the runtime home. `mode`: string; the file's `MODE` value. | 不支援 | `SELECTED_PROVIDER`: defaults to "deepseek-flash"; `PROVIDERS`: a registry containing the selected provider. | `怎麼做`: write a **resolvable** YAML at `{home}/{config_path}` whose `MODE` is `{mode}` and whose `PROVIDERS` registry contains the selected provider — so config load + validation succeed. `權威狀態落地`: the file is loadable and valid, declaring `{mode}`. `回寫`: the configuration file. |
| `the mode override is "{mode}"` | `mode`: string; the mode supplied by the environment override. | 不支援 | `來源`: arranged via `TELL_ME_MODE`. | `怎麼做`: set `TELL_ME_MODE` to `{mode}`. `權威狀態落地`: the effective mode is taken from the environment first. `回寫`: none. |
| `the runtime home is not set` | 無 | 不支援 | 無 | `怎麼做`: ensure `TELL_ME_HOME` is unset in the process environment. `權威狀態落地`: no runtime home is resolvable. `回寫`: none. |
| `the workspace path "{workspace_path}" already exists as a regular file` | `workspace_path`: string; operator-facing path rooted in `{home}`. | 不支援 | 無 | `怎麼做`: create a regular file (not a directory) at the resolved workspace path `<runtime_home>/output/<mode>`. `權威狀態落地`: the workspace path is a file. `回寫`: the regular file. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme creates the session workspace "{workspace_path}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. | 不支援 | 無 | `必查`: `權威狀態`: the resolved workspace directory `<runtime_home>/output/<mode>` now exists and is a directory. `呈現結果`: the run prepared the workspace for the effective mode. |
| `tellme reports the session workspace "{workspace_path}"` | `workspace_path`: string; the expected operator-facing path reported on stdout (e.g. "ait-tmg/output/butler"). | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the resolved workspace path `{workspace_path}`. `權威狀態`: the reported path matches the directory that exists. |
| `tellme reuses the session workspace "{workspace_path}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. | 不支援 | 無 | `必查`: `權威狀態`: the workspace directory `<runtime_home>/output/<mode>` is the same directory as before (not re-created or replaced). `呈現結果`: the run reported the reused workspace. |
| `the workspace "{workspace_path}" still holds the file "{file_name}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. `file_name`: string; the preserved file name. | 不支援 | 無 | `必查`: `權威狀態`: `{file_name}` still exists with its sentinel body in the resolved workspace directory `<runtime_home>/output/<mode>`. `不該發生`: the reuse run must not delete or overwrite workspace state. |
| `tellme exits with the environment error code` | 無 | 不支援 | `碼值`: a dedicated, deterministic non-zero code distinct from the success code and from every other error class (usage, configuration, diagnostic). | `必查`: `呈現結果`: the exit code is non-zero and equals **its** code dedicated to environment errors. `不該發生`: it must not collapse to the success code or to any other error class. |
