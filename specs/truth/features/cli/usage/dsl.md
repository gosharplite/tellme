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

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `the operator starts tellme pointing at the configuration "{config_path}" with the unrecognized flag "{flag}"` | `config_path`: string; path relative to the runtime home. `flag`: string; an unrecognized command-line flag. | Not supported | `Flag order`: the unrecognized flag may appear with `-c`; argument parsing precedes any resolution. | `How`: run `tellme -c $TELL_ME_HOME/{config_path} {flag}`. `State landing`: parsing runs before resolution; the unrecognized flag is a usage error even when a valid configuration is present. `Write-back`: captured exit code, stdout, stderr. |
| `the operator runs tellme's diagnostic with "--json"` | none | Not supported | `Flags`: the run is invoked with `-d --json`. | `How`: run `tellme -d --json`. `State landing`: parsing runs before the diagnostic dispatch; `--json` is not a flag, so the run is a usage error even though `-d` is present. `Write-back`: captured exit code, stdout, stderr. |
| `the operator starts tellme with "--version" and the unrecognized flag "--json"` | none | Not supported | `Flags`: the run is invoked with `--version --json`. | `How`: run `tellme --version --json`. `State landing`: parsing runs before the version path; `--json` is not a flag, so the run is a usage error even though `--version` is present. `Write-back`: captured exit code, stdout, stderr. |

## Then

| DSL Sentence | Gherkin Params | Data Table Params | Default Params | StepDef Implementation Semantics |
| --- | --- | --- | --- | --- |
| `tellme exits with the usage error code` | none | Not supported | `Code`: `2` (**pinned**, round 002 / FR-005) — a non-zero code distinct from success (0) and from every other error class (configuration 3, environment 4, diagnostic 5). | `Must-check`: `Presented Result`: the exit code **equals the pinned usage error code `2`**. `Should Not Happen`: it must not collapse to the success code or to any other error class. |
| `tellme prints its flag list` | none | Not supported | `Channel`: standard output. `Source`: the round-074 help path. | `Must-check`: `Presented Result`: the captured standard output carries the flag list — the new `-h`/`--help` and `-v`/`--version` flags appear, alongside the existing `-c`, `-d`, `-i`, `-l`, `-r`, `-t`, `--new`, `--tool-usage`. `Should Not Happen`: the help must not be empty, and it must not omit the new `-h`/`--help` or `-v`/`--version` flags. |
| `the help is reported as a success` | none | Not supported | `Channel`: standard output + standard error. `Source`: the round-074 help path. | `Must-check`: `Presented Result`: the exit code equals the success code (0) and the captured **standard error is empty** — no `tellme: …` class phrase. `Should Not Happen`: a help request must not be reported as a failure (no `tellme: ` line, no non-zero exit code). (The "went to standard **output**" half is carried by the paired `tellme prints its flag list` Then — the two are asserted together.) |
