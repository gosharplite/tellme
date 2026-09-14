# Feature Specification: tellme post-turn status lines — per-turn metrics + session summary with cost (round 018)

**Feature Branch**: `018-post-turn-status-lines`

**Created**: 2026-09-14

**Status**: Draft — post-turn surface (metrics line + Ready line) locked; pricing = config-only

**Input**: Operator request, verbatim: add the reference's **post-turn** block to `tellme` — a per-turn **metrics line** and a closing **`╰─⠿ Ready`** line — with two locked edits:

- **simplify line 2** → drop the reference's trailing `($cost)` and `[timing]` segments, keeping only `[<provider>] M: … H: … C: … Th: …`;
- **change line 3** → group the summary with ` - ` separators and put `O` before the percentage: `╰─⠿ Ready ($a $b $c - M: … H: … O: … - …%)`.

Grounded against the reference: `tell-me-go` `internal/ui/renderer_metrics.go` (`renderMetricsLineLocked`) and `internal/ui/renderer.go` (`formatFinalCost` / `renderFinalSummary`), plus the reference's per-mode `output/<mode>/tokens.log` usage log.

Behaviour intent: **ADD** the reference's post-turn operator status — a **per-turn metrics line** and a closing **`╰─⠿ Ready` session summary** (with real costs) — to `tellme`'s prompt turns, backed by a **config-supplied pricing table** and a **per-mode per-call usage log**. This is the deliberate counterpart of round 017, which added only the *pre*-turn chrome and explicitly deferred the post-turn lines (round-017 spec **A7**).

