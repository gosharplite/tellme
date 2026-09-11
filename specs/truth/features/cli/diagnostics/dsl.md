# diagnostics module DSL

Module-specific rows for the `diagnostics` module. Cross-module rows live in
[`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> `tellme performs no network access` is a host-harness assertion, not a black-box observation
> (black-box cannot witness an absence). It is mediated by the E2E harness — a no-egress sandbox /
> hostile-env differential witness plus the build-graph **capability** guard — per the round's
> technical-research Decision 5.
>
> Capability-guard semantics (**corrected**, round-001 implementation): the guard checks **network
> capability**, not bare package presence — no `net/http` in the binary's dependency closure, and no
> dialing/listening symbol in the linked binary (`go tool nm`). Bare `net` / `net/netip` may be
> linked transitively by `spf13/pflag`'s IP-flag parsing without performing any I/O, so they are
> **not** indicators.
>
> `--json` output contract (**pinned**, grill #4 fix A2): a single JSON object with keys —
> `status` (string, `"resolved"` or `"unresolved"`), `reason` (string, present only when unresolved:
> one of `config-missing`, `config-invalid`, `provider-mismatch`, `home-unset`, `home-unusable`),
> `runtime_home` (string, present when the home resolved), `session_workspace` (string, present when
> the workspace resolved). The plain report and `--json` must agree on `status` and `reason`.
>
> Path normalization convention (**pinned**, grill #5 fix Q5):
> - `{home}` is the operator-facing name standing for the actual runtime home (`TELL_ME_HOME`).
> - `{workspace_path}` is an operator-facing path expressed with `{home}` as its root (e.g. `"ait-tmg/output/butler"`).

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `a configuration "{config_path}" that does not resolve` | `config_path`: string; path relative to the runtime home. | 不支援 | `失效原因`: defaults to a missing or malformed file so the diagnostic reports unresolved. | `怎麼做`: arrange a configuration at `{home}/{config_path}` that does not reach a ready state. `權威狀態落地`: setup resolution is not ready. `回寫`: the arranged (unusable) configuration. |

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator runs tellme with "--version"` | 無 | 不支援 | `旗標`: the run is invoked with `--version`. | `怎麼做`: run `tellme --version`. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr. |
| `the operator runs tellme's diagnostic` | 無 | 不支援 | `旗標`: the run is invoked with `-d` (the reporting path). | `怎麼做`: run `tellme -d` under the current environment. `權威狀態落地`: the diagnostic report is produced regardless of the resolution outcome. `回寫`: captured exit code, stdout, stderr. |
| `the operator runs tellme's diagnostic with "--json"` | 無 | 不支援 | `旗標`: the run is invoked with `-d --json`. | `怎麼做`: run `tellme -d --json` under the current environment. `權威狀態落地`: the structured report is produced regardless of the resolution outcome. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme prints the build version` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout contains the build version string (the E2E build uses a distinctive sentinel so a missed injection fails). |
| `tellme reports the configuration resolved` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the configuration resolved. `權威狀態`: the resolution matches a ready configuration + effective provider. |
| `tellme reports the runtime home resolved to "{home}"` | `home`: string; the operator-facing runtime home name. | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the runtime home resolved to `{home}` (the arranged `TELL_ME_HOME`). |
| `tellme reports the session workspace resolved to "{workspace_path}"` | `workspace_path`: string; the expected operator-facing path reported on stdout (e.g. "ait-tmg/output/butler"). | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the session workspace resolved to `{workspace_path}`. |
| `tellme reports the configuration did not resolve` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout reports the configuration did not resolve. `權威狀態`: the reported status matches the arranged unresolved setup. |
| `tellme reports the reason the configuration did not resolve` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout names the unresolved category (one of `config-missing`, `config-invalid`, `provider-mismatch`, `home-unset`, `home-unusable`) consistent with the arranged setup. |
| `tellme emits the resolution status as structured output` | 無 | 不支援 | `格式`: the pinned `--json` object (`status: "resolved"`, plus `runtime_home` and `session_workspace`). | `必查`: `呈現結果`: stdout parses as the pinned JSON object; assert `status == "resolved"` and that `runtime_home` / `session_workspace` equal the values the plain form reported. |
| `tellme emits the unresolved status as structured output` | 無 | 不支援 | `格式`: the pinned `--json` object (`status: "unresolved"`, plus `reason`). | `必查`: `呈現結果`: stdout parses as the pinned JSON object; assert `status == "unresolved"` and that `reason` equals the unresolved category the plain form reported. |
| `tellme performs no network access` | 無 | 不支援 | 無 | `必查`: harness-mediated — `呈現結果` / `權威狀態`: with egress blocked the diagnostic's exit code and output are unchanged (differential witness), and the binary has **no network capability** — no `net/http` in its dependency closure and no dialing/listening symbol in the linked binary (`go tool nm` capability guard). Bare `net` / `net/netip` may be linked transitively by `spf13/pflag` IP-flag parsing without implying network I/O and are **not** indicators. This is a differential + capability-absence witness, not a black-box observation. |
| `tellme exits with the diagnostic error code` | 無 | 不支援 | `碼值`: a dedicated, deterministic non-zero code distinct from the success code and from every other error class (usage, configuration, environment). | `必查`: `呈現結果`: the report was produced and the exit code is non-zero and equals **its** code dedicated to the diagnostic "unresolved" outcome. `不該發生`: it must not collapse to the success code or to any other error class. |
