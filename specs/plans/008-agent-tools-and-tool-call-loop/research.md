# Phase 0 Research: tellme Agent Tools & the Tool-Call Loop (Round 008)

Topic: give tellme an **agentic loop** — `tellme "<prompt>"` may call declared **read-only filesystem tools** and iterate (think → act → observe) until a final answer; persist the turn's tool activity; surface the loop live; and land **history summarisation as an agent tool** ([#22](https://github.com/gosharplite/tellme/issues/22)), while **token-budget pruning stays a settled exclusion**.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), output rendering (glamour), and session history (append-only JSON-Lines) were locked in rounds 001–007. The system still has **one CLI end** and adds **no new system end**, **no new external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. **Settled exclusions** (out of scope, not deferred): `-b`/`--retry`, history pinning, streaming responses, token-budget pruning, `SafePath`/interactive consent. **Deferred to this round**: history summarisation — **as an agent tool**.

---

## Decision 1: Tool abstraction — a network-free domain port + a registry

- **Decision**: Add a network-free domain port `internal/domain/tools` — a `Tool` value type (`Name`, `Description`, `Parameters` JSON-Schema) plus a `Registry` the loop dispatches against, and a concrete read-only filesystem adapter in `internal/infrastructure/tools` (or a `.../toolsfilesystem` sub-package). The loop resolves a model's requested tool by name through the registry and calls its executor.
- **Rationale**: Mirrors the existing provider seaming (`llm.Gateway` domain port + `infrastructure/llm` adapter) — the CLI/orchestration layer stays free of concrete tool code, and the loop is fake-testable. A registry keeps tool dispatch a data structure rather than a growing switch.
- **Alternatives considered**:
  - **Inline tool switch inside the CLI turn** — couples the loop to every tool; the CLI grows a domain concern — rejected.
  - **A single god-function that lists/reads by branch** — collapses two tools into one un-schema'd capability; not extensible — rejected.
  - **A plugin/registry loaded from config** — far beyond the first slice — rejected.

## Decision 2: Provider port widening — tool definitions + structured tool-call I/O

- **Decision**: Extend the existing domain port rather than add a parallel one. `llm.Request` gains the tool **definitions** (name/description/parameters) alongside the already-widened `Messages` (round 007); `llm.Response` is extended from `{Text}` to also expose the model's **tool-call requests** (id, name, arguments) and whether the response is a final answer vs a tool request. The OpenAI adapter maps this to the wire `tools` array, assistant `tool_calls`, and `tool`-role result messages.
- **Rationale**: The adapter already assembles the `messages` array (round 007) and normalizes `choices[0]`; widening the existing value types is a smaller change than a second port, and it preserves the round-004–007 call sites/fakes. This is the **minimal** tool-call shape — not the reference's full provider-agnostic `Thought` model (still deferred).
- **Alternatives considered**:
  - **A separate `ToolCaller` port** — duplicates transport ownership for one capability — rejected.
  - **Parse raw provider JSON in the CLI** — leaks the wire shape into the orchestration layer — rejected.
  - **Introduce the full `Thought`/`Part` model now** — larger surface than the round needs — rejected (deferred).

## Decision 3: The loop — bounded iteration inside one turn

