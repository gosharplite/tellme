# Research — round 097 `097-prompt-ack-blue-accent`

Truth owner: `/axb-technical-research` (the **Techstack** owner). **No** techstack truth change
(a `NOOP` row: no dependency, dev-tool, or gate change) + one decision record (**ADR 0067**).

The round's engineering work sits in `/axb-dsl-refine` (the colouring contract) and `/axb-implement`
(the one receiver + one formatter). The research question is narrow but decisive: **which colour, and
where does the decision live?**

## Decision 1 — the accent is **bright blue** `\033[1;34m` (`colorBlue`), whole line

- **Decision**: wrap the **whole** input-capture acknowledgement line — timestamp included — in the
  repo's existing bright-blue SGR pair (`internal/ui/colour.go:62`, `colorBlue = "\033[1;34m"`), reusing
  `blue()`/`wrap()`.
- **Rationale — blue is the least-loaded accent, and its one meaning is the same concept.**
  Palette inventory (verified against `internal/ui/**`):

  | Colour | Meanings today | Count |
  | --- | --- | --- |
  | green `0;32` | `[Tool Reason]` · MODE token (both payload lines) · measured token count · **session cost `$`** | **4** |
  | yellow `0;33` | `[Tool Action]` · `[TOOLS] - M (N calls)` (`-l`) | 2 |
  | grey `0;90` | `[Tool Output]` block · rules · timestamps | 2 |
  | magenta `1;35` | `[MODEL] - N` (`-l`) | 1 |
  | **blue `1;34`** | **`[USER] - N`** (`-l`) | **1** |
  | red `0;31` | payload over budget | 1 |
  | cyan `0;36` | **absent from tellme entirely** | 0 |

- **Why not green (the parity argument that *looks* strongest).** `tell-me-go` colours this exact line
  green (`tell-me-go/internal/ui/capture.go:237-238`), and tellme copied the line verbatim in round 017
  while dropping the colour — so greening it reads as *restoring a dropped element*. **Rejected** because
  green's dominant observed instance in tellme is a **dollar figure**: `internal/ui/metrics.go:38-48`
  wraps `$%.4f` on `╰─⠿ Ready`, three lines below the ack. The operator named this explicitly ("not green,
  it's for dollar"). ADR 0023's own *Recorded divergence* paragraph settles the authority question:
  tellme **"reuses the reference's code on tellme's element set"** — reference parity does not outrank
  palette coherence.
- **Why not cyan (the "obviously free" slot).** `grep -rn cyan` over `internal/ui/`, `specs/truth/`,
  `docs/decisions/`, `docs/domain-model/*.yaml` → **no match**: cyan is not a tellme colour at all. Its
  only precedent is the reference's `[Tool Engine]` line, which tellme renders **plain**
  (`internal/ui/toolcall.go:54` vs `tell-me-go/internal/ui/renderer_metrics.go:326`). Choosing cyan means
  a **new constant + helper + element** to honour a line tellme deliberately does not colour.
- **Adjacency check (the test green failed).** Green's other meanings live in the **same** `stderr` turn
  block as the ack (the payload/`Ready` lines, three lines below). Blue's one existing use is the `-l`
  listing's `[USER] - N` header on **`stdout`** — an **offline** surface that never renders in the same
  block. So blue's second meaning is a real but **non-colliding** stretch ("operator-authored text" →
  "the operator's channel").
- **Alternatives considered**:
  - **Green (reference parity)** — rejected (the `$` accent; adjacency collision).
  - **Cyan (an unused slot)** — rejected (palette growth; tool-loop association; no precedent in tellme).
  - **Magenta (pair with `[USER]`)** — rejected: magenta is `[MODEL]`; inverting the pairing is worse.
  - **A bold/underline attribute instead of a colour** — rejected: the repo's chrome accents are all SGR
    colours; a new attribute class would be a bigger, unargued change.

## Decision 2 — the gate is `chromeColour`, unchanged; the accent lives in the `ui.Lines` adapter

- **Decision**: the call site passes `dp.NewLines(chromeColour(opts, env))` instead of the hard-wired
  `dp.NewLines(false)` (`internal/cli/cli.go:858`). `chromeColour` (`:1116`) = `spinnerGate` = *terminal
  `stderr`* AND `-r` off AND a prompt-bearing surface. The colour bytes stay owned by `internal/ui`.
- **Rationale**: ADR 0023 **D2** already defines the single gate and **D5** already places the colour in
  the adapters ("the pure formatters keep their signatures; `ui.Lines` carries a `colour bool`"). The ack
  is the **only** chrome line still hard-wired plain — a leftover from before round 054 (when no chrome
  line was coloured), not a deliberate exclusion. No second predicate is introduced (FR-002).
- **Verified hazard (recorded so the implementer cannot fall into it)**: the ack is gated on
  `opts.chrome`, **not** on `opts.raw` — so a `if raw { skip }` refactor would **delete a real line** and
  red both `dsl.md`'s announced row and `presenting-the-turn.feature`. Under `-r` the line stays and only
  the colour drops (EC-002).
