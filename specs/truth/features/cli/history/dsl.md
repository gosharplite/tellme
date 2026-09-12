# history module DSL

Module-specific rows for the `history` module — the persisted session store and its `--new` / `-l`
session-lifecycle surfaces. Cross-module rows live in [`../dsl.md`](../dsl.md). Merged with the
interface root, every step in this module's feature must match exactly one row.

> The session history is an append-only JSON-Lines file (`history.jsonl`) under the per-mode session
> workspace (`$TELL_ME_HOME/output/<mode>/`); `--new` archives its lines into `history.archive.jsonl`.
> `{home}` stands for the runtime home (`TELL_ME_HOME`), per the workspace-module path convention.
> The store is read wholesale on resume and written one line per completed turn (round 007).

## When

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the operator starts a fresh session with "--new"` | 無 | 不支援 | `旗標`: `tellme --new` with no `-c` and no positional prompt. | `怎麼做`: run `tellme --new` under the current environment. `權威狀態落地`: the process has run to completion; the active history was archived and reset. `回寫`: captured exit code, stdout, stderr, and the resulting workspace files. |
| `the operator asks tellme to list the last {count} messages` | `count`: integer; the number of most recent messages to list. | 不支援 | `旗標`: `tellme -l {count}` with no `-c` and no positional prompt. | `怎麼做`: run `tellme -l {count}` under the current environment. `權威狀態落地`: the process has run to completion and makes no provider request. `回寫`: captured exit code, stdout, stderr. |

## Then

| DSL 句型 | Gherkin 參數 | Data Table 參數 | 預設參數 | StepDef 實作語意 |
| --- | --- | --- | --- | --- |
| `the active session history holds no exchanges` | 無 | 不支援 | 無 | `必查`: `權威狀態`: the active history file `<runtime_home>/output/<mode>/history.jsonl` holds no exchange lines. `不該發生`: no active exchange may survive a fresh-session start. |
| `the archived session history holds the exchange "{prompt}" and "{answer}"` | `prompt`: string; the earlier prompt. `answer`: string; the earlier answer. | 不支援 | 無 | `必查`: `權威狀態`: the archive file `<runtime_home>/output/<mode>/history.archive.jsonl` holds a line whose prompt and answer are `{prompt}` and `{answer}`. `不該發生`: a fresh-session start must not destroy the previous conversation. |
| `tellme lists the last {count} messages` | `count`: integer; the number of messages listed. | 不支援 | `來源`: the arranged exchanges (the `the session history already holds the exchanges:` Given). | `必查`: `呈現結果`: stdout carries the last `{count}` messages of the arranged history, in order. `不該發生`: the listing must not include exchanges older than the last `{count}`. |
| `tellme lists no messages` | 無 | 不支援 | 無 | `必查`: `呈現結果`: stdout carries no messages. `不該發生`: nothing may be listed for a session with no persisted history. |
| `tellme sends no request to any provider` | 無 | 不支援 | 無 | `必查`: `呈現結果` / `權威狀態`: every configured fake provider recorded zero requests. `不該發生`: a listing run must not contact a provider. |
