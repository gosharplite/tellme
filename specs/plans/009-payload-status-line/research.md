# Phase 0 Research: tellme Payload Status Line (Round 009)

Topic: give tellme **payload-budget visibility** — a prompt-bearing run reports, per turn, the size of the payload it is about to send (a pre-flight **estimate**) and the size the provider actually reported (**actual**), each measured against a configurable **payload budget** (`MAX_HISTORY_TOKENS`, default **1000000**), reproducing the reference's `[HH:MM:SS] Payload: <tokens>/<max> tokens - <mode> - <model>` status line. The round is **observe-only**: it introduces **no** token-budget pruning or automatic summarisation.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), output rendering (glamour), session history (append-only JSON-Lines), and the agent tool loop were locked in rounds 001–008. The system still has **one CLI end**, adds **no new system end**, **no new external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. **Settled exclusions** (out of scope, not deferred): `-b`/`--retry`, history pinning, streaming responses, **token-budget pruning**, `SafePath`/interactive consent. **IN this round**: a token estimator, a payload budget, and the status line.

---

## Decision 1: Token counting — a deterministic, stdlib-only heuristic behind an injectable seam

- **Decision**: Estimate a payload's size with a **hand-written, stdlib-only heuristic** over the assembled conversation text (a fixed bytes-per-token ratio), exposed behind a small injectable seam so tests are deterministic and provider-free. No external tokenizer.
- **Rationale**: The pre-flight figure is an **estimate** (the reference prefixes it with `~`); a deterministic offline heuristic needs zero dependencies and is fully testable, matching the reference's heuristic-estimator role. A real sub-word tokenizer would add a dependency and still would not match every provider's accounting.
- **Alternatives considered**:
  - **`tiktoken`/BPE tokenizer** — a new dependency and per-provider mismatch for a figure that is explicitly an estimate — rejected.
  - **A provider "count tokens" endpoint** — a network round trip before every turn for an estimate — rejected.
  - **Word/character count without a ratio** — materially worse accuracy for no gain — rejected.

## Decision 2: The "actual" figure — widen the gateway `Response` with the provider's reported usage

- **Decision**: Extend the provider gateway `Response` (the round-008 shape) with the provider's reported **usage** (prompt/completion/total tokens) and parse the OpenAI-compatible `usage` block in the adapter. The post-turn status line reads `<actual>` from it; when the provider reports no usage, the post-turn line is **omitted**. The usage is surfaced to the CLI by **widening the agent-loop seam**: `AgentLoop.Run` (`internal/agent/agentloop.go`) returns an `AgentResult{Answer, Steps, Usage}` — the usage of the **final** completion — instead of dropping `resp.Usage`, so `internal/cli` can render the post-turn line (BLOCKER-2).
- **Rationale**: The actual figure must be the provider's own accounting (the reference's post-call line uses the response's `prompt_tokens`). OpenAI-compatible providers already return `usage`; tellme currently discards it, so the change is a minimal widening of an existing value type with no new port.
- **Alternatives considered**:
  - **Re-estimate locally and call it "actual"** — it is not the provider's number; the `~`-vs-bare distinction would be vacuous — rejected.
  - **A separate usage event/bus** — tellme has no event bus and needs none for one field — rejected.
  - **Parse `usage` in the CLI** — leaks the wire shape into the orchestration layer — rejected.

## Decision 3: The status line — reference-parity format on `stderr`, always-on

- **Decision**: Emit two status lines per prompt-bearing run, both to the **diagnostic stream (`stderr`)**: a **pre-flight** `[HH:MM:SS] Payload: ~<est>/<budget> tokens - <mode> - <model>` before the provider request, and a **post-turn** `[HH:MM:SS] Payload: <actual>/<budget> tokens - <mode> - <model>` after it. The line is **always-on** — **not** gated by whether the stream is a terminal and **not** suppressed by `-r/--raw` — and carries **no** `tellme: ` prefix. Non-prompt paths (`--version`, `-d`, `-l`, prompt-less boot, prompt-less `--new`) emit **no** status line. The `<model>` field is the active provider's configured **`MODEL`** attribute (e.g. `deepseek-v4-flash`) — **not** the registry key — matching the reference (`session_manager.go` uses `ts.Model`; TD-2); `<mode>` is the effective mode.
- **Rationale**: Clarify Q1 (stderr) keeps `stdout` byte-exact for piping/`-r` and matches the reference's interactive path as well as tellme's round-008 `stderr` diagnostics; Clarify Q3 (always-on) matches the reference default and avoids the open PR #16 Obs 1 stdout probe. Excluding non-prompt paths keeps the offline commands' output contract intact.
- **Alternatives considered**:
  - **`stdout`** — interleaves with (and perturbs) the piped/redirected answer — rejected.
  - **TTY-gated** — requires wiring the stdout-TTY probe (Obs 1) this round — rejected.
  - **Flag/env-gated (off by default)** — diverges from the reference's always-on default — rejected.

## Decision 4: Line format & timestamp — an injectable clock seam

- **Decision**: Keep the reference's visible shape (`[HH:MM:SS] Payload: … tokens - <mode> - <model>`) and render the timestamp through an **injected clock seam** (defaulting to `time.Now`) so assertions are deterministic.
- **Rationale**: Reference parity for the operator-visible format; the clock seam lets E2E/unit tests either inject a fixed time or match the timestamp by pattern, without a pty or a `time.Sleep` (ADR-036 discipline).
- **Alternatives considered**:
  - **No timestamp** — diverges from the reference's line — rejected.
  - **`time.Now()` directly** — nondeterministic assertions — rejected.
  - **A full clock package** — over-provisioned for one format call — rejected (a single injected `func() time.Time` seam suffices).