- **Alternatives considered**:
  - **Add a `Lines.InputCapturedColour(...)` port method with a bool argument** — rejected: the value
    receiver `(l Lines)` is the established shape (`PayloadMeasured`, `ToolReason`); a bool parameter on
    the port would push the gate decision into `internal/cli` twice.
  - **Colour in `internal/cli` directly** — rejected: violates the "bytes are owned by `internal/ui`"
    rule (ADR 0023 D5; the RULE-E layer baseline stays 0).

## Decision 3 — a **new ADR (0067)** extends ADR 0023's element set; the 0023 body stays verbatim

- **Decision**: **ADR 0067** records the fifth element (the ack, bright blue), the green rejection, the
  recorded divergence from the reference, and the palette-pressure note. `docs/decisions/README.md`'s
  **ADR 0023 index row** gains a forward pointer (`amends 0023 D3`); ADR 0023's **body is not edited**.
- **Rationale**: **ADR 0023 D3** is an explicit **operator-locked four-element set**, and the repo's
  Accepted-ADR bodies are immutable (round 096 **I-3**; the ADR 0065 → 0030 D2 and ADR 0057 → 0045
  precedents — both extend by new ADR + index annotation). Editing D3 in place would silently rewrite an
  operator lock and break `adr-index-consistent`'s spirit.
- **Alternatives considered**:
  - **Amend ADR 0023's D3 in place** — rejected: violates immutability and the "operator-locked" status.
  - **No ADR (a truth row only)** — rejected: the palette is a durable design surface with an existing
    ADR home (0023/0027/0045/0057); a fifth element with a *rejected* stronger alternative must be
    durable, or the next round re-litigates green.

## Decision 4 — the truth home is the `chat` module's colouring feature + its DSL rows

- **Decision**: the element is added to
  `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (a new Rule + two Examples) and to
  `…/chat/dsl.md` (one **ADD** row for the blue contract, one **MODIFY** row widening the prohibition row,
  one **MODIFY** row scoping the round-017 announced row to the non-terminal class).
- **Rationale**: the colouring feature already owns every chrome accent and every negative
  (`the session chrome carries no colour`, `the session turn log carries no decoration`). Splitting the
  ack's colour into a second feature would give one contract two authorities (`dsl-single-authority`).
  Scoping the announced row mirrors the round-039 `[Tool Reason]` row's existing qualifier verbatim
  ("scoped to a **non-terminal** `stderr` — on a terminal the line carries tellme's own … accent as a
  distinct, gated class").
- **Alternatives considered**:
  - **Put the colour clause on the round-017 announced row only** — rejected: that row is exercised by
    **many** non-terminal Examples across `presenting-the-turn`, `continuing-the-interactive-prompt`, and
    `reading-a-multi-line-prompt`; a colour clause there would be vacuous or contradictory on those
    fixtures.
  - **A new feature file** — rejected (two authorities for one contract).

## Decision 5 — the domain model **is** updated (it is modelled behaviour)

- **Decision**: extend the `Chrome` entity's description and the `chrome-colour-terminal-gated`
  invariant in `docs/domain-model/tellme.modelith.yaml` to name the acknowledgement as the fifth accent
  (bright blue), then `make modelith-render`.
- **Rationale**: ADR 0041 makes the model load-bearing; the `Chrome` entity already **enumerates** the
  accented elements and the invariant already **enumerates** the element set (round 073/082/086
  extensions). A fifth element is a modelled change; `modelith-check` (a `verify` member) fails on drift
  if the YAML and the rendered Markdown disagree.
- **Alternatives considered**: leave the model alone — rejected: the invariant would then be **false**
  (it would omit an element that exists), exactly the semantic rot ADR 0041 exists to prevent.

## Decision 6 — out of scope, deliberately

- **The multi-line hint** (`MultiLineHint` / `TUIHint`) — the reference colours its hint **yellow on
  `stdout`** (`tell-me-go/internal/ui/capture.go:270`) while tellme prints it **plain on `stderr`**
  (`internal/cli/prompt.go:117`). A different element on a different stream; bundling it would re-create
  the multi-surface split that sank issue #204 revision 1. Its own issue if wanted.
- **The `-i` submitted-prompt echo** — stays **plain** (issue #204 revision 1, withdrawn).
- **The reference's `\r` + `EL` prefix** before the coloured ack (`capture.go:285-291`) — adopting it
  changes **non-colour** bytes and interacts with spinner state. Recorded as a retained divergence.
- **The terminal's own keystroke echo** — requires raw mode + a bespoke line editor; rejected in #204 §2.

## Verification of the one-line change

`emitInputCaptured(env, dp.NewLines(false))` (`internal/cli/cli.go:858`) →
`emitInputCaptured(env, dp.NewLines(chromeColour(opts, env)))`. Traced: `chromeColour` → `spinnerGate` →
`opts.chrome && !opts.raw && env.stderrIsTerminal()`; all five prompt surfaces set `chrome: true`
(`:295` `-i`, `:429` positional/piped/`-b`, `:479` Ctrl+D). `emitInputCaptured` writes to `env.stderr`
only (`:1087-1088`), never to `env.turnsLog` — so FR-006 holds by construction.