- **Decision**: Replace the single `gw.Complete(...)` in `runTurn` with a bounded loop: send the conversation → if the response requests tools, execute each requested tool, append the tool result(s) as conversation messages, and repeat → terminate on a final answer or when the iteration count reaches **`MAX_TOOL_LOOP`** (default **`1000`**, env/config override). Each individual tool execution is bounded by a per-tool timeout.
- **Rationale**: Matches the reference's `maxToolTurns` Think→Act→Observe recursion (its sample `MAX_TURNS` is 1000). `MAX_TOOL_LOOP` is a tellme-specific name so it is not confused with the (already-different) `MAX_TURNS`. A large default gives headroom; the bound still guarantees termination and bounds per-turn cost (each iteration re-sends the growing conversation).
- **Alternatives considered**:
  - **Reference name `MAX_TURNS`** — collides with the reference's *user-turn* semantics and tellme has no such knob — rejected; the distinct `MAX_TOOL_LOOP` is unambiguous.
  - **A small fixed bound (e.g. 16)** — rejected by the clarify decision (alignment with the reference's generous default).
  - **Recursive/goroutine orchestration** — harder to bound and to assert deterministically — rejected; a simple counted loop is testable.

## Decision 4: The tools — two read-only filesystem capabilities, size-bounded, no path boundary

- **Decision**: Ship exactly two tools, both read-only and local: **`list files`** (enumerate a directory's entries) and **`read files`** (return a file's contents). `read files` output is **size-bounded by a fixed cap (1 MiB, matching the round-005 stdin cap)** and truncated with a marker if exceeded. The tools have **no path/safety boundary** (Clarify Q3); they read whatever path the model gives.
- **Rationale**: Read-only tools need no consent/`SafePath`, keeping the slice's safety surface empty (a settled exclusion is preserved). The size cap is required because **token-budget pruning is out of scope**: an unbounded file read could otherwise exhaust the assembled context. The 1 MiB cap deliberately reuses the round-005 `io.LimitReader` convention for consistency.
- **Alternatives considered**:
  - **Unbounded reads** — a single large file could exhaust the context — rejected.
  - **A path boundary / `SafePath`** — a settled exclusion; re-opening it is out of scope — rejected.
  - **`read files` with line-range arguments** — extra parameter surface beyond the first slice — deferred.

## Decision 5: Persist the turn's tool activity — widen `history_entry`, replay on resume

- **Decision**: Widen the persisted record (round 007 `history.Entry` / `history_entry`) to embed the completed turn's tool steps — e.g. `{prompt, answer, steps:[{tool, args, result}]}` — still written **append-after-complete** as one JSON-Lines line with **fixed field order and no timestamp/id** (byte-determinism preserved). On resume, the prior steps are **replayed** into the conversation sent to the provider (assistant tool-call + tool-result messages), so the model sees earlier tool activity. `-l N` continues to read **only** `prompt`/`answer`.
- **Rationale**: Clarify Q2 chose `tell-me-go` parity (the reference persists a Turn with all its ToolCalls). Embedding the steps in the same line keeps a single append-only file (no sidecar), and replay gives resume true fidelity. `-l` staying prompt/answer preserves the round-007 contract (Clarify: `-l` is operator-facing).
- **Alternatives considered**:
  - **Keep `{prompt, answer}`** — rejected by Clarify Q2 (loses tool activity).
  - **A separate sidecar file for tool steps** — more files/IO to keep consistent — rejected.
  - **Persist but do not replay** — the model would reason over a conversation missing its own tool results — rejected.

## Decision 6: Failure contract — new class phrase + exit code for an incomplete loop

- **Decision**: A tool-using run that **cannot complete** (iteration bound reached, or an unrecoverable tool/protocol failure such as a requested tool that is not declared) is reported with the **new dedicated frozen class phrase** `tellme: the tool request failed` and the **new exit code `7`**. A tool that merely **returns an error** is fed back to the model as the tool's result (non-terminal) and the loop continues.
- **Rationale**: Clarify Round 2 Q1 — an agent/tool failure is a distinct operational category from a provider failure (different remediation), warranting its own class, matching round 004 (provider → `6`) and round 006 (configuration → `3`). Keeping recoverable tool errors in-loop matches the reference's in-turn recovery.
- **Alternatives considered**:
  - **Reuse the provider phrase + code `6`** — mislabels a tool failure as a provider failure — rejected.
  - **Treat any tool error as terminal** — loses the model's ability to recover (the reference feeds results back) — rejected.

## Decision 7: Loop visibility — discrete log lines on `stderr`

- **Decision**: Emit a **discrete log line** to the **diagnostic stream (`stderr`)** for each tool-loop step (tool name, arguments, result/error), as the loop proceeds; `stdout` remains the answer stream (rendered / raw). This is **not** token-level streaming — the loop is a sequence of complete provider calls.
- **Rationale**: Clarify (extra Q2) — the operator must follow the loop live; `stderr` keeps `stdout` pipe-friendly (`… | other`, `-r/--raw`), consistent with the round-005/006 stream discipline. Keeping the log off `stdout` also leaves the final-answer contract unchanged.
- **Alternatives considered**:
  - **Log on `stdout`** — interleaves with (and corrupts) the piped/redirected answer — rejected.
  - **A TUI progress view** — a new dependency/surface — deferred.

## Decision 8: Testing — extend the fake provider; unit-test the tools and the loop

- **Decision**:
  - Extend the **local fake provider** (`net/http/httptest`) to **serve tool-call responses** and to **record the sent tool definitions + messages**, so a scenario can assert the loop (a tool was offered, requested, executed, and its result fed back).
  - Add **pure-helper / unit tests** for: the tool registry dispatch, the two tools (`list files`, `read files` incl. the size cap and a missing-path error), the loop bound (`MAX_TOOL_LOOP`) + timeout + the failure contract, and the widened history record (append + reload + `-l` reading only prompt/answer).
  - Keep the **offline no-network** set unchanged (`--version`, `-d`, `-l`, prompt-less boot); the tool-using turn is a chat path and may dial.
- **Rationale**: The loop's claims (tools offered/executed/result-fed-back, bound reached, failure phrase) are only falsifiable if the fake records what was sent and returns scripted tool requests; the tools and store are pure and cheaply unit-testable.
- **Alternatives considered**:
  - **Assert the loop via answer text only** — weaker, script-dependent — rejected in favour of recording the wire.
  - **Skip the offline extension** — not needed: the offline set is unchanged — accepted (no change).

## Decision 9: No new dependency

- **Decision**: The round is **stdlib-only** (`os`, `path/filepath`, `io`, `encoding/json`, `context`, `time`); `go.mod` / `go.sum` are untouched.
- **Rationale**: Read-only filesystem tools and a counted loop need nothing beyond the standard library; the project has stayed dependency-free since round 007 and the reference's tool layer is not imported.
- **Alternatives considered**:
  - **A JSON-schema library for tool parameters** — hand-written structs suffice for two tools — rejected.
  - **A third-party loop/agent framework** — out of proportion and a supply-chain cost — rejected.

---

## Residual risks / forward links

- **`history_entry` shape change is a data-truth MODIFY** → `/axb-data-plan` owns it (the DBML gains the tool-step embed); recorded via `truth-delta.md`.
- **Byte-determinism** (round-007 NFR-002): the widened line must keep fixed field order and store no timestamp/id so `history.jsonl` bytes stay reproducible.
- **Context growth without pruning**: the loop re-sends the growing conversation each iteration and the read cap bounds any single tool result — but there is **no** automatic pruning (settled exclusion); long tool loops are bounded by `MAX_TOOL_LOOP`, not by token budget.
- **`-l` contract**: unchanged (prompt/answer only); the widened steps are not surfaced by `-l`.
- **Class-phrase vocabulary 10 → 11** and exit codes `0/2/3/4/5/6 → +7`; both must be published in `specs/truth/features/cli/**/dsl.md` by `/axb-dsl-refine`.
- **`MAX_TOOL_LOOP` default 1000** is generous (reference parity); the per-turn input cost scales with iterations × re-sent history — a known cost characteristic, not a defect.
- **Deferred, still out of scope**: the full provider-agnostic `Thought` model, streaming responses, MCP, memory, and any write/shell tools.
