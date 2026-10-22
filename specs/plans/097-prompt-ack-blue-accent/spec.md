# Round 097 — `097-prompt-ack-blue-accent`

**Theme**: give the **input-capture acknowledgement** — `[HH:MM:SS] Input captured. Processing...` —
a **bright-blue whole-line accent** on a colour-enabled terminal, so the one line tellme owns that
marks *"your prompt has entered the session"* is visually distinct from the plain text around it.
The accent rides the **existing** round-054 chrome-colour gate unchanged, applies **uniformly on
every prompt surface**, and is **byte-identical** off the gate. **No other chrome element changes.**

**Branch base**: off `dev` `0f72b06` (session-88 closeout head).

**Anchor issue**: [#204](https://github.com/gosharplite/tellme/issues/204) — *"Colour the
input-capture acknowledgement `Input captured. Processing...` bright blue — the operator-voice
accent, uniform across all prompt surfaces"* (revision 3, operator-selected colour).

**Input (operator directive, 2026-10-22)**: *"Please start a new round in a new branch from dev
branch. The goal is to close issue #204."* — with the colour locked to **blue** after the operator
rejected green (*"Not green, it's for dollar"*).

---

## 1. Why this round

The operator's prompt is the one thing tellme hands to a provider on the operator's behalf, and the
session's first tellme-owned diagnostic line about it — the input-capture acknowledgement — is
currently **plain**, sandwiched between the (uncolourable) terminal echo and the already-accented
turn frame:

```
[13:02:41] Input captured. Processing...                          ← plain (this round colours it)
                                                                  ← blank
────────────…──────────────────────────────────────────────────    ← plain rule
╭─⠿ Turn 9 - butler
[13:02:41] Payload: +0 ~71627 tokens - butler - deepseek-flash     ← green MODE token
```

Two design constraints follow from the recorded history, and both are honoured here:

1. **It must be one line on *all* surfaces.** The obvious alternative — colouring the operator's
   prompt text — was investigated and **withdrawn** (issue #204 revision 1) because the prompt text
   appears **blue / white / absent** across `-i`-typing (`-i` while typing is the TUI's unstyled
   editor, white), `-i`-submit (tellme's re-echo, at `internal/cli/cli.go:855`), the Ctrl+D
   surface (the **terminal driver's** echo — unreachable: `internal/cli/prompt.go:116-122` never
   reads a character), and the positional surface (no echo at all). The ack, by contrast, is
   emitted inside a **single** `if opts.chrome` block (`internal/cli/cli.go:850-859`) that **every**
   prompt-bearing surface reaches ⇒ uniformity is structural.
2. **It must not be green.** Green is tellme's most-loaded accent, and its loudest observed
   instance is a **dollar figure** (`internal/ui/metrics.go:38-48` wraps `$%.4f` on the `╰─⠿ Ready`
   line) — three lines below the ack. **Blue** is the palette's least-loaded accent (exactly one
   meaning: `[USER] - N` in the `-l` listing, `internal/ui/listing.go:84`), it is the *same
   concept* (the operator's side of the conversation), and — decisively — its existing use is on a
   **different surface** (`stdout`, the offline listing) so the two never share a visual block.

## 2. The change (surfaces)

1. **`internal/cli/cli.go`** — the ack is emitted with the colour adapter's flag **hard-wired
   false**: `emitInputCaptured(env, dp.NewLines(false))` (`:858`). It becomes
   `dp.NewLines(chromeColour(opts, env))` — the **existing** single gate (terminal `stderr` AND `-r`
   off AND a prompt-bearing surface; `:1116`, the `spinnerGate` expression). No second predicate.
2. **`internal/ui/render_ports.go`** — `Lines.InputCaptured` gains the value receiver `(l Lines)`
   and returns the colour variant, mirroring `PayloadMeasured` (`:41`) and `ToolReason` (`:57`).
3. **`internal/ui/turn.go`** — `FormatInputCaptured` stays the **single plain entry point**
   (byte-identical, no signature change) and gains a sibling `formatInputCapturedColour(t, colour)`
   — the established pattern (`formatPayloadEstimateColour`, `formatToolReasonColour`).
4. **Truth** — `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` gains the
   element; `specs/truth/features/cli/chat/dsl.md` gains one colour row and its sibling rows are
   scoped/extended (noted below).
5. **Decision record** — a **new ADR 0067** extends **ADR 0023**'s element set (whose **body stays
   verbatim** — ADRs are immutable; the index row gains a forward pointer). ADR 0023 **D3** is an
   operator-locked **four-element** set, so a fifth element cannot be added by editing it.

**Not a behaviour change beyond the accent**: no text change, no ordering change, no new line, no
`stdout` change, no `turns.log` change, no gate change, no palette growth (`colorBlue`/`blue()`
already exist, `internal/ui/colour.go:62`).

---

## 3. User stories

### Story 1 — the acknowledgement is accented at a terminal (P1)

As the tellme operator working in a terminal, I want the input-capture acknowledgement to carry the
operator-voice accent, so that the moment my prompt enters the session is visually distinct from the
plain rule/header around it — without having to colour my own prompt text (which tellme cannot do
uniformly).

**Why this priority**: it is the whole round — the operator-visible value, and the only surface where
the accent can be consistent.

**Independent verification method**: with the diagnostics forced onto a terminal (`TELL_ME_FORCE_STDERR_TTY`),
the captured `stderr` carries the blue SGR pair around the whole ack line; with `-r`, and with a
redirected `stderr`, the ack bytes are **unchanged** from today.

**Acceptance shape**: **Given** a terminal `stderr` and `-r` off, **When** a prompt-bearing turn
runs, **Then** the ack line is wrapped in `\033[1;34m…\033[0m`; and **Given** `-r` (or a redirected
`stderr`), **Then** the ack is still printed, **plain**.

**Functional Requirements (FR)**:

- **FR-001**: on a colour-enabled surface, the input-capture acknowledgement line MUST be wrapped —
  **the whole line, its timestamp included** — in the existing bright-blue SGR pair
  `\033[1;34m` / `\033[0m`. [Verification Intent: observable → the colouring Example "A prompt turn
  at a terminal accents the acknowledgement in blue"]
- **FR-002**: the accent MUST be gated by the **existing** `chromeColour` predicate (a terminal
  `stderr` AND `-r` off AND a prompt-bearing surface) and MUST NOT introduce a second colour gate.
  [Verification Intent: observable → the negative Examples (redirected / `-r`) + the `carries no
  colour` row]
- **FR-003**: off the gate the emitted ack bytes MUST be **byte-identical** to today's plain form
  `[HH:MM:SS] Input captured. Processing...`. [Verification Intent: unobservable → the plain-form
  unit pin + the existing E2E non-terminal assertions]
- **FR-004**: the ack MUST still be **emitted** under `-r` — only its colour is suppressed.
  [Verification Intent: observable → an Example that asserts the ack is present AND uncoloured
  under `-r`]

**Non-Functional Requirements (NFR)**:

- **NFR-001**: the accent MUST add **no** palette growth and **no** new colour mechanism — it reuses
  `colorBlue` / `blue()` / `wrap()` (`internal/ui/colour.go`). [Verification Intent: unobservable →
  the `internal/ui` diff (no new constant/helper)]

### Story 2 — every prompt surface gets the same accent (P1)

As the tellme operator, I want the accent on **whichever** prompt surface I use — positional, piped,
Ctrl+D multi-line, `-b <N> "prompt"`, `-i` submit — so that the accent is a property of the
acknowledgement and never a per-surface accident (the defect that sank the prompt-text proposal).

**Why this priority**: uniformity **is** the rationale for choosing this line over the prompt text.
If the accent lands on one surface only, the round's premise is void.

**Independent verification method**: two Examples on **different** surfaces (one non-`-i`
positional; one `-i` submit) both assert the blue pair on the ack.

**Acceptance shape**: **Given** a terminal `stderr`, **When** the turn is driven on the positional
surface **and**, separately, on the `-i` surface, **Then** both captured `stderr` streams carry the
blue-wrapped ack.

**Functional Requirements (FR)**:

- **FR-005**: the accent MUST appear on **every** prompt-bearing surface — positional, piped, Ctrl+D
  multi-line, `-b <N> "prompt"`, and `-i` submit. [Verification Intent: observable → the two
  cross-surface colouring Examples]
- **FR-006**: the accent MUST NOT reach `stdout` or the session `turns.log`. [Verification Intent:
  observable → the existing `the session turn log carries no decoration` row + the `stdout`
  byte-exactness assertions]

---

## 4. Requirements (global)

> No cross-story global FR/NFR; Stories 1–2 carry the whole round.

### Key entities

- **Input-capture acknowledgement** — the round-017 prompt-surface diagnostic line
  `[HH:MM:SS] Input captured. Processing...`, emitted to `stderr` once per prompt-bearing turn and
  deliberately **not** recorded in `turns.log` (round 053 fold F-53-3(i)).
- **Chrome colour gate** — `chromeColour`: a terminal `stderr` AND `-r` off AND a prompt-bearing
  surface; the single predicate shared by the spinner and every accent (ADR 0023 D2).

## 5. Success criteria

### Measurable outcomes

- **SC-001**: with the diagnostics forced onto a terminal and `-r` off, the captured `stderr` carries
  `\033[1;34m[HH:MM:SS] Input captured. Processing...\033[0m`.
  [Verification Intent: observable → the colouring Example]
- **SC-002**: off the gate (`-r`, or a redirected `stderr`) the ack is present and carries **no**
  SGR byte. [Verification Intent: observable → the negative Examples]
- **SC-003**: the E2E contract stays green; the `make verify` aggregate and the domain-model drift
  gate are green. [Verification Intent: unobservable → `make check`]
- **SC-004**: no palette growth; no new `make verify` member; `go.mod`/`go.sum` unchanged.
  [Verification Intent: unobservable → the diff]

## 6. Edge cases

- **EC-001**: a **piped / redirected** `stderr` (the E2E default) MUST stay byte-plain — the accent
  is terminal-gated, and every existing E2E assertion runs off a terminal.
  [Verification Intent: observable → the `carries no colour` row]
- **EC-002**: `-r` MUST keep the **line** and drop only the **colour** (the ack gates on
  `opts.chrome`, **not** on `raw`; a naive `if raw { skip }` would delete a real line).
  [Verification Intent: observable → the `-r` Example]
- **EC-003**: a **non-prompt** path (the offline readers `-l` / `-t` / `--tool-usage`, a prompt-less
  `--new`) MUST NOT emit an ack at all — unchanged behaviour, no new line.
  [Verification Intent: observable → the existing offline readers' assertions]
- **EC-004**: the accent MUST NOT appear in `turns.log` — the file leg is a **separate** writer
  (`dp.NewLines(false)`, `internal/cli/cli.go:177`), so colour cannot leak by construction.
  [Verification Intent: observable → the turn-log `no decoration` row]

## 7. Invariants

- **I-1** — the accent is **presentation-only**: no provider request, no persisted state, no
  `stdout` byte changes, no `turns.log` changes.
- **I-2** — the colour-off path is **byte-identical** by construction (the `wrap()` helper returns
  the input verbatim when disabled).
- **I-3** — **no** second colour gate: the round reuses `chromeColour`.
- **I-4** — no Accepted-ADR **body** is edited (ADR 0023's D3 element set is extended by a **new**
  ADR; the index row gains a forward pointer only).
- **I-5** — frozen `specs/plans/NNN-*/**` packages are never modified.

## 8. Assumptions

- **A1** — **blue** is the operator-locked colour (issued after green was rejected as "the dollar
  accent"); the round does not re-open the colour debate. The theme readability pre-check
  (`[USER] - N` in `-l`) is the operator's, not a repo artifact.
- **A2** — a **new ADR (0067)** is the correct home rather than an edit to ADR 0023, because ADR
  0023 **D3** is an operator-locked four-element set and the repo's ADR bodies are immutable
  (supersede/forward-pointer, never edit). This mirrors ADR 0065 (supersedes 0030 D2) and ADR 0057
  (amends 0045 by index annotation).
- **A3** — the domain model **is** updated: `Chrome` + the `chrome-colour-terminal-gated` invariant
  are modelled behaviour (ADR 0041), so the ack joins the enumerated accent set in the **YAML**
  source, re-rendered with `make modelith-render`.
- **A4** — the multi-line hint (`MultiLineHint` / `TUIHint`) is **out of scope**: it is a different
  element on a different stream (the reference colours it **yellow on stdout**,
  `tell-me-go/internal/ui/capture.go:270`), and bundling it would re-create the multi-surface
  inconsistency.