**Locked decisions (operator interview, 2026-09-14)**:
- **`Th` always shown** — `Th: 0` is rendered (deviation from the reference's suppress-when-0).
- **Pricing** — introduce a pricing table (cost accounting); **no built-in rates**: rates come only from a config `MODELS` override; an un-priced model shows `$0.0000` (the `$` group still renders).
- **Three costs** — `#1` = the **API call that just returned**; `#2` = **all API calls this turn** (prompt + tool loop); `#3` = the **current session**.
- **Line-2 source** — line 2's `M/H/C/Th` and `$`#1 both come from the **call that just returned**.
- **Storage** — persist per-API-call usage to a per-mode `output/<mode>/tokens.log` (reference mechanism); the session total reads it; `--new` resets it.
- **Presence** — lines 2–3 appear on **every prompt-bearing turn** (mirroring the round-009 payload line), on `stderr`, plain text; **suppressed only when the provider reports no usage**.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The turn reports its token metrics (Priority: P1)

As an operator, I want `tellme` to show the just-finished API call's token breakdown after the answer, so I can see how the prompt and reply consumed the model's window.

**Why this priority**: it is the first of the two requested lines and the source of the cache accounting the summary line reuses.

**Independent verification**: run a prompt turn hermetically (injected streams + injected clock); assert the diagnostic stream carries one metrics line `[HH:MM:SS] [<provider>] M: … H: … C: … Th: …` after the answer, that its fields match the just-returned call's usage, and that `Th:` is present even at 0.

**Acceptance Scenarios**:

1. **Given** a resolved setup and a prompt, **When** the turn answers, **Then** the diagnostic stream carries one line `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>` after the answer.
2. **Given** a provider that reports `thinking_tokens: 0` (or omits it), **When** the turn answers, **Then** the line still ends `… C: <n> Th: 0`.
3. **Given** a tool-loop turn (several API calls), **When** it answers, **Then** the metrics line reflects the **last** call (the one that just returned), not the sum.
4. **Given** a provider response with no usage block, **When** the turn answers, **Then** no metrics line is written.

**Functional Requirements**:

- **FR-001**: On every prompt-bearing turn, after the answer, the system MUST write exactly one metrics line to the diagnostic stream: `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>`.
- **FR-002**: The fields MUST come from the API call that just returned: `M = prompt_tokens − cached_tokens`, `H = cached_tokens`, `C = completion_tokens`, `Th = thinking_tokens`.
- **FR-003**: The `Th:` segment MUST always be rendered, including `Th: 0`.
- **FR-004**: `[<provider>]` MUST be the active provider's registry label.

**Non-Functional Requirements**:

- **NFR-001**: The line MUST be plain text (no ANSI), written to `stderr` only, with its timestamp from the injected clock seam.

---

### User Story 2 - The turn reports its cost and the session summary (Priority: P1)

As an operator, I want a closing `╰─⠿ Ready` line that shows the cost of the last call, of the whole turn, and of the session — plus the session's token totals and cache-hit rate — so I can see what each request cost.

**Why this priority**: it is the other requested line and the round's reason for introducing pricing; without it the `$` figures would have no meaning.

**Independent verification**: run a prompt turn against a config `MODELS` pricing entry; assert the Ready line `╰─⠿ Ready ($<call> $<turn> $<session> - M: … H: … O: … - …%)` appears after the metrics line, that the three costs equal the call / turn / session costs, and that the session totals + hit-rate match.

**Acceptance Scenarios**:

1. **Given** a single-call turn on a fresh session, **When** it answers, **Then** the diagnostic stream carries `╰─⠿ Ready ($0.0005 $0.0005 $0.0005 - M: 236 H: 56576 O: 70 - 99.6%)` (all three costs equal; session == turn == call).
2. **Given** a tool-loop turn, **When** it answers, **Then** `$<call>` is less than `$<turn>` (the turn sums the tool-loop calls) and `$<session>` includes the turn.
3. **Given** a resumed session with prior turns, **When** a new prompt answers, **Then** the Ready line's session totals and `$<session>` include the prior turns.
4. **Given** `--new`, **When** the first prompt answers, **Then** the session totals reset (session == turn == call).
5. **Given** an active model with no `MODELS` pricing entry, **When** the turn answers, **Then** the `$` group renders `$0.0000` and the token totals + hit-rate still render.

**Functional Requirements**:

- **FR-005**: After the metrics line, the system MUST write exactly one summary line to the diagnostic stream: `╰─⠿ Ready ($<call> $<turn> $<session> - M: <sM> H: <sH> O: <sO> - <hit%>%)`.
- **FR-006**: `$<call>` MUST be the cost of the API call that just returned; `$<turn>` the summed cost of every API call in the current turn (the prompt completion and each tool-loop call); `$<session>` the cumulative cost of the current session.
- **FR-007**: `M/H/O` MUST be the session-cumulative miss / cached / output token totals (`O = C + Th`), and `<hit%>` the session cache-hit rate `H / (M + H)` formatted `%.1f%%`.
- **FR-008**: Costs MUST be computed from a **config-supplied** pricing table (`MODELS: { <model>: { PRICING: { HIT, MISS, COMP } } }`, env-over-file); the system MUST NOT ship built-in pricing rates.
- **FR-009**: When the active model has no pricing entry, the `$` figures MUST render `$0.0000` (the `$` group MUST still render).
- **FR-010**: Per-API-call usage (tokens + cost) MUST be persisted to a per-mode `tokens.log` after each call; the session total reads it; `--new` MUST reset the session's usage totals.

**Non-Functional Requirements**:

- **NFR-002**: The costs MUST be formatted `$%.4f` and the hit-rate `%.1f%%`; token counts MUST be integers (no float drift in the totals).

---

### User Story 3 - The new lines are bounded and leave the rest untouched (Priority: P2)

As an operator, I want the post-turn lines limited to the prompt surfaces and off `stdout`, so piping, rendering, the `-i` TUI, and the non-prompt commands keep their current behaviour.

**Why this priority**: it bounds the change and protects rounds 004–017; it is a constraint on Stories 1–2.

**Independent verification**: run each non-prompt path and a `-r`/piped run hermetically; assert lines 2–3 appear on `stderr` only, that `stdout` is byte-exact, and that a no-usage provider suppresses both lines.

**Acceptance Scenarios**:

1. **Given** a `-r`/piped run, **When** the turn answers, **Then** the two lines are on `stderr` and `stdout` is byte-exact.
2. **Given** a provider response with no usage, **When** the turn answers, **Then** neither line is written (the answer and any class phrase are unchanged).
3. **Given** `--version`, `-d`, `-l N`, boot, or a prompt-less `--new`, **When** it runs, **Then** neither line is written.
4. **Given** a failure path (provider / tool / history), **When** it fails, **Then** the class phrase + exit code are unchanged and no post-turn line is written.

**Functional Requirements**:

- **FR-011**: Lines 2–3 MUST be emitted only on prompt-bearing turns (positional, piped, the round-012 reader, and the `-i` submit path), on `stderr`, as plain text; they MUST NOT be written to `stdout`.
- **FR-012**: Lines 2–3 MUST be suppressed when the provider reports no usage; otherwise they MUST be emitted.
- **FR-013**: `stdout` MUST remain byte-exact; the frozen class-phrase vocabulary MUST be unchanged (the lines carry no `tellme: ` prefix); the round-017 chrome and every non-prompt path MUST be unchanged.

**Non-Functional Requirements**:

- **NFR-003**: The new behaviour MUST be deterministic (no `time.Sleep`, no real terminal); the injected clock + streams are the verification surface.

---

### Edge Cases

- **No usage block** (provider omits usage) or a **provider/tool failure** → lines 2–3 are suppressed; the class phrase + exit code are unchanged.
- **Tool-loop turn** → line 2 reflects the **last** call; `$<turn>` exceeds `$<call>`; line 3's session totals still accumulate.
- **Un-priced model / no `MODELS`** → the `$` group shows `$0.0000`.
- **`--new`** → session totals (and the `tokens.log`) reset; the first turn shows call == turn == session.
- **Resumed session** → session totals include the prior turns read from `tokens.log`.
- **`-r` / non-terminal stderr** → the lines are still emitted (plain text).
- **Provider omits cached/thinking** → `H = 0` (so `M = prompt`) and `Th: 0`.
- **Very large counters** → no overflow/panic in formatting.

## Requirements *(mandatory)*

> Story-specific FR / NFR are attached under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-014**: The round MUST NOT introduce a new frozen class phrase (the class-phrase vocabulary is unchanged), and the round-009 payload line and the round-017 turn chrome MUST be unchanged.

#### Non-Functional Requirements

- **NFR-004**: The round MUST add **no new third-party dependency** (pricing is pure arithmetic over already-parsed usage).

### Key Entities *(include if feature involves data)*

- **Usage record** (`output/<mode>/tokens.log`): one JSON line per API call — timestamp, provider, model, cached / prompt / response / thinking / total tokens, and the computed cost.
- **Pricing table** (`MODELS` config): per-model `HIT` / `MISS` / `COMP` rates (USD per million tokens), env-over-file; no built-in defaults.
- **Turn metrics line**: the per-call `[HH:MM:SS] [<provider>] M: … H: … C: … Th: …` line.
- **Session summary (Ready) line**: the `╰─⠿ Ready (…)` line — three costs, the session token totals, and the cache-hit rate.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A prompt turn's `stderr` carries, in order, the pre-flight payload line, the answer (on `stdout`), the measured payload line, the metrics line, and the Ready line — verified hermetically (covers FR-001–FR-009).
- **SC-002**: `Th:` is present on every metrics line, including `Th: 0` (covers FR-003).
- **SC-003**: The three costs equal the call / turn / session costs; an un-priced model renders `$0.0000`; a resumed session accumulates and `--new` resets (covers FR-006, FR-009, FR-010).
- **SC-004**: Both lines are suppressed exactly when the provider reports no usage, and `stdout` is byte-exact (covers FR-011–FR-013).
- **SC-005**: The behaviour is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (reference)**: the parity target is `tell-me-go` `renderer_metrics.go` (metrics line) + `renderer.go` (`formatFinalCost` / `renderFinalSummary`), with the two locked deltas (line 2 drops the `($cost)` and `[timing]` segments; line 3 groups with ` - ` and puts `O` before the `%`).
- **A2 (surfaces)**: lines 2–3 share the round-009 payload line's surfaces — every prompt-bearing turn (positional / piped / the round-012 reader / the `-i` submit path); `stderr`; plain text (no ANSI).
- **A3 (pricing source)**: config `MODELS` override only, **no built-in defaults** (operator decision); an un-priced model renders `$0.0000`, group rendered.
- **A4 (storage)**: per-call usage lives in `output/<mode>/tokens.log` (the reference mechanism); `--new` resets the session totals; a resumed session accumulates. [NEEDS CLARIFICATION: whether `--new` archives/rotates `tokens.log` exactly as it archives `history.jsonl` — an `/axb-data-plan` determination]
- **A5 (Th)**: always rendered, including 0 (operator decision; a deviation from the reference).
- **A6 (fallbacks)**: when the provider omits cached/thinking tokens, `H = 0` (so `M = prompt`) and `Th: 0`.
- **A7 (derived)**: `O = C + Th`; hit-rate = `H / (M + H)`; costs `$%.4f`; token counts are integers.
- **A8 (config shape)**: `MODELS: { <model>: { PRICING: { HIT, MISS, COMP } } }`, env-over-file; the exact keys / env name are a `/axb-technical-research` determination. [NEEDS CLARIFICATION: exact config key names / env override variable]
- **A9 (platform / verification)**: POSIX-only; no new dependency; verification is hermetic (injected clock + streams; no pty).
- **A10 (format stability)**: the `$` group is always rendered (never omitted), even at `$0.0000`.
