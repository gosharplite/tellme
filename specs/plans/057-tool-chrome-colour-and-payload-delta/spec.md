# Feature Specification: tool-chrome colour, a 500-rune `[Tool Action]` cap, and a payload increment (round 057)

**Feature Branch**: `057-tool-chrome-colour-and-payload-delta`

**Created**: 2026-09-19

**Status**: Draft — produced by `/axb-specify` from operator tasking. **Clarify round 1 OPEN** (asked one question at a time; the payload-increment semantics carry the open gaps). No `specs/truth/**` file is written by this skill.

**Input (operator, 2026-09-19, this session)** — three terse chrome requests, given after the operator inspected `tell-me-go`'s palette:

> *(1)* **`[Tool Output]` and the two horizontal lines will be grey in tellme.** *"`[Tool Action]` will be yellow in tellme."*
> *(2)* **`argValueCap` will become 500.**
> *(3)* On the payload status line, **`~203148/1000000` will become `+100 ~203148`** — *"`+100` is the increment of last `~203048` (for example)."*

**Behaviour intent**: **MODIFY (CLI-visible diagnostic chrome + one rendered-value cap + the payload status line).** All three requests are presentation-layer changes to `stderr` chrome; none touches persistence, the wire protocol, the tool contract, or a tool's behaviour. The round is deliberately a small **chrome** round except for the payload increment, whose backing state (where "the previous payload" comes from) is the round's one real design question.

---

## Grounded in the current system *(measured this session, `dev` @ `63a54bc`)*

| Site | Current shape |
| --- | --- |
| `internal/ui/colour.go` | The round-054 colour policy (**ADR 0023**): ONE constant pair — `colorGreen = "\033[0;32m"`, `colorReset = "\033[0m"` — and a `green(s, enabled)` wrapper. The doc comment records that the **ELEMENT SET is tellme's own** (a recorded divergence from the reference): the whole `[Tool Reason]` line, the `MODE` token in both `Payload` lines, the measured token number, and the `Ready` session cost. **Nothing else is coloured today** — there is no grey and no yellow constant in tellme. |
| `internal/ui/tooloutput.go` | `const ToolOutputSeparator = "------------------------------------------------------------"` (60 ASCII hyphens) + `ToolOutputHeader`; `Begin` writes the header then the opening separator, `EndWith`/`End` write the closing separator. **All three are emitted plain** (the round-038 sanitizer/sanitize policy governs the *streamed content*, and the block always closes the terminal in a neutral state). |
| `internal/ui/toolcall.go` | `const argValueCap = 189` — the maximum rendered rune length of ONE `[Tool Action]` argument **value** (fold → sanitize → cap; truncate iff > 189 → first 188 runes + one U+2026 counted inside the cap, rune-boundary cut). Argument **keys** are sanitized but **uncapped**; the joined list is **uncapped**. The cap set is documented as `{189, 200, 200}` (`argValueCap`, `resultValueCap`, `reasonValueCap`). |
| `internal/ui/status.go` | `FormatPayloadStatus(t, tokens, budget, mode, model, estimated)` renders `[HH:MM:SS] Payload: ~<tokens>/<budget> tokens - <mode> - <model>` (estimated pre-flight, `~`) or `… <tokens>/<budget> tokens …` (measured, no `~`). `formatPayloadStatusColour` adds the round-054 green accents (the `MODE` token in **both** lines; the **measured** token number only). The **budget** is the only place the effective token budget is displayed (round-024). |
| `internal/cli/call_renderer.go` | The payload lines are emitted **per AI-endpoint call** (round 034): `OnCallBegin` emits the **estimated** line from `llm.EstimatePayload(persona, toolDefs, messages)`; `OnCallEnd` emits the **measured** line from `usage.PromptTokens` (only when `usage.Reported`). A tool-using turn therefore emits several payload lines (one estimate + one measured per call). |
| `internal/domain/history/history.go` | `Entry{Prompt, Answer, Calls, Steps}` — the persisted record **carries no payload/token count** (`Calls` was added by round 027 as the precedent for a persisted counter). |
| `tests/e2e/steps/`, `specs/truth/features/cli/chat/dsl.md` | The payload line and the action-value cap are pinned in tests + interface truth (e.g. `watching-the-tool-loop.feature` *"A very long argument value is shortened to 189 runes"*, `Then … shortened to at most 189 runes`). |

