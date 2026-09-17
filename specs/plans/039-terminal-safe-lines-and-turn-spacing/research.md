# Technical Research: terminal-safe `[Tool …]` lines + turn-output blank-line grouping (round 039)

**Plan Package**: `specs/plans/039-terminal-safe-lines-and-turn-spacing`

**Status**: Decisions D1–D9 (ratifies the operator Q1–Q8 locked in `spec.md`).

## Context

Two folded workstreams on the `stderr`-bound diagnostic surface:

- **#80** — round 038 generalized nothing: `sanitizeControl` lives in `internal/ui/tooloutput.go` and is consumed by **one** formatter (`FormatToolOutputLine`). Its three siblings render the **same provenance class** (model-authored or externally-sourced text) and remain raw: `FormatToolReason` (the model's `reason`), `FormatToolResult` (a tool-result snippet — e.g. a `read_files` of an ESC-bearing log/binary, or an escape-bearing filename), and `FormatToolAction`'s argument keys/values. A `reason` containing `\x1b[31m` still reaches the stream and can tint the rest of the run.
- **Operator request** — the live turn output is a dense wall of lines. The operator wants it grouped by blank lines (before each call's begin block; before the trailing grouped `[Tool Reason]` block; before the post-status group).

### Reference check (`tell-me-go`)

- **Sanitization:** the reference has **no** tool-log control sanitizer (its only `sanitizeForTerminal` is a LaTeX→Unicode helper; its tool-log lines are written with its own colour codes and the raw model text). So generalizing tellme's sanitize is a **recorded divergence** (the same divergence round 038 recorded for `[Tool Output]`).
- **Spacing:** the reference writes `[Tool Engine]` → `[Tool Reason]`/`[Tool Action]` → `[Tool Result]` **immediately**, and its post-call tail (`[Tool Reason]` group + token line + metrics line) **without** blanks (`internal/ui/renderer_metrics.go` `LogToolCall` / `LogToolResult` / `renderPostCallStatus`). So the blank-line grouping is a **recorded divergence** too (an operator-requested readability choice).

Both are therefore recorded divergences, not silent parity breaks.

## Decisions

### D1 — One single-owned terminal-safe-line policy, consumed by all four `[Tool …]` formatters

`sanitizeControl` + its helpers (`isControlRune`, `escSequenceLen`, `csiLen`, `oscLen`, `genericEscLen`, `csiScanLimit`, `oscScanLimit`) move into a dedicated `internal/ui/sanitize.go` — **one** definition of the class — and are applied inside **every** `[Tool …]` formatter that renders externally-sourced text: `FormatToolOutputLine` (already), `FormatToolReason`, `FormatToolResult`, `FormatToolAction`. `FormatToolEngine` renders only literals + integers, so it has nothing to sanitize (it is a non-consumer by nature, not by omission). Reasons:

- **Single ownership (FR-003 / #69 theme).** One definition, one place to change; a future class tweak touches one file.
- **Presentation-only (FR-002).** The formatters are pure `ui` functions; the tool **result** and the call's raw arguments are other consumers and stay raw.
- **Regression guard.** The `[Tool Output]` behavior must stay **byte-identical** after the move (a round-038 regression pin).

Rejected: leaving the helper in `tooloutput.go` and importing it (a `[Tool Output]`-named file owning a general policy is the wrong home — the issue's own "current vs. scalable" note); sanitizing in `internal/agent` (would reach near the model-facing text and its stored steps).

### D2 — Same removed class and same order (fold → sanitize → cap)

The class is **unchanged** from round 038 (7-bit ESC-introduced sequences + stray C0/DEL; TAB kept; interior CR dropped; bytes `≥ 0x80` untouched; ASCII-gated ESC consumption; per-kind bounded window). For the reason and result the application order is **fold + trim → sanitize → rune cap** (Q3), so the cap bounds the **visible** output and the round-036 one-line/cap contract is preserved (cap set stays `{189, 200, 200}`). For the action, each rendered **value** is folded → sanitized → capped, and each **key** is sanitized (keys are uncapped today, left uncapped); the composed list is otherwise unchanged (sorted ascending, `reason` excluded).

Rationale: sanitizing **after** the cap would let a removed-sequence value print "short" while the visible text lost characters to the cap; sanitizing **before** the cap makes the cap a bound on what the operator sees — the round-036 intent.

### D3 — The blank-value suppression reads the sanitized value

The round-036 blank-reason suppression (`agentloop.logAction`, `agentloop.reasonsOf`, the defensive `callRenderer.OnCallEnd` guard) currently tests `TrimSpace(raw)`. It is **widened** to test the **folded + sanitized + trimmed** value, so a reason that is *only* control data (non-blank before, blank after removal — e.g. `"\x1b[31m"`) emits **no** `[Tool Reason]` row (extending round-036's "no dangling row" intent). This is an *extension of the input*, not a consolidation of the three sites — the single-ownership refactor stays on **#69**.

### D4 — The neutral-close restore stays `[Tool Output]`-scoped (Q4)

No per-line reset is added to `FormatToolReason`/`FormatToolResult`/`FormatToolAction`. Reasons:

- Those formatters render a **single** line and never open a multi-line block, so there is no block state to leave neutral; a reset line per tool-log row would be pure noise (and would be a *new* visible artefact, not a fix).
- Round 038's restore exists for the `[Tool Output]` **block** (a multi-line, byte-budgeted, killable stream); its invariant is about *the block*, not the whole stream.
- With D1's stripping, a coloured `reason`/`result` is already neutralized at the source; the only residual is what the *operator's own* prior terminal state was, which is out of scope.

Recorded so it is not silently re-litigated. The ADR is superseded (see D9) to carry the generalized scope.

### D5 — Blank-line grouping: three insertion sites, exact semantics (Q5–Q8)

| # | Site | Rule | Mechanism |
| --- | --- | --- | --- |
| 1 | call begin | one blank before the call block's **first** line — the `[Tool Reason]` line, else the `[Tool Action]` line | `agentloop.logAction` writes a leading `\n` (via the existing `withToolLog` guard) before the reason/action; **per call** (a `k`-call round ⇒ `k` blanks; the blank before the round's first reason is just the first instance) |
| 2 | post-call tail | one blank before the grouped `[Tool Reason]` block, **none inside it** | `callRenderer.OnCallEnd` `emit` writes one `\n` before the reason loop, only when `roundReasons` is non-empty |
| 3 | post-status | one blank before the measured `Payload:` + metrics + `Ready` group | the same `emit` writes one `\n` before the measured payload line, only when `usage.Reported` |

A reason-less call (no `[Tool Reason]` line) still gets a blank before its `[Tool Action]` (Q6) — the blank is tied to the **call block**, not the reason line. The blank lines are plain `\n` on `stderr`; nothing is added to a non-tool turn, `stdout`, a persisted record, or the offline paths.

Rejected: a round-level blank only (does not separate multiple calls — the operator explicitly wants **per-call**); blanks between the grouped tail lines (Q7 explicitly says none); a blank before every post-turn line (noise).

### D6 — Scope guard

Unchanged: the `[Tool Output]` header/separator literals, the block bound/stop/spinner-yield, `output_file`, the block's restore, the tool schemas, the tool **result**, the call's stored arguments, `stdout`, `-r`, flags, exit codes, the class-phrase vocabulary, the provider transport, and every persisted record. The blank lines ride the same diagnostic stream the tool-log lines already do (incl. `raw` / `-i`; not a history-replay projection like `-l`). Stdlib-only; POSIX-only; no new dependency.

### D7 — Witnesses (falsifiability)

| Layer | Witness |
| --- | --- |
| Unit (`internal/ui`) | hostile-fixture pins for `FormatToolReason` / `FormatToolResult` / `FormatToolAction` (ESC-bearing reason, result, key, and value) asserting **control-free + valid UTF-8** with visible text and the rune caps preserved (SC-001); a `[Tool Output]` **byte-identical** regression pin (D1's guard); an escape-only-reason suppression pin (D3). |
| Unit (`internal/agent`) | the loop's blank-before-the-call-block ordering (the leading `\n` precedes the reason line; a reason-less call's `\n` precedes the action line). |
| E2E (`tests/e2e`) | a scripted run whose reason/result carry an escape ⇒ the emitted `[Tool Reason]`/`[Tool Result]` lines are control-free (SC-002); blank-line **position** assertions for a multi-call round, a reason-less call, the grouped tail block, and the post-status group (SC-003). |

Falsifiability: (a) revert a sibling's sanitize → its unit pin fails; (b) remove a blank insertion → the corresponding E2E position assertion fails; (c) narrow the reason guard back to the raw value → the escape-only-reason pin fails — each reproduced then reverted.

### D8 — Truth & governance impact (ratified)

- `specs/truth/techstack.md` — **MODIFY** the **Agent tool loop** row (and slightly the **Agent command tool** row where the sibling-scope forward item is stated): the sanitize policy is single-owned and applies to **all four** `[Tool …]` formatters, and the live turn output is **blank-line grouped**.
- `docs/decisions/` — **ADD ADR 0008** ("Terminal-safe `[Tool …]` line policy + turn-output blank-line grouping"), **superseding ADR 0007** (whose recorded scope boundary — "the `[Tool Output]` surface only" — this round closes; 0007's `Status` line flips to `Superseded by 0008`, per the decisions README's immutability rule). ADR 0008 records D1–D5 + the recorded divergences.
- `specs/truth/features/cli/chat/**` + `chat/dsl.md` — **MODIFY** (a control-free-siblings Rule + a blank-line-grouping Rule + their `DSLRow`s).
- `/axb-api-plan` — **NOOP** (no HTTP surface). `/axb-data-plan` — checked **NOOP** (no persisted state).

### D9 — ADR lifecycle: supersede, don't edit

ADR 0007 is `Accepted` and its own text says a change to the removed class / neutral-close policy / sanitize seam **supersedes** it rather than editing it. Generalizing the policy's scope **is** such a change (it revises 0007's recorded "closed on `[Tool Output]` only" consequence). So round 039 writes **ADR 0008** and marks 0007 `Superseded by 0008` — the ADR 0005→0006 precedent.

## Residual risks (forward)

- Sanitizing a **key** (not just a value) is a small extension over the issue's literal ask (it said "argument values"); it is included because a key is equally model-authored and escapes in a key would leak identically. No cap is added to keys (unchanged shape).
- The blank-line grouping is a presentation **divergence** from the reference; if reference parity is ever re-sought for the tool log, this is the recorded place to revisit.
- The three-site blank-reason predicate (round 036) is only **extended**, not consolidated — its single ownership remains on **#69**.
