# Feature Specification: The history listing reads like the reference (round 073)

**Feature Branch**: `073-list-role-headers-and-rendered-body`

**Created**: 2026-09-21

**Status**: Draft (specified — clarify pending; see §Clarify strategy)

**Anchor**: **operator request** — no anchor issue.

**Input (operator, 2026-09-21)**:

> *"tellme needs to have `[USER] / [MODEL] header lines (blue/magenta on a TTY stdout), body rendered as Markdown by glamour, blank line between messages`."*

…following the operator's own comparison of tellme's `-l` against the reference's:

| | tellme today | reference (`tell-me-go`) |
| --- | --- | --- |
| Output | plain `role: content` lines (lowercase, colon) | `[USER]` / `[MODEL]` header lines (colour on a TTY stdout) |
| Body | never rendered | Markdown rendered by glamour |
| Separator | none (one line per message, back to back) | a blank line between messages |

**Behaviour intent**: **MODIFY (an existing presentation surface)** — the offline `-l` / `--list` history listing must present each persisted message the way the reference does: a **role header line** (`[USER]` / `[MODEL]`, terminal-gated colour) followed by the message **body rendered as Markdown** and a **blank line** between messages. **No `specs/truth/**` file is written by this skill.**

---

## ⚠️ Read first

- **Only the presentation changes.** `-l` stays the same *command*: strictly offline, no provider request, terminal (it lists and exits), with `-l N` still meaning **the last N messages** (2 per persisted exchange — `prompt` then `answer`). This round does **not** adopt the reference's composability (`-l 5 "prompt"` listing *and* then chatting) nor its `-b/--back` pairing — see **A1**.
- **The role naming is the reference's, not tellme's stored naming.** tellme persists `{prompt, answer, …}` and maps them to `user` / `assistant` for display (`toMessages`). The reference's history holds `user` / `model`, so its headers read **`[USER]`** and **`[MODEL]`**. The operator asked for `[USER] / [MODEL]` → the assistant side must be **presented** as `MODEL`. This is a **presentation mapping only** — the frozen `history.jsonl` format is untouched (**I-2**).
- **Colour moves to the `stdout` axis.** Every tellme accent so far (rounds 054/057/058) is gated on **`stderr`** terminality, because that chrome is a diagnostic stream. This listing writes to **`stdout`**, so its colour is gated on **`stdout` terminality** — the reference's own axis for this surface (`useColor = isTTY(stdout) && !rawOutput`). The operator said "on a TTY stdout" → locked (**S-3**).
- **The colour codes are the reference's** (`tell-me-go/internal/ui/colors.go`): `colorBlue = "\033[1;34m"` (the `user` header) and `colorMagenta = "\033[1;35m"` (every non-`user` header). tellme's existing palette (`colour.go`) has green/gray/yellow only; two new codes are added (**S-4**).
- **glamour is already a dependency** (the answer path: `internal/ui` renders the answer's Markdown → ANSI). The listing reuses the same renderer + the same `WRAP_WIDTH` / `TELL_ME_WRAP_WIDTH` resolution (round 006) — no new module (**I-4**).

---

## Grounded in the current system *(measured 2026-09-21, `dev` @ `7ab84f6`)*

| Site | Current shape |
| --- | --- |
| `internal/cli/cli.go` — `renderHistoryList` | resolves the workspace (round 053: the `-c` config's `MODE`), loads entries, keeps the last `n`, then `fmt.Fprintf(env.stdout, "%s: %s\n", m.Role, m.Content)` per message — **plain, one line, no header, no rendering, no blank line**. |
| `internal/cli/cli.go` — `toMessages` | expands each stored exchange into `{user: prompt}` + `{assistant: answer}` — the **only** two roles the listing emits. |
| `internal/cli/cli.go` — `writeAnswer` | the answer path's render-vs-raw gate: `-r` alone (never stdout terminality) selects the raw bytes; otherwise `env.renderer.Render(answer, width)`, trimmed, + `"\n\n"`. |
| `internal/cli/cli.go` — `dispatchReporting` | `-d` → `-l` → `-t` → `--tool-usage`, each **terminal** (returns + exits); a positional prompt beside `-l` is **ignored** (unlike the reference's Phase 1/Phase 3 composition). |
| `internal/ui/colour.go` | the palette: `colorGreen` / `colorGray` / `colorYellow` + `colorReset`; `wrap(s, code, enabled)` — empty never wrapped, disabled ⇒ verbatim. **No blue, no magenta.** |
| `internal/cli/cli.go` — `runtimeEnv` | `isTTY func(any) bool` (a general stream probe, wired to `stdin` today) + `stderrTTY` (the diagnostic-stream probe). A **`stdout`** probe is expressible (`env.isTTY(env.stdout)`) but is not wired to any behaviour yet — round-006 / PR #16 **Obs 1** (no `stdout` probe) stays **OPEN**. |
| Reference `tell-me-go/internal/ui/history.go` — `renderHistory` / `historyRenderer` | `roleStr := "[" + strings.ToUpper(role) + "]"`; colour: `colorBlue` when `role == "user"`, else `colorMagenta`, only when `useColor`; body: `renderer.Render(text)` when not raw (glamour, `WithWordWrap(WrapWidth)` when `> 0`), else the text verbatim + a newline; `Fprintln(w)` after **every** content block (the blank separator); empty history ⇒ `No history found.` |
| Reference `history` roles | `user` / `model` (`internal/infrastructure/history/history.go:254-257`; the orchestrator writes `Role: "model"`). |

---

## Design (the shape is a `/axb-technical-research` decision; residual choices marked)

| # | Decision | Status |
| --- | --- | --- |
| **S-1** | The listing's role **header line** is `[USER]` for an operator message and `[MODEL]` for a model message (uppercase, bracketed) — the reference's shape. | locked (operator) |
| **S-2** | A message's **body** is rendered as **Markdown by glamour** (the existing renderer), not printed as raw Markdown source. | locked (operator) |
| **S-3** | The header **colour** is blue (`colorBlue = "\033[1;34m"`, operator) / magenta (`colorMagenta = "\033[1;35m"`, model), emitted **only when `stdout` is a terminal**; redirected/piped `stdout` stays byte-plain. | locked (operator) |
| **S-4** | One **blank line** separates consecutive messages. | locked (operator) |
| **S-5** | **Which roles** does the listing show — tellme's existing operator-messages-only (prompt + answer), or the reference's tool activity too (`[Tool Call]` / `[Tool Response]`)? | **clarify Q2** |
| **S-6** | Does the listing honour **`-r`/`--raw`** (reference parity: raw ⇒ body verbatim, no Markdown, no colour), or is the listing always rendered? | **clarify Q1** |
| **S-7** | Is the **operator prompt** body Markdown-rendered as well, or only the model answer? | **clarify Q3** |
| **S-8** | The exact renderer construction / the `WRAP_WIDTH` application / the `stdout`-terminal probe seam are **technical** choices. | **research decision** (D-x) |

**Non-negotiable invariants (proposed, not open):**

- **I-1 — Strictly offline & terminal** — the listing makes **no provider request**, reads no config beyond the mode, and still exits after listing; its precedence (`-d` → `-l` → `-t` → `--tool-usage`) and its prompt-ignoring behaviour are unchanged.
- **I-2 — The persisted format is frozen** — `history.jsonl` / `history.archive.jsonl` are untouched; the `[MODEL]` naming and the rendering are **presentation-only** (no migration, no write).
- **I-3 — Colour is `stdout`-gated and never leaks** — an escape sequence appears only when `stdout` is a terminal; a redirected `stdout` is byte-identical to the plain form; colour never reaches `stderr` and never enters `turns.log`.
- **I-4 — No new dependency; stdlib + the existing glamour; POSIX-only; hermetic.**
- **I-5 — The count semantics are unchanged** — `-l N` still selects the last `N` **messages**, the same slice today's code selects; only the bytes per message change.

---

## Clarify strategy

**Escalate (1–3 questions, one at a time).** Three high-impact gaps would change user-story splitting, acceptance, or the visible surface, so they go to `/axb-clarify` before the spec is treated as settled:

- **Q1 → S-6** — does `-l` honour `-r/--raw`? *(It changes the acceptance for every scenario and the agent-to-agent relay recipe `env -u TELL_ME_MODE … -l 1 -r -c "<target>.yaml"`; the reference's `-l` is `-r`-aware.)*
- **Q2 → S-5** — does the listing surface tool activity (`[Tool Call]` / `[Tool Response]`) as the reference does, or keep tellme's operator-messages-only listing? *(It changes **what** is listed — a round-007 truth explicitly asserts the tool activity is omitted.)*
- **Q3 → S-7** — is the operator prompt's body rendered too, or only the model answer's? *(It changes the per-message contract.)*

**Vetoable assumptions (disclosed, not asked — low impact, the operator asked only about the output format):**

- **A1 — Terminal, not composable.** The listing keeps tellme's terminal/offline shape (a prompt beside `-l` is ignored; no `-b` pairing). The reference's Phase 1/Phase 3 composition is **not** adopted this round; if the operator wants it, it is a separate request.
- **A2 — Empty stays silent.** An empty/absent session keeps printing **nothing** and exiting 0 (tellme's round-007 truth), i.e. the reference's `No history found.` sentence is **not** adopted. *(Vetoable: say the word and it is folded in.)*
- **A3 — The header is the carrier, colour is additive.** `[USER]` / `[MODEL]` is printed **always**; colour is a terminal-only accent. So a redirected listing still reads `[USER]` / `[MODEL]`.
- **A4 — No tool lines by default** *(pending Q2)* — if Q2 keeps the operator-only listing, the widened tool activity stays out, as today.

**No `NEEDS CLARIFICATION` beyond Q1–Q3**; the residual technical choices (S-8) defer to `/axb-technical-research`.

---

## User Stories (proposed)

### US1 — A listed message reads as a role head + a rendered body (Priority: P1)

When the operator lists the session, each message appears as a `[USER]` / `[MODEL]` header line followed by the message's Markdown-rendered body, with a blank line between messages — instead of the flat `role: content` line.

**Why P1**: it is the operator's whole request — the parity of the listing surface.

**Independent verification**: run `-l` over an arranged history whose answer carries Markdown emphasis; assert the header lines, the rendered body (markers absent), and the blank separators.

**Acceptance (proposed)**:

1. **Given** a session whose history holds an exchange whose answer carries Markdown emphasis, **When** the operator lists the last 2 messages, **Then** the listing shows `[USER]` before the prompt and `[MODEL]` before the answer, and the answer carries the **rendered** form (the literal `**` markers are absent).
2. **Given** the same session, **When** the operator lists, **Then** a blank line separates the prompt from the answer's header and the answer's body from the end.

**Functional requirements (FR)**:

- **FR-001**: A listed operator message MUST be headed by a `[USER]` line; a listed model message MUST be headed by a `[MODEL]` line (uppercase, bracketed).
- **FR-002**: A listed message's body MUST be rendered as Markdown (the existing glamour renderer), not emitted as raw Markdown source.
- **FR-003**: Consecutive messages MUST be separated by exactly one blank line.

**Non-functional requirements (NFR)**:

- **NFR-001**: The listing MUST stay strictly offline (no provider request) and MUST NOT change the persisted history.

---

### US2 — The listing is colour-correct on a terminal and byte-plain otherwise (Priority: P1)

On a terminal `stdout`, the header line is accented blue (`[USER]`) / magenta (`[MODEL]`); a redirected/piped `stdout` carries no escape sequences.

**Why P1**: the operator named the colour *and* its gate ("on a TTY stdout"); a piped listing must stay safe to parse.

**Independent verification**: drive the listing under a forced-`stdout`-terminal seam (colour present) and a pipe (no `\033`), and assert both.

**Acceptance (proposed)**:

1. **Given** `stdout` is a terminal, **When** the operator lists, **Then** the `[USER]` header line is wrapped in the blue SGR pair and the `[MODEL]` header line in the magenta pair.
2. **Given** `stdout` is a pipe, **When** the operator lists, **Then** the captured output carries no `\033` bytes, and the headers still read `[USER]` / `[MODEL]`.

**Functional requirements (FR)**:

- **FR-004**: The listing MUST emit blue (`\033[1;34m`) around the operator header and magenta (`\033[1;35m`) around the model header, **only** when `stdout` is a terminal.
- **FR-005**: The listing MUST emit no escape sequence when `stdout` is not a terminal (byte-identical to the plain form).

**Non-functional requirements (NFR)**:

- **NFR-002**: Colour MUST NOT reach `stderr` and MUST NOT enter `turns.log`.

---

### US3 — The count and selection semantics are unchanged (Priority: P2)

`-l N` still lists the last `N` **messages** of the session selected by the `-c` configuration's mode, offline.

**Why P2**: the round must not regress the round-007/053 contract while changing the bytes.

**Independent verification**: the existing `-l` E2E selection/`-c` scenarios, updated for the new per-message bytes.

**Acceptance (proposed)**:

1. **Given** a session with three exchanges, **When** the operator lists the last 2 messages, **Then** exactly the last exchange (prompt + answer) is listed, in order, offline
2. **Given** two configurations in different modes, **When** the operator lists via one of them, **Then** that mode's session is listed (round 053 behaviour preserved).

**Functional requirements (FR)**:

- **FR-006**: `-l N` MUST keep selecting the last `N` messages (2 per exchange) of the `-c`-selected session, with no provider request.
- **FR-007**: A missing/empty session MUST keep its current behaviour (see **A2**).

**Non-functional requirements (NFR)**:

- **NFR-003**: The change MUST NOT introduce a new dependency, a network access, or a non-deterministic byte.

---

### Edge cases

- When a message body is **empty**, the system MUST still emit its header and the separator (no crash, no stray blank-only message).
- When a body carries **terminal control sequences**, the system MUST apply the existing `internal/ui` sanitization policy rather than forwarding raw escapes to the operator's terminal (the round-038/039 rule; exact scope → `/axb-technical-research`).
- When `WRAP_WIDTH` / `TELL_ME_WRAP_WIDTH` is set, the rendered body MUST honour it (the round-006 resolution).
- When the listing is piped, the system MUST emit the plain (uncoloured) form.
- When a stored body already ends in a newline, the separator MUST remain exactly one blank line (no double gap).

## Requirements

### Global requirements

#### Functional requirements

- **FR-008**: The listing MUST reuse the one Markdown renderer + the one `WRAP_WIDTH` resolution the answer path uses (no second rendering policy).

#### Non-functional requirements

- **NFR-004**: The listing MUST remain hermetic and deterministic (no clock-dependent bytes; the E2E stays pty-free via a seam).

### Key entities

- **History message**: one displayed `{role, body}` — `role` ∈ {`user` → `[USER]`, model → `[MODEL]`}; `body` = the persisted prompt or answer text.

## Success criteria

### Measurable outcomes

- **SC-001**: 100 % of the `-l` E2E scenarios assert the new per-message shape (header + rendered body + blank separator), with `stdout` byte-exact on the piped path.
- **SC-002**: A terminal-`stdout` run shows the blue/magenta header accents; a piped run contains **zero** `\033` bytes.
- **SC-003**: The existing `-l` selection/`-c`/`offline`/`no provider request` assertions all still hold.
- **SC-004**: `make verify` + `go test -count=1 ./...` are green; no `go.mod` change; the topology audit adds no new error.

## Assumptions

- **A1** — Terminal, not composable (no `-l N "prompt"` chat pairing; no `-b` pairing).
- **A2** — An empty session keeps printing nothing (the reference's `No history found.` not adopted).
- **A3** — The header text is always printed; colour is a terminal-only accent.
- **A4** — Tool activity stays out of the listing unless **Q2** says otherwise.
- **A5** — The persisted `history.jsonl` format and the `-l N` count semantics are unchanged (presentation-only round).
