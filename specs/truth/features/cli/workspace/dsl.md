# workspace module DSL

Module-specific rows for the `workspace` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> Path normalization convention (**pinned**, grill #5 fix Q5):
> - `{home}` is the operator-facing name standing for the actual runtime home (`TELL_ME_HOME`).
> - `{workspace_path}` is an operator-facing path expressed with `{home}` as its root (e.g. `"ait-tmg/output/butler"`). Stepdefs resolve the leading `{home}` name segment to the actual `TELL_ME_HOME` directory path before performing filesystem assertions.
> - Contrast with `{config_path}` (e.g. `"configs/butler.yaml"`), which is a home-relative input path resolved under `TELL_ME_HOME`.

## Given

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `no session workspace exists under "{home}"` | `home`: string; the operator-facing runtime home name. | Not supported | `Actual path`: the arranged `TELL_ME_HOME` temp directory carries no `output/` yet. | `How`: ensure the arranged runtime home has no `output/<mode>/` directory. `State landing`: no session workspace exists. `Write-back`: none. |
| `the session workspace "{workspace_path}" already exists` | `workspace_path`: string; operator-facing path rooted in `{home}` (e.g. "ait-tmg/output/butler"). | Not supported | none | `How`: resolve `{home}` to the arranged `TELL_ME_HOME` and create the directory (e.g. `<runtime_home>/output/<mode>`). `State landing`: the workspace directory exists. `Write-back`: the directory. |
| `the workspace "{workspace_path}" already holds a file "{file_name}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. `file_name`: string; the name of a file already stored in the workspace. | Not supported | `Content`: a non-empty sentinel body. | `How`: write a sentinel file `{file_name}` into the resolved workspace directory `<runtime_home>/output/<mode>`. `State landing`: the workspace holds pre-existing content. `Write-back`: the file. |
| `the configuration "{config_path}" declares the mode "{mode}"` | `config_path`: string; path relative to the runtime home. `mode`: string; the file's `MODE` value. | Not supported | `SELECTED_PROVIDER`: defaults to "deepseek-flash"; `PROVIDERS`: a registry containing the selected provider. | `How`: write a **resolvable** YAML at `{home}/{config_path}` whose `MODE` is `{mode}` and whose `PROVIDERS` registry contains the selected provider — so config load + validation succeed. `State landing`: the file is loadable and valid, declaring `{mode}`. `Write-back`: the configuration file. |
| `the mode override is "{mode}"` | `mode`: string; the mode supplied by the environment override. | Not supported | `Source`: arranged via `TELL_ME_MODE`. | `How`: set `TELL_ME_MODE` to `{mode}`. `State landing`: the effective mode is taken from the environment first. `Write-back`: none. |
| `the workspace path "{workspace_path}" already exists as a regular file` | `workspace_path`: string; operator-facing path rooted in `{home}`. | Not supported | none | `How`: create a regular file (not a directory) at the resolved workspace path `<runtime_home>/output/<mode>`. `State landing`: the workspace path is a file. `Write-back`: the regular file. |

## Then

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `tellme creates the session workspace "{workspace_path}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. | Not supported | none | `Must-check`: `Authoritative State`: the resolved workspace directory `<runtime_home>/output/<mode>` now exists and is a directory. `Presented Result`: the run prepared the workspace for the effective mode. |
| `tellme reports the session workspace "{workspace_path}"` | `workspace_path`: string; the expected operator-facing path reported on stdout (e.g. "ait-tmg/output/butler"). | Not supported | none | `Must-check`: `Presented Result`: stdout reports the resolved workspace path `{workspace_path}`. `Authoritative State`: the reported path matches the directory that exists. |
| `tellme reuses the session workspace "{workspace_path}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. | Not supported | none | `Must-check`: `Authoritative State`: the workspace directory `<runtime_home>/output/<mode>` is the same directory as before (not re-created or replaced). `Presented Result`: the run reported the reused workspace. |
| `the workspace "{workspace_path}" still holds the file "{file_name}"` | `workspace_path`: string; operator-facing path rooted in `{home}`. `file_name`: string; the preserved file name. | Not supported | none | `Must-check`: `Authoritative State`: `{file_name}` still exists with its sentinel body in the resolved workspace directory `<runtime_home>/output/<mode>`. `Should Not Happen`: the reuse run must not delete or overwrite workspace state. |
| `tellme exits with the environment error code` | none | Not supported | `Code`: `4` (**pinned**, round 002 / FR-005) — a non-zero code distinct from success (0) and from every other error class (usage 2, configuration 3, diagnostic 5). | `Must-check`: `Presented Result`: the exit code **equals the pinned environment error code `4`**. `Should Not Happen`: it must not collapse to the success code or to any other error class. |
