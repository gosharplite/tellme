# CLI shared DSL (interface root)

Interface: `cli` (the operator terminal end). This file holds the **cross-module** rows — sentence
patterns used identically by two or more CLI modules. Module-specific rows live in `{module}/dsl.md`.
Reading a feature merges this root with the feature's own module `dsl.md`; every step must match
**exactly one** row.

> The `cli` end is verified **end-to-end** (black-box): the E2E suite builds the `tellme` binary once
> and invokes it as a subprocess under a controlled environment, asserting exit code, stdout, stderr,
> and filesystem effects. Steps name only operator-facing surface (flags, env vars, paths).

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator has a runnable tellme installation` | 無 | 不支援 | `binary`: the E2E suite builds the `tellme` binary once per run and passes its path (e.g. `TELLME_BIN`). | `怎麼做`: assert the built binary exists and is executable. `權威狀態落地`: a runnable `tellme` binary is available for every scenario. `不必查`: no provider is contacted. |
| `the runtime home is "{home}"` | `home`: string; the operator-facing name of the runtime home (e.g. "ait-tmg"). | 不支援 | `實際路徑`: echo case uses a fresh empty temp directory to stand for `{home}`, exposed as `TELL_ME_HOME`. | `怎麼做`: create a fresh empty temp directory and set `TELL_ME_HOME` to it. `權威狀態落地`: `TELL_ME_HOME` is set in the process environment. `回寫`: none (Arrange state only). |
| `a well-formed configuration "{config_path}"` | `config_path`: string; the configuration path relative to the runtime home. | 不支援 | `SELECTED_PROVIDER`: defaults to "deepseek-flash"; `PROVIDERS`: defaults to a single matching entry. | `怎麼做`: write a usable YAML file at `{home}/{config_path}` carrying `MODE`, `PERSON`, `SELECTED_PROVIDER`, and a `PROVIDERS` registry. `權威狀態落地`: the file exists and parses. `回寫`: the configuration file. |

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts tellme` | 無 | 不支援 | `設定來源`: no `-c` is passed, so the run seeks the default `$TELL_ME_HOME/configs/<effective mode>.yaml`. | `怎麼做`: run the built `tellme` binary with no `-c`, under the current environment (`TELL_ME_HOME` / `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER`). `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and filesystem changes. |
| `the operator starts tellme pointing at the configuration "{config_path}"` | `config_path`: string; the configuration path relative to the runtime home. | 不支援 | `旗標`: the run is invoked with `-c` pointing at `$TELL_ME_HOME/{config_path}`. | `怎麼做`: run `tellme -c $TELL_ME_HOME/{config_path}`. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and filesystem changes. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme exits successfully` | 無 | 不支援 | 無 | `必查`: `呈現結果`: the exit code equals the success code (0). `不該發生`: no non-zero exit code. |
| `tellme refuses to proceed` | 無 | 不支援 | 無 | `必查`: `呈現結果`: the exit code is non-zero (the run refused). `不該發生`: the run must not be reported as success. |
| `tellme explains on stderr that "{reason}"` | `reason`: string; the **frozen class phrase** for a failure class (verbatim phrase). | 不支援 | `片語詞彙`: the seven frozen class phrases — `the configuration could not be found`, `no configuration could be found`, `the configuration could not be parsed`, `the selected provider is not in the registry`, `the workspace path is not a directory`, `the runtime home is not usable` (one class for the *not-usable* **and** *unset* paths), `the command-line usage is invalid`. | `必查`: `呈現結果`: exactly **one** stderr line begins with `tellme: `, and that line starts with `tellme: {reason}`; any class-specific trailing detail (a path, a provider name, an OS error string) is **not part of the frozen contract**. `不該發生`: the failure reason must not be swallowed or emitted only on stdout, and no second `tellme: `-prefixed line may exist. |