**Consequence (current behaviour):** the tool chrome is plain except for the four round-054 green accents; a `[Tool Action]` argument value longer than 189 runes is elided with one `…`; and every payload line prints `<tokens>/<budget> tokens`, repeating a **constant** budget (`/1000000`) on every line while giving no sense of how much the payload *moved*.

---

## Settled design *(operator-stated this session — the plan is built to this shape)*

| # | Decision |
| --- | --- |
| **S-1** | **Grey = `[Tool Output]` header line + the two horizontal separators.** The whole `[Tool Output]` header line and **both** separator lines (the opening one and the closing one) are wrapped grey — a whole-line wrap, the round-054 idiom, not a marker-only tint. The streamed content lines are **not** in this request (assumed plain; flagged). |
| **S-2** | **Yellow = `[Tool Action]`.** The whole `[Tool Action] <tool>(<args>)` line is wrapped yellow. |
| **S-3** | **`argValueCap` 189 → 500.** The mechanic is unchanged — only the number changes: truncate iff > **500** runes → first **499 runes + one U+2026** counted inside the cap, rune-boundary cut, applied after `sanitizeControl(oneLine(...))` on the folded value. |
| **S-4** | **Payload line = `+<delta> <tokens>` and the `/budget` is dropped.** The rendered token segment becomes `+<delta> <tokens>` (e.g. `+100 ~203148`), where `<delta>` is the **increment of the current payload over the previous payload**. |
| **S-5** | **Colour follows the round-054 policy.** The new grey/yellow accents obey the same gate as round 054: emitted **only** when the diagnostic stream is a terminal **and** `-r` is off; never in `stdout`; never in `turns.log` (the round-053 plain-file leg). Colour codes are the reference's (`tell-me-go/internal/ui/colors.go`: `colorGray = "\033[0;90m"`, `colorYellow = "\033[0;33m"`). |

