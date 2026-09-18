# history module DSL

Module-specific rows for the `history` module — the persisted session store and its `--new` / `-l`
session-lifecycle surfaces. Cross-module rows live in [`../dsl.md`](../dsl.md). Merged with the
interface root, every step in this module's feature must match exactly one row.

> The session history is an append-only JSON-Lines file (`history.jsonl`) under the per-mode session
> workspace (`$TELL_ME_HOME/output/<mode>/`); `--new` archives its lines into `history.archive.jsonl`.
> `{home}` stands for the runtime home (`TELL_ME_HOME`), per the workspace-module path convention.
> The store is read wholesale on resume and written one line per completed turn (round 007).

> **Round 014 (`014-session-replay-fidelity`):** the persisted tool step gains an optional, provider-agnostic `signature` — the token the model emitted for the call (for example the Gemini 3 `thoughtSignature`), persisted only when the provider supplies one (omitted when empty, so existing lines stay byte-identical) and replayed on resume so a tool-using Gemini session resumes faithfully. This module's own features are unchanged; the signed-exchange Given and the replay assertions live in the `chat` module.

## Given

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the session history already holds a tool-using exchange` | 無 | 不支援 | `工作區`: writes into `$TELL_ME_HOME/output/<mode>/history.jsonl` (the effective mode's session workspace). | `怎麼做`: create the per-mode session workspace directory if absent, then append one JSON line carrying `prompt`, `answer`, `calls: 2` (a tool-using turn makes two provider inference rounds), and an ordered `steps` array with one step (the widened, round-027 record from `specs/truth/data/data-model.dbml`). `權威狀態落地`: the active history holds a completed tool-using turn. `回寫`: the session history file. |
| `the runtime home holds the configuration "{config}" in mode "{mode}" holding the answer "{answer}"` | `config`: the config file name under `configs/`. `mode`: its `MODE`. `answer`: the seeded session's last answer. | 不支援 | `工作區`: writes `configs/{config}` (a resolvable config whose `MODE` is `{mode}`) and seeds `$TELL_ME_HOME/output/{mode}/history.jsonl` with one exchange (prompt "hello", answer `{answer}`). | `怎麼做`: write the named configuration and create the mode's session workspace with one `history.jsonl` line (round 053). `權威狀態落地`: the configuration exists and its `MODE`'s session holds that answer. `回寫`: the config file and the session history file. |
| `the runtime home holds the configuration "{config}" in mode "{mode}" whose turn log holds "{content}"` | `config`: the config file name. `mode`: its `MODE`. `content`: the turn log contents. | 不支援 | `工作區`: writes `configs/{config}` and seeds `$TELL_ME_HOME/output/{mode}/turns.log` with `{content}`. | `怎麼做`: write the named configuration and create the mode's session workspace with a `turns.log` file holding `{content}` (round 053). `權威狀態落地`: the configuration exists and its `MODE`'s session has a turn log. `回寫`: the config file and the turn-log file. |

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts a fresh session with "--new"` | 無 | 不支援 | `旗標`: `tellme --new` with no `-c` and no positional prompt. | `怎麼做`: run `tellme --new` under the current environment. `權威狀態落地`: the process has run to completion; the active history was archived and reset. `回寫`: captured exit code, stdout, stderr, and the resulting workspace files. |
| `the operator asks tellme to list the last {count} messages` | `count`: integer; the number of most recent messages to list. | 不支援 | `旗標`: `tellme -l {count}` with no `-c` and no positional prompt. | `怎麼做`: run `tellme -l {count}` under the current environment. `權威狀態落地`: the process has run to completion and makes no provider request. `回寫`: captured exit code, stdout, stderr. |
| `the operator asks tellme to list the last {count} messages of the configuration "{config}"` | `count`: integer. `config`: the config file name. | 不支援 | `旗標`: `tellme -l {count} -c <home>/configs/{config}` (no positional prompt). | `怎麼做`: run `tellme -l {count} -c <home>/configs/{config}` under the current environment; no session mode is forced by the environment. `權威狀態落地`: the process resolves the session named by `{config}`'s `MODE` and makes no provider request. `回寫`: captured exit code, stdout, stderr. |
| `the operator reviews the turn log of the configuration "{config}"` | `config`: the config file name. | 不支援 | `旗標`: `tellme -t -c <home>/configs/{config}` (no positional prompt). | `怎麼做`: run `tellme -t -c <home>/configs/{config}` under the current environment. `權威狀態落地`: the process resolves the session named by `{config}`'s `MODE`, prints its turn log, and makes no provider request. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the active session history holds no exchanges` | 無 | 不支援 | 無 | `必查`: `權威狀態`: the active history file `<runtime_home>/output/<mode>/history.jsonl` holds no exchange lines. `不該發生`: no active exchange may survive a fresh-session start. |
| `the archived session history holds the exchange "{prompt}" and "{answer}"` | `prompt`: string; the earlier prompt. `answer`: string; the earlier answer. | 不支援 | 無 | `必查`: `權威狀態`: the archive file `<runtime_home>/output/<mode>/history.archive.jsonl` holds a line whose prompt and answer are `{prompt}` and `{answer}`. `不該發生`: a fresh-session start must not destroy the previous conversation. |
| `tellme lists the last {count} messages` | `count`: integer; the number of messages listed. | 不支援 | `來源`: the arranged exchanges (the `the session history already holds the exchanges:` Given). | `必查`: `呈現結果`: stdout carries the last `{count}` messages of the arranged history, in order. `不該發生`: the listing must not include exchanges older than the last `{count}`. |
| `tellme lists the assistant message "{answer}"` | `answer`: the expected last assistant message. | 不支援 | `來源`: the session's persisted history. | `必查`: `呈現結果`: the LAST line on stdout is `assistant: {answer}`. `不該發生`: a listing must not report any other session's message (the #103 regression: the default session's text). |
| `tellme prints exactly the turn log line "{content}"` | `content`: the expected turn-log contents. | 不支援 | 無 | `必查`: `呈現結果`: stdout is exactly `{content}`. `不該發生`: the `-t` read must not add, reorder, or decorate the stored turn log. |
| `tellme prints nothing` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout carries no characters (a missing/empty turn log). `不該發生`: a missing turn log must not be an error. |
| `tellme lists no messages` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout carries no messages. `不該發生`: nothing may be listed for a session with no persisted history. |
| `tellme sends no request to any provider` | 無 | 不支援 | 無 | `必查`: `呈現結果` / `權威狀態`: every configured fake provider recorded zero requests. `不該發生`: a listing run must not contact a provider. |
| `tellme lists only the operator's messages` | 無 | 不支援 | `來源`: the arranged tool-using history (the `the session history already holds a tool-using exchange` Given). | `必查`: `呈現結果`: stdout carries the stored prompt and answer and **no** tool-step text. `不該發生`: the widened tool activity must not be surfaced by `-l`. |
| `no payload status is reported` | 無 | 不支援 | `通道`: standard error (the diagnostic stream). | `必查`: `呈現結果`: the captured standard error carries no payload status line (neither a `~`-prefixed estimate nor a measured line). `不該發生`: a non-prompt run must not report a payload status. |
