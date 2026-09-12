# chat module DSL

Module-specific rows for the `chat` module — the single-prompt reasoning turn. Cross-module rows live
in [`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> Round 004 (`004-first-reasoning-turn`): one `tellme "<prompt>"` performs **exactly one** non-streaming
> provider request and prints the answer. The provider transport is OpenAI-compatible; the request
> endpoint is the resolved provider `URL`. The E2E suite arranges all provider behaviour on an
> in-process **fake provider**, so "answers with", "error status", "unreachable", and "no usable
> answer" are all scripted on that fake (and a reachable fake doubles as the recording sink for the
> offline-path witness in the root DSL).

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `a configured provider "{provider}" whose endpoint answers with "{answer}"` | `provider`: string; the selected provider key. `answer`: string; the answer text the endpoint returns. | 不支援 | `config`: the default `$TELL_ME_HOME/configs/<effective mode>.yaml` with `SELECTED_PROVIDER` = `{provider}`; endpoint URL = the fake provider. | `怎麼做`: write a resolvable default configuration whose `SELECTED_PROVIDER` is `{provider}` and whose `PROVIDERS.{provider}.URL` points at the fake provider; script the fake to answer `{answer}`. `權威狀態落地`: the configuration is resolvable and the fake is ready to answer. `回寫`: the configuration file + the fake's scripted response. |
| `configured providers "{provider_a}" and "{provider_b}" whose endpoints answer` | `provider_a`, `provider_b`: string; two provider keys. | 不支援 | `config`: default config with both providers in the registry; each endpoint = the fake. | `怎麼做`: write a resolvable default configuration with two providers (`{provider_a}`, `{provider_b}`), both pointing at the fake (each answering a canned text). `權威狀態落地`: the registry holds both providers and both fakes are ready. `回寫`: the configuration file. |
| `a configured provider "{provider}" whose endpoint answers with an error status` | `provider`: string; the selected provider key. | 不支援 | `config`: default config with `SELECTED_PROVIDER` = `{provider}`; endpoint = the fake. | `怎麼做`: write a resolvable default configuration selecting `{provider}` whose endpoint points at the fake; script the fake to return an error status (e.g. `500`). `權威狀態落地`: the fake is scripted to answer with a non-2xx status. `回寫`: the configuration file + the fake's scripted response. |
| `a configured provider "{provider}" whose endpoint is unreachable` | `provider`: string; the selected provider key. | 不支援 | `config`: default config with `SELECTED_PROVIDER` = `{provider}`; endpoint = a closed/unroutable address. | `怎麼做`: write a resolvable default configuration selecting `{provider}` whose endpoint points at a closed port / unroutable address, so the connection fails. `權威狀態落地`: the endpoint refuses the connection. `回寫`: the configuration file. |
| `a configured provider "{provider}" whose endpoint answers with no usable answer` | `provider`: string; the selected provider key. | 不支援 | `config`: default config with `SELECTED_PROVIDER` = `{provider}`; endpoint = the fake. | `怎麼做`: write a resolvable default configuration selecting `{provider}` whose endpoint points at the fake; script the fake to return a body that carries no usable answer (e.g. an empty `choices`). `權威狀態落地`: the fake is scripted to answer with an uninterpretable body. `回寫`: the configuration file + the fake's scripted response. |

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts tellme with the prompt "{prompt}"` | `prompt`: string; the operator's prompt. | 不支援 | `旗標`: no `-c`; the configuration is the default for the effective mode. | `怎麼做`: run `tellme "{prompt}"` under the current environment. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |
| `the operator runs tellme's diagnostic with the prompt "{prompt}"` | `prompt`: string; a positional prompt supplied alongside the diagnostic flag. | 不支援 | `旗標`: the run is invoked with `-d` plus a positional prompt. | `怎麼做`: run `tellme -d "{prompt}"` under the current environment. `權威狀態落地`: `-d` dispatch precedes the prompt turn, so the diagnostic report is produced. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme sends exactly one request to the provider "{provider}"` | `provider`: string; the provider that must have received the request. | 不支援 | 無 | `必查`: `呈現結果`: the fake provider for `{provider}` recorded **exactly one** request. `權威狀態`: that request carried the prompt with the resolved model and credential for `{provider}`. `不該發生`: no request reached any other provider. |
| `tellme prints the provider's answer "{answer}"` | `answer`: string; the expected answer text. | 不支援 | 無 | `必查`: `呈現結果`: stdout contains the provider's answer `{answer}`. `不該發生`: the answer must not be swallowed, and must not be emitted only on stderr. |
| `tellme exits with the provider error code` | 無 | 不支援 | `碼值`: `6` (**pinned**, round 004 / Clarify Q3) — a non-zero code distinct from success (0) and from every other error class (usage 2, configuration 3, environment 4, diagnostic 5). | `必查`: `呈現結果`: the exit code **equals the pinned provider error code `6`**. `不該發生`: it must not collapse to the success code or to any other error class. |