## Decision 5: Payload budget — `MAX_HISTORY_TOKENS`, default 1000000, env-over-file

- **Decision**: Add `MAX_HISTORY_TOKENS` to the configuration, resolved **environment-over-file** with a default of **1000000** (user-locked), `>= 0` enforced; a negative value is a **configuration error** reusing the existing general configuration-invalid class phrase (round 006). Resolution mirrors the existing `MAX_TOOL_LOOP` resolver.
- **Rationale**: Reference parity (`MAX_HISTORY_TOKENS`), a user-locked default, and reuse of an established resolver shape and the existing class phrase (the frozen vocabulary is unchanged). `0`/non-positive handling mirrors the tool-loop resolver (falls back to the default) — a disclosed default.
- **Alternatives considered**:
  - **Reuse the model's context window** — tellme has no context-window/pricing table — rejected.
  - **Hard-code the budget** — not operator-choosable — rejected.
  - **A new knob name** — needless divergence from the reference — rejected.

## Decision 6: What the estimate counts

- **Decision**: The estimate counts the **assembled conversation the turn will send** — the resumed prior turns plus the current prompt (message text). Tool definitions are **excluded** (constant per run, small). The projection MUST be the single exported `agent.BuildMessages(prior)` — the same one the loop uses, including tool steps — not the legacy `cli.toMessages` which drops tool steps; otherwise a tool-using turn is undercounted (TD-1).
- **Rationale**: What the operator wants to gauge is the conversation's growth; excluding the constant tool schema keeps the number stable across turns and cheap to compute.
- **Alternatives considered**:
  - **Include the tool-schema JSON** — a larger, mostly-constant surface with marginal value — rejected.
  - **Count only the prompt** — understates the payload — rejected.

## Decision 7: No persistence — recompute per turn

- **Decision**: Per-turn token counts are computed **for display only**; the session-history record (`history_entry`) is **not** widened. On resume the pre-flight estimate is recomputed from the replayed history text.
- **Rationale**: The pre-flight figure is an estimate of what will be sent (recomputed from history); the actual is meaningful only for the just-completed call. Leaving `history_entry` unchanged keeps the data truth a **NOOP** and preserves the round-007 byte-determinism of `history.jsonl`.
- **Alternatives considered**:
  - **Persist a token count per entry** — a data-truth MODIFY that the displayed behaviour does not require — rejected.
  - **Persist only the latest count** — a sidecar to keep consistent for no observable gain — rejected.

## Decision 8: No new dependency

- **Decision**: The round is **stdlib-only** (`time`, `strings`/`unicode/utf8` for the estimate); `go.mod` / `go.sum` are untouched.
- **Rationale**: A heuristic estimator, a config resolver, a status-line formatter, and one extra parsed JSON field need nothing beyond the standard library.
- **Alternatives considered**:
  - **A tokenizer library** — see Decision 1 — rejected.
  - **A rendering/table library for the line** — the line is a fixed `fmt.Fprintf` — rejected.

## Decision 9: Testing & BDD techstack — unchanged

- **Decision**: No new system end and no change to the BDD techstack or strategy. E2E (`godog` against the built binary) covers the acceptance path; the estimator, the `MAX_HISTORY_TOKENS` resolver, and the status-line formatting are **pure-helper unit tests**. Extend the **local fake provider** to report usage (and a no-usage variant); assert the `stderr` status lines (timestamp matched by pattern) alongside a **byte-exact `stdout`**.
- **Rationale**: The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are not re-decided; the round adds no dependency and reuses the round-008 fake.
- **Alternatives considered**:
  - **Assert the line via `stdout`** — that is exactly what the round must not do — rejected.
  - **A pty harness for the timestamp** — unnecessary with a clock seam — rejected.

---

## Residual risks / forward links

- **Class-phrase contract**: the status line MUST NOT carry the reserved `tellme: ` prefix, so the "exactly one `tellme: ` line" contract and the frozen class-phrase vocabulary are untouched (FR-014). `/axb-dsl-refine` publishes the line's format as a CLI truth row.
- **Estimate vs actual**: the pre-flight figure is a heuristic estimate (`~`) and the post-turn figure is the provider's actual — the divergence is expected and encoded by the `~`, not a defect.
- **`usage`-absent providers**: the post-turn line is omitted (edge case); the pre-flight line still appears.
- **Budget semantics**: `MAX_HISTORY_TOKENS` is **displayed, not enforced** — this round introduces **no** pruning or automatic summarisation (a settled exclusion). A `usage`-driven budget is a known forward link, not a this-round behaviour.
- **Clock determinism**: the timestamp flows through an injected clock seam so E2E assertions stay deterministic without a pty.
- **Metrics line out of scope**: the reference's `M:`/`H:`/`C:` token counts and `$cost` summary are **not** reproduced (tellme has no pricing table); only the payload **status** line is.
- **`MAX_HISTORY_TOKENS: 0`** mirrors the tool-loop resolver (non-positive → default) — a disclosed default, not a spec change.
- **Deferred, still out of scope**: the full provider-agnostic `Thought` model, streaming responses, MCP, memory, and any write/shell tools.