**The requested payload shapes** (the operator's example is the **estimated** line, which carries the `~`):

```
[11:52:03] Payload: +100 ~203148 tokens - butler - deepseek-flash
```

*(today: `[11:52:03] Payload: ~203148/1000000 tokens - butler - deepseek-flash` — budget removed, delta added.)*

---

## Clarify round 1 *(asked one question at a time; ≤ 3 questions this round; ≤ 5 per session)* — **OPEN**

| # | Question | Status |
| --- | --- | --- |
| **Q1** | **Which payload line(s) carry the increment, and does the measured post-turn line change too?** The estimated pre-flight line (your example, `~…`) always changes; the measured line (`Payload: 203148/1000000 tokens …`, no `~`) is a separate emission from `usage.PromptTokens`. Options: **(A)** both lines change — each prints its own delta and drops the budget (estimated prints its delta vs the previous estimate; measured prints its delta vs the previous measured reading); **(B)** only the **estimated** line changes (delta + no budget); the measured line keeps the absolute `tokens/budget` form; **(C)** only the **measured** line changes; the estimated line keeps today's form. | ⏳ **OPEN** |
| **Q2** | **What is "the previous payload" — and is it held in memory or persisted?** Options: **(A)** the last payload value **emitted in this session**, held in memory by the CLI (zero persistence; resets on `--new`/restart — a resumed session starts with no predecessor); **(B)** the previous **measured** `prompt_tokens` **persisted** on the turn record (a round-027-`Calls`-style `Entry` field; survives `--new`/restart; a data-model change under `data/**`); **(C)** the previous call **within the current turn** rather than the previous turn. | ⏳ **OPEN** |
| **Q3** | **Budget removal, no-predecessor, and negative deltas.** (a) Is `/budget` dropped from **every** payload line, or only from the line(s) that gain a delta? (b) With no predecessor (first payload of a session, or right after `--new`/restart), what prints — `+0`, an omitted delta, or the absolute-only form? (c) When the current payload is **smaller** than the previous one (compaction, pruned/summarised history, `--new`), does the delta print negative (`-1234`), or clamp to `+0`? | ⏳ **OPEN** |

**Non-negotiable invariants (proposed, not open):**

- **I-1** — the **mechanic** of every changed element is unchanged; only the operator-stated attribute changes (colour added / the cap number / the token-segment shape). Nothing else on those lines moves.
- **I-2** — colour **never** enters `stdout`, and `turns.log` keeps the **plain** leg by construction (the round-054 rule); the offline/`-r`/non-terminal paths stay **byte-identical** to the plain form.
- **I-3** — the payload line stays a **chrome** line: no persistence, transport, or model-visible change follows from the delta (unless Q2 → B, which would add a persisted `Entry` field — a data-model change the later phases must own).
- **I-4** — a **payload increment** is display-only: it must not alter the round-026 tool-usage accounting, the round-024 effective budget, or any CLI exit code.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the tool chrome is colour-coded (grey output block, yellow action line) (Priority: P1)

As the **operator**, when a tool runs I want the `[Tool Output]` block and the `[Tool Action]` line to carry colour — grey and yellow respectively — so the tool stream is faster to parse at a glance, exactly as the reference's palette colours them.

**Why this priority**: it is the round's lead request (given first, and the only one the operator checked for clarity).

**Independent verification**: on a terminal run with `-r` off, the captured `stderr` carries a grey-wrapped `[Tool Output]` header, grey-wrapped opening and closing separators, and a yellow-wrapped `[Tool Action]` line; with `-r` on (or a piped stream) none of those bytes appear.

**Acceptance Scenarios**:

1. **Given** a terminal run with `-r` off, **When** a shell-class tool streams output, **Then** the `[Tool Output]` header line and both horizontal separator lines are wrapped grey (`\033[0;90m … \033[0m`) and the block still closes the terminal in a neutral state.
2. **Given** the same run, **When** a tool call is logged, **Then** the whole `[HH:MM:SS] [Tool Action] <tool>(<args>)` line is wrapped yellow (`\033[0;33m … \033[0m`).
3. **Given** a run whose diagnostic stream is **not** a terminal, **or** `-r` is on, **Then** the `[Tool Output]` header, both separators, and the `[Tool Action]` line are **byte-identical** to today's plain forms (no escapes).
4. **Given** a terminal run, **When** the turn log is inspected, **Then** `turns.log` contains the **plain** header/separators/action line (no escapes) — the file leg stays plain by construction.

**Functional Requirements**:

- **FR-001**: tellme MUST wrap the `[Tool Output]` **header line** and **both** horizontal separator lines grey on a colour-enabled diagnostic stream (S-1/S-5).
- **FR-002**: tellme MUST wrap the whole `[Tool Action] <tool>(<args>)` **line** yellow on a colour-enabled diagnostic stream (S-2/S-5).
- **FR-003**: the colour MUST be emitted **only** when the diagnostic stream is a terminal **and** `-r` is off (the round-054 gate, S-5) — the plain form is the default everywhere else and is **byte-identical** to today (I-2).
- **FR-004**: the `turns.log` leg MUST stay **plain** (I-2) — colour MUST NOT reach `turns.log` or `stdout`.
- **FR-005**: the `[Tool Output]` neutral-close guarantee (round 038: the block always restores a neutral terminal state at close) MUST be preserved with the separators coloured.

### User Story 2 - the payload line shows how much the payload moved, without the constant budget (Priority: P1)

As the **operator**, when I read a turn's payload line I want to see the **increment** over the previous payload (`+100`) and **not** the constant `/1000000` budget, so the line tells me what this turn *added* instead of repeating a number that never changes.

**Why this priority**: it is the operator's third request and the round's only one with a real design decision behind it (where "the previous payload" comes from).

**Independent verification**: a two-turn session prints, on the second turn's payload line, a `+<delta> <tokens>` segment whose delta equals the difference from the first turn's payload; the `/budget` substring is absent; the run exits 0.

**Acceptance Scenarios**:

1. **Given** a session whose previous payload was `203048` and whose current estimate is `203148`, **When** the payload line is rendered, **Then** the token segment is `+100 ~203148` (S-4) and the line no longer contains `/1000000`.
2. **Given** a first payload of a session (no predecessor), **When** the line is rendered, **Then** the delta follows the Q3 → (b) decision (a defined, non-crashing form).
3. **Given** a current payload **smaller** than the previous one, **When** the line is rendered, **Then** the delta follows the Q3 → (c) decision (a defined form).
4. **Given** any payload line, **When** it is rendered, **Then** the rest of the line (timestamp, `Payload:`, `tokens`, ` - <mode> - <model>`) is unchanged (I-1) and `turns.log` still receives the plain form.

**Functional Requirements**:

- **FR-006**: the payload status line MUST render its token segment as `+<delta> <tokens>` (delta resolving per Q1/Q2), with the `/budget` portion removed per Q3 → (a) (S-4).
- **FR-007**: the line MUST resolve its delta and its no-predecessor / negative cases per the clarify decisions, deterministically (Q2/Q3).
- **FR-008**: the payload line's remaining content MUST be unchanged (timestamp, label, `tokens`, mode, model — I-1), and the round-054 green accents MUST be preserved or extended per the delta's colour decision.
- **FR-009**: the delta MUST be display-only — no persistence beyond whatever Q2 requires, no change to the round-024 effective budget, the round-026 tool-usage accounting, or any exit code (I-3/I-4).

### User Story 3 - a `[Tool Action]` argument value keeps more of its text (Priority: P2)

As the **operator**, when a tool call carries a long argument value I want to see up to **500** runes of it instead of 189, so I can read a large path/query/JSON fragment without it being elided so early.

**Why this priority**: it is a one-constant change (lowest risk, lowest urgency), so it is delivered last.

**Independent verification**: a `[Tool Action]` argument value of 501 runes renders 499 runes + one `…` (500 total); the unit pin's fixture and its `argValueCap` assertion move to 500, and the falsifiability witness (raise the cap → red) is re-established at the new value.

**Acceptance Scenarios**:

1. **Given** an argument value of exactly 500 runes, **When** `FormatToolAction` renders it, **Then** it is printed **whole** (no ellipsis).
2. **Given** an argument value of 501 runes, **When** it renders, **Then** it is truncated to the first 499 runes + one `…` (500 runes total, one U+2026 **inside** the cap).
3. **Given** an argument value containing multi-byte runes near the cap, **When** it renders, **Then** the cut is on a rune boundary and the result is valid UTF-8 (unchanged mechanic).

**Functional Requirements**:

- **FR-010**: ONE argument **value** on a `[Tool Action]` line MUST be capped at **500 rendered runes** (one U+2026 counted inside the cap), i.e. `argValueCap = 500` (S-3).
- **FR-011**: the cap MUST keep every existing property — sanitize-before-cap, fold, rune-boundary cut, valid UTF-8, keys uncapped, joined list uncapped, values never mutated (I-1).
- **FR-012**: the interface truth that names the number (the `[Tool Action]` value-cap Example/`Then` + the `dsl.md` note + the `techstack.md` cap-set) MUST be updated to 500, and the **recorded reference-parity divergence** MUST be extended: the reference caps at 189 **bytes**; tellme now caps at 500 **runes** — the magnitude diverges as well as the unit (ADR 0005 D5 lineage).

---

## Edge Cases

- **Tool-less turn / no `[Tool Output]` block** — nothing to colour; unchanged.
- **A non-terminal stream, `-r`, `--tool-usage` (offline), or the assembler gate** — all colour is absent; the plain bytes are identical to today (I-2).
- **`[Tool Output]` content lines** — **out of scope** (the operator named the header and the two separators only). Recorded assumption A3.
- **A very long `[Tool Action]` line** — the **list** stays uncapped (only each value is capped), so the line can still be long; the cap rise enlarges the worst case. No new whole-line cap is introduced.
- **`turns.log`** — plain leg (I-2); the round-053 tee renders the file leg plain by construction, and the new accents must not leak into it.
- **First payload of a session / after `--new` / after restart** — a defined delta form (Q3 → (b)).
- **Current payload < previous payload** (compaction / summarised history / pruned turns) — a defined form (Q3 → (c)).
- **Tool-using turn** — several payload lines per turn (one estimate + one measured per call); how the "previous payload" chains across them is Q2's scope.
- **MCP / error / spinner interplay** — untouched: the round-056 reason gate, the round-040 coordinator, and the round-054 green accents are unaffected.

## Key Entities

- **The chrome colour policy** — the round-054 single-owned colour constants + gate; this round adds `colorGray` and `colorYellow` to it (element set = tellme's own).
- **The `[Tool Action]` argument-value cap** — the rendered-rune ceiling on one argument value (189 → 500).
- **The payload status line** — the per-call `Payload: …` chrome line; its token segment changes from `<tokens>/<budget>` to `+<delta> <tokens>`.
- **The previous payload** — the value the delta is computed against (definition + storage per Q2).

## Success Criteria

- **SC-001**: On a colour-enabled terminal run, the `[Tool Output]` header, both separators, and the `[Tool Action]` line carry the grey/yellow wraps; with `-r`/non-terminal, all three are byte-identical to today.
- **SC-002**: `turns.log` and `stdout` never contain the new escapes (diff-level check).
- **SC-003**: A two-turn session renders `+<delta> <tokens>` whose delta equals the difference between the two payloads, with no `/budget` substring on the affected line(s).
- **SC-004**: A 501-rune argument value renders exactly 500 rendered runes (499 + one `…`); a 500-rune value renders whole.
- **SC-005**: The round's own gates hold: `gofmt`/`go vet` clean · `make verify` **OK** · `go test -count=1 ./...` green (all Examples) · `go.mod`/`go.sum` unchanged.
- **SC-006**: Every changed surface has an updated carrier (unit pin + interface-truth Example/`dsl.md` row + `techstack.md` row as applicable), and each changed behaviour has a reproduced falsifiability witness.

## Assumptions

- **A1**: The term **"the two horizontal lines"** = the two `ToolOutputSeparator` lines (the opening one, written by `Begin`, and the closing one, written by `EndWith`/`End`). **Not** the turn-rule line (`────…`) that frames a turn header.
- **A2**: **"grey"** and **"yellow"** are whole-line wraps using the reference's codes (`colorGray = "\033[0;90m"`, `colorYellow = "\033[0;33m"`); the **element set** is tellme's own (the round-054 divergence convention).
- **A3**: The `[Tool Output]` **content lines** (the streamed command output) stay **plain** — the request named the header and the two separators. (The reference tints its whole block yellow; tellme's sanitize/neutral-close policy governs the content and is unchanged.)
- **A4**: The delta is a **chrome** concern; unless Q2 → B (a persisted `Entry` field), `/axb-data-plan` is **NOOP** and `/axb-api-plan` is **NOOP**. `/axb-spec-by-example` is **NOT** NOOP (all three are operator-visible chrome) and `/axb-dsl-refine` is **NOT** NOOP (the action-cap Example and the payload-line rows are interface truth).
- **A5**: Truth impact is expected in `specs/truth/features/cli/chat/**` (the payload-line row(s), the `[Tool Output]` block rows, the `[Tool Action]` value-cap Example/`Then`) + `chat/dsl.md`, and `specs/truth/techstack.md` (the tool-loop / payload-status / colour rows); an **ADR** records the new palette elements, the cap change, and the delta semantics, with a §Forward.
- **A6**: The three requests are one round (a small, coherent chrome round), not three.

## Out of scope (recorded forward items)

- Colouring any **other** element (the round-054 RF-54-1 exclusion stands: until this round only the four round-054 accents were coloured; after it, exactly the six elements named here).
- A **non-terminal colour mode** (`--color=always`) — the gate stays terminal-only (round-054 RF-54-2).
- A whole-line cap for `[Tool Action]`, a cap on argument **keys**, or capping the joined list.
- Tinting the `[Tool Output]` **content** lines (A3).
- The locked exclusions: no security/consent layer · no Windows · sequential tool calls ([#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
