# chat module DSL

Module-specific rows for the `chat` module — the single-prompt reasoning turn. Cross-module rows live
in [`../dsl.md`](../dsl.md). Merged with the interface root, every step in this module's feature must
match exactly one row.

> **Parameter escapes (round 005, grill Q6):** a quoted parameter on a single Gherkin line supports the
> escapes `\n`, `\t`, `\\`, and `\xHH` (hex byte); the step definitions decode them. This lets a
> newline-terminated or control-byte-bearing answer be written without a docstring, so the output
> contract is falsifiable.

> **Output rows (round 005, grill Q4/Q5):** the `is exactly` row is the byte-fidelity assertion and the
> `carries no terminal decoration` row is the **FR-007 intent carrier** — they are mechanically
> equivalent while tellme has no renderer (both assert `stdout == answer bytes + one appended newline`);
> the distinction is intent-only, kept so Rule 1 retains a named decoration clause. The decoration row
> relies on the scenario's scripted answer (the `a configured provider … answers with …` Given) and
> fails loudly if it is absent.

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
| `the operator pipes "{content}" into tellme` | `content`: string; the text written to standard input. | 不支援 | `旗標`: no `-c` and no positional prompt; the configuration is the default for the effective mode. `輸入`: `{content}` is written to a piped standard input, then EOF. | `怎麼做`: run `tellme` (no positional prompt) with `{content}` on a piped standard input, under the current environment. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |
| `the operator pipes "{content}" into tellme with the instruction "{instruction}"` | `content`: string; the text written to standard input. `instruction`: string; the positional instruction argument. | 不支援 | `旗標`: no `-c`; the configuration is the default for the effective mode. `輸入`: `{instruction}` is the positional argument; `{content}` is written to a piped standard input. | `怎麼做`: run `tellme "{instruction}"` with `{content}` on a piped standard input. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |
| `the operator pipes nothing into tellme` | 無 | 不支援 | `旗標`: no `-c` and no positional prompt; the standard input is an empty pipe (immediate EOF). | `怎麼做`: run `tellme` with an empty piped standard input. `權威狀態落地`: the process has run to completion. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |
| `the operator runs tellme's diagnostic with "{content}" piped in` | `content`: string; the text written to standard input. | 不支援 | `旗標`: the run is invoked with `-d`; `{content}` is written to a piped standard input. | `怎麼做`: run `tellme -d` with `{content}` on a piped standard input. `權威狀態落地`: `-d` dispatch precedes any prompt handling, so the diagnostic report is produced. `回寫`: captured exit code, stdout, stderr, and the fake's recorded requests. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `tellme sends exactly one request to the provider "{provider}"` | `provider`: string; the provider that must have received the request. | 不支援 | 無 | `必查`: `呈現結果`: the fake provider for `{provider}` recorded **exactly one** request. `權威狀態`: that request carried the prompt with the resolved model and credential for `{provider}`. `不該發生`: no request reached any other provider. |
| `tellme prints the provider's answer "{answer}"` | `answer`: string; the expected answer text. | 不支援 | 無 | `必查`: `呈現結果`: stdout contains the provider's answer `{answer}`. `不該發生`: the answer must not be swallowed, and must not be emitted only on stderr. |
| `tellme exits with the provider error code` | 無 | 不支援 | `碼值`: `6` (**pinned**, round 004 / Clarify Q3) — a non-zero code distinct from success (0) and from every other error class (usage 2, configuration 3, environment 4, diagnostic 5). | `必查`: `呈現結果`: the exit code **equals the pinned provider error code `6`**. `不該發生`: it must not collapse to the success code or to any other error class. |
| `the request carried the piped content "{content}"` | `content`: string; the expected prompt text. | 不支援 | 無 | `必查`: `呈現結果`: the fake provider recorded exactly one request whose prompt is `{content}` (trimmed). `不該發生`: the recorded request must not carry extra or missing text. |
| `the request carried the instruction "{instruction}" followed by the piped content "{content}"` | `instruction`: string; the positional instruction. `content`: string; the piped content. | 不支援 | 無 | `必查`: `呈現結果`: the recorded request's prompt equals `{instruction}`, then a newline, then `{content}`. `不該發生`: the piped content must not replace or precede the instruction. |
| `the captured standard output is exactly "{answer}"` | `answer`: string; the expected answer text (decoded through the escape convention). | 不支援 | 無 | `必查`: `呈現結果`: the captured standard output equals `{answer}` — the answer bytes **verbatim** — followed by exactly one newline **appended by the CLI**. `不該發生`: no leading bytes and no system-added bytes beyond the single appended terminating newline. |
| `the captured standard output carries no terminal decoration` | 無 | 不支援 | 無 | `必查`: `呈現結果`: the captured standard output adds **no presentation control codes introduced by tellme** — the redirected stream equals the answer bytes verbatim, plus the single CLI-appended terminating newline (any control bytes the **answer itself** carries are passed through and are **not** flagged). `不該發生`: the system must not wrap or decorate the redirected answer with presentation control codes when stdout is not a terminal. |
| `tellme completes the turn without waiting for terminal input` | 無 | 不支援 | 無 | `必查`: `呈現結果`: the run completed (an exit code was captured) **within a bounded deadline** with the prompt read from the piped standard input; the process did not block for a terminal. `不該發生`: the run must not hang awaiting interactive input (a hang fails with an explicit deadline error, not a suite timeout). |
